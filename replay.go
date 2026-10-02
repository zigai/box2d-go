// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

const (
	recKeyframeIntervalDefault int  = 16
	recKeyframeBudgetDefault   uint = 512 * 1024 * 1024
)

// The kind of a recorded spatial query, matching the public query and cast functions.
type ReplayQueryType int

const (
	ReplayQueryOverlapAABB ReplayQueryType = iota
	ReplayQueryOverlapShape
	ReplayQueryCastRay
	ReplayQueryCastShape
	ReplayQueryCollideMover
	ReplayQueryCastRayClosest
	ReplayQueryCastMover
	ReplayQueryShapeTestPoint
	ReplayQueryShapeRayCast
)

type recQueryKind int

const (
	recqOverlapAABB recQueryKind = iota
	recqOverlapShape
	recqCastRay
	recqCastShape
	recqCollideMover
	recqCastRayClosest
	recqCastMover
	recqShapeTestPoint
	recqShapeRayCast
)

// Static metadata describing a recording, resolved once when the player opens the file.
type ReplayInfo struct {
	FrameCount   int
	WorkerCount  int
	TimeStep     float32
	SubStepCount int
	LengthScale  float32
	Bounds       AABB
}

// A spatial query recorded during a replayed frame, exposed for inspection.
type ReplayQueryInfo struct {
	Type        ReplayQueryType
	Filter      QueryFilter
	AABB        AABB
	Origin      Pos
	Translation Vec2
	Shape       Shape
	HitCount    int
}

// One result of a recorded spatial query.
type ReplayQueryHit struct {
	Shape    Shape
	Point    Pos
	Normal   Vec2
	Fraction float32
}

type recRecordedHit struct {
	id          Shape
	point       Pos
	normal      Vec2
	fraction    float32
	plane       PlaneResult
	userReturnF float32
	userReturnB bool
}

type recDrawQuery struct {
	kind         int
	filter       QueryFilter
	aabb         AABB
	proxy        ShapeProxy
	mover        Capsule
	origin       Pos
	translation  Vec2
	boolResult   bool
	castFraction float32
	castOut      WorldCastOutput
	shape        Shape
	hitStart     int
	hitCount     int
}

type recReader struct {
	data             []byte
	size             int
	cursor           int
	replayWorldID    World
	workerCount      int
	ok               bool
	diverged         bool
	chainPoints      []Vec2
	chainPointCap    int
	chainMaterials   []SurfaceMaterial
	chainMaterialCap int
	hits             []recRecordedHit
	hitCap           int
	owner            *Replay
}

type recKeyframe struct {
	image         []byte
	imageSize     int
	imageCapacity int
	frame         int
	cursor        int
	bodyIDs       []Body
	bodyIDCount   int
	divergeFrame  int
	diverged      bool
}

// Opaque handle for incremental playback of a recording.
type Replay struct {
	data                 []byte
	size                 int
	headerEnd            int
	lengthScale          float32
	previousLengthScale  float32
	frame                int
	frameCount           int
	recordedWorkerCount  int
	recordedDt           float32
	recordedSubStepCount int
	bounds               AABB
	divergeFrame         int
	atEnd                bool
	rdr                  recReader
	frameQueries         []recDrawQuery
	frameQueryCount      int
	frameQueryCap        int
	frameHits            []recRecordedHit
	frameHitCount        int
	frameHitCap          int
	bodyIDs              []Body
	bodyIDCount          int
	bodyIDCap            int
	frame0Image          []byte
	frame0Size           int
	frame0BodyIDs        []Body
	frame0BodyIDCount    int
	keyframes            []recKeyframe
	keyframeCount        int
	keyframeCapacity     int
	keyframeBudget       uint
	keyframeBytes        uint
	keyframeMinInterval  int
	keyframeInterval     int
	lastKeyframeFrame    int
}

type recReplayQueryCtx struct {
	rdr    *recReader
	hits   []recRecordedHit
	count  int
	cursor int
}

func recRdrCheck(rdr *recReader, size int) {
	todo()
}

func recRU8(rdr *recReader) uint8 {
	todo()
	var result uint8
	return result
}

func recRU16(rdr *recReader) uint16 {
	todo()
	var result uint16
	return result
}

func recRU24(rdr *recReader) uint32 {
	todo()
	var result uint32
	return result
}

func recRU32(rdr *recReader) uint32 {
	todo()
	var result uint32
	return result
}

