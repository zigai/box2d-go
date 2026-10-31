package box2d

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

type dynamicsOracle struct {
	*oracle.Reader

	t           *testing.T
	expected    [][36]uint64
	observation int
}

func dynamicsOpen(t *testing.T, name string) *dynamicsOracle {
	t.Helper()
	r := &dynamicsOracle{Reader: oracle.Open(t, name), t: t}

	precision := "float"
	if oracle.DoublePrecision {
		precision = "double"
	}

	data, err := oracle.ReadFile(precision, name)
	if err != nil {
		t.Fatal(err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))

	for {
		var record restitutionRecord

		err := decoder.Decode(&record)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasSuffix(record.Key, "/digests") {
			continue
		}

		raw, err := hex.DecodeString(record.Value)
		if err != nil {
			t.Fatal(err)
		}

		if len(raw) != 36*8 {
			t.Fatal("invalid dynamics short row width")
		}

		var row [36]uint64
		for i := range row {
			row[i] = binary.LittleEndian.Uint64(raw[8*i:])
		}

		r.expected = append(r.expected, row)
	}

	r.Bytes("dynamics/schema", []byte(DynamicsSchema))
	t.Cleanup(func() {
		if !t.Failed() && r.observation != len(r.expected) {
			t.Errorf("dynamics observations C=%d Go=%d", len(r.expected), r.observation)
		}
	})

	return r
}

func dynamicsWorld(t *testing.T, def WorldDef) *World {
	t.Helper()

	w := NewWorld(&def)
	CleanupWorld(t, w)

	return w
}

func dynamicsBox(w *World, typ BodyType, x, y float32, hx, hy float32, material SurfaceMaterial) (Body, Shape) {
	bd := DefaultBodyDef()
	bd.Type = typ
	bd.Position = Pos{oracle.Real(x), oracle.Real(y)}
	b := w.CreateBody(&bd)
	sd := DefaultShapeDef()
	sd.Material = material
	sd.EnableContactEvents = true
	sd.EnableHitEvents = true
	box := MakeBox(hx, hy)

	return b, b.CreatePolygonShape(&sd, &box)
}

func dynamicsSteps(r *dynamicsOracle, w *World, n int, prefix string) {
	for i := range n {
		w.Step(1.0/60.0, 4)
		dynamicsShort(r, w, fmt.Sprintf("%sstep/%d/", prefix, i+1), true)
	}
}

func dynamicsRestitution(a float32, ia uint64, b float32, ib uint64) float32 {
	if ia == 7 || ib == 7 {
		return .35
	}

	return .65
}

func dynamicsFriction(a float32, ia uint64, b float32, ib uint64) float32 {
	if ia == 7 || ib == 7 {
		return .1
	}

	return .8
}

func TestDynamicsIntegration(t *testing.T) {
	// Loaded step effects complement the direct body API table, including locks and cap neighbors.
	for partition := range 12 {
		t.Run(fmt.Sprintf("partition%d", partition), func(t *testing.T) {
			r := dynamicsOpen(t, fmt.Sprintf("dynamics/integration%d", partition))
			wd := DefaultWorldDef()
			wd.Gravity = Vec2{2, -10}
			wd.EnableSleep = false
			wd.MaximumLinearSpeed = 10
			w := dynamicsWorld(t, wd)
			m := DefaultSurfaceMaterial()
			b, _ := dynamicsBox(w, DynamicBody, 3, 4, .6, .4, m)
			b.SetLinearVelocity(Vec2{3, 2})
			b.SetAngularVelocity(2)

			switch partition {
			case 0:
				b.SetGravityScale(0)
			case 1:
				b.SetGravityScale(-1)
			case 2:
				b.SetLinearDamping(.7)
				b.SetAngularDamping(1.3)
			case 3:
				b.ApplyForce(Vec2{8, 11}, Pos{4, 5}, true)
			case 4:
				b.ApplyTorque(9, true)
			case 5:
				b.ApplyLinearImpulse(Vec2{2, 3}, Pos{4, 5}, true)
			case 6:
				b.ApplyAngularImpulse(3, true)
			case 7:
				b.SetMotionLocks(MotionLocks{LinearX: true})
			case 8:
				b.SetMotionLocks(MotionLocks{LinearY: true, AngularZ: true})
			case 9, 10, 11:
				speed := dynamicsThresholdNeighbor(10, partition-9)

				b.SetGravityScale(0)
				b.SetLinearVelocity(Vec2{speed, 0})
			}

			dynamicsShort(r, w, "initial/", true)
			dynamicsSteps(r, w, 30, "")
			w.SetGravity(Vec2{-3, 4})
			w.SetMaximumLinearSpeed(5)
			b.SetLinearDamping(.2)
			b.SetAngularDamping(.4)
			dynamicsSteps(r, w, 12, "mutated/")
		})
	}
}

