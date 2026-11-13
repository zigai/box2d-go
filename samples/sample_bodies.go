// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
)

const wakeTouchingCount = 10

type BodyType struct {
	Sample

	attachmentID       box2d.Body
	secondAttachmentID box2d.Body
	platformID         box2d.Body
	secondPayloadID    box2d.Body
	touchingBodyID     box2d.Body
	floatingBodyID     box2d.Body
	bodyType           box2d.BodyType
	speed              float32
	isEnabled          bool
}

type Weeble struct {
	Sample

	weebleID           box2d.Body
	explosionPosition  box2d.Pos
	explosionRadius    float32
	explosionMagnitude float32
}

type Sleep struct {
	Sample

	pendulumID     box2d.Body
	staticBodyID   box2d.Body
	groundShapeID  box2d.Shape
	sensorIDs      [2]box2d.Shape
	sensorTouching [2]bool
}

type BadBody struct {
	Sample

	badBodyID box2d.Body
}

type Pivot struct {
	Sample

	bodyID box2d.Body
	lever  float32
}

type Kinematic struct {
	Sample

	bodyID    box2d.Body
	amplitude float32
	time      float32
}

type MixedLocks struct {
	Sample
}

type SetVelocity struct {
	Sample

	bodyID box2d.Body
}

type WakeTouching struct {
	Sample

	groundID box2d.Body
}

