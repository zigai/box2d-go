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

func CreateHuman(human *Human, worldID *box2d.World, position box2d.Pos, scale, frictionTorque, hertz, dampingRatio float32, groupIndex int, userData any, colorize bool) {
	if !(human.IsSpawned == false) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == false"})
	}
	for i := 0; i < int(BoneCount); i++ {
		human.Bones[i].BodyID = box2d.Body{}
		human.Bones[i].JointID = box2d.Joint{}
		human.Bones[i].FrictionScale = float32(1.0)
		human.Bones[i].ParentIndex = -1
	}
	human.OriginalScale = scale
	human.Scale = scale
	human.FrictionTorque = frictionTorque
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.SleepThreshold = float32(0.1)
	bodyDef.UserData = userData
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.Material.Friction = float32(0.2)
	shapeDef.Filter.GroupIndex = -groupIndex
	shapeDef.Filter.CategoryBits = 2
	shapeDef.Filter.MaskBits = (1 | 2)
	var footShapeDef box2d.ShapeDef = shapeDef
	footShapeDef.Material.Friction = float32(0.05)
	footShapeDef.Filter.CategoryBits = 2
	footShapeDef.Filter.MaskBits = 1
	if colorize {
		footShapeDef.Material.CustomColor = uint32(box2d.ColorSaddleBrown)
	}
	var s float32 = scale
	var maxTorque float32 = frictionTorque * s
	var enableMotor bool = true
	var enableLimit bool = true
	var drawSize float32 = float32(0.05)
	var shirtColor box2d.HexColor = box2d.ColorMediumTurquoise
	var pantColor box2d.HexColor = box2d.ColorDodgerBlue
	var skinColors [4]box2d.HexColor = [4]box2d.HexColor{box2d.ColorNavajoWhite, box2d.ColorLightYellow, box2d.ColorPeru, box2d.ColorTan}
	var skinColor box2d.HexColor = skinColors[groupIndex%4]
	{
		var bone *Bone = &human.Bones[int(BoneHip)]
		bone.ParentIndex = -1
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.95) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "hip"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		if colorize {
			shapeDef.Material.CustomColor = uint32(pantColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.02) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.02) * s}, Radius: float32(0.095) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
	}
	{
		var bone *Bone = &human.Bones[int(BoneTorso)]
		bone.ParentIndex = int(BoneHip)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.2) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "torso"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.5)
		bodyDef.Type = box2d.DynamicBody
		if colorize {
			shapeDef.Material.CustomColor = uint32(shirtColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.135) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.135) * s}, Radius: float32(0.09) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.0) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.25) * box2d.Pi
		jointDef.UpperAngle = float32(0.0)
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneHead)]
		bone.ParentIndex = int(BoneTorso)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0) * s, Y: float32(1.475) * s})
		bodyDef.LinearDamping = float32(0.1)
		bodyDef.Name = "head"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.25)
		if colorize {
			shapeDef.Material.CustomColor = uint32(skinColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.038) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.039) * s}, Radius: float32(0.075) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.4) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.3) * box2d.Pi
		jointDef.UpperAngle = float32(0.1) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperLeftLeg)]
		bone.ParentIndex = int(BoneHip)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.775) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "upper_left_leg"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(1.0)
		if colorize {
			shapeDef.Material.CustomColor = uint32(pantColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.125) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.06) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.9) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.05) * box2d.Pi
		jointDef.UpperAngle = float32(0.4) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	var points [4]box2d.Vec2 = [4]box2d.Vec2{box2d.Vec2{X: -float32(0.03) * s, Y: -float32(0.185) * s}, box2d.Vec2{X: float32(0.11) * s, Y: -float32(0.185) * s}, box2d.Vec2{X: float32(0.11) * s, Y: -float32(0.16) * s}, box2d.Vec2{X: -float32(0.03) * s, Y: -float32(0.14) * s}}
	var footHull box2d.Hull = box2d.ComputeHull(points[:])
	var footPolygon box2d.Polygon = box2d.MakePolygon(&footHull, float32(0.015)*s)
	{
		var bone *Bone = &human.Bones[int(BoneLowerLeftLeg)]
		bone.ParentIndex = int(BoneUpperLeftLeg)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.475) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "lower_left_leg"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.5)
		if colorize {
			shapeDef.Material.CustomColor = uint32(pantColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.155) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.045) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		bone.BodyID.CreatePolygonShape(&footShapeDef, &footPolygon)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.625) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.5) * box2d.Pi
		jointDef.UpperAngle = -float32(0.02) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperRightLeg)]
		bone.ParentIndex = int(BoneHip)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.775) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "upper_right_leg"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(1.0)
		if colorize {
			shapeDef.Material.CustomColor = uint32(pantColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.125) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.06) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.9) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.05) * box2d.Pi
		jointDef.UpperAngle = float32(0.4) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerRightLeg)]
		bone.ParentIndex = int(BoneUpperRightLeg)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.475) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "lower_right_leg"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.5)
		if colorize {
			shapeDef.Material.CustomColor = uint32(pantColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.155) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.045) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		bone.BodyID.CreatePolygonShape(&footShapeDef, &footPolygon)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.625) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.5) * box2d.Pi
		jointDef.UpperAngle = -float32(0.02) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperLeftArm)]
		bone.ParentIndex = int(BoneTorso)
		bone.FrictionScale = float32(0.5)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.225) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "upper_left_arm"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		if colorize {
			shapeDef.Material.CustomColor = uint32(shirtColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.125) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.035) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.35) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.1) * box2d.Pi
		jointDef.UpperAngle = float32(0.8) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerLeftArm)]
		bone.ParentIndex = int(BoneUpperLeftArm)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.975) * s})
		bodyDef.LinearDamping = float32(0.1)
		bodyDef.Name = "lower_left_arm"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.1)
		if colorize {
			shapeDef.Material.CustomColor = uint32(skinColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.125) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.03) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.1) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameA.Q = box2d.MakeRot(float32(0.25) * box2d.Pi)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.2) * box2d.Pi
		jointDef.UpperAngle = float32(0.3) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperRightArm)]
		bone.ParentIndex = int(BoneTorso)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.225) * s})
		bodyDef.LinearDamping = float32(0.0)
		bodyDef.Name = "upper_right_arm"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.5)
		if colorize {
			shapeDef.Material.CustomColor = uint32(shirtColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.125) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.035) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.35) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.1) * box2d.Pi
		jointDef.UpperAngle = float32(0.8) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerRightArm)]
		bone.ParentIndex = int(BoneUpperRightArm)
		bodyDef.Position = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.975) * s})
		bodyDef.LinearDamping = float32(0.1)
		bodyDef.Name = "lower_right_arm"
		bone.BodyID = worldID.CreateBody(&bodyDef)
		bone.FrictionScale = float32(0.1)
		if colorize {
			shapeDef.Material.CustomColor = uint32(skinColor)
		}
		var capsule box2d.Capsule = box2d.Capsule{Center1: box2d.Vec2{X: float32(0.0), Y: -float32(0.125) * s}, Center2: box2d.Vec2{X: float32(0.0), Y: float32(0.125) * s}, Radius: float32(0.03) * s}
		bone.BodyID.CreateCapsuleShape(&shapeDef, &capsule)
		var pivot box2d.Pos = position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.1) * s})
		var jointDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
		jointDef.Base.BodyB = bone.BodyID
		jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
		jointDef.Base.LocalFrameA.Q = box2d.MakeRot(float32(0.25) * box2d.Pi)
		jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
		jointDef.EnableLimit = enableLimit
		jointDef.LowerAngle = -float32(0.2) * box2d.Pi
		jointDef.UpperAngle = float32(0.3) * box2d.Pi
		jointDef.EnableMotor = enableMotor
		jointDef.MaxMotorTorque = bone.FrictionScale * maxTorque
		jointDef.EnableSpring = hertz > float32(0.0)
		jointDef.Hertz = hertz
		jointDef.DampingRatio = dampingRatio
		jointDef.Base.DrawScale = drawSize
		bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
	}
	human.IsSpawned = true
}

