// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type graphColor struct {
	bodySet             bitSet
	contactSims         []contactSim
	jointSims           []jointSim
	wideConstraints     []contactConstraintWide
	overflowConstraints []contactConstraint
	wideConstraintCount int
}

type constraintGraph struct {
	colors [GraphColorCount]graphColor
}

const (
	overflowIndex     = GraphColorCount - 1
	dynamicColorCount = GraphColorCount - 4
	forceOverflow     = 0
)

var graphColors = [GraphColorCount]HexColor{
	ColorRed, ColorOrange, ColorYellow, ColorLimeGreen, ColorSpringGreen,
	ColorAqua, ColorDodgerBlue, ColorBlueViolet, ColorMagenta, ColorDeepPink,
	ColorCrimson, ColorCoral, ColorGold, ColorGreenYellow, ColorMediumSeaGreen,
	ColorTurquoise, ColorDeepSkyBlue, ColorCornflowerBlue, ColorMediumSlateBlue, ColorMediumOrchid,
	ColorHotPink, ColorTomato, ColorKhaki, ColorSilver,
}

func createGraph(graph *constraintGraph, capacity *Capacity) {
	todo()
}

func destroyGraph(graph *constraintGraph) {
	todo()
}

func addContactToGraph(world *world, contactSim *contactSim, contact *contact) {
	todo()
}

func removeContactFromGraph(world *world, bodyIDA int, bodyIDB int, colorIndex int, localIndex int) {
	todo()
}

func assignJointColor(graph *constraintGraph, bodyIDA int, bodyIDB int, typeA BodyType, typeB BodyType) int {
	todo()
	return 0
}

func createJointInGraph(world *world, joint *joint) *jointSim {
	todo()
	return nil
}

func addJointToGraph(world *world, jointSim *jointSim, joint *joint) {
	todo()
}

func removeJointFromGraph(world *world, bodyIDA int, bodyIDB int, colorIndex int, localIndex int) {
	todo()
}

// Get the visualization color assigned to a constraint graph color slot. The last index
// (GraphColorCount - 1) is the overflow color.
func GetGraphColor(index int) HexColor {
	todo()
	return 0
}