func CreateBodyType(context *SampleContext) Sampler {
	const (
		cameraCenterX        = 0.8
		cameraCenterY        = 6.4
		baseCameraZoom       = 25.0
		cameraZoomScale      = 0.4
		groundHalfLength     = 20.0
		attachmentHeight     = 3.0
		attachmentHalfWidth  = 0.5
		attachmentHalfHeight = 2.0
		secondAttachmentX    = 3.0
		platformSpeed        = 3.0
	)

	s := &BodyType{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		attachmentID:       box2d.Body{},
		secondAttachmentID: box2d.Body{},
		platformID:         box2d.Body{},
		secondPayloadID:    box2d.Body{},
		touchingBodyID:     box2d.Body{},
		floatingBodyID:     box2d.Body{},
		bodyType:           box2d.StaticBody,
		speed:              platformSpeed,
		isEnabled:          true,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "ground"
		groundID = s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -2.0, Y: attachmentHeight}
		bodyDef.Name = "attach1"
		s.attachmentID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(attachmentHalfWidth, attachmentHalfHeight)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		s.attachmentID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: secondAttachmentX, Y: attachmentHeight}
		bodyDef.Name = "attach2"
		s.secondAttachmentID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(attachmentHalfWidth, attachmentHalfHeight)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		s.secondAttachmentID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		const (
			platformHeight      = 5.0
			platformHalfWidth   = 0.5
			platformHalfLength  = 4.0
			quarterTurnPiFactor = 0.5
			translationLimit    = 10.0
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: -platformHalfLength, Y: platformHeight}
		bodyDef.Name = "platform"
		s.platformID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(
			platformHalfWidth,
			platformHalfLength,
			box2d.Vec2{X: platformHalfLength, Y: 0.0},
			box2d.MakeRot(quarterTurnPiFactor*box2d.Pi),
		)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0
		s.platformID.CreatePolygonShape(&shapeDef, &box)

		revoluteDef := box2d.DefaultRevoluteJointDef()
		pivot := box2d.Pos{X: -2.0, Y: platformHeight}
		revoluteDef.Base.BodyA = s.attachmentID
		revoluteDef.Base.BodyB = s.platformID
		revoluteDef.Base.LocalFrameA.P = s.attachmentID.LocalPoint(pivot)
		revoluteDef.Base.LocalFrameB.P = s.platformID.LocalPoint(pivot)
		revoluteDef.MaxMotorTorque = 50.0
		revoluteDef.EnableMotor = true
		s.worldID.CreateRevoluteJoint(&revoluteDef)

		pivot = box2d.Pos{X: secondAttachmentX, Y: platformHeight}
		revoluteDef.Base.BodyA = s.secondAttachmentID
		revoluteDef.Base.BodyB = s.platformID
		revoluteDef.Base.LocalFrameA.P = s.secondAttachmentID.LocalPoint(pivot)
		revoluteDef.Base.LocalFrameB.P = s.platformID.LocalPoint(pivot)
		revoluteDef.MaxMotorTorque = 50.0
		revoluteDef.EnableMotor = true
		s.worldID.CreateRevoluteJoint(&revoluteDef)

		prismaticDef := box2d.DefaultPrismaticJointDef()
		anchor := box2d.Pos{X: 0.0, Y: platformHeight}
		prismaticDef.Base.BodyA = groundID
		prismaticDef.Base.BodyB = s.platformID
		prismaticDef.Base.LocalFrameA.P = groundID.LocalPoint(anchor)
		prismaticDef.Base.LocalFrameB.P = s.platformID.LocalPoint(anchor)
		prismaticDef.MaxMotorForce = 1000.0
		prismaticDef.MotorSpeed = 0.0
		prismaticDef.EnableMotor = true
		prismaticDef.LowerTranslation = -translationLimit
		prismaticDef.UpperTranslation = translationLimit
		prismaticDef.EnableLimit = true

		s.worldID.CreatePrismaticJoint(&prismaticDef)
	}

	{
		const (
			crateX          = -3.0
			crateHeight     = 8.0
			crateHalfExtent = 0.75
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: crateX, Y: crateHeight}
		bodyDef.Name = "crate1"
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(crateHalfExtent, crateHalfExtent)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		const (
			crateX          = 2.0
			crateHeight     = 8.0
			crateHalfExtent = 0.75
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: crateX, Y: crateHeight}
		bodyDef.Name = "crate2"
		s.secondPayloadID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(crateHalfExtent, crateHalfExtent)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		s.secondPayloadID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		const (
			debrisX      = 8.0
			debrisHeight = 0.2
			debrisRadius = 0.25
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: debrisX, Y: debrisHeight}
		bodyDef.Name = "debris"
		s.touchingBodyID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 0.0},
			Center2: box2d.Vec2{X: 1.0, Y: 0.0},
			Radius:  debrisRadius,
		}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		s.touchingBodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		const (
			debrisX      = 8.5
			debrisHeight = 0.2
			debrisRadius = 0.5
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.StaticBody
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: debrisX, Y: debrisHeight}
		bodyDef.Name = "debris"
		bodyID := s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 0.0},
			Center2: box2d.Vec2{X: 1.0, Y: 0.0},
			Radius:  debrisRadius,
		}

		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		const (
			floaterHeight = 12.0
			circleOffsetY = 0.5
			circleRadius  = 0.25
		)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: -8.0, Y: floaterHeight}
		bodyDef.GravityScale = 0.0
		bodyDef.Name = "floater"
		s.floatingBodyID = s.worldID.CreateBody(&bodyDef)

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: circleOffsetY}, Radius: circleRadius}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		s.floatingBodyID.CreateCircleShape(&shapeDef, &circle)
	}

	return s
}

func (s *BodyType) DrawControls() bool {
	if imgui.RadioButtonBool("Static", s.bodyType == box2d.StaticBody) {
		s.bodyType = box2d.StaticBody
		s.platformID.SetType(box2d.StaticBody)
		s.secondAttachmentID.SetType(box2d.StaticBody)
		s.secondPayloadID.SetType(box2d.StaticBody)
		s.touchingBodyID.SetType(box2d.StaticBody)
		s.floatingBodyID.SetType(box2d.StaticBody)
	}

	if imgui.RadioButtonBool("Kinematic", s.bodyType == box2d.KinematicBody) {
		s.bodyType = box2d.KinematicBody
		s.platformID.SetType(box2d.KinematicBody)
		s.platformID.SetLinearVelocity(box2d.Vec2{X: -s.speed, Y: 0.0})
		s.platformID.SetAngularVelocity(0.0)

		s.secondAttachmentID.SetType(box2d.KinematicBody)
		s.secondAttachmentID.SetLinearVelocity(box2d.Vec2Zero)
		s.secondAttachmentID.SetAngularVelocity(0.0)

		s.secondPayloadID.SetType(box2d.KinematicBody)
		s.touchingBodyID.SetType(box2d.KinematicBody)
		s.floatingBodyID.SetType(box2d.KinematicBody)
	}

	if imgui.RadioButtonBool("Dynamic", s.bodyType == box2d.DynamicBody) {
		s.bodyType = box2d.DynamicBody
		s.platformID.SetType(box2d.DynamicBody)
		s.secondAttachmentID.SetType(box2d.DynamicBody)
		s.secondPayloadID.SetType(box2d.DynamicBody)
		s.touchingBodyID.SetType(box2d.DynamicBody)
		s.floatingBodyID.SetType(box2d.DynamicBody)
	}

	if imgui.Checkbox("Enable", &s.isEnabled) {
		if s.isEnabled {
			s.attachmentID.Enable()
			s.secondPayloadID.Enable()
			s.floatingBodyID.Enable()
		} else {
			s.attachmentID.Disable()
			s.secondPayloadID.Disable()
			s.floatingBodyID.Disable()
		}
	}

	return true
}

