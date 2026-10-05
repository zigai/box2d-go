// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

// These fixtures are produced by api_bodies.c calling the pinned C engine.
// Identity observations compare complete handles within each world, not masked IDs.
func apiVec(r *oracle.Reader, key string, v box2d.Vec2) { r.Vec2(key, v.X, v.Y) }
func apiPos(r *oracle.Reader, key string, v box2d.Pos)  { r.Pos(key, v.X, v.Y) }
func apiMass(r *oracle.Reader, key string, m box2d.MassData) {
	r.F32(key+"/mass", m.Mass)
	apiVec(r, key+"/center", m.Center)
	r.F32(key+"/inertia", m.RotationalInertia)
}

func apiBounds(r *oracle.Reader, key string, a box2d.AABB) {
	apiVec(r, key+"/lower", a.LowerBound)
	apiVec(r, key+"/upper", a.UpperBound)
}

func apiMaterial(r *oracle.Reader, key string, m box2d.SurfaceMaterial) {
	r.F32(key+"/friction", m.Friction)
	r.F32(key+"/restitution", m.Restitution)
	r.F32(key+"/rolling", m.RollingResistance)
	r.F32(key+"/tangent", m.TangentSpeed)
	r.U64(key+"/tag", m.UserMaterialID)
	r.U32(key+"/color", m.CustomColor)
}

func apiBody(r *oracle.Reader, key string, b box2d.Body) {
	apiPos(r, key+"/position", b.Position())
	q := b.Rotation()
	r.Rot(key+"/rotation", q.C, q.S)
	apiVec(r, key+"/velocity", b.LinearVelocity())
	r.F32(key+"/omega", b.AngularVelocity())
	r.I32(key+"/type", int32(b.Type()))
	r.Bool(key+"/awake", b.Awake())
	apiMass(r, key+"/massData", b.MassData())
	r.F32(key+"/mass", b.Mass())
	r.F32(key+"/inertia", b.RotationalInertia())
	apiVec(r, key+"/localCenter", b.LocalCenter())
	apiPos(r, key+"/worldCenter", b.WorldCenter())
	r.F32(key+"/linearDamping", b.LinearDamping())
	r.F32(key+"/angularDamping", b.AngularDamping())
	r.F32(key+"/gravity", b.GravityScale())
	locks := b.MotionLocks()
	r.Bool(key+"/lockX", locks.LinearX)
	r.Bool(key+"/lockY", locks.LinearY)
	r.Bool(key+"/lockZ", locks.AngularZ)
	r.I32(key+"/shapeCount", int32(b.ShapeCount()))

	n := 0
	for s := range b.Shapes() {
		p := fmt.Sprintf("%s/shape/%d", key, n)
		r.I32(p+"/tag", apiTag(s))
		r.Bool(p+"/owner", s.Body() == b)

		n++
	}

	r.I32(key+"/enumerated", int32(n))
	apiBounds(r, key+"/bounds", b.ComputeAABB())
}

func apiShape(r *oracle.Reader, key string, s box2d.Shape, b box2d.Body) {
	r.I32(key+"/type", int32(s.Type()))
	r.I32(key+"/user", apiTag(s))
	r.Bool(key+"/body", s.Body() == b)
	r.Bool(key+"/world", s.World() == b.World())
	r.Bool(key+"/orphan", s.ParentChain() == (box2d.Chain{}))

	switch s.Type() {
	case box2d.CircleShape:
		c := s.Circle()
		apiVec(r, key+"/circle/center", c.Center)
		r.F32(key+"/circle/radius", c.Radius)
	case box2d.CapsuleShape:
		c := s.Capsule()
		apiVec(r, key+"/capsule/a", c.Center1)
		apiVec(r, key+"/capsule/b", c.Center2)
		r.F32(key+"/capsule/radius", c.Radius)
	case box2d.SegmentShape:
		c := s.Segment()
		apiVec(r, key+"/segment/a", c.Point1)
		apiVec(r, key+"/segment/b", c.Point2)
	case box2d.PolygonShape:
		c := s.Polygon()
		r.I32(key+"/polygon/count", int32(c.Count))
		r.F32(key+"/polygon/radius", c.Radius)
		apiVec(r, key+"/polygon/center", c.Centroid)

		for i := range c.Count {
			apiVec(r, fmt.Sprintf("%s/polygon/vertex/%d", key, i), c.Vertices[i])
			apiVec(r, fmt.Sprintf("%s/polygon/normal/%d", key, i), c.Normals[i])
		}
	case box2d.ChainSegmentShape:
		c := s.ChainSegment()
		apiVec(r, key+"/chain/ghost1", c.Ghost1)
		apiVec(r, key+"/chain/a", c.Segment.Point1)
		apiVec(r, key+"/chain/b", c.Segment.Point2)
		apiVec(r, key+"/chain/ghost2", c.Ghost2)
	case box2d.ShapeTypeCount:
		// The shape count sentinel has no geometry payload.
	}

	apiMaterial(r, key+"/material", s.SurfaceMaterial())
	r.F32(key+"/density", s.Density())
	r.F32(key+"/friction", s.Friction())
	r.F32(key+"/restitution", s.Restitution())
	r.U64(key+"/materialTag", s.UserMaterial())
	f := s.Filter()
	r.U64(key+"/category", f.CategoryBits)
	r.U64(key+"/mask", f.MaskBits)
	r.I32(key+"/group", int32(f.GroupIndex))
	r.Bool(key+"/sensor", s.IsSensor())
	r.Bool(key+"/sensorEvents", s.SensorEventsEnabled())
	r.Bool(key+"/contactEvents", s.ContactEventsEnabled())
	r.Bool(key+"/hitEvents", s.HitEventsEnabled())
	r.Bool(key+"/preSolveEvents", s.PreSolveEventsEnabled())
	apiMass(r, key+"/massData", s.ComputeMassData())
	apiBounds(r, key+"/bounds", s.AABB())
}

