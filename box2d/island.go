// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"slices"

	"github.com/zigai/box2d-go/internal/array"
	"github.com/zigai/box2d-go/internal/core"
	"github.com/zigai/box2d-go/internal/timer"
)

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
	core.Assert(setIndex == awakeSet || setIndex >= firstSleepingSet, "setIndex == b2_awakeSet || setIndex >= b2_firstSleepingSet")

	islandID := world.islandIDPool.Alloc()

	if islandID == len(world.islands) {
		emptyIsland := island{}
		world.islands = array.Push(world.islands, emptyIsland)
	} else {
		core.Assert(world.islands[islandID].setIndex == nullIndex, "world->islands.data[islandId].setIndex == B2_NULL_INDEX")
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

	set.islandSims, islandSim = array.Emplace(set.islandSims)
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
		core.Assert(0 <= localIndex && localIndex <= lastIndex, "0 <= localIndex && localIndex <= lastIndex")
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

	core.Validate(island.localIndex == nullIndex, "island->localIndex == B2_NULL_INDEX")

	world.islandIDPool.Free(islandID)
}

func mergeIslands(world *World, islandIDA int, islandIDB int) int {
	if islandIDA == islandIDB {
		return islandIDA
	}

	if islandIDA == nullIndex {
		core.Assert(islandIDB != nullIndex, "islandIdB != B2_NULL_INDEX")
		return islandIDB
	}

	if islandIDB == nullIndex {
		core.Assert(islandIDA != nullIndex, "islandIdA != B2_NULL_INDEX")
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
	bigIsland.bodies = slices.Grow(bigIsland.bodies, len(smallIsland.bodies))

	// Move bodies from smaller island to larger island
	for i := range len(smallIsland.bodies) {
		bodyID := smallIsland.bodies[i]
		body := &world.bodies[bodyID]
		core.Validate(body.islandID == smallIsland.islandID, "body->islandId == smallIsland->islandId")
		body.islandID = bigIslandID
		body.islandIndex = len(bigIsland.bodies)
		bigIsland.bodies = array.Push(bigIsland.bodies, bodyID)
	}

	// Migrate contacts from smaller island to larger island
	if len(smallIsland.contacts) > 0 {
		bigIsland.contacts = slices.Grow(bigIsland.contacts, len(smallIsland.contacts))

		for i := range len(smallIsland.contacts) {
			link := &smallIsland.contacts[i]
			contact := &world.contacts[link.contactID]
			contact.islandID = bigIslandID
			contact.islandIndex = len(bigIsland.contacts)
			bigIsland.contacts = array.Push(bigIsland.contacts, *link)
		}
	}

	// Migrate joints from smaller island to larger island
	if len(smallIsland.joints) > 0 {
		bigIsland.joints = slices.Grow(bigIsland.joints, len(smallIsland.joints))

		for i := range len(smallIsland.joints) {
			link := &smallIsland.joints[i]
			joint := &world.joints[link.jointID]
			joint.islandID = bigIslandID
			joint.islandIndex = len(bigIsland.joints)
			bigIsland.joints = array.Push(bigIsland.joints, *link)
		}
	}

	// Track removed constraints
	bigIsland.constraintRemoveCount += smallIsland.constraintRemoveCount

	destroyIsland(world, smallIsland.islandID)

	validateIsland(world, bigIslandID)

	return bigIslandID
}

func addContactToIsland(world *World, islandID int, contact *contact) {
	core.Assert(contact.islandID == nullIndex, "contact->islandId == B2_NULL_INDEX")
	core.Assert(contact.islandIndex == nullIndex, "contact->islandIndex == B2_NULL_INDEX")

	island := &world.islands[islandID]

	contact.islandID = islandID
	contact.islandIndex = len(island.contacts)

	var link contactLink

	link.contactID = contact.contactID
	link.bodyIDA = contact.edges[0].bodyID
	link.bodyIDB = contact.edges[1].bodyID
	island.contacts = array.Push(island.contacts, link)

	validateIsland(world, islandID)
}

func linkContact(world *World, contact *contact) {
	core.Assert((contact.flags&contactTouchingFlag) != 0, "( contact->flags & b2_contactTouchingFlag ) != 0")

	bodyIDA := contact.edges[0].bodyID
	bodyIDB := contact.edges[1].bodyID

	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	core.Assert(bodyA.setIndex != disabledSet && bodyB.setIndex != disabledSet, "bodyA->setIndex != b2_disabledSet && bodyB->setIndex != b2_disabledSet")
	core.Assert(bodyA.setIndex != staticSet || bodyB.setIndex != staticSet, "bodyA->setIndex != b2_staticSet || bodyB->setIndex != b2_staticSet")

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
	core.Assert(bodyA.setIndex != staticSet || islandIDA == nullIndex, "bodyA->setIndex != b2_staticSet || islandIdA == B2_NULL_INDEX")
	core.Assert(bodyB.setIndex != staticSet || islandIDB == nullIndex, "bodyB->setIndex != b2_staticSet || islandIdB == B2_NULL_INDEX")
	core.Assert(islandIDA != nullIndex || islandIDB != nullIndex, "islandIdA != B2_NULL_INDEX || islandIdB != B2_NULL_INDEX")

	// Merge islands. This will destroy one of the islands.
	finalIslandID := mergeIslands(world, islandIDA, islandIDB)

	// Add contact to the island that survived
	addContactToIsland(world, finalIslandID, contact)
}

func unlinkContact(world *World, contact *contact) {
	core.Assert(contact.islandID != nullIndex, "contact->islandId != B2_NULL_INDEX")

	// remove from island
	islandID := contact.islandID
	island := &world.islands[islandID]

	removeIndex := contact.islandIndex
	core.Assert(0 <= removeIndex && removeIndex < len(island.contacts), "0 <= removeIndex && removeIndex < island->contacts.count")
	core.Assert(island.contacts[removeIndex].contactID == contact.contactID, "island->contacts.data[removeIndex].contactId == contact->contactId")

	var movedIndex int

	island.contacts, movedIndex = array.RemoveSwap(island.contacts, removeIndex)
	if movedIndex != nullIndex {
		// Fix islandIndex on the contact that was swapped into removeIndex
		movedLink := &island.contacts[removeIndex]
		movedContact := &world.contacts[movedLink.contactID]
		core.Assert(movedContact.islandIndex == movedIndex, "movedContact->islandIndex == movedIndex")
		movedContact.islandIndex = removeIndex
	}

	contact.islandID = nullIndex
	contact.islandIndex = nullIndex
	island.constraintRemoveCount += 1

	validateIsland(world, islandID)
}

func addJointToIsland(world *World, islandID int, joint *joint) {
	core.Assert(joint.islandID == nullIndex, "joint->islandId == B2_NULL_INDEX")
	core.Assert(joint.islandIndex == nullIndex, "joint->islandIndex == B2_NULL_INDEX")

	island := &world.islands[islandID]

	joint.islandID = islandID
	joint.islandIndex = len(island.joints)

	var link jointLink

	link.jointID = joint.jointID
	link.bodyIDA = joint.edges[0].bodyID
	link.bodyIDB = joint.edges[1].bodyID
	island.joints = array.Push(island.joints, link)

	validateIsland(world, islandID)
}

func linkJoint(world *World, joint *joint) {
	bodyA := &world.bodies[joint.edges[0].bodyID]
	bodyB := &world.bodies[joint.edges[1].bodyID]

	core.Assert(bodyA.bodyType == DynamicBody || bodyB.bodyType == DynamicBody, "bodyA->type == b2_dynamicBody || bodyB->type == b2_dynamicBody")

	if bodyA.setIndex == awakeSet && bodyB.setIndex >= firstSleepingSet {
		wakeSolverSet(world, bodyB.setIndex)
	} else if bodyB.setIndex == awakeSet && bodyA.setIndex >= firstSleepingSet {
		wakeSolverSet(world, bodyA.setIndex)
	}

	islandIDA := bodyA.islandID
	islandIDB := bodyB.islandID

	core.Assert(islandIDA != nullIndex || islandIDB != nullIndex, "islandIdA != B2_NULL_INDEX || islandIdB != B2_NULL_INDEX")

	// Merge islands. This will destroy one of the islands.
	finalIslandID := mergeIslands(world, islandIDA, islandIDB)

	// Add joint the island that survived
	addJointToIsland(world, finalIslandID, joint)
}

func unlinkJoint(world *World, joint *joint) {
	if joint.islandID == nullIndex {
		return
	}

	// remove from island
	islandID := joint.islandID
	island := &world.islands[islandID]

	removeIndex := joint.islandIndex
	core.Assert(0 <= removeIndex && removeIndex < len(island.joints), "0 <= removeIndex && removeIndex < island->joints.count")
	core.Assert(island.joints[removeIndex].jointID == joint.jointID, "island->joints.data[removeIndex].jointId == joint->jointId")

	var movedIndex int

	island.joints, movedIndex = array.RemoveSwap(island.joints, removeIndex)
	if movedIndex != nullIndex {
		// Fix islandIndex on the joint that was swapped into removeIndex
		movedLink := &island.joints[removeIndex]
		movedJoint := &world.joints[movedLink.jointID]
		core.Assert(movedJoint.islandIndex == movedIndex, "movedJoint->islandIndex == movedIndex")
		movedJoint.islandIndex = removeIndex
	}

	joint.islandID = nullIndex
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
		switch {
		case ranks[root1] < ranks[root2]:
			parents[root1] = root2
			contactCounts[root2] += contactCounts[root1]
			jointCounts[root2] += jointCounts[root1]
		case ranks[root1] > ranks[root2]:
			parents[root2] = root1
			contactCounts[root1] += contactCounts[root2]
			jointCounts[root1] += jointCounts[root2]
		default:
			parents[root2] = root1
			ranks[root1] += 1
			contactCounts[root1] += contactCounts[root2]
			jointCounts[root1] += jointCounts[root2]
		}
	}
}

