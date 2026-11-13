// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const (
	shapePoint proxyShapeType = iota
	shapeSegment
	shapeTriangle
	shapeBox
)

const simplexCapacity = 20

const (
	castModeAny = iota
	castModeClosest
	castModeMultiple
	castModeSorted
)

const (
	castTypeRay = iota
	castTypeCircle
	castTypeCapsule
	castTypePolygon
)

const castWorldMaxCount = 64

const (
	overlapCircleShape = iota
	overlapCapsuleShape
	overlapBoxShape
)

const (
	overlapMaxCount  = 64
	overlapMaxDoomed = 16
)

const (
	smoothCircleShape = iota
	smoothBoxShape
)

const maxCastHits = 3

var proxyShapeTypes = []string{"point", "segment", "triangle", "box"}

type proxyShapeType int32

type proxyShapes struct {
	box      box2d.Polygon
	triangle box2d.Polygon
	point    box2d.Vec2
	segment  box2d.Segment
}

type ShapeDistance struct {
	Sample
	proxyShapes

	typeA   proxyShapeType
	typeB   proxyShapeType
	radiusA float32
	radiusB float32
	proxyA  box2d.ShapeProxy
	proxyB  box2d.ShapeProxy

	cache        box2d.SimplexCache             //exhaustruct:optional
	simplexes    [simplexCapacity]box2d.Simplex //exhaustruct:optional
	simplexCount int32
	simplexIndex int32

	transform box2d.Transform
	angle     float32

	basePosition box2d.Vec2 //exhaustruct:optional
	startPoint   box2d.Pos  //exhaustruct:optional
	baseAngle    float32    //exhaustruct:optional

	dragging    bool
	rotating    bool
	showIndices bool
	useCache    bool
	drawSimplex bool
}

type treeProxy struct {
	box        box2d.AABB
	fatBox     box2d.AABB
	position   box2d.Vec2
	width      box2d.Vec2
	proxyID    int
	rayStamp   int
	queryStamp int
	moved      bool
}

type DynamicTree struct {
	Sample

	tree         box2d.DynamicTree //exhaustruct:optional
	rowCount     int32
	columnCount  int32
	proxies      []treeProxy //exhaustruct:optional
	timeStamp    int
	fill         float32
	moveFraction float32
	moveDelta    float32
	ratio        float32
	grid         float32

	startPoint box2d.Pos //exhaustruct:optional
	endPoint   box2d.Pos //exhaustruct:optional

	rayDrag   bool
	queryDrag bool
	validate  bool
}

type RayCast struct {
	Sample

	box      box2d.Polygon
	triangle box2d.Polygon
	circle   box2d.Circle
	capsule  box2d.Capsule
	segment  box2d.Segment

	transform box2d.Transform
	angle     float32

	rayStart box2d.Pos
	rayEnd   box2d.Pos

	basePosition box2d.Vec2 //exhaustruct:optional
	baseAngle    float32    //exhaustruct:optional

	startPosition box2d.Pos //exhaustruct:optional

	rayDrag      bool
	translating  bool
	rotating     bool
	showFraction bool
}

type ShapeUserData struct {
	index  int
	ignore bool
}

type CastContext struct {
	points    [maxCastHits]box2d.Pos
	normals   [maxCastHits]box2d.Vec2
	fractions [maxCastHits]float32
	count     int
}

// This sample shows how to use the ray and shape cast functions on a b2World. This
// sample is configured to ignore initial overlap.
type CastWorld struct {
	Sample

	bodyIndex int
	bodyIDs   [castWorldMaxCount]box2d.Body    //exhaustruct:optional
	userData  [castWorldMaxCount]ShapeUserData //exhaustruct:optional
	polygons  [2]box2d.Polygon
	box       box2d.Polygon
	capsule   box2d.Capsule
	circle    box2d.Circle
	segment   box2d.Segment

	simple bool

	mode        int32
	ignoreIndex int

	castType   int32
	castRadius float32

	angleAnchor box2d.Pos //exhaustruct:optional
	baseAngle   float32   //exhaustruct:optional
	angle       float32
	rotating    bool

	rayStart box2d.Pos
	rayEnd   box2d.Pos
	dragging bool
}

type OverlapWorld struct {
	Sample

	bodyIndex   int
	bodyIDs     [overlapMaxCount]box2d.Body    //exhaustruct:optional
	userData    [overlapMaxCount]ShapeUserData //exhaustruct:optional
	polygons    [4]box2d.Polygon
	capsule     box2d.Capsule
	circle      box2d.Circle
	segment     box2d.Segment
	ignoreIndex int

	doomIDs   [overlapMaxDoomed]box2d.Shape //exhaustruct:optional
	doomCount int

	shapeType int32

	startPosition box2d.Pos //exhaustruct:optional

	position  box2d.Pos
	angle     float32
	baseAngle float32 //exhaustruct:optional

	dragging bool
	rotating bool
}

// Tests manifolds and contact points.
type Manifold struct {
	Sample

	smgroxCache1 box2d.SimplexCache //exhaustruct:optional
	smgroxCache2 box2d.SimplexCache //exhaustruct:optional
	smgcapCache1 box2d.SimplexCache //exhaustruct:optional
	smgcapCache2 box2d.SimplexCache //exhaustruct:optional

	wedge box2d.Hull

	transform box2d.Transform
	angle     float32
	round     float32

	basePosition box2d.Vec2 //exhaustruct:optional
	startPoint   box2d.Pos  //exhaustruct:optional
	baseAngle    float32    //exhaustruct:optional

	dragging       bool
	rotating       bool
	showCount      bool
	showIDs        bool
	showAnchors    bool
	showSeparation bool
	enableCaching  bool
}

type SmoothManifold struct {
	Sample

	shapeType int32

	segments []box2d.ChainSegment

	transform box2d.Transform
	angle     float32
	round     float32

	basePosition box2d.Vec2 //exhaustruct:optional
	startPoint   box2d.Pos  //exhaustruct:optional
	baseAngle    float32    //exhaustruct:optional

	dragging       bool
	rotating       bool
	showIDs        bool
	showAnchors    bool
	showSeparation bool
}

type ShapeCast struct {
	Sample
	proxyShapes

	typeA   proxyShapeType
	typeB   proxyShapeType
	radiusA float32
	radiusB float32
	proxyA  box2d.ShapeProxy
	proxyB  box2d.ShapeProxy

	transform   box2d.Transform
	angle       float32
	translation box2d.Vec2

	basePosition box2d.Vec2 //exhaustruct:optional
	startPoint   box2d.Pos  //exhaustruct:optional
	baseAngle    float32    //exhaustruct:optional

	dragging    bool
	sweeping    bool
	rotating    bool
	showIndices bool
	drawSimplex bool
	encroach    bool
}

type TimeOfImpact struct {
	Sample

	verticesA [4]box2d.Vec2
	verticesB [2]box2d.Vec2

	radiusA float32
	radiusB float32
}

func CreateShapeDistance(context *SampleContext) Sampler {
	const segmentHalfLength = 0.5

	const (
		cameraZoom        = 3.0
		triangleHalfWidth = 0.5
		boxHalfExtent     = 0.5
	)

	points := []box2d.Vec2{
		{X: -triangleHalfWidth, Y: 0.0},
		{X: triangleHalfWidth, Y: 0.0},
		{X: 0.0, Y: 1.0},
	}
	hull := box2d.ComputeHull(points)
	shapes := proxyShapes{
		point: box2d.Vec2Zero,
		segment: box2d.Segment{
			Point1: box2d.Vec2{X: -segmentHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: segmentHalfLength, Y: 0.0},
		},
		triangle: box2d.MakePolygon(&hull, 0.0),
		box:      box2d.MakeSquare(boxHalfExtent),
	}

	s := &ShapeDistance{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		proxyShapes:     shapes,
		proxyA:          shapes.makeProxy(shapeBox, 0.0),
		proxyB:          shapes.makeProxy(shapeBox, 0.0),
		typeA:           shapeBox,
		typeB:           shapeBox,
		transform:       box2d.TransformIdentity,
		radiusA:         0.0,
		radiusB:         0.0,
		simplexCount:    0,
		simplexIndex:    0,
		angle:           0.0,
		dragging:        false,
		rotating:        false,
		showIndices:     false,
		useCache:        false,
		drawSimplex:     false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = cameraZoom
	}

	return s
}

func (s *ShapeDistance) DrawControls() bool {
	const (
		controlWidthInFonts  = 6.0
		maxShapeRadius       = 0.5
		maxTranslationOffset = 2.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	shapeType := int32(s.typeA)
	if imgui.ComboStrarr("shape A", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeA = proxyShapeType(shapeType)
		s.proxyA = s.makeProxy(s.typeA, s.radiusA)
	}

	if imgui.SliderFloatV("radius A", &s.radiusA, 0.0, maxShapeRadius, "%.2f", 0) {
		s.proxyA.Radius = s.radiusA
	}

	shapeType = int32(s.typeB)
	if imgui.ComboStrarr("shape B", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeB = proxyShapeType(shapeType)
		s.proxyB = s.makeProxy(s.typeB, s.radiusB)
	}

	if imgui.SliderFloatV("radius B", &s.radiusB, 0.0, maxShapeRadius, "%.2f", 0) {
		s.proxyB.Radius = s.radiusB
	}

	imgui.Separator()

	imgui.SliderFloatV(
		"x offset",
		&s.transform.P.X,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)
	imgui.SliderFloatV(
		"y offset",
		&s.transform.P.Y,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)

	if imgui.SliderFloatV("angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.PopItemWidth()

	imgui.Separator()

	imgui.Checkbox("show indices", &s.showIndices)
	imgui.Checkbox("use cache", &s.useCache)

	imgui.Separator()

	if imgui.Checkbox("draw simplex", &s.drawSimplex) {
		s.simplexIndex = 0
	}

	if s.drawSimplex && s.simplexCount > 0 {
		imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())
		imgui.SliderInt("index", &s.simplexIndex, 0, s.simplexCount-1)
		s.simplexIndex = min(max(s.simplexIndex, 0), s.simplexCount-1)

		imgui.PopItemWidth()
	}

	return true
}

func (s *ShapeDistance) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == 0 && !s.rotating {
			s.dragging = true
			s.startPoint = p
			s.basePosition = s.transform.P
		} else if mods == glfw.ModShift && !s.dragging {
			s.rotating = true
			s.startPoint = p
			s.baseAngle = s.angle
		}
	}
}

func (s *ShapeDistance) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.dragging = false
		s.rotating = false
	}
}

func (s *ShapeDistance) MouseMove(p box2d.Pos) {
	const (
		translationSensitivity = 0.5
		rotationSensitivity    = 1.0
	)

	d := p.Sub(s.startPoint)

	if s.dragging {
		s.transform.P = s.basePosition.MulAdd(translationSensitivity, d)
	} else if s.rotating {
		s.angle, s.transform.Q = collisionMouseRotation(s.baseAngle, d.X, rotationSensitivity)
	}
}

func (s *ShapeDistance) Step() {
	const pointSimplexCount = 1

	const (
		segmentSimplexCount  = 2
		triangleSimplexCount = 3
		witnessPointSize     = 10.0
		normalLength         = 0.5
	)

	var input box2d.DistanceInput

	input.ProxyA = s.proxyA
	input.ProxyB = s.proxyB
	input.Transform = s.transform
	input.UseRadii = s.radiusA > 0.0 || s.radiusB > 0.0

	if !s.useCache {
		s.cache.Count = 0
	}

	output := box2d.ShapeDistance(&input, &s.cache, s.simplexes[:])

	s.simplexCount = int32(output.SimplexCount)

	s.drawShape(s.draw, s.typeA, box2d.WorldTransformIdentity, s.radiusA, box2d.ColorCyan)
	s.drawShape(
		s.draw,
		s.typeB,
		box2d.MakeWorldTransform(s.transform),
		s.radiusB,
		box2d.ColorBisque,
	)

	if s.drawSimplex {
		simplex := &s.simplexes[s.simplexIndex]
		vertices := [3]*box2d.SimplexVertex{&simplex.V1, &simplex.V2, &simplex.V3}

		if s.simplexIndex > 0 {
			pointA, pointB := computeSimplexWitnessPoints(simplex)

			s.draw.DrawLine(pointA.Pos(), pointB.Pos(), box2d.ColorWhite)
			s.draw.DrawPoint(pointA.Pos(), witnessPointSize, box2d.ColorWhite)
			s.draw.DrawPoint(pointB.Pos(), witnessPointSize, box2d.ColorWhite)
		}

		colors := [3]box2d.HexColor{box2d.ColorRed, box2d.ColorGreen, box2d.ColorBlue}

		for i := range simplex.Count {
			vertex := vertices[i]
			s.draw.DrawPoint(vertex.WA.Pos(), witnessPointSize, colors[i])
			s.draw.DrawPoint(vertex.WB.Pos(), witnessPointSize, colors[i])
		}
	} else {
		s.draw.DrawLine(output.PointA.Pos(), output.PointB.Pos(), box2d.ColorDimGray)
		s.draw.DrawPoint(output.PointA.Pos(), witnessPointSize, box2d.ColorWhite)
		s.draw.DrawPoint(output.PointB.Pos(), witnessPointSize, box2d.ColorWhite)

		s.draw.DrawLine(
			output.PointA.Pos(),
			output.PointA.MulAdd(normalLength, output.Normal).Pos(),
			box2d.ColorYellow,
		)
	}

	if s.showIndices {
		for i := range s.proxyA.Count {
			s.draw.DrawStringf(s.camera, s.proxyA.Points[i].Pos(), box2d.ColorWhite, " %d", i)
		}

		transformB := box2d.MakeWorldTransform(s.transform)
		for i := range s.proxyB.Count {
			p := transformB.Apply(s.proxyB.Points[i])
			s.draw.DrawStringf(s.camera, p, box2d.ColorWhite, " %d", i)
		}
	}

	s.DrawScreenTextLinef("mouse button 1: drag")
	s.DrawScreenTextLinef("mouse button 1 + shift: rotate")
	s.DrawScreenTextLinef("distance = %.2f, iterations = %d", output.Distance, output.Iterations)

	switch s.cache.Count {
	case pointSimplexCount:
		s.DrawScreenTextLinef("cache = {%d}, {%d}", s.cache.IndexA[0], s.cache.IndexB[0])
	case segmentSimplexCount:
		s.DrawScreenTextLinef(
			"cache = {%d, %d}, {%d, %d}",
			s.cache.IndexA[0],
			s.cache.IndexA[1],
			s.cache.IndexB[0],
			s.cache.IndexB[1],
		)
	case triangleSimplexCount:
		s.DrawScreenTextLinef(
			"cache = {%d, %d, %d}, {%d, %d, %d}",
			s.cache.IndexA[0],
			s.cache.IndexA[1],
			s.cache.IndexA[2],
			s.cache.IndexB[0],
			s.cache.IndexB[1],
			s.cache.IndexB[2],
		)
	}
}