func apiCast(r *oracle.Reader, key string, out box2d.WorldCastOutput) {
	r.Bool(key+"/hit", out.Hit)
	// Only a hit defines these payload fields.
	if out.Hit {
		apiPos(r, key+"/point", out.Point)
		apiVec(r, key+"/normal", out.Normal)
		r.F32(key+"/fraction", out.Fraction)
		r.I32(key+"/iterations", int32(out.Iterations))
	}
}

func apiTag(s box2d.Shape) int32 {
	if !s.IsValid() {
		return -1
	}

	tag, ok := s.UserData().(int)
	if !ok {
		panic("body parity fixture shape user data must be an int tag")
	}

	return int32(tag)
}

func apiContacts(r *oracle.Reader, key string, w *box2d.World, b box2d.Body) {
	count := 0
	for c := range b.ContactData() {
		p := fmt.Sprintf("%s/contact/%d", key, count)
		r.I32(p+"/a", apiTag(c.ShapeA))
		r.I32(p+"/b", apiTag(c.ShapeB))
		r.Bool(p+"/valid", c.Contact.IsValid())
		data := c.Contact.Data()
		r.Bool(p+"/identity", data.Contact == c.Contact && data.ShapeA == c.ShapeA && data.ShapeB == c.ShapeB)
		r.Bool(p+"/body", c.ShapeA.Body() == b || c.ShapeB.Body() == b)
		m := c.Manifold
		apiVec(r, p+"/normal", m.Normal)
		r.F32(p+"/rolling", m.RollingImpulse)
		r.I32(p+"/points", int32(m.PointCount))

		for i := range m.PointCount {
			q := fmt.Sprintf("%s/point/%d", p, i)
			v := m.Points[i]
			apiVec(r, q+"/anchorA", v.AnchorA)
			apiVec(r, q+"/anchorB", v.AnchorB)
			r.F32(q+"/separation", v.Separation)
			r.F32(q+"/normalImpulse", v.NormalImpulse)
			r.F32(q+"/tangentImpulse", v.TangentImpulse)
			r.F32(q+"/totalImpulse", v.TotalNormalImpulse)
			r.U32(q+"/id", uint32(v.ID))
			r.Bool(q+"/persisted", v.Persisted)
		}

		count++
	}

	r.I32(key+"/contacts", int32(count))

	beginCount, endCount, hitCount := 0, 0, 0
	for range w.ContactBeginEvents() {
		beginCount++
	}

	for range w.ContactEndEvents() {
		endCount++
	}

	for range w.ContactHitEvents() {
		hitCount++
	}

	r.I32(key+"/begin", int32(beginCount))
	r.I32(key+"/end", int32(endCount))
	r.I32(key+"/hits", int32(hitCount))

	i := 0
	for c := range w.ContactBeginEvents() {
		p := fmt.Sprintf("%s/begin/%d", key, i)
		r.I32(p+"/a", apiTag(c.ShapeA))
		r.I32(p+"/b", apiTag(c.ShapeB))
		r.Bool(p+"/valid", c.Contact.IsValid())

		i++
	}

	i = 0
	for c := range w.ContactEndEvents() {
		p := fmt.Sprintf("%s/end/%d", key, i)
		r.I32(p+"/a", apiTag(c.ShapeA))
		r.I32(p+"/b", apiTag(c.ShapeB))
		r.Bool(p+"/valid", c.Contact.IsValid())

		i++
	}

	i = 0
	for c := range w.ContactHitEvents() {
		p := fmt.Sprintf("%s/hit/%d", key, i)
		r.I32(p+"/a", apiTag(c.ShapeA))
		r.I32(p+"/b", apiTag(c.ShapeB))
		r.Bool(p+"/valid", c.Contact.IsValid())
		apiPos(r, p+"/point", c.Point)
		apiVec(r, p+"/normal", c.Normal)
		r.F32(p+"/speed", c.ApproachSpeed)

		i++
	}
}

func apiQuery(r *oracle.Reader, key string, w *box2d.World, base box2d.Pos, y float32, mask uint64) {
	q := box2d.QueryFilter{CategoryBits: 1, MaskBits: mask}
	n := 0

	w.OverlapAABB(base, box2d.AABB{LowerBound: box2d.Vec2{X: -.1, Y: y - .1}, UpperBound: box2d.Vec2{X: .1, Y: y + .1}}, q, func(s box2d.Shape) bool {
		r.I32(fmt.Sprintf("%s/tag/%d", key, n), apiTag(s))
		n++

		return true
	})
	r.I32(key+"/count", int32(n))
}

func apiWorld(t *testing.T, gravity box2d.Vec2) *box2d.World {
	t.Helper()

	wd := box2d.DefaultWorldDef()
	wd.Gravity = gravity
	wd.EnableContinuous = false
	w := box2d.NewWorld(&wd)
	box2d.BTWorldCleanup(t, w)

	return w
}

func apiBase(distant bool) box2d.Pos {
	if distant {
		return box2d.Pos{X: 0x1p30, Y: -0x1p29}
	}

	return box2d.Pos{X: 7, Y: -3}
}

func apiCreateGeometry(b box2d.Body, sd *box2d.ShapeDef, family int) box2d.Shape {
	switch family {
	case 0:
		g := box2d.Circle{Center: box2d.Vec2{}, Radius: 1}
		return b.CreateCircleShape(sd, &g)
	case 1:
		g := box2d.Capsule{Center1: box2d.Vec2{X: -1}, Center2: box2d.Vec2{X: 1}, Radius: .5}
		return b.CreateCapsuleShape(sd, &g)
	case 2:
		g := box2d.Segment{Point1: box2d.Vec2{X: -2}, Point2: box2d.Vec2{X: 2}}
		return b.CreateSegmentShape(sd, &g)
	case 3:
		g := box2d.MakeOffsetBox(2, .5, box2d.Vec2{}, box2d.Rot{C: 1})
		return b.CreatePolygonShape(sd, &g)
	default:
		g := box2d.ChainSegment{Ghost1: box2d.Vec2{X: -4}, Segment: box2d.Segment{Point1: box2d.Vec2{X: -2}, Point2: box2d.Vec2{X: 2}}, Ghost2: box2d.Vec2{X: 4}}
		return b.CreateChainSegmentShape(sd, &g)
	}
}

