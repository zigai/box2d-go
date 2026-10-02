// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"testing"
)

func TestThread(t *testing.T) {

	t.Run("SemaphoreCreateDestroyTest", func(t *testing.T) {

		s := createSemaphore(0)
		if !(s != nil) {
			t.Fatal("condition false")
		}
		destroySemaphore(s)
	})
	t.Run("SemaphoreSignalWaitTest", func(t *testing.T) {

		sem := createSemaphore(0)
		value := 0
		worker := createThread(func() {
			value = 99
			signalSemaphore(sem)
		}, "sem test")
		waitSemaphore(sem)
		if !(value == 99) {
			t.Fatal("condition false")
		}
		joinThread(worker)
		destroySemaphore(sem)
	})
	t.Run("SemaphoreInitialCountTest", func(t *testing.T) {

		s := createSemaphore(3)
		waitSemaphore(s)
		waitSemaphore(s)
		waitSemaphore(s)
		signalSemaphore(s)
		waitSemaphore(s)
		destroySemaphore(s)
	})
	t.Run("ThreadCreateJoinTest", func(t *testing.T) {

		sem := createSemaphore(0)
		value := 0
		worker := createThread(func() {
			value = 99
			signalSemaphore(sem)
		}, "join test")
		joinThread(worker)
		if !(value == 99) {
			t.Fatal("condition false")
		}
		destroySemaphore(sem)
	})
	t.Run("ThreadMultipleTest", func(t *testing.T) {

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
		if !(sum == threadCount*1000) {
			t.Fatal("condition false")
		}
		destroyMutex(mutex)
	})
}
