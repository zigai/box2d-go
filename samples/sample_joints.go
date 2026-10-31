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

func sinf(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

func cosf(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

const distanceJointMaxCount = 10

// Test the distance joint and all options
type DistanceJoint struct {
	Sample

	groundID         box2d.Body
	bodyIDs          [distanceJointMaxCount]box2d.Body
	jointIDs         [distanceJointMaxCount]box2d.DistanceJoint
	count            int32
	hertz            float32
	dampingRatio     float32
	length           float32
	tensionForce     float32
	compressionForce float32
	minLength        float32
	maxLength        float32
	enableSpring     bool
	enableLimit      bool
}

func CreateDistanceJoint(context *SampleContext) Sampler {
	s := &DistanceJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 12.0}
		s.context.camera.zoom = 25.0 * 0.35
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		s.groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.count = 0
	s.hertz = 5.0
	s.dampingRatio = 0.5
	s.length = 1.0
	s.minLength = s.length
	s.maxLength = s.length
	s.tensionForce = 2000.0
	s.compressionForce = 100.0
	s.enableSpring = false
	s.enableLimit = false

	s.CreateScene(1)
	return s
}

func (s *DistanceJoint) CreateScene(newCount int32) {
	for i := range s.count {
		s.jointIDs[i].Destroy()
		s.jointIDs[i] = box2d.DistanceJoint{}
	}

	for i := range s.count {
		s.bodyIDs[i].Destroy()
		s.bodyIDs[i] = box2d.Body{}
	}

	s.count = newCount

	radius := float32(0.25)
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: radius}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 20.0

	yOffset := float32(20.0)

	jointDef := box2d.DefaultDistanceJointDef()
	jointDef.Hertz = s.hertz
	jointDef.DampingRatio = s.dampingRatio
	jointDef.Length = s.length
	jointDef.LowerSpringForce = -s.tensionForce
	jointDef.UpperSpringForce = s.compressionForce
	jointDef.MinLength = s.minLength
	jointDef.MaxLength = s.maxLength
	jointDef.EnableSpring = s.enableSpring
	jointDef.EnableLimit = s.enableLimit

	prevBodyID := s.groundID
	for i := range s.count {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.AngularDamping = 1.0
		bodyDef.Position = box2d.Pos{X: posScalar(s.length * (float32(i) + 1.0)), Y: posScalar(yOffset)}
		s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[i].CreateCircleShape(&shapeDef, &circle)

		pivotA := box2d.Pos{X: posScalar(s.length * float32(i)), Y: posScalar(yOffset)}
		pivotB := box2d.Pos{X: posScalar(s.length * (float32(i) + 1.0)), Y: posScalar(yOffset)}
		jointDef.Base.BodyA = prevBodyID
		jointDef.Base.BodyB = s.bodyIDs[i]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivotA)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivotB)
		s.jointIDs[i] = s.worldID.CreateDistanceJoint(&jointDef)

		prevBodyID = s.bodyIDs[i]
	}
}

func (s *DistanceJoint) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Length", &s.length, 0.1, 4.0, "%3.1f", 0) {
		for i := range s.count {
			s.jointIDs[i].SetLength(s.length)
			s.jointIDs[i].WakeBodies()
		}
	}

	if imgui.Checkbox("Spring", &s.enableSpring) {
		for i := range s.count {
			s.jointIDs[i].SetSpringEnabled(s.enableSpring)
			s.jointIDs[i].WakeBodies()
		}
	}

	if s.enableSpring {
		if imgui.SliderFloat("Tension", &s.tensionForce, 0.0, 4000.0) {
			for i := range s.count {
				s.jointIDs[i].SetSpringForceRange(-s.tensionForce, s.compressionForce)
				s.jointIDs[i].WakeBodies()
			}
		}

		if imgui.SliderFloat("Compression", &s.compressionForce, 0.0, 200.0) {
			for i := range s.count {
				s.jointIDs[i].SetSpringForceRange(-s.tensionForce, s.compressionForce)
				s.jointIDs[i].WakeBodies()
			}
		}

		if imgui.SliderFloatV("Hertz", &s.hertz, 0.0, 15.0, "%3.1f", 0) {
			for i := range s.count {
				s.jointIDs[i].SetSpringHertz(s.hertz)
				s.jointIDs[i].WakeBodies()
			}
		}

		if imgui.SliderFloatV("Damping", &s.dampingRatio, 0.0, 4.0, "%3.1f", 0) {
			for i := range s.count {
				s.jointIDs[i].SetSpringDampingRatio(s.dampingRatio)
				s.jointIDs[i].WakeBodies()
			}
		}
	}

	if imgui.Checkbox("Limit", &s.enableLimit) {
		for i := range s.count {
			s.jointIDs[i].SetLimitEnabled(s.enableLimit)
			s.jointIDs[i].WakeBodies()
		}
	}

	if s.enableLimit {
		if imgui.SliderFloatV("Min Length", &s.minLength, 0.1, 4.0, "%3.1f", 0) {
			for i := range s.count {
				s.jointIDs[i].SetLengthRange(s.minLength, s.maxLength)
				s.jointIDs[i].WakeBodies()
			}
		}

		if imgui.SliderFloatV("Max Length", &s.maxLength, 0.1, 4.0, "%3.1f", 0) {
			for i := range s.count {
				s.jointIDs[i].SetLengthRange(s.minLength, s.maxLength)
				s.jointIDs[i].WakeBodies()
			}
		}
	}

	count := s.count
	if imgui.SliderInt("Count", &count, 1, distanceJointMaxCount) {
		s.CreateScene(count)
	}

	imgui.PopItemWidth()

	return true
}

// This test shows how to use a motor joint. A motor joint
// can be used to animate a dynamic body. With finite motor forces
// the body can be blocked by collision with other bodies.
type MotorJoint struct {
	Sample

	targetID  box2d.Body
	bodyID    box2d.Body
	jointID   box2d.MotorJoint
	transform box2d.WorldTransform
	time      float32
	speed     float32
	maxForce  float32
	maxTorque float32
}

func CreateMotorJoint(context *SampleContext) Sampler {
	s := &MotorJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 7.0}
		s.context.camera.zoom = 25.0 * 0.4
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.transform = box2d.WorldTransform{P: box2d.Pos{X: 0.0, Y: 8.0}, Q: box2d.RotIdentity}

	// Define a target body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = s.transform.P
		s.targetID = s.worldID.CreateBody(&bodyDef)
	}

	// Define motorized body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = s.transform.P
		s.bodyID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(2.0, 0.5)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyID.CreatePolygonShape(&shapeDef, &box)

		s.maxForce = 5000.0
		s.maxTorque = 500.0

		jointDef := box2d.DefaultMotorJointDef()
		jointDef.Base.BodyA = s.targetID
		jointDef.Base.BodyB = s.bodyID
		jointDef.LinearHertz = 4.0
		jointDef.LinearDampingRatio = 0.7
		jointDef.AngularHertz = 4.0
		jointDef.AngularDampingRatio = 0.7
		jointDef.MaxSpringForce = s.maxForce
		jointDef.MaxSpringTorque = s.maxTorque

		s.jointID = s.worldID.CreateMotorJoint(&jointDef)
	}

	// Define spring body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -2.0, Y: 2.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeSquare(0.5)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultMotorJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = bodyDef.Position.Vec2().Add(box2d.Vec2{X: 0.25, Y: 0.25})
		jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.25, Y: 0.25}
		jointDef.LinearHertz = 7.5
		jointDef.LinearDampingRatio = 0.7
		jointDef.AngularHertz = 7.5
		jointDef.AngularDampingRatio = 0.7
		jointDef.MaxSpringForce = 500.0
		jointDef.MaxSpringTorque = 10.0

		s.worldID.CreateMotorJoint(&jointDef)
	}

	s.speed = 1.0
	s.time = 0.0
	return s
}

func (s *MotorJoint) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	imgui.SliderFloatV("Speed", &s.speed, -5.0, 5.0, "%.0f", 0)

	if imgui.SliderFloatV("Max Force", &s.maxForce, 0.0, 10000.0, "%.0f", 0) {
		s.jointID.SetMaxSpringForce(s.maxForce)
	}

	if imgui.SliderFloatV("Max Torque", &s.maxTorque, 0.0, 10000.0, "%.0f", 0) {
		s.jointID.SetMaxSpringTorque(s.maxTorque)
	}

	imgui.PopItemWidth()

	if imgui.Button("Apply Impulse") {
		s.bodyID.ApplyLinearImpulseToCenter(box2d.Vec2{X: 100.0, Y: 0.0}, true)
	}

	return true
}

func (s *MotorJoint) Step() {
	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}

	if s.context.pause {
		if s.context.singleStep == 0 {
			timeStep = 0.0
		}
	}

	if timeStep > 0.0 {
		s.time += s.speed * timeStep

		var linearOffset box2d.Pos
		linearOffset.X = posScalar(6.0 * sinf(2.0*s.time))
		linearOffset.Y = posScalar(8.0 + 4.0*sinf(1.0*s.time))

		angularOffset := 2.0 * s.time
		s.transform = box2d.WorldTransform{P: linearOffset, Q: box2d.MakeRot(angularOffset)}

		wake := true
		s.targetID.SetTargetTransform(s.transform, timeStep, wake)
	}

	s.draw.DrawTransform(s.transform, 1.0)

	s.Sample.Step()

	force := s.jointID.ConstraintForce()
	torque := s.jointID.ConstraintTorque()

	s.DrawScreenTextLine("force = {%3.f, %3.f}, torque = %3.f", force.X, force.Y, torque)
}

type TopDownFriction struct {
	Sample
}

func CreateTopDownFriction(context *SampleContext) Sampler {
	s := &TopDownFriction{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 10.0}
		s.context.camera.zoom = 12.0
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: -10.0, Y: 20.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 20.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 20.0}, Point2: box2d.Vec2{X: 10.0, Y: 20.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	jointDef := box2d.DefaultMotorJointDef()
	jointDef.Base.BodyA = groundID
	jointDef.Base.CollideConnected = true
	jointDef.MaxVelocityForce = 10.0
	jointDef.MaxVelocityTorque = 10.0

	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.25, Y: 0.0}, Center2: box2d.Vec2{X: 0.25, Y: 0.0}, Radius: 0.25}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.35}
	square := box2d.MakeSquare(0.35)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.0
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 0.8

	n := 10
	x, y := float32(-5.0), float32(15.0)
	for i := range n {
		for j := range n {
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)

			remainder := (n*i + j) % 4
			if remainder == 0 {
				bodyID.CreateCapsuleShape(&shapeDef, &capsule)
			} else if remainder == 1 {
				bodyID.CreateCircleShape(&shapeDef, &circle)
			} else if remainder == 2 {
				bodyID.CreatePolygonShape(&shapeDef, &square)
			} else {
				poly := scenes.RandomPolygon(0.75)
				poly.Radius = 0.1
				bodyID.CreatePolygonShape(&shapeDef, &poly)
			}

			jointDef.Base.BodyB = bodyID
			s.worldID.CreateMotorJoint(&jointDef)

			x += 1.0
		}

		x = -5.0
		y -= 1.0
	}
	return s
}

func (s *TopDownFriction) DrawControls() bool {
	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()
		def.Position = box2d.Pos{X: 0.0, Y: 10.0}
		def.Radius = 10.0
		def.Falloff = 5.0
		def.ImpulsePerLength = 10.0
		s.worldID.Explode(&def)

		s.draw.DrawCircle(def.Position, 10.0, box2d.ColorWhite)
	}

	return true
}

// This sample shows how to use a filter joint to prevent collision between two bodies.
// This is more specific than shape filters. It also shows that sleeping is coupled by the filter joint.
type FilterJoint struct {
	Sample
}

