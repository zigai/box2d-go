// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type WeldJoint struct {
	Joint
}

// Weld joint definition
// Connects two bodies together rigidly. This constraint provides springs to mimic
// soft-body simulation.
type WeldJointDef struct {
	// Base joint definition
	Base JointDef
	// Linear stiffness expressed as Hertz (cycles per second). Use zero for maximum stiffness.
	LinearHertz float32
	// Angular stiffness as Hertz (cycles per second). Use zero for maximum stiffness.
	AngularHertz float32
	// Linear damping ratio, non-dimensional. Use 1 for critical damping.
	LinearDampingRatio float32
	// Linear damping ratio, non-dimensional. Use 1 for critical damping.
	AngularDampingRatio float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Set the weld joint linear stiffness in Hertz. 0 is rigid.
func (w WeldJoint) SetLinearHertz(hertz float32) {
	if !checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWeldJointSetLinearHertz(world.recording, &recArgsWeldJointSetLinearHertz{joint: w.Joint, hertz: hertz})
	}
	assert(IsValidFloat(hertz) && hertz >= float32(0.0), "b2IsValidFloat( hertz ) && hertz >= 0.0f")
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	joint.weldJoint.linearHertz = hertz
}

// Get the weld joint linear stiffness in Hertz
func (w WeldJoint) LinearHertz() float32 {
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	return joint.weldJoint.linearHertz
}

// Set the weld joint linear damping ratio (non-dimensional)
func (w WeldJoint) SetLinearDampingRatio(dampingRatio float32) {
	if !checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWeldJointSetLinearDampingRatio(world.recording, &recArgsWeldJointSetLinearDampingRatio{joint: w.Joint, dampingRatio: dampingRatio})
	}
	assert(IsValidFloat(dampingRatio) && dampingRatio >= float32(0.0), "b2IsValidFloat( dampingRatio ) && dampingRatio >= 0.0f")
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	joint.weldJoint.linearDampingRatio = dampingRatio
}

// Get the weld joint linear damping ratio (non-dimensional)
func (w WeldJoint) LinearDampingRatio() float32 {
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	return joint.weldJoint.linearDampingRatio
}

// Set the weld joint angular stiffness in Hertz. 0 is rigid.
func (w WeldJoint) SetAngularHertz(hertz float32) {
	if !checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWeldJointSetAngularHertz(world.recording, &recArgsWeldJointSetAngularHertz{joint: w.Joint, hertz: hertz})
	}
	assert(IsValidFloat(hertz) && hertz >= float32(0.0), "b2IsValidFloat( hertz ) && hertz >= 0.0f")
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	joint.weldJoint.angularHertz = hertz
}

// Get the weld joint angular stiffness in Hertz
func (w WeldJoint) AngularHertz() float32 {
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	return joint.weldJoint.angularHertz
}

// Set weld joint angular damping ratio, non-dimensional
func (w WeldJoint) SetAngularDampingRatio(dampingRatio float32) {
	if !checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWeldJointSetAngularDampingRatio(world.recording, &recArgsWeldJointSetAngularDampingRatio{joint: w.Joint, dampingRatio: dampingRatio})
	}
	assert(IsValidFloat(dampingRatio) && dampingRatio >= float32(0.0), "b2IsValidFloat( dampingRatio ) && dampingRatio >= 0.0f")
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	joint.weldJoint.angularDampingRatio = dampingRatio
}

// Get the weld joint angular damping ratio, non-dimensional
func (w WeldJoint) AngularDampingRatio() float32 {
	joint := getJointSimCheckType(w.Joint, WeldJointType)
	return joint.weldJoint.angularDampingRatio
}

func getWeldJointForce(world *world, base *jointSim) Vec2 {
	force := Vec2{X: float32((world.invH) * base.weldJoint.linearImpulse.X), Y: float32((world.invH) * base.weldJoint.linearImpulse.Y)}
	return force
}

func getWeldJointTorque(world *world, base *jointSim) float32 {
	return float32(world.invH * base.weldJoint.angularImpulse)
}

