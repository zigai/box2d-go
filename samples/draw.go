// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	_ "embed"
	"fmt"
	"math"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d"
)

var (
	//go:embed data/background.vs
	backgroundVS string
	//go:embed data/background.fs
	backgroundFS string
	//go:embed data/circle.vs
	circleVS string
	//go:embed data/circle.fs
	circleFS string
	//go:embed data/line.vs
	lineVS string
	//go:embed data/line.fs
	lineFS string
	//go:embed data/point.vs
	pointVS string
	//go:embed data/point.fs
	pointFS string
	//go:embed data/solid_capsule.vs
	solidCapsuleVS string
	//go:embed data/solid_capsule.fs
	solidCapsuleFS string
	//go:embed data/solid_circle.vs
	solidCircleVS string
	//go:embed data/solid_circle.fs
	solidCircleFS string
	//go:embed data/solid_polygon.vs
	solidPolygonVS string
	//go:embed data/solid_polygon.fs
	solidPolygonFS string
)

type Camera struct {
	center box2d.Pos
	zoom   float32
	width  float32
	height float32
}

type rgba8 struct {
	r, g, b, a uint8
}

func makeRGBA8(c box2d.HexColor, alpha float32) rgba8 {
	return rgba8{
		uint8((c >> 16) & 0xFF),
		uint8((c >> 8) & 0xFF),
		uint8(c & 0xFF),
		uint8(0xFF * alpha),
	}
}

func GetDefaultCamera() Camera {
	return Camera{
		center: box2d.Pos{X: 0.0, Y: 20.0},
		zoom:   1.0,
		width:  1920.0,
		height: 1080.0,
	}
}

func ResetView(camera *Camera) {
	camera.center = box2d.Pos{X: 0.0, Y: 20.0}
	camera.zoom = 1.0
}

func ConvertScreenToWorld(camera *Camera, screenPoint box2d.Vec2) box2d.Pos {
	w := camera.width
	h := camera.height
	u := screenPoint.X / w
	v := (h - screenPoint.Y) / h

	ratio := w / h
	extents := box2d.Vec2{X: camera.zoom * ratio, Y: camera.zoom}

	offset := box2d.Vec2{X: extents.X * (2.0*u - 1.0), Y: extents.Y * (2.0*v - 1.0)}
	return camera.center.Add(offset)
}

func ConvertViewToScreen(camera *Camera, viewPoint box2d.Vec2) box2d.Vec2 {
	w := camera.width
	h := camera.height
	ratio := w / h

	extents := box2d.Vec2{X: camera.zoom * ratio, Y: camera.zoom}

	u := (viewPoint.X + extents.X) / (2.0 * extents.X)
	v := (viewPoint.Y + extents.Y) / (2.0 * extents.Y)

	return box2d.Vec2{X: u * w, Y: (1.0 - v) * h}
}

func ConvertWorldToScreen(camera *Camera, worldPoint box2d.Pos) box2d.Vec2 {
	return ConvertViewToScreen(camera, worldPoint.Sub(camera.center))
}

func buildProjectionMatrix(camera *Camera, m *[16]float32, zBias float32) {
	ratio := camera.width / camera.height
	extents := box2d.Vec2{X: camera.zoom * ratio, Y: camera.zoom}

	w := 2.0 * extents.X
	h := 2.0 * extents.Y

	m[0] = 2.0 / w
	m[1] = 0.0
	m[2] = 0.0
	m[3] = 0.0

	m[4] = 0.0
	m[5] = 2.0 / h
	m[6] = 0.0
	m[7] = 0.0

	m[8] = 0.0
	m[9] = 0.0
	m[10] = -1.0
	m[11] = 0.0

	m[12] = 0.0
	m[13] = 0.0
	m[14] = zBias
	m[15] = 1.0
}

func GetViewBounds(camera *Camera) box2d.AABB {
	if camera.height == 0.0 || camera.width == 0.0 {
		return box2d.AABB{}
	}

	lower := ConvertScreenToWorld(camera, box2d.Vec2{X: 0.0, Y: camera.height})
	upper := ConvertScreenToWorld(camera, box2d.Vec2{X: camera.width, Y: 0.0})

	var bounds box2d.AABB
	bounds.LowerBound = box2d.Vec2{X: box2d.RoundDownFloat(float64(lower.X)), Y: box2d.RoundDownFloat(float64(lower.Y))}
	bounds.UpperBound = box2d.Vec2{X: box2d.RoundUpFloat(float64(upper.X)), Y: box2d.RoundUpFloat(float64(upper.Y))}
	return bounds
}