func CreateFilterJoint(context *SampleContext) Sampler {
	s := &FilterJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 7.0}
		s.context.camera.zoom = 25.0 * 0.4
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -4.0, Y: 2.0}
		bodyID1 := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeSquare(2.0)
		shapeDef := box2d.DefaultShapeDef()
		bodyID1.CreatePolygonShape(&shapeDef, &box)

		bodyDef.Position = box2d.Pos{X: 4.0, Y: 2.0}
		bodyID2 := s.worldID.CreateBody(&bodyDef)
		bodyID2.CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultFilterJointDef()
		jointDef.Base.BodyA = bodyID1
		jointDef.Base.BodyB = bodyID2

		s.worldID.CreateFilterJoint(&jointDef)
	}
	return s
}

type RevoluteJoint struct {
	Sample

	ball          box2d.Body
	jointID1      box2d.RevoluteJoint
	jointID2      box2d.RevoluteJoint
	motorSpeed    float32
	motorTorque   float32
	hertz         float32
	dampingRatio  float32
	targetDegrees float32
	enableSpring  bool
	enableMotor   bool
	enableLimit   bool
}

func CreateRevoluteJoint(context *SampleContext) Sampler {
	s := &RevoluteJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 15.5}
		s.context.camera.zoom = 25.0 * 0.7
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(40.0, 1.0)

		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.enableSpring = false
	s.enableLimit = false
	s.enableMotor = false
	s.hertz = 2.0
	s.dampingRatio = 0.5
	s.targetDegrees = 45.0
	s.motorSpeed = 1.0
	s.motorTorque = 1000.0

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -10.0, Y: 20.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -1.0}, Center2: box2d.Vec2{X: 0.0, Y: 6.0}, Radius: 0.5}
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)

		pivot := box2d.Pos{X: -10.0, Y: 20.5}
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.Q = box2d.MakeRot(0.5 * box2d.Pi)
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.TargetAngle = box2d.Pi * s.targetDegrees / 180.0
		jointDef.EnableSpring = s.enableSpring
		jointDef.Hertz = s.hertz
		jointDef.DampingRatio = s.dampingRatio
		jointDef.MotorSpeed = s.motorSpeed
		jointDef.MaxMotorTorque = s.motorTorque
		jointDef.EnableMotor = s.enableMotor
		jointDef.LowerAngle = -0.5 * box2d.Pi
		jointDef.UpperAngle = 0.05 * box2d.Pi
		jointDef.EnableLimit = s.enableLimit

		s.jointID1 = s.worldID.CreateRevoluteJoint(&jointDef)

		s.jointID1.SetConstraintTuning(60.0, 20.0)
	}

	{
		var circle box2d.Circle
		circle.Radius = 2.0

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 5.0, Y: 30.0}
		s.ball = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0

		s.ball.CreateCircleShape(&shapeDef, &circle)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 20.0, Y: 10.0}
		bodyDef.Type = box2d.DynamicBody
		body := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeOffsetBox(10.0, 0.5, box2d.Vec2{X: -10.0, Y: 0.0}, box2d.RotIdentity)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		body.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: 19.0, Y: 10.0}
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = body
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.LowerAngle = -0.25 * box2d.Pi
		jointDef.UpperAngle = 0.0 * box2d.Pi
		jointDef.EnableLimit = true
		jointDef.EnableMotor = true
		jointDef.MotorSpeed = 0.0
		jointDef.MaxMotorTorque = s.motorTorque

		s.jointID2 = s.worldID.CreateRevoluteJoint(&jointDef)
	}
	return s
}

func (s *RevoluteJoint) DrawControls() bool {
	if imgui.Checkbox("Limit", &s.enableLimit) {
		s.jointID1.SetLimitEnabled(s.enableLimit)
		s.jointID1.WakeBodies()
	}

	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.jointID1.SetMotorEnabled(s.enableMotor)
		s.jointID1.WakeBodies()
	}

	if s.enableMotor {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		if imgui.SliderFloatV("Max Torque", &s.motorTorque, 0.0, 5000.0, "%.0f", 0) {
			s.jointID1.SetMaxMotorTorque(s.motorTorque)
			s.jointID1.WakeBodies()
		}

		if imgui.SliderFloatV("Speed", &s.motorSpeed, -20.0, 20.0, "%.0f", 0) {
			s.jointID1.SetMotorSpeed(s.motorSpeed)
			s.jointID1.WakeBodies()
		}
		imgui.PopItemWidth()
	}

	if imgui.Checkbox("Spring", &s.enableSpring) {
		s.jointID1.SetSpringEnabled(s.enableSpring)
		s.jointID1.WakeBodies()
	}

	if s.enableSpring {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		if imgui.SliderFloatV("Hertz", &s.hertz, 0.0, 30.0, "%.1f", 0) {
			s.jointID1.SetSpringHertz(s.hertz)
			s.jointID1.WakeBodies()
		}

		if imgui.SliderFloatV("Damping", &s.dampingRatio, 0.0, 2.0, "%.1f", 0) {
			s.jointID1.SetSpringDampingRatio(s.dampingRatio)
			s.jointID1.WakeBodies()
		}

		if imgui.SliderFloatV("Degrees", &s.targetDegrees, -180.0, 180.0, "%.0f", 0) {
			s.jointID1.SetTargetAngle(box2d.Pi * s.targetDegrees / 180.0)
			s.jointID1.WakeBodies()
		}
		imgui.PopItemWidth()
	}

	return true
}

func (s *RevoluteJoint) Step() {
	s.Sample.Step()

	angle1 := s.jointID1.Angle()
	s.DrawScreenTextLine("Angle (Deg) 1 = %2.1f", angle1)

	torque1 := s.jointID1.MotorTorque()
	s.DrawScreenTextLine("Motor Torque 1 = %4.1f", torque1)

	torque2 := s.jointID2.MotorTorque()
	s.DrawScreenTextLine("Motor Torque 2 = %4.1f", torque2)
}

type PrismaticJoint struct {
	Sample

	jointID      box2d.PrismaticJoint
	motorSpeed   float32
	motorForce   float32
	hertz        float32
	dampingRatio float32
	translation  float32
	enableSpring bool
	enableMotor  bool
	enableLimit  bool
}

func CreatePrismaticJoint(context *SampleContext) Sampler {
	s := &PrismaticJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 8.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.enableSpring = false
	s.enableLimit = true
	s.enableMotor = false
	s.motorSpeed = 2.0
	s.motorForce = 25.0
	s.hertz = 1.0
	s.dampingRatio = 0.5
	s.translation = 0.0

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 10.0}
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(0.5, 2.0)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: 0.0, Y: 9.0}
		axis := box2d.Vec2{X: 1.0, Y: 1.0}.Normalize()
		jointDef := box2d.DefaultPrismaticJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameA.Q = box2d.MakeRotFromUnitVector(axis)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.Q = box2d.MakeRotFromUnitVector(axis)
		jointDef.Base.DrawScale = 2.0
		jointDef.MotorSpeed = s.motorSpeed
		jointDef.MaxMotorForce = s.motorForce
		jointDef.EnableMotor = s.enableMotor
		jointDef.LowerTranslation = -10.0
		jointDef.UpperTranslation = 10.0
		jointDef.EnableLimit = s.enableLimit
		jointDef.EnableSpring = s.enableSpring
		jointDef.Hertz = s.hertz
		jointDef.DampingRatio = s.dampingRatio

		s.jointID = s.worldID.CreatePrismaticJoint(&jointDef)
	}
	return s
}

func (s *PrismaticJoint) DrawControls() bool {
	if imgui.Checkbox("Limit", &s.enableLimit) {
		s.jointID.SetLimitEnabled(s.enableLimit)
		s.jointID.WakeBodies()
	}

	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.jointID.SetMotorEnabled(s.enableMotor)
		s.jointID.WakeBodies()
	}

	if s.enableMotor {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		if imgui.SliderFloatV("Max Force", &s.motorForce, 0.0, 200.0, "%.0f", 0) {
			s.jointID.SetMaxMotorForce(s.motorForce)
			s.jointID.WakeBodies()
		}

		if imgui.SliderFloatV("Speed", &s.motorSpeed, -40.0, 40.0, "%.0f", 0) {
			s.jointID.SetMotorSpeed(s.motorSpeed)
			s.jointID.WakeBodies()
		}
		imgui.PopItemWidth()
	}

	if imgui.Checkbox("Spring", &s.enableSpring) {
		s.jointID.SetSpringEnabled(s.enableSpring)
		s.jointID.WakeBodies()
	}

	if s.enableSpring {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		if imgui.SliderFloatV("Hertz", &s.hertz, 0.0, 10.0, "%.1f", 0) {
			s.jointID.SetSpringHertz(s.hertz)
			s.jointID.WakeBodies()
		}

		if imgui.SliderFloatV("Damping", &s.dampingRatio, 0.0, 2.0, "%.1f", 0) {
			s.jointID.SetSpringDampingRatio(s.dampingRatio)
			s.jointID.WakeBodies()
		}

		if imgui.SliderFloatV("Translation", &s.translation, -15.0, 15.0, "%.1f", 0) {
			s.jointID.SetTargetTranslation(s.translation)
			s.jointID.WakeBodies()
		}
		imgui.PopItemWidth()
	}

	return true
}

func (s *PrismaticJoint) Step() {
	s.Sample.Step()

	force := s.jointID.MotorForce()
	s.DrawScreenTextLine("Motor Force = %4.1f", force)

	translation := s.jointID.Translation()
	s.DrawScreenTextLine("Translation = %4.1f", translation)

	speed := s.jointID.Speed()
	s.DrawScreenTextLine("Speed = %4.8f", speed)
}

type WheelJoint struct {
	Sample

	jointID      box2d.WheelJoint
	hertz        float32
	dampingRatio float32
	motorSpeed   float32
	motorTorque  float32
	enableSpring bool
	enableMotor  bool
	enableLimit  bool
}

func CreateWheelJoint(context *SampleContext) Sampler {
	s := &WheelJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 10.0}
		s.context.camera.zoom = 25.0 * 0.15
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.enableSpring = true
	s.enableLimit = true
	s.enableMotor = true
	s.motorSpeed = 2.0
	s.motorTorque = 5.0
	s.hertz = 1.0
	s.dampingRatio = 0.7

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 10.25}
	bodyDef.Type = box2d.DynamicBody
	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.5}, Center2: box2d.Vec2{X: 0.0, Y: 0.5}, Radius: 0.5}
	bodyID.CreateCapsuleShape(&shapeDef, &capsule)

	pivot := box2d.Pos{X: 0.0, Y: 10.0}
	axis := box2d.Vec2{X: 1.0, Y: 1.0}.Normalize()
	jointDef := box2d.DefaultWheelJointDef()
	jointDef.Base.BodyA = groundID
	jointDef.Base.BodyB = bodyID
	jointDef.Base.LocalFrameA.Q = box2d.MakeRotFromUnitVector(axis)
	jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
	jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
	jointDef.MotorSpeed = s.motorSpeed
	jointDef.MaxMotorTorque = s.motorTorque
	jointDef.EnableMotor = s.enableMotor
	jointDef.LowerTranslation = -3.0
	jointDef.UpperTranslation = 3.0
	jointDef.EnableLimit = s.enableLimit
	jointDef.Hertz = s.hertz
	jointDef.DampingRatio = s.dampingRatio

	s.jointID = s.worldID.CreateWheelJoint(&jointDef)
	return s
}

func (s *WheelJoint) DrawControls() bool {
	if imgui.Checkbox("Limit", &s.enableLimit) {
		s.jointID.SetLimitEnabled(s.enableLimit)
	}

	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.jointID.SetMotorEnabled(s.enableMotor)
	}

	if s.enableMotor {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		if imgui.SliderFloatV("Torque", &s.motorTorque, 0.0, 20.0, "%.0f", 0) {
			s.jointID.SetMaxMotorTorque(s.motorTorque)
		}

		if imgui.SliderFloatV("Speed", &s.motorSpeed, -20.0, 20.0, "%.0f", 0) {
			s.jointID.SetMotorSpeed(s.motorSpeed)
		}
		imgui.PopItemWidth()
	}

	if imgui.Checkbox("Spring", &s.enableSpring) {
		s.jointID.SetSpringEnabled(s.enableSpring)
	}

	if s.enableSpring {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		if imgui.SliderFloatV("Hertz", &s.hertz, 0.0, 10.0, "%.1f", 0) {
			s.jointID.SetSpringHertz(s.hertz)
		}

		if imgui.SliderFloatV("Damping", &s.dampingRatio, 0.0, 2.0, "%.1f", 0) {
			s.jointID.SetSpringDampingRatio(s.dampingRatio)
		}
		imgui.PopItemWidth()
	}

	return true
}

