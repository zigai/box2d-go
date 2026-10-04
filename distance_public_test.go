package box2d_test

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func distancePublicBox(radius float32) box2d.ShapeProxy {
	return box2d.MakeProxy([]box2d.Vec2{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}, radius)
}

func distancePublicProxy(r *oracle.Reader, key string, p box2d.ShapeProxy) {
	r.I32(key+"/count", int32(p.Count))
	r.F32(key+"/radius", p.Radius)
	for i := range p.Count {
		r.Vec2(fmt.Sprintf("%s/point/%d", key, i), p.Points[i].X, p.Points[i].Y)
	}
}

func distancePublicOutput(r *oracle.Reader, key string, o box2d.DistanceOutput, cache box2d.SimplexCache, debug []box2d.Simplex) {
	r.Vec2(key+"/pointA", o.PointA.X, o.PointA.Y)
	r.Vec2(key+"/pointB", o.PointB.X, o.PointB.Y)
	r.F32(key+"/distance", o.Distance)
	// The public contract declares the normal invalid at zero distance.
	if o.Distance > 0 {
		r.Vec2(key+"/normal", o.Normal.X, o.Normal.Y)
	}
	r.I32(key+"/iterations", int32(o.Iterations))
	r.I32(key+"/simplexCount", int32(o.SimplexCount))
	r.U32(key+"/cache/count", uint32(cache.Count))
	for i := range int(cache.Count) {
		r.U32(fmt.Sprintf("%s/cache/%d/indexA", key, i), uint32(cache.IndexA[i]))
		r.U32(fmt.Sprintf("%s/cache/%d/indexB", key, i), uint32(cache.IndexB[i]))
	}
	for i := range o.SimplexCount {
		s := debug[i]
		k := fmt.Sprintf("%s/simplex/%d", key, i)
		r.I32(k+"/count", int32(s.Count))
		vertices := [3]box2d.SimplexVertex{s.V1, s.V2, s.V3}
		for j := range s.Count {
			v := vertices[j]
			vk := fmt.Sprintf("%s/vertex/%d", k, j)
			r.Vec2(vk+"/wA", v.WA.X, v.WA.Y)
			r.Vec2(vk+"/wB", v.WB.X, v.WB.Y)
			r.Vec2(vk+"/w", v.W.X, v.W.Y)
			r.F32(vk+"/a", v.A)
			r.I32(vk+"/indexA", int32(v.IndexA))
			r.I32(vk+"/indexB", int32(v.IndexB))
		}
	}
}

func distancePublicCast(r *oracle.Reader, o box2d.CastOutput) {
	r.Bool("hit", o.Hit)
	r.Vec2("point", o.Point.X, o.Point.Y)
	r.Vec2("normal", o.Normal.X, o.Normal.Y)
	r.F32("fraction", o.Fraction)
	r.I32("iterations", int32(o.Iterations))
}

func distancePublicTOI(r *oracle.Reader, key string, o box2d.TOIOutput) {
	r.I32(key+"/state", int32(o.State))
	r.F32(key+"/fraction", o.Fraction)
	// C initializes these fields on every exit; non-hit zeros are compared as
	// returned payload, not interpreted as a contact witness.
	r.Vec2(key+"/point", o.Point.X, o.Point.Y)
	r.Vec2(key+"/normal", o.Normal.X, o.Normal.Y)
}

