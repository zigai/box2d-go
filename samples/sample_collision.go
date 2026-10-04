// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/scenes"
)

type proxyShapeType int32

const (
	shapePoint proxyShapeType = iota
	shapeSegment
	shapeTriangle
	shapeBox
)

var proxyShapeTypes = []string{"point", "segment", "triangle", "box"}

type proxyShapes struct {
	box      box2d.Polygon
	triangle box2d.Polygon
	point    box2d.Vec2
	segment  box2d.Segment
}

func (p *proxyShapes) makeProxy(shapeType proxyShapeType, radius float32) box2d.ShapeProxy {
	var proxy box2d.ShapeProxy
	proxy.Radius = radius

	switch shapeType {
	case shapePoint:
		proxy.Points[0] = box2d.Vec2Zero
		proxy.Count = 1

	case shapeSegment:
		proxy.Points[0] = p.segment.Point1
		proxy.Points[1] = p.segment.Point2
		proxy.Count = 2

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
		proxy.Count = 4

	default:
		panic("unknown shape type")
	}

	return proxy
}

func (p *proxyShapes) drawShape(draw *Draw, shapeType proxyShapeType, transform box2d.WorldTransform, radius float32, color box2d.HexColor) {
	switch shapeType {
	case shapePoint:
		point := transform.Apply(p.point)
		if radius > 0.0 {
			draw.DrawSolidCircle(box2d.WorldTransform{P: point, Q: transform.Q}, box2d.Vec2Zero, radius, color)
		} else {
			draw.DrawPoint(point, 5.0, color)
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

const simplexCapacity = 20

type ShapeDistance struct {
	Sample
	proxyShapes

	typeA   proxyShapeType
	typeB   proxyShapeType
	radiusA float32
	radiusB float32
	proxyA  box2d.ShapeProxy
	proxyB  box2d.ShapeProxy

	cache        box2d.SimplexCache
	simplexes    [simplexCapacity]box2d.Simplex
	simplexCount int32
	simplexIndex int32

	transform box2d.Transform
	angle     float32

	basePosition box2d.Vec2
	startPoint   box2d.Pos
	baseAngle    float32

	dragging    bool
	rotating    bool
	showIndices bool
	useCache    bool
	drawSimplex bool
}

func CreateShapeDistance(context *SampleContext) Sampler {
	s := &ShapeDistance{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.0}
		s.context.camera.zoom = 3.0
	}

	s.point = box2d.Vec2Zero
	s.segment = box2d.Segment{Point1: box2d.Vec2{X: -0.5, Y: 0.0}, Point2: box2d.Vec2{X: 0.5, Y: 0.0}}

	{
		points := []box2d.Vec2{{X: -0.5, Y: 0.0}, {X: 0.5, Y: 0.0}, {X: 0.0, Y: 1.0}}
		hull := box2d.ComputeHull(points)
		s.triangle = box2d.MakePolygon(&hull, 0.0)
	}

	s.box = box2d.MakeSquare(0.5)

	s.transform = box2d.Transform{P: box2d.Vec2{X: 0.0, Y: 0.0}, Q: box2d.RotIdentity}
	s.angle = 0.0

	s.cache = box2d.SimplexCache{}
	s.simplexCount = 0
	s.simplexIndex = 0
	s.startPoint = box2d.Pos{X: 0.0, Y: 0.0}
	s.basePosition = box2d.Vec2{X: 0.0, Y: 0.0}
	s.baseAngle = 0.0

	s.dragging = false
	s.rotating = false
	s.showIndices = false
	s.useCache = false
	s.drawSimplex = false

	s.typeA = shapeBox
	s.typeB = shapeBox
	s.radiusA = 0.0
	s.radiusB = 0.0

	s.proxyA = s.makeProxy(s.typeA, s.radiusA)
	s.proxyB = s.makeProxy(s.typeB, s.radiusB)
	return s
}

func (s *ShapeDistance) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	shapeType := int32(s.typeA)
	if imgui.ComboStrarr("shape A", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeA = proxyShapeType(shapeType)
		s.proxyA = s.makeProxy(s.typeA, s.radiusA)
	}

	if imgui.SliderFloatV("radius A", &s.radiusA, 0.0, 0.5, "%.2f", 0) {
		s.proxyA.Radius = s.radiusA
	}

	shapeType = int32(s.typeB)
	if imgui.ComboStrarr("shape B", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeB = proxyShapeType(shapeType)
		s.proxyB = s.makeProxy(s.typeB, s.radiusB)
	}

	if imgui.SliderFloatV("radius B", &s.radiusB, 0.0, 0.5, "%.2f", 0) {
		s.proxyB.Radius = s.radiusB
	}

	imgui.Separator()

	imgui.SliderFloatV("x offset", &s.transform.P.X, -2.0, 2.0, "%.2f", 0)
	imgui.SliderFloatV("y offset", &s.transform.P.Y, -2.0, 2.0, "%.2f", 0)

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
		imgui.PushItemWidth(6.0 * imgui.FontSize())
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
	d := p.Sub(s.startPoint)

	if s.dragging {
		s.transform.P = s.basePosition.MulAdd(0.5, d)
	} else if s.rotating {
		s.angle = clampFloat(s.baseAngle+1.0*d.X, -box2d.Pi, box2d.Pi)
		s.transform.Q = box2d.MakeRot(s.angle)
	}
}

func weight2(a1 float32, w1 box2d.Vec2, a2 float32, w2 box2d.Vec2) box2d.Vec2 {
	return box2d.Vec2{X: a1*w1.X + a2*w2.X, Y: a1*w1.Y + a2*w2.Y}
}

func weight3(a1 float32, w1 box2d.Vec2, a2 float32, w2 box2d.Vec2, a3 float32, w3 box2d.Vec2) box2d.Vec2 {
	return box2d.Vec2{X: a1*w1.X + a2*w2.X + a3*w3.X, Y: a1*w1.Y + a2*w2.Y + a3*w3.Y}
}

func computeSimplexWitnessPoints(simplex *box2d.Simplex) (box2d.Vec2, box2d.Vec2) {
	switch simplex.Count {
	case 1:
		return simplex.V1.WA, simplex.V1.WB

	case 2:
		a := weight2(simplex.V1.A, simplex.V1.WA, simplex.V2.A, simplex.V2.WA)
		b := weight2(simplex.V1.A, simplex.V1.WB, simplex.V2.A, simplex.V2.WB)
		return a, b

	case 3:
		a := weight3(simplex.V1.A, simplex.V1.WA, simplex.V2.A, simplex.V2.WA, simplex.V3.A, simplex.V3.WA)
		return a, a

	default:
		panic("invalid simplex count")
	}
}

func (s *ShapeDistance) Step() {
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
	s.drawShape(s.draw, s.typeB, box2d.MakeWorldTransform(s.transform), s.radiusB, box2d.ColorBisque)

	if s.drawSimplex {
		simplex := &s.simplexes[s.simplexIndex]
		vertices := [3]*box2d.SimplexVertex{&simplex.V1, &simplex.V2, &simplex.V3}

		if s.simplexIndex > 0 {
			pointA, pointB := computeSimplexWitnessPoints(simplex)

			s.draw.DrawLine(pointA.Pos(), pointB.Pos(), box2d.ColorWhite)
			s.draw.DrawPoint(pointA.Pos(), 10.0, box2d.ColorWhite)
			s.draw.DrawPoint(pointB.Pos(), 10.0, box2d.ColorWhite)
		}

		colors := [3]box2d.HexColor{box2d.ColorRed, box2d.ColorGreen, box2d.ColorBlue}

		for i := range simplex.Count {
			vertex := vertices[i]
			s.draw.DrawPoint(vertex.WA.Pos(), 10.0, colors[i])
			s.draw.DrawPoint(vertex.WB.Pos(), 10.0, colors[i])
		}
	} else {
		s.draw.DrawLine(output.PointA.Pos(), output.PointB.Pos(), box2d.ColorDimGray)
		s.draw.DrawPoint(output.PointA.Pos(), 10.0, box2d.ColorWhite)
		s.draw.DrawPoint(output.PointB.Pos(), 10.0, box2d.ColorWhite)

		s.draw.DrawLine(output.PointA.Pos(), output.PointA.MulAdd(0.5, output.Normal).Pos(), box2d.ColorYellow)
	}

	if s.showIndices {
		for i := range s.proxyA.Count {
			s.draw.DrawString(s.camera, s.proxyA.Points[i].Pos(), box2d.ColorWhite, " %d", i)
		}

		transformB := box2d.MakeWorldTransform(s.transform)
		for i := range s.proxyB.Count {
			p := transformB.Apply(s.proxyB.Points[i])
			s.draw.DrawString(s.camera, p, box2d.ColorWhite, " %d", i)
		}
	}

	s.DrawScreenTextLine("mouse button 1: drag")
	s.DrawScreenTextLine("mouse button 1 + shift: rotate")
	s.DrawScreenTextLine("distance = %.2f, iterations = %d", output.Distance, output.Iterations)

	switch s.cache.Count {
	case 1:
		s.DrawScreenTextLine("cache = {%d}, {%d}", s.cache.IndexA[0], s.cache.IndexB[0])
	case 2:
		s.DrawScreenTextLine("cache = {%d, %d}, {%d, %d}", s.cache.IndexA[0], s.cache.IndexA[1], s.cache.IndexB[0], s.cache.IndexB[1])
	case 3:
		s.DrawScreenTextLine("cache = {%d, %d, %d}, {%d, %d, %d}", s.cache.IndexA[0], s.cache.IndexA[1], s.cache.IndexA[2],
			s.cache.IndexB[0], s.cache.IndexB[1], s.cache.IndexB[2])
	}
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

	tree         box2d.DynamicTree
	rowCount     int32
	columnCount  int32
	proxies      []treeProxy
	timeStamp    int
	fill         float32
	moveFraction float32
	moveDelta    float32
	ratio        float32
	grid         float32

	startPoint box2d.Pos
	endPoint   box2d.Pos

	rayDrag   bool
	queryDrag bool
	validate  bool
}

func CreateDynamicTree(context *SampleContext) Sampler {
	s := &DynamicTree{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 500.0, Y: 500.0}
		s.context.camera.zoom = 25.0 * 21.0
	}

	s.fill = 0.25
	s.moveFraction = 0.05
	s.moveDelta = 0.1
	s.ratio = 5.0
	s.grid = 1.0

	if isDebug {
		s.rowCount = 100
		s.columnCount = 100
	} else {
		s.rowCount = 1000
		s.columnCount = 1000
	}
	s.BuildTree()
	s.timeStamp = 0

	s.startPoint = box2d.Pos{X: 0.0, Y: 0.0}
	s.endPoint = box2d.Pos{X: 0.0, Y: 0.0}
	s.queryDrag = false
	s.rayDrag = false
	s.validate = true
	return s
}

