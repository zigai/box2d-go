// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
)

const (
	varyingRestitutionCircleShape = 0
	varyingRestitutionBoxShape    = 1
	varyingRestitutionCount       = 40
)

const (
	circleStackImpactSpeed   = 5.0
	circleStackMaxStackCount = 3
)

const (
	restitutionPropagationColumnCount = 3
	restitutionPropagationDropHeight  = 5.0
)

const boxRestitutionImpactSpeed = 5.0

const (
	rotatedBoxRestitutionCount      = 12
	rotatedBoxRestitutionDropHeight = 10.0
)

const circleRestitutionCount = 10

var restitutionPropagationLabels = [restitutionPropagationColumnCount]string{
	"ground",
	"one crate",
	"two crates",
}

// Restitution is approximate since Box2D uses speculative collision.
type VaryingRestitution struct {
	Sample

	bodyIDs   [varyingRestitutionCount]box2d.Body //exhaustruct:optional
	shapeType int32
}

// Tests how a single bouncing box behaves.
type SingleBoxRestitution struct {
	Sample

	bodyID      box2d.Body //exhaustruct:optional
	maxY        float32
	startEnergy float32 //exhaustruct:optional
	peakEnergy  float32 //exhaustruct:optional
	height      float32
}

type SingleCircleRestitution struct {
	Sample

	bodyID      box2d.Body //exhaustruct:optional
	maxY        float32
	startEnergy float32 //exhaustruct:optional
	peakEnergy  float32 //exhaustruct:optional
}

// Similar to MeasureSupportedBounce unit test.
type CircleStackRestitution struct {
	Sample

	bodyIDs     [circleStackMaxStackCount + 1]box2d.Body //exhaustruct:optional
	impactorID  box2d.Body                               //exhaustruct:optional
	bodyCount   int                                      //exhaustruct:optional
	stackCount  int32
	restitution float32
	startHeight float32 //exhaustruct:optional
	coefficient float32 //exhaustruct:optional
	startEnergy float32 //exhaustruct:optional
	peakEnergy  float32 //exhaustruct:optional
	gravity     bool
	latched     bool //exhaustruct:optional
	hit         bool //exhaustruct:optional
}

// A bouncy ball landing on a resting body. Three balls fall together from the same height onto bare
// ground, one crate and two crates. Without restitution propagation the restitution stage only
// visits the ball's own contact, so the rebound is shared with the crate, which is driven into the
// ground and pushed back the next step. With propagation the ground contact under the crate can react
// in the restitution sover.
type RestitutionPropagation struct {
	Sample

	bodyIDs     [2 * restitutionPropagationColumnCount]box2d.Body //exhaustruct:optional
	ballIDs     [restitutionPropagationColumnCount]box2d.Body     //exhaustruct:optional
	crateIDs    [restitutionPropagationColumnCount]box2d.Body     //exhaustruct:optional
	restHeight  [restitutionPropagationColumnCount]float32        //exhaustruct:optional
	apex        [restitutionPropagationColumnCount]float32        //exhaustruct:optional
	rebounding  [restitutionPropagationColumnCount]bool           //exhaustruct:optional
	done        [restitutionPropagationColumnCount]bool           //exhaustruct:optional
	bodyCount   int                                               //exhaustruct:optional
	crateCount  int                                               //exhaustruct:optional
	restitution float32
	kick        float32 //exhaustruct:optional
}

// Similar to MeasureFlatBounce and SpinTest unit tests.
type BoxRestitution struct {
	Sample

	boxID       box2d.Body //exhaustruct:optional
	restitution float32
	spin        float32
	coefficient float32 //exhaustruct:optional
	startEnergy float32 //exhaustruct:optional
	peakEnergy  float32 //exhaustruct:optional
	latched     bool    //exhaustruct:optional
	hit         bool    //exhaustruct:optional
}

type RotatedBoxRestitution struct {
	Sample

	bodyIDs      [rotatedBoxRestitutionCount]box2d.Body //exhaustruct:optional
	startEnergy  [rotatedBoxRestitutionCount]float32    //exhaustruct:optional
	restitution  float32
	friction     float32
	safetyFactor float32
}

