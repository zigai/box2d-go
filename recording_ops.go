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

type recArgsDestroyWorld struct{}

type recArgsStep struct {
	dt           float32
	subStepCount int32
}

type recArgsWorldEnableSleeping struct {
	flag bool
}

type recArgsWorldEnableContinuous struct {
	flag bool
}

type recArgsWorldSetRestitutionThreshold struct {
	value float32
}

type recArgsWorldSetHitEventThreshold struct {
	value float32
}

type recArgsWorldSetGravity struct {
	gravity Vec2
}

type recArgsWorldExplode struct {
	def ExplosionDef
}

type recArgsWorldSetContactTuning struct {
	hertz        float32
	dampingRatio float32
	pushSpeed    float32
}

type recArgsWorldSetContactRecycleDistance struct {
	recycleDistance float32
}

type recArgsWorldSetMaximumLinearSpeed struct {
	maximumLinearSpeed float32
}

type recArgsWorldEnableWarmStarting struct {
	flag bool
}

type recArgsWorldRebuildStaticTree struct{}

type recArgsWorldSetRestitutionIterations struct {
	iterations int32
}

type recArgsWorldEnableRestitutionPropagation struct {
	flag bool
}

type recArgsCreateBody struct {
	def BodyDef
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
	def DistanceJointDef
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
	def FilterJointDef
}

type recArgsCreateMotorJoint struct {
	def MotorJointDef
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
	def MoverJointDef
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
	def PogoJointDef
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
	def PrismaticJointDef
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
	def RevoluteJointDef
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
	def WeldJointDef
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
	def WheelJointDef
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
	origin Pos
	aabb   AABB
	filter QueryFilter
}

type recArgsQueryOverlapShape struct {
	origin Pos
	proxy  ShapeProxy
	filter QueryFilter
}

type recArgsQueryCastRay struct {
	origin      Pos
	translation Vec2
	filter      QueryFilter
}

type recArgsQueryCastShape struct {
	origin      Pos
	proxy       ShapeProxy
	translation Vec2
	filter      QueryFilter
}

type recArgsQueryCollideMover struct {
	origin Pos
	mover  Capsule
	filter QueryFilter
}

type recArgsQueryCastRayClosest struct {
	origin      Pos
	translation Vec2
	filter      QueryFilter
}

type recArgsQueryCastMover struct {
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
	hash uint64
}

type recArgsRecordingBounds struct {
	bounds AABB
}

func recWriteArgsDestroyWorld(rec *Recording, a *recArgsDestroyWorld) {
}

func recWriteDestroyWorld(rec *Recording, a *recArgsDestroyWorld) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x01)
	recWriteArgsDestroyWorld(rec, a)
	recEndRecord(rec)
}

func recWriteArgsStep(rec *Recording, a *recArgsStep) {
	recWF32(&rec.buffer, a.dt)
	recWI32(&rec.buffer, int32(a.subStepCount))
}

