//go:build box2d_stress

// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

type queryDigest struct{ h uint64 }

func newQueryDigest() queryDigest { return queryDigest{14695981039346656037} }
func (d *queryDigest) bytes(p []byte) {
	for _, v := range p {
		d.h ^= uint64(v)
		d.h *= 1099511628211
	}
}

func (d *queryDigest) u(v uint32) {
	var p [4]byte
	binary.LittleEndian.PutUint32(p[:], v)
	d.bytes(p[:])
}
func (d *queryDigest) f(v float32)      { d.u(math.Float32bits(v)) }
func (d *queryDigest) vec(v box2d.Vec2) { d.f(v.X); d.f(v.Y) }
func (d *queryDigest) pos(v box2d.Pos) {
	if unsafe.Sizeof(v) == 16 {
		var p [16]byte
		binary.LittleEndian.PutUint64(p[:], math.Float64bits(float64(v.X)))
		binary.LittleEndian.PutUint64(p[8:], math.Float64bits(float64(v.Y)))
		d.bytes(p[:])
	} else {
		d.f(float32(v.X))
		d.f(float32(v.Y))
	}
}

func (d *queryDigest) id(id box2d.Shape) {
	v := id.ID()
	d.u(uint32(v >> 32))
	d.u(uint32(v & 0xffff))
}

func (d *queryDigest) stats(s box2d.TreeStats) { d.u(uint32(s.NodeVisits)); d.u(uint32(s.LeafVisits)) }

func (d *queryDigest) hit(r box2d.RayResult) {
	if r.Hit {
		d.u(1)
		d.id(r.Shape)
		d.pos(r.Point)
		d.vec(r.Normal)
		d.f(r.Fraction)
	} else {
		d.u(0)
	}

	d.u(uint32(r.NodeVisits))
	d.u(uint32(r.LeafVisits))
}

func (d *queryDigest) overlaps(ids []uint64) {
	slices.Sort(ids)
	d.u(uint32(len(ids)))

	for _, id := range ids {
		d.u(uint32(id >> 32))
		d.u(uint32(id & 0xffff))
	}
}
func queryIdentity(id box2d.Shape) uint64 { return id.ID() & 0xffffffff0000ffff }
func queryObserve(r *oracle.Reader, step, partition int, kind, field string, d queryDigest) {
	r.U64(fmt.Sprintf("%d.%d.%s.%s", step, partition, kind, field), d.h)
}

func observeQueriesPayload(r *oracle.Reader, world *box2d.World, step int) {
	for f := range 2 {
		filter := box2d.DefaultQueryFilter()
		if f == 1 {
			filter.MaskBits = 1
		}

		for kind, name := range []string{"ray", "circle", "overlap"} {
			input := digestQueryInputs(filter, kind)

			queryObserve(r, step, f, name, "inputs", input)

			d := newQueryDigest()
			circle := box2d.MakeProxy([]box2d.Vec2{{}}, .1)

			for i := range queryCount {
				d.u(uint32(i))
				origin, translation := gQueryBenchmark.origins[i], gQueryBenchmark.translations[i]

				switch kind {
				case 0:
					d.hit(world.CastRayClosest(origin, translation, filter))
				case 1:
					var result box2d.RayResult

					stats := world.CastShape(origin, &circle, translation, filter, func(id box2d.Shape, p box2d.Pos, n box2d.Vec2, fraction float32) float32 {
						result = box2d.RayResult{Shape: id, Point: p, Normal: n, Fraction: fraction, Hit: true}
						return fraction
					})
					result.NodeVisits = stats.NodeVisits
					result.LeafVisits = stats.LeafVisits
					d.hit(result)
				case 2:
					var ids []uint64

					stats := world.OverlapAABB(origin, box2d.AABB{LowerBound: box2d.Vec2{X: -5, Y: -5}, UpperBound: box2d.Vec2{X: 5, Y: 5}}, filter, func(id box2d.Shape) bool { ids = append(ids, queryIdentity(id)); return true })

					d.overlaps(ids)
					d.stats(stats)
				}
			}

			queryObserve(r, step, f, name, "payload", d)
		}
	}
}

func observeTreePayload(r *oracle.Reader, step int) {
	for k := range treeCastLengthCount {
		for kind, name := range []string{"rayAll", "rayMask1", "box"} {
			mask := box2d.DefaultMaskBits
			if kind == 1 {
				mask = 1
			}

			input := newQueryDigest()
			input.u(uint32(mask))
			input.u(uint32(mask >> 32))

			for i := range treeCastRayCount {
				input.u(uint32(i))
				input.vec(gTreeCast.origins[i])
				input.vec(gTreeCast.translations[k][i])
				input.f(1)

				if kind == 2 {
					input.f(treeCastBoxExtent)
				}
			}

			queryObserve(r, step, k, name, "inputs", input)

			d := newQueryDigest()

			for i := range treeCastRayCount {
				d.treeCast(k, kind, i, mask)
			}

			queryObserve(r, step, k, name, "payload", d)
		}
	}
}

