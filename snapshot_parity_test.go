// SPDX-License-Identifier: MIT

package box2d

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

type snapshotFamilyScene struct {
	world    *World
	origin   Pos
	shapes   [5]Shape
	joints   [5]Joint
	edge     Shape
	disabled Body
	sleeping Body
	sensor   Shape
	visitor  Body
	circle   Circle
}

func snapshotVec2(r *oracle.Reader, key string, v Vec2) { r.Vec2(key, v.X, v.Y) }
func snapshotPos(r *oracle.Reader, key string, v Pos)   { r.Pos(key, v.X, v.Y) }

// Native images never cross this boundary: only typed public observations do.
func snapshotEvents(r *oracle.Reader, w *World, key string) {
	r.U64(key+"/state", w.StateHash())
	snapshotSensorEvents(r, w, key)
	snapshotContactEvents(r, w, key)
	snapshotJointEvents(r, w, key)
}

func snapshotQuery(t *testing.T, r *oracle.Reader, w *World, origin Pos, key string) {
	t.Helper()

	var (
		mask        uint64
		generations [64]uint32
	)

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

func TestSnapshotRejectsDuplicateFreeBodyIDs(t *testing.T) {
	worldDef := DefaultWorldDef()
	worldDef.WorkerCount = 0
	world := NewWorld(&worldDef)
	BTWorldCleanup(t, world)

	bodyDef := DefaultBodyDef()
	first := world.CreateBody(&bodyDef)
	second := world.CreateBody(&bodyDef)

	first.Destroy()
	second.Destroy()

	image, err := world.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	config := recBuffer{countOnly: true}
	serWorldConfig(&config, world)

	poolOffset := 16 + config.size
	if count := binary.LittleEndian.Uint32(image[poolOffset+4:]); count != 2 {
		t.Fatalf("free body count = %d, want 2", count)
	}

	copy(image[poolOffset+12:poolOffset+16], image[poolOffset+8:poolOffset+12])

	loaded, err := NewWorldFromSnapshot(image, 0)
	BTWorldCleanup(t, loaded)

	if !errors.Is(err, ErrCorrupt) || loaded != nil {
		t.Fatalf("duplicate free body IDs: world = %p, error = %v, want nil and ErrCorrupt", loaded, err)
	}
}

func TestSnapshotRejectsFreeIDOfLiveBody(t *testing.T) {
	def := DefaultWorldDef()
	def.WorkerCount = 0
	world := NewWorld(&def)
	BTWorldCleanup(t, world)

	bodyDef := DefaultBodyDef()
	world.CreateBody(&bodyDef)
	world.CreateBody(&bodyDef).Destroy()
	image := BTSnapshot(world)
	config := recBuffer{countOnly: true}
	serWorldConfig(&config, world)
	binary.LittleEndian.PutUint32(image[16+config.size+8:], 0)

	loaded, err := NewWorldFromSnapshot(image, 0)
	BTWorldCleanup(t, loaded)

	if loaded != nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("free ID of live body: world = %p, error = %v, want nil and ErrCorrupt", loaded, err)
	}
}

func TestSnapshotRejectsLiveProxyOnFreeList(t *testing.T) {
	def := DefaultWorldDef()
	def.WorkerCount = 0
	world := NewWorld(&def)
	BTWorldCleanup(t, world)

	bodyDef := DefaultBodyDef()
	body := world.CreateBody(&bodyDef)
	shapeDef := DefaultShapeDef()
	body.CreateCircleShape(&shapeDef, &Circle{Radius: 1})

	world.broadPhase.trees[StaticBody].proxyFreeList = 0

	loaded, err := NewWorldFromSnapshot(BTSnapshot(world), 0)
	BTWorldCleanup(t, loaded)

	if loaded != nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("live proxy on free list: world = %p, error = %v, want nil and ErrCorrupt", loaded, err)
	}
}

