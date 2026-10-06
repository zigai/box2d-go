# Experimental AABB SIMD backend

## Scope and selection

The ordinary build retains the original scalar AABB overlap expressions. An amd64 build with `GOEXPERIMENT=simd` selects `simd_accel_amd64.go`, using Go 1.27's experimental `simd/archsimd` intrinsics. No assembly files, cgo, or additional module dependencies are required.

The backend vectorizes AABB overlap checks in dynamic-tree traversal and broad-phase pair rejection. It does **not** vectorize the contact solver or imply parity with every upstream SIMD backend.

| Build / startup configuration | AABB overlap implementation |
| --- | --- |
| Normal build, any `BOX2D_SIMD` value | Original scalar Go |
| `GOEXPERIMENT=simd`, amd64 with AVX, `BOX2D_SIMD` unset or `1` | AVX SIMD |
| Same experimental build, `BOX2D_SIMD=0` | Scalar fallback |
| Experimental build on amd64 without AVX | Scalar fallback |
| Experimental build on another architecture | Original scalar Go |

Only the exact override `BOX2D_SIMD=0` disables an available experimental backend. Selection is fixed at package initialization; changing the environment later does not switch existing worlds. `GOEXPERIMENT` is a build-time setting, whereas `BOX2D_SIMD` also works at startup with an already-built experimental executable. AVX2 is not required. Go's AVX detection includes operating-system support for the register state.

Go's SIMD APIs remain experimental and outside the usual compatibility guarantee. Toolchain upgrades may require updating the opt-in implementation; normal builds exclude its import and implementation.

## Profile and opportunity

Scenario: TreeCast, one worker, `GOMAXPROCS=2`, seed 12345, 60 steps for initial profiling; 300 steps for final timing. Normal optimized Go binaries retain profiling information. No CPU governor, kernel, or cache tuning was applied. Timing runs had no other agent-owned heavy jobs running concurrently.

The baseline CPU profile ranked:

| Rank | Location | Flat CPU |
| --- | --- | ---: |
| 1 | `DynamicTree.CastRay` | 29.33% |
| 2 | `overlapNode` | 17.16% |
| 3 | `DynamicTree.CastBox` | 14.22% |
| 4 | `Vec2.Dot` | 9.80% |
| 5 | `orderCastHits` | 4.66% |

Hypothesis: replace four branch-dependent scalar comparisons with a single four-lane ordered comparison. Initial opportunity score: impact 3 × confidence 5 / effort 2 = 7.5. One production lever: the AABB overlap backend. Traversal, hit sorting, callbacks, solver arithmetic, scene construction, and seeds stay unchanged.

Initial prototypes were rejected: a two-comparison implementation was approximately 28% slower, and simplifying it sufficiently to inline was still approximately 20% slower. Copying the query as a four-float array avoided packing individual float registers immediately before a vector load. Two shuffles then formed the exact operands for one comparison. Keeping the final helper within the compiler's inlining budget mattered: the final `overlapNode` cost is 79 against the Go 1.27.1 budget of 80. This is toolchain-sensitive, not an API guarantee.

The final profile still ranks `overlapNode` second (21.57% flat CPU), after `CastRay` (27.26%), followed by `Vec2.Dot` (11.56%) and `CastBox` (10.96%). Whole-workload time decreases, but the flat attribution of inlined overlap code does not; do not interpret the profile as evidence that every individual stage improved.

## Behavior and memory proof

For bounds laid out as `[lowerX, lowerY, upperX, upperY]`, the vector comparison is:

```text
lower = [left.lowerX, left.lowerY, right.lowerX, right.lowerY]
upper = [right.upperX, right.upperY, left.upperX, left.upperY]
overlap = all four lanes of (lower <= upper)
```

These are the four original comparisons, including inclusive boundaries. The ordered comparison returns false for a NaN in either operand; sign and payload do not affect that result. Signed zeros, subnormals, infinities, and invalid boxes retain the original comparison semantics. There is no new floating-point arithmetic, fusion, rounding, or reduction.

- Ordering preserved: traversal and callback code is unchanged; each overlap decision agrees with the scalar expression.
- Tie-breaking unchanged: no hit-sort or priority logic changes.
- Floating-point results preserved: inputs are loaded without conversion or arithmetic, and comparison output is a boolean.
- RNG seeds and workloads unchanged.
- Query load, conditional store, and pair-union helpers preserve all four stored float32 bit patterns. Union still delegates to the original `AABB.Union`.

The unsafe conversions reinterpret an existing, live 16-byte `AABB` as four float32 values. There is no pointer arithmetic, larger read, allocation, or retained pointer. The array and AABB have the same alignment and contain no Go pointers. `TestAABBVectorLayout` verifies size, alignment, and all component offsets in both position precisions. Tests also cover unchanged input bits and conditional-store behavior. The lint configuration has a narrow G103 exclusion limited to `unsafe.Pointer` lines in this backend; the layout and aliasing audit above is the justification, not a general suppression of security checks.

AVX operations occur only after the cached CPU/startup guard. The generated `CastRay` assembly checks `simdEnabled` before `VMOVDQU`, `VSHUFPS`, `VCMPLEPS`, and `VMOVMSKPS`. Before the guard, query copies use ordinary SSE `MOVUPS`/`MOVSS`. The array load/store helpers themselves do not require AVX. Disabling AVX through `GODEBUG=cpu.avx=off` passes the normal suites in both precisions; this tests feature dispatch on an AVX-capable host, not execution on physically older hardware.

## Final measurements

