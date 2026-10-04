// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"iter"
	"math"
	"strings"
)

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
	Body      Body
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
	var v2 float32 = state.linearVelocity.LengthSquared()
	if v2 > maxLinearSpeed*maxLinearSpeed {
		state.linearVelocity = state.linearVelocity.Scale(maxLinearSpeed / float32(math.Sqrt(float64(v2))))
	}
}

func refreshBodyContactIndices(world *World, body *body) {
	encodedIndex := encodeBodySimIndex(body)

	contactKey := body.headContactKey
	for contactKey != nullIndex {
		contactID := contactKey >> 1
		edgeIndex := contactKey & 1

		contact := &world.contacts[contactID]
		contactSim := getContactSim(world, contact)

		if edgeIndex == 0 {
			contactSim.encodedBodySimA = encodedIndex
		} else {
			contactSim.encodedBodySimB = encodedIndex
		}

		contactKey = contact.edges[edgeIndex].nextKey
	}
}

func getBodyFullId(world *World, bodyID Body) *body {
	assert(bodyID.IsValid(), "b2Body_IsValid( bodyId )")

	// id index starts at one so that zero can represent null
	return &world.bodies[bodyID.index1-1]
}

func getBodyTransformQuick(world *World, body *body) WorldTransform {
	set := &world.solverSets[body.setIndex]
	bodySim := &set.bodySims[body.localIndex]
	return bodySim.transform
}

func getBodyTransform(world *World, bodyID int) WorldTransform {
	body := &world.bodies[bodyID]
	return getBodyTransformQuick(world, body)
}

func makeBodyId(world *World, bodyID int) Body {
	body := &world.bodies[bodyID]
	return Body{int32(bodyID + 1), world, body.generation}
}

func syncBodyFlags(world *World, body *body) {
	bodySim := getBodySim(world, body)

	// Preserve the sim only flags: fast for contact recycling, time of impact for debug draw.
	bodySim.flags = (bodySim.flags & (isFast | hadTimeOfImpact)) | (body.flags & ^bodyTransientFlags)

	bodyState := getBodyState(world, body)
	if bodyState != nil {
		bodyState.flags = body.flags & ^bodyTransientFlags
	}
}

func createIslandForBody(world *World, setIndex int, body *body) {
	assert(body.islandID == nullIndex, "body->islandId == B2_NULL_INDEX")
	assert(setIndex != disabledSet, "setIndex != b2_disabledSet")

	island := createIsland(world, setIndex)
	island.bodies = arrayPush(island.bodies, body.id)
	body.islandID = island.islandID
	body.islandIndex = 0

	validateIsland(world, island.islandID)
}

func removeBodyFromIsland(world *World, body *body) {
	if body.islandID == nullIndex {
		assert(body.islandIndex == nullIndex, "body->islandIndex == B2_NULL_INDEX")
		return
	}

	islandID := body.islandID
	island := &world.islands[islandID]
	{
		localIndex := body.islandIndex
		movedBodyID := island.bodies[len(island.bodies)-1]
		island.bodies[localIndex] = movedBodyID
		validate(world.bodies[movedBodyID].islandIndex == len(island.bodies)-1, "world->bodies.data[movedBodyId].islandIndex == island->bodies.count - 1")
		world.bodies[movedBodyID].islandIndex = localIndex
		island.bodies = island.bodies[:len(island.bodies)-1]
	}

	if len(island.bodies) == 0 {
		// Destroy empty island
		assert(len(island.contacts) == 0, "island->contacts.count == 0")
		assert(len(island.joints) == 0, "island->joints.count == 0")

		// Free the island
		destroyIsland(world, island.islandID)
	} else {
		validateIsland(world, islandID)
	}

	body.islandID = nullIndex
	body.islandIndex = nullIndex
}

func destroyBodyContacts(world *World, body *body) {
	// Destroy the attached contacts
	edgeKey := body.headContactKey
	for edgeKey != nullIndex {
		contactID := edgeKey >> 1
		edgeIndex := edgeKey & 1

		contact := &world.contacts[contactID]
		edgeKey = contact.edges[edgeIndex].nextKey
		destroyContact(world, contact)
	}

	validateSolverSets(world)
}

// Create a rigid body given a definition. No reference to the definition is retained. So you can create the definition
// on the stack and pass it as a pointer.
//
//	bodyDef := DefaultBodyDef()
//	myBodyID := myWorldID.CreateBody(&bodyDef)
//
// Warning: This function is locked during callbacks.
func (w *World) CreateBody(def *BodyDef) Body {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(def.Position.IsValid(), "b2IsValidPosition( def->position )")
	checkInput(def.Rotation.IsValid(), "b2IsValidRotation( def->rotation )")
	checkInput(def.LinearVelocity.IsValid(), "b2IsValidVec2( def->linearVelocity )")
	checkInput(IsValidFloat(def.AngularVelocity), "b2IsValidFloat( def->angularVelocity )")
	checkInput(IsValidFloat(def.LinearDamping) && def.LinearDamping >= 0.0, "b2IsValidFloat( def->linearDamping ) && def->linearDamping >= 0.0f")
	checkInput(IsValidFloat(def.AngularDamping) && def.AngularDamping >= 0.0, "b2IsValidFloat( def->angularDamping ) && def->angularDamping >= 0.0f")
	checkInput(IsValidFloat(def.SleepThreshold) && def.SleepThreshold >= 0.0, "b2IsValidFloat( def->sleepThreshold ) && def->sleepThreshold >= 0.0f")
	checkInput(IsValidFloat(def.SafetyFactor) && def.SafetyFactor >= 0.0, "b2IsValidFloat( def->safetyFactor ) && def->safetyFactor >= 0.0f")
	checkInput(IsValidFloat(def.GravityScale), "b2IsValidFloat( def->gravityScale )")

	world := w
	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return Body{}
	}

	isAwake := (def.IsAwake || def.EnableSleep == false) && def.IsEnabled

	// determine the solver set
	var setID int
	if def.IsEnabled == false {
		// any body type can be disabled
		setID = disabledSet
	} else if def.Type == StaticBody {
		setID = staticSet
	} else if isAwake == true {
		setID = awakeSet
	} else {
		// new set for a sleeping body in its own island
		setID = allocID(&world.solverSetIDPool)
		if setID == len(world.solverSets) {
			// Create a zero initialized solver set. All sub-arrays are also zero initialized.
			world.solverSets = arrayPush(world.solverSets, solverSet{})
		} else {
			assert(world.solverSets[setID].setIndex == nullIndex, "world->solverSets.data[setId].setIndex == B2_NULL_INDEX")
		}

		world.solverSets[setID].setIndex = setID
	}

	assert(0 <= setID && setID < len(world.solverSets), "0 <= setId && setId < world->solverSets.count")

	bodyID := allocID(&world.bodyIDPool)

	var lockFlags uint32 = 0
	if def.MotionLocks.LinearX {
		lockFlags |= lockLinearX
	}
	if def.MotionLocks.LinearY {
		lockFlags |= lockLinearY
	}
	if def.MotionLocks.AngularZ {
		lockFlags |= lockAngularZ
	}

	set := &world.solverSets[setID]
	var sim *bodySim
	set.bodySims, sim = arrayEmplace(set.bodySims)
	*sim = bodySim{}
	sim.transform.P = def.Position
	sim.transform.Q = def.Rotation
	sim.center = def.Position
	sim.rotation0 = sim.transform.Q
	sim.center0 = sim.center
	sim.minExtent = Huge()
	sim.maxExtent = 0.0
	sim.linearDamping = def.LinearDamping
	sim.angularDamping = def.AngularDamping
	sim.gravityScale = def.GravityScale
	sim.bodyID = bodyID
	sim.flags = lockFlags
	if def.IsBullet {
		sim.flags |= isBullet
	}
	if def.AllowFastRotation {
		sim.flags |= allowFastRotation
	}
	if def.Type == DynamicBody {
		sim.flags |= dynamicFlag
	}
	if def.EnableSleep {
		sim.flags |= enableSleep
	}
	if def.EnableContactRecycling {
		sim.flags |= bodyEnableContactRecycling
	}

	if setID == awakeSet {
		var state *bodyState
		set.bodyStates, state = arrayEmplace(set.bodyStates)

		*state = bodyState{}
		state.linearVelocity = def.LinearVelocity
		state.angularVelocity = def.AngularVelocity
		state.deltaRotation = Rot{C: 1}
		state.flags = sim.flags
	}

	if bodyID == len(world.bodies) {
		world.bodies = arrayPush(world.bodies, body{})
	} else {
		assert(world.bodies[bodyID].id == nullIndex, "world->bodies.data[bodyId].id == B2_NULL_INDEX")
	}

	body := &world.bodies[bodyID]

	clear(body.name[:])
	name := def.Name
	if end := strings.IndexByte(name, 0); end >= 0 {
		name = name[:end]
	}
	copy(body.name[:NameLength], name)

	body.userData = def.UserData
	body.setIndex = setID
	body.localIndex = len(set.bodySims) - 1
	body.generation += 1
	body.headShapeID = nullIndex
	body.shapeCount = 0
	body.headChainID = nullIndex
	body.headContactKey = nullIndex
	body.contactCount = 0
	body.headJointKey = nullIndex
	body.jointCount = 0
	body.islandID = nullIndex
	body.islandIndex = nullIndex
	body.bodyMoveIndex = nullIndex
	body.id = bodyID
	body.mass = 0.0
	body.inertia = 0.0
	body.sleepThreshold = def.SleepThreshold
	body.sleepTime = 0.0
	body.safetyFactor = def.SafetyFactor
	body.bodyType = def.Type
	body.flags = sim.flags

	// dynamic and kinematic bodies that are enabled need a island
	if setID >= awakeSet {
		createIslandForBody(world, setID, body)
	}

	validateSolverSets(world)

	id := Body{int32(bodyID + 1), world, body.generation}

	if world.recording != nil {
		recWriteRetCreateBody(world.recording, &recArgsCreateBody{*def}, id)
	}

	return id
}

