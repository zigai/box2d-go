// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func traceTransforms(r *oracle.Reader, worldID *World, key string, step int) {
	w := worldID
	full := step == 0 || (step+1)%60 == 0
	count := 0

	for i := range w.bodies {
		if w.bodies[i].id == i {
			count++
		}
	}

	if full {
		r.I32(key+"/count", int32(count))
	}

	hash := fnv.New64a()

	var data [24]byte

	for i := range w.bodies {
		body := &w.bodies[i]
		if body.id != i {
			continue
		}

		xf := getBodySim(w, body).transform
		size := 16

		if IsDoublePrecision() {
			binary.LittleEndian.PutUint64(data[0:8], math.Float64bits(float64(xf.P.X)))
			binary.LittleEndian.PutUint64(data[8:16], math.Float64bits(float64(xf.P.Y)))

			size = 24
		} else {
			binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(float32(xf.P.X)))
			binary.LittleEndian.PutUint32(data[4:8], math.Float32bits(float32(xf.P.Y)))
		}

		binary.LittleEndian.PutUint32(data[size-8:size-4], math.Float32bits(xf.Q.C))
		binary.LittleEndian.PutUint32(data[size-4:size], math.Float32bits(xf.Q.S))
		_, _ = hash.Write(data[:size])

		if full {
			prefix := fmt.Sprintf("%s/body/%d", key, i)
			r.I32(prefix+"/id", int32(i))
			r.Pos(prefix+"/position", xf.P.X, xf.P.Y)
			r.Rot(prefix+"/rotation", xf.Q.C, xf.Q.S)
		}
	}

	r.U64(key, hash.Sum64())
}

func largeWorldBodyRelativeCenter(worldID *World, bodyID Body, base Pos) Vec2 {
	w := worldID
	b := getBodyFullID(w, bodyID)

	return getBodySim(w, b).center.Sub(base)
}

type largeWorldPyramidResult struct {
	sleepStep int
	rel       [9]Vec2
}

func largeWorldRunPyramid(t *testing.T, r *oracle.Reader, base Pos) largeWorldPyramidResult {
	t.Helper()

	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	bd := DefaultBodyDef()
	bd.Position = base
	ground := w.CreateBody(&bd)
	box := MakeBox(10, 0.5)
	sd := DefaultShapeDef()
	ground.CreatePolygonShape(&sd, &box)

	offsets := [9]Vec2{{X: -2, Y: 1}, {X: -1, Y: 1}, {Y: 1}, {X: 1, Y: 1}, {X: 2, Y: 1}, {X: -1, Y: 2}, {Y: 2}, {X: 1, Y: 2}, {Y: 3}}

	var bodies [9]Body

	box = MakeBox(0.5, 0.5)
	sd = DefaultShapeDef()

	for i := range bodies {
		bd := DefaultBodyDef()
		bd.Type = DynamicBody
		bd.Position = Pos{X: base.X + oracle.Real(offsets[i].X), Y: base.Y + oracle.Real(offsets[i].Y)}
		bodies[i] = w.CreateBody(&bd)
		bodies[i].CreatePolygonShape(&sd, &box)
	}

	result := largeWorldPyramidResult{sleepStep: -1}

	for step := range 250 {
		w.Step(float32(1)/60, 4)
		traceTransforms(r, w, fmt.Sprintf("L75/step/%d", step), step)

		if result.sleepStep < 0 {
			moves := 0
			for range w.BodyMoveEvents() {
				moves++
			}

			if moves == 0 && w.AwakeBodyCount() == 0 {
				result.sleepStep = step
			}
		}
	}

	for i := range bodies {
		result.rel[i] = largeWorldBodyRelativeCenter(w, bodies[i], base)
	}

	w.Close()

	return result
}

