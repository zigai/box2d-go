// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

const (
	staticSet        int = 0
	disabledSet      int = 1
	awakeSet         int = 2
	firstSleepingSet int = 3
)

type solverSet struct {
	bodySims    []bodySim
	bodyStates  []bodyState
	jointSims   []jointSim
	contactSims []contactSim
	islandSims  []islandSim
	setIndex    int
}

func destroySolverSet(world *World, setIndex int) {
	set := &world.solverSets[setIndex]
	set.bodySims = nil
	set.bodyStates = nil
	set.contactSims = nil
	set.jointSims = nil
	set.islandSims = nil

	freeID(&world.solverSetIDPool, setIndex)

	*set = solverSet{}
	set.setIndex = nullIndex
}

func wakeSolverSet(world *World, setIndex int) {
	assert(setIndex >= firstSleepingSet, "setIndex >= b2_firstSleepingSet")
	set := &world.solverSets[setIndex]
	awake := &world.solverSets[awakeSet]
	disabled := &world.solverSets[disabledSet]

	bodies := world.bodies

	bodyCount := len(set.bodySims)
	for i := range bodyCount {
		simSrc := &set.bodySims[i]

		body := &bodies[simSrc.bodyID]
		assert(body.setIndex == setIndex, "body->setIndex == setIndex")
		body.setIndex = awakeSet
		body.localIndex = len(awake.bodySims)

		// Reset sleep timer
		body.sleepTime = 0.0

		awake.bodySims, _ = arrayEmplace(awake.bodySims)
		simDst := &awake.bodySims[len(awake.bodySims)-1]
		*simDst = *simSrc

		awake.bodyStates, _ = arrayEmplace(awake.bodyStates)
		state := &awake.bodyStates[len(awake.bodyStates)-1]
		*state = identityBodyState
		state.flags = body.flags

		wakeBodyNonTouchingContacts(world, body, setIndex, awake, disabled)
	}

	// transfer touching contacts from sleeping set to contact graph
	{
		contactCount := len(set.contactSims)
		for i := range contactCount {
			contactSim := &set.contactSims[i]
			contact := &world.contacts[contactSim.contactID]
			assert((contact.flags&contactTouchingFlag) != 0, "contact->flags & b2_contactTouchingFlag")
			assert((contactSim.simFlags&simTouchingFlag) != 0, "contactSim->simFlags & b2_simTouchingFlag")
			assert(contactSim.manifold.PointCount > 0, "contactSim->manifold.pointCount > 0")
			assert(contact.setIndex == setIndex, "contact->setIndex == setIndex")
			addContactToGraph(world, contactSim, contact)
			contact.setIndex = awakeSet
		}
	}

	// transfer joints from sleeping set to awake set
	{
		jointCount := len(set.jointSims)
		for i := range jointCount {
			jointSim := &set.jointSims[i]
			joint := &world.joints[jointSim.jointID]
			assert(joint.setIndex == setIndex, "joint->setIndex == setIndex")
			addJointToGraph(world, jointSim, joint)
			joint.setIndex = awakeSet
		}
	}

	// transfer island from sleeping set to awake set
	// Usually a sleeping set has only one island, but it is possible
	// that joints are created between sleeping islands and they
	// are moved to the same sleeping set.
	{
		islandCount := len(set.islandSims)
		for i := range islandCount {
			islandSrc := &set.islandSims[i]
			island := &world.islands[islandSrc.islandID]
			island.setIndex = awakeSet
			island.localIndex = len(awake.islandSims)
			awake.islandSims, _ = arrayEmplace(awake.islandSims)
			islandDst := &awake.islandSims[len(awake.islandSims)-1]
			*islandDst = *islandSrc
		}
	}

	// destroy the sleeping set
	destroySolverSet(world, setIndex)
}

