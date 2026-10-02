// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

func recurseHull(p1 Vec2, p2 Vec2, ps []Vec2, count int) Hull {
	var hull Hull
	if count == 0 {
		return hull
	}
	e := Normalize(Sub(p2, p1))
	var rightPoints [MaxPolygonVertices]Vec2
	rightCount := 0
	bestIndex := 0
	bestDistance := Cross(Sub(ps[bestIndex], p1), e)
	if bestDistance > 0 {
		rightPoints[rightCount] = ps[bestIndex]
		rightCount++
	}
	for i := 1; i < count; i++ {
		distance := Cross(Sub(ps[i], p1), e)
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
	for i := 0; i < hull1.Count; i++ {
		hull.Points[hull.Count] = hull1.Points[i]
		hull.Count++
	}
	hull.Points[hull.Count] = bestPoint
	hull.Count++
	for i := 0; i < hull2.Count; i++ {
		hull.Points[hull.Count] = hull2.Points[i]
		hull.Count++
	}
	assert(hull.Count < MaxPolygonVertices, "hull.count < B2_MAX_POLYGON_VERTICES")
	return hull
}

// Compute the convex hull of a set of points. Returns an empty hull if it fails.
// Some failure cases:
// - all points very close together
// - all points on a line
// - less than 3 points
// - more than MaxPolygonVertices points
// This welds close points and removes collinear points.
// @warning Do not modify a hull once it has been computed
func ComputeHull(points []Vec2, count int) Hull {
	var hull Hull
	if count < 3 || count > MaxPolygonVertices {
		return hull
	}
	count = MinInt(count, MaxPolygonVertices)
	aabb := AABB{Vec2{math.MaxFloat32, math.MaxFloat32}, Vec2{-math.MaxFloat32, -math.MaxFloat32}}
	var ps [MaxPolygonVertices]Vec2
	n := 0
	linearSlop := LinearSlop()
	tolSqr := 16 * linearSlop * linearSlop
	for i := 0; i < count; i++ {
		aabb.LowerBound = Min(aabb.LowerBound, points[i])
		aabb.UpperBound = Max(aabb.UpperBound, points[i])
		vi := points[i]
		unique := true
		for j := 0; j < i; j++ {
			vj := points[j]
			distSqr := DistanceSquared(vi, vj)
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
	if n < 3 {
		return hull
	}
	c := aabb.Center()
	f1 := 0
	dsq1 := DistanceSquared(c, ps[f1])
	for i := 1; i < n; i++ {
		dsq := DistanceSquared(c, ps[i])
		if dsq > dsq1 {
			f1 = i
			dsq1 = dsq
		}
	}
	p1 := ps[f1]
	ps[f1] = ps[n-1]
	n--
	f2 := 0
	dsq2 := DistanceSquared(p1, ps[f2])
	for i := 1; i < n; i++ {
		dsq := DistanceSquared(p1, ps[i])
		if dsq > dsq2 {
			f2 = i
			dsq2 = dsq
		}
	}
	p2 := ps[f2]
	ps[f2] = ps[n-1]
	n--
	var rightPoints [MaxPolygonVertices - 2]Vec2
	rightCount := 0
	var leftPoints [MaxPolygonVertices - 2]Vec2
	leftCount := 0
	e := Normalize(Sub(p2, p1))
	for i := 0; i < n; i++ {
		d := Cross(Sub(ps[i], p1), e)
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
	for i := 0; i < hull1.Count; i++ {
		hull.Points[hull.Count] = hull1.Points[i]
		hull.Count++
	}
	hull.Points[hull.Count] = p2
	hull.Count++
	for i := 0; i < hull2.Count; i++ {
		hull.Points[hull.Count] = hull2.Points[i]
		hull.Count++
	}
	assert(hull.Count <= MaxPolygonVertices, "hull.count <= B2_MAX_POLYGON_VERTICES")
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
			r := Normalize(Sub(s3, s1))
			distance := Cross(Sub(s2, s1), r)
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
	for i := 0; i < hull.Count; i++ {
		i1 := i
		i2 := 0
		if i < hull.Count-1 {
			i2 = i1 + 1
		}
		p := hull.Points[i1]
		e := Normalize(Sub(hull.Points[i2], p))
		for j := 0; j < hull.Count; j++ {
			if j == i1 || j == i2 {
				continue
			}
			distance := Cross(Sub(hull.Points[j], p), e)
			if distance >= 0 {
				return false
			}
		}
	}
	linearSlop := LinearSlop()
	for i := 0; i < hull.Count; i++ {
		i1 := i
		i2 := (i + 1) % hull.Count
		i3 := (i + 2) % hull.Count
		p1 := hull.Points[i1]
		p2 := hull.Points[i2]
		p3 := hull.Points[i3]
		e := Normalize(Sub(p3, p1))
		distance := Cross(Sub(p2, p1), e)
		if distance <= linearSlop {
			return false
		}
	}
	return true
}
