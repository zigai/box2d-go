// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"sync/atomic"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
)

const (
	characterStaticBit  uint64 = 0x0001
	characterMoverBit   uint64 = 0x0002
	characterDynamicBit uint64 = 0x0004

	characterAllBits uint64 = 0xFFFFFFFF

	characterCameraX            = 20.0
	characterCameraY            = 9.0
	characterSpawnY             = 8.0
	characterCapsuleHalfLength  = 0.5
	characterCapsuleRadius      = 0.3
	characterPathScale          = float32(0.2)
	characterSecondGroundX      = 98.0
	characterBallX              = 7.0
	characterBallBaseY          = 7.0
	characterBallRadius         = 0.25
	characterElevatorHalfWidth  = 2.0
	characterElevatorHalfHeight = 0.1
	characterStackHalfWidth     = float32(0.25)
	characterPendulumX          = 160.0
	characterJumpSpeedLimit     = 40.0
	characterMaxSpeedLimit      = 20.0
	characterStopSpeedLimit     = 10.0
	characterAccelerationLimit  = 100.0
	characterFrictionLimit      = 10.0
	characterGravityLimit       = 100.0
	characterGravityScaleLimit  = 4.0
	characterPogoHertzLimit     = 30.0
	characterPogoDampingLimit   = 4.0
	characterElevatorX          = 112.0
	characterElevatorY          = 10.0
	characterElevatorTravel     = float32(4.0)
)

const characterGroundPath1 = "M -34.395834,201.08333 H 293.68751 v -47.625 h -2.64584 l " +
	"-10.58333,7.9375 -13.22916,7.9375 -13.24648,5.29167 " +
	"-31.73269,7.9375 -21.16667,2.64583 -23.8125,10.58333 -15.875,2e-5 v -3.96875 l -17.19792,2e-5 v 2.64582 l " +
	"-17.19791,0 v -1.32292 h -3.96875 v -1.32292 h -3.96875 v -1.32292 h -3.96875 v -1.32291 h -3.96875 v -1.32292 " +
	"h -3.96875 v -1.32292 h -3.96875 v -1.32291 l -3.968754,0 v -1.32292 h -3.96875 v -1.32292 h -3.968751 v " +
	"-1.32291 h -3.96875 v -1.32292 h -3.96875 v -1.32292 h -3.96875 v -1.32291 h -3.96875 v -1.32292 l -3.96875,0 v " +
	"-1.32292 h -3.968751 v -1.32291 h -3.96875 l -2e-6,-1.32294 H 52.916669 L 39.6875,177.27083 h -5.291667 l " +
	"-7.937499,5.29167 H 15.875001 l -47.625002,-50.27083 v -26.45834 h -2.645834 l 10e-7,95.25"

const characterDynamicGroundPath1 = "M -34.395834,201.08333 H 293.68751 v -47.625 h -2.64584 l " +
	"-10.58333,7.9375 -13.22916,7.9375 -13.24648,5.29167 " +
	"-31.73269,7.9375 -21.16667,2.64583 -23.8125,13.22919 -15.875,2e-5 0,-6.61461 -17.19792,2e-5 v 2.64582 h " +
	"-17.19791 v -1.32292 h -3.96875 v -1.32292 h -3.96875 v -1.32292 h -3.96875 v -1.32291 h -3.96875 v -1.32292 h " +
	"-3.96875 v -1.32292 h -3.96875 v -1.32291 h -3.968754 v -1.32292 h -3.96875 v -1.32292 h -3.968751 v -1.32291 h " +
	"-3.96875 v -1.32292 h -3.96875 v -1.32292 h -3.96875 v -1.32291 h -3.96875 v -1.32292 h -3.96875 v -1.32292 h " +
	"-3.968751 v -1.32291 h -3.96875 l -2e-6,-1.32294 H 52.916669 L 39.6875,177.27083 h -5.291667 l " +
	"-7.937499,5.29167 H 15.875001 l -47.625002,-50.27083 v -26.45834 h -2.645834 l 10e-7,95.25"

