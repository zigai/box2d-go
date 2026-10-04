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

package box2d

const qThresh int = 16

func qsort[T any](s []T, less func(a, b T) bool) {
	if len(s) <= 1 {
		return
	}

	l, r, sp := 0, len(s)-1, 0

	var stack [48]struct{ l, r int }
	for {
		if r-l+1 >= qThresh {
			m := l + ((r - l) >> 1)

			a1, a2, a3 := l+1, m, r
			if less(s[a2], s[a1]) {
				if less(s[a3], s[a2]) {
					s[a1], s[a3] = s[a3], s[a1]
				} else {
					s[a1], s[a2] = s[a2], s[a1]
					if less(s[a3], s[a2]) {
						s[a2], s[a3] = s[a3], s[a2]
					}
				}
			} else if less(s[a3], s[a2]) {
				s[a2], s[a3] = s[a3], s[a2]
				if less(s[a2], s[a1]) {
					s[a1], s[a2] = s[a2], s[a1]
				}
			}

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

			i = j + 1
			s[l], s[j] = s[j], s[l]
			j--

			l1, r1, l2, r2 := l, j, i, r
			if j-l < r-i {
				l1, r1, l2, r2 = i, r, l, j
			}

			if l2 == r2 {
				l, r = l1, r1
			} else {
				stack[sp].l = l1
				stack[sp].r = r1
				sp++
				l, r = l2, r2
			}
		} else {
			for i := l + 1; i <= r; i++ {
				for j := i; j > l && less(s[j], s[j-1]); j-- {
					s[j], s[j-1] = s[j-1], s[j]
				}
			}

			if sp == 0 {
				break
			}

			sp--
			l, r = stack[sp].l, stack[sp].r
		}
	}
}
