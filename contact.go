// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

const (
	contactTouchingFlag        uint32 = 0x00000001
	contactHitEventFlag        uint32 = 0x00000002
	contactEnableContactEvents uint32 = 0x00000004
	contactRecycleFlag         uint32 = 0x00000008
	simTouchingFlag            uint32 = 0x00010000
	simDisjoint                uint32 = 0x00020000
	simStartedTouching         uint32 = 0x00040000
	simStoppedTouching         uint32 = 0x00080000
	simEnableHitEvent          uint32 = 0x00100000
	simEnablePreSolveEvents    uint32 = 0x00200000
	simRelativeTransformValid  uint32 = 0x00400000
)

var contactRegistry = [ShapeTypeCount][ShapeTypeCount]contactRegister{
	CircleShape: {
		CircleShape:       {fcn: circleManifold, primary: true},
		CapsuleShape:      {fcn: capsuleAndCircleManifold, primary: false},
		SegmentShape:      {fcn: segmentAndCircleManifold, primary: false},
		PolygonShape:      {fcn: polygonAndCircleManifold, primary: false},
		ChainSegmentShape: {fcn: chainSegmentAndCircleManifold, primary: false},
	},
	CapsuleShape: {
		CircleShape:       {fcn: capsuleAndCircleManifold, primary: true},
		CapsuleShape:      {fcn: capsuleManifold, primary: true},
		SegmentShape:      {fcn: segmentAndCapsuleManifold, primary: false},
		PolygonShape:      {fcn: polygonAndCapsuleManifold, primary: false},
		ChainSegmentShape: {fcn: chainSegmentAndCapsuleManifold, primary: false},
	},
	SegmentShape: {
		CircleShape:  {fcn: segmentAndCircleManifold, primary: true},
		CapsuleShape: {fcn: segmentAndCapsuleManifold, primary: true},
		PolygonShape: {fcn: segmentAndPolygonManifold, primary: true},
	},
	PolygonShape: {
		CircleShape:       {fcn: polygonAndCircleManifold, primary: true},
		CapsuleShape:      {fcn: polygonAndCapsuleManifold, primary: true},
		SegmentShape:      {fcn: segmentAndPolygonManifold, primary: false},
		PolygonShape:      {fcn: polygonManifold, primary: true},
		ChainSegmentShape: {fcn: chainSegmentAndPolygonManifold, primary: false},
	},
	ChainSegmentShape: {
		CircleShape:  {fcn: chainSegmentAndCircleManifold, primary: true},
		CapsuleShape: {fcn: chainSegmentAndCapsuleManifold, primary: true},
		PolygonShape: {fcn: chainSegmentAndPolygonManifold, primary: true},
	},
}

// A begin touch event is generated when two shapes begin touching.
type ContactBeginTouchEvent struct {
	// Id of the first shape
	ShapeA Shape
	// Id of the second shape
	ShapeB Shape
	// The transient contact id. This contact maybe destroyed automatically when the world is modified or simulated.
	// Used Contact.IsValid before using this id.
	Contact Contact
}

// An end touch event is generated when two shapes stop touching.
//
//	You will get an end event if you do anything that destroys contacts previous to the last
//	world step. These include things like setting the transform, destroying a body
//	or shape, or changing a filter or body type.
type ContactEndTouchEvent struct {
	// Id of the first shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	ShapeA Shape
	// Id of the second shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	ShapeB Shape
	// Id of the contact.
	// 	Warning: this contact may have been destroyed
	// 	See: Contact.IsValid
	Contact Contact
}

// A hit touch event is generated when two shapes collide with a speed faster than the hit speed threshold.
// This may be reported for speculative contacts that have a confirmed impulse.
type ContactHitEvent struct {
	// Id of the first shape
	ShapeA Shape
	// Id of the second shape
	ShapeB Shape
	// Id of the contact.
	// 	Warning: this contact may have been destroyed
	// 	See: Contact.IsValid
	Contact Contact
	// Point where the shapes hit. This is a mid-point between the two surfaces. It could be
	// a speculative point where the shapes were not touching at the beginning of the time
	// step. The point is reconstructed after the step, so when both bodies are moving it may
	// trail the impact by up to one step of motion.
	Point Pos
	// Normal vector pointing from shape A to shape B
	Normal Vec2
	// The speed the shapes are approaching. Always positive. Typically in meters per second.
	ApproachSpeed float32
}

