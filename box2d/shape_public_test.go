// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

// These standalone tables supplement, rather than replace, the four pinned Shape cases.
func shapePublicPolygon(o *oracle.Reader, key string, p box2d.Polygon) {
	o.I32(key+"/count", int32(p.Count))
	o.Vec2(key+"/centroid", p.Centroid.X, p.Centroid.Y)
	o.F32(key+"/radius", p.Radius)

	for i := range p.Count {
		o.Vec2(fmt.Sprintf("%s/vertex/%d", key, i), p.Vertices[i].X, p.Vertices[i].Y)
		o.Vec2(fmt.Sprintf("%s/normal/%d", key, i), p.Normals[i].X, p.Normals[i].Y)
	}
}

func shapePublicMass(o *oracle.Reader, key string, m box2d.MassData) {
	o.F32(key+"/mass", m.Mass)
	o.Vec2(key+"/center", m.Center.X, m.Center.Y)
	o.F32(key+"/inertia", m.RotationalInertia)
}

func shapePublicAABB(o *oracle.Reader, key string, a box2d.AABB) {
	o.Vec2(key+"/lower", a.LowerBound.X, a.LowerBound.Y)
	o.Vec2(key+"/upper", a.UpperBound.X, a.UpperBound.Y)
}

func shapePublicCast(o *oracle.Reader, key string, c box2d.CastOutput) {
	// The pinned C functions initialize the complete payload, including on misses.
	o.Bool(key+"/hit", c.Hit)
	o.Vec2(key+"/point", c.Point.X, c.Point.Y)
	o.Vec2(key+"/normal", c.Normal.X, c.Normal.Y)
	o.F32(key+"/fraction", c.Fraction)
	o.I32(key+"/iterations", int32(c.Iterations))
}

func TestRayCastLongTranslationRange(t *testing.T) {
	unitCircle := box2d.Circle{Radius: 1}
	circle := box2d.Circle{Radius: 5}
	capsule := box2d.Capsule{Center1: box2d.Vec2{X: -4}, Center2: box2d.Vec2{X: 4}, Radius: 1}

	shapes := []struct {
		name  string
		cast  func(*box2d.RayCastInput) box2d.CastOutput
		input box2d.RayCastInput
		want  box2d.CastOutput
	}{
		{
			name: "circle_squared_translation_overflow",
			cast: func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCircle(&unitCircle, input) },
			input: box2d.RayCastInput{
				Origin:      box2d.Vec2{X: -2},
				Translation: box2d.Vec2{X: 0x1p64},
			},
			want: box2d.CastOutput{
				Hit:      true,
				Point:    box2d.Vec2{X: -1},
				Normal:   box2d.Vec2{X: -1},
				Fraction: 0x1p-64,
			},
		},
		{
			name: "circle",
			cast: func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCircle(&circle, input) },
			input: box2d.RayCastInput{
				Origin:      box2d.Vec2{X: -8, Y: -6},
				Translation: box2d.Vec2{X: 0x1.cp127, Y: 0x1.5p127},
			},
			want: box2d.CastOutput{
				Hit:      true,
				Point:    box2d.Vec2{X: -4, Y: -3},
				Normal:   box2d.Vec2{X: -0.8, Y: -0.6},
				Fraction: 0x1p-123 / 7,
			},
		},
		{
			name: "capsule",
			cast: func(input *box2d.RayCastInput) box2d.CastOutput { return box2d.RayCastCapsule(&capsule, input) },
			input: box2d.RayCastInput{
				Origin:      box2d.Vec2{X: -1.5, Y: -3},
				Translation: box2d.Vec2{X: 0x1.5p127, Y: 0x1.cp127},
			},
			want: box2d.CastOutput{
				Hit:      true,
				Point:    box2d.Vec2{Y: -1},
				Normal:   box2d.Vec2{Y: -1},
				Fraction: 0x1p-124 / 7,
			},
		},
	}
	for _, shape := range shapes {
		t.Run(shape.name, func(t *testing.T) {
			cases := []struct {
				name        string
				maxFraction float32
				want        box2d.CastOutput
			}{
				{"hit", 1, shape.want},
				{"zero_limit", 0, box2d.CastOutput{}},
				{"before_hit", 0x1p-125 / 7, box2d.CastOutput{}},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					input := shape.input
					input.MaxFraction = tc.maxFraction

					if !input.IsValid() {
						t.Fatal("finite long ray is not valid")
					}

					got := shape.cast(&input)
					if got.Hit != tc.want.Hit || got.Point != tc.want.Point ||
						got.Normal != tc.want.Normal ||
						got.Fraction != tc.want.Fraction {
						t.Fatalf("long ray cast = %+v, want %+v", got, tc.want)
					}
				})
			}
		})
	}
}

