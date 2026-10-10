// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/scenes"
)

const (
	funnelDonut = 1
	funnelHuman = 2
	funnelCount = 32
)

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
	s := &SensorFunnel{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 1.333
	}

	s.context.debugDraw.DrawJoints = false

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		points := []box2d.Vec2{
			{X: -16.8672504, Y: 31.088623},
			{X: 16.8672485, Y: 31.088623},
			{X: 16.8672485, Y: 17.1978741},
			{X: 8.26824951, Y: 11.906374},
			{X: 16.8672485, Y: 11.906374},
			{X: 16.8672485, Y: -0.661376953},
			{X: 8.26824951, Y: -5.953125},
			{X: 16.8672485, Y: -5.953125},
			{X: 16.8672485, Y: -13.229126},
			{X: 3.63799858, Y: -23.151123},
			{X: 3.63799858, Y: -31.088623},
			{X: -3.63800049, Y: -31.088623},
			{X: -3.63800049, Y: -23.151123},
			{X: -16.8672504, Y: -13.229126},
			{X: -16.8672504, Y: -5.953125},
			{X: -8.26825142, Y: -5.953125},
			{X: -16.8672504, Y: -0.661376953},
			{X: -16.8672504, Y: 11.906374},
			{X: -8.26825142, Y: 11.906374},
			{X: -16.8672504, Y: 17.1978741},
		}

		var material box2d.SurfaceMaterial
		material.Friction = 0.2

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points
		chainDef.IsLoop = true
		chainDef.Materials = []box2d.SurfaceMaterial{material}
		groundID.CreateChain(&chainDef)

		sign := float32(1.0)
		y := float32(14.0)
		for range 3 {
			bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(y)}
			bodyDef.Type = box2d.DynamicBody

			bodyID := s.worldID.CreateBody(&bodyDef)

			box := box2d.MakeBox(6.0, 0.5)
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
			revoluteDef.MotorSpeed = 2.0 * sign
			revoluteDef.EnableMotor = true

			s.worldID.CreateRevoluteJoint(&revoluteDef)

			y -= 14.0
			sign = -sign
		}

		{
			box := box2d.MakeOffsetBox(4.0, 1.0, box2d.Vec2{X: 0.0, Y: -30.5}, box2d.RotIdentity)
			shapeDef := box2d.DefaultShapeDef()
			shapeDef.IsSensor = true
			shapeDef.EnableSensorEvents = true

			groundID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	s.wait = 0.5
	s.side = -15.0
	s.elemType = funnelHuman

	s.CreateElement()
	return s
}

func (s *SensorFunnel) CreateElement() {
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

	center := box2d.Pos{X: posScalar(s.side), Y: 29.5}

	if s.elemType == funnelDonut {
		donut := &s.donuts[index]
		donut.Create(s.worldID, center, 1.0, 0, true, donut)
	} else {
		human := &s.humans[index]
		scale := float32(2.0)
		jointFriction := float32(0.05)
		jointHertz := float32(6.0)
		jointDamping := float32(0.5)
		colorize := true
		scenes.CreateHuman(human, s.worldID, center, scale, jointFriction, jointHertz, jointDamping, index+1, human, colorize)
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

	// Discover rings that touch the bottom sensor
	var deferredDestruction [funnelCount]bool
	for event := range s.worldID.SensorBeginEvents() {
		bodyID := event.VisitorShape.Body()

		if s.elemType == funnelDonut {
			if donut, ok := bodyID.UserData().(*Donut); ok && donut != nil {
				for index := range funnelCount {
					if &s.donuts[index] == donut {
						// Defer destruction to avoid double destruction and event invalidation (orphaned shape ids)
						deferredDestruction[index] = true
					}
				}
			}
		} else {
			if human, ok := bodyID.UserData().(*scenes.Human); ok && human != nil {
				for index := range funnelCount {
					if &s.humans[index] == human {
						// Defer destruction to avoid double destruction and event invalidation (orphaned shape ids)
						deferredDestruction[index] = true
					}
				}
			}
		}
	}

	// Safely destroy rings that hit the bottom sensor
	for i := range funnelCount {
		if deferredDestruction[i] {
			s.DestroyElement(i)
		}
	}

	if s.context.hertz > 0.0 && !s.context.pause {
		s.wait -= 1.0 / s.context.hertz
		if s.wait < 0.0 {
			s.CreateElement()
			s.wait += 0.5
		}
	}
}

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

func CreateSensorBookend(context *SampleContext) Sampler {
	s := &SensorBookend{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 6.0}
		s.context.camera.zoom = 7.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		groundSegment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: -10.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 10.0}}
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
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 2.0, Y: 1.0}
	s.sensorBodyID2 = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	shapeDef.EnableSensorEvents = true

	box := box2d.MakeRoundedBox(0.5, 0.5, 0.5)
	s.sensorShapeID2 = s.sensorBodyID2.CreatePolygonShape(&shapeDef, &box)

	// Solid middle
	shapeDef.IsSensor = false
	shapeDef.EnableSensorEvents = false
	box = box2d.MakeSquare(0.5)
	s.sensorBodyID2.CreatePolygonShape(&shapeDef, &box)
}