func recRU64(rdr *recReader) uint64 {
	todo()
	var result uint64
	return result
}

func recRI32(rdr *recReader) int32 {
	todo()
	var result int32
	return result
}

func recRF32(rdr *recReader) float32 {
	todo()
	var result float32
	return result
}

func recRBOOL(rdr *recReader) bool {
	todo()
	var result bool
	return result
}

func recRVEC2(rdr *recReader) Vec2 {
	todo()
	var result Vec2
	return result
}

func recRROT(rdr *recReader) Rot {
	todo()
	var result Rot
	return result
}

func recRXF(rdr *recReader) Transform {
	todo()
	var result Transform
	return result
}

func recRF64(rdr *recReader) float64 {
	todo()
	var result float64
	return result
}

func recRPOSITION(rdr *recReader) Pos {
	todo()
	var result Pos
	return result
}

func recRWORLDXF(rdr *recReader) WorldTransform {
	todo()
	var result WorldTransform
	return result
}

func recRWORLDID(rdr *recReader) World {
	todo()
	var result World
	return result
}

func recRBODYID(rdr *recReader) Body {
	todo()
	var result Body
	return result
}

func recRSHAPEID(rdr *recReader) Shape {
	todo()
	var result Shape
	return result
}

func recRCHAINID(rdr *recReader) Chain {
	todo()
	var result Chain
	return result
}

func recRJOINTID(rdr *recReader) Joint {
	todo()
	var result Joint
	return result
}

func recRdrBlob(rdr *recReader, out []byte, size int) {
	todo()
}

func recRCIRCLE(rdr *recReader) Circle {
	todo()
	var result Circle
	return result
}

func recRCAPSULE(rdr *recReader) Capsule {
	todo()
	var result Capsule
	return result
}

func recRSEGMENT(rdr *recReader) Segment {
	todo()
	var result Segment
	return result
}

func recRPOLYGON(rdr *recReader) Polygon {
	todo()
	var result Polygon
	return result
}

func recRCHAINSEG(rdr *recReader) ChainSegment {
	todo()
	var result ChainSegment
	return result
}

func recRFILTER(rdr *recReader) Filter {
	todo()
	var result Filter
	return result
}

func recRMATERIAL(rdr *recReader) SurfaceMaterial {
	todo()
	var result SurfaceMaterial
	return result
}

func recRMASSDATA(rdr *recReader) MassData {
	todo()
	var result MassData
	return result
}

func recRLOCKS(rdr *recReader) MotionLocks {
	todo()
	var result MotionLocks
	return result
}

func recRSTR(rdr *recReader) string {
	todo()
	var result string
	return result
}

func recRBODYDEF(rdr *recReader) BodyDef {
	todo()
	var result BodyDef
	return result
}

func recRSHAPEDEF(rdr *recReader) ShapeDef {
	todo()
	var result ShapeDef
	return result
}

func recRCHAINDEF(rdr *recReader) ChainDef {
	todo()
	var result ChainDef
	return result
}

func recREXPLOSIONDEF(rdr *recReader) ExplosionDef {
	todo()
	var result ExplosionDef
	return result
}

func recRJointBase(rdr *recReader, base *JointDef) {
	todo()
}

func recRDISTANCEJOINTDEF(rdr *recReader) DistanceJointDef {
	todo()
	var result DistanceJointDef
	return result
}

func recRFILTERJOINTDEF(rdr *recReader) FilterJointDef {
	todo()
	var result FilterJointDef
	return result
}

func recRMOTORJOINTDEF(rdr *recReader) MotorJointDef {
	todo()
	var result MotorJointDef
	return result
}

func recRMOVERJOINTDEF(rdr *recReader) MoverJointDef {
	todo()
	var result MoverJointDef
	return result
}

func recRPRISMATICJOINTDEF(rdr *recReader) PrismaticJointDef {
	todo()
	var result PrismaticJointDef
	return result
}

func recRPOGOJOINTDEF(rdr *recReader) PogoJointDef {
	todo()
	var result PogoJointDef
	return result
}

func recRREVOLUTEJOINTDEF(rdr *recReader) RevoluteJointDef {
	todo()
	var result RevoluteJointDef
	return result
}

func recRWELDJOINTDEF(rdr *recReader) WeldJointDef {
	todo()
	var result WeldJointDef
	return result
}

func recRWHEELJOINTDEF(rdr *recReader) WheelJointDef {
	todo()
	var result WheelJointDef
	return result
}

