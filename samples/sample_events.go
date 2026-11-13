// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const (
	funnelDonut = iota + 1
	funnelHuman
)

const (
	funnelCount         = 32
	funnelSpawnInterval = 0.5
	sensorHitsSpeed     = 0.5
)

const (
	footGround = 0x00000001
	footPlayer = 0x00000002
	footFoot   = 0x00000004
)

const (
	contactEventCount = 20
	bodyMoveCount     = 50
)

const (
	sensorTypesGround  = 0x00000001
	sensorTypesSensor  = 0x00000002
	sensorTypesDefault = 0x00000004
)

const (
	jointEventCount             = 6
	sensorHitsTransformCapacity = 20
)

type SensorBookend struct {
	Sample

	sensorBodyID1  box2d.Body
	sensorShapeID1 box2d.Shape

	sensorBodyID2  box2d.Body
	sensorShapeID2 box2d.Shape

	visitorBodyID  box2d.Body
	visitorShapeID box2d.Shape

	isVisiting1         bool
	isVisiting2         bool
	sensorsOverlapCount int
}
type FootSensor struct {
	Sample

	playerID     box2d.Body
	sensorID     box2d.Shape
	overlapCount int
}
type ContactEvent struct {
	Sample

	playerID     box2d.Body
	coreShapeID  box2d.Shape
	debrisIDs    [contactEventCount]box2d.Body
	bodyUserData [contactEventCount]int
	force        float32
	wait         float32
}

// This shows how to process body events.
type BodyMove struct {
	Sample

	bodyIDs            [bodyMoveCount]box2d.Body
	sleeping           [bodyMoveCount]bool
	count              int
	sleepCount         int
	explosionPosition  box2d.Pos
	explosionRadius    float32
	explosionMagnitude float32
}
type SensorTypes struct {
	Sample

	staticSensorID    box2d.Shape
	kinematicSensorID box2d.Shape
	dynamicSensorID   box2d.Shape

	kinematicBodyID box2d.Body
}

// This sample breaks joints using reaction force events instead of polling.
type JointEvent struct {
	Sample

	jointIDs [jointEventCount]box2d.Joint
}
type PersistentContact struct {
	Sample

	contactID box2d.Contact
}
type SensorHits struct {
	Sample

	staticSensorID    box2d.Shape
	kinematicSensorID box2d.Shape
	dynamicSensorID   box2d.Shape

	kinematicBodyID box2d.Body
	dynamicBodyID   box2d.Body
	jointID         box2d.PrismaticJoint

	bodyID  box2d.Body
	shapeID box2d.Shape

	transformCount int
	transforms     [sensorHitsTransformCapacity]box2d.WorldTransform

	isBullet   bool
	beginCount int
	endCount   int
}

// This shows how to create a projectile that explodes on impact.
type ProjectileEvent struct {
	Sample

	projectileID      box2d.Body
	projectileShapeID box2d.Shape
	point1            box2d.Pos
	point2            box2d.Pos
	dragging          bool
}
type circleImpulseEvent struct {
	impulse      float32
	totalImpulse float32
	speed        float32
}
type CircleImpulse struct {
	Sample

	mass           float32
	events         []circleImpulseEvent
	bodyID         box2d.Body
	gravity        float32
	restitution    float32
	useGravity     bool
	useRestitution bool
}
type SensorFunnel struct {
	Sample

	humans    [funnelCount]scenes.Human
	donuts    [funnelCount]Donut
	isSpawned [funnelCount]bool
	elemType  int
	wait      float32
	side      float32
}

func CreateSensorFunnel(context *SampleContext) Sampler {
	const (
		cameraBaseZoom        = 25.0
		cameraZoomScale       = 1.333
		rightWallX            = 16.8672485
		topY                  = 31.088623
		upperSlopeY           = 17.1978741
		upperShelfY           = 11.906374
		rightShelfX           = 8.26824951
		outletRightX          = 3.63799858
		paddleSpacing         = 14.0
		paddleHalfWidth       = 6.0
		paddleHalfHeight      = 0.5
		paddleMotorSpeed      = 2.0
		outletSensorHalfWidth = 4.0
		initialSpawnSide      = 15.0
	)

	s := &SensorFunnel{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		humans:    [funnelCount]scenes.Human{},
		donuts:    [funnelCount]Donut{},
		isSpawned: [funnelCount]bool{},
		elemType:  funnelHuman,
		wait:      funnelSpawnInterval,
		side:      -initialSpawnSide,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	s.context.debugDraw.DrawJoints = false

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		points := []box2d.Vec2{
			{X: -16.8672504, Y: topY},
			{X: rightWallX, Y: topY},
			{X: rightWallX, Y: upperSlopeY},
			{X: rightShelfX, Y: upperShelfY},
			{X: rightWallX, Y: upperShelfY},
			{X: rightWallX, Y: -0.661376953},
			{X: rightShelfX, Y: -5.953125},
			{X: rightWallX, Y: -5.953125},
			{X: rightWallX, Y: -13.229126},
			{X: outletRightX, Y: -23.151123},
			{X: outletRightX, Y: -topY},
			{X: -3.63800049, Y: -31.088623},
			{X: -3.63800049, Y: -23.151123},
			{X: -16.8672504, Y: -13.229126},
			{X: -16.8672504, Y: -5.953125},
			{X: -8.26825142, Y: -5.953125},
			{X: -16.8672504, Y: -0.661376953},
			{X: -16.8672504, Y: upperShelfY},
			{X: -8.26825142, Y: upperShelfY},
			{X: -16.8672504, Y: upperSlopeY},
		}

		var material box2d.SurfaceMaterial

		material.Friction = 0.2

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}
		groundID.CreateChain(&chainDef)

		sign := float32(1.0)

		y := float32(paddleSpacing)
		for range 3 {
			bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(y)}
			bodyDef.Type = box2d.DynamicBody

			bodyID := s.worldID.CreateBody(&bodyDef)

			box := box2d.MakeBox(paddleHalfWidth, paddleHalfHeight)
			shapeDef := box2d.DefaultShapeDef()
			shapeDef.Material.Friction = 0.1
			shapeDef.Material.Restitution = 1.0
			shapeDef.Density = 1.0

			bodyID.CreatePolygonShape(&shapeDef, &box)

			revoluteDef := box2d.DefaultRevoluteJointDef()
			revoluteDef.Base.BodyA = groundID
			revoluteDef.Base.BodyB = bodyID
			revoluteDef.Base.LocalFrameA.P = bodyDef.Position.Vec2()
			revoluteDef.Base.LocalFrameB.P = box2d.Vec2Zero
			revoluteDef.MaxMotorTorque = 200.0
			revoluteDef.MotorSpeed = paddleMotorSpeed * sign
			revoluteDef.EnableMotor = true

			s.worldID.CreateRevoluteJoint(&revoluteDef)

			y -= paddleSpacing
			sign = -sign
		}

		{
			box := box2d.MakeOffsetBox(
				outletSensorHalfWidth,
				1.0,
				box2d.Vec2{X: 0.0, Y: -30.5},
				box2d.RotIdentity,
			)
			shapeDef := box2d.DefaultShapeDef()
			shapeDef.IsSensor = true
			shapeDef.EnableSensorEvents = true

			groundID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	s.wait = funnelSpawnInterval
	s.side = -initialSpawnSide
	s.elemType = funnelHuman

	s.CreateElement()

	return s
}