func TestSnapshotRejectsDuplicatePairKeys(t *testing.T) {
	def := DefaultWorldDef()
	def.WorkerCount = 0
	def.Gravity = Vec2{}
	world := NewWorld(&def)
	BTWorldCleanup(t, world)

	bodyDef := DefaultBodyDef()
	ground := world.CreateBody(&bodyDef)
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: 1.75}
	body := world.CreateBody(&bodyDef)
	shapeDef := DefaultShapeDef()
	circle := Circle{Radius: 1}
	ground.CreateCircleShape(&shapeDef, &circle)
	body.CreateCircleShape(&shapeDef, &circle)
	world.Step(1.0/60, 1)

	pairs := &world.broadPhase.pairSet
	if pairs.Count != 1 {
		t.Fatalf("fixture pair count = %d, want 1", pairs.Count)
	}

	for index, item := range pairs.Items {
		if item.Key == 1 {
			pairs.Items[(index+1)&int(pairs.Capacity-1)].Key = 1
			pairs.Count++

			break
		}
	}

	if pairs.Count != 2 {
		t.Fatal("fixture did not contain pair key 1")
	}

	loaded, err := NewWorldFromSnapshot(BTSnapshot(world), 0)
	BTWorldCleanup(t, loaded)

	if loaded != nil || !errors.Is(err, ErrCorrupt) {
		t.Fatalf("duplicate pair keys: world = %p, error = %v, want nil and ErrCorrupt", loaded, err)
	}
}