const characterGroundPath2 = "M 2.6458333,201.08333 H 399.52085 v -23.8125 h -23.8125 l " +
	"21.16667,18.52084 -232.83335,-1e-5 -0.0575,2.64584 h " +
	"-5.29166 v -2.64583 l -7.86855,-1e-5 -0.0114,-2.64583 h -2.64583 l -2.64583,2.64584 h -7.9375 l " +
	"-2.64584,2.64583 -2.58891,-2.64584 -10.64031,2e-5 v -2.64583 l -7.9375,0 v -2.64584 l -7.9375,0 v -2.64583 l " +
	"-7.9375,0 v -2.64583 l -7.9375,0 v -2.64583 l -7.937501,-1e-5 v -2.64584 l -7.9375,1e-5 v -2.64583 l " +
	"-7.937501,-2e-5 V 174.625 l -7.937501,1e-5 v -2.64584 l -5.291669,0 -7.9375,-2.64584 -7.9375,-2.64583 " +
	"-5.291667,-5.29167 H 21.166667 L 13.229167,158.75 5.2916668,153.45833 H 2.6458334 l -10e-8,47.625"

var (
	characterElevatorBase      = box2d.Vec2{X: characterElevatorX, Y: characterElevatorY}
	characterElevatorAmplitude = characterElevatorTravel
)

type GeometryMoverSample struct {
	Sample

	mover         GeometricMover
	elevatorID    box2d.Body         //exhaustruct:optional
	friendlyShape MoverShapeUserData //exhaustruct:optional
	elevatorShape MoverShapeUserData //exhaustruct:optional
	time          float32            //exhaustruct:optional
	jumpReleased  bool               //exhaustruct:optional
	lockCamera    bool               //exhaustruct:optional
}
type DynamicMoverSample struct {
	Sample

	mover        DynamicMover
	elevatorID   box2d.Body //exhaustruct:optional
	time         float32    //exhaustruct:optional
	preSolve     bool       //exhaustruct:optional
	jumpReleased bool       //exhaustruct:optional
	lockCamera   bool       //exhaustruct:optional
}

// This shows how to mitigate ghost collisions using the presolve callback.
// Based on: https://briansemrau.github.io/dealing-with-ghost-collisions/
type GhostCollisionSample struct {
	Sample

	playerIDs     [2]box2d.Body //exhaustruct:optional
	rejectedCount atomic.Int32  //exhaustruct:optional
	speed         float32       //exhaustruct:optional
}

