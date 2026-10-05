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

## License

[MIT](LICENSE)