// Based on https://github.com/erincatto/box2d/discussions/957
type TwoBoxRestitution struct {
	Sample

	bodyIDs     [2]box2d.Body //exhaustruct:optional
	startEnergy float32       //exhaustruct:optional
	peakEnergy  float32       //exhaustruct:optional
}

type CircleRestitution struct {
	Sample
}

func CreateVaryingRestitution(context *SampleContext) Sampler {
	const (
		cameraCenterX = 4.0
		cameraCenterY = 17.0
	)

	s := &VaryingRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		shapeType:       varyingRestitutionCircleShape,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = 27.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		h := float32(1.0 * varyingRestitutionCount)
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -h, Y: 0.0},
			Point2: box2d.Vec2{X: h, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.CreateBodies()

	return s
}

func (s *VaryingRestitution) CreateBodies() {
	const (
		shapeHalfExtent = 0.5
		bodySpacing     = 2.0
		dropHeight      = 40.0
	)

	for i := range varyingRestitutionCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}
		}
	}

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: shapeHalfExtent}

	box := box2d.MakeBox(shapeHalfExtent, shapeHalfExtent)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 0.0
	shapeDef.Material.Friction = 0.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	dr := 1.0 / float32(max(varyingRestitutionCount-1, 1))
	x := float32(-1.0 * (varyingRestitutionCount - 1))
	dx := float32(bodySpacing)

	for i := range varyingRestitutionCount {
		bodyDef.Name = fmt.Sprintf("%.2f", shapeDef.Material.Restitution)
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: dropHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)

		s.bodyIDs[i] = bodyID

		if s.shapeType == varyingRestitutionCircleShape {
			bodyID.CreateCircleShape(&shapeDef, &circle)
		} else {
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}

		shapeDef.Material.Restitution += dr
		x += dx
	}
}

func (s *VaryingRestitution) DrawControls() bool {
	const (
		controlWidthEm = 6.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	changed := false
	shapeTypes := []string{"Circle", "Box"}

	changed = changed ||
		imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))

	imgui.PopItemWidth()

	changed = changed || imgui.Button("Reset")

	if changed {
		s.CreateBodies()
	}

	return true
}

func (s *VaryingRestitution) Step() {
	const (
		dropMarkerHeight = 40.5
	)

	s.Sample.Step()

	h := posScalar(1.0 * varyingRestitutionCount)
	s.draw.DrawLine(
		box2d.Pos{X: -h, Y: dropMarkerHeight},
		box2d.Pos{X: h, Y: dropMarkerHeight},
		box2d.ColorRed,
	)
}

func CreateSingleBoxRestitution(context *SampleContext) Sampler {
	const (
		initialHeight   = 5.0
		groundHalfWidth = 4.0
		boxHalfExtent   = 0.5
	)

	s := &SingleBoxRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		height:          initialHeight,
		maxY:            initialHeight,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: initialHeight}
		s.context.camera.zoom = 8.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	box := box2d.MakeBox(boxHalfExtent, boxHalfExtent)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.EnableHitEvents = true

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(s.height)}
	bodyDef.SafetyFactor = 0.01
	s.bodyID = s.worldID.CreateBody(&bodyDef)
	s.bodyID.CreatePolygonShape(&shapeDef, &box)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.bodyID})
	s.startEnergy = linear + angular + potential
	s.peakEnergy = s.startEnergy

	return s
}

