// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"strconv"
	"testing"
)

func addSolverArenaBody(w *World, x Pos) (Body, Body) {
	groundDef := DefaultBodyDef()
	groundDef.Position = Pos{X: x.X, Y: -0.5}
	ground := w.CreateBody(&groundDef)
	groundBox := MakeBox(2, 0.5)

	shapeDef := DefaultShapeDef()
	for range GraphColorCount + 1 {
		ground.CreatePolygonShape(&shapeDef, &groundBox)
	}

	bodyDef := DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: x.X, Y: 0.5}
	body := w.CreateBody(&bodyDef)
	box := MakeBox(0.5, 0.5)
	body.CreatePolygonShape(&shapeDef, &box)
	body.SetSleepEnabled(false)

	return body, ground
}

func checkSolverArenaCapacity(t *testing.T, state *World) {
	t.Helper()

	wideCount := 0
	for i := range overflowIndex {
		wideCount += (len(state.constraintGraph.colors[i].contactSims) + simdWidth - 1) / simdWidth
	}

	overflowCount := len(state.constraintGraph.colors[overflowIndex].contactSims)
	if cap(state.solverWideScratch) < wideCount || cap(state.solverOverflowScratch) < overflowCount {
		t.Fatalf("scratch capacity is below active need: wide=%d want>=%d overflow=%d want>=%d",
			cap(state.solverWideScratch), wideCount, cap(state.solverOverflowScratch), overflowCount)
	}
}

func stepSolverArenaZero(t *testing.T, w *World) {
	t.Helper()

	state := w
	enqueue := state.enqueueTaskFcn
	checked := false

	state.enqueueTaskFcn = func(task func()) func() {
		if !checked {
			checked = observeSolverArenaZero(t, state)
		}

		return enqueue(task)
	}
	defer func() { state.enqueueTaskFcn = enqueue }()

	w.Step(1.0/60.0, 4)

	if !checked {
		t.Fatal("did not observe active scratch before solver task launch")
	}

	checkSolverScratchReleased(t, w)
}

func poisonSolverArena(state *World) {
	for i := range state.solverWideScratch {
		state.solverWideScratch[i] = contactConstraintWide{
			indexA:              [simdWidth]int{91, 92, 93, 94},
			normalImpulse1:      floatW{11, 12, 13, 14},
			normalImpulse2:      floatW{21, 22, 23, 24},
			totalNormalImpulse1: floatW{31, 32, 33, 34},
			tangentImpulse2:     floatW{41, 42, 43, 44},
			rollingImpulse:      floatW{51, 52, 53, 54},
		}
	}

	for i := range state.solverOverflowScratch {
		state.solverOverflowScratch[i] = contactConstraint{
			indexA: 91, pointCount: 2, rollingImpulse: 51,
			points: [2]contactConstraintPoint{{normalImpulse: 11, totalNormalImpulse: 31}, {normalImpulse: 21, tangentImpulse: 41}},
		}
	}
}

func TestSolverArena(t *testing.T) {
	for _, workers := range []int{1, 4} {
		t.Run(strconv.Itoa(workers), func(t *testing.T) {
			t.Run("Reuse", func(t *testing.T) {
				testSolverArenaReuse(t, workers)
			})
			t.Run("Zero", func(t *testing.T) {
				testSolverArenaZero(t, workers)
			})
			t.Run("GrowthShrink", func(t *testing.T) {
				testSolverArenaGrowthShrink(t, workers)
			})
			t.Run("Sleep", func(t *testing.T) {
				testSolverArenaSleep(t, workers)
			})
			t.Run("RestoreDestroy", func(t *testing.T) {
				testSolverArenaRestoreDestroy(t, workers)
			})
		})
	}
}

func testSolverArenaReuse(t *testing.T, workers int) {
	w, body, observed := newSolverScratchWorld(t, workers)
	w.SetGravity(Vec2{})
	body.SetSleepEnabled(false)
	w.Step(1.0/60.0, 4)

	if len(observed.wide) == 0 || len(observed.overflow) == 0 {
		t.Fatal("contact world did not expose both constraint buffers")
	}

	wide, overflow := observed.wide, observed.overflow
	for step := range 8 {
		*observed = solverScratchObservation{}

		w.Step(1.0/60.0, 4)

		if len(observed.wide) != len(wide) || len(observed.overflow) != len(overflow) {
			t.Fatalf("step %d changed stable constraint counts", step)
		}

		if &observed.wide[0] != &wide[0] || &observed.overflow[0] != &overflow[0] {
			t.Fatalf("step %d did not reuse stable wide and overflow scratch", step)
		}

		checkSolverScratchReleased(t, w)
	}
}

