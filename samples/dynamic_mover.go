// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import "github.com/zigai/box2d-go/box2d"

type DynamicMoverDef struct {
	position             box2d.Pos
	capsule              box2d.Capsule
	density              float32
	filter               box2d.Filter
	jumpSpeed            float32
	maxSpeed             float32
	minSpeed             float32
	stopSpeed            float32
	accelerate           float32
	airSteer             float32
	friction             float32
	gravityScale         float32
	maxGroundForce       float32
	maxAirForce          float32
	pogoRestLength       float32
	pogoHertz            float32
	pogoDampingRatio     float32
	pogoCompressionScale float32
	pogoTensionScale     float32
	enablePreSolveEvents bool
}

func DefaultDynamicMoverDef() DynamicMoverDef {
	return DynamicMoverDef{
		density:              1.0,
		filter:               box2d.DefaultFilter(),
		jumpSpeed:            7.0,
		maxSpeed:             6.0,
		minSpeed:             0.1,
		stopSpeed:            3.0,
		accelerate:           20.0,
		airSteer:             0.5,
		friction:             8.0,
		gravityScale:         1.5,
		maxGroundForce:       70.0,
		maxAirForce:          20.0,
		pogoRestLength:       0.9,
		pogoHertz:            5.0,
		pogoDampingRatio:     0.8,
		pogoCompressionScale: 100.0,
		pogoTensionScale:     100.0,
	}
}

type DynamicMoverCastResult struct {
	point    box2d.Pos
	normal   box2d.Vec2
	bodyID   box2d.Body
	fraction float32
	hit      bool
}

// Dynamic character mover. The capsule is a rigid body with a locked rotation, so it
// collides and is pushed like everything else in the world. A mover joint drives it to
// the target velocity and a pogo joint holds it above the ground so it can climb
// steps without touching the capsule.
//
// The dynamic mover should be vendored into your project so you can alter it to meet
// your game's specific needs.
//
// This system is inspired by Tom Waterson's presentation:
// https://www.youtube.com/watch?v=_jRLlTDqoGI
type DynamicMover struct {
	jumpSpeed  float32
	maxSpeed   float32
	minSpeed   float32
	stopSpeed  float32
	accelerate float32

	// Fraction of the ground acceleration available in the air.
	airSteer float32

	// Ground friction in units of 1/time.
	friction float32

	// A character usually wants to fall faster than the rest of the world.
	gravityScale float32

	// Force budget the mover joint has for reaching the target velocity. The air value
	// is small so the character keeps its momentum while jumping.
	maxGroundForce float32
	maxAirForce    float32

	pogoRestLength   float32
	pogoHertz        float32
	pogoDampingRatio float32

	// Pogo force limits as a multiple of the mover weight.
	pogoCompressionScale float32
	pogoTensionScale     float32

	// Surfaces steeper than this are not ground.
	minGroundNormalY float32

	capsule box2d.Capsule
	filter  box2d.Filter

	worldID      *box2d.World
	groundID     box2d.Body
	moverID      box2d.Body
	moverJointID box2d.MoverJoint
	pogoJointID  box2d.PogoJoint

	castResult DynamicMoverCastResult

	velocity  box2d.Vec2
	onGround  bool
	walkable  bool
	jumping   bool
	jumpTicks int

	// Pogo joint state, carried across the joint being recreated each solve.
	pogoImpulse  float32
	pogoVelocity float32
	pogoLength   float32

	// Ground probe from the last solve, kept for debug drawing. The probe ends at
	// the origin plus the fraction of the translation.
	pogoOrigin      box2d.Pos
	pogoTranslation box2d.Vec2
	pogoFraction    float32
	pogoHit         bool
}

func NewDynamicMover() DynamicMover {
	return DynamicMover{
		jumpSpeed:            5.0,
		maxSpeed:             6.0,
		minSpeed:             0.1,
		stopSpeed:            3.0,
		accelerate:           20.0,
		airSteer:             0.2,
		friction:             8.0,
		gravityScale:         1.5,
		maxGroundForce:       70.0,
		maxAirForce:          20.0,
		pogoRestLength:       0.9,
		pogoHertz:            5.0,
		pogoDampingRatio:     0.8,
		pogoCompressionScale: 100.0,
		pogoTensionScale:     100.0,
		minGroundNormalY:     0.7,

		capsule: box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.5}, Center2: box2d.Vec2{X: 0.0, Y: 0.5}, Radius: 0.3},

		filter: box2d.DefaultFilter(),

		pogoFraction: 1.0,
	}
}

