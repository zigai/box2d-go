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

func createIsland(world *world, setIndex int) *island {
	todo()
	return nil
}

func destroyIsland(world *world, islandID int) {
	todo()
}

func mergeIslands(world *world, islandIDA int, islandIDB int) int {
	todo()
	return 0
}

func addContactToIsland(world *world, islandID int, contact *contact) {
	todo()
}

func linkContact(world *world, contact *contact) {
	todo()
}

func unlinkContact(world *world, contact *contact) {
	todo()
}

func addJointToIsland(world *world, islandID int, joint *joint) {
	todo()
}

func linkJoint(world *world, joint *joint) {
	todo()
}

func unlinkJoint(world *world, joint *joint) {
	todo()
}

func islandFindParent(parents []int, node int) int {
	todo()
	return 0
}

func islandUnion(parents []int, ranks []int, node1 int, node2 int, contactCounts []int, jointCounts []int) {
	todo()
}

func splitIsland(world *world, baseID int) {
	todo()
}

func splitIslandTask(context *world) {
	todo()
}

func validateIsland(world *world, islandID int) {
	todo()
}
