// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/zigai/box2d-go/box2d"
)

const (
	fallingHingePrecision        = int(unsafe.Sizeof(box2d.Pos{}))/8 - 1
	ExpectedSleepStep            = 295*(1-fallingHingePrecision) + 256*fallingHingePrecision
	ExpectedHash          uint32 = uint32(0x36564C94*(1-fallingHingePrecision) + 0x29D0B92D*fallingHingePrecision)
)

type FallingHingeData struct {
	BodyIDs   []box2d.Body
	BodyCount int
	StepCount int
	SleepStep int
	Hash      uint32
}

func CreateFallingHinges(worldID *box2d.World) FallingHingeData {
	{
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Position = box2d.Pos{X: 0, Y: -1}
		groundID := worldID.CreateBody(&bodyDef)
		box := box2d.MakeBox(40, 1)
		shapeDef := box2d.DefaultShapeDef()
		groundID.CreatePolygonShape(&shapeDef, &box)
	}

	columnCount := 4
	rowCount := 20
	bodyCount := rowCount * columnCount
	bodyIDs := make([]box2d.Body, bodyCount)
	h := float32(0.25)
	box := box2d.MakeSquare(h)
	shapeDef := box2d.DefaultShapeDef()
	// shapeDef.Material.Friction = 0.3f;
	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.EnableLimit = true
	jointDef.LowerAngle = -float32(0.1) * box2d.Pi
	jointDef.UpperAngle = float32(0.2) * box2d.Pi
	jointDef.EnableSpring = true
	jointDef.Hertz = 1
	jointDef.DampingRatio = 1
	jointDef.EnableMotor = true
	jointDef.MaxMotorTorque = 0.25
	jointDef.Base.LocalFrameA.P = box2d.Vec2{X: -h, Y: h}
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: -h, Y: -h}
	jointDef.Base.ConstraintHertz = 60
	jointDef.Base.ConstraintDampingRatio = 0
	jointDef.Base.DrawScale = 0.5
	bodyIndex := 0
	offset := float32(0.4) * h
	dx := float32(10) * h

	xBase := -float32(0.5) * dx * (float32(columnCount) - float32(1))
	for j := range columnCount {
		x := xBase + float32(float32(j)*dx)
		prevBodyID := box2d.Body{}

		for i := range rowCount {
			bodyDef := box2d.DefaultBodyDef()
			bodyDef.Type = box2d.DynamicBody
			bodyDef.Position = box2d.Vec2{X: x + float32(offset*float32(i)), Y: h + float32(float32(2)*h*float32(i))}.Pos()
			// this tests the deterministic cosine and sine functions
			angle := float32(0.1)
			if i&1 == 0 {
				angle = -float32(0.1)
			}

			bodyDef.Rotation = box2d.MakeRot(angle)

			bodyID := worldID.CreateBody(&bodyDef)
			if i&1 == 0 {
				prevBodyID = bodyID
			} else {
				jointDef.Base.BodyA = prevBodyID
				jointDef.Base.BodyB = bodyID
				worldID.CreateRevoluteJoint(&jointDef)

				prevBodyID = box2d.Body{}
			}

			bodyID.CreatePolygonShape(&shapeDef, &box)

			if bodyIndex >= bodyCount {
				panic(&box2d.AssertError{Condition: "bodyIndex < bodyCount"})
			}

			bodyIDs[bodyIndex] = bodyID
			bodyIndex += 1
		}
	}

	if bodyIndex != bodyCount {
		panic(&box2d.AssertError{Condition: "bodyIndex == bodyCount"})
	}

	data := FallingHingeData{BodyIDs: bodyIDs, BodyCount: bodyCount, StepCount: 0, SleepStep: -1, Hash: 0}

	return data
}

func UpdateFallingHinges(worldID *box2d.World, data *FallingHingeData) bool {
	if data.Hash == 0 {
		moveCount := 0
		for range worldID.BodyMoveEvents() {
			moveCount++
		}

		if moveCount == 0 {
			if worldID.AwakeBodyCount() != 0 {
				panic(&box2d.AssertError{Condition: "b2World_GetAwakeBodyCount( worldId ) == 0"})
			}

			data.Hash = box2d.HashInit
			hashFallingHingeTransforms(data)

			data.SleepStep = data.StepCount
		}
	}

	data.StepCount += 1

	return data.Hash != 0
}

func DestroyFallingHinges(data *FallingHingeData) { data.BodyIDs = nil }

func hashFallingHingeTransforms(data *FallingHingeData) {
	for i := range data.BodyCount {
		transform := data.BodyIDs[i].Transform()

		var bytes [int(unsafe.Sizeof(box2d.WorldTransform{}))]byte
		if fallingHingePrecision == 1 {
			binary.LittleEndian.PutUint64(bytes[0:8], math.Float64bits(float64(transform.P.X)))
			binary.LittleEndian.PutUint64(bytes[8:16], math.Float64bits(float64(transform.P.Y)))
		} else {
			binary.LittleEndian.PutUint32(bytes[0:4], math.Float32bits(float32(transform.P.X)))
			binary.LittleEndian.PutUint32(bytes[4:8], math.Float32bits(float32(transform.P.Y)))
		}

		offset := int(unsafe.Sizeof(box2d.Pos{}))
		binary.LittleEndian.PutUint32(bytes[offset:offset+4], math.Float32bits(transform.Q.C))
		binary.LittleEndian.PutUint32(bytes[offset+4:offset+8], math.Float32bits(transform.Q.S))
		data.Hash = box2d.Hash(data.Hash, bytes[:])
	}
}
