//go:build box2d_stress

// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"fmt"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func TestSceneHelperBoundaries(t *testing.T) {
	rain, queries, seed := gRainData, gQueryBenchmark, gRandomSeed

	t.Cleanup(func() { gRainData, gQueryBenchmark, gRandomSeed = rain, queries, seed })

	gRandomSeed = RandSeed
	def := box2d.DefaultWorldDef()
	cleanupWorld := func(id *box2d.World) {
		t.Cleanup(func() {
			if id.IsValid() {
				id.Close()
			}
		})
	}
	world := box2d.NewWorld(&def)
	cleanupWorld(world)

	r := oracle.Open(t, "scenes/Helpers")
	counts := func(step int) {
		t.Helper()

		c := world.Counters()
		r.I32(fmt.Sprintf("%d.b", step), int32(c.BodyCount))
		r.I32(fmt.Sprintf("%d.s", step), int32(c.ShapeCount))
		r.I32(fmt.Sprintf("%d.j", step), int32(c.JointCount))
	}
	flags := func(step int) { observeRainHelperFlags(t, r, step) }

	CreateRain(world)
	counts(-1)

	for row := range rainRowCount {
		CreateGroup(world, row, rainColumnCount-1)
	}

	counts(0)
	flags(0)

	oldBody := gRainData.groups[rainColumnCount-1].humans[0].Bones[BoneTorso].BodyID

	oldJoint := gRainData.groups[rainColumnCount-1].humans[0].Bones[BoneTorso].JointID
	for row := range rainRowCount {
		DestroyGroup(row, rainColumnCount-1)
	}

	counts(1)
	flags(1)
	r.Bool("1.oldBody", oldBody.IsValid())
	r.Bool("1.oldJoint", oldJoint.IsValid())

	for row := range rainRowCount {
		CreateGroup(world, row, rainColumnCount-1)
	}

	gRainData.columnCount, gRainData.columnIndex = rainColumnCount, rainColumnCount-1
	oldBody = gRainData.groups[rainColumnCount-1].humans[0].Bones[BoneTorso].BodyID
	oldJoint = gRainData.groups[rainColumnCount-1].humans[0].Bones[BoneTorso].JointID

	r.F32("2.v", StepRain(world, 320))
	counts(2)
	flags(2)
	r.Bool("2.oldBody", oldBody.IsValid())
	r.Bool("2.oldJoint", oldJoint.IsValid())

	if gRainData.columnIndex != 0 || gRainData.columnCount != rainColumnCount {
		t.Fatal("condition false")
	}

	world.Close()
	world = box2d.NewWorld(&def)
	cleanupWorld(world)
	CreateQueries(world)

	positions := [4]float32{gQueryBenchmark.paddleMaxX + 1, gQueryBenchmark.paddleMaxX, gQueryBenchmark.paddleMinX - 1, gQueryBenchmark.paddleMinX}
	for i, x := range positions {
		velocity := float32(5)
		if i >= 2 {
			velocity = -5
		}

		gQueryBenchmark.paddleID.SetTransform(box2d.Vec2{X: x}.Pos(), box2d.Rot{C: 1})
		gQueryBenchmark.paddleID.SetLinearVelocity(box2d.Vec2{X: velocity})
		r.F32(fmt.Sprintf("%d.q", i), StepQueries(world, i))

		p := gQueryBenchmark.paddleID.Position()
		v := gQueryBenchmark.paddleID.LinearVelocity()

		r.Pos(fmt.Sprintf("%d.position", i), p.X, p.Y)
		r.Vec2(fmt.Sprintf("%d.velocity", i), v.X, v.Y)

		s := GetQueryBenchmarkStats()
		r.I32(fmt.Sprintf("%d.n", i), int32(s.NodeVisits))
		r.I32(fmt.Sprintf("%d.l", i), int32(s.LeafVisits))

		want := [4]float32{-5, 5, 5, -5}
		if !(v == (box2d.Vec2{X: want[i]})) {
			t.Fatal("condition false")
		}
	}
}

func observeRainHelperFlags(t *testing.T, r *oracle.Reader, step int) {
	t.Helper()

	spawned, cleared := true, true

	for row := range rainRowCount {
		for i := range rainGroupSize {
			h := &gRainData.groups[row*rainColumnCount+rainColumnCount-1].humans[i]

			spawned = spawned && h.IsSpawned
			for _, b := range h.Bones {
				cleared = cleared && b.BodyID == (box2d.Body{}) && b.JointID == (box2d.Joint{})
			}
		}
	}

	r.Bool(fmt.Sprintf("%d.spawned", step), spawned)
	r.Bool(fmt.Sprintf("%d.cleared", step), cleared)
	r.I32(fmt.Sprintf("%d.column", step), int32(gRainData.columnIndex))

	if spawned != (step != 1) || cleared != (step == 1) {
		t.Fatal("condition false")
	}
}
