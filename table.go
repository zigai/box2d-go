// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math/bits"

type setItem struct {
	key uint64
}

type hashSet struct {
	items    []setItem
	capacity uint32
	count    uint32
}

func createSet(capacity int) hashSet {
	var set hashSet
	if capacity > 16 {
		set.capacity = uint32(roundUpPowerOf2(capacity))
	} else {
		set.capacity = 16
	}
	set.items = make([]setItem, set.capacity)
	return set
}

func destroySet(set *hashSet) {
	set.items = nil
	set.count = 0
	set.capacity = 0
}

func clearSet(set *hashSet) {
	set.count = 0
	clear(set.items[:set.capacity])
}

func findSlot(set *hashSet, key uint64, hash uint64) int {
	capacity := set.capacity
	index := uint32(hash) & (capacity - 1)
	items := set.items
	for items[index].key != 0 && items[index].key != key {
		index = (index + 1) & (capacity - 1)
	}
	return int(index)
}

func addKeyHaveCapacity(set *hashSet, key uint64, hash uint64) {
	index := findSlot(set, key, hash)
	assert(set.items[index].key == 0, "items[index].key == 0")
	set.items[index].key = key
	set.count++
}

func growTable(set *hashSet) {
	oldCount := set.count
	oldCapacity := set.capacity
	oldItems := set.items
	set.count = 0
	set.capacity = 2 * oldCapacity
	set.items = nil
	if set.capacity > 0 {
		set.items = make([]setItem, set.capacity)
	}
	for i := uint32(0); i < oldCapacity; i++ {
		item := &oldItems[i]
		if item.key == 0 {
			continue
		}
		hash := keyHash(item.key)
		addKeyHaveCapacity(set, item.key, hash)
	}
	assert(set.count == oldCount, "set.count == oldCount")
}

func containsKey(set *hashSet, key uint64) bool {
	validate(key != 0, "key != 0")
	hash := keyHash(key)
	index := findSlot(set, key, hash)
	return set.items[index].key == key
}

func containsHashedKey(set *hashSet, key uint64, hash uint64) bool {
	validate(key != 0, "key != 0")
	index := findSlot(set, key, hash)
	return set.items[index].key == key
}

func getHashSetBytes(set *hashSet) int {
	return int(set.capacity) * 8
}

func addKey(set *hashSet, key uint64) bool {
	assert(key != 0, "key != 0")
	hash := keyHash(key)
	assert(hash != 0, "hash != 0")
	index := findSlot(set, key, hash)
	if set.items[index].key != 0 {
		assert(set.items[index].key == key, "items[index].key == key")
		return true
	}
	if 2*set.count >= set.capacity {
		growTable(set)
	}
	addKeyHaveCapacity(set, key, hash)
	return false
}

func removeKey(set *hashSet, key uint64) bool {
	hash := keyHash(key)
	i := findSlot(set, key, hash)
	items := set.items
	if items[i].key == 0 {
		return false
	}
	items[i].key = 0
	assert(set.count > 0, "set.count > 0")
	set.count--
	j := i
	capacity := set.capacity
	for {
		j = int((uint32(j) + 1) & (capacity - 1))
		if items[j].key == 0 {
			break
		}
		hashJ := keyHash(items[j].key)
		k := int(hashJ & uint64(capacity-1))
		if i <= j {
			if i < k && k <= j {
				continue
			}
		} else {
			if i < k || k <= j {
				continue
			}
		}
		items[i] = items[j]
		items[j].key = 0
		i = j
	}
	return true
}

func countSetBits(bitSet *bitSet) int {
	popCount := 0
	for i := uint32(0); i < bitSet.blockCount; i++ {
		popCount += bits.OnesCount64(bitSet.bits[i])
	}
	return popCount
}

func keyHash(key uint64) uint64 {
	h := key
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	return h
}

func getSetCount(set *hashSet) int {
	return int(set.count)
}

func getSetCapacity(set *hashSet) int {
	return int(set.capacity)
}

func shapePairKey(k1, k2 int) uint64 {
	if k1 < k2 {
		return uint64(k1)<<32 | uint64(k2)
	}
	return uint64(k2)<<32 | uint64(k1)
}
