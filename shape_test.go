// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
	"math"
	"testing"
)

const collisionShapeN = 4

func TestShape(t *testing.T) {
	defer porttest.Guard(t)
	t.Run("ShapeMassTest", func(t *testing.T) {
		defer porttest.Guard(t)

		o := oracle.Open(t, "shape/ShapeMassTest")
		box := box2d.MakeBox(1, 1)
		capsule := box2d.Capsule{box2d.Vec2{-1, 0}, box2d.Vec2{1, 0}, 1}
		circle := box2d.Circle{box2d.Vec2{1, 0}, 1}

		{
			md := box2d.ComputeCircleMass(&circle, 1.0)
			o.F32("shape:23/value0", md.Mass)
			o.F32("shape:23", md.Mass-box2d.Pi)
			porttest.EnsureSmall(t, md.Mass-box2d.Pi, float32(0x1p-23))
			o.F32("shape:24/value0", md.Center.X)
			o.F32("shape:24/value1", md.Center.Y)
			o.Bool("shape:24", md.Center.X == 1.0 && md.Center.Y == 0.0)
			porttest.Ensure(t, md.Center.X == 1.0 && md.Center.Y == 0.0)
			o.F32("shape:25/value0", md.RotationalInertia)
			o.F32("shape:25", md.RotationalInertia-0.5*box2d.Pi)
			porttest.EnsureSmall(t, md.RotationalInertia-0.5*box2d.Pi, float32(0x1p-23))
		}

		{
			radius := capsule.Radius
			length := box2d.Distance(capsule.Center1, capsule.Center2)

			md := box2d.ComputeCapsuleMass(&capsule, 1.0)

			// Box that full contains capsule
			r := box2d.MakeBox(radius+float32(0.5*length), radius)
			mdUpper := box2d.ComputePolygonMass(&r, 1.0)

			// Approximate capsule using convex hull
			var points [2 * collisionShapeN]box2d.Vec2
			d := box2d.Pi / (collisionShapeN - 1.0)
			angle := -0.5 * box2d.Pi
			for i := int(0); i < collisionShapeN; i++ {
				points[i].X = 1.0 + float32(radius*float32(math.Cos(float64(angle))))
				points[i].Y = radius * float32(math.Sin(float64(angle)))
				angle += d
			}

			angle = 0.5 * box2d.Pi
			for i := collisionShapeN; i < 2*collisionShapeN; i++ {
				points[i].X = -1.0 + float32(radius*float32(math.Cos(float64(angle))))
				points[i].Y = radius * float32(math.Sin(float64(angle)))
				angle += d
			}

			hull := box2d.ComputeHull(points[:], 2*collisionShapeN)
			ac := box2d.MakePolygon(&hull, 0.0)
			mdLower := box2d.ComputePolygonMass(&ac, 1.0)

			o.F32("shape:60/value0", mdLower.Mass)
			o.F32("shape:60/value1", md.Mass)
			o.F32("shape:60/value2", mdUpper.Mass)
			o.Bool("shape:60", mdLower.Mass < md.Mass && md.Mass < mdUpper.Mass)

			porttest.Ensure(t, mdLower.Mass < md.Mass && md.Mass < mdUpper.Mass)
			o.F32("shape:62/value0", mdLower.RotationalInertia)
			o.F32("shape:62/value1", md.RotationalInertia)
			o.F32("shape:62/value2", mdUpper.RotationalInertia)
			o.Bool("shape:62", mdLower.RotationalInertia < md.RotationalInertia && md.RotationalInertia < mdUpper.RotationalInertia)
			porttest.Ensure(t, mdLower.RotationalInertia < md.RotationalInertia && md.RotationalInertia < mdUpper.RotationalInertia)
		}

		{
			md := box2d.ComputePolygonMass(&box, 1.0)
			o.F32("shape:67/value0", md.Mass)
			o.F32("shape:67", md.Mass-4.0)
			porttest.EnsureSmall(t, md.Mass-4.0, float32(0x1p-23))
			o.F32("shape:68/value0", md.Center.X)
			o.F32("shape:68", md.Center.X)
			porttest.EnsureSmall(t, md.Center.X, float32(0x1p-23))
			o.F32("shape:69/value0", md.Center.Y)
			o.F32("shape:69", md.Center.Y)
			porttest.EnsureSmall(t, md.Center.Y, float32(0x1p-23))
			o.F32("shape:70/value0", md.RotationalInertia)
			o.F32("shape:70", md.RotationalInertia-8.0/3.0)
			porttest.EnsureSmall(t, md.RotationalInertia-8.0/3.0, 2.0*float32(0x1p-23))
		}

		{
			offset := box2d.Vec2{0.4, -0.7}
			b1 := box2d.MakeBox(0.25, 0.5)
			b2 := box2d.MakeOffsetBox(0.25, 0.5, offset, box2d.RotIdentity)

			m1 := box2d.ComputePolygonMass(&b1, 1.0)
			m2 := box2d.ComputePolygonMass(&b2, 1.0)

			o.F32("shape:80/value0", m1.Mass)
			o.F32("shape:80/value1", m2.Mass)
			o.F32("shape:80", m1.Mass-m2.Mass)

			porttest.EnsureSmall(t, m1.Mass-m2.Mass, float32(0x1p-23))
			o.F32("shape:82/value0", m1.RotationalInertia)
			o.F32("shape:82/value1", m2.RotationalInertia)
			o.F32("shape:82", m1.RotationalInertia-m2.RotationalInertia)
			porttest.EnsureSmall(t, m1.RotationalInertia-m2.RotationalInertia, float32(0x1p-23))
			o.F32("shape:83/value0", m2.Center.X)
			o.F32("shape:83/value1", offset.X)
			o.F32("shape:83", m2.Center.X-offset.X)
			porttest.EnsureSmall(t, m2.Center.X-offset.X, float32(0x1p-23))
			o.F32("shape:84/value0", m2.Center.Y)
			o.F32("shape:84/value1", offset.Y)
			o.F32("shape:84", m2.Center.Y-offset.Y)
			porttest.EnsureSmall(t, m2.Center.Y-offset.Y, float32(0x1p-23))
		}

	})
	t.Run("ShapeAABBTest", func(t *testing.T) {
		defer porttest.Guard(t)

		o := oracle.Open(t, "shape/ShapeAABBTest")
		box := box2d.MakeBox(1, 1)
		circle := box2d.Circle{box2d.Vec2{1, 0}, 1}
		segment := box2d.Segment{box2d.Vec2{0, 1}, box2d.Vec2{0, -1}}

		{
			b := box2d.ComputeCircleAABB(&circle, box2d.WorldTransformIdentity)
			o.F32("shape:96/value0", b.LowerBound.X)
			o.F32("shape:96", b.LowerBound.X)
			porttest.EnsureSmall(t, b.LowerBound.X, float32(0x1p-23))
			o.F32("shape:97/value0", b.LowerBound.Y)
			o.F32("shape:97", b.LowerBound.Y+1.0)
			porttest.EnsureSmall(t, b.LowerBound.Y+1.0, float32(0x1p-23))
			o.F32("shape:98/value0", b.UpperBound.X)
			o.F32("shape:98", b.UpperBound.X-2.0)
			porttest.EnsureSmall(t, b.UpperBound.X-2.0, float32(0x1p-23))
			o.F32("shape:99/value0", b.UpperBound.Y)
			o.F32("shape:99", b.UpperBound.Y-1.0)
			porttest.EnsureSmall(t, b.UpperBound.Y-1.0, float32(0x1p-23))
		}

		{
			b := box2d.ComputePolygonAABB(&box, box2d.WorldTransformIdentity)
			o.F32("shape:104/value0", b.LowerBound.X)
			o.F32("shape:104", b.LowerBound.X+1.0)
			porttest.EnsureSmall(t, b.LowerBound.X+1.0, float32(0x1p-23))
			o.F32("shape:105/value0", b.LowerBound.Y)
			o.F32("shape:105", b.LowerBound.Y+1.0)
			porttest.EnsureSmall(t, b.LowerBound.Y+1.0, float32(0x1p-23))
			o.F32("shape:106/value0", b.UpperBound.X)
			o.F32("shape:106", b.UpperBound.X-1.0)
			porttest.EnsureSmall(t, b.UpperBound.X-1.0, float32(0x1p-23))
			o.F32("shape:107/value0", b.UpperBound.Y)
			o.F32("shape:107", b.UpperBound.Y-1.0)
			porttest.EnsureSmall(t, b.UpperBound.Y-1.0, float32(0x1p-23))
		}

		{
			b := box2d.ComputeSegmentAABB(&segment, box2d.WorldTransformIdentity)
			o.F32("shape:112/value0", b.LowerBound.X)
			o.F32("shape:112", b.LowerBound.X)
			porttest.EnsureSmall(t, b.LowerBound.X, float32(0x1p-23))
			o.F32("shape:113/value0", b.LowerBound.Y)
			o.F32("shape:113", b.LowerBound.Y+1.0)
			porttest.EnsureSmall(t, b.LowerBound.Y+1.0, float32(0x1p-23))
			o.F32("shape:114/value0", b.UpperBound.X)
			o.F32("shape:114", b.UpperBound.X)
			porttest.EnsureSmall(t, b.UpperBound.X, float32(0x1p-23))
			o.F32("shape:115/value0", b.UpperBound.Y)
			o.F32("shape:115", b.UpperBound.Y-1.0)
			porttest.EnsureSmall(t, b.UpperBound.Y-1.0, float32(0x1p-23))
		}

	})
	t.Run("PointInShapeTest", func(t *testing.T) {
		defer porttest.Guard(t)

		o := oracle.Open(t, "shape/PointInShapeTest")
		box := box2d.MakeBox(1, 1)
		circle := box2d.Circle{box2d.Vec2{1, 0}, 1}

		p1 := box2d.Vec2{0.5, 0.5}
		p2 := box2d.Vec2{4.0, -4.0}

		{
			var hit bool
			hit = box2d.PointInCircle(&circle, p1)
			o.Bool("shape:131", hit == true)
			porttest.Ensure(t, hit == true)
			hit = box2d.PointInCircle(&circle, p2)
			o.Bool("shape:133", hit == false)
			porttest.Ensure(t, hit == false)
		}

		{
			var hit bool
			hit = box2d.PointInPolygon(&box, p1)
			o.Bool("shape:139", hit == true)
			porttest.Ensure(t, hit == true)
			hit = box2d.PointInPolygon(&box, p2)
			o.Bool("shape:141", hit == false)
			porttest.Ensure(t, hit == false)
		}

	})
	t.Run("RayCastShapeTest", func(t *testing.T) {
		defer porttest.Guard(t)

		o := oracle.Open(t, "shape/RayCastShapeTest")
		box := box2d.MakeBox(1, 1)
		circle := box2d.Circle{box2d.Vec2{1, 0}, 1}
		segment := box2d.Segment{box2d.Vec2{0, 1}, box2d.Vec2{0, -1}}

		input := box2d.RayCastInput{box2d.Vec2{-4.0, 0.0}, box2d.Vec2{8.0, 0.0}, 1.0}

		{
			output := box2d.RayCastCircle(&circle, &input)
			o.Bool("shape:155/value0", output.Hit)
			o.Bool("shape:155", output.Hit)
			porttest.Ensure(t, output.Hit)
			o.F32("shape:156/value0", output.Normal.X)
			o.F32("shape:156", output.Normal.X+1.0)
			porttest.EnsureSmall(t, output.Normal.X+1.0, float32(0x1p-23))
			o.F32("shape:157/value0", output.Normal.Y)
			o.F32("shape:157", output.Normal.Y)
			porttest.EnsureSmall(t, output.Normal.Y, float32(0x1p-23))
			o.F32("shape:158/value0", output.Fraction)
			o.F32("shape:158", output.Fraction-0.5)
			porttest.EnsureSmall(t, output.Fraction-0.5, float32(0x1p-23))
		}

		{
			output := box2d.RayCastPolygon(&box, &input)
			o.Bool("shape:163/value0", output.Hit)
			o.Bool("shape:163", output.Hit)
			porttest.Ensure(t, output.Hit)
			o.F32("shape:164/value0", output.Normal.X)
			o.F32("shape:164", output.Normal.X+1.0)
			porttest.EnsureSmall(t, output.Normal.X+1.0, float32(0x1p-23))
			o.F32("shape:165/value0", output.Normal.Y)
			o.F32("shape:165", output.Normal.Y)
			porttest.EnsureSmall(t, output.Normal.Y, float32(0x1p-23))
			o.F32("shape:166/value0", output.Fraction)
			o.F32("shape:166", output.Fraction-3.0/8.0)
			porttest.EnsureSmall(t, output.Fraction-3.0/8.0, float32(0x1p-23))
		}

		{
			output := box2d.RayCastSegment(&segment, &input, true)
			o.Bool("shape:171/value0", output.Hit)
			o.Bool("shape:171", output.Hit)
			porttest.Ensure(t, output.Hit)
			o.F32("shape:172/value0", output.Normal.X)
			o.F32("shape:172", output.Normal.X+1.0)
			porttest.EnsureSmall(t, output.Normal.X+1.0, float32(0x1p-23))
			o.F32("shape:173/value0", output.Normal.Y)
			o.F32("shape:173", output.Normal.Y)
			porttest.EnsureSmall(t, output.Normal.Y, float32(0x1p-23))
			o.F32("shape:174/value0", output.Fraction)
			o.F32("shape:174", output.Fraction-0.5)
			porttest.EnsureSmall(t, output.Fraction-0.5, float32(0x1p-23))
		}

	})
}
