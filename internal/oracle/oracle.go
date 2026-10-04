package oracle

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type Reader struct {
	t    testing.TB
	path string
	recs []record
	next int
}

type record struct {
	Key  string `json:"k"`
	Type string `json:"t"`
	Val  string `json:"v"`
}

func Open(t testing.TB, name string) *Reader {
	t.Helper()

	if name == "" || path.Clean(name) != name || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") {
		t.Fatalf("oracle: invalid fixture name %q", name)
	}

	root, err := repoRoot()
	if err != nil {
		t.Fatalf("oracle: %v", err)
	}

	file := filepath.Join(root, "testdata", "oracle", config, filepath.FromSlash(name)+".jsonl")
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("oracle: cannot read fixture %q: %v\nrun `just oracle` to generate fixtures", name, err)
	}

	reference, err := fixtureReference(root)
	if err != nil {
		t.Fatalf("oracle: %v\nrun `just oracle` to explicitly regenerate and review C expectations", err)
	}
	records, err := verifyCapture(reference, name, data)
	if err != nil {
		t.Fatalf("oracle: %v", err)
	}
	r := Read(t, "testdata/oracle/"+config+"/"+name+".jsonl", data)
	if r.Len() != records {
		t.Fatalf("oracle: fixture %q has %d records; completed C capture requires %d", name, r.Len(), records)
	}
	t.Logf("C reference fixture: %s (%d observations)", name, records)
	return r
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found above the working directory")
		}
		dir = parent
	}
}

// Read compares a complete typed capture supplied by a fresh C execution.
// Approved offline files should use Open, which also verifies their provenance.
func Read(t testing.TB, displayPath string, data []byte) *Reader {
	t.Helper()

	r := &Reader{t: t, path: displayPath}

	for i, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}

		rec, err := decodeRecord(line)
		if err != nil {
			t.Fatalf("oracle: %s:%d: malformed record: %v", displayPath, i+1, err)
		}

		r.recs = append(r.recs, rec)
	}

	t.Cleanup(func() {
		if t.Failed() || t.Skipped() {
			return
		}

		if r.next < len(r.recs) {
			rec := r.recs[r.next]
			t.Errorf("oracle: %s: %d of %d records were never read; the first is record %d, key %q (%s)",
				r.path, len(r.recs)-r.next, len(r.recs), r.next+1, rec.Key, rec.Type)
		}
	})

	return r
}

func decodeRecord(line []byte) (record, error) {
	var rec record
	dec := json.NewDecoder(bytes.NewReader(line))
	open, err := dec.Token()
	if err != nil || open != json.Delim('{') {
		return rec, fmt.Errorf("expected a record object")
	}
	var fields uint8
	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return rec, err
		}
		key, ok := token.(string)
		if !ok {
			return rec, fmt.Errorf("expected a string member")
		}
		var bit uint8
		switch key {
		case "k":
			bit = 1
		case "t":
			bit = 2
		case "v":
			bit = 4
		default:
			return rec, fmt.Errorf("unknown member %q", key)
		}
		if fields&bit != 0 {
			return rec, fmt.Errorf("duplicate member %q", key)
		}
		fields |= bit
		token, err = dec.Token()
		if err != nil {
			return rec, err
		}
		value, ok := token.(string)
		if !ok {
			return rec, fmt.Errorf("member %q must be a string", key)
		}
		switch key {
		case "k":
			rec.Key = value
		case "t":
			rec.Type = value
		case "v":
			rec.Val = value
		}
	}
	close, err := dec.Token()
	if err != nil || close != json.Delim('}') || fields != 7 {
		return rec, fmt.Errorf("record requires exactly k, t and v")
	}
	if _, err := dec.Token(); err != io.EOF {
		return rec, fmt.Errorf("trailing content after record")
	}
	switch rec.Type {
	case "f32", "vec2", "rot", "pos", "i32", "u32", "u64", "bool", "bytes":
	default:
		return rec, fmt.Errorf("unknown record type %q", rec.Type)
	}
	return rec, nil
}

