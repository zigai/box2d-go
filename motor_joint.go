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
	todo()
}

// Get the desired relative linear velocity in meters per second
func (m MotorJoint) LinearVelocity() Vec2 {
	todo()
	return Vec2{}
}

// Set the desired relative angular velocity in radians per second
func (m MotorJoint) SetAngularVelocity(velocity float32) {
	todo()
}

// Get the desired relative angular velocity in radians per second
func (m MotorJoint) AngularVelocity() float32 {
	todo()
	return 0
}

// Set the motor joint maximum torque, usually in newton-meters
func (m MotorJoint) SetMaxVelocityTorque(maxTorque float32) {
	todo()
}

// Get the motor joint maximum torque, usually in newton-meters
func (m MotorJoint) MaxVelocityTorque() float32 {
	todo()
	return 0
}

// Set the motor joint maximum force, usually in newtons
func (m MotorJoint) SetMaxVelocityForce(maxForce float32) {
	todo()
}

// Get the motor joint maximum force, usually in newtons
func (m MotorJoint) MaxVelocityForce() float32 {
	todo()
	return 0
}

// Set the spring linear hertz stiffness
func (m MotorJoint) SetLinearHertz(hertz float32) {
	todo()
}

// Get the spring linear hertz stiffness
func (m MotorJoint) LinearHertz() float32 {
	todo()
	return 0
}

// Set the spring linear damping ratio. Use 1.0 for critical damping.
func (m MotorJoint) SetLinearDampingRatio(damping float32) {
	todo()
}

// Get the spring linear damping ratio.
func (m MotorJoint) LinearDampingRatio() float32 {
	todo()
	return 0
}

// Set the spring angular hertz stiffness
func (m MotorJoint) SetAngularHertz(hertz float32) {
	todo()
}

// Get the spring angular hertz stiffness
func (m MotorJoint) AngularHertz() float32 {
	todo()
	return 0
}

// Set the spring angular damping ratio. Use 1.0 for critical damping.
func (m MotorJoint) SetAngularDampingRatio(damping float32) {
	todo()
}

// Get the spring angular damping ratio.
func (m MotorJoint) AngularDampingRatio() float32 {
	todo()
	return 0
}

// Set the maximum spring force in newtons.
func (m MotorJoint) SetMaxSpringForce(maxForce float32) {
	todo()
}

// Get the maximum spring force in newtons.
func (m MotorJoint) MaxSpringForce() float32 {
	todo()
	return 0
}

// Set the maximum spring torque in newtons * meters
func (m MotorJoint) SetMaxSpringTorque(maxTorque float32) {
	todo()
}

// Get the maximum spring torque in newtons * meters
func (m MotorJoint) MaxSpringTorque() float32 {
	todo()
	return 0
}

func getMotorJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func getMotorJointTorque(world *world, base *jointSim) float32 {
	todo()
	return 0
}

func prepareMotorJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartMotorJoint(base *jointSim, context *stepContext) {
	todo()
}

func solveMotorJoint(base *jointSim, context *stepContext) {
	todo()
}