func (s *SensorFunnel) CreateElement() {
	const (
		spawnHeight        = 29.5
		humanScale         = 2.0
		humanJointFriction = 0.05
		humanJointHertz    = 6.0
		humanJointDamping  = 0.5
	)

	index := -1

	for i := range funnelCount {
		if !s.isSpawned[i] {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	center := box2d.Pos{X: posScalar(s.side), Y: spawnHeight}

	if s.elemType == funnelDonut {
		donut := &s.donuts[index]
		donut.Create(s.worldID, center, 1.0, 0, true, donut)
	} else {
		human := &s.humans[index]
		scale := float32(humanScale)
		jointFriction := float32(humanJointFriction)
		jointHertz := float32(humanJointHertz)
		jointDamping := float32(humanJointDamping)
		colorize := true
		scenes.CreateHuman(
			human,
			s.worldID,
			center,
			scale,
			jointFriction,
			jointHertz,
			jointDamping,
			index+1,
			human,
			colorize,
		)
		human.EnableSensorEvents(true)
	}

	s.isSpawned[index] = true
	s.side = -s.side
}

func (s *SensorFunnel) DestroyElement(index int) {
	if s.elemType == funnelDonut {
		s.donuts[index].Destroy()
	} else {
		scenes.DestroyHuman(&s.humans[index])
	}

	s.isSpawned[index] = false
}

func (s *SensorFunnel) Clear() {
	for i := range funnelCount {
		if s.isSpawned[i] {
			if s.elemType == funnelDonut {
				s.donuts[i].Destroy()
			} else {
				scenes.DestroyHuman(&s.humans[i])
			}

			s.isSpawned[i] = false
		}
	}
}

func (s *SensorFunnel) DrawControls() bool {
	if imgui.RadioButtonBool("donut", s.elemType == funnelDonut) {
		s.Clear()
		s.elemType = funnelDonut
	}

	if imgui.RadioButtonBool("human", s.elemType == funnelHuman) {
		s.Clear()
		s.elemType = funnelHuman
	}

	return true
}

func (s *SensorFunnel) Step() {
	s.Sample.Step()
	s.destroyExitedElements()

	if s.context.hertz > 0.0 && !s.context.pause {
		s.wait -= 1.0 / s.context.hertz
		if s.wait < 0.0 {
			s.CreateElement()
			s.wait += funnelSpawnInterval
		}
	}
}

func (s *SensorFunnel) destroyExitedElements() {
	// Discover rings that touch the bottom sensor
	var deferredDestruction [funnelCount]bool

	s.markExitedElements(&deferredDestruction)

	// Safely destroy rings that hit the bottom sensor
	for i := range funnelCount {
		if deferredDestruction[i] {
			s.DestroyElement(i)
		}
	}
}

func (s *SensorFunnel) markExitedElements(deferredDestruction *[funnelCount]bool) {
	for event := range s.worldID.SensorBeginEvents() {
		bodyID := event.VisitorShape.Body()

		if s.elemType == funnelDonut {
			s.markExitedDonut(bodyID, deferredDestruction)
		} else {
			s.markExitedHuman(bodyID, deferredDestruction)
		}
	}
}

func (s *SensorFunnel) markExitedDonut(bodyID box2d.Body, deferredDestruction *[funnelCount]bool) {
	donut, ok := bodyID.UserData().(*Donut)
	if !ok || donut == nil {
		return
	}

	for index := range funnelCount {
		if &s.donuts[index] == donut {
			deferredDestruction[index] = true
		}
	}
}

func (s *SensorFunnel) markExitedHuman(bodyID box2d.Body, deferredDestruction *[funnelCount]bool) {
	human, ok := bodyID.UserData().(*scenes.Human)
	if !ok || human == nil {
		return
	}

	for index := range funnelCount {
		if &s.humans[index] == human {
			deferredDestruction[index] = true
		}
	}
}

func CreateSensorBookend(context *SampleContext) Sampler {
	const (
		cameraHeight  = 6.0
		roomHalfWidth = 10.0
		roomHeight    = 10.0
	)

	s := &SensorBookend{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		sensorBodyID1:       box2d.Body{},
		sensorShapeID1:      box2d.Shape{},
		sensorBodyID2:       box2d.Body{},
		sensorShapeID2:      box2d.Shape{},
		visitorBodyID:       box2d.Body{},
		visitorShapeID:      box2d.Shape{},
		isVisiting1:         false,
		isVisiting2:         false,
		sensorsOverlapCount: 0,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 7.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		groundSegment := box2d.Segment{
			Point1: box2d.Vec2{X: -roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: roomHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{
			Point1: box2d.Vec2{X: -roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: -roomHalfWidth, Y: roomHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{
			Point1: box2d.Vec2{X: roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: roomHalfWidth, Y: roomHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		s.isVisiting1 = false
		s.isVisiting2 = false
		s.sensorsOverlapCount = 0
	}

	s.CreateSensor1()
	s.CreateSensor2()
	s.CreateVisitor()

	return s
}

func (s *SensorBookend) CreateSensor1() {
	bodyDef := box2d.DefaultBodyDef()

	bodyDef.Position = box2d.Pos{X: -2.0, Y: 1.0}
	s.sensorBodyID1 = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	shapeDef.EnableSensorEvents = true

	box := box2d.MakeSquare(1.0)
	s.sensorShapeID1 = s.sensorBodyID1.CreatePolygonShape(&shapeDef, &box)
}

func (s *SensorBookend) CreateSensor2() {
	const (
		sensorPositionX    = 2.0
		sensorHalfExtent   = 0.5
		sensorCornerRadius = 0.5
		solidHalfExtent    = 0.5
	)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: sensorPositionX, Y: 1.0}
	s.sensorBodyID2 = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	shapeDef.EnableSensorEvents = true

	box := box2d.MakeRoundedBox(sensorHalfExtent, sensorHalfExtent, sensorCornerRadius)
	s.sensorShapeID2 = s.sensorBodyID2.CreatePolygonShape(&shapeDef, &box)

	// Solid middle
	shapeDef.IsSensor = false
	shapeDef.EnableSensorEvents = false
	box = box2d.MakeSquare(solidHalfExtent)
	s.sensorBodyID2.CreatePolygonShape(&shapeDef, &box)
}

func (s *SensorBookend) CreateVisitor() {
	const visitorRadius = 0.5

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: -4.0, Y: 1.0}
	bodyDef.Type = box2d.DynamicBody

	s.visitorBodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: visitorRadius}
	s.visitorShapeID = s.visitorBodyID.CreateCircleShape(&shapeDef, &circle)
}

func (s *SensorBookend) DrawControls() bool {
	switch {
	case s.visitorBodyID == (box2d.Body{}):
		if imgui.Button("create visitor") {
			s.CreateVisitor()
		}
	case imgui.Button("destroy visitor"):
		s.visitorBodyID.Destroy()
		s.visitorBodyID = box2d.Body{}
		// Retain visitorShapeID for end events.
	default:
		drawSensorEventsCheckbox("visitor events", s.visitorShapeID)
		drawBodyEnableCheckbox("enable visitor body", s.visitorBodyID)
	}

	imgui.Separator()

	s.drawSensorControls()

	return true
}

func (s *SensorBookend) Step() {
	s.Sample.Step()

	for event := range s.worldID.SensorBeginEvents() {
		switch {
		case event.VisitorShape != s.visitorShapeID:
			s.sensorsOverlapCount += 1
		case event.SensorShape == s.sensorShapeID1:
			s.isVisiting1 = true
		default:
			s.isVisiting2 = true
		}
	}

	for event := range s.worldID.SensorEndEvents() {
		switch {
		case event.VisitorShape != s.visitorShapeID:
			s.sensorsOverlapCount -= 1
		case event.SensorShape == s.sensorShapeID1:
			s.isVisiting1 = false
		default:
			s.isVisiting2 = false
		}
	}

	// Nullify invalid shape ids after end events are processed.
	if !s.visitorShapeID.IsValid() {
		s.visitorShapeID = box2d.Shape{}
	}

	if !s.sensorShapeID1.IsValid() {
		s.sensorShapeID1 = box2d.Shape{}
	}

	if !s.sensorShapeID2.IsValid() {
		s.sensorShapeID2 = box2d.Shape{}
	}

	s.DrawScreenTextLinef("visiting 1 == %t", s.isVisiting1)
	s.DrawScreenTextLinef("visiting 2 == %t", s.isVisiting2)
	s.DrawScreenTextLinef("sensors overlap count == %d", s.sensorsOverlapCount)
}

func CreateFootSensor(context *SampleContext) Sampler {
	const (
		cameraHeight     = 6.0
		firstGroundX     = 9.0
		firstGhostX      = 10.0
		playerHalfLength = 0.5
		playerRadius     = 0.5
		footHalfWidth    = 0.5
		footHalfHeight   = 0.25
	)

	s := &FootSensor{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		playerID:     box2d.Body{},
		sensorID:     box2d.Shape{},
		overlapCount: 0,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 7.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		var points [18]box2d.Vec2

		x := float32(firstGroundX)
		for i := range 18 {
			points[i] = box2d.Vec2{X: x, Y: 0.0}
			x -= 1.0
		}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.Ghost1 = box2d.Vec2{X: firstGhostX, Y: 0.0}
		chainDef.Ghost2 = box2d.Vec2{X: -firstGroundX, Y: 0.0}
		chainDef.Filter.CategoryBits = footGround
		chainDef.Filter.MaskBits = footFoot | footPlayer
		chainDef.IsLoop = false
		chainDef.EnableSensorEvents = true

		groundID.CreateChain(&chainDef)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.MotionLocks.AngularZ = true
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
		s.playerID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = footPlayer
		shapeDef.Filter.MaskBits = footGround
		shapeDef.Material.Friction = 0.3
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -playerHalfLength},
			Center2: box2d.Vec2{X: 0.0, Y: playerHalfLength},
			Radius:  playerRadius,
		}
		s.playerID.CreateCapsuleShape(&shapeDef, &capsule)

		box := box2d.MakeOffsetBox(
			footHalfWidth,
			footHalfHeight,
			box2d.Vec2{X: 0.0, Y: -1.0},
			box2d.RotIdentity,
		)
		shapeDef.Filter.CategoryBits = footFoot
		shapeDef.Filter.MaskBits = footGround
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		s.sensorID = s.playerID.CreatePolygonShape(&shapeDef, &box)
	}

	s.overlapCount = 0

	return s
}

func (s *FootSensor) Step() {
	const (
		playerForce      = 50.0
		overlapPointSize = 10.0
	)
	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.playerID.ApplyForceToCenter(box2d.Vec2{X: -playerForce, Y: 0.0}, true)
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.playerID.ApplyForceToCenter(box2d.Vec2{X: playerForce, Y: 0.0}, true)
	}

	s.Sample.Step()

	for event := range s.worldID.SensorBeginEvents() {
		if event.SensorShape == s.sensorID {
			s.overlapCount += 1
		}
	}

	for event := range s.worldID.SensorEndEvents() {
		if event.SensorShape == s.sensorID {
			s.overlapCount -= 1
		}
	}

	s.DrawScreenTextLinef("count == %d", s.overlapCount)

	for shapeID := range s.sensorID.SensorData() {
		aabb := shapeID.AABB()
		point := aabb.Center()
		s.draw.DrawPoint(point.Pos(), overlapPointSize, box2d.ColorWhite)
	}
}

func CreateContactEvent(context *SampleContext) Sampler {
	const (
		cameraBaseZoom     = 25.0
		cameraZoomScale    = 1.75
		arenaHalfExtent    = 40.0
		initialPlayerForce = 200.0
		initialSpawnWait   = 0.5
	)

	s := &ContactEvent{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		playerID:     box2d.Body{},
		coreShapeID:  box2d.Shape{},
		debrisIDs:    [contactEventCount]box2d.Body{},
		bodyUserData: [contactEventCount]int{},
		force:        initialPlayerForce,
		wait:         initialSpawnWait,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		points := []box2d.Vec2{
			{X: arenaHalfExtent, Y: -arenaHalfExtent},
			{X: -arenaHalfExtent, Y: -arenaHalfExtent},
			{X: -arenaHalfExtent, Y: arenaHalfExtent},
			{X: arenaHalfExtent, Y: arenaHalfExtent},
		}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points
		chainDef.IsLoop = true

		groundID.CreateChain(&chainDef)
	}

	// Player
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.GravityScale = 0.0
		bodyDef.LinearDamping = 0.5
		bodyDef.AngularDamping = 0.5
		bodyDef.IsBullet = true
		s.playerID = s.worldID.CreateBody(&bodyDef)

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 1.0}
		shapeDef := box2d.DefaultShapeDef()

		// Enable contact events for the player shape
		shapeDef.EnableContactEvents = true

		s.coreShapeID = s.playerID.CreateCircleShape(&shapeDef, &circle)
	}

	for i := range contactEventCount {
		s.debrisIDs[i] = box2d.Body{}
		s.bodyUserData[i] = i
	}

	s.wait = 0.5
	s.force = 200.0

	return s
}

