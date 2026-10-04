// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"math"
)

type JointType int32

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

// Base joint definition used by all joint types.
// The local frames are measured from the body's origin rather than the center of mass because:
// 1. you might not know where the center of mass will be
// 2. if you add/remove shapes from a body and recompute the mass, the joints will be broken
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
// Note: this data becomes invalid if joints are destroyed
type JointEvents struct {
	// Array of events
	JointEvents []JointEvent
	// Number of events
	Count int
}

type jointEdge struct {
	bodyId  int
	prevKey int
	nextKey int
}

type joint struct {
	userData         any
	setIndex         int
	colorIndex       int
	localIndex       int
	edges            [2]jointEdge
	jointId          int
	islandId         int
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
	jointId                int
	bodyIdA                int
	bodyIdB                int
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

// Use this to initialize your joint definition
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

// Use this to initialize your joint definition
func DefaultFilterJointDef() FilterJointDef {
	return FilterJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultMotorJointDef() MotorJointDef {
	return MotorJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultMoverJointDef() MoverJointDef {
	return MoverJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultPogoJointDef() PogoJointDef {
	return PogoJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultPrismaticJointDef() PrismaticJointDef {
	return PrismaticJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition.
func DefaultRevoluteJointDef() RevoluteJointDef {
	return RevoluteJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultWeldJointDef() WeldJointDef {
	return WeldJointDef{Base: defaultJointDef(), internalValue: secretCookie}
}

// Use this to initialize your joint definition
func DefaultWheelJointDef() WheelJointDef {
	return WheelJointDef{Base: defaultJointDef(), internalValue: secretCookie, EnableSpring: true, Hertz: 1, DampingRatio: 0.7}
}

// Use this to initialize your explosion definition
func DefaultExplosionDef() ExplosionDef {
	return ExplosionDef{MaskBits: DefaultMaskBits}
}

func getJointFullId(world *World, jointID Joint) *joint {
	var id int = int(jointID.index1) - 1
	var jnt *joint = &world.joints[id]
	assert(jnt.jointId == id && jnt.generation == jointID.generation, "joint->jointId == id && joint->generation == jointId.generation")
	return jnt
}

func getJointSim(world *World, jnt *joint) *jointSim {
	if jnt.setIndex == awakeSet {
		assert(0 <= jnt.colorIndex && jnt.colorIndex < GraphColorCount, "0 <= joint->colorIndex && joint->colorIndex < B2_GRAPH_COLOR_COUNT")
		var color *graphColor = &world.constraintGraph.colors[jnt.colorIndex]
		return &color.jointSims[jnt.localIndex]
	}

	var set *solverSet = &world.solverSets[jnt.setIndex]
	return &set.jointSims[jnt.localIndex]
}

func getJointSimCheckType(jointID Joint, jointType JointType) *jointSim {
	var world *World = jointID.world
	assert(world.locked == false, "world->locked == false")
	if world.locked {
		return nil
	}

	var jnt *joint = getJointFullId(world, jointID)
	assert(jnt.jointType == jointType, "joint->type == type")
	var sim *jointSim = getJointSim(world, jnt)
	assert(sim.jointType == jointType, "jointSim->type == type")
	return sim
}

func destroyContactsBetweenBodies(world *World, bodyA *body, bodyB *body) {
	var contactKey int
	var otherBodyId int

	// use the smaller of the two contact lists
	if bodyA.contactCount < bodyB.contactCount {
		contactKey = bodyA.headContactKey
		otherBodyId = bodyB.id
	} else {
		contactKey = bodyB.headContactKey
		otherBodyId = bodyA.id
	}

	// destroy the contacts
	for contactKey != nullIndex {
		var contactId int = contactKey >> 1
		var edgeIndex int = contactKey & 1

		var contact *contact = &world.contacts[contactId]
		contactKey = contact.edges[edgeIndex].nextKey

		var otherEdgeIndex int = edgeIndex ^ 1
		if contact.edges[otherEdgeIndex].bodyID == otherBodyId {
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
	assert(def.LocalFrameA.IsValid(), "b2IsValidTransform( def->localFrameA )")
	assert(def.LocalFrameB.IsValid(), "b2IsValidTransform( def->localFrameB )")
	assert(world == def.BodyA.world, "world->worldId == def->bodyIdA.world")
	assert(world == def.BodyB.world, "world->worldId == def->bodyIdB.world")
	assert((def.BodyA == def.BodyB) == false, "B2_ID_EQUALS( def->bodyIdA, def->bodyIdB ) == false")

	var bodyA *body = getBodyFullId(world, def.BodyA)
	var bodyB *body = getBodyFullId(world, def.BodyB)

	// If the joint prevents collisions, then destroy all contacts between attached bodies.
	// Do this first so world arrays are not disturbed in the following code.
	if def.CollideConnected == false {
		destroyContactsBetweenBodies(world, bodyA, bodyB)
	}

	var bodyIdA int = bodyA.id
	var bodyIdB int = bodyB.id
	var maxSetIndex int = max(bodyA.setIndex, bodyB.setIndex)

	// Create joint id and joint
	var jointId int = allocID(&world.jointIDPool)
	if jointId == len(world.joints) {
		world.joints = arrayPush(world.joints, joint{})
	}

	var jnt *joint = &world.joints[jointId]
	jnt.jointId = jointId
	jnt.userData = def.UserData
	jnt.generation += 1
	jnt.setIndex = nullIndex
	jnt.colorIndex = nullIndex
	jnt.localIndex = nullIndex
	jnt.islandId = nullIndex
	jnt.islandIndex = nullIndex
	jnt.drawScale = def.DrawScale
	jnt.jointType = jointType
	jnt.collideConnected = def.CollideConnected

	// Doubly linked list on bodyA
	jnt.edges[0].bodyId = bodyIdA
	jnt.edges[0].prevKey = nullIndex
	jnt.edges[0].nextKey = bodyA.headJointKey

	var keyA int = (jointId << 1) | 0
	if bodyA.headJointKey != nullIndex {
		var jointA *joint = &world.joints[bodyA.headJointKey>>1]
		var edgeA *jointEdge = &jointA.edges[(bodyA.headJointKey & 1)]
		edgeA.prevKey = keyA
	}
	bodyA.headJointKey = keyA
	bodyA.jointCount += 1

	// Doubly linked list on bodyB
	jnt.edges[1].bodyId = bodyIdB
	jnt.edges[1].prevKey = nullIndex
	jnt.edges[1].nextKey = bodyB.headJointKey

	var keyB int = (jointId << 1) | 1
	if bodyB.headJointKey != nullIndex {
		var jointB *joint = &world.joints[bodyB.headJointKey>>1]
		var edgeB *jointEdge = &jointB.edges[(bodyB.headJointKey & 1)]
		edgeB.prevKey = keyB
	}
	bodyB.headJointKey = keyB
	bodyB.jointCount += 1

	var sim *jointSim

	if bodyA.setIndex == disabledSet || bodyB.setIndex == disabledSet {
		// if either body is disabled, create in disabled set
		var set *solverSet = &world.solverSets[disabledSet]
		jnt.setIndex = disabledSet
		jnt.localIndex = len(set.jointSims)

		set.jointSims, sim = arrayEmplace(set.jointSims)
		*sim = jointSim{}

		sim.jointId = jointId
		sim.bodyIdA = bodyIdA
		sim.bodyIdB = bodyIdB
	} else if bodyA.bodyType != DynamicBody && bodyB.bodyType != DynamicBody {
		// joint is not attached to a dynamic body
		var set *solverSet = &world.solverSets[staticSet]
		jnt.setIndex = staticSet
		jnt.localIndex = len(set.jointSims)

		set.jointSims, sim = arrayEmplace(set.jointSims)
		*sim = jointSim{}

		sim.jointId = jointId
		sim.bodyIdA = bodyIdA
		sim.bodyIdB = bodyIdB
	} else if bodyA.setIndex == awakeSet || bodyB.setIndex == awakeSet {
		// if either body is sleeping, wake it
		if maxSetIndex >= firstSleepingSet {
			wakeSolverSet(world, maxSetIndex)
		}

		jnt.setIndex = awakeSet

		sim = createJointInGraph(world, jnt)
		sim.jointId = jointId
		sim.bodyIdA = bodyIdA
		sim.bodyIdB = bodyIdB
	} else {
		// joint connected between sleeping and/or static bodies
		assert(bodyA.setIndex >= firstSleepingSet || bodyB.setIndex >= firstSleepingSet, "bodyA->setIndex >= b2_firstSleepingSet || bodyB->setIndex >= b2_firstSleepingSet")
		assert(bodyA.setIndex != staticSet || bodyB.setIndex != staticSet, "bodyA->setIndex != b2_staticSet || bodyB->setIndex != b2_staticSet")

		// joint should go into the sleeping set (not static set)
		var setIndex int = maxSetIndex

		var set *solverSet = &world.solverSets[setIndex]
		jnt.setIndex = setIndex
		jnt.localIndex = len(set.jointSims)

		set.jointSims, sim = arrayEmplace(set.jointSims)
		*sim = jointSim{}

		// These must be set to accommodate the merge below
		sim.jointId = jointId
		sim.bodyIdA = bodyIdA
		sim.bodyIdB = bodyIdB

		if bodyA.setIndex != bodyB.setIndex && bodyA.setIndex >= firstSleepingSet &&
			bodyB.setIndex >= firstSleepingSet {
			// merge sleeping sets
			mergeSolverSets(world, bodyA.setIndex, bodyB.setIndex)
			assert(bodyA.setIndex == bodyB.setIndex, "bodyA->setIndex == bodyB->setIndex")

			// fix potentially invalid set index
			setIndex = bodyA.setIndex

			var mergedSet *solverSet = &world.solverSets[setIndex]

			// Careful! The joint sim pointer was orphaned by the set merge.
			sim = &mergedSet.jointSims[jnt.localIndex]
		}

		assert(jnt.setIndex == setIndex, "joint->setIndex == setIndex")
	}

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

	assert(IsValidFloat(def.ForceThreshold) && def.ForceThreshold >= 0.0, "b2IsValidFloat( def->forceThreshold ) && def->forceThreshold >= 0.0f")
	assert(IsValidFloat(def.TorqueThreshold) && def.TorqueThreshold >= 0.0, "b2IsValidFloat( def->torqueThreshold ) && def->torqueThreshold >= 0.0f")

	sim.forceThreshold = def.ForceThreshold
	sim.torqueThreshold = def.TorqueThreshold

	assert(sim.jointId == jointId, "jointSim->jointId == jointId")
	assert(sim.bodyIdA == bodyIdA, "jointSim->bodyIdA == bodyIdA")
	assert(sim.bodyIdB == bodyIdB, "jointSim->bodyIdB == bodyIdB")

	if jnt.setIndex > disabledSet {
		// Add edge to island graph
		linkJoint(world, jnt)
	}

	validateSolverSets(world)

	return jointPair{jnt, sim}
}

// Create a distance joint
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
	assert(def.LowerSpringForce <= def.UpperSpringForce, "def->lowerSpringForce <= def->upperSpringForce")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return DistanceJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, DistanceJointType)

	var jnt *jointSim = pair.jointSim

	var empty distanceJoint = distanceJoint{}
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

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateDistanceJoint(world.recording, &recArgsCreateDistanceJoint{def: *def}, jointId)
	}

	return DistanceJoint{Joint: jointId}
}

// Create a filter joint.
func (w *World) CreateFilterJoint(def *FilterJointDef) FilterJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return FilterJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, FilterJointType)

	var jnt *jointSim = pair.jointSim

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateFilterJoint(world.recording, &recArgsCreateFilterJoint{def: *def}, jointId)
	}

	return FilterJoint{Joint: jointId}
}

// Create a motor joint
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

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return MotorJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, MotorJointType)
	var jnt *jointSim = pair.jointSim

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

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateMotorJoint(world.recording, &recArgsCreateMotorJoint{def: *def}, jointId)
	}

	return MotorJoint{Joint: jointId}
}

