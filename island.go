// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type contactLink struct {
	contactID int
	bodyIDA   int
	bodyIDB   int
}

type jointLink struct {
	jointID int
	bodyIDA int
	bodyIDB int
}

type island struct {
	setIndex              int
	localIndex            int
	islandID              int
	constraintRemoveCount int
	bodies                []int
	contacts              []contactLink
	joints                []jointLink
}

type islandSim struct {
	islandID int
}

func createIsland(world *World, setIndex int) *island {
	assert(setIndex == awakeSet || setIndex >= firstSleepingSet, "setIndex == b2_awakeSet || setIndex >= b2_firstSleepingSet")

	islandID := allocID(&world.islandIDPool)

	if islandID == len(world.islands) {
		emptyIsland := island{}
		world.islands = arrayPush(world.islands, emptyIsland)
	} else {
		assert(world.islands[islandID].setIndex == nullIndex, "world->islands.data[islandId].setIndex == B2_NULL_INDEX")
	}

	set := &world.solverSets[setIndex]

	island := &world.islands[islandID]
	island.setIndex = setIndex
	island.localIndex = len(set.islandSims)
	island.islandID = islandID
	island.bodies = nil
	island.contacts = nil
	island.joints = nil
	island.constraintRemoveCount = 0

	var islandSim *islandSim

	set.islandSims, islandSim = arrayEmplace(set.islandSims)
	islandSim.islandID = islandID

	return island
}

func destroyIsland(world *World, islandID int) {
	if world.splitIslandID == islandID {
		world.splitIslandID = nullIndex
	}

	// assume island is empty
	island := &world.islands[islandID]
	set := &world.solverSets[island.setIndex]
	{
		localIndex := island.localIndex
		lastIndex := len(set.islandSims) - 1
		assert(0 <= localIndex && localIndex <= lastIndex, "0 <= localIndex && localIndex <= lastIndex")
		moveIslandID := set.islandSims[lastIndex].islandID
		set.islandSims[localIndex] = set.islandSims[lastIndex]
		world.islands[moveIslandID].localIndex = localIndex
		set.islandSims = set.islandSims[:len(set.islandSims)-1]
	}

	// Free island and id (preserve island revision)
	island.bodies = nil
	island.contacts = nil
	island.joints = nil
	island.constraintRemoveCount = 0
	island.localIndex = nullIndex
	island.islandID = nullIndex
	island.setIndex = nullIndex

	validate(island.localIndex == nullIndex, "island->localIndex == B2_NULL_INDEX")

	freeID(&world.islandIDPool, islandID)
}

func mergeIslands(world *World, islandIDA int, islandIDB int) int {
	if islandIDA == islandIDB {
		return islandIDA
	}

	if islandIDA == nullIndex {
		assert(islandIDB != nullIndex, "islandIdB != B2_NULL_INDEX")
		return islandIDB
	}

	if islandIDB == nullIndex {
		assert(islandIDA != nullIndex, "islandIdA != B2_NULL_INDEX")
		return islandIDA
	}

	var (
		smallIsland *island
		bigIsland   *island
	)
	{
		islandA := &world.islands[islandIDA]
		islandB := &world.islands[islandIDB]

		// Keep the biggest island to reduce cache misses
		if len(islandA.bodies) >= len(islandB.bodies) {
			bigIsland = islandA
			smallIsland = islandB
		} else {
			bigIsland = islandB
			smallIsland = islandA
		}
	}

	bigIslandID := bigIsland.islandID
	bigIsland.bodies = arrayReserve(bigIsland.bodies, len(bigIsland.bodies)+len(smallIsland.bodies))

	// Move bodies from smaller island to larger island
	for i := range len(smallIsland.bodies) {
		bodyID := smallIsland.bodies[i]
		body := &world.bodies[bodyID]
		validate(body.islandID == smallIsland.islandID, "body->islandId == smallIsland->islandId")
		body.islandID = bigIslandID
		body.islandIndex = len(bigIsland.bodies)
		bigIsland.bodies = arrayPush(bigIsland.bodies, bodyID)
	}

	// Migrate contacts from smaller island to larger island
	if len(smallIsland.contacts) > 0 {
		bigIsland.contacts = arrayReserve(bigIsland.contacts, len(bigIsland.contacts)+len(smallIsland.contacts))

		for i := range len(smallIsland.contacts) {
			link := &smallIsland.contacts[i]
			contact := &world.contacts[link.contactID]
			contact.islandID = bigIslandID
			contact.islandIndex = len(bigIsland.contacts)
			bigIsland.contacts = arrayPush(bigIsland.contacts, *link)
		}
	}

	// Migrate joints from smaller island to larger island
	if len(smallIsland.joints) > 0 {
		bigIsland.joints = arrayReserve(bigIsland.joints, len(bigIsland.joints)+len(smallIsland.joints))

		for i := range len(smallIsland.joints) {
			link := &smallIsland.joints[i]
			joint := &world.joints[link.jointID]
			joint.islandId = bigIslandID
			joint.islandIndex = len(bigIsland.joints)
			bigIsland.joints = arrayPush(bigIsland.joints, *link)
		}
	}

	// Track removed constraints
	bigIsland.constraintRemoveCount += smallIsland.constraintRemoveCount

	destroyIsland(world, smallIsland.islandID)

	validateIsland(world, bigIslandID)

	return bigIslandID
}