func TestHullRetainsTriangleWithOverflowingSquaredEdges(t *testing.T) {
	points := []box2d.Vec2{{}, {X: 0x1p64}, {Y: 0x1p64}}

	hull := box2d.ComputeHull(points)
	if hull.Count != 3 {
		t.Fatalf("hull count = %d, want 3", hull.Count)
	}

	want := map[box2d.Vec2]bool{{}: true, {X: 0x1p64}: true, {Y: 0x1p64}: true}
	for _, point := range hull.Points[:hull.Count] {
		if !want[point] {
			t.Errorf("unexpected or duplicate hull vertex %v", point)
		}

		delete(want, point)
	}

	if len(want) != 0 {
		t.Errorf("missing hull vertices: %v", want)
	}
}

func TestShapePublicHullAndConstructors(t *testing.T) {
	cases := []struct {
		name   string
		count  int
		points []box2d.Vec2
	}{
		{"count0", 0, []box2d.Vec2{{0, 0}}},
		{"count1", 1, []box2d.Vec2{{1, 2}}},
		{"count2", 2, []box2d.Vec2{{-1, 0}, {1, 0}}},
		{"count3", 3, []box2d.Vec2{{0, 2}, {-2, -1}, {2, -1}}},
		{
			"count8",
			8,
			[]box2d.Vec2{{2, 1}, {-1, -2}, {-2, 1}, {1, 2}, {1, -2}, {-2, -1}, {-1, 2}, {2, -1}},
		},
		{
			"count9",
			9,
			[]box2d.Vec2{
				{2, 1},
				{-1, -2},
				{-2, 1},
				{1, 2},
				{1, -2},
				{-2, -1},
				{-1, 2},
				{2, -1},
				{0, 0},
			},
		},
		{"duplicates", 6, []box2d.Vec2{{-1, -1}, {1, 1}, {-1, 1}, {1, -1}, {1, 1}, {-1, -1}}},
		{"near_coincident", 5, []box2d.Vec2{{-1, -1}, {1, 1}, {-1, 1}, {1, -1}, {1.001, 1.001}}},
		{"all_coincident", 3, []box2d.Vec2{{1, 1}, {1.001, 1}, {1, 1.001}}},
		{"collinear", 5, []box2d.Vec2{{2, 2}, {-2, -2}, {0, 0}, {1, 1}, {-1, -1}}},
		{"collinear_edge", 5, []box2d.Vec2{{-1, -1}, {0, -1}, {1, -1}, {1, 1}, {-1, 1}}},
		{"unordered_convex", 6, []box2d.Vec2{{3, 0}, {-2, 1}, {1, -2}, {1, 2}, {-2, -1}, {0, 0}}},
	}
	q := box2d.Rot{C: 0.6, S: 0.8}
	offset := box2d.Vec2{2.5, -3}
	xf := box2d.Transform{P: box2d.Vec2{-4, 1.5}, Q: box2d.Rot{C: 0, S: 1}}

	o := oracle.Open(t, "shape_public/HullAndConstructors")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := box2d.ComputeHull(tc.points[:tc.count])
			o.I32(tc.name+"/count", int32(h.Count))

			for i := range h.Count {
				o.Vec2(fmt.Sprintf("%s/point/%d", tc.name, i), h.Points[i].X, h.Points[i].Y)
			}

			valid := box2d.ValidateHull(&h)
			o.Bool(tc.name+"/valid", valid)
			// Invalid hulls are not legal inputs to any MakePolygon constructor.
			if valid {
				plain := box2d.MakePolygon(&h, 0)
				rounded := box2d.MakePolygon(&h, 0.25)

				shapePublicPolygon(o, tc.name+"/plain", plain)
				shapePublicPolygon(o, tc.name+"/rounded", rounded)
				shapePublicPolygon(o, tc.name+"/offset", box2d.MakeOffsetPolygon(&h, offset, q))
				shapePublicPolygon(
					o,
					tc.name+"/offset_rounded",
					box2d.MakeOffsetRoundedPolygon(&h, offset, q, 0.25),
				)
				shapePublicPolygon(o, tc.name+"/transformed", box2d.TransformPolygon(xf, &rounded))
			}
		})
	}

	shapePublicPolygon(o, "square", box2d.MakeSquare(1.25))
	shapePublicPolygon(o, "box", box2d.MakeBox(1.5, 0.75))

	roundedBox := box2d.MakeRoundedBox(1.5, 0.75, 0.2)
	shapePublicPolygon(o, "rounded_box", roundedBox)
	shapePublicPolygon(o, "offset_box", box2d.MakeOffsetBox(1.5, 0.75, offset, q))
	shapePublicPolygon(
		o,
		"offset_rounded_box",
		box2d.MakeOffsetRoundedBox(1.5, 0.75, offset, q, 0.2),
	)
	shapePublicPolygon(o, "transformed_box", box2d.TransformPolygon(xf, &roundedBox))
}

