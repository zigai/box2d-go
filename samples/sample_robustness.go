// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d"
)

// Pyramid with heavy box on top
type HighMassRatio1 struct {
	Sample
}

func CreateHighMassRatio1(context *SampleContext) Sampler {
	s := &HighMassRatio1{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 3.0, Y: 14.0}
		s.context.camera.zoom = 25.0
	}

	extent := float32(1.0)

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(50.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		box := box2d.MakeBox(extent, extent)
		shapeDef := box2d.DefaultShapeDef()

		for j := range 3 {
			count := 10
			offset := -20.0*extent + 2.0*(float32(count)+1.0)*extent*float32(j)
			y := extent
			for count > 0 {
				for i := range count {
					coeff := float32(i) - 0.5*float32(count)

					yy := y
					if count == 1 {
						yy = y + 2.0
					}
					bodyDef.Position = box2d.Pos{X: posScalar(2.0*coeff*extent + offset), Y: posScalar(yy)}
					bodyID := s.worldID.CreateBody(&bodyDef)

					shapeDef.Density = 1.0
					if count == 1 {
						shapeDef.Density = (float32(j) + 1.0) * 100.0
					}
					bodyID.CreatePolygonShape(&shapeDef, &box)
				}

				count--
				y += 2.0 * extent
			}
		}
	}
	return s
}

// Big box on small boxes
type HighMassRatio2 struct {
	Sample
}

func CreateHighMassRatio2(context *SampleContext) Sampler {
	s := &HighMassRatio2{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 16.5}
		s.context.camera.zoom = 25.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(50.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		shapeDef := box2d.DefaultShapeDef()

		extent := float32(1.0)
		smallBox := box2d.MakeBox(0.5*extent, 0.5*extent)
		bigBox := box2d.MakeBox(10.0*extent, 10.0*extent)

		{
			bodyDef.Position = box2d.Pos{X: posScalar(-9.0 * extent), Y: posScalar(0.5 * extent)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallBox)
		}

		{
			bodyDef.Position = box2d.Pos{X: posScalar(9.0 * extent), Y: posScalar(0.5 * extent)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallBox)
		}

		{
			bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar((10.0 + 16.0) * extent)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &bigBox)
		}
	}
	return s
}

// Big box on small triangles
type HighMassRatio3 struct {
	Sample
}

func CreateHighMassRatio3(context *SampleContext) Sampler {
	s := &HighMassRatio3{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 16.5}
		s.context.camera.zoom = 25.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(50.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		shapeDef := box2d.DefaultShapeDef()

		extent := float32(1.0)
		points := []box2d.Vec2{{X: -0.5 * extent, Y: 0.0}, {X: 0.5 * extent, Y: 0.0}, {X: 0.0, Y: 1.0 * extent}}
		hull := box2d.ComputeHull(points)
		smallTriangle := box2d.MakePolygon(&hull, 0.0)
		bigBox := box2d.MakeBox(10.0*extent, 10.0*extent)

		{
			bodyDef.Position = box2d.Pos{X: posScalar(-9.0 * extent), Y: posScalar(0.5 * extent)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallTriangle)
		}

		{
			bodyDef.Position = box2d.Pos{X: posScalar(9.0 * extent), Y: posScalar(0.5 * extent)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallTriangle)
		}

		{
			bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar((10.0 + 4.0) * extent)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &bigBox)
		}
	}
	return s
}

type OverlapRecovery struct {
	Sample

	bodyIDs      []box2d.Body
	baseCount    int32
	overlap      float32
	extent       float32
	speed        float32
	hertz        float32
	dampingRatio float32
}

func CreateOverlapRecovery(context *SampleContext) Sampler {
	s := &OverlapRecovery{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.5}
		s.context.camera.zoom = 3.75
	}

	s.bodyIDs = nil
	s.baseCount = 4
	s.overlap = 0.25
	s.extent = 0.5
	s.speed = 3.0
	s.hertz = 30.0
	s.dampingRatio = 10.0

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)
	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{Point1: box2d.Vec2{X: -40.0, Y: 0.0}, Point2: box2d.Vec2{X: 40.0, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.CreateScene()
	return s
}

func (s *OverlapRecovery) CreateScene() {
	for _, bodyID := range s.bodyIDs {
		bodyID.Destroy()
	}

	s.worldID.SetContactTuning(s.hertz, s.dampingRatio, s.speed)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	box := box2d.MakeBox(s.extent, s.extent)
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0

	baseCount := int(s.baseCount)
	s.bodyIDs = make([]box2d.Body, 0, baseCount*(baseCount+1)/2)

	fraction := 1.0 - s.overlap
	y := s.extent
	for i := range baseCount {
		x := fraction * s.extent * float32(i-baseCount)
		for j := i; j < baseCount; j++ {
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)

			bodyID.CreatePolygonShape(&shapeDef, &box)

			s.bodyIDs = append(s.bodyIDs, bodyID)

			x += 2.0 * fraction * s.extent
		}

		y += 2.0 * fraction * s.extent
	}
}