func addContactToIsland(world *World, islandID int, contact *contact) {
	assert(contact.islandID == nullIndex, "contact->islandId == B2_NULL_INDEX")
	assert(contact.islandIndex == nullIndex, "contact->islandIndex == B2_NULL_INDEX")

	island := &world.islands[islandID]

	contact.islandID = islandID
	contact.islandIndex = len(island.contacts)

	var link contactLink

	link.contactID = contact.contactID
	link.bodyIDA = contact.edges[0].bodyID
	link.bodyIDB = contact.edges[1].bodyID
	island.contacts = arrayPush(island.contacts, link)

	validateIsland(world, islandID)
}

func linkContact(world *World, contact *contact) {
	assert((contact.flags&contactTouchingFlag) != 0, "( contact->flags & b2_contactTouchingFlag ) != 0")

	bodyIDA := contact.edges[0].bodyID
	bodyIDB := contact.edges[1].bodyID

	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	assert(bodyA.setIndex != disabledSet && bodyB.setIndex != disabledSet, "bodyA->setIndex != b2_disabledSet && bodyB->setIndex != b2_disabledSet")
	assert(bodyA.setIndex != staticSet || bodyB.setIndex != staticSet, "bodyA->setIndex != b2_staticSet || bodyB->setIndex != b2_staticSet")

	// Wake bodyB if bodyA is awake and bodyB is sleeping
	if bodyA.setIndex == awakeSet && bodyB.setIndex >= firstSleepingSet {
		wakeSolverSet(world, bodyB.setIndex)
	}

	// Wake bodyA if bodyB is awake and bodyA is sleeping
	if bodyB.setIndex == awakeSet && bodyA.setIndex >= firstSleepingSet {
		wakeSolverSet(world, bodyA.setIndex)
	}

	islandIDA := bodyA.islandID
	islandIDB := bodyB.islandID

	// Static bodies have null island indices.
	assert(bodyA.setIndex != staticSet || islandIDA == nullIndex, "bodyA->setIndex != b2_staticSet || islandIdA == B2_NULL_INDEX")
	assert(bodyB.setIndex != staticSet || islandIDB == nullIndex, "bodyB->setIndex != b2_staticSet || islandIdB == B2_NULL_INDEX")
	assert(islandIDA != nullIndex || islandIDB != nullIndex, "islandIdA != B2_NULL_INDEX || islandIdB != B2_NULL_INDEX")

	// Merge islands. This will destroy one of the islands.
	finalIslandID := mergeIslands(world, islandIDA, islandIDB)

	// Add contact to the island that survived
	addContactToIsland(world, finalIslandID, contact)
}

