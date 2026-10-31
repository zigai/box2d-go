// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"
	"slices"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const isDebug = false

func millisecondsSince(start time.Time) float32 {
	return float32(time.Since(start).Seconds() * 1000.0)
}

func millisecondsAndReset(start *time.Time) float32 {
	now := time.Now()
	ms := float32(now.Sub(*start).Seconds() * 1000.0)
	*start = now
	return ms
}

const (
	barrelCircleShape = iota
	barrelCapsuleShape
	barrelMixShape
	barrelCompoundShape
	barrelHumanShape
)

const (
	barrelMaxColumns = 26
	barrelMaxRows    = 150
)

type BenchmarkBarrel struct {
	Sample
	bodies      [barrelMaxRows * barrelMaxColumns]box2d.Body
	humans      [barrelMaxRows * barrelMaxColumns]scenes.Human
	columnCount int
	rowCount    int

	shapeType int32
}

func CreateBenchmarkBarrel(context *SampleContext) Sampler {
	s := &BenchmarkBarrel{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 8.0, Y: 53.0}
		s.context.camera.zoom = 25.0 * 2.35
	}

	s.context.debugDraw.DrawJoints = false

	{
		gridSize := float32(1.0)

		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		y := float32(0.0)
		x := -40.0 * gridSize
		for range 81 {
			box := box2d.MakeOffsetBox(0.55*gridSize, 0.5*gridSize, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
			groundID.CreatePolygonShape(&shapeDef, &box)
			x += gridSize
		}

		y = gridSize
		x = -40.0 * gridSize
		for range 100 {
			box := box2d.MakeOffsetBox(0.5*gridSize, 0.55*gridSize, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
			groundID.CreatePolygonShape(&shapeDef, &box)
			y += gridSize
		}

		y = gridSize
		x = 40.0 * gridSize
		for range 100 {
			box := box2d.MakeOffsetBox(0.5*gridSize, 0.55*gridSize, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
			groundID.CreatePolygonShape(&shapeDef, &box)
			y += gridSize
		}

		segment := box2d.Segment{Point1: box2d.Vec2{X: -800.0, Y: -80.0}, Point2: box2d.Vec2{X: 800.0, Y: -80.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.shapeType = barrelCompoundShape

	s.CreateScene()
	return s
}

func (s *BenchmarkBarrel) CreateScene() {
	scenes.SetRandomSeed(42)

	for i := range barrelMaxRows * barrelMaxColumns {
		if s.bodies[i] != (box2d.Body{}) {
			s.bodies[i].Destroy()
			s.bodies[i] = box2d.Body{}
		}

		if s.humans[i].IsSpawned {
			scenes.DestroyHuman(&s.humans[i])
		}
	}

	s.columnCount = barrelMaxColumns
	s.rowCount = barrelMaxRows
	if isDebug {
		s.columnCount = 10
		s.rowCount = 40
	}

	if s.shapeType == barrelCompoundShape {
		if !isDebug {
			s.columnCount = 20
		}
	} else if s.shapeType == barrelHumanShape {
		if isDebug {
			s.rowCount = 5
			s.columnCount = 10
		} else {
			s.rowCount = 30
		}
	}

	rad := float32(0.5)

	shift := float32(1.15)
	centerx := shift * float32(s.columnCount) / 2.0
	centery := shift / 2.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	if s.shapeType == barrelMixShape {
		bodyDef.AngularDamping = 0.3
	}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.5

	capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.25}, Center2: box2d.Vec2{X: 0.0, Y: 0.25}, Radius: rad}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: rad}

	points := []box2d.Vec2{{X: -0.1, Y: -0.5}, {X: 0.1, Y: -0.5}, {X: 0.0, Y: 0.5}}
	wedgeHull := box2d.ComputeHull(points)
	wedge := box2d.MakePolygon(&wedgeHull, 0.0)

	var vertices [3]box2d.Vec2
	vertices[0] = box2d.Vec2{X: -1.0, Y: 0.0}
	vertices[1] = box2d.Vec2{X: 0.5, Y: 1.0}
	vertices[2] = box2d.Vec2{X: 0.0, Y: 2.0}
	hull := box2d.ComputeHull(vertices[:])
	left := box2d.MakePolygon(&hull, 0.0)

	vertices[0] = box2d.Vec2{X: 1.0, Y: 0.0}
	vertices[1] = box2d.Vec2{X: -0.5, Y: 1.0}
	vertices[2] = box2d.Vec2{X: 0.0, Y: 2.0}
	hull = box2d.ComputeHull(vertices[:])
	right := box2d.MakePolygon(&hull, 0.0)

	side := float32(-0.1)
	extray := float32(0.5)

	if s.shapeType == barrelCompoundShape {
		extray = 0.25
		side = 0.25
		shift = 2.0
		centerx = shift*float32(s.columnCount)/2.0 - 1.0
	} else if s.shapeType == barrelHumanShape {
		extray = 0.5
		side = 0.55
		shift = 2.5
		centerx = shift * float32(s.columnCount) / 2.0
	}

	index := 0
	yStart := float32(100.0)
	if s.shapeType == barrelHumanShape {
		yStart = 2.0
	}

	for i := range s.columnCount {
		x := float32(i)*shift - centerx

		for j := range s.rowCount {
			y := float32(j)*(shift+extray) + centery + yStart

			bodyDef.Position = box2d.Pos{X: posScalar(x + side), Y: posScalar(y)}
			side = -side

			switch s.shapeType {
			case barrelCircleShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)
				circle.Radius = scenes.RandomFloatRange(0.25, 0.75)
				shapeDef.Material.RollingResistance = 0.2
				s.bodies[index].CreateCircleShape(&shapeDef, &circle)
			case barrelCapsuleShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)
				capsule.Radius = scenes.RandomFloatRange(0.25, 0.5)
				length := scenes.RandomFloatRange(0.25, 1.0)
				capsule.Center1 = box2d.Vec2{X: 0.0, Y: -0.5 * length}
				capsule.Center2 = box2d.Vec2{X: 0.0, Y: 0.5 * length}
				shapeDef.Material.RollingResistance = 0.2
				s.bodies[index].CreateCapsuleShape(&shapeDef, &capsule)
			case barrelMixShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)

				mod := index % 3
				switch mod {
				case 0:
					circle.Radius = scenes.RandomFloatRange(0.25, 0.75)
					s.bodies[index].CreateCircleShape(&shapeDef, &circle)
				case 1:
					capsule.Radius = scenes.RandomFloatRange(0.25, 0.5)
					length := scenes.RandomFloatRange(0.25, 1.0)
					capsule.Center1 = box2d.Vec2{X: 0.0, Y: -0.5 * length}
					capsule.Center2 = box2d.Vec2{X: 0.0, Y: 0.5 * length}
					s.bodies[index].CreateCapsuleShape(&shapeDef, &capsule)
				case 2:
					width := scenes.RandomFloatRange(0.1, 0.5)
					height := scenes.RandomFloatRange(0.5, 0.75)
					box := box2d.MakeBox(width, height)

					value := scenes.RandomFloatRange(-1.0, 1.0)
					box.Radius = 0.25 * max(0.0, value)
					s.bodies[index].CreatePolygonShape(&shapeDef, &box)
				default:
					wedge.Radius = scenes.RandomFloatRange(0.1, 0.25)
					s.bodies[index].CreatePolygonShape(&shapeDef, &wedge)
				}
			case barrelCompoundShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)

				s.bodies[index].CreatePolygonShape(&shapeDef, &left)
				s.bodies[index].CreatePolygonShape(&shapeDef, &right)
			case barrelHumanShape:
				scale := float32(3.5)
				jointFriction := float32(0.05)
				jointHertz := float32(5.0)
				jointDamping := float32(0.5)
				scenes.CreateHuman(&s.humans[index], s.worldID, bodyDef.Position, scale, jointFriction, jointHertz, jointDamping,
					index+1, nil, false)
			}

			index += 1
		}
	}
}