// Create a mover joint
func (w *World) CreateMoverJoint(def *MoverJointDef) MoverJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(def.LinearVelocity.IsValid(), "b2IsValidVec2( def->linearVelocity )")
	checkInput(def.MaxVelocityForce.IsValid(), "b2IsValidVec2( def->maxVelocityForce )")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return MoverJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, MoverJointType)
	var jnt *jointSim = pair.jointSim

	jnt.moverJoint = moverJoint{}
	jnt.moverJoint.linearVelocity = def.LinearVelocity
	jnt.moverJoint.maxVelocityForce = def.MaxVelocityForce

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateMoverJoint(world.recording, &recArgsCreateMoverJoint{def: *def}, jointId)
	}

	return MoverJoint{Joint: jointId}
}

// Create a pog joint
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

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return PogoJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, PogoJointType)

	var jnt *jointSim = pair.jointSim

	jnt.pogoJoint = pogoJoint{}
	jnt.pogoJoint.restLength = def.RestLength
	jnt.pogoJoint.hertz = def.Hertz
	jnt.pogoJoint.dampingRatio = def.DampingRatio
	jnt.pogoJoint.maxTensionForce = def.MaxTensionForce
	jnt.pogoJoint.maxCompressionForce = def.MaxCompressionForce
	jnt.pogoJoint.normal = def.Normal
	jnt.pogoJoint.impulse = def.Impulse
	jnt.pogoJoint.velocity = def.Velocity

	var jointId Joint = Joint{
		index1:     int32(jnt.jointId + 1),
		world:      world,
		generation: pair.joint.generation,
	}

	if world.recording != nil {
		recWriteRetCreatePogoJoint(world.recording, &recArgsCreatePogoJoint{def: *def}, jointId)
	}

	return PogoJoint{Joint: jointId}
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
	assert(def.LowerTranslation <= def.UpperTranslation, "def->lowerTranslation <= def->upperTranslation")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return PrismaticJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, PrismaticJointType)

	var jnt *jointSim = pair.jointSim

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

	var jointId Joint = Joint{
		index1:     int32(jnt.jointId + 1),
		world:      world,
		generation: pair.joint.generation,
	}

	if world.recording != nil {
		recWriteRetCreatePrismaticJoint(world.recording, &recArgsCreatePrismaticJoint{def: *def}, jointId)
	}

	return PrismaticJoint{Joint: jointId}
}

