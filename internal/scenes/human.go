// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import "github.com/zigai/box2d"

type BoneID int

const (
	BoneHip BoneID = iota
	BoneTorso
	BoneHead
	BoneUpperLeftLeg
	BoneLowerLeftLeg
	BoneUpperRightLeg
	BoneLowerRightLeg
	BoneUpperLeftArm
	BoneLowerLeftArm
	BoneUpperRightArm
	BoneLowerRightArm
	BoneCount
)

type Bone struct {
	BodyID        box2d.Body
	JointID       box2d.Joint
	FrictionScale float32
	MaxTorque     float32
	ParentIndex   int
}

type Human struct {
	Bones          [BoneCount]Bone
	FrictionTorque float32
	OriginalScale  float32
	Scale          float32
	IsSpawned      bool
}

func CreateHuman(human *Human, worldID box2d.World, position box2d.Pos, scale, frictionTorque, hertz, dampingRatio float32, groupIndex int, userData any, colorize bool) {
	todo()
}

func DestroyHuman(human *Human) {
	todo()
}

func (human *Human) SetVelocity(velocity box2d.Vec2) {
	todo()
}

func (human *Human) ApplyRandomAngularImpulse(magnitude float32) {
	todo()
}

func (human *Human) SetJointFrictionTorque(torque float32) {
	todo()
}

func (human *Human) SetJointSpringHertz(hertz float32) {
	todo()
}

func (human *Human) SetJointDampingRatio(dampingRatio float32) {
	todo()
}

func (human *Human) EnableSensorEvents(enable bool) {
	todo()
}

func (human *Human) SetScale(scale float32) {
	todo()
}