func (s *SingleBoxRestitution) Step() {
	const (
		markerHalfWidth = 4.0
		boxHalfExtent   = 0.5
		percentScale    = 100.0
	)

	s.Sample.Step()

	s.draw.DrawLine(
		box2d.Pos{X: -markerHalfWidth, Y: posScalar(s.height + boxHalfExtent)},
		box2d.Pos{X: markerHalfWidth, Y: posScalar(s.height + boxHalfExtent)},
		box2d.ColorRed,
	)

	p := s.bodyID.Position()
	if hitEventCount(s.worldID) == 1 {
		s.maxY = float32(p.Y)
	} else {
		s.maxY = max(s.maxY, float32(p.Y))
	}

	s.DrawScreenTextLinef("maxY = %.2f", s.maxY)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.bodyID})
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = percentScale / s.startEnergy
	}

	s.DrawScreenTextLinef("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLinef("potential = %.3f J", potential)
	s.DrawScreenTextLinef(
		"total     = %.3f J (%.2f%% of start, peak %.2f%%)",
		total,
		scale*total,
		scale*s.peakEnergy,
	)
}

func CreateSingleCircleRestitution(context *SampleContext) Sampler {
	const (
		cameraCenterY   = 5.0
		groundHalfWidth = 4.0
		circleRadius    = 0.5
		dropHeight      = 10.0
	)

	s := &SingleCircleRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		maxY:            dropHeight,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 8.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: circleRadius}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.EnableHitEvents = true

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: dropHeight}
	bodyDef.SafetyFactor = 0.01
	s.bodyID = s.worldID.CreateBody(&bodyDef)
	s.bodyID.CreateCircleShape(&shapeDef, &circle)

	s.maxY = float32(bodyDef.Position.Y)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.bodyID})
	s.startEnergy = linear + angular + potential
	s.peakEnergy = s.startEnergy

	return s
}

func (s *SingleCircleRestitution) Step() {
	const (
		markerHalfWidth  = 4.0
		dropMarkerHeight = 10.5
		percentScale     = 100.0
	)

	s.Sample.Step()

	p := s.bodyID.Position()
	if hitEventCount(s.worldID) == 1 {
		s.maxY = float32(p.Y)
	} else {
		s.maxY = max(s.maxY, float32(p.Y))
	}

	s.draw.DrawLine(
		box2d.Pos{X: -markerHalfWidth, Y: dropMarkerHeight},
		box2d.Pos{X: markerHalfWidth, Y: dropMarkerHeight},
		box2d.ColorRed,
	)

	s.DrawScreenTextLinef("maxY = %.2f", s.maxY)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.bodyID})
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = percentScale / s.startEnergy
	}

	s.DrawScreenTextLinef("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLinef("potential = %.3f J", potential)
	s.DrawScreenTextLinef(
		"total     = %.3f J (%.2f%% of start, peak %.2f%%)",
		total,
		scale*total,
		scale*s.peakEnergy,
	)
}

func CreateCircleStackRestitution(context *SampleContext) Sampler {
	const (
		initialStackCount = 2
		cameraCenterY     = 2.5
		groundHalfWidth   = 40.0
	)

	s := &CircleStackRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		stackCount:      initialStackCount,
		restitution:     1.0,
		gravity:         false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 6.0
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = 0.0
	box := box2d.MakeBox(groundHalfWidth, 1.0)
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.CreateScene()

	return s
}

func (s *CircleStackRestitution) CreateBall(
	y float32, velocityY float32, restitution float32,
	hitEvents bool,
) box2d.Body {
	const (
		ballRadius = 0.5
	)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(y)}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: velocityY}
	bodyDef.EnableSleep = false
	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = restitution
	shapeDef.EnableHitEvents = hitEvents
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}
	bodyID.CreateCircleShape(&shapeDef, &circle)

	return bodyID
}