func TestDynamicsContactTuning(t *testing.T) {
	for partition := range 12 {
		t.Run(fmt.Sprintf("partition%d", partition), func(t *testing.T) {
			r := dynamicsOpen(t, fmt.Sprintf("dynamics/contact%d", partition))
			wd := DefaultWorldDef()
			wd.EnableSleep = false
			w := dynamicsWorld(t, wd)
			m := DefaultSurfaceMaterial()
			m.Friction = .6
			_, ground := dynamicsBox(w, StaticBody, 0, -1, 20, 1, m)
			b, shape := dynamicsBox(w, DynamicBody, 0, .51, .5, .5, m)
			dynamicsBox(w, DynamicBody, 0, 1.53, .5, .5, m)
			b.SetLinearVelocity(Vec2{5, 0})
			b.SetAngularVelocity(1)

			dynamicsTuneContactWorld(w, partition)
			dynamicsTuneContactShape(b, shape, &m, partition)

			dynamicsSteps(r, w, 30, "")
			ground.SetFriction(.1)

			m = shape.SurfaceMaterial()
			m.Friction = .9
			m.RollingResistance = .2
			m.TangentSpeed = -2
			shape.SetSurfaceMaterial(&m)
			w.SetWarmStartingEnabled(true)
			w.SetContactTuning(15, .5, 2)
			w.SetFrictionCallback(nil)
			dynamicsSteps(r, w, 30, "mutated/")
		})
	}
}

func TestDynamicsRestitutionBoundary(t *testing.T) {
	dynamicsCheckRestitutionThresholdNeighbors(t)

	dynamicsCheckRestitutionMaterialMixing(t)

	dynamicsCheckRestitutionPropagation(t)
}

func TestDynamicsContinuous(t *testing.T) {
	for target := range 3 {
		for bullet := range 2 {
			for continuous := range 2 {
				for geometry := range 4 {
					name := fmt.Sprintf("target%d_bullet%d_continuous%d_geometry%d", target, bullet, continuous, geometry)
					t.Run(name, func(t *testing.T) {
						r := dynamicsOpen(t, "dynamics/"+name)
						wd := DefaultWorldDef()
						wd.Gravity = Vec2{}
						wd.EnableSleep = false
						wd.EnableContinuous = continuous != 0
						w := dynamicsWorld(t, wd)
						m := DefaultSurfaceMaterial()
						m.Friction = .2

						targetBody, _ := dynamicsBox(w, BodyType(target), 0, 0, .05, 3, m)
						if target != 0 {
							targetBody.SetLinearVelocity(Vec2{-1, 0})
						}

						bd := dynamicsContinuousBodyDef(geometry, bullet)

						b := w.CreateBody(&bd)
						sd := DefaultShapeDef()
						sd.Material = m
						sd.EnableHitEvents = true
						sd.EnableContactEvents = true
						box := MakeBox(.15, .4)
						b.CreatePolygonShape(&sd, &box)
						dynamicsSteps(r, w, 30, "")
						w.SetContinuousEnabled(true)
						b.SetBullet(true)
						b.SetSafetyFactor(.1)
						dynamicsSteps(r, w, 12, "mutated/")
					})
				}
			}
		}
	}
}

