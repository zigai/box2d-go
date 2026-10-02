// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

func recurseHull(p1 Vec2, p2 Vec2, ps []Vec2, count int) Hull {
	todo()
	return Hull{}
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
	todo()
	return Hull{}
}

// This determines if a hull is valid. Checks for:
// - convexity
// - collinear points
// This is expensive and should not be called at runtime.
func ValidateHull(hull *Hull) bool {
	todo()
	return false
}
