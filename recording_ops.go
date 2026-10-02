// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

const (
	recOpDestroyWorld                        uint8 = 0x01
	recOpStep                                uint8 = 0x80
	recOpWorldEnableSleeping                 uint8 = 0x02
	recOpWorldEnableContinuous               uint8 = 0x03
	recOpWorldSetRestitutionThreshold        uint8 = 0x04
	recOpWorldSetHitEventThreshold           uint8 = 0x05
	recOpWorldSetGravity                     uint8 = 0x06
	recOpWorldExplode                        uint8 = 0x07
	recOpWorldSetContactTuning               uint8 = 0x08
	recOpWorldSetContactRecycleDistance      uint8 = 0x09
	recOpWorldSetMaximumLinearSpeed          uint8 = 0x0A
	recOpWorldEnableWarmStarting             uint8 = 0x0B
	recOpWorldRebuildStaticTree              uint8 = 0x0C
	recOpWorldSetRestitutionIterations       uint8 = 0x0D
	recOpWorldEnableRestitutionPropagation   uint8 = 0x0E
	recOpCreateBody                          uint8 = 0x10
	recOpDestroyBody                         uint8 = 0x11
	recOpBodySetTransform                    uint8 = 0x20
	recOpBodySetLinearVelocity               uint8 = 0x21
	recOpBodySetType                         uint8 = 0x22
	recOpBodySetName                         uint8 = 0x23
	recOpBodySetAngularVelocity              uint8 = 0x24
	recOpBodySetTargetTransform              uint8 = 0x25
	recOpBodyApplyForce                      uint8 = 0x26
	recOpBodyApplyForceToCenter              uint8 = 0x27
	recOpBodyApplyTorque                     uint8 = 0x28
	recOpBodyClearForces                     uint8 = 0x29
	recOpBodyApplyLinearImpulse              uint8 = 0x2A
	recOpBodyApplyLinearImpulseToCenter      uint8 = 0x2B
	recOpBodyApplyAngularImpulse             uint8 = 0x2C
	recOpBodySetMassData                     uint8 = 0x2D
	recOpBodyUpdateMassFromShapes            uint8 = 0x2E
	recOpBodySetLinearDamping                uint8 = 0x2F
	recOpBodySetAngularDamping               uint8 = 0x30
	recOpBodySetGravityScale                 uint8 = 0x31
	recOpBodySetAwake                        uint8 = 0x32
	recOpBodyWakeTouching                    uint8 = 0x33
	recOpBodyEnableSleep                     uint8 = 0x34
	recOpBodySetSleepThreshold               uint8 = 0x35
	recOpBodyDisable                         uint8 = 0x36
	recOpBodyEnable                          uint8 = 0x37
	recOpBodySetMotionLocks                  uint8 = 0x38
	recOpBodySetBullet                       uint8 = 0x39
	recOpBodyEnableContactRecycling          uint8 = 0x3A
	recOpBodyEnableContactEvents             uint8 = 0x3B
	recOpBodyEnableHitEvents                 uint8 = 0x3C
	recOpBodySetSafetyFactor                 uint8 = 0x3D
	recOpCreateCircleShape                   uint8 = 0x40
	recOpCreateCapsuleShape                  uint8 = 0x41
	recOpCreateSegmentShape                  uint8 = 0x42
	recOpCreatePolygonShape                  uint8 = 0x43
	recOpCreateChainSegmentShape             uint8 = 0x44
	recOpDestroyShape                        uint8 = 0x45
	recOpShapeSetDensity                     uint8 = 0x50
	recOpShapeSetFriction                    uint8 = 0x51
	recOpShapeSetRestitution                 uint8 = 0x52
	recOpShapeSetUserMaterial                uint8 = 0x53
	recOpShapeSetSurfaceMaterial             uint8 = 0x54
	recOpShapeSetFilter                      uint8 = 0x55
	recOpShapeEnableSensorEvents             uint8 = 0x56
	recOpShapeEnableContactEvents            uint8 = 0x57
	recOpShapeEnablePreSolveEvents           uint8 = 0x58
	recOpShapeEnableHitEvents                uint8 = 0x59
	recOpShapeSetCircle                      uint8 = 0x5A
	recOpShapeSetCapsule                     uint8 = 0x5B
	recOpShapeSetSegment                     uint8 = 0x5C
	recOpShapeSetPolygon                     uint8 = 0x5D
	recOpShapeSetChainSegment                uint8 = 0x5E
	recOpShapeApplyWind                      uint8 = 0x5F
	recOpCreateChain                         uint8 = 0x70
	recOpDestroyChain                        uint8 = 0x71
	recOpChainSetSurfaceMaterial             uint8 = 0x72
	recOpChainSetAllSurfaceMaterials         uint8 = 0x73
	recOpDestroyJoint                        uint8 = 0x90
	recOpJointSetLocalFrameA                 uint8 = 0x91
	recOpJointSetLocalFrameB                 uint8 = 0x92
	recOpJointSetCollideConnected            uint8 = 0x93
	recOpJointWakeBodies                     uint8 = 0x94
	recOpJointSetConstraintTuning            uint8 = 0x95
	recOpJointSetForceThreshold              uint8 = 0x96
	recOpJointSetTorqueThreshold             uint8 = 0x97
	recOpCreateDistanceJoint                 uint8 = 0x98
	recOpDistanceJointSetLength              uint8 = 0x99
	recOpDistanceJointEnableSpring           uint8 = 0x9A
	recOpDistanceJointSetSpringForceRange    uint8 = 0x9B
	recOpDistanceJointSetSpringHertz         uint8 = 0x9C
	recOpDistanceJointSetSpringDampingRatio  uint8 = 0x9D
	recOpDistanceJointEnableLimit            uint8 = 0x9E
	recOpDistanceJointSetLengthRange         uint8 = 0x9F
	recOpDistanceJointEnableMotor            uint8 = 0xA0
	recOpDistanceJointSetMotorSpeed          uint8 = 0xA1
	recOpDistanceJointSetMaxMotorForce       uint8 = 0xA2
	recOpCreateFilterJoint                   uint8 = 0xA3
	recOpCreateMotorJoint                    uint8 = 0xA4
	recOpMotorJointSetLinearVelocity         uint8 = 0xA5
	recOpMotorJointSetAngularVelocity        uint8 = 0xA6
	recOpMotorJointSetMaxVelocityForce       uint8 = 0xA7
	recOpMotorJointSetMaxVelocityTorque      uint8 = 0xA8
	recOpMotorJointSetLinearHertz            uint8 = 0xA9
	recOpMotorJointSetLinearDampingRatio     uint8 = 0xAA
	recOpMotorJointSetAngularHertz           uint8 = 0xAB
	recOpMotorJointSetAngularDampingRatio    uint8 = 0xAC
	recOpMotorJointSetMaxSpringForce         uint8 = 0xAD
	recOpMotorJointSetMaxSpringTorque        uint8 = 0xAE
	recOpCreateMoverJoint                    uint8 = 0xAF
	recOpMoverJointSetLinearVelocity         uint8 = 0xB0
	recOpMoverJointSetMaxVelocityForce       uint8 = 0xB1
	recOpCreatePogoJoint                     uint8 = 0xB2
	recOpPogoJointSetRestLength              uint8 = 0xB3
	recOpPogoJointSetSpringHertz             uint8 = 0xB4
	recOpPogoJointSetSpringDampingRatio      uint8 = 0xB5
	recOpCreatePrismaticJoint                uint8 = 0xB6
	recOpPrismaticJointEnableSpring          uint8 = 0xB7
	recOpPrismaticJointSetSpringHertz        uint8 = 0xB8
	recOpPrismaticJointSetSpringDampingRatio uint8 = 0xB9
	recOpPrismaticJointSetTargetTranslation  uint8 = 0xBA
	recOpPrismaticJointEnableLimit           uint8 = 0xBB
	recOpPrismaticJointSetLimits             uint8 = 0xBC
	recOpPrismaticJointEnableMotor           uint8 = 0xBD
	recOpPrismaticJointSetMotorSpeed         uint8 = 0xBE
	recOpPrismaticJointSetMaxMotorForce      uint8 = 0xBF
	recOpCreateRevoluteJoint                 uint8 = 0xC0
	recOpRevoluteJointEnableSpring           uint8 = 0xC1
	recOpRevoluteJointSetSpringHertz         uint8 = 0xC2
	recOpRevoluteJointSetSpringDampingRatio  uint8 = 0xC3
	recOpRevoluteJointSetTargetAngle         uint8 = 0xC4
	recOpRevoluteJointEnableLimit            uint8 = 0xC5
	recOpRevoluteJointSetLimits              uint8 = 0xC6
	recOpRevoluteJointEnableMotor            uint8 = 0xC7
	recOpRevoluteJointSetMotorSpeed          uint8 = 0xC8
	recOpRevoluteJointSetMaxMotorTorque      uint8 = 0xC9
	recOpCreateWeldJoint                     uint8 = 0xCA
	recOpWeldJointSetLinearHertz             uint8 = 0xCB
	recOpWeldJointSetLinearDampingRatio      uint8 = 0xCC
	recOpWeldJointSetAngularHertz            uint8 = 0xCD
	recOpWeldJointSetAngularDampingRatio     uint8 = 0xCE
	recOpCreateWheelJoint                    uint8 = 0xCF
	recOpWheelJointEnableSpring              uint8 = 0xD0
	recOpWheelJointSetSpringHertz            uint8 = 0xD1
	recOpWheelJointSetSpringDampingRatio     uint8 = 0xD2
	recOpWheelJointEnableLimit               uint8 = 0xD3
	recOpWheelJointSetLimits                 uint8 = 0xD4
	recOpWheelJointEnableMotor               uint8 = 0xD5
	recOpWheelJointSetMotorSpeed             uint8 = 0xD6
	recOpWheelJointSetMaxMotorTorque         uint8 = 0xD7
	recOpQueryOverlapAABB                    uint8 = 0xE0
	recOpQueryOverlapShape                   uint8 = 0xE1
	recOpQueryCastRay                        uint8 = 0xE2
	recOpQueryCastShape                      uint8 = 0xE3
	recOpQueryCollideMover                   uint8 = 0xE4
	recOpQueryCastRayClosest                 uint8 = 0xE5
	recOpQueryCastMover                      uint8 = 0xE6
	recOpShapeTestPoint                      uint8 = 0xE7
	recOpShapeRayCast                        uint8 = 0xE8
	recOpStateHash                           uint8 = 0xF1
	recOpRecordingBounds                     uint8 = 0xF2
)

