// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type PrismaticJoint struct {
	Joint
}

// Prismatic joint definition
// Body B may slide along the x-axis in local frame A. Body B cannot rotate relative to body A.
// The joint translation is zero when the local frame origins coincide in world space.
type PrismaticJointDef struct {
	// Base joint definition
	Base JointDef
	// Enable a linear spring along the prismatic joint axis
	EnableSpring bool
	// The spring stiffness Hertz, cycles per second
	Hertz float32
	// The spring damping ratio, non-dimensional
	DampingRatio float32
	// The target translation for the joint in meters. The spring-damper will drive
	// to this translation.
	TargetTranslation float32
	// Enable/disable the joint limit
	EnableLimit bool
	// The lower translation limit
	LowerTranslation float32
	// The upper translation limit
	UpperTranslation float32
	// Enable/disable the joint motor
	EnableMotor bool
	// The maximum motor force, typically in newtons
	MaxMotorForce float32
	// The desired motor speed, typically in meters per second
	MotorSpeed float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Enable/disable the joint spring.
func (p PrismaticJoint) EnableSpring(enableSpring bool) {
	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointEnableSpring(world.recording, &recArgsPrismaticJointEnableSpring{joint: p.Joint, enableSpring: enableSpring})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	if enableSpring != joint.prismaticJoint.enableSpring {
		joint.prismaticJoint.enableSpring = enableSpring
		joint.prismaticJoint.springImpulse = float32(0.0)
	}
}

// Is the prismatic joint spring enabled or not?
func (p PrismaticJoint) IsSpringEnabled() bool {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.enableSpring
}

// Set the prismatic joint stiffness in Hertz.
// This should usually be less than a quarter of the simulation rate. For example, if the simulation
// runs at 60Hz then the joint stiffness should be 15Hz or less.
func (p PrismaticJoint) SetSpringHertz(hertz float32) {
	if !checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointSetSpringHertz(world.recording, &recArgsPrismaticJointSetSpringHertz{joint: p.Joint, hertz: hertz})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	joint.prismaticJoint.hertz = hertz
}

// Get the prismatic joint stiffness in Hertz
func (p PrismaticJoint) SpringHertz() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.hertz
}

// Set the prismatic joint damping ratio (non-dimensional)
func (p PrismaticJoint) SetSpringDampingRatio(dampingRatio float32) {
	if !checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointSetSpringDampingRatio(world.recording, &recArgsPrismaticJointSetSpringDampingRatio{joint: p.Joint, dampingRatio: dampingRatio})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	joint.prismaticJoint.dampingRatio = dampingRatio
}

// Get the prismatic spring damping ratio (non-dimensional)
func (p PrismaticJoint) SpringDampingRatio() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.dampingRatio
}

// Set the prismatic joint spring target angle, usually in meters
func (p PrismaticJoint) SetTargetTranslation(translation float32) {
	if !checkInput(IsValidFloat(translation), "b2IsValidFloat( translation )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointSetTargetTranslation(world.recording, &recArgsPrismaticJointSetTargetTranslation{joint: p.Joint, translation: translation})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	joint.prismaticJoint.targetTranslation = translation
}

// Get the prismatic joint spring target translation, usually in meters
func (p PrismaticJoint) TargetTranslation() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.targetTranslation
}

// Enable/disable a prismatic joint limit
func (p PrismaticJoint) EnableLimit(enableLimit bool) {
	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointEnableLimit(world.recording, &recArgsPrismaticJointEnableLimit{joint: p.Joint, enableLimit: enableLimit})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	if enableLimit != joint.prismaticJoint.enableLimit {
		joint.prismaticJoint.enableLimit = enableLimit
		joint.prismaticJoint.lowerImpulse = float32(0.0)
		joint.prismaticJoint.upperImpulse = float32(0.0)
	}
}

// Is the prismatic joint limit enabled?
func (p PrismaticJoint) IsLimitEnabled() bool {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.enableLimit
}

// Get the prismatic joint lower limit
func (p PrismaticJoint) LowerLimit() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.lowerTranslation
}

// Get the prismatic joint upper limit
func (p PrismaticJoint) UpperLimit() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.upperTranslation
}