func (s *BodyType) Step() {
	if s.bodyType == box2d.KinematicBody {
		p := s.platformID.Position()
		v := s.platformID.LinearVelocity()

		if (p.X < -14.0 && v.X < 0.0) || (p.X > 6.0 && v.X > 0.0) {
			v.X = -v.X
			s.platformID.SetLinearVelocity(v)
		}
	}

	s.Sample.Step()
}

func CreateWeeble(context *SampleContext) Sampler {
	const (
		cameraCenterX       = 2.3
		cameraCenterY       = 10.0
		baseCameraZoom      = 25.0
		cameraZoomScale     = 0.5
		contactFriction     = 0.1
		groundHalfLength    = 20.0
		bodyStartHeight     = 3.0
		initialAnglePiRatio = 0.25
		massCenterOffset    = 1.5
		initialBlastRadius  = 2.0
		initialBlastImpulse = 8.0
	)

	s := &Weeble{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		weebleID:           box2d.Body{},
		explosionPosition:  box2d.Pos{X: 0, Y: 0},
		explosionRadius:    initialBlastRadius,
		explosionMagnitude: initialBlastImpulse,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	s.worldID.SetFrictionCallback(
		func(float32, uint64, float32, uint64) float32 { return contactFriction },
	)
	s.worldID.SetRestitutionCallback(func(float32, uint64, float32, uint64) float32 { return 1.0 })

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: bodyStartHeight}
		bodyDef.Rotation = box2d.MakeRot(initialAnglePiRatio * box2d.Pi)
		s.weebleID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -1.0},
			Center2: box2d.Vec2{X: 0.0, Y: 1.0},
			Radius:  1.0,
		}
		shapeDef := box2d.DefaultShapeDef()
		s.weebleID.CreateCapsuleShape(&shapeDef, &capsule)

		mass := s.weebleID.Mass()
		inertiaTensor := s.weebleID.RotationalInertia()

		offset := float32(massCenterOffset)

		inertiaTensor += mass * offset * offset

		massData := box2d.MassData{
			Mass:              mass,
			Center:            box2d.Vec2{X: 0.0, Y: -offset},
			RotationalInertia: inertiaTensor,
		}
		s.weebleID.SetMassData(massData)
	}

	return s
}

func (s *Weeble) DrawControls() bool {
	const (
		teleportHeight       = 5.0
		teleportAnglePiRatio = 0.95
		controlWidthEm       = 6.0
		magnitudeLimit       = 100.0
	)

	if imgui.Button("Teleport") {
		s.weebleID.SetTransform(
			box2d.Pos{X: 0.0, Y: teleportHeight},
			box2d.MakeRot(teleportAnglePiRatio*box2d.Pi),
		)
	}

	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()
		def.Position = s.explosionPosition
		def.Radius = s.explosionRadius
		def.Falloff = 0.1
		def.ImpulsePerLength = s.explosionMagnitude
		s.worldID.Explode(&def)
	}

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())
	imgui.SliderFloatV(
		"Magnitude",
		&s.explosionMagnitude,
		-magnitudeLimit,
		magnitudeLimit,
		"%.1f",
		0,
	)
	imgui.PopItemWidth()

	return true
}

