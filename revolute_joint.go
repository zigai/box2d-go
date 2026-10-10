// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "fmt"

type RevoluteJoint struct {
	Joint
}

// Revolute joint definition
// A point on body B is fixed to a point on body A. Allows relative rotation.
type RevoluteJointDef struct {
	// Base joint definition
	Base JointDef
	// The target angle for the joint in radians. The spring-damper will drive
	// to this angle.
	TargetAngle float32
	// Enable a rotational spring on the revolute hinge axis
	EnableSpring bool
	// The spring stiffness Hertz, cycles per second
	Hertz float32
	// The spring damping ratio, non-dimensional
	DampingRatio float32
	// A flag to enable joint limits
	EnableLimit bool
	// The lower angle for the joint limit in radians. Minimum of -0.99*pi radians.
	LowerAngle float32
	// The upper angle for the joint limit in radians. Maximum of 0.99*pi radians.
	UpperAngle float32
	// A flag to enable the joint motor
	EnableMotor bool
	// The maximum motor torque, typically in newton-meters
	MaxMotorTorque float32
	// The desired motor speed in radians per second
	MotorSpeed float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Enable/disable the revolute joint spring.
func (r RevoluteJoint) SetSpringEnabled(enableSpring bool) {
	// Point-to-point constraint
	// C = pB - pA
	// Cdot = vB - vA
	//      = vB + cross(wB, rB) - vA - cross(wA, rA)
	// J = [-E -skew(rA) E skew(rB) ]
	// Identity used:
	// w k % (rx i + ry j) = w * (-ry i + rx j)
	// Motor constraint
	// Cdot = wB - wA
	// J = [0 0 -1 0 0 1]
	// K = invIA + invIB
	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointEnableSpring(world.recording, &recArgsRevoluteJointEnableSpring{joint: r.Joint, enableSpring: enableSpring})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	if enableSpring != joint.revoluteJoint.enableSpring {
		joint.revoluteJoint.enableSpring = enableSpring
		joint.revoluteJoint.springImpulse = float32(0.0)
	}
}

// It the revolute angular spring enabled?
func (r RevoluteJoint) SpringEnabled() bool {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.enableSpring
}

// Set the revolute joint spring stiffness in Hertz.
func (r RevoluteJoint) SetSpringHertz(hertz float32) {
	checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")

	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointSetSpringHertz(world.recording, &recArgsRevoluteJointSetSpringHertz{joint: r.Joint, hertz: hertz})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	joint.revoluteJoint.hertz = hertz
}

// Get the revolute joint spring stiffness in Hertz.
func (r RevoluteJoint) SpringHertz() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.hertz
}

// Set the revolute joint spring damping ratio, non-dimensional.
func (r RevoluteJoint) SetSpringDampingRatio(dampingRatio float32) {
	checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )")

	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointSetSpringDampingRatio(world.recording, &recArgsRevoluteJointSetSpringDampingRatio{joint: r.Joint, dampingRatio: dampingRatio})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	joint.revoluteJoint.dampingRatio = dampingRatio
}

// Get the revolute joint spring damping ratio, non-dimensional.
func (r RevoluteJoint) SpringDampingRatio() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.dampingRatio
}

// Set the revolute joint spring target angle, radians.
func (r RevoluteJoint) SetTargetAngle(angle float32) {
	checkInput(IsValidFloat(angle), "b2IsValidFloat( angle )")

	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointSetTargetAngle(world.recording, &recArgsRevoluteJointSetTargetAngle{joint: r.Joint, angle: angle})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	joint.revoluteJoint.targetAngle = angle
}

// Get the revolute joint spring target angle, radians.
func (r RevoluteJoint) TargetAngle() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.targetAngle
}

