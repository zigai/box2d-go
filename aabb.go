// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

type aabbRayInterval struct {
	minimum float32
	maximum float32
}

// Is this a valid bounding box? Not Nan or infinity. Upper bound greater than or equal to lower bound.
func (a AABB) IsValid() bool {
	d := a.UpperBound.Sub(a.LowerBound)
	valid := d.X >= 0 && d.Y >= 0
	valid = valid && a.LowerBound.IsValid() && a.UpperBound.IsValid()

	return valid
}

// Ray cast an AABB.
func aabbRayCast(a AABB, p1 Vec2, p2 Vec2) CastOutput {
	var output CastOutput

	interval := aabbRayInterval{-math.MaxFloat32, math.MaxFloat32}
	d := p2.Sub(p1)

	var normal Vec2

	sign, changed, hit := interval.clip(a.LowerBound.X, a.UpperBound.X, p1.X, d.X)
	if !hit {
		return output
	}

	if changed {
		normal.Y = 0
		normal.X = sign
	}

	sign, changed, hit = interval.clip(a.LowerBound.Y, a.UpperBound.Y, p1.Y, d.Y)
	if !hit {
		return output
	}

	if changed {
		normal.X = 0
		normal.Y = sign
	}

	tMin := interval.minimum

	if tMin < 0 {
		return output
	}

	if 1 < tMin {
		return output
	}

	output.Fraction = tMin
	output.Normal = normal
	output.Point = p1.Lerp(p2, tMin)
	output.Hit = true

	return output
}

func (interval *aabbRayInterval) clip(lower, upper, origin, direction float32) (float32, bool, bool) {
	var changed bool

	if absFloat(direction) < float32(0x1p-23) {
		return 0, false, !(origin < lower || upper < origin)
	}

	invD := 1 / direction
	t1 := (lower - origin) * invD
	t2 := (upper - origin) * invD
	sign := float32(-1)

	if t1 > t2 {
		t1, t2 = t2, t1
		sign = 1
	}

	if t1 > interval.minimum {
		interval.minimum = t1
		changed = true
	}

	interval.maximum = minFloat(interval.maximum, t2)

	return sign, changed, !(interval.minimum > interval.maximum)
}

// Get surface area of an AABB (the perimeter length).
func perimeter(a AABB) float32 {
	wx := a.UpperBound.X - a.LowerBound.X
	wy := a.UpperBound.Y - a.LowerBound.Y

	return 2 * (wx + wy)
}

// Enlarge a to contain b
// @return true if the AABB grew.
func enlargeAABB(a *AABB, b AABB) bool {
	changed := false

	if b.LowerBound.X < a.LowerBound.X {
		a.LowerBound.X = b.LowerBound.X
		changed = true
	}

	if b.LowerBound.Y < a.LowerBound.Y {
		a.LowerBound.Y = b.LowerBound.Y
		changed = true
	}

	if a.UpperBound.X < b.UpperBound.X {
		a.UpperBound.X = b.UpperBound.X
		changed = true
	}

	if a.UpperBound.Y < b.UpperBound.Y {
		a.UpperBound.Y = b.UpperBound.Y
		changed = true
	}

	return changed
}

// Translate a relative AABB into world space, rounding outward so the float box always contains
// the true box far from the origin where a float coordinate cannot resolve the shape extent.
// Float ULP at 1e8 dwarfs the AABB margin, plain truncation could clip a shape out of its own
// box. The broadphase pair order rides on the deterministic directed rounding.
func offsetAABB(box AABB, origin Pos) AABB {
	var result AABB

	result.LowerBound.X = RoundDownFloat(float64(origin.X) + float64(box.LowerBound.X))
	result.LowerBound.Y = RoundDownFloat(float64(origin.Y) + float64(box.LowerBound.Y))
	result.UpperBound.X = RoundUpFloat(float64(origin.X) + float64(box.UpperBound.X))
	result.UpperBound.Y = RoundUpFloat(float64(origin.Y) + float64(box.UpperBound.Y))

	return result
}
