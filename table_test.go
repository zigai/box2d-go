// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

const tableItemCount = (317*317 - 317) / 2

func tableBoolByte(v bool) byte {
	if v {
		return 1
	}

	return 0
}

func TestTable(t *testing.T) {
	r := oracle.Open(t, "table/TableTest")

	power := boundingPowerOf2(3008)
	if power != 12 {
		t.Fatal("condition false")
	}

	r.I32("line_427", int32(power))

	nextPowerOf2 := roundUpPowerOf2(3008)
	if nextPowerOf2 != (1 << power) {
		t.Fatal("condition false")
	}

	r.I32("line_430", int32(nextPowerOf2))

	t.Run("BasicHashSetTest", runTableBasicHashSetTest)
	t.Run("HashSetCapacityTest", runTableHashSetCapacityTest)
	t.Run("HashSetAddRemoveTest", runTableHashSetAddRemoveTest)
	t.Run("HashSetClearTest", runTableHashSetClearTest)
	t.Run("HashSetGrowthTest", runTableHashSetGrowthTest)
	t.Run("HashSetEdgeCasesTest", runTableHashSetEdgeCasesTest)
	t.Run("HashSetRemovalReorganizationTest", runTableHashSetRemovalReorganizationTest)
	t.Run("HashSetStressTest", runTableHashSetStressTest)
	t.Run("HashSetShapePairKeyTest", runTableHashSetShapePairKeyTest)
	t.Run("HashSetBytesTest", runTableHashSetBytesTest)
	t.Run("HashSetTest", runTableHashSetTest)
}

func runTableBasicHashSetTest(t *testing.T) {
	r := oracle.Open(t, "table/BasicHashSetTest")

	set := createSet(16)
	if getSetCount(&set) != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_18", int32(getSetCount(&set)))

	if getSetCapacity(&set) != 16 {
		t.Fatal("condition false")
	}

	r.I32("line_19", int32(getSetCapacity(&set)))

	destroySet(&set)

	if set.items != nil {
		t.Fatal("condition false")
	}

	r.Bool("line_22", set.items == nil)

	if set.count != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_23", int32(set.count))

	if set.capacity != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_24", int32(set.capacity))
}

func runTableHashSetCapacityTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetCapacityTest")

	{
		set := createSet(1)
		if getSetCapacity(&set) != 16 {
			t.Fatal("condition false")
		}

		r.I32("line_34", int32(getSetCapacity(&set)))
		destroySet(&set)
	}

	{
		set := createSet(15)
		if getSetCapacity(&set) != 16 {
			t.Fatal("condition false")
		}

		r.I32("line_40", int32(getSetCapacity(&set)))
		destroySet(&set)
	}

	{
		set := createSet(32)
		if getSetCapacity(&set) != 32 {
			t.Fatal("condition false")
		}

		r.I32("line_46", int32(getSetCapacity(&set)))
		destroySet(&set)
	}

	{
		set := createSet(33)
		if getSetCapacity(&set) != 64 {
			t.Fatal("condition false")
		}

		r.I32("line_52", int32(getSetCapacity(&set)))
		destroySet(&set)
	}
}

func runTableHashSetAddRemoveTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetAddRemoveTest")

	set := createSet(16)

	found := addKey(&set, 42)
	if found != false {
		t.Fatal("condition false")
	}

	r.Bool("line_65", found)

	if getSetCount(&set) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_66", int32(getSetCount(&set)))

	found = addKey(&set, 123)
	if found != false {
		t.Fatal("condition false")
	}

	r.Bool("line_69", found)

	if getSetCount(&set) != 2 {
		t.Fatal("condition false")
	}

	r.I32("line_70", int32(getSetCount(&set)))

	found = addKey(&set, 42)
	if found != true {
		t.Fatal("condition false")
	}

	r.Bool("line_74", found)

	if getSetCount(&set) != 2 {
		t.Fatal("condition false")
	}

	r.I32("line_75", int32(getSetCount(&set)))

	if containsKey(&set, 42) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_78", containsKey(&set, 42))

	if containsKey(&set, 123) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_79", containsKey(&set, 123))

	if containsKey(&set, 999) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_80", containsKey(&set, 999))

	removed := removeKey(&set, 42)
	if removed != true {
		t.Fatal("condition false")
	}

	r.Bool("line_84", removed)

	if getSetCount(&set) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_85", int32(getSetCount(&set)))

	if containsKey(&set, 42) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_86", containsKey(&set, 42))

	if containsKey(&set, 123) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_87", containsKey(&set, 123))

	removed = removeKey(&set, 999)
	if removed != false {
		t.Fatal("condition false")
	}

	r.Bool("line_91", removed)

	if getSetCount(&set) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_92", int32(getSetCount(&set)))

	removed = removeKey(&set, 42)
	if removed != false {
		t.Fatal("condition false")
	}

	r.Bool("line_96", removed)

	if getSetCount(&set) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_97", int32(getSetCount(&set)))

	destroySet(&set)
}

func runTableHashSetClearTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetClearTest")

	set := createSet(16)

	addKey(&set, 10)
	addKey(&set, 20)
	addKey(&set, 30)

	if getSetCount(&set) != 3 {
		t.Fatal("condition false")
	}

	r.I32("line_111", int32(getSetCount(&set)))

	clearSet(&set)

	if getSetCount(&set) != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_115", int32(getSetCount(&set)))

	if containsKey(&set, 10) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_116", containsKey(&set, 10))

	if containsKey(&set, 20) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_117", containsKey(&set, 20))

	if containsKey(&set, 30) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_118", containsKey(&set, 30))

	addKey(&set, 40)

	if getSetCount(&set) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_122", int32(getSetCount(&set)))

	if containsKey(&set, 40) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_123", containsKey(&set, 40))

	destroySet(&set)
}

func runTableHashSetGrowthTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetGrowthTest")

	set := createSet(16)
	initialCapacity := getSetCapacity(&set)

	for i := range uint64(8) {
		addKey(&set, i+1)
	}

	newCapacity := getSetCapacity(&set)
	if newCapacity < initialCapacity {
		t.Fatal("condition false")
	}

	r.I32("line_143", int32(newCapacity))
	r.I32("line_143_initial", int32(initialCapacity))

	if getSetCount(&set) != 8 {
		t.Fatal("condition false")
	}

	r.I32("line_144", int32(getSetCount(&set)))

	for i := uint64(1); i <= 8; i++ {
		if containsKey(&set, i) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_149", containsKey(&set, i))
	}

	destroySet(&set)
}

func runTableHashSetEdgeCasesTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetEdgeCasesTest")

	set := createSet(16)

	largeKey := uint64(0xFFFFFFFFFFFFFFFF) - 1
	addKey(&set, largeKey)

	if containsKey(&set, largeKey) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_163", containsKey(&set, largeKey))

	if getSetCount(&set) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_164", int32(getSetCount(&set)))

	key1 := uint64(0x123456789ABCDEF)
	key2 := uint64(0x987654321FEDCBA)

	addKey(&set, key1)
	addKey(&set, key2)

	if containsKey(&set, key1) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_171", containsKey(&set, key1))

	if containsKey(&set, key2) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_172", containsKey(&set, key2))

	for i := uint64(0x1000); i < 0x1010; i++ {
		addKey(&set, i)
	}

	for i := uint64(0x1000); i < 0x1010; i++ {
		if containsKey(&set, i) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_182", containsKey(&set, i))
	}

	destroySet(&set)
}

func runTableHashSetRemovalReorganizationTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetRemovalReorganizationTest")

	set := createSet(16)

	keys := [...]uint64{100, 116, 132, 148, 164}
	keyCount := len(keys)

	for i := range keyCount {
		addKey(&set, keys[i])
	}

	for i := range keyCount {
		if containsKey(&set, keys[i]) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_205", containsKey(&set, keys[i]))
	}

	removeKey(&set, keys[2])

	if containsKey(&set, keys[2]) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_210", containsKey(&set, keys[2]))

	for i := range keyCount {
		if i != 2 {
			if containsKey(&set, keys[i]) != true {
				t.Fatal("condition false")
			}

			r.Bool("line_217", containsKey(&set, keys[i]))
		}
	}

	destroySet(&set)
}

func runTableHashSetStressTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetStressTest")

	set := createSet(32)

	testSize := 1000

	var keys [1000]uint64

	for i := range testSize {
		keys[i] = uint64(i*7 + 13)
	}

	for i := range testSize {
		found := addKey(&set, keys[i])
		if found != false {
			t.Fatal("condition false")
		}

		r.Bool("line_244", found)
	}

	if getSetCount(&set) != testSize {
		t.Fatal("condition false")
	}

	r.I32("line_247", int32(getSetCount(&set)))

	for i := range testSize {
		if containsKey(&set, keys[i]) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_252", containsKey(&set, keys[i]))
	}

	removedCount := 0

	for i := 0; i < testSize; i += 2 {
		removed := removeKey(&set, keys[i])
		if removed != true {
			t.Fatal("condition false")
		}

		r.Bool("line_260", removed)

		removedCount++
	}

	if getSetCount(&set) != testSize-removedCount {
		t.Fatal("condition false")
	}

	r.I32("line_264", int32(getSetCount(&set)))

	for i := range testSize {
		shouldBePresent := (i%2 == 1)
		if containsKey(&set, keys[i]) != shouldBePresent {
			t.Fatal("condition false")
		}

		r.Bool("line_270", containsKey(&set, keys[i]))
	}

	destroySet(&set)
}

func runTableHashSetShapePairKeyTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetShapePairKeyTest")

	set := createSet(16)

	key1 := shapePairKey(5, 10)

	key2 := shapePairKey(10, 5)
	if key1 != key2 {
		t.Fatal("condition false")
	}

	r.U64("line_284", key1)
	r.U64("line_284_other", key2)

	addKey(&set, key1)

	if containsKey(&set, key1) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_287", containsKey(&set, key1))

	if containsKey(&set, key2) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_288", containsKey(&set, key2))

	key3 := shapePairKey(1, 2)

	key4 := shapePairKey(2, 3)
	if key3 == key4 {
		t.Fatal("condition false")
	}

	r.U64("line_293", key3)
	r.U64("line_293_other", key4)

	addKey(&set, key3)
	addKey(&set, key4)

	if getSetCount(&set) != 3 {
		t.Fatal("condition false")
	}

	r.I32("line_297", int32(getSetCount(&set)))

	destroySet(&set)
}

func runTableHashSetBytesTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetBytesTest")

	set := createSet(32)

	bytes := getHashSetBytes(&set)

	expectedBytes := 32 * 8
	if bytes != expectedBytes {
		t.Fatal("condition false")
	}

	r.I32("line_309", int32(bytes))

	addKey(&set, 100)
	addKey(&set, 200)

	bytesAfterAdd := getHashSetBytes(&set)
	if bytesAfterAdd != expectedBytes {
		t.Fatal("condition false")
	}

	r.I32("line_316", int32(bytesAfterAdd))

	destroySet(&set)
}

func runTableHashSetTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetTest")

	var checks []byte

	n := 317
	itemCount := tableItemCount

	var removed [tableItemCount]bool

	for range 1 {
		set := createSet(16)

		for i := range n {
			for j := i + 1; j < n; j++ {
				key := shapePairKey(i, j)

				found := addKey(&set, key)
				if found != false {
					t.Fatal("condition false")
				}

				checks = append(checks, tableBoolByte(found))
			}
		}

		r.Bytes("fill", checks)
		checks = checks[:0]

		if getSetCount(&set) != itemCount {
			t.Fatal("condition false")
		}

		r.I32("line_343", int32(getSetCount(&set)))

		k := 0
		removeCount := 0

		for i := range n {
			for j := i + 1; j < n; j++ {
				if j == i+1 {
					key := shapePairKey(i, j)
					size1 := getSetCount(&set)

					found := removeKey(&set, key)
					if !found {
						t.Fatal("condition false")
					}

					r.Bool("line_357", found)

					size2 := getSetCount(&set)
					if size2 != size1-1 {
						t.Fatal("condition false")
					}

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

		if getSetCount(&set) != (itemCount - removeCount) {
			t.Fatal("condition false")
		}

		r.I32("line_370", int32(getSetCount(&set)))

		ticks := currentTicks()

		k = 0

		for i := range n {
			for j := i + 1; j < n; j++ {
				key := shapePairKey(j, i)

				found := containsKey(&set, key)
				if !found && !removed[k] {
					t.Fatal("condition false")
				}

				checks = append(checks, tableBoolByte(found))
				checks = append(checks, tableBoolByte(removed[k]))
				k += 1
			}
		}

		r.Bytes("search", checks)
		checks = checks[:0]

		ms := milliseconds(ticks)
		t.Logf("set: count = %d, containsKey = %.5f ms, ave = %.5f us\n", itemCount, ms, 1000.0*ms/float32(itemCount))

		for i := range n {
			for j := i + 1; j < n; j++ {
				key := shapePairKey(i, j)
				removeKey(&set, key)
			}
		}

		if getSetCount(&set) != 0 {
			t.Fatal("condition false")
		}

		r.I32("line_415", int32(getSetCount(&set)))

		destroySet(&set)
	}
}
