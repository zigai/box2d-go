// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"runtime"
	"time"
)

var tickOrigin = time.Now()

// Get the absolute number of system ticks. The value is platform specific.
func currentTicks() uint64 {
	return uint64(time.Since(tickOrigin).Nanoseconds())
}

// Get the milliseconds passed from an initial tick value.
func milliseconds(ticks uint64) float32 {
	ticks2 := currentTicks()
	return float32(float64(ticks2-ticks) / 1000000)
}

// Get the milliseconds passed from an initial tick value. Resets the passed in
// value to the current tick value.
func millisecondsAndReset(ticks *uint64) float32 {
	ticks2 := currentTicks()
	milliseconds := float32(float64(ticks2-*ticks) / 1000000)
	*ticks = ticks2

	return milliseconds
}

// yieldThread to be used in a busy loop.
func yieldThread() {
	runtime.Gosched()
}
