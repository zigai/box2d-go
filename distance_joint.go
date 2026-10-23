// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "github.com/zigai/box2d/internal/core"

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

// Set the rest length of a distance joint.
func (d DistanceJoint) SetLength(length float32) {
	checkInput(IsValidFloat(length), "b2IsValidFloat( length )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetLength(world.recording, &recArgsDistanceJointSetLength{joint: d.Joint, length: length})
	}

	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	jnt.length = clampFloat(length, LinearSlop(), Huge())
	jnt.impulse = 0.0
	jnt.lowerImpulse = 0.0
	jnt.upperImpulse = 0.0
}

// Get the rest length of a distance joint.
func (d DistanceJoint) Length() float32 {
	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	return jnt.length
}

// Enable joint limit. The limit only works if the joint spring is enabled. Otherwise the joint is rigid
// and the limit has no effect.
func (d DistanceJoint) SetLimitEnabled(enableLimit bool) {
	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointEnableLimit(world.recording, &recArgsDistanceJointEnableLimit{joint: d.Joint, enableLimit: enableLimit})
	}

	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	jnt.enableLimit = enableLimit
}

// Is the distance joint limit enabled?
func (d DistanceJoint) LimitEnabled() bool {
	jnt := getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.enableLimit
}

// Set the minimum and maximum length parameters of a distance joint.
func (d DistanceJoint) SetLengthRange(minLength float32, maxLength float32) {
	checkInput(IsValidFloat(minLength), "b2IsValidFloat( minLength )")
	checkInput(IsValidFloat(maxLength), "b2IsValidFloat( maxLength )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetLengthRange(world.recording, &recArgsDistanceJointSetLengthRange{joint: d.Joint, minLength: minLength, maxLength: maxLength})
	}

	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	minLength = clampFloat(minLength, LinearSlop(), Huge())
	maxLength = clampFloat(maxLength, LinearSlop(), Huge())
	jnt.minLength = minFloat(minLength, maxLength)
	jnt.maxLength = maxFloat(minLength, maxLength)
}

// Get the distance joint minimum length.
func (d DistanceJoint) MinLength() float32 {
	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	return jnt.minLength
}

// Get the distance joint maximum length.
func (d DistanceJoint) MaxLength() float32 {
	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	return jnt.maxLength
}

// Get the current length of a distance joint.
func (d DistanceJoint) CurrentLength() float32 {
	base := getJointSimCheckType(d.Joint, DistanceJointType)

	world := d.Joint.world
	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return 0.0
	}

	// Relative to body A so the difference stays in float32 precision far from the origin
	var (
		wxfA       = getBodyTransform(world, base.bodyIDA)
		transformA = wxfA.Relative(wxfA.P)
		transformB = getBodyTransform(world, base.bodyIDB).Relative(wxfA.P)
	)

	var (
		pA     = transformA.Apply(base.localFrameA.P)
		pB     = transformB.Apply(base.localFrameB.P)
		delta  = pB.Sub(pA)
		length = delta.Length()
	)

	return length
}

// Enable/disable the distance joint spring. When disabled the distance joint is rigid.
func (d DistanceJoint) SetSpringEnabled(enableSpring bool) {
	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointEnableSpring(world.recording, &recArgsDistanceJointEnableSpring{joint: d.Joint, enableSpring: enableSpring})
	}

	base := getJointSimCheckType(d.Joint, DistanceJointType)

	base.distanceJoint.enableSpring = enableSpring
}

// Is the distance joint spring enabled?
func (d DistanceJoint) SpringEnabled() bool {
	base := getJointSimCheckType(d.Joint, DistanceJointType)
	return base.distanceJoint.enableSpring
}

// Set the force range for the spring.
func (d DistanceJoint) SetSpringForceRange(lowerForce float32, upperForce float32) {
	checkInput(IsValidFloat(lowerForce), "b2IsValidFloat( lowerForce )")
	checkInput(IsValidFloat(upperForce), "b2IsValidFloat( upperForce )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetSpringForceRange(world.recording, &recArgsDistanceJointSetSpringForceRange{joint: d.Joint, lowerForce: lowerForce, upperForce: upperForce})
	}

	core.Assert(lowerForce <= upperForce, "lowerForce <= upperForce")

	base := getJointSimCheckType(d.Joint, DistanceJointType)

	base.distanceJoint.lowerSpringForce = minFloat(lowerForce, upperForce)
	base.distanceJoint.upperSpringForce = maxFloat(lowerForce, upperForce)
}