func apiSetGeometry(s box2d.Shape, family int, y float32) {
	switch family {
	case 0:
		g := box2d.Circle{Center: box2d.Vec2{Y: y}, Radius: 1.25}
		s.SetCircle(&g)
	case 1:
		g := box2d.Capsule{Center1: box2d.Vec2{X: -1.5, Y: y}, Center2: box2d.Vec2{X: 1.5, Y: y}, Radius: .75}
		s.SetCapsule(&g)
	case 2:
		g := box2d.Segment{Point1: box2d.Vec2{X: -3, Y: y}, Point2: box2d.Vec2{X: 3, Y: y}}
		s.SetSegment(&g)
	case 3:
		g := box2d.MakeOffsetRoundedBox(2.5, .75, box2d.Vec2{Y: y}, box2d.Rot{C: 1}, .1)
		s.SetPolygon(&g)
	case 4:
		g := box2d.ChainSegment{Ghost1: box2d.Vec2{X: -5, Y: y - .5}, Segment: box2d.Segment{Point1: box2d.Vec2{X: -3, Y: y}, Point2: box2d.Vec2{X: 3, Y: y}}, Ghost2: box2d.Vec2{X: 5, Y: y + .5}, ChainID: 42}
		s.SetChainSegment(&g)
	}
}

func TestBodiesConstructionCopiesCallerInputs(t *testing.T) {
	for bodyType := range 3 {
		for family := range 5 {
			name := fmt.Sprintf("body%d_shape%d", bodyType, family)
			t.Run(name, func(t *testing.T) {
				r := oracle.Open(t, "api/bodies/construction/"+name)
				w := apiWorld(t, box2d.Vec2{Y: -2})
				bd := box2d.DefaultBodyDef()
				bd.Type = box2d.BodyType(bodyType)
				bd.Position = apiBase(false)
				bd.Rotation = box2d.Rot{C: .6, S: .8}
				bd.LinearVelocity = box2d.Vec2{X: 2, Y: 1}
				bd.AngularVelocity = .5
				bd.LinearDamping = .2
				bd.AngularDamping = .3
				bd.GravityScale = .7
				bd.UserData = 11
				b := w.CreateBody(&bd)
				sd := box2d.DefaultShapeDef()
				sd.UserData = 7
				sd.Density = 2
				sd.Material.Friction = .4
				sd.Material.UserMaterialID = 21
				// Each constructor receives a mutable geometry value which is changed afterwards.
				circle := box2d.Circle{Center: box2d.Vec2{X: 1, Y: .5}, Radius: .75}
				capsule := box2d.Capsule{Center1: box2d.Vec2{X: -1, Y: .5}, Center2: box2d.Vec2{X: 1, Y: .5}, Radius: .4}
				segment := box2d.Segment{Point1: box2d.Vec2{X: -2}, Point2: box2d.Vec2{X: 2}}
				polygon := box2d.MakeOffsetBox(1, .5, box2d.Vec2{X: .5, Y: .25}, box2d.Rot{C: 1})
				chain := box2d.ChainSegment{Ghost1: box2d.Vec2{X: -4}, Segment: segment, Ghost2: box2d.Vec2{X: 4}}

				var s box2d.Shape

				switch family {
				case 0:
					s = b.CreateCircleShape(&sd, &circle)
				case 1:
					s = b.CreateCapsuleShape(&sd, &capsule)
				case 2:
					s = b.CreateSegmentShape(&sd, &segment)
				case 3:
					s = b.CreatePolygonShape(&sd, &polygon)
				case 4:
					s = b.CreateChainSegmentShape(&sd, &chain)
				}

				circle.Center.X = 100
				circle.Radius = 8
				capsule.Center1.X = 100
				capsule.Center2.Y = 100
				capsule.Radius = 8
				segment.Point1.X = 100
				polygon.Vertices[0].X = 100
				polygon.Normals[0].Y = 100
				polygon.Centroid.X = 100
				polygon.Radius = 8
				chain.Ghost1.X = 100
				chain.Segment.Point2.Y = 100
				sd.Density = 9
				sd.Material.Friction = .9
				sd.Material.UserMaterialID = 99
				sd.Filter.MaskBits = 0
				sd.UserData = 99
				bd.Position = box2d.Pos{X: 100, Y: 100}
				bd.Type = box2d.StaticBody
				bd.LinearVelocity = box2d.Vec2{}
				bd.Rotation = box2d.Rot{C: 1}
				bd.UserData = 99
				other := w.CreateBody(&bd)

				bodyUser, ok := b.UserData().(int)
				if !ok {
					t.Fatalf("body user data has type %T, want int", b.UserData())
				}

				reusedUser, ok := other.UserData().(int)
				if !ok {
					t.Fatalf("reused body user data has type %T, want int", other.UserData())
				}

				r.I32("bodyUser", int32(bodyUser))
				r.I32("reusedUser", int32(reusedUser))
				apiShape(r, "copied", s, b)
				apiBody(r, "constructed", b)

				reused := apiCreateGeometry(other, &sd, family)
				apiShape(r, "reusedShape", reused, other)
				// Give line families positive body mass before simulating them.
				ballastDef := box2d.DefaultShapeDef()
				ballastDef.UserData = 8
				ballast := box2d.Circle{Center: box2d.Vec2{X: -10}, Radius: .5}
				b.CreateCircleShape(&ballastDef, &ballast)
				apiCast(r, "ray", s.RayCast(box2d.Pos{X: 3, Y: -3}, box2d.Vec2{X: 8}))
				apiPos(r, "closest", s.ClosestPoint(box2d.Pos{X: 10, Y: 1}))

				for step := range 3 {
					w.Step(1.0/60, 4)
					apiBody(r, fmt.Sprintf("step/%d", step), b)
				}
			})
		}
	}
}

