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
	shapeChoiceCircle  = 0
	shapeChoiceCapsule = 1
	shapeChoiceBox     = 2
)

var shapeChoiceNames = []string{"Circle", "Capsule", "Box"}

type ChainShape struct {
	Sample

	groundID  box2d.Body
	bodyID    box2d.Body
	chainID   box2d.Chain
	shapeType int32
	shapeID   box2d.Shape
	material  box2d.SurfaceMaterial
}

func CreateChainShape(context *SampleContext) Sampler {
	s := &ChainShape{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 1.75
	}

	s.groundID = box2d.Body{}
	s.bodyID = box2d.Body{}
	s.chainID = box2d.Chain{}
	s.shapeID = box2d.Shape{}
	s.shapeType = shapeChoiceCircle

	s.material = box2d.DefaultSurfaceMaterial()
	s.material.Friction = 0.2
	s.material.CustomColor = uint32(box2d.ColorSteelBlue)
	s.material.UserMaterialID = 42

	s.CreateScene()
	s.Launch()
	return s
}

func (s *ChainShape) CreateScene() {
	if s.groundID != (box2d.Body{}) {
		s.groundID.Destroy()
	}

	// https://betravis.github.io/shape-tools/path-to-polygon/
	points := []box2d.Vec2{
		{X: -56.885498, Y: 12.8985004},
		{X: -56.885498, Y: 16.2057495},
		{X: 56.885498, Y: 16.2057495},
		{X: 56.885498, Y: -16.2057514},
		{X: 51.5935059, Y: -16.2057514},
		{X: 43.6559982, Y: -10.9139996},
		{X: 35.7184982, Y: -10.9139996},
		{X: 27.7809982, Y: -10.9139996},
		{X: 21.1664963, Y: -14.2212505},
		{X: 11.9059982, Y: -16.2057514},
		{X: 0, Y: -16.2057514},
		{X: -10.5835037, Y: -14.8827496},
		{X: -17.1980019, Y: -13.5597477},
		{X: -21.1665001, Y: -12.2370014},
		{X: -25.1355019, Y: -9.5909977},
		{X: -31.75, Y: -3.63799858},
		{X: -38.3644981, Y: 6.2840004},
		{X: -42.3334999, Y: 9.59125137},
		{X: -47.625, Y: 11.5755005},
	}

	chainDef := box2d.DefaultChainDef()
	chainDef.Points = points
	chainDef.Materials = []box2d.SurfaceMaterial{s.material}
	chainDef.IsLoop = true

	bodyDef := box2d.DefaultBodyDef()
	s.groundID = s.worldID.CreateBody(&bodyDef)

	s.chainID = s.groundID.CreateChain(&chainDef)
}

func (s *ChainShape) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -55.0, Y: 13.5}
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material = s.material

	if s.shapeType == shapeChoiceCircle {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
	} else if s.shapeType == shapeChoiceCapsule {
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -0.5, Y: 0.0},
			Center2: box2d.Vec2{X: 0.5, Y: 0.0},
			Radius:  0.25,
		}
		s.shapeID = s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		h := float32(0.5)
		box := box2d.MakeBox(h, h)
		s.shapeID = s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	s.stepCount = 0
}

func (s *ChainShape) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.ComboStrarr("Shape", &s.shapeType, shapeChoiceNames, int32(len(shapeChoiceNames))) {
		s.Launch()
	}

	if imgui.SliderFloatV("Friction", &s.material.Friction, 0.0, 1.0, "%.2f", 0) {
		s.shapeID.SetSurfaceMaterial(&s.material)
		s.chainID.SetAllSurfaceMaterials(&s.material)
	}

	if imgui.SliderFloatV("Restitution", &s.material.Restitution, 0.0, 2.0, "%.1f", 0) {
		s.shapeID.SetSurfaceMaterial(&s.material)
	}

	imgui.PopItemWidth()

	if imgui.Button("Launch") {
		s.Launch()
	}

	return true
}

func (s *ChainShape) Step() {
	s.Sample.Step()

	s.draw.DrawLine(box2d.Pos{}, box2d.Vec2{X: 0.5, Y: 0.0}.Pos(), box2d.ColorRed)
	s.draw.DrawLine(box2d.Pos{}, box2d.Vec2{X: 0.0, Y: 0.5}.Pos(), box2d.ColorGreen)
}

const (
	chainSegmentSegmentCount = 32
	chainSegmentPointCount   = chainSegmentSegmentCount + 3
)

type ChainSegmentShape struct {
	Sample

	bodyID        box2d.Body
	shapeType     int32
	segmentShapes [chainSegmentSegmentCount]box2d.Shape
	points        [chainSegmentPointCount]box2d.Vec2
	mutateIndex   int
}

func CreateChainSegmentShape(context *SampleContext) Sampler {
	s := &ChainSegmentShape{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 25.0 * 1.0
	}

	s.bodyID = box2d.Body{}
	s.shapeType = shapeChoiceCircle
	s.mutateIndex = 0

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		// Walk right-to-left so the right-perpendicular normal of (point2 - point1) points up.
		for i := range chainSegmentPointCount {
			x := 25.0 - 50.0*float32(i)/(chainSegmentPointCount-1)
			y := 1.5 * sinf(0.18*x)
			s.points[i] = box2d.Vec2{X: x, Y: y}
		}

		shapeDef := box2d.DefaultShapeDef()
		for i := range chainSegmentSegmentCount {
			var chainSegment box2d.ChainSegment
			chainSegment.Ghost1 = s.points[i]
			chainSegment.Segment.Point1 = s.points[i+1]
			chainSegment.Segment.Point2 = s.points[i+2]
			chainSegment.Ghost2 = s.points[i+3]
			chainSegment.ChainID = -1
			s.segmentShapes[i] = groundID.CreateChainSegmentShape(&shapeDef, &chainSegment)
		}
	}

	s.Launch()
	return s
}

