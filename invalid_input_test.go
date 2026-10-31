// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

var badInputValues = []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))}

func inputIsBodyFinite(bodyID box2d.Body) bool {
	return bodyID.Position().IsValid() && bodyID.Rotation().IsValid() && bodyID.LinearVelocity().IsValid() && box2d.IsValidFloat(bodyID.AngularVelocity())
}

func inputReadBodyStep(r *oracle.Reader, trace *transformTrace, worldID *box2d.World, bodyID box2d.Body, step int) {
	trace.read(r, worldID, fmt.Sprintf("step.%d.hash", step), true)

	p := bodyID.Position()
	r.Pos(fmt.Sprintf("step.%d.position", step), p.X, p.Y)

	q := bodyID.Rotation()
	r.Rot(fmt.Sprintf("step.%d.rotation", step), q.C, q.S)

	v := bodyID.LinearVelocity()
	r.Vec2(fmt.Sprintf("step.%d.velocity", step), v.X, v.Y)
	r.F32(fmt.Sprintf("step.%d.angular", step), bodyID.AngularVelocity())
}

func inputReadRejectedSettings(r *oracle.Reader, world *box2d.World) {
	hertz, damping, push := box2d.WorldTuning(world)
	r.F32("settings.contact.hertz", hertz)
	r.F32("settings.contact.damping", damping)
	r.F32("settings.contact.push", push)

	for _, body := range box2d.BodiesInSlotOrder(world) {
		m := body.MassData()
		r.F32("settings.mass", m.Mass)
		r.Vec2("settings.center", m.Center.X, m.Center.Y)
		r.F32("settings.inertia", m.RotationalInertia)
		r.F32("settings.linearDamping", body.LinearDamping())
		r.F32("settings.angularDamping", body.AngularDamping())
		r.F32("settings.gravityScale", body.GravityScale())
		r.F32("settings.sleepThreshold", body.SleepThreshold())
		r.F32("settings.safetyFactor", body.SafetyFactor())

		shapeCount := 0
		for shape := range body.Shapes() {
			shapeCount++

			r.Bool("settings.shape.valid", shape.IsValid())
			r.Bool("settings.shape.owner", shape.Body() == body)
			r.F32("settings.shape.density", shape.Density())
			material := shape.SurfaceMaterial()
			r.F32("settings.material.friction", material.Friction)
			r.F32("settings.material.restitution", material.Restitution)
			r.F32("settings.material.rolling", material.RollingResistance)
			r.F32("settings.material.tangent", material.TangentSpeed)
			r.U64("settings.material.tag", material.UserMaterialID)
			r.U32("settings.material.color", material.CustomColor)

			bounds := shape.AABB()
			r.Vec2("settings.aabb.lower", bounds.LowerBound.X, bounds.LowerBound.Y)
			r.Vec2("settings.aabb.upper", bounds.UpperBound.X, bounds.UpperBound.Y)
			r.I32("settings.shape.type", int32(shape.Type()))

			if shape.Type() == box2d.PolygonShape {
				polygon := shape.Polygon()
				r.I32("settings.polygon.count", int32(polygon.Count))
				r.F32("settings.polygon.radius", polygon.Radius)
				r.Vec2("settings.polygon.centroid", polygon.Centroid.X, polygon.Centroid.Y)

				for i := range polygon.Count {
					r.Vec2("settings.polygon.vertex", polygon.Vertices[i].X, polygon.Vertices[i].Y)
					r.Vec2("settings.polygon.normal", polygon.Normals[i].X, polygon.Normals[i].Y)
				}
			}
		}

		r.I32("settings.body.shapes", int32(shapeCount))
	}
}

func inputReadRejectedJoint(r *oracle.Reader, joint box2d.RevoluteJoint) {
	a, b := joint.LocalFrameA(), joint.LocalFrameB()
	r.Vec2("joint.frameA.p", a.P.X, a.P.Y)
	r.Rot("joint.frameA.q", a.Q.C, a.Q.S)
	r.Vec2("joint.frameB.p", b.P.X, b.P.Y)
	r.Rot("joint.frameB.q", b.Q.C, b.Q.S)

	hertz, damping := joint.ConstraintTuning()
	r.F32("joint.hertz", hertz)
	r.F32("joint.damping", damping)
	r.F32("joint.forceThreshold", joint.ForceThreshold())
	r.F32("joint.torqueThreshold", joint.TorqueThreshold())
	r.F32("joint.springHertz", joint.SpringHertz())
	r.F32("joint.springDamping", joint.SpringDampingRatio())
	r.F32("joint.target", joint.TargetAngle())
	r.F32("joint.lower", joint.LowerLimit())
	r.F32("joint.upper", joint.UpperLimit())
	r.F32("joint.motorSpeed", joint.MotorSpeed())
	r.F32("joint.maxTorque", joint.MaxMotorTorque())
}

