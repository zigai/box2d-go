# Bench

Benchmarks that time the Go port against the Box2D C engine on the same scenes and settings.

- **Matched workloads:** every benchmark scene runs in Go and in C with 1 and 4 workers.
- **Both precisions:** the float and double builds are measured separately.
- **Checked results:** a verifier rejects any run whose measurements are incomplete or were taken with different settings.
- **Recorded environment:** each run saves the machine, toolchain, and build details next to the numbers.

## Usage

Run the full comparison from the repository root:

```sh
just bench
```

The first run builds the C benchmark, which needs the same tools as `just oracle`. It also needs `benchstat` (`go install golang.org/x/perf/cmd/benchstat@latest`). The command prints a C versus Go table for each precision and saves everything to `bench/results/`.

Pass samples, steps, and CPUs to change the run, in that order:

```sh
just bench 3 100
```

Pin the run to fixed CPUs for steadier numbers, and rebuild the C side with `just oracle-bench` after changing it:

```sh
taskset -c 2,3 just bench
```

Measure only the Go side:

```sh
just bench-go
```

| Argument | Default | Effect |
| --- | --- | --- |
| `samples` | `6` | Samples per scene and worker count |
| `steps` | `300` | Steps per scene |
| `cpus` | `2` | Concurrency for both engines |