func CreateDynamicTree(context *SampleContext) Sampler {
	const (
		cameraCenter        = 500.0
		baseCameraZoom      = 25.0
		treeZoomScale       = 21.0
		initialTreeRows     = 1000
		initialTreeColumns  = 1000
		initialFillFraction = 0.25
		initialMoveFraction = 0.05
		initialMoveDelta    = 0.1
		initialAspectRatio  = 5.0
	)

	rowCount := int32(initialTreeRows)
	columnCount := int32(initialTreeColumns)

	if isDebug {
		rowCount = 100
		columnCount = 100
	}

	s := &DynamicTree{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		rowCount:        rowCount,
		columnCount:     columnCount,
		fill:            initialFillFraction,
		moveFraction:    initialMoveFraction,
		moveDelta:       initialMoveDelta,
		ratio:           initialAspectRatio,
		grid:            1.0,
		validate:        true,
		timeStamp:       0,
		rayDrag:         false,
		queryDrag:       false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenter, Y: cameraCenter}
		s.context.camera.zoom = baseCameraZoom * treeZoomScale
	}

	s.BuildTree()

	return s
}

func (s *DynamicTree) BuildTree() {
	const (
		initialTreeCapacity = 16
		aabbMarginWidth     = 0.1
		minProxyWidth       = 0.1
		maxProxyWidth       = 0.5
	)

	proxyCapacity := int(s.rowCount * s.columnCount)
	s.proxies = make([]treeProxy, 0, proxyCapacity)

	y := float32(-4.0)

	s.tree = box2d.NewDynamicTree(initialTreeCapacity)

	aabbMargin := box2d.Vec2{X: aabbMarginWidth, Y: aabbMarginWidth}

	for range s.rowCount {
		x := float32(-40.0)

		for range s.columnCount {
			fillTest := scenes.RandomFloatRange(0.0, 1.0)
			if fillTest <= s.fill {
				var p treeProxy

				p.position = box2d.Vec2{X: x, Y: y}

				ratio := scenes.RandomFloatRange(1.0, s.ratio)

				width := scenes.RandomFloatRange(minProxyWidth, maxProxyWidth)
				if scenes.RandomFloat() > 0.0 {
					p.width.X = ratio * width
					p.width.Y = width
				} else {
					p.width.X = width
					p.width.Y = ratio * width
				}

				p.box.LowerBound = box2d.Vec2{X: x, Y: y}
				p.box.UpperBound = box2d.Vec2{X: x + p.width.X, Y: y + p.width.Y}
				p.fatBox.LowerBound = p.box.LowerBound.Sub(aabbMargin)
				p.fatBox.UpperBound = p.box.UpperBound.Add(aabbMargin)

				p.proxyID = s.tree.CreateProxy(
					p.fatBox,
					box2d.DefaultCategoryBits,
					uint64(len(s.proxies)),
				)
				p.rayStamp = -1
				p.queryStamp = -1
				p.moved = false
				s.proxies = append(s.proxies, p)
			}

			x += s.grid
		}

		y += s.grid
	}
}

func (s *DynamicTree) DrawControls() bool {
	const (
		controlWidthInFonts = 6.0
		maxGridCount        = 1000
		minGridSpacing      = 0.5
		maxGridSpacing      = 2.0
		maxAspectRatio      = 10.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	changed := imgui.SliderIntV("rows", &s.rowCount, 0, maxGridCount, "%d", 0)

	if imgui.SliderIntV("columns", &s.columnCount, 0, maxGridCount, "%d", 0) {
		changed = true
	}

	if imgui.SliderFloatV("fill", &s.fill, 0.0, 1.0, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("grid", &s.grid, minGridSpacing, maxGridSpacing, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("ratio", &s.ratio, 1.0, maxAspectRatio, "%.2f", 0) {
		changed = true
	}

	imgui.SliderFloatV("move", &s.moveFraction, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("delta", &s.moveDelta, 0.0, 1.0, "%.2f", 0)

	imgui.PopItemWidth()

	imgui.Separator()

	imTextf("mouse button 1: ray cast")
	imTextf("mouse button 1 + shift: query")

	if changed {
		s.BuildTree()
	}

	return true
}

func (s *DynamicTree) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == 0 && !s.queryDrag {
			s.rayDrag = true
			s.startPoint = p
			s.endPoint = p
		} else if mods == glfw.ModShift && !s.rayDrag {
			s.queryDrag = true
			s.startPoint = p
			s.endPoint = p
		}
	}
}

func (s *DynamicTree) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.queryDrag = false
		s.rayDrag = false
	}
}

func (s *DynamicTree) MouseMove(p box2d.Pos) {
	s.endPoint = p
}

func (s *DynamicTree) Step() {
	const binaryTreeBranchFactor = 2.0

	s.queryProxies()

	s.drawAndMoveProxies()

	{
		ticks := time.Now()

		for i := range s.proxies {
			p := &s.proxies[i]
			if p.moved {
				s.tree.MoveProxy(p.proxyID, p.fatBox)
			}
		}

		ms := millisecondsSince(ticks)
		s.DrawScreenTextLinef("incremental : %.3f ms", ms)
	}

	height := s.tree.Height()
	areaRatio := s.tree.AreaRatio()

	proxyCount := len(s.proxies)
	hmin := int(
		float32(
			math.Ceil(
				float64(
					float32(
						math.Log(float64(proxyCount)),
					)/float32(
						math.Log(binaryTreeBranchFactor),
					) - 1.0,
				),
			),
		),
	)
	s.DrawScreenTextLinef(
		"proxies = %d, height = %d, hmin = %d, area ratio = %.1f",
		proxyCount,
		height,
		hmin,
		areaRatio,
	)

	s.tree.Validate()

	s.timeStamp += 1
}

func CreateRayCast(context *SampleContext) Sampler {
	const (
		cameraCenterY        = 20.0
		rayCircleRadius      = 2.0
		rayCapsuleRadius     = 1.5
		rayBoxHalfExtent     = 2.0
		rayTriangleHalfWidth = 2.0
		rayTriangleHeight    = 3.0
		raySegmentHalfLength = 3.0
		rayStartHeight       = 30.0
	)

	vertices := []box2d.Vec2{
		{X: -rayTriangleHalfWidth, Y: 0.0},
		{X: rayTriangleHalfWidth, Y: 0.0},
		{X: rayTriangleHalfWidth, Y: rayTriangleHeight},
	}
	hull := box2d.ComputeHull(vertices)

	s := &RayCast{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		circle:          box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: rayCircleRadius},
		capsule: box2d.Capsule{
			Center1: box2d.Vec2{X: -1.0, Y: 1.0},
			Center2: box2d.Vec2{X: 1.0, Y: -1.0},
			Radius:  rayCapsuleRadius,
		},
		box:      box2d.MakeBox(rayBoxHalfExtent, rayBoxHalfExtent),
		triangle: box2d.MakePolygon(&hull, 0.0),
		segment: box2d.Segment{
			Point1: box2d.Vec2{X: -raySegmentHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: raySegmentHalfLength, Y: 0.0},
		},
		transform:    box2d.TransformIdentity,
		rayStart:     box2d.Pos{X: 0.0, Y: rayStartHeight},
		angle:        0.0,
		rayEnd:       box2d.Pos{X: 0.0, Y: 0.0},
		rayDrag:      false,
		translating:  false,
		rotating:     false,
		showFraction: false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: cameraCenterY}
		s.context.camera.zoom = 17.5
	}

	return s
}

func (s *RayCast) DrawControls() bool {
	const (
		controlWidthInFonts  = 6.0
		maxTranslationOffset = 2.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	imgui.SliderFloatV(
		"x offset",
		&s.transform.P.X,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)
	imgui.SliderFloatV(
		"y offset",
		&s.transform.P.Y,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)

	if imgui.SliderFloatV("angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.PopItemWidth()

	imgui.Checkbox("show fraction", &s.showFraction)

	if imgui.Button("Reset") {
		s.transform = box2d.TransformIdentity
		s.angle = 0.0
	}

	imgui.Separator()

	imTextf("mouse btn 1: ray cast")
	imTextf("mouse btn 1 + shft: translate")
	imTextf("mouse btn 1 + ctrl: rotate")

	return true
}

func (s *RayCast) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		s.startPosition = p

		//nolint:exhaustive // GLFW modifiers are bit flags; other combinations leave the drag mode unchanged.
		switch mods {
		case 0:
			s.rayStart = p
			s.rayDrag = true
		case glfw.ModShift:
			s.translating = true
			s.basePosition = s.transform.P
		case glfw.ModControl:
			s.rotating = true
			s.baseAngle = s.angle
		default:
			return
		}
	}
}

func (s *RayCast) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.rayDrag = false
		s.rotating = false
		s.translating = false
	}
}

func (s *RayCast) MouseMove(p box2d.Pos) {
	const (
		translationSensitivity = 0.5
		rayRotationSensitivity = 0.5
	)

	d := p.Sub(s.startPosition)

	switch {
	case s.rayDrag:
		s.rayEnd = p
	case s.translating:
		s.transform.P.X = s.basePosition.X + translationSensitivity*d.X
		s.transform.P.Y = s.basePosition.Y + translationSensitivity*d.Y
	case s.rotating:
		s.angle, s.transform.Q = collisionMouseRotation(s.baseAngle, d.X, rayRotationSensitivity)
	}
}

func (s *RayCast) DrawRay(output *box2d.CastOutput, hitPoint box2d.Pos) {
	const (
		pointSize    = 5.0
		labelOffsetX = 0.05
	)

	p1 := s.rayStart
	p2 := s.rayEnd
	d := p2.Sub(p1)

	if output.Hit {
		var p box2d.Pos

		if output.Fraction == 0.0 {
			p = hitPoint
			s.draw.DrawPoint(hitPoint, pointSize, box2d.ColorPeru)
		} else {
			p = p1.Add(d.Scale(output.Fraction))
			s.draw.DrawLine(p1, p, box2d.ColorWhite)
			s.draw.DrawPoint(p1, pointSize, box2d.ColorGreen)
			s.draw.DrawPoint(hitPoint, pointSize, box2d.ColorWhite)

			s.draw.DrawLine(p, p.Add(output.Normal), box2d.ColorViolet)
		}

		if s.showFraction {
			s.draw.DrawStringf(
				s.camera,
				p.Add(box2d.Vec2{X: labelOffsetX, Y: -0.02}),
				box2d.ColorWhite,
				"%.2f",
				output.Fraction,
			)
		}
	} else {
		s.draw.DrawLine(p1, p2, box2d.ColorWhite)
		s.draw.DrawPoint(p1, pointSize, box2d.ColorGreen)
		s.draw.DrawPoint(p2, pointSize, box2d.ColorRed)
	}
}

func (s *RayCast) Step() {
	const (
		rayLayoutOffset  = 20.0
		rayLayoutSpacing = 10.0
	)

	offset := box2d.Vec2{X: -rayLayoutOffset, Y: rayLayoutOffset}
	increment := box2d.Vec2{X: rayLayoutSpacing, Y: 0.0}

	color1 := box2d.ColorYellow

	var (
		output   box2d.CastOutput
		hitPoint box2d.Pos
	)

	maxFraction := float32(1.0)

	cast := func(transform box2d.WorldTransform, castFcn func(input *box2d.RayCastInput) box2d.CastOutput) {
		start := transform.InvApply(s.rayStart)
		translation := transform.Q.InvRotate(s.rayEnd.Sub(s.rayStart))
		input := box2d.RayCastInput{
			Origin:      start,
			Translation: translation,
			MaxFraction: maxFraction,
		}

		localOutput := castFcn(&input)
		if localOutput.Hit {
			output = localOutput
			hitPoint = transform.Apply(localOutput.Point)
			output.Normal = transform.Q.Rotate(localOutput.Normal)
			maxFraction = localOutput.Fraction
		}

		offset = offset.Add(increment)
	}

	{
		transform := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)
		s.draw.DrawSolidCircle(transform, s.circle.Center, s.circle.Radius, color1)

		cast(
			transform,
			func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCircle(&s.circle, input) },
		)
	}

	{
		transform := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)
		v1 := transform.Apply(s.capsule.Center1)
		v2 := transform.Apply(s.capsule.Center2)
		s.draw.DrawCapsule(v1, v2, s.capsule.Radius, color1)

		cast(
			transform,
			func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCapsule(&s.capsule, input) },
		)
	}

	{
		transform := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)
		s.draw.DrawSolidPolygon(transform, s.box.Vertices[:s.box.Count], 0.0, color1)

		cast(
			transform,
			func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastPolygon(&s.box, input) },
		)
	}

	{
		transform := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)
		s.draw.DrawSolidPolygon(transform, s.triangle.Vertices[:s.triangle.Count], 0.0, color1)

		cast(
			transform,
			func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastPolygon(&s.triangle, input) },
		)
	}

	{
		transform := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)

		p1 := transform.Apply(s.segment.Point1)
		p2 := transform.Apply(s.segment.Point2)
		s.draw.DrawLine(p1, p2, color1)

		cast(transform, func(input *box2d.RayCastInput) box2d.CastOutput {
			return box2d.RayCastSegment(&s.segment, input, false)
		})
	}

	s.DrawRay(&output, hitPoint)
}

