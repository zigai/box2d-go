package porttest_test

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"testing"

	"github.com/zigai/box2d/internal/porttest"
)

type fatalRecorder struct {
	testing.TB
	failed  bool
	message string
}

func (r *fatalRecorder) Helper() {}

func (r *fatalRecorder) Fatal(args ...any) {
	r.failed = true
	r.message = fmt.Sprint(args...)
	runtime.Goexit()
}

func (r *fatalRecorder) Fatalf(format string, args ...any) {
	r.failed = true
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func runEnsure(body func(t testing.TB)) *fatalRecorder {
	rec := &fatalRecorder{}
	var wg sync.WaitGroup
	wg.Go(func() { body(rec) })
	wg.Wait()
	return rec
}

func TestEnsure(t *testing.T) {
	if rec := runEnsure(func(t testing.TB) { porttest.Ensure(t, true) }); rec.failed {
		t.Errorf("Ensure(true) failed: %s", rec.message)
	}
	if rec := runEnsure(func(t testing.TB) { porttest.Ensure(t, false) }); !rec.failed {
		t.Error("Ensure(false) did not fail")
	}
}

func TestEnsureSmall(t *testing.T) {
	nan := float32(math.NaN())
	tests := []struct {
		name     string
		v, tol   float32
		wantFail bool
	}{
		{"inside", 0.5, 1, false},
		{"at the bound", -1, 1, false},
		{"above", 1.5, 1, true},
		{"below", -1.5, 1, true},
		{"NaN passes as in C", nan, 1, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := runEnsure(func(tb testing.TB) { porttest.EnsureSmall(tb, tc.v, tc.tol) })
			if rec.failed != tc.wantFail {
				t.Errorf("EnsureSmall(%v, %v) failed = %v, want %v (%s)", tc.v, tc.tol, rec.failed, tc.wantFail, rec.message)
			}
		})
	}

	if rec := runEnsure(func(tb testing.TB) { porttest.EnsureSmall(tb, 1e-9, 1e-6) }); rec.failed {
		t.Errorf("EnsureSmall on float64 failed: %s", rec.message)
	}
}
