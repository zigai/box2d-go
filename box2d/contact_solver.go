// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"unsafe"

	"github.com/zigai/box2d-go/internal/bitset"
	"github.com/zigai/box2d-go/internal/core"
)

var zeroContactSim contactSim

type contactConstraintPoint struct {
	anchorA            Vec2
	anchorB            Vec2
	baseSeparation     float32
	relativeVelocity   float32
	normalImpulse      float32
	tangentImpulse     float32
	totalNormalImpulse float32
	restitutionImpulse float32
	normalMass         float32
	tangentMass        float32
}

type contactConstraint struct {
	indexA            int
	indexB            int
	points            [2]contactConstraintPoint
	normal            Vec2
	invMassA          float32
	invMassB          float32
	invIA             float32
	invIB             float32
	friction          float32
	restitution       float32
	tangentSpeed      float32
	rollingResistance float32
	rollingMass       float32
	rollingImpulse    float32
	softness          softness
	pointCount        int
}

type floatW struct {
	x float32
	y float32
	z float32
	w float32
}

type vec2W struct {
	x floatW
	y floatW
}

type rotW struct {
	c floatW
	s floatW
}

type contactConstraintWide struct {
	indexA              [simdWidth]int
	indexB              [simdWidth]int
	invMassA            floatW
	invMassB            floatW
	invIA               floatW
	invIB               floatW
	normal              vec2W
	anchorA1            vec2W
	anchorB1            vec2W
	anchorA2            vec2W
	anchorB2            vec2W
	normalMass1         floatW
	normalMass2         floatW
	baseSeparation1     floatW
	baseSeparation2     floatW
	normalImpulse1      floatW
	normalImpulse2      floatW
	totalNormalImpulse1 floatW
	totalNormalImpulse2 floatW
	tangentImpulse1     floatW
	tangentImpulse2     floatW
	rollingImpulse      floatW
	friction            floatW
	tangentSpeed        floatW
	rollingResistance   floatW
	tangentMass1        floatW
	tangentMass2        floatW
	restitution         floatW
	relativeVelocity1   floatW
	relativeVelocity2   floatW
	restitutionImpulse1 floatW
	restitutionImpulse2 floatW
}

type bodyStateW struct {
	v  vec2W
	w  floatW
	dp vec2W
	dq rotW
}

// contact separation for sub-stepping
// s = s0 + dot(cB + rB - cA - rA, normal)
// normal is held constant
// body positions c can translation and anchors r can rotate
// s(t) = s0 + dot(cB(t) + rB(t) - cA(t) - rA(t), normal)
// s(t) = s0 + dot(cB0 + dpB + rot(dqB, rB0) - cA0 - dpA - rot(dqA, rA0), normal)
// s(t) = s0 + dot(cB0 - cA0, normal) + dot(dpB - dpA + rot(dqB, rB0) - rot(dqA, rA0), normal)
// s_base = s0 + dot(cB0 - cA0, normal)

func prepareContactsOverflow(context *stepContext) {
	world := context.world
	graph := context.graph
	color := &graph.colors[overflowIndex]
	constraints := color.overflowConstraints
	contactCount := len(color.contactSims)
	contacts := color.contactSims

	warmStartScale := float32(0.0)
	if world.enableWarmStarting {
		warmStartScale = 1.0
	}

	anyRestitution := false

	for i := range contactCount {
		contactSim := &contacts[i]

		manifold := &contactSim.manifold
		pointCount := manifold.PointCount

		core.Assert(0 < pointCount && pointCount <= 2, "0 < pointCount && pointCount <= 2")

		indexA := decodeAwakeIndex(contactSim.encodedBodySimA)
		indexB := decodeAwakeIndex(contactSim.encodedBodySimB)

		if core.ValidationEnabled {
			bodyA := &world.bodies[contactSim.bodyIDA]
			core.Assert(contactSim.encodedBodySimA == encodeBodySimIndex(bodyA), "contactSim.encodedBodySimA == encodeBodySimIndex( bodyA )")

			bodyB := &world.bodies[contactSim.bodyIDB]
			core.Assert(contactSim.encodedBodySimB == encodeBodySimIndex(bodyB), "contactSim.encodedBodySimB == encodeBodySimIndex( bodyB )")
		}

		constraint := &constraints[i]

		// 0 is null
		constraint.indexA = indexA + 1
		constraint.indexB = indexB + 1
		constraint.normal = manifold.Normal
		constraint.friction = contactSim.friction
		constraint.restitution = contactSim.restitution
		constraint.rollingResistance = contactSim.rollingResistance
		constraint.rollingImpulse = float32(warmStartScale * manifold.RollingImpulse)
		constraint.tangentSpeed = contactSim.tangentSpeed
		constraint.pointCount = pointCount

		haveRestitution := contactSim.restitution > 0.0
		hitEvents := (contactSim.simFlags & simEnableHitEvent) != 0
		sampleVelocity := haveRestitution || hitEvents
		anyRestitution = anyRestitution || haveRestitution

		prepareOverflowContactDynamics(context, contactSim, constraint, indexA, indexB, warmStartScale, sampleVelocity)
	}

	if anyRestitution {
		// Check if set first? Store per task context and OR?
		context.anyRestitution.Store(1)
	}
}

func warmStartContactsOverflow(context *stepContext) {
	graph := context.graph
	color := &graph.colors[overflowIndex]
	constraints := color.overflowConstraints
	contactCount := len(color.contactSims)
	world := context.world
	awakeSet := &world.solverSets[awakeSet]
	states := awakeSet.bodyStates

	// This is a dummy state to represent a static body because static bodies don't have a solver body.
	dummyState := identityBodyState

	for i := range contactCount {
		constraint := &constraints[i]

		indexA := constraint.indexA - 1
		indexB := constraint.indexB - 1

		stateA := &dummyState
		if indexA != nullIndex {
			stateA = &states[indexA]
		}

		stateB := &dummyState
		if indexB != nullIndex {
			stateB = &states[indexB]
		}

		vA := stateA.linearVelocity
		wA := stateA.angularVelocity
		vB := stateB.linearVelocity
		wB := stateB.angularVelocity

		mA := constraint.invMassA
		iA := constraint.invIA
		mB := constraint.invMassB
		iB := constraint.invIB

		// Stiffer for static contacts to avoid bodies getting pushed through the ground
		normal := constraint.normal
		tangent := constraint.normal.RightPerp()
		pointCount := constraint.pointCount

		for j := range pointCount {
			cp := &constraint.points[j]

			// fixed anchors
			rA := cp.anchorA
			rB := cp.anchorB

			p := normal.Scale(cp.normalImpulse).Add(tangent.Scale(cp.tangentImpulse))

			cp.totalNormalImpulse += cp.normalImpulse

			wA -= float32(iA * rA.Cross(p))
			vA = vA.MulAdd(-mA, p)
			wB += float32(iB * rB.Cross(p))
			vB = vB.MulAdd(mB, p)
		}

		wA -= float32(iA * constraint.rollingImpulse)
		wB += float32(iB * constraint.rollingImpulse)

		if stateA.flags&dynamicFlag != 0 {
			stateA.linearVelocity = vA
			stateA.angularVelocity = wA
		}

		if stateB.flags&dynamicFlag != 0 {
			stateB.linearVelocity = vB
			stateB.angularVelocity = wB
		}
	}
}

// Solve the non-penetration constraints with the soft bias. No friction and no restitution.
func pushContactsOverflow(context *stepContext) {
	graph := context.graph
	color := &graph.colors[overflowIndex]
	constraints := color.overflowConstraints
	contactCount := len(color.contactSims)
	world := context.world
	awakeSet := &world.solverSets[awakeSet]
	states := awakeSet.bodyStates

	invH := context.invH
	contactSpeed := context.world.contactSpeed

	// This is a dummy body to represent a static body since static bodies don't have a solver body.
	dummyState := identityBodyState

	for i := range contactCount {
		constraint := &constraints[i]
		mA := constraint.invMassA
		iA := constraint.invIA
		mB := constraint.invMassB
		iB := constraint.invIB

		indexA := constraint.indexA - 1
		indexB := constraint.indexB - 1

		stateA := &dummyState
		if indexA != nullIndex {
			stateA = &states[indexA]
		}

		vA := stateA.linearVelocity
		wA := stateA.angularVelocity
		dqA := stateA.deltaRotation

		stateB := &dummyState
		if indexB != nullIndex {
			stateB = &states[indexB]
		}

		vB := stateB.linearVelocity
		wB := stateB.angularVelocity
		dqB := stateB.deltaRotation

		dp := stateB.deltaPosition.Sub(stateA.deltaPosition)

		normal := constraint.normal
		softness := constraint.softness
		pointCount := constraint.pointCount

		for j := range pointCount {
			cp := &constraint.points[j]

			// fixed anchor points
			rA := cp.anchorA
			rB := cp.anchorB

			// compute current separation
			// this is subject to round-off error if the anchor is far from the body center of mass
			ds := dp.Add(dqB.Rotate(rB).Sub(dqA.Rotate(rA)))
			s := cp.baseSeparation + ds.Dot(normal)

			var (
				velocityBias float32
				massScale    float32
				impulseScale float32
			)

			if s > 0.0 {
				// speculative bias is positive
				velocityBias = float32(s * invH)
				massScale = 1.0
				impulseScale = 0.0
			} else {
				// overlap bias is negative
				velocityBias = maxFloat(float32(softness.massScale*softness.biasRate*s), -contactSpeed)
				massScale = softness.massScale
				impulseScale = softness.impulseScale
			}

			// relative normal velocity at contact
			vrA := vA.Add(rA.ScalarCross(wA))
			vrB := vB.Add(rB.ScalarCross(wB))
			vn := vrB.Sub(vrA).Dot(normal)

			// incremental normal impulse
			impulse := float32(-cp.normalMass*(float32(massScale*vn)+velocityBias)) - float32(impulseScale*cp.normalImpulse)

			// clamp the accumulated impulse
			newImpulse := maxFloat(cp.normalImpulse+impulse, 0.0)
			impulse = newImpulse - cp.normalImpulse
			cp.normalImpulse = newImpulse
			cp.totalNormalImpulse += impulse

			// apply normal impulse
			p := normal.Scale(impulse)
			vA = vA.MulSub(mA, p)
			wA -= float32(iA * rA.Cross(p))

			vB = vB.MulAdd(mB, p)
			wB += float32(iB * rB.Cross(p))
		}

		if stateA.flags&dynamicFlag != 0 {
			stateA.linearVelocity = vA
			stateA.angularVelocity = wA
		}

		if stateB.flags&dynamicFlag != 0 {
			stateB.linearVelocity = vB
			stateB.angularVelocity = wB
		}
	}
}

