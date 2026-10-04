package main

import (
	"fmt"
	"math"
)

type Rounding = float32

//go:noinline
func result(a, b, c float32) float32 {
	float32 := 7
	if float32 != 7 {
		panic("unrelated binding changed")
	}
	return a*b + c
}

func main() {
	a := math.Float32frombits(0x3f800001)
	b := math.Float32frombits(0x3f7ffffe)
	c := math.Float32frombits(0xbf800000)
	fmt.Printf("result %08x\n", math.Float32bits(result(a, b, c)))
}
