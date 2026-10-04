package main

import (
	"fmt"
	"math"
)

type Real = float32
type Meters float32
type vector struct{ X float32 }

// The exact product of these inputs is 1 - 2^-46. Rounded to
// float32 it is 1; a fused multiply-add with -1 is -2^-46.
//
//go:noinline
func exercise(a, b, c float32) {
	fmt.Printf("add %08x\n", math.Float32bits(a*b+c))
	fmt.Printf("right %08x\n", math.Float32bits(c+a*b))
	fmt.Printf("subtract %08x\n", math.Float32bits(a*b-(-c)))
	fmt.Printf("right_subtract %08x\n", math.Float32bits((-c)-a*b))
	fmt.Printf("unary_plus %08x\n", math.Float32bits(+(a*b)+c))
	fmt.Printf("unary_minus %08x\n", math.Float32bits(-(a*b)-c))
	fmt.Printf("negative_zero %08x\n", math.Float32bits(-(a*b + c)))
	fmt.Printf("grouping %08x\n", math.Float32bits(a*(1+(b-1))+c))
	x := c
	x += a * b
	fmt.Printf("compound_add %08x\n", math.Float32bits(x))
	x = -c
	x -= a * b
	fmt.Printf("compound_subtract %08x\n", math.Float32bits(x))
	v := vector{X: c}
	v.X += a * b
	fmt.Printf("field %08x\n", math.Float32bits(v.X))
	s := []float32{c}
	s[0] += a * b
	fmt.Printf("index %08x\n", math.Float32bits(s[0]))
	fmt.Printf("both_products %08x\n", math.Float32bits(a*b+(-a)*b))

	var ra, rb, rc Real = a, b, c
	fmt.Printf("alias %08x\n", math.Float32bits(ra*rb+rc))
	ma, mb, mc := Meters(a), Meters(b), Meters(c)
	fmt.Printf("named %08x\n", math.Float32bits(float32(ma*mb+mc)))
}

// These nonzero outcomes independently discriminate grouping and unary signs.
//
//go:noinline
func grouping(a, b, c float32) {
	fmt.Printf("nonzero_grouping %08x\n", math.Float32bits(a*(b+c)-c))
	fmt.Printf("nonzero_outer_minus %08x\n", math.Float32bits(-(a*b - c)))
	fmt.Printf("nonzero_nested_sign %08x\n", math.Float32bits(+(-a*b)+c))
}

func main() {
	a := math.Float32frombits(0x3f800001)
	b := math.Float32frombits(0x3f7ffffe)
	c := math.Float32frombits(0xbf800000)
	exercise(a, b, c)
	grouping(2, 3, 5)
	fmt.Printf("fused %08x\n", math.Float32bits(float32(math.FMA(float64(a), float64(b), float64(c)))))
}
