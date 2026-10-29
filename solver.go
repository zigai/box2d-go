// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"math/bits"
	"runtime"
	"slices"
	"sync/atomic"

	"github.com/zigai/box2d/internal/array"
	"github.com/zigai/box2d/internal/bitset"
	"github.com/zigai/box2d/internal/core"
	"github.com/zigai/box2d/internal/timer"
)

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
	world                  *World
	graph                  *constraintGraph
	states                 []bodyState
	sims                   []bodySim
	bulletBodies           []int
	bulletBodyCount        atomic.Int32
	anyRestitution         atomic.Int32
	collideSpans           []contactCollideSpan
	wideContactConstraints []contactConstraintWide
	contactPrepareSpans    []contactPrepareSpan
	wideContactCount       int
	jointPrepareSpans      []jointPrepareSpan
	activeColorCount       int
	workerCount            int
	stages                 []solverStage
	stageCount             int
	enableWarmStarting     bool
	_                      [64]byte
	atomicSyncBits         atomic.Uint32
	_                      [64]byte
	mainClaimed            atomic.Int32
	_                      [64]byte
}

type workerContext struct {
	context     *stepContext
	workerIndex int
	userTask    func()
}

type continuousContext struct {
	world           *World
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

type solverBlockType int

func pause() { runtime.Gosched() }

func integrateVelocitiesTask(block solverBlock, context *stepContext) {
	states := context.states
	sims := context.sims

	core.Validate(block.startIndex+int(block.count) <= len(context.world.solverSets[awakeSet].bodyStates), "block.startIndex + block.count <= context->world->solverSets.data[b2_awakeSet].bodyStates.count")

	gravity := context.world.gravity
	h := context.h

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		sim := &sims[i]
		state := &states[i]

		v := state.linearVelocity
		w := state.angularVelocity

		// Apply forces, torque, gravity, and damping
		// Apply damping.
		// Differential equation: dv/dt + c * v = 0
		// Solution: v(t) = v0 * exp(-c * t)
		// Time step: v(t + dt) = v0 * exp(-c * (t + dt)) = v0 * exp(-c * t) * exp(-c * dt) = v(t) * exp(-c * dt)
		// v2 = exp(-c * dt) * v1
		// Pade approximation:
		// v2 = v1 * 1 / (1 + c * dt)
		linearDamping := 1.0 / (1.0 + float32(h*sim.linearDamping))
		angularDamping := 1.0 / (1.0 + float32(h*sim.angularDamping))

		// Gravity scale will be zero for kinematic bodies
		var gravityScale float32
		if sim.invMass > 0 {
			gravityScale = sim.gravityScale
		}

		// lvd = h * im * f + h * g
		linearVelocityDelta := sim.force.Scale(h * sim.invMass).Add(gravity.Scale(h * gravityScale))
		angularVelocityDelta := float32(h * sim.invInertia * sim.torque)

		v = linearVelocityDelta.MulAdd(linearDamping, v)
		w = angularVelocityDelta + float32(angularDamping*w)

		state.linearVelocity = v
		state.angularVelocity = w
	}
}

func integratePositionsTask(block solverBlock, context *stepContext) {
	core.Validate(block.startIndex+int(block.count) <= len(context.world.solverSets[awakeSet].bodyStates), "block.startIndex + block.count <= context->world->solverSets.data[b2_awakeSet].bodyStates.count")

	states := context.states
	h := context.h
	maxLinearSpeed := context.maxLinearVelocity
	maxAngularSpeed := MaxRotation * context.invDt
	maxAngularSpeedSquared := maxAngularSpeed * maxAngularSpeed

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		state := &states[i]

		v := state.linearVelocity
		w := state.angularVelocity

		// Motion locks - these can be viewed as a constraint that comes last
		if state.flags&lockLinearX != 0 {
			v.X = 0
		}

		if state.flags&lockLinearY != 0 {
			v.Y = 0
		}

		if state.flags&lockAngularZ != 0 {
			w = 0
		}

		state.linearVelocity = v
		if limitVelocity(state, maxLinearSpeed) {
			state.flags |= isSpeedCapped
		}

		// Clamp to max angular speed
		if w*w > maxAngularSpeedSquared && (state.flags&allowFastRotation) == 0 {
			ratio := maxAngularSpeed / absFloat(w)
			w *= ratio
			state.flags |= isSpeedCapped
		}

		state.angularVelocity = w
		state.deltaPosition = state.deltaPosition.MulAdd(h, state.linearVelocity)
		state.deltaRotation = state.deltaRotation.Integrate(h * state.angularVelocity)
	}
}

func continuousQueryCallback(userData uint64, context *continuousContext) bool {
	shapeID := int(userData)
	continuousContext := context
	fastShape := continuousContext.fastShape
	fastBodySim := continuousContext.fastBodySim

	core.Assert(fastShape.sensorIndex == nullIndex, "fastShape->sensorIndex == B2_NULL_INDEX")

	// Skip same shape
	if shapeID == fastShape.id {
		return true
	}

	world := continuousContext.world
	shape := &world.shapes[shapeID]

	// Skip same body
	if shape.bodyID == fastShape.bodyID {
		return true
	}

	bodySim, collide := continuousCollisionBody(continuousContext, shape)
	if !collide {
		return true
	}

	isSensor := shape.sensorIndex != nullIndex

	// Early out on fast parallel movement over a chain shape.
	if shape.shapeType == ChainSegmentShape {
		transform := bodySim.transform.Relative(continuousContext.base)
		p1 := transform.Apply(shape.chainSegment.Segment.Point1)
		p2 := transform.Apply(shape.chainSegment.Segment.Point2)
		e := p2.Sub(p1)

		length, e := e.LengthAndNormalize()
		if length > LinearSlop() {
			c1 := continuousContext.centroid1
			separation1 := c1.Sub(p1).Cross(e)
			c2 := continuousContext.centroid2
			separation2 := c2.Sub(p1).Cross(e)

			coreDistance := coreFraction * fastBodySim.minExtent

			if separation1 < 0.0 || (separation1-separation2 < coreDistance && separation2 > coreDistance) {
				// Minimal clipping
				return true
			}
		}
	}

	// todo_erin testing early out for segments

	var input TOIInput

	input.ProxyA = makeShapeDistanceProxy(shape)
	input.ProxyB = makeShapeDistanceProxy(fastShape)
	input.SweepA = makeRelativeSweep(bodySim, continuousContext.base)
	input.SweepB = continuousContext.sweep
	input.MaxFraction = continuousContext.fraction

	output := TimeOfImpact(&input)
	if isSensor {
		// Only accept a sensor hit that is sooner than the current solid hit.
		if output.Fraction <= continuousContext.fraction && continuousContext.sensorCount < maxContinuousSensorHits {
			index := continuousContext.sensorCount

			// The hit shape is a sensor
			sensorHit := sensorHit{
				sensorID:  shape.id,
				visitorID: fastShape.id,
			}

			continuousContext.sensorHits[index] = sensorHit
			continuousContext.sensorFractions[index] = output.Fraction
			continuousContext.sensorCount += 1
		}

		return true
	}

	resolveContinuousSolidHit(continuousContext, shape, &input, output)

	// Continue query
	return true
}

