// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"strconv"
	"testing"
)

type solverScratchObservation struct {
	wide      []contactConstraintWide
	wideColor int
	overflow  []contactConstraint
}

func newSolverScratchWorld(t *testing.T, workers int) (*World, Body, *solverScratchObservation) {
	t.Helper()

	var state *World

	observed := new(solverScratchObservation)
	observe := func() {
		for i := range state.constraintGraph.colors {
			color := &state.constraintGraph.colors[i]
			if len(color.wideConstraints) > 0 {
				observed.wide = color.wideConstraints
				observed.wideColor = i
			}

			if len(color.overflowConstraints) > 0 {
				observed.overflow = color.overflowConstraints
			}
		}
	}
	def := DefaultWorldDef()
	def.WorkerCount = workers
	def.EnqueueTask = func(task func()) func() {
		if workers == 1 {
			task()
			observe()

			return nil
		}

		done := make(chan struct{})
		go func() {
			defer close(done)

			task()
		}()

		return func() {
			<-done
			observe()
		}
	}
	w := NewWorld(&def)
	t.Cleanup(w.Close)
	state = w
	groundDef := DefaultBodyDef()
	groundDef.Position = Pos{Y: -0.5}
	ground := w.CreateBody(&groundDef)
	groundBox := MakeBox(2, 0.5)

	shapeDef := DefaultShapeDef()
	for range GraphColorCount + 1 {
		ground.CreatePolygonShape(&shapeDef, &groundBox)
	}

	bodyDef := DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{Y: 0.5}
	body := w.CreateBody(&bodyDef)
	box := MakeBox(0.5, 0.5)
	body.CreatePolygonShape(&shapeDef, &box)

	return w, body, observed
}

func checkSolverScratchReleased(t *testing.T, w *World) {
	t.Helper()

	state := w
	for i := range state.constraintGraph.colors {
		color := &state.constraintGraph.colors[i]
		if color.wideConstraints != nil || color.overflowConstraints != nil || color.wideConstraintCount != 0 {
			t.Errorf("color %d retains solver scratch: wide=%d/%d overflow=%d/%d count=%d", i,
				len(color.wideConstraints), cap(color.wideConstraints),
				len(color.overflowConstraints), cap(color.overflowConstraints), color.wideConstraintCount)
		}
	}
}

func TestSolverScratchLifetime(t *testing.T) {
	for _, workers := range []int{1, 4} {
		t.Run(strconv.Itoa(workers), func(t *testing.T) {
			t.Run("RepeatedContacts", func(t *testing.T) {
				testSolverScratchRepeatedContacts(t, workers)
			})
			t.Run("InactiveContacts", func(t *testing.T) {
				testSolverScratchInactiveContacts(t, workers)
			})
			t.Run("SleepingContacts", func(t *testing.T) {
				testSolverScratchSleepingContacts(t, workers)
			})
		})
	}
}

func testSolverScratchRepeatedContacts(t *testing.T, workers int) {
	w, body, observed := newSolverScratchWorld(t, workers)
	body.SetSleepEnabled(false)

	for step := range 8 {
		*observed = solverScratchObservation{}

		w.Step(1.0/60.0, 4)

		if len(observed.wide) == 0 || len(observed.overflow) == 0 {
			t.Fatalf("step %d did not expose both scratch buffers at task completion: wide=%d overflow=%d", step, len(observed.wide), len(observed.overflow))
		}

		checkSolverScratchReleased(t, w)
	}
}

func testSolverScratchInactiveContacts(t *testing.T, workers int) {
	w, body, observed := newSolverScratchWorld(t, workers)
	w.Step(1.0/60.0, 4)

	if len(observed.wide) == 0 || len(observed.overflow) == 0 {
		t.Fatalf("contact step did not use both scratch buffers: wide=%d overflow=%d", len(observed.wide), len(observed.overflow))
	}

	body.SetTransform(Pos{X: 100, Y: 0.5}, Rot{C: 1})
	w.Step(1.0/60.0, 4)

	if !body.Awake() {
		t.Fatal("separated body is not awake")
	}

	state := w
	for i := range state.constraintGraph.colors {
		if len(state.constraintGraph.colors[i].contactSims) != 0 {
			t.Fatalf("color %d still has active contacts after separation", i)
		}
	}

	checkSolverScratchReleased(t, w)
}

func testSolverScratchSleepingContacts(t *testing.T, workers int) {
	w, body, observed := newSolverScratchWorld(t, workers)
	w.Step(1.0/60.0, 4)

	if len(observed.wide) == 0 || len(observed.overflow) == 0 {
		t.Fatalf("contact step did not use both scratch buffers: wide=%d overflow=%d", len(observed.wide), len(observed.overflow))
	}

	body.SetAwake(false)

	if body.Awake() || w.AwakeBodyCount() != 0 {
		t.Fatal("world did not transition to no awake bodies")
	}

	state := w
	color := &state.constraintGraph.colors[observed.wideColor]
	color.wideConstraints = observed.wide
	color.wideConstraintCount = len(observed.wide)
	state.constraintGraph.colors[overflowIndex].overflowConstraints = observed.overflow

	w.Step(1.0/60.0, 4)
	checkSolverScratchReleased(t, w)
}