func unlinkContact(world *World, contact *contact) {
	assert(contact.islandID != nullIndex, "contact->islandId != B2_NULL_INDEX")

	// remove from island
	islandID := contact.islandID
	island := &world.islands[islandID]

	removeIndex := contact.islandIndex
	assert(0 <= removeIndex && removeIndex < len(island.contacts), "0 <= removeIndex && removeIndex < island->contacts.count")
	assert(island.contacts[removeIndex].contactID == contact.contactID, "island->contacts.data[removeIndex].contactId == contact->contactId")

	var movedIndex int

	island.contacts, movedIndex = arrayRemoveSwap(island.contacts, removeIndex)
	if movedIndex != nullIndex {
		// Fix islandIndex on the contact that was swapped into removeIndex
		movedLink := &island.contacts[removeIndex]
		movedContact := &world.contacts[movedLink.contactID]
		assert(movedContact.islandIndex == movedIndex, "movedContact->islandIndex == movedIndex")
		movedContact.islandIndex = removeIndex
	}

	contact.islandID = nullIndex
	contact.islandIndex = nullIndex
	island.constraintRemoveCount += 1

	validateIsland(world, islandID)
}

func addJointToIsland(world *World, islandID int, joint *joint) {
	assert(joint.islandId == nullIndex, "joint->islandId == B2_NULL_INDEX")
	assert(joint.islandIndex == nullIndex, "joint->islandIndex == B2_NULL_INDEX")

	island := &world.islands[islandID]

	joint.islandId = islandID
	joint.islandIndex = len(island.joints)

	var link jointLink

	link.jointID = joint.jointId
	link.bodyIDA = joint.edges[0].bodyId
	link.bodyIDB = joint.edges[1].bodyId
	island.joints = arrayPush(island.joints, link)

	validateIsland(world, islandID)
}

func linkJoint(world *World, joint *joint) {
	bodyA := &world.bodies[joint.edges[0].bodyId]
	bodyB := &world.bodies[joint.edges[1].bodyId]

	assert(bodyA.bodyType == DynamicBody || bodyB.bodyType == DynamicBody, "bodyA->type == b2_dynamicBody || bodyB->type == b2_dynamicBody")

	if bodyA.setIndex == awakeSet && bodyB.setIndex >= firstSleepingSet {
		wakeSolverSet(world, bodyB.setIndex)
	} else if bodyB.setIndex == awakeSet && bodyA.setIndex >= firstSleepingSet {
		wakeSolverSet(world, bodyA.setIndex)
	}

	islandIDA := bodyA.islandID
	islandIDB := bodyB.islandID

	assert(islandIDA != nullIndex || islandIDB != nullIndex, "islandIdA != B2_NULL_INDEX || islandIdB != B2_NULL_INDEX")

	// Merge islands. This will destroy one of the islands.
	finalIslandID := mergeIslands(world, islandIDA, islandIDB)

	// Add joint the island that survived
	addJointToIsland(world, finalIslandID, joint)
}

func unlinkJoint(world *World, joint *joint) {
	if joint.islandId == nullIndex {
		return
	}

	// remove from island
	islandID := joint.islandId
	island := &world.islands[islandID]

	removeIndex := joint.islandIndex
	assert(0 <= removeIndex && removeIndex < len(island.joints), "0 <= removeIndex && removeIndex < island->joints.count")
	assert(island.joints[removeIndex].jointID == joint.jointId, "island->joints.data[removeIndex].jointId == joint->jointId")

	var movedIndex int

	island.joints, movedIndex = arrayRemoveSwap(island.joints, removeIndex)
	if movedIndex != nullIndex {
		// Fix islandIndex on the joint that was swapped into removeIndex
		movedLink := &island.joints[removeIndex]
		movedJoint := &world.joints[movedLink.jointID]
		assert(movedJoint.islandIndex == movedIndex, "movedJoint->islandIndex == movedIndex")
		movedJoint.islandIndex = removeIndex
	}

	joint.islandId = nullIndex
	joint.islandIndex = nullIndex
	island.constraintRemoveCount += 1

	validateIsland(world, islandID)
}