func FocusOnBounds(camera *Camera, bounds box2d.AABB) {
	if camera.width == 0 {
		return
	}

	extents := bounds.Extents()

	if extents.X < box2d.LinearSlop() || extents.Y < box2d.LinearSlop() {
		return
	}

	invRatio := camera.height / camera.width
	camera.zoom = max(extents.X*invRatio, extents.Y)

	camera.zoom = max(camera.zoom, 0.01)

	camera.center = bounds.Center().Pos()
}

type background struct {
	vaoID             uint32
	vboID             uint32
	programID         uint32
	timeUniform       int32
	resolutionUniform int32
	baseColorUniform  int32
}

func createBackground() background {
	var b background

	b.programID = createProgramFromStrings(backgroundVS, backgroundFS)
	b.timeUniform = gl.GetUniformLocation(b.programID, gl.Str("time\x00"))
	b.resolutionUniform = gl.GetUniformLocation(b.programID, gl.Str("resolution\x00"))
	b.baseColorUniform = gl.GetUniformLocation(b.programID, gl.Str("baseColor\x00"))
	var vertexAttribute uint32 = 0

	gl.GenVertexArrays(1, &b.vaoID)
	gl.GenBuffers(1, &b.vboID)

	gl.BindVertexArray(b.vaoID)
	gl.EnableVertexAttribArray(vertexAttribute)

	vertices := []box2d.Vec2{{X: -1.0, Y: 1.0}, {X: -1.0, Y: -1.0}, {X: 1.0, Y: 1.0}, {X: 1.0, Y: -1.0}}
	gl.BindBuffer(gl.ARRAY_BUFFER, b.vboID)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(vertices[0])), unsafe.Pointer(&vertices[0]), gl.STATIC_DRAW)
	gl.VertexAttribPointerWithOffset(vertexAttribute, 2, gl.FLOAT, false, 0, 0)

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return b
}

func (b *background) destroy() {
	if b.vaoID != 0 {
		gl.DeleteVertexArrays(1, &b.vaoID)
		gl.DeleteBuffers(1, &b.vboID)
		b.vaoID = 0
		b.vboID = 0
	}

	if b.programID != 0 {
		gl.DeleteProgram(b.programID)
		b.programID = 0
	}
}

func (b *background) render(camera *Camera) {
	gl.UseProgram(b.programID)

	time := float32(glfw.GetTime())
	time = float32(math.Mod(float64(time), 100.0))

	gl.Uniform1f(b.timeUniform, time)
	gl.Uniform2f(b.resolutionUniform, camera.width, camera.height)

	gl.Uniform3f(b.baseColorUniform, 0.2, 0.2, 0.2)

	gl.BindVertexArray(b.vaoID)

	gl.BindBuffer(gl.ARRAY_BUFFER, b.vboID)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)
}

const pointBatchSize = 2048

type pointData struct {
	position box2d.Vec2
	size     float32
	rgba     rgba8
}

type pointRender struct {
	points            []pointData
	vaoID             uint32
	vboID             uint32
	programID         uint32
	projectionUniform int32
}