func (s *ChainSegmentShape) Launch() {
	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -18.0, Y: 5.0}
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	if s.shapeType == shapeChoiceCircle {
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.25}
		s.bodyID.CreateCircleShape(&shapeDef, &circle)
	} else if s.shapeType == shapeChoiceCapsule {
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -0.5, Y: 0.0},
			Center2: box2d.Vec2{X: 0.5, Y: 0.0},
			Radius:  0.25,
		}
		s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	} else {
		box := box2d.MakeSquare(0.5)
		s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}
}

func (s *ChainSegmentShape) Mutate() {
	// Get an index in [1,pointCount - 2]
	// index 0 and pointCount-1 are ghost vertices and are not mutated
	index := s.mutateIndex + 1

	s.mutateIndex++
	if s.mutateIndex == chainSegmentSegmentCount {
		s.mutateIndex = 0
	}

	s.points[index].Y += 0.25

	var cs box2d.ChainSegment
	cs.Ghost1 = s.points[index-1]
	cs.Segment.Point1 = s.points[index]
	cs.Segment.Point2 = s.points[index+1]
	cs.Ghost2 = s.points[index+2]
	cs.ChainID = -1

	s.segmentShapes[index-1].SetChainSegment(&cs)

	if index-1 > 0 {
		var cs2 box2d.ChainSegment
		cs2.Ghost1 = s.points[index-2]
		cs2.Segment.Point1 = s.points[index-1]
		cs2.Segment.Point2 = s.points[index]
		cs2.Ghost2 = s.points[index+1]
		cs2.ChainID = -1
		s.segmentShapes[index-2].SetChainSegment(&cs2)
	}

	if index+1 < chainSegmentPointCount-2 {
		var cs3 box2d.ChainSegment
		cs3.Ghost1 = s.points[index]
		cs3.Segment.Point1 = s.points[index+1]
		cs3.Segment.Point2 = s.points[index+2]
		cs3.Ghost2 = s.points[index+3]
		cs3.ChainID = -1
		s.segmentShapes[index].SetChainSegment(&cs3)
	}
}

func (s *ChainSegmentShape) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.ComboStrarr("Shape", &s.shapeType, shapeChoiceNames, int32(len(shapeChoiceNames))) {
		s.Launch()
	}

	imgui.PopItemWidth()

	if imgui.Button("Launch") {
		s.Launch()
	}

	if imgui.Button("Mutate") {
		s.Mutate()
	}

	return true
}

// This sample shows how careful creation of compound shapes leads to better simulation and avoids
// objects getting stuck.
// This also shows how to get the combined AABB for the body.
type CompoundShapes struct {
	Sample

	table1ID      box2d.Body
	table2ID      box2d.Body
	ship1ID       box2d.Body
	ship2ID       box2d.Body
	drawBodyAABBs bool
}

func CreateCompoundShapes(context *SampleContext) Sampler {
	s := &CompoundShapes{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 6.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: 50.0, Y: 0.0},
			Point2: box2d.Vec2{X: -50.0, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Table 1
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -15.0, Y: 1.0}
		s.table1ID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		top := box2d.MakeOffsetBox(3.0, 0.5, box2d.Vec2{X: 0.0, Y: 3.5}, box2d.RotIdentity)
		leftLeg := box2d.MakeOffsetBox(0.5, 1.5, box2d.Vec2{X: -2.5, Y: 1.5}, box2d.RotIdentity)
		rightLeg := box2d.MakeOffsetBox(0.5, 1.5, box2d.Vec2{X: 2.5, Y: 1.5}, box2d.RotIdentity)

		s.table1ID.CreatePolygonShape(&shapeDef, &top)
		s.table1ID.CreatePolygonShape(&shapeDef, &leftLeg)
		s.table1ID.CreatePolygonShape(&shapeDef, &rightLeg)
	}

	// Table 2
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -5.0, Y: 1.0}
		s.table2ID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		top := box2d.MakeOffsetBox(3.0, 0.5, box2d.Vec2{X: 0.0, Y: 3.5}, box2d.RotIdentity)
		leftLeg := box2d.MakeOffsetBox(0.5, 2.0, box2d.Vec2{X: -2.5, Y: 2.0}, box2d.RotIdentity)
		rightLeg := box2d.MakeOffsetBox(0.5, 2.0, box2d.Vec2{X: 2.5, Y: 2.0}, box2d.RotIdentity)

		s.table2ID.CreatePolygonShape(&shapeDef, &top)
		s.table2ID.CreatePolygonShape(&shapeDef, &leftLeg)
		s.table2ID.CreatePolygonShape(&shapeDef, &rightLeg)
	}

	// Spaceship 1
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 5.0, Y: 1.0}
		s.ship1ID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		vertices := make([]box2d.Vec2, 3)

		vertices[0] = box2d.Vec2{X: -2.0, Y: 0.0}
		vertices[1] = box2d.Vec2{X: 0.0, Y: 4.0 / 3.0}
		vertices[2] = box2d.Vec2{X: 0.0, Y: 4.0}
		hull := box2d.ComputeHull(vertices)
		left := box2d.MakePolygon(&hull, 0.0)

		vertices[0] = box2d.Vec2{X: 2.0, Y: 0.0}
		vertices[1] = box2d.Vec2{X: 0.0, Y: 4.0 / 3.0}
		vertices[2] = box2d.Vec2{X: 0.0, Y: 4.0}
		hull = box2d.ComputeHull(vertices)
		right := box2d.MakePolygon(&hull, 0.0)

		s.ship1ID.CreatePolygonShape(&shapeDef, &left)
		s.ship1ID.CreatePolygonShape(&shapeDef, &right)
	}

	// Spaceship 2
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 15.0, Y: 1.0}
		s.ship2ID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		vertices := make([]box2d.Vec2, 3)

		vertices[0] = box2d.Vec2{X: -2.0, Y: 0.0}
		vertices[1] = box2d.Vec2{X: 1.0, Y: 2.0}
		vertices[2] = box2d.Vec2{X: 0.0, Y: 4.0}
		hull := box2d.ComputeHull(vertices)
		left := box2d.MakePolygon(&hull, 0.0)

		vertices[0] = box2d.Vec2{X: 2.0, Y: 0.0}
		vertices[1] = box2d.Vec2{X: -1.0, Y: 2.0}
		vertices[2] = box2d.Vec2{X: 0.0, Y: 4.0}
		hull = box2d.ComputeHull(vertices)
		right := box2d.MakePolygon(&hull, 0.0)

		s.ship2ID.CreatePolygonShape(&shapeDef, &left)
		s.ship2ID.CreatePolygonShape(&shapeDef, &right)
	}

	s.drawBodyAABBs = false
	return s
}