func observeTilePayload(r *oracle.Reader, world *box2d.World, step int) {
	for kind, name := range []string{"ray", "overlap"} {
		input := newQueryDigest()
		state := uint32(1234) + 7919*uint32(step)

		rays := make([]struct {
			origin      box2d.Pos
			translation box2d.Vec2
		}, tileQueryCount)
		for i := range tileQueryCount {
			origin := box2d.Vec2{X: gTileWorld.xLeft + tileRandom(&state, 0, float32(gTileWorld.groundCount)*gTileWorld.bodySpan), Y: tileRandom(&state, -2, 18)}.Pos()
			angle := tileRandom(&state, 0, 2*box2d.Pi)
			sine, cosine := boundedSinCos(angle)
			translation := box2d.Vec2{X: tileRayLength * cosine, Y: tileRayLength * sine}
			rays[i].origin, rays[i].translation = origin, translation
			input.u(uint32(i))
			input.pos(origin)

			if kind == 0 {
				input.vec(translation)
			} else {
				input.vec(box2d.Vec2{X: -5, Y: -5})
				input.vec(box2d.Vec2{X: 5, Y: 5})
			}
		}

		queryObserve(r, step, 0, name, "inputs", input)

		d := newQueryDigest()
		filter := box2d.DefaultQueryFilter()

		for i, ray := range rays {
			d.u(uint32(i))

			if kind == 0 {
				d.hit(world.CastRayClosest(ray.origin, ray.translation, filter))
			} else {
				var ids []uint64

				stats := world.OverlapAABB(ray.origin, box2d.AABB{LowerBound: box2d.Vec2{X: -5, Y: -5}, UpperBound: box2d.Vec2{X: 5, Y: 5}}, filter, func(id box2d.Shape) bool { ids = append(ids, queryIdentity(id)); return true })

				d.overlaps(ids)
				d.stats(stats)
			}
		}

		queryObserve(r, step, 0, name, "payload", d)
	}
}

func TestScenePerQueryInputsAndPayloads(t *testing.T) {
	for _, workers := range []int{1, 4} {
		for _, name := range []string{"Queries", "TreeCast", "TileWorld"} {
			t.Run(fmt.Sprintf("%s/workers=%d", name, workers), func(t *testing.T) {
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers

				world := box2d.NewWorld(&def)
				defer world.Close()

				r := oracle.Open(t, fmt.Sprintf("scenes/%sPayload_workers%d", name, workers))
				switch name {
				case "Queries":
					CreateQueries(world)
				case "TreeCast":
					CreateTreeCast(world)

					defer DestroyTreeCast()
				case "TileWorld":
					CreateTileWorld(world)

					defer DestroyTileWorld()
				}

				for step := range 300 {
					switch name {
					case "Queries":
						observeQueriesPayload(r, world, step)
						StepQueries(world, step)
					case "TreeCast":
						observeTreePayload(r, step)
						StepTreeCast(world, step)
					case "TileWorld":
						StepTileWorld(world, step)
						observeTilePayload(r, world, step)
					}

					world.Step(float32(1)/60, 4)
				}
			})
		}
	}
}

func digestQueryInputs(filter box2d.QueryFilter, kind int) queryDigest {
	input := newQueryDigest()
	input.u(uint32(filter.MaskBits))
	input.u(uint32(filter.MaskBits >> 32))

	for i := range queryCount {
		input.u(uint32(i))
		input.pos(gQueryBenchmark.origins[i])

		if kind != 2 {
			input.vec(gQueryBenchmark.translations[i])
		} else {
			input.vec(box2d.Vec2{X: -5, Y: -5})
			input.vec(box2d.Vec2{X: 5, Y: 5})
		}

		if kind == 1 {
			input.f(.1)
		}
	}

	return input
}

func (d *queryDigest) treeCast(k, kind, i int, mask uint64) {
	context := treeCastContext{tightBoxes: gTreeCast.tightBoxes, fraction: 1}
	proxy := -1
	identity := uint64(math.MaxUint64)

	var stats box2d.TreeStats

	if kind < 2 {
		input := box2d.RayCastInput{Origin: gTreeCast.origins[i], Translation: gTreeCast.translations[k][i], MaxFraction: 1}
		stats = gTreeCast.tree.CastRay(&input, mask, func(in *box2d.RayCastInput, id int, value uint64) float32 {
			f := treeCastRayCallback(in, value, &context)
			if f >= 0 {
				proxy = id
				identity = value
			}

			return f
		})
	} else {
		o := gTreeCast.origins[i]
		input := box2d.BoxCastInput{Box: box2d.AABB{LowerBound: box2d.Vec2{X: o.X - treeCastBoxExtent, Y: o.Y - treeCastBoxExtent}, UpperBound: box2d.Vec2{X: o.X + treeCastBoxExtent, Y: o.Y + treeCastBoxExtent}}, Translation: gTreeCast.translations[k][i], MaxFraction: 1}
		stats = gTreeCast.tree.CastBox(&input, mask, func(in *box2d.BoxCastInput, id int, value uint64) float32 {
			f := treeCastBoxCallback(in, value, &context)
			if f >= 0 {
				proxy = id
				identity = value
			}

			return f
		})
	}

	d.u(uint32(i))
	d.u(uint32(proxy))
	d.u(uint32(identity))
	d.u(uint32(identity >> 32))
	d.f(context.fraction)
	d.stats(stats)
}