func recWriteStep(rec *Recording, a *recArgsStep) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x80)
	recWriteArgsStep(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldEnableSleeping(rec *Recording, a *recArgsWorldEnableSleeping) {
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteWorldEnableSleeping(rec *Recording, a *recArgsWorldEnableSleeping) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x02)
	recWriteArgsWorldEnableSleeping(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldEnableContinuous(rec *Recording, a *recArgsWorldEnableContinuous) {
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteWorldEnableContinuous(rec *Recording, a *recArgsWorldEnableContinuous) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x03)
	recWriteArgsWorldEnableContinuous(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetRestitutionThreshold(rec *Recording, a *recArgsWorldSetRestitutionThreshold) {
	recWF32(&rec.buffer, a.value)
}

func recWriteWorldSetRestitutionThreshold(rec *Recording, a *recArgsWorldSetRestitutionThreshold) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x04)
	recWriteArgsWorldSetRestitutionThreshold(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetHitEventThreshold(rec *Recording, a *recArgsWorldSetHitEventThreshold) {
	recWF32(&rec.buffer, a.value)
}

func recWriteWorldSetHitEventThreshold(rec *Recording, a *recArgsWorldSetHitEventThreshold) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x05)
	recWriteArgsWorldSetHitEventThreshold(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetGravity(rec *Recording, a *recArgsWorldSetGravity) {
	recWVEC2(&rec.buffer, a.gravity)
}

func recWriteWorldSetGravity(rec *Recording, a *recArgsWorldSetGravity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x06)
	recWriteArgsWorldSetGravity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldExplode(rec *Recording, a *recArgsWorldExplode) {
	recWEXPLOSIONDEF(&rec.buffer, a.def)
}

func recWriteWorldExplode(rec *Recording, a *recArgsWorldExplode) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x07)
	recWriteArgsWorldExplode(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetContactTuning(rec *Recording, a *recArgsWorldSetContactTuning) {
	recWF32(&rec.buffer, a.hertz)
	recWF32(&rec.buffer, a.dampingRatio)
	recWF32(&rec.buffer, a.pushSpeed)
}

func recWriteWorldSetContactTuning(rec *Recording, a *recArgsWorldSetContactTuning) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x08)
	recWriteArgsWorldSetContactTuning(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetContactRecycleDistance(rec *Recording, a *recArgsWorldSetContactRecycleDistance) {
	recWF32(&rec.buffer, a.recycleDistance)
}

func recWriteWorldSetContactRecycleDistance(rec *Recording, a *recArgsWorldSetContactRecycleDistance) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x09)
	recWriteArgsWorldSetContactRecycleDistance(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetMaximumLinearSpeed(rec *Recording, a *recArgsWorldSetMaximumLinearSpeed) {
	recWF32(&rec.buffer, a.maximumLinearSpeed)
}

func recWriteWorldSetMaximumLinearSpeed(rec *Recording, a *recArgsWorldSetMaximumLinearSpeed) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x0A)
	recWriteArgsWorldSetMaximumLinearSpeed(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldEnableWarmStarting(rec *Recording, a *recArgsWorldEnableWarmStarting) {
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteWorldEnableWarmStarting(rec *Recording, a *recArgsWorldEnableWarmStarting) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x0B)
	recWriteArgsWorldEnableWarmStarting(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldRebuildStaticTree(rec *Recording, a *recArgsWorldRebuildStaticTree) {
}

func recWriteWorldRebuildStaticTree(rec *Recording, a *recArgsWorldRebuildStaticTree) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x0C)
	recWriteArgsWorldRebuildStaticTree(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldSetRestitutionIterations(rec *Recording, a *recArgsWorldSetRestitutionIterations) {
	recWI32(&rec.buffer, int32(a.iterations))
}

func recWriteWorldSetRestitutionIterations(rec *Recording, a *recArgsWorldSetRestitutionIterations) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x0D)
	recWriteArgsWorldSetRestitutionIterations(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWorldEnableRestitutionPropagation(rec *Recording, a *recArgsWorldEnableRestitutionPropagation) {
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteWorldEnableRestitutionPropagation(rec *Recording, a *recArgsWorldEnableRestitutionPropagation) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x0E)
	recWriteArgsWorldEnableRestitutionPropagation(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateBody(rec *Recording, a *recArgsCreateBody) {
	recWBODYDEF(&rec.buffer, a.def)
}

func recWriteCreateBody(rec *Recording, a *recArgsCreateBody) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x10)
	recWriteArgsCreateBody(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateBody(rec *Recording, a *recArgsCreateBody, id Body) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x10)
	recWriteArgsCreateBody(rec, a)
	recWBODYID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsDestroyBody(rec *Recording, a *recArgsDestroyBody) {
	recWBODYID(&rec.buffer, a.body)
}

func recWriteDestroyBody(rec *Recording, a *recArgsDestroyBody) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x11)
	recWriteArgsDestroyBody(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetTransform(rec *Recording, a *recArgsBodySetTransform) {
	recWBODYID(&rec.buffer, a.body)
	recWPOSITION(&rec.buffer, a.position)
	recWROT(&rec.buffer, a.rotation)
}

func recWriteBodySetTransform(rec *Recording, a *recArgsBodySetTransform) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x20)
	recWriteArgsBodySetTransform(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetLinearVelocity(rec *Recording, a *recArgsBodySetLinearVelocity) {
	recWBODYID(&rec.buffer, a.body)
	recWVEC2(&rec.buffer, a.v)
}

func recWriteBodySetLinearVelocity(rec *Recording, a *recArgsBodySetLinearVelocity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x21)
	recWriteArgsBodySetLinearVelocity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetType(rec *Recording, a *recArgsBodySetType) {
	recWBODYID(&rec.buffer, a.body)
	recWI32(&rec.buffer, int32(a.bodyType))
}

func recWriteBodySetType(rec *Recording, a *recArgsBodySetType) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x22)
	recWriteArgsBodySetType(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetName(rec *Recording, a *recArgsBodySetName) {
	recWBODYID(&rec.buffer, a.body)
	recWSTR(&rec.buffer, a.name)
}

func recWriteBodySetName(rec *Recording, a *recArgsBodySetName) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x23)
	recWriteArgsBodySetName(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetAngularVelocity(rec *Recording, a *recArgsBodySetAngularVelocity) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.w)
}

func recWriteBodySetAngularVelocity(rec *Recording, a *recArgsBodySetAngularVelocity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x24)
	recWriteArgsBodySetAngularVelocity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetTargetTransform(rec *Recording, a *recArgsBodySetTargetTransform) {
	recWBODYID(&rec.buffer, a.body)
	recWWORLDXF(&rec.buffer, a.target)
	recWF32(&rec.buffer, a.timeStep)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodySetTargetTransform(rec *Recording, a *recArgsBodySetTargetTransform) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x25)
	recWriteArgsBodySetTargetTransform(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyApplyForce(rec *Recording, a *recArgsBodyApplyForce) {
	recWBODYID(&rec.buffer, a.body)
	recWVEC2(&rec.buffer, a.force)
	recWPOSITION(&rec.buffer, a.point)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodyApplyForce(rec *Recording, a *recArgsBodyApplyForce) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x26)
	recWriteArgsBodyApplyForce(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyApplyForceToCenter(rec *Recording, a *recArgsBodyApplyForceToCenter) {
	recWBODYID(&rec.buffer, a.body)
	recWVEC2(&rec.buffer, a.force)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodyApplyForceToCenter(rec *Recording, a *recArgsBodyApplyForceToCenter) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x27)
	recWriteArgsBodyApplyForceToCenter(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyApplyTorque(rec *Recording, a *recArgsBodyApplyTorque) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.torque)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodyApplyTorque(rec *Recording, a *recArgsBodyApplyTorque) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x28)
	recWriteArgsBodyApplyTorque(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyClearForces(rec *Recording, a *recArgsBodyClearForces) {
	recWBODYID(&rec.buffer, a.body)
}

func recWriteBodyClearForces(rec *Recording, a *recArgsBodyClearForces) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x29)
	recWriteArgsBodyClearForces(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyApplyLinearImpulse(rec *Recording, a *recArgsBodyApplyLinearImpulse) {
	recWBODYID(&rec.buffer, a.body)
	recWVEC2(&rec.buffer, a.impulse)
	recWPOSITION(&rec.buffer, a.point)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodyApplyLinearImpulse(rec *Recording, a *recArgsBodyApplyLinearImpulse) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x2A)
	recWriteArgsBodyApplyLinearImpulse(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyApplyLinearImpulseToCenter(rec *Recording, a *recArgsBodyApplyLinearImpulseToCenter) {
	recWBODYID(&rec.buffer, a.body)
	recWVEC2(&rec.buffer, a.impulse)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodyApplyLinearImpulseToCenter(rec *Recording, a *recArgsBodyApplyLinearImpulseToCenter) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x2B)
	recWriteArgsBodyApplyLinearImpulseToCenter(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyApplyAngularImpulse(rec *Recording, a *recArgsBodyApplyAngularImpulse) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.impulse)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteBodyApplyAngularImpulse(rec *Recording, a *recArgsBodyApplyAngularImpulse) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x2C)
	recWriteArgsBodyApplyAngularImpulse(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetMassData(rec *Recording, a *recArgsBodySetMassData) {
	recWBODYID(&rec.buffer, a.body)
	recWMASSDATA(&rec.buffer, a.massData)
}

func recWriteBodySetMassData(rec *Recording, a *recArgsBodySetMassData) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x2D)
	recWriteArgsBodySetMassData(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyUpdateMassFromShapes(rec *Recording, a *recArgsBodyUpdateMassFromShapes) {
	recWBODYID(&rec.buffer, a.body)
}

func recWriteBodyUpdateMassFromShapes(rec *Recording, a *recArgsBodyUpdateMassFromShapes) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x2E)
	recWriteArgsBodyUpdateMassFromShapes(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetLinearDamping(rec *Recording, a *recArgsBodySetLinearDamping) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.damping)
}

func recWriteBodySetLinearDamping(rec *Recording, a *recArgsBodySetLinearDamping) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x2F)
	recWriteArgsBodySetLinearDamping(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetAngularDamping(rec *Recording, a *recArgsBodySetAngularDamping) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.damping)
}

func recWriteBodySetAngularDamping(rec *Recording, a *recArgsBodySetAngularDamping) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x30)
	recWriteArgsBodySetAngularDamping(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetGravityScale(rec *Recording, a *recArgsBodySetGravityScale) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.scale)
}

