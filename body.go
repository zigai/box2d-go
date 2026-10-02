// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "iter"

// The body simulation type.
// Each body is one of these three types. The type determines how the body behaves in the simulation.
type BodyType int

const (
	// zero mass, zero velocity, may be manually moved
	StaticBody BodyType = 0
	// zero mass, velocity set by user, moved by solver
	KinematicBody BodyType = 1
	// positive mass, velocity determined by forces, moved by solver
	DynamicBody BodyType = 2
	// number of body types
	BodyTypeCount BodyType = 3
)

// Motion locks to restrict the body movement
type MotionLocks struct {
	// Prevent translation along the x-axis
	LinearX bool
	// Prevent translation along the y-axis
	LinearY bool
	// Prevent rotation around the z-axis
	AngularZ bool
}

// A body definition holds all the data needed to construct a rigid body.
// You can safely re-use body definitions. Shapes are added to a body after construction.
// Body definitions are temporary objects used to bundle creation parameters.
// Must be initialized using DefaultBodyDef().
type BodyDef struct {
	// The body type: static, kinematic, or dynamic.
	Type BodyType
	// The initial world position of the body. Bodies should be created with the desired position.
	// Note: Creating bodies at the origin and then moving them nearly doubles the cost of body creation, especially
	// if the body is moved after shapes have been added.
	Position Pos
	// The initial world rotation of the body. Use MakeRot() if you have an angle.
	Rotation Rot
	// The initial linear velocity of the body's origin. Usually in meters per second.
	LinearVelocity Vec2
	// The initial angular velocity of the body. Radians per second.
	AngularVelocity float32
	// Linear damping is used to reduce the linear velocity. The damping parameter
	// can be larger than 1 but the damping effect becomes sensitive to the
	// time step when the damping parameter is large.
	// Generally linear damping is undesirable because it makes objects move slowly
	// as if they are floating.
	LinearDamping float32
	// Angular damping is used to reduce the angular velocity. The damping parameter
	// can be larger than 1.0f but the damping effect becomes sensitive to the
	// time step when the damping parameter is large.
	// Angular damping can be use slow down rotating bodies.
	AngularDamping float32
	// Scale the gravity applied to this body. Non-dimensional.
	GravityScale float32
	// Sleep speed threshold, default is 0.05 meters per second
	SleepThreshold float32
	// Continuous collision safety factor. The solver only uses continuous collision if there is a
	// risk of tunneling. If the body is moving fast enough to risk tunneling then it is considered a "fast body".
	// This improves performance and prevents movement hitches. If a body moving N meter risks tunneling, then the
	// body will be considered fast if it moves more than a safetyFactor times N meters over one full time step.
	// Non-dimensional. Recommended range [0.01, 0.5]. Default is 0.5 for high performance with low tunneling risk.
	SafetyFactor float32
	// Optional body name for debugging. Up to NameLength characters
	Name string
	// Use this to store application specific body data.
	UserData any
	// Motions locks to restrict linear and angular movement.
	// Caution: may lead to softer constraints along the locked direction
	MotionLocks MotionLocks
	// Set this flag to false if this body should never fall asleep.
	EnableSleep bool
	// Is this body initially awake or sleeping?
	IsAwake bool
	// Treat this body as a high speed object that performs continuous collision detection
	// against dynamic and kinematic bodies, but not other bullet bodies.
	// Warning: Bullets should be used sparingly. They are not a solution for general dynamic-versus-dynamic
	// continuous collision. They do not guarantee accurate collision if both bodies are fast moving because
	// the bullet does a continuous check after all non-bullet bodies have moved. You could get unlucky and have
	// the bullet body end a time step very close to a non-bullet body and the non-bullet body then moves over
	// the bullet body. In continuous collision, initial overlap is ignored to avoid freezing bodies in place.
	// I do not recommend using them for game projectiles if precise collision timing is needed. Instead consider
	// using a ray or shape cast. You can use a marching ray or shape cast for projectile that moves over time.
	// If you want a fast moving projectile to collide with a fast moving target, you need to consider the relative
	// movement in your ray or shape cast. This is out of the scope of Box2D.
	// So what are good use cases for bullets? Pinball games or games with dynamic containers that hold other objects.
	// It should be a use case where it doesn't break the game if there is a collision missed, but having them
	// captured improves the quality of the game.
	IsBullet bool
	// Used to disable a body. A disabled body does not move or collide.
	IsEnabled bool
	// This allows this body to bypass rotational speed limits. Should only be used
	// for circular objects, like wheels.
	AllowFastRotation bool
	// Enable contact recycling. True by default. Leaving this enabled improves performance
	// but may lead to ghost collision that should be avoided on characters.
	EnableContactRecycling bool
	internalValue          int32
}

