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
	if !checkInput(IsValidFloat(lengthUnits) && lengthUnits > 0, "IsValidFloat(lengthUnits) && lengthUnits > 0.0f") {
		return
	}
	lengthUnitsPerMeter = lengthUnits
}

func GetLengthUnitsPerMeter() float32 {
	return lengthUnitsPerMeter
}

func checkInput(ok bool, condition string) bool {
	if !ok {
		pc, file, line, _ := runtime.Caller(1)
		function := runtime.FuncForPC(pc).Name()
		log("invalid input: %s in %s\n", condition, function)
		if assertsEnabled {
			assertFail(condition, file, line)
		}
		return false
	}
	return true
}

func checkInputReturn(ok bool, condition string) bool {
	if !ok {
		pc, file, line, _ := runtime.Caller(1)
		function := runtime.FuncForPC(pc).Name()
		log("invalid input: %s in %s\n", condition, function)
		if assertsEnabled {
			assertFail(condition, file, line)
		}
		return false
	}
	return true
}