// Get the force range for the spring.
func (d DistanceJoint) SpringForceRange() (float32, float32) {
	base := getJointSimCheckType(d.Joint, DistanceJointType)

	lowerForce := base.distanceJoint.lowerSpringForce
	upperForce := base.distanceJoint.upperSpringForce

	return lowerForce, upperForce
}

// Set the spring stiffness in Hertz.
func (d DistanceJoint) SetSpringHertz(hertz float32) {
	checkInput(IsValidFloat(hertz), "b2IsValidFloat( hertz )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetSpringHertz(world.recording, &recArgsDistanceJointSetSpringHertz{joint: d.Joint, hertz: hertz})
	}

	base := getJointSimCheckType(d.Joint, DistanceJointType)

	base.distanceJoint.hertz = hertz
}

// Set the spring damping ratio, non-dimensional.
func (d DistanceJoint) SetSpringDampingRatio(dampingRatio float32) {
	checkInput(IsValidFloat(dampingRatio), "b2IsValidFloat( dampingRatio )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetSpringDampingRatio(world.recording, &recArgsDistanceJointSetSpringDampingRatio{joint: d.Joint, dampingRatio: dampingRatio})
	}

	base := getJointSimCheckType(d.Joint, DistanceJointType)

	base.distanceJoint.dampingRatio = dampingRatio
}

// Get the spring Hertz.
func (d DistanceJoint) SpringHertz() float32 {
	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	return jnt.hertz
}

// Get the spring damping ratio.
func (d DistanceJoint) SpringDampingRatio() float32 {
	var (
		base = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt  = &base.distanceJoint
	)

	return jnt.dampingRatio
}

// Get the current spring force.
func (d DistanceJoint) SpringForce() float32 {
	var (
		world = d.Joint.world
		base  = getJointSimCheckType(d.Joint, DistanceJointType)
		jnt   = &base.distanceJoint
	)
	if jnt.enableSpring {
		return float32(jnt.impulse * world.invH)
	}

	return 0.0
}

// Enable/disable the distance joint motor.
func (d DistanceJoint) SetMotorEnabled(enableMotor bool) {
	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointEnableMotor(world.recording, &recArgsDistanceJointEnableMotor{joint: d.Joint, enableMotor: enableMotor})
	}

	jnt := getJointSimCheckType(d.Joint, DistanceJointType)
	if enableMotor != jnt.distanceJoint.enableMotor {
		jnt.distanceJoint.enableMotor = enableMotor
		jnt.distanceJoint.motorImpulse = 0.0
	}
}

// Is the distance joint motor enabled?
func (d DistanceJoint) MotorEnabled() bool {
	jnt := getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.enableMotor
}

// Set the distance joint motor speed, usually in meters per second.
func (d DistanceJoint) SetMotorSpeed(motorSpeed float32) {
	checkInput(IsValidFloat(motorSpeed), "b2IsValidFloat( motorSpeed )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetMotorSpeed(world.recording, &recArgsDistanceJointSetMotorSpeed{joint: d.Joint, motorSpeed: motorSpeed})
	}

	jnt := getJointSimCheckType(d.Joint, DistanceJointType)

	jnt.distanceJoint.motorSpeed = motorSpeed
}

// Get the distance joint motor speed, usually in meters per second.
func (d DistanceJoint) MotorSpeed() float32 {
	jnt := getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.motorSpeed
}

// Get the distance joint current motor force, usually in newtons.
func (d DistanceJoint) MotorForce() float32 {
	var (
		world = d.Joint.world
		base  = getJointSimCheckType(d.Joint, DistanceJointType)
	)

	return float32(world.invH * base.distanceJoint.motorImpulse)
}

// Set the distance joint maximum motor force, usually in newtons.
func (d DistanceJoint) SetMaxMotorForce(force float32) {
	checkInput(IsValidFloat(force), "b2IsValidFloat( force )")

	world := d.Joint.world
	if world.recording != nil {
		recWriteDistanceJointSetMaxMotorForce(world.recording, &recArgsDistanceJointSetMaxMotorForce{joint: d.Joint, force: force})
	}

	jnt := getJointSimCheckType(d.Joint, DistanceJointType)

	jnt.distanceJoint.maxMotorForce = force
}