func wakeBody(world *World, body *body) bool {
	if body.setIndex >= firstSleepingSet {
		wakeSolverSet(world, body.setIndex)
		validateSolverSets(world)
		return true
	}

	return false
}

// Destroy a rigid body given an id. This destroys all shapes and joints attached to the body.
// Do not keep references to the associated shapes and joints.
func (b Body) Destroy() {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	// Record before destroying (body must still be valid)
	if world.recording != nil {
		recWriteDestroyBody(world.recording, &recArgsDestroyBody{b})
	}

	body := getBodyFullId(world, b)

	// Destroy the attached joints
	edgeKey := body.headJointKey
	for edgeKey != nullIndex {
		jointID := edgeKey >> 1
		edgeIndex := edgeKey & 1

		joint := &world.joints[jointID]
		edgeKey = joint.edges[edgeIndex].nextKey

		// Careful because this modifies the list being traversed
		destroyJointInternal(world, joint)
	}

	// Destroy all contacts attached to this body.
	destroyBodyContacts(world, body)

	// Destroy the attached shapes and their broad-phase proxies.
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]

		if shape.sensorIndex != nullIndex {
			destroySensor(world, shape)
		}

		destroyShapeProxy(shape, &world.broadPhase)

		// Return shape to free list.
		freeID(&world.shapeIDPool, shapeID)
		shape.id = nullIndex

		shapeID = shape.nextShapeID
	}

	// Destroy the attached chains. The associated shapes have already been destroyed above.
	chainID := body.headChainID
	for chainID != nullIndex {
		chain := &world.chainShapes[chainID]

		freeChainData(chain)

		// Return chain to free list.
		freeID(&world.chainIDPool, chainID)
		chain.id = nullIndex

		chainID = chain.nextChainID
	}

	removeBodyFromIsland(world, body)

	// Remove body sim from solver set that owns it
	set := &world.solverSets[body.setIndex]
	movedBody := removeBodySim(world, set, body.localIndex)
	if movedBody != nil {
		refreshBodyContactIndices(world, movedBody)
	}

	// Remove body state from awake set
	if body.setIndex == awakeSet {
		set.bodyStates, _ = arrayRemoveSwap(set.bodyStates, body.localIndex)
	} else if set.setIndex >= firstSleepingSet && len(set.bodySims) == 0 {
		// Remove solver set if it is empty
		destroySolverSet(world, set.setIndex)
	}

	// Free body and id (preserve body generation)
	freeID(&world.bodyIDPool, body.id)

	body.setIndex = nullIndex
	body.localIndex = nullIndex
	body.id = nullIndex

	validateSolverSets(world)
}

// Get the maximum capacity required for retrieving all the touching contacts on a body
func (b Body) ContactCapacity() int {
	world := getWorldLocked(b.world)
	if world == nil {
		return 0
	}

	body := getBodyFullId(world, b)

	// Conservative and fast
	return body.contactCount
}

// Get the touching contact data for a body.
// Note: Box2D uses speculative collision so some contact points may be separated.
func (b Body) ContactData() iter.Seq[ContactData] {
	world := getWorldLocked(b.world)
	if world == nil {
		return func(yield func(ContactData) bool) {}
	}
	body := getBodyFullId(world, b)
	return func(yield func(ContactData) bool) {
		contactKey := body.headContactKey
		for contactKey != nullIndex {
			contactID := contactKey >> 1
			edgeIndex := contactKey & 1
			contact := &world.contacts[contactID]
			// Is contact touching?
			if contact.flags&contactTouchingFlag != 0 {
				shapeA := &world.shapes[contact.shapeIDA]
				shapeB := &world.shapes[contact.shapeIDB]
				data := ContactData{
					Contact: Contact{int32(contact.contactID + 1), b.world, contact.generation},
					ShapeA:  Shape{int32(shapeA.id + 1), b.world, shapeA.generation},
					ShapeB:  Shape{int32(shapeB.id + 1), b.world, shapeB.generation},
				}
				contactSim := getContactSim(world, contact)
				data.Manifold = contactSim.manifold
				if !yield(data) {
					return
				}
			}
			contactKey = contact.edges[edgeIndex].nextKey
		}
	}
}

// Get the current world AABB that contains all the attached shapes. Note that this may not encompass the body origin.
// If there are no shapes attached then the returned AABB is empty and centered on the body origin.
func (b Body) ComputeAABB() AABB {
	world := getWorldLocked(b.world)
	if world == nil {
		return AABB{}
	}

	body := getBodyFullId(world, b)
	if body.headShapeID == nullIndex {
		// No shapes, bracket the body origin so the box still contains the true position far away
		transform := getBodyTransform(world, body.id)
		lower := Vec2{RoundDownFloat(float64(transform.P.X)), RoundDownFloat(float64(transform.P.Y))}
		upper := Vec2{RoundUpFloat(float64(transform.P.X)), RoundUpFloat(float64(transform.P.Y))}
		return AABB{lower, upper}
	}

	shape := &world.shapes[body.headShapeID]
	aabb := shape.aabb
	for shape.nextShapeID != nullIndex {
		shape = &world.shapes[shape.nextShapeID]
		aabb = aabb.Union(shape.aabb)
	}

	return aabb
}

