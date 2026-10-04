package oracle

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }

func fixture(lines ...string) []byte { return []byte(strings.Join(lines, "\n") + "\n") }

func readerOf(f *fakeTB, lines ...string) *Reader {
	return Read(f, "testdata/oracle/float/test.jsonl", fixture(lines...))
}

func mustContain(t *testing.T, got string, wants ...string) {
	t.Helper()

	for _, w := range wants {
		if !strings.Contains(got, w) {
			t.Errorf("message lacks %q:\n%s", w, got)
		}
	}
}

func TestReaderFirstDivergence(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		r := readerOf(f,
			`{"k":"a","t":"f32","v":"3f800000"}`,
			`{"k":"b","t":"f32","v":"3f000000"}`,
			`{"k":"c","t":"i32","v":"00000001"}`,
		)
		r.F32("a", 1)
		r.F32("b", math.Float32frombits(0x3f000002))
		r.I32("c", 1)
	})

	mustContain(t, f.fatal,
		"first divergence in testdata/oracle/float/test.jsonl, record 2",
		`key:      "b"`,
		"type:     f32",
		"expected: 0.5 (0x3f000000)   (C oracle)",
		"got:      0.5000001 (0x3f000002)   (Go)",
		"ulp distance: 2",
	)

	if len(f.errors) != 0 {
		t.Errorf("a fatal check must not also report unread records: %v", f.errors)
	}
}

func TestReaderComponents(t *testing.T) {
	tests := []struct {
		name  string
		check func(r *Reader)
		wants []string
	}{
		{"vec2 x", func(r *Reader) { r.Vec2("v", 1.5, 2) }, []string{"type:     vec2", "field:    x"}},
		{"vec2 y", func(r *Reader) { r.Vec2("v", 1, 2.5) }, []string{"type:     vec2", "field:    y", "ulp distance: 2097152"}},
		{"rot c", func(r *Reader) { r.Rot("v", 0.5, 2) }, []string{"type:     rot", "field:    c"}},
		{"rot s", func(r *Reader) { r.Rot("v", 1, 2.5) }, []string{"type:     rot", "field:    s"}},
		{"pos x", func(r *Reader) { r.Pos("v", 9, 2) }, []string{"type:     pos", "field:    x"}},
		{"pos y", func(r *Reader) { r.Pos("v", 1, 9) }, []string{"type:     pos", "field:    y"}},
	}

	lines := map[string]string{
		"vec2 x": `{"k":"v","t":"vec2","v":"3f800000,40000000"}`,
		"vec2 y": `{"k":"v","t":"vec2","v":"3f800000,40000000"}`,
		"rot c":  `{"k":"v","t":"rot","v":"3f800000,40000000"}`,
		"rot s":  `{"k":"v","t":"rot","v":"3f800000,40000000"}`,
	}

	posLine := `{"k":"v","t":"pos","v":"3f800000,40000000"}`
	if DoublePrecision {
		posLine = `{"k":"v","t":"pos","v":"3ff0000000000000,4000000000000000"}`
	}

	lines["pos x"] = posLine
	lines["pos y"] = posLine

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeTB{}
			f.run(func() { tt.check(readerOf(f, lines[tt.name])) })
			mustContain(t, f.fatal, tt.wants...)
		})
	}
}

func TestReaderSignedZeroIsOneUlp(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		readerOf(f, `{"k":"z","t":"f32","v":"00000000"}`).F32("z", float32(math.Copysign(0, -1)))
	})
	mustContain(t, f.fatal, "expected: 0 (0x00000000)", "got:      -0 (0x80000000)", "ulp distance: 1")
}

func TestReaderNaN(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		readerOf(f, `{"k":"n","t":"f32","v":"7fc00000"}`).F32("n", math.Float32frombits(0x7fc00001))
	})
	mustContain(t, f.fatal, "expected: NaN (0x7fc00000)", "got:      NaN (0x7fc00001)", "ulp distance: n/a (NaN)")
}

