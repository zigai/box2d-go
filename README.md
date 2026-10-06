# box2d

Go port of [Box2D](https://github.com/erincatto/box2d) v3 for rigid-body simulation, at the upstream commit in `UPSTREAM`.

- **Native Go:** Simulate bodies, contacts, and joints without a C runtime dependency.
- **Large worlds:** Select double-precision positions with the `box2d_double` build tag.
- **Behavioral comparisons:** Check the port against independently generated observations from the pinned C engine.

## Installation

```sh
go get github.com/zigai/box2d
```

Requires Go 1.27 or newer.

## Experimental SIMD

Default builds use scalar Go and do not depend on experimental SIMD APIs. To try the optional amd64 AVX backend, set the build-time environment variable `GOEXPERIMENT=simd`:

```sh
go test ./...
GOEXPERIMENT=simd go test ./...
GOEXPERIMENT=simd BOX2D_SIMD=0 go test ./...
```

The second command enables SIMD; the third compiles the same experimental backend but selects its scalar fallback. `BOX2D_SIMD=0` is read once when the process starts, so it also works with an already-built experimental executable. Leaving it unset, or setting it to `1`, permits SIMD when available. Setting `BOX2D_SIMD=1` alone cannot enable SIMD in a normal build.

This first backend vectorizes AABB overlap checks used by the dynamic tree and broad phase, not the contact solver. Unsupported CPUs and non-amd64 architectures retain scalar behavior; amd64 SIMD requires AVX, not AVX2. Both position precision modes are supported. Check the selected mode with `GOEXPERIMENT=simd go test -run '^TestSIMDConfiguration$' -v .`.

The Go SIMD API is experimental and may change between toolchain versions. Keep this opt-in for evaluation, and benchmark your own workload rather than assuming every scene becomes faster. For example:

```sh
GOEXPERIMENT=simd BENCH_STEPS=300 GOMAXPROCS=2 go test -run '^$' -bench '^BenchmarkScenes/TreeCast/workers=1$' -benchtime=1x -count=6 -benchmem ./bench
```

Add `BOX2D_SIMD=0` to compare the same experimental build with its fallback; omit `GOEXPERIMENT` to measure a normal scalar build. See `bench/simd_optimization.md` for correctness checks, measurements, and reproduction details.

## Examples

### Step a falling body

```go
package main

import (
	"fmt"

	"github.com/zigai/box2d"
)

func main() {
	definition := box2d.DefaultWorldDef()
	world := box2d.NewWorld(&definition)
	defer world.Close()
	bodyDefinition := box2d.DefaultBodyDef()
	bodyDefinition.Type = box2d.DynamicBody
	bodyDefinition.Position = box2d.Pos{Y: 10}
	body := world.CreateBody(&bodyDefinition)
	shapeDefinition := box2d.DefaultShapeDef()
	circle := box2d.Circle{Radius: 0.5}
	body.CreateCircleShape(&shapeDefinition, &circle)
	for range 60 {
		world.Step(1.0/60, 4)
	}
	fmt.Println(body.Position().Y < 10) // true
}
```

## Conformance checks

Run `go test ./...` for default checks, or `just test` for both float and double precision. Run `just stress` for full-size scene and replay comparisons, or `just test-all` for complete normal and race coverage in both precisions. Tests never regenerate expectations.

Full-size checks require the `box2d_stress` build tag. CI runs default and stress suites explicitly. `test/run.sh` separates one-worker and four-worker scenes into different processes, preserving all cases, scene horizons, assertions, and C fixtures.

Run `just oracle` to explicitly rebuild expectations with `CMake`, GCC, and the exact `UPSTREAM` checkout; review the resulting fixture changes. The generator also runs the unchanged upstream suite and cross-checks SIMD and alternate-compiler profiles. The designated assertion-enabled scalar reference disables floating-point contraction. Requested distance-simplex histories use a separate Debug C profile.

Run `just oracle-check` after generation to execute the original C suites again, compare fresh operation histories, and reconcile the [case inventory](oracle/cases.json) with collected Go results. It rejects missing cases, applicable skips, and orphaned C fixtures. `ShapeCastShortEdgeTest` remains explicitly upstream-disabled; the unregistered recording diagnostic is tracked separately.

Explore additional histories with `go test -tags box2d_oracle ./test/integration -run '^$' -fuzz '^FuzzDifferentialSequence$' -fuzztime=30s`; add `box2d_double` to the tags for double precision. Failed inputs are retained by Go's fuzz runner.

Snapshots and recordings are created and consumed within each engine, not treated as interchangeable C/Go memory images. Named cases and profiles establish finite agreement, not universal equivalence. The C memory-stat dump has no Go counterpart and is not claimed as parity coverage.

## Benchmarks

Scene benchmarks default to 300 simulation steps per workload. Set `BENCH_STEPS` to an integer from 1 to 2147483647 to change the horizon for every selected scene:

```sh
BENCH_STEPS=500 GOMAXPROCS=4 go test -run '^$' -bench '^BenchmarkScenes$' -benchtime=1x -count=6 -benchmem ./bench
```

Select a particular scene and worker count with a benchmark filter, for example `-bench '^BenchmarkScenes/JointGrid/workers=4$'`. Each reported operation is one complete workload of `BENCH_STEPS` steps, not one simulation step. Setup and teardown are excluded from the Go benchmark timer. The configuration header records the actual step count; the timestep, substep count, and seed remain unchanged.

For matched Go/C measurements, first build the pinned C benchmark with `just oracle-bench`, then run:

```sh
BENCH_STEPS=500 BENCH_SAMPLES=6 GOMAXPROCS=4 ./bench/run.sh
```

The comparison script requires `benchstat`, passes the same step count to both engines, and verifies both configuration headers before comparison. `BENCH_SAMPLES` defaults to 6; `BENCH_OUT` selects the output directory. The C benchmark also accepts an optional fourth argument: `bench_scenes [samples] [workers] [cpus] [steps]`. Standalone measurement verification accepts `go run ./bench/verify -steps 500 <go-output> <c-output>`, with matching sample, CPU, and precision flags when needed.

Changing the horizon alone does not reproduce the website's benchmark setup, hardware, SIMD backend, or warm-up policy.

## License

[MIT](LICENSE)