func prepareWeldJoint(base *jointSim, context *stepContext) {
	// Point-to-point constraint
	// C = p2 - p1
	// Cdot = v2 - v1
	//      = v2 + cross(w2, r2) - v1 - cross(w1, r1)
	// J = [-E -r1_skew E r2_skew ]
	// Identity used:
	// w k % (rx i + ry j) = w * (-ry i + rx j)
	// Angle constraint
	// C = angle2 - angle1 - referenceAngle
	// Cdot = w2 - w1
	// J = [0 0 -1 0 0 1]
	// K = invI1 + invI2
	// 3x3 Block
	// K = [J1] * invM * [J1T J2T]
	//     [J2]
	//   = [J1] * [invM * J1T invM * J2T]
	//     [J2]
	//   = [J1 * invM * J1T J1 * invM * J2T]
	//     [J2 * invM * J1T J2 * invM * J2T]

	assert(base.jointType == WeldJointType, "base->type == b2_weldJoint")

	// chase body id to the solver set where the body lives
	idA := base.bodyIdA
	idB := base.bodyIdB

	world := context.world

	bodyA := &world.bodies[idA]
	bodyB := &world.bodies[idB]

	assert(bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet, "bodyA->setIndex == b2_awakeSet || bodyB->setIndex == b2_awakeSet")
	setA := &world.solverSets[bodyA.setIndex]
	setB := &world.solverSets[bodyB.setIndex]

	localIndexA := bodyA.localIndex
	localIndexB := bodyB.localIndex

	bodySimA := &setA.bodySims[localIndexA]
	bodySimB := &setB.bodySims[localIndexB]

	mA := bodySimA.invMass
	iA := bodySimA.invInertia
	mB := bodySimB.invMass
	iB := bodySimB.invInertia

	base.invMassA = mA
	base.invMassB = mB
	base.invIA = iA
	base.invIB = iB

	joint := &base.weldJoint
	if bodyA.setIndex == awakeSet {
		joint.indexA = localIndexA
	} else {
		joint.indexA = nullIndex
	}
	if bodyB.setIndex == awakeSet {
		joint.indexB = localIndexB
	} else {
		joint.indexB = nullIndex
	}

	// Compute joint anchor frames with world space rotation, relative to center of mass
	joint.frameA.Q = MulRot(bodySimA.transform.Q, base.localFrameA.Q)
	joint.frameA.P = RotateVector(bodySimA.transform.Q, Sub(base.localFrameA.P, bodySimA.localCenter))
	joint.frameB.Q = MulRot(bodySimB.transform.Q, base.localFrameB.Q)
	joint.frameB.P = RotateVector(bodySimB.transform.Q, Sub(base.localFrameB.P, bodySimB.localCenter))

	// Compute the initial center delta. Incremental position updates are relative to this.
	joint.deltaCenter = SubPos(bodySimB.center, bodySimA.center)

	ka := iA + iB
	if ka > float32(0.0) {
		joint.axialMass = float32(1.0) / ka
	} else {
		joint.axialMass = float32(0.0)
	}

	if joint.linearHertz == float32(0.0) {
		joint.linearSpring = base.constraintSoftness
	} else {
		joint.linearSpring = makeSoft(joint.linearHertz, joint.linearDampingRatio, context.h)
	}

	if joint.angularHertz == float32(0.0) {
		joint.angularSpring = base.constraintSoftness
	} else {
		joint.angularSpring = makeSoft(joint.angularHertz, joint.angularDampingRatio, context.h)
	}

	if context.enableWarmStarting == false {
		joint.linearImpulse = Vec2{}
		joint.angularImpulse = float32(0.0)
	}
}

func warmStartWeldJoint(base *jointSim, context *stepContext) {
	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.weldJoint

	var stateA *bodyState
	if joint.indexA == nullIndex {
		stateA = &dummyState
	} else {
		stateA = &context.states[joint.indexA]
	}
	var stateB *bodyState
	if joint.indexB == nullIndex {
		stateB = &dummyState
	} else {
		stateB = &context.states[joint.indexB]
	}

	rA := RotateVector(stateA.deltaRotation, joint.frameA.P)
	rB := RotateVector(stateB.deltaRotation, joint.frameB.P)

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, joint.linearImpulse)
		stateA.angularVelocity -= float32(iA * (Cross(rA, joint.linearImpulse) + joint.angularImpulse))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, joint.linearImpulse)
		stateB.angularVelocity += float32(iB * (Cross(rB, joint.linearImpulse) + joint.angularImpulse))
	}
}

