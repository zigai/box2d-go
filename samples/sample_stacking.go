// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/scenes"
)

type SingleBox struct {
	Sample

	bodyID box2d.Body
}

func CreateSingleBox(context *SampleContext) Sampler {
	s := &SingleBox{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.5}
		s.context.camera.zoom = 3.5
	}

	extent := float32(1.0)

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	groundWidth := 66.0 * extent
	shapeDef := box2d.DefaultShapeDef()
	// shapeDef.friction = 0.5f;

	segment := box2d.Segment{Point1: box2d.Vec2{X: -0.5 * 2.0 * groundWidth, Y: 0.0}, Point2: box2d.Vec2{X: 0.5 * 2.0 * groundWidth, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &segment)
	bodyDef.Type = box2d.DynamicBody

	box := box2d.MakeBox(extent, extent)
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
	bodyDef.LinearVelocity = box2d.Vec2{X: 5.0, Y: 0.0}
	s.bodyID = s.worldID.CreateBody(&bodyDef)
	s.bodyID.CreatePolygonShape(&shapeDef, &box)
	return s
}

func (s *SingleBox) Step() {
	s.Sample.Step()

	position := s.bodyID.Position()
	s.DrawScreenTextLine("(x, y) = (%.2g, %.2g)", position.X, position.Y)
}

const (
	tiltedStackColumns = 10
	tiltedStackRows    = 10
)

type TiltedStack struct {
	Sample

	bodies [tiltedStackRows * tiltedStackColumns]box2d.Body
}

