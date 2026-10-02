// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// Is this a valid bounding box? Not Nan or infinity. Upper bound greater than or equal to lower bound.
func IsValidAABB(a AABB) bool {
	todo()
	return false
}

// Ray cast an AABB
func aabbRayCast(a AABB, p1 Vec2, p2 Vec2) CastOutput {
	todo()
	return CastOutput{}
}

// Get surface area of an AABB (the perimeter length)
func perimeter(a AABB) float32 {
	todo()
	return 0
}

// Enlarge a to contain b
// @return true if the AABB grew
func enlargeAABB(a *AABB, b AABB) bool {
	todo()
	return false
}

// Translate a relative AABB into world space, rounding outward so the float box always contains
// the true box far from the origin where a float coordinate cannot resolve the shape extent.
// Float ULP at 1e8 dwarfs the AABB margin, plain truncation could clip a shape out of its own
// box. The broadphase pair order rides on the deterministic directed rounding.
func offsetAABB(box AABB, origin Pos) AABB {
	todo()
	return AABB{}
}
