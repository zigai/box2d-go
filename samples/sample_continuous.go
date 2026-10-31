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

func CreateBounceHouse(context *SampleContext) Sampler {
	s := &BounceHouse{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 0.45
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: -10.0}, Point2: box2d.Vec2{X: 10.0, Y: -10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: -10.0}, Point2: box2d.Vec2{X: 10.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 10.0}, Point2: box2d.Vec2{X: -10.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 10.0}, Point2: box2d.Vec2{X: -10.0, Y: -10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.shapeType = continuousCircleShape
	s.bodyID = box2d.Body{}
	s.enableHitEvents = true

	s.Launch()
	return s
}

func (s *BounceHouse) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.LinearVelocity = box2d.Vec2{X: 10.0, Y: 20.0}
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
	bodyDef.GravityScale = 0.0
	bodyDef.IsBullet = true

	// Circle shapes centered on the body can spin fast without risk of tunnelling.
	bodyDef.AllowFastRotation = s.shapeType == continuousCircleShape

	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.EnableHitEvents = s.enableHitEvents

	if s.shapeType == continuousCircleShape {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		s.bodyID.CreateCircleShape(&shapeDef, &circle)
	} else if s.shapeType == continuousCapsuleShape {
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
		s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		h := float32(0.1)
		box := box2d.MakeBox(20.0*h, h)
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *BounceHouse) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
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
			s.draw.DrawCircle(e.point, 0.1, box2d.ColorOrangeRed)
			s.draw.DrawString(s.camera, e.point, box2d.ColorWhite, "%.1f", e.speed)
		}
	}
}

type BounceHumans struct {
	Sample

	humans     [5]scenes.Human
	humanCount int
	countDown  float32
	time       float32
}

func CreateBounceHumans(context *SampleContext) Sampler {
	s := &BounceHumans{}
	s.init(context, s, true)

	s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
	s.context.camera.zoom = 12.0

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 1.3
	shapeDef.Material.Friction = 0.1

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: -10.0}, Point2: box2d.Vec2{X: 10.0, Y: -10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: -10.0}, Point2: box2d.Vec2{X: 10.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 10.0}, Point2: box2d.Vec2{X: -10.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 10.0}, Point2: box2d.Vec2{X: -10.0, Y: -10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 2.0}
	shapeDef.Material.Restitution = 2.0
	groundID.CreateCircleShape(&shapeDef, &circle)
	return s
}

func (s *BounceHumans) Step() {
	if s.humanCount < 5 && s.countDown <= 0.0 {
		jointFrictionTorque := float32(0.0)
		jointHertz := float32(1.0)
		jointDampingRatio := float32(0.1)

		scenes.CreateHuman(&s.humans[s.humanCount], s.worldID, box2d.Pos{X: 0.0, Y: 5.0}, 1.0, jointFrictionTorque, jointHertz,
			jointDampingRatio, 1, nil, true)

		s.countDown = 2.0
		s.humanCount += 1
	}

	timeStep := float32(1.0 / 60.0)
	cs1 := box2d.ComputeCosSin(0.5 * s.time)
	cs2 := box2d.ComputeCosSin(s.time)
	gravity := float32(10.0)
	gravityVec := box2d.Vec2{X: gravity * cs1.Sine, Y: gravity * cs2.Cosine}
	s.draw.DrawLine(box2d.Pos{}, box2d.Vec2{X: 3.0 * cs1.Sine, Y: 3.0 * cs2.Cosine}.Pos(), box2d.ColorWhite)
	s.time += timeStep
	s.countDown -= timeStep
	s.worldID.SetGravity(gravityVec)

	s.Sample.Step()
}

type ChainDrop struct {
	Sample

	bodyID  box2d.Body
	shapeID box2d.Shape
	yOffset float32
	speed   float32
}

