// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
)

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

func CreateBodyType(context *SampleContext) Sampler {
	s := &BodyType{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.8, Y: 6.4}
		s.context.camera.zoom = 25.0 * 0.4
	}

	s.bodyType = box2d.StaticBody
	s.isEnabled = true

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "ground"
		groundID = s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -2.0, Y: 3.0}
		bodyDef.Name = "attach1"
		s.attachmentID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(0.5, 2.0)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		s.attachmentID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: 3.0, Y: 3.0}
		bodyDef.Name = "attach2"
		s.secondAttachmentID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(0.5, 2.0)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		s.secondAttachmentID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: -4.0, Y: 5.0}
		bodyDef.Name = "platform"
		s.platformID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(
			0.5,
			4.0,
			box2d.Vec2{X: 4.0, Y: 0.0},
			box2d.MakeRot(0.5*box2d.Pi),
		)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0
		s.platformID.CreatePolygonShape(&shapeDef, &box)

		revoluteDef := box2d.DefaultRevoluteJointDef()
		pivot := box2d.Pos{X: -2.0, Y: 5.0}
		revoluteDef.Base.BodyA = s.attachmentID
		revoluteDef.Base.BodyB = s.platformID
		revoluteDef.Base.LocalFrameA.P = s.attachmentID.LocalPoint(pivot)
		revoluteDef.Base.LocalFrameB.P = s.platformID.LocalPoint(pivot)
		revoluteDef.MaxMotorTorque = 50.0
		revoluteDef.EnableMotor = true
		s.worldID.CreateRevoluteJoint(&revoluteDef)

		pivot = box2d.Pos{X: 3.0, Y: 5.0}
		revoluteDef.Base.BodyA = s.secondAttachmentID
		revoluteDef.Base.BodyB = s.platformID
		revoluteDef.Base.LocalFrameA.P = s.secondAttachmentID.LocalPoint(pivot)
		revoluteDef.Base.LocalFrameB.P = s.platformID.LocalPoint(pivot)
		revoluteDef.MaxMotorTorque = 50.0
		revoluteDef.EnableMotor = true
		s.worldID.CreateRevoluteJoint(&revoluteDef)

		prismaticDef := box2d.DefaultPrismaticJointDef()
		anchor := box2d.Pos{X: 0.0, Y: 5.0}
		prismaticDef.Base.BodyA = groundID
		prismaticDef.Base.BodyB = s.platformID
		prismaticDef.Base.LocalFrameA.P = groundID.LocalPoint(anchor)
		prismaticDef.Base.LocalFrameB.P = s.platformID.LocalPoint(anchor)
		prismaticDef.MaxMotorForce = 1000.0
		prismaticDef.MotorSpeed = 0.0
		prismaticDef.EnableMotor = true
		prismaticDef.LowerTranslation = -10.0
		prismaticDef.UpperTranslation = 10.0
		prismaticDef.EnableLimit = true

		s.worldID.CreatePrismaticJoint(&prismaticDef)

		s.speed = 3.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -3.0, Y: 8.0}
		bodyDef.Name = "crate1"
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(0.75, 0.75)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: 2.0, Y: 8.0}
		bodyDef.Name = "crate2"
		s.secondPayloadID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(0.75, 0.75)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		s.secondPayloadID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: 8.0, Y: 0.2}
		bodyDef.Name = "debris"
		s.touchingBodyID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 0.0},
			Center2: box2d.Vec2{X: 1.0, Y: 0.0},
			Radius:  0.25,
		}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 2.0

		s.touchingBodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.StaticBody
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: 8.5, Y: 0.2}
		bodyDef.Name = "debris"
		bodyID := s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 0.0},
			Center2: box2d.Vec2{X: 1.0, Y: 0.0},
			Radius:  0.5,
		}

		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = s.bodyType
		bodyDef.IsEnabled = s.isEnabled
		bodyDef.Position = box2d.Pos{X: -8.0, Y: 12.0}
		bodyDef.GravityScale = 0.0
		bodyDef.Name = "floater"
		s.floatingBodyID = s.worldID.CreateBody(&bodyDef)

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.5}, Radius: 0.25}

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

