// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d"
)

func computeEnergy(worldID *box2d.World, bodyIDs []box2d.Body) (linear, angular, potential float32) {
	gravity := worldID.Gravity()

	linearSum := float32(0.0)
	angularSum := float32(0.0)
	potentialSum := float32(0.0)

	for _, bodyID := range bodyIDs {
		massData := bodyID.MassData()
		v := bodyID.LinearVelocity()
		w := bodyID.AngularVelocity()

		linearSum += 0.5 * massData.Mass * v.Dot(v)
		angularSum += 0.5 * massData.RotationalInertia * w * w
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

const (
	varyingRestitutionCircleShape = 0
	varyingRestitutionBoxShape    = 1
	varyingRestitutionCount       = 40
)

// Restitution is approximate since Box2D uses speculative collision
type VaryingRestitution struct {
	Sample

	bodyIDs   [varyingRestitutionCount]box2d.Body
	shapeType int32
}

func CreateVaryingRestitution(context *SampleContext) Sampler {
	s := &VaryingRestitution{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 4.0, Y: 17.0}
		s.context.camera.zoom = 27.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		h := float32(1.0 * varyingRestitutionCount)
		segment := box2d.Segment{Point1: box2d.Vec2{X: -h, Y: 0.0}, Point2: box2d.Vec2{X: h, Y: 0.0}}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.shapeType = varyingRestitutionCircleShape

	s.CreateBodies()
	return s
}

func (s *VaryingRestitution) CreateBodies() {
	for i := range varyingRestitutionCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}
		}
	}

	circle := box2d.Circle{}
	circle.Radius = 0.5

	box := box2d.MakeBox(0.5, 0.5)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 0.0
	shapeDef.Material.Friction = 0.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	dr := 1.0 / float32(max(varyingRestitutionCount-1, 1))
	x := float32(-1.0 * (varyingRestitutionCount - 1))
	dx := float32(2.0)

	for i := range varyingRestitutionCount {
		bodyDef.Name = fmt.Sprintf("%.2f", shapeDef.Material.Restitution)
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: 40.0}
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
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false
	shapeTypes := []string{"Circle", "Box"}

	changed = changed || imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))

	imgui.PopItemWidth()

	changed = changed || imgui.Button("Reset")

	if changed {
		s.CreateBodies()
	}

	return true
}

func (s *VaryingRestitution) Step() {
	s.Sample.Step()

	h := posScalar(1.0 * varyingRestitutionCount)
	s.draw.DrawLine(box2d.Pos{X: -h, Y: 40.5}, box2d.Pos{X: h, Y: 40.5}, box2d.ColorRed)
}

// Tests how a single bouncing box behaves.
type SingleBoxRestitution struct {
	Sample

	bodyID      box2d.Body
	maxY        float32
	startEnergy float32
	peakEnergy  float32
	height      float32
}

func CreateSingleBoxRestitution(context *SampleContext) Sampler {
	s := &SingleBoxRestitution{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 8.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{Point1: box2d.Vec2{X: -4.0, Y: 0.0}, Point2: box2d.Vec2{X: 4.0, Y: 0.0}}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.height = 5.0
	s.maxY = 5.0

	box := box2d.MakeBox(0.5, 0.5)

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
	s.Sample.Step()

	s.draw.DrawLine(box2d.Pos{X: -4.0, Y: posScalar(s.height + 0.5)}, box2d.Pos{X: 4.0, Y: posScalar(s.height + 0.5)}, box2d.ColorRed)

	p := s.bodyID.Position()
	if hitEventCount(s.worldID) == 1 {
		s.maxY = float32(p.Y)
	} else {
		s.maxY = max(s.maxY, float32(p.Y))
	}

	s.DrawScreenTextLine("maxY = %.2f", s.maxY)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.bodyID})
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = 100.0 / s.startEnergy
	}
	s.DrawScreenTextLine("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLine("potential = %.3f J", potential)
	s.DrawScreenTextLine("total     = %.3f J (%.2f%% of start, peak %.2f%%)", total, scale*total, scale*s.peakEnergy)
}

type SingleCircleRestitution struct {
	Sample

	bodyID      box2d.Body
	maxY        float32
	startEnergy float32
	peakEnergy  float32
}

func CreateSingleCircleRestitution(context *SampleContext) Sampler {
	s := &SingleCircleRestitution{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 8.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{Point1: box2d.Vec2{X: -4.0, Y: 0.0}, Point2: box2d.Vec2{X: 4.0, Y: 0.0}}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: 0.5}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.EnableHitEvents = true

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 10.0}
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
	s.Sample.Step()

	p := s.bodyID.Position()
	if hitEventCount(s.worldID) == 1 {
		s.maxY = float32(p.Y)
	} else {
		s.maxY = max(s.maxY, float32(p.Y))
	}

	s.draw.DrawLine(box2d.Pos{X: -4.0, Y: 10.5}, box2d.Pos{X: 4.0, Y: 10.5}, box2d.ColorRed)

	s.DrawScreenTextLine("maxY = %.2f", s.maxY)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.bodyID})
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = 100.0 / s.startEnergy
	}
	s.DrawScreenTextLine("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLine("potential = %.3f J", potential)
	s.DrawScreenTextLine("total     = %.3f J (%.2f%% of start, peak %.2f%%)", total, scale*total, scale*s.peakEnergy)
}

