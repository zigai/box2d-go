// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/porttest"
)

func TestThread(t *testing.T) {
	defer porttest.Guard(t)
	t.Run("SemaphoreCreateDestroyTest", func(t *testing.T) {
		defer porttest.Guard(t)
		s := createSemaphore(0)
		porttest.Ensure(t, s != nil)
		destroySemaphore(s)
	})
	t.Run("SemaphoreSignalWaitTest", func(t *testing.T) {
		defer porttest.Guard(t)
		sem := createSemaphore(0)
		value := 0
		worker := createThread(func() {
			value = 99
			signalSemaphore(sem)
		}, "sem test")
		waitSemaphore(sem)
		porttest.Ensure(t, value == 99)
		joinThread(worker)
		destroySemaphore(sem)
	})
	t.Run("SemaphoreInitialCountTest", func(t *testing.T) {
		defer porttest.Guard(t)
		s := createSemaphore(3)
		waitSemaphore(s)
		waitSemaphore(s)
		waitSemaphore(s)
		signalSemaphore(s)
		waitSemaphore(s)
		destroySemaphore(s)
	})
	t.Run("ThreadCreateJoinTest", func(t *testing.T) {
		defer porttest.Guard(t)
		sem := createSemaphore(0)
		value := 0
		worker := createThread(func() {
			value = 99
			signalSemaphore(sem)
		}, "join test")
		joinThread(worker)
		porttest.Ensure(t, value == 99)
		destroySemaphore(sem)
	})
	t.Run("ThreadMultipleTest", func(t *testing.T) {
		defer porttest.Guard(t)
		mutex := createMutex()
		sum := 0
		const threadCount = 4
		var threads [threadCount]*thread
		for i := 0; i < threadCount; i++ {
			threads[i] = createThread(func() {
				for j := 0; j < 1000; j++ {
					lockMutex(mutex)
					sum += 1
					unlockMutex(mutex)
				}
			}, fmt.Sprintf("sum test %d", i))
		}
		for i := 0; i < threadCount; i++ {
			joinThread(threads[i])
		}
		porttest.Ensure(t, sum == threadCount*1000)
		destroyMutex(mutex)
	})
}
