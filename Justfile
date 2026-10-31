set shell := ["bash", "-euo", "pipefail", "-c"]

tags := "box2d_double"

# Show the recipe list
default: help

# List recipes with descriptions
help:
    @just --list

# Run the default tests in float and double
test:
    go test -timeout 60m ./...
    go test -timeout 60m -tags {{ tags }} ./...

# Print test coverage in float and double
coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    coverage_file=$(mktemp)
    trap 'rm -f "$coverage_file"' EXIT
    for tag in "" "{{ tags }}"; do
        if [ -n "$tag" ]; then
            go test -timeout 60m -tags "$tag" -coverprofile="$coverage_file" ./...
        else
            go test -timeout 60m -coverprofile="$coverage_file" ./...
        fi
        go tool cover -func="$coverage_file"
    done

# Tidy go.mod in the root and lint modules
tidy:
    go mod tidy
    go -C internal/lint mod tidy

# Fail if go.mod or go.sum need tidying
mod-check:
    go mod tidy -diff
    go -C internal/lint mod tidy -diff

# Format Go, C and shell sources
format:
    golangci-lint run --fix --enable-only whitespace --issues-exit-code=0 --build-tags box2d_stress
    golangci-lint run --fix --enable-only whitespace --issues-exit-code=0 --build-tags box2d_stress,{{ tags }}
    golangci-lint fmt
    git ls-files -z 'oracle/*.c' 'oracle/*.h' | xargs -0 clang-format -i
    git ls-files -z '*.sh' | xargs -0 shfmt -w

# Apply golangci-lint fixes in both precisions
fix:
    golangci-lint run --fix --build-tags box2d_stress
    golangci-lint run --fix --build-tags box2d_stress,{{ tags }}
    golangci-lint fmt

# Run the tests with the race detector in both precisions
race:
    go test -race -timeout 90m ./...
    go test -race -timeout 90m -tags {{ tags }} ./...

# Run the merge gate: module check, race tests and lint
check: mod-check race lint
    golangci-lint fmt --diff

# Run full-size scene and replay comparisons in both precisions
stress:
    ./test/run.sh stress -timeout 60m -tags box2d_stress
    ./test/run.sh stress -timeout 60m -tags box2d_stress,{{ tags }}

# Run normal and race tests with stress cases in both precisions (2+ hours)
test-all:
    ./test/run.sh all -timeout 60m -tags box2d_stress
    ./test/run.sh all -timeout 60m -tags box2d_stress,{{ tags }}
    ./test/run.sh all -race -timeout 90m -tags box2d_stress
    ./test/run.sh all -race -timeout 90m -tags box2d_stress,{{ tags }}

# Lint Go, C and shell sources in both precisions
lint: clang-tidy
    golangci-lint run --build-tags box2d_stress
    golangci-lint run --build-tags box2d_stress,{{ tags }}
    go vet -tags box2d_stress ./...
    go vet -tags box2d_stress,{{ tags }} ./...
    go -C internal/lint test ./...
    go -C internal/lint build -o "$PWD/bin/box2dlint" ./cmd/box2dlint
    go vet -vettool="$PWD/bin/box2dlint" -tags box2d_stress ./...
    go vet -vettool="$PWD/bin/box2dlint" -tags box2d_stress,{{ tags }} ./...
    git ls-files -z 'oracle/*.c' 'oracle/*.h' | xargs -0 clang-format --dry-run --Werror
    git ls-files -z '*.sh' | xargs -0 shfmt -d

# Run clang-tidy on the C reference programs
clang-tidy:
    cmake -S oracle -B oracle/.build/tidy -DCMAKE_EXPORT_COMPILE_COMMANDS=ON -DCMAKE_BUILD_TYPE=RelWithDebInfo -DCMAKE_C_COMPILER=clang -DBOX2D_UPSTREAM_DIR="$PWD/oracle/.upstream/box2d" -DBOX2D_DISABLE_SIMD=ON >/dev/null
    git ls-files -z 'oracle/*.c' ':!oracle/debug/*' | xargs -0 -P "$(nproc)" -n 1 clang-tidy -p oracle/.build/tidy --quiet

# Count tests skipped as unported
skips:
    #!/usr/bin/env bash
    set -euo pipefail
    for t in "" "{{ tags }}"; do
        label="${t:-default}"
        go test -timeout 60m -tags "box2d_stress $t" -json ./... | jq -rs --arg label "$label" '
            [.[] | select(.Action == "output" and .Test != null and (.Output | test("\\bunported: ")))
                 | {pkg: .Package, test: .Test}] | unique
            | "unported skips (\($label) build): \(length)"'
    done

# Regenerate the C reference fixtures and manifest (about 25 min)
oracle:
    ./oracle/build.sh

# Compare the Go port against the live C engine
oracle-check:
    go test -timeout 60m -tags box2d_oracle -v ./test/integration
    go test -timeout 60m -tags box2d_oracle,{{ tags }} -v ./test/integration

# Build only the C benchmark binaries
oracle-bench:
    ORACLE_BENCH_ONLY=1 ./oracle/build.sh

# Compare Go and C speed on the benchmark scenes
bench samples="6" steps="300" cpus="2":
    [ -x oracle/.build/float/bench/bench_scenes ] && [ -x oracle/.build/double/bench/bench_scenes ] || just oracle-bench
    BENCH_SAMPLES={{ samples }} BENCH_STEPS={{ steps }} GOMAXPROCS={{ cpus }} ./bench/run.sh

# Run the Go scene benchmarks only
bench-go:
    go test -run '^$' -bench BenchmarkScenes ./bench

# Open the sample viewer in float or double precision
samples precision="float" *args:
    go -C samples run {{ if precision == "double" { "-tags " + tags } else if precision == "float" { "" } else { error("precision must be float or double") } }} . {{ args }}

alias cov := coverage
alias fmt := format
