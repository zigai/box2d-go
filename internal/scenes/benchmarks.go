// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"github.com/zigai/box2d"
	"math"
)

const (
	benchmarkDebug              = false
	rainRowCount                = 5
	rainColumnCount             = 40
	rainGroupSize               = 5
	spinnerPointCount           = 360
	sleepPyramidCount           = 10
	queryCount                  = 1000
	treeCastRayCount            = 1000
	treeCastLengthCount         = 3
	treeCastBoxExtent   float32 = 0.1
	tilePeriod          float32 = 40
	tileCycleCount              = 600
	tileGridSize        float32 = 1
	tileColumnsPerBody          = 10
	tileQueryCount              = 500
	tileRayLength       float32 = 25
)

type group struct {
	humans [rainGroupSize]Human
}

type rainData struct {
	groups      [rainRowCount * rainColumnCount]group
	gridSize    float32
	gridCount   int
	columnCount int
	columnIndex int
}

type spinnerData struct {
	spinnerID box2d.Joint
}

type junkyardData struct {
	pusherID box2d.Body
}

type sleepData struct {
	bodyIDA [sleepPyramidCount]box2d.Body
	bodyIDB [sleepPyramidCount]box2d.Body
}

type queryBenchmark struct {
	origins      [1000]box2d.Pos
	translations [1000]box2d.Vec2
	paddleID     box2d.Body
	stats        box2d.TreeStats
	extent       float32
	paddleMinX   float32
	paddleMaxX   float32
}

type treeCastBenchmark struct {
	tree          box2d.DynamicTree
	tightBoxes    []box2d.AABB
	proxyCount    int
	proxyCapacity int
	origins       [1000]box2d.Vec2
	translations  [treeCastLengthCount][1000]box2d.Vec2
	stats         box2d.TreeStats
	created       bool
}

type treeCastContext struct {
	tightBoxes []box2d.AABB
	fraction   float32
}

type tileWorldBenchmark struct {
	groundIDs   []box2d.Body
	groundCount int
	groundHead  int
	bodySpan    float32
	xLeft       float32
	xNext       float32
	stats       box2d.TreeStats
}

var gRainData rainData
var gSpinnerData spinnerData
var gJunkyardData junkyardData
var gSleepData sleepData
var gQueryBenchmark queryBenchmark
var gQueryRandomState uint32
var gTreeCast treeCastBenchmark
var gTileWorld tileWorldBenchmark

func CreateJointGrid(worldID box2d.World) {
	worldID.EnableSleeping(false)
	var N int = 100
	var bodies []box2d.Body = make([]box2d.Body, N*N)
	var index int = 0
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Filter.CategoryBits = 2
	shapeDef.Filter.MaskBits = uint64(^uint32(2))
	var circle box2d.Circle = box2d.Circle{Center: box2d.Vec2{X: float32(0.0), Y: float32(0.0)}, Radius: float32(0.4)}
	var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
	jointDef.Base.DrawScale = float32(0.4)
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	for k := 0; k < N; k++ {
		for i := 0; i < N; i++ {
			var fk float32 = float32(k)
			var fi float32 = float32(i)
			if k >= N/2-3 && k <= N/2+3 && i == 0 {
				bodyDef.Type = box2d.StaticBody
			} else {
				bodyDef.Type = box2d.DynamicBody
			}
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: fk, Y: -fi})
			var body box2d.Body = worldID.CreateBody(&bodyDef)
			body.CreateCircleShape(&shapeDef, &circle)
			if i > 0 {
				jointDef.Base.BodyIDA = bodies[index-1]
				jointDef.Base.BodyIDB = body
				jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.0), Y: -float32(0.5)}
				jointDef.Base.LocalFrameB.P = box2d.Vec2{X: float32(0.0), Y: float32(0.5)}
				worldID.CreateRevoluteJoint(&jointDef)
			}
			if k > 0 {
				jointDef.Base.BodyIDA = bodies[index-N]
				jointDef.Base.BodyIDB = body
				jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.5), Y: float32(0.0)}
				jointDef.Base.LocalFrameB.P = box2d.Vec2{X: -float32(0.5), Y: float32(0.0)}
				worldID.CreateRevoluteJoint(&jointDef)
			}
			bodies[index] = body
			index++
		}
	}
}

