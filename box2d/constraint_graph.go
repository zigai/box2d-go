// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"github.com/zigai/box2d-go/internal/array"
	"github.com/zigai/box2d-go/internal/bitset"
	"github.com/zigai/box2d-go/internal/core"
)

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

type graphColor struct {
	bodySet             bitset.Set
	contactSims         []contactSim
	jointSims           []jointSim
	wideConstraints     []contactConstraintWide
	overflowConstraints []contactConstraint
	wideConstraintCount int
}

type constraintGraph struct {
	colors [GraphColorCount]graphColor
}

// GraphColor returns the visualization color assigned to a constraint graph color slot. The last index
// (GraphColorCount - 1) is the overflow color.
func GraphColor(index int) HexColor {
	core.Assert(0 <= index && index < GraphColorCount, "0 <= index && index < B2_GRAPH_COLOR_COUNT")
	return graphColors[index]
}

func createGraph(graph *constraintGraph, capacity *Capacity) {
	*graph = constraintGraph{}

	bodyCapacity := max(capacity.StaticBodyCount+capacity.DynamicBodyCount, 16)

	// Initialize graph color bit set.
	// No bitset for overflow color.
	for i := range overflowIndex {
		color := &graph.colors[i]
		color.bodySet = bitset.New(uint32(bodyCapacity))
		color.bodySet.SetBitCountAndClear(uint32(bodyCapacity))
		color.contactSims = array.Reserve(color.contactSims, 16)
	}
}

func destroyGraph(graph *constraintGraph) {
	for i := range GraphColorCount {
		color := &graph.colors[i]

		// The bit set should never be used on the overflow color
		core.Assert(i != overflowIndex || color.bodySet.Bits == nil, "i != B2_OVERFLOW_INDEX || color->bodySet.bits == NULL")

		color.bodySet.Destroy()
		color.contactSims = nil
		color.jointSims = nil
	}
}

func addContactToGraph(world *World, contactSim *contactSim, contact *contact) {
	core.Assert(contactSim.manifold.PointCount > 0, "contactSim->manifold.pointCount > 0")
	core.Assert((contactSim.simFlags&simTouchingFlag) != 0, "contactSim->simFlags & b2_simTouchingFlag")
	core.Assert((contact.flags&contactTouchingFlag) != 0, "contact->flags & b2_contactTouchingFlag")

	graph := &world.constraintGraph

	bodyIDA := contact.edges[0].bodyID
	bodyIDB := contact.edges[1].bodyID
	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	typeA := bodyA.bodyType
	typeB := bodyB.bodyType
	core.Assert(typeA == DynamicBody || typeB == DynamicBody, "typeA == b2_dynamicBody || typeB == b2_dynamicBody")

	colorIndex := assignConstraintColor(graph, bodyIDA, bodyIDB, typeA, typeB)
	color := &graph.colors[colorIndex]
	contact.colorIndex = colorIndex
	contact.localIndex = len(color.contactSims)

	color.contactSims, _ = array.Emplace(color.contactSims)
	newContact := &color.contactSims[len(color.contactSims)-1]
	*newContact = *contactSim

	// todo perhaps skip this if the contact is already awake

	if typeA == StaticBody {
		newContact.encodedBodySimA = encodeBodySimIndex(bodyA)
		newContact.invMassA = 0.0
		newContact.invIA = 0.0
	} else {
		core.Assert(bodyA.setIndex == awakeSet, "bodyA->setIndex == b2_awakeSet")
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
		core.Assert(bodyB.setIndex == awakeSet, "bodyB->setIndex == b2_awakeSet")
		awakeSet := &world.solverSets[awakeSet]

		localIndex := bodyB.localIndex
		newContact.encodedBodySimB = localIndex

		bodySimB := &awakeSet.bodySims[localIndex]
		newContact.invMassB = bodySimB.invMass
		newContact.invIB = bodySimB.invInertia
	}
}

