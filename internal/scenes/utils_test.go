// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
)

func TestRandomHelpers(t *testing.T) {
	defer porttest.Guard(t)
	seed := gRandomSeed
	t.Cleanup(func() { gRandomSeed = seed })
	gRandomSeed = randSeed
	var r *oracle.Reader
	porttest.Ensure(t, GetNumberOfCores() > 0)
	for step := 0; step < 300; step++ {
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
		for i := 0; i < polygon.Count; i++ {
			r.Vec2(fmt.Sprintf("%d.vertex", step), polygon.Vertices[i].X, polygon.Vertices[i].Y)
			r.Vec2(fmt.Sprintf("%d.normal", step), polygon.Normals[i].X, polygon.Normals[i].Y)
		}
	}
}
