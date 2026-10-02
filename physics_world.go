// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"iter"
	"sync"
)

type EnqueueTaskFunc func(task func()) (finish func())
type FrictionFunc func(frictionA float32, userMaterialIDA uint64, frictionB float32, userMaterialIDB uint64) float32
type RestitutionFunc func(restitutionA float32, userMaterialIDA uint64, restitutionB float32, userMaterialIDB uint64) float32
type CustomFilterFunc func(shapeIDA, shapeIDB Shape) bool
type PreSolveFunc func(shapeIDA, shapeIDB Shape, manifold *Manifold)
type PreContinuousFunc func(shapeIDA, shapeIDB Shape, point Pos, normal Vec2) bool

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
	internalValue                int32
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
	contactStateBitSet   bitSet
	hitEventBitSet       bitSet
	hasHitEvents         bool
	jointStateBitSet     bitSet
	enlargedSimBitSet    bitSet
	awakeIslandBitSet    bitSet
	splitSleepTime       float32
	splitIslandID        int
	recycledContactCount int
}

type world struct {
	broadPhase                   broadPhase
	constraintGraph              constraintGraph
	bodyIDPool                   idPool
	bodies                       []body
	solverSetIDPool              idPool
	solverSets                   []solverSet
	jointIDPool                  idPool
	joints                       []joint
	contactIDPool                idPool
	contacts                     []contact
	islandIDPool                 idPool
	islands                      []island
	shapeIDPool                  idPool
	chainIDPool                  idPool
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
	debugBodySet                 bitSet
	debugJointSet                bitSet
	debugContactSet              bitSet
	debugIslandSet               bitSet
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
	generation                   uint16
	profile                      Profile
	maxCapacity                  Capacity
	preSolveFcn                  PreSolveFunc
	preContinuousFcn             PreContinuousFunc
	customFilterFcn              CustomFilterFunc
	workerCount                  int
	enqueueTaskFcn               EnqueueTaskFunc
	userTreeTask                 func()
	scheduler                    *scheduler
	userData                     any
	recording                    *Recording
	invH                         float32
	invDT                        float32
	activeTaskCount              int
	taskCount                    int
	worldID                      uint16
	enableSleep                  bool
	locked                       bool
	enableWarmStarting           bool
	enableContinuous             bool
	inUse                        bool
}

var worlds [MaxWorlds]world
var worldsMutex sync.Mutex

func getUnlockedWorldFromId(id World) *world {
	todo()
	return nil
}

func getWorldFromId(id World) *world {
	todo()
	return nil
}

func getWorld(index int) *world {
	todo()
	return nil
}

func getWorldLocked(index int) *world {
	todo()
	return nil
}

func defaultAddTaskFcn(task func()) func() {
	todo()
	return nil
}

func defaultFinishTaskFcn(finish func()) {
	todo()
}

func createWorkerContexts(w *world) {
	todo()
}

func destroyWorkerContexts(w *world) {
	todo()
}

func CreateWorld(def *WorldDef) World {
	todo()
	return World{}
}

func (w World) Destroy() {
	todo()
}

func relativeCos(a, b Rot) float32 {
	todo()
	return 0
}

func resolveContactBodySim(w *world, awakeSims, staticSims []bodySim, encodedIndex, bodyID int) *bodySim {
	todo()
	return nil
}

func collideTask(startIndex, endIndex, workerIndex int, context *stepContext) {
	todo()
}

func addNonTouchingContact(w *world, c *contact, sim *contactSim) {
	todo()
}

func removeNonTouchingContact(w *world, setIndex, localIndex int) {
	todo()
}

func collide(context *stepContext) {
	todo()
}

func (w World) Step(timeStep float32, subStepCount int) {
	todo()
}

func computeWorldBounds(w *world, bounds *AABB) bool {
	todo()
	return false
}

func (w World) Bounds() AABB {
	todo()
	return AABB{}
}

func (w World) BodyMoveEvents() iter.Seq[BodyMoveEvent] {
	todo()
	return nil
}

