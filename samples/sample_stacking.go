// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const (
	tiltedStackColumns = 10
	tiltedStackRows    = 10
)

const (
	verticalStackMaxColumns = 10
	verticalStackMaxRows    = 80
	verticalStackMaxBullets = 8

	verticalStackCircleShape = 0
	verticalStackBoxShape    = 1
)

const (
	confinedGridCount = 25
	confinedMaxCount  = confinedGridCount * confinedGridCount
)

type SingleBox struct {
	Sample

	//exhaustruct:optional
	bodyID box2d.Body
}
type TiltedStack struct {
	Sample

	//exhaustruct:optional
	bodies [tiltedStackRows * tiltedStackColumns]box2d.Body
}

// This sample shows some aspects of Box2D continuous collision:
// - bullet dynamic bodies which support continuous collision with non-bullet dynamic bodies
// - prevention of chain reaction tunneling
// Try disabling continuous collision and firing a bullet. You might see a bullet push a
// through the static wall.
type VerticalStack struct {
	Sample

	//exhaustruct:optional
	bullets [verticalStackMaxBullets]box2d.Body
	//exhaustruct:optional
	bodies      [verticalStackMaxRows * verticalStackMaxColumns]box2d.Body
	columnCount int32
	rowCount    int32
	bulletCount int32
	shapeType   int32
	bulletType  int32
}
type circleStackEvent struct {
	indexA int
	indexB int
}

// A simple circle stack that also shows how to collect hit events.
type CircleStack struct {
	Sample

	//exhaustruct:optional
	events []circleStackEvent
}
type CapsuleStack struct {
	Sample
}
type Cliff struct {
	Sample

	//exhaustruct:optional
	bodyIDs [9]box2d.Body
	//exhaustruct:optional
	flip bool
}
type Arch struct {
	Sample
}
type DoubleDomino struct {
	Sample
}
type Confined struct {
	Sample

	//exhaustruct:optional
	row int
	//exhaustruct:optional
	column int
	//exhaustruct:optional
	count int
}

// From PEEL.
type CardHouse struct {
	Sample
}

func CreateSingleBox(context *SampleContext) Sampler {
	const (
		cameraHeight           = 2.5
		groundWidthInExtents   = 66.0
		initialHorizontalSpeed = 5.0
	)

	s := &SingleBox{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 3.5
	}

	extent := float32(1.0)

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	groundWidth := groundWidthInExtents * extent
	shapeDef := box2d.DefaultShapeDef()
	// shapeDef.friction = 0.5f;

	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -0.5 * 2.0 * groundWidth, Y: 0.0},
		Point2: box2d.Vec2{X: 0.5 * 2.0 * groundWidth, Y: 0.0},
	}
	groundID.CreateSegmentShape(&shapeDef, &segment)

	bodyDef.Type = box2d.DynamicBody

	box := box2d.MakeBox(extent, extent)
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
	bodyDef.LinearVelocity = box2d.Vec2{X: initialHorizontalSpeed, Y: 0.0}
	s.bodyID = s.worldID.CreateBody(&bodyDef)
	s.bodyID.CreatePolygonShape(&shapeDef, &box)

	return s
}

func (s *SingleBox) Step() {
	s.Sample.Step()

	position := s.bodyID.Position()
	s.DrawScreenTextLinef("(x, y) = (%.2g, %.2g)", position.X, position.Y)
}

