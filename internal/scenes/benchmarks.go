// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import "github.com/zigai/box2d"

const (
	benchmarkDebug              = false
	rainRowCount                = 5
	rainColumnCount             = 40
	rainGroupSize               = 5
	spinnerPointCount           = 360
	sleepPyramidCount           = 10
	queryCount                  = 1000
	treeCastRayCount            = 1000
	treeCastLengthCount         = 3
	treeCastBoxExtent   float32 = 0.1
	tilePeriod          float32 = 40
	tileCycleCount              = 600
	tileGridSize        float32 = 1
	tileColumnsPerBody          = 10
	tileQueryCount              = 500
	tileRayLength       float32 = 25
)

type group struct {
	humans [rainGroupSize]Human
}

type rainData struct {
	groups      [rainRowCount * rainColumnCount]group
	gridSize    float32
	gridCount   int
	columnCount int
	columnIndex int
}

type spinnerData struct {
	spinnerID box2d.Joint
}

type junkyardData struct {
	pusherID box2d.Body
}

type sleepData struct {
	bodyIDA [sleepPyramidCount]box2d.Body
	bodyIDB [sleepPyramidCount]box2d.Body
}

type queryBenchmark struct {
	origins      [1000]box2d.Pos
	translations [1000]box2d.Vec2
	paddleID     box2d.Body
	stats        box2d.TreeStats
	extent       float32
	paddleMinX   float32
	paddleMaxX   float32
}

type treeCastBenchmark struct {
	tree          box2d.DynamicTree
	tightBoxes    []box2d.AABB
	proxyCount    int
	proxyCapacity int
	origins       [1000]box2d.Vec2
	translations  [treeCastLengthCount][1000]box2d.Vec2
	stats         box2d.TreeStats
	created       bool
}

type treeCastContext struct {
	tightBoxes []box2d.AABB
	fraction   float32
}

type tileWorldBenchmark struct {
	groundIDs   []box2d.Body
	groundCount int
	groundHead  int
	bodySpan    float32
	xLeft       float32
	xNext       float32
	stats       box2d.TreeStats
}

var gRainData rainData
var gSpinnerData spinnerData
var gJunkyardData junkyardData
var gSleepData sleepData
var gQueryBenchmark queryBenchmark
var gQueryRandomState uint32
var gTreeCast treeCastBenchmark
var gTileWorld tileWorldBenchmark

func CreateJointGrid(worldID box2d.World) {
	todo()
}

func CreateLargePyramid(worldID box2d.World) {
	todo()
}

func createSmallPyramid(worldID box2d.World, baseCount int, extent, centerX, baseY float32) {
	todo()
}

func CreateManyPyramids(worldID box2d.World) {
	todo()
}

func GetManyPyramidsCapacity() box2d.Capacity {
	todo()
	return box2d.Capacity{}
}

func CreateRain(worldID box2d.World) {
	todo()
}

func CreateGroup(worldID box2d.World, rowIndex, columnIndex int) {
	todo()
}

func DestroyGroup(rowIndex, columnIndex int) {
	todo()
}

func StepRain(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func CreateSpinner(worldID box2d.World) {
	todo()
}

func StepSpinner(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func CreateSmash(worldID box2d.World) {
	todo()
}

func CreateTumbler(worldID box2d.World) {
	todo()
}

func CreateWasher(worldID box2d.World) {
	todo()
}

func CreateJunkyard(worldID box2d.World) {
	todo()
}

func StepJunkyard(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func createSleepPyramid(worldID box2d.World, baseCount int, extent, centerX float32, bodyIDA, bodyIDB *box2d.Body) {
	todo()
}

func CreateSleep(worldID box2d.World) {
	todo()
}

func StepSleep(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func CreateCompounds(worldID box2d.World) {
	todo()
}

func queryRandom(lower, upper float32) float32 {
	todo()
	return 0
}

func CreateQueries(worldID box2d.World) {
	todo()
}

func queryCastClosest(shapeID box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32, result *float32) float32 {
	todo()
	return 0
}

func queryOverlapCount(shapeID box2d.Shape, count *int) bool {
	todo()
	return false
}

func StepQueries(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func GetQueryBenchmarkStats() box2d.TreeStats {
	todo()
	return box2d.TreeStats{}
}

func GetQueryBenchmarkCount() int {
	todo()
	return 0
}

func GetQueryBenchmarkExtent() float32 {
	todo()
	return 0
}

func GetQueryBenchmarkRay(index int) (box2d.Pos, box2d.Vec2) {
	todo()
	return box2d.Pos{}, box2d.Vec2{}
}

func treeCastRayBox(box box2d.AABB, p, d box2d.Vec2, maxFraction float32) float32 {
	todo()
	return 0
}

func treeCastRayCallback(input *box2d.RayCastInput, proxyID int, userData uint64, context *treeCastContext) float32 {
	todo()
	return 0
}

func treeCastBoxCallback(input *box2d.BoxCastInput, proxyID int, userData uint64, context *treeCastContext) float32 {
	todo()
	return 0
}

func treeCastAddBox(center box2d.Vec2, hx, hy float32, categoryBits uint64) {
	todo()
}

func DestroyTreeCast() {
	todo()
}

func CreateTreeCast(worldID box2d.World) {
	todo()
}

func StepTreeCast(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func GetTreeCastBenchmarkStats() box2d.TreeStats {
	todo()
	return box2d.TreeStats{}
}

func tileRandom(state *uint32, lower, upper float32) float32 {
	todo()
	return 0
}

func createTileGround(worldID box2d.World, xBody float32) box2d.Body {
	todo()
	return box2d.Body{}
}

func DestroyTileWorld() {
	todo()
}

func CreateTileWorld(worldID box2d.World) {
	todo()
}

func tileOverlapCount(shapeID box2d.Shape, count *int) bool {
	todo()
	return false
}

func StepTileWorld(worldID box2d.World, stepCount int) float32 {
	todo()
	return 0
}

func GetTileWorldBenchmarkStats() box2d.TreeStats {
	todo()
	return box2d.TreeStats{}
}