func updateBodyMassData(world *World, body *body) {
	bodySim := getBodySim(world, body)

	// Mass is no longer dirty
	body.flags &= ^dirtyMass

	// Compute mass data from shapes. Each shape has its own density.
	body.mass = 0.0
	body.inertia = 0.0

	bodySim.invMass = 0.0
	bodySim.invInertia = 0.0
	bodySim.localCenter = Vec2{}
	bodySim.minExtent = Huge()
	bodySim.maxExtent = 0.0

	// Static and kinematic sims have zero mass.
	if body.bodyType != DynamicBody {
		bodySim.center = bodySim.transform.P
		bodySim.center0 = bodySim.center

		// Need extents for kinematic bodies for sleeping to work correctly.
		if body.bodyType == KinematicBody {
			shapeID := body.headShapeID
			for shapeID != nullIndex {
				s := &world.shapes[shapeID]

				extent := computeShapeExtent(s, Vec2{})
				bodySim.minExtent = minFloat(bodySim.minExtent, extent.minExtent)
				bodySim.maxExtent = maxFloat(bodySim.maxExtent, extent.maxExtent)

				shapeID = s.nextShapeID
			}
		}

		return
	}

	shapeCount := body.shapeCount
	masses := make([]MassData, shapeCount)

	// Accumulate mass over all shapes.
	localCenter := Vec2{}
	shapeID := body.headShapeID
	shapeIndex := 0
	for shapeID != nullIndex {
		s := &world.shapes[shapeID]
		shapeID = s.nextShapeID

		if s.density == 0.0 {
			masses[shapeIndex] = MassData{}
			shapeIndex += 1
			continue
		}

		massData := computeShapeMass(s)
		body.mass += massData.Mass
		localCenter = localCenter.MulAdd(massData.Mass, massData.Center)

		masses[shapeIndex] = massData
		shapeIndex += 1
	}

	// Compute center of mass.
	if body.mass > 0.0 {
		bodySim.invMass = 1.0 / body.mass
		localCenter = localCenter.Scale(bodySim.invMass)
	}

	// Second loop to accumulate the rotational inertia about the center of mass
	for shapeIndex = 0; shapeIndex < shapeCount; shapeIndex++ {
		massData := masses[shapeIndex]
		if massData.Mass == 0.0 {
			continue
		}

		// Shift to center of mass. This is safe because it can only increase.
		offset := localCenter.Sub(massData.Center)
		var inertia float32 = massData.RotationalInertia + float32(massData.Mass*offset.Dot(offset))
		body.inertia += inertia
	}

	masses = nil

	assert(body.inertia >= 0.0, "body->inertia >= 0.0f")

	if body.inertia > 0.0 && (body.flags&fixedRotation) == 0 {
		bodySim.invInertia = 1.0 / body.inertia
	} else {
		body.inertia = 0.0
		bodySim.invInertia = 0.0
	}

	// Move center of mass.
	oldCenter := bodySim.center
	bodySim.localCenter = localCenter
	bodySim.center = bodySim.transform.Apply(bodySim.localCenter)
	bodySim.center0 = bodySim.center

	// Update center of mass velocity
	state := getBodyState(world, body)
	if state != nil {
		deltaLinear := bodySim.center.Sub(oldCenter).ScalarCross(state.angularVelocity)
		state.linearVelocity = state.linearVelocity.Add(deltaLinear)
	}

	// Compute body extents relative to center of mass
	shapeID = body.headShapeID
	for shapeID != nullIndex {
		s := &world.shapes[shapeID]

		extent := computeShapeExtent(s, localCenter)
		bodySim.minExtent = minFloat(bodySim.minExtent, extent.minExtent)
		bodySim.maxExtent = maxFloat(bodySim.maxExtent, extent.maxExtent)

		shapeID = s.nextShapeID
	}

	// When the center of mass changes, any cached contact manifold becomes invalid.
	edgeKey := body.headContactKey
	for edgeKey != nullIndex {
		contactID := edgeKey >> 1
		edgeIndex := edgeKey & 1

		contact := &world.contacts[contactID]
		contactSim := getContactSim(world, contact)
		contactSim.simFlags &= ^simRelativeTransformValid

		edgeKey = contact.edges[edgeIndex].nextKey
	}
}

// Get the world position of a body. This is the location of the body origin.
func (b Body) Position() Pos {
	world := b.world
	body := getBodyFullId(world, b)
	transform := getBodyTransformQuick(world, body)
	return transform.P
}

// Get the world rotation of a body as a cosine/sine pair (complex number)
func (b Body) Rotation() Rot {
	world := b.world
	body := getBodyFullId(world, b)
	transform := getBodyTransformQuick(world, body)
	return transform.Q
}

// Get the world transform of a body.
func (b Body) Transform() WorldTransform {
	world := b.world
	body := getBodyFullId(world, b)
	return getBodyTransformQuick(world, body)
}

// Get a local point on a body given a world point
func (b Body) LocalPoint(worldPoint Pos) Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	transform := getBodyTransformQuick(world, body)
	return transform.InvApply(worldPoint)
}

// Get a world point on a body given a local point
func (b Body) WorldPoint(localPoint Vec2) Pos {
	world := b.world
	body := getBodyFullId(world, b)
	transform := getBodyTransformQuick(world, body)
	return transform.Apply(localPoint)
}

// Get a local vector on a body given a world vector
func (b Body) LocalVector(worldVector Vec2) Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	transform := getBodyTransformQuick(world, body)
	return transform.Q.InvRotate(worldVector)
}

// Get a world vector on a body given a local vector
func (b Body) WorldVector(localVector Vec2) Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	transform := getBodyTransformQuick(world, body)
	return transform.Q.Rotate(localVector)
}

// Set the world transform of a body. This acts as a teleport and is fairly expensive.
// Note: Generally you should create a body with then intended transform.
// See: BodyDef.Position and BodyDef.Rotation
func (b Body) SetTransform(position Pos, rotation Rot) {
	checkInput(position.IsValid(), "b2IsValidPosition( position )")
	checkInput(rotation.IsValid(), "b2IsValidRotation( rotation )")
	assert(b.IsValid(), "b2Body_IsValid( bodyId )")
	world := b.world
	assert(world.locked == false, "world->locked == false")

	if world.recording != nil {
		recWriteBodySetTransform(world.recording, &recArgsBodySetTransform{b, position, rotation})
	}

	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)

	bodySim.transform.P = position
	bodySim.transform.Q = rotation
	bodySim.center = bodySim.transform.Apply(bodySim.localCenter)

	bodySim.rotation0 = bodySim.transform.Q
	bodySim.center0 = bodySim.center

	broadPhase := &world.broadPhase

	transform := bodySim.transform
	var speculativeDistance float32 = SpeculativeDistance()

	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]
		aabb := computeFatShapeAABB(shape, transform, speculativeDistance)
		shape.aabb = aabb

		shapeFatAABB := &world.fatAABBs[shapeID]
		if shapeFatAABB.Contains(aabb) == false {
			var margin float32 = shape.aabbMargin
			var fatAABB AABB
			fatAABB.LowerBound.X = aabb.LowerBound.X - margin
			fatAABB.LowerBound.Y = aabb.LowerBound.Y - margin
			fatAABB.UpperBound.X = aabb.UpperBound.X + margin
			fatAABB.UpperBound.Y = aabb.UpperBound.Y + margin
			*shapeFatAABB = fatAABB

			// They body could be disabled
			if shape.proxyKey != nullIndex {
				broadPhase.moveProxy(shape.proxyKey, fatAABB)
			}
		}

		shapeID = shape.nextShapeID
	}
}

// Get the linear velocity of a body's center of mass. Usually in meters per second.
func (b Body) LinearVelocity() Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	state := getBodyState(world, body)
	if state != nil {
		return state.linearVelocity
	}
	return Vec2{}
}

// Get the angular velocity of a body in radians per second
func (b Body) AngularVelocity() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	state := getBodyState(world, body)
	if state != nil {
		return state.angularVelocity
	}
	return 0.0
}

