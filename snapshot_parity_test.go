// SPDX-License-Identifier: MIT
package box2d

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func snapshotVec2(r *oracle.Reader, key string, v Vec2) { r.Vec2(key, v.X, v.Y) }
func snapshotPos(r *oracle.Reader, key string, v Pos)   { r.Pos(key, v.X, v.Y) }

// Native images never cross this boundary: only typed public observations do.
func snapshotEvents(r *oracle.Reader, w World, key string) {
	r.U64(key+"/state", w.StateHash())
	var sb []SensorBeginTouchEvent
	for e := range w.SensorBeginEvents() {
		sb = append(sb, e)
	}
	r.I32(key+"/sensorBegin", int32(len(sb)))
	for i, e := range sb {
		p := fmt.Sprintf("%s/sensorBegin/%d", key, i)
		r.I32(p+"/sensor", e.SensorShapeID.index1)
		r.U32(p+"/sensorGeneration", uint32(e.SensorShapeID.generation))
		r.I32(p+"/visitor", e.VisitorShapeID.index1)
		r.U32(p+"/visitorGeneration", uint32(e.VisitorShapeID.generation))
	}
	var se []SensorEndTouchEvent
	for e := range w.SensorEndEvents() {
		se = append(se, e)
	}
	r.I32(key+"/sensorEnd", int32(len(se)))
	for i, e := range se {
		p := fmt.Sprintf("%s/sensorEnd/%d", key, i)
		r.I32(p+"/sensor", e.SensorShapeID.index1)
		r.U32(p+"/sensorGeneration", uint32(e.SensorShapeID.generation))
		r.I32(p+"/visitor", e.VisitorShapeID.index1)
		r.U32(p+"/visitorGeneration", uint32(e.VisitorShapeID.generation))
	}
	var cb []ContactBeginTouchEvent
	for e := range w.ContactBeginEvents() {
		cb = append(cb, e)
	}
	r.I32(key+"/contactBegin", int32(len(cb)))
	for i, e := range cb {
		p := fmt.Sprintf("%s/contactBegin/%d", key, i)
		r.I32(p+"/a", e.ShapeIDA.index1)
		r.U32(p+"/aGeneration", uint32(e.ShapeIDA.generation))
		r.I32(p+"/b", e.ShapeIDB.index1)
		r.U32(p+"/bGeneration", uint32(e.ShapeIDB.generation))
		r.I32(p+"/contact", e.ContactID.index1)
		r.U32(p+"/contactGeneration", e.ContactID.generation)
	}
	var ce []ContactEndTouchEvent
	for e := range w.ContactEndEvents() {
		ce = append(ce, e)
	}
	r.I32(key+"/contactEnd", int32(len(ce)))
	for i, e := range ce {
		p := fmt.Sprintf("%s/contactEnd/%d", key, i)
		r.I32(p+"/a", e.ShapeIDA.index1)
		r.U32(p+"/aGeneration", uint32(e.ShapeIDA.generation))
		r.I32(p+"/b", e.ShapeIDB.index1)
		r.U32(p+"/bGeneration", uint32(e.ShapeIDB.generation))
		r.I32(p+"/contact", e.ContactID.index1)
		r.U32(p+"/contactGeneration", e.ContactID.generation)
	}
	var ch []ContactHitEvent
	for e := range w.ContactHitEvents() {
		ch = append(ch, e)
	}
	r.I32(key+"/contactHit", int32(len(ch)))
	for i, e := range ch {
		p := fmt.Sprintf("%s/contactHit/%d", key, i)
		r.I32(p+"/a", e.ShapeIDA.index1)
		r.I32(p+"/b", e.ShapeIDB.index1)
		r.U32(p+"/aGeneration", uint32(e.ShapeIDA.generation))
		r.U32(p+"/bGeneration", uint32(e.ShapeIDB.generation))
		r.I32(p+"/contact", e.ContactID.index1)
		r.U32(p+"/contactGeneration", e.ContactID.generation)
		snapshotPos(r, p+"/point", e.Point)
		snapshotVec2(r, p+"/normal", e.Normal)
		r.F32(p+"/speed", e.ApproachSpeed)
	}
	var je []JointEvent
	for e := range w.JointEvents() {
		je = append(je, e)
	}
	r.I32(key+"/jointEvents", int32(len(je)))
	for i, e := range je {
		p := fmt.Sprintf("%s/jointEvent/%d", key, i)
		r.I32(p+"/id", e.JointID.index1)
		r.U32(p+"/generation", uint32(e.JointID.generation))
		r.Bool(p+"/userDataNil", e.UserData == nil)
	}
}