func TestShapePublicMass(t *testing.T) {
	circles := []box2d.Circle{
		{Center: box2d.Vec2{0, 0}, Radius: 1},
		{Center: box2d.Vec2{2, -3}, Radius: 0.75},
		{Center: box2d.Vec2{-4, 2}, Radius: 0},
	}
	capsules := []box2d.Capsule{
		{Center1: box2d.Vec2{-1, 0}, Center2: box2d.Vec2{1, 0}, Radius: 0.5},
		{Center1: box2d.Vec2{2, -3}, Center2: box2d.Vec2{5, 1}, Radius: 0.75},
		{Center1: box2d.Vec2{1, 2}, Center2: box2d.Vec2{1, 2}, Radius: 0.5},
	}

	h := box2d.ComputeHull([]box2d.Vec2{{0, 2}, {-2, -1}, {2, -1}})
	if !box2d.ValidateHull(&h) {
		t.Fatal("mass input triangle hull is invalid")
	}

	polygons := []box2d.Polygon{
		box2d.MakeBox(1.5, 0.75), box2d.MakeRoundedBox(1.5, 0.75, 0.25),
		box2d.MakeOffsetRoundedBox(1.5, 0.75, box2d.Vec2{2, -3}, box2d.Rot{C: 0.6, S: 0.8}, 0.25),
		box2d.MakePolygon(&h, 0.125),
	}
	densities := []float32{0, 0.75, 2.5}
	o := oracle.Open(t, "shape_public/Mass")

	for d, density := range densities {
		for i := range 3 {
			shapePublicMass(
				o,
				fmt.Sprintf("density/%d/circle/%d", d, i),
				box2d.ComputeCircleMass(&circles[i], density),
			)
			shapePublicMass(
				o,
				fmt.Sprintf("density/%d/capsule/%d", d, i),
				box2d.ComputeCapsuleMass(&capsules[i], density),
			)
		}

		for i := range 4 {
			shapePublicMass(
				o,
				fmt.Sprintf("density/%d/polygon/%d", d, i),
				box2d.ComputePolygonMass(&polygons[i], density),
			)
		}
	}
}