// Get the revolute joint current angle in radians relative to the reference angle.
func (r RevoluteJoint) Angle() float32 {
	world := r.Joint.world
	jointSim := getJointSimCheckType(r.Joint, RevoluteJointType)
	qA := getBodyTransform(world, jointSim.bodyIDA).Q.Mul(jointSim.localFrameA.Q)
	qB := getBodyTransform(world, jointSim.bodyIDB).Q.Mul(jointSim.localFrameB.Q)

	angle := qA.RelativeAngle(qB)

	return angle
}

// Enable/disable the revolute joint limit.
func (r RevoluteJoint) SetLimitEnabled(enableLimit bool) {
	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointEnableLimit(world.recording, &recArgsRevoluteJointEnableLimit{joint: r.Joint, enableLimit: enableLimit})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	if enableLimit != joint.revoluteJoint.enableLimit {
		joint.revoluteJoint.enableLimit = enableLimit
		joint.revoluteJoint.lowerImpulse = float32(0.0)
		joint.revoluteJoint.upperImpulse = float32(0.0)
	}
}

// Is the revolute joint limit enabled?
func (r RevoluteJoint) LimitEnabled() bool {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.enableLimit
}

// Get the revolute joint lower limit in radians.
func (r RevoluteJoint) LowerLimit() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.lowerAngle
}

// Get the revolute joint upper limit in radians.
func (r RevoluteJoint) UpperLimit() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.upperAngle
}

// Set the revolute joint limits in radians. It is expected that lower <= upper
// and that -0.99 * B2_PI <= lower && upper <= -0.99 * B2_PI.
func (r RevoluteJoint) SetLimits(lower float32, upper float32) {
	checkInput(IsValidFloat(lower), "b2IsValidFloat( lower )")
	checkInput(IsValidFloat(upper), "b2IsValidFloat( upper )")

	assert(lower <= upper, "lower <= upper")

	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointSetLimits(world.recording, &recArgsRevoluteJointSetLimits{joint: r.Joint, lower: lower, upper: upper})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	lowerAngle := minFloat(lower, upper)
	upperAngle := maxFloat(lower, upper)
	joint.revoluteJoint.lowerAngle = clampFloat(lowerAngle, float32(-float32(0.99)*Pi), float32(float32(0.99)*Pi))
	joint.revoluteJoint.upperAngle = clampFloat(upperAngle, float32(-float32(0.99)*Pi), float32(float32(0.99)*Pi))
}

// Enable/disable a revolute joint motor.
func (r RevoluteJoint) SetMotorEnabled(enableMotor bool) {
	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointEnableMotor(world.recording, &recArgsRevoluteJointEnableMotor{joint: r.Joint, enableMotor: enableMotor})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	if enableMotor != joint.revoluteJoint.enableMotor {
		joint.revoluteJoint.enableMotor = enableMotor
		joint.revoluteJoint.motorImpulse = float32(0.0)
	}
}

// Is the revolute joint motor enabled?
func (r RevoluteJoint) MotorEnabled() bool {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.enableMotor
}

// Set the revolute joint motor speed in radians per second.
func (r RevoluteJoint) SetMotorSpeed(motorSpeed float32) {
	checkInput(IsValidFloat(motorSpeed), "b2IsValidFloat( motorSpeed )")

	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointSetMotorSpeed(world.recording, &recArgsRevoluteJointSetMotorSpeed{joint: r.Joint, motorSpeed: motorSpeed})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	joint.revoluteJoint.motorSpeed = motorSpeed
}

// Get the revolute joint motor speed in radians per second.
func (r RevoluteJoint) MotorSpeed() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.motorSpeed
}

// Get the revolute joint current motor torque, usually in newton-meters.
func (r RevoluteJoint) MotorTorque() float32 {
	world := r.Joint.world
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)

	return float32(world.invH * joint.revoluteJoint.motorImpulse)
}

// Set the revolute joint maximum motor torque, usually in newton-meters.
func (r RevoluteJoint) SetMaxMotorTorque(torque float32) {
	checkInput(IsValidFloat(torque), "b2IsValidFloat( torque )")

	world := r.Joint.world
	if world.recording != nil {
		recWriteRevoluteJointSetMaxMotorTorque(world.recording, &recArgsRevoluteJointSetMaxMotorTorque{joint: r.Joint, torque: torque})
	}

	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	joint.revoluteJoint.maxMotorTorque = torque
}

