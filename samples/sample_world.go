// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const (
	farRagdollsColumnCount = 5
	farRagdollsRowCount    = 5
	farSceneOrigin         = 10.0e6
	kilometersPerMeter     = 0.001
	farGateMotorSpeedLimit = 0.3
	farGateMotorSpeedStep  = 0.01
)

type TileWorld struct {
	Sample

	car          Car //exhaustruct:optional
	viewPosition box2d.Pos
	period       float32
	cycleCount   int
	cycleIndex   int
	gridCount    float32
	gridSize     float32
	speed        float32

	explosionPosition box2d.Pos
	explode           bool
	followCar         bool
}

// A pyramid built far from the origin to exercise double precision world positions. The contact
// solver runs in delta space, so the stack settles the same as one at the origin while the body
// transforms hold their full double coordinate. Record it (R) and open it in the Replay Viewer:
// the recorded doubles survive the snapshot and the motion reproduces with no divergence.
type FarPyramid struct {
	Sample
}

type FarRagdolls struct {
	Sample

	humans     [farRagdollsColumnCount * farRagdollsRowCount]scenes.Human //exhaustruct:optional
	humanCount int                                                        //exhaustruct:optional
}

// The Gear Lift mechanism from the Joints samples, rebuilt 1000 km from the origin. This sample is
// not guarded, so it compiles and renders in both precision modes. The debug draw subtracts the
// camera center before narrowing to float, so the view is correct either way. Build double and the
// gears mesh cleanly out here. Build single and the float grid (printed on screen) is coarser than
// the gear teeth, so the same mechanism grinds and jitters. One scene, two builds, a side by side.
type FarGate struct {
	Sample

	origin      box2d.Pos
	driverID    box2d.RevoluteJoint //exhaustruct:optional
	motorTorque float32
	motorSpeed  float32
	enableMotor bool
}

func CreateTileWorld(context *SampleContext) Sampler {
	const (
		initialViewHeight = 15.0
		initialZoom       = 25.0
		sceneCycleKinds   = 3
		boxStackOffset    = 3.0
		boxStackHeight    = 10.0
		boxHalfWidth      = 0.3
		boxHalfHeight     = 0.2
		ragdollOffset     = 2.0
		ragdollScale      = 1.5
		ragdollFriction   = 0.05
		donutOffset       = 4.0
		ragdollHeight     = 10.0
		donutHeight       = 12.0
		donutScale        = 0.75
		carOffset         = 20.0
		carHeight         = 40.0
		carScale          = 10.0
		carHertz          = 2.0
		carDampingRatio   = 0.7
		carTorque         = 2000.0
		explosionHeight   = 7.0
		tilePeriod        = 40.0
		cycleMidpoint     = 0.5
	)

	period := float32(tilePeriod)
	omega := 2.0 * box2d.Pi / period //nolint:mnd // Two pi is one full angular cycle.

	cycleCount := 600
	if isDebug {
		cycleCount = 10
	}

	gridSize := float32(1.0)
	gridCount := float32(int(float32(cycleCount) * period / gridSize))
	xStart := -0.5 * (float32(cycleCount) * period)
	cycleIndex := 0
	explosionPosition := box2d.Pos{
		X: posScalar(
			(cycleMidpoint+float32(cycleIndex))*period + xStart,
		),
		Y: explosionHeight,
	}
	s := &TileWorld{
		context:           context,
		camera:            &context.camera,
		draw:              context.draw,
		screenTextX:       initialScreenTextX,
		mouseForceScale:   initialMouseForceScale,
		period:            period,
		cycleCount:        cycleCount,
		cycleIndex:        cycleIndex,
		gridSize:          gridSize,
		gridCount:         gridCount,
		viewPosition:      box2d.Pos{X: posScalar(xStart), Y: initialViewHeight},
		explosionPosition: explosionPosition,
		speed:             0,
		explode:           true,
		followCar:         false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = s.viewPosition
		s.context.camera.zoom = initialZoom * 1.0
		s.context.debugDraw.DrawJoints = false
	}

	s.createGround(xStart, omega)

	humanIndex := 0

	for cycleIndex := range s.cycleCount {
		xbase := (0.5+float32(cycleIndex))*s.period + xStart //nolint:mnd // One half selects the cycle midpoint.

		remainder := cycleIndex % sceneCycleKinds
		switch remainder {
		case 0:
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(xbase - boxStackOffset), Y: boxStackHeight}

			shapeDef := box2d.DefaultShapeDef()
			box := box2d.MakeBox(boxHalfWidth, boxHalfHeight)

			for range 10 {
				bodyDef.Position.Y = boxStackHeight
				for range 5 {
					bodyID := s.worldID.CreateBody(&bodyDef)
					bodyID.CreatePolygonShape(&shapeDef, &box)

					bodyDef.Position.Y += 0.5
				}

				bodyDef.Position.X += 0.6
			}
		case 1:
			position := box2d.Pos{X: posScalar(xbase - ragdollOffset), Y: ragdollHeight}

			for range 5 {
				human := scenes.Human{
					Bones:          [scenes.BoneCount]scenes.Bone{},
					FrictionTorque: 0,
					OriginalScale:  0,
					Scale:          0,
					IsSpawned:      false,
				}
				scenes.CreateHuman(
					&human,
					s.worldID,
					position,
					ragdollScale,
					ragdollFriction,
					0.0,
					0.0,
					humanIndex+1,
					nil,
					false,
				)
				humanIndex++
				position.X += 1.0
			}
		default:
			position := box2d.Pos{X: posScalar(xbase - donutOffset), Y: donutHeight}

			for range 5 {
				var donut Donut
				donut.Create(s.worldID, position, donutScale, 0, false, nil)
				position.X += 2.0
			}
		}
	}

	s.car.Spawn(
		s.worldID,
		box2d.Pos{X: posScalar(xStart + carOffset), Y: carHeight},
		carScale,
		carHertz,
		carDampingRatio,
		carTorque,
	)

	return s
}