func solveContinuous(world *World, bodySimIndex int, taskContext *taskContext, dt float32) {
	awakeSet := &world.solverSets[awakeSet]
	fastBodySim := &awakeSet.bodySims[bodySimIndex]
	core.Validate((fastBodySim.flags&isFast) != 0, "fastBodySim->flags & b2_isFast")

	// Re-center the sweep on the fast body so the TOI and the swept query stay in float precision
	base := fastBodySim.center0
	sweep := makeRelativeSweep(fastBodySim, base)

	var xf1 Transform

	xf1.Q = sweep.Q1
	xf1.P = sweep.C1.Sub(sweep.Q1.Rotate(sweep.LocalCenter))

	var xf2 Transform

	xf2.Q = sweep.Q2
	xf2.P = sweep.C2.Sub(sweep.Q2.Rotate(sweep.LocalCenter))

	staticTree := &world.broadPhase.trees[StaticBody]
	kinematicTree := &world.broadPhase.trees[KinematicBody]
	dynamicTree := &world.broadPhase.trees[DynamicBody]
	fastBody := &world.bodies[fastBodySim.bodyID]

	context := continuousContext{}
	context.world = world
	context.sweep = sweep
	context.base = base
	context.fastBodySim = fastBodySim
	context.fraction = 1.0

	isBullet := (fastBodySim.flags & isBullet) != 0

	shapeID := fastBody.headShapeID
	for shapeID != nullIndex {
		fastShape := &world.shapes[shapeID]
		shapeID = fastShape.nextShapeID

		context.fastShape = fastShape
		context.centroid1 = xf1.Apply(fastShape.localCentroid)
		context.centroid2 = xf2.Apply(fastShape.localCentroid)

		box1 := fastShape.aabb

		// xf2 is in the base frame, compute the tight box near the origin then lift to world
		box2 := offsetAABB(computeShapeAABB(fastShape, MakeWorldTransform(xf2)), base)

		// Store this to avoid double computation in the case there is no impact event
		fastShape.aabb = box2

		// No continuous collision for sensors (but still need the updated bounds)
		if fastShape.sensorIndex != nullIndex {
			continue
		}

		sweptBox := box1.Union(box2)

		staticTree.Query(sweptBox, DefaultMaskBits, func(_ int, userData uint64) bool { return continuousQueryCallback(userData, &context) })

		if isBullet {
			kinematicTree.Query(sweptBox, DefaultMaskBits, func(_ int, userData uint64) bool { return continuousQueryCallback(userData, &context) })
			dynamicTree.Query(sweptBox, DefaultMaskBits, func(_ int, userData uint64) bool { return continuousQueryCallback(userData, &context) })
		}
	}

	speculativeDistance := SpeculativeDistance()

	if context.fraction < 1.0 {
		// Handle time of impact event
		q := sweep.Q1.NLerp(sweep.Q2, context.fraction)
		c := sweep.C1.Lerp(sweep.C2, context.fraction)
		origin := c.Sub(q.Rotate(sweep.LocalCenter))

		// Advance body, lifting the base frame result back to world
		fastBodySim.transform.Q = q
		fastBodySim.transform.P = base.Add(origin)
		fastBodySim.center = base.Add(c)
		fastBodySim.rotation0 = q
		fastBodySim.center0 = fastBodySim.center

		// Warning: writing to the body sim flags means we should not read from other body sim flags in this function.
		fastBodySim.flags |= hadTimeOfImpact

		// Timeloss means there is a lost gravity contribution. Other forces and torques are ignored for now.
		fastBodyState := &awakeSet.bodyStates[bodySimIndex]
		v := fastBodyState.linearVelocity
		timeLoss := (1.0 - context.fraction) * dt

		dv := world.gravity.Scale(-timeLoss * fastBodySim.gravityScale)
		if fastBodyState.flags&lockLinearX != 0 {
			dv.X = 0
		}

		if fastBodyState.flags&lockLinearY != 0 {
			dv.Y = 0
		}

		fastBodyState.linearVelocity = v.Add(dv)

		// Update body move event
		event := &world.bodyMoveEvents[bodySimIndex]
		event.Transform = fastBodySim.transform

		// Prepare AABBs for broad-phase.
		// Even though a body is fast, it may not move much. So the AABB may not need enlargement.

		shapeID = fastBody.headShapeID
		for shapeID != nullIndex {
			shape := &world.shapes[shapeID]

			// Must recompute aabb at the interpolated transform
			aabb := computeFatShapeAABB(shape, fastBodySim.transform, speculativeDistance)
			shape.aabb = aabb

			enlargeContinuousShapeBounds(world, fastBodySim, shape, shapeID, isBullet)

			shapeID = shape.nextShapeID
		}
	} else {
		// No time of impact event

		// Advance body
		fastBodySim.rotation0 = fastBodySim.transform.Q
		fastBodySim.center0 = fastBodySim.center

		// Prepare AABBs for broad-phase
		shapeID = fastBody.headShapeID
		for shapeID != nullIndex {
			shape := &world.shapes[shapeID]

			// shape.aabb is still valid from above

			enlargeContinuousShapeBounds(world, fastBodySim, shape, shapeID, isBullet)

			shapeID = shape.nextShapeID
		}
	}

	// Push sensor hits on the task context for serial processing.
	for i := range context.sensorCount {
		// Skip any sensor hits that occurred after a solid hit
		if context.sensorFractions[i] < context.fraction {
			taskContext.sensorHits = array.Push(taskContext.sensorHits, context.sensorHits[i])
		}
	}
}

func finalizeBodiesTask(startIndex int, endIndex int, workerIndex int, context *stepContext) {
	stepContext := context
	world := stepContext.world
	bodies := world.bodies
	states := stepContext.states
	sims := stepContext.sims

	core.Assert(endIndex <= len(world.bodyMoveEvents), "endIndex <= world->bodyMoveEvents.count")

	sleepEnabled := world.enableSleep
	enableContinuous := world.enableContinuous
	timeStep := stepContext.dt
	invTimeStep := stepContext.invDt

	// The body move event array should already have the correct size
	moveEvents := world.bodyMoveEvents

	taskContext := &world.taskContexts[workerIndex]
	enlargedSimBitSet := &taskContext.enlargedSimBitSet
	awakeIslandBitSet := &taskContext.awakeIslandBitSet

	speculativeDistance := SpeculativeDistance()

	for simIndex := startIndex; simIndex < endIndex; simIndex++ {
		state := &states[simIndex]
		sim := &sims[simIndex]

		v := state.linearVelocity
		w := state.angularVelocity

		if !v.IsValid() || !IsValidFloat(w) {
			debugBody := &bodies[sim.bodyID]
			world.logger.Warn("box2d: unstable body", "name", string(bytes.TrimRight(debugBody.name[:], "\x00")))
		}

		core.Assert(v.IsValid(), "b2IsValidVec2( v )")
		core.Assert(IsValidFloat(w), "b2IsValidFloat( w )")

		sim.center = sim.center.Add(state.deltaPosition)
		sim.transform.Q = state.deltaRotation.Mul(sim.transform.Q).Normalize()

		// Use the velocity of the farthest point on the body to account for rotation.
		maxVelocity := v.Length() + float32(absFloat(w)*sim.maxExtent)

		// Sleep needs to observe position correction as well as true velocity.
		maxDeltaPosition := state.deltaPosition.Length() + float32(absFloat(state.deltaRotation.S)*sim.maxExtent)

		// Position correction is not as important for sleep as true velocity.
		positionSleepFactor := float32(0.5)
		sleepVelocity := maxFloat(maxVelocity, positionSleepFactor*invTimeStep*maxDeltaPosition)

		// reset state deltas
		state.deltaPosition = Vec2{}
		state.deltaRotation = Rot{C: 1}

		sim.transform.P = sim.center.Add(sim.transform.Q.Rotate(sim.localCenter).Neg())

		// cache miss here, however I need the shape list below
		body := &bodies[sim.bodyID]
		body.bodyMoveIndex = simIndex
		moveEvents[simIndex].Transform = sim.transform
		moveEvents[simIndex].Body = Body{index1: int32(sim.bodyID + 1), world: world, generation: body.generation}
		moveEvents[simIndex].UserData = body.userData
		moveEvents[simIndex].FellAsleep = false

		// reset applied force and torque
		sim.force = Vec2{}
		sim.torque = 0.0

		// If you hit this then it means you deferred mass computation but never called b2Body_UpdateMassFromShapes
		core.Assert((body.flags&dirtyMass) == 0, "( body->flags & b2_dirtyMass ) == 0")

		// Clear the transient flags (fast, speed capped, had TOI). These flags are conditionally set
		// as part of the code below.
		body.flags &= ^bodyTransientFlags
		sim.flags &= ^(isFast | bodyTransientFlags)

		// The body state flag knows about speed capping (used for debug draw).
		body.flags |= (state.flags & isSpeedCapped)
		state.flags &= ^bodyTransientFlags

		if !sleepEnabled || (body.flags&enableSleep) == 0 || sleepVelocity > body.sleepThreshold {
			// Body is not sleepy
			body.sleepTime = 0.0

			maxMotion := maxFloat(maxDeltaPosition, maxVelocity*timeStep)
			advanceAwakeBody(stepContext, taskContext, body, sim, simIndex, maxMotion, enableContinuous)
		} else {
			// Body is safe to advance and is falling asleep
			sim.center0 = sim.center
			sim.rotation0 = sim.transform.Q
			body.sleepTime += timeStep
		}

		markBodyIslandAwake(world, body, taskContext, awakeIslandBitSet)

		finalizeBodyShapeBounds(world, body, sim, simIndex, enlargedSimBitSet, speculativeDistance)
	}
}

func computeBlockCount(itemCount int, minSize int, maxBlockCount int) blockDim {
	dim := blockDim{}
	if itemCount == 0 {
		return dim
	}

	if itemCount <= minSize*maxBlockCount {
		dim.size = minSize
	} else {
		dim.size = (itemCount + maxBlockCount - 1) / maxBlockCount
	}

	dim.count = (itemCount + dim.size - 1) / dim.size

	core.Assert(dim.count >= 1, "dim.count >= 1")
	core.Assert(dim.size*dim.count >= itemCount, "dim.size * dim.count >= itemCount")

	return dim
}