func recWriteBodySetGravityScale(rec *Recording, a *recArgsBodySetGravityScale) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x31)
	recWriteArgsBodySetGravityScale(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetAwake(rec *Recording, a *recArgsBodySetAwake) {
	recWBODYID(&rec.buffer, a.body)
	recWBOOL(&rec.buffer, a.awake)
}

func recWriteBodySetAwake(rec *Recording, a *recArgsBodySetAwake) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x32)
	recWriteArgsBodySetAwake(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyWakeTouching(rec *Recording, a *recArgsBodyWakeTouching) {
	recWBODYID(&rec.buffer, a.body)
}

func recWriteBodyWakeTouching(rec *Recording, a *recArgsBodyWakeTouching) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x33)
	recWriteArgsBodyWakeTouching(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyEnableSleep(rec *Recording, a *recArgsBodyEnableSleep) {
	recWBODYID(&rec.buffer, a.body)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteBodyEnableSleep(rec *Recording, a *recArgsBodyEnableSleep) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x34)
	recWriteArgsBodyEnableSleep(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetSleepThreshold(rec *Recording, a *recArgsBodySetSleepThreshold) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.threshold)
}

func recWriteBodySetSleepThreshold(rec *Recording, a *recArgsBodySetSleepThreshold) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x35)
	recWriteArgsBodySetSleepThreshold(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyDisable(rec *Recording, a *recArgsBodyDisable) {
	recWBODYID(&rec.buffer, a.body)
}

func recWriteBodyDisable(rec *Recording, a *recArgsBodyDisable) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x36)
	recWriteArgsBodyDisable(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyEnable(rec *Recording, a *recArgsBodyEnable) {
	recWBODYID(&rec.buffer, a.body)
}

func recWriteBodyEnable(rec *Recording, a *recArgsBodyEnable) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x37)
	recWriteArgsBodyEnable(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetMotionLocks(rec *Recording, a *recArgsBodySetMotionLocks) {
	recWBODYID(&rec.buffer, a.body)
	recWLOCKS(&rec.buffer, a.locks)
}

func recWriteBodySetMotionLocks(rec *Recording, a *recArgsBodySetMotionLocks) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x38)
	recWriteArgsBodySetMotionLocks(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetBullet(rec *Recording, a *recArgsBodySetBullet) {
	recWBODYID(&rec.buffer, a.body)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteBodySetBullet(rec *Recording, a *recArgsBodySetBullet) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x39)
	recWriteArgsBodySetBullet(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyEnableContactRecycling(rec *Recording, a *recArgsBodyEnableContactRecycling) {
	recWBODYID(&rec.buffer, a.body)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteBodyEnableContactRecycling(rec *Recording, a *recArgsBodyEnableContactRecycling) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x3A)
	recWriteArgsBodyEnableContactRecycling(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyEnableContactEvents(rec *Recording, a *recArgsBodyEnableContactEvents) {
	recWBODYID(&rec.buffer, a.body)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteBodyEnableContactEvents(rec *Recording, a *recArgsBodyEnableContactEvents) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x3B)
	recWriteArgsBodyEnableContactEvents(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodyEnableHitEvents(rec *Recording, a *recArgsBodyEnableHitEvents) {
	recWBODYID(&rec.buffer, a.body)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteBodyEnableHitEvents(rec *Recording, a *recArgsBodyEnableHitEvents) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x3C)
	recWriteArgsBodyEnableHitEvents(rec, a)
	recEndRecord(rec)
}

func recWriteArgsBodySetSafetyFactor(rec *Recording, a *recArgsBodySetSafetyFactor) {
	recWBODYID(&rec.buffer, a.body)
	recWF32(&rec.buffer, a.value)
}

func recWriteBodySetSafetyFactor(rec *Recording, a *recArgsBodySetSafetyFactor) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x3D)
	recWriteArgsBodySetSafetyFactor(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateCircleShape(rec *Recording, a *recArgsCreateCircleShape) {
	recWBODYID(&rec.buffer, a.body)
	recWSHAPEDEF(&rec.buffer, a.def)
	recWCIRCLE(&rec.buffer, a.circle)
}

func recWriteCreateCircleShape(rec *Recording, a *recArgsCreateCircleShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x40)
	recWriteArgsCreateCircleShape(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateCircleShape(rec *Recording, a *recArgsCreateCircleShape, id Shape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x40)
	recWriteArgsCreateCircleShape(rec, a)
	recWSHAPEID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsCreateCapsuleShape(rec *Recording, a *recArgsCreateCapsuleShape) {
	recWBODYID(&rec.buffer, a.body)
	recWSHAPEDEF(&rec.buffer, a.def)
	recWCAPSULE(&rec.buffer, a.capsule)
}

func recWriteCreateCapsuleShape(rec *Recording, a *recArgsCreateCapsuleShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x41)
	recWriteArgsCreateCapsuleShape(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateCapsuleShape(rec *Recording, a *recArgsCreateCapsuleShape, id Shape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x41)
	recWriteArgsCreateCapsuleShape(rec, a)
	recWSHAPEID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsCreateSegmentShape(rec *Recording, a *recArgsCreateSegmentShape) {
	recWBODYID(&rec.buffer, a.body)
	recWSHAPEDEF(&rec.buffer, a.def)
	recWSEGMENT(&rec.buffer, a.segment)
}

func recWriteCreateSegmentShape(rec *Recording, a *recArgsCreateSegmentShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x42)
	recWriteArgsCreateSegmentShape(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateSegmentShape(rec *Recording, a *recArgsCreateSegmentShape, id Shape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x42)
	recWriteArgsCreateSegmentShape(rec, a)
	recWSHAPEID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsCreatePolygonShape(rec *Recording, a *recArgsCreatePolygonShape) {
	recWBODYID(&rec.buffer, a.body)
	recWSHAPEDEF(&rec.buffer, a.def)
	recWPOLYGON(&rec.buffer, a.polygon)
}

func recWriteCreatePolygonShape(rec *Recording, a *recArgsCreatePolygonShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x43)
	recWriteArgsCreatePolygonShape(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreatePolygonShape(rec *Recording, a *recArgsCreatePolygonShape, id Shape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x43)
	recWriteArgsCreatePolygonShape(rec, a)
	recWSHAPEID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsCreateChainSegmentShape(rec *Recording, a *recArgsCreateChainSegmentShape) {
	recWBODYID(&rec.buffer, a.body)
	recWSHAPEDEF(&rec.buffer, a.def)
	recWCHAINSEG(&rec.buffer, a.chainSegment)
}

func recWriteCreateChainSegmentShape(rec *Recording, a *recArgsCreateChainSegmentShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x44)
	recWriteArgsCreateChainSegmentShape(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateChainSegmentShape(rec *Recording, a *recArgsCreateChainSegmentShape, id Shape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x44)
	recWriteArgsCreateChainSegmentShape(rec, a)
	recWSHAPEID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsDestroyShape(rec *Recording, a *recArgsDestroyShape) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWBOOL(&rec.buffer, a.updateBodyMass)
}

func recWriteDestroyShape(rec *Recording, a *recArgsDestroyShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x45)
	recWriteArgsDestroyShape(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetDensity(rec *Recording, a *recArgsShapeSetDensity) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWF32(&rec.buffer, a.density)
	recWBOOL(&rec.buffer, a.updateBodyMass)
}

func recWriteShapeSetDensity(rec *Recording, a *recArgsShapeSetDensity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x50)
	recWriteArgsShapeSetDensity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetFriction(rec *Recording, a *recArgsShapeSetFriction) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWF32(&rec.buffer, a.friction)
}

func recWriteShapeSetFriction(rec *Recording, a *recArgsShapeSetFriction) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x51)
	recWriteArgsShapeSetFriction(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetRestitution(rec *Recording, a *recArgsShapeSetRestitution) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWF32(&rec.buffer, a.restitution)
}

func recWriteShapeSetRestitution(rec *Recording, a *recArgsShapeSetRestitution) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x52)
	recWriteArgsShapeSetRestitution(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetUserMaterial(rec *Recording, a *recArgsShapeSetUserMaterial) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWU64(&rec.buffer, a.material)
}

func recWriteShapeSetUserMaterial(rec *Recording, a *recArgsShapeSetUserMaterial) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x53)
	recWriteArgsShapeSetUserMaterial(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetSurfaceMaterial(rec *Recording, a *recArgsShapeSetSurfaceMaterial) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWMATERIAL(&rec.buffer, a.material)
}

func recWriteShapeSetSurfaceMaterial(rec *Recording, a *recArgsShapeSetSurfaceMaterial) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x54)
	recWriteArgsShapeSetSurfaceMaterial(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetFilter(rec *Recording, a *recArgsShapeSetFilter) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWFILTER(&rec.buffer, a.filter)
}

func recWriteShapeSetFilter(rec *Recording, a *recArgsShapeSetFilter) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x55)
	recWriteArgsShapeSetFilter(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeEnableSensorEvents(rec *Recording, a *recArgsShapeEnableSensorEvents) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteShapeEnableSensorEvents(rec *Recording, a *recArgsShapeEnableSensorEvents) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x56)
	recWriteArgsShapeEnableSensorEvents(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeEnableContactEvents(rec *Recording, a *recArgsShapeEnableContactEvents) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteShapeEnableContactEvents(rec *Recording, a *recArgsShapeEnableContactEvents) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x57)
	recWriteArgsShapeEnableContactEvents(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeEnablePreSolveEvents(rec *Recording, a *recArgsShapeEnablePreSolveEvents) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteShapeEnablePreSolveEvents(rec *Recording, a *recArgsShapeEnablePreSolveEvents) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x58)
	recWriteArgsShapeEnablePreSolveEvents(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeEnableHitEvents(rec *Recording, a *recArgsShapeEnableHitEvents) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWBOOL(&rec.buffer, a.flag)
}

func recWriteShapeEnableHitEvents(rec *Recording, a *recArgsShapeEnableHitEvents) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x59)
	recWriteArgsShapeEnableHitEvents(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetCircle(rec *Recording, a *recArgsShapeSetCircle) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWCIRCLE(&rec.buffer, a.circle)
}

