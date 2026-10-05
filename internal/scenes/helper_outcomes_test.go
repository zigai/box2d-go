//go:build box2d_stress

// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func helperKey(family string, op, item int, field string) string {
	return fmt.Sprintf("%s.%d.%d.%s", family, op, item, field)
}

func helperCounts(r *oracle.Reader, w *box2d.World, family string, op int) {
	c := w.Counters()
	r.I32(helperKey(family, op, 0, "bodies"), int32(c.BodyCount))
	r.I32(helperKey(family, op, 0, "shapes"), int32(c.ShapeCount))
	r.I32(helperKey(family, op, 0, "joints"), int32(c.JointCount))
}

func observeHumanHelper(r *oracle.Reader, h *Human, op int) {
	r.Bool(helperKey("human", op, 0, "spawned"), h.IsSpawned)
	r.F32(helperKey("human", op, 0, "scale"), h.Scale)

	for i, bone := range h.Bones {
		b, j := bone.BodyID, bone.JointID
		key := func(field string) string { return helperKey("human", op, i, field) }
		r.Bool(key("bodyValid"), b.IsValid())
		r.Bool(key("jointValid"), j.IsValid())

		if b != (box2d.Body{}) {
			v := b.LinearVelocity()
			r.Vec2(key("velocity"), v.X, v.Y)
			r.F32(key("angular"), b.AngularVelocity())
			m := b.MassData()
			r.F32(key("mass"), m.Mass)
			r.Vec2(key("center"), m.Center.X, m.Center.Y)
			r.F32(key("inertia"), m.RotationalInertia)

			var shapes []box2d.Shape
			for id := range b.Shapes() {
				shapes = append(shapes, id)
			}

			r.I32(key("shapes"), int32(len(shapes)))

			for _, s := range shapes {
				r.I32(key("shapeType"), int32(s.Type()))
				r.Bool(key("sensor"), s.SensorEventsEnabled())

				switch s.Type() {
				case box2d.CapsuleShape:
					c := s.Capsule()
					r.Vec2(key("c1"), c.Center1.X, c.Center1.Y)
					r.Vec2(key("c2"), c.Center2.X, c.Center2.Y)
					r.F32(key("radius"), c.Radius)
				case box2d.PolygonShape:
					p := s.Polygon()
					r.I32(key("vertices"), int32(p.Count))
					r.F32(key("radius"), p.Radius)
					r.Vec2(key("centroid"), p.Centroid.X, p.Centroid.Y)

					for v := range p.Count {
						r.Vec2(key("vertex"), p.Vertices[v].X, p.Vertices[v].Y)
						r.Vec2(key("normal"), p.Normals[v].X, p.Normals[v].Y)
					}
				case box2d.CircleShape, box2d.SegmentShape, box2d.ChainSegmentShape, box2d.ShapeTypeCount:
					// Human fixtures contain only capsules and polygons.
				}
			}
		}

		if j != (box2d.Joint{}) {
			revolute := box2d.RevoluteJoint{Joint: j}
			r.Bool(key("motor"), revolute.MotorEnabled())
			r.F32(key("torque"), revolute.MaxMotorTorque())
			r.Bool(key("spring"), revolute.SpringEnabled())
			r.F32(key("hertz"), revolute.SpringHertz())
			r.F32(key("damping"), revolute.SpringDampingRatio())

			a, b := j.LocalFrameA(), j.LocalFrameB()
			r.Vec2(key("anchorA"), a.P.X, a.P.Y)
			r.Rot(key("rotationA"), a.Q.C, a.Q.S)
			r.Vec2(key("anchorB"), b.P.X, b.P.Y)
			r.Rot(key("rotationB"), b.Q.C, b.Q.S)
		}
	}
}