func (s *TileWorld) DrawControls() bool {
	const (
		controlWidth       = 6.0
		maxViewSpeed       = 400.0
		metersPerKilometer = 1000.0
	)
	imgui.PushItemWidth(controlWidth * imgui.FontSize())
	imgui.SliderFloatV("speed", &s.speed, -maxViewSpeed, maxViewSpeed, "%.0f", 0)
	imgui.PopItemWidth()

	if imgui.Button("stop") {
		s.speed = 0.0
	}

	imgui.Checkbox("explode", &s.explode)
	imgui.Checkbox("follow car", &s.followCar)

	imTextf("world size = %.6g kilometers", s.gridSize*s.gridCount/metersPerKilometer)

	return true
}

func (s *TileWorld) Step() {
	const (
		explosionRadius = 2.0
		carForwardSpeed = 20.0
		carReverseSpeed = -5.0
		cycleMidpoint   = 0.5
	)

	span := 0.5 * (s.period * float32(s.cycleCount)) //nolint:mnd // One half converts the full world width to its span.

	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}

	if s.context.pause {
		timeStep = 0.0
	}

	s.viewPosition.X += posScalar(timeStep * s.speed)
	s.viewPosition.X = posScalar(clampFloat(float32(s.viewPosition.X), -span, span))

	if s.speed != 0.0 {
		s.context.camera.center = s.viewPosition
	}

	if s.followCar {
		s.context.camera.center.X = s.car.chassisID.Position().X
	}

	radius := float32(explosionRadius)

	if s.stepCount&0x1 == 0x1 && s.explode {
		s.explosionPosition.X = posScalar(
			(cycleMidpoint+float32(s.cycleIndex))*s.period - span,
		)

		def := box2d.DefaultExplosionDef()
		def.Position = s.explosionPosition
		def.Radius = radius
		def.Falloff = 0.1
		def.ImpulsePerLength = 1.0
		s.worldID.Explode(&def)

		s.cycleIndex = (s.cycleIndex + 1) % s.cycleCount
	}

	if s.explode {
		s.draw.DrawCircle(s.explosionPosition, radius, box2d.ColorAzure)
	}

	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.car.SetSpeed(carForwardSpeed)
	}

	if s.context.window.GetKey(glfw.KeyS) == glfw.Press {
		s.car.SetSpeed(0.0)
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.car.SetSpeed(carReverseSpeed)
	}

	s.Sample.Step()
}