// Body move events triggered when a body moves.
// Triggered when a body moves due to simulation. Not reported for bodies moved by the user.
// This also has a flag to indicate that the body went to sleep so the application can also
// sleep that actor/entity/object associated with the body.
// On the other hand if the flag does not indicate the body went to sleep then the application
// can treat the actor/entity/object associated with the body as awake.
// This is an efficient way for an application to update game object transforms rather than
// calling functions such as Body.Transform() because this data is delivered as a contiguous array
// and it is only populated with bodies that have moved.
// Note: If sleeping is disabled all dynamic and kinematic bodies will trigger move events.
type BodyMoveEvent struct {
	// The body user data.
	UserData  any
	Transform WorldTransform
	BodyID    Body
	// Did the body fall asleep this time step?
	FellAsleep bool
}

// Body events are buffered in the Box2D world and are available
// as event arrays after the time step is complete.
// Note: this data becomes invalid if bodies are destroyed
type BodyEvents struct {
	// Array of move events
	MoveEvents []BodyMoveEvent
	// Number of move events
	MoveCount int
}

type body struct {
	userData       any
	setIndex       int
	localIndex     int
	headContactKey int
	contactCount   int
	headShapeID    int
	shapeCount     int
	headChainID    int
	headJointKey   int
	jointCount     int
	islandID       int
	islandIndex    int
	mass           float32
	inertia        float32
	sleepThreshold float32
	sleepTime      float32
	safetyFactor   float32
	bodyMoveIndex  int
	id             int
	flags          uint32
	bodyType       BodyType
	generation     uint16
	name           [NameLength + 1]byte
}

type bodyState struct {
	linearVelocity  Vec2
	angularVelocity float32
	flags           uint32
	deltaPosition   Vec2
	deltaRotation   Rot
}

type bodySim struct {
	transform      WorldTransform
	center         Pos
	rotation0      Rot
	center0        Pos
	localCenter    Vec2
	force          Vec2
	torque         float32
	invMass        float32
	invInertia     float32
	minExtent      float32
	maxExtent      float32
	linearDamping  float32
	angularDamping float32
	gravityScale   float32
	bodyID         int
	flags          uint32
}

type bodyFlags uint32

const (
	lockLinearX                uint32 = 0x00000001
	lockLinearY                uint32 = 0x00000002
	lockAngularZ               uint32 = 0x00000004
	isFast                     uint32 = 0x00000008
	isBullet                   uint32 = 0x00000010
	isSpeedCapped              uint32 = 0x00000020
	hadTimeOfImpact            uint32 = 0x00000040
	allowFastRotation          uint32 = 0x00000080
	enlargeBulletBounds        uint32 = 0x00000100
	dynamicFlag                uint32 = 0x00000200
	dirtyMass                  uint32 = 0x00000400
	enableSleep                uint32 = 0x00000800
	bodyEnableContactRecycling uint32 = 0x00001000
	allLocks                   uint32 = lockAngularZ | lockLinearX | lockLinearY
	fixedRotation              uint32 = lockAngularZ
	bodyTransientFlags         uint32 = isSpeedCapped | hadTimeOfImpact
)

var identityBodyState = bodyState{deltaRotation: Rot{C: 1}}

func limitVelocity(state *bodyState, maxLinearSpeed float32) {
	todo()
}

func refreshBodyContactIndices(world *world, body *body) {
	todo()
}

func getBodyFullId(world *world, bodyID Body) *body {
	todo()
	return nil
}

func getBodyTransformQuick(world *world, body *body) WorldTransform {
	todo()
	return WorldTransform{}
}

func getBodyTransform(world *world, bodyID int) WorldTransform {
	todo()
	return WorldTransform{}
}

func makeBodyId(world *world, bodyID int) Body {
	todo()
	return Body{}
}

func syncBodyFlags(world *world, body *body) {
	todo()
}

func createIslandForBody(world *world, setIndex int, body *body) {
	todo()
}

func removeBodyFromIsland(world *world, body *body) {
	todo()
}

func destroyBodyContacts(world *world, body *body) {
	todo()
}

// Create a rigid body given a definition. No reference to the definition is retained. So you can create the definition
// on the stack and pass it as a pointer.
//
//	bodyDef := DefaultBodyDef()
//	myBodyID := myWorldID.CreateBody(&bodyDef)
//
// Warning: This function is locked during callbacks.
func (w World) CreateBody(def *BodyDef) Body {
	todo()
	return Body{}
}

func wakeBody(world *world, body *body) bool {
	todo()
	return false
}