func CreateChainDrop(context *SampleContext) Sampler {
	s := &ChainDrop{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 0.35
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -6.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	points := []box2d.Vec2{{X: -10.0, Y: -2.0}, {X: 10.0, Y: -2.0}, {X: 10.0, Y: 1.0}, {X: -10.0, Y: 1.0}}

	chainDef := box2d.DefaultChainDef()
	chainDef.Points = points
	chainDef.IsLoop = true

	groundID.CreateChain(&chainDef)

	s.bodyID = box2d.Body{}
	s.yOffset = -0.1
	s.speed = -42.0

	s.Launch()
	return s
}

func (s *ChainDrop) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: s.speed}
	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(10.0 + s.yOffset)}
	bodyDef.Rotation = box2d.MakeRot(0.5 * box2d.Pi)
	bodyDef.MotionLocks.AngularZ = true
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
}

func (s *ChainDrop) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderFloatV("Speed", &s.speed, -100.0, 0.0, "%.0f", 0)
	imgui.SliderFloatV("Y Offset", &s.yOffset, -1.0, 1.0, "%.1f", 0)
	imgui.PopItemWidth()

	if imgui.Button("Launch") {
		s.Launch()
	}

	return true
}

type ChainSlide struct {
	Sample
}

func CreateChainSlide(context *SampleContext) Sampler {
	s := &ChainSlide{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 10.0}
		s.context.camera.zoom = 15.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		const count = 80
		var points [count]box2d.Vec2

		w := float32(2.0)
		h := float32(1.0)
		x, y := float32(20.0), float32(0.0)
		for i := 0; i < 20; i++ {
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
		bodyDef.LinearVelocity = box2d.Vec2{X: 100.0, Y: 0.0}
		bodyDef.Position = box2d.Pos{X: -19.5, Y: 0.0 + 0.5}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.0
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
	return s
}

type SegmentSlide struct {
	Sample
}

func CreateSegmentSlide(context *SampleContext) Sampler {
	s := &SegmentSlide{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 10.0}
		s.context.camera.zoom = 15.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -40.0, Y: 0.0}, Point2: box2d.Vec2{X: 40.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: 40.0, Y: 0.0}, Point2: box2d.Vec2{X: 40.0, Y: 10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.LinearVelocity = box2d.Vec2{X: 100.0, Y: 0.0}
		bodyDef.Position = box2d.Pos{X: -20.0, Y: 0.7}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
	return s
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

func CreateSkinnyBox(context *SampleContext) Sampler {
	s := &SkinnyBox{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.0, Y: 5.0}
		s.context.camera.zoom = 25.0 * 0.25
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.9
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(0.1, 1.0, box2d.Vec2{X: 0.0, Y: 1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.autoTest = false
	s.bullet = false
	s.capsule = false

	s.bodyID = box2d.Body{}
	s.bulletID = box2d.Body{}

	s.Launch()
	return s
}

func (s *SkinnyBox) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	if s.bulletID != (box2d.Body{}) {
		s.bulletID.Destroy()
	}

	s.angularVelocity = scenes.RandomFloatRange(-50.0, 50.0)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 8.0}
	bodyDef.AngularVelocity = s.angularVelocity
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.9

	s.bodyID = s.worldID.CreateBody(&bodyDef)

	if s.capsule {
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -1.0}, Center2: box2d.Vec2{X: 0.0, Y: 1.0}, Radius: 0.1}
		s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		polygon := box2d.MakeBox(2.0, 0.05)
		s.bodyID.CreatePolygonShape(&shapeDef, &polygon)
	}

	if s.bullet {
		polygon := box2d.MakeBox(0.25, 0.25)
		s.x = scenes.RandomFloatRange(-1.0, 1.0)
		bodyDef.Position = box2d.Pos{X: posScalar(s.x), Y: 10.0}
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

// This sample shows ghost bumps
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

func CreateGhostBumps(context *SampleContext) Sampler {
	s := &GhostBumps{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.5, Y: 16.0}
		s.context.camera.zoom = 25.0 * 0.8
	}

	s.groundID = box2d.Body{}
	s.bodyID = box2d.Body{}
	s.shapeID = box2d.Shape{}
	s.shapeType = continuousCircleShape
	s.round = 0.0
	s.friction = 0.2
	s.bevel = 0.0
	s.useChain = true

	s.CreateScene()
	s.Launch()
	return s
}

func (s *GhostBumps) CreateScene() {
	if s.groundID != (box2d.Body{}) {
		s.groundID.Destroy()
	}

	s.shapeID = box2d.Shape{}

	bodyDef := box2d.DefaultBodyDef()
	s.groundID = s.worldID.CreateBody(&bodyDef)

	m := float32(1.0 / math.Sqrt(2.0))
	mm := float32(2.0 * (math.Sqrt(2.0) - 1.0))
	hx, hy := float32(4.0), float32(0.25)

	if s.useChain {
		var points [20]box2d.Vec2
		points[0] = box2d.Vec2{X: -3.0 * hx, Y: hy}
		points[1] = points[0].Add(box2d.Vec2{X: -2.0 * hx * m, Y: 2.0 * hx * m})
		points[2] = points[1].Add(box2d.Vec2{X: -2.0 * hx * m, Y: 2.0 * hx * m})
		points[3] = points[2].Add(box2d.Vec2{X: -2.0 * hx * m, Y: 2.0 * hx * m})
		points[4] = points[3].Add(box2d.Vec2{X: -2.0 * hy * m, Y: -2.0 * hy * m})
		points[5] = points[4].Add(box2d.Vec2{X: 2.0 * hx * m, Y: -2.0 * hx * m})
		points[6] = points[5].Add(box2d.Vec2{X: 2.0 * hx * m, Y: -2.0 * hx * m})
		points[7] = points[6].Add(box2d.Vec2{X: 2.0*hx*m + 2.0*hy*(1.0-m), Y: -2.0*hx*m - 2.0*hy*(1.0-m)})
		points[8] = points[7].Add(box2d.Vec2{X: 2.0*hx + hy*mm, Y: 0.0})
		points[9] = points[8].Add(box2d.Vec2{X: 2.0 * hx, Y: 0.0})
		points[10] = points[9].Add(box2d.Vec2{X: 2.0*hx + hy*mm, Y: 0.0})
		points[11] = points[10].Add(box2d.Vec2{X: 2.0*hx*m + 2.0*hy*(1.0-m), Y: 2.0*hx*m + 2.0*hy*(1.0-m)})
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
				{X: hx + hb, Y: hy - 0.05},
				{X: hx, Y: hy},
				{X: -hx, Y: hy},
				{X: -hx - hb, Y: hy - 0.05},
				{X: -hx - hb, Y: -hy + 0.05},
				{X: -hx, Y: -hy},
				{X: hx, Y: -hy},
				{X: hx + hb, Y: -hy + 0.05},
			}
			hull = box2d.ComputeHull(vs)
		} else {
			vs := []box2d.Vec2{{X: hx, Y: hy}, {X: -hx, Y: hy}, {X: -hx, Y: -hy}, {X: hx, Y: -hy}}
			hull = box2d.ComputeHull(vs)
		}

		var transform box2d.Transform
		var x, y float32

		createSegment := func() {
			transform.P = box2d.Vec2{X: x, Y: y}
			polygon := box2d.MakeOffsetPolygon(&hull, transform.P, transform.Q)
			s.groundID.CreatePolygonShape(&shapeDef, &polygon)
		}

		// Left slope
		x = -3.0*hx - m*hx - m*hy
		y = hy + m*hx - m*hy
		transform.Q = box2d.MakeRot(-0.25 * box2d.Pi)

		for range 3 {
			createSegment()
			x -= 2.0 * m * hx
			y += 2.0 * m * hx
		}

		x = -2.0 * hx
		y = 0.0
		transform.Q = box2d.MakeRot(0.0)

		for range 3 {
			createSegment()
			x += 2.0 * hx
		}

		x = 3.0*hx + m*hx + m*hy
		y = hy + m*hx - m*hy
		transform.Q = box2d.MakeRot(0.25 * box2d.Pi)

		for range 3 {
			createSegment()
			x += 2.0 * m * hx
			y += 2.0 * m * hx
		}
	}
}

