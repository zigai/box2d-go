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
	continuousCircleShape = iota
	continuousCapsuleShape
	continuousBoxShape
)

type bounceHitEvent struct {
	point     box2d.Pos
	speed     float32
	stepIndex int
}

type BounceHouse struct {
	Sample

	hitEvents       [4]bounceHitEvent
	bodyID          box2d.Body
	shapeType       int32
	enableHitEvents bool
}

type BounceHumans struct {
	Sample

	humans     [5]scenes.Human
	humanCount int
	countDown  float32
	time       float32
}

type ChainDrop struct {
	Sample

	bodyID  box2d.Body
	shapeID box2d.Shape
	yOffset float32
	speed   float32
}

type ChainSlide struct {
	Sample
}

type SegmentSlide struct {
	Sample
}

type SkinnyBox struct {
	Sample

	bodyID          box2d.Body
	bulletID        box2d.Body
	angularVelocity float32
	x               float32
	capsule         bool
	autoTest        bool
	bullet          bool
}

// This sample shows ghost bumps.
type GhostBumps struct {
	Sample

	groundID  box2d.Body
	bodyID    box2d.Body
	shapeID   box2d.Shape
	shapeType int32
	round     float32
	friction  float32
	bevel     float32
	useChain  bool
}

// Speculative collision failure case suggested by Dirk Gregorius. This uses
// a simple fallback scheme to prevent tunneling.
type SpeculativeFallback struct {
	Sample
}

type SpeculativeSliver struct {
	Sample
}

// Speculative collision does not cause speculative ghost collisions at small distances.
type SpeculativeGhost struct {
	Sample
}

// This shows that Box2D does not have pixel perfect collision.
type PixelImperfect struct {
	Sample

	ballID box2d.Body
}

type RestitutionThreshold struct {
	Sample

	ballID box2d.Body
}

type Drop struct {
	Sample

	groundIDs   []box2d.Body
	bodyIDs     []box2d.Body
	human       scenes.Human
	frameSkip   int
	frameCount  int
	continuous  bool
	speculative bool
}

// This shows a fast moving body that uses continuous collision versus static and dynamic bodies.
// This is achieved by setting the ball body as a *bullet*.
type Pinball struct {
	Sample

	leftJointID  box2d.RevoluteJoint
	rightJointID box2d.RevoluteJoint
	ballID       box2d.Body
}

// This shows the importance of secondary collisions in continuous physics.
// This also shows a difficult setup for the solver with an acute angle.
type Wedge struct {
	Sample
}

// This shows how adjusting the CCD safety factor can engage continuous collision at
// lower speeds to avoid the slight overlap that happens with discrete collision detection.
// The best way to run this sample is:
// 1. press pause (P)
// 2. restart (R) or press the Drop button
// 3. then single step (.)
// Then look at the metrics and overlap. The color of the shape gets darker when it is using
// continuous collision detection.
type SafetyFactor struct {
	Sample

	bodyID       box2d.Body
	extent       float32
	height       float32
	safetyFactor float32
	overlap      float32
}

func CreateBounceHouse(context *SampleContext) Sampler {
	const (
		baseCameraZoom      = 25.0
		cameraZoomScale     = 0.45
		enclosureHalfExtent = 10.0
	)

	s := &BounceHouse{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		hitEvents:       [4]bounceHitEvent{},
		bodyID:          box2d.Body{},
		shapeType:       continuousCircleShape,
		enableHitEvents: true,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -enclosureHalfExtent, Y: -enclosureHalfExtent},
			Point2: box2d.Vec2{X: enclosureHalfExtent, Y: -enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: enclosureHalfExtent, Y: -enclosureHalfExtent},
			Point2: box2d.Vec2{X: enclosureHalfExtent, Y: enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: enclosureHalfExtent, Y: enclosureHalfExtent},
			Point2: box2d.Vec2{X: -enclosureHalfExtent, Y: enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -enclosureHalfExtent, Y: enclosureHalfExtent},
			Point2: box2d.Vec2{X: -enclosureHalfExtent, Y: -enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.Launch()

	return s
}

func (s *BounceHouse) Launch() {
	const (
		launchVelocityX   = 10.0
		launchVelocityY   = 20.0
		circleRadius      = 0.5
		capsuleHalfLength = 0.5
		capsuleRadius     = 0.25
		boxHalfHeight     = 0.1
		boxAspectRatio    = 20.0
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.LinearVelocity = box2d.Vec2{X: launchVelocityX, Y: launchVelocityY}
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
	bodyDef.GravityScale = 0.0
	bodyDef.IsBullet = true

	// Circle shapes centered on the body can spin fast without risk of tunneling.
	bodyDef.AllowFastRotation = s.shapeType == continuousCircleShape

	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.EnableHitEvents = s.enableHitEvents

	switch s.shapeType {
	case continuousCircleShape:
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
		s.bodyID.CreateCircleShape(&shapeDef, &circle)
	case continuousCapsuleShape:
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
			Radius:  capsuleRadius,
		}
		s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	default:
		h := float32(boxHalfHeight)
		box := box2d.MakeBox(boxAspectRatio*h, h)
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *BounceHouse) DrawControls() bool {
	const controlWidthEm = 6.0

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	shapeTypes := []string{"Circle", "Capsule", "Box"}
	if imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes))) {
		s.Launch()
	}

	imgui.PopItemWidth()

	if imgui.Checkbox("hit events", &s.enableHitEvents) {
		s.bodyID.SetHitEventsEnabled(s.enableHitEvents)
	}

	return true
}

func (s *BounceHouse) Step() {
	const hitMarkerRadius = 0.1

	s.Sample.Step()

	for event := range s.worldID.ContactHitEvents() {
		e := &s.hitEvents[0]
		for j := 1; j < 4; j++ {
			if s.hitEvents[j].stepIndex < e.stepIndex {
				e = &s.hitEvents[j]
			}
		}

		e.point = event.Point
		e.speed = event.ApproachSpeed
		e.stepIndex = s.stepCount
	}

	for i := range 4 {
		e := &s.hitEvents[i]
		if e.stepIndex > 0 && s.stepCount <= e.stepIndex+30 {
			s.draw.DrawCircle(e.point, hitMarkerRadius, box2d.ColorOrangeRed)
			s.draw.DrawStringf(s.camera, e.point, box2d.ColorWhite, "%.1f", e.speed)
		}
	}
}