func recRAABB(rdr *recReader) AABB {
	todo()
	var result AABB
	return result
}

func recRQUERYFILTER(rdr *recReader) QueryFilter {
	todo()
	var result QueryFilter
	return result
}

func recRSHAPEPROXY(rdr *recReader) ShapeProxy {
	todo()
	var result ShapeProxy
	return result
}

func recRWORLDCASTOUTPUT(rdr *recReader) WorldCastOutput {
	todo()
	var result WorldCastOutput
	return result
}

func recRRAYRESULT(rdr *recReader) RayResult {
	todo()
	var result RayResult
	return result
}

func recRPLANERESULT(rdr *recReader) PlaneResult {
	todo()
	var result PlaneResult
	return result
}

func recRTREESTATS(rdr *recReader) TreeStats {
	todo()
	var result TreeStats
	return result
}

func recReserveScratch[T any](rdr *recReader, data *[]T, capacity *int, need int, elemSize int) bool {
	todo()
	var result bool
	return result
}

func recGrow[T any](data *[]T, capacity *int, need int, keep int, elemSize int) {
	todo()
}

func recEnsureHits(rdr *recReader, n int) {
	todo()
}

func recMakeBodyID(rdr *recReader, recorded Body) Body {
	todo()
	var result Body
	return result
}

func recMakeShapeID(rdr *recReader, recorded Shape) Shape {
	todo()
	var result Shape
	return result
}

func recMakeChainID(rdr *recReader, recorded Chain) Chain {
	todo()
	var result Chain
	return result
}

func recMakeJointID(rdr *recReader, recorded Joint) Joint {
	todo()
	var result Joint
	return result
}

func recCheckID(rdr *recReader, kind string, gotIndex int, gotGen uint32, recIndex int, recGen uint32) {
	todo()
}

func recCheckBodyID(rdr *recReader, got Body, rec Body) {
	todo()
}

func recCheckShapeID(rdr *recReader, got Shape, rec Shape) {
	todo()
}

func recCheckChainID(rdr *recReader, got Chain, rec Chain) {
	todo()
}

func recCheckJointID(rdr *recReader, got Joint, rec Joint) {
	todo()
}

func recDispatchDestroyWorld(a *recArgsDestroyWorld, rdr *recReader) {
	todo()
}

func recDispatchStep(a *recArgsStep, rdr *recReader) {
	todo()
}

func recDispatchWorldEnableSleeping(a *recArgsWorldEnableSleeping, rdr *recReader) {
	todo()
}

func recDispatchWorldEnableContinuous(a *recArgsWorldEnableContinuous, rdr *recReader) {
	todo()
}

func recDispatchWorldSetRestitutionThreshold(a *recArgsWorldSetRestitutionThreshold, rdr *recReader) {
	todo()
}

func recDispatchWorldSetHitEventThreshold(a *recArgsWorldSetHitEventThreshold, rdr *recReader) {
	todo()
}

func recDispatchWorldSetGravity(a *recArgsWorldSetGravity, rdr *recReader) {
	todo()
}

func recDispatchWorldExplode(a *recArgsWorldExplode, rdr *recReader) {
	todo()
}

func recDispatchWorldSetContactTuning(a *recArgsWorldSetContactTuning, rdr *recReader) {
	todo()
}

func recDispatchWorldSetContactRecycleDistance(a *recArgsWorldSetContactRecycleDistance, rdr *recReader) {
	todo()
}

func recDispatchWorldSetMaximumLinearSpeed(a *recArgsWorldSetMaximumLinearSpeed, rdr *recReader) {
	todo()
}

func recDispatchWorldEnableWarmStarting(a *recArgsWorldEnableWarmStarting, rdr *recReader) {
	todo()
}

func recDispatchWorldRebuildStaticTree(a *recArgsWorldRebuildStaticTree, rdr *recReader) {
	todo()
}

func recDispatchWorldSetRestitutionIterations(a *recArgsWorldSetRestitutionIterations, rdr *recReader) {
	todo()
}

func recDispatchWorldEnableRestitutionPropagation(a *recArgsWorldEnableRestitutionPropagation, rdr *recReader) {
	todo()
}

func recTrackBodyCreate(player *Replay, id Body) {
	todo()
}

func recTrackBodyDestroy(player *Replay, id Body) {
	todo()
}

func recSeedBodyIDs(player *Replay) {
	todo()
}