func (s *SensorBookend) CreateVisitor() {
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: -4.0, Y: 1.0}
	bodyDef.Type = box2d.DynamicBody

	s.visitorBodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	s.visitorShapeID = s.visitorBodyID.CreateCircleShape(&shapeDef, &circle)
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

func (s *SensorBookend) DrawControls() bool {
	if s.visitorBodyID == (box2d.Body{}) {
		if imgui.Button("create visitor") {
			s.CreateVisitor()
		}
	} else {
		if imgui.Button("destroy visitor") {
			s.visitorBodyID.Destroy()
			s.visitorBodyID = box2d.Body{}
			// Retain visitorShapeID for end events.
		} else {
			drawSensorEventsCheckbox("visitor events", s.visitorShapeID)
			drawBodyEnableCheckbox("enable visitor body", s.visitorBodyID)
		}
	}

	imgui.Separator()

	if s.sensorBodyID1 == (box2d.Body{}) {
		if imgui.Button("create sensor1") {
			s.CreateSensor1()
		}
	} else {
		if imgui.Button("destroy sensor1") {
			s.sensorBodyID1.Destroy()
			s.sensorBodyID1 = box2d.Body{}
			// Retain sensorShapeID1 for end events.
		} else {
			drawSensorEventsCheckbox("sensor 1 events", s.sensorShapeID1)
			drawBodyEnableCheckbox("enable sensor1 body", s.sensorBodyID1)
		}
	}

	imgui.Separator()

	if s.sensorBodyID2 == (box2d.Body{}) {
		if imgui.Button("create sensor2") {
			s.CreateSensor2()
		}
	} else {
		if imgui.Button("destroy sensor2") {
			s.sensorBodyID2.Destroy()
			s.sensorBodyID2 = box2d.Body{}
			// Retain sensorShapeID2 for end events.
		} else {
			drawSensorEventsCheckbox("sensor2 events", s.sensorShapeID2)
			drawBodyEnableCheckbox("enable sensor2 body", s.sensorBodyID2)
		}
	}

	return true
}

