// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"math"

	"github.com/zigai/box2d-go/box2d"
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

var (
	gRainData         rainData
	gSpinnerData      spinnerData
	gJunkyardData     junkyardData
	gSleepData        sleepData
	gQueryBenchmark   queryBenchmark
	gQueryRandomState uint32
	gTreeCast         treeCastBenchmark
	gTileWorld        tileWorldBenchmark
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

func CreateJointGrid(worldID *box2d.World) {
	worldID.SetSleepingEnabled(false)

	var (
		gridSize = 100
		bodies   = make([]box2d.Body, gridSize*gridSize)
		index    = 0
		shapeDef = box2d.DefaultShapeDef()
	)

	shapeDef.Density = float32(1.0)
	shapeDef.Filter.CategoryBits = 2
	shapeDef.Filter.MaskBits = uint64(^uint32(2))

	var (
		circle = box2d.Circle{
			Center: box2d.Vec2{X: float32(0.0), Y: float32(0.0)},
			Radius: float32(0.4),
		}
		jointDef = box2d.DefaultRevoluteJointDef()
	)

	jointDef.Base.DrawScale = float32(0.4)

	bodyDef := box2d.DefaultBodyDef()

	for k := range gridSize {
		for i := range gridSize {
			var (
				fk = float32(k)
				fi = float32(i)
			)
			if k >= gridSize/2-3 && k <= gridSize/2+3 && i == 0 {
				bodyDef.Type = box2d.StaticBody
			} else {
				bodyDef.Type = box2d.DynamicBody
			}

			bodyDef.Position = box2d.Vec2{X: fk, Y: -fi}.Pos()

			body := worldID.CreateBody(&bodyDef)
			body.CreateCircleShape(&shapeDef, &circle)

			if i > 0 {
				jointDef.Base.BodyA = bodies[index-1]
				jointDef.Base.BodyB = body
				jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.0), Y: -float32(0.5)}
				jointDef.Base.LocalFrameB.P = box2d.Vec2{X: float32(0.0), Y: float32(0.5)}
				worldID.CreateRevoluteJoint(&jointDef)
			}

			if k > 0 {
				jointDef.Base.BodyA = bodies[index-gridSize]
				jointDef.Base.BodyB = body
				jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.5), Y: float32(0.0)}
				jointDef.Base.LocalFrameB.P = box2d.Vec2{X: -float32(0.5), Y: float32(0.0)}
				worldID.CreateRevoluteJoint(&jointDef)
			}

			bodies[index] = body
			index++
		}
	}
}