func testSolverArenaZero(t *testing.T, workers int) {
	w, body, _ := newSolverScratchWorld(t, workers)
	w.SetGravity(Vec2{})
	body.SetSleepEnabled(false)

	for i := 1; i < 5; i++ {
		addSolverArenaBody(w, Pos{X: posScalar(10 * i)})
	}

	w.Step(1.0/60.0, 4)
	state := w
	checkSolverArenaCapacity(t, state)

	fullSlot := false
	for i := range overflowIndex {
		fullSlot = fullSlot || len(state.constraintGraph.colors[i].contactSims) >= simdWidth
	}

	if !fullSlot {
		t.Fatal("scene did not exercise a full wide slot in addition to tail lanes")
	}

	image := MarshalSnapshot(w)

	clone, err := NewWorldFromSnapshot(image, workers)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(clone.Close)
	poisonSolverArena(state)

	if !bytes.Equal(image, MarshalSnapshot(w)) {
		t.Fatal("transient scratch affected snapshot bytes")
	}

	stepSolverArenaZero(t, w)
	clone.Step(1.0/60.0, 4)

	if w.StateHash() != clone.StateHash() {
		t.Fatal("poisoned scratch changed simulation state after reset")
	}
}

func testSolverArenaGrowthShrink(t *testing.T, workers int) {
	w, body, _ := newSolverScratchWorld(t, workers)
	w.SetGravity(Vec2{})
	body.SetSleepEnabled(false)
	w.Step(1.0/60.0, 4)
	state := w
	checkSolverArenaCapacity(t, state)
	wideCap, overflowCap := cap(state.solverWideScratch), cap(state.solverOverflowScratch)

	var added []Body

	for i := 1; i < 9; i++ {
		body, ground := addSolverArenaBody(w, Pos{X: posScalar(10 * i)})
		added = append(added, body, ground)
	}

	stepSolverArenaZero(t, w)
	checkSolverArenaCapacity(t, state)

	if cap(state.solverWideScratch) <= wideCap || cap(state.solverOverflowScratch) <= overflowCap {
		t.Fatal("added contacts did not grow both scratch buffers")
	}

	wide, overflow := state.solverWideScratch, state.solverOverflowScratch

	for _, b := range added {
		b.Destroy()
	}

	stepSolverArenaZero(t, w)

	if &state.solverWideScratch[0] != &wide[0] || &state.solverOverflowScratch[0] != &overflow[0] || cap(state.solverWideScratch) != cap(wide) || cap(state.solverOverflowScratch) != cap(overflow) {
		t.Fatal("shrinking contacts replaced or shrank high-water scratch")
	}

	poisonSolverArena(state)

	for i := 1; i < 9; i++ {
		addSolverArenaBody(w, Pos{X: posScalar(10 * i)})
	}

	stepSolverArenaZero(t, w)

	if &state.solverWideScratch[0] != &wide[0] || &state.solverOverflowScratch[0] != &overflow[0] {
		t.Fatal("regrowth within high-water capacity reallocated scratch")
	}
}

func testSolverArenaSleep(t *testing.T, workers int) {
	w, body, _ := newSolverScratchWorld(t, workers)
	w.SetGravity(Vec2{})
	w.Step(1.0/60.0, 4)
	state := w
	wide, overflow := state.solverWideScratch, state.solverOverflowScratch

	body.SetAwake(false)

	if w.AwakeBodyCount() != 0 {
		t.Fatal("world did not become inactive")
	}

	w.Step(1.0/60.0, 4)
	checkSolverScratchReleased(t, w)

	if cap(state.solverWideScratch) != cap(wide) || cap(state.solverOverflowScratch) != cap(overflow) {
		t.Fatal("inactive world changed scratch capacity")
	}

	poisonSolverArena(state)
	body.SetAwake(true)
	stepSolverArenaZero(t, w)

	if &state.solverWideScratch[0] != &wide[0] || &state.solverOverflowScratch[0] != &overflow[0] {
		t.Fatal("waking world replaced scratch")
	}
}