func CreateCastWorld(context *SampleContext) Sampler {
	const octagonHalfWidthFraction = 0.5

	const octagonDiagonalSquare = 2.0

	const (
		baseCameraZoom          = 25.0
		cameraCenterX           = 2.0
		cameraCenterYCast       = 14.0
		castZoomScale           = 0.75
		groundHalfLength        = 40.0
		narrowTriangleHalfWidth = 0.1
		worldTriangleHeight     = 1.5
		worldBoxHalfExtent      = 0.5
		worldCapsuleHalfLength  = 0.5
		ignoredBodyIndex        = 7
		initialCastRadius       = 0.5
		worldCapsuleRadius      = 0.25
		worldCircleRadius       = 0.5
		rayHalfLength           = 20.0
		worldRayHeight          = 10.0
	)

	var polygons [2]box2d.Polygon

	{
		vertices := []box2d.Vec2{
			{X: -narrowTriangleHalfWidth, Y: 0.0},
			{X: narrowTriangleHalfWidth, Y: 0.0},
			{X: 0.0, Y: worldTriangleHeight},
		}
		hull := box2d.ComputeHull(vertices)
		polygons[0] = box2d.MakePolygon(&hull, 0.0)
		polygons[0].Radius = 0.5
	}

	{
		w := float32(1.0)
		b := w / (octagonDiagonalSquare + float32(math.Sqrt(octagonDiagonalSquare)))
		sq := float32(math.Sqrt(octagonDiagonalSquare)) * b

		vertices := []box2d.Vec2{
			{X: octagonHalfWidthFraction * sq, Y: 0.0},
			{X: octagonHalfWidthFraction * w, Y: b},
			{X: octagonHalfWidthFraction * w, Y: b + sq},
			{X: octagonHalfWidthFraction * sq, Y: w},
			{X: -octagonHalfWidthFraction * sq, Y: w},
			{X: -octagonHalfWidthFraction * w, Y: b + sq},
			{X: -octagonHalfWidthFraction * w, Y: b},
			{X: -octagonHalfWidthFraction * sq, Y: 0.0},
		}

		hull := box2d.ComputeHull(vertices)
		polygons[1] = box2d.MakePolygon(&hull, 0.0)
	}

	box := box2d.MakeBox(worldBoxHalfExtent, worldBoxHalfExtent)
	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: -worldCapsuleHalfLength, Y: 0.0},
		Center2: box2d.Vec2{X: worldCapsuleHalfLength, Y: 0.0},
		Radius:  worldCapsuleRadius,
	}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: worldCircleRadius}
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -1.0, Y: 0.0},
		Point2: box2d.Vec2{X: 1.0, Y: 0.0},
	}
	s := &CastWorld{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		polygons:        polygons,
		box:             box,
		capsule:         capsule,
		circle:          circle,
		segment:         segment,
		mode:            castModeClosest,
		ignoreIndex:     ignoredBodyIndex,
		castType:        castTypeRay,
		castRadius:      initialCastRadius,
		rayStart:        box2d.Pos{X: -rayHalfLength, Y: worldRayHeight},
		rayEnd:          box2d.Pos{X: rayHalfLength, Y: worldRayHeight},
		bodyIndex:       0,
		simple:          false,
		angle:           0.0,
		rotating:        false,
		dragging:        false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: cameraCenterYCast}
		s.context.camera.zoom = baseCameraZoom * castZoomScale
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -groundHalfLength, Y: 0.0},
			Point2: box2d.Vec2{X: groundHalfLength, Y: 0.0},
		}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	return s
}

func (s *CastWorld) Create(index int) {
	const (
		dynamicBodyRemainder = 2
		worldCircle          = 2
		worldCapsule         = 3
	)

	const (
		spawnExtent            = 20.0
		bodyTypeCount          = 3
		staticBodyRemainder    = 0
		kinematicBodyRemainder = 1
		worldPolygon           = 0
		worldBox               = 1
		worldSegment           = 4
	)

	if s.bodyIDs[s.bodyIndex] != (box2d.Body{}) {
		s.bodyIDs[s.bodyIndex].Destroy()
		s.bodyIDs[s.bodyIndex] = box2d.Body{}
	}

	x := scenes.RandomFloatRange(-spawnExtent, spawnExtent)
	y := scenes.RandomFloatRange(0.0, spawnExtent)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
	bodyDef.Rotation = box2d.MakeRot(scenes.RandomFloatRange(-box2d.Pi, box2d.Pi))

	mod := s.bodyIndex % bodyTypeCount
	switch mod {
	case staticBodyRemainder:
		bodyDef.Type = box2d.StaticBody
	case kinematicBodyRemainder:
		bodyDef.Type = box2d.KinematicBody
	case dynamicBodyRemainder:
		bodyDef.Type = box2d.DynamicBody
		bodyDef.GravityScale = 0.0
	}

	s.bodyIDs[s.bodyIndex] = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.UserData = &s.userData[s.bodyIndex]

	s.userData[s.bodyIndex].ignore = false
	if s.bodyIndex == s.ignoreIndex {
		s.userData[s.bodyIndex].ignore = true
	}

	switch index {
	case worldPolygon:
		polygonIndex := s.bodyIndex & 1
		s.bodyIDs[s.bodyIndex].CreatePolygonShape(&shapeDef, &s.polygons[polygonIndex])
	case worldBox:
		s.bodyIDs[s.bodyIndex].CreatePolygonShape(&shapeDef, &s.box)
	case worldCircle:
		s.bodyIDs[s.bodyIndex].CreateCircleShape(&shapeDef, &s.circle)
	case worldCapsule:
		s.bodyIDs[s.bodyIndex].CreateCapsuleShape(&shapeDef, &s.capsule)
	case worldSegment:
		s.bodyIDs[s.bodyIndex].CreateSegmentShape(&shapeDef, &s.segment)
	default:
		points := []box2d.Vec2{
			{X: 1.0, Y: 0.0},
			{X: -1.0, Y: 0.0},
			{X: -1.0, Y: -1.0},
			{X: 1.0, Y: -1.0},
		}
		chainDef := box2d.DefaultChainDef()
		chainDef.UserData = shapeDef.UserData
		chainDef.Points = points
		chainDef.IsLoop = true
		s.bodyIDs[s.bodyIndex].CreateChain(&chainDef)
	}

	s.bodyIndex = (s.bodyIndex + 1) % castWorldMaxCount
}

func (s *CastWorld) CreateN(index int, count int) {
	for range count {
		s.Create(index)
	}
}

func (s *CastWorld) DestroyBody() {
	for i := range castWorldMaxCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}

			return
		}
	}
}

func (s *CastWorld) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == 0 && !s.rotating {
			s.rayStart = p
			s.rayEnd = p
			s.dragging = true
		} else if mods == glfw.ModShift && !s.dragging {
			s.rotating = true
			s.angleAnchor = p
			s.baseAngle = s.angle
		}
	}
}

func (s *CastWorld) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.dragging = false
		s.rotating = false
	}
}

func (s *CastWorld) MouseMove(p box2d.Pos) {
	if s.dragging {
		s.rayEnd = p
	} else if s.rotating {
		dx := p.Sub(s.angleAnchor).X
		s.angle = s.baseAngle + 1.0*dx
	}
}

func (s *CastWorld) DrawControls() bool {
	const (
		controlWidthInFonts = 6.0
		maxCastRadius       = 2.0
	)

	imgui.Checkbox("Simple", &s.simple)

	if !s.simple {
		imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

		castTypes := []string{"Ray", "Circle", "Capsule", "Polygon"}
		imgui.ComboStrarr("Type", &s.castType, castTypes, int32(len(castTypes)))

		if s.castType != castTypeRay {
			imgui.SliderFloatV("Radius", &s.castRadius, 0.0, maxCastRadius, "%.1f", 0)
		}

		modes := []string{"Any", "Closest", "Multiple", "Sorted"}
		imgui.ComboStrarr("Mode", &s.mode, modes, int32(len(modes)))
		imgui.PopItemWidth()
	}

	controls := [...]struct {
		label      string
		batchLabel string
	}{
		{label: "Polygon", batchLabel: "10x##Poly"},
		{label: "Box", batchLabel: "10x##Box"},
		{label: "Circle", batchLabel: "10x##Circle"},
		{label: "Capsule", batchLabel: "10x##Capsule"},
		{label: "Segment", batchLabel: "10x##Segment"},
		{label: "Chain", batchLabel: "10x##Chain"},
	}
	for index, control := range controls {
		drawCollisionSpawnButtons(control.label, control.batchLabel, index, s.CreateN)
	}

	if imgui.Button("Destroy Shape") {
		s.DestroyBody()
	}

	return true
}

func (s *CastWorld) Step() {
	const (
		pointSize = 5.0
	)

	s.Sample.Step()

	s.DrawScreenTextLinef("Click left mouse button and drag to modify ray cast")
	s.DrawScreenTextLinef("Shape 7 is intentionally ignored by the ray")

	rayTranslation := s.rayEnd.Sub(s.rayStart)
	if s.simple {
		s.drawSimpleRayCast(rayTranslation)
	} else {
		s.drawWorldCast(rayTranslation)
	}

	s.draw.DrawPoint(s.rayStart, pointSize, box2d.ColorGreen)

	if s.bodyIDs[s.ignoreIndex] != (box2d.Body{}) {
		p := s.bodyIDs[s.ignoreIndex].Position()
		p.X -= 0.2
		s.draw.DrawStringf(s.camera, p, box2d.ColorWhite, "ign")
	}
}

func CreateOverlapWorld(context *SampleContext) Sampler {
	const octagonHalfWidthFraction = 0.5

	const octagonDiagonalSquare = 2.0

	const (
		baseCameraZoom          = 25.0
		narrowTriangleHalfWidth = 0.1
		worldTriangleHeight     = 1.5
		worldBoxHalfExtent      = 0.5
		worldCapsuleHalfLength  = 0.5
		worldCapsuleRadius      = 0.25
		worldCircleRadius       = 0.5
		overlapCenterHeight     = 10.0
		overlapZoomScale        = 0.7
		wideTriangleHalfWidth   = 0.5
		ignoredBodyIndex        = 7
		initialBodyCount        = 10
	)

	var polygons [4]box2d.Polygon

	{
		vertices := []box2d.Vec2{
			{X: -wideTriangleHalfWidth, Y: 0.0},
			{X: wideTriangleHalfWidth, Y: 0.0},
			{X: 0.0, Y: worldTriangleHeight},
		}
		hull := box2d.ComputeHull(vertices)
		polygons[0] = box2d.MakePolygon(&hull, 0.0)
	}

	{
		vertices := []box2d.Vec2{
			{X: -narrowTriangleHalfWidth, Y: 0.0},
			{X: narrowTriangleHalfWidth, Y: 0.0},
			{X: 0.0, Y: worldTriangleHeight},
		}
		hull := box2d.ComputeHull(vertices)
		polygons[1] = box2d.MakePolygon(&hull, 0.0)
	}

	{
		w := float32(1.0)
		b := w / (octagonDiagonalSquare + float32(math.Sqrt(octagonDiagonalSquare)))
		sq := float32(math.Sqrt(octagonDiagonalSquare)) * b

		vertices := []box2d.Vec2{
			{X: octagonHalfWidthFraction * sq, Y: 0.0},
			{X: octagonHalfWidthFraction * w, Y: b},
			{X: octagonHalfWidthFraction * w, Y: b + sq},
			{X: octagonHalfWidthFraction * sq, Y: w},
			{X: -octagonHalfWidthFraction * sq, Y: w},
			{X: -octagonHalfWidthFraction * w, Y: b + sq},
			{X: -octagonHalfWidthFraction * w, Y: b},
			{X: -octagonHalfWidthFraction * sq, Y: 0.0},
		}

		hull := box2d.ComputeHull(vertices)
		polygons[2] = box2d.MakePolygon(&hull, 0.0)
	}

	polygons[3] = box2d.MakeBox(worldBoxHalfExtent, worldBoxHalfExtent)
	capsule := box2d.Capsule{
		Center1: box2d.Vec2{X: -worldCapsuleHalfLength, Y: 0.0},
		Center2: box2d.Vec2{X: worldCapsuleHalfLength, Y: 0.0},
		Radius:  worldCapsuleRadius,
	}
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: worldCircleRadius}
	segment := box2d.Segment{
		Point1: box2d.Vec2{X: -1.0, Y: 0.0},
		Point2: box2d.Vec2{X: 1.0, Y: 0.0},
	}
	s := &OverlapWorld{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		polygons:        polygons,
		capsule:         capsule,
		circle:          circle,
		segment:         segment,
		ignoreIndex:     ignoredBodyIndex,
		shapeType:       overlapCircleShape,
		position:        box2d.Pos{X: 0.0, Y: overlapCenterHeight},
		bodyIndex:       0,
		doomCount:       0,
		angle:           0.0,
		dragging:        false,
		rotating:        false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: overlapCenterHeight}
		s.context.camera.zoom = baseCameraZoom * overlapZoomScale
	}

	s.CreateN(0, initialBodyCount)

	return s
}

func (s *OverlapWorld) Create(index int) {
	const (
		spawnExtent = 20.0
	)

	if s.bodyIDs[s.bodyIndex] != (box2d.Body{}) {
		s.bodyIDs[s.bodyIndex].Destroy()
		s.bodyIDs[s.bodyIndex] = box2d.Body{}
	}

	x := scenes.RandomFloatRange(-spawnExtent, spawnExtent)
	y := scenes.RandomFloatRange(0.0, spawnExtent)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
	bodyDef.Rotation = box2d.MakeRot(scenes.RandomFloatRange(-box2d.Pi, box2d.Pi))

	s.bodyIDs[s.bodyIndex] = s.worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.UserData = &s.userData[s.bodyIndex]
	s.userData[s.bodyIndex].index = s.bodyIndex

	s.userData[s.bodyIndex].ignore = false
	if s.bodyIndex == s.ignoreIndex {
		s.userData[s.bodyIndex].ignore = true
	}

	switch {
	case index < len(s.polygons):
		s.bodyIDs[s.bodyIndex].CreatePolygonShape(&shapeDef, &s.polygons[index])
	case index == len(s.polygons):
		s.bodyIDs[s.bodyIndex].CreateCircleShape(&shapeDef, &s.circle)
	case index == len(s.polygons)+1:
		s.bodyIDs[s.bodyIndex].CreateCapsuleShape(&shapeDef, &s.capsule)
	default:
		s.bodyIDs[s.bodyIndex].CreateSegmentShape(&shapeDef, &s.segment)
	}

	s.bodyIndex = (s.bodyIndex + 1) % overlapMaxCount
}