// Destroy a rigid body given an id. This destroys all shapes and joints attached to the body.
// Do not keep references to the associated shapes and joints.
func (b Body) Destroy() {
	todo()
}

// Get the maximum capacity required for retrieving all the touching contacts on a body
func (b Body) ContactCapacity() int {
	todo()
	return 0
}

// Get the touching contact data for a body.
// Note: Box2D uses speculative collision so some contact points may be separated.
func (b Body) ContactData() iter.Seq[ContactData] {
	todo()
	return nil
}

// Get the current world AABB that contains all the attached shapes. Note that this may not encompass the body origin.
// If there are no shapes attached then the returned AABB is empty and centered on the body origin.
func (b Body) ComputeAABB() AABB {
	todo()
	return AABB{}
}

func updateBodyMassData(world *world, body *body) {
	todo()
}

// Get the world position of a body. This is the location of the body origin.
func (b Body) Position() Pos {
	todo()
	return Pos{}
}

// Get the world rotation of a body as a cosine/sine pair (complex number)
func (b Body) Rotation() Rot {
	todo()
	return Rot{}
}

// Get the world transform of a body.
func (b Body) Transform() WorldTransform {
	todo()
	return WorldTransform{}
}

// Get a local point on a body given a world point
func (b Body) LocalPoint(worldPoint Pos) Vec2 {
	todo()
	return Vec2{}
}

// Get a world point on a body given a local point
func (b Body) WorldPoint(localPoint Vec2) Pos {
	todo()
	return Pos{}
}

