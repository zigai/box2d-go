// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import "github.com/zigai/box2d"

type Car struct {
	chassisID    box2d.Body
	rearWheelID  box2d.Body
	frontWheelID box2d.Body
	rearAxleID   box2d.WheelJoint
	frontAxleID  box2d.WheelJoint
	isSpawned    bool
}

func createWheelAxle(worldID *box2d.World, chassisID, wheelID box2d.Body, scale, hertz, dampingRatio, torque float32) box2d.WheelJoint {
	pivot := wheelID.Position()

	jointDef := box2d.DefaultWheelJointDef()
	jointDef.Base.BodyA = chassisID
	jointDef.Base.BodyB = wheelID
	jointDef.Base.LocalFrameA.Q = box2d.MakeRot(0.5 * box2d.Pi)
	jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
	jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
	jointDef.MotorSpeed = 0.0
	jointDef.MaxMotorTorque = torque
	jointDef.EnableMotor = true
	jointDef.Hertz = hertz
	jointDef.DampingRatio = dampingRatio
	jointDef.LowerTranslation = -0.25 * scale
	jointDef.UpperTranslation = 0.25 * scale
	jointDef.EnableLimit = true
	return worldID.CreateWheelJoint(&jointDef)
}

func (c *Car) Spawn(worldID *box2d.World, position box2d.Pos, scale, hertz, dampingRatio, torque float32, userData any) {
	vertices := []box2d.Vec2{
		{X: -1.5, Y: -0.5}, {X: 1.5, Y: -0.5}, {X: 1.5, Y: 0.0}, {X: 0.0, Y: 0.9}, {X: -1.15, Y: 0.9}, {X: -1.5, Y: 0.2},
	}

	for i := range vertices {
		vertices[i].X *= 0.85 * scale
		vertices[i].Y *= 0.85 * scale
	}

	hull := box2d.ComputeHull(vertices)
	chassis := box2d.MakePolygon(&hull, 0.15*scale)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0 / scale
	shapeDef.Material.Friction = 0.2

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.4 * scale}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = position.Add(box2d.Vec2{X: 0.0, Y: 1.0 * scale})
	c.chassisID = worldID.CreateBody(&bodyDef)
	c.chassisID.CreatePolygonShape(&shapeDef, &chassis)

	shapeDef.Density = 2.0 / scale
	shapeDef.Material.Friction = 1.5
	shapeDef.Material.RollingResistance = 0.1

	bodyDef.Position = position.Add(box2d.Vec2{X: -1.0 * scale, Y: 0.35 * scale})
	bodyDef.AllowFastRotation = true
	c.rearWheelID = worldID.CreateBody(&bodyDef)
	c.rearWheelID.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = position.Add(box2d.Vec2{X: 1.0 * scale, Y: 0.4 * scale})
	bodyDef.AllowFastRotation = true
	c.frontWheelID = worldID.CreateBody(&bodyDef)
	c.frontWheelID.CreateCircleShape(&shapeDef, &circle)

	c.rearAxleID = createWheelAxle(worldID, c.chassisID, c.rearWheelID, scale, hertz, dampingRatio, torque)
	c.frontAxleID = createWheelAxle(worldID, c.chassisID, c.frontWheelID, scale, hertz, dampingRatio, torque)
}

func (c *Car) Despawn() {
	c.rearAxleID.Destroy()
	c.frontAxleID.Destroy()
	c.rearWheelID.Destroy()
	c.frontWheelID.Destroy()
	c.chassisID.Destroy()
	c.isSpawned = false
}

func (c *Car) SetSpeed(speed float32) {
	c.rearAxleID.SetMotorSpeed(speed)
	c.frontAxleID.SetMotorSpeed(speed)
	c.rearAxleID.WakeBodies()
}

func (c *Car) SetTorque(torque float32) {
	c.rearAxleID.SetMaxMotorTorque(torque)
	c.frontAxleID.SetMaxMotorTorque(torque)
}

func (c *Car) SetHertz(hertz float32) {
	c.rearAxleID.SetSpringHertz(hertz)
	c.frontAxleID.SetSpringHertz(hertz)
}

