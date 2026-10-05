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
    golangci-lint fmt

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

lint:
    golangci-lint run --build-tags box2d_stress
    golangci-lint run --build-tags box2d_stress,{{ tags }}
    go vet -tags box2d_stress ./...
    go vet -tags box2d_stress,{{ tags }} ./...
    go -C internal/lint test ./...
    go -C internal/lint build -o "$PWD/bin/box2dlint" ./cmd/box2dlint
    go vet -vettool="$PWD/bin/box2dlint" -tags box2d_stress ./...
    go vet -vettool="$PWD/bin/box2dlint" -tags box2d_stress,{{ tags }} ./...

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

alias cov := coverage
alias fmt := format
