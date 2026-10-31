// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

//go:build box2d_stress

package box2d_test

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"

	"github.com/zigai/box2d-go/internal/scenes"
)

func recordingRestepRaceStress(t *testing.T, r *oracle.Reader, trace *recordingTrace, build func(*box2d.World), name string, workerCount, churnFrame, restepCount, reps int) int {
	t.Helper()

	const dt float32 = 1.0 / 60.0

	subSteps := 4

	wd := box2d.DefaultWorldDef()
	wd.WorkerCount = workerCount
	liveID := box2d.NewWorld(&wd)
	box2d.CleanupWorld(t, liveID)
	build(liveID)

	for i := range churnFrame {
		liveID.Step(dt, subSteps)
		trace.Check(t, r, liveID, "step 1074", i+1)
	}

	image := box2d.MarshalSnapshot(liveID)
	liveID.Close()

	cd := box2d.DefaultWorldDef()
	cd.WorkerCount = workerCount
	cloneID := box2d.NewWorld(&cd)
	box2d.CleanupWorld(t, cloneID)
	clone := cloneID

	var ref uint64

	result := 0
	for rep := range reps {
		if result != 0 {
			break
		}

		check1091 := cloneID.UnmarshalBinary(image) == nil
		if !check1091 {
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

	cloneID.Close()

	return result
}

func TestReStepRace(t *testing.T) {
	r := oracle.Open(t, "recording/ReStepRaceTest")
	trace := newRecordingTrace(t, "ReStepRaceTest")

	check1123 := recordingRestepRaceStress(t, r, trace, scenes.CreateJunkyard, "junkyard", 2, 150, 50, 16) == 0
	if !check1123 {
		t.Fatal("condition false")
	}

	r.Bool("check 1123", check1123)
}
