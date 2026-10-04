// SPDX-License-Identifier: MIT
// Package sequence executes the bounded oracle operation histories shared by
// offline corpus comparisons and native fresh-C differential fuzzing.
package sequence

import (
	"slices"
	"testing"

	box2d "github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

// InputLimit bounds a history to 64 fixed-width operations.
const InputLimit = 256

// Seed identifies an independently authored native C corpus input.
type Seed struct {
	Name  string
	Input []byte
}

// Execute compares a Go history against the supplied C observations. Inputs
// beyond InputLimit use the bounded prefix; the C stdin runner uses that prefix.
func Execute(t *testing.T, r *oracle.Reader, input []byte) {
	t.Helper()
	s := differentialSequence{t: t, r: r}
	s.execute(input[:min(len(input), InputLimit)])
}

// The protocol has 21 four-byte operations, not a simulation language. Every
// byte string is legal: absent operands are zero, indices are bounded, and an
// operation whose object precondition is absent reports acted=false.
// See oracle/differential.c for the independent native implementation.
type differentialSequence struct {
	t                 *testing.T
	r                 *oracle.Reader
	world             box2d.World
	bodies            [4]box2d.Body
	shapes            [4]box2d.Shape
	joints            [2]box2d.Joint
	savedBodies       [4]box2d.Body
	savedShapes       [4]box2d.Shape
	savedJoints       [2]box2d.Joint
	oldBodies         []box2d.Body
	oldShapes         []box2d.Shape
	oldJoints         []box2d.Joint
	image             []byte
	epoch             uint32
	initialGeneration uint16
	initialSlot       uint16
}

func differentialScalar(x byte) float32 { return float32(int8(x)) / 16 }
func differentialPosition(x, y byte) box2d.Pos {
	return box2d.OffsetPos(box2d.Pos{}, box2d.Vec2{X: differentialScalar(x), Y: differentialScalar(y)})
}

// Map only the world slot, which is process-local. Keep every allocation index
// and object generation bit. Never access state through a retained invalid ID.
func differentialID(id uint64) uint64 { return id &^ uint64(0xffff0000) }
func differentialWorld() box2d.World {
	d := box2d.DefaultWorldDef()
	d.Gravity = box2d.Vec2{Y: -10}
	d.EnableSleep = false
	d.WorkerCount = 0
	return box2d.CreateWorld(&d)
}
func (s *differentialSequence) vec(key string, v box2d.Vec2) { s.r.Vec2(key, v.X, v.Y) }
func (s *differentialSequence) pos(key string, p box2d.Pos) {
	s.r.Pos(key, oracle.Real(p.X), oracle.Real(p.Y))
}
func (s *differentialSequence) rot(key string, q box2d.Rot) { s.r.Rot(key, q.C, q.S) }
func (s *differentialSequence) pair(a, b box2d.Shape, c box2d.Contact) {
	s.r.U64("pair.a", differentialID(a.Store()))
	s.r.U64("pair.b", differentialID(b.Store()))
	s.r.Bool("pair.a.valid", a.IsValid())
	s.r.Bool("pair.b.valid", b.IsValid())
	id := c.Store()
	s.r.U32("contact.index", id[0])
	s.r.U32("contact.generation", id[2])
	s.r.Bool("contact.valid", c.IsValid())
}
func (s *differentialSequence) events(w box2d.World) {
	begin := slices.Collect(w.ContactBeginEvents())
	s.r.I32("begin.count", int32(len(begin)))
	for _, e := range begin {
		s.pair(e.ShapeIDA, e.ShapeIDB, e.ContactID)
	}
	end := slices.Collect(w.ContactEndEvents())
	s.r.I32("end.count", int32(len(end)))
	for _, e := range end {
		s.pair(e.ShapeIDA, e.ShapeIDB, e.ContactID)
	}
	hit := slices.Collect(w.ContactHitEvents())
	s.r.I32("hit.count", int32(len(hit)))
	for _, h := range hit {
		s.pair(h.ShapeIDA, h.ShapeIDB, h.ContactID)
		s.pos("hit.point", h.Point)
		s.vec("hit.normal", h.Normal)
		s.r.F32("hit.speed", h.ApproachSpeed)
	}
	moves := slices.Collect(w.BodyMoveEvents())
	s.r.I32("move.count", int32(len(moves)))
	for _, b := range moves {
		s.r.U64("move.body", differentialID(b.BodyID.Store()))
		s.pos("move.position", b.Transform.P)
		s.rot("move.rotation", b.Transform.Q)
		s.r.Bool("move.sleep", b.FellAsleep)
	}
}
func (s *differentialSequence) body(b box2d.Body) {
	s.r.U64("body.id", differentialID(b.Store()))
	valid := b.IsValid()
	s.r.Bool("body.valid", valid)
	if !valid {
		return
	}
	s.pos("body.position", b.Position())
	s.rot("body.rotation", b.Rotation())
	s.vec("body.velocity", b.LinearVelocity())
	s.r.F32("body.angular", b.AngularVelocity())
	s.r.I32("body.type", int32(b.Type()))
	s.r.Bool("body.enabled", b.IsEnabled())
	s.r.Bool("body.awake", b.IsAwake())
	s.r.F32("body.mass", b.Mass())
	contacts := slices.Collect(b.ContactData())
	s.r.I32("body.contacts", int32(len(contacts)))
	for _, c := range contacts {
		s.pair(c.ShapeIDA, c.ShapeIDB, c.ContactID)
		s.vec("manifold.normal", c.Manifold.Normal)
		s.r.F32("manifold.rolling", c.Manifold.RollingImpulse)
		s.r.I32("manifold.count", int32(c.Manifold.PointCount))
		for _, p := range c.Manifold.Points[:c.Manifold.PointCount] {
			s.vec("point.a", p.AnchorA)
			s.vec("point.b", p.AnchorB)
			s.r.F32("point.separation", p.Separation)
			s.r.F32("point.normal", p.NormalImpulse)
			s.r.F32("point.tangent", p.TangentImpulse)
			s.r.F32("point.total", p.TotalNormalImpulse)
			s.r.F32("point.velocity", p.NormalVelocity)
			s.r.U32("point.id", uint32(p.ID))
			s.r.Bool("point.persisted", p.Persisted)
		}
	}
}
func (s *differentialSequence) state(w box2d.World) {
	c := w.Counters()
	s.r.I32("world.bodies", int32(c.BodyCount))
	s.r.I32("world.shapes", int32(c.ShapeCount))
	s.r.I32("world.joints", int32(c.JointCount))
	s.r.I32("world.contacts", int32(c.ContactCount))
	s.r.I32("world.islands", int32(c.IslandCount))
	s.events(w)
}
func (s *differentialSequence) observe() {
	s.r.U32("world.epoch", s.epoch)
	s.r.U32("world.generation.delta", uint32(uint16(s.world.Store())-s.initialGeneration))
	s.r.Bool("world.slot.reused", uint16(s.world.Store()>>16) == s.initialSlot)
	s.state(s.world)
	for i := range 4 {
		s.r.I32("slot", int32(i))
		s.body(s.bodies[i])
		sh := s.shapes[i]
		s.r.U64("shape.id", differentialID(sh.Store()))
		valid := sh.IsValid()
		s.r.Bool("shape.valid", valid)
		if valid {
			f := sh.Filter()
			s.r.U64("shape.category", f.CategoryBits)
			s.r.U64("shape.mask", f.MaskBits)
			s.r.I32("shape.group", int32(f.GroupIndex))
		}
	}
	for _, j := range s.joints {
		s.r.U64("joint.id", differentialID(j.Store()))
		valid := j.IsValid()
		s.r.Bool("joint.valid", valid)
		if valid {
			s.r.U64("joint.a", differentialID(j.BodyA().Store()))
			s.r.U64("joint.b", differentialID(j.BodyB().Store()))
			s.r.F32("joint.length", (box2d.DistanceJoint{Joint: j}).Length())
			s.vec("joint.force", j.ConstraintForce())
		}
	}
	s.r.I32("retained.bodies", int32(len(s.oldBodies)))
	for _, b := range s.oldBodies {
		s.r.U64("retained.body", differentialID(b.Store()))
		s.r.Bool("retained.body.valid", b.IsValid())
	}
	s.r.I32("retained.shapes", int32(len(s.oldShapes)))
	for _, sh := range s.oldShapes {
		s.r.U64("retained.shape", differentialID(sh.Store()))
		s.r.Bool("retained.shape.valid", sh.IsValid())
	}
	s.r.I32("retained.joints", int32(len(s.oldJoints)))
	for _, j := range s.oldJoints {
		s.r.U64("retained.joint", differentialID(j.Store()))
		s.r.Bool("retained.joint.valid", j.IsValid())
	}
	s.r.U64("world.hash", s.world.StateHash())
}
func (s *differentialSequence) query(w box2d.World, x, y byte) {
	p := differentialPosition(x, y)
	a := box2d.AABB{LowerBound: box2d.Vec2{X: -2, Y: -2}, UpperBound: box2d.Vec2{X: 2, Y: 2}}
	f := box2d.DefaultQueryFilter()
	var hits []uint64
	w.OverlapAABB(p, a, f, func(sh box2d.Shape) bool {
		hits = append(hits, differentialID(sh.Store()))
		return true
	})
	// Query traversal order is not promised; retain duplicate hits and generation
	// bits in the multiset. Cancellation is observed separately.
	slices.Sort(hits)
	s.r.I32("query.count", int32(len(hits)))
	for _, id := range hits {
		s.r.U64("query.shape", id)
	}
	count := int32(0)
	w.OverlapAABB(p, a, f, func(box2d.Shape) bool { count++; return false })
	s.r.I32("query.cancel.count", count)
	origin := box2d.OffsetPos(box2d.Pos{}, box2d.Vec2{X: -10, Y: differentialScalar(y)})
	r := w.CastRayClosest(origin, box2d.Vec2{X: 20}, f)
	s.r.Bool("ray.hit", r.Hit)
	if r.Hit {
		s.r.U64("ray.shape", differentialID(r.ShapeID.Store()))
		s.pos("ray.point", r.Point)
		s.vec("ray.normal", r.Normal)
		s.r.F32("ray.fraction", r.Fraction)
	}
}
func (s *differentialSequence) replay(p *box2d.Replay) {
	s.t.Logf("replay observation frame=%d end=%t", p.Frame(), p.IsAtEnd())
	s.r.I32("replay.frame", int32(p.Frame()))
	s.r.Bool("replay.end", p.IsAtEnd())
	s.r.Bool("replay.diverged", p.HasDiverged())
	s.r.I32("replay.diverge.frame", int32(p.DivergeFrame()))
	s.state(p.WorldID())
	s.r.I32("replay.bodies", int32(p.BodyCount()))
	for i := range p.BodyCount() {
		s.body(p.BodyID(i))
	}
	s.r.I32("replay.queries", int32(p.FrameQueryCount()))
	for i := range p.FrameQueryCount() {
		q := p.FrameQuery(i)
		s.r.I32("replay.query.type", int32(q.Type))
		s.r.I32("replay.query.hits", int32(q.HitCount))
		for j := range q.HitCount {
			h := p.FrameQueryHit(i, j)
			s.r.U64("replay.query.shape", differentialID(h.Shape.Store()))
			if q.Type != box2d.ReplayQueryOverlapAABB && q.Type != box2d.ReplayQueryOverlapShape {
				s.pos("replay.query.point", h.Point)
				s.vec("replay.query.normal", h.Normal)
				s.r.F32("replay.query.fraction", h.Fraction)
			}
		}
	}
	s.r.U64("world.hash", p.WorldID().StateHash())
}
func (s *differentialSequence) recording(x, y byte) {
	r := box2d.CreateRecording(0)
	defer r.Destroy()
	s.world.StartRecording(r)
	defer s.world.StopRecording()
	bd := box2d.DefaultBodyDef()
	bd.Type, bd.Position = box2d.DynamicBody, differentialPosition(x, y)
	transient := s.world.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	sd.EnableContactEvents, sd.EnableHitEvents = true, true
	circle := box2d.Circle{Radius: 0.375}
	first := transient.CreateCircleShape(&sd, &circle)
	current := first
	var transientJoint box2d.Joint
	if s.bodies[0].IsValid() {
		jd := box2d.DefaultDistanceJointDef()
		jd.Base.BodyIDA, jd.Base.BodyIDB, jd.Base.CollideConnected = s.bodies[0], transient, true
		transientJoint = s.world.CreateDistanceJoint(&jd).Joint
	}
	for i := range 3 {
		if i == 1 {
			current.Destroy(true)
			current = transient.CreateCircleShape(&sd, &circle)
		}
		if i == 2 {
			transient.Destroy()
		}
		if b := s.bodies[i]; b.IsValid() {
			b.SetLinearVelocity(box2d.Vec2{X: differentialScalar(x), Y: differentialScalar(y)})
		}
		s.query(s.world, x, y)
		s.world.Step(1.0/60.0, 4)
		s.r.I32("record.step", int32(i))
		s.state(s.world)
		s.body(transient)
		s.r.U64("record.first.shape", differentialID(first.Store()))
		s.r.Bool("record.first.shape.valid", first.IsValid())
		s.r.U64("record.current.shape", differentialID(current.Store()))
		s.r.Bool("record.current.shape.valid", current.IsValid())
		s.r.U64("record.joint", differentialID(transientJoint.Store()))
		s.r.Bool("record.joint.valid", transientJoint.IsValid())
		for _, b := range s.bodies {
			s.body(b)
		}
		s.r.U64("world.hash", s.world.StateHash())
	}
	s.world.StopRecording()
	p, err := box2d.CreateReplay(r.Data(), 0)
	s.r.Bool("replay.created", err == nil && p != nil)
	if err != nil {
		s.t.Fatalf("native Go recording cannot replay: %v", err)
	}
	defer p.Destroy()
	info := p.Info()
	s.r.I32("replay.frames", int32(info.FrameCount))
	s.r.I32("replay.workers", int32(info.WorkerCount))
	s.r.F32("replay.dt", info.TimeStep)
	s.r.I32("replay.substeps", int32(info.SubStepCount))
	s.replay(p)
	for _, target := range [...]int{3, 1, 2, 0, 3} {
		p.SeekFrame(target)
		s.replay(p)
		s.r.Bool("replay.step", p.StepFrame())
		s.replay(p)
	}
	p.SeekFrame(1)
	rw := p.WorldID()
	snap := rw.Snapshot()
	for i := range 2 {
		err := rw.Restore(snap)
		s.r.Bool("replay.restore", err == nil)
		if err != nil {
			s.t.Fatalf("restoring native Go replay snapshot: %v", err)
		}
		s.query(rw, x, y)
		rw.Step(1.0/60.0, 4)
		s.r.I32("replay.restep", int32(i))
		s.state(rw)
		for j := range p.BodyCount() {
			s.body(p.BodyID(j))
		}
		s.r.U64("world.hash", rw.StateHash())
	}
	for i := range p.BodyCount() {
		if b := p.BodyID(i); b.IsValid() {
			b.ApplyForceToCenter(box2d.Vec2{X: differentialScalar(y), Y: differentialScalar(x)}, true)
		}
	}
	rw.Step(1.0/60.0, 4)
	s.state(rw)
	for j := range p.BodyCount() {
		s.body(p.BodyID(j))
	}
	s.r.U64("world.hash", rw.StateHash())
}
func (s *differentialSequence) execute(input []byte) {
	s.world = differentialWorld()
	s.initialGeneration, s.initialSlot = uint16(s.world.Store()), uint16(s.world.Store()>>16)
	defer func() {
		if s.world.IsValid() {
			s.world.Destroy()
		}
	}()
	s.r.Bytes("input", input)
	s.observe()
	for offset := 0; offset < len(input); offset += 4 {
		var op [4]byte
		copy(op[:], input[offset:min(offset+4, len(input))])
		code, slot, joint := op[0]%21, int(op[1]%4), int(op[1]%2)
		s.t.Logf("operation %d: opcode=%d slot=%d operands=%02x,%02x", offset/4, code, slot, op[2], op[3])
		b, sh := s.bodies[slot], s.shapes[slot]
		bv, sv, acted := b.IsValid(), sh.IsValid(), false
		s.r.U32("operation.index", uint32(offset/4))
		s.r.U32("operation.code", uint32(code))
		switch code {
		case 0:
			if !bv {
				d := box2d.DefaultBodyDef()
				d.Type, d.Position = box2d.DynamicBody, differentialPosition(op[2], op[3])
				s.bodies[slot] = s.world.CreateBody(&d)
				s.oldBodies = append(s.oldBodies, s.bodies[slot])
				acted = true
			}
		case 1:
			if bv {
				b.Destroy()
				acted = true
			}
		case 2:
			if bv && !sv {
				d := box2d.DefaultShapeDef()
				d.EnableContactEvents, d.EnableHitEvents = true, true
				c := box2d.Circle{Radius: 0.25 + float32(float32(op[2]%5)*0.125)}
				s.shapes[slot] = b.CreateCircleShape(&d, &c)
				s.oldShapes = append(s.oldShapes, s.shapes[slot])
				acted = true
			}
		case 3:
			if sv {
				sh.Destroy(true)
				acted = true
			}
		case 4:
			if bv {
				b.SetTransform(differentialPosition(op[2], op[3]), box2d.MakeRot(float32(op[2]%9)*0.125))
				acted = true
			}
		case 5:
			if bv {
				b.SetLinearVelocity(box2d.Vec2{X: differentialScalar(op[2]), Y: differentialScalar(op[3])})
				acted = true
			}
		case 6:
			if bv {
				b.SetAngularVelocity(differentialScalar(op[2]))
				acted = true
			}
		case 7:
			if bv {
				b.ApplyForceToCenter(box2d.Vec2{X: differentialScalar(op[2]), Y: differentialScalar(op[3])}, true)
				acted = true
			}
		case 8:
			if sv {
				f := box2d.DefaultFilter()
				f.CategoryBits, f.MaskBits = uint64(1)<<(op[2]%4), uint64(op[3]&15)
				sh.SetFilter(f)
				acted = true
			}
		case 9:
			if bv {
				b.SetType(box2d.BodyType(op[2] % 3))
				acted = true
			}
		case 10:
			if bv {
				b.Disable()
				acted = true
			}
		case 11:
			if bv {
				b.Enable()
				acted = true
			}
		case 12:
			dt := float32(1.0 / 60.0)
			if op[2]&1 != 0 {
				dt = 0
			}
			s.world.Step(dt, 1+int(op[3]%4))
			acted = true
		case 13:
			s.query(s.world, op[2], op[3])
			acted = true
		case 14:
			other := int(op[2] % 4)
			if !s.joints[joint].IsValid() && bv && other != slot && s.bodies[other].IsValid() &&
				(b.Type() == box2d.DynamicBody || s.bodies[other].Type() == box2d.DynamicBody) {
				d := box2d.DefaultDistanceJointDef()
				d.Base.BodyIDA, d.Base.BodyIDB = b, s.bodies[other]
				d.Length, d.Base.CollideConnected = 0.5+float32(float32(op[3]%8)*0.25), true
				s.joints[joint] = s.world.CreateDistanceJoint(&d).Joint
				s.oldJoints = append(s.oldJoints, s.joints[joint])
				acted = true
			}
		case 15:
			if s.joints[joint].IsValid() {
				s.joints[joint].Destroy()
				acted = true
			}
		case 16:
			s.image = s.world.Snapshot()
			s.savedBodies, s.savedShapes, s.savedJoints = s.bodies, s.shapes, s.joints
			acted = true
		case 17:
			if s.image != nil {
				err := s.world.Restore(s.image)
				s.r.Bool("snapshot.restore", err == nil)
				if err != nil {
					s.t.Fatalf("restoring native Go snapshot: %v", err)
				}
				s.bodies, s.shapes, s.joints = s.savedBodies, s.savedShapes, s.savedJoints
				acted = true
			}
		case 18:
			if bv {
				b.ApplyLinearImpulseToCenter(box2d.Vec2{X: differentialScalar(op[2]), Y: differentialScalar(op[3])}, true)
				acted = true
			}
		case 19:
			s.recording(op[2], op[3])
			acted = true
		case 20:
			old := s.world
			old.Destroy()
			s.r.Bool("world.destroyed", !old.IsValid())
			s.world = differentialWorld()
			s.r.Bool("world.old.valid", old.IsValid())
			s.epoch++
			s.image = nil
			s.bodies, s.shapes, s.joints = [4]box2d.Body{}, [4]box2d.Shape{}, [2]box2d.Joint{}
			s.oldBodies, s.oldShapes, s.oldJoints = nil, nil, nil
			acted = true
		}
		s.r.Bool("operation.acted", acted)
		s.observe()
	}
	s.r.U32("operation.count", uint32((len(input)+3)/4))
	s.world.Destroy()
	s.r.Bool("world.final.invalid", !s.world.IsValid())
}

// Corpus returns the eight deterministic histories also authored in the C
// differential_corpus runner. Each call returns newly owned input slices.
func Corpus() []Seed {
	maximum := make([]byte, InputLimit)
	for i := range 64 {
		copy(maximum[4*i:], []byte{byte(i % 21), 255, 255, 255})
	}
	for i := range 4 {
		copy(maximum[8*i:], []byte{0, byte(i), 255, 255, 2, byte(i), 255, 255})
	}
	return []Seed{
		{"empty", nil},
		{"zero_partial", []byte{0}},
		{"maximum_operands", maximum},
		{"lifecycle_reuse", []byte{0, 0, 0, 0, 2, 0, 0, 0, 0, 1, 24, 0, 2, 1, 2, 0, 14, 0, 1, 3, 15, 0, 0, 0, 14, 0, 1, 3, 3, 0, 0, 0, 2, 0, 3, 0, 1, 1, 0, 0, 0, 1, 24, 0, 2, 1, 2, 0, 14, 1, 0, 3, 12, 0, 0, 3, 13, 0, 0, 0, 20, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0}},
		{"contact_transitions", []byte{0, 0, 0, 0, 2, 0, 4, 0, 9, 0, 0, 0, 0, 1, 12, 0, 2, 1, 4, 0, 12, 0, 0, 3, 12, 0, 0, 3, 8, 1, 1, 0, 12, 0, 0, 3, 8, 1, 0, 15, 12, 0, 0, 3, 10, 1, 0, 0, 12, 0, 0, 3, 11, 1, 0, 0, 4, 1, 12, 0, 12, 0, 0, 3, 9, 1, 1, 0, 12, 0, 0, 3, 9, 1, 2, 0, 12, 0, 0, 3, 1, 0, 0, 0, 12, 0, 0, 3}},
		{"snapshot_continuation", []byte{0, 0, 0, 0, 2, 0, 4, 0, 9, 0, 0, 0, 0, 1, 12, 0, 2, 1, 4, 0, 12, 0, 0, 3, 16, 0, 0, 0, 10, 1, 0, 0, 0, 2, 32, 32, 2, 2, 2, 0, 14, 1, 2, 3, 12, 0, 0, 3, 17, 0, 0, 0, 12, 0, 0, 3, 4, 1, 48, 0, 13, 0, 0, 0, 12, 0, 0, 3}},
		{"replay_seek_restep", []byte{0, 0, 0, 0, 2, 0, 4, 0, 9, 0, 0, 0, 0, 1, 12, 0, 2, 1, 4, 0, 0, 2, 24, 24, 2, 2, 2, 0, 14, 1, 2, 3, 12, 0, 0, 3, 19, 0, 16, 240, 13, 0, 0, 0, 12, 0, 0, 3}},
		{"force_velocity_zero_step", []byte{0, 0, 128, 127, 2, 0, 4, 0, 5, 0, 255, 127, 6, 0, 128, 0, 7, 0, 127, 128, 12, 0, 1, 3, 18, 0, 127, 128, 12, 0, 0, 3, 13, 0, 128, 127}},
	}
}