func islandFindParent(parents []int, node int) int {
	// Walk the chain of parents to find the node that is its own parent (the root)
	for parents[node] != node {
		grandParent := parents[parents[node]]
		parents[node] = grandParent
		node = grandParent
	}

	return node
}

func islandUnion(parents []int, ranks []int, node1 int, node2 int, contactCounts []int, jointCounts []int) {
	root1 := islandFindParent(parents, node1)

	root2 := islandFindParent(parents, node2)
	if root1 != root2 {
		if ranks[root1] < ranks[root2] {
			parents[root1] = root2
			contactCounts[root2] += contactCounts[root1]
			jointCounts[root2] += jointCounts[root1]
		} else if ranks[root1] > ranks[root2] {
			parents[root2] = root1
			contactCounts[root1] += contactCounts[root2]
			jointCounts[root1] += jointCounts[root2]
		} else {
			parents[root2] = root1
			ranks[root1] += 1
			contactCounts[root1] += contactCounts[root2]
			jointCounts[root1] += jointCounts[root2]
		}
	}
}

func splitIsland(world *World, baseID int) {
	baseIsland := &world.islands[baseID]
	assert(baseIsland.constraintRemoveCount > 0, "baseIsland->constraintRemoveCount > 0")
	assert(baseIsland.setIndex == awakeSet, "baseIsland->setIndex == b2_awakeSet")

	validateIsland(world, baseID)

	// Cache base island fields before createIsland, which may reallocate
	// world.islands and invalidate the baseIsland pointer.
	baseBodyCount := len(baseIsland.bodies)
	baseBodyIDs := baseIsland.bodies

	baseContactCount := len(baseIsland.contacts)
	baseContacts := baseIsland.contacts

	baseJointCount := len(baseIsland.joints)
	baseJoints := baseIsland.joints

	parents := make([]int, baseBodyCount)
	contactCounts := make([]int, baseBodyCount)
	jointCounts := make([]int, baseBodyCount)

	ranks := make([]int, baseBodyCount)
	for i := range baseBodyCount {
		parents[i] = i
		ranks[i] = 0
		contactCounts[i] = 0
		jointCounts[i] = 0
	}

	bodies := world.bodies

	// Union over contacts, tracking per-component contact counts
	for i := range baseContactCount {
		bodyIDA := baseContacts[i].bodyIDA
		bodyIDB := baseContacts[i].bodyIDB

		validate(0 <= bodyIDA && bodyIDA < len(world.bodies), "0 <= bodyIdA && bodyIdA < world->bodies.count")
		validate(0 <= bodyIDB && bodyIDB < len(world.bodies), "0 <= bodyIdB && bodyIdB < world->bodies.count")
		bodyA := &bodies[bodyIDA]
		bodyB := &bodies[bodyIDB]
		islandIndexA := bodyA.islandIndex
		islandIndexB := bodyB.islandIndex

		// Only connect non-static bodies
		if islandIndexA != nullIndex && islandIndexB != nullIndex {
			validate(0 <= islandIndexA && islandIndexA < baseBodyCount, "0 <= islandIndexA && islandIndexA < baseBodyCount")
			validate(0 <= islandIndexB && islandIndexB < baseBodyCount, "0 <= islandIndexB && islandIndexB < baseBodyCount")
			islandUnion(parents, ranks, islandIndexA, islandIndexB, contactCounts, jointCounts)
			root := islandFindParent(parents, islandIndexA)
			contactCounts[root] += 1
		} else {
			islandIndex := islandIndexB
			if islandIndexA != nullIndex {
				islandIndex = islandIndexA
			}

			root := islandFindParent(parents, islandIndex)
			contactCounts[root] += 1
		}
	}

	// Union over joints, tracking per-component joint counts
	for i := range baseJointCount {
		bodyIDA := baseJoints[i].bodyIDA
		bodyIDB := baseJoints[i].bodyIDB

		validate(0 <= bodyIDA && bodyIDA < len(world.bodies), "0 <= bodyIdA && bodyIdA < world->bodies.count")
		validate(0 <= bodyIDB && bodyIDB < len(world.bodies), "0 <= bodyIdB && bodyIdB < world->bodies.count")
		bodyA := &bodies[bodyIDA]
		bodyB := &bodies[bodyIDB]
		islandIndexA := bodyA.islandIndex
		islandIndexB := bodyB.islandIndex

		// Only connect non-static bodies
		if islandIndexA != nullIndex && islandIndexB != nullIndex {
			validate(0 <= islandIndexA && islandIndexA < baseBodyCount, "0 <= islandIndexA && islandIndexA < baseBodyCount")
			validate(0 <= islandIndexB && islandIndexB < baseBodyCount, "0 <= islandIndexB && islandIndexB < baseBodyCount")
			islandUnion(parents, ranks, islandIndexA, islandIndexB, contactCounts, jointCounts)
			root := islandFindParent(parents, islandIndexA)
			jointCounts[root] += 1
		} else {
			islandIndex := islandIndexB
			if islandIndexA != nullIndex {
				islandIndex = islandIndexA
			}

			root := islandFindParent(parents, islandIndex)
			jointCounts[root] += 1
		}
	}

	// Done with ranks
	ranks = nil

	// Flatten all parent indices and count connected components.
	componentCount := 0

	for i := range baseBodyCount {
		parents[i] = islandFindParent(parents, i)
		if parents[i] == i {
			componentCount += 1
		}
	}

	// Early return — island is still fully connected, no split needed.
	if componentCount == 1 {
		baseIsland.constraintRemoveCount = 0
		return
	}

	// Detach body/contact/joint arrays from base island so destroyIsland won't free them
	baseIsland.bodies = nil

	baseIsland.contacts = nil

	baseIsland.joints = nil

	// Null so code below doesn't accidentally use this.
	baseIsland = nil

	// Map from body index to new island index. Only set for root bodies.
	rootMap := make([]int, baseBodyCount)
	for i := range baseBodyCount {
		rootMap[i] = nullIndex
	}

	componentBodyCounts := make([]int, componentCount)
	componentContactCounts := make([]int, componentCount)
	componentJointCounts := make([]int, componentCount)
	islandCount := 0

	// Find the root body for each body and create islands as needed.
	// Extract per-component counts from the root nodes' accumulated counts.
	for i := range baseBodyCount {
		rootIndex := parents[i]
		if rootMap[rootIndex] == nullIndex {
			rootMap[rootIndex] = islandCount
			componentBodyCounts[islandCount] = 0
			componentContactCounts[islandCount] = contactCounts[rootIndex]
			componentJointCounts[islandCount] = jointCounts[rootIndex]
			islandCount += 1
		}

		componentBodyCounts[rootMap[rootIndex]] += 1
	}

	assert(islandCount == componentCount, "islandCount == componentCount")

	// Map from new island index to island id
	islandIDs := make([]int, islandCount)

	// Create new islands and reserve body/contact/joint arrays
	for i := range islandCount {
		// WARNING: this invalidates baseIsland pointer
		newIsland := createIsland(world, awakeSet)
		islandIDs[i] = newIsland.islandID

		// Reserve arrays to avoid wasteful growth and memcpy.
		newIsland.bodies = arrayReserve(newIsland.bodies, componentBodyCounts[i])
		newIsland.contacts = arrayReserve(newIsland.contacts, componentContactCounts[i])
		newIsland.joints = arrayReserve(newIsland.joints, componentJointCounts[i])
	}

	// Assign bodies to new islands
	for i := range baseBodyCount {
		bodyID := baseBodyIDs[i]
		root := islandFindParent(parents, i)
		newIslandID := islandIDs[rootMap[root]]

		body := &world.bodies[bodyID]
		newIsland := &world.islands[newIslandID]

		body.islandID = newIslandID
		body.islandIndex = len(newIsland.bodies)

		// Ensure the array has the correct capacity
		validate(len(newIsland.bodies) < cap(newIsland.bodies), "newIsland->bodies.count < newIsland->bodies.capacity")
		newIsland.bodies = arrayPush(newIsland.bodies, bodyID)
	}

	// Assign contacts to the island of their bodies
	for i := range baseContactCount {
		link := &baseContacts[i]
		contact := &world.contacts[link.contactID]

		// Static bodies don't have an island id.
		bodyA := &world.bodies[link.bodyIDA]
		bodyB := &world.bodies[link.bodyIDB]

		targetIslandID := bodyB.islandID
		if bodyA.islandID != nullIndex {
			targetIslandID = bodyA.islandID
		}

		targetIsland := &world.islands[targetIslandID]
		contact.islandID = targetIslandID
		contact.islandIndex = len(targetIsland.contacts)

		// Ensure the array has the correct capacity
		validate(len(targetIsland.contacts) < cap(targetIsland.contacts), "targetIsland->contacts.count < targetIsland->contacts.capacity")
		targetIsland.contacts = arrayPush(targetIsland.contacts, *link)
	}

	// Assign joints to the island of their bodies
	for i := range baseJointCount {
		link := &baseJoints[i]
		joint := &world.joints[link.jointID]

		// Static bodies don't have an island id.
		bodyA := &world.bodies[link.bodyIDA]
		bodyB := &world.bodies[link.bodyIDB]

		targetIslandID := bodyB.islandID
		if bodyA.islandID != nullIndex {
			targetIslandID = bodyA.islandID
		}

		targetIsland := &world.islands[targetIslandID]
		joint.islandId = targetIslandID
		joint.islandIndex = len(targetIsland.joints)

		// Ensure the array has the correct capacity
		validate(len(targetIsland.joints) < cap(targetIsland.joints), "targetIsland->joints.count < targetIsland->joints.capacity")
		targetIsland.joints = arrayPush(targetIsland.joints, *link)
	}

	// Destroy the base island
	destroyIsland(world, baseID)
}

