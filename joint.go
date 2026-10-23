// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"math"

	"github.com/zigai/box2d/internal/array"
	"github.com/zigai/box2d/internal/core"
	"github.com/zigai/box2d/internal/idpool"
)

const (
	DistanceJointType  JointType = 0
	FilterJointType    JointType = 1
	MotorJointType     JointType = 2
	MoverJointType     JointType = 3
	PogoJointType      JointType = 4
	PrismaticJointType JointType = 5
	RevoluteJointType  JointType = 6
	WeldJointType      JointType = 7
	WheelJointType     JointType = 8
)

type JointType int32

// Base joint definition used by all joint types.
// The local frames are measured from the body's origin rather than the center of mass because:
// 1. you might not know where the center of mass will be
// 2. if you add/remove shapes from a body and recompute the mass, the joints will be broken.
type JointDef struct {
	// User data pointer
	UserData any
	// The first attached body
	BodyA Body
	// The second attached body
	BodyB Body
	// The first local joint frame
	LocalFrameA Transform
	// The second local joint frame
	LocalFrameB Transform
	// Force threshold for joint events
	ForceThreshold float32
	// Torque threshold for joint events
	TorqueThreshold float32
	// Constraint hertz (advanced feature)
	ConstraintHertz float32
	// Constraint damping ratio (advanced feature)
	ConstraintDampingRatio float32
	// Debug draw scale
	DrawScale float32
	// Set this flag to true if the attached bodies should collide
	CollideConnected bool
}

type FilterJoint struct {
	Joint
}

// A filter joint is used to disable collision between two specific bodies.
type FilterJointDef struct {
	// Base joint definition
	Base JointDef
	// Used internally to detect a valid definition. DO NOT SET.
	internalValue int32
}

// Joint events report joints that are awake and have a force and/or torque exceeding the threshold
// The observed forces and torques are not returned for efficiency reasons.
type JointEvent struct {
	// The joint id
	Joint Joint
	// The user data from the joint for convenience
	UserData any
}

// Joint events are buffered in the world and are available
// as event arrays after the time step is complete.
// Note: this data becomes invalid if joints are destroyed.
type JointEvents struct {
	// Array of events
	JointEvents []JointEvent
	// Number of events
	Count int
}

type jointEdge struct {
	bodyID  int
	prevKey int
	nextKey int
}

type joint struct {
	userData         any
	setIndex         int
	colorIndex       int
	localIndex       int
	edges            [2]jointEdge
	jointID          int
	islandID         int
	islandIndex      int
	drawScale        float32
	jointType        JointType
	generation       uint16
	collideConnected bool
}

type pogoJoint struct {
	normal              Vec2
	restLength          float32
	hertz               float32
	dampingRatio        float32
	maxTensionForce     float32
	maxCompressionForce float32
	impulse             float32
	indexA              int
	indexB              int
	frameA              Transform
	frameB              Transform
	deltaCenter         Vec2
	linearMass          float32
	velocity            float32
}

type distanceJoint struct {
	length           float32
	hertz            float32
	dampingRatio     float32
	lowerSpringForce float32
	upperSpringForce float32
	minLength        float32
	maxLength        float32
	maxMotorForce    float32
	motorSpeed       float32
	impulse          float32
	lowerImpulse     float32
	upperImpulse     float32
	motorImpulse     float32
	indexA           int
	indexB           int
	anchorA          Vec2
	anchorB          Vec2
	deltaCenter      Vec2
	distanceSoftness softness
	axialMass        float32
	enableSpring     bool
	enableLimit      bool
	enableMotor      bool
}

type motorJoint struct {
	linearVelocity         Vec2
	maxVelocityForce       float32
	angularVelocity        float32
	maxVelocityTorque      float32
	linearHertz            float32
	linearDampingRatio     float32
	maxSpringForce         float32
	angularHertz           float32
	angularDampingRatio    float32
	maxSpringTorque        float32
	linearVelocityImpulse  Vec2
	angularVelocityImpulse float32
	linearSpringImpulse    Vec2
	angularSpringImpulse   float32
	linearSpring           softness
	angularSpring          softness
	indexA                 int
	indexB                 int
	frameA                 Transform
	frameB                 Transform
	deltaCenter            Vec2
	linearMass             Mat22
	angularMass            float32
}

type moverJoint struct {
	linearVelocity        Vec2
	maxVelocityForce      Vec2
	linearVelocityImpulse Vec2
	indexA                int
	indexB                int
	frameA                Transform
	frameB                Transform
	linearMass            float32
}

type prismaticJoint struct {
	impulse           Vec2
	springImpulse     float32
	motorImpulse      float32
	lowerImpulse      float32
	upperImpulse      float32
	hertz             float32
	dampingRatio      float32
	targetTranslation float32
	maxMotorForce     float32
	motorSpeed        float32
	lowerTranslation  float32
	upperTranslation  float32
	indexA            int
	indexB            int
	frameA            Transform
	frameB            Transform
	deltaCenter       Vec2
	springSoftness    softness
	enableSpring      bool
	enableLimit       bool
	enableMotor       bool
}

type revoluteJoint struct {
	linearImpulse  Vec2
	springImpulse  float32
	motorImpulse   float32
	lowerImpulse   float32
	upperImpulse   float32
	hertz          float32
	dampingRatio   float32
	targetAngle    float32
	maxMotorTorque float32
	motorSpeed     float32
	lowerAngle     float32
	upperAngle     float32
	indexA         int
	indexB         int
	frameA         Transform
	frameB         Transform
	deltaCenter    Vec2
	axialMass      float32
	springSoftness softness
	enableSpring   bool
	enableMotor    bool
	enableLimit    bool
}

type weldJoint struct {
	linearHertz         float32
	linearDampingRatio  float32
	angularHertz        float32
	angularDampingRatio float32
	linearSpring        softness
	angularSpring       softness
	linearImpulse       Vec2
	angularImpulse      float32
	indexA              int
	indexB              int
	frameA              Transform
	frameB              Transform
	deltaCenter         Vec2
	axialMass           float32
}

type wheelJoint struct {
	perpImpulse      float32
	motorImpulse     float32
	springImpulse    float32
	lowerImpulse     float32
	upperImpulse     float32
	maxMotorTorque   float32
	motorSpeed       float32
	lowerTranslation float32
	upperTranslation float32
	hertz            float32
	dampingRatio     float32
	indexA           int
	indexB           int
	frameA           Transform
	frameB           Transform
	deltaCenter      Vec2
	perpMass         float32
	motorMass        float32
	axialMass        float32
	springSoftness   softness
	enableSpring     bool
	enableMotor      bool
	enableLimit      bool
}

type jointSim struct {
	jointID                int
	bodyIDA                int
	bodyIDB                int
	jointType              JointType
	localFrameA            Transform
	localFrameB            Transform
	invMassA               float32
	invMassB               float32
	invIA                  float32
	invIB                  float32
	constraintHertz        float32
	constraintDampingRatio float32
	constraintSoftness     softness
	forceThreshold         float32
	torqueThreshold        float32
	distanceJoint          distanceJoint
	motorJoint             motorJoint
	moverJoint             moverJoint
	pogoJoint              pogoJoint
	revoluteJoint          revoluteJoint
	prismaticJoint         prismaticJoint
	weldJoint              weldJoint
	wheelJoint             wheelJoint
}

type jointPair struct {
	joint    *joint
	jointSim *jointSim
}

// Use this to initialize your joint definition.
func DefaultDistanceJointDef() DistanceJointDef {
	return DistanceJointDef{
		Base:             defaultJointDef(),
		Length:           1,
		LowerSpringForce: -math.MaxFloat32,
		UpperSpringForce: math.MaxFloat32,
		MaxLength:        Huge(),
		internalValue:    secretCookie,
	}
}

