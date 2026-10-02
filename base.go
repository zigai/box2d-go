// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

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
	todo()
	var result bool
	return result
}

// Override the default assert function
func SetAssertFunc(assertFcn AssertFunc) {
	todo()
}

func assertFail(condition string, fileName string, lineNumber int) bool {
	todo()
	var result bool
	return result
}

func defaultLogFcn(message string) {
	todo()
}

// Override the default log function
func SetLogFunc(logFcn LogFunc) {
	todo()
}

func log(format string, args ...any) {
	todo()
}

// Get the current version of Box2D
func GetVersion() Version {
	todo()
	var result Version
	return result
}

// Return true if the library was built with box2d_double (large world mode).
func IsDoublePrecision() bool {
	todo()
	var result bool
	return result
}

// Simple djb2 hash function for determinism testing
func Hash(hash uint32, data []byte) uint32 {
	todo()
	var result uint32
	return result
}

func assert(ok bool, condition string) {
	todo()
}

func validate(ok bool, condition string) {
	todo()
}

func isDenormalFlushEnabled() bool {
	todo()
	var result bool
	return result
}

func (e *AssertError) Error() string {
	todo()
	var result string
	return result
}