func CreateTiltedStack(context *SampleContext) Sampler {
	s := &TiltedStack{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 7.5, Y: 7.5}
		s.context.camera.zoom = 20.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(1000.0, 1.0)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	box := box2d.MakeRoundedBox(0.45, 0.45, 0.05)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.3

	offset := float32(0.2)
	dx := float32(5.0)
	xroot := -0.5 * dx * (tiltedStackColumns - 1.0)

	for j := range tiltedStackColumns {
		x := xroot + float32(j)*dx

		for i := range tiltedStackRows {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody

			n := j*tiltedStackRows + i

			bodyDef.Position = box2d.Pos{X: posScalar(x + offset*float32(i)), Y: posScalar(0.5 + 1.0*float32(i))}
			bodyID := s.worldID.CreateBody(&bodyDef)

			s.bodies[n] = bodyID

			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
	return s
}

const (
	verticalStackMaxColumns = 10
	verticalStackMaxRows    = 80
	verticalStackMaxBullets = 8

	verticalStackCircleShape = 0
	verticalStackBoxShape    = 1
)

// This sample shows some aspects of Box2D continuous collision:
// - bullet dynamic bodies which support continuous collision with non-bullet dynamic bodies
// - prevention of chain reaction tunneling
// Try disabling continuous collision and firing a bullet. You might see a bullet push a
// a through the static wall.
type VerticalStack struct {
	Sample

	bullets     [verticalStackMaxBullets]box2d.Body
	bodies      [verticalStackMaxRows * verticalStackMaxColumns]box2d.Body
	columnCount int32
	rowCount    int32
	bulletCount int32
	shapeType   int32
	bulletType  int32
}

func CreateVerticalStack(context *SampleContext) Sampler {
	s := &VerticalStack{}
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

		segment := box2d.Segment{Point1: box2d.Vec2{X: 10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 20.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{Point1: box2d.Vec2{X: -30.0, Y: 0.0}, Point2: box2d.Vec2{X: 30.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.shapeType = verticalStackBoxShape
	s.rowCount = 12
	s.columnCount = 1
	s.bulletCount = 1
	s.bulletType = verticalStackCircleShape

	s.CreateStacks()
	return s
}

func (s *VerticalStack) CreateStacks() {
	for i := range s.bodies {
		if s.bodies[i] != (box2d.Body{}) {
			s.bodies[i].Destroy()
			s.bodies[i] = box2d.Body{}
		}
	}

	circle := box2d.Circle{}
	circle.Radius = 0.5

	box := box2d.MakeRoundedBox(0.45, 0.45, 0.05)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.3

	var offset float32

	if s.shapeType == verticalStackCircleShape {
		offset = 0.0
	} else {
		offset = 0.01
	}

	dx := float32(-3.0)
	xroot := float32(8.0)

	for j := range int(s.columnCount) {
		x := xroot + float32(j)*dx

		for i := range int(s.rowCount) {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody

			n := j*int(s.rowCount) + i

			shift := offset
			if i%2 == 0 {
				shift = -offset
			}
			bodyDef.Position = box2d.Pos{X: posScalar(x + shift), Y: posScalar(0.5 + 1.0*float32(i))}
			// bodyDef.position = {x + shift, 1.0f + 1.51f * i};
			bodyID := s.worldID.CreateBody(&bodyDef)

			s.bodies[n] = bodyID

			if s.shapeType == verticalStackCircleShape {
				bodyID.CreateCircleShape(&shapeDef, &circle)
			} else {
				bodyID.CreatePolygonShape(&shapeDef, &box)
			}
		}
	}
}

func (s *VerticalStack) DestroyBody() {
	for j := range int(s.columnCount) {
		for i := range int(s.rowCount) {
			n := j*int(s.rowCount) + i

			if s.bodies[n] != (box2d.Body{}) {
				s.bodies[n].Destroy()
				s.bodies[n] = box2d.Body{}
				break
			}
		}
	}
}

func (s *VerticalStack) DestroyBullets() {
	for i := range verticalStackMaxBullets {
		bullet := s.bullets[i]

		if bullet != (box2d.Body{}) {
			bullet.Destroy()
			s.bullets[i] = box2d.Body{}
		}
	}
}

func (s *VerticalStack) FireBullets() {
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.25}
	box := box2d.MakeBox(0.25, 0.25)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 4.0

	for i := range int(s.bulletCount) {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-26.7 - float32(i)), Y: 6.0}
		speed := scenes.RandomFloatRange(200.0, 300.0)
		bodyDef.LinearVelocity = box2d.Vec2{X: speed, Y: 0.0}
		bodyDef.IsBullet = true

		bullet := s.worldID.CreateBody(&bodyDef)

		if s.bulletType == verticalStackBoxShape {
			bullet.CreatePolygonShape(&shapeDef, &box)
		} else {
			bullet.CreateCircleShape(&shapeDef, &circle)
		}
		s.bullets[i] = bullet
	}
}

func (s *VerticalStack) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	consumed := false

	switch key {
	case glfw.KeyB:
		s.FireBullets()
		consumed = true
	}

	if !consumed {
		s.Sample.Keyboard(key, action, mods)
	}
}

func (s *VerticalStack) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false
	shapeTypes := []string{"Circle", "Box"}

	changed = changed || imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))

	changed = changed || imgui.SliderInt("Rows", &s.rowCount, 1, verticalStackMaxRows)
	changed = changed || imgui.SliderInt("Columns", &s.columnCount, 1, verticalStackMaxColumns)

	imgui.SliderInt("Bullets", &s.bulletCount, 1, verticalStackMaxBullets)

	imgui.ComboStrarr("Bullet Shape", &s.bulletType, shapeTypes, int32(len(shapeTypes)))

	imgui.PopItemWidth()

	if imgui.Button("Fire Bullets") {
		s.DestroyBullets()
		s.FireBullets()
	}

	if imgui.Button("Destroy Body") {
		s.DestroyBody()
	}

	changed = changed || imgui.Button("Reset Stack")

	if changed {
		s.DestroyBullets()
		s.CreateStacks()
	}

	return true
}

type circleStackEvent struct {
	indexA, indexB int
}

// A simple circle stack that also shows how to collect hit events
type CircleStack struct {
	Sample

	events []circleStackEvent
}

