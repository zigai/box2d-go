// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package array

import "github.com/zigai/box2d-go/internal/core"

func Push[T any](s []T, value T) []T {
	if len(s) >= cap(s) {
		newCapacity := 8
		if cap(s) != 0 {
			newCapacity = 2 * cap(s)
		}

		s = Reserve(s, newCapacity)
	}

	return append(s, value)
}

func Emplace[T any](s []T) ([]T, *T) {
	var value T

	s = Push(s, value)

	return s, &s[len(s)-1]
}

func Pop[T any](s []T) ([]T, T) {
	core.Assert(len(s) > 0, "0 < count")
	last := len(s) - 1
	value := s[last]

	var zero T

	s[last] = zero

	return s[:last], value
}

// RemoveSwap moves the last element into index and returns the old index of the
// moved element, or -1 when index was the last element.
func RemoveSwap[T any](s []T, index int) ([]T, int) {
	core.Assert(0 <= index && index < len(s), "0 <= index && index < count && Array index out of bounds")
	last := len(s) - 1
	moved := -1

	if index != last {
		s[index] = s[last]
		moved = last
	}

	var zero T

	s[last] = zero

	return s[:last], moved
}

func Reserve[T any](s []T, capacity int) []T {
	if cap(s) < capacity {
		grown := make([]T, len(s), capacity)
		copy(grown[:cap(s)], s[:cap(s)])
		s = grown
	}

	return s
}