func (s *DynamicTree) BuildTree() {
	proxyCapacity := int(s.rowCount * s.columnCount)
	s.proxies = make([]treeProxy, 0, proxyCapacity)

	y := float32(-4.0)

	s.tree = box2d.NewDynamicTree(16)

	aabbMargin := box2d.Vec2{X: 0.1, Y: 0.1}

	for range s.rowCount {
		x := float32(-40.0)

		for range s.columnCount {
			fillTest := scenes.RandomFloatRange(0.0, 1.0)
			if fillTest <= s.fill {
				var p treeProxy
				p.position = box2d.Vec2{X: x, Y: y}

				ratio := scenes.RandomFloatRange(1.0, s.ratio)
				width := scenes.RandomFloatRange(0.1, 0.5)
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

				p.proxyID = s.tree.CreateProxy(p.fatBox, box2d.DefaultCategoryBits, uint64(len(s.proxies)))
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
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	changed := false
	if imgui.SliderIntV("rows", &s.rowCount, 0, 1000, "%d", 0) {
		changed = true
	}

	if imgui.SliderIntV("columns", &s.columnCount, 0, 1000, "%d", 0) {
		changed = true
	}

	if imgui.SliderFloatV("fill", &s.fill, 0.0, 1.0, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("grid", &s.grid, 0.5, 2.0, "%.2f", 0) {
		changed = true
	}

	if imgui.SliderFloatV("ratio", &s.ratio, 1.0, 10.0, "%.2f", 0) {
		changed = true
	}

	imgui.SliderFloatV("move", &s.moveFraction, 0.0, 1.0, "%.2f", 0)
	imgui.SliderFloatV("delta", &s.moveDelta, 0.0, 1.0, "%.2f", 0)

	imgui.PopItemWidth()

	imgui.Separator()

	imText("mouse button 1: ray cast")
	imText("mouse button 1 + shift: query")

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
		input := box2d.RayCastInput{Origin: s.startPoint.Vec2(), Translation: s.endPoint.Sub(s.startPoint), MaxFraction: 1.0}
		result := s.tree.CastRay(&input, box2d.DefaultMaskBits, func(input *box2d.RayCastInput, proxyID int, userData uint64) float32 {
			proxy := &s.proxies[userData]
			if proxy.proxyID != proxyID {
				panic("proxy id mismatch")
			}
			proxy.rayStamp = s.timeStamp
			return input.MaxFraction
		})

		s.draw.DrawLine(s.startPoint, s.endPoint, box2d.ColorWhite)
		s.draw.DrawPoint(s.startPoint, 5.0, box2d.ColorGreen)
		s.draw.DrawPoint(s.endPoint, 5.0, box2d.ColorRed)

		s.DrawScreenTextLine("node visits = %d, leaf visits = %d", result.NodeVisits, result.LeafVisits)
	}

	c := box2d.ColorBlue
	qc := box2d.ColorGreen

	aabbMargin := box2d.Vec2{X: 0.1, Y: 0.1}

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

	{
		ticks := time.Now()
		for i := range s.proxies {
			p := &s.proxies[i]
			if p.moved {
				s.tree.MoveProxy(p.proxyID, p.fatBox)
			}
		}
		ms := millisecondsSince(ticks)
		s.DrawScreenTextLine("incremental : %.3f ms", ms)
	}

	height := s.tree.Height()
	areaRatio := s.tree.AreaRatio()

	proxyCount := len(s.proxies)
	hmin := int(float32(math.Ceil(float64(float32(math.Log(float64(proxyCount)))/float32(math.Log(2.0)) - 1.0))))
	s.DrawScreenTextLine("proxies = %d, height = %d, hmin = %d, area ratio = %.1f", proxyCount, height, hmin, areaRatio)

	s.tree.Validate()

	s.timeStamp += 1
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

	basePosition box2d.Vec2
	baseAngle    float32

	startPosition box2d.Pos

	rayDrag      bool
	translating  bool
	rotating     bool
	showFraction bool
}

func CreateRayCast(context *SampleContext) Sampler {
	s := &RayCast{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 20.0}
		s.context.camera.zoom = 17.5
	}

	s.circle = box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 2.0}
	s.capsule = box2d.Capsule{Center1: box2d.Vec2{X: -1.0, Y: 1.0}, Center2: box2d.Vec2{X: 1.0, Y: -1.0}, Radius: 1.5}
	s.box = box2d.MakeBox(2.0, 2.0)

	vertices := []box2d.Vec2{{X: -2.0, Y: 0.0}, {X: 2.0, Y: 0.0}, {X: 2.0, Y: 3.0}}
	hull := box2d.ComputeHull(vertices)
	s.triangle = box2d.MakePolygon(&hull, 0.0)

	s.segment = box2d.Segment{Point1: box2d.Vec2{X: -3.0, Y: 0.0}, Point2: box2d.Vec2{X: 3.0, Y: 0.0}}

	s.transform = box2d.TransformIdentity
	s.angle = 0.0

	s.basePosition = box2d.Vec2{X: 0.0, Y: 0.0}
	s.baseAngle = 0.0
	s.startPosition = box2d.Pos{X: 0.0, Y: 0.0}

	s.rayStart = box2d.Pos{X: 0.0, Y: 30.0}
	s.rayEnd = box2d.Pos{X: 0.0, Y: 0.0}

	s.rayDrag = false
	s.translating = false
	s.rotating = false

	s.showFraction = false
	return s
}

func (s *RayCast) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	imgui.SliderFloatV("x offset", &s.transform.P.X, -2.0, 2.0, "%.2f", 0)
	imgui.SliderFloatV("y offset", &s.transform.P.Y, -2.0, 2.0, "%.2f", 0)

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

	imText("mouse btn 1: ray cast")
	imText("mouse btn 1 + shft: translate")
	imText("mouse btn 1 + ctrl: rotate")

	return true
}

func (s *RayCast) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		s.startPosition = p

		if mods == 0 {
			s.rayStart = p
			s.rayDrag = true
		} else if mods == glfw.ModShift {
			s.translating = true
			s.basePosition = s.transform.P
		} else if mods == glfw.ModControl {
			s.rotating = true
			s.baseAngle = s.angle
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
	d := p.Sub(s.startPosition)

	if s.rayDrag {
		s.rayEnd = p
	} else if s.translating {
		s.transform.P.X = s.basePosition.X + 0.5*d.X
		s.transform.P.Y = s.basePosition.Y + 0.5*d.Y
	} else if s.rotating {
		s.angle = clampFloat(s.baseAngle+0.5*d.X, -box2d.Pi, box2d.Pi)
		s.transform.Q = box2d.MakeRot(s.angle)
	}
}