func (s *SensorBookend) Step() {
	s.Sample.Step()

	for event := range s.worldID.SensorBeginEvents() {
		if event.SensorShape == s.sensorShapeID1 {
			if event.VisitorShape == s.visitorShapeID {
				s.isVisiting1 = true
			} else {
				s.sensorsOverlapCount += 1
			}
		} else {
			if event.VisitorShape == s.visitorShapeID {
				s.isVisiting2 = true
			} else {
				s.sensorsOverlapCount += 1
			}
		}
	}

	for event := range s.worldID.SensorEndEvents() {
		if event.SensorShape == s.sensorShapeID1 {
			if event.VisitorShape == s.visitorShapeID {
				s.isVisiting1 = false
			} else {
				s.sensorsOverlapCount -= 1
			}
		} else {
			if event.VisitorShape == s.visitorShapeID {
				s.isVisiting2 = false
			} else {
				s.sensorsOverlapCount -= 1
			}
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

	s.DrawScreenTextLine("visiting 1 == %t", s.isVisiting1)
	s.DrawScreenTextLine("visiting 2 == %t", s.isVisiting2)
	s.DrawScreenTextLine("sensors overlap count == %d", s.sensorsOverlapCount)
}

const (
	footGround = 0x00000001
	footPlayer = 0x00000002
	footFoot   = 0x00000004
)

type FootSensor struct {
	Sample

	playerID     box2d.Body
	sensorID     box2d.Shape
	overlapCount int
}

func CreateFootSensor(context *SampleContext) Sampler {
	s := &FootSensor{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 6.0}
		s.context.camera.zoom = 7.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		var points [18]box2d.Vec2
		x := float32(9.0)
		for i := range 18 {
			points[i] = box2d.Vec2{X: x, Y: 0.0}
			x -= 1.0
		}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.Ghost1 = box2d.Vec2{X: 10.0, Y: 0.0}
		chainDef.Ghost2 = box2d.Vec2{X: -9.0, Y: 0.0}
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
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.5}, Center2: box2d.Vec2{X: 0.0, Y: 0.5}, Radius: 0.5}
		s.playerID.CreateCapsuleShape(&shapeDef, &capsule)

		box := box2d.MakeOffsetBox(0.5, 0.25, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
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
	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.playerID.ApplyForceToCenter(box2d.Vec2{X: -50.0, Y: 0.0}, true)
	}

	if s.context.window.GetKey(glfw.KeyD) == glfw.Press {
		s.playerID.ApplyForceToCenter(box2d.Vec2{X: 50.0, Y: 0.0}, true)
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

	s.DrawScreenTextLine("count == %d", s.overlapCount)

	for shapeID := range s.sensorID.SensorData() {
		aabb := shapeID.AABB()
		point := aabb.Center()
		s.draw.DrawPoint(point.Pos(), 10.0, box2d.ColorWhite)
	}
}

const contactEventCount = 20

type ContactEvent struct {
	Sample

	playerID     box2d.Body
	coreShapeID  box2d.Shape
	debrisIDs    [contactEventCount]box2d.Body
	bodyUserData [contactEventCount]int
	force        float32
	wait         float32
}

func CreateContactEvent(context *SampleContext) Sampler {
	s := &ContactEvent{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 1.75
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		points := []box2d.Vec2{{X: 40.0, Y: -40.0}, {X: -40.0, Y: -40.0}, {X: -40.0, Y: 40.0}, {X: 40.0, Y: 40.0}}

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
	x := scenes.RandomFloatRange(-38.0, 38.0)
	y := scenes.RandomFloatRange(-38.0, 38.0)
	bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
	bodyDef.Rotation = box2d.MakeRot(scenes.RandomFloatRange(-box2d.Pi, box2d.Pi))
	vx := scenes.RandomFloatRange(-5.0, 5.0)
	vy := scenes.RandomFloatRange(-5.0, 5.0)
	bodyDef.LinearVelocity = box2d.Vec2{X: vx, Y: vy}
	bodyDef.AngularVelocity = scenes.RandomFloatRange(-1.0, 1.0)
	bodyDef.GravityScale = 0.0
	bodyDef.UserData = &s.bodyUserData[index]
	s.debrisIDs[index] = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 0.8

	// No events when debris hits debris
	shapeDef.EnableContactEvents = false

	if (index+1)%3 == 0 {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		s.debrisIDs[index].CreateCircleShape(&shapeDef, &circle)
	} else if (index+1)%2 == 0 {
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.25}, Center2: box2d.Vec2{X: 0.0, Y: 0.25}, Radius: 0.25}
		s.debrisIDs[index].CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		box := box2d.MakeBox(0.4, 0.6)
		s.debrisIDs[index].CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *ContactEvent) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderFloatV("force", &s.force, 100.0, 500.0, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *ContactEvent) drawEventContacts(shapeID, otherShapeID box2d.Shape, color box2d.HexColor) {
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
				s.draw.DrawPoint(p1, 10.0, box2d.ColorWhite)
			}
		}
	}
}