func removeContactFromGraph(world *World, bodyIDA int, bodyIDB int, colorIndex int, localIndex int) {
	graph := &world.constraintGraph

	core.Assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")
	color := &graph.colors[colorIndex]

	if colorIndex != overflowIndex {
		// This might clear a bit for a kinematic or static body, but this has no effect
		color.bodySet.ClearBit(uint32(bodyIDA))
		color.bodySet.ClearBit(uint32(bodyIDB))
	}

	var movedIndex int

	color.contactSims, movedIndex = array.RemoveSwap(color.contactSims, localIndex)
	if movedIndex != nullIndex {
		// Fix index on swapped contact
		movedContactSim := &color.contactSims[localIndex]

		// Fix moved contact
		movedID := movedContactSim.contactID
		movedContact := &world.contacts[movedID]
		core.Assert(movedContact.setIndex == awakeSet, "movedContact->setIndex == b2_awakeSet")
		core.Assert(movedContact.colorIndex == colorIndex, "movedContact->colorIndex == colorIndex")
		core.Assert(movedContact.localIndex == movedIndex, "movedContact->localIndex == movedIndex")
		movedContact.localIndex = localIndex
	}
}

func assignConstraintColor(graph *constraintGraph, bodyIDA int, bodyIDB int, typeA BodyType, typeB BodyType) int {
	core.Assert(typeA == DynamicBody || typeB == DynamicBody, "typeA == b2_dynamicBody || typeB == b2_dynamicBody")

	if typeA == DynamicBody && typeB == DynamicBody {
		// Dynamic constraint colors cannot encroach on colors reserved for static constraints
		for i := range dynamicColorCount {
			color := &graph.colors[i]
			if color.bodySet.GetBit(uint32(bodyIDA)) || color.bodySet.GetBit(uint32(bodyIDB)) {
				continue
			}

			color.bodySet.SetBitGrow(uint32(bodyIDA))
			color.bodySet.SetBitGrow(uint32(bodyIDB))

			return i
		}

		return overflowIndex
	}

	bodyID := bodyIDB
	if typeA == DynamicBody {
		bodyID = bodyIDA
	} else if typeB != DynamicBody {
		return overflowIndex
	}

	// Static constraint colors build from the end to get higher priority than dyn-dyn constraints
	for i := overflowIndex - 1; i >= 1; i-- {
		color := &graph.colors[i]
		if color.bodySet.GetBit(uint32(bodyID)) {
			continue
		}

		color.bodySet.SetBitGrow(uint32(bodyID))

		return i
	}

	return overflowIndex
}

func createJointInGraph(world *World, joint *joint) *jointSim {
	graph := &world.constraintGraph

	bodyIDA := joint.edges[0].bodyID
	bodyIDB := joint.edges[1].bodyID
	bodyA := &world.bodies[bodyIDA]
	bodyB := &world.bodies[bodyIDB]

	colorIndex := assignConstraintColor(graph, bodyIDA, bodyIDB, bodyA.bodyType, bodyB.bodyType)

	graph.colors[colorIndex].jointSims, _ = array.Emplace(graph.colors[colorIndex].jointSims)
	jointSim := &graph.colors[colorIndex].jointSims[len(graph.colors[colorIndex].jointSims)-1]

	joint.colorIndex = colorIndex
	joint.localIndex = len(graph.colors[colorIndex].jointSims) - 1

	return jointSim
}

func addJointToGraph(world *World, jointSim *jointSim, joint *joint) {
	jointDst := createJointInGraph(world, joint)
	*jointDst = *jointSim
}

func removeJointFromGraph(world *World, bodyIDA int, bodyIDB int, colorIndex int, localIndex int) {
	graph := &world.constraintGraph

	core.Assert(0 <= colorIndex && colorIndex < GraphColorCount, "0 <= colorIndex && colorIndex < B2_GRAPH_COLOR_COUNT")
	color := &graph.colors[colorIndex]

	if colorIndex != overflowIndex {
		// May clear static bodies, no effect
		color.bodySet.ClearBit(uint32(bodyIDA))
		color.bodySet.ClearBit(uint32(bodyIDB))
	}

	var movedIndex int

	color.jointSims, movedIndex = array.RemoveSwap(color.jointSims, localIndex)
	if movedIndex != nullIndex {
		// Fix moved joint
		movedJointSim := &color.jointSims[localIndex]
		movedID := movedJointSim.jointID
		movedJoint := &world.joints[movedID]
		core.Assert(movedJoint.setIndex == awakeSet, "movedJoint->setIndex == b2_awakeSet")
		core.Assert(movedJoint.colorIndex == colorIndex, "movedJoint->colorIndex == colorIndex")
		core.Assert(movedJoint.localIndex == movedIndex, "movedJoint->localIndex == movedIndex")
		movedJoint.localIndex = localIndex
	}
}