func (s *CircleStackRestitution) CreateScene() {
	const (
		halfStepFraction = 0.5
		ballRadius       = 0.5
		stepsPerSecond   = 60.0
	)

	for i := range s.bodyCount {
		s.bodyIDs[i].Destroy()
	}

	s.bodyCount = 0

	if s.gravity {
		s.worldID.SetGravity(box2d.Vec2{X: 0.0, Y: -10.0})
	} else {
		s.worldID.SetGravity(box2d.Vec2Zero)
	}

	for i := range int(s.stackCount) {
		s.bodyIDs[s.bodyCount] = s.CreateBall(ballRadius+1.0*float32(i), 0.0, 0.0, false)
		s.bodyCount++
	}

	// Start half a step of travel above contact so the impact lands mid step, matching the test
	s.startHeight = ballRadius + 1.0*float32(
		s.stackCount,
	) + halfStepFraction*circleStackImpactSpeed*(1.0/stepsPerSecond)

	s.impactorID = s.CreateBall(s.startHeight, -circleStackImpactSpeed, s.restitution, true)
	s.bodyIDs[s.bodyCount] = s.impactorID
	s.bodyCount++

	linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[:s.bodyCount])
	s.startEnergy = linear + angular + potential
	s.peakEnergy = s.startEnergy

	s.coefficient = 0.0
	s.latched = false
	s.hit = false
}

func (s *CircleStackRestitution) DrawControls() bool {
	const (
		controlWidthEm = 6.0
	)

	rebuild := false

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.SliderFloatV("Restitution", &s.restitution, 0.0, 1.0, "%.2f", 0) {
		rebuild = true
	}

	if imgui.SliderInt("Stack", &s.stackCount, 0, circleStackMaxStackCount) {
		rebuild = true
	}

	imgui.PopItemWidth()

	if imgui.Checkbox("Gravity", &s.gravity) {
		rebuild = true
	}

	if imgui.Button("Reset") {
		rebuild = true
	}

	if rebuild {
		s.CreateScene()
	}

	return true
}

func (s *CircleStackRestitution) Step() {
	const (
		markerHalfWidth = 2.0
		percentScale    = 100.0
	)

	s.Sample.Step()

	if hitEventCount(s.worldID) > 0 {
		s.hit = true
	}

	vy := s.impactorID.LinearVelocity().Y

	// The rebound only means something once the impact has happened and the impactor is leaving
	if !s.latched && s.hit && vy > 0.0 {
		s.coefficient = vy / circleStackImpactSpeed
		s.latched = true
	}

	supportSpeed := float32(0.0)
	for i := range int(s.stackCount) {
		supportSpeed = max(supportSpeed, s.bodyIDs[i].LinearVelocity().Length())
	}

	s.draw.DrawLine(
		box2d.Pos{X: -markerHalfWidth, Y: posScalar(s.startHeight)},
		box2d.Pos{X: markerHalfWidth, Y: posScalar(s.startHeight)},
		box2d.ColorRed,
	)

	linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[:s.bodyCount])
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	s.DrawScreenTextLinef("impactor vy = %.3f m/s", vy)

	s.DrawScreenTextLinef(
		"coefficient = %.4f live (target %.2f)",
		vy/circleStackImpactSpeed,
		s.restitution,
	)

	if s.latched {
		s.DrawScreenTextLinef("coefficient = %.4f at first rebound", s.coefficient)
	}

	s.DrawScreenTextLinef("support speed = %.4f m/s", supportSpeed)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = percentScale / s.startEnergy
	}

	s.DrawScreenTextLinef("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLinef("potential = %.3f J", potential)
	s.DrawScreenTextLinef(
		"total     = %.3f J (%.2f%% of start, peak %.2f%%)",
		total,
		scale*total,
		scale*s.peakEnergy,
	)
}