func (s *Weeble) Step() {
	const (
		velocityProbeHeight = 2.0
		velocityLineOffset  = 0.05
	)

	s.Sample.Step()

	s.context.draw.DrawCircle(s.explosionPosition, s.explosionRadius, box2d.ColorAzure)

	localPoint := box2d.Vec2{X: 0.0, Y: velocityProbeHeight}
	worldPoint := s.weebleID.WorldPoint(localPoint)

	v1 := s.weebleID.LocalPointVelocity(localPoint)
	v2 := s.weebleID.WorldPointVelocity(worldPoint)

	offset := box2d.Vec2{X: velocityLineOffset, Y: 0.0}

	s.context.draw.DrawLine(worldPoint, worldPoint.Add(v1), box2d.ColorRed)
	s.context.draw.DrawLine(
		worldPoint.Add(offset),
		worldPoint.Add(v2).Add(offset),
		box2d.ColorGreen,
	)
}

func CreateSleep(context *SampleContext) Sampler {
	const (
		cameraCenterX       = 3.0
		cameraCenterY       = 50.0
		baseCameraZoom      = 25.0
		cameraZoomScale     = 2.2
		groundHalfLength    = 40.0
		sensorBaseHeight    = 3.0
		sensorHeightStep    = 2.0
		sensorCapsuleRadius = 0.75
		sleepCircleHeight   = 3.0
		awakeBoxX           = 5.0
		awakeBoxHeight      = 3.0
		awakeBoxAnglePi     = 0.25
		sleepBoxX           = 5.0
		pendulumHeight      = 100.0
		pendulumLength      = 90.0
		pendulumRadius      = 0.25
	)

	s := &Sleep{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
		pendulumID:     box2d.Body{},
		staticBodyID:   box2d.Body{},
		groundShapeID:  box2d.Shape{},
		sensorIDs:      [2]box2d.Shape{},
		sensorTouching: [2]bool{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableSensorEvents = true
		s.groundShapeID = groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	for i := range 2 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{
			X: -4.0,
			Y: posScalar(sensorBaseHeight + sensorHeightStep*float32(i)),
		}
		bodyDef.IsAwake = false
		bodyDef.EnableSleep = true
		bodyID := s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 1.0},
			Center2: box2d.Vec2{X: 1.0, Y: 1.0},
			Radius:  sensorCapsuleRadius,
		}
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)

		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true
		capsule.Radius = 1.0
		s.sensorIDs[i] = bodyID.CreateCapsuleShape(&shapeDef, &capsule)
		s.sensorTouching[i] = false
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: sleepCircleHeight}
		bodyDef.IsAwake = false
		bodyDef.EnableSleep = false
		bodyID := s.worldID.CreateBody(&bodyDef)

		circle := box2d.Circle{Center: box2d.Vec2{X: 1.0, Y: 1.0}, Radius: 1.0}
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: awakeBoxX, Y: awakeBoxHeight}
		bodyDef.IsAwake = true
		bodyDef.EnableSleep = false
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(
			1.0,
			1.0,
			box2d.Vec2{X: 0.0, Y: 1.0},
			box2d.MakeRot(awakeBoxAnglePi*box2d.Pi),
		)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: sleepBoxX, Y: 1.0}
		bodyDef.IsAwake = false
		bodyDef.EnableSleep = true
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeSquare(1.0)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: pendulumHeight}
		bodyDef.AngularDamping = 0.5
		bodyDef.SleepThreshold = 0.05
		s.pendulumID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 0.0},
			Center2: box2d.Vec2{X: pendulumLength, Y: 0.0},
			Radius:  pendulumRadius,
		}
		shapeDef := box2d.DefaultShapeDef()
		s.pendulumID.CreateCapsuleShape(&shapeDef, &capsule)

		pivot := bodyDef.Position
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.pendulumID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		s.worldID.CreateRevoluteJoint(&jointDef)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -10.0, Y: 1.0}
		bodyDef.IsAwake = false
		bodyDef.EnableSleep = true
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeSquare(1.0)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *Sleep) ToggleInvoker() {
	const (
		invokerHeight       = 3.0
		invokerHalfWidth    = 2.0
		invokerHalfHeight   = 0.1
		invokerAnglePiRatio = 0.25
	)

	if s.staticBodyID == (box2d.Body{}) {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -10.5, Y: invokerHeight}
		s.staticBodyID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(
			invokerHalfWidth,
			invokerHalfHeight,
			box2d.Vec2{X: 0.0, Y: 0.0},
			box2d.MakeRot(invokerAnglePiRatio*box2d.Pi),
		)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.InvokeContactCreation = true
		s.staticBodyID.CreatePolygonShape(&shapeDef, &box)
	} else {
		s.staticBodyID.Destroy()
		s.staticBodyID = box2d.Body{}
	}
}

