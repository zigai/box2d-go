// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"sync/atomic"
)

const (
	schedulerFree schedulerTaskStatus = iota
	schedulerPending
	schedulerClaimed
	schedulerComplete
)

type schedulerTaskStatus int

type schedulerTask struct {
	callback func()
	status   atomic.Int32
}

type schedulerWorkerContext struct {
	scheduler   *scheduler
	threadIndex int
}

type scheduler struct {
	threads        [MaxWorkers]*thread
	workerContexts [MaxWorkers]schedulerWorkerContext
	workerCount    int
	threadCount    int
	tasks          [MaxTasks]schedulerTask
	nextSlot       atomic.Int32
	taskSemaphore  *semaphore
	shutdown       atomic.Int32
}

func schedulerExecuteOne(s *scheduler) bool {
	taskCount := int(s.nextSlot.Load())
	for i := range taskCount {
		task := &s.tasks[i]
		if !task.status.CompareAndSwap(int32(schedulerPending), int32(schedulerClaimed)) {
			continue
		}

		task.callback()
		task.callback = nil
		task.status.Store(int32(schedulerComplete))

		return true
	}

	return false
}

func schedulerWorkerMain(context *schedulerWorkerContext) {
	s := context.scheduler
	for {
		waitSemaphore(s.taskSemaphore)

		if s.shutdown.Load() != 0 {
			return
		}

		for schedulerExecuteOne(s) {
		}
	}
}

func createScheduler(workerCount int) *scheduler {
	assert(0 < workerCount && workerCount <= MaxWorkers, "0 < workerCount && workerCount <= B2_MAX_WORKERS")

	s := &scheduler{workerCount: workerCount, threadCount: workerCount - 1, taskSemaphore: createSemaphore(0)}
	for i := range s.threadCount {
		context := &s.workerContexts[i]
		context.scheduler = s
		context.threadIndex = i + 1
		s.threads[i] = createThread(func() { schedulerWorkerMain(context) }, fmt.Sprintf("box2d_worker_%02d", i+1))
	}

	return s
}

func destroyScheduler(s *scheduler) {
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

func resetScheduler(s *scheduler) {
	for i := range int(s.nextSlot.Load()) {
		s.tasks[i].callback = nil
	}

	s.nextSlot.Store(0)
}

func schedulerEnqueueTask(task func(), s *scheduler) func() {
	slot := int(s.nextSlot.Add(1) - 1)
	assert(slot < MaxTasks, "slot < B2_MAX_TASKS")
	schedulerTask := &s.tasks[slot]
	schedulerTask.callback = task
	schedulerTask.status.Store(int32(schedulerPending))
	signalSemaphore(s.taskSemaphore)

	return func() { schedulerFinishTask(schedulerTask, s) }
}

func schedulerFinishTask(task *schedulerTask, s *scheduler) {
	if task == nil {
		return
	}

	for task.status.Load() != int32(schedulerComplete) {
		if !schedulerExecuteOne(s) {
			yieldThread()
		}
	}
}