// Create a revolute joint
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
	assert(def.LowerAngle <= def.UpperAngle, "def->lowerAngle <= def->upperAngle")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return RevoluteJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, RevoluteJointType)

	var jnt *jointSim = pair.jointSim

	var empty revoluteJoint = revoluteJoint{}
	jnt.revoluteJoint = empty

	jnt.revoluteJoint.targetAngle = clampFloat(def.TargetAngle, -Pi, Pi)
	jnt.revoluteJoint.hertz = def.Hertz
	jnt.revoluteJoint.dampingRatio = def.DampingRatio

	var lowerAngle float32 = minFloat(def.LowerAngle, def.UpperAngle)
	var upperAngle float32 = maxFloat(def.LowerAngle, def.UpperAngle)
	jnt.revoluteJoint.lowerAngle = clampFloat(lowerAngle, float32(-0.99*Pi), float32(0.99*Pi))
	jnt.revoluteJoint.upperAngle = clampFloat(upperAngle, float32(-0.99*Pi), float32(0.99*Pi))

	jnt.revoluteJoint.maxMotorTorque = def.MaxMotorTorque
	jnt.revoluteJoint.motorSpeed = def.MotorSpeed
	jnt.revoluteJoint.enableSpring = def.EnableSpring
	jnt.revoluteJoint.enableLimit = def.EnableLimit
	jnt.revoluteJoint.enableMotor = def.EnableMotor

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateRevoluteJoint(world.recording, &recArgsCreateRevoluteJoint{def: *def}, jointId)
	}

	return RevoluteJoint{Joint: jointId}
}

// Create a weld joint
func (w *World) CreateWeldJoint(def *WeldJointDef) WeldJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.LinearHertz), "b2IsValidFloat( def->linearHertz )")
	checkInput(IsValidFloat(def.AngularHertz), "b2IsValidFloat( def->angularHertz )")
	checkInput(IsValidFloat(def.LinearDampingRatio), "b2IsValidFloat( def->linearDampingRatio )")
	checkInput(IsValidFloat(def.AngularDampingRatio), "b2IsValidFloat( def->angularDampingRatio )")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return WeldJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, WeldJointType)

	var jnt *jointSim = pair.jointSim

	var empty weldJoint = weldJoint{}
	jnt.weldJoint = empty
	jnt.weldJoint.linearHertz = def.LinearHertz
	jnt.weldJoint.linearDampingRatio = def.LinearDampingRatio
	jnt.weldJoint.angularHertz = def.AngularHertz
	jnt.weldJoint.angularDampingRatio = def.AngularDampingRatio
	jnt.weldJoint.linearImpulse = Vec2Zero
	jnt.weldJoint.angularImpulse = 0.0

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateWeldJoint(world.recording, &recArgsCreateWeldJoint{def: *def}, jointId)
	}

	return WeldJoint{Joint: jointId}
}

