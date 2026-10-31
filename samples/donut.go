// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"

	"github.com/zigai/box2d-go/box2d"
)

const donutSides = 7

type Donut struct {
	bodyIDs   [donutSides]box2d.Body
	jointIDs  [donutSides]box2d.Joint
	isSpawned bool
}

func (d *Donut) Create(worldID *box2d.World, position box2d.Pos, scale float32, groupIndex int, enableSensorEvents bool, userData any) {
	radius := 1.0 * scale
	deltaAngle := 2.0 * box2d.Pi / donutSides
	length := 2.0 * box2d.Pi * radius / donutSides

	capsule := box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.5 * length}, Center2: box2d.Vec2{X: 0.0, Y: 0.5 * length}, Radius: 0.25 * scale}

	center := position

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.UserData = userData

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = enableSensorEvents
	shapeDef.Filter.GroupIndex = -groupIndex
	shapeDef.Material.Friction = 0.3

	angle := float32(0.0)
	for i := range donutSides {
		cosAngle := float32(math.Cos(float64(angle)))
		sinAngle := float32(math.Sin(float64(angle)))
		bodyDef.Position = box2d.Pos{X: posScalar(radius*cosAngle) + center.X, Y: posScalar(radius*sinAngle) + center.Y}
		bodyDef.Rotation = box2d.MakeRot(angle)

		d.bodyIDs[i] = worldID.CreateBody(&bodyDef)
		d.bodyIDs[i].CreateCapsuleShape(&shapeDef, &capsule)

		angle += deltaAngle
	}

	weldDef := box2d.DefaultWeldJointDef()
	weldDef.AngularHertz = 5.0
	weldDef.AngularDampingRatio = 0.0
	weldDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.0, Y: 0.5 * length}
	weldDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: -0.5 * length}
	weldDef.Base.DrawScale = 0.5 * scale

	prevBodyID := d.bodyIDs[donutSides-1]
	for i := range donutSides {
		weldDef.Base.BodyA = prevBodyID
		weldDef.Base.BodyB = d.bodyIDs[i]
		qA := prevBodyID.Rotation()
		qB := d.bodyIDs[i].Rotation()
		weldDef.Base.LocalFrameA.Q = qA.InvMul(qB)
		d.jointIDs[i] = worldID.CreateWeldJoint(&weldDef).Joint
		prevBodyID = weldDef.Base.BodyB
	}

	d.isSpawned = true
}

func (d *Donut) Destroy() {
	for i := range donutSides {
		d.bodyIDs[i].Destroy()
		d.bodyIDs[i] = box2d.Body{}
		d.jointIDs[i] = box2d.Joint{}
	}

	d.isSpawned = false
}