func TestBodiesMotionOperationsAndWake(t *testing.T) {
	apiMotionOperationCases(t)
	apiMotionLockCases(t)
	apiTargetTransformCases(t)
}

func apiMotionOperationCase(t *testing.T, name string, kind, distant, sleeping, action, wake int) {
	t.Helper()
	r := oracle.Open(t, "api/bodies/motion/"+name)
	base := apiBase(distant == 1)
	w := apiWorld(t, box2d.Vec2{X: 1, Y: -4})
	bd := box2d.DefaultBodyDef()
	bd.Type = box2d.BodyType(kind)
	bd.Position = base
	bd.Rotation = box2d.Rot{C: .6, S: .8}
	bd.UserData = 11
	b := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	sd.UserData = 7
	sd.Density = 2
	g := box2d.Circle{Center: box2d.Vec2{X: 1, Y: .5}, Radius: .75}
	b.CreateCircleShape(&sd, &g)
	b.SetLinearVelocity(box2d.Vec2{X: 2, Y: -1})
	b.SetAngularVelocity(.75)
	apiVec(r, "localPoint", b.LocalPoint(box2d.Pos{X: base.X + 3, Y: base.Y - 2}))
	apiPos(r, "worldPoint", b.WorldPoint(box2d.Vec2{X: 1.25, Y: -.75}))
	apiVec(r, "localVector", b.LocalVector(box2d.Vec2{X: 3, Y: -2}))
	apiVec(r, "worldVector", b.WorldVector(box2d.Vec2{X: 1.25, Y: -.75}))
	apiVec(r, "localVelocity", b.LocalPointVelocity(box2d.Vec2{X: 1.25, Y: -.75}))
	apiVec(r, "worldVelocity", b.WorldPointVelocity(box2d.Pos{X: base.X + 3, Y: base.Y - 2}))
	b.SetLinearDamping(.4)
	b.SetAngularDamping(.7)
	b.SetGravityScale(.3)

	if sleeping == 1 {
		b.SetAwake(false)
	}

	point := box2d.Pos{X: base.X + 2, Y: base.Y + 1}
	flag := wake == 1

	switch action {
	case 0:
		b.ApplyForce(box2d.Vec2{X: 5, Y: 9}, point, flag)
	case 1:
		b.ApplyForceToCenter(box2d.Vec2{X: 5, Y: 9}, flag)
	case 2:
		b.ApplyTorque(7, flag)
	case 3:
		b.ApplyLinearImpulse(box2d.Vec2{X: 3, Y: 4}, point, flag)
	case 4:
		b.ApplyLinearImpulseToCenter(box2d.Vec2{X: 3, Y: 4}, flag)
	case 5:
		b.ApplyAngularImpulse(2, flag)
	case 6:
		b.SetTargetTransform(box2d.WorldTransform{P: point, Q: box2d.Rot{C: 0, S: 1}}, 1.0/60, flag)
	case 7:
		b.SetTargetTransform(box2d.WorldTransform{P: box2d.Pos{X: base.X + .0001, Y: base.Y}, Q: box2d.Rot{C: .6, S: .8}}, 1.0/60, flag)
	case 8:
		b.ApplyForce(box2d.Vec2{X: 5, Y: 9}, point, flag)
		b.ApplyTorque(7, flag)
		b.ClearForces()
	}

	apiBody(r, "immediate", b)

	for step := range 3 {
		w.Step(1.0/60, 4)
		apiBody(r, fmt.Sprintf("step/%d", step), b)
	}
}

func apiMotionOperationCases(t *testing.T) {
	t.Helper()

	for distant := range 2 {
		if distant == 1 && !box2d.IsDoublePrecision() {
			continue
		}

		for kind := range 3 {
			for sleeping := range 2 {
				for action := range 9 {
					for wake := range 2 {
						name := fmt.Sprintf("type%d_base%d_sleep%d_action%d_wake%d", kind, distant, sleeping, action, wake)
						t.Run(name, func(t *testing.T) {
							apiMotionOperationCase(t, name, kind, distant, sleeping, action, wake)
						})
					}
				}
			}
		}
	}
}

func apiMotionLockCases(t *testing.T) {
	t.Helper()

	for locks := range 8 {
		for runtime := range 2 {
			name := fmt.Sprintf("locks%d_runtime%d", locks, runtime)
			t.Run(name, func(t *testing.T) {
				r := oracle.Open(t, "api/bodies/motion/"+name)
				w := apiWorld(t, box2d.Vec2{X: 1, Y: -4})
				m := box2d.MotionLocks{LinearX: locks&1 != 0, LinearY: locks&2 != 0, AngularZ: locks&4 != 0}
				bd := box2d.DefaultBodyDef()
				bd.Type = box2d.DynamicBody
				bd.Position = apiBase(false)
				bd.LinearVelocity = box2d.Vec2{X: 2, Y: -1}

				bd.AngularVelocity = .75
				if runtime == 0 {
					bd.MotionLocks = m
				}

				b := w.CreateBody(&bd)
				sd := box2d.DefaultShapeDef()
				sd.UserData = 7
				apiCreateGeometry(b, &sd, 3)

				if runtime == 1 {
					b.SetMotionLocks(m)
				}

				b.SetLinearDamping(.4)
				b.SetAngularDamping(.7)
				b.SetGravityScale(.3)
				b.ApplyForce(box2d.Vec2{X: 5, Y: 9}, box2d.Pos{X: 9, Y: -2}, true)
				b.ApplyTorque(7, true)
				b.ApplyLinearImpulse(box2d.Vec2{X: 3, Y: 4}, box2d.Pos{X: 9, Y: -2}, true)
				b.ApplyAngularImpulse(2, true)
				apiBody(r, "loaded", b)

				for step := range 3 {
					w.Step(1.0/60, 4)
					apiBody(r, fmt.Sprintf("step/%d", step), b)
				}

				b.SetMotionLocks(box2d.MotionLocks{})
				b.ApplyTorque(7, true)
				b.ApplyForceToCenter(box2d.Vec2{X: 5, Y: 9}, true)
				w.Step(1.0/60, 4)
				apiBody(r, "unlocked", b)
			})
		}
	}
}