// Get the distance joint maximum motor force, usually in newtons.
func (d DistanceJoint) MaxMotorForce() float32 {
	jnt := getJointSimCheckType(d.Joint, DistanceJointType)
	return jnt.distanceJoint.maxMotorForce
}

func getDistanceJointForce(world *World, base *jointSim) Vec2 {
	jnt := &base.distanceJoint

	var (
		worldTransformA = getBodyTransform(world, base.bodyIDA)
		transformA      = worldTransformA.Relative(worldTransformA.P)
		worldTransformB = getBodyTransform(world, base.bodyIDB)
		transformB      = worldTransformB.Relative(worldTransformA.P)
	)

	var (
		pA    = transformA.Apply(base.localFrameA.P)
		pB    = transformB.Apply(base.localFrameB.P)
		d     = pB.Sub(pA)
		axis  = d.Normalize()
		force = float32((jnt.impulse + jnt.lowerImpulse - jnt.upperImpulse + jnt.motorImpulse) * world.invH)
	)

	return axis.Scale(force)
}

func prepareDistanceJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	// chase body id to the solver set where the body lives
	var (
		idA = base.bodyIDA
		idB = base.bodyIDB
	)

	var (
		world = context.world
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

	jnt := &base.distanceJoint

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
	jnt.anchorA = bodySimA.transform.Q.Rotate(base.localFrameA.P.Sub(bodySimA.localCenter))
	jnt.anchorB = bodySimB.transform.Q.Rotate(base.localFrameB.P.Sub(bodySimB.localCenter))
	jnt.deltaCenter = bodySimB.center.Sub(bodySimA.center)

	var (
		rA         = jnt.anchorA
		rB         = jnt.anchorB
		separation = rB.Sub(rA).Add(jnt.deltaCenter)
		axis       = separation.Normalize()
	)

	// compute effective mass
	var (
		crA = rA.Cross(axis)
		crB = rB.Cross(axis)
		k   = mA + mB + float32(float32(iA*crA)*crA) + float32(float32(iB*crB)*crB)
	)
	if k > 0.0 {
		jnt.axialMass = 1.0 / k
	} else {
		jnt.axialMass = 0.0
	}

	jnt.distanceSoftness = makeSoft(jnt.hertz, jnt.dampingRatio, context.h)

	if !context.enableWarmStarting {
		jnt.impulse = 0.0
		jnt.lowerImpulse = 0.0
		jnt.upperImpulse = 0.0
		jnt.motorImpulse = 0.0
	}
}

func warmStartDistanceJoint(base *jointSim, context *stepContext) {
	core.Assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	var (
		mA = base.invMassA
		mB = base.invMassB
		iA = base.invIA
		iB = base.invIB
	)

	// dummy state for static bodies
	dummyState := identityBodyState

	var (
		jnt    = &base.distanceJoint
		stateA = jointSolverBodyState(context, jnt.indexA, &dummyState)
	)

	stateB := jointSolverBodyState(context, jnt.indexB, &dummyState)

	var (
		rA = stateA.deltaRotation.Rotate(jnt.anchorA)
		rB = stateB.deltaRotation.Rotate(jnt.anchorB)
	)

	var (
		ds         = stateB.deltaPosition.Sub(stateA.deltaPosition).Add(rB.Sub(rA))
		separation = jnt.deltaCenter.Add(ds)
		axis       = separation.Normalize()
	)

	var (
		axialImpulse  = jnt.impulse + jnt.lowerImpulse - jnt.upperImpulse + jnt.motorImpulse
		impulseVector = axis.Scale(axialImpulse)
	)

	if stateA.flags&dynamicFlag != 0 {
		stateA.linearVelocity = stateA.linearVelocity.MulSub(mA, impulseVector)
		stateA.angularVelocity -= float32(iA * rA.Cross(impulseVector))
	}

	if stateB.flags&dynamicFlag != 0 {
		stateB.linearVelocity = stateB.linearVelocity.MulAdd(mB, impulseVector)
		stateB.angularVelocity += float32(iB * rB.Cross(impulseVector))
	}
}