func CreateBounceHumans(context *SampleContext) Sampler {
	const (
		enclosureHalfExtent = 10.0
		centerBumperRadius  = 2.0
	)

	s := &BounceHumans{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		humans:     [5]scenes.Human{},
		humanCount: 0,
		countDown:  0,
		time:       0,
	}
	s.init(context, s, true)

	s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
	s.context.camera.zoom = 12.0

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 1.3
	shapeDef.Material.Friction = 0.1

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -enclosureHalfExtent, Y: -enclosureHalfExtent},
			Point2: box2d.Vec2{X: enclosureHalfExtent, Y: -enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: enclosureHalfExtent, Y: -enclosureHalfExtent},
			Point2: box2d.Vec2{X: enclosureHalfExtent, Y: enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: enclosureHalfExtent, Y: enclosureHalfExtent},
			Point2: box2d.Vec2{X: -enclosureHalfExtent, Y: enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -enclosureHalfExtent, Y: enclosureHalfExtent},
			Point2: box2d.Vec2{X: -enclosureHalfExtent, Y: -enclosureHalfExtent},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: centerBumperRadius}
	shapeDef.Material.Restitution = 2.0
	groundID.CreateCircleShape(&shapeDef, &circle)

	return s
}

func (s *BounceHumans) Step() {
	const (
		jointDampingRatioValue = 0.1
		spawnHeight            = 5.0
		stepsPerSecond         = 60.0
		gravityPhaseRate       = 0.5
		gravityMagnitude       = 10.0
		gravityIndicatorLength = 3.0
	)

	if s.humanCount < 5 && s.countDown <= 0.0 {
		jointFrictionTorque := float32(0.0)
		jointHertz := float32(1.0)
		jointDampingRatio := float32(jointDampingRatioValue)

		scenes.CreateHuman(
			&s.humans[s.humanCount],
			s.worldID,
			box2d.Pos{X: 0.0, Y: spawnHeight},
			1.0,
			jointFrictionTorque,
			jointHertz,
			jointDampingRatio,
			1,
			nil,
			true,
		)

		s.countDown = 2.0
		s.humanCount += 1
	}

	timeStep := float32(1.0 / stepsPerSecond)
	cs1 := box2d.ComputeCosSin(gravityPhaseRate * s.time)
	cs2 := box2d.ComputeCosSin(s.time)
	gravity := float32(gravityMagnitude)
	gravityVec := box2d.Vec2{X: gravity * cs1.Sine, Y: gravity * cs2.Cosine}
	s.draw.DrawLine(
		box2d.Pos{X: 0, Y: 0},
		box2d.Vec2{
			X: gravityIndicatorLength * cs1.Sine,
			Y: gravityIndicatorLength * cs2.Cosine,
		}.Pos(),
		box2d.ColorWhite,
	)
	s.time += timeStep
	s.countDown -= timeStep
	s.worldID.SetGravity(gravityVec)

	s.Sample.Step()
}

func CreateChainDrop(context *SampleContext) Sampler {
	const (
		baseCameraZoom   = 25.0
		cameraZoomScale  = 0.35
		chainHalfWidth   = 10.0
		chainBottom      = -2.0
		chainTop         = 1.0
		initialYOffset   = -0.1
		initialDropSpeed = -42.0
	)

	s := &ChainDrop{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		bodyID:  box2d.Body{},
		shapeID: box2d.Shape{},
		yOffset: initialYOffset,
		speed:   initialDropSpeed,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -6.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	points := []box2d.Vec2{
		{X: -chainHalfWidth, Y: chainBottom},
		{X: chainHalfWidth, Y: chainBottom},
		{X: chainHalfWidth, Y: chainTop},
		{X: -chainHalfWidth, Y: chainTop},
	}

	chainDef := box2d.DefaultChainDef()
	chainDef.Points = points
	chainDef.IsLoop = true

	groundID.CreateChain(&chainDef)

	s.Launch()

	return s
}

func (s *ChainDrop) Launch() {
	const (
		dropHeight            = 10.0
		quarterTurnPiFraction = 0.5
		circleRadius          = 0.5
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: s.speed}
	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(dropHeight + s.yOffset)}
	bodyDef.Rotation = box2d.MakeRot(quarterTurnPiFraction * box2d.Pi)
	bodyDef.MotionLocks.AngularZ = true
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
	s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
}

func (s *ChainDrop) DrawControls() bool {
	const controlWidthEm = 6.0

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())
	imgui.SliderFloatV("Speed", &s.speed, -100.0, 0.0, "%.0f", 0)
	imgui.SliderFloatV("Y Offset", &s.yOffset, -1.0, 1.0, "%.1f", 0)
	imgui.PopItemWidth()

	if imgui.Button("Launch") {
		s.Launch()
	}

	return true
}

