package partitions

import units "measures"

type Real = float32
type Meters float32

func unaryPlus(a, b, c float32) {
	_ = +(a*b) + c // want `float32 product`
	_ = c - +(a*b) // want `float32 product`
	c += +(a*b) // want `float32 product`
}

func imported(a, b, c units.Value) {
	_ = a*b + c // want `float32 product`
	_ = units.Value(a*b) + c
}

func shadowFloat32(a, b, c float32) {
	float32 := 1
	_ = float32
	_ = a*b + c // want `float32 product`
}

func shadowAlias(a, b, c Real) {
	Real := 1
	_ = Real
	_ = a*b + c // want `float32 product`
}

func shadowNamed(a, b, c Meters) {
	Meters := 1
	_ = Meters
	_ = a*b + c // want `float32 product`
}