func recDispatchCreateBody(a *recArgsCreateBody, rdr *recReader) {
	todo()
}

func recDispatchDestroyBody(a *recArgsDestroyBody, rdr *recReader) {
	todo()
}

func recDispatchBodySetTransform(a *recArgsBodySetTransform, rdr *recReader) {
	todo()
}

func recDispatchBodySetLinearVelocity(a *recArgsBodySetLinearVelocity, rdr *recReader) {
	todo()
}

func recDispatchBodySetType(a *recArgsBodySetType, rdr *recReader) {
	todo()
}

func recDispatchBodySetName(a *recArgsBodySetName, rdr *recReader) {
	todo()
}

func recDispatchBodySetAngularVelocity(a *recArgsBodySetAngularVelocity, rdr *recReader) {
	todo()
}

func recDispatchBodySetTargetTransform(a *recArgsBodySetTargetTransform, rdr *recReader) {
	todo()
}

func recDispatchBodyApplyForce(a *recArgsBodyApplyForce, rdr *recReader) {
	todo()
}

func recDispatchBodyApplyForceToCenter(a *recArgsBodyApplyForceToCenter, rdr *recReader) {
	todo()
}

func recDispatchBodyApplyTorque(a *recArgsBodyApplyTorque, rdr *recReader) {
	todo()
}

func recDispatchBodyClearForces(a *recArgsBodyClearForces, rdr *recReader) {
	todo()
}

func recDispatchBodyApplyLinearImpulse(a *recArgsBodyApplyLinearImpulse, rdr *recReader) {
	todo()
}

func recDispatchBodyApplyLinearImpulseToCenter(a *recArgsBodyApplyLinearImpulseToCenter, rdr *recReader) {
	todo()
}

func recDispatchBodyApplyAngularImpulse(a *recArgsBodyApplyAngularImpulse, rdr *recReader) {
	todo()
}

func recDispatchBodySetMassData(a *recArgsBodySetMassData, rdr *recReader) {
	todo()
}

func recDispatchBodyUpdateMassFromShapes(a *recArgsBodyUpdateMassFromShapes, rdr *recReader) {
	todo()
}

func recDispatchBodySetLinearDamping(a *recArgsBodySetLinearDamping, rdr *recReader) {
	todo()
}

func recDispatchBodySetAngularDamping(a *recArgsBodySetAngularDamping, rdr *recReader) {
	todo()
}

func recDispatchBodySetGravityScale(a *recArgsBodySetGravityScale, rdr *recReader) {
	todo()
}

func recDispatchBodySetAwake(a *recArgsBodySetAwake, rdr *recReader) {
	todo()
}

func recDispatchBodyWakeTouching(a *recArgsBodyWakeTouching, rdr *recReader) {
	todo()
}

func recDispatchBodyEnableSleep(a *recArgsBodyEnableSleep, rdr *recReader) {
	todo()
}

func recDispatchBodySetSleepThreshold(a *recArgsBodySetSleepThreshold, rdr *recReader) {
	todo()
}

func recDispatchBodySetSafetyFactor(a *recArgsBodySetSafetyFactor, rdr *recReader) {
	todo()
}

func recDispatchBodyDisable(a *recArgsBodyDisable, rdr *recReader) {
	todo()
}

func recDispatchBodyEnable(a *recArgsBodyEnable, rdr *recReader) {
	todo()
}

func recDispatchBodySetMotionLocks(a *recArgsBodySetMotionLocks, rdr *recReader) {
	todo()
}

func recDispatchBodySetBullet(a *recArgsBodySetBullet, rdr *recReader) {
	todo()
}

func recDispatchBodyEnableContactRecycling(a *recArgsBodyEnableContactRecycling, rdr *recReader) {
	todo()
}

func recDispatchBodyEnableContactEvents(a *recArgsBodyEnableContactEvents, rdr *recReader) {
	todo()
}

func recDispatchBodyEnableHitEvents(a *recArgsBodyEnableHitEvents, rdr *recReader) {
	todo()
}

func recDispatchCreateCircleShape(a *recArgsCreateCircleShape, rdr *recReader) {
	todo()
}

func recDispatchCreateCapsuleShape(a *recArgsCreateCapsuleShape, rdr *recReader) {
	todo()
}

func recDispatchCreateSegmentShape(a *recArgsCreateSegmentShape, rdr *recReader) {
	todo()
}

func recDispatchCreatePolygonShape(a *recArgsCreatePolygonShape, rdr *recReader) {
	todo()
}