func (s *RayCast) DrawRay(output *box2d.CastOutput, hitPoint box2d.Pos) {
	p1 := s.rayStart
	p2 := s.rayEnd
	d := p2.Sub(p1)

	if output.Hit {
		var p box2d.Pos

		if output.Fraction == 0.0 {
			p = hitPoint
			s.draw.DrawPoint(hitPoint, 5.0, box2d.ColorPeru)
		} else {
			p = p1.Add(d.Scale(output.Fraction))
			s.draw.DrawLine(p1, p, box2d.ColorWhite)
			s.draw.DrawPoint(p1, 5.0, box2d.ColorGreen)
			s.draw.DrawPoint(hitPoint, 5.0, box2d.ColorWhite)

			s.draw.DrawLine(p, p.Add(output.Normal), box2d.ColorViolet)
		}

		if s.showFraction {
			s.draw.DrawString(s.camera, p.Add(box2d.Vec2{X: 0.05, Y: -0.02}), box2d.ColorWhite, "%.2f", output.Fraction)
		}
	} else {
		s.draw.DrawLine(p1, p2, box2d.ColorWhite)
		s.draw.DrawPoint(p1, 5.0, box2d.ColorGreen)
		s.draw.DrawPoint(p2, 5.0, box2d.ColorRed)
	}
}

func (s *RayCast) Step() {
	offset := box2d.Vec2{X: -20.0, Y: 20.0}
	increment := box2d.Vec2{X: 10.0, Y: 0.0}

	color1 := box2d.ColorYellow

	var output box2d.CastOutput
	var hitPoint box2d.Pos
	maxFraction := float32(1.0)

	cast := func(transform box2d.WorldTransform, castFcn func(input *box2d.RayCastInput) box2d.CastOutput) {
		start := transform.InvApply(s.rayStart)
		translation := transform.Q.InvRotate(s.rayEnd.Sub(s.rayStart))
		input := box2d.RayCastInput{Origin: start, Translation: translation, MaxFraction: maxFraction}

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
		transform := box2d.MakeWorldTransform(box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q})
		s.draw.DrawSolidCircle(transform, s.circle.Center, s.circle.Radius, color1)

		cast(transform, func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCircle(&s.circle, input) })
	}

	{
		transform := box2d.MakeWorldTransform(box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q})
		v1 := transform.Apply(s.capsule.Center1)
		v2 := transform.Apply(s.capsule.Center2)
		s.draw.DrawCapsule(v1, v2, s.capsule.Radius, color1)

		cast(transform, func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCapsule(&s.capsule, input) })
	}

	{
		transform := box2d.MakeWorldTransform(box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q})
		s.draw.DrawSolidPolygon(transform, s.box.Vertices[:s.box.Count], 0.0, color1)

		cast(transform, func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastPolygon(&s.box, input) })
	}

	{
		transform := box2d.MakeWorldTransform(box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q})
		s.draw.DrawSolidPolygon(transform, s.triangle.Vertices[:s.triangle.Count], 0.0, color1)

		cast(transform, func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastPolygon(&s.triangle, input) })
	}

	{
		transform := box2d.MakeWorldTransform(box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q})

		p1 := transform.Apply(s.segment.Point1)
		p2 := transform.Apply(s.segment.Point2)
		s.draw.DrawLine(p1, p2, color1)

		cast(transform, func(input *box2d.RayCastInput) box2d.CastOutput {
			return box2d.RayCastSegment(&s.segment, input, false)
		})
	}

	s.DrawRay(&output, hitPoint)
}

type ShapeUserData struct {
	index  int
	ignore bool
}

