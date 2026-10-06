//go:build !goexperiment.simd || !amd64

package box2d

const (
	simdSupported = false
	simdEnabled   = false
)

type aabbV = AABB

func loadAABBV(aabb *AABB) aabbV {
	return *aabb
}

func overlapNode(av aabbV, node *TreeNode) bool {
	bounds := &node.AABB
	return av.LowerBound.X <= bounds.UpperBound.X && av.LowerBound.Y <= bounds.UpperBound.Y && bounds.LowerBound.X <= av.UpperBound.X && bounds.LowerBound.Y <= av.UpperBound.Y
}

func overlapV(left, right *AABB) bool {
	return left.LowerBound.X <= right.UpperBound.X && left.LowerBound.Y <= right.UpperBound.Y && right.LowerBound.X <= left.UpperBound.X && right.LowerBound.Y <= left.UpperBound.Y
}

func unionPairV(pair []TreeNode) aabbV {
	return pair[0].AABB.Union(pair[1].AABB)
}

func storeAABBV(aabb *AABB, value aabbV, condition bool) {
	if condition {
		*aabb = value
	}
}
