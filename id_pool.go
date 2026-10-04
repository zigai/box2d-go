// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "slices"

type idPool struct {
	freeArray []int
	nextIndex int
}

func createIDPool() idPool {
	return idPool{freeArray: make([]int, 0, 32)}
}

func destroyIDPool(pool *idPool) {
	*pool = idPool{}
}

func allocID(pool *idPool) int {
	if len(pool.freeArray) > 0 {
		var id int

		pool.freeArray, id = arrayPop(pool.freeArray)

		return id
	}

	id := pool.nextIndex
	pool.nextIndex++

	return id
}

func freeID(pool *idPool, id int) {
	assert(pool.nextIndex > 0, "pool.nextIndex > 0")
	assert(0 <= id && id < pool.nextIndex, "0 <= id && id < pool.nextIndex")
	pool.freeArray = arrayPush(pool.freeArray, id)
}

func validateFreeID(pool *idPool, id int) {
	if !validationEnabled {
		return
	}

	if slices.Contains(pool.freeArray, id) {
		return
	}

	assert(false, "0")
}

func validateUsedID(pool *idPool, id int) {
	if !validationEnabled {
		return
	}

	for _, freeID := range pool.freeArray {
		if freeID == id {
			assert(false, "0")
		}
	}
}

func getIDCount(pool *idPool) int {
	return pool.nextIndex - len(pool.freeArray)
}

func getIDCapacity(pool *idPool) int {
	return pool.nextIndex
}

func getIDBytes(pool *idPool) int {
	return cap(pool.freeArray) * 4
}
