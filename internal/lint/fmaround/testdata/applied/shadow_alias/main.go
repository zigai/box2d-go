package main

import (
	"fmt"
	"math"
)

type Real = float32
type Rounding = Real

//go:noinline
func result(a, b, c Real) Real {
	Real := 7
	if Real != 7 {
		panic("unrelated binding changed")
	}
	return a*b + c
}

func main() {
	a := Real(math.Float32frombits(0x3f800001))
	b := Real(math.Float32frombits(0x3f7ffffe))
	c := Real(math.Float32frombits(0xbf800000))
	fmt.Printf("result %08x\n", math.Float32bits(result(a, b, c)))
}