func snapshotQuery(t *testing.T, r *oracle.Reader, w World, origin Pos, key string) {
	t.Helper()
	var mask uint64
	var generations [64]uint32
	count := 0
	w.OverlapAABB(origin, AABB{LowerBound: Vec2{-30, -5}, UpperBound: Vec2{30, 15}}, DefaultQueryFilter(), func(s Shape) bool {
		if s.index1 <= 0 || s.index1 >= 64 {
			t.Fatalf("%s query shape index %d outside bounded observation", key, s.index1)
		}
		mask |= uint64(1) << s.index1
		count++
		generations[s.index1] = uint32(s.generation)
		return true
	})
	r.I32(key+"/queryCount", int32(count))
	r.U64(key+"/querySet", mask)
	for i := range 64 {
		if mask&(uint64(1)<<i) != 0 {
			r.U32(fmt.Sprintf("%s/query/%d/generation", key, i), generations[i])
		}
	}
}

func TestSnapshotFamiliesRollback(t *testing.T) {
	cases := []struct {
		name   string
		origin Pos
	}{{"Families", Pos{}}}
	if IsDoublePrecision() {
		cases = append(cases, struct {
			name   string
			origin Pos
		}{"DistantFamilies", Pos{1e7, -1e7}})
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := oracle.Open(t, "snapshot/"+tc.name)
			origin := tc.origin
			wd := DefaultWorldDef()
			wd.Gravity = Vec2{0, -2}
			w := CreateWorld(&wd)
			BTWorldCleanup(t, w)
			bd := DefaultBodyDef()
			bd.Position = origin
			ground := w.CreateBody(&bd)
			sd := DefaultShapeDef()
			sd.EnableContactEvents = true
			sd.EnableHitEvents = true
			sd.EnableSensorEvents = true
			segment := Segment{Point1: Vec2{-25, 0}, Point2: Vec2{25, 0}}
			edge := ground.CreateSegmentShape(&sd, &segment)
			var bodies [5]Body
			var shapes [5]Shape
			var joints [5]Joint
			for i := range 5 {
				bd = DefaultBodyDef()
				bd.Type = DynamicBody
				bd.Position = Pos{origin.X + oracle.Real(3*i), origin.Y + 3}
				bodies[i] = w.CreateBody(&bd)
				if i%2 == 0 {
					c := Circle{Center: Vec2{0.125, 0}, Radius: 0.4}
					shapes[i] = bodies[i].CreateCircleShape(&sd, &c)
				} else {
					c := Capsule{Center1: Vec2{-0.3, 0}, Center2: Vec2{0.3, 0}, Radius: 0.25}
					shapes[i] = bodies[i].CreateCapsuleShape(&sd, &c)
				}
				base := DefaultFilterJointDef().Base
				base.BodyIDA = ground
				base.BodyIDB = bodies[i]
				base.LocalFrameA.P = Vec2{float32(3 * i), 3}
				base.ForceThreshold = 0.5
				base.TorqueThreshold = 0.5
				switch i {
				case 0:
					d := DefaultWheelJointDef()
					d.Base = base
					d.EnableMotor = true
					d.MotorSpeed = 2
					d.MaxMotorTorque = 3
					d.EnableLimit = true
					d.LowerTranslation = -1
					d.UpperTranslation = 1
					joints[i] = w.CreateWheelJoint(&d).Joint
				case 1:
					d := DefaultMotorJointDef()
					d.Base = base
					d.LinearVelocity = Vec2{0.5, -0.25}
					d.MaxVelocityForce = 4
					d.AngularVelocity = 1
					d.MaxVelocityTorque = 2
					d.LinearHertz = 2
					d.LinearDampingRatio = 0.7
					d.MaxSpringForce = 5
					d.AngularHertz = 1
					d.AngularDampingRatio = 0.5
					d.MaxSpringTorque = 3
					joints[i] = w.CreateMotorJoint(&d).Joint
				case 2:
					d := DefaultMoverJointDef()
					d.Base = base
					d.LinearVelocity = Vec2{1, -0.5}
					d.MaxVelocityForce = Vec2{2, 3}
					joints[i] = w.CreateMoverJoint(&d).Joint
				case 3:
					d := DefaultPogoJointDef()
					d.Base = base
					d.Normal = Vec2{0, 1}
					d.Hertz = 3
					d.DampingRatio = 0.6
					d.RestLength = 0.5
					d.MaxTensionForce = 2
					d.MaxCompressionForce = 6
					d.Impulse = 0.125
					d.Velocity = 0.25
					joints[i] = w.CreatePogoJoint(&d).Joint
				case 4:
					d := DefaultFilterJointDef()
					d.Base = base
					joints[i] = w.CreateFilterJoint(&d).Joint
				}
			}
			bd = DefaultBodyDef()
			bd.Type = DynamicBody
			bd.Position = Pos{origin.X - 10, origin.Y + 5}
			bd.IsEnabled = false
			disabled := w.CreateBody(&bd)
			circle := Circle{Radius: 0.5}
			disabled.CreateCircleShape(&sd, &circle)
			bd.Position.X = origin.X - 15
			bd.IsEnabled = true
			bd.IsAwake = false
			sleeping := w.CreateBody(&bd)
			sleeping.CreateCircleShape(&sd, &circle)
			bd = DefaultBodyDef()
			bd.Position = Pos{origin.X - 5, origin.Y + 4}
			sensorBody := w.CreateBody(&bd)
			sensorDef := DefaultShapeDef()
			sensorDef.IsSensor = true
			sensorDef.EnableSensorEvents = true
			circle.Radius = 1.25
			sensor := sensorBody.CreateCircleShape(&sensorDef, &circle)
			bd.Type = DynamicBody
			bd.Position.X = origin.X - 8
			bd.LinearVelocity = Vec2{4, 0}
			bd.GravityScale = 0
			visitor := w.CreateBody(&bd)
			circle.Radius = 0.25
			visitor.CreateCircleShape(&sd, &circle)
			for range 20 {
				w.Step(float32(1.0)/60, 4)
			}
			image := w.Snapshot()
			initial := w.StateHash()
			var tail [90]uint64
			r.U64("capture/state", initial)
			r.Bool("capture/disabled", disabled.IsEnabled())
			r.Bool("capture/sleeping", sleeping.IsAwake())
			for i := range 90 {
				if i == 5 {
					disabled.Enable()
				}
				if i == 9 {
					sleeping.SetAwake(true)
				}
				w.Step(float32(1.0)/60, 4)
				tail[i] = w.StateHash()
				key := fmt.Sprintf("reference/%d", i)
				snapshotEvents(r, w, key)
				snapshotQuery(t, r, w, origin, key)
			}
			disabled.Enable()
			sleeping.SetAwake(true)
			visitor.SetLinearVelocity(Vec2{-10, 2})
			WheelJoint{joints[0]}.SetMotorSpeed(-7)
			MotorJoint{joints[1]}.SetLinearVelocity(Vec2{-4, 5})
			MoverJoint{joints[2]}.SetLinearVelocity(Vec2{-3, 2})
			PogoJoint{joints[3]}.SetRestLength(2)
			shapes[0].SetCircle(&circle)
			joints[4].Destroy()
			changedCapsule := Capsule{Center1: Vec2{-1, 0}, Center2: Vec2{1, 0}, Radius: 0.75}
			shapes[1].SetCapsule(&changedCapsule)
			changedSegment := Segment{Point1: Vec2{-10, -1}, Point2: Vec2{10, -1}}
			edge.SetSegment(&changedSegment)
			visitor.SetTransform(Pos{origin.X - 5, origin.Y + 4}, Rot{C: 1})
			w.Step(float32(1.0)/60, 4)
			visitor.Disable()
			w.Step(float32(1.0)/60, 4)
			internal := getWorldFromId(w)
			queuedSensorEnds := len(internal.sensorEndEvents[0]) + len(internal.sensorEndEvents[1])
			if queuedSensorEnds < 1 {
				t.Fatal("sensor corruption history did not queue a real end event")
			}
			r.I32("mutation/queuedSensorEnds", int32(queuedSensorEnds))
			if err := w.Restore(image); err != nil {
				t.Fatalf("restore family image: %v", err)
			}
			if w.StateHash() != initial {
				t.Fatalf("restore canonical state: got %x, capture %x", w.StateHash(), initial)
			}
			r.U64("restore/state", w.StateHash())
			r.I32("restore/sensorEnd0", int32(len(internal.sensorEndEvents[0])))
			r.I32("restore/sensorEnd1", int32(len(internal.sensorEndEvents[1])))
			if len(internal.sensorEndEvents[0]) != 0 || len(internal.sensorEndEvents[1]) != 0 {
				t.Fatal("restore retained queued sensor ends")
			}
			r.Bool("restore/disabled", disabled.IsEnabled())
			r.Bool("restore/sleeping", sleeping.IsAwake())
			r.F32("restore/wheelMotor", WheelJoint{joints[0]}.MotorSpeed())
			snapshotVec2(r, "restore/motorVelocity", MotorJoint{joints[1]}.LinearVelocity())
			snapshotVec2(r, "restore/moverVelocity", MoverJoint{joints[2]}.LinearVelocity())
			r.F32("restore/pogoLength", PogoJoint{joints[3]}.RestLength())
			for i := range 5 {
				r.I32(fmt.Sprintf("restore/joint/%d/type", i), int32(joints[i].Type()))
				r.I32(fmt.Sprintf("restore/shape/%d/type", i), int32(shapes[i].Type()))
			}
			c := shapes[0].Circle()
			snapshotVec2(r, "restore/circleCenter", c.Center)
			r.F32("restore/circleRadius", c.Radius)
			cap := shapes[1].Capsule()
			snapshotVec2(r, "restore/capsule1", cap.Center1)
			snapshotVec2(r, "restore/capsule2", cap.Center2)
			r.F32("restore/capsuleRadius", cap.Radius)
			segment = edge.Segment()
			snapshotVec2(r, "restore/segment1", segment.Point1)
			snapshotVec2(r, "restore/segment2", segment.Point2)
			r.I32("restore/sensorOverlaps", int32(sensor.SensorCapacity()))
			for i := range 90 {
				if i == 5 {
					disabled.Enable()
				}
				if i == 9 {
					sleeping.SetAwake(true)
				}
				w.Step(float32(1.0)/60, 4)
				if w.StateHash() != tail[i] {
					t.Fatalf("rollback canonical state at step %d: got %x reference %x", i, w.StateHash(), tail[i])
				}
				key := fmt.Sprintf("rollback/%d", i)
				snapshotEvents(r, w, key)
				snapshotQuery(t, r, w, origin, key)
			}
		})
	}
}

