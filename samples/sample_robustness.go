// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
)

// Pyramid with heavy box on top.
type HighMassRatio1 struct {
	Sample
}

// Big box on small boxes.
type HighMassRatio2 struct {
	Sample
}

// Big box on small triangles.
type HighMassRatio3 struct {
	Sample
}

type OverlapRecovery struct {
	Sample

	bodyIDs      []box2d.Body //exhaustruct:optional
	baseCount    int32
	overlap      float32
	extent       float32
	speed        float32
	hertz        float32
	dampingRatio float32
}

// A pyramid of 5cm squares. Stacking tiny objects is challenging for physics engines due to rotational effects.
// This is also challenging for Box2D because of the AABB margin and linear slop are close to the shape size. This
// leads to many collision pairs and some shape overlap.
type TinyPyramid struct {
	Sample

	extent float32
}

// High gravity and high mass ratio. This shows how to tune contact and joint stiffness values to
// achieve an improved result at a high sub-step count.
// There is still a fair bit of bounce with some settings.
type Cart struct {
	Sample

	chassisID box2d.Body          //exhaustruct:optional
	wheelID1  box2d.Body          //exhaustruct:optional
	wheelID2  box2d.Body          //exhaustruct:optional
	jointID1  box2d.RevoluteJoint //exhaustruct:optional
	jointID2  box2d.RevoluteJoint //exhaustruct:optional

	contactHertz           float32
	contactDampingRatio    float32
	contactSpeed           float32
	constraintHertz        float32
	constraintDampingRatio float32
}

// Ensure prismatic joint stability when highly distorted.
type MultiplePrismatic struct {
	Sample
}

func CreateHighMassRatio1(context *SampleContext) Sampler {
	const (
		cameraCenterX     = 3.0
		cameraCenterY     = 14.0
		groundHalfWidth   = 50.0
		boxDiameterFactor = 2.0
		topBoxClearance   = 2.0
		rowCenterFactor   = 0.5
		topDensityScale   = 100.0
	)

	s := &HighMassRatio1{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = 25.0
	}

	extent := float32(1.0)

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			groundHalfWidth,
			1.0,
			box2d.Vec2{X: 0.0, Y: -1.0},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		box := box2d.MakeBox(extent, extent)
		shapeDef := box2d.DefaultShapeDef()

		for j := range 3 {
			count := 10
			offset := -20.0*extent + boxDiameterFactor*(float32(count)+1.0)*extent*float32(j)
			y := extent

			for count > 0 {
				for i := range count {
					coeff := float32(i) - rowCenterFactor*float32(count)

					yy := y
					if count == 1 {
						yy = y + topBoxClearance
					}

					bodyDef.Position = box2d.Pos{
						X: posScalar(boxDiameterFactor*coeff*extent + offset),
						Y: posScalar(yy),
					}
					bodyID := s.worldID.CreateBody(&bodyDef)

					shapeDef.Density = 1.0
					if count == 1 {
						shapeDef.Density = (float32(j) + 1.0) * topDensityScale
					}

					bodyID.CreatePolygonShape(&shapeDef, &box)
				}

				count--
				y += boxDiameterFactor * extent
			}
		}
	}

	return s
}

func CreateHighMassRatio2(context *SampleContext) Sampler {
	const (
		cameraCenterY            = 16.5
		groundHalfWidth          = 50.0
		supportHalfExtentFactor  = 0.5
		largeBoxHalfExtentFactor = 10.0
		supportOffsetFactor      = 9.0
		dropClearanceFactor      = 16.0
	)

	s := &HighMassRatio2{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 25.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			groundHalfWidth,
			1.0,
			box2d.Vec2{X: 0.0, Y: -1.0},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		shapeDef := box2d.DefaultShapeDef()

		extent := float32(1.0)
		smallBox := box2d.MakeBox(supportHalfExtentFactor*extent, supportHalfExtentFactor*extent)
		bigBox := box2d.MakeBox(largeBoxHalfExtentFactor*extent, largeBoxHalfExtentFactor*extent)

		{
			bodyDef.Position = box2d.Pos{
				X: posScalar(-supportOffsetFactor * extent),
				Y: posScalar(supportHalfExtentFactor * extent),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallBox)
		}

		{
			bodyDef.Position = box2d.Pos{
				X: posScalar(supportOffsetFactor * extent),
				Y: posScalar(supportHalfExtentFactor * extent),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallBox)
		}

		{
			bodyDef.Position = box2d.Pos{
				X: 0.0,
				Y: posScalar((largeBoxHalfExtentFactor + dropClearanceFactor) * extent),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &bigBox)
		}
	}

	return s
}