func (s *OverlapRecovery) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false
	changed = changed || imgui.SliderFloatV("Extent", &s.extent, 0.1, 1.0, "%.1f", 0)
	changed = changed || imgui.SliderInt("Base Count", &s.baseCount, 1, 10)
	changed = changed || imgui.SliderFloatV("Overlap", &s.overlap, 0.0, 1.0, "%.2f", 0)
	changed = changed || imgui.SliderFloatV("Speed", &s.speed, 0.0, 10.0, "%.1f", 0)
	changed = changed || imgui.SliderFloatV("Hertz", &s.hertz, 0.0, 240.0, "%.f", 0)
	changed = changed || imgui.SliderFloatV("Damping Ratio", &s.dampingRatio, 0.0, 20.0, "%.1f", 0)
	changed = changed || imgui.Button("Reset Scene")

	if changed {
		s.CreateScene()
	}

	imgui.PopItemWidth()

	return true
}

// A pyramid of 5cm squares. Stacking tiny objects is challenging for physics engines due to rotational effects.
// This is also challenging for Box2D because of the AABB margin and linear slop are close to the shape size. This
// leads to many collision pairs and some shape overlap.
type TinyPyramid struct {
	Sample

	extent float32
}

func CreateTinyPyramid(context *SampleContext) Sampler {
	s := &TinyPyramid{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.8}
		s.context.camera.zoom = 1.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(5.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		s.extent = 0.025
		baseCount := 30

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shapeDef := box2d.DefaultShapeDef()

		box := box2d.MakeSquare(s.extent)

		for i := range baseCount {
			y := (2.0*float32(i) + 1.0) * s.extent

			for j := i; j < baseCount; j++ {
				x := (float32(i)+1.0)*s.extent + 2.0*float32(j-i)*s.extent - float32(baseCount)*s.extent
				bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}

				bodyID := s.worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &box)
			}
		}
	}
	return s
}

func (s *TinyPyramid) Step() {
	s.DrawScreenTextLine("%.1fcm squares", 200.0*s.extent)
	s.Sample.Step()
}

// High gravity and high mass ratio. This shows how to tune contact and joint stiffness values to
// achieve an improved result at a high sub-step count.
// There is still a fair bit of bounce with some settings.
type Cart struct {
	Sample

	chassisID box2d.Body
	wheelID1  box2d.Body
	wheelID2  box2d.Body
	jointID1  box2d.RevoluteJoint
	jointID2  box2d.RevoluteJoint

	contactHertz           float32
	contactDampingRatio    float32
	contactSpeed           float32
	constraintHertz        float32
	constraintDampingRatio float32
}

func CreateCart(context *SampleContext) Sampler {
	s := &Cart{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.0}
		s.context.camera.zoom = 1.5
		s.context.subStepCount = 12
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		groundBox := box2d.MakeBox(20.0, 1.0)
		groundID.CreatePolygonShape(&shapeDef, &groundBox)
	}

	s.worldID.SetGravity(box2d.Vec2{X: 0, Y: -22.0})

	s.contactHertz = 240.0
	s.contactDampingRatio = 10.0
	s.contactSpeed = 0.5
	s.worldID.SetContactTuning(s.contactHertz, s.contactDampingRatio, s.contactSpeed)

	s.constraintHertz = 240.0
	s.constraintDampingRatio = 0.0

	s.chassisID = box2d.Body{}
	s.wheelID1 = box2d.Body{}
	s.wheelID2 = box2d.Body{}
	s.jointID1 = box2d.RevoluteJoint{}
	s.jointID2 = box2d.RevoluteJoint{}

	s.CreateScene()
	return s
}