func inputReadConstructionCounts(t *testing.T, r *oracle.Reader, world *box2d.World, bodyCount int) {
	t.Helper()

	counts := world.Counters()
	r.I32("construction.bodies", int32(counts.BodyCount))
	r.I32("construction.shapes", int32(counts.ShapeCount))
	r.I32("construction.joints", int32(counts.JointCount))
	r.I32("construction.contacts", int32(counts.ContactCount))
	r.I32("construction.chains", int32(box2d.ChainCount(world)))

	if counts.BodyCount != bodyCount || counts.ShapeCount != 1 || counts.JointCount != 0 || counts.ContactCount != 0 || box2d.ChainCount(world) != 0 {
		t.Fatalf("failed construction leaked objects: %+v, chains=%d", counts, box2d.ChainCount(world))
	}

	if len(box2d.BodiesInSlotOrder(world)) != bodyCount {
		t.Fatal("body enumeration disagrees with construction counts")
	}

	shapeCount, jointCount := inputConstructionMembership(t, world)

	r.I32("construction.enumerated.shapes", int32(shapeCount))
	r.I32("construction.enumerated.joints", int32(jointCount))

	if shapeCount != 1 || jointCount != 0 {
		t.Fatal("failed construction changed enumeration")
	}
}

func inputIndependentWorld(t *testing.T, r *oracle.Reader) {
	t.Helper()

	def := box2d.DefaultWorldDef()
	def.WorkerCount = 4
	def.Gravity = box2d.Vec2{X: 3, Y: -7}
	world := box2d.NewWorld(&def)
	box2d.CleanupWorld(t, world)
	r.Bool("independent.world.valid", world.IsValid())

	if !world.IsValid() {
		t.Fatal("valid world creation poisoned by failed creation")
	}

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.LinearVelocity = box2d.Vec2{X: 2, Y: 1}
	body := world.CreateBody(&bodyDef)
	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	shape := body.CreatePolygonShape(&shapeDef, &box)
	r.Bool("independent.body.valid", body.IsValid())
	r.Bool("independent.shape.valid", shape.IsValid())

	if !body.IsValid() || !shape.IsValid() {
		t.Fatal("valid object creation poisoned by rejection")
	}

	trace := transformTrace{}

	for i := range 2 {
		world.Step(1.0/60.0, 4)
		trace.read(r, world, fmt.Sprintf("independent.%d", i), false)
	}

	world.Close()
	r.Bool("independent.world.destroyed", !world.IsValid())

	if world.IsValid() {
		t.Fatal("independent world remains valid after destroy")
	}
}

func TestInvalidInput(t *testing.T) {
	t.Run("WorldInputTest", func(t *testing.T) { checkWorldInputTest(t) })
	t.Run("BodyInputTest", func(t *testing.T) { checkBodyInputTest(t) })
	t.Run("ShapeInputTest", func(t *testing.T) { checkShapeInputTest(t) })
	t.Run("JointInputTest", func(t *testing.T) { checkJointInputTest(t) })
	t.Run("CreateInputTest", func(t *testing.T) { checkCreateInputTest(t) })
	t.Run("RecordReplayTest", func(t *testing.T) { inputCheckRecordReplayTest(t) })
}

// rejectedInput reports whether call panicked on invalid input and counts each rejection.
func rejectedInput(count *int, call func()) bool {
	rejected := false

	func() {
		defer func() {
			if v := recover(); v != nil {
				if _, ok := v.(*box2d.AssertError); !ok {
					panic(v)
				}

				*count++
				rejected = true
			}
		}()

		call()
	}()

	return rejected
}

type inputCallChecks struct {
	t     *testing.T
	r     *oracle.Reader
	world *box2d.World
	count *int
}

