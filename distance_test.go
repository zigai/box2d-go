// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
	"testing"
)

func TestDistance(t *testing.T) {
	defer porttest.Guard(t)
	t.Run("SegmentDistanceTest", func(t *testing.T) {
		defer porttest.Guard(t)

		r := oracle.Open(t, "distance/SegmentDistanceTest")
		p1 := box2d.Vec2{-1.0, -1.0}
		q1 := box2d.Vec2{-1.0, 1.0}
		p2 := box2d.Vec2{2.0, 0.0}
		q2 := box2d.Vec2{1.0, 0.0}

		result := box2d.SegmentDistance(p1, q1, p2, q2)

		r.F32("distance:21/value0", result.Fraction1)
		r.F32("distance:21", result.Fraction1-0.5)

		porttest.EnsureSmall(t, result.Fraction1-0.5, float32(0x1p-23))
		r.F32("distance:23/value0", result.Fraction2)
		r.F32("distance:23", result.Fraction2-1.0)
		porttest.EnsureSmall(t, result.Fraction2-1.0, float32(0x1p-23))
		r.F32("distance:24/value0", result.Closest1.X)
		r.F32("distance:24", result.Closest1.X+1.0)
		porttest.EnsureSmall(t, result.Closest1.X+1.0, float32(0x1p-23))
		r.F32("distance:25/value0", result.Closest1.Y)
		r.F32("distance:25", result.Closest1.Y)
		porttest.EnsureSmall(t, result.Closest1.Y, float32(0x1p-23))
		r.F32("distance:26/value0", result.Closest2.X)
		r.F32("distance:26", result.Closest2.X-1.0)
		porttest.EnsureSmall(t, result.Closest2.X-1.0, float32(0x1p-23))
		r.F32("distance:27/value0", result.Closest2.Y)
		r.F32("distance:27", result.Closest2.Y)
		porttest.EnsureSmall(t, result.Closest2.Y, float32(0x1p-23))
		r.F32("distance:28/value0", result.DistanceSquared)
		r.F32("distance:28", result.DistanceSquared-4.0)
		porttest.EnsureSmall(t, result.DistanceSquared-4.0, float32(0x1p-23))

	})
	t.Run("ShapeDistanceTest", func(t *testing.T) {
		defer porttest.Guard(t)

		r := oracle.Open(t, "distance/ShapeDistanceTest")
		vas := []box2d.Vec2{box2d.Vec2{-1.0, -1.0}, box2d.Vec2{1.0, -1.0}, box2d.Vec2{1.0, 1.0}, box2d.Vec2{-1.0, 1.0}}

		vbs := []box2d.Vec2{box2d.Vec2{2.0, -1.0}, box2d.Vec2{2.0, 1.0}}

		var input box2d.DistanceInput
		input.ProxyA = box2d.MakeProxy(vas, 0.0)
		input.ProxyB = box2d.MakeProxy(vbs, 0.0)
		input.Transform = box2d.TransformIdentity
		input.UseRadii = false

		cache := box2d.SimplexCache{}
		output := box2d.ShapeDistance(&input, &cache, nil)

		r.F32("distance:52/value0", output.Distance)
		r.F32("distance:52", output.Distance-1.0)

		porttest.EnsureSmall(t, output.Distance-1.0, float32(0x1p-23))

	})
	t.Run("ShapeCastTest", func(t *testing.T) {
		defer porttest.Guard(t)

		r := oracle.Open(t, "distance/ShapeCastTest")
		vas := []box2d.Vec2{box2d.Vec2{-1.0, -1.0}, box2d.Vec2{1.0, -1.0}, box2d.Vec2{1.0, 1.0}, box2d.Vec2{-1.0, 1.0}}

		vbs := []box2d.Vec2{box2d.Vec2{2.0, -1.0}, box2d.Vec2{2.0, 1.0}}

		input := box2d.ShapeCastPairInput{}
		input.ProxyA = box2d.MakeProxy(vas, 0.0)
		input.ProxyB = box2d.MakeProxy(vbs, 0.0)
		input.Transform = box2d.TransformIdentity
		input.TranslationB = box2d.Vec2{-2.0, 0.0}
		input.MaxFraction = 1.0
		input.CanEncroach = false

		output := box2d.ShapeCast(&input)

		r.Bool("distance:78/value0", output.Hit)
		r.Bool("distance:78", output.Hit)

		porttest.Ensure(t, output.Hit)
		r.F32("distance:80/value0", output.Fraction)
		r.F32("distance:80", output.Fraction-0.5)
		porttest.EnsureSmall(t, output.Fraction-0.5, 0.005)

	})
	t.Run("ShapeDistanceShortEdgeTest", func(t *testing.T) {
		defer porttest.Guard(t)

		r := oracle.Open(t, "distance/ShapeDistanceShortEdgeTest")
		halfThickness := float32(0.0016)

		corners := []box2d.Vec2{box2d.Vec2{-0.5, -halfThickness}, box2d.Vec2{0.5, -halfThickness}, box2d.Vec2{0.5, halfThickness}, box2d.Vec2{-0.5, halfThickness}}

		point := box2d.Vec2{0.5 + box2d.LinearSlop(), 0.0}

		input := box2d.DistanceInput{}
		input.ProxyA = box2d.MakeProxy(corners, 0.0)
		input.ProxyB = box2d.MakeProxy([]box2d.Vec2{point}, 0.0)
		input.Transform = box2d.TransformIdentity
		input.UseRadii = false

		cache := box2d.SimplexCache{}
		output := box2d.ShapeDistance(&input, &cache, nil)

		r.F32("distance:118/value0", output.Distance)
		r.Bool("distance:118", output.Distance > 0.0)

		porttest.Ensure(t, output.Distance > 0.0)
		r.Vec2("distance:120/value0", output.Normal.X, output.Normal.Y)
		r.Bool("distance:120", box2d.IsNormalized(output.Normal))
		porttest.Ensure(t, box2d.IsNormalized(output.Normal))
		r.F32("distance:121/value0", output.Distance)
		r.F32("distance:121", output.Distance-box2d.LinearSlop())
		porttest.EnsureSmall(t, output.Distance-box2d.LinearSlop(), 1e-6)

	})
	t.Run("ShapeCastShortEdgeTest", func(t *testing.T) {
		defer porttest.Guard(t)

		t.Skip("disabled upstream: B2_SHORT_EDGE_CAST_REPRO 0")

		r := oracle.Open(t, "distance/ShapeCastShortEdgeTest")
		halfThickness := float32(0.0002)

		corners := []box2d.Vec2{box2d.Vec2{-0.5, -halfThickness}, box2d.Vec2{0.5, -halfThickness}, box2d.Vec2{0.5, halfThickness}, box2d.Vec2{-0.5, halfThickness}}

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
		r.Bool("distance:162", output.Hit == false || box2d.IsNormalized(output.Normal))

		porttest.Ensure(t, output.Hit == false || box2d.IsNormalized(output.Normal))

	})
	t.Run("TimeOfImpactTest", func(t *testing.T) {
		defer porttest.Guard(t)

		r := oracle.Open(t, "distance/TimeOfImpactTest")
		vas := []box2d.Vec2{box2d.Vec2{-1.0, -1.0}, box2d.Vec2{1.0, -1.0}, box2d.Vec2{1.0, 1.0}, box2d.Vec2{-1.0, 1.0}}

		vbs := []box2d.Vec2{box2d.Vec2{2.0, -1.0}, box2d.Vec2{2.0, 1.0}}

		var input box2d.TOIInput
		input.ProxyA = box2d.MakeProxy(vas, 0.0)
		input.ProxyB = box2d.MakeProxy(vbs, 0.0)
		input.SweepA = box2d.Sweep{box2d.Vec2Zero, box2d.Vec2Zero, box2d.Vec2Zero, box2d.RotIdentity, box2d.RotIdentity}
		input.SweepB = box2d.Sweep{box2d.Vec2Zero, box2d.Vec2Zero, box2d.Vec2{-2, 0}, box2d.RotIdentity, box2d.RotIdentity}
		input.MaxFraction = 1.0

		output := box2d.TimeOfImpact(&input)

		r.I32("distance:188/value0", int32(output.State))
		r.Bool("distance:188", output.State == box2d.TOIStateHit)

		porttest.Ensure(t, output.State == box2d.TOIStateHit)
		r.F32("distance:190/value0", output.Fraction)
		r.F32("distance:190", output.Fraction-0.5)
		porttest.EnsureSmall(t, output.Fraction-0.5, 0.005)

	})
}
