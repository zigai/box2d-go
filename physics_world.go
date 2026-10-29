// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"iter"
	"log/slog"
	"math"
	"math/bits"

	"github.com/zigai/box2d/internal/array"
	"github.com/zigai/box2d/internal/bitset"
	"github.com/zigai/box2d/internal/core"
	"github.com/zigai/box2d/internal/idpool"
	"github.com/zigai/box2d/internal/scheduler"
	"github.com/zigai/box2d/internal/timer"
)

type (
	EnqueueTaskFunc   func(task func()) (finish func())
	FrictionFunc      func(frictionA float32, userMaterialIDA uint64, frictionB float32, userMaterialIDB uint64) float32
	RestitutionFunc   func(restitutionA float32, userMaterialIDA uint64, restitutionB float32, userMaterialIDB uint64) float32
	CustomFilterFunc  func(shapeA, shapeB Shape) bool
	PreSolveFunc      func(shapeA, shapeB Shape, manifold *Manifold)
	PreContinuousFunc func(shapeA, shapeB Shape, point Pos, normal Vec2) bool
)

type Capacity struct {
	StaticShapeCount  int
	DynamicShapeCount int
	StaticBodyCount   int
	DynamicBodyCount  int
	ContactCount      int
}

type WorldDef struct {
	Gravity                      Vec2
	RestitutionThreshold         float32
	RestitutionIterations        int
	EnableRestitutionPropagation bool
	HitEventThreshold            float32
	ContactHertz                 float32
	ContactDampingRatio          float32
	ContactSpeed                 float32
	MaximumLinearSpeed           float32
	FrictionCallback             FrictionFunc
	RestitutionCallback          RestitutionFunc
	EnableSleep                  bool
	EnableContinuous             bool
	WorkerCount                  int
	EnqueueTask                  EnqueueTaskFunc
	UserData                     any
	Capacity                     Capacity
	// Logger receives simulation warnings. DefaultWorldDef sets it to slog.Default().
	Logger        *slog.Logger
	internalValue int32
}

type Profile struct {
	Step                float32
	Pairs               float32
	Collide             float32
	Solve               float32
	SolverSetup         float32
	Constraints         float32
	PrepareConstraints  float32
	IntegrateVelocities float32
	WarmStart           float32
	SolveImpulses       float32
	IntegratePositions  float32
	RelaxImpulses       float32
	Restitution         float32
	StoreImpulses       float32
	SplitIslands        float32
	Transforms          float32
	SensorHits          float32
	JointEvents         float32
	HitEvents           float32
	Refit               float32
	Bullets             float32
	SleepIslands        float32
	Sensors             float32
}

type Counters struct {
	ByteCount            int64
	BodyCount            int
	ShapeCount           int
	ContactCount         int
	JointCount           int
	IslandCount          int
	StackUsed            int
	StaticTreeHeight     int
	TreeHeight           int
	TaskCount            int
	ColorCounts          [24]int
	AwakeContactCount    int
	RecycledContactCount int
}

type taskContext struct {
	sensorHits           []sensorHit
	pairKeys             []uint64
	contactStateBitSet   bitset.Set
	hitEventBitSet       bitset.Set
	hasHitEvents         bool
	jointStateBitSet     bitset.Set
	enlargedSimBitSet    bitset.Set
	awakeIslandBitSet    bitset.Set
	splitSleepTime       float32
	splitIslandID        int
	recycledContactCount int
}

// World is a physics simulation. Create one with NewWorld and release it with Close.
type World struct {
	broadPhase                   broadPhase
	constraintGraph              constraintGraph
	solverWideScratch            []contactConstraintWide
	solverOverflowScratch        []contactConstraint
	stepScratch                  stepContext
	collideSpanScratch           [GraphColorCount + 2]contactCollideSpan
	contactPrepareSpanScratch    [GraphColorCount + 1]contactPrepareSpan
	jointPrepareSpanScratch      [GraphColorCount + 1]jointPrepareSpan
	solverStageScratch           []solverStage
	solverBodyBlockScratch       []syncBlock
	solverContactBlockScratch    []syncBlock
	solverJointBlockScratch      []syncBlock
	solverGraphBlockScratch      []syncBlock
	solverWorkerScratch          [MaxWorkers]workerContext
	solverBulletScratch          []int
	parallelForScratch           parallelForShared
	parallelTaskScratch          [MaxWorkers]parallelForTask
	parallelHandleScratch        [MaxWorkers]func()
	bodyIDPool                   idpool.Pool
	bodies                       []body
	solverSetIDPool              idpool.Pool
	solverSets                   []solverSet
	jointIDPool                  idpool.Pool
	joints                       []joint
	contactIDPool                idpool.Pool
	contacts                     []contact
	islandIDPool                 idpool.Pool
	islands                      []island
	shapeIDPool                  idpool.Pool
	chainIDPool                  idpool.Pool
	shapes                       []shape
	fatAABBs                     []AABB
	chainShapes                  []chainShape
	sensors                      []sensor
	taskContexts                 []taskContext
	sensorTaskContexts           []sensorTaskContext
	bodyMoveEvents               []BodyMoveEvent
	sensorBeginEvents            []SensorBeginTouchEvent
	contactBeginEvents           []ContactBeginTouchEvent
	sensorEndEvents              [2][]SensorEndTouchEvent
	contactEndEvents             [2][]ContactEndTouchEvent
	endEventArrayIndex           int
	contactHitEvents             []ContactHitEvent
	jointEvents                  []JointEvent
	debugBodySet                 bitset.Set
	debugJointSet                bitset.Set
	debugContactSet              bitset.Set
	debugIslandSet               bitset.Set
	stepIndex                    uint64
	splitIslandID                int
	gravity                      Vec2
	hitEventThreshold            float32
	restitutionThreshold         float32
	restitutionIterations        int
	enableRestitutionPropagation bool
	maxLinearSpeed               float32
	contactSpeed                 float32
	contactHertz                 float32
	contactDampingRatio          float32
	contactRecycleDistance       float32
	frictionCallback             FrictionFunc
	restitutionCallback          RestitutionFunc
	profile                      Profile
	maxCapacity                  Capacity
	preSolveFcn                  PreSolveFunc
	preContinuousFcn             PreContinuousFunc
	customFilterFcn              CustomFilterFunc
	workerCount                  int
	enqueueTaskFcn               EnqueueTaskFunc
	userTreeTask                 func()
	scheduler                    *scheduler.Scheduler
	userData                     any
	logger                       *slog.Logger
	recording                    *Recording
	invH                         float32
	invDT                        float32
	activeTaskCount              int
	taskCount                    int
	enableSleep                  bool
	locked                       bool
	enableWarmStarting           bool
	enableContinuous             bool
	inUse                        bool
}

