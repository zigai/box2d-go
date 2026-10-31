// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "testing"

// Observe actual worker exits, not scheduler addresses, callback identities, or
// the number of internal contexts used to implement a public worker setting.
func TestSchedulerWorldOwnedWorkersShutdown(t *testing.T) {
	for _, initial := range []int{1, 4} {
		for range 8 {
			testSchedulerWorldWorkerLifecycle(t, initial)
		}
	}
}

func testSchedulerWorldWorkerLifecycle(t *testing.T, initial int) {
	t.Helper()

	def := DefaultWorldDef()
	def.WorkerCount = initial
	w := NewWorld(&def)

	t.Cleanup(func() {
		if w.IsValid() {
			w.Close()
		}
	})

	bodyDef := DefaultBodyDef()
	bodyDef.Type = DynamicBody
	body := w.CreateBody(&bodyDef)
	shapeDef := DefaultShapeDef()
	circle := Circle{Radius: 1}
	body.CreateCircleShape(&shapeDef, &circle)

	started := map[<-chan struct{}]bool{}
	observeWorkers := func() {
		s := w.scheduler
		if s == nil {
			return
		}

		for _, done := range s.WorkerDone() {
			started[done] = true
		}
	}
	observeWorkers()

	for _, count := range []int{4, 1, 8, 2} {
		w.SetWorkerCount(count)
		observeWorkers()
		w.Step(1.0/60, 4)

		if s := w.scheduler; s != nil && s.RetainsCallbacks() {
			t.Error("a task retains a callback capturing a step context")
		}
	}

	w.Close()

	for done := range started {
		select {
		case <-done:
		default:
			t.Fatal("World.Destroy returned before an owned worker exited")
		}
	}
}