func CreateLargePyramid(worldID box2d.World) {
	worldID.EnableSleeping(false)
	var baseCount int = 100
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: float32(0.0), Y: -float32(1.0)})
		var groundID box2d.Body = worldID.CreateBody(&bodyDef)
		var box box2d.Polygon = box2d.MakeBox(float32(100.0), float32(1.0))
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	var a float32 = float32(0.5)
	var box box2d.Polygon = box2d.MakeSquare(a)
	var shift float32 = float32(1.0) * a
	for i := 0; i < baseCount; i++ {
		var y float32 = (float32(float32(2.0)*float32(i)) + float32(1.0)) * shift
		for j := i; j < baseCount; j++ {
			var x float32 = float32((float32(i)+float32(1.0))*shift) + float32(float32(2.0)*(float32(j)-float32(i))*shift) - float32(a*float32(baseCount))
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
}

func createSmallPyramid(worldID box2d.World, baseCount int, extent, centerX, baseY float32) {
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var box box2d.Polygon = box2d.MakeSquare(extent)
	for i := 0; i < baseCount; i++ {
		var y float32 = float32((float32(float32(2.0)*float32(i))+float32(1.0))*extent) + baseY
		for j := i; j < baseCount; j++ {
			var x float32 = float32((float32(i)+float32(1.0))*extent) + float32(float32(2.0)*(float32(j)-float32(i))*extent) + centerX - float32(0.5)
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
}

func CreateManyPyramids(worldID box2d.World) {
	worldID.EnableSleeping(false)
	var baseCount int = 10
	var extent float32 = float32(0.5)
	var rowCount int = 20
	var columnCount int = 20
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	var groundID box2d.Body = worldID.CreateBody(&bodyDef)
	var groundDeltaY float32 = float32(2.0) * extent * (float32(baseCount) + float32(1.0))
	var groundWidth float32 = float32(2.0) * extent * float32(columnCount) * (float32(baseCount) + float32(1.0))
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var groundY float32 = float32(0.0)
	for i := 0; i < rowCount; i++ {
		var segment box2d.Segment = box2d.Segment{Point1: box2d.Vec2{X: -float32(0.5) * groundWidth, Y: groundY}, Point2: box2d.Vec2{X: float32(0.5) * groundWidth, Y: groundY}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
		groundY += groundDeltaY
	}
	var baseWidth float32 = float32(2.0) * extent * float32(baseCount)
	var baseY float32 = float32(0.0)
	for i := 0; i < rowCount; i++ {
		for j := 0; j < columnCount; j++ {
			var centerX float32 = float32(-float32(0.5)*groundWidth) + float32(float32(j)*(baseWidth+float32(float32(2.0)*extent))) + float32(float32(2.0)*extent)
			createSmallPyramid(worldID, baseCount, extent, centerX, baseY)
		}
		baseY += groundDeltaY
	}
}

func GetManyPyramidsCapacity() box2d.Capacity {
	var c box2d.Capacity = box2d.Capacity{StaticShapeCount: 20, StaticBodyCount: 1, DynamicShapeCount: 22000, DynamicBodyCount: 22000, ContactCount: 58000}
	return c
}

func CreateRain(worldID box2d.World) {
	gRainData = rainData{}
	gRainData.gridSize = float32(0.5)
	gRainData.gridCount = 500
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		var groundID box2d.Body = worldID.CreateBody(&bodyDef)
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		var y float32 = float32(0.0)
		var width float32 = gRainData.gridSize
		var height float32 = gRainData.gridSize
		for i := 0; i < rainRowCount; i++ {
			var x float32 = -float32(0.5) * float32(gRainData.gridCount) * gRainData.gridSize
			for j := 0; j <= gRainData.gridCount; j++ {
				var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.5)*width, float32(0.5)*height, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
				groundID.CreatePolygonShape(&shapeDef, &box)
				x += gRainData.gridSize
			}
			y += float32(45.0)
		}
	}
	gRainData.columnCount = 0
	gRainData.columnIndex = 0
}

func CreateGroup(worldID box2d.World, rowIndex, columnIndex int) {
	if !(rowIndex < rainRowCount && columnIndex < rainColumnCount) {
		panic(&box2d.AssertError{Condition: "rowIndex < RAIN_ROW_COUNT && columnIndex < RAIN_COLUMN_COUNT"})
	}
	var groupIndex int = rowIndex*rainColumnCount + columnIndex
	var span float32 = float32(gRainData.gridCount) * gRainData.gridSize
	var groupDistance float32 = float32(1.0) * span / float32(rainColumnCount)
	var position box2d.Pos
	position.X = box2d.ToPos(box2d.Vec2{X: float32(-float32(0.5)*span) + float32(groupDistance*(float32(columnIndex)+float32(0.5)))}).X
	position.Y = box2d.ToPos(box2d.Vec2{Y: float32(40.0) + float32(float32(45.0)*float32(rowIndex))}).Y
	var scale float32 = float32(1.0)
	var jointFriction float32 = float32(0.05)
	var jointHertz float32 = float32(5.0)
	var jointDamping float32 = float32(0.5)
	for i := 0; i < rainGroupSize; i++ {
		var human *Human = &gRainData.groups[groupIndex].humans[i]
		CreateHuman(human, worldID, position, scale, jointFriction, jointHertz, jointDamping, i+1, nil, false)
		position.X += 0.5
	}
}

func DestroyGroup(rowIndex, columnIndex int) {
	if !(rowIndex < rainRowCount && columnIndex < rainColumnCount) {
		panic(&box2d.AssertError{Condition: "rowIndex < RAIN_ROW_COUNT && columnIndex < RAIN_COLUMN_COUNT"})
	}
	var groupIndex int = rowIndex*rainColumnCount + columnIndex
	for i := 0; i < rainGroupSize; i++ {
		DestroyHuman(&gRainData.groups[groupIndex].humans[i])
	}
}

func StepRain(worldID box2d.World, stepCount int) float32 {
	var delay int = 0x7
	if (stepCount & delay) == 0 {
		if gRainData.columnCount < rainColumnCount {
			for i := 0; i < rainRowCount; i++ {
				CreateGroup(worldID, i, gRainData.columnCount)
			}
			gRainData.columnCount += 1
		} else {
			for i := 0; i < rainRowCount; i++ {
				DestroyGroup(i, gRainData.columnIndex)
				CreateGroup(worldID, i, gRainData.columnIndex)
			}
			gRainData.columnIndex = (gRainData.columnIndex + 1) % rainColumnCount
		}
	}
	return float32(0.0)
}

func CreateSpinner(worldID box2d.World) {
	var groundID box2d.Body
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		groundID = worldID.CreateBody(&bodyDef)
		var points [spinnerPointCount]box2d.Vec2
		var q box2d.Rot = box2d.MakeRot(-float32(2.0) * box2d.Pi / spinnerPointCount)
		var p box2d.Vec2 = box2d.Vec2{X: float32(40.0), Y: float32(0.0)}
		for i := 0; i < spinnerPointCount; i++ {
			points[i] = box2d.Vec2{X: p.X, Y: p.Y + float32(32.0)}
			p = box2d.RotateVector(q, p)
		}
		var material box2d.SurfaceMaterial = box2d.SurfaceMaterial{}
		material.Friction = float32(0.1)
		var chainDef box2d.ChainDef = box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.PointCount = spinnerPointCount
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}
		chainDef.MaterialCount = 1
		groundID.CreateChain(&chainDef)
	}
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: 0.0, Y: float32(12.0)})
		bodyDef.EnableSleep = false
		var spinnerID box2d.Body = worldID.CreateBody(&bodyDef)
		var box box2d.Polygon = box2d.MakeRoundedBox(float32(0.4), float32(20.0), float32(0.2))
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		shapeDef.Material.Friction = float32(0.0)
		spinnerID.CreatePolygonShape(&shapeDef, &box)
		var motorSpeed float32 = float32(5.0)
		var maxMotorTorque float32 = float32(math.MaxFloat32)
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyIDA = groundID
		jointDef.Base.BodyIDB = spinnerID
		jointDef.Base.LocalFrameA.P = groundID.LocalPoint(bodyDef.Position)
		jointDef.EnableMotor = true
		jointDef.MotorSpeed = motorSpeed
		jointDef.MaxMotorTorque = maxMotorTorque
		gSpinnerData.spinnerID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: -float32(0.25), Y: float32(0.0)}, Center2: box2d.Vec2{X: float32(0.25), Y: float32(0.0)}, Radius: float32(0.25)}
	var circle box2d.Circle = box2d.Circle{Center: box2d.Vec2{X: float32(0.0), Y: float32(0.0)}, Radius: float32(0.35)}
	var square box2d.Polygon = box2d.MakeSquare(float32(0.35))
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.Material.Friction = float32(0.1)
	shapeDef.Material.Restitution = float32(0.1)
	shapeDef.Density = float32(0.25)
	var bodyCount int = 2 * 3038
	var x float32 = -float32(23.0)
	var y float32 = float32(2.0)
	for i := 0; i < bodyCount; i++ {
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
		var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
		var remainder int = i % 3
		if remainder == 0 {
			bodyID.CreateCapsuleShape(&shapeDef, &capsule)
		} else if remainder == 1 {
			bodyID.CreateCircleShape(&shapeDef, &circle)
		} else if remainder == 2 {
			bodyID.CreatePolygonShape(&shapeDef, &square)
		}
		x += float32(0.5)
		if x >= float32(23.0) {
			x = -float32(23.0)
			y += float32(0.5)
		}
	}
}