func (s *ContactEvent) Step() {
	s.DrawScreenTextLine("move using WASD")

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

	var debrisToAttach [contactEventCount]int
	var shapesToDestroy [contactEventCount]box2d.Shape
	attachCount := 0
	destroyCount := 0

	// Process contact begin touch events.
	for event := range s.worldID.ContactBeginEvents() {
		bodyIDA := event.ShapeA.Body()
		bodyIDB := event.ShapeB.Body()

		// The begin touch events have the contact manifolds, but the impulses are zero. This is because the manifolds
		// are gathered before the contact solver is run.

		// We can get the final contact data from the shapes. The manifold is shared by the two shapes, so we just need the
		// contact data from one of the shapes. Choose the one with the smallest number of contacts.

		capacityA := event.ShapeA.ContactCapacity()
		capacityB := event.ShapeB.ContactCapacity()

		if capacityA < capacityB {
			s.drawEventContacts(event.ShapeA, event.ShapeB, box2d.ColorBlueViolet)
		} else {
			s.drawEventContacts(event.ShapeB, event.ShapeA, box2d.ColorYellowGreen)
		}

		addShapeToDestroy := func(shapeID box2d.Shape) {
			if shapeID != s.coreShapeID && destroyCount < contactEventCount {
				// player non-core shape hit the wall

				found := false
				for j := range destroyCount {
					if shapeID == shapesToDestroy[j] {
						found = true
						break
					}
				}

				// avoid double deletion
				if !found {
					shapesToDestroy[destroyCount] = shapeID
					destroyCount += 1
				}
			}
		}

		if bodyIDA == s.playerID {
			userDataB, _ := bodyIDB.UserData().(*int)
			if userDataB == nil {
				addShapeToDestroy(event.ShapeA)
			} else if attachCount < contactEventCount {
				debrisToAttach[attachCount] = *userDataB
				attachCount += 1
			}
		} else {
			// Only expect events for the player
			userDataA, _ := bodyIDA.UserData().(*int)
			if userDataA == nil {
				addShapeToDestroy(event.ShapeB)
			} else if attachCount < contactEventCount {
				debrisToAttach[attachCount] = *userDataA
				attachCount += 1
			}
		}
	}

	// Attach debris to player body
	for i := range attachCount {
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

		default:
			panic("unexpected shape type")
		}

		debrisID.Destroy()
		s.debrisIDs[index] = box2d.Body{}
	}

	for i := range destroyCount {
		updateMass := false
		shapesToDestroy[i].Destroy(updateMass)
	}

	if destroyCount > 0 {
		// Update mass just once
		s.playerID.UpdateMassFromShapes()
	}

	if s.context.hertz > 0.0 && !s.context.pause {
		s.wait -= 1.0 / s.context.hertz
		if s.wait < 0.0 {
			s.SpawnDebris()
			s.wait += 0.5
		}
	}
}

const bodyMoveCount = 50

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

