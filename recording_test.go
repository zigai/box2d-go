// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"

	"github.com/zigai/box2d/internal/scenes"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var recordingOverlapCount int

func recordingOverlap(_ box2d.Shape) bool {
	recordingOverlapCount++
	return true
}

func recordingIssueAllQueries(worldID box2d.World, groundShapeID box2d.Shape) {
	filter := box2d.DefaultQueryFilter()

	baseOffset := box2d.Vec2{X: 3.0, Y: -2.0}
	base := box2d.ToPos(baseOffset)

	aabb := box2d.AABB{LowerBound: box2d.Vec2{X: -5.0 - baseOffset.X, Y: -15.0 - baseOffset.Y}, UpperBound: box2d.Vec2{X: 5.0 - baseOffset.X, Y: 5.0 - baseOffset.Y}}
	worldID.OverlapAABB(base, aabb, filter, recordingOverlap)

	proxy := box2d.MakeProxy([]box2d.Vec2{{X: -0.5, Y: -0.5}, {X: 0.5, Y: -0.5}, {X: 0.5, Y: 0.5}, {X: -0.5, Y: 0.5}}, 0)
	worldID.OverlapShape(box2d.Pos{}, &proxy, filter, recordingOverlap)

	rayOrigin := box2d.Pos{X: 0.0, Y: 10.0}
	rayDir := box2d.Vec2{X: 0.0, Y: -20.0}
	worldID.CastRay(rayOrigin, rayDir, filter, func(_ box2d.Shape, _ box2d.Pos, _ box2d.Vec2, fraction float32) float32 { return fraction })

	worldID.CastRayClosest(rayOrigin, rayDir, filter)

	circProxy := box2d.MakeProxy([]box2d.Vec2{{X: -baseOffset.X, Y: -baseOffset.Y}}, 0.3)
	worldID.CastShape(base, &circProxy, rayDir, filter, func(_ box2d.Shape, _ box2d.Pos, _ box2d.Vec2, fraction float32) float32 { return fraction })

	moverCap := box2d.Capsule{Center1: box2d.Vec2{X: -0.3 - baseOffset.X, Y: -baseOffset.Y}, Center2: box2d.Vec2{X: 0.3 - baseOffset.X, Y: -baseOffset.Y}, Radius: 0.5}
	worldID.CollideMover(base, &moverCap, filter, func(box2d.Shape, *box2d.PlaneResult) bool { return true })

	moverTranslation := box2d.Vec2{X: 0.0, Y: -5.0}
	worldID.CastMover(base, &moverCap, moverTranslation, filter)

	groundShapeID.TestPoint(box2d.Pos{X: 0.0, Y: -10.0})
	groundShapeID.TestPoint(box2d.Pos{X: 0.0, Y: 100.0})

	rayStart := box2d.OffsetPos(base, box2d.Vec2{X: -baseOffset.X, Y: 5.0 - baseOffset.Y})
	groundShapeID.RayCast(rayStart, box2d.Vec2{X: 0.0, Y: -20.0})
}

func recordingReplaySafetyFactor(player *box2d.Replay, name string) float32 {
	count := player.BodyCount()
	for i := range count {
		bodyID := player.BodyID(i)
		if bodyID.IsValid() && bodyID.Name() == name {
			return bodyID.SafetyFactor()
		}
	}

	return -1.0
}