func (s *CompoundShapes) Spawn() {
	// Table 1 obstruction
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = s.table1ID.Position()
		bodyDef.Rotation = s.table1ID.Rotation()
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(4.0, 0.1, box2d.Vec2{X: 0.0, Y: 3.0}, box2d.RotIdentity)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	// Table 2 obstruction
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = s.table2ID.Position()
		bodyDef.Rotation = s.table2ID.Rotation()
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(4.0, 0.1, box2d.Vec2{X: 0.0, Y: 3.0}, box2d.RotIdentity)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	// Ship 1 obstruction
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = s.ship1ID.Position()
		bodyDef.Rotation = s.ship1ID.Rotation()
		// bodyDef.gravityScale = 0.0f;
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 2.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	// Ship 2 obstruction
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = s.ship2ID.Position()
		bodyDef.Rotation = s.ship2ID.Rotation()
		// bodyDef.gravityScale = 0.0f;
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 2.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
}

func (s *CompoundShapes) DrawControls() bool {
	if imgui.Button("Intrude") {
		s.Spawn()
	}

	imgui.Checkbox("Body AABBs", &s.drawBodyAABBs)

	return true
}

func (s *CompoundShapes) Step() {
	s.Sample.Step()

	if s.drawBodyAABBs {
		aabb := s.table1ID.ComputeAABB()
		s.draw.DrawBounds(aabb, box2d.ColorYellow)

		aabb = s.table2ID.ComputeAABB()
		s.draw.DrawBounds(aabb, box2d.ColorYellow)

		aabb = s.ship1ID.ComputeAABB()
		s.draw.DrawBounds(aabb, box2d.ColorYellow)

		aabb = s.ship2ID.ComputeAABB()
		s.draw.DrawBounds(aabb, box2d.ColorYellow)
	}
}

const (
	shapeFilterGround  = 0x00000001
	shapeFilterTeam1   = 0x00000002
	shapeFilterTeam2   = 0x00000004
	shapeFilterTeam3   = 0x00000008
	shapeFilterAllBits = 0xFFFFFFFF
)

type ShapeFilter struct {
	Sample

	player1ID box2d.Body
	player2ID box2d.Body
	player3ID box2d.Body

	shape1ID box2d.Shape
	shape2ID box2d.Shape
	shape3ID box2d.Shape
}

func CreateShapeFilter(context *SampleContext) Sampler {
	s := &ShapeFilter{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 0.5
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = shapeFilterGround
		shapeDef.Filter.MaskBits = shapeFilterAllBits

		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		bodyDef.Position = box2d.Pos{X: 0.0, Y: 2.0}
		s.player1ID = s.worldID.CreateBody(&bodyDef)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: 5.0}
		s.player2ID = s.worldID.CreateBody(&bodyDef)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: 8.0}
		s.player3ID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(2.0, 1.0)

		shapeDef := box2d.DefaultShapeDef()

		shapeDef.Filter.CategoryBits = shapeFilterTeam1
		shapeDef.Filter.MaskBits = shapeFilterGround | shapeFilterTeam2 | shapeFilterTeam3
		s.shape1ID = s.player1ID.CreatePolygonShape(&shapeDef, &box)

		shapeDef.Filter.CategoryBits = shapeFilterTeam2
		shapeDef.Filter.MaskBits = shapeFilterGround | shapeFilterTeam1 | shapeFilterTeam3
		s.shape2ID = s.player2ID.CreatePolygonShape(&shapeDef, &box)

		shapeDef.Filter.CategoryBits = shapeFilterTeam3
		shapeDef.Filter.MaskBits = shapeFilterGround | shapeFilterTeam1 | shapeFilterTeam2
		s.shape3ID = s.player3ID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

