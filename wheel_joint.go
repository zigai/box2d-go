// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type WheelJoint struct {
	Joint
}

// Wheel joint definition
// Body B is a wheel that may rotate freely and slide along the local x-axis in frame A.
// The joint translation is zero when the local frame origins coincide in world space.
type WheelJointDef struct {
	// Base joint definition
	Base JointDef
	// Enable a linear spring along the local axis
	EnableSpring bool
	// Spring stiffness in Hertz
	Hertz float32
	// Spring damping ratio, non-dimensional
	DampingRatio float32
	// Enable/disable the joint linear limit
	EnableLimit bool
	// The lower translation limit
	LowerTranslation float32
	// The upper translation limit
	UpperTranslation float32
	// Enable/disable the joint rotational motor
	EnableMotor bool
	// The maximum motor torque, typically in newton-meters
	MaxMotorTorque float32
	// The desired motor speed in radians per second
	MotorSpeed float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Enable/disable the wheel joint spring
func (w WheelJoint) EnableSpring(enableSpring bool) {
	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointEnableSpring(world.recording, &recArgsWheelJointEnableSpring{joint: w.Joint, enableSpring: enableSpring})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)

	if enableSpring != joint.wheelJoint.enableSpring {
		joint.wheelJoint.enableSpring = enableSpring
		joint.wheelJoint.springImpulse = float32(0.0)
	}
}

// Is the wheel joint spring enabled?
func (w WheelJoint) IsSpringEnabled() bool {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.enableSpring
}

// Set the wheel joint stiffness in Hertz
func (w WheelJoint) SetSpringHertz(hertz float32) {
	if !checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointSetSpringHertz(world.recording, &recArgsWheelJointSetSpringHertz{joint: w.Joint, hertz: hertz})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	joint.wheelJoint.hertz = hertz
}

// Get the wheel joint stiffness in Hertz
func (w WheelJoint) SpringHertz() float32 {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.hertz
}

// Set the wheel joint damping ratio, non-dimensional
func (w WheelJoint) SetSpringDampingRatio(dampingRatio float32) {
	if !checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointSetSpringDampingRatio(world.recording, &recArgsWheelJointSetSpringDampingRatio{joint: w.Joint, dampingRatio: dampingRatio})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	joint.wheelJoint.dampingRatio = dampingRatio
}

// Get the wheel joint damping ratio, non-dimensional
func (w WheelJoint) SpringDampingRatio() float32 {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.dampingRatio
}

// Enable/disable the wheel joint limit
func (w WheelJoint) EnableLimit(enableLimit bool) {
	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointEnableLimit(world.recording, &recArgsWheelJointEnableLimit{joint: w.Joint, enableLimit: enableLimit})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	if joint.wheelJoint.enableLimit != enableLimit {
		joint.wheelJoint.lowerImpulse = float32(0.0)
		joint.wheelJoint.upperImpulse = float32(0.0)
		joint.wheelJoint.enableLimit = enableLimit
	}
}

// Is the wheel joint limit enabled?
func (w WheelJoint) IsLimitEnabled() bool {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.enableLimit
}

// Get the wheel joint lower limit
func (w WheelJoint) LowerLimit() float32 {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.lowerTranslation
}

// Get the wheel joint upper limit
func (w WheelJoint) UpperLimit() float32 {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.upperTranslation
}

// Set the wheel joint limits
func (w WheelJoint) SetLimits(lower float32, upper float32) {
	if !checkInput(IsValidFloat(lower), "b2IsValidFloat( lower )") {
		return
	}
	if !checkInput(IsValidFloat(upper), "b2IsValidFloat( upper )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointSetLimits(world.recording, &recArgsWheelJointSetLimits{joint: w.Joint, lower: lower, upper: upper})
	}
	assert(lower <= upper, "lower <= upper")

	joint := getJointSimCheckType(w.Joint, WheelJointType)
	joint.wheelJoint.lowerTranslation = MinFloat(lower, upper)
	joint.wheelJoint.upperTranslation = MaxFloat(lower, upper)
}

