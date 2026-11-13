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

const (
	chainSegmentSegmentCount = 32
	chainSegmentPointCount   = chainSegmentSegmentCount + 3
)

const (
	shapeFilterGround  = 0x00000001
	shapeFilterTeam1   = 0x00000002
	shapeFilterTeam2   = 0x00000004
	shapeFilterTeam3   = 0x00000008
	shapeFilterAllBits = 0xFFFFFFFF
)

const customFilterCount = 10

const tangentSpeedTotalCount = 200

const windMaxCount = 60

var shapeChoiceNames = []string{"Circle", "Capsule", "Box"}

type ChainShape struct {
	Sample

	groundID  box2d.Body  //exhaustruct:optional
	bodyID    box2d.Body  //exhaustruct:optional
	chainID   box2d.Chain //exhaustruct:optional
	shapeType int32
	shapeID   box2d.Shape //exhaustruct:optional
	material  box2d.SurfaceMaterial
}

type ChainSegmentShape struct {
	Sample

	bodyID        box2d.Body //exhaustruct:optional
	shapeType     int32
	segmentShapes [chainSegmentSegmentCount]box2d.Shape //exhaustruct:optional
	points        [chainSegmentPointCount]box2d.Vec2    //exhaustruct:optional
	mutateIndex   int                                   //exhaustruct:optional
}

// This sample shows how careful creation of compound shapes leads to better simulation and avoids
// objects getting stuck.
// This also shows how to get the combined AABB for the body.
type CompoundShapes struct {
	Sample

	table1ID      box2d.Body //exhaustruct:optional
	table2ID      box2d.Body //exhaustruct:optional
	ship1ID       box2d.Body //exhaustruct:optional
	ship2ID       box2d.Body //exhaustruct:optional
	drawBodyAABBs bool       //exhaustruct:optional
}

type ShapeFilter struct {
	Sample

	player1ID box2d.Body //exhaustruct:optional
	player2ID box2d.Body //exhaustruct:optional
	player3ID box2d.Body //exhaustruct:optional

	shape1ID box2d.Shape //exhaustruct:optional
	shape2ID box2d.Shape //exhaustruct:optional
	shape3ID box2d.Shape //exhaustruct:optional
}

// This shows how to use custom filtering.
type CustomFilter struct {
	Sample

	bodyIDs  [customFilterCount]box2d.Body  //exhaustruct:optional
	shapeIDs [customFilterCount]box2d.Shape //exhaustruct:optional
}

type Friction struct {
	Sample
}

type RollingResistance struct {
	Sample

	resistScale float32
	lift        float32 //exhaustruct:optional
}

type ConveyorBelt struct {
	Sample
}

type TangentSpeed struct {
	Sample

	bodyIDs           []box2d.Body //exhaustruct:optional
	friction          float32
	rollingResistance float32
}

// This sample shows how to modify the geometry on an existing shape. This is only supported on
// dynamic and kinematic shapes because static shapes don't look for new collisions.
type ModifyGeometry struct {
	Sample

	shapeID   box2d.Shape //exhaustruct:optional
	shapeType box2d.ShapeType
	scale     float32

	circle  box2d.Circle
	capsule box2d.Capsule //exhaustruct:optional
	segment box2d.Segment //exhaustruct:optional
	polygon box2d.Polygon //exhaustruct:optional
}

// Shows how to link chain shapes together to build large game levels with smooth collision.
type ChainLink struct {
	Sample
}

type RoundedShapes struct {
	Sample
}

type EllipseShape struct {
	Sample
}

type OffsetShapes struct {
	Sample
}

// This shows how to use explosions and demonstrates the projected perimeter.
type Explosion struct {
	Sample

	jointIDs       []box2d.Joint //exhaustruct:optional
	radius         float32
	falloff        float32
	impulse        float32
	referenceAngle float32 //exhaustruct:optional
}

// This sample tests a static shape being recreated every step.
type RecreateStatic struct {
	Sample

	groundID box2d.Body //exhaustruct:optional
}

type Wind struct {
	Sample

	shapeType int32
	wind      box2d.Vec2
	drag      float32
	lift      float32
	noise     box2d.Vec2               //exhaustruct:optional
	groundID  box2d.Body               //exhaustruct:optional
	bodyIDs   [windMaxCount]box2d.Body //exhaustruct:optional
	count     int32
}

