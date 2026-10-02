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
	todo()
}

// Is the prismatic joint spring enabled or not?
func (p PrismaticJoint) IsSpringEnabled() bool {
	todo()
	return false
}

// Set the prismatic joint stiffness in Hertz.
// This should usually be less than a quarter of the simulation rate. For example, if the simulation
// runs at 60Hz then the joint stiffness should be 15Hz or less.
func (p PrismaticJoint) SetSpringHertz(hertz float32) {
	todo()
}

// Get the prismatic joint stiffness in Hertz
func (p PrismaticJoint) SpringHertz() float32 {
	todo()
	return 0
}

// Set the prismatic joint damping ratio (non-dimensional)
func (p PrismaticJoint) SetSpringDampingRatio(dampingRatio float32) {
	todo()
}

// Get the prismatic spring damping ratio (non-dimensional)
func (p PrismaticJoint) SpringDampingRatio() float32 {
	todo()
	return 0
}

// Set the prismatic joint spring target angle, usually in meters
func (p PrismaticJoint) SetTargetTranslation(translation float32) {
	todo()
}

// Get the prismatic joint spring target translation, usually in meters
func (p PrismaticJoint) TargetTranslation() float32 {
	todo()
	return 0
}

// Enable/disable a prismatic joint limit
func (p PrismaticJoint) EnableLimit(enableLimit bool) {
	todo()
}

// Is the prismatic joint limit enabled?
func (p PrismaticJoint) IsLimitEnabled() bool {
	todo()
	return false
}

// Get the prismatic joint lower limit
func (p PrismaticJoint) LowerLimit() float32 {
	todo()
	return 0
}

// Get the prismatic joint upper limit
func (p PrismaticJoint) UpperLimit() float32 {
	todo()
	return 0
}

// Set the prismatic joint limits
func (p PrismaticJoint) SetLimits(lower float32, upper float32) {
	todo()
}

// Enable/disable a prismatic joint motor
func (p PrismaticJoint) EnableMotor(enableMotor bool) {
	todo()
}

// Is the prismatic joint motor enabled?
func (p PrismaticJoint) IsMotorEnabled() bool {
	todo()
	return false
}

// Set the prismatic joint motor speed, usually in meters per second
func (p PrismaticJoint) SetMotorSpeed(motorSpeed float32) {
	todo()
}

// Get the prismatic joint motor speed, usually in meters per second
func (p PrismaticJoint) MotorSpeed() float32 {
	todo()
	return 0
}

// Get the prismatic joint current motor force, usually in newtons
func (p PrismaticJoint) MotorForce() float32 {
	todo()
	return 0
}

// Set the prismatic joint maximum motor force, usually in newtons
func (p PrismaticJoint) SetMaxMotorForce(force float32) {
	todo()
}

// Get the prismatic joint maximum motor force, usually in newtons
func (p PrismaticJoint) MaxMotorForce() float32 {
	todo()
	return 0
}

// Get the current joint translation, usually in meters.
func (p PrismaticJoint) Translation() float32 {
	todo()
	return 0
}

// Get the current joint translation speed, usually in meters per second.
func (p PrismaticJoint) Speed() float32 {
	todo()
	return 0
}

func getPrismaticJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func getPrismaticJointTorque(world *world, base *jointSim) float32 {
	todo()
	return 0
}

func preparePrismaticJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartPrismaticJoint(base *jointSim, context *stepContext) {
	todo()
}

func solvePrismaticJoint(base *jointSim, context *stepContext, useBias bool) {
	todo()
}

func drawPrismaticJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	todo()
}