func initBlocks(blocks []syncBlock, dim blockDim, itemCount int, blockType uint8, colorIndex uint8) {
	if dim.count == 0 {
		return
	}

	core.Assert(itemCount >= dim.count, "itemCount >= dim.count")

	// Compute the number of elements per block
	blockSize := dim.size

	// Simulation too big
	core.Assert(blockSize <= 65535, "blockSize <= UINT16_MAX")

	for i := range dim.count {
		blocks[i].block.startIndex = i * blockSize
		blocks[i].block.count = uint16(blockSize)
		blocks[i].block.blockType = blockType
		blocks[i].block.colorIndex = colorIndex
		blocks[i].syncIndex.Store(int32(0))
	}

	// The last block may not be full
	blocks[dim.count-1].block.count = uint16(itemCount - (dim.count-1)*blockSize)

	core.Validate(int(blocks[dim.count-1].block.count) <= blockSize, "blocks[dim.count - 1].block.count <= blockSize")
	core.Validate((dim.count-1)*dim.size+int(blocks[dim.count-1].block.count) == itemCount, "( dim.count - 1 ) * dim.size + blocks[dim.count - 1].block.count == itemCount")
}

func initStage(stage []solverStage, stageType solverStageType, blocks []syncBlock, blockCount int, colorIndex uint8) []solverStage {
	stage[0].stageType = stageType
	stage[0].blocks = blocks
	stage[0].blockCount = blockCount
	stage[0].colorIndex = colorIndex
	stage[0].completionCount.Store(0)

	return stage[1:]
}

func initColorStages(stage []solverStage, stageType solverStageType, activeColorCount int, colorBlocks [][]syncBlock, colorBlockCounts []int, activeColorIndices []int) []solverStage {
	for i := range activeColorCount {
		stage = initStage(stage, stageType, colorBlocks[i], colorBlockCounts[i], uint8(activeColorIndices[i]))
	}

	return stage
}

func executeBlock(stage *solverStage, context *stepContext, block solverBlock, workerIndex int) {
	stageType := stage.stageType
	blockType := solverBlockType(block.blockType)

	switch stageType {
	case stagePrepareJoints:
		prepareJointsTask(block, context)

	case stagePrepareContacts:
		prepareContactsWide(block, context)

	case stageIntegrateVelocities:
		integrateVelocitiesTask(block, context)

	case stageWarmStart, stageSolve, stageRelax:
		executeGraphBlock(stageType, blockType, context, block, workerIndex)

	case stageIntegratePositions:
		integratePositionsTask(block, context)

	case stageRestitution:
		if blockType == graphContactBlock {
			applyRestitutionWide(block, context)
		}

	case stageStoreImpulses:
		storeImpulsesWide(block, context, workerIndex)
	}
}

func getWorkerStartIndex(workerIndex int, blockCount int, workerCount int) int {
	if blockCount <= workerCount {
		if workerIndex < blockCount {
			return workerIndex
		}

		return nullIndex
	}

	blocksPerWorker := blockCount / workerCount
	remainder := blockCount - blocksPerWorker*workerCount

	return blocksPerWorker*workerIndex + min(remainder, workerIndex)
}

func executeStage(stage *solverStage, context *stepContext, previousSyncIndex int, syncIndex int, workerIndex int) {
	completedCount := 0
	blocks := stage.blocks
	blockCount := stage.blockCount

	startIndex := getWorkerStartIndex(workerIndex, blockCount, context.workerCount)
	if startIndex == nullIndex {
		return
	}

	core.Assert(0 <= startIndex && startIndex < blockCount, "0 <= startIndex && startIndex < blockCount")

	blockIndex := startIndex
	for range blockCount {
		if blocks[blockIndex].syncIndex.CompareAndSwap(int32(previousSyncIndex), int32(syncIndex)) {
			core.Assert(stage.stageType != stagePrepareContacts || syncIndex < 2, "stage->type != b2_stagePrepareContacts || syncIndex < 2")
			core.Assert(completedCount < blockCount, "completedCount < blockCount")

			// Pass the descriptor by value -- the wrapping syncBlock holds the atomic
			// syncIndex but we only copy .block, so the struct copy never aliases the CAS target.
			executeBlock(stage, context, blocks[blockIndex].block, workerIndex)

			completedCount += 1
		}

		blockIndex += 1
		if blockIndex >= blockCount {
			blockIndex = 0
		}
	}

	stage.completionCount.Add(int32(completedCount))
}

func executeMainStage(stage *solverStage, context *stepContext, syncBits uint32) {
	blockCount := stage.blockCount
	if blockCount == 0 {
		return
	}

	workerIndex := 0

	if blockCount == 1 {
		executeBlock(stage, context, stage.blocks[0].block, workerIndex)
	} else {
		context.atomicSyncBits.Store(syncBits)

		syncIndex := int((syncBits >> 16) & 0xFFFF)
		core.Assert(syncIndex > 0, "syncIndex > 0")
		previousSyncIndex := syncIndex - 1

		executeStage(stage, context, previousSyncIndex, syncIndex, workerIndex)

		// Spin waiting for thieves to finish
		for int(stage.completionCount.Load()) != blockCount {
			pause()
		}

		stage.completionCount.Store(int32(0))
	}
}