type recArgsDestroyWorld struct {
	world World
}

type recArgsStep struct {
	world        World
	dt           float32
	subStepCount int32
}

type recArgsWorldEnableSleeping struct {
	world World
	flag  bool
}

type recArgsWorldEnableContinuous struct {
	world World
	flag  bool
}

type recArgsWorldSetRestitutionThreshold struct {
	world World
	value float32
}

type recArgsWorldSetHitEventThreshold struct {
	world World
	value float32
}

type recArgsWorldSetGravity struct {
	world   World
	gravity Vec2
}

type recArgsWorldExplode struct {
	world World
	def   ExplosionDef
}

type recArgsWorldSetContactTuning struct {
	world        World
	hertz        float32
	dampingRatio float32
	pushSpeed    float32
}

type recArgsWorldSetContactRecycleDistance struct {
	world           World
	recycleDistance float32
}

type recArgsWorldSetMaximumLinearSpeed struct {
	world              World
	maximumLinearSpeed float32
}

type recArgsWorldEnableWarmStarting struct {
	world World
	flag  bool
}

type recArgsWorldRebuildStaticTree struct {
	world World
}

type recArgsWorldSetRestitutionIterations struct {
	world      World
	iterations int32
}

type recArgsWorldEnableRestitutionPropagation struct {
	world World
	flag  bool
}

type recArgsCreateBody struct {
	world World
	def   BodyDef
}

type recArgsDestroyBody struct {
	body Body
}

type recArgsBodySetTransform struct {
	body     Body
	position Pos
	rotation Rot
}

type recArgsBodySetLinearVelocity struct {
	body Body
	v    Vec2
}

type recArgsBodySetType struct {
	body     Body
	bodyType int32
}

type recArgsBodySetName struct {
	body Body
	name string
}

type recArgsBodySetAngularVelocity struct {
	body Body
	w    float32
}

type recArgsBodySetTargetTransform struct {
	body     Body
	target   WorldTransform
	timeStep float32
	wake     bool
}

type recArgsBodyApplyForce struct {
	body  Body
	force Vec2
	point Pos
	wake  bool
}

type recArgsBodyApplyForceToCenter struct {
	body  Body
	force Vec2
	wake  bool
}

type recArgsBodyApplyTorque struct {
	body   Body
	torque float32
	wake   bool
}

type recArgsBodyClearForces struct {
	body Body
}

type recArgsBodyApplyLinearImpulse struct {
	body    Body
	impulse Vec2
	point   Pos
	wake    bool
}

type recArgsBodyApplyLinearImpulseToCenter struct {
	body    Body
	impulse Vec2
	wake    bool
}

type recArgsBodyApplyAngularImpulse struct {
	body    Body
	impulse float32
	wake    bool
}

type recArgsBodySetMassData struct {
	body     Body
	massData MassData
}

type recArgsBodyUpdateMassFromShapes struct {
	body Body
}

type recArgsBodySetLinearDamping struct {
	body    Body
	damping float32
}

type recArgsBodySetAngularDamping struct {
	body    Body
	damping float32
}

type recArgsBodySetGravityScale struct {
	body  Body
	scale float32
}

type recArgsBodySetAwake struct {
	body  Body
	awake bool
}

type recArgsBodyWakeTouching struct {
	body Body
}

type recArgsBodyEnableSleep struct {
	body Body
	flag bool
}

type recArgsBodySetSleepThreshold struct {
	body      Body
	threshold float32
}

type recArgsBodyDisable struct {
	body Body
}

type recArgsBodyEnable struct {
	body Body
}

type recArgsBodySetMotionLocks struct {
	body  Body
	locks MotionLocks
}

type recArgsBodySetBullet struct {
	body Body
	flag bool
}

type recArgsBodyEnableContactRecycling struct {
	body Body
	flag bool
}

type recArgsBodyEnableContactEvents struct {
	body Body
	flag bool
}

type recArgsBodyEnableHitEvents struct {
	body Body
	flag bool
}

type recArgsBodySetSafetyFactor struct {
	body  Body
	value float32
}

type recArgsCreateCircleShape struct {
	body   Body
	def    ShapeDef
	circle Circle
}

type recArgsCreateCapsuleShape struct {
	body    Body
	def     ShapeDef
	capsule Capsule
}

type recArgsCreateSegmentShape struct {
	body    Body
	def     ShapeDef
	segment Segment
}

type recArgsCreatePolygonShape struct {
	body    Body
	def     ShapeDef
	polygon Polygon
}

type recArgsCreateChainSegmentShape struct {
	body         Body
	def          ShapeDef
	chainSegment ChainSegment
}

type recArgsDestroyShape struct {
	shape          Shape
	updateBodyMass bool
}

type recArgsShapeSetDensity struct {
	shape          Shape
	density        float32
	updateBodyMass bool
}

type recArgsShapeSetFriction struct {
	shape    Shape
	friction float32
}

type recArgsShapeSetRestitution struct {
	shape       Shape
	restitution float32
}

type recArgsShapeSetUserMaterial struct {
	shape    Shape
	material uint64
}

type recArgsShapeSetSurfaceMaterial struct {
	shape    Shape
	material SurfaceMaterial
}

type recArgsShapeSetFilter struct {
	shape  Shape
	filter Filter
}

type recArgsShapeEnableSensorEvents struct {
	shape Shape
	flag  bool
}

type recArgsShapeEnableContactEvents struct {
	shape Shape
	flag  bool
}

type recArgsShapeEnablePreSolveEvents struct {
	shape Shape
	flag  bool
}

type recArgsShapeEnableHitEvents struct {
	shape Shape
	flag  bool
}

type recArgsShapeSetCircle struct {
	shape  Shape
	circle Circle
}

type recArgsShapeSetCapsule struct {
	shape   Shape
	capsule Capsule
}

type recArgsShapeSetSegment struct {
	shape   Shape
	segment Segment
}

type recArgsShapeSetPolygon struct {
	shape   Shape
	polygon Polygon
}

type recArgsShapeSetChainSegment struct {
	shape        Shape
	chainSegment ChainSegment
}

type recArgsShapeApplyWind struct {
	shape Shape
	wind  Vec2
	drag  float32
	lift  float32
	wake  bool
}

type recArgsCreateChain struct {
	body Body
	def  ChainDef
}

type recArgsDestroyChain struct {
	chain Chain
}

type recArgsChainSetSurfaceMaterial struct {
	chain        Chain
	material     SurfaceMaterial
	segmentIndex int32
}

