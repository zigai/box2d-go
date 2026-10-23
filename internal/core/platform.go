// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package core

import "math/bits"

func BoundingPowerOf2(x int) int {
	if x <= 1 {
		return 1
	}

	return 32 - bits.LeadingZeros32(uint32(x)-1)
}

func RoundUpPowerOf2(x int) int {
	if x <= 1 {
		return 1
	}

	return 1 << (32 - bits.LeadingZeros32(uint32(x)-1))
}