func CreateTiltedStack(context *SampleContext) Sampler {
	const (
		cameraCenterX       = 7.5
		cameraCenterY       = 7.5
		groundHalfWidth     = 1000.0
		boxHalfExtent       = 0.45
		boxCornerRadius     = 0.05
		rowHorizontalOffset = 0.2
		columnSpacing       = 5.0
		firstRowHeight      = 0.5
	)

	s := &TiltedStack{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = 20.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(groundHalfWidth, 1.0)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	box := box2d.MakeRoundedBox(boxHalfExtent, boxHalfExtent, boxCornerRadius)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	shapeDef.Material.Friction = 0.3

	offset := float32(rowHorizontalOffset)
	dx := float32(columnSpacing)
	xroot := -0.5 * dx * (tiltedStackColumns - 1.0)

	for j := range tiltedStackColumns {
		x := xroot + float32(j)*dx

		for i := range tiltedStackRows {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody

			n := j*tiltedStackRows + i

			bodyDef.Position = box2d.Pos{
				X: posScalar(x + offset*float32(i)),
				Y: posScalar(firstRowHeight + 1.0*float32(i)),
			}
			bodyID := s.worldID.CreateBody(&bodyDef)

			s.bodies[n] = bodyID

			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	return s
}

func CreateVerticalStack(context *SampleContext) Sampler {
	const (
		cameraHeight    = 9.0
		wallX           = 10.0
		wallHeight      = 20.0
		groundHalfWidth = 30.0
		initialRowCount = 12
	)

	s := &VerticalStack{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		columnCount:     1,
		rowCount:        initialRowCount,
		bulletCount:     1,
		shapeType:       verticalStackBoxShape,
		bulletType:      verticalStackCircleShape,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: -7.0, Y: cameraHeight}
		s.context.camera.zoom = 14.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: wallX, Y: 0.0},
			Point2: box2d.Vec2{X: wallX, Y: wallHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		segment = box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.CreateStacks()

	return s
}

func (s *VerticalStack) CreateStacks() {
	const (
		boxHalfExtent   = 0.45
		boxCornerRadius = 0.05
		firstColumnX    = 8.0
		firstRowHeight  = 0.5
		circleRadius    = 0.5
	)

	for i := range s.bodies {
		if s.bodies[i] != (box2d.Body{}) {
			s.bodies[i].Destroy()
			s.bodies[i] = box2d.Body{}
		}
	}

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}

	box := box2d.MakeRoundedBox(boxHalfExtent, boxHalfExtent, boxCornerRadius)

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
	xroot := float32(firstColumnX)

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

			bodyDef.Position = box2d.Pos{
				X: posScalar(x + shift),
				Y: posScalar(firstRowHeight + 1.0*float32(i)),
			}
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
	const (
		bulletRadius     = 0.25
		bulletHalfExtent = 0.25
		bulletHeight     = 6.0
		minBulletSpeed   = 200.0
		maxBulletSpeed   = 300.0
	)

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: bulletRadius}
	box := box2d.MakeBox(bulletHalfExtent, bulletHalfExtent)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 4.0

	for i := range int(s.bulletCount) {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-26.7 - float32(i)), Y: bulletHeight}
		speed := scenes.RandomFloatRange(minBulletSpeed, maxBulletSpeed)
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

	if key == glfw.KeyB {
		s.FireBullets()

		consumed = true
	}

	if !consumed {
		s.Sample.Keyboard(key, action, mods)
	}
}

func (s *VerticalStack) DrawControls() bool {
	const controlWidthInFontSizes = 6.0
	imgui.PushItemWidth(controlWidthInFontSizes * imgui.FontSize())

	changed := false
	shapeTypes := []string{"Circle", "Box"}

	changed = changed ||
		imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))

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

func CreateCircleStack(context *SampleContext) Sampler {
	const (
		cameraHeight             = 5.0
		groundHalfWidth          = 10.0
		contactFrequencyFraction = 0.25
		contactBaseFrequency     = 360.0
		contactDampingRatio      = 10.0
		contactPushSpeed         = 3.0
		firstCircleHeight        = 0.75
		circleRadius             = 0.5
		densityIncrement         = 4.0
	)

	s := &CircleStack{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 6.0
	}

	shapeIndex := 0

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.UserData = shapeIndex
		shapeIndex++

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	s.worldID.SetGravity(box2d.Vec2{X: 0.0, Y: -20.0})
	s.worldID.SetContactTuning(
		contactFrequencyFraction*contactBaseFrequency,
		contactDampingRatio,
		contactPushSpeed,
	)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableHitEvents = true
	// shapeDef.rollingResistance = 0.2f;
	shapeDef.Material.Friction = 0.0
	shapeDef.Material.Restitution = 0.8

	y := float32(firstCircleHeight)

	for i := range 4 {
		bodyDef.Position.Y = posScalar(y)

		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef.UserData = shapeIndex
		shapeDef.Density = 1.0 + densityIncrement*float32(i)
		shapeIndex++

		bodyID.CreateCircleShape(&shapeDef, &circle)

		y += 1.25
	}

	return s
}

func (s *CircleStack) Step() {
	const hitPointSize = 10.0

	s.Sample.Step()

	for event := range s.worldID.ContactHitEvents() {
		indexA, ok := event.ShapeA.UserData().(int)
		if !ok {
			panic("circle stack shape A user data must contain its integer index")
		}

		indexB, ok := event.ShapeB.UserData().(int)
		if !ok {
			panic("circle stack shape B user data must contain its integer index")
		}

		s.draw.DrawPoint(event.Point, hitPointSize, box2d.ColorWhite)

		s.events = append(s.events, circleStackEvent{indexA, indexB})
	}

	for _, event := range s.events {
		s.DrawScreenTextLinef("%d, %d", event.indexA, event.indexB)
	}
}