// Contact events are buffered in the Box2D world and are available
// as event arrays after the time step is complete.
// Note: these may become invalid if bodies and/or shapes are destroyed.
type ContactEvents struct {
	// Array of begin touch events
	BeginEvents []ContactBeginTouchEvent
	// Array of end touch events
	EndEvents []ContactEndTouchEvent
	// Array of hit events
	HitEvents []ContactHitEvent
	// Number of begin touch events
	BeginCount int
	// Number of end touch events
	EndCount int
	// Number of hit events
	HitCount int
}

// The contact data for two shapes. By convention the manifold normal points
// from shape A to shape B.
// See: Shape.ContactData() and Body.ContactData().
type ContactData struct {
	// The contact id. You may hold onto this to track a contact across time steps.
	// This id may become orphaned. Use Contact.IsValid before using it for other functions.
	Contact Contact
	// The first shape id.
	ShapeA Shape
	// The second shape id.
	ShapeB Shape
	// The manifold copied from the contact.
	Manifold Manifold
}

type contactEdge struct {
	bodyID  int
	prevKey int
	nextKey int
}

type contact struct {
	edges       [2]contactEdge
	islandID    int
	islandIndex int
	setIndex    int
	colorIndex  int
	localIndex  int
	shapeIDA    int
	shapeIDB    int
	contactID   int
	flags       uint32
	generation  uint32
}

type contactSim struct {
	contactID          int
	cachedRotationA    Rot
	cachedRotationB    Rot
	cachedRelativePose Transform
	bodyIDA            int
	bodyIDB            int
	encodedBodySimA    int
	encodedBodySimB    int
	simFlags           uint32
	shapeIDA           int
	shapeIDB           int
	invMassA           float32
	invIA              float32
	invMassB           float32
	invIB              float32
	manifold           Manifold
	friction           float32
	restitution        float32
	rollingResistance  float32
	tangentSpeed       float32
	cache              SimplexCache
}

type contactRegister struct {
	fcn     manifoldFunc
	primary bool
}

type manifoldFunc func(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold

func circleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideCircles(&shapeA.circle, &shapeB.circle, xf)
}

func capsuleAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideCapsuleAndCircle(&shapeA.capsule, &shapeB.circle, xf)
}

func segmentAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideSegmentAndCircle(&shapeA.segment, &shapeB.circle, xf)
}

func polygonAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollidePolygonAndCircle(&shapeA.polygon, &shapeB.circle, xf)
}

func chainSegmentAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideChainSegmentAndCircle(&shapeA.chainSegment, &shapeB.circle, xf)
}

func capsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideCapsules(&shapeA.capsule, &shapeB.capsule, xf)
}

func segmentAndCapsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideSegmentAndCapsule(&shapeA.segment, &shapeB.capsule, xf)
}

func polygonAndCapsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollidePolygonAndCapsule(&shapeA.polygon, &shapeB.capsule, xf)
}

func chainSegmentAndCapsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	return CollideChainSegmentAndCapsule(&shapeA.chainSegment, &shapeB.capsule, xf, cache)
}

func segmentAndPolygonManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollideSegmentAndPolygon(&shapeA.segment, &shapeB.polygon, xf)
}

func polygonManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	_ = cache
	return CollidePolygons(&shapeA.polygon, &shapeB.polygon, xf)
}

func chainSegmentAndPolygonManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	return CollideChainSegmentAndPolygon(&shapeA.chainSegment, &shapeB.polygon, xf, cache)
}

func getContactFullID(world *World, contactID Contact) *contact {
	id := int(contactID.index1) - 1
	contact := &world.contacts[id]
	assert(contact.contactID == id && contact.generation == contactID.generation, "contact.contactID == id && contact.generation == contactID.generation")

	return contact
}

