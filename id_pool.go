// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type idPool struct {
	freeArray []int
	nextIndex int
}

func createIDPool() idPool {
	todo()
	var result idPool
	return result
}

func destroyIDPool(pool *idPool) {
	todo()
}

func allocID(pool *idPool) int {
	todo()
	var result int
	return result
}

func freeID(pool *idPool, id int) {
	todo()
}

func validateFreeID(pool *idPool, id int) {
	todo()
}

func validateUsedID(pool *idPool, id int) {
	todo()
}

func getIDCount(pool *idPool) int {
	todo()
	var result int
	return result
}

func getIDCapacity(pool *idPool) int {
	todo()
	var result int
	return result
}

func getIDBytes(pool *idPool) int {
	todo()
	var result int
	return result
}
