// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package bitset

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func TestBitset(t *testing.T) {
	r := oracle.Open(t, "bitset/BitSetTest")

	const count = 169

	bitSet := New(count)
	bitSet.SetBitCountAndClear(count)

	var values [count]bool

	i1, i2 := int32(0), int32(1)
	bitSet.SetBit(uint32(i1))
	values[i1] = true

	for i2 < count {
		bitSet.SetBit(uint32(i2))
		values[i2] = true
		next := i1 + i2
		i1 = i2
		i2 = next
	}

	for i := range int32(count) {
		value := bitSet.GetBit(uint32(i))
		if value != values[i] {
			t.Fatal("condition false")
		}

		r.Bool(fmt.Sprintf("bit/%d", i), value)
	}

	bitSet.Destroy()
}
