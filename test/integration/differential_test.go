//go:build box2d_oracle

package integration_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/zigai/box2d-go/internal/oracle"
	"github.com/zigai/box2d-go/internal/oracle/sequence"
)

// Both corpus replay and fuzzing execute a new native process for every input.
// Failure output contains the effective bytes and digest; testing.F retains and
// minimizes the original input in its native testdata/fuzz corpus automatically.
func runDifferential(t *testing.T, raw []byte) {
	t.Helper()
	input := raw[:min(len(raw), sequence.InputLimit)]
	sum := sha256.Sum256(input)
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		t.Logf("live differential input sha256=%x effective=%x original_bytes=%d; replay with go test -tags box2d_oracle ./test/integration -run 'FuzzDifferentialSequence/<saved-corpus-id>'", sum, input, len(raw))
		if dir := os.Getenv("BOX2D_DIFFERENTIAL_FAILURE_DIR"); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Errorf("retaining differential failure: %v", err)
				return
			}
			data := "go test fuzz v1\n[]byte(" + strconv.Quote(string(input)) + ")\n"
			path := filepath.Join(dir, fmt.Sprintf("%x", sum))
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Errorf("retaining differential failure %s: %v", path, err)
			}
		}
	})
	binary := filepath.Join(oracleBuildDir(t), "tools", "differential")
	if info, err := os.Stat(binary); err != nil || info.IsDir() {
		t.Fatalf("explicit box2d_oracle mode requires the pinned C runner %s: %v; build with just oracle", binary, err)
	}
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Dir = repoRoot(t)
	cmd.Stdin = bytes.NewReader(input)
	// Replace inherited output routing rather than creating duplicate env keys.
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "ORACLE_OUT=") {
			cmd.Env = append(cmd.Env, e)
		}
	}
	cmd.Env = append(cmd.Env, "ORACLE_OUT="+dir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fresh pinned C execution sha256=%x input=%x: %v (context=%v)\n%s", sum, input, err, ctx.Err(), output)
	}
	path := filepath.Join(dir, "differential", "sequence.jsonl")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("missing fresh C observations: %v\n%s", err, output)
	}
	if info.Size() > 32<<20 {
		t.Fatalf("bounded C sequence produced excessive output: %d bytes", info.Size())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading fresh C observations: %v\n%s", err, output)
	}
	r := oracle.Read(t, fmt.Sprintf("live-C/%x/differential/sequence", sum), data)
	sequence.Execute(t, r, input)
}

func TestDifferentialCorpus(t *testing.T) {
	for _, seed := range sequence.Corpus() {
		t.Run(seed.Name, func(t *testing.T) { runDifferential(t, seed.Input) })
	}
}

func FuzzDifferentialSequence(f *testing.F) {
	for _, seed := range sequence.Corpus() {
		f.Add(seed.Input)
	}
	f.Fuzz(func(t *testing.T, input []byte) { runDifferential(t, input) })
}
