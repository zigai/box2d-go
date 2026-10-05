package oracle

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestTypes(t *testing.T) {
	checkTypes(Open(t, "types"))
}

func checkTypes(r *Reader) {
	r.F32("f32/one", 1)
	r.F32("f32/neg_zero", float32(math.Copysign(0, -1)))
	r.F32("f32/pi", math.Float32frombits(0x40490fdb))
	r.F32("f32/denormal", math.Float32frombits(1))
	r.F32("f32/inf", float32(math.Inf(1)))
	r.F32("f32/neg_inf", float32(math.Inf(-1)))
	r.F32("f32/nan", math.Float32frombits(0x7fc00000))
	r.F32("f32/max", math.MaxFloat32)

	r.Vec2("vec2/basic", 1.5, -2.25)
	r.Vec2("vec2/zeros", 0, float32(math.Copysign(0, -1)))

	r.Rot("rot/identity", 1, 0)
	r.Rot("rot/basic", 0.6, -0.8)

	negZero := Real(math.Copysign(0, -1))

	r.Pos("pos/basic", Real(0.1), Real(-1.0e-3))
	r.Pos("pos/far", Real(123456.789), negZero)

	r.I32("i32/zero", 0)
	r.I32("i32/minus_one", -1)
	r.I32("i32/min", math.MinInt32)
	r.I32("i32/max", math.MaxInt32)

	r.U32("u32/zero", 0)
	r.U32("u32/max", math.MaxUint32)

	r.U64("u64/zero", 0)
	r.U64("u64/high", 0x8000000000000001)
	r.U64("u64/max", math.MaxUint64)

	r.Bool("bool/true", true)
	r.Bool("bool/false", false)

	r.Bytes("bytes/five", []byte{0x00, 0x01, 0x7f, 0x80, 0xff})
	r.Bytes("bytes/empty", []byte{})

	r.Bool("key \"quoted\" \\ back\tslash/\u00e9", true)

	r.U64("hash/init", HashInit)
	r.U64("hash/bytes", HashBytes(HashInit, []byte("abc")))
	r.U64("hash/u32", HashU32(HashInit, 0x01020304))
	r.U64("hash/u64", HashU64(HashInit, 0x0102030405060708))
	r.U64("hash/f32", HashF32(HashInit, 1.5))
	r.U64("hash/f64", HashF64(HashInit, -2.5))
	r.U64("hash/pos", HashPos(HashInit, 1.5, -2.25))
	r.U64("hash/rot", HashRot(HashInit, 0.6, -0.8))
	r.U64("hash/transform", HashTransform(HashInit, 1.5, -2.25, 0.6, -0.8))
}

func TestTypesRealCaptureComparisonIsReadOnly(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}

	capture, err := os.ReadFile(filepath.Join(root, "testdata", "oracle", config, "types.jsonl"))
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "types.jsonl")
	if err := os.WriteFile(path, capture, 0o400); err != nil { //nolint:gosec // G703: fixed filename under t.TempDir's private test-owned tree.
		t.Fatal(err)
	}

	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	checkTypes(Read(t, path, data))
	checkRealCaptureUnchanged(t, path, before, capture, data)
}

func checkRealCaptureUnchanged(t *testing.T, path string, before os.FileInfo, capture, data []byte) {
	t.Helper()

	afterData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(capture, afterData) || !bytes.Equal(capture, data) {
		t.Fatal("comparison modified real C capture or caller-owned input bytes")
	}

	if !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("comparison replaced or modified read-only real C capture")
	}
}
