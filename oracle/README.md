# Oracle

C programs that record what the Box2D C engine at the commit in [`UPSTREAM`](../UPSTREAM) does. The Go tests replay these recordings offline and must match them bit for bit.

- Recordings go to `testdata/oracle/float/` and `testdata/oracle/double/`, with floats stored as exact bits.
- A C program and its Go test share a fixture name. Search for the name to find both.
- The upstream test suite also runs here, with recording calls added through patches.
- `manifest.json` hashes every source here. Changing one fails the Go tests until the fixtures are regenerated.

## Usage

```sh
just oracle         # build the C engine and regenerate every fixture
just oracle-check   # run the live C comparisons
just oracle-bench   # build only the C side of the benchmarks
```

`just oracle` needs `cmake`, `gcc`, `clang`, `git`, `gzip`, and network access. Commit the regenerated fixtures together with the change that caused them.

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

The C code has its own `.clang-format` and `.clang-tidy` and builds with warnings as errors. `just format` formats it, and `just lint` checks it.
