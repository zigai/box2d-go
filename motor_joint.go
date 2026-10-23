// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "github.com/zigai/box2d/internal/core"

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

// Set the desired relative linear velocity in meters per second.
func (m MotorJoint) SetLinearVelocity(velocity Vec2) {
	checkInput(velocity.IsValid(), "b2IsValidVec2( velocity )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetLinearVelocity(world.recording, &recArgsMotorJointSetLinearVelocity{joint: m.Joint, velocity: velocity})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.linearVelocity = velocity
}

// Get the desired relative linear velocity in meters per second.
func (m MotorJoint) LinearVelocity() Vec2 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.linearVelocity
}

// Set the desired relative angular velocity in radians per second.
func (m MotorJoint) SetAngularVelocity(velocity float32) {
	checkInput(IsValidFloat(velocity), "b2IsValidFloat( velocity )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetAngularVelocity(world.recording, &recArgsMotorJointSetAngularVelocity{joint: m.Joint, velocity: velocity})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.angularVelocity = velocity
}

// Get the desired relative angular velocity in radians per second.
func (m MotorJoint) AngularVelocity() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.angularVelocity
}

// Set the motor joint maximum torque, usually in newton-meters.
func (m MotorJoint) SetMaxVelocityTorque(maxTorque float32) {
	checkInput(IsValidFloat(maxTorque), "b2IsValidFloat( maxTorque )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetMaxVelocityTorque(world.recording, &recArgsMotorJointSetMaxVelocityTorque{joint: m.Joint, maxTorque: maxTorque})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.maxVelocityTorque = maxTorque
}

// Get the motor joint maximum torque, usually in newton-meters.
func (m MotorJoint) MaxVelocityTorque() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxVelocityTorque
}

// Set the motor joint maximum force, usually in newtons.
func (m MotorJoint) SetMaxVelocityForce(maxForce float32) {
	checkInput(IsValidFloat(maxForce), "b2IsValidFloat( maxForce )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetMaxVelocityForce(world.recording, &recArgsMotorJointSetMaxVelocityForce{joint: m.Joint, maxForce: maxForce})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.maxVelocityForce = maxForce
}

// Get the motor joint maximum force, usually in newtons.
func (m MotorJoint) MaxVelocityForce() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxVelocityForce
}

// Set the spring linear hertz stiffness.
func (m MotorJoint) SetLinearHertz(hertz float32) {
	checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetLinearHertz(world.recording, &recArgsMotorJointSetLinearHertz{joint: m.Joint, hertz: hertz})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.linearHertz = hertz
}

// Get the spring linear hertz stiffness.
func (m MotorJoint) LinearHertz() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.linearHertz
}

// Set the spring linear damping ratio. Use 1.0 for critical damping.
func (m MotorJoint) SetLinearDampingRatio(damping float32) {
	checkInput(IsValidFloat(damping), "b2IsValidFloat( damping )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetLinearDampingRatio(world.recording, &recArgsMotorJointSetLinearDampingRatio{joint: m.Joint, damping: damping})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.linearDampingRatio = damping
}

// Get the spring linear damping ratio.
func (m MotorJoint) LinearDampingRatio() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.linearDampingRatio
}

// Set the spring angular hertz stiffness.
func (m MotorJoint) SetAngularHertz(hertz float32) {
	checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetAngularHertz(world.recording, &recArgsMotorJointSetAngularHertz{joint: m.Joint, hertz: hertz})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.angularHertz = hertz
}

// Get the spring angular hertz stiffness.
func (m MotorJoint) AngularHertz() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.angularHertz
}

// Set the spring angular damping ratio. Use 1.0 for critical damping.
func (m MotorJoint) SetAngularDampingRatio(damping float32) {
	checkInput(IsValidFloat(damping), "b2IsValidFloat( damping )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetAngularDampingRatio(world.recording, &recArgsMotorJointSetAngularDampingRatio{joint: m.Joint, damping: damping})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.angularDampingRatio = damping
}

// Get the spring angular damping ratio.
func (m MotorJoint) AngularDampingRatio() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.angularDampingRatio
}

