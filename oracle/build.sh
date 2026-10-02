#!/bin/sh

set -eu
(set -o pipefail) 2>/dev/null && set -o pipefail

LC_ALL=C
export LC_ALL

ORACLE_CC=${ORACLE_CC:-gcc}
ORACLE_CHECKS=${ORACLE_CHECKS:-1}
ORACLE_CLANG=${ORACLE_CLANG:-clang}
ORACLE_STRICT=${ORACLE_STRICT:-0}
if [ -z "${ORACLE_GENERATOR:-}" ] && command -v ninja >/dev/null 2>&1; then
	ORACLE_GENERATOR=Ninja
fi
ORACLE_GENERATOR=${ORACLE_GENERATOR:-}
UPSTREAM_URL=https://github.com/erincatto/box2d
BUDGET_BYTES=1048576

here=$(cd "$(dirname "$0")" && pwd)
root=$(dirname "$here")
upstream="$here/.upstream/box2d"
build="$here/.build"
fixtures="$root/testdata/oracle"

die() {
	echo "oracle: $*" >&2
	exit 1
}

warn() {
	echo "oracle: WARNING: $*" >&2
}

pin=$(sed -e 's/#.*//' -e 's/[[:space:]]//g' "$root/UPSTREAM" | grep -v '^$' | head -n 1) || true
case $pin in
*[!0-9a-f]* | '') die "UPSTREAM must hold a 40 digit commit hash, got '$pin'" ;;
esac
[ ${#pin} -eq 40 ] || die "UPSTREAM must hold a 40 digit commit hash, got '$pin'"

command -v cmake >/dev/null 2>&1 || die "cmake is required"
command -v git >/dev/null 2>&1 || die "git is required"
command -v "$ORACLE_CC" >/dev/null 2>&1 || die "$ORACLE_CC is required"

fetch_upstream() {
	if [ -d "$upstream/.git" ] &&
		[ "$(git -C "$upstream" rev-parse HEAD 2>/dev/null || true)" = "$pin" ] &&
		[ -z "$(git -C "$upstream" status --porcelain --ignored 2>/dev/null)" ]; then
		echo "oracle: upstream at $pin already present"
		return
	fi

	url=${BOX2D_UPSTREAM_DIR:-$UPSTREAM_URL}
	echo "oracle: fetching upstream $pin from $url"
	rm -rf "$here/.upstream"
	mkdir -p "$upstream"
	git -C "$upstream" init --quiet
	git -C "$upstream" remote add origin "$url"
	if ! git -C "$upstream" fetch --quiet --depth 1 origin "$pin" 2>/dev/null; then
		git -C "$upstream" fetch --quiet origin
	fi
	git -C "$upstream" -c advice.detachedHead=false checkout --quiet --detach "$pin"
	[ "$(git -C "$upstream" rev-parse HEAD)" = "$pin" ] || die "checkout of $pin failed"
}

check_run() {
	run=$1
	prog=$2
	count=$(find "$run" -type f | wc -l | tr -d ' ')
	[ "$count" -gt 0 ] || die "$prog wrote no fixtures (ORACLE_OPEN missing?)"
	size=$(find "$run" -type f -exec cat {} + | wc -c | tr -d ' ')
	budget=$BUDGET_BYTES
	[ "$prog" != tests_scenes_benchmarks ] || budget=16777216
	[ "$size" -le "$budget" ] ||
		die "$prog wrote $size bytes of fixtures, over the $budget byte budget; hash large scenes per step"
}

build_variant() {
	name=$1
	precision=$2
	scalar=$3
	cc=$4
	dir="$build/$name"
	out="$build/out/$name"

	double=OFF
	[ "$precision" = double ] && double=ON
	disable_simd=OFF
	[ "$scalar" = scalar ] && disable_simd=ON

	echo "oracle: building $name ($precision, $scalar, $cc)"
	cmake -S "$here" -B "$dir" ${ORACLE_GENERATOR:+-G "$ORACLE_GENERATOR"} \
		-DCMAKE_BUILD_TYPE=RelWithDebInfo \
		-DCMAKE_C_COMPILER="$cc" \
		-DBOX2D_UPSTREAM_DIR="$upstream" \
		-DBOX2D_DISABLE_SIMD="$disable_simd" \
		-DBOX2D_DOUBLE_PRECISION="$double" \
		-DBOX2D_VALIDATE=OFF \
		>"$dir.configure.log" 2>&1 || {
		cat "$dir.configure.log" >&2
		die "cmake configure failed for $name"
	}
	cmake --build "$dir" --parallel >"$dir.build.log" 2>&1 || {
		cat "$dir.build.log" >&2
		die "build failed for $name"
	}

	mkdir -p "$out"
	for exe in "$dir"/bin/*; do
		[ -x "$exe" ] || continue
		prog=$(basename "$exe")
		run="$build/run/$name/$prog"
		rm -rf "$run"
		mkdir -p "$run"
		echo "oracle:   running $prog"
		ORACLE_OUT="$run" "$exe" || die "$prog failed in $name"
		check_run "$run" "$prog"
		(cd "$run" && find . -type f | sort) | while read -r file; do
			[ ! -e "$out/$file" ] || die "$prog writes fixture $file, which another program already wrote"
			mkdir -p "$out/$(dirname "$file")"
			cp "$run/$file" "$out/$file"
		done
	done
}

compare() {
	if diff -r -q "$build/out/$1" "$build/out/$2" >"$build/diff.txt" 2>&1; then
		echo "oracle: check passed: $3 gives identical fixtures"
		return
	fi
	cat "$build/diff.txt" >&2
	if [ "$ORACLE_STRICT" = 1 ]; then
		die "check failed: $3 gives different fixtures (listed above)"
	fi
	warn "$3 gives different fixtures (listed above). The scalar $ORACLE_CC build stays the reference."
}

fetch_upstream

rm -rf "$build"
mkdir -p "$build"

build_variant float float scalar "$ORACLE_CC"
build_variant double double scalar "$ORACLE_CC"

if [ "$ORACLE_CHECKS" = 1 ]; then
	build_variant float-simd float simd "$ORACLE_CC"
	compare float float-simd "default SIMD build (float)"
	build_variant double-simd double simd "$ORACLE_CC"
	compare double double-simd "default SIMD build (double)"

	if command -v "$ORACLE_CLANG" >/dev/null 2>&1; then
		build_variant float-clang float scalar "$ORACLE_CLANG"
		compare float float-clang "$ORACLE_CLANG (float)"
		build_variant double-clang double scalar "$ORACLE_CLANG"
		compare double double-clang "$ORACLE_CLANG (double)"
	else
		warn "$ORACLE_CLANG not found, the clang cross-check was skipped"
	fi
fi

for config in float double; do
	rm -rf "${fixtures:?}/$config"
	mkdir -p "$fixtures/$config"
	cp -R "$build/out/$config/." "$fixtures/$config/"
done

echo "oracle: fixtures written to $fixtures"