func CreateHighMassRatio3(context *SampleContext) Sampler {
	const (
		cameraCenterY            = 16.5
		groundHalfWidth          = 50.0
		supportHalfExtentFactor  = 0.5
		largeBoxHalfExtentFactor = 10.0
		supportOffsetFactor      = 9.0
		dropClearanceFactor      = 4.0
	)

	s := &HighMassRatio3{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 25.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			groundHalfWidth,
			1.0,
			box2d.Vec2{X: 0.0, Y: -1.0},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		shapeDef := box2d.DefaultShapeDef()

		extent := float32(1.0)
		points := []box2d.Vec2{
			{X: -supportHalfExtentFactor * extent, Y: 0.0},
			{X: supportHalfExtentFactor * extent, Y: 0.0},
			{X: 0.0, Y: 1.0 * extent},
		}
		hull := box2d.ComputeHull(points)
		smallTriangle := box2d.MakePolygon(&hull, 0.0)
		bigBox := box2d.MakeBox(largeBoxHalfExtentFactor*extent, largeBoxHalfExtentFactor*extent)

		{
			bodyDef.Position = box2d.Pos{
				X: posScalar(-supportOffsetFactor * extent),
				Y: posScalar(supportHalfExtentFactor * extent),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallTriangle)
		}

		{
			bodyDef.Position = box2d.Pos{
				X: posScalar(supportOffsetFactor * extent),
				Y: posScalar(supportHalfExtentFactor * extent),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &smallTriangle)
		}

		{
			bodyDef.Position = box2d.Pos{
				X: 0.0,
				Y: posScalar((largeBoxHalfExtentFactor + dropClearanceFactor) * extent),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &bigBox)
		}
	}

	return s
}

func CreateOverlapRecovery(context *SampleContext) Sampler {
	const (
		cameraCenterY       = 2.5
		groundHalfWidth     = 40.0
		initialBaseCount    = 4
		initialOverlap      = 0.25
		initialExtent       = 0.5
		initialSpeed        = 3.0
		initialHertz        = 30.0
		initialDampingRatio = 10.0
	)

	s := &OverlapRecovery{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		baseCount:       initialBaseCount,
		overlap:         initialOverlap,
		extent:          initialExtent,
		speed:           initialSpeed,
		hertz:           initialHertz,
		dampingRatio:    initialDampingRatio,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 3.75
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)
	shapeDef := box2d.DefaultShapeDef()
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
		Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.CreateScene()

	return s
}

func (s *OverlapRecovery) CreateScene() {
	const (
		triangleCountDivisor = 2
		boxDiameterFactor    = 2.0
	)

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
	s.bodyIDs = make([]box2d.Body, 0, baseCount*(baseCount+1)/triangleCountDivisor)

	fraction := 1.0 - s.overlap

	y := s.extent
	for i := range baseCount {
		x := fraction * s.extent * float32(i-baseCount)
		for j := i; j < baseCount; j++ {
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)

			bodyID.CreatePolygonShape(&shapeDef, &box)

			s.bodyIDs = append(s.bodyIDs, bodyID)

			x += boxDiameterFactor * fraction * s.extent
		}

		y += boxDiameterFactor * fraction * s.extent
	}
}

