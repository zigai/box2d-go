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
	return new(sync.Mutex)
}

func destroyMutex(m *sync.Mutex) {}

func lockMutex(m *sync.Mutex) {
	m.Lock()
}

func unlockMutex(m *sync.Mutex) {
	m.Unlock()
}

func createSemaphore(initCount int) *semaphore {
	s := &semaphore{count: initCount}
	s.condition = sync.NewCond(&s.mutex)
	return s
}

func destroySemaphore(s *semaphore) {}

func waitSemaphore(s *semaphore) {
	s.mutex.Lock()
	for s.count == 0 {
		s.condition.Wait()
	}
	s.count--
	s.mutex.Unlock()
}

func signalSemaphore(s *semaphore) {
	s.mutex.Lock()
	s.count++
	s.condition.Signal()
	s.mutex.Unlock()
}

func setCurrentThreadName(name string) {}

func threadStart(t *thread) {
	defer close(t.done)
	t.function()
}

func createThread(function threadFunction, name string) *thread {
	t := &thread{done: make(chan struct{}), function: function, name: name}
	go threadStart(t)
	return t
}

func joinThread(t *thread) {
	<-t.done
}