func (s *BenchmarkBarrel) DrawControls() bool {
	changed := false
	shapeTypes := []string{"Circle", "Capsule", "Mix", "Compound", "Human"}

	shapeType := s.shapeType
	changed = changed || imgui.ComboStrarr("Shape", &shapeType, shapeTypes, int32(len(shapeTypes)))
	s.shapeType = shapeType

	changed = changed || imgui.Button("Reset Scene")

	if changed {
		s.CreateScene()
	}

	return true
}

type BenchmarkBarrel24 struct {
	Sample
}

func CreateBenchmarkBarrel24(context *SampleContext) Sampler {
	s := &BenchmarkBarrel24{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 8.0, Y: 53.0}
		s.context.camera.zoom = 25.0 * 2.35
	}

	groundSize := float32(25.0)

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(groundSize, 1.2)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)

		bodyDef.Rotation = box2d.MakeRot(0.5 * box2d.Pi)
		bodyDef.Position = box2d.Pos{X: posScalar(groundSize), Y: posScalar(2.0 * groundSize)}
		groundID = s.worldID.CreateBody(&bodyDef)

		box = box2d.MakeBox(2.0*groundSize, 1.2)
		groundID.CreatePolygonShape(&shapeDef, &box)

		bodyDef.Position = box2d.Pos{X: posScalar(-groundSize), Y: posScalar(2.0 * groundSize)}
		groundID = s.worldID.CreateBody(&bodyDef)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	num := 26
	rad := float32(0.5)

	shift := rad * 2.0
	centerx := shift * float32(num) / 2.0
	centery := shift / 2.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.5

	cuboid := box2d.MakeSquare(0.5)

	numj := 5 * num
	if isDebug {
		numj = 5
	}
	for i := range num {
		x := float32(i)*shift - centerx

		for j := range numj {
			y := float32(j)*shift + centery + 2.0

			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}

			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &cuboid)
		}
	}

	return s
}

type BenchmarkCompounds struct {
	Sample
}

func CreateBenchmarkCompounds(context *SampleContext) Sampler {
	s := &BenchmarkCompounds{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 50.0}
		s.context.camera.zoom = 25.0 * 2.2
		s.context.enableSleep = false
	}

	scenes.CreateCompounds(s.worldID)
	return s
}

type BenchmarkTumbler struct {
	Sample
}

func CreateBenchmarkTumbler(context *SampleContext) Sampler {
	s := &BenchmarkTumbler{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.5, Y: 10.0}
		s.context.camera.zoom = 15.0
	}

	scenes.CreateTumbler(s.worldID)
	return s
}

type BenchmarkWasher struct {
	Sample
}

func CreateBenchmarkWasher(context *SampleContext) Sampler {
	s := &BenchmarkWasher{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.5, Y: 10.0}
		s.context.camera.zoom = 20.0
	}

	scenes.CreateWasher(s.worldID)
	return s
}

func (s *BenchmarkWasher) Step() {
	s.Sample.Step()

	hitCount := 0
	for range s.worldID.ContactHitEvents() {
		hitCount++
	}
	s.DrawScreenTextLine("hits = %d", hitCount)
}

type BenchmarkManyTumblers struct {
	Sample
	groundID box2d.Body

	rowCount    int32
	columnCount int32

	tumblerIDs   []box2d.Body
	positions    []box2d.Pos
	tumblerCount int

	bodyIDs   []box2d.Body
	bodyCount int
	bodyIndex int

	angularSpeed float32
}

func CreateBenchmarkManyTumblers(context *SampleContext) Sampler {
	s := &BenchmarkManyTumblers{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.0, Y: -5.5}
		s.context.camera.zoom = 25.0 * 3.4
		s.context.debugDraw.DrawJoints = false
	}

	bodyDef := box2d.DefaultBodyDef()
	s.groundID = s.worldID.CreateBody(&bodyDef)

	s.rowCount = 19
	s.columnCount = 19
	if isDebug {
		s.rowCount = 2
		s.columnCount = 2
	}

	s.angularSpeed = 25.0

	s.CreateScene()
	return s
}

