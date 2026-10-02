// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type PogoJoint struct {
	Joint
}

// Pogo joint definition
type PogoJointDef struct {
	// Base joint definition
	Base JointDef
	// Normal vector at the contact
	Normal Vec2
	// The spring stiffness Hertz, cycles per second
	Hertz float32
	// The spring damping ratio, non-dimensional
	DampingRatio float32
	// Spring length.
	RestLength float32
	// Maximum tension force. Controls how sticky the mover is to ground.
	MaxTensionForce float32
	// Maximum compression force. How hard the pogo pushes the mover up from the ground.
	MaxCompressionForce float32
	// The initial pogo internal impulse. The pogo is typically recreated every frame. This
	// gives a way to move the internal state forward for smoother spring behavior.
	Impulse float32
	// The initial pogo internal velocity. The pogo is typically recreated every frame. This
	// gives a way to move the internal state forward for smoother spring behavior.
	Velocity float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

var pogoAxis = Vec2{X: 0.0, Y: 1.0}

// Set the rest length of a pogo joint
func (p PogoJoint) SetRestLength(length float32) {
	if !checkInput(IsValidFloat(length), "b2IsValidFloat( length )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePogoJointSetRestLength(world.recording, &recArgsPogoJointSetRestLength{joint: p.Joint, length: length})
	}

	joint := getJointSimCheckType(p.Joint, PogoJointType)
	joint.pogoJoint.restLength = length
}

// Get the rest length of a pogo joint
func (p PogoJoint) RestLength() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.restLength
}

// Get the spring Hertz
func (p PogoJoint) SpringHertz() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.hertz
}

// Set the spring stiffness in Hertz
func (p PogoJoint) SetSpringHertz(hertz float32) {
	if !checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePogoJointSetSpringHertz(world.recording, &recArgsPogoJointSetSpringHertz{joint: p.Joint, hertz: hertz})
	}

	joint := getJointSimCheckType(p.Joint, PogoJointType)
	joint.pogoJoint.hertz = hertz
}

// Set the spring damping ratio, non-dimensional
func (p PogoJoint) SetSpringDampingRatio(dampingRatio float32) {
	if !checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePogoJointSetSpringDampingRatio(world.recording, &recArgsPogoJointSetSpringDampingRatio{joint: p.Joint, dampingRatio: dampingRatio})
	}

	joint := getJointSimCheckType(p.Joint, PogoJointType)
	joint.pogoJoint.dampingRatio = dampingRatio
}

// Get the spring damping ratio
func (p PogoJoint) SpringDampingRatio() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.dampingRatio
}

// Get the current length of the pogo.
func (p PogoJoint) Length() float32 {
	world := getWorld(int(p.Joint.world0))
	jointSim := getJointSimCheckType(p.Joint, PogoJointType)

	// Relative to body A so the difference stays in float precision far from the origin
	wxfA := getBodyTransform(world, jointSim.bodyIdA)
	transformA := ToRelativeTransform(wxfA, wxfA.P)
	transformB := ToRelativeTransform(getBodyTransform(world, jointSim.bodyIdB), wxfA.P)

	axis := RotateVector(jointSim.localFrameB.Q, pogoAxis)
	axis = RotateVector(transformB.Q, axis)
	pA := TransformPoint(transformA, jointSim.localFrameA.P)
	pB := TransformPoint(transformB, jointSim.localFrameB.P)
	d := Sub(pB, pA)
	length := Dot(d, axis)
	return length
}

// Get the internal pogo impulse.
func (p PogoJoint) Impulse() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.impulse
}

// Get the internal pogo velocity.
func (p PogoJoint) Velocity() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.velocity
}

func getPogoJointForce(world *world, base *jointSim) Vec2 {
	force := Vec2{X: float32((float32(world.invH * base.pogoJoint.impulse)) * base.pogoJoint.normal.X), Y: float32((float32(world.invH * base.pogoJoint.impulse)) * base.pogoJoint.normal.Y)}
	return force
}

func preparePogoJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == PogoJointType, "base->type == b2_pogoJoint")

	world := context.world

	// chase body id to the solver set where the body lives
	idA := base.bodyIdA
	idB := base.bodyIdB

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

	joint := &base.pogoJoint
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

	rA := joint.frameA.P
	rB := joint.frameB.P

	// compute effective mass
	crA := Cross(rA, joint.normal)
	crB := Cross(rB, joint.normal)
	k := mA + mB + float32(float32(iA*crA)*crA) + float32(float32(iB*crB)*crB)
	if k > float32(0.0) {
		joint.linearMass = float32(1.0) / k
	} else {
		joint.linearMass = float32(0.0)
	}

	if context.enableWarmStarting == false {
		joint.impulse = float32(0.0)
	}
}

func warmStartPogoJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == PogoJointType, "base->type == b2_pogoJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	joint := &base.pogoJoint

	if joint.hertz == float32(0.0) {
		joint.impulse = float32(0.0)
		return
	}

	// dummy state for static bodies
	dummyState := identityBodyState

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

	linearImpulse := Vec2{X: float32((joint.impulse) * joint.normal.X), Y: float32((joint.impulse) * joint.normal.Y)}

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, linearImpulse)
		stateA.angularVelocity -= float32(iA * Cross(rA, linearImpulse))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, linearImpulse)
		stateB.angularVelocity += float32(iB * Cross(rB, linearImpulse))
	}
}

func solvePogoJoint(base *jointSim, context *stepContext, useBias bool) {
	assert(base.jointType == PogoJointType, "base->type == b2_pogoJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.pogoJoint
	if joint.hertz == float32(0.0) {
		joint.impulse = float32(0.0)
		return
	}

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

	rA := RotateVector(stateA.deltaRotation, joint.frameA.P)
	rB := RotateVector(stateB.deltaRotation, joint.frameB.P)

	bias := float32(0.0)
	if useBias {
		// This is a custom pogo position correction designed to avoid feeding velocity into
		// the mover body. Otherwise the mover can make huge jumps going up stairs. This still
		// works with continuous collision.
		// However, this position delta is not observer contact forces and can lead to tunnelling.
		dcA := stateA.deltaPosition
		dcB := stateB.deltaPosition
		d := Add(Add(Sub(dcB, dcA), Sub(rB, rA)), joint.deltaCenter)

		// The pogo axis is usually fixed as the world y-axis. Different than the ground normal.
		pogoAxis := RotateVector(joint.frameB.Q, pogoAxis)
		c := Dot(pogoAxis, d) - joint.restLength

		// This could be divided by dot(joint->normal, pogoAxis) to account for the constraint
		// direction, but I'd rather diminish the pogo recovery rate than risk making it huge.
		joint.velocity = SpringDamper(joint.hertz, joint.dampingRatio, c, joint.velocity, context.h)
		bias = -joint.velocity
	}

	vr := Sub(Add(vB, Vec2{X: float32(-(wB) * rB.Y), Y: float32((wB) * rB.X)}), Add(vA, Vec2{X: float32(-(wA) * rA.Y), Y: float32((wA) * rA.X)}))
	cdot := Dot(joint.normal, vr)

	maxTensionImpulse := float32(context.h * joint.maxTensionForce)
	maxCompressionImpulse := float32(context.h * joint.maxCompressionForce)
	oldImpulse := joint.impulse
	impulse := float32(-joint.linearMass * (cdot + bias))
	joint.impulse = ClampFloat(joint.impulse+impulse, -maxTensionImpulse, maxCompressionImpulse)
	impulse = joint.impulse - oldImpulse

	P := Vec2{X: float32((impulse) * joint.normal.X), Y: float32((impulse) * joint.normal.Y)}
	vA = MulSub(vA, mA, P)
	wA -= float32(iA * Cross(rA, P))
	vB = MulAdd(vB, mB, P)
	wB += float32(iB * Cross(rB, P))

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = vA
		stateA.angularVelocity = wA
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = vB
		stateB.angularVelocity = wB
	}
}