func CreateCapsuleStack(context *SampleContext) Sampler {
	const (
		cameraHeight             = 5.0
		groundHalfWidth          = 10.0
		capsuleRadius            = 0.25
		capsuleHalfLengthInRadii = 4.0
		firstHeightInRadii       = 2.0
		rowSpacingInRadii        = 3.0
	)

	s := &CapsuleStack{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = 6.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		polygon := box2d.MakeBox(groundHalfWidth, 1.0)
		groundID.CreatePolygonShape(&shapeDef, &polygon)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	a := float32(capsuleRadius)
	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: -capsuleHalfLengthInRadii * a, Y: 0.0},
		Center2: box2d.Vec2{X: capsuleHalfLengthInRadii * a, Y: 0.0},
		Radius:  a,
	}

	shapeDef := box2d.DefaultShapeDef()

	// rolling resistance increases stacking stability
	// shapeDef.rollingResistance = 0.2f;

	y := firstHeightInRadii * a

	for range 20 {
		bodyDef.Position.Y = posScalar(y)
		// bodyDef.position.x += ( i & 1 ) == 1 ? -0.5f * a : 0.5f * a;
		// bodyDef.linearVelocity = { 0.0f, -10.0f };
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreateCapsuleShape(&shapeDef, &capsule)

		y += rowSpacingInRadii * a
	}

	return s
}

func CreateCliff(context *SampleContext) Sampler {
	const (
		cameraBaseZoom     = 25.0
		cameraZoomScale    = 0.5
		cameraHeight       = 5.0
		groundHalfWidth    = 100.0
		platformHeight     = 4.0
		platformHalfWidth  = 3.0
		platformHalfHeight = 0.5
		capsuleStartX      = 8.5
		capsuleEndX        = 13.5
		capsuleRadius      = 0.5
	)

	s := &Cliff{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			groundHalfWidth,
			1.0,
			box2d.Vec2{X: 0.0, Y: -1.0},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -14.0, Y: platformHeight},
			Point2: box2d.Vec2{X: -8.0, Y: platformHeight},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box = box2d.MakeOffsetBox(
			platformHalfWidth,
			platformHalfHeight,
			box2d.Vec2{X: 0.0, Y: platformHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: capsuleStartX, Y: platformHeight},
			Center2: box2d.Vec2{X: capsuleEndX, Y: platformHeight},
			Radius:  capsuleRadius,
		}
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
	const (
		capsuleHalfLength  = 0.25
		capsuleRadius      = 0.25
		squareHalfExtent   = 0.5
		circleRadius       = 0.5
		capsuleSpeed       = 2.0
		flipOffset         = 4.0
		leftCapsuleHeight  = 4.25
		centerCapsuleX     = 2.0
		outerCapsuleX      = 13.0
		outerCapsuleHeight = 4.75
		squareSpeed        = 2.5
		leftSquareHeight   = 4.5
		upperBodyHeight    = 5.0
		rightSquareX       = 11.0
		circleSpeed        = 1.5
		leftCircleX        = -13.0
		centerCircleX      = -2.0
		rightCircleX       = 9.0
	)

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

	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
		Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
		Radius:  capsuleRadius,
	}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
	square := box2d.MakeSquare(squareHalfExtent)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	{
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.01
		bodyDef.LinearVelocity = box2d.Vec2{X: capsuleSpeed * sign, Y: 0.0}

		offset := posScalar(0.0)
		if s.flip {
			offset = -flipOffset
		}

		bodyDef.Position = box2d.Pos{X: -9.0 + offset, Y: leftCapsuleHeight}
		s.bodyIDs[0] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[0].CreateCapsuleShape(&shapeDef, &capsule)

		bodyDef.Position = box2d.Pos{X: centerCapsuleX + offset, Y: outerCapsuleHeight}
		s.bodyIDs[1] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[1].CreateCapsuleShape(&shapeDef, &capsule)

		bodyDef.Position = box2d.Pos{X: outerCapsuleX + offset, Y: outerCapsuleHeight}
		s.bodyIDs[2] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[2].CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.01
		bodyDef.LinearVelocity = box2d.Vec2{X: squareSpeed * sign, Y: 0.0}

		bodyDef.Position = box2d.Pos{X: -11.0, Y: leftSquareHeight}
		s.bodyIDs[3] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[3].CreatePolygonShape(&shapeDef, &square)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: upperBodyHeight}
		s.bodyIDs[4] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[4].CreatePolygonShape(&shapeDef, &square)

		bodyDef.Position = box2d.Pos{X: rightSquareX, Y: upperBodyHeight}
		s.bodyIDs[5] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[5].CreatePolygonShape(&shapeDef, &square)
	}

	{
		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.2
		bodyDef.LinearVelocity = box2d.Vec2{X: circleSpeed * sign, Y: 0.0}

		offset := posScalar(0.0)
		if s.flip {
			offset = flipOffset
		}

		bodyDef.Position = box2d.Pos{X: leftCircleX + offset, Y: leftSquareHeight}
		s.bodyIDs[6] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[6].CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: centerCircleX + offset, Y: upperBodyHeight}
		s.bodyIDs[7] = s.worldID.CreateBody(&bodyDef)
		s.bodyIDs[7].CreateCircleShape(&shapeDef, &circle)

		bodyDef.Position = box2d.Pos{X: rightCircleX + offset, Y: upperBodyHeight}
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