func (s *ContactEvent) SpawnDebris() {
	const (
		spawnExtent       = 38.0
		maxInitialSpeed   = 5.0
		circleRadius      = 0.5
		capsuleHalfLength = 0.25
		capsuleRadius     = 0.25
		boxHalfWidth      = 0.4
		boxHalfHeight     = 0.6
	)

	index := -1

	for i := range contactEventCount {
		if s.debrisIDs[i] == (box2d.Body{}) {
			index = i
			break
		}
	}

	if index == -1 {
		return
	}

	// Debris
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	x := scenes.RandomFloatRange(-spawnExtent, spawnExtent)
	y := scenes.RandomFloatRange(-spawnExtent, spawnExtent)
	bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
	bodyDef.Rotation = box2d.MakeRot(scenes.RandomFloatRange(-box2d.Pi, box2d.Pi))
	vx := scenes.RandomFloatRange(-maxInitialSpeed, maxInitialSpeed)
	vy := scenes.RandomFloatRange(-maxInitialSpeed, maxInitialSpeed)
	bodyDef.LinearVelocity = box2d.Vec2{X: vx, Y: vy}
	bodyDef.AngularVelocity = scenes.RandomFloatRange(-1.0, 1.0)
	bodyDef.GravityScale = 0.0
	bodyDef.UserData = &s.bodyUserData[index]
	s.debrisIDs[index] = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 0.8

	// No events when debris hits debris
	shapeDef.EnableContactEvents = false

	switch {
	case (index+1)%3 == 0:
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
		s.debrisIDs[index].CreateCircleShape(&shapeDef, &circle)
	case (index+1)%2 == 0:
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -capsuleHalfLength},
			Center2: box2d.Vec2{X: 0.0, Y: capsuleHalfLength},
			Radius:  capsuleRadius,
		}
		s.debrisIDs[index].CreateCapsuleShape(&shapeDef, &capsule)
	default:
		box := box2d.MakeBox(boxHalfWidth, boxHalfHeight)
		s.debrisIDs[index].CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *ContactEvent) DrawControls() bool {
	const (
		controlWidth = 6.0
		minForce     = 100.0
		maxForce     = 500.0
	)
	imgui.PushItemWidth(controlWidth * imgui.FontSize())
	imgui.SliderFloatV("force", &s.force, minForce, maxForce, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *ContactEvent) Step() {
	s.DrawScreenTextLinef("move using WASD")

	position := s.playerID.Position()

	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.playerID.ApplyForce(box2d.Vec2{X: -s.force, Y: 0.0}, position, true)
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.playerID.ApplyForce(box2d.Vec2{X: s.force, Y: 0.0}, position, true)
	}

	if s.context.window.GetKey(glfw.KeyW) == glfw.Press {
		s.playerID.ApplyForce(box2d.Vec2{X: 0.0, Y: s.force}, position, true)
	}

	if s.context.window.GetKey(glfw.KeyS) == glfw.Press {
		s.playerID.ApplyForce(box2d.Vec2{X: 0.0, Y: -s.force}, position, true)
	}

	s.Sample.Step()
	s.processContactEvents()

	if s.context.hertz > 0.0 && !s.context.pause {
		s.wait -= 1.0 / s.context.hertz
		if s.wait < 0.0 {
			s.SpawnDebris()
			s.wait += 0.5
		}
	}
}