type recArgsChainSetAllSurfaceMaterials struct {
	chain    Chain
	material SurfaceMaterial
}

type recArgsDestroyJoint struct {
	joint Joint
}

type recArgsJointSetLocalFrameA struct {
	joint      Joint
	localFrame Transform
}

type recArgsJointSetLocalFrameB struct {
	joint      Joint
	localFrame Transform
}

type recArgsJointSetCollideConnected struct {
	joint         Joint
	shouldCollide bool
}

type recArgsJointWakeBodies struct {
	joint Joint
}

type recArgsJointSetConstraintTuning struct {
	joint        Joint
	hertz        float32
	dampingRatio float32
}

type recArgsJointSetForceThreshold struct {
	joint     Joint
	threshold float32
}

type recArgsJointSetTorqueThreshold struct {
	joint     Joint
	threshold float32
}

type recArgsCreateDistanceJoint struct {
	world World
	def   DistanceJointDef
}

type recArgsDistanceJointSetLength struct {
	joint  Joint
	length float32
}

type recArgsDistanceJointEnableSpring struct {
	joint        Joint
	enableSpring bool
}

type recArgsDistanceJointSetSpringForceRange struct {
	joint      Joint
	lowerForce float32
	upperForce float32
}

type recArgsDistanceJointSetSpringHertz struct {
	joint Joint
	hertz float32
}

type recArgsDistanceJointSetSpringDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsDistanceJointEnableLimit struct {
	joint       Joint
	enableLimit bool
}

type recArgsDistanceJointSetLengthRange struct {
	joint     Joint
	minLength float32
	maxLength float32
}

type recArgsDistanceJointEnableMotor struct {
	joint       Joint
	enableMotor bool
}

type recArgsDistanceJointSetMotorSpeed struct {
	joint      Joint
	motorSpeed float32
}

type recArgsDistanceJointSetMaxMotorForce struct {
	joint Joint
	force float32
}

type recArgsCreateFilterJoint struct {
	world World
	def   FilterJointDef
}

type recArgsCreateMotorJoint struct {
	world World
	def   MotorJointDef
}

type recArgsMotorJointSetLinearVelocity struct {
	joint    Joint
	velocity Vec2
}

type recArgsMotorJointSetAngularVelocity struct {
	joint    Joint
	velocity float32
}

type recArgsMotorJointSetMaxVelocityForce struct {
	joint    Joint
	maxForce float32
}

type recArgsMotorJointSetMaxVelocityTorque struct {
	joint     Joint
	maxTorque float32
}

type recArgsMotorJointSetLinearHertz struct {
	joint Joint
	hertz float32
}

type recArgsMotorJointSetLinearDampingRatio struct {
	joint   Joint
	damping float32
}

type recArgsMotorJointSetAngularHertz struct {
	joint Joint
	hertz float32
}

type recArgsMotorJointSetAngularDampingRatio struct {
	joint   Joint
	damping float32
}

type recArgsMotorJointSetMaxSpringForce struct {
	joint    Joint
	maxForce float32
}

type recArgsMotorJointSetMaxSpringTorque struct {
	joint     Joint
	maxTorque float32
}

type recArgsCreateMoverJoint struct {
	world World
	def   MoverJointDef
}

type recArgsMoverJointSetLinearVelocity struct {
	joint    Joint
	velocity Vec2
}

type recArgsMoverJointSetMaxVelocityForce struct {
	joint    Joint
	maxForce Vec2
}

type recArgsCreatePogoJoint struct {
	world World
	def   PogoJointDef
}

type recArgsPogoJointSetRestLength struct {
	joint  Joint
	length float32
}

type recArgsPogoJointSetSpringHertz struct {
	joint Joint
	hertz float32
}

type recArgsPogoJointSetSpringDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsCreatePrismaticJoint struct {
	world World
	def   PrismaticJointDef
}

type recArgsPrismaticJointEnableSpring struct {
	joint        Joint
	enableSpring bool
}

type recArgsPrismaticJointSetSpringHertz struct {
	joint Joint
	hertz float32
}

type recArgsPrismaticJointSetSpringDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsPrismaticJointSetTargetTranslation struct {
	joint       Joint
	translation float32
}

type recArgsPrismaticJointEnableLimit struct {
	joint       Joint
	enableLimit bool
}

type recArgsPrismaticJointSetLimits struct {
	joint Joint
	lower float32
	upper float32
}

type recArgsPrismaticJointEnableMotor struct {
	joint       Joint
	enableMotor bool
}

type recArgsPrismaticJointSetMotorSpeed struct {
	joint      Joint
	motorSpeed float32
}

type recArgsPrismaticJointSetMaxMotorForce struct {
	joint Joint
	force float32
}

type recArgsCreateRevoluteJoint struct {
	world World
	def   RevoluteJointDef
}

type recArgsRevoluteJointEnableSpring struct {
	joint        Joint
	enableSpring bool
}

type recArgsRevoluteJointSetSpringHertz struct {
	joint Joint
	hertz float32
}

type recArgsRevoluteJointSetSpringDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsRevoluteJointSetTargetAngle struct {
	joint Joint
	angle float32
}

type recArgsRevoluteJointEnableLimit struct {
	joint       Joint
	enableLimit bool
}

type recArgsRevoluteJointSetLimits struct {
	joint Joint
	lower float32
	upper float32
}

type recArgsRevoluteJointEnableMotor struct {
	joint       Joint
	enableMotor bool
}

type recArgsRevoluteJointSetMotorSpeed struct {
	joint      Joint
	motorSpeed float32
}

type recArgsRevoluteJointSetMaxMotorTorque struct {
	joint  Joint
	torque float32
}

type recArgsCreateWeldJoint struct {
	world World
	def   WeldJointDef
}

type recArgsWeldJointSetLinearHertz struct {
	joint Joint
	hertz float32
}

type recArgsWeldJointSetLinearDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsWeldJointSetAngularHertz struct {
	joint Joint
	hertz float32
}

type recArgsWeldJointSetAngularDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsCreateWheelJoint struct {
	world World
	def   WheelJointDef
}

type recArgsWheelJointEnableSpring struct {
	joint        Joint
	enableSpring bool
}

type recArgsWheelJointSetSpringHertz struct {
	joint Joint
	hertz float32
}

type recArgsWheelJointSetSpringDampingRatio struct {
	joint        Joint
	dampingRatio float32
}

type recArgsWheelJointEnableLimit struct {
	joint       Joint
	enableLimit bool
}

type recArgsWheelJointSetLimits struct {
	joint Joint
	lower float32
	upper float32
}

type recArgsWheelJointEnableMotor struct {
	joint       Joint
	enableMotor bool
}

type recArgsWheelJointSetMotorSpeed struct {
	joint      Joint
	motorSpeed float32
}

type recArgsWheelJointSetMaxMotorTorque struct {
	joint  Joint
	torque float32
}

type recArgsQueryOverlapAABB struct {
	world  World
	origin Pos
	aabb   AABB
	filter QueryFilter
}

type recArgsQueryOverlapShape struct {
	world  World
	origin Pos
	proxy  ShapeProxy
	filter QueryFilter
}

type recArgsQueryCastRay struct {
	world       World
	origin      Pos
	translation Vec2
	filter      QueryFilter
}

type recArgsQueryCastShape struct {
	world       World
	origin      Pos
	proxy       ShapeProxy
	translation Vec2
	filter      QueryFilter
}

type recArgsQueryCollideMover struct {
	world  World
	origin Pos
	mover  Capsule
	filter QueryFilter
}

type recArgsQueryCastRayClosest struct {
	world       World
	origin      Pos
	translation Vec2
	filter      QueryFilter
}

type recArgsQueryCastMover struct {
	world       World
	origin      Pos
	mover       Capsule
	translation Vec2
	filter      QueryFilter
}

type recArgsShapeTestPoint struct {
	shape Shape
	point Pos
}

type recArgsShapeRayCast struct {
	shape       Shape
	origin      Pos
	translation Vec2
}

type recArgsStateHash struct {
	world World
	hash  uint64
}

type recArgsRecordingBounds struct {
	bounds AABB
}

func recWriteArgsDestroyWorld(rec *Recording, a *recArgsDestroyWorld) {
	todo()
}

func recWriteDestroyWorld(rec *Recording, a *recArgsDestroyWorld) {
	todo()
}

func recWriteArgsStep(rec *Recording, a *recArgsStep) {
	todo()
}

func recWriteStep(rec *Recording, a *recArgsStep) {
	todo()
}

