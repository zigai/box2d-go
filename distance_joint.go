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
	todo()
}

// Get the rest length of a distance joint
func (d DistanceJoint) Length() float32 {
	todo()
	return 0
}

// Enable joint limit. The limit only works if the joint spring is enabled. Otherwise the joint is rigid
// and the limit has no effect.
func (d DistanceJoint) EnableLimit(enableLimit bool) {
	todo()
}

// Is the distance joint limit enabled?
func (d DistanceJoint) IsLimitEnabled() bool {
	todo()
	return false
}

// Set the minimum and maximum length parameters of a distance joint
func (d DistanceJoint) SetLengthRange(minLength float32, maxLength float32) {
	todo()
}

// Get the distance joint minimum length
func (d DistanceJoint) MinLength() float32 {
	todo()
	return 0
}

// Get the distance joint maximum length
func (d DistanceJoint) MaxLength() float32 {
	todo()
	return 0
}

// Get the current length of a distance joint
func (d DistanceJoint) CurrentLength() float32 {
	todo()
	return 0
}

// Enable/disable the distance joint spring. When disabled the distance joint is rigid.
func (d DistanceJoint) EnableSpring(enableSpring bool) {
	todo()
}

// Is the distance joint spring enabled?
func (d DistanceJoint) IsSpringEnabled() bool {
	todo()
	return false
}

// Set the force range for the spring.
func (d DistanceJoint) SetSpringForceRange(lowerForce float32, upperForce float32) {
	todo()
}

// Get the force range for the spring.
func (d DistanceJoint) SpringForceRange() (lowerForce float32, upperForce float32) {
	todo()
	return 0, 0
}

// Set the spring stiffness in Hertz
func (d DistanceJoint) SetSpringHertz(hertz float32) {
	todo()
}

// Set the spring damping ratio, non-dimensional
func (d DistanceJoint) SetSpringDampingRatio(dampingRatio float32) {
	todo()
}

// Get the spring Hertz
func (d DistanceJoint) SpringHertz() float32 {
	todo()
	return 0
}

// Get the spring damping ratio
func (d DistanceJoint) SpringDampingRatio() float32 {
	todo()
	return 0
}

// Get the current spring force.
func (d DistanceJoint) SpringForce() float32 {
	todo()
	return 0
}

// Enable/disable the distance joint motor
func (d DistanceJoint) EnableMotor(enableMotor bool) {
	todo()
}

// Is the distance joint motor enabled?
func (d DistanceJoint) IsMotorEnabled() bool {
	todo()
	return false
}

// Set the distance joint motor speed, usually in meters per second
func (d DistanceJoint) SetMotorSpeed(motorSpeed float32) {
	todo()
}

// Get the distance joint motor speed, usually in meters per second
func (d DistanceJoint) MotorSpeed() float32 {
	todo()
	return 0
}

// Get the distance joint current motor force, usually in newtons
func (d DistanceJoint) MotorForce() float32 {
	todo()
	return 0
}

// Set the distance joint maximum motor force, usually in newtons
func (d DistanceJoint) SetMaxMotorForce(force float32) {
	todo()
}

// Get the distance joint maximum motor force, usually in newtons
func (d DistanceJoint) MaxMotorForce() float32 {
	todo()
	return 0
}

func getDistanceJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func prepareDistanceJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartDistanceJoint(base *jointSim, context *stepContext) {
	todo()
}

func solveDistanceJoint(base *jointSim, context *stepContext, useBias bool) {
	todo()
}

func drawDistanceJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform) {
	todo()
}
