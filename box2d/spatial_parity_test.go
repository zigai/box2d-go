// SPDX-License-Identifier: MIT

package box2d

import (
	"math"
	"slices"
	"testing"

	"github.com/zigai/box2d-go/internal/oracle"
)

func spatialBox(r *oracle.Reader, key string, a AABB) {
	r.Vec2(key+"/lower", a.LowerBound.X, a.LowerBound.Y)
	r.Vec2(key+"/upper", a.UpperBound.X, a.UpperBound.Y)
}

func spatialStats(r *oracle.Reader, s TreeStats) {
	r.I32("nodes", int32(s.NodeVisits))
	r.I32("leaves", int32(s.LeafVisits))
}

func observeTreeQuery(r *oracle.Reader, tree *DynamicTree, a AABB, mask uint64, all, stop bool) {
	type hit struct {
		id   int
		data uint64
	}

	var hits []hit

	callback := func(id int, data uint64) bool { hits = append(hits, hit{id, data}); return !stop }

	var s TreeStats
	if all {
		s = tree.QueryAll(a, callback)
	} else {
		s = tree.Query(a, mask, callback)
	}

	if !stop {
		slices.SortStableFunc(hits, func(a, b hit) int { return a.id - b.id })
	}

	spatialStats(r, s)
	r.I32("count", int32(len(hits)))

	for _, h := range hits {
		r.I32("id", int32(h.id))
		r.U64("data", h.data)
	}
}

func TestPublicTreeHistory(t *testing.T) {
	r := oracle.Open(t, "spatial/PublicTreeHistory")
	tree := NewDynamicTree(4)

	var (
		ids  [4]int
		live [4]bool
	)

	observe := func() {
		r.I32("proxies", int32(tree.ProxyCount()))
		r.I32("height", int32(tree.Height()))
		r.F32("ratio", tree.AreaRatio())
		spatialBox(r, "root", tree.RootBounds())

		for i := range 4 {
			if live[i] {
				r.I32("proxy", int32(ids[i]))
				r.U64("category", tree.CategoryBits(ids[i]))
				r.U64("userdata", tree.UserData(ids[i]))
				spatialBox(r, "aabb", tree.AABB(ids[i]))
			}
		}

		queries := []AABB{{Vec2{-10, -10}, Vec2{10, 10}}, {Vec2{-.25, -.25}, Vec2{.25, .25}}, {Vec2{20, 20}, Vec2{21, 21}}}
		masks := []uint64{math.MaxUint64, 1, 2, 1 << 63}

		query := func(a AABB, mask uint64, all, stop bool) {
			observeTreeQuery(r, &tree, a, mask, all, stop)
		}
		for _, a := range queries {
			for _, mask := range masks {
				query(a, mask, false, false)
			}
		}

		query(queries[0], 0, true, false)
		query(queries[0], 0, true, true)
	}
	observe()

	for i := range 4 {
		category := uint64(1) << i
		if i == 3 {
			category = 1 << 63
		}

		ids[i] = tree.CreateProxy(AABB{Vec2{float32(i) - .5, -.5}, Vec2{float32(i) + .5, .5}}, category, uint64(100+i))
		live[i] = true

		observe()
	}

	tree.MoveProxy(ids[0], AABB{Vec2{4, -1}, Vec2{5, 1}})
	observe()
	tree.SetCategoryBits(ids[2], 1<<63|1)
	observe()
	tree.DestroyProxy(ids[1])

	live[1] = false

	observe()

	ids[1] = tree.CreateProxy(AABB{Vec2{-.5, -.5}, Vec2{.5, .5}}, 2, 999)
	live[1] = true

	observe()

	for i := range 4 {
		tree.DestroyProxy(ids[i])
		live[i] = false

		observe()
	}
}

