// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"

	"github.com/zigai/box2d/internal/scenes"
)

func inputSingleMultithreading(t *testing.T, r *oracle.Reader, trace *inputTransformTrace, workerCount int) int {
	t.Helper()

	worldDef := box2d.DefaultWorldDef()
	worldDef.WorkerCount = workerCount
	worldID := box2d.NewWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldID)
	data := scenes.CreateFallingHinges(worldID)
	timeStep := float32(1.0 / 60.0)

	stepLimit := 500
	for range stepLimit {
		subStepCount := 4
		worldID.Step(timeStep, subStepCount)
		done := scenes.UpdateFallingHinges(worldID, &data)
		trace.read(r, worldID, fmt.Sprintf("w%d.s%d", workerCount, data.StepCount), data.StepCount%60 == 0)

		if done {
			break
		}
	}

	worldID.Close()
	r.I32("check.49", int32(data.SleepStep))

	if data.SleepStep != scenes.ExpectedSleepStep {
		t.Fatal("condition false")
	}

	r.U32("check.50", data.Hash)

	if data.Hash != scenes.ExpectedHash {
		t.Fatal("condition false")
	}

	scenes.DestroyFallingHinges(&data)

	return 0
}

func TestDeterminism(t *testing.T) {
	// Test multithreaded determinism.
	t.Run("MultithreadingTest", func(t *testing.T) { inputCheckMultithreadingTest(t) })
	// Test determinism using the built-in scheduler (no external task system).
	t.Run("BuiltInSchedulerTest", func(t *testing.T) { inputCheckBuiltInSchedulerTest(t) })
	// Test cross-platform determinism.
	t.Run("CrossPlatformTest", func(t *testing.T) { inputCheckCrossPlatformTest(t) })
}

type inputTransformTrace struct {
	snapshots [][]box2d.WorldTransform
}

func inputSameTransforms(a, b []box2d.WorldTransform) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if math.Float64bits(float64(a[i].P.X)) != math.Float64bits(float64(b[i].P.X)) ||
			math.Float64bits(float64(a[i].P.Y)) != math.Float64bits(float64(b[i].P.Y)) ||
			math.Float32bits(a[i].Q.C) != math.Float32bits(b[i].Q.C) ||
			math.Float32bits(a[i].Q.S) != math.Float32bits(b[i].Q.S) {
			return false
		}
	}

	return true
}

func (trace *inputTransformTrace) read(r *oracle.Reader, worldID *box2d.World, key string, fullState bool) {
	transforms := box2d.InputBodyTransforms(worldID)

	hash := oracle.HashInit
	for _, transform := range transforms {
		hash = oracle.HashTransform(hash, transform.P.X, transform.P.Y, transform.Q.C, transform.Q.S)
	}

	r.U64(key, hash)
	r.U64(key+".canonical", worldID.StateHash())
	counters := worldID.Counters()
	r.I32(key+".bodies", int32(counters.BodyCount))
	r.I32(key+".shapes", int32(counters.ShapeCount))
	r.I32(key+".contacts", int32(counters.ContactCount))
	r.I32(key+".joints", int32(counters.JointCount))

	bodies := box2d.InputBodies(worldID)

	inputReadBodyMoves(r, worldID, key, fullState, bodies)

	for _, body := range bodies {
		if fullState {
			v := body.LinearVelocity()
			r.Vec2(key+".body.velocity", v.X, v.Y)
			r.F32(key+".body.angular", body.AngularVelocity())
			r.Bool(key+".body.awake", body.Awake())
			m := body.MassData()
			r.F32(key+".body.mass", m.Mass)
			r.Vec2(key+".body.center", m.Center.X, m.Center.Y)
			r.F32(key+".body.inertia", m.RotationalInertia)
		}
	}

	inputReadEventCounts(r, worldID, key)

	if !fullState {
		return
	}

	index := 0
	for index < len(trace.snapshots) && !inputSameTransforms(transforms, trace.snapshots[index]) {
		index++
	}

	r.I32("state", int32(index))

	if index < len(trace.snapshots) {
		return
	}

	trace.snapshots = append(trace.snapshots, transforms)
	r.I32(fmt.Sprintf("state.%d.count", index), int32(len(transforms)))

	for i, transform := range transforms {
		r.Pos(fmt.Sprintf("state.%d.b%d.p", index, i), transform.P.X, transform.P.Y)
		r.Rot(fmt.Sprintf("state.%d.b%d.q", index, i), transform.Q.C, transform.Q.S)
	}
}

func inputCheckMultithreadingTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "input/determinism/MultithreadingTest")
	trace := inputTransformTrace{}

	for range 3 {
		for workerCount := 1; workerCount < 16; workerCount += 2 {
			result := inputSingleMultithreading(t, r, &trace, workerCount)
			r.Bool("check.65", result == 0)

			if result != 0 {
				t.Fatal("condition false")
			}
		}

		for workerCount := 32; workerCount >= 0; workerCount -= 5 {
			result := inputSingleMultithreading(t, r, &trace, workerCount)
			r.Bool("check.71", result == 0)

			if result != 0 {
				t.Fatal("condition false")
			}
		}
	}
}

func inputCheckBuiltInSchedulerTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "input/determinism/BuiltInSchedulerTest")
	trace := inputTransformTrace{}

	for workerCount := 2; workerCount <= 8; workerCount += 2 {
		worldDef := box2d.DefaultWorldDef()
		worldDef.WorkerCount = workerCount
		worldID := box2d.NewWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)
		data := scenes.CreateFallingHinges(worldID)
		timeStep := float32(1.0 / 60.0)

		stepLimit := 1000
		for range stepLimit {
			subStepCount := 4
			worldID.Step(timeStep, subStepCount)
			done := scenes.UpdateFallingHinges(worldID, &data)
			trace.read(r, worldID, fmt.Sprintf("w%d.s%d", workerCount, data.StepCount), data.StepCount%60 == 0)

			if done {
				break
			}
		}

		worldID.Close()
		r.I32("check.111", int32(data.SleepStep))

		if data.SleepStep != scenes.ExpectedSleepStep {
			t.Fatal("condition false")
		}

		r.U32("check.112", data.Hash)

		if data.Hash != scenes.ExpectedHash {
			t.Fatal("condition false")
		}

		scenes.DestroyFallingHinges(&data)
	}
}

func inputCheckCrossPlatformTest(t *testing.T) {
	t.Helper()
	r := oracle.Open(t, "input/determinism/CrossPlatformTest")
	trace := inputTransformTrace{}
	worldDef := box2d.DefaultWorldDef()
	worldID := box2d.NewWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldID)
	data := scenes.CreateFallingHinges(worldID)
	timeStep := float32(1.0 / 60.0)

	done := false
	for done == false {
		subStepCount := 4
		worldID.Step(timeStep, subStepCount)
		done = scenes.UpdateFallingHinges(worldID, &data)
		trace.read(r, worldID, fmt.Sprintf("w%d.s%d", worldDef.WorkerCount, data.StepCount), data.StepCount%60 == 0)
	}

	r.I32("check.145", int32(data.SleepStep))

	if data.SleepStep != scenes.ExpectedSleepStep {
		t.Fatal("condition false")
	}

	r.U32("check.146", data.Hash)

	if data.Hash != scenes.ExpectedHash {
		t.Fatal("condition false")
	}

	scenes.DestroyFallingHinges(&data)
	worldID.Close()
}

func inputReadEventCounts(r *oracle.Reader, worldID *box2d.World, key string) {
	contactBegin, contactEnd, contactHit, sensorBegin, sensorEnd, jointEvents := 0, 0, 0, 0, 0, 0
	for range worldID.ContactBeginEvents() {
		contactBegin++
	}

	for range worldID.ContactEndEvents() {
		contactEnd++
	}

	for range worldID.ContactHitEvents() {
		contactHit++
	}

	for range worldID.SensorBeginEvents() {
		sensorBegin++
	}

	for range worldID.SensorEndEvents() {
		sensorEnd++
	}

	for range worldID.JointEvents() {
		jointEvents++
	}

	r.I32(key+".contact.begin", int32(contactBegin))
	r.I32(key+".contact.end", int32(contactEnd))
	r.I32(key+".contact.hit", int32(contactHit))
	r.I32(key+".sensor.begin", int32(sensorBegin))
	r.I32(key+".sensor.end", int32(sensorEnd))
	r.I32(key+".joint.events", int32(jointEvents))
}

func inputReadBodyMoves(r *oracle.Reader, worldID *box2d.World, key string, fullState bool, bodies []box2d.Body) {
	moveCount := 0
	for event := range worldID.BodyMoveEvents() {
		moveCount++

		if !fullState {
			continue
		}

		ordinal := -1

		for i, body := range bodies {
			if event.Body == body {
				ordinal = i
				break
			}
		}

		r.I32(key+".move.body", int32(ordinal))
		r.Bool(key+".move.sleep", event.FellAsleep)
		r.Pos(key+".move.p", event.Transform.P.X, event.Transform.P.Y)
		r.Rot(key+".move.q", event.Transform.Q.C, event.Transform.Q.S)
	}

	r.I32(key+".move.count", int32(moveCount))
}