// Create a wheel joint
func (w *World) CreateWheelJoint(def *WheelJointDef) WheelJoint {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(isValidJointDef(&def.Base), "b2IsValidJointDef( &def->base )")
	checkInput(IsValidFloat(def.Hertz), "b2IsValidFloat( def->hertz )")
	checkInput(IsValidFloat(def.DampingRatio), "b2IsValidFloat( def->dampingRatio )")
	checkInput(IsValidFloat(def.LowerTranslation), "b2IsValidFloat( def->lowerTranslation )")
	checkInput(IsValidFloat(def.UpperTranslation), "b2IsValidFloat( def->upperTranslation )")
	checkInput(IsValidFloat(def.MaxMotorTorque), "b2IsValidFloat( def->maxMotorTorque )")
	checkInput(IsValidFloat(def.MotorSpeed), "b2IsValidFloat( def->motorSpeed )")
	assert(def.LowerTranslation <= def.UpperTranslation, "def->lowerTranslation <= def->upperTranslation")

	var world *World = w

	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return WheelJoint{}
	}

	var pair jointPair = createJoint(world, &def.Base, WheelJointType)

	var jnt *jointSim = pair.jointSim

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

	var jointId Joint = Joint{int32(jnt.jointId + 1), world, pair.joint.generation}

	if world.recording != nil {
		recWriteRetCreateWheelJoint(world.recording, &recArgsCreateWheelJoint{def: *def}, jointId)
	}

	return WheelJoint{Joint: jointId}
}

func destroyJointInternal(world *World, jnt *joint) {
	var jointId int = jnt.jointId

	var edgeA *jointEdge = &jnt.edges[0]
	var edgeB *jointEdge = &jnt.edges[1]

	var idA int = edgeA.bodyId
	var idB int = edgeB.bodyId
	var bodyA *body = &world.bodies[idA]
	var bodyB *body = &world.bodies[idB]

	// Remove from body A
	if edgeA.prevKey != nullIndex {
		var prevJoint *joint = &world.joints[edgeA.prevKey>>1]
		var prevEdge *jointEdge = &prevJoint.edges[(edgeA.prevKey & 1)]
		prevEdge.nextKey = edgeA.nextKey
	}

	if edgeA.nextKey != nullIndex {
		var nextJoint *joint = &world.joints[edgeA.nextKey>>1]
		var nextEdge *jointEdge = &nextJoint.edges[(edgeA.nextKey & 1)]
		nextEdge.prevKey = edgeA.prevKey
	}

	var edgeKeyA int = (jointId << 1) | 0
	if bodyA.headJointKey == edgeKeyA {
		bodyA.headJointKey = edgeA.nextKey
	}

	bodyA.jointCount -= 1

	// Remove from body B
	if edgeB.prevKey != nullIndex {
		var prevJoint *joint = &world.joints[edgeB.prevKey>>1]
		var prevEdge *jointEdge = &prevJoint.edges[(edgeB.prevKey & 1)]
		prevEdge.nextKey = edgeB.nextKey
	}

	if edgeB.nextKey != nullIndex {
		var nextJoint *joint = &world.joints[edgeB.nextKey>>1]
		var nextEdge *jointEdge = &nextJoint.edges[(edgeB.nextKey & 1)]
		nextEdge.prevKey = edgeB.prevKey
	}

	var edgeKeyB int = (jointId << 1) | 1
	if bodyB.headJointKey == edgeKeyB {
		bodyB.headJointKey = edgeB.nextKey
	}

	bodyB.jointCount -= 1

	if jnt.islandId != nullIndex {
		assert(jnt.setIndex > disabledSet, "joint->setIndex > b2_disabledSet")
		unlinkJoint(world, jnt)
	} else {
		assert(jnt.setIndex <= disabledSet, "joint->setIndex <= b2_disabledSet")
	}

	// Remove joint from solver set that owns it
	var setIndex int = jnt.setIndex
	var localIndex int = jnt.localIndex

	if setIndex == awakeSet {
		removeJointFromGraph(world, jnt.edges[0].bodyId, jnt.edges[1].bodyId, jnt.colorIndex, localIndex)
	} else {
		var set *solverSet = &world.solverSets[setIndex]
		var movedIndex int
		set.jointSims, movedIndex = arrayRemoveSwap(set.jointSims, localIndex)
		if movedIndex != nullIndex {
			// Fix moved joint
			var movedJointSim *jointSim = &set.jointSims[localIndex]
			var movedId int = movedJointSim.jointId
			var movedJoint *joint = &world.joints[movedId]
			assert(movedJoint.localIndex == movedIndex, "movedJoint->localIndex == movedIndex")
			movedJoint.localIndex = localIndex
		}
	}

	// Free joint and id (preserve joint generation)
	jnt.setIndex = nullIndex
	jnt.localIndex = nullIndex
	jnt.colorIndex = nullIndex
	jnt.jointId = nullIndex
	freeID(&world.jointIDPool, jointId)

	wakeBody(world, bodyA)
	wakeBody(world, bodyB)

	validateSolverSets(world)
}

// Destroy a joint.
func (j Joint) Destroy() {
	var world *World = j.world
	assert(world.locked == false, "world->locked == false")

	if world.locked {
		return
	}

	if world.recording != nil {
		recWriteDestroyJoint(world.recording, &recArgsDestroyJoint{joint: j})
	}

	var jnt *joint = getJointFullId(world, j)

	destroyJointInternal(world, jnt)
}

// Get the joint type
func (j Joint) Type() JointType {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return jnt.jointType
}

// Get body A id on a joint
func (j Joint) BodyA() Body {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return makeBodyId(world, jnt.edges[0].bodyId)
}