type eventSlice[E any] []E

func NewWorld(def *WorldDef) *World {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(def.Logger != nil, "def.Logger != nil")
	checkInput(def.Gravity.IsValid(), "b2IsValidVec2( def->gravity )")
	checkInput(IsValidFloat(def.RestitutionThreshold), "b2IsValidFloat( def->restitutionThreshold )")
	checkInput(def.RestitutionIterations >= 0, "def->restitutionIterations >= 0")
	checkInput(IsValidFloat(def.HitEventThreshold), "b2IsValidFloat( def->hitEventThreshold )")
	checkInput(IsValidFloat(def.ContactHertz), "b2IsValidFloat( def->contactHertz )")
	checkInput(IsValidFloat(def.ContactDampingRatio), "b2IsValidFloat( def->contactDampingRatio )")
	checkInput(IsValidFloat(def.ContactSpeed), "b2IsValidFloat( def->contactSpeed )")
	checkInput(IsValidFloat(def.MaximumLinearSpeed) && def.MaximumLinearSpeed > 0, "b2IsValidFloat( def->maximumLinearSpeed ) && def->maximumLinearSpeed > 0.0f")

	w := &World{inUse: true}
	createBroadPhase(&w.broadPhase, &def.Capacity)
	createGraph(&w.constraintGraph, &def.Capacity)
	// pools
	w.bodyIDPool = idpool.New()
	bodyCapacity := max(16, def.Capacity.StaticBodyCount+def.Capacity.DynamicBodyCount)
	w.bodies = make([]body, 0, bodyCapacity)
	w.solverSets = make([]solverSet, 0, 8)
	// add empty static, active, and disabled body sets
	w.solverSetIDPool = idpool.New()
	// static set
	w.solverSets = array.Push(w.solverSets, solverSet{setIndex: w.solverSetIDPool.Alloc()})
	w.solverSets[staticSet].bodySims = array.Reserve(w.solverSets[staticSet].bodySims, max(16, def.Capacity.StaticBodyCount))
	core.Assert(w.solverSets[staticSet].setIndex == staticSet, "world->solverSets.data[b2_staticSet].setIndex == b2_staticSet")
	// disabled set
	w.solverSets = array.Push(w.solverSets, solverSet{setIndex: w.solverSetIDPool.Alloc()})
	core.Assert(w.solverSets[disabledSet].setIndex == disabledSet, "world->solverSets.data[b2_disabledSet].setIndex == b2_disabledSet")
	// awake set
	w.solverSets = array.Push(w.solverSets, solverSet{setIndex: w.solverSetIDPool.Alloc()})
	w.solverSets[awakeSet].bodySims = array.Reserve(w.solverSets[awakeSet].bodySims, max(16, def.Capacity.DynamicBodyCount))
	w.solverSets[awakeSet].bodyStates = array.Reserve(w.solverSets[awakeSet].bodyStates, max(16, def.Capacity.DynamicBodyCount))
	w.solverSets[awakeSet].contactSims = array.Reserve(w.solverSets[awakeSet].contactSims, max(16, def.Capacity.ContactCount))
	core.Assert(w.solverSets[awakeSet].setIndex == awakeSet, "world->solverSets.data[b2_awakeSet].setIndex == b2_awakeSet")
	w.shapeIDPool = idpool.New()
	shapeCapacity := max(16, def.Capacity.StaticShapeCount+def.Capacity.DynamicShapeCount)
	w.shapes = make([]shape, 0, shapeCapacity)
	w.fatAABBs = make([]AABB, 0, shapeCapacity)
	w.chainIDPool = idpool.New()
	w.chainShapes = make([]chainShape, 0, 4)
	w.contactIDPool = idpool.New()
	w.contacts = make([]contact, 0, max(16, def.Capacity.ContactCount))
	w.jointIDPool = idpool.New()
	w.joints = make([]joint, 0, 16)
	w.islandIDPool = idpool.New()
	w.islands = make([]island, 0, max(16, def.Capacity.DynamicBodyCount))
	w.sensors = make([]sensor, 0, 4)
	w.bodyMoveEvents = make([]BodyMoveEvent, 0, 4)
	w.sensorBeginEvents = make([]SensorBeginTouchEvent, 0, 4)
	w.sensorEndEvents[0] = make([]SensorEndTouchEvent, 0, 4)
	w.sensorEndEvents[1] = make([]SensorEndTouchEvent, 0, 4)
	w.contactBeginEvents = make([]ContactBeginTouchEvent, 0, 4)
	w.contactEndEvents[0] = make([]ContactEndTouchEvent, 0, 4)
	w.contactEndEvents[1] = make([]ContactEndTouchEvent, 0, 4)
	w.contactHitEvents = make([]ContactHitEvent, 0, 4)
	w.jointEvents = make([]JointEvent, 0, 4)
	w.splitIslandID = nullIndex
	w.gravity = def.Gravity
	w.hitEventThreshold = clampFloat(def.HitEventThreshold, 0, math.MaxFloat32)
	w.restitutionThreshold = clampFloat(def.RestitutionThreshold, 0, math.MaxFloat32)
	// Clamp this to avoid overflowing the solver sync flags.
	w.restitutionIterations = min(max(def.RestitutionIterations, 0), MaxRestitutionIterations)
	w.enableRestitutionPropagation = def.EnableRestitutionPropagation
	w.maxLinearSpeed = def.MaximumLinearSpeed
	w.contactSpeed = clampFloat(def.ContactSpeed, 0, math.MaxFloat32)
	w.contactHertz = clampFloat(def.ContactHertz, 0, math.MaxFloat32)
	w.contactDampingRatio = clampFloat(def.ContactDampingRatio, 0, math.MaxFloat32)
	w.contactRecycleDistance = ContactRecycleDistance()
	w.frictionCallback = def.FrictionCallback
	w.restitutionCallback = def.RestitutionCallback
	w.enableSleep = def.EnableSleep
	w.enableWarmStarting = true
	w.enableContinuous = def.EnableContinuous
	w.userData = def.UserData

	w.logger = def.Logger
	if def.WorkerCount > 0 && def.EnqueueTask != nil {
		// External task system
		w.workerCount = min(def.WorkerCount, MaxWorkers)
		w.enqueueTaskFcn = def.EnqueueTask
	} else {
		w.workerCount = max(1, min(def.WorkerCount, MaxWorkers))
		if w.workerCount > 1 {
			w.scheduler = scheduler.New(w.workerCount, MaxTasks)
			w.enqueueTaskFcn = func(task func()) func() { return w.scheduler.Enqueue(task) }
		} else {
			// Serial fallback
			w.enqueueTaskFcn = defaultAddTaskFcn
		}
	}

	createWorkerContexts(w)
	w.debugBodySet = bitset.New(256)
	w.debugJointSet = bitset.New(256)
	w.debugContactSet = bitset.New(256)
	w.debugIslandSet = bitset.New(256)

	return w
}