func (s *TileWorld) createGround(xStart float32, omega float32) {
	const (
		groundWaveHeight = 4.0
		groundBaseTiles  = 12
		tileHalfExtent   = 0.4
	)
	{
		bodyDef := box2d.DefaultBodyDef()
		shapeDef := box2d.DefaultShapeDef()

		// Setting this to false significantly reduces the cost of creating
		// static bodies and shapes.
		shapeDef.InvokeContactCreation = false

		height := float32(groundWaveHeight)
		xBody := xStart
		xShape := xStart

		var groundID box2d.Body

		for i := range int(s.gridCount) {
			// Create a new body regularly so that shapes are not too far from the body origin.
			// Most algorithms in Box2D work in local coordinates, but contact points are computed
			// relative to the body origin.
			// This makes a noticeable improvement in stability far from the origin.
			if i%10 == 0 {
				bodyDef.Position.X = posScalar(xBody)
				groundID = s.worldID.CreateBody(&bodyDef)
				xShape = 0.0
			}

			y := float32(0.0)

			ycount := int(float32(math.Round(float64(height*cosf(omega*xBody))))) + groundBaseTiles

			for range ycount {
				square := box2d.MakeOffsetBox(
					tileHalfExtent*s.gridSize,
					tileHalfExtent*s.gridSize,
					box2d.Vec2{X: xShape, Y: y},
					box2d.RotIdentity,
				)
				square.Radius = 0.1
				groundID.CreatePolygonShape(&shapeDef, &square)

				y += s.gridSize
			}

			xBody += s.gridSize
			xShape += s.gridSize
		}
	}
}