// Get the world that owns this contact.
func (c Contact) World() *World {
	return c.world
}

// Get the data for a contact. The manifold may have no points if the contact is not touching.
func (c Contact) Data() ContactData {
	world := c.world
	contact := getContactFullID(world, c)
	contactSim := getContactSim(world, contact)
	shapeA := &world.shapes[contact.shapeIDA]
	shapeB := &world.shapes[contact.shapeIDB]

	return ContactData{
		Contact:  c,
		ShapeA:   Shape{index1: int32(shapeA.id + 1), world: c.world, generation: shapeA.generation},
		ShapeB:   Shape{index1: int32(shapeB.id + 1), world: c.world, generation: shapeB.generation},
		Manifold: contactSim.manifold,
	}
}

func canCollide(typeA ShapeType, typeB ShapeType) bool {
	return contactRegistry[typeA][typeB].fcn != nil
}

func initializeContactFlags(contact *contact, bodyA, bodyB *body, shapeA, shapeB *shape) {
	contact.flags = 0

	// Both bodies must enable recycling
	if (bodyA.flags&bodyEnableContactRecycling) != 0 && (bodyB.flags&bodyEnableContactRecycling) != 0 {
		contact.flags |= contactRecycleFlag
	}

	assert(shapeA.sensorIndex == nullIndex && shapeB.sensorIndex == nullIndex, "shapeA.sensorIndex == nullIndex && shapeB.sensorIndex == nullIndex")

	if shapeA.enableContactEvents || shapeB.enableContactEvents {
		contact.flags |= contactEnableContactEvents
	}
}

func mixContactMaterials(world *World, contactSim *contactSim, shapeA, shapeB *shape) {
	if world.frictionCallback == nil {
		contactSim.friction = mixFriction(shapeA.material.Friction, shapeB.material.Friction)
	} else {
		contactSim.friction = world.frictionCallback(shapeA.material.Friction, shapeA.material.UserMaterialID,
			shapeB.material.Friction, shapeB.material.UserMaterialID)
	}

	if world.restitutionCallback == nil {
		contactSim.restitution = mixRestitution(shapeA.material.Restitution, shapeB.material.Restitution)
	} else {
		contactSim.restitution = world.restitutionCallback(shapeA.material.Restitution, shapeA.material.UserMaterialID,
			shapeB.material.Restitution, shapeB.material.UserMaterialID)
	}
}

func connectContactToBody(world *World, contact *contact, body *body, bodyID, edgeIndex int) {
	contact.edges[edgeIndex].bodyID = bodyID
	contact.edges[edgeIndex].prevKey = nullIndex
	contact.edges[edgeIndex].nextKey = body.headContactKey
	key := (contact.contactID << 1) | edgeIndex

	headContactKey := body.headContactKey
	if headContactKey != nullIndex {
		headContact := &world.contacts[headContactKey>>1]
		headContact.edges[headContactKey&1].prevKey = key
	}

	body.headContactKey = key
	body.contactCount += 1
}

func initializeContactSim(world *World, contactSim *contactSim, contact *contact, shapeA, shapeB *shape) {
	bodyA := &world.bodies[shapeA.bodyID]
	bodyB := &world.bodies[shapeB.bodyID]
	shapeIDA := shapeA.id
	shapeIDB := shapeB.id

	if validationEnabled {
		contactSim.bodyIDA = shapeA.bodyID
		contactSim.bodyIDB = shapeB.bodyID
	}

	contactSim.encodedBodySimA = encodeBodySimIndex(bodyA)
	contactSim.encodedBodySimB = encodeBodySimIndex(bodyB)
	contactSim.invMassA = 0.0
	contactSim.invIA = 0.0
	contactSim.invMassB = 0.0
	contactSim.invIB = 0.0
	contactSim.shapeIDA = shapeIDA
	contactSim.shapeIDB = shapeIDB
	contactSim.cache = SimplexCache{}
	contactSim.manifold = Manifold{}

	// These get updated in the narrow phase, but these are needed for first touch
	mixContactMaterials(world, contactSim, shapeA, shapeB)

	contactSim.tangentSpeed = 0.0
	contactSim.simFlags = contact.flags

	if shapeA.enablePreSolveEvents || shapeB.enablePreSolveEvents {
		contactSim.simFlags |= simEnablePreSolveEvents
	}
}