// Set the linear velocity of a body. Usually in meters per second.
func (b Body) SetLinearVelocity(linearVelocity Vec2) {
	checkInput(linearVelocity.IsValid(), "b2IsValidVec2( linearVelocity )")

	world := b.world

	if world.recording != nil {
		recWriteBodySetLinearVelocity(world.recording, &recArgsBodySetLinearVelocity{b, linearVelocity})
	}

	body := getBodyFullId(world, b)

	if body.bodyType == StaticBody {
		return
	}

	if linearVelocity.LengthSquared() > 0.0 {
		wakeBody(world, body)
	}

	state := getBodyState(world, body)
	if state == nil {
		return
	}

	state.linearVelocity = linearVelocity
}

// Set the angular velocity of a body in radians per second
func (b Body) SetAngularVelocity(angularVelocity float32) {
	checkInput(IsValidFloat(angularVelocity), "b2IsValidFloat( angularVelocity )")

	world := b.world
	if world.recording != nil {
		recWriteBodySetAngularVelocity(world.recording, &recArgsBodySetAngularVelocity{b, angularVelocity})
	}
	body := getBodyFullId(world, b)

	if body.bodyType == StaticBody || (body.flags&lockAngularZ) != 0 {
		return
	}

	if angularVelocity != 0.0 {
		wakeBody(world, body)
	}

	state := getBodyState(world, body)
	if state == nil {
		return
	}

	state.angularVelocity = angularVelocity
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
	checkInput(target.IsValid(), "b2IsValidWorldTransform( target )")
	checkInput(IsValidFloat(timeStep), "b2IsValidFloat( timeStep )")

	world := b.world

	if world.recording != nil {
		recWriteBodySetTargetTransform(world.recording, &recArgsBodySetTargetTransform{b, target, timeStep, wake})
	}
	body := getBodyFullId(world, b)

	if body.setIndex == disabledSet {
		return
	}

	if body.bodyType == StaticBody || timeStep <= 0.0 {
		return
	}

	if body.setIndex != awakeSet && wake == false {
		return
	}

	sim := getBodySim(world, body)

	// Compute linear velocity. The center difference is taken in world precision then demoted
	delta := target.Apply(sim.localCenter).Sub(sim.center)
	var invTimeStep float32 = 1.0 / timeStep
	linearVelocity := delta.Scale(invTimeStep)

	// Compute angular velocity
	q1 := sim.transform.Q
	q2 := target.Q
	var deltaAngle float32 = q1.RelativeAngle(q2)
	var angularVelocity float32 = invTimeStep * deltaAngle

	// Early out if the body is asleep already and the desired movement is small
	if body.setIndex != awakeSet {
		var maxVelocity float32 = linearVelocity.Length() + float32(absFloat(angularVelocity)*sim.maxExtent)

		// Return if velocity would be sleepy
		if maxVelocity < body.sleepThreshold {
			return
		}

		// Must wake for state to exist
		wakeBody(world, body)
	}

	assert(body.setIndex == awakeSet, "body->setIndex == b2_awakeSet")

	state := getBodyState(world, body)
	state.linearVelocity = linearVelocity
	state.angularVelocity = angularVelocity
}

// Get the linear velocity of a local point attached to a body. Usually in meters per second.
func (b Body) LocalPointVelocity(localPoint Vec2) Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	state := getBodyState(world, body)
	if state == nil {
		return Vec2{}
	}

	set := &world.solverSets[body.setIndex]
	bodySim := &set.bodySims[body.localIndex]

	r := bodySim.transform.Q.Rotate(localPoint.Sub(bodySim.localCenter))
	v := state.linearVelocity.Add(r.ScalarCross(state.angularVelocity))
	return v
}

// Get the linear velocity of a world point attached to a body. Usually in meters per second.
func (b Body) WorldPointVelocity(worldPoint Pos) Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	state := getBodyState(world, body)
	if state == nil {
		return Vec2{}
	}

	set := &world.solverSets[body.setIndex]
	bodySim := &set.bodySims[body.localIndex]

	r := worldPoint.Sub(bodySim.center)
	v := state.linearVelocity.Add(r.ScalarCross(state.angularVelocity))
	return v
}

// Apply a force at a world point. If the force is not applied at the center of mass,
// it will generate a torque and affect the angular velocity. This optionally wakes up the body.
// The force is ignored if the body is not awake.
// bodyId: The body id
// force: The world force vector, usually in newtons (N)
// point: The world position of the point of application
// wake: Option to wake up the body
func (b Body) ApplyForce(force Vec2, point Pos, wake bool) {
	checkInput(force.IsValid(), "b2IsValidVec2( force )")
	checkInput(point.IsValid(), "b2IsValidPosition( point )")

	world := b.world
	if world.recording != nil {
		recWriteBodyApplyForce(world.recording, &recArgsBodyApplyForce{b, force, point, wake})
	}
	body := getBodyFullId(world, b)

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if wake && body.setIndex >= firstSleepingSet {
		wakeBody(world, body)
	}

	if body.setIndex == awakeSet {
		bodySim := getBodySim(world, body)
		bodySim.force = bodySim.force.Add(force)
		bodySim.torque += point.Sub(bodySim.center).Cross(force)
	}
}

// Apply a force to the center of mass. This optionally wakes up the body.
// The force is ignored if the body is not awake.
// bodyId: The body id
// force: the world force vector, usually in newtons (N).
// wake: also wake up the body
func (b Body) ApplyForceToCenter(force Vec2, wake bool) {
	checkInput(force.IsValid(), "b2IsValidVec2( force )")

	world := b.world
	if world.recording != nil {
		recWriteBodyApplyForceToCenter(world.recording, &recArgsBodyApplyForceToCenter{b, force, wake})
	}
	body := getBodyFullId(world, b)

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if wake && body.setIndex >= firstSleepingSet {
		wakeBody(world, body)
	}

	if body.setIndex == awakeSet {
		bodySim := getBodySim(world, body)
		bodySim.force = bodySim.force.Add(force)
	}
}

// Apply a torque. This affects the angular velocity without affecting the linear velocity.
// This optionally wakes the body. The torque is ignored if the body is not awake.
// bodyId: The body id
// torque: about the z-axis (out of the screen), usually in N*m.
// wake: also wake up the body
func (b Body) ApplyTorque(torque float32, wake bool) {
	checkInput(IsValidFloat(torque), "b2IsValidFloat( torque )")

	world := b.world
	if world.recording != nil {
		recWriteBodyApplyTorque(world.recording, &recArgsBodyApplyTorque{b, torque, wake})
	}
	body := getBodyFullId(world, b)

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if wake && body.setIndex >= firstSleepingSet {
		wakeBody(world, body)
	}

	if body.setIndex == awakeSet {
		bodySim := getBodySim(world, body)
		bodySim.torque += torque
	}
}

// Clear the force and torque on this body. Forces and torques are automatically cleared after each world
// step. So this only needs to be called if the application wants to remove the effect of previous
// calls to apply forces and torques before the world step is called.
// bodyId: The body id
func (b Body) ClearForces() {
	world := b.world
	if world.recording != nil {
		recWriteBodyClearForces(world.recording, &recArgsBodyClearForces{b})
	}
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	bodySim.force = Vec2{}
	bodySim.torque = 0.0
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
	checkInput(impulse.IsValid(), "b2IsValidVec2( impulse )")
	checkInput(point.IsValid(), "b2IsValidPosition( point )")

	world := b.world
	if world.recording != nil {
		recWriteBodyApplyLinearImpulse(world.recording, &recArgsBodyApplyLinearImpulse{b, impulse, point, wake})
	}
	body := getBodyFullId(world, b)

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if wake && body.setIndex >= firstSleepingSet {
		wakeBody(world, body)
	}

	if body.setIndex == awakeSet {
		localIndex := body.localIndex
		set := &world.solverSets[awakeSet]
		state := &set.bodyStates[localIndex]
		bodySim := &set.bodySims[localIndex]
		state.linearVelocity = state.linearVelocity.MulAdd(bodySim.invMass, impulse)
		state.angularVelocity += float32(bodySim.invInertia * point.Sub(bodySim.center).Cross(impulse))

		limitVelocity(state, world.maxLinearSpeed)
	}
}