func getWorldLocked(w *World) *World {
	if w.locked {
		core.Assert(false, "false")
		return nil
	}

	return w
}

func defaultAddTaskFcn(task func()) func() {
	task()
	return nil
}

func createWorkerContexts(w *World) {
	w.taskContexts = make([]taskContext, w.workerCount)

	w.sensorTaskContexts = make([]sensorTaskContext, w.workerCount)
	for i := range w.workerCount {
		c := &w.taskContexts[i]
		c.sensorHits = make([]sensorHit, 0, 8)
		c.contactStateBitSet = bitset.New(1024)
		c.hitEventBitSet = bitset.New(1024)
		c.jointStateBitSet = bitset.New(1024)
		c.enlargedSimBitSet = bitset.New(256)
		c.awakeIslandBitSet = bitset.New(256)
		c.splitIslandID = nullIndex
		w.sensorTaskContexts[i].eventBits = bitset.New(128)
	}
}

func destroyWorkerContexts(w *World) {
	for i := range w.workerCount {
		c := &w.taskContexts[i]
		c.sensorHits = nil
		c.pairKeys = nil
		c.contactStateBitSet.Destroy()
		c.hitEventBitSet.Destroy()
		c.jointStateBitSet.Destroy()
		c.enlargedSimBitSet.Destroy()
		c.awakeIslandBitSet.Destroy()
		w.sensorTaskContexts[i].eventBits.Destroy()
	}

	w.taskContexts = nil
	w.sensorTaskContexts = nil
}

func (w *World) Close() {
	state := w
	// Detach any recording before teardown. The user owns and frees the recording buffer.
	stopRecordingInternal(state)

	if state.scheduler != nil {
		state.scheduler.Destroy()
		state.scheduler = nil
	}

	state.debugBodySet.Destroy()
	state.debugJointSet.Destroy()
	state.debugContactSet.Destroy()
	state.debugIslandSet.Destroy()
	destroyWorkerContexts(state)
	state.bodyMoveEvents = nil
	state.sensorBeginEvents = nil
	state.sensorEndEvents = [2][]SensorEndTouchEvent{}
	state.contactBeginEvents = nil
	state.contactEndEvents = [2][]ContactEndTouchEvent{}
	state.contactHitEvents = nil

	state.jointEvents = nil
	for i := range state.chainShapes {
		chain := &state.chainShapes[i]
		if chain.id != nullIndex {
			freeChainData(chain)
		} else {
			core.Assert(chain.shapeIndices == nil, "chain->shapeIndices == NULL")
		}
	}

	for i := range state.sensors {
		state.sensors[i].hits = nil
		state.sensors[i].overlaps1 = nil
		state.sensors[i].overlaps2 = nil
	}

	state.sensors = nil
	state.bodies = nil
	state.shapes = nil
	state.fatAABBs = nil
	state.chainShapes = nil
	state.contacts = nil

	state.joints = nil
	for i := range state.islands {
		state.islands[i].bodies = nil
		state.islands[i].contacts = nil
		state.islands[i].joints = nil
	}

	state.islands = nil
	// The entire solver-set pool is being discarded, so no ids need recycling.
	// A failed snapshot rebuild may leave the sets and id pool inconsistent.
	// Dropping the owning slice releases even partially decoded set storage.
	state.solverSets = nil
	destroyGraph(&state.constraintGraph)
	destroyBroadPhase(&state.broadPhase)
	state.bodyIDPool.Destroy()
	state.shapeIDPool.Destroy()
	state.chainIDPool.Destroy()
	state.contactIDPool.Destroy()
	state.jointIDPool.Destroy()
	state.islandIDPool.Destroy()
	state.solverSetIDPool.Destroy()
	*state = World{}
}

func relativeCos(a, b Rot) float32 {
	return float32(a.C*b.C) + float32(a.S*b.S)
}

func resolveContactBodySim(w *World, awakeSims, staticSims []bodySim, encodedIndex, bodyID int) *bodySim {
	if encodedIndex >= 0 {
		return &awakeSims[encodedIndex]
	}

	if isStaticSimIndex(encodedIndex) {
		return &staticSims[-encodedIndex-2]
	}

	return getBodySim(w, &w.bodies[bodyID])
}

func tryRecycleContact(sim *contactSim, bodySimA, bodySimB *bodySim, transformA, transformB *WorldTransform, recycleDistance, recycleDistanceNonTouching float32) bool {
	encodedA := sim.encodedBodySimA
	encodedB := sim.encodedBodySimB
	wasTouching := sim.simFlags&simTouchingFlag != 0
	cachedQA := sim.cachedRotationA
	cachedQB := sim.cachedRotationB
	xfc := sim.cachedRelativePose
	xf := transformA.InvMul(*transformB)
	cosA := relativeCos(transformA.Q, cachedQA)
	cosB := relativeCos(transformB.Q, cachedQB)
	minCos := minFloat(cosA, cosB)

	var maxExtentA, maxExtentB float32
	if !isStaticSimIndex(encodedA) {
		maxExtentA = bodySimA.maxExtent
	}

	if !isStaticSimIndex(encodedB) {
		maxExtentB = bodySimB.maxExtent
	}

	maxExtent := maxFloat(maxExtentA, maxExtentB)
	distance := xf.P.Distance(xfc.P)
	qr := xf.Q.InvMul(xfc.Q)
	// This metric is used for fast bodies and sleeping. It comes from conservative advancement.
	// Note that qr.S == sin(theta) ~= theta for small angles.
	// Need a tighter tolerance for non-touching shapes so that contacts are not missed.
	tolerance := recycleDistanceNonTouching
	if wasTouching {
		tolerance = recycleDistance
	}

	if minCos > ContactRecycleCosAngle && distance+float32(maxExtent*absFloat(qr.S)) < tolerance {
		dqA := transformA.Q.Mul(cachedQA.Invert())
		dqB := transformB.Q.Mul(cachedQB.Invert())
		normal := sim.manifold.Normal
		// Minimize round-off
		dc := bodySimB.center.Sub(bodySimA.center)

		for i := range sim.manifold.PointCount {
			// Keep anchors but update separation, same as sub-stepping. This eliminates jitter.
			mp := &sim.manifold.Points[i]
			rA := dqA.Rotate(mp.AnchorA)
			rB := dqB.Rotate(mp.AnchorB)
			dp := dc.Add(rB.Sub(rA))
			mp.Separation = mp.BaseSeparation + dp.Dot(normal)
			mp.NormalVelocity = 0
			mp.Persisted = true
		}

		return true
	}

	return false
}