func (s *OverlapWorld) CreateN(index int, count int) {
	for range count {
		s.Create(index)
	}
}

func (s *OverlapWorld) DestroyBody() {
	for i := range overlapMaxCount {
		if s.bodyIDs[i] != (box2d.Body{}) {
			s.bodyIDs[i].Destroy()
			s.bodyIDs[i] = box2d.Body{}

			return
		}
	}
}

func (s *OverlapWorld) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == 0 && !s.rotating {
			s.dragging = true
			s.position = p
		} else if mods == glfw.ModShift && !s.dragging {
			s.rotating = true
			s.startPosition = p
			s.baseAngle = s.angle
		}
	}
}

func (s *OverlapWorld) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.dragging = false
		s.rotating = false
	}
}

func (s *OverlapWorld) MouseMove(p box2d.Pos) {
	if s.dragging {
		s.position = p
	} else if s.rotating {
		dx := p.Sub(s.startPosition).X
		s.angle = s.baseAngle + 1.0*dx
	}
}

func (s *OverlapWorld) DrawControls() bool {
	controls := [...]struct {
		label      string
		batchLabel string
	}{
		{label: "Polygon 1", batchLabel: "10x##Poly1"},
		{label: "Polygon 2", batchLabel: "10x##Poly2"},
		{label: "Polygon 3", batchLabel: "10x##Poly3"},
		{label: "Box", batchLabel: "10x##Box"},
		{label: "Circle", batchLabel: "10x##Circle"},
		{label: "Capsule", batchLabel: "10x##Capsule"},
		{label: "Segment", batchLabel: "10x##Segment"},
	}
	for index, control := range controls {
		drawCollisionSpawnButtons(control.label, control.batchLabel, index, s.CreateN)
	}

	if imgui.Button("Destroy Shape") {
		s.DestroyBody()
	}

	imgui.Separator()
	imTextf("Overlap Shape")
	imgui.RadioButtonIntPtr("Circle##Overlap", &s.shapeType, overlapCircleShape)
	imgui.RadioButtonIntPtr("Capsule##Overlap", &s.shapeType, overlapCapsuleShape)
	imgui.RadioButtonIntPtr("Box##Overlap", &s.shapeType, overlapBoxShape)

	return true
}

func (s *OverlapWorld) Step() {
	const (
		overlapCapsuleRadius = 0.5
		overlapBoxHalfWidth  = 2.0
		overlapBoxHalfHeight = 0.5
	)

	s.Sample.Step()

	s.DrawScreenTextLinef("left mouse button: drag query shape")
	s.DrawScreenTextLinef("left mouse button + shift: rotate query shape")

	s.doomCount = 0

	// Build the query shape near the origin and place it with the world-space query origin so it
	// stays precise far from the origin in large world mode.
	rotation := box2d.MakeRot(s.angle)

	var proxy box2d.ShapeProxy

	switch s.shapeType {
	case overlapCircleShape:
		circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: 1.0}
		proxy = box2d.MakeProxy([]box2d.Vec2{circle.Center}, circle.Radius)
		s.draw.DrawSolidCircle(
			box2d.WorldTransform{P: s.position, Q: box2d.RotIdentity},
			circle.Center,
			circle.Radius,
			box2d.ColorWhite,
		)
	case overlapCapsuleShape:
		capsule := box2d.Capsule{
			Center1: rotation.Rotate(box2d.Vec2{X: -1.0, Y: 0.0}),
			Center2: rotation.Rotate(box2d.Vec2{X: 1.0, Y: 0.0}),
			Radius:  overlapCapsuleRadius,
		}
		proxy = box2d.MakeProxy([]box2d.Vec2{capsule.Center1, capsule.Center2}, capsule.Radius)
		s.draw.DrawCapsule(
			s.position.Add(capsule.Center1),
			s.position.Add(capsule.Center2),
			capsule.Radius,
			box2d.ColorWhite,
		)
	case overlapBoxShape:
		box := box2d.MakeOffsetBox(
			overlapBoxHalfWidth,
			overlapBoxHalfHeight,
			box2d.Vec2Zero,
			rotation,
		)
		proxy = box2d.MakeProxy(box.Vertices[:box.Count], box.Radius)
		s.draw.DrawPolygon(
			box2d.WorldTransform{P: s.position, Q: box2d.RotIdentity},
			box.Vertices[:box.Count],
			box2d.ColorWhite,
		)
	}

	s.worldID.OverlapShape(s.position, &proxy, box2d.DefaultQueryFilter(), s.overlapResult)

	if s.bodyIDs[s.ignoreIndex] != (box2d.Body{}) {
		p := s.bodyIDs[s.ignoreIndex].Position()
		p.X -= 0.2
		s.draw.DrawStringf(s.camera, p, box2d.ColorWhite, "skip")
	}

	for i := range s.doomCount {
		shapeID := s.doomIDs[i]

		userData, _ := shapeID.UserData().(*ShapeUserData)
		if userData == nil {
			continue
		}

		index := userData.index

		s.bodyIDs[index].Destroy()
		s.bodyIDs[index] = box2d.Body{}
	}
}

func CreateManifold(context *SampleContext) Sampler {
	const (
		baseCameraZoom        = 25.0
		cameraCenterXManifold = 1.8
		manifoldZoomScale     = 0.45
		wedgeHalfWidth        = 0.1
		wedgeHalfHeight       = 0.5
		initialPositionX      = 0.17
		initialPositionY      = 1.12
		initialRoundRadius    = 0.1
	)

	points := []box2d.Vec2{
		{X: -wedgeHalfWidth, Y: -wedgeHalfHeight},
		{X: wedgeHalfWidth, Y: -wedgeHalfHeight},
		{X: 0.0, Y: wedgeHalfHeight},
	}
	wedge := box2d.ComputeHull(points)

	s := &Manifold{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		wedge:           wedge,
		transform: box2d.Transform{
			P: box2d.Vec2{X: initialPositionX, Y: initialPositionY},
			Q: box2d.RotIdentity,
		},
		round:          initialRoundRadius,
		enableCaching:  true,
		angle:          0.0,
		dragging:       false,
		rotating:       false,
		showCount:      false,
		showIDs:        false,
		showAnchors:    false,
		showSeparation: false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterXManifold, Y: 0.0}
		s.context.camera.zoom = baseCameraZoom * manifoldZoomScale
	}

	return s
}

func (s *Manifold) DrawControls() bool {
	const (
		controlWidthInFonts  = 6.0
		maxTranslationOffset = 2.0
		maxRoundingRadius    = 0.4
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	imgui.SliderFloatV(
		"x offset",
		&s.transform.P.X,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)
	imgui.SliderFloatV(
		"y offset",
		&s.transform.P.Y,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)

	if imgui.SliderFloatV("angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.SliderFloatV("round", &s.round, 0.0, maxRoundingRadius, "%.1f", 0)

	imgui.PopItemWidth()

	imgui.Separator()

	imgui.Checkbox("show count", &s.showCount)
	imgui.Checkbox("show ids", &s.showIDs)
	imgui.Checkbox("show separation", &s.showSeparation)
	imgui.Checkbox("show anchors", &s.showAnchors)
	imgui.Checkbox("enable caching", &s.enableCaching)

	if imgui.Button("Reset") {
		s.transform = box2d.TransformIdentity
		s.angle = 0.0
	}

	return true
}

func (s *Manifold) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	handled := false

	if button == glfw.MouseButton1 {
		if mods == 0 && !s.rotating {
			s.dragging = true
			s.startPoint = p
			s.basePosition = s.transform.P
			handled = true
		} else if mods == glfw.ModShift && !s.dragging {
			s.rotating = true
			s.startPoint = p
			s.baseAngle = s.angle
			handled = true
		}
	}

	if !handled {
		s.Sample.MouseDown(p, button, mods)
	}
}

func (s *Manifold) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.dragging = false
		s.rotating = false
	}
}

func (s *Manifold) MouseMove(p box2d.Pos) {
	const (
		translationSensitivity = 0.5
		rotationSensitivity    = 1.0
	)

	d := p.Sub(s.startPoint)

	switch {
	case s.dragging:
		s.transform.P.X = s.basePosition.X + translationSensitivity*d.X
		s.transform.P.Y = s.basePosition.Y + translationSensitivity*d.Y
	case s.rotating:
		s.angle, s.transform.Q = collisionMouseRotation(s.baseAngle, d.X, rotationSensitivity)
	default:
		s.Sample.MouseMove(p)
	}
}

func (s *Manifold) DrawManifold(
	manifold *box2d.LocalManifold,
	transformA box2d.WorldTransform, transformB box2d.WorldTransform,
) {
	const (
		pointSize        = 5.0
		normalLength     = 0.5
		contactPointSize = 10.0
		midpointFraction = 0.5
	)

	normal := transformA.Q.Rotate(manifold.Normal)

	if s.showCount {
		p := transformA.P.Lerp(transformB.P, midpointFraction)
		s.draw.DrawStringf(s.camera, p, box2d.ColorWhite, "%d", manifold.PointCount)
	}

	for i := range manifold.PointCount {
		mp := &manifold.Points[i]

		p1 := transformA.P.Add(transformA.Q.Rotate(mp.Point))
		p2 := p1.Add(normal.Scale(normalLength))
		s.draw.DrawLine(p1, p2, box2d.ColorViolet)

		if s.showAnchors {
			s.draw.DrawPoint(p1, pointSize, box2d.ColorRed)
		} else {
			s.draw.DrawPoint(p1, contactPointSize, box2d.ColorBlue)
		}

		drawContactLabels(&s.Sample, mp, p1, s.showIDs, s.showSeparation)
	}
}

func (s *Manifold) Step() {
	if !s.enableCaching {
		s.smgroxCache1 = box2d.SimplexCache{Count: 0, IndexA: [3]uint8{}, IndexB: [3]uint8{}}
		s.smgroxCache2 = box2d.SimplexCache{Count: 0, IndexA: [3]uint8{}, IndexB: [3]uint8{}}
		s.smgcapCache1 = box2d.SimplexCache{Count: 0, IndexA: [3]uint8{}, IndexB: [3]uint8{}}
		s.smgcapCache2 = box2d.SimplexCache{Count: 0, IndexA: [3]uint8{}, IndexB: [3]uint8{}}
	}

	s.drawRoundGeometry()
	s.drawPolygonGeometry()
	s.drawChainGeometry()
	s.DrawScreenTextLinef("mouse button 1: drag")
	s.DrawScreenTextLinef("mouse button 1 + shift: rotate")
}

func CreateSmoothManifold(context *SampleContext) Sampler {
	const (
		smoothCameraCenterX = 2.0
		smoothStartHeight   = 20.0
		westEntryX          = 20.58325
		westEntryY          = 14.54175
		westRiseX           = 21.90625
		westRiseY           = 15.8645
		westPeakX           = 24.552
		westPeakY           = 17.1875
		westNotchX          = 27.198
		westNotchY          = 11.89575
		farWestShoulderX    = 29.84375
		farWestShoulderY    = 15.8645
		farWestRiseX        = 29.84375
		farWestRiseY        = 21.15625
		northwestShelfX     = 25.875
		northwestShelfY     = 23.802
		northwestInsetX     = 20.58325
		northwestInsetY     = 25.125
		northwestPeakX      = 25.875
		northwestPeakY      = 29.09375
		northwestCrownX     = 20.58325
		northwestCrownY     = 31.7395
		northInsetX         = 11.0089998
		northInsetY         = 23.2290001
		northInnerShelfX    = 8.67700005
		northInnerShelfY    = 21.15625
		northInnerCornerX   = 6.03125
		northInnerCornerY   = 21.15625
		northInnerPeakX     = 7.35424995
		northInnerPeakY     = 29.09375
		northInnerCrownX    = 3.38549995
		northInnerCrownY    = 29.09375
		northBridgeX        = 1.90625
		northBridgeY        = 30.41675
		centralNotchX       = 5.875
		centralNotchY       = 17.1875
		northeastInsetX     = 11.16675
		northeastInsetY     = 25.125
		northeastPeakX      = 9.84375
		northeastPeakY      = 29.09375
		northeastCrownX     = 13.8125
		northeastCrownY     = 31.7395
		northeastBridgeX    = 21.75
		northeastBridgeY    = 30.41675
		eastTipX            = 28.3644981
		eastTipY            = 26.448
		eastShoulderX       = 25.71875
		eastShoulderY       = 18.5105
		eastInsetX          = 24.3957481
		eastInsetY          = 13.21875
		southeastShoulderX  = 17.78125
		southeastShoulderY  = 11.89575
		southeastInsetX     = 15.1355
		southeastInsetY     = 7.92700005
		southShelfX         = 5.875
		southShelfY         = 9.25
		southwestInsetX     = 1.90625
		southwestInsetY     = 11.89575
		innerSouthCornerX   = 3.25
		innerSouthCornerY   = 11.89575
		innerSouthNotchX    = 3.25
		innerSouthNotchY    = 9.9375
		innerSouthShelfX    = 4.70825005
		innerSouthShelfY    = 9.25
		innerWestShelfX     = 8.67700005
		innerWestShelfY     = 9.25
		innerWestRiseX      = 11.323
		innerWestRiseY      = 11.89575
		innerWestCornerX    = 13.96875
		innerWestCornerY    = 11.89575
		entryRiseX          = 15.29175
		entryRiseY          = 14.54175
		entryCornerX        = 19.2605
		entryCornerY        = 14.54175
	)

	// https://betravis.github.io/shape-tools/path-to-polygon/
	points := []box2d.Vec2{
		{X: -westEntryX, Y: westEntryY},
		{X: -westRiseX, Y: westRiseY},
		{X: -westPeakX, Y: westPeakY},
		{X: -westNotchX, Y: westNotchY},
		{X: -farWestShoulderX, Y: farWestShoulderY},
		{X: -farWestRiseX, Y: farWestRiseY},
		{X: -northwestShelfX, Y: northwestShelfY},
		{X: -northwestInsetX, Y: northwestInsetY},
		{X: -northwestPeakX, Y: northwestPeakY},
		{X: -northwestCrownX, Y: northwestCrownY},
		{X: -northInsetX, Y: northInsetY},
		{X: -northInnerShelfX, Y: northInnerShelfY},
		{X: -northInnerCornerX, Y: northInnerCornerY},
		{X: -northInnerPeakX, Y: northInnerPeakY},
		{X: -northInnerCrownX, Y: northInnerCrownY},
		{X: northBridgeX, Y: northBridgeY},
		{X: centralNotchX, Y: centralNotchY},
		{X: northeastInsetX, Y: northeastInsetY},
		{X: northeastPeakX, Y: northeastPeakY},
		{X: northeastCrownX, Y: northeastCrownY},
		{X: northeastBridgeX, Y: northeastBridgeY},
		{X: eastTipX, Y: eastTipY},
		{X: eastShoulderX, Y: eastShoulderY},
		{X: eastInsetX, Y: eastInsetY},
		{X: southeastShoulderX, Y: southeastShoulderY},
		{X: southeastInsetX, Y: southeastInsetY},
		{X: southShelfX, Y: southShelfY},
		{X: southwestInsetX, Y: southwestInsetY},
		{X: -innerSouthCornerX, Y: innerSouthCornerY},
		{X: -innerSouthNotchX, Y: innerSouthNotchY},
		{X: -innerSouthShelfX, Y: innerSouthShelfY},
		{X: -innerWestShelfX, Y: innerWestShelfY},
		{X: -innerWestRiseX, Y: innerWestRiseY},
		{X: -innerWestCornerX, Y: innerWestCornerY},
		{X: -entryRiseX, Y: entryRiseY},
		{X: -entryCornerX, Y: entryCornerY},
	}

	count := len(points)
	segments := make([]box2d.ChainSegment, count)

	for i := range count {
		i0 := count - 1
		if i > 0 {
			i0 = i - 1
		}

		i1 := i

		i2 := 0
		if i1 < count-1 {
			i2 = i1 + 1
		}

		i3 := 0
		if i2 < count-1 {
			i3 = i2 + 1
		}

		g1 := points[i0]
		p1 := points[i1]
		p2 := points[i2]
		g2 := points[i3]

		segments[i] = box2d.ChainSegment{
			Ghost1:  g1,
			Segment: box2d.Segment{Point1: p1, Point2: p2},
			Ghost2:  g2,
			ChainID: -1,
		}
	}

	s := &SmoothManifold{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		segments:        segments,
		shapeType:       smoothBoxShape,
		transform: box2d.Transform{
			P: box2d.Vec2{X: 0.0, Y: smoothStartHeight},
			Q: box2d.RotIdentity,
		},
		angle:          0.0,
		round:          0.0,
		dragging:       false,
		rotating:       false,
		showIDs:        false,
		showAnchors:    false,
		showSeparation: false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: smoothCameraCenterX, Y: smoothStartHeight}
		s.context.camera.zoom = 21.0
	}

	return s
}

