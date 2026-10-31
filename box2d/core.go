// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "runtime"

const (
	secretCookie       int32 = 1152023
	simdWidth                = 4
	snoopTableCounters       = false
	snoopPairCounters        = false
	snoopTOICounters         = false
)

var lengthUnitsPerMeter float32 = 1

func SetLengthUnitsPerMeter(lengthUnits float32) {
	checkInput(
		IsValidFloat(lengthUnits) && lengthUnits > 0,
		"IsValidFloat(lengthUnits) && lengthUnits > 0.0f",
	)
	lengthUnitsPerMeter = lengthUnits
}

func LengthUnitsPerMeter() float32 {
	return lengthUnitsPerMeter
}

// checkInput panics when an API argument is invalid. Invalid input is a programmer error, so it
// is rejected regardless of the box2d_noassert build tag.
func checkInput(ok bool, condition string) {
	if !ok {
		_, file, line, _ := runtime.Caller(1)
		panic(&AssertError{Condition: condition, File: file, Line: line})
	}
}