// Solve contacts: normal, friction and rolling resistance.
func solveContactsOverflow(context *stepContext) {
	graph := context.graph
	color := &graph.colors[overflowIndex]
	constraints := color.overflowConstraints
	contactCount := len(color.contactSims)
	world := context.world
	awakeSet := &world.solverSets[awakeSet]
	states := awakeSet.bodyStates

	invH := context.invH

	// This is a dummy body to represent a static body since static bodies don't have a solver body.
	dummyState := identityBodyState

	for i := range contactCount {
		constraint := &constraints[i]
		mA := constraint.invMassA
		iA := constraint.invIA
		mB := constraint.invMassB
		iB := constraint.invIB

		indexA := constraint.indexA - 1
		indexB := constraint.indexB - 1

		stateA := &dummyState
		if indexA != nullIndex {
			stateA = &states[indexA]
		}

		vA := stateA.linearVelocity
		wA := stateA.angularVelocity
		dqA := stateA.deltaRotation

		stateB := &dummyState
		if indexB != nullIndex {
			stateB = &states[indexB]
		}

		vB := stateB.linearVelocity
		wB := stateB.angularVelocity
		dqB := stateB.deltaRotation

		dp := stateB.deltaPosition.Sub(stateA.deltaPosition)

		normal := constraint.normal
		tangent := normal.RightPerp()
		friction := constraint.friction

		pointCount := constraint.pointCount
		totalNormalImpulse := float32(0.0)

		for j := range pointCount {
			cp := &constraint.points[j]

			// fixed anchor points
			rA := cp.anchorA
			rB := cp.anchorB

			// compute current separation
			// this is subject to round-off error if the anchor is far from the body center of mass
			ds := dp.Add(dqB.Rotate(rB).Sub(dqA.Rotate(rA)))
			s := cp.baseSeparation + ds.Dot(normal)

			// speculative bias, zero when overlapped
			velocityBias := float32(0.0)
			if s > 0.0 {
				velocityBias = float32(s * invH)
			}

			// relative normal velocity at contact
			vrA := vA.Add(rA.ScalarCross(wA))
			vrB := vB.Add(rB.ScalarCross(wB))
			vn := vrB.Sub(vrA).Dot(normal)

			// incremental normal impulse
			impulse := float32(-cp.normalMass * (vn + velocityBias))

			// clamp the accumulated impulse
			newImpulse := maxFloat(cp.normalImpulse+impulse, 0.0)
			impulse = newImpulse - cp.normalImpulse
			cp.normalImpulse = newImpulse
			cp.totalNormalImpulse += impulse

			// This is used for rolling resistance
			totalNormalImpulse += newImpulse

			// apply normal impulse
			p := normal.Scale(impulse)
			vA = vA.MulSub(mA, p)
			wA -= float32(iA * rA.Cross(p))

			vB = vB.MulAdd(mB, p)
			wB += float32(iB * rB.Cross(p))
		}

		// Rolling resistance
		{
			deltaLambda := float32(-constraint.rollingMass * (wB - wA))
			lambda := constraint.rollingImpulse
			maxLambda := float32(constraint.rollingResistance * totalNormalImpulse)
			constraint.rollingImpulse = clampFloat(lambda+deltaLambda, -maxLambda, maxLambda)
			deltaLambda = constraint.rollingImpulse - lambda

			wA -= float32(iA * deltaLambda)
			wB += float32(iB * deltaLambda)
		}

		// Friction
		for j := range pointCount {
			cp := &constraint.points[j]

			// fixed anchor points
			rA := cp.anchorA
			rB := cp.anchorB

			// relative tangent velocity at contact
			vrB := vB.Add(rB.ScalarCross(wB))
			vrA := vA.Add(rA.ScalarCross(wA))

			// vt = dot(vrB - sB * tangent - (vrA + sA * tangent), tangent)
			//    = dot(vrB - vrA, tangent) - (sA + sB)

			vt := vrB.Sub(vrA).Dot(tangent) - constraint.tangentSpeed

			// incremental tangent impulse
			impulse := float32(cp.tangentMass * (-vt))

			// clamp the accumulated force
			maxFriction := float32(friction * cp.normalImpulse)
			newImpulse := clampFloat(cp.tangentImpulse+impulse, -maxFriction, maxFriction)
			impulse = newImpulse - cp.tangentImpulse
			cp.tangentImpulse = newImpulse

			// apply tangent impulse
			p := tangent.Scale(impulse)
			vA = vA.MulSub(mA, p)
			wA -= float32(iA * rA.Cross(p))
			vB = vB.MulAdd(mB, p)
			wB += float32(iB * rB.Cross(p))
		}

		if stateA.flags&dynamicFlag != 0 {
			stateA.linearVelocity = vA
			stateA.angularVelocity = wA
		}

		if stateB.flags&dynamicFlag != 0 {
			stateB.linearVelocity = vB
			stateB.angularVelocity = wB
		}
	}
}

func applyRestitutionOverflow(context *stepContext) {
	graph := context.graph
	color := &graph.colors[overflowIndex]
	constraints := color.overflowConstraints
	contactCount := len(color.contactSims)
	world := context.world
	awakeSet := &world.solverSets[awakeSet]
	states := awakeSet.bodyStates

	threshold := world.restitutionThreshold
	invH := context.invH
	propagate := world.enableRestitutionPropagation

	dummyState := identityBodyState

	for i := range contactCount {
		constraint := &constraints[i]
		if !propagate && constraint.restitution == 0.0 {
			continue
		}

		mA := constraint.invMassA
		iA := constraint.invIA
		mB := constraint.invMassB
		iB := constraint.invIB

		indexA := constraint.indexA - 1
		indexB := constraint.indexB - 1

		stateA := &dummyState
		if indexA != nullIndex {
			stateA = &states[indexA]
		}

		vA := stateA.linearVelocity
		wA := stateA.angularVelocity
		dqA := stateA.deltaRotation

		stateB := &dummyState
		if indexB != nullIndex {
			stateB = &states[indexB]
		}

		vB := stateB.linearVelocity
		wB := stateB.angularVelocity
		dqB := stateB.deltaRotation

		dp := stateB.deltaPosition.Sub(stateA.deltaPosition)

		normal := constraint.normal
		restitution := constraint.restitution
		pointCount := constraint.pointCount

		for j := range pointCount {
			cp := &constraint.points[j]

			rA := cp.anchorA
			rB := cp.anchorB

			// The total normal impulse is 0 for speculative points.
			compressionImpulse := cp.totalNormalImpulse - cp.restitutionImpulse
			armed := restitution > 0.0 && cp.relativeVelocity < -threshold && compressionImpulse > 0.0

			velocityBias := overflowRestitutionBias(cp, normal, dp, dqA, dqB, restitution, invH, armed)

			vrA := vA.Add(rA.ScalarCross(wA))
			vrB := vB.Add(rB.ScalarCross(wB))
			vn := vrB.Sub(vrA).Dot(normal)

			impulse := float32(-cp.normalMass * (vn + velocityBias))

			newImpulse := maxFloat(cp.normalImpulse+impulse, 0.0)
			impulse = newImpulse - cp.normalImpulse

			approachImpulse := float32(minFloat(maxFloat(-cp.normalMass*vn, 0.0), maxFloat(impulse, 0.0)))

			if armed {
				// Poisson kinetic restitution guarantees no energy gain.
				allowance := float32(restitution*(compressionImpulse+approachImpulse)) - cp.restitutionImpulse
				impulse = minFloat(impulse, approachImpulse+maxFloat(allowance, 0.0))
			}

			cp.normalImpulse += impulse
			cp.restitutionImpulse += impulse - approachImpulse
			cp.totalNormalImpulse += impulse

			p := normal.Scale(impulse)
			vA = vA.MulSub(mA, p)
			wA -= float32(iA * rA.Cross(p))

			vB = vB.MulAdd(mB, p)
			wB += float32(iB * rB.Cross(p))
		}

		if stateA.flags&dynamicFlag != 0 {
			stateA.linearVelocity = vA
			stateA.angularVelocity = wA
		}

		if stateB.flags&dynamicFlag != 0 {
			stateB.linearVelocity = vB
			stateB.angularVelocity = wB
		}
	}
}

func storeImpulsesOverflow(context *stepContext) {
	world := context.world
	graph := context.graph
	color := &graph.colors[overflowIndex]
	constraints := color.overflowConstraints
	contactSims := color.contactSims
	taskContext := &world.taskContexts[0]
	hitEventBitSet := &taskContext.hitEventBitSet
	contactCount := len(color.contactSims)
	negHitThreshold := -world.hitEventThreshold
	hasHitEvents := taskContext.hasHitEvents

	for i := range contactCount {
		constraint := &constraints[i]
		contactSim := &contactSims[i]
		manifold := &contactSim.manifold
		pointCount := manifold.PointCount

		for j := range pointCount {
			mp := &manifold.Points[j]
			mp.NormalImpulse = constraint.points[j].normalImpulse
			mp.TangentImpulse = constraint.points[j].tangentImpulse
			mp.TotalNormalImpulse = constraint.points[j].totalNormalImpulse
		}

		if contactSim.simFlags&simEnableHitEvent != 0 {
			for j := range contactSim.manifold.PointCount {
				mp := &manifold.Points[j]

				// Need to check total impulse because the point may be speculative and not colliding
				if mp.NormalVelocity < negHitThreshold && mp.TotalNormalImpulse > 0.0 {
					hitEventBitSet.SetBit(uint32(contactSim.contactID))

					hasHitEvents = true

					break
				}
			}
		}

		manifold.RollingImpulse = constraint.rollingImpulse
	}

	taskContext.hasHitEvents = hasHitEvents
}

