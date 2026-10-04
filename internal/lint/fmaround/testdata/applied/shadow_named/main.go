package main

import (
	"fmt"
	"math"
)

type Meters float32
type Rounding = Meters

//go:noinline
func result(a, b, c Meters) Meters {
	Meters := 7
	if Meters != 7 {
		panic("unrelated binding changed")
	}
	return a*b + c
}

func main() {
	a := Meters(math.Float32frombits(0x3f800001))
	b := Meters(math.Float32frombits(0x3f7ffffe))
	c := Meters(math.Float32frombits(0xbf800000))
	fmt.Printf("result %08x\n", math.Float32bits(float32(result(a, b, c))))
}