// Apply an impulse to the center of mass. This immediately modifies the velocity.
// The impulse is ignored if the body is not awake. This optionally wakes the body.
// bodyId: The body id
// impulse: the world impulse vector, usually in N*s or kg*m/s.
// wake: also wake up the body
// Warning: This should be used for one-shot impulses. If you need a steady force,
// use a force instead, which will work better with the sub-stepping solver.
func (b Body) ApplyLinearImpulseToCenter(impulse Vec2, wake bool) {
	checkInput(impulse.IsValid(), "b2IsValidVec2( impulse )")

	world := b.world
	if world.recording != nil {
		recWriteBodyApplyLinearImpulseToCenter(world.recording, &recArgsBodyApplyLinearImpulseToCenter{b, impulse, wake})
	}
	body := getBodyFullId(world, b)

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if wake && body.setIndex >= firstSleepingSet {
		wakeBody(world, body)
	}

	if body.setIndex == awakeSet {
		localIndex := body.localIndex
		set := &world.solverSets[awakeSet]
		state := &set.bodyStates[localIndex]
		bodySim := &set.bodySims[localIndex]
		state.linearVelocity = state.linearVelocity.MulAdd(bodySim.invMass, impulse)

		limitVelocity(state, world.maxLinearSpeed)
	}
}

// Apply an angular impulse. The impulse is ignored if the body is not awake.
// This optionally wakes the body.
// bodyId: The body id
// impulse: the angular impulse, usually in units of kg*m*m/s
// wake: also wake up the body
// Warning: This should be used for one-shot impulses. If you need a steady torque,
// use a torque instead, which will work better with the sub-stepping solver.
func (b Body) ApplyAngularImpulse(impulse float32, wake bool) {
	checkInput(IsValidFloat(impulse), "b2IsValidFloat( impulse )")
	assert(b.IsValid(), "b2Body_IsValid( bodyId )")
	world := b.world
	if world.recording != nil {
		recWriteBodyApplyAngularImpulse(world.recording, &recArgsBodyApplyAngularImpulse{b, impulse, wake})
	}
	body := getBodyFullId(world, b)

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if wake && body.setIndex >= firstSleepingSet {
		// this will not invalidate body pointer
		wakeBody(world, body)
	}

	if body.setIndex == awakeSet {
		localIndex := body.localIndex
		set := &world.solverSets[awakeSet]
		state := &set.bodyStates[localIndex]
		bodySim := &set.bodySims[localIndex]
		state.angularVelocity += float32(bodySim.invInertia * impulse)
	}
}

// Get the body type: static, kinematic, or dynamic
func (b Body) Type() BodyType {
	world := b.world
	body := getBodyFullId(world, b)
	return body.bodyType
}

// Change the body type. This is an expensive operation. This automatically updates the mass
// properties regardless of the automatic mass setting.
func (b Body) SetType(bodyType BodyType) {
	world := b.world
	if world.recording != nil {
		recWriteBodySetType(world.recording, &recArgsBodySetType{b, int32(bodyType)})
	}
	body := getBodyFullId(world, b)

	originalType := body.bodyType
	if originalType == bodyType {
		return
	}

	// Stage 1: skip disabled bodies
	if body.setIndex == disabledSet {
		// Disabled bodies don't change solver sets or islands when they change type.
		body.bodyType = bodyType

		if bodyType == DynamicBody {
			body.flags |= dynamicFlag
		} else {
			body.flags &= ^dynamicFlag
		}

		syncBodyFlags(world, body)

		// Body type affects the mass properties
		updateBodyMassData(world, body)
		return
	}

	// Stage 2: destroy all contacts
	destroyBodyContacts(world, body)

	// Stage 3: wake this body (does nothing if body is static), otherwise it will also wake
	// all bodies in the same sleeping solver set.
	wakeBody(world, body)

	// Stage 4: move joints to temporary storage
	staticSolverSet := &world.solverSets[staticSet]

	jointKey := body.headJointKey
	for jointKey != nullIndex {
		jointID := jointKey >> 1
		edgeIndex := jointKey & 1

		joint := &world.joints[jointID]
		jointKey = joint.edges[edgeIndex].nextKey

		// Joint may be disabled by other body
		if joint.setIndex == disabledSet {
			continue
		}

		// Wake attached bodies. The wakeBody call above does not wake bodies
		// attached to a static body. But it is necessary because the body may have
		// no joints.
		bodyA := &world.bodies[joint.edges[0].bodyId]
		bodyB := &world.bodies[joint.edges[1].bodyId]
		wakeBody(world, bodyA)
		wakeBody(world, bodyB)

		// Remove joint from island
		unlinkJoint(world, joint)

		// It is necessary to transfer all joints to the static set
		// so they can be added to the constraint graph below and acquire consistent colors.
		jointSourceSet := &world.solverSets[joint.setIndex]
		transferJoint(world, staticSolverSet, jointSourceSet, joint)
	}

	// Stage 5: change the body type and transfer body
	body.bodyType = bodyType

	if bodyType == DynamicBody {
		body.flags |= dynamicFlag
	} else {
		body.flags &= ^dynamicFlag
	}

	awakeSolverSet := &world.solverSets[awakeSet]
	sourceSet := &world.solverSets[body.setIndex]
	targetSet := awakeSolverSet
	if bodyType == StaticBody {
		targetSet = staticSolverSet
	}

	// Transfer body
	transferBody(world, targetSet, sourceSet, body)

	// Stage 6: update island participation for the body
	if originalType == StaticBody {
		// Create island for body
		createIslandForBody(world, awakeSet, body)
	} else if bodyType == StaticBody {
		// Remove body from island.
		removeBodyFromIsland(world, body)
	}

	// Stage 7: Transfer joints to the target set
	jointKey = body.headJointKey
	for jointKey != nullIndex {
		jointID := jointKey >> 1
		edgeIndex := jointKey & 1

		joint := &world.joints[jointID]

		jointKey = joint.edges[edgeIndex].nextKey

		// Joint may be disabled by other body
		if joint.setIndex == disabledSet {
			continue
		}

		// All joints were transferred to the static set in an earlier stage
		assert(joint.setIndex == staticSet, "joint->setIndex == b2_staticSet")

		bodyA := &world.bodies[joint.edges[0].bodyId]
		bodyB := &world.bodies[joint.edges[1].bodyId]
		assert(bodyA.setIndex == staticSet || bodyA.setIndex == awakeSet, "bodyA->setIndex == b2_staticSet || bodyA->setIndex == b2_awakeSet")
		assert(bodyB.setIndex == staticSet || bodyB.setIndex == awakeSet, "bodyB->setIndex == b2_staticSet || bodyB->setIndex == b2_awakeSet")

		if bodyA.bodyType == DynamicBody || bodyB.bodyType == DynamicBody {
			transferJoint(world, awakeSolverSet, staticSolverSet, joint)
		}
	}

	// Recreate shape proxies in broadphase
	transform := getBodyTransformQuick(world, body)
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]
		shapeID = shape.nextShapeID
		destroyShapeProxy(shape, &world.broadPhase)
		forcePairCreation := true
		createShapeProxy(world, shape, bodyType, transform, forcePairCreation)
	}

	// Relink all joints
	jointKey = body.headJointKey
	for jointKey != nullIndex {
		jointID := jointKey >> 1
		edgeIndex := jointKey & 1

		joint := &world.joints[jointID]
		jointKey = joint.edges[edgeIndex].nextKey

		otherEdgeIndex := edgeIndex ^ 1
		otherBodyID := joint.edges[otherEdgeIndex].bodyId
		otherBody := &world.bodies[otherBodyID]

		if otherBody.setIndex == disabledSet {
			continue
		}

		if body.bodyType != DynamicBody && otherBody.bodyType != DynamicBody {
			continue
		}

		linkJoint(world, joint)
	}

	syncBodyFlags(world, body)

	// Body type affects the mass
	updateBodyMassData(world, body)

	validateSolverSets(world)
	validateIsland(world, body.islandID)
}