func CreateChainShape(context *SampleContext) Sampler {
	const (
		cameraBaseZoom  = 25.0
		cameraZoomScale = 1.75
	)

	s := &ChainShape{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		shapeType:       shapeChoiceCircle,
		material:        box2d.DefaultSurfaceMaterial(),
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

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
	//nolint:mnd // These are the authored path's sampled polygon coordinates.
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
	const (
		launchHeight      = 13.5
		circleRadius      = 0.5
		capsuleHalfLength = 0.5
		capsuleRadius     = 0.25
		boxHalfExtent     = 0.5
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -55.0, Y: launchHeight}
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material = s.material

	switch s.shapeType {
	case shapeChoiceCircle:
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
		s.shapeID = s.bodyID.CreateCircleShape(&shapeDef, &circle)
	case shapeChoiceCapsule:
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
			Radius:  capsuleRadius,
		}
		s.shapeID = s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	default:
		h := float32(boxHalfExtent)
		box := box2d.MakeBox(h, h)
		s.shapeID = s.bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	s.stepCount = 0
}

func (s *ChainShape) DrawControls() bool {
	const (
		controlWidthEm        = 6.0
		maxSurfaceCoefficient = 2.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.ComboStrarr("Shape", &s.shapeType, shapeChoiceNames, int32(len(shapeChoiceNames))) {
		s.Launch()
	}

	if imgui.SliderFloatV("Friction", &s.material.Friction, 0.0, 1.0, "%.2f", 0) {
		s.shapeID.SetSurfaceMaterial(&s.material)
		s.chainID.SetAllSurfaceMaterials(&s.material)
	}

	if imgui.SliderFloatV(
		"Restitution",
		&s.material.Restitution,
		0.0,
		maxSurfaceCoefficient,
		"%.1f",
		0,
	) {
		s.shapeID.SetSurfaceMaterial(&s.material)
	}

	imgui.PopItemWidth()

	if imgui.Button("Launch") {
		s.Launch()
	}

	return true
}

func (s *ChainShape) Step() {
	const (
		axisLength = 0.5
	)

	s.Sample.Step()

	var origin box2d.Pos

	s.draw.DrawLine(origin, box2d.Vec2{X: axisLength, Y: 0.0}.Pos(), box2d.ColorRed)
	s.draw.DrawLine(origin, box2d.Vec2{X: 0.0, Y: axisLength}.Pos(), box2d.ColorGreen)
}

func CreateChainSegmentShape(context *SampleContext) Sampler {
	const (
		cameraBaseZoom   = 25.0
		terrainHalfWidth = 25.0
		terrainWidth     = 50.0
		terrainAmplitude = 1.5
		terrainFrequency = 0.18
	)

	s := &ChainSegmentShape{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		shapeType:       shapeChoiceCircle,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = cameraBaseZoom * 1.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		// Walk right-to-left so the right-perpendicular normal of (point2 - point1) points up.
		for i := range chainSegmentPointCount {
			x := terrainHalfWidth - terrainWidth*float32(i)/(chainSegmentPointCount-1)
			y := terrainAmplitude * sinf(terrainFrequency*x)
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
	const (
		launchHeight      = 5.0
		circleRadius      = 0.25
		capsuleHalfLength = 0.5
		capsuleRadius     = 0.25
		boxHalfExtent     = 0.5
	)

	if s.bodyID != (box2d.Body{}) {
		s.bodyID.Destroy()
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: -18.0, Y: launchHeight}
	s.bodyID = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()

	switch s.shapeType {
	case shapeChoiceCircle:
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
		s.bodyID.CreateCircleShape(&shapeDef, &circle)
	case shapeChoiceCapsule:
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
			Radius:  capsuleRadius,
		}
		s.bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	default:
		box := box2d.MakeSquare(boxHalfExtent)
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
	const (
		controlWidthEm = 6.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

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

func CreateCompoundShapes(context *SampleContext) Sampler {
	const (
		cameraBaseZoom     = 25.0
		cameraCenterHeight = 6.0
		cameraZoomScale    = 0.5
		groundHalfWidth    = 50.0
		tableHalfWidth     = 3.0
		tableHalfThickness = 0.5
		tableTopHeight     = 3.5
		tableLegHalfWidth  = 0.5
		shortLegHalfHeight = 1.5
		tableLegOffset     = 2.5
		tallLegHalfHeight  = 2.0
		leftShipX          = 5.0
		shipVertexCount    = 3
		shipHeight         = 4.0
		shipNotchDivisor   = 3.0
		shipHalfWidth      = 2.0
		rightShipX         = 15.0
		shipNotchHeight    = 2.0
	)

	s := &CompoundShapes{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
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
		top := box2d.MakeOffsetBox(
			tableHalfWidth,
			tableHalfThickness,
			box2d.Vec2{X: 0.0, Y: tableTopHeight},
			box2d.RotIdentity,
		)
		leftLeg := box2d.MakeOffsetBox(
			tableLegHalfWidth,
			shortLegHalfHeight,
			box2d.Vec2{X: -tableLegOffset, Y: shortLegHalfHeight},
			box2d.RotIdentity,
		)
		rightLeg := box2d.MakeOffsetBox(
			tableLegHalfWidth,
			shortLegHalfHeight,
			box2d.Vec2{X: tableLegOffset, Y: shortLegHalfHeight},
			box2d.RotIdentity,
		)

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
		top := box2d.MakeOffsetBox(
			tableHalfWidth,
			tableHalfThickness,
			box2d.Vec2{X: 0.0, Y: tableTopHeight},
			box2d.RotIdentity,
		)
		leftLeg := box2d.MakeOffsetBox(
			tableLegHalfWidth,
			tallLegHalfHeight,
			box2d.Vec2{X: -tableLegOffset, Y: tallLegHalfHeight},
			box2d.RotIdentity,
		)
		rightLeg := box2d.MakeOffsetBox(
			tableLegHalfWidth,
			tallLegHalfHeight,
			box2d.Vec2{X: tableLegOffset, Y: tallLegHalfHeight},
			box2d.RotIdentity,
		)

		s.table2ID.CreatePolygonShape(&shapeDef, &top)
		s.table2ID.CreatePolygonShape(&shapeDef, &leftLeg)
		s.table2ID.CreatePolygonShape(&shapeDef, &rightLeg)
	}

	// Spaceship 1
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: leftShipX, Y: 1.0}
		s.ship1ID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		vertices := make([]box2d.Vec2, shipVertexCount)

		vertices[0] = box2d.Vec2{X: -shipHalfWidth, Y: 0.0}
		vertices[1] = box2d.Vec2{X: 0.0, Y: shipHeight / shipNotchDivisor}
		vertices[2] = box2d.Vec2{X: 0.0, Y: shipHeight}
		hull := box2d.ComputeHull(vertices)
		left := box2d.MakePolygon(&hull, 0.0)

		vertices[0] = box2d.Vec2{X: shipHalfWidth, Y: 0.0}
		vertices[1] = box2d.Vec2{X: 0.0, Y: shipHeight / shipNotchDivisor}
		vertices[2] = box2d.Vec2{X: 0.0, Y: shipHeight}
		hull = box2d.ComputeHull(vertices)
		right := box2d.MakePolygon(&hull, 0.0)

		s.ship1ID.CreatePolygonShape(&shapeDef, &left)
		s.ship1ID.CreatePolygonShape(&shapeDef, &right)
	}

	// Spaceship 2
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: rightShipX, Y: 1.0}
		s.ship2ID = s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		vertices := make([]box2d.Vec2, shipVertexCount)

		vertices[0] = box2d.Vec2{X: -shipHalfWidth, Y: 0.0}
		vertices[1] = box2d.Vec2{X: 1.0, Y: shipNotchHeight}
		vertices[2] = box2d.Vec2{X: 0.0, Y: shipHeight}
		hull := box2d.ComputeHull(vertices)
		left := box2d.MakePolygon(&hull, 0.0)

		vertices[0] = box2d.Vec2{X: shipHalfWidth, Y: 0.0}
		vertices[1] = box2d.Vec2{X: -1.0, Y: shipNotchHeight}
		vertices[2] = box2d.Vec2{X: 0.0, Y: shipHeight}
		hull = box2d.ComputeHull(vertices)
		right := box2d.MakePolygon(&hull, 0.0)

		s.ship2ID.CreatePolygonShape(&shapeDef, &left)
		s.ship2ID.CreatePolygonShape(&shapeDef, &right)
	}

	return s
}

func (s *CompoundShapes) Spawn() {
	const (
		tableObstructionHalfWidth  = 4.0
		tableObstructionHalfHeight = 0.1
		tableObstructionHeight     = 3.0
		shipObstructionHeight      = 2.0
		shipObstructionRadius      = 0.5
	)

	// Table 1 obstruction
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = s.table1ID.Position()
		bodyDef.Rotation = s.table1ID.Rotation()
		bodyID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			tableObstructionHalfWidth,
			tableObstructionHalfHeight,
			box2d.Vec2{X: 0.0, Y: tableObstructionHeight},
			box2d.RotIdentity,
		)
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
		box := box2d.MakeOffsetBox(
			tableObstructionHalfWidth,
			tableObstructionHalfHeight,
			box2d.Vec2{X: 0.0, Y: tableObstructionHeight},
			box2d.RotIdentity,
		)
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
		circle := box2d.Circle{
			Center: box2d.Vec2{X: 0.0, Y: shipObstructionHeight},
			Radius: shipObstructionRadius,
		}
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
		circle := box2d.Circle{
			Center: box2d.Vec2{X: 0.0, Y: shipObstructionHeight},
			Radius: shipObstructionRadius,
		}
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

func CreateShapeFilter(context *SampleContext) Sampler {
	const (
		cameraBaseZoom     = 25.0
		cameraZoomScale    = 0.5
		cameraCenterHeight = 5.0
		groundHalfWidth    = 20.0
		lowerPlayerHeight  = 2.0
		middlePlayerHeight = 5.0
		upperPlayerHeight  = 8.0
		playerHalfWidth    = 2.0
	)

	s := &ShapeFilter{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Filter.CategoryBits = shapeFilterGround
		shapeDef.Filter.MaskBits = shapeFilterAllBits

		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody

		bodyDef.Position = box2d.Pos{X: 0.0, Y: lowerPlayerHeight}
		s.player1ID = s.worldID.CreateBody(&bodyDef)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: middlePlayerHeight}
		s.player2ID = s.worldID.CreateBody(&bodyDef)

		bodyDef.Position = box2d.Pos{X: 0.0, Y: upperPlayerHeight}
		s.player3ID = s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(playerHalfWidth, 1.0)

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
	const (
		labelOffsetX = 0.5
	)

	s.Sample.Step()

	p1 := s.player1ID.Position()
	s.draw.DrawStringf(
		s.camera,
		box2d.Pos{X: p1.X - labelOffsetX, Y: p1.Y},
		box2d.ColorWhite,
		"player 1",
	)

	p2 := s.player2ID.Position()
	s.draw.DrawStringf(
		s.camera,
		box2d.Pos{X: p2.X - labelOffsetX, Y: p2.Y},
		box2d.ColorWhite,
		"player 2",
	)

	p3 := s.player3ID.Position()
	s.draw.DrawStringf(
		s.camera,
		box2d.Pos{X: p3.X - labelOffsetX, Y: p3.Y},
		box2d.ColorWhite,
		"player 3",
	)
}

func CreateCustomFilter(context *SampleContext) Sampler {
	const (
		cameraCenterHeight = 5.0
		groundHalfWidth    = 40.0
		bodyHeight         = 5.0
	)

	s := &CustomFilter{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
		s.context.camera.zoom = 10.0
	}

	// Register custom filter
	s.worldID.SetCustomFilterCallback(s.ShouldCollide)

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

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableCustomFiltering = true
	box := box2d.MakeSquare(1.0)
	x := posScalar(-customFilterCount)

	for i := range customFilterCount {
		bodyDef.Position = box2d.Pos{X: x, Y: bodyHeight}
		s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)

		shapeDef.UserData = i + 1
		s.shapeIDs[i] = s.bodyIDs[i].CreatePolygonShape(&shapeDef, &box)
		x += 2.0
	}

	return s
}

func (s *CustomFilter) Step() {
	s.DrawScreenTextLinef("Custom filter disables collision between odd and even shapes")

	s.Sample.Step()

	for i := range customFilterCount {
		p := s.bodyIDs[i].Position()
		s.draw.DrawStringf(s.camera, box2d.Pos{X: p.X, Y: p.Y}, box2d.ColorWhite, "%d", i)
	}
}

func (s *CustomFilter) ShouldCollide(shapeIDA box2d.Shape, shapeIDB box2d.Shape) bool {
	userDataA := shapeIDA.UserData()
	userDataB := shapeIDB.UserData()

	if userDataA == nil || userDataB == nil {
		return true
	}

	indexA, ok := userDataA.(int)
	if !ok {
		panic(
			"custom filter: shape A user data must be the integer index set by CreateCustomFilter",
		)
	}

	indexB, ok := userDataB.(int)
	if !ok {
		panic(
			"custom filter: shape B user data must be the integer index set by CreateCustomFilter",
		)
	}

	return ((indexA & 1) + (indexB & 1)) != 1
}

func CreateFriction(context *SampleContext) Sampler {
	const (
		cameraBaseZoom     = 25.0
		cameraCenterHeight = 14.0
		cameraZoomScale    = 0.6
		groundHalfWidth    = 40.0
		rampHalfLength     = 13.0
		rampHalfThickness  = 0.25
		rampAngle          = 0.25
		upperRampHeight    = 22.0
		stopHalfWidth      = 0.25
		stopOffsetX        = 10.5
		upperStopHeight    = 19.0
		rampOffsetX        = 4.0
		middleRampHeight   = 14.0
		lowerStopHeight    = 11.0
		lowerRampHeight    = 6.0
		boxHalfExtent      = 0.5
		boxStartHeight     = 28.0
	)

	s := &Friction{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.2

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		box := box2d.MakeOffsetBox(
			rampHalfLength,
			rampHalfThickness,
			box2d.Vec2{X: -rampOffsetX, Y: upperRampHeight},
			box2d.MakeRot(-rampAngle),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			stopHalfWidth,
			1.0,
			box2d.Vec2{X: stopOffsetX, Y: upperStopHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			rampHalfLength,
			rampHalfThickness,
			box2d.Vec2{X: rampOffsetX, Y: middleRampHeight},
			box2d.MakeRot(rampAngle),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			stopHalfWidth,
			1.0,
			box2d.Vec2{X: -stopOffsetX, Y: lowerStopHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			rampHalfLength,
			rampHalfThickness,
			box2d.Vec2{X: -rampOffsetX, Y: lowerRampHeight},
			box2d.MakeRot(-rampAngle),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		box := box2d.MakeBox(boxHalfExtent, boxHalfExtent)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Density = 25.0

		friction := [5]float32{0.75, 0.5, 0.35, 0.1, 0.0}

		for i := range 5 {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Pos{X: posScalar(-15.0 + 4.0*float32(i)), Y: boxStartHeight}
			bodyID := s.worldID.CreateBody(&bodyDef)

			shapeDef.Material.Friction = friction[i]
			bodyID.CreatePolygonShape(&shapeDef, &box)
		}
	}

	return s
}

func CreateRollingResistance(context *SampleContext) Sampler {
	const (
		cameraCenterX          = 5.0
		cameraCenterY          = 20.0
		initialResistanceScale = 0.02
	)

	s := &RollingResistance{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		resistScale:     initialResistanceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = 27.5
	}

	s.CreateScene()

	return s
}

func (s *RollingResistance) CreateScene() {
	const (
		circleRadius        = 0.5
		laneSpacing         = 2.0
		groundHalfWidth     = 40.0
		circleStartHeight   = 0.75
		initialAngularSpeed = 10.0
		initialLinearSpeed  = 5.0
	)

	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: circleRadius}

	shapeDef := box2d.DefaultShapeDef()

	for i := range 20 {
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: laneSpacing * float32(i)},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: laneSpacing*float32(i) + s.lift},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)

		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: -39.5, Y: posScalar(2.0*float32(i) + circleStartHeight)}
		bodyDef.AngularVelocity = -initialAngularSpeed
		bodyDef.LinearVelocity = box2d.Vec2{X: initialLinearSpeed, Y: 0.0}

		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef.Material.RollingResistance = s.resistScale * float32(i)
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
}

