// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "sync/atomic"

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
		end := start + blockSize
		if end > itemCount {
			end = itemCount
		}
		callback(start, end, workerIndex)
	}
}

func parallelFor(w *World, callback parallelForFunc, itemCount, minRange int) {
	if itemCount <= 0 {
		return
	}
	assert(minRange > 0, "minRange > 0")
	workerCount := w.workerCount
	assert(0 < workerCount && workerCount <= MaxWorkers, "0 < workerCount && workerCount <= B2_MAX_WORKERS")
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
	assert(blockCount >= 1, "blockCount >= 1")
	assert(blockSize*blockCount >= itemCount, "blockSize * blockCount >= itemCount")
	// No point enqueueing more tasks than blocks.
	taskCount := min(workerCount, blockCount)
	shared := parallelForShared{blockCount: blockCount, blockSize: blockSize, itemCount: itemCount, callback: callback}
	shared.nextBlock.Store(0)
	var tasks [MaxWorkers]parallelForTask
	var handles [MaxWorkers]func()
	for i := 0; i < taskCount; i++ {
		tasks[i].shared = &shared
		tasks[i].workerIndex = i
		if w.taskCount < MaxTasks {
			handles[i] = w.enqueueTaskFcn(func() { parallelForTrampoline(&tasks[i]) })
			w.taskCount++
		} else {
			parallelForTrampoline(&tasks[i])
		}
	}
	for i := 0; i < taskCount; i++ {
		if handles[i] != nil {
			handles[i]()
		}
	}
}