func CreateRestitutionPropagation(context *SampleContext) Sampler {
	const (
		initialRestitution = 0.9
		cameraCenterY      = 4.0
		groundHalfWidth    = 40.0
	)

	s := &RestitutionPropagation{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		restitution:     initialRestitution,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 10.0
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.6
	box := box2d.MakeBox(groundHalfWidth, 1.0)
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.CreateScene()

	return s
}

func (s *RestitutionPropagation) CreateScene() {
	const (
		bodyHalfExtent = 0.5
		columnSpacing  = 4.0
	)

	for i := range s.bodyCount {
		s.bodyIDs[i].Destroy()
	}

	s.bodyCount = 0
	s.crateCount = 0
	s.kick = 0.0

	box := box2d.MakeBox(bodyHalfExtent, bodyHalfExtent)
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: bodyHalfExtent}

	for column := range restitutionPropagationColumnCount {
		x := columnSpacing * float32(column-1)

		for k := range column {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{
				X: posScalar(x),
				Y: posScalar(bodyHalfExtent + 1.0*float32(k)),
			}
			bodyDef.EnableSleep = false
			crateID := s.worldID.CreateBody(&bodyDef)

			shapeDef := box2d.DefaultShapeDef()
			crateID.CreatePolygonShape(&shapeDef, &box)

			s.bodyIDs[s.bodyCount] = crateID
			s.bodyCount++
			s.crateIDs[s.crateCount] = crateID
			s.crateCount++
		}

		s.restHeight[column] = bodyHalfExtent + 1.0*float32(column)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{
			X: posScalar(x),
			Y: posScalar(s.restHeight[column] + restitutionPropagationDropHeight),
		}
		bodyDef.EnableSleep = false
		ballID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Restitution = s.restitution
		ballID.CreateCircleShape(&shapeDef, &circle)

		s.bodyIDs[s.bodyCount] = ballID
		s.bodyCount++
		s.ballIDs[column] = ballID
		s.apex[column] = 0.0
		s.rebounding[column] = false
		s.done[column] = false
	}
}

func (s *RestitutionPropagation) DrawControls() bool {
	const (
		controlWidthEm = 6.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	rebuild := imgui.SliderFloatV("Restitution", &s.restitution, 0.0, 1.0, "%.2f", 0)

	imgui.PopItemWidth()

	if imgui.Button("Reset") {
		rebuild = true
	}

	if rebuild {
		s.CreateScene()
	}

	return true
}

func (s *RestitutionPropagation) Step() {
	const (
		columnSpacing = 4.0
		percentScale  = 100.0
	)

	s.Sample.Step()

	for column := range restitutionPropagationColumnCount {
		v := s.ballIDs[column].LinearVelocity()
		height := float32(s.ballIDs[column].Position().Y) - s.restHeight[column]

		// The first rebound apex: the ball starts at rest and falls, so upward velocity means it bounced
		if !s.rebounding[column] && v.Y > 0.0 {
			s.rebounding[column] = true
		}

		if s.rebounding[column] && !s.done[column] {
			s.apex[column] = max(s.apex[column], height)
			if v.Y < 0.0 {
				s.done[column] = true
			}
		}

		x := posScalar(columnSpacing * float32(column-1))
		dropY := posScalar(s.restHeight[column] + restitutionPropagationDropHeight)
		s.draw.DrawLine(
			box2d.Pos{X: x - 1.0, Y: dropY},
			box2d.Pos{X: x + 1.0, Y: dropY},
			box2d.ColorRed,
		)

		if s.apex[column] > 0.0 {
			apexY := posScalar(s.restHeight[column] + s.apex[column])
			s.draw.DrawLine(
				box2d.Pos{X: x - 1.0, Y: apexY},
				box2d.Pos{X: x + 1.0, Y: apexY},
				box2d.ColorGreen,
			)
		}

		s.DrawScreenTextLinef(
			"%s: apex %.2f m (%.0f%% of drop)",
			restitutionPropagationLabels[column],
			s.apex[column],
			percentScale*s.apex[column]/restitutionPropagationDropHeight,
		)
	}

	for i := range s.crateCount {
		s.kick = max(s.kick, s.crateIDs[i].LinearVelocity().Length())
	}

	s.DrawScreenTextLinef("peak crate speed %.2f m/s", s.kick)
}

func CreateBoxRestitution(context *SampleContext) Sampler {
	const (
		initialRestitution = 0.9
		cameraCenterY      = 8.0
		groundHalfWidth    = 40.0
	)

	s := &BoxRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		restitution:     initialRestitution,
		spin:            0.0,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 10.0
	}

	// Gravity would bias the measured coefficient
	s.worldID.SetGravity(box2d.Vec2Zero)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = 0.0
	box := box2d.MakeBox(groundHalfWidth, 1.0)
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.CreateScene()

	return s
}

