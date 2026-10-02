// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"testing"

	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
)

const tableItemCount = (317*317 - 317) / 2

func tableBoolByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}

func TestTable(t *testing.T) {
	defer porttest.Guard(t)
	r := oracle.Open(t, "table/TableTest")

	power := boundingPowerOf2(3008)
	porttest.Ensure(t, power == 12)
	r.I32("line_427", int32(power))

	nextPowerOf2 := roundUpPowerOf2(3008)
	porttest.Ensure(t, nextPowerOf2 == (1<<power))
	r.I32("line_430", int32(nextPowerOf2))

	t.Run("BasicHashSetTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/BasicHashSetTest")

		set := createSet(16)
		porttest.Ensure(t, getSetCount(&set) == 0)
		r.I32("line_18", int32(getSetCount(&set)))
		porttest.Ensure(t, getSetCapacity(&set) == 16)
		r.I32("line_19", int32(getSetCapacity(&set)))

		destroySet(&set)
		porttest.Ensure(t, set.items == nil)
		r.Bool("line_22", set.items == nil)
		porttest.Ensure(t, set.count == 0)
		r.I32("line_23", int32(set.count))
		porttest.Ensure(t, set.capacity == 0)
		r.I32("line_24", int32(set.capacity))

	})
	t.Run("HashSetCapacityTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetCapacityTest")

		{
			set := createSet(1)
			porttest.Ensure(t, getSetCapacity(&set) == 16)
			r.I32("line_34", int32(getSetCapacity(&set)))
			destroySet(&set)
		}

		{
			set := createSet(15)
			porttest.Ensure(t, getSetCapacity(&set) == 16)
			r.I32("line_40", int32(getSetCapacity(&set)))
			destroySet(&set)
		}

		{
			set := createSet(32)
			porttest.Ensure(t, getSetCapacity(&set) == 32)
			r.I32("line_46", int32(getSetCapacity(&set)))
			destroySet(&set)
		}

		{
			set := createSet(33)
			porttest.Ensure(t, getSetCapacity(&set) == 64)
			r.I32("line_52", int32(getSetCapacity(&set)))
			destroySet(&set)
		}

	})
	t.Run("HashSetAddRemoveTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetAddRemoveTest")

		set := createSet(16)

		found := addKey(&set, 42)
		porttest.Ensure(t, found == false)
		r.Bool("line_65", found)
		porttest.Ensure(t, getSetCount(&set) == 1)
		r.I32("line_66", int32(getSetCount(&set)))

		found = addKey(&set, 123)
		porttest.Ensure(t, found == false)
		r.Bool("line_69", found)
		porttest.Ensure(t, getSetCount(&set) == 2)
		r.I32("line_70", int32(getSetCount(&set)))

		found = addKey(&set, 42)
		porttest.Ensure(t, found == true)
		r.Bool("line_74", found)
		porttest.Ensure(t, getSetCount(&set) == 2)
		r.I32("line_75", int32(getSetCount(&set)))

		porttest.Ensure(t, containsKey(&set, 42) == true)
		r.Bool("line_78", containsKey(&set, 42))
		porttest.Ensure(t, containsKey(&set, 123) == true)
		r.Bool("line_79", containsKey(&set, 123))
		porttest.Ensure(t, containsKey(&set, 999) == false)
		r.Bool("line_80", containsKey(&set, 999))

		removed := removeKey(&set, 42)
		porttest.Ensure(t, removed == true)
		r.Bool("line_84", removed)
		porttest.Ensure(t, getSetCount(&set) == 1)
		r.I32("line_85", int32(getSetCount(&set)))
		porttest.Ensure(t, containsKey(&set, 42) == false)
		r.Bool("line_86", containsKey(&set, 42))
		porttest.Ensure(t, containsKey(&set, 123) == true)
		r.Bool("line_87", containsKey(&set, 123))

		removed = removeKey(&set, 999)
		porttest.Ensure(t, removed == false)
		r.Bool("line_91", removed)
		porttest.Ensure(t, getSetCount(&set) == 1)
		r.I32("line_92", int32(getSetCount(&set)))

		removed = removeKey(&set, 42)
		porttest.Ensure(t, removed == false)
		r.Bool("line_96", removed)
		porttest.Ensure(t, getSetCount(&set) == 1)
		r.I32("line_97", int32(getSetCount(&set)))

		destroySet(&set)

	})
	t.Run("HashSetClearTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetClearTest")

		set := createSet(16)

		addKey(&set, 10)
		addKey(&set, 20)
		addKey(&set, 30)
		porttest.Ensure(t, getSetCount(&set) == 3)
		r.I32("line_111", int32(getSetCount(&set)))

		clearSet(&set)
		porttest.Ensure(t, getSetCount(&set) == 0)
		r.I32("line_115", int32(getSetCount(&set)))
		porttest.Ensure(t, containsKey(&set, 10) == false)
		r.Bool("line_116", containsKey(&set, 10))
		porttest.Ensure(t, containsKey(&set, 20) == false)
		r.Bool("line_117", containsKey(&set, 20))
		porttest.Ensure(t, containsKey(&set, 30) == false)
		r.Bool("line_118", containsKey(&set, 30))

		addKey(&set, 40)
		porttest.Ensure(t, getSetCount(&set) == 1)
		r.I32("line_122", int32(getSetCount(&set)))
		porttest.Ensure(t, containsKey(&set, 40) == true)
		r.Bool("line_123", containsKey(&set, 40))

		destroySet(&set)

	})
	t.Run("HashSetGrowthTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetGrowthTest")

		set := createSet(16)
		initialCapacity := getSetCapacity(&set)

		for i := uint64(0); i < 8; i++ {
			addKey(&set, i+1)
		}

		newCapacity := getSetCapacity(&set)
		porttest.Ensure(t, newCapacity >= initialCapacity)
		r.I32("line_143", int32(newCapacity))
		r.I32("line_143_initial", int32(initialCapacity))
		porttest.Ensure(t, getSetCount(&set) == 8)
		r.I32("line_144", int32(getSetCount(&set)))

		for i := uint64(1); i <= 8; i++ {
			porttest.Ensure(t, containsKey(&set, i) == true)
			r.Bool("line_149", containsKey(&set, i))
		}

		destroySet(&set)

	})
	t.Run("HashSetEdgeCasesTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetEdgeCasesTest")

		set := createSet(16)

		largeKey := uint64(0xFFFFFFFFFFFFFFFF) - 1
		addKey(&set, largeKey)
		porttest.Ensure(t, containsKey(&set, largeKey) == true)
		r.Bool("line_163", containsKey(&set, largeKey))
		porttest.Ensure(t, getSetCount(&set) == 1)
		r.I32("line_164", int32(getSetCount(&set)))

		key1 := uint64(0x123456789ABCDEF)
		key2 := uint64(0x987654321FEDCBA)
		addKey(&set, key1)
		addKey(&set, key2)
		porttest.Ensure(t, containsKey(&set, key1) == true)
		r.Bool("line_171", containsKey(&set, key1))
		porttest.Ensure(t, containsKey(&set, key2) == true)
		r.Bool("line_172", containsKey(&set, key2))

		for i := uint64(0x1000); i < 0x1010; i++ {
			addKey(&set, i)
		}

		for i := uint64(0x1000); i < 0x1010; i++ {
			porttest.Ensure(t, containsKey(&set, i) == true)
			r.Bool("line_182", containsKey(&set, i))
		}

		destroySet(&set)

	})
	t.Run("HashSetRemovalReorganizationTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetRemovalReorganizationTest")

		set := createSet(16)

		keys := [...]uint64{100, 116, 132, 148, 164}
		keyCount := len(keys)

		for i := 0; i < keyCount; i++ {
			addKey(&set, keys[i])
		}

		for i := 0; i < keyCount; i++ {
			porttest.Ensure(t, containsKey(&set, keys[i]) == true)
			r.Bool("line_205", containsKey(&set, keys[i]))
		}

		removeKey(&set, keys[2])
		porttest.Ensure(t, containsKey(&set, keys[2]) == false)
		r.Bool("line_210", containsKey(&set, keys[2]))

		for i := 0; i < keyCount; i++ {
			if i != 2 {
				porttest.Ensure(t, containsKey(&set, keys[i]) == true)
				r.Bool("line_217", containsKey(&set, keys[i]))
			}
		}

		destroySet(&set)

	})
	t.Run("HashSetStressTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetStressTest")

		set := createSet(32)

		testSize := 1000
		var keys [1000]uint64

		for i := 0; i < testSize; i++ {
			keys[i] = uint64(i*7 + 13)
		}

		for i := 0; i < testSize; i++ {
			found := addKey(&set, keys[i])
			porttest.Ensure(t, found == false)
			r.Bool("line_244", found)
		}

		porttest.Ensure(t, getSetCount(&set) == testSize)
		r.I32("line_247", int32(getSetCount(&set)))

		for i := 0; i < testSize; i++ {
			porttest.Ensure(t, containsKey(&set, keys[i]) == true)
			r.Bool("line_252", containsKey(&set, keys[i]))
		}

		removedCount := 0
		for i := 0; i < testSize; i += 2 {
			removed := removeKey(&set, keys[i])
			porttest.Ensure(t, removed == true)
			r.Bool("line_260", removed)
			removedCount++
		}

		porttest.Ensure(t, getSetCount(&set) == testSize-removedCount)
		r.I32("line_264", int32(getSetCount(&set)))

		for i := 0; i < testSize; i++ {
			shouldBePresent := (i%2 == 1)
			porttest.Ensure(t, containsKey(&set, keys[i]) == shouldBePresent)
			r.Bool("line_270", containsKey(&set, keys[i]))
		}

		destroySet(&set)

	})
	t.Run("HashSetShapePairKeyTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetShapePairKeyTest")

		set := createSet(16)

		key1 := shapePairKey(5, 10)
		key2 := shapePairKey(10, 5)
		porttest.Ensure(t, key1 == key2)
		r.U64("line_284", key1)
		r.U64("line_284_other", key2)

		addKey(&set, key1)
		porttest.Ensure(t, containsKey(&set, key1) == true)
		r.Bool("line_287", containsKey(&set, key1))
		porttest.Ensure(t, containsKey(&set, key2) == true)
		r.Bool("line_288", containsKey(&set, key2))

		key3 := shapePairKey(1, 2)
		key4 := shapePairKey(2, 3)
		porttest.Ensure(t, key3 != key4)
		r.U64("line_293", key3)
		r.U64("line_293_other", key4)

		addKey(&set, key3)
		addKey(&set, key4)
		porttest.Ensure(t, getSetCount(&set) == 3)
		r.I32("line_297", int32(getSetCount(&set)))

		destroySet(&set)

	})
	t.Run("HashSetBytesTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetBytesTest")

		set := createSet(32)

		bytes := getHashSetBytes(&set)
		expectedBytes := 32 * 8
		porttest.Ensure(t, bytes == expectedBytes)
		r.I32("line_309", int32(bytes))

		addKey(&set, 100)
		addKey(&set, 200)

		bytesAfterAdd := getHashSetBytes(&set)
		porttest.Ensure(t, bytesAfterAdd == expectedBytes)
		r.I32("line_316", int32(bytesAfterAdd))

		destroySet(&set)

	})
	t.Run("HashSetTest", func(t *testing.T) {
		defer porttest.Guard(t)
		r := oracle.Open(t, "table/HashSetTest")

		var checks []byte

		N := 317
		itemCount := tableItemCount
		var removed [tableItemCount]bool

		for iter := 0; iter < 1; iter++ {
			set := createSet(16)

			for i := 0; i < N; i++ {
				for j := i + 1; j < N; j++ {
					key := shapePairKey(i, j)
					found := addKey(&set, key)
					porttest.Ensure(t, found == false)
					checks = append(checks, tableBoolByte(found))
				}
			}

			r.Bytes("fill", checks)
			checks = checks[:0]
			porttest.Ensure(t, getSetCount(&set) == itemCount)
			r.I32("line_343", int32(getSetCount(&set)))

			k := 0
			removeCount := 0
			for i := 0; i < N; i++ {
				for j := i + 1; j < N; j++ {
					if j == i+1 {
						key := shapePairKey(i, j)
						size1 := getSetCount(&set)
						found := removeKey(&set, key)
						porttest.Ensure(t, found)
						r.Bool("line_357", found)
						size2 := getSetCount(&set)
						porttest.Ensure(t, size2 == size1-1)
						r.I32("line_359", int32(size2))
						removed[k] = true
						k++
						removeCount += 1
					} else {
						removed[k] = false
						k++
					}
				}
			}

			porttest.Ensure(t, getSetCount(&set) == (itemCount-removeCount))
			r.I32("line_370", int32(getSetCount(&set)))

			ticks := GetTicks()

			k = 0
			for i := 0; i < N; i++ {
				for j := i + 1; j < N; j++ {
					key := shapePairKey(j, i)
					found := containsKey(&set, key)
					porttest.Ensure(t, found || removed[k])
					checks = append(checks, tableBoolByte(found))
					checks = append(checks, tableBoolByte(removed[k]))
					k += 1
				}
			}

			r.Bytes("search", checks)
			checks = checks[:0]

			ms := GetMilliseconds(ticks)
			t.Logf("set: count = %d, containsKey = %.5f ms, ave = %.5f us\n", itemCount, ms, 1000.0*ms/float32(itemCount))

			for i := 0; i < N; i++ {
				for j := i + 1; j < N; j++ {
					key := shapePairKey(i, j)
					removeKey(&set, key)
				}
			}

			porttest.Ensure(t, getSetCount(&set) == 0)
			r.I32("line_415", int32(getSetCount(&set)))

			destroySet(&set)
		}

	})
}
