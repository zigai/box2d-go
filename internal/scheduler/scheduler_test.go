// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scheduler

import (
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

const taskCapacity = 256

type capturedContext struct {
	constraints []int
}

func TestSchedulerParallelFinish(t *testing.T) {
	s := New(4, taskCapacity)
	defer s.Destroy()

	entered := make(chan struct{}, 3)
	release := make(chan struct{})

	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()

	var (
		count  atomic.Int32
		values [3]int
		finish [3]func()
	)
	for i := range finish {
		finish[i] = s.Enqueue(func() {
			entered <- struct{}{}

			<-release

			values[i] = i + 1

			count.Add(1)
		})
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
}

func TestSchedulerFinishResetDestroy(t *testing.T) {
	s := New(4, taskCapacity)

	var (
		count  atomic.Int32
		finish [taskCapacity]func()
	)
	for range 8 {
		for i := range finish {
			finish[i] = s.Enqueue(func() { count.Add(1) })
		}

		for _, f := range finish {
			f()
		}

		s.Reset()
	}

	for i := range finish {
		finish[i] = s.Enqueue(func() { count.Add(1) })
	}

	for _, f := range finish {
		f()
	}

	s.Destroy()

	if count.Load() != 9*taskCapacity {
		t.Fatalf("count = %d, want %d", count.Load(), 9*taskCapacity)
	}
}

func TestSchedulerCooperativeFinish(t *testing.T) {
	s := New(1, taskCapacity)
	defer s.Destroy()

	var values []int

	first := s.Enqueue(func() { values = append(values, 1) })
	last := s.Enqueue(func() { values = append(values, 2) })
	last()
	first()
	last()

	if !reflect.DeepEqual(values, []int{1, 2}) {
		t.Fatalf("values = %v, want [1 2]", values)
	}
}

func TestSchedulerResetDestroyUnfinished(t *testing.T) {
	s := New(1, taskCapacity)
	count := 0

	s.Enqueue(func() { count++ })
	s.Reset()
	s.Enqueue(func() { count++ })
	s.Destroy()

	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

func enqueueSchedulerContext(s *Scheduler, count *atomic.Int32) func() {
	context := &capturedContext{constraints: make([]int, 1)}

	return s.Enqueue(func() {
		count.Add(int32(len(context.constraints)))
	})
}

func requireSchedulerCallbacksReleased(t *testing.T, s *Scheduler) {
	t.Helper()

	for i := range s.tasks {
		if s.tasks[i].callback != nil {
			t.Errorf("task %d retains a callback capturing a step context", i)
		}
	}
}

func TestSchedulerCallbackLifetimeCompleted(t *testing.T) {
	for _, test := range []struct {
		name    string
		workers int
	}{{"Serial", 1}, {"Parallel", 4}} {
		t.Run(test.name, func(t *testing.T) {
			s := New(test.workers, taskCapacity)
			defer s.Destroy()

			var (
				count  atomic.Int32
				finish func()
			)
			if test.workers == 1 {
				finish = enqueueSchedulerContext(s, &count)
			} else {
				context := &capturedContext{constraints: make([]int, 1)}
				entered := make(chan struct{})
				release := make(chan struct{})

				released := false
				defer func() {
					if !released {
						close(release)
					}
				}()

				finish = s.Enqueue(func() {
					close(entered)
					<-release
					count.Add(int32(len(context.constraints)))
				})

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
}

func TestSchedulerCallbackLifetimeShrinkingBatch(t *testing.T) {
	s := New(4, taskCapacity)
	defer s.Destroy()

	var (
		count  atomic.Int32
		finish [8]func()
	)
	for i := range finish {
		finish[i] = enqueueSchedulerContext(s, &count)
	}

	for _, f := range finish {
		f()
	}

	s.Reset()
	small := enqueueSchedulerContext(s, &count)
	small()
	small()

	if count.Load() != 9 {
		t.Fatal("shrinking batch did not execute exactly once")
	}

	requireSchedulerCallbacksReleased(t, s)
	s.Reset()
	requireSchedulerCallbacksReleased(t, s)
}

func TestSchedulerCallbackLifetimeResetUnclaimed(t *testing.T) {
	s := New(1, taskCapacity)
	defer s.Destroy()

	var count atomic.Int32
	for range 3 {
		enqueueSchedulerContext(s, &count)
	}

	s.Reset()

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
}

func TestSchedulerCallbackLifetimeDestroyUnclaimed(t *testing.T) {
	s := New(1, taskCapacity)

	var count atomic.Int32
	for range 3 {
		enqueueSchedulerContext(s, &count)
	}

	s.Destroy()

	if count.Load() != 0 {
		t.Fatal("destroy changed unclaimed task execution or lifecycle")
	}

	requireSchedulerCallbacksReleased(t, s)
}