func (s *BenchmarkManyTumblers) CreateTumbler(position box2d.Pos, index int) {
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.KinematicBody
	bodyDef.Position = position
	bodyDef.AngularVelocity = (box2d.Pi / 180.0) * s.angularSpeed
	bodyID := s.worldID.CreateBody(&bodyDef)
	s.tumblerIDs[index] = bodyID

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 50.0

	polygon := box2d.MakeOffsetBox(0.25, 2.0, box2d.Vec2{X: 2.0, Y: 0.0}, box2d.RotIdentity)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
	polygon = box2d.MakeOffsetBox(0.25, 2.0, box2d.Vec2{X: -2.0, Y: 0.0}, box2d.RotIdentity)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
	polygon = box2d.MakeOffsetBox(2.0, 0.25, box2d.Vec2{X: 0.0, Y: 2.0}, box2d.RotIdentity)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
	polygon = box2d.MakeOffsetBox(2.0, 0.25, box2d.Vec2{X: 0.0, Y: -2.0}, box2d.RotIdentity)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
}

func (s *BenchmarkManyTumblers) CreateScene() {
	for i := range s.bodyCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
		}
	}

	for i := range s.tumblerCount {
		s.tumblerIDs[i].Destroy()
	}

	s.tumblerCount = int(s.rowCount * s.columnCount)
	s.tumblerIDs = make([]box2d.Body, s.tumblerCount)
	s.positions = make([]box2d.Pos, s.tumblerCount)

	index := 0
	x := -4.0 * float32(s.rowCount)
	for range s.rowCount {
		y := -4.0 * float32(s.columnCount)
		for range s.columnCount {
			s.positions[index] = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			s.CreateTumbler(s.positions[index], index)
			index++
			y += 8.0
		}

		x += 8.0
	}

	bodiesPerTumbler := 50
	if isDebug {
		bodiesPerTumbler = 8
	}
	s.bodyCount = bodiesPerTumbler * s.tumblerCount

	s.bodyIDs = make([]box2d.Body, s.bodyCount)
	s.bodyIndex = 0
}

func (s *BenchmarkManyTumblers) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false
	changed = changed || imgui.SliderInt("Row Count", &s.rowCount, 1, 32)
	changed = changed || imgui.SliderInt("Column Count", &s.columnCount, 1, 32)

	if changed {
		s.CreateScene()
	}

	if imgui.SliderFloatV("Speed", &s.angularSpeed, 0.0, 100.0, "%.f", 0) {
		for i := range s.tumblerCount {
			s.tumblerIDs[i].SetAngularVelocity((box2d.Pi / 180.0) * s.angularSpeed)
			s.tumblerIDs[i].SetAwake(true)
		}
	}

	imgui.PopItemWidth()

	return true
}

func (s *BenchmarkManyTumblers) Step() {
	s.Sample.Step()

	if s.bodyIndex < s.bodyCount && (s.stepCount&0x7) == 0 {
		shapeDef := box2d.DefaultShapeDef()

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.1, Y: 0.0}, Center2: box2d.Vec2{X: 0.1, Y: 0.0}, Radius: 0.075}

		for i := range s.tumblerCount {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = s.positions[i]
			s.bodyIDs[s.bodyIndex] = s.worldID.CreateBody(&bodyDef)
			s.bodyIDs[s.bodyIndex].CreateCapsuleShape(&shapeDef, &capsule)

			s.bodyIndex += 1
		}
	}
}

type BenchmarkLargePyramid struct {
	Sample
}

func CreateBenchmarkLargePyramid(context *SampleContext) Sampler {
	s := &BenchmarkLargePyramid{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 50.0}
		s.context.camera.zoom = 25.0 * 2.2
		s.context.enableSleep = false
	}

	scenes.CreateLargePyramid(s.worldID)
	return s
}

type BenchmarkManyPyramids struct {
	Sample
}

func CreateBenchmarkManyPyramids(context *SampleContext) Sampler {
	s := &BenchmarkManyPyramids{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 23.0, Y: 72.5}
		s.context.camera.zoom = 165.0
		s.context.enableSleep = false
	}

	scenes.CreateManyPyramids(s.worldID)
	return s
}

const (
	createDestroyMaxBaseCount = 100
	createDestroyMaxBodyCount = createDestroyMaxBaseCount * (createDestroyMaxBaseCount + 1) / 2
)

type BenchmarkCreateDestroy struct {
	Sample
	createTime  float32
	destroyTime float32
	bodies      [createDestroyMaxBodyCount]box2d.Body
	bodyCount   int
	baseCount   int
	iterations  int
}

func CreateBenchmarkCreateDestroy(context *SampleContext) Sampler {
	s := &BenchmarkCreateDestroy{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 50.0}
		s.context.camera.zoom = 25.0 * 2.2
	}

	groundSize := float32(100.0)

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	box := box2d.MakeBox(groundSize, 1.0)
	shapeDef := box2d.DefaultShapeDef()
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.createTime = 0.0
	s.destroyTime = 0.0

	s.baseCount = 100
	s.iterations = 10
	if isDebug {
		s.baseCount = 40
		s.iterations = 1
	}
	s.bodyCount = 0
	return s
}

func (s *BenchmarkCreateDestroy) CreateScene() {
	ticks := time.Now()

	for i := range createDestroyMaxBodyCount {
		if s.bodies[i] != (box2d.Body{}) {
			s.bodies[i].Destroy()
			s.bodies[i] = box2d.Body{}
		}
	}

	s.destroyTime += millisecondsAndReset(&ticks)

	count := s.baseCount
	rad := float32(0.5)
	shift := rad * 2.0
	centerx := shift * float32(count) / 2.0
	centery := shift/2.0 + 1.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.5

	h := float32(0.5)
	box := box2d.MakeRoundedBox(h, h, 0.0)

	index := 0

	for i := range count {
		y := float32(i)*shift + centery

		for j := i; j < count; j++ {
			x := 0.5*float32(i)*shift + float32(j-i)*shift - centerx
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}

			s.bodies[index] = s.worldID.CreateBody(&bodyDef)
			s.bodies[index].CreatePolygonShape(&shapeDef, &box)

			index += 1
		}
	}

	s.createTime += millisecondsSince(ticks)

	s.bodyCount = index

	s.worldID.Step(1.0/60.0, 4)
}

func (s *BenchmarkCreateDestroy) Step() {
	s.createTime = 0.0
	s.destroyTime = 0.0

	for range s.iterations {
		s.CreateScene()
	}

	s.DrawScreenTextLine("total: create = %.6g ms, destroy = %.6g ms", s.createTime, s.destroyTime)

	createPerBody := 1000.0 * s.createTime / float32(s.iterations) / float32(s.bodyCount)
	destroyPerBody := 1000.0 * s.destroyTime / float32(s.iterations) / float32(s.bodyCount)
	s.DrawScreenTextLine("body: create = %.6g us, destroy = %.6g us", createPerBody, destroyPerBody)

	s.Sample.Step()
}