func TestReaderUlpDistance(t *testing.T) {
	tests := []struct {
		a, b  uint64
		width int
		want  string
	}{
		{0x3f800000, 0x3f800001, 32, "1"},
		{0x3f800001, 0x3f800000, 32, "1"},
		{0x00000000, 0x80000000, 32, "1"},
		{0x00000001, 0x80000001, 32, "3"},
		{0x7f7fffff, 0xff7fffff, 32, "4278190079"},
		{0x3ff0000000000000, 0x3ff0000000000003, 64, "3"},
		{0x0000000000000000, 0x8000000000000000, 64, "1"},
		{0x7fefffffffffffff, 0xffefffffffffffff, 64, "18437736874454810623"},
	}
	for _, tt := range tests {
		if got := ulpDistance(tt.a, tt.b, tt.width); got != tt.want {
			t.Errorf("ulpDistance(%#x, %#x, %d) = %s, want %s", tt.a, tt.b, tt.width, got, tt.want)
		}
	}
}

func TestReaderKeyAndTypeMismatch(t *testing.T) {
	line := `{"k":"a","t":"f32","v":"3f800000"}`

	f := &fakeTB{}
	f.run(func() { readerOf(f, line).F32("b", 1) })
	mustContain(t, f.fatal, "record 1 does not match the order of checks", `key "a", type f32`, `key "b", type f32`)

	f = &fakeTB{}
	f.run(func() { readerOf(f, line).U32("a", 1) })
	mustContain(t, f.fatal, `key "a", type f32`, `key "a", type u32`)
}

func TestReaderPastTheEnd(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		r := readerOf(f, `{"k":"a","t":"bool","v":"1"}`)
		r.Bool("a", true)
		r.Bool("b", true)
	})
	mustContain(t, f.fatal, `checked key "b" (bool) as record 2, but the fixture has only 1 records`)
}

func TestReaderUnreadRecordsFailAtCleanup(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		r := readerOf(f,
			`{"k":"a","t":"bool","v":"1"}`,
			`{"k":"b","t":"bool","v":"0"}`,
			`{"k":"c","t":"bool","v":"0"}`,
		)
		r.Bool("a", true)
	})

	if f.fatal != "" {
		t.Fatalf("unexpected fatal: %s", f.fatal)
	}

	if len(f.errors) != 1 {
		t.Fatalf("got %d errors, want 1: %v", len(f.errors), f.errors)
	}

	mustContain(t, f.errors[0], "2 of 3 records were never read", `record 2, key "b" (bool)`)
}

func TestReaderUnreadRecordsAreIgnoredWhenSkippedOrFailed(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		readerOf(f, `{"k":"a","t":"bool","v":"1"}`)
		f.skipped = true
	})

	if len(f.errors) != 0 {
		t.Errorf("a skipped test must not report unread records: %v", f.errors)
	}

	f = &fakeTB{}
	f.run(func() {
		readerOf(f, `{"k":"a","t":"bool","v":"1"}`)
		f.failed = true
	})

	if len(f.errors) != 0 {
		t.Errorf("a failed test must not report unread records: %v", f.errors)
	}
}

func TestReaderIntegersBoolsBytes(t *testing.T) {
	f := &fakeTB{}
	f.run(func() {
		r := readerOf(f,
			`{"k":"i","t":"i32","v":"ffffffff"}`,
			`{"k":"u","t":"u32","v":"00000007"}`,
			`{"k":"w","t":"u64","v":"8000000000000001"}`,
			`{"k":"t","t":"bool","v":"1"}`,
			`{"k":"b","t":"bytes","v":"00017f"}`,
			`{"k":"e","t":"bytes","v":""}`,
			`{"k":"x","t":"u64","v":"0000000000000001"}`,
		)
		r.I32("i", -1)
		r.U32("u", 7)
		r.U64("w", 1<<63|1)
		r.Bool("t", true)
		r.Bytes("b", []byte{0, 1, 0x7f})
		r.Bytes("e", nil)
		r.U64("x", 2)
	})

	mustContain(t, f.fatal, `key:      "x"`, "expected: 1 (0x0000000000000001)", "got:      2 (0x0000000000000002)")

	f = &fakeTB{}
	f.run(func() {
		readerOf(f, `{"k":"b","t":"bytes","v":"00017f"}`).Bytes("b", []byte{0, 1, 0x7e, 9})
	})
	mustContain(t, f.fatal, "expected: 3 bytes 0x00017f", "got:      4 bytes 0x00017e09", "first difference at byte 2; lengths 3 (C) and 4 (Go)")

	f = &fakeTB{}
	f.run(func() { readerOf(f, `{"k":"i","t":"i32","v":"80000000"}`).I32("i", math.MaxInt32) })
	mustContain(t, f.fatal, "expected: -2147483648 (0x80000000)", "got:      2147483647 (0x7fffffff)")

	f = &fakeTB{}
	f.run(func() { readerOf(f, `{"k":"t","t":"bool","v":"0"}`).Bool("t", true) })
	mustContain(t, f.fatal, "expected: false", "got:      true")
}