func TestShapePublicAABB(t *testing.T) {
	circle := box2d.Circle{Center: box2d.Vec2{1.25, -2}, Radius: 0.7}
	capsule := box2d.Capsule{
		Center1: box2d.Vec2{-1.25, 0.5},
		Center2: box2d.Vec2{2, -0.75},
		Radius:  0.3,
	}
	segment := box2d.Segment{Point1: box2d.Vec2{-1.25, 0.5}, Point2: box2d.Vec2{2, -0.75}}
	polygon := box2d.MakeOffsetRoundedBox(
		1.5,
		0.75,
		box2d.Vec2{0.25, -0.5},
		box2d.Rot{C: 0.6, S: 0.8},
		0.2,
	)
	frames := []struct {
		name string
		xf   box2d.WorldTransform
	}{
		{"identity", box2d.WorldTransform{P: box2d.Pos{0, 0}, Q: box2d.Rot{C: 1, S: 0}}},
		{"translated", box2d.WorldTransform{P: box2d.Pos{7.25, -4.5}, Q: box2d.Rot{C: 1, S: 0}}},
		{"rotated", box2d.WorldTransform{P: box2d.Pos{7.25, -4.5}, Q: box2d.Rot{C: 0.6, S: 0.8}}},
		{
			"distant",
			box2d.WorldTransform{P: box2d.Pos{10000000, -10000000}, Q: box2d.Rot{C: 0.6, S: 0.8}},
		},
	}

	o := oracle.Open(t, "shape_public/AABB")
	for _, frame := range frames {
		shapePublicAABB(o, frame.name+"/circle", box2d.ComputeCircleAABB(&circle, frame.xf))
		shapePublicAABB(o, frame.name+"/capsule", box2d.ComputeCapsuleAABB(&capsule, frame.xf))
		shapePublicAABB(o, frame.name+"/segment", box2d.ComputeSegmentAABB(&segment, frame.xf))
		shapePublicAABB(o, frame.name+"/polygon", box2d.ComputePolygonAABB(&polygon, frame.xf))
	}
}

func TestShapePublicContainment(t *testing.T) {
	circle := box2d.Circle{Center: box2d.Vec2{2, -1}, Radius: 1}
	capsule := box2d.Capsule{Center1: box2d.Vec2{-1, 0}, Center2: box2d.Vec2{1, 0}, Radius: 0.5}
	degenerate := box2d.Capsule{Center1: box2d.Vec2{2, -1}, Center2: box2d.Vec2{2, -1}, Radius: 1}
	box := box2d.MakeBox(1, 1)
	rounded := box2d.MakeRoundedBox(1, 1, 0.5)
	offset := box2d.MakeOffsetBox(1, 1, box2d.Vec2{2, -1}, box2d.Rot{C: 0, S: 1})
	circlePoints := []box2d.Vec2{{2, -1}, {4, -1}, {3, -1}, {2, 0}, {2.999, -1}, {3.001, -1}}
	capsulePoints := []box2d.Vec2{{0, 0}, {0, 2}, {0, 0.5}, {-1.5, 0}, {1.5, 0}, {1.501, 0}}
	polygonPoints := []box2d.Vec2{{0, 0}, {3, 0}, {1, 0}, {1, 1}, {1.5, 0}, {1.501, 0}, {1.5, 1.5}}

	o := oracle.Open(t, "shape_public/Containment")
	for i := range 6 {
		o.Bool(fmt.Sprintf("circle/%d", i), box2d.PointInCircle(&circle, circlePoints[i]))
		o.Bool(fmt.Sprintf("capsule/%d", i), box2d.PointInCapsule(&capsule, capsulePoints[i]))
		o.Bool(
			fmt.Sprintf("degenerate_capsule/%d", i),
			box2d.PointInCapsule(&degenerate, circlePoints[i]),
		)
		o.Bool(fmt.Sprintf("offset_polygon/%d", i), box2d.PointInPolygon(&offset, circlePoints[i]))
	}

	for i := range 7 {
		o.Bool(fmt.Sprintf("polygon/%d", i), box2d.PointInPolygon(&box, polygonPoints[i]))
		o.Bool(
			fmt.Sprintf("rounded_polygon/%d", i),
			box2d.PointInPolygon(&rounded, polygonPoints[i]),
		)
	}
}