func DestroyHuman(human *Human) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	for i := 0; i < int(BoneCount); i++ {
		if human.Bones[i].JointID == (box2d.Joint{}) {
			continue
		}
		human.Bones[i].JointID.Destroy()
		human.Bones[i].JointID = box2d.Joint{}
	}
	for i := 0; i < int(BoneCount); i++ {
		if human.Bones[i].BodyID == (box2d.Body{}) {
			continue
		}
		human.Bones[i].BodyID.Destroy()
		human.Bones[i].BodyID = box2d.Body{}
	}
	human.IsSpawned = false
}

func (human *Human) SetVelocity(velocity box2d.Vec2) {
	for i := 0; i < int(BoneCount); i++ {
		var bodyID box2d.Body = human.Bones[i].BodyID
		if bodyID == (box2d.Body{}) {
			continue
		}
		bodyID.SetLinearVelocity(velocity)
	}
}

func (human *Human) ApplyRandomAngularImpulse(magnitude float32) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	var impulse float32 = RandomFloatRange(-magnitude, magnitude)
	human.Bones[int(BoneTorso)].BodyID.ApplyAngularImpulse(impulse, true)
}

func (human *Human) SetJointFrictionTorque(torque float32) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	if torque == float32(0.0) {
		for i := 1; i < int(BoneCount); i++ {
			box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetMotorEnabled(false)
		}
	} else {
		for i := 1; i < int(BoneCount); i++ {
			box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetMotorEnabled(true)
			var scale float32 = human.Scale * human.Bones[i].FrictionScale
			box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetMaxMotorTorque(scale * torque)
		}
	}
}

