# Bit-preserving absolute-value optimization

## Scope and environment

One production lever: replace the sign-dependent branch in `absFloat` with magnitude-bit extraction, while explicitly retaining signed zero and every NaN payload.

Measured on 2026-10-05 at base revision `494e07acad3b3f1027225a77b0bed1732d8d3a33`, AMD EPYC 9J14, Linux amd64, Go 1.27.1, optimized compiler settings with normal DWARF information. `GOMAXPROCS=2`; no CPU governor, kernel, or global configuration changes. Existing unrelated sample/integration edits were not touched. Same-host, warm-cache runs; performance measurements were not overlapped with test jobs.

Existing workload: `BenchmarkScenes/TreeCast/workers=1`, float precision, seed 12345, release scene configuration, all 300 steps starting at step zero, timestep 1/60, four substeps. Go benchmark time excludes construction/destruction. Hyperfine wall time includes setup, teardown, and process startup. Results below are quantiles of **complete workload runs**, not per-frame latency.

## Profile and selection

Baseline CPU profile, five complete workloads:

| Rank | Location | Flat CPU | Cumulative CPU |
| --- | --- | ---: | ---: |
| 1 | `DynamicTree.CastRay` | 26.93% | 69.38% |
| 2 | `overlapNode` | 13.58% | 13.58% |
| 3 | `DynamicTree.CastBox` | 12.35% | 29.27% |
| 4 | `Vec2.Sub` | 10.83% | 10.83% |
| 5 | `absFloat` | 9.66% | 9.66% |

Opportunity score for `absFloat`: impact 3 × confidence 5 ÷ effort 2 = **7.5**, above the 2.0 threshold.

Hypothesis ledger:
- Sign-dependent rejection arithmetic: supported by the CPU profile and before/after measurements. Bit extraction avoids the unpredictable sign branch for normal values.
- Allocation pressure: supported by allocation profiles, but not selected for the final patch. An initial lazy callback-input experiment reduced bytes by 28% but increased elapsed time by 3%; it was fully reverted.
- I/O bottleneck: rejected for this in-memory workload; resource captures show 100% process CPU and no filesystem inputs/outputs.
- Contention bottleneck: rejected for this workload; baseline mutex samples total approximately 101 microseconds over 17 seconds. Block samples are test-harness channel waits, not engine waits.

Re-profile: `absFloat` falls to 1.57% flat CPU (2.18% cumulative including bit extraction). Tree traversal, overlap rejection and dot products remain dominant. No additional lever is included.

## Results

Hyperfine 1.19.0, three warmups and 20 measured runs **per binary**:

| Metric | Before | After |
| --- | ---: | ---: |
| Mean workload wall time | 3.417899 s | 2.941718 s |
| p50 | 3.415708 s | 2.938841 s |
| p95 (nearest rank) | 3.432152 s | 2.965734 s |
| p99 / maximum | 3.447048 s | 2.965972 s |
| p99.9 / p99.99 | 3.447048 s | 2.965972 s |
| Coefficient of variation | 0.30% | 0.45% |
| Mean step-equivalent throughput | 87.77 steps/s | 101.98 steps/s |
| Timed allocation bytes / 300 steps | 72,153,600 | 72,153,600 |
| Timed allocations / 300 steps | 2,700,300 | 2,700,300 |
| Peak RSS, separate five-workload profile | 26,232 KiB | 25,816 KiB |
| Process CPU, separate profile | 100% | 100% |

Mean elapsed time decreases **13.93%**; throughput increases **16.19%**. p95 decreases **13.59%**. With only 20 observations, p99 and higher tails collapse to the maximum; especially p99.9/p99.99 are conservative observed-sample values, not reliable tail estimates. RSS captures include profiler overhead and do not establish a memory improvement; allocations are unchanged.

A one-sample survey of all 16 existing scene families, with one/four workers at fixed `GOMAXPROCS=2`, shows no >10% elapsed regression. This is a screening check, not a statistically supported speedup claim for those other scenes. No thread-scaling law or workload-size scaling claim is made. Byte throughput is N/A (no payload I/O). Per-core CPU, PSS and heap high-water marks were not separately instrumented.

## Isomorphism proof

- **Ordering preserved:** yes; traversal, callbacks, iteration order, and collection layout are unchanged.
- **Tie-breaking unchanged:** yes; no comparison used for ordering is changed.
- **Floating-point:** identical output bits for every float32 encoding:
  - Positive finite nonzero values and positive infinity already have a clear sign bit.
  - Negative finite nonzero values and negative infinity have nonzero magnitude at most `0x7f800000`; clearing the sign bit matches unary negation exactly.
  - Positive and negative zero return the original value, preserving negative zero as the previous `value < 0` predicate did.
  - All NaN magnitudes exceed `0x7f800000`; the original bits are returned unchanged, including sign and signaling/quiet payload.
  - `Float32bits` / `Float32frombits` reinterpret storage without arithmetic conversion or rounding.