func solverTask(taskContext *workerContext) {
	workerContext := taskContext
	workerIndex := workerContext.workerIndex
	context := workerContext.context
	activeColorCount := context.activeColorCount
	stages := context.stages
	profile := &context.world.profile

	if workerIndex == 0 {
		// The orchestrator slot is a race. The calling thread of b2World_Step also enters here
		// as worker 0, so progress is guaranteed even if the user's task system schedules tasks
		// out of order, has fewer threads than workerCount, or runs the task synchronously
		// inside enqueueTaskFcn. Whoever wins the CAS becomes the orchestrator; the loser
		// returns and lets the spinner-only path handle workers >0.
		if !context.mainClaimed.CompareAndSwap(int32(0), int32(1)) {
			return
		}

		// Main thread synchronizes the workers and does work itself.
		//
		// This single task is able to fully complete all work even if all other workers are
		// blocked, so a fully serial task system still drives the simulation forward.

		// Stages are re-used by loops so that I don't need more stages for large substep counts.
		// The sync indices grow monotonically for the body/graph/constraint groupings because they share solver blocks.
		// The stage index and sync indices are combined in to sync bits for atomic synchronization.
		// The workers need to compute the previous sync index for a given stage so that CAS works correctly. This
		// setup makes this easy to do.

		/*
			Stage sequence
			stagePrepareJoints,
			stagePrepareContacts,
			stageIntegrateVelocities,
			stageWarmStart,
			stageSolve,
			stageIntegratePositions,
			stageRelax,
			stageRestitution,
			stageStoreImpulses
		*/

		ticks := timer.Ticks()

		bodySyncIndex := 1
		stageIndex := 0

		// Prepare joint constraints
		jointSyncIndex := uint32(1)
		syncBits := (uint32(jointSyncIndex) << 16) | uint32(stageIndex)
		core.Assert(stages[stageIndex].stageType == stagePrepareJoints, "stages[stageIndex].type == b2_stagePrepareJoints")
		executeMainStage(&stages[stageIndex], context, syncBits)
		stageIndex += 1

		// Prepare contact constraints
		contactSyncIndex := uint32(1)
		syncBits = (uint32(contactSyncIndex) << 16) | uint32(stageIndex)
		core.Assert(stages[stageIndex].stageType == stagePrepareContacts, "stages[stageIndex].type == b2_stagePrepareContacts")
		executeMainStage(&stages[stageIndex], context, syncBits)
		stageIndex += 1
		contactSyncIndex += 1

		// Single-threaded overflow work. These constraints don't fit in the graph coloring.
		prepareJointsOverflow(context)
		prepareContactsOverflow(context)

		profile.PrepareConstraints += timer.MillisecondsAndReset(&ticks)

		graphSyncIndex := 1

		subStepCount := context.subStepCount
		for range subStepCount {
			// stage index restarted each iteration
			// syncBits still increases monotonically because the upper bits increase each iteration
			iterationStageIndex := stageIndex

			// Integrate velocities
			syncBits = (uint32(bodySyncIndex) << 16) | uint32(iterationStageIndex)
			core.Assert(stages[iterationStageIndex].stageType == stageIntegrateVelocities, "stages[iterationStageIndex].type == b2_stageIntegrateVelocities")
			executeMainStage(&stages[iterationStageIndex], context, syncBits)
			iterationStageIndex += 1
			bodySyncIndex += 1

			profile.IntegrateVelocities += timer.MillisecondsAndReset(&ticks)

			// Warm start constraints
			warmStartJointsOverflow(context)
			warmStartContactsOverflow(context)

			for range activeColorCount {
				syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
				core.Assert(stages[iterationStageIndex].stageType == stageWarmStart, "stages[iterationStageIndex].type == b2_stageWarmStart")
				executeMainStage(&stages[iterationStageIndex], context, syncBits)
				iterationStageIndex += 1
			}

			graphSyncIndex += 1

			profile.WarmStart += timer.MillisecondsAndReset(&ticks)

			// Solve constraints
			useBias := true
			for range iterations {
				// Overflow constraints have lower priority. Typically these are dynamic-vs-dynamic.
				solveJointsOverflow(context, useBias)
				pushContactsOverflow(context)

				for range activeColorCount {
					syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
					core.Assert(stages[iterationStageIndex].stageType == stageSolve, "stages[iterationStageIndex].type == b2_stageSolve")
					executeMainStage(&stages[iterationStageIndex], context, syncBits)
					iterationStageIndex += 1
				}

				graphSyncIndex += 1
			}

			profile.SolveImpulses += timer.MillisecondsAndReset(&ticks)

			// Integrate positions
			core.Assert(stages[iterationStageIndex].stageType == stageIntegratePositions, "stages[iterationStageIndex].type == b2_stageIntegratePositions")
			syncBits = (uint32(bodySyncIndex) << 16) | uint32(iterationStageIndex)
			executeMainStage(&stages[iterationStageIndex], context, syncBits)
			iterationStageIndex += 1
			bodySyncIndex += 1

			profile.IntegratePositions += timer.MillisecondsAndReset(&ticks)

			// Relax constraints
			useBias = false
			for range relaxIterations {
				solveJointsOverflow(context, useBias)
				solveContactsOverflow(context)

				for range activeColorCount {
					syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
					core.Assert(stages[iterationStageIndex].stageType == stageRelax, "stages[iterationStageIndex].type == b2_stageRelax")
					executeMainStage(&stages[iterationStageIndex], context, syncBits)
					iterationStageIndex += 1
				}

				graphSyncIndex += 1
			}

			profile.RelaxImpulses += timer.MillisecondsAndReset(&ticks)
		}

		// Advance the stage according to the sub-stepping tasks just completed
		// integrate velocities / warm start / solve / integrate positions / relax
		stageIndex += 1 + activeColorCount + iterations*activeColorCount + 1 + relaxIterations*activeColorCount

		restitutionIterations := context.world.restitutionIterations
		if restitutionIterations > 0 && int(context.anyRestitution.Load()) != 0 {
			for range restitutionIterations {
				applyRestitutionOverflow(context)

				iterationStageIndex := stageIndex
				for range activeColorCount {
					syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
					core.Assert(stages[iterationStageIndex].stageType == stageRestitution, "stages[iterationStageIndex].type == b2_stageRestitution")
					executeMainStage(&stages[iterationStageIndex], context, syncBits)
					iterationStageIndex += 1
				}

				graphSyncIndex += 1
			}

			profile.Restitution += timer.MillisecondsAndReset(&ticks)
		}

		stageIndex += activeColorCount

		// Store impulses
		storeImpulsesOverflow(context)

		syncBits = (uint32(contactSyncIndex) << 16) | uint32(stageIndex)
		core.Assert(stages[stageIndex].stageType == stageStoreImpulses, "stages[stageIndex].type == b2_stageStoreImpulses")
		executeMainStage(&stages[stageIndex], context, syncBits)

		profile.StoreImpulses += timer.MillisecondsAndReset(&ticks)

		// Signal workers to finish
		context.atomicSyncBits.Store(^uint32(0))

		core.Assert(stageIndex+1 == context.stageCount, "stageIndex + 1 == context->stageCount")

		return
	}

	// Worker spins and waits for work
	lastSyncBits := uint32(0)
	// maxSpinTime := 10
	for {
		// Spin until main thread bumps changes the sync bits. This can waste significant time overall, but it is necessary for
		// parallel simulation with graph coloring.
		// todo improve this spinner
		var syncBits uint32

		spinCount := 0

		for {
			syncBits = context.atomicSyncBits.Load()
			if syncBits != lastSyncBits {
				break
			}

			if spinCount > 5 {
				timer.Yield()

				spinCount = 0
			} else {
				// Using the cycle counter helps to account for variation in mm_pause timing across different
				// CPUs. However, this is X64 only.
				// prev := __rdtsc()
				// do
				//{
				//	_mm_pause()
				//}
				// while ((__rdtsc() - prev) < maxSpinTime)
				// maxSpinTime += 10
				pause()
				pause()

				spinCount += 1
			}
		}

		if syncBits == ^uint32(0) {
			// sentinel hit
			break
		}

		stageIndex := int(syncBits & 0xFFFF)
		core.Assert(stageIndex < context.stageCount, "stageIndex < context->stageCount")

		syncIndex := int((syncBits >> 16) & 0xFFFF)
		core.Assert(syncIndex > 0, "syncIndex > 0")

		previousSyncIndex := syncIndex - 1

		stage := &stages[stageIndex]
		executeStage(stage, context, previousSyncIndex, syncIndex, workerIndex)

		lastSyncBits = syncBits
	}
}

func bulletBodyTask(startIndex int, endIndex int, workerIndex int, context *stepContext) {
	stepContext := context
	taskContext := &stepContext.world.taskContexts[workerIndex]

	core.Assert(startIndex <= endIndex, "startIndex <= endIndex")

	for i := startIndex; i < endIndex; i++ {
		simIndex := stepContext.bulletBodies[i]
		solveContinuous(stepContext.world, simIndex, taskContext, stepContext.dt)
	}
}

func clearSolverScratch(graph *constraintGraph) {
	for i := range graph.colors {
		color := &graph.colors[i]
		color.wideConstraints = nil
		color.overflowConstraints = nil
		color.wideConstraintCount = 0
	}
}

func startSplitIslandTask(world *World) func() {
	if world.splitIslandID == nullIndex {
		return nil
	}

	if world.taskCount >= MaxTasks {
		splitIslandTask(world)
		return nil
	}

	splitTask := world.enqueueTaskFcn(func() { splitIslandTask(world) })
	world.taskCount += 1

	if splitTask != nil {
		world.activeTaskCount += 1
	}

	return splitTask
}

func runSolverWorkers(world *World, stepContext *stepContext, workerContexts []workerContext, workerCount int, splitTask func()) {
	constraintTicks := timer.Ticks()
	jointIDCapacity := world.jointIDPool.Capacity()
	contactIDCapacity := world.contactIDPool.Capacity()

	for i := range workerCount {
		taskContext := &world.taskContexts[i]
		taskContext.jointStateBitSet.SetBitCountAndClear(uint32(jointIDCapacity))
		taskContext.hitEventBitSet.SetBitCountAndClear(uint32(contactIDCapacity))
		taskContext.hasHitEvents = false

		workerContexts[i].context = stepContext
		workerContexts[i].workerIndex = i

		if world.taskCount < MaxTasks {
			workerContexts[i].userTask = world.enqueueTaskFcn(func() { solverTask(&workerContexts[i]) })
			world.taskCount += 1

			if workerContexts[i].userTask != nil {
				world.activeTaskCount += 1
			}
		} else {
			workerContexts[i].userTask = nil
			solverTask(&workerContexts[i])
		}
	}

	// The calling thread of b2World_Step also enters solverTask as worker 0 and races for the
	// orchestrator slot via the CAS inside. This guarantees progress even when the user's task
	// system can't run the queued worker 0 promptly: it might schedule out of order, have fewer
	// threads than workerCount, or invert priority by parking the calling thread in finishTaskFcn.
	// Whoever wins the CAS becomes the orchestrator; the loser returns and lets the spinner-only
	// path handle workers >0.
	callerContext := workerContext{context: stepContext}
	solverTask(&callerContext)

	// Finish constraint solve
	for i := range workerCount {
		if workerContexts[i].userTask != nil {
			workerContexts[i].userTask()

			world.activeTaskCount -= 1
		}
	}

	// Finish island split
	if splitTask != nil {
		splitTask()

		world.activeTaskCount -= 1
	}

	world.splitIslandID = nullIndex
	world.profile.Constraints = timer.MillisecondsAndReset(&constraintTicks)
}

