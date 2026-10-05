// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type bitSet struct {
	bits          []uint64
	blockCapacity uint32
	blockCount    uint32
}

func createBitSet(bitCapacity uint32) bitSet {
	blockCapacity := uint32((uint64(bitCapacity) + 63) / 64)

	var data []uint64
	if blockCapacity > 0 {
		data = make([]uint64, blockCapacity)
	}

	return bitSet{bits: data, blockCapacity: blockCapacity}
}

func destroyBitSet(bitSet *bitSet) {
	bitSet.bits = nil
	bitSet.blockCapacity = 0
	bitSet.blockCount = 0
}

func setBitCountAndClear(bitSet *bitSet, bitCount uint32) {
	blockCount := uint32((uint64(bitCount) + 63) / 64)
	if bitSet.blockCapacity < blockCount {
		destroyBitSet(bitSet)

		newBitCapacity := bitCount + (bitCount >> 1)
		*bitSet = createBitSet(newBitCapacity)
	}

	bitSet.blockCount = blockCount
	clear(bitSet.bits[:blockCount])
}

func growBitSet(bitSet *bitSet, blockCount uint32) {
	assert(blockCount > bitSet.blockCount, "blockCount > bitSet.blockCount")

	if blockCount > bitSet.blockCapacity {
		oldCapacity := bitSet.blockCapacity
		bitSet.blockCapacity = blockCount + blockCount/2
		newBits := make([]uint64, bitSet.blockCapacity)
		assert(bitSet.bits != nil, "bitSet.bits != NULL")
		copy(newBits, bitSet.bits[:oldCapacity])
		bitSet.bits = newBits
	}

	bitSet.blockCount = blockCount
}

func inPlaceUnion(setA *bitSet, setB *bitSet) {
	assert(setA.blockCount == setB.blockCount, "setA.blockCount == setB.blockCount")

	for i := range uint32(setA.blockCount) {
		setA.bits[i] |= setB.bits[i]
	}
}

func setBit(bitSet *bitSet, bitIndex uint32) {
	blockIndex := bitIndex / 64
	assert(blockIndex < bitSet.blockCount, "blockIndex < bitSet.blockCount")
	bitSet.bits[blockIndex] |= uint64(1) << (bitIndex % 64)
}

func setBitGrow(bitSet *bitSet, bitIndex uint32) {
	blockIndex := bitIndex / 64
	if blockIndex >= bitSet.blockCount {
		growBitSet(bitSet, blockIndex+1)
	}

	bitSet.bits[blockIndex] |= uint64(1) << (bitIndex % 64)
}

func clearBit(bitSet *bitSet, bitIndex uint32) {
	blockIndex := bitIndex / 64
	if blockIndex >= bitSet.blockCount {
		return
	}

	bitSet.bits[blockIndex] &^= uint64(1) << (bitIndex % 64)
}

func getBit(bitSet *bitSet, bitIndex uint32) bool {
	blockIndex := bitIndex / 64
	if blockIndex >= bitSet.blockCount {
		return false
	}

	return bitSet.bits[blockIndex]&(uint64(1)<<(bitIndex%64)) != 0
}