type BenchmarkSleep struct {
	Sample
}

func CreateBenchmarkSleep(context *SampleContext) Sampler {
	s := &BenchmarkSleep{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 20.0}
		s.context.camera.zoom = 25.0 * 6.0
	}

	scenes.CreateSleep(s.worldID)
	return s
}

func (s *BenchmarkSleep) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		scenes.StepSleep(s.worldID, s.stepCount)
	}

	s.Sample.Step()
}

type BenchmarkJointGrid struct {
	Sample
}

func CreateBenchmarkJointGrid(context *SampleContext) Sampler {
	s := &BenchmarkJointGrid{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 60.0, Y: -57.0}
		s.context.camera.zoom = 25.0 * 2.5
		s.context.enableSleep = false
	}

	scenes.CreateJointGrid(s.worldID)
	return s
}

type BenchmarkSmash struct {
	Sample
}

func CreateBenchmarkSmash(context *SampleContext) Sampler {
	s := &BenchmarkSmash{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 60.0, Y: 6.0}
		s.context.camera.zoom = 25.0 * 1.6
	}

	scenes.CreateSmash(s.worldID)
	return s
}

type BenchmarkLargeCompounds struct {
	Sample
}

func CreateBenchmarkLargeCompounds(context *SampleContext) Sampler {
	s := &BenchmarkLargeCompounds{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 18.0, Y: 115.0}
		s.context.camera.zoom = 25.0 * 5.5
	}

	grid := float32(1.0)
	height := 200
	width := 200
	if isDebug {
		height = 100
		width = 100
	}
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		for i := range height {
			y := grid * float32(i)
			for j := i; j < width; j++ {
				x := grid * float32(j)
				square := box2d.MakeOffsetBox(0.5*grid, 0.5*grid, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
				groundID.CreatePolygonShape(&shapeDef, &square)
			}
		}

		for i := range height {
			y := grid * float32(i)
			for j := i; j < width; j++ {
				x := -grid * float32(j)
				square := box2d.MakeOffsetBox(0.5*grid, 0.5*grid, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
				groundID.CreatePolygonShape(&shapeDef, &square)
			}
		}
	}

	{
		span := 20
		count := 5
		if isDebug {
			span = 5
		}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.UpdateBodyMass = false

		for m := range count {
			ybody := (100.0 + float32(m*span)) * grid

			for n := range count {
				xbody := -0.5*grid*float32(count*span) + float32(n*span)*grid
				bodyDef.Position = box2d.Pos{X: posScalar(xbody), Y: posScalar(ybody)}
				bodyID := s.worldID.CreateBody(&bodyDef)

				for i := range span {
					y := float32(i) * grid
					for j := range span {
						x := float32(j) * grid
						square := box2d.MakeOffsetBox(0.5*grid, 0.5*grid, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
						bodyID.CreatePolygonShape(&shapeDef, &square)
					}
				}

				bodyID.UpdateMassFromShapes()
			}
		}
	}

	return s
}

type BenchmarkKinematic struct {
	Sample
}

func CreateBenchmarkKinematic(context *SampleContext) Sampler {
	s := &BenchmarkKinematic{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 150.0
	}

	grid := float32(1.0)

	span := 100
	if isDebug {
		span = 20
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.KinematicBody
	bodyDef.AngularVelocity = 1.0

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Filter.CategoryBits = 1
	shapeDef.Filter.MaskBits = 2

	shapeDef.UpdateBodyMass = false

	bodyID := s.worldID.CreateBody(&bodyDef)

	for i := -span; i < span; i++ {
		y := float32(i) * grid
		for j := -span; j < span; j++ {
			x := float32(j) * grid
			square := box2d.MakeOffsetBox(0.5*grid, 0.5*grid, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
			bodyID.CreatePolygonShape(&shapeDef, &square)
		}
	}

	bodyID.UpdateMassFromShapes()
	return s
}

const (
	queryRayCast = iota
	queryCircleCast
	queryOverlap
)

type BenchmarkCast struct {
	Sample
	queryType int32

	origins      []box2d.Pos
	translations []box2d.Vec2
	minTime      float32
	buildTime    float32

	rowCount, columnCount int32
	drawIndex             int
	radius                float32
	fill                  float32
	ratio                 float32
	grid                  float32
	topDown               bool
}

func CreateBenchmarkCast(context *SampleContext) Sampler {
	s := &BenchmarkCast{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 500.0, Y: 500.0}
		s.context.camera.zoom = 25.0 * 21.0
	}

	s.queryType = queryCircleCast
	s.ratio = 5.0
	s.grid = 1.0
	s.fill = 0.1
	s.rowCount = 1000
	s.columnCount = 1000
	if isDebug {
		s.rowCount = 100
		s.columnCount = 100
	}
	s.minTime = 1e6
	s.drawIndex = 0
	s.topDown = false
	s.buildTime = 0.0
	s.radius = 0.1

	scenes.SetRandomSeed(1234)
	sampleCount := 10000
	if isDebug {
		sampleCount = 100
	}
	s.origins = make([]box2d.Pos, sampleCount)
	s.translations = make([]box2d.Vec2, sampleCount)
	extent := float32(s.rowCount) * s.grid

	for i := range sampleCount {
		rayStart := scenes.RandomPos(0.0, extent)
		rayEnd := scenes.RandomPos(0.0, extent)

		s.origins[i] = rayStart
		s.translations[i] = rayEnd.Sub(rayStart)
	}

	s.BuildScene()
	return s
}

func (s *BenchmarkCast) BuildScene() {
	scenes.SetRandomSeed(1234)
	s.worldID.Close()
	worldDef := box2d.DefaultWorldDef()
	s.worldID = box2d.NewWorld(&worldDef)

	ticks := time.Now()

	bodyDef := box2d.DefaultBodyDef()
	shapeDef := box2d.DefaultShapeDef()

	y := float32(0.0)

	for range s.rowCount {
		x := float32(0.0)

		for range s.columnCount {
			fillTest := scenes.RandomFloatRange(0.0, 1.0)
			if fillTest <= s.fill {
				bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
				bodyID := s.worldID.CreateBody(&bodyDef)

				ratio := scenes.RandomFloatRange(1.0, s.ratio)
				halfWidth := scenes.RandomFloatRange(0.05, 0.25)

				var box box2d.Polygon
				if scenes.RandomFloat() > 0.0 {
					box = box2d.MakeBox(ratio*halfWidth, halfWidth)
				} else {
					box = box2d.MakeBox(halfWidth, ratio*halfWidth)
				}

				category := scenes.RandomIntRange(0, 2)
				shapeDef.Filter.CategoryBits = 1 << category
				switch category {
				case 0:
					shapeDef.Material.CustomColor = uint32(box2d.ColorBox2DBlue)
				case 1:
					shapeDef.Material.CustomColor = uint32(box2d.ColorBox2DYellow)
				default:
					shapeDef.Material.CustomColor = uint32(box2d.ColorBox2DGreen)
				}

				bodyID.CreatePolygonShape(&shapeDef, &box)
			}

			x += s.grid
		}

		y += s.grid
	}

	if s.topDown {
		s.worldID.RebuildStaticTree()
	}

	s.buildTime = millisecondsSince(ticks)
	s.minTime = 1e6
}

func (s *BenchmarkCast) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false

	queryTypes := []string{"Ray", "Circle", "Overlap"}
	queryType := s.queryType
	if imgui.ComboStrarr("Query", &queryType, queryTypes, int32(len(queryTypes))) {
		s.queryType = queryType
		if s.queryType == queryOverlap {
			s.radius = 5.0
		} else {
			s.radius = 0.1
		}

		changed = true
	}

	if imgui.SliderIntV("rows", &s.rowCount, 0, 1000, "%d", 0) {
		changed = true
	}

	if imgui.SliderIntV("columns", &s.columnCount, 0, 1000, "%d", 0) {
		changed = true
	}

	if imgui.SliderFloatV("fill", &s.fill, 0.0, 1.0, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("grid", &s.grid, 0.5, 2.0, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("ratio", &s.ratio, 1.0, 10.0, "%.2f", 0) {
		changed = true
	}

	imgui.PopItemWidth()

	if imgui.Checkbox("top down", &s.topDown) {
		changed = true
	}

	if imgui.Button("Draw Next") {
		s.drawIndex = (s.drawIndex + 1) % len(s.origins)
	}

	if changed {
		s.BuildScene()
	}

	return true
}

type castResult struct {
	point    box2d.Pos
	fraction float32
	hit      bool
}

type overlapResult struct {
	points [32]box2d.Vec2
	count  int
}

func (s *BenchmarkCast) Step() {
	s.Sample.Step()

	filter := box2d.DefaultQueryFilter()
	filter.MaskBits = 1
	hitCount := 0
	nodeVisits := 0
	leafVisits := 0
	var ms float32
	sampleCount := len(s.origins)

	switch s.queryType {
	case queryRayCast:
		ticks := time.Now()

		var drawResult box2d.RayResult

		for i := range sampleCount {
			origin := s.origins[i]
			translation := s.translations[i]

			result := s.worldID.CastRayClosest(origin, translation, filter)

			if i == s.drawIndex {
				drawResult = result
			}

			nodeVisits += result.NodeVisits
			leafVisits += result.LeafVisits
			if result.Hit {
				hitCount += 1
			}
		}

		ms = millisecondsSince(ticks)

		s.minTime = min(s.minTime, ms)

		p1 := s.origins[s.drawIndex]
		p2 := s.origins[s.drawIndex].Add(s.translations[s.drawIndex])
		s.context.draw.DrawLine(p1, p2, box2d.ColorWhite)
		s.context.draw.DrawPoint(p1, 5.0, box2d.ColorGreen)
		s.context.draw.DrawPoint(p2, 5.0, box2d.ColorRed)
		if drawResult.Hit {
			s.context.draw.DrawPoint(drawResult.Point, 5.0, box2d.ColorWhite)
		}
	case queryCircleCast:
		ticks := time.Now()

		var drawResult castResult

		for i := range sampleCount {
			proxy := box2d.MakeProxy([]box2d.Vec2{box2d.Vec2Zero}, s.radius)
			translation := s.translations[i]

			var result castResult
			traversalResult := s.worldID.CastShape(s.origins[i], &proxy, translation, filter,
				func(shape box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
					result.point = point
					result.fraction = fraction
					result.hit = true
					return fraction
				})

			if i == s.drawIndex {
				drawResult = result
			}

			nodeVisits += traversalResult.NodeVisits
			leafVisits += traversalResult.LeafVisits
			if result.hit {
				hitCount += 1
			}
		}

		ms = millisecondsSince(ticks)

		s.minTime = min(s.minTime, ms)

		p1 := s.origins[s.drawIndex]
		p2 := s.origins[s.drawIndex].Add(s.translations[s.drawIndex])
		s.context.draw.DrawLine(p1, p2, box2d.ColorWhite)
		s.context.draw.DrawPoint(p1, 5.0, box2d.ColorGreen)
		s.context.draw.DrawPoint(p2, 5.0, box2d.ColorRed)
		if drawResult.hit {
			t := s.origins[s.drawIndex].Add(s.translations[s.drawIndex].Scale(drawResult.fraction))
			s.context.draw.DrawCircle(t, s.radius, box2d.ColorWhite)
			s.context.draw.DrawPoint(drawResult.point, 5.0, box2d.ColorWhite)
		}
	case queryOverlap:
		ticks := time.Now()

		var drawResult overlapResult
		extent := box2d.Vec2{X: s.radius, Y: s.radius}
		var result overlapResult

		for i := range sampleCount {
			aabb := box2d.AABB{LowerBound: extent.Neg(), UpperBound: extent}

			result.count = 0
			traversalResult := s.worldID.OverlapAABB(s.origins[i], aabb, filter, func(shape box2d.Shape) bool {
				if result.count < 32 {
					result.points[result.count] = shape.AABB().Center()
					result.count += 1
				}

				return true
			})

			if i == s.drawIndex {
				drawResult = result
			}

			nodeVisits += traversalResult.NodeVisits
			leafVisits += traversalResult.LeafVisits
			hitCount += result.count
		}

		ms = millisecondsSince(ticks)

		s.minTime = min(s.minTime, ms)

		origin := s.origins[s.drawIndex]
		aabb := box2d.AABB{
			LowerBound: box2d.Vec2{X: box2d.RoundDownFloat(float64(origin.X) - float64(extent.X)), Y: box2d.RoundDownFloat(float64(origin.Y) - float64(extent.Y))},
			UpperBound: box2d.Vec2{X: box2d.RoundUpFloat(float64(origin.X) + float64(extent.X)), Y: box2d.RoundUpFloat(float64(origin.Y) + float64(extent.Y))},
		}

		s.context.draw.DrawBounds(aabb, box2d.ColorWhite)

		for i := range drawResult.count {
			s.context.draw.DrawPoint(drawResult.points[i].Pos(), 5.0, box2d.ColorHotPink)
		}
	}

	s.DrawScreenTextLine("build time ms = %.6g", s.buildTime)
	s.DrawScreenTextLine("hit count = %d, node visits = %d, leaf visits = %d", hitCount, nodeVisits, leafVisits)
	s.DrawScreenTextLine("total ms = %.3f", ms)
	s.DrawScreenTextLine("min total ms = %.3f", s.minTime)

	aveRayCost := 1000.0 * s.minTime / float32(sampleCount)
	s.DrawScreenTextLine("average us = %.2f", aveRayCost)
}

type BenchmarkSpinner struct {
	Sample
}

func CreateBenchmarkSpinner(context *SampleContext) Sampler {
	s := &BenchmarkSpinner{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 32.0}
		s.context.camera.zoom = 42.0
	}

	scenes.CreateSpinner(s.worldID)
	return s
}

type BenchmarkRain struct {
	Sample
}

func CreateBenchmarkRain(context *SampleContext) Sampler {
	s := &BenchmarkRain{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 110.0}
		s.context.camera.zoom = 125.0
		s.context.enableSleep = true
	}

	s.context.debugDraw.DrawJoints = false

	scenes.CreateRain(s.worldID)
	return s
}

func (s *BenchmarkRain) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		scenes.StepRain(s.worldID, s.stepCount)
	}

	s.Sample.Step()
}