func (s *RollingResistance) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	const rampLift = 5.0

	//nolint:exhaustive // GLFW's external key set is open-ended; other keys pass to Sample.Keyboard.
	switch key {
	case glfw.Key1:
		s.lift = 0.0
		s.CreateWorld()
		s.CreateScene()

	case glfw.Key2:
		s.lift = rampLift
		s.CreateWorld()
		s.CreateScene()

	case glfw.Key3:
		s.lift = -rampLift
		s.CreateWorld()
		s.CreateScene()

	default:
		s.Sample.Keyboard(key, action, mods)
	}
}

func (s *RollingResistance) Step() {
	s.Sample.Step()

	for i := range 20 {
		s.draw.DrawStringf(
			s.camera,
			box2d.Pos{X: -41.5, Y: posScalar(2.0*float32(i) + 1.0)},
			box2d.ColorWhite,
			"%.2f",
			s.resistScale*float32(i),
		)
	}
}

func CreateConveyorBelt(context *SampleContext) Sampler {
	const (
		cameraCenterX      = 2.0
		cameraCenterY      = 7.5
		groundHalfWidth    = 20.0
		platformHeight     = 5.0
		platformOffsetX    = 5.0
		platformHalfHeight = 0.25
		platformHalfWidth  = 10.0
		platformRounding   = 0.25
		boxHalfExtent      = 0.5
		boxStartHeight     = 7.0
	)

	s := &ConveyorBelt{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
		s.context.camera.zoom = 12.0
	}

	// Ground
	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	// Platform
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -platformOffsetX, Y: platformHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeRoundedBox(platformHalfWidth, platformHalfHeight, platformRounding)

		shapeDef := box2d.DefaultShapeDef()
		shapeDef.Material.Friction = 0.8
		shapeDef.Material.TangentSpeed = 2.0

		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	// Boxes
	shapeDef := box2d.DefaultShapeDef()
	cube := box2d.MakeSquare(boxHalfExtent)

	for i := range 5 {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: posScalar(-10.0 + 2.0*float32(i)), Y: boxStartHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)

		bodyID.CreatePolygonShape(&shapeDef, &cube)
	}

	return s
}

