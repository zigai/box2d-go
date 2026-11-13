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

const (
	benchmarkRandomSeed        = 42
	castRandomSeed             = 1234
	controlWidthInFonts        = 6.0
	degreesPerHalfTurn         = 180.0
	castPointSize              = 5.0
	microsecondsPerMillisecond = 1000.0
)

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

const (
	createDestroyMaxBaseCount = 100
	createDestroyMaxBodyCount = createDestroyMaxBaseCount * (createDestroyMaxBaseCount + 1) / 2
)

const (
	queryRayCast = iota
	queryCircleCast
	queryOverlap
)
const shapeDistanceCount = 10000
const (
	sensorColumnCount     = 40
	sensorRowCount        = 40
	sensorGridSpacing     = 5.0
	sensorGridStartHeight = 10.0
)

type BenchmarkBarrel struct {
	Sample

	bodies      [barrelMaxRows * barrelMaxColumns]box2d.Body   //exhaustruct:optional
	humans      [barrelMaxRows * barrelMaxColumns]scenes.Human //exhaustruct:optional
	columnCount int
	rowCount    int

	shapeType int32
}

type BenchmarkBarrel24 struct {
	Sample
}
type BenchmarkCompounds struct {
	Sample
}
type BenchmarkTumbler struct {
	Sample
}
type BenchmarkWasher struct {
	Sample
}
type BenchmarkManyTumblers struct {
	Sample

	groundID box2d.Body //exhaustruct:optional

	rowCount    int32
	columnCount int32

	tumblerIDs   []box2d.Body //exhaustruct:optional
	positions    []box2d.Pos  //exhaustruct:optional
	tumblerCount int          //exhaustruct:optional

	bodyIDs   []box2d.Body //exhaustruct:optional
	bodyCount int          //exhaustruct:optional
	bodyIndex int          //exhaustruct:optional

	angularSpeed float32
}
type BenchmarkLargePyramid struct {
	Sample
}
type BenchmarkManyPyramids struct {
	Sample
}
type BenchmarkCreateDestroy struct {
	Sample

	createTime  float32                               //exhaustruct:optional
	destroyTime float32                               //exhaustruct:optional
	bodies      [createDestroyMaxBodyCount]box2d.Body //exhaustruct:optional
	bodyCount   int                                   //exhaustruct:optional
	baseCount   int
	iterations  int
}
type BenchmarkSleep struct {
	Sample
}
type BenchmarkJointGrid struct {
	Sample
}
type BenchmarkSmash struct {
	Sample
}
type BenchmarkLargeCompounds struct {
	Sample
}
type BenchmarkKinematic struct {
	Sample
}
type BenchmarkCast struct {
	Sample

	queryType int32

	origins      []box2d.Pos  //exhaustruct:optional
	translations []box2d.Vec2 //exhaustruct:optional
	minTime      float32
	buildTime    float32 //exhaustruct:optional

	rowCount, columnCount int32
	drawIndex             int //exhaustruct:optional
	radius                float32
	fill                  float32
	ratio                 float32
	grid                  float32
	topDown               bool //exhaustruct:optional
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
type BenchmarkSpinner struct {
	Sample
}
type BenchmarkRain struct {
	Sample
}
type BenchmarkShapeDistance struct {
	Sample

	transformAs     []box2d.WorldTransform //exhaustruct:optional
	transformBs     []box2d.WorldTransform //exhaustruct:optional
	outputs         []box2d.DistanceOutput //exhaustruct:optional
	polygonA        box2d.Polygon
	polygonB        box2d.Polygon
	minMilliseconds float32
	drawIndex       int32 //exhaustruct:optional
}
type shapeUserData struct {
	row    int
	active bool
}
type BenchmarkSensor struct {
	Sample

	maxBeginCount  int                           //exhaustruct:optional
	maxEndCount    int                           //exhaustruct:optional
	passiveSensors [sensorRowCount]shapeUserData //exhaustruct:optional
	activeSensor   shapeUserData
	lastStepCount  int //exhaustruct:optional
	filterRow      int
}
type BenchmarkCapacity struct {
	Sample

	square     box2d.Polygon
	reachCount int  //exhaustruct:optional
	done       bool //exhaustruct:optional
}
type BenchmarkJunkyard struct {
	Sample
}
type BenchmarkQueries struct {
	Sample

	drawIndex  int32 //exhaustruct:optional
	minTime    float32
	lastTime   float32 //exhaustruct:optional
	hitCount   int     //exhaustruct:optional
	nodeVisits int     //exhaustruct:optional
	leafVisits int     //exhaustruct:optional
}

func CreateBenchmarkBarrel(context *SampleContext) Sampler {
	const (
		cameraX           = 8.0
		cameraY           = 53.0
		baseZoom          = 25.0
		cameraZoomScale   = 2.35
		wallHalfLength    = 0.55
		wallHalfThickness = 0.5
		wallGridHalfWidth = 40.0
		floorHalfWidth    = 800.0
	)

	s := &BenchmarkBarrel{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		shapeType: barrelCompoundShape, columnCount: barrelMaxColumns, rowCount: barrelMaxRows,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = baseZoom * cameraZoomScale
	}

	s.context.debugDraw.DrawJoints = false

	{
		gridSize := float32(1.0)

		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		y := float32(0.0)

		x := -wallGridHalfWidth * gridSize
		for range 81 {
			box := box2d.MakeOffsetBox(
				wallHalfLength*gridSize,
				wallHalfThickness*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.RotIdentity,
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			x += gridSize
		}

		y = gridSize

		x = -wallGridHalfWidth * gridSize
		for range 100 {
			box := box2d.MakeOffsetBox(
				wallHalfThickness*gridSize,
				wallHalfLength*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.RotIdentity,
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			y += gridSize
		}

		y = gridSize

		x = wallGridHalfWidth * gridSize
		for range 100 {
			box := box2d.MakeOffsetBox(
				wallHalfThickness*gridSize,
				wallHalfLength*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.RotIdentity,
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			y += gridSize
		}

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -floorHalfWidth, Y: -80.0},
			Point2: box2d.Vec2{X: floorHalfWidth, Y: -80.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.CreateScene()

	return s
}

func (s *BenchmarkBarrel) CreateScene() {
	const (
		initialRadius         = 0.5
		initialSpacing        = 1.15
		centerDivisor         = 2.0
		capsuleHalfLength     = 0.25
		wedgeHalfWidth        = 0.1
		wedgeHalfHeight       = 0.5
		compoundInnerX        = 0.5
		compoundHeight        = 2.0
		initialExtraY         = 0.5
		shapeStartY           = 100.0
		circleMinRadius       = 0.25
		circleMaxRadius       = 0.75
		capsuleMinRadius      = 0.25
		capsuleMaxRadius      = 0.5
		capsuleCenterFraction = 0.5
		capsuleMinLength      = 0.25
		capsuleMaxLength      = 1.0
		humanScale            = 3.5
		humanJointFriction    = 0.05
		humanJointHertz       = 5.0
		humanJointDamping     = 0.5
	)

	scenes.SetRandomSeed(benchmarkRandomSeed)

	s.resetBodies()
	s.configureGrid()

	rad := float32(initialRadius)

	shift := float32(initialSpacing)
	centerx := shift * float32(s.columnCount) / centerDivisor
	centery := shift / centerDivisor

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	if s.shapeType == barrelMixShape {
		bodyDef.AngularDamping = 0.3
	}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.5

	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: 0.0, Y: -capsuleHalfLength},
		Center2: box2d.Vec2{X: 0.0, Y: capsuleHalfLength},
		Radius:  rad,
	}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: rad}

	points := []box2d.Vec2{
		{X: -wedgeHalfWidth, Y: -wedgeHalfHeight},
		{X: wedgeHalfWidth, Y: -wedgeHalfHeight},
		{X: 0.0, Y: wedgeHalfHeight},
	}
	wedgeHull := box2d.ComputeHull(points)
	wedge := box2d.MakePolygon(&wedgeHull, 0.0)

	var vertices [3]box2d.Vec2

	vertices[0] = box2d.Vec2{X: -1.0, Y: 0.0}
	vertices[1] = box2d.Vec2{X: compoundInnerX, Y: 1.0}
	vertices[2] = box2d.Vec2{X: 0.0, Y: compoundHeight}
	hull := box2d.ComputeHull(vertices[:])
	left := box2d.MakePolygon(&hull, 0.0)

	vertices[0] = box2d.Vec2{X: 1.0, Y: 0.0}
	vertices[1] = box2d.Vec2{X: -compoundInnerX, Y: 1.0}
	vertices[2] = box2d.Vec2{X: 0.0, Y: compoundHeight}
	hull = box2d.ComputeHull(vertices[:])
	right := box2d.MakePolygon(&hull, 0.0)

	side := float32(-0.1)
	extray := float32(initialExtraY)

	switch s.shapeType {
	case barrelCompoundShape:
		extray = 0.25
		side = 0.25
		shift = 2.0
		centerx = shift*float32(s.columnCount)/centerDivisor - 1.0
	case barrelHumanShape:
		extray = 0.5
		side = 0.55
		shift = 2.5
		centerx = shift * float32(s.columnCount) / centerDivisor
	}

	index := 0

	yStart := float32(shapeStartY)
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
				circle.Radius = scenes.RandomFloatRange(circleMinRadius, circleMaxRadius)
				shapeDef.Material.RollingResistance = 0.2
				s.bodies[index].CreateCircleShape(&shapeDef, &circle)
			case barrelCapsuleShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)
				capsule.Radius = scenes.RandomFloatRange(capsuleMinRadius, capsuleMaxRadius)
				length := scenes.RandomFloatRange(capsuleMinLength, capsuleMaxLength)
				capsule.Center1 = box2d.Vec2{X: 0.0, Y: -capsuleCenterFraction * length}
				capsule.Center2 = box2d.Vec2{X: 0.0, Y: capsuleCenterFraction * length}
				shapeDef.Material.RollingResistance = 0.2
				s.bodies[index].CreateCapsuleShape(&shapeDef, &capsule)
			case barrelMixShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)

				s.createMixedShape(index, &shapeDef, &circle, &capsule, &wedge)
			case barrelCompoundShape:
				s.bodies[index] = s.worldID.CreateBody(&bodyDef)

				s.bodies[index].CreatePolygonShape(&shapeDef, &left)
				s.bodies[index].CreatePolygonShape(&shapeDef, &right)
			case barrelHumanShape:
				scale := float32(humanScale)
				jointFriction := float32(humanJointFriction)
				jointHertz := float32(humanJointHertz)
				jointDamping := float32(humanJointDamping)
				scenes.CreateHuman(
					&s.humans[index],
					s.worldID,
					bodyDef.Position,
					scale,
					jointFriction,
					jointHertz,
					jointDamping,
					index+1,
					nil,
					false,
				)
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

func (s *BenchmarkBarrel) resetBodies() {
	for i := range barrelMaxRows * barrelMaxColumns {
		if s.bodies[i] != (box2d.Body{}) {
			s.bodies[i].Destroy()
			s.bodies[i] = box2d.Body{}
		}

		if s.humans[i].IsSpawned {
			scenes.DestroyHuman(&s.humans[i])
		}
	}
}

func (s *BenchmarkBarrel) configureGrid() {
	s.columnCount = barrelMaxColumns

	s.rowCount = barrelMaxRows
	if isDebug {
		s.columnCount = 10
		s.rowCount = 40
	}

	switch s.shapeType {
	case barrelCompoundShape:
		if !isDebug {
			s.columnCount = 20
		}
	case barrelHumanShape:
		if isDebug {
			s.rowCount = 5
			s.columnCount = 10
		} else {
			s.rowCount = 30
		}
	}
}

func (s *BenchmarkBarrel) createMixedShape(
	index int, shapeDef *box2d.ShapeDef, circle *box2d.Circle, capsule *box2d.Capsule,
	wedge *box2d.Polygon,
) {
	const (
		mixedShapeCount       = 3
		mixedBoxIndex         = 2
		circleMinRadius       = 0.25
		circleMaxRadius       = 0.75
		capsuleMinRadius      = 0.25
		capsuleMaxRadius      = 0.5
		capsuleMinLength      = 0.25
		capsuleMaxLength      = 1.0
		capsuleCenterFraction = 0.5
		boxMinHalfWidth       = 0.1
		boxMaxHalfWidth       = 0.5
		boxMinHalfHeight      = 0.5
		boxMaxHalfHeight      = 0.75
		boxMaxRadius          = 0.25
		wedgeMinRadius        = 0.1
		wedgeMaxRadius        = 0.25
	)

	mod := index % mixedShapeCount
	switch mod {
	case 0:
		circle.Radius = scenes.RandomFloatRange(circleMinRadius, circleMaxRadius)
		s.bodies[index].CreateCircleShape(shapeDef, circle)
	case 1:
		capsule.Radius = scenes.RandomFloatRange(capsuleMinRadius, capsuleMaxRadius)
		length := scenes.RandomFloatRange(capsuleMinLength, capsuleMaxLength)
		capsule.Center1 = box2d.Vec2{X: 0.0, Y: -capsuleCenterFraction * length}
		capsule.Center2 = box2d.Vec2{X: 0.0, Y: capsuleCenterFraction * length}
		s.bodies[index].CreateCapsuleShape(shapeDef, capsule)
	case mixedBoxIndex:
		width := scenes.RandomFloatRange(boxMinHalfWidth, boxMaxHalfWidth)
		height := scenes.RandomFloatRange(boxMinHalfHeight, boxMaxHalfHeight)
		box := box2d.MakeBox(width, height)
		value := scenes.RandomFloatRange(-1.0, 1.0)
		box.Radius = boxMaxRadius * max(0.0, value)
		s.bodies[index].CreatePolygonShape(shapeDef, &box)
	default:
		wedge.Radius = scenes.RandomFloatRange(wedgeMinRadius, wedgeMaxRadius)
		s.bodies[index].CreatePolygonShape(shapeDef, wedge)
	}
}

func CreateBenchmarkBarrel24(context *SampleContext) Sampler {
	const (
		cameraX           = 8.0
		cameraY           = 53.0
		baseZoom          = 25.0
		zoomScale         = 2.35
		groundHalfWidth   = 25.0
		wallHalfThickness = 1.2
		quarterTurn       = 0.5
		wallHeightFactor  = 2.0
		bodyRadius        = 0.5
		diameterFactor    = 2.0
		centerDivisor     = 2.0
		bodyHalfExtent    = 0.5
		rowsPerColumn     = 5
		bodyStartHeight   = 2.0
	)

	s := &BenchmarkBarrel24{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = baseZoom * zoomScale
	}

	groundSize := float32(groundHalfWidth)

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(groundSize, wallHalfThickness)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)

		bodyDef.Rotation = box2d.MakeRot(quarterTurn * box2d.Pi)
		bodyDef.Position = box2d.Pos{
			X: posScalar(groundSize),
			Y: posScalar(wallHeightFactor * groundSize),
		}
		groundID = s.worldID.CreateBody(&bodyDef)

		box = box2d.MakeBox(wallHeightFactor*groundSize, wallHalfThickness)
		groundID.CreatePolygonShape(&shapeDef, &box)

		bodyDef.Position = box2d.Pos{
			X: posScalar(-groundSize),
			Y: posScalar(wallHeightFactor * groundSize),
		}
		groundID = s.worldID.CreateBody(&bodyDef)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	num := 26
	rad := float32(bodyRadius)

	shift := rad * diameterFactor
	centerx := shift * float32(num) / centerDivisor
	centery := shift / centerDivisor

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.5

	cuboid := box2d.MakeSquare(bodyHalfExtent)

	numj := rowsPerColumn * num
	if isDebug {
		numj = 5
	}

	for i := range num {
		x := float32(i)*shift - centerx

		for j := range numj {
			y := float32(j)*shift + centery + bodyStartHeight

			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}

			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &cuboid)
		}
	}

	return s
}