func CreateChainSlide(context *SampleContext) Sampler {
	const (
		cameraHeight        = 10.0
		chainHorizontalStep = 2.0
		chainVerticalStep   = 1.0
		chainStartX         = 20.0
		slideSpeed          = 100.0
		ballRadius          = 0.5
	)

	s := &ChainSlide{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 15.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		const count = 80

		var points [count]box2d.Vec2

		w := float32(chainHorizontalStep)
		h := float32(chainVerticalStep)

		x, y := float32(chainStartX), float32(0.0)
		for i := range 20 {
			points[i] = box2d.Vec2{X: x, Y: y}
			x -= w
		}

		for i := 20; i < 40; i++ {
			points[i] = box2d.Vec2{X: x, Y: y}
			y += h
		}

		for i := 40; i < 60; i++ {
			points[i] = box2d.Vec2{X: x, Y: y}
			x += w
		}

		for i := 60; i < 80; i++ {
			points[i] = box2d.Vec2{X: x, Y: y}
			y -= h
		}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.IsLoop = true

		groundID.CreateChain(&chainDef)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.LinearVelocity = box2d.Vec2{X: slideSpeed, Y: 0.0}
		bodyDef.Position = box2d.Pos{X: -19.5, Y: 0.0 + ballRadius}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.0
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	return s
}

func CreateSegmentSlide(context *SampleContext) Sampler {
	const (
		cameraHeight     = 10.0
		groundHalfLength = 40.0
		wallHeight       = 10.0
		slideSpeed       = 100.0
		startHeight      = 0.7
		ballRadius       = 0.5
	)

	s := &SegmentSlide{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 15.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{
			Point1: box2d.Vec2{X: groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: wallHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.LinearVelocity = box2d.Vec2{X: slideSpeed, Y: 0.0}
		bodyDef.Position = box2d.Pos{X: -20.0, Y: startHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	return s
}

func CreateSkinnyBox(context *SampleContext) Sampler {
	const (
		cameraHeight       = 5.0
		baseCameraZoom     = 25.0
		cameraZoomScale    = 0.25
		groundHalfLength   = 10.0
		obstacleHalfWidth  = 0.1
		obstacleHalfHeight = 1.0
		obstacleCenterY    = 1.0
	)

	s := &SkinnyBox{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		bodyID:          box2d.Body{},
		bulletID:        box2d.Body{},
		angularVelocity: 0,
		x:               0,
		capsule:         false,
		autoTest:        false,
		bullet:          false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.0, Y: cameraHeight}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.9
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(
			obstacleHalfWidth,
			obstacleHalfHeight,
			box2d.Vec2{X: 0.0, Y: obstacleCenterY},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.Launch()

	return s
}

func (s *SkinnyBox) Launch() {
	const (
		maxAngularSpeed  = 50.0
		launchHeight     = 8.0
		capsuleRadius    = 0.1
		boxHalfWidth     = 2.0
		boxHalfHeight    = 0.05
		bulletHalfExtent = 0.25
		bulletHeight     = 10.0
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	if s.bulletID != (box2d.Body{}) {
		s.bulletID.Destroy()
	}

	s.angularVelocity = scenes.RandomFloatRange(-maxAngularSpeed, maxAngularSpeed)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: launchHeight}
	bodyDef.AngularVelocity = s.angularVelocity
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.9

	s.bodyID = s.worldID.CreateBody(&bodyDef)

	if s.capsule {
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -1.0},
			Center2: box2d.Vec2{X: 0.0, Y: 1.0},
			Radius:  capsuleRadius,
		}
		s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		polygon := box2d.MakeBox(boxHalfWidth, boxHalfHeight)
		s.bodyID.CreatePolygonShape(&shapeDef, &polygon)
	}

	if s.bullet {
		polygon := box2d.MakeBox(bulletHalfExtent, bulletHalfExtent)
		s.x = scenes.RandomFloatRange(-1.0, 1.0)
		bodyDef.Position = box2d.Pos{X: posScalar(s.x), Y: bulletHeight}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -50.0}
		s.bulletID = s.worldID.CreateBody(&bodyDef)
		s.bulletID.CreatePolygonShape(&shapeDef, &polygon)
	}
}

func (s *SkinnyBox) DrawControls() bool {
	imgui.Checkbox("Capsule", &s.capsule)

	if imgui.Button("Launch") {
		s.Launch()
	}

	imgui.Checkbox("Auto Test", &s.autoTest)

	return true
}

func (s *SkinnyBox) Step() {
	s.Sample.Step()

	if s.autoTest && s.stepCount%60 == 0 {
		s.Launch()
	}
}

func CreateGhostBumps(context *SampleContext) Sampler {
	const (
		initialGroundFriction = 0.2
		cameraCenterX         = 1.5
		cameraCenterY         = 16.0
		baseCameraZoom        = 25.0
		cameraZoomScale       = 0.8
	)

	s := &GhostBumps{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		groundID:  box2d.Body{},
		bodyID:    box2d.Body{},
		shapeID:   box2d.Shape{},
		shapeType: continuousCircleShape,
		round:     0,
		friction:  initialGroundFriction,
		bevel:     0,
		useChain:  true,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	s.CreateScene()
	s.Launch()

	return s
}

func (s *GhostBumps) CreateScene() {
	const (
		diagonalDimensions = 2.0
		segmentHalfLength  = 4.0
		segmentHalfHeight  = 0.25
		bevelInset         = 0.05
		segmentSpanFactor  = 2.0
		slopeAnglePiRatio  = 0.25
	)

	if s.groundID != (box2d.Body{}) {
		s.groundID.Destroy()
	}

	s.shapeID = box2d.Shape{}

	bodyDef := box2d.DefaultBodyDef()
	s.groundID = s.worldID.CreateBody(&bodyDef)

	m := float32(1.0 / math.Sqrt(diagonalDimensions))
	mm := float32(segmentSpanFactor * (math.Sqrt(diagonalDimensions) - 1.0))
	hx, hy := float32(segmentHalfLength), float32(segmentHalfHeight)

	if s.useChain {
		var points [20]box2d.Vec2

		points[0] = box2d.Vec2{X: -3.0 * hx, Y: hy}
		points[1] = points[0].Add(box2d.Vec2{X: -2.0 * hx * m, Y: 2.0 * hx * m})
		points[2] = points[1].Add(box2d.Vec2{X: -2.0 * hx * m, Y: 2.0 * hx * m})
		points[3] = points[2].Add(box2d.Vec2{X: -2.0 * hx * m, Y: 2.0 * hx * m})
		points[4] = points[3].Add(box2d.Vec2{X: -2.0 * hy * m, Y: -2.0 * hy * m})
		points[5] = points[4].Add(box2d.Vec2{X: 2.0 * hx * m, Y: -2.0 * hx * m})
		points[6] = points[5].Add(box2d.Vec2{X: 2.0 * hx * m, Y: -2.0 * hx * m})
		points[7] = points[6].Add(
			box2d.Vec2{X: 2.0*hx*m + 2.0*hy*(1.0-m), Y: -2.0*hx*m - 2.0*hy*(1.0-m)},
		)
		points[8] = points[7].Add(box2d.Vec2{X: 2.0*hx + hy*mm, Y: 0.0})
		points[9] = points[8].Add(box2d.Vec2{X: segmentSpanFactor * hx, Y: 0.0})
		points[10] = points[9].Add(box2d.Vec2{X: 2.0*hx + hy*mm, Y: 0.0})
		points[11] = points[10].Add(
			box2d.Vec2{X: 2.0*hx*m + 2.0*hy*(1.0-m), Y: 2.0*hx*m + 2.0*hy*(1.0-m)},
		)
		points[12] = points[11].Add(box2d.Vec2{X: 2.0 * hx * m, Y: 2.0 * hx * m})
		points[13] = points[12].Add(box2d.Vec2{X: 2.0 * hx * m, Y: 2.0 * hx * m})
		points[14] = points[13].Add(box2d.Vec2{X: -2.0 * hy * m, Y: 2.0 * hy * m})
		points[15] = points[14].Add(box2d.Vec2{X: -2.0 * hx * m, Y: -2.0 * hx * m})
		points[16] = points[15].Add(box2d.Vec2{X: -2.0 * hx * m, Y: -2.0 * hx * m})
		points[17] = points[16].Add(box2d.Vec2{X: -2.0 * hx * m, Y: -2.0 * hx * m})
		points[18] = points[17].Add(box2d.Vec2{X: -2.0 * hx, Y: 0.0})
		points[19] = points[18].Add(box2d.Vec2{X: -2.0 * hx, Y: 0.0})

		var material box2d.SurfaceMaterial

		material.Friction = s.friction

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}

		s.groundID.CreateChain(&chainDef)
	} else {
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = s.friction

		var hull box2d.Hull

		if s.bevel > 0.0 {
			hb := s.bevel
			vs := []box2d.Vec2{
				{X: hx + hb, Y: hy - bevelInset},
				{X: hx, Y: hy},
				{X: -hx, Y: hy},
				{X: -hx - hb, Y: hy - bevelInset},
				{X: -hx - hb, Y: -hy + bevelInset},
				{X: -hx, Y: -hy},
				{X: hx, Y: -hy},
				{X: hx + hb, Y: -hy + bevelInset},
			}
			hull = box2d.ComputeHull(vs)
		} else {
			vs := []box2d.Vec2{{X: hx, Y: hy}, {X: -hx, Y: hy}, {X: -hx, Y: -hy}, {X: hx, Y: -hy}}
			hull = box2d.ComputeHull(vs)
		}

		var (
			transform box2d.Transform
			x, y      float32
		)

		createSegment := func() {
			transform.P = box2d.Vec2{X: x, Y: y}
			polygon := box2d.MakeOffsetPolygon(&hull, transform.P, transform.Q)
			s.groundID.CreatePolygonShape(&shapeDef, &polygon)
		}

		// Left slope
		x = -3.0*hx - m*hx - m*hy
		y = hy + m*hx - m*hy
		transform.Q = box2d.MakeRot(-slopeAnglePiRatio * box2d.Pi)

		for range 3 {
			createSegment()

			x -= segmentSpanFactor * m * hx
			y += segmentSpanFactor * m * hx
		}

		x = -2.0 * hx
		y = 0.0
		transform.Q = box2d.MakeRot(0.0)

		for range 3 {
			createSegment()

			x += segmentSpanFactor * hx
		}

		x = 3.0*hx + m*hx + m*hy
		y = hy + m*hx - m*hy
		transform.Q = box2d.MakeRot(slopeAnglePiRatio * box2d.Pi)

		for range 3 {
			createSegment()

			x += segmentSpanFactor * m * hx
			y += segmentSpanFactor * m * hx
		}
	}
}

func (s *GhostBumps) Launch() {
	const (
		launchHeight      = 18.0
		circleRadius      = 0.5
		capsuleHalfLength = 0.5
		capsuleRadius     = 0.25
		boxHalfWidth      = 0.5
		boxAspectRatio    = 2.0
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
		s.shapeID = box2d.Shape{}
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -28.0, Y: launchHeight}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: 0.0}
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = s.friction

	switch s.shapeType {
	case continuousCircleShape:
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
		s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
	case continuousCapsuleShape:
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
			Radius:  capsuleRadius,
		}
		s.shapeID = s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	default:
		h := boxHalfWidth - s.round
		box := box2d.MakeRoundedBox(h, boxAspectRatio*h, s.round)
		s.shapeID = s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *GhostBumps) DrawControls() bool {
	const (
		controlWidthEm = 6.0
		maxRoundRadius = 0.4
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.Checkbox("Chain", &s.useChain) {
		s.CreateScene()
	}

	if !s.useChain {
		if imgui.SliderFloatV("Bevel", &s.bevel, 0.0, 1.0, "%.2f", 0) {
			s.CreateScene()
		}
	}

	{
		shapeTypes := []string{"Circle", "Capsule", "Box"}
		imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))
	}

	if s.shapeType == continuousBoxShape {
		imgui.SliderFloatV("Round", &s.round, 0.0, maxRoundRadius, "%.1f", 0)
	}

	if imgui.SliderFloatV("Friction", &s.friction, 0.0, 1.0, "%.1f", 0) {
		if s.shapeID != (box2d.Shape{}) {
			s.shapeID.SetFriction(s.friction)
		}

		s.CreateScene()
	}

	if imgui.Button("Launch") {
		s.Launch()
	}

	imgui.PopItemWidth()

	return true
}

func CreateSpeculativeFallback(context *SampleContext) Sampler {
	const (
		cameraHeight      = 5.0
		baseCameraZoom    = 25.0
		cameraZoomScale   = 0.25
		groundHalfLength  = 10.0
		platformHalfWidth = 2.0
		platformBottom    = 4.0
		platformStep      = 4.1
		platformTop       = 4.2
		platformCornerX   = -0.5
		shapeOffset       = 8.0
		dropHeight        = 12.0
		boxHalfWidth      = 2.0
		boxHalfHeight     = 0.05
	)

	s := &SpeculativeFallback{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.0, Y: cameraHeight}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		points := []box2d.Vec2{
			{X: -platformHalfWidth, Y: platformBottom},
			{X: platformHalfWidth, Y: platformBottom},
			{X: platformHalfWidth, Y: platformStep},
			{X: platformCornerX, Y: platformTop},
			{X: -platformHalfWidth, Y: platformTop},
		}
		hull := box2d.ComputeHull(points)
		poly := box2d.MakePolygon(&hull, 0.0)
		groundID.CreatePolygonShape(&shapeDef, &poly)
	}

	// Fast moving skinny box. Also testing a large shape offset.
	{
		offset := float32(shapeOffset)
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(offset), Y: dropHeight}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			boxHalfWidth,
			boxHalfHeight,
			box2d.Vec2{X: -offset, Y: 0.0},
			box2d.MakeRot(box2d.Pi),
		)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func CreateSpeculativeSliver(context *SampleContext) Sampler {
	const (
		cameraHeight     = 1.75
		groundHalfLength = 10.0
		dropHeight       = 12.0
		sliverHalfWidth  = 2.0
		sliverTipHeight  = 0.5
	)

	s := &SpeculativeSliver{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 2.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: dropHeight}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		points := []box2d.Vec2{
			{
				X: -sliverHalfWidth,
				Y: 0.0,
			},
			{X: -1.0, Y: 0.0},
			{X: sliverHalfWidth, Y: sliverTipHeight},
		}
		hull := box2d.ComputeHull(points)
		poly := box2d.MakePolygon(&hull, 0.0)
		bodyID.CreatePolygonShape(&shapeDef, &poly)
	}

	return s
}

