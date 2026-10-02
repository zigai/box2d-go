// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type MoverJoint struct {
	Joint
}

// A mover joint is used to move a dynamic character mover through velocity commands.
// The x and y directions are handled separately. Does not affect rotation.
type MoverJointDef struct {
	// Base joint definition
	Base JointDef
	// The desired linear velocity
	LinearVelocity Vec2
	// The maximum motor force in newtons.
	MaxVelocityForce Vec2
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Set the desired relative linear velocity in meters per second
func (m MoverJoint) SetLinearVelocity(velocity Vec2) {
	todo()
}

// Get the desired relative linear velocity in meters per second
func (m MoverJoint) LinearVelocity() Vec2 {
	todo()
	return Vec2{}
}

// Set the motor joint maximum force, usually in newtons
func (m MoverJoint) SetMaxVelocityForce(maxForce Vec2) {
	todo()
}

// Get the motor joint maximum force, usually in newtons
func (m MoverJoint) MaxVelocityForce() Vec2 {
	todo()
	return Vec2{}
}

func getMoverJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func prepareMoverJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartMoverJoint(base *jointSim, context *stepContext) {
	todo()
}

func solveMoverJoint(base *jointSim, context *stepContext) {
	todo()
}
