// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "sync/atomic"

type schedulerTaskStatus int

const (
	schedulerFree schedulerTaskStatus = iota
	schedulerPending
	schedulerClaimed
	schedulerComplete
)

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
	todo()
	return false
}

func schedulerWorkerMain(context *schedulerWorkerContext) {
	todo()
}

func createScheduler(workerCount int) *scheduler {
	todo()
	return nil
}

func destroyScheduler(s *scheduler) {
	todo()
}

func resetScheduler(s *scheduler) {
	todo()
}

func schedulerEnqueueTask(task func(), s *scheduler) func() {
	todo()
	return nil
}

func schedulerFinishTask(task *schedulerTask, s *scheduler) {
	todo()
}