func trySleepIsland(world *World, islandID int) {
	island := &world.islands[islandID]
	assert(island.setIndex == awakeSet, "island->setIndex == b2_awakeSet")

	// Cannot put an island to sleep while it has a pending split and more than one body.
	if island.constraintRemoveCount > 0 && len(island.bodies) > 1 {
		return
	}

	// island is sleeping
	// - create new sleeping solver set
	// - move island to sleeping solver set
	// - identify non-touching contacts that should move to sleeping solver set or disabled set
	// - remove old island
	// - fix island
	sleepSetID := allocID(&world.solverSetIDPool)
	if sleepSetID == len(world.solverSets) {
		set := solverSet{}
		set.setIndex = nullIndex
		world.solverSets = arrayPush(world.solverSets, set)
	}

	sleepSet := &world.solverSets[sleepSetID]
	*sleepSet = solverSet{}

	// grab awake set after creating the sleep set because the solver set array may have been resized
	awake := &world.solverSets[awakeSet]
	assert(0 <= island.localIndex && island.localIndex < len(awake.islandSims), "0 <= island->localIndex && island->localIndex < awakeSet->islandSims.count")

	sleepSet.setIndex = sleepSetID
	sleepSet.bodySims = make([]bodySim, 0, len(island.bodies))
	sleepSet.contactSims = make([]contactSim, 0, len(island.contacts))
	sleepSet.jointSims = make([]jointSim, 0, len(island.joints))

	// move awake bodies to sleeping set
	// this shuffles around bodies in the awake set
	{
		disabled := &world.solverSets[disabledSet]

		for i := range len(island.bodies) {
			bodyID := island.bodies[i]
			body := &world.bodies[bodyID]
			assert(body.setIndex == awakeSet, "body->setIndex == b2_awakeSet")
			assert(body.islandID == islandID, "body->islandId == islandId")
			assert(body.islandIndex == i, "body->islandIndex == i")

			body.flags &= ^bodyTransientFlags

			// Update the body move event to indicate this body fell asleep
			// It could happen the body is forced asleep before it ever moves.
			if body.bodyMoveIndex != nullIndex {
				moveEvent := &world.bodyMoveEvents[body.bodyMoveIndex]
				assert(int(moveEvent.Body.index1)-1 == bodyID, "moveEvent->bodyId.index1 - 1 == bodyId")
				assert(moveEvent.Body.generation == body.generation, "moveEvent->bodyId.generation == body->generation")
				moveEvent.FellAsleep = true
				body.bodyMoveIndex = nullIndex
			}

			awakeBodyIndex := body.localIndex
			awakeSim := &awake.bodySims[awakeBodyIndex]
			awakeSim.flags &= ^bodyTransientFlags

			// move body sim to sleep set
			sleepBodyIndex := len(sleepSet.bodySims)
			sleepSet.bodySims, _ = arrayEmplace(sleepSet.bodySims)
			sleepBodySim := &sleepSet.bodySims[len(sleepSet.bodySims)-1]
			*sleepBodySim = *awakeSim

			movedBody := removeBodySim(world, awake, awakeBodyIndex)
			if movedBody != nil && movedBody.islandID != islandID {
				// Update the contact sims for the moved body.
				refreshBodyContactIndices(world, movedBody)
			}

			// Destroy state, no need to clone.
			awake.bodyStates, _ = arrayRemoveSwap(awake.bodyStates, awakeBodyIndex)

			body.setIndex = sleepSetID
			body.localIndex = sleepBodyIndex

			sleepBodyNonTouchingContacts(world, body, awake, disabled)
		}
	}

	moveIslandContactsToSleepingSet(world, island, islandID, sleepSet, sleepSetID)

	moveIslandJointsToSleepingSet(world, island, islandID, sleepSet, sleepSetID)

	// move island struct
	{
		assert(island.setIndex == awakeSet, "island->setIndex == b2_awakeSet")

		islandIndex := island.localIndex
		sleepSet.islandSims, _ = arrayEmplace(sleepSet.islandSims)
		sleepIsland := &sleepSet.islandSims[len(sleepSet.islandSims)-1]
		sleepIsland.islandID = islandID

		var movedIslandIndex int

		awake.islandSims, movedIslandIndex = arrayRemoveSwap(awake.islandSims, islandIndex)
		if movedIslandIndex != nullIndex {
			// fix index on moved element
			movedIslandSim := &awake.islandSims[islandIndex]
			movedIslandID := movedIslandSim.islandID
			movedIsland := &world.islands[movedIslandID]
			assert(movedIsland.localIndex == movedIslandIndex, "movedIsland->localIndex == movedIslandIndex")
			movedIsland.localIndex = islandIndex
		}

		island.setIndex = sleepSetID
		island.localIndex = 0
	}

	if world.splitIslandID == islandID {
		world.splitIslandID = nullIndex
	}

	validateSolverSets(world)
}