type CastContext struct {
	points    [3]box2d.Pos
	normals   [3]box2d.Vec2
	fractions [3]float32
	count     int
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
// NOTE: shape hits are not ordered, so this may not return the closest hit
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

		if rayContext.count == 3 {
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

		index := 3
		for fraction < rayContext.fractions[index-1] {
			index -= 1

			if index == 0 {
				break
			}
		}

		if index == 3 {
			// not closer, continue but tell the caller not to consider fractions further than the largest fraction acquired
			// this only happens once the buffer is full
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
		if count < 3 {
			rayContext.count = count + 1
		} else {
			rayContext.count = 3
		}

		if rayContext.count == 3 {
			return rayContext.fractions[2]
		}

		// By returning 1, we instruct the caller to continue without clipping the ray.
		return 1.0
	}
}

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

// This sample shows how to use the ray and shape cast functions on a b2World. This
// sample is configured to ignore initial overlap.
type CastWorld struct {
	Sample

	bodyIndex int
	bodyIDs   [castWorldMaxCount]box2d.Body
	userData  [castWorldMaxCount]ShapeUserData
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

	angleAnchor box2d.Pos
	baseAngle   float32
	angle       float32
	rotating    bool

	rayStart box2d.Pos
	rayEnd   box2d.Pos
	dragging bool
}

func CreateCastWorld(context *SampleContext) Sampler {
	s := &CastWorld{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 14.0}
		s.context.camera.zoom = 25.0 * 0.75
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -40.0, Y: 0.0}, Point2: box2d.Vec2{X: 40.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	{
		vertices := []box2d.Vec2{{X: -0.1, Y: 0.0}, {X: 0.1, Y: 0.0}, {X: 0.0, Y: 1.5}}
		hull := box2d.ComputeHull(vertices)
		s.polygons[0] = box2d.MakePolygon(&hull, 0.0)
		s.polygons[0].Radius = 0.5
	}

	{
		w := float32(1.0)
		b := w / (2.0 + float32(math.Sqrt(2.0)))
		sq := float32(math.Sqrt(2.0)) * b

		vertices := []box2d.Vec2{{X: 0.5 * sq, Y: 0.0}, {X: 0.5 * w, Y: b}, {X: 0.5 * w, Y: b + sq}, {X: 0.5 * sq, Y: w},
			{X: -0.5 * sq, Y: w}, {X: -0.5 * w, Y: b + sq}, {X: -0.5 * w, Y: b}, {X: -0.5 * sq, Y: 0.0}}

		hull := box2d.ComputeHull(vertices)
		s.polygons[1] = box2d.MakePolygon(&hull, 0.0)
	}

	s.box = box2d.MakeBox(0.5, 0.5)
	s.capsule = box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
	s.circle = box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	s.segment = box2d.Segment{Point1: box2d.Vec2{X: -1.0, Y: 0.0}, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}

	s.bodyIndex = 0

	s.mode = castModeClosest
	s.ignoreIndex = 7

	s.castType = castTypeRay
	s.castRadius = 0.5

	s.rayStart = box2d.Pos{X: -20.0, Y: 10.0}
	s.rayEnd = box2d.Pos{X: 20.0, Y: 10.0}
	s.dragging = false

	s.angle = 0.0
	s.baseAngle = 0.0
	s.angleAnchor = box2d.Pos{X: 0.0, Y: 0.0}
	s.rotating = false

	s.simple = false
	return s
}

func (s *CastWorld) Create(index int) {
	if s.bodyIDs[s.bodyIndex] != (box2d.Body{}) {
		s.bodyIDs[s.bodyIndex].Destroy()
		s.bodyIDs[s.bodyIndex] = box2d.Body{}
	}

	x := scenes.RandomFloatRange(-20.0, 20.0)
	y := scenes.RandomFloatRange(0.0, 20.0)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{X: posScalar(x), Y: posScalar(y)}
	bodyDef.Rotation = box2d.MakeRot(scenes.RandomFloatRange(-box2d.Pi, box2d.Pi))

	mod := s.bodyIndex % 3
	if mod == 0 {
		bodyDef.Type = box2d.StaticBody
	} else if mod == 1 {
		bodyDef.Type = box2d.KinematicBody
	} else if mod == 2 {
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

	if index == 0 {
		polygonIndex := s.bodyIndex & 1
		s.bodyIDs[s.bodyIndex].CreatePolygonShape(&shapeDef, &s.polygons[polygonIndex])
	} else if index == 1 {
		s.bodyIDs[s.bodyIndex].CreatePolygonShape(&shapeDef, &s.box)
	} else if index == 2 {
		s.bodyIDs[s.bodyIndex].CreateCircleShape(&shapeDef, &s.circle)
	} else if index == 3 {
		s.bodyIDs[s.bodyIndex].CreateCapsuleShape(&shapeDef, &s.capsule)
	} else if index == 4 {
		s.bodyIDs[s.bodyIndex].CreateSegmentShape(&shapeDef, &s.segment)
	} else {
		points := []box2d.Vec2{{X: 1.0, Y: 0.0}, {X: -1.0, Y: 0.0}, {X: -1.0, Y: -1.0}, {X: 1.0, Y: -1.0}}
		chainDef := box2d.DefaultChainDef()
		chainDef.UserData = shapeDef.UserData
		chainDef.Points = points
		chainDef.IsLoop = true
		s.bodyIDs[s.bodyIndex].CreateChain(&chainDef)
	}

	s.bodyIndex = (s.bodyIndex + 1) % castWorldMaxCount
}

func (s *CastWorld) CreateN(index, count int) {
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
	imgui.Checkbox("Simple", &s.simple)

	if !s.simple {
		imgui.PushItemWidth(6.0 * imgui.FontSize())
		castTypes := []string{"Ray", "Circle", "Capsule", "Polygon"}
		imgui.ComboStrarr("Type", &s.castType, castTypes, int32(len(castTypes)))

		if s.castType != castTypeRay {
			imgui.SliderFloatV("Radius", &s.castRadius, 0.0, 2.0, "%.1f", 0)
		}

		modes := []string{"Any", "Closest", "Multiple", "Sorted"}
		imgui.ComboStrarr("Mode", &s.mode, modes, int32(len(modes)))
		imgui.PopItemWidth()
	}

	if imgui.Button("Polygon") {
		s.Create(0)
	}
	imgui.SameLine()
	if imgui.Button("10x##Poly") {
		s.CreateN(0, 10)
	}

	if imgui.Button("Box") {
		s.Create(1)
	}
	imgui.SameLine()
	if imgui.Button("10x##Box") {
		s.CreateN(1, 10)
	}

	if imgui.Button("Circle") {
		s.Create(2)
	}
	imgui.SameLine()
	if imgui.Button("10x##Circle") {
		s.CreateN(2, 10)
	}

	if imgui.Button("Capsule") {
		s.Create(3)
	}
	imgui.SameLine()
	if imgui.Button("10x##Capsule") {
		s.CreateN(3, 10)
	}

	if imgui.Button("Segment") {
		s.Create(4)
	}
	imgui.SameLine()
	if imgui.Button("10x##Segment") {
		s.CreateN(4, 10)
	}

	if imgui.Button("Chain") {
		s.Create(5)
	}
	imgui.SameLine()
	if imgui.Button("10x##Chain") {
		s.CreateN(5, 10)
	}

	if imgui.Button("Destroy Shape") {
		s.DestroyBody()
	}

	return true
}

func (s *CastWorld) Step() {
	s.Sample.Step()

	s.DrawScreenTextLine("Click left mouse button and drag to modify ray cast")
	s.DrawScreenTextLine("Shape 7 is intentionally ignored by the ray")

	color1 := box2d.ColorGreen
	color2 := box2d.ColorLightGray
	color3 := box2d.ColorMagenta

	rayTranslation := s.rayEnd.Sub(s.rayStart)

	if s.simple {
		s.DrawScreenTextLine("Simple closest point ray cast")

		// This version doesn't have a callback, but it doesn't skip the ignored shape
		result := s.worldID.CastRayClosest(s.rayStart, rayTranslation, box2d.DefaultQueryFilter())

		if result.Hit && result.Fraction > 0.0 {
			c := s.rayStart.Add(rayTranslation.Scale(result.Fraction))
			s.draw.DrawPoint(result.Point, 5.0, color1)
			s.draw.DrawLine(s.rayStart, c, color2)
			head := result.Point.Add(result.Normal.Scale(0.5))
			s.draw.DrawLine(result.Point, head, color3)
		} else {
			s.draw.DrawLine(s.rayStart, s.rayEnd, color2)
		}
	} else {
		switch s.mode {
		case castModeAny:
			s.DrawScreenTextLine("Cast mode: any - check for obstruction - unsorted")

		case castModeClosest:
			s.DrawScreenTextLine("Cast mode: closest - find closest shape along the cast")

		case castModeMultiple:
			s.DrawScreenTextLine("Cast mode: multiple - gather up to 3 shapes - unsorted")

		case castModeSorted:
			s.DrawScreenTextLine("Cast mode: sorted - gather up to 3 shapes sorted by closeness")

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
		capsule := box2d.Capsule{Center1: rotation.Rotate(box2d.Vec2{X: -0.25, Y: 0.0}), Center2: rotation.Rotate(box2d.Vec2{X: 0.25, Y: 0.0}),
			Radius: s.castRadius}
		box := box2d.MakeOffsetRoundedBox(0.125, 0.25, box2d.Vec2Zero, rotation, s.castRadius)
		var proxy box2d.ShapeProxy

		if s.castType == castTypeRay {
			s.worldID.CastRay(s.rayStart, rayTranslation, box2d.DefaultQueryFilter(), modeFcn)
		} else {
			if s.castType == castTypeCircle {
				proxy = box2d.MakeProxy([]box2d.Vec2{circle.Center}, circle.Radius)
			} else if s.castType == castTypeCapsule {
				proxy = box2d.MakeProxy([]box2d.Vec2{capsule.Center1, capsule.Center2}, capsule.Radius)
			} else {
				proxy = box2d.MakeProxy(box.Vertices[:box.Count], box.Radius)
			}

			s.worldID.CastShape(s.rayStart, &proxy, rayTranslation, box2d.DefaultQueryFilter(), modeFcn)
		}

		if context.count > 0 {
			colors := [3]box2d.HexColor{box2d.ColorRed, box2d.ColorGreen, box2d.ColorBlue}
			for i := range context.count {
				c := s.rayStart.Add(rayTranslation.Scale(context.fractions[i]))
				p := context.points[i]
				n := context.normals[i]
				s.draw.DrawPoint(p, 5.0, colors[i])
				s.draw.DrawLine(s.rayStart, c, color2)
				head := p.Add(n.Scale(1.0))
				s.draw.DrawLine(p, head, color3)

				// The swept shape rests at the contact position c
				if s.castType == castTypeCircle {
					s.draw.DrawSolidCircle(box2d.WorldTransform{P: c, Q: box2d.RotIdentity}, box2d.Vec2Zero, s.castRadius, box2d.ColorYellow)
				} else if s.castType == castTypeCapsule {
					s.draw.DrawCapsule(c.Add(capsule.Center1), c.Add(capsule.Center2), s.castRadius, box2d.ColorYellow)
				} else if s.castType == castTypePolygon {
					s.draw.DrawSolidPolygon(box2d.WorldTransform{P: c, Q: box2d.RotIdentity}, box.Vertices[:box.Count], box.Radius, box2d.ColorYellow)
				}
			}
		} else {
			s.draw.DrawLine(s.rayStart, s.rayEnd, color2)

			// No hit, so the swept shape rests at the end of the ray
			if s.castType == castTypeCircle {
				s.draw.DrawSolidCircle(box2d.WorldTransform{P: s.rayEnd, Q: box2d.RotIdentity}, box2d.Vec2Zero, s.castRadius, box2d.ColorGray)
			} else if s.castType == castTypeCapsule {
				s.draw.DrawCapsule(s.rayEnd.Add(capsule.Center1), s.rayEnd.Add(capsule.Center2), s.castRadius, box2d.ColorYellow)
			} else if s.castType == castTypePolygon {
				s.draw.DrawSolidPolygon(box2d.WorldTransform{P: s.rayEnd, Q: box2d.RotIdentity}, box.Vertices[:box.Count], box.Radius, box2d.ColorYellow)
			}
		}
	}

	s.draw.DrawPoint(s.rayStart, 5.0, box2d.ColorGreen)

	if s.bodyIDs[s.ignoreIndex] != (box2d.Body{}) {
		p := s.bodyIDs[s.ignoreIndex].Position()
		p.X -= 0.2
		s.draw.DrawString(s.camera, p, box2d.ColorWhite, "ign")
	}
}

const (
	overlapCircleShape = iota
	overlapCapsuleShape
	overlapBoxShape
)

const (
	overlapMaxCount  = 64
	overlapMaxDoomed = 16
)

type OverlapWorld struct {
	Sample

	bodyIndex   int
	bodyIDs     [overlapMaxCount]box2d.Body
	userData    [overlapMaxCount]ShapeUserData
	polygons    [4]box2d.Polygon
	capsule     box2d.Capsule
	circle      box2d.Circle
	segment     box2d.Segment
	ignoreIndex int

	doomIDs   [overlapMaxDoomed]box2d.Shape
	doomCount int

	shapeType int32

	startPosition box2d.Pos

	position  box2d.Pos
	angle     float32
	baseAngle float32

	dragging bool
	rotating bool
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

func CreateOverlapWorld(context *SampleContext) Sampler {
	s := &OverlapWorld{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 10.0}
		s.context.camera.zoom = 25.0 * 0.7
	}

	{
		vertices := []box2d.Vec2{{X: -0.5, Y: 0.0}, {X: 0.5, Y: 0.0}, {X: 0.0, Y: 1.5}}
		hull := box2d.ComputeHull(vertices)
		s.polygons[0] = box2d.MakePolygon(&hull, 0.0)
	}

	{
		vertices := []box2d.Vec2{{X: -0.1, Y: 0.0}, {X: 0.1, Y: 0.0}, {X: 0.0, Y: 1.5}}
		hull := box2d.ComputeHull(vertices)
		s.polygons[1] = box2d.MakePolygon(&hull, 0.0)
	}

	{
		w := float32(1.0)
		b := w / (2.0 + float32(math.Sqrt(2.0)))
		sq := float32(math.Sqrt(2.0)) * b

		vertices := []box2d.Vec2{{X: 0.5 * sq, Y: 0.0}, {X: 0.5 * w, Y: b}, {X: 0.5 * w, Y: b + sq}, {X: 0.5 * sq, Y: w},
			{X: -0.5 * sq, Y: w}, {X: -0.5 * w, Y: b + sq}, {X: -0.5 * w, Y: b}, {X: -0.5 * sq, Y: 0.0}}

		hull := box2d.ComputeHull(vertices)
		s.polygons[2] = box2d.MakePolygon(&hull, 0.0)
	}

	s.polygons[3] = box2d.MakeBox(0.5, 0.5)
	s.capsule = box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
	s.circle = box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	s.segment = box2d.Segment{Point1: box2d.Vec2{X: -1.0, Y: 0.0}, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}

	s.bodyIndex = 0

	s.ignoreIndex = 7

	s.shapeType = overlapCircleShape

	s.position = box2d.Pos{X: 0.0, Y: 10.0}
	s.angle = 0.0
	s.dragging = false
	s.rotating = false

	s.doomCount = 0

	s.CreateN(0, 10)
	return s
}

func (s *OverlapWorld) Create(index int) {
	if s.bodyIDs[s.bodyIndex] != (box2d.Body{}) {
		s.bodyIDs[s.bodyIndex].Destroy()
		s.bodyIDs[s.bodyIndex] = box2d.Body{}
	}

	x := scenes.RandomFloatRange(-20.0, 20.0)
	y := scenes.RandomFloatRange(0.0, 20.0)

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

	if index < 4 {
		s.bodyIDs[s.bodyIndex].CreatePolygonShape(&shapeDef, &s.polygons[index])
	} else if index == 4 {
		s.bodyIDs[s.bodyIndex].CreateCircleShape(&shapeDef, &s.circle)
	} else if index == 5 {
		s.bodyIDs[s.bodyIndex].CreateCapsuleShape(&shapeDef, &s.capsule)
	} else {
		s.bodyIDs[s.bodyIndex].CreateSegmentShape(&shapeDef, &s.segment)
	}

	s.bodyIndex = (s.bodyIndex + 1) % overlapMaxCount
}

func (s *OverlapWorld) CreateN(index, count int) {
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
	if imgui.Button("Polygon 1") {
		s.Create(0)
	}
	imgui.SameLine()
	if imgui.Button("10x##Poly1") {
		s.CreateN(0, 10)
	}

	if imgui.Button("Polygon 2") {
		s.Create(1)
	}
	imgui.SameLine()
	if imgui.Button("10x##Poly2") {
		s.CreateN(1, 10)
	}

	if imgui.Button("Polygon 3") {
		s.Create(2)
	}
	imgui.SameLine()
	if imgui.Button("10x##Poly3") {
		s.CreateN(2, 10)
	}

	if imgui.Button("Box") {
		s.Create(3)
	}
	imgui.SameLine()
	if imgui.Button("10x##Box") {
		s.CreateN(3, 10)
	}

	if imgui.Button("Circle") {
		s.Create(4)
	}
	imgui.SameLine()
	if imgui.Button("10x##Circle") {
		s.CreateN(4, 10)
	}

	if imgui.Button("Capsule") {
		s.Create(5)
	}
	imgui.SameLine()
	if imgui.Button("10x##Capsule") {
		s.CreateN(5, 10)
	}

	if imgui.Button("Segment") {
		s.Create(6)
	}
	imgui.SameLine()
	if imgui.Button("10x##Segment") {
		s.CreateN(6, 10)
	}

	if imgui.Button("Destroy Shape") {
		s.DestroyBody()
	}

	imgui.Separator()
	imText("Overlap Shape")
	imgui.RadioButtonIntPtr("Circle##Overlap", &s.shapeType, overlapCircleShape)
	imgui.RadioButtonIntPtr("Capsule##Overlap", &s.shapeType, overlapCapsuleShape)
	imgui.RadioButtonIntPtr("Box##Overlap", &s.shapeType, overlapBoxShape)

	return true
}

func (s *OverlapWorld) Step() {
	s.Sample.Step()

	s.DrawScreenTextLine("left mouse button: drag query shape")
	s.DrawScreenTextLine("left mouse button + shift: rotate query shape")

	s.doomCount = 0

	// Build the query shape near the origin and place it with the world-space query origin so it
	// stays precise far from the origin in large world mode.
	rotation := box2d.MakeRot(s.angle)
	var proxy box2d.ShapeProxy

	if s.shapeType == overlapCircleShape {
		circle := box2d.Circle{Center: box2d.Vec2Zero, Radius: 1.0}
		proxy = box2d.MakeProxy([]box2d.Vec2{circle.Center}, circle.Radius)
		s.draw.DrawSolidCircle(box2d.WorldTransform{P: s.position, Q: box2d.RotIdentity}, circle.Center, circle.Radius, box2d.ColorWhite)
	} else if s.shapeType == overlapCapsuleShape {
		capsule := box2d.Capsule{
			Center1: rotation.Rotate(box2d.Vec2{X: -1.0, Y: 0.0}),
			Center2: rotation.Rotate(box2d.Vec2{X: 1.0, Y: 0.0}),
			Radius:  0.5,
		}
		proxy = box2d.MakeProxy([]box2d.Vec2{capsule.Center1, capsule.Center2}, capsule.Radius)
		s.draw.DrawCapsule(s.position.Add(capsule.Center1), s.position.Add(capsule.Center2), capsule.Radius, box2d.ColorWhite)
	} else if s.shapeType == overlapBoxShape {
		box := box2d.MakeOffsetBox(2.0, 0.5, box2d.Vec2Zero, rotation)
		proxy = box2d.MakeProxy(box.Vertices[:box.Count], box.Radius)
		s.draw.DrawPolygon(box2d.WorldTransform{P: s.position, Q: box2d.RotIdentity}, box.Vertices[:box.Count], box2d.ColorWhite)
	}

	s.worldID.OverlapShape(s.position, &proxy, box2d.DefaultQueryFilter(), s.overlapResult)

	if s.bodyIDs[s.ignoreIndex] != (box2d.Body{}) {
		p := s.bodyIDs[s.ignoreIndex].Position()
		p.X -= 0.2
		s.draw.DrawString(s.camera, p, box2d.ColorWhite, "skip")
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

// Tests manifolds and contact points
type Manifold struct {
	Sample

	smgroxCache1 box2d.SimplexCache
	smgroxCache2 box2d.SimplexCache
	smgcapCache1 box2d.SimplexCache
	smgcapCache2 box2d.SimplexCache

	wedge box2d.Hull

	transform box2d.Transform
	angle     float32
	round     float32

	basePosition box2d.Vec2
	startPoint   box2d.Pos
	baseAngle    float32

	dragging       bool
	rotating       bool
	showCount      bool
	showIDs        bool
	showAnchors    bool
	showSeparation bool
	enableCaching  bool
}

func CreateManifold(context *SampleContext) Sampler {
	s := &Manifold{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 1.8, Y: 0.0}
		s.context.camera.zoom = 25.0 * 0.45
	}

	s.smgroxCache1 = box2d.SimplexCache{}
	s.smgroxCache2 = box2d.SimplexCache{}
	s.smgcapCache1 = box2d.SimplexCache{}
	s.smgcapCache2 = box2d.SimplexCache{}

	s.transform = box2d.TransformIdentity
	s.transform.P.X = 0.17
	s.transform.P.Y = 1.12
	s.angle = 0.0
	s.round = 0.1

	s.startPoint = box2d.Pos{X: 0.0, Y: 0.0}
	s.basePosition = box2d.Vec2{X: 0.0, Y: 0.0}
	s.baseAngle = 0.0

	s.dragging = false
	s.rotating = false
	s.showCount = false
	s.showIDs = false
	s.showSeparation = false
	s.showAnchors = false
	s.enableCaching = true

	points := []box2d.Vec2{{X: -0.1, Y: -0.5}, {X: 0.1, Y: -0.5}, {X: 0.0, Y: 0.5}}
	s.wedge = box2d.ComputeHull(points)
	return s
}

func (s *Manifold) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	imgui.SliderFloatV("x offset", &s.transform.P.X, -2.0, 2.0, "%.2f", 0)
	imgui.SliderFloatV("y offset", &s.transform.P.Y, -2.0, 2.0, "%.2f", 0)

	if imgui.SliderFloatV("angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.SliderFloatV("round", &s.round, 0.0, 0.4, "%.1f", 0)

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
	d := p.Sub(s.startPoint)

	if s.dragging {
		s.transform.P.X = s.basePosition.X + 0.5*d.X
		s.transform.P.Y = s.basePosition.Y + 0.5*d.Y
	} else if s.rotating {
		s.angle = clampFloat(s.baseAngle+1.0*d.X, -box2d.Pi, box2d.Pi)
		s.transform.Q = box2d.MakeRot(s.angle)
	} else {
		s.Sample.MouseMove(p)
	}
}

func (s *Manifold) DrawManifold(manifold *box2d.LocalManifold, transformA, transformB box2d.WorldTransform) {
	normal := transformA.Q.Rotate(manifold.Normal)

	if s.showCount {
		p := transformA.P.Lerp(transformB.P, 0.5)
		s.draw.DrawString(s.camera, p, box2d.ColorWhite, "%d", manifold.PointCount)
	}

	for i := range manifold.PointCount {
		mp := &manifold.Points[i]

		p1 := transformA.P.Add(transformA.Q.Rotate(mp.Point))
		p2 := p1.Add(normal.Scale(0.5))
		s.draw.DrawLine(p1, p2, box2d.ColorViolet)

		if s.showAnchors {
			s.draw.DrawPoint(p1, 5.0, box2d.ColorRed)
		} else {
			s.draw.DrawPoint(p1, 10.0, box2d.ColorBlue)
		}

		if s.showIDs {
			s.draw.DrawString(s.camera, p1.Add(box2d.Vec2{X: 0.05, Y: -0.02}), box2d.ColorWhite, "0x%04x", mp.ID)
		}

		if s.showSeparation {
			s.draw.DrawString(s.camera, p1.Add(box2d.Vec2{X: 0.05, Y: 0.03}), box2d.ColorWhite, "%.3f", mp.Separation)
		}
	}
}

func (s *Manifold) Step() {
	offset := box2d.Vec2{X: -10.0, Y: -5.0}
	increment := box2d.Vec2{X: 4.0, Y: 0.0}

	color1 := box2d.ColorAquamarine
	color2 := box2d.ColorPaleGoldenRod

	if !s.enableCaching {
		s.smgroxCache1 = box2d.SimplexCache{}
		s.smgroxCache2 = box2d.SimplexCache{}
		s.smgcapCache1 = box2d.SimplexCache{}
		s.smgcapCache2 = box2d.SimplexCache{}
	}

	transforms := func() (box2d.WorldTransform, box2d.WorldTransform) {
		transform1 := box2d.MakeWorldTransform(box2d.Transform{P: offset, Q: box2d.RotIdentity})
		transform2 := box2d.MakeWorldTransform(box2d.Transform{P: s.transform.P.Add(offset), Q: s.transform.Q})
		return transform1, transform2
	}

	// circle-circle
	{
		circle1 := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
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
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

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
		segment := box2d.Segment{Point1: box2d.Vec2{X: -1.0, Y: 0.0}, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

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
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		box := box2d.MakeSquare(0.5)
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
		capsule1 := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
		capsule2 := box2d.Capsule{Center1: box2d.Vec2{X: 0.25, Y: 0.0}, Center2: box2d.Vec2{X: 1.0, Y: 0.0}, Radius: 0.1}

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
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.4, Y: 0.0}, Center2: box2d.Vec2{X: -0.1, Y: 0.0}, Radius: 0.1}
		box := box2d.MakeOffsetBox(0.25, 1.0, box2d.Vec2{X: 1.0, Y: -1.0}, box2d.MakeRot(0.25*box2d.Pi))

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
		segment := box2d.Segment{Point1: box2d.Vec2{X: -1.0, Y: 0.0}, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}

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

	offset = box2d.Vec2{X: -10.0, Y: 0.0}

	// square-square
	{
		box1 := box2d.MakeSquare(0.5)
		box := box2d.MakeSquare(0.5)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&box1, &box, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box1.Vertices[:box1.Count], box1.Radius, color1)
		s.draw.DrawSolidPolygon(transform2, box.Vertices[:box.Count], box.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// box-box
	{
		box1 := box2d.MakeBox(2.0, 0.1)
		box := box2d.MakeSquare(0.25)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&box1, &box, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, box1.Vertices[:box1.Count], box1.Radius, color1)
		s.draw.DrawSolidPolygon(transform2, box.Vertices[:box.Count], box.Radius, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	// box-rox
	{
		box := box2d.MakeSquare(0.5)
		h := 0.5 - s.round
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
		h := 0.5 - s.round
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
		segment := box2d.Segment{Point1: box2d.Vec2{X: -1.0, Y: 0.0}, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}
		h := 0.5 - s.round
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
		p1s := []box2d.Vec2{{X: 0.175740838, Y: 0.224936664}, {X: -0.301293969, Y: 0.194021404}, {X: -0.105151534, Y: -0.432157338}}
		p2s := []box2d.Vec2{{X: -0.427884758, Y: -0.225028217}, {X: 0.0566576123, Y: -0.128772855}, {X: 0.176625848, Y: 0.338923335}}

		h1 := box2d.ComputeHull(p1s)
		h2 := box2d.ComputeHull(p2s)
		w1 := box2d.MakePolygon(&h1, 0.158798501)
		w2 := box2d.MakePolygon(&h2, 0.205900759)

		transform1, transform2 := transforms()

		m := box2d.CollidePolygons(&w1, &w2, transform1.InvMul(transform2))

		s.draw.DrawSolidPolygon(transform1, w1.Vertices[:w1.Count], w1.Radius, color1)
		s.draw.DrawSolidPolygon(transform1, w1.Vertices[:w1.Count], 0.0, color1)
		s.draw.DrawSolidPolygon(transform2, w2.Vertices[:w2.Count], w2.Radius, color2)
		s.draw.DrawSolidPolygon(transform2, w2.Vertices[:w2.Count], 0.0, color2)

		s.DrawManifold(&m, transform1, transform2)

		offset = offset.Add(increment)
	}

	offset = box2d.Vec2{X: -10.0, Y: 5.0}

	// box-triangle
	{
		box := box2d.MakeBox(1.0, 1.0)
		points := []box2d.Vec2{{X: -0.05, Y: 0.0}, {X: 0.05, Y: 0.0}, {X: 0.0, Y: 0.1}}
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
		segment := box2d.ChainSegment{Ghost1: box2d.Vec2{X: 2.0, Y: 1.0}, Segment: box2d.Segment{Point1: box2d.Vec2{X: 1.0, Y: 1.0}, Point2: box2d.Vec2{X: -1.0, Y: 0.0}},
			Ghost2: box2d.Vec2{X: -2.0, Y: 0.0}, ChainID: -1}
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}

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

		offset.X += 2.0 * increment.X
	}

	segment1 := box2d.ChainSegment{Ghost1: box2d.Vec2{X: 2.0, Y: 1.0}, Segment: box2d.Segment{Point1: box2d.Vec2{X: 1.0, Y: 1.0}, Point2: box2d.Vec2{X: -1.0, Y: 0.0}},
		Ghost2: box2d.Vec2{X: -2.0, Y: 0.0}, ChainID: -1}
	segment2 := box2d.ChainSegment{Ghost1: box2d.Vec2{X: 3.0, Y: 1.0}, Segment: box2d.Segment{Point1: box2d.Vec2{X: 2.0, Y: 1.0}, Point2: box2d.Vec2{X: 1.0, Y: 1.0}},
		Ghost2: box2d.Vec2{X: -1.0, Y: 0.0}, ChainID: -1}

	drawChainSegments := func(transform1 box2d.WorldTransform) {
		{
			g2 := transform1.Apply(segment1.Ghost2)
			p1 := transform1.Apply(segment1.Segment.Point1)
			p2 := transform1.Apply(segment1.Segment.Point2)
			s.draw.DrawLine(p1, p2, color1)
			s.draw.DrawPoint(p1, 4.0, color1)
			s.draw.DrawPoint(p2, 4.0, color1)
			s.draw.DrawLine(p2, g2, box2d.ColorLightGray)
		}

		{
			g1 := transform1.Apply(segment2.Ghost1)
			p1 := transform1.Apply(segment2.Segment.Point1)
			p2 := transform1.Apply(segment2.Segment.Point2)
			s.draw.DrawLine(g1, p1, box2d.ColorLightGray)
			s.draw.DrawLine(p1, p2, color1)
			s.draw.DrawPoint(p1, 4.0, color1)
			s.draw.DrawPoint(p2, 4.0, color1)
		}
	}

	// chain-segment vs rounded polygon
	{
		h := 0.5 - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)

		transform1, transform2 := transforms()

		xf := transform1.InvMul(transform2)
		m1 := box2d.CollideChainSegmentAndPolygon(&segment1, &rox, xf, &s.smgroxCache1)
		m2 := box2d.CollideChainSegmentAndPolygon(&segment2, &rox, xf, &s.smgroxCache2)

		drawChainSegments(transform1)

		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)
		s.draw.DrawPoint(transform2.Apply(rox.Centroid), 5.0, box2d.ColorGainsboro)

		s.DrawManifold(&m1, transform1, transform2)
		s.DrawManifold(&m2, transform1, transform2)

		offset.X += 2.0 * increment.X
	}

	// chain-segment vs capsule
	{
		capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}

		transform1, transform2 := transforms()

		xf := transform1.InvMul(transform2)
		m1 := box2d.CollideChainSegmentAndCapsule(&segment1, &capsule, xf, &s.smgcapCache1)
		m2 := box2d.CollideChainSegmentAndCapsule(&segment2, &capsule, xf, &s.smgcapCache2)

		drawChainSegments(transform1)

		p1 := transform2.Apply(capsule.Center1)
		p2 := transform2.Apply(capsule.Center2)
		s.draw.DrawCapsule(p1, p2, capsule.Radius, color2)

		s.draw.DrawPoint(p1.Lerp(p2, 0.5), 5.0, box2d.ColorGainsboro)

		s.DrawManifold(&m1, transform1, transform2)
		s.DrawManifold(&m2, transform1, transform2)

		offset.X += 2.0 * increment.X
	}

	s.DrawScreenTextLine("mouse button 1: drag")
	s.DrawScreenTextLine("mouse button 1 + shift: rotate")
}

const (
	smoothCircleShape = iota
	smoothBoxShape
)

type SmoothManifold struct {
	Sample

	shapeType int32

	segments []box2d.ChainSegment

	transform box2d.Transform
	angle     float32
	round     float32

	basePosition box2d.Vec2
	startPoint   box2d.Pos
	baseAngle    float32

	dragging       bool
	rotating       bool
	showIDs        bool
	showAnchors    bool
	showSeparation bool
}

func CreateSmoothManifold(context *SampleContext) Sampler {
	s := &SmoothManifold{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 2.0, Y: 20.0}
		s.context.camera.zoom = 21.0
	}

	s.shapeType = smoothBoxShape
	s.transform = box2d.Transform{P: box2d.Vec2{X: 0.0, Y: 20.0}, Q: box2d.RotIdentity}
	s.angle = 0.0
	s.round = 0.0

	s.startPoint = box2d.Pos{X: 0.0, Y: 0.0}
	s.basePosition = box2d.Vec2{X: 0.0, Y: 0.0}
	s.baseAngle = 0.0

	s.dragging = false
	s.rotating = false
	s.showIDs = false
	s.showAnchors = false
	s.showSeparation = false

	// https://betravis.github.io/shape-tools/path-to-polygon/
	points := []box2d.Vec2{
		{X: -20.58325, Y: 14.54175},
		{X: -21.90625, Y: 15.8645},
		{X: -24.552, Y: 17.1875},
		{X: -27.198, Y: 11.89575},
		{X: -29.84375, Y: 15.8645},
		{X: -29.84375, Y: 21.15625},
		{X: -25.875, Y: 23.802},
		{X: -20.58325, Y: 25.125},
		{X: -25.875, Y: 29.09375},
		{X: -20.58325, Y: 31.7395},
		{X: -11.0089998, Y: 23.2290001},
		{X: -8.67700005, Y: 21.15625},
		{X: -6.03125, Y: 21.15625},
		{X: -7.35424995, Y: 29.09375},
		{X: -3.38549995, Y: 29.09375},
		{X: 1.90625, Y: 30.41675},
		{X: 5.875, Y: 17.1875},
		{X: 11.16675, Y: 25.125},
		{X: 9.84375, Y: 29.09375},
		{X: 13.8125, Y: 31.7395},
		{X: 21.75, Y: 30.41675},
		{X: 28.3644981, Y: 26.448},
		{X: 25.71875, Y: 18.5105},
		{X: 24.3957481, Y: 13.21875},
		{X: 17.78125, Y: 11.89575},
		{X: 15.1355, Y: 7.92700005},
		{X: 5.875, Y: 9.25},
		{X: 1.90625, Y: 11.89575},
		{X: -3.25, Y: 11.89575},
		{X: -3.25, Y: 9.9375},
		{X: -4.70825005, Y: 9.25},
		{X: -8.67700005, Y: 9.25},
		{X: -11.323, Y: 11.89575},
		{X: -13.96875, Y: 11.89575},
		{X: -15.29175, Y: 14.54175},
		{X: -19.2605, Y: 14.54175},
	}

	count := len(points)
	s.segments = make([]box2d.ChainSegment, count)

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

		s.segments[i] = box2d.ChainSegment{Ghost1: g1, Segment: box2d.Segment{Point1: p1, Point2: p2}, Ghost2: g2, ChainID: -1}
	}
	return s
}

