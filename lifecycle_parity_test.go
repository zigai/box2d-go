// SPDX-License-Identifier: MIT

package box2d_test

import (
	"slices"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

// World slots are process-local; retain all object slot and generation bits.
func lifecycleShapeKey(s box2d.Shape) uint64 { return s.ID() & 0xffffffff0000ffff }
func lifecycleBodyKey(b box2d.Body) uint64   { return b.ID() & 0xffffffff0000ffff }

func TestLifecycleNullHandles(t *testing.T) {
	r := oracle.Open(t, "lifecycle/NullHandles")
	lifecycleValidity(r, nil, box2d.Body{}, box2d.Shape{}, box2d.Chain{}, box2d.Joint{}, box2d.Contact{})
}

func lifecycleChain(b box2d.Body) box2d.Chain {
	d := box2d.DefaultChainDef()
	d.Points = []box2d.Vec2{{X: 10}, {X: 14}, {X: 14, Y: 4}, {X: 10, Y: 4}}
	d.IsLoop = true
	d.Materials = []box2d.SurfaceMaterial{box2d.DefaultSurfaceMaterial()}

	return b.CreateChain(&d)
}

func lifecycleJoint(w *box2d.World, a, b box2d.Body) box2d.Joint {
	d := box2d.DefaultRevoluteJointDef()
	d.Base.BodyA = a
	d.Base.BodyB = b
	d.Base.LocalFrameA.P = a.LocalPoint(b.Position())
	d.Base.CollideConnected = true

	return w.CreateRevoluteJoint(&d).Joint
}

func lifecycleValidity(r *oracle.Reader, w *box2d.World, b box2d.Body, s box2d.Shape, c box2d.Chain, j box2d.Joint, contact box2d.Contact) {
	r.Bool("world.valid", w.IsValid())
	r.Bool("body.valid", b.IsValid())
	r.Bool("shape.valid", s.IsValid())
	r.Bool("chain.valid", c.IsValid())
	r.Bool("joint.valid", j.IsValid())
	r.Bool("contact.valid", contact.IsValid())
}

func lifecycleMembership(t *testing.T, r *oracle.Reader, w *box2d.World, b, peer box2d.Body, s box2d.Shape, c box2d.Chain, j box2d.Joint) {
	t.Helper()
	r.Bool("body.world", b.World() == w)
	r.Bool("shape.world", s.World() == w)
	r.Bool("shape.body", s.Body() == b)
	r.Bool("chain.world", c.World() == w)
	r.Bool("joint.world", j.World() == w)
	r.Bool("joint.a", j.BodyA() == b)
	r.Bool("joint.b", j.BodyB() == peer)

	before := w.StateHash()

	for range b.Shapes() {
		break
	}

	for range b.Joints() {
		break
	}

	for range c.Segments() {
		break
	}

	if w.StateHash() != before {
		t.Fatal("early-stop enumeration mutated canonical state")
	}

	shapes := slices.Collect(b.Shapes())
	r.I32("shape.count", int32(b.ShapeCount()))
	r.I32("shape.enumerated", int32(len(shapes)))

	seen := map[box2d.Shape]bool{}
	for _, shape := range shapes {
		if seen[shape] {
			t.Fatal("duplicate body shape")
		}

		seen[shape] = true
		r.U64("shape.member", lifecycleShapeKey(shape))
	}

	joints := slices.Collect(b.Joints())
	r.I32("joint.count", int32(b.JointCount()))
	r.I32("joint.enumerated", int32(len(joints)))

	for _, joint := range joints {
		r.Bool("joint.member", joint == j)
	}

	segments := slices.Collect(c.Segments())
	r.I32("chain.count", int32(c.SegmentCount()))
	r.I32("chain.enumerated", int32(len(segments)))

	seen = map[box2d.Shape]bool{}
	for _, segment := range segments {
		if seen[segment] {
			t.Fatal("duplicate chain segment")
		}

		seen[segment] = true
		r.U64("segment", lifecycleShapeKey(segment))
		r.Bool("segment.parent", segment.ParentChain() == c)
		r.Bool("segment.body", segment.Body() == b)
	}
}

func TestLifecycleReuseAndDestruction(t *testing.T) {
	r := oracle.Open(t, "lifecycle/ReuseAndDestruction")

	var (
		oldw       *box2d.World
		oldb       box2d.Body
		olds       box2d.Shape
		oldc       box2d.Chain
		oldj       box2d.Joint
		oldcontact box2d.Contact
	)

	for range 4 {
		wd := box2d.DefaultWorldDef()
		wd.Gravity = box2d.Vec2{}
		w := box2d.NewWorld(&wd)
		box2d.BTWorldCleanup(t, w)
		lifecycleValidity(r, oldw, oldb, olds, oldc, oldj, oldcontact)

		bd := box2d.DefaultBodyDef()
		ground := w.CreateBody(&bd)
		sd := box2d.DefaultShapeDef()
		sd.EnableContactEvents = true
		circle := box2d.Circle{Radius: 1}
		ground.CreateCircleShape(&sd, &circle)

		bd.Type = box2d.DynamicBody
		b := w.CreateBody(&bd)
		s := b.CreateCircleShape(&sd, &circle)
		circle.Center = box2d.Vec2{X: 4}
		extra := b.CreateCircleShape(&sd, &circle)
		circle.Center = box2d.Vec2{}
		bd.Position = box2d.Pos{Y: 8}
		peer := w.CreateBody(&bd)
		c := lifecycleChain(b)
		j := lifecycleJoint(w, b, peer)
		lifecycleMembership(t, r, w, b, peer, s, c, j)
		lifecycleValidity(r, oldw, oldb, olds, oldc, oldj, oldcontact)
		w.Step(1.0/60, 4)

		data := slices.Collect(s.ContactData())
		r.I32("contact.count", int32(len(data)))

		if len(data) != 1 {
			t.Fatalf("initial contacts=%d, want 1", len(data))
		}

		contact := data[0].Contact
		r.Bool("contact.member", data[0].ShapeA == s || data[0].ShapeB == s)
		r.I32("contact.points", int32(data[0].Manifold.PointCount))
		lifecycleValidity(r, w, b, s, c, j, contact)
		direct := contact.Data()
		r.Bool("contact.direct.id", direct.Contact == contact)
		r.U64("contact.direct.a", lifecycleShapeKey(direct.ShapeA))
		r.U64("contact.direct.b", lifecycleShapeKey(direct.ShapeB))
		r.Vec2("contact.normal", direct.Manifold.Normal.X, direct.Manifold.Normal.Y)
		r.I32("contact.direct.points", int32(direct.Manifold.PointCount))
		lifecycleCheckContactViews(t, r, w, b, s, contact)

		b.SetTransform(box2d.Pos{X: 30}, box2d.Rot{C: 1})
		w.Step(1.0/60, 4)
		r.Bool("separated.contact.valid", contact.IsValid())
		b.SetTransform(box2d.Pos{}, box2d.Rot{C: 1})
		w.Step(1.0/60, 4)

		data = slices.Collect(s.ContactData())
		r.I32("returned.contacts", int32(len(data)))

		if len(data) != 1 {
			t.Fatalf("returned contacts=%d, want 1", len(data))
		}

		next := data[0].Contact
		r.Bool("contact.reused", next.ID()>>32 == contact.ID()>>32)
		r.U32("contact.generation.delta", uint32(next.ID())-uint32(contact.ID()))
		r.Bool("old.contact.valid", contact.IsValid())
		contact = next

		extra.Destroy(true)

		replacement := b.CreateCircleShape(&sd, &circle)

		r.Bool("shape.old.valid", extra.IsValid())
		r.Bool("shape.reused", extra.ID()>>32 == replacement.ID()>>32)
		r.U32("shape.generation.delta", uint32(uint16(replacement.ID()-extra.ID())))

		segments := slices.Collect(c.Segments())
		c.Destroy()

		nc := lifecycleChain(b)

		r.Bool("chain.old.valid", c.IsValid())
		r.Bool("chain.reused", c.ID()>>32 == nc.ID()>>32)
		r.U32("chain.generation.delta", uint32(uint16(nc.ID()-c.ID())))

		for _, segment := range segments {
			r.Bool("segment.old.valid", segment.IsValid())
		}

		c = nc

		j.Destroy()

		nj := lifecycleJoint(w, b, peer)

		r.Bool("joint.old.valid", j.IsValid())
		r.Bool("joint.reused", j.ID()>>32 == nj.ID()>>32)
		r.U32("joint.generation.delta", uint32(uint16(nj.ID()-j.ID())))
		j = nj
		lifecycleMembership(t, r, w, b, peer, s, c, j)
		segments = slices.Collect(c.Segments())

		b.Destroy()
		lifecycleValidity(r, w, b, s, c, j, contact)
		r.Bool("replacement.valid", replacement.IsValid())

		for _, segment := range segments {
			r.Bool("dependent.segment.valid", segment.IsValid())
		}

		r.I32("peer.joints", int32(peer.JointCount()))

		bd.Position = box2d.Pos{X: 20}
		nb := w.CreateBody(&bd)
		r.Bool("body.reused", b.ID()>>32 == nb.ID()>>32)
		r.U32("body.generation.delta", uint32(uint16(nb.ID()-b.ID())))
		r.Bool("body.old.valid", b.IsValid())
		w.Step(1.0/60, 4)
		counts := w.Counters()
		r.I32("bodies", int32(counts.BodyCount))
		r.I32("shapes", int32(counts.ShapeCount))
		r.I32("joints", int32(counts.JointCount))
		r.I32("contacts", int32(counts.ContactCount))

		p := peer.Position()
		r.Pos("peer.position", p.X, p.Y)
		nb.SetTransform(box2d.Pos{}, box2d.Rot{C: 1})
		oldw, oldb = w, nb
		olds = nb.CreateCircleShape(&sd, &circle)
		oldc = lifecycleChain(nb)
		oldj = lifecycleJoint(w, nb, peer)
		w.Step(1.0/60, 4)

		data = slices.Collect(olds.ContactData())
		r.I32("world.destroy.contacts", int32(len(data)))

		if len(data) != 1 {
			t.Fatalf("pre-world-destroy contacts=%d, want 1", len(data))
		}

		oldcontact = data[0].Contact
		lifecycleValidity(r, oldw, oldb, olds, oldc, oldj, oldcontact)
		w.Close()
		lifecycleValidity(r, oldw, oldb, olds, oldc, oldj, oldcontact)
	}
}

func TestLifecycleMultiworldDestroyRecreate(t *testing.T) {
	r := oracle.Open(t, "lifecycle/Multiworld")
	a := newOwnershipScene(t, 0, 1, nil)
	b := newOwnershipScene(t, 1, 4, nil)

	old := a.w
	for step := range 12 {
		a.w.Step(1.0/60, 4)
		b.w.Step(1.0/60, 4)
		observeOwnershipScene(r, a, step)
		observeOwnershipScene(r, b, step)
	}

	a.w.Close()
	a = newOwnershipScene(t, 0, 2, nil)

	r.Bool("old.world.valid", old.IsValid())
	r.Bool("worlds.distinct", a.w != b.w)

	for step := 12; step < 24; step++ {
		a.w.Step(1.0/60, 4)
		b.w.Step(1.0/60, 4)
		observeOwnershipScene(r, a, step)
		observeOwnershipScene(r, b, step)
	}

	a.w.Close()
	b.w.Close()
}

func lifecycleCollectionObservation(t *testing.T, r *oracle.Reader, w *box2d.World, b box2d.Body) {
	t.Helper()

	before := w.StateHash()

	for range b.Shapes() {
		break
	}

	for range b.Joints() {
		break
	}

	for range b.ContactData() {
		break
	}

	if w.StateHash() != before {
		t.Fatal("early-stop collection mutated simulation")
	}

	shapes := slices.Collect(b.Shapes())
	r.I32("shapes", int32(len(shapes)))

	for _, s := range shapes {
		r.U64("shape", lifecycleShapeKey(s))
	}

	joints := slices.Collect(b.Joints())
	r.I32("joints", int32(len(joints)))

	for _, j := range joints {
		r.U64("joint", j.ID()&0xffffffff0000ffff)
	}

	r.I32("contact.capacity", int32(b.ContactCapacity()))
	data := slices.Collect(b.ContactData())
	r.I32("contacts", int32(len(data)))

	for _, d := range data {
		id := d.Contact.ID()
		r.U32("contact.index", uint32(id>>32))
		r.U32("contact.generation", uint32(id))
		r.U64("contact.a", lifecycleShapeKey(d.ShapeA))
		r.U64("contact.b", lifecycleShapeKey(d.ShapeB))
		r.Vec2("contact.normal", d.Manifold.Normal.X, d.Manifold.Normal.Y)
		r.I32("contact.points", int32(d.Manifold.PointCount))
	}
}

func TestLifecycleCollectionTransitions(t *testing.T) {
	r := oracle.Open(t, "lifecycle/CollectionTransitions")
	wd := box2d.DefaultWorldDef()
	wd.Gravity = box2d.Vec2{}
	w := box2d.NewWorld(&wd)
	box2d.BTWorldCleanup(t, w)

	bd := box2d.DefaultBodyDef()
	ground := w.CreateBody(&bd)
	bd.Type = box2d.DynamicBody
	b := w.CreateBody(&bd)
	lifecycleCollectionObservation(t, r, w, b)

	sd := box2d.DefaultShapeDef()
	circle := box2d.Circle{Radius: 2}
	ground.CreateCircleShape(&sd, &circle)
	circle.Radius = 0.5
	a := b.CreateCircleShape(&sd, &circle)
	lifecycleCollectionObservation(t, r, w, b)

	circle.Center.X = 0.2
	second := b.CreateCircleShape(&sd, &circle)
	lifecycleCollectionObservation(t, r, w, b)
	w.Step(1.0/60, 4)
	lifecycleCollectionObservation(t, r, w, b)

	data := slices.Collect(b.ContactData())
	if len(data) != 2 {
		t.Fatalf("two-shape contacts=%d, want 2", len(data))
	}

	held := [2]box2d.Contact{data[0].Contact, data[1].Contact}

	second.Destroy(true)
	w.Step(1.0/60, 4)
	lifecycleCollectionObservation(t, r, w, b)

	for _, c := range held {
		r.Bool("held.contact.valid", c.IsValid())
	}

	bd.Position = box2d.Pos{Y: 5}
	peer := w.CreateBody(&bd)
	j1 := lifecycleJoint(w, b, peer)
	lifecycleCollectionObservation(t, r, w, b)
	j2 := lifecycleJoint(w, b, peer)
	lifecycleCollectionObservation(t, r, w, b)
	j1.Destroy()
	lifecycleCollectionObservation(t, r, w, b)
	r.Bool("removed.joint.valid", j1.IsValid())
	r.Bool("survivor.joint.valid", j2.IsValid())
	b.Destroy()
	r.Bool("dependent.shape.valid", a.IsValid())
	r.Bool("dependent.joint.valid", j2.IsValid())
	w.Step(1.0/60, 4)
	lifecycleCollectionObservation(t, r, w, ground)
	lifecycleCollectionObservation(t, r, w, peer)

	for _, c := range held {
		r.Bool("held.contact.valid", c.IsValid())
	}

	w.Close()
}

func lifecycleCheckContactViews(t *testing.T, r *oracle.Reader, w *box2d.World, b box2d.Body, s box2d.Shape, contact box2d.Contact) {
	t.Helper()
	// Body and shape contact views must contain the same live contact exactly once.
	bodyData := slices.Collect(b.ContactData())
	r.I32("body.contact.count", int32(len(bodyData)))

	matches := 0

	for _, d := range bodyData {
		if d.Contact == contact {
			matches++
		}
	}

	r.I32("body.contact.multiplicity", int32(matches))

	if matches != 1 {
		t.Fatalf("body contact multiplicity=%d", matches)
	}

	before := w.StateHash()

	for range s.ContactData() {
		break
	}

	for range b.ContactData() {
		break
	}

	if w.StateHash() != before {
		t.Fatal("early-stop contacts mutated state")
	}
}
