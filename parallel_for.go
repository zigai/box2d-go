// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "sync/atomic"

type parallelForFunc func(startIndex, endIndex, workerIndex int)

type parallelForShared struct {
	nextBlock  atomic.Int32
	blockCount int
	blockSize  int
	itemCount  int
	callback   parallelForFunc
}

type parallelForTask struct {
	shared      *parallelForShared
	workerIndex int
}

func parallelForTrampoline(task *parallelForTask) {
	todo()
}

func parallelFor(w *world, callback parallelForFunc, itemCount, minRange int) {
	todo()
}