func (s *SmoothManifold) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	{
		shapeTypes := []string{"Circle", "Box"}
		imgui.ComboStrarr("Shape", &s.shapeType, shapeTypes, int32(len(shapeTypes)))
	}

	imgui.SliderFloatV("x Offset", &s.transform.P.X, -2.0, 2.0, "%.2f", 0)
	imgui.SliderFloatV("y Offset", &s.transform.P.Y, -2.0, 2.0, "%.2f", 0)

	if imgui.SliderFloatV("Angle", &s.angle, -box2d.Pi, box2d.Pi, "%.2f", 0) {
		s.transform.Q = box2d.MakeRot(s.angle)
	}

	imgui.SliderFloatV("Round", &s.round, 0.0, 0.4, "%.1f", 0)

	imgui.PopItemWidth()

	imgui.Checkbox("Show Ids", &s.showIDs)
	imgui.Checkbox("Show Separation", &s.showSeparation)
	imgui.Checkbox("Show Anchors", &s.showAnchors)

	if imgui.Button("Reset") {
		s.transform = box2d.TransformIdentity
		s.angle = 0.0
	}

	imgui.Separator()

	imText("mouse button 1: drag")
	imText("mouse button 1 + shift: rotate")

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
	d := p.Sub(s.startPoint)

	if s.dragging {
		s.transform.P.X = s.basePosition.X + d.X
		s.transform.P.Y = s.basePosition.Y + d.Y
	} else if s.rotating {
		s.angle = clampFloat(s.baseAngle+1.0*d.X, -box2d.Pi, box2d.Pi)
		s.transform.Q = box2d.MakeRot(s.angle)
	}
}