func createPointDrawData() pointRender {
	var render pointRender
	render.points = make([]pointData, 0, pointBatchSize)
	render.programID = createProgramFromStrings(pointVS, pointFS)
	render.projectionUniform = gl.GetUniformLocation(render.programID, gl.Str("projectionMatrix\x00"))
	var vertexAttribute uint32 = 0
	var sizeAttribute uint32 = 1
	var colorAttribute uint32 = 2

	gl.GenVertexArrays(1, &render.vaoID)
	gl.GenBuffers(1, &render.vboID)

	gl.BindVertexArray(render.vaoID)
	gl.EnableVertexAttribArray(vertexAttribute)
	gl.EnableVertexAttribArray(sizeAttribute)
	gl.EnableVertexAttribArray(colorAttribute)

	var data pointData
	stride := int32(unsafe.Sizeof(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboID)
	gl.BufferData(gl.ARRAY_BUFFER, pointBatchSize*int(stride), nil, gl.DYNAMIC_DRAW)

	gl.VertexAttribPointerWithOffset(vertexAttribute, 2, gl.FLOAT, false, stride, unsafe.Offsetof(data.position))
	gl.VertexAttribPointerWithOffset(sizeAttribute, 1, gl.FLOAT, false, stride, unsafe.Offsetof(data.size))
	gl.VertexAttribPointerWithOffset(colorAttribute, 4, gl.UNSIGNED_BYTE, true, stride, unsafe.Offsetof(data.rgba))

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return render
}

func (render *pointRender) destroy() {
	if render.vaoID != 0 {
		gl.DeleteVertexArrays(1, &render.vaoID)
		gl.DeleteBuffers(1, &render.vboID)
	}

	if render.programID != 0 {
		gl.DeleteProgram(render.programID)
	}

	*render = pointRender{}
}

func (render *pointRender) add(v box2d.Vec2, size float32, c box2d.HexColor) {
	render.points = append(render.points, pointData{v, size, makeRGBA8(c, 1.0)})
}

func (render *pointRender) flush(camera *Camera) {
	count := len(render.points)
	if count == 0 {
		return
	}

	gl.UseProgram(render.programID)

	var proj [16]float32
	buildProjectionMatrix(camera, &proj, 0.0)

	gl.UniformMatrix4fv(render.projectionUniform, 1, false, &proj[0])
	gl.BindVertexArray(render.vaoID)

	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboID)
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	stride := int(unsafe.Sizeof(render.points[0]))
	base := 0
	for count > 0 {
		batchCount := min(count, pointBatchSize)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, batchCount*stride, unsafe.Pointer(&render.points[base]))
		gl.DrawArrays(gl.POINTS, 0, int32(batchCount))

		checkOpenGL()

		count -= pointBatchSize
		base += pointBatchSize
	}

	gl.Disable(gl.PROGRAM_POINT_SIZE)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	render.points = render.points[:0]
}

const lineBatchSize = 2 * 2048

type vertexData struct {
	position box2d.Vec2
	rgba     rgba8
}

type lineRender struct {
	points            []vertexData
	vaoID             uint32
	vboID             uint32
	programID         uint32
	projectionUniform int32
}

func createLineRender() lineRender {
	var render lineRender
	render.points = make([]vertexData, 0, lineBatchSize)
	render.programID = createProgramFromStrings(lineVS, lineFS)
	render.projectionUniform = gl.GetUniformLocation(render.programID, gl.Str("projectionMatrix\x00"))
	var vertexAttribute uint32 = 0
	var colorAttribute uint32 = 1

	gl.GenVertexArrays(1, &render.vaoID)
	gl.GenBuffers(1, &render.vboID)

	gl.BindVertexArray(render.vaoID)
	gl.EnableVertexAttribArray(vertexAttribute)
	gl.EnableVertexAttribArray(colorAttribute)

	var data vertexData
	stride := int32(unsafe.Sizeof(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboID)
	gl.BufferData(gl.ARRAY_BUFFER, lineBatchSize*int(stride), nil, gl.DYNAMIC_DRAW)

	gl.VertexAttribPointerWithOffset(vertexAttribute, 2, gl.FLOAT, false, stride, unsafe.Offsetof(data.position))
	gl.VertexAttribPointerWithOffset(colorAttribute, 4, gl.UNSIGNED_BYTE, true, stride, unsafe.Offsetof(data.rgba))

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return render
}

func (render *lineRender) destroy() {
	if render.vaoID != 0 {
		gl.DeleteVertexArrays(1, &render.vaoID)
		gl.DeleteBuffers(1, &render.vboID)
	}

	if render.programID != 0 {
		gl.DeleteProgram(render.programID)
	}

	*render = lineRender{}
}

func (render *lineRender) add(p1, p2 box2d.Vec2, c box2d.HexColor) {
	rgba := makeRGBA8(c, 1.0)
	render.points = append(render.points, vertexData{p1, rgba}, vertexData{p2, rgba})
}

func (render *lineRender) flush(camera *Camera) {
	count := len(render.points)
	if count == 0 {
		return
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.UseProgram(render.programID)

	var proj [16]float32
	buildProjectionMatrix(camera, &proj, 0.1)

	gl.UniformMatrix4fv(render.projectionUniform, 1, false, &proj[0])

	gl.BindVertexArray(render.vaoID)

	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboID)

	stride := int(unsafe.Sizeof(render.points[0]))
	base := 0
	for count > 0 {
		batchCount := min(count, lineBatchSize)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, batchCount*stride, unsafe.Pointer(&render.points[base]))

		gl.DrawArrays(gl.LINES, 0, int32(batchCount))

		checkOpenGL()

		count -= lineBatchSize
		base += lineBatchSize
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	gl.Disable(gl.BLEND)

	render.points = render.points[:0]
}

var quadVertices = [6]box2d.Vec2{{X: -1.1, Y: -1.1}, {X: 1.1, Y: -1.1}, {X: -1.1, Y: 1.1}, {X: 1.1, Y: -1.1}, {X: 1.1, Y: 1.1}, {X: -1.1, Y: 1.1}}

func bufferQuad(vboID uint32, vertexAttribute uint32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, vboID)
	gl.BufferData(gl.ARRAY_BUFFER, int(unsafe.Sizeof(quadVertices)), unsafe.Pointer(&quadVertices[0]), gl.STATIC_DRAW)
	gl.VertexAttribPointerWithOffset(vertexAttribute, 2, gl.FLOAT, false, 0, 0)
}