func apiTargetTransformCases(t *testing.T) {
	t.Helper()

	for kind := range 3 {
		name := fmt.Sprintf("target_type%d", kind)
		t.Run(name, func(t *testing.T) {
			r := oracle.Open(t, "api/bodies/motion/"+name)
			w := apiWorld(t, box2d.Vec2{})
			bd := box2d.DefaultBodyDef()
			bd.Type = box2d.BodyType(kind)
			bd.Position = apiBase(false)
			bd.Rotation = box2d.Rot{C: 0, S: 1}
			b := w.CreateBody(&bd)
			sd := box2d.DefaultShapeDef()
			sd.UserData = 7
			apiCreateGeometry(b, &sd, 0)
			b.SetTransform(box2d.Pos{X: 8, Y: -2}, box2d.Rot{C: .6, S: .8})
			apiBody(r, "teleport", b)
			b.SetTargetTransform(box2d.WorldTransform{P: box2d.Pos{X: 9, Y: -1}, Q: box2d.Rot{C: 0, S: 1}}, .25, true)
			apiBody(r, "target", b)
			w.Step(.25, 4)
			apiBody(r, "arrived", b)
		})
	}
}

func TestBodiesMassMutationTriggersAndLoadedResponse(t *testing.T) {
	r := oracle.Open(t, "api/bodies/mass")
	w := apiWorld(t, box2d.Vec2{Y: -2})
	bd := box2d.DefaultBodyDef()
	bd.Type = box2d.DynamicBody
	bd.Position = apiBase(false)
	bd.Rotation = box2d.Rot{C: .6, S: .8}
	bd.AngularVelocity = .75
	b := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	sd.UserData = 7
	sd.Density = 2
	circle := box2d.Circle{Center: box2d.Vec2{X: 1, Y: .5}, Radius: .75}
	a := b.CreateCircleShape(&sd, &circle)
	observe := func(key string) { apiBody(r, key, b) }
	load := func(key string) {
		b.SetLinearVelocity(box2d.Vec2{X: 2, Y: -1})
		b.SetAngularVelocity(.75)
		b.ApplyLinearImpulse(box2d.Vec2{X: 3, Y: 4}, box2d.Pos{X: 9, Y: -2}, true)
		b.ApplyTorque(7, true)
		b.ApplyForceToCenter(box2d.Vec2{X: 5, Y: 9}, true)
		observe(key + "/loaded")
		w.Step(1.0/60, 4)
		observe(key + "/step")
	}

	observe("circle")
	load("circle")

	sd.UserData = 8
	sd.Density = 0
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -2}, Center2: box2d.Vec2{X: -1}, Radius: .4}
	z := b.CreateCapsuleShape(&sd, &capsule)

	observe("zeroAdded")
	load("zeroAdded")

	sd.UserData = 9
	sd.Density = 3
	sd.IsSensor = true
	sd.UpdateBodyMass = false
	polygon := box2d.MakeOffsetBox(.5, 1, box2d.Vec2{X: -1, Y: 2}, box2d.Rot{C: 1})
	p := b.CreatePolygonShape(&sd, &polygon)

	observe("deferredAdded")
	b.UpdateMassFromShapes()
	observe("updatedAdded")
	load("updatedAdded")
	a.SetDensity(4, false)
	observe("densityDeferred")
	load("densityDeferred")
	b.UpdateMassFromShapes()
	observe("densityUpdated")
	load("densityUpdated")
	z.SetDensity(2, true)
	observe("densityAutomatic")
	load("densityAutomatic")
	z.SetDensity(0, true)
	observe("densityBackToZero")
	load("densityBackToZero")
	b.SetMassData(box2d.MassData{Mass: 7, Center: box2d.Vec2{X: -.25, Y: 1.5}, RotationalInertia: 11})
	observe("override")
	load("override")
	b.UpdateMassFromShapes()
	observe("restoreOverride")
	load("restoreOverride")
	b.SetMassData(box2d.MassData{Mass: 7, Center: box2d.Vec2{X: -.25, Y: 1.5}, RotationalInertia: 11})
	observe("overrideBeforeRemove")
	z.Destroy(false)
	r.Bool("removedZeroValid", z.IsValid())
	observe("removeDeferred")
	load("removeDeferred")
	b.UpdateMassFromShapes()
	observe("removeUpdated")
	load("removeUpdated")
	b.SetMassData(box2d.MassData{Mass: 9, Center: box2d.Vec2{X: .5, Y: -.5}, RotationalInertia: 13})

	sd.IsSensor = false
	sd.UpdateBodyMass = true
	sd.UserData = 10
	sd.Density = 1
	extra := b.CreateCircleShape(&sd, &circle)

	observe("addResetsOverride")
	load("addResetsOverride")
	extra.Destroy(true)
	observe("removeAutomatic")
	load("removeAutomatic")

	for kind := range 3 {
		b.SetMassData(box2d.MassData{Mass: 9, Center: box2d.Vec2{X: .5, Y: -.5}, RotationalInertia: 13})
		b.SetType(box2d.BodyType(kind))
		key := fmt.Sprintf("type/%d", kind)
		observe(key)
		load(key)
	}

	b.SetMotionLocks(box2d.MotionLocks{AngularZ: true})
	b.SetMassData(box2d.MassData{Mass: 7, Center: box2d.Vec2{X: -.25, Y: 1.5}, RotationalInertia: 11})
	observe("lockedOverride")
	load("lockedOverride")
	b.SetMotionLocks(box2d.MotionLocks{})
	observe("unlockRecomputes")
	load("unlockRecomputes")
	p.Destroy(true)
	a.Destroy(true)
	observe("empty")
	load("empty")
}

