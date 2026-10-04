// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"math"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func TestVectorProductRounding(t *testing.T) {

	cases := []struct {
		name string
		a, b [2]uint32
		s    uint32
		want [2]uint32
	}{
		{"MulSV", [2]uint32{0x3f6ceba4, 0x3f5e42b3}, [2]uint32{0xbf46a876, 0xbf2a855d}, 0xbf4fc718, [2]uint32{0x3fc7142e, 0x3fb45494}},
		{"CrossSV", [2]uint32{0xbf279f97, 0x3f47380a}, [2]uint32{0xbf5ac7e1, 0x3f27bfcc}, 0xbf4af8bb, [2]uint32{0xbe0a7d10, 0x3fba572a}},
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
				product = MulSV(s, b)
				got = Add(a, MulSV(s, b))
			} else {
				product = CrossSV(s, b)
				got = Add(a, CrossSV(s, b))
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