func recWriteArgsWorldEnableSleeping(rec *Recording, a *recArgsWorldEnableSleeping) {
	todo()
}

func recWriteWorldEnableSleeping(rec *Recording, a *recArgsWorldEnableSleeping) {
	todo()
}

func recWriteArgsWorldEnableContinuous(rec *Recording, a *recArgsWorldEnableContinuous) {
	todo()
}

func recWriteWorldEnableContinuous(rec *Recording, a *recArgsWorldEnableContinuous) {
	todo()
}

func recWriteArgsWorldSetRestitutionThreshold(rec *Recording, a *recArgsWorldSetRestitutionThreshold) {
	todo()
}

func recWriteWorldSetRestitutionThreshold(rec *Recording, a *recArgsWorldSetRestitutionThreshold) {
	todo()
}

func recWriteArgsWorldSetHitEventThreshold(rec *Recording, a *recArgsWorldSetHitEventThreshold) {
	todo()
}

func recWriteWorldSetHitEventThreshold(rec *Recording, a *recArgsWorldSetHitEventThreshold) {
	todo()
}

func recWriteArgsWorldSetGravity(rec *Recording, a *recArgsWorldSetGravity) {
	todo()
}

func recWriteWorldSetGravity(rec *Recording, a *recArgsWorldSetGravity) {
	todo()
}

func recWriteArgsWorldExplode(rec *Recording, a *recArgsWorldExplode) {
	todo()
}

func recWriteWorldExplode(rec *Recording, a *recArgsWorldExplode) {
	todo()
}

func recWriteArgsWorldSetContactTuning(rec *Recording, a *recArgsWorldSetContactTuning) {
	todo()
}

func recWriteWorldSetContactTuning(rec *Recording, a *recArgsWorldSetContactTuning) {
	todo()
}

func recWriteArgsWorldSetContactRecycleDistance(rec *Recording, a *recArgsWorldSetContactRecycleDistance) {
	todo()
}

func recWriteWorldSetContactRecycleDistance(rec *Recording, a *recArgsWorldSetContactRecycleDistance) {
	todo()
}

func recWriteArgsWorldSetMaximumLinearSpeed(rec *Recording, a *recArgsWorldSetMaximumLinearSpeed) {
	todo()
}

func recWriteWorldSetMaximumLinearSpeed(rec *Recording, a *recArgsWorldSetMaximumLinearSpeed) {
	todo()
}

func recWriteArgsWorldEnableWarmStarting(rec *Recording, a *recArgsWorldEnableWarmStarting) {
	todo()
}

func recWriteWorldEnableWarmStarting(rec *Recording, a *recArgsWorldEnableWarmStarting) {
	todo()
}

func recWriteArgsWorldRebuildStaticTree(rec *Recording, a *recArgsWorldRebuildStaticTree) {
	todo()
}

func recWriteWorldRebuildStaticTree(rec *Recording, a *recArgsWorldRebuildStaticTree) {
	todo()
}

func recWriteArgsWorldSetRestitutionIterations(rec *Recording, a *recArgsWorldSetRestitutionIterations) {
	todo()
}

func recWriteWorldSetRestitutionIterations(rec *Recording, a *recArgsWorldSetRestitutionIterations) {
	todo()
}

func recWriteArgsWorldEnableRestitutionPropagation(rec *Recording, a *recArgsWorldEnableRestitutionPropagation) {
	todo()
}

func recWriteWorldEnableRestitutionPropagation(rec *Recording, a *recArgsWorldEnableRestitutionPropagation) {
	todo()
}

func recWriteArgsCreateBody(rec *Recording, a *recArgsCreateBody) {
	todo()
}

func recWriteCreateBody(rec *Recording, a *recArgsCreateBody) {
	todo()
}

func recWriteRetCreateBody(rec *Recording, a *recArgsCreateBody, id Body) {
	todo()
}

func recWriteArgsDestroyBody(rec *Recording, a *recArgsDestroyBody) {
	todo()
}

func recWriteDestroyBody(rec *Recording, a *recArgsDestroyBody) {
	todo()
}

func recWriteArgsBodySetTransform(rec *Recording, a *recArgsBodySetTransform) {
	todo()
}

func recWriteBodySetTransform(rec *Recording, a *recArgsBodySetTransform) {
	todo()
}

func recWriteArgsBodySetLinearVelocity(rec *Recording, a *recArgsBodySetLinearVelocity) {
	todo()
}

func recWriteBodySetLinearVelocity(rec *Recording, a *recArgsBodySetLinearVelocity) {
	todo()
}

func recWriteArgsBodySetType(rec *Recording, a *recArgsBodySetType) {
	todo()
}

func recWriteBodySetType(rec *Recording, a *recArgsBodySetType) {
	todo()
}

func recWriteArgsBodySetName(rec *Recording, a *recArgsBodySetName) {
	todo()
}

func recWriteBodySetName(rec *Recording, a *recArgsBodySetName) {
	todo()
}

func recWriteArgsBodySetAngularVelocity(rec *Recording, a *recArgsBodySetAngularVelocity) {
	todo()
}

func recWriteBodySetAngularVelocity(rec *Recording, a *recArgsBodySetAngularVelocity) {
	todo()
}

func recWriteArgsBodySetTargetTransform(rec *Recording, a *recArgsBodySetTargetTransform) {
	todo()
}

func recWriteBodySetTargetTransform(rec *Recording, a *recArgsBodySetTargetTransform) {
	todo()
}

func recWriteArgsBodyApplyForce(rec *Recording, a *recArgsBodyApplyForce) {
	todo()
}

func recWriteBodyApplyForce(rec *Recording, a *recArgsBodyApplyForce) {
	todo()
}

func recWriteArgsBodyApplyForceToCenter(rec *Recording, a *recArgsBodyApplyForceToCenter) {
	todo()
}

func recWriteBodyApplyForceToCenter(rec *Recording, a *recArgsBodyApplyForceToCenter) {
	todo()
}

func recWriteArgsBodyApplyTorque(rec *Recording, a *recArgsBodyApplyTorque) {
	todo()
}

func recWriteBodyApplyTorque(rec *Recording, a *recArgsBodyApplyTorque) {
	todo()
}

func recWriteArgsBodyClearForces(rec *Recording, a *recArgsBodyClearForces) {
	todo()
}

func recWriteBodyClearForces(rec *Recording, a *recArgsBodyClearForces) {
	todo()
}

func recWriteArgsBodyApplyLinearImpulse(rec *Recording, a *recArgsBodyApplyLinearImpulse) {
	todo()
}

func recWriteBodyApplyLinearImpulse(rec *Recording, a *recArgsBodyApplyLinearImpulse) {
	todo()
}

func recWriteArgsBodyApplyLinearImpulseToCenter(rec *Recording, a *recArgsBodyApplyLinearImpulseToCenter) {
	todo()
}

func recWriteBodyApplyLinearImpulseToCenter(rec *Recording, a *recArgsBodyApplyLinearImpulseToCenter) {
	todo()
}

func recWriteArgsBodyApplyAngularImpulse(rec *Recording, a *recArgsBodyApplyAngularImpulse) {
	todo()
}

func recWriteBodyApplyAngularImpulse(rec *Recording, a *recArgsBodyApplyAngularImpulse) {
	todo()
}

func recWriteArgsBodySetMassData(rec *Recording, a *recArgsBodySetMassData) {
	todo()
}

func recWriteBodySetMassData(rec *Recording, a *recArgsBodySetMassData) {
	todo()
}

func recWriteArgsBodyUpdateMassFromShapes(rec *Recording, a *recArgsBodyUpdateMassFromShapes) {
	todo()
}

func recWriteBodyUpdateMassFromShapes(rec *Recording, a *recArgsBodyUpdateMassFromShapes) {
	todo()
}

func recWriteArgsBodySetLinearDamping(rec *Recording, a *recArgsBodySetLinearDamping) {
	todo()
}

func recWriteBodySetLinearDamping(rec *Recording, a *recArgsBodySetLinearDamping) {
	todo()
}

func recWriteArgsBodySetAngularDamping(rec *Recording, a *recArgsBodySetAngularDamping) {
	todo()
}

func recWriteBodySetAngularDamping(rec *Recording, a *recArgsBodySetAngularDamping) {
	todo()
}

func recWriteArgsBodySetGravityScale(rec *Recording, a *recArgsBodySetGravityScale) {
	todo()
}

func recWriteBodySetGravityScale(rec *Recording, a *recArgsBodySetGravityScale) {
	todo()
}

