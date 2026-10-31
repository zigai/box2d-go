// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"math"

	"github.com/zigai/box2d-go/internal/core"
)

// Compute the convex hull of a set of points. Returns an empty hull if it fails.
// Some failure cases:
// - all points very close together
// - all points on a line
// - less than 3 points
// - more than MaxPolygonVertices points
// This welds close points and removes collinear points.
// @warning Do not modify a hull once it has been computed.
func ComputeHull(points []Vec2) Hull {
	count := len(points)

	var hull Hull
	if count < 3 || count > MaxPolygonVertices {
		return hull
	}

	count = min(count, MaxPolygonVertices)

	var ps [MaxPolygonVertices]Vec2

	linearSlop := LinearSlop()
	aabb, n := weldHullPoints(points[:count], &ps, linearSlop)

	if n < 3 {
		return hull
	}

	f1 := farthestHullPoint(ps[:n], aabb.Center())

	p1 := ps[f1]
	ps[f1] = ps[n-1]
	n--
	f2 := farthestHullPoint(ps[:n], p1)

	p2 := ps[f2]
	ps[f2] = ps[n-1]
	n--

	var rightPoints [MaxPolygonVertices - 2]Vec2

	rightCount := 0

	var leftPoints [MaxPolygonVertices - 2]Vec2

	leftCount := 0

	e := p2.Sub(p1).Normalize()
	for i := range n {
		d := ps[i].Sub(p1).Cross(e)
		if d >= 2*linearSlop {
			rightPoints[rightCount] = ps[i]
			rightCount++
		} else if d <= -2*linearSlop {
			leftPoints[leftCount] = ps[i]
			leftCount++
		}
	}

	hull1 := recurseHull(p1, p2, rightPoints[:], rightCount)

	hull2 := recurseHull(p2, p1, leftPoints[:], leftCount)
	if hull1.Count == 0 && hull2.Count == 0 {
		return hull
	}

	hull.Points[hull.Count] = p1

	hull.Count++
	for i := range hull1.Count {
		hull.Points[hull.Count] = hull1.Points[i]
		hull.Count++
	}

	hull.Points[hull.Count] = p2

	hull.Count++
	for i := range hull2.Count {
		hull.Points[hull.Count] = hull2.Points[i]
		hull.Count++
	}

	core.Assert(hull.Count <= MaxPolygonVertices, "hull.count <= B2_MAX_POLYGON_VERTICES")

	hull.removeCollinearPoints(linearSlop)

	return hull
}

// This determines if a hull is valid. Checks for:
// - convexity
// - collinear points
// This is expensive and should not be called at runtime.
func ValidateHull(hull *Hull) bool {
	if hull.Count < 3 || MaxPolygonVertices < hull.Count {
		return false
	}

	for i := range hull.Count {
		i1 := i

		i2 := 0
		if i < hull.Count-1 {
			i2 = i1 + 1
		}

		p := hull.Points[i1]

		e := hull.Points[i2].Sub(p).Normalize()
		for j := range hull.Count {
			if j == i1 || j == i2 {
				continue
			}

			distance := hull.Points[j].Sub(p).Cross(e)
			if distance >= 0 {
				return false
			}
		}
	}

	linearSlop := LinearSlop()

	for i := range hull.Count {
		i1 := i
		i2 := (i + 1) % hull.Count
		i3 := (i + 2) % hull.Count
		p1 := hull.Points[i1]
		p2 := hull.Points[i2]
		p3 := hull.Points[i3]
		e := p3.Sub(p1).Normalize()

		distance := p2.Sub(p1).Cross(e)
		if distance <= linearSlop {
			return false
		}
	}

	return true
}

func recurseHull(p1 Vec2, p2 Vec2, ps []Vec2, count int) Hull {
	var hull Hull
	if count == 0 {
		return hull
	}

	e := p2.Sub(p1).Normalize()

	var rightPoints [MaxPolygonVertices]Vec2

	rightCount := 0
	bestIndex := 0

	bestDistance := ps[bestIndex].Sub(p1).Cross(e)
	if bestDistance > 0 {
		rightPoints[rightCount] = ps[bestIndex]
		rightCount++
	}

	for i := 1; i < count; i++ {
		distance := ps[i].Sub(p1).Cross(e)
		if distance > bestDistance {
			bestIndex = i
			bestDistance = distance
		}

		if distance > 0 {
			rightPoints[rightCount] = ps[i]
			rightCount++
		}
	}

	if bestDistance < 2*LinearSlop() {
		return hull
	}

	bestPoint := ps[bestIndex]
	hull1 := recurseHull(p1, bestPoint, rightPoints[:], rightCount)
	hull2 := recurseHull(bestPoint, p2, rightPoints[:], rightCount)

	for i := range hull1.Count {
		hull.Points[hull.Count] = hull1.Points[i]
		hull.Count++
	}

	hull.Points[hull.Count] = bestPoint

	hull.Count++
	for i := range hull2.Count {
		hull.Points[hull.Count] = hull2.Points[i]
		hull.Count++
	}

	core.Assert(hull.Count < MaxPolygonVertices, "hull.count < B2_MAX_POLYGON_VERTICES")

	return hull
}

func weldHullPoints(points []Vec2, ps *[MaxPolygonVertices]Vec2, linearSlop float32) (AABB, int) {
	aabb := AABB{Vec2{math.MaxFloat32, math.MaxFloat32}, Vec2{-math.MaxFloat32, -math.MaxFloat32}}
	n := 0
	tolSqr := 16 * linearSlop * linearSlop

	for i := range points {
		aabb.LowerBound = aabb.LowerBound.Min(points[i])
		aabb.UpperBound = aabb.UpperBound.Max(points[i])
		vi := points[i]
		unique := true

		for j := range i {
			vj := points[j]

			distSqr := vi.DistanceSquared(vj)
			if distSqr < tolSqr {
				unique = false
				break
			}
		}

		if unique {
			ps[n] = vi
			n++
		}
	}

	return aabb, n
}

func farthestHullPoint(points []Vec2, origin Vec2) int {
	index := 0

	distanceSquared := origin.DistanceSquared(points[index])
	for i := 1; i < len(points); i++ {
		dsq := origin.DistanceSquared(points[i])
		if dsq > distanceSquared {
			index = i
			distanceSquared = dsq
		}
	}

	return index
}

func (hull *Hull) removeCollinearPoints(linearSlop float32) {
	searching := true
	for searching && hull.Count > 2 {
		searching = false

		for i := 0; i < hull.Count; i++ {
			i1 := i
			i2 := (i + 1) % hull.Count
			i3 := (i + 2) % hull.Count
			s1 := hull.Points[i1]
			s2 := hull.Points[i2]
			s3 := hull.Points[i3]
			r := s3.Sub(s1).Normalize()

			distance := s2.Sub(s1).Cross(r)
			if distance <= 2*linearSlop {
				for j := i2; j < hull.Count-1; j++ {
					hull.Points[j] = hull.Points[j+1]
				}

				hull.Count--
				searching = true

				break
			}
		}
	}

	if hull.Count < 3 {
		hull.Count = 0
	}
}