func (s *ContactEvent) processContactEvents() {
	var (
		debrisToAttach  [contactEventCount]int
		shapesToDestroy [contactEventCount]box2d.Shape
	)

	attachCount := 0
	destroyCount := 0

	// Process contact begin touch events.
	for event := range s.worldID.ContactBeginEvents() {
		bodyIDA := event.ShapeA.Body()
		bodyIDB := event.ShapeB.Body()

		// The begin touch events have the contact manifolds, but the impulses are zero. This is because the manifolds
		// are gathered before the contact solver is run.

		// Both shapes share the final contact manifold. Read it from the shape with fewer contacts.

		capacityA := event.ShapeA.ContactCapacity()
		capacityB := event.ShapeB.ContactCapacity()

		if capacityA < capacityB {
			s.drawEventContacts(event.ShapeA, event.ShapeB, box2d.ColorBlueViolet)
		} else {
			s.drawEventContacts(event.ShapeB, event.ShapeA, box2d.ColorYellowGreen)
		}

		addShapeToDestroy := func(shapeID box2d.Shape) {
			if shapeID == s.coreShapeID || destroyCount >= contactEventCount ||
				slices.Contains(shapesToDestroy[:destroyCount], shapeID) {
				return
			}

			shapesToDestroy[destroyCount] = shapeID
			destroyCount += 1
		}

		playerShape, otherBody := event.ShapeB, bodyIDA
		if bodyIDA == s.playerID {
			playerShape, otherBody = event.ShapeA, bodyIDB
		}

		userData, _ := otherBody.UserData().(*int)
		if userData == nil {
			addShapeToDestroy(playerShape)
		} else if attachCount < contactEventCount {
			debrisToAttach[attachCount] = *userData
			attachCount += 1
		}
	}

	s.attachDebris(debrisToAttach[:attachCount])

	for i := range destroyCount {
		updateMass := false
		shapesToDestroy[i].Destroy(updateMass)
	}

	if destroyCount > 0 {
		// Update mass just once
		s.playerID.UpdateMassFromShapes()
	}
}

func (s *ContactEvent) attachDebris(debrisToAttach []int) {
	// Attach debris to player body
	for i := range debrisToAttach {
		index := debrisToAttach[i]

		debrisID := s.debrisIDs[index]
		if !debrisID.IsValid() {
			continue
		}

		playerTransform := s.playerID.Transform()
		debrisTransform := debrisID.Transform()
		relativeTransform := playerTransform.InvMul(debrisTransform)

		shapeCount := debrisID.ShapeCount()
		if shapeCount == 0 {
			continue
		}

		var shapeID box2d.Shape
		for shape := range debrisID.Shapes() {
			shapeID = shape
			break
		}

		shapeType := shapeID.Type()

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableContactEvents = true

		switch shapeType {
		case box2d.CircleShape:
			circle := shapeID.Circle()
			circle.Center = relativeTransform.Apply(circle.Center)

			s.playerID.CreateCircleShape(&shapeDef, &circle)

		case box2d.CapsuleShape:
			capsule := shapeID.Capsule()
			capsule.Center1 = relativeTransform.Apply(capsule.Center1)
			capsule.Center2 = relativeTransform.Apply(capsule.Center2)

			s.playerID.CreateCapsuleShape(&shapeDef, &capsule)

		case box2d.PolygonShape:
			originalPolygon := shapeID.Polygon()
			polygon := box2d.TransformPolygon(relativeTransform, &originalPolygon)

			s.playerID.CreatePolygonShape(&shapeDef, &polygon)

		case box2d.SegmentShape, box2d.ChainSegmentShape, box2d.ShapeTypeCount:
			panic("unexpected shape type")
		default:
			panic("unexpected shape type")
		}

		debrisID.Destroy()

		s.debrisIDs[index] = box2d.Body{}
	}
}

func (s *ContactEvent) drawEventContacts(
	shapeID box2d.Shape,
	otherShapeID box2d.Shape,
	color box2d.HexColor,
) {
	const contactPointSize = 10.0
	// The count may be less than the capacity
	for data := range shapeID.ContactData() {
		idA := data.ShapeA

		idB := data.ShapeB
		if idA == otherShapeID || idB == otherShapeID {
			manifold := data.Manifold
			normal := manifold.Normal
			bidA := idA.Body()

			for k := range manifold.PointCount {
				point := manifold.Points[k]
				posA := bidA.WorldCenter()
				p1 := posA.Add(point.AnchorA)
				p2 := p1.Add(normal.Scale(point.TotalNormalImpulse))

				s.draw.DrawLine(p1, p2, color)
				s.draw.DrawPoint(p1, contactPointSize, box2d.ColorWhite)
			}
		}
	}
}