func TestSnapshotDestroyedReusedHandlesAndNewWorld(t *testing.T) {
	r := oracle.Open(t, "snapshot/Handles")
	wd := DefaultWorldDef()
	w := CreateWorld(&wd)
	BTWorldCleanup(t, w)
	bd := DefaultBodyDef()
	ground := w.CreateBody(&bd)
	bd.Type = DynamicBody
	bd.Position = Pos{0, 4}
	body := w.CreateBody(&bd)
	sd := DefaultShapeDef()
	c := Circle{Radius: 0.5}
	shape := body.CreateCircleShape(&sd, &c)
	jd := DefaultFilterJointDef()
	jd.Base.BodyIDA = ground
	jd.Base.BodyIDB = body
	joint := w.CreateFilterJoint(&jd).Joint
	mat := DefaultSurfaceMaterial()
	cd := DefaultChainDef()
	cd.Points = []Vec2{{-2, 0}, {0, 0}, {2, 0}}
	cd.PointCount = 3
	cd.Ghost1 = Vec2{-4, 0}
	cd.Ghost2 = Vec2{4, 0}
	cd.Materials = []SurfaceMaterial{mat}
	cd.MaterialCount = 1
	chain := ground.CreateChain(&cd)
	image := w.Snapshot()
	joint.Destroy()
	chain.Destroy()
	shape.Destroy(true)
	body.Destroy()
	replacement := w.CreateBody(&bd)
	replacementShape := replacement.CreateCircleShape(&sd, &c)
	jd.Base.BodyIDB = replacement
	replacementJoint := w.CreateFilterJoint(&jd).Joint
	replacementChain := ground.CreateChain(&cd)
	extra := w.CreateBody(&bd)
	extraShape := extra.CreateCircleShape(&sd, &c)
	jd.Base.BodyIDB = extra
	extraJoint := w.CreateFilterJoint(&jd).Joint
	extraChain := ground.CreateChain(&cd)
	r.Bool("reuse/bodySlot", body.index1 == replacement.index1)
	r.Bool("reuse/bodyGeneration", body.generation != replacement.generation)
	r.Bool("reuse/shapeSlot", shape.index1 == replacementShape.index1)
	r.Bool("reuse/shapeGeneration", shape.generation != replacementShape.generation)
	r.Bool("reuse/jointSlot", joint.index1 == replacementJoint.index1)
	r.Bool("reuse/jointGeneration", joint.generation != replacementJoint.generation)
	r.Bool("reuse/chainSlot", chain.index1 == replacementChain.index1)
	r.Bool("reuse/chainGeneration", chain.generation != replacementChain.generation)
	if err := w.Restore(image); err != nil {
		t.Fatalf("restore destroyed/reused handles: %v", err)
	}
	r.Bool("held/body", body.IsValid())
	r.Bool("held/shape", shape.IsValid())
	r.Bool("held/joint", joint.IsValid())
	r.Bool("held/chain", chain.IsValid())
	r.Bool("replacement/body", replacement.IsValid())
	r.Bool("replacement/shape", replacementShape.IsValid())
	r.Bool("replacement/joint", replacementJoint.IsValid())
	r.Bool("replacement/chain", replacementChain.IsValid())
	r.Bool("extra/body", extra.IsValid())
	r.Bool("extra/shape", extraShape.IsValid())
	r.Bool("extra/joint", extraJoint.IsValid())
	r.Bool("extra/chain", extraChain.IsValid())
	if !body.IsValid() || !shape.IsValid() || !joint.IsValid() || !chain.IsValid() {
		t.Fatal("snapshot-time handles must be revived before operation")
	}
	body.SetLinearVelocity(Vec2{2, 1})
	shape.SetDensity(2, true)
	mat.Friction = 0.75
	chain.SetAllSurfaceMaterials(&mat)
	joint.SetCollideConnected(true)
	w.Step(float32(1.0)/60, 4)
	snapshotPos(r, "held/position", body.Position())
	r.F32("held/density", shape.Density())
	r.F32("held/chainFriction", chain.SurfaceMaterial(0).Friction)
	r.Bool("held/collide", joint.CollideConnected())
	r.U64("held/state", w.StateHash())
	for _, workers := range []int{0, 4} {
		clone, err := CreateWorldFromSnapshot(image, workers)
		if err != nil {
			t.Fatalf("load workers %d: %v", workers, err)
		}
		BTWorldCleanup(t, clone)
		key := fmt.Sprintf("new/%d", workers)
		r.Bool(key+"/identity", clone.index1 != w.index1)
		r.I32(key+"/workers", int32(getWorldFromId(clone).workerCount))
		mappedBody := body
		mappedBody.world0 = clone.index1 - 1
		mappedShape := shape
		mappedShape.world0 = clone.index1 - 1
		mappedJoint := joint
		mappedJoint.world0 = clone.index1 - 1
		mappedChain := chain
		mappedChain.world0 = clone.index1 - 1
		r.Bool(key+"/bodyValid", mappedBody.IsValid())
		r.Bool(key+"/shapeBody", mappedShape.Body() == mappedBody)
		r.Bool(key+"/jointBody", mappedJoint.BodyB() == mappedBody)
		r.Bool(key+"/bodyWorld", mappedBody.World() == clone)
		r.Bool(key+"/shapeWorld", mappedShape.World() == clone)
		r.Bool(key+"/jointWorld", mappedJoint.World() == clone)
		r.Bool(key+"/chainWorld", mappedChain.World() == clone)
		r.Bool(key+"/originOwner", body.World() == w)
		clone.Step(float32(1.0)/60, 4)
		r.U64(key+"/state", clone.StateHash())
		clone.Destroy()
	}
}

