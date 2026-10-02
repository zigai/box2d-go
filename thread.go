// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "sync"

type threadFunction func()

type semaphore struct {
	mutex     sync.Mutex
	condition *sync.Cond
	count     int
}

type thread struct {
	done     chan struct{}
	function threadFunction
	name     string
}

func createMutex() *sync.Mutex {
	todo()
	return nil
}

func destroyMutex(m *sync.Mutex) {
	todo()
}

func lockMutex(m *sync.Mutex) {
	todo()
}

func unlockMutex(m *sync.Mutex) {
	todo()
}

func createSemaphore(initCount int) *semaphore {
	todo()
	return nil
}

func destroySemaphore(s *semaphore) {
	todo()
}

func waitSemaphore(s *semaphore) {
	todo()
}

func signalSemaphore(s *semaphore) {
	todo()
}

func setCurrentThreadName(name string) {
	todo()
}

func threadStart(t *thread) {
	todo()
}

func createThread(function threadFunction, name string) *thread {
	todo()
	return nil
}

func joinThread(t *thread) {
	todo()
}