// Get body B id on a joint
func (j Joint) BodyB() Body {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return makeBodyId(world, jnt.edges[1].bodyId)
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

// Get the world that owns this joint
func (j Joint) World() *World {
	return j.world
}

// Set the local frame on body A
func (j Joint) SetLocalFrameA(localFrame Transform) {
	checkInput(localFrame.IsValid(), "b2IsValidTransform( localFrame )")

	var world *World = j.world
	if world.recording != nil {
		recWriteJointSetLocalFrameA(world.recording, &recArgsJointSetLocalFrameA{joint: j, localFrame: localFrame})
	}
	var jnt *joint = getJointFullId(world, j)
	var sim *jointSim = getJointSim(world, jnt)
	sim.localFrameA = localFrame
}

// Get the local frame on body A
func (j Joint) LocalFrameA() Transform {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var sim *jointSim = getJointSim(world, jnt)
	return sim.localFrameA
}

// Set the local frame on body B
func (j Joint) SetLocalFrameB(localFrame Transform) {
	checkInput(localFrame.IsValid(), "b2IsValidTransform( localFrame )")

	var world *World = j.world
	if world.recording != nil {
		recWriteJointSetLocalFrameB(world.recording, &recArgsJointSetLocalFrameB{joint: j, localFrame: localFrame})
	}
	var jnt *joint = getJointFullId(world, j)
	var sim *jointSim = getJointSim(world, jnt)
	sim.localFrameB = localFrame
}

// Get the local frame on body B
func (j Joint) LocalFrameB() Transform {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var sim *jointSim = getJointSim(world, jnt)
	return sim.localFrameB
}

// Toggle collision between connected bodies
func (j Joint) SetCollideConnected(shouldCollide bool) {
	var world *World = getWorldLocked(j.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteJointSetCollideConnected(world.recording, &recArgsJointSetCollideConnected{joint: j, shouldCollide: shouldCollide})
	}

	var jnt *joint = getJointFullId(world, j)
	if jnt.collideConnected == shouldCollide {
		return
	}

	jnt.collideConnected = shouldCollide

	var bodyA *body = &world.bodies[jnt.edges[0].bodyId]
	var bodyB *body = &world.bodies[jnt.edges[1].bodyId]

	if shouldCollide {
		var typeA BodyType = bodyA.bodyType
		var typeB BodyType = bodyB.bodyType

		var shapeId int = nullIndex
		if typeA == DynamicBody && typeB == DynamicBody {
			if bodyA.shapeCount < bodyB.shapeCount {
				shapeId = bodyA.headShapeID
			} else {
				shapeId = bodyB.headShapeID
			}
		} else if typeA == DynamicBody {
			shapeId = bodyA.headShapeID
		} else if typeB == DynamicBody {
			shapeId = bodyB.headShapeID
		}

		for shapeId != nullIndex {
			var shape *shape = &world.shapes[shapeId]

			if shape.proxyKey != nullIndex {
				world.broadPhase.markProxyMovedSerial(shape.proxyKey)
			}

			shapeId = shape.nextShapeID
		}
	} else {
		destroyContactsBetweenBodies(world, bodyA, bodyB)
	}
}

// Is collision allowed between connected bodies?
func (j Joint) CollideConnected() bool {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return jnt.collideConnected
}

// Set the user data on a joint
func (j Joint) SetUserData(userData any) {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	jnt.userData = userData
}

// Get the user data on a joint
func (j Joint) UserData() any {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return jnt.userData
}

// Wake the bodies connect to this joint
func (j Joint) WakeBodies() {
	var world *World = getWorldLocked(j.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteJointWakeBodies(world.recording, &recArgsJointWakeBodies{joint: j})
	}

	var jnt *joint = getJointFullId(world, j)
	var bodyA *body = &world.bodies[jnt.edges[0].bodyId]
	var bodyB *body = &world.bodies[jnt.edges[1].bodyId]

	wakeBody(world, bodyA)
	wakeBody(world, bodyB)
}

func getJointReaction(sim *jointSim, invTimeStep float32) (force float32, torque float32) {
	var linearImpulse float32 = 0.0
	var angularImpulse float32 = 0.0

	switch sim.jointType {
	case DistanceJointType:
		{
			var jnt *distanceJoint = &sim.distanceJoint
			linearImpulse = absFloat(jnt.impulse + jnt.lowerImpulse - jnt.upperImpulse + jnt.motorImpulse)
		}
		break

	case MotorJointType:
		{
			var jnt *motorJoint = &sim.motorJoint
			linearImpulse = jnt.linearVelocityImpulse.Add(jnt.linearSpringImpulse).Length()
			angularImpulse = absFloat(jnt.angularVelocityImpulse + jnt.angularSpringImpulse)
		}
		break

	case MoverJointType:
		{
			var jnt *moverJoint = &sim.moverJoint
			linearImpulse = jnt.linearVelocityImpulse.Length()
		}
		break

	case PogoJointType:
		{
			var jnt *pogoJoint = &sim.pogoJoint
			linearImpulse = absFloat(jnt.impulse)
		}
		break

	case PrismaticJointType:
		{
			var jnt *prismaticJoint = &sim.prismaticJoint
			var perpImpulse float32 = jnt.impulse.X
			var axialImpulse float32 = jnt.motorImpulse + jnt.lowerImpulse - jnt.upperImpulse
			linearImpulse = float32(math.Sqrt(float64(float32(perpImpulse*perpImpulse) + float32(axialImpulse*axialImpulse))))
			angularImpulse = absFloat(jnt.impulse.Y)
		}
		break

	case RevoluteJointType:
		{
			var jnt *revoluteJoint = &sim.revoluteJoint

			linearImpulse = jnt.linearImpulse.Length()
			angularImpulse = absFloat(jnt.motorImpulse + jnt.lowerImpulse - jnt.upperImpulse)
		}
		break

	case WeldJointType:
		{
			var jnt *weldJoint = &sim.weldJoint
			linearImpulse = jnt.linearImpulse.Length()
			angularImpulse = absFloat(jnt.angularImpulse)
		}
		break

	case WheelJointType:
		{
			var jnt *wheelJoint = &sim.wheelJoint
			var perpImpulse float32 = jnt.perpImpulse
			var axialImpulse float32 = jnt.springImpulse + jnt.lowerImpulse - jnt.upperImpulse
			linearImpulse = float32(math.Sqrt(float64(float32(perpImpulse*perpImpulse) + float32(axialImpulse*axialImpulse))))
			angularImpulse = absFloat(jnt.motorImpulse)
		}
		break

	default:
		break
	}

	force = float32(linearImpulse * invTimeStep)
	torque = float32(angularImpulse * invTimeStep)

	return
}

func getJointConstraintForce(world *World, jnt *joint) Vec2 {
	var base *jointSim = getJointSim(world, jnt)

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
		assert(false, "false")
		return Vec2Zero
	}
}

