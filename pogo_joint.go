// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type PogoJoint struct {
	Joint
}

// Pogo joint definition
type PogoJointDef struct {
	// Base joint definition
	Base JointDef
	// Normal vector at the contact
	Normal Vec2
	// The spring stiffness Hertz, cycles per second
	Hertz float32
	// The spring damping ratio, non-dimensional
	DampingRatio float32
	// Spring length.
	RestLength float32
	// Maximum tension force. Controls how sticky the mover is to ground.
	MaxTensionForce float32
	// Maximum compression force. How hard the pogo pushes the mover up from the ground.
	MaxCompressionForce float32
	// The initial pogo internal impulse. The pogo is typically recreated every frame. This
	// gives a way to move the internal state forward for smoother spring behavior.
	Impulse float32
	// The initial pogo internal velocity. The pogo is typically recreated every frame. This
	// gives a way to move the internal state forward for smoother spring behavior.
	Velocity float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

var pogoAxis = Vec2{X: 0.0, Y: 1.0}

// Set the rest length of a pogo joint
func (p PogoJoint) SetRestLength(length float32) {
	todo()
}

// Get the rest length of a pogo joint
func (p PogoJoint) RestLength() float32 {
	todo()
	return 0
}

// Get the spring Hertz
func (p PogoJoint) SpringHertz() float32 {
	todo()
	return 0
}

// Set the spring stiffness in Hertz
func (p PogoJoint) SetSpringHertz(hertz float32) {
	todo()
}

// Set the spring damping ratio, non-dimensional
func (p PogoJoint) SetSpringDampingRatio(dampingRatio float32) {
	todo()
}

// Get the spring damping ratio
func (p PogoJoint) SpringDampingRatio() float32 {
	todo()
	return 0
}

// Get the current length of the pogo.
func (p PogoJoint) Length() float32 {
	todo()
	return 0
}

// Get the internal pogo impulse.
func (p PogoJoint) Impulse() float32 {
	todo()
	return 0
}

// Get the internal pogo velocity.
func (p PogoJoint) Velocity() float32 {
	todo()
	return 0
}

func getPogoJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func preparePogoJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartPogoJoint(base *jointSim, context *stepContext) {
	todo()
}

func solvePogoJoint(base *jointSim, context *stepContext, useBias bool) {
	todo()
}
