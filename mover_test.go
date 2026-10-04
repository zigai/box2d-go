// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"

	"math"
	"testing"
)

func collisionMakePlane(normal box2d.Vec2, pushLimit float32) box2d.CollisionPlane {
	return box2d.CollisionPlane{Plane: box2d.Plane{normal, 0}, PushLimit: pushLimit, Push: 0, ClipVelocity: true}
}

func TestMover(t *testing.T) {

	t.Run("NoPlanesTest", func(t *testing.T) {

		r := oracle.Open(t, "mover/NoPlanesTest")
		targetDelta := box2d.Vec2{0.5, -0.25}
		result := box2d.SolvePlanes(targetDelta, nil)
		geometryMoverObserve(r, result, nil)

		r.F32("mover:27/value0", result.Delta.X)
		r.F32("mover:27/value1", targetDelta.X)
		r.Bool("mover:27", result.Delta.X == targetDelta.X)

		if !(result.Delta.X == targetDelta.X) {
			t.Fatal("condition false")
		}
		r.F32("mover:29/value0", result.Delta.Y)
		r.F32("mover:29/value1", targetDelta.Y)
		r.Bool("mover:29", result.Delta.Y == targetDelta.Y)
		if !(result.Delta.Y == targetDelta.Y) {
			t.Fatal("condition false")
		}
		r.I32("mover:30/value0", int32(result.IterationCount))
		r.Bool("mover:30", result.IterationCount == 0)
		if !(result.IterationCount == 0) {
			t.Fatal("condition false")
		}

	})
	t.Run("SolvePlaneTest", func(t *testing.T) {

		r := oracle.Open(t, "mover/SolvePlaneTest")
		planes := []box2d.CollisionPlane{collisionMakePlane(box2d.Vec2{0.0, 1.0}, float32(math.MaxFloat32))}
		result := box2d.SolvePlanes(box2d.Vec2{0.0, -1.0}, planes)
		geometryMoverObserve(r, result, planes)

		// An overshoot would eject the mover out the far side of the floor
		r.F32("mover:43/value0", result.Delta.Y)
		r.Bool("mover:43", result.Delta.Y <= 0.0)
		if !(result.Delta.Y <= 0.0) {
			t.Fatal("condition false")
		}

		// The mover comes to rest a slop under the surface, not through it
		r.F32("mover:46/value0", result.Delta.Y)
		r.Bool("mover:46", result.Delta.Y > -10.0*box2d.LinearSlop())
		if !(result.Delta.Y > -10.0*box2d.LinearSlop()) {
			t.Fatal("condition false")
		}
		r.F32("mover:47/value0", result.Delta.X)
		r.Bool("mover:47", result.Delta.X == 0.0)
		if !(result.Delta.X == 0.0) {
			t.Fatal("condition false")
		}
		r.F32("mover:48/value0", planes[0].Push)
		r.Bool("mover:48", planes[0].Push > 0.0)
		if !(planes[0].Push > 0.0) {
			t.Fatal("condition false")
		}

	})
	t.Run("PushLimitTest", func(t *testing.T) {

		r := oracle.Open(t, "mover/PushLimitTest")
		planes := []box2d.CollisionPlane{collisionMakePlane(box2d.Vec2{0.0, 1.0}, 0.1)}
		result := box2d.SolvePlanes(box2d.Vec2{0.0, -1.0}, planes)
		geometryMoverObserve(r, result, planes)

		// Soft collision, so the mover sinks once the limit is spent
		r.F32("mover:61/value0", planes[0].Push)
		r.F32("mover:61", planes[0].Push-0.1)
		if v := planes[0].Push - 0.1; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}
		r.F32("mover:62/value0", result.Delta.Y)
		r.F32("mover:62", result.Delta.Y+0.9)
		if v := result.Delta.Y + 0.9; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

	})
	t.Run("CornerTest", func(t *testing.T) {

		r := oracle.Open(t, "mover/CornerTest")
		planes := []box2d.CollisionPlane{collisionMakePlane(box2d.Vec2{0.0, 1.0}, float32(math.MaxFloat32)), collisionMakePlane(box2d.Vec2{1.0, 0.0}, float32(math.MaxFloat32))}

		result := box2d.SolvePlanes(box2d.Vec2{-1.0, -1.0}, planes)
		geometryMoverObserve(r, result, planes)

		// Both planes hold at once, rather than the second relaxing the first
		for i := int(0); i < len(planes); i++ {
			separation := planes[i].Plane.Separation(result.Delta)
			r.F32("mover:82", separation+box2d.LinearSlop())
			if v := separation + box2d.LinearSlop(); v < -(box2d.LinearSlop()) || (box2d.LinearSlop()) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, box2d.LinearSlop())
			}
			r.Bool("mover:83", planes[i].Push > 0.0)
			if !(planes[i].Push > 0.0) {
				t.Fatal("condition false")
			}
		}

	})
	t.Run("WedgeTest", func(t *testing.T) {

		r := oracle.Open(t, "mover/WedgeTest")
		// An acute wedge takes several passes to settle, unlike a square corner
		planes := []box2d.CollisionPlane{collisionMakePlane(box2d.Vec2{0.0, 1.0}, float32(math.MaxFloat32)), collisionMakePlane(box2d.Vec2{0.8, 0.6}, float32(math.MaxFloat32))}

		result := box2d.SolvePlanes(box2d.Vec2{-1.0, -1.0}, planes)
		geometryMoverObserve(r, result, planes)

		// Settling means resting on both faces. Stopping early leaves the mover shoved
		// clear of one of them, which reads as a large positive separation
		for i := int(0); i < len(planes); i++ {
			separation := planes[i].Plane.Separation(result.Delta)
			r.F32("mover:106", separation+box2d.LinearSlop())
			if v := separation + box2d.LinearSlop(); v < -(box2d.LinearSlop()) || (box2d.LinearSlop()) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, box2d.LinearSlop())
			}
		}

	})
	t.Run("ClipVectorTest", func(t *testing.T) {

		r := oracle.Open(t, "mover/ClipVectorTest")
		planes := []box2d.CollisionPlane{collisionMakePlane(box2d.Vec2{0.0, 1.0}, float32(math.MaxFloat32))}
		planes[0].Push = 0.5

		v := box2d.ClipVector(box2d.Vec2{1.0, -1.0}, planes)
		geometryMoverClipObserve(r, v)
		r.F32("mover:120/value0", v.X)
		r.F32("mover:120", v.X-1.0)
		if v := v.X - 1.0; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}
		r.F32("mover:121/value0", v.Y)
		r.F32("mover:121", v.Y)
		if v := v.Y; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		// Separating velocity passes through untouched
		v = box2d.ClipVector(box2d.Vec2{1.0, 1.0}, planes)
		geometryMoverClipObserve(r, v)
		r.F32("mover:125/value0", v.X)
		r.F32("mover:125", v.X-1.0)
		if v := v.X - 1.0; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}
		r.F32("mover:126/value0", v.Y)
		r.F32("mover:126", v.Y-1.0)
		if v := v.Y - 1.0; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		planes[0].ClipVelocity = false
		v = box2d.ClipVector(box2d.Vec2{1.0, -1.0}, planes)
		geometryMoverClipObserve(r, v)
		r.F32("mover:130/value0", v.Y)
		r.F32("mover:130", v.Y+1.0)
		if v := v.Y + 1.0; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		// Zero push means the solver never touched this plane
		planes[0].ClipVelocity = true
		planes[0].Push = 0.0
		v = box2d.ClipVector(box2d.Vec2{1.0, -1.0}, planes)
		geometryMoverClipObserve(r, v)
		r.F32("mover:136/value0", v.Y)
		r.F32("mover:136", v.Y+1.0)
		if v := v.Y + 1.0; v < -(float32(0x1p-23)) || (float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

	})
}