func CreateSpeculativeGhost(context *SampleContext) Sampler {
	const (
		cameraHeight     = 1.75
		groundHalfLength = 10.0
		ledgeHalfHeight  = 0.1
		ledgeCenterY     = 0.9
		bodyStartX       = 0.015
		bodyStartY       = 2.515
		boxHalfExtent    = 0.25
	)

	s := &SpeculativeGhost{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 2.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(
			1.0, ledgeHalfHeight, box2d.Vec2{X: 0.0, Y: ledgeCenterY}, box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		// The speculative distance is 0.02 meters, so this avoid it
		bodyDef.Position = box2d.Pos{X: bodyStartX, Y: bodyStartY}
		bodyDef.LinearVelocity = box2d.Vec2{
			X: 0.1 * 1.25 * s.context.hertz,
			Y: -0.1 * 1.25 * s.context.hertz,
		}
		bodyDef.GravityScale = 0.0
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeSquare(boxHalfExtent)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func CreatePixelImperfect(context *SampleContext) Sampler {
	const (
		cameraX              = 7.0
		cameraY              = 5.0
		pixelScale           = 30.0
		blockXPixel          = 175.0
		blockYPixel          = 150.0
		blockHalfWidthPixel  = 20.0
		blockHalfHeightPixel = 10.0
		ballXPixel           = 200.0
		ballYPixel           = 275.0
		ballHalfExtentPixel  = 4.0
		ballRoundingPixel    = 0.9
	)

	s := &PixelImperfect{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		ballID:          box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = 6.0
	}

	pixelsPerMeter := float32(pixelScale)

	{
		block4BodyDef := box2d.DefaultBodyDef()
		block4BodyDef.Type = box2d.StaticBody
		block4BodyDef.Position = box2d.Pos{
			X: posScalar(blockXPixel / pixelsPerMeter),
			Y: posScalar(blockYPixel / pixelsPerMeter),
		}
		block4BodyID := s.worldID.CreateBody(&block4BodyDef)
		block4Shape := box2d.MakeBox(
			blockHalfWidthPixel/pixelsPerMeter,
			blockHalfHeightPixel/pixelsPerMeter,
		)
		block4ShapeDef := box2d.DefaultShapeDef()
		block4ShapeDef.Material.Friction = 0.0
		block4BodyID.CreatePolygonShape(&block4ShapeDef, &block4Shape)
	}

	{
		ballBodyDef := box2d.DefaultBodyDef()
		ballBodyDef.Type = box2d.DynamicBody
		ballBodyDef.Position = box2d.Pos{
			X: posScalar(ballXPixel / pixelsPerMeter),
			Y: posScalar(ballYPixel / pixelsPerMeter),
		}
		ballBodyDef.GravityScale = 0.0

		s.ballID = s.worldID.CreateBody(&ballBodyDef)
		ballShape := box2d.MakeRoundedBox(
			ballHalfExtentPixel/pixelsPerMeter,
			ballHalfExtentPixel/pixelsPerMeter,
			ballRoundingPixel/pixelsPerMeter,
		)
		ballShapeDef := box2d.DefaultShapeDef()
		ballShapeDef.Material.Friction = 0.0
		s.ballID.CreatePolygonShape(&ballShapeDef, &ballShape)
		s.ballID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: -5.0})
		s.ballID.SetMotionLocks(box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: true})
	}

	return s
}