func recWriteArgsBodySetAwake(rec *Recording, a *recArgsBodySetAwake) {
	todo()
}

func recWriteBodySetAwake(rec *Recording, a *recArgsBodySetAwake) {
	todo()
}

func recWriteArgsBodyWakeTouching(rec *Recording, a *recArgsBodyWakeTouching) {
	todo()
}

func recWriteBodyWakeTouching(rec *Recording, a *recArgsBodyWakeTouching) {
	todo()
}

func recWriteArgsBodyEnableSleep(rec *Recording, a *recArgsBodyEnableSleep) {
	todo()
}

func recWriteBodyEnableSleep(rec *Recording, a *recArgsBodyEnableSleep) {
	todo()
}

func recWriteArgsBodySetSleepThreshold(rec *Recording, a *recArgsBodySetSleepThreshold) {
	todo()
}

func recWriteBodySetSleepThreshold(rec *Recording, a *recArgsBodySetSleepThreshold) {
	todo()
}

func recWriteArgsBodyDisable(rec *Recording, a *recArgsBodyDisable) {
	todo()
}

func recWriteBodyDisable(rec *Recording, a *recArgsBodyDisable) {
	todo()
}

func recWriteArgsBodyEnable(rec *Recording, a *recArgsBodyEnable) {
	todo()
}

func recWriteBodyEnable(rec *Recording, a *recArgsBodyEnable) {
	todo()
}

func recWriteArgsBodySetMotionLocks(rec *Recording, a *recArgsBodySetMotionLocks) {
	todo()
}

func recWriteBodySetMotionLocks(rec *Recording, a *recArgsBodySetMotionLocks) {
	todo()
}

func recWriteArgsBodySetBullet(rec *Recording, a *recArgsBodySetBullet) {
	todo()
}

func recWriteBodySetBullet(rec *Recording, a *recArgsBodySetBullet) {
	todo()
}

func recWriteArgsBodyEnableContactRecycling(rec *Recording, a *recArgsBodyEnableContactRecycling) {
	todo()
}

func recWriteBodyEnableContactRecycling(rec *Recording, a *recArgsBodyEnableContactRecycling) {
	todo()
}

func recWriteArgsBodyEnableContactEvents(rec *Recording, a *recArgsBodyEnableContactEvents) {
	todo()
}

func recWriteBodyEnableContactEvents(rec *Recording, a *recArgsBodyEnableContactEvents) {
	todo()
}

func recWriteArgsBodyEnableHitEvents(rec *Recording, a *recArgsBodyEnableHitEvents) {
	todo()
}

func recWriteBodyEnableHitEvents(rec *Recording, a *recArgsBodyEnableHitEvents) {
	todo()
}

func recWriteArgsBodySetSafetyFactor(rec *Recording, a *recArgsBodySetSafetyFactor) {
	todo()
}

func recWriteBodySetSafetyFactor(rec *Recording, a *recArgsBodySetSafetyFactor) {
	todo()
}

func recWriteArgsCreateCircleShape(rec *Recording, a *recArgsCreateCircleShape) {
	todo()
}

func recWriteCreateCircleShape(rec *Recording, a *recArgsCreateCircleShape) {
	todo()
}

func recWriteRetCreateCircleShape(rec *Recording, a *recArgsCreateCircleShape, id Shape) {
	todo()
}

func recWriteArgsCreateCapsuleShape(rec *Recording, a *recArgsCreateCapsuleShape) {
	todo()
}

func recWriteCreateCapsuleShape(rec *Recording, a *recArgsCreateCapsuleShape) {
	todo()
}

func recWriteRetCreateCapsuleShape(rec *Recording, a *recArgsCreateCapsuleShape, id Shape) {
	todo()
}

func recWriteArgsCreateSegmentShape(rec *Recording, a *recArgsCreateSegmentShape) {
	todo()
}

func recWriteCreateSegmentShape(rec *Recording, a *recArgsCreateSegmentShape) {
	todo()
}

func recWriteRetCreateSegmentShape(rec *Recording, a *recArgsCreateSegmentShape, id Shape) {
	todo()
}

func recWriteArgsCreatePolygonShape(rec *Recording, a *recArgsCreatePolygonShape) {
	todo()
}

func recWriteCreatePolygonShape(rec *Recording, a *recArgsCreatePolygonShape) {
	todo()
}

func recWriteRetCreatePolygonShape(rec *Recording, a *recArgsCreatePolygonShape, id Shape) {
	todo()
}

func recWriteArgsCreateChainSegmentShape(rec *Recording, a *recArgsCreateChainSegmentShape) {
	todo()
}

func recWriteCreateChainSegmentShape(rec *Recording, a *recArgsCreateChainSegmentShape) {
	todo()
}

func recWriteRetCreateChainSegmentShape(rec *Recording, a *recArgsCreateChainSegmentShape, id Shape) {
	todo()
}

func recWriteArgsDestroyShape(rec *Recording, a *recArgsDestroyShape) {
	todo()
}

func recWriteDestroyShape(rec *Recording, a *recArgsDestroyShape) {
	todo()
}

func recWriteArgsShapeSetDensity(rec *Recording, a *recArgsShapeSetDensity) {
	todo()
}

func recWriteShapeSetDensity(rec *Recording, a *recArgsShapeSetDensity) {
	todo()
}

func recWriteArgsShapeSetFriction(rec *Recording, a *recArgsShapeSetFriction) {
	todo()
}

func recWriteShapeSetFriction(rec *Recording, a *recArgsShapeSetFriction) {
	todo()
}

func recWriteArgsShapeSetRestitution(rec *Recording, a *recArgsShapeSetRestitution) {
	todo()
}

func recWriteShapeSetRestitution(rec *Recording, a *recArgsShapeSetRestitution) {
	todo()
}

func recWriteArgsShapeSetUserMaterial(rec *Recording, a *recArgsShapeSetUserMaterial) {
	todo()
}

func recWriteShapeSetUserMaterial(rec *Recording, a *recArgsShapeSetUserMaterial) {
	todo()
}

func recWriteArgsShapeSetSurfaceMaterial(rec *Recording, a *recArgsShapeSetSurfaceMaterial) {
	todo()
}

func recWriteShapeSetSurfaceMaterial(rec *Recording, a *recArgsShapeSetSurfaceMaterial) {
	todo()
}

func recWriteArgsShapeSetFilter(rec *Recording, a *recArgsShapeSetFilter) {
	todo()
}

func recWriteShapeSetFilter(rec *Recording, a *recArgsShapeSetFilter) {
	todo()
}

func recWriteArgsShapeEnableSensorEvents(rec *Recording, a *recArgsShapeEnableSensorEvents) {
	todo()
}

func recWriteShapeEnableSensorEvents(rec *Recording, a *recArgsShapeEnableSensorEvents) {
	todo()
}

func recWriteArgsShapeEnableContactEvents(rec *Recording, a *recArgsShapeEnableContactEvents) {
	todo()
}

func recWriteShapeEnableContactEvents(rec *Recording, a *recArgsShapeEnableContactEvents) {
	todo()
}

func recWriteArgsShapeEnablePreSolveEvents(rec *Recording, a *recArgsShapeEnablePreSolveEvents) {
	todo()
}

func recWriteShapeEnablePreSolveEvents(rec *Recording, a *recArgsShapeEnablePreSolveEvents) {
	todo()
}

func recWriteArgsShapeEnableHitEvents(rec *Recording, a *recArgsShapeEnableHitEvents) {
	todo()
}

func recWriteShapeEnableHitEvents(rec *Recording, a *recArgsShapeEnableHitEvents) {
	todo()
}

func recWriteArgsShapeSetCircle(rec *Recording, a *recArgsShapeSetCircle) {
	todo()
}

func recWriteShapeSetCircle(rec *Recording, a *recArgsShapeSetCircle) {
	todo()
}

func recWriteArgsShapeSetCapsule(rec *Recording, a *recArgsShapeSetCapsule) {
	todo()
}

func recWriteShapeSetCapsule(rec *Recording, a *recArgsShapeSetCapsule) {
	todo()
}

func recWriteArgsShapeSetSegment(rec *Recording, a *recArgsShapeSetSegment) {
	todo()
}

func recWriteShapeSetSegment(rec *Recording, a *recArgsShapeSetSegment) {
	todo()
}

func recWriteArgsShapeSetPolygon(rec *Recording, a *recArgsShapeSetPolygon) {
	todo()
}

func recWriteShapeSetPolygon(rec *Recording, a *recArgsShapeSetPolygon) {
	todo()
}