func (s *Cart) CreateScene() {
	if s.chassisID != (box2d.Body{}) {
		s.chassisID.Destroy()
	}

	if s.wheelID1 != (box2d.Body{}) {
		s.wheelID1.Destroy()
	}

	if s.wheelID2 != (box2d.Body{}) {
		s.wheelID2.Destroy()
	}

	yBase := posScalar(2.0)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: yBase}
	s.chassisID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1000.0

	box := box2d.MakeOffsetBox(1.0, 0.25, box2d.Vec2{X: 0.0, Y: 0.25}, box2d.RotIdentity)
	s.chassisID.CreatePolygonShape(&shapeDef, &box)

	shapeDef = box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.02
	shapeDef.Density = 50.0

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: 0.1}
	bodyDef.Position = box2d.Pos{X: -0.9, Y: yBase - 0.15}
	s.wheelID1 = s.worldID.CreateBody(&bodyDef)
	s.wheelID1.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = box2d.Pos{X: 0.9, Y: yBase - 0.15}
	s.wheelID2 = s.worldID.CreateBody(&bodyDef)
	s.wheelID2.CreateCircleShape(&shapeDef, &circle)

	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.ConstraintHertz = 120.0
	jointDef.Base.ConstraintDampingRatio = 0.0

	jointDef.Base.BodyA = s.chassisID
	jointDef.Base.BodyB = s.wheelID1
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: -0.9, Y: -0.15}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 0.0}

	s.jointID1 = s.worldID.CreateRevoluteJoint(&jointDef)
	s.jointID1.SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)

	jointDef.Base.BodyA = s.chassisID
	jointDef.Base.BodyB = s.wheelID2
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.9, Y: -0.15}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 0.0}

	s.jointID2 = s.worldID.CreateRevoluteJoint(&jointDef)
	s.jointID2.SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)
}

func (s *Cart) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false
	imgui.Text("Contact")
	changed = changed || imgui.SliderFloatV("Hertz##contact", &s.contactHertz, 0.0, 240.0, "%.f", 0)
	changed = changed || imgui.SliderFloatV("Damping Ratio##contact", &s.contactDampingRatio, 0.0, 100.0, "%.f", 0)
	changed = changed || imgui.SliderFloatV("Speed", &s.contactSpeed, 0.0, 5.0, "%.1f", 0)

	if changed {
		s.worldID.SetContactTuning(s.contactHertz, s.contactDampingRatio, s.contactSpeed)
		s.CreateScene()
	}

	imgui.Separator()

	changed = false
	imgui.Text("Joint")
	changed = changed || imgui.SliderFloatV("Hertz##joint", &s.constraintHertz, 0.0, 240.0, "%.f", 0)
	changed = changed || imgui.SliderFloatV("Damping Ratio##joint", &s.constraintDampingRatio, 0.0, 20.0, "%.f", 0)

	imgui.Separator()

	changed = changed || imgui.Button("Reset Scene")

	if changed {
		s.jointID1.SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)
		s.jointID2.SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)
		s.CreateScene()
	}

	imgui.PopItemWidth()

	return true
}

// Ensure prismatic joint stability when highly distorted
type MultiplePrismatic struct {
	Sample
}

func CreateMultiplePrismatic(context *SampleContext) Sampler {
	s := &MultiplePrismatic{}
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

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	jointDef := box2d.DefaultPrismaticJointDef()
	jointDef.Base.BodyA = groundID
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.0}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: -0.6}
	jointDef.Base.DrawScale = 1.0
	jointDef.Base.ConstraintHertz = 240.0
	jointDef.LowerTranslation = -6.0
	jointDef.UpperTranslation = 6.0
	jointDef.EnableLimit = true

	for i := range 6 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(0.6 + 1.2*float32(i))}
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreatePolygonShape(&shapeDef, &box)

		jointDef.Base.BodyB = bodyID
		s.worldID.CreatePrismaticJoint(&jointDef)

		jointDef.Base.BodyA = bodyID
		jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.6}
	}

	// Increase the mouse force
	s.mouseForceScale = 100000.0
	return s
}

func init() {
	RegisterSample("Robustness", "HighMassRatio1", CreateHighMassRatio1)
	RegisterSample("Robustness", "HighMassRatio2", CreateHighMassRatio2)
	RegisterSample("Robustness", "HighMassRatio3", CreateHighMassRatio3)
	RegisterSample("Robustness", "Overlap Recovery", CreateOverlapRecovery)
	RegisterSample("Robustness", "Tiny Pyramid", CreateTinyPyramid)
	RegisterSample("Robustness", "Cart", CreateCart)
	RegisterSample("Robustness", "Multiple Prismatic", CreateMultiplePrismatic)
}
