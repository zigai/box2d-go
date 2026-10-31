// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package bitset

import (
	"math/bits"

	"github.com/zigai/box2d-go/internal/core"
)

type Set struct {
	Bits          []uint64
	BlockCapacity uint32
	BlockCount    uint32
}

func New(bitCapacity uint32) Set {
	blockCapacity := uint32((uint64(bitCapacity) + 63) / 64)

	var data []uint64
	if blockCapacity > 0 {
		data = make([]uint64, blockCapacity)
	}

	return Set{Bits: data, BlockCapacity: blockCapacity}
}

func (s *Set) Destroy() {
	s.Bits = nil
	s.BlockCapacity = 0
	s.BlockCount = 0
}

func (s *Set) SetBitCountAndClear(bitCount uint32) {
	blockCount := uint32((uint64(bitCount) + 63) / 64)
	if s.BlockCapacity < blockCount {
		s.Destroy()

		newBitCapacity := bitCount + (bitCount >> 1)
		*s = New(newBitCapacity)
	}

	s.BlockCount = blockCount
	clear(s.Bits[:blockCount])
}

func (s *Set) InPlaceUnion(other *Set) {
	core.Assert(s.BlockCount == other.BlockCount, "setA.blockCount == setB.blockCount")

	for i := range uint32(s.BlockCount) {
		s.Bits[i] |= other.Bits[i]
	}
}

func (s *Set) SetBit(bitIndex uint32) {
	blockIndex := bitIndex / 64
	core.Assert(blockIndex < s.BlockCount, "blockIndex < bitSet.blockCount")
	s.Bits[blockIndex] |= uint64(1) << (bitIndex % 64)
}

func (s *Set) SetBitGrow(bitIndex uint32) {
	blockIndex := bitIndex / 64
	if blockIndex >= s.BlockCount {
		s.grow(blockIndex + 1)
	}

	s.Bits[blockIndex] |= uint64(1) << (bitIndex % 64)
}

func (s *Set) ClearBit(bitIndex uint32) {
	blockIndex := bitIndex / 64
	if blockIndex >= s.BlockCount {
		return
	}

	s.Bits[blockIndex] &^= uint64(1) << (bitIndex % 64)
}

func (s *Set) GetBit(bitIndex uint32) bool {
	blockIndex := bitIndex / 64
	if blockIndex >= s.BlockCount {
		return false
	}

	return s.Bits[blockIndex]&(uint64(1)<<(bitIndex%64)) != 0
}

func (s *Set) CountSetBits() int {
	popCount := 0
	for i := range uint32(s.BlockCount) {
		popCount += bits.OnesCount64(s.Bits[i])
	}

	return popCount
}

func (s *Set) grow(blockCount uint32) {
	core.Assert(blockCount > s.BlockCount, "blockCount > bitSet.blockCount")

	if blockCount > s.BlockCapacity {
		oldCapacity := s.BlockCapacity
		s.BlockCapacity = blockCount + blockCount/2
		newBits := make([]uint64, s.BlockCapacity)
		core.Assert(s.Bits != nil, "bitSet.bits != NULL")
		copy(newBits, s.Bits[:oldCapacity])
		s.Bits = newBits
	}

	s.BlockCount = blockCount
}