func CreateFarPyramid(context *SampleContext) Sampler {
	const (
		viewHeight      = 12.0
		boxHalfSize     = 0.25
		groundHalfWidth = 40.0
	)

	s := &FarPyramid{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	// 1e7 is exactly representable in float, so integer box offsets stay exact and a run here
	// can be compared against one at the origin.
	origin := box2d.Pos{X: farSceneOrigin, Y: 0.0}

	if !s.context.restart {
		s.context.camera.center = origin.Add(box2d.Vec2{X: 0.0, Y: viewHeight})
		s.context.camera.zoom = 17.0
	}

	h := float32(boxHalfSize)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = origin
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
		Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	box := box2d.MakeBox(h, h)
	bodyDef.Type = box2d.DynamicBody

	baseCount := 50
	for i := range baseCount {
		y := (2.0*float32(i) + 1.0) * h
		for j := i; j < baseCount; j++ {
			x := (float32(i)+1.0)*h + 2.0*float32(j-i)*h - h*float32(baseCount)
			bodyDef.Position = origin.Add(box2d.Vec2{X: x, Y: y})
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	return s
}

func (s *FarPyramid) Step() {
	s.Sample.Step()

	if box2d.IsDoublePrecision() {
		s.DrawScreenTextLinef("Double precision")
	} else {
		s.DrawScreenTextLinef("Single precision")
	}

	c := s.context.camera.center
	s.DrawScreenTextLinef("view center %.0f km from world origin", kilometersPerMeter*float32(c.X))
}

func CreateFarRagdolls(context *SampleContext) Sampler {
	const (
		viewHeight         = 6.0
		containerHalfWidth = 6.0
		containerHeight    = 12.0
		spawnJitter        = 0.3
		firstRowHeight     = 2.0
		jointFriction      = 0.05
		restitution        = 0.5
	)

	s := &FarRagdolls{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	origin := box2d.Pos{X: farSceneOrigin, Y: 0.0}

	if !s.context.restart {
		s.context.camera.center = origin.Add(box2d.Vec2{X: 0.0, Y: viewHeight})
		s.context.camera.zoom = 10.0
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = origin
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	w := float32(containerHalfWidth)
	h := float32(containerHeight)
	floor := box2d.Segment{Point1: box2d.Vec2{X: -w, Y: 0.0}, Point2: box2d.Vec2{X: w, Y: 0.0}}
	leftWall := box2d.Segment{Point1: box2d.Vec2{X: -w, Y: 0.0}, Point2: box2d.Vec2{X: -w, Y: h}}
	rightWall := box2d.Segment{Point1: box2d.Vec2{X: w, Y: 0.0}, Point2: box2d.Vec2{X: w, Y: h}}

	groundID.CreateSegmentShape(&shapeDef, &floor)
	groundID.CreateSegmentShape(&shapeDef, &leftWall)
	groundID.CreateSegmentShape(&shapeDef, &rightWall)

	scale := float32(1.0)
	index := 0

	for i := range farRagdollsRowCount {
		for j := range farRagdollsColumnCount {
			x := 2.4*scale*(float32(j)-0.5*(farRagdollsColumnCount-1)) + scenes.RandomFloatRange(
				-spawnJitter,
				spawnJitter,
			)
			y := firstRowHeight + 2.2*scale*float32(i)
			p := origin.Add(box2d.Vec2{X: x, Y: y})
			scenes.CreateHuman(
				&s.humans[index],
				s.worldID,
				p,
				scale,
				jointFriction,
				1.0,
				restitution,
				index+1,
				nil,
				false,
			)
			index++
		}
	}

	s.humanCount = index

	return s
}

func (s *FarRagdolls) Step() {
	s.Sample.Step()

	if box2d.IsDoublePrecision() {
		s.DrawScreenTextLinef("Double precision")
	} else {
		s.DrawScreenTextLinef("Single precision")
	}

	c := s.context.camera.center
	s.DrawScreenTextLinef(
		"%d ragdolls piled %.0f km from the world origin",
		s.humanCount,
		kilometersPerMeter*float32(c.X),
	)
}

func CreateFarGate(context *SampleContext) Sampler {
	const (
		originDistance      = 1.0e6
		viewHeight          = 6.0
		pathScale           = 0.2
		gearToothHalfWidth  = 0.09
		gearToothHalfHeight = 0.06
		gearToothRadius     = 0.03
		chainLinkHalfLength = 0.07
		chainLinkRadius     = 0.05
		chainLinkCount      = 40
		gateHalfHeight      = 1.5
		driverHeight        = 9.75
		gearSpacingX        = 2.0
		followerAngle       = 0.25
		followerUpperAngle  = 0.8
		initialMotorTorque  = 80.0
	)

	s := &FarGate{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		origin:          box2d.Pos{X: originDistance, Y: 0.0},
		motorTorque:     initialMotorTorque,
		motorSpeed:      -farGateMotorSpeedLimit,
		enableMotor:     true,
	}
	s.init(context, s, true)

	// 1e6 is exactly representable in float so the placement adds no error of its own. Any
	// difference between the two builds comes from how positions are stored as the sim runs.
	origin := s.origin

	if !s.context.restart {
		s.context.camera.center = origin.Add(box2d.Vec2{X: 0.0, Y: viewHeight})
		s.context.camera.zoom = 7.0
		s.context.debugDraw.DrawJoints = false
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = origin
		groundID = s.worldID.CreateBody(&bodyDef)

		var points [128]box2d.Vec2

		offset := box2d.Vec2{X: -120.0, Y: -200.0}
		scale := float32(pathScale)
		count := ParsePath(gearLiftPath, offset, points[:64], scale)

		material := box2d.DefaultSurfaceMaterial()
		material.CustomColor = uint32(box2d.ColorDarkSeaGreen)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}

		groundID.CreateChain(&chainDef)
	}

	gearRadius := float32(1.0)
	toothHalfWidth := float32(gearToothHalfWidth)
	toothHalfHeight := float32(gearToothHalfHeight)
	toothRadius := float32(gearToothRadius)
	linkHalfLength := float32(chainLinkHalfLength)
	linkRadius := float32(chainLinkRadius)
	linkCount := float32(chainLinkCount)
	doorHalfHeight := float32(gateHalfHeight)

	gearPosition1 := origin.Add(box2d.Vec2{X: -4.25, Y: driverHeight})
	gearPosition2 := gearPosition1.Add(box2d.Vec2{X: gearSpacingX, Y: 1.0})
	linkAttachPosition := gearPosition2.Add(
		box2d.Vec2{X: gearRadius + 2.0*toothHalfWidth + toothRadius, Y: 0.0},
	)
	doorPosition := linkAttachPosition.Add(
		box2d.Vec2{X: 0.0, Y: 2.0*linkCount*linkHalfLength + doorHalfHeight}.Neg(),
	)

	{
		position := gearPosition1
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = position

		bodyID := s.worldID.CreateBody(&bodyDef)

		createGear(bodyID, toothHalfHeight)

		revoluteDef := box2d.DefaultRevoluteJointDef()

		revoluteDef.Base.BodyA = groundID
		revoluteDef.Base.BodyB = bodyID
		revoluteDef.Base.LocalFrameA.P = groundID.LocalPoint(position)
		revoluteDef.Base.LocalFrameB.P = box2d.Vec2Zero
		revoluteDef.EnableMotor = s.enableMotor
		revoluteDef.MaxMotorTorque = s.motorTorque
		revoluteDef.MotorSpeed = s.motorSpeed
		s.driverID = s.worldID.CreateRevoluteJoint(&revoluteDef)
	}

	var followerID box2d.Body

	{
		position := gearPosition2
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = position

		followerID = s.worldID.CreateBody(&bodyDef)

		createGear(followerID, toothHalfWidth)

		revoluteDef := box2d.DefaultRevoluteJointDef()

		revoluteDef.Base.BodyA = groundID
		revoluteDef.Base.BodyB = followerID
		revoluteDef.Base.LocalFrameA.P = groundID.LocalPoint(position)
		revoluteDef.Base.LocalFrameA.Q = box2d.MakeRot(followerAngle * box2d.Pi)
		revoluteDef.Base.LocalFrameB.P = box2d.Vec2Zero
		revoluteDef.EnableMotor = true
		revoluteDef.MaxMotorTorque = 0.5
		revoluteDef.LowerAngle = -0.3 * box2d.Pi
		revoluteDef.UpperAngle = followerUpperAngle * box2d.Pi
		revoluteDef.EnableLimit = true
		s.worldID.CreateRevoluteJoint(&revoluteDef)
	}

	createLiftChainAndDoor(s.worldID, groundID, followerID, linkAttachPosition, doorPosition,
		linkHalfLength, linkRadius, doorHalfHeight)

	s.createDebris(origin)

	return s
}

func (s *FarGate) DrawControls() bool {
	const (
		controlWidth   = 6.0
		maxMotorTorque = 100.0
	)

	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.driverID.SetMotorEnabled(s.enableMotor)
		s.driverID.WakeBodies()
	}

	imgui.PushItemWidth(controlWidth * imgui.FontSize())

	if imgui.SliderFloatV("Max Torque", &s.motorTorque, 0.0, maxMotorTorque, "%.0f", 0) {
		s.driverID.SetMaxMotorTorque(s.motorTorque)
		s.driverID.WakeBodies()
	}

	if imgui.SliderFloatV(
		"Speed",
		&s.motorSpeed,
		-farGateMotorSpeedLimit,
		farGateMotorSpeedLimit,
		"%.2f",
		0,
	) {
		s.driverID.SetMotorSpeed(s.motorSpeed)
		s.driverID.WakeBodies()
	}

	imgui.PopItemWidth()

	return true
}

func (s *FarGate) Step() {
	const metersPerKilometer = 1000.0

	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.motorSpeed = max(-farGateMotorSpeedLimit, s.motorSpeed-farGateMotorSpeedStep)
		s.driverID.SetMotorSpeed(s.motorSpeed)
		s.driverID.WakeBodies()
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.motorSpeed = min(farGateMotorSpeedLimit, s.motorSpeed+farGateMotorSpeedStep)
		s.driverID.SetMotorSpeed(s.motorSpeed)
		s.driverID.WakeBodies()
	}

	s.Sample.Step()

	fx := float32(s.origin.X)
	gridStep := math.Nextafter32(fx, math.MaxFloat32) - fx

	if box2d.IsDoublePrecision() {
		s.DrawScreenTextLinef("Double precision")
	} else {
		s.DrawScreenTextLinef("Single precision")
	}

	s.DrawScreenTextLinef("%.0f km from the origin", float64(s.origin.X)/metersPerKilometer)
	s.DrawScreenTextLinef("float grid step here = %.6g m, gear teeth are about 0.12 m", gridStep)
}

func (s *FarGate) createDebris(origin box2d.Pos) {
	const (
		firstRowHeight   = 4.25
		polygonExtent    = 0.1
		minPolygonRadius = 0.01
		maxPolygonRadius = 0.02
	)
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.RollingResistance = 0.3

		colors := [5]box2d.HexColor{
			box2d.ColorGray,
			box2d.ColorGainsboro,
			box2d.ColorLightGray,
			box2d.ColorLightSlateGray,
			box2d.ColorDarkGray,
		}

		y := float32(firstRowHeight)

		xCount, yCount := 10, 20
		for range yCount {
			x := float32(-3.15)
			for range xCount {
				bodyDef.Position = origin.Add(box2d.Vec2{X: x, Y: y})
				bodyID := s.worldID.CreateBody(&bodyDef)
				poly := scenes.RandomPolygon(polygonExtent)
				poly.Radius = scenes.RandomFloatRange(minPolygonRadius, maxPolygonRadius)

				colorIndex := scenes.RandomIntRange(0, len(colors)-1)
				shapeDef.Material.CustomColor = uint32(colors[colorIndex])

				bodyID.CreatePolygonShape(&shapeDef, &poly)

				x += 0.2
			}

			y += 0.2
		}
	}
}

func registerWorldSamples() {
	RegisterSample("World", "Tiles", CreateTileWorld)
	RegisterSample("World", "Far Pyramid", CreateFarPyramid)
	RegisterSample("World", "Far Ragdolls", CreateFarRagdolls)
	RegisterSample("World", "Far Gate", CreateFarGate)
}