func (s *SmoothManifold) DrawManifold(manifold *box2d.LocalManifold, transformA box2d.WorldTransform) {
	normal := transformA.Q.Rotate(manifold.Normal)

	for i := range manifold.PointCount {
		mp := &manifold.Points[i]

		p1 := transformA.P.Add(transformA.Q.Rotate(mp.Point))
		p2 := p1.Add(normal.Scale(0.5))
		s.draw.DrawLine(p1, p2, box2d.ColorWhite)

		if s.showAnchors {
			s.draw.DrawPoint(p1, 5.0, box2d.ColorGreen)
		} else {
			s.draw.DrawPoint(p1, 5.0, box2d.ColorGreen)
		}

		if s.showIDs {
			s.draw.DrawString(s.camera, p1.Add(box2d.Vec2{X: 0.05, Y: -0.02}), box2d.ColorWhite, "0x%04x", mp.ID)
		}

		if s.showSeparation {
			s.draw.DrawString(s.camera, p1.Add(box2d.Vec2{X: 0.05, Y: 0.03}), box2d.ColorWhite, "%.3f", mp.Separation)
		}
	}
}

func (s *SmoothManifold) Step() {
	color1 := box2d.ColorYellow
	color2 := box2d.ColorMagenta

	transform1 := box2d.WorldTransformIdentity
	transform2 := box2d.MakeWorldTransform(s.transform)

	for i := range s.segments {
		segment := &s.segments[i]
		p1 := transform1.Apply(segment.Segment.Point1)
		p2 := transform1.Apply(segment.Segment.Point2)
		s.draw.DrawLine(p1, p2, color1)
		s.draw.DrawPoint(p1, 4.0, color1)
	}

	// chain-segment vs circle
	if s.shapeType == smoothCircleShape {
		radius := float32(0.5)
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: radius}
		s.draw.DrawSolidCircle(transform2, circle.Center, circle.Radius, color2)

		for i := range s.segments {
			segment := &s.segments[i]
			m := box2d.CollideChainSegmentAndCircle(segment, &circle, transform1.InvMul(transform2))
			s.DrawManifold(&m, transform1)
		}
	} else if s.shapeType == smoothBoxShape {
		h := 0.5 - s.round
		rox := box2d.MakeRoundedBox(h, h, s.round)
		s.draw.DrawSolidPolygon(transform2, rox.Vertices[:rox.Count], rox.Radius, color2)

		for i := range s.segments {
			segment := &s.segments[i]
			var cache box2d.SimplexCache
			m := box2d.CollideChainSegmentAndPolygon(segment, &rox, transform1.InvMul(transform2), &cache)
			s.DrawManifold(&m, transform1)
		}
	}
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

	basePosition box2d.Vec2
	startPoint   box2d.Pos
	baseAngle    float32

	dragging    bool
	sweeping    bool
	rotating    bool
	showIndices bool
	drawSimplex bool
	encroach    bool
}

