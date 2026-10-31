// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d-go/internal/oracle"
)

func geometryCheckAABBCast(r *oracle.Reader, output CastOutput) {
	r.Bool("full/cast/hit", output.Hit)
	r.I32("full/cast/iterations", int32(output.Iterations))

	if output.Hit {
		r.F32("full/cast/fraction", output.Fraction)
		r.Vec2("full/cast/normal", output.Normal.X, output.Normal.Y)
		r.Vec2("full/cast/point", output.Point.X, output.Point.Y)
	}
}

func geometryCheckManifold(r *oracle.Reader, key string, m LocalManifold) {
	r.I32(key+"/count", int32(m.PointCount))
	// The normal and dead point slots have no contact meaning when count is zero.
	if m.PointCount > 0 {
		r.Vec2(key+"/normal", m.Normal.X, m.Normal.Y)
	}

	for i := range m.PointCount {
		prefix := fmt.Sprintf("%s/%d", key, i)
		r.Vec2(prefix+"/point", m.Points[i].Point.X, m.Points[i].Point.Y)
		r.F32(prefix+"/separation", m.Points[i].Separation)
		r.U32(prefix+"/id", uint32(m.Points[i].ID))
	}
}

func geometryCheckCache(r *oracle.Reader, key string, cache SimplexCache) {
	r.U32(key+"/count", uint32(cache.Count))

	for i := range int(cache.Count) {
		r.U32(fmt.Sprintf("%s/%d/a", key, i), uint32(cache.IndexA[i]))
		r.U32(fmt.Sprintf("%s/%d/b", key, i), uint32(cache.IndexB[i]))
	}
}

func TestCollisionPublicPairs(t *testing.T) {
	r := oracle.Open(t, "collision/PublicPairs")
	circle := Circle{Vec2{0, 0}, 0.5}
	capsule := Capsule{Vec2{-0.75, 0}, Vec2{0.75, 0}, 0.25}
	segment := Segment{Vec2{-1.5, 0}, Vec2{1.5, 0}}
	polygon := MakeRoundedBox(0.75, 0.5, 0.1)
	// Straight, convex and concave ghost joins exercise front/back and both endpoint seams.
	chains := []ChainSegment{
		{Vec2{-3, 0}, segment, Vec2{3, 0}, 0},
		{Vec2{-2.5, -1}, segment, Vec2{2.5, -1}, 0},
		{Vec2{-2.5, 1}, segment, Vec2{2.5, 1}, 0},
	}
	positions := []Vec2{
		{0, 3},
		{0, 1.21},
		{0, 1.2},
		{0, 1.01},
		{0, 1},
		{0, 0.86},
		{0, 0.85},
		{0, 0.76},
		{0, 0.75},
		{0, 0.61},
		{0, 0.6},
		{0, 0.51},
		{0, 0.5},
		{0, 0.26},
		{0, 0.25},
		{0, 0},
		{0, -0.25},
		{0, -0.26},
		{0, -0.5},
		{0, -0.51},
		{0, -0.6},
		{0, -0.61},
		{0, -1},
		{0, -3},
		{-1.6, -0.2},
		{-1.5, -0.2},
		{-1.4, -0.2},
		{1.4, -0.2},
		{1.5, -0.2},
		{1.6, -0.2},
		{-1.5, 0.2},
		{1.5, 0.2},
		{0.25, -0.45},
	}

	angles := []float32{0, 0.35, -0.65}
	for a, angle := range angles {
		var capsuleCache, polygonCache [3]SimplexCache

		for p, position := range positions {
			xf := Transform{position, MakeRot(angle)}
			key := fmt.Sprintf("%d/%d/", a, p)
			geometryCheckManifold(r, key+"circles", CollideCircles(&circle, &circle, xf))
			geometryCheckManifold(r, key+"capsule-circle", CollideCapsuleAndCircle(&capsule, &circle, xf))
			geometryCheckManifold(r, key+"segment-circle", CollideSegmentAndCircle(&segment, &circle, xf))
			geometryCheckManifold(r, key+"polygon-circle", CollidePolygonAndCircle(&polygon, &circle, xf))
			geometryCheckManifold(r, key+"capsules", CollideCapsules(&capsule, &capsule, xf))
			geometryCheckManifold(r, key+"segment-capsule", CollideSegmentAndCapsule(&segment, &capsule, xf))
			geometryCheckManifold(r, key+"polygon-capsule", CollidePolygonAndCapsule(&polygon, &capsule, xf))
			geometryCheckManifold(r, key+"polygons", CollidePolygons(&polygon, &polygon, xf))
			geometryCheckManifold(r, key+"segment-polygon", CollideSegmentAndPolygon(&segment, &polygon, xf))

			for c := range chains {
				prefix := fmt.Sprintf("%schain%d", key, c)
				geometryCheckManifold(r, prefix+"-circle", CollideChainSegmentAndCircle(&chains[c], &circle, xf))
				geometryCheckManifold(r, prefix+"-capsule", CollideChainSegmentAndCapsule(&chains[c], &capsule, xf, &capsuleCache[c]))
				geometryCheckCache(r, prefix+"-capsule-cache", capsuleCache[c])
				geometryCheckManifold(r, prefix+"-polygon", CollideChainSegmentAndPolygon(&chains[c], &polygon, xf, &polygonCache[c]))
				geometryCheckCache(r, prefix+"-polygon-cache", polygonCache[c])
			}
		}
	}
}
