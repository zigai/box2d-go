// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import "github.com/zigai/box2d-go/box2d"

type Car struct {
	chassisID    box2d.Body
	rearWheelID  box2d.Body
	frontWheelID box2d.Body
	rearAxleID   box2d.WheelJoint
	frontAxleID  box2d.WheelJoint
	isSpawned    bool
}

type Truck struct {
	chassisID    box2d.Body
	rearWheelID  box2d.Body
	frontWheelID box2d.Body
	rearAxleID   box2d.WheelJoint
	frontAxleID  box2d.WheelJoint
	isSpawned    bool
}

func createWheelAxle(
	worldID *box2d.World,
	chassisID box2d.Body, wheelID box2d.Body,
	scale float32, hertz float32, dampingRatio float32, torque float32,
) box2d.WheelJoint {
	const (
		axleAngle        = 0.5 * box2d.Pi
		suspensionTravel = 0.25
	)

	pivot := wheelID.Position()

	jointDef := box2d.DefaultWheelJointDef()
	jointDef.Base.BodyA = chassisID
	jointDef.Base.BodyB = wheelID
	jointDef.Base.LocalFrameA.Q = box2d.MakeRot(axleAngle)
	jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
	jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
	jointDef.MotorSpeed = 0.0
	jointDef.MaxMotorTorque = torque
	jointDef.EnableMotor = true
	jointDef.Hertz = hertz
	jointDef.DampingRatio = dampingRatio
	jointDef.LowerTranslation = -suspensionTravel * scale
	jointDef.UpperTranslation = suspensionTravel * scale
	jointDef.EnableLimit = true

	return worldID.CreateWheelJoint(&jointDef)
}

func (c *Car) Spawn(
	worldID *box2d.World,
	position box2d.Pos,
	scale float32, hertz float32, dampingRatio float32, torque float32,
) {
	const (
		rearX             = -1.5
		frontX            = 1.5
		bottomY           = -0.5
		roofY             = 0.9
		roofRearX         = -1.15
		trunkY            = 0.2
		hullScale         = 0.85
		cornerRadius      = 0.15
		chassisFriction   = 0.2
		wheelRadius       = 0.4
		wheelDensity      = 2.0
		wheelFriction     = 1.5
		rollingResistance = 0.1
		rearWheelX        = -1.0
		rearWheelY        = 0.35
		frontWheelX       = 1.0
		frontWheelY       = 0.4
	)

	vertices := []box2d.Vec2{
		{
			X: rearX,
			Y: bottomY,
		},
		{X: frontX, Y: bottomY},
		{X: frontX, Y: 0.0},
		{X: 0.0, Y: roofY},
		{X: roofRearX, Y: roofY},
		{X: rearX, Y: trunkY},
	}

	for i := range vertices {
		vertices[i].X *= hullScale * scale
		vertices[i].Y *= hullScale * scale
	}

	hull := box2d.ComputeHull(vertices)
	chassis := box2d.MakePolygon(&hull, cornerRadius*scale)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0 / scale
	shapeDef.Material.Friction = chassisFriction

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: wheelRadius * scale}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = position.Add(box2d.Vec2{X: 0.0, Y: 1.0 * scale})
	c.chassisID = worldID.CreateBody(&bodyDef)
	c.chassisID.CreatePolygonShape(&shapeDef, &chassis)

	shapeDef.Density = wheelDensity / scale
	shapeDef.Material.Friction = wheelFriction
	shapeDef.Material.RollingResistance = rollingResistance

	bodyDef.Position = position.Add(box2d.Vec2{X: rearWheelX * scale, Y: rearWheelY * scale})
	bodyDef.AllowFastRotation = true
	c.rearWheelID = worldID.CreateBody(&bodyDef)
	c.rearWheelID.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = position.Add(box2d.Vec2{X: frontWheelX * scale, Y: frontWheelY * scale})
	bodyDef.AllowFastRotation = true
	c.frontWheelID = worldID.CreateBody(&bodyDef)
	c.frontWheelID.CreateCircleShape(&shapeDef, &circle)

	c.rearAxleID = createWheelAxle(
		worldID,
		c.chassisID,
		c.rearWheelID,
		scale,
		hertz,
		dampingRatio,
		torque,
	)
	c.frontAxleID = createWheelAxle(
		worldID,
		c.chassisID,
		c.frontWheelID,
		scale,
		hertz,
		dampingRatio,
		torque,
	)
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