func (s *BoxRestitution) CreateScene() {
	const (
		boxHalfHeight    = 0.25
		halfStepFraction = 0.5
		stepsPerSecond   = 60.0
	)

	if s.boxID != (box2d.Body{}) {
		s.boxID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{
		X: 0.0,
		Y: posScalar(
			float32(
				boxHalfHeight + halfStepFraction*boxRestitutionImpactSpeed*(1.0/stepsPerSecond),
			),
		),
	}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -boxRestitutionImpactSpeed}
	bodyDef.AngularVelocity = s.spin
	bodyDef.EnableSleep = false
	s.boxID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = s.restitution
	shapeDef.EnableHitEvents = true
	box := box2d.MakeBox(1.0, boxHalfHeight)
	s.boxID.CreatePolygonShape(&shapeDef, &box)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.boxID})
	s.startEnergy = linear + angular + potential
	s.peakEnergy = s.startEnergy

	s.coefficient = 0.0
	s.latched = false
	s.hit = false
}

func (s *BoxRestitution) DrawControls() bool {
	const (
		controlWidthEm = 6.0
		maxSpin        = 2.0
	)

	rebuild := false

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.SliderFloatV("Restitution", &s.restitution, 0.0, 1.0, "%.2f", 0) {
		rebuild = true
	}

	if imgui.SliderFloatV("Spin", &s.spin, 0.0, maxSpin, "%.2f", 0) {
		rebuild = true
	}

	imgui.PopItemWidth()

	if imgui.Button("Reset") {
		rebuild = true
	}

	if rebuild {
		s.CreateScene()
	}

	return true
}

func (s *BoxRestitution) Step() {
	const (
		cornerPointSize = 8.0
		percentScale    = 100.0
	)

	s.Sample.Step()

	if hitEventCount(s.worldID) > 0 {
		s.hit = true
	}

	vy := s.boxID.LinearVelocity().Y
	w := s.boxID.AngularVelocity()

	if !s.latched && s.hit && vy > 0.0 {
		s.coefficient = vy / boxRestitutionImpactSpeed
		s.latched = true
	}

	s.draw.DrawPoint(
		s.boxID.WorldPoint(box2d.Vec2{X: -1.0, Y: -0.25}),
		cornerPointSize,
		box2d.ColorYellow,
	)
	s.draw.DrawPoint(
		s.boxID.WorldPoint(box2d.Vec2{X: 1.0, Y: -0.25}),
		cornerPointSize,
		box2d.ColorYellow,
	)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.boxID})
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	s.DrawScreenTextLinef("vy = %.3f m/s", vy)

	s.DrawScreenTextLinef(
		"coefficient = %.4f live (target %.2f)",
		vy/boxRestitutionImpactSpeed,
		s.restitution,
	)

	if s.latched {
		s.DrawScreenTextLinef("coefficient = %.4f at first rebound", s.coefficient)
	}

	s.DrawScreenTextLinef(
		"spin in = %.2f, spin out = %.4f (ideal %.2f at e = 1)",
		s.spin,
		w,
		-s.spin,
	)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = percentScale / s.startEnergy
	}

	s.DrawScreenTextLinef("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLinef("potential = %.3f J", potential)
	s.DrawScreenTextLinef(
		"total     = %.3f J (%.2f%% of start, peak %.2f%%)",
		total,
		scale*total,
		scale*s.peakEnergy,
	)
}

