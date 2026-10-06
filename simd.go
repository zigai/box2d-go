// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

func unionV(a, b AABB) AABB {
	return a.Union(b)
}