func CreateLargePyramid(worldID *box2d.World) {
	worldID.SetSleepingEnabled(false)

	baseCount := 100
	{
		bodyDef := box2d.DefaultBodyDef()

		bodyDef.Position = box2d.Vec2{X: float32(0.0), Y: -float32(1.0)}.Pos()

		var (
			groundID = worldID.CreateBody(&bodyDef)
			box      = box2d.MakeBox(float32(100.0), float32(1.0))
			shapeDef = box2d.DefaultShapeDef()
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()

	shapeDef.Density = float32(1.0)

	var (
		a     = float32(0.5)
		box   = box2d.MakeSquare(a)
		shift = float32(1.0) * a
	)
	for i := range baseCount {
		y := (float32(float32(2.0)*float32(i)) + float32(1.0)) * shift
		for j := i; j < baseCount; j++ {
			x := float32(
				(float32(i)+float32(1.0))*shift,
			) + float32(
				float32(2.0)*(float32(j)-float32(i))*shift,
			) - float32(
				a*float32(baseCount),
			)

			bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
}

func CreateManyPyramids(worldID *box2d.World) {
	worldID.SetSleepingEnabled(false)

	var (
		baseCount    = 10
		extent       = float32(0.5)
		rowCount     = 20
		columnCount  = 20
		bodyDef      = box2d.DefaultBodyDef()
		groundID     = worldID.CreateBody(&bodyDef)
		groundDeltaY = float32(2.0) * extent * (float32(baseCount) + float32(1.0))
		groundWidth  = float32(
			2.0,
		) * extent * float32(
			columnCount,
		) * (float32(baseCount) + float32(1.0))
		shapeDef = box2d.DefaultShapeDef()
		groundY  = float32(0.0)
	)
	for range rowCount {
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -float32(0.5) * groundWidth, Y: groundY},
			Point2: box2d.Vec2{X: float32(0.5) * groundWidth, Y: groundY},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		groundY += groundDeltaY
	}

	var (
		baseWidth = float32(2.0) * extent * float32(baseCount)
		baseY     = float32(0.0)
	)

	for range rowCount {
		for j := range columnCount {
			centerX := float32(
				-float32(0.5)*groundWidth,
			) + float32(
				float32(j)*(baseWidth+float32(float32(2.0)*extent)),
			) + float32(
				float32(2.0)*extent,
			)
			createSmallPyramid(worldID, baseCount, extent, centerX, baseY)
		}

		baseY += groundDeltaY
	}
}

func GetManyPyramidsCapacity() box2d.Capacity {
	c := box2d.Capacity{
		StaticShapeCount:  20,
		StaticBodyCount:   1,
		DynamicShapeCount: 22000,
		DynamicBodyCount:  22000,
		ContactCount:      58000,
	}

	return c
}

func CreateRain(worldID *box2d.World) {
	gRainData = rainData{}
	gRainData.gridSize = float32(0.5)
	gRainData.gridCount = 500
	{
		var (
			bodyDef  = box2d.DefaultBodyDef()
			groundID = worldID.CreateBody(&bodyDef)
			shapeDef = box2d.DefaultShapeDef()
			y        = float32(0.0)
			width    = gRainData.gridSize
			height   = gRainData.gridSize
		)
		for range rainRowCount {
			x := -float32(0.5) * float32(gRainData.gridCount) * gRainData.gridSize
			for j := 0; j <= gRainData.gridCount; j++ {
				box := box2d.MakeOffsetBox(
					float32(0.5)*width,
					float32(0.5)*height,
					box2d.Vec2{X: x, Y: y},
					box2d.Rot{C: 1},
				)
				groundID.CreatePolygonShape(&shapeDef, &box)

				x += gRainData.gridSize
			}

			y += float32(45.0)
		}
	}

	gRainData.columnCount = 0
	gRainData.columnIndex = 0
}

func CreateGroup(worldID *box2d.World, rowIndex int, columnIndex int) {
	if rowIndex >= rainRowCount || columnIndex >= rainColumnCount {
		panic(
			&box2d.AssertError{
				Condition: "rowIndex < RAIN_ROW_COUNT && columnIndex < RAIN_COLUMN_COUNT",
			},
		)
	}

	var (
		groupIndex    = rowIndex*rainColumnCount + columnIndex
		span          = float32(gRainData.gridCount) * gRainData.gridSize
		groupDistance = float32(1.0) * span / float32(rainColumnCount)
		position      box2d.Pos
	)

	position.X = box2d.Vec2{
		X: float32(-float32(0.5)*span) + float32(groupDistance*(float32(columnIndex)+float32(0.5))),
	}.Pos().X
	position.Y = box2d.Vec2{Y: float32(40.0) + float32(float32(45.0)*float32(rowIndex))}.Pos().Y

	var (
		scale         = float32(1.0)
		jointFriction = float32(0.05)
		jointHertz    = float32(5.0)
		jointDamping  = float32(0.5)
	)

	for i := range rainGroupSize {
		human := &gRainData.groups[groupIndex].humans[i]
		CreateHuman(
			human,
			worldID,
			position,
			scale,
			jointFriction,
			jointHertz,
			jointDamping,
			i+1,
			nil,
			false,
		)
		position.X += 0.5
	}
}

func DestroyGroup(rowIndex int, columnIndex int) {
	if rowIndex >= rainRowCount || columnIndex >= rainColumnCount {
		panic(
			&box2d.AssertError{
				Condition: "rowIndex < RAIN_ROW_COUNT && columnIndex < RAIN_COLUMN_COUNT",
			},
		)
	}

	groupIndex := rowIndex*rainColumnCount + columnIndex
	for i := range rainGroupSize {
		DestroyHuman(&gRainData.groups[groupIndex].humans[i])
	}
}

func StepRain(worldID *box2d.World, stepCount int) float32 {
	delay := 0x7
	if (stepCount & delay) == 0 {
		if gRainData.columnCount < rainColumnCount {
			for i := range rainRowCount {
				CreateGroup(worldID, i, gRainData.columnCount)
			}

			gRainData.columnCount += 1
		} else {
			for i := range rainRowCount {
				DestroyGroup(i, gRainData.columnIndex)
				CreateGroup(worldID, i, gRainData.columnIndex)
			}

			gRainData.columnIndex = (gRainData.columnIndex + 1) % rainColumnCount
		}
	}

	return float32(0.0)
}

func CreateSpinner(worldID *box2d.World) {
	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()

		groundID = worldID.CreateBody(&bodyDef)

		var (
			points [spinnerPointCount]box2d.Vec2
			q      = box2d.MakeRot(-float32(2.0) * box2d.Pi / spinnerPointCount)
			p      = box2d.Vec2{X: float32(40.0), Y: float32(0.0)}
		)
		for i := range spinnerPointCount {
			points[i] = box2d.Vec2{X: p.X, Y: p.Y + float32(32.0)}
			p = q.Rotate(p)
		}

		material := box2d.SurfaceMaterial{}

		material.Friction = float32(0.1)

		chainDef := box2d.DefaultChainDef()

		chainDef.Points = points[:]
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}
		groundID.CreateChain(&chainDef)
	}
	{
		bodyDef := box2d.DefaultBodyDef()

		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Vec2{X: 0.0, Y: float32(12.0)}.Pos()
		bodyDef.EnableSleep = false

		var (
			spinnerID = worldID.CreateBody(&bodyDef)
			box       = box2d.MakeRoundedBox(float32(0.4), float32(20.0), float32(0.2))
			shapeDef  = box2d.DefaultShapeDef()
		)

		shapeDef.Material.Friction = float32(0.0)
		spinnerID.CreatePolygonShape(&shapeDef, &box)

		var (
			motorSpeed     = float32(5.0)
			maxMotorTorque = float32(math.MaxFloat32)
			jointDef       = box2d.DefaultRevoluteJointDef()
		)

		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = spinnerID
		jointDef.Base.LocalFrameA.P = groundID.LocalPoint(bodyDef.Position)
		jointDef.EnableMotor = true
		jointDef.MotorSpeed = motorSpeed
		jointDef.MaxMotorTorque = maxMotorTorque
		gSpinnerData.spinnerID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}

	var (
		capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: -float32(0.25), Y: float32(0.0)},
			Center2: box2d.Vec2{X: float32(0.25), Y: float32(0.0)},
			Radius:  float32(0.25),
		}
		circle = box2d.Circle{
			Center: box2d.Vec2{X: float32(0.0), Y: float32(0.0)},
			Radius: float32(0.35),
		}
		square  = box2d.MakeSquare(float32(0.35))
		bodyDef = box2d.DefaultBodyDef()
	)

	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()

	shapeDef.Material.Friction = float32(0.1)
	shapeDef.Material.Restitution = float32(0.1)
	shapeDef.Density = float32(0.25)

	var (
		bodyCount = 2 * 3038
		x         = -float32(23.0)
		y         = float32(2.0)
	)
	for i := range bodyCount {
		bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

		var (
			bodyID    = worldID.CreateBody(&bodyDef)
			remainder = i % 3
		)
		switch remainder {
		case 0:
			bodyID.CreateCapsuleShape(&shapeDef, &capsule)
		case 1:
			bodyID.CreateCircleShape(&shapeDef, &circle)
		case 2:
			bodyID.CreatePolygonShape(&shapeDef, &square)
		}

		x += float32(0.5)
		if x >= float32(23.0) {
			x = -float32(23.0)
			y += float32(0.5)
		}
	}
}

func StepSpinner(worldID *box2d.World, stepCount int) float32 {
	return box2d.RevoluteJoint{Joint: gSpinnerData.spinnerID}.Angle()
}