func updateContactTouchingState(task *taskContext, sim *contactSim, touching, wasTouching bool) {
	contactID := sim.contactID
	// State changes that affect island connectivity. Also affects contact events.
	if touching && !wasTouching {
		sim.simFlags |= simStartedTouching

		task.contactStateBitSet.SetBit(uint32(contactID))
	} else if !touching && wasTouching {
		sim.simFlags |= simStoppedTouching

		task.contactStateBitSet.SetBit(uint32(contactID))
	}

	if touching {
		for i := range sim.manifold.PointCount {
			// Cache separation
			mp := &sim.manifold.Points[i]
			mp.BaseSeparation = mp.Separation
		}
	}
}

func collideTask(startIndex, endIndex, workerIndex int, context *stepContext) {
	w := context.world
	task := &w.taskContexts[workerIndex]
	spans := context.collideSpans
	shapes := w.shapes
	fatAABBs := w.fatAABBs
	awakeSims := w.solverSets[awakeSet].bodySims
	staticSims := w.solverSets[staticSet].bodySims

	core.Assert(startIndex < endIndex, "startIndex < endIndex")

	spanIndex := 0
	for spans[spanIndex+1].start <= startIndex {
		spanIndex++
	}

	spanStart := spans[spanIndex].start
	spanEnd := spans[spanIndex+1].start
	spanBase := spans[spanIndex].contacts
	recycleDistance := w.contactRecycleDistance
	speculativeDistance := SpeculativeDistance()
	recycleDistanceNonTouching := minFloat(recycleDistance, speculativeDistance)

	for contactIndex := startIndex; contactIndex < endIndex; contactIndex++ {
		if contactIndex == spanEnd {
			spanIndex++
			spanStart = spans[spanIndex].start
			spanEnd = spans[spanIndex+1].start
			spanBase = spans[spanIndex].contacts
		}

		sim := &spanBase[contactIndex-spanStart]
		contactID := sim.contactID
		shapeIDA := sim.shapeIDA
		shapeIDB := sim.shapeIDB
		// Do proxies still overlap?
		overlap := overlapV(&fatAABBs[shapeIDA], &fatAABBs[shapeIDB])
		if !overlap {
			sim.simFlags |= simDisjoint
			sim.simFlags &^= simTouchingFlag

			task.contactStateBitSet.SetBit(uint32(contactID))

			continue
		}

		shapeA := &shapes[shapeIDA]
		shapeB := &shapes[shapeIDB]
		wasTouching := sim.simFlags&simTouchingFlag != 0
		// Update contact respecting shape/body order (A,B)
		encodedA := sim.encodedBodySimA
		encodedB := sim.encodedBodySimB
		bodySimA := resolveContactBodySim(w, awakeSims, staticSims, encodedA, shapeA.bodyID)
		bodySimB := resolveContactBodySim(w, awakeSims, staticSims, encodedB, shapeB.bodyID)
		transformA := bodySimA.transform
		transformB := bodySimB.transform
		// These may not be skipped by relative transform check below
		sim.invMassA = bodySimA.invMass
		sim.invIA = bodySimA.invInertia
		sim.invMassB = bodySimB.invMass
		sim.invIB = bodySimB.invInertia
		fast := (bodySimA.flags|bodySimB.flags)&isFast != 0
		// Contact recycling optimization. Please cite this code if you use this optimization.
		// This is inspired by persistent contact manifolds used in some physics engines, such as PhysX.
		// However, this allows larger relative motion and has fewer tuning parameters (just one).
		if !fast && recycleDistance > 0 && sim.simFlags&simRelativeTransformValid != 0 && sim.simFlags&contactRecycleFlag != 0 {
			if tryRecycleContact(sim, bodySimA, bodySimB, &transformA, &transformB, recycleDistance, recycleDistanceNonTouching) {
				task.recycledContactCount++
				// Contact is recycled. This also skips updating other aspects of the contact
				// such as material parameters.
				continue
			}
		}
		// Caching for contact recycling.
		sim.cachedRotationA = transformA.Q
		sim.cachedRotationB = transformB.Q
		sim.cachedRelativePose = transformA.InvMul(transformB)
		sim.simFlags |= simRelativeTransformValid
		centerOffsetA := transformA.Q.Rotate(bodySimA.localCenter)
		centerOffsetB := transformB.Q.Rotate(bodySimB.localCenter)
		// This updates solid contacts
		touching := updateContact(w, sim, shapeA, transformA, centerOffsetA, shapeB, transformB, centerOffsetB)
		updateContactTouchingState(task, sim, touching, wasTouching)
	}
}

func addNonTouchingContact(w *World, c *contact, sim *contactSim) {
	core.Assert(c.setIndex == awakeSet, "contact->setIndex == b2_awakeSet")
	set := &w.solverSets[awakeSet]
	c.colorIndex = nullIndex
	c.localIndex = len(set.contactSims)
	set.contactSims = array.Push(set.contactSims, *sim)
}

func removeNonTouchingContact(w *World, setIndex, localIndex int) {
	set := &w.solverSets[setIndex]

	var movedIndex int

	set.contactSims, movedIndex = array.RemoveSwap(set.contactSims, localIndex)
	if movedIndex != nullIndex {
		movedSim := &set.contactSims[localIndex]
		moved := &w.contacts[movedSim.contactID]
		core.Assert(moved.setIndex == setIndex, "movedContact->setIndex == setIndex")
		core.Assert(moved.localIndex == movedIndex, "movedContact->localIndex == movedIndex")
		core.Assert(moved.colorIndex == nullIndex, "movedContact->colorIndex == B2_NULL_INDEX")
		moved.localIndex = localIndex
	}
}