func TestShapesMutationBoundsQueriesAndCollisions(t *testing.T) {
	for family := range 5 {
		name := fmt.Sprintf("shape%d", family)
		t.Run(name, func(t *testing.T) {
			r := oracle.Open(t, "api/shapes/mutation/"+name)
			base := apiBase(false)
			w := apiWorld(t, box2d.Vec2{})
			bd := box2d.DefaultBodyDef()
			bd.Position = base
			ground := w.CreateBody(&bd)
			sd := box2d.DefaultShapeDef()
			sd.UserData = 7
			sd.Filter.CategoryBits = 4
			sd.EnableContactEvents = true
			sd.EnableHitEvents = true
			sd.Material.Friction = .3
			s := apiCreateGeometry(ground, &sd, family)
			bd.Type = box2d.DynamicBody
			bd.Position.Y -= oracle.Real(float32(.3))
			bd.EnableSleep = false
			visitor := w.CreateBody(&bd)
			sd.Filter.CategoryBits = 2
			sd.UserData = 9
			g := box2d.Circle{Radius: .5}
			visitor.CreateCircleShape(&sd, &g)

			observe := func(key string) {
				apiShape(r, key+"/shape", s, ground)
				apiQuery(r, key+"/oldQuery", w, base, 0, 4)
				apiQuery(r, key+"/newQuery", w, base, 3, 4)
				apiBody(r, key+"/body", visitor)
				apiContacts(r, key+"/effects", w, visitor)
			}
			impact := func(key string, y float32) {
				visitor.SetTransform(box2d.Pos{X: base.X, Y: base.Y + oracle.Real(y) - oracle.Real(float32(.3))}, box2d.Rot{C: 1})
				visitor.SetLinearVelocity(box2d.Vec2{X: 2, Y: 3})
				visitor.SetAngularVelocity(.5)

				for step := range 4 {
					w.Step(1.0/60, 4)
					observe(fmt.Sprintf("%s/%d", key, step))
				}
			}
			impact("initial", 0)
			apiSetGeometry(s, family, 3)
			observe("geometryImmediate")
			w.Step(1.0/60, 4)
			observe("oldLocationCleared")
			impact("geometryImpact", 3)
			s.SetDensity(2, true)
			impact("density", 3)
			s.SetFriction(.8)
			impact("friction", 3)
			s.SetRestitution(.7)
			impact("restitution", 3)
			// The application policy gives material 91 a distinct physical response.
			w.SetFrictionCallback(func(_ float32, a uint64, _ float32, b uint64) float32 {
				if a == 91 || b == 91 {
					return .05
				}

				return .6
			})
			s.SetUserMaterial(91)
			impact("tag", 3)
			w.SetFrictionCallback(nil)

			material := box2d.SurfaceMaterial{Friction: .6, Restitution: .2, RollingResistance: .15, TangentSpeed: 2, UserMaterialID: 92, CustomColor: 0x123456}
			s.SetSurfaceMaterial(&material)
			material.TangentSpeed = 99

			impact("surface", 3)

			filter := s.Filter()
			filter.MaskBits = 0
			s.SetFilter(filter)
			observe("filterImmediate")
			impact("filtered", 3)

			filter.MaskBits = ^uint64(0)
			filter.CategoryBits = 8
			s.SetFilter(filter)
			impact("categoryChanged", 3)
			apiQuery(r, "category8", w, base, 3, 8)

			filter.CategoryBits = 4
			s.SetFilter(filter)
			s.SetSensorEventsEnabled(true)
			s.SetContactEventsEnabled(false)
			s.SetHitEventsEnabled(false)
			s.SetPreSolveEventsEnabled(true)
			impact("flagsOff", 3)
			s.SetSensorEventsEnabled(false)
			s.SetContactEventsEnabled(true)
			s.SetHitEventsEnabled(true)
			s.SetPreSolveEventsEnabled(false)
			impact("flagsOn", 3)
			// Every legal cross-family setter is also exercised on a live shape.
			for next := range 5 {
				apiSetGeometry(s, next, 3)
				impact(fmt.Sprintf("converted/%d", next), 3)
			}

			s.Destroy(true)
			r.Bool("destroyed", s.IsValid())
			w.Step(1.0/60, 4)
			apiQuery(r, "destroyedQuery", w, base, 3, 4)
			apiBody(r, "destroyedBody", visitor)
			apiContacts(r, "destroyedEffects", w, visitor)
		})
	}
}

func apiEnvironmentalForceCase(r *oracle.Reader, w *box2d.World, b box2d.Body, s box2d.Shape, base box2d.Pos, mode, wake int) {
	key := fmt.Sprintf("force/%d/wake/%d", mode, wake)

	b.SetTransform(base, box2d.Rot{S: 1})
	b.SetLinearVelocity(box2d.Vec2{X: 1, Y: -.5})
	b.SetAngularVelocity(.3)
	b.SetAwake(false)

	if mode == 1 {
		b.SetAwake(true)
		b.SetLinearVelocity(box2d.Vec2{X: 1, Y: -.5})
		b.SetAngularVelocity(.3)
	}

	if mode < 2 {
		s.ApplyWind(box2d.Vec2{X: 5, Y: 2}, .4, .2, wake == 1)
	} else {
		if wake == 1 {
			b.SetAwake(true)
		}

		def := box2d.DefaultExplosionDef()
		def.MaskBits = 1
		def.Position = box2d.Pos{X: base.X - 2, Y: base.Y - 1}
		def.Radius = 1
		def.Falloff = 5

		def.ImpulsePerLength = 3
		if mode == 3 {
			def.ImpulsePerLength = -3
		}

		w.Explode(&def)
	}

	apiBody(r, key+"/immediate", b)

	for step := range 3 {
		w.Step(1.0/60, 4)
		apiBody(r, fmt.Sprintf("%s/step/%d", key, step), b)
	}
}

