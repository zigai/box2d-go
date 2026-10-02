// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type DistanceJoint struct {
	Joint
}

// Distance joint definition
// Connects a point on body A with a point on body B by a segment.
// Useful for ropes and springs.
type DistanceJointDef struct {
	// Base joint definition
	Base JointDef
	// The rest length of this joint. Clamped to a stable minimum value.
	Length float32
	// Enable the distance constraint to behave like a spring. If false
	// then the distance joint will be rigid, overriding the limit and motor.
	EnableSpring bool
	// The lower spring force controls how much tension it can sustain
	LowerSpringForce float32
	// The upper spring force controls how much compression it an sustain
	UpperSpringForce float32
	// The spring linear stiffness Hertz, cycles per second
	Hertz float32
	// The spring linear damping ratio, non-dimensional
	DampingRatio float32
	// Enable/disable the joint limit
	EnableLimit bool
	// Minimum length for limit. Clamped to a stable minimum value.
	MinLength float32
	// Maximum length for limit. Must be greater than or equal to the minimum length.
	MaxLength float32
	// Enable/disable the joint motor
	EnableMotor bool
	// The maximum motor force, usually in newtons
	MaxMotorForce float32
	// The desired motor speed, usually in meters per second
	MotorSpeed float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Set the rest length of a distance joint
func (d DistanceJoint) SetLength(length float32) {
	if !(checkInput(IsValidFloat(length), "b2IsValidFloat( length )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetLength(world.recording, &recArgsDistanceJointSetLength{joint: d.Joint, length: length})
	}
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint

	jnt.length = ClampFloat(length, LinearSlop(), Huge())
	jnt.impulse = 0.0
	jnt.lowerImpulse = 0.0
	jnt.upperImpulse = 0.0
}

// Get the rest length of a distance joint
func (d DistanceJoint) Length() float32 {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	return jnt.length
}

// Enable joint limit. The limit only works if the joint spring is enabled. Otherwise the joint is rigid
// and the limit has no effect.
func (d DistanceJoint) EnableLimit(enableLimit bool) {
	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointEnableLimit(world.recording, &recArgsDistanceJointEnableLimit{joint: d.Joint, enableLimit: enableLimit})
	}
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	jnt.enableLimit = enableLimit
}

// Is the distance joint limit enabled?
func (d DistanceJoint) IsLimitEnabled() bool {
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.enableLimit
}

// Set the minimum and maximum length parameters of a distance joint
func (d DistanceJoint) SetLengthRange(minLength float32, maxLength float32) {
	if !(checkInput(IsValidFloat(minLength), "b2IsValidFloat( minLength )")) {
		return
	}
	if !(checkInput(IsValidFloat(maxLength), "b2IsValidFloat( maxLength )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetLengthRange(world.recording, &recArgsDistanceJointSetLengthRange{joint: d.Joint, minLength: minLength, maxLength: maxLength})
	}
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint

	minLength = ClampFloat(minLength, LinearSlop(), Huge())
	maxLength = ClampFloat(maxLength, LinearSlop(), Huge())
	jnt.minLength = MinFloat(minLength, maxLength)
	jnt.maxLength = MaxFloat(minLength, maxLength)
}

// Get the distance joint minimum length
func (d DistanceJoint) MinLength() float32 {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	return jnt.minLength
}

// Get the distance joint maximum length
func (d DistanceJoint) MaxLength() float32 {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	return jnt.maxLength
}

// Get the current length of a distance joint
func (d DistanceJoint) CurrentLength() float32 {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)

	var world *world = getWorld(int(d.Joint.world0))
	assert(world.locked == false, "world->locked == false")
	if world.locked {
		return 0.0
	}

	// Relative to body A so the difference stays in float32 precision far from the origin
	var wxfA WorldTransform = getBodyTransform(world, base.bodyIdA)
	var transformA Transform = ToRelativeTransform(wxfA, wxfA.P)
	var transformB Transform = ToRelativeTransform(getBodyTransform(world, base.bodyIdB), wxfA.P)

	var pA Vec2 = TransformPoint(transformA, base.localFrameA.P)
	var pB Vec2 = TransformPoint(transformB, base.localFrameB.P)
	var delta Vec2 = Sub(pB, pA)
	var length float32 = Length(delta)
	return length
}

// Enable/disable the distance joint spring. When disabled the distance joint is rigid.
func (d DistanceJoint) EnableSpring(enableSpring bool) {
	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointEnableSpring(world.recording, &recArgsDistanceJointEnableSpring{joint: d.Joint, enableSpring: enableSpring})
	}
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	base.distanceJoint.enableSpring = enableSpring
}