func zeroW() floatW {
	return floatW{0.0, 0.0, 0.0, 0.0}
}

func splatW(scalar float32) floatW {
	return floatW{scalar, scalar, scalar, scalar}
}

func setW(a float32, b float32, c float32, d float32) floatW {
	return floatW{a, b, c, d}
}

func negW(a floatW) floatW {
	return floatW{-a.x, -a.y, -a.z, -a.w}
}

func addW(a floatW, b floatW) floatW {
	return floatW{a.x + b.x, a.y + b.y, a.z + b.z, a.w + b.w}
}

func subW(a floatW, b floatW) floatW {
	return floatW{a.x - b.x, a.y - b.y, a.z - b.z, a.w - b.w}
}

func mulW(a floatW, b floatW) floatW {
	return floatW{float32(a.x * b.x), float32(a.y * b.y), float32(a.z * b.z), float32(a.w * b.w)}
}

func divW(a floatW, b floatW) floatW {
	return floatW{a.x / b.x, a.y / b.y, a.z / b.z, a.w / b.w}
}

func mulAddW(a floatW, b floatW, c floatW) floatW {
	return floatW{a.x + float32(b.x*c.x), a.y + float32(b.y*c.y), a.z + float32(b.z*c.z), a.w + float32(b.w*c.w)}
}

func mulSubW(a floatW, b floatW, c floatW) floatW {
	return floatW{a.x - float32(b.x*c.x), a.y - float32(b.y*c.y), a.z - float32(b.z*c.z), a.w - float32(b.w*c.w)}
}

func minW(a floatW, b floatW) floatW {
	var r floatW
	if a.x <= b.x {
		r.x = a.x
	} else {
		r.x = b.x
	}

	if a.y <= b.y {
		r.y = a.y
	} else {
		r.y = b.y
	}

	if a.z <= b.z {
		r.z = a.z
	} else {
		r.z = b.z
	}

	if a.w <= b.w {
		r.w = a.w
	} else {
		r.w = b.w
	}

	return r
}

func maxW(a floatW, b floatW) floatW {
	var r floatW
	if a.x >= b.x {
		r.x = a.x
	} else {
		r.x = b.x
	}

	if a.y >= b.y {
		r.y = a.y
	} else {
		r.y = b.y
	}

	if a.z >= b.z {
		r.z = a.z
	} else {
		r.z = b.z
	}

	if a.w >= b.w {
		r.w = a.w
	} else {
		r.w = b.w
	}

	return r
}

// a = clamp(a, -b, b).
func symClampW(a floatW, b floatW) floatW {
	var r floatW

	r.x = clampFloat(a.x, -b.x, b.x)
	r.y = clampFloat(a.y, -b.y, b.y)
	r.z = clampFloat(a.z, -b.z, b.z)
	r.w = clampFloat(a.w, -b.w, b.w)

	return r
}

func andW(a floatW, b floatW) floatW {
	var r floatW
	if a.x != 0.0 && b.x != 0.0 {
		r.x = 1.0
	} else {
		r.x = 0.0
	}

	if a.y != 0.0 && b.y != 0.0 {
		r.y = 1.0
	} else {
		r.y = 0.0
	}

	if a.z != 0.0 && b.z != 0.0 {
		r.z = 1.0
	} else {
		r.z = 0.0
	}

	if a.w != 0.0 && b.w != 0.0 {
		r.w = 1.0
	} else {
		r.w = 0.0
	}

	return r
}

func greaterThanW(a floatW, b floatW) floatW {
	var r floatW
	if a.x > b.x {
		r.x = 1.0
	} else {
		r.x = 0.0
	}

	if a.y > b.y {
		r.y = 1.0
	} else {
		r.y = 0.0
	}

	if a.z > b.z {
		r.z = 1.0
	} else {
		r.z = 0.0
	}

	if a.w > b.w {
		r.w = 1.0
	} else {
		r.w = 0.0
	}

	return r
}

func lessThanW(a floatW, b floatW) floatW {
	var r floatW
	if a.x < b.x {
		r.x = 1.0
	} else {
		r.x = 0.0
	}

	if a.y < b.y {
		r.y = 1.0
	} else {
		r.y = 0.0
	}

	if a.z < b.z {
		r.z = 1.0
	} else {
		r.z = 0.0
	}

	if a.w < b.w {
		r.w = 1.0
	} else {
		r.w = 0.0
	}

	return r
}

func anyTrueW(mask floatW) bool {
	return mask.x != 0.0 || mask.y != 0.0 || mask.z != 0.0 || mask.w != 0.0
}

func allZeroW(a floatW) bool {
	return a.x == 0.0 && a.y == 0.0 && a.z == 0.0 && a.w == 0.0
}

func loadU(data []float32) floatW {
	return floatW{data[0], data[1], data[2], data[3]}
}

func storeW(data []float32, a floatW) {
	data[0] = a.x
	data[1] = a.y
	data[2] = a.z
	data[3] = a.w
}

// component-wise returns mask ? b : a.
func blendW(a floatW, b floatW, mask floatW) floatW {
	var r floatW
	if mask.x != 0.0 {
		r.x = b.x
	} else {
		r.x = a.x
	}

	if mask.y != 0.0 {
		r.y = b.y
	} else {
		r.y = a.y
	}

	if mask.z != 0.0 {
		r.z = b.z
	} else {
		r.z = a.z
	}

	if mask.w != 0.0 {
		r.w = b.w
	} else {
		r.w = a.w
	}

	return r
}

func unpackLoW(a floatW, b floatW) floatW {
	return floatW{a.x, b.x, a.y, b.y}
}

func unpackHiW(a floatW, b floatW) floatW {
	return floatW{a.z, b.z, a.w, b.w}
}

func softMaskW(indexA []int, indexB []int) floatW {
	var r floatW
	if indexA[0] == 0 || indexB[0] == 0 {
		r.x = 1.0
	} else {
		r.x = 0.0
	}

	if indexA[1] == 0 || indexB[1] == 0 {
		r.y = 1.0
	} else {
		r.y = 0.0
	}

	if indexA[2] == 0 || indexB[2] == 0 {
		r.z = 1.0
	} else {
		r.z = 0.0
	}

	if indexA[3] == 0 || indexB[3] == 0 {
		r.w = 1.0
	} else {
		r.w = 0.0
	}

	return r
}

func subVW(a vec2W, b vec2W) vec2W {
	return vec2W{x: subW(a.x, b.x), y: subW(a.y, b.y)}
}

func rightPerpW(a vec2W) vec2W {
	return vec2W{x: a.y, y: negW(a.x)}
}

func dotW(a vec2W, b vec2W) floatW {
	return floatW{float32(a.x.x*b.x.x) + float32(a.y.x*b.y.x), float32(a.x.y*b.x.y) + float32(a.y.y*b.y.y), float32(a.x.z*b.x.z) + float32(a.y.z*b.y.z), float32(a.x.w*b.x.w) + float32(a.y.w*b.y.w)}
}

func crossW(a vec2W, b vec2W) floatW {
	return floatW{float32(a.x.x*b.y.x) - float32(a.y.x*b.x.x), float32(a.x.y*b.y.y) - float32(a.y.y*b.x.y), float32(a.x.z*b.y.z) - float32(a.y.z*b.x.z), float32(a.x.w*b.y.w) - float32(a.y.w*b.x.w)}
}

func rotateVectorW(q rotW, v vec2W) vec2W {
	return vec2W{subW(mulW(q.c, v.x), mulW(q.s, v.y)), addW(mulW(q.s, v.x), mulW(q.c, v.y))}
}

func getWideContactConstraintByteCount() int {
	return int(unsafe.Sizeof(contactConstraintWide{}))
}

func gatherBodies(states []bodyState, indices []int, simdBody *bodyStateW, includePosition bool) {
	core.Validate(indices[0] >= 0 && indices[1] >= 0 && indices[2] >= 0 && indices[3] >= 0, "indices[0] >= 0 && indices[1] >= 0 && indices[2] >= 0 && indices[3] >= 0")

	// zero means null
	i1 := indices[0] - 1
	i2 := indices[1] - 1
	i3 := indices[2] - 1
	i4 := indices[3] - 1

	s1 := &identityBodyState
	if i1 != nullIndex {
		s1 = &states[i1]
	}

	s2 := &identityBodyState
	if i2 != nullIndex {
		s2 = &states[i2]
	}

	s3 := &identityBodyState
	if i3 != nullIndex {
		s3 = &states[i3]
	}

	s4 := &identityBodyState
	if i4 != nullIndex {
		s4 = &states[i4]
	}

	simdBody.v.x = floatW{s1.linearVelocity.X, s2.linearVelocity.X, s3.linearVelocity.X, s4.linearVelocity.X}
	simdBody.v.y = floatW{s1.linearVelocity.Y, s2.linearVelocity.Y, s3.linearVelocity.Y, s4.linearVelocity.Y}
	simdBody.w = floatW{s1.angularVelocity, s2.angularVelocity, s3.angularVelocity, s4.angularVelocity}

	if includePosition {
		simdBody.dp.x = floatW{s1.deltaPosition.X, s2.deltaPosition.X, s3.deltaPosition.X, s4.deltaPosition.X}
		simdBody.dp.y = floatW{s1.deltaPosition.Y, s2.deltaPosition.Y, s3.deltaPosition.Y, s4.deltaPosition.Y}
		simdBody.dq.c = floatW{s1.deltaRotation.C, s2.deltaRotation.C, s3.deltaRotation.C, s4.deltaRotation.C}
		simdBody.dq.s = floatW{s1.deltaRotation.S, s2.deltaRotation.S, s3.deltaRotation.S, s4.deltaRotation.S}
	}
}