func TestRecording(t *testing.T) {

	r := oracle.Open(t, "recording/RecordingTest")
	trace := newRecordingTrace(t, "RecordingTest")
	recPath := filepath.Join(t.TempDir(), "recording_test.b2rec")

	worldDef := box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{X: 0.0, Y: -10.0}
	worldDef.WorkerCount = 1

	worldID := box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldID)
	if !(worldID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 165", worldID.IsValid())

	rec := box2d.CreateRecording(0)
	worldID.StartRecording(rec)

	groundDef := box2d.DefaultBodyDef()
	groundDef.Position = box2d.Pos{X: 0.0, Y: -10.0}
	groundID := worldID.CreateBody(&groundDef)
	if !(groundID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 175", groundID.IsValid())

	groundShapeDef := box2d.DefaultShapeDef()
	groundCircle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 10.0}
	groundShapeID := groundID.CreateCircleShape(&groundShapeDef, &groundCircle)
	if !(groundShapeID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 180", groundShapeID.IsValid())

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 0.0, Y: 4.0}
	bodyDef.Name = "testBodyWithLongName"
	bodyID := worldID.CreateBody(&bodyDef)
	if !(bodyID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 189", bodyID.IsValid())

	shapeDef := box2d.DefaultShapeDef()
	shapeDef.Density = 1.0
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
	shapeID := bodyID.CreateCircleShape(&shapeDef, &circle)
	if !(shapeID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 195", shapeID.IsValid())

	box := box2d.MakeBox(0.25, 0.25)
	boxDef := box2d.DefaultShapeDef()
	boxDef.Density = 2.0
	boxShapeID := bodyID.CreatePolygonShape(&boxDef, &box)
	if !(boxShapeID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 202", boxShapeID.IsValid())

	capsuleBodyDef := box2d.DefaultBodyDef()
	capsuleBodyDef.Type = box2d.DynamicBody
	capsuleBodyDef.Position = box2d.Pos{X: 2.0, Y: 6.0}
	capsuleBodyID := worldID.CreateBody(&capsuleBodyDef)
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -0.5, Y: 0.0}, Center2: box2d.Vec2{X: 0.5, Y: 0.0}, Radius: 0.25}
	capsuleDef := box2d.DefaultShapeDef()
	capsuleDef.Density = 1.0
	capsuleShapeID := capsuleBodyID.CreateCapsuleShape(&capsuleDef, &capsule)
	if !(capsuleShapeID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 213", capsuleShapeID.IsValid())

	segment := box2d.Segment{Point1: box2d.Vec2{X: -20.0, Y: 0.0}, Point2: box2d.Vec2{X: 20.0, Y: 0.0}}
	segmentDef := box2d.DefaultShapeDef()
	segmentShapeID := groundID.CreateSegmentShape(&segmentDef, &segment)
	if !(segmentShapeID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 219", segmentShapeID.IsValid())

	chainSeg := box2d.ChainSegment{Ghost1: box2d.Vec2{X: -2.0, Y: 1.0}, Segment: box2d.Segment{Point1: box2d.Vec2{X: -1.0, Y: 1.0}, Point2: box2d.Vec2{X: 1.0, Y: 1.0}}, Ghost2: box2d.Vec2{X: 2.0, Y: 1.0}, ChainID: -1}
	chainSegShapeID := groundID.CreateChainSegmentShape(&segmentDef, &chainSeg)
	if !(chainSegShapeID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 224", chainSegShapeID.IsValid())

	boxShapeID.SetFriction(0.3)
	capsuleShapeID.SetRestitution(0.5)
	boxShapeID.SetDensity(3.0, true)
	boxShapeID.SetUserMaterial(0x1234)
	surface := box2d.DefaultSurfaceMaterial()
	surface.Friction = 0.7
	surface.Restitution = 0.1
	capsuleShapeID.SetSurfaceMaterial(&surface)
	filter := box2d.DefaultFilter()
	filter.CategoryBits = 0x2
	boxShapeID.SetFilter(filter)
	capsuleShapeID.EnableContactEvents(true)
	capsuleShapeID.EnableSensorEvents(true)
	boxShapeID.EnableHitEvents(true)
	boxShapeID.EnablePreSolveEvents(true)
	capsuleShapeID.ApplyWind(box2d.Vec2{X: 1.0, Y: 0.0}, 0.1, 0.0, true)

	newCircle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.4}
	shapeID.SetCircle(&newCircle)

	tmpCircle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.1}
	tmpShapeID := capsuleBodyID.CreateCircleShape(&capsuleDef, &tmpCircle)
	tmpShapeID.Destroy(true)

	kinematicDef := box2d.DefaultBodyDef()
	kinematicDef.Type = box2d.KinematicBody
	kinematicDef.Position = box2d.Pos{X: -3.0, Y: 5.0}
	kinematicID := worldID.CreateBody(&kinematicDef)
	kinematicShapeDef := box2d.DefaultShapeDef()
	kinematicCircle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.3}
	kinematicID.CreateCircleShape(&kinematicShapeDef, &kinematicCircle)

	disableDef := box2d.DefaultBodyDef()
	disableDef.Type = box2d.DynamicBody
	disableDef.Position = box2d.Pos{X: 5.0, Y: 5.0}
	disableDef.Name = "dropBody"
	disableID := worldID.CreateBody(&disableDef)
	disableCircle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.3}
	disableID.CreateCircleShape(&shapeDef, &disableCircle)

	fastDef := box2d.DefaultBodyDef()
	fastDef.Type = box2d.DynamicBody
	fastDef.Position = box2d.Pos{X: -14.0, Y: 30.0}
	fastDef.LinearVelocity = box2d.Vec2{X: 0.0, Y: -8.0}
	fastDef.SafetyFactor = 0.1
	fastDef.Name = "fastBody"
	fastID := worldID.CreateBody(&fastDef)
	if !(fastID.SafetyFactor() == 0.1) {
		t.Fatal("condition false")
	}
	r.Bool("check 281", fastID.SafetyFactor() == 0.1)
	r.F32("value 281", fastID.SafetyFactor())
	fastBox := box2d.MakeSquare(0.5)
	fastID.CreatePolygonShape(&shapeDef, &fastBox)

	bodyID.SetTransform(box2d.Pos{X: 1.0, Y: 5.0}, box2d.Rot{C: 1})
	bodyID.SetLinearVelocity(box2d.Vec2{X: 0.5, Y: 0.0})
	bodyID.SetAngularVelocity(0.25)
	bodyID.SetName("renamedBody")
	bodyID.SetLinearDamping(0.1)
	bodyID.SetAngularDamping(0.05)
	bodyID.SetGravityScale(0.9)
	bodyID.SetSleepThreshold(0.02)
	disableID.SetSafetyFactor(0.25)
	bodyID.EnableSleep(false)
	bodyID.SetBullet(true)
	bodyID.EnableContactRecycling(false)
	bodyID.EnableContactEvents(true)
	bodyID.EnableHitEvents(true)
	bodyID.SetMotionLocks(box2d.MotionLocks{LinearX: false, LinearY: false, AngularZ: true})
	md := box2d.MassData{Mass: 2.0, Center: box2d.Vec2{X: 0.0, Y: 0.0}, RotationalInertia: 0.5}
	bodyID.SetMassData(md)
	bodyID.UpdateMassFromShapes()
	capsuleBodyID.SetType(box2d.KinematicBody)
	capsuleBodyID.SetType(box2d.DynamicBody)
	kinematicID.SetTargetTransform(box2d.WorldTransform{P: box2d.Pos{X: -2.0, Y: 5.0}, Q: box2d.Rot{C: 1}}, 1.0/60.0, true)
	disableID.Disable()
	disableID.Enable()
	bodyID.SetAwake(true)
	bodyID.WakeTouching()

	bodyID.ApplyForce(box2d.Vec2{X: 0.0, Y: 50.0}, box2d.Pos{X: 1.0, Y: 5.0}, true)
	bodyID.ApplyForceToCenter(box2d.Vec2{X: 5.0, Y: 0.0}, true)
	bodyID.ApplyTorque(1.0, true)
	bodyID.ApplyLinearImpulse(box2d.Vec2{X: 0.1, Y: 0.0}, box2d.Pos{X: 1.0, Y: 5.0}, true)
	bodyID.ApplyLinearImpulseToCenter(box2d.Vec2{X: 0.0, Y: 0.1}, true)
	bodyID.ApplyAngularImpulse(0.05, true)

	chainBodyDef := box2d.DefaultBodyDef()
	chainBodyDef.Position = box2d.Pos{X: 0.0, Y: -2.0}
	chainBodyID := worldID.CreateBody(&chainBodyDef)
	chainPoints := []box2d.Vec2{{X: -8}, {X: -4}, {}, {X: 4}, {X: 8}, {X: 8, Y: 4}}
	chainMats := []box2d.SurfaceMaterial{box2d.DefaultSurfaceMaterial()}
	chainDef := box2d.DefaultChainDef()
	chainDef.Points = chainPoints
	chainDef.PointCount = 6
	chainDef.Ghost1 = box2d.Vec2{X: -12.0, Y: 0.0}
	chainDef.Ghost2 = box2d.Vec2{X: 8.0, Y: 8.0}
	chainDef.Materials = chainMats
	chainDef.MaterialCount = 1
	chainDef.IsLoop = false
	chainID := chainBodyID.CreateChain(&chainDef)
	if !(chainID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 335", chainID.IsValid())

	chainSurface := box2d.DefaultSurfaceMaterial()
	chainSurface.Friction = 0.4
	chainID.SetAllSurfaceMaterials(&chainSurface)

	chainSurface.Friction = 0.7
	chainID.SetSurfaceMaterial(&chainSurface, 3)

	tmpChainID := chainBodyID.CreateChain(&chainDef)
	tmpChainID.Destroy()

	var jb [8]box2d.Body
	for i := range 8 {
		jbd := box2d.DefaultBodyDef()
		jbd.Type = box2d.DynamicBody
		jbd.Position = box2d.ToPos(box2d.Vec2{X: -7.0 + float32(i), Y: 8.0})
		jb[i] = worldID.CreateBody(&jbd)
		jc := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.25}
		jb[i].CreateCircleShape(&shapeDef, &jc)
	}

	revDef := box2d.DefaultRevoluteJointDef()
	revDef.Base.BodyIDA = jb[0]
	revDef.Base.BodyIDB = jb[1]
	revDef.Base.LocalFrameA.P = box2d.Vec2{X: 0.5, Y: 0.0}
	revDef.Base.LocalFrameB.P = box2d.Vec2{X: -0.5, Y: 0.0}
	revID := worldID.CreateRevoluteJoint(&revDef)
	if !(revID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 367", revID.IsValid())
	revID.EnableLimit(true)
	revID.SetLimits(-1.0, 1.0)
	revID.EnableMotor(true)
	revID.SetMotorSpeed(0.5)
	revID.SetMaxMotorTorque(10.0)
	revID.EnableSpring(true)
	revID.SetSpringHertz(2.0)
	revID.SetSpringDampingRatio(0.5)
	revID.SetTargetAngle(0.25)
	revID.SetLocalFrameA(box2d.Transform{P: box2d.Vec2{X: 0.5, Y: 0.0}, Q: box2d.Rot{C: 1}})
	revID.SetLocalFrameB(box2d.Transform{P: box2d.Vec2{X: -0.5, Y: 0.0}, Q: box2d.Rot{C: 1}})
	revID.SetConstraintTuning(60.0, 2.0)
	revID.SetForceThreshold(100.0)
	revID.SetTorqueThreshold(50.0)
	revID.SetCollideConnected(false)
	revID.WakeBodies()

	distDef := box2d.DefaultDistanceJointDef()
	distDef.Base.BodyIDA = jb[1]
	distDef.Base.BodyIDB = jb[2]
	distDef.Length = 1.0
	distID := worldID.CreateDistanceJoint(&distDef)
	distID.SetLength(1.2)
	distID.EnableSpring(true)
	distID.SetSpringHertz(3.0)
	distID.SetSpringDampingRatio(0.4)
	distID.SetSpringForceRange(-50.0, 50.0)
	distID.EnableLimit(true)
	distID.SetLengthRange(0.5, 2.0)
	distID.EnableMotor(true)
	distID.SetMotorSpeed(0.3)
	distID.SetMaxMotorForce(5.0)

	prisDef := box2d.DefaultPrismaticJointDef()
	prisDef.Base.BodyIDA = jb[2]
	prisDef.Base.BodyIDB = jb[3]
	prisID := worldID.CreatePrismaticJoint(&prisDef)
	prisID.EnableSpring(true)
	prisID.SetSpringHertz(2.0)
	prisID.SetSpringDampingRatio(0.5)
	prisID.SetTargetTranslation(0.1)
	prisID.EnableLimit(true)
	prisID.SetLimits(-1.0, 1.0)
	prisID.EnableMotor(true)
	prisID.SetMotorSpeed(0.2)
	prisID.SetMaxMotorForce(8.0)

	wheelDef := box2d.DefaultWheelJointDef()
	wheelDef.Base.BodyIDA = jb[3]
	wheelDef.Base.BodyIDB = jb[4]
	wheelID := worldID.CreateWheelJoint(&wheelDef)
	wheelID.EnableSpring(true)
	wheelID.SetSpringHertz(4.0)
	wheelID.SetSpringDampingRatio(0.7)
	wheelID.EnableLimit(true)
	wheelID.SetLimits(-0.5, 0.5)
	wheelID.EnableMotor(true)
	wheelID.SetMotorSpeed(1.0)
	wheelID.SetMaxMotorTorque(6.0)

	weldDef := box2d.DefaultWeldJointDef()
	weldDef.Base.BodyIDA = jb[4]
	weldDef.Base.BodyIDB = jb[5]
	weldID := worldID.CreateWeldJoint(&weldDef)
	weldID.SetLinearHertz(5.0)
	weldID.SetLinearDampingRatio(0.6)
	weldID.SetAngularHertz(5.0)
	weldID.SetAngularDampingRatio(0.6)

	motorDef := box2d.DefaultMotorJointDef()
	motorDef.Base.BodyIDA = jb[5]
	motorDef.Base.BodyIDB = jb[6]
	motorID := worldID.CreateMotorJoint(&motorDef)
	motorID.SetLinearVelocity(box2d.Vec2{X: 0.1, Y: 0.0})
	motorID.SetAngularVelocity(0.2)
	motorID.SetMaxVelocityForce(10.0)
	motorID.SetMaxVelocityTorque(10.0)
	motorID.SetLinearHertz(2.0)
	motorID.SetLinearDampingRatio(0.5)
	motorID.SetAngularHertz(2.0)
	motorID.SetAngularDampingRatio(0.5)
	motorID.SetMaxSpringForce(20.0)
	motorID.SetMaxSpringTorque(20.0)

	filterDef := box2d.DefaultFilterJointDef()
	filterDef.Base.BodyIDA = jb[6]
	filterDef.Base.BodyIDB = jb[7]
	filterID := worldID.CreateFilterJoint(&filterDef)
	if !(filterID.IsValid()) {
		t.Fatal("condition false")
	}
	r.Bool("check 462", filterID.IsValid())

	tmpJointDef := box2d.DefaultDistanceJointDef()
	tmpJointDef.Base.BodyIDA = jb[0]
	tmpJointDef.Base.BodyIDB = jb[7]
	tmpJointDef.Length = 5.0
	tmpJointID := worldID.CreateDistanceJoint(&tmpJointDef)
	tmpJointID.Destroy()

	worldID.SetGravity(box2d.Vec2{X: 0.0, Y: -9.8})
	worldID.EnableSleeping(true)
	worldID.EnableContinuous(true)
	worldID.EnableWarmStarting(true)
	worldID.SetRestitutionThreshold(1.5)
	worldID.SetRestitutionIterations(2)
	worldID.EnableRestitutionPropagation(true)
	worldID.SetHitEventThreshold(2.0)
	worldID.SetContactTuning(30.0, 10.0, 3.0)
	worldID.SetContactRecycleDistance(0.05)
	worldID.SetMaximumLinearSpeed(100.0)
	worldID.RebuildStaticTree()
	explosion := box2d.DefaultExplosionDef()
	explosion.Position = box2d.Pos{X: 0.0, Y: 4.0}
	explosion.Radius = 3.0
	explosion.Falloff = 1.0
	explosion.ImpulsePerLength = 5.0
	worldID.Explode(&explosion)

	recordingIssueAllQueries(worldID, groundShapeID)

	const timeStep float32 = 1.0 / 60.0
	subStepCount := 4
	for i := range 60 {

		if i == 30 {
			capsuleBodyID.ApplyLinearImpulseToCenter(box2d.Vec2{X: 2.0, Y: 0.0}, true)
			bodyID.ClearForces()
			bodyID.SetGravityScale(1.0)
			fastID.SetSafetyFactor(0.4)
		}

		if i == 15 {
			recordingIssueAllQueries(worldID, groundShapeID)
		}

		worldID.Step(timeStep, subStepCount)
		trace.Check(t, r, worldID, "step 513", i+1)
	}

	worldID.StopRecording()
	worldID.Destroy()

	recData := rec.Data()
	recSize := rec.Size()
	if !(recSize > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 522", recSize > 0)
	check525 := box2d.ValidateReplay(recData, 0)
	if !(check525) {
		t.Fatal("condition false")
	}
	r.Bool("check 525", check525)
	check528 := box2d.ValidateReplay(recData, 4)
	if !(check528) {
		t.Fatal("condition false")
	}
	r.Bool("check 528", check528)

	{
		patched := make([]byte, recSize)
		copy(patched, recData)
		patched[8] = 0xAB
		patched[9] = 0xCD
		patched[10] = 0xEF
		patched[11] = 0x12
		patched[16] = 0x34
		check540 := box2d.ValidateReplay(patched, 0)
		if !(check540) {
			t.Fatal("condition false")
		}
		r.Bool("check 540", check540)

	}

	check545 := rec.SaveToFile(recPath) == nil
	if !(check545) {
		t.Fatal("condition false")
	}
	r.Bool("check 545", check545)
	loaded, err := box2d.LoadRecordingFromFile(recPath)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(loaded != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 547", loaded != nil)
	if !(loaded.Size() == recSize) {
		t.Fatal("condition false")
	}
	r.Bool("check 548", loaded.Size() == recSize)
	check549 := box2d.ValidateReplay(loaded.Data(), 0)
	if !(check549) {
		t.Fatal("condition false")
	}
	r.Bool("check 549", check549)
	check550 := box2d.ValidateReplay(loaded.Data(), 4)
	if !(check550) {
		t.Fatal("condition false")
	}
	r.Bool("check 550", check550)
	loaded.Destroy()

	{
		player, err := box2d.CreateReplay(recData, 0)
		destroyPlayer := box2d.BTReplayCleanup(t, player)
		if !(err == nil) {
			t.Fatal("condition false")
		}
		if !(player != nil) {
			t.Fatal("condition false")
		}
		r.Bool("check 557", player != nil)

		recBounds := player.Info().Bounds
		recExtents := recBounds.Extents()
		r.Vec2("bounds lower", recBounds.LowerBound.X, recBounds.LowerBound.Y)
		r.Vec2("bounds upper", recBounds.UpperBound.X, recBounds.UpperBound.Y)
		r.Vec2("bounds extents", recExtents.X, recExtents.Y)
		if !(recExtents.X > 0.0 && recExtents.Y > 0.0) {
			t.Fatal("condition false")
		}
		r.Bool("check 563", recExtents.X > 0.0 && recExtents.Y > 0.0)
		if !(recBounds.LowerBound.X <= -20.0 && recBounds.UpperBound.X >= 20.0) {
			t.Fatal("condition false")
		}
		r.Bool("check 564", recBounds.LowerBound.X <= -20.0 && recBounds.UpperBound.X >= 20.0)
		if !(recBounds.LowerBound.Y <= -20.0) {
			t.Fatal("condition false")
		}
		r.Bool("check 565", recBounds.LowerBound.Y <= -20.0)

		dd := box2d.DefaultDebugDraw()
		dd.DrawLineFunc = func(box2d.Pos, box2d.Pos, box2d.HexColor) {}
		dd.DrawPointFunc = func(box2d.Pos, float32, box2d.HexColor) {}
		dd.DrawPolygonFunc = func(box2d.WorldTransform, []box2d.Vec2, box2d.HexColor) {}
		dd.DrawSolidCapsuleFunc = func(box2d.Pos, box2d.Pos, float32, box2d.HexColor) {}

		frames := 0
		for player.StepFrame() {
			trace.Check(t, r, player.WorldID(), "step 576", player.Frame())
			if player.Frame() == 1 || player.Frame() == 16 || player.Frame() == 60 {
				recordingObserveQueries(t, r, player, fmt.Sprintf("inspect/%d", player.Frame()))
			}

			if frames%2 == 0 {
				player.DrawFrameQueries(&dd, -1)
			}
			frames += 1
		}
		if !(frames == 60) {
			t.Fatal("condition false")
		}
		r.Bool("check 584", frames == 60)
		r.I32("value 584", int32(frames))
		if !(player.Frame() == 60) {
			t.Fatal("condition false")
		}
		r.Bool("check 585", player.Frame() == 60)
		r.I32("value 585", int32(player.Frame()))
		if !(player.IsAtEnd()) {
			t.Fatal("condition false")
		}
		r.Bool("check 586", player.IsAtEnd())
		if !(player.HasDiverged() == false) {
			t.Fatal("condition false")
		}
		r.Bool("check 587", player.HasDiverged() == false)
		if !(player.WorldID().IsValid()) {
			t.Fatal("condition false")
		}
		r.Bool("check 591", player.WorldID().IsValid())

		player.Restart()
		if !(player.Frame() == 0) {
			t.Fatal("condition false")
		}
		r.Bool("check 595", player.Frame() == 0)
		r.I32("value 595", int32(player.Frame()))
		if !(player.IsAtEnd() == false) {
			t.Fatal("condition false")
		}
		r.Bool("check 596", player.IsAtEnd() == false)

		frames2 := 0
		for player.StepFrame() {
			trace.Check(t, r, player.WorldID(), "step 600", player.Frame())
			frames2 += 1
		}
		if !(frames2 == 60) {
			t.Fatal("condition false")
		}
		r.Bool("check 603", frames2 == 60)
		r.I32("value 603", int32(frames2))
		if !(player.HasDiverged() == false) {
			t.Fatal("condition false")
		}
		r.Bool("check 604", player.HasDiverged() == false)

		destroyPlayer()
	}

	{
		player, err := box2d.CreateReplay(recData, 0)
		destroyPlayer := box2d.BTReplayCleanup(t, player)
		if !(err == nil) {
			t.Fatal("condition false")
		}
		if !(player != nil) {
			t.Fatal("condition false")
		}
		r.Bool("check 613", player != nil)
		check616 := player.StepFrame()
		if !(check616) {
			t.Fatal("condition false")
		}
		r.Bool("check 616", check616)
		trace.Check(t, r, player.WorldID(), "step 616", player.Frame())
		if !(recordingReplaySafetyFactor(player, "fastBody") == 0.1) {
			t.Fatal("condition false")
		}
		r.Bool("check 617", recordingReplaySafetyFactor(player, "fastBody") == 0.1)
		r.F32("value 617", recordingReplaySafetyFactor(player, "fastBody"))
		if !(recordingReplaySafetyFactor(player, "dropBody") == 0.25) {
			t.Fatal("condition false")
		}
		r.Bool("check 618", recordingReplaySafetyFactor(player, "dropBody") == 0.25)
		r.F32("value 618", recordingReplaySafetyFactor(player, "dropBody"))

		for player.StepFrame() {
			trace.Check(t, r, player.WorldID(), "step 622", player.Frame())
		}
		if !(recordingReplaySafetyFactor(player, "fastBody") == 0.4) {
			t.Fatal("condition false")
		}
		r.Bool("check 624", recordingReplaySafetyFactor(player, "fastBody") == 0.4)
		r.F32("value 624", recordingReplaySafetyFactor(player, "fastBody"))

		destroyPlayer()
	}

	rec.Destroy()
	if !(os.Remove(recPath) == nil) {
		t.Fatal("condition false")
	}
	return
}

func TestRecordingOutliner(t *testing.T) {

	r := oracle.Open(t, "recording/RecordingOutlinerTest")
	trace := newRecordingTrace(t, "RecordingOutlinerTest")
	worldDef := box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{X: 0.0, Y: -10.0}
	worldDef.WorkerCount = 1
	worldID := box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldID)

	groundDef := box2d.DefaultBodyDef()
	groundID := worldID.CreateBody(&groundDef)
	groundShapeDef := box2d.DefaultShapeDef()
	groundCircle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 10.0}
	groundID.CreateCircleShape(&groundShapeDef, &groundCircle)

	dynamicCount := 3
	for i := range dynamicCount {
		bodyDef := box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.Position = box2d.ToPos(box2d.Vec2{X: float32(i), Y: 4.0})
		bodyID := worldID.CreateBody(&bodyDef)
		shapeDef := box2d.DefaultShapeDef()
		circle := box2d.Circle{Center: box2d.Vec2{X: 0.0, Y: 0.0}, Radius: 0.5}
		bodyID.CreateCircleShape(&shapeDef, &circle)
	}
	expectedBodies := 1 + dynamicCount

	worldID.Step(1.0/60.0, 4)
	trace.Check(t, r, worldID, "step 665", 1)

	rec := box2d.CreateRecording(0)
	worldID.StartRecording(rec)
	for i := range 10 {
		worldID.Step(1.0/60.0, 4)
		trace.Check(t, r, worldID, "step 671", i+2)
	}
	worldID.StopRecording()
	worldID.Destroy()

	recData := rec.Data()
	recSize := rec.Size()
	if !(recSize > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 678", recSize > 0)

	player, err := box2d.CreateReplay(recData, 0)
	destroyPlayer := box2d.BTReplayCleanup(t, player)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(player != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 681", player != nil)

	replayWorldID := player.WorldID()
	seedCount := player.BodyCount()
	if !(seedCount == expectedBodies) {
		t.Fatal("condition false")
	}
	r.Bool("check 687", seedCount == expectedBodies)
	r.I32("value 687", int32(seedCount))
	if !(seedCount == replayWorldID.Counters().BodyCount) {
		t.Fatal("condition false")
	}
	r.Bool("check 688", seedCount == replayWorldID.Counters().BodyCount)
	r.I32("value 688", int32(seedCount))

	for ord := range seedCount {
		if !(player.BodyID(ord).IsValid()) {
			t.Fatal("condition false")
		}
		r.Bool("check 693", player.BodyID(ord).IsValid())
	}

	for player.StepFrame() {
		trace.Check(t, r, player.WorldID(), "step 697", player.Frame())
	}

	player.Restart()
	if !(player.BodyCount() == seedCount) {
		t.Fatal("condition false")
	}
	r.Bool("check 702", player.BodyCount() == seedCount)
	r.I32("value 702", int32(player.BodyCount()))

	destroyPlayer()
	rec.Destroy()
	return
}

func recordingReplayDeepHash(player *box2d.Replay) uint64 {
	return box2d.RecordingHashWorldStateDeep(box2d.RecordingGetWorldFromID(player.WorldID()))
}

func recordingCheckKeyframeSeek(t *testing.T, r *oracle.Reader, trace *recordingTrace, recData []byte, recSize, workerCount, budgetBytes, minInterval int) int {

	ref, err := box2d.CreateReplay(recData, workerCount)
	destroyRef := box2d.BTReplayCleanup(t, ref)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(ref != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 725", ref != nil)
	frameCount := ref.Info().FrameCount
	if !(frameCount > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 727", frameCount > 0)

	refHash := make([]uint64, frameCount+1)
	if !(refHash != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 730", refHash != nil)
	refHash[0] = recordingReplayDeepHash(ref)
	r.U64("reference seed/hash", refHash[0])
	trace.Check(t, r, ref.WorldID(), "reference seed", 0)
	for f := 1; f <= frameCount; f++ {
		check734 := ref.StepFrame()
		if !(check734) {
			t.Fatal("condition false")
		}
		r.Bool("check 734", check734)
		trace.Check(t, r, ref.WorldID(), "step 734", ref.Frame())
		refHash[f] = recordingReplayDeepHash(ref)
	}
	if !(ref.HasDiverged() == false) {
		t.Fatal("condition false")
	}
	r.Bool("check 737", ref.HasDiverged() == false)
	destroyRef()

	player, err := box2d.CreateReplay(recData, workerCount)
	destroyPlayer := box2d.BTReplayCleanup(t, player)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(player != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 743", player != nil)
	if budgetBytes > 0 {
		player.SetKeyframePolicy(uint(budgetBytes), minInterval)
	}
	for player.StepFrame() {
		trace.Check(t, r, player.WorldID(), "step 749", player.Frame())
	}
	if !(player.Frame() == frameCount) {
		t.Fatal("condition false")
	}
	r.Bool("check 751", player.Frame() == frameCount)
	r.I32("value 751", int32(player.Frame()))

	targets := []int{frameCount, 1, frameCount - 1, 290, 17, 271, 256, 128, 33, 200, 5, 300, 100, frameCount}
	for i := range len(targets) {
		target := targets[i]
		if target > frameCount {
			target = frameCount
		}
		player.SeekFrame(target)
		if !(player.Frame() == target) {
			t.Fatal("condition false")
		}
		r.Bool("check 764", player.Frame() == target)
		r.I32("value 764", int32(player.Frame()))
		if !(player.HasDiverged() == false) {
			t.Fatal("condition false")
		}
		r.Bool("check 765", player.HasDiverged() == false)
		got := recordingReplayDeepHash(player)
		r.U64(fmt.Sprintf("seek 766/%d/hash", target), got)
		trace.Check(t, r, player.WorldID(), "seek 766", target)
		if got != refHash[target] {
			t.Logf("keyframe seek mismatch at frame %d (wc %d): got %d want %d\n", target, workerCount, got,
				refHash[target])

			destroyPlayer()
			return 1
		}
	}

	destroyPlayer()

	return 0
}

func TestRecordingKeyframe(t *testing.T) {

	r := oracle.Open(t, "recording/RecordingKeyframeTest")
	trace := newRecordingTrace(t, "RecordingKeyframeTest")
	worldDef := box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{X: 0.0, Y: -10.0}
	worldDef.WorkerCount = 1
	worldID := box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldID)

	groundDef := box2d.DefaultBodyDef()
	groundID := worldID.CreateBody(&groundDef)
	groundShapeDef := box2d.DefaultShapeDef()
	groundBox := box2d.MakeBox(20.0, 1.0)
	groundID.CreatePolygonShape(&groundShapeDef, &groundBox)

	for i := range 8 {
		bd := box2d.DefaultBodyDef()
		bd.Type = box2d.DynamicBody
		bd.Position = box2d.ToPos(box2d.Vec2{X: 0.05 * float32(i), Y: 2.0 + float32(1.1*float32(i))})
		id := worldID.CreateBody(&bd)
		sd := box2d.DefaultShapeDef()
		sd.Density = 1.0
		box := box2d.MakeBox(0.5, 0.5)
		id.CreatePolygonShape(&sd, &box)
	}

	rec := box2d.CreateRecording(0)
	worldID.StartRecording(rec)

	for i := range 320 {
		worldID.Step(1.0/60.0, 4)
		trace.Check(t, r, worldID, "step 819", i+1)
	}

	worldID.StopRecording()
	worldID.Destroy()

	recData := rec.Data()
	recSize := rec.Size()
	if !(recSize > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 827", recSize > 0)

	probe, err := box2d.CreateReplay(recData, 0)
	destroyProbe := box2d.BTReplayCleanup(t, probe)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(probe != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 832", probe != nil)
	snapSize := len(probe.WorldID().Snapshot())
	destroyProbe()
	if !(snapSize > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 835", snapSize > 0)
	tightBudget := 6 * snapSize
	check840 := recordingCheckKeyframeSeek(t, r, trace, recData, recSize, 0, 0, 0) == 0
	if !(check840) {
		t.Fatal("condition false")
	}
	r.Bool("check 840", check840)
	check841 := recordingCheckKeyframeSeek(t, r, trace, recData, recSize, 4, 0, 0) == 0
	if !(check841) {
		t.Fatal("condition false")
	}
	r.Bool("check 841", check841)
	check842 := recordingCheckKeyframeSeek(t, r, trace, recData, recSize, 0, tightBudget, 8) == 0
	if !(check842) {
		t.Fatal("condition false")
	}
	r.Bool("check 842", check842)
	check843 := recordingCheckKeyframeSeek(t, r, trace, recData, recSize, 4, tightBudget, 8) == 0
	if !(check843) {
		t.Fatal("condition false")
	}
	r.Bool("check 843", check843)

	rec.Destroy()
	return
}

func recordingBuildScrubPyramid(worldID box2d.World, baseCount int) {
	bd := box2d.DefaultBodyDef()
	bd.Position = box2d.Pos{X: 0.0, Y: -1.0}
	groundID := worldID.CreateBody(&bd)
	groundBox := box2d.MakeBox(40.0, 1.0)
	gsd := box2d.DefaultShapeDef()
	groundID.CreatePolygonShape(&gsd, &groundBox)

	const h float32 = 0.5
	pitch := 2.0*h + 0.05
	box := box2d.MakeBox(h, h)
	sd := box2d.DefaultShapeDef()
	sd.Density = 1.0

	for row := range baseCount {
		count := baseCount - row
		y := h + float32(float32(row)*pitch)
		xStart := float32(-0.5 * float32(count-1) * pitch)
		for col := range count {
			body := box2d.DefaultBodyDef()
			body.Type = box2d.DynamicBody
			body.Position = box2d.ToPos(box2d.Vec2{X: xStart + float32(float32(col)*pitch), Y: y})
			id := worldID.CreateBody(&body)
			id.CreatePolygonShape(&sd, &box)
		}
	}
}

func recordingBuildPyramidScene(worldID box2d.World) {
	recordingBuildScrubPyramid(worldID, 6)
}

func recordingIssuePileQueries(worldID box2d.World) {
	filter := box2d.DefaultQueryFilter()

	aabb := box2d.AABB{LowerBound: box2d.Vec2{X: -12.0, Y: -2.0}, UpperBound: box2d.Vec2{X: 12.0, Y: 22.0}}
	worldID.OverlapAABB(box2d.Pos{}, aabb, filter, recordingOverlap)

	worldID.CastRay(box2d.Pos{X: -12.0, Y: 10.0}, box2d.Vec2{X: 24.0, Y: 0.0}, filter, func(box2d.Shape, box2d.Pos, box2d.Vec2, float32) float32 { return 1 })
	worldID.CastRay(box2d.Pos{X: 0.0, Y: 22.0}, box2d.Vec2{X: 0.0, Y: -24.0}, filter, func(box2d.Shape, box2d.Pos, box2d.Vec2, float32) float32 { return 1 })
}

func recordingRecordSceneEx(t *testing.T, r *oracle.Reader, trace *recordingTrace, build func(box2d.World), workerCount, stepCount int, withQueries bool) *box2d.Recording {
	wd := box2d.DefaultWorldDef()
	wd.WorkerCount = workerCount
	worldID := box2d.CreateWorld(&wd)
	box2d.BTWorldCleanup(t, worldID)
	build(worldID)

	rec := box2d.CreateRecording(0)
	worldID.StartRecording(rec)
	for i := range stepCount {
		worldID.Step(1.0/60.0, 4)
		trace.Check(t, r, worldID, "step 925", i+1)
		if withQueries {
			recordingIssuePileQueries(worldID)
		}
	}
	worldID.StopRecording()
	worldID.Destroy()
	return rec
}

func recordingRecordScene(t *testing.T, r *oracle.Reader, trace *recordingTrace, build func(box2d.World), workerCount, stepCount int) *box2d.Recording {
	return recordingRecordSceneEx(t, r, trace, build, workerCount, stepCount, false)
}

func recordingCheckScrubAllFrames(t *testing.T, r *oracle.Reader, trace *recordingTrace, recData []byte, recSize, workerCount, budgetBytes, minInterval int) int {
	ref, err := box2d.CreateReplay(recData, workerCount)
	destroyRef := box2d.BTReplayCleanup(t, ref)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(ref != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 948", ref != nil)
	frameCount := ref.Info().FrameCount
	if !(frameCount > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 950", frameCount > 0)

	refHash := make([]uint64, frameCount+1)
	if !(refHash != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 953", refHash != nil)
	refHash[0] = recordingReplayDeepHash(ref)
	r.U64("reference seed/hash", refHash[0])
	trace.Check(t, r, ref.WorldID(), "reference seed", 0)
	for f := 1; f <= frameCount; f++ {
		check957 := ref.StepFrame()
		if !(check957) {
			t.Fatal("condition false")
		}
		r.Bool("check 957", check957)
		trace.Check(t, r, ref.WorldID(), "step 957", ref.Frame())
		refHash[f] = recordingReplayDeepHash(ref)
	}
	if !(ref.HasDiverged() == false) {
		t.Fatal("condition false")
	}
	r.Bool("check 960", ref.HasDiverged() == false)
	destroyRef()

	player, err := box2d.CreateReplay(recData, workerCount)
	destroyPlayer := box2d.BTReplayCleanup(t, player)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(player != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 964", player != nil)
	if budgetBytes > 0 {
		player.SetKeyframePolicy(uint(budgetBytes), minInterval)
	}
	for player.StepFrame() {
		trace.Check(t, r, player.WorldID(), "step 970", player.Frame())
	}

	for target := frameCount; target >= 0; target-- {
		player.SeekFrame(target)
		if !(player.Frame() == target) {
			t.Fatal("condition false")
		}
		r.Bool("check 976", player.Frame() == target)
		r.I32("value 976", int32(player.Frame()))
		got := recordingReplayDeepHash(player)
		r.U64(fmt.Sprintf("seek 977/%d/hash", target), got)
		trace.Check(t, r, player.WorldID(), "seek 977", target)
		diverged := player.HasDiverged()
		r.Bool(fmt.Sprintf("scrub %d/diverged", target), diverged)
		if got != refHash[target] || diverged {

			kind := "state"
			if got == refHash[target] {
				kind = "query-order"
			}
			t.Logf("scrub mismatch at frame %d (wc %d budget %d): %s divergence (deep got %d want %d)\n", target, workerCount,
				budgetBytes, kind, got, refHash[target])

			destroyPlayer()
			return 1
		}
	}

	destroyPlayer()
	return 0
}

func recordingScrubRecording(t *testing.T, r *oracle.Reader, trace *recordingTrace, rec *box2d.Recording, workerCount int) int {
	recData := rec.Data()
	recSize := rec.Size()
	if !(recSize > 0) {
		t.Fatal("condition false")
	}
	r.Bool("check 1003", recSize > 0)

	probe, err := box2d.CreateReplay(recData, 0)
	destroyProbe := box2d.BTReplayCleanup(t, probe)
	if !(err == nil) {
		t.Fatal("condition false")
	}
	if !(probe != nil) {
		t.Fatal("condition false")
	}
	r.Bool("check 1006", probe != nil)
	snapSize := len(probe.WorldID().Snapshot())
	destroyProbe()
	tightBudget := 4 * snapSize
	check1011 := recordingCheckScrubAllFrames(t, r, trace, recData, recSize, workerCount, tightBudget, 8) == 0
	if !(check1011) {
		t.Fatal("condition false")
	}
	r.Bool("check 1011", check1011)
	check1012 := recordingCheckScrubAllFrames(t, r, trace, recData, recSize, workerCount, 0, 0) == 0
	if !(check1012) {
		t.Fatal("condition false")
	}
	r.Bool("check 1012", check1012)
	return 0
}

func TestRecordingScrub(t *testing.T) {

	r := oracle.Open(t, "recording/RecordingScrubTest")
	trace := newRecordingTrace(t, "RecordingScrubTest")
	rec := recordingRecordScene(t, r, trace, recordingBuildPyramidScene, 1, 80)
	check1023 := recordingScrubRecording(t, r, trace, rec, 1) == 0
	if !(check1023) {
		t.Fatal("condition false")
	}
	r.Bool("check 1023", check1023)
	rec.Destroy()
	return
}

func TestRecordingQueryScrub(t *testing.T) {

	r := oracle.Open(t, "recording/RecordingQueryScrubTest")
	trace := newRecordingTrace(t, "RecordingQueryScrubTest")
	rec := recordingRecordSceneEx(t, r, trace, recordingBuildPyramidScene, 1, 80, true)
	check1034 := recordingScrubRecording(t, r, trace, rec, 1) == 0
	if !(check1034) {
		t.Fatal("condition false")
	}
	r.Bool("check 1034", check1034)
	rec.Destroy()
	return
}

func recordingRestepRaceStress(t *testing.T, r *oracle.Reader, trace *recordingTrace, build func(box2d.World), name string, workerCount, churnFrame, restepCount, reps int) int {
	const dt float32 = 1.0 / 60.0
	subSteps := 4

	wd := box2d.DefaultWorldDef()
	wd.WorkerCount = workerCount
	liveID := box2d.CreateWorld(&wd)
	box2d.BTWorldCleanup(t, liveID)
	build(liveID)
	for i := range churnFrame {
		liveID.Step(dt, subSteps)
		trace.Check(t, r, liveID, "step 1074", i+1)
	}

	image := liveID.Snapshot()
	liveID.Destroy()

	cd := box2d.DefaultWorldDef()
	cd.WorkerCount = workerCount
	cloneID := box2d.CreateWorld(&cd)
	box2d.BTWorldCleanup(t, cloneID)
	clone := box2d.RecordingGetWorldFromID(cloneID)

	var ref uint64
	result := 0
	for rep := range reps {
		if result != 0 {
			break
		}
		check1091 := cloneID.Restore(image) == nil
		if !(check1091) {
			t.Fatal("condition false")
		}
		r.Bool("check 1091", check1091)
		for s := range restepCount {
			cloneID.Step(dt, subSteps)
			trace.Check(t, r, cloneID, fmt.Sprintf("step 1094/rep %d", rep), churnFrame+s+1)
		}
		h := box2d.RecordingHashWorldStateDeep(clone)
		r.U64(fmt.Sprintf("restep 1096/%d/hash", rep), h)
		if rep == 0 {
			ref = h
		} else if h != ref {
			t.Logf("restep race: %s rep %d differs (wc %d churn %d restep %d): %d vs ref %d\n", name, rep, workerCount,
				churnFrame, restepCount, h, ref)
			result = 1
		}
	}

	cloneID.Destroy()

	return result
}

func TestReStepRace(t *testing.T) {

	r := oracle.Open(t, "recording/ReStepRaceTest")
	trace := newRecordingTrace(t, "ReStepRaceTest")
	check1123 := recordingRestepRaceStress(t, r, trace, scenes.CreateJunkyard, "junkyard", 2, 150, 50, 16) == 0
	if !(check1123) {
		t.Fatal("condition false")
	}
	r.Bool("check 1123", check1123)
	return
}

type recordingTraceRecord struct {
	key       string
	reference int
	hash      uint64
	hashSeen  bool
}

type recordingTrace struct {
	expected []recordingTraceRecord
	packets  [][]byte
	states   [][]byte
	next     int
	started  bool
	hashes   int
	rows     [][14]uint64
	rowNext  int
	reader   *oracle.Reader
	history  []byte
	previous [14]uint64
	run      uint32
}

type recordingTransformBits struct {
	index int
	bits  [4]uint64
}

func newRecordingTrace(t *testing.T, name string) *recordingTrace {
	t.Helper()
	config := "float"
	if oracle.DoublePrecision {
		config = "double"
	}
	data, err := os.ReadFile(filepath.Join("testdata", "oracle", config, "recording", name+".jsonl"))
	if !(err == nil) {
		t.Fatal("condition false")
	}
	trace := &recordingTrace{}
	schemaSeen, historySeen := false, false
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var record struct {
			Key   string `json:"k"`
			Type  string `json:"t"`
			Value string `json:"v"`
		}
		if !(json.Unmarshal(line, &record) == nil) {
			t.Fatal("condition false")
		}
		if record.Key == "trace/schema" {
			schema, err := hex.DecodeString(record.Value)
			if err != nil || string(schema) != recordingTraceSchema || schemaSeen {
				t.Fatalf("invalid recording trace schema: %v", err)
			}
			schemaSeen = true
		} else if record.Key == "trace/rle-u32-14u64-le" {
			packed, err := hex.DecodeString(record.Value)
			if err != nil || historySeen {
				t.Fatalf("invalid recording trace packet: %v", err)
			}
			trace.rows, err = recordingUnpackHistory(packed)
			if err != nil {
				t.Fatal(err)
			}
			historySeen = true
		}
		if strings.HasSuffix(record.Key, "/state/ref") {
			reference, err := strconv.ParseInt(record.Value, 16, 32)
			if !(err == nil && reference >= 0 && int(reference) <= len(trace.packets)) {
				t.Fatal("condition false")
			}
			trace.expected = append(trace.expected, recordingTraceRecord{key: record.Key, reference: int(reference)})
		} else if strings.HasSuffix(record.Key, "/state/packed") {
			packet, err := hex.DecodeString(record.Value)
			if !(err == nil) {
				t.Fatal("condition false")
			}
			trace.packets = append(trace.packets, packet)
		} else if record.Type == "u64" && strings.HasSuffix(record.Key, "/xf") {
			trace.hashes++
			if len(trace.expected) == 0 {
				continue
			}
			latest := &trace.expected[len(trace.expected)-1]
			if strings.TrimSuffix(latest.key, "/state/ref")+"/xf" == record.Key {
				hash, err := strconv.ParseUint(record.Value, 16, 64)
				if !(err == nil) {
					t.Fatal("condition false")
				}
				latest.hash = hash
				latest.hashSeen = true
			}
		}
	}
	if !schemaSeen || !historySeen {
		t.Fatal("missing complete named-field recording history")
	}
	trace.hashes = len(trace.rows)
	t.Cleanup(func() {

		if trace.started && !t.Skipped() && !t.Failed() {
			if trace.next != len(trace.expected) || trace.rowNext != len(trace.rows) {
				t.Fatalf("recording history incomplete: transform checkpoints %d/%d, frames %d/%d",
					trace.next, len(trace.expected), trace.rowNext, len(trace.rows))
			}
			trace.flushRun()
			trace.reader.Bytes("trace/schema", []byte(recordingTraceSchema))
			trace.reader.Bytes("trace/rle-u32-14u64-le", trace.history)
		}
	})
	return trace
}

const recordingTraceSchema = "transform,canonicalState,moveCount,moveHash,sensorBeginCount,sensorEndCount,sensorHash," +
	"contactBeginCount,contactEndCount,hitCount,contactHash,jointCount,jointHash,frameIdentity"

var recordingTraceFields = strings.Split(recordingTraceSchema, ",")

func recordingUnpackHistory(packed []byte) ([][14]uint64, error) {
	var rows [][14]uint64
	for len(packed) != 0 {
		if len(packed) < 116 {
			return nil, fmt.Errorf("incomplete recording RLE row")
		}
		count := binary.LittleEndian.Uint32(packed[:4])
		if count == 0 || count > 1<<20 || len(rows)+int(count) > 1<<20 {
			return nil, fmt.Errorf("invalid recording RLE count %d", count)
		}
		var row [14]uint64
		for field := range row {
			row[field] = binary.LittleEndian.Uint64(packed[4+field*8:])
		}
		for range count {
			rows = append(rows, row)
		}
		packed = packed[116:]
	}
	return rows, nil
}

func (trace *recordingTrace) flushRun() {
	if trace.run == 0 {
		return
	}
	trace.history = binary.LittleEndian.AppendUint32(trace.history, trace.run)
	for _, field := range trace.previous {
		trace.history = binary.LittleEndian.AppendUint64(trace.history, field)
	}
	trace.run = 0
}

func (trace *recordingTrace) compareRow(t *testing.T, r *oracle.Reader, prefix string, frame int, row [14]uint64) {
	t.Helper()
	trace.reader = r
	if trace.rowNext == len(trace.rows) {
		t.Fatalf("%s/%d: unexpected recording history frame %d", prefix, frame, trace.rowNext)
	}
	want := trace.rows[trace.rowNext]
	for field, name := range recordingTraceFields {
		if row[field] != want[field] {
			t.Fatalf("%s/%d: recording history frame ordinal %d field %s: C %016x, Go %016x",
				prefix, frame, trace.rowNext, name, want[field], row[field])
		}
	}
	trace.rowNext++
	if trace.run != 0 && trace.previous != row {
		trace.flushRun()
	}
	trace.previous = row
	trace.run++
}

func recordingPackTransforms(transforms []box2d.RecordingBodyTransform) []byte {
	width := byte(4)
	if oracle.DoublePrecision {
		width = 8
	}
	raw := binary.AppendUvarint([]byte{width}, uint64(len(transforms)))
	previousIndex := -1
	var previous [4]uint64
	for _, body := range transforms {
		raw = binary.AppendUvarint(raw, uint64(body.Index-previousIndex))
		previousIndex = body.Index
		xf := body.Transform
		bits := [4]uint64{uint64(math.Float32bits(float32(xf.P.X))), uint64(math.Float32bits(float32(xf.P.Y))), uint64(math.Float32bits(xf.Q.C)), uint64(math.Float32bits(xf.Q.S))}
		if oracle.DoublePrecision {
			bits[0] = math.Float64bits(float64(xf.P.X))
			bits[1] = math.Float64bits(float64(xf.P.Y))
		}
		for component := range bits {
			raw = binary.AppendUvarint(raw, bits[component]^previous[component])
			previous[component] = bits[component]
		}
	}
	return raw
}

func recordingLZKey(data []byte) int {
	return (int(data[0])*251 + int(data[1])*31 + int(data[2])) & 65535
}

func recordingCompress(raw []byte) []byte {
	packed := make([]byte, 0, len(raw))
	recent := make([]int, 65536)
	for i := range recent {
		recent[i] = -1
	}
	pos, literal := 0, 0
	for pos < len(raw) {
		length, distance := 0, 0
		if pos+2 < len(raw) {
			candidate := recent[recordingLZKey(raw[pos:])]
			if candidate >= 0 && pos-candidate <= 65535 {
				distance = pos - candidate
				for length < 130 && pos+length < len(raw) && raw[candidate+length] == raw[pos+length] {
					length++
				}
			}
		}
		if length >= 4 {
			if pos > literal {
				packed = append(packed, byte(pos-literal-1))
				packed = append(packed, raw[literal:pos]...)
			}
			packed = append(packed, byte(128+length-3), byte(distance), byte(distance>>8))
			for j := range length {
				if pos+j+2 < len(raw) {
					recent[recordingLZKey(raw[pos+j:])] = pos + j
				}
			}
			pos += length
			literal = pos
		} else {
			if pos+2 < len(raw) {
				recent[recordingLZKey(raw[pos:])] = pos
			}
			pos++
			if pos-literal == 128 || pos == len(raw) {
				packed = append(packed, byte(pos-literal-1))
				packed = append(packed, raw[literal:pos]...)
				literal = pos
			}
		}
	}
	return packed
}

func recordingDecompress(packed []byte) ([]byte, error) {
	var raw []byte
	for pos := 0; pos < len(packed); {
		control := packed[pos]
		pos++
		if control < 128 {
			length := int(control) + 1
			if length > len(packed)-pos {
				return nil, fmt.Errorf("truncated literal")
			}
			raw = append(raw, packed[pos:pos+length]...)
			pos += length
		} else {
			if len(packed)-pos < 2 {
				return nil, fmt.Errorf("truncated match")
			}
			length := int(control&127) + 3
			distance := int(packed[pos]) | int(packed[pos+1])<<8
			pos += 2
			if distance == 0 || distance > len(raw) {
				return nil, fmt.Errorf("invalid match distance %d", distance)
			}
			for range length {
				raw = append(raw, raw[len(raw)-distance])
			}
		}
		if len(raw) > 4<<20 {
			return nil, fmt.Errorf("decoded transform packet exceeds limit")
		}
	}
	return raw, nil
}

func recordingUnpackTransforms(raw []byte) ([]recordingTransformBits, error) {
	if len(raw) == 0 || (raw[0] != 4 && raw[0] != 8) {
		return nil, fmt.Errorf("invalid position width")
	}
	width := raw[0]
	raw = raw[1:]
	count, n := binary.Uvarint(raw)
	if n <= 0 || count > 1<<20 {
		return nil, fmt.Errorf("invalid body count")
	}
	raw = raw[n:]
	bodies := make([]recordingTransformBits, int(count))
	previousIndex := -1
	var previous [4]uint64
	for ordinal := range bodies {
		delta, n := binary.Uvarint(raw)
		if n <= 0 || delta == 0 || delta > math.MaxInt32 || previousIndex > math.MaxInt32-int(delta) {
			return nil, fmt.Errorf("invalid body index at ordinal %d", ordinal)
		}
		raw = raw[n:]
		previousIndex += int(delta)
		bodies[ordinal].index = previousIndex
		for component := range previous {
			xor, n := binary.Uvarint(raw)
			if n <= 0 {
				return nil, fmt.Errorf("invalid component at body %d component %d", ordinal, component)
			}
			raw = raw[n:]
			previous[component] ^= xor
			if (width == 4 || component >= 2) && previous[component] > math.MaxUint32 {
				return nil, fmt.Errorf("invalid float32 bits")
			}
			bodies[ordinal].bits[component] = previous[component]
		}
	}
	if len(raw) != 0 {
		return nil, fmt.Errorf("trailing transform bytes")
	}
	return bodies, nil
}

func (trace *recordingTrace) Check(t *testing.T, r *oracle.Reader, world box2d.World, prefix string, frame int) {
	t.Helper()
	trace.started = true
	transforms := box2d.RecordingBodyTransforms(world)
	if frame%60 == 0 {
		key := fmt.Sprintf("%s/%d/state/ref", prefix, frame)
		if !(trace.next < len(trace.expected)) {
			t.Fatal("condition false")
		}
		expected := trace.expected[trace.next]
		trace.next++
		if !(expected.key == key && expected.reference < len(trace.packets)) {
			t.Fatal("condition false")
		}
		wantRaw, err := recordingDecompress(trace.packets[expected.reference])
		if !(err == nil) {
			t.Fatal("condition false")
		}
		want, err := recordingUnpackTransforms(wantRaw)
		if !(err == nil) {
			t.Fatal("condition false")
		}
		raw := recordingPackTransforms(transforms)
		got, err := recordingUnpackTransforms(raw)
		if !(err == nil) {
			t.Fatal("condition false")
		}
		if len(want) != len(got) {
			t.Fatalf("%s: body count: oracle %d, Go %d", key, len(want), len(got))
		}
		components := [4]string{"position.x", "position.y", "rotation.c", "rotation.s"}
		for ordinal := range want {
			if want[ordinal].index != got[ordinal].index {
				t.Fatalf("%s: body ordinal %d: oracle index %d, Go index %d", key, ordinal, want[ordinal].index, got[ordinal].index)
			}
			for component := range components {
				if want[ordinal].bits[component] != got[ordinal].bits[component] {
					t.Fatalf("%s: body ordinal %d index %d component %s: oracle bits %016x, Go bits %016x", key, ordinal, got[ordinal].index, components[component], want[ordinal].bits[component], got[ordinal].bits[component])
				}
			}
		}
		reference := 0
		for reference < len(trace.states) && !bytes.Equal(trace.states[reference], raw) {
			reference++
		}
		r.I32(key, int32(reference))
		if reference == len(trace.states) {
			trace.states = append(trace.states, raw)
			r.Bytes(fmt.Sprintf("%s/%d/state/packed", prefix, frame), recordingCompress(raw))
		}
	}
	hash := oracle.HashInit
	for _, body := range transforms {
		xf := body.Transform
		hash = oracle.HashTransform(hash, oracle.Real(xf.P.X), oracle.Real(xf.P.Y), xf.Q.C, xf.Q.S)
	}
	if frame%60 == 0 {
		r.U64(fmt.Sprintf("%s/%d/xf", prefix, frame), hash)
	}
	row := [14]uint64{hash, world.StateHash()}
	events := recordingEventValues(world)
	copy(row[2:], events[:])
	row[13] = oracle.HashBytes(oracle.HashInit, []byte(fmt.Sprintf("%s/%d", prefix, frame)))
	trace.compareRow(t, r, prefix, frame, row)
	if frame == 1 || frame == 31 || frame%60 == 0 {
		recordingProperties(r, world, fmt.Sprintf("%s/%d/properties", prefix, frame))
	}
}

func TestRecordingTransformPacking(t *testing.T) {

	transforms := []box2d.RecordingBodyTransform{
		{Index: 0, Transform: box2d.WorldTransform{P: box2d.Pos{X: -1, Y: 2}, Q: box2d.Rot{C: 1}}},
		{Index: 4, Transform: box2d.WorldTransform{P: box2d.Pos{X: 3, Y: -4}, Q: box2d.Rot{C: -1, S: math.Float32frombits(1 << 31)}}},
	}
	for i := range 1000 {
		transforms = append(transforms, box2d.RecordingBodyTransform{Index: 5 + i, Transform: transforms[i%2].Transform})
	}
	raw := recordingPackTransforms(transforms)
	packed := recordingCompress(raw)
	if !(len(packed) < len(raw)) {
		t.Fatal("condition false")
	}
	decoded, err := recordingDecompress(packed)
	if !(err == nil && bytes.Equal(raw, decoded)) {
		t.Fatal("condition false")
	}
	bits, err := recordingUnpackTransforms(decoded)
	if !(err == nil && len(bits) == len(transforms)) {
		t.Fatal("condition false")
	}
	if !(bits[0].index == 0 && bits[1].index == 4 && bits[1].bits[3] == 1<<31) {
		t.Fatal("condition false")
	}
	for _, invalid := range [][]byte{{0}, {128}, {128, 0, 0}, {128, 1, 0}} {
		_, err := recordingDecompress(invalid)
		if !(err != nil) {
			t.Fatal("condition false")
		}
	}
	for _, invalid := range [][]byte{{}, {1}, {4}, {4, 1}, {4, 0, 0}} {
		_, err := recordingUnpackTransforms(invalid)
		if !(err != nil) {
			t.Fatal("condition false")
		}
	}
	for _, name := range []string{"RecordingTest", "RecordingOutlinerTest", "RecordingKeyframeTest", "RecordingScrubTest", "RecordingQueryScrubTest", "ReStepRaceTest"} {
		t.Run(name, func(t *testing.T) {

			trace := newRecordingTrace(t, name)
			if !(trace.hashes > 0) {
				t.Fatal("condition false")
			}
			if name != "RecordingOutlinerTest" {
				if !(len(trace.expected) > 0 && len(trace.packets) > 0) {
					t.Fatal("condition false")
				}
			}
			for index, packet := range trace.packets {
				raw, err := recordingDecompress(packet)
				if !(err == nil) {
					t.Fatal("condition false")
				}
				if !(bytes.Equal(packet, recordingCompress(raw))) {
					t.Fatal("condition false")
				}
				_, err = recordingUnpackTransforms(raw)
				if err != nil {
					t.Fatalf("packet %d: %v", index, err)
				}
			}
			for _, record := range trace.expected {
				if !(record.hashSeen) {
					t.Fatal("condition false")
				}
				raw, err := recordingDecompress(trace.packets[record.reference])
				if !(err == nil) {
					t.Fatal("condition false")
				}
				bodies, err := recordingUnpackTransforms(raw)
				if !(err == nil) {
					t.Fatal("condition false")
				}
				hash := oracle.HashInit
				for _, body := range bodies {
					for component, bits := range body.bits {
						if component < 2 && raw[0] == 8 {
							hash = oracle.HashU64(hash, bits)
						} else {
							hash = oracle.HashU32(hash, uint32(bits))
						}
					}
				}
				if hash != record.hash {
					t.Fatalf("%s: unpacked transforms hash %016x, C stream hash %016x", record.key, hash, record.hash)
				}
			}
		})
	}
}

func TestRecordingHistoryPacking(t *testing.T) {
	// Hand-authored little-endian run: two frames, transform=1,
	// canonicalState=128, all eleven event fields and frame identity zero.
	packed := make([]byte, 116)
	packed[0], packed[4], packed[12] = 2, 1, 128
	rows, err := recordingUnpackHistory(packed)
	want := [14]uint64{1, 128}
	if err != nil || len(rows) != 2 || rows[0] != want || rows[1] != want {
		t.Fatalf("decoding literal recording run: rows %v, error %v", rows, err)
	}
	trace := &recordingTrace{previous: want, run: 2}
	trace.flushRun()
	if !bytes.Equal(trace.history, packed) || trace.run != 0 {
		t.Fatalf("encoding literal recording run: got %x, want %x", trace.history, packed)
	}
	zeroCount := make([]byte, 116)
	excessCount := make([]byte, 116)
	excessCount[3] = 1
	for _, invalid := range [][]byte{packed[:115], append(append([]byte(nil), packed...), 0), zeroCount, excessCount} {
		if _, err := recordingUnpackHistory(invalid); err == nil {
			t.Fatalf("accepted invalid recording run %x", invalid)
		}
	}
}
