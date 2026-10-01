//go:build !box2d_double

package oracle

import "math"

type Real = float32

const (
	DoublePrecision = false

	config = "float"

	realWidth = 32
)

func realBits(v Real) uint64 { return uint64(math.Float32bits(v)) }
