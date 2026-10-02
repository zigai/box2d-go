// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"math/bits"
	"runtime"
	"sync/atomic"
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

func pause() { runtime.Gosched() }

func integrateVelocitiesTask(block solverBlock, context *stepContext) {
	states := context.states
	sims := context.sims

	validate(block.startIndex+int(block.count) <= len(context.world.solverSets[awakeSet].bodyStates), "block.startIndex + block.count <= context->world->solverSets.data[b2_awakeSet].bodyStates.count")

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
		linearVelocityDelta := Add(MulSV(h*sim.invMass, sim.force), MulSV(h*gravityScale, gravity))
		angularVelocityDelta := float32(h * sim.invInertia * sim.torque)

		v = MulAdd(linearVelocityDelta, linearDamping, v)
		w = angularVelocityDelta + float32(angularDamping*w)

		state.linearVelocity = v
		state.angularVelocity = w
	}
}

func integratePositionsTask(block solverBlock, context *stepContext) {
	validate(block.startIndex+int(block.count) <= len(context.world.solverSets[awakeSet].bodyStates), "block.startIndex + block.count <= context->world->solverSets.data[b2_awakeSet].bodyStates.count")

	states := context.states
	h := context.h
	maxLinearSpeed := context.maxLinearVelocity
	maxAngularSpeed := MaxRotation * context.invDt
	maxLinearSpeedSquared := maxLinearSpeed * maxLinearSpeed
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

		// Clamp to max linear speed
		if Dot(v, v) > maxLinearSpeedSquared {
			ratio := maxLinearSpeed / Length(v)
			v = MulSV(ratio, v)
			state.flags |= isSpeedCapped
		}

		// Clamp to max angular speed
		if w*w > maxAngularSpeedSquared && (state.flags&allowFastRotation) == 0 {
			ratio := maxAngularSpeed / AbsFloat(w)
			w *= ratio
			state.flags |= isSpeedCapped
		}

		state.linearVelocity = v
		state.angularVelocity = w
		state.deltaPosition = MulAdd(state.deltaPosition, h, state.linearVelocity)
		state.deltaRotation = IntegrateRotation(state.deltaRotation, h*state.angularVelocity)
	}
}