func (s *WheelJoint) Step() {
	s.Sample.Step()

	torque := s.jointID.MotorTorque()
	s.DrawScreenTextLine("Motor Torque = %4.1f", torque)
}

const bridgeCount = 160

// A suspension bridge
type Bridge struct {
	Sample

	bodyIDs                [bridgeCount]box2d.Body
	jointIDs               [bridgeCount + 1]box2d.RevoluteJoint
	frictionTorque         float32
	constraintHertz        float32
	constraintDampingRatio float32
	springHertz            float32
	springDampingRatio     float32
}

func CreateBridge(context *SampleContext) Sampler {
	s := &Bridge{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 2.5
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	{
		s.constraintHertz = 60.0
		s.constraintDampingRatio = 0.0
		s.springHertz = 2.0
		s.springDampingRatio = 0.7
		s.frictionTorque = 200.0

		box := box2d.MakeBox(0.5, 0.125)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 20.0

		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.EnableMotor = true
		jointDef.MaxMotorTorque = s.frictionTorque
		jointDef.EnableSpring = true
		jointDef.Hertz = s.springHertz
		jointDef.DampingRatio = s.springDampingRatio

		jointIndex := 0

		xbase := float32(-80.0)

		prevBodyID := groundID
		for i := range bridgeCount {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(xbase + 0.5 + 1.0*float32(i)), Y: 20.0}
			bodyDef.LinearDamping = 0.1
			bodyDef.AngularDamping = 0.1

			s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)

			s.bodyIDs[i].CreatePolygonShape(&shapeDef, &box)

			pivot := box2d.Pos{X: posScalar(xbase + 1.0*float32(i)), Y: 20.0}
			jointDef.Base.BodyA = prevBodyID
			jointDef.Base.BodyB = s.bodyIDs[i]
			jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
			jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
			s.jointIDs[jointIndex] = s.worldID.CreateRevoluteJoint(&jointDef)
			jointIndex++

			prevBodyID = s.bodyIDs[i]
		}

		pivot := box2d.Pos{X: posScalar(xbase + 1.0*bridgeCount), Y: 20.0}
		jointDef.Base.BodyA = prevBodyID
		jointDef.Base.BodyB = groundID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		s.jointIDs[jointIndex] = s.worldID.CreateRevoluteJoint(&jointDef)
	}

	for i := range 2 {
		vertices := []box2d.Vec2{{X: -0.5, Y: 0.0}, {X: 0.5, Y: 0.0}, {X: 0.0, Y: 1.5}}

		hull := box2d.ComputeHull(vertices)
		triangle := box2d.MakePolygon(&hull, 0.0)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 20.0

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-8.0 + 8.0*float32(i)), Y: 22.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &triangle)
	}

	for i := range 3 {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 20.0

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-6.0 + 6.0*float32(i)), Y: 25.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
	return s
}

func (s *Bridge) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	updateFriction := imgui.SliderFloatV("Joint Friction", &s.frictionTorque, 0.0, 10000.0, "%2.f", 0)
	if updateFriction {
		for i := 0; i <= bridgeCount; i++ {
			s.jointIDs[i].SetMaxMotorTorque(s.frictionTorque)
		}
	}

	if imgui.SliderFloatV("Spring hertz", &s.springHertz, 0.0, 30.0, "%.0f", 0) {
		for i := 0; i <= bridgeCount; i++ {
			s.jointIDs[i].SetSpringHertz(s.springHertz)
		}
	}

	if imgui.SliderFloatV("Spring damping", &s.springDampingRatio, 0.0, 2.0, "%.1f", 0) {
		for i := 0; i <= bridgeCount; i++ {
			s.jointIDs[i].SetSpringDampingRatio(s.springDampingRatio)
		}
	}

	if imgui.SliderFloatV("Constraint hertz", &s.constraintHertz, 15.0, 240.0, "%.0f", 0) {
		for i := 0; i <= bridgeCount; i++ {
			s.jointIDs[i].SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)
		}
	}

	if imgui.SliderFloatV("Constraint damping", &s.constraintDampingRatio, 0.0, 10.0, "%.1f", 0) {
		for i := 0; i <= bridgeCount; i++ {
			s.jointIDs[i].SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)
		}
	}

	imgui.PopItemWidth()

	return true
}

const ballAndChainCount = 30

type BallAndChain struct {
	Sample

	jointIDs       [ballAndChainCount + 1]box2d.RevoluteJoint
	frictionTorque float32
}

func CreateBallAndChain(context *SampleContext) Sampler {
	s := &BallAndChain{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: -8.0}
		s.context.camera.zoom = 27.5
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.frictionTorque = 100.0

	{
		hx := float32(0.5)
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -hx, Y: 0.0}, Center2: box2d.Vec2{X: hx, Y: 0.0}, Radius: 0.125}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 20.0
		shapeDef.Filter.CategoryBits = 0x1
		shapeDef.Filter.MaskBits = 0x2
		jointDef := box2d.DefaultRevoluteJointDef()

		jointIndex := 0

		prevBodyID := groundID
		for i := range ballAndChainCount {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar((1.0 + 2.0*float32(i)) * hx), Y: posScalar(ballAndChainCount * hx)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreateCapsuleShape(&shapeDef, &capsule)

			pivot := box2d.Pos{X: posScalar((2.0 * float32(i)) * hx), Y: posScalar(ballAndChainCount * hx)}
			jointDef.Base.BodyA = prevBodyID
			jointDef.Base.BodyB = bodyID
			jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
			jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
			jointDef.EnableMotor = true
			jointDef.MaxMotorTorque = s.frictionTorque
			jointDef.EnableSpring = i > 0
			jointDef.Hertz = 4.0
			s.jointIDs[jointIndex] = s.worldID.CreateRevoluteJoint(&jointDef)
			jointIndex++

			prevBodyID = bodyID
		}

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 4.0}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar((1.0+2.0*ballAndChainCount)*hx + circle.Radius - hx), Y: posScalar(ballAndChainCount * hx)}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef.Filter.CategoryBits = 0x2
		shapeDef.Filter.MaskBits = 0x1
		bodyID.CreateCircleShape(&shapeDef, &circle)

		pivot := box2d.Pos{X: posScalar((2.0 * ballAndChainCount) * hx), Y: posScalar(ballAndChainCount * hx)}
		jointDef.Base.BodyA = prevBodyID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableMotor = true
		jointDef.MaxMotorTorque = s.frictionTorque
		jointDef.EnableSpring = true
		jointDef.Hertz = 4.0
		s.jointIDs[jointIndex] = s.worldID.CreateRevoluteJoint(&jointDef)
	}
	return s
}

func (s *BallAndChain) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	updateFriction := imgui.SliderFloatV("Joint Friction", &s.frictionTorque, 0.0, 1000.0, "%2.f", 0)
	imgui.PopItemWidth()
	if updateFriction {
		for i := 0; i <= ballAndChainCount; i++ {
			s.jointIDs[i].SetMaxMotorTorque(s.frictionTorque)
		}
	}

	return true
}

const cantileverCount = 8

// This sample shows the limitations of an iterative solver. The cantilever sags even though the weld
// joint is stiff as possible.
type Cantilever struct {
	Sample

	linearHertz         float32
	linearDampingRatio  float32
	angularHertz        float32
	angularDampingRatio float32
	gravityScale        float32
	tipID               box2d.Body
	bodyIDs             [cantileverCount]box2d.Body
	jointIDs            [cantileverCount]box2d.WeldJoint
	collideConnected    bool
}

func CreateCantilever(context *SampleContext) Sampler {
	s := &Cantilever{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 0.35
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	{
		s.linearHertz = 15.0
		s.linearDampingRatio = 0.5
		s.angularHertz = 5.0
		s.angularDampingRatio = 0.5
		s.gravityScale = 1.0
		s.collideConnected = false

		hx := float32(0.5)
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -hx, Y: 0.0}, Center2: box2d.Vec2{X: hx, Y: 0.0}, Radius: 0.125}
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 20.0

		jointDef := box2d.DefaultWeldJointDef()

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.IsAwake = false

		prevBodyID := groundID
		for i := range cantileverCount {
			bodyDef.Position = box2d.Pos{X: posScalar((1.0 + 2.0*float32(i)) * hx), Y: 0.0}
			s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)
			s.bodyIDs[i].CreateCapsuleShape(&shapeDef, &capsule)

			pivot := box2d.Pos{X: posScalar((2.0 * float32(i)) * hx), Y: 0.0}
			jointDef.Base.BodyA = prevBodyID
			jointDef.Base.BodyB = s.bodyIDs[i]
			jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
			jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
			jointDef.LinearHertz = s.linearHertz
			jointDef.LinearDampingRatio = s.linearDampingRatio
			jointDef.AngularHertz = s.angularHertz
			jointDef.AngularDampingRatio = s.angularDampingRatio
			jointDef.Base.CollideConnected = s.collideConnected
			s.jointIDs[i] = s.worldID.CreateWeldJoint(&jointDef)

			// Experimental tuning
			s.jointIDs[i].SetConstraintTuning(120.0, 10.0)

			prevBodyID = s.bodyIDs[i]
		}

		s.tipID = prevBodyID
	}
	return s
}

func (s *Cantilever) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Linear Hertz", &s.linearHertz, 0.0, 20.0, "%.0f", 0) {
		for i := range cantileverCount {
			s.jointIDs[i].SetLinearHertz(s.linearHertz)
		}
	}

	if imgui.SliderFloatV("Linear Damping Ratio", &s.linearDampingRatio, 0.0, 10.0, "%.1f", 0) {
		for i := range cantileverCount {
			s.jointIDs[i].SetLinearDampingRatio(s.linearDampingRatio)
		}
	}

	if imgui.SliderFloatV("Angular Hertz", &s.angularHertz, 0.0, 20.0, "%.0f", 0) {
		for i := range cantileverCount {
			s.jointIDs[i].SetAngularHertz(s.angularHertz)
		}
	}

	if imgui.SliderFloatV("Angular Damping Ratio", &s.angularDampingRatio, 0.0, 10.0, "%.1f", 0) {
		for i := range cantileverCount {
			s.jointIDs[i].SetAngularDampingRatio(s.angularDampingRatio)
		}
	}

	if imgui.SliderFloatV("Gravity Scale", &s.gravityScale, -1.0, 1.0, "%.1f", 0) {
		for i := range cantileverCount {
			s.bodyIDs[i].SetGravityScale(s.gravityScale)
		}
	}

	imgui.PopItemWidth()

	if imgui.Checkbox("Collide Connected", &s.collideConnected) {
		for i := range cantileverCount {
			s.jointIDs[i].SetCollideConnected(s.collideConnected)
		}
	}

	return true
}

func (s *Cantilever) Step() {
	s.Sample.Step()

	tipPosition := s.tipID.Position()
	s.DrawScreenTextLine("tip-y = %.2f", tipPosition.Y)
}

const motionLocksCount = 6

// This test ensures joints work correctly with bodies that have motion locks
type MotionLocks struct {
	Sample

	bodyIDs     [motionLocksCount]box2d.Body
	motionLocks box2d.MotionLocks
}