const shapeDistanceCount = 10000

type BenchmarkShapeDistance struct {
	Sample
	transformAs     []box2d.WorldTransform
	transformBs     []box2d.WorldTransform
	outputs         []box2d.DistanceOutput
	polygonA        box2d.Polygon
	polygonB        box2d.Polygon
	minMilliseconds float32
	drawIndex       int32
}

func CreateBenchmarkShapeDistance(context *SampleContext) Sampler {
	s := &BenchmarkShapeDistance{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 3.0
	}

	{
		var points [8]box2d.Vec2
		q := box2d.MakeRot(2.0 * box2d.Pi / 8.0)
		p := box2d.Vec2{X: 0.5, Y: 0.0}
		points[0] = p
		for i := 1; i < 8; i++ {
			points[i] = q.Rotate(points[i-1])
		}

		hull := box2d.ComputeHull(points[:])
		s.polygonA = box2d.MakePolygon(&hull, 0.0)
	}

	{
		var points [8]box2d.Vec2
		q := box2d.MakeRot(2.0 * box2d.Pi / 8.0)
		p := box2d.Vec2{X: 0.5, Y: 0.0}
		points[0] = p
		for i := 1; i < 8; i++ {
			points[i] = q.Rotate(points[i-1])
		}

		hull := box2d.ComputeHull(points[:])
		s.polygonB = box2d.MakePolygon(&hull, 0.1)
	}

	s.transformAs = make([]box2d.WorldTransform, shapeDistanceCount)
	s.transformBs = make([]box2d.WorldTransform, shapeDistanceCount)
	s.outputs = make([]box2d.DistanceOutput, shapeDistanceCount)

	scenes.SetRandomSeed(42)
	for i := range shapeDistanceCount {
		s.transformAs[i] = box2d.WorldTransform{P: scenes.RandomPos(-0.1, 0.1), Q: scenes.RandomRot()}
		s.transformBs[i] = box2d.WorldTransform{P: scenes.RandomPos(0.25, 2.0), Q: scenes.RandomRot()}
	}

	s.drawIndex = 0
	s.minMilliseconds = math.MaxFloat32
	return s
}