func CreateBodyMove(context *SampleContext) Sampler {
	s := &BodyMove{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 8.0}
		s.context.camera.zoom = 25.0 * 0.55
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.1

		box := box2d.MakeOffsetBox(12.0, 0.1, box2d.Vec2{X: -10.0, Y: -0.1}, box2d.MakeRot(-0.15*box2d.Pi))
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(12.0, 0.1, box2d.Vec2{X: 10.0, Y: -0.1}, box2d.MakeRot(0.15*box2d.Pi))
		groundID.CreatePolygonShape(&shapeDef, &box)

		shapeDef.Material.Restitution = 0.8

		box = box2d.MakeOffsetBox(0.1, 10.0, box2d.Vec2{X: 19.9, Y: 10.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(0.1, 10.0, box2d.Vec2{X: -19.9, Y: 10.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(20.0, 0.1, box2d.Vec2{X: 0.0, Y: 20.1}, box2d.RotIdentity)
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
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.25, Y: 0.0}, Center2: box2d.Vec2{X: 0.25, Y: 0.0}, Radius: 0.25}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.35}
	square := box2d.MakeSquare(0.35)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()

	x, y := float32(-5.0), float32(10.0)
	for i := 0; i < 10 && s.count < bodyMoveCount; i++ {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
		bodyDef.IsBullet = s.count%12 == 0
		bodyDef.UserData = s.count
		s.bodyIDs[s.count] = s.worldID.CreateBody(&bodyDef)
		s.sleeping[s.count] = false

		remainder := s.count % 4
		if remainder == 0 {
			s.bodyIDs[s.count].CreateCapsuleShape(&shapeDef, &capsule)
		} else if remainder == 1 {
			s.bodyIDs[s.count].CreateCircleShape(&shapeDef, &circle)
		} else if remainder == 2 {
			s.bodyIDs[s.count].CreatePolygonShape(&shapeDef, &square)
		} else {
			poly := scenes.RandomPolygon(0.75)
			poly.Radius = 0.1
			s.bodyIDs[s.count].CreatePolygonShape(&shapeDef, &poly)
		}

		s.count += 1
		x += 1.0
	}
}

func (s *BodyMove) DrawControls() bool {
	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()
		def.Position = s.explosionPosition
		def.Radius = s.explosionRadius
		def.Falloff = 0.1
		def.ImpulsePerLength = s.explosionMagnitude
		s.worldID.Explode(&def)
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderFloatV("Magnitude", &s.explosionMagnitude, -20.0, 20.0, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *BodyMove) Step() {
	if !s.context.pause && (s.stepCount&15) == 15 && s.count < bodyMoveCount {
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
		} else {
			if *sleeping {
				*sleeping = false
				s.sleepCount -= 1
			}
		}
	}

	s.draw.DrawCircle(s.explosionPosition, s.explosionRadius, box2d.ColorAzure)

	s.DrawScreenTextLine("sleep count: %d", s.sleepCount)
}

const (
	sensorTypesGround  = 0x00000001
	sensorTypesSensor  = 0x00000002
	sensorTypesDefault = 0x00000004
)

type SensorTypes struct {
	Sample

	staticSensorID    box2d.Shape
	kinematicSensorID box2d.Shape
	dynamicSensorID   box2d.Shape

	kinematicBodyID box2d.Body
}

func CreateSensorTypes(context *SampleContext) Sampler {
	s := &SensorTypes{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 3.0}
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

		groundSegment := box2d.Segment{Point1: box2d.Vec2{X: -6.0, Y: 0.0}, Point2: box2d.Vec2{X: 6.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{Point1: box2d.Vec2{X: -6.0, Y: 0.0}, Point2: box2d.Vec2{X: -6.0, Y: 4.0}}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{Point1: box2d.Vec2{X: 6.0, Y: 0.0}, Point2: box2d.Vec2{X: 6.0, Y: 4.0}}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "static sensor"
		bodyDef.Type = box2d.StaticBody
		bodyDef.Position = box2d.Pos{X: -3.0, Y: 0.8}
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
		bodyDef.Position = box2d.Pos{X: 3.0, Y: 1.0}
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
		box = box2d.MakeSquare(0.8)
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

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
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

	s.DrawScreenTextLine("%s", buffer.String())
}

func (s *SensorTypes) Step() {
	position := s.kinematicBodyID.Position()
	if position.Y < 0.0 {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: 1.0})
	} else if position.Y > 3.0 {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: 0.0, Y: -1.0})
	}

	s.Sample.Step()

	s.PrintOverlaps(s.staticSensorID, "static")
	s.PrintOverlaps(s.kinematicSensorID, "kinematic")
	s.PrintOverlaps(s.dynamicSensorID, "dynamic")

	origin := box2d.Pos{X: 5.0, Y: 1.0}
	translation := box2d.Vec2{X: -10.0, Y: 0.0}
	result := s.worldID.CastRayClosest(origin, translation, box2d.DefaultQueryFilter())
	s.draw.DrawLine(origin, origin.Add(translation), box2d.ColorDimGray)

	if result.Hit {
		s.draw.DrawPoint(result.Point, 10.0, box2d.ColorCyan)
	}
}

