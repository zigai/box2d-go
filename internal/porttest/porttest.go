package porttest

import "testing"

type unported interface {
	Unported() string
}

func Guard(t testing.TB) {
	t.Helper()
	r := recover()
	if r == nil {
		return
	}
	if u, ok := r.(unported); ok {
		t.Skipf("unported: %s", u.Unported())
		return
	}
	panic(r)
}