func TestReaderRejectsMalformedRecordGrammar(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{"not JSON", `not json`},
		{"unknown member", `{"k":"a","t":"bytes","v":"","extra":1}`},
		{"missing key", `{"t":"bytes","v":""}`},
		{"missing type", `{"k":"a","v":""}`},
		{"missing value", `{"k":"a","t":"bytes"}`},
		{"null key", `{"k":null,"t":"bytes","v":""}`},
		{"null type", `{"k":"a","t":null,"v":""}`},
		{"null value", `{"k":"a","t":"bytes","v":null}`},
		{"numeric key", `{"k":1,"t":"bytes","v":""}`},
		{"boolean type", `{"k":"a","t":true,"v":""}`},
		{"array value", `{"k":"a","t":"bytes","v":[]}`},
		{"object key", `{"k":{},"t":"bytes","v":""}`},
		{"array type", `{"k":"a","t":[],"v":""}`},
		{"numeric value", `{"k":"a","t":"bytes","v":0}`},
		{"duplicate key", `{"k":"ignored","k":"a","t":"bytes","v":""}`},
		{"duplicate type", `{"k":"a","t":"u32","t":"bytes","v":""}`},
		{"duplicate value", `{"k":"a","t":"bytes","v":"ff","v":""}`},
		{"identical duplicate", `{"k":"a","t":"bytes","v":"","v":""}`},
		{"escaped duplicate member", `{"k":"a","t":"bytes","v":"","\u0076":""}`},
		{"unknown type", `{"k":"a","t":"unknown","v":""}`},
		{"empty type", `{"k":"a","t":"","v":""}`},
		{"second object", `{"k":"a","t":"bytes","v":""} {"k":"hidden","t":"bool","v":"1"}`},
		{"trailing null", `{"k":"a","t":"bytes","v":""} null`},
		{"trailing token", `{"k":"a","t":"bytes","v":""} garbage`},
		{"trailing comment", `{"k":"a","t":"bytes","v":""} // ignored`},
		{"truncated object", `{"k":"a","t":"bytes","v":""`},
		{"truncated string", `{"k":"a","t":"bytes","v":"`},
		{"top-level array", `[{"k":"a","t":"bytes","v":""}]`},
		{"top-level null", `null`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeTB{}
			f.run(func() {
				Read(f, "capture.jsonl", fixture(`{"k":"prefix","t":"bool","v":"1"}`, tt.line))
			})

			if f.fatal == "" {
				t.Fatalf("accepted malformed record %s; cleanup errors are not schema rejection: %v", tt.line, f.errors)
			}

			mustContain(t, f.fatal, "capture.jsonl:2:")
		})
	}
}

func TestReaderBadHexWidth(t *testing.T) {
	f := &fakeTB{}
	f.run(func() { readerOf(f, `{"k":"a","t":"f32","v":"3f80"}`).F32("a", 1) })
	mustContain(t, f.fatal, `f32 value "3f80" is not 8 hex digits`)
}

func TestOpenErrors(t *testing.T) {
	f := &fakeTB{}
	f.run(func() { Open(f, "no_such_fixture") })
	mustContain(t, f.fatal, `cannot read fixture "no_such_fixture"`, "run `just oracle`")

	for _, name := range []string{"", "../x", "/abs", "a/../b"} {
		f = &fakeTB{}
		f.run(func() { Open(f, name) })
		mustContain(t, f.fatal, "invalid fixture name")
	}
}

