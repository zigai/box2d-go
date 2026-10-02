// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"os"
	"runtime"
)

const (
	nullIndex int = -1
	// Simple djb2 hash function for determinism testing
	HashInit uint32 = 5381
	// Significant API and behavior changes.
	VersionMajor int = 3
	// API changes and features.
	VersionMinor int = 2
	// Bug fixes. No API changes.
	VersionRevision int = 0
)

// Prototype for the user assert callback. Return false to skip the panic.
type AssertFunc func(condition, fileName string, lineNumber int) bool

// Prototype for user log callback. Used to log warnings.
type LogFunc func(message string)

// Version numbering scheme.
type Version struct {
	Major    int
	Minor    int
	Revision int
}

type AssertError struct {
	Condition string
	File      string
	Line      int
}

var (
	assertHandler AssertFunc = defaultAssertFcn
	logHandler    LogFunc    = defaultLogFcn
)

func defaultAssertFcn(condition string, fileName string, lineNumber int) bool {
	fmt.Fprintf(os.Stderr, "BOX2D ASSERTION: %s, %s, line %d\n", condition, fileName, lineNumber)
	return true
}

// Override the default assert function
func SetAssertFunc(assertFcn AssertFunc) {
	assert(assertFcn != nil, "assertFcn != NULL")
	assertHandler = assertFcn
}

func assertFail(condition string, fileName string, lineNumber int) bool {
	result := assertHandler(condition, fileName, lineNumber)
	if result {
		panic(&AssertError{condition, fileName, lineNumber})
	}
	return result
}

func defaultLogFcn(message string) {
	fmt.Printf("Box2D: %s\n", message)
}

// Override the default log function
func SetLogFunc(logFcn LogFunc) {
	assert(logFcn != nil, "logFcn != NULL")
	logHandler = logFcn
}

func log(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if len(message) > 511 {
		message = message[:511]
	}
	logHandler(message)
}

// Get the current version of Box2D
func GetVersion() Version {
	return Version{Major: VersionMajor, Minor: VersionMinor, Revision: VersionRevision}
}

// Return true if the library was built with box2d_double (large world mode).
func IsDoublePrecision() bool {
	return doublePrecision
}

// Simple djb2 hash function for determinism testing
func Hash(hash uint32, data []byte) uint32 {
	for _, b := range data {
		hash = ((hash << 5) + hash) + uint32(b)
	}
	return hash
}

func assert(ok bool, condition string) {
	if assertsEnabled && !ok {
		_, file, line, _ := runtime.Caller(1)
		assertFail(condition, file, line)
	}
}

func validate(ok bool, condition string) {
	if validationEnabled && assertsEnabled && !ok {
		_, file, line, _ := runtime.Caller(1)
		assertFail(condition, file, line)
	}
}

func isDenormalFlushEnabled() bool {
	return false
}

func (e *AssertError) Error() string {
	return fmt.Sprintf("BOX2D ASSERTION: %s, %s, line %d", e.Condition, e.File, e.Line)
}