func (s *ShapeFilter) DrawControls() bool {
	imgui.Text("Player 1 Collides With")
	filter1 := s.shape1ID.Filter()
	team2 := (filter1.MaskBits & shapeFilterTeam2) == shapeFilterTeam2
	if imgui.Checkbox("Team 2##1", &team2) {
		if team2 {
			filter1.MaskBits |= shapeFilterTeam2
		} else {
			filter1.MaskBits &^= shapeFilterTeam2
		}

		s.shape1ID.SetFilter(filter1)
	}

	team3 := (filter1.MaskBits & shapeFilterTeam3) == shapeFilterTeam3
	if imgui.Checkbox("Team 3##1", &team3) {
		if team3 {
			filter1.MaskBits |= shapeFilterTeam3
		} else {
			filter1.MaskBits &^= shapeFilterTeam3
		}

		s.shape1ID.SetFilter(filter1)
	}

	return true
}

func (s *ShapeFilter) Step() {
	s.Sample.Step()

	p1 := s.player1ID.Position()
	s.draw.DrawString(s.camera, box2d.Pos{X: p1.X - 0.5, Y: p1.Y}, box2d.ColorWhite, "player 1")

	p2 := s.player2ID.Position()
	s.draw.DrawString(s.camera, box2d.Pos{X: p2.X - 0.5, Y: p2.Y}, box2d.ColorWhite, "player 2")

	p3 := s.player3ID.Position()
	s.draw.DrawString(s.camera, box2d.Pos{X: p3.X - 0.5, Y: p3.Y}, box2d.ColorWhite, "player 3")
}

const customFilterCount = 10

// This shows how to use custom filtering
type CustomFilter struct {
	Sample

	bodyIDs  [customFilterCount]box2d.Body
	shapeIDs [customFilterCount]box2d.Shape
}

func CreateCustomFilter(context *SampleContext) Sampler {
	s := &CustomFilter{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 10.0
	}

	// Register custom filter
	s.worldID.SetCustomFilterCallback(s.ShouldCollide)

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -40.0, Y: 0.0},
			Point2: box2d.Vec2{X: 40.0, Y: 0.0},
		}

		shapeDef := box2d.DefaultShapeDef()

		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableCustomFiltering = true
	box := box2d.MakeSquare(1.0)
	x := posScalar(-customFilterCount)

	for i := range customFilterCount {
		bodyDef.Position = box2d.Pos{X: x, Y: 5.0}
		s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)

		shapeDef.UserData = i + 1
		s.shapeIDs[i] = s.bodyIDs[i].CreatePolygonShape(&shapeDef, &box)
		x += 2.0
	}
	return s
}

func (s *CustomFilter) Step() {
	s.DrawScreenTextLine("Custom filter disables collision between odd and even shapes")

	s.Sample.Step()

	for i := range customFilterCount {
		p := s.bodyIDs[i].Position()
		s.draw.DrawString(s.camera, box2d.Pos{X: p.X, Y: p.Y}, box2d.ColorWhite, "%d", i)
	}
}

func (s *CustomFilter) ShouldCollide(shapeIDA, shapeIDB box2d.Shape) bool {
	userDataA := shapeIDA.UserData()
	userDataB := shapeIDB.UserData()

	if userDataA == nil || userDataB == nil {
		return true
	}

	indexA := userDataA.(int)
	indexB := userDataB.(int)

	return ((indexA & 1) + (indexB & 1)) != 1
}

type Friction struct {
	Sample
}

func CreateFriction(context *SampleContext) Sampler {
	s := &Friction{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 14.0}
		s.context.camera.zoom = 25.0 * 0.6
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.2

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -40.0, Y: 0.0},
			Point2: box2d.Vec2{X: 40.0, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(13.0, 0.25, box2d.Vec2{X: -4.0, Y: 22.0}, box2d.MakeRot(-0.25))
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(0.25, 1.0, box2d.Vec2{X: 10.5, Y: 19.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(13.0, 0.25, box2d.Vec2{X: 4.0, Y: 14.0}, box2d.MakeRot(0.25))
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(0.25, 1.0, box2d.Vec2{X: -10.5, Y: 11.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(13.0, 0.25, box2d.Vec2{X: -4.0, Y: 6.0}, box2d.MakeRot(-0.25))
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		box := box2d.MakeBox(0.5, 0.5)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 25.0

		friction := [5]float32{0.75, 0.5, 0.35, 0.1, 0.0}

		for i := range 5 {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(-15.0 + 4.0*float32(i)), Y: 28.0}
			bodyID := s.worldID.CreateBody(&bodyDef)

			shapeDef.Material.Friction = friction[i]
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}
	return s
}

type RollingResistance struct {
	Sample

	resistScale float32
	lift        float32
}

func CreateRollingResistance(context *SampleContext) Sampler {
	s := &RollingResistance{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 5.0, Y: 20.0}
		s.context.camera.zoom = 27.5
	}

	s.lift = 0.0
	s.resistScale = 0.02
	s.CreateScene()
	return s
}

func (s *RollingResistance) CreateScene() {
	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: 0.5}

	shapeDef := box2d.DefaultShapeDef()

	for i := range 20 {
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -40.0, Y: 2.0 * float32(i)},
			Point2: box2d.Vec2{X: 40.0, Y: 2.0*float32(i) + s.lift},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -39.5, Y: posScalar(2.0*float32(i) + 0.75)}
		bodyDef.AngularVelocity = -10.0
		bodyDef.LinearVelocity = box2d.Vec2{X: 5.0, Y: 0.0}

		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef.Material.RollingResistance = s.resistScale * float32(i)
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
}

func (s *RollingResistance) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.Key1:
		s.lift = 0.0
		s.CreateWorld()
		s.CreateScene()

	case glfw.Key2:
		s.lift = 5.0
		s.CreateWorld()
		s.CreateScene()

	case glfw.Key3:
		s.lift = -5.0
		s.CreateWorld()
		s.CreateScene()

	default:
		s.Sample.Keyboard(key, action, mods)
	}
}

