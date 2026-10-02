#!/bin/sh
set -eu
GOMAXPROCS=${GOMAXPROCS:-2}
export GOMAXPROCS
samples=${BENCH_SAMPLES:-6}
output=${BENCH_OUT:-bench/results}
mkdir -p "$output"
{
 printf 'GOMAXPROCS=%s\n' "$GOMAXPROCS"
 uname -a
 go version
 getconf _NPROCESSORS_ONLN
 if [ -r /proc/self/status ]; then
  awk '/^Cpus_allowed_list:/ {print}' /proc/self/status
 fi
 for precision in float double; do
  awk '/^CMAKE_C_COMPILER:|^CMAKE_C_FLAGS[^:]*:|^CMAKE_BUILD_TYPE:|^BOX2D_DISABLE_SIMD:|^BOX2D_DOUBLE_PRECISION:/ {print}' "oracle/.build/$precision/CMakeCache.txt"
 done
 "${ORACLE_CC:-gcc}" --version
} > "$output/environment.txt"
for precision in float double; do
 tags=
 [ "$precision" != double ] || tags=box2d_double
 go test -tags "$tags" -cpu "$GOMAXPROCS" -run '^$' -bench '^BenchmarkScenes$' -benchtime=1x -count="$samples" -benchmem ./bench > "$output/go-$precision.txt"
 awk '/^Benchmark/ {exit} {print}' "$output/go-$precision.txt" > "$output/c-$precision.txt"
 for workers in 1 4; do
  "oracle/.build/$precision/bench/bench_scenes" "$samples" "$workers" "$GOMAXPROCS" >> "$output/c-$precision.txt"
 done
 benchstat "$output/c-$precision.txt" "$output/go-$precision.txt" > "$output/benchstat-$precision.txt"
done