func largeWorldRunBullet(t *testing.T, r *oracle.Reader, base Pos) float32 {
	t.Helper()

	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	bd := DefaultBodyDef()
	bd.Type = StaticBody
	bd.Position = base
	wall := w.CreateBody(&bd)
	box := MakeBox(0.05, 5)
	sd := DefaultShapeDef()
	wall.CreatePolygonShape(&sd, &box)

	bd = DefaultBodyDef()
	bd.Type = DynamicBody
	bd.IsBullet = true
	bd.GravityScale = 0
	bd.Position = Pos{X: base.X + 10, Y: base.Y}
	bd.LinearVelocity = Vec2{X: -200}
	bullet := w.CreateBody(&bd)
	circle := Circle{Radius: 0.1}
	sd = DefaultShapeDef()
	bullet.CreateCircleShape(&sd, &circle)

	for step := range 30 {
		w.Step(float32(1)/60, 4)
		traceTransforms(r, w, fmt.Sprintf("L150/step/%d", step), step)
	}

	x := largeWorldBodyRelativeCenter(w, bullet, base).X
	w.Close()

	return x
}

func largeWorldRunRayCast(t *testing.T, base Pos) Vec2 {
	t.Helper()

	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	bd := DefaultBodyDef()
	bd.Position = base
	body := w.CreateBody(&bd)
	box := MakeBox(0.5, 0.5)
	sd := DefaultShapeDef()
	body.CreatePolygonShape(&sd, &box)

	origin := Pos{X: base.X - 5, Y: base.Y}
	result := w.CastRayClosest(origin, Vec2{X: 10}, DefaultQueryFilter())
	rel := result.Point.Sub(base)

	w.Close()

	return rel
}

type largeWorldOriginQueryData struct {
	overlapCount  int
	castPoint     Pos
	castFraction  float32
	moverFraction float32
	planeCount    int
	insidePoint   bool
	shapeRayPoint Pos
	shapeRayHit   bool
}

func largeWorldRunOriginQueries(t *testing.T, base Pos) largeWorldOriginQueryData {
	t.Helper()

	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	bd := DefaultBodyDef()
	bd.Position = base
	body := w.CreateBody(&bd)
	box := MakeBox(0.5, 0.5)
	sd := DefaultShapeDef()
	shape := body.CreatePolygonShape(&sd, &box)
	filter := DefaultQueryFilter()
	data := largeWorldOriginQueryData{castFraction: 1}
	center := []Vec2{{}}
	overlap := MakeProxy(center, 0.1)
	w.OverlapShape(base, &overlap, filter, func(Shape) bool { data.overlapCount++; return true })

	start := []Vec2{{X: -5}}
	cast := MakeProxy(start, 0.1)
	w.CastShape(base, &cast, Vec2{X: 10}, filter, func(_ Shape, point Pos, _ Vec2, fraction float32) float32 {
		data.castPoint = point
		data.castFraction = fraction

		return fraction
	})
	mover := Capsule{Center1: Vec2{X: -5, Y: -0.2}, Center2: Vec2{X: -5, Y: 0.2}, Radius: 0.3}
	data.moverFraction = w.CastMover(base, &mover, Vec2{X: 10}, filter)
	touching := Capsule{Center1: Vec2{X: -0.9, Y: -0.2}, Center2: Vec2{X: -0.9, Y: 0.2}, Radius: 0.5}
	w.CollideMover(base, &touching, filter, func(_ Shape, _ *PlaneResult) bool { data.planeCount++; return true })
	data.insidePoint = shape.TestPoint(base)
	ray := shape.RayCast(base.Add(Vec2{X: -5}), Vec2{X: 10})
	data.shapeRayHit = ray.Hit
	data.shapeRayPoint = ray.Point

	w.Close()

	return data
}

func TestLargeWorld(t *testing.T) {
	t.Run("LargeWorldPyramidTest", func(t *testing.T) { checkLargeWorldPyramidTest(t) })
	t.Run("LargeWorldBulletTest", func(t *testing.T) { checkLargeWorldBulletTest(t) })
	t.Run("LargeWorldRayCastTest", func(t *testing.T) { checkLargeWorldRayCastTest(t) })
	t.Run("LargeWorldOriginQueryTest", func(t *testing.T) { checkLargeWorldOriginQueryTest(t) })
	t.Run("LargeWorldRecordingTest", func(t *testing.T) { checkLargeWorldRecordingTest(t) })
}