// This writes only the velocities back to the solver bodies.
func scatterBodies(states []bodyState, indices []int, simdBody *bodyStateW) {
	core.Validate(indices[0] >= 0 && indices[1] >= 0 && indices[2] >= 0 && indices[3] >= 0, "indices[0] >= 0 && indices[1] >= 0 && indices[2] >= 0 && indices[3] >= 0")

	// zero means null
	i1 := indices[0] - 1
	i2 := indices[1] - 1
	i3 := indices[2] - 1
	i4 := indices[3] - 1

	if i1 != nullIndex && (states[i1].flags&dynamicFlag != 0) {
		state := &states[i1]
		state.linearVelocity.X = simdBody.v.x.x
		state.linearVelocity.Y = simdBody.v.y.x
		state.angularVelocity = simdBody.w.x
	}

	if i2 != nullIndex && (states[i2].flags&dynamicFlag != 0) {
		state := &states[i2]
		state.linearVelocity.X = simdBody.v.x.y
		state.linearVelocity.Y = simdBody.v.y.y
		state.angularVelocity = simdBody.w.y
	}

	if i3 != nullIndex && (states[i3].flags&dynamicFlag != 0) {
		state := &states[i3]
		state.linearVelocity.X = simdBody.v.x.z
		state.linearVelocity.Y = simdBody.v.y.z
		state.angularVelocity = simdBody.w.z
	}

	if i4 != nullIndex && (states[i4].flags&dynamicFlag != 0) {
		state := &states[i4]
		state.linearVelocity.X = simdBody.v.x.w
		state.linearVelocity.Y = simdBody.v.y.w
		state.angularVelocity = simdBody.w.w
	}
}

// Prepare wide contact constraints.
// Potential optimization: for SSE2 load 4 manifold points and transpose, like the body gather/scatter.
// I would need to reorder the manifold point so that constraint data comes first.
func prepareContactsWide(block solverBlock, context *stepContext) {
	world := context.world
	states := context.states
	spans := context.contactPrepareSpans
	wideBase := context.wideContactConstraints

	zero := zeroW()
	oneW := splatW(1.0)

	warmStartScale := zero
	if world.enableWarmStarting {
		warmStartScale = oneW
	}

	anyRestitution := false

	wideIndex := block.startIndex
	endWideIndex := block.startIndex + int(block.count)

	// Find color for start index. Linear search but fast.
	colorIndex := 0
	for spans[colorIndex+1].start <= wideIndex {
		colorIndex += 1
	}

	// Loop over block
	for wideIndex < endWideIndex {
		colorWideStart := spans[colorIndex].start
		colorWideEndIndex := min(spans[colorIndex+1].start, endWideIndex)
		colorContactCount := spans[colorIndex].count
		contactSims := spans[colorIndex].contacts

		if core.ValidationEnabled {
			expectedWide := 0
			if colorContactCount > 0 {
				expectedWide = ((colorContactCount - 1) >> simdShift) + 1
			}

			core.Assert(spans[colorIndex+1].start-spans[colorIndex].start == expectedWide, "spans[colorIndex + 1].start - spans[colorIndex].start == expectedWide")
		}

		// Loop over color
		for ; wideIndex < colorWideEndIndex; wideIndex++ {
			cw := &wideBase[wideIndex]
			localWideIndex := wideIndex - colorWideStart

			var contactLanes [simdWidth]*contactSim

			hitEvents := gatherWideContactLanes(world, cw, &contactLanes, contactSims, localWideIndex, colorContactCount)

			prepareWideContactGeometry(cw, &contactLanes, warmStartScale, zero, oneW)

			// Only sample contact point normal velocity if needed.
			restitutionMask := greaterThanW(cw.restitution, zero)
			haveRestitution := anyTrueW(restitutionMask)
			anyRestitution = anyRestitution || haveRestitution

			if (hitEvents&simEnableHitEvent) != 0 || haveRestitution {
				sampleWideContactVelocity(cw, &contactLanes, states)
			}
		}

		// Advance to next color
		colorIndex += 1
	}

	if anyRestitution {
		context.anyRestitution.Store(1)
	}
}

func warmStartContactPointWide(c *contactConstraintWide, bA, bB *bodyStateW, rA, rB vec2W, normalImpulse, tangentImpulse floatW, tangent vec2W) {
	var p vec2W

	p.x = addW(mulW(normalImpulse, c.normal.x), mulW(tangentImpulse, tangent.x))
	p.y = addW(mulW(normalImpulse, c.normal.y), mulW(tangentImpulse, tangent.y))
	bA.w = mulSubW(bA.w, c.invIA, crossW(rA, p))
	bA.v.x = mulSubW(bA.v.x, c.invMassA, p.x)
	bA.v.y = mulSubW(bA.v.y, c.invMassA, p.y)
	bB.w = mulAddW(bB.w, c.invIB, crossW(rB, p))
	bB.v.x = mulAddW(bB.v.x, c.invMassB, p.x)
	bB.v.y = mulAddW(bB.v.y, c.invMassB, p.y)
}

func applyContactImpulseWide(c *contactConstraintWide, bA, bB *bodyStateW, rA, rB *vec2W, px, py *floatW) {
	bA.v.x = mulSubW(bA.v.x, c.invMassA, *px)
	bA.v.y = mulSubW(bA.v.y, c.invMassA, *py)
	bA.w = mulSubW(bA.w, c.invIA, subW(mulW(rA.x, *py), mulW(rA.y, *px)))

	bB.v.x = mulAddW(bB.v.x, c.invMassB, *px)
	bB.v.y = mulAddW(bB.v.y, c.invMassB, *py)
	bB.w = mulAddW(bB.w, c.invIB, subW(mulW(rB.x, *py), mulW(rB.y, *px)))
}

func warmStartContactsWide(block solverBlock, context *stepContext) {
	states := context.states
	constraints := context.graph.colors[block.colorIndex].wideConstraints

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		c := &constraints[i]

		var bA, bB bodyStateW
		gatherBodies(states, c.indexA[:], &bA, false)
		gatherBodies(states, c.indexB[:], &bB, false)

		tangentX := c.normal.y
		tangentY := subW(zeroW(), c.normal.x)

		// fixed anchors
		warmStartContactPointWide(c, &bA, &bB, c.anchorA1, c.anchorB1, c.normalImpulse1, c.tangentImpulse1, vec2W{tangentX, tangentY})
		c.totalNormalImpulse1 = addW(c.totalNormalImpulse1, c.normalImpulse1)

		// fixed anchors
		warmStartContactPointWide(c, &bA, &bB, c.anchorA2, c.anchorB2, c.normalImpulse2, c.tangentImpulse2, vec2W{tangentX, tangentY})
		c.totalNormalImpulse2 = addW(c.totalNormalImpulse2, c.normalImpulse2)

		bA.w = mulSubW(bA.w, c.invIA, c.rollingImpulse)
		bB.w = mulAddW(bB.w, c.invIB, c.rollingImpulse)

		scatterBodies(states, c.indexA[:], &bA)
		scatterBodies(states, c.indexB[:], &bB)
	}
}

