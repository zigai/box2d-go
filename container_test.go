// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"testing"
)

type containerFoo struct {
	a int
	b float32
}
type containerBar struct{ a []int }

func TestContainer(t *testing.T) {

	t.Run("TestCreateDestroy", func(t *testing.T) {

		var a []int
		a = nil
		a = nil
		_ = a

	})
	t.Run("TestAccess", func(t *testing.T) {

		var a []int
		a = nil
		a = arrayPush(a, 42)
		element := &a[0]
		if !(*element == 42) {
			t.Fatal("condition false")
		}
		a = nil

	})
	t.Run("TestIteration", func(t *testing.T) {

		var a []int
		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a = arrayPush(a, 3)

		sum := 0
		for i := 0; i < len(a); i++ {
			sum += a[i]
		}

		if !(sum == 6) {
			t.Fatal("condition false")
		}
		a = nil

	})
	t.Run("TestArrayOfStruct", func(t *testing.T) {

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

		if !(sum1 == 6) {
			t.Fatal("condition false")
		}
		if !(sum2 == 18.0) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestStructWithArray", func(t *testing.T) {

		var a containerBar
		a.a = nil
		a.a = arrayPush(a.a, 1)
		a.a = arrayPush(a.a, 2)
		a.a = arrayPush(a.a, 3)

		sum1 := 0
		for i := 0; i < len(a.a); i++ {
			sum1 += a.a[i]
		}

		if !(sum1 == 6) {
			t.Fatal("condition false")
		}

		a.a = nil

	})
	t.Run("TestArrayEmplace", func(t *testing.T) {

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

		if !(sum == 100*99/2) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayRemove", func(t *testing.T) {

		var a []int16

		n := 10
		a = arrayReserve(a, n)
		if !(cap(a) == n && len(a) == 0) {
			t.Fatal("condition false")
		}

		for i := 0; i < n; i++ {
			a = arrayPush(a, int16(i))
		}

		if !(len(a) == n) {
			t.Fatal("condition false")
		}

		sum := 0
		for i := 0; i < n; i++ {
			var temp int
			a, temp = arrayRemoveSwap(a, 0)
			sum += temp
		}

		if !(sum == n*(n-1)/2-1) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayPop", func(t *testing.T) {

		var a []uint8

		n := 100
		a = arrayReserve(a, n)
		a = a[:n]
		if !(cap(a) == n && len(a) == n) {
			t.Fatal("condition false")
		}

		for i := 0; i < n; i++ {
			a = arrayPush(a, uint8(i))
		}

		sum := 0
		for i := 0; i < n; i++ {
			var popped uint8
			a, popped = arrayPop(a)
			sum += int(popped)
		}

		if !(sum == 100*99/2) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestEmptyArrayProperties", func(t *testing.T) {

		var a []int
		a = nil
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}
		a = nil
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}
		if !(cap(a) == 0) {
			t.Fatal("condition false")
		}
		if !(a == nil) {
			t.Fatal("condition false")
		}

	})
	t.Run("TestArrayReserveNoop", func(t *testing.T) {

		var a []int
		a = arrayReserve(a, 16)
		if !(cap(a) >= 16) {
			t.Fatal("condition false")
		}
		oldCapacity := cap(a)

		a = arrayReserve(a, 8)
		if !(cap(a) == oldCapacity) {
			t.Fatal("condition false")
		}
		a = arrayReserve(a, oldCapacity)
		if !(cap(a) == oldCapacity) {
			t.Fatal("condition false")
		}
		a = nil

	})
	t.Run("TestArrayResizeDown", func(t *testing.T) {

		var a []int
		for i := 0; i < 10; i++ {
			a = arrayPush(a, i*10)
		}
		if !(len(a) == 10) {
			t.Fatal("condition false")
		}

		a = arrayReserve(a, 5)
		a = a[:5]
		if !(len(a) == 5) {
			t.Fatal("condition false")
		}

		for i := 0; i < 5; i++ {
			if !(a[i] == i*10) {
				t.Fatal("condition false")
			}
		}

		a = nil

	})
	t.Run("TestArrayResizeUp", func(t *testing.T) {

		var a []int

		a = arrayReserve(a, 10)
		a = a[:10]
		if !(len(a) == 10) {
			t.Fatal("condition false")
		}
		if !(cap(a) >= 10) {
			t.Fatal("condition false")
		}

		for i := 0; i < 10; i++ {
			a[i] = i + 1
		}

		a = arrayReserve(a, 20)
		a = a[:20]
		if !(len(a) == 20) {
			t.Fatal("condition false")
		}
		for i := 0; i < 10; i++ {
			if !(a[i] == i+1) {
				t.Fatal("condition false")
			}
		}

		a = nil

	})
	t.Run("TestArrayPopOrder", func(t *testing.T) {

		var popped int

		var a []int
		a = arrayPush(a, 10)
		a = arrayPush(a, 20)
		a = arrayPush(a, 30)

		a, popped = arrayPop(a)
		if !(popped == 30) {
			t.Fatal("condition false")
		}
		if !(len(a) == 2) {
			t.Fatal("condition false")
		}
		a, popped = arrayPop(a)
		if !(popped == 20) {
			t.Fatal("condition false")
		}
		if !(len(a) == 1) {
			t.Fatal("condition false")
		}
		a, popped = arrayPop(a)
		if !(popped == 10) {
			t.Fatal("condition false")
		}
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayRemoveSwapContents", func(t *testing.T) {

		var a []int
		a = arrayPush(a, 100)
		a = arrayPush(a, 200)
		a = arrayPush(a, 300)
		a = arrayPush(a, 400)

		a, _ = arrayRemoveSwap(a, 1)
		if !(len(a) == 3) {
			t.Fatal("condition false")
		}
		if !(a[0] == 100) {
			t.Fatal("condition false")
		}
		if !(a[1] == 400) {
			t.Fatal("condition false")
		}
		if !(a[2] == 300) {
			t.Fatal("condition false")
		}

		a, _ = arrayRemoveSwap(a, 2)
		if !(len(a) == 2) {
			t.Fatal("condition false")
		}
		if !(a[0] == 100) {
			t.Fatal("condition false")
		}
		if !(a[1] == 400) {
			t.Fatal("condition false")
		}

		a, _ = arrayRemoveSwap(a, 0)
		if !(len(a) == 1) {
			t.Fatal("condition false")
		}
		if !(a[0] == 400) {
			t.Fatal("condition false")
		}

		a, _ = arrayRemoveSwap(a, 0)
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayRemoveSwapAllFromEnd", func(t *testing.T) {

		var a []int
		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a = arrayPush(a, 3)

		a, _ = arrayRemoveSwap(a, len(a)-1)
		if !(len(a) == 2) {
			t.Fatal("condition false")
		}
		if !(a[0] == 1) {
			t.Fatal("condition false")
		}
		if !(a[1] == 2) {
			t.Fatal("condition false")
		}

		a, _ = arrayRemoveSwap(a, len(a)-1)
		if !(len(a) == 1) {
			t.Fatal("condition false")
		}
		if !(a[0] == 1) {
			t.Fatal("condition false")
		}

		a, _ = arrayRemoveSwap(a, len(a)-1)
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayGrowthIntegrity", func(t *testing.T) {

		var a []int

		for i := 0; i < 1000; i++ {
			a = arrayPush(a, i)
		}

		if !(len(a) == 1000) {
			t.Fatal("condition false")
		}
		if !(cap(a) >= 1000) {
			t.Fatal("condition false")
		}

		for i := 0; i < 1000; i++ {
			if !(a[i] == i) {
				t.Fatal("condition false")
			}
		}

		a = nil

	})
	t.Run("TestArrayInterleavedPushPop", func(t *testing.T) {

		var popped int

		var a []int

		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a, popped = arrayPop(a)
		if !(popped == 2) {
			t.Fatal("condition false")
		}
		if !(len(a) == 1) {
			t.Fatal("condition false")
		}

		a = arrayPush(a, 3)
		a = arrayPush(a, 4)
		if !(len(a) == 3) {
			t.Fatal("condition false")
		}
		a, popped = arrayPop(a)
		if !(popped == 4) {
			t.Fatal("condition false")
		}
		a, popped = arrayPop(a)
		if !(popped == 3) {
			t.Fatal("condition false")
		}
		a, popped = arrayPop(a)
		if !(popped == 1) {
			t.Fatal("condition false")
		}
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}

		a = arrayPush(a, 99)
		if !(len(a) == 1) {
			t.Fatal("condition false")
		}
		if !(a[0] == 99) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayEmplaceStruct", func(t *testing.T) {

		var a []containerFoo

		for i := 0; i < 50; i++ {
			var f *containerFoo
			a, f = arrayEmplace(a)
			f.a = i
			f.b = float32(i) * 2.0
		}

		if !(len(a) == 50) {
			t.Fatal("condition false")
		}

		for i := 0; i < 50; i++ {
			f := &a[i]
			if !(f.a == i) {
				t.Fatal("condition false")
			}
			if !(f.b == float32(i)*2.0) {
				t.Fatal("condition false")
			}
		}

		a = nil

	})
	t.Run("TestArrayPushAfterReserve", func(t *testing.T) {

		var a []int

		a = arrayReserve(a, 50)
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}
		if !(cap(a) >= 50) {
			t.Fatal("condition false")
		}

		for i := 0; i < 50; i++ {
			a = arrayPush(a, i*3)
		}

		if !(len(a) == 50) {
			t.Fatal("condition false")
		}
		for i := 0; i < 50; i++ {
			if !(a[i] == i*3) {
				t.Fatal("condition false")
			}
		}

		a = nil

	})
	t.Run("TestArraySingleElement", func(t *testing.T) {

		var a []containerFoo

		var f *containerFoo
		a, f = arrayEmplace(a)
		f.a = 7
		f.b = 3.14

		if !(len(a) == 1) {
			t.Fatal("condition false")
		}
		if !(a[0].a == 7) {
			t.Fatal("condition false")
		}
		if !(a[0].b == 3.14) {
			t.Fatal("condition false")
		}

		a, popped := arrayPop(a)
		if !(popped.a == 7) {
			t.Fatal("condition false")
		}
		if !(popped.b == 3.14) {
			t.Fatal("condition false")
		}
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}

		a = nil

	})
	t.Run("TestArrayCreateN", func(t *testing.T) {

		var a []int
		a = make([]int, 0, 16)
		if !(len(a) == 0) {
			t.Fatal("condition false")
		}
		if !(cap(a) == 16) {
			t.Fatal("condition false")
		}
		if !(a != nil) {
			t.Fatal("condition false")
		}

		for i := 0; i < 16; i++ {
			a = arrayPush(a, i*5)
		}

		if !(len(a) == 16) {
			t.Fatal("condition false")
		}
		for i := 0; i < 16; i++ {
			if !(a[i] == i*5) {
				t.Fatal("condition false")
			}
		}

		a = nil

	})
}