func splitIsland(world *World, baseID int) {
	baseIsland := &world.islands[baseID]
	core.Assert(baseIsland.constraintRemoveCount > 0, "baseIsland->constraintRemoveCount > 0")
	core.Assert(baseIsland.setIndex == awakeSet, "baseIsland->setIndex == b2_awakeSet")

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

	// Union over contacts, tracking per-component contact counts
	for i := range baseContactCount {
		bodyIDA := baseContacts[i].bodyIDA
		bodyIDB := baseContacts[i].bodyIDB

		root := unionIslandBodies(world, parents, ranks, contactCounts, jointCounts, bodyIDA, bodyIDB)
		contactCounts[root] += 1
	}

	// Union over joints, tracking per-component joint counts
	for i := range baseJointCount {
		bodyIDA := baseJoints[i].bodyIDA
		bodyIDB := baseJoints[i].bodyIDB

		root := unionIslandBodies(world, parents, ranks, contactCounts, jointCounts, bodyIDA, bodyIDB)
		jointCounts[root] += 1
	}

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

	core.Assert(islandCount == componentCount, "islandCount == componentCount")

	// Map from new island index to island id
	islandIDs := make([]int, islandCount)

	// Create new islands and reserve body/contact/joint arrays
	for i := range islandCount {
		// WARNING: this invalidates baseIsland pointer
		newIsland := createIsland(world, awakeSet)
		islandIDs[i] = newIsland.islandID

		// Reserve arrays to avoid wasteful growth and memcpy.
		newIsland.bodies = array.Reserve(newIsland.bodies, componentBodyCounts[i])
		newIsland.contacts = array.Reserve(newIsland.contacts, componentContactCounts[i])
		newIsland.joints = array.Reserve(newIsland.joints, componentJointCounts[i])
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
		core.Validate(len(newIsland.bodies) < cap(newIsland.bodies), "newIsland->bodies.count < newIsland->bodies.capacity")
		newIsland.bodies = array.Push(newIsland.bodies, bodyID)
	}

	reassignSplitIslandContacts(world, baseContacts)

	reassignSplitIslandJoints(world, baseJoints)

	// Destroy the base island
	destroyIsland(world, baseID)
}