func StepSpinner(worldID box2d.World, stepCount int) float32 {
	return box2d.RevoluteJoint{Joint: gSpinnerData.spinnerID}.Angle()
}

func CreateSmash(worldID box2d.World) {
	worldID.SetGravity(box2d.Vec2{})
	{
		var box box2d.Polygon = box2d.MakeBox(float32(4.0), float32(4.0))
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: -float32(20.0), Y: float32(0.0)})
		bodyDef.LinearVelocity = box2d.Vec2{X: float32(40.0), Y: float32(0.0)}
		var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		shapeDef.Density = float32(8.0)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	var d float32 = float32(0.4)
	var box box2d.Polygon = box2d.MakeSquare(float32(0.5) * d)
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.IsAwake = false
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var columns int = 120
	var rows int = 80
	for i := 0; i < columns; i++ {
		for j := 0; j < rows; j++ {
			bodyDef.Position.X = box2d.ToPos(box2d.Vec2{X: float32(float32(i)*d) + float32(30.0)}).X
			bodyDef.Position.Y = box2d.ToPos(box2d.Vec2{Y: (float32(j) - float32(rows)/float32(2.0)) * d}).Y
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
}

func CreateTumbler(worldID box2d.World) {
	var groundID box2d.Body
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		groundID = worldID.CreateBody(&bodyDef)
	}
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: float32(0.0), Y: float32(10.0)})
		var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		shapeDef.Density = float32(50.0)
		var polygon box2d.Polygon
		polygon = box2d.MakeOffsetBox(float32(0.5), float32(10.0), box2d.Vec2{X: float32(10.0), Y: float32(0.0)}, box2d.Rot{C: 1})
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		polygon = box2d.MakeOffsetBox(float32(0.5), float32(10.0), box2d.Vec2{X: -float32(10.0), Y: float32(0.0)}, box2d.Rot{C: 1})
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		polygon = box2d.MakeOffsetBox(float32(10.0), float32(0.5), box2d.Vec2{X: float32(0.0), Y: float32(10.0)}, box2d.Rot{C: 1})
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		polygon = box2d.MakeOffsetBox(float32(10.0), float32(0.5), box2d.Vec2{X: float32(0.0), Y: -float32(10.0)}, box2d.Rot{C: 1})
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		var motorSpeed float32 = float32(25.0)
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyIDA = groundID
		jointDef.Base.BodyIDB = bodyID
		jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.0), Y: float32(10.0)}
		jointDef.Base.LocalFrameB.P = box2d.Vec2{X: float32(0.0), Y: float32(0.0)}
		jointDef.MotorSpeed = (box2d.Pi / float32(180.0)) * motorSpeed
		jointDef.MaxMotorTorque = float32(1e8)
		jointDef.EnableMotor = true
		worldID.CreateRevoluteJoint(&jointDef)
	}
	var gridCount int = 45
	var polygon box2d.Polygon = box2d.MakeBox(float32(0.125), float32(0.125))
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var y float32 = float32(-float32(0.2)*float32(gridCount)) + float32(10.0)
	for i := 0; i < gridCount; i++ {
		var x float32 = -float32(0.2) * float32(gridCount)
		for j := 0; j < gridCount; j++ {
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &polygon)
			x += float32(0.4)
		}
		y += float32(0.4)
	}
}

func CreateWasher(worldID box2d.World) {
	var kinematic bool = true
	var groundID box2d.Body
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		groundID = worldID.CreateBody(&bodyDef)
	}
	{
		var motorSpeed float32 = float32(25.0)
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: float32(0.0), Y: float32(10.0)})
		if kinematic == true {
			bodyDef.Type = box2d.KinematicBody
			bodyDef.AngularVelocity = (box2d.Pi / float32(180.0)) * motorSpeed
			bodyDef.LinearVelocity = box2d.Vec2{X: float32(0.001), Y: -float32(0.002)}
		} else {
			bodyDef.Type = box2d.DynamicBody
		}
		var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		var r0 float32 = float32(14.0)
		var r1 float32 = float32(16.0)
		var r2 float32 = float32(18.0)
		var angle float32 = box2d.Pi / float32(18.0)
		sine, cosine := boundedSinCos(angle)
		var q box2d.Rot = box2d.Rot{C: cosine, S: sine}
		sine, cosine = boundedSinCos(float32(0.1) * angle)
		var qo box2d.Rot = box2d.Rot{C: cosine, S: sine}
		var u1 box2d.Vec2 = box2d.Vec2{X: float32(1.0), Y: float32(0.0)}
		for i := 0; i < 36; i++ {
			var u2 box2d.Vec2
			if i == 35 {
				u2 = box2d.Vec2{X: float32(1.0), Y: float32(0.0)}
			} else {
				u2 = box2d.RotateVector(q, u1)
			}
			{
				var a1 box2d.Vec2 = box2d.InvRotateVector(qo, u1)
				var a2 box2d.Vec2 = box2d.RotateVector(qo, u2)
				var p1 box2d.Vec2 = box2d.MulSV(r1, a1)
				var p2 box2d.Vec2 = box2d.MulSV(r2, a1)
				var p3 box2d.Vec2 = box2d.MulSV(r1, a2)
				var p4 box2d.Vec2 = box2d.MulSV(r2, a2)
				var points [4]box2d.Vec2 = [4]box2d.Vec2{p1, p2, p3, p4}
				var hull box2d.Hull = box2d.ComputeHull(points[:], 4)
				var polygon box2d.Polygon = box2d.MakePolygon(&hull, float32(0.0))
				bodyID.CreatePolygonShape(&shapeDef, &polygon)
			}
			if i%9 == 0 {
				var p1 box2d.Vec2 = box2d.MulSV(r0, u1)
				var p2 box2d.Vec2 = box2d.MulSV(r1, u1)
				var p3 box2d.Vec2 = box2d.MulSV(r0, u2)
				var p4 box2d.Vec2 = box2d.MulSV(r1, u2)
				var points [4]box2d.Vec2 = [4]box2d.Vec2{p1, p2, p3, p4}
				var hull box2d.Hull = box2d.ComputeHull(points[:], 4)
				var polygon box2d.Polygon = box2d.MakePolygon(&hull, float32(0.0))
				bodyID.CreatePolygonShape(&shapeDef, &polygon)
			}
			u1 = u2
		}
		if kinematic == false {
			var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
			jointDef.Base.BodyIDA = groundID
			jointDef.Base.BodyIDB = bodyID
			jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.0), Y: float32(10.0)}
			jointDef.Base.LocalFrameB.P = box2d.Vec2{X: float32(0.0), Y: float32(0.0)}
			jointDef.MotorSpeed = (box2d.Pi / float32(180.0)) * motorSpeed
			jointDef.MaxMotorTorque = float32(1e8)
			jointDef.EnableMotor = true
			worldID.CreateRevoluteJoint(&jointDef)
		}
	}
	var gridCount int = 90
	var a float32 = float32(0.1)
	var polygon box2d.Polygon = box2d.MakeSquare(a)
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.EnableHitEvents = true
	var y float32 = float32(-float32(1.1)*a*float32(gridCount)) + float32(10.0)
	for i := 0; i < gridCount; i++ {
		var x float32 = -float32(1.1) * a * float32(gridCount)
		for j := 0; j < gridCount; j++ {
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &polygon)
			x += float32(float32(2.1) * a)
		}
		y += float32(float32(2.1) * a)
	}
}