func (s *PixelImperfect) Step() {
	p := s.ballID.Position()
	v := s.ballID.LinearVelocity()
	s.DrawScreenTextLinef("p.x = %.9f, v.y = %.9f", p.X, v.Y)

	s.Sample.Step()
}

func CreateRestitutionThreshold(context *SampleContext) Sampler {
	const (
		cameraX              = 7.0
		cameraY              = 5.0
		pixelScale           = 30.0
		restitutionThreshold = 0.1
		blockXPixel          = 205.0
		blockYPixel          = 120.0
		degreesPerHalfTurn   = 180.0
		blockHalfWidthPixel  = 50.0
		blockHalfHeightPixel = 5.0
		ballXPixel           = 200.0
		ballYPixel           = 250.0
		ballRadiusPixel      = 5.0
	)

	s := &RestitutionThreshold{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		ballID:          box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = 6.0
	}

	pixelsPerMeter := float32(pixelScale)

	// With the default threshold the ball will not bounce.
	s.worldID.SetRestitutionThreshold(restitutionThreshold)

	{
		block0BodyDef := box2d.DefaultBodyDef()
		block0BodyDef.Type = box2d.StaticBody
		block0BodyDef.Position = box2d.Pos{
			X: posScalar(blockXPixel / pixelsPerMeter),
			Y: posScalar(blockYPixel / pixelsPerMeter),
		}
		block0BodyDef.Rotation = box2d.MakeRot(70.0 * 3.14 / degreesPerHalfTurn)
		block0BodyID := s.worldID.CreateBody(&block0BodyDef)
		block0Shape := box2d.MakeBox(
			blockHalfWidthPixel/pixelsPerMeter,
			blockHalfHeightPixel/pixelsPerMeter,
		)
		block0ShapeDef := box2d.DefaultShapeDef()
		block0ShapeDef.Material.Friction = 0.0
		block0BodyID.CreatePolygonShape(&block0ShapeDef, &block0Shape)
	}

	{
		// Make a ball
		ballBodyDef := box2d.DefaultBodyDef()
		ballBodyDef.Type = box2d.DynamicBody
		ballBodyDef.Position = box2d.Pos{
			X: posScalar(ballXPixel / pixelsPerMeter),
			Y: posScalar(ballYPixel / pixelsPerMeter),
		}
		s.ballID = s.worldID.CreateBody(&ballBodyDef)

		var ballShape box2d.Circle

		ballShape.Radius = ballRadiusPixel / pixelsPerMeter
		ballShapeDef := box2d.DefaultShapeDef()
		ballShapeDef.Material.Friction = 0.0
		ballShapeDef.Material.Restitution = 1.0
		s.ballID.CreateCircleShape(&ballShapeDef, &ballShape)

		s.ballID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: -2.9})
		s.ballID.SetMotionLocks(box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: true})
	}

	return s
}