// Is the distance joint spring enabled?
func (d DistanceJoint) IsSpringEnabled() bool {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	return base.distanceJoint.enableSpring
}

// Set the force range for the spring.
func (d DistanceJoint) SetSpringForceRange(lowerForce float32, upperForce float32) {
	if !(checkInput(IsValidFloat(lowerForce), "b2IsValidFloat( lowerForce )")) {
		return
	}
	if !(checkInput(IsValidFloat(upperForce), "b2IsValidFloat( upperForce )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetSpringForceRange(world.recording, &recArgsDistanceJointSetSpringForceRange{joint: d.Joint, lowerForce: lowerForce, upperForce: upperForce})
	}
	assert(lowerForce <= upperForce, "lowerForce <= upperForce")
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	base.distanceJoint.lowerSpringForce = MinFloat(lowerForce, upperForce)
	base.distanceJoint.upperSpringForce = MaxFloat(lowerForce, upperForce)
}

// Get the force range for the spring.
func (d DistanceJoint) SpringForceRange() (lowerForce float32, upperForce float32) {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	lowerForce = base.distanceJoint.lowerSpringForce
	upperForce = base.distanceJoint.upperSpringForce

	return
}

// Set the spring stiffness in Hertz
func (d DistanceJoint) SetSpringHertz(hertz float32) {
	if !(checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetSpringHertz(world.recording, &recArgsDistanceJointSetSpringHertz{joint: d.Joint, hertz: hertz})
	}
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	base.distanceJoint.hertz = hertz
}

// Set the spring damping ratio, non-dimensional
func (d DistanceJoint) SetSpringDampingRatio(dampingRatio float32) {
	if !(checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetSpringDampingRatio(world.recording, &recArgsDistanceJointSetSpringDampingRatio{joint: d.Joint, dampingRatio: dampingRatio})
	}
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	base.distanceJoint.dampingRatio = dampingRatio
}

// Get the spring Hertz
func (d DistanceJoint) SpringHertz() float32 {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	return jnt.hertz
}

// Get the spring damping ratio
func (d DistanceJoint) SpringDampingRatio() float32 {
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	return jnt.dampingRatio
}

// Get the current spring force.
func (d DistanceJoint) SpringForce() float32 {
	var world *world = getWorld(int(d.Joint.world0))
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	var jnt *distanceJoint = &base.distanceJoint
	if jnt.enableSpring {
		return float32(jnt.impulse * world.invH)
	}

	return 0.0
}

// Enable/disable the distance joint motor
func (d DistanceJoint) EnableMotor(enableMotor bool) {
	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointEnableMotor(world.recording, &recArgsDistanceJointEnableMotor{joint: d.Joint, enableMotor: enableMotor})
	}
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	if enableMotor != jnt.distanceJoint.enableMotor {
		jnt.distanceJoint.enableMotor = enableMotor
		jnt.distanceJoint.motorImpulse = 0.0
	}
}

// Is the distance joint motor enabled?
func (d DistanceJoint) IsMotorEnabled() bool {
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.enableMotor
}