// Set the body name. Up to 31 characters excluding 0 termination.
func (b Body) SetName(name string) {
	world := b.world
	if world.recording != nil {
		recWriteBodySetName(world.recording, &recArgsBodySetName{b, name})
	}
	body := getBodyFullId(world, b)

	clear(body.name[:])
	if end := strings.IndexByte(name, 0); end >= 0 {
		name = name[:end]
	}
	copy(body.name[:NameLength], name)
}

// Get the body name.
func (b Body) Name() string {
	world := b.world
	body := getBodyFullId(world, b)
	return string(body.name[:strings.IndexByte(string(body.name[:]), 0)])
}

// Set the user data for a body
func (b Body) SetUserData(userData any) {
	world := b.world
	body := getBodyFullId(world, b)
	body.userData = userData
}

// Get the user data stored in a body
func (b Body) UserData() any {
	world := b.world
	body := getBodyFullId(world, b)
	return body.userData
}

// Get the mass of the body, usually in kilograms
func (b Body) Mass() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	return body.mass
}

// Get the rotational inertia of the body, usually in kg*m^2
func (b Body) RotationalInertia() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	return body.inertia
}

// Get the center of mass position of the body in local space.
func (b Body) LocalCenter() Vec2 {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.localCenter
}

// Get the center of mass position of the body in world space.
func (b Body) WorldCenter() Pos {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.center
}

// Override the body's mass properties. Normally this is computed automatically using the
// shape geometry and density. This information is lost if a shape is added or removed or if the
// body type changes.
// If motion locks are used to fix rotation then the rotational inertia value is ignored and
// rotational inertia remains fixed at zero.
func (b Body) SetMassData(massData MassData) {
	checkInput(IsValidFloat(massData.Mass) && massData.Mass >= 0.0, "b2IsValidFloat( massData.mass ) && massData.mass >= 0.0f")
	checkInput(IsValidFloat(massData.RotationalInertia) && massData.RotationalInertia >= 0.0, "b2IsValidFloat( massData.rotationalInertia ) && massData.rotationalInertia >= 0.0f")
	checkInput(massData.Center.IsValid(), "b2IsValidVec2( massData.center )")

	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetMassData(world.recording, &recArgsBodySetMassData{b, massData})
	}

	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)

	body.mass = massData.Mass
	body.inertia = massData.RotationalInertia
	bodySim.localCenter = massData.Center

	oldCenter := bodySim.center
	center := bodySim.transform.Apply(massData.Center)
	bodySim.center = center
	bodySim.center0 = center

	// Update center of mass velocity
	state := getBodyState(world, body)
	if state != nil {
		deltaLinear := bodySim.center.Sub(oldCenter).ScalarCross(state.angularVelocity)
		state.linearVelocity = state.linearVelocity.Add(deltaLinear)
	}

	if body.mass > 0.0 {
		bodySim.invMass = 1.0 / body.mass
	} else {
		bodySim.invMass = 0.0
	}
	if body.inertia > 0.0 {
		bodySim.invInertia = 1.0 / body.inertia
	} else {
		bodySim.invInertia = 0.0
	}

	// Update extents using supplied mass center.
	bodySim.minExtent = Huge()
	bodySim.maxExtent = 0.0
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		s := &world.shapes[shapeID]
		extent := computeShapeExtent(s, massData.Center)
		bodySim.minExtent = minFloat(bodySim.minExtent, extent.minExtent)
		bodySim.maxExtent = maxFloat(bodySim.maxExtent, extent.maxExtent)
		shapeID = s.nextShapeID
	}

	// When the center of mass changes, any cached contact manifold becomes invalid.
	edgeKey := body.headContactKey
	for edgeKey != nullIndex {
		contactID := edgeKey >> 1
		edgeIndex := edgeKey & 1

		contact := &world.contacts[contactID]
		contactSim := getContactSim(world, contact)
		contactSim.simFlags &= ^simRelativeTransformValid

		edgeKey = contact.edges[edgeIndex].nextKey
	}

	// Motion locks take priority over mass data.
	if body.flags&fixedRotation != 0 {
		body.inertia = 0.0
		bodySim.invInertia = 0.0
	}

	body.flags &= ^dirtyMass
}

// Get the mass data for a body
func (b Body) MassData() MassData {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	massData := MassData{body.mass, bodySim.localCenter, body.inertia}
	return massData
}

// This updates the mass properties to the sum of the mass properties of the shapes.
// This normally does not need to be called unless you called SetMassData to override
// the mass and you later want to reset the mass.
// You may also use this when automatic mass computation has been disabled.
// You should call this regardless of body type.
// Note that sensor shapes may have mass.
func (b Body) UpdateMassFromShapes() {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodyUpdateMassFromShapes(world.recording, &recArgsBodyUpdateMassFromShapes{b})
	}

	body := getBodyFullId(world, b)
	updateBodyMassData(world, body)
}

// Adjust the linear damping. Normally this is set in BodyDef before creation.
func (b Body) SetLinearDamping(linearDamping float32) {
	checkInput(IsValidFloat(linearDamping) && linearDamping >= 0.0, "b2IsValidFloat( linearDamping ) && linearDamping >= 0.0f")

	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetLinearDamping(world.recording, &recArgsBodySetLinearDamping{b, linearDamping})
	}

	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	bodySim.linearDamping = linearDamping
}

// Get the current linear damping.
func (b Body) LinearDamping() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.linearDamping
}

// Adjust the angular damping. Normally this is set in BodyDef before creation.
func (b Body) SetAngularDamping(angularDamping float32) {
	checkInput(IsValidFloat(angularDamping) && angularDamping >= 0.0, "b2IsValidFloat( angularDamping ) && angularDamping >= 0.0f")

	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetAngularDamping(world.recording, &recArgsBodySetAngularDamping{b, angularDamping})
	}

	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	bodySim.angularDamping = angularDamping
}

// Get the current angular damping.
func (b Body) AngularDamping() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.angularDamping
}

// Adjust the gravity scale. Normally this is set in BodyDef before creation.
// See: BodyDef.GravityScale
func (b Body) SetGravityScale(gravityScale float32) {
	assert(b.IsValid(), "b2Body_IsValid( bodyId )")
	checkInput(IsValidFloat(gravityScale), "b2IsValidFloat( gravityScale )")

	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetGravityScale(world.recording, &recArgsBodySetGravityScale{b, gravityScale})
	}

	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	bodySim.gravityScale = gravityScale
}

// Get the current gravity scale
func (b Body) GravityScale() float32 {
	assert(b.IsValid(), "b2Body_IsValid( bodyId )")
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.gravityScale
}

// Returns: true if this body is awake
func (b Body) Awake() bool {
	world := b.world
	body := getBodyFullId(world, b)
	return body.setIndex == awakeSet
}

// Wake a body from sleep. This wakes the entire island the body is touching.
// Warning: Putting a body to sleep will put the entire island of bodies touching this body to sleep,
// which can be expensive and possibly unintuitive.
func (b Body) SetAwake(awake bool) {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetAwake(world.recording, &recArgsBodySetAwake{b, awake})
	}

	body := getBodyFullId(world, b)

	if awake && body.setIndex >= firstSleepingSet {
		wakeBody(world, body)
	} else if awake == false && body.setIndex == awakeSet {
		island := &world.islands[body.islandID]
		if island.constraintRemoveCount > 0 {
			// Must split the island before sleeping. This is expensive.
			splitIsland(world, body.islandID)
		}

		trySleepIsland(world, body.islandID)
	}
}