- **RNG seeds:** unchanged; the production helper has no RNG.
- **Golden outputs:** regenerated after the change, then `sha256sum -c golden_checksums.txt` passed in both precisions. Each capture contains 300 actual TreeCast scene hit-count/visit-stat records and 2,000 deterministic ray/box input pairs with ordered callbacks, all exposed input float bits, clipping/ignore/termination policies, masks and visit counts (4,850 records per precision).
- **Golden SHA-256:** both precision captures match `99933f24ca22064fdcf2124a59b208e2019c50abc48a2c289a3b843045341e35`.

The permanent `TestAbsFloatBitEquivalence` checks all 16,777,214 signed NaN encodings, signed edge cases including subnormals/zeros/infinities, and one million deterministic pseudo-random bit patterns against the original branch implementation. This test complements, rather than substitutes for, the all-encodings proof above.

## Validation

Passed:
- `go test ./...`
- `go test -tags box2d_double ./...`
- `go test -race ./...`
- `go test -race -tags box2d_double ./...`
- Targeted `TestScenes` and `TestParallelScenes` oracle fixtures for Queries, TreeCast, and TileWorld, separately in float/double precision.
- `go vet ./...` and `go vet -tags box2d_double ./...`
- `golangci-lint run --build-tags box2d_stress` and `box2d_stress,box2d_double`
- `golangci-lint fmt --diff math_functions.go math_abs_test.go`
- `git diff --check`

The complete stress suite and external C fixture regeneration were not run; existing expectations were not regenerated or modified.

## Reproduction and local evidence

Local artifacts remain in `/tmp/box2d-perf/`: `fingerprint.json`, `scenario.md`, original/optimized benchmark binaries, `baseline.json`, `abs-comparison.json`, CPU/allocation profiles and ranked text reports, resource captures, compiler escape analysis, test/lint logs, `capture.go`, `golden_outputs/`, and `golden_checksums.txt`. The rejected allocation experiment is retained as `comparison.json` and documented in `abs-decision.md`.

The temporary capture source was run from `bench/.perf_capture/main.go` to permit imports of internal scene helpers, then removed from the repository. To repeat golden verification, place `/tmp/box2d-perf/capture.go` at that temporary path, regenerate the two outputs with `go run` and `go run -tags box2d_double`, and check the saved checksums.

Build the existing benchmark with `go test -c -o /tmp/box2d-perf/abs-after.test ./bench`. Baseline binary `before.test` was captured before modifying the helper. No benchmark implementation changes or PGO were used.

```sh
hyperfine --warmup 3 --runs 20 --export-json /tmp/box2d-perf/abs-comparison.json \
  'GOMAXPROCS=2 /tmp/box2d-perf/before.test -test.run "^$" -test.bench "^BenchmarkScenes/TreeCast/workers=1$" -test.benchtime=1x -test.count=1 -test.benchmem' \
  'GOMAXPROCS=2 /tmp/box2d-perf/abs-after.test -test.run "^$" -test.bench "^BenchmarkScenes/TreeCast/workers=1$" -test.benchtime=1x -test.count=1 -test.benchmem'
```

Rollback before committing: restore the helper from `/tmp/box2d-perf/math_functions.before.go` and remove only the newly added `math_abs_test.go` and this report. No commits or branches were created. After a future isolated commit, revert that commit and rerun both precision suites.

## Raw wall-time samples

Seconds, in collection order; binaries measured sequentially, not pair-interleaved.

| Sample | Before | After |
| --- | ---: | ---: |
| 1 | 3.409152290 | 2.946395983 |
| 2 | 3.406847021 | 2.932378665 |
| 3 | 3.412036093 | 2.937085400 |
| 4 | 3.427364360 | 2.935825217 |
| 5 | 3.412919040 | 2.965734453 |
| 6 | 3.432152048 | 2.964891346 |
| 7 | 3.404239821 | 2.965971534 |
| 8 | 3.416146026 | 2.949424776 |
| 9 | 3.420914605 | 2.948213868 |
| 10 | 3.417749824 | 2.940595809 |
| 11 | 3.428815716 | 2.932922205 |
| 12 | 3.412772516 | 2.928725990 |
| 13 | 3.415985700 | 2.935400118 |
| 14 | 3.411947157 | 2.929466070 |
| 15 | 3.421612161 | 2.944704942 |
| 16 | 3.447048168 | 2.925687482 |
| 17 | 3.410569565 | 2.940649011 |
| 18 | 3.415429742 | 2.954475106 |
| 19 | 3.425165314 | 2.932348159 |
| 20 | 3.409110206 | 2.923457988 |
