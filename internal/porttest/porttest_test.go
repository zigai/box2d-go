package porttest_test

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/zigai/box2d/internal/porttest"
)

type notPorted struct{ name string }

func (n notPorted) Unported() string { return n.name }

type recorder struct {
	testing.TB
	skipped bool
	message string
}

func (r *recorder) Helper() {}

func (r *recorder) Skipf(format string, args ...any) {
	r.skipped = true
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func runGuarded(body func()) (rec *recorder, escaped any) {
	rec = &recorder{}
	var wg sync.WaitGroup
	wg.Go(func() {
		defer func() { escaped = recover() }()
		defer porttest.Guard(rec)
		body()
	})
	wg.Wait()
	return rec, escaped
}

func TestGuardSkipsUnported(t *testing.T) {
	rec, escaped := runGuarded(func() { panic(notPorted{"box2d.World.Step"}) })
	if escaped != nil {
		t.Fatalf("Guard let an unported panic escape: %v", escaped)
	}
	if !rec.skipped {
		t.Fatal("Guard did not skip")
	}
	if !strings.Contains(rec.message, "unported") || !strings.Contains(rec.message, "box2d.World.Step") {
		t.Fatalf("skip message %q does not name the unported function", rec.message)
	}
}

func TestGuardReraisesOtherPanics(t *testing.T) {
	for name, value := range map[string]any{
		"string": "boom",
		"error":  os.ErrNotExist,
		"int":    42,
	} {
		t.Run(name, func(t *testing.T) {
			rec, escaped := runGuarded(func() { panic(value) })
			if rec.skipped {
				t.Fatal("Guard skipped on a panic that is not unported")
			}
			if escaped != value {
				t.Fatalf("escaped panic = %v, want %v", escaped, value)
			}
		})
	}
}

func TestGuardPassesWithoutPanic(t *testing.T) {
	rec, escaped := runGuarded(func() {})
	if rec.skipped || escaped != nil {
		t.Fatalf("Guard interfered with a test that did not panic: skipped=%v escaped=%v", rec.skipped, escaped)
	}
}

const guardChildEnv = "BOX2D_PORTTEST_CHILD"

func TestGuardChild(t *testing.T) {
	kind := os.Getenv(guardChildEnv)
	if kind == "" {
		t.Skip("only runs as a subprocess of TestGuardInSubprocess")
	}
	defer porttest.Guard(t)
	switch kind {
	case "unported":
		panic(notPorted{"box2d.Body.SetLinearVelocity"})
	case "other":
		panic("a real bug")
	}
}

func TestGuardInSubprocess(t *testing.T) {
	if runtime.GOOS == "js" || runtime.GOOS == "wasip1" {
		t.Skip("cannot start subprocesses on " + runtime.GOOS)
	}
	run := func(kind string) (string, error) {
		cmd := exec.Command(os.Args[0], "-test.run=^TestGuardChild$", "-test.v")
		cmd.Env = append(os.Environ(), guardChildEnv+"="+kind)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	if _, err := run("none"); err != nil {
		if _, exited := err.(*exec.ExitError); !exited {
			t.Skipf("cannot re-execute the test binary: %v", err)
		}
	}

	t.Run("unported skips", func(t *testing.T) {
		out, err := run("unported")
		if err != nil {
			t.Fatalf("child failed, want skip: %v\n%s", err, out)
		}
		if !strings.Contains(out, "--- SKIP: TestGuardChild") || !strings.Contains(out, "unported: box2d.Body.SetLinearVelocity") {
			t.Fatalf("child output does not show an unported skip:\n%s", out)
		}
	})

	t.Run("other panic fails", func(t *testing.T) {
		out, err := run("other")
		if err == nil {
			t.Fatalf("child passed, want failure:\n%s", out)
		}
		if !strings.Contains(out, "a real bug") || strings.Contains(out, "--- SKIP: TestGuardChild") {
			t.Fatalf("child output does not show the re-raised panic:\n%s", out)
		}
	})
}
