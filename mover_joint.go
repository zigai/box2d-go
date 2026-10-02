// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type MoverJoint struct {
	Joint
}

// A mover joint is used to move a dynamic character mover through velocity commands.
// The x and y directions are handled separately. Does not affect rotation.
type MoverJointDef struct {
	// Base joint definition
	Base JointDef
	// The desired linear velocity
	LinearVelocity Vec2
	// The maximum motor force in newtons.
	MaxVelocityForce Vec2
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Set the desired relative linear velocity in meters per second
func (m MoverJoint) SetLinearVelocity(velocity Vec2) {
	if !checkInput(IsValidVec2(velocity), "b2IsValidVec2( velocity )") {
		return
	}

	world := getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMoverJointSetLinearVelocity(world.recording, &recArgsMoverJointSetLinearVelocity{joint: m.Joint, velocity: velocity})
	}

	joint := getJointSimCheckType(m.Joint, MoverJointType)
	joint.moverJoint.linearVelocity = velocity
}

// Get the desired relative linear velocity in meters per second
func (m MoverJoint) LinearVelocity() Vec2 {
	joint := getJointSimCheckType(m.Joint, MoverJointType)
	return joint.moverJoint.linearVelocity
}

// Set the motor joint maximum force, usually in newtons
func (m MoverJoint) SetMaxVelocityForce(maxForce Vec2) {
	if !checkInput(IsValidVec2(maxForce), "b2IsValidVec2( maxForce )") {
		return
	}

	world := getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMoverJointSetMaxVelocityForce(world.recording, &recArgsMoverJointSetMaxVelocityForce{joint: m.Joint, maxForce: maxForce})
	}

	joint := getJointSimCheckType(m.Joint, MoverJointType)
	joint.moverJoint.maxVelocityForce = maxForce
}

// Get the motor joint maximum force, usually in newtons
func (m MoverJoint) MaxVelocityForce() Vec2 {
	joint := getJointSimCheckType(m.Joint, MoverJointType)
	return joint.moverJoint.maxVelocityForce
}

func getMoverJointForce(world *world, base *jointSim) Vec2 {
	force := Vec2{X: float32((world.invH) * base.moverJoint.linearVelocityImpulse.X), Y: float32((world.invH) * base.moverJoint.linearVelocityImpulse.Y)}
	return force
}

func prepareMoverJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == MoverJointType, "base->type == b2_moverJoint")

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
	mB := bodySimB.invMass

	base.invMassA = mA
	base.invMassB = mB
	base.invIA = float32(0.0)
	base.invIB = float32(0.0)

	joint := &base.moverJoint
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

	k := mA + mB
	if k > float32(0.0) {
		joint.linearMass = float32(1.0) / k
	} else {
		joint.linearMass = float32(0.0)
	}

	if context.enableWarmStarting == false {
		joint.linearVelocityImpulse = Vec2{}
	}
}

func warmStartMoverJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == MoverJointType, "base->type == b2_moverJoint")

	mA := base.invMassA
	mB := base.invMassB

	joint := &base.moverJoint

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

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, joint.linearVelocityImpulse)
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, joint.linearVelocityImpulse)
	}
}

func solveMoverJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == MoverJointType, "base->type == b2_moverJoint")

	mA := base.invMassA
	mB := base.invMassB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.moverJoint
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
	vB := stateB.linearVelocity

	// linear velocity
	if joint.maxVelocityForce.X > float32(0.0) || joint.maxVelocityForce.Y > float32(0.0) {
		cdot := Sub(vB, vA)
		cdot = Sub(cdot, joint.linearVelocity)
		b := Vec2{X: float32((joint.linearMass) * cdot.X), Y: float32((joint.linearMass) * cdot.Y)}
		impulse := Vec2{X: -b.X, Y: -b.Y}

		oldImpulse := joint.linearVelocityImpulse
		joint.linearVelocityImpulse = Add(joint.linearVelocityImpulse, impulse)
		maxImpulse := Vec2{X: float32((context.h) * joint.maxVelocityForce.X), Y: float32((context.h) * joint.maxVelocityForce.Y)}

		// clamp x and y separately
		joint.linearVelocityImpulse = Clamp(joint.linearVelocityImpulse, Neg(maxImpulse), maxImpulse)

		impulse = Sub(joint.linearVelocityImpulse, oldImpulse)
		vA = MulSub(vA, mA, impulse)
		vB = MulAdd(vB, mB, impulse)
	} else {
		joint.linearVelocityImpulse = Vec2{}
	}

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = vA
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = vB
	}
}