func (c *Car) SetDampingRadio(dampingRatio float32) {
	c.rearAxleID.SetSpringDampingRatio(dampingRatio)
	c.frontAxleID.SetSpringDampingRatio(dampingRatio)
}

type Truck struct {
	chassisID    box2d.Body
	rearWheelID  box2d.Body
	frontWheelID box2d.Body
	rearAxleID   box2d.WheelJoint
	frontAxleID  box2d.WheelJoint
	isSpawned    bool
}

func (t *Truck) Spawn(worldID *box2d.World, position box2d.Pos, scale, hertz, dampingRatio, torque, density float32, userData any) {
	vertices := []box2d.Vec2{
		{X: -0.65, Y: -0.4}, {X: 1.5, Y: -0.4}, {X: 1.5, Y: 0.0}, {X: 0.0, Y: 0.9}, {X: -0.65, Y: 0.9},
	}

	for i := range vertices {
		vertices[i].X *= 0.85 * scale
		vertices[i].Y *= 0.85 * scale
	}

	hull := box2d.ComputeHull(vertices)
	chassis := box2d.MakePolygon(&hull, 0.15*scale)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = density
	shapeDef.Material.Friction = 0.2
	shapeDef.Material.CustomColor = uint32(box2d.ColorHotPink)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = position.Add(box2d.Vec2{X: 0.0, Y: 1.0 * scale})
	t.chassisID = worldID.CreateBody(&bodyDef)
	t.chassisID.CreatePolygonShape(&shapeDef, &chassis)

	box := box2d.MakeOffsetBox(1.25*scale, 0.1*scale, box2d.Vec2{X: -2.05 * scale, Y: -0.275 * scale}, box2d.RotIdentity)
	box.Radius = 0.1 * scale
	t.chassisID.CreatePolygonShape(&shapeDef, &box)

	box = box2d.MakeOffsetBox(0.05*scale, 0.35*scale, box2d.Vec2{X: -3.25 * scale, Y: 0.375 * scale}, box2d.RotIdentity)
	box.Radius = 0.1 * scale
	t.chassisID.CreatePolygonShape(&shapeDef, &box)

	shapeDef.Density = 2.0 * density
	shapeDef.Material.Friction = 2.5
	shapeDef.Material.CustomColor = uint32(box2d.ColorSilver)

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.4 * scale}
	bodyDef.Position = position.Add(box2d.Vec2{X: -2.75 * scale, Y: 0.3 * scale})
	t.rearWheelID = worldID.CreateBody(&bodyDef)
	t.rearWheelID.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = position.Add(box2d.Vec2{X: 0.8 * scale, Y: 0.3 * scale})
	t.frontWheelID = worldID.CreateBody(&bodyDef)
	t.frontWheelID.CreateCircleShape(&shapeDef, &circle)

	t.rearAxleID = createWheelAxle(worldID, t.chassisID, t.rearWheelID, scale, hertz, dampingRatio, torque)
	t.frontAxleID = createWheelAxle(worldID, t.chassisID, t.frontWheelID, scale, hertz, dampingRatio, torque)
}

func (t *Truck) Despawn() {
	t.rearAxleID.Destroy()
	t.frontAxleID.Destroy()
	t.rearWheelID.Destroy()
	t.frontWheelID.Destroy()
	t.chassisID.Destroy()
	t.isSpawned = false
}

func (t *Truck) SetSpeed(speed float32) {
	t.rearAxleID.SetMotorSpeed(speed)
	t.frontAxleID.SetMotorSpeed(speed)
	t.rearAxleID.WakeBodies()
}

func (t *Truck) SetTorque(torque float32) {
	t.rearAxleID.SetMaxMotorTorque(torque)
	t.frontAxleID.SetMaxMotorTorque(torque)
}

func (t *Truck) SetHertz(hertz float32) {
	t.rearAxleID.SetSpringHertz(hertz)
	t.frontAxleID.SetSpringHertz(hertz)
}

func (t *Truck) SetDampingRadio(dampingRatio float32) {
	t.rearAxleID.SetSpringDampingRatio(dampingRatio)
	t.frontAxleID.SetSpringDampingRatio(dampingRatio)
}
