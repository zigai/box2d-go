set shell := ["bash", "-euo", "pipefail", "-c"]

TAGS := "box2d_double"

# Run fast unit tests in float and double
[positional-arguments]
test *args:
    #!/usr/bin/env bash
    set -euo pipefail
    if (( $# )); then
        go test "$@"
        go test -tags {{ TAGS }} "$@"
    else
        go test ./...
        go test -tags {{ TAGS }} ./...
    fi

# Measure all local test suites and print coverage
coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    rm -rf coverage
    mkdir -p coverage
    coverage_package=$(awk '$1 == "module" {gsub(/"/, "", $2); print $2; exit}' go.mod)
    go test -coverpkg=./... ./... -args -test.gocoverdir="$PWD/coverage"
    go test -coverpkg=./... -tags "{{ TAGS }}" ./... -args -test.gocoverdir="$PWD/coverage"
    go tool covdata percent -i coverage -pkg "$coverage_package/..."
    go tool covdata textfmt -i coverage -o coverage/profile.txt -pkg "$coverage_package/..."
    go tool cover -func coverage/profile.txt | awk '$1 == "total:"'
    module_coverage() (
        cd "$1"
        rm -rf coverage
        mkdir -p coverage
        coverage_package=$(awk '$1 == "module" {gsub(/"/, "", $2); print $2; exit}' go.mod)
        go test -coverpkg=./... ./... -args -test.gocoverdir="$PWD/coverage"
        go tool covdata percent -i coverage -pkg "$coverage_package/..."
        go tool covdata textfmt -i coverage -o coverage/profile.txt -pkg "$coverage_package/..."
        go tool cover -func coverage/profile.txt | awk '$1 == "total:"'
    )
    module_coverage internal/lint
    (
        cd samples
        rm -rf coverage
        mkdir -p coverage
        coverage_package=$(awk '$1 == "module" {gsub(/"/, "", $2); print $2; exit}' go.mod)
        xvfb-run -a go test -coverpkg=./... ./... -args -test.gocoverdir="$PWD/coverage"
        go tool covdata percent -i coverage -pkg "$coverage_package/..."
        go tool covdata textfmt -i coverage -o coverage/profile.txt -pkg "$coverage_package/..."
        go tool cover -func coverage/profile.txt | awk '$1 == "total:"'
    )

# Format every supported first-party file type
format:
    @golangci-lint run --fix --enable-only whitespace --issues-exit-code=0 --build-tags box2d_stress
    @golangci-lint run --fix --enable-only whitespace --issues-exit-code=0 --build-tags box2d_stress,{{ TAGS }}
    @cd internal/lint && golangci-lint run --fix --enable-only whitespace --issues-exit-code=0
    @cd samples && golangci-lint run --fix --enable-only whitespace --issues-exit-code=0
    @golangci-lint fmt
    @cd internal/lint && golangci-lint fmt
    @cd samples && golangci-lint fmt
    @oxfmt '!**/*.md' '!**/testdata/**' '!**/.clang-*'
    @git ls-files -co --exclude-standard -z | while IFS= read -r -d '' file; do if [[ -f $file ]]; then printf '%s\0' "$file"; fi; done | xargs -0 -r shfmt -f | xargs -r -d '\n' shfmt -w
    @git ls-files -co --exclude-standard -- Justfile justfile "*.just" | while IFS= read -r file; do just --justfile "$file" --fmt; done
    @just markdown fmt

# Apply safe lint autofixes, then format every supported file type
fix: && format
    @just clang-tidy-fix
    golangci-lint run --fix --build-tags box2d_stress
    golangci-lint run --fix --build-tags box2d_stress,{{ TAGS }}
    @cd internal/lint && golangci-lint run --fix
    @cd samples && golangci-lint run --fix
    go vet -fix -tags box2d_stress ./...
    go vet -fix -tags box2d_stress,{{ TAGS }} ./...
    go -C internal/lint build -o "$PWD/bin/box2dlint" ./cmd/box2dlint
    go vet -fix -vettool="$PWD/bin/box2dlint" -tags box2d_stress ./...
    go vet -fix -vettool="$PWD/bin/box2dlint" -tags box2d_stress,{{ TAGS }} ./...

# Run ordinary and stress-enabled race tests in both precisions
race:
    go test -race -timeout 90m ./...
    go test -race -timeout 90m -tags {{ TAGS }} ./...
    ./test/run.sh all -race -timeout 90m -tags box2d_stress
    ./test/run.sh all -race -timeout 90m -tags box2d_stress,{{ TAGS }}

# Run the commit checks
check: (run-checks "formatcheck" "lockcheck" "lint" "test")

# Run full-size scene and replay integration tests in both precisions
integration:
    ./test/run.sh all -timeout 60m -tags box2d_stress
    ./test/run.sh all -timeout 60m -tags box2d_stress,{{ TAGS }}

# Run every configured linter on the whole project, stopping on failure
lint: clang-tidy
    @golangci-lint run --build-tags box2d_stress
    @golangci-lint run --build-tags box2d_stress,{{ TAGS }}
    @cd internal/lint && golangci-lint run
    @cd samples && golangci-lint run
    @go vet -tags box2d_stress ./...
    @go vet -tags box2d_stress,{{ TAGS }} ./...
    @go -C internal/lint test ./...
    @go -C internal/lint build -o "$PWD/bin/box2dlint" ./cmd/box2dlint
    @go vet -vettool="$PWD/bin/box2dlint" -tags box2d_stress ./...
    @go vet -vettool="$PWD/bin/box2dlint" -tags box2d_stress,{{ TAGS }} ./...

# Run clang-tidy on the C reference programs
clang-tidy:
    @cmake -S oracle -B oracle/.build/tidy -DCMAKE_EXPORT_COMPILE_COMMANDS=ON -DCMAKE_BUILD_TYPE=RelWithDebInfo -DCMAKE_C_COMPILER=clang -DBOX2D_UPSTREAM_DIR="$PWD/oracle/.upstream/box2d" -DBOX2D_DISABLE_SIMD=ON >/dev/null
    @git ls-files -z 'oracle/*.c' ':!oracle/debug/*' | xargs -0 -P "$(nproc)" -n 1 clang-tidy -p oracle/.build/tidy --quiet

# Count tests skipped as unported
skips:
    #!/usr/bin/env bash
    set -euo pipefail
    for t in "" "{{ TAGS }}"; do
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
    go test -timeout 60m -tags box2d_oracle,{{ TAGS }} -v ./test/integration

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
    go -C samples run {{ if precision == "double" { "-tags " + TAGS } else if precision == "float" { "" } else { error("precision must be float or double") } }} . {{ args }}

# Remove generated output and caches
clean: (remove "dist" "coverage.out" "*.test" "*.prof" "bin" "oracle/.build" "oracle/.upstream" "bench/results" "coverage" "internal/lint/coverage" "samples/coverage")

# Check the same files and tools without writing
formatcheck:
    @golangci-lint run --enable-only whitespace --build-tags box2d_stress
    @golangci-lint run --enable-only whitespace --build-tags box2d_stress,{{ TAGS }}
    @cd internal/lint && golangci-lint run --enable-only whitespace
    @cd samples && golangci-lint run --enable-only whitespace
    @golangci-lint fmt --diff
    @cd internal/lint && golangci-lint fmt --diff
    @cd samples && golangci-lint fmt --diff
    @oxfmt --check '!**/*.md' '!**/testdata/**' '!**/.clang-*'
    @git ls-files -co --exclude-standard -z | while IFS= read -r -d '' file; do if [[ -f $file ]]; then printf '%s\0' "$file"; fi; done | xargs -0 -r shfmt -f | xargs -r -d '\n' shfmt -d
    @git ls-files -co --exclude-standard -- Justfile justfile "*.just" | while IFS= read -r file; do just --justfile "$file" --fmt --check; done
    @just markdown check

# Run the commit checks and slow suites
checkfull: (run-checks "formatcheck" "lockcheck" "lint" "test" "race" "integration" "bench")

# Synchronize dependency locks without upgrading dependencies
[no-exit-message]
lock:
    @go mod tidy
    @cd "internal/lint" && go mod tidy
    @cd "samples" && go mod tidy

# Verify dependency manifests and locks without changing files
[no-exit-message]
lockcheck:
    @go mod tidy -diff
    @cd "internal/lint" && go mod tidy -diff
    @cd "samples" && go mod tidy -diff

# Install repository dependencies
setup:
    @go mod download
    @cd "internal/lint" && go mod download
    @cd "samples" && go mod download

# List available recipes
[default]
help:
    @just --list

alias cov := coverage
alias fmt := format

[positional-arguments]
[private]
remove +paths:
    #!/usr/bin/env bash
    set -euo pipefail
    shopt -s nullglob globstar
    size() { LC_ALL=C numfmt --from-unit=1024 --to=iec-i --suffix=B --format=%.1f "$1"; }
    rows=() total=0
    for pattern in "$@"; do
        found=()
        for path in $pattern; do
            if [[ -e $path || -L $path ]]; then found+=("$path"); fi
        done
        if (( ${#found[@]} == 0 )); then continue; fi
        kib=$(du -sck -- "${found[@]}" | tail -n1 | cut -f1)
        rm -rf -- "${found[@]}"
        label=$pattern
        if (( ${#found[@]} > 1 )); then label+=" (${#found[@]})"; fi
        rows+=("removed  $label|$(size "$kib")")
        total=$(( total + kib ))
    done
    if [[ -n ${CLEAN_TOTAL_FILE:-} ]]; then echo "$total" >> "$CLEAN_TOTAL_FILE"; fi
    if (( ${#rows[@]} == 0 )); then echo "nothing to clean"; exit 0; fi
    printf '%s\n' "${rows[@]}" "total freed|$(size "$total")" | column -t -s '|' -R 2

[positional-arguments]
[private]
run-checks +steps:
    #!/usr/bin/env bash
    set -uo pipefail
    LC_NUMERIC=C
    log_dir="${XDG_CACHE_HOME:-$HOME/.cache}/just-check"
    mkdir -p "$log_dir"
    log=$(mktemp "$log_dir/step.XXXXXX")
    trap 'rm -f "$log"' EXIT
    width=0 failed=()
    for step in "$@"; do (( ${#step} > width )) && width=${#step}; done
    for step in "$@"; do
        read -ra command <<< "$step"
        printf 'check  %-*s  ' "$width" "$step"
        start=$EPOCHREALTIME
        run=(just --justfile {{ quote(justfile()) }} "${command[@]}")
        if [[ -t 1 ]]; then run=(script -qefc "${run[*]@Q}" /dev/null); fi
        if "${run[@]}" < /dev/null > "$log" 2>&1; then status=ok; else status=FAIL; failed+=("$step"); fi
        printf '%-4s  %7.1fs\n' "$status" "$(LC_NUMERIC=C awk -v a="$start" -v b="$EPOCHREALTIME" 'BEGIN { print b - a }')"
        if [[ "$status" == FAIL ]]; then sed 's/\r$//; s/^/    /' "$log"; fi
    done
    if (( ${#failed[@]} )); then printf '%d of %d failed: %s\n' "${#failed[@]}" "$#" "${failed[*]}"; exit 1; fi
    printf 'all %d passed\n' "$#"

[private]
clang-tidy-fix:
    @cmake -S oracle -B oracle/.build/tidy -DCMAKE_EXPORT_COMPILE_COMMANDS=ON -DCMAKE_BUILD_TYPE=RelWithDebInfo -DCMAKE_C_COMPILER=clang -DBOX2D_UPSTREAM_DIR="$PWD/oracle/.upstream/box2d" -DBOX2D_DISABLE_SIMD=ON >/dev/null
    @git ls-files -z 'oracle/*.c' ':!oracle/debug/*' | xargs -0 -P 2 -n 1 clang-tidy --fix -p oracle/.build/tidy --quiet

[private]
markdown *args:
    rumdl {{ args }} --disable MD001,MD013,MD024,MD025,MD026,MD031,MD032,MD033,MD034,MD036,MD041,MD058 --config 'MD060.enabled=true' --config 'MD060.style="aligned"' --config 'MD003.style="atx"' --config 'MD004.style="dash"' --config 'MD029.style="ordered"' --config 'MD046.style="fenced"' --config 'MD048.style="backtick"' --config 'MD049.style="asterisk"' --config 'MD050.style="asterisk"' --config 'MD055.style="leading_and_trailing"'