func (w World) SensorBeginEvents() iter.Seq[SensorBeginTouchEvent] {
	todo()
	return nil
}

func (w World) SensorEndEvents() iter.Seq[SensorEndTouchEvent] {
	todo()
	return nil
}

func (w World) ContactBeginEvents() iter.Seq[ContactBeginTouchEvent] {
	todo()
	return nil
}

func (w World) ContactEndEvents() iter.Seq[ContactEndTouchEvent] {
	todo()
	return nil
}

func (w World) ContactHitEvents() iter.Seq[ContactHitEvent] {
	todo()
	return nil
}

func (w World) JointEvents() iter.Seq[JointEvent] {
	todo()
	return nil
}

func (w World) IsValid() bool {
	todo()
	return false
}

func (b Body) IsValid() bool {
	todo()
	return false
}

func (s Shape) IsValid() bool {
	todo()
	return false
}

func (c Chain) IsValid() bool {
	todo()
	return false
}

func (j Joint) IsValid() bool {
	todo()
	return false
}

func (c Contact) IsValid() bool {
	todo()
	return false
}

func (w World) EnableSleeping(flag bool) {
	todo()
}

func (w World) IsSleepingEnabled() bool {
	todo()
	return false
}

func (w World) EnableWarmStarting(flag bool) {
	todo()
}

func (w World) IsWarmStartingEnabled() bool {
	todo()
	return false
}

func (w World) AwakeBodyCount() int {
	todo()
	return 0
}

func (w World) EnableContinuous(flag bool) {
	todo()
}

func (w World) IsContinuousEnabled() bool {
	todo()
	return false
}

func (w World) SetRestitutionThreshold(value float32) {
	todo()
}

func (w World) RestitutionThreshold() float32 {
	todo()
	return 0
}

func (w World) SetRestitutionIterations(iterations int) {
	todo()
}

func (w World) RestitutionIterations() int {
	todo()
	return 0
}

func (w World) EnableRestitutionPropagation(flag bool) {
	todo()
}

func (w World) IsRestitutionPropagationEnabled() bool {
	todo()
	return false
}

func (w World) SetHitEventThreshold(value float32) {
	todo()
}

func (w World) HitEventThreshold() float32 {
	todo()
	return 0
}

func (w World) SetContactTuning(hertz, dampingRatio, pushSpeed float32) {
	todo()
}

func (w World) SetContactRecycleDistance(recycleDistance float32) {
	todo()
}

func (w World) ContactRecycleDistance() float32 {
	todo()
	return 0
}

func (w World) SetMaximumLinearSpeed(maximumLinearSpeed float32) {
	todo()
}

func (w World) MaximumLinearSpeed() float32 {
	todo()
	return 0
}

func (w World) Profile() Profile {
	todo()
	return Profile{}
}

func (w World) Counters() Counters {
	todo()
	return Counters{}
}

func (w World) MaxCapacity() Capacity {
	todo()
	return Capacity{}
}

func (w World) SetUserData(userData any) {
	todo()
}

func (w World) UserData() any {
	todo()
	return nil
}

func (w World) SetFrictionCallback(callback FrictionFunc) {
	todo()
}

func (w World) SetRestitutionCallback(callback RestitutionFunc) {
	todo()
}

func (w World) SetWorkerCount(count int) {
	todo()
}

func (w World) WorkerCount() int {
	todo()
	return 0
}

func (w World) StartRecording(recording *Recording) {
	todo()
}

func (w World) StopRecording() {
	todo()
}

func (w World) SetCustomFilterCallback(fcn CustomFilterFunc) {
	todo()
}

func (w World) SetPreSolveCallback(preSolveFcn PreSolveFunc, preContinuousFcn PreContinuousFunc) {
	todo()
}

func (w World) SetGravity(gravity Vec2) {
	todo()
}

func (w World) Gravity() Vec2 {
	todo()
	return Vec2{}
}

func (w World) RebuildStaticTree() {
	todo()
}