// Get a local vector on a body given a world vector
func (b Body) LocalVector(worldVector Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Get a world vector on a body given a local vector
func (b Body) WorldVector(localVector Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Set the world transform of a body. This acts as a teleport and is fairly expensive.
// Note: Generally you should create a body with then intended transform.
// See: BodyDef.Position and BodyDef.Rotation
func (b Body) SetTransform(position Pos, rotation Rot) {
	todo()
}

// Get the linear velocity of a body's center of mass. Usually in meters per second.
func (b Body) LinearVelocity() Vec2 {
	todo()
	return Vec2{}
}

// Get the angular velocity of a body in radians per second
func (b Body) AngularVelocity() float32 {
	todo()
	return 0
}

// Set the linear velocity of a body. Usually in meters per second.
func (b Body) SetLinearVelocity(linearVelocity Vec2) {
	todo()
}

// Set the angular velocity of a body in radians per second
func (b Body) SetAngularVelocity(angularVelocity float32) {
	todo()
}

// Set the velocity to reach the given transform after a given time step.
// The result will be close but maybe not exact. This is meant for kinematic bodies.
// The target is not applied if the velocity would be below the sleep threshold and
// the body is currently asleep.
// bodyId: The body id
// target: The target transform for the body
// timeStep: The time step of the next call to world_Step
// wake: Option to wake the body or not
func (b Body) SetTargetTransform(target WorldTransform, timeStep float32, wake bool) {
	todo()
}

// Get the linear velocity of a local point attached to a body. Usually in meters per second.
func (b Body) LocalPointVelocity(localPoint Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Get the linear velocity of a world point attached to a body. Usually in meters per second.
func (b Body) WorldPointVelocity(worldPoint Pos) Vec2 {
	todo()
	return Vec2{}
}

// Apply a force at a world point. If the force is not applied at the center of mass,
// it will generate a torque and affect the angular velocity. This optionally wakes up the body.
// The force is ignored if the body is not awake.
// bodyId: The body id
// force: The world force vector, usually in newtons (N)
// point: The world position of the point of application
// wake: Option to wake up the body
func (b Body) ApplyForce(force Vec2, point Pos, wake bool) {
	todo()
}

// Apply a force to the center of mass. This optionally wakes up the body.
// The force is ignored if the body is not awake.
// bodyId: The body id
// force: the world force vector, usually in newtons (N).
// wake: also wake up the body
func (b Body) ApplyForceToCenter(force Vec2, wake bool) {
	todo()
}

// Apply a torque. This affects the angular velocity without affecting the linear velocity.
// This optionally wakes the body. The torque is ignored if the body is not awake.
// bodyId: The body id
// torque: about the z-axis (out of the screen), usually in N*m.
// wake: also wake up the body
func (b Body) ApplyTorque(torque float32, wake bool) {
	todo()
}

// Clear the force and torque on this body. Forces and torques are automatically cleared after each world
// step. So this only needs to be called if the application wants to remove the effect of previous
// calls to apply forces and torques before the world step is called.
// bodyId: The body id
func (b Body) ClearForces() {
	todo()
}

// Apply an impulse at a point. This immediately modifies the velocity.
// It also modifies the angular velocity if the point of application
// is not at the center of mass. This optionally wakes the body.
// The impulse is ignored if the body is not awake.
// bodyId: The body id
// impulse: the world impulse vector, usually in N*s or kg*m/s.
// point: the world position of the point of application.
// wake: also wake up the body
// Warning: This should be used for one-shot impulses. If you need a steady force,
// use a force instead, which will work better with the sub-stepping solver.
func (b Body) ApplyLinearImpulse(impulse Vec2, point Pos, wake bool) {
	todo()
}

// Apply an impulse to the center of mass. This immediately modifies the velocity.
// The impulse is ignored if the body is not awake. This optionally wakes the body.
// bodyId: The body id
// impulse: the world impulse vector, usually in N*s or kg*m/s.
// wake: also wake up the body
// Warning: This should be used for one-shot impulses. If you need a steady force,
// use a force instead, which will work better with the sub-stepping solver.
func (b Body) ApplyLinearImpulseToCenter(impulse Vec2, wake bool) {
	todo()
}

// Apply an angular impulse. The impulse is ignored if the body is not awake.
// This optionally wakes the body.
// bodyId: The body id
// impulse: the angular impulse, usually in units of kg*m*m/s
// wake: also wake up the body
// Warning: This should be used for one-shot impulses. If you need a steady torque,
// use a torque instead, which will work better with the sub-stepping solver.
func (b Body) ApplyAngularImpulse(impulse float32, wake bool) {
	todo()
}

// Get the body type: static, kinematic, or dynamic
func (b Body) Type() BodyType {
	todo()
	return 0
}

// Change the body type. This is an expensive operation. This automatically updates the mass
// properties regardless of the automatic mass setting.
func (b Body) SetType(bodyType BodyType) {
	todo()
}

// Set the body name. Up to 31 byteacters excluding 0 termination.
func (b Body) SetName(name string) {
	todo()
}

// Get the body name.
func (b Body) Name() string {
	todo()
	return ""
}

// Set the user data for a body
func (b Body) SetUserData(userData any) {
	todo()
}

// Get the user data stored in a body
func (b Body) UserData() any {
	todo()
	return nil
}

// Get the mass of the body, usually in kilograms
func (b Body) Mass() float32 {
	todo()
	return 0
}

// Get the rotational inertia of the body, usually in kg*m^2
func (b Body) RotationalInertia() float32 {
	todo()
	return 0
}

// Get the center of mass position of the body in local space.
func (b Body) LocalCenter() Vec2 {
	todo()
	return Vec2{}
}

// Get the center of mass position of the body in world space.
func (b Body) WorldCenter() Pos {
	todo()
	return Pos{}
}

// Override the body's mass properties. Normally this is computed automatically using the
// shape geometry and density. This information is lost if a shape is added or removed or if the
// body type changes.
// If motion locks are used to fix rotation then the rotational inertia value is ignored and
// rotational inertia remains fixed at zero.
func (b Body) SetMassData(massData MassData) {
	todo()
}

// Get the mass data for a body
func (b Body) MassData() MassData {
	todo()
	return MassData{}
}

// This updates the mass properties to the sum of the mass properties of the shapes.
// This normally does not need to be called unless you called SetMassData to override
// the mass and you later want to reset the mass.
// You may also use this when automatic mass computation has been disabled.
// You should call this regardless of body type.
// Note that sensor shapes may have mass.
func (b Body) UpdateMassFromShapes() {
	todo()
}

// Adjust the linear damping. Normally this is set in BodyDef before creation.
func (b Body) SetLinearDamping(linearDamping float32) {
	todo()
}

// Get the current linear damping.
func (b Body) LinearDamping() float32 {
	todo()
	return 0
}

// Adjust the angular damping. Normally this is set in BodyDef before creation.
func (b Body) SetAngularDamping(angularDamping float32) {
	todo()
}

// Get the current angular damping.
func (b Body) AngularDamping() float32 {
	todo()
	return 0
}

// Adjust the gravity scale. Normally this is set in BodyDef before creation.
// See: BodyDef.GravityScale
func (b Body) SetGravityScale(gravityScale float32) {
	todo()
}

// Get the current gravity scale
func (b Body) GravityScale() float32 {
	todo()
	return 0
}

// Returns: true if this body is awake
func (b Body) IsAwake() bool {
	todo()
	return false
}

// Wake a body from sleep. This wakes the entire island the body is touching.
// Warning: Putting a body to sleep will put the entire island of bodies touching this body to sleep,
// which can be expensive and possibly unintuitive.
func (b Body) SetAwake(awake bool) {
	todo()
}

// Wake bodies touching this body. Works for static bodies.
func (b Body) WakeTouching() {
	todo()
}

// Returns true if this body is enabled
func (b Body) IsEnabled() bool {
	todo()
	return false
}

// Returns true if sleeping is enabled for this body
func (b Body) IsSleepEnabled() bool {
	todo()
	return false
}

// Set the sleep threshold, usually in meters per second
func (b Body) SetSleepThreshold(sleepThreshold float32) {
	todo()
}

// Get the sleep threshold, usually in meters per second.
func (b Body) SleepThreshold() float32 {
	todo()
	return 0
}

// Set the continuous collision safety factor. Smaller is safer but can lead to hitching. Recommended range [0.01, 0.5]. Non-dimensional.
func (b Body) SetSafetyFactor(safetyFactor float32) {
	todo()
}

// Get the continuous collision safety factor. Non-dimensional.
func (b Body) SafetyFactor() float32 {
	todo()
	return 0
}

// Enable or disable sleeping for this body. If sleeping is disabled the body will wake (and the entire island).
func (b Body) EnableSleep(enableSleep bool) {
	todo()
}

// Disable a body by removing it completely from the simulation. This is expensive.
func (b Body) Disable() {
	todo()
}

// Enable a body by adding it to the simulation. This is expensive.
func (b Body) Enable() {
	todo()
}

// Set the motion locks on this body. This causes mass properties to be recomputed because
// fixed rotation sets the rotational inertia to zero.
func (b Body) SetMotionLocks(locks MotionLocks) {
	todo()
}

// Get the motion locks for this body.
func (b Body) MotionLocks() MotionLocks {
	todo()
	return MotionLocks{}
}

// Set this body to be a bullet. A bullet does continuous collision detection
// against dynamic bodies (but not other bullets).
func (b Body) SetBullet(flag bool) {
	todo()
}

// Is this body a bullet?
func (b Body) IsBullet() bool {
	todo()
	return false
}

// Enable or disable contact recycling for this body. Contact recycling is a performance optimization
// that reuses contact manifolds when bodies move slightly. Disabling it can a ghost collisions
// on byteacters at the cost of higher per-step work. Existing contacts retain their prior setting;
// only contacts created after this call see the new value.
// See: BodyDef.EnableContactRecycling
func (b Body) EnableContactRecycling(flag bool) {
	todo()
}

// Is contact recycling enabled on this body?
func (b Body) IsContactRecyclingEnabled() bool {
	todo()
	return false
}

// Enable/disable contact events on all shapes.
// See: ShapeDef.EnableContactEvents
// Warning: changing this at runtime may cause mismatched begin/end touch events
func (b Body) EnableContactEvents(flag bool) {
	todo()
}

// Enable/disable hit events on all shapes
// See: ShapeDef.EnableHitEvents
func (b Body) EnableHitEvents(flag bool) {
	todo()
}

// Get the world that owns this body
func (b Body) World() World {
	todo()
	return World{}
}

// Get the number of shapes on this body
func (b Body) ShapeCount() int {
	todo()
	return 0
}

// Get the shape ids for all shapes on this body.
func (b Body) Shapes() iter.Seq[Shape] {
	todo()
	return nil
}

// Get the number of joints on this body
func (b Body) JointCount() int {
	todo()
	return 0
}

// Get the joint ids for all joints on this body
func (b Body) Joints() iter.Seq[Joint] {
	todo()
	return nil
}

func shouldBodiesCollide(world *world, bodyA *body, bodyB *body) bool {
	todo()
	return false
}

// The minimum distance from any point on the body shapes to the center of mass.
func (b Body) MinExtent() float32 {
	todo()
	return 0
}

// The maximum distance from any point on the body shapes to the center of mass.
func (b Body) MaxExtent() float32 {
	todo()
	return 0
}

// The maximum distance from any point on the body shapes to the body origin. Conservative.
func (b Body) MaxExtentOrigin() float32 {
	todo()
	return 0
}

func makeRelativeSweep(bodySim *bodySim, base Pos) Sweep {
	todo()
	return Sweep{}
}

func getBodySim(world *world, body *body) *bodySim {
	todo()
	return nil
}

func getBodyState(world *world, body *body) *bodyState {
	todo()
	return nil
}

func encodeBodySimIndex(body *body) int {
	todo()
	return 0
}

func isStaticSimIndex(encodedBodySimIndex int) bool {
	todo()
	return false
}

func decodeAwakeIndex(encodedBodySimIndex int) int {
	todo()
	return 0
}

func sleepBodySimIndex(encodedIndex int) int {
	todo()
	return 0
}
