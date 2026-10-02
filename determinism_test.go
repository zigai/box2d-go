// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
	"github.com/zigai/box2d/internal/scenes"
)

func inputSingleMultithreading(t *testing.T, r *oracle.Reader, trace *inputTransformTrace, workerCount int) int {
	t.Helper()
	worldDef := box2d.DefaultWorldDef()
	worldDef.WorkerCount = workerCount
	worldID := box2d.CreateWorld(&worldDef)
	data := scenes.CreateFallingHinges(worldID)
	timeStep := float32(1.0 / 60.0)
	stepLimit := 500
	for i := 0; i < stepLimit; i++ {
		subStepCount := 4
		worldID.Step(timeStep, subStepCount)
		done := scenes.UpdateFallingHinges(worldID, &data)
		trace.read(r, worldID, fmt.Sprintf("w%d.s%d", workerCount, data.StepCount), data.StepCount%60 == 0)
		if done {
			break
		}
	}
	worldID.Destroy()
	r.I32("check.49", int32(data.SleepStep))
	porttest.Ensure(t, data.SleepStep == scenes.ExpectedSleepStep)
	r.U32("check.50", data.Hash)
	porttest.Ensure(t, data.Hash == scenes.ExpectedHash)
	scenes.DestroyFallingHinges(&data)
	return 0
}

func TestDeterminism(t *testing.T) {
	defer porttest.Guard(t)
	// Test multithreaded determinism.
	t.Run("MultithreadingTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/determinism/MultithreadingTest")
		trace := inputTransformTrace{}
		for run := 0; run < 3; run++ {
			for workerCount := 1; workerCount < 16; workerCount += 2 {
				result := inputSingleMultithreading(t, r, &trace, workerCount)
				r.Bool("check.65", result == 0)
				porttest.Ensure(t, result == 0)
			}
			for workerCount := 32; workerCount >= 0; workerCount -= 5 {
				result := inputSingleMultithreading(t, r, &trace, workerCount)
				r.Bool("check.71", result == 0)
				porttest.Ensure(t, result == 0)
			}
		}
	})
	// Test determinism using the built-in scheduler (no external task system).
	t.Run("BuiltInSchedulerTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/determinism/BuiltInSchedulerTest")
		trace := inputTransformTrace{}
		for workerCount := 2; workerCount <= 8; workerCount += 2 {
			worldDef := box2d.DefaultWorldDef()
			worldDef.WorkerCount = workerCount
			worldID := box2d.CreateWorld(&worldDef)
			data := scenes.CreateFallingHinges(worldID)
			timeStep := float32(1.0 / 60.0)
			stepLimit := 1000
			for i := 0; i < stepLimit; i++ {
				subStepCount := 4
				worldID.Step(timeStep, subStepCount)
				done := scenes.UpdateFallingHinges(worldID, &data)
				trace.read(r, worldID, fmt.Sprintf("w%d.s%d", workerCount, data.StepCount), data.StepCount%60 == 0)
				if done {
					break
				}
			}
			worldID.Destroy()
			r.I32("check.111", int32(data.SleepStep))
			porttest.Ensure(t, data.SleepStep == scenes.ExpectedSleepStep)
			r.U32("check.112", data.Hash)
			porttest.Ensure(t, data.Hash == scenes.ExpectedHash)
			scenes.DestroyFallingHinges(&data)
		}
	})
	// Test cross-platform determinism.
	t.Run("CrossPlatformTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "input/determinism/CrossPlatformTest")
		trace := inputTransformTrace{}
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
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
		porttest.Ensure(t, data.SleepStep == scenes.ExpectedSleepStep)
		r.U32("check.146", data.Hash)
		porttest.Ensure(t, data.Hash == scenes.ExpectedHash)
		scenes.DestroyFallingHinges(&data)
		worldID.Destroy()
	})
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

func (trace *inputTransformTrace) read(r *oracle.Reader, worldID box2d.World, key string, fullState bool) {
	transforms := box2d.InputBodyTransforms(worldID)
	hash := oracle.HashInit
	for _, transform := range transforms {
		hash = oracle.HashTransform(hash, transform.P.X, transform.P.Y, transform.Q.C, transform.Q.S)
	}
	r.U64(key, hash)
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
