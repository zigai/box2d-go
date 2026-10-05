//go:build box2d_double

package oracle

import "math"

const (
	DoublePrecision = true

	config = "double"

	realWidth = 64
)

type Real = float64

func realBits(v Real) uint64 { return math.Float64bits(v) }