func createContact(world *World, shapeA *shape, shapeB *shape) {
	type1 := shapeA.shapeType
	type2 := shapeB.shapeType

	assert(0 <= type1 && type1 < ShapeTypeCount, "0 <= type1 && type1 < ShapeTypeCount")
	assert(0 <= type2 && type2 < ShapeTypeCount, "0 <= type2 && type2 < ShapeTypeCount")

	if contactRegistry[type1][type2].fcn == nil {
		// For example, no segment vs segment collision
		return
	}

	if !contactRegistry[type1][type2].primary {
		// flip order
		createContact(world, shapeB, shapeA)
		return
	}

	bodyA := &world.bodies[shapeA.bodyID]
	bodyB := &world.bodies[shapeB.bodyID]

	assert(bodyA.setIndex != disabledSet && bodyB.setIndex != disabledSet, "bodyA.setIndex != disabledSet && bodyB.setIndex != disabledSet")
	assert(bodyA.setIndex != staticSet || bodyB.setIndex != staticSet, "bodyA.setIndex != staticSet || bodyB.setIndex != staticSet")

	var setIndex int
	if bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet {
		setIndex = awakeSet
	} else {
		// sleeping and non-touching contacts live in the disabled set
		// later if this set is found to be touching then the sleeping
		// islands will be linked and the contact moved to the merged island
		setIndex = disabledSet
	}

	set := &world.solverSets[setIndex]

	// Create contact key and contact
	contactID := allocID(&world.contactIDPool)
	if contactID == len(world.contacts) {
		world.contacts = arrayPush(world.contacts, contact{})
	}

	shapeIDA := shapeA.id
	shapeIDB := shapeB.id

	contact := &world.contacts[contactID]
	contact.contactID = contactID
	contact.generation += 1
	contact.setIndex = setIndex
	contact.colorIndex = nullIndex
	contact.localIndex = len(set.contactSims)
	contact.islandID = nullIndex
	contact.islandIndex = nullIndex
	contact.shapeIDA = shapeIDA
	contact.shapeIDB = shapeIDB
	initializeContactFlags(contact, bodyA, bodyB, shapeA, shapeB)

	// Connect to body A
	connectContactToBody(world, contact, bodyA, shapeA.bodyID, 0)
	// Connect to body B
	connectContactToBody(world, contact, bodyB, shapeB.bodyID, 1)

	// Add to pair set for fast lookup.
	pairKey := shapePairKey(shapeIDA, shapeIDB)
	addKey(&world.broadPhase.pairSet, pairKey)

	// Contacts are created as non-touching. Later if they are found to be touching
	// they will link islands and be moved into the constraint graph.
	var contactSim *contactSim

	set.contactSims, contactSim = arrayEmplace(set.contactSims)
	contactSim.contactID = contactID
	initializeContactSim(world, contactSim, contact, shapeA, shapeB)
}

func disconnectContactFromBody(world *World, contact *contact, body *body, edgeIndex int) {
	edge := &contact.edges[edgeIndex]
	if edge.prevKey != nullIndex {
		prevContact := &world.contacts[edge.prevKey>>1]
		prevEdge := &prevContact.edges[edge.prevKey&1]
		prevEdge.nextKey = edge.nextKey
	}

	if edge.nextKey != nullIndex {
		nextContact := &world.contacts[edge.nextKey>>1]
		nextEdge := &nextContact.edges[edge.nextKey&1]
		nextEdge.prevKey = edge.prevKey
	}

	edgeKey := (contact.contactID << 1) | edgeIndex
	if body.headContactKey == edgeKey {
		body.headContactKey = edge.nextKey
	}

	body.contactCount -= 1
}

