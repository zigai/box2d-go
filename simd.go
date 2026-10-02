// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type aabbV = AABB

func loadAABBV(aabb *AABB) aabbV {
	return *aabb
}

func overlapNode(av aabbV, node *TreeNode) bool {
	bv := &node.AABB
	return av.LowerBound.X <= bv.UpperBound.X && av.LowerBound.Y <= bv.UpperBound.Y && bv.LowerBound.X <= av.UpperBound.X && bv.LowerBound.Y <= av.UpperBound.Y
}

func overlapV(a, b *AABB) bool {
	return a.LowerBound.X <= b.UpperBound.X && a.LowerBound.Y <= b.UpperBound.Y && b.LowerBound.X <= a.UpperBound.X && b.LowerBound.Y <= a.UpperBound.Y
}

func unionV(a, b AABB) AABB {
	return a.Union(b)
}

func unionPairV(pair []TreeNode) aabbV {
	return pair[0].AABB.Union(pair[1].AABB)
}

func storeAABBV(aabb *AABB, value aabbV, condition bool) {
	if condition {
		*aabb = value
	}
}