func CreateCircleStack(context *SampleContext) Sampler {
	s := &CircleStack{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 6.0
	}

	shapeIndex := 0

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.UserData = shapeIndex
		shapeIndex++

		segment := box2d.Segment{Point1: box2d.Vec2{X: -10.0, Y: 0.0}, Point2: box2d.Vec2{X: 10.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.worldID.SetGravity(box2d.Vec2{X: 0.0, Y: -20.0})
	s.worldID.SetContactTuning(0.25*360.0, 10.0, 3.0)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	circle := box2d.Circle{}
	circle.Radius = 0.5

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableHitEvents = true
	// shapeDef.rollingResistance = 0.2f;
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = 0.8

	y := float32(0.75)

	for i := range 4 {
		bodyDef.Position.Y = posScalar(y)

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef.UserData = shapeIndex
		shapeDef.Density = 1.0 + 4.0*float32(i)
		shapeIndex++
		bodyID.CreateCircleShape(&shapeDef, &circle)

		y += 1.25
	}
	return s
}

func (s *CircleStack) Step() {
	s.Sample.Step()

	for event := range s.worldID.ContactHitEvents() {
		indexA := event.ShapeA.UserData().(int)
		indexB := event.ShapeB.UserData().(int)

		s.draw.DrawPoint(event.Point, 10.0, box2d.ColorWhite)

		s.events = append(s.events, circleStackEvent{indexA, indexB})
	}

	for _, event := range s.events {
		s.DrawScreenTextLine("%d, %d", event.indexA, event.indexB)
	}
}

type CapsuleStack struct {
	Sample
}

func CreateCapsuleStack(context *SampleContext) Sampler {
	s := &CapsuleStack{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		polygon := box2d.MakeBox(10.0, 1.0)
		groundID.CreatePolygonShape(&shapeDef, &polygon)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	a := float32(0.25)
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -4.0 * a, Y: 0.0}, Center2: box2d.Vec2{X: 4.0 * a, Y: 0.0}, Radius: a}

	shapeDef := box2d.DefaultShapeDef()

	// rolling resistance increases stacking stability
	// shapeDef.rollingResistance = 0.2f;

	y := 2.0 * a

	for range 20 {
		bodyDef.Position.Y = posScalar(y)
		// bodyDef.position.x += ( i & 1 ) == 1 ? -0.5f * a : 0.5f * a;
		// bodyDef.linearVelocity = { 0.0f, -10.0f };
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreateCapsuleShape(&shapeDef, &capsule)

		y += 3.0 * a
	}
	return s
}

type Cliff struct {
	Sample

	bodyIDs [9]box2d.Body
	flip    bool
}

func CreateCliff(context *SampleContext) Sampler {
	s := &Cliff{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 0.5
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(100.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		segment := box2d.Segment{Point1: box2d.Vec2{X: -14.0, Y: 4.0}, Point2: box2d.Vec2{X: -8.0, Y: 4.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box = box2d.MakeOffsetBox(3.0, 0.5, box2d.Vec2{X: 0.0, Y: 4.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		capsule := box2d.Capsule{Center1: box2d.Vec2{X: 8.5, Y: 4.0}, Center2: box2d.Vec2{X: 13.5, Y: 4.0}, Radius: 0.5}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	s.flip = false

	for i := range 9 {
		s.bodyIDs[i] = box2d.Body{}
	}

	s.CreateBodies()
	return s
}

func (s *Cliff) CreateBodies() {
	for i := range 9 {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}
		}
	}

	sign := float32(1.0)
	if s.flip {
		sign = -1.0
	}

	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.25, Y: 0.0}, Center2: box2d.Vec2{X: 0.25, Y: 0.0}, Radius: 0.25}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	square := box2d.MakeSquare(0.5)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	{
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.01
		bodyDef.LinearVelocity = box2d.Vec2{X: 2.0 * sign, Y: 0.0}

		offset := posScalar(0.0)
		if s.flip {
			offset = -4.0
		}

		bodyDef.Position = box2d.Pos{X: -9.0 + offset, Y: 4.25}
		s.bodyIDs[0] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[0].CreateCapsuleShape(&shapeDef, &capsule)

		bodyDef.Position = box2d.Pos{X: 2.0 + offset, Y: 4.75}
		s.bodyIDs[1] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[1].CreateCapsuleShape(&shapeDef, &capsule)

		bodyDef.Position = box2d.Pos{X: 13.0 + offset, Y: 4.75}
		s.bodyIDs[2] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[2].CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.01
		bodyDef.LinearVelocity = box2d.Vec2{X: 2.5 * sign, Y: 0.0}

		bodyDef.Position = box2d.Pos{X: -11.0, Y: 4.5}
		s.bodyIDs[3] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[3].CreatePolygonShape(&shapeDef, &square)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: 5.0}
		s.bodyIDs[4] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[4].CreatePolygonShape(&shapeDef, &square)

		bodyDef.Position = box2d.Pos{X: 11.0, Y: 5.0}
		s.bodyIDs[5] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[5].CreatePolygonShape(&shapeDef, &square)
	}

	{
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.2
		bodyDef.LinearVelocity = box2d.Vec2{X: 1.5 * sign, Y: 0.0}

		offset := posScalar(0.0)
		if s.flip {
			offset = 4.0
		}

		bodyDef.Position = box2d.Pos{X: -13.0 + offset, Y: 4.5}
		s.bodyIDs[6] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[6].CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: -2.0 + offset, Y: 5.0}
		s.bodyIDs[7] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[7].CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: 9.0 + offset, Y: 5.0}
		s.bodyIDs[8] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[8].CreateCircleShape(&shapeDef, &circle)
	}
}