func rainHelperOutcomes(r *oracle.Reader, def *box2d.WorldDef) {
	w := box2d.NewWorld(def)
	defer w.Close()

	CreateRain(w)

	var (
		first      box2d.Body
		firstJoint box2d.Joint
	)

	for step := range 1000 {
		r.F32(helperKey("rain", step, 0, "result"), StepRain(w, step))

		if step == 0 {
			first = gRainData.groups[0].humans[0].Bones[BoneTorso].BodyID
			firstJoint = gRainData.groups[0].humans[0].Bones[BoneTorso].JointID
		}

		r.I32(helperKey("rain", step, 0, "columns"), int32(gRainData.columnCount))
		r.I32(helperKey("rain", step, 0, "head"), int32(gRainData.columnIndex))
		r.Bool(helperKey("rain", step, 0, "firstBody"), first.IsValid())
		r.Bool(helperKey("rain", step, 0, "firstJoint"), firstJoint.IsValid())
		helperCounts(r, w, "rain", step)

		if step == 320 && (first.IsValid() || firstJoint.IsValid()) {
			panic("normal Rain recycle retained stale handles")
		}

		w.Step(float32(1)/60, 4)
	}
}

func sleepHelperOutcomes(r *oracle.Reader, def *box2d.WorldDef) {
	w := box2d.NewWorld(def)
	defer w.Close()

	CreateSleep(w)
	w.Step(float32(1)/60, 4)

	for i := range sleepPyramidCount {
		gSleepData.bodyIDA[i].SetAwake(false)
	}

	for op := range 3 {
		if op > 0 {
			r.F32(helperKey("sleep", op, 0, "result"), StepSleep(w, op))
		}

		for i := range sleepPyramidCount {
			r.Bool(helperKey("sleep", op, i, "a"), gSleepData.bodyIDA[i].Awake())
			r.Bool(helperKey("sleep", op, i, "b"), gSleepData.bodyIDB[i].Awake())
		}

		helperCounts(r, w, "sleep", op)
	}
}

func lifecycleHelperOutcomes(t *testing.T, r *oracle.Reader, def *box2d.WorldDef) {
	t.Helper()

	create := func() *box2d.World {
		w := box2d.NewWorld(def)

		t.Cleanup(func() {
			if w.IsValid() {
				w.Close()
			}
		})

		return w
	}

	t.Cleanup(DestroyTreeCast)
	t.Cleanup(DestroyTileWorld)

	for cycle := range 2 {
		w := create()
		CreateTreeCast(w)
		r.Bool(helperKey("tree", cycle, 0, "created"), gTreeCast.created)
		r.I32(helperKey("tree", cycle, 0, "proxies"), int32(gTreeCast.proxyCount))
		r.F32(helperKey("tree", cycle, 0, "result"), StepTreeCast(w, 0))

		s := GetTreeCastBenchmarkStats()
		r.I32(helperKey("tree", cycle, 0, "nodes"), int32(s.NodeVisits))
		r.I32(helperKey("tree", cycle, 0, "leaves"), int32(s.LeafVisits))
		DestroyTreeCast()
		DestroyTreeCast()
		r.Bool(helperKey("tree", cycle, 0, "destroyed"), !gTreeCast.created)
		r.I32(helperKey("tree", cycle, 0, "cleared"), int32(gTreeCast.proxyCount))

		s = GetTreeCastBenchmarkStats()
		r.I32(helperKey("tree", cycle, 0, "retainedNodes"), int32(s.NodeVisits))
		r.I32(helperKey("tree", cycle, 0, "retainedLeaves"), int32(s.LeafVisits))
		w.Close()
		w = create()
		CreateTileWorld(w)

		gTileWorld.groundHead = gTileWorld.groundCount - 1
		old := gTileWorld.groundIDs[gTileWorld.groundHead]

		r.F32(helperKey("tile", cycle, 0, "result"), StepTileWorld(w, 0))
		r.I32(helperKey("tile", cycle, 0, "head"), int32(gTileWorld.groundHead))
		r.Bool(helperKey("tile", cycle, 0, "old"), old.IsValid())
		r.F32(helperKey("tile", cycle, 0, "left"), gTileWorld.xLeft)
		r.F32(helperKey("tile", cycle, 0, "next"), gTileWorld.xNext)
		helperCounts(r, w, "tile", cycle)

		s = GetTileWorldBenchmarkStats()
		r.I32(helperKey("tile", cycle, 0, "nodes"), int32(s.NodeVisits))
		r.I32(helperKey("tile", cycle, 0, "leaves"), int32(s.LeafVisits))
		DestroyTileWorld()
		DestroyTileWorld()
		r.I32(helperKey("tile", cycle, 0, "cleared"), int32(gTileWorld.groundCount))

		s = GetTileWorldBenchmarkStats()
		r.I32(helperKey("tile", cycle, 0, "retainedNodes"), int32(s.NodeVisits))
		r.I32(helperKey("tile", cycle, 0, "retainedLeaves"), int32(s.LeafVisits))
		w.Close()
	}
}