func TestDynamicsSleepWake(t *testing.T) {
	for operation := range 12 {
		for wake := range 2 {
			name := fmt.Sprintf("sleep_operation%d_wake%d", operation, wake)
			t.Run(name, func(t *testing.T) {
				r := dynamicsOpen(t, "dynamics/"+name)
				wd := DefaultWorldDef()
				w := dynamicsWorld(t, wd)
				m := DefaultSurfaceMaterial()
				m.Friction = .6
				dynamicsBox(w, StaticBody, 0, -1, 20, 1, m)
				lower, _ := dynamicsBox(w, DynamicBody, 0, .51, .5, .5, m)
				upper, _ := dynamicsBox(w, DynamicBody, 0, 1.53, .5, .5, m)
				dynamicsSettleIsland(r, w, lower, upper)

				dynamicsApplySleepingBodyLoad(lower, operation, wake != 0)
				dynamicsMutateSleepingIsland(w, lower, operation)

				dynamicsShort(r, w, "operation/", true)
				dynamicsSteps(r, w, 12, "after/")
				lower.Enable()
				lower.SetType(DynamicBody)
				lower.SetAwake(true)
				dynamicsSteps(r, w, 12, "restored/")
			})
		}
	}
}

func dynamicsCheckRestitutionThresholdNeighbors(t *testing.T) {
	t.Helper()

	for thresholdCase := range 2 {
		for neighbor := range 3 {
			name := fmt.Sprintf("threshold%d_neighbor%d", thresholdCase, neighbor)
			t.Run(name, func(t *testing.T) {
				r := dynamicsOpen(t, "dynamics/"+name)
				wd := DefaultWorldDef()
				wd.Gravity = Vec2{}
				wd.EnableSleep = false

				wd.RestitutionIterations = 4
				if thresholdCase == 1 {
					wd.RestitutionThreshold = 2.5
				}

				speed := dynamicsThresholdNeighbor(wd.RestitutionThreshold, neighbor)

				w := dynamicsWorld(t, wd)
				m := DefaultSurfaceMaterial()
				m.Friction = 0
				m.Restitution = .8
				dynamicsBox(w, StaticBody, 0, -1, 20, 1, m)

				bd := DefaultBodyDef()
				bd.Type = DynamicBody
				bd.Position = Pos{0, .5}
				bd.LinearVelocity = Vec2{0, -speed}
				b := w.CreateBody(&bd)
				sd := DefaultShapeDef()
				sd.Material = m
				sd.EnableHitEvents = true
				sd.EnableContactEvents = true
				circle := Circle{Radius: .5}
				b.CreateCircleShape(&sd, &circle)
				r.F32("approachSpeed", speed)
				dynamicsSteps(r, w, 60, "")
			})
		}
	}
}

func dynamicsCheckRestitutionMaterialMixing(t *testing.T) {
	t.Helper()

	for mode := range 3 {
		for swapped := range 2 {
			name := fmt.Sprintf("mixing%d_swapped%d", mode, swapped)
			t.Run(name, func(t *testing.T) {
				r := dynamicsOpen(t, "dynamics/"+name)
				wd := DefaultWorldDef()
				wd.Gravity = Vec2{}
				wd.EnableSleep = false
				wd.RestitutionIterations = 4

				w := dynamicsWorld(t, wd)
				if mode != 0 {
					w.SetRestitutionCallback(dynamicsRestitution)
				}

				if mode == 2 {
					w.SetRestitutionCallback(nil)
				}

				a, b := float32(.2), float32(.8)
				if swapped == 1 {
					a, b = b, a
				}

				m := DefaultSurfaceMaterial()
				m.Friction = 0
				m.Restitution = a
				m.UserMaterialID = 7
				dynamicsBox(w, StaticBody, 0, -1, 20, 1, m)
				m.Restitution = b
				m.UserMaterialID = 11
				ball, _ := dynamicsBox(w, DynamicBody, 0, .6, .5, .5, m)
				ball.SetLinearVelocity(Vec2{0, -5})
				dynamicsSteps(r, w, 60, "")
			})
		}
	}
}

