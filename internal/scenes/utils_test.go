// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

func TestRandomHelpers(t *testing.T) {
	seed := gRandomSeed

	t.Cleanup(func() { gRandomSeed = seed })

	gRandomSeed = RandSeed

	var r *oracle.Reader

	if GetNumberOfCores() <= 0 {
		t.Fatal("condition false")
	}

	for step := range 300 {
		integer := RandomInt()
		integerRange := RandomIntRange(-7, 13)
		scalar := RandomFloat()
		scalarRange := RandomFloatRange(-0.75, 2.5)
		vec := RandomVec2(-1.5, 3)
		pos := RandomPos(-2, 4)
		rot := RandomRot()
		polygon := RandomPolygon(1)

		if r == nil {
			r = oracle.Open(t, "scenes/Utils")
		}

		r.I32(fmt.Sprintf("%d.int", step), int32(integer))
		r.I32(fmt.Sprintf("%d.range", step), int32(integerRange))
		r.F32(fmt.Sprintf("%d.float", step), scalar)
		r.F32(fmt.Sprintf("%d.floatRange", step), scalarRange)
		r.Vec2(fmt.Sprintf("%d.vec", step), vec.X, vec.Y)
		r.Pos(fmt.Sprintf("%d.pos", step), pos.X, pos.Y)
		r.Rot(fmt.Sprintf("%d.rot", step), rot.C, rot.S)
		r.I32(fmt.Sprintf("%d.count", step), int32(polygon.Count))
		r.F32(fmt.Sprintf("%d.radius", step), polygon.Radius)
		r.Vec2(fmt.Sprintf("%d.centroid", step), polygon.Centroid.X, polygon.Centroid.Y)

		for i := range polygon.Count {
			r.Vec2(fmt.Sprintf("%d.vertex", step), polygon.Vertices[i].X, polygon.Vertices[i].Y)
			r.Vec2(fmt.Sprintf("%d.normal", step), polygon.Normals[i].X, polygon.Normals[i].Y)
		}
	}

	gRandomSeed = RandSeed

	r.I32("equal.int", int32(RandomIntRange(7, 7)))
	r.F32("equal.float", RandomFloatRange(1.25, 1.25))

	v := RandomVec2(-2, -2)
	r.Vec2("equal.vec", v.X, v.Y)

	p := RandomPos(0.5, 0.5)
	r.Pos("equal.pos", p.X, p.Y)
	r.U32("equal.seed", gRandomSeed)
	gRandomSeed = 0

	fallback := RandomPolygon(1)
	if fallback.Count != 4 {
		t.Fatal("degenerate random polygon did not take square fallback")
	}

	r.I32("fallback.count", int32(fallback.Count))
	r.F32("fallback.radius", fallback.Radius)
	r.Vec2("fallback.centroid", fallback.Centroid.X, fallback.Centroid.Y)

	for i := range fallback.Count {
		r.Vec2("fallback.vertex", fallback.Vertices[i].X, fallback.Vertices[i].Y)
		r.Vec2("fallback.normal", fallback.Normals[i].X, fallback.Normals[i].Y)
	}
}
