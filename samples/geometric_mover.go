// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"math"

	"github.com/zigai/box2d-go/box2d"
)

type GeometricMoverDef struct {
	position             box2d.Pos
	capsule              box2d.Capsule
	filter               box2d.QueryFilter
	jumpSpeed            float32
	maxSpeed             float32
	minSpeed             float32
	stopSpeed            float32
	accelerate           float32
	airSteer             float32
	friction             float32
	gravity              float32
	pogoRestLength       float32
	pogoHertz            float32
	pogoDampingRatio     float32
	pogoCompressionScale float32
	pogoTensionScale     float32
}

func DefaultGeometricMoverDef() GeometricMoverDef {
	return GeometricMoverDef{
		capsule:              box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.5}, Center2: box2d.Vec2{X: 0.0, Y: 0.5}, Radius: 0.3},
		filter:               box2d.DefaultQueryFilter(),
		jumpSpeed:            10.0,
		maxSpeed:             6.0,
		minSpeed:             0.1,
		stopSpeed:            3.0,
		accelerate:           20.0,
		airSteer:             0.2,
		friction:             8.0,
		gravity:              30.0,
		pogoRestLength:       0.9,
		pogoHertz:            5.0,
		pogoDampingRatio:     0.8,
		pogoCompressionScale: 100.0,
		pogoTensionScale:     100.0,
	}
}

// Optional shape user data read by GeometryMover when it gathers collision planes.
// Shapes without user data are treated as rigid and fully clipping.
type MoverShapeUserData struct {
	// Distance limit on the push this surface may apply to the mover. FLT_MAX is rigid.
	// Use a small value for surfaces that must not shove the mover, such as an elevator
	// that would otherwise push it through the floor.
	maxPush float32

	// False keeps the velocity along this surface, which is what soft collision needs.
	clipVelocity bool
}

// Extra data kept alongside each collision plane so the mover can push rigid bodies.
type PlaneExtra struct {
	point   box2d.Pos
	shapeID box2d.Shape
}

const geometricMoverPlaneCapacity = 8

// Geometric character mover. The capsule is moved by casting
// and by solving collision planes, so the game keeps full control of the motion.
//
// A pogo spring holds the capsule above the ground so it can move up steps.
type GeometricMover struct {
	jumpSpeed  float32
	maxSpeed   float32
	minSpeed   float32
	stopSpeed  float32
	accelerate float32

	// Fraction of the ground acceleration available in the air.
	airSteer float32

	// Ground friction in units of 1/time.
	friction float32

	// The mover is not simulated so it needs its own gravity.
	gravity float32

	pogoRestLength   float32
	pogoHertz        float32
	pogoDampingRatio float32

	// Downward force applied to whatever the pogo stick is resting on.
	pogoGroundForce float32

	// Surfaces steeper than this are not ground.
	minGroundNormalY float32

	// Push touched rigid bodies out of the way by solving a normal constraint against
	// each collision plane. Without this the mover slides along a crate instead of
	// shoving it.
	pushBodies bool

	// The capsule is relative to the mover position.
	capsule box2d.Capsule

	filter box2d.QueryFilter

	worldID *box2d.World

	// The center of the capsule.
	position box2d.Pos

	velocity     box2d.Vec2
	pogoVelocity float32
	onGround     bool

	// Planes from the last solve, kept for debug drawing.
	planes          [geometricMoverPlaneCapacity]box2d.CollisionPlane
	planeExtras     [geometricMoverPlaneCapacity]PlaneExtra
	planeCount      int
	totalIterations int

	pogoOrigin      box2d.Pos
	pogoTranslation box2d.Vec2
	pogoFraction    float32
	pogoHit         bool
}

func NewGeometricMover() GeometricMover {
	return GeometricMover{
		jumpSpeed:        10.0,
		maxSpeed:         6.0,
		minSpeed:         0.1,
		stopSpeed:        3.0,
		accelerate:       20.0,
		airSteer:         0.2,
		friction:         8.0,
		gravity:          30.0,
		pogoRestLength:   0.9,
		pogoHertz:        5.0,
		pogoDampingRatio: 0.8,
		pogoGroundForce:  50.0,
		pushBodies:       true,

		capsule: box2d.Capsule{Center1: box2d.Vec2{X: 0.0, Y: -0.5}, Center2: box2d.Vec2{X: 0.0, Y: 0.5}, Radius: 0.3},

		filter: box2d.DefaultQueryFilter(),

		minGroundNormalY: 0.7,

		pogoFraction: 1.0,
	}
}