func (t *Truck) Spawn(
	worldID *box2d.World,
	position box2d.Pos,
	scale float32, hertz float32, dampingRatio float32, torque float32, density float32,
) {
	const (
		cabRearX           = -0.65
		cabFrontX          = 1.5
		cabBottomY         = -0.4
		cabRoofY           = 0.9
		hullScale          = 0.85
		cabCornerRadius    = 0.15
		chassisFriction    = 0.2
		bedHalfWidth       = 1.25
		bedHalfHeight      = 0.1
		bedX               = -2.05
		bedY               = -0.275
		cornerRadius       = 0.1
		tailgateHalfWidth  = 0.05
		tailgateHalfHeight = 0.35
		tailgateX          = -3.25
		tailgateY          = 0.375
		wheelDensityScale  = 2.0
		wheelFriction      = 2.5
		wheelRadius        = 0.4
		rearWheelX         = -2.75
		frontWheelX        = 0.8
		wheelY             = 0.3
	)

	vertices := []box2d.Vec2{
		{
			X: cabRearX,
			Y: cabBottomY,
		},
		{X: cabFrontX, Y: cabBottomY},
		{X: cabFrontX, Y: 0.0},
		{X: 0.0, Y: cabRoofY},
		{X: cabRearX, Y: cabRoofY},
	}

	for i := range vertices {
		vertices[i].X *= hullScale * scale
		vertices[i].Y *= hullScale * scale
	}

	hull := box2d.ComputeHull(vertices)
	chassis := box2d.MakePolygon(&hull, cabCornerRadius*scale)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = density
	shapeDef.Material.Friction = chassisFriction
	shapeDef.Material.CustomColor = uint32(box2d.ColorHotPink)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = position.Add(box2d.Vec2{X: 0.0, Y: 1.0 * scale})
	t.chassisID = worldID.CreateBody(&bodyDef)
	t.chassisID.CreatePolygonShape(&shapeDef, &chassis)

	box := box2d.MakeOffsetBox(
		bedHalfWidth*scale,
		bedHalfHeight*scale,
		box2d.Vec2{X: bedX * scale, Y: bedY * scale},
		box2d.RotIdentity,
	)
	box.Radius = cornerRadius * scale
	t.chassisID.CreatePolygonShape(&shapeDef, &box)

	box = box2d.MakeOffsetBox(
		tailgateHalfWidth*scale,
		tailgateHalfHeight*scale,
		box2d.Vec2{X: tailgateX * scale, Y: tailgateY * scale},
		box2d.RotIdentity,
	)
	box.Radius = cornerRadius * scale
	t.chassisID.CreatePolygonShape(&shapeDef, &box)

	shapeDef.Density = wheelDensityScale * density
	shapeDef.Material.Friction = wheelFriction
	shapeDef.Material.CustomColor = uint32(box2d.ColorSilver)

	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: wheelRadius * scale}
	bodyDef.Position = position.Add(box2d.Vec2{X: rearWheelX * scale, Y: wheelY * scale})
	t.rearWheelID = worldID.CreateBody(&bodyDef)
	t.rearWheelID.CreateCircleShape(&shapeDef, &circle)

	bodyDef.Position = position.Add(box2d.Vec2{X: frontWheelX * scale, Y: wheelY * scale})
	t.frontWheelID = worldID.CreateBody(&bodyDef)
	t.frontWheelID.CreateCircleShape(&shapeDef, &circle)

	t.rearAxleID = createWheelAxle(
		worldID,
		t.chassisID,
		t.rearWheelID,
		scale,
		hertz,
		dampingRatio,
		torque,
	)
	t.frontAxleID = createWheelAxle(
		worldID,
		t.chassisID,
		t.frontWheelID,
		scale,
		hertz,
		dampingRatio,
		torque,
	)
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
