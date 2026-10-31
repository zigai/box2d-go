// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "github.com/zigai/box2d-go/internal/core"

var pogoAxis = Vec2{X: 0.0, Y: 1.0}

type PogoJoint struct {
	Joint
}

// Pogo joint definition.
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

// Set the rest length of a pogo joint.
func (p PogoJoint) SetRestLength(length float32) {
	checkInput(IsValidFloat(length), "b2IsValidFloat( length )")

	world := p.Joint.world
	if world.recording != nil {
		recWritePogoJointSetRestLength(
			world.recording,
			&recArgsPogoJointSetRestLength{joint: p.Joint, length: length},
		)
	}

	joint := getJointSimCheckType(p.Joint, PogoJointType)
	joint.pogoJoint.restLength = length
}

// Get the rest length of a pogo joint.
func (p PogoJoint) RestLength() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.restLength
}

// Get the spring Hertz.
func (p PogoJoint) SpringHertz() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.hertz
}

// Set the spring stiffness in Hertz.
func (p PogoJoint) SetSpringHertz(hertz float32) {
	checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")

	world := p.Joint.world
	if world.recording != nil {
		recWritePogoJointSetSpringHertz(
			world.recording,
			&recArgsPogoJointSetSpringHertz{joint: p.Joint, hertz: hertz},
		)
	}

	joint := getJointSimCheckType(p.Joint, PogoJointType)
	joint.pogoJoint.hertz = hertz
}

// Set the spring damping ratio, non-dimensional.
func (p PogoJoint) SetSpringDampingRatio(dampingRatio float32) {
	checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )")

	world := p.Joint.world
	if world.recording != nil {
		recWritePogoJointSetSpringDampingRatio(
			world.recording,
			&recArgsPogoJointSetSpringDampingRatio{joint: p.Joint, dampingRatio: dampingRatio},
		)
	}

	joint := getJointSimCheckType(p.Joint, PogoJointType)
	joint.pogoJoint.dampingRatio = dampingRatio
}

// Get the spring damping ratio.
func (p PogoJoint) SpringDampingRatio() float32 {
	joint := getJointSimCheckType(p.Joint, PogoJointType)
	return joint.pogoJoint.dampingRatio
}

// Get the current length of the pogo.
func (p PogoJoint) Length() float32 {
	world := p.Joint.world
	jointSim := getJointSimCheckType(p.Joint, PogoJointType)

	// Relative to body A so the difference stays in float precision far from the origin
	wxfA := getBodyTransform(world, jointSim.bodyIDA)
	transformA := wxfA.Relative(wxfA.P)
	transformB := getBodyTransform(world, jointSim.bodyIDB).Relative(wxfA.P)

	axis := jointSim.localFrameB.Q.Rotate(pogoAxis)
	axis = transformB.Q.Rotate(axis)
	pA := transformA.Apply(jointSim.localFrameA.P)
	pB := transformB.Apply(jointSim.localFrameB.P)
	d := pB.Sub(pA)
	length := d.Dot(axis)

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

func getPogoJointForce(world *World, base *jointSim) Vec2 {
	force := Vec2{
		X: float32(float32(world.invH*base.pogoJoint.impulse) * base.pogoJoint.normal.X),
		Y: float32(float32(world.invH*base.pogoJoint.impulse) * base.pogoJoint.normal.Y),
	}

	return force
}

func preparePogoJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == PogoJointType, "base->type == b2_pogoJoint")

	world := context.world

	// chase body id to the solver set where the body lives
	idA := base.bodyIDA
	idB := base.bodyIDB

	bodyA := &world.bodies[idA]
	bodyB := &world.bodies[idB]

	core.Assert(
		bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet,
		"bodyA->setIndex == b2_awakeSet || bodyB->setIndex == b2_awakeSet",
	)

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
	joint.frameA.Q = bodySimA.transform.Q.Mul(base.localFrameA.Q)
	joint.frameA.P = bodySimA.transform.Q.Rotate(base.localFrameA.P.Sub(bodySimA.localCenter))
	joint.frameB.Q = bodySimB.transform.Q.Mul(base.localFrameB.Q)
	joint.frameB.P = bodySimB.transform.Q.Rotate(base.localFrameB.P.Sub(bodySimB.localCenter))

	// Compute the initial center delta. Incremental position updates are relative to this.
	joint.deltaCenter = bodySimB.center.Sub(bodySimA.center)

	rA := joint.frameA.P
	rB := joint.frameB.P

	// compute effective mass
	crA := rA.Cross(joint.normal)
	crB := rB.Cross(joint.normal)

	k := mA + mB + float32(float32(iA*crA)*crA) + float32(float32(iB*crB)*crB)
	if k > float32(0.0) {
		joint.linearMass = float32(1.0) / k
	} else {
		joint.linearMass = float32(0.0)
	}

	if !context.enableWarmStarting {
		joint.impulse = float32(0.0)
	}
}

func warmStartPogoJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == PogoJointType, "base->type == b2_pogoJoint")

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

	rA := stateA.deltaRotation.Rotate(joint.frameA.P)
	rB := stateB.deltaRotation.Rotate(joint.frameB.P)

	linearImpulse := Vec2{
		X: float32(joint.impulse * joint.normal.X),
		Y: float32(joint.impulse * joint.normal.Y),
	}

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = stateA.linearVelocity.MulSub(mA, linearImpulse)
		stateA.angularVelocity -= float32(iA * rA.Cross(linearImpulse))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = stateB.linearVelocity.MulAdd(mB, linearImpulse)
		stateB.angularVelocity += float32(iB * rB.Cross(linearImpulse))
	}
}

func solvePogoJoint(base *jointSim, context *stepContext, useBias bool) {
	core.Assert(base.jointType == PogoJointType, "base->type == b2_pogoJoint")

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

	rA := stateA.deltaRotation.Rotate(joint.frameA.P)
	rB := stateB.deltaRotation.Rotate(joint.frameB.P)

	bias := float32(0.0)

	if useBias {
		// This is a custom pogo position correction designed to avoid feeding velocity into
		// the mover body. Otherwise the mover can make huge jumps going up stairs. This still
		// works with continuous collision.
		// However, this position delta is not observer contact forces and can lead to tunneling.
		dcA := stateA.deltaPosition
		dcB := stateB.deltaPosition
		d := dcB.Sub(dcA).Add(rB.Sub(rA)).Add(joint.deltaCenter)

		// The pogo axis is usually fixed as the world y-axis. Different than the ground normal.
		pogoAxis := joint.frameB.Q.Rotate(pogoAxis)
		c := pogoAxis.Dot(d) - joint.restLength

		// This could be divided by dot(joint->normal, pogoAxis) to account for the constraint
		// direction, but I'd rather diminish the pogo recovery rate than risk making it huge.
		joint.velocity = SpringDamper(joint.hertz, joint.dampingRatio, c, joint.velocity, context.h)
		bias = -joint.velocity
	}

	vr := vB.Add(Vec2{X: float32(-wB * rB.Y), Y: float32(wB * rB.X)}).
		Sub(vA.Add(Vec2{X: float32(-wA * rA.Y), Y: float32(wA * rA.X)}))
	cdot := joint.normal.Dot(vr)

	maxTensionImpulse := float32(context.h * joint.maxTensionForce)
	maxCompressionImpulse := float32(context.h * joint.maxCompressionForce)
	oldImpulse := joint.impulse
	impulse := float32(-joint.linearMass * (cdot + bias))
	joint.impulse = clampFloat(joint.impulse+impulse, -maxTensionImpulse, maxCompressionImpulse)
	impulse = joint.impulse - oldImpulse

	impulseVector := Vec2{
		X: float32(impulse * joint.normal.X),
		Y: float32(impulse * joint.normal.Y),
	}
	vA = vA.MulSub(mA, impulseVector)
	wA -= float32(iA * rA.Cross(impulseVector))
	vB = vB.MulAdd(mB, impulseVector)
	wB += float32(iB * rB.Cross(impulseVector))

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = vA
		stateA.angularVelocity = wA
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = vB
		stateB.angularVelocity = wB
	}
}
