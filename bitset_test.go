// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func TestBitset(t *testing.T) {

	r := oracle.Open(t, "bitset/BitSetTest")
	const count = 169
	bitSet := createBitSet(count)
	setBitCountAndClear(&bitSet, count)
	var values [count]bool
	i1, i2 := int32(0), int32(1)
	setBit(&bitSet, uint32(i1))
	values[i1] = true
	for i2 < count {
		setBit(&bitSet, uint32(i2))
		values[i2] = true
		next := i1 + i2
		i1 = i2
		i2 = next
	}
	for i := int32(0); i < count; i++ {
		value := getBit(&bitSet, uint32(i))
		if !(value == values[i]) {
			t.Fatal("condition false")
		}
		r.Bool(fmt.Sprintf("bit/%d", i), value)
	}
	destroyBitSet(&bitSet)
}