// Use this to initialize your joint definition.
func DefaultFilterJointDef() FilterJointDef {
	return FilterJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultMotorJointDef() MotorJointDef {
	return MotorJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultMoverJointDef() MoverJointDef {
	return MoverJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultPogoJointDef() PogoJointDef {
	return PogoJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultPrismaticJointDef() PrismaticJointDef {
	return PrismaticJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultRevoluteJointDef() RevoluteJointDef {
	return RevoluteJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultWeldJointDef() WeldJointDef {
	return WeldJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultWheelJointDef() WheelJointDef {
	return WheelJointDef{Base: defaultJointDef(), internalValue: secretCookie, EnableSpring: true, Hertz: 1, DampingRatio: 0.7}
}

// Use this to initialize your explosion definition.
func DefaultExplosionDef() ExplosionDef {
	return ExplosionDef{MaskBits: DefaultMaskBits}
}

func defaultJointDef() JointDef {
	return JointDef{
		LocalFrameA:            Transform{Q: Rot{C: 1}},
		LocalFrameB:            Transform{Q: Rot{C: 1}},
		ForceThreshold:         math.MaxFloat32,
		TorqueThreshold:        math.MaxFloat32,
		ConstraintHertz:        60,
		ConstraintDampingRatio: 2,
		DrawScale:              lengthUnitsPerMeter,
	}
}

func getJointFullID(world *World, jointID Joint) *joint {
	var (
		id  = int(jointID.index1) - 1
		jnt = &world.joints[id]
	)
	core.Assert(jnt.jointID == id && jnt.generation == jointID.generation, "joint->jointId == id && joint->generation == jointId.generation")

	return jnt
}

func getJointSim(world *World, jnt *joint) *jointSim {
	if jnt.setIndex == awakeSet {
		core.Assert(0 <= jnt.colorIndex && jnt.colorIndex < GraphColorCount, "0 <= joint->colorIndex && joint->colorIndex < B2_GRAPH_COLOR_COUNT")

		color := &world.constraintGraph.colors[jnt.colorIndex]

		return &color.jointSims[jnt.localIndex]
	}

	set := &world.solverSets[jnt.setIndex]

	return &set.jointSims[jnt.localIndex]
}

func getJointSimCheckType(jointID Joint, jointType JointType) *jointSim {
	world := jointID.world
	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return nil
	}

	jnt := getJointFullID(world, jointID)
	core.Assert(jnt.jointType == jointType, "joint->type == type")

	sim := getJointSim(world, jnt)
	core.Assert(sim.jointType == jointType, "jointSim->type == type")

	return sim
}

func destroyContactsBetweenBodies(world *World, bodyA *body, bodyB *body) {
	var (
		contactKey  int
		otherBodyID int
	)

	// use the smaller of the two contact lists
	if bodyA.contactCount < bodyB.contactCount {
		contactKey = bodyA.headContactKey
		otherBodyID = bodyB.id
	} else {
		contactKey = bodyB.headContactKey
		otherBodyID = bodyA.id
	}

	// destroy the contacts
	for contactKey != nullIndex {
		var (
			contactID = contactKey >> 1
			edgeIndex = contactKey & 1
		)

		contact := &world.contacts[contactID]

		contactKey = contact.edges[edgeIndex].nextKey

		otherEdgeIndex := edgeIndex ^ 1
		if contact.edges[otherEdgeIndex].bodyID == otherBodyID {
			// Careful, this removes the contact from the current doubly linked list
			destroyContact(world, contact)
		}
	}

	validateSolverSets(world)
}

func isValidJointDef(def *JointDef) bool {
	return def.LocalFrameA.IsValid() && def.LocalFrameB.IsValid() &&
		IsValidFloat(def.ForceThreshold) && def.ForceThreshold >= 0.0 && IsValidFloat(def.TorqueThreshold) &&
		def.TorqueThreshold >= 0.0 && IsValidFloat(def.ConstraintHertz) && def.ConstraintHertz >= 0.0 &&
		IsValidFloat(def.ConstraintDampingRatio) && def.ConstraintDampingRatio >= 0.0 &&
		IsValidFloat(def.DrawScale)
}

func createJoint(world *World, def *JointDef, jointType JointType) jointPair {
	core.Assert(def.LocalFrameA.IsValid(), "b2IsValidTransform( def->localFrameA )")
	core.Assert(def.LocalFrameB.IsValid(), "b2IsValidTransform( def->localFrameB )")
	core.Assert(world == def.BodyA.world, "world->worldId == def->bodyIdA.world")
	core.Assert(world == def.BodyB.world, "world->worldId == def->bodyIdB.world")
	core.Assert(def.BodyA != def.BodyB, "B2_ID_EQUALS( def->bodyIdA, def->bodyIdB ) == false")

	var (
		bodyA = getBodyFullID(world, def.BodyA)
		bodyB = getBodyFullID(world, def.BodyB)
	)

	// If the joint prevents collisions, then destroy all contacts between attached bodies.
	// Do this first so world arrays are not disturbed in the following code.
	if !def.CollideConnected {
		destroyContactsBetweenBodies(world, bodyA, bodyB)
	}

	var (
		bodyIDA     = bodyA.id
		bodyIDB     = bodyB.id
		maxSetIndex = max(bodyA.setIndex, bodyB.setIndex)
	)

	// Create joint id and joint
	jointID := idpool.AllocWithGeneration(&world.jointIDPool, func(id int) uint16 {
		return world.joints[id].generation
	})

	if jointID == len(world.joints) {
		world.joints = array.Push(world.joints, joint{})
	}

	jnt := &world.joints[jointID]

	jnt.jointID = jointID
	jnt.userData = def.UserData
	jnt.generation += 1
	jnt.setIndex = nullIndex
	jnt.colorIndex = nullIndex
	jnt.localIndex = nullIndex
	jnt.islandID = nullIndex
	jnt.islandIndex = nullIndex
	jnt.drawScale = def.DrawScale
	jnt.jointType = jointType
	jnt.collideConnected = def.CollideConnected

	// Doubly linked list on bodyA
	jnt.edges[0].bodyID = bodyIDA
	jnt.edges[0].prevKey = nullIndex
	jnt.edges[0].nextKey = bodyA.headJointKey

	keyA := jointID << 1

	if bodyA.headJointKey != nullIndex {
		var (
			jointA = &world.joints[bodyA.headJointKey>>1]
			edgeA  = &jointA.edges[(bodyA.headJointKey & 1)]
		)

		edgeA.prevKey = keyA
	}

	bodyA.headJointKey = keyA
	bodyA.jointCount += 1

	// Doubly linked list on bodyB
	jnt.edges[1].bodyID = bodyIDB
	jnt.edges[1].prevKey = nullIndex
	jnt.edges[1].nextKey = bodyB.headJointKey

	keyB := (jointID << 1) | 1

	if bodyB.headJointKey != nullIndex {
		var (
			jointB = &world.joints[bodyB.headJointKey>>1]
			edgeB  = &jointB.edges[(bodyB.headJointKey & 1)]
		)

		edgeB.prevKey = keyB
	}

	bodyB.headJointKey = keyB
	bodyB.jointCount += 1

	sim := placeJointInSolverSet(world, jnt, bodyA, bodyB, maxSetIndex)

	sim.localFrameA = def.LocalFrameA
	sim.localFrameB = def.LocalFrameB
	sim.jointType = jointType
	sim.constraintHertz = def.ConstraintHertz
	sim.constraintDampingRatio = def.ConstraintDampingRatio
	sim.constraintSoftness = softness{
		biasRate:     0.0,
		massScale:    1.0,
		impulseScale: 0.0,
	}

	core.Assert(IsValidFloat(def.ForceThreshold) && def.ForceThreshold >= 0.0, "b2IsValidFloat( def->forceThreshold ) && def->forceThreshold >= 0.0f")
	core.Assert(IsValidFloat(def.TorqueThreshold) && def.TorqueThreshold >= 0.0, "b2IsValidFloat( def->torqueThreshold ) && def->torqueThreshold >= 0.0f")

	sim.forceThreshold = def.ForceThreshold
	sim.torqueThreshold = def.TorqueThreshold

	core.Assert(sim.jointID == jointID, "jointSim->jointId == jointId")
	core.Assert(sim.bodyIDA == bodyIDA, "jointSim->bodyIdA == bodyIdA")
	core.Assert(sim.bodyIDB == bodyIDB, "jointSim->bodyIdB == bodyIdB")

	if jnt.setIndex > disabledSet {
		// Add edge to island graph
		linkJoint(world, jnt)
	}

	validateSolverSets(world)

	return jointPair{jnt, sim}
}

// Create a distance joint.
func (w *World) CreateDistanceJoint(def *DistanceJointDef) DistanceJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.Length) && def.Length > 0.0, "b2IsValidFloat( def->length ) && def->length > 0.0f")
	checkInput(IsValidFloat(def.Hertz), "b2IsValidFloat( def->hertz )")
	checkInput(IsValidFloat(def.DampingRatio), "b2IsValidFloat( def->dampingRatio )")
	checkInput(IsValidFloat(def.MinLength), "b2IsValidFloat( def->minLength )")
	checkInput(IsValidFloat(def.MaxLength), "b2IsValidFloat( def->maxLength )")
	checkInput(IsValidFloat(def.MaxMotorForce), "b2IsValidFloat( def->maxMotorForce )")
	checkInput(IsValidFloat(def.MotorSpeed), "b2IsValidFloat( def->motorSpeed )")
	checkInput(IsValidFloat(def.LowerSpringForce), "b2IsValidFloat( def->lowerSpringForce )")
	checkInput(IsValidFloat(def.UpperSpringForce), "b2IsValidFloat( def->upperSpringForce )")
	core.Assert(def.LowerSpringForce <= def.UpperSpringForce, "def->lowerSpringForce <= def->upperSpringForce")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return DistanceJoint{}
	}

	pair := createJoint(world, &def.Base, DistanceJointType)

	jnt := pair.jointSim

	empty := distanceJoint{}

	jnt.distanceJoint = empty
	jnt.distanceJoint.length = maxFloat(def.Length, LinearSlop())
	jnt.distanceJoint.hertz = def.Hertz
	jnt.distanceJoint.dampingRatio = def.DampingRatio
	jnt.distanceJoint.minLength = maxFloat(def.MinLength, LinearSlop())
	jnt.distanceJoint.maxLength = maxFloat(def.MinLength, def.MaxLength)
	jnt.distanceJoint.maxMotorForce = def.MaxMotorForce
	jnt.distanceJoint.motorSpeed = def.MotorSpeed
	jnt.distanceJoint.enableSpring = def.EnableSpring
	jnt.distanceJoint.lowerSpringForce = minFloat(def.LowerSpringForce, def.UpperSpringForce)
	jnt.distanceJoint.upperSpringForce = maxFloat(def.LowerSpringForce, def.UpperSpringForce)
	jnt.distanceJoint.enableLimit = def.EnableLimit
	jnt.distanceJoint.enableMotor = def.EnableMotor
	jnt.distanceJoint.impulse = 0.0
	jnt.distanceJoint.lowerImpulse = 0.0
	jnt.distanceJoint.upperImpulse = 0.0
	jnt.distanceJoint.motorImpulse = 0.0

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateDistanceJoint(world.recording, &recArgsCreateDistanceJoint{def: *def}, jointID)
	}

	return DistanceJoint{Joint: jointID}
}

// Create a filter joint.
func (w *World) CreateFilterJoint(def *FilterJointDef) FilterJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return FilterJoint{}
	}

	pair := createJoint(world, &def.Base, FilterJointType)

	jnt := pair.jointSim

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateFilterJoint(world.recording, &recArgsCreateFilterJoint{def: *def}, jointID)
	}

	return FilterJoint{Joint: jointID}
}

