// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"sync"
	"testing"
)

func BTWorldCleanup(tb testing.TB, id *World) {
	tb.Helper()
	tb.Cleanup(func() {
		if id.IsValid() {
			id.Close()
		}
	})
}

func BTReplayCleanup(tb testing.TB, player *Replay) func() {
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
	WorldStaticSet = staticSet
	WorldAwakeSet  = awakeSet
)

func WorldDynamicTree(id *World) *DynamicTree   { return &id.broadPhase.trees[DynamicBody] }
func WorldJointSetIndex(w *World, id Joint) int { return getJointFullID(w, id).setIndex }
func WorldCountEnlargedNodes(tree *DynamicTree) int {
	count := 0

	for i := range int32(tree.nodeEnd) {
		if isNodeMoved(&tree.nodes[i]) {
			count++
		}
	}

	return count
}

func WorldShapeIndex(s Shape) int32 { return s.index1 }

type WorldBodyTransform struct {
	Index     int32
	Transform WorldTransform
}

func WorldBodyTransforms(id *World) []WorldBodyTransform {
	w := id

	var transforms []WorldBodyTransform

	for i := range w.bodies {
		body := &w.bodies[i]
		if body.id != i {
			continue
		}

		transforms = append(transforms, WorldBodyTransform{int32(i), getBodySim(w, body).transform})
	}

	return transforms
}

func BTSnapshot(w *World) []byte {
	image, err := w.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return image
}

// BTFuncDrawer forwards each draw command to its field when set.
type BTFuncDrawer struct {
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

func (d *BTFuncDrawer) DrawPolygon(transform WorldTransform, vertices []Vec2, color HexColor) {
	if d.Polygon != nil {
		d.Polygon(transform, vertices, color)
	}
}

func (d *BTFuncDrawer) DrawSolidPolygon(transform WorldTransform, vertices []Vec2, radius float32, color HexColor) {
	if d.SolidPolygon != nil {
		d.SolidPolygon(transform, vertices, radius, color)
	}
}

func (d *BTFuncDrawer) DrawCircle(center Pos, radius float32, color HexColor) {
	if d.Circle != nil {
		d.Circle(center, radius, color)
	}
}

func (d *BTFuncDrawer) DrawSolidCircle(transform WorldTransform, center Vec2, radius float32, color HexColor) {
	if d.SolidCircle != nil {
		d.SolidCircle(transform, center, radius, color)
	}
}

func (d *BTFuncDrawer) DrawSolidCapsule(p1, p2 Pos, radius float32, color HexColor) {
	if d.SolidCapsule != nil {
		d.SolidCapsule(p1, p2, radius, color)
	}
}

func (d *BTFuncDrawer) DrawLine(p1, p2 Pos, color HexColor) {
	if d.Line != nil {
		d.Line(p1, p2, color)
	}
}

func (d *BTFuncDrawer) DrawTransform(transform WorldTransform) {
	if d.TransformFunc != nil {
		d.TransformFunc(transform)
	}
}

func (d *BTFuncDrawer) DrawPoint(p Pos, size float32, color HexColor) {
	if d.Point != nil {
		d.Point(p, size, color)
	}
}

func (d *BTFuncDrawer) DrawString(p Pos, s string, color HexColor) {
	if d.String != nil {
		d.String(p, s, color)
	}
}

func (d *BTFuncDrawer) DrawAABB(aabb AABB, color HexColor) {
	if d.Bounds != nil {
		d.Bounds(aabb, color)
	}
}