func (s *RollingResistance) Step() {
	s.Sample.Step()

	for i := range 20 {
		s.draw.DrawString(
			s.camera,
			box2d.Pos{X: -41.5, Y: posScalar(2.0*float32(i) + 1.0)},
			box2d.ColorWhite,
			"%.2f",
			s.resistScale*float32(i),
		)
	}
}

type ConveyorBelt struct {
	Sample
}

func CreateConveyorBelt(context *SampleContext) Sampler {
	s := &ConveyorBelt{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 7.5}
		s.context.camera.zoom = 12.0
	}

	// Ground
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -20.0, Y: 0.0},
			Point2: box2d.Vec2{X: 20.0, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Platform
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -5.0, Y: 5.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeRoundedBox(10.0, 0.25, 0.25)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.8
		shapeDef.Material.TangentSpeed = 2.0

		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	// Boxes
	shapeDef := box2d.DefaultShapeDef()
	cube := box2d.MakeSquare(0.5)
	for i := range 5 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-10.0 + 2.0*float32(i)), Y: 7.0}
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreatePolygonShape(&shapeDef, &cube)
	}
	return s
}

const tangentSpeedTotalCount = 200

type TangentSpeed struct {
	Sample

	bodyIDs           []box2d.Body
	friction          float32
	rollingResistance float32
}

func CreateTangentSpeed(context *SampleContext) Sampler {
	s := &TangentSpeed{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 60.0, Y: -15.0}
		s.context.camera.zoom = 38.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		path := "m 613.8334,185.20833 -42.33338,0 h -37.04166 l -34.39581,0 -29.10417,-2.64583 -26.45834,-7.9375 " +
			"-26.45833,-13.22917 -23.81251,-21.16666 h -13.22916 v 44.97916 H 68.791712 V 0 h -21.16671 v " +
			"206.375 l 566.208398,-1e-5 z"

		offset := box2d.Vec2{X: -47.375002, Y: 0.25}

		scale := float32(0.2)
		var points [20]box2d.Vec2
		count := ParsePath(path, offset, points[:], scale, true)

		var materials [20]box2d.SurfaceMaterial
		for i := range 20 {
			materials[i].Friction = 0.6
		}

		materials[0].TangentSpeed = -10.0
		materials[0].CustomColor = uint32(box2d.ColorDarkBlue)
		materials[1].TangentSpeed = -20.0
		materials[1].CustomColor = uint32(box2d.ColorDarkCyan)
		materials[2].TangentSpeed = -30.0
		materials[2].CustomColor = uint32(box2d.ColorDarkGoldenRod)
		materials[3].TangentSpeed = -40.0
		materials[3].CustomColor = uint32(box2d.ColorDarkGray)
		materials[4].TangentSpeed = -50.0
		materials[4].CustomColor = uint32(box2d.ColorDarkGreen)
		materials[5].TangentSpeed = -60.0
		materials[5].CustomColor = uint32(box2d.ColorDarkKhaki)
		materials[6].TangentSpeed = -70.0
		materials[6].CustomColor = uint32(box2d.ColorDarkMagenta)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true
		chainDef.Materials = materials[:count]

		groundID.CreateChain(&chainDef)

		s.friction = 0.6
		s.rollingResistance = 0.3
	}
	return s
}

func (s *TangentSpeed) DropBall() box2d.Body {
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 110.0, Y: -30.0}
	bodyID := s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.Friction = s.friction
	shapeDef.Material.RollingResistance = s.rollingResistance
	bodyID.CreateCircleShape(&shapeDef, &circle)
	return bodyID
}

func (s *TangentSpeed) Reset() {
	for _, bodyID := range s.bodyIDs {
		bodyID.Destroy()
	}

	s.bodyIDs = s.bodyIDs[:0]
}

func (s *TangentSpeed) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.SliderFloatV("Friction", &s.friction, 0.0, 2.0, "%.2f", 0) {
		s.Reset()
	}

	if imgui.SliderFloatV("Rolling Resistance", &s.rollingResistance, 0.0, 1.0, "%.2f", 0) {
		s.Reset()
	}

	imgui.PopItemWidth()

	return true
}

func (s *TangentSpeed) Step() {
	if s.stepCount%25 == 0 && len(s.bodyIDs) < tangentSpeedTotalCount && !s.context.pause {
		id := s.DropBall()
		s.bodyIDs = append(s.bodyIDs, id)
	}

	s.Sample.Step()
}

// This sample shows how to modify the geometry on an existing shape. This is only supported on
// dynamic and kinematic shapes because static shapes don't look for new collisions.
type ModifyGeometry struct {
	Sample

	shapeID   box2d.Shape
	shapeType box2d.ShapeType
	scale     float32

	circle  box2d.Circle
	capsule box2d.Capsule
	segment box2d.Segment
	polygon box2d.Polygon
}