const circleBatchSize = 2048

type circleData struct {
	position box2d.Vec2
	radius   float32
	rgba     rgba8
}

type circleRender struct {
	circles           []circleData
	vaoID             uint32
	vboIDs            [2]uint32
	programID         uint32
	projectionUniform int32
	pixelScaleUniform int32
}

func createCircles() circleRender {
	var render circleRender
	render.circles = make([]circleData, 0, circleBatchSize)
	render.programID = createProgramFromStrings(circleVS, circleFS)
	render.projectionUniform = gl.GetUniformLocation(render.programID, gl.Str("projectionMatrix\x00"))
	render.pixelScaleUniform = gl.GetUniformLocation(render.programID, gl.Str("pixelScale\x00"))
	var vertexAttribute uint32 = 0
	var positionInstance uint32 = 1
	var radiusInstance uint32 = 2
	var colorInstance uint32 = 3

	gl.GenVertexArrays(1, &render.vaoID)
	gl.GenBuffers(2, &render.vboIDs[0])

	gl.BindVertexArray(render.vaoID)
	gl.EnableVertexAttribArray(vertexAttribute)
	gl.EnableVertexAttribArray(positionInstance)
	gl.EnableVertexAttribArray(radiusInstance)
	gl.EnableVertexAttribArray(colorInstance)

	bufferQuad(render.vboIDs[0], vertexAttribute)

	var data circleData
	stride := int32(unsafe.Sizeof(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.BufferData(gl.ARRAY_BUFFER, circleBatchSize*int(stride), nil, gl.DYNAMIC_DRAW)

	gl.VertexAttribPointerWithOffset(positionInstance, 2, gl.FLOAT, false, stride, unsafe.Offsetof(data.position))
	gl.VertexAttribPointerWithOffset(radiusInstance, 1, gl.FLOAT, false, stride, unsafe.Offsetof(data.radius))
	gl.VertexAttribPointerWithOffset(colorInstance, 4, gl.UNSIGNED_BYTE, true, stride, unsafe.Offsetof(data.rgba))

	gl.VertexAttribDivisor(positionInstance, 1)
	gl.VertexAttribDivisor(radiusInstance, 1)
	gl.VertexAttribDivisor(colorInstance, 1)

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return render
}

func (render *circleRender) destroy() {
	if render.vaoID != 0 {
		gl.DeleteVertexArrays(1, &render.vaoID)
		gl.DeleteBuffers(2, &render.vboIDs[0])
	}

	if render.programID != 0 {
		gl.DeleteProgram(render.programID)
	}

	*render = circleRender{}
}

func (render *circleRender) add(center box2d.Vec2, radius float32, color box2d.HexColor) {
	render.circles = append(render.circles, circleData{center, radius, makeRGBA8(color, 1.0)})
}

func (render *circleRender) flush(camera *Camera) {
	count := len(render.circles)
	if count == 0 {
		return
	}

	gl.UseProgram(render.programID)

	var proj [16]float32
	buildProjectionMatrix(camera, &proj, 0.2)

	gl.UniformMatrix4fv(render.projectionUniform, 1, false, &proj[0])
	gl.Uniform1f(render.pixelScaleUniform, camera.height/camera.zoom)

	gl.BindVertexArray(render.vaoID)

	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	stride := int(unsafe.Sizeof(render.circles[0]))
	base := 0
	for count > 0 {
		batchCount := min(count, circleBatchSize)

		gl.BufferSubData(gl.ARRAY_BUFFER, 0, batchCount*stride, unsafe.Pointer(&render.circles[base]))
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, int32(batchCount))

		checkOpenGL()

		count -= circleBatchSize
		base += circleBatchSize
	}

	gl.Disable(gl.BLEND)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	render.circles = render.circles[:0]
}

type solidCircle struct {
	transform box2d.Transform
	radius    float32
	rgba      rgba8
}

const solidCircleBatchSize = 2048

type solidCircles struct {
	circles           []solidCircle
	vaoID             uint32
	vboIDs            [2]uint32
	programID         uint32
	projectionUniform int32
	pixelScaleUniform int32
}

func createSolidCircles() solidCircles {
	var render solidCircles
	render.circles = make([]solidCircle, 0, solidCircleBatchSize)
	render.programID = createProgramFromStrings(solidCircleVS, solidCircleFS)
	render.projectionUniform = gl.GetUniformLocation(render.programID, gl.Str("projectionMatrix\x00"))
	render.pixelScaleUniform = gl.GetUniformLocation(render.programID, gl.Str("pixelScale\x00"))

	gl.GenVertexArrays(1, &render.vaoID)
	gl.GenBuffers(2, &render.vboIDs[0])

	gl.BindVertexArray(render.vaoID)

	var vertexAttribute uint32 = 0
	var transformInstance uint32 = 1
	var radiusInstance uint32 = 2
	var colorInstance uint32 = 3
	gl.EnableVertexAttribArray(vertexAttribute)
	gl.EnableVertexAttribArray(transformInstance)
	gl.EnableVertexAttribArray(radiusInstance)
	gl.EnableVertexAttribArray(colorInstance)

	bufferQuad(render.vboIDs[0], vertexAttribute)

	var data solidCircle
	stride := int32(unsafe.Sizeof(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.BufferData(gl.ARRAY_BUFFER, solidCircleBatchSize*int(stride), nil, gl.DYNAMIC_DRAW)

	gl.VertexAttribPointerWithOffset(transformInstance, 4, gl.FLOAT, false, stride, unsafe.Offsetof(data.transform))
	gl.VertexAttribPointerWithOffset(radiusInstance, 1, gl.FLOAT, false, stride, unsafe.Offsetof(data.radius))
	gl.VertexAttribPointerWithOffset(colorInstance, 4, gl.UNSIGNED_BYTE, true, stride, unsafe.Offsetof(data.rgba))

	gl.VertexAttribDivisor(transformInstance, 1)
	gl.VertexAttribDivisor(radiusInstance, 1)
	gl.VertexAttribDivisor(colorInstance, 1)

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return render
}

func (render *solidCircles) destroy() {
	if render.vaoID != 0 {
		gl.DeleteVertexArrays(1, &render.vaoID)
		gl.DeleteBuffers(2, &render.vboIDs[0])
	}

	if render.programID != 0 {
		gl.DeleteProgram(render.programID)
	}

	*render = solidCircles{}
}

func (render *solidCircles) add(transform box2d.Transform, radius float32, color box2d.HexColor) {
	render.circles = append(render.circles, solidCircle{transform, radius, makeRGBA8(color, 1.0)})
}

func (render *solidCircles) flush(camera *Camera) {
	count := len(render.circles)
	if count == 0 {
		return
	}

	gl.UseProgram(render.programID)

	var proj [16]float32
	buildProjectionMatrix(camera, &proj, 0.2)

	gl.UniformMatrix4fv(render.projectionUniform, 1, false, &proj[0])
	gl.Uniform1f(render.pixelScaleUniform, camera.height/camera.zoom)

	gl.BindVertexArray(render.vaoID)

	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	stride := int(unsafe.Sizeof(render.circles[0]))
	base := 0
	for count > 0 {
		batchCount := min(count, solidCircleBatchSize)

		gl.BufferSubData(gl.ARRAY_BUFFER, 0, batchCount*stride, unsafe.Pointer(&render.circles[base]))
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, int32(batchCount))

		checkOpenGL()

		count -= solidCircleBatchSize
		base += solidCircleBatchSize
	}

	gl.Disable(gl.BLEND)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	render.circles = render.circles[:0]
}

type capsuleData struct {
	transform box2d.Transform
	radius    float32
	length    float32
	rgba      rgba8
}

const capsuleBatchSize = 2048

type capsules struct {
	capsules          []capsuleData
	vaoID             uint32
	vboIDs            [2]uint32
	programID         uint32
	projectionUniform int32
	pixelScaleUniform int32
}

func createCapsules() capsules {
	var render capsules
	render.capsules = make([]capsuleData, 0, capsuleBatchSize)
	render.programID = createProgramFromStrings(solidCapsuleVS, solidCapsuleFS)
	render.projectionUniform = gl.GetUniformLocation(render.programID, gl.Str("projectionMatrix\x00"))
	render.pixelScaleUniform = gl.GetUniformLocation(render.programID, gl.Str("pixelScale\x00"))

	var vertexAttribute uint32 = 0
	var transformInstance uint32 = 1
	var radiusInstance uint32 = 2
	var lengthInstance uint32 = 3
	var colorInstance uint32 = 4

	gl.GenVertexArrays(1, &render.vaoID)
	gl.GenBuffers(2, &render.vboIDs[0])

	gl.BindVertexArray(render.vaoID)
	gl.EnableVertexAttribArray(vertexAttribute)
	gl.EnableVertexAttribArray(transformInstance)
	gl.EnableVertexAttribArray(radiusInstance)
	gl.EnableVertexAttribArray(lengthInstance)
	gl.EnableVertexAttribArray(colorInstance)

	bufferQuad(render.vboIDs[0], vertexAttribute)

	var data capsuleData
	stride := int32(unsafe.Sizeof(data))
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.BufferData(gl.ARRAY_BUFFER, capsuleBatchSize*int(stride), nil, gl.DYNAMIC_DRAW)

	gl.VertexAttribPointerWithOffset(transformInstance, 4, gl.FLOAT, false, stride, unsafe.Offsetof(data.transform))
	gl.VertexAttribPointerWithOffset(radiusInstance, 1, gl.FLOAT, false, stride, unsafe.Offsetof(data.radius))
	gl.VertexAttribPointerWithOffset(lengthInstance, 1, gl.FLOAT, false, stride, unsafe.Offsetof(data.length))
	gl.VertexAttribPointerWithOffset(colorInstance, 4, gl.UNSIGNED_BYTE, true, stride, unsafe.Offsetof(data.rgba))

	gl.VertexAttribDivisor(transformInstance, 1)
	gl.VertexAttribDivisor(radiusInstance, 1)
	gl.VertexAttribDivisor(lengthInstance, 1)
	gl.VertexAttribDivisor(colorInstance, 1)

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return render
}

func (render *capsules) destroy() {
	if render.vaoID != 0 {
		gl.DeleteVertexArrays(1, &render.vaoID)
		gl.DeleteBuffers(2, &render.vboIDs[0])
	}

	if render.programID != 0 {
		gl.DeleteProgram(render.programID)
	}

	*render = capsules{}
}

func (render *capsules) add(p1, p2 box2d.Vec2, radius float32, c box2d.HexColor) {
	d := p2.Sub(p1)
	length := d.Length()
	if length < 0.001 {
		fmt.Printf("WARNING: sample app: capsule too short!\n")
		return
	}

	axis := box2d.Vec2{X: d.X / length, Y: d.Y / length}
	var transform box2d.Transform
	transform.P = p1.Lerp(p2, 0.5)
	transform.Q.C = axis.X
	transform.Q.S = axis.Y

	render.capsules = append(render.capsules, capsuleData{transform, radius, length, makeRGBA8(c, 1.0)})
}

func (render *capsules) flush(camera *Camera) {
	count := len(render.capsules)
	if count == 0 {
		return
	}

	gl.UseProgram(render.programID)

	var proj [16]float32
	buildProjectionMatrix(camera, &proj, 0.2)

	gl.UniformMatrix4fv(render.projectionUniform, 1, false, &proj[0])
	gl.Uniform1f(render.pixelScaleUniform, camera.height/camera.zoom)

	gl.BindVertexArray(render.vaoID)

	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	stride := int(unsafe.Sizeof(render.capsules[0]))
	base := 0
	for count > 0 {
		batchCount := min(count, capsuleBatchSize)

		gl.BufferSubData(gl.ARRAY_BUFFER, 0, batchCount*stride, unsafe.Pointer(&render.capsules[base]))
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, int32(batchCount))

		checkOpenGL()

		count -= capsuleBatchSize
		base += capsuleBatchSize
	}

	gl.Disable(gl.BLEND)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	render.capsules = render.capsules[:0]
}

type polygonData struct {
	transform box2d.Transform
	points    [8]box2d.Vec2
	count     int32
	radius    float32
	color     rgba8
}

const polygonBatchSize = 2048

type polygons struct {
	polygons          []polygonData
	vaoID             uint32
	vboIDs            [2]uint32
	programID         uint32
	projectionUniform int32
	pixelScaleUniform int32
}

func createPolygons() polygons {
	var render polygons
	render.polygons = make([]polygonData, 0, 10*polygonBatchSize)
	render.programID = createProgramFromStrings(solidPolygonVS, solidPolygonFS)
	render.projectionUniform = gl.GetUniformLocation(render.programID, gl.Str("projectionMatrix\x00"))
	render.pixelScaleUniform = gl.GetUniformLocation(render.programID, gl.Str("pixelScale\x00"))

	var vertexAttribute uint32 = 0
	var instanceTransform uint32 = 1
	var instancePoint12 uint32 = 2
	var instancePoint34 uint32 = 3
	var instancePoint56 uint32 = 4
	var instancePoint78 uint32 = 5
	var instancePointCount uint32 = 6
	var instanceRadius uint32 = 7
	var instanceColor uint32 = 8

	gl.GenVertexArrays(1, &render.vaoID)
	gl.GenBuffers(2, &render.vboIDs[0])

	gl.BindVertexArray(render.vaoID)
	gl.EnableVertexAttribArray(vertexAttribute)
	gl.EnableVertexAttribArray(instanceTransform)
	gl.EnableVertexAttribArray(instancePoint12)
	gl.EnableVertexAttribArray(instancePoint34)
	gl.EnableVertexAttribArray(instancePoint56)
	gl.EnableVertexAttribArray(instancePoint78)
	gl.EnableVertexAttribArray(instancePointCount)
	gl.EnableVertexAttribArray(instanceRadius)
	gl.EnableVertexAttribArray(instanceColor)

	bufferQuad(render.vboIDs[0], vertexAttribute)

	var data polygonData
	stride := int32(unsafe.Sizeof(data))
	pointOffset := unsafe.Offsetof(data.points)
	pointSize := unsafe.Sizeof(data.points[0])
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])
	gl.BufferData(gl.ARRAY_BUFFER, polygonBatchSize*int(stride), nil, gl.DYNAMIC_DRAW)
	gl.VertexAttribPointerWithOffset(instanceTransform, 4, gl.FLOAT, false, stride, unsafe.Offsetof(data.transform))
	gl.VertexAttribPointerWithOffset(instancePoint12, 4, gl.FLOAT, false, stride, pointOffset)
	gl.VertexAttribPointerWithOffset(instancePoint34, 4, gl.FLOAT, false, stride, pointOffset+2*pointSize)
	gl.VertexAttribPointerWithOffset(instancePoint56, 4, gl.FLOAT, false, stride, pointOffset+4*pointSize)
	gl.VertexAttribPointerWithOffset(instancePoint78, 4, gl.FLOAT, false, stride, pointOffset+6*pointSize)
	gl.VertexAttribIPointerWithOffset(instancePointCount, 1, gl.INT, stride, unsafe.Offsetof(data.count))
	gl.VertexAttribPointerWithOffset(instanceRadius, 1, gl.FLOAT, false, stride, unsafe.Offsetof(data.radius))
	gl.VertexAttribPointerWithOffset(instanceColor, 4, gl.UNSIGNED_BYTE, true, stride, unsafe.Offsetof(data.color))

	gl.VertexAttribDivisor(instanceTransform, 1)
	gl.VertexAttribDivisor(instancePoint12, 1)
	gl.VertexAttribDivisor(instancePoint34, 1)
	gl.VertexAttribDivisor(instancePoint56, 1)
	gl.VertexAttribDivisor(instancePoint78, 1)
	gl.VertexAttribDivisor(instancePointCount, 1)
	gl.VertexAttribDivisor(instanceRadius, 1)
	gl.VertexAttribDivisor(instanceColor, 1)

	checkOpenGL()

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	return render
}