type Weeble struct {
	Sample
	weebleID           box2d.Body
	explosionPosition  box2d.Pos
	explosionRadius    float32
	explosionMagnitude float32
}

func CreateWeeble(context *SampleContext) Sampler {
	s := &Weeble{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 2.3, Y: 10.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	s.worldID.SetFrictionCallback(func(float32, uint64, float32, uint64) float32 { return 0.1 })
	s.worldID.SetRestitutionCallback(func(float32, uint64, float32, uint64) float32 { return 1.0 })

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 3.0}
		bodyDef.Rotation = box2d.MakeRot(0.25 * box2d.Pi)
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

		offset := float32(1.5)

		inertiaTensor += mass * offset * offset

		massData := box2d.MassData{
			Mass:              mass,
			Center:            box2d.Vec2{X: 0.0, Y: -offset},
			RotationalInertia: inertiaTensor,
		}
		s.weebleID.SetMassData(massData)
	}

	s.explosionPosition = box2d.Pos{X: 0.0, Y: 0.0}
	s.explosionRadius = 2.0
	s.explosionMagnitude = 8.0
	return s
}

func (s *Weeble) DrawControls() bool {
	if imgui.Button("Teleport") {
		s.weebleID.SetTransform(box2d.Pos{X: 0.0, Y: 5.0}, box2d.MakeRot(0.95*box2d.Pi))
	}

	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()
		def.Position = s.explosionPosition
		def.Radius = s.explosionRadius
		def.Falloff = 0.1
		def.ImpulsePerLength = s.explosionMagnitude
		s.worldID.Explode(&def)
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderFloatV("Magnitude", &s.explosionMagnitude, -100.0, 100.0, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *Weeble) Step() {
	s.Sample.Step()

	s.context.draw.DrawCircle(s.explosionPosition, s.explosionRadius, box2d.ColorAzure)

	localPoint := box2d.Vec2{X: 0.0, Y: 2.0}
	worldPoint := s.weebleID.WorldPoint(localPoint)

	v1 := s.weebleID.LocalPointVelocity(localPoint)
	v2 := s.weebleID.WorldPointVelocity(worldPoint)

	offset := box2d.Vec2{X: 0.05, Y: 0.0}
	s.context.draw.DrawLine(worldPoint, worldPoint.Add(v1), box2d.ColorRed)
	s.context.draw.DrawLine(
		worldPoint.Add(offset),
		worldPoint.Add(v2).Add(offset),
		box2d.ColorGreen,
	)
}

type Sleep struct {
	Sample
	pendulumID     box2d.Body
	staticBodyID   box2d.Body
	groundShapeID  box2d.Shape
	sensorIDs      [2]box2d.Shape
	sensorTouching [2]bool
}

func CreateSleep(context *SampleContext) Sampler {
	s := &Sleep{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 3.0, Y: 50.0}
		s.context.camera.zoom = 25.0 * 2.2
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -40.0, Y: 0.0},
			Point2: box2d.Vec2{X: 40.0, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableSensorEvents = true
		s.groundShapeID = groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	for i := range 2 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -4.0, Y: posScalar(3.0 + 2.0*float32(i))}
		bodyDef.IsAwake = false
		bodyDef.EnableSleep = true
		bodyID := s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 1.0},
			Center2: box2d.Vec2{X: 1.0, Y: 1.0},
			Radius:  0.75,
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
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 3.0}
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
		bodyDef.Position = box2d.Pos{X: 5.0, Y: 3.0}
		bodyDef.IsAwake = true
		bodyDef.EnableSleep = false
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(
			1.0,
			1.0,
			box2d.Vec2{X: 0.0, Y: 1.0},
			box2d.MakeRot(0.25*box2d.Pi),
		)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 5.0, Y: 1.0}
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
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 100.0}
		bodyDef.AngularDamping = 0.5
		bodyDef.SleepThreshold = 0.05
		s.pendulumID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: 0.0, Y: 0.0},
			Center2: box2d.Vec2{X: 90.0, Y: 0.0},
			Radius:  0.25,
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

	s.staticBodyID = box2d.Body{}
	return s
}