func pushContactsWide(block solverBlock, context *stepContext) {
	states := context.states
	color := &context.graph.colors[block.colorIndex]
	constraints := color.wideConstraints
	invH := splatW(context.invH)
	contactSpeed := splatW(-context.world.contactSpeed)
	oneW := splatW(1.0)

	dynamicBiasRate := splatW(context.contactSoftness.massScale * context.contactSoftness.biasRate)
	dynamicMassScale := splatW(context.contactSoftness.massScale)
	dynamicImpulseScale := splatW(context.contactSoftness.impulseScale)
	staticBiasRate := splatW(context.staticSoftness.massScale * context.staticSoftness.biasRate)
	staticMassScale := splatW(context.staticSoftness.massScale)
	staticImpulseScale := splatW(context.staticSoftness.impulseScale)

	for wideIndex := block.startIndex; wideIndex < block.startIndex+int(block.count); wideIndex++ {
		c := &constraints[wideIndex]

		var bA, bB bodyStateW
		gatherBodies(states, c.indexA[:], &bA, true)
		gatherBodies(states, c.indexB[:], &bB, true)

		softMask := softMaskW(c.indexA[:], c.indexB[:])
		biasRate := blendW(dynamicBiasRate, staticBiasRate, softMask)
		massScale := blendW(dynamicMassScale, staticMassScale, softMask)
		impulseScale := blendW(dynamicImpulseScale, staticImpulseScale, softMask)

		dp := vec2W{subW(bB.dp.x, bA.dp.x), subW(bB.dp.y, bA.dp.y)}

		// point1 non-penetration constraint
		{
			// Fixed anchors for impulses
			rA := c.anchorA1
			rB := c.anchorB1

			// Moving anchors for current separation
			rsA := rotateVectorW(bA.dq, rA)
			rsB := rotateVectorW(bB.dq, rB)

			// compute current separation
			// this is subject to round-off error if the anchor is far from the body center of mass
			ds := vec2W{addW(dp.x, subW(rsB.x, rsA.x)), addW(dp.y, subW(rsB.y, rsA.y))}
			s := addW(dotW(c.normal, ds), c.baseSeparation1)

			// Apply speculative bias if separation is greater than zero, otherwise apply soft constraint bias
			// The contactSpeed is meant to limit stiffness, not increase it.
			separated := greaterThanW(s, zeroW())

			// Speculative bias - positive
			specBias := mulW(s, invH)

			// Overlap bias - negative when overlapped
			overlapBias := maxW(mulW(biasRate, s), contactSpeed)

			// velocity bias = s > 0 ? spec : overlap
			velocityBias := blendW(overlapBias, specBias, separated)

			pointMassScale := blendW(massScale, oneW, separated)
			pointImpulseScale := blendW(impulseScale, zeroW(), separated)

			// Relative velocity at contact
			vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &c.normal)

			// Compute normal impulse
			negImpulse := addW(mulW(c.normalMass1, addW(mulW(pointMassScale, vn), velocityBias)), mulW(pointImpulseScale, c.normalImpulse1))

			// Clamp the accumulated impulse
			newImpulse := maxW(subW(c.normalImpulse1, negImpulse), zeroW())
			impulse := subW(newImpulse, c.normalImpulse1)
			c.normalImpulse1 = newImpulse
			c.totalNormalImpulse1 = addW(c.totalNormalImpulse1, impulse)

			// Apply contact impulse
			px := mulW(impulse, c.normal.x)
			py := mulW(impulse, c.normal.y)

			bA.v.x = mulSubW(bA.v.x, c.invMassA, px)
			bA.v.y = mulSubW(bA.v.y, c.invMassA, py)
			bA.w = mulSubW(bA.w, c.invIA, subW(mulW(rA.x, py), mulW(rA.y, px)))

			bB.v.x = mulAddW(bB.v.x, c.invMassB, px)
			bB.v.y = mulAddW(bB.v.y, c.invMassB, py)
			bB.w = mulAddW(bB.w, c.invIB, subW(mulW(rB.x, py), mulW(rB.y, px)))
		}

		// second point non-penetration constraint
		{
			// moving anchors for current separation
			rsA := rotateVectorW(bA.dq, c.anchorA2)
			rsB := rotateVectorW(bB.dq, c.anchorB2)

			// compute current separation
			ds := vec2W{addW(dp.x, subW(rsB.x, rsA.x)), addW(dp.y, subW(rsB.y, rsA.y))}
			s := addW(dotW(c.normal, ds), c.baseSeparation2)

			separated := greaterThanW(s, zeroW())

			specBias := mulW(s, invH)
			overlapBias := maxW(mulW(biasRate, s), contactSpeed)
			velocityBias := blendW(overlapBias, specBias, separated)

			pointMassScale := blendW(massScale, oneW, separated)
			pointImpulseScale := blendW(impulseScale, zeroW(), separated)

			// fixed anchors for Jacobians
			rA := c.anchorA2
			rB := c.anchorB2

			// Relative velocity at contact
			vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &c.normal)

			// Compute normal impulse
			negImpulse := addW(mulW(c.normalMass2, addW(mulW(pointMassScale, vn), velocityBias)), mulW(pointImpulseScale, c.normalImpulse2))

			// Clamp the accumulated impulse
			newImpulse := maxW(subW(c.normalImpulse2, negImpulse), zeroW())
			impulse := subW(newImpulse, c.normalImpulse2)
			c.normalImpulse2 = newImpulse
			c.totalNormalImpulse2 = addW(c.totalNormalImpulse2, impulse)

			// Apply contact impulse
			px := mulW(impulse, c.normal.x)
			py := mulW(impulse, c.normal.y)

			bA.v.x = mulSubW(bA.v.x, c.invMassA, px)
			bA.v.y = mulSubW(bA.v.y, c.invMassA, py)
			bA.w = mulSubW(bA.w, c.invIA, subW(mulW(rA.x, py), mulW(rA.y, px)))

			bB.v.x = mulAddW(bB.v.x, c.invMassB, px)
			bB.v.y = mulAddW(bB.v.y, c.invMassB, py)
			bB.w = mulAddW(bB.w, c.invIB, subW(mulW(rB.x, py), mulW(rB.y, px)))
		}

		scatterBodies(states, c.indexA[:], &bA)
		scatterBodies(states, c.indexB[:], &bB)
	}
}

// Solve the normal constraint, friction, and rolling resistance.
func solveContactsWide(block solverBlock, context *stepContext) {
	states := context.states
	color := &context.graph.colors[block.colorIndex]
	constraints := color.wideConstraints
	invH := splatW(context.invH)

	for wideIndex := block.startIndex; wideIndex < block.startIndex+int(block.count); wideIndex++ {
		c := &constraints[wideIndex]

		var bA, bB bodyStateW
		gatherBodies(states, c.indexA[:], &bA, true)
		gatherBodies(states, c.indexB[:], &bB, true)

		totalNormalImpulse := zeroW()

		dp := vec2W{subW(bB.dp.x, bA.dp.x), subW(bB.dp.y, bA.dp.y)}

		var velocityBias1, velocityBias2 floatW
		{
			// Moving anchors for current separation
			rsA := rotateVectorW(bA.dq, c.anchorA1)
			rsB := rotateVectorW(bB.dq, c.anchorB1)

			// compute current separation
			// this is subject to round-off error if the anchor is far from the body center of mass
			ds := vec2W{addW(dp.x, subW(rsB.x, rsA.x)), addW(dp.y, subW(rsB.y, rsA.y))}
			s := addW(dotW(c.normal, ds), c.baseSeparation1)

			// Speculative bias, positive if separated and zero if overlapped
			velocityBias1 = maxW(zeroW(), mulW(s, invH))
		}

		{
			rsA := rotateVectorW(bA.dq, c.anchorA2)
			rsB := rotateVectorW(bB.dq, c.anchorB2)

			ds := vec2W{addW(dp.x, subW(rsB.x, rsA.x)), addW(dp.y, subW(rsB.y, rsA.y))}
			s := addW(dotW(c.normal, ds), c.baseSeparation2)

			velocityBias2 = maxW(zeroW(), mulW(s, invH))
		}

		// point1 non-penetration constraint
		{
			// Fixed anchors for impulses
			rA := c.anchorA1
			rB := c.anchorB1

			// Relative velocity at contact
			vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &c.normal)

			newImpulse, impulse := normalContactImpulseWide(c.normalMass1, c.normalImpulse1, vn, velocityBias1)
			c.normalImpulse1 = newImpulse
			c.totalNormalImpulse1 = addW(c.totalNormalImpulse1, impulse)

			totalNormalImpulse = addW(totalNormalImpulse, newImpulse)

			// Apply contact impulse
			px := mulW(impulse, c.normal.x)
			py := mulW(impulse, c.normal.y)

			applyContactImpulseWide(c, &bA, &bB, &rA, &rB, &px, &py)
		}

		// second point non-penetration constraint
		{
			// fixed anchors for Jacobians
			rA := c.anchorA2
			rB := c.anchorB2

			// Relative velocity at contact
			vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &c.normal)

			newImpulse, impulse := normalContactImpulseWide(c.normalMass2, c.normalImpulse2, vn, velocityBias2)
			c.normalImpulse2 = newImpulse
			c.totalNormalImpulse2 = addW(c.totalNormalImpulse2, impulse)

			totalNormalImpulse = addW(totalNormalImpulse, newImpulse)

			// Apply contact impulse
			px := mulW(impulse, c.normal.x)
			py := mulW(impulse, c.normal.y)

			applyContactImpulseWide(c, &bA, &bB, &rA, &rB, &px, &py)
		}

		// Rolling resistance
		if !allZeroW(c.rollingResistance) {
			k := addW(c.invIA, c.invIB)
			mask := greaterThanW(k, zeroW())
			deltaLambda := blendW(zeroW(), divW(subW(bA.w, bB.w), k), mask)
			lambda := c.rollingImpulse
			maxLambda := mulW(c.rollingResistance, totalNormalImpulse)
			c.rollingImpulse = symClampW(addW(lambda, deltaLambda), maxLambda)
			deltaLambda = subW(c.rollingImpulse, lambda)

			bA.w = mulSubW(bA.w, c.invIA, deltaLambda)
			bB.w = mulAddW(bB.w, c.invIB, deltaLambda)
		}

		tangentX := c.normal.y
		tangentY := subW(zeroW(), c.normal.x)

		// point 1 friction constraint
		{
			// Fixed anchor points for applying impulses
			rA := c.anchorA1
			rB := c.anchorB1

			// Relative velocity at contact
			vt := contactPointVelocityWide(&bA, &bB, &rA, &rB, &vec2W{tangentX, tangentY})

			// Tangent speed (conveyor belt)
			vt = subW(vt, c.tangentSpeed)

			newImpulse, impulse := frictionContactImpulseWide(c.friction, c.normalImpulse1, c.tangentMass1, c.tangentImpulse1, vt)
			c.tangentImpulse1 = newImpulse

			// Apply contact impulse
			px := mulW(impulse, tangentX)
			py := mulW(impulse, tangentY)

			applyContactImpulseWide(c, &bA, &bB, &rA, &rB, &px, &py)
		}

		// second point friction constraint
		{
			// fixed anchors for Jacobians
			rA := c.anchorA2
			rB := c.anchorB2

			// Relative velocity at contact
			vt := contactPointVelocityWide(&bA, &bB, &rA, &rB, &vec2W{tangentX, tangentY})

			// Tangent speed (conveyor belt)
			vt = subW(vt, c.tangentSpeed)

			newImpulse, impulse := frictionContactImpulseWide(c.friction, c.normalImpulse2, c.tangentMass2, c.tangentImpulse2, vt)
			c.tangentImpulse2 = newImpulse

			// Apply contact impulse
			px := mulW(impulse, tangentX)
			py := mulW(impulse, tangentY)

			applyContactImpulseWide(c, &bA, &bB, &rA, &rB, &px, &py)
		}

		scatterBodies(states, c.indexA[:], &bA)
		scatterBodies(states, c.indexB[:], &bB)
	}
}