func CreateJunkyard(worldID box2d.World) {
	{
		var gridSize float32 = float32(1.0)
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		var groundID box2d.Body = worldID.CreateBody(&bodyDef)
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		var y float32 = float32(0.0)
		var x float32 = -float32(80.0) * gridSize
		for i := 0; i < 161; i++ {
			var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.55)*gridSize, float32(0.5)*gridSize, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
			groundID.CreatePolygonShape(&shapeDef, &box)
			x += gridSize
		}
		y = gridSize
		x = -float32(80.0) * gridSize
		for i := 0; i < 50; i++ {
			var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.5)*gridSize, float32(0.55)*gridSize, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
			groundID.CreatePolygonShape(&shapeDef, &box)
			y += gridSize
		}
		y = gridSize
		x = float32(80.0) * gridSize
		for i := 0; i < 50; i++ {
			var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.5)*gridSize, float32(0.55)*gridSize, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
			groundID.CreatePolygonShape(&shapeDef, &box)
			y += gridSize
		}
	}
	var columnCount int = 200
	var rowCount int = 40
	var radius float32 = float32(0.25)
	var polygon box2d.Polygon
	{
		var phi float32 = box2d.Pi * (float32(math.Sqrt(float64(float32(5.0)))) - float32(1.0))
		var points [5]box2d.Vec2
		for i := 0; i < 5; i++ {
			var theta float32 = phi * float32(i)
			var cs box2d.CosSin = box2d.ComputeCosSin(theta)
			points[i].X = radius * cs.Cosine
			points[i].Y = radius * cs.Sine
		}
		var hull box2d.Hull = box2d.ComputeHull(points[:], 5)
		polygon = box2d.MakePolygon(&hull, float32(0.0))
	}
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var side float32 = -float32(0.1)
	var yStart float32 = float32(15.0)
	for i := 0; i < columnCount; i++ {
		var x float32 = float32(1.5) * (float32(float32(2.0)*float32(i)) - float32(columnCount)) * radius
		for j := 0; j < rowCount; j++ {
			var y float32 = float32(float32(4.0)*float32(j)*radius) + yStart
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x + side, Y: y})
			side = -side
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &polygon)
		}
	}
	bodyDef.Type = box2d.KinematicBody
	bodyDef.Position = box2d.Pos{}
	gJunkyardData.pusherID = worldID.CreateBody(&bodyDef)
	var box box2d.Polygon = box2d.MakeOffsetBox(float32(2.0), float32(4.0), box2d.Vec2{X: float32(0.0), Y: float32(4.0)}, box2d.Rot{C: 1})
	gJunkyardData.pusherID.CreatePolygonShape(&shapeDef, &box)
}

func StepJunkyard(worldID box2d.World, stepCount int) float32 {
	var timeStep float32 = float32(1.0) / float32(60.0)
	var time float32 = timeStep * float32(stepCount)
	var cs box2d.CosSin = box2d.ComputeCosSin(float32(0.2) * time)
	var target box2d.WorldTransform = box2d.WorldTransform{P: box2d.ToPos(box2d.Vec2{X: float32(60.0) * cs.Sine, Y: float32(0.0)}), Q: box2d.Rot{C: 1}}
	gJunkyardData.pusherID.SetTargetTransform(target, timeStep, true)
	return float32(0.0)
}

func createSleepPyramid(worldID box2d.World, baseCount int, extent, centerX float32, bodyIDA, bodyIDB *box2d.Body) {
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.SleepThreshold = float32(1.0)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var box box2d.Polygon = box2d.MakeSquare(extent)
	for i := 0; i < baseCount; i++ {
		var y float32 = (float32(float32(2.0)*float32(i)) + float32(1.0)) * extent
		for j := i; j < baseCount; j++ {
			var x float32 = float32((float32(i)+float32(1.0))*extent) + float32(float32(2.0)*(float32(j)-float32(i))*extent) + centerX
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
			if i == 0 && j == 0 {
				*bodyIDA = bodyID
			} else if i == 1 && j == 1 {
				*bodyIDB = bodyID
			}
		}
	}
}

func CreateSleep(worldID box2d.World) {
	gSleepData = sleepData{}
	var baseCount int = 60
	var extent float32 = float32(0.5)
	var baseWidth float32 = float32(2.0) * extent * float32(baseCount)
	var pitch float32 = baseWidth + float32(float32(8.0)*extent)
	var span float32 = pitch * (float32(sleepPyramidCount) - 1)
	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: float32(0.0), Y: -float32(1.0)})
		var groundID box2d.Body = worldID.CreateBody(&bodyDef)
		var box box2d.Polygon = box2d.MakeBox(float32(float32(0.5)*span)+baseWidth, float32(1.0))
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}
	for i := 0; i < sleepPyramidCount; i++ {
		var centerX float32 = float32(-float32(0.5)*span) + float32(float32(i)*pitch) - float32(float32(0.5)*baseWidth)
		createSleepPyramid(worldID, baseCount, extent, centerX, &gSleepData.bodyIDA[i], &gSleepData.bodyIDB[i])
	}
}

func StepSleep(worldID box2d.World, stepCount int) float32 {
	for i := 0; i < sleepPyramidCount; i++ {
		if gSleepData.bodyIDA[i].IsAwake() == false {
			var jointDef box2d.FilterJointDef = box2d.DefaultFilterJointDef()
			jointDef.Base.BodyIDA = gSleepData.bodyIDA[i]
			jointDef.Base.BodyIDB = gSleepData.bodyIDB[i]
			var jointID box2d.Joint = worldID.CreateFilterJoint(&jointDef).Joint
			jointID.Destroy()
			break
		}
	}
	return float32(0.0)
}

