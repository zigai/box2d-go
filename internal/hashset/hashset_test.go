// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package hashset

import (
	"testing"

	"github.com/zigai/box2d-go/internal/core"
	"github.com/zigai/box2d-go/internal/oracle"
	"github.com/zigai/box2d-go/internal/timer"
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

	power := core.BoundingPowerOf2(3008)
	if power != 12 {
		t.Fatal("condition false")
	}

	r.I32("line_427", int32(power))

	nextPowerOf2 := core.RoundUpPowerOf2(3008)
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

	set := New(16)
	if int(set.Count) != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_18", int32(int(set.Count)))

	if int(set.Capacity) != 16 {
		t.Fatal("condition false")
	}

	r.I32("line_19", int32(int(set.Capacity)))

	set.Destroy()

	if set.Items != nil {
		t.Fatal("condition false")
	}

	r.Bool("line_22", set.Items == nil)

	if set.Count != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_23", int32(set.Count))

	if set.Capacity != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_24", int32(set.Capacity))
}

func runTableHashSetCapacityTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetCapacityTest")

	{
		set := New(1)
		if int(set.Capacity) != 16 {
			t.Fatal("condition false")
		}

		r.I32("line_34", int32(int(set.Capacity)))
		set.Destroy()
	}

	{
		set := New(15)
		if int(set.Capacity) != 16 {
			t.Fatal("condition false")
		}

		r.I32("line_40", int32(int(set.Capacity)))
		set.Destroy()
	}

	{
		set := New(32)
		if int(set.Capacity) != 32 {
			t.Fatal("condition false")
		}

		r.I32("line_46", int32(int(set.Capacity)))
		set.Destroy()
	}

	{
		set := New(33)
		if int(set.Capacity) != 64 {
			t.Fatal("condition false")
		}

		r.I32("line_52", int32(int(set.Capacity)))
		set.Destroy()
	}
}

func runTableHashSetAddRemoveTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetAddRemoveTest")

	set := New(16)

	found := set.Add(42)
	if found != false {
		t.Fatal("condition false")
	}

	r.Bool("line_65", found)

	if int(set.Count) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_66", int32(int(set.Count)))

	found = set.Add(123)
	if found != false {
		t.Fatal("condition false")
	}

	r.Bool("line_69", found)

	if int(set.Count) != 2 {
		t.Fatal("condition false")
	}

	r.I32("line_70", int32(int(set.Count)))

	found = set.Add(42)
	if found != true {
		t.Fatal("condition false")
	}

	r.Bool("line_74", found)

	if int(set.Count) != 2 {
		t.Fatal("condition false")
	}

	r.I32("line_75", int32(int(set.Count)))

	if set.Contains(42) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_78", set.Contains(42))

	if set.Contains(123) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_79", set.Contains(123))

	if set.Contains(999) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_80", set.Contains(999))

	removed := set.Remove(42)
	if removed != true {
		t.Fatal("condition false")
	}

	r.Bool("line_84", removed)

	if int(set.Count) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_85", int32(int(set.Count)))

	if set.Contains(42) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_86", set.Contains(42))

	if set.Contains(123) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_87", set.Contains(123))

	removed = set.Remove(999)
	if removed != false {
		t.Fatal("condition false")
	}

	r.Bool("line_91", removed)

	if int(set.Count) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_92", int32(int(set.Count)))

	removed = set.Remove(42)
	if removed != false {
		t.Fatal("condition false")
	}

	r.Bool("line_96", removed)

	if int(set.Count) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_97", int32(int(set.Count)))

	set.Destroy()
}

func runTableHashSetClearTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetClearTest")

	set := New(16)

	set.Add(10)
	set.Add(20)
	set.Add(30)

	if int(set.Count) != 3 {
		t.Fatal("condition false")
	}

	r.I32("line_111", int32(int(set.Count)))

	set.Clear()

	if int(set.Count) != 0 {
		t.Fatal("condition false")
	}

	r.I32("line_115", int32(int(set.Count)))

	if set.Contains(10) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_116", set.Contains(10))

	if set.Contains(20) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_117", set.Contains(20))

	if set.Contains(30) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_118", set.Contains(30))

	set.Add(40)

	if int(set.Count) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_122", int32(int(set.Count)))

	if set.Contains(40) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_123", set.Contains(40))

	set.Destroy()
}

func runTableHashSetGrowthTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetGrowthTest")

	set := New(16)
	initialCapacity := int(set.Capacity)

	for i := range uint64(8) {
		set.Add(i + 1)
	}

	newCapacity := int(set.Capacity)
	if newCapacity < initialCapacity {
		t.Fatal("condition false")
	}

	r.I32("line_143", int32(newCapacity))
	r.I32("line_143_initial", int32(initialCapacity))

	if int(set.Count) != 8 {
		t.Fatal("condition false")
	}

	r.I32("line_144", int32(int(set.Count)))

	for i := uint64(1); i <= 8; i++ {
		if set.Contains(i) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_149", set.Contains(i))
	}

	set.Destroy()
}

func runTableHashSetEdgeCasesTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetEdgeCasesTest")

	set := New(16)

	largeKey := uint64(0xFFFFFFFFFFFFFFFF) - 1
	set.Add(largeKey)

	if set.Contains(largeKey) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_163", set.Contains(largeKey))

	if int(set.Count) != 1 {
		t.Fatal("condition false")
	}

	r.I32("line_164", int32(int(set.Count)))

	key1 := uint64(0x123456789ABCDEF)
	key2 := uint64(0x987654321FEDCBA)

	set.Add(key1)
	set.Add(key2)

	if set.Contains(key1) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_171", set.Contains(key1))

	if set.Contains(key2) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_172", set.Contains(key2))

	for i := uint64(0x1000); i < 0x1010; i++ {
		set.Add(i)
	}

	for i := uint64(0x1000); i < 0x1010; i++ {
		if set.Contains(i) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_182", set.Contains(i))
	}

	set.Destroy()
}

func runTableHashSetRemovalReorganizationTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetRemovalReorganizationTest")

	set := New(16)

	keys := [...]uint64{100, 116, 132, 148, 164}
	keyCount := len(keys)

	for i := range keyCount {
		set.Add(keys[i])
	}

	for i := range keyCount {
		if set.Contains(keys[i]) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_205", set.Contains(keys[i]))
	}

	set.Remove(keys[2])

	if set.Contains(keys[2]) != false {
		t.Fatal("condition false")
	}

	r.Bool("line_210", set.Contains(keys[2]))

	for i := range keyCount {
		if i != 2 {
			if set.Contains(keys[i]) != true {
				t.Fatal("condition false")
			}

			r.Bool("line_217", set.Contains(keys[i]))
		}
	}

	set.Destroy()
}

func runTableHashSetStressTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetStressTest")

	set := New(32)

	testSize := 1000

	var keys [1000]uint64

	for i := range testSize {
		keys[i] = uint64(i*7 + 13)
	}

	for i := range testSize {
		found := set.Add(keys[i])
		if found != false {
			t.Fatal("condition false")
		}

		r.Bool("line_244", found)
	}

	if int(set.Count) != testSize {
		t.Fatal("condition false")
	}

	r.I32("line_247", int32(int(set.Count)))

	for i := range testSize {
		if set.Contains(keys[i]) != true {
			t.Fatal("condition false")
		}

		r.Bool("line_252", set.Contains(keys[i]))
	}

	removedCount := 0

	for i := 0; i < testSize; i += 2 {
		removed := set.Remove(keys[i])
		if removed != true {
			t.Fatal("condition false")
		}

		r.Bool("line_260", removed)

		removedCount++
	}

	if int(set.Count) != testSize-removedCount {
		t.Fatal("condition false")
	}

	r.I32("line_264", int32(int(set.Count)))

	for i := range testSize {
		shouldBePresent := (i%2 == 1)
		if set.Contains(keys[i]) != shouldBePresent {
			t.Fatal("condition false")
		}

		r.Bool("line_270", set.Contains(keys[i]))
	}

	set.Destroy()
}

func runTableHashSetShapePairKeyTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetShapePairKeyTest")

	set := New(16)

	key1 := ShapePairKey(5, 10)

	key2 := ShapePairKey(10, 5)
	if key1 != key2 {
		t.Fatal("condition false")
	}

	r.U64("line_284", key1)
	r.U64("line_284_other", key2)

	set.Add(key1)

	if set.Contains(key1) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_287", set.Contains(key1))

	if set.Contains(key2) != true {
		t.Fatal("condition false")
	}

	r.Bool("line_288", set.Contains(key2))

	key3 := ShapePairKey(1, 2)

	key4 := ShapePairKey(2, 3)
	if key3 == key4 {
		t.Fatal("condition false")
	}

	r.U64("line_293", key3)
	r.U64("line_293_other", key4)

	set.Add(key3)
	set.Add(key4)

	if int(set.Count) != 3 {
		t.Fatal("condition false")
	}

	r.I32("line_297", int32(int(set.Count)))

	set.Destroy()
}

func runTableHashSetBytesTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetBytesTest")

	set := New(32)

	bytes := set.Bytes()

	expectedBytes := 32 * 8
	if bytes != expectedBytes {
		t.Fatal("condition false")
	}

	r.I32("line_309", int32(bytes))

	set.Add(100)
	set.Add(200)

	bytesAfterAdd := set.Bytes()
	if bytesAfterAdd != expectedBytes {
		t.Fatal("condition false")
	}

	r.I32("line_316", int32(bytesAfterAdd))

	set.Destroy()
}

func runTableHashSetTest(t *testing.T) {
	r := oracle.Open(t, "table/HashSetTest")

	var checks []byte

	n := 317
	itemCount := tableItemCount

	var removed [tableItemCount]bool

	for range 1 {
		set := New(16)

		for i := range n {
			for j := i + 1; j < n; j++ {
				key := ShapePairKey(i, j)

				found := set.Add(key)
				if found != false {
					t.Fatal("condition false")
				}

				checks = append(checks, tableBoolByte(found))
			}
		}

		r.Bytes("fill", checks)
		checks = checks[:0]

		if int(set.Count) != itemCount {
			t.Fatal("condition false")
		}

		r.I32("line_343", int32(int(set.Count)))

		k := 0
		removeCount := 0

		for i := range n {
			for j := i + 1; j < n; j++ {
				if j == i+1 {
					key := ShapePairKey(i, j)
					size1 := int(set.Count)

					found := set.Remove(key)
					if !found {
						t.Fatal("condition false")
					}

					r.Bool("line_357", found)

					size2 := int(set.Count)
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

		if int(set.Count) != (itemCount - removeCount) {
			t.Fatal("condition false")
		}

		r.I32("line_370", int32(int(set.Count)))

		ticks := timer.Ticks()

		k = 0

		for i := range n {
			for j := i + 1; j < n; j++ {
				key := ShapePairKey(j, i)

				found := set.Contains(key)
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

		ms := timer.Milliseconds(ticks)
		t.Logf(
			"set: count = %d, containsKey = %.5f ms, ave = %.5f us\n",
			itemCount,
			ms,
			1000.0*ms/float32(itemCount),
		)

		for i := range n {
			for j := i + 1; j < n; j++ {
				key := ShapePairKey(i, j)
				set.Remove(key)
			}
		}

		if int(set.Count) != 0 {
			t.Fatal("condition false")
		}

		r.I32("line_415", int32(int(set.Count)))

		set.Destroy()
	}
}