func testSolverArenaRestoreDestroy(t *testing.T, workers int) {
	var state *World

	t.Cleanup(func() {
		if state != nil && (state.solverWideScratch != nil || state.solverOverflowScratch != nil) {
			t.Error("destroyed world retains owned scratch")
		}
	})
	w, body, _ := newSolverScratchWorld(t, workers)
	w.SetGravity(Vec2{})
	body.SetSleepEnabled(false)
	w.Step(1.0/60.0, 4)
	state = w
	wide, overflow := state.solverWideScratch, state.solverOverflowScratch
	image := MarshalSnapshot(w)

	clone, err := NewWorldFromSnapshot(image, workers)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(clone.Close)

	other := clone
	if other.solverWideScratch != nil || other.solverOverflowScratch != nil {
		t.Fatal("snapshot created scratch ownership")
	}

	poisonSolverArena(state)

	if err := w.UnmarshalBinary(image); err != nil {
		t.Fatal(err)
	}

	checkSolverScratchReleased(t, w)
	stepSolverArenaZero(t, w)
	clone.Step(1.0/60.0, 4)

	if w.StateHash() != clone.StateHash() {
		t.Fatal("restored cache changed persistent state")
	}

	if &state.solverWideScratch[0] != &wide[0] || &state.solverOverflowScratch[0] != &overflow[0] {
		t.Fatal("Restore replaced world scratch ownership")
	}

	if &other.solverWideScratch[0] == &wide[0] || &other.solverOverflowScratch[0] == &overflow[0] {
		t.Fatal("snapshot worlds share scratch")
	}
}

func observeSolverArenaZero(t *testing.T, state *World) bool {
	t.Helper()

	active := false

	for i := range state.constraintGraph.colors {
		if len(state.constraintGraph.colors[i].wideConstraints) != 0 {
			active = true
			break
		}
	}

	if !active {
		return false
	}

	checkSolverArenaZeroConstraints(t, state)

	return true
}

func checkSolverArenaZeroConstraints(t *testing.T, state *World) {
	t.Helper()

	for i := range state.constraintGraph.colors {
		color := &state.constraintGraph.colors[i]
		for j, constraint := range color.wideConstraints {
			if constraint != (contactConstraintWide{}) {
				t.Errorf("wide constraint %d/%d is not fresh-make zero before solver tasks", i, j)
			}
		}

		for j, constraint := range color.overflowConstraints {
			if constraint != (contactConstraint{}) {
				t.Errorf("overflow constraint %d is not fresh-make zero before solver tasks", j)
			}
		}
	}
}

func TestStepStorageRestoreAndIsolation(t *testing.T) {
	for _, workers := range []int{1, 4} {
		t.Run(strconv.Itoa(workers), func(t *testing.T) {
			def := DefaultWorldDef()
			def.WorkerCount = workers
			def.Gravity = Vec2{}
			w := NewWorld(&def)
			t.Cleanup(w.Close)
			addSolverArenaBody(w, Pos{})

			for range 8 {
				w.Step(1.0/60.0, 4)
			}

			image, err := w.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			clone, err := NewWorldFromSnapshot(image, workers)
			if err != nil {
				t.Fatal(err)
			}

			t.Cleanup(clone.Close)

			for range 8 {
				before := clone.StateHash()

				w.Step(1.0/60.0, 4)

				if clone.StateHash() != before {
					t.Fatal("stepping one world changed another world's state")
				}

				clone.Step(1.0/60.0, 4)

				if w.StateHash() != clone.StateHash() {
					t.Fatal("snapshot worlds diverged while reusing step storage")
				}
			}

			if err := w.UnmarshalBinary(image); err != nil {
				t.Fatal(err)
			}

			if err := clone.UnmarshalBinary(image); err != nil {
				t.Fatal(err)
			}

			w.Step(0, 4)
			clone.Step(0, 4)

			for range 8 {
				before := clone.StateHash()

				w.Step(1.0/60.0, 4)

				if clone.StateHash() != before {
					t.Fatal("stepping a restored world changed another world's state")
				}

				clone.Step(1.0/60.0, 4)

				if w.StateHash() != clone.StateHash() {
					t.Fatal("restored worlds diverged while reusing step storage")
				}
			}
		})
	}
}

func BenchmarkNativeWarmStep(b *testing.B) {
	for _, workers := range []int{1, 4} {
		b.Run(strconv.Itoa(workers), func(b *testing.B) {
			def := DefaultWorldDef()
			def.WorkerCount = workers
			def.Gravity = Vec2{}
			w := NewWorld(&def)
			b.Cleanup(w.Close)

			for i := range 8 {
				body, _ := addSolverArenaBody(w, Pos{X: posScalar(5 * i)})
				body.SetLinearVelocity(Vec2{X: 0.01})
			}

			for range 120 {
				w.Step(1.0/60.0, 4)
			}

			b.ReportAllocs()

			for b.Loop() {
				w.Step(1.0/60.0, 4)
			}
		})
	}
}

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