func (s *BenchmarkShapeDistance) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderInt("draw index", &s.drawIndex, 0, shapeDistanceCount-1)
	imgui.PopItemWidth()

	return true
}

func (s *BenchmarkShapeDistance) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		var input box2d.DistanceInput
		input.ProxyA = box2d.MakeProxy(s.polygonA.Vertices[:s.polygonA.Count], s.polygonA.Radius)
		input.ProxyB = box2d.MakeProxy(s.polygonB.Vertices[:s.polygonB.Count], s.polygonB.Radius)
		input.UseRadii = true
		totalIterations := 0

		start := time.Now()
		for i := range shapeDistanceCount {
			var cache box2d.SimplexCache
			input.Transform = s.transformAs[i].InvMul(s.transformBs[i])
			s.outputs[i] = box2d.ShapeDistance(&input, &cache, nil)
			totalIterations += s.outputs[i].Iterations
		}

		ms := millisecondsSince(start)
		s.minMilliseconds = min(s.minMilliseconds, ms)

		s.DrawScreenTextLine("count = %d", shapeDistanceCount)
		s.DrawScreenTextLine("min ms = %.6g, ave us = %.6g", s.minMilliseconds, 1000.0*s.minMilliseconds/float32(shapeDistanceCount))
		s.DrawScreenTextLine("average iterations = %.6g", float32(totalIterations)/float32(shapeDistanceCount))
	}

	xfA := s.transformAs[s.drawIndex]
	xfB := s.transformBs[s.drawIndex]
	output := s.outputs[s.drawIndex]

	worldPointA := xfA.Apply(output.PointA)
	worldPointB := xfA.Apply(output.PointB)
	worldNormal := xfA.Q.Rotate(output.Normal)

	s.context.draw.DrawSolidPolygon(xfA, s.polygonA.Vertices[:s.polygonA.Count], s.polygonA.Radius, box2d.ColorBox2DGreen)
	s.context.draw.DrawSolidPolygon(xfB, s.polygonB.Vertices[:s.polygonB.Count], s.polygonB.Radius, box2d.ColorBox2DBlue)
	s.context.draw.DrawLine(worldPointA, worldPointB, box2d.ColorDimGray)
	s.context.draw.DrawPoint(worldPointA, 10.0, box2d.ColorWhite)
	s.context.draw.DrawPoint(worldPointB, 10.0, box2d.ColorWhite)
	s.context.draw.DrawLine(worldPointA, worldPointA.Add(worldNormal.Scale(0.5)), box2d.ColorYellow)
	s.DrawScreenTextLine("distance = %.6g", output.Distance)

	s.Sample.Step()
}