Host: AMD EPYC 9J14, linux/amd64, Go 1.27.1. Hyperfine: 3 warmups, 20 measured runs per configuration. Each process executes one TreeCast/workers=1 workload of **300 steps**, with `GOMAXPROCS=2`.

| Configuration | Mean wall time | p50 | p95 | p99 / max | Steps/s |
| --- | ---: | ---: | ---: | ---: | ---: |
| Normal scalar build | 2.938982 s | 2.933442 s | 2.959016 s | 2.987861 s | 102.08 |
| Experimental SIMD | 2.796800 s | 2.788752 s | 2.849843 s | 2.859449 s | 107.27 |
| Experimental build, SIMD disabled | 2.976459 s | 2.966744 s | 3.011063 s | 3.062585 s | 100.79 |

SIMD reduces mean workload wall time by **4.84%** relative to the normal scalar build and by **6.04%** relative to its same-binary fallback. The fallback is approximately 1.28% slower than the normal build, which has neither the runtime branch nor the experimental query representation. Use a normal build when you do not want that overhead. These are workload measurements, not per-frame p95, and no engine-wide speedup is claimed. Percentiles use nearest ranks; p99 is the maximum of only 20 samples and does not estimate a production tail reliably.

The 60-step profiling runs have 540,060 allocations per workload in both implementations; bytes remain approximately 14.43 MB, with profiling noise. Peak process RSS was 25,716 KiB scalar and 26,828 KiB SIMD, measured with the CPU/memory profilers enabled. These are single diagnostic RSS samples, not evidence of a memory improvement.

A smoke survey ran all 16 scenes with one/four workers, 5 steps, and three samples in normal/SIMD builds. All workloads completed. Its sample count is insufficient for confident per-scene performance conclusions; it is not a universal regression or speedup guarantee.

## Verification

- Normal suites passed in both precisions for normal builds, SIMD builds, runtime-disabled SIMD builds, and simulated unavailable AVX: eight configurations.
- SIMD full race suites passed in both precisions; targeted overlap/configuration race tests also passed with SIMD disabled.
- Overlap tests compare the original scalar expression with both overlap entry points over 7,744 special-value placements and one million deterministic bit-pattern pairs, checking that inputs stay unchanged. The special values include both zeros, subnormals, finite extremes, infinities, signaling/quiet NaNs, signs, and representative payloads.
- Load/store and union/store tests check 10,000 deterministic bit-pattern pairs.
- Full-size Queries, TreeCast, and TileWorld C-oracle scene checks passed in both precisions and with one/four workers.
- Before/after workload captures covering 300 TreeCast steps plus 2,000 deterministic ray/box casts, callback policies, and node/leaf statistics remain byte-identical. Both precision captures have SHA-256 `99933f24ca22064fdcf2124a59b208e2019c50abc48a2c289a3b843045341e35`.
- ARM64 normal/experimental builds and 386 experimental builds cross-compiled the scalar fallback; these binaries were not executed.
- Configured scalar/SIMD lint and SIMD vet checks passed, including stress/double-precision configurations.
- Matched Go/C benchmark runs passed for all 16 scenes, one/four workers, both precisions, one step, and SIMD enabled/disabled. The runner records `GOEXPERIMENT`, `BOX2D_SIMD`, and `GODEBUG` without changing workload-verification headers or the C adapter.

## Reproduction

```sh
go test ./...
GOEXPERIMENT=simd go test ./...
GOEXPERIMENT=simd BOX2D_SIMD=0 go test ./...
GOEXPERIMENT=simd GODEBUG=cpu.avx=off go test ./...
GOEXPERIMENT=simd go test -tags box2d_double ./...
GOEXPERIMENT=simd go test -race ./...
GOEXPERIMENT=simd go test -race -tags box2d_double ./...
GOEXPERIMENT=simd go test -tags box2d_stress -run '^Test(Parallel)?Scenes/(Queries|TreeCast|TileWorld)$' .
GOEXPERIMENT=simd go test -tags box2d_stress,box2d_double -run '^Test(Parallel)?Scenes/(Queries|TreeCast|TileWorld)$' .

go test -c -o /tmp/box2d-scalar.test ./bench
GOEXPERIMENT=simd go test -c -o /tmp/box2d-simd.test ./bench
hyperfine --warmup 3 --runs 20 \
  'BENCH_STEPS=300 GOMAXPROCS=2 /tmp/box2d-scalar.test -test.run "^$" -test.bench "^BenchmarkScenes/TreeCast/workers=1$" -test.benchtime=1x -test.benchmem' \
  'BENCH_STEPS=300 GOMAXPROCS=2 /tmp/box2d-simd.test -test.run "^$" -test.bench "^BenchmarkScenes/TreeCast/workers=1$" -test.benchtime=1x -test.benchmem' \
  'BOX2D_SIMD=0 BENCH_STEPS=300 GOMAXPROCS=2 /tmp/box2d-simd.test -test.run "^$" -test.bench "^BenchmarkScenes/TreeCast/workers=1$" -test.benchtime=1x -test.benchmem'
```

Raw profiles, timings, fingerprint, golden checksums, assembly, and test logs from this run are under `/tmp/box2d-perf/simd-*`. The scratch directory is local evidence, not a required repository dependency.

## Rollback

To disable the experiment without changing code, omit `GOEXPERIMENT=simd`, or start an experimental executable with `BOX2D_SIMD=0`. To remove the implementation, restore the original `simd.go` helpers, remove the backend and SIMD test files, and remove the SIMD-only documentation, benchmark metadata additions, and narrow lint exclusion. Preserve the earlier absFloat/benchmark-horizon changes and unrelated user edits. No commits or branches were created.