func dynamicsCheckRestitutionPropagation(t *testing.T) {
	t.Helper()

	for propagation := range 2 {
		for iterationCase := range 3 {
			for supportsCase := range 2 {
				name := fmt.Sprintf("propagation%d_iterations%d_supports%d", propagation, iterationCase, supportsCase)
				t.Run(name, func(t *testing.T) {
					r := dynamicsOpen(t, "dynamics/"+name)
					wd := DefaultWorldDef()
					wd.Gravity = Vec2{}
					wd.EnableSleep = false
					wd.RestitutionIterations = 1 << iterationCase
					wd.EnableRestitutionPropagation = propagation != 0
					w := dynamicsWorld(t, wd)
					m := DefaultSurfaceMaterial()
					m.Friction = 0
					m.Restitution = .8
					dynamicsBox(w, StaticBody, 0, -1, 20, 1, m)

					supports := 1 + 2*supportsCase
					for i := range supports {
						dynamicsBox(w, DynamicBody, 0, .5+float32(i), .5, .5, m)
					}

					ball, _ := dynamicsBox(w, DynamicBody, 0, float32(supports)+.6, .5, .5, m)
					ball.SetLinearVelocity(Vec2{0, -5})
					dynamicsSteps(r, w, 60, "")
				})
			}
		}
	}
}

func dynamicsContinuousBodyDef(geometry, bullet int) BodyDef {
	x := float32(-2)
	if geometry == 1 {
		x = -.1
	}

	bd := DefaultBodyDef()
	bd.Type = DynamicBody
	bd.Position = Pos{oracle.Real(x), 0}
	bd.LinearVelocity = Vec2{80, 0}

	bd.IsBullet = bullet != 0
	if geometry == 0 {
		bd.SafetyFactor = .01
	} else {
		bd.SafetyFactor = .5
	}

	if geometry == 2 {
		bd.AngularVelocity = 10
	}

	return bd
}

func dynamicsSettleIsland(r *dynamicsOracle, w *World, lower, upper Body) {
	r.t.Helper()

	for i := range 900 {
		w.Step(1.0/60.0, 4)
		dynamicsShort(r, w, fmt.Sprintf("settle/%d/", i+1), i == 0 || (i+1)%60 == 0)

		if !lower.Awake() && !upper.Awake() {
			r.I32("sleepStep", int32(i+1))
			return
		}
	}

	r.t.Fatal("loaded island did not sleep within source locator bound 900")
}

func dynamicsThresholdNeighbor(value float32, neighbor int) float32 {
	if neighbor == 0 {
		return math.Nextafter32(value, float32(math.Inf(-1)))
	}

	if neighbor == 2 {
		return math.Nextafter32(value, float32(math.Inf(1)))
	}

	return value
}

func dynamicsTuneContactWorld(w *World, partition int) {
	switch partition {
	case 0:
		w.SetWarmStartingEnabled(false)
	case 1:
		w.SetContactTuning(5, .2, .5)
	case 2:
		w.SetContactTuning(120, 3, 10)
	case 6:
		w.SetFrictionCallback(dynamicsFriction)
	case 9:
		w.SetContactTuning(30, 1, 0)
	case 10:
		w.SetMaximumLinearSpeed(2)
	case 11:
		w.SetGravity(Vec2{2, -5})
	}
}

func dynamicsTuneContactShape(b Body, shape Shape, m *SurfaceMaterial, partition int) {
	switch partition {
	case 3:
		shape.SetFriction(0)
	case 4:
		m.RollingResistance = .7
		shape.SetSurfaceMaterial(m)
	case 5:
		m.TangentSpeed = 3
		shape.SetSurfaceMaterial(m)
	case 7:
		m.RollingResistance = .4
		shape.SetSurfaceMaterial(m)
	case 8:
		b.SetContactRecyclingEnabled(false)
	}
}