func (render *polygons) destroy() {
	if render.vaoID != 0 {
		gl.DeleteVertexArrays(1, &render.vaoID)
		gl.DeleteBuffers(2, &render.vboIDs[0])
	}

	if render.programID != 0 {
		gl.DeleteProgram(render.programID)
	}

	*render = polygons{}
}

func (render *polygons) add(transform box2d.Transform, points []box2d.Vec2, radius float32, color box2d.HexColor) {
	var data polygonData
	data.transform = transform

	n := copy(data.points[:], points)

	data.count = int32(n)
	data.radius = radius
	data.color = makeRGBA8(color, 1.0)

	render.polygons = append(render.polygons, data)
}

func (render *polygons) flush(camera *Camera) {
	count := len(render.polygons)
	if count == 0 {
		return
	}

	gl.UseProgram(render.programID)

	var proj [16]float32
	buildProjectionMatrix(camera, &proj, 0.2)

	gl.UniformMatrix4fv(render.projectionUniform, 1, false, &proj[0])
	gl.Uniform1f(render.pixelScaleUniform, camera.height/camera.zoom)

	gl.BindVertexArray(render.vaoID)
	gl.BindBuffer(gl.ARRAY_BUFFER, render.vboIDs[1])

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	stride := int(unsafe.Sizeof(render.polygons[0]))
	base := 0
	for count > 0 {
		batchCount := min(count, polygonBatchSize)

		gl.BufferSubData(gl.ARRAY_BUFFER, 0, batchCount*stride, unsafe.Pointer(&render.polygons[base]))
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, int32(batchCount))
		checkOpenGL()

		count -= polygonBatchSize
		base += polygonBatchSize
	}

	gl.Disable(gl.BLEND)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)

	render.polygons = render.polygons[:0]
}