const (
	circleStackImpactSpeed   = 5.0
	circleStackMaxStackCount = 3
)

// Similar to MeasureSupportedBounce unit test.
type CircleStackRestitution struct {
	Sample

	bodyIDs     [circleStackMaxStackCount + 1]box2d.Body
	impactorID  box2d.Body
	bodyCount   int
	stackCount  int32
	restitution float32
	startHeight float32
	coefficient float32
	startEnergy float32
	peakEnergy  float32
	gravity     bool
	latched     bool
	hit         bool
}

func CreateCircleStackRestitution(context *SampleContext) Sampler {
	s := &CircleStackRestitution{stackCount: 2, restitution: 1.0}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.5}
		s.context.camera.zoom = 6.0
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = 0.0
	box := box2d.MakeBox(40.0, 1.0)
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.CreateScene()
	return s
}

func (s *CircleStackRestitution) CreateBall(y, velocityY, restitution float32, hitEvents bool) box2d.Body {
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
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	bodyID.CreateCircleShape(&shapeDef, &circle)

	return bodyID
}

func (s *CircleStackRestitution) CreateScene() {
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
		s.bodyIDs[s.bodyCount] = s.CreateBall(0.5+1.0*float32(i), 0.0, 0.0, false)
		s.bodyCount++
	}

	// Start half a step of travel above contact so the impact lands mid step, matching the test
	s.startHeight = 0.5 + 1.0*float32(s.stackCount) + 0.5*circleStackImpactSpeed*(1.0/60.0)

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
	rebuild := false

	imgui.PushItemWidth(6.0 * imgui.FontSize())

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

	s.draw.DrawLine(box2d.Pos{X: -2.0, Y: posScalar(s.startHeight)}, box2d.Pos{X: 2.0, Y: posScalar(s.startHeight)}, box2d.ColorRed)

	linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[:s.bodyCount])
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	s.DrawScreenTextLine("impactor vy = %.3f m/s", vy)

	s.DrawScreenTextLine("coefficient = %.4f live (target %.2f)", vy/circleStackImpactSpeed, s.restitution)

	if s.latched {
		s.DrawScreenTextLine("coefficient = %.4f at first rebound", s.coefficient)
	}

	s.DrawScreenTextLine("support speed = %.4f m/s", supportSpeed)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = 100.0 / s.startEnergy
	}
	s.DrawScreenTextLine("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLine("potential = %.3f J", potential)
	s.DrawScreenTextLine("total     = %.3f J (%.2f%% of start, peak %.2f%%)", total, scale*total, scale*s.peakEnergy)
}

const (
	restitutionPropagationColumnCount = 3
	restitutionPropagationDropHeight  = 5.0
)

var restitutionPropagationLabels = [restitutionPropagationColumnCount]string{"ground", "one crate", "two crates"}

// A bouncy ball landing on a resting body. Three balls fall together from the same height onto bare
// ground, one crate and two crates. Without restitution propagation the restitution stage only
// visits the ball's own contact, so the rebound is shared with the crate, which is driven into the
// ground and pushed back the next step. With propagation the ground contact under the crate can react
// in the restitution sover.
type RestitutionPropagation struct {
	Sample

	bodyIDs     [2 * restitutionPropagationColumnCount]box2d.Body
	ballIDs     [restitutionPropagationColumnCount]box2d.Body
	crateIDs    [restitutionPropagationColumnCount]box2d.Body
	restHeight  [restitutionPropagationColumnCount]float32
	apex        [restitutionPropagationColumnCount]float32
	rebounding  [restitutionPropagationColumnCount]bool
	done        [restitutionPropagationColumnCount]bool
	bodyCount   int
	crateCount  int
	restitution float32
	kick        float32
}

func CreateRestitutionPropagation(context *SampleContext) Sampler {
	s := &RestitutionPropagation{restitution: 0.9}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 4.0}
		s.context.camera.zoom = 10.0
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.6
	box := box2d.MakeBox(40.0, 1.0)
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.CreateScene()
	return s
}