type shapeUserData struct {
	row    int
	active bool
}

const (
	sensorColumnCount = 40
	sensorRowCount    = 40
)

type BenchmarkSensor struct {
	Sample
	maxBeginCount  int
	maxEndCount    int
	passiveSensors [sensorRowCount]shapeUserData
	activeSensor   shapeUserData
	lastStepCount  int
	filterRow      int
}

func CreateBenchmarkSensor(context *SampleContext) Sampler {
	s := &BenchmarkSensor{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 105.0}
		s.context.camera.zoom = 125.0
	}

	s.worldID.SetCustomFilterCallback(s.Filter)

	s.activeSensor.row = 0
	s.activeSensor.active = true

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	{
		gridSize := float32(3.0)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		shapeDef.UserData = &s.activeSensor

		y := float32(0.0)
		x := -40.0 * gridSize
		for range 81 {
			box := box2d.MakeOffsetBox(0.45*gridSize, 0.45*gridSize, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity)
			groundID.CreatePolygonShape(&shapeDef, &box)
			x += gridSize
		}
	}

	scenes.SetRandomSeed(42)

	shift := float32(5.0)
	xCenter := 0.5 * shift * sensorColumnCount

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	shapeDef.EnableSensorEvents = true

	yStart := float32(10.0)
	s.filterRow = sensorRowCount >> 1

	for j := range sensorRowCount {
		s.passiveSensors[j].row = j
		s.passiveSensors[j].active = false
		shapeDef.UserData = &s.passiveSensors[j]

		if j == s.filterRow {
			shapeDef.EnableCustomFiltering = true
			shapeDef.Material.CustomColor = uint32(box2d.ColorFuchsia)
		} else {
			shapeDef.EnableCustomFiltering = false
			shapeDef.Material.CustomColor = 0
		}

		y := float32(j)*shift + yStart
		for i := range sensorColumnCount {
			x := float32(i)*shift - xCenter
			box := box2d.MakeOffsetRoundedBox(0.5, 0.5, box2d.Vec2{X: x, Y: y}, box2d.RotIdentity, 0.1)
			groundID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	s.maxBeginCount = 0
	s.maxEndCount = 0
	s.lastStepCount = 0
	return s
}

func (s *BenchmarkSensor) CreateRow(y float32) {
	shift := float32(5.0)
	xCenter := 0.5 * shift * sensorColumnCount

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.0
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -5.0}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	for i := range sensorColumnCount {
		yOffset := scenes.RandomFloatRange(-1.0, 1.0)
		bodyDef.Position = box2d.Pos{X: posScalar(shift*float32(i) - xCenter), Y: posScalar(y + yOffset)}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
}

func (s *BenchmarkSensor) Step() {
	s.Sample.Step()

	if s.stepCount == s.lastStepCount {
		return
	}

	zombies := map[box2d.Body]struct{}{}

	beginCount := 0
	for event := range s.worldID.SensorBeginEvents() {
		beginCount++

		userData := event.SensorShape.UserData().(*shapeUserData)

		if userData.active {
			zombies[event.VisitorShape.Body()] = struct{}{}
		} else {
			if userData.row == s.filterRow {
				panic("custom filter failed")
			}

			surfaceMaterial := event.VisitorShape.SurfaceMaterial()
			surfaceMaterial.CustomColor = uint32(box2d.ColorLime)
			event.VisitorShape.SetSurfaceMaterial(&surfaceMaterial)
		}
	}

	endCount := 0
	for event := range s.worldID.SensorEndEvents() {
		endCount++

		if !event.VisitorShape.IsValid() {
			continue
		}

		surfaceMaterial := event.VisitorShape.SurfaceMaterial()
		surfaceMaterial.CustomColor = 0
		event.VisitorShape.SetSurfaceMaterial(&surfaceMaterial)
	}

	sortedZombies := make([]box2d.Body, 0, len(zombies))
	for bodyID := range zombies {
		sortedZombies = append(sortedZombies, bodyID)
	}
	slices.SortFunc(sortedZombies, func(a, b box2d.Body) int {
		switch {
		case a.ID() < b.ID():
			return -1
		case a.ID() > b.ID():
			return 1
		}
		return 0
	})
	for _, bodyID := range sortedZombies {
		bodyID.Destroy()
	}

	delay := 0x1F

	if (s.stepCount & delay) == 0 {
		s.CreateRow(10.0 + sensorRowCount*5.0)
	}

	s.lastStepCount = s.stepCount

	s.maxBeginCount = max(beginCount, s.maxBeginCount)
	s.maxEndCount = max(endCount, s.maxEndCount)
	s.DrawScreenTextLine("max begin touch events = %d", s.maxBeginCount)
	s.DrawScreenTextLine("max end touch events = %d", s.maxEndCount)
}

func (s *BenchmarkSensor) Filter(idA, idB box2d.Shape) bool {
	var userData *shapeUserData
	if idA.IsSensor() {
		userData, _ = idA.UserData().(*shapeUserData)
	} else if idB.IsSensor() {
		userData, _ = idB.UserData().(*shapeUserData)
	}

	if userData != nil {
		return userData.active || userData.row != s.filterRow
	}

	return true
}

type BenchmarkCapacity struct {
	Sample
	square     box2d.Polygon
	reachCount int
	done       bool
}

func CreateBenchmarkCapacity(context *SampleContext) Sampler {
	s := &BenchmarkCapacity{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 150.0}
		s.context.camera.zoom = 200.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position.Y = -5.0
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(800.0, 5.0)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.square = box2d.MakeSquare(0.5)
	s.done = false
	s.reachCount = 0
	return s
}

func (s *BenchmarkCapacity) Step() {
	s.Sample.Step()

	millisecondLimit := float32(20.0)

	profile := s.worldID.Profile()
	if profile.Step > millisecondLimit {
		s.reachCount += 1
		if s.reachCount > 60 {
			s.done = true
		}
	} else {
		s.reachCount = 0
	}

	if s.done {
		return
	}

	if (s.stepCount & 0x1F) != 0x1F {
		return
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position.Y = 200.0

	shapeDef := box2d.DefaultShapeDef()

	count := 200
	x := -1.0 * float32(count)
	for range count {
		bodyDef.Position.X = posScalar(x)
		bodyDef.Position.Y += 0.5

		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &s.square)

		x += 2.0
	}
}

type BenchmarkJunkyard struct {
	Sample
}

func CreateBenchmarkJunkyard(context *SampleContext) Sampler {
	s := &BenchmarkJunkyard{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 8.0, Y: 25.0}
		s.context.camera.zoom = 60.0
	}

	scenes.CreateJunkyard(s.worldID)
	return s
}