func CreateModifyGeometry(context *SampleContext) Sampler {
	s := &ModifyGeometry{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 0.25
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(10.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 4.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(1.0, 1.0)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		s.shapeType = box2d.CircleShape
		s.scale = 1.0
		s.circle = box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.KinematicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		s.shapeID = bodyID.CreateCircleShape(&shapeDef, &s.circle)
	}
	return s
}

func (s *ModifyGeometry) UpdateShape() {
	switch s.shapeType {
	case box2d.CircleShape:
		s.circle = box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5 * s.scale}
		s.shapeID.SetCircle(&s.circle)

	case box2d.CapsuleShape:
		s.capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: -0.5 * s.scale, Y: 0.0},
			Center2: box2d.Vec2{X: 0.0, Y: 0.5 * s.scale},
			Radius:  0.5 * s.scale,
		}
		s.shapeID.SetCapsule(&s.capsule)

	case box2d.SegmentShape:
		s.segment = box2d.Segment{
			Point1: box2d.Vec2{X: -0.5 * s.scale, Y: 0.0},
			Point2: box2d.Vec2{X: 0.75 * s.scale, Y: 0.0},
		}
		s.shapeID.SetSegment(&s.segment)

	case box2d.PolygonShape:
		s.polygon = box2d.MakeBox(0.5*s.scale, 0.75*s.scale)
		s.shapeID.SetPolygon(&s.polygon)
	}

	bodyID := s.shapeID.Body()
	bodyID.UpdateMassFromShapes()
}

func (s *ModifyGeometry) DrawControls() bool {
	if imgui.RadioButtonBool("Circle", s.shapeType == box2d.CircleShape) {
		s.shapeType = box2d.CircleShape
		s.UpdateShape()
	}

	if imgui.RadioButtonBool("Capsule", s.shapeType == box2d.CapsuleShape) {
		s.shapeType = box2d.CapsuleShape
		s.UpdateShape()
	}

	if imgui.RadioButtonBool("Segment", s.shapeType == box2d.SegmentShape) {
		s.shapeType = box2d.SegmentShape
		s.UpdateShape()
	}

	if imgui.RadioButtonBool("Polygon", s.shapeType == box2d.PolygonShape) {
		s.shapeType = box2d.PolygonShape
		s.UpdateShape()
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())
	if imgui.SliderFloatV("Scale", &s.scale, 0.1, 10.0, "%.2f", 0) {
		s.UpdateShape()
	}
	imgui.PopItemWidth()

	bodyID := s.shapeID.Body()
	bodyType := bodyID.Type()

	if imgui.RadioButtonBool("Static", bodyType == box2d.StaticBody) {
		bodyID.SetType(box2d.StaticBody)
	}

	if imgui.RadioButtonBool("Kinematic", bodyType == box2d.KinematicBody) {
		bodyID.SetType(box2d.KinematicBody)
	}

	if imgui.RadioButtonBool("Dynamic", bodyType == box2d.DynamicBody) {
		bodyID.SetType(box2d.DynamicBody)
	}

	return true
}

// Shows how to link to chain shapes together. This is a useful technique for building large game levels with smooth collision.
type ChainLink struct {
	Sample
}