func CreateCompounds(worldID box2d.World) {
	{
		var gridSize float32 = float32(1.0)
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		var groundID box2d.Body = worldID.CreateBody(&bodyDef)
		var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
		var y float32 = float32(0.0)
		var x float32 = -float32(40.0) * gridSize
		for i := 0; i < 81; i++ {
			var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.55)*gridSize, float32(0.5)*gridSize, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
			groundID.CreatePolygonShape(&shapeDef, &box)
			x += gridSize
		}
		y = gridSize
		x = -float32(40.0) * gridSize
		for i := 0; i < 100; i++ {
			var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.5)*gridSize, float32(0.55)*gridSize, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
			groundID.CreatePolygonShape(&shapeDef, &box)
			y += gridSize
		}
		y = gridSize
		x = float32(40.0) * gridSize
		for i := 0; i < 100; i++ {
			var box box2d.Polygon = box2d.MakeOffsetBox(float32(0.5)*gridSize, float32(0.55)*gridSize, box2d.Vec2{X: x, Y: y}, box2d.Rot{C: 1})
			groundID.CreatePolygonShape(&shapeDef, &box)
			y += gridSize
		}
		var segment box2d.Segment = box2d.Segment{Point1: box2d.Vec2{X: -float32(800.0), Y: -float32(80.0)}, Point2: box2d.Vec2{X: float32(800.0), Y: -float32(80.0)}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}
	var columnCount int = 20
	var rowCount int = 150
	var leftPoints [3]box2d.Vec2 = [3]box2d.Vec2{box2d.Vec2{X: -float32(1.0), Y: float32(0.0)}, box2d.Vec2{X: float32(0.5), Y: float32(1.0)}, box2d.Vec2{X: float32(0.0), Y: float32(2.0)}}
	var leftHull box2d.Hull = box2d.ComputeHull(leftPoints[:], 3)
	var left box2d.Polygon = box2d.MakePolygon(&leftHull, float32(0.0))
	var rightPoints [3]box2d.Vec2 = [3]box2d.Vec2{box2d.Vec2{X: float32(1.0), Y: float32(0.0)}, box2d.Vec2{X: -float32(0.5), Y: float32(1.0)}, box2d.Vec2{X: float32(0.0), Y: float32(2.0)}}
	var rightHull box2d.Hull = box2d.ComputeHull(rightPoints[:], 3)
	var right box2d.Polygon = box2d.MakePolygon(&rightHull, float32(0.0))
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Material.Friction = float32(0.5)
	var shift float32 = float32(2.0)
	var extray float32 = float32(0.25)
	var side float32 = float32(0.25)
	var centerx float32 = shift*float32(columnCount)/float32(2.0) - float32(1.0)
	var centery float32 = float32(1.15) / float32(2.0)
	var yStart float32 = float32(5.0)
	for i := 0; i < columnCount; i++ {
		var x float32 = float32(float32(i)*shift) - centerx
		for j := 0; j < rowCount; j++ {
			var y float32 = float32(float32(j)*(shift+extray)) + centery + yStart
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x + side, Y: y})
			side = -side
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &left)
			bodyID.CreatePolygonShape(&shapeDef, &right)
		}
	}
}

func queryRandom(lower, upper float32) float32 {
	gQueryRandomState = uint32(1664525)*gQueryRandomState + uint32(1013904223)
	var unit float32 = float32((gQueryRandomState >> 8)) * (float32(1.0) / float32(16777216.0))
	return lower + float32((upper-lower)*unit)
}

func CreateQueries(worldID box2d.World) {
	gQueryRandomState = 1234
	var extent float32 = float32(500.0)
	var cellCount int = 500
	var fill float32 = float32(0.1)
	var pileX float32 = float32(0.4) * extent
	var pileY float32 = float32(0.3) * extent
	var clear box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: pileX - float32(10.0), Y: pileY - float32(10.0)}, UpperBound: box2d.Vec2{X: pileX + float32(50.0), Y: pileY + float32(60.0)}}
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	for i := 0; i < cellCount; i++ {
		var y float32 = float32(i)
		for j := 0; j < cellCount; j++ {
			var x float32 = float32(j)
			var fillTest float32 = queryRandom(float32(0.0), float32(1.0))
			var ratio float32 = queryRandom(float32(1.0), float32(5.0))
			var halfWidth float32 = queryRandom(float32(0.05), float32(0.25))
			var orientation float32 = queryRandom(float32(0.0), float32(1.0))
			var category int = int(queryRandom(float32(0.0), float32(2.999)))
			if fillTest > fill {
				continue
			}
			if clear.LowerBound.X <= x && x <= clear.UpperBound.X && clear.LowerBound.Y <= y && y <= clear.UpperBound.Y {
				continue
			}
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: x, Y: y})
			var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
			var box box2d.Polygon
			if orientation > float32(0.5) {
				box = box2d.MakeBox(ratio*halfWidth, halfWidth)
			} else {
				box = box2d.MakeBox(halfWidth, ratio*halfWidth)
			}
			shapeDef.Filter.CategoryBits = uint64(1) << category
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
	shapeDef.Filter.CategoryBits = 1
	{
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: float32(0.0), Y: float32(0.0)})
		var groundID box2d.Body = worldID.CreateBody(&bodyDef)
		var segment box2d.Segment = box2d.Segment{Point1: box2d.Vec2{X: pileX - float32(10.0), Y: pileY}, Point2: box2d.Vec2{X: pileX + float32(50.0), Y: pileY}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}
	{
		var sensorDef box2d.ShapeDef = box2d.DefaultShapeDef()
		sensorDef.IsSensor = true
		var sensorBox box2d.Polygon = box2d.MakeSquare(float32(2.0))
		for i := 0; i < 8; i++ {
			bodyDef.Position = box2d.ToPos(box2d.Vec2{X: pileX + float32(float32(4.0)*float32(i)), Y: pileY + float32(2.0)})
			var sensorId box2d.Body = worldID.CreateBody(&bodyDef)
			sensorId.CreatePolygonShape(&sensorDef, &sensorBox)
		}
	}
	{
		var columnCount int = 40
		var rowCount int = 50
		var spacing float32 = float32(0.55)
		var box box2d.Polygon = box2d.MakeSquare(float32(0.25))
		bodyDef.Type = box2d.DynamicBody
		for i := 0; i < rowCount; i++ {
			for j := 0; j < columnCount; j++ {
				bodyDef.Position = box2d.ToPos(box2d.Vec2{X: pileX + float32(spacing*float32(j)), Y: pileY + float32(1.0) + float32(spacing*float32(i))})
				var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &box)
			}
		}
	}
	{
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: pileX - float32(5.0), Y: pileY + float32(1.2)})
		bodyDef.LinearVelocity = box2d.Vec2{X: float32(5.0), Y: float32(0.0)}
		gQueryBenchmark.paddleID = worldID.CreateBody(&bodyDef)
		var paddle box2d.Polygon = box2d.MakeBox(float32(1.0), float32(1.0))
		gQueryBenchmark.paddleID.CreatePolygonShape(&shapeDef, &paddle)
		gQueryBenchmark.paddleMinX = pileX - float32(5.0)
		gQueryBenchmark.paddleMaxX = pileX + float32(30.0)
	}
	queryTotal := queryCount
	for i := 0; i < queryTotal; i++ {
		var x1 float32 = queryRandom(float32(0.0), extent)
		var y1 float32 = queryRandom(float32(0.0), extent)
		var x2 float32 = queryRandom(float32(0.0), extent)
		var y2 float32 = queryRandom(float32(0.0), extent)
		gQueryBenchmark.origins[i] = box2d.ToPos(box2d.Vec2{X: x1, Y: y1})
		gQueryBenchmark.translations[i] = box2d.Vec2{X: x2 - x1, Y: y2 - y1}
	}
	gQueryBenchmark.stats = box2d.TreeStats{}
	gQueryBenchmark.extent = extent
}

