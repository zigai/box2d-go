// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func TestDistance(t *testing.T) {
	t.Run("SegmentDistanceTest", func(t *testing.T) {
		r := oracle.Open(t, "distance/SegmentDistanceTest")
		p1 := box2d.Vec2{-1.0, -1.0}
		q1 := box2d.Vec2{-1.0, 1.0}
		p2 := box2d.Vec2{2.0, 0.0}
		q2 := box2d.Vec2{1.0, 0.0}

		result := box2d.SegmentDistance(p1, q1, p2, q2)

		r.F32("distance:21/value0", result.Fraction1)
		r.F32("distance:21", result.Fraction1-0.5)

		if v := result.Fraction1 - 0.5; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("distance:23/value0", result.Fraction2)
		r.F32("distance:23", result.Fraction2-1.0)

		if v := result.Fraction2 - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("distance:24/value0", result.Closest1.X)
		r.F32("distance:24", result.Closest1.X+1.0)

		if v := result.Closest1.X + 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("distance:25/value0", result.Closest1.Y)
		r.F32("distance:25", result.Closest1.Y)

		if v := result.Closest1.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("distance:26/value0", result.Closest2.X)
		r.F32("distance:26", result.Closest2.X-1.0)

		if v := result.Closest2.X - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("distance:27/value0", result.Closest2.Y)
		r.F32("distance:27", result.Closest2.Y)

		if v := result.Closest2.Y; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}

		r.F32("distance:28/value0", result.DistanceSquared)
		r.F32("distance:28", result.DistanceSquared-4.0)

		if v := result.DistanceSquared - 4.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}
	})
	t.Run("ShapeDistanceTest", func(t *testing.T) {
		r := oracle.Open(t, "distance/ShapeDistanceTest")
		vas := []box2d.Vec2{{-1.0, -1.0}, {1.0, -1.0}, {1.0, 1.0}, {-1.0, 1.0}}

		vbs := []box2d.Vec2{{2.0, -1.0}, {2.0, 1.0}}

		var input box2d.DistanceInput

		input.ProxyA = box2d.MakeProxy(vas, 0.0)
		input.ProxyB = box2d.MakeProxy(vbs, 0.0)
		input.Transform = box2d.TransformIdentity
		input.UseRadii = false

		cache := box2d.SimplexCache{}
		output := box2d.ShapeDistance(&input, &cache, nil)
		geometryOriginalDistance(r, output, cache)

		r.F32("distance:52/value0", output.Distance)
		r.F32("distance:52", output.Distance-1.0)

		if v := output.Distance - 1.0; v < -float32(0x1p-23) || float32(0x1p-23) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0x1p-23))
		}
	})
	t.Run("ShapeCastTest", func(t *testing.T) {
		r := oracle.Open(t, "distance/ShapeCastTest")
		vas := []box2d.Vec2{{-1.0, -1.0}, {1.0, -1.0}, {1.0, 1.0}, {-1.0, 1.0}}

		vbs := []box2d.Vec2{{2.0, -1.0}, {2.0, 1.0}}

		input := box2d.ShapeCastPairInput{}
		input.ProxyA = box2d.MakeProxy(vas, 0.0)
		input.ProxyB = box2d.MakeProxy(vbs, 0.0)
		input.Transform = box2d.TransformIdentity
		input.TranslationB = box2d.Vec2{-2.0, 0.0}
		input.MaxFraction = 1.0
		input.CanEncroach = false

		output := box2d.ShapeCast(&input)
		geometryOriginalCast(r, output)

		r.Bool("distance:78/value0", output.Hit)
		r.Bool("distance:78", output.Hit)

		if !output.Hit {
			t.Fatal("condition false")
		}

		r.F32("distance:80/value0", output.Fraction)
		r.F32("distance:80", output.Fraction-0.5)

		if v := output.Fraction - 0.5; v < -0.005 || 0.005 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.005)
		}
	})
	t.Run("ShapeDistanceShortEdgeTest", func(t *testing.T) {
		r := oracle.Open(t, "distance/ShapeDistanceShortEdgeTest")
		halfThickness := float32(0.0016)

		corners := []box2d.Vec2{{-0.5, -halfThickness}, {0.5, -halfThickness}, {0.5, halfThickness}, {-0.5, halfThickness}}

		point := box2d.Vec2{0.5 + box2d.LinearSlop(), 0.0}

		input := box2d.DistanceInput{}
		input.ProxyA = box2d.MakeProxy(corners, 0.0)
		input.ProxyB = box2d.MakeProxy([]box2d.Vec2{point}, 0.0)
		input.Transform = box2d.TransformIdentity
		input.UseRadii = false

		cache := box2d.SimplexCache{}
		output := box2d.ShapeDistance(&input, &cache, nil)
		geometryOriginalDistance(r, output, cache)

		r.F32("distance:118/value0", output.Distance)
		r.Bool("distance:118", output.Distance > 0.0)

		if !(output.Distance > 0.0) {
			t.Fatal("condition false")
		}

		r.Vec2("distance:120/value0", output.Normal.X, output.Normal.Y)
		r.Bool("distance:120", output.Normal.IsNormalized())

		if !output.Normal.IsNormalized() {
			t.Fatal("condition false")
		}

		r.F32("distance:121/value0", output.Distance)
		r.F32("distance:121", output.Distance-box2d.LinearSlop())

		if v := output.Distance - box2d.LinearSlop(); v < -1e-6 || 1e-6 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 1e-6)
		}
	})
	t.Run("ShapeCastShortEdgeTest", func(t *testing.T) {
		t.Skip("disabled upstream: B2_SHORT_EDGE_CAST_REPRO 0")

		r := oracle.Open(t, "distance/ShapeCastShortEdgeTest")
		halfThickness := float32(0.0002)

		corners := []box2d.Vec2{{-0.5, -halfThickness}, {0.5, -halfThickness}, {0.5, halfThickness}, {-0.5, halfThickness}}

		// falls past the end of the plank and drifts into it, so the second query lands on the end face
		start := box2d.Vec2{0.506, 0.0035}

		input := box2d.ShapeCastPairInput{}
		input.ProxyA = box2d.MakeProxy(corners, 0.0)
		input.ProxyB = box2d.MakeProxy([]box2d.Vec2{start}, 0.0)
		input.Transform = box2d.TransformIdentity
		input.TranslationB = box2d.Vec2{-0.01, -0.5}
		input.MaxFraction = 1.0
		input.CanEncroach = false

		output := box2d.ShapeCast(&input)

		r.Bool("distance:162/value0", output.Hit)
		r.Vec2("distance:162/value1", output.Normal.X, output.Normal.Y)
		r.Bool("distance:162", output.Hit == false || output.Normal.IsNormalized())

		if output.Hit != false && !output.Normal.IsNormalized() {
			t.Fatal("condition false")
		}
	})
	t.Run("TimeOfImpactTest", func(t *testing.T) {
		r := oracle.Open(t, "distance/TimeOfImpactTest")
		vas := []box2d.Vec2{{-1.0, -1.0}, {1.0, -1.0}, {1.0, 1.0}, {-1.0, 1.0}}

		vbs := []box2d.Vec2{{2.0, -1.0}, {2.0, 1.0}}

		var input box2d.TOIInput

		input.ProxyA = box2d.MakeProxy(vas, 0.0)
		input.ProxyB = box2d.MakeProxy(vbs, 0.0)
		input.SweepA = box2d.Sweep{box2d.Vec2Zero, box2d.Vec2Zero, box2d.Vec2Zero, box2d.RotIdentity, box2d.RotIdentity}
		input.SweepB = box2d.Sweep{box2d.Vec2Zero, box2d.Vec2Zero, box2d.Vec2{-2, 0}, box2d.RotIdentity, box2d.RotIdentity}
		input.MaxFraction = 1.0

		output := box2d.TimeOfImpact(&input)
		r.I32("full/toi/state", int32(output.State))
		r.F32("full/toi/fraction", output.Fraction)

		if output.State == box2d.TOIStateHit {
			r.Vec2("full/toi/point", output.Point.X, output.Point.Y)
			r.Vec2("full/toi/normal", output.Normal.X, output.Normal.Y)
		}

		r.I32("distance:188/value0", int32(output.State))
		r.Bool("distance:188", output.State == box2d.TOIStateHit)

		if output.State != box2d.TOIStateHit {
			t.Fatal("condition false")
		}

		r.F32("distance:190/value0", output.Fraction)
		r.F32("distance:190", output.Fraction-0.5)

		if v := output.Fraction - 0.5; v < -0.005 || 0.005 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.005)
		}
	})
}
