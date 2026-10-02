// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type MotorJoint struct {
	Joint
}

// A motor joint is used to control the relative velocity and or transform between two bodies.
// With a velocity of zero this acts like top-down friction.
type MotorJointDef struct {
	// Base joint definition
	Base JointDef
	// The desired linear velocity
	LinearVelocity Vec2
	// The maximum motor force in newtons
	MaxVelocityForce float32
	// The desired angular velocity
	AngularVelocity float32
	// The maximum motor torque in newton-meters
	MaxVelocityTorque float32
	// Linear spring hertz for position control
	LinearHertz float32
	// Linear spring damping ratio
	LinearDampingRatio float32
	// Maximum spring force in newtons
	MaxSpringForce float32
	// Angular spring hertz for position control
	AngularHertz float32
	// Angular spring damping ratio
	AngularDampingRatio float32
	// Maximum spring torque in newton-meters
	MaxSpringTorque float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Set the desired relative linear velocity in meters per second
func (m MotorJoint) SetLinearVelocity(velocity Vec2) {
	if !(checkInput(IsValidVec2(velocity), "b2IsValidVec2( velocity )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetLinearVelocity(world.recording, &recArgsMotorJointSetLinearVelocity{joint: m.Joint, velocity: velocity})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.linearVelocity = velocity
}

// Get the desired relative linear velocity in meters per second
func (m MotorJoint) LinearVelocity() Vec2 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.linearVelocity
}

// Set the desired relative angular velocity in radians per second
func (m MotorJoint) SetAngularVelocity(velocity float32) {
	if !(checkInput(IsValidFloat(velocity), "b2IsValidFloat( velocity )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetAngularVelocity(world.recording, &recArgsMotorJointSetAngularVelocity{joint: m.Joint, velocity: velocity})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.angularVelocity = velocity
}

// Get the desired relative angular velocity in radians per second
func (m MotorJoint) AngularVelocity() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.angularVelocity
}

// Set the motor joint maximum torque, usually in newton-meters
func (m MotorJoint) SetMaxVelocityTorque(maxTorque float32) {
	if !(checkInput(IsValidFloat(maxTorque), "b2IsValidFloat( maxTorque )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetMaxVelocityTorque(world.recording, &recArgsMotorJointSetMaxVelocityTorque{joint: m.Joint, maxTorque: maxTorque})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.maxVelocityTorque = maxTorque
}

// Get the motor joint maximum torque, usually in newton-meters
func (m MotorJoint) MaxVelocityTorque() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxVelocityTorque
}

// Set the motor joint maximum force, usually in newtons
func (m MotorJoint) SetMaxVelocityForce(maxForce float32) {
	if !(checkInput(IsValidFloat(maxForce), "b2IsValidFloat( maxForce )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetMaxVelocityForce(world.recording, &recArgsMotorJointSetMaxVelocityForce{joint: m.Joint, maxForce: maxForce})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.maxVelocityForce = maxForce
}

// Get the motor joint maximum force, usually in newtons
func (m MotorJoint) MaxVelocityForce() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxVelocityForce
}

// Set the spring linear hertz stiffness
func (m MotorJoint) SetLinearHertz(hertz float32) {
	if !(checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetLinearHertz(world.recording, &recArgsMotorJointSetLinearHertz{joint: m.Joint, hertz: hertz})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.linearHertz = hertz
}

// Get the spring linear hertz stiffness
func (m MotorJoint) LinearHertz() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.linearHertz
}

// Set the spring linear damping ratio. Use 1.0 for critical damping.
func (m MotorJoint) SetLinearDampingRatio(damping float32) {
	if !(checkInput(IsValidFloat(damping), "b2IsValidFloat( damping )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetLinearDampingRatio(world.recording, &recArgsMotorJointSetLinearDampingRatio{joint: m.Joint, damping: damping})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.linearDampingRatio = damping
}

// Get the spring linear damping ratio.
func (m MotorJoint) LinearDampingRatio() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.linearDampingRatio
}

// Set the spring angular hertz stiffness
func (m MotorJoint) SetAngularHertz(hertz float32) {
	if !(checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetAngularHertz(world.recording, &recArgsMotorJointSetAngularHertz{joint: m.Joint, hertz: hertz})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.angularHertz = hertz
}

// Get the spring angular hertz stiffness
func (m MotorJoint) AngularHertz() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.angularHertz
}

// Set the spring angular damping ratio. Use 1.0 for critical damping.
func (m MotorJoint) SetAngularDampingRatio(damping float32) {
	if !(checkInput(IsValidFloat(damping), "b2IsValidFloat( damping )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetAngularDampingRatio(world.recording, &recArgsMotorJointSetAngularDampingRatio{joint: m.Joint, damping: damping})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.angularDampingRatio = damping
}

// Get the spring angular damping ratio.
func (m MotorJoint) AngularDampingRatio() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.angularDampingRatio
}

// Set the maximum spring force in newtons.
func (m MotorJoint) SetMaxSpringForce(maxForce float32) {
	if !(checkInput(IsValidFloat(maxForce), "b2IsValidFloat( maxForce )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetMaxSpringForce(world.recording, &recArgsMotorJointSetMaxSpringForce{joint: m.Joint, maxForce: maxForce})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.maxSpringForce = MaxFloat(0.0, maxForce)
}

// Get the maximum spring force in newtons.
func (m MotorJoint) MaxSpringForce() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxSpringForce
}

// Set the maximum spring torque in newtons * meters
func (m MotorJoint) SetMaxSpringTorque(maxTorque float32) {
	if !(checkInput(IsValidFloat(maxTorque), "b2IsValidFloat( maxTorque )")) {
		return
	}

	var world *world = getWorld(int(m.Joint.world0))
	if world.recording != nil {
		recWriteMotorJointSetMaxSpringTorque(world.recording, &recArgsMotorJointSetMaxSpringTorque{joint: m.Joint, maxTorque: maxTorque})
	}
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	jnt.motorJoint.maxSpringTorque = MaxFloat(0.0, maxTorque)
}

// Get the maximum spring torque in newtons * meters
func (m MotorJoint) MaxSpringTorque() float32 {
	var jnt *jointSim = getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxSpringTorque
}

func getMotorJointForce(world *world, base *jointSim) Vec2 {
	var force Vec2 = MulSV(world.invH, Add(base.motorJoint.linearVelocityImpulse, base.motorJoint.linearSpringImpulse))
	return force
}

func getMotorJointTorque(world *world, base *jointSim) float32 {
	return float32(world.invH * (base.motorJoint.angularVelocityImpulse + base.motorJoint.angularSpringImpulse))
}

func prepareMotorJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == MotorJointType, "base->type == b2_motorJoint")

	// chase body id to the solver set where the body lives
	var idA int = base.bodyIdA
	var idB int = base.bodyIdB

	var world *world = context.world

	var bodyA *body = &world.bodies[idA]
	var bodyB *body = &world.bodies[idB]

	assert(bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet, "bodyA->setIndex == b2_awakeSet || bodyB->setIndex == b2_awakeSet")

	var setA *solverSet = &world.solverSets[bodyA.setIndex]
	var setB *solverSet = &world.solverSets[bodyB.setIndex]

	var localIndexA int = bodyA.localIndex
	var localIndexB int = bodyB.localIndex

	var bodySimA *bodySim = &setA.bodySims[localIndexA]
	var bodySimB *bodySim = &setB.bodySims[localIndexB]

	var mA float32 = bodySimA.invMass
	var iA float32 = bodySimA.invInertia
	var mB float32 = bodySimB.invMass
	var iB float32 = bodySimB.invInertia

	base.invMassA = mA
	base.invMassB = mB
	base.invIA = iA
	base.invIB = iB

	var jnt *motorJoint = &base.motorJoint
	if bodyA.setIndex == awakeSet {
		jnt.indexA = localIndexA
	} else {
		jnt.indexA = nullIndex
	}
	if bodyB.setIndex == awakeSet {
		jnt.indexB = localIndexB
	} else {
		jnt.indexB = nullIndex
	}

	// Compute joint anchor frames with world space rotation, relative to center of mass
	jnt.frameA.Q = MulRot(bodySimA.transform.Q, base.localFrameA.Q)
	jnt.frameA.P = RotateVector(bodySimA.transform.Q, Sub(base.localFrameA.P, bodySimA.localCenter))
	jnt.frameB.Q = MulRot(bodySimB.transform.Q, base.localFrameB.Q)
	jnt.frameB.P = RotateVector(bodySimB.transform.Q, Sub(base.localFrameB.P, bodySimB.localCenter))

	// Compute the initial center delta. Incremental position updates are relative to this.
	jnt.deltaCenter = SubPos(bodySimB.center, bodySimA.center)

	var rA Vec2 = jnt.frameA.P
	var rB Vec2 = jnt.frameB.P

	jnt.linearSpring = makeSoft(jnt.linearHertz, jnt.linearDampingRatio, context.h)
	jnt.angularSpring = makeSoft(jnt.angularHertz, jnt.angularDampingRatio, context.h)

	var kl Mat22
	kl.Cx.X = mA + mB + float32(float32(rA.Y*rA.Y)*iA) + float32(float32(rB.Y*rB.Y)*iB)
	kl.Cx.Y = float32(float32(-rA.Y*rA.X)*iA) - float32(float32(rB.Y*rB.X)*iB)
	kl.Cy.X = kl.Cx.Y
	kl.Cy.Y = mA + mB + float32(float32(rA.X*rA.X)*iA) + float32(float32(rB.X*rB.X)*iB)
	jnt.linearMass = GetInverse22(kl)

	var ka float32 = iA + iB
	if ka > 0.0 {
		jnt.angularMass = 1.0 / ka
	} else {
		jnt.angularMass = 0.0
	}

	if context.enableWarmStarting == false {
		jnt.linearVelocityImpulse = Vec2Zero
		jnt.angularVelocityImpulse = 0.0
		jnt.linearSpringImpulse = Vec2Zero
		jnt.angularSpringImpulse = 0.0
	}
}

func warmStartMotorJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == MotorJointType, "base->type == b2_motorJoint")

	var mA float32 = base.invMassA
	var mB float32 = base.invMassB
	var iA float32 = base.invIA
	var iB float32 = base.invIB

	var jnt *motorJoint = &base.motorJoint

	// dummy state for static bodies
	var dummyState bodyState = identityBodyState

	var stateA *bodyState
	if jnt.indexA == nullIndex {
		stateA = &dummyState
	} else {
		stateA = &context.states[jnt.indexA]
	}
	var stateB *bodyState
	if jnt.indexB == nullIndex {
		stateB = &dummyState
	} else {
		stateB = &context.states[jnt.indexB]
	}

	var rA Vec2 = RotateVector(stateA.deltaRotation, jnt.frameA.P)
	var rB Vec2 = RotateVector(stateB.deltaRotation, jnt.frameB.P)

	var linearImpulse Vec2 = Add(jnt.linearVelocityImpulse, jnt.linearSpringImpulse)
	var angularImpulse float32 = jnt.angularVelocityImpulse + jnt.angularSpringImpulse

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, linearImpulse)
		stateA.angularVelocity -= float32(iA * (Cross(rA, linearImpulse) + angularImpulse))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, linearImpulse)
		stateB.angularVelocity += float32(iB * (Cross(rB, linearImpulse) + angularImpulse))
	}
}

func solveMotorJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == MotorJointType, "base->type == b2_motorJoint")

	var mA float32 = base.invMassA
	var mB float32 = base.invMassB
	var iA float32 = base.invIA
	var iB float32 = base.invIB

	// dummy state for static bodies
	var dummyState bodyState = identityBodyState

	var jnt *motorJoint = &base.motorJoint
	var stateA *bodyState
	if jnt.indexA == nullIndex {
		stateA = &dummyState
	} else {
		stateA = &context.states[jnt.indexA]
	}
	var stateB *bodyState
	if jnt.indexB == nullIndex {
		stateB = &dummyState
	} else {
		stateB = &context.states[jnt.indexB]
	}

	var vA Vec2 = stateA.linearVelocity
	var wA float32 = stateA.angularVelocity
	var vB Vec2 = stateB.linearVelocity
	var wB float32 = stateB.angularVelocity

	// angular spring
	if jnt.maxSpringTorque > 0.0 && jnt.angularHertz > 0.0 {
		var qA Rot = MulRot(stateA.deltaRotation, jnt.frameA.Q)
		var qB Rot = MulRot(stateB.deltaRotation, jnt.frameB.Q)
		var relQ Rot = InvMulRot(qA, qB)

		var c float32 = relQ.Angle()
		var bias float32 = float32(jnt.angularSpring.biasRate * c)
		var massScale float32 = jnt.angularSpring.massScale
		var impulseScale float32 = jnt.angularSpring.impulseScale

		var cdot float32 = wB - wA

		var maxImpulse float32 = float32(context.h * jnt.maxSpringTorque)
		var oldImpulse float32 = jnt.angularSpringImpulse
		var impulse float32 = float32(float32(-massScale*jnt.angularMass)*(cdot+bias)) - float32(impulseScale*oldImpulse)
		jnt.angularSpringImpulse = ClampFloat(oldImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = jnt.angularSpringImpulse - oldImpulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	// angular velocity
	if jnt.maxVelocityTorque > 0.0 {
		var cdot float32 = wB - wA - jnt.angularVelocity
		var impulse float32 = float32(-jnt.angularMass * cdot)

		var maxImpulse float32 = float32(context.h * jnt.maxVelocityTorque)
		var oldImpulse float32 = jnt.angularVelocityImpulse
		jnt.angularVelocityImpulse = ClampFloat(oldImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = jnt.angularVelocityImpulse - oldImpulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	var rA Vec2 = RotateVector(stateA.deltaRotation, jnt.frameA.P)
	var rB Vec2 = RotateVector(stateB.deltaRotation, jnt.frameB.P)

	// linear spring
	if jnt.maxSpringForce > 0.0 && jnt.linearHertz > 0.0 {
		var dcA Vec2 = stateA.deltaPosition
		var dcB Vec2 = stateB.deltaPosition
		var c Vec2 = Add(Add(Sub(dcB, dcA), Sub(rB, rA)), jnt.deltaCenter)

		var bias Vec2 = MulSV(jnt.linearSpring.biasRate, c)
		var massScale float32 = jnt.linearSpring.massScale
		var impulseScale float32 = jnt.linearSpring.impulseScale

		var cdot Vec2 = Sub(Add(vB, CrossSV(wB, rB)), Add(vA, CrossSV(wA, rA)))
		cdot = Add(cdot, bias)

		// Updating the effective mass here may be overkill
		var kl Mat22
		kl.Cx.X = mA + mB + float32(float32(rA.Y*rA.Y)*iA) + float32(float32(rB.Y*rB.Y)*iB)
		kl.Cx.Y = float32(float32(-rA.Y*rA.X)*iA) - float32(float32(rB.Y*rB.X)*iB)
		kl.Cy.X = kl.Cx.Y
		kl.Cy.Y = mA + mB + float32(float32(rA.X*rA.X)*iA) + float32(float32(rB.X*rB.X)*iB)
		jnt.linearMass = GetInverse22(kl)

		var b Vec2 = MulMV(jnt.linearMass, cdot)

		var oldImpulse Vec2 = jnt.linearSpringImpulse
		var impulse Vec2 = Vec2{
			float32(-massScale*b.X) - float32(impulseScale*oldImpulse.X),
			float32(-massScale*b.Y) - float32(impulseScale*oldImpulse.Y),
		}

		var maxImpulse float32 = float32(context.h * jnt.maxSpringForce)
		jnt.linearSpringImpulse = Add(jnt.linearSpringImpulse, impulse)

		if LengthSquared(jnt.linearSpringImpulse) > float32(maxImpulse*maxImpulse) {
			jnt.linearSpringImpulse = Normalize(jnt.linearSpringImpulse)
			jnt.linearSpringImpulse.X *= maxImpulse
			jnt.linearSpringImpulse.Y *= maxImpulse
		}

		impulse = Sub(jnt.linearSpringImpulse, oldImpulse)

		vA = MulSub(vA, mA, impulse)
		wA -= float32(iA * Cross(rA, impulse))
		vB = MulAdd(vB, mB, impulse)
		wB += float32(iB * Cross(rB, impulse))
	}

	// linear velocity
	if jnt.maxVelocityForce > 0.0 {
		var cdot Vec2 = Sub(Add(vB, CrossSV(wB, rB)), Add(vA, CrossSV(wA, rA)))
		cdot = Sub(cdot, jnt.linearVelocity)
		var b Vec2 = MulMV(jnt.linearMass, cdot)
		var impulse Vec2 = Vec2{-b.X, -b.Y}

		var oldImpulse Vec2 = jnt.linearVelocityImpulse
		var maxImpulse float32 = float32(context.h * jnt.maxVelocityForce)
		jnt.linearVelocityImpulse = Add(jnt.linearVelocityImpulse, impulse)

		if LengthSquared(jnt.linearVelocityImpulse) > float32(maxImpulse*maxImpulse) {
			jnt.linearVelocityImpulse = Normalize(jnt.linearVelocityImpulse)
			jnt.linearVelocityImpulse.X *= maxImpulse
			jnt.linearVelocityImpulse.Y *= maxImpulse
		}

		impulse = Sub(jnt.linearVelocityImpulse, oldImpulse)

		vA = MulSub(vA, mA, impulse)
		wA -= float32(iA * Cross(rA, impulse))
		vB = MulAdd(vB, mB, impulse)
		wB += float32(iB * Cross(rB, impulse))
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
