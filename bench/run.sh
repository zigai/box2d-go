#!/bin/sh
set -eu
GOMAXPROCS=${GOMAXPROCS:-2}
export GOMAXPROCS
samples=${BENCH_SAMPLES:-6}
steps=${BENCH_STEPS:-300}
output=${BENCH_OUT:-bench/results}
build=${BOX2D_ORACLE_DIR:-oracle/.build}
goexperiment=$(go env GOEXPERIMENT)
case "$samples" in ''|*[!0-9]*|0) printf '%s\n' 'BENCH_SAMPLES must be a positive finite integer' >&2; exit 2;; esac
case "$GOMAXPROCS" in ''|*[!0-9]*|0) printf '%s\n' 'GOMAXPROCS must be a positive integer' >&2; exit 2;; esac
case "$steps" in ''|*[!0-9]*) printf '%s\n' 'BENCH_STEPS must be an integer in 1..2147483647' >&2; exit 2;; esac
steps=$(printf '%s' "$steps" | sed 's/^0*//')
[ -n "$steps" ] && [ "${#steps}" -le 10 ] && [ "$steps" -gt 0 ] && [ "$steps" -le 2147483647 ] || { printf '%s\n' 'BENCH_STEPS must be an integer in 1..2147483647' >&2; exit 2; }
BENCH_STEPS=$steps
export BENCH_STEPS
[ "$samples" -gt 0 ] && [ "$GOMAXPROCS" -gt 0 ] || { printf '%s\n' 'Samples and concurrency must be positive integers' >&2; exit 2; }
mkdir -p "$output"
{
 printf 'GOMAXPROCS=%s\nBENCH_SAMPLES=%s\nBENCH_STEPS=%s\nBOX2D_ORACLE_DIR=%s\n' "$GOMAXPROCS" "$samples" "$steps" "$build"
 printf 'GOEXPERIMENT=%s\nBOX2D_SIMD=%s\nGODEBUG=%s\n' "$goexperiment" "${BOX2D_SIMD-<unset>}" "${GODEBUG-<unset>}"
 printf '%s\n' 'upstream-pin=956ce4e1e8acddd05a21102629e268ff9321ec50'
 printf 'workload=custom%s; steps=%s; first-step=0; timestep=1/60; substeps=4; seed=12345; workers=1,4; continuous=true; rebuild-static=false; creation/destruction excluded\n' "$steps" "$steps"
 printf 'Go checkout revision: '; git rev-parse HEAD
 printf 'C checkout revision: '; git -C oracle/.upstream/box2d rev-parse HEAD
 uname -a
 go version
 getconf _NPROCESSORS_ONLN
 if [ -r /proc/self/status ]; then
  awk '/^Cpus_allowed_list:/ {print}' /proc/self/status
 fi
 "${ORACLE_CC:-gcc}" --version
} > "$output/environment.txt"
go env -json > "$output/go-environment.json"
for precision in float double; do
 cp "$build/$precision/CMakeCache.txt" "$output/cmake-$precision.txt"
 sha256sum "$build/$precision/bench/bench_scenes" > "$output/c-binary-$precision.sha256"
 tags=
 [ "$precision" != double ] || tags=box2d_double
 printf 'GOEXPERIMENT=%s BOX2D_SIMD=%s BENCH_STEPS=%s go test -timeout 60m -tags %s -cpu %s -run ^$ -bench ^BenchmarkScenes$ -benchtime=1x -count=%s -benchmem ./bench\n' "$goexperiment" "${BOX2D_SIMD-}" "$steps" "$tags" "$GOMAXPROCS" "$samples" > "$output/command-$precision.txt"
 go test -timeout 60m -tags "$tags" -cpu "$GOMAXPROCS" -run '^$' -bench '^BenchmarkScenes$' -benchtime=1x -count="$samples" -benchmem ./bench > "$output/go-$precision.txt" 2> "$output/go-$precision.stderr.txt"
 # Copy only real platform headers, never fabricate C configuration metadata.
 awk '/^goos:|^goarch:|^pkg:|^cpu:/ {print}' "$output/go-$precision.txt" > "$output/c-$precision.txt"
 : > "$output/c-$precision.stderr.txt"
 for workers in 1 4; do
  printf '%s %s %s %s %s\n' "$build/$precision/bench/bench_scenes" "$samples" "$workers" "$GOMAXPROCS" "$steps" >> "$output/command-$precision.txt"
  "$build/$precision/bench/bench_scenes" "$samples" "$workers" "$GOMAXPROCS" "$steps" >> "$output/c-$precision.txt" 2>> "$output/c-$precision.stderr.txt"
 done
 go run ./bench/verify -samples "$samples" -steps "$steps" -cpus "$GOMAXPROCS" -precision "$precision" "$output/go-$precision.txt" "$output/c-$precision.txt" > "$output/verified-$precision.txt" 2> "$output/verify-$precision.stderr.txt"
 benchstat "$output/c-$precision.txt" "$output/go-$precision.txt" > "$output/benchstat-$precision.txt" 2> "$output/benchstat-$precision.stderr.txt"
done