// Create a motor joint.
func (w *World) CreateMotorJoint(def *MotorJointDef) MotorJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(def.LinearVelocity.IsValid(), "b2IsValidVec2( def->linearVelocity )")
	checkInput(IsValidFloat(def.MaxVelocityForce), "b2IsValidFloat( def->maxVelocityForce )")
	checkInput(IsValidFloat(def.AngularVelocity), "b2IsValidFloat( def->angularVelocity )")
	checkInput(IsValidFloat(def.MaxVelocityTorque), "b2IsValidFloat( def->maxVelocityTorque )")
	checkInput(IsValidFloat(def.LinearHertz), "b2IsValidFloat( def->linearHertz )")
	checkInput(IsValidFloat(def.LinearDampingRatio), "b2IsValidFloat( def->linearDampingRatio )")
	checkInput(IsValidFloat(def.MaxSpringForce), "b2IsValidFloat( def->maxSpringForce )")
	checkInput(IsValidFloat(def.AngularHertz), "b2IsValidFloat( def->angularHertz )")
	checkInput(IsValidFloat(def.AngularDampingRatio), "b2IsValidFloat( def->angularDampingRatio )")
	checkInput(IsValidFloat(def.MaxSpringTorque), "b2IsValidFloat( def->maxSpringTorque )")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return MotorJoint{}
	}

	var (
		pair = createJoint(world, &def.Base, MotorJointType)
		jnt  = pair.jointSim
	)

	jnt.motorJoint = motorJoint{}
	jnt.motorJoint.linearVelocity = def.LinearVelocity
	jnt.motorJoint.maxVelocityForce = def.MaxVelocityForce
	jnt.motorJoint.angularVelocity = def.AngularVelocity
	jnt.motorJoint.maxVelocityTorque = def.MaxVelocityTorque
	jnt.motorJoint.linearHertz = def.LinearHertz
	jnt.motorJoint.linearDampingRatio = def.LinearDampingRatio
	jnt.motorJoint.maxSpringForce = def.MaxSpringForce
	jnt.motorJoint.angularHertz = def.AngularHertz
	jnt.motorJoint.angularDampingRatio = def.AngularDampingRatio
	jnt.motorJoint.maxSpringTorque = def.MaxSpringTorque

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateMotorJoint(world.recording, &recArgsCreateMotorJoint{def: *def}, jointID)
	}

	return MotorJoint{Joint: jointID}
}

// Create a mover joint.
func (w *World) CreateMoverJoint(def *MoverJointDef) MoverJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(def.LinearVelocity.IsValid(), "b2IsValidVec2( def->linearVelocity )")
	checkInput(def.MaxVelocityForce.IsValid(), "b2IsValidVec2( def->maxVelocityForce )")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return MoverJoint{}
	}

	var (
		pair = createJoint(world, &def.Base, MoverJointType)
		jnt  = pair.jointSim
	)

	jnt.moverJoint = moverJoint{}
	jnt.moverJoint.linearVelocity = def.LinearVelocity
	jnt.moverJoint.maxVelocityForce = def.MaxVelocityForce

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateMoverJoint(world.recording, &recArgsCreateMoverJoint{def: *def}, jointID)
	}

	return MoverJoint{Joint: jointID}
}

// Create a pog joint.
func (w *World) CreatePogoJoint(def *PogoJointDef) PogoJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(def.Normal.IsValid(), "b2IsValidVec2( def->normal )")
	checkInput(IsValidFloat(def.RestLength) && def.RestLength >= 0.0, "b2IsValidFloat( def->restLength ) && def->restLength >= 0.0f")
	checkInput(IsValidFloat(def.Hertz) && def.Hertz >= 0.0, "b2IsValidFloat( def->hertz ) && def->hertz >= 0.0f")
	checkInput(IsValidFloat(def.DampingRatio) && def.DampingRatio >= 0.0, "b2IsValidFloat( def->dampingRatio ) && def->dampingRatio >= 0.0f")
	checkInput(IsValidFloat(def.MaxTensionForce) && def.MaxTensionForce >= 0.0, "b2IsValidFloat( def->maxTensionForce ) && def->maxTensionForce >= 0.0f")
	checkInput(IsValidFloat(def.MaxCompressionForce) && def.MaxCompressionForce >= 0.0, "b2IsValidFloat( def->maxCompressionForce ) && def->maxCompressionForce >= 0.0f")
	checkInput(IsValidFloat(def.Impulse), "b2IsValidFloat( def->impulse )")
	checkInput(IsValidFloat(def.Velocity), "b2IsValidFloat( def->velocity )")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return PogoJoint{}
	}

	pair := createJoint(world, &def.Base, PogoJointType)

	jnt := pair.jointSim

	jnt.pogoJoint = pogoJoint{}
	jnt.pogoJoint.restLength = def.RestLength
	jnt.pogoJoint.hertz = def.Hertz
	jnt.pogoJoint.dampingRatio = def.DampingRatio
	jnt.pogoJoint.maxTensionForce = def.MaxTensionForce
	jnt.pogoJoint.maxCompressionForce = def.MaxCompressionForce
	jnt.pogoJoint.normal = def.Normal
	jnt.pogoJoint.impulse = def.Impulse
	jnt.pogoJoint.velocity = def.Velocity

	jointID := Joint{
		index1:     int32(jnt.jointID + 1),
		world:      world,
		generation: pair.joint.generation,
	}

	if world.recording != nil {
		recWriteRetCreatePogoJoint(world.recording, &recArgsCreatePogoJoint{def: *def}, jointID)
	}

	return PogoJoint{Joint: jointID}
}

