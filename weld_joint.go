// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type WeldJoint struct {
	Joint
}

// Weld joint definition
// Connects two bodies together rigidly. This constraint provides springs to mimic
// soft-body simulation.
type WeldJointDef struct {
	// Base joint definition
	Base JointDef
	// Linear stiffness expressed as Hertz (cycles per second). Use zero for maximum stiffness.
	LinearHertz float32
	// Angular stiffness as Hertz (cycles per second). Use zero for maximum stiffness.
	AngularHertz float32
	// Linear damping ratio, non-dimensional. Use 1 for critical damping.
	LinearDampingRatio float32
	// Linear damping ratio, non-dimensional. Use 1 for critical damping.
	AngularDampingRatio float32
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Set the weld joint linear stiffness in Hertz. 0 is rigid.
func (w WeldJoint) SetLinearHertz(hertz float32) {
	todo()
}

// Get the weld joint linear stiffness in Hertz
func (w WeldJoint) LinearHertz() float32 {
	todo()
	return 0
}

// Set the weld joint linear damping ratio (non-dimensional)
func (w WeldJoint) SetLinearDampingRatio(dampingRatio float32) {
	todo()
}

// Get the weld joint linear damping ratio (non-dimensional)
func (w WeldJoint) LinearDampingRatio() float32 {
	todo()
	return 0
}

// Set the weld joint angular stiffness in Hertz. 0 is rigid.
func (w WeldJoint) SetAngularHertz(hertz float32) {
	todo()
}

// Get the weld joint angular stiffness in Hertz
func (w WeldJoint) AngularHertz() float32 {
	todo()
	return 0
}

// Set weld joint angular damping ratio, non-dimensional
func (w WeldJoint) SetAngularDampingRatio(dampingRatio float32) {
	todo()
}

// Get the weld joint angular damping ratio, non-dimensional
func (w WeldJoint) AngularDampingRatio() float32 {
	todo()
	return 0
}

func getWeldJointForce(world *world, base *jointSim) Vec2 {
	todo()
	return Vec2{}
}

func getWeldJointTorque(world *world, base *jointSim) float32 {
	todo()
	return 0
}

func prepareWeldJoint(base *jointSim, context *stepContext) {
	todo()
}

func warmStartWeldJoint(base *jointSim, context *stepContext) {
	todo()
}

func solveWeldJoint(base *jointSim, context *stepContext, useBias bool) {
	todo()
}

func drawWeldJoint(draw *DebugDraw, base *jointSim, transformA WorldTransform, transformB WorldTransform, drawScale float32) {
	todo()
}