func processContactStateChange(w *World, contactID int, awake *solverSet, endEventArrayIndex int, shapes []shape) {
	colors := &w.constraintGraph.colors
	c := &w.contacts[contactID]
	core.Assert(c.setIndex == awakeSet, "contact->setIndex == b2_awakeSet")
	colorIndex := c.colorIndex
	localIndex := c.localIndex

	var sim *contactSim

	if colorIndex != nullIndex {
		// contact lives in constraint graph
		core.Assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")
		sim = &colors[colorIndex].contactSims[localIndex]
	} else {
		sim = &awake.contactSims[localIndex]
	}

	flags := c.flags

	simFlags := sim.simFlags
	switch {
	case simFlags&simDisjoint != 0:
		// Bounding boxes no longer overlap
		destroyContact(w, c)
	case simFlags&simStartedTouching != 0:
		core.Assert(c.islandID == nullIndex, "contact->islandId == B2_NULL_INDEX")

		if flags&contactEnableContactEvents != 0 {
			shapeA := &shapes[c.shapeIDA]
			shapeB := &shapes[c.shapeIDB]
			shapeIDA := Shape{index1: int32(shapeA.id + 1), world: w, generation: shapeA.generation}
			shapeIDB := Shape{index1: int32(shapeB.id + 1), world: w, generation: shapeB.generation}
			fullID := Contact{index1: int32(contactID + 1), world: w, generation: c.generation}
			w.contactBeginEvents = array.Push(w.contactBeginEvents, ContactBeginTouchEvent{ShapeA: shapeIDA, ShapeB: shapeIDB, Contact: fullID})
		}

		core.Assert(sim.manifold.PointCount > 0, "contactSim->manifold.pointCount > 0")
		core.Assert(c.setIndex == awakeSet, "contact->setIndex == b2_awakeSet")
		// Link first because this wakes colliding bodies and ensures the body sims
		// are in the correct place.
		c.flags |= contactTouchingFlag
		linkContact(w, c)
		// Make sure these didn't change
		core.Assert(c.colorIndex == nullIndex, "contact->colorIndex == B2_NULL_INDEX")
		core.Assert(c.localIndex == localIndex, "contact->localIndex == localIndex")
		// Contact sim pointer may have become orphaned due to awake set growth,
		// so I just need to refresh it.
		sim = &awake.contactSims[localIndex]
		sim.simFlags &^= simStartedTouching
		// Add first for memcpy
		addContactToGraph(w, sim, c)
		// This destroys the contact sim
		removeNonTouchingContact(w, awakeSet, localIndex)
	case simFlags&simStoppedTouching != 0:
		sim.simFlags &^= simStoppedTouching

		c.flags &^= contactTouchingFlag
		if c.flags&contactEnableContactEvents != 0 {
			shapeA := &shapes[c.shapeIDA]
			shapeB := &shapes[c.shapeIDB]
			shapeIDA := Shape{index1: int32(shapeA.id + 1), world: w, generation: shapeA.generation}
			shapeIDB := Shape{index1: int32(shapeB.id + 1), world: w, generation: shapeB.generation}
			fullID := Contact{index1: int32(contactID + 1), world: w, generation: c.generation}
			w.contactEndEvents[endEventArrayIndex] = array.Push(w.contactEndEvents[endEventArrayIndex], ContactEndTouchEvent{ShapeA: shapeIDA, ShapeB: shapeIDB, Contact: fullID})
		}

		core.Assert(sim.manifold.PointCount == 0, "contactSim->manifold.pointCount == 0")
		unlinkContact(w, c)
		bodyIDA := c.edges[0].bodyID
		bodyIDB := c.edges[1].bodyID
		// Add first for memcpy
		addNonTouchingContact(w, c, sim)
		removeContactFromGraph(w, bodyIDA, bodyIDB, colorIndex, localIndex)
	}
}

func collide(context *stepContext) {
	w := context.world
	core.Assert(w.workerCount > 0, "world->workerCount > 0")
	// gather contacts into a single array for easier parallel-for
	contactCount := 0

	colors := &w.constraintGraph.colors
	for i := range GraphColorCount {
		contactCount += len(colors[i].contactSims)
	}

	nonTouchingCount := len(w.solverSets[awakeSet].contactSims)

	contactCount += nonTouchingCount
	if contactCount == 0 {
		return
	}

	spans := w.collideSpanScratch[:]

	spanCount := 0
	contactIndex := 0

	for i := range GraphColorCount {
		color := &colors[i]

		count := len(color.contactSims)
		if count > 0 {
			spans[spanCount] = contactCollideSpan{start: contactIndex, contacts: color.contactSims}
			spanCount++
			contactIndex += count
		}
	}

	if nonTouchingCount > 0 {
		spans[spanCount] = contactCollideSpan{start: contactIndex, contacts: w.solverSets[awakeSet].contactSims}
		spanCount++
		contactIndex += nonTouchingCount
	}

	core.Assert(contactIndex == contactCount, "contactIndex == contactCount")
	core.Assert(spanCount <= GraphColorCount+1, "spanCount <= B2_GRAPH_COLOR_COUNT + 1")
	spans[spanCount].start = contactCount
	context.collideSpans = spans[:spanCount+1]
	// Contact bit set on ids because contact pointers are unstable as they move between touching and not touching.
	contactIDCapacity := w.contactIDPool.Capacity()
	for i := range w.workerCount {
		w.taskContexts[i].contactStateBitSet.SetBitCountAndClear(uint32(contactIDCapacity))
		w.taskContexts[i].recycledContactCount = 0
	}
	// Task should take at least 40us on a 4GHz CPU (10K cycles)
	minRange := 64
	parallelFor(w, func(start, end, worker int) { collideTask(start, end, worker, context) }, contactCount, minRange)

	context.collideSpans = nil

	clear(spans)
	// Serially update contact state
	// Bitwise OR all contact bits
	bitSet := &w.taskContexts[0].contactStateBitSet
	for i := 1; i < w.workerCount; i++ {
		bitSet.InPlaceUnion(&w.taskContexts[i].contactStateBitSet)
	}

	awake := &w.solverSets[awakeSet]
	endEventArrayIndex := w.endEventArrayIndex
	shapes := w.shapes
	// Process contact state changes. Iterate over set bits
	for k := range uint32(bitSet.BlockCount) {
		word := bitSet.Bits[k]
		for word != 0 {
			ctz := bits.TrailingZeros64(word)
			contactID := int(64*k) + ctz
			processContactStateChange(w, contactID, awake, endEventArrayIndex, shapes)
			// Clear the smallest set bit
			word &= word - 1
		}
	}

	validateSolverSets(w)
	validateContacts(w)
}