const jointEventCount = 6

// This sample shows how to break joints when the internal reaction force becomes large. Instead of polling, this uses events.
type JointEvent struct {
	Sample

	jointIDs [jointEventCount]box2d.Joint
}

func CreateJointEvent(context *SampleContext) Sampler {
	s := &JointEvent{}
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

	forceThreshold := float32(20000.0)
	torqueThreshold := float32(10000.0)

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
			index := event.UserData.(int)
			event.Joint.Destroy()
			s.jointIDs[index] = box2d.Joint{}
		}
	}
}

type PersistentContact struct {
	Sample

	contactID box2d.Contact
}

func CreatePersistentContact(context *SampleContext) Sampler {
	s := &PersistentContact{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 6.0}
		s.context.camera.zoom = 7.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		var points [22]box2d.Vec2
		x := float32(10.0)
		for i := range 20 {
			points[i] = box2d.Vec2{X: x, Y: 0.0}
			x -= 1.0
		}

		points[20] = box2d.Vec2{X: -9.0, Y: 10.0}
		points[21] = box2d.Vec2{X: 10.0, Y: 10.0}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:]
		chainDef.IsLoop = true

		groundID.CreateChain(&chainDef)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -8.0, Y: 1.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: 2.0, Y: 0.0}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableContactEvents = true
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	s.contactID = box2d.Contact{}
	return s
}

func (s *PersistentContact) Step() {
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
			s.draw.DrawPoint(p1, 6.0, box2d.ColorCrimson)
			s.draw.DrawString(s.camera, p1, box2d.ColorWhite, "%.2f", manifoldPoint.TotalNormalImpulse)
		}
	} else {
		s.contactID = box2d.Contact{}
	}
}

const sensorHitsTransformCapacity = 20

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

func CreateSensorHits(context *SampleContext) Sampler {
	s := &SensorHits{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 7.5
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "ground"

		groundID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()

		groundSegment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &groundSegment)

		groundSegment = box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 10.0}}
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

		segment := box2d.Segment{Point1: box2d.Vec2{X: 0.0, Y: 0.0}, Point2: box2d.Vec2{X: 0.0, Y: 10.0}}
		s.staticSensorID = bodyID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Kinematic sensor
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "kinematic sensor"
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.5, Y: 0.0}

		s.kinematicBodyID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true

		segment := box2d.Segment{Point1: box2d.Vec2{X: 0.0, Y: 0.0}, Point2: box2d.Vec2{X: 0.0, Y: 10.0}}
		s.kinematicSensorID = s.kinematicBodyID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Dynamic sensor
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Name = "dynamic sensor"
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 4.0, Y: 1.0}

		s.dynamicBodyID = s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.IsSensor = true
		shapeDef.EnableSensorEvents = true

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: 1.0}, Center2: box2d.Vec2{X: 0.0, Y: 9.0}, Radius: 0.1}
		s.dynamicSensorID = s.dynamicBodyID.CreateCapsuleShape(&shapeDef, &capsule)

		pivot := bodyDef.Position.Add(box2d.Vec2{X: 0.0, Y: 6.0})
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
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	s.transformCount = 0
	s.beginCount = 0
	s.endCount = 0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -26.7, Y: 6.0}
	speed := scenes.RandomFloatRange(200.0, 300.0)
	bodyDef.LinearVelocity = box2d.Vec2{X: speed, Y: 0.0}
	bodyDef.IsBullet = s.isBullet
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true
	shapeDef.Material.Friction = 0.8
	shapeDef.Material.RollingResistance = 0.01
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.25}
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
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: -0.5, Y: 0.0})
	} else if p.X < -1.0 {
		s.kinematicBodyID.SetLinearVelocity(box2d.Vec2{X: 0.5, Y: 0.0})
	}

	x := s.jointID.Translation()
	if x > 1.0 {
		s.jointID.SetMotorSpeed(-0.5)
	} else if x < -1.0 {
		s.jointID.SetMotorSpeed(0.5)
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

	s.DrawScreenTextLine("begin touch count = %d", s.beginCount)
	s.DrawScreenTextLine("end touch count = %d", s.endCount)
}

