// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"encoding/binary"
	"math"
	"slices"
	"unsafe"
)

type SceneBodyTransformForTest struct {
	Index     int
	Transform WorldTransform
}

func SceneBodyTransformsForTest(worldID *World) []SceneBodyTransformForTest {
	world := worldID

	var transforms []SceneBodyTransformForTest

	for index := range world.bodies {
		body := &world.bodies[index]
		if body.setIndex == nullIndex {
			continue
		}

		transforms = append(transforms, SceneBodyTransformForTest{Index: index, Transform: getBodyTransformQuick(world, body)})
	}

	return transforms
}

type SceneObservationForTest struct {
	Count int
	Hash  uint64
}

type sceneDigest struct{ hash uint64 }

func newSceneDigest() sceneDigest { return sceneDigest{14695981039346656037} }
func (d *sceneDigest) data(p []byte) {
	for _, b := range p {
		d.hash ^= uint64(b)
		d.hash *= 1099511628211
	}
}

func (d *sceneDigest) u32(v uint32) {
	var p [4]byte
	binary.LittleEndian.PutUint32(p[:], v)
	d.data(p[:])
}

// Identity encoding retains slot and generation; only the world routing slot is omitted.
func (d *sceneDigest) id(v uint64) { d.u32(uint32(v >> 32)); d.u32(uint32(v & 0xffff)) }
func (d *sceneDigest) f(v float32) { d.u32(math.Float32bits(v)) }
func (d *sceneDigest) vec(v Vec2)  { d.f(v.X); d.f(v.Y) }
func (d *sceneDigest) flag(v bool) {
	if v {
		d.u32(1)
	} else {
		d.u32(0)
	}
}
func (d *sceneDigest) frame(v Transform) { d.vec(v.P); d.f(v.Q.C); d.f(v.Q.S) }
func (d *sceneDigest) manifold(m Manifold, hitEvents bool) {
	d.flag(hitEvents)
	d.vec(m.Normal)
	d.f(m.RollingImpulse)
	d.u32(uint32(m.PointCount))

	for _, p := range m.Points[:m.PointCount] {
		d.vec(p.AnchorA)
		d.vec(p.AnchorB)
		d.f(p.Separation)
		d.f(p.NormalImpulse)
		d.f(p.TangentImpulse)
		d.f(p.TotalNormalImpulse)
		d.f(p.BaseSeparation)
		d.u32(uint32(p.ID))
		d.flag(p.Persisted)
		// NormalVelocity is specified only for contacts with enabled hit events.
		if hitEvents {
			d.f(p.NormalVelocity)
		}
	}
}

// Each lane is a canonical, explicitly encoded public outcome, not a native image.
func SceneStateForTest(w *World) [3]SceneObservationForTest {
	world := w

	var result [3]SceneObservationForTest

	d := newSceneDigest()

	for i := range world.bodies {
		b := &world.bodies[i]
		if b.setIndex == nullIndex {
			continue
		}

		id := makeBodyID(world, i)
		d.id(id.ID())
		d.u32(uint32(id.Type()))
		d.vec(id.LinearVelocity())
		d.f(id.AngularVelocity())
		d.flag(id.Awake())
		d.flag(id.Enabled())
		mass := id.MassData()
		d.f(mass.Mass)
		d.vec(mass.Center)
		d.f(mass.RotationalInertia)

		result[0].Count++
	}

	result[0].Hash = d.hash
	d = newSceneDigest()

	for i := range world.contacts {
		c := &world.contacts[i]
		if c.contactID == nullIndex {
			continue
		}

		d.u32(uint32(i + 1))
		d.u32(c.generation)
		a, b := &world.shapes[c.shapeIDA], &world.shapes[c.shapeIDB]
		d.u32(uint32(c.shapeIDA + 1))
		d.u32(uint32(a.generation))
		d.u32(uint32(c.shapeIDB + 1))
		d.u32(uint32(b.generation))
		d.manifold(getContactSim(world, c).manifold, a.enableHitEvents || b.enableHitEvents)

		result[1].Count++
	}

	result[1].Hash = d.hash
	d = newSceneDigest()

	for i := range world.joints {
		j := &world.joints[i]
		if j.setIndex == nullIndex {
			continue
		}

		id := Joint{int32(i + 1), world, j.generation}
		d.id(id.ID())
		d.u32(uint32(id.Type()))
		d.id(id.BodyA().ID())
		d.id(id.BodyB().ID())
		d.frame(id.LocalFrameA())
		d.frame(id.LocalFrameB())
		d.vec(id.ConstraintForce())
		d.f(id.ConstraintTorque())

		if id.Type() == RevoluteJointType {
			r := RevoluteJoint{Joint: id}
			d.flag(r.MotorEnabled())
			d.f(r.MotorSpeed())
			d.f(r.MaxMotorTorque())
			d.f(r.MotorTorque())
			d.flag(r.SpringEnabled())
			d.f(r.SpringHertz())
			d.f(r.SpringDampingRatio())
			d.flag(r.LimitEnabled())
			d.f(r.LowerLimit())
			d.f(r.UpperLimit())
		}

		result[2].Count++
	}

	result[2].Hash = d.hash

	return result
}