func mergeSolverSets(world *World, setID1 int, setID2 int) {
	assert(setID1 >= firstSleepingSet, "setId1 >= b2_firstSleepingSet")
	assert(setID2 >= firstSleepingSet, "setId2 >= b2_firstSleepingSet")

	set1 := &world.solverSets[setID1]
	set2 := &world.solverSets[setID2]

	// Move the fewest number of bodies
	if len(set1.bodySims) < len(set2.bodySims) {
		set1, set2 = set2, set1

		setID1, setID2 = setID2, setID1
	}

	// transfer bodies
	{
		bodies := world.bodies

		bodyCount := len(set2.bodySims)
		for i := range bodyCount {
			simSrc := &set2.bodySims[i]

			body := &bodies[simSrc.bodyID]
			assert(body.setIndex == setID2, "body->setIndex == setId2")
			body.setIndex = setID1
			body.localIndex = len(set1.bodySims)

			set1.bodySims, _ = arrayEmplace(set1.bodySims)
			simDst := &set1.bodySims[len(set1.bodySims)-1]
			*simDst = *simSrc
		}
	}

	// transfer contacts
	{
		contactCount := len(set2.contactSims)
		for i := range contactCount {
			contactSrc := &set2.contactSims[i]

			contact := &world.contacts[contactSrc.contactID]
			assert(contact.setIndex == setID2, "contact->setIndex == setId2")
			contact.setIndex = setID1
			contact.localIndex = len(set1.contactSims)

			set1.contactSims, _ = arrayEmplace(set1.contactSims)
			contactDst := &set1.contactSims[len(set1.contactSims)-1]
			*contactDst = *contactSrc
		}
	}

	// transfer joints
	{
		jointCount := len(set2.jointSims)
		for i := range jointCount {
			jointSrc := &set2.jointSims[i]

			joint := &world.joints[jointSrc.jointID]
			assert(joint.setIndex == setID2, "joint->setIndex == setId2")
			joint.setIndex = setID1
			joint.localIndex = len(set1.jointSims)

			set1.jointSims, _ = arrayEmplace(set1.jointSims)
			jointDst := &set1.jointSims[len(set1.jointSims)-1]
			*jointDst = *jointSrc
		}
	}

	// transfer islands
	{
		islandCount := len(set2.islandSims)
		for i := range islandCount {
			islandSrc := &set2.islandSims[i]
			islandID := islandSrc.islandID

			island := &world.islands[islandID]
			island.setIndex = setID1
			island.localIndex = len(set1.islandSims)

			set1.islandSims, _ = arrayEmplace(set1.islandSims)
			islandDst := &set1.islandSims[len(set1.islandSims)-1]
			*islandDst = *islandSrc
		}
	}

	// destroy the merged set
	destroySolverSet(world, setID2)

	validateSolverSets(world)
}

func removeBodySim(world *World, set *solverSet, localIndex int) *body {
	bodySims := set.bodySims
	assert(0 <= localIndex && localIndex < len(bodySims), "0 <= localIndex && localIndex < bodySims->count")
	lastIndex := len(bodySims) - 1
	set.bodySims = set.bodySims[:lastIndex]

	if localIndex == lastIndex {
		return nil
	}

	bodySims[localIndex] = bodySims[lastIndex]
	bodySims[lastIndex] = bodySim{}
	movedBody := &world.bodies[bodySims[localIndex].bodyID]
	assert(movedBody.localIndex == lastIndex, "movedBody->localIndex == lastIndex")
	movedBody.localIndex = localIndex

	return movedBody
}

