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

type TileWorld struct {
	Sample

	car          Car
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

func CreateTileWorld(context *SampleContext) Sampler {
	s := &TileWorld{}
	s.init(context, s, true)

	s.period = 40.0
	omega := 2.0 * box2d.Pi / s.period
	s.cycleCount = 600
	if isDebug {
		s.cycleCount = 10
	}
	s.gridSize = 1.0
	s.gridCount = float32(int(float32(s.cycleCount) * s.period / s.gridSize))

	xStart := -0.5 * (float32(s.cycleCount) * s.period)

	s.viewPosition = box2d.Pos{X: posScalar(xStart), Y: 15.0}

	if !s.context.restart {
		s.context.camera.center = s.viewPosition
		s.context.camera.zoom = 25.0 * 1.0
		s.context.debugDraw.DrawJoints = false
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		shapeDef := box2d.DefaultShapeDef()

		// Setting this to false significantly reduces the cost of creating
		// static bodies and shapes.
		shapeDef.InvokeContactCreation = false

		height := float32(4.0)
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

			ycount := int(float32(math.Round(float64(height*cosf(omega*xBody))))) + 12

			for range ycount {
				square := box2d.MakeOffsetBox(
					0.4*s.gridSize,
					0.4*s.gridSize,
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

	humanIndex := 0
	for cycleIndex := range s.cycleCount {
		xbase := (0.5+float32(cycleIndex))*s.period + xStart

		remainder := cycleIndex % 3
		if remainder == 0 {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(xbase - 3.0), Y: 10.0}

			shapeDef := box2d.DefaultShapeDef()
			box := box2d.MakeBox(0.3, 0.2)

			for range 10 {
				bodyDef.Position.Y = 10.0
				for range 5 {
					bodyID := s.worldID.CreateBody(&bodyDef)
					bodyID.CreatePolygonShape(&shapeDef, &box)
					bodyDef.Position.Y += 0.5
				}
				bodyDef.Position.X += 0.6
			}
		} else if remainder == 1 {
			position := box2d.Pos{X: posScalar(xbase - 2.0), Y: 10.0}
			for range 5 {
				human := scenes.Human{}
				scenes.CreateHuman(
					&human,
					s.worldID,
					position,
					1.5,
					0.05,
					0.0,
					0.0,
					humanIndex+1,
					nil,
					false,
				)
				humanIndex++
				position.X += 1.0
			}
		} else {
			position := box2d.Pos{X: posScalar(xbase - 4.0), Y: 12.0}

			for range 5 {
				var donut Donut
				donut.Create(s.worldID, position, 0.75, 0, false, nil)
				position.X += 2.0
			}
		}
	}

	s.car.Spawn(
		s.worldID,
		box2d.Pos{X: posScalar(xStart + 20.0), Y: 40.0},
		10.0,
		2.0,
		0.7,
		2000.0,
		nil,
	)

	s.cycleIndex = 0
	s.speed = 0.0
	s.explosionPosition = box2d.Pos{
		X: posScalar((0.5+float32(s.cycleIndex))*s.period + xStart),
		Y: 7.0,
	}
	s.explode = true
	s.followCar = false
	return s
}

func (s *TileWorld) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderFloatV("speed", &s.speed, -400.0, 400.0, "%.0f", 0)
	imgui.PopItemWidth()
	if imgui.Button("stop") {
		s.speed = 0.0
	}

	imgui.Checkbox("explode", &s.explode)
	imgui.Checkbox("follow car", &s.followCar)

	imText("world size = %.6g kilometers", s.gridSize*s.gridCount/1000.0)

	return true
}

func (s *TileWorld) Step() {
	span := 0.5 * (s.period * float32(s.cycleCount))
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

	radius := float32(2.0)
	if s.stepCount&0x1 == 0x1 && s.explode {
		s.explosionPosition.X = posScalar((0.5+float32(s.cycleIndex))*s.period - span)

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
		s.car.SetSpeed(20.0)
	}

	if s.context.window.GetKey(glfw.KeyS) == glfw.Press {
		s.car.SetSpeed(0.0)
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.car.SetSpeed(-5.0)
	}

	s.Sample.Step()
}

// A pyramid built far from the origin to exercise double precision world positions. The contact
// solver runs in delta space, so the stack settles the same as one at the origin while the body
// transforms hold their full double coordinate. Record it (R) and open it in the Replay Viewer:
// the recorded doubles survive the snapshot and the motion reproduces with no divergence.
type FarPyramid struct {
	Sample
}

func CreateFarPyramid(context *SampleContext) Sampler {
	s := &FarPyramid{}
	s.init(context, s, true)

	// 1e7 is exactly representable in float, so integer box offsets stay exact and a run here
	// can be compared against one at the origin.
	origin := box2d.Pos{X: 10.0e6, Y: 0.0}

	if !s.context.restart {
		s.context.camera.center = origin.Add(box2d.Vec2{X: 0.0, Y: 12.0})
		s.context.camera.zoom = 17.0
	}

	h := float32(0.25)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = origin
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -40.0, Y: 0.0},
		Point2: box2d.Vec2{X: 40.0, Y: 0.0},
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
		s.DrawScreenTextLine("Double precision")
	} else {
		s.DrawScreenTextLine("Single precision")
	}

	c := s.context.camera.center
	s.DrawScreenTextLine("view center %.0f km from world origin", 0.001*float32(c.X))
}