func CreateShapeCast(context *SampleContext) Sampler {
	s := &ShapeCast{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 0.25}
		s.context.camera.zoom = 3.0
	}

	s.point = box2d.Vec2Zero
	s.segment = box2d.Segment{Point1: box2d.Vec2{X: 0.0, Y: 0.0}, Point2: box2d.Vec2{X: 0.5, Y: 0.0}}

	{
		points := []box2d.Vec2{{X: -0.5, Y: 0.0}, {X: 0.5, Y: 0.0}, {X: 0.0, Y: 1.0}}
		hull := box2d.ComputeHull(points)
		s.triangle = box2d.MakePolygon(&hull, 0.0)
	}

	s.box = box2d.MakeBox(8.984375, 0.5)

	s.transform = box2d.Transform{P: box2d.Vec2{X: -0.6, Y: 0.0}, Q: box2d.RotIdentity}
	s.translation = box2d.Vec2{X: 2.0, Y: 0.0}
	s.angle = 0.0
	s.startPoint = box2d.Pos{X: 0.0, Y: 0.0}
	s.basePosition = box2d.Vec2{X: 0.0, Y: 0.0}
	s.baseAngle = 0.0

	s.dragging = false
	s.sweeping = false
	s.rotating = false
	s.showIndices = false
	s.drawSimplex = false
	s.encroach = false

	s.typeA = shapeBox
	s.typeB = shapePoint
	s.radiusA = 0.0
	s.radiusB = 0.2

	s.proxyA = s.makeProxy(s.typeA, s.radiusA)
	s.proxyB = s.makeProxy(s.typeB, s.radiusB)
	return s
}

