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
	*graph = constraintGraph{}

	bodyCapacity := max(capacity.StaticBodyCount+capacity.DynamicBodyCount, 16)

	// Initialize graph color bit set.
	// No bitset for overflow color.
	for i := 0; i < overflowIndex; i++ {
		color := &graph.colors[i]
		color.bodySet = createBitSet(uint32(bodyCapacity))
		setBitCountAndClear(&color.bodySet, uint32(bodyCapacity))
		color.contactSims = arrayReserve(color.contactSims, 16)
	}
}

func destroyGraph(graph *constraintGraph) {
	for i := 0; i < GraphColorCount; i++ {
		color := &graph.colors[i]

		// The bit set should never be used on the overflow color
		assert(i != overflowIndex || color.bodySet.bits == nil, "i != B2_OVERFLOW_INDEX || color->bodySet.bits == NULL")

		destroyBitSet(&color.bodySet)
		color.contactSims = nil
		color.jointSims = nil
	}
}

func addContactToGraph(world *world, contactSim *contactSim, contact *contact) {
	assert(contactSim.manifold.PointCount > 0, "contactSim->manifold.pointCount > 0")
	assert((contactSim.simFlags&simTouchingFlag) != 0, "contactSim->simFlags & b2_simTouchingFlag")
	assert((contact.flags&contactTouchingFlag) != 0, "contact->flags & b2_contactTouchingFlag")

	graph := &world.constraintGraph
	colorIndex := overflowIndex

	bodyIDA := contact.edges[0].bodyID
	bodyIDB := contact.edges[1].bodyID
	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	typeA := bodyA.bodyType
	typeB := bodyB.bodyType
	assert(typeA == DynamicBody || typeB == DynamicBody, "typeA == b2_dynamicBody || typeB == b2_dynamicBody")

	if typeA == DynamicBody && typeB == DynamicBody {
		// Dynamic constraint colors cannot encroach on colors reserved for static constraints
		for i := 0; i < dynamicColorCount; i++ {
			color := &graph.colors[i]
			if getBit(&color.bodySet, uint32(bodyIDA)) || getBit(&color.bodySet, uint32(bodyIDB)) {
				continue
			}

			setBitGrow(&color.bodySet, uint32(bodyIDA))
			setBitGrow(&color.bodySet, uint32(bodyIDB))
			colorIndex = i
			break
		}
	} else if typeA == DynamicBody {
		// Static constraint colors build from the end to get higher priority than dyn-dyn constraints
		for i := overflowIndex - 1; i >= 1; i-- {
			color := &graph.colors[i]
			if getBit(&color.bodySet, uint32(bodyIDA)) {
				continue
			}

			setBitGrow(&color.bodySet, uint32(bodyIDA))
			colorIndex = i
			break
		}
	} else if typeB == DynamicBody {
		// Static constraint colors build from the end to get higher priority than dyn-dyn constraints
		for i := overflowIndex - 1; i >= 1; i-- {
			color := &graph.colors[i]
			if getBit(&color.bodySet, uint32(bodyIDB)) {
				continue
			}

			setBitGrow(&color.bodySet, uint32(bodyIDB))
			colorIndex = i
			break
		}
	}

	color := &graph.colors[colorIndex]
	contact.colorIndex = colorIndex
	contact.localIndex = len(color.contactSims)

	color.contactSims, _ = arrayEmplace(color.contactSims)
	newContact := &color.contactSims[len(color.contactSims)-1]
	*newContact = *contactSim

	// todo perhaps skip this if the contact is already awake

	if typeA == StaticBody {
		newContact.encodedBodySimA = encodeBodySimIndex(bodyA)
		newContact.invMassA = 0.0
		newContact.invIA = 0.0
	} else {
		assert(bodyA.setIndex == awakeSet, "bodyA->setIndex == b2_awakeSet")
		awakeSet := &world.solverSets[awakeSet]

		localIndex := bodyA.localIndex
		newContact.encodedBodySimA = localIndex

		bodySimA := &awakeSet.bodySims[localIndex]
		newContact.invMassA = bodySimA.invMass
		newContact.invIA = bodySimA.invInertia
	}

	if typeB == StaticBody {
		newContact.encodedBodySimB = encodeBodySimIndex(bodyB)
		newContact.invMassB = 0.0
		newContact.invIB = 0.0
	} else {
		assert(bodyB.setIndex == awakeSet, "bodyB->setIndex == b2_awakeSet")
		awakeSet := &world.solverSets[awakeSet]

		localIndex := bodyB.localIndex
		newContact.encodedBodySimB = localIndex

		bodySimB := &awakeSet.bodySims[localIndex]
		newContact.invMassB = bodySimB.invMass
		newContact.invIB = bodySimB.invInertia
	}
}