func CreateBenchmarkCompounds(context *SampleContext) Sampler {
	const (
		cameraHeight = 50.0
		baseZoom     = 25.0
		zoomScale    = 2.2
	)

	s := &BenchmarkCompounds{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = baseZoom * zoomScale
		s.context.enableSleep = false
	}

	scenes.CreateCompounds(s.worldID)

	return s
}

func CreateBenchmarkTumbler(context *SampleContext) Sampler {
	const (
		cameraX = 1.5
		cameraY = 10.0
	)

	s := &BenchmarkTumbler{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = 15.0
	}

	scenes.CreateTumbler(s.worldID)

	return s
}

func CreateBenchmarkWasher(context *SampleContext) Sampler {
	const (
		cameraX = 1.5
		cameraY = 10.0
	)

	s := &BenchmarkWasher{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
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

	s.DrawScreenTextLinef("hits = %d", hitCount)
}

func CreateBenchmarkManyTumblers(context *SampleContext) Sampler {
	const (
		baseZoom            = 25.0
		zoomScale           = 3.4
		initialGridRows     = 19
		initialGridColumns  = 19
		initialAngularSpeed = 25.0
	)

	s := &BenchmarkManyTumblers{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		rowCount:     initialGridRows,
		columnCount:  initialGridColumns,
		angularSpeed: initialAngularSpeed,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.0, Y: -5.5}
		s.context.camera.zoom = baseZoom * zoomScale
		s.context.debugDraw.DrawJoints = false
	}

	bodyDef := box2d.DefaultBodyDef()
	s.groundID = s.worldID.CreateBody(&bodyDef)

	if isDebug {
		s.rowCount = 2
		s.columnCount = 2
	}

	s.CreateScene()

	return s
}

