// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scheduler

import (
	"fmt"
	"sync/atomic"

	"github.com/zigai/box2d-go/internal/core"
	"github.com/zigai/box2d-go/internal/timer"
)

const (
	taskFree taskStatus = iota
	taskPending
	taskClaimed
	taskComplete
)

type taskStatus int

type task struct {
	callback func()
	status   atomic.Int32
}

type workerContext struct {
	scheduler   *Scheduler
	threadIndex int
}

type Scheduler struct {
	threads        []*thread
	workerContexts []workerContext
	workerCount    int
	threadCount    int
	tasks          []task
	nextSlot       atomic.Int32
	taskSemaphore  *semaphore
	shutdown       atomic.Int32
}

// New starts workerCount-1 worker goroutines that run tasks alongside the caller.
// At most taskCapacity tasks can be enqueued between resets.
func New(workerCount int, taskCapacity int) *Scheduler {
	core.Assert(0 < workerCount, "0 < workerCount")

	s := &Scheduler{
		threads:        make([]*thread, workerCount-1),
		workerContexts: make([]workerContext, workerCount-1),
		workerCount:    workerCount,
		threadCount:    workerCount - 1,
		tasks:          make([]task, taskCapacity),
		taskSemaphore:  createSemaphore(0),
	}
	for i := range s.threadCount {
		context := &s.workerContexts[i]
		context.scheduler = s
		context.threadIndex = i + 1
		s.threads[i] = createThread(
			func() { workerMain(context) },
			fmt.Sprintf("box2d_worker_%02d", i+1),
		)
	}

	return s
}

func workerMain(context *workerContext) {
	s := context.scheduler
	for {
		waitSemaphore(s.taskSemaphore)

		if s.shutdown.Load() != 0 {
			return
		}

		for s.executeOne() {
		}
	}
}

func (s *Scheduler) Destroy() {
	s.shutdown.Store(1)

	for range s.threadCount {
		signalSemaphore(s.taskSemaphore)
	}

	for i := range s.threadCount {
		joinThread(s.threads[i])
		s.threads[i] = nil
	}

	for i := range int(s.nextSlot.Load()) {
		s.tasks[i].callback = nil
	}

	destroySemaphore(s.taskSemaphore)
}

func (s *Scheduler) Reset() {
	for i := range int(s.nextSlot.Load()) {
		s.tasks[i].callback = nil
	}

	s.nextSlot.Store(0)
}

// Enqueue schedules callback and returns a function that blocks until it has run,
// helping with pending tasks while it waits.
func (s *Scheduler) Enqueue(callback func()) func() {
	slot := int(s.nextSlot.Add(1) - 1)
	core.Assert(slot < len(s.tasks), "slot < B2_MAX_TASKS")
	task := &s.tasks[slot]
	task.callback = callback
	task.status.Store(int32(taskPending))
	signalSemaphore(s.taskSemaphore)

	return func() { s.finishTask(task) }
}

// WorkerDone returns, for each live worker goroutine, a channel closed when it exits.
func (s *Scheduler) WorkerDone() []<-chan struct{} {
	var done []<-chan struct{}

	for _, worker := range s.threads {
		if worker != nil {
			done = append(done, worker.done)
		}
	}

	return done
}

// RetainsCallbacks reports whether any task slot still references its callback.
func (s *Scheduler) RetainsCallbacks() bool {
	for i := range s.tasks {
		if s.tasks[i].callback != nil {
			return true
		}
	}

	return false
}

func (s *Scheduler) executeOne() bool {
	taskCount := int(s.nextSlot.Load())
	for i := range taskCount {
		task := &s.tasks[i]
		if !task.status.CompareAndSwap(int32(taskPending), int32(taskClaimed)) {
			continue
		}

		task.callback()
		task.callback = nil
		task.status.Store(int32(taskComplete))

		return true
	}

	return false
}

func (s *Scheduler) finishTask(task *task) {
	if task == nil {
		return
	}

	for task.status.Load() != int32(taskComplete) {
		if !s.executeOne() {
			timer.Yield()
		}
	}
}