func CreateGeometryMoverSample(context *SampleContext) Sampler {
	s := &GeometryMoverSample{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		mover:           NewGeometricMover(),
	}
	s.init(context, s, true)

	if !context.restart {
		s.camera.center = box2d.Pos{X: characterCameraX, Y: characterCameraY}
		s.camera.zoom = 10.0
	}

	context.debugDraw.DrawJoints = false

	moverDef := DefaultGeometricMoverDef()
	moverDef.position = box2d.Pos{X: 0.0, Y: characterSpawnY}
	moverDef.capsule = box2d.Capsule{
		Center1: box2d.Vec2{X: 0.0, Y: -characterCapsuleHalfLength},
		Center2: box2d.Vec2{X: 0.0, Y: characterCapsuleHalfLength},
		Radius:  characterCapsuleRadius,
	}
	moverDef.filter = box2d.QueryFilter{
		CategoryBits: characterMoverBit,
		MaskBits:     characterStaticBit | characterDynamicBit | characterMoverBit,
	}
	s.mover.Create(s.worldID, &moverDef)

	var groundID1 box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "svg"
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID1 = s.worldID.CreateBody(&bodyDef)

		var points [64]box2d.Vec2

		offset := box2d.Vec2{X: -50.0, Y: -200.0}
		scale := characterPathScale

		count := ParsePath(characterGroundPath1, offset, points[:], scale)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true

		groundID1.CreateChain(&chainDef)
	}

	var groundID2 box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: characterSecondGroundX, Y: 0.0}
		groundID2 = s.worldID.CreateBody(&bodyDef)

		var points [64]box2d.Vec2

		offset := box2d.Vec2{X: 0.0, Y: -200.0}
		scale := characterPathScale

		count := ParsePath(characterGroundPath2, offset, points[:], scale)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true

		groundID2.CreateChain(&chainDef)
	}

	{
		const (
			friendlyX = 32.0
			friendlyY = 4.5
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: friendlyX, Y: friendlyY}

		shapeDef := box2d.DefaultShapeDef()
		s.friendlyShape.maxPush = 0.025
		s.friendlyShape.clipVelocity = false

		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterMoverBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}
		shapeDef.UserData = &s.friendlyShape
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCapsuleShape(&shapeDef, &s.mover.capsule)
	}

	createCharacterBridge(s.worldID, groundID1, groundID2)

	for i := range 5 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{
			X: characterBallX,
			Y: posScalar(characterBallBaseY + 0.5*float32(i)),
		}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterDynamicBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}
		shapeDef.Material.Restitution = 0.7
		shapeDef.Material.RollingResistance = 0.2

		circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: characterBallRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Pos{
			X: posScalar(characterElevatorBase.X),
			Y: posScalar(characterElevatorBase.Y - characterElevatorAmplitude),
		}
		s.elevatorID = s.worldID.CreateBody(&bodyDef)

		const elevatorMaxPush = 0.1

		s.elevatorShape = MoverShapeUserData{
			maxPush:      elevatorMaxPush,
			clipVelocity: true,
		}
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterDynamicBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}
		shapeDef.UserData = &s.elevatorShape

		box := box2d.MakeBox(characterElevatorHalfWidth, characterElevatorHalfHeight)
		s.elevatorID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position.X = 140.0
		a := characterStackHalfWidth
		square := box2d.MakeSquare(a)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterDynamicBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}

		for i := range 10 {
			bodyDef.Position.Y = posScalar((2.0*float32(i)+1.0)*a + 1.0)
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &square)
		}
	}

	const geometricPendulumY = 5.0
	createCharacterPendulum(
		s.worldID,
		groundID2,
		box2d.Pos{X: characterPendulumX, Y: geometricPendulumY},
	)

	s.jumpReleased = true
	s.lockCamera = true
	s.time = 0.0

	return s
}