// Place the mover. The capsule and tuning may be changed at any time.
func (m *GeometricMover) Create(worldID *box2d.World, def *GeometricMoverDef) {
	m.capsule = def.capsule
	m.filter = def.filter
	m.jumpSpeed = def.jumpSpeed
	m.maxSpeed = def.maxSpeed
	m.minSpeed = def.minSpeed
	m.stopSpeed = def.stopSpeed
	m.accelerate = def.accelerate
	m.airSteer = def.airSteer
	m.friction = def.friction
	m.gravity = def.gravity
	m.pogoRestLength = def.pogoRestLength
	m.pogoHertz = def.pogoHertz
	m.pogoDampingRatio = def.pogoDampingRatio
	m.minGroundNormalY = 0.7

	m.worldID = worldID
	m.position = def.position
	m.velocity = box2d.Vec2Zero
	m.pogoVelocity = 0.0
	m.onGround = false
	m.planeCount = 0
	m.totalIterations = 0
	m.pogoFraction = 1.0
	m.pogoHit = false
}

// Jump if grounded. Returns false when airborne.
func (m *GeometricMover) Jump() bool {
	if !m.onGround {
		return false
	}

	m.velocity.Y = m.jumpSpeed
	m.onGround = false
	return true
}

// Solve a normal constraint between the mover and each touched rigid body. This is the
// same velocity constraint the solver uses for a contact, except the mover is treated as
// having infinite mass, so all of the impulse lands on the body.
func (m *GeometricMover) PushBodies() {
	for i := range m.planeCount {
		bodyID := m.planeExtras[i].shapeID.Body()
		if bodyID.Type() != box2d.DynamicBody {
			continue
		}

		// Plane normals point at the mover
		point := m.planeExtras[i].point
		normal := m.planes[i].Plane.Normal.Neg()

		invMassA := float32(0.0)

		massB := bodyID.Mass()
		invMassB := float32(0.0)
		if massB > 0.0 {
			invMassB = 1.0 / massB
		}

		inertiaB := bodyID.RotationalInertia()
		invIB := float32(0.0)
		if inertiaB > 0.0 {
			invIB = 1.0 / inertiaB
		}

		centerB := bodyID.WorldCenter()
		rB := point.Sub(centerB)

		rnB := rB.Cross(normal)
		kNormal := invMassA + invMassB + invIB*rnB*rnB
		normalMass := float32(0.0)
		if kNormal > 0.0 {
			normalMass = 1.0 / kNormal
		}

		vB := bodyID.LinearVelocity()
		omegaB := bodyID.AngularVelocity()
		vrB := vB.Add(box2d.Vec2{X: -omegaB * rB.Y, Y: omegaB * rB.X})
		vn := vrB.Sub(m.velocity).Dot(normal)

		// Push only, a separating body is left alone
		impulse := max(-normalMass*vn, 0.0)

		P := normal.Scale(impulse)
		m.velocity = m.velocity.MulSub(invMassA, P)

		bodyID.ApplyLinearImpulse(P, point, true)
	}
}