// Get the revolute joint maximum motor torque, usually in newton-meters.
func (r RevoluteJoint) MaxMotorTorque() float32 {
	joint := getJointSimCheckType(r.Joint, RevoluteJointType)
	return joint.revoluteJoint.maxMotorTorque
}

func getRevoluteJointForce(world *World, base *jointSim) Vec2 {
	force := Vec2{X: float32(world.invH * base.revoluteJoint.linearImpulse.X), Y: float32(world.invH * base.revoluteJoint.linearImpulse.Y)}
	return force
}

func getRevoluteJointTorque(world *World, base *jointSim) float32 {
	revolute := &base.revoluteJoint
	torque := float32(world.invH * (revolute.motorImpulse + revolute.lowerImpulse - revolute.upperImpulse))

	return torque
}

func prepareRevoluteJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == RevoluteJointType, "base->type == b2_revoluteJoint")

	// chase body id to the solver set where the body lives
	idA := base.bodyIDA
	idB := base.bodyIDB

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

	joint := &base.revoluteJoint

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

	// Compute joint anchor frames with world space rotation, relative to center of mass.
	// Avoid round-off here as much as possible.
	// Vec2 pf = (xf.p - c) + rot(xf.q, f.p)
	// pf = xf.p - (xf.p + rot(xf.q, lc)) + rot(xf.q, f.p)
	// pf = rot(xf.q, f.p - lc)
	joint.frameA.Q = bodySimA.transform.Q.Mul(base.localFrameA.Q)
	joint.frameA.P = bodySimA.transform.Q.Rotate(base.localFrameA.P.Sub(bodySimA.localCenter))
	joint.frameB.Q = bodySimB.transform.Q.Mul(base.localFrameB.Q)
	joint.frameB.P = bodySimB.transform.Q.Rotate(base.localFrameB.P.Sub(bodySimB.localCenter))

	// Compute the initial center delta. Incremental position updates are relative to this.
	joint.deltaCenter = bodySimB.center.Sub(bodySimA.center)

	k := iA + iB
	if k > float32(0.0) {
		joint.axialMass = float32(1.0) / k
	} else {
		joint.axialMass = float32(0.0)
	}

	joint.springSoftness = makeSoft(joint.hertz, joint.dampingRatio, context.h)

	if !context.enableWarmStarting {
		joint.linearImpulse = Vec2{}
		joint.springImpulse = float32(0.0)
		joint.motorImpulse = float32(0.0)
		joint.lowerImpulse = float32(0.0)
		joint.upperImpulse = float32(0.0)
	}
}

func warmStartRevoluteJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == RevoluteJointType, "base->type == b2_revoluteJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.revoluteJoint

	stateA := jointSolverBodyState(context, joint.indexA, &dummyState)

	stateB := jointSolverBodyState(context, joint.indexB, &dummyState)

	rA := stateA.deltaRotation.Rotate(joint.frameA.P)
	rB := stateB.deltaRotation.Rotate(joint.frameB.P)

	axialImpulse := joint.springImpulse + joint.motorImpulse + joint.lowerImpulse - joint.upperImpulse

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = stateA.linearVelocity.MulSub(mA, joint.linearImpulse)
		stateA.angularVelocity -= float32(iA * (rA.Cross(joint.linearImpulse) + axialImpulse))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = stateB.linearVelocity.MulAdd(mB, joint.linearImpulse)
		stateB.angularVelocity += float32(iB * (rB.Cross(joint.linearImpulse) + axialImpulse))
	}
}