func TestSnapshotContactHandleStaysInvalidAfterGenerationExhaustion(t *testing.T) {
	def := DefaultWorldDef()
	def.WorkerCount = 0
	def.Gravity = Vec2{}
	world := NewWorld(&def)
	BTWorldCleanup(t, world)

	bodyDef := DefaultBodyDef()
	ground := world.CreateBody(&bodyDef)
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: 1.5}
	moving := world.CreateBody(&bodyDef)
	shapeDef := DefaultShapeDef()
	circle := Circle{Radius: 1}
	ground.CreateCircleShape(&shapeDef, &circle)
	moving.CreateCircleShape(&shapeDef, &circle)
	world.Step(1.0/60, 1)

	if world.Counters().ContactCount != 1 {
		t.Fatal("fixture must contain one contact")
	}

	world.contacts[0].generation = math.MaxUint32
	image := BTSnapshot(world)

	loaded, err := NewWorldFromSnapshot(image, 0)
	if err != nil {
		t.Fatal(err)
	}

	BTWorldCleanup(t, loaded)
	stale := Contact{index1: 1, world: loaded, generation: 1}
	moving.world = loaded

	for range 2 {
		moving.SetTransform(Pos{X: 10}, Rot{C: 1})
		loaded.Step(1.0/60, 1)
		moving.SetTransform(Pos{X: 1.5}, Rot{C: 1})
		loaded.Step(1.0/60, 1)
	}

	if stale.IsValid() {
		t.Error("destroyed contact handle became valid after slot reuse")
	}

	if count := loaded.Counters().ContactCount; count != 1 {
		t.Errorf("contact count = %d, want 1", count)
	}

	for data := range moving.ContactData() {
		if !data.Contact.IsValid() {
			t.Error("replacement contact is invalid")
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
			scene := snapshotFamilyScene{origin: origin}
			snapshotBuildFamilyScene(t, &scene)
			w := scene.world
			shapes, joints := &scene.shapes, &scene.joints
			edge, sensor := scene.edge, scene.sensor
			disabled, sleeping, visitor := scene.disabled, scene.sleeping, scene.visitor
			circle := scene.circle

			image := BTSnapshot(w)
			initial := w.StateHash()

			var tail [90]uint64

			r.U64("capture/state", initial)
			r.Bool("capture/disabled", disabled.Enabled())
			r.Bool("capture/sleeping", sleeping.Awake())

			for i := range 90 {
				scene.step(i)
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
			internal := w

			queuedSensorEnds := len(internal.sensorEndEvents[0]) + len(internal.sensorEndEvents[1])
			if queuedSensorEnds < 1 {
				t.Fatal("sensor corruption history did not queue a real end event")
			}

			r.I32("mutation/queuedSensorEnds", int32(queuedSensorEnds))

			if err := w.UnmarshalBinary(image); err != nil {
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

			r.Bool("restore/disabled", disabled.Enabled())
			r.Bool("restore/sleeping", sleeping.Awake())
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

			capsule := shapes[1].Capsule()
			snapshotVec2(r, "restore/capsule1", capsule.Center1)
			snapshotVec2(r, "restore/capsule2", capsule.Center2)
			r.F32("restore/capsuleRadius", capsule.Radius)

			segment := edge.Segment()
			snapshotVec2(r, "restore/segment1", segment.Point1)
			snapshotVec2(r, "restore/segment2", segment.Point2)
			r.I32("restore/sensorOverlaps", int32(sensor.SensorCapacity()))

			for i := range 90 {
				scene.step(i)

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
	w := NewWorld(&wd)
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
	jd.Base.BodyA = ground
	jd.Base.BodyB = body
	joint := w.CreateFilterJoint(&jd).Joint
	mat := DefaultSurfaceMaterial()
	cd := DefaultChainDef()
	cd.Points = []Vec2{{-2, 0}, {0, 0}, {2, 0}}
	cd.Ghost1 = Vec2{-4, 0}
	cd.Ghost2 = Vec2{4, 0}
	cd.Materials = []SurfaceMaterial{mat}
	chain := ground.CreateChain(&cd)
	image := BTSnapshot(w)

	joint.Destroy()
	chain.Destroy()
	shape.Destroy(true)
	body.Destroy()

	replacement := w.CreateBody(&bd)
	replacementShape := replacement.CreateCircleShape(&sd, &c)
	jd.Base.BodyB = replacement
	replacementJoint := w.CreateFilterJoint(&jd).Joint
	replacementChain := ground.CreateChain(&cd)
	extra := w.CreateBody(&bd)
	extraShape := extra.CreateCircleShape(&sd, &c)
	jd.Base.BodyB = extra
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

	if err := w.UnmarshalBinary(image); err != nil {
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
		clone, err := NewWorldFromSnapshot(image, workers)
		if err != nil {
			t.Fatalf("load workers %d: %v", workers, err)
		}

		BTWorldCleanup(t, clone)

		key := fmt.Sprintf("new/%d", workers)
		r.Bool(key+"/identity", clone != w)
		r.I32(key+"/workers", int32(clone.workerCount))

		mappedBody := body
		mappedBody.world = clone
		mappedShape := shape
		mappedShape.world = clone
		mappedJoint := joint
		mappedJoint.world = clone
		mappedChain := chain
		mappedChain.world = clone

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
		clone.Close()
	}
}

func TestSnapshotHeaderRejectionAndLateCleanup(t *testing.T) {
	r := oracle.Open(t, "snapshot/Headers")
	wd := DefaultWorldDef()
	w := NewWorld(&wd)
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
	jd.Base.BodyA = anchor
	jd.Base.BodyB = body
	joint := w.CreateFilterJoint(&jd).Joint
	material := DefaultSurfaceMaterial()
	material.Friction = 0.375
	cd := DefaultChainDef()
	cd.Points = []Vec2{{-2, 0}, {0, 0}, {2, 0}}
	cd.Ghost1 = Vec2{-4, 0}
	cd.Ghost2 = Vec2{4, 0}
	cd.Materials = []SurfaceMaterial{material}
	chain := anchor.CreateChain(&cd)
	image := BTSnapshot(w)
	hash := w.StateHash()

	cases := []struct {
		name     string
		offset   int
		category error
	}{{"nil", -2, nil}, {"short", -1, nil}, {"magic", 0, ErrBadMagic}, {"version", 4, ErrVersion}, {"layout", 8, ErrLayout}, {"precision", 12, ErrLayout}}
	for _, tc := range cases {
		bad := snapshotCorruptHeader(image, tc.offset)

		err := w.UnmarshalBinary(bad)

		snapshotRequireError(t, err, tc.category, tc.name, "restore")

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

		invalid, err := NewWorldFromSnapshot(bad, 0)
		r.Bool(tc.name+"/newRejected", !invalid.IsValid())

		snapshotRequireError(t, err, tc.category, tc.name, "new-world")
	}
	// Each engine truncates its own native image at the final payload boundary.
	err := w.UnmarshalBinary(image[:len(image)-1])

	var snapshotErr *SnapshotError

	r.Bool("late/rejected", err != nil)

	if err == nil || !errors.As(err, &snapshotErr) {
		t.Fatalf("late restore: %v", err)
	}

	w.Close()

	invalid, err := NewWorldFromSnapshot(image[:len(image)-1], 0)
	r.Bool("late/newRejected", !invalid.IsValid())

	if err == nil || !errors.As(err, &snapshotErr) {
		t.Fatalf("late load: %v", err)
	}

	wd = DefaultWorldDef()
	fresh := NewWorld(&wd)
	BTWorldCleanup(t, fresh)
	body = fresh.CreateBody(&bd)
	r.Bool("cleanup/newBody", body.IsValid())
}

func snapshotContactMaterials(r *oracle.Reader, id *World, key string) {
	w := id
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

	var (
		oldTag, liveTag                                                                                                     int
		tasks, finishes, oldFilter, liveFilter, oldPre, livePre, oldFriction, liveFriction, oldRestitution, liveRestitution int
	)

	wd := DefaultWorldDef()
	wd.WorkerCount = 2
	wd.UserData = &oldTag
	wd.EnqueueTask = func(task func()) func() { tasks++; task(); return func() { finishes++ } }
	wd.FrictionCallback = func(float32, uint64, float32, uint64) float32 { oldFriction++; return 0.9 }
	wd.RestitutionCallback = func(float32, uint64, float32, uint64) float32 { oldRestitution++; return 0.9 }
	w := NewWorld(&wd)
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
	jd.Base.BodyA = ground
	jd.Base.BodyB = body
	jd.Base.CollideConnected = true
	jd.Base.UserData = &oldTag
	joint := w.CreateFilterJoint(&jd).Joint
	image := BTSnapshot(w)
	w.SetUserData(&liveTag)
	body.SetUserData(&liveTag)
	shape.SetUserData(&liveTag)
	joint.SetUserData(&liveTag)
	w.SetCustomFilterCallback(func(Shape, Shape) bool { liveFilter++; return true })
	w.SetPreSolveCallback(func(_ Shape, _ Shape, m *Manifold) { livePre++; m.PointCount = 0 }, nil)
	w.SetFrictionCallback(func(float32, uint64, float32, uint64) float32 { liveFriction++; return 0.125 })
	w.SetRestitutionCallback(func(float32, uint64, float32, uint64) float32 { liveRestitution++; return 0.375 })

	if err := w.UnmarshalBinary(image); err != nil {
		t.Fatalf("restore wired destination: %v", err)
	}

	r.Bool("restore/worldData", w.UserData() == &liveTag)
	r.Bool("restore/bodyData", body.UserData() == nil)
	r.Bool("restore/shapeData", shape.UserData() == nil)
	r.Bool("restore/jointData", joint.UserData() == nil)
	r.I32("restore/workers", int32(w.workerCount))
	w.Step(float32(1.0)/60, 4)

	snapshotRequireTasksComplete(t, tasks, finishes)

	if liveFilter == 0 || livePre == 0 || liveFriction == 0 || liveRestitution == 0 {
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

	clone, err := NewWorldFromSnapshot(image, 0)
	if err != nil {
		t.Fatalf("load default wiring: %v", err)
	}

	BTWorldCleanup(t, clone)

	cb := body
	cb.world = clone
	cs := shape
	cs.world = clone
	cj := joint
	cj.world = clone
	r.Bool("new/worldData", clone.UserData() == nil)
	r.Bool("new/bodyData", cb.UserData() == nil)
	r.Bool("new/shapeData", cs.UserData() == nil)
	r.Bool("new/jointData", cj.UserData() == nil)
	r.I32("new/workers", int32(clone.workerCount))
	clone.Step(float32(1.0)/60, 4)
	r.Bool("new/noLiveCalls", liveFilter == beforeFilter && livePre == beforePre && liveFriction == beforeFriction && liveRestitution == beforeRestitution && tasks == beforeTasks)
	snapshotPos(r, "new/position", cb.Position())
	r.U64("new/state", clone.StateHash())
	snapshotContactMaterials(r, clone, "new")
}

func snapshotSensorEvents(r *oracle.Reader, w *World, key string) {
	var sb []SensorBeginTouchEvent
	for e := range w.SensorBeginEvents() {
		sb = append(sb, e)
	}

	r.I32(key+"/sensorBegin", int32(len(sb)))

	for i, e := range sb {
		p := fmt.Sprintf("%s/sensorBegin/%d", key, i)
		r.I32(p+"/sensor", e.SensorShape.index1)
		r.U32(p+"/sensorGeneration", uint32(e.SensorShape.generation))
		r.I32(p+"/visitor", e.VisitorShape.index1)
		r.U32(p+"/visitorGeneration", uint32(e.VisitorShape.generation))
	}

	var se []SensorEndTouchEvent
	for e := range w.SensorEndEvents() {
		se = append(se, e)
	}

	r.I32(key+"/sensorEnd", int32(len(se)))

	for i, e := range se {
		p := fmt.Sprintf("%s/sensorEnd/%d", key, i)
		r.I32(p+"/sensor", e.SensorShape.index1)
		r.U32(p+"/sensorGeneration", uint32(e.SensorShape.generation))
		r.I32(p+"/visitor", e.VisitorShape.index1)
		r.U32(p+"/visitorGeneration", uint32(e.VisitorShape.generation))
	}
}

func snapshotContactEvents(r *oracle.Reader, w *World, key string) {
	var cb []ContactBeginTouchEvent
	for e := range w.ContactBeginEvents() {
		cb = append(cb, e)
	}

	r.I32(key+"/contactBegin", int32(len(cb)))

	for i, e := range cb {
		p := fmt.Sprintf("%s/contactBegin/%d", key, i)
		r.I32(p+"/a", e.ShapeA.index1)
		r.U32(p+"/aGeneration", uint32(e.ShapeA.generation))
		r.I32(p+"/b", e.ShapeB.index1)
		r.U32(p+"/bGeneration", uint32(e.ShapeB.generation))
		r.I32(p+"/contact", e.Contact.index1)
		r.U32(p+"/contactGeneration", e.Contact.generation)
	}

	var ce []ContactEndTouchEvent
	for e := range w.ContactEndEvents() {
		ce = append(ce, e)
	}

	r.I32(key+"/contactEnd", int32(len(ce)))

	for i, e := range ce {
		p := fmt.Sprintf("%s/contactEnd/%d", key, i)
		r.I32(p+"/a", e.ShapeA.index1)
		r.U32(p+"/aGeneration", uint32(e.ShapeA.generation))
		r.I32(p+"/b", e.ShapeB.index1)
		r.U32(p+"/bGeneration", uint32(e.ShapeB.generation))
		r.I32(p+"/contact", e.Contact.index1)
		r.U32(p+"/contactGeneration", e.Contact.generation)
	}

	var ch []ContactHitEvent
	for e := range w.ContactHitEvents() {
		ch = append(ch, e)
	}

	r.I32(key+"/contactHit", int32(len(ch)))

	for i, e := range ch {
		p := fmt.Sprintf("%s/contactHit/%d", key, i)
		r.I32(p+"/a", e.ShapeA.index1)
		r.I32(p+"/b", e.ShapeB.index1)
		r.U32(p+"/aGeneration", uint32(e.ShapeA.generation))
		r.U32(p+"/bGeneration", uint32(e.ShapeB.generation))
		r.I32(p+"/contact", e.Contact.index1)
		r.U32(p+"/contactGeneration", e.Contact.generation)
		snapshotPos(r, p+"/point", e.Point)
		snapshotVec2(r, p+"/normal", e.Normal)
		r.F32(p+"/speed", e.ApproachSpeed)
	}
}

func snapshotJointEvents(r *oracle.Reader, w *World, key string) {
	var je []JointEvent
	for e := range w.JointEvents() {
		je = append(je, e)
	}

	r.I32(key+"/jointEvents", int32(len(je)))

	for i, e := range je {
		p := fmt.Sprintf("%s/jointEvent/%d", key, i)
		r.I32(p+"/id", e.Joint.index1)
		r.U32(p+"/generation", uint32(e.Joint.generation))
		r.Bool(p+"/userDataNil", e.UserData == nil)
	}
}

func snapshotCorruptHeader(image []byte, offset int) []byte {
	bad := append([]byte(nil), image...)

	switch offset {
	case -2:
		return nil
	case -1:
		return bad[:15]
	default:
		value := binary.LittleEndian.Uint32(bad[offset:])
		bit := uint32(1)

		if offset == 12 {
			bit = 2
		}

		binary.LittleEndian.PutUint32(bad[offset:], value^bit)

		return bad
	}
}

func snapshotRequireError(t *testing.T, err, category error, name, operation string) {
	t.Helper()

	var snapshotErr *SnapshotError
	if err == nil || !errors.As(err, &snapshotErr) {
		t.Fatalf("%s %s error category: %v", name, operation, err)
	}

	if category != nil && !errors.Is(err, category) {
		t.Fatalf("%s %s: got %v want errors.Is(%v)", name, operation, err, category)
	}
}

func snapshotBuildFamilyScene(t *testing.T, scene *snapshotFamilyScene) {
	t.Helper()

	origin := scene.origin
	shapes, joints := &scene.shapes, &scene.joints
	wd := DefaultWorldDef()
	wd.Gravity = Vec2{0, -2}
	w := NewWorld(&wd)
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
		base.BodyA = ground
		base.BodyB = bodies[i]
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

	scene.world = w
	scene.edge = edge
	scene.disabled = disabled
	scene.sleeping = sleeping
	scene.sensor = sensor
	scene.visitor = visitor
	scene.circle = circle

	for range 20 {
		w.Step(float32(1.0)/60, 4)
	}
}

func (scene *snapshotFamilyScene) step(i int) {
	if i == 5 {
		scene.disabled.Enable()
	}

	if i == 9 {
		scene.sleeping.SetAwake(true)
	}

	scene.world.Step(float32(1.0)/60, 4)
}

func snapshotRequireTasksComplete(t *testing.T, tasks, finishes int) {
	t.Helper()

	if tasks == 0 || tasks != finishes {
		t.Fatalf("restored tasks incomplete: tasks %d finished %d", tasks, finishes)
	}
}