func prepareSolverBlocks(world *World, bodyCount, contactCount, jointCount, graphCount int) ([]syncBlock, []syncBlock, []syncBlock, []syncBlock) {
	if bodyCount > cap(world.solverBodyBlockScratch) {
		world.solverBodyBlockScratch = make([]syncBlock, bodyCount)
	}

	bodyBlocks := world.solverBodyBlockScratch[:bodyCount]
	clear(bodyBlocks)

	if contactCount > cap(world.solverContactBlockScratch) {
		world.solverContactBlockScratch = make([]syncBlock, contactCount)
	}

	contactBlocks := world.solverContactBlockScratch[:contactCount]
	clear(contactBlocks)

	if jointCount > cap(world.solverJointBlockScratch) {
		world.solverJointBlockScratch = make([]syncBlock, jointCount)
	}

	jointBlocks := world.solverJointBlockScratch[:jointCount]
	clear(jointBlocks)

	if graphCount > cap(world.solverGraphBlockScratch) {
		world.solverGraphBlockScratch = make([]syncBlock, graphCount)
	}

	graphBlocks := world.solverGraphBlockScratch[:graphCount]
	clear(graphBlocks)

	return bodyBlocks, contactBlocks, jointBlocks, graphBlocks
}

func finalizeSolverBodies(world *World, stepContext *stepContext, awake *solverSet, awakeBodyCount int) {
	transformTicks := timer.Ticks()

	// Prepare contact, enlarged body, and island bit sets used in body finalization.
	awakeIslandCount := len(awake.islandSims)

	for i := range world.workerCount {
		taskContext := &world.taskContexts[i]
		taskContext.sensorHits = taskContext.sensorHits[:0]
		taskContext.enlargedSimBitSet.SetBitCountAndClear(uint32(awakeBodyCount))
		taskContext.awakeIslandBitSet.SetBitCountAndClear(uint32(awakeIslandCount))
		taskContext.splitIslandID = nullIndex
		taskContext.splitSleepTime = 0.0
	}

	// Finish the user tree task that was queued earlier in the time step. This must be complete before touching the
	// broad-phase.
	if world.userTreeTask != nil {
		world.userTreeTask()
		world.userTreeTask = nil
		world.activeTaskCount -= 1
	}

	validateNoEnlarged(&world.broadPhase)

	// Finalize bodies. Must happen after the constraint solver and after island splitting.
	parallelFor(world, func(start, end, worker int) { finalizeBodiesTask(start, end, worker, stepContext) }, awakeBodyCount, 64)

	world.profile.Transforms = timer.Milliseconds(transformTicks)
}