type Draw struct {
	background    background
	points        pointRender
	lines         lineRender
	hollowCircles circleRender
	circles       solidCircles
	capsules      capsules
	polygons      polygons

	origin box2d.Pos
}

func CreateDraw() *Draw {
	return &Draw{
		background:    createBackground(),
		points:        createPointDrawData(),
		lines:         createLineRender(),
		hollowCircles: createCircles(),
		circles:       createSolidCircles(),
		capsules:      createCapsules(),
		polygons:      createPolygons(),
	}
}

func (draw *Draw) Destroy() {
	draw.background.destroy()
	draw.points.destroy()
	draw.lines.destroy()
	draw.hollowCircles.destroy()
	draw.circles.destroy()
	draw.capsules.destroy()
	draw.polygons.destroy()
}

func (draw *Draw) SetOrigin(origin box2d.Pos) {
	draw.origin = origin
}

func (draw *Draw) DrawPoint(p box2d.Pos, size float32, color box2d.HexColor) {
	draw.points.add(p.Sub(draw.origin), size, color)
}

func (draw *Draw) DrawLine(p1, p2 box2d.Pos, color box2d.HexColor) {
	draw.lines.add(p1.Sub(draw.origin), p2.Sub(draw.origin), color)
}

func (draw *Draw) DrawCircle(center box2d.Pos, radius float32, color box2d.HexColor) {
	draw.hollowCircles.add(center.Sub(draw.origin), radius, color)
}