func recWriteArgsShapeSetChainSegment(rec *Recording, a *recArgsShapeSetChainSegment) {
	todo()
}

func recWriteShapeSetChainSegment(rec *Recording, a *recArgsShapeSetChainSegment) {
	todo()
}

func recWriteArgsShapeApplyWind(rec *Recording, a *recArgsShapeApplyWind) {
	todo()
}

func recWriteShapeApplyWind(rec *Recording, a *recArgsShapeApplyWind) {
	todo()
}

func recWriteArgsCreateChain(rec *Recording, a *recArgsCreateChain) {
	todo()
}

func recWriteCreateChain(rec *Recording, a *recArgsCreateChain) {
	todo()
}

func recWriteRetCreateChain(rec *Recording, a *recArgsCreateChain, id Chain) {
	todo()
}

func recWriteArgsDestroyChain(rec *Recording, a *recArgsDestroyChain) {
	todo()
}

func recWriteDestroyChain(rec *Recording, a *recArgsDestroyChain) {
	todo()
}

func recWriteArgsChainSetSurfaceMaterial(rec *Recording, a *recArgsChainSetSurfaceMaterial) {
	todo()
}

func recWriteChainSetSurfaceMaterial(rec *Recording, a *recArgsChainSetSurfaceMaterial) {
	todo()
}

func recWriteArgsChainSetAllSurfaceMaterials(rec *Recording, a *recArgsChainSetAllSurfaceMaterials) {
	todo()
}

func recWriteChainSetAllSurfaceMaterials(rec *Recording, a *recArgsChainSetAllSurfaceMaterials) {
	todo()
}

func recWriteArgsDestroyJoint(rec *Recording, a *recArgsDestroyJoint) {
	todo()
}

func recWriteDestroyJoint(rec *Recording, a *recArgsDestroyJoint) {
	todo()
}

func recWriteArgsJointSetLocalFrameA(rec *Recording, a *recArgsJointSetLocalFrameA) {
	todo()
}

func recWriteJointSetLocalFrameA(rec *Recording, a *recArgsJointSetLocalFrameA) {
	todo()
}

func recWriteArgsJointSetLocalFrameB(rec *Recording, a *recArgsJointSetLocalFrameB) {
	todo()
}

func recWriteJointSetLocalFrameB(rec *Recording, a *recArgsJointSetLocalFrameB) {
	todo()
}

func recWriteArgsJointSetCollideConnected(rec *Recording, a *recArgsJointSetCollideConnected) {
	todo()
}

func recWriteJointSetCollideConnected(rec *Recording, a *recArgsJointSetCollideConnected) {
	todo()
}

func recWriteArgsJointWakeBodies(rec *Recording, a *recArgsJointWakeBodies) {
	todo()
}

func recWriteJointWakeBodies(rec *Recording, a *recArgsJointWakeBodies) {
	todo()
}

func recWriteArgsJointSetConstraintTuning(rec *Recording, a *recArgsJointSetConstraintTuning) {
	todo()
}

func recWriteJointSetConstraintTuning(rec *Recording, a *recArgsJointSetConstraintTuning) {
	todo()
}

func recWriteArgsJointSetForceThreshold(rec *Recording, a *recArgsJointSetForceThreshold) {
	todo()
}

func recWriteJointSetForceThreshold(rec *Recording, a *recArgsJointSetForceThreshold) {
	todo()
}

func recWriteArgsJointSetTorqueThreshold(rec *Recording, a *recArgsJointSetTorqueThreshold) {
	todo()
}

func recWriteJointSetTorqueThreshold(rec *Recording, a *recArgsJointSetTorqueThreshold) {
	todo()
}

func recWriteArgsCreateDistanceJoint(rec *Recording, a *recArgsCreateDistanceJoint) {
	todo()
}

func recWriteCreateDistanceJoint(rec *Recording, a *recArgsCreateDistanceJoint) {
	todo()
}

func recWriteRetCreateDistanceJoint(rec *Recording, a *recArgsCreateDistanceJoint, id Joint) {
	todo()
}

func recWriteArgsDistanceJointSetLength(rec *Recording, a *recArgsDistanceJointSetLength) {
	todo()
}

func recWriteDistanceJointSetLength(rec *Recording, a *recArgsDistanceJointSetLength) {
	todo()
}

func recWriteArgsDistanceJointEnableSpring(rec *Recording, a *recArgsDistanceJointEnableSpring) {
	todo()
}

func recWriteDistanceJointEnableSpring(rec *Recording, a *recArgsDistanceJointEnableSpring) {
	todo()
}

func recWriteArgsDistanceJointSetSpringForceRange(rec *Recording, a *recArgsDistanceJointSetSpringForceRange) {
	todo()
}

func recWriteDistanceJointSetSpringForceRange(rec *Recording, a *recArgsDistanceJointSetSpringForceRange) {
	todo()
}

func recWriteArgsDistanceJointSetSpringHertz(rec *Recording, a *recArgsDistanceJointSetSpringHertz) {
	todo()
}

func recWriteDistanceJointSetSpringHertz(rec *Recording, a *recArgsDistanceJointSetSpringHertz) {
	todo()
}

func recWriteArgsDistanceJointSetSpringDampingRatio(rec *Recording, a *recArgsDistanceJointSetSpringDampingRatio) {
	todo()
}

func recWriteDistanceJointSetSpringDampingRatio(rec *Recording, a *recArgsDistanceJointSetSpringDampingRatio) {
	todo()
}

func recWriteArgsDistanceJointEnableLimit(rec *Recording, a *recArgsDistanceJointEnableLimit) {
	todo()
}

func recWriteDistanceJointEnableLimit(rec *Recording, a *recArgsDistanceJointEnableLimit) {
	todo()
}

func recWriteArgsDistanceJointSetLengthRange(rec *Recording, a *recArgsDistanceJointSetLengthRange) {
	todo()
}

func recWriteDistanceJointSetLengthRange(rec *Recording, a *recArgsDistanceJointSetLengthRange) {
	todo()
}

func recWriteArgsDistanceJointEnableMotor(rec *Recording, a *recArgsDistanceJointEnableMotor) {
	todo()
}

func recWriteDistanceJointEnableMotor(rec *Recording, a *recArgsDistanceJointEnableMotor) {
	todo()
}

func recWriteArgsDistanceJointSetMotorSpeed(rec *Recording, a *recArgsDistanceJointSetMotorSpeed) {
	todo()
}

func recWriteDistanceJointSetMotorSpeed(rec *Recording, a *recArgsDistanceJointSetMotorSpeed) {
	todo()
}

func recWriteArgsDistanceJointSetMaxMotorForce(rec *Recording, a *recArgsDistanceJointSetMaxMotorForce) {
	todo()
}

func recWriteDistanceJointSetMaxMotorForce(rec *Recording, a *recArgsDistanceJointSetMaxMotorForce) {
	todo()
}

func recWriteArgsCreateFilterJoint(rec *Recording, a *recArgsCreateFilterJoint) {
	todo()
}

func recWriteCreateFilterJoint(rec *Recording, a *recArgsCreateFilterJoint) {
	todo()
}

func recWriteRetCreateFilterJoint(rec *Recording, a *recArgsCreateFilterJoint, id Joint) {
	todo()
}

func recWriteArgsCreateMotorJoint(rec *Recording, a *recArgsCreateMotorJoint) {
	todo()
}

func recWriteCreateMotorJoint(rec *Recording, a *recArgsCreateMotorJoint) {
	todo()
}

func recWriteRetCreateMotorJoint(rec *Recording, a *recArgsCreateMotorJoint, id Joint) {
	todo()
}

func recWriteArgsMotorJointSetLinearVelocity(rec *Recording, a *recArgsMotorJointSetLinearVelocity) {
	todo()
}

func recWriteMotorJointSetLinearVelocity(rec *Recording, a *recArgsMotorJointSetLinearVelocity) {
	todo()
}

func recWriteArgsMotorJointSetAngularVelocity(rec *Recording, a *recArgsMotorJointSetAngularVelocity) {
	todo()
}

func recWriteMotorJointSetAngularVelocity(rec *Recording, a *recArgsMotorJointSetAngularVelocity) {
	todo()
}

func recWriteArgsMotorJointSetMaxVelocityForce(rec *Recording, a *recArgsMotorJointSetMaxVelocityForce) {
	todo()
}

func recWriteMotorJointSetMaxVelocityForce(rec *Recording, a *recArgsMotorJointSetMaxVelocityForce) {
	todo()
}

