//go:build goexperiment.simd && amd64

package box2d

import (
	"os"
	"simd/archsimd"
	"unsafe"
)

var (
	simdSupported = archsimd.X86.AVX()
	simdEnabled   = simdSupported && os.Getenv("BOX2D_SIMD") != "0"
)

type aabbV [4]float32

func loadAABBV(aabb *AABB) aabbV {
	return *(*aabbV)(unsafe.Pointer(aabb))
}

func overlapNode(av aabbV, node *TreeNode) bool {
	if !simdEnabled {
		bounds := &node.AABB
		return av[0] <= bounds.UpperBound.X && av[1] <= bounds.UpperBound.Y && bounds.LowerBound.X <= av[2] && bounds.LowerBound.Y <= av[3]
	}

	leftVector := archsimd.LoadFloat32x4Array((*[4]float32)(&av))
	rightVector := archsimd.LoadFloat32x4Array((*[4]float32)(unsafe.Pointer(&node.AABB)))
	lower := leftVector.ConcatPermuteScalars(0, 1, 4, 5, rightVector)

	return lower.LessEqual(rightVector.ConcatPermuteScalars(2, 3, 6, 7, leftVector)).ToBits() == 15
}

func overlapV(left, right *AABB) bool {
	if !simdEnabled {
		return left.LowerBound.X <= right.UpperBound.X && left.LowerBound.Y <= right.UpperBound.Y && right.LowerBound.X <= left.UpperBound.X && right.LowerBound.Y <= left.UpperBound.Y
	}

	leftVector := archsimd.LoadFloat32x4Array((*[4]float32)(unsafe.Pointer(left)))
	rightVector := archsimd.LoadFloat32x4Array((*[4]float32)(unsafe.Pointer(right)))
	lower := leftVector.ConcatPermuteScalars(0, 1, 4, 5, rightVector)

	return lower.LessEqual(rightVector.ConcatPermuteScalars(2, 3, 6, 7, leftVector)).ToBits() == 15
}

func unionPairV(pair []TreeNode) aabbV {
	bounds := pair[0].AABB.Union(pair[1].AABB)
	return loadAABBV(&bounds)
}

func storeAABBV(aabb *AABB, value aabbV, condition bool) {
	if condition {
		*(*aabbV)(unsafe.Pointer(aabb)) = value
	}
}