// Set the maximum spring force in newtons.
func (m MotorJoint) SetMaxSpringForce(maxForce float32) {
	checkInput(IsValidFloat(maxForce), "b2IsValidFloat( maxForce )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetMaxSpringForce(world.recording, &recArgsMotorJointSetMaxSpringForce{joint: m.Joint, maxForce: maxForce})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.maxSpringForce = maxFloat(0.0, maxForce)
}

// Get the maximum spring force in newtons.
func (m MotorJoint) MaxSpringForce() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxSpringForce
}

// Set the maximum spring torque in newtons * meters.
func (m MotorJoint) SetMaxSpringTorque(maxTorque float32) {
	checkInput(IsValidFloat(maxTorque), "b2IsValidFloat( maxTorque )")

	world := m.Joint.world
	if world.recording != nil {
		recWriteMotorJointSetMaxSpringTorque(world.recording, &recArgsMotorJointSetMaxSpringTorque{joint: m.Joint, maxTorque: maxTorque})
	}

	jnt := getJointSimCheckType(m.Joint, MotorJointType)

	jnt.motorJoint.maxSpringTorque = maxFloat(0.0, maxTorque)
}

// Get the maximum spring torque in newtons * meters.
func (m MotorJoint) MaxSpringTorque() float32 {
	jnt := getJointSimCheckType(m.Joint, MotorJointType)
	return jnt.motorJoint.maxSpringTorque
}

func getMotorJointForce(world *World, base *jointSim) Vec2 {
	force := base.motorJoint.linearVelocityImpulse.Add(base.motorJoint.linearSpringImpulse).Scale(world.invH)
	return force
}

func getMotorJointTorque(world *World, base *jointSim) float32 {
	return float32(world.invH * (base.motorJoint.angularVelocityImpulse + base.motorJoint.angularSpringImpulse))
}

func prepareMotorJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == MotorJointType, "base->type == b2_motorJoint")

	// chase body id to the solver set where the body lives
	var (
		idA = base.bodyIDA
		idB = base.bodyIDB
	)

	world := context.world

	var (
		bodyA = &world.bodies[idA]
		bodyB = &world.bodies[idB]
	)

	core.Assert(bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet, "bodyA->setIndex == b2_awakeSet || bodyB->setIndex == b2_awakeSet")

	var (
		setA = &world.solverSets[bodyA.setIndex]
		setB = &world.solverSets[bodyB.setIndex]
	)

	var (
		localIndexA = bodyA.localIndex
		localIndexB = bodyB.localIndex
	)

	var (
		bodySimA = &setA.bodySims[localIndexA]
		bodySimB = &setB.bodySims[localIndexB]
	)

	var (
		mA = bodySimA.invMass
		iA = bodySimA.invInertia
		mB = bodySimB.invMass
		iB = bodySimB.invInertia
	)

	base.invMassA = mA
	base.invMassB = mB
	base.invIA = iA
	base.invIB = iB

	jnt := &base.motorJoint
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
	jnt.frameA.Q = bodySimA.transform.Q.Mul(base.localFrameA.Q)
	jnt.frameA.P = bodySimA.transform.Q.Rotate(base.localFrameA.P.Sub(bodySimA.localCenter))
	jnt.frameB.Q = bodySimB.transform.Q.Mul(base.localFrameB.Q)
	jnt.frameB.P = bodySimB.transform.Q.Rotate(base.localFrameB.P.Sub(bodySimB.localCenter))

	// Compute the initial center delta. Incremental position updates are relative to this.
	jnt.deltaCenter = bodySimB.center.Sub(bodySimA.center)

	var (
		rA = jnt.frameA.P
		rB = jnt.frameB.P
	)

	jnt.linearSpring = makeSoft(jnt.linearHertz, jnt.linearDampingRatio, context.h)
	jnt.angularSpring = makeSoft(jnt.angularHertz, jnt.angularDampingRatio, context.h)

	var kl Mat22

	kl.Cx.X = mA + mB + float32(float32(rA.Y*rA.Y)*iA) + float32(float32(rB.Y*rB.Y)*iB)
	kl.Cx.Y = float32(float32(-rA.Y*rA.X)*iA) - float32(float32(rB.Y*rB.X)*iB)
	kl.Cy.X = kl.Cx.Y
	kl.Cy.Y = mA + mB + float32(float32(rA.X*rA.X)*iA) + float32(float32(rB.X*rB.X)*iB)
	jnt.linearMass = kl.Inverse()

	ka := iA + iB
	if ka > 0.0 {
		jnt.angularMass = 1.0 / ka
	} else {
		jnt.angularMass = 0.0
	}

	if !context.enableWarmStarting {
		jnt.linearVelocityImpulse = Vec2Zero
		jnt.angularVelocityImpulse = 0.0
		jnt.linearSpringImpulse = Vec2Zero
		jnt.angularSpringImpulse = 0.0
	}
}

func warmStartMotorJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == MotorJointType, "base->type == b2_motorJoint")

	var (
		mA = base.invMassA
		mB = base.invMassB
		iA = base.invIA
		iB = base.invIB
	)

	jnt := &base.motorJoint

	// dummy state for static bodies
	dummyState := identityBodyState

	stateA := jointSolverBodyState(context, jnt.indexA, &dummyState)

	stateB := jointSolverBodyState(context, jnt.indexB, &dummyState)

	var (
		rA = stateA.deltaRotation.Rotate(jnt.frameA.P)
		rB = stateB.deltaRotation.Rotate(jnt.frameB.P)
	)

	var (
		linearImpulse  = jnt.linearVelocityImpulse.Add(jnt.linearSpringImpulse)
		angularImpulse = jnt.angularVelocityImpulse + jnt.angularSpringImpulse
	)

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = stateA.linearVelocity.MulSub(mA, linearImpulse)
		stateA.angularVelocity -= float32(iA * (rA.Cross(linearImpulse) + angularImpulse))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = stateB.linearVelocity.MulAdd(mB, linearImpulse)
		stateB.angularVelocity += float32(iB * (rB.Cross(linearImpulse) + angularImpulse))
	}
}

func solveMotorJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == MotorJointType, "base->type == b2_motorJoint")

	var (
		mA = base.invMassA
		mB = base.invMassB
		iA = base.invIA
		iB = base.invIB
	)

	// dummy state for static bodies
	dummyState := identityBodyState

	var (
		jnt    = &base.motorJoint
		stateA = jointSolverBodyState(context, jnt.indexA, &dummyState)
	)

	stateB := jointSolverBodyState(context, jnt.indexB, &dummyState)

	var (
		vA = stateA.linearVelocity
		wA = stateA.angularVelocity
		vB = stateB.linearVelocity
		wB = stateB.angularVelocity
	)

	// angular spring
	if jnt.maxSpringTorque > 0.0 && jnt.angularHertz > 0.0 {
		var (
			qA   = stateA.deltaRotation.Mul(jnt.frameA.Q)
			qB   = stateB.deltaRotation.Mul(jnt.frameB.Q)
			relQ = qA.InvMul(qB)
		)

		var (
			c            = relQ.Angle()
			bias         = float32(jnt.angularSpring.biasRate * c)
			massScale    = jnt.angularSpring.massScale
			impulseScale = jnt.angularSpring.impulseScale
		)

		cdot := wB - wA

		var (
			maxImpulse = float32(context.h * jnt.maxSpringTorque)
			oldImpulse = jnt.angularSpringImpulse
			impulse    = float32(float32(-massScale*jnt.angularMass)*(cdot+bias)) - float32(impulseScale*oldImpulse)
		)

		jnt.angularSpringImpulse = clampFloat(oldImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = jnt.angularSpringImpulse - oldImpulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	// angular velocity
	if jnt.maxVelocityTorque > 0.0 {
		var (
			cdot    = wB - wA - jnt.angularVelocity
			impulse = float32(-jnt.angularMass * cdot)
		)

		var (
			maxImpulse = float32(context.h * jnt.maxVelocityTorque)
			oldImpulse = jnt.angularVelocityImpulse
		)

		jnt.angularVelocityImpulse = clampFloat(oldImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = jnt.angularVelocityImpulse - oldImpulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	var (
		rA = stateA.deltaRotation.Rotate(jnt.frameA.P)
		rB = stateB.deltaRotation.Rotate(jnt.frameB.P)
	)

	// linear spring
	if jnt.maxSpringForce > 0.0 && jnt.linearHertz > 0.0 {
		var (
			dcA = stateA.deltaPosition
			dcB = stateB.deltaPosition
			c   = dcB.Sub(dcA).Add(rB.Sub(rA)).Add(jnt.deltaCenter)
		)

		var (
			bias         = c.Scale(jnt.linearSpring.biasRate)
			massScale    = jnt.linearSpring.massScale
			impulseScale = jnt.linearSpring.impulseScale
		)

		cdot := vB.Add(rB.ScalarCross(wB)).Sub(vA.Add(rA.ScalarCross(wA)))

		cdot = cdot.Add(bias)

		// Updating the effective mass here may be overkill
		var kl Mat22

		kl.Cx.X = mA + mB + float32(float32(rA.Y*rA.Y)*iA) + float32(float32(rB.Y*rB.Y)*iB)
		kl.Cx.Y = float32(float32(-rA.Y*rA.X)*iA) - float32(float32(rB.Y*rB.X)*iB)
		kl.Cy.X = kl.Cx.Y
		kl.Cy.Y = mA + mB + float32(float32(rA.X*rA.X)*iA) + float32(float32(rB.X*rB.X)*iB)
		jnt.linearMass = kl.Inverse()

		b := jnt.linearMass.MulVec(cdot)

		var (
			oldImpulse = jnt.linearSpringImpulse
			impulse    = Vec2{
				float32(-massScale*b.X) - float32(impulseScale*oldImpulse.X),
				float32(-massScale*b.Y) - float32(impulseScale*oldImpulse.Y),
			}
		)

		maxImpulse := float32(context.h * jnt.maxSpringForce)

		jnt.linearSpringImpulse = jnt.linearSpringImpulse.Add(impulse)

		if jnt.linearSpringImpulse.LengthSquared() > float32(maxImpulse*maxImpulse) {
			jnt.linearSpringImpulse = jnt.linearSpringImpulse.Normalize()
			jnt.linearSpringImpulse.X *= maxImpulse
			jnt.linearSpringImpulse.Y *= maxImpulse
		}

		impulse = jnt.linearSpringImpulse.Sub(oldImpulse)

		vA = vA.MulSub(mA, impulse)
		wA -= float32(iA * rA.Cross(impulse))
		vB = vB.MulAdd(mB, impulse)
		wB += float32(iB * rB.Cross(impulse))
	}

	// linear velocity
	if jnt.maxVelocityForce > 0.0 {
		cdot := vB.Add(rB.ScalarCross(wB)).Sub(vA.Add(rA.ScalarCross(wA)))

		cdot = cdot.Sub(jnt.linearVelocity)

		var (
			b       = jnt.linearMass.MulVec(cdot)
			impulse = Vec2{-b.X, -b.Y}
		)

		var (
			oldImpulse = jnt.linearVelocityImpulse
			maxImpulse = float32(context.h * jnt.maxVelocityForce)
		)

		jnt.linearVelocityImpulse = jnt.linearVelocityImpulse.Add(impulse)

		if jnt.linearVelocityImpulse.LengthSquared() > float32(maxImpulse*maxImpulse) {
			jnt.linearVelocityImpulse = jnt.linearVelocityImpulse.Normalize()
			jnt.linearVelocityImpulse.X *= maxImpulse
			jnt.linearVelocityImpulse.Y *= maxImpulse
		}

		impulse = jnt.linearVelocityImpulse.Sub(oldImpulse)

		vA = vA.MulSub(mA, impulse)
		wA -= float32(iA * rA.Cross(impulse))
		vB = vB.MulAdd(mB, impulse)
		wB += float32(iB * rB.Cross(impulse))
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
