// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

type containerFoo struct {
	a int
	b float32
}
type containerBar struct{ a []int }

// These fixtures preserve upstream-internal helper cases. C allocation/free
// assertions run in C; only initialized contents and helper semantics are shared.
func observeContainerArray[T ~int | ~int16 | ~uint8](r *oracle.Reader, a []T) {
	r.I32("count", int32(len(a)))

	for _, value := range a {
		r.I32("element", int32(value))
	}
}

func observeContainerFooArray(r *oracle.Reader, a []containerFoo) {
	r.I32("count", int32(len(a)))

	for _, value := range a {
		r.I32("a", int32(value.a))
		r.F32("b", value.b)
	}
}

func TestContainer(t *testing.T) {
	t.Run("TestCreateDestroy", func(t *testing.T) {
		r := oracle.Open(t, "container/TestCreateDestroy")

		var a []int
		for i := range 2 {
			a = arrayReserve(a, 16)
			observeContainerArray(r, a)

			value := 42
			if i == 1 {
				value = 99
			}

			a = arrayPush(a, value)
			observeContainerArray(r, a)

			var popped int

			a, popped = arrayPop(a)
			r.I32("popped", int32(popped))
			observeContainerArray(r, a)
			a = nil
		}
	})
	t.Run("TestAccess", func(t *testing.T) {
		r := oracle.Open(t, "container/TestAccess")

		var a []int

		a = nil
		a = arrayPush(a, 42)

		element := &a[0]
		if *element != 42 {
			t.Fatal("condition false")
		}

		r.I32("access", int32(*element))
		observeContainerArray(r, a)
		a = nil
	})
	t.Run("TestIteration", func(t *testing.T) {
		r := oracle.Open(t, "container/TestIteration")

		var a []int

		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a = arrayPush(a, 3)

		sum := 0
		for i := range a {
			sum += a[i]
		}

		if sum != 6 {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum))
		observeContainerArray(r, a)
		a = nil
	})
	t.Run("TestArrayOfStruct", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayOfStruct")

		var a []containerFoo

		a = nil
		a = arrayPush(a, containerFoo{a: 1, b: 5.0})
		a = arrayPush(a, containerFoo{a: 2, b: 6.0})
		a = arrayPush(a, containerFoo{a: 3, b: 7.0})

		sum1 := 0
		sum2 := float32(0)

		for i := range a {
			sum1 += a[i].a
			sum2 += a[i].b
		}

		if sum1 != 6 {
			t.Fatal("condition false")
		}

		if !(sum2 == 18.0) {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum1))
		r.F32("sumFloat", sum2)
		observeContainerFooArray(r, a)

		a = nil
	})
	t.Run("TestStructWithArray", func(t *testing.T) {
		r := oracle.Open(t, "container/TestStructWithArray")

		var a containerBar

		a.a = nil
		a.a = arrayPush(a.a, 1)
		a.a = arrayPush(a.a, 2)
		a.a = arrayPush(a.a, 3)

		sum1 := 0
		for i := range len(a.a) {
			sum1 += a.a[i]
		}

		if sum1 != 6 {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum1))
		observeContainerArray(r, a.a)

		a.a = nil
	})
	t.Run("TestArrayEmplace", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayEmplace")

		var a []uint64

		for i := range 100 {
			var j *uint64

			a, j = arrayEmplace(a)
			*j = uint64(i)
		}

		sum := 0
		for i := range a {
			sum += int(a[i])
		}

		if sum != 100*99/2 {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum))
		r.I32("count", int32(len(a)))

		for _, value := range a {
			r.U64("element", value)
		}

		a = nil
	})
	t.Run("TestArrayRemove", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayRemove")

		var a []int16

		n := 10

		a = arrayReserve(a, n)
		if cap(a) != n || len(a) != 0 {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)

		for i := range n {
			a = arrayPush(a, int16(i))
		}

		if len(a) != n {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)

		sum := 0

		for range n {
			var temp int

			a, temp = arrayRemoveSwap(a, 0)
			sum += temp
			r.I32("moved", int32(temp))
			observeContainerArray(r, a)
		}

		if sum != n*(n-1)/2-1 {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum))

		a = nil
	})
	t.Run("TestArrayPop", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayPop")

		var a []uint8

		n := 100
		a = arrayReserve(a, n)

		a = a[:n]
		if cap(a) != n || len(a) != n {
			t.Fatal("condition false")
		}

		r.I32("resizedCount", int32(len(a)))

		for i := range n {
			a = arrayPush(a, uint8(i))
		}

		sum := 0

		for range n {
			var popped uint8

			a, popped = arrayPop(a)
			sum += int(popped)
			r.I32("popped", int32(popped))
			r.I32("count", int32(len(a)))
		}

		if sum != 100*99/2 {
			t.Fatal("condition false")
		}

		r.I32("sum", int32(sum))

		a = nil
	})
	t.Run("TestEmptyArrayProperties", func(t *testing.T) {
		r := oracle.Open(t, "container/TestEmptyArrayProperties")

		var a []int

		a = arrayReserve(a, 0)
		if len(a) != 0 {
			t.Fatal("condition false")
		}

		r.I32("count", int32(len(a)))

		a = nil
		if len(a) != 0 {
			t.Fatal("condition false")
		}

		if cap(a) != 0 {
			t.Fatal("condition false")
		}

		if a != nil {
			t.Fatal("condition false")
		}

		r.I32("count", int32(len(a)))
		a = arrayPush(a, 7)
		observeContainerArray(r, a)
		a, popped := arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)
	})
	t.Run("TestArrayReserveNoop", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayReserveNoop")

		var a []int

		a = arrayReserve(a, 16)
		if cap(a) < 16 {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)
		oldCapacity := cap(a)

		a = arrayReserve(a, 8)
		if cap(a) != oldCapacity {
			t.Fatal("condition false")
		}

		r.Bool("capacityUnchanged", cap(a) == oldCapacity)
		observeContainerArray(r, a)

		a = arrayReserve(a, oldCapacity)
		if cap(a) != oldCapacity {
			t.Fatal("condition false")
		}

		r.Bool("capacityUnchanged", cap(a) == oldCapacity)
		a = arrayPush(a, 42)
		observeContainerArray(r, a)
		a = nil
	})
	t.Run("TestArrayResizeDown", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayResizeDown")

		var a []int
		for i := range 10 {
			a = arrayPush(a, i*10)
		}

		if len(a) != 10 {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)

		a = arrayReserve(a, 5)

		a = a[:5]
		if len(a) != 5 {
			t.Fatal("condition false")
		}

		for i := range 5 {
			if a[i] != i*10 {
				t.Fatal("condition false")
			}
		}

		observeContainerArray(r, a)

		a = nil
	})
	t.Run("TestArrayResizeUp", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayResizeUp")

		var a []int

		a = arrayReserve(a, 10)

		a = a[:10]
		if len(a) != 10 {
			t.Fatal("condition false")
		}

		if cap(a) < 10 {
			t.Fatal("condition false")
		}

		r.I32("resizedCount", int32(len(a)))

		for i := range 10 {
			a[i] = i + 1
		}

		a = arrayReserve(a, 20)

		a = a[:20]
		if len(a) != 20 {
			t.Fatal("condition false")
		}

		for i := range 10 {
			if a[i] != i+1 {
				t.Fatal("condition false")
			}
		}

		r.I32("resizedCount", int32(len(a)))

		for _, value := range a[:10] {
			r.I32("element", int32(value))
		}

		a = nil
	})
	t.Run("TestArrayPopOrder", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayPopOrder")

		var popped int

		var a []int

		a = arrayPush(a, 10)
		a = arrayPush(a, 20)
		a = arrayPush(a, 30)

		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 30 {
			t.Fatal("condition false")
		}

		if len(a) != 2 {
			t.Fatal("condition false")
		}

		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 20 {
			t.Fatal("condition false")
		}

		if len(a) != 1 {
			t.Fatal("condition false")
		}

		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 10 {
			t.Fatal("condition false")
		}

		if len(a) != 0 {
			t.Fatal("condition false")
		}

		a = nil
	})
	t.Run("TestArrayRemoveSwapContents", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayRemoveSwapContents")

		var moved int

		var a []int

		a = arrayPush(a, 100)
		a = arrayPush(a, 200)
		a = arrayPush(a, 300)
		a = arrayPush(a, 400)

		a, moved = arrayRemoveSwap(a, 1)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 3 {
			t.Fatal("condition false")
		}

		if a[0] != 100 {
			t.Fatal("condition false")
		}

		if a[1] != 400 {
			t.Fatal("condition false")
		}

		if a[2] != 300 {
			t.Fatal("condition false")
		}

		a, moved = arrayRemoveSwap(a, 2)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 2 {
			t.Fatal("condition false")
		}

		if a[0] != 100 {
			t.Fatal("condition false")
		}

		if a[1] != 400 {
			t.Fatal("condition false")
		}

		a, moved = arrayRemoveSwap(a, 0)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 1 {
			t.Fatal("condition false")
		}

		if a[0] != 400 {
			t.Fatal("condition false")
		}

		a, moved = arrayRemoveSwap(a, 0)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 0 {
			t.Fatal("condition false")
		}

		a = nil
	})
	t.Run("TestArrayRemoveSwapAllFromEnd", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayRemoveSwapAllFromEnd")

		var moved int

		var a []int

		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a = arrayPush(a, 3)

		a, moved = arrayRemoveSwap(a, len(a)-1)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 2 {
			t.Fatal("condition false")
		}

		if a[0] != 1 {
			t.Fatal("condition false")
		}

		if a[1] != 2 {
			t.Fatal("condition false")
		}

		a, moved = arrayRemoveSwap(a, len(a)-1)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 1 {
			t.Fatal("condition false")
		}

		if a[0] != 1 {
			t.Fatal("condition false")
		}

		a, moved = arrayRemoveSwap(a, len(a)-1)
		r.I32("moved", int32(moved))
		observeContainerArray(r, a)

		if len(a) != 0 {
			t.Fatal("condition false")
		}

		a = nil
	})
	t.Run("TestArrayGrowthIntegrity", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayGrowthIntegrity")

		var a []int

		for i := range 1000 {
			a = arrayPush(a, i)
		}

		if len(a) != 1000 {
			t.Fatal("condition false")
		}

		if cap(a) < 1000 {
			t.Fatal("condition false")
		}

		for i := range 1000 {
			if a[i] != i {
				t.Fatal("condition false")
			}
		}

		observeContainerArray(r, a)

		a = nil
	})
	t.Run("TestArrayInterleavedPushPop", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayInterleavedPushPop")

		var popped int

		var a []int

		a = arrayPush(a, 1)
		a = arrayPush(a, 2)
		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 2 {
			t.Fatal("condition false")
		}

		if len(a) != 1 {
			t.Fatal("condition false")
		}

		a = arrayPush(a, 3)

		a = arrayPush(a, 4)
		if len(a) != 3 {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)
		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 4 {
			t.Fatal("condition false")
		}

		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 3 {
			t.Fatal("condition false")
		}

		a, popped = arrayPop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)

		if popped != 1 {
			t.Fatal("condition false")
		}

		if len(a) != 0 {
			t.Fatal("condition false")
		}

		a = arrayPush(a, 99)
		if len(a) != 1 {
			t.Fatal("condition false")
		}

		if a[0] != 99 {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)

		a = nil
	})
	t.Run("TestArrayEmplaceStruct", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayEmplaceStruct")

		var a []containerFoo

		for i := range 50 {
			var f *containerFoo

			a, f = arrayEmplace(a)
			f.a = i
			f.b = float32(i) * 2.0
		}

		if len(a) != 50 {
			t.Fatal("condition false")
		}

		for i := range 50 {
			f := &a[i]
			if f.a != i {
				t.Fatal("condition false")
			}

			if !(f.b == float32(i)*2.0) {
				t.Fatal("condition false")
			}
		}

		observeContainerFooArray(r, a)

		a = nil
	})
	t.Run("TestArrayPushAfterReserve", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayPushAfterReserve")

		var a []int

		a = arrayReserve(a, 50)
		if len(a) != 0 {
			t.Fatal("condition false")
		}

		if cap(a) < 50 {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)

		for i := range 50 {
			a = arrayPush(a, i*3)
		}

		if len(a) != 50 {
			t.Fatal("condition false")
		}

		for i := range 50 {
			if a[i] != i*3 {
				t.Fatal("condition false")
			}
		}

		observeContainerArray(r, a)

		a = nil
	})
	t.Run("TestArraySingleElement", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArraySingleElement")

		var a []containerFoo

		var f *containerFoo

		a, f = arrayEmplace(a)
		f.a = 7
		f.b = 3.14

		if len(a) != 1 {
			t.Fatal("condition false")
		}

		if a[0].a != 7 {
			t.Fatal("condition false")
		}

		if !(a[0].b == 3.14) {
			t.Fatal("condition false")
		}

		observeContainerFooArray(r, a)

		a, popped := arrayPop(a)
		if popped.a != 7 {
			t.Fatal("condition false")
		}

		if !(popped.b == 3.14) {
			t.Fatal("condition false")
		}

		if len(a) != 0 {
			t.Fatal("condition false")
		}

		r.I32("poppedA", int32(popped.a))
		r.F32("poppedB", popped.b)
		observeContainerFooArray(r, a)

		a = nil
	})
	t.Run("TestArrayCreateN", func(t *testing.T) {
		r := oracle.Open(t, "container/TestArrayCreateN")

		var a []int

		a = arrayReserve(a, 16)
		if len(a) != 0 {
			t.Fatal("condition false")
		}

		if cap(a) != 16 {
			t.Fatal("condition false")
		}

		if a == nil {
			t.Fatal("condition false")
		}

		observeContainerArray(r, a)

		for i := range 16 {
			a = arrayPush(a, i*5)
		}

		if len(a) != 16 {
			t.Fatal("condition false")
		}

		for i := range 16 {
			if a[i] != i*5 {
				t.Fatal("condition false")
			}
		}

		observeContainerArray(r, a)

		a = nil
	})
}