const (
	farRagdollsColumnCount = 5
	farRagdollsRowCount    = 5
)

type FarRagdolls struct {
	Sample

	humans     [farRagdollsColumnCount * farRagdollsRowCount]scenes.Human
	humanCount int
}

func CreateFarRagdolls(context *SampleContext) Sampler {
	s := &FarRagdolls{}
	s.init(context, s, true)

	origin := box2d.Pos{X: 10.0e6, Y: 0.0}

	if !s.context.restart {
		s.context.camera.center = origin.Add(box2d.Vec2{X: 0.0, Y: 6.0})
		s.context.camera.zoom = 10.0
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = origin
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	w := float32(6.0)
	h := float32(12.0)
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
				-0.3,
				0.3,
			)
			y := 2.0 + 2.2*scale*float32(i)
			p := origin.Add(box2d.Vec2{X: x, Y: y})
			scenes.CreateHuman(
				&s.humans[index],
				s.worldID,
				p,
				scale,
				0.05,
				1.0,
				0.5,
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
		s.DrawScreenTextLine("Double precision")
	} else {
		s.DrawScreenTextLine("Single precision")
	}

	c := s.context.camera.center
	s.DrawScreenTextLine(
		"%d ragdolls piled %.0f km from the world origin",
		s.humanCount,
		0.001*float32(c.X),
	)
}

// The Gear Lift mechanism from the Joints samples, rebuilt 1000 km from the origin. This sample is
// not guarded, so it compiles and renders in both precision modes. The debug draw subtracts the
// camera center before narrowing to float, so the view is correct either way. Build double and the
// gears mesh cleanly out here. Build single and the float grid (printed on screen) is coarser than
// the gear teeth, so the same mechanism grinds and jitters. One scene, two builds, a side by side.
type FarGate struct {
	Sample

	origin      box2d.Pos
	driverID    box2d.RevoluteJoint
	motorTorque float32
	motorSpeed  float32
	enableMotor bool
}

