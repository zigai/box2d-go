# Oracle

The C reference for the Go port. It builds the Box2D C engine at the commit in [`UPSTREAM`](../UPSTREAM), runs small C programs against it, and records what the engine does. The Go tests replay those recordings offline and require the Go port to match them bit for bit.

- **Fixtures:** each program writes typed records, with floats stored as exact bits, to `testdata/oracle/float/` and `testdata/oracle/double/`.
- **Paired tests:** a C program and its Go test share a fixture name, so a search for the name finds both sides.
- **Upstream tests:** the upstream test suite runs with recording calls added through patches, so its cases check the Go port too.
- **Cross-checks:** SIMD and clang builds are compared with the scalar `gcc` reference, and any fixture that differs is reported.
- **Freshness:** `manifest.json` hashes every source here. Changing one makes the Go tests fail as stale until the fixtures are regenerated.

## Usage

Regenerate every fixture and the manifest after changing anything in this folder or `UPSTREAM`:

```sh
just oracle
```

This takes about 25 minutes. It needs `cmake`, `gcc`, `clang`, `git`, `gzip`, and network access to fetch the pinned upstream into `oracle/.upstream/`. Commit the regenerated fixtures together with the change that caused them.

Run the live comparisons against the built C programs, including generated operation histories:

```sh
just oracle-check
```

Build only the C side of the scene benchmarks:

```sh
just oracle-bench
```

## Adding a case

1. Write a program in `tests/` that calls `ORACLE_OPEN("area/Case")` and records values with the `ORACLE_*` macros from `oracle.h`.
2. Write the Go test that calls `oracle.Open(t, "area/Case")` and reads the same records in the same order.
3. Run `just oracle` and commit the new fixture with the code.

A Go test fails when it leaves records unread, so the C and Go sides can't drift apart silently.

## Layout

| Path | Contents |
| --- | --- |
| `oracle.c`, `oracle.h` | Recording library: fixture files, typed records, and hashing helpers |
| `tests/` | One fixture program per area, and shared recording helpers |
| `tests/scenes/` | Drivers for the large scene programs, which reuse the upstream `shared/` scenes |
| `patches/tests/` | Recording calls added to the upstream tests, applied at configure time |
| `patches/world_overlap_stop.patch` | Upstream fix: a world overlap query stops when its callback returns false |
| `sample/` | Small probes of numeric types and rounding |
| `debug/` | Distance diagnostics that need the Debug engine build |
| `differential.c` | Runner for generated operation histories, used by `just oracle-check` |
| `bench/` | C side of the scene benchmarks |
| `cases.json` | Map from each upstream test to its Go test and fixture |

## Options

`build.sh` reads these environment variables:

| Variable | Default | Effect |
| --- | --- | --- |
| `ORACLE_CC` | `gcc` | Compiler for the reference build |
| `ORACLE_CLANG` | `clang` | Compiler for the clang cross-check, skipped when missing |
| `ORACLE_CHECKS` | `1` | Set to `0` to skip the SIMD and clang cross-checks |
| `ORACLE_STRICT` | `0` | Set to `1` to fail when a cross-check differs, instead of warning |
| `ORACLE_GENERATOR` | Ninja when installed | Build system generator for `cmake` |

## Code style

The C code has its own `.clang-format` and `.clang-tidy`, and builds with warnings as errors. `just format` formats it, and `just lint` checks formatting and runs clang-tidy.