func (m *DynamicMover) Create(worldID *box2d.World, def *DynamicMoverDef) {
	m.capsule = def.capsule
	m.filter = def.filter
	m.jumpSpeed = def.jumpSpeed
	m.maxSpeed = def.maxSpeed
	m.minSpeed = def.minSpeed
	m.stopSpeed = def.stopSpeed
	m.accelerate = def.accelerate
	m.airSteer = def.airSteer
	m.friction = def.friction
	m.gravityScale = def.gravityScale
	m.maxGroundForce = def.maxGroundForce
	m.maxAirForce = def.maxAirForce
	m.pogoRestLength = def.pogoRestLength
	m.pogoHertz = def.pogoHertz
	m.pogoDampingRatio = def.pogoDampingRatio
	m.pogoCompressionScale = def.pogoCompressionScale
	m.pogoTensionScale = def.pogoTensionScale
	m.minGroundNormalY = 0.7

	m.worldID = worldID
	m.velocity = box2d.Vec2Zero
	m.onGround = false
	m.walkable = false
	m.jumping = false
	m.pogoImpulse = 0.0
	m.pogoVelocity = 0.0
	m.pogoLength = 0.0
	m.pogoJointID = box2d.PogoJoint{}

	bodyDef := box2d.DefaultBodyDef()
	m.groundID = worldID.CreateBody(&bodyDef)

	// Position is the center of the capsule.
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = def.position
	bodyDef.GravityScale = m.gravityScale
	bodyDef.MotionLocks.AngularZ = true
	bodyDef.EnableSleep = false
	bodyDef.Name = "mover"

	m.moverID = worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = def.density

	// The mover joint and the pogo joint handle all the movement, surface friction would fight them
	shapeDef.Material.Friction = 0.0
	shapeDef.Filter = def.filter
	shapeDef.EnablePreSolveEvents = def.enablePreSolveEvents

	m.moverID.CreateCapsuleShape(&shapeDef, &m.capsule)

	moverDef := box2d.DefaultMoverJointDef()
	moverDef.LinearVelocity = m.velocity
	moverDef.MaxVelocityForce = box2d.Vec2{X: m.maxGroundForce, Y: 0.0}
	moverDef.Base.BodyA = m.groundID
	moverDef.Base.BodyB = m.moverID
	moverDef.Base.CollideConnected = true

	m.moverJointID = worldID.CreateMoverJoint(&moverDef)
}

func (m *DynamicMover) Destroy() {
	if m.moverID != (box2d.Body{}) {
		m.moverID.Destroy()
	}

	if m.groundID != (box2d.Body{}) {
		m.groundID.Destroy()
	}

	m.groundID = box2d.Body{}
	m.moverID = box2d.Body{}

	// Joints are implicitly destroyed.
	m.moverJointID = box2d.MoverJoint{}
	m.pogoJointID = box2d.PogoJoint{}
}

// Gravity scale is used to size the pogo forces, so change it here.
func (m *DynamicMover) SetGravityScale(gravityScale float32) {
	m.gravityScale = gravityScale

	if m.moverID != (box2d.Body{}) {
		m.moverID.SetGravityScale(gravityScale)
	}
}

// Jump if grounded. Returns false when airborne.
func (m *DynamicMover) Jump() bool {
	if !m.onGround || !m.walkable {
		return false
	}

	surfaceVelocity := float32(0.0)
	if m.castResult.bodyID.IsValid() {
		v := m.castResult.bodyID.WorldPointVelocity(m.castResult.point)
		surfaceVelocity = v.Y
	}

	mass := m.moverID.Mass()
	vy := m.moverID.LinearVelocity().Y

	// Remove the pogo constraint velocity but add the surface velocity.
	dv := max(0.0, m.jumpSpeed-vy) + surfaceVelocity
	m.moverID.ApplyLinearImpulseToCenter(box2d.Vec2{X: 0.0, Y: mass * dv}, true)

	// This removes the pogo step down on the next update.
	m.onGround = false
	m.walkable = false
	m.jumping = true
	m.jumpTicks = 0
	return true
}

// Movement follows the Quake ground and air model:
// https://github.com/id-Software/Quake/blob/master/QW/client/pmove.c#L390