func (s *GeometryMoverSample) DrawControls() bool {
	imgui.SliderFloatV("Jump Speed", &s.mover.jumpSpeed, 0.0, characterJumpSpeedLimit, "%.0f", 0)
	imgui.SliderFloatV("Min Speed", &s.mover.minSpeed, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("Max Speed", &s.mover.maxSpeed, 0.0, characterMaxSpeedLimit, "%.0f", 0)
	imgui.SliderFloatV("Stop Speed", &s.mover.stopSpeed, 0.0, characterStopSpeedLimit, "%.1f", 0)
	imgui.SliderFloatV(
		"Accelerate",
		&s.mover.accelerate,
		0.0,
		characterAccelerationLimit,
		"%.0f",
		0,
	)
	imgui.SliderFloatV("Friction", &s.mover.friction, 0.0, characterFrictionLimit, "%.1f", 0)
	imgui.SliderFloatV("Gravity", &s.mover.gravity, 0.0, characterGravityLimit, "%.1f", 0)
	imgui.SliderFloatV("Air Steer", &s.mover.airSteer, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("Pogo Hertz", &s.mover.pogoHertz, 0.0, characterPogoHertzLimit, "%.0f", 0)
	imgui.SliderFloatV(
		"Pogo Damping",
		&s.mover.pogoDampingRatio,
		0.0,
		characterPogoDampingLimit,
		"%.1f",
		0,
	)

	imgui.Separator()

	imgui.Checkbox("Push Bodies", &s.mover.pushBodies)
	imgui.Checkbox("Lock Camera", &s.lockCamera)

	return true
}

func (s *GeometryMoverSample) Step() {
	s.DrawScreenTextLinef("left/right/jump = A/D/Space")

	pause := false
	if s.context.pause {
		pause = s.context.singleStep == 0
	}

	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}

	if pause {
		timeStep = 0.0
	}

	if timeStep > 0.0 {
		point := box2d.Pos{
			X: posScalar(characterElevatorBase.X),
			Y: posScalar(
				characterElevatorAmplitude*cosf(1.0*s.time+box2d.Pi) + characterElevatorBase.Y,
			),
		}

		wake := true
		s.elevatorID.SetTargetTransform(
			box2d.WorldTransform{P: point, Q: box2d.RotIdentity},
			timeStep,
			wake,
		)
	}

	s.time += timeStep

	s.Sample.Step()

	if !pause {
		throttle := characterThrottle(s.context.window)
		if s.context.window.GetKey(glfw.KeySpace) == glfw.Press {
			if s.jumpReleased && s.mover.Jump() {
				s.jumpReleased = false
			}
		} else {
			s.jumpReleased = true
		}

		s.mover.Update(timeStep, throttle)
	}

	drawPogo(
		s.draw,
		s.mover.pogoOrigin,
		s.mover.pogoTranslation,
		s.mover.pogoFraction,
		s.mover.pogoHit,
	)

	const transformScale = 0.25
	s.draw.DrawTransform(
		box2d.WorldTransform{P: s.mover.position, Q: box2d.RotIdentity},
		transformScale,
	)

	capsule := s.mover.capsule

	count := s.mover.planeCount
	for i := range count {
		const (
			planeNormalLength = 0.1
			planePointSize    = 5.0
		)

		plane := s.mover.planes[i].Plane
		p1 := s.mover.position.Add(plane.Normal.Scale(plane.Offset - capsule.Radius))
		p2 := p1.Add(plane.Normal.Scale(planeNormalLength))
		s.draw.DrawPoint(p1, planePointSize, box2d.ColorYellow)
		s.draw.DrawLine(p1, p2, box2d.ColorYellow)
	}

	p1 := s.mover.position.Add(capsule.Center1)
	p2 := s.mover.position.Add(capsule.Center2)

	color := box2d.ColorAquamarine
	if s.mover.onGround {
		color = box2d.ColorOrange
	}

	s.draw.DrawCapsule(p1, p2, capsule.Radius, color)
	s.draw.DrawLine(s.mover.position, s.mover.position.Add(s.mover.velocity), box2d.ColorPurple)

	p := s.mover.position
	s.DrawScreenTextLinef("position %.2f %.2f", p.X, p.Y)
	s.DrawScreenTextLinef("velocity %.2f %.2f", s.mover.velocity.X, s.mover.velocity.Y)
	s.DrawScreenTextLinef("iterations %d", s.mover.totalIterations)

	if s.lockCamera {
		s.camera.center.X = s.mover.position.X
	}
}

func CreateDynamicMoverSample(context *SampleContext) Sampler {
	s := &DynamicMoverSample{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		mover:           NewDynamicMover(),
	}
	s.init(context, s, true)

	if !context.restart {
		s.camera.center = box2d.Pos{X: characterCameraX, Y: characterCameraY}
		s.camera.zoom = 10.0
	}

	context.debugDraw.DrawJoints = false

	// This is the mover the player controls
	{
		def := DefaultDynamicMoverDef()
		def.position = box2d.Pos{X: 0.0, Y: characterSpawnY}
		def.capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -characterCapsuleHalfLength},
			Center2: box2d.Vec2{X: 0.0, Y: characterCapsuleHalfLength},
			Radius:  characterCapsuleRadius,
		}
		def.filter = box2d.DefaultFilter()
		def.filter.CategoryBits = characterMoverBit
		def.filter.MaskBits = characterStaticBit | characterDynamicBit
		def.enablePreSolveEvents = true
		s.mover.Create(s.worldID, &def)
	}

	s.elevatorID = box2d.Body{}

	var groundID1 box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "svg"
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID1 = s.worldID.CreateBody(&bodyDef)

		var points [64]box2d.Vec2

		offset := box2d.Vec2{X: -50.0, Y: -200.0}
		scale := characterPathScale

		count := ParsePath(characterDynamicGroundPath1, offset, points[:], scale)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true

		groundID1.CreateChain(&chainDef)
	}

	var groundID2 box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: characterSecondGroundX, Y: 0.0}
		groundID2 = s.worldID.CreateBody(&bodyDef)

		var points [64]box2d.Vec2

		offset := box2d.Vec2{X: 0.0, Y: -200.0}
		scale := characterPathScale

		count := ParsePath(characterGroundPath2, offset, points[:], scale)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true

		groundID2.CreateChain(&chainDef)
	}

	createCharacterBridge(s.worldID, groundID1, groundID2)

	for i := range 5 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{
			X: characterBallX,
			Y: posScalar(characterBallBaseY + 0.5*float32(i)),
		}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterDynamicBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}
		shapeDef.Material.Restitution = 0.7
		shapeDef.Material.RollingResistance = 0.2

		circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: characterBallRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Pos{
			X: posScalar(characterElevatorBase.X),
			Y: posScalar(characterElevatorBase.Y - characterElevatorAmplitude),
		}
		s.elevatorID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterDynamicBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}

		box := box2d.MakeBox(characterElevatorHalfWidth, characterElevatorHalfHeight)
		s.elevatorID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position.X = 140.0
		a := characterStackHalfWidth
		square := box2d.MakeSquare(a)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter = box2d.Filter{
			CategoryBits: characterDynamicBit,
			MaskBits:     characterAllBits,
			GroupIndex:   0,
		}

		for i := range 10 {
			bodyDef.Position.Y = posScalar((2.0*float32(i)+1.0)*a + 1.0)
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &square)
		}
	}

	const dynamicPendulumY = 4.0
	createCharacterPendulum(
		s.worldID,
		groundID2,
		box2d.Pos{X: characterPendulumX, Y: dynamicPendulumY},
	)

	s.preSolve = true

	// This is used for one-way collision on the kinematic elevator. But pre-solve doesn't work with recording.
	if s.preSolve {
		s.worldID.SetPreSolveCallback(s.PreSolve, s.PreContinuous)
	}

	s.jumpReleased = true
	s.lockCamera = true
	s.time = 0.0

	return s
}