func (w *World) Step(timeStep float32, subStepCount int) {
	checkInput(IsValidFloat(timeStep), "b2IsValidFloat( timeStep )")
	core.Assert(0 < subStepCount, "0 < subStepCount")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}
	// Record step inputs before simulation runs
	if state.recording != nil {
		recWriteStep(state.recording, &recArgsStep{dt: timeStep, subStepCount: int32(subStepCount)})
	}
	// Prepare to capture events
	// Ensure user does not access stale data if there is an early return
	state.bodyMoveEvents = state.bodyMoveEvents[:0]
	state.sensorBeginEvents = state.sensorBeginEvents[:0]
	state.contactBeginEvents = state.contactBeginEvents[:0]
	state.contactHitEvents = state.contactHitEvents[:0]
	state.jointEvents = state.jointEvents[:0]
	state.profile = Profile{}
	state.locked = true
	state.activeTaskCount = 0

	state.taskCount = 0
	if state.scheduler != nil {
		state.scheduler.Reset()
	}

	stepTicks := timer.Ticks()
	{
		c := &state.maxCapacity
		c.StaticShapeCount = max(c.StaticShapeCount, int(state.broadPhase.trees[StaticBody].proxyCount))
		c.DynamicShapeCount = max(c.DynamicShapeCount, int(state.broadPhase.trees[DynamicBody].proxyCount))
		staticBodyCount := len(state.solverSets[staticSet].bodySims)
		c.StaticBodyCount = max(c.StaticBodyCount, staticBodyCount)
		// this includes kinematic bodies
		totalBodyCount := state.bodyIDPool.Count()
		c.DynamicBodyCount = max(c.DynamicBodyCount, totalBodyCount-staticBodyCount)
		totalContactCount := state.contactIDPool.Count()
		c.ContactCount = max(c.ContactCount, totalContactCount)
	}
	// Update collision pairs and create contacts
	{
		pairTicks := timer.Ticks()

		updateBroadPhasePairs(state)
		state.profile.Pairs = timer.Milliseconds(pairTicks)
	}

	context := &state.stepScratch
	context.world = state
	context.dt = timeStep

	context.subStepCount = max(1, subStepCount)
	if timeStep > 0 {
		context.invDt = 1 / timeStep
		context.h = timeStep / float32(context.subStepCount)
		context.invH = float32(context.subStepCount) * context.invDt
	}

	state.invH = context.invH
	state.invDT = context.invDt
	// Hertz values get reduced for large time steps
	contactHertz := minFloat(state.contactHertz, 0.125*context.invH)
	context.contactSoftness = makeSoft(contactHertz, state.contactDampingRatio, context.h)
	context.staticSoftness = makeSoft(2*contactHertz, state.contactDampingRatio, context.h)
	context.restitutionThreshold = state.restitutionThreshold
	context.maxLinearVelocity = state.maxLinearSpeed
	context.enableWarmStarting = state.enableWarmStarting
	// Narrow phase : update contacts
	{
		collideTicks := timer.Ticks()

		collide(context)

		state.profile.Collide = timer.Milliseconds(collideTicks)
	}
	// Integrate velocities, solve velocity constraints, and integrate positions.
	if timeStep > 0 {
		solveTicks := timer.Ticks()

		solve(state, context)
		state.profile.Solve = timer.Milliseconds(solveTicks)
	}
	// Finish the tree task in case solve didn't finish it
	if state.userTreeTask != nil {
		state.userTreeTask()
		state.userTreeTask = nil
		state.activeTaskCount--
		validateNoEnlarged(&state.broadPhase)
	}
	// Update sensors
	{
		sensorTicks := timer.Ticks()

		overlapSensors(state)
		state.profile.Sensors = timer.Milliseconds(sensorTicks)
	}

	state.profile.Step = timer.Milliseconds(stepTicks)
	// Make sure all tasks that were started were also finished
	core.Assert(state.activeTaskCount == 0, "world->activeTaskCount == 0")
	// Swap end event array buffers
	state.endEventArrayIndex = 1 - state.endEventArrayIndex
	state.sensorEndEvents[state.endEventArrayIndex] = state.sensorEndEvents[state.endEventArrayIndex][:0]

	state.contactEndEvents[state.endEventArrayIndex] = state.contactEndEvents[state.endEventArrayIndex][:0]
	if state.recording != nil {
		// Write the per-step StateHash while the world is still locked. Queries early return while
		// locked, so this keeps the shared recording buffer single-writer without a lock. StateHash
		// proves the simulation reproduced exactly on replay.
		hash := hashWorldState(state)
		recWriteStateHash(state.recording, &recArgsStateHash{hash: hash})
		// Grow the recorded bounds so a replay can frame the whole motion, not just frame 0
		var bounds AABB
		if computeWorldBounds(state, &bounds) {
			recAccumulateBounds(state.recording, bounds)
		}
	}

	state.stepScratch = stepContext{}
	clear(state.contactPrepareSpanScratch[:])
	clear(state.jointPrepareSpanScratch[:])
	clear(state.solverWorkerScratch[:])
	clear(state.solverStageScratch)
	state.locked = false
}

func computeWorldBounds(w *World, bounds *AABB) bool {
	worldBounds := AABB{}
	haveBounds := false

	for i := range int(BodyTypeCount) {
		tree := &w.broadPhase.trees[i]
		if tree.ProxyCount() == 0 {
			continue
		}

		treeBounds := tree.RootBounds()
		if haveBounds {
			worldBounds = worldBounds.Union(treeBounds)
		} else {
			worldBounds = treeBounds
		}

		haveBounds = true
	}

	*bounds = worldBounds

	return haveBounds
}

func (w *World) Bounds() AABB {
	state := getWorldLocked(w)
	if state == nil {
		return AABB{}
	}

	var bounds AABB
	computeWorldBounds(state, &bounds)

	return bounds
}

func (w *World) BodyMoveEvents() iter.Seq[BodyMoveEvent] {
	return eventSlice[BodyMoveEvent](checkedEventSlice(w, w.bodyMoveEvents)).values
}

func (w *World) SensorBeginEvents() iter.Seq[SensorBeginTouchEvent] {
	return eventSlice[SensorBeginTouchEvent](checkedEventSlice(w, w.sensorBeginEvents)).values
}

func (w *World) SensorEndEvents() iter.Seq[SensorEndTouchEvent] {
	// Careful to use previous buffer
	return eventSlice[SensorEndTouchEvent](checkedEndEventSlice(w, &w.sensorEndEvents)).values
}

func (w *World) ContactBeginEvents() iter.Seq[ContactBeginTouchEvent] {
	return eventSlice[ContactBeginTouchEvent](checkedEventSlice(w, w.contactBeginEvents)).values
}