func warmStartContactManifold(contactSim *contactSim, oldManifold *Manifold, centerOffsetA, centerOffsetB Vec2) {
	pointCount := contactSim.manifold.PointCount
	if pointCount > 0 {
		contactSim.manifold.RollingImpulse = oldManifold.RollingImpulse

		// Match old contact ids to new contact ids and copy the
		// stored impulses to warm start the solver.
		unmatchedCount := 0

		for i := range pointCount {
			mp2 := &contactSim.manifold.Points[i]

			// shift anchors to be center of mass relative
			mp2.AnchorA = mp2.AnchorA.Sub(centerOffsetA)
			mp2.AnchorB = mp2.AnchorB.Sub(centerOffsetB)

			mp2.NormalImpulse = 0.0
			mp2.TangentImpulse = 0.0
			mp2.TotalNormalImpulse = 0.0
			mp2.NormalVelocity = 0.0
			mp2.Persisted = false

			id2 := mp2.ID

			for j := range oldManifold.PointCount {
				mp1 := &oldManifold.Points[j]

				if mp1.ID == id2 {
					mp2.NormalImpulse = mp1.NormalImpulse
					mp2.TangentImpulse = mp1.TangentImpulse
					mp2.Persisted = true

					break
				}
			}

			if !mp2.Persisted {
				unmatchedCount++
			}
		}

		_ = unmatchedCount
	}
}

func destroyContact(world *World, contact *contact) {
	// Remove pair from set
	pairKey := shapePairKey(contact.shapeIDA, contact.shapeIDB)
	removeKey(&world.broadPhase.pairSet, pairKey)

	edgeA := &contact.edges[0]
	edgeB := &contact.edges[1]

	bodyIDA := edgeA.bodyID
	bodyIDB := edgeB.bodyID
	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	flags := contact.flags
	touching := (flags & contactTouchingFlag) != 0

	// End touch event
	if touching && (flags&contactEnableContactEvents) != 0 {
		shapeA := &world.shapes[contact.shapeIDA]
		shapeB := &world.shapes[contact.shapeIDB]
		shapeIDA := Shape{index1: int32(shapeA.id + 1), world: world, generation: shapeA.generation}
		shapeIDB := Shape{index1: int32(shapeB.id + 1), world: world, generation: shapeB.generation}

		contactID := Contact{index1: int32(contact.contactID + 1), world: world, generation: contact.generation}

		event := ContactEndTouchEvent{ShapeA: shapeIDA, ShapeB: shapeIDB, Contact: contactID}

		world.contactEndEvents[world.endEventArrayIndex] = arrayPush(world.contactEndEvents[world.endEventArrayIndex], event)
	}

	contactID := contact.contactID
	// Remove from body A
	disconnectContactFromBody(world, contact, bodyA, 0)
	// Remove from body B
	disconnectContactFromBody(world, contact, bodyB, 1)

	// Remove contact from island.
	if contact.islandID != nullIndex {
		unlinkContact(world, contact)
	}

	if contact.colorIndex != nullIndex {
		// contact is an active constraint
		assert(contact.setIndex == awakeSet, "contact.setIndex == awakeSet")
		removeContactFromGraph(world, bodyIDA, bodyIDB, contact.colorIndex, contact.localIndex)
	} else {
		// contact is non-touching or is sleeping
		assert(contact.setIndex != awakeSet || (contact.flags&contactTouchingFlag) == 0, "contact.setIndex != awakeSet || ( contact.flags & contactTouchingFlag ) == 0")
		set := &world.solverSets[contact.setIndex]

		var movedIndex int

		set.contactSims, movedIndex = arrayRemoveSwap(set.contactSims, contact.localIndex)
		if movedIndex != nullIndex {
			movedContactSim := &set.contactSims[contact.localIndex]
			movedContact := &world.contacts[movedContactSim.contactID]
			movedContact.localIndex = contact.localIndex
		}
	}

	// Free contact and id (preserve generation)
	contact.contactID = nullIndex
	contact.setIndex = nullIndex
	contact.colorIndex = nullIndex
	contact.localIndex = nullIndex

	freeID(&world.contactIDPool, contactID)

	if touching {
		wakeBody(world, bodyA)
		wakeBody(world, bodyB)
	}
}

