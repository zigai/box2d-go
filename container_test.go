// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"testing"

	"github.com/zigai/box2d/internal/porttest"
)

type containerFoo struct {
	a int
	b float32
}
type containerBar struct{ a []int }

func TestContainer(t *testing.T) {
	defer porttest.Guard(t)
	t.Run("TestCreateDestroy", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = nil
		a = nil
		_ = a

	})
	t.Run("TestAccess", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = nil
		a = arrayPush(a, 42)
		element := &a[0]
		porttest.Ensure(t, *element == 42)
		a = nil

	})
	t.Run("TestIteration", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a = arrayPush(a, 3)

		sum := 0
		for i := 0; i < len(a); i++ {
			sum += a[i]
		}

		porttest.Ensure(t, sum == 6)
		a = nil

	})
	t.Run("TestArrayOfStruct", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []containerFoo
		a = nil
		a = arrayPush(a, containerFoo{a: 1, b: 5.0})
		a = arrayPush(a, containerFoo{a: 2, b: 6.0})
		a = arrayPush(a, containerFoo{a: 3, b: 7.0})

		sum1 := 0
		sum2 := float32(0)
		for i := 0; i < len(a); i++ {
			sum1 += a[i].a
			sum2 += a[i].b
		}

		porttest.Ensure(t, sum1 == 6)
		porttest.Ensure(t, sum2 == 18.0)

		a = nil

	})
	t.Run("TestStructWithArray", func(t *testing.T) {
		defer porttest.Guard(t)

		var a containerBar
		a.a = nil
		a.a = arrayPush(a.a, 1)
		a.a = arrayPush(a.a, 2)
		a.a = arrayPush(a.a, 3)

		sum1 := 0
		for i := 0; i < len(a.a); i++ {
			sum1 += a.a[i]
		}

		porttest.Ensure(t, sum1 == 6)

		a.a = nil

	})
	t.Run("TestArrayEmplace", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []uint64

		for i := 0; i < 100; i++ {
			var j *uint64
			a, j = arrayEmplace(a)
			*j = uint64(i)
		}

		sum := 0
		for i := 0; i < len(a); i++ {
			sum += int(a[i])
		}

		porttest.Ensure(t, sum == 100*99/2)

		a = nil

	})
	t.Run("TestArrayRemove", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int16

		n := 10
		a = arrayReserve(a, n)
		porttest.Ensure(t, cap(a) == n && len(a) == 0)

		for i := 0; i < n; i++ {
			a = arrayPush(a, int16(i))
		}

		porttest.Ensure(t, len(a) == n)

		sum := 0
		for i := 0; i < n; i++ {
			var temp int
			a, temp = arrayRemoveSwap(a, 0)
			sum += temp
		}

		porttest.Ensure(t, sum == n*(n-1)/2-1)

		a = nil

	})
	t.Run("TestArrayPop", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []uint8

		n := 100
		a = arrayReserve(a, n)
		a = a[:n]
		porttest.Ensure(t, cap(a) == n && len(a) == n)

		for i := 0; i < n; i++ {
			a = arrayPush(a, uint8(i))
		}

		sum := 0
		for i := 0; i < n; i++ {
			var popped uint8
			a, popped = arrayPop(a)
			sum += int(popped)
		}

		porttest.Ensure(t, sum == 100*99/2)

		a = nil

	})
	t.Run("TestEmptyArrayProperties", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = nil
		porttest.Ensure(t, len(a) == 0)
		a = nil
		porttest.Ensure(t, len(a) == 0)
		porttest.Ensure(t, cap(a) == 0)
		porttest.Ensure(t, a == nil)

	})
	t.Run("TestArrayReserveNoop", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = arrayReserve(a, 16)
		porttest.Ensure(t, cap(a) >= 16)
		oldCapacity := cap(a)

		a = arrayReserve(a, 8)
		porttest.Ensure(t, cap(a) == oldCapacity)
		a = arrayReserve(a, oldCapacity)
		porttest.Ensure(t, cap(a) == oldCapacity)
		a = nil

	})
	t.Run("TestArrayResizeDown", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		for i := 0; i < 10; i++ {
			a = arrayPush(a, i*10)
		}
		porttest.Ensure(t, len(a) == 10)

		a = arrayReserve(a, 5)
		a = a[:5]
		porttest.Ensure(t, len(a) == 5)

		for i := 0; i < 5; i++ {
			porttest.Ensure(t, a[i] == i*10)
		}

		a = nil

	})
	t.Run("TestArrayResizeUp", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int

		a = arrayReserve(a, 10)
		a = a[:10]
		porttest.Ensure(t, len(a) == 10)
		porttest.Ensure(t, cap(a) >= 10)

		for i := 0; i < 10; i++ {
			a[i] = i + 1
		}

		a = arrayReserve(a, 20)
		a = a[:20]
		porttest.Ensure(t, len(a) == 20)
		for i := 0; i < 10; i++ {
			porttest.Ensure(t, a[i] == i+1)
		}

		a = nil

	})
	t.Run("TestArrayPopOrder", func(t *testing.T) {
		defer porttest.Guard(t)

		var popped int

		var a []int
		a = arrayPush(a, 10)
		a = arrayPush(a, 20)
		a = arrayPush(a, 30)

		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 30)
		porttest.Ensure(t, len(a) == 2)
		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 20)
		porttest.Ensure(t, len(a) == 1)
		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 10)
		porttest.Ensure(t, len(a) == 0)

		a = nil

	})
	t.Run("TestArrayRemoveSwapContents", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = arrayPush(a, 100)
		a = arrayPush(a, 200)
		a = arrayPush(a, 300)
		a = arrayPush(a, 400)

		a, _ = arrayRemoveSwap(a, 1)
		porttest.Ensure(t, len(a) == 3)
		porttest.Ensure(t, a[0] == 100)
		porttest.Ensure(t, a[1] == 400)
		porttest.Ensure(t, a[2] == 300)

		a, _ = arrayRemoveSwap(a, 2)
		porttest.Ensure(t, len(a) == 2)
		porttest.Ensure(t, a[0] == 100)
		porttest.Ensure(t, a[1] == 400)

		a, _ = arrayRemoveSwap(a, 0)
		porttest.Ensure(t, len(a) == 1)
		porttest.Ensure(t, a[0] == 400)

		a, _ = arrayRemoveSwap(a, 0)
		porttest.Ensure(t, len(a) == 0)

		a = nil

	})
	t.Run("TestArrayRemoveSwapAllFromEnd", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a = arrayPush(a, 3)

		a, _ = arrayRemoveSwap(a, len(a)-1)
		porttest.Ensure(t, len(a) == 2)
		porttest.Ensure(t, a[0] == 1)
		porttest.Ensure(t, a[1] == 2)

		a, _ = arrayRemoveSwap(a, len(a)-1)
		porttest.Ensure(t, len(a) == 1)
		porttest.Ensure(t, a[0] == 1)

		a, _ = arrayRemoveSwap(a, len(a)-1)
		porttest.Ensure(t, len(a) == 0)

		a = nil

	})
	t.Run("TestArrayGrowthIntegrity", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int

		for i := 0; i < 1000; i++ {
			a = arrayPush(a, i)
		}

		porttest.Ensure(t, len(a) == 1000)
		porttest.Ensure(t, cap(a) >= 1000)

		for i := 0; i < 1000; i++ {
			porttest.Ensure(t, a[i] == i)
		}

		a = nil

	})
	t.Run("TestArrayInterleavedPushPop", func(t *testing.T) {
		defer porttest.Guard(t)

		var popped int

		var a []int

		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 2)
		porttest.Ensure(t, len(a) == 1)

		a = arrayPush(a, 3)
		a = arrayPush(a, 4)
		porttest.Ensure(t, len(a) == 3)
		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 4)
		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 3)
		a, popped = arrayPop(a)
		porttest.Ensure(t, popped == 1)
		porttest.Ensure(t, len(a) == 0)

		a = arrayPush(a, 99)
		porttest.Ensure(t, len(a) == 1)
		porttest.Ensure(t, a[0] == 99)

		a = nil

	})
	t.Run("TestArrayEmplaceStruct", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []containerFoo

		for i := 0; i < 50; i++ {
			var f *containerFoo
			a, f = arrayEmplace(a)
			f.a = i
			f.b = float32(i) * 2.0
		}

		porttest.Ensure(t, len(a) == 50)

		for i := 0; i < 50; i++ {
			f := &a[i]
			porttest.Ensure(t, f.a == i)
			porttest.Ensure(t, f.b == float32(i)*2.0)
		}

		a = nil

	})
	t.Run("TestArrayPushAfterReserve", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int

		a = arrayReserve(a, 50)
		porttest.Ensure(t, len(a) == 0)
		porttest.Ensure(t, cap(a) >= 50)

		for i := 0; i < 50; i++ {
			a = arrayPush(a, i*3)
		}

		porttest.Ensure(t, len(a) == 50)
		for i := 0; i < 50; i++ {
			porttest.Ensure(t, a[i] == i*3)
		}

		a = nil

	})
	t.Run("TestArraySingleElement", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []containerFoo

		var f *containerFoo
		a, f = arrayEmplace(a)
		f.a = 7
		f.b = 3.14

		porttest.Ensure(t, len(a) == 1)
		porttest.Ensure(t, a[0].a == 7)
		porttest.Ensure(t, a[0].b == 3.14)

		a, popped := arrayPop(a)
		porttest.Ensure(t, popped.a == 7)
		porttest.Ensure(t, popped.b == 3.14)
		porttest.Ensure(t, len(a) == 0)

		a = nil

	})
	t.Run("TestArrayCreateN", func(t *testing.T) {
		defer porttest.Guard(t)

		var a []int
		a = make([]int, 0, 16)
		porttest.Ensure(t, len(a) == 0)
		porttest.Ensure(t, cap(a) == 16)
		porttest.Ensure(t, a != nil)

		for i := 0; i < 16; i++ {
			a = arrayPush(a, i*5)
		}

		porttest.Ensure(t, len(a) == 16)
		for i := 0; i < 16; i++ {
			porttest.Ensure(t, a[i] == i*5)
		}

		a = nil

	})
}