func CreateBodyMove(context *SampleContext) Sampler {
	const (
		cameraX                   = 2.0
		cameraY                   = 8.0
		cameraBaseZoom            = 25.0
		cameraZoomScale           = 0.55
		rampHalfWidth             = 12.0
		wallHalfThickness         = 0.1
		rampCenterX               = 10.0
		rampAngle                 = 0.15
		wallHalfHeight            = 10.0
		wallCenterX               = 19.9
		ceilingHalfWidth          = 20.0
		ceilingY                  = 20.1
		initialExplosionRadius    = 10.0
		initialExplosionMagnitude = 10.0
	)

	s := &BodyMove{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		bodyIDs:            [bodyMoveCount]box2d.Body{},
		sleeping:           [bodyMoveCount]bool{},
		count:              0,
		sleepCount:         0,
		explosionPosition:  box2d.Pos{X: 0.0, Y: -5.0},
		explosionRadius:    initialExplosionRadius,
		explosionMagnitude: initialExplosionMagnitude,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraX, Y: cameraY}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.1

		box := box2d.MakeOffsetBox(
			rampHalfWidth,
			wallHalfThickness,
			box2d.Vec2{X: -rampCenterX, Y: -wallHalfThickness},
			box2d.MakeRot(-rampAngle*box2d.Pi),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			rampHalfWidth,
			wallHalfThickness,
			box2d.Vec2{X: rampCenterX, Y: -wallHalfThickness},
			box2d.MakeRot(rampAngle*box2d.Pi),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		shapeDef.Material.Restitution = 0.8

		box = box2d.MakeOffsetBox(
			wallHalfThickness, wallHalfHeight,
			box2d.Vec2{X: wallCenterX, Y: wallHalfHeight}, box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			wallHalfThickness, wallHalfHeight,
			box2d.Vec2{X: -wallCenterX, Y: wallHalfHeight}, box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			ceilingHalfWidth, wallHalfThickness, box2d.Vec2{X: 0.0, Y: ceilingY}, box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.sleepCount = 0
	s.count = 0

	s.explosionPosition = box2d.Pos{X: 0.0, Y: -5.0}
	s.explosionRadius = 10.0
	s.explosionMagnitude = 10.0

	return s
}

func (s *BodyMove) CreateBodies() {
	const (
		capsuleHalfLength   = 0.25
		capsuleRadius       = 0.25
		circleRadius        = 0.35
		squareHalfExtent    = 0.35
		spawnHeight         = 10.0
		bulletInterval      = 12
		shapeCount          = 4
		squareShapeIndex    = 2
		randomPolygonRadius = 0.75
	)

	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
		Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
		Radius:  capsuleRadius,
	}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
	square := box2d.MakeSquare(squareHalfExtent)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()

	x, y := float32(-5.0), float32(spawnHeight)
	for i := 0; i < 10 && s.count < bodyMoveCount; i++ {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
		bodyDef.IsBullet = s.count%bulletInterval == 0
		bodyDef.UserData = s.count
		s.bodyIDs[s.count] = s.worldID.CreateBody(&bodyDef)
		s.sleeping[s.count] = false

		remainder := s.count % shapeCount
		switch remainder {
		case 0:
			s.bodyIDs[s.count].CreateCapsuleShape(&shapeDef, &capsule)
		case 1:
			s.bodyIDs[s.count].CreateCircleShape(&shapeDef, &circle)
		case squareShapeIndex:
			s.bodyIDs[s.count].CreatePolygonShape(&shapeDef, &square)
		default:
			poly := scenes.RandomPolygon(randomPolygonRadius)
			poly.Radius = 0.1
			s.bodyIDs[s.count].CreatePolygonShape(&shapeDef, &poly)
		}

		s.count += 1
		x += 1.0
	}
}

func (s *BodyMove) DrawControls() bool {
	const (
		controlWidth = 6.0
		maxMagnitude = 20.0
	)

	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()
		def.Position = s.explosionPosition
		def.Radius = s.explosionRadius
		def.Falloff = 0.1
		def.ImpulsePerLength = s.explosionMagnitude
		s.worldID.Explode(&def)
	}

	imgui.PushItemWidth(controlWidth * imgui.FontSize())
	imgui.SliderFloatV("Magnitude", &s.explosionMagnitude, -maxMagnitude, maxMagnitude, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *BodyMove) Step() {
	const spawnStepMask = 15
	if !s.context.pause && (s.stepCount&spawnStepMask) == spawnStepMask && s.count < bodyMoveCount {
		s.CreateBodies()
	}

	s.Sample.Step()

	// Process body events
	for event := range s.worldID.BodyMoveEvents() {
		index, ok := event.UserData.(int)
		if !ok {
			// The mouse joint body has no user data
			continue
		}

		// draw the transform of every body that moved (not sleeping)
		s.draw.DrawTransform(event.Transform, 1.0)

		// this shows a somewhat contrived way to track body sleeping
		sleeping := &s.sleeping[index]

		if event.FellAsleep {
			*sleeping = true
			s.sleepCount += 1
		} else if *sleeping {
			*sleeping = false
			s.sleepCount -= 1
		}
	}

	s.draw.DrawCircle(s.explosionPosition, s.explosionRadius, box2d.ColorAzure)

	s.DrawScreenTextLinef("sleep count: %d", s.sleepCount)
}

func CreateSensorTypes(context *SampleContext) Sampler {
	const (
		cameraHeight           = 3.0
		roomHalfWidth          = 6.0
		roomHeight             = 4.0
		staticSensorHeight     = 0.8
		dynamicSensorX         = 3.0
		dynamicSolidHalfExtent = 0.8
		ballRadius             = 0.5
	)

	s := &SensorTypes{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		staticSensorID:    box2d.Shape{},
		kinematicSensorID: box2d.Shape{},
		dynamicSensorID:   box2d.Shape{},
		kinematicBodyID:   box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 4.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "ground"

		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		// Enable sensor events, but filter them out as a test
		shapeDef.Filter.CategoryBits = sensorTypesGround
		shapeDef.Filter.MaskBits = sensorTypesDefault
		shapeDef.EnableSensorEvents = true

		groundSegment := box2d.Segment{
			Point1: box2d.Vec2{X: -roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: roomHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{
			Point1: box2d.Vec2{X: -roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: -roomHalfWidth, Y: roomHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{
			Point1: box2d.Vec2{X: roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: roomHalfWidth, Y: roomHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "static sensor"
		bodyDef.Type = box2d.StaticBody
		bodyDef.Position = box2d.Pos{X: -3.0, Y: staticSensorHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = sensorTypesSensor
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		box := box2d.MakeSquare(1.0)
		s.staticSensorID = bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "kinematic sensor"
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: 1.0}
		s.kinematicBodyID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = sensorTypesSensor
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		box := box2d.MakeSquare(1.0)
		s.kinematicSensorID = s.kinematicBodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "dynamic sensor"
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: dynamicSensorX, Y: 1.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = sensorTypesSensor
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		box := box2d.MakeSquare(1.0)
		s.dynamicSensorID = bodyID.CreatePolygonShape(&shapeDef, &box)

		// Add some real collision so the dynamic body is valid
		shapeDef.Filter.CategoryBits = sensorTypesDefault
		shapeDef.IsSensor = false
		shapeDef.EnableSensorEvents = false
		box = box2d.MakeSquare(dynamicSolidHalfExtent)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "ball_01"
		bodyDef.Position = box2d.Pos{X: -5.0, Y: 1.0}
		bodyDef.Type = box2d.DynamicBody

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = sensorTypesDefault
		shapeDef.Filter.MaskBits = sensorTypesGround | sensorTypesDefault | sensorTypesSensor
		shapeDef.EnableSensorEvents = true

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	return s
}

func (s *SensorTypes) PrintOverlaps(sensorShapeID box2d.Shape, prefix string) {
	var buffer strings.Builder
	fmt.Fprintf(&buffer, "%s: ", prefix)

	for visitorID := range sensorShapeID.SensorData() {
		if !visitorID.IsValid() {
			continue
		}

		bodyID := visitorID.Body()
		name := bodyID.Name()

		fmt.Fprintf(&buffer, "%s, ", name)
	}

	s.DrawScreenTextLinef("%s", buffer.String())
}

func (s *SensorTypes) Step() {
	const (
		kinematicUpperLimit = 3.0
		rayOriginX          = 5.0
		hitPointSize        = 10.0
	)

	position := s.kinematicBodyID.Position()
	if position.Y < 0.0 {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: 1.0})
	} else if position.Y > kinematicUpperLimit {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: -1.0})
	}

	s.Sample.Step()

	s.PrintOverlaps(s.staticSensorID, "static")
	s.PrintOverlaps(s.kinematicSensorID, "kinematic")
	s.PrintOverlaps(s.dynamicSensorID, "dynamic")

	origin := box2d.Pos{X: rayOriginX, Y: 1.0}
	translation := box2d.Vec2{X: -10.0, Y: 0.0}
	result := s.worldID.CastRayClosest(origin, translation, box2d.DefaultQueryFilter())
	s.draw.DrawLine(origin, origin.Add(translation), box2d.ColorDimGray)

	if result.Hit {
		s.draw.DrawPoint(result.Point, hitPointSize, box2d.ColorCyan)
	}
}

func CreateJointEvent(context *SampleContext) Sampler {
	const (
		cameraHeight         = 8.0
		cameraBaseZoom       = 25.0
		cameraZoomScale      = 0.7
		groundHalfWidth      = 40.0
		initialBodyHeight    = 10.0
		jointForceThreshold  = 20000.0
		jointTorqueThreshold = 10000.0
		distanceJointLength  = 2.0
	)

	s := &JointEvent{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		jointIDs: [jointEventCount]box2d.Joint{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
		Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	position := box2d.Pos{X: -12.5, Y: initialBodyHeight}
	bodyDef.Type = box2d.DynamicBody
	bodyDef.EnableSleep = false

	box := box2d.MakeBox(1.0, 1.0)

	index := 0

	forceThreshold := float32(jointForceThreshold)
	torqueThreshold := float32(jointTorqueThreshold)

	// distance joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		length := float32(distanceJointLength)
		pivot1 := box2d.Pos{X: position.X, Y: position.Y + 1.0 + posScalar(length)}
		pivot2 := box2d.Pos{X: position.X, Y: position.Y + 1.0}
		jointDef := box2d.DefaultDistanceJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot1)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot2)
		jointDef.Length = length
		jointDef.Base.ForceThreshold = forceThreshold
		jointDef.Base.TorqueThreshold = torqueThreshold
		jointDef.Base.CollideConnected = true
		jointDef.Base.UserData = index
		s.jointIDs[index] = s.worldID.CreateDistanceJoint(&jointDef).Joint
	}

	position.X += 5.0
	index++

	// motor joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultMotorJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = position.Vec2()
		jointDef.MaxVelocityForce = 1000.0
		jointDef.MaxVelocityTorque = 20.0
		jointDef.Base.ForceThreshold = forceThreshold
		jointDef.Base.TorqueThreshold = torqueThreshold
		jointDef.Base.CollideConnected = true
		jointDef.Base.UserData = index
		s.jointIDs[index] = s.worldID.CreateMotorJoint(&jointDef).Joint
	}

	position.X += 5.0
	index++

	// prismatic joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultPrismaticJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Base.ForceThreshold = forceThreshold
		jointDef.Base.TorqueThreshold = torqueThreshold
		jointDef.Base.CollideConnected = true
		jointDef.Base.UserData = index
		s.jointIDs[index] = s.worldID.CreatePrismaticJoint(&jointDef).Joint
	}

	position.X += 5.0
	index++

	// revolute joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Base.ForceThreshold = forceThreshold
		jointDef.Base.TorqueThreshold = torqueThreshold
		jointDef.Base.CollideConnected = true
		jointDef.Base.UserData = index
		s.jointIDs[index] = s.worldID.CreateRevoluteJoint(&jointDef).Joint
	}

	position.X += 5.0
	index++

	// weld joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultWeldJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.AngularHertz = 2.0
		jointDef.AngularDampingRatio = 0.5
		jointDef.Base.ForceThreshold = forceThreshold
		jointDef.Base.TorqueThreshold = torqueThreshold
		jointDef.Base.CollideConnected = true
		jointDef.Base.UserData = index
		s.jointIDs[index] = s.worldID.CreateWeldJoint(&jointDef).Joint
	}

	position.X += 5.0
	index++

	// wheel joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultWheelJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Hertz = 1.0
		jointDef.DampingRatio = 0.7
		jointDef.LowerTranslation = -1.0
		jointDef.UpperTranslation = 1.0
		jointDef.EnableLimit = true
		jointDef.EnableMotor = true
		jointDef.MaxMotorTorque = 10.0
		jointDef.MotorSpeed = 1.0
		jointDef.Base.ForceThreshold = forceThreshold
		jointDef.Base.TorqueThreshold = torqueThreshold
		jointDef.Base.CollideConnected = true
		jointDef.Base.UserData = index
		s.jointIDs[index] = s.worldID.CreateWheelJoint(&jointDef).Joint
	}

	return s
}