func queryCastClosest(shapeID box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32, result *float32) float32 {
	*result = fraction
	return fraction
}

func queryOverlapCount(shapeID box2d.Shape, count *int) bool {
	*count += 1
	return true
}

func StepQueries(worldID box2d.World, stepCount int) float32 {
	var data *queryBenchmark = &gQueryBenchmark
	var paddlePosition box2d.Pos = data.paddleID.Position()
	if paddlePosition.X > box2d.ToPos(box2d.Vec2{X: data.paddleMaxX}).X {
		data.paddleID.SetLinearVelocity(box2d.Vec2{X: -float32(5.0), Y: float32(0.0)})
	} else if paddlePosition.X < box2d.ToPos(box2d.Vec2{X: data.paddleMinX}).X {
		data.paddleID.SetLinearVelocity(box2d.Vec2{X: float32(5.0), Y: float32(0.0)})
	}
	var filters [2]box2d.QueryFilter = [2]box2d.QueryFilter{box2d.DefaultQueryFilter(), box2d.DefaultQueryFilter()}
	filters[1].MaskBits = 1
	var circleCenter box2d.Vec2 = box2d.Vec2{}
	var circle box2d.ShapeProxy = box2d.MakeProxy([]box2d.Vec2{circleCenter}, float32(0.1))
	queryTotal := queryCount
	var hitCount int = 0
	var stats box2d.TreeStats = box2d.TreeStats{}
	for f := 0; f < 2; f++ {
		var filter box2d.QueryFilter = filters[f]
		for i := 0; i < queryTotal; i++ {
			var result box2d.RayResult = worldID.CastRayClosest(data.origins[i], data.translations[i], filter)
			stats.NodeVisits += result.NodeVisits
			stats.LeafVisits += result.LeafVisits
			if result.Hit {
				hitCount++
			}
		}
		for i := 0; i < queryTotal; i++ {
			var fraction float32 = float32(1.0)
			var castStats box2d.TreeStats = worldID.CastShape(data.origins[i], &circle, data.translations[i], filter, func(id box2d.Shape, point box2d.Pos, normal box2d.Vec2, value float32) float32 {
				return queryCastClosest(id, point, normal, value, &fraction)
			})
			stats.NodeVisits += castStats.NodeVisits
			stats.LeafVisits += castStats.LeafVisits
			if fraction < float32(1.0) {
				hitCount++
			}
		}
		for i := 0; i < queryTotal; i++ {
			var overlapCount int = 0
			var aabb box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: -float32(5.0), Y: -float32(5.0)}, UpperBound: box2d.Vec2{X: float32(5.0), Y: float32(5.0)}}
			var overlapStats box2d.TreeStats = worldID.OverlapAABB(data.origins[i], aabb, filter, func(id box2d.Shape) bool { return queryOverlapCount(id, &overlapCount) })
			stats.NodeVisits += overlapStats.NodeVisits
			stats.LeafVisits += overlapStats.LeafVisits
			hitCount += overlapCount
		}
	}
	data.stats.NodeVisits += stats.NodeVisits
	data.stats.LeafVisits += stats.LeafVisits
	return float32(hitCount)
}

func GetQueryBenchmarkStats() box2d.TreeStats {
	return gQueryBenchmark.stats
}

func GetQueryBenchmarkCount() int {
	return queryCount
}

func GetQueryBenchmarkExtent() float32 {
	return gQueryBenchmark.extent
}

func GetQueryBenchmarkRay(index int) (box2d.Pos, box2d.Vec2) {
	return gQueryBenchmark.origins[index], gQueryBenchmark.translations[index]
}

func treeCastRayBox(box box2d.AABB, p, d box2d.Vec2, maxFraction float32) float32 {
	var tmin float32 = float32(0.0)
	var tmax float32 = maxFraction
	if d.X == float32(0.0) {
		if p.X < box.LowerBound.X || box.UpperBound.X < p.X {
			return -float32(1.0)
		}
	} else {
		var inv float32 = float32(1.0) / d.X
		var t1 float32 = (box.LowerBound.X - p.X) * inv
		var t2 float32 = (box.UpperBound.X - p.X) * inv
		if t1 > t2 {
			var t float32 = t1
			t1 = t2
			t2 = t
		}
		tmin = box2d.MaxFloat(tmin, t1)
		tmax = box2d.MinFloat(tmax, t2)
		if tmin > tmax {
			return -float32(1.0)
		}
	}
	if d.Y == float32(0.0) {
		if p.Y < box.LowerBound.Y || box.UpperBound.Y < p.Y {
			return -float32(1.0)
		}
	} else {
		var inv float32 = float32(1.0) / d.Y
		var t1 float32 = (box.LowerBound.Y - p.Y) * inv
		var t2 float32 = (box.UpperBound.Y - p.Y) * inv
		if t1 > t2 {
			var t float32 = t1
			t1 = t2
			t2 = t
		}
		tmin = box2d.MaxFloat(tmin, t1)
		tmax = box2d.MinFloat(tmax, t2)
		if tmin > tmax {
			return -float32(1.0)
		}
	}
	if tmin <= float32(0.0) {
		return -float32(1.0)
	}
	return tmin
}

func treeCastRayCallback(input *box2d.RayCastInput, proxyID int, userData uint64, context *treeCastContext) float32 {
	var cast *treeCastContext = context
	var fraction float32 = treeCastRayBox(cast.tightBoxes[userData], input.Origin, input.Translation, input.MaxFraction)
	if fraction < float32(0.0) {
		return -float32(1.0)
	}
	cast.fraction = fraction
	return fraction
}

func treeCastBoxCallback(input *box2d.BoxCastInput, proxyID int, userData uint64, context *treeCastContext) float32 {
	var cast *treeCastContext = context
	var tight box2d.AABB = cast.tightBoxes[userData]
	var extents box2d.Vec2 = (input.Box).Extents()
	var grown box2d.AABB = box2d.AABB{LowerBound: box2d.Sub(tight.LowerBound, extents), UpperBound: box2d.Add(tight.UpperBound, extents)}
	var fraction float32 = treeCastRayBox(grown, (input.Box).Center(), input.Translation, input.MaxFraction)
	if fraction < float32(0.0) {
		return -float32(1.0)
	}
	cast.fraction = fraction
	return fraction
}