func (s *RestitutionPropagation) CreateScene() {
	for i := range s.bodyCount {
		s.bodyIDs[i].Destroy()
	}
	s.bodyCount = 0
	s.crateCount = 0
	s.kick = 0.0

	box := box2d.MakeBox(0.5, 0.5)
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

	for column := range restitutionPropagationColumnCount {
		x := 4.0 * float32(column-1)

		for k := range column {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(0.5 + 1.0*float32(k))}
			bodyDef.EnableSleep = false
			crateID := s.worldID.CreateBody(&bodyDef)

			shapeDef := box2d.DefaultShapeDef()
			crateID.CreatePolygonShape(&shapeDef, &box)

			s.bodyIDs[s.bodyCount] = crateID
			s.bodyCount++
			s.crateIDs[s.crateCount] = crateID
			s.crateCount++
		}

		s.restHeight[column] = 0.5 + 1.0*float32(column)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(s.restHeight[column] + restitutionPropagationDropHeight)}
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
	imgui.PushItemWidth(6.0 * imgui.FontSize())
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

		x := posScalar(4.0 * float32(column-1))
		dropY := posScalar(s.restHeight[column] + restitutionPropagationDropHeight)
		s.draw.DrawLine(box2d.Pos{X: x - 1.0, Y: dropY}, box2d.Pos{X: x + 1.0, Y: dropY}, box2d.ColorRed)
		if s.apex[column] > 0.0 {
			apexY := posScalar(s.restHeight[column] + s.apex[column])
			s.draw.DrawLine(box2d.Pos{X: x - 1.0, Y: apexY}, box2d.Pos{X: x + 1.0, Y: apexY}, box2d.ColorGreen)
		}

		s.DrawScreenTextLine("%s: apex %.2f m (%.0f%% of drop)", restitutionPropagationLabels[column], s.apex[column],
			100.0*s.apex[column]/restitutionPropagationDropHeight)
	}

	for i := range s.crateCount {
		s.kick = max(s.kick, s.crateIDs[i].LinearVelocity().Length())
	}

	s.DrawScreenTextLine("peak crate speed %.2f m/s", s.kick)
}

const boxRestitutionImpactSpeed = 5.0

// Similar to MeasureFlatBounce and SpinTest unit tests.
type BoxRestitution struct {
	Sample

	boxID       box2d.Body
	restitution float32
	spin        float32
	coefficient float32
	startEnergy float32
	peakEnergy  float32
	latched     bool
	hit         bool
}

func CreateBoxRestitution(context *SampleContext) Sampler {
	s := &BoxRestitution{restitution: 0.9}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 8.0}
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
	box := box2d.MakeBox(40.0, 1.0)
	groundID.CreatePolygonShape(&shapeDef, &box)

	s.CreateScene()
	return s
}

func (s *BoxRestitution) CreateScene() {
	if s.boxID != (box2d.Body{}) {
		s.boxID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(float32(0.25 + 0.5*boxRestitutionImpactSpeed*(1.0/60.0)))}
	bodyDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -boxRestitutionImpactSpeed}
	bodyDef.AngularVelocity = s.spin
	bodyDef.EnableSleep = false
	s.boxID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = s.restitution
	shapeDef.EnableHitEvents = true
	box := box2d.MakeBox(1.0, 0.25)
	s.boxID.CreatePolygonShape(&shapeDef, &box)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.boxID})
	s.startEnergy = linear + angular + potential
	s.peakEnergy = s.startEnergy

	s.coefficient = 0.0
	s.latched = false
	s.hit = false
}

func (s *BoxRestitution) DrawControls() bool {
	rebuild := false

	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Restitution", &s.restitution, 0.0, 1.0, "%.2f", 0) {
		rebuild = true
	}

	if imgui.SliderFloatV("Spin", &s.spin, 0.0, 2.0, "%.2f", 0) {
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

	s.draw.DrawPoint(s.boxID.WorldPoint(box2d.Vec2{X: -1.0, Y: -0.25}), 8.0, box2d.ColorYellow)
	s.draw.DrawPoint(s.boxID.WorldPoint(box2d.Vec2{X: 1.0, Y: -0.25}), 8.0, box2d.ColorYellow)

	linear, angular, potential := computeEnergy(s.worldID, []box2d.Body{s.boxID})
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	s.DrawScreenTextLine("vy = %.3f m/s", vy)

	s.DrawScreenTextLine("coefficient = %.4f live (target %.2f)", vy/boxRestitutionImpactSpeed, s.restitution)

	if s.latched {
		s.DrawScreenTextLine("coefficient = %.4f at first rebound", s.coefficient)
	}

	s.DrawScreenTextLine("spin in = %.2f, spin out = %.4f (ideal %.2f at e = 1)", s.spin, w, -s.spin)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = 100.0 / s.startEnergy
	}
	s.DrawScreenTextLine("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLine("potential = %.3f J", potential)
	s.DrawScreenTextLine("total     = %.3f J (%.2f%% of start, peak %.2f%%)", total, scale*total, scale*s.peakEnergy)
}