func (w *World) ContactEndEvents() iter.Seq[ContactEndTouchEvent] {
	// Careful to use previous buffer
	return eventSlice[ContactEndTouchEvent](checkedEndEventSlice(w, &w.contactEndEvents)).values
}

func (w *World) ContactHitEvents() iter.Seq[ContactHitEvent] {
	return eventSlice[ContactHitEvent](checkedEventSlice(w, w.contactHitEvents)).values
}

func (w *World) JointEvents() iter.Seq[JointEvent] {
	return eventSlice[JointEvent](checkedEventSlice(w, w.jointEvents)).values
}

func (events eventSlice[E]) values(yield func(E) bool) {
	for _, event := range events {
		if !yield(event) {
			return
		}
	}
}

//go:noinline
func checkedEventSlice[E any](w *World, events []E) []E {
	core.Assert(!w.locked, "world->locked == false")

	if w.locked {
		return nil
	}

	return events
}

//go:noinline
func checkedEndEventSlice[E any](w *World, events *[2][]E) []E {
	return checkedEventSlice(w, events[1-w.endEventArrayIndex])
}

func (w *World) IsValid() bool {
	return w != nil && w.inUse
}

func (b Body) IsValid() bool {
	state := b.world
	if !state.IsValid() {
		return false
	}

	index := int(b.index1) - 1
	if index < 0 || index >= len(state.bodies) {
		return false
	}

	item := &state.bodies[index]
	if item.setIndex == nullIndex {
		return false
	}

	core.Assert(item.localIndex != nullIndex, "body->localIndex != B2_NULL_INDEX")

	return item.generation == b.generation
}

func (s Shape) IsValid() bool {
	state := s.world
	if !state.IsValid() {
		return false
	}

	index := int(s.index1) - 1
	if index < 0 || index >= len(state.shapes) {
		return false
	}

	item := &state.shapes[index]
	if item.id == nullIndex {
		return false
	}

	core.Assert(item.id == index, "shape->id == shapeId")

	return item.generation == s.generation
}

func (c Chain) IsValid() bool {
	state := c.world
	if !state.IsValid() {
		return false
	}

	index := int(c.index1) - 1
	if index < 0 || index >= len(state.chainShapes) {
		return false
	}

	item := &state.chainShapes[index]
	if item.id == nullIndex {
		return false
	}

	core.Assert(item.id == index, "chain->id == chainId")

	return item.generation == c.generation
}

func (j Joint) IsValid() bool {
	state := j.world
	if !state.IsValid() {
		return false
	}

	index := int(j.index1) - 1
	if index < 0 || index >= len(state.joints) {
		return false
	}

	item := &state.joints[index]
	if item.jointID == nullIndex {
		return false
	}

	core.Assert(item.jointID == index, "joint->jointId == jointId")

	return item.generation == j.generation
}

func (c Contact) IsValid() bool {
	state := c.world
	if !state.IsValid() {
		return false
	}

	index := int(c.index1) - 1
	if index < 0 || index >= len(state.contacts) {
		return false
	}

	item := &state.contacts[index]
	if item.contactID == nullIndex {
		return false
	}

	core.Assert(item.contactID == index, "contact->contactId == contactId")

	return item.generation == c.generation
}

func (w *World) SetSleepingEnabled(flag bool) {
	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldEnableSleeping(state.recording, &recArgsWorldEnableSleeping{flag: flag})
	}

	if flag == state.enableSleep {
		return
	}

	state.enableSleep = flag
	if !flag {
		for i := firstSleepingSet; i < len(state.solverSets); i++ {
			if len(state.solverSets[i].bodySims) > 0 {
				wakeSolverSet(state, i)
			}
		}
	}
}

func (w *World) SleepingEnabled() bool {
	return w.enableSleep
}

func (w *World) SetWarmStartingEnabled(flag bool) {
	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldEnableWarmStarting(state.recording, &recArgsWorldEnableWarmStarting{flag: flag})
	}

	state.enableWarmStarting = flag
}

func (w *World) WarmStartingEnabled() bool {
	return w.enableWarmStarting
}

func (w *World) AwakeBodyCount() int {
	return len(w.solverSets[awakeSet].bodySims)
}

func (w *World) SetContinuousEnabled(flag bool) {
	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldEnableContinuous(state.recording, &recArgsWorldEnableContinuous{flag: flag})
	}

	state.enableContinuous = flag
}

func (w *World) ContinuousEnabled() bool {
	return w.enableContinuous
}

func (w *World) SetRestitutionThreshold(value float32) {
	checkInput(IsValidFloat(value), "b2IsValidFloat( value )")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldSetRestitutionThreshold(state.recording, &recArgsWorldSetRestitutionThreshold{value: value})
	}

	state.restitutionThreshold = clampFloat(value, 0, math.MaxFloat32)
}

func (w *World) RestitutionThreshold() float32 {
	return w.restitutionThreshold
}

func (w *World) SetRestitutionIterations(iterations int) {
	checkInput(iterations >= 0, "iterations >= 0")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	iterations = min(max(iterations, 0), MaxRestitutionIterations)
	if state.recording != nil {
		recWriteWorldSetRestitutionIterations(state.recording, &recArgsWorldSetRestitutionIterations{iterations: int32(iterations)})
	}

	state.restitutionIterations = iterations
}

func (w *World) RestitutionIterations() int {
	return w.restitutionIterations
}

func (w *World) SetRestitutionPropagationEnabled(flag bool) {
	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldEnableRestitutionPropagation(state.recording, &recArgsWorldEnableRestitutionPropagation{flag: flag})
	}

	state.enableRestitutionPropagation = flag
}

func (w *World) RestitutionPropagationEnabled() bool {
	return w.enableRestitutionPropagation
}

func (w *World) SetHitEventThreshold(value float32) {
	checkInput(IsValidFloat(value), "b2IsValidFloat( value )")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldSetHitEventThreshold(state.recording, &recArgsWorldSetHitEventThreshold{value: value})
	}

	state.hitEventThreshold = clampFloat(value, 0, math.MaxFloat32)
}

func (w *World) HitEventThreshold() float32 {
	return w.hitEventThreshold
}

func (w *World) SetContactTuning(hertz, dampingRatio, pushSpeed float32) {
	checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")
	checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )")
	checkInput(IsValidFloat(pushSpeed), "b2IsValidFloat( pushSpeed )")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldSetContactTuning(state.recording, &recArgsWorldSetContactTuning{hertz: hertz, dampingRatio: dampingRatio, pushSpeed: pushSpeed})
	}

	state.contactHertz = clampFloat(hertz, 0, math.MaxFloat32)
	state.contactDampingRatio = clampFloat(dampingRatio, 0, math.MaxFloat32)
	state.contactSpeed = clampFloat(pushSpeed, 0, math.MaxFloat32)
}