// Wake bodies touching this body. Works for static bodies.
func (b Body) WakeTouching() {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodyWakeTouching(world.recording, &recArgsBodyWakeTouching{b})
	}
	body := getBodyFullId(world, b)

	contactKey := body.headContactKey
	for contactKey != nullIndex {
		contactID := contactKey >> 1
		edgeIndex := contactKey & 1

		contact := &world.contacts[contactID]
		shapeA := &world.shapes[contact.shapeIDA]
		shapeB := &world.shapes[contact.shapeIDB]

		if shapeA.bodyID == int(b.index1)-1 {
			otherBody := &world.bodies[shapeB.bodyID]
			wakeBody(world, otherBody)
		} else {
			otherBody := &world.bodies[shapeA.bodyID]
			wakeBody(world, otherBody)
		}

		contactKey = contact.edges[edgeIndex].nextKey
	}
}

// Returns true if this body is enabled
func (b Body) Enabled() bool {
	world := b.world
	body := getBodyFullId(world, b)
	return body.setIndex != disabledSet
}

// Returns true if sleeping is enabled for this body
func (b Body) SleepEnabled() bool {
	world := b.world
	body := getBodyFullId(world, b)
	return (body.flags & enableSleep) == enableSleep
}

// Set the sleep threshold, usually in meters per second
func (b Body) SetSleepThreshold(sleepThreshold float32) {
	checkInput(IsValidFloat(sleepThreshold) && sleepThreshold >= 0.0, "b2IsValidFloat( sleepThreshold ) && sleepThreshold >= 0.0f")

	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetSleepThreshold(world.recording, &recArgsBodySetSleepThreshold{b, sleepThreshold})
	}
	body := getBodyFullId(world, b)
	body.sleepThreshold = sleepThreshold
}

// Get the sleep threshold, usually in meters per second.
func (b Body) SleepThreshold() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	return body.sleepThreshold
}

// Set the continuous collision safety factor. Smaller is safer but can lead to hitching. Recommended range [0.01, 0.5]. Non-dimensional.
func (b Body) SetSafetyFactor(safetyFactor float32) {
	checkInput(IsValidFloat(safetyFactor) && safetyFactor >= 0.0, "b2IsValidFloat( safetyFactor ) && safetyFactor >= 0.0f")
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetSafetyFactor(world.recording, &recArgsBodySetSafetyFactor{b, safetyFactor})
	}
	body := getBodyFullId(world, b)
	body.safetyFactor = safetyFactor
}

// Get the continuous collision safety factor. Non-dimensional.
func (b Body) SafetyFactor() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	return body.safetyFactor
}

// Enable or disable sleeping for this body. If sleeping is disabled the body will wake (and the entire island).
func (b Body) SetSleepEnabled(enable bool) {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodyEnableSleep(world.recording, &recArgsBodyEnableSleep{b, enable})
	}

	body := getBodyFullId(world, b)

	flag := (body.flags & enableSleep) == enableSleep
	if enable == flag {
		return
	}

	if enable {
		body.flags = body.flags | enableSleep
	} else {
		body.flags = body.flags & ^enableSleep
	}
	syncBodyFlags(world, body)

	if enable == false {
		wakeBody(world, body)
	}
}

// Disable a body by removing it completely from the simulation. This is expensive.
func (b Body) Disable() {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodyDisable(world.recording, &recArgsBodyDisable{b})
	}

	body := getBodyFullId(world, b)
	if body.setIndex == disabledSet {
		return
	}

	// Destroy contacts and wake bodies touching this body. This avoid floating bodies.
	// This is necessary even for static bodies.
	destroyBodyContacts(world, body)

	// The current solver set of the body
	set := &world.solverSets[body.setIndex]

	// Disabled bodies and connected joints are moved to the disabled set
	disabledSolverSet := &world.solverSets[disabledSet]

	// Unlink joints and transfer them to the disabled set
	jointKey := body.headJointKey
	for jointKey != nullIndex {
		jointID := jointKey >> 1
		edgeIndex := jointKey & 1

		joint := &world.joints[jointID]
		jointKey = joint.edges[edgeIndex].nextKey

		// joint may already be disabled by other body
		if joint.setIndex == disabledSet {
			continue
		}

		assert(joint.setIndex == set.setIndex || set.setIndex == staticSet || joint.setIndex == staticSet, "joint->setIndex == set->setIndex || set->setIndex == b2_staticSet || joint->setIndex == b2_staticSet")

		// Remove joint from island
		unlinkJoint(world, joint)

		// Transfer joint to disabled set
		jointSet := &world.solverSets[joint.setIndex]
		transferJoint(world, disabledSolverSet, jointSet, joint)
	}

	// Remove shapes from broad-phase
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]
		shapeID = shape.nextShapeID
		destroyShapeProxy(shape, &world.broadPhase)
	}

	// Disabled bodies are not in an island. If the island becomes empty it will be destroyed.
	removeBodyFromIsland(world, body)

	// Transfer body sim
	transferBody(world, disabledSolverSet, set, body)

	validateConnectivity(world)
	validateSolverSets(world)
}

// Enable a body by adding it to the simulation. This is expensive.
func (b Body) Enable() {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodyEnable(world.recording, &recArgsBodyEnable{b})
	}

	body := getBodyFullId(world, b)
	if body.setIndex != disabledSet {
		return
	}

	disabledSolverSet := &world.solverSets[disabledSet]
	setID := awakeSet
	if body.bodyType == StaticBody {
		setID = staticSet
	}
	targetSet := &world.solverSets[setID]

	transferBody(world, targetSet, disabledSolverSet, body)

	transform := getBodyTransformQuick(world, body)

	// Add shapes to broad-phase
	proxyType := body.bodyType
	forcePairCreation := true
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]
		shapeID = shape.nextShapeID

		createShapeProxy(world, shape, proxyType, transform, forcePairCreation)
	}

	if setID != staticSet {
		createIslandForBody(world, setID, body)
	}

	// Transfer joints. If the other body is disabled, don't transfer.
	// If the other body is sleeping, wake it.
	jointKey := body.headJointKey
	for jointKey != nullIndex {
		jointID := jointKey >> 1
		edgeIndex := jointKey & 1

		joint := &world.joints[jointID]
		assert(joint.setIndex == disabledSet, "joint->setIndex == b2_disabledSet")
		assert(joint.islandId == nullIndex, "joint->islandId == B2_NULL_INDEX")

		jointKey = joint.edges[edgeIndex].nextKey

		bodyA := &world.bodies[joint.edges[0].bodyId]
		bodyB := &world.bodies[joint.edges[1].bodyId]

		if bodyA.setIndex == disabledSet || bodyB.setIndex == disabledSet {
			// one body is still disabled
			continue
		}

		// Transfer joint first
		var jointSetID int
		if bodyA.bodyType != DynamicBody && bodyB.bodyType != DynamicBody {
			jointSetID = staticSet
		} else if bodyA.setIndex == staticSet {
			jointSetID = bodyB.setIndex
		} else {
			jointSetID = bodyA.setIndex
		}

		jointSet := &world.solverSets[jointSetID]
		transferJoint(world, jointSet, disabledSolverSet, joint)

		// Now that the joint is in the correct set, I can link the joint in the island.
		if jointSetID != staticSet {
			linkJoint(world, joint)
		}
	}

	validateSolverSets(world)
}

// Set the motion locks on this body. This causes mass properties to be recomputed because
// fixed rotation sets the rotational inertia to zero.
func (b Body) SetMotionLocks(locks MotionLocks) {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetMotionLocks(world.recording, &recArgsBodySetMotionLocks{b, locks})
	}

	var newFlags uint32 = 0
	if locks.LinearX {
		newFlags |= lockLinearX
	}
	if locks.LinearY {
		newFlags |= lockLinearY
	}
	if locks.AngularZ {
		newFlags |= lockAngularZ
	}

	body := getBodyFullId(world, b)
	if (body.flags & allLocks) != newFlags {
		body.flags &= ^allLocks
		body.flags |= newFlags

		syncBodyFlags(world, body)

		state := getBodyState(world, body)

		if state != nil {
			if locks.LinearX {
				state.linearVelocity.X = 0.0
			}

			if locks.LinearY {
				state.linearVelocity.Y = 0.0
			}

			if locks.AngularZ {
				state.angularVelocity = 0.0
			}
		}

		// Motion locks can affect mass properties.
		updateBodyMassData(world, body)
	}
}