// Create a prismatic (slider) joint.
func (w *World) CreatePrismaticJoint(def *PrismaticJointDef) PrismaticJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.Hertz), "b2IsValidFloat( def->hertz )")
	checkInput(IsValidFloat(def.DampingRatio), "b2IsValidFloat( def->dampingRatio )")
	checkInput(IsValidFloat(def.TargetTranslation), "b2IsValidFloat( def->targetTranslation )")
	checkInput(IsValidFloat(def.LowerTranslation), "b2IsValidFloat( def->lowerTranslation )")
	checkInput(IsValidFloat(def.UpperTranslation), "b2IsValidFloat( def->upperTranslation )")
	checkInput(IsValidFloat(def.MaxMotorForce), "b2IsValidFloat( def->maxMotorForce )")
	checkInput(IsValidFloat(def.MotorSpeed), "b2IsValidFloat( def->motorSpeed )")
	core.Assert(def.LowerTranslation <= def.UpperTranslation, "def->lowerTranslation <= def->upperTranslation")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return PrismaticJoint{}
	}

	pair := createJoint(world, &def.Base, PrismaticJointType)

	jnt := pair.jointSim

	jnt.prismaticJoint = prismaticJoint{}
	jnt.prismaticJoint.hertz = def.Hertz
	jnt.prismaticJoint.dampingRatio = def.DampingRatio
	jnt.prismaticJoint.targetTranslation = def.TargetTranslation
	jnt.prismaticJoint.lowerTranslation = minFloat(def.LowerTranslation, def.UpperTranslation)
	jnt.prismaticJoint.upperTranslation = maxFloat(def.LowerTranslation, def.UpperTranslation)
	jnt.prismaticJoint.maxMotorForce = def.MaxMotorForce
	jnt.prismaticJoint.motorSpeed = def.MotorSpeed
	jnt.prismaticJoint.enableSpring = def.EnableSpring
	jnt.prismaticJoint.enableLimit = def.EnableLimit
	jnt.prismaticJoint.enableMotor = def.EnableMotor

	jointID := Joint{
		index1:     int32(jnt.jointID + 1),
		world:      world,
		generation: pair.joint.generation,
	}

	if world.recording != nil {
		recWriteRetCreatePrismaticJoint(world.recording, &recArgsCreatePrismaticJoint{def: *def}, jointID)
	}

	return PrismaticJoint{Joint: jointID}
}

// Create a revolute joint.
func (w *World) CreateRevoluteJoint(def *RevoluteJointDef) RevoluteJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.TargetAngle), "b2IsValidFloat( def->targetAngle )")
	checkInput(IsValidFloat(def.Hertz), "b2IsValidFloat( def->hertz )")
	checkInput(IsValidFloat(def.DampingRatio), "b2IsValidFloat( def->dampingRatio )")
	checkInput(IsValidFloat(def.LowerAngle), "b2IsValidFloat( def->lowerAngle )")
	checkInput(IsValidFloat(def.UpperAngle), "b2IsValidFloat( def->upperAngle )")
	checkInput(IsValidFloat(def.MaxMotorTorque), "b2IsValidFloat( def->maxMotorTorque )")
	checkInput(IsValidFloat(def.MotorSpeed), "b2IsValidFloat( def->motorSpeed )")
	core.Assert(def.LowerAngle <= def.UpperAngle, "def->lowerAngle <= def->upperAngle")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return RevoluteJoint{}
	}

	pair := createJoint(world, &def.Base, RevoluteJointType)

	jnt := pair.jointSim

	empty := revoluteJoint{}

	jnt.revoluteJoint = empty

	jnt.revoluteJoint.targetAngle = clampFloat(def.TargetAngle, -Pi, Pi)
	jnt.revoluteJoint.hertz = def.Hertz
	jnt.revoluteJoint.dampingRatio = def.DampingRatio

	var (
		lowerAngle = minFloat(def.LowerAngle, def.UpperAngle)
		upperAngle = maxFloat(def.LowerAngle, def.UpperAngle)
	)

	jnt.revoluteJoint.lowerAngle = clampFloat(lowerAngle, float32(-0.99*Pi), float32(0.99*Pi))
	jnt.revoluteJoint.upperAngle = clampFloat(upperAngle, float32(-0.99*Pi), float32(0.99*Pi))

	jnt.revoluteJoint.maxMotorTorque = def.MaxMotorTorque
	jnt.revoluteJoint.motorSpeed = def.MotorSpeed
	jnt.revoluteJoint.enableSpring = def.EnableSpring
	jnt.revoluteJoint.enableLimit = def.EnableLimit
	jnt.revoluteJoint.enableMotor = def.EnableMotor

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateRevoluteJoint(world.recording, &recArgsCreateRevoluteJoint{def: *def}, jointID)
	}

	return RevoluteJoint{Joint: jointID}
}

// Create a weld joint.
func (w *World) CreateWeldJoint(def *WeldJointDef) WeldJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.LinearHertz), "b2IsValidFloat( def->linearHertz )")
	checkInput(IsValidFloat(def.AngularHertz), "b2IsValidFloat( def->angularHertz )")
	checkInput(IsValidFloat(def.LinearDampingRatio), "b2IsValidFloat( def->linearDampingRatio )")
	checkInput(IsValidFloat(def.AngularDampingRatio), "b2IsValidFloat( def->angularDampingRatio )")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return WeldJoint{}
	}

	pair := createJoint(world, &def.Base, WeldJointType)

	jnt := pair.jointSim

	empty := weldJoint{}

	jnt.weldJoint = empty
	jnt.weldJoint.linearHertz = def.LinearHertz
	jnt.weldJoint.linearDampingRatio = def.LinearDampingRatio
	jnt.weldJoint.angularHertz = def.AngularHertz
	jnt.weldJoint.angularDampingRatio = def.AngularDampingRatio
	jnt.weldJoint.linearImpulse = Vec2Zero
	jnt.weldJoint.angularImpulse = 0.0

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateWeldJoint(world.recording, &recArgsCreateWeldJoint{def: *def}, jointID)
	}

	return WeldJoint{Joint: jointID}
}

// Create a wheel joint.
func (w *World) CreateWheelJoint(def *WheelJointDef) WheelJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.Hertz), "b2IsValidFloat( def->hertz )")
	checkInput(IsValidFloat(def.DampingRatio), "b2IsValidFloat( def->dampingRatio )")
	checkInput(IsValidFloat(def.LowerTranslation), "b2IsValidFloat( def->lowerTranslation )")
	checkInput(IsValidFloat(def.UpperTranslation), "b2IsValidFloat( def->upperTranslation )")
	checkInput(IsValidFloat(def.MaxMotorTorque), "b2IsValidFloat( def->maxMotorTorque )")
	checkInput(IsValidFloat(def.MotorSpeed), "b2IsValidFloat( def->motorSpeed )")
	core.Assert(def.LowerTranslation <= def.UpperTranslation, "def->lowerTranslation <= def->upperTranslation")

	world := w

	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return WheelJoint{}
	}

	pair := createJoint(world, &def.Base, WheelJointType)

	jnt := pair.jointSim

	jnt.wheelJoint = wheelJoint{}
	jnt.wheelJoint.perpMass = 0.0
	jnt.wheelJoint.axialMass = 0.0
	jnt.wheelJoint.motorImpulse = 0.0
	jnt.wheelJoint.lowerImpulse = 0.0
	jnt.wheelJoint.upperImpulse = 0.0
	jnt.wheelJoint.lowerTranslation = minFloat(def.LowerTranslation, def.UpperTranslation)
	jnt.wheelJoint.upperTranslation = maxFloat(def.LowerTranslation, def.UpperTranslation)
	jnt.wheelJoint.maxMotorTorque = def.MaxMotorTorque
	jnt.wheelJoint.motorSpeed = def.MotorSpeed
	jnt.wheelJoint.hertz = def.Hertz
	jnt.wheelJoint.dampingRatio = def.DampingRatio
	jnt.wheelJoint.enableSpring = def.EnableSpring
	jnt.wheelJoint.enableLimit = def.EnableLimit
	jnt.wheelJoint.enableMotor = def.EnableMotor

	jointID := Joint{int32(jnt.jointID + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateWheelJoint(world.recording, &recArgsCreateWheelJoint{def: *def}, jointID)
	}

	return WheelJoint{Joint: jointID}
}