func apiShapeQueryCase(t *testing.T, name string, distant, family int) {
	t.Helper()
	r := oracle.Open(t, "api/shapes/queries/"+name)

	base := apiBase(distant == 1)
	if distant == 2 {
		base = box2d.Pos{}
	}

	w := apiWorld(t, box2d.Vec2{})
	bd := box2d.DefaultBodyDef()
	bd.Type = box2d.DynamicBody
	bd.Position = base
	bd.Rotation = box2d.Rot{S: 1}
	b := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	sd.UserData = 7
	s := apiCreateGeometry(b, &sd, family)
	ballastDef := box2d.DefaultShapeDef()
	ballastDef.UserData = 8
	ballastDef.Filter.CategoryBits = 2
	ballast := box2d.Circle{Center: box2d.Vec2{X: -10}, Radius: .5}
	b.CreateCircleShape(&ballastDef, &ballast)
	apiShape(r, "shape", s, b)

	if family == 0 || family == 1 || family == 3 {
		boundary := box2d.Pos{X: base.X + .5, Y: base.Y}
		if family == 0 {
			boundary.X = base.X + 1
		}

		for i, p := range []box2d.Pos{{X: base.X, Y: base.Y}, {X: base.X + .5, Y: base.Y + .25}, {X: base.X + 4, Y: base.Y + 4}, boundary} {
			r.Bool(fmt.Sprintf("point/%d", i), s.TestPoint(p))
		}
	}

	rays := [7]struct {
		p box2d.Pos
		v box2d.Vec2
	}{
		{box2d.Pos{X: base.X, Y: base.Y - 4}, box2d.Vec2{Y: 8}},
		{box2d.Pos{X: base.X, Y: base.Y + 4}, box2d.Vec2{Y: -8}},
		{box2d.Pos{X: base.X + 4, Y: base.Y - 4}, box2d.Vec2{Y: 8}},
		{base, box2d.Vec2{Y: 8}},
		{box2d.Pos{X: base.X - 4, Y: base.Y}, box2d.Vec2{X: 8}},
		{box2d.Pos{X: base.X, Y: base.Y - 4}, box2d.Vec2{Y: 1}},
		{box2d.Pos{X: base.X + .5, Y: base.Y - 4}, box2d.Vec2{Y: 8}},
	}
	for i, ray := range rays {
		apiCast(r, fmt.Sprintf("ray/%d", i), s.RayCast(ray.p, ray.v))
	}

	for i, p := range []box2d.Pos{{X: base.X + 3, Y: base.Y + 5}, base, {X: base.X - 4, Y: base.Y - 2}} {
		apiPos(r, fmt.Sprintf("closest/%d", i), s.ClosestPoint(p))
	}
	// Sleeping and already-moving wind cases distinguish wake from relative drag.
	for mode := range 4 {
		for wake := range 2 {
			apiEnvironmentalForceCase(r, w, b, s, base, mode, wake)
		}
	}
}

func TestShapesTransformedQueriesAndEnvironmentalForces(t *testing.T) {
	for distant := range 3 {
		if distant == 1 && !box2d.IsDoublePrecision() {
			continue
		}

		for family := range 5 {
			name := fmt.Sprintf("base%d_shape%d", distant, family)
			t.Run(name, func(t *testing.T) {
				apiShapeQueryCase(t, name, distant, family)
			})
		}
	}
}

func apiChainTopology(topology, reverse int) ([]box2d.Vec2, bool) {
	points := []box2d.Vec2{{X: -2}, {X: 2}}
	loop := false

	switch topology {
	case 1:
		points = []box2d.Vec2{{X: -2}, {}, {X: 2}, {X: 3, Y: 1}}
	case 2:
		points = []box2d.Vec2{{X: -2}, {X: 2}, {X: 0, Y: 3}}
		loop = true
	case 3:
		points = []box2d.Vec2{{X: -2}, {X: 2}, {X: 2, Y: 3}, {X: -2, Y: 3}}
		loop = true
	case 4:
		points = []box2d.Vec2{{X: -.02}, {}, {X: .02}}
	}

	if reverse == 1 {
		for i := range len(points) / 2 {
			j := len(points) - 1 - i
			points[i], points[j] = points[j], points[i]
		}
	}

	return points, loop
}

func apiChainMaterialImpact(r *oracle.Reader, key string, segment box2d.Shape, ground, visitor box2d.Body, w *box2d.World) {
	edge := segment.ChainSegment().Segment
	tangent := edge.Point2.Sub(edge.Point1).Normalize()
	normal := tangent.RightPerp()
	start := edge.Point1.Lerp(edge.Point2, .5).Add(normal.Scale(.3))
	visitor.SetTransform(ground.WorldPoint(start), box2d.Rot{C: 1})
	visitor.SetLinearVelocity(normal.Scale(-2).Add(tangent.Scale(.5)))
	visitor.SetAngularVelocity(.5)

	for step := range 8 {
		w.Step(1.0/60, 4)

		p := fmt.Sprintf("%s/%d", key, step)
		apiBody(r, p+"/body", visitor)
		apiContacts(r, p+"/effects", w, visitor)
	}
}

