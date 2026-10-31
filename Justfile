set shell := ["bash", "-euo", "pipefail", "-c"]

tags := "box2d_double"

default: help

help:
    @just --list

test:
    go test -timeout 60m ./...
    go test -timeout 60m -tags {{ tags }} ./...

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

tidy:
    go mod tidy
    go -C internal/lint mod tidy

mod-check:
    go mod tidy -diff
    go -C internal/lint mod tidy -diff

format:
    golangci-lint run --fix --enable-only whitespace --issues-exit-code=0 --build-tags box2d_stress
    golangci-lint run --fix --enable-only whitespace --issues-exit-code=0 --build-tags box2d_stress,{{ tags }}
    golangci-lint fmt
    git ls-files -z 'oracle/*.c' 'oracle/*.h' | xargs -0 clang-format -i
    git ls-files -z '*.sh' | xargs -0 shfmt -w

fix:
    golangci-lint run --fix --build-tags box2d_stress
    golangci-lint run --fix --build-tags box2d_stress,{{ tags }}
    golangci-lint fmt

race:
    go test -race -timeout 90m ./...
    go test -race -timeout 90m -tags {{ tags }} ./...

check: mod-check race lint
    golangci-lint fmt --diff

stress:
    ./test/run.sh stress -timeout 60m -tags box2d_stress
    ./test/run.sh stress -timeout 60m -tags box2d_stress,{{ tags }}

test-all:
    ./test/run.sh all -timeout 60m -tags box2d_stress
    ./test/run.sh all -timeout 60m -tags box2d_stress,{{ tags }}
    ./test/run.sh all -race -timeout 90m -tags box2d_stress
    ./test/run.sh all -race -timeout 90m -tags box2d_stress,{{ tags }}

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

clang-tidy:
    cmake -S oracle -B oracle/.build/tidy -DCMAKE_EXPORT_COMPILE_COMMANDS=ON -DCMAKE_BUILD_TYPE=RelWithDebInfo -DCMAKE_C_COMPILER=clang -DBOX2D_UPSTREAM_DIR="$PWD/oracle/.upstream/box2d" -DBOX2D_DISABLE_SIMD=ON >/dev/null
    git ls-files -z 'oracle/*.c' ':!oracle/debug/*' | xargs -0 -P "$(nproc)" -n 1 clang-tidy -p oracle/.build/tidy --quiet

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

oracle:
    ./oracle/build.sh

oracle-check:
    go test -timeout 60m -tags box2d_oracle -v ./test/integration
    go test -timeout 60m -tags box2d_oracle,{{ tags }} -v ./test/integration

oracle-bench:
    ORACLE_BENCH_ONLY=1 ./oracle/build.sh

bench samples="6" steps="300" cpus="2":
    [ -x oracle/.build/float/bench/bench_scenes ] && [ -x oracle/.build/double/bench/bench_scenes ] || just oracle-bench
    BENCH_SAMPLES={{ samples }} BENCH_STEPS={{ steps }} GOMAXPROCS={{ cpus }} ./bench/run.sh

bench-go:
    go test -run '^$' -bench BenchmarkScenes ./bench

samples *args:
    go -C samples run . {{ args }}

samples-double *args:
    go -C samples run -tags {{ tags }} . {{ args }}

alias cov := coverage
alias fmt := format