func CreateMotionLocks(context *SampleContext) Sampler {
	s := &MotionLocks{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 8.0}
		s.context.camera.zoom = 25.0 * 0.7
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	s.motionLocks = box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: true}

	position := box2d.Pos{X: -12.5, Y: 10.0}
	bodyDef.Type = box2d.DynamicBody
	bodyDef.MotionLocks = s.motionLocks

	box := box2d.MakeBox(1.0, 1.0)

	index := 0

	// distance joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		length := float32(2.0)
		pivot1 := box2d.Pos{X: position.X, Y: position.Y + 1.0 + posScalar(length)}
		pivot2 := box2d.Pos{X: position.X, Y: position.Y + 1.0}
		jointDef := box2d.DefaultDistanceJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot1)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot2)
		jointDef.Length = length
		s.worldID.CreateDistanceJoint(&jointDef)
	}

	position.X += 5.0
	index++

	// motor joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultMotorJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = position.Vec2()
		jointDef.MaxVelocityForce = 200.0
		jointDef.MaxVelocityTorque = 200.0
		s.worldID.CreateMotorJoint(&jointDef)
	}

	position.X += 5.0
	index++

	// prismatic joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultPrismaticJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		s.worldID.CreatePrismaticJoint(&jointDef)
	}

	position.X += 5.0
	index++

	// revolute joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		s.worldID.CreateRevoluteJoint(&jointDef)
	}

	position.X += 5.0
	index++

	// weld joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultWeldJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.AngularHertz = 1.0
		jointDef.AngularDampingRatio = 0.5
		jointDef.LinearHertz = 1.0
		jointDef.LinearDampingRatio = 0.5
		s.worldID.CreateWeldJoint(&jointDef)
	}

	position.X += 5.0
	index++

	// wheel joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultWheelJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
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
		s.worldID.CreateWheelJoint(&jointDef)
	}
	return s
}

func (s *MotionLocks) applyLocks() {
	for i := range motionLocksCount {
		s.bodyIDs[i].SetMotionLocks(s.motionLocks)
		s.bodyIDs[i].SetAwake(true)
	}
}

func (s *MotionLocks) DrawControls() bool {
	if imgui.Checkbox("Lock Linear X", &s.motionLocks.LinearX) {
		s.applyLocks()
	}

	if imgui.Checkbox("Lock Linear Y", &s.motionLocks.LinearY) {
		s.applyLocks()
	}

	if imgui.Checkbox("Lock Angular Z", &s.motionLocks.AngularZ) {
		s.applyLocks()
	}

	if s.context.window.GetKey(glfw.KeyL) == glfw.Press {
		s.bodyIDs[0].ApplyLinearImpulseToCenter(box2d.Vec2{X: 100.0, Y: 0.0}, true)
	}

	return true
}

const breakableJointCount = 6

// This sample shows how to break joints when the internal reaction force becomes large.
type BreakableJoint struct {
	Sample

	jointIDs   [breakableJointCount]box2d.Joint
	breakForce float32
}

func CreateBreakableJoint(context *SampleContext) Sampler {
	s := &BreakableJoint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 8.0}
		s.context.camera.zoom = 25.0 * 0.7
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{Point1: box2d.Vec2{X: -40.0, Y: 0.0}, Point2: box2d.Vec2{X: 40.0, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	position := box2d.Pos{X: -12.5, Y: 10.0}
	bodyDef.Type = box2d.DynamicBody
	bodyDef.EnableSleep = false

	box := box2d.MakeBox(1.0, 1.0)

	index := 0

	// distance joint
	{
		bodyDef.Position = position
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		length := float32(2.0)
		pivot1 := box2d.Pos{X: position.X, Y: position.Y + 1.0 + posScalar(length)}
		pivot2 := box2d.Pos{X: position.X, Y: position.Y + 1.0}
		jointDef := box2d.DefaultDistanceJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot1)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot2)
		jointDef.Length = length
		jointDef.Base.CollideConnected = true
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
		jointDef.Base.CollideConnected = true
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
		jointDef.Base.CollideConnected = true
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
		jointDef.Base.CollideConnected = true
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
		jointDef.Base.CollideConnected = true
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
		jointDef.Base.CollideConnected = true
		s.jointIDs[index] = s.worldID.CreateWheelJoint(&jointDef).Joint
	}

	s.breakForce = 1000.0
	return s
}

func (s *BreakableJoint) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	imgui.SliderFloatV("break force", &s.breakForce, 0.0, 10000.0, "%.1f", 0)

	gravity := s.worldID.Gravity()
	if imgui.SliderFloatV("gravity", &gravity.Y, -50.0, 50.0, "%.1f", 0) {
		s.worldID.SetGravity(gravity)
	}

	imgui.PopItemWidth()

	return true
}

func (s *BreakableJoint) Step() {
	for i := range breakableJointCount {
		if s.jointIDs[i] == (box2d.Joint{}) {
			continue
		}

		force := s.jointIDs[i].ConstraintForce()
		if force.LengthSquared() > s.breakForce*s.breakForce {
			s.jointIDs[i].Destroy()
			s.jointIDs[i] = box2d.Joint{}
		} else {
			localFrame := s.jointIDs[i].LocalFrameA()
			s.draw.DrawString(s.camera, localFrame.P.Pos(), box2d.ColorWhite, "(%.1f, %.1f)", force.X, force.Y)
		}
	}

	s.Sample.Step()
}

const jointSeparationCount = 5

// This sample shows how to measure joint separation. This is the unresolved constraint error.
type JointSeparation struct {
	Sample

	bodyIDs           [jointSeparationCount]box2d.Body
	jointIDs          [jointSeparationCount]box2d.Joint
	impulse           float32
	jointHertz        float32
	jointDampingRatio float32
}

func CreateJointSeparation(context *SampleContext) Sampler {
	s := &JointSeparation{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 8.0}
		s.context.camera.zoom = 25.0
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{Point1: box2d.Vec2{X: -40.0, Y: 0.0}, Point2: box2d.Vec2{X: 40.0, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	position := box2d.Pos{X: -20.0, Y: 10.0}
	bodyDef.Type = box2d.DynamicBody
	bodyDef.EnableSleep = false

	box := box2d.MakeBox(1.0, 1.0)

	index := 0

	// distance joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		length := float32(2.0)
		pivot1 := box2d.Pos{X: position.X, Y: position.Y + 1.0 + posScalar(length)}
		pivot2 := box2d.Pos{X: position.X, Y: position.Y + 1.0}
		jointDef := box2d.DefaultDistanceJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot1)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot2)
		jointDef.Length = length
		jointDef.Base.CollideConnected = true
		s.jointIDs[index] = s.worldID.CreateDistanceJoint(&jointDef).Joint
	}

	position.X += 10.0
	index++

	// prismatic joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultPrismaticJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Base.CollideConnected = true
		s.jointIDs[index] = s.worldID.CreatePrismaticJoint(&jointDef).Joint
	}

	position.X += 10.0
	index++

	// revolute joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Base.CollideConnected = true
		s.jointIDs[index] = s.worldID.CreateRevoluteJoint(&jointDef).Joint
	}

	position.X += 10.0
	index++

	// weld joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultWeldJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.Base.CollideConnected = true
		s.jointIDs[index] = s.worldID.CreateWeldJoint(&jointDef).Joint
	}

	position.X += 10.0
	index++

	// wheel joint
	{
		bodyDef.Position = position
		s.bodyIDs[index] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[index].CreatePolygonShape(&shapeDef, &box)

		pivot := box2d.Pos{X: position.X - 1.0, Y: position.Y}
		jointDef := box2d.DefaultWheelJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.bodyIDs[index]
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
		jointDef.Base.CollideConnected = true
		s.jointIDs[index] = s.worldID.CreateWheelJoint(&jointDef).Joint
	}

	s.impulse = 500.0
	s.jointHertz = 60.0
	s.jointDampingRatio = 2.0
	return s
}

func (s *JointSeparation) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	gravity := s.worldID.Gravity()
	if imgui.SliderFloatV("gravity", &gravity.Y, -500.0, 500.0, "%.0f", 0) {
		s.worldID.SetGravity(gravity)
	}

	imgui.SliderFloatV("magnitude", &s.impulse, 0.0, 1000.0, "%.0f", 0)

	if imgui.SliderFloatV("hertz", &s.jointHertz, 15.0, 120.0, "%.0f", 0) {
		for i := range jointSeparationCount {
			s.jointIDs[i].SetConstraintTuning(s.jointHertz, s.jointDampingRatio)
		}
	}

	if imgui.SliderFloatV("damping", &s.jointDampingRatio, 0.0, 10.0, "%.1f", 0) {
		for i := range jointSeparationCount {
			s.jointIDs[i].SetConstraintTuning(s.jointHertz, s.jointDampingRatio)
		}
	}

	imgui.PopItemWidth()

	if imgui.Button("impulse") {
		for i := range jointSeparationCount {
			p := s.bodyIDs[i].WorldPoint(box2d.Vec2{X: 1.0, Y: 1.0})
			s.bodyIDs[i].ApplyLinearImpulse(box2d.Vec2{X: s.impulse, Y: -s.impulse}, p, true)
		}
	}

	return true
}

func (s *JointSeparation) Step() {
	for i := range jointSeparationCount {
		if s.jointIDs[i] == (box2d.Joint{}) {
			continue
		}

		linear := s.jointIDs[i].LinearSeparation()
		angular := s.jointIDs[i].AngularSeparation()
		localFrame := s.jointIDs[i].LocalFrameA()
		s.draw.DrawString(s.camera, localFrame.P.Pos(), box2d.ColorWhite, "%.2f m, %.1f deg", linear, 180.0*angular/box2d.Pi)
	}

	s.Sample.Step()
}

var (
	userConstraintHertz    float32 = 3.0
	userConstraintZeta     float32 = 0.7
	userConstraintMaxForce float32 = 1000.0
)

// This shows how you can implement a constraint outside of Box2D
type UserConstraint struct {
	Sample

	bodyID   box2d.Body
	impulses [2]float32
}

func CreateUserConstraint(context *SampleContext) Sampler {
	s := &UserConstraint{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 3.0, Y: -1.0}
		s.context.camera.zoom = 25.0 * 0.15
	}

	box := box2d.MakeBox(1.0, 0.5)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 20.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 1.0
	bodyDef.AngularDamping = 0.5
	bodyDef.LinearDamping = 0.2
	s.bodyID = s.worldID.CreateBody(&bodyDef)
	s.bodyID.CreatePolygonShape(&shapeDef, &box)

	s.impulses[0] = 0.0
	s.impulses[1] = 0.0
	return s
}

func (s *UserConstraint) Step() {
	s.Sample.Step()

	s.draw.DrawTransform(box2d.WorldTransformIdentity, 1.0)

	if s.context.pause {
		return
	}

	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}
	if timeStep == 0.0 {
		return
	}

	invTimeStep := s.context.hertz

	omega := 2.0 * box2d.Pi * userConstraintHertz
	sigma := 2.0*userConstraintZeta + timeStep*omega
	sv := timeStep * omega * sigma
	impulseCoefficient := 1.0 / (1.0 + sv)
	massCoefficient := sv * impulseCoefficient
	biasCoefficient := omega / sigma

	localAnchors := [2]box2d.Vec2{{X: 1.0, Y: -0.5}, {X: 1.0, Y: 0.5}}
	mass := s.bodyID.Mass()
	invMass := float32(0.0)
	if mass >= 0.0001 {
		invMass = 1.0 / mass
	}
	inertiaTensor := s.bodyID.RotationalInertia()
	invI := float32(0.0)
	if inertiaTensor >= 0.0001 {
		invI = 1.0 / inertiaTensor
	}

	vB := s.bodyID.LinearVelocity()
	omegaB := s.bodyID.AngularVelocity()
	pB := s.bodyID.WorldCenter()

	for i := range 2 {
		anchorA := box2d.Pos{X: 3.0, Y: 0.0}
		anchorB := s.bodyID.WorldPoint(localAnchors[i])

		deltaAnchor := anchorB.Sub(anchorA)

		slackLength := float32(1.0)
		length := deltaAnchor.Length()
		C := length - slackLength
		if C < 0.0 || length < 0.001 {
			s.draw.DrawLine(anchorA, anchorB, box2d.ColorLightCyan)
			s.impulses[i] = 0.0
			continue
		}

		s.draw.DrawLine(anchorA, anchorB, box2d.ColorViolet)
		axis := deltaAnchor.Normalize()

		rB := anchorB.Sub(pB)
		Jb := rB.Cross(axis)
		K := invMass + Jb*invI*Jb
		invK := float32(0.0)
		if K >= 0.0001 {
			invK = 1.0 / K
		}

		Cdot := vB.Dot(axis) + Jb*omegaB
		impulse := -massCoefficient * invK * (Cdot + biasCoefficient*C)
		appliedImpulse := clampFloat(impulse, -userConstraintMaxForce*timeStep, 0.0)

		vB = vB.MulAdd(invMass*appliedImpulse, axis)
		omegaB += appliedImpulse * invI * Jb

		s.impulses[i] = appliedImpulse
	}

	s.bodyID.SetLinearVelocity(vB)
	s.bodyID.SetAngularVelocity(omegaB)

	s.DrawScreenTextLine("forces = %.6g, %.6g", s.impulses[0]*invTimeStep, s.impulses[1]*invTimeStep)
}