func recDispatchCreateChainSegmentShape(a *recArgsCreateChainSegmentShape, rdr *recReader) {
	todo()
}

func recDispatchDestroyShape(a *recArgsDestroyShape, rdr *recReader) {
	todo()
}

func recDispatchShapeSetDensity(a *recArgsShapeSetDensity, rdr *recReader) {
	todo()
}

func recDispatchShapeSetFriction(a *recArgsShapeSetFriction, rdr *recReader) {
	todo()
}

func recDispatchShapeSetRestitution(a *recArgsShapeSetRestitution, rdr *recReader) {
	todo()
}

func recDispatchShapeSetUserMaterial(a *recArgsShapeSetUserMaterial, rdr *recReader) {
	todo()
}

func recDispatchShapeSetSurfaceMaterial(a *recArgsShapeSetSurfaceMaterial, rdr *recReader) {
	todo()
}

func recDispatchShapeSetFilter(a *recArgsShapeSetFilter, rdr *recReader) {
	todo()
}

func recDispatchShapeEnableSensorEvents(a *recArgsShapeEnableSensorEvents, rdr *recReader) {
	todo()
}

func recDispatchShapeEnableContactEvents(a *recArgsShapeEnableContactEvents, rdr *recReader) {
	todo()
}

func recDispatchShapeEnablePreSolveEvents(a *recArgsShapeEnablePreSolveEvents, rdr *recReader) {
	todo()
}

func recDispatchShapeEnableHitEvents(a *recArgsShapeEnableHitEvents, rdr *recReader) {
	todo()
}

func recDispatchShapeSetCircle(a *recArgsShapeSetCircle, rdr *recReader) {
	todo()
}

func recDispatchShapeSetCapsule(a *recArgsShapeSetCapsule, rdr *recReader) {
	todo()
}

func recDispatchShapeSetSegment(a *recArgsShapeSetSegment, rdr *recReader) {
	todo()
}

func recDispatchShapeSetPolygon(a *recArgsShapeSetPolygon, rdr *recReader) {
	todo()
}

func recDispatchShapeSetChainSegment(a *recArgsShapeSetChainSegment, rdr *recReader) {
	todo()
}

func recDispatchShapeApplyWind(a *recArgsShapeApplyWind, rdr *recReader) {
	todo()
}

func recDispatchCreateChain(a *recArgsCreateChain, rdr *recReader) {
	todo()
}

func recDispatchDestroyChain(a *recArgsDestroyChain, rdr *recReader) {
	todo()
}

func recDispatchChainSetSurfaceMaterial(a *recArgsChainSetSurfaceMaterial, rdr *recReader) {
	todo()
}

func recDispatchChainSetAllSurfaceMaterials(a *recArgsChainSetAllSurfaceMaterials, rdr *recReader) {
	todo()
}

func recDispatchCreateDistanceJoint(a *recArgsCreateDistanceJoint, rdr *recReader) {
	todo()
}

func recDispatchCreateFilterJoint(a *recArgsCreateFilterJoint, rdr *recReader) {
	todo()
}

func recDispatchCreateMotorJoint(a *recArgsCreateMotorJoint, rdr *recReader) {
	todo()
}

func recDispatchCreateMoverJoint(a *recArgsCreateMoverJoint, rdr *recReader) {
	todo()
}

func recDispatchCreatePrismaticJoint(a *recArgsCreatePrismaticJoint, rdr *recReader) {
	todo()
}

func recDispatchCreatePogoJoint(a *recArgsCreatePogoJoint, rdr *recReader) {
	todo()
}

func recDispatchCreateRevoluteJoint(a *recArgsCreateRevoluteJoint, rdr *recReader) {
	todo()
}

func recDispatchCreateWeldJoint(a *recArgsCreateWeldJoint, rdr *recReader) {
	todo()
}

func recDispatchCreateWheelJoint(a *recArgsCreateWheelJoint, rdr *recReader) {
	todo()
}

func recDispatchDestroyJoint(a *recArgsDestroyJoint, rdr *recReader) {
	todo()
}

func recDispatchJointSetLocalFrameA(a *recArgsJointSetLocalFrameA, rdr *recReader) {
	todo()
}

func recDispatchJointSetLocalFrameB(a *recArgsJointSetLocalFrameB, rdr *recReader) {
	todo()
}