func TestSnapshotHeaderRejectionAndLateCleanup(t *testing.T) {
	r := oracle.Open(t, "snapshot/Headers")
	wd := DefaultWorldDef()
	w := CreateWorld(&wd)
	BTWorldCleanup(t, w)
	bd := DefaultBodyDef()
	bd.Type = DynamicBody
	bd.Position = Pos{2, 4}
	bd.Name = "header-held"
	body := w.CreateBody(&bd)
	sd := DefaultShapeDef()
	sd.Material.Friction = 0.625
	c := Circle{Radius: 0.5}
	shape := body.CreateCircleShape(&sd, &c)
	staticDef := DefaultBodyDef()
	anchor := w.CreateBody(&staticDef)
	jd := DefaultFilterJointDef()
	jd.Base.BodyIDA = anchor
	jd.Base.BodyIDB = body
	joint := w.CreateFilterJoint(&jd).Joint
	material := DefaultSurfaceMaterial()
	material.Friction = 0.375
	cd := DefaultChainDef()
	cd.Points = []Vec2{{-2, 0}, {0, 0}, {2, 0}}
	cd.PointCount = 3
	cd.Ghost1 = Vec2{-4, 0}
	cd.Ghost2 = Vec2{4, 0}
	cd.Materials = []SurfaceMaterial{material}
	cd.MaterialCount = 1
	chain := anchor.CreateChain(&cd)
	image := w.Snapshot()
	hash := w.StateHash()
	cases := []struct {
		name     string
		offset   int
		category error
	}{{"nil", -2, nil}, {"short", -1, nil}, {"magic", 0, ErrBadMagic}, {"version", 4, ErrVersion}, {"layout", 8, ErrLayout}, {"precision", 12, ErrLayout}}
	for _, tc := range cases {
		bad := append([]byte(nil), image...)
		switch tc.offset {
		case -2:
			bad = nil
		case -1:
			bad = bad[:15]
		default:
			value := binary.LittleEndian.Uint32(bad[tc.offset:])
			bit := uint32(1)
			if tc.offset == 12 {
				bit = 2
			}
			binary.LittleEndian.PutUint32(bad[tc.offset:], value^bit)
		}
		err := w.Restore(bad)
		var snapshotErr *SnapshotError
		if err == nil || !errors.As(err, &snapshotErr) {
			t.Fatalf("%s restore error category: %v", tc.name, err)
		}
		if tc.category != nil && !errors.Is(err, tc.category) {
			t.Fatalf("%s restore: got %v want errors.Is(%v)", tc.name, err, tc.category)
		}
		if w.StateHash() != hash {
			t.Fatalf("%s early rejection changed canonical state", tc.name)
		}
		r.Bool(tc.name+"/rejected", err != nil)
		r.U64(tc.name+"/state", w.StateHash())
		r.Bool(tc.name+"/bodyValid", body.IsValid())
		r.Bool(tc.name+"/shapeValid", shape.IsValid())
		snapshotPos(r, tc.name+"/position", body.Position())
		r.Bool(tc.name+"/name", body.Name() == "header-held")
		r.F32(tc.name+"/friction", shape.Friction())
		r.Bool(tc.name+"/jointValid", joint.IsValid())
		r.Bool(tc.name+"/chainValid", chain.IsValid())
		r.Bool(tc.name+"/jointBody", joint.BodyB() == body)
		r.F32(tc.name+"/chainFriction", chain.SurfaceMaterial(0).Friction)
		snapshotVec2(r, tc.name+"/gravity", w.Gravity())
		invalid, err := CreateWorldFromSnapshot(bad, 0)
		r.Bool(tc.name+"/newRejected", !invalid.IsValid())
		if err == nil || !errors.As(err, &snapshotErr) {
			t.Fatalf("%s new-world error category: %v", tc.name, err)
		}
		if tc.category != nil && !errors.Is(err, tc.category) {
			t.Fatalf("%s new-world: got %v want %v", tc.name, err, tc.category)
		}
	}
	// Each engine truncates its own native image at the final payload boundary.
	err := w.Restore(image[:len(image)-1])
	var snapshotErr *SnapshotError
	r.Bool("late/rejected", err != nil)
	if err == nil || !errors.As(err, &snapshotErr) {
		t.Fatalf("late restore: %v", err)
	}
	w.Destroy()
	invalid, err := CreateWorldFromSnapshot(image[:len(image)-1], 0)
	r.Bool("late/newRejected", !invalid.IsValid())
	if err == nil || !errors.As(err, &snapshotErr) {
		t.Fatalf("late load: %v", err)
	}
	wd = DefaultWorldDef()
	fresh := CreateWorld(&wd)
	BTWorldCleanup(t, fresh)
	body = fresh.CreateBody(&bd)
	r.Bool("cleanup/newBody", body.IsValid())
}

