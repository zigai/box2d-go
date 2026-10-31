// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d-go/internal/oracle"
)

func TestVectorProductRounding(t *testing.T) {
	cases := []struct {
		name string
		a, b [2]uint32
		s    uint32
		want [2]uint32
	}{
		{
			"MulSV",
			[2]uint32{0x3f6ceba4, 0x3f5e42b3},
			[2]uint32{0xbf46a876, 0xbf2a855d},
			0xbf4fc718,
			[2]uint32{0x3fc7142e, 0x3fb45494},
		},
		{
			"CrossSV",
			[2]uint32{0xbf279f97, 0x3f47380a},
			[2]uint32{0xbf5ac7e1, 0x3f27bfcc},
			0xbf4af8bb,
			[2]uint32{0xbe0a7d10, 0x3fba572a},
		},
	}
	checks := oracle.Open(t, "vector_rounding")
	checks.I32("count", int32(len(cases)))

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := Vec2{math.Float32frombits(tc.a[0]), math.Float32frombits(tc.a[1])}
			b := Vec2{math.Float32frombits(tc.b[0]), math.Float32frombits(tc.b[1])}
			s := math.Float32frombits(tc.s)
			checks.Vec2(tc.name+"/a", a.X, a.Y)
			checks.Vec2(tc.name+"/b", b.X, b.Y)
			checks.F32(tc.name+"/s", s)

			var product, got Vec2
			if tc.name == "MulSV" {
				product = b.Scale(s)
				got = a.Add(b.Scale(s))
			} else {
				product = b.ScalarCross(s)
				got = a.Add(b.ScalarCross(s))
			}

			checks.Vec2(tc.name+"/product", product.X, product.Y)
			checks.Vec2(tc.name+"/result", got.X, got.Y)

			bits := [2]uint32{math.Float32bits(got.X), math.Float32bits(got.Y)}
			if bits != tc.want {
				t.Fatalf("got %08x, want %08x", bits, tc.want)
			}
		})
	}
}

func TestMathPublicScalars(t *testing.T) {
	r := oracle.Open(t, "math/PublicScalars")
	integers := []int{0, 1, -1, 17, -17, math.MaxInt32, math.MinInt32 + 1}
	r.I32("integers/count", int32(len(integers)))

	for i, a := range integers {
		b := integers[len(integers)-1-i]
		r.I32(fmt.Sprintf("integers/%d/min", i), int32(min(a, b)))
		r.I32(fmt.Sprintf("integers/%d/max", i), int32(max(a, b)))
		r.I32(fmt.Sprintf("integers/%d/abs", i), int32(absInt(a)))
		r.I32(fmt.Sprintf("integers/%d/clamp", i), int32(min(max(a, -1), 1)))
	}

	divisions := [][2]int{{0, 3}, {1, 3}, {3, 3}, {4, 3}, {17, 5}, {math.MaxInt32 - 2, 2}}
	r.I32("ceiling/count", int32(len(divisions)))

	for i, pair := range divisions {
		r.I32(fmt.Sprintf("ceiling/%d", i), int32(ceilingInt(pair[0], pair[1])))
	}

	negativeZero := math.Float32frombits(0x80000000)
	floats := []float32{0, negativeZero, -3.25, 3.25, -2, 2, 0x1p-126, 0x1p-149, math.MaxFloat32}
	r.I32("floats/count", int32(len(floats)))

	for i, a := range floats {
		b := floats[len(floats)-1-i]
		r.F32(fmt.Sprintf("floats/%d/min", i), minFloat(a, b))
		r.F32(fmt.Sprintf("floats/%d/max", i), maxFloat(a, b))
		r.F32(fmt.Sprintf("floats/%d/abs", i), absFloat(a))
		r.F32(fmt.Sprintf("floats/%d/clamp", i), clampFloat(a, -2, 2))
	}

	r.F32("zero/min-negative-first", minFloat(negativeZero, 0))
	r.F32("zero/min-positive-first", minFloat(0, negativeZero))
	r.F32("zero/max-negative-first", maxFloat(negativeZero, 0))
	r.F32("zero/max-positive-first", maxFloat(0, negativeZero))
}

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

	for _, magnitude := range []uint32{
		0, 1, 0x007fffff, 0x00800000, 0x3f800000, 0x7f7fffff,
		0x7f800000, 0x7f800001, 0x7fbfffff, 0x7fc00000, 0x7fffffff,
	} {
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