func (s *ShapeCast) MouseDown(p box2d.Pos, button glfw.MouseButton, mods glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		if mods == 0 {
			s.dragging = true
			s.sweeping = false
			s.rotating = false
			s.startPoint = p
			s.basePosition = s.transform.P
		} else if mods == glfw.ModShift {
			s.dragging = false
			s.sweeping = false
			s.rotating = true
			s.startPoint = p
			s.baseAngle = s.angle
		} else if mods == glfw.ModControl {
			s.dragging = false
			s.sweeping = true
			s.rotating = false
			s.startPoint = p
			s.basePosition = box2d.Vec2Zero
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
	d := p.Sub(s.startPoint)

	if s.dragging {
		s.transform.P = s.basePosition.MulAdd(0.5, d)
	} else if s.rotating {
		s.angle = clampFloat(s.baseAngle+1.0*d.X, -box2d.Pi, box2d.Pi)
		s.transform.Q = box2d.MakeRot(s.angle)
	} else if s.sweeping {
		s.translation = d
	}
}

func (s *ShapeCast) DrawControls() bool {
	imgui.PushItemWidth(6.0 * imgui.FontSize())

	shapeType := int32(s.typeA)
	if imgui.ComboStrarr("shape A", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeA = proxyShapeType(shapeType)
		s.proxyA = s.makeProxy(s.typeA, s.radiusA)
	}

	if imgui.SliderFloatV("radius A", &s.radiusA, 0.0, 0.5, "%.2f", 0) {
		s.proxyA.Radius = s.radiusA
	}

	shapeType = int32(s.typeB)
	if imgui.ComboStrarr("shape B", &shapeType, proxyShapeTypes, int32(len(proxyShapeTypes))) {
		s.typeB = proxyShapeType(shapeType)
		s.proxyB = s.makeProxy(s.typeB, s.radiusB)
	}

	if imgui.SliderFloatV("radius B", &s.radiusB, 0.0, 0.5, "%.2f", 0) {
		s.proxyB.Radius = s.radiusB
	}

	imgui.Separator()

	imgui.SliderFloatV("x offset", &s.transform.P.X, -2.0, 2.0, "%.2f", 0)
	imgui.SliderFloatV("y offset", &s.transform.P.Y, -2.0, 2.0, "%.2f", 0)

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

	s.DrawScreenTextLine("hit = %t, iterations = %d, fraction = %.6g, distance = %.6g", output.Hit,
		output.Iterations, output.Fraction, distanceOutput.Distance)

	s.drawShape(s.draw, s.typeA, box2d.WorldTransformIdentity, s.radiusA, box2d.ColorCyan)
	s.drawShape(s.draw, s.typeB, box2d.MakeWorldTransform(s.transform), s.radiusB, box2d.ColorLightGreen)
	transform2 := box2d.Transform{P: s.transform.P.Add(s.translation), Q: s.transform.Q}
	s.drawShape(s.draw, s.typeB, box2d.MakeWorldTransform(transform2), s.radiusB, box2d.ColorIndianRed)

	if output.Hit {
		s.drawShape(s.draw, s.typeB, box2d.MakeWorldTransform(transform), s.radiusB, box2d.ColorPlum)

		if output.Fraction > 0.0 {
			s.draw.DrawPoint(output.Point.Pos(), 5.0, box2d.ColorWhite)
			s.draw.DrawLine(output.Point.Pos(), output.Point.MulAdd(0.5, output.Normal).Pos(), box2d.ColorYellow)
		} else {
			s.draw.DrawPoint(output.Point.Pos(), 5.0, box2d.ColorPeru)
		}
	}

	if s.showIndices {
		for i := range s.proxyA.Count {
			s.draw.DrawString(s.camera, s.proxyA.Points[i].Pos(), box2d.ColorWhite, " %d", i)
		}

		transformB := box2d.MakeWorldTransform(s.transform)
		for i := range s.proxyB.Count {
			p := transformB.Apply(s.proxyB.Points[i])
			s.draw.DrawString(s.camera, p, box2d.ColorWhite, " %d", i)
		}
	}

	s.DrawScreenTextLine("mouse button 1: drag")
	s.DrawScreenTextLine("mouse button 1 + shift: rotate")
	s.DrawScreenTextLine("mouse button 1 + control: sweep")
	s.DrawScreenTextLine("distance = %.2f, iterations = %d", distanceOutput.Distance, output.Iterations)
}

type TimeOfImpact struct {
	Sample

	verticesA [4]box2d.Vec2
	verticesB [2]box2d.Vec2

	radiusA float32
	radiusB float32
}

func CreateTimeOfImpact(context *SampleContext) Sampler {
	s := &TimeOfImpact{
		verticesA: [4]box2d.Vec2{{X: -16.25, Y: 44.75}, {X: -15.75, Y: 44.75}, {X: -15.75, Y: 45.25}, {X: -16.25, Y: 45.25}},
		verticesB: [2]box2d.Vec2{{X: 0.0, Y: -0.125000000}, {X: 0.0, Y: 0.125000000}},
		radiusA:   0.0,
		radiusB:   0.0299999993,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.6, Y: 2.0}
		s.context.camera.center = box2d.Pos{X: -16, Y: 45}
		s.context.camera.zoom = 5.0
	}
	return s
}

func (s *TimeOfImpact) Step() {
	s.Sample.Step()

	sweepA := box2d.Sweep{
		LocalCenter: box2d.Vec2Zero, C1: box2d.Vec2{X: 0.0, Y: 0.0}, C2: box2d.Vec2{X: 0.0, Y: 0.0}, Q1: box2d.RotIdentity, Q2: box2d.RotIdentity,
	}
	sweepB := box2d.Sweep{
		LocalCenter: box2d.Vec2Zero,
		C1:          box2d.Vec2{X: -15.8332710, Y: 45.3520279},
		C2:          box2d.Vec2{X: -15.8324337, Y: 45.3413048},
		Q1:          box2d.Rot{C: -0.540891349, S: 0.841092527},
		Q2:          box2d.Rot{C: -0.457797021, S: 0.889056742},
	}

	var input box2d.TOIInput
	input.ProxyA = box2d.MakeProxy(s.verticesA[:], s.radiusA)
	input.ProxyB = box2d.MakeProxy(s.verticesB[:], s.radiusB)
	input.SweepA = sweepA
	input.SweepB = sweepB
	input.MaxFraction = 1.0

	output := box2d.TimeOfImpact(&input)

	s.DrawScreenTextLine("toi = %.6g", output.Fraction)

	// Draw A
	transformA := box2d.MakeWorldTransform(sweepA.Transform(0.0))
	s.draw.DrawPolygon(transformA, s.verticesA[:], box2d.ColorGray)

	// Draw B at t = 0
	transformB := box2d.MakeWorldTransform(sweepB.Transform(0.0))
	s.draw.DrawCapsule(transformB.Apply(s.verticesB[0]), transformB.Apply(s.verticesB[1]), s.radiusB, box2d.ColorGreen)

	// Draw B at t = hit_time
	transformB = box2d.MakeWorldTransform(sweepB.Transform(output.Fraction))
	s.draw.DrawPolygon(transformB, s.verticesB[:], box2d.ColorOrange)

	// Draw B at t = 1
	transformB = box2d.MakeWorldTransform(sweepB.Transform(1.0))
	s.draw.DrawCapsule(transformB.Apply(s.verticesB[0]), transformB.Apply(s.verticesB[1]), s.radiusB, box2d.ColorRed)

	if output.State == box2d.TOIStateHit {
		var distanceInput box2d.DistanceInput
		distanceInput.ProxyA = input.ProxyA
		distanceInput.ProxyB = input.ProxyB
		distanceInput.Transform = sweepA.Transform(output.Fraction).InvMul(sweepB.Transform(output.Fraction))
		distanceInput.UseRadii = false
		var cache box2d.SimplexCache
		distanceOutput := box2d.ShapeDistance(&distanceInput, &cache, nil)
		s.DrawScreenTextLine("distance = %.6g", distanceOutput.Distance)
	}
}

func init() {
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
