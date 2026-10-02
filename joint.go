// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

type JointType int32

const (
	DistanceJointType  JointType = 0
	FilterJointType    JointType = 1
	MotorJointType     JointType = 2
	MoverJointType     JointType = 3
	PogoJointType      JointType = 4
	PrismaticJointType JointType = 5
	RevoluteJointType  JointType = 6
	WeldJointType      JointType = 7
	WheelJointType     JointType = 8
)

// Base joint definition used by all joint types.
// The local frames are measured from the body's origin rather than the center of mass because:
// 1. you might not know where the center of mass will be
// 2. if you add/remove shapes from a body and recompute the mass, the joints will be broken
type JointDef struct {
	// User data pointer
	UserData any
	// The first attached body
	BodyIDA Body
	// The second attached body
	BodyIDB Body
	// The first local joint frame
	LocalFrameA Transform
	// The second local joint frame
	LocalFrameB Transform
	// Force threshold for joint events
	ForceThreshold float32
	// Torque threshold for joint events
	TorqueThreshold float32
	// Constraint hertz (advanced feature)
	ConstraintHertz float32
	// Constraint damping ratio (advanced feature)
	ConstraintDampingRatio float32
	// Debug draw scale
	DrawScale float32
	// Set this flag to true if the attached bodies should collide
	CollideConnected bool
}

type FilterJoint struct {
	Joint
}