func treeCastAddBox(center box2d.Vec2, hx, hy float32, categoryBits uint64) {
	var data *treeCastBenchmark = &gTreeCast
	if !(data.proxyCount < data.proxyCapacity) {
		panic(&box2d.AssertError{Condition: "data->proxyCount < data->proxyCapacity"})
	}
	var tight box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: center.X - hx, Y: center.Y - hy}, UpperBound: box2d.Vec2{X: center.X + hx, Y: center.Y + hy}}
	var fat box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: tight.LowerBound.X - float32(2*box2d.SpeculativeDistance()), Y: tight.LowerBound.Y - float32(2*box2d.SpeculativeDistance())}, UpperBound: box2d.Vec2{X: tight.UpperBound.X + float32(2*box2d.SpeculativeDistance()), Y: tight.UpperBound.Y + float32(2*box2d.SpeculativeDistance())}}
	var index int = data.proxyCount
	data.tightBoxes[index] = tight
	data.proxyCount += 1
	data.tree.CreateProxy(fat, categoryBits, uint64(index))
}

func DestroyTreeCast() {
	var data *treeCastBenchmark = &gTreeCast
	if data.created {
		data.tree.Destroy()
	}
	var stats box2d.TreeStats = data.stats
	*data = treeCastBenchmark{}
	data.stats = stats
}

func CreateTreeCast(worldID box2d.World) {
	DestroyTreeCast()
	var data *treeCastBenchmark = &gTreeCast
	gQueryRandomState = 1234
	var extent float32 = float32(500.0)
	var cellCount int = 500
	var fill float32 = float32(0.1)
	var pileX float32 = float32(0.4) * extent
	var pileY float32 = float32(0.3) * extent
	var clear box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: pileX - float32(10.0), Y: pileY - float32(10.0)}, UpperBound: box2d.Vec2{X: pileX + float32(50.0), Y: pileY + float32(60.0)}}
	data.proxyCapacity = cellCount * cellCount / 5
	data.tightBoxes = make([]box2d.AABB, data.proxyCapacity)
	data.tree = box2d.CreateDynamicTree(data.proxyCapacity)
	data.created = true
	for i := 0; i < cellCount; i++ {
		var y float32 = float32(i)
		for j := 0; j < cellCount; j++ {
			var x float32 = float32(j)
			var fillTest float32 = queryRandom(float32(0.0), float32(1.0))
			var ratio float32 = queryRandom(float32(1.0), float32(5.0))
			var halfWidth float32 = queryRandom(float32(0.05), float32(0.25))
			var orientation float32 = queryRandom(float32(0.0), float32(1.0))
			var category int = int(queryRandom(float32(0.0), float32(2.999)))
			if fillTest > fill {
				continue
			}
			if clear.LowerBound.X <= x && x <= clear.UpperBound.X && clear.LowerBound.Y <= y && y <= clear.UpperBound.Y {
				continue
			}
			var hx float32
			if orientation > float32(0.5) {
				hx = ratio * halfWidth
			} else {
				hx = halfWidth
			}
			var hy float32
			if orientation > float32(0.5) {
				hy = halfWidth
			} else {
				hy = ratio * halfWidth
			}
			treeCastAddBox(box2d.Vec2{X: x, Y: y}, hx, hy, uint64(1)<<category)
		}
	}
	treeCastAddBox(box2d.Vec2{X: pileX + float32(20.0), Y: pileY}, float32(30.0), float32(0.0), 1)
	for i := 0; i < 8; i++ {
		treeCastAddBox(box2d.Vec2{X: pileX + float32(float32(4.0)*float32(i)), Y: pileY + float32(2.0)}, float32(2.0), float32(2.0), 1)
	}
	var lengths [treeCastLengthCount]float32 = [treeCastLengthCount]float32{float32(0.0), float32(25.0), float32(5.0)}
	for i := 0; i < treeCastRayCount; i++ {
		var x1 float32 = queryRandom(float32(0.0), extent)
		var y1 float32 = queryRandom(float32(0.0), extent)
		var x2 float32 = queryRandom(float32(0.0), extent)
		var y2 float32 = queryRandom(float32(0.0), extent)
		var t box2d.Vec2 = box2d.Vec2{X: x2 - x1, Y: y2 - y1}
		data.origins[i] = box2d.Vec2{X: x1, Y: y1}
		data.translations[0][i] = t
		var length float32 = box2d.Length(t)
		for k := 1; k < treeCastLengthCount; k++ {
			var scale float32
			if length > float32(0.0) {
				scale = lengths[k] / length
			} else {
				scale = float32(0.0)
			}
			data.translations[k][i] = box2d.MulSV(scale, t)
		}
	}
	data.stats = box2d.TreeStats{}
}

func StepTreeCast(worldID box2d.World, stepCount int) float32 {
	var data *treeCastBenchmark = &gTreeCast
	var tree *box2d.DynamicTree = &data.tree
	var masks [2]uint64 = [2]uint64{box2d.DefaultMaskBits, 1}
	var rayCount int = treeCastRayCount
	var stats box2d.TreeStats = box2d.TreeStats{}
	var hitCount int = 0
	for k := 0; k < treeCastLengthCount; k++ {
		translations := data.translations[k]
		for m := 0; m < 2; m++ {
			for i := 0; i < rayCount; i++ {
				var input box2d.RayCastInput = box2d.RayCastInput{Origin: data.origins[i], Translation: translations[i], MaxFraction: float32(1.0)}
				context := treeCastContext{tightBoxes: data.tightBoxes, fraction: 1}
				var castStats box2d.TreeStats = tree.CastRay(&input, masks[m], func(input *box2d.RayCastInput, id int, value uint64) float32 {
					return treeCastRayCallback(input, id, value, &context)
				})
				stats.NodeVisits += castStats.NodeVisits
				stats.LeafVisits += castStats.LeafVisits
				if context.fraction < float32(1.0) {
					hitCount++
				}
			}
		}
		for i := 0; i < rayCount; i++ {
			var o box2d.Vec2 = data.origins[i]
			var box box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: o.X - treeCastBoxExtent, Y: o.Y - treeCastBoxExtent}, UpperBound: box2d.Vec2{X: o.X + treeCastBoxExtent, Y: o.Y + treeCastBoxExtent}}
			var input box2d.BoxCastInput = box2d.BoxCastInput{Box: box, Translation: translations[i], MaxFraction: float32(1.0)}
			context := treeCastContext{tightBoxes: data.tightBoxes, fraction: 1}
			var castStats box2d.TreeStats = tree.CastBox(&input, box2d.DefaultMaskBits, func(input *box2d.BoxCastInput, id int, value uint64) float32 {
				return treeCastBoxCallback(input, id, value, &context)
			})
			stats.NodeVisits += castStats.NodeVisits
			stats.LeafVisits += castStats.LeafVisits
			if context.fraction < float32(1.0) {
				hitCount++
			}
		}
	}
	data.stats.NodeVisits += stats.NodeVisits
	data.stats.LeafVisits += stats.LeafVisits
	return float32(hitCount)
}

func GetTreeCastBenchmarkStats() box2d.TreeStats {
	return gTreeCast.stats
}