func (draw *Draw) DrawCapsule(p1, p2 box2d.Pos, radius float32, color box2d.HexColor) {
	draw.capsules.add(p1.Sub(draw.origin), p2.Sub(draw.origin), radius, color)
}

func (draw *Draw) DrawPolygon(transform box2d.WorldTransform, vertices []box2d.Vec2, color box2d.HexColor) {
	xf := transform.Relative(draw.origin)
	p1 := xf.Apply(vertices[len(vertices)-1])
	for _, v := range vertices {
		p2 := xf.Apply(v)
		draw.lines.add(p1, p2, color)
		p1 = p2
	}
}

func (draw *Draw) DrawSolidCircle(transform box2d.WorldTransform, center box2d.Vec2, radius float32, color box2d.HexColor) {
	xf := box2d.WorldTransform{P: transform.Apply(center), Q: transform.Q}
	draw.circles.add(xf.Relative(draw.origin), radius, color)
}

func (draw *Draw) DrawSolidPolygon(transform box2d.WorldTransform, vertices []box2d.Vec2, radius float32, color box2d.HexColor) {
	draw.polygons.add(transform.Relative(draw.origin), vertices, radius, color)
}

func (draw *Draw) DrawTransform(transform box2d.WorldTransform, scale float32) {
	xf := transform.Relative(draw.origin)

	p1 := xf.P

	p2 := p1.MulAdd(scale, xf.Q.XAxis())
	draw.lines.add(p1, p2, box2d.ColorRed)

	p2 = p1.MulAdd(scale, xf.Q.YAxis())
	draw.lines.add(p1, p2, box2d.ColorGreen)
}

func (draw *Draw) DrawBounds(aabb box2d.AABB, color box2d.HexColor) {
	lower := aabb.LowerBound.Pos().Sub(draw.origin)
	upper := aabb.UpperBound.Pos().Sub(draw.origin)

	p1 := lower
	p2 := box2d.Vec2{X: upper.X, Y: lower.Y}
	p3 := upper
	p4 := box2d.Vec2{X: lower.X, Y: upper.Y}

	draw.lines.add(p1, p2, color)
	draw.lines.add(p2, p3, color)
	draw.lines.add(p3, p4, color)
	draw.lines.add(p4, p1, color)
}

func (draw *Draw) Flush(camera *Camera) {
	draw.circles.flush(camera)
	draw.capsules.flush(camera)
	draw.polygons.flush(camera)
	draw.hollowCircles.flush(camera)
	draw.lines.flush(camera)
	draw.points.flush(camera)
	checkOpenGL()
}

func (draw *Draw) DrawBackground(camera *Camera) {
	draw.background.render(camera)
}