func getContactSim(world *World, contact *contact) *contactSim {
	if contact.setIndex == awakeSet && contact.colorIndex != nullIndex {
		// contact lives in constraint graph
		assert(0 <= contact.colorIndex && contact.colorIndex < GraphColorCount, "0 <= contact.colorIndex && contact.colorIndex < GraphColorCount")
		color := &world.constraintGraph.colors[contact.colorIndex]

		return &color.contactSims[contact.localIndex]
	}

	set := &world.solverSets[contact.setIndex]

	return &set.contactSims[contact.localIndex]
}

func updateContact(world *World, contactSim *contactSim, shapeA *shape, transformA WorldTransform, centerOffsetA Vec2, shapeB *shape, transformB WorldTransform, centerOffsetB Vec2) bool {
	// Save old manifold
	oldManifold := contactSim.manifold

	// Compute the manifold in frame A, then marshal it to world anchors relative to each shape origin.
	// The relative pose differences the two world positions before the narrow phase runs in frame A,
	// so precision is retained far from the origin.
	// anchorB = worldPoint - pB = rot(qA, localAnchorA) + pA - pB = anchorA + (pA - pB)
	relativeTransform := transformA.InvMul(transformB)
	fcn := contactRegistry[shapeA.shapeType][shapeB.shapeType].fcn
	local := fcn(shapeA, shapeB, relativeTransform, &contactSim.cache)

	contactSim.manifold = Manifold{}
	contactSim.manifold.Normal = transformA.Q.Rotate(local.Normal)
	contactSim.manifold.PointCount = local.PointCount

	originDelta := transformA.P.Sub(transformB.P)

	for i := range local.PointCount {
		mp := &contactSim.manifold.Points[i]
		mp.AnchorA = transformA.Q.Rotate(local.Points[i].Point)
		mp.AnchorB = mp.AnchorA.Add(originDelta)
		mp.Separation = local.Points[i].Separation
		mp.ID = local.Points[i].ID
	}

	// Keep these updated in case the values on the shapes are modified
	mixContactMaterials(world, contactSim, shapeA, shapeB)

	if shapeA.material.RollingResistance > 0.0 || shapeB.material.RollingResistance > 0.0 {
		radiusA := getShapeRadius(shapeA)
		radiusB := getShapeRadius(shapeB)
		maxRadius := maxFloat(radiusA, radiusB)
		contactSim.rollingResistance =
			maxFloat(shapeA.material.RollingResistance, shapeB.material.RollingResistance) * maxRadius
	} else {
		contactSim.rollingResistance = 0.0
	}

	contactSim.tangentSpeed = shapeA.material.TangentSpeed + shapeB.material.TangentSpeed

	pointCount := contactSim.manifold.PointCount
	touching := pointCount > 0

	if touching && world.preSolveFcn != nil && (contactSim.simFlags&simEnablePreSolveEvents) != 0 {
		shapeIDA := Shape{index1: int32(shapeA.id + 1), world: world, generation: shapeA.generation}
		shapeIDB := Shape{index1: int32(shapeB.id + 1), world: world, generation: shapeB.generation}

		// This call assumes thread safety.
		world.preSolveFcn(shapeIDA, shapeIDB, &contactSim.manifold)

		// Keep these current.
		pointCount = contactSim.manifold.PointCount
		touching = pointCount > 0
	}

	if touching && (shapeA.enableHitEvents || shapeB.enableHitEvents) {
		contactSim.simFlags |= simEnableHitEvent
	} else {
		contactSim.simFlags &= ^simEnableHitEvent
	}

	warmStartContactManifold(contactSim, &oldManifold, centerOffsetA, centerOffsetB)

	if touching {
		contactSim.simFlags |= simTouchingFlag
	} else {
		contactSim.simFlags &= ^simTouchingFlag
	}

	return touching
}

func mixFriction(frictionA float32, frictionB float32) float32 {
	return float32(math.Sqrt(float64(frictionA * frictionB)))
}

func mixRestitution(restitutionA float32, restitutionB float32) float32 {
	return maxFloat(restitutionA, restitutionB)
}
