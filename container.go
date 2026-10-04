// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

func arrayPush[T any](s []T, value T) []T {
	if len(s) >= cap(s) {
		newCapacity := 8
		if cap(s) != 0 {
			newCapacity = 2 * cap(s)
		}

		s = arrayReserve(s, newCapacity)
	}

	return append(s, value)
}

func arrayEmplace[T any](s []T) ([]T, *T) {
	var value T

	s = arrayPush(s, value)

	return s, &s[len(s)-1]
}

func arrayPop[T any](s []T) ([]T, T) {
	assert(len(s) > 0, "0 < count")
	last := len(s) - 1
	value := s[last]

	var zero T

	s[last] = zero

	return s[:last], value
}

func arrayRemoveSwap[T any](s []T, index int) ([]T, int) {
	assert(0 <= index && index < len(s), "0 <= index && index < count && Array index out of bounds")
	last := len(s) - 1
	moved := nullIndex

	if index != last {
		s[index] = s[last]
		moved = last
	}

	var zero T

	s[last] = zero

	return s[:last], moved
}

func arrayReserve[T any](s []T, capacity int) []T {
	if cap(s) < capacity {
		grown := make([]T, len(s), capacity)
		copy(grown[:cap(s)], s[:cap(s)])
		s = grown
	}

	return s
}