func recWriteShapeSetCircle(rec *Recording, a *recArgsShapeSetCircle) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x5A)
	recWriteArgsShapeSetCircle(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetCapsule(rec *Recording, a *recArgsShapeSetCapsule) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWCAPSULE(&rec.buffer, a.capsule)
}

func recWriteShapeSetCapsule(rec *Recording, a *recArgsShapeSetCapsule) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x5B)
	recWriteArgsShapeSetCapsule(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetSegment(rec *Recording, a *recArgsShapeSetSegment) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWSEGMENT(&rec.buffer, a.segment)
}

func recWriteShapeSetSegment(rec *Recording, a *recArgsShapeSetSegment) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x5C)
	recWriteArgsShapeSetSegment(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetPolygon(rec *Recording, a *recArgsShapeSetPolygon) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWPOLYGON(&rec.buffer, a.polygon)
}

func recWriteShapeSetPolygon(rec *Recording, a *recArgsShapeSetPolygon) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x5D)
	recWriteArgsShapeSetPolygon(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeSetChainSegment(rec *Recording, a *recArgsShapeSetChainSegment) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWCHAINSEG(&rec.buffer, a.chainSegment)
}

func recWriteShapeSetChainSegment(rec *Recording, a *recArgsShapeSetChainSegment) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x5E)
	recWriteArgsShapeSetChainSegment(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeApplyWind(rec *Recording, a *recArgsShapeApplyWind) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWVEC2(&rec.buffer, a.wind)
	recWF32(&rec.buffer, a.drag)
	recWF32(&rec.buffer, a.lift)
	recWBOOL(&rec.buffer, a.wake)
}

func recWriteShapeApplyWind(rec *Recording, a *recArgsShapeApplyWind) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x5F)
	recWriteArgsShapeApplyWind(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateChain(rec *Recording, a *recArgsCreateChain) {
	recWBODYID(&rec.buffer, a.body)
	recWCHAINDEF(&rec.buffer, a.def)
}

func recWriteCreateChain(rec *Recording, a *recArgsCreateChain) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x70)
	recWriteArgsCreateChain(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateChain(rec *Recording, a *recArgsCreateChain, id Chain) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x70)
	recWriteArgsCreateChain(rec, a)
	recWCHAINID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsDestroyChain(rec *Recording, a *recArgsDestroyChain) {
	recWCHAINID(&rec.buffer, a.chain)
}

func recWriteDestroyChain(rec *Recording, a *recArgsDestroyChain) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x71)
	recWriteArgsDestroyChain(rec, a)
	recEndRecord(rec)
}

func recWriteArgsChainSetSurfaceMaterial(rec *Recording, a *recArgsChainSetSurfaceMaterial) {
	recWCHAINID(&rec.buffer, a.chain)
	recWMATERIAL(&rec.buffer, a.material)
	recWI32(&rec.buffer, int32(a.segmentIndex))
}

func recWriteChainSetSurfaceMaterial(rec *Recording, a *recArgsChainSetSurfaceMaterial) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x72)
	recWriteArgsChainSetSurfaceMaterial(rec, a)
	recEndRecord(rec)
}

func recWriteArgsChainSetAllSurfaceMaterials(rec *Recording, a *recArgsChainSetAllSurfaceMaterials) {
	recWCHAINID(&rec.buffer, a.chain)
	recWMATERIAL(&rec.buffer, a.material)
}