func (s *SmoothManifold) DrawControls() bool {
	const (
		controlWidthInFonts  = 6.0
		maxTranslationOffset = 2.0
		maxRoundingRadius    = 0.4
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	{
		shapeTypes := []string{"Circle", "Box"}
		imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))
	}

	imgui.SliderFloatV(
		"x Offset",
		&s.transform.P.X,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)
	imgui.SliderFloatV(
		"y Offset",
		&s.transform.P.Y,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)

	if imgui.SliderFloatV("Angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.SliderFloatV("Round", &s.round, 0.0, maxRoundingRadius, "%.1f", 0)

	imgui.PopItemWidth()

	imgui.Checkbox("Show Ids", &s.showIDs)
	imgui.Checkbox("Show Separation", &s.showSeparation)
	imgui.Checkbox("Show Anchors", &s.showAnchors)

	if imgui.Button("Reset") {
		s.transform = box2d.TransformIdentity
		s.angle = 0.0
	}

	imgui.Separator()

	imTextf("mouse button 1: drag")
	imTextf("mouse button 1 + shift: rotate")

	return true
}

func (s *SmoothManifold) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == 0 && !s.rotating {
			s.dragging = true
			s.startPoint = p
			s.basePosition = s.transform.P
		} else if mods == glfw.ModShift && !s.dragging {
			s.rotating = true
			s.startPoint = p
			s.baseAngle = s.angle
		}
	}
}

func (s *SmoothManifold) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.dragging = false
		s.rotating = false
	}
}

func (s *SmoothManifold) MouseMove(p box2d.Pos) {
	const (
		rotationSensitivity = 1.0
	)

	d := p.Sub(s.startPoint)

	if s.dragging {
		s.transform.P.X = s.basePosition.X + d.X
		s.transform.P.Y = s.basePosition.Y + d.Y
	} else if s.rotating {
		s.angle, s.transform.Q = collisionMouseRotation(s.baseAngle, d.X, rotationSensitivity)
	}
}

func (s *SmoothManifold) DrawManifold(
	manifold *box2d.LocalManifold,
	transformA box2d.WorldTransform,
) {
	const (
		pointSize    = 5.0
		normalLength = 0.5
	)

	normal := transformA.Q.Rotate(manifold.Normal)

	for i := range manifold.PointCount {
		mp := &manifold.Points[i]

		p1 := transformA.P.Add(transformA.Q.Rotate(mp.Point))
		p2 := p1.Add(normal.Scale(normalLength))
		s.draw.DrawLine(p1, p2, box2d.ColorWhite)

		s.draw.DrawPoint(p1, pointSize, box2d.ColorGreen)

		drawContactLabels(&s.Sample, mp, p1, s.showIDs, s.showSeparation)
	}
}

func (s *SmoothManifold) Step() {
	const (
		manifoldCircleRadius  = 0.5
		manifoldBoxHalfExtent = 0.5
		chainVertexSize       = 4.0
	)

	color1 := box2d.ColorYellow
	color2 := box2d.ColorMagenta

	transform1 := box2d.WorldTransformIdentity
	transform2 := box2d.MakeWorldTransform(s.transform)

	for i := range s.segments {
		segment := &s.segments[i]
		p1 := transform1.Apply(segment.Segment.Point1)
		p2 := transform1.Apply(segment.Segment.Point2)
		s.draw.DrawLine(p1, p2, color1)
		s.draw.DrawPoint(p1, chainVertexSize, color1)
	}

	// chain-segment vs circle
	switch s.shapeType {
	case smoothCircleShape:
		radius := float32(manifoldCircleRadius)
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: radius}
		s.draw.DrawSolidCircle(transform2, circle.Center, circle.Radius, color2)

		for i := range s.segments {
			segment := &s.segments[i]
			m := box2d.CollideChainSegmentAndCircle(segment, &circle, transform1.InvMul(transform2))
			s.DrawManifold(&m, transform1)
		}
	case smoothBoxShape:
		h := manifoldBoxHalfExtent - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)
		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)

		for i := range s.segments {
			segment := &s.segments[i]

			var cache box2d.SimplexCache

			m := box2d.CollideChainSegmentAndPolygon(
				segment,
				&rox,
				transform1.InvMul(transform2),
				&cache,
			)
			s.DrawManifold(&m, transform1)
		}
	}
}

func CreateShapeCast(context *SampleContext) Sampler {
	const castSegmentLength = 0.5

	const (
		cameraZoom            = 3.0
		triangleHalfWidth     = 0.5
		shapeCastCameraHeight = 0.25
		castBoxHalfLength     = 8.984375
		castBoxHalfThickness  = 0.5
		sweepDistance         = 2.0
		initialPointRadius    = 0.2
	)

	points := []box2d.Vec2{
		{X: -triangleHalfWidth, Y: 0.0},
		{X: triangleHalfWidth, Y: 0.0},
		{X: 0.0, Y: 1.0},
	}
	hull := box2d.ComputeHull(points)
	shapes := proxyShapes{
		point: box2d.Vec2Zero,
		segment: box2d.Segment{
			Point1: box2d.Vec2{X: 0.0, Y: 0.0},
			Point2: box2d.Vec2{X: castSegmentLength, Y: 0.0},
		},
		triangle: box2d.MakePolygon(&hull, 0.0),
		box:      box2d.MakeBox(castBoxHalfLength, castBoxHalfThickness),
	}

	s := &ShapeCast{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		proxyShapes:     shapes,
		proxyA:          shapes.makeProxy(shapeBox, 0.0),
		proxyB:          shapes.makeProxy(shapePoint, initialPointRadius),
		typeA:           shapeBox,
		typeB:           shapePoint,
		radiusB:         initialPointRadius,
		transform:       box2d.Transform{P: box2d.Vec2{X: -0.6, Y: 0.0}, Q: box2d.RotIdentity},
		translation:     box2d.Vec2{X: sweepDistance, Y: 0.0},
		radiusA:         0.0,
		angle:           0.0,
		dragging:        false,
		sweeping:        false,
		rotating:        false,
		showIndices:     false,
		drawSimplex:     false,
		encroach:        false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: shapeCastCameraHeight}
		s.context.camera.zoom = cameraZoom
	}

	return s
}

func (s *ShapeCast) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		//nolint:exhaustive // GLFW modifiers are bit flags; other combinations leave the drag mode unchanged.
		switch mods {
		case 0:
			s.dragging = true
			s.sweeping = false
			s.rotating = false
			s.startPoint = p
			s.basePosition = s.transform.P
		case glfw.ModShift:
			s.dragging = false
			s.sweeping = false
			s.rotating = true
			s.startPoint = p
			s.baseAngle = s.angle
		case glfw.ModControl:
			s.dragging = false
			s.sweeping = true
			s.rotating = false
			s.startPoint = p
			s.basePosition = box2d.Vec2Zero
		default:
			return
		}
	}
}

func (s *ShapeCast) MouseUp(p box2d.Pos, button glfw.MouseButton) {
	if button == glfw.MouseButton1 {
		s.dragging = false
		s.sweeping = false
		s.rotating = false
	}
}

func (s *ShapeCast) MouseMove(p box2d.Pos) {
	const (
		translationSensitivity = 0.5
		rotationSensitivity    = 1.0
	)

	d := p.Sub(s.startPoint)

	switch {
	case s.dragging:
		s.transform.P = s.basePosition.MulAdd(translationSensitivity, d)
	case s.rotating:
		s.angle, s.transform.Q = collisionMouseRotation(s.baseAngle, d.X, rotationSensitivity)
	case s.sweeping:
		s.translation = d
	}
}

func (s *ShapeCast) DrawControls() bool {
	const (
		controlWidthInFonts  = 6.0
		maxShapeRadius       = 0.5
		maxTranslationOffset = 2.0
	)

	imgui.PushItemWidth(controlWidthInFonts * imgui.FontSize())

	shapeType := int32(s.typeA)
	if imgui.ComboStrarr("shape A", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeA = proxyShapeType(shapeType)
		s.proxyA = s.makeProxy(s.typeA, s.radiusA)
	}

	if imgui.SliderFloatV("radius A", &s.radiusA, 0.0, maxShapeRadius, "%.2f", 0) {
		s.proxyA.Radius = s.radiusA
	}

	shapeType = int32(s.typeB)
	if imgui.ComboStrarr("shape B", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeB = proxyShapeType(shapeType)
		s.proxyB = s.makeProxy(s.typeB, s.radiusB)
	}

	if imgui.SliderFloatV("radius B", &s.radiusB, 0.0, maxShapeRadius, "%.2f", 0) {
		s.proxyB.Radius = s.radiusB
	}

	imgui.Separator()

	imgui.SliderFloatV(
		"x offset",
		&s.transform.P.X,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)
	imgui.SliderFloatV(
		"y offset",
		&s.transform.P.Y,
		-maxTranslationOffset,
		maxTranslationOffset,
		"%.2f",
		0,
	)

	if imgui.SliderFloatV("angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.PopItemWidth()

	imgui.Separator()

	imgui.Checkbox("show indices", &s.showIndices)
	imgui.Checkbox("encroach", &s.encroach)

	return true
}

func (s *ShapeCast) Step() {
	const (
		pointSize    = 5.0
		normalLength = 0.5
	)

	s.Sample.Step()

	var input box2d.ShapeCastPairInput

	input.ProxyA = s.proxyA
	input.ProxyB = s.proxyB
	input.Transform = s.transform
	input.TranslationB = s.translation
	input.MaxFraction = 1.0
	input.CanEncroach = s.encroach

	output := box2d.ShapeCast(&input)

	t := float32(1.0)
	if output.Hit {
		t = output.Fraction
	}

	var transform box2d.Transform

	transform.Q = s.transform.Q
	transform.P = s.transform.P.MulAdd(t, input.TranslationB)

	var distanceInput box2d.DistanceInput

	distanceInput.ProxyA = s.proxyA
	distanceInput.ProxyB = s.proxyB
	distanceInput.Transform = transform
	distanceInput.UseRadii = false

	var distanceCache box2d.SimplexCache

	distanceOutput := box2d.ShapeDistance(&distanceInput, &distanceCache, nil)

	s.DrawScreenTextLinef("hit = %t, iterations = %d, fraction = %.6g, distance = %.6g", output.Hit,
		output.Iterations, output.Fraction, distanceOutput.Distance)

	s.drawShape(s.draw, s.typeA, box2d.WorldTransformIdentity, s.radiusA, box2d.ColorCyan)
	s.drawShape(
		s.draw,
		s.typeB,
		box2d.MakeWorldTransform(s.transform),
		s.radiusB,
		box2d.ColorLightGreen,
	)
	transform2 := box2d.Transform{P: s.transform.P.Add(s.translation), Q: s.transform.Q}
	s.drawShape(
		s.draw,
		s.typeB,
		box2d.MakeWorldTransform(transform2),
		s.radiusB,
		box2d.ColorIndianRed,
	)

	if output.Hit {
		s.drawShape(
			s.draw,
			s.typeB,
			box2d.MakeWorldTransform(transform),
			s.radiusB,
			box2d.ColorPlum,
		)

		if output.Fraction > 0.0 {
			s.draw.DrawPoint(output.Point.Pos(), pointSize, box2d.ColorWhite)
			s.draw.DrawLine(
				output.Point.Pos(),
				output.Point.MulAdd(normalLength, output.Normal).Pos(),
				box2d.ColorYellow,
			)
		} else {
			s.draw.DrawPoint(output.Point.Pos(), pointSize, box2d.ColorPeru)
		}
	}

	if s.showIndices {
		for i := range s.proxyA.Count {
			s.draw.DrawStringf(s.camera, s.proxyA.Points[i].Pos(), box2d.ColorWhite, " %d", i)
		}

		transformB := box2d.MakeWorldTransform(s.transform)
		for i := range s.proxyB.Count {
			p := transformB.Apply(s.proxyB.Points[i])
			s.draw.DrawStringf(s.camera, p, box2d.ColorWhite, " %d", i)
		}
	}

	s.DrawScreenTextLinef("mouse button 1: drag")
	s.DrawScreenTextLinef("mouse button 1 + shift: rotate")
	s.DrawScreenTextLinef("mouse button 1 + control: sweep")
	s.DrawScreenTextLinef(
		"distance = %.2f, iterations = %d",
		distanceOutput.Distance,
		output.Iterations,
	)
}

func CreateTimeOfImpact(context *SampleContext) Sampler {
	const (
		toiCapsuleHalfLength = 0.125000000
		toiCapsuleRadius     = 0.0299999993
	)

	const (
		toiBoxBottom         = 44.75
		toiBoxTop            = 45.25
		initialCameraCenterX = 0.6
		initialCameraCenterY = 2.0
		toiCameraCenterY     = 45
	)

	s := &TimeOfImpact{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		verticesA: [4]box2d.Vec2{
			{X: -16.25, Y: toiBoxBottom},
			{X: -15.75, Y: toiBoxBottom},
			{X: -15.75, Y: toiBoxTop},
			{X: -16.25, Y: toiBoxTop},
		},
		verticesB: [2]box2d.Vec2{
			{X: 0.0, Y: -toiCapsuleHalfLength},
			{X: 0.0, Y: toiCapsuleHalfLength},
		},
		radiusA: 0.0,
		radiusB: toiCapsuleRadius,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: initialCameraCenterX, Y: initialCameraCenterY}
		s.context.camera.center = box2d.Pos{X: -16, Y: toiCameraCenterY}
		s.context.camera.zoom = 5.0
	}

	return s
}

