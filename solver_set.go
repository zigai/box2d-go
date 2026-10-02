// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type solverSet struct {
	bodySims    []bodySim
	bodyStates  []bodyState
	jointSims   []jointSim
	contactSims []contactSim
	islandSims  []islandSim
	setIndex    int
}

type solverSetType int

const (
	staticSet        int = 0
	disabledSet      int = 1
	awakeSet         int = 2
	firstSleepingSet int = 3
)

func destroySolverSet(world *world, setIndex int) {
	todo()
}

func wakeSolverSet(world *world, setIndex int) {
	todo()
}

func trySleepIsland(world *world, islandID int) {
	todo()
}

func mergeSolverSets(world *world, setId1 int, setId2 int) {
	todo()
}

func removeBodySim(world *world, set *solverSet, localIndex int) *body {
	todo()
	return nil
}

func transferBody(world *world, targetSet *solverSet, sourceSet *solverSet, body *body) {
	todo()
}

func transferJoint(world *world, targetSet *solverSet, sourceSet *solverSet, joint *joint) {
	todo()
}