// This shows how to create a projectile that explodes on impact
type ProjectileEvent struct {
	Sample

	projectileID      box2d.Body
	projectileShapeID box2d.Shape
	point1            box2d.Pos
	point2            box2d.Pos
	dragging          bool
}

func CreateProjectileEvent(context *SampleContext) Sampler {
	s := &ProjectileEvent{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: -7.0, Y: 9.0}
		s.context.camera.zoom = 14.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.EnableSensorEvents = true

		segment := box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 20.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: -30.0, Y: 0.0}, Point2: box2d.Vec2{X: 30.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.projectileID = box2d.Body{}
	s.projectileShapeID = box2d.Shape{}
	s.dragging = false
	s.point1 = box2d.Pos{}
	s.point2 = box2d.Pos{}

	box := box2d.MakeRoundedBox(0.45, 0.45, 0.05)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true

	offset := float32(0.01)

	x := float32(8.0)
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	for i := range 8 {
		shift := offset
		if i%2 == 0 {
			shift = -offset
		}
		bodyDef.Position = box2d.Pos{X: posScalar(x + shift), Y: posScalar(0.5 + 1.0*float32(i))}

		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

func (s *ProjectileEvent) FireProjectile() {
	if s.projectileID != (box2d.Body{}) {
		s.projectileID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = s.point1
	bodyDef.LinearVelocity = s.point2.Sub(s.point1).Scale(4.0)
	bodyDef.IsBullet = true

	s.projectileID = s.worldID.CreateBody(&bodyDef)

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.25}
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
	s.DrawScreenTextLine("Use Ctrl + Left Mouse to drag and shoot a projectile")

	s.Sample.Step()

	if s.dragging {
		s.draw.DrawLine(s.point1, s.point2, box2d.ColorWhite)
		s.draw.DrawPoint(s.point1, 5.0, box2d.ColorGreen)
		s.draw.DrawPoint(s.point2, 5.0, box2d.ColorRed)
	}

	for event := range s.worldID.ContactBeginEvents() {
		if event.ShapeA == s.projectileShapeID || event.ShapeB == s.projectileShapeID {
			if event.Contact.IsValid() {
				data := event.Contact.Data()

				if data.Manifold.PointCount > 0 {
					explosionDef := box2d.DefaultExplosionDef()
					explosionDef.Position = data.ShapeA.Body().WorldCenter().Add(data.Manifold.Points[0].AnchorA)
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

func CreateCircleImpulse(context *SampleContext) Sampler {
	s := &CircleImpulse{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.7}
		s.context.camera.zoom = 3.4
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
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
	bodyDef.LinearVelocity.Y = -25.0
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
	s.Sample.Step()

	for event := range s.worldID.ContactHitEvents() {
		s.draw.DrawPoint(event.Point, 10.0, box2d.ColorWhite)

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
	s.DrawScreenTextLine("mass = %.6g, gravity = %.6g, restitution = %.6g", s.mass, gravity, restitution)

	for _, e := range s.events {
		s.DrawScreenTextLine("hit speed = %.6g, hit momentum = %.6g, final impulse = %.6g, total impulse = %.6g", e.speed,
			s.mass*e.speed, e.impulse, e.totalImpulse)
	}
}

func init() {
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