func TestHashKnownAnswers(t *testing.T) {
	for _, tt := range []struct {
		in   string
		want uint64
	}{
		{"", 0xcbf29ce484222325},
		{"a", 0xaf63dc4c8601ec8c},
		{"foobar", 0x85944171f73967e8},
	} {
		if got := HashBytes(HashInit, []byte(tt.in)); got != tt.want {
			t.Errorf("HashBytes(%q) = %#x, want %#x", tt.in, got, tt.want)
		}
	}

	if got, want := HashU32(HashInit, 0x01020304), HashBytes(HashInit, []byte{4, 3, 2, 1}); got != want {
		t.Errorf("HashU32 = %#x, want %#x", got, want)
	}

	if got, want := HashU64(HashInit, 0x0102030405060708), HashBytes(HashInit, []byte{8, 7, 6, 5, 4, 3, 2, 1}); got != want {
		t.Errorf("HashU64 = %#x, want %#x", got, want)
	}
}

func TestReaderRejectsMalformedValues(t *testing.T) {
	posOne := "3f800000"

	posWrongWidth := "3ff0000000000000"
	if DoublePrecision {
		posOne, posWrongWidth = posWrongWidth, posOne
	}

	tests := []struct {
		name  string
		typ   string
		value string
		check func(*Reader)
	}{
		{"f32 short", "f32", "3f80", func(r *Reader) { r.F32("value", 1) }},
		{"f32 long", "f32", "03f800000", func(r *Reader) { r.F32("value", 1) }},
		{"f32 nonhex", "f32", "gg800000", func(r *Reader) { r.F32("value", 1) }},
		{"f32 signed", "f32", "+0000001", func(r *Reader) { r.F32("value", math.Float32frombits(1)) }},
		{"f32 prefix", "f32", "0x800000", func(r *Reader) { r.F32("value", 0) }},
		{"i32 short", "i32", "fffffff", func(r *Reader) { r.I32("value", -1) }},
		{"i32 overflow", "i32", "100000000", func(r *Reader) { r.I32("value", 0) }},
		{"u32 short", "u32", "0000000", func(r *Reader) { r.U32("value", 0) }},
		{"u32 nonhex", "u32", "0000000g", func(r *Reader) { r.U32("value", 0) }},
		{"u64 short", "u64", "000000000000000", func(r *Reader) { r.U64("value", 0) }},
		{"u64 overflow", "u64", "10000000000000000", func(r *Reader) { r.U64("value", 0) }},
		{"u64 nonhex", "u64", "000000000000000g", func(r *Reader) { r.U64("value", 0) }},
		{"bool integer", "bool", "2", func(r *Reader) { r.Bool("value", false) }},
		{"bool word", "bool", "true", func(r *Reader) { r.Bool("value", true) }},
		{"bool empty", "bool", "", func(r *Reader) { r.Bool("value", false) }},
		{"bytes odd", "bytes", "0", func(r *Reader) { r.Bytes("value", nil) }},
		{"bytes nonhex", "bytes", "0g", func(r *Reader) { r.Bytes("value", nil) }},
		{"vec2 one component", "vec2", "3f800000", func(r *Reader) { r.Vec2("value", 1, 1) }},
		{"vec2 three components", "vec2", "3f800000,3f800000,3f800000", func(r *Reader) { r.Vec2("value", 1, 1) }},
		{"vec2 empty x", "vec2", ",3f800000", func(r *Reader) { r.Vec2("value", 1, 1) }},
		{"vec2 bad y", "vec2", "3f800000,gg800000", func(r *Reader) { r.Vec2("value", 1, 1) }},
		{"rot one component", "rot", "3f800000", func(r *Reader) { r.Rot("value", 1, 1) }},
		{"rot three components", "rot", "3f800000,3f800000,3f800000", func(r *Reader) { r.Rot("value", 1, 1) }},
		{"rot empty s", "rot", "3f800000,", func(r *Reader) { r.Rot("value", 1, 1) }},
		{"pos one component", "pos", posOne, func(r *Reader) { r.Pos("value", 1, 1) }},
		{"pos three components", "pos", posOne + "," + posOne + "," + posOne, func(r *Reader) { r.Pos("value", 1, 1) }},
		{"pos wrong precision x", "pos", posWrongWidth + "," + posOne, func(r *Reader) { r.Pos("value", 1, 1) }},
		{"pos wrong precision y", "pos", posOne + "," + posWrongWidth, func(r *Reader) { r.Pos("value", 1, 1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeTB{}
			f.run(func() {
				r := readerOf(f, fmt.Sprintf(`{"k":"value","t":%q,"v":%q}`, tt.typ, tt.value))
				tt.check(r)
			})

			if f.fatal == "" {
				t.Fatalf("accepted invalid %s encoding %q", tt.typ, tt.value)
			}

			mustContain(t, f.fatal, "testdata/oracle/float/test.jsonl", "record 1", `"value"`)
		})
	}
}

func TestReaderAcceptsRepeatedKeysAndCompletedFinalLine(t *testing.T) {
	for _, ending := range []string{"", "\n", "\n\n", "\t \r\n"} {
		t.Run(fmt.Sprintf("ending %q", ending), func(t *testing.T) {
			data := []byte(`{"v":"","t":"bytes","k":""}` + "\n" +
				`{"k":"sample","t":"i32","v":"00000001"}` + "\n" +
				`{"k":"sample","t":"i32","v":"00000002"}` + ending)
			f := &fakeTB{}
			f.run(func() {
				r := Read(f, "completed.jsonl", data)
				if r.Len() != 3 {
					t.Errorf("decoded %d records, want 3", r.Len())
				}

				r.Bytes("", nil)
				r.I32("sample", 1)
				r.I32("sample", 2)
			})

			if f.fatal != "" || len(f.errors) != 0 {
				t.Fatalf("valid repeated keys/empty bytes rejected: fatal=%q errors=%v", f.fatal, f.errors)
			}
		})
	}
}

func TestReaderRejectsInteriorRecordMutationsWithRepeatedKeys(t *testing.T) {
	first := `{"k":"sample","t":"i32","v":"00000001"}`
	second := `{"k":"sample","t":"i32","v":"00000002"}`
	marker := `{"k":"checkpoint","t":"bool","v":"1"}`
	third := `{"k":"sample","t":"i32","v":"00000003"}`
	tail := `{"k":"tail","t":"u32","v":"00000008"}`

	tests := []struct {
		name  string
		lines []string
	}{
		{"insert duplicate interior", []string{first, first, second, marker, third, tail}},
		{"delete repeated interior", []string{first, marker, third, tail}},
		{"reorder repeated interior", []string{first, third, marker, second, tail}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeTB{}
			f.run(func() {
				r := readerOf(f, tt.lines...)
				r.I32("sample", 1)
				r.I32("sample", 2)
				r.Bool("checkpoint", true)
				r.I32("sample", 3)
				r.U32("tail", 8)
			})
			mustContain(t, f.fatal, "record 2", "sample", "i32")

			if len(f.errors) != 0 {
				t.Errorf("first divergence obscured by cleanup: %v", f.errors)
			}
		})
	}

	t.Run("extra repeated tail", func(t *testing.T) {
		f := &fakeTB{}
		f.run(func() {
			r := readerOf(f, first, second, marker, third, tail, third)
			r.I32("sample", 1)
			r.I32("sample", 2)
			r.Bool("checkpoint", true)
			r.I32("sample", 3)
			r.U32("tail", 8)
		})

		if f.fatal != "" || len(f.errors) != 1 {
			t.Fatalf("extra record should fail exhaustion once: fatal=%q errors=%v", f.fatal, f.errors)
		}

		mustContain(t, f.errors[0], "1 of 6 records", "record 6", `"sample"`, "i32")
	})
}