func unionIslandBodies(world *World, parents, ranks, contactCounts, jointCounts []int, bodyIDA, bodyIDB int) int {
	baseBodyCount := len(parents)
	bodies := world.bodies
	core.Validate(0 <= bodyIDA && bodyIDA < len(world.bodies), "0 <= bodyIdA && bodyIdA < world->bodies.count")
	core.Validate(0 <= bodyIDB && bodyIDB < len(world.bodies), "0 <= bodyIdB && bodyIdB < world->bodies.count")
	bodyA := &bodies[bodyIDA]
	bodyB := &bodies[bodyIDB]
	islandIndexA := bodyA.islandIndex
	islandIndexB := bodyB.islandIndex

	// Only connect non-static bodies
	if islandIndexA != nullIndex && islandIndexB != nullIndex {
		core.Validate(0 <= islandIndexA && islandIndexA < baseBodyCount, "0 <= islandIndexA && islandIndexA < baseBodyCount")
		core.Validate(0 <= islandIndexB && islandIndexB < baseBodyCount, "0 <= islandIndexB && islandIndexB < baseBodyCount")
		islandUnion(parents, ranks, islandIndexA, islandIndexB, contactCounts, jointCounts)

		return islandFindParent(parents, islandIndexA)
	}

	islandIndex := islandIndexB
	if islandIndexA != nullIndex {
		islandIndex = islandIndexA
	}

	return islandFindParent(parents, islandIndex)
}

