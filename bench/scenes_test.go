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
	steps, err := parseBenchmarkSteps(os.Getenv("BENCH_STEPS"))
	if err != nil {
		b.Fatal(err)
	}

	precision := "float"
	if unsafe.Sizeof(box2d.Pos{}) == 16 {
		precision = "double"
	}

	if _, err := fmt.Fprintf(os.Stdout, "box2d-bench: precision=%s scene-config=release steps=%d timestep=1/60 substeps=4 seed=12345 continuous=true rebuild-static=false cpus=%d\n", precision, steps, runtime.GOMAXPROCS(0)); err != nil {
		b.Fatalf("write benchmark configuration: %v", err)
	}

	for _, workers := range []int{1, 4} {
		b.Run(fmt.Sprintf("JointGrid/workers=%d", workers), func(b *testing.B) {
			benchmarkJointGrid(b, workers, steps)
		})
		b.Run(fmt.Sprintf("LargePyramid/workers=%d", workers), func(b *testing.B) {
			benchmarkLargePyramid(b, workers, steps)
		})
		b.Run(fmt.Sprintf("ManyPyramids/workers=%d", workers), func(b *testing.B) {
			benchmarkManyPyramids(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Rain/workers=%d", workers), func(b *testing.B) {
			benchmarkRain(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Spinner/workers=%d", workers), func(b *testing.B) {
			benchmarkSpinner(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Smash/workers=%d", workers), func(b *testing.B) {
			benchmarkSmash(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Tumbler/workers=%d", workers), func(b *testing.B) {
			benchmarkTumbler(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Washer/workers=%d", workers), func(b *testing.B) {
			benchmarkWasher(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Junkyard/workers=%d", workers), func(b *testing.B) {
			benchmarkJunkyard(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Sleep/workers=%d", workers), func(b *testing.B) {
			benchmarkSleep(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Compounds/workers=%d", workers), func(b *testing.B) {
			benchmarkCompounds(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Queries/workers=%d", workers), func(b *testing.B) {
			benchmarkQueries(b, workers, steps)
		})
		b.Run(fmt.Sprintf("TreeCast/workers=%d", workers), func(b *testing.B) {
			benchmarkTreeCast(b, workers, steps)
		})
		b.Run(fmt.Sprintf("TileWorld/workers=%d", workers), func(b *testing.B) {
			benchmarkTileWorld(b, workers, steps)
		})
		b.Run(fmt.Sprintf("FallingHinges/workers=%d", workers), func(b *testing.B) {
			benchmarkFallingHinges(b, workers, steps)
		})
		b.Run(fmt.Sprintf("Human/workers=%d", workers), func(b *testing.B) {
			benchmarkHuman(b, workers, steps)
		})
	}
}

func benchmarkJointGrid(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkLargePyramid(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkManyPyramids(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkRain(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepRain(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkSpinner(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepSpinner(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkSmash(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkTumbler(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkWasher(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkJunkyard(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepJunkyard(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkSleep(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepSleep(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkCompounds(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkQueries(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepQueries(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkTreeCast(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepTreeCast(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		scenes.DestroyTreeCast()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkTileWorld(b *testing.B, workers, steps int) {
	b.Helper()
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

		for step := range steps {
			scenes.StepTileWorld(world, step)
			world.Step(float32(1)/float32(60), 4)
		}

		b.StopTimer()
		scenes.DestroyTileWorld()
		world.Close()
		b.StartTimer()
	}
}

func benchmarkFallingHinges(b *testing.B, workers, steps int) {
	b.Helper()
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

		for range steps {
			world.Step(float32(1)/float32(60), 4)
			scenes.UpdateFallingHinges(world, &data)
		}

		b.StopTimer()
		scenes.DestroyFallingHinges(&data)
		world.Close()
		b.StartTimer()
	}
}

func benchmarkHuman(b *testing.B, workers, steps int) {
	b.Helper()
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

		runHumanWorkload(world, &human, steps)

		b.StopTimer()
		scenes.DestroyHuman(&human)
		world.Close()
		b.StartTimer()
	}
}

func runHumanWorkload(world *box2d.World, human *scenes.Human, steps int) {
	for step := range steps {
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
}