func recDispatchJointSetCollideConnected(a *recArgsJointSetCollideConnected, rdr *recReader) {
	todo()
}

func recDispatchJointWakeBodies(a *recArgsJointWakeBodies, rdr *recReader) {
	todo()
}

func recDispatchJointSetConstraintTuning(a *recArgsJointSetConstraintTuning, rdr *recReader) {
	todo()
}

func recDispatchJointSetForceThreshold(a *recArgsJointSetForceThreshold, rdr *recReader) {
	todo()
}

func recDispatchJointSetTorqueThreshold(a *recArgsJointSetTorqueThreshold, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetLength(a *recArgsDistanceJointSetLength, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointEnableSpring(a *recArgsDistanceJointEnableSpring, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetSpringForceRange(a *recArgsDistanceJointSetSpringForceRange, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetSpringHertz(a *recArgsDistanceJointSetSpringHertz, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetSpringDampingRatio(a *recArgsDistanceJointSetSpringDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointEnableLimit(a *recArgsDistanceJointEnableLimit, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetLengthRange(a *recArgsDistanceJointSetLengthRange, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointEnableMotor(a *recArgsDistanceJointEnableMotor, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetMotorSpeed(a *recArgsDistanceJointSetMotorSpeed, rdr *recReader) {
	todo()
}

func recDispatchDistanceJointSetMaxMotorForce(a *recArgsDistanceJointSetMaxMotorForce, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetLinearVelocity(a *recArgsMotorJointSetLinearVelocity, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetAngularVelocity(a *recArgsMotorJointSetAngularVelocity, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetMaxVelocityForce(a *recArgsMotorJointSetMaxVelocityForce, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetMaxVelocityTorque(a *recArgsMotorJointSetMaxVelocityTorque, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetLinearHertz(a *recArgsMotorJointSetLinearHertz, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetLinearDampingRatio(a *recArgsMotorJointSetLinearDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetAngularHertz(a *recArgsMotorJointSetAngularHertz, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetAngularDampingRatio(a *recArgsMotorJointSetAngularDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetMaxSpringForce(a *recArgsMotorJointSetMaxSpringForce, rdr *recReader) {
	todo()
}

func recDispatchMotorJointSetMaxSpringTorque(a *recArgsMotorJointSetMaxSpringTorque, rdr *recReader) {
	todo()
}

func recDispatchMoverJointSetLinearVelocity(a *recArgsMoverJointSetLinearVelocity, rdr *recReader) {
	todo()
}

func recDispatchMoverJointSetMaxVelocityForce(a *recArgsMoverJointSetMaxVelocityForce, rdr *recReader) {
	todo()
}

func recDispatchPogoJointSetRestLength(a *recArgsPogoJointSetRestLength, rdr *recReader) {
	todo()
}

func recDispatchPogoJointSetSpringHertz(a *recArgsPogoJointSetSpringHertz, rdr *recReader) {
	todo()
}

func recDispatchPogoJointSetSpringDampingRatio(a *recArgsPogoJointSetSpringDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointEnableSpring(a *recArgsPrismaticJointEnableSpring, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointSetSpringHertz(a *recArgsPrismaticJointSetSpringHertz, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointSetSpringDampingRatio(a *recArgsPrismaticJointSetSpringDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointSetTargetTranslation(a *recArgsPrismaticJointSetTargetTranslation, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointEnableLimit(a *recArgsPrismaticJointEnableLimit, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointSetLimits(a *recArgsPrismaticJointSetLimits, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointEnableMotor(a *recArgsPrismaticJointEnableMotor, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointSetMotorSpeed(a *recArgsPrismaticJointSetMotorSpeed, rdr *recReader) {
	todo()
}

func recDispatchPrismaticJointSetMaxMotorForce(a *recArgsPrismaticJointSetMaxMotorForce, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointEnableSpring(a *recArgsRevoluteJointEnableSpring, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointSetSpringHertz(a *recArgsRevoluteJointSetSpringHertz, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointSetSpringDampingRatio(a *recArgsRevoluteJointSetSpringDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointSetTargetAngle(a *recArgsRevoluteJointSetTargetAngle, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointEnableLimit(a *recArgsRevoluteJointEnableLimit, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointSetLimits(a *recArgsRevoluteJointSetLimits, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointEnableMotor(a *recArgsRevoluteJointEnableMotor, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointSetMotorSpeed(a *recArgsRevoluteJointSetMotorSpeed, rdr *recReader) {
	todo()
}

func recDispatchRevoluteJointSetMaxMotorTorque(a *recArgsRevoluteJointSetMaxMotorTorque, rdr *recReader) {
	todo()
}

func recDispatchWeldJointSetLinearHertz(a *recArgsWeldJointSetLinearHertz, rdr *recReader) {
	todo()
}

func recDispatchWeldJointSetLinearDampingRatio(a *recArgsWeldJointSetLinearDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchWeldJointSetAngularHertz(a *recArgsWeldJointSetAngularHertz, rdr *recReader) {
	todo()
}

func recDispatchWeldJointSetAngularDampingRatio(a *recArgsWeldJointSetAngularDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchWheelJointEnableSpring(a *recArgsWheelJointEnableSpring, rdr *recReader) {
	todo()
}

func recDispatchWheelJointSetSpringHertz(a *recArgsWheelJointSetSpringHertz, rdr *recReader) {
	todo()
}

func recDispatchWheelJointSetSpringDampingRatio(a *recArgsWheelJointSetSpringDampingRatio, rdr *recReader) {
	todo()
}

func recDispatchWheelJointEnableLimit(a *recArgsWheelJointEnableLimit, rdr *recReader) {
	todo()
}

func recDispatchWheelJointSetLimits(a *recArgsWheelJointSetLimits, rdr *recReader) {
	todo()
}

func recDispatchWheelJointEnableMotor(a *recArgsWheelJointEnableMotor, rdr *recReader) {
	todo()
}

func recDispatchWheelJointSetMotorSpeed(a *recArgsWheelJointSetMotorSpeed, rdr *recReader) {
	todo()
}

func recDispatchWheelJointSetMaxMotorTorque(a *recArgsWheelJointSetMaxMotorTorque, rdr *recReader) {
	todo()
}

func recF32Differs(a float32, b float32) bool {
	todo()
	var result bool
	return result
}

func recVec2Differs(a Vec2, b Vec2) bool {
	todo()
	var result bool
	return result
}

func recGrowFrameQueries(player *Replay) {
	todo()
}

func recGrowFrameHits(player *Replay, need int) {
	todo()
}

func recStashQueryBegin(player *Replay, kind int, hits []recRecordedHit, hitCount int) *recDrawQuery {
	todo()
	var result *recDrawQuery
	return result
}

func recReplayOverlapTrampoline(id Shape, ctx *recReplayQueryCtx) bool {
	todo()
	var result bool
	return result
}

func recDispatchQueryOverlapAABB(a *recArgsQueryOverlapAABB, rdr *recReader) {
	todo()
}

func recDispatchQueryOverlapShape(a *recArgsQueryOverlapShape, rdr *recReader) {
	todo()
}

func recReplayCastTrampoline(id Shape, point Pos, normal Vec2, fraction float32, ctx *recReplayQueryCtx) float32 {
	todo()
	var result float32
	return result
}

func recDispatchQueryCastRay(a *recArgsQueryCastRay, rdr *recReader) {
	todo()
}

func recDispatchQueryCastShape(a *recArgsQueryCastShape, rdr *recReader) {
	todo()
}

func recReplayPlaneTrampoline(id Shape, plane *PlaneResult, ctx *recReplayQueryCtx) bool {
	todo()
	var result bool
	return result
}

func recDispatchQueryCollideMover(a *recArgsQueryCollideMover, rdr *recReader) {
	todo()
}

func recDispatchQueryCastRayClosest(a *recArgsQueryCastRayClosest, rdr *recReader) {
	todo()
}

func recDispatchQueryCastMover(a *recArgsQueryCastMover, rdr *recReader) {
	todo()
}

func recDispatchShapeTestPoint(a *recArgsShapeTestPoint, rdr *recReader) {
	todo()
}

func recDispatchShapeRayCast(a *recArgsShapeRayCast, rdr *recReader) {
	todo()
}

func recDispatchStateHash(a *recArgsStateHash, rdr *recReader) {
	todo()
}

func recDispatchRecordingBounds(a *recArgsRecordingBounds, rdr *recReader) {
	todo()
}

func recDispatchOne(player *Replay) int {
	todo()
	var result int
	return result
}

func recScanFile(player *Replay) {
	todo()
}

// Open a recording for incremental playback. The player copies the bytes,
// so you may free or destroy the source buffer immediately after this call.
func CreateReplay(data []byte, workerCount int) (*Replay, error) {
	todo()
	return nil, nil
}

func freeKeyframe(kf *recKeyframe) {
	todo()
}

func recCaptureKeyframe(player *Replay) {
	todo()
}

func replayRestoreKeyframe(player *Replay, kf *recKeyframe) {
	todo()
}

// Advance the replay by one recorded step.
func (r *Replay) StepFrame() bool {
	todo()
	var result bool
	return result
}

// Rewind the player to the first step, recreating the replay world from the file.
func (r *Replay) Restart() {
	todo()
}

// Get the id of the replayed world.
func (r *Replay) WorldID() World {
	todo()
	var result World
	return result
}

// Get the number of steps replayed so far.
func (r *Replay) Frame() int {
	todo()
	var result int
	return result
}

// Seek to a recorded step. Seeking backward rewinds and re-runs from the start, so the
// cost grows with the target frame. Clamps to the recording bounds.
func (r *Replay) SeekFrame(targetFrame int) {
	todo()
}

// Get static metadata for the recording (frame count, recorded tuning, time).
func (r *Replay) Info() ReplayInfo {
	todo()
	var result ReplayInfo
	return result
}

// Returns true once the end of the recording has been reached.
func (r *Replay) IsAtEnd() bool {
	todo()
	var result bool
	return result
}

// Returns true if a recorded state hash failed to reproduce, meaning replay diverged.
func (r *Replay) HasDiverged() bool {
	todo()
	var result bool
	return result
}

// Get the first step at which replay diverged, or -1 if it has not diverged.
func (r *Replay) DivergeFrame() int {
	todo()
	var result int
	return result
}

// Tune the keyframe ring used to speed up backward seeking. A keyframe is a periodic snapshot the
// player restores from instead of replaying from the start, trading memory for seek speed.
// A zero budget or a non-positive interval keeps that value. Clears the existing ring, so call
// Restart afterward to repopulate it under the new policy.
func (r *Replay) SetKeyframePolicy(budgetBytes uint, minIntervalFrames int) {
	todo()
}

// Get the keyframe memory budget in bytes.
func (r *Replay) KeyframeBudget() uint {
	todo()
	var result uint
	return result
}

// Get the finest keyframe spacing in frames.
func (r *Replay) KeyframeMinInterval() int {
	todo()
	var result int
	return result
}

// Get the current keyframe spacing in frames. Starts at the min interval and doubles as the ring
// evicts to stay under budget, so it reflects the effective backward-seek granularity right now.
func (r *Replay) KeyframeInterval() int {
	todo()
	var result int
	return result
}

// Get the memory currently held by keyframe snapshots, in bytes.
func (r *Replay) KeyframeBytes() uint {
	todo()
	var result uint
	return result
}

// Close a player and free its replay world and file buffer.
func (r *Replay) Destroy() {
	todo()
}

func recDrawHitAABBs(player *Replay, q *recDrawQuery, draw *DebugDraw) {
	todo()
}

// Draw spatial queries recorded during the most recently replayed frame.
// Call after World_Draw so queries are layered on top of the world.
func (r *Replay) DrawFrameQueries(draw *DebugDraw, queryIndex int) {
	todo()
}

// Get the number of spatial queries recorded for the most recently replayed frame.
func (r *Replay) FrameQueryCount() int {
	todo()
	var result int
	return result
}

// Get a recorded query from the most recently replayed frame by index.
func (r *Replay) FrameQuery(index int) ReplayQueryInfo {
	todo()
	var result ReplayQueryInfo
	return result
}

// Get one result of a recorded query from the most recently replayed frame.
func (r *Replay) FrameQueryHit(queryIndex int, hitIndex int) ReplayQueryHit {
	todo()
	var result ReplayQueryHit
	return result
}

// Get the number of body slots tracked for the outliner. This is the creation-order span and
// includes holes for destroyed bodies, so it only grows as the replay advances.
func (r *Replay) BodyCount() int {
	todo()
	var result int
	return result
}

// Get a tracked body by creation ordinal. Returns b2_nullBodyId for a destroyed slot, an ordinal not
// yet reached at the current frame, or an out-of-range index. Validate with Body_IsValid.
func (r *Replay) BodyID(index int) Body {
	todo()
	var result Body
	return result
}

// Replay a recording by re-running the engine and asserting recorded ids and state match.
func ValidateReplay(data []byte, workerCount int) bool {
	todo()
	var result bool
	return result
}