func (s *BenchmarkJunkyard) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		scenes.StepJunkyard(s.worldID, s.stepCount)
	}

	s.Sample.Step()
}

type BenchmarkQueries struct {
	Sample
	drawIndex  int32
	minTime    float32
	lastTime   float32
	hitCount   int
	nodeVisits int
	leafVisits int
}

func CreateBenchmarkQueries(context *SampleContext) Sampler {
	s := &BenchmarkQueries{}
	s.init(context, s, true)

	scenes.CreateQueries(s.worldID)

	if !s.context.restart {
		extent := scenes.GetQueryBenchmarkExtent()
		s.context.camera.center = box2d.Pos{X: posScalar(0.5 * extent), Y: posScalar(0.5 * extent)}
		s.context.camera.zoom = 0.55 * extent
	}

	s.drawIndex = 0
	s.minTime = 1e6
	s.lastTime = 0.0
	s.hitCount = 0
	s.nodeVisits = 0
	s.leafVisits = 0
	return s
}

func (s *BenchmarkQueries) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderIntV("draw", &s.drawIndex, 0, int32(scenes.GetQueryBenchmarkCount()-1), "%d", 0)
	imgui.PopItemWidth()
	return false
}

func (s *BenchmarkQueries) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		before := scenes.GetQueryBenchmarkStats()
		ticks := time.Now()
		s.hitCount = int(scenes.StepQueries(s.worldID, s.stepCount))
		s.lastTime = millisecondsSince(ticks)
		s.minTime = min(s.minTime, s.lastTime)
		after := scenes.GetQueryBenchmarkStats()
		s.nodeVisits = after.NodeVisits - before.NodeVisits
		s.leafVisits = after.LeafVisits - before.LeafVisits
	}

	s.Sample.Step()

	origin, translation := scenes.GetQueryBenchmarkRay(int(s.drawIndex))
	end := origin.Add(translation)
	s.context.draw.DrawLine(origin, end, box2d.ColorWhite)
	s.context.draw.DrawPoint(origin, 5.0, box2d.ColorGreen)
	s.context.draw.DrawPoint(end, 5.0, box2d.ColorRed)

	box := box2d.AABB{
		LowerBound: box2d.Vec2{X: box2d.RoundDownFloat(float64(origin.X) - 5.0), Y: box2d.RoundDownFloat(float64(origin.Y) - 5.0)},
		UpperBound: box2d.Vec2{X: box2d.RoundUpFloat(float64(origin.X) + 5.0), Y: box2d.RoundUpFloat(float64(origin.Y) + 5.0)},
	}
	s.context.draw.DrawBounds(box, box2d.ColorWhite)

	s.DrawScreenTextLine("%d queries of each kind and filter, %d hits", scenes.GetQueryBenchmarkCount(), s.hitCount)
	s.DrawScreenTextLine("node visits = %d, leaf visits = %d", s.nodeVisits, s.leafVisits)
	s.DrawScreenTextLine("queries ms = %.3f, min = %.3f", s.lastTime, s.minTime)
}

func init() {
	RegisterSample("Benchmark", "Barrel", CreateBenchmarkBarrel)
	RegisterSample("Benchmark", "Barrel 2.4", CreateBenchmarkBarrel24)
	RegisterSample("Benchmark", "Compounds", CreateBenchmarkCompounds)
	RegisterSample("Benchmark", "Tumbler", CreateBenchmarkTumbler)
	RegisterSample("Benchmark", "Washer", CreateBenchmarkWasher)
	RegisterSample("Benchmark", "Many Tumblers", CreateBenchmarkManyTumblers)
	RegisterSample("Benchmark", "Large Pyramid", CreateBenchmarkLargePyramid)
	RegisterSampleWithCapacity("Benchmark", "Many Pyramids", CreateBenchmarkManyPyramids, scenes.GetManyPyramidsCapacity)
	RegisterSample("Benchmark", "CreateDestroy", CreateBenchmarkCreateDestroy)
	RegisterSample("Benchmark", "Sleep", CreateBenchmarkSleep)
	RegisterSample("Benchmark", "Joint Grid", CreateBenchmarkJointGrid)
	RegisterSample("Benchmark", "Smash", CreateBenchmarkSmash)
	RegisterSample("Benchmark", "Large Compounds", CreateBenchmarkLargeCompounds)
	RegisterSample("Benchmark", "Kinematic", CreateBenchmarkKinematic)
	RegisterSample("Benchmark", "Cast", CreateBenchmarkCast)
	RegisterSample("Benchmark", "Spinner", CreateBenchmarkSpinner)
	RegisterSample("Benchmark", "Rain", CreateBenchmarkRain)
	RegisterSample("Benchmark", "Shape Distance", CreateBenchmarkShapeDistance)
	RegisterSample("Benchmark", "Sensor", CreateBenchmarkSensor)
	RegisterSample("Benchmark", "Capacity", CreateBenchmarkCapacity)
	RegisterSample("Benchmark", "Junkyard", CreateBenchmarkJunkyard)
	RegisterSample("Benchmark", "Queries", CreateBenchmarkQueries)
}