func humanHelperOutcomes(r *oracle.Reader, def *box2d.WorldDef) {
	w := box2d.NewWorld(def)
	defer w.Close()

	var h Human
	CreateHuman(&h, w, box2d.Pos{Y: 5}, 1, .05, 5, .5, 1, nil, true)
	observeHumanHelper(r, &h, 0)
	h.SetVelocity(box2d.Vec2{X: 1, Y: -2})
	observeHumanHelper(r, &h, 1)
	h.ApplyRandomAngularImpulse(.5)
	observeHumanHelper(r, &h, 2)
	h.SetJointFrictionTorque(0)
	observeHumanHelper(r, &h, 3)
	h.SetJointFrictionTorque(.1)
	observeHumanHelper(r, &h, 4)
	h.SetJointSpringHertz(0)
	observeHumanHelper(r, &h, 5)
	h.SetJointSpringHertz(2)
	observeHumanHelper(r, &h, 6)
	h.SetJointDampingRatio(.75)
	observeHumanHelper(r, &h, 7)
	h.EnableSensorEvents(true)
	observeHumanHelper(r, &h, 8)
	h.SetScale(1.5)
	observeHumanHelper(r, &h, 9)
	h.EnableSensorEvents(false)
	observeHumanHelper(r, &h, 10)
	old := h
	DestroyHuman(&h)
	observeHumanHelper(r, &h, 11)

	for i, bone := range old.Bones {
		r.Bool(helperKey("human", 11, i, "oldBody"), bone.BodyID.IsValid())
		r.Bool(helperKey("human", 11, i, "oldJoint"), bone.JointID.IsValid())
	}

	helperCounts(r, w, "human", 11)

	var sparse Human

	bd := box2d.DefaultBodyDef()
	bd.Type = box2d.DynamicBody
	sparse.Bones[BoneHip].BodyID = w.CreateBody(&bd)
	sparse.Bones[BoneHead].BodyID = w.CreateBody(&bd)
	sparse.SetVelocity(box2d.Vec2{X: 2, Y: -3})
	v := sparse.Bones[BoneHip].BodyID.LinearVelocity()
	r.Vec2("human.sparse.hip", v.X, v.Y)
	v = sparse.Bones[BoneHead].BodyID.LinearVelocity()
	r.Vec2("human.sparse.head", v.X, v.Y)
	r.Bool("human.sparse.torsoNull", sparse.Bones[BoneTorso].BodyID == (box2d.Body{}))
}

func TestSceneHelperImmediateOutcomesAndNormalRecycle(t *testing.T) {
	seed := gRandomSeed

	t.Cleanup(func() { gRandomSeed = seed })

	for _, workers := range []int{1, 4} {
		t.Run(fmt.Sprintf("workers=%d", workers), func(t *testing.T) {
			gRandomSeed = 12345
			def := box2d.DefaultWorldDef()
			def.WorkerCount = workers
			r := oracle.Open(t, fmt.Sprintf("scenes/ExtendedHelpers_workers%d", workers))
			rainHelperOutcomes(r, &def)
			sleepHelperOutcomes(r, &def)
			lifecycleHelperOutcomes(t, r, &def)
			humanHelperOutcomes(r, &def)
		})
	}
}
