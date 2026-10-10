package box2d

import (
	"math"
	"testing"
)

func TestAbsFloatBitEquivalence(t *testing.T) {
	check := func(bits uint32) {
		value := math.Float32frombits(bits)

		expected := value
		if value < 0 {
			expected = -value
		}

		if actual := math.Float32bits(absFloat(value)); actual != math.Float32bits(expected) {
			t.Fatalf("absFloat(%08x) = %08x, want %08x", bits, actual, math.Float32bits(expected))
		}
	}

	for _, magnitude := range []uint32{0, 1, 0x007fffff, 0x00800000, 0x3f800000, 0x7f7fffff, 0x7f800000, 0x7f800001, 0x7fbfffff, 0x7fc00000, 0x7fffffff} {
		check(magnitude)
		check(magnitude | 0x80000000)
	}

	for payload := uint32(1); payload <= 0x007fffff; payload++ {
		check(0x7f800000 | payload)
		check(0xff800000 | payload)
	}

	bits := uint32(12345)
	for range 1_000_000 {
		bits = 1664525*bits + 1013904223
		check(bits)
	}
}