func destroyJointInternal(world *World, jnt *joint) {
	jointID := jnt.jointID

	var (
		edgeA = &jnt.edges[0]
		edgeB = &jnt.edges[1]
	)

	var (
		idA   = edgeA.bodyID
		idB   = edgeB.bodyID
		bodyA = &world.bodies[idA]
		bodyB = &world.bodies[idB]
	)

	// Remove from body A
	if edgeA.prevKey != nullIndex {
		var (
			prevJoint = &world.joints[edgeA.prevKey>>1]
			prevEdge  = &prevJoint.edges[(edgeA.prevKey & 1)]
		)

		prevEdge.nextKey = edgeA.nextKey
	}

	if edgeA.nextKey != nullIndex {
		var (
			nextJoint = &world.joints[edgeA.nextKey>>1]
			nextEdge  = &nextJoint.edges[(edgeA.nextKey & 1)]
		)

		nextEdge.prevKey = edgeA.prevKey
	}

	edgeKeyA := jointID << 1
	if bodyA.headJointKey == edgeKeyA {
		bodyA.headJointKey = edgeA.nextKey
	}

	bodyA.jointCount -= 1

	// Remove from body B
	if edgeB.prevKey != nullIndex {
		var (
			prevJoint = &world.joints[edgeB.prevKey>>1]
			prevEdge  = &prevJoint.edges[(edgeB.prevKey & 1)]
		)

		prevEdge.nextKey = edgeB.nextKey
	}

	if edgeB.nextKey != nullIndex {
		var (
			nextJoint = &world.joints[edgeB.nextKey>>1]
			nextEdge  = &nextJoint.edges[(edgeB.nextKey & 1)]
		)

		nextEdge.prevKey = edgeB.prevKey
	}

	edgeKeyB := (jointID << 1) | 1
	if bodyB.headJointKey == edgeKeyB {
		bodyB.headJointKey = edgeB.nextKey
	}

	bodyB.jointCount -= 1

	if jnt.islandID != nullIndex {
		core.Assert(jnt.setIndex > disabledSet, "joint->setIndex > b2_disabledSet")
		unlinkJoint(world, jnt)
	} else {
		core.Assert(jnt.setIndex <= disabledSet, "joint->setIndex <= b2_disabledSet")
	}

	// Remove joint from solver set that owns it
	var (
		setIndex   = jnt.setIndex
		localIndex = jnt.localIndex
	)

	if setIndex == awakeSet {
		removeJointFromGraph(world, jnt.edges[0].bodyID, jnt.edges[1].bodyID, jnt.colorIndex, localIndex)
	} else {
		var (
			set        = &world.solverSets[setIndex]
			movedIndex int
		)

		set.jointSims, movedIndex = array.RemoveSwap(set.jointSims, localIndex)
		if movedIndex != nullIndex {
			// Fix moved joint
			var (
				movedJointSim = &set.jointSims[localIndex]
				movedID       = movedJointSim.jointID
				movedJoint    = &world.joints[movedID]
			)
			core.Assert(movedJoint.localIndex == movedIndex, "movedJoint->localIndex == movedIndex")
			movedJoint.localIndex = localIndex
		}
	}

	// Free joint and id (preserve joint generation)
	jnt.setIndex = nullIndex
	jnt.localIndex = nullIndex
	jnt.colorIndex = nullIndex
	jnt.jointID = nullIndex
	world.jointIDPool.Free(jointID)

	wakeBody(world, bodyA)
	wakeBody(world, bodyB)

	validateSolverSets(world)
}

// Destroy a joint.
func (j Joint) Destroy() {
	world := j.world
	core.Assert(!world.locked, "world->locked == false")

	if world.locked {
		return
	}

	if world.recording != nil {
		recWriteDestroyJoint(world.recording, &recArgsDestroyJoint{joint: j})
	}

	jnt := getJointFullID(world, j)

	destroyJointInternal(world, jnt)
}

// Get the joint type.
func (j Joint) Type() JointType {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return jnt.jointType
}

// Get body A id on a joint.
func (j Joint) BodyA() Body {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return makeBodyID(world, jnt.edges[0].bodyID)
}

// Get body B id on a joint.
func (j Joint) BodyB() Body {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return makeBodyID(world, jnt.edges[1].bodyID)
}

// Distance returns j as a DistanceJoint, or false when j is not a distance joint.
func (j Joint) Distance() (DistanceJoint, bool) {
	if j.Type() != DistanceJointType {
		return DistanceJoint{}, false
	}

	return DistanceJoint{j}, true
}

// Filter returns j as a FilterJoint, or false when j is not a filter joint.
func (j Joint) Filter() (FilterJoint, bool) {
	if j.Type() != FilterJointType {
		return FilterJoint{}, false
	}

	return FilterJoint{j}, true
}

// Motor returns j as a MotorJoint, or false when j is not a motor joint.
func (j Joint) Motor() (MotorJoint, bool) {
	if j.Type() != MotorJointType {
		return MotorJoint{}, false
	}

	return MotorJoint{j}, true
}

// Mover returns j as a MoverJoint, or false when j is not a mover joint.
func (j Joint) Mover() (MoverJoint, bool) {
	if j.Type() != MoverJointType {
		return MoverJoint{}, false
	}

	return MoverJoint{j}, true
}

// Pogo returns j as a PogoJoint, or false when j is not a pogo joint.
func (j Joint) Pogo() (PogoJoint, bool) {
	if j.Type() != PogoJointType {
		return PogoJoint{}, false
	}

	return PogoJoint{j}, true
}

// Prismatic returns j as a PrismaticJoint, or false when j is not a prismatic joint.
func (j Joint) Prismatic() (PrismaticJoint, bool) {
	if j.Type() != PrismaticJointType {
		return PrismaticJoint{}, false
	}

	return PrismaticJoint{j}, true
}

// Revolute returns j as a RevoluteJoint, or false when j is not a revolute joint.
func (j Joint) Revolute() (RevoluteJoint, bool) {
	if j.Type() != RevoluteJointType {
		return RevoluteJoint{}, false
	}

	return RevoluteJoint{j}, true
}

// Weld returns j as a WeldJoint, or false when j is not a weld joint.
func (j Joint) Weld() (WeldJoint, bool) {
	if j.Type() != WeldJointType {
		return WeldJoint{}, false
	}

	return WeldJoint{j}, true
}

// Wheel returns j as a WheelJoint, or false when j is not a wheel joint.
func (j Joint) Wheel() (WheelJoint, bool) {
	if j.Type() != WheelJointType {
		return WheelJoint{}, false
	}

	return WheelJoint{j}, true
}

// Get the world that owns this joint.
func (j Joint) World() *World {
	return j.world
}

// Set the local frame on body A.
func (j Joint) SetLocalFrameA(localFrame Transform) {
	checkInput(localFrame.IsValid(), "b2IsValidTransform( localFrame )")

	world := j.world
	if world.recording != nil {
		recWriteJointSetLocalFrameA(world.recording, &recArgsJointSetLocalFrameA{joint: j, localFrame: localFrame})
	}

	var (
		jnt = getJointFullID(world, j)
		sim = getJointSim(world, jnt)
	)

	sim.localFrameA = localFrame
}

// Get the local frame on body A.
func (j Joint) LocalFrameA() Transform {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		sim   = getJointSim(world, jnt)
	)

	return sim.localFrameA
}

// Set the local frame on body B.
func (j Joint) SetLocalFrameB(localFrame Transform) {
	checkInput(localFrame.IsValid(), "b2IsValidTransform( localFrame )")

	world := j.world
	if world.recording != nil {
		recWriteJointSetLocalFrameB(world.recording, &recArgsJointSetLocalFrameB{joint: j, localFrame: localFrame})
	}

	var (
		jnt = getJointFullID(world, j)
		sim = getJointSim(world, jnt)
	)

	sim.localFrameB = localFrame
}

// Get the local frame on body B.
func (j Joint) LocalFrameB() Transform {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		sim   = getJointSim(world, jnt)
	)

	return sim.localFrameB
}

// Toggle collision between connected bodies.
func (j Joint) SetCollideConnected(shouldCollide bool) {
	world := getWorldLocked(j.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteJointSetCollideConnected(world.recording, &recArgsJointSetCollideConnected{joint: j, shouldCollide: shouldCollide})
	}

	jnt := getJointFullID(world, j)
	if jnt.collideConnected == shouldCollide {
		return
	}

	jnt.collideConnected = shouldCollide

	var (
		bodyA = &world.bodies[jnt.edges[0].bodyID]
		bodyB = &world.bodies[jnt.edges[1].bodyID]
	)

	if !shouldCollide {
		destroyContactsBetweenBodies(world, bodyA, bodyB)
		return
	}

	var (
		typeA = bodyA.bodyType
		typeB = bodyB.bodyType
	)

	shapeID := nullIndex

	if typeA == DynamicBody && typeB == DynamicBody {
		if bodyA.shapeCount < bodyB.shapeCount {
			shapeID = bodyA.headShapeID
		} else {
			shapeID = bodyB.headShapeID
		}
	} else if typeA == DynamicBody {
		shapeID = bodyA.headShapeID
	} else if typeB == DynamicBody {
		shapeID = bodyB.headShapeID
	}

	for shapeID != nullIndex {
		shape := &world.shapes[shapeID]

		if shape.proxyKey != nullIndex {
			world.broadPhase.markProxyMovedSerial(shape.proxyKey)
		}

		shapeID = shape.nextShapeID
	}
}