// Movement follows the Quake ground and air model:
// https://github.com/id-Software/Quake/blob/master/QW/client/pmove.c#L390
func (m *GeometricMover) Update(timeStep, throttle float32) {
	// Friction
	speed := m.velocity.Length()
	if speed < m.minSpeed {
		m.velocity.X = 0.0
		m.velocity.Y = 0.0
	} else if m.onGround {
		// Linear damping above stopSpeed and fixed reduction below stopSpeed
		control := speed
		if speed < m.stopSpeed {
			control = m.stopSpeed
		}

		// friction has units of 1/time
		drop := control * m.friction * timeStep
		newSpeed := max(0.0, speed-drop)
		m.velocity = m.velocity.Scale(newSpeed / speed)
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
		if m.onGround {
			steer = 1.0
		}
		accelSpeed := steer * m.accelerate * m.maxSpeed * timeStep
		if accelSpeed > addSpeed {
			accelSpeed = addSpeed
		}

		m.velocity = m.velocity.Add(desiredDirection.Scale(accelSpeed))
	}

	// The probe reaches past the rest length so the spring can be stretched
	rayLength := m.pogoRestLength
	if m.onGround {
		rayLength = 2.0 * m.pogoRestLength
	}
	translation := box2d.Vec2{X: 0.0, Y: -rayLength}
	origin := m.position.Add(m.capsule.Center1)

	var castResult DynamicMoverCastResult
	m.worldID.CastRay(origin, translation, m.filter, func(shapeID box2d.Shape, point box2d.Pos, normal box2d.Vec2, fraction float32) float32 {
		castResult.point = point
		castResult.normal = normal
		castResult.bodyID = shapeID.Body()
		castResult.fraction = fraction
		castResult.hit = true
		return fraction
	})

	m.pogoOrigin = origin
	m.pogoTranslation = translation
	m.pogoFraction = 1.0
	if castResult.hit {
		m.pogoFraction = castResult.fraction
	}
	m.pogoHit = castResult.hit

	// Avoid snapping to ground if still going up
	haveGround := castResult.hit && castResult.normal.Y > m.minGroundNormalY

	if !m.onGround {
		m.onGround = haveGround && m.velocity.Y <= 0.01
	} else {
		m.onGround = haveGround
	}

	if !m.onGround {
		m.velocity.Y -= m.gravity * timeStep
		m.pogoVelocity = 0.0
	} else {
		m.velocity.Y = 0.0

		pogoCurrentLength := castResult.fraction * rayLength

		offset := pogoCurrentLength - m.pogoRestLength
		m.pogoVelocity = box2d.SpringDamper(m.pogoHertz, m.pogoDampingRatio, offset, m.pogoVelocity, timeStep)

		castResult.bodyID.ApplyForce(box2d.Vec2{X: 0.0, Y: -m.pogoGroundForce}, castResult.point, true)
	}

	target := m.position.Add(m.velocity.Scale(timeStep)).Add(box2d.Vec2{X: 0.0, Y: 1.0}.Scale(timeStep * m.pogoVelocity))

	m.totalIterations = 0
	tolerance := float32(0.01)

	// Don't cast against other movers.
	castFilter := box2d.QueryFilter{
		CategoryBits: m.filter.CategoryBits,
		MaskBits:     m.filter.MaskBits &^ m.filter.CategoryBits,
	}

	for range 5 {
		m.planeCount = 0

		m.worldID.CollideMover(m.position, &m.capsule, m.filter, m.collectPlane)
		result := box2d.SolvePlanes(target.Sub(m.position), m.planes[:m.planeCount])

		m.totalIterations += result.IterationCount

		fraction := m.worldID.CastMover(m.position, &m.capsule, result.Delta, castFilter)

		delta := result.Delta.Scale(fraction)
		m.position = m.position.Add(delta)

		if delta.LengthSquared() < tolerance*tolerance {
			break
		}
	}

	if m.pushBodies {
		m.PushBodies()
	}

	// Stop pushing into surfaces so speed can't accumulate against a wall
	m.velocity = box2d.ClipVector(m.velocity, m.planes[:m.planeCount])
}

func (m *GeometricMover) collectPlane(shapeID box2d.Shape, planeResult *box2d.PlaneResult) bool {
	maxPush := float32(math.MaxFloat32)
	clipVelocity := true
	if userData, ok := shapeID.UserData().(*MoverShapeUserData); ok {
		maxPush = userData.maxPush
		clipVelocity = userData.clipVelocity
	}

	if m.planeCount < geometricMoverPlaneCapacity {
		m.planes[m.planeCount] = box2d.CollisionPlane{Plane: planeResult.Plane, PushLimit: maxPush, Push: 0.0, ClipVelocity: clipVelocity}

		// The plane point is relative to the origin passed to the collide call
		m.planeExtras[m.planeCount] = PlaneExtra{
			point:   m.position.Add(planeResult.Point),
			shapeID: shapeID,
		}
		m.planeCount++
	}

	return true
}
