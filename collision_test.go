// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"testing"
	"unsafe"

	"github.com/zigai/box2d/internal/oracle"
)

func TestCollision(t *testing.T) {
	t.Run("AABBTest", func(t *testing.T) {
		r := oracle.Open(t, "collision/AABBTest")

		var a AABB

		a.LowerBound = Vec2{-1.0, -1.0}
		a.UpperBound = Vec2{-2.0, -2.0}

		r.Bool("collision:17", a.IsValid() == false)

		if a.IsValid() != false {
			t.Fatal("condition false")
		}

		a.UpperBound = Vec2{1.0, 1.0}
		r.Bool("collision:21", a.IsValid() == true)

		if a.IsValid() != true {
			t.Fatal("condition false")
		}

		b := AABB{Vec2{2.0, 2.0}, Vec2{4.0, 4.0}}
		r.Bool("collision:24", a.Overlaps(b) == false)

		if a.Overlaps(b) != false {
			t.Fatal("condition false")
		}

		r.Bool("collision:25", a.Contains(b) == false)

		if a.Contains(b) != false {
			t.Fatal("condition false")
		}
	})
	t.Run("AABBRayCastTest", func(t *testing.T) {
		r := oracle.Open(t, "collision/AABBRayCastTest")
		// Test AABB centered at origin with bounds [-1, -1] to [1, 1]
		aabb := AABB{Vec2{-1.0, -1.0}, Vec2{1.0, 1.0}}

		// Test 1: Ray hits AABB from left side
		{
			p1 := Vec2{-3.0, 0.0}
			p2 := Vec2{3.0, 0.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:42/value0", output.Hit)
			r.Bool("collision:42", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:44/value0", output.Fraction)
			r.F32("collision:44", output.Fraction-1.0/3.0)

			if v := output.Fraction - 1.0/3.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:45/value0", output.Normal.X)
			r.F32("collision:45", output.Normal.X+1.0)

			if v := output.Normal.X + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:46/value0", output.Normal.Y)
			r.F32("collision:46", output.Normal.Y)

			if v := output.Normal.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:47/value0", output.Point.X)
			r.F32("collision:47", output.Point.X+1.0)

			if v := output.Point.X + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:48/value0", output.Point.Y)
			r.F32("collision:48", output.Point.Y)

			if v := output.Point.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}

		// Test 2: Ray hits AABB from right side
		{
			p1 := Vec2{3.0, 0.0}
			p2 := Vec2{-3.0, 0.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:56/value0", output.Hit)
			r.Bool("collision:56", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:58/value0", output.Fraction)
			r.F32("collision:58", output.Fraction-1.0/3.0)

			if v := output.Fraction - 1.0/3.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:59/value0", output.Normal.X)
			r.F32("collision:59", output.Normal.X-1.0)

			if v := output.Normal.X - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:60/value0", output.Normal.Y)
			r.F32("collision:60", output.Normal.Y)

			if v := output.Normal.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:61/value0", output.Point.X)
			r.F32("collision:61", output.Point.X-1.0)

			if v := output.Point.X - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:62/value0", output.Point.Y)
			r.F32("collision:62", output.Point.Y)

			if v := output.Point.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}

		// Test 3: Ray hits AABB from bottom
		{
			p1 := Vec2{0.0, -3.0}
			p2 := Vec2{0.0, 3.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:70/value0", output.Hit)
			r.Bool("collision:70", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:72/value0", output.Fraction)
			r.F32("collision:72", output.Fraction-1.0/3.0)

			if v := output.Fraction - 1.0/3.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:73/value0", output.Normal.X)
			r.F32("collision:73", output.Normal.X)

			if v := output.Normal.X; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:74/value0", output.Normal.Y)
			r.F32("collision:74", output.Normal.Y+1.0)

			if v := output.Normal.Y + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:75/value0", output.Point.X)
			r.F32("collision:75", output.Point.X)

			if v := output.Point.X; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:76/value0", output.Point.Y)
			r.F32("collision:76", output.Point.Y+1.0)

			if v := output.Point.Y + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}

		// Test 4: Ray hits AABB from top
		{
			p1 := Vec2{0.0, 3.0}
			p2 := Vec2{0.0, -3.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:84/value0", output.Hit)
			r.Bool("collision:84", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:86/value0", output.Fraction)
			r.F32("collision:86", output.Fraction-1.0/3.0)

			if v := output.Fraction - 1.0/3.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:87/value0", output.Normal.X)
			r.F32("collision:87", output.Normal.X)

			if v := output.Normal.X; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:88/value0", output.Normal.Y)
			r.F32("collision:88", output.Normal.Y-1.0)

			if v := output.Normal.Y - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:89/value0", output.Point.X)
			r.F32("collision:89", output.Point.X)

			if v := output.Point.X; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:90/value0", output.Point.Y)
			r.F32("collision:90", output.Point.Y-1.0)

			if v := output.Point.Y - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}

		// Test 5: Ray misses AABB completely (parallel to x-axis)
		{
			p1 := Vec2{-3.0, 2.0}
			p2 := Vec2{3.0, 2.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:98/value0", output.Hit)
			r.Bool("collision:98", output.Hit == false)

			if output.Hit != false {
				t.Fatal("condition false")
			}
		}

		// Test 6: Ray misses AABB completely (parallel to y-axis)
		{
			p1 := Vec2{2.0, -3.0}
			p2 := Vec2{2.0, 3.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:107/value0", output.Hit)
			r.Bool("collision:107", output.Hit == false)

			if output.Hit != false {
				t.Fatal("condition false")
			}
		}

		// Test 7: Ray starts inside AABB
		{
			p1 := Vec2{0.0, 0.0}
			p2 := Vec2{2.0, 0.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:116/value0", output.Hit)
			r.Bool("collision:116", output.Hit == false)

			if output.Hit != false {
				t.Fatal("condition false")
			}
		}

		// Test 8: Ray hits corner of AABB (diagonal ray)
		{
			p1 := Vec2{-2.0, -2.0}
			p2 := Vec2{2.0, 2.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:125/value0", output.Hit)
			r.Bool("collision:125", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:127/value0", output.Fraction)
			r.F32("collision:127", output.Fraction-0.25)

			if v := output.Fraction - 0.25; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
			// Normal should be either (-1, 0) or (0, -1) depending on which edge is hit first
			r.F32("collision:129/value0", output.Normal.X)
			r.F32("collision:129/value1", output.Normal.Y)
			r.Bool("collision:129", (output.Normal.X == -1.0 && output.Normal.Y == 0.0) ||
				(output.Normal.X == 0.0 && output.Normal.Y == -1.0))

			if !((output.Normal.X == -1.0 && output.Normal.Y == 0.0) ||
				(output.Normal.X == 0.0 && output.Normal.Y == -1.0)) {
				t.Fatal("condition false")
			}
		}

		// Test 9: Ray parallel to AABB edge but outside
		{
			p1 := Vec2{-2.0, 1.5}
			p2 := Vec2{2.0, 1.5}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:138/value0", output.Hit)
			r.Bool("collision:138", output.Hit == false)

			if output.Hit != false {
				t.Fatal("condition false")
			}
		}

		// Test 10: Ray parallel to AABB edge and exactly on boundary
		{
			p1 := Vec2{-2.0, 1.0}
			p2 := Vec2{2.0, 1.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:147/value0", output.Hit)
			r.Bool("collision:147", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:149/value0", output.Fraction)
			r.F32("collision:149", output.Fraction-0.25)

			if v := output.Fraction - 0.25; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:150/value0", output.Normal.X)
			r.F32("collision:150", output.Normal.X+1.0)

			if v := output.Normal.X + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:151/value0", output.Normal.Y)
			r.F32("collision:151", output.Normal.Y)

			if v := output.Normal.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}

		// Test 11: Very short ray that doesn't reach AABB
		{
			p1 := Vec2{-3.0, 0.0}
			p2 := Vec2{-2.5, 0.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:159/value0", output.Hit)
			r.Bool("collision:159", output.Hit == false)

			if output.Hit != false {
				t.Fatal("condition false")
			}
		}

		// Test 12: Zero-length ray (degenerate case)
		{
			p1 := Vec2{0.0, 0.0}
			p2 := Vec2{0.0, 0.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:168/value0", output.Hit)
			r.Bool("collision:168", output.Hit == false)

			if output.Hit != false {
				t.Fatal("condition false")
			}
		}

		// Test 13: Ray hits AABB at exact boundary condition (t = 1.0)
		{
			p1 := Vec2{-2.0, 0.0}
			p2 := Vec2{-1.0, 0.0}
			output := aabbRayCast(aabb, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:177/value0", output.Hit)
			r.Bool("collision:177", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:179/value0", output.Fraction)
			r.F32("collision:179", output.Fraction-1.0)

			if v := output.Fraction - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:180/value0", output.Normal.X)
			r.F32("collision:180", output.Normal.X+1.0)

			if v := output.Normal.X + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:181/value0", output.Normal.Y)
			r.F32("collision:181", output.Normal.Y)

			if v := output.Normal.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}

		// Test 14: Different AABB position (not centered at origin)
		{
			offsetAABB := AABB{Vec2{2.0, 3.0}, Vec2{4.0, 5.0}}
			p1 := Vec2{0.0, 4.0}
			p2 := Vec2{6.0, 4.0}
			output := aabbRayCast(offsetAABB, p1, p2)
			geometryCheckAABBCast(r, output)

			r.Bool("collision:190/value0", output.Hit)
			r.Bool("collision:190", output.Hit == true)

			if output.Hit != true {
				t.Fatal("condition false")
			}

			r.F32("collision:192/value0", output.Fraction)
			r.F32("collision:192", output.Fraction-1.0/3.0)

			if v := output.Fraction - 1.0/3.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:193/value0", output.Normal.X)
			r.F32("collision:193", output.Normal.X+1.0)

			if v := output.Normal.X + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:194/value0", output.Normal.Y)
			r.F32("collision:194", output.Normal.Y)

			if v := output.Normal.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:195/value0", output.Point.X)
			r.F32("collision:195", output.Point.X-2.0)

			if v := output.Point.X - 2.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}

			r.F32("collision:196/value0", output.Point.Y)
			r.F32("collision:196", output.Point.Y-4.0)

			if v := output.Point.Y - 4.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
			}
		}
	})
	t.Run("LargeWorldManifoldTest", func(t *testing.T) {
		r := oracle.Open(t, "collision/LargeWorldManifoldTest")
		boxA := MakeBox(0.5, 0.5)
		boxB := MakeBox(0.5, 0.5)

		// Centers 0.9 apart so the boxes overlap by 0.1 along x
		sep := Vec2{0.9, 0.0}

		xfAo := WorldTransformIdentity
		xfBo := WorldTransform{PosZero.Add(sep), RotIdentity}
		mOrigin := CollidePolygons(&boxA, &boxB, xfAo.InvMul(xfBo))
		geometryCheckManifold(r, "full/origin", mOrigin)

		r.I32("collision:218/value0", int32(mOrigin.PointCount))
		r.Bool("collision:218", mOrigin.PointCount == 2)

		if mOrigin.PointCount != 2 {
			t.Fatal("condition false")
		}

		r.F32("collision:220/value0", mOrigin.Points[0].Separation)
		r.F32("collision:220", mOrigin.Points[0].Separation+0.1)

		if v := mOrigin.Points[0].Separation + 0.1; v < -0.01 || 0.01 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.01)
		}

		r.F32("collision:221/value0", mOrigin.Points[1].Separation)
		r.F32("collision:221", mOrigin.Points[1].Separation+0.1)

		if v := mOrigin.Points[1].Separation + 0.1; v < -0.01 || 0.01 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.01)
		}

		if unsafe.Sizeof(Pos{}.X) == 8 {
			// Same relative configuration shifted far from the origin. The relative pose differences the
			// world positions in double, so in double the frame A manifold is preserved to float precision.
			// In float it would collapse since the offset is below the ULP.
			base := PosZero.Add(Vec2{1.0e7, 1.0e7})
			xfAl := WorldTransform{base, RotIdentity}
			xfBl := WorldTransform{base.Add(sep), RotIdentity}
			mLarge := CollidePolygons(&boxA, &boxB, xfAl.InvMul(xfBl))
			geometryCheckManifold(r, "full/large", mLarge)

			r.I32("collision:231/value0", int32(mLarge.PointCount))
			r.I32("collision:231/value1", int32(mOrigin.PointCount))
			r.Bool("collision:231", mLarge.PointCount == mOrigin.PointCount)

			if mLarge.PointCount != mOrigin.PointCount {
				t.Fatal("condition false")
			}

			r.F32("collision:233/value0", mLarge.Normal.X)
			r.F32("collision:233/value1", mOrigin.Normal.X)
			r.F32("collision:233", mLarge.Normal.X-mOrigin.Normal.X)

			if v := mLarge.Normal.X - mOrigin.Normal.X; v < -1e-4 || 1e-4 < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 1e-4)
			}

			r.F32("collision:234/value0", mLarge.Normal.Y)
			r.F32("collision:234/value1", mOrigin.Normal.Y)
			r.F32("collision:234", mLarge.Normal.Y-mOrigin.Normal.Y)

			if v := mLarge.Normal.Y - mOrigin.Normal.Y; v < -1e-4 || 1e-4 < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 1e-4)
			}

			for i := range mLarge.PointCount {
				r.F32("collision:237/value0", mLarge.Points[i].Separation)
				r.F32("collision:237/value1", mOrigin.Points[i].Separation)
				r.F32("collision:237", mLarge.Points[i].Separation-mOrigin.Points[i].Separation)

				if v := mLarge.Points[i].Separation - mOrigin.Points[i].Separation; v < -1e-4 || 1e-4 < v {
					t.Fatalf("condition false: abs(%v) < %v", v, 1e-4)
				}

				r.F32("collision:238/value0", mLarge.Points[i].Point.X)
				r.F32("collision:238/value1", mOrigin.Points[i].Point.X)
				r.F32("collision:238", mLarge.Points[i].Point.X-mOrigin.Points[i].Point.X)

				if v := mLarge.Points[i].Point.X - mOrigin.Points[i].Point.X; v < -1e-4 || 1e-4 < v {
					t.Fatalf("condition false: abs(%v) < %v", v, 1e-4)
				}

				r.F32("collision:239/value0", mLarge.Points[i].Point.Y)
				r.F32("collision:239/value1", mOrigin.Points[i].Point.Y)
				r.F32("collision:239", mLarge.Points[i].Point.Y-mOrigin.Points[i].Point.Y)

				if v := mLarge.Points[i].Point.Y - mOrigin.Points[i].Point.Y; v < -1e-4 || 1e-4 < v {
					t.Fatalf("condition false: abs(%v) < %v", v, 1e-4)
				}
			}
		}
	})
	t.Run("LargeWorldAABBTest", func(t *testing.T) {
		r := oracle.Open(t, "collision/LargeWorldAABBTest")
		// Rounded box: 0.5 half extents plus 0.1 radius, so the tight extent is 0.6 each way
		box := MakeRoundedBox(0.5, 0.5, 0.1)

		aabbOrigin := ComputePolygonAABB(&box, WorldTransformIdentity)
		r.F32("collision:257/value0", aabbOrigin.LowerBound.X)
		r.F32("collision:257", aabbOrigin.LowerBound.X+0.6)

		if v := aabbOrigin.LowerBound.X + 0.6; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("collision:258/value0", aabbOrigin.LowerBound.Y)
		r.F32("collision:258", aabbOrigin.LowerBound.Y+0.6)

		if v := aabbOrigin.LowerBound.Y + 0.6; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("collision:259/value0", aabbOrigin.UpperBound.X)
		r.F32("collision:259", aabbOrigin.UpperBound.X-0.6)

		if v := aabbOrigin.UpperBound.X - 0.6; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("collision:260/value0", aabbOrigin.UpperBound.Y)
		r.F32("collision:260", aabbOrigin.UpperBound.Y-0.6)

		if v := aabbOrigin.UpperBound.Y - 0.6; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		if unsafe.Sizeof(Pos{}.X) == 8 {
			d := 1.0e7
			xfLarge := WorldTransform{Pos{1.0e7, 1.0e7}, RotIdentity}

			// Tight world AABB still contains the 0.6 m extent
			tight := ComputePolygonAABB(&box, xfLarge)
			r.F32("collision:268/value0", tight.LowerBound.X)
			r.Bool("collision:268", float64(tight.LowerBound.X) <= d-0.6)

			if !(float64(tight.LowerBound.X) <= d-0.6) {
				t.Fatal("condition false")
			}

			r.F32("collision:269/value0", tight.LowerBound.Y)
			r.Bool("collision:269", float64(tight.LowerBound.Y) <= d-0.6)

			if !(float64(tight.LowerBound.Y) <= d-0.6) {
				t.Fatal("condition false")
			}

			r.F32("collision:270/value0", tight.UpperBound.X)
			r.Bool("collision:270", float64(tight.UpperBound.X) >= d+0.6)

			if !(float64(tight.UpperBound.X) >= d+0.6) {
				t.Fatal("condition false")
			}

			r.F32("collision:271/value0", tight.UpperBound.Y)
			r.Bool("collision:271", float64(tight.UpperBound.Y) >= d+0.6)

			if !(float64(tight.UpperBound.Y) >= d+0.6) {
				t.Fatal("condition false")
			}

			// The fat helper folds the extra into the double step before the single outward rounding, so a
			// margin smaller than a float ULP at this range survives instead of becoming a no-op subtract.
			extra := float32(0.05)
			shape := shape{}
			shape.shapeType = PolygonShape
			shape.polygon = box
			fat := computeFatShapeAABB(&shape, xfLarge, extra)
			r.F32("collision:280/value0", fat.LowerBound.X)
			r.Bool("collision:280", float64(fat.LowerBound.X) <= d-0.6-float64(extra))

			if !(float64(fat.LowerBound.X) <= d-0.6-float64(extra)) {
				t.Fatal("condition false")
			}

			r.F32("collision:281/value0", fat.LowerBound.Y)
			r.Bool("collision:281", float64(fat.LowerBound.Y) <= d-0.6-float64(extra))

			if !(float64(fat.LowerBound.Y) <= d-0.6-float64(extra)) {
				t.Fatal("condition false")
			}

			r.F32("collision:282/value0", fat.UpperBound.X)
			r.Bool("collision:282", float64(fat.UpperBound.X) >= d+0.6+float64(extra))

			if !(float64(fat.UpperBound.X) >= d+0.6+float64(extra)) {
				t.Fatal("condition false")
			}

			r.F32("collision:283/value0", fat.UpperBound.Y)
			r.Bool("collision:283", float64(fat.UpperBound.Y) >= d+0.6+float64(extra))

			if !(float64(fat.UpperBound.Y) >= d+0.6+float64(extra)) {
				t.Fatal("condition false")
			}
		}
	})
}