func transferBody(world *World, targetSet *solverSet, sourceSet *solverSet, body *body) {
	assert(body.headContactKey == nullIndex, "body->headContactKey == B2_NULL_INDEX")

	if targetSet == sourceSet {
		return
	}

	// The use cases for transfer body (set body type, disable, etc) destroy all the contacts
	// first. Double check here because this function wouldn't work correctly otherwise.
	assert(body.headContactKey == nullIndex, "body->headContactKey == B2_NULL_INDEX")

	sourceIndex := body.localIndex
	sourceSim := &sourceSet.bodySims[sourceIndex]

	targetIndex := len(targetSet.bodySims)
	targetSet.bodySims, _ = arrayEmplace(targetSet.bodySims)
	targetSim := &targetSet.bodySims[len(targetSet.bodySims)-1]
	*targetSim = *sourceSim

	// Clear transient body flags
	body.flags &= ^bodyTransientFlags
	targetSim.flags &= ^(isFast | bodyTransientFlags)

	// Remove body sim from solver set that owns it
	movedBody := removeBodySim(world, sourceSet, sourceIndex)
	if movedBody != nil {
		// Fixup the moved body's contact sims.
		refreshBodyContactIndices(world, movedBody)
	}

	if sourceSet.setIndex == awakeSet {
		sourceSet.bodyStates, _ = arrayRemoveSwap(sourceSet.bodyStates, sourceIndex)
	} else if targetSet.setIndex == awakeSet {
		targetSet.bodyStates, _ = arrayEmplace(targetSet.bodyStates)
		state := &targetSet.bodyStates[len(targetSet.bodyStates)-1]
		*state = identityBodyState
		state.flags = body.flags
	}

	body.setIndex = targetSet.setIndex
	body.localIndex = targetIndex
}

func transferJoint(world *World, targetSet *solverSet, sourceSet *solverSet, joint *joint) {
	if targetSet == sourceSet {
		return
	}

	localIndex := joint.localIndex
	colorIndex := joint.colorIndex

	// Retrieve source.
	var sourceSim *jointSim

	if sourceSet.setIndex == awakeSet {
		assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")
		color := &world.constraintGraph.colors[colorIndex]

		sourceSim = &color.jointSims[localIndex]
	} else {
		assert(colorIndex == nullIndex, "colorIndex == B2_NULL_INDEX")

		sourceSim = &sourceSet.jointSims[localIndex]
	}

	// Create target and copy. Fix joint.
	if targetSet.setIndex == awakeSet {
		addJointToGraph(world, sourceSim, joint)
		joint.setIndex = awakeSet
	} else {
		joint.setIndex = targetSet.setIndex
		joint.localIndex = len(targetSet.jointSims)
		joint.colorIndex = nullIndex

		targetSet.jointSims, _ = arrayEmplace(targetSet.jointSims)
		targetSim := &targetSet.jointSims[len(targetSet.jointSims)-1]
		*targetSim = *sourceSim
	}

	// Destroy source.
	if sourceSet.setIndex == awakeSet {
		removeJointFromGraph(world, joint.edges[0].bodyID, joint.edges[1].bodyID, colorIndex, localIndex)
	} else {
		var movedIndex int

		sourceSet.jointSims, movedIndex = arrayRemoveSwap(sourceSet.jointSims, localIndex)
		if movedIndex != nullIndex {
			// fix swapped element
			movedJointSim := &sourceSet.jointSims[localIndex]
			movedID := movedJointSim.jointID
			movedJoint := &world.joints[movedID]
			movedJoint.localIndex = localIndex
		}
	}
}

