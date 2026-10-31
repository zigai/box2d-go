# Box2D port conventions

`github.com/zigai/box2d-go/box2d` is a Go port of the Box2D C engine. The port has two
goals: it must produce the same bits as the C engine, and it must stay easy to
compare with upstream when upstream changes. These conventions explain how C patterns map to Go so
that a change you make keeps both goals. Read them before you change the
port.

## Concepts

*Upstream* is the Box2D C source at the commit pinned in `UPSTREAM`. It is
version 3.2.0 and includes snapshot, recording and replay. Box2D is MIT
licensed, Copyright (c) 2023 Erin Catto.

*Bit-exact* means that every float the port computes has the same bits as
upstream built with `-ffp-contract=off` and scalar math, in both the float
and double configurations. The tests check this against *oracle fixtures*:
values captured from the pinned C engine and stored in `testdata/oracle/`.

Most conventions follow from these two ideas. Go code reads like the C file it
came from, so a diff against upstream stays small. Where Go and C disagree on
float behavior, the port spells out the C behavior. The module needs Go 1.27
or newer.

## Names

The public API is one package, `box2d`. Upstream's standalone helpers live
in packages under `internal/`, described in
[Repository layout](#repository-layout). Public names
are upstream's names without the `b2` or `B2_` prefix, in Go capitalization.
The following table shows common cases:

| Upstream | Go |
|---|---|
| `b2Vec2`, `b2Transform`, `b2Polygon` | `Vec2`, `Transform`, `Polygon` |
| `b2BodyId`, `b2ShapeId`, `b2WorldId` | `Body`, `Shape`, `*World` |
| `b2BodyDef`, `b2DefaultBodyDef` | `BodyDef`, `DefaultBodyDef` |
| `b2CreateWorld`, `b2DestroyWorld` | `NewWorld`, `(*World).Close` |
| `b2CreateBody( worldId, def )` | `world.CreateBody(def)` |
| `b2CreateCircleShape( bodyId, def, circle )` | `body.CreateCircleShape(def, circle)` |
| `b2DestroyBody( bodyId )` | `body.Destroy()` |
| `b2Body_SetLinearVelocity`, `b2Body_GetLinearVelocity` | `Body.SetLinearVelocity`, `Body.LinearVelocity` |
| `b2Body_GetShapes` | `Body.Shapes`, an `iter.Seq` |
| `b2Add`, `b2MulSV( s, v )`, `b2RotateVector( q, v )` | `a.Add(b)`, `v.Scale(s)`, `q.Rotate(v)` |
| `b2TransformPoint( t, p )`, `b2IntegrateRotation` | `t.Apply(p)`, `Rot.Integrate` |
| `b2GetVersion`, `b2GetSweepTransform` | `CurrentVersion`, `(*Sweep).Transform` |
| `b2_staticBody`, `b2_distanceJoint` | `StaticBody`, `DistanceJointType` |
| `B2_HUGE`, `B2_LINEAR_SLOP` | `Huge()`, `LinearSlop()` |

The table follows these rules:

- A function named `b2<Type>_<Method>` becomes a method on the Go type. A
  `b2Create*` function that takes a handle becomes a method on that handle.
- A math function whose first argument is a value type becomes a method on
  that type, often with a shorter name, as in `Vec2.Scale` and
  `Transform.Apply`. Free math functions that build a value stay functions,
  such as `MakeRot`, `MakeBox` and `Atan2`.
- A function that returns a new object the caller owns uses `New`:
  `NewWorld`, `NewDynamicTree`, `NewRecording`, `NewReplay`. Objects that a
  world owns keep `Create`.
- A leading `Get` is dropped from getters. `Is`, `Are`, `Set`, `Enable` and
  `Disable` are kept, as in `Body.IsAwake`.
- `Id` becomes `ID`, and callback types end in `Func`: `b2CustomFilterFcn`
  becomes `CustomFilterFunc`.
- An enum constant `b2_<name>` becomes `<Name>`. If that name is taken by a
  type, the enum type's name is added: `b2_distanceJoint` becomes
  `DistanceJointType`, because `DistanceJoint` is a handle. Numeric values
  stay the same, because recordings and hashes depend on them.
- A macro that depends on `b2_lengthUnitsPerMeter`, such as `B2_HUGE`, becomes
  a function.
- Public data structs, such as defs, events, geometry and results, have
  exported fields: `Vec2.X`, `BodyDef.GravityScale`.

Everything else is unexported. An internal name is the upstream name without
`b2`, with underscores removed and the first letter lower-cased: `b2BodySim`
is `bodySim`, `b2_awakeSet` is `awakeSet`. Internal struct fields keep
upstream's camelCase, as in `bodySim.center`. Functions in `internal/scenes`
keep upstream's exported names, such as `CreateJunkyard`.

## Worlds and handles

`NewWorld` returns a `*World`. A world is used from one goroutine, plus its
own workers. There is no global world table.

`Body`, `Shape`, `Chain`, `Joint` and `Contact` are small value structs that
hold upstream's `index1` and `generation` fields plus the owning `*World`.
They are comparable with `==`, and the zero value refers to nothing. Typed
joint handles such as `DistanceJoint` embed `Joint`.

`IsValid` on a handle is a port of `b2Body_IsValid` and friends. It returns
false for a zero handle, an out-of-range slot, a free slot or a stale
generation, and never panics. Any other method on an invalid handle asserts,
in the same place where upstream's lookup functions assert. Don't add checks
that upstream lacks.

To store a handle outside the world, use its `ID` method. It returns a
`uint64` built from the slot and generation. `World.Body(id)` and
`World.Shape(id)` turn it back into a handle, or return the zero handle when
the ID is stale. An ID is local to its world: resolve it in the world that
produced it, or in a world restored from that world's snapshot.

## Defs and the cookie

Defs are exported structs. `DefaultBodyDef`, `DefaultShapeDef`,
`DefaultWorldDef` and the other `Default*Def` functions return upstream's
defaults and set an unexported `internalValue` field to `1152023`
(`B2_SECRET_COOKIE`). Every constructor that takes a def checks the cookie
first, so a def that did not come from a `Default*Def` function is rejected.
Joint defs hold the shared fields in a `Base JointDef` field.

The following example creates a dynamic body:

```go
def := box2d.DefaultBodyDef()
def.Type = box2d.DynamicBody
def.Position = box2d.Pos{X: 0, Y: 4}
body := world.CreateBody(&def)
```

The constructor copies what it needs, so you can reuse `def` for the next
body.

We don't make the zero value a valid def, because upstream's defaults are not
zero: gravity scale, density, friction and sleep thresholds all have nonzero
defaults. Recordings and the replay reader also start from `Default*Def` and
overlay fields. A zero def that quietly simulates differently is the bug the
cookie exists to catch.

## Callbacks, user data, events and queries

### Callbacks

A C callback with a `void* context` becomes a Go func value, and the closure
replaces the context. The `context` parameter is gone from every signature.
Func types are named, so upstream's thread-safety warnings stay on the type's
doc comment.

```go
world.SetCustomFilterCallback(func(a, b box2d.Shape) bool { return a != b })
```

Upstream's `enqueueTask`, `finishTask` and `userTaskContext` become one
`WorldDef.EnqueueTask` field of type `EnqueueTaskFunc`. It runs a task now or
later, and returns a function that waits for it, or nil if the task already
ran. When `EnqueueTask` is nil and `WorkerCount` is greater than 1, the world
runs its own goroutine scheduler. With one worker, tasks run inline.

### User data

`void* userData` becomes `any`, in defs and in `SetUserData` and `UserData`.
It is never serialized, hashed or recorded. Upstream writes it as `NULL` in
snapshot images, so a restored body reads back `nil`. `UnmarshalBinary` keeps
the world's own callbacks and user data, as upstream keeps host wiring.
`NewWorldFromSnapshot` starts with none.

### Events

Each event array is one `iter.Seq` method on `World`: `BodyMoveEvents`,
`SensorBeginEvents`, `SensorEndEvents`, `ContactBeginEvents`,
`ContactEndEvents`, `ContactHitEvents` and `JointEvents`. They yield values in
upstream's order, and the values are valid until the next `Step`. The
internal double buffering of end events is unchanged.

```go
world.Step(1.0/60, 4)
for e := range world.ContactBeginEvents() {
	handleTouch(e.ShapeA, e.ShapeB)
}
```

An iterator makes the lifetime clear: a caller can't hold on to a slice that
the next step overwrites.

### Queries

`OverlapAABB`, `OverlapShape`, `CastRay`, `CastShape` and `CollideMover` keep
upstream's callback form and return `TreeStats`. A cast callback returns a
clip fraction (`-1` to filter, `0` to stop, or a fraction to clip), which an
iterator cannot express. The benchmark scenes read the `TreeStats`.
`CastRayClosest` returns a `RayResult`.

Upstream getters that fill a caller's `(array, capacity)` and return a count
become `iter.Seq` methods: `Body.Shapes`, `Body.Joints`, `Body.ContactData`,
`Shape.ContactData`, `Shape.SensorData` and `Chain.Segments`. The count and
capacity getters stay.

```go
for s := range body.Shapes() {
	s.SetFriction(0.2)
}
```

## Failure handling

Upstream has several check macros. Each keeps its place in the code, and its
message is upstream's condition text, so you can search for the message and
find the C line.

| Upstream | Go | Behavior |
|---|---|---|
| `B2_ASSERT( c )` | `assert(c, "c")` | Panics with `*AssertError` when false. Removed by the `box2d_noassert` tag. |
| `B2_CHECK_INPUT( c )`, `B2_CHECK_DEF( d )` | `checkInput(c, "c")` | Panics with `*AssertError` when false, in every build. |
| `B2_VALIDATE( c )` | `validate(c, "c")` | Like `assert`, but only with the `box2d_validate` tag. |

Bad input is a programmer error, so it panics instead of returning a Go
`error`. Adding `error` returns to hundreds of methods would change every call
site for a case upstream treats as a bug. `checkInput` ignores
`box2d_noassert`, so turning off internal asserts never lets bad input
through. There are no assert or log hooks. Simulation warnings, such as an
unsupported callback while recording, go to `WorldDef.Logger`, which defaults
to `slog.Default()`.

Functions that read files or untrusted bytes return an `error`:
`NewWorldFromSnapshot`, `World.UnmarshalBinary`, `ReadRecording`,
`Recording.WriteTo`, `NewReplay` and `ValidateReplay`. A bad snapshot returns
a `*SnapshotError` that wraps a sentinel such as `ErrBadMagic`, `ErrVersion`,
`ErrLayout` or `ErrCorrupt`. `ValidateReplay` returns `ErrDiverged` when a
replay does not match.

Snapshot images are written field by field, little-endian and in a fixed
order. They round-trip in Go but are not byte-compatible with C images,
because Go can't copy structs that hold slices and `any` as raw memory. They
carry the same state, and `World.StateHash` matches the oracle.

## The float rule

This rule is what makes the port bit-exact. Go and C round floats
differently in a few places, and Go's compiler may fuse a multiply and an add
into one instruction with one rounding. Each rule writes the C behavior out
explicitly:

1. **Use float32 where upstream uses `float`**, and `float64` where upstream
   uses `double`, as in `RoundDownFloat`.
2. **Round every product that feeds an addition or subtraction.** In C,
   `a*b + c` rounds twice. Go may fuse it into one rounding. Write
   `float32(a*b) + c`. The same applies to `c - a*b`, `x += a*b` and sums of
   products: `float32(a*b) + float32(c*d)`. Assigning the product to a local
   variable does not prevent fusion. A product that feeds only a comparison,
   a multiplication or a division needs no rounding. The `fmaround` analyzer
   in `internal/lint` flags unrounded forms, and `just lint` runs it.
3. **Use upstream's approximations, not libm.** Use the ported
   `ComputeCosSin`, `Atan2`, `MakeRot`, `Rot.NLerp`, `Rot.Integrate` and
   `UnwindAngle`. Never call `math.Sin`, `math.Cos`, `math.Atan2` or
   `math.Pow`. `float32(math.Sqrt(float64(x)))` is the one allowed libm call,
   because it is correctly rounded for float32 inputs.
4. **Use typed constants.** Go folds an untyped constant expression exactly,
   while C folds `float` literals one operation at a time. Write
   `const x float32 = 0.1` and compute with typed operands.
5. **Don't use builtin `min`, `max` or `math.Abs` on floats** in simulation
   code. They differ from `b2MinFloat`, `b2MaxFloat` and `b2AbsFloat` on `-0`
   and NaN. Use the ported `minFloat`, `maxFloat` and `absFloat`. The
   builtins are fine for integers.
6. **Don't convert out-of-range floats to integers.** The result is
   implementation-specific in Go and undefined in C. Port upstream's casts as
   written and don't add new ones.
7. **Don't use `math.FMA`**, or any libm function other than `Sqrt`.
8. **Keep the order of operations from the C source.** Don't reassociate,
   factor or simplify.

The following example shows rules 2 and 4:

```go
// upstream: float d = a.x * b.x + a.y * b.y;
d := float32(a.X*b.X) + float32(a.Y*b.Y)

// upstream: pi2 = pi * pi
const pi float32 = 3.14159265359
pi2 := float32(pi * pi)
```

The dot product rounds both products, because the compiler would otherwise
fuse one of them into the add. `pi` is typed, so `pi * pi` rounds like the C
expression.

The test `TestMathPublicSpecialAngles` in `math_public_test.go` checks the
rule against the `make_rot` fixture. Replacing its roundings with fused
operations changes 16 of its 166 results.

One exception lives in `internal/scenes`. The benchmark scenes call C `sinf`
and `cosf` to build their inputs, not the simulation's approximations.
`boundedSinCos` reproduces them with published Arm coefficients, float32 input
and output, and explicitly rounded float64 intermediates. It is scoped to the
finite TileWorld and Washer inputs, and its tests compare all 150,000 ray
inputs and the boundary cases with C. Don't use it in simulation code. Its MIT
notice is in the repository's `LICENSE`.

## Build tags

The package has three tags that change what it computes or checks, plus two
for tests:

- `box2d_double` selects upstream's `BOX2D_DOUBLE_PRECISION`. It changes
  exactly what upstream changes: `Pos` and `WorldTransform`, and the functions
  on them, use `float64`. `Vec2`, `Rot`, `Transform` and all local math stay
  `float32`.
- `box2d_validate` turns on `B2_VALIDATE`, upstream's `BOX2D_VALIDATE`.
- `box2d_noassert` turns off internal asserts. It doesn't change any
  computed value, and input checks still run.
- `box2d_stress` adds the full-size scene and replay comparisons.
- `box2d_oracle` builds the live C comparisons in `test/integration`.

Precision is selected by file pairs with `//go:build box2d_double` and
`//go:build !box2d_double`. `pos.go` defines `Pos` as a struct of `posScalar`,
and `pos_float.go` and `pos_double.go` set `posScalar` to `float32` or
`float64`. `IsDoublePrecision` reports the build. Upstream's link-time guard,
`b2CreateWorldDoublePrecision`, has no Go equivalent, because a build tag
can't mismatch. Snapshot images record the build's precision, and loading an
image from the other precision returns an error.

```go
//go:build !box2d_double

package box2d

type posScalar = float32
```

`just test` runs every test in both precisions.

## Sorting

`qsort.h` is ported as `internal/qsort`, with its Alexey Tourbin MIT notice.
`broad_phase.c` sorts pair keys with it. Quicksort isn't stable, so another
algorithm could order equal keys differently, change the pair order and
change the simulation.

`sensor.c` uses the C library's `qsort` with `b2CompareVisitors`. Equal
shape IDs are the same shape, and the loop that follows removes duplicates,
so the order of equal keys doesn't matter there. That call uses
`slices.SortFunc`. Don't call `sort` or `slices` sorting anywhere else that
upstream sorts.

## SIMD and the wide contact solver

`simd.h` holds the AABB overlap helpers for the dynamic tree and broad phase.
Its scalar branch is the default, in `simd_scalar.go`, with
`type aabbV = AABB`. An optional amd64 AVX backend in `simd_accel_amd64.go`
builds only with `GOEXPERIMENT=simd`, and `BOX2D_SIMD=0` selects the scalar
fallback at run time. [SIMD notes](simd.md) covers how to try it.

The contact solver's `b2FloatW` layer is the scalar struct `floatW` in
`contact_solver.go`, with four lanes processed one after another.
`simdWidth` is 4, the width of SSE2, NEON and WASM. Contacts group into the
same batches as every non-AVX2 upstream build, so the oracle's SIMD
cross-check applies. Each `b2MulAddW` and `b2MulSubW` helper is `a + b*c` in
C, so in Go it is `a + float32(b*c)`.

```go
func mulAddW(a, b, c floatW) floatW {
	return floatW{a.x + float32(b.x*c.x), a.y + float32(b.y*c.y), a.z + float32(b.z*c.z), a.w + float32(b.w*c.w)}
}
```

## Containers, memory and threads

`b2Array(T)` becomes a `[]T` plus five generic helpers in `internal/array`:
`Push`, `Emplace`, `Pop`, `RemoveSwap` and `Reserve`. They keep `b2Array`'s growth: the first push allocates 8,
then capacity doubles, and a reserve allocates exactly. The container tests
check capacities, and allocation should not differ from the code you diff
against. Code that doesn't depend on growth uses `append`, `slices` and
`clear`.

Upstream's allocator functions and byte counters are not ported, because Go
allocates and collects. The arena allocator is not ported either. Per-step
scratch arrays are slices that the world keeps and reuses.

`b2HashSet`, `b2BitSet` and the ID pool are ported function by function, in
`internal/hashset`, `internal/bitset` and `internal/idpool`.
Atomics become `sync/atomic` typed values, bit scans become `math/bits`,
`b2Mutex` becomes `sync.Mutex` and a thread becomes a goroutine.
`b2ParallelFor` keeps the same block-claiming split, so the work split stays
deterministic. Alignment macros, `B2_RESTRICT` and `B2_INLINE` are dropped.

## Repository layout

The repository has the following layout:

```
github.com/zigai/box2d-go
  box2d/                the library, flat; one Go file per upstream file (body.c, body.h -> body.go)
  internal/scenes/      benchmark, human and determinism scenes; uses only the public API
  internal/<helper>/    array, bitset, core, hashset, idpool, qsort, scheduler, timer
  internal/oracle/      fixture reader; takes primitive values, so package box2d can import it
  internal/lint/        fmaround analyzer and the box2dlint command, in its own module
  oracle/               C harness at the pin; oracle/tests/*.c are fixture programs
  oracle/patches/tests/ dump calls added to upstream tests, applied at build time
  oracle/tests/scenes/  scene drivers shared by the scenes_*.c programs
  testdata/oracle/      gzipped fixtures, in float/ and double/, plus manifest.json
  test/integration/     live C comparisons and differential fuzzing
  bench/                scene benchmarks and the Go and C comparison script
  samples/              interactive samples, in their own module
```

`AGENTS.md` covers the oracle workflow, which checks each change needs, and
how long they take.

## File headers and comments

Each ported Go file starts with upstream's two SPDX lines, copied from the
upstream file, so the year differs per file. A blank line follows, which keeps
the header out of the package doc. `internal/qsort/qsort.go` carries the
Alexey Tourbin notice instead.

```go
// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d
```

Doc comments come from upstream's header comments, adapted to Go names.
Comments inside function bodies are upstream's, kept where the code they
describe is kept. API that only exists in Go, such as `Body.ID` or
`WorldDef.Logger`, gets a doc comment that states its contract. Don't add
comments that explain the porting, the rounding or the harness; the reasons
belong in this page.

## Tests

Each upstream test maps to Go tests with the same names, and each case reads
an oracle fixture captured from C:

- One Go test per upstream top-level case, and one subtest per upstream
  subtest, with the same names. `WorldTest` becomes `TestWorld` with subtests
  such as `HelloWorld`. `oracle/cases.json` maps each upstream test to its Go
  test and fixture.
- Checks keep upstream's expressions, tolerances and iteration counts. An
  `ENSURE( c )` becomes `if !(c) { t.Fatal(...) }`. There are no shared
  `Ensure` helpers.
- Each case has one fixture, named `<file>/<case>`, such as
  `world/HelloWorld`. The C program calls `ORACLE_OPEN( "world/HelloWorld" )`
  and the Go test calls `oracle.Open(t, "world/HelloWorld")`, then both make
  the same sequence of checks. Keys are stable strings. Search for a fixture
  name to find the C and Go pair.
- A test fails if it leaves fixture records unread, unless it failed or was
  skipped first.
- Scenes record the world hash at every step and the full state every 60
  steps.
- A test that only uses the public API, or imports `internal/scenes`, is in
  `package box2d_test`. `internal/scenes` imports `box2d`, so an in-package
  test can't import it. A test that reaches internals and needs no scenes,
  such as `solver_test.go`, is in `package box2d`. Tests of an `internal/`
  helper live in that helper's package.
- A `box2d_test` test that needs internals uses an `export_*_test.go` file in
  `package box2d`, such as `export_world_test.go`, that exports only what that
  test needs.

Fixtures are generated by `just oracle` and never edited by hand.
`manifest.json` hashes every C source under `oracle/`, so editing the harness
or `UPSTREAM` makes the tests fail as stale until you regenerate.

### What counts as a match

- The reference is the C engine at the `UPSTREAM` commit, built with scalar
  math. SIMD and clang builds are cross-checks, not the reference.
- Floats match bit for bit. Enums, booleans, integers and IDs match exactly.
  There are no tolerances, rounding or scrubbing. Float and double builds have
  separate fixtures.
- Collections are compared in order when C defines the order, otherwise as
  multisets that keep duplicates.
- Tests compare return values, event payloads and the state each step
  changes. Large scenes compare a defined state hash and report fields at the
  first step that differs. Hashes of raw memory don't count.
- Snapshots and recordings are not shared between C and Go. Each engine saves
  and restores its own, and tests compare the behavior that results.
- Error tests cover only rejections C defines. Undefined C behavior is not a
  reference.
- Go-only behavior, such as goroutine ownership, races, Go error types and the
  rounding analyzer, has native Go tests. Those don't count as C parity.
- Tests never regenerate or accept fixtures.

## Go 1.27 findings

[The float rule](#the-float-rule) rests on probe programs built with
`go1.27.1` for amd64, arm64, riscv64 and wasm. This table shows the
instructions each target emits, from `-gcflags=-S`:

| Target | `a*b + c` | `float32(a*b) + c` | `t := a*b; t + c` | `c - a*b`, `c -= a*b` | `c += a*b` | `a*b + c*d` |
|---|---|---|---|---|---|---|
| arm64 | `FMADDS` (fused) | `FMULS`, `FADDS` | `FMADDS` (fused) | `FMSUBS` (fused) | `FMADDS` (fused) | `FMULS`, `FMADDS` |
| riscv64 | `FMADDS` (fused) | `FMULS`, `FADDS` | `FMADDS` (fused) | `FNMSUBS` (fused) | `FMADDS` (fused) | `FMULS`, `FMADDS` |
| amd64, `GOAMD64=v3` | `VFMADD231SS` (fused) | `MULSS`, `ADDSS` | `VFMADD231SS` (fused) | `MULSS`, `SUBSS` | `VFMADD231SS` (fused) | `MULSS`, `VFMADD231SS` |
| amd64, default (`v1`) | `MULSS`, `ADDSS` | same | same | same | same | `MULSS`, `MULSS`, `ADDSS` |
| wasm | `F32Mul`, `F32Add` | same | same | `F32Mul`, `F32Sub` | not run | not run |

- Only an explicit `float32(...)` conversion prevents fusion. A local
  variable does not. In a sum of two products, the compiler fuses one of them
  into the add. Wasm never fuses, and amd64 fuses from `GOAMD64=v3`, so CI
  runs `v3` cells on amd64 as well as arm64 cells.
- Typed float32 constants round per operation, like C: with
  `const a float32 = 0.1`, `const b float32 = a * a` gives bits `3c23d70b`.
  The untyped `0.1 * 0.1` converted to float32 gives `3c23d70a`.
- `float32(math.Sqrt(float64(x)))` matched a correctly rounded result on
  2,000,000 random float32 inputs.
- Builtin `min(-0, +0)` is `-0` and `min(NaN, 1)` is NaN, where `b2MinFloat`
  gives `+0` and `1`. `math.Abs(-0)` is `+0`, where `b2AbsFloat(-0)` is `-0`.
- `int32(float32(3.0e9))` is `-2147483648` on amd64. C leaves it undefined.
- The portable `simd` package's vector width depends on the hardware, so it
  can't give the fixed width 4 that contact batching needs. `simd/archsimd`
  is amd64 only and outside the Go 1 compatibility promise.
