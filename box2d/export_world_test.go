// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"sync"
	"testing"
)

func CleanupWorld(tb testing.TB, id *World) {
	tb.Helper()
	tb.Cleanup(func() {
		if id.IsValid() {
			id.Close()
		}
	})
}

func CleanupReplay(tb testing.TB, player *Replay) func() {
	tb.Helper()

	destroy := sync.OnceFunc(func() {
		if player != nil {
			player.Close()
		}
	})
	tb.Cleanup(destroy)

	return destroy
}

const (
	StaticSet = staticSet
	AwakeSet  = awakeSet
)

func WorldDynamicTree(id *World) *DynamicTree { return &id.broadPhase.trees[DynamicBody] }
func JointSetIndex(w *World, id Joint) int    { return getJointFullID(w, id).setIndex }
func CountEnlargedNodes(tree *DynamicTree) int {
	count := 0

	for i := range tree.nodeEnd {
		if isNodeMoved(&tree.nodes[i]) {
			count++
		}
	}

	return count
}

func ShapeSlot(s Shape) int32 { return s.index1 }

type BodyTransform struct {
	Index     int32
	Transform WorldTransform
}

func BodyTransforms(id *World) []BodyTransform {
	w := id

	var transforms []BodyTransform

	for i := range w.bodies {
		body := &w.bodies[i]
		if body.id != i {
			continue
		}

		transforms = append(transforms, BodyTransform{int32(i), getBodySim(w, body).transform})
	}

	return transforms
}

func MarshalSnapshot(w *World) []byte {
	image, err := w.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return image
}

// FuncDrawer forwards each draw command to its field when set.
type FuncDrawer struct {
	Polygon       func(transform WorldTransform, vertices []Vec2, color HexColor)
	SolidPolygon  func(transform WorldTransform, vertices []Vec2, radius float32, color HexColor)
	Circle        func(center Pos, radius float32, color HexColor)
	SolidCircle   func(transform WorldTransform, center Vec2, radius float32, color HexColor)
	SolidCapsule  func(p1, p2 Pos, radius float32, color HexColor)
	Line          func(p1, p2 Pos, color HexColor)
	TransformFunc func(transform WorldTransform)
	Point         func(p Pos, size float32, color HexColor)
	String        func(p Pos, s string, color HexColor)
	Bounds        func(aabb AABB, color HexColor)
}

func (d *FuncDrawer) DrawPolygon(transform WorldTransform, vertices []Vec2, color HexColor) {
	if d.Polygon != nil {
		d.Polygon(transform, vertices, color)
	}
}

func (d *FuncDrawer) DrawSolidPolygon(
	transform WorldTransform,
	vertices []Vec2,
	radius float32,
	color HexColor,
) {
	if d.SolidPolygon != nil {
		d.SolidPolygon(transform, vertices, radius, color)
	}
}

func (d *FuncDrawer) DrawCircle(center Pos, radius float32, color HexColor) {
	if d.Circle != nil {
		d.Circle(center, radius, color)
	}
}

func (d *FuncDrawer) DrawSolidCircle(
	transform WorldTransform,
	center Vec2,
	radius float32,
	color HexColor,
) {
	if d.SolidCircle != nil {
		d.SolidCircle(transform, center, radius, color)
	}
}

func (d *FuncDrawer) DrawSolidCapsule(p1 Pos, p2 Pos, radius float32, color HexColor) {
	if d.SolidCapsule != nil {
		d.SolidCapsule(p1, p2, radius, color)
	}
}

func (d *FuncDrawer) DrawLine(p1 Pos, p2 Pos, color HexColor) {
	if d.Line != nil {
		d.Line(p1, p2, color)
	}
}

func (d *FuncDrawer) DrawTransform(transform WorldTransform) {
	if d.TransformFunc != nil {
		d.TransformFunc(transform)
	}
}

func (d *FuncDrawer) DrawPoint(p Pos, size float32, color HexColor) {
	if d.Point != nil {
		d.Point(p, size, color)
	}
}

func (d *FuncDrawer) DrawString(p Pos, s string, color HexColor) {
	if d.String != nil {
		d.String(p, s, color)
	}
}

func (d *FuncDrawer) DrawAABB(aabb AABB, color HexColor) {
	if d.Bounds != nil {
		d.Bounds(aabb, color)
	}
}

func TestCleanupWorldSkipped(t *testing.T) {
	var ids [2]*World

	t.Run("skip", func(t *testing.T) {
		def := DefaultWorldDef()
		for i := range ids {
			ids[i] = NewWorld(&def)
			CleanupWorld(t, ids[i])

			if !ids[i].IsValid() {
				t.Fatal("condition false")
			}
		}

		t.SkipNow()
	})

	for _, id := range ids {
		if id.IsValid() {
			t.Fatal("skipped test left a live world")
		}
	}
}

func TestCleanupWorldExplicitDestroy(t *testing.T) {
	var replacement *World

	t.Run("destroy and recycle", func(child *testing.T) {
		def := DefaultWorldDef()
		id := NewWorld(&def)
		CleanupWorld(child, id)

		if !id.IsValid() {
			child.Fatal("condition false")
		}

		id.Close()

		if id.IsValid() {
			child.Fatal("condition false")
		}

		replacement = NewWorld(&def)
		CleanupWorld(t, replacement)

		if !replacement.IsValid() {
			child.Fatal("condition false")
		}
	})

	if !replacement.IsValid() {
		t.Fatal("condition false")
	}

	t.Run("null", func(t *testing.T) {
		CleanupWorld(t, nil)
	})
}