func (human *Human) SetJointSpringHertz(hertz float32) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	if hertz == float32(0.0) {
		for i := 1; i < int(BoneCount); i++ {
			box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetSpringEnabled(false)
		}
	} else {
		for i := 1; i < int(BoneCount); i++ {
			box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetSpringEnabled(true)
			box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetSpringHertz(hertz)
		}
	}
}

func (human *Human) SetJointDampingRatio(dampingRatio float32) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	for i := 1; i < int(BoneCount); i++ {
		box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetSpringDampingRatio(dampingRatio)
	}
}

func (human *Human) EnableSensorEvents(enable bool) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	var bodyID box2d.Body = human.Bones[int(BoneTorso)].BodyID
	var shapeID box2d.Shape
	count := 0
	for id := range bodyID.Shapes() {
		shapeID = id
		count++
		break
	}
	if count == 1 {
		shapeID.SetSensorEventsEnabled(enable)
	}
}

func (human *Human) SetScale(scale float32) {
	if !(human.IsSpawned == true) {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}
	if !(float32(0.01) < scale && scale < float32(100.0)) {
		panic(&box2d.AssertError{Condition: "0.01f < scale && scale < 100.0f"})
	}
	if !(float32(0.0) < human.Scale) {
		panic(&box2d.AssertError{Condition: "0.0f < human->scale"})
	}
	var ratio float32 = scale / human.Scale
	var originalRatio float32 = scale / human.OriginalScale
	var frictionTorque float32 = (originalRatio * originalRatio * originalRatio) * human.FrictionTorque
	var origin box2d.Pos = human.Bones[0].BodyID.Position()
	for boneIndex := 0; boneIndex < int(BoneCount); boneIndex++ {
		var bone *Bone = &human.Bones[boneIndex]
		if boneIndex > 0 {
			var worldTransform box2d.WorldTransform = bone.BodyID.Transform()
			var d box2d.Vec2 = worldTransform.P.Sub(origin).Scale(ratio)
			bone.BodyID.SetTransform(origin.Add(d), worldTransform.Q)
			var localFrameA box2d.Transform = bone.JointID.LocalFrameA()
			var localFrameB box2d.Transform = bone.JointID.LocalFrameB()
			localFrameA.P = localFrameA.P.Scale(ratio)
			localFrameB.P = localFrameB.P.Scale(ratio)
			bone.JointID.SetLocalFrameA(localFrameA)
			bone.JointID.SetLocalFrameB(localFrameB)
			var shapeType box2d.JointType = bone.JointID.Type()
			if shapeType == box2d.RevoluteJointType {
				box2d.RevoluteJoint{Joint: bone.JointID}.SetMaxMotorTorque(bone.FrictionScale * frictionTorque)
			}
		}
		var shapeIDs [2]box2d.Shape
		shapeCount := 0
		for id := range bone.BodyID.Shapes() {
			shapeIDs[shapeCount] = id
			shapeCount++
			if shapeCount == 2 {
				break
			}
		}
		for shapeIndex := 0; shapeIndex < shapeCount; shapeIndex++ {
			var shapeType box2d.ShapeType = shapeIDs[shapeIndex].Type()
			if shapeType == box2d.CapsuleShape {
				var capsule box2d.Capsule = shapeIDs[shapeIndex].Capsule()
				capsule.Center1 = capsule.Center1.Scale(ratio)
				capsule.Center2 = capsule.Center2.Scale(ratio)
				capsule.Radius *= ratio
				shapeIDs[shapeIndex].SetCapsule(&capsule)
			} else if shapeType == box2d.PolygonShape {
				var polygon box2d.Polygon = shapeIDs[shapeIndex].Polygon()
				for pointIndex := 0; pointIndex < polygon.Count; pointIndex++ {
					polygon.Vertices[pointIndex] = polygon.Vertices[pointIndex].Scale(ratio)
				}
				polygon.Centroid = polygon.Centroid.Scale(ratio)
				polygon.Radius *= ratio
				shapeIDs[shapeIndex].SetPolygon(&polygon)
			}
		}
		bone.BodyID.UpdateMassFromShapes()
	}
	human.Scale = scale
}
