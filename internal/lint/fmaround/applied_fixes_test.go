package fmaround_test

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"

	"github.com/zigai/box2d-go/internal/lint/fmaround"
)

// Unlike the text goldens, these tests apply the analyzer's actual edits to
// separately authored programs and check the resulting executable behavior.
// These are native Go tooling tests, not physics/C-parity expectations.
func TestAppliedSuggestedFixes(t *testing.T) {
	basic := map[string]uint32{
		"add": 0, "right": 0, "subtract": 0, "right_subtract": 0,
		"unary_plus": 0, "unary_minus": 0, "negative_zero": 0x80000000,
		"grouping": 0, "compound_add": 0, "compound_subtract": 0,
		"field": 0, "index": 0, "both_products": 0, "alias": 0, "named": 0,
		"fused":            0xa8800000,
		"nonzero_grouping": 0x41300000, "nonzero_outer_minus": 0xbf800000,
		"nonzero_nested_sign": 0xbf800000,
	}
	cases := []struct {
		name        string
		diagnostics int
		want        map[string]uint32
	}{
		{"basic", 19, basic},
		{"imported", 1, map[string]uint32{"result": 0}},
		{"dot_import", 1, map[string]uint32{"result": 0}},
		{"shadow_float32", 1, map[string]uint32{"result": 0}},
		{"shadow_alias", 1, map[string]uint32{"result": 0}},
		{"shadow_named", 1, map[string]uint32{"result": 0}},
		{"shadow_import", 1, map[string]uint32{"result": 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			mainFile := filepath.Join(dir, "main.go")
			source, err := os.ReadFile(filepath.Join("testdata", "applied", tc.name, "main.go"))
			if err != nil {
				t.Fatal(err)
			}
			writeAppliedFile(t, mainFile, source)
			writeAppliedFile(t, filepath.Join(dir, "go.mod"), []byte("module roundingfixture\n\ngo 1.27\n"))
			if err := os.Mkdir(filepath.Join(dir, "measures"), 0o700); err != nil {
				t.Fatal(err)
			}
			measure, err := os.ReadFile(filepath.Join("testdata", "src", "measures", "measures.go"))
			if err != nil {
				t.Fatal(err)
			}
			writeAppliedFile(t, filepath.Join(dir, "measures", "measures.go"), measure)

			pkg := loadAppliedPackage(t, dir)
			diagnostics := analyzeAppliedPackage(t, pkg)
			if len(diagnostics) != tc.diagnostics {
				t.Fatalf("original diagnostics = %d, want %d; %+v", len(diagnostics), tc.diagnostics, diagnostics)
			}
			var edits []analysis.TextEdit
			for i, diagnostic := range diagnostics {
				if len(diagnostic.SuggestedFixes) != 1 {
					t.Fatalf("diagnostic %d has %d fixes, want one applicable rounding fix", i, len(diagnostic.SuggestedFixes))
				}
				fix := diagnostic.SuggestedFixes[0]
				if len(fix.TextEdits) == 0 {
					t.Fatalf("diagnostic %d offers an empty fix", i)
				}
				edits = append(edits, fix.TextEdits...)
			}
			sort.SliceStable(edits, func(i, j int) bool { return edits[i].Pos > edits[j].Pos })
			lastStart := len(source)
			for _, edit := range edits {
				file := pkg.Fset.File(edit.Pos)
				if file == nil || file.Name() != mainFile || pkg.Fset.File(edit.End) != file {
					t.Fatalf("fix edits an unexpected file: %+v", edit)
				}
				start, end := file.Offset(edit.Pos), file.Offset(edit.End)
				if start > end || end > lastStart {
					t.Fatalf("fix has reversed or overlapping ranges: [%d,%d), previous start %d", start, end, lastStart)
				}
				source = append(append(append([]byte(nil), source[:start]...), edit.NewText...), source[end:]...)
				lastStart = start
			}
			writeAppliedFile(t, mainFile, source)

			// Loading with NeedTypes/NeedTypesInfo rejects invalid conversions,
			// including local identifiers that shadow a suggested type spelling.
			fixed := loadAppliedPackage(t, dir)
			if remaining := analyzeAppliedPackage(t, fixed); len(remaining) != 0 {
				t.Fatalf("applied fixes leave %d diagnostics: %+v", len(remaining), remaining)
			}
			cmd := exec.CommandContext(t.Context(), "go", "run", ".")
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "GOWORK=off")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("compiling/executing actual fixes: %v\n%s", err, output)
			}
			seen := make(map[string]bool)
			scanner := bufio.NewScanner(strings.NewReader(string(output)))
			for scanner.Scan() {
				fields := strings.Fields(scanner.Text())
				if len(fields) != 2 {
					t.Fatalf("invalid observation %q", scanner.Text())
				}
				key := fields[0]
				want, ok := tc.want[key]
				if !ok || seen[key] {
					t.Fatalf("unexpected or duplicate observation %q", key)
				}
				seen[key] = true
				bits, err := strconv.ParseUint(fields[1], 16, 32)
				if err != nil || uint32(bits) != want {
					t.Errorf("%s = %s (%v), want %08x", key, fields[1], err, want)
				}
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			for key := range tc.want {
				if !seen[key] {
					t.Errorf("missing numeric observation %q", key)
				}
			}
		})
	}
}

func writeAppliedFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func loadAppliedPackage(t *testing.T, dir string) *packages.Package {
	t.Helper()
	loaded, err := packages.Load(&packages.Config{
		Mode: packages.LoadSyntax,
		Dir:  dir,
		Env:  append(os.Environ(), "GOWORK=off"),
	}, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded %d fixture packages, want one", len(loaded))
	}
	pkg := loaded[0]
	if len(pkg.Errors) != 0 {
		t.Fatalf("fixture does not type-check: %v", pkg.Errors)
	}
	return pkg
}

func analyzeAppliedPackage(t *testing.T, pkg *packages.Package) []analysis.Diagnostic {
	t.Helper()
	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Analyzer:   fmaround.Analyzer,
		Fset:       pkg.Fset,
		Files:      pkg.Syntax,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		ResultOf:   map[*analysis.Analyzer]any{inspect.Analyzer: inspector.New(pkg.Syntax)},
		Report:     func(diagnostic analysis.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
		ReadFile:   os.ReadFile,
	}
	if _, err := fmaround.Analyzer.Run(pass); err != nil {
		t.Fatal(fmt.Errorf("running analyzer: %w", err))
	}
	return diagnostics
}