// Set the distance joint motor speed, usually in meters per second
func (d DistanceJoint) SetMotorSpeed(motorSpeed float32) {
	if !(checkInput(IsValidFloat(motorSpeed), "b2IsValidFloat( motorSpeed )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetMotorSpeed(world.recording, &recArgsDistanceJointSetMotorSpeed{joint: d.Joint, motorSpeed: motorSpeed})
	}
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	jnt.distanceJoint.motorSpeed = motorSpeed
}

// Get the distance joint motor speed, usually in meters per second
func (d DistanceJoint) MotorSpeed() float32 {
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.motorSpeed
}

// Get the distance joint current motor force, usually in newtons
func (d DistanceJoint) MotorForce() float32 {
	var world *world = getWorld(int(d.Joint.world0))
	var base *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	return float32(world.invH * base.distanceJoint.motorImpulse)
}

// Set the distance joint maximum motor force, usually in newtons
func (d DistanceJoint) SetMaxMotorForce(force float32) {
	if !(checkInput(IsValidFloat(force), "b2IsValidFloat( force )")) {
		return
	}

	var world *world = getWorld(int(d.Joint.world0))
	if world.recording != nil {
		recWriteDistanceJointSetMaxMotorForce(world.recording, &recArgsDistanceJointSetMaxMotorForce{joint: d.Joint, force: force})
	}
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	jnt.distanceJoint.maxMotorForce = force
}

// Get the distance joint maximum motor force, usually in newtons
func (d DistanceJoint) MaxMotorForce() float32 {
	var jnt *jointSim = getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.maxMotorForce
}

func getDistanceJointForce(world *world, base *jointSim) Vec2 {
	var jnt *distanceJoint = &base.distanceJoint

	var worldTransformA WorldTransform = getBodyTransform(world, base.bodyIdA)
	var transformA Transform = ToRelativeTransform(worldTransformA, worldTransformA.P)
	var worldTransformB WorldTransform = getBodyTransform(world, base.bodyIdB)
	var transformB Transform = ToRelativeTransform(worldTransformB, worldTransformA.P)

	var pA Vec2 = TransformPoint(transformA, base.localFrameA.P)
	var pB Vec2 = TransformPoint(transformB, base.localFrameB.P)
	var d Vec2 = Sub(pB, pA)
	var axis Vec2 = Normalize(d)
	var force float32 = float32((jnt.impulse + jnt.lowerImpulse - jnt.upperImpulse + jnt.motorImpulse) * world.invH)
	return MulSV(force, axis)
}

func prepareDistanceJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

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

	var jnt *distanceJoint = &base.distanceJoint

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

	// initial anchors in world space
	jnt.anchorA = RotateVector(bodySimA.transform.Q, Sub(base.localFrameA.P, bodySimA.localCenter))
	jnt.anchorB = RotateVector(bodySimB.transform.Q, Sub(base.localFrameB.P, bodySimB.localCenter))
	jnt.deltaCenter = SubPos(bodySimB.center, bodySimA.center)

	var rA Vec2 = jnt.anchorA
	var rB Vec2 = jnt.anchorB
	var separation Vec2 = Add(Sub(rB, rA), jnt.deltaCenter)
	var axis Vec2 = Normalize(separation)

	// compute effective mass
	var crA float32 = Cross(rA, axis)
	var crB float32 = Cross(rB, axis)
	var k float32 = mA + mB + float32(float32(iA*crA)*crA) + float32(float32(iB*crB)*crB)
	if k > 0.0 {
		jnt.axialMass = 1.0 / k
	} else {
		jnt.axialMass = 0.0
	}

	jnt.distanceSoftness = makeSoft(jnt.hertz, jnt.dampingRatio, context.h)

	if context.enableWarmStarting == false {
		jnt.impulse = 0.0
		jnt.lowerImpulse = 0.0
		jnt.upperImpulse = 0.0
		jnt.motorImpulse = 0.0
	}
}

func warmStartDistanceJoint(base *jointSim, context *stepContext) {
	assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	var mA float32 = base.invMassA
	var mB float32 = base.invMassB
	var iA float32 = base.invIA
	var iB float32 = base.invIB

	// dummy state for static bodies
	var dummyState bodyState = identityBodyState

	var jnt *distanceJoint = &base.distanceJoint
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

	var rA Vec2 = RotateVector(stateA.deltaRotation, jnt.anchorA)
	var rB Vec2 = RotateVector(stateB.deltaRotation, jnt.anchorB)

	var ds Vec2 = Add(Sub(stateB.deltaPosition, stateA.deltaPosition), Sub(rB, rA))
	var separation Vec2 = Add(jnt.deltaCenter, ds)
	var axis Vec2 = Normalize(separation)

	var axialImpulse float32 = jnt.impulse + jnt.lowerImpulse - jnt.upperImpulse + jnt.motorImpulse
	var P Vec2 = MulSV(axialImpulse, axis)

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = MulSub(stateA.linearVelocity, mA, P)
		stateA.angularVelocity -= float32(iA * Cross(rA, P))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = MulAdd(stateB.linearVelocity, mB, P)
		stateB.angularVelocity += float32(iB * Cross(rB, P))
	}
}

