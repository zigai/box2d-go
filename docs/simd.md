# SIMD backend

An optional amd64 backend speeds up AABB overlap checks in the dynamic tree and
broad phase. It uses Go's experimental `simd/archsimd` package and is built only
with `GOEXPERIMENT=simd`. It does not touch the contact solver.

| Build | Overlap checks |
| --- | --- |
| Normal build | Scalar Go |
| `GOEXPERIMENT=simd`, amd64 with AVX | AVX |
| `GOEXPERIMENT=simd` with `BOX2D_SIMD=0` | Scalar fallback |
| `GOEXPERIMENT=simd`, amd64 without AVX | Scalar fallback |
| `GOEXPERIMENT=simd`, other architectures | Scalar Go |

The backend is chosen once, at package initialization. Results are bit-identical
to the scalar build. On the TreeCast scene it cut wall time by about 5%. The
fallback in a SIMD build is about 1% slower than a normal build.

`simd/archsimd` is outside the Go 1 compatibility promise, so a Go upgrade may
require changes to `simd_accel_amd64.go`.

To test it:

```sh
GOEXPERIMENT=simd go test ./...
GOEXPERIMENT=simd go test -tags box2d_double ./...
GOEXPERIMENT=simd BOX2D_SIMD=0 go test ./...
GOEXPERIMENT=simd GODEBUG=cpu.avx=off go test ./...
```