func (s *DynamicMoverSample) PreSolve(
	shapeIDA box2d.Shape,
	shapeIDB box2d.Shape,
	manifold *box2d.Manifold,
) {
	idA := shapeIDA.Body()
	idB := shapeIDB.Body()

	if idA == s.elevatorID && idB == s.mover.moverID {
		if manifold.Normal.Y < 0.0 {
			manifold.PointCount = 0
		}
	} else if idA == s.mover.moverID && idB == s.elevatorID {
		if manifold.Normal.Y > 0.0 {
			manifold.PointCount = 0
		}
	}
}

func (s *DynamicMoverSample) PreContinuous(
	shapeIDA box2d.Shape, shapeIDB box2d.Shape,
	point box2d.Pos,
	normal box2d.Vec2,
) bool {
	idA := shapeIDA.Body()
	idB := shapeIDB.Body()

	if idA == s.elevatorID && idB == s.mover.moverID {
		return !(normal.Y < 0.0)
	}

	if idA == s.mover.moverID && idB == s.elevatorID {
		return !(normal.Y > 0.0)
	}

	return true
}

func (s *DynamicMoverSample) DrawControls() bool {
	imgui.SliderFloatV("Jump Speed", &s.mover.jumpSpeed, 0.0, characterJumpSpeedLimit, "%.0f", 0)
	imgui.SliderFloatV("Min Speed", &s.mover.minSpeed, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("Max Speed", &s.mover.maxSpeed, 0.0, characterMaxSpeedLimit, "%.0f", 0)
	imgui.SliderFloatV("Stop Speed", &s.mover.stopSpeed, 0.0, characterStopSpeedLimit, "%.1f", 0)
	imgui.SliderFloatV(
		"Accelerate",
		&s.mover.accelerate,
		0.0,
		characterAccelerationLimit,
		"%.0f",
		0,
	)
	imgui.SliderFloatV("Friction", &s.mover.friction, 0.0, characterFrictionLimit, "%.1f", 0)

	gravityScale := s.mover.gravityScale
	if imgui.SliderFloatV(
		"Gravity Scale",
		&gravityScale,
		0.0,
		characterGravityScaleLimit,
		"%.1f",
		0,
	) {
		s.mover.SetGravityScale(gravityScale)
	}

	imgui.SliderFloatV("Air Steer", &s.mover.airSteer, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("Pogo Hertz", &s.mover.pogoHertz, 0.0, characterPogoHertzLimit, "%.0f", 0)
	imgui.SliderFloatV(
		"Pogo Damping",
		&s.mover.pogoDampingRatio,
		0.0,
		characterPogoDampingLimit,
		"%.1f",
		0,
	)

	imgui.Separator()

	imgui.Checkbox("Lock Camera", &s.lockCamera)

	return true
}

func (s *DynamicMoverSample) Step() {
	s.DrawScreenTextLinef("left/right/jump = A/D/Space")

	pause := false
	if s.context.pause {
		pause = s.context.singleStep == 0
	}

	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}

	if pause {
		timeStep = 0.0
	}

	if timeStep > 0.0 && s.elevatorID != (box2d.Body{}) {
		point := box2d.Pos{
			X: posScalar(characterElevatorBase.X),
			Y: posScalar(
				characterElevatorAmplitude*cosf(1.0*s.time+box2d.Pi) + characterElevatorBase.Y,
			),
		}

		wake := true
		s.elevatorID.SetTargetTransform(
			box2d.WorldTransform{P: point, Q: box2d.RotIdentity},
			timeStep,
			wake,
		)
	}

	s.time += timeStep

	if !pause {
		s.updateInput(timeStep)
	}

	s.Sample.Step()

	drawPogo(
		s.draw,
		s.mover.pogoOrigin,
		s.mover.pogoTranslation,
		s.mover.pogoFraction,
		s.mover.pogoHit,
	)

	position := s.mover.moverID.Position()

	color := box2d.ColorAquamarine
	if s.mover.onGround {
		color = box2d.ColorOrange
	}

	s.draw.DrawLine(position, position.Add(s.mover.velocity), color)

	s.DrawScreenTextLinef("position %.2f %.2f", position.X, position.Y)
	s.DrawScreenTextLinef("velocity %.2f %.2f", s.mover.velocity.X, s.mover.velocity.Y)
	s.DrawScreenTextLinef(
		"pogo: impulse %.3f, velocity %.3f",
		s.mover.pogoImpulse,
		s.mover.pogoVelocity,
	)
	s.DrawScreenTextLinef("pogo: length = %.2f/%.2f", s.mover.pogoLength, s.mover.pogoRestLength)
	s.DrawScreenTextLinef("on ground: %d", boolToInt(s.mover.onGround))

	normalY := float32(0.0)
	if s.mover.castResult.hit {
		normalY = s.mover.castResult.normal.Y
	}

	s.DrawScreenTextLinef(
		"walkable: %d, normal y: %.3f/%.3f",
		boolToInt(s.mover.walkable),
		normalY,
		s.mover.minGroundNormalY,
	)
	s.DrawScreenTextLinef("jumping/ticks: %d/%d", boolToInt(s.mover.jumping), s.mover.jumpTicks)

	if s.lockCamera {
		s.camera.center.X = position.X
	}
}

func CreateGhostCollisionSample(context *SampleContext) Sampler {
	const (
		cameraHeight   = 1.5
		tileSpacing    = 2.0
		tileHalfHeight = 0.25
		wallHalfWidth  = 0.25
		wallHalfHeight = 1.25
		wallX          = 10.25
	)

	s := &GhostCollisionSample{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !context.restart {
		s.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.camera.zoom = 12.0
	}

	groundBodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&groundBodyDef)

	surfaceHeights := [2]float32{3.0, 0.0}
	for _, surfaceHeight := range surfaceHeights {
		for i := range 10 {
			x := -9.0 + tileSpacing*float32(i)
			tile := box2d.MakeOffsetBox(
				1.0,
				tileHalfHeight,
				box2d.Vec2{X: x, Y: surfaceHeight - tileHalfHeight},
				box2d.RotIdentity,
			)

			shapeDef := box2d.DefaultShapeDef()
			shapeDef.Material.Friction = 0.0

			shapeDef.Material.CustomColor = uint32(box2d.ColorDarkGray)
			if i%2 == 0 {
				shapeDef.Material.CustomColor = uint32(box2d.ColorGray)
			}

			groundID.CreatePolygonShape(&shapeDef, &tile)
		}

		wall := box2d.MakeOffsetBox(
			wallHalfWidth,
			wallHalfHeight,
			box2d.Vec2{X: wallX, Y: surfaceHeight + wallHalfHeight},
			box2d.RotIdentity,
		)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.0
		shapeDef.Material.CustomColor = uint32(box2d.ColorDarkGray)
		groundID.CreatePolygonShape(&shapeDef, &wall)
	}

	s.speed = 10.0
	s.playerIDs[0] = box2d.Body{}
	s.playerIDs[1] = box2d.Body{}
	s.rejectedCount.Store(0)

	s.worldID.SetPreSolveCallback(s.PreSolve, nil)
	s.Launch()

	return s
}

func (s *GhostCollisionSample) Launch() {
	const playerHalfExtent = 0.5

	s.rejectedCount.Store(0)

	surfaceHeights := [2]float32{3.0, 0.0}
	colors := [2]box2d.HexColor{box2d.ColorBox2DRed, box2d.ColorBox2DGreen}

	for i := range 2 {
		if s.playerIDs[i] != (box2d.Body{}) {
			s.playerIDs[i].Destroy()
		}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -9.0, Y: posScalar(surfaceHeights[i] + playerHalfExtent)}
		bodyDef.LinearVelocity = box2d.Vec2{X: s.speed, Y: 0.0}
		bodyDef.MotionLocks.AngularZ = true

		// Pre-solve must see a fresh manifold at every tile boundary.
		bodyDef.EnableContactRecycling = false
		s.playerIDs[i] = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		shapeDef.Material.Friction = 0.0
		shapeDef.Material.CustomColor = uint32(colors[i])
		shapeDef.EnablePreSolveEvents = i == 1

		player := box2d.MakeBox(playerHalfExtent, playerHalfExtent)
		s.playerIDs[i].CreatePolygonShape(&shapeDef, &player)
	}
}