func (s *JointEvent) Step() {
	s.Sample.Step()

	// Process joint events
	for event := range s.worldID.JointEvents() {
		// Destroy the joint if it is still valid
		if event.Joint.IsValid() {
			index, ok := event.UserData.(int)
			if !ok {
				panic("JointEvent joint user data must be its integer slot index")
			}

			event.Joint.Destroy()

			s.jointIDs[index] = box2d.Joint{}
		}
	}
}

func CreatePersistentContact(context *SampleContext) Sampler {
	const (
		cameraHeight = 6.0
		firstGroundX = 10.0
		roomHeight   = 10.0
		initialSpeed = 2.0
		ballRadius   = 0.5
	)

	s := &PersistentContact{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		contactID: box2d.Contact{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 7.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		var points [22]box2d.Vec2

		x := float32(firstGroundX)
		for i := range 20 {
			points[i] = box2d.Vec2{X: x, Y: 0.0}
			x -= 1.0
		}

		points[20] = box2d.Vec2{X: -9.0, Y: roomHeight}
		points[21] = box2d.Vec2{X: firstGroundX, Y: roomHeight}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.IsLoop = true

		groundID.CreateChain(&chainDef)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -8.0, Y: 1.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: initialSpeed, Y: 0.0}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableContactEvents = true
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	s.contactID = box2d.Contact{}

	return s
}

func (s *PersistentContact) Step() {
	const contactPointSize = 6.0

	s.Sample.Step()

	for event := range s.worldID.ContactBeginEvents() {
		s.contactID = event.Contact
		break
	}

	for event := range s.worldID.ContactEndEvents() {
		if s.contactID == event.Contact {
			s.contactID = box2d.Contact{}
			break
		}
	}

	if s.contactID != (box2d.Contact{}) && s.contactID.IsValid() {
		data := s.contactID.Data()

		for i := range data.Manifold.PointCount {
			manifoldPoint := &data.Manifold.Points[i]
			bodyIDA := data.ShapeA.Body()
			p1 := bodyIDA.WorldCenter().Add(manifoldPoint.AnchorA)
			p2 := p1.Add(data.Manifold.Normal.Scale(manifoldPoint.TotalNormalImpulse))
			s.draw.DrawLine(p1, p2, box2d.ColorCrimson)
			s.draw.DrawPoint(p1, contactPointSize, box2d.ColorCrimson)
			s.draw.DrawStringf(
				s.camera,
				p1,
				box2d.ColorWhite,
				"%.2f",
				manifoldPoint.TotalNormalImpulse,
			)
		}
	} else {
		s.contactID = box2d.Contact{}
	}
}

func CreateSensorHits(context *SampleContext) Sampler {
	const (
		cameraHeight     = 5.0
		roomHalfWidth    = 10.0
		roomHeight       = 10.0
		sensorHeight     = 10.0
		dynamicSensorX   = 4.0
		capsuleTop       = 9.0
		capsuleRadius    = 0.1
		jointPivotHeight = 6.0
	)

	s := &SensorHits{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		staticSensorID:    box2d.Shape{},
		kinematicSensorID: box2d.Shape{},
		dynamicSensorID:   box2d.Shape{},
		kinematicBodyID:   box2d.Body{},
		dynamicBodyID:     box2d.Body{},
		jointID:           box2d.PrismaticJoint{Joint: box2d.Joint{}},
		bodyID:            box2d.Body{},
		shapeID:           box2d.Shape{},
		transformCount:    0,
		transforms:        [sensorHitsTransformCapacity]box2d.WorldTransform{},
		isBullet:          true,
		beginCount:        0,
		endCount:          0,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 7.5
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "ground"

		groundID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		groundSegment := box2d.Segment{
			Point1: box2d.Vec2{X: -roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: roomHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{
			Point1: box2d.Vec2{X: roomHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: roomHalfWidth, Y: roomHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)
	}

	// Static sensor
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "static sensor"
		bodyDef.Position = box2d.Pos{X: -4.0, Y: 1.0}

		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: 0.0, Y: 0.0},
			Point2: box2d.Vec2{X: 0.0, Y: sensorHeight},
		}
		s.staticSensorID = bodyID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Kinematic sensor
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "kinematic sensor"
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: sensorHitsSpeed, Y: 0.0}

		s.kinematicBodyID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: 0.0, Y: 0.0},
			Point2: box2d.Vec2{X: 0.0, Y: sensorHeight},
		}
		s.kinematicSensorID = s.kinematicBodyID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Dynamic sensor
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "dynamic sensor"
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: dynamicSensorX, Y: 1.0}

		s.dynamicBodyID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 1.0},
			Center2: box2d.Vec2{X: 0.0, Y: capsuleTop},
			Radius:  capsuleRadius,
		}
		s.dynamicSensorID = s.dynamicBodyID.CreateCapsuleShape(&shapeDef, &capsule)

		pivot := bodyDef.Position.Add(box2d.Vec2{X: 0.0, Y: jointPivotHeight})
		axis := box2d.Vec2{X: 1.0, Y: 0.0}
		jointDef := box2d.DefaultPrismaticJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.dynamicBodyID
		jointDef.Base.LocalFrameA.Q = box2d.MakeRotFromUnitVector(axis)
		jointDef.Base.LocalFrameA.P = groundID.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.Q = box2d.MakeRotFromUnitVector(axis)
		jointDef.Base.LocalFrameB.P = s.dynamicBodyID.LocalPoint(pivot)
		jointDef.EnableMotor = true
		jointDef.MaxMotorForce = 1000.0
		jointDef.MotorSpeed = 0.5

		s.jointID = s.worldID.CreatePrismaticJoint(&jointDef)
	}

	s.beginCount = 0
	s.endCount = 0
	s.bodyID = box2d.Body{}
	s.shapeID = box2d.Shape{}
	s.transformCount = 0
	s.isBullet = true

	s.Launch()

	return s
}