// This is a fun demo that shows off the wheel joint
type Driving struct {
	Sample

	car Car

	throttle     float32
	hertz        float32
	dampingRatio float32
	torque       float32
	speed        float32
}

func CreateDriving(context *SampleContext) Sampler {
	s := &Driving{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center.Y = 5.0
		s.context.camera.zoom = 25.0 * 0.4
		s.context.debugDraw.DrawJoints = false
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)

		var points [25]box2d.Vec2
		count := 24

		// fill in reverse to match line list convention
		points[count] = box2d.Vec2{X: -20.0, Y: -20.0}
		count--
		points[count] = box2d.Vec2{X: -20.0, Y: 0.0}
		count--
		points[count] = box2d.Vec2{X: 20.0, Y: 0.0}
		count--

		hs := [10]float32{0.25, 1.0, 4.0, 0.0, 0.0, -1.0, -2.0, -2.0, -1.25, 0.0}
		x, dx := float32(20.0), float32(5.0)

		for range 2 {
			for i := range 10 {
				y2 := hs[i]
				points[count] = box2d.Vec2{X: x + dx, Y: y2}
				count--
				x += dx
			}
		}

		// flat before bridge
		points[count] = box2d.Vec2{X: x + 40.0, Y: 0.0}
		count--
		points[count] = box2d.Vec2{X: x + 40.0, Y: -20.0}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.IsLoop = true
		groundID.CreateChain(&chainDef)

		// flat after bridge
		x += 80.0
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: x, Y: 0.0}, Point2: box2d.Vec2{X: x + 40.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		// jump ramp
		x += 40.0
		segment = box2d.Segment{Point1: box2d.Vec2{X: x, Y: 0.0}, Point2: box2d.Vec2{X: x + 10.0, Y: 5.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		// final corner
		x += 20.0
		segment = box2d.Segment{Point1: box2d.Vec2{X: x, Y: 0.0}, Point2: box2d.Vec2{X: x + 40.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		x += 40.0
		segment = box2d.Segment{Point1: box2d.Vec2{X: x, Y: 0.0}, Point2: box2d.Vec2{X: x, Y: 20.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Teeter
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 140.0, Y: 1.0}
		bodyDef.AngularVelocity = 1.0
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(10.0, 0.25)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		pivot := bodyDef.Position
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = bodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.LowerAngle = -8.0 * box2d.Pi / 180.0
		jointDef.UpperAngle = 8.0 * box2d.Pi / 180.0
		jointDef.EnableLimit = true
		s.worldID.CreateRevoluteJoint(&jointDef)
	}

	// Bridge
	{
		N := 20
		shapeDef := box2d.DefaultShapeDef()
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -1.0, Y: 0.0}, Center2: box2d.Vec2{X: 1.0, Y: 0.0}, Radius: 0.125}

		jointDef := box2d.DefaultRevoluteJointDef()

		prevBodyID := groundID
		for i := range N {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(161.0 + 2.0*float32(i)), Y: -0.125}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreateCapsuleShape(&shapeDef, &capsule)

			pivot := box2d.Pos{X: posScalar(160.0 + 2.0*float32(i)), Y: -0.125}
			jointDef.Base.BodyA = prevBodyID
			jointDef.Base.BodyB = bodyID
			jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
			jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
			s.worldID.CreateRevoluteJoint(&jointDef)

			prevBodyID = bodyID
		}

		pivot := box2d.Pos{X: posScalar(160.0 + 2.0*float32(N)), Y: -0.125}
		jointDef.Base.BodyA = prevBodyID
		jointDef.Base.BodyB = groundID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableMotor = true
		jointDef.MaxMotorTorque = 50.0
		s.worldID.CreateRevoluteJoint(&jointDef)
	}

	// Boxes
	{
		box := box2d.MakeBox(0.5, 0.5)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.25
		shapeDef.Material.Restitution = 0.25
		shapeDef.Density = 0.25

		for _, y := range []posScalar{0.5, 1.5, 2.5, 3.5, 4.5} {
			bodyDef.Position = box2d.Pos{X: 230.0, Y: y}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	// Car

	s.throttle = 0.0
	s.speed = 35.0
	s.torque = 5.0
	s.hertz = 5.0
	s.dampingRatio = 0.7

	s.car.Spawn(s.worldID, box2d.Pos{X: 0.0, Y: 0.0}, 1.0, s.hertz, s.dampingRatio, s.torque, nil)
	return s
}

func (s *Driving) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	if imgui.SliderFloatV("Spring Hertz", &s.hertz, 0.0, 20.0, "%.0f", 0) {
		s.car.SetHertz(s.hertz)
	}

	if imgui.SliderFloatV("Damping Ratio", &s.dampingRatio, 0.0, 10.0, "%.1f", 0) {
		s.car.SetDampingRadio(s.dampingRatio)
	}

	if imgui.SliderFloatV("Speed", &s.speed, 0.0, 50.0, "%.0f", 0) {
		s.car.SetSpeed(s.throttle * s.speed)
	}

	if imgui.SliderFloatV("Torque", &s.torque, 0.0, 10.0, "%.1f", 0) {
		s.car.SetTorque(s.torque)
	}
	imgui.PopItemWidth()

	return true
}

func (s *Driving) Step() {
	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.throttle = 1.0
		s.car.SetSpeed(s.speed)
	}

	if s.context.window.GetKey(glfw.KeyS) == glfw.Press {
		s.throttle = 0.0
		s.car.SetSpeed(0.0)
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.throttle = -1.0
		s.car.SetSpeed(-s.speed)
	}

	s.DrawScreenTextLine("Keys: left = a, brake = s, right = d")

	linearVelocity := s.car.chassisID.LinearVelocity()
	kph := linearVelocity.X * 3.6
	s.DrawScreenTextLine("speed in kph: %.2g", kph)

	carPosition := s.car.chassisID.Position()
	s.context.camera.center.X = carPosition.X

	s.Sample.Step()
}

type Ragdoll struct {
	Sample

	human               scenes.Human
	jointFrictionTorque float32
	jointHertz          float32
	jointDampingRatio   float32
}

func CreateRagdoll(context *SampleContext) Sampler {
	s := &Ragdoll{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 12.0}
		s.context.camera.zoom = 16.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.jointFrictionTorque = 0.03
	s.jointHertz = 5.0
	s.jointDampingRatio = 0.5

	s.human = scenes.Human{}

	s.Spawn()

	s.worldID.SetContactTuning(240.0, 0.0, 2.0)
	return s
}

func (s *Ragdoll) Spawn() {
	scenes.CreateHuman(&s.human, s.worldID, box2d.Pos{X: 0.0, Y: 25.0}, 1.0, s.jointFrictionTorque, s.jointHertz, s.jointDampingRatio, 1,
		nil, false)
}

func (s *Ragdoll) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Friction", &s.jointFrictionTorque, 0.0, 1.0, "%3.2f", 0) {
		s.human.SetJointFrictionTorque(s.jointFrictionTorque)
	}

	if imgui.SliderFloatV("Hertz", &s.jointHertz, 0.0, 10.0, "%3.1f", 0) {
		s.human.SetJointSpringHertz(s.jointHertz)
	}

	if imgui.SliderFloatV("Damping", &s.jointDampingRatio, 0.0, 4.0, "%3.1f", 0) {
		s.human.SetJointDampingRatio(s.jointDampingRatio)
	}

	imgui.PopItemWidth()

	if imgui.Button("Respawn") {
		scenes.DestroyHuman(&s.human)
		s.Spawn()
	}

	return true
}

type SoftBody struct {
	Sample

	donut Donut
}

