package porttest

import "testing"

func Ensure(t testing.TB, cond bool) {
	t.Helper()
	if !cond {
		t.Fatal("condition false")
	}
}

func EnsureSmall[F ~float32 | ~float64](t testing.TB, v, tol F) {
	t.Helper()
	if v < -tol || tol < v {
		t.Fatalf("condition false: abs(%v) < %v", v, tol)
	}
}