func (s *TimeOfImpact) Step() {
	const (
		sweepStartY    = 45.3520279
		sweepEndY      = 45.3413048
		sweepStartSine = 0.841092527
		sweepEndSine   = 0.889056742
	)

	s.Sample.Step()

	sweepA := box2d.Sweep{
		LocalCenter: box2d.Vec2Zero,
		C1:          box2d.Vec2{X: 0.0, Y: 0.0},
		C2:          box2d.Vec2{X: 0.0, Y: 0.0},
		Q1:          box2d.RotIdentity,
		Q2:          box2d.RotIdentity,
	}
	sweepB := box2d.Sweep{
		LocalCenter: box2d.Vec2Zero,
		C1:          box2d.Vec2{X: -15.8332710, Y: sweepStartY},
		C2:          box2d.Vec2{X: -15.8324337, Y: sweepEndY},
		Q1:          box2d.Rot{C: -0.540891349, S: sweepStartSine},
		Q2:          box2d.Rot{C: -0.457797021, S: sweepEndSine},
	}

	var input box2d.TOIInput

	input.ProxyA = box2d.MakeProxy(s.verticesA[:], s.radiusA)
	input.ProxyB = box2d.MakeProxy(s.verticesB[:], s.radiusB)
	input.SweepA = sweepA
	input.SweepB = sweepB
	input.MaxFraction = 1.0

	output := box2d.TimeOfImpact(&input)

	s.DrawScreenTextLinef("toi = %.6g", output.Fraction)

	// Draw A
	transformA := box2d.MakeWorldTransform(sweepA.Transform(0.0))
	s.draw.DrawPolygon(transformA, s.verticesA[:], box2d.ColorGray)

	// Draw B at t = 0
	transformB := box2d.MakeWorldTransform(sweepB.Transform(0.0))
	s.draw.DrawCapsule(
		transformB.Apply(s.verticesB[0]),
		transformB.Apply(s.verticesB[1]),
		s.radiusB,
		box2d.ColorGreen,
	)

	// Draw B at t = hit_time
	transformB = box2d.MakeWorldTransform(sweepB.Transform(output.Fraction))
	s.draw.DrawPolygon(transformB, s.verticesB[:], box2d.ColorOrange)

	// Draw B at t = 1
	transformB = box2d.MakeWorldTransform(sweepB.Transform(1.0))
	s.draw.DrawCapsule(
		transformB.Apply(s.verticesB[0]),
		transformB.Apply(s.verticesB[1]),
		s.radiusB,
		box2d.ColorRed,
	)

	if output.State == box2d.TOIStateHit {
		var distanceInput box2d.DistanceInput

		distanceInput.ProxyA = input.ProxyA
		distanceInput.ProxyB = input.ProxyB
		distanceInput.Transform = sweepA.Transform(output.Fraction).
			InvMul(sweepB.Transform(output.Fraction))
		distanceInput.UseRadii = false

		var cache box2d.SimplexCache

		distanceOutput := box2d.ShapeDistance(&distanceInput, &cache, nil)
		s.DrawScreenTextLinef("distance = %.6g", distanceOutput.Distance)
	}
}

func (p *proxyShapes) makeProxy(shapeType proxyShapeType, radius float32) box2d.ShapeProxy {
	const pointVertexCount = 1

	const (
		segmentPointCount = 2
		boxPointCount     = 4
	)

	var proxy box2d.ShapeProxy

	proxy.Radius = radius

	switch shapeType {
	case shapePoint:
		proxy.Points[0] = box2d.Vec2Zero
		proxy.Count = pointVertexCount

	case shapeSegment:
		proxy.Points[0] = p.segment.Point1
		proxy.Points[1] = p.segment.Point2
		proxy.Count = segmentPointCount

	case shapeTriangle:
		for i := range p.triangle.Count {
			proxy.Points[i] = p.triangle.Vertices[i]
		}

		proxy.Count = p.triangle.Count

	case shapeBox:
		proxy.Points[0] = p.box.Vertices[0]
		proxy.Points[1] = p.box.Vertices[1]
		proxy.Points[2] = p.box.Vertices[2]
		proxy.Points[3] = p.box.Vertices[3]
		proxy.Count = boxPointCount

	default:
		panic("unknown shape type")
	}

	return proxy
}

func (p *proxyShapes) drawShape(
	draw *Draw,
	shapeType proxyShapeType,
	transform box2d.WorldTransform,
	radius float32,
	color box2d.HexColor,
) {
	const (
		pointSize = 5.0
	)

	switch shapeType {
	case shapePoint:
		point := transform.Apply(p.point)
		if radius > 0.0 {
			draw.DrawSolidCircle(
				box2d.WorldTransform{P: point, Q: transform.Q},
				box2d.Vec2Zero,
				radius,
				color,
			)
		} else {
			draw.DrawPoint(point, pointSize, color)
		}

	case shapeSegment:
		p1 := transform.Apply(p.segment.Point1)
		p2 := transform.Apply(p.segment.Point2)

		if radius > 0.0 {
			draw.DrawCapsule(p1, p2, radius, color)
		} else {
			draw.DrawLine(p1, p2, color)
		}

	case shapeTriangle:
		draw.DrawSolidPolygon(transform, p.triangle.Vertices[:p.triangle.Count], radius, color)

	case shapeBox:
		draw.DrawSolidPolygon(transform, p.box.Vertices[:p.box.Count], radius, color)

	default:
		panic("unknown shape type")
	}
}

func weight2(a1 float32, w1 box2d.Vec2, a2 float32, w2 box2d.Vec2) box2d.Vec2 {
	return box2d.Vec2{X: a1*w1.X + a2*w2.X, Y: a1*w1.Y + a2*w2.Y}
}

func weight3(
	a1 float32,
	w1 box2d.Vec2,
	a2 float32,
	w2 box2d.Vec2,
	a3 float32,
	w3 box2d.Vec2,
) box2d.Vec2 {
	return box2d.Vec2{X: a1*w1.X + a2*w2.X + a3*w3.X, Y: a1*w1.Y + a2*w2.Y + a3*w3.Y}
}

func computeSimplexWitnessPoints(simplex *box2d.Simplex) (box2d.Vec2, box2d.Vec2) {
	const pointSimplexCount = 1

	const (
		segmentSimplexCount  = 2
		triangleSimplexCount = 3
	)

	switch simplex.Count {
	case pointSimplexCount:
		return simplex.V1.WA, simplex.V1.WB

	case segmentSimplexCount:
		a := weight2(simplex.V1.A, simplex.V1.WA, simplex.V2.A, simplex.V2.WA)
		b := weight2(simplex.V1.A, simplex.V1.WB, simplex.V2.A, simplex.V2.WB)

		return a, b

	case triangleSimplexCount:
		a := weight3(
			simplex.V1.A,
			simplex.V1.WA,
			simplex.V2.A,
			simplex.V2.WA,
			simplex.V3.A,
			simplex.V3.WA,
		)

		return a, a

	default:
		panic("invalid simplex count")
	}
}

func shapeIgnored(shape box2d.Shape) bool {
	userData, _ := shape.UserData().(*ShapeUserData)
	return userData != nil && userData.ignore
}

func rayCastClosestCallback(rayContext *CastContext) box2d.CastResultFunc {
	return func(shape box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
		// Ignore a specific shape. Also ignore initial overlap.
		if shapeIgnored(shape) || fraction == 0.0 {
			// By returning -1, we instruct the calling code to ignore this shape and
			// continue the ray-cast to the next shape.
			return -1.0
		}

		rayContext.points[0] = point
		rayContext.normals[0] = normal
		rayContext.fractions[0] = fraction
		rayContext.count = 1

		// By returning the current fraction, we instruct the calling code to clip the ray and
		// continue the ray-cast to the next shape. WARNING: do not assume that shapes
		// are reported in order. However, by clipping, we can always get the closest shape.
		return fraction
	}
}

// This callback finds any hit. For this type of query we are usually just checking for obstruction,
// so the hit data is not relevant.
// NOTE: shape hits are not ordered, so this may not return the closest hit.
func rayCastAnyCallback(rayContext *CastContext) box2d.CastResultFunc {
	return func(shape box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
		if shapeIgnored(shape) || fraction == 0.0 {
			return -1.0
		}

		rayContext.points[0] = point
		rayContext.normals[0] = normal
		rayContext.fractions[0] = fraction
		rayContext.count = 1

		// At this point we have a hit, so we know the ray is obstructed.
		// By returning 0, we instruct the calling code to terminate the ray-cast.
		return 0.0
	}
}

// This ray cast collects multiple hits along the ray.
// The shapes are not necessary reported in order, so we might not capture
// the closest shape.
// NOTE: shape hits are not ordered, so this may return hits in any order. This means that
// if you limit the number of results, you may discard the closest hit. You can see this
// behavior in the sample.
func rayCastMultipleCallback(rayContext *CastContext) box2d.CastResultFunc {
	return func(shape box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
		if shapeIgnored(shape) || fraction == 0.0 {
			return -1.0
		}

		count := rayContext.count

		rayContext.points[count] = point
		rayContext.normals[count] = normal
		rayContext.fractions[count] = fraction
		rayContext.count = count + 1

		if rayContext.count == maxCastHits {
			// At this point the buffer is full.
			// By returning 0, we instruct the calling code to terminate the ray-cast.
			return 0.0
		}

		// By returning 1, we instruct the caller to continue without clipping the ray.
		return 1.0
	}
}

// This ray cast collects multiple hits along the ray and sorts them.
func rayCastSortedCallback(rayContext *CastContext) box2d.CastResultFunc {
	return func(shape box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
		if shapeIgnored(shape) || fraction == 0.0 {
			return -1.0
		}

		count := rayContext.count

		index := maxCastHits
		for fraction < rayContext.fractions[index-1] {
			index -= 1

			if index == 0 {
				break
			}
		}

		if index == maxCastHits {
			// Once full, ignore hits farther than the largest fraction already acquired.
			return rayContext.fractions[2]
		}

		for j := 2; j > index; j-- {
			rayContext.points[j] = rayContext.points[j-1]
			rayContext.normals[j] = rayContext.normals[j-1]
			rayContext.fractions[j] = rayContext.fractions[j-1]
		}

		rayContext.points[index] = point
		rayContext.normals[index] = normal

		rayContext.fractions[index] = fraction
		if count < maxCastHits {
			rayContext.count = count + 1
		} else {
			rayContext.count = maxCastHits
		}

		if rayContext.count == maxCastHits {
			return rayContext.fractions[2]
		}

		// By returning 1, we instruct the caller to continue without clipping the ray.
		return 1.0
	}
}

func (s *OverlapWorld) overlapResult(shape box2d.Shape) bool {
	if shapeIgnored(shape) {
		// continue the query
		return true
	}

	if s.doomCount < overlapMaxDoomed {
		index := s.doomCount
		s.doomIDs[index] = shape
		s.doomCount += 1
	}

	// continue the query
	return true
}