func recWriteChainSetAllSurfaceMaterials(rec *Recording, a *recArgsChainSetAllSurfaceMaterials) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x73)
	recWriteArgsChainSetAllSurfaceMaterials(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDestroyJoint(rec *Recording, a *recArgsDestroyJoint) {
	recWJOINTID(&rec.buffer, a.joint)
}

func recWriteDestroyJoint(rec *Recording, a *recArgsDestroyJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x90)
	recWriteArgsDestroyJoint(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointSetLocalFrameA(rec *Recording, a *recArgsJointSetLocalFrameA) {
	recWJOINTID(&rec.buffer, a.joint)
	recWXF(&rec.buffer, a.localFrame)
}

func recWriteJointSetLocalFrameA(rec *Recording, a *recArgsJointSetLocalFrameA) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x91)
	recWriteArgsJointSetLocalFrameA(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointSetLocalFrameB(rec *Recording, a *recArgsJointSetLocalFrameB) {
	recWJOINTID(&rec.buffer, a.joint)
	recWXF(&rec.buffer, a.localFrame)
}

func recWriteJointSetLocalFrameB(rec *Recording, a *recArgsJointSetLocalFrameB) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x92)
	recWriteArgsJointSetLocalFrameB(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointSetCollideConnected(rec *Recording, a *recArgsJointSetCollideConnected) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.shouldCollide)
}

func recWriteJointSetCollideConnected(rec *Recording, a *recArgsJointSetCollideConnected) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x93)
	recWriteArgsJointSetCollideConnected(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointWakeBodies(rec *Recording, a *recArgsJointWakeBodies) {
	recWJOINTID(&rec.buffer, a.joint)
}

func recWriteJointWakeBodies(rec *Recording, a *recArgsJointWakeBodies) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x94)
	recWriteArgsJointWakeBodies(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointSetConstraintTuning(rec *Recording, a *recArgsJointSetConstraintTuning) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWriteJointSetConstraintTuning(rec *Recording, a *recArgsJointSetConstraintTuning) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x95)
	recWriteArgsJointSetConstraintTuning(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointSetForceThreshold(rec *Recording, a *recArgsJointSetForceThreshold) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.threshold)
}

func recWriteJointSetForceThreshold(rec *Recording, a *recArgsJointSetForceThreshold) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x96)
	recWriteArgsJointSetForceThreshold(rec, a)
	recEndRecord(rec)
}

func recWriteArgsJointSetTorqueThreshold(rec *Recording, a *recArgsJointSetTorqueThreshold) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.threshold)
}

func recWriteJointSetTorqueThreshold(rec *Recording, a *recArgsJointSetTorqueThreshold) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x97)
	recWriteArgsJointSetTorqueThreshold(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateDistanceJoint(rec *Recording, a *recArgsCreateDistanceJoint) {
	recWDISTANCEJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateDistanceJoint(rec *Recording, a *recArgsCreateDistanceJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x98)
	recWriteArgsCreateDistanceJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateDistanceJoint(rec *Recording, a *recArgsCreateDistanceJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x98)
	recWriteArgsCreateDistanceJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetLength(rec *Recording, a *recArgsDistanceJointSetLength) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.length)
}

func recWriteDistanceJointSetLength(rec *Recording, a *recArgsDistanceJointSetLength) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x99)
	recWriteArgsDistanceJointSetLength(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointEnableSpring(rec *Recording, a *recArgsDistanceJointEnableSpring) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableSpring)
}

func recWriteDistanceJointEnableSpring(rec *Recording, a *recArgsDistanceJointEnableSpring) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x9A)
	recWriteArgsDistanceJointEnableSpring(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetSpringForceRange(rec *Recording, a *recArgsDistanceJointSetSpringForceRange) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.lowerForce)
	recWF32(&rec.buffer, a.upperForce)
}

func recWriteDistanceJointSetSpringForceRange(rec *Recording, a *recArgsDistanceJointSetSpringForceRange) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x9B)
	recWriteArgsDistanceJointSetSpringForceRange(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetSpringHertz(rec *Recording, a *recArgsDistanceJointSetSpringHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteDistanceJointSetSpringHertz(rec *Recording, a *recArgsDistanceJointSetSpringHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x9C)
	recWriteArgsDistanceJointSetSpringHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetSpringDampingRatio(rec *Recording, a *recArgsDistanceJointSetSpringDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWriteDistanceJointSetSpringDampingRatio(rec *Recording, a *recArgsDistanceJointSetSpringDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x9D)
	recWriteArgsDistanceJointSetSpringDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointEnableLimit(rec *Recording, a *recArgsDistanceJointEnableLimit) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableLimit)
}

func recWriteDistanceJointEnableLimit(rec *Recording, a *recArgsDistanceJointEnableLimit) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x9E)
	recWriteArgsDistanceJointEnableLimit(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetLengthRange(rec *Recording, a *recArgsDistanceJointSetLengthRange) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.minLength)
	recWF32(&rec.buffer, a.maxLength)
}

func recWriteDistanceJointSetLengthRange(rec *Recording, a *recArgsDistanceJointSetLengthRange) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0x9F)
	recWriteArgsDistanceJointSetLengthRange(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointEnableMotor(rec *Recording, a *recArgsDistanceJointEnableMotor) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableMotor)
}

func recWriteDistanceJointEnableMotor(rec *Recording, a *recArgsDistanceJointEnableMotor) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA0)
	recWriteArgsDistanceJointEnableMotor(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetMotorSpeed(rec *Recording, a *recArgsDistanceJointSetMotorSpeed) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.motorSpeed)
}

func recWriteDistanceJointSetMotorSpeed(rec *Recording, a *recArgsDistanceJointSetMotorSpeed) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA1)
	recWriteArgsDistanceJointSetMotorSpeed(rec, a)
	recEndRecord(rec)
}

func recWriteArgsDistanceJointSetMaxMotorForce(rec *Recording, a *recArgsDistanceJointSetMaxMotorForce) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.force)
}

func recWriteDistanceJointSetMaxMotorForce(rec *Recording, a *recArgsDistanceJointSetMaxMotorForce) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA2)
	recWriteArgsDistanceJointSetMaxMotorForce(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateFilterJoint(rec *Recording, a *recArgsCreateFilterJoint) {
	recWFILTERJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateFilterJoint(rec *Recording, a *recArgsCreateFilterJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA3)
	recWriteArgsCreateFilterJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateFilterJoint(rec *Recording, a *recArgsCreateFilterJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA3)
	recWriteArgsCreateFilterJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsCreateMotorJoint(rec *Recording, a *recArgsCreateMotorJoint) {
	recWMOTORJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateMotorJoint(rec *Recording, a *recArgsCreateMotorJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA4)
	recWriteArgsCreateMotorJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateMotorJoint(rec *Recording, a *recArgsCreateMotorJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA4)
	recWriteArgsCreateMotorJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetLinearVelocity(rec *Recording, a *recArgsMotorJointSetLinearVelocity) {
	recWJOINTID(&rec.buffer, a.joint)
	recWVEC2(&rec.buffer, a.velocity)
}

func recWriteMotorJointSetLinearVelocity(rec *Recording, a *recArgsMotorJointSetLinearVelocity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA5)
	recWriteArgsMotorJointSetLinearVelocity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetAngularVelocity(rec *Recording, a *recArgsMotorJointSetAngularVelocity) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.velocity)
}

func recWriteMotorJointSetAngularVelocity(rec *Recording, a *recArgsMotorJointSetAngularVelocity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA6)
	recWriteArgsMotorJointSetAngularVelocity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetMaxVelocityForce(rec *Recording, a *recArgsMotorJointSetMaxVelocityForce) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.maxForce)
}