func CreateChainLink(context *SampleContext) Sampler {
	s := &ChainLink{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 5.0}
		s.context.camera.zoom = 25.0 * 0.5
	}

	points1 := []box2d.Vec2{
		{X: 0.0, Y: 0.0},
		{X: -40.0, Y: 0.0},
		{X: -40.0, Y: -1.0},
		{X: 0.0, Y: -1.0},
	}
	points2 := []box2d.Vec2{
		{X: 0.0, Y: -1.0},
		{X: 40.0, Y: -1.0},
		{X: 40.0, Y: 0.0},
		{X: 0.0, Y: 0.0},
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	{
		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points1

		// The other chain carries on past both ends of this one
		chainDef.Ghost1 = box2d.Vec2{X: 40.0, Y: 1.0}
		chainDef.Ghost2 = box2d.Vec2{X: 40.0, Y: -1.0}
		chainDef.IsLoop = false
		groundID.CreateChain(&chainDef)
	}

	{
		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points2
		chainDef.Ghost1 = box2d.Vec2{X: -40.0, Y: -1.0}
		chainDef.Ghost2 = box2d.Vec2{X: -40.0, Y: 0.0}
		chainDef.IsLoop = false
		groundID.CreateChain(&chainDef)
	}

	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()

	{
		bodyDef.Position = box2d.Pos{X: -5.0, Y: 2.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	{
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 2.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -0.5, Y: 0.0},
			Center2: box2d.Vec2{X: 0.5, Y: 0.0},
			Radius:  0.25,
		}
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		bodyDef.Position = box2d.Pos{X: 5.0, Y: 2.0}
		bodyID := s.worldID.CreateBody(&bodyDef)
		h := float32(0.5)
		box := box2d.MakeBox(h, h)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

func (s *ChainLink) Step() {
	s.Sample.Step()

	s.DrawScreenTextLine("This shows how to link together two chain shapes")
}

type RoundedShapes struct {
	Sample
}

func CreateRoundedShapes(context *SampleContext) Sampler {
	s := &RoundedShapes{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 0.55
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 8.0}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(20.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(1.0, 5.0, box2d.Vec2{X: 19.0, Y: 5.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(1.0, 5.0, box2d.Vec2{X: -19.0, Y: 5.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.3

	y := float32(2.0)
	xcount, ycount := 10, 10

	for range ycount {
		x := float32(-5.0)
		for range xcount {
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)

			poly := scenes.RandomPolygon(0.5)
			poly.Radius = scenes.RandomFloatRange(0.05, 0.25)
			bodyID.CreatePolygonShape(&shapeDef, &poly)

			x += 1.0
		}

		y += 1.0
	}
	return s
}

type EllipseShape struct {
	Sample
}

func CreateEllipseShape(context *SampleContext) Sampler {
	s := &EllipseShape{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 0.55
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 8.0}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(20.0, 1.0, box2d.Vec2{X: 0.0, Y: -1.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(1.0, 5.0, box2d.Vec2{X: 19.0, Y: 5.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(1.0, 5.0, box2d.Vec2{X: -19.0, Y: 5.0}, box2d.RotIdentity)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	points := []box2d.Vec2{
		{
			X: 0.0,
			Y: -0.25,
		},
		{X: 0.0, Y: 0.25},
		{X: 0.05, Y: 0.075},
		{X: -0.05, Y: 0.075},
		{X: 0.05, Y: -0.075},
		{X: -0.05, Y: -0.075},
	}
	diamondHull := box2d.ComputeHull(points)
	poly := box2d.MakePolygon(&diamondHull, 0.2)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.2

	y := float32(2.0)
	xCount, yCount := 10, 10

	for range yCount {
		x := float32(-5.0)
		for range xCount {
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)
			bodyID.CreatePolygonShape(&shapeDef, &poly)

			x += 1.0
		}

		y += 1.0
	}
	return s
}

type OffsetShapes struct {
	Sample
}

func CreateOffsetShapes(context *SampleContext) Sampler {
	s := &OffsetShapes{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = 25.0 * 0.55
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 8.0}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -1.0, Y: 1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			1.0,
			1.0,
			box2d.Vec2{X: 10.0, Y: -2.0},
			box2d.MakeRot(0.5*box2d.Pi),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -5.0, Y: 1.0},
			Center2: box2d.Vec2{X: -4.0, Y: 1.0},
			Radius:  0.25,
		}
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 13.5, Y: -0.75}
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		box := box2d.MakeOffsetBox(
			0.75,
			0.5,
			box2d.Vec2{X: 9.0, Y: 2.0},
			box2d.MakeRot(0.5*box2d.Pi),
		)
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0.0, Y: 0.0}
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}
	return s
}

func (s *OffsetShapes) Step() {
	s.Sample.Step()

	s.draw.DrawTransform(box2d.WorldTransformIdentity, 1.0)
}

// This shows how to use explosions and demonstrates the projected perimeter
type Explosion struct {
	Sample

	jointIDs       []box2d.Joint
	radius         float32
	falloff        float32
	impulse        float32
	referenceAngle float32
}

func CreateExplosion(context *SampleContext) Sampler {
	s := &Explosion{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 14.0
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.0
	shapeDef := box2d.DefaultShapeDef()

	s.referenceAngle = 0.0

	weldDef := box2d.DefaultWeldJointDef()
	weldDef.Base.BodyA = groundID
	weldDef.Base.LocalFrameA.Q = box2d.MakeRot(s.referenceAngle)
	weldDef.Base.LocalFrameB.P = box2d.Vec2Zero
	weldDef.AngularHertz = 0.5
	weldDef.AngularDampingRatio = 0.7
	weldDef.LinearHertz = 0.5
	weldDef.LinearDampingRatio = 0.7

	r := float32(8.0)
	for angle := float32(0.0); angle < 360.0; angle += 30.0 {
		cosSin := box2d.ComputeCosSin(angle * box2d.Pi / 180.0)
		bodyDef.Position = box2d.Pos{X: posScalar(r * cosSin.Cosine), Y: posScalar(r * cosSin.Sine)}
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(1.0, 0.1)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		weldDef.Base.LocalFrameA.P = bodyDef.Position.Vec2()
		weldDef.Base.BodyB = bodyID

		jointID := s.worldID.CreateWeldJoint(&weldDef)
		s.jointIDs = append(s.jointIDs, jointID.Joint)
	}

	s.radius = 7.0
	s.falloff = 3.0
	s.impulse = 10.0
	return s
}

func (s *Explosion) DrawControls() bool {
	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()
		def.Position = box2d.Pos{}
		def.Radius = s.radius
		def.Falloff = s.falloff
		def.ImpulsePerLength = s.impulse
		s.worldID.Explode(&def)
	}

	imgui.PushItemWidth(6.0 * imgui.FontSize())
	imgui.SliderFloatV("radius", &s.radius, 0.0, 20.0, "%.1f", 0)
	imgui.SliderFloatV("falloff", &s.falloff, 0.0, 20.0, "%.1f", 0)
	imgui.SliderFloatV("impulse", &s.impulse, -20.0, 20.0, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *Explosion) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		if s.context.hertz > 0.0 {
			s.referenceAngle += 60.0 * box2d.Pi / 180.0 / s.context.hertz
		}
		s.referenceAngle = box2d.UnwindAngle(s.referenceAngle)

		for _, jointID := range s.jointIDs {
			localFrameA := jointID.LocalFrameA()
			localFrameA.Q = box2d.MakeRot(s.referenceAngle)
			jointID.SetLocalFrameA(localFrameA)
		}
	}

	s.Sample.Step()

	s.DrawScreenTextLine("reference angle = %.6g", s.referenceAngle)

	s.draw.DrawCircle(box2d.Pos{}, s.radius+s.falloff, box2d.ColorBox2DBlue)
	s.draw.DrawCircle(box2d.Pos{}, s.radius, box2d.ColorBox2DYellow)
}

// This sample tests a static shape being recreated every step.
type RecreateStatic struct {
	Sample

	groundID box2d.Body
}

func CreateRecreateStatic(context *SampleContext) Sampler {
	s := &RecreateStatic{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 2.5}
		s.context.camera.zoom = 3.5
	}

	bodyDef := box2d.DefaultBodyDef()
	shapeDef := box2d.DefaultShapeDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
	bodyID := s.worldID.CreateBody(&bodyDef)

	box := box2d.MakeBox(1.0, 1.0)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	s.groundID = box2d.Body{}
	return s
}

func (s *RecreateStatic) Step() {
	if s.groundID != (box2d.Body{}) {
		s.groundID.Destroy()
		s.groundID = box2d.Body{}
	}

	bodyDef := box2d.DefaultBodyDef()
	s.groundID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	// Invoke contact creation so that contact points are created immediately
	// on a static body.
	shapeDef.InvokeContactCreation = true

	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -10.0, Y: 0.0},
		Point2: box2d.Vec2{X: 10.0, Y: 0.0},
	}
	s.groundID.CreateSegmentShape(&shapeDef, &segment)

	s.Sample.Step()
}