// A filter joint is used to disable collision between two specific bodies.
type FilterJointDef struct {
	// Base joint definition
	Base JointDef
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Joint events report joints that are awake and have a force and/or torque exceeding the threshold
// The observed forces and torques are not returned for efficiency reasons.
type JointEvent struct {
	// The joint id
	JointID Joint
	// The user data from the joint for convenience
	UserData any
}

// Joint events are buffered in the world and are available
// as event arrays after the time step is complete.
// Note: this data becomes invalid if joints are destroyed
type JointEvents struct {
	// Array of events
	JointEvents []JointEvent
	// Number of events
	Count int
}

type jointEdge struct {
	bodyId  int
	prevKey int
	nextKey int
}

type joint struct {
	userData         any
	setIndex         int
	colorIndex       int
	localIndex       int
	edges            [2]jointEdge
	jointId          int
	islandId         int
	islandIndex      int
	drawScale        float32
	jointType        JointType
	generation       uint16
	collideConnected bool
}

type pogoJoint struct {
	normal              Vec2
	restLength          float32
	hertz               float32
	dampingRatio        float32
	maxTensionForce     float32
	maxCompressionForce float32
	impulse             float32
	indexA              int
	indexB              int
	frameA              Transform
	frameB              Transform
	deltaCenter         Vec2
	linearMass          float32
	velocity            float32
}

type distanceJoint struct {
	length           float32
	hertz            float32
	dampingRatio     float32
	lowerSpringForce float32
	upperSpringForce float32
	minLength        float32
	maxLength        float32
	maxMotorForce    float32
	motorSpeed       float32
	impulse          float32
	lowerImpulse     float32
	upperImpulse     float32
	motorImpulse     float32
	indexA           int
	indexB           int
	anchorA          Vec2
	anchorB          Vec2
	deltaCenter      Vec2
	distanceSoftness softness
	axialMass        float32
	enableSpring     bool
	enableLimit      bool
	enableMotor      bool
}

type motorJoint struct {
	linearVelocity         Vec2
	maxVelocityForce       float32
	angularVelocity        float32
	maxVelocityTorque      float32
	linearHertz            float32
	linearDampingRatio     float32
	maxSpringForce         float32
	angularHertz           float32
	angularDampingRatio    float32
	maxSpringTorque        float32
	linearVelocityImpulse  Vec2
	angularVelocityImpulse float32
	linearSpringImpulse    Vec2
	angularSpringImpulse   float32
	linearSpring           softness
	angularSpring          softness
	indexA                 int
	indexB                 int
	frameA                 Transform
	frameB                 Transform
	deltaCenter            Vec2
	linearMass             Mat22
	angularMass            float32
}

type moverJoint struct {
	linearVelocity        Vec2
	maxVelocityForce      Vec2
	linearVelocityImpulse Vec2
	indexA                int
	indexB                int
	frameA                Transform
	frameB                Transform
	linearMass            float32
}

type prismaticJoint struct {
	impulse           Vec2
	springImpulse     float32
	motorImpulse      float32
	lowerImpulse      float32
	upperImpulse      float32
	hertz             float32
	dampingRatio      float32
	targetTranslation float32
	maxMotorForce     float32
	motorSpeed        float32
	lowerTranslation  float32
	upperTranslation  float32
	indexA            int
	indexB            int
	frameA            Transform
	frameB            Transform
	deltaCenter       Vec2
	springSoftness    softness
	enableSpring      bool
	enableLimit       bool
	enableMotor       bool
}

type revoluteJoint struct {
	linearImpulse  Vec2
	springImpulse  float32
	motorImpulse   float32
	lowerImpulse   float32
	upperImpulse   float32
	hertz          float32
	dampingRatio   float32
	targetAngle    float32
	maxMotorTorque float32
	motorSpeed     float32
	lowerAngle     float32
	upperAngle     float32
	indexA         int
	indexB         int
	frameA         Transform
	frameB         Transform
	deltaCenter    Vec2
	axialMass      float32
	springSoftness softness
	enableSpring   bool
	enableMotor    bool
	enableLimit    bool
}

type weldJoint struct {
	linearHertz         float32
	linearDampingRatio  float32
	angularHertz        float32
	angularDampingRatio float32
	linearSpring        softness
	angularSpring       softness
	linearImpulse       Vec2
	angularImpulse      float32
	indexA              int
	indexB              int
	frameA              Transform
	frameB              Transform
	deltaCenter         Vec2
	axialMass           float32
}

type wheelJoint struct {
	perpImpulse      float32
	motorImpulse     float32
	springImpulse    float32
	lowerImpulse     float32
	upperImpulse     float32
	maxMotorTorque   float32
	motorSpeed       float32
	lowerTranslation float32
	upperTranslation float32
	hertz            float32
	dampingRatio     float32
	indexA           int
	indexB           int
	frameA           Transform
	frameB           Transform
	deltaCenter      Vec2
	perpMass         float32
	motorMass        float32
	axialMass        float32
	springSoftness   softness
	enableSpring     bool
	enableMotor      bool
	enableLimit      bool
}

type jointSim struct {
	jointId                int
	bodyIdA                int
	bodyIdB                int
	jointType              JointType
	localFrameA            Transform
	localFrameB            Transform
	invMassA               float32
	invMassB               float32
	invIA                  float32
	invIB                  float32
	constraintHertz        float32
	constraintDampingRatio float32
	constraintSoftness     softness
	forceThreshold         float32
	torqueThreshold        float32
	distanceJoint          distanceJoint
	motorJoint             motorJoint
	moverJoint             moverJoint
	pogoJoint              pogoJoint
	revoluteJoint          revoluteJoint
	prismaticJoint         prismaticJoint
	weldJoint              weldJoint
	wheelJoint             wheelJoint
}

type jointPair struct {
	joint    *joint
	jointSim *jointSim
}

func defaultJointDef() JointDef {
	return JointDef{
		LocalFrameA:            Transform{Q: Rot{C: 1}},
		LocalFrameB:            Transform{Q: Rot{C: 1}},
		ForceThreshold:         math.MaxFloat32,
		TorqueThreshold:        math.MaxFloat32,
		ConstraintHertz:        60,
		ConstraintDampingRatio: 2,
		DrawScale:              lengthUnitsPerMeter,
	}
}

// Use this to initialize your joint definition
func DefaultDistanceJointDef() DistanceJointDef {
	return DistanceJointDef{
		Base:             defaultJointDef(),
		Length:           1,
		LowerSpringForce: -math.MaxFloat32,
		UpperSpringForce: math.MaxFloat32,
		MaxLength:        Huge(),
		internalValue:    secretCookie,
	}
}

// Use this to initialize your joint definition
func DefaultFilterJointDef() FilterJointDef {
	return FilterJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultMotorJointDef() MotorJointDef {
	return MotorJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultMoverJointDef() MoverJointDef {
	return MoverJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultPogoJointDef() PogoJointDef {
	return PogoJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultPrismaticJointDef() PrismaticJointDef {
	return PrismaticJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultRevoluteJointDef() RevoluteJointDef {
	return RevoluteJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultWeldJointDef() WeldJointDef {
	return WeldJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultWheelJointDef() WheelJointDef {
	return WheelJointDef{Base: defaultJointDef(), internalValue: secretCookie, EnableSpring: true, Hertz: 1, DampingRatio: 0.7}
}

// Use this to initialize your explosion definition
func DefaultExplosionDef() ExplosionDef {
	return ExplosionDef{MaskBits: DefaultMaskBits}
}

func getJointFullId(world *world, jointID Joint) *joint {
	todo()
	return nil
}

func getJointSim(world *world, joint *joint) *jointSim {
	todo()
	return nil
}

func getJointSimCheckType(jointID Joint, jointType JointType) *jointSim {
	todo()
	return nil
}

func destroyContactsBetweenBodies(world *world, bodyA *body, bodyB *body) {
	todo()
}

func isValidJointDef(def *JointDef) bool {
	todo()
	return false
}

func createJoint(world *world, def *JointDef, jointType JointType) jointPair {
	todo()
	return jointPair{}
}

// Create a distance joint
func (w World) CreateDistanceJoint(def *DistanceJointDef) DistanceJoint {
	todo()
	return DistanceJoint{}
}

// Create a filter joint.
func (w World) CreateFilterJoint(def *FilterJointDef) FilterJoint {
	todo()
	return FilterJoint{}
}

// Create a motor joint
func (w World) CreateMotorJoint(def *MotorJointDef) MotorJoint {
	todo()
	return MotorJoint{}
}

// Create a mover joint
func (w World) CreateMoverJoint(def *MoverJointDef) MoverJoint {
	todo()
	return MoverJoint{}
}

// Create a pog joint
func (w World) CreatePogoJoint(def *PogoJointDef) PogoJoint {
	todo()
	return PogoJoint{}
}

// Create a prismatic (slider) joint.
func (w World) CreatePrismaticJoint(def *PrismaticJointDef) PrismaticJoint {
	todo()
	return PrismaticJoint{}
}

// Create a revolute joint
func (w World) CreateRevoluteJoint(def *RevoluteJointDef) RevoluteJoint {
	todo()
	return RevoluteJoint{}
}

// Create a weld joint
func (w World) CreateWeldJoint(def *WeldJointDef) WeldJoint {
	todo()
	return WeldJoint{}
}

// Create a wheel joint
func (w World) CreateWheelJoint(def *WheelJointDef) WheelJoint {
	todo()
	return WheelJoint{}
}

func destroyJointInternal(world *world, joint *joint) {
	todo()
}

// Destroy a joint.
func (j Joint) Destroy() {
	todo()
}

// Get the joint type
func (j Joint) Type() JointType {
	todo()
	return 0
}

// Get body A id on a joint
func (j Joint) BodyA() Body {
	todo()
	return Body{}
}

// Get body B id on a joint
func (j Joint) BodyB() Body {
	todo()
	return Body{}
}

// Get the world that owns this joint
func (j Joint) World() World {
	todo()
	return World{}
}

// Set the local frame on body A
func (j Joint) SetLocalFrameA(localFrame Transform) {
	todo()
}

// Get the local frame on body A
func (j Joint) LocalFrameA() Transform {
	todo()
	return Transform{}
}

// Set the local frame on body B
func (j Joint) SetLocalFrameB(localFrame Transform) {
	todo()
}

// Get the local frame on body B
func (j Joint) LocalFrameB() Transform {
	todo()
	return Transform{}
}

// Toggle collision between connected bodies
func (j Joint) SetCollideConnected(shouldCollide bool) {
	todo()
}

// Is collision allowed between connected bodies?
func (j Joint) CollideConnected() bool {
	todo()
	return false
}

// Set the user data on a joint
func (j Joint) SetUserData(userData any) {
	todo()
}

// Get the user data on a joint
func (j Joint) UserData() any {
	todo()
	return nil
}

// Wake the bodies connect to this joint
func (j Joint) WakeBodies() {
	todo()
}

func getJointReaction(sim *jointSim, invTimeStep float32) (force float32, torque float32) {
	todo()
	return 0, 0
}

func getJointConstraintForce(world *world, joint *joint) Vec2 {
	todo()
	return Vec2{}
}

func getJointConstraintTorque(world *world, joint *joint) float32 {
	todo()
	return 0
}

// Get the current constraint force for this joint. Usually in Newtons.
func (j Joint) ConstraintForce() Vec2 {
	todo()
	return Vec2{}
}

// Get the current constraint torque for this joint. Usually in Newton * meters.
func (j Joint) ConstraintTorque() float32 {
	todo()
	return 0
}

// Get the current linear separation error for this joint. Does not consider admissible movement. Usually in meters.
func (j Joint) LinearSeparation() float32 {
	todo()
	return 0
}

// Get the current angular separation error for this joint. Does not consider admissible movement. Usually in meters.
func (j Joint) AngularSeparation() float32 {
	todo()
	return 0
}

// Set the joint constraint tuning. Advanced feature.
func (j Joint) SetConstraintTuning(hertz float32, dampingRatio float32) {
	todo()
}

// Get the joint constraint tuning. Advanced feature.
func (j Joint) ConstraintTuning() (hertz float32, dampingRatio float32) {
	todo()
	return 0, 0
}

// Set the force threshold for joint events (Newtons)
func (j Joint) SetForceThreshold(threshold float32) {
	todo()
}

// Get the force threshold for joint events (Newtons)
func (j Joint) ForceThreshold() float32 {
	todo()
	return 0
}

// Set the torque threshold for joint events (N-m)
func (j Joint) SetTorqueThreshold(threshold float32) {
	todo()
}

// Get the torque threshold for joint events (N-m)
func (j Joint) TorqueThreshold() float32 {
	todo()
	return 0
}

func prepareJoint(joint *jointSim, context *stepContext) {
	todo()
}

func warmStartJoint(joint *jointSim, context *stepContext) {
	todo()
}

func solveJoint(joint *jointSim, context *stepContext, useBias bool) {
	todo()
}

func prepareJointsOverflow(context *stepContext) {
	todo()
}

func warmStartJointsOverflow(context *stepContext) {
	todo()
}

func solveJointsOverflow(context *stepContext, useBias bool) {
	todo()
}

func prepareJointsTask(block solverBlock, context *stepContext) {
	todo()
}

func warmStartJointsTask(block solverBlock, context *stepContext) {
	todo()
}

func solveJointsTask(block solverBlock, context *stepContext, useBias bool, workerIndex int) {
	todo()
}

func drawJoint(draw *DebugDraw, world *world, joint *joint) {
	todo()
}