func (r *Reader) Len() int { return len(r.recs) }

func (r *Reader) take(key, typ string) record {
	r.t.Helper()

	if r.next >= len(r.recs) {
		r.t.Fatalf("oracle: %s: the Go test checked key %q (%s) as record %d, but the fixture has only %d records",
			r.path, key, typ, r.next+1, len(r.recs))
	}

	rec := r.recs[r.next]
	if rec.Key != key || rec.Type != typ {
		r.t.Fatalf("oracle: %s: record %d does not match the order of checks\n"+
			"  fixture (C) has: key %q, type %s\n"+
			"  Go test checks:  key %q, type %s",
			r.path, r.next+1, rec.Key, rec.Type, key, typ)
	}

	r.next++
	return rec
}

func (r *Reader) bitsOf(rec record, field string, width int) uint64 {
	r.t.Helper()

	if len(field) != width/4 {
		r.t.Fatalf("oracle: %s: record %d, key %q: %s value %q is not %d hex digits",
			r.path, r.next, rec.Key, rec.Type, field, width/4)
	}

	v, err := strconv.ParseUint(field, 16, 64)
	if err != nil {
		r.t.Fatalf("oracle: %s: record %d, key %q: bad hex %q: %v", r.path, r.next, rec.Key, field, err)
	}
	return v
}

func (r *Reader) fields(rec record, n int) []string {
	r.t.Helper()

	parts := strings.Split(rec.Val, ",")
	if len(parts) != n {
		r.t.Fatalf("oracle: %s: record %d, key %q: %s value %q has %d components, want %d",
			r.path, r.next, rec.Key, rec.Type, rec.Val, len(parts), n)
	}
	return parts
}

func (r *Reader) mismatch(key, typ, component, expected, got, extra string) {
	r.t.Helper()

	var b strings.Builder
	fmt.Fprintf(&b, "oracle: first divergence in %s, record %d\n", r.path, r.next)
	fmt.Fprintf(&b, "  key:      %q\n", key)
	fmt.Fprintf(&b, "  type:     %s\n", typ)
	if component != "" {
		fmt.Fprintf(&b, "  field:    %s\n", component)
	}
	fmt.Fprintf(&b, "  expected: %s   (C oracle)\n", expected)
	fmt.Fprintf(&b, "  got:      %s   (Go)\n", got)
	b.WriteString(extra)

	r.t.Fatalf("%s", strings.TrimRight(b.String(), "\n"))
}

func (r *Reader) checkFloat(rec record, component, field string, width int, got uint64) {
	r.t.Helper()

	want := r.bitsOf(rec, field, width)
	if want == got {
		return
	}

	r.mismatch(rec.Key, rec.Type, component, formatFloat(want, width), formatFloat(got, width),
		fmt.Sprintf("  ulp distance: %s\n", ulpDistance(want, got, width)))
}

func (r *Reader) F32(key string, v float32) {
	r.t.Helper()

	rec := r.take(key, "f32")
	r.checkFloat(rec, "", rec.Val, 32, uint64(math.Float32bits(v)))
}

func (r *Reader) Vec2(key string, x, y float32) {
	r.t.Helper()

	rec := r.take(key, "vec2")
	parts := r.fields(rec, 2)
	r.checkFloat(rec, "x", parts[0], 32, uint64(math.Float32bits(x)))
	r.checkFloat(rec, "y", parts[1], 32, uint64(math.Float32bits(y)))
}

func (r *Reader) Rot(key string, c, s float32) {
	r.t.Helper()

	rec := r.take(key, "rot")
	parts := r.fields(rec, 2)
	r.checkFloat(rec, "c", parts[0], 32, uint64(math.Float32bits(c)))
	r.checkFloat(rec, "s", parts[1], 32, uint64(math.Float32bits(s)))
}

func (r *Reader) Pos(key string, x, y Real) {
	r.t.Helper()

	rec := r.take(key, "pos")
	parts := r.fields(rec, 2)
	r.checkFloat(rec, "x", parts[0], realWidth, realBits(x))
	r.checkFloat(rec, "y", parts[1], realWidth, realBits(y))
}

