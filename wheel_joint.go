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
	todo()
}

// Is the wheel joint spring enabled?
func (w WheelJoint) IsSpringEnabled() bool {
	todo()
	return false
}

// Set the wheel joint stiffness in Hertz
func (w WheelJoint) SetSpringHertz(hertz float32) {
	todo()
}

// Get the wheel joint stiffness in Hertz
func (w WheelJoint) SpringHertz() float32 {
	todo()
	return 0
}

// Set the wheel joint damping ratio, non-dimensional
func (w WheelJoint) SetSpringDampingRatio(dampingRatio float32) {
	todo()
}

// Get the wheel joint damping ratio, non-dimensional
func (w WheelJoint) SpringDampingRatio() float32 {
	todo()
	return 0
}

// Enable/disable the wheel joint limit
func (w WheelJoint) EnableLimit(enableLimit bool) {
	todo()
}

// Is the wheel joint limit enabled?
func (w WheelJoint) IsLimitEnabled() bool {
	todo()
	return false
}

// Get the wheel joint lower limit
func (w WheelJoint) LowerLimit() float32 {
	todo()
	return 0
}

// Get the wheel joint upper limit
func (w WheelJoint) UpperLimit() float32 {
	todo()
	return 0
}

// Set the wheel joint limits
func (w WheelJoint) SetLimits(lower float32, upper float32) {
	todo()
}

// Enable/disable the wheel joint motor
func (w WheelJoint) EnableMotor(enableMotor bool) {
	todo()
}

// Is the wheel joint motor enabled?
func (w WheelJoint) IsMotorEnabled() bool {
	todo()
	return false
}

// Set the wheel joint motor speed in radians per second
func (w WheelJoint) SetMotorSpeed(motorSpeed float32) {
	todo()
}

// Get the wheel joint motor speed in radians per second
func (w WheelJoint) MotorSpeed() float32 {
	todo()
	return 0
}

// Get the wheel joint current motor torque, usually in newton-meters
func (w WheelJoint) MotorTorque() float32 {
	todo()
	return 0
}

// Set the wheel joint maximum motor torque, usually in newton-meters
func (w WheelJoint) SetMaxMotorTorque(torque float32) {
	todo()
}

// Get the wheel joint maximum motor torque, usually in newton-meters
func (w WheelJoint) MaxMotorTorque() float32 {
	todo()
	return 0
}

func getWheelJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func getWheelJointTorque(world *world, base *jointSim) float32 {
	todo()
	return 0
}

func prepareWheelJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartWheelJoint(base *jointSim, context *stepContext) {
	todo()
}

func solveWheelJoint(base *jointSim, context *stepContext, useBias bool) {
	todo()
}

func drawWheelJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	todo()
}