func (s *SensorHits) Launch() {
	const (
		launchHeight     = 6.0
		minLaunchSpeed   = 200.0
		maxLaunchSpeed   = 300.0
		projectileRadius = 0.25
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	s.transformCount = 0
	s.beginCount = 0
	s.endCount = 0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -26.7, Y: launchHeight}
	speed := scenes.RandomFloatRange(minLaunchSpeed, maxLaunchSpeed)
	bodyDef.LinearVelocity = box2d.Vec2{X: speed, Y: 0.0}
	bodyDef.IsBullet = s.isBullet
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true
	shapeDef.Material.Friction = 0.8
	shapeDef.Material.RollingResistance = 0.01
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: projectileRadius}
	s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
}

func (s *SensorHits) DrawControls() bool {
	imgui.Checkbox("Bullet", &s.isBullet)

	if imgui.Button("Launch") || s.context.window.GetKey(glfw.KeyB) == glfw.Press {
		s.Launch()
	}

	return true
}

func (s *SensorHits) CollectTransforms(sensorShapeID box2d.Shape) {
	const capacity = 5

	count := 0
	for range sensorShapeID.SensorData() {
		if count == capacity {
			break
		}

		count++
	}

	for i := 0; i < count && s.transformCount < sensorHitsTransformCapacity; i++ {
		sensorBodyID := sensorShapeID.Body()
		s.transforms[s.transformCount] = sensorBodyID.Transform()
		s.transformCount += 1
	}
}

func (s *SensorHits) Step() {
	p := s.kinematicBodyID.Position()
	if p.X > 1.0 {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: -sensorHitsSpeed, Y: 0.0})
	} else if p.X < -1.0 {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: sensorHitsSpeed, Y: 0.0})
	}

	x := s.jointID.Translation()
	if x > 1.0 {
		s.jointID.SetMotorSpeed(-sensorHitsSpeed)
	} else if x < -1.0 {
		s.jointID.SetMotorSpeed(sensorHitsSpeed)
	}

	s.Sample.Step()

	for i := range s.transformCount {
		s.draw.DrawTransform(s.transforms[i], 1.0)
	}

	for event := range s.worldID.SensorBeginEvents() {
		s.beginCount += 1
		if event.SensorShape.IsValid() {
			s.CollectTransforms(event.SensorShape)
		}
	}

	for range s.worldID.SensorEndEvents() {
		s.endCount += 1
	}

	s.DrawScreenTextLinef("begin touch count = %d", s.beginCount)
	s.DrawScreenTextLinef("end touch count = %d", s.endCount)
}

func CreateProjectileEvent(context *SampleContext) Sampler {
	const (
		cameraHeight      = 9.0
		wallX             = 10.0
		wallHeight        = 20.0
		groundHalfWidth   = 30.0
		blockHalfExtent   = 0.45
		blockCornerRadius = 0.05
		blockOffset       = 0.01
		blockX            = 8.0
		firstBlockHeight  = 0.5
	)

	s := &ProjectileEvent{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		projectileID:      box2d.Body{},
		projectileShapeID: box2d.Shape{},
		point1:            box2d.Pos{X: 0, Y: 0},
		point2:            box2d.Pos{X: 0, Y: 0},
		dragging:          false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: -7.0, Y: cameraHeight}
		s.context.camera.zoom = 14.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableSensorEvents = true

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: wallX, Y: 0.0},
			Point2: box2d.Vec2{X: wallX, Y: wallHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.projectileID = box2d.Body{}
	s.projectileShapeID = box2d.Shape{}
	s.dragging = false
	s.point1 = box2d.Pos{X: 0, Y: 0}
	s.point2 = box2d.Pos{X: 0, Y: 0}

	box := box2d.MakeRoundedBox(blockHalfExtent, blockHalfExtent, blockCornerRadius)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true

	offset := float32(blockOffset)

	x := float32(blockX)
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	for i := range 8 {
		shift := offset
		if i%2 == 0 {
			shift = -offset
		}

		bodyDef.Position = box2d.Pos{
			X: posScalar(x + shift),
			Y: posScalar(firstBlockHeight + 1.0*float32(i)),
		}

		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *ProjectileEvent) FireProjectile() {
	const (
		launchVelocityScale = 4.0
		projectileRadius    = 0.25
	)

	if s.projectileID != (box2d.Body{}) {
		s.projectileID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = s.point1
	bodyDef.LinearVelocity = s.point2.Sub(s.point1).Scale(launchVelocityScale)
	bodyDef.IsBullet = true

	s.projectileID = s.worldID.CreateBody(&bodyDef)

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: projectileRadius}
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableContactEvents = true
	s.projectileShapeID = s.projectileID.CreateCircleShape(&shapeDef, &circle)
}

func (s *ProjectileEvent) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == glfw.ModControl {
			s.dragging = true
			s.point1 = p
		}
	}
}