func (r *Reader) I32(key string, v int32) {
	r.t.Helper()

	rec := r.take(key, "i32")
	want := int32(uint32(r.bitsOf(rec, rec.Val, 32)))
	if want != v {
		r.mismatch(key, rec.Type, "",
			fmt.Sprintf("%d (0x%08x)", want, uint32(want)), fmt.Sprintf("%d (0x%08x)", v, uint32(v)),
			fmt.Sprintf("  difference: %d\n", int64(v)-int64(want)))
	}
}

func (r *Reader) U32(key string, v uint32) {
	r.t.Helper()

	rec := r.take(key, "u32")
	want := uint32(r.bitsOf(rec, rec.Val, 32))
	if want != v {
		r.mismatch(key, rec.Type, "",
			fmt.Sprintf("%d (0x%08x)", want, want), fmt.Sprintf("%d (0x%08x)", v, v), "")
	}
}

func (r *Reader) U64(key string, v uint64) {
	r.t.Helper()

	rec := r.take(key, "u64")
	want := r.bitsOf(rec, rec.Val, 64)
	if want != v {
		r.mismatch(key, rec.Type, "",
			fmt.Sprintf("%d (0x%016x)", want, want), fmt.Sprintf("%d (0x%016x)", v, v), "")
	}
}

func (r *Reader) Bool(key string, v bool) {
	r.t.Helper()

	rec := r.take(key, "bool")
	if rec.Val != "0" && rec.Val != "1" {
		r.t.Fatalf("oracle: %s: record %d, key %q: bool value %q is not 0 or 1", r.path, r.next, key, rec.Val)
	}

	want := rec.Val == "1"
	if want != v {
		r.mismatch(key, rec.Type, "", strconv.FormatBool(want), strconv.FormatBool(v), "")
	}
}

func (r *Reader) Bytes(key string, v []byte) {
	r.t.Helper()

	rec := r.take(key, "bytes")
	want, err := hex.DecodeString(rec.Val)
	if err != nil {
		r.t.Fatalf("oracle: %s: record %d, key %q: bad bytes %q: %v", r.path, r.next, key, rec.Val, err)
	}

	if bytes.Equal(want, v) {
		return
	}

	i := 0
	for i < len(want) && i < len(v) && want[i] == v[i] {
		i++
	}

	extra := fmt.Sprintf("  first difference at byte %d; lengths %d (C) and %d (Go)\n", i, len(want), len(v))
	r.mismatch(key, rec.Type, "", formatBytes(want), formatBytes(v), extra)
}

func formatBytes(p []byte) string {
	const limit = 64
	if len(p) <= limit {
		return fmt.Sprintf("%d bytes 0x%x", len(p), p)
	}
	return fmt.Sprintf("%d bytes 0x%x...", len(p), p[:limit])
}

func formatFloat(bits uint64, width int) string {
	if width == 32 {
		return fmt.Sprintf("%s (0x%08x)", strconv.FormatFloat(float64(math.Float32frombits(uint32(bits))), 'g', -1, 32), uint32(bits))
	}
	return fmt.Sprintf("%s (0x%016x)", strconv.FormatFloat(math.Float64frombits(bits), 'g', -1, 64), bits)
}

func ulpDistance(a, b uint64, width int) string {
	if isNaNBits(a, width) || isNaNBits(b, width) {
		return "n/a (NaN)"
	}

	oa, ob := orderedFloat(a, width), orderedFloat(b, width)
	if oa < ob {
		oa, ob = ob, oa
	}

	return strconv.FormatUint(uint64(oa)-uint64(ob), 10)
}

func orderedFloat(bits uint64, width int) int64 {
	sign := uint64(1) << (width - 1)
	if bits&sign != 0 {
		return -int64(bits&^sign) - 1
	}
	return int64(bits)
}

func isNaNBits(bits uint64, width int) bool {
	if width == 32 {
		return math.IsNaN(float64(math.Float32frombits(uint32(bits))))
	}
	return math.IsNaN(math.Float64frombits(bits))
}
