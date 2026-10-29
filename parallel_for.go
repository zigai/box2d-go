// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"sync/atomic"

	"github.com/zigai/box2d/internal/core"
)

type parallelForFunc func(startIndex, endIndex, workerIndex int)

// Shared state for one parallelFor invocation. Workers race on nextBlock to
// claim work, so a slow chunk can't strand the other threads.
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
	shared := task.shared
	workerIndex := task.workerIndex
	callback := shared.callback
	blockCount := shared.blockCount
	blockSize := shared.blockSize

	itemCount := shared.itemCount
	for {
		blockIndex := int(shared.nextBlock.Add(1) - 1)
		if blockIndex >= blockCount {
			break
		}

		start := blockIndex * blockSize

		end := min(start+blockSize, itemCount)

		callback(start, end, workerIndex)
	}
}

func parallelFor(w *World, callback parallelForFunc, itemCount, minRange int) {
	if itemCount <= 0 {
		return
	}

	core.Assert(minRange > 0, "minRange > 0")

	workerCount := w.workerCount
	core.Assert(0 < workerCount && workerCount <= MaxWorkers, "0 < workerCount && workerCount <= B2_MAX_WORKERS")
	// Target multiple blocks per worker to reduce thread stalls.
	// block size grows once items exceed maxBlockCount * minRange
	// so the block count stays bounded and per-block sync overhead stays low.
	blocksPerWorker := 8
	maxBlockCount := blocksPerWorker * workerCount

	var blockSize, blockCount int
	if itemCount <= minRange*maxBlockCount {
		blockSize = minRange
		blockCount = (itemCount + blockSize - 1) / blockSize
	} else {
		blockSize = (itemCount + maxBlockCount - 1) / maxBlockCount
		blockCount = (itemCount + blockSize - 1) / blockSize
	}

	core.Assert(blockCount >= 1, "blockCount >= 1")
	core.Assert(blockSize*blockCount >= itemCount, "blockSize * blockCount >= itemCount")
	// No point enqueueing more tasks than blocks.
	taskCount := min(workerCount, blockCount)
	shared := &w.parallelForScratch
	shared.blockCount = blockCount
	shared.blockSize = blockSize
	shared.itemCount = itemCount
	shared.callback = callback
	shared.nextBlock.Store(0)

	tasks := w.parallelTaskScratch[:taskCount]
	handles := w.parallelHandleScratch[:taskCount]

	for i := range taskCount {
		tasks[i].shared = shared

		tasks[i].workerIndex = i
		if w.taskCount < MaxTasks {
			handles[i] = w.enqueueTaskFcn(func() { parallelForTrampoline(&tasks[i]) })
			w.taskCount++
		} else {
			parallelForTrampoline(&tasks[i])
		}
	}

	for i := range taskCount {
		if handles[i] != nil {
			handles[i]()
		}
	}

	clear(tasks)
	clear(handles)

	w.parallelForScratch = parallelForShared{}
}
