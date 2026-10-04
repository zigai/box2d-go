package box2d

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zigai/box2d/internal/oracle"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

type dynamicsOracle struct {
	*oracle.Reader
	t           *testing.T
	expected    [][36]uint64
	observation int
}

func dynamicsOpen(t *testing.T, name string) *dynamicsOracle {
	r := &dynamicsOracle{Reader: oracle.Open(t, name), t: t}
	precision := "float"
	if oracle.DoublePrecision {
		precision = "double"
	}
	data, err := os.ReadFile("testdata/oracle/" + precision + "/" + name + ".jsonl")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var record restitutionRecord
		err := decoder.Decode(&record)
		if err == io.EOF {
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
	r.Bytes("dynamics/schema", []byte(BTDynamicsSchema))
	t.Cleanup(func() {
		if !t.Failed() && r.observation != len(r.expected) {
			t.Errorf("dynamics observations C=%d Go=%d", len(r.expected), r.observation)
		}
	})
	return r
}

func dynamicsWorld(t *testing.T, def WorldDef) World {
	w := CreateWorld(&def)
	BTWorldCleanup(t, w)
	return w
}
func dynamicsBox(w World, typ BodyType, x, y float32, hx, hy float32, material SurfaceMaterial) (Body, Shape) {
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
func dynamicsSteps(r *dynamicsOracle, w World, n int, prefix string) {
	for i := range n {
		w.Step(1.0/60.0, 4)
		btDynamicsShort(r, w, fmt.Sprintf("%sstep/%d/", prefix, i+1), true)
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
				speed := float32(10)
				if partition == 9 {
					speed = math.Nextafter32(speed, float32(math.Inf(-1)))
				}
				if partition == 11 {
					speed = math.Nextafter32(speed, float32(math.Inf(1)))
				}
				b.SetGravityScale(0)
				b.SetLinearVelocity(Vec2{speed, 0})
			}
			btDynamicsShort(r, w, "initial/", true)
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
			switch partition {
			case 0:
				w.EnableWarmStarting(false)
			case 1:
				w.SetContactTuning(5, .2, .5)
			case 2:
				w.SetContactTuning(120, 3, 10)
			case 3:
				shape.SetFriction(0)
			case 4:
				m.RollingResistance = .7
				shape.SetSurfaceMaterial(&m)
			case 5:
				m.TangentSpeed = 3
				shape.SetSurfaceMaterial(&m)
			case 6:
				w.SetFrictionCallback(dynamicsFriction)
			case 7:
				m.RollingResistance = .4
				shape.SetSurfaceMaterial(&m)
			case 8:
				b.EnableContactRecycling(false)
			case 9:
				w.SetContactTuning(30, 1, 0)
			case 10:
				w.SetMaximumLinearSpeed(2)
			case 11:
				w.SetGravity(Vec2{2, -5})
			}
			dynamicsSteps(r, w, 30, "")
			ground.SetFriction(.1)
			m = shape.SurfaceMaterial()
			m.Friction = .9
			m.RollingResistance = .2
			m.TangentSpeed = -2
			shape.SetSurfaceMaterial(&m)
			w.EnableWarmStarting(true)
			w.SetContactTuning(15, .5, 2)
			w.SetFrictionCallback(nil)
			dynamicsSteps(r, w, 30, "mutated/")
		})
	}
}

func TestDynamicsRestitutionBoundary(t *testing.T) {
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
				speed := wd.RestitutionThreshold
				if neighbor == 0 {
					speed = math.Nextafter32(speed, float32(math.Inf(-1)))
				}
				if neighbor == 2 {
					speed = math.Nextafter32(speed, float32(math.Inf(1)))
				}
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
						b := w.CreateBody(&bd)
						sd := DefaultShapeDef()
						sd.Material = m
						sd.EnableHitEvents = true
						sd.EnableContactEvents = true
						box := MakeBox(.15, .4)
						b.CreatePolygonShape(&sd, &box)
						dynamicsSteps(r, w, 30, "")
						w.EnableContinuous(true)
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
				asleep := false
				for i := range 900 {
					w.Step(1.0/60.0, 4)
					btDynamicsShort(r, w, fmt.Sprintf("settle/%d/", i+1), i == 0 || (i+1)%60 == 0)
					if !lower.IsAwake() && !upper.IsAwake() {
						asleep = true
						r.I32("sleepStep", int32(i+1))
						break
					}
				}
				if !asleep {
					t.Fatal("loaded island did not sleep within source locator bound 900")
				}
				switch operation {
				case 0:
					lower.ApplyForceToCenter(Vec2{20, 5}, wake != 0)
				case 1:
					lower.ApplyForce(Vec2{20, 5}, Pos{1, 1}, wake != 0)
				case 2:
					lower.ApplyTorque(10, wake != 0)
				case 3:
					lower.ApplyLinearImpulseToCenter(Vec2{2, 3}, wake != 0)
				case 4:
					lower.ApplyLinearImpulse(Vec2{2, 3}, Pos{1, 1}, wake != 0)
				case 5:
					lower.ApplyAngularImpulse(2, wake != 0)
				case 6:
					lower.SetAwake(true)
				case 7:
					lower.EnableSleep(false)
				case 8:
					w.EnableSleeping(false)
				case 9:
					lower.Disable()
				case 10:
					lower.SetType(KinematicBody)
					lower.SetLinearVelocity(Vec2{1, 0})
				case 11:
					lower.SetTransform(Pos{1, .5}, Rot{C: 1})
				}
				btDynamicsShort(r, w, "operation/", true)
				dynamicsSteps(r, w, 12, "after/")
				lower.Enable()
				lower.SetType(DynamicBody)
				lower.SetAwake(true)
				dynamicsSteps(r, w, 12, "restored/")
			})
		}
	}
}