func (s *Manifold) drawRoundGeometry() {
	const (
		worldCapsuleHalfLength = 0.5
		worldCapsuleRadius     = 0.25
		geometryColumnSpacing  = 4.0
		manifoldCircleRadius   = 0.5
		manifoldBoxHalfExtent  = 0.5
		offsetCapsuleStartX    = 0.25
		thinCapsuleRadius      = 0.1
		tiltedBoxHalfWidth     = 0.25
		tiltedBoxAngleInPi     = 0.25
		geometryStartX         = 10.0
		roundGeometryRowY      = 5.0
	)

	offset := box2d.Vec2{X: -geometryStartX, Y: -roundGeometryRowY}
	increment := box2d.Vec2{X: geometryColumnSpacing, Y: 0.0}
	color1 := box2d.ColorAquamarine
	color2 := box2d.ColorPaleGoldenRod
	transforms := func() (box2d.WorldTransform, box2d.WorldTransform) {
		transform1 := box2d.MakeWorldTransform(box2d.Transform{P: offset, Q: box2d.RotIdentity})
		transform2 := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)

		return transform1, transform2
	}

	// circle-circle
	{
		circle1 := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: manifoldCircleRadius}
		circle2 := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 1.0}

		transform1, transform2 := transforms()

		m := box2d.CollideCircles(&circle1, &circle2, transform1.InvMul(transform2))

		s.draw.DrawSolidCircle(transform1, circle1.Center, circle1.Radius, color1)
		s.draw.DrawSolidCircle(transform2, circle2.Center, circle2.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// capsule-circle
	{
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -worldCapsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: worldCapsuleHalfLength, Y: 0.0},
			Radius:  worldCapsuleRadius,
		}
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: manifoldCircleRadius}

		transform1, transform2 := transforms()

		m := box2d.CollideCapsuleAndCircle(&capsule, &circle, transform1.InvMul(transform2))

		v1 := transform1.Apply(capsule.Center1)
		v2 := transform1.Apply(capsule.Center2)
		s.draw.DrawCapsule(v1, v2, capsule.Radius, color1)
		s.draw.DrawSolidCircle(transform2, circle.Center, circle.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// segment-circle
	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -1.0, Y: 0.0},
			Point2: box2d.Vec2{X: 1.0, Y: 0.0},
		}
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: manifoldCircleRadius}

		transform1, transform2 := transforms()

		m := box2d.CollideSegmentAndCircle(&segment, &circle, transform1.InvMul(transform2))

		p1 := transform1.Apply(segment.Point1)
		p2 := transform1.Apply(segment.Point2)
		s.draw.DrawLine(p1, p2, color1)
		s.draw.DrawSolidCircle(transform2, circle.Center, circle.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// box-circle
	{
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: manifoldCircleRadius}
		box := box2d.MakeSquare(manifoldBoxHalfExtent)
		box.Radius = s.round

		transform1, transform2 := transforms()

		m := box2d.CollidePolygonAndCircle(&box, &circle, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box.Vertices[:box.Count], s.round, color1)
		s.draw.DrawSolidCircle(transform2, circle.Center, circle.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// capsule-capsule
	{
		capsule1 := box2d.Capsule{
			Center1: box2d.Vec2{X: -worldCapsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: worldCapsuleHalfLength, Y: 0.0},
			Radius:  worldCapsuleRadius,
		}
		capsule2 := box2d.Capsule{
			Center1: box2d.Vec2{X: offsetCapsuleStartX, Y: 0.0},
			Center2: box2d.Vec2{X: 1.0, Y: 0.0},
			Radius:  thinCapsuleRadius,
		}

		transform1, transform2 := transforms()

		m := box2d.CollideCapsules(&capsule1, &capsule2, transform1.InvMul(transform2))

		v1 := transform1.Apply(capsule1.Center1)
		v2 := transform1.Apply(capsule1.Center2)
		s.draw.DrawCapsule(v1, v2, capsule1.Radius, color1)

		v1 = transform2.Apply(capsule2.Center1)
		v2 = transform2.Apply(capsule2.Center2)
		s.draw.DrawCapsule(v1, v2, capsule2.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// box-capsule
	{
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -0.4, Y: 0.0},
			Center2: box2d.Vec2{X: -0.1, Y: 0.0},
			Radius:  thinCapsuleRadius,
		}
		box := box2d.MakeOffsetBox(
			tiltedBoxHalfWidth,
			1.0,
			box2d.Vec2{X: 1.0, Y: -1.0},
			box2d.MakeRot(tiltedBoxAngleInPi*box2d.Pi),
		)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygonAndCapsule(&box, &capsule, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box.Vertices[:box.Count], box.Radius, color1)

		v1 := transform2.Apply(capsule.Center1)
		v2 := transform2.Apply(capsule.Center2)
		s.draw.DrawCapsule(v1, v2, capsule.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// segment-capsule
	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -1.0, Y: 0.0},
			Point2: box2d.Vec2{X: 1.0, Y: 0.0},
		}
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -worldCapsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: worldCapsuleHalfLength, Y: 0.0},
			Radius:  worldCapsuleRadius,
		}

		transform1, transform2 := transforms()

		m := box2d.CollideSegmentAndCapsule(&segment, &capsule, transform1.InvMul(transform2))

		p1 := transform1.Apply(segment.Point1)
		p2 := transform1.Apply(segment.Point2)
		s.draw.DrawLine(p1, p2, color1)

		p1 = transform2.Apply(capsule.Center1)
		p2 = transform2.Apply(capsule.Center2)
		s.draw.DrawCapsule(p1, p2, capsule.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}
}

func (s *Manifold) drawPolygonGeometry() {
	const (
		geometryColumnSpacing = 4.0
		manifoldBoxHalfExtent = 0.5
		flatBoxHalfWidth      = 2.0
		flatBoxHalfHeight     = 0.1
		smallBoxHalfExtent    = 0.25
		wedgeFirstTipX        = 0.175740838
		wedgeFirstTipY        = 0.224936664
		wedgeFirstShoulderY   = 0.194021404
		wedgeSecondShoulderX  = 0.0566576123
		wedgeSecondTipX       = 0.176625848
		wedgeSecondTipY       = 0.338923335
		firstWedgeRadius      = 0.158798501
		secondWedgeRadius     = 0.205900759
		geometryStartX        = 10.0
	)

	offset := box2d.Vec2{X: -geometryStartX, Y: 0.0}
	increment := box2d.Vec2{X: geometryColumnSpacing, Y: 0.0}
	color1 := box2d.ColorAquamarine
	color2 := box2d.ColorPaleGoldenRod
	transforms := func() (box2d.WorldTransform, box2d.WorldTransform) {
		transform1 := box2d.MakeWorldTransform(box2d.Transform{P: offset, Q: box2d.RotIdentity})
		transform2 := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)

		return transform1, transform2
	}

	// square-square
	{
		box1 := box2d.MakeSquare(manifoldBoxHalfExtent)
		box := box2d.MakeSquare(manifoldBoxHalfExtent)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&box1, &box, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box1.Vertices[:box1.Count], box1.Radius, color1)
		s.draw.DrawSolidPolygon(transform2, box.Vertices[:box.Count], box.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// box-box
	{
		box1 := box2d.MakeBox(flatBoxHalfWidth, flatBoxHalfHeight)
		box := box2d.MakeSquare(smallBoxHalfExtent)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&box1, &box, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box1.Vertices[:box1.Count], box1.Radius, color1)
		s.draw.DrawSolidPolygon(transform2, box.Vertices[:box.Count], box.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// box-rox
	{
		box := box2d.MakeSquare(manifoldBoxHalfExtent)
		h := manifoldBoxHalfExtent - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&box, &rox, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box.Vertices[:box.Count], box.Radius, color1)
		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// rox-rox
	{
		h := manifoldBoxHalfExtent - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&rox, &rox, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, rox.Vertices[:rox.Count], rox.Radius, color1)
		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// segment-rox
	{
		segment := box2d.Segment{
			Point1: box2d.Vec2{X: -1.0, Y: 0.0},
			Point2: box2d.Vec2{X: 1.0, Y: 0.0},
		}
		h := manifoldBoxHalfExtent - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)

		transform1, transform2 := transforms()

		m := box2d.CollideSegmentAndPolygon(&segment, &rox, transform1.InvMul(transform2))

		p1 := transform1.Apply(segment.Point1)
		p2 := transform1.Apply(segment.Point2)
		s.draw.DrawLine(p1, p2, color1)
		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// wox-wox
	{
		wox := box2d.MakePolygon(&s.wedge, s.round)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&wox, &wox, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, wox.Vertices[:wox.Count], wox.Radius, color1)
		s.draw.DrawSolidPolygon(transform1, wox.Vertices[:wox.Count], 0.0, color1)
		s.draw.DrawSolidPolygon(transform2, wox.Vertices[:wox.Count], wox.Radius, color2)
		s.draw.DrawSolidPolygon(transform2, wox.Vertices[:wox.Count], 0.0, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// wox-wox
	{
		p1s := []box2d.Vec2{
			{X: wedgeFirstTipX, Y: wedgeFirstTipY},
			{X: -0.301293969, Y: wedgeFirstShoulderY},
			{X: -0.105151534, Y: -0.432157338},
		}
		p2s := []box2d.Vec2{
			{X: -0.427884758, Y: -0.225028217},
			{X: wedgeSecondShoulderX, Y: -0.128772855},
			{X: wedgeSecondTipX, Y: wedgeSecondTipY},
		}

		h1 := box2d.ComputeHull(p1s)
		h2 := box2d.ComputeHull(p2s)
		w1 := box2d.MakePolygon(&h1, firstWedgeRadius)
		w2 := box2d.MakePolygon(&h2, secondWedgeRadius)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&w1, &w2, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, w1.Vertices[:w1.Count], w1.Radius, color1)
		s.draw.DrawSolidPolygon(transform1, w1.Vertices[:w1.Count], 0.0, color1)
		s.draw.DrawSolidPolygon(transform2, w2.Vertices[:w2.Count], w2.Radius, color2)
		s.draw.DrawSolidPolygon(transform2, w2.Vertices[:w2.Count], 0.0, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}
}

func (s *Manifold) drawChainGeometry() {
	const (
		worldCapsuleHalfLength  = 0.5
		worldCapsuleRadius      = 0.25
		geometryColumnSpacing   = 4.0
		manifoldCircleRadius    = 0.5
		manifoldBoxHalfExtent   = 0.5
		smallTriangleHalfWidth  = 0.05
		smallTriangleHeight     = 0.1
		chainGhostX             = 2.0
		nextChainGhostX         = 3.0
		nextChainStartX         = 2.0
		chainColumnSpan         = 2.0
		chainVertexSize         = 4.0
		centroidPointSize       = 5.0
		geometryStartX          = 10.0
		chainGeometryRowY       = 5.0
		segmentMidpointFraction = 0.5
	)

	offset := box2d.Vec2{X: -geometryStartX, Y: chainGeometryRowY}
	increment := box2d.Vec2{X: geometryColumnSpacing, Y: 0.0}
	color1 := box2d.ColorAquamarine
	color2 := box2d.ColorPaleGoldenRod
	transforms := func() (box2d.WorldTransform, box2d.WorldTransform) {
		transform1 := box2d.MakeWorldTransform(box2d.Transform{P: offset, Q: box2d.RotIdentity})
		transform2 := box2d.MakeWorldTransform(
			box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q},
		)

		return transform1, transform2
	}

	// box-triangle
	{
		box := box2d.MakeBox(1.0, 1.0)
		points := []box2d.Vec2{
			{X: -smallTriangleHalfWidth, Y: 0.0},
			{X: smallTriangleHalfWidth, Y: 0.0},
			{X: 0.0, Y: smallTriangleHeight},
		}
		hull := box2d.ComputeHull(points)
		tri := box2d.MakePolygon(&hull, 0.0)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&box, &tri, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box.Vertices[:box.Count], 0.0, color1)
		s.draw.DrawSolidPolygon(transform2, tri.Vertices[:tri.Count], 0.0, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// chain-segment vs circle
	{
		segment := box2d.ChainSegment{
			Ghost1: box2d.Vec2{
				X: chainGhostX,
				Y: 1.0,
			},
			Segment: box2d.Segment{
				Point1: box2d.Vec2{X: 1.0, Y: 1.0},
				Point2: box2d.Vec2{X: -1.0, Y: 0.0},
			},
			Ghost2:  box2d.Vec2{X: -2.0, Y: 0.0},
			ChainID: -1,
		}
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: manifoldCircleRadius}

		transform1, transform2 := transforms()

		m := box2d.CollideChainSegmentAndCircle(&segment, &circle, transform1.InvMul(transform2))

		g1 := transform1.Apply(segment.Ghost1)
		g2 := transform1.Apply(segment.Ghost2)
		p1 := transform1.Apply(segment.Segment.Point1)
		p2 := transform1.Apply(segment.Segment.Point2)

		s.draw.DrawLine(g1, p1, box2d.ColorLightGray)
		s.draw.DrawLine(p1, p2, color1)
		s.draw.DrawLine(p2, g2, box2d.ColorLightGray)
		s.draw.DrawSolidCircle(transform2, circle.Center, circle.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset.X += chainColumnSpan * increment.X
	}

	segment1 := box2d.ChainSegment{
		Ghost1: box2d.Vec2{
			X: chainGhostX,
			Y: 1.0,
		},
		Segment: box2d.Segment{
			Point1: box2d.Vec2{X: 1.0, Y: 1.0},
			Point2: box2d.Vec2{X: -1.0, Y: 0.0},
		},
		Ghost2:  box2d.Vec2{X: -2.0, Y: 0.0},
		ChainID: -1,
	}
	segment2 := box2d.ChainSegment{
		Ghost1: box2d.Vec2{
			X: nextChainGhostX,
			Y: 1.0,
		},
		Segment: box2d.Segment{
			Point1: box2d.Vec2{X: nextChainStartX, Y: 1.0},
			Point2: box2d.Vec2{X: 1.0, Y: 1.0},
		},
		Ghost2:  box2d.Vec2{X: -1.0, Y: 0.0},
		ChainID: -1,
	}

	drawChainSegments := func(transform1 box2d.WorldTransform) {
		{
			g2 := transform1.Apply(segment1.Ghost2)
			p1 := transform1.Apply(segment1.Segment.Point1)
			p2 := transform1.Apply(segment1.Segment.Point2)
			s.draw.DrawLine(p1, p2, color1)
			s.draw.DrawPoint(p1, chainVertexSize, color1)
			s.draw.DrawPoint(p2, chainVertexSize, color1)
			s.draw.DrawLine(p2, g2, box2d.ColorLightGray)
		}

		{
			g1 := transform1.Apply(segment2.Ghost1)
			p1 := transform1.Apply(segment2.Segment.Point1)
			p2 := transform1.Apply(segment2.Segment.Point2)

			s.draw.DrawLine(g1, p1, box2d.ColorLightGray)
			s.draw.DrawLine(p1, p2, color1)
			s.draw.DrawPoint(p1, chainVertexSize, color1)
			s.draw.DrawPoint(p2, chainVertexSize, color1)
		}
	}

	// chain-segment vs rounded polygon
	{
		h := manifoldBoxHalfExtent - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)

		transform1, transform2 := transforms()

		xf := transform1.InvMul(transform2)
		m1 := box2d.CollideChainSegmentAndPolygon(&segment1, &rox, xf, &s.smgroxCache1)
		m2 := box2d.CollideChainSegmentAndPolygon(&segment2, &rox, xf, &s.smgroxCache2)

		drawChainSegments(transform1)

		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)
		s.draw.DrawPoint(transform2.Apply(rox.Centroid), centroidPointSize, box2d.ColorGainsboro)

		s.DrawManifold(&m1, transform1, transform2)
		s.DrawManifold(&m2, transform1, transform2)

		offset.X += chainColumnSpan * increment.X
	}

	// chain-segment vs capsule
	{
		capsule := box2d.Capsule{
			Center1: box2d.Vec2{X: -worldCapsuleHalfLength, Y: 0.0},
			Center2: box2d.Vec2{X: worldCapsuleHalfLength, Y: 0.0},
			Radius:  worldCapsuleRadius,
		}

		transform1, transform2 := transforms()

		xf := transform1.InvMul(transform2)
		m1 := box2d.CollideChainSegmentAndCapsule(&segment1, &capsule, xf, &s.smgcapCache1)
		m2 := box2d.CollideChainSegmentAndCapsule(&segment2, &capsule, xf, &s.smgcapCache2)

		drawChainSegments(transform1)

		p1 := transform2.Apply(capsule.Center1)
		p2 := transform2.Apply(capsule.Center2)
		s.draw.DrawCapsule(p1, p2, capsule.Radius, color2)

		s.draw.DrawPoint(
			p1.Lerp(p2, segmentMidpointFraction),
			centroidPointSize,
			box2d.ColorGainsboro,
		)

		s.DrawManifold(&m1, transform1, transform2)
		s.DrawManifold(&m2, transform1, transform2)

		offset.X += chainColumnSpan * increment.X
	}
}