func recWriteArgsMotorJointSetMaxVelocityTorque(rec *Recording, a *recArgsMotorJointSetMaxVelocityTorque) {
	todo()
}

func recWriteMotorJointSetMaxVelocityTorque(rec *Recording, a *recArgsMotorJointSetMaxVelocityTorque) {
	todo()
}

func recWriteArgsMotorJointSetLinearHertz(rec *Recording, a *recArgsMotorJointSetLinearHertz) {
	todo()
}

func recWriteMotorJointSetLinearHertz(rec *Recording, a *recArgsMotorJointSetLinearHertz) {
	todo()
}

func recWriteArgsMotorJointSetLinearDampingRatio(rec *Recording, a *recArgsMotorJointSetLinearDampingRatio) {
	todo()
}

func recWriteMotorJointSetLinearDampingRatio(rec *Recording, a *recArgsMotorJointSetLinearDampingRatio) {
	todo()
}

func recWriteArgsMotorJointSetAngularHertz(rec *Recording, a *recArgsMotorJointSetAngularHertz) {
	todo()
}

func recWriteMotorJointSetAngularHertz(rec *Recording, a *recArgsMotorJointSetAngularHertz) {
	todo()
}

func recWriteArgsMotorJointSetAngularDampingRatio(rec *Recording, a *recArgsMotorJointSetAngularDampingRatio) {
	todo()
}

func recWriteMotorJointSetAngularDampingRatio(rec *Recording, a *recArgsMotorJointSetAngularDampingRatio) {
	todo()
}

func recWriteArgsMotorJointSetMaxSpringForce(rec *Recording, a *recArgsMotorJointSetMaxSpringForce) {
	todo()
}

func recWriteMotorJointSetMaxSpringForce(rec *Recording, a *recArgsMotorJointSetMaxSpringForce) {
	todo()
}

func recWriteArgsMotorJointSetMaxSpringTorque(rec *Recording, a *recArgsMotorJointSetMaxSpringTorque) {
	todo()
}

func recWriteMotorJointSetMaxSpringTorque(rec *Recording, a *recArgsMotorJointSetMaxSpringTorque) {
	todo()
}

func recWriteArgsCreateMoverJoint(rec *Recording, a *recArgsCreateMoverJoint) {
	todo()
}

func recWriteCreateMoverJoint(rec *Recording, a *recArgsCreateMoverJoint) {
	todo()
}

func recWriteRetCreateMoverJoint(rec *Recording, a *recArgsCreateMoverJoint, id Joint) {
	todo()
}

func recWriteArgsMoverJointSetLinearVelocity(rec *Recording, a *recArgsMoverJointSetLinearVelocity) {
	todo()
}

func recWriteMoverJointSetLinearVelocity(rec *Recording, a *recArgsMoverJointSetLinearVelocity) {
	todo()
}

func recWriteArgsMoverJointSetMaxVelocityForce(rec *Recording, a *recArgsMoverJointSetMaxVelocityForce) {
	todo()
}

func recWriteMoverJointSetMaxVelocityForce(rec *Recording, a *recArgsMoverJointSetMaxVelocityForce) {
	todo()
}

func recWriteArgsCreatePogoJoint(rec *Recording, a *recArgsCreatePogoJoint) {
	todo()
}

func recWriteCreatePogoJoint(rec *Recording, a *recArgsCreatePogoJoint) {
	todo()
}

func recWriteRetCreatePogoJoint(rec *Recording, a *recArgsCreatePogoJoint, id Joint) {
	todo()
}

func recWriteArgsPogoJointSetRestLength(rec *Recording, a *recArgsPogoJointSetRestLength) {
	todo()
}

func recWritePogoJointSetRestLength(rec *Recording, a *recArgsPogoJointSetRestLength) {
	todo()
}

func recWriteArgsPogoJointSetSpringHertz(rec *Recording, a *recArgsPogoJointSetSpringHertz) {
	todo()
}

func recWritePogoJointSetSpringHertz(rec *Recording, a *recArgsPogoJointSetSpringHertz) {
	todo()
}

func recWriteArgsPogoJointSetSpringDampingRatio(rec *Recording, a *recArgsPogoJointSetSpringDampingRatio) {
	todo()
}

func recWritePogoJointSetSpringDampingRatio(rec *Recording, a *recArgsPogoJointSetSpringDampingRatio) {
	todo()
}

func recWriteArgsCreatePrismaticJoint(rec *Recording, a *recArgsCreatePrismaticJoint) {
	todo()
}

func recWriteCreatePrismaticJoint(rec *Recording, a *recArgsCreatePrismaticJoint) {
	todo()
}

func recWriteRetCreatePrismaticJoint(rec *Recording, a *recArgsCreatePrismaticJoint, id Joint) {
	todo()
}

func recWriteArgsPrismaticJointEnableSpring(rec *Recording, a *recArgsPrismaticJointEnableSpring) {
	todo()
}

func recWritePrismaticJointEnableSpring(rec *Recording, a *recArgsPrismaticJointEnableSpring) {
	todo()
}

func recWriteArgsPrismaticJointSetSpringHertz(rec *Recording, a *recArgsPrismaticJointSetSpringHertz) {
	todo()
}

func recWritePrismaticJointSetSpringHertz(rec *Recording, a *recArgsPrismaticJointSetSpringHertz) {
	todo()
}

func recWriteArgsPrismaticJointSetSpringDampingRatio(rec *Recording, a *recArgsPrismaticJointSetSpringDampingRatio) {
	todo()
}

func recWritePrismaticJointSetSpringDampingRatio(rec *Recording, a *recArgsPrismaticJointSetSpringDampingRatio) {
	todo()
}

func recWriteArgsPrismaticJointSetTargetTranslation(rec *Recording, a *recArgsPrismaticJointSetTargetTranslation) {
	todo()
}

func recWritePrismaticJointSetTargetTranslation(rec *Recording, a *recArgsPrismaticJointSetTargetTranslation) {
	todo()
}

func recWriteArgsPrismaticJointEnableLimit(rec *Recording, a *recArgsPrismaticJointEnableLimit) {
	todo()
}

func recWritePrismaticJointEnableLimit(rec *Recording, a *recArgsPrismaticJointEnableLimit) {
	todo()
}

func recWriteArgsPrismaticJointSetLimits(rec *Recording, a *recArgsPrismaticJointSetLimits) {
	todo()
}

func recWritePrismaticJointSetLimits(rec *Recording, a *recArgsPrismaticJointSetLimits) {
	todo()
}

func recWriteArgsPrismaticJointEnableMotor(rec *Recording, a *recArgsPrismaticJointEnableMotor) {
	todo()
}

func recWritePrismaticJointEnableMotor(rec *Recording, a *recArgsPrismaticJointEnableMotor) {
	todo()
}

func recWriteArgsPrismaticJointSetMotorSpeed(rec *Recording, a *recArgsPrismaticJointSetMotorSpeed) {
	todo()
}

func recWritePrismaticJointSetMotorSpeed(rec *Recording, a *recArgsPrismaticJointSetMotorSpeed) {
	todo()
}

func recWriteArgsPrismaticJointSetMaxMotorForce(rec *Recording, a *recArgsPrismaticJointSetMaxMotorForce) {
	todo()
}

func recWritePrismaticJointSetMaxMotorForce(rec *Recording, a *recArgsPrismaticJointSetMaxMotorForce) {
	todo()
}

func recWriteArgsCreateRevoluteJoint(rec *Recording, a *recArgsCreateRevoluteJoint) {
	todo()
}

func recWriteCreateRevoluteJoint(rec *Recording, a *recArgsCreateRevoluteJoint) {
	todo()
}

func recWriteRetCreateRevoluteJoint(rec *Recording, a *recArgsCreateRevoluteJoint, id Joint) {
	todo()
}

func recWriteArgsRevoluteJointEnableSpring(rec *Recording, a *recArgsRevoluteJointEnableSpring) {
	todo()
}

func recWriteRevoluteJointEnableSpring(rec *Recording, a *recArgsRevoluteJointEnableSpring) {
	todo()
}

func recWriteArgsRevoluteJointSetSpringHertz(rec *Recording, a *recArgsRevoluteJointSetSpringHertz) {
	todo()
}

func recWriteRevoluteJointSetSpringHertz(rec *Recording, a *recArgsRevoluteJointSetSpringHertz) {
	todo()
}

func recWriteArgsRevoluteJointSetSpringDampingRatio(rec *Recording, a *recArgsRevoluteJointSetSpringDampingRatio) {
	todo()
}