func solveRevoluteJoint(base *jointSim, context *stepContext, useBias bool) {
	assert(base.jointType == RevoluteJointType, "base->type == b2_revoluteJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.revoluteJoint

	stateA := jointSolverBodyState(context, joint.indexA, &dummyState)

	stateB := jointSolverBodyState(context, joint.indexB, &dummyState)

	vA := stateA.linearVelocity
	wA := stateA.angularVelocity
	vB := stateB.linearVelocity
	wB := stateB.angularVelocity

	fixedRotation := (iA+iB == float32(0.0))

	var relQ Rot

	if joint.enableSpring || joint.enableLimit {
		qA := stateA.deltaRotation.Mul(joint.frameA.Q)
		qB := stateB.deltaRotation.Mul(joint.frameB.Q)
		relQ = qA.InvMul(qB)
	}

	// Solve spring.
	if joint.enableSpring && !fixedRotation {
		jointAngle := relQ.Angle()
		jointAngleDelta := UnwindAngle(jointAngle - joint.targetAngle)

		separation := jointAngleDelta
		bias := float32(joint.springSoftness.biasRate * separation)
		massScale := joint.springSoftness.massScale
		impulseScale := joint.springSoftness.impulseScale

		velocityError := wB - wA
		impulse := float32(float32(-massScale*joint.axialMass)*(velocityError+bias)) - float32(impulseScale*joint.springImpulse)
		joint.springImpulse += impulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	// Solve motor constraint.
	if joint.enableMotor && !fixedRotation {
		velocityError := wB - wA - joint.motorSpeed
		impulse := float32(-joint.axialMass * velocityError)
		oldImpulse := joint.motorImpulse
		maxImpulse := float32(context.h * joint.maxMotorTorque)
		joint.motorImpulse = clampFloat(joint.motorImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = joint.motorImpulse - oldImpulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	if joint.enableLimit && !fixedRotation {
		jointAngle := relQ.Angle()

		// Lower limit
		{
			separation := jointAngle - joint.lowerAngle
			bias, massScale, impulseScale := jointLimitBias(separation, context.invH, useBias, base.constraintSoftness)

			velocityError := wB - wA
			oldImpulse := joint.lowerImpulse
			impulse := float32(float32(-massScale*joint.axialMass)*(velocityError+bias)) - float32(impulseScale*oldImpulse)
			joint.lowerImpulse = maxFloat(oldImpulse+impulse, float32(0.0))
			impulse = joint.lowerImpulse - oldImpulse

			wA -= float32(iA * impulse)
			wB += float32(iB * impulse)
		}

		// Upper limit
		// Note: signs are flipped to keep C positive when the constraint is satisfied.
		// This also keeps the impulse positive when the limit is active.
		{
			separation := joint.upperAngle - jointAngle
			bias, massScale, impulseScale := jointLimitBias(separation, context.invH, useBias, base.constraintSoftness)

			// sign flipped on velocityError
			velocityError := wA - wB
			oldImpulse := joint.upperImpulse
			impulse := float32(float32(-massScale*joint.axialMass)*(velocityError+bias)) - float32(impulseScale*oldImpulse)
			joint.upperImpulse = maxFloat(oldImpulse+impulse, float32(0.0))
			impulse = joint.upperImpulse - oldImpulse

			// sign flipped on applied impulse
			wA += float32(iA * impulse)
			wB -= float32(iB * impulse)
		}
	}

	// Solve point-to-point constraint
	{
		// J = [-I -r1_skew I r2_skew]
		// r_skew = [-ry; rx]
		// K = [ mA+r1y^2*iA+mB+r2y^2*iB,  -r1y*iA*r1x-r2y*iB*r2x]
		//     [  -r1y*iA*r1x-r2y*iB*r2x, mA+r1x^2*iA+mB+r2x^2*iB]

		// current anchors
		rA := stateA.deltaRotation.Rotate(joint.frameA.P)
		rB := stateB.deltaRotation.Rotate(joint.frameB.P)

		velocityError := vB.Add(Vec2{X: float32(-wB * rB.Y), Y: float32(wB * rB.X)}).Sub(vA.Add(Vec2{X: float32(-wA * rA.Y), Y: float32(wA * rA.X)}))

		bias := Vec2{}
		massScale := float32(1.0)
		impulseScale := float32(0.0)

		if useBias {
			dcA := stateA.deltaPosition
			dcB := stateB.deltaPosition

			separation := dcB.Sub(dcA).Add(rB.Sub(rA)).Add(joint.deltaCenter)
			bias = Vec2{X: float32(base.constraintSoftness.biasRate * separation.X), Y: float32(base.constraintSoftness.biasRate * separation.Y)}
			massScale = base.constraintSoftness.massScale
			impulseScale = base.constraintSoftness.impulseScale
		}

		var massMatrix Mat22

		massMatrix.Cx.X = mA + mB + float32(float32(rA.Y*rA.Y)*iA) + float32(float32(rB.Y*rB.Y)*iB)
		massMatrix.Cy.X = float32(float32(-rA.Y*rA.X)*iA) - float32(float32(rB.Y*rB.X)*iB)
		massMatrix.Cx.Y = massMatrix.Cy.X
		massMatrix.Cy.Y = mA + mB + float32(float32(rA.X*rA.X)*iA) + float32(float32(rB.X*rB.X)*iB)
		b := massMatrix.Solve(velocityError.Add(bias))
		b.X = float32(b.X)
		b.Y = float32(b.Y)

		var impulse Vec2

		impulse.X = float32(-massScale*b.X) - float32(impulseScale*joint.linearImpulse.X)
		impulse.Y = float32(-massScale*b.Y) - float32(impulseScale*joint.linearImpulse.Y)
		joint.linearImpulse.X += impulse.X
		joint.linearImpulse.Y += impulse.Y

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

func drawRevoluteJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	assert(base.jointType == RevoluteJointType, "base->type == b2_revoluteJoint")

	joint := &base.revoluteJoint

	frameA := transformA.Mul(base.localFrameA)
	frameB := transformB.Mul(base.localFrameB)

	radius := float32(float32(0.25) * drawScale)
	draw.DrawCircle(frameB.P, radius, ColorGray)

	rx := Vec2{X: radius, Y: float32(0.0)}
	r := frameA.Q.Rotate(rx)
	draw.DrawLine(frameA.P, frameA.P.Add(r), ColorGray)

	r = frameB.Q.Rotate(rx)
	draw.DrawLine(frameB.P, frameB.P.Add(r), ColorBlue)

	if draw.DrawJointExtras {
		jointAngle := frameA.Q.RelativeAngle(frameB.Q)

		buffer := fmt.Sprintf(" %.1f deg", float32(float32(180.0)*jointAngle)/Pi)
		draw.DrawString(frameA.P.Add(r), buffer, ColorWhite)
	}

	lowerAngle := joint.lowerAngle
	upperAngle := joint.upperAngle

	if joint.enableLimit {
		rotLo := frameA.Q.Mul(MakeRot(lowerAngle))
		rlo := rotLo.Rotate(rx)

		rotHi := frameA.Q.Mul(MakeRot(upperAngle))
		rhi := rotHi.Rotate(rx)

		draw.DrawLine(frameB.P, frameB.P.Add(rlo), ColorGreen)
		draw.DrawLine(frameB.P, frameB.P.Add(rhi), ColorRed)
	}

	if joint.enableSpring {
		q := frameA.Q.Mul(MakeRot(joint.targetAngle))
		v := q.Rotate(rx)
		draw.DrawLine(frameB.P, frameB.P.Add(v), ColorViolet)
	}

	color := ColorGold
	draw.DrawLine(frameA.P, frameB.P, color)

	if draw.DrawJointExtras {
		draw.DrawLine(transformA.P, frameA.P, color)
		draw.DrawLine(transformB.P, frameB.P, color)
	}

	// char buffer[32];
	// sprintf(buffer, "%.1f", Length(joint->impulse));
	// draw->DrawString(pA, buffer, draw->context);
}
