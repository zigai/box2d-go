#!/bin/sh

set -eu

# Run every package under the race detector. The box2d package runs as
# concurrent shards balanced by test/race-durations.txt; the last shard runs
# every test the other shards do not.
usage='usage: test/race.sh <shards> <tags> [-test.* flags]'
shards=${1:?$usage}
tags=${2?$usage}
shift 2

weights=$(pwd)/test/race-durations.txt
dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT

go test -c -race -tags "$tags" -o "$dir/box2d.test" ./box2d
(cd box2d && "$dir/box2d.test" -test.list '.*') | grep -E '^(Test|Fuzz|Example)' >"$dir/tests"

awk -v shards="$shards" -v dir="$dir" -v weights="$weights" '
    FILENAME == weights {
        if ($0 !~ /^#/) weight[$2] = $1 + 0
        next
    }
    {
        count++
        name[count] = $0
        estimate[count] = ($0 in weight) ? weight[$0] : 1
    }
    END {
        for (i = 1; i <= count; i++) order[i] = i
        for (i = 2; i <= count; i++) {
            key = order[i]
            for (j = i - 1; j > 0 && estimate[order[j]] < estimate[key]; j--) order[j + 1] = order[j]
            order[j + 1] = key
        }
        for (i = 1; i <= count; i++) {
            target = 0
            for (s = 1; s < shards; s++) if (load[s] < load[target]) target = s
            load[target] += estimate[order[i]]
            shard[order[i]] = target
        }
        for (s = 0; s < shards - 1; s++) {
            pattern = ""
            for (i = 1; i <= count; i++) if (shard[i] == s) pattern = pattern (pattern == "" ? "" : "|") name[i]
            print "run " pattern >(dir "/plan." s)
            skip = skip (skip == "" || pattern == "" ? "" : "|") pattern
        }
        print "skip " skip >(dir "/plan." (shards - 1))
    }
' "$weights" "$dir/tests"

status=0
pids=
shard=0
while [ "$shard" -lt "$shards" ]; do
    read -r mode pattern <"$dir/plan.$shard"
    (cd box2d && exec "$dir/box2d.test" -test."$mode" "^(${pattern})$" "$@") &
    pids="$pids $!"
    shard=$((shard + 1))
done

others=$(go list -tags "$tags" ./... | grep -vx 'github.com/zigai/box2d-go/box2d')
# shellcheck disable=SC2086 # one package per word
go test -race -tags "$tags" "$@" $others || status=1

for pid in $pids; do
    wait "$pid" || status=1
done
exit "$status"