func applyRestitutionWide(block solverBlock, context *stepContext) {
	states := context.states
	color := &context.graph.colors[block.colorIndex]
	constraints := color.wideConstraints
	invH := splatW(context.invH)
	negRestitutionThreshold := splatW(-context.world.restitutionThreshold)
	zeroW := zeroW()
	propagate := context.world.enableRestitutionPropagation

	for wideIndex := block.startIndex; wideIndex < block.startIndex+int(block.count); wideIndex++ {
		c := &constraints[wideIndex]
		if !propagate && allZeroW(c.restitution) {
			continue
		}

		var bA, bB bodyStateW
		gatherBodies(states, c.indexA[:], &bA, true)
		gatherBodies(states, c.indexB[:], &bB, true)

		restitutionMask := greaterThanW(c.restitution, zeroW)

		normalMass1 := blendW(zeroW, c.normalMass1, restitutionMask)
		if propagate {
			normalMass1 = c.normalMass1
		}

		normalMass2 := blendW(zeroW, c.normalMass2, restitutionMask)
		if propagate {
			normalMass2 = c.normalMass2
		}

		dp := vec2W{subW(bB.dp.x, bA.dp.x), subW(bB.dp.y, bA.dp.y)}

		{
			rA := c.anchorA1
			rB := c.anchorB1

			compressionImpulse := subW(c.totalNormalImpulse1, c.restitutionImpulse1)
			armed := andW(andW(restitutionMask, lessThanW(c.relativeVelocity1, negRestitutionThreshold)), greaterThanW(compressionImpulse, zeroW))

			rsA := rotateVectorW(bA.dq, rA)
			rsB := rotateVectorW(bB.dq, rB)
			ds := vec2W{addW(dp.x, subW(rsB.x, rsA.x)), addW(dp.y, subW(rsB.y, rsA.y))}
			s := addW(dotW(c.normal, ds), c.baseSeparation1)

			specBias := maxW(zeroW, mulW(s, invH))
			velocityBias := blendW(specBias, mulW(c.restitution, c.relativeVelocity1), armed)

			vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &c.normal)

			impulse, approachImpulse := restitutionContactImpulseWide(normalMass1, c.normalImpulse1, vn, velocityBias, c.restitution, compressionImpulse, c.restitutionImpulse1, armed, zeroW)

			c.normalImpulse1 = addW(c.normalImpulse1, impulse)
			c.restitutionImpulse1 = addW(c.restitutionImpulse1, subW(impulse, approachImpulse))
			c.totalNormalImpulse1 = addW(c.totalNormalImpulse1, impulse)

			px := mulW(impulse, c.normal.x)
			py := mulW(impulse, c.normal.y)

			applyContactImpulseWide(c, &bA, &bB, &rA, &rB, &px, &py)
		}

		{
			rA := c.anchorA2
			rB := c.anchorB2

			compressionImpulse := subW(c.totalNormalImpulse2, c.restitutionImpulse2)
			armed := andW(andW(restitutionMask, lessThanW(c.relativeVelocity2, negRestitutionThreshold)), greaterThanW(compressionImpulse, zeroW))

			rsA := rotateVectorW(bA.dq, rA)
			rsB := rotateVectorW(bB.dq, rB)
			ds := vec2W{addW(dp.x, subW(rsB.x, rsA.x)), addW(dp.y, subW(rsB.y, rsA.y))}
			s := addW(dotW(c.normal, ds), c.baseSeparation2)

			specBias := maxW(zeroW, mulW(s, invH))
			velocityBias := blendW(specBias, mulW(c.restitution, c.relativeVelocity2), armed)

			vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &c.normal)

			impulse, approachImpulse := restitutionContactImpulseWide(normalMass2, c.normalImpulse2, vn, velocityBias, c.restitution, compressionImpulse, c.restitutionImpulse2, armed, zeroW)

			c.normalImpulse2 = addW(c.normalImpulse2, impulse)
			c.restitutionImpulse2 = addW(c.restitutionImpulse2, subW(impulse, approachImpulse))
			c.totalNormalImpulse2 = addW(c.totalNormalImpulse2, impulse)

			px := mulW(impulse, c.normal.x)
			py := mulW(impulse, c.normal.y)

			applyContactImpulseWide(c, &bA, &bB, &rA, &rB, &px, &py)
		}

		scatterBodies(states, c.indexA[:], &bA)
		scatterBodies(states, c.indexB[:], &bB)
	}
}

// I tried adding this to the last relax iterations but it was slower.
//
// Runs as a flat parallel-for over the whole wide constraint range. Per-color
// contact sims are looked up through the prepareSpans cursor rather than the
// block's colorIndex, matching the layout of prepareContactsWide.
//
// I could store the manifold pointer in the contactConstraintWide to simplify
// this. But it is important to keep the constraint small.
func storeImpulsesWide(block solverBlock, context *stepContext, workerIndex int) {
	world := context.world
	spans := context.contactPrepareSpans
	wideBase := context.wideContactConstraints
	taskContext := &world.taskContexts[workerIndex]
	hitEventBitSet := &taskContext.hitEventBitSet
	hasHitEvents := taskContext.hasHitEvents
	negHitThreshold := -world.hitEventThreshold

	wideIndex := block.startIndex
	endWideIndex := block.startIndex + int(block.count)

	// Find color for start index
	colorIndex := 0
	for spans[colorIndex+1].start <= wideIndex {
		colorIndex += 1
	}

	for wideIndex < endWideIndex {
		colorWideEndIndex := min(spans[colorIndex+1].start, endWideIndex)
		colorWideStart := spans[colorIndex].start
		colorContactCount := spans[colorIndex].count
		contactSims := spans[colorIndex].contacts

		for ; wideIndex < colorWideEndIndex; wideIndex++ {
			c := &wideBase[wideIndex]
			rollingImpulse := [simdWidth]float32{c.rollingImpulse.x, c.rollingImpulse.y, c.rollingImpulse.z, c.rollingImpulse.w}
			normalImpulse1 := [simdWidth]float32{c.normalImpulse1.x, c.normalImpulse1.y, c.normalImpulse1.z, c.normalImpulse1.w}
			normalImpulse2 := [simdWidth]float32{c.normalImpulse2.x, c.normalImpulse2.y, c.normalImpulse2.z, c.normalImpulse2.w}
			tangentImpulse1 := [simdWidth]float32{c.tangentImpulse1.x, c.tangentImpulse1.y, c.tangentImpulse1.z, c.tangentImpulse1.w}
			tangentImpulse2 := [simdWidth]float32{c.tangentImpulse2.x, c.tangentImpulse2.y, c.tangentImpulse2.z, c.tangentImpulse2.w}
			totalNormalImpulse1 := [simdWidth]float32{c.totalNormalImpulse1.x, c.totalNormalImpulse1.y, c.totalNormalImpulse1.z, c.totalNormalImpulse1.w}
			totalNormalImpulse2 := [simdWidth]float32{c.totalNormalImpulse2.x, c.totalNormalImpulse2.y, c.totalNormalImpulse2.z, c.totalNormalImpulse2.w}

			localWideIndex := wideIndex - colorWideStart
			baseIndex := simdWidth * localWideIndex

			for laneIndex := range simdWidth {
				contactIndex := baseIndex + laneIndex
				if contactIndex >= colorContactCount {
					break
				}

				contactSim := &contactSims[contactIndex]
				m := &contactSim.manifold
				m.RollingImpulse = rollingImpulse[laneIndex]

				m.Points[0].NormalImpulse = normalImpulse1[laneIndex]
				m.Points[0].TangentImpulse = tangentImpulse1[laneIndex]
				m.Points[0].TotalNormalImpulse = totalNormalImpulse1[laneIndex]

				m.Points[1].NormalImpulse = normalImpulse2[laneIndex]
				m.Points[1].TangentImpulse = tangentImpulse2[laneIndex]
				m.Points[1].TotalNormalImpulse = totalNormalImpulse2[laneIndex]

				if markContactHitEvent(contactSim, hitEventBitSet, negHitThreshold) {
					hasHitEvents = true
				}
			}
		}

		// Advance to next color
		colorIndex += 1
	}

	taskContext.hasHitEvents = hasHitEvents
}

func prepareOverflowContactDynamics(context *stepContext, contactSim *contactSim, constraint *contactConstraint, indexA, indexB int, warmStartScale float32, sampleVelocity bool) {
	states := context.states
	manifold := &contactSim.manifold
	pointCount := constraint.pointCount
	// Stiffer for static contacts to avoid bodies getting pushed through the ground
	contactSoftness := context.contactSoftness
	staticSoftness := context.staticSoftness
	vA := Vec2Zero
	wA := float32(0.0)
	vB := Vec2Zero
	wB := float32(0.0)

	if sampleVelocity {
		if indexA != nullIndex {
			vA = states[indexA].linearVelocity
			wA = states[indexA].angularVelocity
		}

		if indexB != nullIndex {
			vB = states[indexB].linearVelocity
			wB = states[indexB].angularVelocity
		}
	}

	mA := contactSim.invMassA
	iA := contactSim.invIA
	mB := contactSim.invMassB
	iB := contactSim.invIB

	if indexA == nullIndex || indexB == nullIndex {
		constraint.softness = staticSoftness
	} else {
		constraint.softness = contactSoftness
	}

	// copy mass into constraint to avoid cache misses during sub-stepping
	constraint.invMassA = mA
	constraint.invIA = iA
	constraint.invMassB = mB
	constraint.invIB = iB

	{
		k := iA + iB
		if k > 0.0 {
			constraint.rollingMass = 1.0 / k
		} else {
			constraint.rollingMass = 0.0
		}
	}

	normal := constraint.normal
	tangent := constraint.normal.RightPerp()

	for j := range pointCount {
		mp := &manifold.Points[j]
		cp := &constraint.points[j]

		cp.normalImpulse = float32(warmStartScale * mp.NormalImpulse)
		cp.tangentImpulse = float32(warmStartScale * mp.TangentImpulse)
		cp.totalNormalImpulse = 0.0
		cp.restitutionImpulse = 0.0

		rA := mp.AnchorA
		rB := mp.AnchorB

		cp.anchorA = rA
		cp.anchorB = rB
		cp.baseSeparation = mp.Separation - rB.Sub(rA).Dot(normal)

		rnA := rA.Cross(normal)
		rnB := rB.Cross(normal)

		kNormal := mA + mB + float32(iA*rnA*rnA) + float32(iB*rnB*rnB)
		if kNormal > 0.0 {
			cp.normalMass = 1.0 / kNormal
		} else {
			cp.normalMass = 0.0
		}

		rtA := rA.Cross(tangent)
		rtB := rB.Cross(tangent)

		kTangent := mA + mB + float32(iA*rtA*rtA) + float32(iB*rtB*rtB)
		if kTangent > 0.0 {
			cp.tangentMass = 1.0 / kTangent
		} else {
			cp.tangentMass = 0.0
		}

		// Only compute the normal velocity terms if needed.
		if sampleVelocity {
			vrA := vA.Add(rA.ScalarCross(wA))
			vrB := vB.Add(rB.ScalarCross(wB))
			vn := normal.Dot(vrB.Sub(vrA))

			cp.relativeVelocity = vn
			mp.NormalVelocity = vn
		} else {
			cp.relativeVelocity = 0.0
			mp.NormalVelocity = 0.0
		}
	}
}