func CreateRotatedBoxRestitution(context *SampleContext) Sampler {
	const (
		initialRestitution  = 0.9
		initialSafetyFactor = 0.5
		cameraCenterY       = 4.0
	)

	s := &RotatedBoxRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		restitution:     initialRestitution,
		friction:        0.0,
		safetyFactor:    initialSafetyFactor,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 14.0
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	h := float32(2.0*rotatedBoxRestitutionCount + 1.0)
	segment := box2d.Segment{Point1: box2d.Vec2{X: -h, Y: 0.0}, Point2: box2d.Vec2{X: h, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	s.CreateScene()

	return s
}

func (s *RotatedBoxRestitution) CreateScene() {
	const (
		boxHalfExtent = 0.5
	)

	for i := range rotatedBoxRestitutionCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}
		}
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.SafetyFactor = s.safetyFactor

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = s.friction
	shapeDef.Material.Restitution = s.restitution

	// Boxes in a negative group never collide with each other.
	shapeDef.Filter.GroupIndex = -1

	box := box2d.MakeBox(boxHalfExtent, boxHalfExtent)

	// A square repeats every quarter turn
	const quarterTurnPiFactor = 0.5

	dq := quarterTurnPiFactor * box2d.Pi / rotatedBoxRestitutionCount
	x := float32(-1.0 * (rotatedBoxRestitutionCount - 1))

	for i := range rotatedBoxRestitutionCount {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: rotatedBoxRestitutionDropHeight}
		bodyDef.Rotation = box2d.MakeRot(float32(i) * dq)
		s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[i].CreatePolygonShape(&shapeDef, &box)

		linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[i:i+1])
		s.startEnergy[i] = linear + angular + potential

		x += 2.0
	}
}

func (s *RotatedBoxRestitution) DrawControls() bool {
	const (
		controlWidthEm  = 6.0
		minSafetyFactor = 0.01
		maxSafetyFactor = 0.5
	)

	rebuild := false

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.SliderFloatV("Restitution", &s.restitution, 0.0, 1.0, "%.1f", 0) {
		rebuild = true
	}

	if imgui.SliderFloatV("Friction", &s.friction, 0.0, 1.0, "%.1f", 0) {
		rebuild = true
	}

	if imgui.SliderFloatV(
		"Safety Factor",
		&s.safetyFactor,
		minSafetyFactor,
		maxSafetyFactor,
		"%.2f",
		0,
	) {
		rebuild = true
	}

	imgui.PopItemWidth()

	if imgui.Button("Reset") {
		rebuild = true
	}

	if rebuild {
		s.CreateScene()
	}

	return true
}

func (s *RotatedBoxRestitution) Step() {
	const (
		boxHalfExtent           = 0.5
		percentScale            = 100.0
		energyLabelHeightOffset = 1.5
	)

	s.Sample.Step()

	h := posScalar(1.0 * rotatedBoxRestitutionCount)
	s.draw.DrawLine(
		box2d.Pos{X: -h, Y: rotatedBoxRestitutionDropHeight + boxHalfExtent},
		box2d.Pos{X: h, Y: rotatedBoxRestitutionDropHeight + boxHalfExtent},
		box2d.ColorRed,
	)

	minPercent := float32(0.0)
	maxPercent := float32(0.0)
	x := float32(-1.0 * (rotatedBoxRestitutionCount - 1))

	for i := range rotatedBoxRestitutionCount {
		linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[i:i+1])
		total := linear + angular + potential

		percent := float32(0.0)
		if s.startEnergy[i] != 0.0 {
			percent = percentScale * total / s.startEnergy[i]
		}

		if i == 0 {
			minPercent = percent
			maxPercent = percent
		} else {
			minPercent = min(minPercent, percent)
			maxPercent = max(maxPercent, percent)
		}

		s.draw.DrawStringf(
			s.camera,
			box2d.Pos{
				X: posScalar(x - boxHalfExtent),
				Y: rotatedBoxRestitutionDropHeight + energyLabelHeightOffset,
			},
			box2d.ColorWhite,
			"%.1f%%",
			percent,
		)
		x += 2.0
	}

	s.DrawScreenTextLinef("energy: min %.2f%%, max %.2f%% of start", minPercent, maxPercent)
}

