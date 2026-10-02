// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "sync/atomic"

type solverBlock struct {
	startIndex int
	count      uint16
	blockType  uint8
	colorIndex uint8
}

type syncBlock struct {
	block     solverBlock
	syncIndex atomic.Int32
}

type solverStage struct {
	blocks          []syncBlock
	stageType       solverStageType
	blockCount      int
	colorIndex      uint8
	completionCount atomic.Int32
}

type softness struct {
	biasRate     float32
	massScale    float32
	impulseScale float32
}

type contactPrepareSpan struct {
	start    int
	count    int
	contacts []contactSim
}

type contactCollideSpan struct {
	start    int
	contacts []contactSim
}

type jointPrepareSpan struct {
	start  int
	count  int
	joints []jointSim
}

type stepContext struct {
	dt                     float32
	invDt                  float32
	h                      float32
	invH                   float32
	subStepCount           int
	contactSoftness        softness
	staticSoftness         softness
	restitutionThreshold   float32
	maxLinearVelocity      float32
	world                  *world
	graph                  *constraintGraph
	states                 []bodyState
	sims                   []bodySim
	enlargedShapes         []int
	enlargedShapeCount     int
	bulletBodies           []int
	bulletBodyCount        atomic.Int32
	anyRestitution         atomic.Int32
	collideSpans           []contactCollideSpan
	wideContactConstraints []contactConstraintWide
	contactPrepareSpans    []contactPrepareSpan
	wideContactCount       int
	jointPrepareSpans      []jointPrepareSpan
	jointCount             int
	activeColorCount       int
	workerCount            int
	stages                 []solverStage
	stageCount             int
	enableWarmStarting     bool
	padding1               [64]byte
	atomicSyncBits         atomic.Uint32
	padding2               [64]byte
	mainClaimed            atomic.Int32
	padding3               [64]byte
}

type workerContext struct {
	context     *stepContext
	workerIndex int
	userTask    func()
}

type continuousContext struct {
	world           *world
	fastBodySim     *bodySim
	fastShape       *shape
	centroid1       Vec2
	centroid2       Vec2
	sweep           Sweep
	base            Pos
	fraction        float32
	sensorHits      [maxContinuousSensorHits]sensorHit
	sensorFractions [maxContinuousSensorHits]float32
	sensorCount     int
}

type blockDim struct {
	size  int
	count int
}

type solverStageType int

const (
	stagePrepareJoints       solverStageType = 0
	stagePrepareContacts     solverStageType = 1
	stageIntegrateVelocities solverStageType = 2
	stageWarmStart           solverStageType = 3
	stageSolve               solverStageType = 4
	stageIntegratePositions  solverStageType = 5
	stageRelax               solverStageType = 6
	stageRestitution         solverStageType = 7
	stageStoreImpulses       solverStageType = 8
)

type solverBlockType int

const (
	bodyBlock         solverBlockType = 0
	jointBlock        solverBlockType = 1
	contactBlock      solverBlockType = 2
	graphJointBlock   solverBlockType = 3
	graphContactBlock solverBlockType = 4
)

const (
	simdShift                       = 2
	iterations                      = 1
	relaxIterations                 = 1
	maxContinuousSensorHits         = 8
	coreFraction            float32 = 0.25
)

func pause() {
	todo()
}

func integrateVelocitiesTask(block solverBlock, context *stepContext) {
	todo()
}

func integratePositionsTask(block solverBlock, context *stepContext) {
	todo()
}

func continuousQueryCallback(proxyID int, userData uint64, context *continuousContext) bool {
	todo()
	return false
}

func solveContinuous(world *world, bodySimIndex int, taskContext *taskContext, dt float32) {
	todo()
}

func finalizeBodiesTask(startIndex int, endIndex int, workerIndex int, context *stepContext) {
	todo()
}

func computeBlockCount(itemCount int, minSize int, maxBlockCount int) blockDim {
	todo()
	return blockDim{}
}

func initBlocks(blocks []syncBlock, dim blockDim, itemCount int, blockType uint8, colorIndex uint8) {
	todo()
}

func initStage(stage []solverStage, stageType solverStageType, blocks []syncBlock, blockCount int, colorIndex uint8) []solverStage {
	todo()
	return nil
}

func initColorStages(stage []solverStage, stageType solverStageType, iterations int, activeColorCount int, colorBlocks [][]syncBlock, colorBlockCounts []int, activeColorIndices []int) []solverStage {
	todo()
	return nil
}

func executeBlock(stage *solverStage, context *stepContext, block solverBlock, workerIndex int) {
	todo()
}

func getWorkerStartIndex(workerIndex int, blockCount int, workerCount int) int {
	todo()
	return 0
}

func executeStage(stage *solverStage, context *stepContext, previousSyncIndex int, syncIndex int, workerIndex int) {
	todo()
}

func executeMainStage(stage *solverStage, context *stepContext, syncBits uint32) {
	todo()
}

func solverTask(taskContext *workerContext) {
	todo()
}

func bulletBodyTask(startIndex int, endIndex int, workerIndex int, context *stepContext) {
	todo()
}

func solve(world *world, stepContext *stepContext) {
	todo()
}

func makeSoft(hertz float32, zeta float32, h float32) softness {
	todo()
	return softness{}
}