func CreateSmash(worldID *box2d.World) {
	worldID.SetGravity(box2d.Vec2{})
	{
		var (
			box     = box2d.MakeBox(float32(4.0), float32(4.0))
			bodyDef = box2d.DefaultBodyDef()
		)

		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Vec2{X: -float32(20.0), Y: float32(0.0)}.Pos()
		bodyDef.LinearVelocity = box2d.Vec2{X: float32(40.0), Y: float32(0.0)}

		var (
			bodyID   = worldID.CreateBody(&bodyDef)
			shapeDef = box2d.DefaultShapeDef()
		)

		shapeDef.Density = float32(8.0)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	var (
		d       = float32(0.4)
		box     = box2d.MakeSquare(float32(0.5) * d)
		bodyDef = box2d.DefaultBodyDef()
	)

	bodyDef.Type = box2d.DynamicBody
	bodyDef.IsAwake = false

	var (
		shapeDef = box2d.DefaultShapeDef()
		columns  = 120
		rows     = 80
	)
	for i := range columns {
		for j := range rows {
			bodyDef.Position.X = box2d.Vec2{X: float32(float32(i)*d) + float32(30.0)}.Pos().X
			bodyDef.Position.Y = box2d.Vec2{
				Y: (float32(j) - float32(rows)/float32(2.0)) * d,
			}.Pos().Y

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
}

func CreateTumbler(worldID *box2d.World) {
	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()

		groundID = worldID.CreateBody(&bodyDef)
	}
	{
		bodyDef := box2d.DefaultBodyDef()

		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Vec2{X: float32(0.0), Y: float32(10.0)}.Pos()

		var (
			bodyID   = worldID.CreateBody(&bodyDef)
			shapeDef = box2d.DefaultShapeDef()
		)

		shapeDef.Density = float32(50.0)

		var polygon box2d.Polygon

		polygon = box2d.MakeOffsetBox(
			float32(0.5),
			float32(10.0),
			box2d.Vec2{X: float32(10.0), Y: float32(0.0)},
			box2d.Rot{C: 1},
		)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		polygon = box2d.MakeOffsetBox(
			float32(0.5),
			float32(10.0),
			box2d.Vec2{X: -float32(10.0), Y: float32(0.0)},
			box2d.Rot{C: 1},
		)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		polygon = box2d.MakeOffsetBox(
			float32(10.0),
			float32(0.5),
			box2d.Vec2{X: float32(0.0), Y: float32(10.0)},
			box2d.Rot{C: 1},
		)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
		polygon = box2d.MakeOffsetBox(
			float32(10.0),
			float32(0.5),
			box2d.Vec2{X: float32(0.0), Y: -float32(10.0)},
			box2d.Rot{C: 1},
		)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)

		var (
			motorSpeed = float32(25.0)
			jointDef   = box2d.DefaultRevoluteJointDef()
		)

		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.0), Y: float32(10.0)}
		jointDef.Base.LocalFrameB.P = box2d.Vec2{X: float32(0.0), Y: float32(0.0)}
		jointDef.MotorSpeed = (box2d.Pi / float32(180.0)) * motorSpeed
		jointDef.MaxMotorTorque = float32(1e8)
		jointDef.EnableMotor = true
		worldID.CreateRevoluteJoint(&jointDef)
	}

	var (
		gridCount = 45
		polygon   = box2d.MakeBox(float32(0.125), float32(0.125))
		bodyDef   = box2d.DefaultBodyDef()
	)

	bodyDef.Type = box2d.DynamicBody

	var (
		shapeDef = box2d.DefaultShapeDef()
		y        = float32(-float32(0.2)*float32(gridCount)) + float32(10.0)
	)
	for range gridCount {
		x := -float32(0.2) * float32(gridCount)
		for range gridCount {
			bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &polygon)

			x += float32(0.4)
		}

		y += float32(0.4)
	}
}

func CreateWasher(worldID *box2d.World) {
	var (
		kinematic = true
		groundID  box2d.Body
	)
	{
		bodyDef := box2d.DefaultBodyDef()

		groundID = worldID.CreateBody(&bodyDef)
	}
	{
		var (
			motorSpeed = float32(25.0)
			bodyDef    = box2d.DefaultBodyDef()
		)

		bodyDef.Position = box2d.Vec2{X: float32(0.0), Y: float32(10.0)}.Pos()
		if kinematic {
			bodyDef.Type = box2d.KinematicBody
			bodyDef.AngularVelocity = (box2d.Pi / float32(180.0)) * motorSpeed
			bodyDef.LinearVelocity = box2d.Vec2{X: float32(0.001), Y: -float32(0.002)}
		} else {
			bodyDef.Type = box2d.DynamicBody
		}

		var (
			bodyID   = worldID.CreateBody(&bodyDef)
			shapeDef = box2d.DefaultShapeDef()
			r0       = float32(14.0)
			r1       = float32(16.0)
			r2       = float32(18.0)
			angle    = box2d.Pi / float32(18.0)
		)

		sine, cosine := boundedSinCos(angle)

		q := box2d.Rot{C: cosine, S: sine}

		sine, cosine = boundedSinCos(float32(0.1) * angle)

		var (
			qo = box2d.Rot{C: cosine, S: sine}
			u1 = box2d.Vec2{X: float32(1.0), Y: float32(0.0)}
		)

		for i := range 36 {
			var u2 box2d.Vec2
			if i == 35 {
				u2 = box2d.Vec2{X: float32(1.0), Y: float32(0.0)}
			} else {
				u2 = q.Rotate(u1)
			}
			{
				var (
					a1      = qo.InvRotate(u1)
					a2      = qo.Rotate(u2)
					p1      = a1.Scale(r1)
					p2      = a1.Scale(r2)
					p3      = a2.Scale(r1)
					p4      = a2.Scale(r2)
					points  = [4]box2d.Vec2{p1, p2, p3, p4}
					hull    = box2d.ComputeHull(points[:])
					polygon = box2d.MakePolygon(&hull, float32(0.0))
				)
				bodyID.CreatePolygonShape(&shapeDef, &polygon)
			}

			if i%9 == 0 {
				var (
					p1      = u1.Scale(r0)
					p2      = u1.Scale(r1)
					p3      = u2.Scale(r0)
					p4      = u2.Scale(r1)
					points  = [4]box2d.Vec2{p1, p2, p3, p4}
					hull    = box2d.ComputeHull(points[:])
					polygon = box2d.MakePolygon(&hull, float32(0.0))
				)
				bodyID.CreatePolygonShape(&shapeDef, &polygon)
			}

			u1 = u2
		}

		if !kinematic {
			jointDef := box2d.DefaultRevoluteJointDef()

			jointDef.Base.BodyA = groundID
			jointDef.Base.BodyB = bodyID
			jointDef.Base.LocalFrameA.P = box2d.Vec2{X: float32(0.0), Y: float32(10.0)}
			jointDef.Base.LocalFrameB.P = box2d.Vec2{X: float32(0.0), Y: float32(0.0)}
			jointDef.MotorSpeed = (box2d.Pi / float32(180.0)) * motorSpeed
			jointDef.MaxMotorTorque = float32(1e8)
			jointDef.EnableMotor = true
			worldID.CreateRevoluteJoint(&jointDef)
		}
	}

	var (
		gridCount = 90
		a         = float32(0.1)
		polygon   = box2d.MakeSquare(a)
		bodyDef   = box2d.DefaultBodyDef()
	)

	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()

	shapeDef.EnableHitEvents = true

	y := float32(-float32(1.1)*a*float32(gridCount)) + float32(10.0)
	for range gridCount {
		x := -float32(1.1) * a * float32(gridCount)
		for range gridCount {
			bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &polygon)

			x += float32(float32(2.1) * a)
		}

		y += float32(float32(2.1) * a)
	}
}