func (s *Sleep) ToggleInvoker() {
	if s.staticBodyID == (box2d.Body{}) {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -10.5, Y: 3.0}
		s.staticBodyID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(
			2.0,
			0.1,
			box2d.Vec2{X: 0.0, Y: 0.0},
			box2d.MakeRot(0.25*box2d.Pi),
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
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	imgui.Text("Pendulum Tuning")

	sleepVelocity := s.pendulumID.SleepThreshold()
	if imgui.SliderFloatV("sleep velocity", &sleepVelocity, 0.0, 1.0, "%.2f", 0) {
		s.pendulumID.SetSleepThreshold(sleepVelocity)
		s.pendulumID.SetAwake(true)
	}

	angularDamping := s.pendulumID.AngularDamping()
	if imgui.SliderFloatV("angular damping", &angularDamping, 0.0, 2.0, "%.2f", 0) {
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
			if event.SensorShape == s.sensorIDs[0] {
				s.sensorTouching[0] = true
			} else if event.SensorShape == s.sensorIDs[1] {
				s.sensorTouching[1] = true
			}
		}
	}

	for event := range s.worldID.SensorEndEvents() {
		if event.VisitorShape == s.groundShapeID {
			if event.SensorShape == s.sensorIDs[0] {
				s.sensorTouching[0] = false
			} else if event.SensorShape == s.sensorIDs[1] {
				s.sensorTouching[1] = false
			}
		}
	}

	for i := range 2 {
		s.DrawScreenTextLine("sensor touch %d = %t", i, s.sensorTouching[i])
	}
}

type BadBody struct {
	Sample
	badBodyID box2d.Body
}

func CreateBadBody(context *SampleContext) Sampler {
	s := &BadBody{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 2.3, Y: 10.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 3.0}
		bodyDef.AngularVelocity = 0.5
		bodyDef.Rotation = box2d.MakeRot(0.25 * box2d.Pi)

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
		bodyDef.Position = box2d.Pos{X: 2.0, Y: 3.0}
		bodyDef.Rotation = box2d.MakeRot(0.25 * box2d.Pi)

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
	s.Sample.Step()

	s.DrawScreenTextLine(
		"A bad body is a dynamic body with no mass and behaves like a kinematic body.",
	)
	s.DrawScreenTextLine(
		"Bad bodies are considered invalid and a user bug. Behavior is not guaranteed.",
	)

	s.badBodyID.ApplyForceToCenter(box2d.Vec2{X: 0.0, Y: 10.0}, true)
}

type Pivot struct {
	Sample
	bodyID box2d.Body
	lever  float32
}

func CreatePivot(context *SampleContext) Sampler {
	s := &Pivot{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.8, Y: 6.4}
		s.context.camera.zoom = 25.0 * 0.4
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		v := box2d.Vec2{X: 5.0, Y: 0.0}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 3.0}
		bodyDef.GravityScale = 1.0
		bodyDef.LinearVelocity = v

		s.bodyID = s.worldID.CreateBody(&bodyDef)

		s.lever = 3.0
		r := box2d.Vec2{X: 0.0, Y: -s.lever}

		omega := v.Cross(r) / r.Dot(r)
		s.bodyID.SetAngularVelocity(omega)

		box := box2d.MakeBox(0.1, s.lever)

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
	s.DrawScreenTextLine("pivot velocity = (%.6g, %.6g)", vp.X, vp.Y)
}

