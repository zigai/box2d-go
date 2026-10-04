// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

package main

import "github.com/zigai/box2d"

type Doohickey struct {
	wheelID1  box2d.Body
	wheelID2  box2d.Body
	barID1    box2d.Body
	barID2    box2d.Body
	axleID1   box2d.Joint
	axleID2   box2d.Joint
	sliderID  box2d.Joint
	isSpawned bool
}

func (d *Doohickey) Spawn(worldID *box2d.World, position box2d.Pos, scale float32) {
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Material.RollingResistance = 0.1

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 1.0 * scale}
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -3.5 * scale, Y: 0.0}, Center2: box2d.Vec2{X: 3.5 * scale, Y: 0.0}, Radius: 0.15 * scale}

	bodyDef.Position = position.Add(box2d.Vec2{X: -5.0, Y: 3.0}.Scale(scale))
	d.wheelID1 = worldID.CreateBody(&bodyDef)
	d.wheelID1.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = position.Add(box2d.Vec2{X: 5.0, Y: 3.0}.Scale(scale))
	d.wheelID2 = worldID.CreateBody(&bodyDef)
	d.wheelID2.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = position.Add(box2d.Vec2{X: -1.5, Y: 3.0}.Scale(scale))
	d.barID1 = worldID.CreateBody(&bodyDef)
	d.barID1.CreateCapsuleShape(&shapeDef, &capsule)

	bodyDef.Position = position.Add(box2d.Vec2{X: 1.5, Y: 3.0}.Scale(scale))
	d.barID2 = worldID.CreateBody(&bodyDef)
	d.barID2.CreateCapsuleShape(&shapeDef, &capsule)

	revoluteDef := box2d.DefaultRevoluteJointDef()

	revoluteDef.Base.BodyA = d.wheelID1
	revoluteDef.Base.BodyB = d.barID1
	revoluteDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.0}
	revoluteDef.Base.LocalFrameB.P = box2d.Vec2{X: -3.5 * scale, Y: 0.0}
	revoluteDef.EnableMotor = true
	revoluteDef.MaxMotorTorque = 2.0 * scale
	worldID.CreateRevoluteJoint(&revoluteDef)

	revoluteDef.Base.BodyA = d.wheelID2
	revoluteDef.Base.BodyB = d.barID2
	revoluteDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.0}
	revoluteDef.Base.LocalFrameB.P = box2d.Vec2{X: 3.5 * scale, Y: 0.0}
	revoluteDef.EnableMotor = true
	revoluteDef.MaxMotorTorque = 2.0 * scale
	worldID.CreateRevoluteJoint(&revoluteDef)

	prismaticDef := box2d.DefaultPrismaticJointDef()
	prismaticDef.Base.BodyA = d.barID1
	prismaticDef.Base.BodyB = d.barID2
	prismaticDef.Base.LocalFrameA.P = box2d.Vec2{X: 2.0 * scale, Y: 0.0}
	prismaticDef.Base.LocalFrameB.P = box2d.Vec2{X: -2.0 * scale, Y: 0.0}
	prismaticDef.LowerTranslation = -2.0 * scale
	prismaticDef.UpperTranslation = 2.0 * scale
	prismaticDef.EnableLimit = true
	prismaticDef.EnableMotor = true
	prismaticDef.MaxMotorForce = 2.0 * scale
	prismaticDef.EnableSpring = true
	prismaticDef.Hertz = 1.0
	prismaticDef.DampingRatio = 0.5
	worldID.CreatePrismaticJoint(&prismaticDef)
}

func (d *Doohickey) Despawn() {
	d.axleID1.Destroy()
	d.axleID2.Destroy()
	d.sliderID.Destroy()

	d.wheelID1.Destroy()
	d.wheelID2.Destroy()
	d.barID1.Destroy()
	d.barID2.Destroy()

	d.isSpawned = false
}