func TestReaderPositionExactBits(t *testing.T) {
	tests := []struct {
		name          string
		want32, got32 uint32
		want64, got64 uint64
	}{
		{"finite one bit", 0x3f800000, 0x3f800001, 0x3ff0000000000000, 0x3ff0000000000001},
		{"signed zero", 0, 0x80000000, 0, 0x8000000000000000},
		{"NaN payload", 0x7fc00000, 0x7fc00001, 0x7ff8000000000000, 0x7ff8000000000001},
		{"infinity versus largest finite", 0x7f800000, 0x7f7fffff, 0x7ff0000000000000, 0x7fefffffffffffff},
		{"subnormal one bit", 1, 2, 1, 2},
	}
	for _, tt := range tests {
		for _, component := range []string{"x", "y"} {
			t.Run(tt.name+"/"+component, func(t *testing.T) {
				wantHex := fmt.Sprintf("%08x", tt.want32)
				gotHex := fmt.Sprintf("%08x", tt.got32)
				want := Real(math.Float32frombits(tt.want32))
				got := Real(math.Float32frombits(tt.got32))
				zeroHex := "00000000"

				if DoublePrecision {
					wantHex = fmt.Sprintf("%016x", tt.want64)
					gotHex = fmt.Sprintf("%016x", tt.got64)
					want = Real(math.Float64frombits(tt.want64))
					got = Real(math.Float64frombits(tt.got64))
					zeroHex = "0000000000000000"
				}

				value := wantHex + "," + zeroHex
				x, y := got, Real(0)

				if component == "y" {
					value = zeroHex + "," + wantHex
					x, y = 0, got
				}

				f := &fakeTB{}
				f.run(func() {
					readerOf(f, fmt.Sprintf(`{"k":"position","t":"pos","v":%q}`, value)).Pos("position", x, y)
				})
				mustContain(t, f.fatal, "record 1", `"position"`, "type:     pos", "field:    "+component, "0x"+wantHex, "0x"+gotHex)

				x, y = want, 0
				if component == "y" {
					x, y = 0, want
				}

				f = &fakeTB{}
				f.run(func() {
					readerOf(f, fmt.Sprintf(`{"k":"position","t":"pos","v":%q}`, value)).Pos("position", x, y)
				})

				if f.fatal != "" || len(f.errors) != 0 {
					t.Fatalf("identical position bits rejected: fatal=%q errors=%v", f.fatal, f.errors)
				}
			})
		}
	}
}