func wakeBodyNonTouchingContacts(world *World, body *body, setIndex int, awake, disabled *solverSet) {
	// Move non-touching contacts from disabled set to awake set.
	encodedBodySimIndex := body.localIndex

	contactKey := body.headContactKey
	for contactKey != nullIndex {
		edgeIndex := contactKey & 1
		contactID := contactKey >> 1

		contact := &world.contacts[contactID]

		contactKey = contact.edges[edgeIndex].nextKey

		if contact.setIndex != disabledSet {
			assert(contact.setIndex == awakeSet || contact.setIndex == setIndex, "contact->setIndex == b2_awakeSet || contact->setIndex == setIndex")

			// If this contact is awake but not touching I need to fixup the body sim index here.
			if contact.setIndex == awakeSet {
				assert(contact.colorIndex == nullIndex, "contact->colorIndex == B2_NULL_INDEX")

				awakeContactSim := &awake.contactSims[contact.localIndex]
				if edgeIndex == 0 {
					awakeContactSim.encodedBodySimA = encodedBodySimIndex
				} else {
					awakeContactSim.encodedBodySimB = encodedBodySimIndex
				}
			}

			continue
		}

		localIndex := contact.localIndex
		contactSim := &disabled.contactSims[localIndex]

		assert((contact.flags&contactTouchingFlag) == 0 && contactSim.manifold.PointCount == 0, "( contact->flags & b2_contactTouchingFlag ) == 0 && contactSim->manifold.pointCount == 0")

		contact.setIndex = awakeSet
		contact.localIndex = len(awake.contactSims)
		awake.contactSims, _ = arrayEmplace(awake.contactSims)
		awakeContactSim := &awake.contactSims[len(awake.contactSims)-1]
		*awakeContactSim = *contactSim

		if edgeIndex == 0 {
			awakeContactSim.encodedBodySimA = encodedBodySimIndex
		} else {
			awakeContactSim.encodedBodySimB = encodedBodySimIndex
		}

		var movedLocalIndex int

		disabled.contactSims, movedLocalIndex = arrayRemoveSwap(disabled.contactSims, localIndex)
		if movedLocalIndex != nullIndex {
			// fix moved element
			movedContactSim := &disabled.contactSims[localIndex]
			movedContact := &world.contacts[movedContactSim.contactID]
			assert(movedContact.localIndex == movedLocalIndex, "movedContact->localIndex == movedLocalIndex")
			movedContact.localIndex = localIndex
		}
	}
}

func sleepBodyNonTouchingContacts(world *World, body *body, awake, disabled *solverSet) {
	// Move non-touching contacts to the disabled set.
	// Non-touching contacts may exist between sleeping islands and there is no clear ownership.
	contactKey := body.headContactKey
	for contactKey != nullIndex {
		contactID := contactKey >> 1
		edgeIndex := contactKey & 1

		contact := &world.contacts[contactID]

		assert(contact.setIndex == awakeSet, "contact->setIndex == b2_awakeSet")

		contactKey = contact.edges[edgeIndex].nextKey

		if contact.colorIndex != nullIndex {
			// contact is touching and will be moved separately
			assert((contact.flags&contactTouchingFlag) != 0, "( contact->flags & b2_contactTouchingFlag ) != 0")
			continue
		}

		localIndex := contact.localIndex
		contactSim := &awake.contactSims[localIndex]

		if edgeIndex == 0 {
			contactSim.encodedBodySimA = nullIndex
		} else {
			contactSim.encodedBodySimB = nullIndex
		}

		// the other body may still be awake, it still may go to sleep and then it will be responsible
		// for moving this contact to the disabled set.
		otherEdgeIndex := edgeIndex ^ 1
		otherBodyID := contact.edges[otherEdgeIndex].bodyID

		otherBody := &world.bodies[otherBodyID]
		if otherBody.setIndex == awakeSet {
			continue
		}

		assert(contactSim.manifold.PointCount == 0, "contactSim->manifold.pointCount == 0")
		assert((contact.flags&contactTouchingFlag) == 0, "( contact->flags & b2_contactTouchingFlag ) == 0")

		// move the non-touching contact to the disabled set
		contact.setIndex = disabledSet
		contact.localIndex = len(disabled.contactSims)
		disabled.contactSims, _ = arrayEmplace(disabled.contactSims)
		disabledContactSim := &disabled.contactSims[len(disabled.contactSims)-1]
		*disabledContactSim = *contactSim

		var movedLocalIndex int

		awake.contactSims, movedLocalIndex = arrayRemoveSwap(awake.contactSims, localIndex)
		if movedLocalIndex != nullIndex {
			// fix moved element
			movedContactSim := &awake.contactSims[localIndex]
			movedContact := &world.contacts[movedContactSim.contactID]
			assert(movedContact.localIndex == movedLocalIndex, "movedContact->localIndex == movedLocalIndex")
			movedContact.localIndex = localIndex
		}
	}
}