func tileRandom(state *uint32, lower, upper float32) float32 {
	*state = uint32(1664525)**state + uint32(1013904223)
	var unit float32 = float32((*state >> 8)) * (float32(1.0) / float32(16777216.0))
	return lower + float32((upper-lower)*unit)
}

func createTileGround(worldID box2d.World, xBody float32) box2d.Body {
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Position.X = box2d.ToPos(box2d.Vec2{X: xBody}).X
	var groundID box2d.Body = worldID.CreateBody(&bodyDef)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.InvokeContactCreation = false
	var omega float32 = float32(2.0) * box2d.Pi / tilePeriod
	var xShape float32 = float32(0.0)
	for i := 0; i < tileColumnsPerBody; i++ {
		var columnCount int = int(math.Round(float64(float32(4.0)*float32(math.Cos(float64(omega*(xBody+xShape))))))) + 12
		var y float32 = float32(0.0)
		for j := 0; j < columnCount; j++ {
			var square box2d.Polygon = box2d.MakeOffsetBox(float32(0.4)*tileGridSize, float32(0.4)*tileGridSize, box2d.Vec2{X: xShape, Y: y}, box2d.Rot{C: 1})
			square.Radius = float32(0.1)
			groundID.CreatePolygonShape(&shapeDef, &square)
			y += tileGridSize
		}
		xShape += tileGridSize
	}
	return groundID
}

func DestroyTileWorld() {
	var data *tileWorldBenchmark = &gTileWorld
	var stats box2d.TreeStats = data.stats
	*data = tileWorldBenchmark{}
	data.stats = stats
}

func CreateTileWorld(worldID box2d.World) {
	DestroyTileWorld()
	var data *tileWorldBenchmark = &gTileWorld
	var gridCount int = int((tileCycleCount * tilePeriod / tileGridSize))
	data.groundCount = gridCount / tileColumnsPerBody
	data.groundIDs = make([]box2d.Body, data.groundCount)
	data.bodySpan = tileColumnsPerBody * tileGridSize
	data.xLeft = -float32(0.5) * (tileCycleCount * tilePeriod)
	data.xNext = data.xLeft
	data.groundHead = 0
	for i := 0; i < data.groundCount; i++ {
		data.groundIDs[i] = createTileGround(worldID, data.xNext)
		data.xNext += data.bodySpan
	}
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var box box2d.Polygon = box2d.MakeBox(float32(0.3), float32(0.2))
	for cycleIndex := tileCycleCount / 5; cycleIndex < tileCycleCount; cycleIndex += 12 {
		var xBase float32 = float32((float32(0.5)+float32(cycleIndex))*tilePeriod) + data.xLeft
		bodyDef.Position.X = box2d.ToPos(box2d.Vec2{X: xBase - float32(3.0)}).X
		for i := 0; i < 10; i++ {
			bodyDef.Position.Y = box2d.ToPos(box2d.Vec2{Y: float32(10.0)}).Y
			for j := 0; j < 5; j++ {
				var bodyID box2d.Body = worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &box)
				bodyDef.Position.Y += 0.5
			}
			bodyDef.Position.X += box2d.ToPos(box2d.Vec2{X: float32(0.6)}).X
		}
	}
	data.stats = box2d.TreeStats{}
}

func tileOverlapCount(shapeID box2d.Shape, count *int) bool {
	*count += 1
	return true
}

func StepTileWorld(worldID box2d.World, stepCount int) float32 {
	var data *tileWorldBenchmark = &gTileWorld
	data.groundIDs[data.groundHead].Destroy()
	data.groundIDs[data.groundHead] = createTileGround(worldID, data.xNext)
	data.groundHead = (data.groundHead + 1) % data.groundCount
	data.xNext += data.bodySpan
	data.xLeft += data.bodySpan
	var state uint32 = uint32(1234) + uint32(7919)*uint32(stepCount)
	var span float32 = float32(data.groundCount) * data.bodySpan
	var filter box2d.QueryFilter = box2d.DefaultQueryFilter()
	var aabb box2d.AABB = box2d.AABB{LowerBound: box2d.Vec2{X: -float32(5.0), Y: -float32(5.0)}, UpperBound: box2d.Vec2{X: float32(5.0), Y: float32(5.0)}}
	var stats box2d.TreeStats = box2d.TreeStats{}
	var hitCount int = 0
	for i := 0; i < tileQueryCount; i++ {
		var origin box2d.Pos = box2d.ToPos(box2d.Vec2{X: data.xLeft + tileRandom(&state, float32(0.0), span), Y: tileRandom(&state, -float32(2.0), float32(18.0))})
		var angle float32 = tileRandom(&state, float32(0.0), float32(2.0)*box2d.Pi)
		sine, cosine := boundedSinCos(angle)
		var translation box2d.Vec2 = box2d.Vec2{X: tileRayLength * cosine, Y: tileRayLength * sine}
		var result box2d.RayResult = worldID.CastRayClosest(origin, translation, filter)
		stats.NodeVisits += result.NodeVisits
		stats.LeafVisits += result.LeafVisits
		if result.Hit {
			hitCount++
		}
		var overlapCount int = 0
		var overlapStats box2d.TreeStats = worldID.OverlapAABB(origin, aabb, filter, func(id box2d.Shape) bool { return tileOverlapCount(id, &overlapCount) })
		stats.NodeVisits += overlapStats.NodeVisits
		stats.LeafVisits += overlapStats.LeafVisits
		hitCount += overlapCount
	}
	data.stats.NodeVisits += stats.NodeVisits
	data.stats.LeafVisits += stats.LeafVisits
	return float32(hitCount)
}

func GetTileWorldBenchmarkStats() box2d.TreeStats {
	return gTileWorld.stats
}

func boundedSinCos(angle float32) (float32, float32) {
	negative := angle < 0
	x := float64(angle)
	if negative {
		x = -x
	}
	quadrant := int(float64(x*0x1.45f306dc9c883p-1) + 0.5)
	r := x - float64(float64(quadrant)*0x1.921fb54442d18p0)
	z := r * r
	sp := -0x1.994eb3774cf24p-13
	sp = 0x1.1107605230bc4p-7 + float64(z*sp)
	sp = -0x1.555545995a603p-3 + float64(z*sp)
	s := r + float64(float64(r*z)*sp)
	cp := 0x1.99343027bf8c3p-16
	cp = -0x1.6c087e89a359dp-10 + float64(z*cp)
	cp = 0x1.55553e1068f19p-5 + float64(z*cp)
	cp = -0x1.ffffffd0c621cp-2 + float64(z*cp)
	c := 1.0 + float64(z*cp)
	sine, cosine := float32(s), float32(c)
	switch quadrant & 3 {
	case 1:
		sine, cosine = cosine, -sine
	case 2:
		sine, cosine = -sine, -cosine
	case 3:
		sine, cosine = -cosine, sine
	}
	if negative {
		sine = -sine
	}
	if angle == 0 {
		sine = angle
	}
	return sine, cosine
}
