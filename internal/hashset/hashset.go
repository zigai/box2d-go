// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package hashset

import "github.com/zigai/box2d-go/internal/core"

type Item struct {
	Key uint64
}

type Set struct {
	Items    []Item
	Capacity uint32
	Count    uint32
}

func New(capacity int) Set {
	var set Set
	if capacity > 16 {
		set.Capacity = uint32(core.RoundUpPowerOf2(capacity))
	} else {
		set.Capacity = 16
	}

	set.Items = make([]Item, set.Capacity)

	return set
}

func (s *Set) Destroy() {
	s.Items = nil
	s.Count = 0
	s.Capacity = 0
}

func (s *Set) Clear() {
	s.Count = 0
	clear(s.Items[:s.Capacity])
}

func (s *Set) Contains(key uint64) bool {
	core.Validate(key != 0, "key != 0")
	hash := Hash(key)
	index := s.findSlot(key, hash)

	return s.Items[index].Key == key
}

func (s *Set) ContainsHashed(key uint64, hash uint64) bool {
	core.Validate(key != 0, "key != 0")
	index := s.findSlot(key, hash)

	return s.Items[index].Key == key
}

func (s *Set) Bytes() int {
	return int(s.Capacity) * 8
}

// Add inserts key and reports whether it was already present.
func (s *Set) Add(key uint64) bool {
	core.Assert(key != 0, "key != 0")
	hash := Hash(key)
	core.Assert(hash != 0, "hash != 0")

	index := s.findSlot(key, hash)
	if s.Items[index].Key != 0 {
		core.Assert(s.Items[index].Key == key, "items[index].key == key")
		return true
	}

	if 2*s.Count >= s.Capacity {
		s.grow()
	}

	s.addKeyHaveCapacity(key, hash)

	return false
}

func (s *Set) Remove(key uint64) bool {
	hash := Hash(key)
	i := s.findSlot(key, hash)

	items := s.Items
	if items[i].Key == 0 {
		return false
	}

	items[i].Key = 0

	core.Assert(s.Count > 0, "set.count > 0")
	s.Count--
	j := i

	capacity := s.Capacity
	for {
		j = int((uint32(j) + 1) & (capacity - 1))
		if items[j].Key == 0 {
			break
		}

		hashJ := Hash(items[j].Key)

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
		items[j].Key = 0
		i = j
	}

	return true
}

func (s *Set) findSlot(key uint64, hash uint64) int {
	capacity := s.Capacity
	index := uint32(hash) & (capacity - 1)

	items := s.Items
	for items[index].Key != 0 && items[index].Key != key {
		index = (index + 1) & (capacity - 1)
	}

	return int(index)
}

func (s *Set) addKeyHaveCapacity(key uint64, hash uint64) {
	index := s.findSlot(key, hash)
	core.Assert(s.Items[index].Key == 0, "items[index].key == 0")
	s.Items[index].Key = key
	s.Count++
}

func (s *Set) grow() {
	oldCount := s.Count
	oldCapacity := s.Capacity
	oldItems := s.Items
	s.Count = 0
	s.Capacity = 2 * oldCapacity

	s.Items = nil
	if s.Capacity > 0 {
		s.Items = make([]Item, s.Capacity)
	}

	for i := range oldCapacity {
		item := &oldItems[i]
		if item.Key == 0 {
			continue
		}

		hash := Hash(item.Key)
		s.addKeyHaveCapacity(item.Key, hash)
	}

	core.Assert(s.Count == oldCount, "set.count == oldCount")
}

func Hash(key uint64) uint64 {
	h := key
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33

	return h
}

func ShapePairKey(k1 int, k2 int) uint64 {
	if k1 < k2 {
		return uint64(k1)<<32 | uint64(k2)
	}

	return uint64(k2)<<32 | uint64(k1)
}