func (s *RestitutionThreshold) Step() {
	p := s.ballID.Position()
	v := s.ballID.LinearVelocity()
	s.DrawScreenTextLinef("p.x = %.9f, v.y = %.9f", p.X, v.Y)

	s.Sample.Step()
}

func CreateDrop(context *SampleContext) Sampler {
	const cameraHeight = 1.5

	var human scenes.Human

	s := &Drop{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		groundIDs:   nil,
		bodyIDs:     nil,
		human:       human,
		frameSkip:   0,
		frameCount:  0,
		continuous:  true,
		speculative: true,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 3.0
		s.context.enableSleep = false
		s.context.debugDraw.DrawJoints = false
	}

	s.Scene1()

	return s
}

func (s *Drop) Clear() {
	for _, bodyID := range s.bodyIDs {
		bodyID.Destroy()
	}

	s.bodyIDs = s.bodyIDs[:0]

	if s.human.IsSpawned {
		scenes.DestroyHuman(&s.human)
	}
}

func (s *Drop) CreateGround1() {
	const groundSegmentWidth = 0.25

	s.destroyGround()

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	w := float32(groundSegmentWidth)
	count := 40
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -0.5 * float32(count) * w, Y: 0.0},
		Point2: box2d.Vec2{X: 0.5 * float32(count) * w, Y: 0.0},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.groundIDs = append(s.groundIDs, groundID)
}

func (s *Drop) CreateGround2() {
	const (
		groundBoxWidth      = 0.25
		groundBoxHalfHeight = 0.05
		halfWidthFactor     = 0.5
	)

	s.destroyGround()

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	w := float32(groundBoxWidth)
	count := 40

	x := -0.5 * float32(count) * w

	h := float32(groundBoxHalfHeight)
	for j := 0; j <= count; j++ {
		box := box2d.MakeOffsetBox(
			halfWidthFactor*w,
			h,
			box2d.Vec2{X: x, Y: 0.0},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		x += w
	}

	s.groundIDs = append(s.groundIDs, groundID)
}

func (s *Drop) CreateGround3() {
	const (
		groundSegmentWidth = 0.25
		wallX              = 3.0
		wallHeight         = 8.0
	)

	s.destroyGround()

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	w := float32(groundSegmentWidth)
	count := 40
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -0.5 * float32(count) * w, Y: 0.0},
		Point2: box2d.Vec2{X: 0.5 * float32(count) * w, Y: 0.0},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)
	segment = box2d.Segment{
		Point1: box2d.Vec2{X: wallX, Y: 0.0},
		Point2: box2d.Vec2{X: wallX, Y: wallHeight},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.groundIDs = append(s.groundIDs, groundID)
}

// ball.
func (s *Drop) Scene1() {
	const (
		dropHeight = 4.0
		ballRadius = 0.125
	)

	s.Clear()
	s.CreateGround2()

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: dropHeight}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}

	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
	bodyID.CreateCircleShape(&shapeDef, &circle)

	s.bodyIDs = append(s.bodyIDs, bodyID)
	s.frameCount = 1
}

// ruler.
func (s *Drop) Scene2() {
	const (
		dropHeight          = 4.0
		initialAnglePiRatio = 0.5
		initialAngularSpeed = -0.5
		rulerHalfWidth      = 0.75
		rulerHalfHeight     = 0.01
	)

	s.Clear()
	s.CreateGround1()

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: dropHeight}
	bodyDef.Rotation = box2d.MakeRot(initialAnglePiRatio * box2d.Pi)
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: 0.0}
	bodyDef.AngularVelocity = initialAngularSpeed

	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(rulerHalfWidth, rulerHalfHeight)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	s.bodyIDs = append(s.bodyIDs, bodyID)
	s.frameCount = 1
}

// ragdoll.
func (s *Drop) Scene3() {
	const (
		frictionTorque = 0.03
		dampingRatio   = 0.5
		dropHeight     = 40.0
	)

	s.Clear()
	s.CreateGround2()

	jointFrictionTorque := float32(frictionTorque)
	jointHertz := float32(1.0)
	jointDampingRatio := float32(dampingRatio)

	scenes.CreateHuman(
		&s.human,
		s.worldID,
		box2d.Pos{X: 0.0, Y: dropHeight},
		1.0,
		jointFrictionTorque,
		jointHertz,
		jointDampingRatio,
		1,
		nil,
		true,
	)

	s.frameCount = 1
}