func (w *World) SetContactRecycleDistance(recycleDistance float32) {
	checkInput(IsValidFloat(recycleDistance), "b2IsValidFloat( recycleDistance )")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldSetContactRecycleDistance(state.recording, &recArgsWorldSetContactRecycleDistance{recycleDistance: recycleDistance})
	}

	state.contactRecycleDistance = clampFloat(recycleDistance, 0, math.MaxFloat32)
}

func (w *World) ContactRecycleDistance() float32 {
	return w.contactRecycleDistance
}

func (w *World) SetMaximumLinearSpeed(maximumLinearSpeed float32) {
	checkInput(IsValidFloat(maximumLinearSpeed) && maximumLinearSpeed > 0, "b2IsValidFloat( maximumLinearSpeed ) && maximumLinearSpeed > 0.0f")

	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldSetMaximumLinearSpeed(state.recording, &recArgsWorldSetMaximumLinearSpeed{maximumLinearSpeed: maximumLinearSpeed})
	}

	state.maxLinearSpeed = maximumLinearSpeed
}

func (w *World) MaximumLinearSpeed() float32 {
	return w.maxLinearSpeed
}

func (w *World) Profile() Profile {
	return w.profile
}

func (w *World) Counters() Counters {
	state := w
	s := Counters{}
	s.BodyCount = state.bodyIDPool.Count()
	s.ShapeCount = state.shapeIDPool.Count()
	s.ContactCount = state.contactIDPool.Count()
	s.JointCount = state.jointIDPool.Count()
	s.IslandCount = state.islandIDPool.Count()
	s.StaticTreeHeight = state.broadPhase.trees[StaticBody].Height()
	s.TreeHeight = max(state.broadPhase.trees[DynamicBody].Height(), state.broadPhase.trees[KinematicBody].Height())

	s.TaskCount = state.taskCount
	for i := range state.workerCount {
		s.RecycledContactCount += state.taskContexts[i].recycledContactCount
	}

	for i := range GraphColorCount {
		color := &state.constraintGraph.colors[i]
		s.ColorCounts[i] = len(color.contactSims) + len(color.jointSims)
		s.AwakeContactCount += len(color.contactSims)
	}

	s.AwakeContactCount += len(state.solverSets[awakeSet].contactSims)

	return s
}

func (w *World) MaxCapacity() Capacity {
	return w.maxCapacity
}

func (w *World) SetUserData(userData any) {
	w.userData = userData
}

func (w *World) UserData() any {
	return w.userData
}

func (w *World) SetFrictionCallback(callback FrictionFunc) {
	state := w
	if state.locked {
		return
	}

	if callback != nil && state.recording != nil {
		state.logger.Warn("box2d: friction callbacks are not supported while recording")
		return
	}

	state.frictionCallback = callback
}

func (w *World) SetRestitutionCallback(callback RestitutionFunc) {
	state := w
	if state.locked {
		return
	}

	if callback != nil && state.recording != nil {
		state.logger.Warn("box2d: restitution callbacks are not supported while recording")
		return
	}

	state.restitutionCallback = callback
}

func (w *World) SetWorkerCount(count int) {
	state := getWorldLocked(w)
	if state == nil {
		return
	}

	if count == state.workerCount {
		return
	}

	destroyWorkerContexts(state)
	state.workerCount = min(max(count, 1), MaxWorkers)
	createWorkerContexts(state)
}

func (w *World) WorkerCount() int {
	state := getWorldLocked(w)
	if state == nil {
		return 0
	}

	return state.workerCount
}

func (w *World) StartRecording(recording *Recording) {
	// Must be a step boundary, so refuse a locked world
	state := getWorldLocked(w)
	if state == nil || recording == nil || state.recording != nil {
		return
	}

	if state.preSolveFcn != nil || state.preContinuousFcn != nil {
		state.logger.Warn("box2d: recording does not support pre-solve callbacks")
		return
	}

	if state.customFilterFcn != nil {
		state.logger.Warn("box2d: recording does not support custom filter callbacks")
		return
	}

	if state.frictionCallback != nil || state.restitutionCallback != nil {
		state.logger.Warn("box2d: recording does not support material callbacks")
		return
	}

	startRecordingIntoBuffer(state, recording)
}

func (w *World) StopRecording() {
	state := getWorldLocked(w)
	if state == nil {
		return
	}

	stopRecordingInternal(state)
}

func (w *World) SetCustomFilterCallback(fcn CustomFilterFunc) {
	state := w
	if fcn != nil && state.recording != nil {
		state.logger.Warn("box2d: custom filter callbacks are not supported while recording")
		core.Assert(false, "false && \"customFilter callbacks are not supported while recording\"")
	}

	state.customFilterFcn = fcn
}

func (w *World) SetPreSolveCallback(preSolveFcn PreSolveFunc, preContinuousFcn PreContinuousFunc) {
	state := w
	if (preSolveFcn != nil || preContinuousFcn != nil) && state.recording != nil {
		state.logger.Warn("box2d: pre-solve callbacks are not supported while recording")
		core.Assert(false, "false && \"preSolve callbacks are not supported while recording\"")
	}

	state.preSolveFcn = preSolveFcn
	state.preContinuousFcn = preContinuousFcn
}

func (w *World) SetGravity(gravity Vec2) {
	checkInput(gravity.IsValid(), "b2IsValidVec2( gravity )")

	state := w
	if state.recording != nil {
		recWriteWorldSetGravity(state.recording, &recArgsWorldSetGravity{gravity: gravity})
	}

	state.gravity = gravity
}

func (w *World) Gravity() Vec2 {
	return w.gravity
}

func (w *World) RebuildStaticTree() {
	state := w
	core.Assert(!state.locked, "world->locked == false")

	if state.locked {
		return
	}

	if state.recording != nil {
		recWriteWorldRebuildStaticTree(state.recording, &recArgsWorldRebuildStaticTree{})
	}

	tree := &state.broadPhase.trees[StaticBody]
	// Moved proxies must retain the flag in the rebuild.
	movedCount := 0

	var movedProxies []int

	proxyCount := tree.ProxyCount()
	if hasTreeMoved(tree) {
		movedProxies = make([]int, proxyCount)
		movedCount = tree.gatherMovedProxies(movedProxies)
	}

	tree.rebuild(true)

	for i := range movedCount {
		tree.markProxyMovedSerial(movedProxies[i])
	}
}
