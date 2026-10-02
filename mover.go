// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// These are the collision planes returned from World_CollideMover.
// The plane and point are relative to the query origin, matching the mover capsule.
type PlaneResult struct {
	// The collision plane between the mover and a convex shape
	Plane Plane
	Point Vec2
	// Did the collision register a hit? If not this plane should be ignored.
	Hit bool
}

// These are collision planes that can be fed to SolvePlanes. Normally
// this is assembled by the user from plane results in PlaneResult.
type CollisionPlane struct {
	// The collision plane between the mover and some shape.
	Plane Plane
	// Setting this to FLT_MAX makes the plane as rigid as possible. Lower values can
	// make the plane collision soft. Usually in meters.
	PushLimit float32
	// The push on the mover determined by SolvePlanes. Usually in meters.
	Push float32
	// Indicates if ClipVector should clip against this plane. Should be false for soft collision.
	ClipVelocity bool
}

// Result returned by SolvePlanes.
type PlaneSolverResult struct {
	// The final relative translation.
	Delta Vec2
	// The number of iterations used by the plane solver. For diagnostics.
	IterationCount int
}

// Solves the position of a mover that satisfies the given collision planes.
// @param targetDelta the desired translation from the position used to generate the collision planes
// @param planes the collision planes
func SolvePlanes(targetDelta Vec2, planes []CollisionPlane) PlaneSolverResult {
	for i := range planes {
		planes[i].Push = 0
	}
	delta := targetDelta
	tolerance := LinearSlop()
	iteration := 0
	for ; iteration < 20; iteration++ {
		var totalPush float32
		for planeIndex := range planes {
			plane := &planes[planeIndex]
			separation := PlaneSeparation(plane.Plane, delta) + LinearSlop()
			push := -separation
			accumulatedPush := plane.Push
			plane.Push = ClampFloat(plane.Push+push, 0, plane.PushLimit)
			push = plane.Push - accumulatedPush
			delta = MulAdd(delta, push, plane.Plane.Normal)
			totalPush += AbsFloat(push)
		}
		if totalPush < tolerance {
			break
		}
	}
	return PlaneSolverResult{Delta: delta, IterationCount: iteration}
}

// Clips the velocity against the given collision planes. Planes with zero push or ClipVelocity
// set to false are skipped.
func ClipVector(vector Vec2, planes []CollisionPlane) Vec2 {
	v := vector
	for planeIndex := range planes {
		plane := &planes[planeIndex]
		if plane.Push == 0 || !plane.ClipVelocity {
			continue
		}
		v = MulSub(v, MinFloat(0, Dot(v, plane.Plane.Normal)), plane.Plane.Normal)
	}
	return v
}