func splitIslandTask(context *World) {
	ticks := currentTicks()

	world := context

	assert(world.splitIslandID != nullIndex, "world->splitIslandId != B2_NULL_INDEX")

	splitIsland(world, world.splitIslandID)

	world.splitIslandID = nullIndex
	world.profile.SplitIslands += milliseconds(ticks)
}

func validateIsland(world *World, islandID int) {
	if !validationEnabled {
		return
	}

	if islandID == nullIndex {
		return
	}

	island := &world.islands[islandID]
	assert(island.islandID == islandID, "island->islandId == islandId")
	assert(island.setIndex != nullIndex, "island->setIndex != B2_NULL_INDEX")
	{
		assert(len(island.bodies) > 0, "island->bodies.count > 0")
		assert(len(island.bodies) <= getIDCount(&world.bodyIDPool), "island->bodies.count <= b2GetIdCount( &world->bodyIdPool )")

		for i := range len(island.bodies) {
			body := &world.bodies[island.bodies[i]]
			assert(body.islandID == islandID, "body->islandId == islandId")
			assert(body.islandIndex == i, "body->islandIndex == i")
			assert(body.setIndex == island.setIndex, "body->setIndex == island->setIndex")
		}
	}

	if len(island.contacts) > 0 {
		assert(len(island.contacts) <= getIDCount(&world.contactIDPool), "island->contacts.count <= b2GetIdCount( &world->contactIdPool )")

		for i := range len(island.contacts) {
			link := &island.contacts[i]
			contact := &world.contacts[link.contactID]
			assert(contact.setIndex == island.setIndex, "contact->setIndex == island->setIndex")
			assert(contact.islandID == islandID, "contact->islandId == islandId")
			assert(contact.islandIndex == i, "contact->islandIndex == i")
		}
	}

	if len(island.joints) > 0 {
		assert(len(island.joints) <= getIDCount(&world.jointIDPool), "island->joints.count <= b2GetIdCount( &world->jointIdPool )")

		for i := range len(island.joints) {
			link := &island.joints[i]
			joint := &world.joints[link.jointID]
			assert(joint.setIndex == island.setIndex, "joint->setIndex == island->setIndex")
			assert(joint.islandId == islandID, "joint->islandId == islandId")
			assert(joint.islandIndex == i, "joint->islandIndex == i")
		}
	}
}
