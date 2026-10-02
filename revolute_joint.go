// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

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

// Enable/disable the revolute joint spring
func (r RevoluteJoint) EnableSpring(enableSpring bool) {
	todo()
}

// It the revolute angular spring enabled?
func (r RevoluteJoint) IsSpringEnabled() bool {
	todo()
	return false
}

// Set the revolute joint spring stiffness in Hertz
func (r RevoluteJoint) SetSpringHertz(hertz float32) {
	todo()
}

// Get the revolute joint spring stiffness in Hertz
func (r RevoluteJoint) SpringHertz() float32 {
	todo()
	return 0
}

// Set the revolute joint spring damping ratio, non-dimensional
func (r RevoluteJoint) SetSpringDampingRatio(dampingRatio float32) {
	todo()
}

// Get the revolute joint spring damping ratio, non-dimensional
func (r RevoluteJoint) SpringDampingRatio() float32 {
	todo()
	return 0
}

// Set the revolute joint spring target angle, radians
func (r RevoluteJoint) SetTargetAngle(angle float32) {
	todo()
}

// Get the revolute joint spring target angle, radians
func (r RevoluteJoint) TargetAngle() float32 {
	todo()
	return 0
}

// Get the revolute joint current angle in radians relative to the reference angle
func (r RevoluteJoint) Angle() float32 {
	todo()
	return 0
}

// Enable/disable the revolute joint limit
func (r RevoluteJoint) EnableLimit(enableLimit bool) {
	todo()
}

// Is the revolute joint limit enabled?
func (r RevoluteJoint) IsLimitEnabled() bool {
	todo()
	return false
}

// Get the revolute joint lower limit in radians
func (r RevoluteJoint) LowerLimit() float32 {
	todo()
	return 0
}

// Get the revolute joint upper limit in radians
func (r RevoluteJoint) UpperLimit() float32 {
	todo()
	return 0
}

// Set the revolute joint limits in radians. It is expected that lower <= upper
// and that -0.99 * B2_PI <= lower && upper <= -0.99 * B2_PI.
func (r RevoluteJoint) SetLimits(lower float32, upper float32) {
	todo()
}

// Enable/disable a revolute joint motor
func (r RevoluteJoint) EnableMotor(enableMotor bool) {
	todo()
}

// Is the revolute joint motor enabled?
func (r RevoluteJoint) IsMotorEnabled() bool {
	todo()
	return false
}

// Set the revolute joint motor speed in radians per second
func (r RevoluteJoint) SetMotorSpeed(motorSpeed float32) {
	todo()
}

// Get the revolute joint motor speed in radians per second
func (r RevoluteJoint) MotorSpeed() float32 {
	todo()
	return 0
}

// Get the revolute joint current motor torque, usually in newton-meters
func (r RevoluteJoint) MotorTorque() float32 {
	todo()
	return 0
}

// Set the revolute joint maximum motor torque, usually in newton-meters
func (r RevoluteJoint) SetMaxMotorTorque(torque float32) {
	todo()
}

// Get the revolute joint maximum motor torque, usually in newton-meters
func (r RevoluteJoint) MaxMotorTorque() float32 {
	todo()
	return 0
}

func getRevoluteJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func getRevoluteJointTorque(world *world, base *jointSim) float32 {
	todo()
	return 0
}

func prepareRevoluteJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartRevoluteJoint(base *jointSim, context *stepContext) {
	todo()
}

func solveRevoluteJoint(base *jointSim, context *stepContext, useBias bool) {
	todo()
}

func drawRevoluteJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	todo()
}