const windMaxCount = 60

type Wind struct {
	Sample

	shapeType int32
	wind      box2d.Vec2
	drag      float32
	lift      float32
	noise     box2d.Vec2
	groundID  box2d.Body
	bodyIDs   [windMaxCount]box2d.Body
	count     int32
}

func CreateWind(context *SampleContext) Sampler {
	s := &Wind{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.0}
		s.context.camera.zoom = 2.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		s.groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.shapeType = shapeChoiceCapsule
	s.wind = box2d.Vec2{X: 6.0, Y: 0.0}
	s.drag = 1.0
	s.lift = 0.75
	s.count = 10
	s.noise = box2d.Vec2{X: 0.0, Y: 0.0}

	s.CreateScene()
	return s
}

func (s *Wind) CreateScene() {
	for i := range windMaxCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}
		}
	}

	radius := float32(0.1)
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: radius}
	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: 0.0, Y: -radius},
		Center2: box2d.Vec2{X: 0.0, Y: radius},
		Radius:  0.25 * radius,
	}
	box := box2d.MakeBox(0.25*radius, 1.25*radius)

	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.BodyA = s.groundID
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 2.0 + radius}
	jointDef.Base.DrawScale = 0.1
	jointDef.Hertz = 0.1
	jointDef.DampingRatio = 0.0
	jointDef.EnableSpring = true

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 20.0

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.GravityScale = 0.5
	bodyDef.EnableSleep = false

	for i := range int(s.count) {
		bodyDef.Position = box2d.Pos{X: 0.0, Y: posScalar(2.0 - 2.0*radius*float32(i))}
		s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)

		if s.shapeType == shapeChoiceCircle {
			s.bodyIDs[i].CreateCircleShape(&shapeDef, &circle)
		} else if s.shapeType == shapeChoiceCapsule {
			s.bodyIDs[i].CreateCapsuleShape(&shapeDef, &capsule)
		} else {
			s.bodyIDs[i].CreatePolygonShape(&shapeDef, &box)
		}

		jointDef.Base.BodyB = s.bodyIDs[i]
		jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: radius}
		s.worldID.CreateRevoluteJoint(&jointDef)

		jointDef.Base.BodyA = s.bodyIDs[i]
		jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: -radius}
	}
}

func (s *Wind) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	if imgui.ComboStrarr("Shape", &s.shapeType, shapeChoiceNames, int32(len(shapeChoiceNames))) {
		s.CreateScene()
	}

	wind := [2]float32{s.wind.X, s.wind.Y}
	if imgui.SliderFloat2V("Wind", &wind, -20.0, 20.0, "%.1f", 0) {
		s.wind = box2d.Vec2{X: wind[0], Y: wind[1]}
	}
	imgui.SliderFloatV("Drag", &s.drag, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("Lift", &s.lift, 0.0, 4.0, "%.2f", 0)
	if imgui.SliderIntV("Count", &s.count, 1, windMaxCount, "%d", 0) {
		s.CreateScene()
	}

	imgui.PopItemWidth()

	return true
}

func (s *Wind) Step() {
	if !s.context.pause || s.context.singleStep > 0 {
		speed, direction := s.wind.LengthAndNormalize()
		wind := direction.Add(s.noise).Scale(speed)

		for i := range int(s.count) {
			for shapeID := range s.bodyIDs[i].Shapes() {
				shapeID.ApplyWind(wind, s.drag, s.lift, true)
				break
			}
		}

		rand := scenes.RandomVec2(-0.3, 0.3)
		s.noise = s.noise.Lerp(rand, 0.05)

		s.draw.DrawLine(box2d.Pos{}, wind.Scale(0.2).Pos(), box2d.ColorFuchsia)
	}

	s.Sample.Step()
}

func init() {
	RegisterSample("Shapes", "Chain Shape", CreateChainShape)
	RegisterSample("Shapes", "Chain Segment", CreateChainSegmentShape)
	RegisterSample("Shapes", "Compound Shapes", CreateCompoundShapes)
	RegisterSample("Shapes", "Filter", CreateShapeFilter)
	RegisterSample("Shapes", "Custom Filter", CreateCustomFilter)
	RegisterSample("Shapes", "Friction", CreateFriction)
	RegisterSample("Shapes", "Rolling Resistance", CreateRollingResistance)
	RegisterSample("Shapes", "Conveyor Belt", CreateConveyorBelt)
	RegisterSample("Shapes", "Tangent Speed", CreateTangentSpeed)
	RegisterSample("Shapes", "Modify Geometry", CreateModifyGeometry)
	RegisterSample("Shapes", "Chain Link", CreateChainLink)
	RegisterSample("Shapes", "Rounded", CreateRoundedShapes)
	RegisterSample("Shapes", "Ellipse", CreateEllipseShape)
	RegisterSample("Shapes", "Offset", CreateOffsetShapes)
	RegisterSample("Shapes", "Explosion", CreateExplosion)
	RegisterSample("Shapes", "Recreate Static", CreateRecreateStatic)
	RegisterSample("Shapes", "Wind", CreateWind)
}
