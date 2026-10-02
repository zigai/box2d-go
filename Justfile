set shell := ["bash", "-euo", "pipefail", "-c"]

tags := "box2d_double"

default:
    @just --list

test:
    go test ./...
    go test -tags {{tags}} ./...
    go test -race ./...
    go test -race -tags {{tags}} ./...

lint:
    go vet ./...
    go vet -tags {{tags}} ./...
    go -C internal/lint test ./...
    go -C internal/lint build -o "$PWD/bin/box2dlint" ./cmd/box2dlint
    go vet -vettool="$PWD/bin/box2dlint" ./...
    go vet -vettool="$PWD/bin/box2dlint" -tags {{tags}} ./...

skips:
    #!/usr/bin/env bash
    set -euo pipefail
    for t in "" "{{tags}}"; do
        label="${t:-default}"
        go test -tags "$t" -json ./... | jq -rs --arg label "$label" '
            [.[] | select(.Action == "output" and .Test != null and (.Output | test("\\bunported: ")))
                 | {pkg: .Package, test: .Test}] | unique
            | "unported skips (\($label) build): \(length)"'
    done

oracle:
    ./oracle/build.sh

oracle-bench:
    ORACLE_BENCH_ONLY=1 ./oracle/build.sh