// Is collision allowed between connected bodies?
func (j Joint) CollideConnected() bool {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return jnt.collideConnected
}

// Set the user data on a joint.
func (j Joint) SetUserData(userData any) {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	jnt.userData = userData
}

// Get the user data on a joint.
func (j Joint) UserData() any {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return jnt.userData
}

// Wake the bodies connect to this joint.
func (j Joint) WakeBodies() {
	world := getWorldLocked(j.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteJointWakeBodies(world.recording, &recArgsJointWakeBodies{joint: j})
	}

	var (
		jnt   = getJointFullID(world, j)
		bodyA = &world.bodies[jnt.edges[0].bodyID]
		bodyB = &world.bodies[jnt.edges[1].bodyID]
	)

	wakeBody(world, bodyA)
	wakeBody(world, bodyB)
}

func getJointReaction(sim *jointSim, invTimeStep float32) (float32, float32) {
	var (
		linearImpulse  float32 = 0.0
		angularImpulse float32 = 0.0
	)

	switch sim.jointType {
	case DistanceJointType:
		{
			jnt := &sim.distanceJoint

			linearImpulse = absFloat(jnt.impulse + jnt.lowerImpulse - jnt.upperImpulse + jnt.motorImpulse)
		}

	case MotorJointType:
		{
			jnt := &sim.motorJoint

			linearImpulse = jnt.linearVelocityImpulse.Add(jnt.linearSpringImpulse).Length()
			angularImpulse = absFloat(jnt.angularVelocityImpulse + jnt.angularSpringImpulse)
		}

	case MoverJointType:
		{
			jnt := &sim.moverJoint

			linearImpulse = jnt.linearVelocityImpulse.Length()
		}

	case PogoJointType:
		{
			jnt := &sim.pogoJoint

			linearImpulse = absFloat(jnt.impulse)
		}

	case PrismaticJointType:
		{
			var (
				jnt          = &sim.prismaticJoint
				perpImpulse  = jnt.impulse.X
				axialImpulse = jnt.motorImpulse + jnt.lowerImpulse - jnt.upperImpulse
			)

			linearImpulse = float32(math.Sqrt(float64(float32(perpImpulse*perpImpulse) + float32(axialImpulse*axialImpulse))))
			angularImpulse = absFloat(jnt.impulse.Y)
		}

	case RevoluteJointType:
		{
			jnt := &sim.revoluteJoint

			linearImpulse = jnt.linearImpulse.Length()
			angularImpulse = absFloat(jnt.motorImpulse + jnt.lowerImpulse - jnt.upperImpulse)
		}

	case WeldJointType:
		{
			jnt := &sim.weldJoint

			linearImpulse = jnt.linearImpulse.Length()
			angularImpulse = absFloat(jnt.angularImpulse)
		}

	case WheelJointType:
		{
			var (
				jnt          = &sim.wheelJoint
				perpImpulse  = jnt.perpImpulse
				axialImpulse = jnt.springImpulse + jnt.lowerImpulse - jnt.upperImpulse
			)

			linearImpulse = float32(math.Sqrt(float64(float32(perpImpulse*perpImpulse) + float32(axialImpulse*axialImpulse))))
			angularImpulse = absFloat(jnt.motorImpulse)
		}

	case FilterJointType:
		// Filter joints have no reaction impulse.

	default:
		break
	}

	force := float32(linearImpulse * invTimeStep)
	torque := float32(angularImpulse * invTimeStep)

	return force, torque
}

func getJointConstraintForce(world *World, jnt *joint) Vec2 {
	base := getJointSim(world, jnt)

	switch jnt.jointType {
	case DistanceJointType:
		return getDistanceJointForce(world, base)

	case MotorJointType:
		return getMotorJointForce(world, base)

	case MoverJointType:
		return getMoverJointForce(world, base)

	case FilterJointType:
		return Vec2Zero

	case PogoJointType:
		return getPogoJointForce(world, base)

	case PrismaticJointType:
		return getPrismaticJointForce(world, base)

	case RevoluteJointType:
		return getRevoluteJointForce(world, base)

	case WeldJointType:
		return getWeldJointForce(world, base)

	case WheelJointType:
		return getWheelJointForce(world, base)

	default:
		core.Assert(false, "false")
		return Vec2Zero
	}
}

func getJointConstraintTorque(world *World, jnt *joint) float32 {
	base := getJointSim(world, jnt)

	switch jnt.jointType {
	case DistanceJointType:
		return 0.0

	case MotorJointType:
		return getMotorJointTorque(world, base)

	case FilterJointType:
		return 0.0

	case PrismaticJointType:
		return getPrismaticJointTorque(world, base)

	case RevoluteJointType:
		return getRevoluteJointTorque(world, base)

	case WeldJointType:
		return getWeldJointTorque(world, base)

	case WheelJointType:
		return getWheelJointTorque(world, base)

	case MoverJointType, PogoJointType:
		// Preserve the C port's unsupported constraint-torque query.
		core.Assert(false, "false")
		return 0.0

	default:
		core.Assert(false, "false")
		return 0.0
	}
}

// Get the current constraint force for this joint. Usually in Newtons.
func (j Joint) ConstraintForce() Vec2 {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return getJointConstraintForce(world, jnt)
}

// Get the current constraint torque for this joint. Usually in Newton * meters.
func (j Joint) ConstraintTorque() float32 {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
	)

	return getJointConstraintTorque(world, jnt)
}

// Get the current linear separation error for this joint. Does not consider admissible movement. Usually in meters.
func (j Joint) LinearSeparation() float32 {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		base  = getJointSim(world, jnt)
	)

	// Relative to body A so the difference stays in float32 precision far from the origin
	var (
		wxfA = getBodyTransform(world, jnt.edges[0].bodyID)
		xfA  = wxfA.Relative(wxfA.P)
		xfB  = getBodyTransform(world, jnt.edges[1].bodyID).Relative(wxfA.P)
	)

	var (
		pA = xfA.Apply(base.localFrameA.P)
		pB = xfB.Apply(base.localFrameB.P)
		dp = pB.Sub(pA)
	)

	switch jnt.jointType {
	case DistanceJointType:
		return distanceLinearSeparation(&base.distanceJoint, dp.Length())

	case MotorJointType:
		return 0.0

	case FilterJointType:
		return 0.0

	case PrismaticJointType:
		joint := &base.prismaticJoint
		return axialLinearSeparation(xfA.Q.Rotate(Vec2{1.0, 0.0}), dp, joint.enableLimit, joint.lowerTranslation, joint.upperTranslation)

	case RevoluteJointType:
		return dp.Length()

	case WeldJointType:
		{
			weldJoint := &base.weldJoint
			if weldJoint.linearHertz == 0.0 {
				return dp.Length()
			}

			return 0.0
		}

	case WheelJointType:
		joint := &base.wheelJoint
		return axialLinearSeparation(xfA.Q.Rotate(Vec2{1.0, 0.0}), dp, joint.enableLimit, joint.lowerTranslation, joint.upperTranslation)

	case MoverJointType, PogoJointType:
		// Preserve the C port's unsupported linear-separation query.
		core.Assert(false, "false")
		return 0.0

	default:
		core.Assert(false, "false")
		return 0.0
	}
}

// Get the current angular separation error for this joint. Does not consider admissible movement. Usually in meters.
func (j Joint) AngularSeparation() float32 {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		base  = getJointSim(world, jnt)
	)

	var (
		qA            = getBodyTransform(world, jnt.edges[0].bodyID).Q
		qB            = getBodyTransform(world, jnt.edges[1].bodyID).Q
		relativeAngle = qA.RelativeAngle(qB)
	)

	switch jnt.jointType {
	case DistanceJointType:
		return 0.0

	case MotorJointType:
		return 0.0

	case FilterJointType:
		return 0.0

	case PrismaticJointType:
		{
			return relativeAngle
		}

	case RevoluteJointType:
		return revoluteAngularSeparation(&base.revoluteJoint, relativeAngle)

	case WeldJointType:
		{
			weldJoint := &base.weldJoint
			if weldJoint.angularHertz == 0.0 {
				return relativeAngle
			}

			return 0.0
		}

	case WheelJointType:
		return 0.0

	case MoverJointType, PogoJointType:
		// Preserve the C port's unsupported angular-separation query.
		core.Assert(false, "false")
		return 0.0

	default:
		core.Assert(false, "false")
		return 0.0
	}
}