func (s *Cliff) DrawControls() bool {
	if imgui.Button("Flip") {
		s.flip = !s.flip
		s.CreateBodies()
	}

	return true
}

type Arch struct {
	Sample
}

func CreateArch(context *SampleContext) Sampler {
	s := &Arch{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 8.0}
		s.context.camera.zoom = 25.0 * 0.35
	}

	ps1 := [9]box2d.Vec2{
		{X: 16.0, Y: 0.0},
		{X: 14.93803712795643, Y: 5.133601056842984},
		{X: 13.79871746027416, Y: 10.24928069555078},
		{X: 12.56252963284711, Y: 15.34107019122473},
		{X: 11.20040987372525, Y: 20.39856541571217},
		{X: 9.66521217819836, Y: 25.40369899225096},
		{X: 7.87179930638133, Y: 30.3179337000085},
		{X: 5.635199558196225, Y: 35.03820717801641},
		{X: 2.405937953536585, Y: 39.09554102558315},
	}

	ps2 := [9]box2d.Vec2{
		{X: 24.0, Y: 0.0},
		{X: 22.33619528222415, Y: 6.02299846205841},
		{X: 20.54936888969905, Y: 12.00964361211476},
		{X: 18.60854610798073, Y: 17.9470321677465},
		{X: 16.46769273811807, Y: 23.81367936585418},
		{X: 14.05325025774858, Y: 29.57079353071012},
		{X: 11.23551045834022, Y: 35.13775818285372},
		{X: 7.752568160730571, Y: 40.30450679009583},
		{X: 3.016931552701656, Y: 44.28891593799322},
	}

	scale := float32(0.25)
	for i := range 9 {
		ps1[i] = ps1[i].Scale(scale)
		ps2[i] = ps2[i].Scale(scale)
	}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.6

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		segment := box2d.Segment{Point1: box2d.Vec2{X: -100.0, Y: 0.0}, Point2: box2d.Vec2{X: 100.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	for i := range 8 {
		bodyID := s.worldID.CreateBody(&bodyDef)
		ps := []box2d.Vec2{ps1[i], ps2[i], ps2[i+1], ps1[i+1]}
		hull := box2d.ComputeHull(ps)
		polygon := box2d.MakePolygon(&hull, 0.0)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
	}

	for i := range 8 {
		bodyID := s.worldID.CreateBody(&bodyDef)
		ps := []box2d.Vec2{
			{X: -ps2[i].X, Y: ps2[i].Y},
			{X: -ps1[i].X, Y: ps1[i].Y},
			{X: -ps1[i+1].X, Y: ps1[i+1].Y},
			{X: -ps2[i+1].X, Y: ps2[i+1].Y},
		}
		hull := box2d.ComputeHull(ps)
		polygon := box2d.MakePolygon(&hull, 0.0)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
	}

	{
		bodyID := s.worldID.CreateBody(&bodyDef)
		ps := []box2d.Vec2{ps1[8], ps2[8], {X: -ps2[8].X, Y: ps2[8].Y}, {X: -ps1[8].X, Y: ps1[8].Y}}
		hull := box2d.ComputeHull(ps)
		polygon := box2d.MakePolygon(&hull, 0.0)
		bodyID.CreatePolygonShape(&shapeDef, &polygon)
	}

	for i := range 4 {
		box := box2d.MakeBox(2.0, 0.5)
		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(0.5 + ps2[8].Y + 1.0*float32(i))}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

type DoubleDomino struct {
	Sample
}

func CreateDoubleDomino(context *SampleContext) Sampler {
	s := &DoubleDomino{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 4.0}
		s.context.camera.zoom = 25.0 * 0.25
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(100.0, 1.0)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	box := box2d.MakeBox(0.125, 0.5)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.6
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	count := 15
	x := float32(-0.5 * float32(count))
	for i := range count {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: 0.5}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)
		if i == 0 {
			bodyID.ApplyLinearImpulse(box2d.Vec2{X: 0.2, Y: 0.0}, box2d.Pos{X: posScalar(x), Y: 1.0}, true)
		}

		x += 1.0
	}
	return s
}