func (s *GhostCollisionSample) PreSolve(
	shapeIDA box2d.Shape,
	shapeIDB box2d.Shape,
	manifold *box2d.Manifold,
) {
	const faceContactPoints = 2
	if manifold.PointCount != faceContactPoints {
		return
	}

	normal := manifold.Normal

	bodyIDA := shapeIDA.Body()
	if bodyIDA != s.playerIDs[1] {
		bodyIDB := shapeIDB.Body()
		if bodyIDB != s.playerIDs[1] {
			return
		}

		normal = normal.Neg()
	}

	velocity := s.playerIDs[1].LinearVelocity()
	if normal.Dot(velocity) < 0.0 {
		return
	}

	// A short face contact is overlap slop at a tile seam, not a wall.
	contactArea := manifold.Points[0].AnchorA.Distance(manifold.Points[1].AnchorA)
	if contactArea < 4.0*box2d.LinearSlop() {
		manifold.PointCount = 0

		// PreSolve is called from a thread and should not have data race.
		s.rejectedCount.Add(1)
	}
}

func (s *GhostCollisionSample) DrawControls() bool {
	const (
		controlWidthInFonts = 6.0
		speedLimit          = 20.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())
	imgui.SliderFloatV("Speed", &s.speed, 1.0, speedLimit, "%.1f", 0)
	imgui.PopItemWidth()

	if imgui.Button("Launch") {
		s.Launch()
	}

	return true
}

