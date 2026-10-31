// SPDX-License-Identifier: MIT

package box2d_test

import (
	"cmp"
	"fmt"
	"iter"
	"math"
	"reflect"
	"slices"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

type evOracle struct {
	t       *testing.T
	r       *oracle.Reader
	w       *box2d.World
	tags    map[box2d.Shape]int
	history []box2d.Contact
	calls   []evCall
	reject  bool
}

func evStart(t *testing.T, name string, gravity box2d.Vec2, sleep bool) *evOracle {
	t.Helper()

	d := box2d.DefaultWorldDef()
	d.Gravity = gravity
	d.EnableSleep = sleep
	d.WorkerCount = 1
	e := &evOracle{t: t, r: oracle.Open(t, "events/"+name), w: box2d.NewWorld(&d), tags: make(map[box2d.Shape]int)}
	box2d.CleanupWorld(t, e.w)

	return e
}

func (e *evOracle) body(tag int, typ box2d.BodyType, p box2d.Pos) box2d.Body {
	d := box2d.DefaultBodyDef()
	d.Type = typ
	d.Position = p
	d.UserData = tag

	return e.w.CreateBody(&d)
}

func (e *evOracle) circle(b box2d.Body, tag int, radius float32, d *box2d.ShapeDef) box2d.Shape {
	s := b.CreateCircleShape(d, &box2d.Circle{Radius: radius})
	e.tags[s] = tag

	return s
}

func evShapeCompare(a, b box2d.Shape) int {
	x, y := a.ID(), b.ID()
	if c := cmp.Compare(uint32(x>>32), uint32(y>>32)); c != 0 {
		return c
	}

	return cmp.Compare(uint16(x), uint16(y))
}

func evPairCompare(a, b, c, d box2d.Shape) int {
	if x := evShapeCompare(a, c); x != 0 {
		return x
	}

	return evShapeCompare(b, d)
}

func (e *evOracle) shape(s box2d.Shape) {
	id := s.ID()
	e.r.I32("shape.index", int32(id>>32))
	e.r.U32("shape.generation", uint32(uint16(id)))
	e.r.Bool("shape.world", s.World() == e.w)
	e.r.Bool("shape.valid", s.IsValid())
	e.r.I32("shape.tag", int32(e.tags[s]))
}

func (e *evOracle) bodyID(b box2d.Body) {
	id := b.ID()
	e.r.I32("body.index", int32(id>>32))
	e.r.U32("body.generation", uint32(uint16(id)))
	e.r.Bool("body.world", b.World() == e.w)
	e.r.Bool("body.valid", b.IsValid())
}

func (e *evOracle) contact(c box2d.Contact) {
	id := c.ID()
	e.r.I32("contact.index", int32(id>>32))
	e.r.U32("contact.generation", uint32(id))
	e.r.Bool("contact.world", c.World() == e.w)
	e.r.Bool("contact.valid", c.IsValid())
}

func (e *evOracle) observe(bodies ...box2d.Body) {
	cb := slices.Collect(e.w.ContactBeginEvents())
	ce := slices.Collect(e.w.ContactEndEvents())
	ch := slices.Collect(e.w.ContactHitEvents())
	sb := slices.Collect(e.w.SensorBeginEvents())
	se := slices.Collect(e.w.SensorEndEvents())
	bm := slices.Collect(e.w.BodyMoveEvents())
	je := slices.Collect(e.w.JointEvents())

	slices.SortFunc(cb, func(a, b box2d.ContactBeginTouchEvent) int {
		return evPairCompare(a.ShapeA, a.ShapeB, b.ShapeA, b.ShapeB)
	})
	slices.SortFunc(ce, func(a, b box2d.ContactEndTouchEvent) int {
		return evPairCompare(a.ShapeA, a.ShapeB, b.ShapeA, b.ShapeB)
	})
	slices.SortFunc(ch, func(a, b box2d.ContactHitEvent) int {
		return evPairCompare(a.ShapeA, a.ShapeB, b.ShapeA, b.ShapeB)
	})
	slices.SortFunc(sb, func(a, b box2d.SensorBeginTouchEvent) int {
		return evPairCompare(a.SensorShape, a.VisitorShape, b.SensorShape, b.VisitorShape)
	})
	slices.SortFunc(se, func(a, b box2d.SensorEndTouchEvent) int {
		return evPairCompare(a.SensorShape, a.VisitorShape, b.SensorShape, b.VisitorShape)
	})
	slices.SortFunc(bm, func(a, b box2d.BodyMoveEvent) int { return cmp.Compare(a.Body.ID()>>32, b.Body.ID()>>32) })
	slices.SortFunc(je, func(a, b box2d.JointEvent) int { return cmp.Compare(a.Joint.ID()>>32, b.Joint.ID()>>32) })

	for _, c := range cb {
		if !slices.Contains(e.history, c.Contact) {
			e.history = append(e.history, c.Contact)
		}
	}

	e.r.I32("history.count", int32(len(e.history)))

	for _, c := range e.history {
		e.contact(c)
	}

	r := e.r
	r.I32("contact.begin", int32(len(cb)))

	for _, c := range cb {
		e.shape(c.ShapeA)
		e.shape(c.ShapeB)
		e.contact(c.Contact)
	}

	r.I32("contact.end", int32(len(ce)))

	for _, c := range ce {
		e.shape(c.ShapeA)
		e.shape(c.ShapeB)
		e.contact(c.Contact)
	}

	r.I32("contact.hit", int32(len(ch)))

	for _, c := range ch {
		e.shape(c.ShapeA)
		e.shape(c.ShapeB)
		e.contact(c.Contact)
		r.Pos("hit.point", c.Point.X, c.Point.Y)
		r.Vec2("hit.normal", c.Normal.X, c.Normal.Y)
		r.F32("hit.speed", c.ApproachSpeed)
	}

	r.I32("sensor.begin", int32(len(sb)))

	for _, s := range sb {
		e.shape(s.SensorShape)
		e.shape(s.VisitorShape)
	}

	r.I32("sensor.end", int32(len(se)))

	for _, s := range se {
		e.shape(s.SensorShape)
		e.shape(s.VisitorShape)
	}

	e.observeOwnedEvents(bm, je)

	e.observeBodyStates(bodies)
}
func (e *evOracle) step(b ...box2d.Body) { e.w.Step(1.0/60.0, 4); e.observe(b...) }

func TestEventsLifecycleParity(t *testing.T) {
	for sensor := range 2 {
		for flags := range 4 {
			name := fmt.Sprintf("Lifecycle-%d-%d", sensor, flags)
			t.Run(name, func(t *testing.T) {
				e := evStart(t, name, box2d.Vec2{}, false)
				a := e.body(1, box2d.StaticBody, box2d.Pos{})
				d := box2d.DefaultShapeDef()
				d.IsSensor = sensor != 0
				d.EnableSensorEvents = flags&1 != 0
				d.EnableContactEvents = flags&1 != 0
				root := e.circle(a, 10, 2, &d)
				d.IsSensor = false
				d.Density = 1
				d.EnableSensorEvents = flags&2 != 0
				d.EnableContactEvents = flags&2 != 0
				b := []box2d.Body{e.body(2, box2d.DynamicBody, box2d.Pos{X: -1.9}), e.body(3, box2d.DynamicBody, box2d.Pos{X: 1.9})}
				v := []box2d.Shape{e.circle(b[0], 20, .25, &d), e.circle(b[1], 30, .25, &d)}
				e.step(b...)
				e.step(b...)
				b[0].SetTransform(box2d.Pos{X: -5}, box2d.Rot{C: 1})
				e.step(b...)
				b[0].SetTransform(box2d.Pos{X: -1.9}, box2d.Rot{C: 1})
				e.step(b...)

				f := box2d.DefaultFilter()
				f.MaskBits = 0
				v[0].SetFilter(f)
				e.step(b...)
				v[0].SetFilter(box2d.DefaultFilter())
				e.step(b...)
				b[0].SetType(box2d.StaticBody)
				e.step(b...)
				b[0].SetType(box2d.DynamicBody)
				e.step(b...)
				b[1].Disable()
				e.step(b...)
				b[1].Enable()
				e.step(b...)
				b[0].Destroy()
				e.step(b...)
				b[0] = e.body(4, box2d.DynamicBody, box2d.Pos{X: -1.9})
				v[0] = e.circle(b[0], 40, .25, &d)
				e.step(b...)
				root.Destroy(true)
				e.step(b...)

				d.IsSensor = sensor != 0
				d.EnableSensorEvents = flags&1 != 0
				d.EnableContactEvents = flags&1 != 0
				root = e.circle(a, 11, 2, &d)
				e.step(b...)
				e.step(b...)
				root.SetSensorEventsEnabled(flags&1 == 0)
				root.SetContactEventsEnabled(flags&1 == 0)
				v[0].SetSensorEventsEnabled(flags&2 == 0)
				v[0].SetContactEventsEnabled(flags&2 == 0)
				b[0].SetTransform(box2d.Pos{X: -5}, box2d.Rot{C: 1})
				e.step(b...)
				b[0].SetTransform(box2d.Pos{X: -1.9}, box2d.Rot{C: 1})
				e.step(b...)
			})
		}
	}
}

func TestEventsMovementParity(t *testing.T) {
	e := evStart(t, "Movement", box2d.Vec2{}, true)
	b := e.body(2, box2d.DynamicBody, box2d.Pos{X: 3, Y: 4})
	d := box2d.DefaultShapeDef()
	d.Density = 1
	e.circle(b, 20, .5, &d)
	b.SetLinearDamping(3)
	b.SetLinearVelocity(box2d.Vec2{X: 2, Y: 1})

	step := 0
	for step < 600 && b.Awake() {
		e.step(b)
		step++
	}

	e.r.I32("sleep.steps", int32(step))

	if step >= 600 {
		t.Fatal("body did not sleep within 600 steps")
	}

	e.step(b)
	b.SetTransform(box2d.Pos{X: 8, Y: 9}, box2d.MakeRot(.3))
	e.observe(b)
	e.step(b)
	b.SetAwake(true)
	e.step(b)
	b.SetAwake(false)
	b.ApplyForceToCenter(box2d.Vec2{X: 5, Y: 2}, false)
	e.step(b)
	b.ApplyForceToCenter(box2d.Vec2{X: 5, Y: 2}, true)
	e.step(b)
	b.SetAwake(false)
	b.SetSleepEnabled(false)
	e.step(b)
}

func TestEventsFastSensorParity(t *testing.T) {
	e := evStart(t, "FastSensor", box2d.Vec2{}, false)
	a := e.body(1, box2d.StaticBody, box2d.Pos{})
	b := e.body(2, box2d.DynamicBody, box2d.Pos{X: -3})
	d := box2d.DefaultShapeDef()
	d.IsSensor = true
	d.EnableSensorEvents = true
	e.circle(a, 10, 1, &d)
	d.IsSensor = false
	d.Density = 1
	e.circle(b, 20, .1, &d)
	b.SetBullet(true)
	b.SetLinearVelocity(box2d.Vec2{X: 20})

	for range 24 {
		e.step(b)
	}
}

func TestEventsHitThresholdParity(t *testing.T) {
	for flags := range 4 {
		for thresholdCase := range 2 {
			for _, side := range []int{-1, 0, 1} {
				name := fmt.Sprintf("Hit-%d-%d-%d", flags, thresholdCase, side)
				t.Run(name, func(t *testing.T) {
					e := evStart(t, name, box2d.Vec2{}, false)

					b := e.hitThresholdScene(flags, thresholdCase, side)

					for range 10 {
						e.step(b)
					}
				})
			}
		}
	}
}

func TestEventsJointThresholdParity(t *testing.T) {
	for axis := range 2 {
		for _, side := range []int{-2, -1, 0, 1, 2} {
			name := fmt.Sprintf("Joint-%d-%d", axis, side)
			t.Run(name, func(t *testing.T) {
				e := evStart(t, name, box2d.Vec2{}, false)
				a := e.body(1, box2d.StaticBody, box2d.Pos{})
				b := e.body(2, box2d.DynamicBody, box2d.Pos{})
				d := box2d.DefaultShapeDef()
				d.Density = 1
				e.circle(b, 20, .5, &d)

				jd := box2d.DefaultWeldJointDef()
				jd.Base.BodyA = a
				jd.Base.BodyB = b
				jd.Base.UserData = 71
				threshold := float32(10)

				switch side {
				case -2:
					threshold = 0
				case -1:
					threshold = math.Nextafter32(10, 0)
				case 1:
					threshold = math.Nextafter32(10, float32(math.Inf(1)))
				case 2:
					threshold = math.MaxFloat32
				}

				if axis == 0 {
					jd.Base.ForceThreshold = threshold
				} else {
					jd.Base.TorqueThreshold = threshold
				}

				j := e.w.CreateWeldJoint(&jd)
				e.r.F32("threshold", threshold)

				for range 4 {
					b.ApplyForceToCenter(box2d.Vec2{X: 10}, true)
					b.ApplyTorque(10, true)
					e.step(b)

					force := j.ConstraintForce()
					e.r.Vec2("joint.force", force.X, force.Y)
					e.r.F32("joint.torque", j.ConstraintTorque())
				}

				b.SetAwake(false)
				e.step(b)
			})
		}
	}
}

type evCall struct {
	kind   int
	a, b   box2d.Shape
	p      box2d.Pos
	n      box2d.Vec2
	fa, fb float32
	ma, mb uint64
	points int
}

func evCallCompare(a, b evCall) int {
	if c := cmp.Compare(a.kind, b.kind); c != 0 {
		return c
	}

	if c := evShapeCompare(a.a, b.a); c != 0 {
		return c
	}

	if c := evShapeCompare(a.b, b.b); c != 0 {
		return c
	}

	if c := cmp.Compare(a.p.X, b.p.X); c != 0 {
		return c
	}

	if c := cmp.Compare(a.p.Y, b.p.Y); c != 0 {
		return c
	}

	if c := cmp.Compare(a.n.X, b.n.X); c != 0 {
		return c
	}

	if c := cmp.Compare(a.n.Y, b.n.Y); c != 0 {
		return c
	}

	if c := cmp.Compare(a.fa, b.fa); c != 0 {
		return c
	}

	if c := cmp.Compare(a.fb, b.fb); c != 0 {
		return c
	}

	if c := cmp.Compare(a.ma, b.ma); c != 0 {
		return c
	}

	if c := cmp.Compare(a.mb, b.mb); c != 0 {
		return c
	}

	return cmp.Compare(a.points, b.points)
}

func (e *evOracle) filter(a, b box2d.Shape) bool {
	e.calls = append(e.calls, evCall{kind: 1, a: a, b: b})
	return !e.reject
}

func (e *evOracle) pre(a, b box2d.Shape, m *box2d.Manifold) {
	e.calls = append(e.calls, evCall{kind: 2, a: a, b: b, n: m.Normal, points: m.PointCount})
	if e.reject {
		m.PointCount = 0
	}
}

func (e *evOracle) continuous(a, b box2d.Shape, p box2d.Pos, n box2d.Vec2) bool {
	e.calls = append(e.calls, evCall{kind: 3, a: a, b: b, p: p, n: n})
	return !e.reject
}

func (e *evOracle) friction(a float32, ma uint64, b float32, mb uint64) float32 {
	e.calls = append(e.calls, evCall{kind: 4, fa: a, fb: b, ma: ma, mb: mb})
	if ma == 11 || mb == 11 {
		return 0
	}

	return 1
}

func (e *evOracle) restitution(a float32, ma uint64, b float32, mb uint64) float32 {
	e.calls = append(e.calls, evCall{kind: 5, fa: a, fb: b, ma: ma, mb: mb})
	if ma == 11 || mb == 11 {
		return 1
	}

	return 0
}

func (e *evOracle) flushCalls() {
	slices.SortFunc(e.calls, evCallCompare)
	r := e.r
	r.I32("callback.count", int32(len(e.calls)))

	for _, c := range e.calls {
		r.I32("callback.kind", int32(c.kind))

		if c.kind <= 3 {
			e.shape(c.a)
			e.shape(c.b)
		}

		r.Pos("callback.point", c.p.X, c.p.Y)
		r.Vec2("callback.normal", c.n.X, c.n.Y)
		r.F32("callback.a", c.fa)
		r.F32("callback.b", c.fb)
		r.U64("callback.materialA", c.ma)
		r.U64("callback.materialB", c.mb)
		r.I32("callback.points", int32(c.points))
	}

	e.calls = nil
}

func TestEventsFilterEligibilityParity(t *testing.T) {
	for row := range 6 {
		for eligible := range 4 {
			for sensor := range 2 {
				for reject := range 2 {
					name := fmt.Sprintf("Filter-%d-%d-%d-%d", row, eligible, sensor, reject)
					t.Run(name, func(t *testing.T) {
						e := evStart(t, name, box2d.Vec2{}, false)
						e.reject = reject != 0
						e.w.SetCustomFilterCallback(e.filter)
						a := e.body(1, box2d.StaticBody, box2d.Pos{})
						b := e.body(2, box2d.DynamicBody, box2d.Pos{Y: .8})
						da, db := evFilterDefinitions(row, eligible, sensor)
						e.circle(a, 10, .5, &da)
						e.circle(b, 20, .5, &db)

						for range 3 {
							e.step(b)
							e.flushCalls()
						}
					})
				}
			}
		}
	}
}

func TestEventsQueryFilterParity(t *testing.T) {
	e := evStart(t, "QueryFilter", box2d.Vec2{}, false)
	a := e.body(1, box2d.StaticBody, box2d.Pos{})

	d := box2d.DefaultShapeDef()
	for i := range 3 {
		d.Filter.CategoryBits = 1 << i

		d.Filter.MaskBits = 3
		if i == 2 {
			d.Filter.MaskBits = 0
		}

		d.Filter.GroupIndex = -7
		if i == 1 {
			d.Filter.GroupIndex = 7
		}

		e.circle(a, 10+i, .5, &d)
	}

	for i := range 4 {
		q := box2d.DefaultQueryFilter()

		q.CategoryBits = 1
		if i == 3 {
			q.CategoryBits = 4
		}

		q.MaskBits = 7

		switch i {
		case 0:
			q.MaskBits = 0
		case 1:
			q.MaskBits = 1
		}

		var ids []int

		e.w.OverlapAABB(box2d.Pos{}, box2d.AABB{LowerBound: box2d.Vec2{X: -1, Y: -1}, UpperBound: box2d.Vec2{X: 1, Y: 1}}, q, func(s box2d.Shape) bool { ids = append(ids, int(s.ID()>>32)); return true })
		slices.Sort(ids)
		e.r.I32("query.count", int32(len(ids)))

		for _, id := range ids {
			e.r.I32("query.index", int32(id))
		}
	}
}

func TestEventsCallbackPolicyParity(t *testing.T) {
	for kind := range 4 {
		for mode := range 3 {
			for domain := range 4 {
				name := fmt.Sprintf("Policy-%d-%d-%d", kind, mode, domain)
				t.Run(name, func(t *testing.T) {
					e := evStart(t, name, box2d.Vec2{Y: -10}, true)

					e.installCallbackPolicy(kind, mode)

					b, steps := e.callbackPolicyScene(kind, domain)

					for range steps {
						e.step(b)
						e.flushCalls()
					}
				})
			}
		}
	}
}

// Native supplemental: Go iterator early termination has no C array analog.
// Every iterator is exercised on a populated event family, then full collection
// and the subsequent world state are compared to an untouched twin world.
func TestEventsEarlyStopNativeSupplemental(t *testing.T) {
	for _, family := range []string{"contact", "sensor", "hit", "move", "joint"} {
		t.Run(family, func(t *testing.T) {
			w, bodies := evEarlyStopWorld(t, family)
			twin, twinBodies := evEarlyStopWorld(t, family)
			check := func(before, after any) {
				t.Helper()

				if !reflect.DeepEqual(before, after) {
					t.Fatal("early stop changed subsequent full collection")
				}
			}

			evCheckContactDataEarlyStop(t, family, bodies)

			switch family {
			case "contact":
				evCheckEventEarlyStop(t, w.ContactBeginEvents, "begin")
			case "sensor":
				evCheckEventEarlyStop(t, w.SensorBeginEvents, "sensor")
			case "hit":
				evCheckEventEarlyStop(t, w.ContactHitEvents, "hit")
			case "move":
				evCheckEventEarlyStop(t, w.BodyMoveEvents, "move")
			case "joint":
				evCheckEventEarlyStop(t, w.JointEvents, "joint")
			}

			evCheckEndEventEarlyStop(t, family, w, twin, bodies, twinBodies)

			for range 3 {
				w.Step(1.0/60, 4)
				twin.Step(1.0/60, 4)

				for i, b := range bodies {
					check(b.Transform(), twinBodies[i].Transform())
					check(b.LinearVelocity(), twinBodies[i].LinearVelocity())
					check(b.AngularVelocity(), twinBodies[i].AngularVelocity())
					check(b.Awake(), twinBodies[i].Awake())
				}
			}
		})
	}
}

func TestEventsSleepingDestroyParity(t *testing.T) {
	e := evStart(t, "SleepingDestroy", box2d.Vec2{}, true)
	a := e.body(1, box2d.StaticBody, box2d.Pos{})
	b := e.body(2, box2d.DynamicBody, box2d.Pos{Y: .4})
	d := box2d.DefaultShapeDef()
	d.IsSensor = true
	d.EnableSensorEvents = true
	e.circle(a, 10, 1, &d)
	d.IsSensor = false
	d.Density = 1
	visitor := e.circle(b, 20, .25, &d)
	e.step(b)
	b.SetAwake(false)
	e.step(b)
	visitor.Destroy(true)
	e.step(b)
	e.circle(b, 21, .25, &d)
	e.step(b)
	b.SetAwake(false)
	e.step(b)
	b.Destroy()
	e.step(b)
	b = e.body(3, box2d.DynamicBody, box2d.Pos{Y: .4})
	e.circle(b, 30, .25, &d)
	e.step(b)
	b.SetAwake(false)
	e.step(b)
	a.Destroy()
	e.step(b)
	e.step(b)
}

func evCheckEventEarlyStop[T any](t *testing.T, events func() iter.Seq[T], setup string) {
	t.Helper()

	full := slices.Collect(events())
	if len(full) != 2 {
		t.Fatalf("%s setup: got %d, want 2", setup, len(full))
	}

	for range events() {
		break
	}

	if !reflect.DeepEqual(full, slices.Collect(events())) {
		t.Fatal("early stop changed subsequent full collection")
	}
}

func evEarlyStopWorld(t *testing.T, family string) (*box2d.World, []box2d.Body) {
	t.Helper()

	wd := box2d.DefaultWorldDef()
	wd.Gravity = box2d.Vec2{}
	wd.EnableSleep = false
	w := box2d.NewWorld(&wd)
	box2d.CleanupWorld(t, w)

	bodies := make([]box2d.Body, 0, 2)

	for i := range 2 {
		bd := box2d.DefaultBodyDef()
		if i != 0 {
			bd.Position.X = 3
		}

		a := w.CreateBody(&bd)
		bd.Type = box2d.DynamicBody
		bd.Position.Y = .99
		b := w.CreateBody(&bd)
		bodies = append(bodies, b)
		sd := box2d.DefaultShapeDef()
		sd.EnableContactEvents = true
		sd.EnableSensorEvents = true
		sd.EnableHitEvents = true
		sd.IsSensor = family == "sensor"
		circle := box2d.Circle{Radius: .5}
		a.CreateCircleShape(&sd, &circle)
		sd.IsSensor = false
		sd.Density = 1
		b.CreateCircleShape(&sd, &circle)

		if family == "joint" {
			jd := box2d.DefaultWeldJointDef()
			jd.Base.BodyA = a
			jd.Base.BodyB = b
			jd.Base.ForceThreshold = 0
			w.CreateWeldJoint(&jd)
		}

		if family == "hit" {
			b.SetLinearVelocity(box2d.Vec2{Y: -5})
		}
	}

	w.Step(1.0/60, 4)

	return w, bodies
}

func (e *evOracle) observeBodyStates(bodies []box2d.Body) {
	r := e.r
	r.I32("state.count", int32(len(bodies)))

	for _, b := range bodies {
		e.bodyID(b)

		if !b.IsValid() {
			continue
		}

		position, rotation, velocity := b.Position(), b.Rotation(), b.LinearVelocity()
		r.Pos("state.position", position.X, position.Y)
		r.Rot("state.rotation", rotation.C, rotation.S)
		r.Vec2("state.velocity", velocity.X, velocity.Y)
		r.F32("state.spin", b.AngularVelocity())
		r.Bool("state.awake", b.Awake())
		r.Bool("state.enabled", b.Enabled())
		data := slices.Collect(b.ContactData())
		slices.SortFunc(data, func(a, b box2d.ContactData) int { return evPairCompare(a.ShapeA, a.ShapeB, b.ShapeA, b.ShapeB) })
		r.I32("state.contacts", int32(len(data)))

		for _, c := range data {
			e.shape(c.ShapeA)
			e.shape(c.ShapeB)
			e.contact(c.Contact)
			m := c.Manifold
			r.Vec2("manifold.normal", m.Normal.X, m.Normal.Y)
			r.F32("manifold.rollingImpulse", m.RollingImpulse)
			r.I32("manifold.count", int32(m.PointCount))

			hit := c.ShapeA.HitEventsEnabled() || c.ShapeB.HitEventsEnabled()
			r.Bool("manifold.hitEnabled", hit)

			for p := range m.PointCount {
				mp := m.Points[p]
				r.Vec2("manifold.anchorA", mp.AnchorA.X, mp.AnchorA.Y)
				r.Vec2("manifold.anchorB", mp.AnchorB.X, mp.AnchorB.Y)
				r.F32("manifold.separation", mp.Separation)
				r.F32("manifold.normalImpulse", mp.NormalImpulse)
				r.F32("manifold.tangentImpulse", mp.TangentImpulse)
				r.F32("manifold.totalImpulse", mp.TotalNormalImpulse)

				if hit {
					r.F32("manifold.normalVelocity", mp.NormalVelocity)
				}

				r.U32("manifold.id", uint32(mp.ID))
				r.Bool("manifold.persisted", mp.Persisted)
			}
		}
	}
}

func evCheckContactDataEarlyStop(t *testing.T, family string, bodies []box2d.Body) {
	t.Helper()

	check := func(before, after any) {
		t.Helper()

		if !reflect.DeepEqual(before, after) {
			t.Fatal("early stop changed subsequent full collection")
		}
	}
	// A populated contact iterator is checked before any removal mutation.
	contacts := slices.Collect(bodies[0].ContactData())
	for range bodies[0].ContactData() {
		break
	}

	check(contacts, slices.Collect(bodies[0].ContactData()))

	if family == "contact" || family == "hit" {
		if len(contacts) != 1 {
			t.Fatalf("contact setup: got %d, want 1", len(contacts))
		}
	}
}

func evCheckEndEventEarlyStop(t *testing.T, family string, w, twin *box2d.World, bodies, twinBodies []box2d.Body) {
	t.Helper()

	if family == "contact" || family == "sensor" {
		for i, b := range bodies {
			p := b.Position()
			p.Y = 5
			b.SetTransform(p, box2d.Rot{C: 1})
			twinBodies[i].SetTransform(p, box2d.Rot{C: 1})
		}

		w.Step(1.0/60, 4)
		twin.Step(1.0/60, 4)

		if family == "contact" {
			evCheckEventEarlyStop(t, w.ContactEndEvents, "end")
		} else {
			evCheckEventEarlyStop(t, w.SensorEndEvents, "sensor end")
		}
	}
}

func (e *evOracle) observeOwnedEvents(bm []box2d.BodyMoveEvent, je []box2d.JointEvent) {
	r := e.r
	r.I32("move.count", int32(len(bm)))

	for _, m := range bm {
		e.bodyID(m.Body)

		tag, ok := m.UserData.(int)
		if !ok {
			e.t.Fatalf("move user data type = %T, want int", m.UserData)
		}

		r.I32("move.tag", int32(tag))
		r.Pos("move.position", oracle.Real(m.Transform.P.X), oracle.Real(m.Transform.P.Y))
		r.Rot("move.rotation", m.Transform.Q.C, m.Transform.Q.S)
		r.Bool("move.sleep", m.FellAsleep)
	}

	r.I32("joint.count", int32(len(je)))

	for _, j := range je {
		id := j.Joint.ID()
		r.I32("joint.index", int32(id>>32))
		r.U32("joint.generation", uint32(uint16(id)))
		r.Bool("joint.world", j.Joint.World() == e.w)
		r.Bool("joint.valid", j.Joint.IsValid())

		tag, ok := j.UserData.(int)
		if !ok {
			e.t.Fatalf("joint user data type = %T, want int", j.UserData)
		}

		r.I32("joint.tag", int32(tag))
	}
}

func (e *evOracle) installCallbackPolicy(kind, mode int) {
	e.reject = mode == 1
	if kind == 0 {
		e.w.SetPreSolveCallback(e.pre, nil)
	}

	if kind == 1 {
		e.w.SetPreSolveCallback(nil, e.continuous)
	}

	if kind == 2 && mode != 0 {
		e.w.SetFrictionCallback(e.friction)
	}

	if kind == 3 && mode != 0 {
		e.w.SetRestitutionCallback(e.restitution)
	}

	if mode == 2 {
		e.w.SetPreSolveCallback(nil, nil)
		e.w.SetFrictionCallback(nil)
		e.w.SetRestitutionCallback(nil)
	}
}

func (e *evOracle) callbackPolicyScene(kind, domain int) (box2d.Body, int) {
	a := e.body(1, box2d.StaticBody, box2d.Pos{})

	position := box2d.Pos{Y: 1.2}
	if kind == 1 {
		position = box2d.Pos{X: -2}
	}

	b := e.body(2, box2d.DynamicBody, position)
	d := box2d.DefaultShapeDef()
	d.Material.Friction = .8
	d.Material.Restitution = .2
	d.Material.UserMaterialID = 11
	d.EnablePreSolveEvents = domain != 3
	d.EnableContactEvents = true
	d.EnableSensorEvents = true

	d.IsSensor = domain == 1
	if kind == 1 {
		e.circle(a, 10, .2, &d)
	} else {
		polygon := box2d.MakeBox(10, .5)
		s := a.CreatePolygonShape(&d, &polygon)
		e.tags[s] = 10
	}

	d.IsSensor = false
	d.Density = 1
	d.Material.Friction = .2
	d.Material.Restitution = .8
	d.Material.UserMaterialID = 22
	e.circle(b, 20, .5, &d)

	steps := 30

	if kind == 1 {
		b.SetBullet(true)
		b.SetGravityScale(0)
		b.SetLinearVelocity(box2d.Vec2{X: 150})

		steps = 3
	} else {
		b.SetLinearVelocity(box2d.Vec2{X: 4, Y: -4})
	}

	if domain == 2 {
		b.SetAwake(false)
	}

	return b, steps
}

func (e *evOracle) hitThresholdScene(flags, thresholdCase, side int) box2d.Body {
	threshold := float32(2)
	if thresholdCase != 0 {
		threshold = 5
	}

	e.w.SetHitEventThreshold(threshold)
	e.w.SetContinuousEnabled(false)
	a := e.body(1, box2d.StaticBody, box2d.Pos{})
	b := e.body(2, box2d.DynamicBody, box2d.Pos{Y: 1.001})
	d := box2d.DefaultShapeDef()
	d.EnableHitEvents = flags&1 != 0
	d.EnableContactEvents = true
	e.circle(a, 10, .5, &d)
	d.Density = 1
	d.EnableHitEvents = flags&2 != 0
	e.circle(b, 20, .5, &d)

	speed := threshold
	if side < 0 {
		speed = math.Nextafter32(threshold, 0)
	} else if side > 0 {
		speed = math.Nextafter32(threshold, float32(math.Inf(1)))
	}

	e.r.F32("input.speed", speed)
	b.SetLinearVelocity(box2d.Vec2{Y: -speed})

	return b
}

func evFilterDefinitions(row, eligible, sensor int) (box2d.ShapeDef, box2d.ShapeDef) {
	da, db := box2d.DefaultShapeDef(), box2d.DefaultShapeDef()
	da.IsSensor = sensor != 0
	da.EnableSensorEvents = true
	db.EnableSensorEvents = true
	da.EnableContactEvents = true
	db.EnableContactEvents = true
	da.EnableCustomFiltering = eligible&1 != 0
	db.EnableCustomFiltering = eligible&2 != 0
	da.Filter.CategoryBits = 1
	db.Filter.CategoryBits = 2
	da.Filter.MaskBits = 2
	db.Filter.MaskBits = 1

	switch row {
	case 1:
		da.Filter.MaskBits = 0
	case 2:
		db.Filter.MaskBits = 0
	case 3:
		da.Filter.MaskBits = 0
		db.Filter.MaskBits = 0
		da.Filter.GroupIndex = 7
		db.Filter.GroupIndex = 7
	case 4:
		da.Filter.GroupIndex = -7
		db.Filter.GroupIndex = -7
	case 5:
		da.Filter.MaskBits = 4
		db.Filter.MaskBits = 4
	}

	db.Density = 1

	return da, db
}
