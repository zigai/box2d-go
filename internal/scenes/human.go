// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import "github.com/zigai/box2d"

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

type BoneID int

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

type humanJointSettings struct {
	maxTorque, hertz, dampingRatio float32
}

func CreateHuman(human *Human, worldID *box2d.World, position box2d.Pos, scale, frictionTorque, hertz, dampingRatio float32, groupIndex int, userData any, colorize bool) {
	if human.IsSpawned {
		panic(&box2d.AssertError{Condition: "human->isSpawned == false"})
	}

	for i := range int(BoneCount) {
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

	var (
		s             float32           = scale
		jointSettings                   = humanJointSettings{maxTorque: frictionTorque * s, hertz: hertz, dampingRatio: dampingRatio}
		shirtColor    box2d.HexColor    = box2d.ColorMediumTurquoise
		pantColor     box2d.HexColor    = box2d.ColorDodgerBlue
		skinColors    [4]box2d.HexColor = [4]box2d.HexColor{box2d.ColorNavajoWhite, box2d.ColorLightYellow, box2d.ColorPeru, box2d.ColorTan}
		skinColor     box2d.HexColor    = skinColors[groupIndex%4]
	)
	{
		var bone *Bone = &human.Bones[int(BoneHip)]

		bone.ParentIndex = -1
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.95) * s}), float32(0.0), "hip")

		setHumanShapeColor(&shapeDef, colorize, pantColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.02)*s, float32(0.02)*s, float32(0.095)*s)
	}
	{
		var bone *Bone = &human.Bones[int(BoneTorso)]

		bone.ParentIndex = int(BoneHip)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.2) * s}), float32(0.0), "torso")
		bone.FrictionScale = float32(0.5)
		bodyDef.Type = box2d.DynamicBody

		setHumanShapeColor(&shapeDef, colorize, shirtColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.135)*s, float32(0.135)*s, float32(0.09)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.0) * s}), -float32(0.25)*box2d.Pi, float32(0.0), box2d.Rot{C: 1}, jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneHead)]

		bone.ParentIndex = int(BoneTorso)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0) * s, Y: float32(1.475) * s}), float32(0.1), "head")
		bone.FrictionScale = float32(0.25)

		setHumanShapeColor(&shapeDef, colorize, skinColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.038)*s, float32(0.039)*s, float32(0.075)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.4) * s}), -float32(0.3)*box2d.Pi, float32(0.1)*box2d.Pi, box2d.Rot{C: 1}, jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperLeftLeg)]

		bone.ParentIndex = int(BoneHip)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.775) * s}), float32(0.0), "upper_left_leg")
		bone.FrictionScale = float32(1.0)

		setHumanShapeColor(&shapeDef, colorize, pantColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.125)*s, float32(0.125)*s, float32(0.06)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.9) * s}), -float32(0.05)*box2d.Pi, float32(0.4)*box2d.Pi, box2d.Rot{C: 1}, jointSettings)
	}

	var (
		points      [4]box2d.Vec2 = [4]box2d.Vec2{{X: -float32(0.03) * s, Y: -float32(0.185) * s}, {X: float32(0.11) * s, Y: -float32(0.185) * s}, {X: float32(0.11) * s, Y: -float32(0.16) * s}, {X: -float32(0.03) * s, Y: -float32(0.14) * s}}
		footHull    box2d.Hull    = box2d.ComputeHull(points[:])
		footPolygon box2d.Polygon = box2d.MakePolygon(&footHull, float32(0.015)*s)
	)

	createLowerLeg := func(bone *Bone, name string) {
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.475) * s}), float32(0.0), name)
		bone.FrictionScale = float32(0.5)

		setHumanShapeColor(&shapeDef, colorize, pantColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.155)*s, float32(0.125)*s, float32(0.045)*s)
		bone.BodyID.CreatePolygonShape(&footShapeDef, &footPolygon)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.625) * s}), -float32(0.5)*box2d.Pi, -float32(0.02)*box2d.Pi, box2d.Rot{C: 1}, jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerLeftLeg)]

		bone.ParentIndex = int(BoneUpperLeftLeg)
		createLowerLeg(bone, "lower_left_leg")
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperRightLeg)]

		bone.ParentIndex = int(BoneHip)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.775) * s}), float32(0.0), "upper_right_leg")
		bone.FrictionScale = float32(1.0)

		setHumanShapeColor(&shapeDef, colorize, pantColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.125)*s, float32(0.125)*s, float32(0.06)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.9) * s}), -float32(0.05)*box2d.Pi, float32(0.4)*box2d.Pi, box2d.Rot{C: 1}, jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerRightLeg)]

		bone.ParentIndex = int(BoneUpperRightLeg)
		createLowerLeg(bone, "lower_right_leg")
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperLeftArm)]

		bone.ParentIndex = int(BoneTorso)
		bone.FrictionScale = float32(0.5)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.225) * s}), float32(0.0), "upper_left_arm")

		setHumanShapeColor(&shapeDef, colorize, shirtColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.125)*s, float32(0.125)*s, float32(0.035)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.35) * s}), -float32(0.1)*box2d.Pi, float32(0.8)*box2d.Pi, box2d.Rot{C: 1}, jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerLeftArm)]

		bone.ParentIndex = int(BoneUpperLeftArm)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.975) * s}), float32(0.1), "lower_left_arm")
		bone.FrictionScale = float32(0.1)

		setHumanShapeColor(&shapeDef, colorize, skinColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.125)*s, float32(0.125)*s, float32(0.03)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.1) * s}), -float32(0.2)*box2d.Pi, float32(0.3)*box2d.Pi, box2d.MakeRot(float32(0.25)*box2d.Pi), jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneUpperRightArm)]

		bone.ParentIndex = int(BoneTorso)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.225) * s}), float32(0.0), "upper_right_arm")
		bone.FrictionScale = float32(0.5)

		setHumanShapeColor(&shapeDef, colorize, shirtColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.125)*s, float32(0.125)*s, float32(0.035)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.35) * s}), -float32(0.1)*box2d.Pi, float32(0.8)*box2d.Pi, box2d.Rot{C: 1}, jointSettings)
	}
	{
		var bone *Bone = &human.Bones[int(BoneLowerRightArm)]

		bone.ParentIndex = int(BoneUpperRightArm)
		createHumanBoneBody(worldID, bone, &bodyDef, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(0.975) * s}), float32(0.1), "lower_right_arm")
		bone.FrictionScale = float32(0.1)

		setHumanShapeColor(&shapeDef, colorize, skinColor)

		createHumanBoneCapsule(bone.BodyID, &shapeDef, -float32(0.125)*s, float32(0.125)*s, float32(0.03)*s)

		createHumanBoneJoint(human, worldID, bone, position.Add(box2d.Vec2{X: float32(0.0), Y: float32(1.1) * s}), -float32(0.2)*box2d.Pi, float32(0.3)*box2d.Pi, box2d.MakeRot(float32(0.25)*box2d.Pi), jointSettings)
	}

	human.IsSpawned = true
}