func (s *GhostCollisionSample) Step() {
	const (
		unfilteredLabelY = 4.2
		filteredLabelY   = 1.2
	)

	s.Sample.Step()

	s.draw.DrawStringf(
		s.camera,
		box2d.Pos{X: -9.8, Y: unfilteredLabelY},
		box2d.ColorWhite,
		"unfiltered",
	)
	s.draw.DrawStringf(
		s.camera,
		box2d.Pos{X: -9.8, Y: filteredLabelY},
		box2d.ColorWhite,
		"shallow contacts rejected",
	)
	s.DrawScreenTextLinef("Rejected contacts: %d", s.rejectedCount.Load())
}

func registerCharacterSamples() {
	RegisterSample("Character", "Geometric Mover", CreateGeometryMoverSample)
	RegisterSample("Character", "Dynamic Mover", CreateDynamicMoverSample)
	RegisterSample("Character", "Ghost Collision", CreateGhostCollisionSample)
}

// The ground probe is a point, a circle, or a segment, all carried by the proxy.
func drawPogo(draw *Draw, origin box2d.Pos, translation box2d.Vec2, fraction float32, hit bool) {
	const pogoPointSize = 10.0

	color := box2d.ColorGray
	if hit {
		color = box2d.ColorPlum
	}

	delta := translation.Scale(fraction)

	draw.DrawLine(origin, origin.Add(delta), box2d.ColorGray)
	draw.DrawPoint(origin.Add(delta), pogoPointSize, color)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}

	return 0
}

