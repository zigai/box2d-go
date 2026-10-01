package box2d

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/zigai/box2d/internal/porttest"
)

func unportedStub() { todo() }

func TestTodoPanicsWithUnportedValue(t *testing.T) {
	defer func() {
		r := recover()
		u, ok := r.(interface{ Unported() string })
		if !ok {
			t.Fatalf("todo panicked with %T (%v), want a value with an Unported() string method", r, r)
		}
		if got := u.Unported(); got != "box2d.unportedStub" {
			t.Fatalf("Unported() = %q, want %q", got, "box2d.unportedStub")
		}
	}()
	unportedStub()
	t.Fatal("todo did not panic")
}

type skipRecorder struct {
	testing.TB
	message string
}

func (r *skipRecorder) Helper() {}

func (r *skipRecorder) Skipf(format string, args ...any) {
	r.message = fmt.Sprintf(format, args...)
	runtime.Goexit()
}

func TestTodoSkipsUnderGuard(t *testing.T) {
	rec := &skipRecorder{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer porttest.Guard(rec)
		unportedStub()
	}()
	<-done
	if want := "unported: box2d.unportedStub"; rec.message != want {
		t.Fatalf("skip message = %q, want %q", rec.message, want)
	}
}

func TestUnportedError(t *testing.T) {
	if got := (unported{fn: "box2d.X"}).Error(); !strings.Contains(got, "box2d.X") {
		t.Fatalf("Error() = %q does not name the function", got)
	}
}