func solveDistanceJoint(base *jointSim, context *stepContext, useBias bool) {
	core.Assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	var (
		mA = base.invMassA
		mB = base.invMassB
		iA = base.invIA
		iB = base.invIB
	)

	// dummy state for static bodies
	dummyState := identityBodyState

	var (
		jnt    = &base.distanceJoint
		stateA = jointSolverBodyState(context, jnt.indexA, &dummyState)
	)

	stateB := jointSolverBodyState(context, jnt.indexB, &dummyState)

	var (
		vA = stateA.linearVelocity
		wA = stateA.angularVelocity
		vB = stateB.linearVelocity
		wB = stateB.angularVelocity
	)

	// current anchors
	var (
		rA = stateA.deltaRotation.Rotate(jnt.anchorA)
		rB = stateB.deltaRotation.Rotate(jnt.anchorB)
	)

	// current separation
	var (
		ds         = stateB.deltaPosition.Sub(stateA.deltaPosition).Add(rB.Sub(rA))
		separation = jnt.deltaCenter.Add(ds)
	)

	var (
		length = separation.Length()
		axis   = separation.Normalize()
	)

	// joint is soft if
	// - spring is enabled
	// - and (joint limit is disabled or limits are not equal)
	if jnt.enableSpring && (jnt.minLength < jnt.maxLength || !jnt.enableLimit) {
		// spring
		if jnt.hertz > 0.0 {
			// Cdot = dot(u, v + cross(w, r))
			var (
				vr   = vB.Sub(vA).Add(rB.ScalarCross(wB).Sub(rA.ScalarCross(wA)))
				cdot = axis.Dot(vr)
				c    = length - jnt.length
				bias = float32(jnt.distanceSoftness.biasRate * c)
			)

			var (
				m          = float32(jnt.distanceSoftness.massScale * jnt.axialMass)
				oldImpulse = jnt.impulse
				impulse    = float32(-m*(cdot+bias)) - float32(jnt.distanceSoftness.impulseScale*oldImpulse)
			)

			h := context.h

			jnt.impulse = clampFloat(jnt.impulse+impulse, float32(jnt.lowerSpringForce*h), float32(jnt.upperSpringForce*h))
			impulse = jnt.impulse - oldImpulse

			impulseVector := axis.Scale(impulse)

			vA = vA.MulSub(mA, impulseVector)
			wA -= float32(iA * rA.Cross(impulseVector))
			vB = vB.MulAdd(mB, impulseVector)
			wB += float32(iB * rB.Cross(impulseVector))
		}

		if jnt.enableMotor {
			var (
				vr            = vB.Sub(vA).Add(rB.ScalarCross(wB).Sub(rA.ScalarCross(wA)))
				velocityError = axis.Dot(vr)
				impulse       = float32(jnt.axialMass * (jnt.motorSpeed - velocityError))
				oldImpulse    = jnt.motorImpulse
				maxImpulse    = float32(context.h * jnt.maxMotorForce)
			)

			jnt.motorImpulse = clampFloat(jnt.motorImpulse+impulse, -maxImpulse, maxImpulse)
			impulse = jnt.motorImpulse - oldImpulse

			impulseVector := axis.Scale(impulse)

			vA = vA.MulSub(mA, impulseVector)
			wA -= float32(iA * rA.Cross(impulseVector))
			vB = vB.MulAdd(mB, impulseVector)
			wB += float32(iB * rB.Cross(impulseVector))
		}

		if jnt.enableLimit {
			// lower limit
			{
				var (
					vr   = vB.Sub(vA).Add(rB.ScalarCross(wB).Sub(rA.ScalarCross(wA)))
					cdot = axis.Dot(vr)
				)

				c := length - jnt.minLength

				bias, massScale, impulseScale := jointLimitBias(c, context.invH, useBias, base.constraintSoftness)

				var (
					impulse    = float32(float32(-massScale*jnt.axialMass)*(cdot+bias)) - float32(impulseScale*jnt.lowerImpulse)
					newImpulse = maxFloat(0.0, jnt.lowerImpulse+impulse)
				)

				impulse = newImpulse - jnt.lowerImpulse
				jnt.lowerImpulse = newImpulse

				impulseVector := axis.Scale(impulse)

				vA = vA.MulSub(mA, impulseVector)
				wA -= float32(iA * rA.Cross(impulseVector))
				vB = vB.MulAdd(mB, impulseVector)
				wB += float32(iB * rB.Cross(impulseVector))
			}

			// upper
			{
				var (
					vr            = vA.Sub(vB).Add(rA.ScalarCross(wA).Sub(rB.ScalarCross(wB)))
					velocityError = axis.Dot(vr)
				)

				separation := jnt.maxLength - length

				bias, massScale, impulseScale := jointLimitBias(separation, context.invH, useBias, base.constraintSoftness)

				var (
					impulse    = float32(float32(-massScale*jnt.axialMass)*(velocityError+bias)) - float32(impulseScale*jnt.upperImpulse)
					newImpulse = maxFloat(0.0, jnt.upperImpulse+impulse)
				)

				impulse = newImpulse - jnt.upperImpulse
				jnt.upperImpulse = newImpulse

				impulseVector := axis.Scale(-impulse)

				vA = vA.MulSub(mA, impulseVector)
				wA -= float32(iA * rA.Cross(impulseVector))
				vB = vB.MulAdd(mB, impulseVector)
				wB += float32(iB * rB.Cross(impulseVector))
			}
		}
	} else {
		// rigid constraint
		var (
			vr            = vB.Sub(vA).Add(rB.ScalarCross(wB).Sub(rA.ScalarCross(wA)))
			velocityError = axis.Dot(vr)
		)

		separation := length - jnt.length

		var (
			bias         float32 = 0.0
			massScale    float32 = 1.0
			impulseScale float32 = 0.0
		)

		if useBias {
			bias = float32(base.constraintSoftness.biasRate * separation)
			massScale = base.constraintSoftness.massScale
			impulseScale = base.constraintSoftness.impulseScale
		}

		impulse := float32(float32(-massScale*jnt.axialMass)*(velocityError+bias)) - float32(impulseScale*jnt.impulse)

		jnt.impulse += impulse

		impulseVector := axis.Scale(impulse)

		vA = vA.MulSub(mA, impulseVector)
		wA -= float32(iA * rA.Cross(impulseVector))
		vB = vB.MulAdd(mB, impulseVector)
		wB += float32(iB * rB.Cross(impulseVector))
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
	core.Assert(base.jointType == DistanceJointType, "base->type == b2_distanceJoint")

	jnt := &base.distanceJoint

	var (
		pA = transformA.Apply(base.localFrameA.P)
		pB = transformB.Apply(base.localFrameB.P)
	)

	axis := pB.Sub(pA).Normalize()

	if jnt.minLength < jnt.maxLength && jnt.enableLimit {
		var (
			pMin   = pA.Add(axis.Scale(jnt.minLength))
			pMax   = pA.Add(axis.Scale(jnt.maxLength))
			offset = axis.RightPerp().Scale(float32(0.05 * LengthUnitsPerMeter()))
		)

		if jnt.minLength > LinearSlop() {
			draw.DrawLine(pMin.Add(offset.Neg()), pMin.Add(offset), ColorLightGreen)
		}

		if jnt.maxLength < Huge() {
			draw.DrawLine(pMax.Add(offset.Neg()), pMax.Add(offset), ColorRed)
		}

		if jnt.minLength > LinearSlop() && jnt.maxLength < Huge() {
			draw.DrawLine(pMin, pMax, ColorGray)
		}
	}

	draw.DrawLine(pA, pB, ColorWhite)
	draw.DrawPoint(pA, 4.0, ColorWhite)
	draw.DrawPoint(pB, 4.0, ColorWhite)

	if jnt.hertz > 0.0 && jnt.enableSpring {
		pRest := pA.Add(axis.Scale(jnt.length))
		draw.DrawPoint(pRest, 4.0, ColorBlue)
	}
}

func jointLimitBias(separation, invH float32, useBias bool, constraintSoftness softness) (float32, float32, float32) {
	if separation > 0.0 {
		// speculative
		return float32(separation * invH), 1.0, 0.0
	}

	if useBias {
		return float32(constraintSoftness.biasRate * separation), constraintSoftness.massScale, constraintSoftness.impulseScale
	}

	return 0.0, 1.0, 0.0
}