func overflowRestitutionBias(cp *contactConstraintPoint, normal, dp Vec2, dqA, dqB Rot, restitution, invH float32, armed bool) float32 {
	if armed {
		return float32(restitution * cp.relativeVelocity)
	}

	ds := dp.Add(dqB.Rotate(cp.anchorB).Sub(dqA.Rotate(cp.anchorA)))

	s := cp.baseSeparation + ds.Dot(normal)
	if s > 0.0 {
		return float32(s * invH)
	}

	return 0.0
}

func gatherWideContactLanes(world *World, cw *contactConstraintWide, contactLanes *[simdWidth]*contactSim, contactSims []contactSim, localWideIndex, colorContactCount int) uint32 {
	hitEvents := uint32(0)

	for laneIndex := range simdWidth {
		contactIndex := simdWidth*localWideIndex + laneIndex
		if contactIndex < colorContactCount {
			c := &contactSims[contactIndex]
			contactLanes[laneIndex] = c

			// index base-1
			cw.indexA[laneIndex] = decodeAwakeIndex(c.encodedBodySimA) + 1
			cw.indexB[laneIndex] = decodeAwakeIndex(c.encodedBodySimB) + 1
			hitEvents |= (c.simFlags & simEnableHitEvent)

			if core.ValidationEnabled {
				bodyA := &world.bodies[c.bodyIDA]
				bodyB := &world.bodies[c.bodyIDB]
				core.Assert(c.encodedBodySimA == encodeBodySimIndex(bodyA), "c.encodedBodySimA == encodeBodySimIndex( bodyA )")
				core.Assert(c.encodedBodySimB == encodeBodySimIndex(bodyB), "c.encodedBodySimB == encodeBodySimIndex( bodyB )")
			}
		} else {
			contactLanes[laneIndex] = &zeroContactSim
		}
	}

	return hitEvents
}

func prepareWideContactGeometry(cw *contactConstraintWide, contactLanes *[simdWidth]*contactSim, warmStartScale, zero, oneW floatW) {
	cw.invMassA = setW(contactLanes[0].invMassA, contactLanes[1].invMassA, contactLanes[2].invMassA, contactLanes[3].invMassA)
	cw.invMassB = setW(contactLanes[0].invMassB, contactLanes[1].invMassB, contactLanes[2].invMassB, contactLanes[3].invMassB)
	cw.invIA = setW(contactLanes[0].invIA, contactLanes[1].invIA, contactLanes[2].invIA, contactLanes[3].invIA)
	cw.invIB = setW(contactLanes[0].invIB, contactLanes[1].invIB, contactLanes[2].invIB, contactLanes[3].invIB)
	cw.normal.x = setW(contactLanes[0].manifold.Normal.X, contactLanes[1].manifold.Normal.X, contactLanes[2].manifold.Normal.X, contactLanes[3].manifold.Normal.X)
	cw.normal.y = setW(contactLanes[0].manifold.Normal.Y, contactLanes[1].manifold.Normal.Y, contactLanes[2].manifold.Normal.Y, contactLanes[3].manifold.Normal.Y)
	cw.friction = setW(contactLanes[0].friction, contactLanes[1].friction, contactLanes[2].friction, contactLanes[3].friction)
	cw.tangentSpeed = setW(contactLanes[0].tangentSpeed, contactLanes[1].tangentSpeed, contactLanes[2].tangentSpeed, contactLanes[3].tangentSpeed)
	cw.rollingResistance = setW(contactLanes[0].rollingResistance, contactLanes[1].rollingResistance, contactLanes[2].rollingResistance, contactLanes[3].rollingResistance)
	cw.rollingImpulse = setW(contactLanes[0].manifold.RollingImpulse, contactLanes[1].manifold.RollingImpulse, contactLanes[2].manifold.RollingImpulse, contactLanes[3].manifold.RollingImpulse)
	cw.rollingImpulse = mulW(warmStartScale, cw.rollingImpulse)

	cw.restitution = setW(contactLanes[0].restitution, contactLanes[1].restitution, contactLanes[2].restitution, contactLanes[3].restitution)

	tangent := rightPerpW(cw.normal)

	pointCounts := setW(float32(contactLanes[0].manifold.PointCount), float32(contactLanes[1].manifold.PointCount), float32(contactLanes[2].manifold.PointCount), float32(contactLanes[3].manifold.PointCount))

	twoPointMask := greaterThanW(pointCounts, oneW)

	{
		// Optimized transpose gather. Similar to gatherBodies.
		m1a := setW(contactLanes[0].manifold.Points[0].AnchorA.X, contactLanes[0].manifold.Points[0].AnchorA.Y, contactLanes[0].manifold.Points[0].AnchorB.X, contactLanes[0].manifold.Points[0].AnchorB.Y)
		m1b := setW(contactLanes[0].manifold.Points[0].Separation, contactLanes[0].manifold.Points[0].NormalImpulse, contactLanes[0].manifold.Points[0].TangentImpulse, contactLanes[0].manifold.Points[0].TotalNormalImpulse)
		m2a := setW(contactLanes[1].manifold.Points[0].AnchorA.X, contactLanes[1].manifold.Points[0].AnchorA.Y, contactLanes[1].manifold.Points[0].AnchorB.X, contactLanes[1].manifold.Points[0].AnchorB.Y)
		m2b := setW(contactLanes[1].manifold.Points[0].Separation, contactLanes[1].manifold.Points[0].NormalImpulse, contactLanes[1].manifold.Points[0].TangentImpulse, contactLanes[1].manifold.Points[0].TotalNormalImpulse)
		m3a := setW(contactLanes[2].manifold.Points[0].AnchorA.X, contactLanes[2].manifold.Points[0].AnchorA.Y, contactLanes[2].manifold.Points[0].AnchorB.X, contactLanes[2].manifold.Points[0].AnchorB.Y)
		m3b := setW(contactLanes[2].manifold.Points[0].Separation, contactLanes[2].manifold.Points[0].NormalImpulse, contactLanes[2].manifold.Points[0].TangentImpulse, contactLanes[2].manifold.Points[0].TotalNormalImpulse)
		m4a := setW(contactLanes[3].manifold.Points[0].AnchorA.X, contactLanes[3].manifold.Points[0].AnchorA.Y, contactLanes[3].manifold.Points[0].AnchorB.X, contactLanes[3].manifold.Points[0].AnchorB.Y)
		m4b := setW(contactLanes[3].manifold.Points[0].Separation, contactLanes[3].manifold.Points[0].NormalImpulse, contactLanes[3].manifold.Points[0].TangentImpulse, contactLanes[3].manifold.Points[0].TotalNormalImpulse)

		t1a := unpackLoW(m1a, m3a)
		t2a := unpackLoW(m2a, m4a)
		t3a := unpackHiW(m1a, m3a)
		t4a := unpackHiW(m2a, m4a)
		t1b := unpackLoW(m1b, m3b)
		t2b := unpackLoW(m2b, m4b)
		t3b := unpackHiW(m1b, m3b)
		t4b := unpackHiW(m2b, m4b)

		cw.anchorA1.x = unpackLoW(t1a, t2a)
		cw.anchorA1.y = unpackHiW(t1a, t2a)
		cw.anchorB1.x = unpackLoW(t3a, t4a)
		cw.anchorB1.y = unpackHiW(t3a, t4a)
		cw.baseSeparation1 = unpackLoW(t1b, t2b)
		cw.normalImpulse1 = unpackHiW(t1b, t2b)
		cw.tangentImpulse1 = unpackLoW(t3b, t4b)

		// separation1 = cw.baseSeparation1;

		offset := dotW(subVW(cw.anchorB1, cw.anchorA1), cw.normal)
		cw.baseSeparation1 = subW(cw.baseSeparation1, offset)
		cw.normalImpulse1 = mulW(warmStartScale, cw.normalImpulse1)
		cw.tangentImpulse1 = mulW(warmStartScale, cw.tangentImpulse1)
		cw.totalNormalImpulse1 = zero
		cw.relativeVelocity1 = zero
		cw.restitutionImpulse1 = zero

		// Effective mass along normal.
		{
			rnA := crossW(cw.anchorA1, cw.normal)
			rnB := crossW(cw.anchorB1, cw.normal)
			k := addW(cw.invMassA, cw.invMassB)
			k = mulAddW(k, cw.invIA, mulW(rnA, rnA))
			k = mulAddW(k, cw.invIB, mulW(rnB, rnB))
			test := greaterThanW(k, zeroW())
			cw.normalMass1 = blendW(zeroW(), divW(oneW, k), test)
		}

		// Effective mass along tangent.
		{
			rtA := crossW(cw.anchorA1, tangent)
			rtB := crossW(cw.anchorB1, tangent)
			k := addW(cw.invMassA, cw.invMassB)
			k = mulAddW(k, cw.invIA, mulW(rtA, rtA))
			k = mulAddW(k, cw.invIB, mulW(rtB, rtB))
			test := greaterThanW(k, zeroW())
			cw.tangentMass1 = blendW(zeroW(), divW(oneW, k), test)
		}
	}

	{
		m1a := setW(contactLanes[0].manifold.Points[1].AnchorA.X, contactLanes[0].manifold.Points[1].AnchorA.Y, contactLanes[0].manifold.Points[1].AnchorB.X, contactLanes[0].manifold.Points[1].AnchorB.Y)
		m1b := setW(contactLanes[0].manifold.Points[1].Separation, contactLanes[0].manifold.Points[1].NormalImpulse, contactLanes[0].manifold.Points[1].TangentImpulse, contactLanes[0].manifold.Points[1].TotalNormalImpulse)
		m2a := setW(contactLanes[1].manifold.Points[1].AnchorA.X, contactLanes[1].manifold.Points[1].AnchorA.Y, contactLanes[1].manifold.Points[1].AnchorB.X, contactLanes[1].manifold.Points[1].AnchorB.Y)
		m2b := setW(contactLanes[1].manifold.Points[1].Separation, contactLanes[1].manifold.Points[1].NormalImpulse, contactLanes[1].manifold.Points[1].TangentImpulse, contactLanes[1].manifold.Points[1].TotalNormalImpulse)
		m3a := setW(contactLanes[2].manifold.Points[1].AnchorA.X, contactLanes[2].manifold.Points[1].AnchorA.Y, contactLanes[2].manifold.Points[1].AnchorB.X, contactLanes[2].manifold.Points[1].AnchorB.Y)
		m3b := setW(contactLanes[2].manifold.Points[1].Separation, contactLanes[2].manifold.Points[1].NormalImpulse, contactLanes[2].manifold.Points[1].TangentImpulse, contactLanes[2].manifold.Points[1].TotalNormalImpulse)
		m4a := setW(contactLanes[3].manifold.Points[1].AnchorA.X, contactLanes[3].manifold.Points[1].AnchorA.Y, contactLanes[3].manifold.Points[1].AnchorB.X, contactLanes[3].manifold.Points[1].AnchorB.Y)
		m4b := setW(contactLanes[3].manifold.Points[1].Separation, contactLanes[3].manifold.Points[1].NormalImpulse, contactLanes[3].manifold.Points[1].TangentImpulse, contactLanes[3].manifold.Points[1].TotalNormalImpulse)

		t1a := unpackLoW(m1a, m3a)
		t2a := unpackLoW(m2a, m4a)
		t3a := unpackHiW(m1a, m3a)
		t4a := unpackHiW(m2a, m4a)
		t1b := unpackLoW(m1b, m3b)
		t2b := unpackLoW(m2b, m4b)
		t3b := unpackHiW(m1b, m3b)
		t4b := unpackHiW(m2b, m4b)

		cw.anchorA2.x = unpackLoW(t1a, t2a)
		cw.anchorA2.y = unpackHiW(t1a, t2a)
		cw.anchorB2.x = unpackLoW(t3a, t4a)
		cw.anchorB2.y = unpackHiW(t3a, t4a)
		cw.baseSeparation2 = unpackLoW(t1b, t2b)
		cw.normalImpulse2 = unpackHiW(t1b, t2b)
		cw.tangentImpulse2 = unpackLoW(t3b, t4b)

		// separation2 = cw.baseSeparation2;

		offset := dotW(subVW(cw.anchorB2, cw.anchorA2), cw.normal)
		cw.baseSeparation2 = subW(cw.baseSeparation2, offset)
		cw.normalImpulse2 = mulW(warmStartScale, cw.normalImpulse2)
		cw.tangentImpulse2 = mulW(warmStartScale, cw.tangentImpulse2)
		cw.totalNormalImpulse2 = zero
		cw.relativeVelocity2 = zero
		cw.restitutionImpulse2 = zero

		// Effective mass along normal.
		{
			rnA := crossW(cw.anchorA2, cw.normal)
			rnB := crossW(cw.anchorB2, cw.normal)
			k := addW(cw.invMassA, cw.invMassB)
			k = mulAddW(k, cw.invIA, mulW(rnA, rnA))
			k = mulAddW(k, cw.invIB, mulW(rnB, rnB))
			test := greaterThanW(k, zeroW())
			cw.normalMass2 = blendW(zeroW(), divW(oneW, k), test)
		}

		// Effective mass along tangent.
		{
			rtA := crossW(cw.anchorA2, tangent)
			rtB := crossW(cw.anchorB2, tangent)
			k := addW(cw.invMassA, cw.invMassB)
			k = mulAddW(k, cw.invIA, mulW(rtA, rtA))
			k = mulAddW(k, cw.invIB, mulW(rtB, rtB))
			test := greaterThanW(k, zeroW())
			cw.tangentMass2 = blendW(zeroW(), divW(oneW, k), test)
		}

		// Zero effective mass on lanes that only have one point.
		cw.normalMass2 = blendW(zero, cw.normalMass2, twoPointMask)
		cw.tangentMass2 = blendW(zero, cw.tangentMass2, twoPointMask)
	}
}

