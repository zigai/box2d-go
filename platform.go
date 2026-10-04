// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math/bits"

func isPowerOf2(x int) bool {
	return x&(x-1) == 0
}

func boundingPowerOf2(x int) int {
	if x <= 1 {
		return 1
	}

	return 32 - bits.LeadingZeros32(uint32(x)-1)
}

func roundUpPowerOf2(x int) int {
	if x <= 1 {
		return 1
	}

	return 1 << (32 - bits.LeadingZeros32(uint32(x)-1))
}