func TestTreeCallbackForwarding(t *testing.T) {
	r := oracle.Open(t, "spatial/TreeCallbackForwarding")

	tree := NewDynamicTree(4)
	for i := range 3 {
		tree.CreateProxy(AABB{Vec2{.5 + float32(.25*float32(i)), -1}, Vec2{.75 + float32(.25*float32(i)), 1}}, 1, uint64(100+i))
	}

	tree.CreateProxy(AABB{Vec2{3, -1}, Vec2{3.25, 1}}, 1, 200)

	for policy := range 4 {
		count := 0
		callback := func(incoming float32, id int, data uint64) float32 {
			result := incoming

			switch policy {
			case 1:
				result = -1
			case 2:
				result = 0
			case 3:
				result = .5
			}

			r.I32("id", int32(id))
			r.U64("data", data)
			r.F32("incoming", incoming)
			r.F32("return", result)

			count++

			return result
		}
		ray := RayCastInput{Vec2{}, Vec2{4, 0}, 1}
		spatialStats(r, tree.CastRay(&ray, 1, func(in *RayCastInput, id int, data uint64) float32 {
			r.Vec2("origin", in.Origin.X, in.Origin.Y)
			r.Vec2("translation", in.Translation.X, in.Translation.Y)

			return callback(in.MaxFraction, id, data)
		}))
		r.I32("count", int32(count))
		count = 0
		box := BoxCastInput{AABB{Vec2{-.1, -.1}, Vec2{.1, .1}}, Vec2{4, 0}, 1}
		spatialStats(r, tree.CastBox(&box, 1, func(in *BoxCastInput, id int, data uint64) float32 {
			spatialBox(r, "input", in.Box)
			r.Vec2("translation", in.Translation.X, in.Translation.Y)

			return callback(in.MaxFraction, id, data)
		}))
		r.I32("count", int32(count))
	}
}

type spatialHit struct {
	id       Shape
	point    Pos
	normal   Vec2
	fraction float32
	plane    PlaneResult
}
type spatialHits struct {
	hits   []spatialHit
	policy int
	world  *World
}

func (h *spatialHits) overlap(id Shape) bool {
	h.hits = append(h.hits, spatialHit{id: id})
	return h.policy != 2
}

func (h *spatialHits) cast(id Shape, p Pos, n Vec2, f float32) float32 {
	h.hits = append(h.hits, spatialHit{id: id, point: p, normal: n, fraction: f})
	switch h.policy {
	case 0:
		return 1
	case 1:
		return -1
	case 2:
		return 0
	default:
		return f
	}
}

func (h *spatialHits) plane(id Shape, p *PlaneResult) bool {
	h.hits = append(h.hits, spatialHit{id: id, plane: *p})
	return h.policy != 2
}

func spatialIdentity(r *oracle.Reader, id Shape, w *World) {
	r.I32("shape/index", id.index1)
	r.U32("shape/generation", uint32(id.generation))
	r.Bool("shape/owner", id.world == w)
}

func (h *spatialHits) observe(r *oracle.Reader, payload int) {
	if h.policy < 2 {
		slices.SortStableFunc(h.hits, func(a, b spatialHit) int { return int(a.id.index1 - b.id.index1) })
	}

	r.I32("count", int32(len(h.hits)))

	for _, p := range h.hits {
		spatialIdentity(r, p.id, h.world)

		if payload == 1 {
			r.Pos("point", p.point.X, p.point.Y)
			r.Vec2("normal", p.normal.X, p.normal.Y)
			r.F32("fraction", p.fraction)
		}

		if payload == 2 {
			r.Bool("hit", p.plane.Hit)
			r.Vec2("normal", p.plane.Plane.Normal.X, p.plane.Plane.Normal.Y)
			r.F32("offset", p.plane.Plane.Offset)
			r.Vec2("point", p.plane.Point.X, p.plane.Point.Y)
		}
	}
}

func spatialInspect(r *oracle.Reader, w *World) {
	c := w.Counters()
	r.I32("bodies", int32(c.BodyCount))
	r.I32("shapes", int32(c.ShapeCount))
	r.I32("joints", int32(c.JointCount))
	r.I32("contacts", int32(c.ContactCount))
	r.I32("islands", int32(c.IslandCount))
	r.I32("staticHeight", int32(c.StaticTreeHeight))
	r.I32("treeHeight", int32(c.TreeHeight))
	r.I32("awakeContacts", int32(c.AwakeContactCount))
	r.I32("recycled", int32(c.RecycledContactCount))

	for _, count := range c.ColorCounts {
		r.I32("color", int32(count))
	}

	spatialBox(r, "bounds", w.Bounds())
}