func CreateSoftBody(context *SampleContext) Sampler {
	s := &SoftBody{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 25.0 * 0.25
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.donut.Create(s.worldID, box2d.Pos{X: 0.0, Y: 10.0}, 2.0, 0, false, nil)
	return s
}

type DoohickeyFarm struct {
	Sample
}

func CreateDoohickeyFarm(context *SampleContext) Sampler {
	s := &DoohickeyFarm{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 25.0 * 0.35
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(1.0, 1.0, box2d.Vec2{X: 0.0, Y: 1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	y := float32(4.0)
	for range 4 {
		var doohickey Doohickey
		doohickey.Spawn(s.worldID, box2d.Pos{X: 0.0, Y: posScalar(y)}, 0.5)
		y += 2.0
	}
	return s
}

type ScissorLift struct {
	Sample

	liftJointID box2d.DistanceJoint
	motorForce  float32
	motorSpeed  float32
	enableMotor bool
}

func CreateScissorLift(context *SampleContext) Sampler {
	s := &ScissorLift{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 9.0}
		s.context.camera.zoom = 25.0 * 0.4
	}

	// Need 8 sub-steps for smoother operation
	s.context.subStepCount = 8

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.SleepThreshold = 0.01

	shapeDef := box2d.DefaultShapeDef()
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -2.5, Y: 0.0}, Center2: box2d.Vec2{X: 2.5, Y: 0.0}, Radius: 0.15}

	baseID1 := groundID
	baseID2 := groundID
	baseAnchor1 := box2d.Vec2{X: -2.5, Y: 0.2}
	baseAnchor2 := box2d.Vec2{X: 2.5, Y: 0.2}
	y := float32(0.5)

	var linkID1 box2d.Body
	N := 3

	constraintDampingRatio := float32(20.0)
	constraintHertz := float32(240.0)

	for i := range N {
		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(y)}
		bodyDef.Rotation = box2d.MakeRot(0.15)
		bodyID1 := s.worldID.CreateBody(&bodyDef)
		bodyID1.CreateCapsuleShape(&shapeDef, &capsule)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(y)}
		bodyDef.Rotation = box2d.MakeRot(-0.15)

		bodyID2 := s.worldID.CreateBody(&bodyDef)
		bodyID2.CreateCapsuleShape(&shapeDef, &capsule)

		if i == 1 {
			linkID1 = bodyID2
		}

		revoluteDef := box2d.DefaultRevoluteJointDef()

		// left pin
		revoluteDef.Base.BodyA = baseID1
		revoluteDef.Base.BodyB = bodyID1
		revoluteDef.Base.LocalFrameA.P = baseAnchor1
		revoluteDef.Base.LocalFrameB.P = box2d.Vec2{X: -2.5, Y: 0.0}
		revoluteDef.Base.CollideConnected = i == 0
		revoluteDef.Base.ConstraintDampingRatio = constraintDampingRatio
		revoluteDef.Base.ConstraintHertz = constraintHertz

		s.worldID.CreateRevoluteJoint(&revoluteDef)

		// right pin
		if i == 0 {
			wheelDef := box2d.DefaultWheelJointDef()
			wheelDef.Base.BodyA = baseID2
			wheelDef.Base.BodyB = bodyID2
			wheelDef.Base.LocalFrameA.P = baseAnchor2
			wheelDef.Base.LocalFrameB.P = box2d.Vec2{X: 2.5, Y: 0.0}
			wheelDef.EnableSpring = false
			wheelDef.Base.CollideConnected = true
			wheelDef.Base.ConstraintDampingRatio = constraintDampingRatio
			wheelDef.Base.ConstraintHertz = constraintHertz

			s.worldID.CreateWheelJoint(&wheelDef)
		} else {
			revoluteDef.Base.BodyA = baseID2
			revoluteDef.Base.BodyB = bodyID2
			revoluteDef.Base.LocalFrameA.P = baseAnchor2
			revoluteDef.Base.LocalFrameB.P = box2d.Vec2{X: 2.5, Y: 0.0}
			revoluteDef.Base.CollideConnected = false
			revoluteDef.Base.ConstraintDampingRatio = constraintDampingRatio
			revoluteDef.Base.ConstraintHertz = constraintHertz

			s.worldID.CreateRevoluteJoint(&revoluteDef)
		}

		// middle pin
		revoluteDef.Base.BodyA = bodyID1
		revoluteDef.Base.BodyB = bodyID2
		revoluteDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.0}
		revoluteDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 0.0}
		revoluteDef.Base.CollideConnected = false
		revoluteDef.Base.ConstraintDampingRatio = constraintDampingRatio
		revoluteDef.Base.ConstraintHertz = constraintHertz

		s.worldID.CreateRevoluteJoint(&revoluteDef)

		baseID1 = bodyID2
		baseID2 = bodyID1
		baseAnchor1 = box2d.Vec2{X: -2.5, Y: 0.0}
		baseAnchor2 = box2d.Vec2{X: 2.5, Y: 0.0}
		y += 1.0
	}

	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(y)}
	bodyDef.Rotation = box2d.RotIdentity
	platformID := s.worldID.CreateBody(&bodyDef)

	box := box2d.MakeBox(3.0, 0.2)
	platformID.CreatePolygonShape(&shapeDef, &box)

	// left pin
	revoluteDef := box2d.DefaultRevoluteJointDef()
	revoluteDef.Base.BodyA = platformID
	revoluteDef.Base.BodyB = baseID1
	revoluteDef.Base.LocalFrameA.P = box2d.Vec2{X: -2.5, Y: -0.4}
	revoluteDef.Base.LocalFrameB.P = baseAnchor1
	revoluteDef.Base.CollideConnected = true
	revoluteDef.Base.ConstraintDampingRatio = constraintDampingRatio
	revoluteDef.Base.ConstraintHertz = constraintHertz
	s.worldID.CreateRevoluteJoint(&revoluteDef)

	// right pin
	wheelDef := box2d.DefaultWheelJointDef()
	wheelDef.Base.BodyA = platformID
	wheelDef.Base.BodyB = baseID2
	wheelDef.Base.LocalFrameA.P = box2d.Vec2{X: 2.5, Y: -0.4}
	wheelDef.Base.LocalFrameB.P = baseAnchor2
	wheelDef.EnableSpring = false
	wheelDef.Base.CollideConnected = true
	wheelDef.Base.ConstraintDampingRatio = constraintDampingRatio
	wheelDef.Base.ConstraintHertz = constraintHertz
	s.worldID.CreateWheelJoint(&wheelDef)

	s.enableMotor = false
	s.motorSpeed = 0.25
	s.motorForce = 2000.0

	distanceDef := box2d.DefaultDistanceJointDef()
	distanceDef.Base.BodyA = groundID
	distanceDef.Base.BodyB = linkID1
	distanceDef.Base.LocalFrameA.P = box2d.Vec2{X: -2.5, Y: 0.2}
	distanceDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.5, Y: 0.0}
	distanceDef.EnableSpring = true
	distanceDef.MinLength = 0.2
	distanceDef.MaxLength = 5.5
	distanceDef.EnableLimit = true
	distanceDef.EnableMotor = s.enableMotor
	distanceDef.MotorSpeed = s.motorSpeed
	distanceDef.MaxMotorForce = s.motorForce
	s.liftJointID = s.worldID.CreateDistanceJoint(&distanceDef)

	var car Car
	car.Spawn(s.worldID, box2d.Pos{X: 0.0, Y: posScalar(y + 2.0)}, 1.0, 3.0, 0.7, 0.0, nil)
	return s
}

func (s *ScissorLift) DrawControls() bool {
	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.liftJointID.SetMotorEnabled(s.enableMotor)
		s.liftJointID.WakeBodies()
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Max Force", &s.motorForce, 0.0, 3000.0, "%.0f", 0) {
		s.liftJointID.SetMaxMotorForce(s.motorForce)
		s.liftJointID.WakeBodies()
	}

	if imgui.SliderFloatV("Speed", &s.motorSpeed, -0.3, 0.3, "%.2f", 0) {
		s.liftJointID.SetMotorSpeed(s.motorSpeed)
		s.liftJointID.WakeBodies()
	}

	imgui.PopItemWidth()

	return true
}

const gearLiftPath = "m 63.500002,201.08333 103.187498,0 1e-5,-37.04166 h -2.64584 l 0,34.39583 h -42.33333 v -2.64583 l " +
	"-2.64584,-1e-5 v -2.64583 h -2.64583 v -2.64584 h -2.64584 v -2.64583 H 111.125 v -2.64583 h -2.64583 v " +
	"-2.64583 h -2.64583 v -2.64584 l -2.64584,1e-5 v -2.64583 l -2.64583,-1e-5 V 174.625 h -2.645834 v -2.64584 l " +
	"-2.645833,1e-5 v -2.64584 H 92.60417 v -2.64583 h -2.645834 v -2.64583 l -26.458334,0 0,37.04166"

type GearLift struct {
	Sample

	driverID    box2d.RevoluteJoint
	motorTorque float32
	motorSpeed  float32
	enableMotor bool
}

func createGear(bodyID box2d.Body, gearRadius, toothOffset, toothHalfWidth, toothHalfHeight, toothRadius float32) {
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.1
	shapeDef.Material.CustomColor = uint32(box2d.ColorSaddleBrown)
	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: gearRadius}
	bodyID.CreateCircleShape(&shapeDef, &circle)

	count := 16
	deltaAngle := 2.0 * box2d.Pi / 16
	dq := box2d.MakeRot(deltaAngle)
	center := box2d.Vec2{X: gearRadius + toothOffset, Y: 0.0}
	rotation := box2d.RotIdentity

	for range count {
		tooth := box2d.MakeOffsetRoundedBox(toothHalfWidth, toothHalfHeight, center, rotation, toothRadius)
		shapeDef.Material.CustomColor = uint32(box2d.ColorGray)
		bodyID.CreatePolygonShape(&shapeDef, &tooth)

		rotation = dq.Mul(rotation)
		center = rotation.Rotate(box2d.Vec2{X: gearRadius + toothOffset, Y: 0.0})
	}
}

func CreateGearLift(context *SampleContext) Sampler {
	s := &GearLift{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 6.0}
		s.context.camera.zoom = 7.0
		s.context.debugDraw.DrawJoints = false
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
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

	gearPosition1 := box2d.Pos{X: -4.25, Y: 9.75}
	gearPosition2 := gearPosition1.Add(box2d.Vec2{X: 2.0, Y: 1.0})
	linkAttachPosition := gearPosition2.Add(box2d.Vec2{X: gearRadius + 2.0*toothHalfWidth + toothRadius, Y: 0.0})
	doorPosition := linkAttachPosition.Add(box2d.Vec2{X: 0.0, Y: 2.0*linkCount*linkHalfLength + doorHalfHeight}.Neg())

	{
		position := gearPosition1
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = position

		bodyID := s.worldID.CreateBody(&bodyDef)

		createGear(bodyID, gearRadius, toothHalfHeight, toothHalfWidth, toothHalfHeight, toothRadius)

		revoluteDef := box2d.DefaultRevoluteJointDef()

		s.motorTorque = 80.0
		s.motorSpeed = 0.0
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

		createGear(followerID, gearRadius, toothHalfWidth, toothHalfWidth, toothHalfHeight, toothRadius)

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
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -linkHalfLength}, Center2: box2d.Vec2{X: 0.0, Y: linkHalfLength}, Radius: linkRadius}

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
			box2d.ColorGray, box2d.ColorGainsboro, box2d.ColorLightGray, box2d.ColorLightSlateGray, box2d.ColorDarkGray,
		}

		y := float32(4.25)
		xCount, yCount := 10, 20
		for range yCount {
			x := float32(-3.15)
			for range xCount {
				bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
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

func (s *GearLift) DrawControls() bool {
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

func (s *GearLift) Step() {
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
}

// A top down door
type Door struct {
	Sample

	doorID            box2d.Body
	jointID           box2d.RevoluteJoint
	impulse           float32
	translationError  float32
	jointHertz        float32
	jointDampingRatio float32
	enableLimit       bool
}

func CreateDoor(context *SampleContext) Sampler {
	s := &Door{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 4.0
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.enableLimit = true
	s.impulse = 50000.0
	s.translationError = 0.0
	s.jointHertz = 240.0
	s.jointDampingRatio = 1.0

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.5}
		bodyDef.GravityScale = 0.0

		s.doorID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1000.0

		box := box2d.MakeBox(0.1, 1.5)
		s.doorID.CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = groundID
		jointDef.Base.BodyB = s.doorID
		jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.0}
		jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: -1.5}
		jointDef.Base.ConstraintHertz = s.jointHertz
		jointDef.Base.ConstraintDampingRatio = s.jointDampingRatio
		jointDef.TargetAngle = 0.0
		jointDef.EnableSpring = true
		jointDef.Hertz = 1.0
		jointDef.DampingRatio = 0.5
		jointDef.MotorSpeed = 0.0
		jointDef.MaxMotorTorque = 0.0
		jointDef.EnableMotor = false
		jointDef.LowerAngle = -0.5 * box2d.Pi
		jointDef.UpperAngle = 0.5 * box2d.Pi
		jointDef.EnableLimit = s.enableLimit

		s.jointID = s.worldID.CreateRevoluteJoint(&jointDef)
	}
	return s
}

func (s *Door) DrawControls() bool {
	if imgui.Button("impulse") {
		p := s.doorID.WorldPoint(box2d.Vec2{X: 0.0, Y: 1.5})
		s.doorID.ApplyLinearImpulse(box2d.Vec2{X: s.impulse, Y: 0.0}, p, true)
		s.translationError = 0.0
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())

	imgui.SliderFloatV("magnitude", &s.impulse, 1000.0, 100000.0, "%.0f", 0)

	if imgui.SliderFloatV("hertz", &s.jointHertz, 15.0, 480.0, "%.0f", 0) {
		s.jointID.SetConstraintTuning(s.jointHertz, s.jointDampingRatio)
	}

	if imgui.SliderFloatV("damping", &s.jointDampingRatio, 0.0, 10.0, "%.1f", 0) {
		s.jointID.SetConstraintTuning(s.jointHertz, s.jointDampingRatio)
	}

	imgui.PopItemWidth()

	if imgui.Checkbox("limit", &s.enableLimit) {
		s.jointID.SetLimitEnabled(s.enableLimit)
	}

	return true
}

func (s *Door) Step() {
	s.Sample.Step()

	p := s.doorID.WorldPoint(box2d.Vec2{X: 0.0, Y: 1.5})
	s.draw.DrawPoint(p, 5.0, box2d.ColorDarkKhaki)

	translationError := s.jointID.LinearSeparation()
	s.translationError = max(s.translationError, translationError)

	s.DrawScreenTextLine("translation error = %.6g", s.translationError)
}

type ScaleRagdoll struct {
	Sample

	human scenes.Human
	scale float32
}