func dynamicsApplySleepingBodyLoad(lower Body, operation int, wake bool) {
	switch operation {
	case 0:
		lower.ApplyForceToCenter(Vec2{20, 5}, wake)
	case 1:
		lower.ApplyForce(Vec2{20, 5}, Pos{1, 1}, wake)
	case 2:
		lower.ApplyTorque(10, wake)
	case 3:
		lower.ApplyLinearImpulseToCenter(Vec2{2, 3}, wake)
	case 4:
		lower.ApplyLinearImpulse(Vec2{2, 3}, Pos{1, 1}, wake)
	case 5:
		lower.ApplyAngularImpulse(2, wake)
	}
}

func dynamicsMutateSleepingIsland(w *World, lower Body, operation int) {
	switch operation {
	case 6:
		lower.SetAwake(true)
	case 7:
		lower.SetSleepEnabled(false)
	case 8:
		w.SetSleepingEnabled(false)
	case 9:
		lower.Disable()
	case 10:
		lower.SetType(KinematicBody)
		lower.SetLinearVelocity(Vec2{1, 0})
	case 11:
		lower.SetTransform(Pos{1, .5}, Rot{C: 1})
	}
}

// ObserveDynamics observes post-step public state, never pre-solve scratch impulses.
// Handle generations are retained; world membership is relative to the tested world.
func ObserveDynamics(r *oracle.Reader, id *World, prefix string, direct bool) [36]uint64 {
	var h [36]uint64
	for i := range h {
		h[i] = oracle.HashInit
	}

	w := id

	h[0] = oracle.HashU32(h[0], uint32(int32(id.Counters().BodyCount)))
	if direct {
		r.I32(prefix+"bodyCount", int32(id.Counters().BodyCount))
	}

	for i := range w.bodies {
		if w.bodies[i].id != i {
			continue
		}

		b := makeBodyID(w, i)

		key := ""
		if direct {
			key = fmt.Sprintf("%sbody/%d/", prefix, i)
		}

		dynamicsBodyState(r, key, b, w, &h, direct)

		dynamicsBodyContacts(r, key, b, w, &h, direct)
	}

	dynamicsMoveEvents(r, id, prefix, &h, direct)

	dynamicsContactTransitions(r, id, prefix, &h, direct)

	dynamicsHitEvents(r, id, prefix, &h, direct)

	dynamicsSensorTransitions(r, id, prefix, &h, direct)

	h[35] = id.StateHash()

	return h
}

func dynamicsShape(r *oracle.Reader, key string, s Shape, worldID *World, h *[36]uint64, direct bool) {
	h[1] = oracle.HashU32(h[1], uint32(s.index1))
	if direct {
		r.I32(key+"index", s.index1)
	}

	h[2] = oracle.HashU32(h[2], uint32(int32(s.generation)))
	if direct {
		r.I32(key+"generation", int32(s.generation))
	}

	h[3] = dynamicsBool(h[3], s.world == worldID)
	if direct {
		r.Bool(key+"world", s.world == worldID)
	}

	h[34] = dynamicsBool(h[34], s.IsValid())
	if direct {
		r.Bool(key+"valid", s.IsValid())
	}
}

func dynamicsContact(r *oracle.Reader, key string, c Contact, a, b Shape, worldID *World, h *[36]uint64, direct bool) {
	h[1] = oracle.HashU32(h[1], uint32(c.index1))
	if direct {
		r.I32(key+"index", c.index1)
	}

	h[2] = oracle.HashU64(h[2], uint64(c.generation))
	if direct {
		r.U64(key+"generation", uint64(c.generation))
	}

	h[3] = dynamicsBool(h[3], c.world == worldID)
	if direct {
		r.Bool(key+"world", c.world == worldID)
	}

	h[34] = dynamicsBool(h[34], c.IsValid())
	if direct {
		r.Bool(key+"valid", c.IsValid())
	}

	dynamicsShape(r, key+"shapeA/", a, worldID, h, direct)
	dynamicsShape(r, key+"shapeB/", b, worldID, h, direct)
}

