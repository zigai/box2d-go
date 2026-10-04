// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func TestThread(t *testing.T) {
	t.Run("SemaphoreCreateDestroyTest", func(t *testing.T) {
		r := oracle.Open(t, "thread/SemaphoreCreateDestroyTest")

		s := createSemaphore(0)

		t.Cleanup(func() { destroySemaphore(s) })

		if s == nil {
			t.Fatal("condition false")
		}

		r.Bool("created", s != nil)
		signalSemaphore(s)
		waitSemaphore(s)
		r.I32("completedWaits", 1)

		recreated := createSemaphore(1)

		t.Cleanup(func() { destroySemaphore(recreated) })
		waitSemaphore(recreated)
		r.I32("recreatedCompletedWaits", 1)
	})
	t.Run("SemaphoreSignalWaitTest", func(t *testing.T) {
		r := oracle.Open(t, "thread/SemaphoreSignalWaitTest")

		sem := createSemaphore(0)

		t.Cleanup(func() { destroySemaphore(sem) })

		value := 0
		worker := createThread(func() {
			value = 99

			signalSemaphore(sem)
		}, "sem test")

		t.Cleanup(func() { joinThread(worker) })
		waitSemaphore(sem)

		afterWait := value

		joinThread(worker)
		r.I32("afterWait", int32(afterWait))
		r.I32("afterJoin", int32(value))

		if afterWait != 99 || value != 99 {
			t.Fatal("condition false")
		}
	})
	t.Run("SemaphoreInitialCountTest", func(t *testing.T) {
		r := oracle.Open(t, "thread/SemaphoreInitialCountTest")

		s := createSemaphore(3)

		t.Cleanup(func() { destroySemaphore(s) })

		completedWaits := int32(0)

		waitSemaphore(s)

		completedWaits++
		r.I32("completedWaits", completedWaits)
		waitSemaphore(s)

		completedWaits++
		r.I32("completedWaits", completedWaits)
		waitSemaphore(s)

		completedWaits++
		r.I32("completedWaits", completedWaits)
		signalSemaphore(s)
		waitSemaphore(s)

		completedWaits++
		r.I32("completedWaits", completedWaits)
	})
	t.Run("ThreadCreateJoinTest", func(t *testing.T) {
		r := oracle.Open(t, "thread/ThreadCreateJoinTest")

		sem := createSemaphore(0)

		t.Cleanup(func() { destroySemaphore(sem) })

		value := 0
		worker := createThread(func() {
			value = 99

			signalSemaphore(sem)
		}, "join test")

		t.Cleanup(func() { joinThread(worker) })
		joinThread(worker)
		waitSemaphore(sem)
		r.I32("afterJoin", int32(value))
		r.I32("completedWaits", 1)

		if value != 99 {
			t.Fatal("condition false")
		}
	})
	t.Run("ThreadMultipleTest", func(t *testing.T) {
		r := oracle.Open(t, "thread/ThreadMultipleTest")

		mutex := createMutex()

		t.Cleanup(func() { destroyMutex(mutex) })

		sum := 0
		completed := 0

		const threadCount = 4

		var threads [threadCount]*thread
		for i := range threadCount {
			threads[i] = createThread(func() {
				for range 1000 {
					lockMutex(mutex)

					sum += 1

					unlockMutex(mutex)
				}

				lockMutex(mutex)

				completed++

				unlockMutex(mutex)
			}, fmt.Sprintf("sum test %d", i))

			t.Cleanup(func() { joinThread(threads[i]) })
		}

		joined := 0

		for i := range threadCount {
			joinThread(threads[i])

			joined++
		}

		if sum != threadCount*1000 {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum))
		r.I32("completedWorkers", int32(completed))
		r.I32("joinedWorkers", int32(joined))
	})
}