func getJointConstraintTorque(world *World, jnt *joint) float32 {
	var base *jointSim = getJointSim(world, jnt)

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

	default:
		assert(false, "false")
		return 0.0
	}
}

// Get the current constraint force for this joint. Usually in Newtons.
func (j Joint) ConstraintForce() Vec2 {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return getJointConstraintForce(world, jnt)
}

// Get the current constraint torque for this joint. Usually in Newton * meters.
func (j Joint) ConstraintTorque() float32 {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	return getJointConstraintTorque(world, jnt)
}

// Get the current linear separation error for this joint. Does not consider admissible movement. Usually in meters.
func (j Joint) LinearSeparation() float32 {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)

	// Relative to body A so the difference stays in float32 precision far from the origin
	var wxfA WorldTransform = getBodyTransform(world, jnt.edges[0].bodyId)
	var xfA Transform = wxfA.Relative(wxfA.P)
	var xfB Transform = getBodyTransform(world, jnt.edges[1].bodyId).Relative(wxfA.P)

	var pA Vec2 = xfA.Apply(base.localFrameA.P)
	var pB Vec2 = xfB.Apply(base.localFrameB.P)
	var dp Vec2 = pB.Sub(pA)

	switch jnt.jointType {
	case DistanceJointType:
		{
			var distanceJoint *distanceJoint = &base.distanceJoint
			var length float32 = dp.Length()
			if distanceJoint.enableSpring {
				if distanceJoint.enableLimit {
					if length < distanceJoint.minLength {
						return distanceJoint.minLength - length
					}

					if length > distanceJoint.maxLength {
						return length - distanceJoint.maxLength
					}

					return 0.0
				}

				return 0.0
			}

			return absFloat(length - distanceJoint.length)
		}

	case MotorJointType:
		return 0.0

	case FilterJointType:
		return 0.0

	case PrismaticJointType:
		{
			var prismaticJoint *prismaticJoint = &base.prismaticJoint
			var axisA Vec2 = xfA.Q.Rotate(Vec2{1.0, 0.0})
			var perpA Vec2 = axisA.LeftPerp()
			var perpendicularSeparation float32 = absFloat(perpA.Dot(dp))
			var limitSeparation float32 = 0.0

			if prismaticJoint.enableLimit {
				var translation float32 = axisA.Dot(dp)
				if translation < prismaticJoint.lowerTranslation {
					limitSeparation = prismaticJoint.lowerTranslation - translation
				}

				if prismaticJoint.upperTranslation < translation {
					limitSeparation = translation - prismaticJoint.upperTranslation
				}
			}

			return float32(math.Sqrt(float64(float32(perpendicularSeparation*perpendicularSeparation) + float32(limitSeparation*limitSeparation))))
		}

	case RevoluteJointType:
		return dp.Length()

	case WeldJointType:
		{
			var weldJoint *weldJoint = &base.weldJoint
			if weldJoint.linearHertz == 0.0 {
				return dp.Length()
			}

			return 0.0
		}

	case WheelJointType:
		{
			var wheelJoint *wheelJoint = &base.wheelJoint
			var axisA Vec2 = xfA.Q.Rotate(Vec2{1.0, 0.0})
			var perpA Vec2 = axisA.LeftPerp()
			var perpendicularSeparation float32 = absFloat(perpA.Dot(dp))
			var limitSeparation float32 = 0.0

			if wheelJoint.enableLimit {
				var translation float32 = axisA.Dot(dp)
				if translation < wheelJoint.lowerTranslation {
					limitSeparation = wheelJoint.lowerTranslation - translation
				}

				if wheelJoint.upperTranslation < translation {
					limitSeparation = translation - wheelJoint.upperTranslation
				}
			}

			return float32(math.Sqrt(float64(float32(perpendicularSeparation*perpendicularSeparation) + float32(limitSeparation*limitSeparation))))
		}

	default:
		assert(false, "false")
		return 0.0
	}
}

// Get the current angular separation error for this joint. Does not consider admissible movement. Usually in meters.
func (j Joint) AngularSeparation() float32 {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)

	var qA Rot = getBodyTransform(world, jnt.edges[0].bodyId).Q
	var qB Rot = getBodyTransform(world, jnt.edges[1].bodyId).Q
	var relativeAngle float32 = qA.RelativeAngle(qB)

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
		{
			var revoluteJoint *revoluteJoint = &base.revoluteJoint
			if revoluteJoint.enableLimit {
				var angle float32 = relativeAngle
				if angle < revoluteJoint.lowerAngle {
					return revoluteJoint.lowerAngle - angle
				}

				if revoluteJoint.upperAngle < angle {
					return angle - revoluteJoint.upperAngle
				}
			}

			return 0.0
		}

	case WeldJointType:
		{
			var weldJoint *weldJoint = &base.weldJoint
			if weldJoint.angularHertz == 0.0 {
				return relativeAngle
			}

			return 0.0
		}

	case WheelJointType:
		return 0.0

	default:
		assert(false, "false")
		return 0.0
	}
}