func (s *Sleep) DrawControls() bool {
	const (
		controlWidthEm    = 6.0
		maxAngularDamping = 2.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	imgui.Text("Pendulum Tuning")

	sleepVelocity := s.pendulumID.SleepThreshold()
	if imgui.SliderFloatV("sleep velocity", &sleepVelocity, 0.0, 1.0, "%.2f", 0) {
		s.pendulumID.SetSleepThreshold(sleepVelocity)
		s.pendulumID.SetAwake(true)
	}

	angularDamping := s.pendulumID.AngularDamping()
	if imgui.SliderFloatV("angular damping", &angularDamping, 0.0, maxAngularDamping, "%.2f", 0) {
		s.pendulumID.SetAngularDamping(angularDamping)
	}

	imgui.PopItemWidth()

	imgui.Separator()

	if s.staticBodyID == (box2d.Body{}) {
		if imgui.Button("Create") {
			s.ToggleInvoker()
		}
	} else {
		if imgui.Button("Destroy") {
			s.ToggleInvoker()
		}
	}

	return true
}

func (s *Sleep) Step() {
	s.Sample.Step()

	for event := range s.worldID.SensorBeginEvents() {
		if event.VisitorShape == s.groundShapeID {
			switch event.SensorShape {
			case s.sensorIDs[0]:
				s.sensorTouching[0] = true
			case s.sensorIDs[1]:
				s.sensorTouching[1] = true
			}
		}
	}

	for event := range s.worldID.SensorEndEvents() {
		if event.VisitorShape == s.groundShapeID {
			switch event.SensorShape {
			case s.sensorIDs[0]:
				s.sensorTouching[0] = false
			case s.sensorIDs[1]:
				s.sensorTouching[1] = false
			}
		}
	}

	for i := range 2 {
		s.DrawScreenTextLinef("sensor touch %d = %t", i, s.sensorTouching[i])
	}
}

func CreateBadBody(context *SampleContext) Sampler {
	const (
		cameraCenterX       = 2.3
		cameraCenterY       = 10.0
		baseCameraZoom      = 25.0
		cameraZoomScale     = 0.5
		groundHalfLength    = 20.0
		bodyHeight          = 3.0
		bodyAnglePiFraction = 0.25
		comparisonBodyX     = 2.0
	)

	s := &BadBody{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		badBodyID:       box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
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
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: bodyHeight}
		bodyDef.AngularVelocity = 0.5
		bodyDef.Rotation = box2d.MakeRot(bodyAnglePiFraction * box2d.Pi)

		s.badBodyID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -1.0},
			Center2: box2d.Vec2{X: 0.0, Y: 1.0},
			Radius:  1.0,
		}
		shapeDef := box2d.DefaultShapeDef()

		shapeDef.Density = 0.0
		s.badBodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: comparisonBodyX, Y: bodyHeight}
		bodyDef.Rotation = box2d.MakeRot(bodyAnglePiFraction * box2d.Pi)

		bodyID := s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: -1.0},
			Center2: box2d.Vec2{X: 0.0, Y: 1.0},
			Radius:  1.0,
		}
		shapeDef := box2d.DefaultShapeDef()

		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	return s
}

