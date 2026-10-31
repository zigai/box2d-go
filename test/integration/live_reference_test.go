//go:build box2d_oracle

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zigai/box2d-go/box2d"
)

func TestLiveReferenceProfileRejectsInvalidSettings(t *testing.T) {
	const child = "BOX2D_TEST_REFERENCE_PROFILE_CHILD"
	if os.Getenv(child) == "1" {
		oracleBuildDir(t)
		return
	}
	precision := "float"
	if box2d.IsDoublePrecision() {
		precision = "double"
	}
	data, err := os.ReadFile(filepath.Join(oracleBuildDir(t), "reference.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, member string
		value        json.RawMessage
		reject       bool
	}{
		{name: "designated"},
		{name: "assertions-disabled", member: "assertions", value: json.RawMessage(`false`), reject: true},
		{name: "release", member: "build_type", value: json.RawMessage(`"Release"`), reject: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var manifest map[string]json.RawMessage
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Fatal(err)
			}
			if tc.member != "" {
				manifest[tc.member] = tc.value
			}
			image, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			base := t.TempDir()
			dir := filepath.Join(base, precision)
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "reference.json"), image, 0o644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLiveReferenceProfileRejectsInvalidSettings$")
			for _, value := range os.Environ() {
				if !strings.HasPrefix(value, child+"=") && !strings.HasPrefix(value, "BOX2D_ORACLE_DIR=") {
					cmd.Env = append(cmd.Env, value)
				}
			}
			cmd.Env = append(cmd.Env, child+"=1", "BOX2D_ORACLE_DIR="+base)
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("reference validation exceeded its bound: %v\n%s", ctx.Err(), output)
			}
			if err != nil {
				if _, ok := errors.AsType[*exec.ExitError](err); !ok {
					t.Fatalf("starting reference validation: %v", err)
				}
			}
			if (err != nil) != tc.reject {
				t.Fatalf("reference profile rejection=%t, want %t\n%s", err != nil, tc.reject, output)
			}
		})
	}
}
