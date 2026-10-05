// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func geometryOriginalDistance(r *oracle.Reader, output box2d.DistanceOutput, cache box2d.SimplexCache) {
	r.F32("full/distance", output.Distance)
	r.Vec2("full/point-a", output.PointA.X, output.PointA.Y)
	r.Vec2("full/point-b", output.PointB.X, output.PointB.Y)

	if output.Distance > 0 {
		r.Vec2("full/normal", output.Normal.X, output.Normal.Y)
	}

	r.I32("full/iterations", int32(output.Iterations))
	r.I32("full/simplex-count", int32(output.SimplexCount))
	r.U32("full/cache/count", uint32(cache.Count))

	for i := range int(cache.Count) {
		r.U32(fmt.Sprintf("full/cache/%d/a", i), uint32(cache.IndexA[i]))
		r.U32(fmt.Sprintf("full/cache/%d/b", i), uint32(cache.IndexB[i]))
	}
}

func geometryOriginalCast(r *oracle.Reader, output box2d.CastOutput) {
	r.Bool("full/cast/hit", output.Hit)
	r.I32("full/cast/iterations", int32(output.Iterations))

	if output.Hit {
		r.F32("full/cast/fraction", output.Fraction)
		r.Vec2("full/cast/normal", output.Normal.X, output.Normal.Y)
		r.Vec2("full/cast/point", output.Point.X, output.Point.Y)
	}
}

func geometryOriginalMass(r *oracle.Reader, md box2d.MassData) {
	r.F32("full/mass", md.Mass)
	r.Vec2("full/mass-center", md.Center.X, md.Center.Y)
	r.F32("full/inertia", md.RotationalInertia)
}

func geometryOriginalHull(r *oracle.Reader, hull box2d.Hull) {
	r.I32("full/hull/count", int32(hull.Count))
	r.Bool("full/hull/valid", box2d.ValidateHull(&hull))

	for i := range hull.Count {
		r.Vec2(fmt.Sprintf("full/hull/%d", i), hull.Points[i].X, hull.Points[i].Y)
	}
}

func geometryOriginalPolygon(r *oracle.Reader, polygon box2d.Polygon) {
	r.I32("full/polygon/count", int32(polygon.Count))
	r.Vec2("full/polygon/centroid", polygon.Centroid.X, polygon.Centroid.Y)
	r.F32("full/polygon/radius", polygon.Radius)

	for i := range polygon.Count {
		r.Vec2(fmt.Sprintf("full/polygon/%d/vertex", i), polygon.Vertices[i].X, polygon.Vertices[i].Y)
		r.Vec2(fmt.Sprintf("full/polygon/%d/normal", i), polygon.Normals[i].X, polygon.Normals[i].Y)
	}
}