func (s *BadBody) Step() {
	const appliedUpwardForce = 10.0

	s.Sample.Step()

	s.DrawScreenTextLinef(
		"A bad body is a dynamic body with no mass and behaves like a kinematic body.",
	)
	s.DrawScreenTextLinef(
		"Bad bodies are considered invalid and a user bug. Behavior is not guaranteed.",
	)

	s.badBodyID.ApplyForceToCenter(box2d.Vec2{X: 0.0, Y: appliedUpwardForce}, true)
}

func CreatePivot(context *SampleContext) Sampler {
	const (
		cameraCenterX    = 0.8
		cameraCenterY    = 6.4
		baseCameraZoom   = 25.0
		cameraZoomScale  = 0.4
		groundHalfLength = 20.0
		initialVelocityX = 5.0
		bodyHeight       = 3.0
		leverHalfWidth   = 0.1
		leverLength      = 3.0
	)

	s := &Pivot{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		bodyID:          box2d.Body{},
		lever:           leverLength,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
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
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		v := box2d.Vec2{X: initialVelocityX, Y: 0.0}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: bodyHeight}
		bodyDef.GravityScale = 1.0
		bodyDef.LinearVelocity = v

		s.bodyID = s.worldID.CreateBody(&bodyDef)

		r := box2d.Vec2{X: 0.0, Y: -s.lever}

		omega := v.Cross(r) / r.Dot(r)
		s.bodyID.SetAngularVelocity(omega)

		box := box2d.MakeBox(leverHalfWidth, s.lever)

		shapeDef := box2d.DefaultShapeDef()
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *Pivot) Step() {
	s.Sample.Step()

	v := s.bodyID.LinearVelocity()
	omega := s.bodyID.AngularVelocity()
	r := s.bodyID.WorldVector(box2d.Vec2{X: 0.0, Y: -s.lever})

	vp := v.Add(r.ScalarCross(omega))
	s.DrawScreenTextLinef("pivot velocity = (%.6g, %.6g)", vp.X, vp.Y)
}

func CreateKinematic(context *SampleContext) Sampler {
	const (
		horizontalAmplitudeScale = 2.0
		bodyHalfWidth            = 0.1
		bodyHalfHeight           = 1.0
		motionAmplitude          = 2.0
	)

	s := &Kinematic{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		bodyID:          box2d.Body{},
		amplitude:       motionAmplitude,
		time:            0,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 4.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position.X = posScalar(horizontalAmplitudeScale * s.amplitude)

		s.bodyID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(bodyHalfWidth, bodyHalfHeight)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *Kinematic) Step() {
	const (
		horizontalAmplitudeScale = 2.0
		verticalFrequencyScale   = 2.0
		angularFrequencyScale    = 2.0
		axisHalfLength           = 0.5
		markerSize               = 10.0
	)

	var timeStep float32
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}

	if s.context.pause && s.context.singleStep == 0 {
		timeStep = 0.0
	}

	if timeStep > 0.0 {
		point := box2d.Pos{
			X: posScalar(
				horizontalAmplitudeScale * s.amplitude * float32(math.Cos(float64(s.time))),
			),
			Y: posScalar(s.amplitude * float32(math.Sin(float64(verticalFrequencyScale*s.time)))),
		}
		rotation := box2d.MakeRot(angularFrequencyScale * s.time)

		axis := rotation.Rotate(box2d.Vec2{X: 0.0, Y: 1.0})
		s.context.draw.DrawLine(
			point.Add(axis.Scale(-axisHalfLength)),
			point.Add(axis.Scale(axisHalfLength)),
			box2d.ColorPlum,
		)
		s.context.draw.DrawPoint(point, markerSize, box2d.ColorPlum)

		wake := true
		s.bodyID.SetTargetTransform(box2d.WorldTransform{P: point, Q: rotation}, timeStep, wake)
	}

	s.Sample.Step()

	s.time += timeStep
}