const (
	confinedGridCount = 25
	confinedMaxCount  = confinedGridCount * confinedGridCount
)

type Confined struct {
	Sample

	row    int
	column int
	count  int
}

func CreateConfined(context *SampleContext) Sampler {
	s := &Confined{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 10.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		var capsule box2d.Capsule
		capsule = box2d.Capsule{Center1: box2d.Vec2{X: -10.5, Y: 0.0}, Center2: box2d.Vec2{X: 10.5, Y: 0.0}, Radius: 0.5}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
		capsule = box2d.Capsule{Center1: box2d.Vec2{X: -10.5, Y: 0.0}, Center2: box2d.Vec2{X: -10.5, Y: 20.5}, Radius: 0.5}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
		capsule = box2d.Capsule{Center1: box2d.Vec2{X: 10.5, Y: 0.0}, Center2: box2d.Vec2{X: 10.5, Y: 20.5}, Radius: 0.5}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
		capsule = box2d.Capsule{Center1: box2d.Vec2{X: -10.5, Y: 20.5}, Center2: box2d.Vec2{X: 10.5, Y: 20.5}, Radius: 0.5}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	s.row = 0
	s.column = 0
	s.count = 0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.0

	shapeDef := box2d.DefaultShapeDef()
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

	for s.count < confinedMaxCount {
		s.row = 0
		for range confinedGridCount {
			x := -8.75 + float32(s.column)*18.0/confinedGridCount
			y := 1.5 + float32(s.row)*18.0/confinedGridCount

			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreateCircleShape(&shapeDef, &circle)

			s.count++
			s.row++
		}
		s.column++
	}
	return s
}

// From PEEL
type CardHouse struct {
	Sample
}

func CreateCardHouse(context *SampleContext) Sampler {
	s := &CardHouse{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.75, Y: 0.9}
		s.context.camera.zoom = 25.0 * 0.05
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -2.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.7

	groundBox := box2d.MakeBox(40.0, 2.0)
	groundID.CreatePolygonShape(&shapeDef, &groundBox)

	cardHeight := float32(0.2)
	cardThickness := float32(0.001)

	angle0 := 25.0 * box2d.Pi / 180.0
	angle1 := -25.0 * box2d.Pi / 180.0
	angle2 := 0.5 * box2d.Pi

	cardBox := box2d.MakeBox(cardThickness, cardHeight)
	bodyDef.Type = box2d.DynamicBody

	Nb := 5
	z0 := float32(0.0)
	y := cardHeight - 0.02
	for Nb != 0 {
		z := z0
		for i := range Nb {
			if i != Nb-1 {
				bodyDef.Position = box2d.Pos{X: posScalar(z + 0.25), Y: posScalar(y + cardHeight - 0.015)}
				bodyDef.Rotation = box2d.MakeRot(angle2)
				bodyID := s.worldID.CreateBody(&bodyDef)
				bodyID.CreatePolygonShape(&shapeDef, &cardBox)
			}

			bodyDef.Position = box2d.Pos{X: posScalar(z), Y: posScalar(y)}
			bodyDef.Rotation = box2d.MakeRot(angle1)
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &cardBox)

			z += 0.175

			bodyDef.Position = box2d.Pos{X: posScalar(z), Y: posScalar(y)}
			bodyDef.Rotation = box2d.MakeRot(angle0)
			bodyID = s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &cardBox)

			z += 0.175
		}
		y += cardHeight*2.0 - 0.03
		z0 += 0.175
		Nb--
	}
	return s
}

func init() {
	RegisterSample("Stacking", "Single Box", CreateSingleBox)
	RegisterSample("Stacking", "Tilted Stack", CreateTiltedStack)
	RegisterSample("Stacking", "Vertical Stack", CreateVerticalStack)
	RegisterSample("Stacking", "Circle Stack", CreateCircleStack)
	RegisterSample("Stacking", "Capsule Stack", CreateCapsuleStack)
	RegisterSample("Stacking", "Cliff", CreateCliff)
	RegisterSample("Stacking", "Arch", CreateArch)
	RegisterSample("Stacking", "Double Domino", CreateDoubleDomino)
	RegisterSample("Stacking", "Confined", CreateConfined)
	RegisterSample("Stacking", "Card House", CreateCardHouse)
}
