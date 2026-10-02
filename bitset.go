// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type bitSet struct {
	bits          []uint64
	blockCapacity uint32
	blockCount    uint32
}

func createBitSet(bitCapacity uint32) bitSet {
	todo()
	var result bitSet
	return result
}

func destroyBitSet(bitSet *bitSet) {
	todo()
}

func setBitCountAndClear(bitSet *bitSet, bitCount uint32) {
	todo()
}

func growBitSet(bitSet *bitSet, blockCount uint32) {
	todo()
}

func inPlaceUnion(setA *bitSet, setB *bitSet) {
	todo()
}

func setBit(bitSet *bitSet, bitIndex uint32) {
	todo()
}

func setBitGrow(bitSet *bitSet, bitIndex uint32) {
	todo()
}

func clearBit(bitSet *bitSet, bitIndex uint32) {
	todo()
}

func getBit(bitSet *bitSet, bitIndex uint32) bool {
	todo()
	var result bool
	return result
}

func getBitSetBytes(bitSet *bitSet) int {
	todo()
	var result int
	return result
}