func CreateFarGate(context *SampleContext) Sampler {
	s := &FarGate{}
	s.init(context, s, true)

	// 1e6 is exactly representable in float so the placement adds no error of its own. Any
	// difference between the two builds comes from how positions are stored as the sim runs.
	origin := box2d.Pos{X: 1.0e6, Y: 0.0}
	s.origin = origin

	if !s.context.restart {
		s.context.camera.center = origin.Add(box2d.Vec2{X: 0.0, Y: 6.0})
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
		scale := float32(0.2)
		count := ParsePath(gearLiftPath, offset, points[:64], scale, false)

		material := box2d.DefaultSurfaceMaterial()
		material.CustomColor = uint32(box2d.ColorDarkSeaGreen)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}

		groundID.CreateChain(&chainDef)
	}

	gearRadius := float32(1.0)
	toothHalfWidth := float32(0.09)
	toothHalfHeight := float32(0.06)
	toothRadius := float32(0.03)
	linkHalfLength := float32(0.07)
	linkRadius := float32(0.05)
	linkCount := float32(40)
	doorHalfHeight := float32(1.5)

	gearPosition1 := origin.Add(box2d.Vec2{X: -4.25, Y: 9.75})
	gearPosition2 := gearPosition1.Add(box2d.Vec2{X: 2.0, Y: 1.0})
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

		createGear(
			bodyID,
			gearRadius,
			toothHalfHeight,
			toothHalfWidth,
			toothHalfHeight,
			toothRadius,
		)

		revoluteDef := box2d.DefaultRevoluteJointDef()

		s.motorTorque = 80.0
		s.motorSpeed = -0.3
		s.enableMotor = true

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

		createGear(
			followerID,
			gearRadius,
			toothHalfWidth,
			toothHalfWidth,
			toothHalfHeight,
			toothRadius,
		)

		revoluteDef := box2d.DefaultRevoluteJointDef()

		revoluteDef.Base.BodyA = groundID
		revoluteDef.Base.BodyB = followerID
		revoluteDef.Base.LocalFrameA.P = groundID.LocalPoint(position)
		revoluteDef.Base.LocalFrameA.Q = box2d.MakeRot(0.25 * box2d.Pi)
		revoluteDef.Base.LocalFrameB.P = box2d.Vec2Zero
		revoluteDef.EnableMotor = true
		revoluteDef.MaxMotorTorque = 0.5
		revoluteDef.LowerAngle = -0.3 * box2d.Pi
		revoluteDef.UpperAngle = 0.8 * box2d.Pi
		revoluteDef.EnableLimit = true
		s.worldID.CreateRevoluteJoint(&revoluteDef)
	}

	var lastLinkID box2d.Body
	{
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -linkHalfLength},
			Center2: box2d.Vec2{X: 0.0, Y: linkHalfLength},
			Radius:  linkRadius,
		}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0
		shapeDef.Material.CustomColor = uint32(box2d.ColorLightSteelBlue)

		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.MaxMotorTorque = 0.05
		jointDef.EnableMotor = true

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		position := linkAttachPosition.Add(box2d.Vec2{X: 0.0, Y: -linkHalfLength})

		count := 40
		prevBodyID := followerID
		for range count {
			bodyDef.Position = position

			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreateCapsuleShape(&shapeDef, &capsule)

			pivot := box2d.Pos{X: position.X, Y: position.Y + posScalar(linkHalfLength)}
			jointDef.Base.BodyA = prevBodyID
			jointDef.Base.BodyB = bodyID
			jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
			jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
			jointDef.Base.DrawScale = 0.2
			s.worldID.CreateRevoluteJoint(&jointDef)

			position.Y -= posScalar(2.0 * linkHalfLength)
			prevBodyID = bodyID
		}

		lastLinkID = prevBodyID
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = doorPosition

		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(0.15, doorHalfHeight)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.1
		shapeDef.Material.CustomColor = uint32(box2d.ColorDarkCyan)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		{
			pivot := doorPosition.Add(box2d.Vec2{X: 0.0, Y: doorHalfHeight})
			revoluteDef := box2d.DefaultRevoluteJointDef()
			revoluteDef.Base.BodyA = lastLinkID
			revoluteDef.Base.BodyB = bodyID
			revoluteDef.Base.LocalFrameA.P = lastLinkID.LocalPoint(pivot)
			revoluteDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: doorHalfHeight}
			revoluteDef.EnableMotor = true
			revoluteDef.MaxMotorTorque = 0.05
			s.worldID.CreateRevoluteJoint(&revoluteDef)
		}

		{
			localAxis := box2d.Vec2{X: 0.0, Y: 1.0}
			jointDef := box2d.DefaultPrismaticJointDef()
			jointDef.Base.BodyA = groundID
			jointDef.Base.BodyB = bodyID
			jointDef.Base.LocalFrameA.P = groundID.LocalPoint(doorPosition)
			jointDef.Base.LocalFrameA.Q = box2d.MakeRotFromUnitVector(localAxis)
			jointDef.Base.LocalFrameB.P = box2d.Vec2Zero
			jointDef.Base.LocalFrameB.Q = box2d.MakeRotFromUnitVector(localAxis)
			jointDef.MaxMotorForce = 0.2
			jointDef.EnableMotor = true
			jointDef.Base.CollideConnected = true
			s.worldID.CreatePrismaticJoint(&jointDef)
		}
	}

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

		y := float32(4.25)
		xCount, yCount := 10, 20
		for range yCount {
			x := float32(-3.15)
			for range xCount {
				bodyDef.Position = origin.Add(box2d.Vec2{X: x, Y: y})
				bodyID := s.worldID.CreateBody(&bodyDef)
				poly := scenes.RandomPolygon(0.1)
				poly.Radius = scenes.RandomFloatRange(0.01, 0.02)

				colorIndex := scenes.RandomIntRange(0, 4)
				shapeDef.Material.CustomColor = uint32(colors[colorIndex])

				bodyID.CreatePolygonShape(&shapeDef, &poly)
				x += 0.2
			}

			y += 0.2
		}
	}
	return s
}

func (s *FarGate) DrawControls() bool {
	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.driverID.SetMotorEnabled(s.enableMotor)
		s.driverID.WakeBodies()
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Max Torque", &s.motorTorque, 0.0, 100.0, "%.0f", 0) {
		s.driverID.SetMaxMotorTorque(s.motorTorque)
		s.driverID.WakeBodies()
	}

	if imgui.SliderFloatV("Speed", &s.motorSpeed, -0.3, 0.3, "%.2f", 0) {
		s.driverID.SetMotorSpeed(s.motorSpeed)
		s.driverID.WakeBodies()
	}

	imgui.PopItemWidth()

	return true
}

func (s *FarGate) Step() {
	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.motorSpeed = max(-0.3, s.motorSpeed-0.01)
		s.driverID.SetMotorSpeed(s.motorSpeed)
		s.driverID.WakeBodies()
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.motorSpeed = min(0.3, s.motorSpeed+0.01)
		s.driverID.SetMotorSpeed(s.motorSpeed)
		s.driverID.WakeBodies()
	}

	s.Sample.Step()

	fx := float32(s.origin.X)
	gridStep := math.Nextafter32(fx, math.MaxFloat32) - fx

	if box2d.IsDoublePrecision() {
		s.DrawScreenTextLine("Double precision")
	} else {
		s.DrawScreenTextLine("Single precision")
	}

	s.DrawScreenTextLine("%.0f km from the origin", float64(s.origin.X)/1000.0)
	s.DrawScreenTextLine("float grid step here = %.6g m, gear teeth are about 0.12 m", gridStep)
}

func init() {
	RegisterSample("World", "Tiles", CreateTileWorld)
	RegisterSample("World", "Far Pyramid", CreateFarPyramid)
	RegisterSample("World", "Far Ragdolls", CreateFarRagdolls)
	RegisterSample("World", "Far Gate", CreateFarGate)
}