func solveWeldJoint(base *jointSim, context *stepContext, useBias bool) {
	assert(base.jointType == WeldJointType, "base->type == b2_weldJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.weldJoint

	var stateA *bodyState
	if joint.indexA == nullIndex {
		stateA = &dummyState
	} else {
		stateA = &context.states[joint.indexA]
	}
	var stateB *bodyState
	if joint.indexB == nullIndex {
		stateB = &dummyState
	} else {
		stateB = &context.states[joint.indexB]
	}

	vA := stateA.linearVelocity
	wA := stateA.angularVelocity
	vB := stateB.linearVelocity
	wB := stateB.angularVelocity

	// Block solve doesn't work correctly with mixed stiffness values

	// angular constraint
	{
		qA := MulRot(stateA.deltaRotation, joint.frameA.Q)
		qB := MulRot(stateB.deltaRotation, joint.frameB.Q)
		relQ := InvMulRot(qA, qB)
		jointAngle := relQ.Angle()

		bias := float32(0.0)
		massScale := float32(1.0)
		impulseScale := float32(0.0)
		if useBias || joint.angularHertz > float32(0.0) {
			C := jointAngle
			bias = float32(joint.angularSpring.biasRate * C)
			massScale = joint.angularSpring.massScale
			impulseScale = joint.angularSpring.impulseScale
		}

		Cdot := wB - wA
		impulse := float32(float32(-massScale*joint.axialMass)*(Cdot+bias)) - float32(impulseScale*joint.angularImpulse)
		joint.angularImpulse += impulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	// linear constraint
	{
		rA := RotateVector(stateA.deltaRotation, joint.frameA.P)
		rB := RotateVector(stateB.deltaRotation, joint.frameB.P)

		bias := Vec2{}
		massScale := float32(1.0)
		impulseScale := float32(0.0)
		if useBias || joint.linearHertz > float32(0.0) {
			dcA := stateA.deltaPosition
			dcB := stateB.deltaPosition
			C := Add(Add(Sub(dcB, dcA), Sub(rB, rA)), joint.deltaCenter)

			bias = Vec2{X: float32((joint.linearSpring.biasRate) * C.X), Y: float32((joint.linearSpring.biasRate) * C.Y)}
			massScale = joint.linearSpring.massScale
			impulseScale = joint.linearSpring.impulseScale
		}

		Cdot := Sub(Add(vB, Vec2{X: float32(-(wB) * rB.Y), Y: float32((wB) * rB.X)}), Add(vA, Vec2{X: float32(-(wA) * rA.Y), Y: float32((wA) * rA.X)}))

		var K Mat22
		K.Cx.X = mA + mB + float32(float32(rA.Y*rA.Y)*iA) + float32(float32(rB.Y*rB.Y)*iB)
		K.Cy.X = float32(float32(-rA.Y*rA.X)*iA) - float32(float32(rB.Y*rB.X)*iB)
		K.Cx.Y = K.Cy.X
		K.Cy.Y = mA + mB + float32(float32(rA.X*rA.X)*iA) + float32(float32(rB.X*rB.X)*iB)
		b := Solve22(K, Add(Cdot, bias))
		b.X = float32(b.X)
		b.Y = float32(b.Y)

		impulse := Vec2{X: float32(-massScale*b.X) - float32(impulseScale*joint.linearImpulse.X), Y: float32(-massScale*b.Y) - float32(impulseScale*joint.linearImpulse.Y)}

		joint.linearImpulse = Add(joint.linearImpulse, impulse)

		vA = MulSub(vA, mA, impulse)
		wA -= float32(iA * Cross(rA, impulse))
		vB = MulAdd(vB, mB, impulse)
		wB += float32(iB * Cross(rB, impulse))
	}

	assert(IsValidVec2(vA), "b2IsValidVec2( vA )")
	assert(IsValidFloat(wA), "b2IsValidFloat( wA )")
	assert(IsValidVec2(vB), "b2IsValidVec2( vB )")
	assert(IsValidFloat(wB), "b2IsValidFloat( wB )")

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = vA
		stateA.angularVelocity = wA
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = vB
		stateB.angularVelocity = wB
	}
}

func drawWeldJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	assert(base.jointType == WeldJointType, "base->type == b2_weldJoint")

	frameA := MulWorldTransforms(transformA, base.localFrameA)
	frameB := MulWorldTransforms(transformB, base.localFrameB)

	box := MakeBox(float32(float32(0.25)*drawScale), float32(float32(0.125)*drawScale))
	draw.DrawPolygonFunc(frameA, box.Vertices[:4], ColorDarkOrange)
	draw.DrawPolygonFunc(frameB, box.Vertices[:4], ColorDarkCyan)
}
