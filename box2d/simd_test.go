package box2d

import (
	"math"
	"os"
	"testing"
	"unsafe"
)

func TestAABBOverlapBitEquivalence(t *testing.T) {
	special := []uint32{
		0, 0x80000000, 1, 0x80000001, 0x007fffff, 0x807fffff,
		0x00800000, 0x80800000, 0x3f800000, 0xbf800000,
		0x7f7fffff, 0xff7fffff, 0x7f800000, 0xff800000,
		0x7f800001, 0xff800001, 0x7fbfffff, 0xffbfffff,
		0x7fc00000, 0xffc00000, 0x7fffffff, 0xffffffff,
	}

	for leftPosition := range 4 {
		for rightPosition := range 4 {
			for _, leftValue := range special {
				for _, rightValue := range special {
					left := [4]uint32{0xbf800000, 0xbf800000, 0x3f800000, 0x3f800000}
					right := [4]uint32{0, 0, 0x40000000, 0x40000000}
					left[leftPosition] = leftValue
					right[rightPosition] = rightValue
					checkAABBOverlap(t, left, right)
				}
			}
		}
	}

	bits := uint32(12345)

	nextBits := func() uint32 {
		bits = 1664525*bits + 1013904223
		return bits
	}
	for range 1_000_000 {
		left := [4]uint32{nextBits(), nextBits(), nextBits(), nextBits()}
		right := [4]uint32{nextBits(), nextBits(), nextBits(), nextBits()}
		checkAABBOverlap(t, left, right)
	}
}

func TestAABBVectorHelpers(t *testing.T) {
	bits := uint32(98765)

	nextBits := func() uint32 {
		bits = 1664525*bits + 1013904223
		return bits
	}
	for range 10_000 {
		left := overlapAABBFromBits([4]uint32{nextBits(), nextBits(), nextBits(), nextBits()})
		right := overlapAABBFromBits([4]uint32{nextBits(), nextBits(), nextBits(), nextBits()})
		original := overlapAABBBits(right)
		value := loadAABBV(&left)

		storeAABBV(&right, value, false)

		if actual := overlapAABBBits(right); actual != original {
			t.Fatalf("conditional store changed %08x to %08x", original, actual)
		}

		storeAABBV(&right, value, true)

		if actual, expected := overlapAABBBits(right), overlapAABBBits(left); actual != expected {
			t.Fatalf("load/store = %08x, want %08x", actual, expected)
		}

		pair := []TreeNode{{AABB: left}, {AABB: overlapAABBFromBits(original)}}
		expected := overlapAABBBits(pair[0].AABB.Union(pair[1].AABB))
		storeAABBV(&right, unionPairV(pair), true)

		if actual := overlapAABBBits(right); actual != expected {
			t.Fatalf("union/store = %08x, want %08x", actual, expected)
		}
	}
}

func TestAABBVectorLayout(t *testing.T) {
	if unsafe.Sizeof(AABB{}) != unsafe.Sizeof([4]float32{}) ||
		unsafe.Alignof(AABB{}) != unsafe.Alignof([4]float32{}) {
		t.Fatal("AABB must have the size and alignment of four float32 values")
	}

	if unsafe.Sizeof(Vec2{}) != 8 || unsafe.Offsetof(Vec2{}.X) != 0 ||
		unsafe.Offsetof(Vec2{}.Y) != 4 {
		t.Fatal("Vec2 must contain two consecutive float32 values")
	}

	if unsafe.Offsetof(AABB{}.LowerBound) != 0 || unsafe.Offsetof(AABB{}.UpperBound) != 8 {
		t.Fatal("AABB must contain two consecutive Vec2 values")
	}
}

func TestSIMDConfiguration(t *testing.T) {
	expected := simdSupported && os.Getenv("BOX2D_SIMD") != "0"
	if simdEnabled != expected {
		t.Fatalf("SIMD active = %t, want %t", simdEnabled, expected)
	}

	t.Logf("SIMD supported=%t active=%t", simdSupported, simdEnabled)
}

func checkAABBOverlap(t *testing.T, leftBits [4]uint32, rightBits [4]uint32) {
	t.Helper()

	left := overlapAABBFromBits(leftBits)
	right := overlapAABBFromBits(rightBits)
	expected := left.LowerBound.X <= right.UpperBound.X &&
		left.LowerBound.Y <= right.UpperBound.Y &&
		right.LowerBound.X <= left.UpperBound.X &&
		right.LowerBound.Y <= left.UpperBound.Y

	if actual := overlapV(&left, &right); actual != expected {
		t.Fatalf("overlapV(%08x, %08x) = %t, want %t", leftBits, rightBits, actual, expected)
	}

	if actual := overlapNode(loadAABBV(&left), &TreeNode{AABB: right}); actual != expected {
		t.Fatalf("overlapNode(%08x, %08x) = %t, want %t", leftBits, rightBits, actual, expected)
	}

	if overlapAABBBits(left) != leftBits || overlapAABBBits(right) != rightBits {
		t.Fatal("overlap changed input bits")
	}
}

func overlapAABBFromBits(bits [4]uint32) AABB {
	return AABB{
		LowerBound: Vec2{X: math.Float32frombits(bits[0]), Y: math.Float32frombits(bits[1])},
		UpperBound: Vec2{X: math.Float32frombits(bits[2]), Y: math.Float32frombits(bits[3])},
	}
}

func overlapAABBBits(value AABB) [4]uint32 {
	return [4]uint32{
		math.Float32bits(value.LowerBound.X), math.Float32bits(value.LowerBound.Y),
		math.Float32bits(value.UpperBound.X), math.Float32bits(value.UpperBound.Y),
	}
}