func solveDistanceJoint(base *jointSim, context *stepContext, useBias bool) {
	assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	var mA float32 = base.invMassA
	var mB float32 = base.invMassB
	var iA float32 = base.invIA
	var iB float32 = base.invIB

	// dummy state for static bodies
	var dummyState bodyState = identityBodyState

	var jnt *distanceJoint = &base.distanceJoint
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

	// current anchors
	var rA Vec2 = RotateVector(stateA.deltaRotation, jnt.anchorA)
	var rB Vec2 = RotateVector(stateB.deltaRotation, jnt.anchorB)

	// current separation
	var ds Vec2 = Add(Sub(stateB.deltaPosition, stateA.deltaPosition), Sub(rB, rA))
	var separation Vec2 = Add(jnt.deltaCenter, ds)

	var length float32 = Length(separation)
	var axis Vec2 = Normalize(separation)

	// joint is soft if
	// - spring is enabled
	// - and (joint limit is disabled or limits are not equal)
	if jnt.enableSpring && (jnt.minLength < jnt.maxLength || jnt.enableLimit == false) {
		// spring
		if jnt.hertz > 0.0 {
			// Cdot = dot(u, v + cross(w, r))
			var vr Vec2 = Add(Sub(vB, vA), Sub(CrossSV(wB, rB), CrossSV(wA, rA)))
			var cdot float32 = Dot(axis, vr)
			var c float32 = length - jnt.length
			var bias float32 = float32(jnt.distanceSoftness.biasRate * c)

			var m float32 = float32(jnt.distanceSoftness.massScale * jnt.axialMass)
			var oldImpulse float32 = jnt.impulse
			var impulse float32 = float32(-m*(cdot+bias)) - float32(jnt.distanceSoftness.impulseScale*oldImpulse)

			var h float32 = context.h
			jnt.impulse = ClampFloat(jnt.impulse+impulse, float32(jnt.lowerSpringForce*h), float32(jnt.upperSpringForce*h))
			impulse = jnt.impulse - oldImpulse

			var P Vec2 = MulSV(impulse, axis)
			vA = MulSub(vA, mA, P)
			wA -= float32(iA * Cross(rA, P))
			vB = MulAdd(vB, mB, P)
			wB += float32(iB * Cross(rB, P))
		}

		if jnt.enableMotor {
			var vr Vec2 = Add(Sub(vB, vA), Sub(CrossSV(wB, rB), CrossSV(wA, rA)))
			var Cdot float32 = Dot(axis, vr)
			var impulse float32 = float32(jnt.axialMass * (jnt.motorSpeed - Cdot))
			var oldImpulse float32 = jnt.motorImpulse
			var maxImpulse float32 = float32(context.h * jnt.maxMotorForce)
			jnt.motorImpulse = ClampFloat(jnt.motorImpulse+impulse, -maxImpulse, maxImpulse)
			impulse = jnt.motorImpulse - oldImpulse

			var P Vec2 = MulSV(impulse, axis)
			vA = MulSub(vA, mA, P)
			wA -= float32(iA * Cross(rA, P))
			vB = MulAdd(vB, mB, P)
			wB += float32(iB * Cross(rB, P))
		}

		if jnt.enableLimit {
			// lower limit
			{
				var vr Vec2 = Add(Sub(vB, vA), Sub(CrossSV(wB, rB), CrossSV(wA, rA)))
				var cdot float32 = Dot(axis, vr)

				var c float32 = length - jnt.minLength

				var bias float32 = 0.0
				var massScale float32 = 1.0
				var impulseScale float32 = 0.0
				if c > 0.0 {
					// speculative
					bias = float32(c * context.invH)
				} else if useBias {
					bias = float32(base.constraintSoftness.biasRate * c)
					massScale = base.constraintSoftness.massScale
					impulseScale = base.constraintSoftness.impulseScale
				}

				var impulse float32 = float32(float32(-massScale*jnt.axialMass)*(cdot+bias)) - float32(impulseScale*jnt.lowerImpulse)
				var newImpulse float32 = MaxFloat(0.0, jnt.lowerImpulse+impulse)
				impulse = newImpulse - jnt.lowerImpulse
				jnt.lowerImpulse = newImpulse

				var P Vec2 = MulSV(impulse, axis)
				vA = MulSub(vA, mA, P)
				wA -= float32(iA * Cross(rA, P))
				vB = MulAdd(vB, mB, P)
				wB += float32(iB * Cross(rB, P))
			}

			// upper
			{
				var vr Vec2 = Add(Sub(vA, vB), Sub(CrossSV(wA, rA), CrossSV(wB, rB)))
				var Cdot float32 = Dot(axis, vr)

				var C float32 = jnt.maxLength - length

				var bias float32 = 0.0
				var massScale float32 = 1.0
				var impulseScale float32 = 0.0
				if C > 0.0 {
					// speculative
					bias = float32(C * context.invH)
				} else if useBias {
					bias = float32(base.constraintSoftness.biasRate * C)
					massScale = base.constraintSoftness.massScale
					impulseScale = base.constraintSoftness.impulseScale
				}

				var impulse float32 = float32(float32(-massScale*jnt.axialMass)*(Cdot+bias)) - float32(impulseScale*jnt.upperImpulse)
				var newImpulse float32 = MaxFloat(0.0, jnt.upperImpulse+impulse)
				impulse = newImpulse - jnt.upperImpulse
				jnt.upperImpulse = newImpulse

				var P Vec2 = MulSV(-impulse, axis)
				vA = MulSub(vA, mA, P)
				wA -= float32(iA * Cross(rA, P))
				vB = MulAdd(vB, mB, P)
				wB += float32(iB * Cross(rB, P))
			}
		}
	} else {
		// rigid constraint
		var vr Vec2 = Add(Sub(vB, vA), Sub(CrossSV(wB, rB), CrossSV(wA, rA)))
		var Cdot float32 = Dot(axis, vr)

		var C float32 = length - jnt.length

		var bias float32 = 0.0
		var massScale float32 = 1.0
		var impulseScale float32 = 0.0
		if useBias {
			bias = float32(base.constraintSoftness.biasRate * C)
			massScale = base.constraintSoftness.massScale
			impulseScale = base.constraintSoftness.impulseScale
		}

		var impulse float32 = float32(float32(-massScale*jnt.axialMass)*(Cdot+bias)) - float32(impulseScale*jnt.impulse)
		jnt.impulse += impulse

		var P Vec2 = MulSV(impulse, axis)
		vA = MulSub(vA, mA, P)
		wA -= float32(iA * Cross(rA, P))
		vB = MulAdd(vB, mB, P)
		wB += float32(iB * Cross(rB, P))
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

func drawDistanceJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform) {
	assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	var jnt *distanceJoint = &base.distanceJoint

	var pA Pos = TransformWorldPoint(transformA, base.localFrameA.P)
	var pB Pos = TransformWorldPoint(transformB, base.localFrameB.P)

	var axis Vec2 = Normalize(SubPos(pB, pA))

	if jnt.minLength < jnt.maxLength && jnt.enableLimit {
		var pMin Pos = OffsetPos(pA, MulSV(jnt.minLength, axis))
		var pMax Pos = OffsetPos(pA, MulSV(jnt.maxLength, axis))
		var offset Vec2 = MulSV(float32(0.05*GetLengthUnitsPerMeter()), RightPerp(axis))

		if jnt.minLength > LinearSlop() {
			draw.DrawLineFunc(OffsetPos(pMin, Neg(offset)), OffsetPos(pMin, offset), ColorLightGreen)
		}

		if jnt.maxLength < Huge() {
			draw.DrawLineFunc(OffsetPos(pMax, Neg(offset)), OffsetPos(pMax, offset), ColorRed)
		}

		if jnt.minLength > LinearSlop() && jnt.maxLength < Huge() {
			draw.DrawLineFunc(pMin, pMax, ColorGray)
		}
	}

	draw.DrawLineFunc(pA, pB, ColorWhite)
	draw.DrawPointFunc(pA, 4.0, ColorWhite)
	draw.DrawPointFunc(pB, 4.0, ColorWhite)

	if jnt.hertz > 0.0 && jnt.enableSpring {
		var pRest Pos = OffsetPos(pA, MulSV(jnt.length, axis))
		draw.DrawPointFunc(pRest, 4.0, ColorBlue)
	}
}