func CreateTangentSpeed(context *SampleContext) Sampler {
	const (
		cameraCenterX            = 60.0
		pathOffsetY              = 0.25
		pathScale                = 0.2
		blueTangentSpeed         = 10.0
		cyanTangentSpeed         = 20.0
		goldenTangentSpeed       = 30.0
		grayTangentSpeed         = 40.0
		greenTangentSpeed        = 50.0
		khakiTangentSpeed        = 60.0
		magentaTangentSpeed      = 70.0
		initialFriction          = 0.6
		initialRollingResistance = 0.3
	)

	s := &TangentSpeed{
		context:           context,
		camera:            &context.camera,
		draw:              context.draw,
		screenTextX:       initialScreenTextX,
		mouseForceScale:   initialMouseForceScale,
		friction:          initialFriction,
		rollingResistance: initialRollingResistance,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: -15.0}
		s.context.camera.zoom = 38.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		path := "m 613.8334,185.20833 -42.33338,0 h -37.04166 l -34.39581,0 -29.10417,-2.64583 -26.45834,-7.9375 " +
			"-26.45833,-13.22917 -23.81251,-21.16666 h -13.22916 v 44.97916 H 68.791712 V 0 h -21.16671 v " +
			"206.375 l 566.208398,-1e-5 z"

		offset := box2d.Vec2{X: -47.375002, Y: pathOffsetY}

		scale := float32(pathScale)

		var points [20]box2d.Vec2

		count := ParsePath(path, offset, points[:], scale)

		var materials [20]box2d.SurfaceMaterial
		for i := range 20 {
			materials[i].Friction = 0.6
		}

		materials[0].TangentSpeed = -blueTangentSpeed
		materials[0].CustomColor = uint32(box2d.ColorDarkBlue)
		materials[1].TangentSpeed = -cyanTangentSpeed
		materials[1].CustomColor = uint32(box2d.ColorDarkCyan)
		materials[2].TangentSpeed = -goldenTangentSpeed
		materials[2].CustomColor = uint32(box2d.ColorDarkGoldenRod)
		materials[3].TangentSpeed = -grayTangentSpeed
		materials[3].CustomColor = uint32(box2d.ColorDarkGray)
		materials[4].TangentSpeed = -greenTangentSpeed
		materials[4].CustomColor = uint32(box2d.ColorDarkGreen)
		materials[5].TangentSpeed = -khakiTangentSpeed
		materials[5].CustomColor = uint32(box2d.ColorDarkKhaki)
		materials[6].TangentSpeed = -magentaTangentSpeed
		materials[6].CustomColor = uint32(box2d.ColorDarkMagenta)

		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points[:count]
		chainDef.IsLoop = true
		chainDef.Materials = materials[:count]

		groundID.CreateChain(&chainDef)
	}

	return s
}