func (s *GhostBumps) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
		s.shapeID = box2d.Shape{}
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -28.0, Y: 18.0}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: 0.0}
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = s.friction

	if s.shapeType == continuousCircleShape {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
	} else if s.shapeType == continuousCapsuleShape {
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
		s.shapeID = s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		h := 0.5 - s.round
		box := box2d.MakeRoundedBox(h, 2.0*h, s.round)
		s.shapeID = s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *GhostBumps) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

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
		imgui.SliderFloatV("Round", &s.round, 0.0, 0.4, "%.1f", 0)
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

// Speculative collision failure case suggested by Dirk Gregorius. This uses
// a simple fallback scheme to prevent tunneling.
type SpeculativeFallback struct {
	Sample
}

func CreateSpeculativeFallback(context *SampleContext) Sampler {
	s := &SpeculativeFallback{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.0, Y: 5.0}
		s.context.camera.zoom = 25.0 * 0.25
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		points := []box2d.Vec2{{X: -2.0, Y: 4.0}, {X: 2.0, Y: 4.0}, {X: 2.0, Y: 4.1}, {X: -0.5, Y: 4.2}, {X: -2.0, Y: 4.2}}
		hull := box2d.ComputeHull(points)
		poly := box2d.MakePolygon(&hull, 0.0)
		groundID.CreatePolygonShape(&shapeDef, &poly)
	}

	// Fast moving skinny box. Also testing a large shape offset.
	{
		offset := float32(8.0)
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(offset), Y: 12.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(2.0, 0.05, box2d.Vec2{X: -offset, Y: 0.0}, box2d.MakeRot(box2d.Pi))
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

type SpeculativeSliver struct {
	Sample
}

func CreateSpeculativeSliver(context *SampleContext) Sampler {
	s := &SpeculativeSliver{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.75}
		s.context.camera.zoom = 2.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 12.0}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		points := []box2d.Vec2{{X: -2.0, Y: 0.0}, {X: -1.0, Y: 0.0}, {X: 2.0, Y: 0.5}}
		hull := box2d.ComputeHull(points)
		poly := box2d.MakePolygon(&hull, 0.0)
		bodyID.CreatePolygonShape(&shapeDef, &poly)
	}
	return s
}

