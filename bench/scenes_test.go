// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package bench

import (
	"fmt"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/scenes"
	"testing"
	_ "unsafe"
)

//go:linkname randomSeed github.com/zigai/box2d/internal/scenes.gRandomSeed
var randomSeed uint32

func BenchmarkScenes(b *testing.B) {
	for _, workers := range []int{1, 4} {
		b.Run(fmt.Sprintf("JointGrid/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateJointGrid(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("LargePyramid/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateLargePyramid(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("ManyPyramids/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateManyPyramids(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Rain/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateRain(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepRain(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Spinner/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateSpinner(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepSpinner(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Smash/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateSmash(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Tumbler/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateTumbler(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Washer/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateWasher(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Junkyard/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateJunkyard(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepJunkyard(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Sleep/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateSleep(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepSleep(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Compounds/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateCompounds(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Queries/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateQueries(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepQueries(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("TreeCast/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateTreeCast(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepTreeCast(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				scenes.DestroyTreeCast()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("TileWorld/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				scenes.CreateTileWorld(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					scenes.StepTileWorld(world, step)
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				scenes.DestroyTileWorld()
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("FallingHinges/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				data := scenes.CreateFallingHinges(world)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					world.Step(float32(1)/float32(60), 4)
					scenes.UpdateFallingHinges(world, &data)
				}
				b.StopTimer()
				scenes.DestroyFallingHinges(&data)
				world.Destroy()
				b.StartTimer()
			}
		})
		b.Run(fmt.Sprintf("Human/workers=%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			seed := randomSeed
			defer func() { randomSeed = seed }()
			for b.Loop() {
				b.StopTimer()
				randomSeed = 12345
				def := box2d.DefaultWorldDef()
				def.WorkerCount = workers
				world := box2d.CreateWorld(&def)
				var human scenes.Human
				scenes.CreateHuman(&human, world, box2d.Pos{X: 0, Y: 5}, 1, 0.05, 5, 0.5, 1, nil, true)
				b.StartTimer()
				for step := 0; step < 300; step++ {
					switch step {
					case 0:
						human.SetVelocity(box2d.Vec2{X: 1})
					case 20:
						human.ApplyRandomAngularImpulse(0.5)
					case 40:
						human.SetJointFrictionTorque(0)
					case 60:
						human.SetJointFrictionTorque(0.1)
					case 80:
						human.SetJointSpringHertz(0)
					case 100:
						human.SetJointSpringHertz(2)
					case 120:
						human.SetJointDampingRatio(0.75)
					case 140:
						human.EnableSensorEvents(true)
					case 160:
						human.SetScale(1.5)
					case 180:
						human.EnableSensorEvents(false)
					}
					world.Step(float32(1)/float32(60), 4)
				}
				b.StopTimer()
				scenes.DestroyHuman(&human)
				world.Destroy()
				b.StartTimer()
			}
		})
	}
}