func recWriteMotorJointSetMaxVelocityForce(rec *Recording, a *recArgsMotorJointSetMaxVelocityForce) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA7)
	recWriteArgsMotorJointSetMaxVelocityForce(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetMaxVelocityTorque(rec *Recording, a *recArgsMotorJointSetMaxVelocityTorque) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.maxTorque)
}

func recWriteMotorJointSetMaxVelocityTorque(rec *Recording, a *recArgsMotorJointSetMaxVelocityTorque) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA8)
	recWriteArgsMotorJointSetMaxVelocityTorque(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetLinearHertz(rec *Recording, a *recArgsMotorJointSetLinearHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteMotorJointSetLinearHertz(rec *Recording, a *recArgsMotorJointSetLinearHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xA9)
	recWriteArgsMotorJointSetLinearHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetLinearDampingRatio(rec *Recording, a *recArgsMotorJointSetLinearDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.damping)
}

func recWriteMotorJointSetLinearDampingRatio(rec *Recording, a *recArgsMotorJointSetLinearDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAA)
	recWriteArgsMotorJointSetLinearDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetAngularHertz(rec *Recording, a *recArgsMotorJointSetAngularHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteMotorJointSetAngularHertz(rec *Recording, a *recArgsMotorJointSetAngularHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAB)
	recWriteArgsMotorJointSetAngularHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetAngularDampingRatio(rec *Recording, a *recArgsMotorJointSetAngularDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.damping)
}

func recWriteMotorJointSetAngularDampingRatio(rec *Recording, a *recArgsMotorJointSetAngularDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAC)
	recWriteArgsMotorJointSetAngularDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetMaxSpringForce(rec *Recording, a *recArgsMotorJointSetMaxSpringForce) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.maxForce)
}

func recWriteMotorJointSetMaxSpringForce(rec *Recording, a *recArgsMotorJointSetMaxSpringForce) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAD)
	recWriteArgsMotorJointSetMaxSpringForce(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMotorJointSetMaxSpringTorque(rec *Recording, a *recArgsMotorJointSetMaxSpringTorque) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.maxTorque)
}

func recWriteMotorJointSetMaxSpringTorque(rec *Recording, a *recArgsMotorJointSetMaxSpringTorque) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAE)
	recWriteArgsMotorJointSetMaxSpringTorque(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateMoverJoint(rec *Recording, a *recArgsCreateMoverJoint) {
	recWMOVERJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateMoverJoint(rec *Recording, a *recArgsCreateMoverJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAF)
	recWriteArgsCreateMoverJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateMoverJoint(rec *Recording, a *recArgsCreateMoverJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xAF)
	recWriteArgsCreateMoverJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsMoverJointSetLinearVelocity(rec *Recording, a *recArgsMoverJointSetLinearVelocity) {
	recWJOINTID(&rec.buffer, a.joint)
	recWVEC2(&rec.buffer, a.velocity)
}

func recWriteMoverJointSetLinearVelocity(rec *Recording, a *recArgsMoverJointSetLinearVelocity) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB0)
	recWriteArgsMoverJointSetLinearVelocity(rec, a)
	recEndRecord(rec)
}

func recWriteArgsMoverJointSetMaxVelocityForce(rec *Recording, a *recArgsMoverJointSetMaxVelocityForce) {
	recWJOINTID(&rec.buffer, a.joint)
	recWVEC2(&rec.buffer, a.maxForce)
}

func recWriteMoverJointSetMaxVelocityForce(rec *Recording, a *recArgsMoverJointSetMaxVelocityForce) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB1)
	recWriteArgsMoverJointSetMaxVelocityForce(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreatePogoJoint(rec *Recording, a *recArgsCreatePogoJoint) {
	recWPOGOJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreatePogoJoint(rec *Recording, a *recArgsCreatePogoJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB2)
	recWriteArgsCreatePogoJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreatePogoJoint(rec *Recording, a *recArgsCreatePogoJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB2)
	recWriteArgsCreatePogoJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsPogoJointSetRestLength(rec *Recording, a *recArgsPogoJointSetRestLength) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.length)
}

func recWritePogoJointSetRestLength(rec *Recording, a *recArgsPogoJointSetRestLength) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB3)
	recWriteArgsPogoJointSetRestLength(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPogoJointSetSpringHertz(rec *Recording, a *recArgsPogoJointSetSpringHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWritePogoJointSetSpringHertz(rec *Recording, a *recArgsPogoJointSetSpringHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB4)
	recWriteArgsPogoJointSetSpringHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPogoJointSetSpringDampingRatio(rec *Recording, a *recArgsPogoJointSetSpringDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWritePogoJointSetSpringDampingRatio(rec *Recording, a *recArgsPogoJointSetSpringDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB5)
	recWriteArgsPogoJointSetSpringDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreatePrismaticJoint(rec *Recording, a *recArgsCreatePrismaticJoint) {
	recWPRISMATICJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreatePrismaticJoint(rec *Recording, a *recArgsCreatePrismaticJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB6)
	recWriteArgsCreatePrismaticJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreatePrismaticJoint(rec *Recording, a *recArgsCreatePrismaticJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB6)
	recWriteArgsCreatePrismaticJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointEnableSpring(rec *Recording, a *recArgsPrismaticJointEnableSpring) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableSpring)
}

func recWritePrismaticJointEnableSpring(rec *Recording, a *recArgsPrismaticJointEnableSpring) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB7)
	recWriteArgsPrismaticJointEnableSpring(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointSetSpringHertz(rec *Recording, a *recArgsPrismaticJointSetSpringHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWritePrismaticJointSetSpringHertz(rec *Recording, a *recArgsPrismaticJointSetSpringHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB8)
	recWriteArgsPrismaticJointSetSpringHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointSetSpringDampingRatio(rec *Recording, a *recArgsPrismaticJointSetSpringDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWritePrismaticJointSetSpringDampingRatio(rec *Recording, a *recArgsPrismaticJointSetSpringDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xB9)
	recWriteArgsPrismaticJointSetSpringDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointSetTargetTranslation(rec *Recording, a *recArgsPrismaticJointSetTargetTranslation) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.translation)
}

func recWritePrismaticJointSetTargetTranslation(rec *Recording, a *recArgsPrismaticJointSetTargetTranslation) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xBA)
	recWriteArgsPrismaticJointSetTargetTranslation(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointEnableLimit(rec *Recording, a *recArgsPrismaticJointEnableLimit) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableLimit)
}

