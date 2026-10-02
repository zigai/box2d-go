// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"strconv"
	"testing"
)

func addSolverArenaBody(w World, x Pos) (Body, Body) {
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
	body.EnableSleep(false)
	return body, ground
}

func checkSolverArenaCapacity(t *testing.T, state *world) {
	t.Helper()
	wideCount := 0
	for i := 0; i < overflowIndex; i++ {
		wideCount += (len(state.constraintGraph.colors[i].contactSims) + simdWidth - 1) / simdWidth
	}
	overflowCount := len(state.constraintGraph.colors[overflowIndex].contactSims)
	if cap(state.solverWideScratch) != wideCount || cap(state.solverOverflowScratch) != overflowCount {
		t.Fatalf("scratch capacity differs from maximum need: wide=%d want=%d overflow=%d want=%d",
			cap(state.solverWideScratch), wideCount, cap(state.solverOverflowScratch), overflowCount)
	}
}

func stepSolverArenaZero(t *testing.T, w World) {
	t.Helper()
	state := getWorldFromId(w)
	enqueue := state.enqueueTaskFcn
	checked := false
	state.enqueueTaskFcn = func(task func()) func() {
		if !checked {
			for i := range state.constraintGraph.colors {
				if len(state.constraintGraph.colors[i].wideConstraints) != 0 {
					checked = true
					break
				}
			}
			if checked {
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

func poisonSolverArena(state *world) {
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
				w, body, observed := newSolverScratchWorld(t, workers)
				w.SetGravity(Vec2{})
				body.EnableSleep(false)
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
			})
			t.Run("Zero", func(t *testing.T) {
				w, body, _ := newSolverScratchWorld(t, workers)
				w.SetGravity(Vec2{})
				body.EnableSleep(false)
				for i := 1; i < 5; i++ {
					addSolverArenaBody(w, Pos{X: posScalar(10 * i)})
				}
				w.Step(1.0/60.0, 4)
				state := getWorldFromId(w)
				checkSolverArenaCapacity(t, state)
				fullSlot := false
				for i := 0; i < overflowIndex; i++ {
					fullSlot = fullSlot || len(state.constraintGraph.colors[i].contactSims) >= simdWidth
				}
				if !fullSlot {
					t.Fatal("scene did not exercise a full wide slot in addition to tail lanes")
				}
				image := w.Snapshot()
				clone, err := CreateWorldFromSnapshot(image, workers)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(clone.Destroy)
				poisonSolverArena(state)
				if !bytes.Equal(image, w.Snapshot()) {
					t.Fatal("transient scratch affected snapshot bytes")
				}
				stepSolverArenaZero(t, w)
				clone.Step(1.0/60.0, 4)
				if w.StateHash() != clone.StateHash() {
					t.Fatal("poisoned scratch changed simulation state after reset")
				}
			})
			t.Run("GrowthShrink", func(t *testing.T) {
				w, body, _ := newSolverScratchWorld(t, workers)
				w.SetGravity(Vec2{})
				body.EnableSleep(false)
				w.Step(1.0/60.0, 4)
				state := getWorldFromId(w)
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
			})
			t.Run("Sleep", func(t *testing.T) {
				w, body, _ := newSolverScratchWorld(t, workers)
				w.SetGravity(Vec2{})
				w.Step(1.0/60.0, 4)
				state := getWorldFromId(w)
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
			})
			t.Run("RestoreDestroy", func(t *testing.T) {
				var state *world
				t.Cleanup(func() {
					if state != nil && (state.solverWideScratch != nil || state.solverOverflowScratch != nil) {
						t.Error("destroyed world retains owned scratch")
					}
				})
				w, body, _ := newSolverScratchWorld(t, workers)
				w.SetGravity(Vec2{})
				body.EnableSleep(false)
				w.Step(1.0/60.0, 4)
				state = getWorldFromId(w)
				wide, overflow := state.solverWideScratch, state.solverOverflowScratch
				image := w.Snapshot()
				clone, err := CreateWorldFromSnapshot(image, workers)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(clone.Destroy)
				other := getWorldFromId(clone)
				if other.solverWideScratch != nil || other.solverOverflowScratch != nil {
					t.Fatal("snapshot created scratch ownership")
				}
				poisonSolverArena(state)
				if err := w.Restore(image); err != nil {
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
			})
		})
	}
}