// This shows that while Box2D uses speculative collision, it does not lead to speculative ghost collisions at small distances
type SpeculativeGhost struct {
	Sample
}

func CreateSpeculativeGhost(context *SampleContext) Sampler {
	s := &SpeculativeGhost{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.75}
		s.context.camera.zoom = 2.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(1.0, 0.1, box2d.Vec2{X: 0.0, Y: 0.9}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		// The speculative distance is 0.02 meters, so this avoid it
		bodyDef.Position = box2d.Pos{X: 0.015, Y: 2.515}
		bodyDef.LinearVelocity = box2d.Vec2{X: 0.1 * 1.25 * s.context.hertz, Y: -0.1 * 1.25 * s.context.hertz}
		bodyDef.GravityScale = 0.0
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeSquare(0.25)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

// This shows that Box2D does not have pixel perfect collision.
type PixelImperfect struct {
	Sample

	ballID box2d.Body
}

func CreatePixelImperfect(context *SampleContext) Sampler {
	s := &PixelImperfect{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 7.0, Y: 5.0}
		s.context.camera.zoom = 6.0
	}

	pixelsPerMeter := float32(30.0)

	{
		block4BodyDef := box2d.DefaultBodyDef()
		block4BodyDef.Type = box2d.StaticBody
		block4BodyDef.Position = box2d.Pos{X: posScalar(175.0 / pixelsPerMeter), Y: posScalar(150.0 / pixelsPerMeter)}
		block4BodyID := s.worldID.CreateBody(&block4BodyDef)
		block4Shape := box2d.MakeBox(20.0/pixelsPerMeter, 10.0/pixelsPerMeter)
		block4ShapeDef := box2d.DefaultShapeDef()
		block4ShapeDef.Material.Friction = 0.0
		block4BodyID.CreatePolygonShape(&block4ShapeDef, &block4Shape)
	}

	{
		ballBodyDef := box2d.DefaultBodyDef()
		ballBodyDef.Type = box2d.DynamicBody
		ballBodyDef.Position = box2d.Pos{X: posScalar(200.0 / pixelsPerMeter), Y: posScalar(275.0 / pixelsPerMeter)}
		ballBodyDef.GravityScale = 0.0

		s.ballID = s.worldID.CreateBody(&ballBodyDef)
		ballShape := box2d.MakeRoundedBox(4.0/pixelsPerMeter, 4.0/pixelsPerMeter, 0.9/pixelsPerMeter)
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
	s.DrawScreenTextLine("p.x = %.9f, v.y = %.9f", p.X, v.Y)

	s.Sample.Step()
}

type RestitutionThreshold struct {
	Sample

	ballID box2d.Body
}

func CreateRestitutionThreshold(context *SampleContext) Sampler {
	s := &RestitutionThreshold{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 7.0, Y: 5.0}
		s.context.camera.zoom = 6.0
	}

	pixelsPerMeter := float32(30.0)

	// With the default threshold the ball will not bounce.
	s.worldID.SetRestitutionThreshold(0.1)

	{
		block0BodyDef := box2d.DefaultBodyDef()
		block0BodyDef.Type = box2d.StaticBody
		block0BodyDef.Position = box2d.Pos{X: posScalar(205.0 / pixelsPerMeter), Y: posScalar(120.0 / pixelsPerMeter)}
		block0BodyDef.Rotation = box2d.MakeRot(70.0 * 3.14 / 180.0)
		block0BodyID := s.worldID.CreateBody(&block0BodyDef)
		block0Shape := box2d.MakeBox(50.0/pixelsPerMeter, 5.0/pixelsPerMeter)
		block0ShapeDef := box2d.DefaultShapeDef()
		block0ShapeDef.Material.Friction = 0.0
		block0BodyID.CreatePolygonShape(&block0ShapeDef, &block0Shape)
	}

	{
		// Make a ball
		ballBodyDef := box2d.DefaultBodyDef()
		ballBodyDef.Type = box2d.DynamicBody
		ballBodyDef.Position = box2d.Pos{X: posScalar(200.0 / pixelsPerMeter), Y: posScalar(250.0 / pixelsPerMeter)}
		s.ballID = s.worldID.CreateBody(&ballBodyDef)

		var ballShape box2d.Circle
		ballShape.Radius = 5.0 / pixelsPerMeter
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
	s.DrawScreenTextLine("p.x = %.9f, v.y = %.9f", p.X, v.Y)

	s.Sample.Step()
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

func CreateDrop(context *SampleContext) Sampler {
	s := &Drop{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.5}
		s.context.camera.zoom = 3.0
		s.context.enableSleep = false
		s.context.debugDraw.DrawJoints = false
	}

	s.human = scenes.Human{}
	s.frameSkip = 0
	s.frameCount = 0
	s.continuous = true
	s.speculative = true

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

func (s *Drop) destroyGround() {
	for _, groundID := range s.groundIDs {
		groundID.Destroy()
	}
	s.groundIDs = s.groundIDs[:0]
}

func (s *Drop) CreateGround1() {
	s.destroyGround()

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	w := float32(0.25)
	count := 40
	segment := box2d.Segment{Point1: box2d.Vec2{X: -0.5 * float32(count) * w, Y: 0.0}, Point2: box2d.Vec2{X: 0.5 * float32(count) * w, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.groundIDs = append(s.groundIDs, groundID)
}

func (s *Drop) CreateGround2() {
	s.destroyGround()

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	w := float32(0.25)
	count := 40

	x := -0.5 * float32(count) * w
	h := float32(0.05)
	for j := 0; j <= count; j++ {
		box := box2d.MakeOffsetBox(0.5*w, h, box2d.Vec2{X: x, Y: 0.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
		x += w
	}

	s.groundIDs = append(s.groundIDs, groundID)
}

func (s *Drop) CreateGround3() {
	s.destroyGround()

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	w := float32(0.25)
	count := 40
	segment := box2d.Segment{Point1: box2d.Vec2{X: -0.5 * float32(count) * w, Y: 0.0}, Point2: box2d.Vec2{X: 0.5 * float32(count) * w, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)
	segment = box2d.Segment{Point1: box2d.Vec2{X: 3.0, Y: 0.0}, Point2: box2d.Vec2{X: 3.0, Y: 8.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.groundIDs = append(s.groundIDs, groundID)
}

// ball
func (s *Drop) Scene1() {
	s.Clear()
	s.CreateGround2()

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 4.0}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -100.0}

	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.125}
	bodyID.CreateCircleShape(&shapeDef, &circle)

	s.bodyIDs = append(s.bodyIDs, bodyID)
	s.frameCount = 1
}

// ruler
func (s *Drop) Scene2() {
	s.Clear()
	s.CreateGround1()

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 4.0}
	bodyDef.Rotation = box2d.MakeRot(0.5 * box2d.Pi)
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: 0.0}
	bodyDef.AngularVelocity = -0.5

	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.75, 0.01)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	s.bodyIDs = append(s.bodyIDs, bodyID)
	s.frameCount = 1
}

// ragdoll
func (s *Drop) Scene3() {
	s.Clear()
	s.CreateGround2()

	jointFrictionTorque := float32(0.03)
	jointHertz := float32(1.0)
	jointDampingRatio := float32(0.5)

	scenes.CreateHuman(&s.human, s.worldID, box2d.Pos{X: 0.0, Y: 40.0}, 1.0, jointFrictionTorque, jointHertz, jointDampingRatio, 1, nil,
		true)

	s.frameCount = 1
}

func (s *Drop) Scene4() {
	s.Clear()
	s.CreateGround3()

	a := float32(0.25)
	box := box2d.MakeSquare(a)

	shapeDef := box2d.DefaultShapeDef()

	offset := float32(0.01)

	for i := range 5 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shift := offset
		if i%2 == 0 {
			shift = -offset
		}
		bodyDef.Position = box2d.Pos{X: posScalar(2.5 + shift), Y: posScalar(a + 2.0*a*float32(i))}
		bodyID := s.worldID.CreateBody(&bodyDef)

		s.bodyIDs = append(s.bodyIDs, bodyID)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.125}
	shapeDef.Density = 4.0

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -7.7, Y: 1.9}
		bodyDef.LinearVelocity = box2d.Vec2{X: 200.0, Y: 0.0}
		bodyDef.IsBullet = true

		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCircleShape(&shapeDef, &circle)
		s.bodyIDs = append(s.bodyIDs, bodyID)
	}

	s.frameCount = 1
}

func (s *Drop) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.Key1:
		s.Scene1()

	case glfw.Key2:
		s.Scene2()

	case glfw.Key3:
		s.Scene3()

	case glfw.Key4:
		s.Scene4()

	case glfw.KeyC:
		s.Clear()
		s.continuous = !s.continuous

	case glfw.KeyS:
		if s.frameSkip > 0 {
			s.frameSkip = 0
		} else {
			s.frameSkip = 60
		}

	default:
		s.Sample.Keyboard(key, action, mods)
	}
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

