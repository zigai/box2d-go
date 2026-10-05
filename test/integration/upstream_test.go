//go:build box2d_oracle

package integration_test

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/zigai/box2d"
)

type conformanceCase struct {
	Go      string `json:"go"`
	Fixture string `json:"fixture"`
}

type conformanceSuite struct {
	Go      string                     `json:"go"`
	Fixture string                     `json:"fixture"`
	Reason  string                     `json:"reason"`
	Cases   map[string]conformanceCase `json:"cases"`
}

type conformanceCatalog struct {
	Schema   int                         `json:"schema"`
	Upstream string                      `json:"upstream"`
	Suites   map[string]conformanceSuite `json:"suites"`
	Disabled []struct {
		Go string `json:"go"`
	} `json:"disabled_upstream"`
	Diagnostics []conformanceCase `json:"unregistered_diagnostics"`
	Families    []struct {
		Subject string   `json:"subject"`
		Go      []string `json:"go"`
	} `json:"public_families"`
}

func repoRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "UPSTREAM")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot find Box2D repository root above working directory")
		}
		dir = parent
	}
}

func oracleBuildDir(t testing.TB) string {
	t.Helper()
	root := repoRoot(t)
	base := os.Getenv("BOX2D_ORACLE_DIR")
	if base == "" {
		base = filepath.Join(root, "oracle", ".build")
	} else if !filepath.IsAbs(base) {
		base = filepath.Join(root, base)
	}
	precision := "float"
	if box2d.IsDoublePrecision() {
		precision = "double"
	}
	dir := filepath.Join(base, precision)
	data, err := os.ReadFile(filepath.Join(dir, "reference.json"))
	if err != nil {
		t.Fatalf("explicit box2d_oracle mode requires an independently built C reference in %s: %v; run just oracle", dir, err)
	}
	var identity struct {
		Upstream    string            `json:"upstream"`
		Precision   string            `json:"precision"`
		Scalar      bool              `json:"scalar"`
		Assertions  bool              `json:"assertions"`
		BuildType   string            `json:"build_type"`
		Contraction string            `json:"contraction"`
		Sources     map[string]string `json:"sources"`
	}
	if err := json.Unmarshal(data, &identity); err != nil {
		t.Fatal(err)
	}
	pin, err := os.ReadFile(filepath.Join(root, "UPSTREAM"))
	if err != nil {
		t.Fatal(err)
	}
	if identity.Upstream != strings.TrimSpace(string(pin)) || identity.Precision != precision ||
		!identity.Scalar || !identity.Assertions || identity.BuildType != "RelWithDebInfo" ||
		identity.Contraction != "off" || len(identity.Sources) == 0 {
		t.Fatalf("C build has wrong reference identity/profile: %s", data)
	}
	for name, expected := range identity.Sources {
		if !strings.HasPrefix(name, "oracle/") || filepath.ToSlash(filepath.Clean(name)) != name || strings.Contains(name, "../") {
			t.Fatalf("invalid C adapter provenance path %q", name)
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(content)
		if hex.EncodeToString(digest[:]) != expected {
			t.Fatalf("C build predates adapter %s; explicitly rebuild with just oracle", name)
		}
	}
	return dir
}

func TestUpstreamSuite(t *testing.T) {
	root := repoRoot(t)
	build := oracleBuildDir(t)
	data, err := os.ReadFile(filepath.Join(root, "oracle", "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog conformanceCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.Schema != 1 || len(catalog.Suites) == 0 {
		t.Fatal("empty or unsupported conformance inventory")
	}
	pin, err := os.ReadFile(filepath.Join(root, "UPSTREAM"))
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Upstream != strings.TrimSpace(string(pin)) {
		t.Fatal("conformance inventory does not match the pinned upstream")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Minute)
	defer cancel()
	native := exec.CommandContext(ctx, filepath.Join(build, "upstream", "upstream_tests"))
	native.Dir = root
	nativeOutput, err := native.CombinedOutput()
	if err != nil {
		t.Fatalf("unchanged C suite failed: %v\n%s", err, nativeOutput)
	}
	artifacts := t.TempDir()
	nativeLog := filepath.Join(artifacts, "upstream.log")
	if err := os.WriteFile(nativeLog, nativeOutput, 0o644); err != nil {
		t.Fatal(err)
	}
	check := exec.CommandContext(ctx, "cmake", "-DLOG="+nativeLog,
		"-DINVENTORY="+filepath.Join(root, "oracle", "cases.json"),
		"-DOUTPUT="+filepath.Join(artifacts, "native.json"), "-P", filepath.Join(root, "oracle", "check-upstream.cmake"))
	if output, err := check.CombinedOutput(); err != nil {
		t.Fatalf("native case collection failed: %v\n%s", err, output)
	}
	args := []string{"test", "-json", "-count=1", "-timeout=40m", "-tags=box2d_stress"}
	precision := "float"
	if box2d.IsDoublePrecision() {
		precision = "double"
		args[4] = "-tags=box2d_stress,box2d_double"
	}
	args = append(args, "./...")
	command := exec.CommandContext(ctx, "go", args...)
	command.Dir = root
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr := &strings.Builder{}
	command.Stderr = stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	passed := make(map[string]int)
	skipped := make(map[string]int)
	used := make(map[string]int)
	fixtureLine := regexp.MustCompile(`C reference fixture: ([^ ]+) \(([0-9]+) observations\)`)
	decoder := json.NewDecoder(bufio.NewReader(stdout))
	var failures []string
	for {
		var event struct {
			Action  string
			Package string
			Test    string
			Output  string
		}
		err := decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = command.Process.Kill()
			_ = command.Wait()
			t.Fatalf("malformed Go collection: %v", err)
		}
		if event.Package == "github.com/zigai/box2d" && event.Test != "" {
			switch event.Action {
			case "pass":
				passed[event.Test]++
			case "skip":
				skipped[event.Test]++
			}
		}
		if event.Action == "fail" {
			failures = append(failures, event.Package+"/"+event.Test)
		}
		if match := fixtureLine.FindStringSubmatch(event.Output); match != nil {
			used[match[1]]++
		}
		if event.Action == "output" && (strings.Contains(event.Output, "FAIL") || strings.Contains(event.Output, "oracle:")) {
			t.Log(strings.TrimSpace(event.Output))
		}
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("Go corpus failed: %v; failed cases %v\n%s", err, failures, stderr)
	}
	require := func(name string) {
		t.Helper()
		if passed[name] != 1 || skipped[name] != 0 {
			t.Errorf("required Go case %s: passed=%d, skipped=%d", name, passed[name], skipped[name])
		}
	}
	active := 0
	for name, suite := range catalog.Suites {
		if suite.Go == "" {
			if suite.Reason == "" || len(suite.Cases) != 0 {
				t.Errorf("upstream suite %s has no Go case and no reason", name)
			}
			active++
			continue
		}
		require(suite.Go)
		if suite.Fixture != "" && used[suite.Fixture] == 0 {
			t.Errorf("upstream suite fixture %s was never consumed", suite.Fixture)
		}
		if len(suite.Cases) == 0 {
			active++
		}
		for _, c := range suite.Cases {
			active++
			require(c.Go)
			if used[c.Fixture] == 0 {
				t.Errorf("upstream case fixture %s was never consumed", c.Fixture)
			}
		}
	}
	for _, family := range catalog.Families {
		for _, name := range family.Go {
			require(name)
		}
	}
	for _, diagnostic := range catalog.Diagnostics {
		require(diagnostic.Go)
		if used[diagnostic.Fixture] == 0 {
			t.Errorf("unregistered diagnostic fixture %s was never consumed", diagnostic.Fixture)
		}
	}
	for _, disabled := range catalog.Disabled {
		if skipped[disabled.Go] != 1 || passed[disabled.Go] != 0 {
			t.Errorf("upstream-disabled case %s is not separately reported", disabled.Go)
		}
	}
	manifestData, err := os.ReadFile(filepath.Join(root, "testdata", "oracle", precision, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Fixtures map[string]json.RawMessage `json:"fixtures"`
	}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatal(err)
	}
	for name := range manifest.Fixtures {
		if used[name] == 0 {
			t.Errorf("completed C fixture %s is orphaned: no Go consumer executed", name)
		}
	}
	for name := range used {
		if _, exists := manifest.Fixtures[name]; !exists {
			t.Errorf("Go consumed undeclared C fixture %s", name)
		}
	}
	t.Logf("%s: %d unchanged C suites / %d active cases; %d C fixtures consumed; %d disabled upstream cases separately reported", precision, len(catalog.Suites), active, len(used), len(catalog.Disabled))
}