func solve(world *World, stepContext *stepContext) {
	// Only count steps that advance the simulation
	world.stepIndex += 1

	// Are there any awake bodies? This scenario should not be important for profiling.
	awake := &world.solverSets[awakeSet]

	awakeBodyCount := len(awake.bodySims)
	if awakeBodyCount == 0 {
		clearSolverScratch(&world.constraintGraph)
		return
	}

	// Solve constraints using graph coloring
	{
		setupTicks := timer.Ticks()

		// Prepare buffers for bullets
		stepContext.bulletBodyCount.Store(int32(0))

		world.solverBulletScratch = array.Reserve(world.solverBulletScratch, awakeBodyCount)
		stepContext.bulletBodies = world.solverBulletScratch[:awakeBodyCount]

		graph := &world.constraintGraph
		colors := graph.colors[:]

		stepContext.sims = awake.bodySims
		stepContext.states = awake.bodyStates

		// prepare for move events
		world.bodyMoveEvents = array.Reserve(world.bodyMoveEvents, awakeBodyCount)
		world.bodyMoveEvents = world.bodyMoveEvents[:awakeBodyCount]

		workerCount := world.workerCount

		// Target 4 blocks per worker to allow work stealing
		maxBlockCount := 4 * workerCount

		// Body blocks are for parallel iteration over bodies directly (integration, update transforms)
		minBodiesPerBlock := 32
		bodyDim := computeBlockCount(awakeBodyCount, minBodiesPerBlock, maxBlockCount)

		minContactsPerBlock := 4
		minJointsPerBlock := 4

		// Configure blocks for tasks parallel-for each active graph color
		// The blocks are a mix of wide contact blocks and joint blocks
		var (
			activeColorIndices [GraphColorCount]int
			colorContactCounts [GraphColorCount]int
			colorJointCounts   [GraphColorCount]int
			graphContactDims   [GraphColorCount]blockDim
			graphJointDims     [GraphColorCount]blockDim
		)

		graphBlockCount := 0

		// c is the active color index
		wideContactCount := 0
		jointCount := 0
		c := 0

		for i := range GraphColorCount - 1 {
			colorContactCount := len(colors[i].contactSims)
			colorJointCount := len(colors[i].jointSims)

			if colorContactCount+colorJointCount == 0 {
				continue
			}

			activeColorIndices[c] = i

			// Ceiling for wide constraint count
			colorContactCountW := 0
			if colorContactCount > 0 {
				colorContactCountW = ((colorContactCount - 1) >> simdShift) + 1
			}

			wideContactCount += colorContactCountW
			colorContactCounts[c] = colorContactCountW

			colorJointCounts[c] = colorJointCount
			jointCount += colorJointCount

			// Graph solver block dimensions
			graphContactDims[c] = computeBlockCount(colorContactCountW, minContactsPerBlock, maxBlockCount)
			graphJointDims[c] = computeBlockCount(colorJointCount, minJointsPerBlock, maxBlockCount)
			graphBlockCount += graphContactDims[c].count + graphJointDims[c].count

			c += 1
		}

		activeColorCount := c

		// Prepare and store run as one flat parallel-for over the entire wide constraint range,
		// partitioned into uniformly sized blocks. Color info is consulted inside the task via
		// a small span array, so blocks do not need to honor color boundaries here.
		contactPrepareDim := computeBlockCount(wideContactCount, minContactsPerBlock, maxBlockCount)
		jointPrepareDim := computeBlockCount(jointCount, minJointsPerBlock, maxBlockCount)

		wideContactConstraints := reserveSolverContactScratch(world, colors, wideContactCount)

		// Build the span table for the flat prepare/store parallel-for while I slice the
		// wide constraint buffer across colors. One entry per active color plus a sentinel
		// at wideContactCount.
		contactPrepareSpans := world.contactPrepareSpanScratch[:]
		jointPrepareSpans := world.jointPrepareSpanScratch[:]

		prepareSolverConstraintSpans(colors, activeColorCount, activeColorIndices[:], wideContactConstraints, wideContactCount, jointCount, contactPrepareSpans, jointPrepareSpans)

		stageCount := 0

		// stagePrepareJoints
		stageCount += 1
		// stagePrepareContacts
		stageCount += 1
		// stageIntegrateVelocities
		stageCount += 1
		// stageWarmStart
		stageCount += activeColorCount
		// stageSolve
		stageCount += iterations * activeColorCount
		// stageIntegratePositions
		stageCount += 1
		// stageRelax
		stageCount += relaxIterations * activeColorCount
		// stageRestitution
		stageCount += activeColorCount
		// stageStoreImpulses
		stageCount += 1

		if stageCount > cap(world.solverStageScratch) {
			world.solverStageScratch = make([]solverStage, stageCount)
		}

		stages := world.solverStageScratch[:stageCount]
		bodyBlocks, contactBlocks, jointBlocks, graphBlocks := prepareSolverBlocks(world, bodyDim.count, contactPrepareDim.count, jointPrepareDim.count, graphBlockCount)

		// Split an awake island. This modifies:
		// - stack allocator
		// - world island array and solver set
		// - island indices on bodies, contacts, and joints
		// I'm squeezing this task in here because it may be expensive and this is a safe place to put it.
		// Note: cannot split islands in parallel with FinalizeBodies
		splitTask := startSplitIslandTask(world)

		// Prepare body blocks
		initBlocks(bodyBlocks, bodyDim, awakeBodyCount, uint8(bodyBlock), 255)

		// Prepare blocks as a single flat parallel-for over the whole constraint range.
		// The task walks spans to decode flat slot indices back to per-color arrays.
		initBlocks(contactBlocks, contactPrepareDim, wideContactCount, uint8(contactBlock), 255)
		initBlocks(jointBlocks, jointPrepareDim, jointCount, uint8(jointBlock), 255)

		// Prepare graph work blocks. Each color gets joint blocks followed by contact blocks.
		var graphColorBlocks [GraphColorCount][]syncBlock

		baseGraphBlock := graphBlocks

		var graphBlockCounts [GraphColorCount]int

		for i := range activeColorCount {
			graphColorBlocks[i] = baseGraphBlock

			colorIndex := uint8(activeColorIndices[i])
			initBlocks(baseGraphBlock, graphJointDims[i], colorJointCounts[i], uint8(graphJointBlock), colorIndex)
			baseGraphBlock = baseGraphBlock[graphJointDims[i].count:]

			initBlocks(baseGraphBlock, graphContactDims[i], colorContactCounts[i], uint8(graphContactBlock), colorIndex)
			baseGraphBlock = baseGraphBlock[graphContactDims[i].count:]

			graphBlockCounts[i] = graphJointDims[i].count + graphContactDims[i].count
		}

		core.Assert(len(graphBlocks)-len(baseGraphBlock) == graphBlockCount, "(ptrdiff_t)( baseGraphBlock - graphBlocks ) == graphBlockCount")

		stage := stages
		stage = initStage(stage, stagePrepareJoints, jointBlocks, jointPrepareDim.count, 255)
		stage = initStage(stage, stagePrepareContacts, contactBlocks, contactPrepareDim.count, 255)
		stage = initStage(stage, stageIntegrateVelocities, bodyBlocks, bodyDim.count, 255)
		stage = initColorStages(stage, stageWarmStart, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initColorStages(stage, stageSolve, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initStage(stage, stageIntegratePositions, bodyBlocks, bodyDim.count, 255)
		stage = initColorStages(stage, stageRelax, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initColorStages(stage, stageRestitution, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initStage(stage, stageStoreImpulses, contactBlocks, contactPrepareDim.count, 255)

		core.Assert(len(stages)-len(stage) == stageCount, "(int)( stage - stages ) == stageCount")

		core.Assert(workerCount <= MaxWorkers, "workerCount <= B2_MAX_WORKERS")

		workerContexts := world.solverWorkerScratch[:]

		stepContext.graph = graph
		stepContext.activeColorCount = activeColorCount
		stepContext.workerCount = workerCount
		stepContext.stageCount = stageCount
		stepContext.stages = stages
		stepContext.wideContactConstraints = wideContactConstraints
		stepContext.contactPrepareSpans = contactPrepareSpans[:activeColorCount+1]
		stepContext.wideContactCount = wideContactCount
		stepContext.jointPrepareSpans = jointPrepareSpans[:activeColorCount+1]
		stepContext.atomicSyncBits.Store(0)
		stepContext.mainClaimed.Store(int32(0))

		world.profile.SolverSetup = timer.MillisecondsAndReset(&setupTicks)

		runSolverWorkers(world, stepContext, workerContexts, workerCount, splitTask)

		finalizeSolverBodies(world, stepContext, awake, awakeBodyCount)
	}

	// Report joint events
	reportSolverJointEvents(world)

	// Report hit events
	reportSolverHitEvents(world)
	{
		refitTicks := timer.Ticks()

		bp := &world.broadPhase
		bp.trees[KinematicBody].refit()
		bp.trees[DynamicBody].refit()

		validateBroadphase(&world.broadPhase)

		world.profile.Refit = timer.Milliseconds(refitTicks)
	}

	// Bullets are processed after the broad-phase refit so they can query the
	// final non-bullet world.
	solveBulletBodies(world, stepContext, awake)

	// Report sensor hits. This may include bullets sensor hits.
	reportSolverSensorHits(world)

	// Island sleeping
	// This must be done last because putting islands to sleep invalidates the enlarged body bits.
	// todo_erin figure out how to do this in parallel with tree refit
	sleepSolverIslands(world, awake)

	clearSolverScratch(&world.constraintGraph)
}

func makeSoft(hertz float32, zeta float32, h float32) softness {
	if hertz == 0 {
		return softness{}
	}

	omega := float32(2) * Pi * hertz
	a1 := float32(float32(2)*zeta) + float32(h*omega)
	a2 := float32(h * omega * a1)
	a3 := float32(1) / (float32(1) + a2)

	return softness{biasRate: omega / a1, massScale: a2 * a3, impulseScale: a3}
}

// Report joint events.
func reportSolverJointEvents(world *World) {
	jointEventTicks := timer.Ticks()

	// Gather bits for all joints that have force/torque events
	jointStateBitSet := &world.taskContexts[0].jointStateBitSet
	for i := 1; i < world.workerCount; i++ {
		jointStateBitSet.InPlaceUnion(&world.taskContexts[i].jointStateBitSet)
	}
	{
		wordCount := jointStateBitSet.BlockCount
		words := jointStateBitSet.Bits

		jointArray := world.joints

		for k := range wordCount {
			word := words[k]
			for word != 0 {
				ctz := bits.TrailingZeros64(word)
				jointID := int(64*k) + ctz

				core.Assert(jointID < cap(world.joints), "jointId < world->joints.capacity")

				joint := &jointArray[jointID]

				core.Assert(joint.setIndex == awakeSet, "joint->setIndex == b2_awakeSet")

				event := JointEvent{Joint: Joint{index1: int32(jointID + 1), world: world, generation: joint.generation}, UserData: joint.userData}

				world.jointEvents = array.Push(world.jointEvents, event)

				// Clear the smallest set bit
				word &= word - 1
			}
		}
	}

	world.profile.JointEvents = timer.Milliseconds(jointEventTicks)
}

// Report hit events.
func reportSolverHitEvents(world *World) {
	hitTicks := timer.Ticks()

	core.Assert(len(world.contactHitEvents) == 0, "world->contactHitEvents.count == 0")

	// Fast path: skip if no worker flagged any hit-event candidates during impulse storing.
	anyHitEvents := false

	for i := range world.workerCount {
		if world.taskContexts[i].hasHitEvents {
			anyHitEvents = true
			break
		}
	}

	if !anyHitEvents {
		world.profile.HitEvents = timer.Milliseconds(hitTicks)
		return
	}
	// Union per-worker bits into worker 0's bit set.
	hitEventBitSet := &world.taskContexts[0].hitEventBitSet
	for i := 1; i < world.workerCount; i++ {
		if world.taskContexts[i].hasHitEvents {
			hitEventBitSet.InPlaceUnion(&world.taskContexts[i].hitEventBitSet)
		}
	}

	threshold := world.hitEventThreshold
	colors := world.constraintGraph.colors[:]
	contactArray := world.contacts

	wordCount := hitEventBitSet.BlockCount

	words := hitEventBitSet.Bits
	for k := range wordCount {
		word := words[k]
		for word != 0 {
			ctz := bits.TrailingZeros64(word)
			contactID := int(64*k) + ctz

			contact := &contactArray[contactID]
			core.Assert(contact.setIndex == awakeSet && contact.colorIndex != nullIndex, "contact->setIndex == b2_awakeSet && contact->colorIndex != B2_NULL_INDEX")

			color := &colors[contact.colorIndex]
			contactSim := &color.contactSims[contact.localIndex]

			reportSolverContactHit(world, contact, contactSim, threshold)

			// Clear the smallest set bit
			word &= word - 1
		}
	}

	world.profile.HitEvents = timer.Milliseconds(hitTicks)
}

// Report sensor hits. This may include bullets sensor hits.
func reportSolverSensorHits(world *World) {
	sensorHitTicks := timer.Ticks()

	workerCount := world.workerCount
	core.Assert(workerCount == len(world.taskContexts), "workerCount == world->taskContexts.count")

	for i := range workerCount {
		taskContext := &world.taskContexts[i]
		hitCount := len(taskContext.sensorHits)
		hits := taskContext.sensorHits

		for j := range hitCount {
			hit := hits[j]
			sensorShape := &world.shapes[hit.sensorID]
			visitingShape := &world.shapes[hit.visitorID]

			sensor := &world.sensors[sensorShape.sensorIndex]
			shapeRef := visitor{shapeID: hit.visitorID, generation: visitingShape.generation}
			sensor.hits = array.Push(sensor.hits, shapeRef)
		}
	}

	world.profile.SensorHits = timer.Milliseconds(sensorHitTicks)
}

func reportSolverContactHit(world *World, contact *contact, contactSim *contactSim, threshold float32) {
	shapeArray := world.shapes
	event := ContactHitEvent{}
	event.ApproachSpeed = threshold

	var bestPoint *ManifoldPoint

	pointCount := contactSim.manifold.PointCount
	for p := range pointCount {
		mp := &contactSim.manifold.Points[p]
		approachSpeed := -mp.NormalVelocity

		// Need to check total impulse because the point may be speculative and not colliding
		if approachSpeed > event.ApproachSpeed && mp.TotalNormalImpulse > 0.0 {
			event.ApproachSpeed = approachSpeed
			bestPoint = mp
		}
	}

	core.Validate(bestPoint != nil, "bestPoint != NULL")

	if bestPoint != nil {
		event.Normal = contactSim.manifold.Normal

		shapeA := &shapeArray[contactSim.shapeIDA]
		shapeB := &shapeArray[contactSim.shapeIDB]

		// World contact point reconstructed from a body center of mass and the matching
		// anchor. The anchors were built with the manifold, so a body that has moved since
		// drags the point with it. A static body has not moved, prefer one so the common
		// case of a fast body striking the world stays exact.
		bodyA := &world.bodies[shapeA.bodyID]

		bodyB := &world.bodies[shapeB.bodyID]
		if bodyA.bodyType != StaticBody && bodyB.bodyType == StaticBody {
			bodySimB := getBodySim(world, bodyB)
			event.Point = bodySimB.center.Add(bestPoint.AnchorB)
		} else {
			bodySimA := getBodySim(world, bodyA)
			event.Point = bodySimA.center.Add(bestPoint.AnchorA)
		}

		event.ShapeA = Shape{index1: int32(shapeA.id + 1), world: world, generation: shapeA.generation}
		event.ShapeB = Shape{index1: int32(shapeB.id + 1), world: world, generation: shapeB.generation}

		event.Contact = Contact{
			index1:     int32(contact.contactID + 1),
			world:      world,
			generation: contact.generation,
		}

		world.contactHitEvents = array.Push(world.contactHitEvents, event)
	}
}

func continuousCollisionBody(context *continuousContext, shape *shape) (*bodySim, bool) {
	world := context.world
	fastShape := context.fastShape
	fastBodySim := context.fastBodySim
	// Skip sensors unless the shapes want sensor events
	isSensor := shape.sensorIndex != nullIndex
	if isSensor && (!shape.enableSensorEvents || !fastShape.enableSensorEvents) {
		return nil, false
	}
	// Skip filtered shapes
	if !shouldShapesCollide(fastShape.filter, shape.filter) {
		return nil, false
	}

	body := &world.bodies[shape.bodyID]
	bodySim := getBodySim(world, body)
	core.Assert(body.bodyType == StaticBody || (fastBodySim.flags&isBullet) != 0, "body->type == b2_staticBody || ( fastBodySim->flags & b2_isBullet )")
	// Skip bullets
	// Warning: it is only safe to read flags from other bodies, not body sims because there are
	// body sim flag writes in the continuous solver.
	if (body.flags & isBullet) != 0 {
		return nil, false
	}
	// Skip filtered bodies
	fastBody := &world.bodies[fastBodySim.bodyID]
	if !shouldBodiesCollide(world, fastBody, body) {
		return nil, false
	}
	// Custom user filtering
	if shape.enableCustomFiltering || fastShape.enableCustomFiltering {
		customFilterFcn := world.customFilterFcn
		if customFilterFcn != nil {
			idA := Shape{index1: int32(shape.id + 1), world: world, generation: shape.generation}

			idB := Shape{index1: int32(fastShape.id + 1), world: world, generation: fastShape.generation}
			if !customFilterFcn(idA, idB) {
				return nil, false
			}
		}
	}

	return bodySim, true
}

func resolveContinuousSolidHit(continuousContext *continuousContext, shape *shape, input *TOIInput, output TOIOutput) {
	world := continuousContext.world
	fastShape := continuousContext.fastShape
	hitFraction := continuousContext.fraction
	didHit := false

	if 0.0 < output.Fraction && output.Fraction < continuousContext.fraction {
		hitFraction = output.Fraction
		didHit = true
	} else if output.Fraction == 0.0 {
		// fallback to TOI of a small circle around the fast shape centroid
		centroid := getShapeCentroid(fastShape)
		extent := computeShapeExtent(fastShape, centroid)
		radius := coreFraction * extent.minExtent
		input.ProxyB = MakeProxy([]Vec2{centroid}, radius)

		output = TimeOfImpact(input)
		if 0.0 < output.Fraction && output.Fraction < continuousContext.fraction {
			hitFraction = output.Fraction
			didHit = true
		}
	}

	if didHit && (shape.enablePreSolveEvents || fastShape.enablePreSolveEvents) && world.preContinuousFcn != nil {
		shapeIDA := Shape{index1: int32(shape.id + 1), world: world, generation: shape.generation}
		shapeIDB := Shape{index1: int32(fastShape.id + 1), world: world, generation: fastShape.generation}

		// TOI runs in the base frame, lift the hit point back to world for the callback
		worldPoint := continuousContext.base.Add(output.Point)
		didHit = world.preContinuousFcn(shapeIDA, shapeIDB, worldPoint, output.Normal)
	}

	if didHit {
		continuousContext.fraction = hitFraction
	}
}

func enlargeContinuousShapeBounds(world *World, fastBodySim *bodySim, shape *shape, shapeID int, bullet bool) {
	shapeFatAABB := &world.fatAABBs[shapeID]
	if shapeFatAABB.Contains(shape.aabb) {
		return
	}

	margin := shape.aabbMargin

	// Note: far from the origin the margin can be snapped to the nearest ULP.
	// This relevant for DP mode. So we lose the broad-phase hysteresis.
	var fatAABB AABB

	fatAABB.LowerBound.X = shape.aabb.LowerBound.X - margin
	fatAABB.LowerBound.Y = shape.aabb.LowerBound.Y - margin
	fatAABB.UpperBound.X = shape.aabb.UpperBound.X + margin
	fatAABB.UpperBound.Y = shape.aabb.UpperBound.Y + margin
	*shapeFatAABB = fatAABB

	// Regular bodies mark the hierarchy as enlarged using atomic operations.
	// Bullets are handled separately at a later stage.
	if !bullet {
		world.broadPhase.markProxyMoved(shape.proxyKey, fatAABB)
	} else {
		fastBodySim.flags |= enlargeBulletBounds
	}
}

func advanceAwakeBody(stepContext *stepContext, taskContext *taskContext, body *body, sim *bodySim, simIndex int, maxMotion float32, enableContinuous bool) {
	if body.bodyType == DynamicBody && enableContinuous && maxMotion > body.safetyFactor*sim.minExtent {
		// This flag is used for debug draw and contact recycling.
		sim.flags |= isFast
		// Store fast bullets for processing later.
		// This is deterministic because the order of TOI sweeps doesn't matter.
		if (sim.flags & isBullet) != 0 {
			bulletIndex := int(stepContext.bulletBodyCount.Add(int32(1))) - int(1)
			stepContext.bulletBodies[bulletIndex] = simIndex
		} else {
			solveContinuous(stepContext.world, simIndex, taskContext, stepContext.dt)
		}
	} else {
		// Body is safe to advance
		sim.center0 = sim.center
		sim.rotation0 = sim.transform.Q
	}
}

func markBodyIslandAwake(world *World, body *body, taskContext *taskContext, awakeIslandBitSet *bitset.Set) {
	// Any single body in an island can keep it awake
	island := &world.islands[body.islandID]
	if body.sleepTime < TimeToSleep {
		// keep island awake
		islandIndex := island.localIndex
		awakeIslandBitSet.SetBit(uint32(islandIndex))
	} else if island.constraintRemoveCount > 0 {
		// Body wants to sleep but its island needs splitting first. Track the sleepiest candidate.
		// Break sleep time ties using the island id to ensure determinism. The cross worker reduction
		// breaks ties the same way.
		if body.sleepTime > taskContext.splitSleepTime ||
			(body.sleepTime == taskContext.splitSleepTime && body.islandID > taskContext.splitIslandID) {
			taskContext.splitIslandID = body.islandID
			taskContext.splitSleepTime = body.sleepTime
		}
	}
}

func finalizeBodyShapeBounds(world *World, body *body, sim *bodySim, simIndex int, enlargedSimBitSet *bitset.Set, speculativeDistance float32) {
	// Update shapes AABBs
	transform := sim.transform
	isFast := (sim.flags & isFast) != 0

	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]

		if isFast {
			// For fast non-bullet bodies the AABB has already been updated in solveContinuous
			// For fast bullet bodies the AABB will be updated at a later stage

			// Add to enlarged shapes regardless of AABB changes.
			// Bit-set to keep the move array sorted
			enlargedSimBitSet.SetBit(uint32(simIndex))
		} else {
			// Update bounds for slow bodies, including slow bullets.
			aabb := computeFatShapeAABB(shape, transform, speculativeDistance)
			shape.aabb = aabb

			shapeFatAABB := &world.fatAABBs[shapeID]
			if !shapeFatAABB.Contains(aabb) {
				margin := shape.aabbMargin

				var fatAABB AABB

				fatAABB.LowerBound.X = aabb.LowerBound.X - margin
				fatAABB.LowerBound.Y = aabb.LowerBound.Y - margin
				fatAABB.UpperBound.X = aabb.UpperBound.X + margin
				fatAABB.UpperBound.Y = aabb.UpperBound.Y + margin
				*shapeFatAABB = fatAABB

				// Mark the hierarchy as enlarged using atomic operations.
				world.broadPhase.markProxyMoved(shape.proxyKey, fatAABB)

				// Bit-set to keep the move array sorted
				enlargedSimBitSet.SetBit(uint32(simIndex))
			}
		}

		shapeID = shape.nextShapeID
	}
}

func solveBulletBodies(world *World, stepContext *stepContext, awake *solverSet) {
	bulletBodyCount := int(stepContext.bulletBodyCount.Load())
	if bulletBodyCount > 0 {
		bulletTicks := timer.Ticks()

		// Fast bullet bodies
		// Note: a bullet body may be moving slow
		minRange := 8
		parallelFor(world, func(start, end, worker int) { bulletBodyTask(start, end, worker, stepContext) }, bulletBodyCount, minRange)

		broadPhase := &world.broadPhase
		dynamicTree := &broadPhase.trees[DynamicBody]

		// Fast array access is important here
		bodyArray := world.bodies
		bodySimArray := awake.bodySims
		shapeArray := world.shapes
		bulletBodySimIndices := stepContext.bulletBodies

		// Serially enlarge broad-phase proxies for bullet shapes.
		// This loop has non-deterministic order but it shouldn't affect the result.
		for i := range bulletBodyCount {
			bulletBodySim := &bodySimArray[bulletBodySimIndices[i]]

			// It is worth tracking this flag because bullets may be moving slowly.
			if (bulletBodySim.flags & enlargeBulletBounds) == 0 {
				continue
			}

			// Clear flag
			bulletBodySim.flags &= ^enlargeBulletBounds

			bodyID := bulletBodySim.bodyID
			core.Assert(0 <= bodyID && bodyID < len(world.bodies), "0 <= bodyId && bodyId < world->bodies.count")
			bulletBody := &bodyArray[bodyID]

			shapeID := bulletBody.headShapeID
			for shapeID != nullIndex {
				shape := &shapeArray[shapeID]
				proxyKey := shape.proxyKey
				proxyID := proxyID(proxyKey)
				core.Validate(proxyType(proxyKey) == DynamicBody, "B2_PROXY_TYPE( proxyKey ) == b2_dynamicBody")

				treeAABB := dynamicTree.AABB(proxyID)
				shapeFatAABB := world.fatAABBs[shapeID]

				// Double check containment because a bullet can have multiple
				// shapes and maybe just one needs an update.
				if !treeAABB.Contains(shapeFatAABB) {
					dynamicTree.enlargeProxy(proxyID, shapeFatAABB)
				}

				shapeID = shape.nextShapeID
			}
		}

		world.profile.Bullets = timer.Milliseconds(bulletTicks)
	}

	// Need to free this even if no bullets got processed.

	stepContext.bulletBodies = nil
	stepContext.bulletBodyCount.Store(int32(0))
}

func sleepSolverIslands(world *World, awake *solverSet) {
	if !world.enableSleep {
		return
	}

	sleepTicks := timer.Ticks()

	// Collect split island candidate for the next time step. No need to split if sleeping is disabled.
	core.Assert(world.splitIslandID == nullIndex, "world->splitIslandId == B2_NULL_INDEX")

	splitSleepTimer := float32(0)

	for i := range world.workerCount {
		taskContext := &world.taskContexts[i]
		if taskContext.splitIslandID != nullIndex && taskContext.splitSleepTime >= splitSleepTimer {
			core.Assert(taskContext.splitSleepTime > 0.0, "taskContext->splitSleepTime > 0.0f")

			// Tie breaking for determinism. Largest island id wins. Needed due to work stealing.
			if taskContext.splitSleepTime == splitSleepTimer && taskContext.splitIslandID < world.splitIslandID {
				continue
			}

			world.splitIslandID = taskContext.splitIslandID
			splitSleepTimer = taskContext.splitSleepTime
		}
	}

	awakeIslandBitSet := &world.taskContexts[0].awakeIslandBitSet
	for i := 1; i < world.workerCount; i++ {
		awakeIslandBitSet.InPlaceUnion(&world.taskContexts[i].awakeIslandBitSet)
	}

	// Need to process in reverse because this moves islands to sleeping solver sets.
	islands := awake.islandSims

	count := len(awake.islandSims)
	for islandIndex := count - 1; islandIndex >= 0; islandIndex -= 1 {
		if awakeIslandBitSet.GetBit(uint32(islandIndex)) {
			// this island is still awake
			continue
		}

		island := &islands[islandIndex]
		islandID := island.islandID

		trySleepIsland(world, islandID)
	}

	validateSolverSets(world)

	world.profile.SleepIslands = timer.Milliseconds(sleepTicks)
}

func executeGraphBlock(stageType solverStageType, blockType solverBlockType, context *stepContext, block solverBlock, workerIndex int) {
	switch blockType {
	case graphContactBlock:
		switch stageType {
		case stageWarmStart:
			warmStartContactsWide(block, context)
		case stageSolve:
			pushContactsWide(block, context)
		case stageRelax:
			solveContactsWide(block, context)
		case stagePrepareJoints, stagePrepareContacts, stageIntegrateVelocities, stageIntegratePositions, stageRestitution, stageStoreImpulses:
			// These phases are handled by executeBlock rather than graph dispatch.
		}
	case graphJointBlock:
		switch stageType {
		case stageWarmStart:
			warmStartJointsTask(block, context)
		case stageSolve:
			useBias := true
			solveJointsTask(block, context, useBias, workerIndex)
		case stageRelax:
			useBias := false
			solveJointsTask(block, context, useBias, workerIndex)
		case stagePrepareJoints, stagePrepareContacts, stageIntegrateVelocities, stageIntegratePositions, stageRestitution, stageStoreImpulses:
			// These phases are handled by executeBlock rather than graph dispatch.
		}
	case bodyBlock, jointBlock, contactBlock:
		// Non-graph blocks have no graph constraint work.
	}
}

func reserveSolverContactScratch(world *World, colors []graphColor, wideContactCount int) []contactConstraintWide {
	if wideContactCount > cap(world.solverWideScratch) {
		world.solverWideScratch = slices.Grow(world.solverWideScratch, wideContactCount-len(world.solverWideScratch))[:wideContactCount]
	}

	clear(world.solverWideScratch[:wideContactCount])

	wideContactConstraints := world.solverWideScratch[:wideContactCount]

	overflow := &colors[overflowIndex]

	overflowCount := len(overflow.contactSims)
	if overflowCount > cap(world.solverOverflowScratch) {
		world.solverOverflowScratch = slices.Grow(world.solverOverflowScratch, overflowCount-len(world.solverOverflowScratch))[:overflowCount]
	}

	clear(world.solverOverflowScratch[:overflowCount])

	overflowContacts := world.solverOverflowScratch[:overflowCount]
	overflow.overflowConstraints = overflowContacts

	return wideContactConstraints
}

func prepareSolverConstraintSpans(colors []graphColor, activeColorCount int, activeColorIndices []int, wideContactConstraints []contactConstraintWide, wideContactCount, jointCount int, contactPrepareSpans []contactPrepareSpan, jointPrepareSpans []jointPrepareSpan) {
	// Distribute transient constraints to each graph color and prepare spans
	{
		wideBase := 0
		jointBase := 0

		for i := range activeColorCount {
			j := activeColorIndices[i]
			color := &colors[j]

			colorContactCount := len(color.contactSims)
			contactPrepareSpans[i].start = wideBase
			contactPrepareSpans[i].count = colorContactCount
			contactPrepareSpans[i].contacts = color.contactSims

			if colorContactCount == 0 {
				color.wideConstraints = nil
				color.wideConstraintCount = 0
			} else {
				color.wideConstraints = wideContactConstraints[wideBase:]

				colorContactCountW := ((colorContactCount - 1) >> simdShift) + 1
				color.wideConstraints = color.wideConstraints[:colorContactCountW]
				color.wideConstraintCount = colorContactCountW

				// Zero remainder lanes in the tail wide slot so prepare workers don't need to
				// initialize them.
				if (colorContactCount & (simdWidth - 1)) != 0 {
					color.wideConstraints[colorContactCountW-1] = contactConstraintWide{}
				}

				wideBase += colorContactCountW
			}

			jointPrepareSpans[i].start = jointBase
			jointPrepareSpans[i].count = len(color.jointSims)
			jointPrepareSpans[i].joints = color.jointSims
			jointBase += len(color.jointSims)
		}

		// Sentinel
		contactPrepareSpans[activeColorCount].start = wideContactCount
		contactPrepareSpans[activeColorCount].count = 0
		contactPrepareSpans[activeColorCount].contacts = nil

		core.Assert(wideBase == wideContactCount, "wideBase == wideContactCount")

		jointPrepareSpans[activeColorCount].start = jointCount
		jointPrepareSpans[activeColorCount].count = 0
		jointPrepareSpans[activeColorCount].joints = nil

		core.Assert(jointBase == jointCount, "jointBase == jointCount")
	}
}