func dynamicsBool(h uint64, value bool) uint64 {
	if value {
		return oracle.HashU32(h, 1)
	}

	return oracle.HashU32(h, 0)
}

func dynamicsBodyState(r *oracle.Reader, key string, b Body, w *World, h *[36]uint64, direct bool) {
	h[1] = oracle.HashU32(h[1], uint32(b.index1))
	if direct {
		r.I32(key+"index", b.index1)
	}

	h[2] = oracle.HashU32(h[2], uint32(int32(b.generation)))
	if direct {
		r.I32(key+"generation", int32(b.generation))
	}

	h[3] = dynamicsBool(h[3], b.world == w)
	if direct {
		r.Bool(key+"world", b.world == w)
	}

	p, q, v := b.Position(), b.Rotation(), b.LinearVelocity()

	h[4] = oracle.HashPos(h[4], p.X, p.Y)
	if direct {
		r.Pos(key+"position", p.X, p.Y)
	}

	h[5] = oracle.HashRot(h[5], q.C, q.S)
	if direct {
		r.Rot(key+"rotation", q.C, q.S)
	}

	h[6] = oracle.HashRot(h[6], v.X, v.Y)
	if direct {
		r.Vec2(key+"velocity", v.X, v.Y)
	}

	h[7] = oracle.HashF32(h[7], b.AngularVelocity())
	if direct {
		r.F32(key+"angularVelocity", b.AngularVelocity())
	}

	h[8] = dynamicsBool(h[8], b.Awake())
	if direct {
		r.Bool(key+"awake", b.Awake())
	}

	h[9] = dynamicsBool(h[9], b.Enabled())
	if direct {
		r.Bool(key+"enabled", b.Enabled())
	}

	h[10] = oracle.HashU32(h[10], uint32(int32(b.Type())))
	if direct {
		r.I32(key+"type", int32(b.Type()))
	}
}

func dynamicsManifoldPoint(r *oracle.Reader, pk string, p ManifoldPoint, h *[36]uint64, direct bool) {
	h[14] = oracle.HashRot(h[14], p.AnchorA.X, p.AnchorA.Y)
	if direct {
		r.Vec2(pk+"anchorA", p.AnchorA.X, p.AnchorA.Y)
	}

	h[15] = oracle.HashRot(h[15], p.AnchorB.X, p.AnchorB.Y)
	if direct {
		r.Vec2(pk+"anchorB", p.AnchorB.X, p.AnchorB.Y)
	}

	h[16] = oracle.HashF32(h[16], p.Separation)
	if direct {
		r.F32(pk+"separation", p.Separation)
	}

	h[17] = oracle.HashF32(h[17], p.NormalImpulse)
	if direct {
		r.F32(pk+"normalImpulse", p.NormalImpulse)
	}

	h[18] = oracle.HashF32(h[18], p.TangentImpulse)
	if direct {
		r.F32(pk+"tangentImpulse", p.TangentImpulse)
	}

	h[19] = oracle.HashF32(h[19], p.TotalNormalImpulse)
	if direct {
		r.F32(pk+"totalNormalImpulse", p.TotalNormalImpulse)
	}

	h[20] = oracle.HashF32(h[20], p.NormalVelocity)
	if direct {
		r.F32(pk+"normalVelocity", p.NormalVelocity)
	}

	h[21] = oracle.HashF32(h[21], p.BaseSeparation)
	if direct {
		r.F32(pk+"baseSeparation", p.BaseSeparation)
	}

	h[22] = oracle.HashU32(h[22], uint32(int32(p.ID)))
	if direct {
		r.I32(pk+"id", int32(p.ID))
	}

	h[23] = dynamicsBool(h[23], p.Persisted)
	if direct {
		r.Bool(pk+"persisted", p.Persisted)
	}
}