func CreateScaleRagdoll(context *SampleContext) Sampler {
	s := &ScaleRagdoll{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 4.5}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		box := box2d.MakeOffsetBox(20.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.scale = 1.0

	s.human = scenes.Human{}

	s.Spawn()
	return s
}

func (s *ScaleRagdoll) Spawn() {
	jointFrictionTorque := float32(0.03)
	jointHertz := float32(1.0)
	jointDampingRatio := float32(0.5)
	scenes.CreateHuman(&s.human, s.worldID, box2d.Pos{X: 0.0, Y: 5.0}, s.scale, jointFrictionTorque, jointHertz, jointDampingRatio, 1, nil,
		false)
	s.human.ApplyRandomAngularImpulse(0.1)
}

func (s *ScaleRagdoll) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Scale", &s.scale, 0.1, 10.0, "%3.2f", imgui.SliderFlagsClampOnInput) {
		s.human.SetScale(s.scale)
	}

	imgui.PopItemWidth()

	return true
}

// Theo Jansen's walking machine, inspired by a contribution from roman_m
// Dimensions scooped from APE (http://www.cove.org/ape/index.htm)
type TheoJansen struct {
	Sample

	offset       box2d.Vec2
	chassisID    box2d.Body
	wheelID      box2d.Body
	motorJointID box2d.RevoluteJoint
	motorSpeed   float32
	enableMotor  bool
}

func CreateTheoJansen(context *SampleContext) Sampler {
	s := &TheoJansen{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 6.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	s.offset = box2d.Vec2{X: 0.0, Y: 8.0}
	s.motorSpeed = 2.0
	s.enableMotor = true
	pivot := box2d.Vec2{X: 0.0, Y: 0.8}

	// Ground
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		segment := box2d.Segment{Point1: box2d.Vec2{X: -50.0, Y: 0.0}, Point2: box2d.Vec2{X: 50.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: -50.0, Y: 0.0}, Point2: box2d.Vec2{X: -50.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: 50.0, Y: 0.0}, Point2: box2d.Vec2{X: 50.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Balls
	for i := range 40 {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.25}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-40.0 + 2.0*float32(i)), Y: 0.5}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	// Chassis
	{
		box := box2d.MakeBox(2.5, 1.0)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = pivot.Pos().Add(s.offset)
		s.chassisID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		shapeDef.Filter.GroupIndex = -1
		s.chassisID.CreatePolygonShape(&shapeDef, &box)
	}

	// Wheel
	{
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 1.6}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = pivot.Pos().Add(s.offset)
		s.wheelID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 1.0
		shapeDef.Filter.GroupIndex = -1
		s.wheelID.CreateCircleShape(&shapeDef, &circle)
	}

	{
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = s.wheelID
		jointDef.Base.BodyB = s.chassisID

		motorPivot := pivot.Pos().Add(s.offset)
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(motorPivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(motorPivot)

		jointDef.EnableMotor = s.enableMotor
		jointDef.MotorSpeed = s.motorSpeed
		jointDef.MaxMotorTorque = 400.0
		s.motorJointID = s.worldID.CreateRevoluteJoint(&jointDef)
	}

	wheelAnchor := pivot.Add(box2d.Vec2{X: 0.0, Y: -0.8})

	s.CreateLeg(-1.0, wheelAnchor)
	s.CreateLeg(1.0, wheelAnchor)

	// Stagger three leg pairs around the wheel so some feet always touch the ground
	s.wheelID.SetTransform(s.wheelID.Position(), box2d.MakeRot(120.0*box2d.Pi/180.0))
	s.CreateLeg(-1.0, wheelAnchor)
	s.CreateLeg(1.0, wheelAnchor)

	s.wheelID.SetTransform(s.wheelID.Position(), box2d.MakeRot(-120.0*box2d.Pi/180.0))
	s.CreateLeg(-1.0, wheelAnchor)
	s.CreateLeg(1.0, wheelAnchor)
	return s
}

func (s *TheoJansen) CreateLeg(sign float32, wheelAnchor box2d.Vec2) {
	p1 := box2d.Vec2{X: 5.4 * sign, Y: -6.1}
	p2 := box2d.Vec2{X: 7.2 * sign, Y: -1.2}
	p3 := box2d.Vec2{X: 4.3 * sign, Y: -1.9}
	p4 := box2d.Vec2{X: 3.1 * sign, Y: 0.8}
	p5 := box2d.Vec2{X: 6.0 * sign, Y: 1.5}
	p6 := box2d.Vec2{X: 2.5 * sign, Y: 3.7}

	// The hull computation handles the mirrored winding for s < 0
	vertices1 := []box2d.Vec2{p1, p2, p3}
	vertices2 := []box2d.Vec2{box2d.Vec2Zero, p5.Sub(p4), p6.Sub(p4)}

	hull1 := box2d.ComputeHull(vertices1)
	hull2 := box2d.ComputeHull(vertices2)
	poly1 := box2d.MakePolygon(&hull1, 0.0)
	poly2 := box2d.MakePolygon(&hull2, 0.0)

	bodyDef1 := box2d.DefaultBodyDef()
	bodyDef1.Type = box2d.DynamicBody
	bodyDef1.Position = s.offset.Pos()
	bodyDef1.AngularDamping = 10.0
	body1 := s.worldID.CreateBody(&bodyDef1)

	bodyDef2 := box2d.DefaultBodyDef()
	bodyDef2.Type = box2d.DynamicBody
	bodyDef2.Position = p4.Pos().Add(s.offset)
	bodyDef2.AngularDamping = 10.0
	body2 := s.worldID.CreateBody(&bodyDef2)

	// Negative group index keeps the legs from colliding with the chassis and wheel
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Filter.GroupIndex = -1
	body1.CreatePolygonShape(&shapeDef, &poly1)
	body2.CreatePolygonShape(&shapeDef, &poly2)

	{
		// Soft distance constraints reduce jitter and act like a suspension system
		jointDef := box2d.DefaultDistanceJointDef()
		jointDef.EnableSpring = true
		jointDef.Hertz = 10.0
		jointDef.DampingRatio = 0.5

		jointDef.Base.BodyA = body1
		jointDef.Base.BodyB = body2

		anchorA := p2.Pos().Add(s.offset)
		anchorB := p5.Pos().Add(s.offset)
		jointDef.Base.LocalFrameA.P = body1.LocalPoint(anchorA)
		jointDef.Base.LocalFrameB.P = body2.LocalPoint(anchorB)
		jointDef.Length = anchorA.Sub(anchorB).Length()
		s.worldID.CreateDistanceJoint(&jointDef)

		anchorA = p3.Pos().Add(s.offset)
		anchorB = p4.Pos().Add(s.offset)
		jointDef.Base.LocalFrameA.P = body1.LocalPoint(anchorA)
		jointDef.Base.LocalFrameB.P = body2.LocalPoint(anchorB)
		jointDef.Length = anchorA.Sub(anchorB).Length()
		s.worldID.CreateDistanceJoint(&jointDef)

		jointDef.Base.BodyB = s.wheelID
		anchorB = wheelAnchor.Pos().Add(s.offset)
		jointDef.Base.LocalFrameB.P = s.wheelID.LocalPoint(anchorB)
		jointDef.Length = anchorA.Sub(anchorB).Length()
		s.worldID.CreateDistanceJoint(&jointDef)

		jointDef.Base.BodyA = body2
		anchorA = p6.Pos().Add(s.offset)
		jointDef.Base.LocalFrameA.P = body2.LocalPoint(anchorA)
		jointDef.Length = anchorA.Sub(anchorB).Length()
		s.worldID.CreateDistanceJoint(&jointDef)
	}

	{
		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = body2
		jointDef.Base.BodyB = s.chassisID

		legPivot := p4.Pos().Add(s.offset)
		jointDef.Base.LocalFrameA.P = body2.LocalPoint(legPivot)
		jointDef.Base.LocalFrameB.P = s.chassisID.LocalPoint(legPivot)
		s.worldID.CreateRevoluteJoint(&jointDef)
	}
}

func (s *TheoJansen) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.Checkbox("Motor", &s.enableMotor) {
		s.motorJointID.SetMotorEnabled(s.enableMotor)
	}

	if imgui.SliderFloatV("Speed", &s.motorSpeed, -5.0, 5.0, "%.1f", 0) {
		s.motorJointID.SetMotorSpeed(s.motorSpeed)
	}

	imgui.PopItemWidth()

	return true
}

func (s *TheoJansen) Step() {
	s.DrawScreenTextLine("Keys: left = a, brake = s, right = d, toggle motor = f")

	s.Sample.Step()
}

func (s *TheoJansen) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyA:
		s.motorJointID.SetMotorSpeed(-s.motorSpeed)

	case glfw.KeyS:
		s.motorJointID.SetMotorSpeed(0.0)

	case glfw.KeyD:
		s.motorJointID.SetMotorSpeed(s.motorSpeed)

	case glfw.KeyF:
		s.enableMotor = !s.motorJointID.MotorEnabled()
		s.motorJointID.SetMotorEnabled(s.enableMotor)
	}
}

const deskLampSpringCount = 2

// The Anglepoise Original 1227 desk lamp.
// https://www.anglepoise.com/usa/product-category/original-1227/
type DeskLamp struct {
	Sample

	springJointIDs [deskLampSpringCount]box2d.DistanceJoint
	postJointID    box2d.RevoluteJoint
	elbowJointID   box2d.RevoluteJoint
	headJointID    box2d.RevoluteJoint
	topLinkID      box2d.Body
	springHertz    [deskLampSpringCount]float32
	springDamping  float32
	lowerFriction  float32
	elbowFriction  float32
	headAngle      float32
}