func TestReaderBytesDistinguishesLongSuffixAndPrefixLength(t *testing.T) {
	tests := []struct {
		name   string
		want   []byte
		got    []byte
		offset int
	}{
		{"suffix beyond display limit", append(bytes.Repeat([]byte{0x12}, 64), 0x34), append(bytes.Repeat([]byte{0x12}, 64), 0x35), 64},
		{"missing last byte", []byte{1, 2, 3}, []byte{1, 2}, 2},
		{"extra last byte", []byte{1, 2}, []byte{1, 2, 3}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeTB{}
			f.run(func() {
				readerOf(f, fmt.Sprintf(`{"k":"bytes","t":"bytes","v":%q}`, hex.EncodeToString(tt.want))).Bytes("bytes", tt.got)
			})
			mustContain(t, f.fatal, "record 1", `"bytes"`, fmt.Sprintf("first difference at byte %d", tt.offset),
				fmt.Sprintf("lengths %d (C) and %d (Go)", len(tt.want), len(tt.got)))
		})
	}
}

func TestReadComparisonLeavesRealBaselineUnchanged(t *testing.T) {
	baseline := fixture(`{"k":"sample","t":"u32","v":"00000007"}`)

	for _, tt := range []struct {
		name  string
		got   uint32
		fails bool
	}{
		{"match", 7, false},
		{"mismatch", 8, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "baseline.jsonl")
			if err := os.WriteFile(path, baseline, 0o444); err != nil {
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

			inputCopy := bytes.Clone(data)
			f := &fakeTB{}
			f.run(func() { Read(f, path, data).U32("sample", tt.got) })

			if (f.fatal != "") != tt.fails || len(f.errors) != 0 {
				t.Fatalf("comparison outcome: fatal=%q errors=%v, want failure=%v", f.fatal, f.errors, tt.fails)
			}

			if tt.fails {
				mustContain(t, f.fatal, path, "record 1", `"sample"`, "type:     u32")
			}

			afterData, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			after, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(afterData, baseline) || !bytes.Equal(data, inputCopy) {
				t.Fatal("comparison rewrote the baseline or modified caller-owned capture bytes")
			}

			if !os.SameFile(before, after) || before.Mode() != after.Mode() || !before.ModTime().Equal(after.ModTime()) {
				t.Fatal("comparison replaced or modified the read-only baseline")
			}
		})
	}
}