// Set the joint constraint tuning. Advanced feature.
func (j Joint) SetConstraintTuning(hertz float32, dampingRatio float32) {
	checkInput(IsValidFloat(hertz) && hertz >= 0.0, "b2IsValidFloat( hertz ) && hertz >= 0.0f")
	checkInput(IsValidFloat(dampingRatio) && dampingRatio >= 0.0, "b2IsValidFloat( dampingRatio ) && dampingRatio >= 0.0f")

	world := j.world
	if world.recording != nil {
		recWriteJointSetConstraintTuning(world.recording, &recArgsJointSetConstraintTuning{joint: j, hertz: hertz, dampingRatio: dampingRatio})
	}

	var (
		jnt  = getJointFullID(world, j)
		base = getJointSim(world, jnt)
	)

	base.constraintHertz = hertz
	base.constraintDampingRatio = dampingRatio
}

// Get the joint constraint tuning. Advanced feature.
func (j Joint) ConstraintTuning() (float32, float32) {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		base  = getJointSim(world, jnt)
	)

	hertz := base.constraintHertz
	dampingRatio := base.constraintDampingRatio

	return hertz, dampingRatio
}

// Set the force threshold for joint events (Newtons).
func (j Joint) SetForceThreshold(threshold float32) {
	checkInput(IsValidFloat(threshold) && threshold >= 0.0, "b2IsValidFloat( threshold ) && threshold >= 0.0f")

	world := j.world
	if world.recording != nil {
		recWriteJointSetForceThreshold(world.recording, &recArgsJointSetForceThreshold{joint: j, threshold: threshold})
	}

	var (
		jnt  = getJointFullID(world, j)
		base = getJointSim(world, jnt)
	)

	base.forceThreshold = threshold
}

// Get the force threshold for joint events (Newtons).
func (j Joint) ForceThreshold() float32 {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		base  = getJointSim(world, jnt)
	)

	return base.forceThreshold
}

// Set the torque threshold for joint events (N-m).
func (j Joint) SetTorqueThreshold(threshold float32) {
	checkInput(IsValidFloat(threshold) && threshold >= 0.0, "b2IsValidFloat( threshold ) && threshold >= 0.0f")

	world := j.world
	if world.recording != nil {
		recWriteJointSetTorqueThreshold(world.recording, &recArgsJointSetTorqueThreshold{joint: j, threshold: threshold})
	}

	var (
		jnt  = getJointFullID(world, j)
		base = getJointSim(world, jnt)
	)

	base.torqueThreshold = threshold
}

// Get the torque threshold for joint events (N-m).
func (j Joint) TorqueThreshold() float32 {
	var (
		world = j.world
		jnt   = getJointFullID(world, j)
		base  = getJointSim(world, jnt)
	)

	return base.torqueThreshold
}

func prepareJoint(jnt *jointSim, context *stepContext) {
	// Clamp joint hertz based on the time step to reduce jitter.
	hertz := minFloat(jnt.constraintHertz, float32(0.25*context.invH))

	jnt.constraintSoftness = makeSoft(hertz, jnt.constraintDampingRatio, context.h)

	switch jnt.jointType {
	case DistanceJointType:
		prepareDistanceJoint(jnt, context)

	case MotorJointType:
		prepareMotorJoint(jnt, context)

	case MoverJointType:
		prepareMoverJoint(jnt, context)

	case FilterJointType:
		break

	case PogoJointType:
		preparePogoJoint(jnt, context)

	case PrismaticJointType:
		preparePrismaticJoint(jnt, context)

	case RevoluteJointType:
		prepareRevoluteJoint(jnt, context)

	case WeldJointType:
		prepareWeldJoint(jnt, context)

	case WheelJointType:
		prepareWheelJoint(jnt, context)

	default:
		core.Assert(false, "false")
	}
}

func warmStartJoint(jnt *jointSim, context *stepContext) {
	switch jnt.jointType {
	case DistanceJointType:
		warmStartDistanceJoint(jnt, context)

	case MotorJointType:
		warmStartMotorJoint(jnt, context)

	case MoverJointType:
		warmStartMoverJoint(jnt, context)

	case FilterJointType:
		break

	case PogoJointType:
		warmStartPogoJoint(jnt, context)

	case PrismaticJointType:
		warmStartPrismaticJoint(jnt, context)

	case RevoluteJointType:
		warmStartRevoluteJoint(jnt, context)

	case WeldJointType:
		warmStartWeldJoint(jnt, context)

	case WheelJointType:
		warmStartWheelJoint(jnt, context)

	default:
		core.Assert(false, "false")
	}
}

func solveJoint(jnt *jointSim, context *stepContext, useBias bool) {
	switch jnt.jointType {
	case DistanceJointType:
		solveDistanceJoint(jnt, context, useBias)

	case MotorJointType:
		solveMotorJoint(jnt, context)

	case MoverJointType:
		solveMoverJoint(jnt, context)

	case FilterJointType:
		break

	case PogoJointType:
		solvePogoJoint(jnt, context, useBias)

	case PrismaticJointType:
		solvePrismaticJoint(jnt, context, useBias)

	case RevoluteJointType:
		solveRevoluteJoint(jnt, context, useBias)

	case WeldJointType:
		solveWeldJoint(jnt, context, useBias)

	case WheelJointType:
		solveWheelJoint(jnt, context, useBias)

	default:
		core.Assert(false, "false")
	}
}

func prepareJointsOverflow(context *stepContext) {
	var (
		graph      = context.graph
		joints     = graph.colors[overflowIndex].jointSims
		jointCount = len(graph.colors[overflowIndex].jointSims)
	)

	for i := range jointCount {
		jnt := &joints[i]
		prepareJoint(jnt, context)
	}
}

func warmStartJointsOverflow(context *stepContext) {
	var (
		graph      = context.graph
		joints     = graph.colors[overflowIndex].jointSims
		jointCount = len(graph.colors[overflowIndex].jointSims)
	)

	for i := range jointCount {
		jnt := &joints[i]
		warmStartJoint(jnt, context)
	}
}

func solveJointsOverflow(context *stepContext, useBias bool) {
	var (
		graph      = context.graph
		joints     = graph.colors[overflowIndex].jointSims
		jointCount = len(graph.colors[overflowIndex].jointSims)
	)

	for i := range jointCount {
		jnt := &joints[i]
		solveJoint(jnt, context, useBias)
	}
}

func prepareJointsTask(block solverBlock, context *stepContext) {
	spans := context.jointPrepareSpans

	var (
		index    = block.startIndex
		endIndex = block.startIndex + int(block.count)
	)

	// Find color for start index. Linear search but fast.
	colorIndex := 0
	for spans[colorIndex+1].start <= index {
		colorIndex += 1
	}

	// Loop over block
	for index < endIndex {
		var (
			colorStart    = spans[colorIndex].start
			colorEndIndex = min(spans[colorIndex+1].start, endIndex)
			joints        = spans[colorIndex].joints
		)

		// Loop over color
		for ; index < colorEndIndex; index++ {
			core.Assert(0 <= index-colorStart && index-colorStart < spans[colorIndex].count, "0 <= index - colorStart && index - colorStart < spans[colorIndex].count")

			jnt := &joints[(index - colorStart)]
			prepareJoint(jnt, context)
		}

		// Advance to next color
		colorIndex += 1
	}
}

func warmStartJointsTask(block solverBlock, context *stepContext) {
	var (
		color  = &context.graph.colors[block.colorIndex]
		joints = color.jointSims
	)

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		jnt := &joints[i]
		warmStartJoint(jnt, context)
	}
}