func CreateDeskLamp(context *SampleContext) Sampler {
	s := &DeskLamp{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.2, Y: 1.42}
		s.context.camera.zoom = 1.6
	}

	// The parallelogram is narrow, so it wants the extra resolution
	s.context.subStepCount = 8

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Dimensions taken off an Original 1227 in meters, then scaled up. At life size the silver
	// link is two centimeters, which is far too close to the linear slop.
	scale := float32(5.0)

	baseHalfWidth := 0.070 * scale
	baseHalfHeight := 0.018 * scale
	postHalfWidth := 0.008 * scale
	pivotHeight := 0.185 * scale
	springAnchorHeight := 0.095 * scale
	topLinkLength := 0.296 * scale
	silverFoot := 0.024 * scale
	topLeverLength := 0.019 * scale
	lowerLeverLength := 0.028 * scale
	upperArmLength := 0.324 * scale
	tubeRadius := 0.0036 * scale
	shadeLength := 0.161 * scale
	shadeNeckRadius := 0.020 * scale
	shadeMouthRadius := 0.061 * scale

	// The silver link measured in the top link frame. The same span separates the two pins on the
	// upper arm, which is what closes the parallelogram.
	silverSpan := box2d.Vec2{X: 0.0212 * scale, Y: 0.0109 * scale}
	silverLength := silverSpan.Length()

	topLinkAngle := float32(-0.878)
	upperArmAngle := float32(1.192)
	shadeTilt := float32(-0.41)

	// The arm spring balances the elevation, the head spring balances the upper arm through the
	// parallelogram. That is what the real lamp uses two springs for.
	s.springHertz[0] = 5.0
	s.springHertz[1] = 5.0
	s.springDamping = 4.0
	s.lowerFriction = 100.0
	s.elbowFriction = 100.0
	s.headAngle = 0.0
	topRotation := box2d.MakeRot(topLinkAngle)
	upperRotation := box2d.MakeRot(upperArmAngle)

	topAxis := topRotation.YAxis()
	silverWorld := topRotation.Rotate(silverSpan)

	postPivot := box2d.Pos{X: 0.0, Y: posScalar(pivotHeight)}
	upperPin := postPivot.Add(topAxis.Scale(topLinkLength))
	silverBase := postPivot.Add(topAxis.Scale(silverFoot))
	lowerFoot := silverBase.Add(silverWorld)
	lowerPin := upperPin.Add(silverWorld)

	// The lower link is parallel to the top link and shorter by the silver link foot
	lowerLinkLength := topLinkLength - silverFoot
	silverRotation := box2d.MakeRotFromUnitVector(silverWorld.Normalize().RightPerp())

	headPivot := upperPin.Add(upperRotation.YAxis().Scale(upperArmLength))

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()

	var baseID box2d.Body
	{
		// The base is static so the linkage is all that moves
		bodyDef.Type = box2d.StaticBody
		bodyDef.Position = box2d.Pos{}
		bodyDef.Rotation = box2d.RotIdentity
		baseID = s.worldID.CreateBody(&bodyDef)

		plinth := box2d.MakeOffsetBox(baseHalfWidth, baseHalfHeight, box2d.Vec2{X: 0.0, Y: baseHalfHeight}, box2d.RotIdentity)
		postHalfHeight := 0.5 * (pivotHeight - 2.0*baseHalfHeight)
		post := box2d.MakeOffsetBox(postHalfWidth, postHalfHeight, box2d.Vec2{X: 0.0, Y: pivotHeight - postHalfHeight}, box2d.RotIdentity)

		baseID.CreatePolygonShape(&shapeDef, &plinth)
		baseID.CreatePolygonShape(&shapeDef, &post)
	}

	bodyDef.Type = box2d.DynamicBody
	shapeDef.Density = 20.0

	// Both long links run past their feet into the levers the springs pull on
	{
		bodyDef.Position = postPivot
		bodyDef.Rotation = topRotation
		s.topLinkID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -topLeverLength}, Center2: box2d.Vec2{X: 0.0, Y: topLinkLength}, Radius: tubeRadius}
		s.topLinkID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	var lowerLinkID box2d.Body
	{
		bodyDef.Position = lowerFoot
		bodyDef.Rotation = topRotation
		lowerLinkID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -lowerLeverLength}, Center2: box2d.Vec2{X: 0.0, Y: lowerLinkLength}, Radius: tubeRadius}
		lowerLinkID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	var silverID box2d.Body
	{
		bodyDef.Position = silverBase
		bodyDef.Rotation = silverRotation
		silverID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: 0.0}, Center2: box2d.Vec2{X: 0.0, Y: silverLength}, Radius: 1.4 * tubeRadius}
		shapeDef.Material.CustomColor = uint32(box2d.ColorSilver)
		silverID.CreateCapsuleShape(&shapeDef, &capsule)
		shapeDef.Material.CustomColor = 0
	}

	var upperArmID box2d.Body
	{
		bodyDef.Position = upperPin
		bodyDef.Rotation = upperRotation
		upperArmID = s.worldID.CreateBody(&bodyDef)

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: 0.0}, Center2: box2d.Vec2{X: 0.0, Y: upperArmLength}, Radius: 0.8 * tubeRadius}
		upperArmID.CreateCapsuleShape(&shapeDef, &capsule)

		// The coupler span reaches back to the lower link pin
		bracket := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: 0.0}, Center2: upperArmID.LocalPoint(lowerPin), Radius: tubeRadius}
		upperArmID.CreateCapsuleShape(&shapeDef, &bracket)
	}

	pinDef := box2d.DefaultRevoluteJointDef()
	pinDef.Base.DrawScale = 0.03 * scale
	pinDef.Base.ConstraintHertz = 240.0
	pinDef.Base.ConstraintDampingRatio = 2.0

	// Only the top link reaches the post. The springs carry it, so it needs little friction.
	pinDef.Base.BodyA = baseID
	pinDef.Base.BodyB = s.topLinkID
	pinDef.Base.LocalFrameA.P = baseID.LocalPoint(postPivot)
	pinDef.Base.LocalFrameB.P = box2d.Vec2Zero
	pinDef.EnableMotor = true
	pinDef.MaxMotorTorque = s.lowerFriction
	s.postJointID = s.worldID.CreateRevoluteJoint(&pinDef)

	pinDef.EnableMotor = false

	// The silver link hangs off the top link and carries the foot of the lower link
	pinDef.Base.BodyA = s.topLinkID
	pinDef.Base.BodyB = silverID
	pinDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: silverFoot}
	pinDef.Base.LocalFrameB.P = box2d.Vec2Zero
	s.worldID.CreateRevoluteJoint(&pinDef)

	pinDef.Base.BodyA = silverID
	pinDef.Base.BodyB = lowerLinkID
	pinDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: silverLength}
	pinDef.Base.LocalFrameB.P = box2d.Vec2Zero
	s.worldID.CreateRevoluteJoint(&pinDef)

	// Both long links reach the upper arm, at points a silver link apart.
	pinDef.Base.BodyA = lowerLinkID
	pinDef.Base.BodyB = upperArmID
	pinDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: lowerLinkLength}
	pinDef.Base.LocalFrameB.P = upperArmID.LocalPoint(lowerPin)
	s.worldID.CreateRevoluteJoint(&pinDef)

	// No spring reaches the elbow, so it is a friction pivot like the real one.
	pinDef.Base.BodyA = s.topLinkID
	pinDef.Base.BodyB = upperArmID
	pinDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: topLinkLength}
	pinDef.Base.LocalFrameB.P = box2d.Vec2Zero
	pinDef.EnableMotor = true
	pinDef.MaxMotorTorque = s.elbowFriction
	s.elbowJointID = s.worldID.CreateRevoluteJoint(&pinDef)

	springDef := box2d.DefaultDistanceJointDef()
	springDef.Base.DrawScale = 0.03 * scale
	springDef.EnableSpring = true
	springDef.DampingRatio = s.springDamping

	// Carwardine's trick. With almost no free length the spring torque about the pivot is
	// k times cross(lever, anchor), which tracks the gravity torque as the arm swings, so the
	// arm balances over its whole range instead of at one angle.
	springDef.Length = 0.01

	springBase := box2d.Pos{X: 0.0, Y: posScalar(springAnchorHeight)}
	leverIDs := [deskLampSpringCount]box2d.Body{s.topLinkID, lowerLinkID}
	leverLengths := [deskLampSpringCount]float32{topLeverLength, lowerLeverLength}

	for k := range deskLampSpringCount {
		springDef.Base.BodyA = baseID
		springDef.Base.BodyB = leverIDs[k]
		springDef.Base.LocalFrameA.P = baseID.LocalPoint(springBase)
		springDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: -leverLengths[k]}
		springDef.Hertz = s.springHertz[k]
		s.springJointIDs[k] = s.worldID.CreateDistanceJoint(&springDef)
	}

	{
		bodyDef.Position = headPivot
		bodyDef.Rotation = box2d.MakeRot(shadeTilt)
		headID := s.worldID.CreateBody(&bodyDef)

		// The pivot is on the shade center line, where the arm capsule ends. The shade opens along
		// the local negative y.
		neck := 0.45 * shadeLength
		mouth := neck - shadeLength
		points := []box2d.Vec2{
			{X: -shadeNeckRadius, Y: neck},
			{X: shadeNeckRadius, Y: neck},
			{X: shadeMouthRadius, Y: mouth},
			{X: -shadeMouthRadius, Y: mouth},
		}
		hull := box2d.ComputeHull(points)
		shade := box2d.MakePolygon(&hull, 0.01*scale)

		// The shade is a thin shell.
		shapeDef.Density = 1.0
		headID.CreatePolygonShape(&shapeDef, &shade)

		// Offset the reference frame so the head slider reads zero at the pose above.
		pinDef.Base.BodyA = upperArmID
		pinDef.Base.BodyB = headID
		pinDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: upperArmLength}
		pinDef.Base.LocalFrameA.Q = upperRotation.InvMul(box2d.MakeRot(shadeTilt))
		pinDef.Base.LocalFrameB.P = box2d.Vec2Zero
		pinDef.EnableMotor = false
		pinDef.MaxMotorTorque = 10.0
		pinDef.EnableSpring = true
		pinDef.Hertz = 8.0
		pinDef.DampingRatio = 2.0
		pinDef.TargetAngle = s.headAngle
		pinDef.EnableLimit = true
		pinDef.LowerAngle = -1.2
		pinDef.UpperAngle = 1.2
		s.headJointID = s.worldID.CreateRevoluteJoint(&pinDef)
	}
	return s
}

func (s *DeskLamp) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Arm Spring", &s.springHertz[0], 0.0, 20.0, "%.2f", 0) {
		s.springJointIDs[0].SetSpringHertz(s.springHertz[0])
		s.springJointIDs[0].WakeBodies()
	}

	if imgui.SliderFloatV("Head Spring", &s.springHertz[1], 0.0, 20.0, "%.2f", 0) {
		s.springJointIDs[1].SetSpringHertz(s.springHertz[1])
		s.springJointIDs[1].WakeBodies()
	}

	if imgui.SliderFloatV("Spring Damping", &s.springDamping, 0.0, 5.0, "%.1f", 0) {
		for k := range deskLampSpringCount {
			s.springJointIDs[k].SetSpringDampingRatio(s.springDamping)
			s.springJointIDs[k].WakeBodies()
		}
	}

	if imgui.SliderFloatV("Post Friction", &s.lowerFriction, 0.0, 20.0, "%.1f", 0) {
		s.postJointID.SetMaxMotorTorque(s.lowerFriction)
		s.postJointID.WakeBodies()
	}

	if imgui.SliderFloatV("Elbow Friction", &s.elbowFriction, 0.0, 60.0, "%.0f", 0) {
		s.elbowJointID.SetMaxMotorTorque(s.elbowFriction)
		s.elbowJointID.WakeBodies()
	}

	if imgui.SliderFloatV("Head", &s.headAngle, -1.2, 1.2, "%.2f", 0) {
		s.headJointID.SetTargetAngle(s.headAngle)
		s.headJointID.WakeBodies()
	}

	imgui.PopItemWidth()

	return true
}

func (s *DeskLamp) Step() {
	s.Sample.Step()

	postTorque := s.postJointID.MotorTorque()
	elbowTorque := s.elbowJointID.MotorTorque()
	s.DrawScreenTextLine("friction carrying: post %.2f, elbow %.2f N m", postTorque, elbowTorque)

	spring1 := s.springJointIDs[0].SpringForce()
	spring2 := s.springJointIDs[1].SpringForce()
	s.DrawScreenTextLine("spring forces: %.2f, %.2f N", spring1, spring2)
}

func init() {
	RegisterSample("Joints", "Distance Joint", CreateDistanceJoint)
	RegisterSample("Joints", "Motor Joint", CreateMotorJoint)
	RegisterSample("Joints", "Top Down Friction", CreateTopDownFriction)
	RegisterSample("Joints", "Filter Joint", CreateFilterJoint)
	RegisterSample("Joints", "Revolute", CreateRevoluteJoint)
	RegisterSample("Joints", "Prismatic", CreatePrismaticJoint)
	RegisterSample("Joints", "Wheel", CreateWheelJoint)
	RegisterSample("Joints", "Bridge", CreateBridge)
	RegisterSample("Joints", "Ball & Chain", CreateBallAndChain)
	RegisterSample("Joints", "Cantilever", CreateCantilever)
	RegisterSample("Joints", "Motion Locks", CreateMotionLocks)
	RegisterSample("Joints", "Breakable", CreateBreakableJoint)
	RegisterSample("Joints", "Separation", CreateJointSeparation)
	RegisterSample("Joints", "User Constraint", CreateUserConstraint)
	RegisterSample("Joints", "Driving", CreateDriving)
	RegisterSample("Joints", "Ragdoll", CreateRagdoll)
	RegisterSample("Joints", "Soft Body", CreateSoftBody)
	RegisterSample("Joints", "Doohickey", CreateDoohickeyFarm)
	RegisterSample("Joints", "Scissor Lift", CreateScissorLift)
	RegisterSample("Joints", "Gear Lift", CreateGearLift)
	RegisterSample("Joints", "Door", CreateDoor)
	RegisterSample("Joints", "Scale Ragdoll", CreateScaleRagdoll)
	RegisterSample("Joints", "Theo Jansen", CreateTheoJansen)
	RegisterSample("Joints", "Desk Lamp", CreateDeskLamp)
}
