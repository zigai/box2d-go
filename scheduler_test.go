// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	t.Run("ParallelFinish", func(t *testing.T) {
		s := createScheduler(4)
		defer destroyScheduler(s)
		entered := make(chan struct{}, 3)
		release := make(chan struct{})
		released := false
		defer func() {
			if !released {
				close(release)
			}
		}()
		var count atomic.Int32
		var values [3]int
		var finish [3]func()
		for i := range finish {
			finish[i] = schedulerEnqueueTask(func() {
				entered <- struct{}{}
				<-release
				values[i] = i + 1
				count.Add(1)
			}, s)
		}
		for range finish {
			select {
			case <-entered:
			case <-time.After(10 * time.Second):
				t.Fatal("parallel workers did not enter callbacks")
			}
		}
		close(release)
		released = true
		for i, f := range finish {
			f()
			f()
			if values[i] != i+1 {
				t.Fatalf("value[%d] = %d, want %d", i, values[i], i+1)
			}
		}
		if count.Load() != 3 {
			t.Fatalf("count = %d, want 3", count.Load())
		}
	})
	t.Run("FinishResetDestroy", func(t *testing.T) {
		s := createScheduler(4)
		var count atomic.Int32
		var finish [MaxTasks]func()
		for range 8 {
			for i := range finish {
				finish[i] = schedulerEnqueueTask(func() { count.Add(1) }, s)
			}
			for _, f := range finish {
				f()
			}
			resetScheduler(s)
		}
		for i := range finish {
			finish[i] = schedulerEnqueueTask(func() { count.Add(1) }, s)
		}
		for _, f := range finish {
			f()
		}
		destroyScheduler(s)
		if count.Load() != 9*MaxTasks {
			t.Fatalf("count = %d, want %d", count.Load(), 9*MaxTasks)
		}
	})
	t.Run("CooperativeFinish", func(t *testing.T) {
		s := createScheduler(1)
		defer destroyScheduler(s)
		var values []int
		first := schedulerEnqueueTask(func() { values = append(values, 1) }, s)
		last := schedulerEnqueueTask(func() { values = append(values, 2) }, s)
		last()
		first()
		last()
		if !reflect.DeepEqual(values, []int{1, 2}) {
			t.Fatalf("values = %v, want [1 2]", values)
		}
	})
	t.Run("ResetDestroyUnfinished", func(t *testing.T) {
		s := createScheduler(1)
		count := 0
		schedulerEnqueueTask(func() { count++ }, s)
		resetScheduler(s)
		schedulerEnqueueTask(func() { count++ }, s)
		destroyScheduler(s)
		if count != 0 {
			t.Fatalf("count = %d, want 0", count)
		}
	})
}

func enqueueSchedulerContext(s *scheduler, count *atomic.Int32) func() {
	context := &stepContext{wideContactConstraints: make([]contactConstraintWide, 1)}
	return schedulerEnqueueTask(func() {
		count.Add(int32(len(context.wideContactConstraints)))
	}, s)
}

func requireSchedulerCallbacksReleased(t *testing.T, s *scheduler) {
	t.Helper()
	for i := range s.tasks {
		if s.tasks[i].callback != nil {
			t.Errorf("task %d retains a callback capturing a step context", i)
		}
	}
}

func TestSchedulerCallbackLifetime(t *testing.T) {
	t.Run("Completed", func(t *testing.T) {
		for _, test := range []struct {
			name    string
			workers int
		}{{"Serial", 1}, {"Parallel", 4}} {
			t.Run(test.name, func(t *testing.T) {
				s := createScheduler(test.workers)
				defer destroyScheduler(s)
				var count atomic.Int32
				var finish func()
				if test.workers == 1 {
					finish = enqueueSchedulerContext(s, &count)
				} else {
					context := &stepContext{wideContactConstraints: make([]contactConstraintWide, 1)}
					entered := make(chan struct{})
					release := make(chan struct{})
					released := false
					defer func() {
						if !released {
							close(release)
						}
					}()
					finish = schedulerEnqueueTask(func() {
						close(entered)
						<-release
						count.Add(int32(len(context.wideContactConstraints)))
					}, s)
					select {
					case <-entered:
					case <-time.After(10 * time.Second):
						t.Fatal("worker did not enter the captured callback")
					}
					if s.tasks[0].callback == nil {
						t.Fatal("in-flight callback was released before completion")
					}
					close(release)
					released = true
				}
				finish()
				finish()
				if count.Load() != 1 {
					t.Fatal("callback did not complete exactly once")
				}
				requireSchedulerCallbacksReleased(t, s)
			})
		}
	})
	t.Run("ShrinkingBatch", func(t *testing.T) {
		s := createScheduler(4)
		defer destroyScheduler(s)
		var count atomic.Int32
		var finish [8]func()
		for i := range finish {
			finish[i] = enqueueSchedulerContext(s, &count)
		}
		for _, f := range finish {
			f()
		}
		resetScheduler(s)
		small := enqueueSchedulerContext(s, &count)
		small()
		small()
		if count.Load() != 9 {
			t.Fatal("shrinking batch did not execute exactly once")
		}
		requireSchedulerCallbacksReleased(t, s)
		resetScheduler(s)
		requireSchedulerCallbacksReleased(t, s)
	})
	t.Run("ResetUnclaimed", func(t *testing.T) {
		s := createScheduler(1)
		defer destroyScheduler(s)
		var count atomic.Int32
		for range 3 {
			enqueueSchedulerContext(s, &count)
		}
		resetScheduler(s)
		if count.Load() != 0 {
			t.Fatal("reset executed or kept the unclaimed batch")
		}
		requireSchedulerCallbacksReleased(t, s)
		finish := enqueueSchedulerContext(s, &count)
		finish()
		finish()
		if count.Load() != 1 {
			t.Fatal("reset discarded callbacks executed")
		}
		requireSchedulerCallbacksReleased(t, s)
	})
	t.Run("DestroyUnclaimed", func(t *testing.T) {
		s := createScheduler(1)
		var count atomic.Int32
		for range 3 {
			enqueueSchedulerContext(s, &count)
		}
		destroyScheduler(s)
		if count.Load() != 0 {
			t.Fatal("destroy changed unclaimed task execution or lifecycle")
		}
		requireSchedulerCallbacksReleased(t, s)
	})
}

// Observe actual worker exits, not scheduler addresses, callback identities, or
// the number of internal contexts used to implement a public worker setting.
func TestSchedulerWorldOwnedWorkersShutdown(t *testing.T) {
	for _, initial := range []int{1, 4} {
		for range 8 {
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
				for _, worker := range s.threads {
					if worker != nil {
						started[worker.done] = true
					}
				}
			}
			observeWorkers()
			for _, count := range []int{4, 1, 8, 2} {
				w.SetWorkerCount(count)
				observeWorkers()
				w.Step(1.0/60, 4)
				if s := w.scheduler; s != nil {
					requireSchedulerCallbacksReleased(t, s)
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
	}
}