type Kinematic struct {
	Sample
	bodyID    box2d.Body
	amplitude float32
	time      float32
}

func CreateKinematic(context *SampleContext) Sampler {
	s := &Kinematic{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 4.0
	}

	s.amplitude = 2.0

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position.X = posScalar(2.0 * s.amplitude)

		s.bodyID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(0.1, 1.0)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	s.time = 0.0
	return s
}

func (s *Kinematic) Step() {
	var timeStep float32
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}
	if s.context.pause && s.context.singleStep == 0 {
		timeStep = 0.0
	}

	if timeStep > 0.0 {
		point := box2d.Pos{
			X: posScalar(2.0 * s.amplitude * float32(math.Cos(float64(s.time)))),
			Y: posScalar(s.amplitude * float32(math.Sin(float64(2.0*s.time)))),
		}
		rotation := box2d.MakeRot(2.0 * s.time)

		axis := rotation.Rotate(box2d.Vec2{X: 0.0, Y: 1.0})
		s.context.draw.DrawLine(
			point.Add(axis.Scale(-0.5)),
			point.Add(axis.Scale(0.5)),
			box2d.ColorPlum,
		)
		s.context.draw.DrawPoint(point, 10.0, box2d.ColorPlum)

		wake := true
		s.bodyID.SetTargetTransform(box2d.WorldTransform{P: point, Q: rotation}, timeStep, wake)
	}

	s.Sample.Step()

	s.time += timeStep
}

type MixedLocks struct {
	Sample
}

func CreateMixedLocks(context *SampleContext) Sampler {
	s := &MixedLocks{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.5}
		s.context.camera.zoom = 3.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -40.0, Y: 0.0},
			Point2: box2d.Vec2{X: 40.0, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	box := box2d.MakeSquare(0.5)
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

	create(box2d.StaticBody, box2d.Pos{X: 2.0, Y: 1.0}, box2d.MotionLocks{}, "static")
	create(box2d.DynamicBody, box2d.Pos{X: 1.0, Y: 1.0}, box2d.MotionLocks{}, "free")
	create(box2d.DynamicBody, box2d.Pos{X: 1.0, Y: 3.0}, box2d.MotionLocks{}, "free")
	create(
		box2d.DynamicBody,
		box2d.Pos{X: -1.0, Y: 1.0},
		box2d.MotionLocks{AngularZ: true},
		"angular z",
	)
	create(
		box2d.DynamicBody,
		box2d.Pos{X: -2.0, Y: 2.0},
		box2d.MotionLocks{LinearX: true},
		"linear x",
	)
	create(
		box2d.DynamicBody,
		box2d.Pos{X: -1.0, Y: 2.5},
		box2d.MotionLocks{LinearY: true, AngularZ: true},
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

type SetVelocity struct {
	Sample
	bodyID box2d.Body
}

func CreateSetVelocity(context *SampleContext) Sampler {
	s := &SetVelocity{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.5}
		s.context.camera.zoom = 3.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -0.25}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(20.0, 0.25)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeSquare(0.5)
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.5}
		s.bodyID = s.worldID.CreateBody(&bodyDef)
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *SetVelocity) Step() {
	s.Sample.Step()

	s.bodyID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: -20.0})

	position := s.bodyID.Position()
	s.DrawScreenTextLine("(x, y) = (%.2g, %.2g)", position.X, position.Y)
}

const wakeTouchingCount = 10

type WakeTouching struct {
	Sample
	groundID box2d.Body
}

func CreateWakeTouching(context *SampleContext) Sampler {
	s := &WakeTouching{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 4.0}
		s.context.camera.zoom = 8.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		s.groundID = s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		s.groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	box := box2d.MakeBox(0.5, 0.5)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	x := -1.0 * float32(wakeTouchingCount-1)

	for range wakeTouchingCount {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: 4.0}
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

func init() {
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