func (s *Drop) Scene4() {
	const (
		boxHalfExtent    = 0.25
		boxStagger       = 0.01
		stackX           = 2.5
		projectileRadius = 0.125
		projectileHeight = 1.9
		projectileSpeed  = 200.0
	)

	s.Clear()
	s.CreateGround3()

	a := float32(boxHalfExtent)
	box := box2d.MakeSquare(a)

	shapeDef := box2d.DefaultShapeDef()

	offset := float32(boxStagger)

	for i := range 5 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shift := offset
		if i%2 == 0 {
			shift = -offset
		}

		bodyDef.Position = box2d.Pos{
			X: posScalar(stackX + shift),
			Y: posScalar(a + 2.0*a*float32(i)),
		}
		bodyID := s.worldID.CreateBody(&bodyDef)

		s.bodyIDs = append(s.bodyIDs, bodyID)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: projectileRadius}
	shapeDef.Density = 4.0

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -7.7, Y: projectileHeight}
		bodyDef.LinearVelocity = box2d.Vec2{X: projectileSpeed, Y: 0.0}
		bodyDef.IsBullet = true

		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCircleShape(&shapeDef, &circle)
		s.bodyIDs = append(s.bodyIDs, bodyID)
	}

	s.frameCount = 1
}

func (s *Drop) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	const skippedFrames = 60

	if key == glfw.Key1 {
		s.Scene1()
		return
	}

	if key == glfw.Key2 {
		s.Scene2()
		return
	}

	if key == glfw.Key3 {
		s.Scene3()
		return
	}

	if key == glfw.Key4 {
		s.Scene4()
		return
	}

	if key == glfw.KeyC {
		s.Clear()
		s.continuous = !s.continuous

		return
	}

	if key == glfw.KeyS {
		if s.frameSkip > 0 {
			s.frameSkip = 0
		} else {
			s.frameSkip = skippedFrames
		}

		return
	}

	s.Sample.Keyboard(key, action, mods)
}

func (s *Drop) Step() {
	s.context.enableContinuous = s.continuous

	if (s.frameSkip == 0 || s.frameCount%s.frameSkip == 0) && !s.context.pause {
		s.Sample.Step()
	} else {
		pause := s.context.pause
		s.context.pause = true
		s.Sample.Step()
		s.context.pause = pause
	}

	s.frameCount += 1
}

func (s *Drop) destroyGround() {
	for _, groundID := range s.groundIDs {
		groundID.Destroy()
	}

	s.groundIDs = s.groundIDs[:0]
}

func CreatePinball(context *SampleContext) Sampler {
	const (
		cameraHeight          = 9.0
		baseCameraZoom        = 25.0
		cameraZoomScale       = 0.5
		tableHalfWidth        = 8.0
		tableSideBottom       = 6.0
		tableTop              = 20.0
		flipperPivotX         = 2.0
		flipperHalfLength     = 1.75
		flipperHalfThickness  = 0.2
		flipperWideLimitDeg   = 30.0
		flipperNarrowLimitDeg = 5.0
		degreesPerHalfTurn    = 180.0
		leftSpinnerHeight     = 17.0
		rightSpinnerX         = 4.0
		rightSpinnerHeight    = 8.0
		spinnerArmHalfLength  = 1.5
		spinnerArmHalfWidth   = 0.125
		leftBumperHeight      = 8.0
		rightBumperX          = 4.0
		rightBumperHeight     = 17.0
		ballHeight            = 15.0
		ballRadius            = 0.2
	)

	s := &Pinball{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		leftJointID:  box2d.RevoluteJoint{Joint: box2d.Joint{}},
		rightJointID: box2d.RevoluteJoint{Joint: box2d.Joint{}},
		ballID:       box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	s.context.debugDraw.DrawJoints = false

	// Ground body
	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)

		vs := []box2d.Vec2{
			{X: -tableHalfWidth, Y: tableSideBottom},
			{X: -tableHalfWidth, Y: tableTop},
			{X: tableHalfWidth, Y: tableTop},
			{X: tableHalfWidth, Y: tableSideBottom},
			{X: 0.0, Y: -2.0},
		}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = vs
		chainDef.IsLoop = true
		groundID.CreateChain(&chainDef)
	}

	// Flippers
	{
		p1, p2 := box2d.Pos{X: -flipperPivotX, Y: 0.0}, box2d.Pos{X: flipperPivotX, Y: 0.0}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.EnableSleep = false

		bodyDef.Position = p1
		leftFlipperID := s.worldID.CreateBody(&bodyDef)

		bodyDef.Position = p2
		rightFlipperID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(flipperHalfLength, flipperHalfThickness)

		shapeDef := box2d.DefaultShapeDef()

		leftFlipperID.CreatePolygonShape(&shapeDef, &box)
		rightFlipperID.CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.LocalFrameB.P = box2d.Vec2Zero
		jointDef.EnableMotor = true
		jointDef.MaxMotorTorque = 1000.0
		jointDef.EnableLimit = true

		jointDef.MotorSpeed = 0.0
		jointDef.Base.LocalFrameA.P = p1.Vec2()
		jointDef.Base.BodyB = leftFlipperID
		jointDef.LowerAngle = -flipperWideLimitDeg * box2d.Pi / degreesPerHalfTurn
		jointDef.UpperAngle = flipperNarrowLimitDeg * box2d.Pi / degreesPerHalfTurn
		s.leftJointID = s.worldID.CreateRevoluteJoint(&jointDef)

		jointDef.MotorSpeed = 0.0
		jointDef.Base.LocalFrameA.P = p2.Vec2()
		jointDef.Base.BodyB = rightFlipperID
		jointDef.LowerAngle = -flipperNarrowLimitDeg * box2d.Pi / degreesPerHalfTurn
		jointDef.UpperAngle = flipperWideLimitDeg * box2d.Pi / degreesPerHalfTurn
		s.rightJointID = s.worldID.CreateRevoluteJoint(&jointDef)
	}

	// Spinners
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -4.0, Y: leftSpinnerHeight}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box1 := box2d.MakeBox(spinnerArmHalfLength, spinnerArmHalfWidth)
		box2 := box2d.MakeBox(spinnerArmHalfWidth, spinnerArmHalfLength)

		bodyID.CreatePolygonShape(&shapeDef, &box1)
		bodyID.CreatePolygonShape(&shapeDef, &box2)

		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = bodyDef.Position.Vec2()
		jointDef.Base.LocalFrameB.P = box2d.Vec2Zero
		jointDef.EnableMotor = true
		jointDef.MaxMotorTorque = 0.1
		s.worldID.CreateRevoluteJoint(&jointDef)

		bodyDef.Position = box2d.Pos{X: rightSpinnerX, Y: rightSpinnerHeight}
		bodyID = s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box1)
		bodyID.CreatePolygonShape(&shapeDef, &box2)

		jointDef.Base.LocalFrameA.P = bodyDef.Position.Vec2()
		jointDef.Base.BodyB = bodyID
		s.worldID.CreateRevoluteJoint(&jointDef)
	}

	// Bumpers
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -4.0, Y: leftBumperHeight}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Restitution = 1.5

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 1.0}
		bodyID.CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: rightBumperX, Y: rightBumperHeight}
		bodyID = s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	// Ball
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 1.0, Y: ballHeight}
		bodyDef.Type = box2d.DynamicBody
		bodyDef.IsBullet = true

		s.ballID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
		s.ballID.CreateCircleShape(&shapeDef, &circle)
	}

	return s
}