const (
	rotatedBoxRestitutionCount      = 12
	rotatedBoxRestitutionDropHeight = 10.0
)

type RotatedBoxRestitution struct {
	Sample

	bodyIDs      [rotatedBoxRestitutionCount]box2d.Body
	startEnergy  [rotatedBoxRestitutionCount]float32
	restitution  float32
	friction     float32
	safetyFactor float32
}

func CreateRotatedBoxRestitution(context *SampleContext) Sampler {
	s := &RotatedBoxRestitution{restitution: 0.9, friction: 0.0, safetyFactor: 0.5}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 4.0}
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

	box := box2d.MakeBox(0.5, 0.5)

	// A square repeats every quarter turn
	dq := 0.5 * box2d.Pi / rotatedBoxRestitutionCount
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
	rebuild := false

	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Restitution", &s.restitution, 0.0, 1.0, "%.1f", 0) {
		rebuild = true
	}

	if imgui.SliderFloatV("Friction", &s.friction, 0.0, 1.0, "%.1f", 0) {
		rebuild = true
	}

	if imgui.SliderFloatV("Safety Factor", &s.safetyFactor, 0.01, 0.5, "%.2f", 0) {
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
	s.Sample.Step()

	h := posScalar(1.0 * rotatedBoxRestitutionCount)
	s.draw.DrawLine(box2d.Pos{X: -h, Y: rotatedBoxRestitutionDropHeight + 0.5}, box2d.Pos{X: h, Y: rotatedBoxRestitutionDropHeight + 0.5}, box2d.ColorRed)

	minPercent := float32(0.0)
	maxPercent := float32(0.0)
	x := float32(-1.0 * (rotatedBoxRestitutionCount - 1))

	for i := range rotatedBoxRestitutionCount {
		linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[i:i+1])
		total := linear + angular + potential

		percent := float32(0.0)
		if s.startEnergy[i] != 0.0 {
			percent = 100.0 * total / s.startEnergy[i]
		}
		if i == 0 {
			minPercent = percent
			maxPercent = percent
		} else {
			minPercent = min(minPercent, percent)
			maxPercent = max(maxPercent, percent)
		}

		s.draw.DrawString(s.camera, box2d.Pos{X: posScalar(x - 0.5), Y: rotatedBoxRestitutionDropHeight + 1.5}, box2d.ColorWhite, "%.1f%%", percent)
		x += 2.0
	}

	s.DrawScreenTextLine("energy: min %.2f%%, max %.2f%% of start", minPercent, maxPercent)
}

// Based on https://github.com/erincatto/box2d/discussions/957
type TwoBoxRestitution struct {
	Sample

	bodyIDs     [2]box2d.Body
	startEnergy float32
	peakEnergy  float32
}

func CreateTwoBoxRestitution(context *SampleContext) Sampler {
	s := &TwoBoxRestitution{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 3.0}
		s.context.camera.zoom = 5.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	box := box2d.MakeSquare(0.5)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Restitution = 0.9

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	bodyDef.Position.Y = 0.5
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
	s.Sample.Step()

	linear, angular, potential := computeEnergy(s.worldID, s.bodyIDs[:])
	total := linear + angular + potential
	s.peakEnergy = max(s.peakEnergy, total)

	scale := float32(0.0)
	if s.startEnergy != 0.0 {
		scale = 100.0 / s.startEnergy
	}
	s.DrawScreenTextLine("kinetic   = %.3f J linear + %.3f J angular", linear, angular)
	s.DrawScreenTextLine("potential = %.3f J", potential)
	s.DrawScreenTextLine("total     = %.3f J (%.2f%% of start, peak %.2f%%)", total, scale*total, scale*s.peakEnergy)
}

const circleRestitutionCount = 10

type CircleRestitution struct {
	Sample
}

func CreateCircleRestitution(context *SampleContext) Sampler {
	s := &CircleRestitution{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 10.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		h := float32(2.0 * circleRestitutionCount)
		segment := box2d.Segment{Point1: box2d.Vec2{X: -h, Y: 0.0}, Point2: box2d.Vec2{X: h, Y: 0.0}}
		shapeDef := box2d.DefaultShapeDef()

		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: 0.5}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Restitution = 0.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	dr := 1.0 / float32(max(circleRestitutionCount-1, 1))
	x := float32(-1.0 * (circleRestitutionCount - 1))
	dx := float32(2.0)

	for range circleRestitutionCount {
		name := fmt.Sprintf("%.2f", shapeDef.Material.Restitution)

		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: 1.0}
		bodyDef.Name = name
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: 4.0}
		bodyDef.Name = name
		bodyID = s.worldID.CreateBody(&bodyDef)

		bodyID.CreateCircleShape(&shapeDef, &circle)

		shapeDef.Material.Restitution += dr
		x += dx
	}
	return s
}

func init() {
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