func TestShapePublicRayCasts(t *testing.T) {
	circle := box2d.Circle{Center: box2d.Vec2{0, 0}, Radius: 1}
	capsule := box2d.Capsule{Center1: box2d.Vec2{-1, 0}, Center2: box2d.Vec2{1, 0}, Radius: 1}
	degenerate := box2d.Capsule{Center1: box2d.Vec2{0, 0}, Center2: box2d.Vec2{0, 0}, Radius: 1}
	segment := box2d.Segment{Point1: box2d.Vec2{0, -1}, Point2: box2d.Vec2{0, 1}}
	reversed := box2d.Segment{Point1: box2d.Vec2{0, 1}, Point2: box2d.Vec2{0, -1}}
	polygon := box2d.MakeBox(1, 1)
	rounded := box2d.MakeRoundedBox(1, 1, 0.25)
	cases := []struct {
		name  string
		input box2d.RayCastInput
	}{
		{
			"hit",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 0},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 1,
			},
		},
		{
			"reverse",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{4, 0},
				Translation: box2d.Vec2{-8, 0},
				MaxFraction: 1,
			},
		},
		{
			"miss",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 4},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 1,
			},
		},
		{
			"away",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 0},
				Translation: box2d.Vec2{-8, 0},
				MaxFraction: 1,
			},
		},
		{
			"overlap",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{0, 0},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 1,
			},
		},
		{
			"boundary",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-1, 0},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 1,
			},
		},
		{
			"graze",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 1},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 1,
			},
		},
		{
			"max_fraction",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 0},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 0.2,
			},
		},
		{
			"endpoint",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 0},
				Translation: box2d.Vec2{8, 0},
				MaxFraction: 0.5,
			},
		},
		{
			"zero_translation",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{-4, 0},
				Translation: box2d.Vec2{0, 0},
				MaxFraction: 1,
			},
		},
		{
			"capsule_cap",
			box2d.RayCastInput{
				Origin:      box2d.Vec2{1, 4},
				Translation: box2d.Vec2{0, -8},
				MaxFraction: 1,
			},
		},
	}

	o := oracle.Open(t, "shape_public/RayCasts")
	for _, tc := range cases {
		shapePublicCast(o, tc.name+"/circle", box2d.RayCastCircle(&circle, &tc.input))
		shapePublicCast(o, tc.name+"/capsule", box2d.RayCastCapsule(&capsule, &tc.input))
		shapePublicCast(
			o,
			tc.name+"/degenerate_capsule",
			box2d.RayCastCapsule(&degenerate, &tc.input),
		)
		shapePublicCast(o, tc.name+"/polygon", box2d.RayCastPolygon(&polygon, &tc.input))
		shapePublicCast(o, tc.name+"/rounded_polygon", box2d.RayCastPolygon(&rounded, &tc.input))
		shapePublicCast(
			o,
			tc.name+"/segment_two_sided",
			box2d.RayCastSegment(&segment, &tc.input, false),
		)
		shapePublicCast(
			o,
			tc.name+"/segment_one_sided",
			box2d.RayCastSegment(&segment, &tc.input, true),
		)
		shapePublicCast(
			o,
			tc.name+"/reversed_two_sided",
			box2d.RayCastSegment(&reversed, &tc.input, false),
		)
		shapePublicCast(
			o,
			tc.name+"/reversed_one_sided",
			box2d.RayCastSegment(&reversed, &tc.input, true),
		)
	}
}