// This shows a fast moving body that uses continuous collision versus static and dynamic bodies.
// This is achieved by setting the ball body as a *bullet*.
type Pinball struct {
	Sample

	leftJointID  box2d.RevoluteJoint
	rightJointID box2d.RevoluteJoint
	ballID       box2d.Body
}

func CreatePinball(context *SampleContext) Sampler {
	s := &Pinball{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 9.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	s.context.debugDraw.DrawJoints = false

	// Ground body
	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)

		vs := []box2d.Vec2{{X: -8.0, Y: 6.0}, {X: -8.0, Y: 20.0}, {X: 8.0, Y: 20.0}, {X: 8.0, Y: 6.0}, {X: 0.0, Y: -2.0}}

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = vs
		chainDef.IsLoop = true
		groundID.CreateChain(&chainDef)
	}

	// Flippers
	{
		p1, p2 := box2d.Pos{X: -2.0, Y: 0.0}, box2d.Pos{X: 2.0, Y: 0.0}

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.EnableSleep = false

		bodyDef.Position = p1
		leftFlipperID := s.worldID.CreateBody(&bodyDef)

		bodyDef.Position = p2
		rightFlipperID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(1.75, 0.2)

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
		jointDef.LowerAngle = -30.0 * box2d.Pi / 180.0
		jointDef.UpperAngle = 5.0 * box2d.Pi / 180.0
		s.leftJointID = s.worldID.CreateRevoluteJoint(&jointDef)

		jointDef.MotorSpeed = 0.0
		jointDef.Base.LocalFrameA.P = p2.Vec2()
		jointDef.Base.BodyB = rightFlipperID
		jointDef.LowerAngle = -5.0 * box2d.Pi / 180.0
		jointDef.UpperAngle = 30.0 * box2d.Pi / 180.0
		s.rightJointID = s.worldID.CreateRevoluteJoint(&jointDef)
	}

	// Spinners
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -4.0, Y: 17.0}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box1 := box2d.MakeBox(1.5, 0.125)
		box2 := box2d.MakeBox(0.125, 1.5)

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

		bodyDef.Position = box2d.Pos{X: 4.0, Y: 8.0}
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
		bodyDef.Position = box2d.Pos{X: -4.0, Y: 8.0}

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Restitution = 1.5

		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 1.0}
		bodyID.CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: 4.0, Y: 17.0}
		bodyID = s.worldID.CreateBody(&bodyDef)
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	// Ball
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 1.0, Y: 15.0}
		bodyDef.Type = box2d.DynamicBody
		bodyDef.IsBullet = true

		s.ballID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.2}
		s.ballID.CreateCircleShape(&shapeDef, &circle)
	}
	return s
}