func recWritePrismaticJointEnableLimit(rec *Recording, a *recArgsPrismaticJointEnableLimit) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xBB)
	recWriteArgsPrismaticJointEnableLimit(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointSetLimits(rec *Recording, a *recArgsPrismaticJointSetLimits) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.lower)
	recWF32(&rec.buffer, a.upper)
}

func recWritePrismaticJointSetLimits(rec *Recording, a *recArgsPrismaticJointSetLimits) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xBC)
	recWriteArgsPrismaticJointSetLimits(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointEnableMotor(rec *Recording, a *recArgsPrismaticJointEnableMotor) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableMotor)
}

func recWritePrismaticJointEnableMotor(rec *Recording, a *recArgsPrismaticJointEnableMotor) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xBD)
	recWriteArgsPrismaticJointEnableMotor(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointSetMotorSpeed(rec *Recording, a *recArgsPrismaticJointSetMotorSpeed) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.motorSpeed)
}

func recWritePrismaticJointSetMotorSpeed(rec *Recording, a *recArgsPrismaticJointSetMotorSpeed) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xBE)
	recWriteArgsPrismaticJointSetMotorSpeed(rec, a)
	recEndRecord(rec)
}

func recWriteArgsPrismaticJointSetMaxMotorForce(rec *Recording, a *recArgsPrismaticJointSetMaxMotorForce) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.force)
}

func recWritePrismaticJointSetMaxMotorForce(rec *Recording, a *recArgsPrismaticJointSetMaxMotorForce) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xBF)
	recWriteArgsPrismaticJointSetMaxMotorForce(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateRevoluteJoint(rec *Recording, a *recArgsCreateRevoluteJoint) {
	recWREVOLUTEJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateRevoluteJoint(rec *Recording, a *recArgsCreateRevoluteJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC0)
	recWriteArgsCreateRevoluteJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateRevoluteJoint(rec *Recording, a *recArgsCreateRevoluteJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC0)
	recWriteArgsCreateRevoluteJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointEnableSpring(rec *Recording, a *recArgsRevoluteJointEnableSpring) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableSpring)
}

func recWriteRevoluteJointEnableSpring(rec *Recording, a *recArgsRevoluteJointEnableSpring) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC1)
	recWriteArgsRevoluteJointEnableSpring(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointSetSpringHertz(rec *Recording, a *recArgsRevoluteJointSetSpringHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteRevoluteJointSetSpringHertz(rec *Recording, a *recArgsRevoluteJointSetSpringHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC2)
	recWriteArgsRevoluteJointSetSpringHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointSetSpringDampingRatio(rec *Recording, a *recArgsRevoluteJointSetSpringDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWriteRevoluteJointSetSpringDampingRatio(rec *Recording, a *recArgsRevoluteJointSetSpringDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC3)
	recWriteArgsRevoluteJointSetSpringDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointSetTargetAngle(rec *Recording, a *recArgsRevoluteJointSetTargetAngle) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.angle)
}

func recWriteRevoluteJointSetTargetAngle(rec *Recording, a *recArgsRevoluteJointSetTargetAngle) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC4)
	recWriteArgsRevoluteJointSetTargetAngle(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointEnableLimit(rec *Recording, a *recArgsRevoluteJointEnableLimit) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableLimit)
}

func recWriteRevoluteJointEnableLimit(rec *Recording, a *recArgsRevoluteJointEnableLimit) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC5)
	recWriteArgsRevoluteJointEnableLimit(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointSetLimits(rec *Recording, a *recArgsRevoluteJointSetLimits) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.lower)
	recWF32(&rec.buffer, a.upper)
}

func recWriteRevoluteJointSetLimits(rec *Recording, a *recArgsRevoluteJointSetLimits) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC6)
	recWriteArgsRevoluteJointSetLimits(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointEnableMotor(rec *Recording, a *recArgsRevoluteJointEnableMotor) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableMotor)
}

func recWriteRevoluteJointEnableMotor(rec *Recording, a *recArgsRevoluteJointEnableMotor) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC7)
	recWriteArgsRevoluteJointEnableMotor(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointSetMotorSpeed(rec *Recording, a *recArgsRevoluteJointSetMotorSpeed) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.motorSpeed)
}

func recWriteRevoluteJointSetMotorSpeed(rec *Recording, a *recArgsRevoluteJointSetMotorSpeed) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC8)
	recWriteArgsRevoluteJointSetMotorSpeed(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRevoluteJointSetMaxMotorTorque(rec *Recording, a *recArgsRevoluteJointSetMaxMotorTorque) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.torque)
}

func recWriteRevoluteJointSetMaxMotorTorque(rec *Recording, a *recArgsRevoluteJointSetMaxMotorTorque) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xC9)
	recWriteArgsRevoluteJointSetMaxMotorTorque(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateWeldJoint(rec *Recording, a *recArgsCreateWeldJoint) {
	recWWELDJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateWeldJoint(rec *Recording, a *recArgsCreateWeldJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCA)
	recWriteArgsCreateWeldJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateWeldJoint(rec *Recording, a *recArgsCreateWeldJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCA)
	recWriteArgsCreateWeldJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsWeldJointSetLinearHertz(rec *Recording, a *recArgsWeldJointSetLinearHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteWeldJointSetLinearHertz(rec *Recording, a *recArgsWeldJointSetLinearHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCB)
	recWriteArgsWeldJointSetLinearHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWeldJointSetLinearDampingRatio(rec *Recording, a *recArgsWeldJointSetLinearDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWriteWeldJointSetLinearDampingRatio(rec *Recording, a *recArgsWeldJointSetLinearDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCC)
	recWriteArgsWeldJointSetLinearDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWeldJointSetAngularHertz(rec *Recording, a *recArgsWeldJointSetAngularHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteWeldJointSetAngularHertz(rec *Recording, a *recArgsWeldJointSetAngularHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCD)
	recWriteArgsWeldJointSetAngularHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWeldJointSetAngularDampingRatio(rec *Recording, a *recArgsWeldJointSetAngularDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWriteWeldJointSetAngularDampingRatio(rec *Recording, a *recArgsWeldJointSetAngularDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCE)
	recWriteArgsWeldJointSetAngularDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsCreateWheelJoint(rec *Recording, a *recArgsCreateWheelJoint) {
	recWWHEELJOINTDEF(&rec.buffer, a.def)
}

func recWriteCreateWheelJoint(rec *Recording, a *recArgsCreateWheelJoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCF)
	recWriteArgsCreateWheelJoint(rec, a)
	recEndRecord(rec)
}

func recWriteRetCreateWheelJoint(rec *Recording, a *recArgsCreateWheelJoint, id Joint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xCF)
	recWriteArgsCreateWheelJoint(rec, a)
	recWJOINTID(&rec.buffer, id)
	recEndRecord(rec)
}

func recWriteArgsWheelJointEnableSpring(rec *Recording, a *recArgsWheelJointEnableSpring) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableSpring)
}

func recWriteWheelJointEnableSpring(rec *Recording, a *recArgsWheelJointEnableSpring) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD0)
	recWriteArgsWheelJointEnableSpring(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointSetSpringHertz(rec *Recording, a *recArgsWheelJointSetSpringHertz) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.hertz)
}