func (s *ProjectileEvent) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		if s.dragging {
			s.dragging = false
			s.FireProjectile()
		}
	}
}

func (s *ProjectileEvent) MouseMove(p box2d.Pos) {
	if s.dragging {
		s.point2 = p
	}
}

func (s *ProjectileEvent) Step() {
	const dragPointSize = 5.0

	s.DrawScreenTextLinef("Use Ctrl + Left Mouse to drag and shoot a projectile")

	s.Sample.Step()

	if s.dragging {
		s.draw.DrawLine(s.point1, s.point2, box2d.ColorWhite)
		s.draw.DrawPoint(s.point1, dragPointSize, box2d.ColorGreen)
		s.draw.DrawPoint(s.point2, dragPointSize, box2d.ColorRed)
	}

	for event := range s.worldID.ContactBeginEvents() {
		if event.ShapeA == s.projectileShapeID || event.ShapeB == s.projectileShapeID {
			if event.Contact.IsValid() {
				data := event.Contact.Data()

				if data.Manifold.PointCount > 0 {
					explosionDef := box2d.DefaultExplosionDef()
					explosionDef.Position = data.ShapeA.Body().
						WorldCenter().
						Add(data.Manifold.Points[0].AnchorA)
					explosionDef.Radius = 1.0
					explosionDef.ImpulsePerLength = 20.0
					s.worldID.Explode(&explosionDef)

					s.projectileID.Destroy()
					s.projectileID = box2d.Body{}
				}
			}

			break
		}
	}
}

func CreateCircleImpulse(context *SampleContext) Sampler {
	const (
		cameraHeight       = 2.7
		groundHalfWidth    = 10.0
		initialGravity     = 10.0
		initialRestitution = 0.25
	)

	s := &CircleImpulse{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		mass:           1.0,
		events:         nil,
		bodyID:         box2d.Body{},
		gravity:        initialGravity,
		restitution:    initialRestitution,
		useGravity:     false,
		useRestitution: false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 3.4
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.gravity = 10.0
	s.restitution = 0.25
	s.useGravity = false
	s.useRestitution = false
	s.mass = 1.0
	s.bodyID = box2d.Body{}

	s.Spawn()

	return s
}

func (s *CircleImpulse) Spawn() {
	const initialDownwardSpeed = 25.0

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
		s.bodyID = box2d.Body{}
	}

	s.events = s.events[:0]

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	bodyDef.GravityScale = 0.0
	if s.useGravity {
		bodyDef.GravityScale = 1.0
	}

	bodyDef.LinearVelocity.Y = -initialDownwardSpeed
	bodyDef.Position.Y = 5.5

	var circle box2d.Circle

	circle.Radius = 0.25

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableHitEvents = true
	shapeDef.Material.Friction = 0.0

	shapeDef.Material.Restitution = 0.0
	if s.useRestitution {
		shapeDef.Material.Restitution = s.restitution
	}

	s.bodyID = s.worldID.CreateBody(&bodyDef)

	s.bodyID.CreateCircleShape(&shapeDef, &circle)

	// Override mass
	massData := s.bodyID.MassData()
	ratio := s.mass / massData.Mass
	massData.Mass = s.mass
	massData.RotationalInertia *= ratio
	s.bodyID.SetMassData(massData)
}

func (s *CircleImpulse) DrawControls() bool {
	if imgui.Checkbox("gravity", &s.useGravity) {
		s.Spawn()
	}

	if imgui.Checkbox("restitution", &s.useRestitution) {
		s.Spawn()
	}

	return true
}

func (s *CircleImpulse) Step() {
	const hitPointSize = 10.0

	s.Sample.Step()

	for event := range s.worldID.ContactHitEvents() {
		s.draw.DrawPoint(event.Point, hitPointSize, box2d.ColorWhite)

		data := event.Contact.Data()

		var e circleImpulseEvent

		e.speed = event.ApproachSpeed
		if data.Manifold.PointCount > 0 {
			e.impulse = data.Manifold.Points[0].NormalImpulse
			e.totalImpulse = data.Manifold.Points[0].TotalNormalImpulse
		}

		s.events = append(s.events, e)
	}

	gravity := float32(0.0)
	if s.useGravity {
		gravity = 10.0
	}

	restitution := float32(0.0)
	if s.useRestitution {
		restitution = s.restitution
	}

	s.DrawScreenTextLinef(
		"mass = %.6g, gravity = %.6g, restitution = %.6g",
		s.mass,
		gravity,
		restitution,
	)

	for _, e := range s.events {
		s.DrawScreenTextLinef(
			"hit speed = %.6g, hit momentum = %.6g, final impulse = %.6g, total impulse = %.6g",
			e.speed,
			s.mass*e.speed,
			e.impulse,
			e.totalImpulse,
		)
	}
}

func registerEventSamples() {
	RegisterSample("Events", "Sensor Funnel", CreateSensorFunnel)
	RegisterSample("Events", "Sensor Bookend", CreateSensorBookend)
	RegisterSample("Events", "Foot Sensor", CreateFootSensor)
	RegisterSample("Events", "Contact", CreateContactEvent)
	RegisterSample("Events", "Body Move", CreateBodyMove)
	RegisterSample("Events", "Sensor Types", CreateSensorTypes)
	RegisterSample("Events", "Joint", CreateJointEvent)
	RegisterSample("Events", "Persistent Contact", CreatePersistentContact)
	RegisterSample("Events", "Sensor Hits", CreateSensorHits)
	RegisterSample("Events", "Projectile Event", CreateProjectileEvent)
	RegisterSample("Events", "Circle Impulse", CreateCircleImpulse)
}

func drawBodyEnableCheckbox(label string, bodyID box2d.Body) {
	enabledBody := bodyID.Enabled()
	if imgui.Checkbox(label, &enabledBody) {
		if enabledBody {
			bodyID.Enable()
		} else {
			bodyID.Disable()
		}
	}
}

func drawSensorEventsCheckbox(label string, shapeID box2d.Shape) {
	enabledEvents := shapeID.SensorEventsEnabled()
	if imgui.Checkbox(label, &enabledEvents) {
		shapeID.SetSensorEventsEnabled(enabledEvents)
	}
}

func (s *SensorBookend) drawSensorControls() {
	switch {
	case s.sensorBodyID1 == (box2d.Body{}):
		if imgui.Button("create sensor1") {
			s.CreateSensor1()
		}
	case imgui.Button("destroy sensor1"):
		s.sensorBodyID1.Destroy()
		s.sensorBodyID1 = box2d.Body{}
		// Retain sensorShapeID1 for end events.
	default:
		drawSensorEventsCheckbox("sensor 1 events", s.sensorShapeID1)
		drawBodyEnableCheckbox("enable sensor1 body", s.sensorBodyID1)
	}

	imgui.Separator()

	switch {
	case s.sensorBodyID2 == (box2d.Body{}):
		if imgui.Button("create sensor2") {
			s.CreateSensor2()
		}
	case imgui.Button("destroy sensor2"):
		s.sensorBodyID2.Destroy()
		s.sensorBodyID2 = box2d.Body{}
		// Retain sensorShapeID2 for end events.
	default:
		drawSensorEventsCheckbox("sensor2 events", s.sensorShapeID2)
		drawBodyEnableCheckbox("enable sensor2 body", s.sensorBodyID2)
	}
}