func CreateTwoBoxRestitution(context *SampleContext) Sampler {
	const (
		cameraCenterY   = 3.0
		groundHalfWidth = 10.0
		boxHalfExtent   = 0.5
	)

	s := &TwoBoxRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 5.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	box := box2d.MakeSquare(boxHalfExtent)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 0.9

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	bodyDef.Position.Y = boxHalfExtent
	s.bodyIDs[0] = s.worldID.CreateBody(&bodyDef)
	s.bodyIDs[0].CreatePolygonShape(&shapeDef, &box)

	bodyDef.Position.X += 0.1
	bodyDef.Position.Y = 6.0
	s.bodyIDs[1] = s.worldID.CreateBody(&bodyDef)
	s.bodyIDs[1].CreatePolygonShape(&shapeDef, &box)

	linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[:])
	s.startEnergy = linear + angular + potential
	s.peakEnergy = s.startEnergy

	return s
}

func (s *TwoBoxRestitution) Step() {
	const (
		percentScale = 100.0
	)

	s.Sample.Step()

	linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[:])
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = percentScale / s.startEnergy
	}

	s.DrawScreenTextLinef("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLinef("potential = %.3f J", potential)
	s.DrawScreenTextLinef(
		"total     = %.3f J (%.2f%% of start, peak %.2f%%)",
		total,
		scale*total,
		scale*s.peakEnergy,
	)
}

func CreateCircleRestitution(context *SampleContext) Sampler {
	const (
		cameraCenterY  = 5.0
		bodySpacing    = 2.0
		circleRadius   = 0.5
		upperRowHeight = 4.0
	)

	s := &CircleRestitution{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 10.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		h := float32(bodySpacing * circleRestitutionCount)
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -h, Y: 0.0},
			Point2: box2d.Vec2{X: h, Y: 0.0},
		}
		shapeDef := box2d.DefaultShapeDef()

		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: circleRadius}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 0.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	dr := 1.0 / float32(max(circleRestitutionCount-1, 1))
	x := float32(-1.0 * (circleRestitutionCount - 1))
	dx := float32(bodySpacing)

	for range circleRestitutionCount {
		name := fmt.Sprintf("%.2f", shapeDef.Material.Restitution)

		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: 1.0}
		bodyDef.Name = name
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: upperRowHeight}
		bodyDef.Name = name
		bodyID = s.worldID.CreateBody(&bodyDef)

		bodyID.CreateCircleShape(&shapeDef, &circle)

		shapeDef.Material.Restitution += dr
		x += dx
	}

	return s
}

func computeEnergy(
	worldID *box2d.World,
	bodyIDs []box2d.Body,
) (float32, float32, float32) {
	const (
		kineticEnergyFactor = 0.5
	)

	gravity := worldID.Gravity()

	linearSum := float32(0.0)
	angularSum := float32(0.0)
	potentialSum := float32(0.0)

	for _, bodyID := range bodyIDs {
		massData := bodyID.MassData()
		v := bodyID.LinearVelocity()
		w := bodyID.AngularVelocity()

		linearSum += kineticEnergyFactor * massData.Mass * v.Dot(v)
		angularSum += kineticEnergyFactor * massData.RotationalInertia * w * w
		potentialSum -= massData.Mass * gravity.Dot(bodyID.WorldCenter().Vec2())
	}

	return linearSum, angularSum, potentialSum
}

func hitEventCount(worldID *box2d.World) int {
	count := 0
	for range worldID.ContactHitEvents() {
		count++
	}

	return count
}

func registerRestitutionSamples() {
	RegisterSample("Restitution", "Varying", CreateVaryingRestitution)
	RegisterSample("Restitution", "Single Box Restitution", CreateSingleBoxRestitution)
	RegisterSample("Restitution", "Single Circle Restitution", CreateSingleCircleRestitution)
	RegisterSample("Restitution", "Circle Stack Restitution", CreateCircleStackRestitution)
	RegisterSample("Restitution", "Restitution Propagation", CreateRestitutionPropagation)
	RegisterSample("Restitution", "Box Restitution", CreateBoxRestitution)
	RegisterSample("Restitution", "Rotated Box Restitution", CreateRotatedBoxRestitution)
	RegisterSample("Restitution", "Two Box Restitution", CreateTwoBoxRestitution)
	RegisterSample("Restitution", "Circle Restitution", CreateCircleRestitution)
}