func sampleWideContactVelocity(cw *contactConstraintWide, contactLanes *[simdWidth]*contactSim, states []bodyState) {
	var bA, bB bodyStateW
	gatherBodies(states, cw.indexA[:], &bA, false)
	gatherBodies(states, cw.indexB[:], &bB, false)

	{
		rA := cw.anchorA1
		rB := cw.anchorB1

		// Relative velocity at contact
		vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &cw.normal)
		cw.relativeVelocity1 = vn

		var normalVelocities [simdWidth]float32
		storeW(normalVelocities[:], vn)

		for lane := range simdWidth {
			// Check the point count to avoid a race writing to zeroContactSim.
			if contactLanes[lane].manifold.PointCount > 0 {
				contactLanes[lane].manifold.Points[0].NormalVelocity = normalVelocities[lane]
			}
		}
	}

	{
		rA := cw.anchorA2
		rB := cw.anchorB2

		// Relative velocity at contact
		vn := contactPointVelocityWide(&bA, &bB, &rA, &rB, &cw.normal)
		cw.relativeVelocity2 = vn

		var normalVelocities [simdWidth]float32
		storeW(normalVelocities[:], vn)

		for lane := range simdWidth {
			if contactLanes[lane].manifold.PointCount > 1 {
				contactLanes[lane].manifold.Points[1].NormalVelocity = normalVelocities[lane]
			}
		}
	}
}

func contactPointVelocityWide(bA, bB *bodyStateW, rA, rB, direction *vec2W) floatW {
	dvx := subW(subW(bB.v.x, mulW(bB.w, rB.y)), subW(bA.v.x, mulW(bA.w, rA.y)))
	dvy := subW(addW(bB.v.y, mulW(bB.w, rB.x)), addW(bA.v.y, mulW(bA.w, rA.x)))

	return addW(mulW(dvx, direction.x), mulW(dvy, direction.y))
}

func normalContactImpulseWide(normalMass, normalImpulse, vn, velocityBias floatW) (floatW, floatW) {
	// Compute normal impulse
	negImpulse := mulW(normalMass, addW(vn, velocityBias))
	// Clamp the accumulated impulse
	newImpulse := maxW(subW(normalImpulse, negImpulse), zeroW())
	impulse := subW(newImpulse, normalImpulse)

	return newImpulse, impulse
}

func frictionContactImpulseWide(friction, normalImpulse, tangentMass, tangentImpulse, vt floatW) (floatW, floatW) {
	// Compute tangent force
	negImpulse := mulW(tangentMass, vt)
	// Clamp the accumulated force
	maxFriction := mulW(friction, normalImpulse)
	newImpulse := subW(tangentImpulse, negImpulse)
	newImpulse = maxW(subW(zeroW(), maxFriction), minW(newImpulse, maxFriction))
	impulse := subW(newImpulse, tangentImpulse)

	return newImpulse, impulse
}

func restitutionContactImpulseWide(normalMass, normalImpulse, vn, velocityBias, restitution, compressionImpulse, restitutionImpulse, armed, zero floatW) (floatW, floatW) {
	negImpulse := mulW(normalMass, addW(vn, velocityBias))
	newImpulse := maxW(subW(normalImpulse, negImpulse), zero)
	impulse := subW(newImpulse, normalImpulse)
	approachImpulse := minW(maxW(negW(mulW(normalMass, vn)), zero), maxW(impulse, zero))
	allowance := subW(mulW(restitution, addW(compressionImpulse, approachImpulse)), restitutionImpulse)
	maxImpulse := addW(approachImpulse, maxW(allowance, zero))
	impulse = blendW(impulse, minW(impulse, maxImpulse), armed)

	return impulse, approachImpulse
}

func markContactHitEvent(contactSim *contactSim, hitEventBitSet *bitset.Set, negHitThreshold float32) bool {
	// Check for hit events to speed up serial processing later in the step
	if contactSim.simFlags&simEnableHitEvent == 0 {
		return false
	}

	m := &contactSim.manifold
	for k := range m.PointCount {
		mp := &m.Points[k]
		// Need to check total impulse because the point may be speculative and not colliding
		if mp.NormalVelocity < negHitThreshold && mp.TotalNormalImpulse > 0.0 {
			hitEventBitSet.SetBit(uint32(contactSim.contactID))
			return true
		}
	}

	return false
}