// Set the prismatic joint limits
func (p PrismaticJoint) SetLimits(lower float32, upper float32) {
	if !checkInput(IsValidFloat(lower), "b2IsValidFloat( lower )") {
		return
	}
	if !checkInput(IsValidFloat(upper), "b2IsValidFloat( upper )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointSetLimits(world.recording, &recArgsPrismaticJointSetLimits{joint: p.Joint, lower: lower, upper: upper})
	}
	assert(lower <= upper, "lower <= upper")

	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	joint.prismaticJoint.lowerTranslation = MinFloat(lower, upper)
	joint.prismaticJoint.upperTranslation = MaxFloat(lower, upper)
}

// Enable/disable a prismatic joint motor
func (p PrismaticJoint) EnableMotor(enableMotor bool) {
	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointEnableMotor(world.recording, &recArgsPrismaticJointEnableMotor{joint: p.Joint, enableMotor: enableMotor})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	if enableMotor != joint.prismaticJoint.enableMotor {
		joint.prismaticJoint.enableMotor = enableMotor
		joint.prismaticJoint.motorImpulse = float32(0.0)
	}
}

// Is the prismatic joint motor enabled?
func (p PrismaticJoint) IsMotorEnabled() bool {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.enableMotor
}

// Set the prismatic joint motor speed, usually in meters per second
func (p PrismaticJoint) SetMotorSpeed(motorSpeed float32) {
	if !checkInput(IsValidFloat(motorSpeed), "b2IsValidFloat( motorSpeed )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointSetMotorSpeed(world.recording, &recArgsPrismaticJointSetMotorSpeed{joint: p.Joint, motorSpeed: motorSpeed})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	joint.prismaticJoint.motorSpeed = motorSpeed
}

// Get the prismatic joint motor speed, usually in meters per second
func (p PrismaticJoint) MotorSpeed() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.motorSpeed
}

// Get the prismatic joint current motor force, usually in newtons
func (p PrismaticJoint) MotorForce() float32 {
	world := getWorld(int(p.Joint.world0))
	base := getJointSimCheckType(p.Joint, PrismaticJointType)
	return float32(world.invH * base.prismaticJoint.motorImpulse)
}

// Set the prismatic joint maximum motor force, usually in newtons
func (p PrismaticJoint) SetMaxMotorForce(force float32) {
	if !checkInput(IsValidFloat(force), "b2IsValidFloat( force )") {
		return
	}

	world := getWorld(int(p.Joint.world0))
	if world.recording != nil {
		recWritePrismaticJointSetMaxMotorForce(world.recording, &recArgsPrismaticJointSetMaxMotorForce{joint: p.Joint, force: force})
	}
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	joint.prismaticJoint.maxMotorForce = force
}

// Get the prismatic joint maximum motor force, usually in newtons
func (p PrismaticJoint) MaxMotorForce() float32 {
	joint := getJointSimCheckType(p.Joint, PrismaticJointType)
	return joint.prismaticJoint.maxMotorForce
}