func snapshotContactMaterials(r *oracle.Reader, id World, key string) {
	w := getWorldFromId(id)
	count := 0
	for i := range w.contacts {
		if w.contacts[i].contactID == i {
			count++
		}
	}
	r.I32(key+"/contactCount", int32(count))
	for i := range w.contacts {
		c := &w.contacts[i]
		if c.contactID != i {
			continue
		}
		sim := getContactSim(w, c)
		p := fmt.Sprintf("%s/contact/%d", key, i)
		r.I32(p+"/id", int32(i))
		r.F32(p+"/friction", sim.friction)
		r.F32(p+"/restitution", sim.restitution)
		r.I32(p+"/points", int32(sim.manifold.PointCount))
	}
}

func TestSnapshotRetainsLiveWiringAndClearsApplicationData(t *testing.T) {
	r := oracle.Open(t, "snapshot/Wiring")
	var oldTag, liveTag int
	var tasks, finishes, oldFilter, liveFilter, oldPre, livePre, oldFriction, liveFriction, oldRestitution, liveRestitution int
	wd := DefaultWorldDef()
	wd.WorkerCount = 2
	wd.UserData = &oldTag
	wd.EnqueueTask = func(task func()) func() { tasks++; task(); return func() { finishes++ } }
	wd.FrictionCallback = func(float32, uint64, float32, uint64) float32 { oldFriction++; return 0.9 }
	wd.RestitutionCallback = func(float32, uint64, float32, uint64) float32 { oldRestitution++; return 0.9 }
	w := CreateWorld(&wd)
	BTWorldCleanup(t, w)
	w.SetCustomFilterCallback(func(Shape, Shape) bool { oldFilter++; return false })
	w.SetPreSolveCallback(func(Shape, Shape, *Manifold) { oldPre++ }, nil)
	bd := DefaultBodyDef()
	bd.UserData = &oldTag
	ground := w.CreateBody(&bd)
	bd.Type = DynamicBody
	bd.Position = Pos{0, 0.75}
	body := w.CreateBody(&bd)
	sd := DefaultShapeDef()
	sd.UserData = &oldTag
	sd.EnableCustomFiltering = true
	sd.EnablePreSolveEvents = true
	sd.EnableContactEvents = true
	box := MakeBox(0.5, 0.5)
	ground.CreatePolygonShape(&sd, &box)
	shape := body.CreatePolygonShape(&sd, &box)
	jd := DefaultFilterJointDef()
	jd.Base.BodyIDA = ground
	jd.Base.BodyIDB = body
	jd.Base.CollideConnected = true
	jd.Base.UserData = &oldTag
	joint := w.CreateFilterJoint(&jd).Joint
	image := w.Snapshot()
	w.SetUserData(&liveTag)
	body.SetUserData(&liveTag)
	shape.SetUserData(&liveTag)
	joint.SetUserData(&liveTag)
	w.SetCustomFilterCallback(func(Shape, Shape) bool { liveFilter++; return true })
	w.SetPreSolveCallback(func(_ Shape, _ Shape, m *Manifold) { livePre++; m.PointCount = 0 }, nil)
	w.SetFrictionCallback(func(float32, uint64, float32, uint64) float32 { liveFriction++; return 0.125 })
	w.SetRestitutionCallback(func(float32, uint64, float32, uint64) float32 { liveRestitution++; return 0.375 })
	if err := w.Restore(image); err != nil {
		t.Fatalf("restore wired destination: %v", err)
	}
	r.Bool("restore/worldData", w.UserData() == &liveTag)
	r.Bool("restore/bodyData", body.UserData() == nil)
	r.Bool("restore/shapeData", shape.UserData() == nil)
	r.Bool("restore/jointData", joint.UserData() == nil)
	r.I32("restore/workers", int32(getWorldFromId(w).workerCount))
	w.Step(float32(1.0)/60, 4)
	if tasks == 0 || tasks != finishes || liveFilter == 0 || livePre == 0 || liveFriction == 0 || liveRestitution == 0 {
		t.Fatalf("restored wiring inactive: tasks %d finished %d filter %d pre %d friction %d restitution %d", tasks, finishes, liveFilter, livePre, liveFriction, liveRestitution)
	}
	r.Bool("live/tasks", tasks > 0)
	r.Bool("live/finished", tasks == finishes)
	r.I32("live/filter", int32(liveFilter))
	r.I32("live/pre", int32(livePre))
	r.I32("live/friction", int32(liveFriction))
	r.I32("live/restitution", int32(liveRestitution))
	r.I32("old/filter", int32(oldFilter))
	r.I32("old/pre", int32(oldPre))
	r.I32("old/friction", int32(oldFriction))
	r.I32("old/restitution", int32(oldRestitution))
	snapshotPos(r, "live/position", body.Position())
	r.U64("live/state", w.StateHash())
	snapshotContactMaterials(r, w, "live")
	beforeFilter, beforePre, beforeFriction, beforeRestitution, beforeTasks := liveFilter, livePre, liveFriction, liveRestitution, tasks
	clone, err := CreateWorldFromSnapshot(image, 0)
	if err != nil {
		t.Fatalf("load default wiring: %v", err)
	}
	BTWorldCleanup(t, clone)
	cb := body
	cb.world0 = clone.index1 - 1
	cs := shape
	cs.world0 = clone.index1 - 1
	cj := joint
	cj.world0 = clone.index1 - 1
	r.Bool("new/worldData", clone.UserData() == nil)
	r.Bool("new/bodyData", cb.UserData() == nil)
	r.Bool("new/shapeData", cs.UserData() == nil)
	r.Bool("new/jointData", cj.UserData() == nil)
	r.I32("new/workers", int32(getWorldFromId(clone).workerCount))
	clone.Step(float32(1.0)/60, 4)
	r.Bool("new/noLiveCalls", liveFilter == beforeFilter && livePre == beforePre && liveFriction == beforeFriction && liveRestitution == beforeRestitution && tasks == beforeTasks)
	snapshotPos(r, "new/position", cb.Position())
	r.U64("new/state", clone.StateHash())
	snapshotContactMaterials(r, clone, "new")
}