func recWriteRevoluteJointSetSpringDampingRatio(rec *Recording, a *recArgsRevoluteJointSetSpringDampingRatio) {
	todo()
}

func recWriteArgsRevoluteJointSetTargetAngle(rec *Recording, a *recArgsRevoluteJointSetTargetAngle) {
	todo()
}

func recWriteRevoluteJointSetTargetAngle(rec *Recording, a *recArgsRevoluteJointSetTargetAngle) {
	todo()
}

func recWriteArgsRevoluteJointEnableLimit(rec *Recording, a *recArgsRevoluteJointEnableLimit) {
	todo()
}

func recWriteRevoluteJointEnableLimit(rec *Recording, a *recArgsRevoluteJointEnableLimit) {
	todo()
}

func recWriteArgsRevoluteJointSetLimits(rec *Recording, a *recArgsRevoluteJointSetLimits) {
	todo()
}

func recWriteRevoluteJointSetLimits(rec *Recording, a *recArgsRevoluteJointSetLimits) {
	todo()
}

func recWriteArgsRevoluteJointEnableMotor(rec *Recording, a *recArgsRevoluteJointEnableMotor) {
	todo()
}

func recWriteRevoluteJointEnableMotor(rec *Recording, a *recArgsRevoluteJointEnableMotor) {
	todo()
}

func recWriteArgsRevoluteJointSetMotorSpeed(rec *Recording, a *recArgsRevoluteJointSetMotorSpeed) {
	todo()
}

func recWriteRevoluteJointSetMotorSpeed(rec *Recording, a *recArgsRevoluteJointSetMotorSpeed) {
	todo()
}

func recWriteArgsRevoluteJointSetMaxMotorTorque(rec *Recording, a *recArgsRevoluteJointSetMaxMotorTorque) {
	todo()
}

func recWriteRevoluteJointSetMaxMotorTorque(rec *Recording, a *recArgsRevoluteJointSetMaxMotorTorque) {
	todo()
}

func recWriteArgsCreateWeldJoint(rec *Recording, a *recArgsCreateWeldJoint) {
	todo()
}

func recWriteCreateWeldJoint(rec *Recording, a *recArgsCreateWeldJoint) {
	todo()
}

func recWriteRetCreateWeldJoint(rec *Recording, a *recArgsCreateWeldJoint, id Joint) {
	todo()
}

func recWriteArgsWeldJointSetLinearHertz(rec *Recording, a *recArgsWeldJointSetLinearHertz) {
	todo()
}

func recWriteWeldJointSetLinearHertz(rec *Recording, a *recArgsWeldJointSetLinearHertz) {
	todo()
}

func recWriteArgsWeldJointSetLinearDampingRatio(rec *Recording, a *recArgsWeldJointSetLinearDampingRatio) {
	todo()
}

func recWriteWeldJointSetLinearDampingRatio(rec *Recording, a *recArgsWeldJointSetLinearDampingRatio) {
	todo()
}

func recWriteArgsWeldJointSetAngularHertz(rec *Recording, a *recArgsWeldJointSetAngularHertz) {
	todo()
}

func recWriteWeldJointSetAngularHertz(rec *Recording, a *recArgsWeldJointSetAngularHertz) {
	todo()
}

func recWriteArgsWeldJointSetAngularDampingRatio(rec *Recording, a *recArgsWeldJointSetAngularDampingRatio) {
	todo()
}

func recWriteWeldJointSetAngularDampingRatio(rec *Recording, a *recArgsWeldJointSetAngularDampingRatio) {
	todo()
}

func recWriteArgsCreateWheelJoint(rec *Recording, a *recArgsCreateWheelJoint) {
	todo()
}

func recWriteCreateWheelJoint(rec *Recording, a *recArgsCreateWheelJoint) {
	todo()
}

func recWriteRetCreateWheelJoint(rec *Recording, a *recArgsCreateWheelJoint, id Joint) {
	todo()
}

func recWriteArgsWheelJointEnableSpring(rec *Recording, a *recArgsWheelJointEnableSpring) {
	todo()
}

func recWriteWheelJointEnableSpring(rec *Recording, a *recArgsWheelJointEnableSpring) {
	todo()
}

func recWriteArgsWheelJointSetSpringHertz(rec *Recording, a *recArgsWheelJointSetSpringHertz) {
	todo()
}

func recWriteWheelJointSetSpringHertz(rec *Recording, a *recArgsWheelJointSetSpringHertz) {
	todo()
}

func recWriteArgsWheelJointSetSpringDampingRatio(rec *Recording, a *recArgsWheelJointSetSpringDampingRatio) {
	todo()
}

func recWriteWheelJointSetSpringDampingRatio(rec *Recording, a *recArgsWheelJointSetSpringDampingRatio) {
	todo()
}

func recWriteArgsWheelJointEnableLimit(rec *Recording, a *recArgsWheelJointEnableLimit) {
	todo()
}

func recWriteWheelJointEnableLimit(rec *Recording, a *recArgsWheelJointEnableLimit) {
	todo()
}

func recWriteArgsWheelJointSetLimits(rec *Recording, a *recArgsWheelJointSetLimits) {
	todo()
}

func recWriteWheelJointSetLimits(rec *Recording, a *recArgsWheelJointSetLimits) {
	todo()
}

func recWriteArgsWheelJointEnableMotor(rec *Recording, a *recArgsWheelJointEnableMotor) {
	todo()
}

func recWriteWheelJointEnableMotor(rec *Recording, a *recArgsWheelJointEnableMotor) {
	todo()
}

func recWriteArgsWheelJointSetMotorSpeed(rec *Recording, a *recArgsWheelJointSetMotorSpeed) {
	todo()
}

func recWriteWheelJointSetMotorSpeed(rec *Recording, a *recArgsWheelJointSetMotorSpeed) {
	todo()
}

func recWriteArgsWheelJointSetMaxMotorTorque(rec *Recording, a *recArgsWheelJointSetMaxMotorTorque) {
	todo()
}

func recWriteWheelJointSetMaxMotorTorque(rec *Recording, a *recArgsWheelJointSetMaxMotorTorque) {
	todo()
}

func recWriteArgsQueryOverlapAABB(rec *Recording, a *recArgsQueryOverlapAABB) {
	todo()
}

func recWriteQueryOverlapAABB(rec *Recording, a *recArgsQueryOverlapAABB) {
	todo()
}

func recWriteArgsQueryOverlapShape(rec *Recording, a *recArgsQueryOverlapShape) {
	todo()
}

func recWriteQueryOverlapShape(rec *Recording, a *recArgsQueryOverlapShape) {
	todo()
}

func recWriteArgsQueryCastRay(rec *Recording, a *recArgsQueryCastRay) {
	todo()
}

func recWriteQueryCastRay(rec *Recording, a *recArgsQueryCastRay) {
	todo()
}

func recWriteArgsQueryCastShape(rec *Recording, a *recArgsQueryCastShape) {
	todo()
}

func recWriteQueryCastShape(rec *Recording, a *recArgsQueryCastShape) {
	todo()
}

func recWriteArgsQueryCollideMover(rec *Recording, a *recArgsQueryCollideMover) {
	todo()
}

func recWriteQueryCollideMover(rec *Recording, a *recArgsQueryCollideMover) {
	todo()
}

func recWriteArgsQueryCastRayClosest(rec *Recording, a *recArgsQueryCastRayClosest) {
	todo()
}

func recWriteQueryCastRayClosest(rec *Recording, a *recArgsQueryCastRayClosest) {
	todo()
}

func recWriteArgsQueryCastMover(rec *Recording, a *recArgsQueryCastMover) {
	todo()
}

func recWriteQueryCastMover(rec *Recording, a *recArgsQueryCastMover) {
	todo()
}

func recWriteArgsShapeTestPoint(rec *Recording, a *recArgsShapeTestPoint) {
	todo()
}

func recWriteShapeTestPoint(rec *Recording, a *recArgsShapeTestPoint) {
	todo()
}

func recWriteArgsShapeRayCast(rec *Recording, a *recArgsShapeRayCast) {
	todo()
}

func recWriteShapeRayCast(rec *Recording, a *recArgsShapeRayCast) {
	todo()
}

func recWriteArgsStateHash(rec *Recording, a *recArgsStateHash) {
	todo()
}

func recWriteStateHash(rec *Recording, a *recArgsStateHash) {
	todo()
}

func recWriteArgsRecordingBounds(rec *Recording, a *recArgsRecordingBounds) {
	todo()
}

func recWriteRecordingBounds(rec *Recording, a *recArgsRecordingBounds) {
	todo()
}