func dynamicsBodyContacts(r *oracle.Reader, key string, b Body, w *World, h *[36]uint64, direct bool) {
	count := 0

	for c := range b.ContactData() {
		ck := ""
		if direct {
			ck = fmt.Sprintf("%scontact/%d/", key, count)
		}

		dynamicsContact(r, ck, c.Contact, c.ShapeA, c.ShapeB, w, h, direct)
		m := c.Manifold

		h[11] = oracle.HashRot(h[11], m.Normal.X, m.Normal.Y)
		if direct {
			r.Vec2(ck+"normal", m.Normal.X, m.Normal.Y)
		}

		h[12] = oracle.HashF32(h[12], m.RollingImpulse)
		if direct {
			r.F32(ck+"rollingImpulse", m.RollingImpulse)
		}

		h[13] = oracle.HashU32(h[13], uint32(int32(m.PointCount)))
		if direct {
			r.I32(ck+"pointCount", int32(m.PointCount))
		}

		for j := range m.PointCount {
			pk := ""
			if direct {
				pk = fmt.Sprintf("%spoint/%d/", ck, j)
			}

			dynamicsManifoldPoint(r, pk, m.Points[j], h, direct)
		}

		count++
	}

	h[24] = oracle.HashU32(h[24], uint32(int32(count)))
	if direct {
		r.I32(key+"contactCount", int32(count))
	}
}

func dynamicsMoveEvents(r *oracle.Reader, id *World, prefix string, h *[36]uint64, direct bool) {
	w := id
	count := 0

	for e := range id.BodyMoveEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%smove/%d/", prefix, count)
		}

		h[1] = oracle.HashU32(h[1], uint32(e.Body.index1))
		if direct {
			r.I32(key+"index", e.Body.index1)
		}

		h[2] = oracle.HashU32(h[2], uint32(int32(e.Body.generation)))
		if direct {
			r.I32(key+"generation", int32(e.Body.generation))
		}

		h[3] = dynamicsBool(h[3], e.Body.world == w)
		if direct {
			r.Bool(key+"world", e.Body.world == w)
		}

		h[4] = oracle.HashPos(h[4], e.Transform.P.X, e.Transform.P.Y)
		if direct {
			r.Pos(key+"position", e.Transform.P.X, e.Transform.P.Y)
		}

		h[5] = oracle.HashRot(h[5], e.Transform.Q.C, e.Transform.Q.S)
		if direct {
			r.Rot(key+"rotation", e.Transform.Q.C, e.Transform.Q.S)
		}

		h[25] = dynamicsBool(h[25], e.FellAsleep)
		if direct {
			r.Bool(key+"fellAsleep", e.FellAsleep)
		}

		count++
	}

	h[26] = oracle.HashU32(h[26], uint32(int32(count)))
	if direct {
		r.I32(prefix+"moveCount", int32(count))
	}
}

func dynamicsContactTransitions(r *oracle.Reader, id *World, prefix string, h *[36]uint64, direct bool) {
	w := id
	count := 0

	for e := range id.ContactBeginEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%sbegin/%d/", prefix, count)
		}

		dynamicsContact(r, key, e.Contact, e.ShapeA, e.ShapeB, w, h, direct)

		count++
	}

	h[27] = oracle.HashU32(h[27], uint32(int32(count)))
	if direct {
		r.I32(prefix+"beginCount", int32(count))
	}

	count = 0

	for e := range id.ContactEndEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%send/%d/", prefix, count)
		}

		dynamicsContact(r, key, e.Contact, e.ShapeA, e.ShapeB, w, h, direct)

		count++
	}

	h[28] = oracle.HashU32(h[28], uint32(int32(count)))
	if direct {
		r.I32(prefix+"endCount", int32(count))
	}
}

func dynamicsHitEvents(r *oracle.Reader, id *World, prefix string, h *[36]uint64, direct bool) {
	w := id
	count := 0

	for e := range id.ContactHitEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%shit/%d/", prefix, count)
		}

		dynamicsContact(r, key, e.Contact, e.ShapeA, e.ShapeB, w, h, direct)

		h[29] = oracle.HashPos(h[29], e.Point.X, e.Point.Y)
		if direct {
			r.Pos(key+"point", e.Point.X, e.Point.Y)
		}

		h[11] = oracle.HashRot(h[11], e.Normal.X, e.Normal.Y)
		if direct {
			r.Vec2(key+"normal", e.Normal.X, e.Normal.Y)
		}

		h[30] = oracle.HashF32(h[30], e.ApproachSpeed)
		if direct {
			r.F32(key+"approachSpeed", e.ApproachSpeed)
		}

		count++
	}

	h[31] = oracle.HashU32(h[31], uint32(int32(count)))
	if direct {
		r.I32(prefix+"hitCount", int32(count))
	}
}