// Set the joint constraint tuning. Advanced feature.
func (j Joint) SetConstraintTuning(hertz float32, dampingRatio float32) {
	checkInput(IsValidFloat(hertz) && hertz >= 0.0, "b2IsValidFloat( hertz ) && hertz >= 0.0f")
	checkInput(IsValidFloat(dampingRatio) && dampingRatio >= 0.0, "b2IsValidFloat( dampingRatio ) && dampingRatio >= 0.0f")

	var world *World = j.world
	if world.recording != nil {
		recWriteJointSetConstraintTuning(world.recording, &recArgsJointSetConstraintTuning{joint: j, hertz: hertz, dampingRatio: dampingRatio})
	}
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)
	base.constraintHertz = hertz
	base.constraintDampingRatio = dampingRatio
}

// Get the joint constraint tuning. Advanced feature.
func (j Joint) ConstraintTuning() (hertz float32, dampingRatio float32) {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)
	hertz = base.constraintHertz
	dampingRatio = base.constraintDampingRatio

	return
}

// Set the force threshold for joint events (Newtons)
func (j Joint) SetForceThreshold(threshold float32) {
	checkInput(IsValidFloat(threshold) && threshold >= 0.0, "b2IsValidFloat( threshold ) && threshold >= 0.0f")

	var world *World = j.world
	if world.recording != nil {
		recWriteJointSetForceThreshold(world.recording, &recArgsJointSetForceThreshold{joint: j, threshold: threshold})
	}
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)
	base.forceThreshold = threshold
}

// Get the force threshold for joint events (Newtons)
func (j Joint) ForceThreshold() float32 {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)
	return base.forceThreshold
}

// Set the torque threshold for joint events (N-m)
func (j Joint) SetTorqueThreshold(threshold float32) {
	checkInput(IsValidFloat(threshold) && threshold >= 0.0, "b2IsValidFloat( threshold ) && threshold >= 0.0f")

	var world *World = j.world
	if world.recording != nil {
		recWriteJointSetTorqueThreshold(world.recording, &recArgsJointSetTorqueThreshold{joint: j, threshold: threshold})
	}
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)
	base.torqueThreshold = threshold
}

// Get the torque threshold for joint events (N-m)
func (j Joint) TorqueThreshold() float32 {
	var world *World = j.world
	var jnt *joint = getJointFullId(world, j)
	var base *jointSim = getJointSim(world, jnt)
	return base.torqueThreshold
}

func prepareJoint(jnt *jointSim, context *stepContext) {
	// Clamp joint hertz based on the time step to reduce jitter.
	var hertz float32 = minFloat(jnt.constraintHertz, float32(0.25*context.invH))
	jnt.constraintSoftness = makeSoft(hertz, jnt.constraintDampingRatio, context.h)

	switch jnt.jointType {
	case DistanceJointType:
		prepareDistanceJoint(jnt, context)
		break

	case MotorJointType:
		prepareMotorJoint(jnt, context)
		break

	case MoverJointType:
		prepareMoverJoint(jnt, context)
		break

	case FilterJointType:
		break

	case PogoJointType:
		preparePogoJoint(jnt, context)
		break

	case PrismaticJointType:
		preparePrismaticJoint(jnt, context)
		break

	case RevoluteJointType:
		prepareRevoluteJoint(jnt, context)
		break

	case WeldJointType:
		prepareWeldJoint(jnt, context)
		break

	case WheelJointType:
		prepareWheelJoint(jnt, context)
		break

	default:
		assert(false, "false")
	}
}

func warmStartJoint(jnt *jointSim, context *stepContext) {
	switch jnt.jointType {
	case DistanceJointType:
		warmStartDistanceJoint(jnt, context)
		break

	case MotorJointType:
		warmStartMotorJoint(jnt, context)
		break

	case MoverJointType:
		warmStartMoverJoint(jnt, context)
		break

	case FilterJointType:
		break

	case PogoJointType:
		warmStartPogoJoint(jnt, context)
		break

	case PrismaticJointType:
		warmStartPrismaticJoint(jnt, context)
		break

	case RevoluteJointType:
		warmStartRevoluteJoint(jnt, context)
		break

	case WeldJointType:
		warmStartWeldJoint(jnt, context)
		break

	case WheelJointType:
		warmStartWheelJoint(jnt, context)
		break

	default:
		assert(false, "false")
	}
}

func solveJoint(jnt *jointSim, context *stepContext, useBias bool) {
	switch jnt.jointType {
	case DistanceJointType:
		solveDistanceJoint(jnt, context, useBias)
		break

	case MotorJointType:
		solveMotorJoint(jnt, context)
		break

	case MoverJointType:
		solveMoverJoint(jnt, context)
		break

	case FilterJointType:
		break

	case PogoJointType:
		solvePogoJoint(jnt, context, useBias)
		break

	case PrismaticJointType:
		solvePrismaticJoint(jnt, context, useBias)
		break

	case RevoluteJointType:
		solveRevoluteJoint(jnt, context, useBias)
		break

	case WeldJointType:
		solveWeldJoint(jnt, context, useBias)
		break

	case WheelJointType:
		solveWheelJoint(jnt, context, useBias)
		break

	default:
		assert(false, "false")
	}
}

func prepareJointsOverflow(context *stepContext) {
	var graph *constraintGraph = context.graph
	var joints []jointSim = graph.colors[overflowIndex].jointSims
	var jointCount int = len(graph.colors[overflowIndex].jointSims)

	for i := 0; i < jointCount; i++ {
		var jnt *jointSim = &joints[i]
		prepareJoint(jnt, context)
	}
}

func warmStartJointsOverflow(context *stepContext) {
	var graph *constraintGraph = context.graph
	var joints []jointSim = graph.colors[overflowIndex].jointSims
	var jointCount int = len(graph.colors[overflowIndex].jointSims)

	for i := 0; i < jointCount; i++ {
		var jnt *jointSim = &joints[i]
		warmStartJoint(jnt, context)
	}
}

