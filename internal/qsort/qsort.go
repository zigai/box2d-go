/*
 * Copyright (c) 2013, 2017 Alexey Tourbin
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package qsort

const qThresh int = 16

func Sort[T any](s []T, less func(a, b T) bool) {
	if len(s) <= 1 {
		return
	}

	l, r, sp := 0, len(s)-1, 0

	var stack [48]struct{ l, r int }
	for {
		if r-l+1 < qThresh {
			insertionSortRange(s, less, l, r)

			if sp == 0 {
				break
			}

			sp--
			l, r = stack[sp].l, stack[sp].r

			continue
		}

		j := partitionQuickSort(s, less, l, r)
		i := j + 1
		j--

		l1, r1, l2, r2 := l, j, i, r
		if j-l < r-i {
			l1, r1, l2, r2 = i, r, l, j
		}

		if l2 == r2 {
			l, r = l1, r1
			continue
		}

		stack[sp].l = l1
		stack[sp].r = r1
		sp++
		l, r = l2, r2
	}
}

func orderQuickSortPivot[T any](s []T, less func(a, b T) bool, l int, m int, r int) {
	a1, a2, a3 := l+1, m, r
	if less(s[a2], s[a1]) {
		if less(s[a3], s[a2]) {
			s[a1], s[a3] = s[a3], s[a1]
			return
		}

		s[a1], s[a2] = s[a2], s[a1]
		if less(s[a3], s[a2]) {
			s[a2], s[a3] = s[a3], s[a2]
		}

		return
	}

	if less(s[a3], s[a2]) {
		s[a2], s[a3] = s[a3], s[a2]
		if less(s[a2], s[a1]) {
			s[a1], s[a2] = s[a2], s[a1]
		}
	}
}

func partitionQuickSort[T any](s []T, less func(a, b T) bool, l int, r int) int {
	m := l + ((r - l) >> 1)
	orderQuickSortPivot(s, less, l, m, r)
	s[l], s[m] = s[m], s[l]

	i, j := l+1, r
	for {
		i++
		for less(s[i], s[l]) {
			i++
		}

		j--
		for less(s[l], s[j]) {
			j--
		}

		if i >= j {
			break
		}

		s[i], s[j] = s[j], s[i]
	}

	s[l], s[j] = s[j], s[l]

	return j
}

func insertionSortRange[T any](s []T, less func(a, b T) bool, l int, r int) {
	for i := l + 1; i <= r; i++ {
		for j := i; j > l && less(s[j], s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