func populateSpatialWorld(r *oracle.Reader, w *World, base Pos, bodies *[5]Body, shapes *[8]Shape) {
	for i := range 5 {
		bd := DefaultBodyDef()

		bd.Position = base.Add(Vec2{3 * float32(i), 0})
		if i < 3 {
			bd.Type = BodyType(i)
		}

		bodies[i] = w.CreateBody(&bd)
		sd := DefaultShapeDef()
		sd.Filter.CategoryBits = uint64(1) << i

		sd.Filter.MaskBits = 1
		if i%2 != 0 {
			sd.Filter.MaskBits = 2
		}

		switch i {
		case 0:
			c := Circle{Vec2{}, .75}
			shapes[i] = bodies[i].CreateCircleShape(&sd, &c)
		case 1:
			c := Capsule{Vec2{0, -.5}, Vec2{0, .5}, .5}
			shapes[i] = bodies[i].CreateCapsuleShape(&sd, &c)
		case 2:
			p := MakeBox(.5, .75)
			shapes[i] = bodies[i].CreatePolygonShape(&sd, &p)
		case 3:
			s := Segment{Vec2{0, 1}, Vec2{0, -1}}
			shapes[i] = bodies[i].CreateSegmentShape(&sd, &s)
		case 4:
			cd := DefaultChainDef()
			cd.Points = []Vec2{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
			cd.IsLoop = true
			cd.Filter = sd.Filter
			chain := bodies[i].CreateChain(&cd)

			count := 0
			for segment := range chain.Segments() {
				shapes[4+count] = segment
				count++
			}

			r.I32("segments", int32(count))
		}
	}
}

func observeFilteredSpatialCasts(r *oracle.Reader, w *World, base Pos, filter QueryFilter, proxy *ShapeProxy) {
	var h spatialHits

	for start := range 3 {
		offset := Vec2{}
		if start == 0 {
			offset.X = -2
		}

		if start == 2 {
			offset.Y = 4
		}

		origin := base.Add(offset)

		for policy := range 4 {
			h = spatialHits{world: w, policy: policy}
			spatialStats(r, w.CastRay(origin, Vec2{18, 0}, filter, h.cast))
			h.observe(r, 1)
			h = spatialHits{world: w, policy: policy}
			spatialStats(r, w.CastShape(origin, proxy, Vec2{18, 0}, filter, h.cast))
			h.observe(r, 1)
		}

		rr := w.CastRayClosest(origin, Vec2{18, 0}, filter)
		r.Bool("closest/hit", rr.Hit)
		spatialStats(r, TreeStats{rr.NodeVisits, rr.LeafVisits})

		if rr.Hit {
			spatialIdentity(r, rr.Shape, w)
			r.Pos("point", rr.Point.X, rr.Point.Y)
			r.Vec2("normal", rr.Normal.X, rr.Normal.Y)
			r.F32("fraction", rr.Fraction)
		}
	}
}

func observeFilteredSpatialQueries(r *oracle.Reader, w *World, base Pos, all QueryFilter, proxy *ShapeProxy) {
	var h spatialHits

	for _, filter := range []QueryFilter{all, {1, 31}, {2, 31}, {1, 2}} {
		for stop := range 2 {
			policy := 0
			if stop != 0 {
				policy = 2
			}

			h = spatialHits{world: w, policy: policy}
			spatialStats(r, w.OverlapAABB(base, AABB{Vec2{-2, -2}, Vec2{16, 2}}, filter, h.overlap))
			h.observe(r, 0)
		}

		for i := range 5 {
			origin := base.Add(Vec2{3 * float32(i), 0})
			h = spatialHits{world: w}
			spatialStats(r, w.OverlapShape(origin, proxy, filter, h.overlap))
			h.observe(r, 0)
		}

		observeFilteredSpatialCasts(r, w, base, filter, proxy)
	}
}

func observeShapeSpatialQueries(r *oracle.Reader, w *World, base Pos, shapes *[8]Shape) {
	for i := range 8 {
		x := 3 * float32(min(i, 4))
		spatialIdentity(r, shapes[i], w)
		r.Bool("inside", shapes[i].TestPoint(base.Add(Vec2{x, 0})))
		r.Bool("outside", shapes[i].TestPoint(base.Add(Vec2{x, 4})))
		origins := []Vec2{{x - 2, 0}, {x + 2, 0}, {x, -2}, {x, 2}}

		translations := []Vec2{{4, 0}, {-4, 0}, {0, 4}, {0, -4}}
		for direction := range 4 {
			out := shapes[i].RayCast(base.Add(origins[direction]), translations[direction])
			r.Bool("local/hit", out.Hit)

			if out.Hit {
				r.Pos("point", out.Point.X, out.Point.Y)
				r.Vec2("normal", out.Normal.X, out.Normal.Y)
				r.F32("fraction", out.Fraction)
				r.I32("iterations", int32(out.Iterations))
			}
		}
	}
}

func observeMoverSpatialQueries(r *oracle.Reader, w *World, base Pos, all QueryFilter) {
	var h spatialHits

	for i := range 5 {
		mover := Capsule{Vec2{0, -.25}, Vec2{0, .25}, .4}
		origin := base.Add(Vec2{float32(3*float32(i)) + .7, 0})

		for stop := range 2 {
			policy := 0
			if stop != 0 {
				policy = 2
			}

			h = spatialHits{world: w, policy: policy}
			w.CollideMover(origin, &mover, all, h.plane)
			h.observe(r, 2)
		}

		h = spatialHits{world: w}
		w.CollideMover(origin, &mover, all, h.plane)
		slices.SortStableFunc(h.hits, func(a, b spatialHit) int { return int(a.id.index1 - b.id.index1) })
		observeSpatialPlaneLimits(r, h.hits, Vec2{-1, .2}, Vec2{-2, .3})

		r.F32("mover/fraction", w.CastMover(base.Add(Vec2{float32(3*float32(i)) - 2, 0}), &mover, Vec2{4, 0}, all))
	}
}

func observeSpatialPlaneLimits(r *oracle.Reader, hits []spatialHit, translation, velocity Vec2) {
	for limit := range 3 {
		pushLimit := float32(math.MaxFloat32)
		if limit == 0 {
			pushLimit = 0
		}

		if limit == 1 {
			pushLimit = .1
		}

		var planes []CollisionPlane

		for _, hit := range hits {
			if hit.plane.Hit {
				planes = append(planes, CollisionPlane{Plane: hit.plane.Plane, PushLimit: pushLimit, ClipVelocity: limit != 1})
			}
		}

		result := SolvePlanes(translation, planes)
		r.Vec2("delta", result.Delta.X, result.Delta.Y)
		r.I32("iterations", int32(result.IterationCount))

		for _, p := range planes {
			r.F32("push", p.Push)
		}

		clip := ClipVector(velocity, planes)
		r.Vec2("clip", clip.X, clip.Y)
	}
}

func TestWorldOverlapStopsAcrossBodyTypes(t *testing.T) {
	wd := DefaultWorldDef()
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	circle := Circle{Radius: 1}
	sd := DefaultShapeDef()

	for _, bodyType := range []BodyType{StaticBody, KinematicBody, DynamicBody} {
		bd := DefaultBodyDef()
		bd.Type = bodyType
		body := w.CreateBody(&bd)
		body.CreateCircleShape(&sd, &circle)
	}

	filter := DefaultQueryFilter()
	proxy := MakeProxy([]Vec2{{}}, 1)

	queries := []struct {
		name  string
		query func(OverlapResultFunc) TreeStats
	}{
		{"AABB", func(callback OverlapResultFunc) TreeStats {
			return w.OverlapAABB(Pos{}, AABB{LowerBound: Vec2{-2, -2}, UpperBound: Vec2{2, 2}}, filter, callback)
		}},
		{"shape", func(callback OverlapResultFunc) TreeStats {
			return w.OverlapShape(Pos{}, &proxy, filter, callback)
		}},
	}
	for _, tc := range queries {
		t.Run(tc.name, func(t *testing.T) {
			count := 0

			tc.query(func(Shape) bool { count++; return true })

			if count != 3 {
				t.Fatalf("continuing callback saw %d shapes, want 3", count)
			}

			count = 0
			stats := tc.query(func(Shape) bool { count++; return false })

			if count != 1 {
				t.Fatalf("stopping callback saw %d shapes, want 1", count)
			}

			if stats.NodeVisits != 1 || stats.LeafVisits != 1 {
				t.Fatalf("stopping traversal visited %d nodes and %d leaves, want 1 each", stats.NodeVisits, stats.LeafVisits)
			}
		})
	}
}

func TestWorldSpatialParity(t *testing.T) {
	frames := []struct {
		name string
		base Pos
	}{{"WorldOrigin", Pos{}}, {"WorldOffset", Pos{32, -16}}}
	if IsDoublePrecision() {
		frames = append(frames, struct {
			name string
			base Pos
		}{"WorldFar", Pos{10000000, -10000000}})
	}

	for _, frame := range frames {
		t.Run(frame.name, func(t *testing.T) {
			r := oracle.Open(t, "spatial/"+frame.name)
			base := frame.base
			wd := DefaultWorldDef()
			wd.Gravity = Vec2{}
			wd.WorkerCount = 1
			w := NewWorld(&wd)
			CleanupWorld(t, w)
			spatialInspect(r, w)

			all := QueryFilter{math.MaxUint64, math.MaxUint64}
			proxy := MakeProxy([]Vec2{{}}, .3)
			h := spatialHits{world: w}
			spatialStats(r, w.OverlapAABB(base, AABB{Vec2{-2, -2}, Vec2{16, 2}}, all, h.overlap))
			h.observe(r, 0)
			h = spatialHits{world: w}
			spatialStats(r, w.CastRay(base, Vec2{16, 0}, all, h.cast))
			h.observe(r, 1)
			h = spatialHits{world: w}
			spatialStats(r, w.OverlapShape(base, &proxy, all, h.overlap))
			h.observe(r, 0)
			h = spatialHits{world: w}
			spatialStats(r, w.CastShape(base, &proxy, Vec2{16, 0}, all, h.cast))
			h.observe(r, 1)

			var (
				bodies [5]Body
				shapes [8]Shape
			)

			populateSpatialWorld(r, w, base, &bodies, &shapes)

			spatialInspect(r, w)

			observeFilteredSpatialQueries(r, w, base, all, &proxy)

			observeShapeSpatialQueries(r, w, base, &shapes)

			observeMoverSpatialQueries(r, w, base, all)

			bodies[2].SetTransform(base.Add(Vec2{6, 2}), MakeRot(.25))
			spatialInspect(r, w)
			w.Step(float32(1)/60, 4)
			spatialInspect(r, w)

			for n := range 5 {
				bodies[4-n].Destroy()
				spatialInspect(r, w)
			}
		})
	}
}

func observeMoverPlaneScenario(t *testing.T, r *oracle.Reader, base Pos, scenario int) {
	t.Helper()

	wd := DefaultWorldDef()
	wd.Gravity = Vec2{}
	wd.WorkerCount = 1
	w := NewWorld(&wd)
	CleanupWorld(t, w)

	filter := QueryFilter{2, 1 << 63}

	if scenario != 0 {
		bd := DefaultBodyDef()
		bd.Position = base
		body := w.CreateBody(&bd)
		sd := DefaultShapeDef()
		sd.Filter.CategoryBits = 1 << 63
		sd.Filter.MaskBits = 2

		wall := MakeOffsetBox(.5, 3, Vec2{1, 0}, RotIdentity)
		if scenario == 4 {
			wall = MakeOffsetBox(3, .5, Vec2{0, -1}, MakeRot(.25))
		}

		body.CreatePolygonShape(&sd, &wall)

		if scenario == 3 {
			floor := MakeOffsetBox(3, .5, Vec2{0, -1}, RotIdentity)
			body.CreatePolygonShape(&sd, &floor)
		}
	}

	mover := Capsule{Vec2{0, -.25}, Vec2{0, .25}, .5}

	offset := Vec2{}
	if scenario == 2 {
		offset.X = .3
	}

	if scenario == 5 {
		offset.X = -.5
	}

	if scenario == 3 {
		offset.Y = -.25
	}

	origin := base.Add(offset)
	h := spatialHits{world: w}
	w.CollideMover(origin, &mover, filter, h.plane)
	h.observe(r, 2)

	observeSpatialPlaneLimits(r, h.hits, Vec2{1, -1}, Vec2{2, -3})

	r.F32("cast", w.CastMover(origin, &mover, Vec2{2, -1}, filter))
	r.F32("filteredCast", w.CastMover(origin, &mover, Vec2{2, -1}, QueryFilter{1, 1 << 63}))
	w.Close()
}

func TestWorldMoverPlanesResolveAndClip(t *testing.T) {
	frames := []struct {
		name string
		base Pos
	}{{"MoverOrigin", Pos{}}, {"MoverOffset", Pos{32, -16}}}
	if IsDoublePrecision() {
		frames = append(frames, struct {
			name string
			base Pos
		}{"MoverFar", Pos{10000000, -10000000}})
	}

	for _, frame := range frames {
		t.Run(frame.name, func(t *testing.T) {
			r := oracle.Open(t, "spatial/"+frame.name)
			base := frame.base

			for scenario := range 6 {
				observeMoverPlaneScenario(t, r, base, scenario)
			}
		})
	}
}