// Get the motion locks for this body.
func (b Body) MotionLocks() MotionLocks {
	world := b.world
	body := getBodyFullId(world, b)

	var locks MotionLocks
	locks.LinearX = (body.flags & lockLinearX) != 0
	locks.LinearY = (body.flags & lockLinearY) != 0
	locks.AngularZ = (body.flags & lockAngularZ) != 0
	return locks
}

// Set this body to be a bullet. A bullet does continuous collision detection
// against dynamic bodies (but not other bullets).
func (b Body) SetBullet(flag bool) {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodySetBullet(world.recording, &recArgsBodySetBullet{b, flag})
	}

	var newFlag uint32
	if flag {
		newFlag = isBullet
	} else {
		newFlag = 0
	}

	body := getBodyFullId(world, b)
	if (body.flags & isBullet) == newFlag {
		return
	}

	body.flags &= ^isBullet
	body.flags |= newFlag

	syncBodyFlags(world, body)
}

// Is this body a bullet?
func (b Body) IsBullet() bool {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return (bodySim.flags & isBullet) != 0
}

// Enable or disable contact recycling for this body. Contact recycling is a performance optimization
// that reuses contact manifolds when bodies move slightly. Disabling it can avoid ghost collisions
// on characters at the cost of higher per-step work. Existing contacts retain their prior setting;
// only contacts created after this call see the new value.
// See: BodyDef.EnableContactRecycling
func (b Body) SetContactRecyclingEnabled(flag bool) {
	world := getWorldLocked(b.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteBodyEnableContactRecycling(world.recording, &recArgsBodyEnableContactRecycling{b, flag})
	}

	var newFlag uint32
	if flag {
		newFlag = bodyEnableContactRecycling
	} else {
		newFlag = 0
	}

	body := getBodyFullId(world, b)
	if (body.flags & bodyEnableContactRecycling) == newFlag {
		return
	}

	body.flags &= ^bodyEnableContactRecycling
	body.flags |= newFlag

	syncBodyFlags(world, body)
}

// Is contact recycling enabled on this body?
func (b Body) ContactRecyclingEnabled() bool {
	world := b.world
	body := getBodyFullId(world, b)
	return (body.flags & bodyEnableContactRecycling) != 0
}

// Enable/disable contact events on all shapes.
// See: ShapeDef.EnableContactEvents
// Warning: changing this at runtime may cause mismatched begin/end touch events
func (b Body) SetContactEventsEnabled(flag bool) {
	world := b.world
	if world.recording != nil {
		recWriteBodyEnableContactEvents(world.recording, &recArgsBodyEnableContactEvents{b, flag})
	}
	body := getBodyFullId(world, b)
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]
		shape.enableContactEvents = flag
		shapeID = shape.nextShapeID
	}
}

// Enable/disable hit events on all shapes
// See: ShapeDef.EnableHitEvents
func (b Body) SetHitEventsEnabled(flag bool) {
	world := b.world
	if world.recording != nil {
		recWriteBodyEnableHitEvents(world.recording, &recArgsBodyEnableHitEvents{b, flag})
	}
	body := getBodyFullId(world, b)
	shapeID := body.headShapeID
	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]
		shape.enableHitEvents = flag
		shapeID = shape.nextShapeID
	}
}

// Get the world that owns this body
func (b Body) World() *World {
	return b.world
}

// Get the number of shapes on this body
func (b Body) ShapeCount() int {
	world := b.world
	body := getBodyFullId(world, b)
	return body.shapeCount
}

// Get the shape ids for all shapes on this body.
func (b Body) Shapes() iter.Seq[Shape] {
	world := b.world
	body := getBodyFullId(world, b)
	return func(yield func(Shape) bool) {
		shapeID := body.headShapeID
		for shapeID != nullIndex {
			shape := &world.shapes[shapeID]
			id := Shape{int32(shape.id + 1), b.world, shape.generation}
			if !yield(id) {
				return
			}
			shapeID = shape.nextShapeID
		}
	}
}

// Get the number of joints on this body
func (b Body) JointCount() int {
	world := b.world
	body := getBodyFullId(world, b)
	return body.jointCount
}

// Get the joint ids for all joints on this body
func (b Body) Joints() iter.Seq[Joint] {
	world := b.world
	body := getBodyFullId(world, b)
	return func(yield func(Joint) bool) {
		jointKey := body.headJointKey
		for jointKey != nullIndex {
			jointID := jointKey >> 1
			edgeIndex := jointKey & 1
			joint := &world.joints[jointID]
			id := Joint{int32(jointID + 1), b.world, joint.generation}
			if !yield(id) {
				return
			}
			jointKey = joint.edges[edgeIndex].nextKey
		}
	}
}

func shouldBodiesCollide(world *World, bodyA *body, bodyB *body) bool {
	if bodyA.bodyType != DynamicBody && bodyB.bodyType != DynamicBody {
		return false
	}

	var jointKey int
	var otherBodyID int
	if bodyA.jointCount < bodyB.jointCount {
		jointKey = bodyA.headJointKey
		otherBodyID = bodyB.id
	} else {
		jointKey = bodyB.headJointKey
		otherBodyID = bodyA.id
	}

	for jointKey != nullIndex {
		jointID := jointKey >> 1
		edgeIndex := jointKey & 1
		otherEdgeIndex := edgeIndex ^ 1

		joint := &world.joints[jointID]
		if joint.collideConnected == false && joint.edges[otherEdgeIndex].bodyId == otherBodyID {
			return false
		}

		jointKey = joint.edges[edgeIndex].nextKey
	}

	return true
}

// The minimum distance from any point on the body shapes to the center of mass.
func (b Body) MinExtent() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.minExtent
}

// The maximum distance from any point on the body shapes to the center of mass.
func (b Body) MaxExtent() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.maxExtent
}

// The maximum distance from any point on the body shapes to the body origin. Conservative.
func (b Body) MaxExtentOrigin() float32 {
	world := b.world
	body := getBodyFullId(world, b)
	bodySim := getBodySim(world, body)
	return bodySim.maxExtent + bodySim.localCenter.Length()
}

func makeRelativeSweep(bodySim *bodySim, base Pos) Sweep {
	var s Sweep
	s.C1 = bodySim.center0.Sub(base)
	s.C2 = bodySim.center.Sub(base)
	s.Q1 = bodySim.rotation0
	s.Q2 = bodySim.transform.Q
	s.LocalCenter = bodySim.localCenter
	return s
}

func getBodySim(world *World, body *body) *bodySim {
	set := &world.solverSets[body.setIndex]
	bodySim := &set.bodySims[body.localIndex]
	return bodySim
}

func getBodyState(world *World, body *body) *bodyState {
	if body.setIndex == awakeSet {
		set := &world.solverSets[awakeSet]
		return &set.bodyStates[body.localIndex]
	}

	return nil
}

func encodeBodySimIndex(body *body) int {
	if body.setIndex == awakeSet {
		return body.localIndex
	}

	if body.setIndex == staticSet {
		return -(body.localIndex + 2)
	}

	return nullIndex
}

func isStaticSimIndex(encodedBodySimIndex int) bool {
	return encodedBodySimIndex < nullIndex
}

func decodeAwakeIndex(encodedBodySimIndex int) int {
	if encodedBodySimIndex >= 0 {
		return encodedBodySimIndex
	}
	return nullIndex
}

func sleepBodySimIndex(encodedIndex int) int {
	if encodedIndex >= 0 {
		return nullIndex
	}
	return encodedIndex
}
