#!/bin/sh

set -eu

# Run from the repository root with box2d_stress enabled in the supplied tags.
# Keep heavy one-worker and four-worker scenes in separate processes.
case "${1:-}" in
stress)
	shift
	go test "$@" -run '^Test(ReStepRace|SceneFixtureInventory|SceneHelperFixtureInventory|SceneHelperBoundaries|SceneHelperImmediateOutcomesAndNormalRecycle|ScenePerQueryInputsAndPayloads)$' ./...
	;;
all)
	shift
	go test "$@" -skip '^Test(Parallel)?Scenes$' ./...
	;;
*)
	printf '%s\n' 'usage: test/run.sh {stress|all} [go test flags including -tags box2d_stress]' >&2
	exit 2
	;;
esac

go test "$@" -run '^TestScenes$' ./...
go test "$@" -run '^TestParallelScenes$' ./...