// Get the current joint translation, usually in meters.
func (p PrismaticJoint) Translation() float32 {
	world := getWorld(int(p.Joint.world0))
	jointSim := getJointSimCheckType(p.Joint, PrismaticJointType)
	// Relative to body A so the difference stays in float precision far from the origin
	wxfA := getBodyTransform(world, jointSim.bodyIdA)
	transformA := ToRelativeTransform(wxfA, wxfA.P)
	transformB := ToRelativeTransform(getBodyTransform(world, jointSim.bodyIdB), wxfA.P)

	localAxisA := RotateVector(jointSim.localFrameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA := RotateVector(transformA.Q, localAxisA)
	pA := TransformPoint(transformA, jointSim.localFrameA.P)
	pB := TransformPoint(transformB, jointSim.localFrameB.P)
	d := Sub(pB, pA)
	translation := Dot(d, axisA)
	return translation
}

// Get the current joint translation speed, usually in meters per second.
func (p PrismaticJoint) Speed() float32 {
	world := getWorld(int(p.Joint.world0))
	joint := getJointFullId(world, p.Joint)
	assert(joint.jointType == PrismaticJointType, "joint->type == b2_prismaticJoint")
	base := getJointSim(world, joint)
	assert(base.jointType == PrismaticJointType, "base->type == b2_prismaticJoint")

	bodyA := &world.bodies[base.bodyIdA]
	bodyB := &world.bodies[base.bodyIdB]
	bodySimA := getBodySim(world, bodyA)
	bodySimB := getBodySim(world, bodyB)
	bodyStateA := getBodyState(world, bodyA)
	bodyStateB := getBodyState(world, bodyB)

	qA := bodySimA.transform.Q
	qB := bodySimB.transform.Q

	localAxisA := RotateVector(base.localFrameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA := RotateVector(qA, localAxisA)
	rA := RotateVector(qA, Sub(base.localFrameA.P, bodySimA.localCenter))
	rB := RotateVector(qB, Sub(base.localFrameB.P, bodySimB.localCenter))

	// Difference the centers in double so the speed stays exact far from the origin
	dc := SubPos(bodySimB.center, bodySimA.center)
	d := Add(dc, Sub(rB, rA))

	var vA Vec2
	if bodyStateA != nil {
		vA = bodyStateA.linearVelocity
	} else {
		vA = Vec2{}
	}
	var vB Vec2
	if bodyStateB != nil {
		vB = bodyStateB.linearVelocity
	} else {
		vB = Vec2{}
	}
	var wA float32
	if bodyStateA != nil {
		wA = bodyStateA.angularVelocity
	} else {
		wA = float32(0.0)
	}
	var wB float32
	if bodyStateB != nil {
		wB = bodyStateB.angularVelocity
	} else {
		wB = float32(0.0)
	}

	vRel := Sub(Add(vB, Vec2{X: float32(-(wB) * rB.Y), Y: float32((wB) * rB.X)}), Add(vA, Vec2{X: float32(-(wA) * rA.Y), Y: float32((wA) * rA.X)}))
	speed := Dot(d, Vec2{X: float32(-(wA) * axisA.Y), Y: float32((wA) * axisA.X)}) + Dot(axisA, vRel)
	return speed
}

func getPrismaticJointForce(world *world, base *jointSim) Vec2 {
	qA := getBodyTransform(world, base.bodyIdA).Q

	joint := &base.prismaticJoint

	localAxisA := RotateVector(base.localFrameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA := RotateVector(qA, localAxisA)
	perpA := LeftPerp(axisA)

	invH := world.invH
	perpForce := float32(invH * joint.impulse.X)
	axialForce := float32(invH * (joint.motorImpulse + joint.lowerImpulse - joint.upperImpulse))

	force := Add(Vec2{X: float32((perpForce) * perpA.X), Y: float32((perpForce) * perpA.Y)}, Vec2{X: float32((axialForce) * axisA.X), Y: float32((axialForce) * axisA.Y)})
	return force
}

func getPrismaticJointTorque(world *world, base *jointSim) float32 {
	return float32(world.invH * base.prismaticJoint.impulse.Y)
}

func preparePrismaticJoint(base *jointSim, context *stepContext) {
	// Linear constraint (point-to-line)
	// d = pB - pA = xB + rB - xA - rA
	// C = dot(perp, d)
	// Cdot = dot(d, cross(wA, perp)) + dot(perp, vB + cross(wB, rB) - vA - cross(wA, rA))
	//      = -dot(perp, vA) - dot(cross(rA + d, perp), wA) + dot(perp, vB) + dot(cross(rB, perp), vB)
	// J = [-perp, -cross(rA + d, perp), perp, cross(rB, perp)]
	//
	// Angular constraint
	// C = aB - aA + a_initial
	// Cdot = wB - wA
	// J = [0 0 -1 0 0 1]
	//
	// K = J * invM * JT
	//
	// J = [-a -sA a sB]
	//     [0  -1  0  1]
	// a = perp
	// sA = cross(rA + d, a) = cross(pB - xA, a)
	// sB = cross(rB, a) = cross(pB - xB, a)
	// Motor/Limit linear constraint
	// C = dot(axA, d)
	// Cdot = -dot(axA, vA) - dot(cross(rA + d, axA), wA) + dot(axA, vB) + dot(cross(rB, axA), vB)
	// J = [-axA -cross(rA + d, axA) axA cross(rB, ax1)]
	// Predictive limit is applied even when the limit is not active.
	// Prevents a constraint speed that can lead to a constraint error in one time step.
	// Want C2 = C1 + h * Cdot >= 0
	// Or:
	// Cdot + C1/h >= 0
	// I do not apply a negative constraint error because that is handled in position correction.
	// So:
	// Cdot + max(C1, 0)/h >= 0
	// Block Solver
	// We develop a block solver that includes the angular and linear constraints. This makes the limit stiffer.
	//
	// The Jacobian has 2 rows:
	// J = [-uT -s1 uT s2] // linear
	//     [0   -1   0  1] // angular
	//
	// u = perp
	// s1 = cross(d + r1, u), s2 = cross(r2, u)
	// a1 = cross(d + r1, v), a2 = cross(r2, v)

	assert(base.jointType == PrismaticJointType, "base->type == b2_prismaticJoint")

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

	joint := &base.prismaticJoint
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

	joint.springSoftness = makeSoft(joint.hertz, joint.dampingRatio, context.h)

	if context.enableWarmStarting == false {
		joint.impulse = Vec2{}
		joint.springImpulse = float32(0.0)
		joint.motorImpulse = float32(0.0)
		joint.lowerImpulse = float32(0.0)
		joint.upperImpulse = float32(0.0)
	}
}

func warmStartPrismaticJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == PrismaticJointType, "base->type == b2_prismaticJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.prismaticJoint

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

	// impulse is applied at anchor point on body B
	a1 := Cross(Add(rA, d), axisA)
	a2 := Cross(rB, axisA)
	axialImpulse := joint.springImpulse + joint.motorImpulse + joint.lowerImpulse - joint.upperImpulse

	// perpendicular constraint
	perpA := LeftPerp(axisA)
	s1 := Cross(Add(rA, d), perpA)
	s2 := Cross(rB, perpA)
	perpImpulse := joint.impulse.X
	angleImpulse := joint.impulse.Y

	P := Add(Vec2{X: float32((axialImpulse) * axisA.X), Y: float32((axialImpulse) * axisA.Y)}, Vec2{X: float32((perpImpulse) * perpA.X), Y: float32((perpImpulse) * perpA.Y)})
	LA := float32(axialImpulse*a1) + float32(perpImpulse*s1) + angleImpulse
	LB := float32(axialImpulse*a2) + float32(perpImpulse*s2) + angleImpulse

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, P)
		stateA.angularVelocity -= float32(iA * LA)
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, P)
		stateB.angularVelocity += float32(iB * LB)
	}
}