func continuousQueryCallback(proxyID int, userData uint64, context *continuousContext) bool {
	shapeId := int(userData)
	continuousContext := context
	fastShape := continuousContext.fastShape
	fastBodySim := continuousContext.fastBodySim

	assert(fastShape.sensorIndex == nullIndex, "fastShape->sensorIndex == B2_NULL_INDEX")

	// Skip same shape
	if shapeId == fastShape.id {
		return true
	}

	world := continuousContext.world
	shape := &world.shapes[shapeId]

	// Skip same body
	if shape.bodyID == fastShape.bodyID {
		return true
	}

	// Skip sensors unless the shapes want sensor events
	isSensor := shape.sensorIndex != nullIndex
	if isSensor && (shape.enableSensorEvents == false || fastShape.enableSensorEvents == false) {
		return true
	}

	// Skip filtered shapes
	canCollide := shouldShapesCollide(fastShape.filter, shape.filter)
	if canCollide == false {
		return true
	}

	body := &world.bodies[shape.bodyID]

	bodySim := getBodySim(world, body)
	assert(body.bodyType == StaticBody || (fastBodySim.flags&isBullet) != 0, "body->type == b2_staticBody || ( fastBodySim->flags & b2_isBullet )")

	// Skip bullets
	// Warning: it is only safe to read flags from other bodies, not body sims because there are
	// body sim flag writes in the continuous solver.
	if (body.flags & isBullet) != 0 {
		return true
	}

	// Skip filtered bodies
	fastBody := &world.bodies[fastBodySim.bodyID]
	canCollide = shouldBodiesCollide(world, fastBody, body)
	if canCollide == false {
		return true
	}

	// Custom user filtering
	if shape.enableCustomFiltering || fastShape.enableCustomFiltering {
		customFilterFcn := world.customFilterFcn
		if customFilterFcn != nil {
			idA := Shape{index1: int32(shape.id + 1), world0: world.worldID, generation: shape.generation}
			idB := Shape{index1: int32(fastShape.id + 1), world0: world.worldID, generation: fastShape.generation}
			canCollide = customFilterFcn(idA, idB)
			if canCollide == false {
				return true
			}
		}
	}

	// Early out on fast parallel movement over a chain shape.
	if shape.shapeType == ChainSegmentShape {
		transform := ToRelativeTransform(bodySim.transform, continuousContext.base)
		p1 := TransformPoint(transform, shape.chainSegment.Segment.Point1)
		p2 := TransformPoint(transform, shape.chainSegment.Segment.Point2)
		e := Sub(p2, p1)
		var length float32
		e = GetLengthAndNormalize(&length, e)
		if length > LinearSlop() {
			c1 := continuousContext.centroid1
			separation1 := Cross(Sub(c1, p1), e)
			c2 := continuousContext.centroid2
			separation2 := Cross(Sub(c2, p1), e)

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
	} else {
		hitFraction := continuousContext.fraction
		didHit := false

		if 0.0 < output.Fraction && output.Fraction < continuousContext.fraction {
			hitFraction = output.Fraction
			didHit = true
		} else if 0.0 == output.Fraction {
			// fallback to TOI of a small circle around the fast shape centroid
			centroid := getShapeCentroid(fastShape)
			extent := computeShapeExtent(fastShape, centroid)
			radius := coreFraction * extent.minExtent
			input.ProxyB = MakeProxy([]Vec2{centroid}, radius)
			output = TimeOfImpact(&input)
			if 0.0 < output.Fraction && output.Fraction < continuousContext.fraction {
				hitFraction = output.Fraction
				didHit = true
			}
		}

		if didHit && (shape.enablePreSolveEvents || fastShape.enablePreSolveEvents) && world.preContinuousFcn != nil {
			shapeIdA := Shape{index1: int32(shape.id + 1), world0: world.worldID, generation: shape.generation}
			shapeIdB := Shape{index1: int32(fastShape.id + 1), world0: world.worldID, generation: fastShape.generation}

			// TOI runs in the base frame, lift the hit point back to world for the callback
			worldPoint := OffsetPos(continuousContext.base, output.Point)
			didHit = world.preContinuousFcn(shapeIdA, shapeIdB, worldPoint, output.Normal)
		}

		if didHit {
			continuousContext.fraction = hitFraction
		}
	}

	// Continue query
	return true
}

func solveContinuous(world *world, bodySimIndex int, taskContext *taskContext, dt float32) {
	awakeSet := &world.solverSets[awakeSet]
	fastBodySim := &awakeSet.bodySims[bodySimIndex]
	validate((fastBodySim.flags&isFast) != 0, "fastBodySim->flags & b2_isFast")

	// Re-center the sweep on the fast body so the TOI and the swept query stay in float precision
	base := fastBodySim.center0
	sweep := makeRelativeSweep(fastBodySim, base)

	var xf1 Transform
	xf1.Q = sweep.Q1
	xf1.P = Sub(sweep.C1, RotateVector(sweep.Q1, sweep.LocalCenter))

	var xf2 Transform
	xf2.Q = sweep.Q2
	xf2.P = Sub(sweep.C2, RotateVector(sweep.Q2, sweep.LocalCenter))

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

	shapeId := fastBody.headShapeID
	for shapeId != nullIndex {
		fastShape := &world.shapes[shapeId]
		shapeId = fastShape.nextShapeID

		context.fastShape = fastShape
		context.centroid1 = TransformPoint(xf1, fastShape.localCentroid)
		context.centroid2 = TransformPoint(xf2, fastShape.localCentroid)

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

		staticTree.Query(sweptBox, DefaultMaskBits, func(proxyID int, userData uint64) bool { return continuousQueryCallback(proxyID, userData, &context) })

		if isBullet {
			kinematicTree.Query(sweptBox, DefaultMaskBits, func(proxyID int, userData uint64) bool { return continuousQueryCallback(proxyID, userData, &context) })
			dynamicTree.Query(sweptBox, DefaultMaskBits, func(proxyID int, userData uint64) bool { return continuousQueryCallback(proxyID, userData, &context) })
		}
	}

	speculativeDistance := SpeculativeDistance()

	if context.fraction < 1.0 {
		// Handle time of impact event
		q := NLerp(sweep.Q1, sweep.Q2, context.fraction)
		c := Lerp(sweep.C1, sweep.C2, context.fraction)
		origin := Sub(c, RotateVector(q, sweep.LocalCenter))

		// Advance body, lifting the base frame result back to world
		fastBodySim.transform.Q = q
		fastBodySim.transform.P = OffsetPos(base, origin)
		fastBodySim.center = OffsetPos(base, c)
		fastBodySim.rotation0 = q
		fastBodySim.center0 = fastBodySim.center

		// Warning: writing to the body sim flags means we should not read from other body sim flags in this function.
		fastBodySim.flags |= hadTimeOfImpact

		// Timeloss means there is a lost gravity contribution. Other forces and torques are ignored for now.
		fastBodyState := &awakeSet.bodyStates[bodySimIndex]
		v := fastBodyState.linearVelocity
		timeLoss := (1.0 - context.fraction) * dt
		dv := MulSV(-timeLoss*fastBodySim.gravityScale, world.gravity)
		if fastBodyState.flags&lockLinearX != 0 {
			dv.X = 0
		}
		if fastBodyState.flags&lockLinearY != 0 {
			dv.Y = 0
		}
		fastBodyState.linearVelocity = Add(v, dv)

		// Update body move event
		event := &world.bodyMoveEvents[bodySimIndex]
		event.Transform = fastBodySim.transform

		// Prepare AABBs for broad-phase.
		// Even though a body is fast, it may not move much. So the AABB may not need enlargement.

		shapeId = fastBody.headShapeID
		for shapeId != nullIndex {
			shape := &world.shapes[shapeId]

			// Must recompute aabb at the interpolated transform
			aabb := computeFatShapeAABB(shape, fastBodySim.transform, speculativeDistance)
			shape.aabb = aabb

			shapeFatAABB := &world.fatAABBs[shapeId]
			if (*shapeFatAABB).Contains(aabb) == false {
				margin := shape.aabbMargin
				var fatAABB AABB
				fatAABB.LowerBound.X = aabb.LowerBound.X - margin
				fatAABB.LowerBound.Y = aabb.LowerBound.Y - margin
				fatAABB.UpperBound.X = aabb.UpperBound.X + margin
				fatAABB.UpperBound.Y = aabb.UpperBound.Y + margin
				*shapeFatAABB = fatAABB

				// Regular bodies mark the hierarchy as enlarged using atomic operations.
				// Bullets are handled separately at a later stage.
				if isBullet == false {
					world.broadPhase.markProxyMoved(shape.proxyKey, fatAABB)
				} else {
					fastBodySim.flags |= enlargeBulletBounds
				}
			}

			shapeId = shape.nextShapeID
		}
	} else {
		// No time of impact event

		// Advance body
		fastBodySim.rotation0 = fastBodySim.transform.Q
		fastBodySim.center0 = fastBodySim.center

		// Prepare AABBs for broad-phase
		shapeId = fastBody.headShapeID
		for shapeId != nullIndex {
			shape := &world.shapes[shapeId]

			// shape.aabb is still valid from above

			shapeFatAABB := &world.fatAABBs[shapeId]
			if (*shapeFatAABB).Contains(shape.aabb) == false {
				margin := shape.aabbMargin

				// Note: far from the origin the margin can be snapped to the nearest ULP.
				// This relevant for DP mode. So we lose the broad-phase hysteresis.
				var fatAABB AABB
				fatAABB.LowerBound.X = shape.aabb.LowerBound.X - margin
				fatAABB.LowerBound.Y = shape.aabb.LowerBound.Y - margin
				fatAABB.UpperBound.X = shape.aabb.UpperBound.X + margin
				fatAABB.UpperBound.Y = shape.aabb.UpperBound.Y + margin
				*shapeFatAABB = fatAABB

				if isBullet == false {
					world.broadPhase.markProxyMoved(shape.proxyKey, fatAABB)
				} else {
					fastBodySim.flags |= enlargeBulletBounds
				}
			}

			shapeId = shape.nextShapeID
		}
	}

	// Push sensor hits on the the task context for serial processing.
	for i := 0; i < context.sensorCount; i++ {
		// Skip any sensor hits that occurred after a solid hit
		if context.sensorFractions[i] < context.fraction {
			taskContext.sensorHits = arrayPush(taskContext.sensorHits, context.sensorHits[i])
		}
	}
}

func finalizeBodiesTask(startIndex int, endIndex int, workerIndex int, context *stepContext) {
	stepContext := context
	world := stepContext.world
	bodies := world.bodies
	states := stepContext.states
	sims := stepContext.sims

	assert(endIndex <= len(world.bodyMoveEvents), "endIndex <= world->bodyMoveEvents.count")

	sleepEnabled := world.enableSleep
	enableContinuous := world.enableContinuous
	timeStep := stepContext.dt
	invTimeStep := stepContext.invDt
	worldId := world.worldID

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

		if IsValidVec2(v) == false || IsValidFloat(w) == false {
			debugBody := &bodies[sim.bodyID]
			log("unstable: %s\n", bytes.TrimRight(debugBody.name[:], "\x00"))
		}

		assert(IsValidVec2(v), "b2IsValidVec2( v )")
		assert(IsValidFloat(w), "b2IsValidFloat( w )")

		sim.center = OffsetPos(sim.center, state.deltaPosition)
		sim.transform.Q = NormalizeRot(MulRot(state.deltaRotation, sim.transform.Q))

		// Use the velocity of the farthest point on the body to account for rotation.
		maxVelocity := Length(v) + float32(AbsFloat(w)*sim.maxExtent)

		// Sleep needs to observe position correction as well as true velocity.
		maxDeltaPosition := Length(state.deltaPosition) + float32(AbsFloat(state.deltaRotation.S)*sim.maxExtent)

		// Position correction is not as important for sleep as true velocity.
		positionSleepFactor := float32(0.5)
		sleepVelocity := MaxFloat(maxVelocity, positionSleepFactor*invTimeStep*maxDeltaPosition)

		// reset state deltas
		state.deltaPosition = Vec2{}
		state.deltaRotation = Rot{C: 1}

		sim.transform.P = OffsetPos(sim.center, Neg(RotateVector(sim.transform.Q, sim.localCenter)))

		// cache miss here, however I need the shape list below
		body := &bodies[sim.bodyID]
		body.bodyMoveIndex = simIndex
		moveEvents[simIndex].Transform = sim.transform
		moveEvents[simIndex].BodyID = Body{index1: int32(sim.bodyID + 1), world0: worldId, generation: body.generation}
		moveEvents[simIndex].UserData = body.userData
		moveEvents[simIndex].FellAsleep = false

		// reset applied force and torque
		sim.force = Vec2{}
		sim.torque = 0.0

		// If you hit this then it means you deferred mass computation but never called b2Body_UpdateMassFromShapes
		assert((body.flags&dirtyMass) == 0, "( body->flags & b2_dirtyMass ) == 0")

		// Clear the transient flags (fast, speed capped, had TOI). These flags are conditionally set
		// as part of the code below.
		body.flags &= ^bodyTransientFlags
		sim.flags &= ^(isFast | bodyTransientFlags)

		// The body state flag knows about speed capping (used for debug draw).
		body.flags |= (state.flags & isSpeedCapped)
		state.flags &= ^bodyTransientFlags

		if sleepEnabled == false || (body.flags&enableSleep) == 0 || sleepVelocity > body.sleepThreshold {
			// Body is not sleepy
			body.sleepTime = 0.0

			safetyFactor := body.safetyFactor
			maxMotion := MaxFloat(maxDeltaPosition, maxVelocity*timeStep)
			if body.bodyType == DynamicBody && enableContinuous && maxMotion > safetyFactor*sim.minExtent {
				// This flag is used for debug draw and contact recycling.
				sim.flags |= isFast

				// Store fast bullets for processing later.
				// This is deterministic because the order of TOI sweeps doesn't matter.
				if (sim.flags & isBullet) != 0 {
					bulletIndex := int(stepContext.bulletBodyCount.Add(int32(1))) - int(1)
					stepContext.bulletBodies[bulletIndex] = simIndex
				} else {
					solveContinuous(world, simIndex, taskContext, stepContext.dt)
				}
			} else {
				// Body is safe to advance
				sim.center0 = sim.center
				sim.rotation0 = sim.transform.Q
			}
		} else {
			// Body is safe to advance and is falling asleep
			sim.center0 = sim.center
			sim.rotation0 = sim.transform.Q
			body.sleepTime += timeStep
		}

		// Any single body in an island can keep it awake
		island := &world.islands[body.islandID]
		if body.sleepTime < TimeToSleep {
			// keep island awake
			islandIndex := island.localIndex
			setBit(awakeIslandBitSet, uint32(islandIndex))
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

		// Update shapes AABBs
		transform := sim.transform
		isFast := (sim.flags & isFast) != 0
		shapeId := body.headShapeID
		for shapeId != nullIndex {
			shape := &world.shapes[shapeId]

			if isFast {
				// For fast non-bullet bodies the AABB has already been updated in solveContinuous
				// For fast bullet bodies the AABB will be updated at a later stage

				// Add to enlarged shapes regardless of AABB changes.
				// Bit-set to keep the move array sorted
				setBit(enlargedSimBitSet, uint32(simIndex))
			} else {
				// Update bounds for slow bodies, including slow bullets.

				aabb := computeFatShapeAABB(shape, transform, speculativeDistance)
				shape.aabb = aabb

				shapeFatAABB := &world.fatAABBs[shapeId]
				if (*shapeFatAABB).Contains(aabb) == false {
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
					setBit(enlargedSimBitSet, uint32(simIndex))
				}
			}

			shapeId = shape.nextShapeID
		}
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

	assert(dim.count >= 1, "dim.count >= 1")
	assert(dim.size*dim.count >= itemCount, "dim.size * dim.count >= itemCount")

	return dim
}

func initBlocks(blocks []syncBlock, dim blockDim, itemCount int, blockType uint8, colorIndex uint8) {
	if dim.count == 0 {
		return
	}

	assert(itemCount >= dim.count, "itemCount >= dim.count")

	// Compute the number of elements per block
	blockSize := dim.size

	// Simulation too big
	assert(blockSize <= 65535, "blockSize <= UINT16_MAX")

	for i := 0; i < dim.count; i++ {
		blocks[i].block.startIndex = i * blockSize
		blocks[i].block.count = uint16(blockSize)
		blocks[i].block.blockType = blockType
		blocks[i].block.colorIndex = colorIndex
		blocks[i].syncIndex.Store(int32(0))
	}

	// The last block may not be full
	blocks[dim.count-1].block.count = uint16(itemCount - (dim.count-1)*blockSize)

	validate(int(blocks[dim.count-1].block.count) <= blockSize, "blocks[dim.count - 1].block.count <= blockSize")
	validate((dim.count-1)*dim.size+int(blocks[dim.count-1].block.count) == itemCount, "( dim.count - 1 ) * dim.size + blocks[dim.count - 1].block.count == itemCount")
}

func initStage(stage []solverStage, stageType solverStageType, blocks []syncBlock, blockCount int, colorIndex uint8) []solverStage {
	stage[0].stageType = stageType
	stage[0].blocks = blocks
	stage[0].blockCount = blockCount
	stage[0].colorIndex = colorIndex
	stage[0].completionCount.Store(0)
	return stage[1:]
}

func initColorStages(stage []solverStage, stageType solverStageType, iterations int, activeColorCount int, colorBlocks [][]syncBlock, colorBlockCounts []int, activeColorIndices []int) []solverStage {
	for j := 0; j < iterations; j++ {
		for i := 0; i < activeColorCount; i++ {
			stage = initStage(stage, stageType, colorBlocks[i], colorBlockCounts[i], uint8(activeColorIndices[i]))
		}
	}
	return stage
}

func executeBlock(stage *solverStage, context *stepContext, block solverBlock, workerIndex int) {
	stageType := stage.stageType
	blockType := solverBlockType(block.blockType)

	switch stageType {
	case stagePrepareJoints:
		prepareJointsTask(block, context)
		break

	case stagePrepareContacts:
		prepareContactsWide(block, context)
		break

	case stageIntegrateVelocities:
		integrateVelocitiesTask(block, context)
		break

	case stageWarmStart:
		if blockType == graphContactBlock {
			warmStartContactsWide(block, context)
		} else if blockType == graphJointBlock {
			warmStartJointsTask(block, context)
		}
		break

	case stageSolve:
		if blockType == graphContactBlock {
			pushContactsWide(block, context)
		} else if blockType == graphJointBlock {
			useBias := true
			solveJointsTask(block, context, useBias, workerIndex)
		}
		break

	case stageIntegratePositions:
		integratePositionsTask(block, context)
		break

	case stageRelax:
		if blockType == graphContactBlock {
			solveContactsWide(block, context)
		} else if blockType == graphJointBlock {
			useBias := false
			solveJointsTask(block, context, useBias, workerIndex)
		}
		break

	case stageRestitution:
		if blockType == graphContactBlock {
			applyRestitutionWide(block, context)
		}
		break

	case stageStoreImpulses:
		storeImpulsesWide(block, context, workerIndex)
		break
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

	assert(0 <= startIndex && startIndex < blockCount, "0 <= startIndex && startIndex < blockCount")

	blockIndex := startIndex
	for i := 0; i < blockCount; i++ {
		if blocks[blockIndex].syncIndex.CompareAndSwap(int32(previousSyncIndex), int32(syncIndex)) {
			assert(stage.stageType != stagePrepareContacts || syncIndex < 2, "stage->type != b2_stagePrepareContacts || syncIndex < 2")
			assert(completedCount < blockCount, "completedCount < blockCount")

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
		assert(syncIndex > 0, "syncIndex > 0")
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
		if context.mainClaimed.CompareAndSwap(int32(0), int32(1)) == false {
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

		ticks := GetTicks()

		bodySyncIndex := 1
		stageIndex := 0

		// Prepare joint constraints
		jointSyncIndex := uint32(1)
		syncBits := (uint32(jointSyncIndex) << 16) | uint32(stageIndex)
		assert(stages[stageIndex].stageType == stagePrepareJoints, "stages[stageIndex].type == b2_stagePrepareJoints")
		executeMainStage(&stages[stageIndex], context, syncBits)
		stageIndex += 1
		jointSyncIndex += 1

		// Prepare contact constraints
		contactSyncIndex := uint32(1)
		syncBits = (uint32(contactSyncIndex) << 16) | uint32(stageIndex)
		assert(stages[stageIndex].stageType == stagePrepareContacts, "stages[stageIndex].type == b2_stagePrepareContacts")
		executeMainStage(&stages[stageIndex], context, syncBits)
		stageIndex += 1
		contactSyncIndex += 1

		// Single-threaded overflow work. These constraints don't fit in the graph coloring.
		prepareJointsOverflow(context)
		prepareContactsOverflow(context)

		profile.PrepareConstraints += GetMillisecondsAndReset(&ticks)

		graphSyncIndex := 1
		subStepCount := context.subStepCount
		for subStepIndex := 0; subStepIndex < subStepCount; subStepIndex++ {
			// stage index restarted each iteration
			// syncBits still increases monotonically because the upper bits increase each iteration
			iterationStageIndex := stageIndex

			// Integrate velocities
			syncBits = (uint32(bodySyncIndex) << 16) | uint32(iterationStageIndex)
			assert(stages[iterationStageIndex].stageType == stageIntegrateVelocities, "stages[iterationStageIndex].type == b2_stageIntegrateVelocities")
			executeMainStage(&stages[iterationStageIndex], context, syncBits)
			iterationStageIndex += 1
			bodySyncIndex += 1

			profile.IntegrateVelocities += GetMillisecondsAndReset(&ticks)

			// Warm start constraints
			warmStartJointsOverflow(context)
			warmStartContactsOverflow(context)

			for colorIndex := 0; colorIndex < activeColorCount; colorIndex++ {
				syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
				assert(stages[iterationStageIndex].stageType == stageWarmStart, "stages[iterationStageIndex].type == b2_stageWarmStart")
				executeMainStage(&stages[iterationStageIndex], context, syncBits)
				iterationStageIndex += 1
			}
			graphSyncIndex += 1

			profile.WarmStart += GetMillisecondsAndReset(&ticks)

			// Solve constraints
			useBias := true
			for j := 0; j < iterations; j++ {
				// Overflow constraints have lower priority. Typically these are dynamic-vs-dynamic.
				solveJointsOverflow(context, useBias)
				pushContactsOverflow(context)

				for colorIndex := 0; colorIndex < activeColorCount; colorIndex++ {
					syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
					assert(stages[iterationStageIndex].stageType == stageSolve, "stages[iterationStageIndex].type == b2_stageSolve")
					executeMainStage(&stages[iterationStageIndex], context, syncBits)
					iterationStageIndex += 1
				}
				graphSyncIndex += 1
			}

			profile.SolveImpulses += GetMillisecondsAndReset(&ticks)

			// Integrate positions
			assert(stages[iterationStageIndex].stageType == stageIntegratePositions, "stages[iterationStageIndex].type == b2_stageIntegratePositions")
			syncBits = (uint32(bodySyncIndex) << 16) | uint32(iterationStageIndex)
			executeMainStage(&stages[iterationStageIndex], context, syncBits)
			iterationStageIndex += 1
			bodySyncIndex += 1

			profile.IntegratePositions += GetMillisecondsAndReset(&ticks)

			// Relax constraints
			useBias = false
			for j := 0; j < relaxIterations; j++ {
				solveJointsOverflow(context, useBias)
				solveContactsOverflow(context)
				for colorIndex := 0; colorIndex < activeColorCount; colorIndex++ {
					syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
					assert(stages[iterationStageIndex].stageType == stageRelax, "stages[iterationStageIndex].type == b2_stageRelax")
					executeMainStage(&stages[iterationStageIndex], context, syncBits)
					iterationStageIndex += 1
				}
				graphSyncIndex += 1
			}

			profile.RelaxImpulses += GetMillisecondsAndReset(&ticks)
		}

		// Advance the stage according to the sub-stepping tasks just completed
		// integrate velocities / warm start / solve / integrate positions / relax
		stageIndex += 1 + activeColorCount + iterations*activeColorCount + 1 + relaxIterations*activeColorCount

		restitutionIterations := context.world.restitutionIterations
		if restitutionIterations > 0 && int(context.anyRestitution.Load()) != 0 {
			for j := 0; j < restitutionIterations; j++ {
				applyRestitutionOverflow(context)

				iterationStageIndex := stageIndex
				for colorIndex := 0; colorIndex < activeColorCount; colorIndex++ {
					syncBits = (uint32(graphSyncIndex) << 16) | uint32(iterationStageIndex)
					assert(stages[iterationStageIndex].stageType == stageRestitution, "stages[iterationStageIndex].type == b2_stageRestitution")
					executeMainStage(&stages[iterationStageIndex], context, syncBits)
					iterationStageIndex += 1
				}
				graphSyncIndex += 1
			}

			profile.Restitution += GetMillisecondsAndReset(&ticks)
		}

		stageIndex += activeColorCount

		// Store impulses
		storeImpulsesOverflow(context)

		syncBits = (uint32(contactSyncIndex) << 16) | uint32(stageIndex)
		assert(stages[stageIndex].stageType == stageStoreImpulses, "stages[stageIndex].type == b2_stageStoreImpulses")
		executeMainStage(&stages[stageIndex], context, syncBits)

		profile.StoreImpulses += GetMillisecondsAndReset(&ticks)

		// Signal workers to finish
		context.atomicSyncBits.Store(^uint32(0))

		assert(stageIndex+1 == context.stageCount, "stageIndex + 1 == context->stageCount")
		return
	}

	// Worker spins and waits for work
	lastSyncBits := uint32(0)
	// maxSpinTime := 10
	for true {
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
				Yield()
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
		assert(stageIndex < context.stageCount, "stageIndex < context->stageCount")

		syncIndex := int((syncBits >> 16) & 0xFFFF)
		assert(syncIndex > 0, "syncIndex > 0")

		previousSyncIndex := syncIndex - 1

		stage := &stages[stageIndex]
		executeStage(stage, context, previousSyncIndex, syncIndex, workerIndex)

		lastSyncBits = syncBits
	}
}

func bulletBodyTask(startIndex int, endIndex int, workerIndex int, context *stepContext) {
	stepContext := context
	taskContext := &stepContext.world.taskContexts[workerIndex]

	assert(startIndex <= endIndex, "startIndex <= endIndex")

	for i := startIndex; i < endIndex; i++ {
		simIndex := stepContext.bulletBodies[i]
		solveContinuous(stepContext.world, simIndex, taskContext, stepContext.dt)
	}
}

func solve(world *world, stepContext *stepContext) {
	// Only count steps that advance the simulation
	world.stepIndex += 1

	// Are there any awake bodies? This scenario should not be important for profiling.
	awake := &world.solverSets[awakeSet]
	awakeBodyCount := len(awake.bodySims)
	if awakeBodyCount == 0 {
		return
	}

	// Solve constraints using graph coloring
	{

		setupTicks := GetTicks()

		// Prepare buffers for bullets
		stepContext.bulletBodyCount.Store(int32(0))
		stepContext.bulletBodies = make([]int, awakeBodyCount)

		graph := &world.constraintGraph
		colors := graph.colors[:]

		stepContext.sims = awake.bodySims
		stepContext.states = awake.bodyStates

		// count contacts, joints, and colors
		activeColorCount := 0
		for i := 0; i < GraphColorCount-1; i++ {
			perColorContactCount := len(colors[i].contactSims)
			perColorJointCount := len(colors[i].jointSims)
			occupancyCount := perColorContactCount + perColorJointCount
			if occupancyCount > 0 {
				activeColorCount += 1
			}
		}

		// prepare for move events
		world.bodyMoveEvents = arrayReserve(world.bodyMoveEvents, awakeBodyCount)
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
		var activeColorIndices [GraphColorCount]int
		var colorContactCounts [GraphColorCount]int
		var colorJointCounts [GraphColorCount]int
		var graphContactDims [GraphColorCount]blockDim
		var graphJointDims [GraphColorCount]blockDim
		graphBlockCount := 0

		// c is the active color index
		wideContactCount := 0
		jointCount := 0
		c := 0
		for i := 0; i < GraphColorCount-1; i++ {
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
		activeColorCount = c

		// Prepare and store run as one flat parallel-for over the entire wide constraint range,
		// partitioned into uniformly sized blocks. Color info is consulted inside the task via
		// a small span array, so blocks do not need to honor color boundaries here.
		contactPrepareDim := computeBlockCount(wideContactCount, minContactsPerBlock, maxBlockCount)
		jointPrepareDim := computeBlockCount(jointCount, minJointsPerBlock, maxBlockCount)

		wideContactConstraints := make([]contactConstraintWide, wideContactCount)

		overflow := &colors[overflowIndex]
		overflowCount := len(overflow.contactSims)
		overflowContacts := make([]contactConstraint, overflowCount)
		overflow.overflowConstraints = overflowContacts

		// Build the span table for the flat prepare/store parallel-for while I slice the
		// wide constraint buffer across colors. One entry per active color plus a sentinel
		// at wideContactCount.
		var contactPrepareSpans [GraphColorCount + 1]contactPrepareSpan
		var jointPrepareSpans [GraphColorCount + 1]jointPrepareSpan

		// Distribute transient constraints to each graph color and prepare spans
		{
			wideBase := 0
			jointBase := 0
			for i := 0; i < activeColorCount; i++ {
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
			assert(wideBase == wideContactCount, "wideBase == wideContactCount")

			jointPrepareSpans[activeColorCount].start = jointCount
			jointPrepareSpans[activeColorCount].count = 0
			jointPrepareSpans[activeColorCount].joints = nil
			assert(jointBase == jointCount, "jointBase == jointCount")
		}

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

		stages := make([]solverStage, stageCount)
		bodyBlocks := make([]syncBlock, bodyDim.count)
		contactBlocks := make([]syncBlock, contactPrepareDim.count)
		jointBlocks := make([]syncBlock, jointPrepareDim.count)
		graphBlocks := make([]syncBlock, graphBlockCount)

		// Split an awake island. This modifies:
		// - stack allocator
		// - world island array and solver set
		// - island indices on bodies, contacts, and joints
		// I'm squeezing this task in here because it may be expensive and this is a safe place to put it.
		// Note: cannot split islands in parallel with FinalizeBodies
		var splitTask func()
		if world.splitIslandID != nullIndex {
			if world.taskCount < MaxTasks {
				splitTask = world.enqueueTaskFcn(func() { splitIslandTask(world) })
				world.taskCount += 1
				if splitTask != nil {
					world.activeTaskCount += 1
				}
			} else {
				splitIslandTask(world)
			}
		}

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
		for i := 0; i < activeColorCount; i++ {
			graphColorBlocks[i] = baseGraphBlock

			colorIndex := uint8(activeColorIndices[i])
			initBlocks(baseGraphBlock, graphJointDims[i], colorJointCounts[i], uint8(graphJointBlock), colorIndex)
			baseGraphBlock = baseGraphBlock[graphJointDims[i].count:]

			initBlocks(baseGraphBlock, graphContactDims[i], colorContactCounts[i], uint8(graphContactBlock), colorIndex)
			baseGraphBlock = baseGraphBlock[graphContactDims[i].count:]

			graphBlockCounts[i] = graphJointDims[i].count + graphContactDims[i].count
		}

		assert(len(graphBlocks)-len(baseGraphBlock) == graphBlockCount, "(ptrdiff_t)( baseGraphBlock - graphBlocks ) == graphBlockCount")

		stage := stages
		stage = initStage(stage, stagePrepareJoints, jointBlocks, jointPrepareDim.count, 255)
		stage = initStage(stage, stagePrepareContacts, contactBlocks, contactPrepareDim.count, 255)
		stage = initStage(stage, stageIntegrateVelocities, bodyBlocks, bodyDim.count, 255)
		stage = initColorStages(stage, stageWarmStart, 1, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initColorStages(stage, stageSolve, iterations, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initStage(stage, stageIntegratePositions, bodyBlocks, bodyDim.count, 255)
		stage = initColorStages(stage, stageRelax, relaxIterations, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initColorStages(stage, stageRestitution, 1, activeColorCount, graphColorBlocks[:], graphBlockCounts[:],
			activeColorIndices[:])
		stage = initStage(stage, stageStoreImpulses, contactBlocks, contactPrepareDim.count, 255)

		assert(len(stages)-len(stage) == stageCount, "(int)( stage - stages ) == stageCount")

		assert(workerCount <= MaxWorkers, "workerCount <= B2_MAX_WORKERS")
		var workerContexts [MaxWorkers]workerContext

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

		world.profile.SolverSetup = GetMillisecondsAndReset(&setupTicks)

		constraintTicks := GetTicks()

		jointIdCapacity := getIDCapacity(&world.jointIDPool)
		contactIdCapacity := getIDCapacity(&world.contactIDPool)
		for i := 0; i < workerCount; i++ {
			taskContext := &world.taskContexts[i]
			setBitCountAndClear(&taskContext.jointStateBitSet, uint32(jointIdCapacity))
			setBitCountAndClear(&taskContext.hitEventBitSet, uint32(contactIdCapacity))
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
		for i := 0; i < workerCount; i++ {
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

		world.profile.Constraints = GetMillisecondsAndReset(&constraintTicks)

		transformTicks := GetTicks()

		// Prepare contact, enlarged body, and island bit sets used in body finalization.
		awakeIslandCount := len(awake.islandSims)
		for i := 0; i < world.workerCount; i++ {
			taskContext := &world.taskContexts[i]
			taskContext.sensorHits = taskContext.sensorHits[:0]
			setBitCountAndClear(&taskContext.enlargedSimBitSet, uint32(awakeBodyCount))
			setBitCountAndClear(&taskContext.awakeIslandBitSet, uint32(awakeIslandCount))
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

		world.profile.Transforms = GetMilliseconds(transformTicks)

	}

	// Report joint events
	{

		jointEventTicks := GetTicks()

		// Gather bits for all joints that have force/torque events
		jointStateBitSet := &world.taskContexts[0].jointStateBitSet
		for i := 1; i < world.workerCount; i++ {
			inPlaceUnion(jointStateBitSet, &world.taskContexts[i].jointStateBitSet)
		}
		{
			wordCount := jointStateBitSet.blockCount
			words := jointStateBitSet.bits

			jointArray := world.joints
			worldIndex0 := world.worldID

			for k := uint32(0); k < wordCount; k++ {
				word := words[k]
				for word != 0 {
					ctz := bits.TrailingZeros64(word)
					jointId := int(64*k) + ctz

					assert(jointId < cap(world.joints), "jointId < world->joints.capacity")

					joint := &jointArray[jointId]

					assert(joint.setIndex == awakeSet, "joint->setIndex == b2_awakeSet")

					event := JointEvent{JointID: Joint{index1: int32(jointId + 1), world0: worldIndex0, generation: joint.generation}, UserData: joint.userData}

					world.jointEvents = arrayPush(world.jointEvents, event)

					// Clear the smallest set bit
					word = word & (word - 1)
				}
			}
		}

		world.profile.JointEvents = GetMilliseconds(jointEventTicks)

	}

	// Report hit events
	{

		hitTicks := GetTicks()

		assert(len(world.contactHitEvents) == 0, "world->contactHitEvents.count == 0")

		// Fast path: skip if no worker flagged any hit-event candidates during impulse storing.
		anyHitEvents := false
		for i := 0; i < world.workerCount; i++ {
			if world.taskContexts[i].hasHitEvents {
				anyHitEvents = true
				break
			}
		}

		if anyHitEvents {
			// Union per-worker bits into worker 0's bit set.
			hitEventBitSet := &world.taskContexts[0].hitEventBitSet
			for i := 1; i < world.workerCount; i++ {
				if world.taskContexts[i].hasHitEvents {
					inPlaceUnion(hitEventBitSet, &world.taskContexts[i].hitEventBitSet)
				}
			}

			threshold := world.hitEventThreshold
			colors := world.constraintGraph.colors[:]
			contactArray := world.contacts
			shapeArray := world.shapes
			worldId := world.worldID

			wordCount := hitEventBitSet.blockCount
			words := hitEventBitSet.bits
			for k := uint32(0); k < wordCount; k++ {
				word := words[k]
				for word != 0 {
					ctz := bits.TrailingZeros64(word)
					contactId := int(64*k) + ctz

					contact := &contactArray[contactId]
					assert(contact.setIndex == awakeSet && contact.colorIndex != nullIndex, "contact->setIndex == b2_awakeSet && contact->colorIndex != B2_NULL_INDEX")

					color := &colors[contact.colorIndex]
					contactSim := &color.contactSims[contact.localIndex]

					event := ContactHitEvent{}
					event.ApproachSpeed = threshold

					var bestPoint *ManifoldPoint
					pointCount := contactSim.manifold.PointCount
					for p := 0; p < pointCount; p++ {
						mp := &contactSim.manifold.Points[p]
						approachSpeed := -mp.NormalVelocity

						// Need to check total impulse because the point may be speculative and not colliding
						if approachSpeed > event.ApproachSpeed && mp.TotalNormalImpulse > 0.0 {
							event.ApproachSpeed = approachSpeed
							bestPoint = mp
						}
					}

					validate(bestPoint != nil, "bestPoint != NULL")

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
							event.Point = OffsetPos(bodySimB.center, bestPoint.AnchorB)
						} else {
							bodySimA := getBodySim(world, bodyA)
							event.Point = OffsetPos(bodySimA.center, bestPoint.AnchorA)
						}

						event.ShapeIDA = Shape{index1: int32(shapeA.id + 1), world0: worldId, generation: shapeA.generation}
						event.ShapeIDB = Shape{index1: int32(shapeB.id + 1), world0: worldId, generation: shapeB.generation}

						event.ContactID = Contact{
							index1:     int32(contact.contactID + 1),
							world0:     worldId,
							padding:    0,
							generation: contact.generation,
						}

						world.contactHitEvents = arrayPush(world.contactHitEvents, event)
					}

					// Clear the smallest set bit
					word = word & (word - 1)
				}
			}
		}

		world.profile.HitEvents = GetMilliseconds(hitTicks)

	}
	{

		refitTicks := GetTicks()

		bp := &world.broadPhase
		bp.trees[KinematicBody].refit()
		bp.trees[DynamicBody].refit()

		validateBroadphase(&world.broadPhase)

		world.profile.Refit = GetMilliseconds(refitTicks)

	}

	// Bullets are processed after the broad-phase refit so they can query the
	// final non-bullet world.
	bulletBodyCount := int(stepContext.bulletBodyCount.Load())
	if bulletBodyCount > 0 {

		bulletTicks := GetTicks()

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
		for i := 0; i < bulletBodyCount; i++ {
			bulletBodySim := &bodySimArray[bulletBodySimIndices[i]]

			// It is worth tracking this flag because bullets may be moving slowly.
			if (bulletBodySim.flags & enlargeBulletBounds) == 0 {
				continue
			}

			// Clear flag
			bulletBodySim.flags &= ^enlargeBulletBounds

			bodyId := bulletBodySim.bodyID
			assert(0 <= bodyId && bodyId < len(world.bodies), "0 <= bodyId && bodyId < world->bodies.count")
			bulletBody := &bodyArray[bodyId]

			shapeId := bulletBody.headShapeID
			for shapeId != nullIndex {
				shape := &shapeArray[shapeId]
				proxyKey := shape.proxyKey
				proxyId := proxyID(proxyKey)
				validate(proxyType(proxyKey) == DynamicBody, "B2_PROXY_TYPE( proxyKey ) == b2_dynamicBody")

				treeAABB := dynamicTree.AABB(proxyId)
				shapeFatAABB := world.fatAABBs[shapeId]

				// Double check containment because a bullet can have multiple
				// shapes and maybe just one needs an update.
				if (treeAABB).Contains(shapeFatAABB) == false {
					dynamicTree.enlargeProxy(proxyId, shapeFatAABB)
				}

				shapeId = shape.nextShapeID
			}
		}

		world.profile.Bullets = GetMilliseconds(bulletTicks)

	}

	// Need to free this even if no bullets got processed.

	stepContext.bulletBodies = nil
	stepContext.bulletBodyCount.Store(int32(0))

	// Report sensor hits. This may include bullets sensor hits.
	{

		sensorHitTicks := GetTicks()

		workerCount := world.workerCount
		assert(workerCount == len(world.taskContexts), "workerCount == world->taskContexts.count")

		for i := 0; i < workerCount; i++ {
			taskContext := &world.taskContexts[i]
			hitCount := len(taskContext.sensorHits)
			hits := taskContext.sensorHits

			for j := 0; j < hitCount; j++ {
				hit := hits[j]
				sensorShape := &world.shapes[hit.sensorID]
				visitingShape := &world.shapes[hit.visitorID]

				sensor := &world.sensors[sensorShape.sensorIndex]
				shapeRef := visitor{shapeID: hit.visitorID, generation: visitingShape.generation}
				sensor.hits = arrayPush(sensor.hits, shapeRef)
			}
		}

		world.profile.SensorHits = GetMilliseconds(sensorHitTicks)

	}

	// Island sleeping
	// This must be done last because putting islands to sleep invalidates the enlarged body bits.
	// todo_erin figure out how to do this in parallel with tree refit
	if world.enableSleep == true {

		sleepTicks := GetTicks()

		// Collect split island candidate for the next time step. No need to split if sleeping is disabled.
		assert(world.splitIslandID == nullIndex, "world->splitIslandId == B2_NULL_INDEX")
		splitSleepTimer := float32(0)
		for i := 0; i < world.workerCount; i++ {
			taskContext := &world.taskContexts[i]
			if taskContext.splitIslandID != nullIndex && taskContext.splitSleepTime >= splitSleepTimer {
				assert(taskContext.splitSleepTime > 0.0, "taskContext->splitSleepTime > 0.0f")

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
			inPlaceUnion(awakeIslandBitSet, &world.taskContexts[i].awakeIslandBitSet)
		}

		// Need to process in reverse because this moves islands to sleeping solver sets.
		islands := awake.islandSims
		count := len(awake.islandSims)
		for islandIndex := count - 1; islandIndex >= 0; islandIndex -= 1 {
			if getBit(awakeIslandBitSet, uint32(islandIndex)) == true {
				// this island is still awake
				continue
			}

			island := &islands[islandIndex]
			islandID := island.islandID

			trySleepIsland(world, islandID)
		}

		validateSolverSets(world)

		world.profile.SleepIslands = GetMilliseconds(sleepTicks)

	}
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