func (s *TangentSpeed) DropBall() box2d.Body {
	const (
		ballRadius = 0.5
		ballStartX = 110.0
	)

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: ballRadius}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: ballStartX, Y: -30.0}
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
	const (
		controlWidthEm        = 6.0
		maxSurfaceCoefficient = 2.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.SliderFloatV("Friction", &s.friction, 0.0, maxSurfaceCoefficient, "%.2f", 0) {
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

func CreateModifyGeometry(context *SampleContext) Sampler {
	const (
		cameraBaseZoom      = 25.0
		cameraZoomScale     = 0.25
		cameraCenterHeight  = 5.0
		groundHalfWidth     = 10.0
		dynamicBodyHeight   = 4.0
		initialCircleRadius = 0.5
	)

	s := &ModifyGeometry{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		shapeType:       box2d.CircleShape,
		scale:           1.0,
		circle: box2d.Circle{
			Center: box2d.Vec2{X: 0.0, Y: 0.0},
			Radius: initialCircleRadius,
		},
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
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
		bodyDef.Position = box2d.Pos{X: 0.0, Y: dynamicBodyHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(1.0, 1.0)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	{
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
	const (
		circleBaseRadius      = 0.5
		capsuleBaseHalfLength = 0.5
		capsuleBaseRadius     = 0.5
		segmentRightExtent    = 0.75
		polygonHalfWidth      = 0.5
		polygonHalfHeight     = 0.75
		segmentLeftExtent     = 0.5
	)

	switch s.shapeType {
	case box2d.CircleShape:
		s.circle = box2d.Circle{
			Center: box2d.Vec2{X: 0.0, Y: 0.0},
			Radius: circleBaseRadius * s.scale,
		}
		s.shapeID.SetCircle(&s.circle)

	case box2d.CapsuleShape:
		s.capsule = box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleBaseHalfLength * s.scale, Y: 0.0},
			Center2: box2d.Vec2{X: 0.0, Y: capsuleBaseHalfLength * s.scale},
			Radius:  capsuleBaseRadius * s.scale,
		}
		s.shapeID.SetCapsule(&s.capsule)

	case box2d.SegmentShape:
		s.segment = box2d.Segment{
			Point1: box2d.Vec2{X: -segmentLeftExtent * s.scale, Y: 0.0},
			Point2: box2d.Vec2{X: segmentRightExtent * s.scale, Y: 0.0},
		}
		s.shapeID.SetSegment(&s.segment)

	case box2d.PolygonShape:
		s.polygon = box2d.MakeBox(polygonHalfWidth*s.scale, polygonHalfHeight*s.scale)
		s.shapeID.SetPolygon(&s.polygon)
	case box2d.ChainSegmentShape, box2d.ShapeTypeCount:
	}

	bodyID := s.shapeID.Body()
	bodyID.UpdateMassFromShapes()
}

func (s *ModifyGeometry) DrawControls() bool {
	const (
		controlWidthEm = 6.0
		minShapeScale  = 0.1
		maxShapeScale  = 10.0
	)

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

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.SliderFloatV("Scale", &s.scale, minShapeScale, maxShapeScale, "%.2f", 0) {
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

func CreateChainLink(context *SampleContext) Sampler {
	const (
		cameraBaseZoom     = 25.0
		cameraCenterHeight = 5.0
		cameraZoomScale    = 0.5
		chainHalfWidth     = 40.0
		bodyStartHeight    = 2.0
		circleRadius       = 0.5
		capsuleHalfLength  = 0.5
		capsuleRadius      = 0.25
		boxStartX          = 5.0
		boxHalfExtent      = 0.5
	)

	s := &ChainLink{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
	}

	points1 := []box2d.Vec2{
		{X: 0.0, Y: 0.0},
		{X: -chainHalfWidth, Y: 0.0},
		{X: -chainHalfWidth, Y: -1.0},
		{X: 0.0, Y: -1.0},
	}
	points2 := []box2d.Vec2{
		{X: 0.0, Y: -1.0},
		{X: chainHalfWidth, Y: -1.0},
		{X: chainHalfWidth, Y: 0.0},
		{X: 0.0, Y: 0.0},
	}

	bodyDef := box2d.DefaultBodyDef()
	groundID := s.worldID.CreateBody(&bodyDef)

	{
		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points1

		// The other chain carries on past both ends of this one
		chainDef.Ghost1 = box2d.Vec2{X: chainHalfWidth, Y: 1.0}
		chainDef.Ghost2 = box2d.Vec2{X: chainHalfWidth, Y: -1.0}
		chainDef.IsLoop = false
		groundID.CreateChain(&chainDef)
	}

	{
		chainDef := box2d.DefaultChainDef()
		chainDef.Points = points2
		chainDef.Ghost1 = box2d.Vec2{X: -chainHalfWidth, Y: -1.0}
		chainDef.Ghost2 = box2d.Vec2{X: -chainHalfWidth, Y: 0.0}
		chainDef.IsLoop = false
		groundID.CreateChain(&chainDef)
	}

	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()

	{
		bodyDef.Position = box2d.Pos{X: -5.0, Y: bodyStartHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: circleRadius}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}

	{
		bodyDef.Position = box2d.Pos{X: 0.0, Y: bodyStartHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -capsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: capsuleHalfLength, Y: 0.0},
			Radius:  capsuleRadius,
		}
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		bodyDef.Position = box2d.Pos{X: boxStartX, Y: bodyStartHeight}
		bodyID := s.worldID.CreateBody(&bodyDef)
		h := float32(boxHalfExtent)
		box := box2d.MakeBox(h, h)
		bodyID.CreatePolygonShape(&shapeDef, &box)
	}

	return s
}

func (s *ChainLink) Step() {
	s.Sample.Step()

	s.DrawScreenTextLinef("This shows how to link together two chain shapes")
}

func CreateRoundedShapes(context *SampleContext) Sampler {
	const (
		cameraBaseZoom    = 25.0
		cameraZoomScale   = 0.55
		cameraCenterX     = 2.0
		cameraCenterY     = 8.0
		groundHalfWidth   = 20.0
		wallHalfHeight    = 5.0
		wallOffsetX       = 19.0
		firstRowHeight    = 2.0
		polygonExtent     = 0.5
		minRoundingRadius = 0.05
		maxRoundingRadius = 0.25
	)

	s := &RoundedShapes{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
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

		box = box2d.MakeOffsetBox(
			1.0,
			wallHalfHeight,
			box2d.Vec2{X: wallOffsetX, Y: wallHalfHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			1.0,
			wallHalfHeight,
			box2d.Vec2{X: -wallOffsetX, Y: wallHalfHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.3

	y := float32(firstRowHeight)
	xcount, ycount := 10, 10

	for range ycount {
		x := float32(-5.0)
		for range xcount {
			bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
			bodyID := s.worldID.CreateBody(&bodyDef)

			poly := scenes.RandomPolygon(polygonExtent)
			poly.Radius = scenes.RandomFloatRange(minRoundingRadius, maxRoundingRadius)
			bodyID.CreatePolygonShape(&shapeDef, &poly)

			x += 1.0
		}

		y += 1.0
	}

	return s
}

func CreateEllipseShape(context *SampleContext) Sampler {
	const (
		cameraBaseZoom        = 25.0
		cameraZoomScale       = 0.55
		cameraCenterX         = 2.0
		cameraCenterY         = 8.0
		groundHalfWidth       = 20.0
		wallHalfHeight        = 5.0
		wallOffsetX           = 19.0
		ellipseHalfHeight     = 0.25
		ellipseHalfWidth      = 0.05
		ellipseShoulderHeight = 0.075
		ellipseRoundingRadius = 0.2
		firstRowHeight        = 2.0
	)

	s := &EllipseShape{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
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

		box = box2d.MakeOffsetBox(
			1.0,
			wallHalfHeight,
			box2d.Vec2{X: wallOffsetX, Y: wallHalfHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)

		box = box2d.MakeOffsetBox(
			1.0,
			wallHalfHeight,
			box2d.Vec2{X: -wallOffsetX, Y: wallHalfHeight},
			box2d.RotIdentity,
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	points := []box2d.Vec2{
		{
			X: 0.0,
			Y: -ellipseHalfHeight,
		},
		{X: 0.0, Y: ellipseHalfHeight},
		{X: ellipseHalfWidth, Y: ellipseShoulderHeight},
		{X: -ellipseHalfWidth, Y: ellipseShoulderHeight},
		{X: ellipseHalfWidth, Y: -ellipseShoulderHeight},
		{X: -ellipseHalfWidth, Y: -ellipseShoulderHeight},
	}
	diamondHull := box2d.ComputeHull(points)
	poly := box2d.MakePolygon(&diamondHull, ellipseRoundingRadius)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.2

	y := float32(firstRowHeight)
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

func CreateOffsetShapes(context *SampleContext) Sampler {
	const (
		cameraBaseZoom      = 25.0
		cameraZoomScale     = 0.55
		cameraCenterX       = 2.0
		cameraCenterY       = 8.0
		staticShapeOffsetX  = 10.0
		quarterTurnFactor   = 0.5
		capsuleRadius       = 0.25
		capsuleBodyX        = 13.5
		offsetBoxHalfWidth  = 0.75
		offsetBoxHalfHeight = 0.5
		dynamicShapeOffsetX = 9.0
		dynamicShapeOffsetY = 2.0
		capsuleBodyY        = 0.75
	)

	s := &OffsetShapes{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.zoom = cameraBaseZoom * cameraZoomScale
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterY}
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: -1.0, Y: 1.0}
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeOffsetBox(
			1.0,
			1.0,
			box2d.Vec2{X: staticShapeOffsetX, Y: -2.0},
			box2d.MakeRot(quarterTurnFactor*box2d.Pi),
		)
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	{
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -5.0, Y: 1.0},
			Center2: box2d.Vec2{X: -4.0, Y: 1.0},
			Radius:  capsuleRadius,
		}
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: capsuleBodyX, Y: -capsuleBodyY}
		bodyDef.Type = box2d.DynamicBody
		bodyID := s.worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		bodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}

	{
		box := box2d.MakeOffsetBox(
			offsetBoxHalfWidth,
			offsetBoxHalfHeight,
			box2d.Vec2{X: dynamicShapeOffsetX, Y: dynamicShapeOffsetY},
			box2d.MakeRot(quarterTurnFactor*box2d.Pi),
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

func CreateExplosion(context *SampleContext) Sampler {
	const (
		bodyRingRadius          = 8.0
		halfTurnDegrees         = 180.0
		bodyHalfHeight          = 0.1
		initialExplosionRadius  = 7.0
		initialExplosionFalloff = 3.0
		initialExplosionImpulse = 10.0
	)

	s := &Explosion{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		radius:          initialExplosionRadius,
		falloff:         initialExplosionFalloff,
		impulse:         initialExplosionImpulse,
	}
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

	weldDef := box2d.DefaultWeldJointDef()
	weldDef.Base.BodyA = groundID
	weldDef.Base.LocalFrameA.Q = box2d.MakeRot(s.referenceAngle)
	weldDef.Base.LocalFrameB.P = box2d.Vec2Zero
	weldDef.AngularHertz = 0.5
	weldDef.AngularDampingRatio = 0.7
	weldDef.LinearHertz = 0.5
	weldDef.LinearDampingRatio = 0.7

	r := float32(bodyRingRadius)

	for angle := float32(0.0); angle < 360.0; angle += 30.0 {
		cosSin := box2d.ComputeCosSin(angle * box2d.Pi / halfTurnDegrees)
		bodyDef.Position = box2d.Pos{X: posScalar(r * cosSin.Cosine), Y: posScalar(r * cosSin.Sine)}
		bodyID := s.worldID.CreateBody(&bodyDef)

		box := box2d.MakeBox(1.0, bodyHalfHeight)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		weldDef.Base.LocalFrameA.P = bodyDef.Position.Vec2()
		weldDef.Base.BodyB = bodyID

		jointID := s.worldID.CreateWeldJoint(&weldDef)
		s.jointIDs = append(s.jointIDs, jointID.Joint)
	}

	return s
}

func (s *Explosion) DrawControls() bool {
	const (
		controlWidthEm      = 6.0
		maxRadius           = 20.0
		maxFalloff          = 20.0
		maxImpulseMagnitude = 20.0
	)

	if imgui.Button("Explode") {
		def := box2d.DefaultExplosionDef()

		var origin box2d.Pos

		def.Position = origin
		def.Radius = s.radius
		def.Falloff = s.falloff
		def.ImpulsePerLength = s.impulse
		s.worldID.Explode(&def)
	}

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())
	imgui.SliderFloatV("radius", &s.radius, 0.0, maxRadius, "%.1f", 0)
	imgui.SliderFloatV("falloff", &s.falloff, 0.0, maxFalloff, "%.1f", 0)
	imgui.SliderFloatV("impulse", &s.impulse, -maxImpulseMagnitude, maxImpulseMagnitude, "%.1f", 0)
	imgui.PopItemWidth()

	return true
}

func (s *Explosion) Step() {
	const (
		halfTurnDegrees = 180.0
	)

	if !s.context.pause || s.context.singleStep > 0 {
		if s.context.hertz > 0.0 {
			s.referenceAngle += 60.0 * box2d.Pi / halfTurnDegrees / s.context.hertz
		}

		s.referenceAngle = box2d.UnwindAngle(s.referenceAngle)

		for _, jointID := range s.jointIDs {
			localFrameA := jointID.LocalFrameA()
			localFrameA.Q = box2d.MakeRot(s.referenceAngle)
			jointID.SetLocalFrameA(localFrameA)
		}
	}

	s.Sample.Step()

	s.DrawScreenTextLinef("reference angle = %.6g", s.referenceAngle)

	var origin box2d.Pos

	s.draw.DrawCircle(origin, s.radius+s.falloff, box2d.ColorBox2DBlue)
	s.draw.DrawCircle(origin, s.radius, box2d.ColorBox2DYellow)
}

func CreateRecreateStatic(context *SampleContext) Sampler {
	const (
		cameraCenterHeight = 2.5
	)

	s := &RecreateStatic{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterHeight}
		s.context.camera.zoom = 3.5
	}

	bodyDef := box2d.DefaultBodyDef()
	shapeDef := box2d.DefaultShapeDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 1.0}
	bodyID := s.worldID.CreateBody(&bodyDef)

	box := box2d.MakeBox(1.0, 1.0)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	return s
}

func (s *RecreateStatic) Step() {
	const (
		groundHalfWidth = 10.0
	)

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
		Point1: box2d.Vec2{X: -groundHalfWidth, Y: 0.0},
		Point2: box2d.Vec2{X: groundHalfWidth, Y: 0.0},
	}
	s.groundID.CreateSegmentShape(&shapeDef, &segment)

	s.Sample.Step()
}

func CreateWind(context *SampleContext) Sampler {
	const (
		initialWindSpeed = 6.0
		initialLift      = 0.75
		initialBodyCount = 10
	)

	s := &Wind{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		shapeType:       shapeChoiceCapsule,
		wind:            box2d.Vec2{X: initialWindSpeed, Y: 0.0},
		drag:            1.0,
		lift:            initialLift,
		count:           initialBodyCount,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.0}
		s.context.camera.zoom = 2.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		s.groundID = s.worldID.CreateBody(&bodyDef)
	}

	s.CreateScene()

	return s
}

func (s *Wind) CreateScene() {
	const (
		shapeRadius        = 0.1
		capsuleRadiusScale = 0.25
		boxHalfWidthScale  = 0.25
		boxHalfHeightScale = 1.25
		firstBodyHeight    = 2.0
		bodySpacingFactor  = 2.0
	)

	for i := range windMaxCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}
		}
	}

	radius := float32(shapeRadius)
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: radius}
	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: 0.0, Y: -radius},
		Center2: box2d.Vec2{X: 0.0, Y: radius},
		Radius:  capsuleRadiusScale * radius,
	}
	box := box2d.MakeBox(boxHalfWidthScale*radius, boxHalfHeightScale*radius)

	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.BodyA = s.groundID
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: firstBodyHeight + radius}
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
		bodyDef.Position = box2d.Pos{
			X: 0.0,
			Y: posScalar(bodySpacingFactor - bodySpacingFactor*radius*float32(i)),
		}
		s.bodyIDs[i] = s.worldID.CreateBody(&bodyDef)

		switch s.shapeType {
		case shapeChoiceCircle:
			s.bodyIDs[i].CreateCircleShape(&shapeDef, &circle)
		case shapeChoiceCapsule:
			s.bodyIDs[i].CreateCapsuleShape(&shapeDef, &capsule)
		default:
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
	const (
		controlWidthEm = 6.0
		maxWindSpeed   = 20.0
		maxLift        = 4.0
	)

	imgui.PushItemWidth(controlWidthEm * imgui.FontSize())

	if imgui.ComboStrarr("Shape", &s.shapeType, shapeChoiceNames, int32(len(shapeChoiceNames))) {
		s.CreateScene()
	}

	wind := [2]float32{s.wind.X, s.wind.Y}
	if imgui.SliderFloat2V("Wind", &wind, -maxWindSpeed, maxWindSpeed, "%.1f", 0) {
		s.wind = box2d.Vec2{X: wind[0], Y: wind[1]}
	}

	imgui.SliderFloatV("Drag", &s.drag, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("Lift", &s.lift, 0.0, maxLift, "%.2f", 0)

	if imgui.SliderIntV("Count", &s.count, 1, windMaxCount, "%d", 0) {
		s.CreateScene()
	}

	imgui.PopItemWidth()

	return true
}

func (s *Wind) Step() {
	const (
		noiseAmplitude = 0.3
		noiseBlend     = 0.05
		windArrowScale = 0.2
	)

	if !s.context.pause || s.context.singleStep > 0 {
		speed, direction := s.wind.LengthAndNormalize()
		wind := direction.Add(s.noise).Scale(speed)

		for i := range int(s.count) {
			for shapeID := range s.bodyIDs[i].Shapes() {
				shapeID.ApplyWind(wind, s.drag, s.lift, true)
				break
			}
		}

		rand := scenes.RandomVec2(-noiseAmplitude, noiseAmplitude)
		s.noise = s.noise.Lerp(rand, noiseBlend)

		var origin box2d.Pos

		s.draw.DrawLine(origin, wind.Scale(windArrowScale).Pos(), box2d.ColorFuchsia)
	}

	s.Sample.Step()
}

func registerShapeSamples() {
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
