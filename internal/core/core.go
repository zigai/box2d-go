// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package core

import (
	"fmt"
	"runtime"
)

// AssertError is the panic value for invalid API input and, unless built with the
// box2d_noassert tag, for broken internal invariants.
type AssertError struct {
	Condition string
	File      string
	Line      int
}

func (e *AssertError) Error() string {
	return fmt.Sprintf("BOX2D ASSERTION: %s, %s, line %d", e.Condition, e.File, e.Line)
}

func Assert(ok bool, condition string) {
	if AssertsEnabled && !ok {
		_, file, line, _ := runtime.Caller(1)
		panic(&AssertError{Condition: condition, File: file, Line: line})
	}
}

func Validate(ok bool, condition string) {
	if ValidationEnabled && AssertsEnabled && !ok {
		_, file, line, _ := runtime.Caller(1)
		panic(&AssertError{Condition: condition, File: file, Line: line})
	}
}