// onGround - controls friction, mover max force, and pogo step-down extension
// walkable - controls air steer
// jumping - blocks the pogo from pulling down
func (m *DynamicMover) Update(timeStep, throttle float32) {
	// Reach further while grounded so the spring can find the ground over a step
	stepDownLength := m.pogoRestLength
	rayLength := m.pogoRestLength
	if m.onGround {
		rayLength = m.pogoRestLength + stepDownLength
	}
	translation := box2d.Vec2{X: 0.0, Y: -rayLength}

	position := m.moverID.Position()
	origin := position.Add(m.capsule.Center1)

	queryFilter := box2d.QueryFilter{
		CategoryBits: m.filter.CategoryBits,
		MaskBits:     m.filter.MaskBits,
	}

	m.castResult = DynamicMoverCastResult{}
	m.worldID.CastRay(origin, translation, queryFilter, func(shapeID box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
		m.castResult.point = point
		m.castResult.normal = normal
		m.castResult.bodyID = shapeID.Body()
		m.castResult.fraction = fraction
		m.castResult.hit = true
		return fraction
	})

	m.pogoOrigin = origin
	m.pogoTranslation = translation
	m.pogoFraction = 1.0
	if m.castResult.hit {
		m.pogoFraction = m.castResult.fraction
	}
	m.pogoHit = m.castResult.hit

	// Should jumping end?
	if m.jumping {
		m.jumpTicks++

		// Jump ticks allow time for the jump to leave the ground. Otherwise jumping ends
		// when the character approaches the current ground.
		if m.jumpTicks > 2 && m.castResult.hit && m.velocity.Dot(m.castResult.normal) <= 0.0 {
			m.jumping = false
		}
	}

	if m.castResult.hit {
		// The cast can still hit while jumping.
		m.onGround = !m.jumping

		// Is the ground too steep to walk?
		m.walkable = m.castResult.normal.Y >= m.minGroundNormalY
	} else {
		m.onGround = false
		m.walkable = false
	}

	m.velocity = m.moverID.LinearVelocity()

	// Friction
	if m.onGround {
		speed := m.velocity.Length()
		if speed < m.minSpeed {
			m.velocity.X = 0.0
		} else {
			// Linear damping above stopSpeed and fixed reduction below stopSpeed
			control := speed
			if speed < m.stopSpeed {
				control = m.stopSpeed
			}

			// friction has units of 1/time
			drop := control * m.friction * timeStep
			newSpeed := max(0.0, speed-drop)
			m.velocity.X *= newSpeed / speed
		}
	}

	// Mover force
	if m.onGround {
		maxForce := float32(0.0)
		if m.walkable {
			maxForce = m.maxGroundForce
		}
		m.moverJointID.SetMaxVelocityForce(box2d.Vec2{X: maxForce, Y: 0.0})
	} else {
		m.moverJointID.SetMaxVelocityForce(box2d.Vec2{X: m.maxAirForce, Y: 0.0})
	}

	desiredVelocity := box2d.Vec2{X: m.maxSpeed * throttle, Y: 0.0}
	desiredSpeed, desiredDirection := desiredVelocity.LengthAndNormalize()

	if desiredSpeed > m.maxSpeed {
		desiredSpeed = m.maxSpeed
	}

	// Accelerate
	currentSpeed := m.velocity.Dot(desiredDirection)
	addSpeed := desiredSpeed - currentSpeed
	if addSpeed > 0.0 {
		steer := m.airSteer
		if m.walkable {
			steer = 1.0
		}
		accelSpeed := steer * m.accelerate * m.maxSpeed * timeStep
		if accelSpeed > addSpeed {
			accelSpeed = addSpeed
		}

		m.velocity.X += accelSpeed * desiredDirection.X
	}

	// The pogo joint is rebuilt every solve because it may land on a different body
	if m.pogoJointID.IsValid() {
		m.pogoLength = m.pogoJointID.Length()
		m.pogoImpulse = m.pogoJointID.Impulse()
		m.pogoVelocity = m.pogoJointID.Velocity()

		m.pogoJointID.Destroy()
		m.pogoJointID = box2d.PogoJoint{}
	}

	if m.castResult.hit {
		moverMass := m.moverID.Mass()
		moverWeight := m.gravityScale * m.worldID.Gravity().Length() * moverMass

		pogoDef := box2d.DefaultPogoJointDef()
		pogoDef.Base.LocalFrameA.P = m.castResult.bodyID.LocalPoint(m.castResult.point)
		pogoDef.Base.LocalFrameB.P = m.capsule.Center1
		pogoDef.Normal = m.castResult.normal
		pogoDef.Base.BodyA = m.castResult.bodyID
		pogoDef.Base.BodyB = m.moverID
		pogoDef.Base.CollideConnected = true
		pogoDef.RestLength = m.pogoRestLength
		pogoDef.Hertz = m.pogoHertz
		pogoDef.DampingRatio = m.pogoDampingRatio
		pogoDef.MaxCompressionForce = m.pogoCompressionScale * moverWeight

		// Warm start from the joint that was just destroyed
		pogoDef.Impulse = m.pogoImpulse
		pogoDef.Velocity = m.pogoVelocity

		if m.jumping {
			// Don't allow the pogo to pull down
			pogoDef.MaxTensionForce = 0.0
		} else {
			// The pogo can pull down at a multiple of the gravity force.
			pogoDef.MaxTensionForce = m.pogoTensionScale * moverWeight
		}

		m.pogoJointID = m.worldID.CreatePogoJoint(&pogoDef)
	} else {
		m.pogoImpulse = 0.0
		m.pogoVelocity = 0.0
	}

	m.moverJointID.SetLinearVelocity(m.velocity)
}