func registerCollisionSamples() {
	RegisterSample("Collision", "Shape Distance", CreateShapeDistance)
	RegisterSample("Collision", "Dynamic Tree", CreateDynamicTree)
	RegisterSample("Collision", "Ray Cast", CreateRayCast)
	RegisterSample("Collision", "Cast World", CreateCastWorld)
	RegisterSample("Collision", "Overlap World", CreateOverlapWorld)
	RegisterSample("Collision", "Manifold", CreateManifold)
	RegisterSample("Collision", "Smooth Manifold", CreateSmoothManifold)
	RegisterSample("Collision", "Shape Cast", CreateShapeCast)
	RegisterSample("Collision", "Time of Impact", CreateTimeOfImpact)
}

func (s *DynamicTree) queryProxies() {
	const (
		pointSize = 5.0
	)

	if s.queryDrag {
		p1 := s.startPoint.Vec2()
		p2 := s.endPoint.Vec2()
		box := box2d.AABB{LowerBound: p1.Min(p2), UpperBound: p1.Max(p2)}
		s.tree.Query(box, box2d.DefaultMaskBits, func(proxyID int, userData uint64) bool {
			proxy := &s.proxies[userData]
			if proxy.proxyID != proxyID {
				panic("proxy id mismatch")
			}

			proxy.queryStamp = s.timeStamp

			return true
		})

		s.draw.DrawBounds(box, box2d.ColorWhite)
	}

	if s.rayDrag {
		input := box2d.RayCastInput{
			Origin:      s.startPoint.Vec2(),
			Translation: s.endPoint.Sub(s.startPoint),
			MaxFraction: 1.0,
		}
		result := s.tree.CastRay(
			&input,
			box2d.DefaultMaskBits,
			func(input *box2d.RayCastInput, proxyID int, userData uint64) float32 {
				proxy := &s.proxies[userData]
				if proxy.proxyID != proxyID {
					panic("proxy id mismatch")
				}

				proxy.rayStamp = s.timeStamp

				return input.MaxFraction
			},
		)

		s.draw.DrawLine(s.startPoint, s.endPoint, box2d.ColorWhite)
		s.draw.DrawPoint(s.startPoint, pointSize, box2d.ColorGreen)
		s.draw.DrawPoint(s.endPoint, pointSize, box2d.ColorRed)

		s.DrawScreenTextLinef(
			"node visits = %d, leaf visits = %d",
			result.NodeVisits,
			result.LeafVisits,
		)
	}
}

func (s *DynamicTree) drawAndMoveProxies() {
	const (
		aabbMarginWidth = 0.1
	)

	c := box2d.ColorBlue
	qc := box2d.ColorGreen

	aabbMargin := box2d.Vec2{X: aabbMarginWidth, Y: aabbMarginWidth}

	for i := range s.proxies {
		p := &s.proxies[i]

		if p.queryStamp == s.timeStamp || p.rayStamp == s.timeStamp {
			s.draw.DrawBounds(p.box, qc)
		} else {
			s.draw.DrawBounds(p.box, c)
		}

		moveTest := scenes.RandomFloatRange(0.0, 1.0)
		if s.moveFraction > moveTest {
			dx := s.moveDelta * scenes.RandomFloat()
			dy := s.moveDelta * scenes.RandomFloat()

			p.position.X += dx
			p.position.Y += dy

			p.box.LowerBound.X = p.position.X + dx
			p.box.LowerBound.Y = p.position.Y + dy
			p.box.UpperBound.X = p.position.X + dx + p.width.X
			p.box.UpperBound.Y = p.position.Y + dy + p.width.Y

			if !p.fatBox.Contains(p.box) {
				p.fatBox.LowerBound = p.box.LowerBound.Sub(aabbMargin)
				p.fatBox.UpperBound = p.box.UpperBound.Add(aabbMargin)
				p.moved = true
			} else {
				p.moved = false
			}
		} else {
			p.moved = false
		}
	}
}

func drawContactLabels(
	sample *Sample,
	mp *box2d.LocalManifoldPoint,
	p1 box2d.Pos,
	showIDs bool,
	showSeparation bool,
) {
	const separationLabelOffsetY = 0.03

	const (
		labelOffsetX = 0.05
	)

	if showIDs {
		sample.draw.DrawStringf(
			sample.camera,
			p1.Add(box2d.Vec2{X: labelOffsetX, Y: -0.02}),
			box2d.ColorWhite,
			"0x%04x",
			mp.ID,
		)
	}

	if showSeparation {
		sample.draw.DrawStringf(
			sample.camera,
			p1.Add(box2d.Vec2{X: labelOffsetX, Y: separationLabelOffsetY}),
			box2d.ColorWhite,
			"%.3f",
			mp.Separation,
		)
	}
}

func drawCollisionSpawnButtons(label string, batchLabel string, index int, create func(int, int)) {
	const batchCount = 10

	if imgui.Button(label) {
		create(index, 1)
	}

	imgui.SameLine()

	if imgui.Button(batchLabel) {
		create(index, batchCount)
	}
}

func (s *CastWorld) drawSimpleRayCast(rayTranslation box2d.Vec2) {
	const (
		pointSize    = 5.0
		normalLength = 0.5
	)

	color1 := box2d.ColorGreen
	color2 := box2d.ColorLightGray
	color3 := box2d.ColorMagenta

	s.DrawScreenTextLinef("Simple closest point ray cast")

	// This version doesn't have a callback, but it doesn't skip the ignored shape
	result := s.worldID.CastRayClosest(s.rayStart, rayTranslation, box2d.DefaultQueryFilter())

	if result.Hit && result.Fraction > 0.0 {
		c := s.rayStart.Add(rayTranslation.Scale(result.Fraction))
		s.draw.DrawPoint(result.Point, pointSize, color1)
		s.draw.DrawLine(s.rayStart, c, color2)

		head := result.Point.Add(result.Normal.Scale(normalLength))
		s.draw.DrawLine(result.Point, head, color3)
	} else {
		s.draw.DrawLine(s.rayStart, s.rayEnd, color2)
	}
}

func (s *CastWorld) castGeometry(
	rayTranslation box2d.Vec2,
	circle *box2d.Circle,
	capsule *box2d.Capsule,
	box *box2d.Polygon,
	modeFcn box2d.CastResultFunc,
) {
	var proxy box2d.ShapeProxy

	if s.castType == castTypeRay {
		s.worldID.CastRay(s.rayStart, rayTranslation, box2d.DefaultQueryFilter(), modeFcn)
	} else {
		switch s.castType {
		case castTypeCircle:
			proxy = box2d.MakeProxy([]box2d.Vec2{circle.Center}, circle.Radius)
		case castTypeCapsule:
			proxy = box2d.MakeProxy(
				[]box2d.Vec2{capsule.Center1, capsule.Center2},
				capsule.Radius,
			)
		default:
			proxy = box2d.MakeProxy(box.Vertices[:box.Count], box.Radius)
		}

		s.worldID.CastShape(
			s.rayStart,
			&proxy,
			rayTranslation,
			box2d.DefaultQueryFilter(),
			modeFcn,
		)
	}
}

func (s *CastWorld) drawCastResults(
	context *CastContext,
	rayTranslation box2d.Vec2,
	capsule *box2d.Capsule,
	box *box2d.Polygon,
) {
	const (
		pointSize = 5.0
	)

	color2 := box2d.ColorLightGray
	color3 := box2d.ColorMagenta

	if context.count > 0 {
		colors := [3]box2d.HexColor{box2d.ColorRed, box2d.ColorGreen, box2d.ColorBlue}

		for i := range context.count {
			c := s.rayStart.Add(rayTranslation.Scale(context.fractions[i]))
			p := context.points[i]
			n := context.normals[i]
			s.draw.DrawPoint(p, pointSize, colors[i])
			s.draw.DrawLine(s.rayStart, c, color2)

			head := p.Add(n.Scale(1.0))
			s.draw.DrawLine(p, head, color3)

			// The swept shape rests at the contact position c
			switch s.castType {
			case castTypeCircle:
				s.draw.DrawSolidCircle(
					box2d.WorldTransform{P: c, Q: box2d.RotIdentity},
					box2d.Vec2Zero,
					s.castRadius,
					box2d.ColorYellow,
				)
			case castTypeCapsule:
				s.draw.DrawCapsule(
					c.Add(capsule.Center1),
					c.Add(capsule.Center2),
					s.castRadius,
					box2d.ColorYellow,
				)
			case castTypePolygon:
				s.draw.DrawSolidPolygon(
					box2d.WorldTransform{P: c, Q: box2d.RotIdentity},
					box.Vertices[:box.Count],
					box.Radius,
					box2d.ColorYellow,
				)
			}
		}
	} else {
		s.draw.DrawLine(s.rayStart, s.rayEnd, color2)

		// No hit, so the swept shape rests at the end of the ray
		switch s.castType {
		case castTypeCircle:
			s.draw.DrawSolidCircle(
				box2d.WorldTransform{P: s.rayEnd, Q: box2d.RotIdentity},
				box2d.Vec2Zero,
				s.castRadius,
				box2d.ColorGray,
			)
		case castTypeCapsule:
			s.draw.DrawCapsule(
				s.rayEnd.Add(capsule.Center1),
				s.rayEnd.Add(capsule.Center2),
				s.castRadius,
				box2d.ColorYellow,
			)
		case castTypePolygon:
			s.draw.DrawSolidPolygon(
				box2d.WorldTransform{P: s.rayEnd, Q: box2d.RotIdentity},
				box.Vertices[:box.Count],
				box.Radius,
				box2d.ColorYellow,
			)
		}
	}
}

func (s *CastWorld) drawWorldCast(rayTranslation box2d.Vec2) {
	const (
		castCapsuleHalfLength = 0.25
		castBoxHalfWidth      = 0.125
		castBoxHalfHeight     = 0.25
	)

	switch s.mode {
	case castModeAny:
		s.DrawScreenTextLinef("Cast mode: any - check for obstruction - unsorted")

	case castModeClosest:
		s.DrawScreenTextLinef("Cast mode: closest - find closest shape along the cast")

	case castModeMultiple:
		s.DrawScreenTextLinef("Cast mode: multiple - gather up to 3 shapes - unsorted")

	case castModeSorted:
		s.DrawScreenTextLinef("Cast mode: sorted - gather up to 3 shapes sorted by closeness")

	default:
		panic("unknown cast mode")
	}

	var context CastContext

	// Must initialize fractions for sorting
	context.fractions[0] = math.MaxFloat32
	context.fractions[1] = math.MaxFloat32
	context.fractions[2] = math.MaxFloat32

	functions := [4]box2d.CastResultFunc{
		rayCastAnyCallback(&context),
		rayCastClosestCallback(&context),
		rayCastMultipleCallback(&context),
		rayCastSortedCallback(&context),
	}
	modeFcn := functions[s.mode]

	rotation := box2d.MakeRot(s.angle)
	circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: s.castRadius}
	capsule := box2d.Capsule{
		Center1: rotation.Rotate(
			box2d.Vec2{X: -castCapsuleHalfLength, Y: 0.0},
		),
		Center2: rotation.Rotate(box2d.Vec2{X: castCapsuleHalfLength, Y: 0.0}),
		Radius:  s.castRadius,
	}
	box := box2d.MakeOffsetRoundedBox(
		castBoxHalfWidth,
		castBoxHalfHeight,
		box2d.Vec2Zero,
		rotation,
		s.castRadius,
	)

	s.castGeometry(rayTranslation, &circle, &capsule, &box, modeFcn)

	s.drawCastResults(&context, rayTranslation, &capsule, &box)
}

func collisionMouseRotation(
	baseAngle float32,
	delta float32,
	sensitivity float32,
) (float32, box2d.Rot) {
	angle := clampFloat(baseAngle+sensitivity*delta, -box2d.Pi, box2d.Pi)
	return angle, box2d.MakeRot(angle)
}
