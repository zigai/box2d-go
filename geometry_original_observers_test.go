// SPDX-License-Identifier: MIT

package box2d_test

import (
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func geometryOriginalCast(r *oracle.Reader, output box2d.CastOutput) {
	r.Bool("full/cast/hit", output.Hit)
	r.I32("full/cast/iterations", int32(output.Iterations))

	if output.Hit {
		r.F32("full/cast/fraction", output.Fraction)
		r.Vec2("full/cast/normal", output.Normal.X, output.Normal.Y)
		r.Vec2("full/cast/point", output.Point.X, output.Point.Y)
	}
}