func DestroyHuman(human *Human) {
	if !human.IsSpawned {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}

	for i := range int(BoneCount) {
		if human.Bones[i].JointID == (box2d.Joint{}) {
			continue
		}

		human.Bones[i].JointID.Destroy()
		human.Bones[i].JointID = box2d.Joint{}
	}

	for i := range int(BoneCount) {
		if human.Bones[i].BodyID == (box2d.Body{}) {
			continue
		}

		human.Bones[i].BodyID.Destroy()
		human.Bones[i].BodyID = box2d.Body{}
	}

	human.IsSpawned = false
}

func (human *Human) SetVelocity(velocity box2d.Vec2) {
	for i := range int(BoneCount) {
		var bodyID box2d.Body = human.Bones[i].BodyID
		if bodyID == (box2d.Body{}) {
			continue
		}

		bodyID.SetLinearVelocity(velocity)
	}
}

func (human *Human) ApplyRandomAngularImpulse(magnitude float32) {
	if !human.IsSpawned {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}

	var impulse float32 = RandomFloatRange(-magnitude, magnitude)
	human.Bones[int(BoneTorso)].BodyID.ApplyAngularImpulse(impulse, true)
}

func (human *Human) SetJointFrictionTorque(torque float32) {
	if !human.IsSpawned {
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
	if !human.IsSpawned {
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
	if !human.IsSpawned {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}

	for i := 1; i < int(BoneCount); i++ {
		box2d.RevoluteJoint{Joint: human.Bones[i].JointID}.SetSpringDampingRatio(dampingRatio)
	}
}

func (human *Human) EnableSensorEvents(enable bool) {
	if !human.IsSpawned {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}

	var (
		bodyID  box2d.Body = human.Bones[int(BoneTorso)].BodyID
		shapeID box2d.Shape
	)

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
	if !human.IsSpawned {
		panic(&box2d.AssertError{Condition: "human->isSpawned == true"})
	}

	if !(float32(0.01) < scale && scale < float32(100.0)) {
		panic(&box2d.AssertError{Condition: "0.01f < scale && scale < 100.0f"})
	}

	if !(float32(0.0) < human.Scale) {
		panic(&box2d.AssertError{Condition: "0.0f < human->scale"})
	}

	var (
		ratio          float32   = scale / human.Scale
		originalRatio  float32   = scale / human.OriginalScale
		frictionTorque float32   = (originalRatio * originalRatio * originalRatio) * human.FrictionTorque
		origin         box2d.Pos = human.Bones[0].BodyID.Position()
	)
	for boneIndex := range int(BoneCount) {
		var bone *Bone = &human.Bones[boneIndex]
		if boneIndex > 0 {
			var (
				worldTransform box2d.WorldTransform = bone.BodyID.Transform()
				d              box2d.Vec2           = worldTransform.P.Sub(origin).Scale(ratio)
			)
			bone.BodyID.SetTransform(origin.Add(d), worldTransform.Q)

			var (
				localFrameA box2d.Transform = bone.JointID.LocalFrameA()
				localFrameB box2d.Transform = bone.JointID.LocalFrameB()
			)

			localFrameA.P = localFrameA.P.Scale(ratio)
			localFrameB.P = localFrameB.P.Scale(ratio)

			bone.JointID.SetLocalFrameA(localFrameA)
			bone.JointID.SetLocalFrameB(localFrameB)

			var shapeType box2d.JointType = bone.JointID.Type()
			if shapeType == box2d.RevoluteJointType {
				box2d.RevoluteJoint{Joint: bone.JointID}.SetMaxMotorTorque(bone.FrictionScale * frictionTorque)
			}
		}

		scaleHumanBoneShapes(bone.BodyID, ratio)

		bone.BodyID.UpdateMassFromShapes()
	}

	human.Scale = scale
}

func createHumanBoneJoint(human *Human, worldID *box2d.World, bone *Bone, pivot box2d.Pos, lower, upper float32, referenceRotation box2d.Rot, settings humanJointSettings) {
	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.BodyA = human.Bones[bone.ParentIndex].BodyID
	jointDef.Base.BodyB = bone.BodyID
	jointDef.Base.LocalFrameA.P = jointDef.Base.BodyA.LocalPoint(pivot)
	jointDef.Base.LocalFrameA.Q = referenceRotation
	jointDef.Base.LocalFrameB.P = jointDef.Base.BodyB.LocalPoint(pivot)
	jointDef.EnableLimit = true
	jointDef.LowerAngle = lower
	jointDef.UpperAngle = upper
	jointDef.EnableMotor = true
	jointDef.MaxMotorTorque = bone.FrictionScale * settings.maxTorque
	jointDef.EnableSpring = settings.hertz > float32(0.0)
	jointDef.Hertz = settings.hertz
	jointDef.DampingRatio = settings.dampingRatio
	jointDef.Base.DrawScale = float32(0.05)
	bone.JointID = worldID.CreateRevoluteJoint(&jointDef).Joint
}

func setHumanShapeColor(shapeDef *box2d.ShapeDef, colorize bool, color box2d.HexColor) {
	if colorize {
		shapeDef.Material.CustomColor = uint32(color)
	}
}

func scaleHumanBoneShapes(bodyID box2d.Body, ratio float32) {
	var shapeIDs [2]box2d.Shape

	shapeCount := 0
	for id := range bodyID.Shapes() {
		shapeIDs[shapeCount] = id

		shapeCount++
		if shapeCount == 2 {
			break
		}
	}

	for shapeIndex := range shapeCount {
		switch shapeIDs[shapeIndex].Type() {
		case box2d.CapsuleShape:
			capsule := shapeIDs[shapeIndex].Capsule()
			capsule.Center1 = capsule.Center1.Scale(ratio)
			capsule.Center2 = capsule.Center2.Scale(ratio)
			capsule.Radius *= ratio
			shapeIDs[shapeIndex].SetCapsule(&capsule)
		case box2d.PolygonShape:
			polygon := shapeIDs[shapeIndex].Polygon()
			for pointIndex := range polygon.Count {
				polygon.Vertices[pointIndex] = polygon.Vertices[pointIndex].Scale(ratio)
			}

			polygon.Centroid = polygon.Centroid.Scale(ratio)
			polygon.Radius *= ratio
			shapeIDs[shapeIndex].SetPolygon(&polygon)
		case box2d.CircleShape, box2d.SegmentShape, box2d.ChainSegmentShape, box2d.ShapeTypeCount:
			// Human bones contain only capsules and polygons.
		}
	}
}

func createHumanBoneBody(worldID *box2d.World, bone *Bone, bodyDef *box2d.BodyDef, position box2d.Pos, damping float32, name string) {
	bodyDef.Position = position
	bodyDef.LinearDamping = damping
	bodyDef.Name = name
	bone.BodyID = worldID.CreateBody(bodyDef)
}

func createHumanBoneCapsule(bodyID box2d.Body, shapeDef *box2d.ShapeDef, lower, upper, radius float32) {
	capsule := box2d.Capsule{Center1: box2d.Vec2{Y: lower}, Center2: box2d.Vec2{Y: upper}, Radius: radius}
	bodyID.CreateCapsuleShape(shapeDef, &capsule)
}