func dynamicsSensorTransitions(r *oracle.Reader, id *World, prefix string, h *[36]uint64, direct bool) {
	w := id
	count := 0

	for e := range id.SensorBeginEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%ssensorBegin/%d/", prefix, count)
		}

		dynamicsShape(r, key+"sensor/", e.SensorShape, w, h, direct)
		dynamicsShape(r, key+"visitor/", e.VisitorShape, w, h, direct)

		count++
	}

	h[32] = oracle.HashU32(h[32], uint32(int32(count)))
	if direct {
		r.I32(prefix+"sensorBeginCount", int32(count))
	}

	count = 0

	for e := range id.SensorEndEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%ssensorEnd/%d/", prefix, count)
		}

		dynamicsShape(r, key+"sensor/", e.SensorShape, w, h, direct)
		dynamicsShape(r, key+"visitor/", e.VisitorShape, w, h, direct)

		count++
	}

	h[33] = oracle.HashU32(h[33], uint32(int32(count)))
	if direct {
		r.I32(prefix+"sensorEndCount", int32(count))
	}
}

const DynamicsSchema = "bodyCount,index,generation,world,position,rotation,velocity,angularVelocity,awake,enabled,type,normal,rollingImpulse,pointCount,anchorA,anchorB,separation,normalImpulse,tangentImpulse,totalNormalImpulse,normalVelocity,baseSeparation,id,persisted,contactCount,fellAsleep,moveCount,beginCount,endCount,point,approachSpeed,hitCount,sensorBeginCount,sensorEndCount,valid,canonicalState"

var dynamicsFields = strings.Split(DynamicsSchema, ",")

func EncodeDynamics(rows [][36]uint64) []byte {
	data := make([]byte, 0, len(rows)*36*8)
	for _, row := range rows {
		for _, v := range row {
			data = binary.LittleEndian.AppendUint64(data, v)
		}
	}

	return data
}

func DecodeDynamics(tb testing.TB, data []byte) [][36]uint64 {
	tb.Helper()

	if len(data)%(36*8) != 0 {
		tb.Fatal("incomplete dynamics row")
	}

	rows := make([][36]uint64, len(data)/(36*8))
	for i := range rows {
		for f := range rows[i] {
			rows[i][f] = binary.LittleEndian.Uint64(data[(i*36+f)*8:])
		}
	}

	return rows
}

func CompareDynamics(tb testing.TB, label string, step int, got [36]uint64, want [][36]uint64) {
	tb.Helper()

	if step > len(want) {
		tb.Fatalf("%s: unexpected observation %d", label, step)
	}

	for i, v := range got {
		if v != want[step-1][i] {
			tb.Fatalf("%s: first divergent observation %d field %s: C=%016x Go=%016x", label, step, dynamicsFields[i], want[step-1][i], v)
		}
	}
}

func EndDynamics(r *oracle.Reader, rows [][36]uint64) {
	r.Bytes("dynamics/schema", []byte(DynamicsSchema))
	r.Bytes("dynamics/rows-36u64-le", EncodeDynamics(rows))
}

func dynamicsShort(r *dynamicsOracle, w *World, prefix string, direct bool) {
	row := ObserveDynamics(r.Reader, w, prefix, direct)
	r.observation++
	CompareDynamics(r.t, prefix, r.observation, row, r.expected)
	r.U64(prefix+"canonicalState", row[35])

	data := make([]byte, 0, 36*8)
	for _, v := range row {
		data = binary.LittleEndian.AppendUint64(data, v)
	}

	r.Bytes(prefix+"digests", data)
}