func (c *inputCallChecks) check(kind string, want int, call func()) {
	c.t.Helper()
	before := *c.count

	stateBefore := c.world.StateHash()
	if want == 1 {
		c.r.U64("reject.before", stateBefore)
	}

	rejectedInput(c.count, call)
	c.r.I32(kind, int32(*c.count-before))

	if want == 1 {
		stateAfter := c.world.StateHash()
		c.r.U64("reject.after", stateAfter)

		if stateAfter != stateBefore {
			c.t.Fatalf("rejected call changed canonical state: %x -> %x", stateBefore, stateAfter)
		}

		inputReadRejectedSettings(c.r, c.world)
	}

	if *c.count != before+want {
		c.t.Fatal("condition false")
	}
}

func checkWorldInputTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "invalid_input/WorldInputTest")
	trace := transformTrace{}
	rejectCount := 0

	checks := inputCallChecks{t: t, r: r, count: &rejectCount}
	checkCall := checks.check
	worldDef := box2d.DefaultWorldDef()
	worldID := box2d.NewWorld(&worldDef)
	checks.world = worldID
	box2d.CleanupWorld(t, worldID)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.LinearVelocity = box2d.Vec2{X: 2, Y: 3}
	body := worldID.CreateBody(&bodyDef)
	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	body.CreatePolygonShape(&shapeDef, &box)

	gravity := worldID.Gravity()
	restitutionThreshold := float32(worldID.RestitutionThreshold())
	hitEventThreshold := float32(worldID.HitEventThreshold())
	recycleDistance := float32(worldID.ContactRecycleDistance())
	maximumLinearSpeed := float32(worldID.MaximumLinearSpeed())
	restitutionIterations := worldID.RestitutionIterations()

	rejectCount = 0

	for i, bad := range badInputValues {
		checkCall("reject", 1, func() { worldID.SetGravity(box2d.Vec2{X: bad, Y: 0.0}) })
		checkCall("reject", 1, func() { worldID.SetGravity(box2d.Vec2{X: 0.0, Y: bad}) })
		checkCall("reject", 1, func() { worldID.SetRestitutionThreshold(bad) })
		checkCall("reject", 1, func() { worldID.SetHitEventThreshold(bad) })
		checkCall("reject", 1, func() { worldID.SetContactRecycleDistance(bad) })
		checkCall("reject", 1, func() { worldID.SetMaximumLinearSpeed(bad) })
		checkCall("reject", 1, func() { worldID.SetContactTuning(bad, 10.0, 3.0) })
		checkCall("reject", 1, func() { worldID.SetContactTuning(30.0, bad, 3.0) })
		checkCall("reject", 1, func() { worldID.SetContactTuning(30.0, 10.0, bad) })
		checkCall("reject", 1, func() { worldID.Step(bad, 4) })
		trace.read(r, worldID, fmt.Sprintf("bad.%d.hash", i), true)
	}

	checkCall("reject", 1, func() { worldID.SetRestitutionIterations(-1) })

	checkCall("accept", 0, func() { worldID.Step(1.0/60.0, 4) })
	trace.read(r, worldID, "step.hash", true)
	gravityStep := worldID.Gravity()
	r.Vec2("step.gravity", gravityStep.X, gravityStep.Y)

	{
		ok := worldID.Gravity().X == gravity.X && worldID.Gravity().Y == gravity.Y
		r.Bool("check.115", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.RestitutionThreshold() == restitutionThreshold
		r.Bool("check.116", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.HitEventThreshold() == hitEventThreshold
		r.Bool("check.117", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.ContactRecycleDistance() == recycleDistance
		r.Bool("check.118", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.MaximumLinearSpeed() == maximumLinearSpeed
		r.Bool("check.119", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.RestitutionIterations() == restitutionIterations
		r.Bool("check.120", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}

	worldID.Close()
}

func checkBodyInputTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "invalid_input/BodyInputTest")
	trace := transformTrace{}
	rejectCount := 0

	checks := inputCallChecks{t: t, r: r, count: &rejectCount}
	checkCall := checks.check
	worldDef := box2d.DefaultWorldDef()
	worldID := box2d.NewWorld(&worldDef)
	checks.world = worldID
	box2d.CleanupWorld(t, worldID)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyID := worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	rejectCount = 0

	for _, bad := range badInputValues {
		badVec := box2d.Vec2{X: bad, Y: 0.0}
		badPos := box2d.Pos{X: oracle.Real(bad), Y: 0.0}
		badRot := box2d.Rot{C: bad, S: 0.0}
		badTarget := box2d.WorldTransform{P: badPos, Q: box2d.RotIdentity}

		checkCall("reject", 1, func() { bodyID.SetLinearVelocity(badVec) })
		checkCall("reject", 1, func() { bodyID.SetAngularVelocity(bad) })
		checkCall("reject", 1, func() { bodyID.SetTransform(badPos, box2d.RotIdentity) })
		checkCall("reject", 1, func() { bodyID.SetTransform(box2d.PosZero, badRot) })
		checkCall("reject", 1, func() { bodyID.SetTargetTransform(badTarget, 1.0/60.0, true) })
		checkCall("reject", 1, func() { bodyID.SetTargetTransform(box2d.WorldTransformIdentity, bad, true) })
		checkCall("reject", 1, func() { bodyID.ApplyForce(badVec, box2d.PosZero, true) })
		checkCall("reject", 1, func() { bodyID.ApplyForce(box2d.Vec2Zero, badPos, true) })
		checkCall("reject", 1, func() { bodyID.ApplyForceToCenter(badVec, true) })
		checkCall("reject", 1, func() { bodyID.ApplyTorque(bad, true) })
		checkCall("reject", 1, func() { bodyID.ApplyLinearImpulse(badVec, box2d.PosZero, true) })
		checkCall("reject", 1, func() { bodyID.ApplyLinearImpulse(box2d.Vec2Zero, badPos, true) })
		checkCall("reject", 1, func() { bodyID.ApplyLinearImpulseToCenter(badVec, true) })
		checkCall("reject", 1, func() { bodyID.ApplyAngularImpulse(bad, true) })
		checkCall("reject", 1, func() { bodyID.SetMassData(box2d.MassData{Mass: bad, Center: box2d.Vec2Zero, RotationalInertia: 1.0}) })
		checkCall("reject", 1, func() { bodyID.SetMassData(box2d.MassData{Mass: 1.0, Center: badVec, RotationalInertia: 1.0}) })
		checkCall("reject", 1, func() { bodyID.SetMassData(box2d.MassData{Mass: 1.0, Center: box2d.Vec2Zero, RotationalInertia: bad}) })
		checkCall("reject", 1, func() { bodyID.SetLinearDamping(bad) })
		checkCall("reject", 1, func() { bodyID.SetAngularDamping(bad) })
		checkCall("reject", 1, func() { bodyID.SetGravityScale(bad) })
		checkCall("reject", 1, func() { bodyID.SetSleepThreshold(bad) })
		checkCall("reject", 1, func() { bodyID.SetSafetyFactor(bad) })
	}

	// A negative value on a quantity the API documents as non-negative is turned away too,
	// since negative damping grows velocity without bound.
	checkCall("reject", 1, func() { bodyID.SetLinearDamping(-1.0) })
	checkCall("reject", 1, func() { bodyID.SetAngularDamping(-1.0) })

	checkCall("accept", 0, func() { bodyID.SetLinearVelocity(box2d.Vec2{X: 1.0, Y: 2.0}) })
	checkCall("accept", 0, func() { bodyID.ApplyForceToCenter(box2d.Vec2{X: 0.0, Y: 10.0}, true) })

	for i := range 30 {
		worldID.Step(1.0/60.0, 4)
		inputReadBodyStep(r, &trace, worldID, bodyID, i)
	}

	{
		ok := inputIsBodyFinite(bodyID)
		r.Bool("check.188", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}

	worldID.Close()
}

func checkShapeInputTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "invalid_input/ShapeInputTest")
	trace := transformTrace{}
	rejectCount := 0

	checks := inputCallChecks{t: t, r: r, count: &rejectCount}
	checkCall := checks.check
	worldDef := box2d.DefaultWorldDef()
	worldID := box2d.NewWorld(&worldDef)
	checks.world = worldID
	box2d.CleanupWorld(t, worldID)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyID := worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	shapeID := bodyID.CreatePolygonShape(&shapeDef, &box)

	density := float32(shapeID.Density())
	friction := float32(shapeID.Friction())
	restitution := float32(shapeID.Restitution())

	rejectCount = 0

	for _, bad := range badInputValues {
		badVec := box2d.Vec2{X: bad, Y: 0.0}

		badCircle := box2d.Circle{Center: badVec, Radius: 1.0}
		badCapsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: bad}
		badSegment := box2d.Segment{Point1: badVec, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}
		badChainSegment := box2d.ChainSegment{Ghost1: badVec, Segment: box2d.Segment{Point1: box2d.Vec2{X: 0.0, Y: 0.0}, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}, Ghost2: box2d.Vec2{X: 2.0, Y: 0.0}, ChainID: -1}

		badPolygon := box
		badPolygon.Radius = bad

		badVertexPolygon := box
		badVertexPolygon.Vertices[0] = badVec

		badMaterial := box2d.DefaultSurfaceMaterial()
		badMaterial.Friction = bad

		checkCall("reject", 1, func() { shapeID.SetDensity(bad, true) })
		checkCall("reject", 1, func() { shapeID.SetFriction(bad) })
		checkCall("reject", 1, func() { shapeID.SetRestitution(bad) })
		checkCall("reject", 1, func() { shapeID.SetSurfaceMaterial(&badMaterial) })
		checkCall("reject", 1, func() { shapeID.ApplyWind(badVec, 1.0, 1.0, true) })
		checkCall("reject", 1, func() { shapeID.ApplyWind(box2d.Vec2Zero, bad, 1.0, true) })
		checkCall("reject", 1, func() { shapeID.ApplyWind(box2d.Vec2Zero, 1.0, bad, true) })
		checkCall("reject", 1, func() { shapeID.SetCircle(&badCircle) })
		checkCall("reject", 1, func() { shapeID.SetCapsule(&badCapsule) })
		checkCall("reject", 1, func() { shapeID.SetSegment(&badSegment) })
		checkCall("reject", 1, func() { shapeID.SetChainSegment(&badChainSegment) })
		checkCall("reject", 1, func() { shapeID.SetPolygon(&badPolygon) })
		checkCall("reject", 1, func() { shapeID.SetPolygon(&badVertexPolygon) })
	}

	checkCall("accept", 0, func() { shapeID.SetFriction(0.4) })

	for i := range 10 {
		worldID.Step(1.0/60.0, 4)
		inputReadBodyStep(r, &trace, worldID, bodyID, i)
	}

	// The rejected geometry never reached the shape, so the proxy is still the original box
	{
		ok := shapeID.Type() == box2d.PolygonShape
		r.Bool("check.257", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := shapeID.AABB().IsValid()
		r.Bool("check.258", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.Bounds().IsValid()
		r.Bool("check.259", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := shapeID.Density() == density
		r.Bool("check.260", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := shapeID.Friction() == 0.4 && friction == friction
		r.Bool("check.261", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := shapeID.Restitution() == restitution
		r.Bool("check.262", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := inputIsBodyFinite(bodyID)
		r.Bool("check.263", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}

	worldID.Close()
}

func checkJointInputTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "invalid_input/JointInputTest")
	trace := transformTrace{}
	rejectCount := 0

	checks := inputCallChecks{t: t, r: r, count: &rejectCount}
	checkCall := checks.check
	worldDef := box2d.DefaultWorldDef()
	worldID := box2d.NewWorld(&worldDef)
	checks.world = worldID
	box2d.CleanupWorld(t, worldID)

	bodyDef := box2d.DefaultBodyDef()
	groundID := worldID.CreateBody(&bodyDef)

	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
	bodyID := worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	jointDef := box2d.DefaultRevoluteJointDef()
	jointDef.Base.BodyA = groundID
	jointDef.Base.BodyB = bodyID
	jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 1.0}
	jointID := worldID.CreateRevoluteJoint(&jointDef)

	rejectCount = 0

	for _, bad := range badInputValues {
		badFrame := box2d.Transform{P: box2d.Vec2{X: bad, Y: 0.0}, Q: box2d.RotIdentity}

		checkCall("reject", 1, func() { jointID.SetLocalFrameA(badFrame) })
		checkCall("reject", 1, func() { jointID.SetLocalFrameB(badFrame) })
		checkCall("reject", 1, func() { jointID.SetConstraintTuning(bad, 1.0) })
		checkCall("reject", 1, func() { jointID.SetConstraintTuning(60.0, bad) })
		checkCall("reject", 1, func() { jointID.SetForceThreshold(bad) })
		checkCall("reject", 1, func() { jointID.SetTorqueThreshold(bad) })
		checkCall("reject", 1, func() { jointID.SetSpringHertz(bad) })
		checkCall("reject", 1, func() { jointID.SetSpringDampingRatio(bad) })
		checkCall("reject", 1, func() { jointID.SetTargetAngle(bad) })
		checkCall("reject", 1, func() { jointID.SetLimits(bad, 1.0) })
		checkCall("reject", 1, func() { jointID.SetLimits(-1.0, bad) })
		checkCall("reject", 1, func() { jointID.SetMotorSpeed(bad) })
		checkCall("reject", 1, func() { jointID.SetMaxMotorTorque(bad) })
		inputReadRejectedJoint(r, jointID)
	}

	checkCall("accept", 0, func() { jointID.SetMotorSpeed(2.0) })
	inputReadRejectedJoint(r, jointID)

	for i := range 30 {
		worldID.Step(1.0/60.0, 4)
		inputReadBodyStep(r, &trace, worldID, bodyID, i)
	}

	{
		ok := inputIsBodyFinite(bodyID)
		r.Bool("check.322", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := jointID.ConstraintForce().IsValid()
		r.Bool("check.323", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := box2d.IsValidFloat(jointID.ConstraintTorque())
		r.Bool("check.324", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}

	worldID.Close()
}

func checkCreateInputTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "invalid_input/CreateInputTest")
	trace := transformTrace{}
	rejectCount := 0

	checks := inputCallChecks{t: t, r: r, count: &rejectCount}
	checkCall := checks.check
	worldDef := box2d.DefaultWorldDef()
	worldID := box2d.NewWorld(&worldDef)
	checks.world = worldID
	box2d.CleanupWorld(t, worldID)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyID := worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	rejectCount = 0

	for i, bad := range badInputValues {
		badVec := box2d.Vec2{X: bad, Y: 0.0}
		constructionHash := worldID.StateHash()

		badWorldDef := box2d.DefaultWorldDef()
		badWorldDef.Gravity = badVec
		{
			ok := rejectedInput(&rejectCount, func() { box2d.NewWorld(&badWorldDef) })
			r.Bool("check.353", ok)

			if !ok {
				t.Fatal("condition false")
			}
		}

		badBodyDef := box2d.DefaultBodyDef()
		badBodyDef.Position = box2d.Pos{X: oracle.Real(bad), Y: 0.0}

		checks.construction("check.357", "check.357.result", constructionHash, "failed body creation changed state", func() { worldID.CreateBody(&badBodyDef) })

		badShapeDef := box2d.DefaultShapeDef()
		badShapeDef.Density = bad

		checks.construction("check.361", "check.361.result", constructionHash, "failed shape creation changed state", func() { bodyID.CreatePolygonShape(&badShapeDef, &box) })

		badCircle := box2d.Circle{Center: badVec, Radius: 1.0}

		checks.construction("check.364", "check.364.result", constructionHash, "failed circle creation changed state", func() { bodyID.CreateCircleShape(&shapeDef, &badCircle) })

		badCapsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: bad}

		checks.construction("check.367", "check.367.result", constructionHash, "failed capsule creation changed state", func() { bodyID.CreateCapsuleShape(&shapeDef, &badCapsule) })

		badSegment := box2d.Segment{Point1: badVec, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}

		checks.construction("check.370", "check.370.result", constructionHash, "failed segment creation changed state", func() { bodyID.CreateSegmentShape(&shapeDef, &badSegment) })

		badPolygon := box
		badPolygon.Radius = bad

		checks.construction("check.374", "check.374.result", constructionHash, "failed polygon creation changed state", func() { bodyID.CreatePolygonShape(&shapeDef, &badPolygon) })

		chainPoints := []box2d.Vec2{{X: -4.0, Y: 0.0}, {X: 0.0, Y: 0.0}, {X: 4.0, Y: 0.0}, {X: 8.0, Y: 0.0}}
		chainPoints[2] = badVec
		material := box2d.DefaultSurfaceMaterial()
		badChainDef := box2d.DefaultChainDef()
		badChainDef.Points = chainPoints
		badChainDef.Materials = []box2d.SurfaceMaterial{material}
		badChainDef.IsLoop = true

		checks.construction("check.385", "check.385.result", constructionHash, "failed chain creation changed state", func() { bodyID.CreateChain(&badChainDef) })

		badJointDef := box2d.DefaultRevoluteJointDef()
		badJointDef.Base.BodyA = bodyID
		badJointDef.Base.BodyB = worldID.CreateBody(&bodyDef)
		badJointDef.Hertz = bad
		constructionHash = worldID.StateHash()
		checks.construction("check.391", "check.391.result", constructionHash, "failed joint creation changed state", func() { worldID.CreateRevoluteJoint(&badJointDef) })

		badExplosionDef := box2d.DefaultExplosionDef()
		badExplosionDef.Radius = bad

		checkCall("reject", 1, func() { worldID.Explode(&badExplosionDef) })
		inputReadConstructionCounts(t, r, worldID, 2+i)
		inputIndependentWorld(t, r)
	}

	for i := range 10 {
		worldID.Step(1.0/60.0, 4)
		inputReadBodyStep(r, &trace, worldID, bodyID, i)
	}

	{
		ok := inputIsBodyFinite(bodyID)
		r.Bool("check.405", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := worldID.Bounds().IsValid()
		r.Bool("check.406", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}

	worldID.Close()
}

func inputCheckRecordReplayTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "invalid_input/RecordReplayTest")
	trace := transformTrace{}
	rejectCount := 0

	checks := inputCallChecks{t: t, r: r, count: &rejectCount}
	checkCall := checks.check
	worldDef := box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{X: 0.0, Y: -10.0}
	worldDef.WorkerCount = 1
	worldID := box2d.NewWorld(&worldDef)
	checks.world = worldID
	box2d.CleanupWorld(t, worldID)

	recording := box2d.NewRecording(0)
	worldID.StartRecording(recording)

	bodyDef := box2d.DefaultBodyDef()
	groundID := worldID.CreateBody(&bodyDef)

	shapeDef := box2d.DefaultShapeDef()
	ground := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
	groundID.CreateSegmentShape(&shapeDef, &ground)

	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 4.0}
	bodyID := worldID.CreateBody(&bodyDef)

	box := box2d.MakeBox(0.5, 0.5)
	bodyID.CreatePolygonShape(&shapeDef, &box)

	rejectCount = 0

	// Interleave turned away calls with real ones. A rejected call panics before it is recorded,
	// so the op stream carries only the accepted calls.
	for i := range 30 {
		bad := float32(badInputValues[i%len(badInputValues)])

		checkCall("reject", 1, func() { bodyID.ApplyForceToCenter(box2d.Vec2{X: bad, Y: 0.0}, true) })
		checkCall("reject", 1, func() { bodyID.SetAngularVelocity(bad) })

		badShapeDef := box2d.DefaultShapeDef()
		badShapeDef.Density = bad
		constructionHash := worldID.StateHash()
		checks.construction("check.449", "check.449.result", constructionHash, "recorded failed creation changed state", func() { bodyID.CreatePolygonShape(&badShapeDef, &box) })

		bodyID.ApplyForceToCenter(box2d.Vec2{X: 0.0, Y: 1.0}, true)
		worldID.Step(1.0/60.0, 4)
		inputReadBodyStep(r, &trace, worldID, bodyID, i)
	}

	{
		ok := inputIsBodyFinite(bodyID)
		r.Bool("check.455", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}

	worldID.StopRecording()
	worldID.Close()

	data := recording.Data()
	size := recording.Size()
	{
		ok := size > 0
		r.Bool("check.462", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := box2d.ValidateReplay(data, 0) == nil

		r.I32("check.463", int32(rejectCount))
		r.Bool("check.463.result", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
	{
		ok := box2d.ValidateReplay(data, 4) == nil

		r.I32("check.464", int32(rejectCount))
		r.Bool("check.464.result", ok)

		if !ok {
			t.Fatal("condition false")
		}
	}
}

func (c *inputCallChecks) construction(key, resultKey string, before uint64, changedMessage string, call func()) {
	c.t.Helper()
	ok := rejectedInput(c.count, call)
	c.r.I32(key, int32(*c.count))
	c.r.Bool(resultKey, ok)
	c.r.U64("construction.state", c.world.StateHash())

	if c.world.StateHash() != before {
		c.t.Fatal(changedMessage)
	}

	if !ok {
		c.t.Fatal("condition false")
	}
}

func inputConstructionMembership(t *testing.T, world *box2d.World) (int, int) {
	t.Helper()

	shapeCount, jointCount := 0, 0

	for _, body := range box2d.BodiesInSlotOrder(world) {
		for shape := range body.Shapes() {
			if !shape.IsValid() || shape.Body() != body {
				t.Fatal("invalid shape membership after rejection")
			}

			shapeCount++
		}

		for joint := range body.Joints() {
			if !joint.IsValid() {
				t.Fatal("invalid joint membership after rejection")
			}

			jointCount++
		}
	}

	return shapeCount, jointCount
}