func apiChainCase(t *testing.T, name string, topology, reverse int) {
	t.Helper()
	r := oracle.Open(t, "api/chains/"+name)
	base := apiBase(false)
	w := apiWorld(t, box2d.Vec2{})
	bd := box2d.DefaultBodyDef()
	bd.Position = base
	ground := w.CreateBody(&bd)
	points, loop := apiChainTopology(topology, reverse)

	materials := make([]box2d.SurfaceMaterial, len(points))
	for i := range materials {
		materials[i] = box2d.DefaultSurfaceMaterial()
		materials[i].Friction = .2
		materials[i].UserMaterialID = uint64(20 + i)
	}

	cd := box2d.DefaultChainDef()
	cd.Points = points

	cd.Materials = materials
	if !loop {
		cd.Materials = materials[:len(points)-1]
	}

	cd.IsLoop = loop
	cd.Ghost1 = box2d.Vec2{X: -4}

	cd.Ghost2 = box2d.Vec2{X: 4, Y: 2}
	if reverse == 1 {
		cd.Ghost1, cd.Ghost2 = cd.Ghost2, cd.Ghost1
	}

	cd.UserData = 7
	cd.Filter.CategoryBits = 4
	chain := ground.CreateChain(&cd)

	for i := range points {
		points[i] = box2d.Vec2{X: 100, Y: 100}
		materials[i].Friction = .9
		materials[i].UserMaterialID = 99
	}

	cd.Ghost1.X = 100
	cd.Ghost2.Y = 100
	cd.Filter.MaskBits = 0
	cd.UserData = 99

	segments := make([]box2d.Shape, 0, chain.SegmentCount())
	for s := range chain.Segments() {
		r.I32(fmt.Sprintf("copiedUser/%d", len(segments)), apiTag(s))
		s.SetUserData(100 + len(segments))
		segments = append(segments, s)
	}

	observe := func(key string) {
		r.I32(key+"/count", int32(chain.SegmentCount()))

		n := 0
		for s := range chain.Segments() {
			p := fmt.Sprintf("%s/segment/%d", key, n)
			r.Bool(p+"/identity", s == segments[n])
			r.Bool(p+"/parent", s.ParentChain() == chain)
			r.Bool(p+"/valid", s.IsValid())
			apiShape(r, p, s, ground)
			apiMaterial(r, p+"/reported", chain.SurfaceMaterial(n))
			n++
		}

		r.I32(key+"/enumerated", int32(n))
		r.Bool(key+"/world", chain.World() == w)
		apiBody(r, key+"/ground", ground)
		apiQuery(r, key+"/query", w, base, 0, 4)
	}
	observe("copied")

	bd.Type = box2d.DynamicBody
	bd.EnableSleep = false
	bd.Position = box2d.Pos{X: base.X, Y: base.Y - 1}
	visitor := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	sd.UserData = 9
	sd.Filter.CategoryBits = 2
	sd.EnableContactEvents = true
	sd.EnableHitEvents = true
	g := box2d.Circle{Radius: .25}
	visitor.CreateCircleShape(&sd, &g)

	materialImpact := func(key string) {
		apiChainMaterialImpact(r, key, segments[0], ground, visitor, w)
	}
	materialImpact("originalMaterialImpact")

	m := box2d.SurfaceMaterial{Friction: .7, Restitution: .4, TangentSpeed: 1.5, RollingResistance: .1, UserMaterialID: 91, CustomColor: 0xabcdef}
	chain.SetSurfaceMaterial(&m, 0)
	m.Friction = .9

	observe("oneMaterial")
	materialImpact("oneMaterialImpact")

	m.UserMaterialID = 92
	chain.SetAllSurfaceMaterials(&m)
	m.TangentSpeed = 99

	observe("allMaterials")
	materialImpact("allMaterialImpact")

	for side := range 2 {
		y := float32(-1)
		speed := float32(2)

		if side == 1 {
			y = 1
			speed = -2
		}

		visitor.SetTransform(box2d.Pos{X: base.X, Y: base.Y + oracle.Real(y)}, box2d.Rot{C: 1})
		visitor.SetLinearVelocity(box2d.Vec2{Y: speed})
		visitor.SetAngularVelocity(0)

		for step := range 60 {
			w.Step(1.0/60, 4)

			key := fmt.Sprintf("side/%d/step/%d", side, step)
			apiBody(r, key+"/body", visitor)
			apiContacts(r, key+"/effects", w, visitor)
		}
	}

	apiQuery(r, "beforeDestroy", w, base, 0, 4)
	chain.Destroy()
	r.Bool("chainValid", chain.IsValid())

	for i, s := range segments {
		r.Bool(fmt.Sprintf("segmentValid/%d", i), s.IsValid())
	}

	apiBody(r, "afterDestroy", ground)
	w.Step(1.0/60, 4)
	apiQuery(r, "afterDestroyQuery", w, base, 0, 4)
	apiContacts(r, "afterDestroyEffects", w, visitor)
	// Reallocated objects cannot resurrect saved full chain/segment handles.
	cd = box2d.DefaultChainDef()
	cd.Points = []box2d.Vec2{{X: -2}, {X: 2}}
	cd.Ghost1 = box2d.Vec2{X: -4}
	cd.Ghost2 = box2d.Vec2{X: 4}
	cd.Materials = []box2d.SurfaceMaterial{box2d.DefaultSurfaceMaterial()}
	replacement := ground.CreateChain(&cd)
	r.Bool("replacementDifferent", replacement != chain)
	r.Bool("oldChainStillInvalid", chain.IsValid())

	for i, s := range segments {
		r.Bool(fmt.Sprintf("oldSegmentStillInvalid/%d", i), s.IsValid())
	}

	ground.Destroy()
	r.Bool("replacementDestroyedWithBody", replacement.IsValid())
}

func TestChainsOwnershipMaterialsWindingAndDestruction(t *testing.T) {
	for topology := range 5 {
		for reverse := range 2 {
			name := fmt.Sprintf("topology%d_reverse%d", topology, reverse)
			t.Run(name, func(t *testing.T) {
				apiChainCase(t, name, topology, reverse)
			})
		}
	}
}