// Enable/disable the wheel joint motor
func (w WheelJoint) EnableMotor(enableMotor bool) {
	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointEnableMotor(world.recording, &recArgsWheelJointEnableMotor{joint: w.Joint, enableMotor: enableMotor})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	if joint.wheelJoint.enableMotor != enableMotor {
		joint.wheelJoint.motorImpulse = float32(0.0)
		joint.wheelJoint.enableMotor = enableMotor
	}
}

// Is the wheel joint motor enabled?
func (w WheelJoint) IsMotorEnabled() bool {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.enableMotor
}

// Set the wheel joint motor speed in radians per second
func (w WheelJoint) SetMotorSpeed(motorSpeed float32) {
	if !checkInput(IsValidFloat(motorSpeed), "b2IsValidFloat( motorSpeed )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointSetMotorSpeed(world.recording, &recArgsWheelJointSetMotorSpeed{joint: w.Joint, motorSpeed: motorSpeed})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	joint.wheelJoint.motorSpeed = motorSpeed
}

// Get the wheel joint motor speed in radians per second
func (w WheelJoint) MotorSpeed() float32 {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.motorSpeed
}

// Get the wheel joint current motor torque, usually in newton-meters
func (w WheelJoint) MotorTorque() float32 {
	world := getWorld(int(w.Joint.world0))
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return float32(world.invH * joint.wheelJoint.motorImpulse)
}

// Set the wheel joint maximum motor torque, usually in newton-meters
func (w WheelJoint) SetMaxMotorTorque(torque float32) {
	if !checkInput(IsValidFloat(torque), "b2IsValidFloat( torque )") {
		return
	}

	world := getWorld(int(w.Joint.world0))
	if world.recording != nil {
		recWriteWheelJointSetMaxMotorTorque(world.recording, &recArgsWheelJointSetMaxMotorTorque{joint: w.Joint, torque: torque})
	}
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	joint.wheelJoint.maxMotorTorque = torque
}

// Get the wheel joint maximum motor torque, usually in newton-meters
func (w WheelJoint) MaxMotorTorque() float32 {
	joint := getJointSimCheckType(w.Joint, WheelJointType)
	return joint.wheelJoint.maxMotorTorque
}

func getWheelJointForce(world *world, base *jointSim) Vec2 {
	qA := getBodyTransform(world, base.bodyIdA).Q

	localAxisA := RotateVector(base.localFrameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA := RotateVector(qA, localAxisA)
	perpA := LeftPerp(axisA)

	joint := &base.wheelJoint

	perpForce := float32(world.invH * joint.perpImpulse)
	axialForce := float32(world.invH * (joint.springImpulse + joint.lowerImpulse - joint.upperImpulse))

	force := Add(Vec2{X: float32((perpForce) * perpA.X), Y: float32((perpForce) * perpA.Y)}, Vec2{X: float32((axialForce) * axisA.X), Y: float32((axialForce) * axisA.Y)})
	return force
}

func getWheelJointTorque(world *world, base *jointSim) float32 {
	return float32(world.invH * base.wheelJoint.motorImpulse)
}

func prepareWheelJoint(base *jointSim, context *stepContext) {
	// Linear constraint (point-to-line)
	// d = pB - pA = xB + rB - xA - rA
	// C = dot(ay, d)
	// Cdot = dot(d, cross(wA, ay)) + dot(ay, vB + cross(wB, rB) - vA - cross(wA, rA))
	//      = -dot(ay, vA) - dot(cross(d + rA, ay), wA) + dot(ay, vB) + dot(cross(rB, ay), vB)
	// J = [-ay, -cross(d + rA, ay), ay, cross(rB, ay)]
	// Spring linear constraint
	// C = dot(ax, d)
	// Cdot = = -dot(ax, vA) - dot(cross(d + rA, ax), wA) + dot(ax, vB) + dot(cross(rB, ax), vB)
	// J = [-ax -cross(d+rA, ax) ax cross(rB, ax)]
	// Motor rotational constraint
	// Cdot = wB - wA
	// J = [0 0 -1 0 0 1]

	assert(base.jointType == WheelJointType, "base->type == b2_wheelJoint")

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

	joint := &base.wheelJoint

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

	d := Add(joint.deltaCenter, Sub(rB, rA))
	axisA := RotateVector(joint.frameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	perpA := LeftPerp(axisA)

	// perpendicular constraint (keep wheel on line)
	s1 := Cross(Add(d, rA), perpA)
	s2 := Cross(rB, perpA)

	kp := mA + mB + float32(float32(iA*s1)*s1) + float32(float32(iB*s2)*s2)
	if kp > float32(0.0) {
		joint.perpMass = float32(1.0) / kp
	} else {
		joint.perpMass = float32(0.0)
	}

	// spring constraint
	a1 := Cross(Add(d, rA), axisA)
	a2 := Cross(rB, axisA)

	ka := mA + mB + float32(float32(iA*a1)*a1) + float32(float32(iB*a2)*a2)
	if ka > float32(0.0) {
		joint.axialMass = float32(1.0) / ka
	} else {
		joint.axialMass = float32(0.0)
	}

	joint.springSoftness = makeSoft(joint.hertz, joint.dampingRatio, context.h)

	km := iA + iB
	if km > float32(0.0) {
		joint.motorMass = float32(1.0) / km
	} else {
		joint.motorMass = float32(0.0)
	}

	if context.enableWarmStarting == false {
		joint.perpImpulse = float32(0.0)
		joint.springImpulse = float32(0.0)
		joint.motorImpulse = float32(0.0)
		joint.lowerImpulse = float32(0.0)
		joint.upperImpulse = float32(0.0)
	}
}

func warmStartWheelJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == WheelJointType, "base->type == b2_wheelJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.wheelJoint

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

	d := Add(Add(Sub(stateB.deltaPosition, stateA.deltaPosition), joint.deltaCenter), Sub(rB, rA))
	axisA := RotateVector(joint.frameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA = RotateVector(stateA.deltaRotation, axisA)
	perpA := LeftPerp(axisA)

	a1 := Cross(Add(d, rA), axisA)
	a2 := Cross(rB, axisA)
	s1 := Cross(Add(d, rA), perpA)
	s2 := Cross(rB, perpA)

	axialImpulse := joint.springImpulse + joint.lowerImpulse - joint.upperImpulse

	P := Add(Vec2{X: float32((axialImpulse) * axisA.X), Y: float32((axialImpulse) * axisA.Y)}, Vec2{X: float32((joint.perpImpulse) * perpA.X), Y: float32((joint.perpImpulse) * perpA.Y)})
	LA := float32(axialImpulse*a1) + float32(joint.perpImpulse*s1) + joint.motorImpulse
	LB := float32(axialImpulse*a2) + float32(joint.perpImpulse*s2) + joint.motorImpulse

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, P)
		stateA.angularVelocity -= float32(iA * LA)
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, P)
		stateB.angularVelocity += float32(iB * LB)
	}
}

