package main

import (
	"fmt"
	"math"
	. "roundingfixture/measures"
)

//go:noinline
func result(a, b, c Value) Value {
	return a*b + c
}

func main() {
	a := Value(math.Float32frombits(0x3f800001))
	b := Value(math.Float32frombits(0x3f7ffffe))
	c := Value(math.Float32frombits(0xbf800000))
	fmt.Printf("result %08x\n", math.Float32bits(float32(result(a, b, c))))
}