func recWriteWheelJointSetSpringHertz(rec *Recording, a *recArgsWheelJointSetSpringHertz) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD1)
	recWriteArgsWheelJointSetSpringHertz(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointSetSpringDampingRatio(rec *Recording, a *recArgsWheelJointSetSpringDampingRatio) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.dampingRatio)
}

func recWriteWheelJointSetSpringDampingRatio(rec *Recording, a *recArgsWheelJointSetSpringDampingRatio) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD2)
	recWriteArgsWheelJointSetSpringDampingRatio(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointEnableLimit(rec *Recording, a *recArgsWheelJointEnableLimit) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableLimit)
}

func recWriteWheelJointEnableLimit(rec *Recording, a *recArgsWheelJointEnableLimit) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD3)
	recWriteArgsWheelJointEnableLimit(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointSetLimits(rec *Recording, a *recArgsWheelJointSetLimits) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.lower)
	recWF32(&rec.buffer, a.upper)
}

func recWriteWheelJointSetLimits(rec *Recording, a *recArgsWheelJointSetLimits) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD4)
	recWriteArgsWheelJointSetLimits(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointEnableMotor(rec *Recording, a *recArgsWheelJointEnableMotor) {
	recWJOINTID(&rec.buffer, a.joint)
	recWBOOL(&rec.buffer, a.enableMotor)
}

func recWriteWheelJointEnableMotor(rec *Recording, a *recArgsWheelJointEnableMotor) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD5)
	recWriteArgsWheelJointEnableMotor(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointSetMotorSpeed(rec *Recording, a *recArgsWheelJointSetMotorSpeed) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.motorSpeed)
}

func recWriteWheelJointSetMotorSpeed(rec *Recording, a *recArgsWheelJointSetMotorSpeed) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD6)
	recWriteArgsWheelJointSetMotorSpeed(rec, a)
	recEndRecord(rec)
}

func recWriteArgsWheelJointSetMaxMotorTorque(rec *Recording, a *recArgsWheelJointSetMaxMotorTorque) {
	recWJOINTID(&rec.buffer, a.joint)
	recWF32(&rec.buffer, a.torque)
}

func recWriteWheelJointSetMaxMotorTorque(rec *Recording, a *recArgsWheelJointSetMaxMotorTorque) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xD7)
	recWriteArgsWheelJointSetMaxMotorTorque(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryOverlapAABB(rec *Recording, a *recArgsQueryOverlapAABB) {
	recWPOSITION(&rec.buffer, a.origin)
	recWAABB(&rec.buffer, a.aabb)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryOverlapAABB(rec *Recording, a *recArgsQueryOverlapAABB) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE0)
	recWriteArgsQueryOverlapAABB(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryOverlapShape(rec *Recording, a *recArgsQueryOverlapShape) {
	recWPOSITION(&rec.buffer, a.origin)
	recWSHAPEPROXY(&rec.buffer, a.proxy)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryOverlapShape(rec *Recording, a *recArgsQueryOverlapShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE1)
	recWriteArgsQueryOverlapShape(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryCastRay(rec *Recording, a *recArgsQueryCastRay) {
	recWPOSITION(&rec.buffer, a.origin)
	recWVEC2(&rec.buffer, a.translation)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryCastRay(rec *Recording, a *recArgsQueryCastRay) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE2)
	recWriteArgsQueryCastRay(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryCastShape(rec *Recording, a *recArgsQueryCastShape) {
	recWPOSITION(&rec.buffer, a.origin)
	recWSHAPEPROXY(&rec.buffer, a.proxy)
	recWVEC2(&rec.buffer, a.translation)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryCastShape(rec *Recording, a *recArgsQueryCastShape) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE3)
	recWriteArgsQueryCastShape(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryCollideMover(rec *Recording, a *recArgsQueryCollideMover) {
	recWPOSITION(&rec.buffer, a.origin)
	recWCAPSULE(&rec.buffer, a.mover)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryCollideMover(rec *Recording, a *recArgsQueryCollideMover) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE4)
	recWriteArgsQueryCollideMover(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryCastRayClosest(rec *Recording, a *recArgsQueryCastRayClosest) {
	recWPOSITION(&rec.buffer, a.origin)
	recWVEC2(&rec.buffer, a.translation)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryCastRayClosest(rec *Recording, a *recArgsQueryCastRayClosest) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE5)
	recWriteArgsQueryCastRayClosest(rec, a)
	recEndRecord(rec)
}

func recWriteArgsQueryCastMover(rec *Recording, a *recArgsQueryCastMover) {
	recWPOSITION(&rec.buffer, a.origin)
	recWCAPSULE(&rec.buffer, a.mover)
	recWVEC2(&rec.buffer, a.translation)
	recWQUERYFILTER(&rec.buffer, a.filter)
}

func recWriteQueryCastMover(rec *Recording, a *recArgsQueryCastMover) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE6)
	recWriteArgsQueryCastMover(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeTestPoint(rec *Recording, a *recArgsShapeTestPoint) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWPOSITION(&rec.buffer, a.point)
}

func recWriteShapeTestPoint(rec *Recording, a *recArgsShapeTestPoint) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE7)
	recWriteArgsShapeTestPoint(rec, a)
	recEndRecord(rec)
}

func recWriteArgsShapeRayCast(rec *Recording, a *recArgsShapeRayCast) {
	recWSHAPEID(&rec.buffer, a.shape)
	recWPOSITION(&rec.buffer, a.origin)
	recWVEC2(&rec.buffer, a.translation)
}

func recWriteShapeRayCast(rec *Recording, a *recArgsShapeRayCast) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xE8)
	recWriteArgsShapeRayCast(rec, a)
	recEndRecord(rec)
}

func recWriteArgsStateHash(rec *Recording, a *recArgsStateHash) {
	recWU64(&rec.buffer, a.hash)
}

func recWriteStateHash(rec *Recording, a *recArgsStateHash) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xF1)
	recWriteArgsStateHash(rec, a)
	recEndRecord(rec)
}

func recWriteArgsRecordingBounds(rec *Recording, a *recArgsRecordingBounds) {
	recWAABB(&rec.buffer, a.bounds)
}

func recWriteRecordingBounds(rec *Recording, a *recArgsRecordingBounds) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	recBeginRecord(rec, 0xF2)
	recWriteArgsRecordingBounds(rec, a)
	recEndRecord(rec)
}