func (s *Pinball) DrawControls() bool {
	imText("Flipper: press A")
	return true
}

func (s *Pinball) Step() {
	s.Sample.Step()

	if s.context.window.GetKey(glfw.KeyA) == glfw.Press {
		s.leftJointID.SetMotorSpeed(20.0)
		s.rightJointID.SetMotorSpeed(-20.0)
	} else {
		s.leftJointID.SetMotorSpeed(-10.0)
		s.rightJointID.SetMotorSpeed(10.0)
	}
}

// This shows the importance of secondary collisions in continuous physics.
// This also shows a difficult setup for the solver with an acute angle.
type Wedge struct {
	Sample
}

func CreateWedge(context *SampleContext) Sampler {
	s := &Wedge{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.5}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -4.0, Y: 8.0}, Point2: box2d.Vec2{X: 0.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
		segment = box2d.Segment{Point1: box2d.Vec2{X: 0.0, Y: 0.0}, Point2: box2d.Vec2{X: 0.0, Y: 8.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -0.45, Y: 10.75}
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

func CreateSafetyFactor(context *SampleContext) Sampler {
	s := &SafetyFactor{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 3.0}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(20.0, 1.0)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	s.bodyID = box2d.Body{}
	s.extent = 0.5
	s.height = 0.3
	s.safetyFactor = 0.1
	s.overlap = 0.0

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

	s.draw.DrawLine(box2d.Vec2{X: -3.0, Y: -s.overlap}.Pos(), box2d.Vec2{X: 3.0, Y: -s.overlap}.Pos(), box2d.ColorRed)

	timeStep := float32(0.0)
	if s.context.hertz > 0.0 {
		timeStep = 1.0 / s.context.hertz
	}
	motion := timeStep * s.bodyID.LinearVelocity().Length()

	s.DrawScreenTextLine("actual movement = %.3f m", motion)
	s.DrawScreenTextLine("fast movement = %.3f m", s.safetyFactor*s.extent)
	s.DrawScreenTextLine("overlap = %.4f m", s.overlap)
}

func (s *SafetyFactor) DrawControls() bool {
	changed := false

	imgui.PushItemWidth(10.0 * imgui.FontSize())
	changed = imgui.SliderFloatV("Height", &s.height, 0.0, 2.0, "%.2f", 0) || changed
	changed = imgui.SliderFloatV("Safety", &s.safetyFactor, 0.0, 1.0, "%.2f", 0) || changed
	imgui.PopItemWidth()

	if imgui.Button("Drop") || changed {
		s.Launch()
	}

	return true
}

func init() {
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
