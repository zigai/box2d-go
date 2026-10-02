// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
	"math"
	"testing"
)

var inputBadValues = []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))}

func inputIsBodyFinite(bodyID box2d.Body) bool {
	return box2d.IsValidPosition(bodyID.Position()) && box2d.IsValidRotation(bodyID.Rotation()) && box2d.IsValidVec2(bodyID.LinearVelocity()) && box2d.IsValidFloat(bodyID.AngularVelocity())
}

func inputReadBodyStep(r *oracle.Reader, trace *inputTransformTrace, worldID box2d.World, bodyID box2d.Body, step int) {
	trace.read(r, worldID, fmt.Sprintf("step.%d.hash", step), true)
	p := bodyID.Position()
	r.Pos(fmt.Sprintf("step.%d.position", step), p.X, p.Y)
	q := bodyID.Rotation()
	r.Rot(fmt.Sprintf("step.%d.rotation", step), q.C, q.S)
	v := bodyID.LinearVelocity()
	r.Vec2(fmt.Sprintf("step.%d.velocity", step), v.X, v.Y)
	r.F32(fmt.Sprintf("step.%d.angular", step), bodyID.AngularVelocity())
}

func TestInvalidInput(t *testing.T) {
	defer porttest.Guard(t)
	t.Run("WorldInputTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/invalid_input/WorldInputTest")
		trace := inputTransformTrace{}
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		logCount := 0
		beginRejectCount := func() {
			logCount = 0
			box2d.SetLogFunc(func(string) { logCount++ })
			box2d.SetAssertFunc(func(string, string, int) bool { return false })
		}
		endRejectCount := restore
		checkCall := func(kind string, want int, call func()) {
			t.Helper()
			before := logCount
			call()
			r.I32(kind, int32(logCount-before))
			porttest.Ensure(t, logCount == before+want)
		}
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		gravity := worldID.Gravity()
		restitutionThreshold := float32(worldID.RestitutionThreshold())
		hitEventThreshold := float32(worldID.HitEventThreshold())
		recycleDistance := float32(worldID.ContactRecycleDistance())
		maximumLinearSpeed := float32(worldID.MaximumLinearSpeed())
		restitutionIterations := worldID.RestitutionIterations()

		beginRejectCount()

		for i, bad := range inputBadValues {

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

		endRejectCount()

		{
			ok := worldID.Gravity().X == gravity.X && worldID.Gravity().Y == gravity.Y
			r.Bool("check.115", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := worldID.RestitutionThreshold() == restitutionThreshold
			r.Bool("check.116", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := worldID.HitEventThreshold() == hitEventThreshold
			r.Bool("check.117", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := worldID.ContactRecycleDistance() == recycleDistance
			r.Bool("check.118", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := worldID.MaximumLinearSpeed() == maximumLinearSpeed
			r.Bool("check.119", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := worldID.RestitutionIterations() == restitutionIterations
			r.Bool("check.120", ok)
			porttest.Ensure(t, ok)
		}

		worldID.Destroy()
	})
	t.Run("BodyInputTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/invalid_input/BodyInputTest")
		trace := inputTransformTrace{}
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		logCount := 0
		beginRejectCount := func() {
			logCount = 0
			box2d.SetLogFunc(func(string) { logCount++ })
			box2d.SetAssertFunc(func(string, string, int) bool { return false })
		}
		endRejectCount := restore
		checkCall := func(kind string, want int, call func()) {
			t.Helper()
			before := logCount
			call()
			r.I32(kind, int32(logCount-before))
			porttest.Ensure(t, logCount == before+want)
		}
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyID := worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(0.5, 0.5)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		beginRejectCount()

		for _, bad := range inputBadValues {
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

		for i := 0; i < 30; i++ {
			worldID.Step(1.0/60.0, 4)
			inputReadBodyStep(r, &trace, worldID, bodyID, i)
		}

		endRejectCount()

		{
			ok := inputIsBodyFinite(bodyID)
			r.Bool("check.188", ok)
			porttest.Ensure(t, ok)
		}

		worldID.Destroy()
	})
	t.Run("ShapeInputTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/invalid_input/ShapeInputTest")
		trace := inputTransformTrace{}
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		logCount := 0
		beginRejectCount := func() {
			logCount = 0
			box2d.SetLogFunc(func(string) { logCount++ })
			box2d.SetAssertFunc(func(string, string, int) bool { return false })
		}
		endRejectCount := restore
		checkCall := func(kind string, want int, call func()) {
			t.Helper()
			before := logCount
			call()
			r.I32(kind, int32(logCount-before))
			porttest.Ensure(t, logCount == before+want)
		}
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyID := worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(0.5, 0.5)
		shapeID := bodyID.CreatePolygonShape(&shapeDef, &box)

		density := float32(shapeID.Density())
		friction := float32(shapeID.Friction())
		restitution := float32(shapeID.Restitution())

		beginRejectCount()

		for _, bad := range inputBadValues {
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

		for i := 0; i < 10; i++ {
			worldID.Step(1.0/60.0, 4)
			inputReadBodyStep(r, &trace, worldID, bodyID, i)
		}

		endRejectCount()

		// The rejected geometry never reached the shape, so the proxy is still the original box
		{
			ok := shapeID.Type() == box2d.PolygonShape
			r.Bool("check.257", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.IsValidAABB(shapeID.AABB())
			r.Bool("check.258", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.IsValidAABB(worldID.Bounds())
			r.Bool("check.259", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := shapeID.Density() == density
			r.Bool("check.260", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := shapeID.Friction() == 0.4 && friction == friction
			r.Bool("check.261", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := shapeID.Restitution() == restitution
			r.Bool("check.262", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := inputIsBodyFinite(bodyID)
			r.Bool("check.263", ok)
			porttest.Ensure(t, ok)
		}

		worldID.Destroy()
	})
	t.Run("JointInputTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/invalid_input/JointInputTest")
		trace := inputTransformTrace{}
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		logCount := 0
		beginRejectCount := func() {
			logCount = 0
			box2d.SetLogFunc(func(string) { logCount++ })
			box2d.SetAssertFunc(func(string, string, int) bool { return false })
		}
		endRejectCount := restore
		checkCall := func(kind string, want int, call func()) {
			t.Helper()
			before := logCount
			call()
			r.I32(kind, int32(logCount-before))
			porttest.Ensure(t, logCount == before+want)
		}
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		groundID := worldID.CreateBody(&bodyDef)

		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.Pos{X: 0.0, Y: -1.0}
		bodyID := worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(0.5, 0.5)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		jointDef := box2d.DefaultRevoluteJointDef()
		jointDef.Base.BodyIDA = groundID
		jointDef.Base.BodyIDB = bodyID
		jointDef.Base.LocalFrameB.P = box2d.Vec2{X: 0.0, Y: 1.0}
		jointID := worldID.CreateRevoluteJoint(&jointDef)

		beginRejectCount()

		for _, bad := range inputBadValues {
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
		}

		checkCall("accept", 0, func() { jointID.SetMotorSpeed(2.0) })

		for i := 0; i < 30; i++ {
			worldID.Step(1.0/60.0, 4)
			inputReadBodyStep(r, &trace, worldID, bodyID, i)
		}

		endRejectCount()

		{
			ok := inputIsBodyFinite(bodyID)
			r.Bool("check.322", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.IsValidVec2(jointID.ConstraintForce())
			r.Bool("check.323", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.IsValidFloat(jointID.ConstraintTorque())
			r.Bool("check.324", ok)
			porttest.Ensure(t, ok)
		}

		worldID.Destroy()
	})
	t.Run("CreateInputTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/invalid_input/CreateInputTest")
		trace := inputTransformTrace{}
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		logCount := 0
		beginRejectCount := func() {
			logCount = 0
			box2d.SetLogFunc(func(string) { logCount++ })
			box2d.SetAssertFunc(func(string, string, int) bool { return false })
		}
		endRejectCount := restore
		checkCall := func(kind string, want int, call func()) {
			t.Helper()
			before := logCount
			call()
			r.I32(kind, int32(logCount-before))
			porttest.Ensure(t, logCount == before+want)
		}
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyID := worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		box := box2d.MakeBox(0.5, 0.5)
		bodyID.CreatePolygonShape(&shapeDef, &box)

		beginRejectCount()

		for _, bad := range inputBadValues {
			badVec := box2d.Vec2{X: bad, Y: 0.0}

			badWorldDef := box2d.DefaultWorldDef()
			badWorldDef.Gravity = badVec
			badWorldID := box2d.CreateWorld(&badWorldDef)
			box2d.BTWorldCleanup(t, badWorldID)
			{
				ok := badWorldID.IsValid() == false
				r.Bool("check.353", ok)
				porttest.Ensure(t, ok)
			}

			badBodyDef := box2d.DefaultBodyDef()
			badBodyDef.Position = box2d.Pos{X: oracle.Real(bad), Y: 0.0}
			{
				ok := worldID.CreateBody(&badBodyDef).IsValid() == false
				r.I32("check.357", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badShapeDef := box2d.DefaultShapeDef()
			badShapeDef.Density = bad
			{
				ok := bodyID.CreatePolygonShape(&badShapeDef, &box).IsValid() == false
				r.I32("check.361", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badCircle := box2d.Circle{Center: badVec, Radius: 1.0}
			{
				ok := bodyID.CreateCircleShape(&shapeDef, &badCircle).IsValid() == false
				r.I32("check.364", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badCapsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: bad}
			{
				ok := bodyID.CreateCapsuleShape(&shapeDef, &badCapsule).IsValid() == false
				r.I32("check.367", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badSegment := box2d.Segment{Point1: badVec, Point2: box2d.Vec2{X: 1.0, Y: 0.0}}
			{
				ok := bodyID.CreateSegmentShape(&shapeDef, &badSegment).IsValid() == false
				r.I32("check.370", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badPolygon := box
			badPolygon.Radius = bad
			{
				ok := bodyID.CreatePolygonShape(&shapeDef, &badPolygon).IsValid() == false
				r.I32("check.374", int32(logCount))
				porttest.Ensure(t, ok)
			}

			chainPoints := []box2d.Vec2{box2d.Vec2{X: -4.0, Y: 0.0}, box2d.Vec2{X: 0.0, Y: 0.0}, box2d.Vec2{X: 4.0, Y: 0.0}, box2d.Vec2{X: 8.0, Y: 0.0}}
			chainPoints[2] = badVec
			material := box2d.DefaultSurfaceMaterial()
			badChainDef := box2d.DefaultChainDef()
			badChainDef.Points = chainPoints
			badChainDef.PointCount = 4
			badChainDef.Materials = []box2d.SurfaceMaterial{material}
			badChainDef.MaterialCount = 1
			badChainDef.IsLoop = true
			{
				ok := bodyID.CreateChain(&badChainDef).IsValid() == false
				r.I32("check.385", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badJointDef := box2d.DefaultRevoluteJointDef()
			badJointDef.Base.BodyIDA = bodyID
			badJointDef.Base.BodyIDB = worldID.CreateBody(&bodyDef)
			badJointDef.Hertz = bad
			{
				ok := worldID.CreateRevoluteJoint(&badJointDef).IsValid() == false
				r.I32("check.391", int32(logCount))
				porttest.Ensure(t, ok)
			}

			badExplosionDef := box2d.DefaultExplosionDef()
			badExplosionDef.Radius = bad
			checkCall("reject", 1, func() { worldID.Explode(&badExplosionDef) })
		}

		for i := 0; i < 10; i++ {
			worldID.Step(1.0/60.0, 4)
			inputReadBodyStep(r, &trace, worldID, bodyID, i)
		}

		endRejectCount()

		{
			ok := inputIsBodyFinite(bodyID)
			r.Bool("check.405", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.IsValidAABB(worldID.Bounds())
			r.Bool("check.406", ok)
			porttest.Ensure(t, ok)
		}

		worldID.Destroy()
	})
	t.Run("RecordReplayTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/invalid_input/RecordReplayTest")
		trace := inputTransformTrace{}
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		logCount := 0
		beginRejectCount := func() {
			logCount = 0
			box2d.SetLogFunc(func(string) { logCount++ })
			box2d.SetAssertFunc(func(string, string, int) bool { return false })
		}
		endRejectCount := restore
		checkCall := func(kind string, want int, call func()) {
			t.Helper()
			before := logCount
			call()
			r.I32(kind, int32(logCount-before))
			porttest.Ensure(t, logCount == before+want)
		}
		worldDef := box2d.DefaultWorldDef()
		worldDef.Gravity = box2d.Vec2{X: 0.0, Y: -10.0}
		worldDef.WorkerCount = 1
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		recording := box2d.CreateRecording(0)
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

		beginRejectCount()

		// Interleave turned away calls with real ones so the op stream carries both. A mutator is
		// rejected before it is recorded, while a create is recorded with the null id it returned.
		for i := 0; i < 30; i++ {
			bad := float32(inputBadValues[i%len(inputBadValues)])

			checkCall("reject", 1, func() { bodyID.ApplyForceToCenter(box2d.Vec2{X: bad, Y: 0.0}, true) })
			checkCall("reject", 1, func() { bodyID.SetAngularVelocity(bad) })

			badShapeDef := box2d.DefaultShapeDef()
			badShapeDef.Density = bad
			{
				ok := bodyID.CreatePolygonShape(&badShapeDef, &box).IsValid() == false
				r.I32("check.449", int32(logCount))
				porttest.Ensure(t, ok)
			}

			bodyID.ApplyForceToCenter(box2d.Vec2{X: 0.0, Y: 1.0}, true)
			worldID.Step(1.0/60.0, 4)
			inputReadBodyStep(r, &trace, worldID, bodyID, i)
		}

		{
			ok := inputIsBodyFinite(bodyID)
			r.Bool("check.455", ok)
			porttest.Ensure(t, ok)
		}

		worldID.StopRecording()
		worldID.Destroy()

		data := recording.Data()
		size := recording.Size()
		{
			ok := size > 0
			r.Bool("check.462", ok)
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.ValidateReplay(data, 0)
			r.I32("check.463", int32(logCount))
			porttest.Ensure(t, ok)
		}
		{
			ok := box2d.ValidateReplay(data, 4)
			r.I32("check.464", int32(logCount))
			porttest.Ensure(t, ok)
		}

		endRejectCount()

		recording.Destroy()
	})
}