func CreateMixedLocks(context *SampleContext) Sampler {
	const (
		cameraHeight         = 2.5
		groundHalfLength     = 40.0
		boxHalfExtent        = 0.5
		staticBodyX          = 2.0
		upperFreeBodyHeight  = 3.0
		horizontalLockX      = -2.0
		horizontalLockHeight = 2.0
		mixedLockHeight      = 2.5
	)

	s := &MixedLocks{
		context: context, camera: &context.camera, draw: context.draw,
		screenTextX: initialScreenTextX, mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 3.5
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

	box := box2d.MakeSquare(boxHalfExtent)
	shapeDef := box2d.DefaultShapeDef()

	create := func(bodyType box2d.BodyType, position box2d.Pos, locks box2d.MotionLocks, name string) {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = bodyType
		bodyDef.Position = position
		bodyDef.MotionLocks = locks
		bodyDef.Name = name

		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	create(
		box2d.StaticBody, box2d.Pos{X: staticBodyX, Y: 1.0},
		box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: false}, "static",
	)
	create(
		box2d.DynamicBody, box2d.Pos{X: 1.0, Y: 1.0},
		box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: false}, "free",
	)
	create(
		box2d.DynamicBody, box2d.Pos{X: 1.0, Y: upperFreeBodyHeight},
		box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: false}, "free",
	)
	create(
		box2d.DynamicBody,
		box2d.Pos{X: -1.0, Y: 1.0},
		box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: true},
		"angular z",
	)
	create(
		box2d.DynamicBody,
		box2d.Pos{X: horizontalLockX, Y: horizontalLockHeight},
		box2d.MotionLocks{LinearX: true, LinearY: false, AngularZ: false},
		"linear x",
	)
	create(
		box2d.DynamicBody,
		box2d.Pos{X: -1.0, Y: mixedLockHeight},
		box2d.MotionLocks{LinearX: false, LinearY: true, AngularZ: true},
		"lin y ang z",
	)
	create(
		box2d.DynamicBody,
		box2d.Pos{X: 0.0, Y: 1.0},
		box2d.MotionLocks{LinearX: true, LinearY: true, AngularZ: true},
		"full",
	)

	return s
}

func CreateSetVelocity(context *SampleContext) Sampler {
	const (
		cameraHeight     = 2.5
		groundHalfWidth  = 20.0
		groundHalfHeight = 0.25
		boxHalfExtent    = 0.5
	)

	s := &SetVelocity{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		bodyID:          box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 3.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -groundHalfHeight}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(groundHalfWidth, groundHalfHeight)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeSquare(boxHalfExtent)
		bodyDef.Position = box2d.Pos{X: 0.0, Y: boxHalfExtent}
		s.bodyID = s.worldID.CreateBody(&bodyDef)
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *SetVelocity) Step() {
	s.Sample.Step()

	s.bodyID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: -20.0})

	position := s.bodyID.Position()
	s.DrawScreenTextLinef("(x, y) = (%.2g, %.2g)", position.X, position.Y)
}

func CreateWakeTouching(context *SampleContext) Sampler {
	const (
		cameraHeight     = 4.0
		groundHalfLength = 20.0
		boxHalfExtent    = 0.5
		dropHeight       = 4.0
	)

	s := &WakeTouching{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		groundID:        box2d.Body{},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 8.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		s.groundID = s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		s.groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	box := box2d.MakeBox(boxHalfExtent, boxHalfExtent)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	x := -1.0 * float32(wakeTouchingCount-1)

	for range wakeTouchingCount {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: dropHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		x += 2.0
	}

	return s
}

func (s *WakeTouching) DrawControls() bool {
	if imgui.Button("Wake Touching") {
		s.groundID.WakeTouching()
	}

	return true
}

func registerBodySamples() {
	RegisterSample("Bodies", "Body Type", CreateBodyType)
	RegisterSample("Bodies", "Weeble", CreateWeeble)
	RegisterSample("Bodies", "Sleep", CreateSleep)
	RegisterSample("Bodies", "Bad", CreateBadBody)
	RegisterSample("Bodies", "Pivot", CreatePivot)
	RegisterSample("Bodies", "Kinematic", CreateKinematic)
	RegisterSample("Bodies", "Mixed Locks", CreateMixedLocks)
	RegisterSample("Bodies", "Set Velocity", CreateSetVelocity)
	RegisterSample("Bodies", "Wake Touching", CreateWakeTouching)
}
