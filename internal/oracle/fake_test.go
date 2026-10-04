package oracle

import (
	"slices"
	"testing"
)

type fatalStop struct{}

type fakeTB struct {
	testing.TB

	fatal    string
	errors   []string
	failed   bool
	skipped  bool
	cleanups []func()
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.fatal = sprintf(format, args...)
	f.failed = true

	panic(fatalStop{})
}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.errors = append(f.errors, sprintf(format, args...))
	f.failed = true
}

func (f *fakeTB) Failed() bool  { return f.failed }
func (f *fakeTB) Skipped() bool { return f.skipped }

func (f *fakeTB) Cleanup(fn func()) { f.cleanups = append(f.cleanups, fn) }

func (f *fakeTB) run(body func()) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(fatalStop); !ok {
					panic(r)
				}
			}
		}()

		body()
	}()

	for _, v := range slices.Backward(f.cleanups) {
		v()
	}
}