func CreateArch(context *SampleContext) Sampler {
	const (
		cameraHeight        = 8.0
		cameraBaseZoom      = 25.0
		cameraZoomScale     = 0.35
		archScale           = 0.25
		groundHalfWidth     = 100.0
		loadHalfWidth       = 2.0
		loadHalfHeight      = 0.5
		innerBaseX          = 16.0
		innerLowerFootX     = 14.93803712795643
		innerLowerFootY     = 5.133601056842984
		innerFootX          = 13.79871746027416
		innerFootY          = 10.24928069555078
		innerLowerFlankX    = 12.56252963284711
		innerLowerFlankY    = 15.34107019122473
		innerMidFlankX      = 11.20040987372525
		innerMidFlankY      = 20.39856541571217
		innerUpperFlankX    = 9.66521217819836
		innerUpperFlankY    = 25.40369899225096
		innerShoulderX      = 7.87179930638133
		innerShoulderY      = 30.3179337000085
		innerUpperShoulderX = 5.635199558196225
		innerUpperShoulderY = 35.03820717801641
		innerCrownX         = 2.405937953536585
		innerCrownY         = 39.09554102558315
		outerBaseX          = 24.0
		outerLowerFootX     = 22.33619528222415
		outerLowerFootY     = 6.02299846205841
		outerFootX          = 20.54936888969905
		outerFootY          = 12.00964361211476
		outerLowerFlankX    = 18.60854610798073
		outerLowerFlankY    = 17.9470321677465
		outerMidFlankX      = 16.46769273811807
		outerMidFlankY      = 23.81367936585418
		outerUpperFlankX    = 14.05325025774858
		outerUpperFlankY    = 29.57079353071012
		outerShoulderX      = 11.23551045834022
		outerShoulderY      = 35.13775818285372
		outerUpperShoulderX = 7.752568160730571
		outerUpperShoulderY = 40.30450679009583
		outerCrownX         = 3.016931552701656
		outerCrownY         = 44.28891593799322
	)

	s := &Arch{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	ps1 := [9]box2d.Vec2{
		{X: innerBaseX, Y: 0.0},
		{X: innerLowerFootX, Y: innerLowerFootY},
		{X: innerFootX, Y: innerFootY},
		{X: innerLowerFlankX, Y: innerLowerFlankY},
		{X: innerMidFlankX, Y: innerMidFlankY},
		{X: innerUpperFlankX, Y: innerUpperFlankY},
		{X: innerShoulderX, Y: innerShoulderY},
		{X: innerUpperShoulderX, Y: innerUpperShoulderY},
		{X: innerCrownX, Y: innerCrownY},
	}

	ps2 := [9]box2d.Vec2{
		{X: outerBaseX, Y: 0.0},
		{X: outerLowerFootX, Y: outerLowerFootY},
		{X: outerFootX, Y: outerFootY},
		{X: outerLowerFlankX, Y: outerLowerFlankY},
		{X: outerMidFlankX, Y: outerMidFlankY},
		{X: outerUpperFlankX, Y: outerUpperFlankY},
		{X: outerShoulderX, Y: outerShoulderY},
		{X: outerUpperShoulderX, Y: outerUpperShoulderY},
		{X: outerCrownX, Y: outerCrownY},
	}

	scale := float32(archScale)
	for i := range 9 {
		ps1[i] = ps1[i].Scale(scale)
		ps2[i] = ps2[i].Scale(scale)
	}

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.6

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
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
		box := box2d.MakeBox(loadHalfWidth, loadHalfHeight)
		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(0.5 + ps2[8].Y + 1.0*float32(i))}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func CreateDoubleDomino(context *SampleContext) Sampler {
	const (
		cameraHeight     = 4.0
		cameraBaseZoom   = 25.0
		cameraZoomScale  = 0.25
		groundHalfWidth  = 100.0
		dominoHalfWidth  = 0.125
		dominoHalfHeight = 0.5
		initialImpulse   = 0.2
	)

	s := &DoubleDomino{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(groundHalfWidth, 1.0)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	box := box2d.MakeBox(dominoHalfWidth, dominoHalfHeight)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.6
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	count := 15

	x := float32(-0.5 * float32(count))
	for i := range count {
		bodyDef.Position = box2d.Pos{X: posScalar(x), Y: dominoHalfHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		if i == 0 {
			bodyID.ApplyLinearImpulse(
				box2d.Vec2{X: initialImpulse, Y: 0.0},
				box2d.Pos{X: posScalar(x), Y: 1.0},
				true,
			)
		}

		x += 1.0
	}

	return s
}

func CreateConfined(context *SampleContext) Sampler {
	const (
		cameraHeight    = 10.0
		cameraBaseZoom  = 25.0
		cameraZoomScale = 0.5
		wallX           = 10.5
		wallTop         = 20.5
		wallRadius      = 0.5
		gridBottom      = 1.5
		circleRadius    = 0.5
	)

	s := &Confined{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()

		var capsule box2d.Capsule

		capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: -wallX, Y: 0.0},
			Center2: box2d.Vec2{X: wallX, Y: 0.0},
			Radius:  wallRadius,
		}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
		capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: -wallX, Y: 0.0},
			Center2: box2d.Vec2{X: -wallX, Y: wallTop},
			Radius:  wallRadius,
		}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
		capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: wallX, Y: 0.0},
			Center2: box2d.Vec2{X: wallX, Y: wallTop},
			Radius:  wallRadius,
		}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
		capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: -wallX, Y: wallTop},
			Center2: box2d.Vec2{X: wallX, Y: wallTop},
			Radius:  wallRadius,
		}
		groundID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	s.row = 0
	s.column = 0
	s.count = 0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.0

	shapeDef := box2d.DefaultShapeDef()
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}

	for s.count < confinedMaxCount {
		s.row = 0
		for range confinedGridCount {
			x := -8.75 + float32(s.column)*18.0/confinedGridCount
			y := gridBottom + float32(s.row)*18.0/confinedGridCount

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

func CreateCardHouse(context *SampleContext) Sampler {
	const (
		cameraCenterX       = 0.75
		cameraCenterY       = 0.9
		cameraBaseZoom      = 25.0
		cameraZoomScale     = 0.05
		groundHalfWidth     = 40.0
		groundHalfHeight    = 2.0
		cardHalfHeight      = 0.2
		cardHalfThickness   = 0.001
		cardTiltDegrees     = 25.0
		degreesPerHalfTurn  = 180.0
		halfTurnFraction    = 0.5
		firstRowOffset      = 0.02
		crossCardOffsetX    = 0.25
		crossCardOffsetY    = 0.015
		rowHeightMultiplier = 2.0
		rowOverlap          = 0.03
	)

	s := &CardHouse{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -2.0}
	groundID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = 0.7

	groundBox := box2d.MakeBox(groundHalfWidth, groundHalfHeight)
	groundID.CreatePolygonShape(&shapeDef, &groundBox)

	cardHeight := float32(cardHalfHeight)
	cardThickness := float32(cardHalfThickness)

	angle0 := cardTiltDegrees * box2d.Pi / degreesPerHalfTurn
	angle1 := -cardTiltDegrees * box2d.Pi / degreesPerHalfTurn
	angle2 := halfTurnFraction * box2d.Pi

	cardBox := box2d.MakeBox(cardThickness, cardHeight)
	bodyDef.Type = box2d.DynamicBody

	remainingPairs := 5
	z0 := float32(0.0)
	y := cardHeight - firstRowOffset

	for remainingPairs != 0 {
		z := z0

		for i := range remainingPairs {
			if i != remainingPairs-1 {
				bodyDef.Position = box2d.Pos{
					X: posScalar(z + crossCardOffsetX),
					Y: posScalar(y + cardHeight - crossCardOffsetY),
				}
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

		y += cardHeight*rowHeightMultiplier - rowOverlap
		z0 += 0.175
		remainingPairs--
	}

	return s
}

func registerStackingSamples() {
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