func (s *OverlapRecovery) DrawControls() bool {
	const (
		controlWidthEm  = 6.0
		minExtent       = 0.1
		maxBaseCount    = 10
		maxSpeed        = 10.0
		maxHertz        = 240.0
		maxDampingRatio = 20.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	changed := false
	changed = changed || imgui.SliderFloatV("Extent", &s.extent, minExtent, 1.0, "%.1f", 0)
	changed = changed || imgui.SliderInt("Base Count", &s.baseCount, 1, maxBaseCount)
	changed = changed || imgui.SliderFloatV("Overlap", &s.overlap, 0.0, 1.0, "%.2f", 0)
	changed = changed || imgui.SliderFloatV("Speed", &s.speed, 0.0, maxSpeed, "%.1f", 0)
	changed = changed || imgui.SliderFloatV("Hertz", &s.hertz, 0.0, maxHertz, "%.f", 0)
	changed = changed ||
		imgui.SliderFloatV("Damping Ratio", &s.dampingRatio, 0.0, maxDampingRatio, "%.1f", 0)
	changed = changed || imgui.Button("Reset Scene")

	if changed {
		s.CreateScene()
	}

	imgui.PopItemWidth()

	return true
}

func CreateTinyPyramid(context *SampleContext) Sampler {
	const (
		cameraCenterY   = 0.8
		groundHalfWidth = 5.0
		initialExtent   = 0.025
	)

	s := &TinyPyramid{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		extent:          initialExtent,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 1.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			groundHalfWidth,
			1.0,
			box2d.Vec2{X: 0.0, Y: -1.0},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		baseCount := 30

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		shapeDef := box2d.DefaultShapeDef()

		box := box2d.MakeSquare(s.extent)

		for i := range baseCount {
			y := (2.0*float32(i) + 1.0) * s.extent

			for j := i; j < baseCount; j++ {
				x := (float32(i)+1.0)*s.extent + 2.0*float32(
					j-i,
				)*s.extent - float32(
					baseCount,
				)*s.extent
				bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}

				bodyID := s.worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &box)
			}
		}
	}

	return s
}

func (s *TinyPyramid) Step() {
	const (
		diameterCentimetersPerMeter = 200.0
	)

	s.DrawScreenTextLinef("%.1fcm squares", diameterCentimetersPerMeter*s.extent)
	s.Sample.Step()
}

func CreateCart(context *SampleContext) Sampler {
	const (
		groundHalfWidth            = 20.0
		initialContactHertz        = 240.0
		initialContactDampingRatio = 10.0
		initialContactSpeed        = 0.5
		initialConstraintHertz     = 240.0
	)

	s := &Cart{
		context:                context,
		camera:                 &context.camera,
		draw:                   context.draw,
		screenTextX:            initialScreenTextX,
		mouseForceScale:        initialMouseForceScale,
		contactHertz:           initialContactHertz,
		contactDampingRatio:    initialContactDampingRatio,
		contactSpeed:           initialContactSpeed,
		constraintHertz:        initialConstraintHertz,
		constraintDampingRatio: 0.0,
	}
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
		groundBox := box2d.MakeBox(groundHalfWidth, 1.0)
		groundID.CreatePolygonShape(&shapeDef, &groundBox)
	}

	s.worldID.SetGravity(box2d.Vec2{X: 0, Y: -22.0})

	s.worldID.SetContactTuning(s.contactHertz, s.contactDampingRatio, s.contactSpeed)

	s.CreateScene()

	return s
}

func (s *Cart) CreateScene() {
	const (
		chassisHeight     = 2.0
		chassisHalfHeight = 0.25
		wheelRadius       = 0.1
		wheelOffset       = 0.9
		wheelDrop         = 0.15
	)

	if s.chassisID != (box2d.Body{}) {
		s.chassisID.Destroy()
	}

	if s.wheelID1 != (box2d.Body{}) {
		s.wheelID1.Destroy()
	}

	if s.wheelID2 != (box2d.Body{}) {
		s.wheelID2.Destroy()
	}

	yBase := posScalar(chassisHeight)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: yBase}
	s.chassisID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1000.0

	box := box2d.MakeOffsetBox(
		1.0,
		chassisHalfHeight,
		box2d.Vec2{X: 0.0, Y: chassisHalfHeight},
		box2d.RotIdentity,
	)
	s.chassisID.CreatePolygonShape(&shapeDef, &box)

	shapeDef = box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.02
	shapeDef.Density = 50.0

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: wheelRadius}
	bodyDef.Position = box2d.Pos{X: -wheelOffset, Y: yBase - wheelDrop}
	s.wheelID1 = s.worldID.CreateBody(&bodyDef)
	s.wheelID1.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = box2d.Pos{X: wheelOffset, Y: yBase - wheelDrop}
	s.wheelID2 = s.worldID.CreateBody(&bodyDef)
	s.wheelID2.CreateCircleShape(&shapeDef, &circle)

	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.ConstraintHertz = 120.0
	jointDef.Base.ConstraintDampingRatio = 0.0

	jointDef.Base.BodyA = s.chassisID
	jointDef.Base.BodyB = s.wheelID1
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: -wheelOffset, Y: -wheelDrop}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 0.0}

	s.jointID1 = s.worldID.CreateRevoluteJoint(&jointDef)
	s.jointID1.SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)

	jointDef.Base.BodyA = s.chassisID
	jointDef.Base.BodyB = s.wheelID2
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: wheelOffset, Y: -wheelDrop}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 0.0}

	s.jointID2 = s.worldID.CreateRevoluteJoint(&jointDef)
	s.jointID2.SetConstraintTuning(s.constraintHertz, s.constraintDampingRatio)
}