func solveJointsTask(block solverBlock, context *stepContext, useBias bool, workerIndex int) {
	var (
		color  = &context.graph.colors[block.colorIndex]
		joints = color.jointSims
	)

	core.Assert(0 <= block.startIndex && block.startIndex+int(block.count) <= len(color.jointSims), "0 <= block.startIndex && block.startIndex + block.count <= color->jointSims.count")

	jointStateBitSet := &context.world.taskContexts[workerIndex].jointStateBitSet

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		jnt := &joints[i]
		solveJoint(jnt, context, useBias)

		if useBias && (jnt.forceThreshold < math.MaxFloat32 || jnt.torqueThreshold < math.MaxFloat32) &&
			!jointStateBitSet.GetBit(uint32(jnt.jointID)) {
			var force, torque float32

			force, torque = getJointReaction(jnt, context.invH)

			// Check thresholds. A zero threshold means all awake joints get reported.
			if force >= jnt.forceThreshold || torque >= jnt.torqueThreshold {
				// Flag this joint for processing.
				jointStateBitSet.SetBit(uint32(jnt.jointID))
			}
		}
	}
}

func drawJoint(draw *DebugDraw, world *World, jnt *joint) {
	var (
		bodyA = &world.bodies[jnt.edges[0].bodyID]
		bodyB = &world.bodies[jnt.edges[1].bodyID]
	)
	if bodyA.setIndex == disabledSet || bodyB.setIndex == disabledSet {
		return
	}

	sim := getJointSim(world, jnt)

	var (
		xfA = getBodyTransformQuick(world, bodyA)
		xfB = getBodyTransformQuick(world, bodyB)
	)

	pA, pB := drawJointGeometry(draw, jnt, sim, xfA, xfB)

	if draw.DrawGraphColors {
		colorIndex := jnt.colorIndex
		if colorIndex != nullIndex {
			p := pA.Lerp(pB, 0.5)
			draw.DrawPoint(p, 5.0, GraphColor(colorIndex))
		}
	}

	if draw.DrawJointExtras {
		var (
			force  = getJointConstraintForce(world, jnt)
			torque = getJointConstraintTorque(world, jnt)
			p      = pA.Lerp(pB, 0.5)
		)

		draw.DrawLine(p, p.Add(force.Scale(0.001)), ColorAzure)

		buffer := fmt.Sprintf("f = [%.6g, %.6g], t = %.6g", force.X, force.Y, torque)
		draw.DrawString(p, buffer, ColorAzure)
	}
}

func placeJointInSolverSet(world *World, jnt *joint, bodyA, bodyB *body, maxSetIndex int) *jointSim {
	jointID := jnt.jointID
	bodyIDA := bodyA.id
	bodyIDB := bodyB.id

	var sim *jointSim

	if bodyA.setIndex == disabledSet || bodyB.setIndex == disabledSet {
		// if either body is disabled, create in disabled set
		set := &world.solverSets[disabledSet]

		jnt.setIndex = disabledSet
		jnt.localIndex = len(set.jointSims)

		set.jointSims, sim = array.Emplace(set.jointSims)
		*sim = jointSim{}

		sim.jointID = jointID
		sim.bodyIDA = bodyIDA
		sim.bodyIDB = bodyIDB
	} else if bodyA.bodyType != DynamicBody && bodyB.bodyType != DynamicBody {
		// joint is not attached to a dynamic body
		set := &world.solverSets[staticSet]

		jnt.setIndex = staticSet
		jnt.localIndex = len(set.jointSims)

		set.jointSims, sim = array.Emplace(set.jointSims)
		*sim = jointSim{}

		sim.jointID = jointID
		sim.bodyIDA = bodyIDA
		sim.bodyIDB = bodyIDB
	} else if bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet {
		// if either body is sleeping, wake it
		if maxSetIndex >= firstSleepingSet {
			wakeSolverSet(world, maxSetIndex)
		}

		jnt.setIndex = awakeSet

		sim = createJointInGraph(world, jnt)
		sim.jointID = jointID
		sim.bodyIDA = bodyIDA
		sim.bodyIDB = bodyIDB
	} else {
		// joint connected between sleeping and/or static bodies
		core.Assert(bodyA.setIndex >= firstSleepingSet || bodyB.setIndex >= firstSleepingSet, "bodyA->setIndex >= b2_firstSleepingSet || bodyB->setIndex >= b2_firstSleepingSet")
		core.Assert(bodyA.setIndex != staticSet || bodyB.setIndex != staticSet, "bodyA->setIndex != b2_staticSet || bodyB->setIndex != b2_staticSet")

		// joint should go into the sleeping set (not static set)
		setIndex := maxSetIndex

		set := &world.solverSets[setIndex]

		jnt.setIndex = setIndex
		jnt.localIndex = len(set.jointSims)

		set.jointSims, sim = array.Emplace(set.jointSims)
		*sim = jointSim{}

		// These must be set to accommodate the merge below
		sim.jointID = jointID
		sim.bodyIDA = bodyIDA
		sim.bodyIDB = bodyIDB

		if bodyA.setIndex != bodyB.setIndex && bodyA.setIndex >= firstSleepingSet &&
			bodyB.setIndex >= firstSleepingSet {
			// merge sleeping sets
			mergeSolverSets(world, bodyA.setIndex, bodyB.setIndex)
			core.Assert(bodyA.setIndex == bodyB.setIndex, "bodyA->setIndex == bodyB->setIndex")

			// fix potentially invalid set index
			setIndex = bodyA.setIndex

			mergedSet := &world.solverSets[setIndex]

			// Careful! The joint sim pointer was orphaned by the set merge.
			sim = &mergedSet.jointSims[jnt.localIndex]
		}

		core.Assert(jnt.setIndex == setIndex, "joint->setIndex == setIndex")
	}

	return sim
}

func distanceLinearSeparation(joint *distanceJoint, length float32) float32 {
	if !joint.enableSpring {
		return absFloat(length - joint.length)
	}

	if !joint.enableLimit {
		return 0.0
	}

	if length < joint.minLength {
		return joint.minLength - length
	}

	if length > joint.maxLength {
		return length - joint.maxLength
	}

	return 0.0
}

func axialLinearSeparation(axisA, dp Vec2, enableLimit bool, lowerTranslation, upperTranslation float32) float32 {
	perpA := axisA.LeftPerp()
	perpendicularSeparation := absFloat(perpA.Dot(dp))
	limitSeparation := float32(0.0)

	if enableLimit {
		translation := axisA.Dot(dp)
		if translation < lowerTranslation {
			limitSeparation = lowerTranslation - translation
		}

		if upperTranslation < translation {
			limitSeparation = translation - upperTranslation
		}
	}

	return float32(math.Sqrt(float64(float32(perpendicularSeparation*perpendicularSeparation) + float32(limitSeparation*limitSeparation))))
}

func revoluteAngularSeparation(joint *revoluteJoint, angle float32) float32 {
	if joint.enableLimit {
		if angle < joint.lowerAngle {
			return joint.lowerAngle - angle
		}

		if joint.upperAngle < angle {
			return angle - joint.upperAngle
		}
	}

	return 0.0
}

func jointSolverBodyState(context *stepContext, index int, dummyState *bodyState) *bodyState {
	if index == nullIndex {
		return dummyState
	}

	return &context.states[index]
}

func drawJointGeometry(draw *DebugDraw, jnt *joint, sim *jointSim, xfA, xfB WorldTransform) (Pos, Pos) {
	var (
		pA = xfA.Apply(sim.localFrameA.P)
		pB = xfB.Apply(sim.localFrameB.P)
	)

	scale := maxFloat(0.0001, float32(draw.JointScale*jnt.drawScale))

	switch jnt.jointType {
	case DistanceJointType:
		drawDistanceJoint(draw, sim, xfA, xfB)

	case FilterJointType:
		draw.DrawPoint(pA, 8.0, ColorLightSkyBlue)
		draw.DrawPoint(pB, 8.0, ColorLightSkyBlue)

		if draw.DrawJointExtras {
			draw.DrawLine(pA, pB, ColorGold)
		}

	case MotorJointType:
		draw.DrawPoint(pA, 8.0, ColorYellowGreen)
		draw.DrawPoint(pB, 8.0, ColorPlum)

		if draw.DrawJointExtras {
			draw.DrawLine(pA, pB, ColorLightGray)
		}

	case MoverJointType:
		draw.DrawPoint(pB, 8.0, ColorPlum)

	case PogoJointType:
		draw.DrawLine(pA, pB, ColorLightGray)

	case PrismaticJointType:
		drawPrismaticJoint(draw, sim, xfA, xfB, scale)

	case RevoluteJointType:
		drawRevoluteJoint(draw, sim, xfA, xfB, scale)

	case WeldJointType:
		drawWeldJoint(draw, sim, xfA, xfB, scale)

	case WheelJointType:
		drawWheelJoint(draw, sim, xfA, xfB, scale)

	default:
		break
	}

	return pA, pB
}