func (s *Pinball) DrawControls() bool {
	imTextf("Flipper: press A")
	return true
}

func (s *Pinball) Step() {
	const (
		activeMotorSpeed = 20.0
		restMotorSpeed   = 10.0
	)

	s.Sample.Step()

	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.leftJointID.SetMotorSpeed(activeMotorSpeed)
		s.rightJointID.SetMotorSpeed(-activeMotorSpeed)
	} else {
		s.leftJointID.SetMotorSpeed(-restMotorSpeed)
		s.rightJointID.SetMotorSpeed(restMotorSpeed)
	}
}

func CreateWedge(context *SampleContext) Sampler {
	const (
		cameraHeight = 5.5
		wedgeTop     = 8.0
		dropHeight   = 10.75
	)

	s := &Wedge{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -4.0, Y: wedgeTop},
			Point2: box2d.Vec2{X: 0.0, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
		segment = box2d.Segment{
			Point1: box2d.Vec2{X: 0.0, Y: 0.0},
			Point2: box2d.Vec2{X: 0.0, Y: wedgeTop},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -0.45, Y: dropHeight}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -200.0}

		bodyID := s.worldID.CreateBody(&bodyDef)

		var circle box2d.Circle

		circle.Radius = 0.3
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.2
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	return s
}

func CreateSafetyFactor(context *SampleContext) Sampler {
	const (
		initialExtent       = 0.5
		initialHeight       = 0.3
		initialSafetyFactor = 0.1
		cameraHeight        = 3.0
		groundHalfWidth     = 20.0
	)

	s := &SafetyFactor{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		bodyID:       box2d.Body{},
		extent:       initialExtent,
		height:       initialHeight,
		safetyFactor: initialSafetyFactor,
		overlap:      0,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(groundHalfWidth, 1.0)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.Launch()

	return s
}

func (s *SafetyFactor) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(s.height + s.extent)}
	bodyDef.SafetyFactor = s.safetyFactor
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeSquare(s.extent)
	s.bodyID.CreatePolygonShape(&shapeDef, &box)

	s.overlap = 0.0
}

func (s *SafetyFactor) Step() {
	const overlapLineHalfWidth = 3.0

	s.Sample.Step()

	if s.didStep {
		// Contacts are computed at the beginning of the step, so the manifold holds the initial separation.
		for data := range s.bodyID.ContactData() {
			for j := range data.Manifold.PointCount {
				s.overlap = max(s.overlap, -data.Manifold.Points[j].Separation)
			}

			break
		}
	}

	s.draw.DrawLine(
		box2d.Vec2{X: -overlapLineHalfWidth, Y: -s.overlap}.Pos(),
		box2d.Vec2{X: overlapLineHalfWidth, Y: -s.overlap}.Pos(),
		box2d.ColorRed,
	)

	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}

	motion := timeStep * s.bodyID.LinearVelocity().Length()

	s.DrawScreenTextLinef("actual movement = %.3f m", motion)
	s.DrawScreenTextLinef("fast movement = %.3f m", s.safetyFactor*s.extent)
	s.DrawScreenTextLinef("overlap = %.4f m", s.overlap)
}

func (s *SafetyFactor) DrawControls() bool {
	const (
		controlWidthEm = 10.0
		maxDropHeight  = 2.0
	)

	changed := false

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	changed = imgui.SliderFloatV("Height", &s.height, 0.0, maxDropHeight, "%.2f", 0) || changed
	changed = imgui.SliderFloatV("Safety", &s.safetyFactor, 0.0, 1.0, "%.2f", 0) || changed

	imgui.PopItemWidth()

	if imgui.Button("Drop") || changed {
		s.Launch()
	}

	return true
}

func registerContinuousSamples() {
	RegisterSample("Continuous", "Bounce House", CreateBounceHouse)
	RegisterSample("Continuous", "Bounce Humans", CreateBounceHumans)
	RegisterSample("Continuous", "Chain Drop", CreateChainDrop)
	RegisterSample("Continuous", "Chain Slide", CreateChainSlide)
	RegisterSample("Continuous", "Segment Slide", CreateSegmentSlide)
	RegisterSample("Continuous", "Skinny Box", CreateSkinnyBox)
	RegisterSample("Continuous", "Ghost Bumps", CreateGhostBumps)
	RegisterSample("Continuous", "Speculative Fallback", CreateSpeculativeFallback)
	RegisterSample("Continuous", "Speculative Sliver", CreateSpeculativeSliver)
	RegisterSample("Continuous", "Speculative Ghost", CreateSpeculativeGhost)
	RegisterSample("Continuous", "Pixel Imperfect", CreatePixelImperfect)
	RegisterSample("Continuous", "Restitution Threshold", CreateRestitutionThreshold)
	RegisterSample("Continuous", "Drop", CreateDrop)
	RegisterSample("Continuous", "Pinball", CreatePinball)
	RegisterSample("Continuous", "Wedge", CreateWedge)
	RegisterSample("Continuous", "Safety Factor", CreateSafetyFactor)
}