func (s *BenchmarkManyTumblers) CreateTumbler(position box2d.Pos, index int) {
	const (
		wallHalfThickness = 0.25
		wallHalfLength    = 2.0
	)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.KinematicBody
	bodyDef.Position = position
	bodyDef.AngularVelocity = (box2d.Pi / degreesPerHalfTurn) * s.angularSpeed
	bodyID := s.worldID.CreateBody(&bodyDef)
	s.tumblerIDs[index] = bodyID

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 50.0

	polygon := box2d.MakeOffsetBox(
		wallHalfThickness,
		wallHalfLength,
		box2d.Vec2{X: wallHalfLength, Y: 0.0},
		box2d.RotIdentity,
	)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
	polygon = box2d.MakeOffsetBox(
		wallHalfThickness,
		wallHalfLength,
		box2d.Vec2{X: -wallHalfLength, Y: 0.0},
		box2d.RotIdentity,
	)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
	polygon = box2d.MakeOffsetBox(
		wallHalfLength,
		wallHalfThickness,
		box2d.Vec2{X: 0.0, Y: wallHalfLength},
		box2d.RotIdentity,
	)
	bodyID.CreatePolygonShape(&shapeDef, &polygon)
	polygon = box2d.MakeOffsetBox(
		wallHalfLength,
		wallHalfThickness,
		box2d.Vec2{X: 0.0, Y: -wallHalfLength},
		box2d.RotIdentity,
	)
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
	const (
		maxGridSize     = 32
		maxAngularSpeed = 100.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	changed := false
	changed = changed || imgui.SliderInt("Row Count", &s.rowCount, 1, maxGridSize)
	changed = changed || imgui.SliderInt("Column Count", &s.columnCount, 1, maxGridSize)

	if changed {
		s.CreateScene()
	}

	if imgui.SliderFloatV("Speed", &s.angularSpeed, 0.0, maxAngularSpeed, "%.f", 0) {
		for i := range s.tumblerCount {
			s.tumblerIDs[i].SetAngularVelocity((box2d.Pi / degreesPerHalfTurn) * s.angularSpeed)
			s.tumblerIDs[i].SetAwake(true)
		}
	}

	imgui.PopItemWidth()

	return true
}

func (s *BenchmarkManyTumblers) Step() {
	const (
		spawnStepMask     = 0x7
		capsuleHalfLength = 0.1
		capsuleRadius     = 0.075
	)

	s.Sample.Step()

	if s.bodyIndex < s.bodyCount && (s.stepCount&spawnStepMask) == 0 {
		shapeDef := box2d.DefaultShapeDef()

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
			Radius:  capsuleRadius,
		}

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

func CreateBenchmarkLargePyramid(context *SampleContext) Sampler {
	const (
		cameraHeight = 50.0
		baseZoom     = 25.0
		zoomScale    = 2.2
	)

	s := &BenchmarkLargePyramid{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = baseZoom * zoomScale
		s.context.enableSleep = false
	}

	scenes.CreateLargePyramid(s.worldID)

	return s
}

func CreateBenchmarkManyPyramids(context *SampleContext) Sampler {
	const (
		cameraX = 23.0
		cameraY = 72.5
	)

	s := &BenchmarkManyPyramids{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = 165.0
		s.context.enableSleep = false
	}

	scenes.CreateManyPyramids(s.worldID)

	return s
}

func CreateBenchmarkCreateDestroy(context *SampleContext) Sampler {
	const (
		cameraHeight      = 50.0
		baseZoom          = 25.0
		zoomScale         = 2.2
		groundHalfWidth   = 100.0
		initialBaseCount  = 100
		initialIterations = 10
	)

	s := &BenchmarkCreateDestroy{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		baseCount: initialBaseCount, iterations: initialIterations,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = baseZoom * zoomScale
	}

	groundSize := float32(groundHalfWidth)

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	box := box2d.MakeBox(groundSize, 1.0)
	shapeDef := box2d.DefaultShapeDef()
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.createTime = 0.0
	s.destroyTime = 0.0

	if isDebug {
		s.baseCount = 40
		s.iterations = 1
	}

	s.bodyCount = 0

	return s
}

func (s *BenchmarkCreateDestroy) CreateScene() {
	const (
		bodyRadius         = 0.5
		diameterFactor     = 2.0
		centerDivisor      = 2.0
		bodyHalfExtent     = 0.5
		simulationHertz    = 60.0
		simulationSubsteps = 4
	)

	ticks := time.Now()

	for i := range createDestroyMaxBodyCount {
		if s.bodies[i] != (box2d.Body{}) {
			s.bodies[i].Destroy()
			s.bodies[i] = box2d.Body{}
		}
	}

	s.destroyTime += millisecondsAndReset(&ticks)

	count := s.baseCount
	rad := float32(bodyRadius)
	shift := rad * diameterFactor
	centerx := shift * float32(count) / centerDivisor
	centery := shift/centerDivisor + 1.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.5

	h := float32(bodyHalfExtent)
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

	s.worldID.Step(1.0/simulationHertz, simulationSubsteps)
}

func (s *BenchmarkCreateDestroy) Step() {
	s.createTime = 0.0
	s.destroyTime = 0.0

	for range s.iterations {
		s.CreateScene()
	}

	s.DrawScreenTextLinef("total: create = %.6g ms, destroy = %.6g ms", s.createTime, s.destroyTime)

	createPerBody := microsecondsPerMillisecond * s.createTime / float32(
		s.iterations,
	) / float32(
		s.bodyCount,
	)
	destroyPerBody := microsecondsPerMillisecond * s.destroyTime / float32(
		s.iterations,
	) / float32(
		s.bodyCount,
	)
	s.DrawScreenTextLinef(
		"body: create = %.6g us, destroy = %.6g us",
		createPerBody,
		destroyPerBody,
	)

	s.Sample.Step()
}

func CreateBenchmarkSleep(context *SampleContext) Sampler {
	const (
		cameraHeight = 20.0
		baseZoom     = 25.0
		zoomScale    = 6.0
	)

	s := &BenchmarkSleep{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = baseZoom * zoomScale
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

func CreateBenchmarkJointGrid(context *SampleContext) Sampler {
	const (
		cameraX   = 60.0
		baseZoom  = 25.0
		zoomScale = 2.5
	)

	s := &BenchmarkJointGrid{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: -57.0}
		s.context.camera.zoom = baseZoom * zoomScale
		s.context.enableSleep = false
	}

	scenes.CreateJointGrid(s.worldID)

	return s
}

func CreateBenchmarkSmash(context *SampleContext) Sampler {
	const (
		cameraX   = 60.0
		cameraY   = 6.0
		baseZoom  = 25.0
		zoomScale = 1.6
	)

	s := &BenchmarkSmash{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = baseZoom * zoomScale
	}

	scenes.CreateSmash(s.worldID)

	return s
}

func CreateBenchmarkLargeCompounds(context *SampleContext) Sampler {
	const (
		cameraX            = 18.0
		cameraY            = 115.0
		baseZoom           = 25.0
		zoomScale          = 5.5
		gridCellHalfExtent = 0.5
		bodyStartHeight    = 100.0
	)

	s := &BenchmarkLargeCompounds{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = baseZoom * zoomScale
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
				square := box2d.MakeOffsetBox(
					gridCellHalfExtent*grid,
					gridCellHalfExtent*grid,
					box2d.Vec2{X: x, Y: y},
					box2d.RotIdentity,
				)
				groundID.CreatePolygonShape(&shapeDef, &square)
			}
		}

		for i := range height {
			y := grid * float32(i)
			for j := i; j < width; j++ {
				x := -grid * float32(j)
				square := box2d.MakeOffsetBox(
					gridCellHalfExtent*grid,
					gridCellHalfExtent*grid,
					box2d.Vec2{X: x, Y: y},
					box2d.RotIdentity,
				)
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
			ybody := (bodyStartHeight + float32(m*span)) * grid

			for n := range count {
				xbody := -0.5*grid*float32(count*span) + float32(n*span)*grid
				bodyDef.Position = box2d.Pos{X: posScalar(xbody), Y: posScalar(ybody)}
				bodyID := s.worldID.CreateBody(&bodyDef)

				for i := range span {
					y := float32(i) * grid
					for j := range span {
						x := float32(j) * grid
						square := box2d.MakeOffsetBox(
							gridCellHalfExtent*grid,
							gridCellHalfExtent*grid,
							box2d.Vec2{X: x, Y: y},
							box2d.RotIdentity,
						)
						bodyID.CreatePolygonShape(&shapeDef, &square)
					}
				}

				bodyID.UpdateMassFromShapes()
			}
		}
	}

	return s
}

func CreateBenchmarkKinematic(context *SampleContext) Sampler {
	const gridCellHalfExtent = 0.5

	s := &BenchmarkKinematic{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
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
			square := box2d.MakeOffsetBox(
				gridCellHalfExtent*grid,
				gridCellHalfExtent*grid,
				box2d.Vec2{X: x, Y: y},
				box2d.RotIdentity,
			)
			bodyID.CreatePolygonShape(&shapeDef, &square)
		}
	}

	bodyID.UpdateMassFromShapes()

	return s
}

func CreateBenchmarkCast(context *SampleContext) Sampler {
	const (
		cameraCenter       = 500.0
		baseZoom           = 25.0
		zoomScale          = 21.0
		initialAspectRatio = 5.0
		initialGridSpacing = 1.0
		initialFill        = 0.1
		initialRows        = 1000
		initialColumns     = 1000
		initialMinTime     = 1e6
		initialRadius      = 0.1
	)

	s := &BenchmarkCast{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		queryType:   queryCircleCast,
		ratio:       initialAspectRatio,
		grid:        initialGridSpacing,
		fill:        initialFill,
		rowCount:    initialRows,
		columnCount: initialColumns,
		minTime:     initialMinTime,
		radius:      initialRadius,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenter, Y: cameraCenter}
		s.context.camera.zoom = baseZoom * zoomScale
	}

	if isDebug {
		s.rowCount = 100
		s.columnCount = 100
	}

	s.drawIndex = 0
	s.topDown = false
	s.buildTime = 0.0

	scenes.SetRandomSeed(castRandomSeed)

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
	const (
		minShapeHalfWidth = 0.05
		maxShapeHalfWidth = 0.25
		maxShapeCategory  = 2
	)

	scenes.SetRandomSeed(castRandomSeed)
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
				halfWidth := scenes.RandomFloatRange(minShapeHalfWidth, maxShapeHalfWidth)

				var box box2d.Polygon
				if scenes.RandomFloat() > 0.0 {
					box = box2d.MakeBox(ratio*halfWidth, halfWidth)
				} else {
					box = box2d.MakeBox(halfWidth, ratio*halfWidth)
				}

				category := scenes.RandomIntRange(0, maxShapeCategory)

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
	const (
		maxGridCount   = 1000
		minGridSpacing = 0.5
		maxGridSpacing = 2.0
		minAspectRatio = 1.0
		maxAspectRatio = 10.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

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

	if imgui.SliderIntV("rows", &s.rowCount, 0, maxGridCount, "%d", 0) {
		changed = true
	}

	if imgui.SliderIntV("columns", &s.columnCount, 0, maxGridCount, "%d", 0) {
		changed = true
	}

	if imgui.SliderFloatV("fill", &s.fill, 0.0, 1.0, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("grid", &s.grid, minGridSpacing, maxGridSpacing, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("ratio", &s.ratio, minAspectRatio, maxAspectRatio, "%.2f", 0) {
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
		hitCount, nodeVisits, leafVisits, ms = s.runRayQueries(filter)
	case queryCircleCast:
		hitCount, nodeVisits, leafVisits, ms = s.runCircleQueries(filter)
	case queryOverlap:
		hitCount, nodeVisits, leafVisits, ms = s.runOverlapQueries(filter)
	}

	s.DrawScreenTextLinef("build time ms = %.6g", s.buildTime)
	s.DrawScreenTextLinef(
		"hit count = %d, node visits = %d, leaf visits = %d",
		hitCount,
		nodeVisits,
		leafVisits,
	)
	s.DrawScreenTextLinef("total ms = %.3f", ms)
	s.DrawScreenTextLinef("min total ms = %.3f", s.minTime)

	aveRayCost := microsecondsPerMillisecond * s.minTime / float32(sampleCount)
	s.DrawScreenTextLinef("average us = %.2f", aveRayCost)
}

func (s *BenchmarkCast) runRayQueries(filter box2d.QueryFilter) (int, int, int, float32) {
	sampleCount := len(s.origins)
	hitCount, nodeVisits, leafVisits := 0, 0, 0

	var ms float32

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
	s.context.draw.DrawPoint(p1, castPointSize, box2d.ColorGreen)
	s.context.draw.DrawPoint(p2, castPointSize, box2d.ColorRed)

	if drawResult.Hit {
		s.context.draw.DrawPoint(drawResult.Point, castPointSize, box2d.ColorWhite)
	}

	return hitCount, nodeVisits, leafVisits, ms
}

func (s *BenchmarkCast) runCircleQueries(filter box2d.QueryFilter) (int, int, int, float32) {
	sampleCount := len(s.origins)
	hitCount, nodeVisits, leafVisits := 0, 0, 0

	var ms float32

	ticks := time.Now()

	var drawResult castResult

	for i := range sampleCount {
		proxy := box2d.MakeProxy([]box2d.Vec2{box2d.Vec2Zero}, s.radius)
		translation := s.translations[i]

		var result castResult

		traversalResult := s.worldID.CastShape(
			s.origins[i],
			&proxy,
			translation,
			filter,
			func(shape box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
				result.point = point
				result.fraction = fraction
				result.hit = true

				return fraction
			},
		)

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
	s.context.draw.DrawPoint(p1, castPointSize, box2d.ColorGreen)
	s.context.draw.DrawPoint(p2, castPointSize, box2d.ColorRed)

	if drawResult.hit {
		t := s.origins[s.drawIndex].Add(s.translations[s.drawIndex].Scale(drawResult.fraction))
		s.context.draw.DrawCircle(t, s.radius, box2d.ColorWhite)
		s.context.draw.DrawPoint(drawResult.point, castPointSize, box2d.ColorWhite)
	}

	return hitCount, nodeVisits, leafVisits, ms
}

func (s *BenchmarkCast) runOverlapQueries(filter box2d.QueryFilter) (int, int, int, float32) {
	sampleCount := len(s.origins)
	hitCount, nodeVisits, leafVisits := 0, 0, 0

	var ms float32

	ticks := time.Now()

	var drawResult overlapResult

	extent := box2d.Vec2{X: s.radius, Y: s.radius}

	var result overlapResult

	for i := range sampleCount {
		aabb := box2d.AABB{LowerBound: extent.Neg(), UpperBound: extent}

		result.count = 0
		traversalResult := s.worldID.OverlapAABB(
			s.origins[i],
			aabb,
			filter,
			func(shape box2d.Shape) bool {
				if result.count < len(result.points) {
					result.points[result.count] = shape.AABB().Center()
					result.count += 1
				}

				return true
			},
		)

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
		LowerBound: box2d.Vec2{
			X: box2d.RoundDownFloat(float64(origin.X) - float64(extent.X)),
			Y: box2d.RoundDownFloat(float64(origin.Y) - float64(extent.Y)),
		},
		UpperBound: box2d.Vec2{
			X: box2d.RoundUpFloat(float64(origin.X) + float64(extent.X)),
			Y: box2d.RoundUpFloat(float64(origin.Y) + float64(extent.Y)),
		},
	}

	s.context.draw.DrawBounds(aabb, box2d.ColorWhite)

	for i := range drawResult.count {
		s.context.draw.DrawPoint(drawResult.points[i].Pos(), castPointSize, box2d.ColorHotPink)
	}

	return hitCount, nodeVisits, leafVisits, ms
}

func CreateBenchmarkSpinner(context *SampleContext) Sampler {
	const cameraHeight = 32.0

	s := &BenchmarkSpinner{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 42.0
	}

	scenes.CreateSpinner(s.worldID)

	return s
}

func CreateBenchmarkRain(context *SampleContext) Sampler {
	const cameraHeight = 110.0

	s := &BenchmarkRain{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
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

func CreateBenchmarkShapeDistance(context *SampleContext) Sampler {
	const (
		polygonVertexCount     = 8
		fullTurnFactor         = 2.0
		polygonRadius          = 0.5
		polygonBRounding       = 0.1
		polygonAPositionExtent = 0.1
		polygonBMinPosition    = 0.25
		polygonBMaxPosition    = 2.0
	)

	var (
		polygonA box2d.Polygon
		polygonB box2d.Polygon
	)
	{
		var points [polygonVertexCount]box2d.Vec2

		q := box2d.MakeRot(fullTurnFactor * box2d.Pi / polygonVertexCount)
		p := box2d.Vec2{X: polygonRadius, Y: 0.0}

		points[0] = p
		for i := 1; i < polygonVertexCount; i++ {
			points[i] = q.Rotate(points[i-1])
		}

		hull := box2d.ComputeHull(points[:])
		polygonA = box2d.MakePolygon(&hull, 0.0)
	}

	{
		var points [polygonVertexCount]box2d.Vec2

		q := box2d.MakeRot(fullTurnFactor * box2d.Pi / polygonVertexCount)
		p := box2d.Vec2{X: polygonRadius, Y: 0.0}

		points[0] = p
		for i := 1; i < polygonVertexCount; i++ {
			points[i] = q.Rotate(points[i-1])
		}

		hull := box2d.ComputeHull(points[:])
		polygonB = box2d.MakePolygon(&hull, polygonBRounding)
	}

	s := &BenchmarkShapeDistance{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		polygonA: polygonA, polygonB: polygonB,
		minMilliseconds: math.MaxFloat32,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 3.0
	}

	s.transformAs = make([]box2d.WorldTransform, shapeDistanceCount)
	s.transformBs = make([]box2d.WorldTransform, shapeDistanceCount)
	s.outputs = make([]box2d.DistanceOutput, shapeDistanceCount)

	scenes.SetRandomSeed(benchmarkRandomSeed)

	for i := range shapeDistanceCount {
		s.transformAs[i] = box2d.WorldTransform{
			P: scenes.RandomPos(-polygonAPositionExtent, polygonAPositionExtent),
			Q: scenes.RandomRot(),
		}
		s.transformBs[i] = box2d.WorldTransform{
			P: scenes.RandomPos(polygonBMinPosition, polygonBMaxPosition),
			Q: scenes.RandomRot(),
		}
	}

	s.drawIndex = 0

	return s
}

func (s *BenchmarkShapeDistance) DrawControls() bool {
	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())
	imgui.SliderInt("draw index", &s.drawIndex, 0, shapeDistanceCount-1)
	imgui.PopItemWidth()

	return true
}

func (s *BenchmarkShapeDistance) Step() {
	const (
		distancePointSize   = 10.0
		normalDisplayLength = 0.5
	)

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

		s.DrawScreenTextLinef("count = %d", shapeDistanceCount)
		s.DrawScreenTextLinef(
			"min ms = %.6g, ave us = %.6g",
			s.minMilliseconds,
			microsecondsPerMillisecond*s.minMilliseconds/float32(shapeDistanceCount),
		)
		s.DrawScreenTextLinef(
			"average iterations = %.6g",
			float32(totalIterations)/float32(shapeDistanceCount),
		)
	}

	xfA := s.transformAs[s.drawIndex]
	xfB := s.transformBs[s.drawIndex]
	output := s.outputs[s.drawIndex]

	worldPointA := xfA.Apply(output.PointA)
	worldPointB := xfA.Apply(output.PointB)
	worldNormal := xfA.Q.Rotate(output.Normal)

	s.context.draw.DrawSolidPolygon(
		xfA,
		s.polygonA.Vertices[:s.polygonA.Count],
		s.polygonA.Radius,
		box2d.ColorBox2DGreen,
	)
	s.context.draw.DrawSolidPolygon(
		xfB,
		s.polygonB.Vertices[:s.polygonB.Count],
		s.polygonB.Radius,
		box2d.ColorBox2DBlue,
	)
	s.context.draw.DrawLine(worldPointA, worldPointB, box2d.ColorDimGray)
	s.context.draw.DrawPoint(worldPointA, distancePointSize, box2d.ColorWhite)
	s.context.draw.DrawPoint(worldPointB, distancePointSize, box2d.ColorWhite)
	s.context.draw.DrawLine(
		worldPointA,
		worldPointA.Add(worldNormal.Scale(normalDisplayLength)),
		box2d.ColorYellow,
	)
	s.DrawScreenTextLinef("distance = %.6g", output.Distance)

	s.Sample.Step()
}

func CreateBenchmarkSensor(context *SampleContext) Sampler {
	const (
		cameraHeight         = 105.0
		activeGridSize       = 3.0
		activeSensorHalfSize = 0.45
		gridCenterFraction   = 0.5
		passiveStartHeight   = 10.0
		passiveHalfExtent    = 0.5
		passiveCornerRadius  = 0.1
	)

	s := &BenchmarkSensor{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		activeSensor: shapeUserData{row: 0, active: true},
		filterRow:    sensorRowCount >> 1,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 125.0
	}

	s.worldID.SetCustomFilterCallback(s.Filter)

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	{
		gridSize := float32(activeGridSize)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		shapeDef.UserData = &s.activeSensor

		y := float32(0.0)

		x := -40.0 * gridSize
		for range 81 {
			box := box2d.MakeOffsetBox(
				activeSensorHalfSize*gridSize,
				activeSensorHalfSize*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.RotIdentity,
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			x += gridSize
		}
	}

	scenes.SetRandomSeed(benchmarkRandomSeed)

	shift := float32(sensorGridSpacing)
	xCenter := gridCenterFraction * shift * sensorColumnCount

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	shapeDef.EnableSensorEvents = true

	yStart := float32(passiveStartHeight)

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
			box := box2d.MakeOffsetRoundedBox(
				passiveHalfExtent,
				passiveHalfExtent,
				box2d.Vec2{X: x, Y: y},
				box2d.RotIdentity,
				passiveCornerRadius,
			)
			groundID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	s.maxBeginCount = 0
	s.maxEndCount = 0
	s.lastStepCount = 0

	return s
}

func (s *BenchmarkSensor) CreateRow(y float32) {
	const (
		gridCenterFraction = 0.5
		bodyRadius         = 0.5
	)

	shift := float32(sensorGridSpacing)
	xCenter := gridCenterFraction * shift * sensorColumnCount

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.0
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -5.0}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: bodyRadius}

	for i := range sensorColumnCount {
		yOffset := scenes.RandomFloatRange(-1.0, 1.0)
		bodyDef.Position = box2d.Pos{
			X: posScalar(shift*float32(i) - xCenter),
			Y: posScalar(y + yOffset),
		}
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

	beginCount := s.collectBeginEvents(zombies)

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
		s.CreateRow(sensorGridStartHeight + sensorRowCount*sensorGridSpacing)
	}

	s.lastStepCount = s.stepCount

	s.maxBeginCount = max(beginCount, s.maxBeginCount)
	s.maxEndCount = max(endCount, s.maxEndCount)
	s.DrawScreenTextLinef("max begin touch events = %d", s.maxBeginCount)
	s.DrawScreenTextLinef("max end touch events = %d", s.maxEndCount)
}

func (s *BenchmarkSensor) Filter(idA box2d.Shape, idB box2d.Shape) bool {
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

func (s *BenchmarkSensor) collectBeginEvents(zombies map[box2d.Body]struct{}) int {
	beginCount := 0
	for event := range s.worldID.SensorBeginEvents() {
		beginCount++

		userData, ok := event.SensorShape.UserData().(*shapeUserData)
		if !ok || userData == nil {
			panic("benchmark sensor shapes must carry non-nil *shapeUserData")
		}

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

	return beginCount
}

func CreateBenchmarkCapacity(context *SampleContext) Sampler {
	const (
		cameraHeight     = 150.0
		groundHalfWidth  = 800.0
		groundHalfHeight = 5.0
		bodyHalfExtent   = 0.5
	)

	s := &BenchmarkCapacity{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		square: box2d.MakeSquare(bodyHalfExtent),
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 200.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position.Y = -groundHalfHeight
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(groundHalfWidth, groundHalfHeight)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.done = false
	s.reachCount = 0

	return s
}

func (s *BenchmarkCapacity) Step() {
	const (
		millisecondLimit = float32(20.0)
		overLimitSteps   = 60
		spawnStepMask    = 0x1F
	)

	s.Sample.Step()

	profile := s.worldID.Profile()
	if profile.Step > millisecondLimit {
		s.reachCount += 1
		if s.reachCount > overLimitSteps {
			s.done = true
		}
	} else {
		s.reachCount = 0
	}

	if s.done {
		return
	}

	if (s.stepCount & spawnStepMask) != spawnStepMask {
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

func CreateBenchmarkJunkyard(context *SampleContext) Sampler {
	const (
		cameraX = 8.0
		cameraY = 25.0
	)

	s := &BenchmarkJunkyard{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
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

func CreateBenchmarkQueries(context *SampleContext) Sampler {
	const initialMinQueryTime = 1e6

	s := &BenchmarkQueries{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		minTime: initialMinQueryTime,
	}
	s.init(context, s, true)

	scenes.CreateQueries(s.worldID)

	if !s.context.restart {
		const (
			cameraCenterFraction = 0.5
			cameraZoomFraction   = 0.55
		)

		extent := scenes.GetQueryBenchmarkExtent()
		s.context.camera.center = box2d.Pos{
			X: posScalar(cameraCenterFraction * extent),
			Y: posScalar(cameraCenterFraction * extent),
		}
		s.context.camera.zoom = cameraZoomFraction * extent
	}

	s.drawIndex = 0
	s.lastTime = 0.0
	s.hitCount = 0
	s.nodeVisits = 0
	s.leafVisits = 0

	return s
}

func (s *BenchmarkQueries) DrawControls() bool {
	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())
	imgui.SliderIntV("draw", &s.drawIndex, 0, int32(scenes.GetQueryBenchmarkCount()-1), "%d", 0)
	imgui.PopItemWidth()

	return false
}

func (s *BenchmarkQueries) Step() {
	const (
		queryPointSize  = 5.0
		queryHalfExtent = 5.0
	)

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
	s.context.draw.DrawPoint(origin, queryPointSize, box2d.ColorGreen)
	s.context.draw.DrawPoint(end, queryPointSize, box2d.ColorRed)

	box := box2d.AABB{
		LowerBound: box2d.Vec2{
			X: box2d.RoundDownFloat(float64(origin.X) - queryHalfExtent),
			Y: box2d.RoundDownFloat(float64(origin.Y) - queryHalfExtent),
		},
		UpperBound: box2d.Vec2{
			X: box2d.RoundUpFloat(float64(origin.X) + queryHalfExtent),
			Y: box2d.RoundUpFloat(float64(origin.Y) + queryHalfExtent),
		},
	}
	s.context.draw.DrawBounds(box, box2d.ColorWhite)

	s.DrawScreenTextLinef(
		"%d queries of each kind and filter, %d hits",
		scenes.GetQueryBenchmarkCount(),
		s.hitCount,
	)
	s.DrawScreenTextLinef("node visits = %d, leaf visits = %d", s.nodeVisits, s.leafVisits)
	s.DrawScreenTextLinef("queries ms = %.3f, min = %.3f", s.lastTime, s.minTime)
}

func millisecondsSince(start time.Time) float32 {
	return float32(time.Since(start).Seconds() * millisecondsPerSecond)
}

func millisecondsAndReset(start *time.Time) float32 {
	now := time.Now()
	ms := float32(now.Sub(*start).Seconds() * millisecondsPerSecond)
	*start = now

	return ms
}

func registerBenchmarkSamples() {
	RegisterSample("Benchmark", "Barrel", CreateBenchmarkBarrel)
	RegisterSample("Benchmark", "Barrel 2.4", CreateBenchmarkBarrel24)
	RegisterSample("Benchmark", "Compounds", CreateBenchmarkCompounds)
	RegisterSample("Benchmark", "Tumbler", CreateBenchmarkTumbler)
	RegisterSample("Benchmark", "Washer", CreateBenchmarkWasher)
	RegisterSample("Benchmark", "Many Tumblers", CreateBenchmarkManyTumblers)
	RegisterSample("Benchmark", "Large Pyramid", CreateBenchmarkLargePyramid)
	RegisterSampleWithCapacity(
		"Benchmark",
		"Many Pyramids",
		CreateBenchmarkManyPyramids,
		scenes.GetManyPyramidsCapacity,
	)
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