func TestDistancePublic(t *testing.T) {
	t.Run("Segments", func(t *testing.T) {
		cases := []struct {
			name           string
			p1, q1, p2, q2 box2d.Vec2
		}{
			{"intersect", box2d.Vec2{-2, 0}, box2d.Vec2{2, 0}, box2d.Vec2{0, -3}, box2d.Vec2{0, 3}},
			{"parallel", box2d.Vec2{-2, 0}, box2d.Vec2{2, 0}, box2d.Vec2{-1, 2}, box2d.Vec2{3, 2}},
			{"parallel-reversed", box2d.Vec2{-2, 0}, box2d.Vec2{2, 0}, box2d.Vec2{3, 2}, box2d.Vec2{-1, 2}},
			{"clamp-start", box2d.Vec2{0, 0}, box2d.Vec2{1, 0}, box2d.Vec2{-2, 1}, box2d.Vec2{-1, 2}},
			{"clamp-end", box2d.Vec2{0, 0}, box2d.Vec2{1, 0}, box2d.Vec2{3, 2}, box2d.Vec2{2, 1}},
			{"first-point", box2d.Vec2{1, 1}, box2d.Vec2{1, 1}, box2d.Vec2{-2, 0}, box2d.Vec2{2, 0}},
			{"second-point", box2d.Vec2{-2, 0}, box2d.Vec2{2, 0}, box2d.Vec2{1, 1}, box2d.Vec2{1, 1}},
			{"both-points", box2d.Vec2{1, 1}, box2d.Vec2{1, 1}, box2d.Vec2{-2, 3}, box2d.Vec2{-2, 3}},
			{"tiny-first", box2d.Vec2{0, 0}, box2d.Vec2{0x1p-25, 0}, box2d.Vec2{-1, 2}, box2d.Vec2{1, 2}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				r := oracle.Open(t, "distance_public/segments/"+c.name)
				o := box2d.SegmentDistance(c.p1, c.q1, c.p2, c.q2)
				r.Vec2("closest1", o.Closest1.X, o.Closest1.Y)
				r.Vec2("closest2", o.Closest2.X, o.Closest2.Y)
				r.F32("fraction1", o.Fraction1)
				r.F32("fraction2", o.Fraction2)
				r.F32("distanceSquared", o.DistanceSquared)
			})
		}
	})

	t.Run("ProxyCopies", func(t *testing.T) {
		for _, count := range []int{1, 8, 9} {
			t.Run(fmt.Sprint(count), func(t *testing.T) {
				r := oracle.Open(t, fmt.Sprintf("distance_public/proxy/%d", count))
				points := []box2d.Vec2{{-2, -1}, {0, -2}, {2, -1}, {3, 0}, {2, 1}, {0, 2}, {-2, 1}, {-3, 0}, {9, 9}}
				p := box2d.MakeProxy(points[:count], 0.375)
				o := box2d.MakeOffsetProxy(points[:count], 0.625, box2d.Vec2{3, -4}, box2d.Rot{0.6, 0.8})
				distancePublicProxy(r, "plain", p)
				distancePublicProxy(r, "offset", o)
				for i := range count {
					points[i] = box2d.Vec2{100 + float32(i), -100 - float32(i)}
				}
				distancePublicProxy(r, "plain-after-source-mutation", p)
				distancePublicProxy(r, "offset-after-source-mutation", o)
			})
		}
	})

	t.Run("SweepTransforms", func(t *testing.T) {
		r := oracle.Open(t, "distance_public/sweep-transforms")
		sweeps := []box2d.Sweep{
			{LocalCenter: box2d.Vec2{2, -3}, C1: box2d.Vec2{-5, 4}, C2: box2d.Vec2{7, -2}, Q1: box2d.RotIdentity, Q2: box2d.Rot{0, 1}},
			{LocalCenter: box2d.Vec2{-0.5, 1.25}, C1: box2d.Vec2{3, 2}, C2: box2d.Vec2{3, 2}, Q1: box2d.Rot{0.6, -0.8}, Q2: box2d.Rot{0.6, 0.8}},
		}
		for i, sweep := range sweeps {
			for j, time := range []float32{0, 0.25, 0.5, 0.75, 1} {
				xf := sweep.Transform(time)
				key := fmt.Sprintf("%d/%d", i, j)
				r.Vec2(key+"/p", xf.P.X, xf.P.Y)
				r.Rot(key+"/q", xf.Q.C, xf.Q.S)
			}
		}
	})

	t.Run("ShapeDistance", func(t *testing.T) {
		segment := box2d.MakeProxy([]box2d.Vec2{{0, -1}, {0, 1}}, 0)
		point := box2d.MakeProxy([]box2d.Vec2{{0, 0}}, 0.5)
		cases := []struct {
			name  string
			a, b  box2d.ShapeProxy
			xf    box2d.Transform
			radii bool
		}{
			{"separated", distancePublicBox(0), segment, box2d.Transform{P: box2d.Vec2{3, 0}, Q: box2d.RotIdentity}, false},
			{"touch", distancePublicBox(0), segment, box2d.Transform{P: box2d.Vec2{1, 0}, Q: box2d.RotIdentity}, false},
			{"overlap", distancePublicBox(0), distancePublicBox(0), box2d.Transform{P: box2d.Vec2{0.25, 0.5}, Q: box2d.RotIdentity}, false},
			{"radii-separated", point, point, box2d.Transform{P: box2d.Vec2{3, 0}, Q: box2d.RotIdentity}, true},
			{"radii-touch", point, point, box2d.Transform{P: box2d.Vec2{1, 0}, Q: box2d.RotIdentity}, true},
			{"radii-overlap", point, point, box2d.Transform{P: box2d.Vec2{0.75, 0}, Q: box2d.RotIdentity}, true},
			{"radii-ignored", point, point, box2d.Transform{P: box2d.Vec2{0.75, 0}, Q: box2d.RotIdentity}, false},
			{"transformed", distancePublicBox(0.125), segment, box2d.Transform{P: box2d.Vec2{3, -0.5}, Q: box2d.Rot{0.6, 0.8}}, true},
			{"max-count", box2d.MakeProxy([]box2d.Vec2{{-2, -1}, {0, -2}, {2, -1}, {3, 0}, {2, 1}, {0, 2}, {-2, 1}, {-3, 0}}, 0.25), point, box2d.Transform{P: box2d.Vec2{4, 2}, Q: box2d.RotIdentity}, true},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				r := oracle.Open(t, "distance_public/distance/"+c.name)
				input := box2d.DistanceInput{ProxyA: c.a, ProxyB: c.b, Transform: c.xf, UseRadii: c.radii}
				for _, capacity := range []int{0, 1, 24} {
					cache := box2d.SimplexCache{}
					var debug []box2d.Simplex
					if capacity != 0 {
						debug = make([]box2d.Simplex, capacity)
					}
					o := box2d.ShapeDistance(&input, &cache, debug)
					distancePublicOutput(r, fmt.Sprintf("capacity/%d", capacity), o, cache, debug)
				}
			})
		}
	})

	t.Run("CacheHistory", func(t *testing.T) {
		r := oracle.Open(t, "distance_public/cache-history")
		input := box2d.DistanceInput{ProxyA: distancePublicBox(0.125), ProxyB: distancePublicBox(0.25), UseRadii: true}
		transforms := []box2d.Transform{
			{P: box2d.Vec2{4, 0.25}, Q: box2d.RotIdentity},
			{P: box2d.Vec2{3.75, 0.5}, Q: box2d.RotIdentity},
			{P: box2d.Vec2{3.5, 1}, Q: box2d.Rot{0.6, 0.8}},
			{P: box2d.Vec2{2, 1}, Q: box2d.Rot{0.6, 0.8}},
			{P: box2d.Vec2{0.5, 0.25}, Q: box2d.RotIdentity},
			{P: box2d.Vec2{-3, -1}, Q: box2d.Rot{0, -1}},
			{P: box2d.Vec2{-3.25, -0.75}, Q: box2d.Rot{0, -1}},
			{P: box2d.Vec2{4, 0.25}, Q: box2d.RotIdentity},
		}
		cache := box2d.SimplexCache{}
		for i, xf := range transforms {
			input.Transform = xf
			debug := make([]box2d.Simplex, 24)
			o := box2d.ShapeDistance(&input, &cache, debug)
			distancePublicOutput(r, fmt.Sprint(i), o, cache, debug)
		}
	})

	t.Run("ShapeCast", func(t *testing.T) {
		point := box2d.MakeProxy([]box2d.Vec2{{0, 0}}, 0.5)
		segment := box2d.MakeProxy([]box2d.Vec2{{0, -1}, {0, 1}}, 0)
		cases := []struct {
			name     string
			a, b     box2d.ShapeProxy
			p        box2d.Vec2
			q        box2d.Rot
			delta    box2d.Vec2
			max      float32
			encroach bool
		}{
			{"hit", distancePublicBox(0), segment, box2d.Vec2{3, 0}, box2d.RotIdentity, box2d.Vec2{-4, 0}, 1, false},
			{"away", distancePublicBox(0), segment, box2d.Vec2{3, 0}, box2d.RotIdentity, box2d.Vec2{4, 0}, 1, false},
			{"stationary", distancePublicBox(0), segment, box2d.Vec2{3, 0}, box2d.RotIdentity, box2d.Vec2{}, 1, false},
			{"too-short", distancePublicBox(0), segment, box2d.Vec2{3, 0}, box2d.RotIdentity, box2d.Vec2{-4, 0}, 0.25, false},
			{"fraction-boundary", distancePublicBox(0), segment, box2d.Vec2{3, 0}, box2d.RotIdentity, box2d.Vec2{-4, 0}, 0.5, false},
			{"initial-touch", distancePublicBox(0), segment, box2d.Vec2{1, 0}, box2d.RotIdentity, box2d.Vec2{-1, 0}, 1, false},
			{"initial-overlap", distancePublicBox(0), distancePublicBox(0), box2d.Vec2{0.25, 0.5}, box2d.RotIdentity, box2d.Vec2{-1, 0}, 1, false},
			{"graze", distancePublicBox(0), box2d.MakeProxy([]box2d.Vec2{{0, 0}}, 0), box2d.Vec2{3, 1}, box2d.RotIdentity, box2d.Vec2{-4, 0}, 1, false},
			{"graze-miss", distancePublicBox(0), box2d.MakeProxy([]box2d.Vec2{{0, 0}}, 0), box2d.Vec2{3, 1.25}, box2d.RotIdentity, box2d.Vec2{-4, 0}, 1, false},
			{"rounded-hit", point, point, box2d.Vec2{3, 0.5}, box2d.RotIdentity, box2d.Vec2{-4, 0}, 1, false},
			{"rounded-touch", point, point, box2d.Vec2{1, 0}, box2d.RotIdentity, box2d.Vec2{-1, 0}, 1, false},
			{"encroach-off", point, point, box2d.Vec2{0.99, 0}, box2d.RotIdentity, box2d.Vec2{-1, 0}, 1, false},
			{"encroach-on", point, point, box2d.Vec2{0.99, 0}, box2d.RotIdentity, box2d.Vec2{-1, 0}, 1, true},
			{"encroach-away", point, point, box2d.Vec2{0.99, 0}, box2d.RotIdentity, box2d.Vec2{1, 0}, 1, true},
			{"transformed", distancePublicBox(0.125), segment, box2d.Vec2{4, 0.5}, box2d.Rot{0.6, 0.8}, box2d.Vec2{-5, -0.5}, 1, false},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				r := oracle.Open(t, "distance_public/cast/"+c.name)
				input := box2d.ShapeCastPairInput{ProxyA: c.a, ProxyB: c.b, Transform: box2d.Transform{P: c.p, Q: c.q}, TranslationB: c.delta, MaxFraction: c.max, CanEncroach: c.encroach}
				distancePublicCast(r, box2d.ShapeCast(&input))
			})
		}
	})

	t.Run("TimeOfImpact", func(t *testing.T) {
		cases := []struct {
			name        string
			start, end  box2d.Vec2
			q1, q2      box2d.Rot
			localCenter box2d.Vec2
			max         float32
		}{
			{"translating-hit", box2d.Vec2{4, 0}, box2d.Vec2{0, 0}, box2d.RotIdentity, box2d.RotIdentity, box2d.Vec2{}, 1},
			{"translating-separated", box2d.Vec2{4, 0}, box2d.Vec2{6, 0}, box2d.RotIdentity, box2d.RotIdentity, box2d.Vec2{}, 1},
			{"initial-overlap", box2d.Vec2{0.25, 0.5}, box2d.Vec2{4, 0}, box2d.RotIdentity, box2d.RotIdentity, box2d.Vec2{}, 1},
			{"initial-touch", box2d.Vec2{2, 0}, box2d.Vec2{0, 0}, box2d.RotIdentity, box2d.RotIdentity, box2d.Vec2{}, 1},
			{"radius-touch", box2d.Vec2{2.25, 0}, box2d.Vec2{0, 0}, box2d.RotIdentity, box2d.RotIdentity, box2d.Vec2{}, 1},
			{"max-fraction", box2d.Vec2{4, 0}, box2d.Vec2{0, 0}, box2d.RotIdentity, box2d.RotIdentity, box2d.Vec2{}, 0.25},
			{"rotating-hit", box2d.Vec2{4, 0.5}, box2d.Vec2{0, 0}, box2d.RotIdentity, box2d.Rot{0, 1}, box2d.Vec2{}, 1},
			{"local-com", box2d.Vec2{5, 0.5}, box2d.Vec2{0, 0}, box2d.Rot{0.6, -0.8}, box2d.Rot{0.6, 0.8}, box2d.Vec2{0.75, -0.25}, 1},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				r := oracle.Open(t, "distance_public/toi/"+c.name)
				input := box2d.TOIInput{ProxyA: distancePublicBox(0.125), ProxyB: distancePublicBox(0.125), SweepA: box2d.Sweep{C1: box2d.Vec2{-1, 2}, C2: box2d.Vec2{-1, 2}, Q1: box2d.RotIdentity, Q2: box2d.RotIdentity}, SweepB: box2d.Sweep{LocalCenter: c.localCenter, C1: c.start.Add(box2d.Vec2{-1, 2}), C2: c.end.Add(box2d.Vec2{-1, 2}), Q1: c.q1, Q2: c.q2}, MaxFraction: c.max}
				distancePublicTOI(r, "output", box2d.TimeOfImpact(&input))
			})
		}
	})

	t.Run("PinnedRotatingTOISample", func(t *testing.T) {
		// The rotating capsule/box input is retained from the pinned Collision /
		// Time of Impact sample, in addition to the synthetic partition table.
		r := oracle.Open(t, "distance_public/toi-pinned-rotating-sample")
		input := box2d.TOIInput{
			ProxyA: box2d.MakeProxy([]box2d.Vec2{{-16.25, 44.75}, {-15.75, 44.75}, {-15.75, 45.25}, {-16.25, 45.25}}, 0),
			ProxyB: box2d.MakeProxy([]box2d.Vec2{{0, -0.125}, {0, 0.125}}, 0.0299999993),
			SweepA: box2d.Sweep{Q1: box2d.RotIdentity, Q2: box2d.RotIdentity},
			SweepB: box2d.Sweep{
				C1: box2d.Vec2{-15.8332710, 45.3520279},
				C2: box2d.Vec2{-15.8324337, 45.3413048},
				Q1: box2d.Rot{-0.540891349, 0.841092527},
				Q2: box2d.Rot{-0.457797021, 0.889056742},
			},
			MaxFraction: 1,
		}
		distancePublicTOI(r, "output", box2d.TimeOfImpact(&input))
	})

	t.Run("RotatingTOIHistory", func(t *testing.T) {
		r := oracle.Open(t, "distance_public/toi-rotating-history")
		// A finite deterministic corpus probes the failed-state branches without
		// pretending that TOI promises to detect every intermediate rotating collision.
		rotations := []box2d.Rot{{1, 0}, {0.6, 0.8}, {0, 1}, {-0.6, 0.8}, {-1, 0}, {-0.6, -0.8}, {0, -1}, {0.6, -0.8}}
		ends := []box2d.Vec2{{-3, -2}, {0, 0}, {3, 2}, {-0.125, 0.25}}
		a := box2d.MakeProxy([]box2d.Vec2{{-2, -0.125}, {2, -0.125}, {2, 0.125}, {-2, 0.125}}, 0.025)
		b := box2d.MakeProxy([]box2d.Vec2{{-0.125, -2}, {0.125, -2}, {0.125, 2}, {-0.125, 2}}, 0.025)
		var states [5]int32
		for i, q := range rotations {
			for j, end := range ends {
				for k, start := range []box2d.Vec2{{3, 0.125}, {0.125, 3}, {2.25, 2.25}, {-3, -0.125}} {
					input := box2d.TOIInput{ProxyA: a, ProxyB: b, SweepA: box2d.Sweep{LocalCenter: box2d.Vec2{0.25, -0.125}, C1: box2d.Vec2{}, C2: box2d.Vec2{0.125, -0.25}, Q1: rotations[(i+1)%len(rotations)], Q2: rotations[(i+2)%len(rotations)]}, SweepB: box2d.Sweep{LocalCenter: box2d.Vec2{-0.125, 0.25}, C1: start, C2: end, Q1: rotations[(i+1)%len(rotations)], Q2: q}, MaxFraction: 1}
					o := box2d.TimeOfImpact(&input)
					distancePublicTOI(r, fmt.Sprintf("%d/%d/%d", i, j, k), o)
					states[o.State]++
				}
			}
		}
		for i := range len(states) {
			r.I32(fmt.Sprintf("state-count/%d", i), states[i])
		}
	})
}