func solveWheelJoint(base *jointSim, context *stepContext, useBias bool) {
	assert(base.jointType == WheelJointType, "base->type == b2_wheelJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.wheelJoint

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

	fixedRotation := (iA+iB == float32(0.0))

	// current anchors
	rA := RotateVector(stateA.deltaRotation, joint.frameA.P)
	rB := RotateVector(stateB.deltaRotation, joint.frameB.P)

	d := Add(Add(Sub(stateB.deltaPosition, stateA.deltaPosition), joint.deltaCenter), Sub(rB, rA))
	axisA := RotateVector(joint.frameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA = RotateVector(stateA.deltaRotation, axisA)
	translation := Dot(axisA, d)

	a1 := Cross(Add(d, rA), axisA)
	a2 := Cross(rB, axisA)

	// motor constraint
	if joint.enableMotor && fixedRotation == false {
		Cdot := wB - wA - joint.motorSpeed
		impulse := float32(-joint.motorMass * Cdot)
		oldImpulse := joint.motorImpulse
		maxImpulse := float32(context.h * joint.maxMotorTorque)
		joint.motorImpulse = ClampFloat(joint.motorImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = joint.motorImpulse - oldImpulse

		wA -= float32(iA * impulse)
		wB += float32(iB * impulse)
	}

	// spring constraint
	if joint.enableSpring {
		// This is a real spring and should be applied even during relax
		C := translation
		bias := float32(joint.springSoftness.biasRate * C)
		massScale := joint.springSoftness.massScale
		impulseScale := joint.springSoftness.impulseScale

		Cdot := Dot(axisA, Sub(vB, vA)) + float32(a2*wB) - float32(a1*wA)
		impulse := float32(float32(-massScale*joint.axialMass)*(Cdot+bias)) - float32(impulseScale*joint.springImpulse)
		joint.springImpulse += impulse

		P := Vec2{X: float32((impulse) * axisA.X), Y: float32((impulse) * axisA.Y)}
		LA := float32(impulse * a1)
		LB := float32(impulse * a2)

		vA = MulSub(vA, mA, P)
		wA -= float32(iA * LA)
		vB = MulAdd(vB, mB, P)
		wB += float32(iB * LB)
	}

	if joint.enableLimit {
		// Lower limit
		{
			C := translation - joint.lowerTranslation
			bias := float32(0.0)
			massScale := float32(1.0)
			impulseScale := float32(0.0)

			if C > float32(0.0) {
				// speculation
				bias = float32(C * context.invH)
			} else if useBias {
				bias = float32(base.constraintSoftness.biasRate * C)
				massScale = base.constraintSoftness.massScale
				impulseScale = base.constraintSoftness.impulseScale
			}

			Cdot := Dot(axisA, Sub(vB, vA)) + float32(a2*wB) - float32(a1*wA)
			impulse := float32(float32(-massScale*joint.axialMass)*(Cdot+bias)) - float32(impulseScale*joint.lowerImpulse)
			oldImpulse := joint.lowerImpulse
			joint.lowerImpulse = MaxFloat(oldImpulse+impulse, float32(0.0))
			impulse = joint.lowerImpulse - oldImpulse

			P := Vec2{X: float32((impulse) * axisA.X), Y: float32((impulse) * axisA.Y)}
			LA := float32(impulse * a1)
			LB := float32(impulse * a2)

			vA = MulSub(vA, mA, P)
			wA -= float32(iA * LA)
			vB = MulAdd(vB, mB, P)
			wB += float32(iB * LB)
		}

		// Upper limit
		// Note: signs are flipped to keep C positive when the constraint is satisfied.
		// This also keeps the impulse positive when the limit is active.
		{
			// sign flipped
			C := joint.upperTranslation - translation
			bias := float32(0.0)
			massScale := float32(1.0)
			impulseScale := float32(0.0)

			if C > float32(0.0) {
				// speculation
				bias = float32(C * context.invH)
			} else if useBias {
				bias = float32(base.constraintSoftness.biasRate * C)
				massScale = base.constraintSoftness.massScale
				impulseScale = base.constraintSoftness.impulseScale
			}

			// sign flipped on Cdot
			Cdot := Dot(axisA, Sub(vA, vB)) + float32(a1*wA) - float32(a2*wB)
			impulse := float32(float32(-massScale*joint.axialMass)*(Cdot+bias)) - float32(impulseScale*joint.upperImpulse)
			oldImpulse := joint.upperImpulse
			joint.upperImpulse = MaxFloat(oldImpulse+impulse, float32(0.0))
			impulse = joint.upperImpulse - oldImpulse

			P := Vec2{X: float32((impulse) * axisA.X), Y: float32((impulse) * axisA.Y)}
			LA := float32(impulse * a1)
			LB := float32(impulse * a2)

			// sign flipped on applied impulse
			vA = MulAdd(vA, mA, P)
			wA += float32(iA * LA)
			vB = MulSub(vB, mB, P)
			wB -= float32(iB * LB)
		}
	}

	// point to line constraint
	{
		perpA := LeftPerp(axisA)

		bias := float32(0.0)
		massScale := float32(1.0)
		impulseScale := float32(0.0)
		if useBias {
			C := Dot(perpA, d)
			bias = float32(base.constraintSoftness.biasRate * C)
			massScale = base.constraintSoftness.massScale
			impulseScale = base.constraintSoftness.impulseScale
		}

		s1 := Cross(Add(d, rA), perpA)
		s2 := Cross(rB, perpA)
		Cdot := Dot(perpA, Sub(vB, vA)) + float32(s2*wB) - float32(s1*wA)

		impulse := float32(float32(-massScale*joint.perpMass)*(Cdot+bias)) - float32(impulseScale*joint.perpImpulse)
		joint.perpImpulse += impulse

		P := Vec2{X: float32((impulse) * perpA.X), Y: float32((impulse) * perpA.Y)}
		LA := float32(impulse * s1)
		LB := float32(impulse * s2)

		vA = MulSub(vA, mA, P)
		wA -= float32(iA * LA)
		vB = MulAdd(vB, mB, P)
		wB += float32(iB * LB)
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

func drawWheelJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	assert(base.jointType == WheelJointType, "base->type == b2_wheelJoint")

	joint := &base.wheelJoint

	frameA := MulWorldTransforms(transformA, base.localFrameA)
	frameB := MulWorldTransforms(transformB, base.localFrameB)
	axisA := RotateVector(frameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})

	c1 := ColorGray
	c2 := ColorGreen
	c3 := ColorRed
	c4 := ColorDimGray
	c5 := ColorBlue

	draw.DrawLineFunc(frameA.P, frameB.P, c5)

	if joint.enableLimit {
		lower := OffsetPos(frameA.P, Vec2{X: float32((joint.lowerTranslation) * axisA.X), Y: float32((joint.lowerTranslation) * axisA.Y)})
		upper := OffsetPos(frameA.P, Vec2{X: float32((joint.upperTranslation) * axisA.X), Y: float32((joint.upperTranslation) * axisA.Y)})
		perp := LeftPerp(axisA)
		draw.DrawLineFunc(lower, upper, c1)
		draw.DrawLineFunc(OffsetPos(lower, Vec2{X: float32((float32(-float32(0.1) * drawScale)) * perp.X), Y: float32((float32(-float32(0.1) * drawScale)) * perp.Y)}),
			OffsetPos(lower, Vec2{X: float32((float32(float32(0.1) * drawScale)) * perp.X), Y: float32((float32(float32(0.1) * drawScale)) * perp.Y)}), c2)
		draw.DrawLineFunc(OffsetPos(upper, Vec2{X: float32((float32(-float32(0.1) * drawScale)) * perp.X), Y: float32((float32(-float32(0.1) * drawScale)) * perp.Y)}),
			OffsetPos(upper, Vec2{X: float32((float32(float32(0.1) * drawScale)) * perp.X), Y: float32((float32(float32(0.1) * drawScale)) * perp.Y)}), c3)
	} else {
		draw.DrawLineFunc(OffsetPos(frameA.P, Neg(axisA)), OffsetPos(frameA.P, axisA), c1)
	}

	draw.DrawPointFunc(frameA.P, float32(5.0), c1)
	draw.DrawPointFunc(frameB.P, float32(5.0), c4)
}