func removeContactFromGraph(world *world, bodyIDA int, bodyIDB int, colorIndex int, localIndex int) {
	graph := &world.constraintGraph

	assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")
	color := &graph.colors[colorIndex]

	if colorIndex != overflowIndex {
		// This might clear a bit for a kinematic or static body, but this has no effect
		clearBit(&color.bodySet, uint32(bodyIDA))
		clearBit(&color.bodySet, uint32(bodyIDB))
	}

	var movedIndex int
	color.contactSims, movedIndex = arrayRemoveSwap(color.contactSims, localIndex)
	if movedIndex != nullIndex {
		// Fix index on swapped contact
		movedContactSim := &color.contactSims[localIndex]

		// Fix moved contact
		movedId := movedContactSim.contactID
		movedContact := &world.contacts[movedId]
		assert(movedContact.setIndex == awakeSet, "movedContact->setIndex == b2_awakeSet")
		assert(movedContact.colorIndex == colorIndex, "movedContact->colorIndex == colorIndex")
		assert(movedContact.localIndex == movedIndex, "movedContact->localIndex == movedIndex")
		movedContact.localIndex = localIndex
	}
}

func assignJointColor(graph *constraintGraph, bodyIDA int, bodyIDB int, typeA BodyType, typeB BodyType) int {
	assert(typeA == DynamicBody || typeB == DynamicBody, "typeA == b2_dynamicBody || typeB == b2_dynamicBody")

	if typeA == DynamicBody && typeB == DynamicBody {
		// Dynamic constraint colors cannot encroach on colors reserved for static constraints
		for i := 0; i < dynamicColorCount; i++ {
			color := &graph.colors[i]
			if getBit(&color.bodySet, uint32(bodyIDA)) || getBit(&color.bodySet, uint32(bodyIDB)) {
				continue
			}

			setBitGrow(&color.bodySet, uint32(bodyIDA))
			setBitGrow(&color.bodySet, uint32(bodyIDB))
			return i
		}
	} else if typeA == DynamicBody {
		// Static constraint colors build from the end to get higher priority than dyn-dyn constraints
		for i := overflowIndex - 1; i >= 1; i-- {
			color := &graph.colors[i]
			if getBit(&color.bodySet, uint32(bodyIDA)) {
				continue
			}

			setBitGrow(&color.bodySet, uint32(bodyIDA))
			return i
		}
	} else if typeB == DynamicBody {
		// Static constraint colors build from the end to get higher priority than dyn-dyn constraints
		for i := overflowIndex - 1; i >= 1; i-- {
			color := &graph.colors[i]
			if getBit(&color.bodySet, uint32(bodyIDB)) {
				continue
			}

			setBitGrow(&color.bodySet, uint32(bodyIDB))
			return i
		}
	}

	return overflowIndex
}

func createJointInGraph(world *world, joint *joint) *jointSim {
	graph := &world.constraintGraph

	bodyIDA := joint.edges[0].bodyId
	bodyIDB := joint.edges[1].bodyId
	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	colorIndex := assignJointColor(graph, bodyIDA, bodyIDB, bodyA.bodyType, bodyB.bodyType)

	graph.colors[colorIndex].jointSims, _ = arrayEmplace(graph.colors[colorIndex].jointSims)
	jointSim := &graph.colors[colorIndex].jointSims[len(graph.colors[colorIndex].jointSims)-1]

	joint.colorIndex = colorIndex
	joint.localIndex = len(graph.colors[colorIndex].jointSims) - 1
	return jointSim
}

func addJointToGraph(world *world, jointSim *jointSim, joint *joint) {
	jointDst := createJointInGraph(world, joint)
	*jointDst = *jointSim
}

func removeJointFromGraph(world *world, bodyIDA int, bodyIDB int, colorIndex int, localIndex int) {
	graph := &world.constraintGraph

	assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")
	color := &graph.colors[colorIndex]

	if colorIndex != overflowIndex {
		// May clear static bodies, no effect
		clearBit(&color.bodySet, uint32(bodyIDA))
		clearBit(&color.bodySet, uint32(bodyIDB))
	}

	var movedIndex int
	color.jointSims, movedIndex = arrayRemoveSwap(color.jointSims, localIndex)
	if movedIndex != nullIndex {
		// Fix moved joint
		movedJointSim := &color.jointSims[localIndex]
		movedId := movedJointSim.jointId
		movedJoint := &world.joints[movedId]
		assert(movedJoint.setIndex == awakeSet, "movedJoint->setIndex == b2_awakeSet")
		assert(movedJoint.colorIndex == colorIndex, "movedJoint->colorIndex == colorIndex")
		assert(movedJoint.localIndex == movedIndex, "movedJoint->localIndex == movedIndex")
		movedJoint.localIndex = localIndex
	}
}

// Get the visualization color assigned to a constraint graph color slot. The last index
// (GraphColorCount - 1) is the overflow color.
func GetGraphColor(index int) HexColor {
	assert(0 <= index && index < GraphColorCount, "0 <= index && index < B2_GRAPH_COLOR_COUNT")
	return graphColors[index]
}