func SceneEventsForTest(w *World) [7]SceneObservationForTest {
	var lanes [7][][]byte

	u := func(p []byte, v uint32) []byte { return binary.LittleEndian.AppendUint32(p, v) }
	id := func(p []byte, v uint64) []byte { p = u(p, uint32(v>>32)); return u(p, uint32(v&0xffff)) }
	contact := func(p []byte, v Contact) []byte { p = u(p, uint32(v.index1)); return u(p, v.generation) }
	f := func(p []byte, v float32) []byte { return u(p, math.Float32bits(v)) }
	vec := func(p []byte, v Vec2) []byte { p = f(p, v.X); return f(p, v.Y) }

	pos := func(p []byte, v Pos) []byte {
		// Pos precision is selected by the build; conversion preserves its exact bits.
		if unsafe.Sizeof(Pos{}) == 16 {
			p = binary.LittleEndian.AppendUint64(p, math.Float64bits(float64(v.X)))
			return binary.LittleEndian.AppendUint64(p, math.Float64bits(float64(v.Y)))
		}

		p = f(p, float32(v.X))

		return f(p, float32(v.Y))
	}
	for e := range w.BodyMoveEvents() {
		p := id(nil, e.Body.ID())
		p = pos(p, e.Transform.P)
		p = f(p, e.Transform.Q.C)
		p = f(p, e.Transform.Q.S)

		v := uint32(0)
		if e.FellAsleep {
			v = 1
		}

		p = u(p, v)

		v = 0
		if e.UserData == nil {
			v = 1
		}

		lanes[0] = append(lanes[0], u(p, v))
	}

	for e := range w.SensorBeginEvents() {
		lanes[1] = append(lanes[1], id(id(nil, e.SensorShape.ID()), e.VisitorShape.ID()))
	}

	for e := range w.SensorEndEvents() {
		lanes[2] = append(lanes[2], id(id(nil, e.SensorShape.ID()), e.VisitorShape.ID()))
	}

	for e := range w.ContactBeginEvents() {
		lanes[3] = append(lanes[3], contact(id(id(nil, e.ShapeA.ID()), e.ShapeB.ID()), e.Contact))
	}

	for e := range w.ContactEndEvents() {
		lanes[4] = append(lanes[4], contact(id(id(nil, e.ShapeA.ID()), e.ShapeB.ID()), e.Contact))
	}

	for e := range w.ContactHitEvents() {
		p := contact(id(id(nil, e.ShapeA.ID()), e.ShapeB.ID()), e.Contact)
		p = pos(p, e.Point)
		p = vec(p, e.Normal)
		lanes[5] = append(lanes[5], f(p, e.ApproachSpeed))
	}

	for e := range w.JointEvents() {
		p := id(nil, e.Joint.ID())

		v := uint32(0)
		if e.UserData == nil {
			v = 1
		}

		lanes[6] = append(lanes[6], u(p, v))
	}

	return sceneEventDigests(&lanes)
}

func sceneEventDigests(lanes *[7][][]byte) [7]SceneObservationForTest {
	var result [7]SceneObservationForTest

	for i, records := range lanes {
		slices.SortFunc(records, bytes.Compare)

		d := newSceneDigest()
		for _, p := range records {
			d.data(p)
		}

		result[i] = SceneObservationForTest{len(records), d.hash}
	}

	return result
}