func createCharacterBridge(world *box2d.World, groundID1 box2d.Body, groundID2 box2d.Body) {
	const (
		plankHalfWidth      = 0.55
		plankHalfHeight     = 0.125
		motorTorque         = 10.0
		springHertz         = 3.0
		springDamping       = 0.8
		bridgeStartX        = float32(48.7)
		bridgeHeight        = float32(9.2)
		plankCount          = 50
		plankAngularDamping = 0.2
	)

	box := box2d.MakeBox(plankHalfWidth, plankHalfHeight)
	shapeDef := box2d.DefaultShapeDef()
	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.MaxMotorTorque = motorTorque
	jointDef.EnableMotor = true
	jointDef.Hertz = springHertz
	jointDef.DampingRatio = springDamping
	jointDef.EnableSpring = true
	prevBodyID := groundID1

	for i := range plankCount {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{
			X: posScalar(bridgeStartX + 0.5 + 1.0*float32(i)),
			Y: posScalar(bridgeHeight),
		}
		bodyDef.AngularDamping = plankAngularDamping
		bodyID := world.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: posScalar(bridgeStartX + 1.0*float32(i)), Y: posScalar(bridgeHeight)}
		jointDef.Base.BodyA = prevBodyID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		world.CreateRevoluteJoint(&jointDef)

		prevBodyID = bodyID
	}

	pivot := box2d.Pos{
		X: posScalar(bridgeStartX + 1.0*float32(plankCount)),
		Y: posScalar(bridgeHeight),
	}
	jointDef.Base.BodyA = prevBodyID
	jointDef.Base.BodyB = groundID2
	jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
	jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
	world.CreateRevoluteJoint(&jointDef)
}

func createCharacterPendulum(world *box2d.World, groundID box2d.Body, position box2d.Pos) {
	const (
		pendulumHalfWidth  = 0.1
		pendulumHalfHeight = 4.0
		motorTorque        = 500.0
	)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = position
	bodyDef.Type = box2d.DynamicBody
	body := world.CreateBody(&bodyDef)
	box := box2d.MakeBox(pendulumHalfWidth, pendulumHalfHeight)
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	body.CreatePolygonShape(&shapeDef, &box)

	pivot := bodyDef.Position.Add(box2d.Vec2{X: 0.0, Y: pendulumHalfHeight})
	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.BodyA = groundID
	jointDef.Base.BodyB = body
	jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
	jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
	jointDef.EnableMotor = true
	jointDef.MotorSpeed = 1.0
	jointDef.MaxMotorTorque = motorTorque
	world.CreateRevoluteJoint(&jointDef)
}

func characterThrottle(window *glfw.Window) float32 {
	throttle := float32(0.0)
	if window.GetKey(glfw.KeyA) == glfw.Press {
		throttle -= 1.0
	}

	if window.GetKey(glfw.KeyD) == glfw.Press {
		throttle += 1.0
	}

	return throttle
}

func (s *DynamicMoverSample) updateInput(timeStep float32) {
	throttle := characterThrottle(s.context.window)
	if s.context.window.GetKey(glfw.KeySpace) == glfw.Press {
		if s.jumpReleased && s.mover.Jump() {
			s.jumpReleased = false
		}
	} else {
		s.jumpReleased = true
	}

	s.mover.Update(timeStep, throttle)
}
