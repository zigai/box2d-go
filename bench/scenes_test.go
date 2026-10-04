// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package bench

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/scenes"
)

//go:linkname randomSeed github.com/zigai/box2d/internal/scenes.gRandomSeed
var randomSeed uint32

func BenchmarkScenes(b *testing.B) {
	// This deliberately matched custom workload times all 300 steps from step 0,
	// excluding construction/destruction. It is not the pinned CLI warm-up/horizon.
	precision := "float"
	if unsafe.Sizeof(box2d.Pos{}) == 16 {
		precision = "double"
	}

	fmt.Fprintf(os.Stdout, "box2d-bench: precision=%s scene-config=release steps=300 timestep=1/60 substeps=4 seed=12345 continuous=true rebuild-static=false cpus=%d\n", precision, runtime.GOMAXPROCS(0))

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
				world := box2d.NewWorld(&def)
				scenes.CreateJointGrid(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateLargePyramid(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateManyPyramids(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateRain(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepRain(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateSpinner(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepSpinner(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateSmash(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateTumbler(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateWasher(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateJunkyard(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepJunkyard(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateSleep(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepSleep(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateCompounds(world)
				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateQueries(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepQueries(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateTreeCast(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepTreeCast(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				scenes.DestroyTreeCast()
				world.Close()
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
				world := box2d.NewWorld(&def)
				scenes.CreateTileWorld(world)
				b.StartTimer()

				for step := range 300 {
					scenes.StepTileWorld(world, step)
					world.Step(float32(1)/float32(60), 4)
				}

				b.StopTimer()
				scenes.DestroyTileWorld()
				world.Close()
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
				world := box2d.NewWorld(&def)
				data := scenes.CreateFallingHinges(world)

				b.StartTimer()

				for range 300 {
					world.Step(float32(1)/float32(60), 4)
					scenes.UpdateFallingHinges(world, &data)
				}

				b.StopTimer()
				scenes.DestroyFallingHinges(&data)
				world.Close()
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
				world := box2d.NewWorld(&def)

				var human scenes.Human
				scenes.CreateHuman(&human, world, box2d.Pos{X: 0, Y: 5}, 1, 0.05, 5, 0.5, 1, nil, true)
				b.StartTimer()

				for step := range 300 {
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
				world.Close()
				b.StartTimer()
			}
		})
	}
}
