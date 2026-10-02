// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// Get the absolute number of system ticks. The value is platform specific.
func GetTicks() uint64 {
	todo()
	var result uint64
	return result
}

// Get the milliseconds passed from an initial tick value.
func GetMilliseconds(ticks uint64) float32 {
	todo()
	var result float32
	return result
}

// Get the milliseconds passed from an initial tick value. Resets the passed in
// value to the current tick value.
func GetMillisecondsAndReset(ticks *uint64) float32 {
	todo()
	var result float32
	return result
}

// Yield to be used in a busy loop.
func Yield() {
	todo()
}