func solveJointsOverflow(context *stepContext, useBias bool) {
	var graph *constraintGraph = context.graph
	var joints []jointSim = graph.colors[overflowIndex].jointSims
	var jointCount int = len(graph.colors[overflowIndex].jointSims)

	for i := 0; i < jointCount; i++ {
		var jnt *jointSim = &joints[i]
		solveJoint(jnt, context, useBias)
	}
}

func prepareJointsTask(block solverBlock, context *stepContext) {
	var spans []jointPrepareSpan = context.jointPrepareSpans

	var index int = block.startIndex
	var endIndex int = block.startIndex + int(block.count)

	// Find color for start index. Linear search but fast.
	var colorIndex int = 0
	for spans[colorIndex+1].start <= index {
		colorIndex += 1
	}

	// Loop over block
	for index < endIndex {
		var colorStart int = spans[colorIndex].start
		var colorEndIndex int = min(spans[colorIndex+1].start, endIndex)
		var joints []jointSim = spans[colorIndex].joints

		// Loop over color
		for ; index < colorEndIndex; index++ {
			assert(0 <= index-colorStart && index-colorStart < spans[colorIndex].count, "0 <= index - colorStart && index - colorStart < spans[colorIndex].count")
			var jnt *jointSim = &joints[(index - colorStart)]
			prepareJoint(jnt, context)
		}

		// Advance to next color
		colorIndex += 1
	}
}

func warmStartJointsTask(block solverBlock, context *stepContext) {
	var color *graphColor = &context.graph.colors[block.colorIndex]
	var joints []jointSim = color.jointSims

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		var jnt *jointSim = &joints[i]
		warmStartJoint(jnt, context)
	}
}

func solveJointsTask(block solverBlock, context *stepContext, useBias bool, workerIndex int) {
	var color *graphColor = &context.graph.colors[block.colorIndex]
	var joints []jointSim = color.jointSims

	assert(0 <= block.startIndex && block.startIndex+int(block.count) <= len(color.jointSims), "0 <= block.startIndex && block.startIndex + block.count <= color->jointSims.count")

	var jointStateBitSet *bitSet = &context.world.taskContexts[workerIndex].jointStateBitSet

	for i := block.startIndex; i < block.startIndex+int(block.count); i++ {
		var jnt *jointSim = &joints[i]
		solveJoint(jnt, context, useBias)

		if useBias && (jnt.forceThreshold < math.MaxFloat32 || jnt.torqueThreshold < math.MaxFloat32) &&
			getBit(jointStateBitSet, uint32(jnt.jointId)) == false {
			var force, torque float32
			force, torque = getJointReaction(jnt, context.invH)

			// Check thresholds. A zero threshold means all awake joints get reported.
			if force >= jnt.forceThreshold || torque >= jnt.torqueThreshold {
				// Flag this joint for processing.
				setBit(jointStateBitSet, uint32(jnt.jointId))
			}
		}
	}
}

func drawJoint(draw *DebugDraw, world *World, jnt *joint) {
	var bodyA *body = &world.bodies[jnt.edges[0].bodyId]
	var bodyB *body = &world.bodies[jnt.edges[1].bodyId]
	if bodyA.setIndex == disabledSet || bodyB.setIndex == disabledSet {
		return
	}

	var sim *jointSim = getJointSim(world, jnt)

	var xfA WorldTransform = getBodyTransformQuick(world, bodyA)
	var xfB WorldTransform = getBodyTransformQuick(world, bodyB)

	var pA Pos = xfA.Apply(sim.localFrameA.P)
	var pB Pos = xfB.Apply(sim.localFrameB.P)

	var scale float32 = maxFloat(0.0001, float32(draw.JointScale*jnt.drawScale))

	switch jnt.jointType {
	case DistanceJointType:
		drawDistanceJoint(draw, sim, xfA, xfB)
		break

	case FilterJointType:
		draw.DrawPoint(pA, 8.0, ColorLightSkyBlue)
		draw.DrawPoint(pB, 8.0, ColorLightSkyBlue)
		if draw.DrawJointExtras {
			draw.DrawLine(pA, pB, ColorGold)
		}
		break

	case MotorJointType:
		draw.DrawPoint(pA, 8.0, ColorYellowGreen)
		draw.DrawPoint(pB, 8.0, ColorPlum)
		if draw.DrawJointExtras {
			draw.DrawLine(pA, pB, ColorLightGray)
		}
		break

	case MoverJointType:
		draw.DrawPoint(pB, 8.0, ColorPlum)
		break

	case PogoJointType:
		draw.DrawLine(pA, pB, ColorLightGray)
		break

	case PrismaticJointType:
		drawPrismaticJoint(draw, sim, xfA, xfB, scale)
		break

	case RevoluteJointType:
		drawRevoluteJoint(draw, sim, xfA, xfB, scale)
		break

	case WeldJointType:
		drawWeldJoint(draw, sim, xfA, xfB, scale)
		break

	case WheelJointType:
		drawWheelJoint(draw, sim, xfA, xfB, scale)
		break

	default:
		break
	}

	if draw.DrawGraphColors {
		var colorIndex int = jnt.colorIndex
		if colorIndex != nullIndex {
			var p Pos = pA.Lerp(pB, 0.5)
			draw.DrawPoint(p, 5.0, colorForGraph(colorIndex))
		}
	}

	if draw.DrawJointExtras {
		var force Vec2 = getJointConstraintForce(world, jnt)
		var torque float32 = getJointConstraintTorque(world, jnt)
		var p Pos = pA.Lerp(pB, 0.5)

		draw.DrawLine(p, p.Add(force.Scale(0.001)), ColorAzure)

		buffer := fmt.Sprintf("f = [%.6g, %.6g], t = %.6g", force.X, force.Y, torque)
		draw.DrawString(p, buffer, ColorAzure)
	}
}