func TestShapePublicShapeCasts(t *testing.T) {
	circle := box2d.Circle{Center: box2d.Vec2{0, 0}, Radius: 1}
	capsule := box2d.Capsule{Center1: box2d.Vec2{-1, 0}, Center2: box2d.Vec2{1, 0}, Radius: 1}
	segment := box2d.Segment{Point1: box2d.Vec2{0, -1}, Point2: box2d.Vec2{0, 1}}
	polygon := box2d.MakeBox(1, 1)
	rounded := box2d.MakeRoundedBox(1, 1, 0.25)
	cases := []struct {
		name                string
		origin, translation box2d.Vec2
		radius, maxFraction float32
		encroach            bool
	}{
		{"hit", box2d.Vec2{-4, 0}, box2d.Vec2{8, 0}, 0.25, 1, false},
		{"reverse", box2d.Vec2{4, 0}, box2d.Vec2{-8, 0}, 0.25, 1, false},
		{"miss", box2d.Vec2{-4, 4}, box2d.Vec2{8, 0}, 0.25, 1, false},
		{"away", box2d.Vec2{-4, 0}, box2d.Vec2{-8, 0}, 0.25, 1, false},
		{"overlap", box2d.Vec2{0, 0}, box2d.Vec2{8, 0}, 0.25, 1, false},
		{"graze", box2d.Vec2{-4, 1.25}, box2d.Vec2{8, 0}, 0.25, 1, false},
		{"max_fraction", box2d.Vec2{-4, 0}, box2d.Vec2{8, 0}, 0.25, 0.2, false},
		{"endpoint", box2d.Vec2{-4, 0}, box2d.Vec2{8, 0}, 0.25, 0.5, false},
		{"zero_translation", box2d.Vec2{-4, 0}, box2d.Vec2{0, 0}, 0.25, 1, false},
		{"zero_radius", box2d.Vec2{-4, 0}, box2d.Vec2{8, 0}, 0, 1, false},
	}
	o := oracle.Open(t, "shape_public/ShapeCasts")

	for _, tc := range cases {
		in := box2d.ShapeCastInput{
			Proxy:       box2d.MakeProxy([]box2d.Vec2{tc.origin}, tc.radius),
			Translation: tc.translation, MaxFraction: tc.maxFraction, CanEncroach: tc.encroach,
		}
		shapePublicCast(o, tc.name+"/circle", box2d.ShapeCastCircle(&circle, &in))
		shapePublicCast(o, tc.name+"/capsule", box2d.ShapeCastCapsule(&capsule, &in))
		shapePublicCast(o, tc.name+"/segment", box2d.ShapeCastSegment(&segment, &in))
		shapePublicCast(o, tc.name+"/polygon", box2d.ShapeCastPolygon(&polygon, &in))
		shapePublicCast(o, tc.name+"/rounded_polygon", box2d.ShapeCastPolygon(&rounded, &in))
	}
	// Shape-specific touching origins keep every target initially on its radius boundary.
	touching := []box2d.Vec2{{-1.25, 0}, {-2.25, 0}, {-0.25, 0}, {-1.25, 0}, {-1.5, 0}}

	for e := range 2 {
		in := box2d.ShapeCastInput{
			Translation: box2d.Vec2{2, 0},
			MaxFraction: 1,
			CanEncroach: e != 0,
		}
		in.Proxy = box2d.MakeProxy(touching[0:1], 0.25)
		shapePublicCast(o, fmt.Sprintf("touch/%d/circle", e), box2d.ShapeCastCircle(&circle, &in))
		in.Proxy = box2d.MakeProxy(touching[1:2], 0.25)
		shapePublicCast(
			o,
			fmt.Sprintf("touch/%d/capsule", e),
			box2d.ShapeCastCapsule(&capsule, &in),
		)
		in.Proxy = box2d.MakeProxy(touching[2:3], 0.25)
		shapePublicCast(
			o,
			fmt.Sprintf("touch/%d/segment", e),
			box2d.ShapeCastSegment(&segment, &in),
		)
		in.Proxy = box2d.MakeProxy(touching[3:4], 0.25)
		shapePublicCast(
			o,
			fmt.Sprintf("touch/%d/polygon", e),
			box2d.ShapeCastPolygon(&polygon, &in),
		)
		in.Proxy = box2d.MakeProxy(touching[4:5], 0.25)
		shapePublicCast(
			o,
			fmt.Sprintf("touch/%d/rounded_polygon", e),
			box2d.ShapeCastPolygon(&rounded, &in),
		)
	}

	capsulePoints := []box2d.Vec2{{-4, -0.5}, {-4, 0.5}}
	boxPoints := []box2d.Vec2{{-4.5, -0.5}, {-3.5, -0.5}, {-3.5, 0.5}, {-4.5, 0.5}}

	proxies := []box2d.ShapeProxy{
		box2d.MakeProxy(capsulePoints, 0.25),
		box2d.MakeProxy(boxPoints, 0),
	}
	for i, proxy := range proxies {
		in := box2d.ShapeCastInput{Proxy: proxy, Translation: box2d.Vec2{8, 0}, MaxFraction: 1}
		shapePublicCast(o, fmt.Sprintf("proxy/%d/circle", i), box2d.ShapeCastCircle(&circle, &in))
		shapePublicCast(
			o,
			fmt.Sprintf("proxy/%d/capsule", i),
			box2d.ShapeCastCapsule(&capsule, &in),
		)
		shapePublicCast(
			o,
			fmt.Sprintf("proxy/%d/segment", i),
			box2d.ShapeCastSegment(&segment, &in),
		)
		shapePublicCast(
			o,
			fmt.Sprintf("proxy/%d/polygon", i),
			box2d.ShapeCastPolygon(&polygon, &in),
		)
		shapePublicCast(
			o,
			fmt.Sprintf("proxy/%d/rounded_polygon", i),
			box2d.ShapeCastPolygon(&rounded, &in),
		)
	}
}
