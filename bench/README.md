# Bench

Benchmarks that time the Go port against the Box2D C engine on the same scenes and settings. Each scene runs with 1 and 4 workers, in both float and double builds.

## Usage

```sh
just bench
```

This prints a C versus Go table for each precision and saves the results to `bench/results/`. It needs `benchstat` and the same tools as `just oracle`.

```sh
just bench 3 100 4          # samples, steps, cpus
taskset -c 2,3 just bench   # pin to fixed CPUs for steadier numbers
just bench-go               # Go side only
just oracle-bench           # rebuild the C side after changing it
```

| Argument | Default | Effect |
| --- | --- | --- |
| `samples` | `6` | Samples per scene and worker count |
| `steps` | `300` | Steps per scene |
| `cpus` | `2` | Concurrency for both engines |