func splitIslandTask(context *World) {
	ticks := timer.Ticks()

	world := context

	core.Assert(world.splitIslandID != nullIndex, "world->splitIslandId != B2_NULL_INDEX")

	splitIsland(world, world.splitIslandID)

	world.splitIslandID = nullIndex
	world.profile.SplitIslands += timer.Milliseconds(ticks)
}

func validateIsland(world *World, islandID int) {
	if !core.ValidationEnabled {
		return
	}

	if islandID == nullIndex {
		return
	}

	island := &world.islands[islandID]
	core.Assert(island.islandID == islandID, "island->islandId == islandId")
	core.Assert(island.setIndex != nullIndex, "island->setIndex != B2_NULL_INDEX")
	{
		core.Assert(len(island.bodies) > 0, "island->bodies.count > 0")
		core.Assert(len(island.bodies) <= world.bodyIDPool.Count(), "island->bodies.count <= b2GetIdCount( &world->bodyIdPool )")

		for i := range len(island.bodies) {
			body := &world.bodies[island.bodies[i]]
			core.Assert(body.islandID == islandID, "body->islandId == islandId")
			core.Assert(body.islandIndex == i, "body->islandIndex == i")
			core.Assert(body.setIndex == island.setIndex, "body->setIndex == island->setIndex")
		}
	}

	if len(island.contacts) > 0 {
		core.Assert(len(island.contacts) <= world.contactIDPool.Count(), "island->contacts.count <= b2GetIdCount( &world->contactIdPool )")

		for i := range len(island.contacts) {
			link := &island.contacts[i]
			contact := &world.contacts[link.contactID]
			core.Assert(contact.setIndex == island.setIndex, "contact->setIndex == island->setIndex")
			core.Assert(contact.islandID == islandID, "contact->islandId == islandId")
			core.Assert(contact.islandIndex == i, "contact->islandIndex == i")
		}
	}

	if len(island.joints) > 0 {
		core.Assert(len(island.joints) <= world.jointIDPool.Count(), "island->joints.count <= b2GetIdCount( &world->jointIdPool )")

		for i := range len(island.joints) {
			link := &island.joints[i]
			joint := &world.joints[link.jointID]
			core.Assert(joint.setIndex == island.setIndex, "joint->setIndex == island->setIndex")
			core.Assert(joint.islandID == islandID, "joint->islandId == islandId")
			core.Assert(joint.islandIndex == i, "joint->islandIndex == i")
		}
	}
}

func reassignSplitIslandContacts(world *World, baseContacts []contactLink) {
	baseContactCount := len(baseContacts)
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
		core.Validate(len(targetIsland.contacts) < cap(targetIsland.contacts), "targetIsland->contacts.count < targetIsland->contacts.capacity")
		targetIsland.contacts = array.Push(targetIsland.contacts, *link)
	}
}

func reassignSplitIslandJoints(world *World, baseJoints []jointLink) {
	baseJointCount := len(baseJoints)
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
		joint.islandID = targetIslandID
		joint.islandIndex = len(targetIsland.joints)

		// Ensure the array has the correct capacity
		core.Validate(len(targetIsland.joints) < cap(targetIsland.joints), "targetIsland->joints.count < targetIsland->joints.capacity")
		targetIsland.joints = array.Push(targetIsland.joints, *link)
	}
}