func (s *Cart) DrawControls() bool {
	const (
		controlWidthEm            = 6.0
		maxHertz                  = 240.0
		maxContactDampingRatio    = 100.0
		maxContactSpeed           = 5.0
		maxConstraintDampingRatio = 20.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	changed := false

	imgui.Text("Contact")

	changed = changed ||
		imgui.SliderFloatV("Hertz##contact", &s.contactHertz, 0.0, maxHertz, "%.f", 0)
	changed = changed ||
		imgui.SliderFloatV(
			"Damping Ratio##contact",
			&s.contactDampingRatio,
			0.0,
			maxContactDampingRatio,
			"%.f",
			0,
		)
	changed = changed ||
		imgui.SliderFloatV("Speed", &s.contactSpeed, 0.0, maxContactSpeed, "%.1f", 0)

	if changed {
		s.worldID.SetContactTuning(s.contactHertz, s.contactDampingRatio, s.contactSpeed)
		s.CreateScene()
	}

	imgui.Separator()

	changed = false

	imgui.Text("Joint")

	changed = changed ||
		imgui.SliderFloatV("Hertz##joint", &s.constraintHertz, 0.0, maxHertz, "%.f", 0)
	changed = changed ||
		imgui.SliderFloatV(
			"Damping Ratio##joint",
			&s.constraintDampingRatio,
			0.0,
			maxConstraintDampingRatio,
			"%.f",
			0,
		)

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

func CreateMultiplePrismatic(context *SampleContext) Sampler {
	const (
		cameraCenterY     = 8.0
		baseCameraZoom    = 25.0
		cameraZoomScale   = 0.5
		jointAnchorOffset = 0.6
		translationLimit  = 6.0
	)

	s := &MultiplePrismatic{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = baseCameraZoom * cameraZoomScale
	}

	var groundID box2d.Body
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID = s.worldID.CreateBody(&bodyDef)
	}

	shapeDef := box2d.DefaultShapeDef()

	const boxHalfExtent = 0.5

	box := box2d.MakeBox(boxHalfExtent, boxHalfExtent)
	jointDef := box2d.DefaultPrismaticJointDef()
	jointDef.Base.BodyA = groundID
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.0}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: -jointAnchorOffset}
	jointDef.Base.DrawScale = 1.0
	jointDef.Base.ConstraintHertz = 240.0
	jointDef.LowerTranslation = -translationLimit
	jointDef.UpperTranslation = translationLimit
	jointDef.EnableLimit = true

	for i := range 6 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(jointAnchorOffset + 1.2*float32(i))}
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreatePolygonShape(&shapeDef, &box)

		jointDef.Base.BodyB = bodyID
		s.worldID.CreatePrismaticJoint(&jointDef)

		jointDef.Base.BodyA = bodyID
		jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: jointAnchorOffset}
	}

	// Increase the mouse force
	s.mouseForceScale = 100000.0

	return s
}

func registerRobustnessSamples() {
	RegisterSample("Robustness", "HighMassRatio1", CreateHighMassRatio1)
	RegisterSample("Robustness", "HighMassRatio2", CreateHighMassRatio2)
	RegisterSample("Robustness", "HighMassRatio3", CreateHighMassRatio3)
	RegisterSample("Robustness", "Overlap Recovery", CreateOverlapRecovery)
	RegisterSample("Robustness", "Tiny Pyramid", CreateTinyPyramid)
	RegisterSample("Robustness", "Cart", CreateCart)
	RegisterSample("Robustness", "Multiple Prismatic", CreateMultiplePrismatic)
}
