package box2d

import "testing"

func TestWorldLookupValidatesSlotAndGeneration(t *testing.T) {
	wd := DefaultWorldDef()
	w := NewWorld(&wd)
	defer w.Close()
	bd := DefaultBodyDef()
	b := w.CreateBody(&bd)
	sd := DefaultShapeDef()
	s := b.CreateCircleShape(&sd, &Circle{Radius: 1})
	if w.Body(b.ID()) != b || w.Shape(s.ID()) != s {
		t.Fatal("live handles did not rebind")
	}
	for _, id := range []uint64{0, 1, ^uint64(0), b.ID() | 1<<16} {
		if w.Body(id).IsValid() || w.Shape(id).IsValid() {
			t.Fatalf("invalid id %x accepted", id)
		}
	}
	bid, sid := b.ID(), s.ID()
	b.Destroy()
	replacement := w.CreateBody(&bd)
	replacement.CreateCircleShape(&sd, &Circle{Radius: 1})
	if w.Body(bid).IsValid() || w.Shape(sid).IsValid() {
		t.Fatal("stale generation accepted")
	}
}

func TestWorldLookupRebindsSnapshotInFreshWorld(t *testing.T) {
	wd := DefaultWorldDef()
	original := NewWorld(&wd)
	defer original.Close()
	bd := DefaultBodyDef()
	b := original.CreateBody(&bd)
	sd := DefaultShapeDef()
	s := b.CreateCircleShape(&sd, &Circle{Radius: 1})
	image, err := original.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := NewWorldFromSnapshot(image, 1)
	if err != nil {
		t.Fatal(err)
	}
	rb, rs := restored.Body(b.ID()), restored.Shape(s.ID())
	if !rb.IsValid() || !rs.IsValid() || rb == b || rs == s || rs.Body() != rb {
		t.Fatal("lookup retained wrong world ownership")
	}
	restored.Close()
	if restored.Body(b.ID()).IsValid() || restored.Shape(s.ID()).IsValid() {
		t.Fatal("closed world accepted handles")
	}
	var absent *World
	if absent.Body(b.ID()).IsValid() || absent.Shape(s.ID()).IsValid() {
		t.Fatal("nil world accepted handles")
	}
}