func largeWorldCheckRelativeCoordinate(t *testing.T, r *oracle.Reader, key string, delta, tolerance float32) {
	t.Helper()

	if delta < -tolerance || tolerance < delta {
		t.Fatalf("condition false: abs(%v) < %v", delta, tolerance)
	}

	r.F32(key, delta)
}

func checkLargeWorldPyramidTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "large_world/LargeWorldPyramidTest")

	origin := largeWorldRunPyramid(t, r, Pos{})
	if origin.sleepStep <= 0 {
		t.Fatal("condition false")
	}

	r.Bool("L102", origin.sleepStep > 0)
	r.I32("L102/value0", int32(origin.sleepStep))

	if IsDoublePrecision() {
		large := largeWorldRunPyramid(t, r, Pos{X: 1e7})
		if large.sleepStep != origin.sleepStep {
			t.Fatal("condition false")
		}

		r.Bool("L107", large.sleepStep == origin.sleepStep)
		r.I32("L107/value0", int32(large.sleepStep))
		r.I32("L107/value1", int32(origin.sleepStep))

		for i := range origin.rel {
			largeWorldCheckRelativeCoordinate(t, r, "L110", large.rel[i].X-origin.rel[i].X, float32(1e-3))

			largeWorldCheckRelativeCoordinate(t, r, "L111", large.rel[i].Y-origin.rel[i].Y, float32(1e-3))
		}
	}
}

func checkLargeWorldBulletTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "large_world/LargeWorldBulletTest")

	x := largeWorldRunBullet(t, r, Pos{})
	if !(x > 0 && x < 0.5) {
		t.Fatal("condition false")
	}

	r.Bool("L165", x > 0 && x < 0.5)
	r.F32("L165/value0", x)

	if IsDoublePrecision() {
		x = largeWorldRunBullet(t, r, Pos{X: 1e7})
		if !(x > 0 && x < 0.5) {
			t.Fatal("condition false")
		}

		r.Bool("L169", x > 0 && x < 0.5)
		r.F32("L169/value0", x)
	}
}

func checkLargeWorldRayCastTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "large_world/LargeWorldRayCastTest")

	rel := largeWorldRunRayCast(t, Pos{})
	largeWorldCheckRelativeCoordinate(t, r, "L207", rel.X+0.5, float32(1e-4))

	largeWorldCheckRelativeCoordinate(t, r, "L208", rel.Y, float32(1e-4))

	if IsDoublePrecision() {
		rel = largeWorldRunRayCast(t, Pos{X: 1e7})
		largeWorldCheckRelativeCoordinate(t, r, "L212", rel.X+0.5, float32(1e-4))

		largeWorldCheckRelativeCoordinate(t, r, "L213", rel.Y, float32(1e-4))
	}
}

func checkLargeWorldOriginQueryTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "large_world/LargeWorldOriginQueryTest")

	origin := largeWorldRunOriginQueries(t, Pos{})
	if origin.overlapCount != 1 {
		t.Fatal("condition false")
	}

	r.Bool("L311", origin.overlapCount == 1)
	r.I32("L311/value0", int32(origin.overlapCount))

	if !(origin.castFraction < 1) {
		t.Fatal("condition false")
	}

	r.Bool("L312", origin.castFraction < 1)
	r.F32("L312/value0", origin.castFraction)

	if !(origin.moverFraction < 1) {
		t.Fatal("condition false")
	}

	r.Bool("L313", origin.moverFraction < 1)
	r.F32("L313/value0", origin.moverFraction)

	if origin.planeCount < 1 {
		t.Fatal("condition false")
	}

	r.Bool("L314", origin.planeCount >= 1)
	r.I32("L314/value0", int32(origin.planeCount))

	if !origin.insidePoint {
		t.Fatal("condition false")
	}

	r.Bool("L315", origin.insidePoint)

	if !origin.shapeRayHit {
		t.Fatal("condition false")
	}

	r.Bool("L316", origin.shapeRayHit)

	castRel := origin.castPoint.Sub(Pos{})
	largeWorldCheckRelativeCoordinate(t, r, "L319", castRel.X+0.5, float32(1e-3))

	rayRel := origin.shapeRayPoint.Sub(Pos{})
	largeWorldCheckRelativeCoordinate(t, r, "L321", rayRel.X+0.5, float32(1e-3))

	if IsDoublePrecision() {
		base := Pos{X: 1e7}

		large := largeWorldRunOriginQueries(t, base)
		if large.overlapCount != origin.overlapCount {
			t.Fatal("condition false")
		}

		r.Bool("L328", large.overlapCount == origin.overlapCount)
		r.I32("L328/value0", int32(large.overlapCount))
		r.I32("L328/value1", int32(origin.overlapCount))

		if large.planeCount != origin.planeCount {
			t.Fatal("condition false")
		}

		r.Bool("L329", large.planeCount == origin.planeCount)
		r.I32("L329/value0", int32(large.planeCount))
		r.I32("L329/value1", int32(origin.planeCount))

		if large.insidePoint != origin.insidePoint {
			t.Fatal("condition false")
		}

		r.Bool("L330", large.insidePoint == origin.insidePoint)

		if large.shapeRayHit != origin.shapeRayHit {
			t.Fatal("condition false")
		}

		r.Bool("L331", large.shapeRayHit == origin.shapeRayHit)

		largeWorldCheckRelativeCoordinate(t, r, "L332", large.castFraction-origin.castFraction, float32(1e-4))

		largeWorldCheckRelativeCoordinate(t, r, "L333", large.moverFraction-origin.moverFraction, float32(1e-4))

		largeCast := large.castPoint.Sub(base)
		largeWorldCheckRelativeCoordinate(t, r, "L336", largeCast.X-castRel.X, float32(1e-3))

		largeWorldCheckRelativeCoordinate(t, r, "L337", largeCast.Y-castRel.Y, float32(1e-3))

		largeRay := large.shapeRayPoint.Sub(base)
		largeWorldCheckRelativeCoordinate(t, r, "L340", largeRay.X-rayRel.X, float32(1e-3))

		largeWorldCheckRelativeCoordinate(t, r, "L341", largeRay.Y-rayRel.Y, float32(1e-3))
	}
}

func checkLargeWorldRecordingTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "large_world/LargeWorldRecordingTest")
	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	base := Pos{}
	if IsDoublePrecision() {
		base = Pos{X: 1e7}
	}

	rec := NewRecording(0)
	w.StartRecording(rec)

	groundDef := DefaultBodyDef()
	groundDef.Position = base
	ground := w.CreateBody(&groundDef)
	groundBox := MakeBox(20, 0.5)
	gsd := DefaultShapeDef()
	ground.CreatePolygonShape(&gsd, &groundBox)

	stackDef := DefaultBodyDef()
	stackDef.Type = DynamicBody
	stackDef.Position = Pos{X: base.X, Y: base.Y + 1}
	stack := w.CreateBody(&stackDef)
	box := MakeBox(0.5, 0.5)
	sd := DefaultShapeDef()
	sd.Density = 1
	stack.CreatePolygonShape(&sd, &box)

	sliderDef := DefaultBodyDef()
	sliderDef.Type = DynamicBody
	sliderDef.GravityScale = 0
	sliderDef.Position = Pos{X: base.X, Y: base.Y + 5}
	sliderDef.LinearVelocity = Vec2{X: 3}
	slider := w.CreateBody(&sliderDef)
	slider.CreatePolygonShape(&sd, &box)

	for step := range 30 {
		w.Step(float32(1)/60, 4)
		traceTransforms(r, w, fmt.Sprintf("L401/step/%d", step), step)
	}

	xf := slider.Transform()
	slider.SetTransform(xf.P, xf.Q)

	for step := range 30 {
		w.Step(float32(1)/60, 4)
		traceTransforms(r, w, fmt.Sprintf("L409/step/%d", step), step+30)
	}

	w.StopRecording()
	w.Close()

	data := rec.Data()

	size := rec.Size()
	if size <= 0 {
		t.Fatal("condition false")
	}

	r.Bool("L417", size > 0)

	if ValidateReplay(data, 0) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L421", true)

	if ValidateReplay(data, 4) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L422", true)
}