func solvePrismaticJoint(base *jointSim, context *stepContext, useBias bool) {
	assert(base.jointType == PrismaticJointType, "base->type == b2_prismaticJoint")

	mA := base.invMassA
	mB := base.invMassB
	iA := base.invIA
	iB := base.invIB

	// dummy state for static bodies
	dummyState := identityBodyState

	joint := &base.prismaticJoint

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

	qA := MulRot(stateA.deltaRotation, joint.frameA.Q)
	qB := MulRot(stateB.deltaRotation, joint.frameB.Q)
	relQ := InvMulRot(qA, qB)

	// current anchors
	rA := RotateVector(stateA.deltaRotation, joint.frameA.P)
	rB := RotateVector(stateB.deltaRotation, joint.frameB.P)

	d := Add(Add(Sub(stateB.deltaPosition, stateA.deltaPosition), joint.deltaCenter), Sub(rB, rA))

	axisA := RotateVector(joint.frameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})
	axisA = RotateVector(stateA.deltaRotation, axisA)
	translation := Dot(axisA, d)

	// These scalars are for torques generated by axial forces
	a1 := Cross(Add(rA, d), axisA)
	a2 := Cross(rB, axisA)

	k := mA + mB + float32(float32(iA*a1)*a1) + float32(float32(iB*a2)*a2)
	var axialMass float32
	if k > float32(0.0) {
		axialMass = float32(1.0) / k
	} else {
		axialMass = float32(0.0)
	}

	softness := base.constraintSoftness

	// spring constraint
	if joint.enableSpring {
		// This is a real spring and should be applied even during relax
		C := translation - joint.targetTranslation
		bias := float32(joint.springSoftness.biasRate * C)
		massScale := joint.springSoftness.massScale
		impulseScale := joint.springSoftness.impulseScale

		Cdot := Dot(axisA, Sub(vB, vA)) + float32(a2*wB) - float32(a1*wA)
		deltaImpulse := float32(float32(-massScale*axialMass)*(Cdot+bias)) - float32(impulseScale*joint.springImpulse)
		joint.springImpulse += deltaImpulse

		P := Vec2{X: float32((deltaImpulse) * axisA.X), Y: float32((deltaImpulse) * axisA.Y)}
		LA := float32(deltaImpulse * a1)
		LB := float32(deltaImpulse * a2)

		vA = MulSub(vA, mA, P)
		wA -= float32(iA * LA)
		vB = MulAdd(vB, mB, P)
		wB += float32(iB * LB)
	}

	// Solve motor constraint
	if joint.enableMotor {
		Cdot := Dot(axisA, Sub(vB, vA)) + float32(a2*wB) - float32(a1*wA)
		impulse := float32(axialMass * (joint.motorSpeed - Cdot))
		oldImpulse := joint.motorImpulse
		maxImpulse := float32(context.h * joint.maxMotorForce)
		joint.motorImpulse = ClampFloat(joint.motorImpulse+impulse, -maxImpulse, maxImpulse)
		impulse = joint.motorImpulse - oldImpulse

		P := Vec2{X: float32((impulse) * axisA.X), Y: float32((impulse) * axisA.Y)}
		LA := float32(impulse * a1)
		LB := float32(impulse * a2)

		vA = MulSub(vA, mA, P)
		wA -= float32(iA * LA)
		vB = MulAdd(vB, mB, P)
		wB += float32(iB * LB)
	}

	if joint.enableLimit {
		// Clamp the speculative distance to a reasonable value
		speculativeDistance := float32(float32(0.25) * (joint.upperTranslation - joint.lowerTranslation))

		// Lower limit
		{
			C := translation - joint.lowerTranslation

			if C < speculativeDistance {
				bias := float32(0.0)
				massScale := float32(1.0)
				impulseScale := float32(0.0)

				if C > float32(0.0) {
					// speculation
					safe := GetLengthUnitsPerMeter()
					bias = float32(MinFloat(C, safe) * context.invH)
				} else if useBias {
					bias = float32(softness.biasRate * C)
					massScale = softness.massScale
					impulseScale = softness.impulseScale
				}

				oldImpulse := joint.lowerImpulse
				Cdot := Dot(axisA, Sub(vB, vA)) + float32(a2*wB) - float32(a1*wA)
				deltaImpulse := float32(float32(-axialMass*massScale)*(Cdot+bias)) - float32(impulseScale*oldImpulse)
				joint.lowerImpulse = MaxFloat(oldImpulse+deltaImpulse, float32(0.0))
				deltaImpulse = joint.lowerImpulse - oldImpulse

				P := Vec2{X: float32((deltaImpulse) * axisA.X), Y: float32((deltaImpulse) * axisA.Y)}
				LA := float32(deltaImpulse * a1)
				LB := float32(deltaImpulse * a2)

				vA = MulSub(vA, mA, P)
				wA -= float32(iA * LA)
				vB = MulAdd(vB, mB, P)
				wB += float32(iB * LB)
			} else {
				joint.lowerImpulse = float32(0.0)
			}
		}

		// Upper limit
		// Note: signs are flipped to keep C positive when the constraint is satisfied.
		// This also keeps the impulse positive when the limit is active.
		{
			// sign flipped
			C := joint.upperTranslation - translation

			if C < speculativeDistance {
				bias := float32(0.0)
				massScale := float32(1.0)
				impulseScale := float32(0.0)

				if C > float32(0.0) {
					// speculation
					safe := GetLengthUnitsPerMeter()
					bias = float32(MinFloat(C, safe) * context.invH)
				} else if useBias {
					bias = float32(softness.biasRate * C)
					massScale = softness.massScale
					impulseScale = softness.impulseScale
				}

				oldImpulse := joint.upperImpulse

				// sign flipped
				Cdot := Dot(axisA, Sub(vA, vB)) + float32(a1*wA) - float32(a2*wB)
				deltaImpulse := float32(float32(-axialMass*massScale)*(Cdot+bias)) - float32(impulseScale*oldImpulse)
				joint.upperImpulse = MaxFloat(oldImpulse+deltaImpulse, float32(0.0))
				deltaImpulse = joint.upperImpulse - oldImpulse

				P := Vec2{X: float32((deltaImpulse) * axisA.X), Y: float32((deltaImpulse) * axisA.Y)}
				LA := float32(deltaImpulse * a1)
				LB := float32(deltaImpulse * a2)

				// sign flipped
				vA = MulAdd(vA, mA, P)
				wA += float32(iA * LA)
				vB = MulSub(vB, mB, P)
				wB -= float32(iB * LB)
			} else {
				joint.upperImpulse = float32(0.0)
			}
		}
	}

	// Solve the prismatic constraint in block form
	{
		perpA := LeftPerp(axisA)

		// These scalars are for torques generated by the perpendicular constraint force
		s1 := Cross(Add(d, rA), perpA)
		s2 := Cross(rB, perpA)

		var Cdot Vec2
		Cdot.X = Dot(perpA, Sub(vB, vA)) + float32(s2*wB) - float32(s1*wA)
		Cdot.Y = wB - wA

		bias := Vec2{}
		massScale := float32(1.0)
		impulseScale := float32(0.0)
		if useBias {
			var C Vec2
			C.X = Dot(perpA, d)
			C.Y = relQ.Angle()

			bias = Vec2{X: float32((softness.biasRate) * C.X), Y: float32((softness.biasRate) * C.Y)}
			massScale = softness.massScale
			impulseScale = softness.impulseScale
		}

		k11 := mA + mB + float32(float32(iA*s1)*s1) + float32(float32(iB*s2)*s2)
		k12 := float32(iA*s1) + float32(iB*s2)
		k22 := iA + iB
		if k22 == float32(0.0) {
			// For bodies with fixed rotation.
			k22 = float32(1.0)
		}

		K := Mat22{Cx: Vec2{X: k11, Y: k12}, Cy: Vec2{X: k12, Y: k22}}

		b := Solve22(K, Add(Cdot, bias))
		b.X = float32(b.X)
		b.Y = float32(b.Y)
		var deltaImpulse Vec2
		deltaImpulse.X = float32(-massScale*b.X) - float32(impulseScale*joint.impulse.X)
		deltaImpulse.Y = float32(-massScale*b.Y) - float32(impulseScale*joint.impulse.Y)

		joint.impulse.X += deltaImpulse.X
		joint.impulse.Y += deltaImpulse.Y

		P := Vec2{X: float32((deltaImpulse.X) * perpA.X), Y: float32((deltaImpulse.X) * perpA.Y)}
		LA := float32(deltaImpulse.X*s1) + deltaImpulse.Y
		LB := float32(deltaImpulse.X*s2) + deltaImpulse.Y

		vA = MulSub(vA, mA, P)
		wA -= float32(iA * LA)
		vB = MulAdd(vB, mB, P)
		wB += float32(iB * LB)
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

func drawPrismaticJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	assert(base.jointType == PrismaticJointType, "base->type == b2_prismaticJoint")

	joint := &base.prismaticJoint

	frameA := MulWorldTransforms(transformA, base.localFrameA)
	frameB := MulWorldTransforms(transformB, base.localFrameB)
	axisA := RotateVector(frameA.Q, Vec2{X: float32(1.0), Y: float32(0.0)})

	draw.DrawLineFunc(frameA.P, frameB.P, ColorDimGray)

	if joint.enableLimit {
		b := float32(float32(0.25) * drawScale)
		lower := OffsetPos(frameA.P, Vec2{X: float32((joint.lowerTranslation) * axisA.X), Y: float32((joint.lowerTranslation) * axisA.Y)})
		upper := OffsetPos(frameA.P, Vec2{X: float32((joint.upperTranslation) * axisA.X), Y: float32((joint.upperTranslation) * axisA.Y)})
		perp := LeftPerp(axisA)
		draw.DrawLineFunc(lower, upper, ColorGray)
		draw.DrawLineFunc(OffsetPos(lower, Vec2{X: float32((-b) * perp.X), Y: float32((-b) * perp.Y)}), OffsetPos(lower, Vec2{X: float32((b) * perp.X), Y: float32((b) * perp.Y)}), ColorGreen)
		draw.DrawLineFunc(OffsetPos(upper, Vec2{X: float32((-b) * perp.X), Y: float32((-b) * perp.Y)}), OffsetPos(upper, Vec2{X: float32((b) * perp.X), Y: float32((b) * perp.Y)}), ColorRed)
	} else {
		draw.DrawLineFunc(OffsetPos(frameA.P, Neg(axisA)), OffsetPos(frameA.P, axisA), ColorGray)
	}

	if joint.enableSpring {
		p := OffsetPos(frameA.P, Vec2{X: float32((joint.targetTranslation) * axisA.X), Y: float32((joint.targetTranslation) * axisA.Y)})
		draw.DrawPointFunc(p, float32(8.0), ColorViolet)
	}

	draw.DrawPointFunc(frameA.P, float32(5.0), ColorGray)
	draw.DrawPointFunc(frameB.P, float32(5.0), ColorBlue)
}
