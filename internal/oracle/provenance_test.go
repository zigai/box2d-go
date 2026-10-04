package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testReference(t *testing.T) (string, string, referenceManifest, []byte) {
	t.Helper()
	root := t.TempDir()
	pin := strings.Repeat("a", 40)
	if err := os.WriteFile(filepath.Join(root, "UPSTREAM"), []byte(pin+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := adapterSources(root)
	if err != nil {
		t.Fatal(err)
	}
	reference := referenceManifest{
		Schema: 1, Upstream: pin, Precision: config, Scalar: true, Assertions: true, Contraction: "off",
		Compiler: "unit transport fixture", CompilerVersion: "1", CompilerTarget: "unit",
		BuildType: "RelWithDebInfo", Flags: "-ffp-contract=off",
		Sources: make(map[string]string), Fixtures: make(map[string]fixtureIdentity),
	}
	for _, name := range files {
		file := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("unit adapter identity: " + name)
		if err := os.WriteFile(file, content, 0o644); err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		reference.Sources[name] = hex.EncodeToString(digest[:])
	}
	capture := []byte("{\"k\":\"result\",\"t\":\"f32\",\"v\":\"3f000000\"}\n")
	digest := sha256.Sum256(capture)
	reference.Fixtures["unit"] = fixtureIdentity{
		Producer: "unit transport", Records: 1, SHA256: hex.EncodeToString(digest[:]), Completed: true,
	}
	file := filepath.Join(root, "manifest.json")
	storeReference(t, file, reference)
	return root, file, reference, capture
}

func storeReference(t *testing.T, file string, reference referenceManifest) {
	t.Helper()
	data, err := json.Marshal(reference)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReferenceRejectsStalePinProfileAndAdapters(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testing.T, string, string, *referenceManifest)
	}{
		{"missing", func(t *testing.T, _, file string, _ *referenceManifest) {
			if err := os.Remove(file); err != nil {
				t.Fatal(err)
			}
		}},
		{"pin", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Upstream = strings.Repeat("b", 40) }},
		{"schema", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Schema = 2 }},
		{"precision", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Precision = "wrong" }},
		{"SIMD", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Scalar = false }},
		{"assertions", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Assertions = false }},
		{"build_type", func(_ *testing.T, _, _ string, r *referenceManifest) { r.BuildType = "Release" }},
		{"contraction", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Contraction = "fast" }},
		{"compiler", func(_ *testing.T, _, _ string, r *referenceManifest) { r.CompilerVersion = "" }},
		{"adapter_changed", func(t *testing.T, root, _ string, _ *referenceManifest) {
			if err := os.WriteFile(filepath.Join(root, "oracle", "oracle.c"), []byte("changed adapter"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"adapter_added", func(t *testing.T, root, _ string, _ *referenceManifest) {
			if err := os.MkdirAll(filepath.Join(root, "oracle", "tests"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, "oracle", "tests", "added.c"), []byte("new case"), 0o644); err != nil {
				t.Fatal(err)
			}
		}},
		{"adapter_missing", func(t *testing.T, root, _ string, _ *referenceManifest) {
			if err := os.Remove(filepath.Join(root, "oracle", "oracle.c")); err != nil {
				t.Fatal(err)
			}
		}},
		{"empty_captures", func(_ *testing.T, _, _ string, r *referenceManifest) { r.Fixtures = nil }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, file, reference, _ := testReference(t)
			if _, err := loadReference(root, file); err != nil {
				t.Fatalf("valid transport provenance rejected: %v", err)
			}
			c.mutate(t, root, file, &reference)
			if c.name != "missing" {
				storeReference(t, file, reference)
			}
			if _, err := loadReference(root, file); err == nil {
				t.Fatal("stale or incomplete reference was accepted")
			}
		})
	}
}

func TestCaptureRejectsMutationMissingAndIncompleteRun(t *testing.T) {
	_, _, reference, data := testReference(t)
	count, err := verifyCapture(&reference, "unit", data)
	if err != nil || count != 1 {
		t.Fatalf("valid capture: count=%d, err=%v", count, err)
	}
	r := Read(t, "approved transport", data)
	r.F32("result", 0.5)
	cases := []struct {
		name  string
		input []byte
	}{
		{"changed_value", []byte("{\"k\":\"result\",\"t\":\"f32\",\"v\":\"3f800000\"}\n")},
		{"truncated", data[:len(data)-2]},
		{"extra_record", append(append([]byte(nil), data...), data...)},
		{"empty", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := verifyCapture(&reference, "unit", c.input); err == nil {
				t.Fatal("unapproved capture bytes accepted")
			}
		})
	}
	if _, err := verifyCapture(&reference, "absent", data); err == nil {
		t.Fatal("missing capture identity accepted")
	}
	identity := reference.Fixtures["unit"]
	identity.Completed = false
	reference.Fixtures["unit"] = identity
	if _, err := verifyCapture(&reference, "unit", data); err == nil {
		t.Fatal("interrupted C capture accepted")
	}
}
