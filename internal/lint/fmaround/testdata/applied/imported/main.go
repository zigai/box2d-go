package main

import (
	"fmt"
	"math"
	units "roundingfixture/measures"
)

//go:noinline
func result(a, b, c units.Value) units.Value {
	return a*b + c
}

func main() {
	a := units.Value(math.Float32frombits(0x3f800001))
	b := units.Value(math.Float32frombits(0x3f7ffffe))
	c := units.Value(math.Float32frombits(0xbf800000))
	fmt.Printf("result %08x\n", math.Float32bits(float32(result(a, b, c))))
}
