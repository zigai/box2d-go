package oracle

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func sprintf(format string, args ...any) string { return fmt.Sprintf(format, args...) }

func fixture(lines ...string) []byte { return []byte(strings.Join(lines, "\n") + "\n") }

func readerOf(f *fakeTB, lines ...string) *Reader {
	return newReader(f, "testdata/oracle/float/test.jsonl", fixture(lines...))
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
		{"vec2 y", func(r *Reader) { r.Vec2("v", 1, 2.5) }, []string{"type:     vec2", "field:    y", "ulp distance: 2097152"}},
		{"rot c", func(r *Reader) { r.Rot("v", 0.5, 2) }, []string{"type:     rot", "field:    c"}},
		{"pos x", func(r *Reader) { r.Pos("v", 9, 2) }, []string{"type:     pos", "field:    x"}},
	}

	lines := map[string]string{
		"vec2 y": `{"k":"v","t":"vec2","v":"3f800000,40000000"}`,
		"rot c":  `{"k":"v","t":"rot","v":"3f800000,40000000"}`,
	}
	posLine := `{"k":"v","t":"pos","v":"3f800000,40000000"}`
	if DoublePrecision {
		posLine = `{"k":"v","t":"pos","v":"3ff0000000000000,4000000000000000"}`
	}
	lines["pos x"] = posLine

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

func TestReaderMalformedFixture(t *testing.T) {
	for line, want := range map[string]string{
		`not json`:                              "malformed record",
		`{"k":"a","t":"f32","v":"0","extra":1}`: "malformed record",
		`{"k":"a","v":"0"}`:                     "record has no type",
	} {
		f := &fakeTB{}
		f.run(func() { readerOf(f, line) })
		mustContain(t, f.fatal, ":1:", want)
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