func moveIslandContactsToSleepingSet(world *World, island *island, islandID int, sleepSet *solverSet, sleepSetID int) {
	// move touching contacts
	// this shuffles contacts in the awake set
	{
		for i := range len(island.contacts) {
			link := &island.contacts[i]
			contact := &world.contacts[link.contactID]
			assert(contact.setIndex == awakeSet, "contact->setIndex == b2_awakeSet")
			assert(contact.islandID == islandID, "contact->islandId == islandId")
			colorIndex := contact.colorIndex
			assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")

			color := &world.constraintGraph.colors[colorIndex]

			// Remove bodies from graph coloring associated with this constraint
			if colorIndex != overflowIndex {
				// might clear a bit for a static body, but this has no effect
				clearBit(&color.bodySet, uint32(contact.edges[0].bodyID))
				clearBit(&color.bodySet, uint32(contact.edges[1].bodyID))
			}

			localIndex := contact.localIndex
			awakeContactSim := &color.contactSims[localIndex]

			sleepContactIndex := len(sleepSet.contactSims)
			sleepSet.contactSims, _ = arrayEmplace(sleepSet.contactSims)
			sleepContactSim := &sleepSet.contactSims[len(sleepSet.contactSims)-1]
			*sleepContactSim = *awakeContactSim

			sleepContactSim.encodedBodySimA = sleepBodySimIndex(sleepContactSim.encodedBodySimA)
			sleepContactSim.encodedBodySimB = sleepBodySimIndex(sleepContactSim.encodedBodySimB)

			var movedLocalIndex int

			color.contactSims, movedLocalIndex = arrayRemoveSwap(color.contactSims, localIndex)
			if movedLocalIndex != nullIndex {
				// fix moved element
				movedContactSim := &color.contactSims[localIndex]
				movedContact := &world.contacts[movedContactSim.contactID]
				assert(movedContact.localIndex == movedLocalIndex, "movedContact->localIndex == movedLocalIndex")
				movedContact.localIndex = localIndex
			}

			contact.setIndex = sleepSetID
			contact.colorIndex = nullIndex
			contact.localIndex = sleepContactIndex
		}
	}
}

func moveIslandJointsToSleepingSet(world *World, island *island, islandID int, sleepSet *solverSet, sleepSetID int) {
	// move joints
	// this shuffles joints in the awake set
	{
		for i := range len(island.joints) {
			link := &island.joints[i]
			joint := &world.joints[link.jointID]
			assert(joint.setIndex == awakeSet, "joint->setIndex == b2_awakeSet")
			assert(joint.islandID == islandID, "joint->islandId == islandId")
			colorIndex := joint.colorIndex
			localIndex := joint.localIndex

			assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")

			color := &world.constraintGraph.colors[colorIndex]

			awakeJointSim := &color.jointSims[localIndex]

			if colorIndex != overflowIndex {
				// might clear a bit for a static body, but this has no effect
				clearBit(&color.bodySet, uint32(joint.edges[0].bodyID))
				clearBit(&color.bodySet, uint32(joint.edges[1].bodyID))
			}

			sleepJointIndex := len(sleepSet.jointSims)
			sleepSet.jointSims, _ = arrayEmplace(sleepSet.jointSims)
			sleepJointSim := &sleepSet.jointSims[len(sleepSet.jointSims)-1]
			*sleepJointSim = *awakeJointSim

			var movedIndex int

			color.jointSims, movedIndex = arrayRemoveSwap(color.jointSims, localIndex)
			if movedIndex != nullIndex {
				// fix moved element
				movedJointSim := &color.jointSims[localIndex]
				movedID := movedJointSim.jointID
				movedJoint := &world.joints[movedID]
				assert(movedJoint.localIndex == movedIndex, "movedJoint->localIndex == movedIndex")
				movedJoint.localIndex = localIndex
			}

			joint.setIndex = sleepSetID
			joint.colorIndex = nullIndex
			joint.localIndex = sleepJointIndex
		}
	}
}