func CreateJunkyard(worldID *box2d.World) {
	{
		var (
			gridSize = float32(1.0)
			bodyDef  = box2d.DefaultBodyDef()
			groundID = worldID.CreateBody(&bodyDef)
			shapeDef = box2d.DefaultShapeDef()
			y        = float32(0.0)
			x        = -float32(80.0) * gridSize
		)
		for range 161 {
			box := box2d.MakeOffsetBox(
				float32(0.55)*gridSize,
				float32(0.5)*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.Rot{C: 1},
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			x += gridSize
		}

		y = gridSize

		x = -float32(80.0) * gridSize
		for range 50 {
			box := box2d.MakeOffsetBox(
				float32(0.5)*gridSize,
				float32(0.55)*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.Rot{C: 1},
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			y += gridSize
		}

		y = gridSize

		x = float32(80.0) * gridSize
		for range 50 {
			box := box2d.MakeOffsetBox(
				float32(0.5)*gridSize,
				float32(0.55)*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.Rot{C: 1},
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			y += gridSize
		}
	}

	var (
		columnCount = 200
		rowCount    = 40
		radius      = float32(0.25)
		polygon     box2d.Polygon
	)
	{
		var (
			phi    = box2d.Pi * (float32(math.Sqrt(float64(float32(5.0)))) - float32(1.0))
			points [5]box2d.Vec2
		)

		for i := range 5 {
			var (
				theta = phi * float32(i)
				cs    = box2d.ComputeCosSin(theta)
			)

			points[i].X = radius * cs.Cosine
			points[i].Y = radius * cs.Sine
		}

		hull := box2d.ComputeHull(points[:])

		polygon = box2d.MakePolygon(&hull, float32(0.0))
	}

	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Type = box2d.DynamicBody

	var (
		shapeDef = box2d.DefaultShapeDef()
		side     = -float32(0.1)
		yStart   = float32(15.0)
	)

	for i := range columnCount {
		x := float32(1.5) * (float32(float32(2.0)*float32(i)) - float32(columnCount)) * radius
		for j := range rowCount {
			y := float32(float32(4.0)*float32(j)*radius) + yStart

			bodyDef.Position = box2d.Vec2{X: x + side, Y: y}.Pos()
			side = -side

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &polygon)
		}
	}

	bodyDef.Type = box2d.KinematicBody
	bodyDef.Position = box2d.Pos{}
	gJunkyardData.pusherID = worldID.CreateBody(&bodyDef)

	box := box2d.MakeOffsetBox(
		float32(2.0),
		float32(4.0),
		box2d.Vec2{X: float32(0.0), Y: float32(4.0)},
		box2d.Rot{C: 1},
	)
	gJunkyardData.pusherID.CreatePolygonShape(&shapeDef, &box)
}

func StepJunkyard(worldID *box2d.World, stepCount int) float32 {
	var (
		timeStep = float32(1.0) / float32(60.0)
		time     = timeStep * float32(stepCount)
		cs       = box2d.ComputeCosSin(float32(0.2) * time)
		target   = box2d.WorldTransform{
			P: box2d.Vec2{X: float32(60.0) * cs.Sine, Y: float32(0.0)}.Pos(),
			Q: box2d.Rot{C: 1},
		}
	)
	gJunkyardData.pusherID.SetTargetTransform(target, timeStep, true)

	return float32(0.0)
}

func CreateSleep(worldID *box2d.World) {
	gSleepData = sleepData{}

	var (
		baseCount = 60
		extent    = float32(0.5)
		baseWidth = float32(2.0) * extent * float32(baseCount)
		pitch     = baseWidth + float32(float32(8.0)*extent)
		span      = pitch * (float32(sleepPyramidCount) - 1)
	)
	{
		bodyDef := box2d.DefaultBodyDef()

		bodyDef.Position = box2d.Vec2{X: float32(0.0), Y: -float32(1.0)}.Pos()

		var (
			groundID = worldID.CreateBody(&bodyDef)
			box      = box2d.MakeBox(float32(float32(0.5)*span)+baseWidth, float32(1.0))
			shapeDef = box2d.DefaultShapeDef()
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	for i := range sleepPyramidCount {
		centerX := float32(
			-float32(0.5)*span,
		) + float32(
			float32(i)*pitch,
		) - float32(
			float32(0.5)*baseWidth,
		)
		createSleepPyramid(
			worldID,
			baseCount,
			extent,
			centerX,
			&gSleepData.bodyIDA[i],
			&gSleepData.bodyIDB[i],
		)
	}
}

func StepSleep(worldID *box2d.World, stepCount int) float32 {
	for i := range sleepPyramidCount {
		if !gSleepData.bodyIDA[i].Awake() {
			jointDef := box2d.DefaultFilterJointDef()

			jointDef.Base.BodyA = gSleepData.bodyIDA[i]
			jointDef.Base.BodyB = gSleepData.bodyIDB[i]

			jointID := worldID.CreateFilterJoint(&jointDef).Joint
			jointID.Destroy()

			break
		}
	}

	return float32(0.0)
}

func CreateCompounds(worldID *box2d.World) {
	{
		var (
			gridSize = float32(1.0)
			bodyDef  = box2d.DefaultBodyDef()
			groundID = worldID.CreateBody(&bodyDef)
			shapeDef = box2d.DefaultShapeDef()
			y        = float32(0.0)
			x        = -float32(40.0) * gridSize
		)
		for range 81 {
			box := box2d.MakeOffsetBox(
				float32(0.55)*gridSize,
				float32(0.5)*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.Rot{C: 1},
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			x += gridSize
		}

		y = gridSize

		x = -float32(40.0) * gridSize
		for range 100 {
			box := box2d.MakeOffsetBox(
				float32(0.5)*gridSize,
				float32(0.55)*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.Rot{C: 1},
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			y += gridSize
		}

		y = gridSize

		x = float32(40.0) * gridSize
		for range 100 {
			box := box2d.MakeOffsetBox(
				float32(0.5)*gridSize,
				float32(0.55)*gridSize,
				box2d.Vec2{X: x, Y: y},
				box2d.Rot{C: 1},
			)
			groundID.CreatePolygonShape(&shapeDef, &box)

			y += gridSize
		}

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -float32(800.0), Y: -float32(80.0)},
			Point2: box2d.Vec2{X: float32(800.0), Y: -float32(80.0)},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	var (
		columnCount = 20
		rowCount    = 150
		leftPoints  = [3]box2d.Vec2{
			{X: -float32(1.0), Y: float32(0.0)},
			{X: float32(0.5), Y: float32(1.0)},
			{X: float32(0.0), Y: float32(2.0)},
		}
		leftHull    = box2d.ComputeHull(leftPoints[:])
		left        = box2d.MakePolygon(&leftHull, float32(0.0))
		rightPoints = [3]box2d.Vec2{
			{X: float32(1.0), Y: float32(0.0)},
			{X: -float32(0.5), Y: float32(1.0)},
			{X: float32(0.0), Y: float32(2.0)},
		}
		rightHull = box2d.ComputeHull(rightPoints[:])
		right     = box2d.MakePolygon(&rightHull, float32(0.0))
		bodyDef   = box2d.DefaultBodyDef()
	)

	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()

	shapeDef.Density = float32(1.0)
	shapeDef.Material.Friction = float32(0.5)

	var (
		shift   = float32(2.0)
		extray  = float32(0.25)
		side    = float32(0.25)
		centerx = shift*float32(columnCount)/float32(2.0) - float32(1.0)
		centery = float32(1.15) / float32(2.0)
		yStart  = float32(5.0)
	)

	for i := range columnCount {
		x := float32(float32(i)*shift) - centerx
		for j := range rowCount {
			y := float32(float32(j)*(shift+extray)) + centery + yStart

			bodyDef.Position = box2d.Vec2{X: x + side, Y: y}.Pos()
			side = -side

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &left)
			bodyID.CreatePolygonShape(&shapeDef, &right)
		}
	}
}

func CreateQueries(worldID *box2d.World) {
	gQueryRandomState = 1234

	var (
		extent    = float32(500.0)
		cellCount = 500
		fill      = float32(0.1)
		pileX     = float32(0.4) * extent
		pileY     = float32(0.3) * extent
		clearance = box2d.AABB{
			LowerBound: box2d.Vec2{X: pileX - float32(10.0), Y: pileY - float32(10.0)},
			UpperBound: box2d.Vec2{X: pileX + float32(50.0), Y: pileY + float32(60.0)},
		}
		bodyDef  = box2d.DefaultBodyDef()
		shapeDef = box2d.DefaultShapeDef()
	)

	createQueryObstacles(worldID, cellCount, fill, clearance, &bodyDef, &shapeDef)

	shapeDef.Filter.CategoryBits = 1
	{
		bodyDef.Position = box2d.Vec2{X: float32(0.0), Y: float32(0.0)}.Pos()

		var (
			groundID = worldID.CreateBody(&bodyDef)
			segment  = box2d.Segment{
				Point1: box2d.Vec2{X: pileX - float32(10.0), Y: pileY},
				Point2: box2d.Vec2{X: pileX + float32(50.0), Y: pileY},
			}
		)
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}
	{
		sensorDef := box2d.DefaultShapeDef()

		sensorDef.IsSensor = true

		sensorBox := box2d.MakeSquare(float32(2.0))

		for i := range 8 {
			bodyDef.Position = box2d.Vec2{
				X: pileX + float32(float32(4.0)*float32(i)),
				Y: pileY + float32(2.0),
			}.Pos()

			sensorID := worldID.CreateBody(&bodyDef)
			sensorID.CreatePolygonShape(&sensorDef, &sensorBox)
		}
	}
	{
		var (
			columnCount = 40
			rowCount    = 50
			spacing     = float32(0.55)
			box         = box2d.MakeSquare(float32(0.25))
		)

		bodyDef.Type = box2d.DynamicBody

		for i := range rowCount {
			for j := range columnCount {
				bodyDef.Position = box2d.Vec2{
					X: pileX + float32(spacing*float32(j)),
					Y: pileY + float32(1.0) + float32(spacing*float32(i)),
				}.Pos()

				bodyID := worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &box)
			}
		}
	}
	{
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Vec2{X: pileX - float32(5.0), Y: pileY + float32(1.2)}.Pos()
		bodyDef.LinearVelocity = box2d.Vec2{X: float32(5.0), Y: float32(0.0)}
		gQueryBenchmark.paddleID = worldID.CreateBody(&bodyDef)

		paddle := box2d.MakeBox(float32(1.0), float32(1.0))
		gQueryBenchmark.paddleID.CreatePolygonShape(&shapeDef, &paddle)
		gQueryBenchmark.paddleMinX = pileX - float32(5.0)
		gQueryBenchmark.paddleMaxX = pileX + float32(30.0)
	}

	queryTotal := queryCount
	for i := range queryTotal {
		var (
			x1 = queryRandom(float32(0.0), extent)
			y1 = queryRandom(float32(0.0), extent)
			x2 = queryRandom(float32(0.0), extent)
			y2 = queryRandom(float32(0.0), extent)
		)

		gQueryBenchmark.origins[i] = box2d.Vec2{X: x1, Y: y1}.Pos()
		gQueryBenchmark.translations[i] = box2d.Vec2{X: x2 - x1, Y: y2 - y1}
	}

	gQueryBenchmark.stats = box2d.TreeStats{}
	gQueryBenchmark.extent = extent
}

func StepQueries(worldID *box2d.World, stepCount int) float32 {
	var (
		data           = &gQueryBenchmark
		paddlePosition = data.paddleID.Position()
	)
	if paddlePosition.X > (box2d.Vec2{X: data.paddleMaxX}).Pos().X {
		data.paddleID.SetLinearVelocity(box2d.Vec2{X: -float32(5.0), Y: float32(0.0)})
	} else if paddlePosition.X < (box2d.Vec2{X: data.paddleMinX}).Pos().X {
		data.paddleID.SetLinearVelocity(box2d.Vec2{X: float32(5.0), Y: float32(0.0)})
	}

	filters := [2]box2d.QueryFilter{box2d.DefaultQueryFilter(), box2d.DefaultQueryFilter()}

	filters[1].MaskBits = 1

	var (
		circleCenter = box2d.Vec2{}
		circle       = box2d.MakeProxy([]box2d.Vec2{circleCenter}, float32(0.1))
	)

	queryTotal := queryCount

	var (
		hitCount = 0
		stats    = box2d.TreeStats{}
	)

	for f := range 2 {
		filter := filters[f]
		for i := range queryTotal {
			result := worldID.CastRayClosest(data.origins[i], data.translations[i], filter)

			stats.NodeVisits += result.NodeVisits

			stats.LeafVisits += result.LeafVisits
			if result.Hit {
				hitCount++
			}
		}

		for i := range queryTotal {
			var (
				fraction  = float32(1.0)
				castStats = worldID.CastShape(
					data.origins[i],
					&circle,
					data.translations[i],
					filter,
					func(id box2d.Shape, point box2d.Pos, normal box2d.Vec2, value float32) float32 {
						return queryCastClosest(id, point, normal, value, &fraction)
					},
				)
			)

			stats.NodeVisits += castStats.NodeVisits
			stats.LeafVisits += castStats.LeafVisits

			if fraction < float32(1.0) {
				hitCount++
			}
		}

		for i := range queryTotal {
			var (
				overlapCount = 0
				aabb         = box2d.AABB{
					LowerBound: box2d.Vec2{X: -float32(5.0), Y: -float32(5.0)},
					UpperBound: box2d.Vec2{X: float32(5.0), Y: float32(5.0)},
				}
				overlapStats = worldID.OverlapAABB(
					data.origins[i],
					aabb,
					filter,
					func(id box2d.Shape) bool { return queryOverlapCount(&overlapCount) },
				)
			)

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

func DestroyTreeCast() {
	data := &gTreeCast

	stats := data.stats

	*data = treeCastBenchmark{}
	data.stats = stats
}

func CreateTreeCast(worldID *box2d.World) {
	DestroyTreeCast()

	data := &gTreeCast

	gQueryRandomState = 1234

	var (
		extent    = float32(500.0)
		cellCount = 500
		fill      = float32(0.1)
		pileX     = float32(0.4) * extent
		pileY     = float32(0.3) * extent
		clearance = box2d.AABB{
			LowerBound: box2d.Vec2{X: pileX - float32(10.0), Y: pileY - float32(10.0)},
			UpperBound: box2d.Vec2{X: pileX + float32(50.0), Y: pileY + float32(60.0)},
		}
	)

	data.proxyCapacity = cellCount * cellCount / 5
	data.tightBoxes = make([]box2d.AABB, data.proxyCapacity)
	data.tree = box2d.NewDynamicTree(data.proxyCapacity)
	data.created = true

	createTreeCastObstacles(cellCount, fill, clearance)

	treeCastAddBox(box2d.Vec2{X: pileX + float32(20.0), Y: pileY}, float32(30.0), float32(0.0), 1)

	for i := range 8 {
		treeCastAddBox(
			box2d.Vec2{X: pileX + float32(float32(4.0)*float32(i)), Y: pileY + float32(2.0)},
			float32(2.0),
			float32(2.0),
			1,
		)
	}

	lengths := [treeCastLengthCount]float32{float32(0.0), float32(25.0), float32(5.0)}

	for i := range treeCastRayCount {
		var (
			x1 = queryRandom(float32(0.0), extent)
			y1 = queryRandom(float32(0.0), extent)
			x2 = queryRandom(float32(0.0), extent)
			y2 = queryRandom(float32(0.0), extent)
			t  = box2d.Vec2{X: x2 - x1, Y: y2 - y1}
		)

		data.origins[i] = box2d.Vec2{X: x1, Y: y1}
		data.translations[0][i] = t

		length := t.Length()

		for k := 1; k < treeCastLengthCount; k++ {
			var scale float32
			if length > float32(0.0) {
				scale = lengths[k] / length
			} else {
				scale = float32(0.0)
			}

			data.translations[k][i] = t.Scale(scale)
		}
	}

	data.stats = box2d.TreeStats{}
}

func StepTreeCast(worldID *box2d.World, stepCount int) float32 {
	var (
		data     = &gTreeCast
		tree     = &data.tree
		masks    = [2]uint64{box2d.DefaultMaskBits, 1}
		rayCount = treeCastRayCount
		stats    = box2d.TreeStats{}
		hitCount = 0
	)

	for k := range treeCastLengthCount {
		translations := data.translations[k]

		for m := range 2 {
			for i := range rayCount {
				input := box2d.RayCastInput{
					Origin:      data.origins[i],
					Translation: translations[i],
					MaxFraction: float32(1.0),
				}

				context := treeCastContext{tightBoxes: data.tightBoxes, fraction: 1}

				castStats := tree.CastRay(
					&input,
					masks[m],
					func(input *box2d.RayCastInput, id int, value uint64) float32 {
						return treeCastRayCallback(input, value, &context)
					},
				)

				stats.NodeVisits += castStats.NodeVisits
				stats.LeafVisits += castStats.LeafVisits

				if context.fraction < float32(1.0) {
					hitCount++
				}
			}
		}

		for i := range rayCount {
			var (
				o   = data.origins[i]
				box = box2d.AABB{
					LowerBound: box2d.Vec2{X: o.X - treeCastBoxExtent, Y: o.Y - treeCastBoxExtent},
					UpperBound: box2d.Vec2{X: o.X + treeCastBoxExtent, Y: o.Y + treeCastBoxExtent},
				}
				input = box2d.BoxCastInput{
					Box:         box,
					Translation: translations[i],
					MaxFraction: float32(1.0),
				}
			)

			context := treeCastContext{tightBoxes: data.tightBoxes, fraction: 1}

			castStats := tree.CastBox(
				&input,
				box2d.DefaultMaskBits,
				func(input *box2d.BoxCastInput, id int, value uint64) float32 {
					return treeCastBoxCallback(input, value, &context)
				},
			)

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

func DestroyTileWorld() {
	var (
		data  = &gTileWorld
		stats = data.stats
	)

	*data = tileWorldBenchmark{}
	data.stats = stats
}

func CreateTileWorld(worldID *box2d.World) {
	DestroyTileWorld()

	var (
		data      = &gTileWorld
		gridCount = int((tileCycleCount * tilePeriod / tileGridSize))
	)

	data.groundCount = gridCount / tileColumnsPerBody
	data.groundIDs = make([]box2d.Body, data.groundCount)
	data.bodySpan = tileColumnsPerBody * tileGridSize
	data.xLeft = -float32(0.5) * (tileCycleCount * tilePeriod)
	data.xNext = data.xLeft

	data.groundHead = 0
	for i := range data.groundCount {
		data.groundIDs[i] = createTileGround(worldID, data.xNext)
		data.xNext += data.bodySpan
	}

	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Type = box2d.DynamicBody

	var (
		shapeDef = box2d.DefaultShapeDef()
		box      = box2d.MakeBox(float32(0.3), float32(0.2))
	)

	for cycleIndex := tileCycleCount / 5; cycleIndex < tileCycleCount; cycleIndex += 12 {
		xBase := float32((float32(0.5)+float32(cycleIndex))*tilePeriod) + data.xLeft

		bodyDef.Position.X = box2d.Vec2{X: xBase - float32(3.0)}.Pos().X
		for range 10 {
			bodyDef.Position.Y = box2d.Vec2{Y: float32(10.0)}.Pos().Y
			for range 5 {
				bodyID := worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &box)

				bodyDef.Position.Y += 0.5
			}

			bodyDef.Position.X += box2d.Vec2{X: float32(0.6)}.Pos().X
		}
	}

	data.stats = box2d.TreeStats{}
}

func StepTileWorld(worldID *box2d.World, stepCount int) float32 {
	data := &gTileWorld
	data.groundIDs[data.groundHead].Destroy()
	data.groundIDs[data.groundHead] = createTileGround(worldID, data.xNext)
	data.groundHead = (data.groundHead + 1) % data.groundCount
	data.xNext += data.bodySpan
	data.xLeft += data.bodySpan

	var (
		state  = uint32(1234) + uint32(7919)*uint32(stepCount)
		span   = float32(data.groundCount) * data.bodySpan
		filter = box2d.DefaultQueryFilter()
		aabb   = box2d.AABB{
			LowerBound: box2d.Vec2{X: -float32(5.0), Y: -float32(5.0)},
			UpperBound: box2d.Vec2{X: float32(5.0), Y: float32(5.0)},
		}
		stats    = box2d.TreeStats{}
		hitCount = 0
	)

	for range tileQueryCount {
		var (
			origin = box2d.Vec2{
				X: data.xLeft + tileRandom(&state, float32(0.0), span),
				Y: tileRandom(&state, -float32(2.0), float32(18.0)),
			}.Pos()
			angle = tileRandom(&state, float32(0.0), float32(2.0)*box2d.Pi)
		)

		sine, cosine := boundedSinCos(angle)

		var (
			translation = box2d.Vec2{X: tileRayLength * cosine, Y: tileRayLength * sine}
			result      = worldID.CastRayClosest(origin, translation, filter)
		)

		stats.NodeVisits += result.NodeVisits

		stats.LeafVisits += result.LeafVisits
		if result.Hit {
			hitCount++
		}

		var (
			overlapCount = 0
			overlapStats = worldID.OverlapAABB(
				origin,
				aabb,
				filter,
				func(id box2d.Shape) bool { return tileOverlapCount(&overlapCount) },
			)
		)

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

func createQueryObstacles(
	worldID *box2d.World,
	cellCount int,
	fill float32,
	clearance box2d.AABB,
	bodyDef *box2d.BodyDef,
	shapeDef *box2d.ShapeDef,
) {
	for i := range cellCount {
		y := float32(i)

		for j := range cellCount {
			var (
				x           = float32(j)
				fillTest    = queryRandom(float32(0.0), float32(1.0))
				ratio       = queryRandom(float32(1.0), float32(5.0))
				halfWidth   = queryRandom(float32(0.05), float32(0.25))
				orientation = queryRandom(float32(0.0), float32(1.0))
				category    = int(queryRandom(float32(0.0), float32(2.999)))
			)

			if fillTest > fill {
				continue
			}

			if clearance.LowerBound.X <= x && x <= clearance.UpperBound.X &&
				clearance.LowerBound.Y <= y &&
				y <= clearance.UpperBound.Y {
				continue
			}

			bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

			var (
				bodyID = worldID.CreateBody(bodyDef)
				box    box2d.Polygon
			)
			if orientation > float32(0.5) {
				box = box2d.MakeBox(ratio*halfWidth, halfWidth)
			} else {
				box = box2d.MakeBox(halfWidth, ratio*halfWidth)
			}

			shapeDef.Filter.CategoryBits = uint64(1) << category
			bodyID.CreatePolygonShape(shapeDef, &box)
		}
	}
}

func createTreeCastObstacles(cellCount int, fill float32, clearance box2d.AABB) {
	for i := range cellCount {
		y := float32(i)

		for j := range cellCount {
			var (
				x           = float32(j)
				fillTest    = queryRandom(float32(0.0), float32(1.0))
				ratio       = queryRandom(float32(1.0), float32(5.0))
				halfWidth   = queryRandom(float32(0.05), float32(0.25))
				orientation = queryRandom(float32(0.0), float32(1.0))
				category    = int(queryRandom(float32(0.0), float32(2.999)))
			)

			if fillTest > fill {
				continue
			}

			if clearance.LowerBound.X <= x && x <= clearance.UpperBound.X &&
				clearance.LowerBound.Y <= y &&
				y <= clearance.UpperBound.Y {
				continue
			}

			var hx, hy float32
			if orientation > float32(0.5) {
				hx, hy = ratio*halfWidth, halfWidth
			} else {
				hx, hy = halfWidth, ratio*halfWidth
			}

			treeCastAddBox(box2d.Vec2{X: x, Y: y}, hx, hy, uint64(1)<<category)
		}
	}
}

func createSmallPyramid(
	worldID *box2d.World,
	baseCount int,
	extent float32,
	centerX float32,
	baseY float32,
) {
	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Type = box2d.DynamicBody

	var (
		shapeDef = box2d.DefaultShapeDef()
		box      = box2d.MakeSquare(extent)
	)
	for i := range baseCount {
		y := float32((float32(float32(2.0)*float32(i))+float32(1.0))*extent) + baseY
		for j := i; j < baseCount; j++ {
			x := float32(
				(float32(i)+float32(1.0))*extent,
			) + float32(
				float32(2.0)*(float32(j)-float32(i))*extent,
			) + centerX - float32(
				0.5,
			)

			bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
}

func createSleepPyramid(
	worldID *box2d.World,
	baseCount int,
	extent float32, centerX float32,
	bodyIDA *box2d.Body, bodyIDB *box2d.Body,
) {
	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Type = box2d.DynamicBody
	bodyDef.SleepThreshold = float32(1.0)

	var (
		shapeDef = box2d.DefaultShapeDef()
		box      = box2d.MakeSquare(extent)
	)
	for i := range baseCount {
		y := (float32(float32(2.0)*float32(i)) + float32(1.0)) * extent
		for j := i; j < baseCount; j++ {
			x := float32(
				(float32(i)+float32(1.0))*extent,
			) + float32(
				float32(2.0)*(float32(j)-float32(i))*extent,
			) + centerX

			bodyDef.Position = box2d.Vec2{X: x, Y: y}.Pos()

			bodyID := worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)

			if i == 0 && j == 0 {
				*bodyIDA = bodyID
			} else if i == 1 && j == 1 {
				*bodyIDB = bodyID
			}
		}
	}
}

func queryRandom(lower float32, upper float32) float32 {
	gQueryRandomState = uint32(1664525)*gQueryRandomState + uint32(1013904223)

	unit := float32((gQueryRandomState >> 8)) * (float32(1.0) / float32(16777216.0))

	return lower + float32((upper-lower)*unit)
}

func queryCastClosest(
	shapeID box2d.Shape,
	point box2d.Pos,
	normal box2d.Vec2,
	fraction float32,
	result *float32,
) float32 {
	*result = fraction
	return fraction
}

func queryOverlapCount(count *int) bool {
	*count += 1
	return true
}

func treeCastRayBox(box box2d.AABB, p box2d.Vec2, d box2d.Vec2, maxFraction float32) float32 {
	tmin, tmax := float32(0.0), maxFraction
	if !treeCastRaySlab(box.LowerBound.X, box.UpperBound.X, p.X, d.X, &tmin, &tmax) {
		return -float32(1.0)
	}

	if !treeCastRaySlab(box.LowerBound.Y, box.UpperBound.Y, p.Y, d.Y, &tmin, &tmax) {
		return -float32(1.0)
	}

	if tmin <= float32(0.0) {
		return -float32(1.0)
	}

	return tmin
}

func treeCastRaySlab(
	lower float32,
	upper float32,
	p float32,
	d float32,
	tmin *float32,
	tmax *float32,
) bool {
	if d == float32(0.0) {
		return !(p < lower || upper < p)
	}

	inv := float32(1.0) / d
	t1, t2 := (lower-p)*inv, (upper-p)*inv

	if t1 > t2 {
		t1, t2 = t2, t1
	}

	if !(*tmin > t1) {
		*tmin = t1
	}

	if !(*tmax < t2) {
		*tmax = t2
	}

	return !(*tmin > *tmax)
}

func treeCastRayCallback(
	input *box2d.RayCastInput,
	userData uint64,
	context *treeCastContext,
) float32 {
	var (
		cast     = context
		fraction = treeCastRayBox(
			cast.tightBoxes[userData],
			input.Origin,
			input.Translation,
			input.MaxFraction,
		)
	)
	if fraction < float32(0.0) {
		return -float32(1.0)
	}

	cast.fraction = fraction

	return fraction
}

func treeCastBoxCallback(
	input *box2d.BoxCastInput,
	userData uint64,
	context *treeCastContext,
) float32 {
	var (
		cast    = context
		tight   = cast.tightBoxes[userData]
		extents = input.Box.Extents()
		grown   = box2d.AABB{
			LowerBound: tight.LowerBound.Sub(extents),
			UpperBound: tight.UpperBound.Add(extents),
		}
		fraction = treeCastRayBox(grown, input.Box.Center(), input.Translation, input.MaxFraction)
	)
	if fraction < float32(0.0) {
		return -float32(1.0)
	}

	cast.fraction = fraction

	return fraction
}

func treeCastAddBox(center box2d.Vec2, hx float32, hy float32, categoryBits uint64) {
	data := &gTreeCast
	if data.proxyCount >= data.proxyCapacity {
		panic(&box2d.AssertError{Condition: "data->proxyCount < data->proxyCapacity"})
	}

	var (
		tight = box2d.AABB{
			LowerBound: box2d.Vec2{X: center.X - hx, Y: center.Y - hy},
			UpperBound: box2d.Vec2{X: center.X + hx, Y: center.Y + hy},
		}
		fat = box2d.AABB{
			LowerBound: box2d.Vec2{
				X: tight.LowerBound.X - float32(2*box2d.SpeculativeDistance()),
				Y: tight.LowerBound.Y - float32(2*box2d.SpeculativeDistance()),
			},
			UpperBound: box2d.Vec2{
				X: tight.UpperBound.X + float32(2*box2d.SpeculativeDistance()),
				Y: tight.UpperBound.Y + float32(2*box2d.SpeculativeDistance()),
			},
		}
		index = data.proxyCount
	)

	data.tightBoxes[index] = tight
	data.proxyCount += 1
	data.tree.CreateProxy(fat, categoryBits, uint64(index))
}

func tileRandom(state *uint32, lower float32, upper float32) float32 {
	*state = uint32(1664525)**state + uint32(1013904223)

	unit := float32((*state >> 8)) * (float32(1.0) / float32(16777216.0))

	return lower + float32((upper-lower)*unit)
}

func createTileGround(worldID *box2d.World, xBody float32) box2d.Body {
	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Position.X = box2d.Vec2{X: xBody}.Pos().X

	var (
		groundID = worldID.CreateBody(&bodyDef)
		shapeDef = box2d.DefaultShapeDef()
	)

	shapeDef.InvokeContactCreation = false

	var (
		omega  = float32(2.0) * box2d.Pi / tilePeriod
		xShape = float32(0.0)
	)
	for range tileColumnsPerBody {
		var (
			columnCount = int(
				math.Round(float64(float32(4.0)*float32(math.Cos(float64(omega*(xBody+xShape)))))),
			) + 12
			y = float32(0.0)
		)
		for range columnCount {
			square := box2d.MakeOffsetBox(
				float32(0.4)*tileGridSize,
				float32(0.4)*tileGridSize,
				box2d.Vec2{X: xShape, Y: y},
				box2d.Rot{C: 1},
			)

			square.Radius = float32(0.1)
			groundID.CreatePolygonShape(&shapeDef, &square)

			y += tileGridSize
		}

		xShape += tileGridSize
	}

	return groundID
}

func tileOverlapCount(count *int) bool {
	*count += 1
	return true
}
