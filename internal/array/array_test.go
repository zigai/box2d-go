// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package array

import (
	"testing"

	"github.com/zigai/box2d-go/internal/oracle"
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
	t.Run("TestCreateDestroy", runContainerTestCreateDestroy)
	t.Run("TestAccess", runContainerTestAccess)
	t.Run("TestIteration", runContainerTestIteration)
	t.Run("TestArrayOfStruct", runContainerTestArrayOfStruct)
	t.Run("TestStructWithArray", runContainerTestStructWithArray)
	t.Run("TestArrayEmplace", runContainerTestArrayEmplace)
	t.Run("TestArrayRemove", runContainerTestArrayRemove)
	t.Run("TestArrayPop", runContainerTestArrayPop)
	t.Run("TestEmptyArrayProperties", runContainerTestEmptyArrayProperties)
	t.Run("TestArrayReserveNoop", runContainerTestArrayReserveNoop)
	t.Run("TestArrayResizeDown", runContainerTestArrayResizeDown)
	t.Run("TestArrayResizeUp", runContainerTestArrayResizeUp)
	t.Run("TestArrayPopOrder", runContainerTestArrayPopOrder)
	t.Run("TestArrayRemoveSwapContents", runContainerTestArrayRemoveSwapContents)
	t.Run("TestArrayRemoveSwapAllFromEnd", runContainerTestArrayRemoveSwapAllFromEnd)
	t.Run("TestArrayGrowthIntegrity", runContainerTestArrayGrowthIntegrity)
	t.Run("TestArrayInterleavedPushPop", runContainerTestArrayInterleavedPushPop)
	t.Run("TestArrayEmplaceStruct", runContainerTestArrayEmplaceStruct)
	t.Run("TestArrayPushAfterReserve", runContainerTestArrayPushAfterReserve)
	t.Run("TestArraySingleElement", runContainerTestArraySingleElement)
	t.Run("TestArrayCreateN", runContainerTestArrayCreateN)
}

func runContainerTestCreateDestroy(t *testing.T) {
	r := oracle.Open(t, "container/TestCreateDestroy")

	var a []int
	for i := range 2 {
		a = Reserve(a, 16)
		observeContainerArray(r, a)

		value := 42
		if i == 1 {
			value = 99
		}

		a = Push(a, value)
		observeContainerArray(r, a)

		var popped int

		a, popped = Pop(a)
		r.I32("popped", int32(popped))
		observeContainerArray(r, a)
		a = nil
	}
}

func runContainerTestAccess(t *testing.T) {
	r := oracle.Open(t, "container/TestAccess")

	var a []int

	a = nil
	a = Push(a, 42)

	element := &a[0]
	if *element != 42 {
		t.Fatal("condition false")
	}

	r.I32("access", int32(*element))
	observeContainerArray(r, a)
}

func runContainerTestIteration(t *testing.T) {
	r := oracle.Open(t, "container/TestIteration")

	var a []int

	a = Push(a, 1)
	a = Push(a, 2)
	a = Push(a, 3)

	sum := 0
	for i := range a {
		sum += a[i]
	}

	if sum != 6 {
		t.Fatal("condition false")
	}

	r.I32("sum", int32(sum))
	observeContainerArray(r, a)
}

func runContainerTestArrayOfStruct(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayOfStruct")

	var a []containerFoo

	a = nil
	a = Push(a, containerFoo{a: 1, b: 5.0})
	a = Push(a, containerFoo{a: 2, b: 6.0})
	a = Push(a, containerFoo{a: 3, b: 7.0})

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
}

func runContainerTestStructWithArray(t *testing.T) {
	r := oracle.Open(t, "container/TestStructWithArray")

	var a containerBar

	a.a = nil
	a.a = Push(a.a, 1)
	a.a = Push(a.a, 2)
	a.a = Push(a.a, 3)

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
}

func runContainerTestArrayEmplace(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayEmplace")

	var a []uint64

	for i := range 100 {
		var j *uint64

		a, j = Emplace(a)
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
}

func runContainerTestArrayRemove(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayRemove")

	var a []int16

	n := 10

	a = Reserve(a, n)
	if cap(a) != n || len(a) != 0 {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)

	for i := range n {
		a = Push(a, int16(i))
	}

	if len(a) != n {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)

	sum := 0

	for range n {
		var temp int

		a, temp = RemoveSwap(a, 0)
		sum += temp
		r.I32("moved", int32(temp))
		observeContainerArray(r, a)
	}

	if sum != n*(n-1)/2-1 {
		t.Fatal("condition false")
	}

	r.I32("sum", int32(sum))
}

func runContainerTestArrayPop(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayPop")

	var a []uint8

	n := 100
	a = Reserve(a, n)

	a = a[:n]
	if cap(a) != n || len(a) != n {
		t.Fatal("condition false")
	}

	r.I32("resizedCount", int32(len(a)))

	for i := range n {
		a = Push(a, uint8(i))
	}

	sum := 0

	for range n {
		var popped uint8

		a, popped = Pop(a)
		sum += int(popped)
		r.I32("popped", int32(popped))
		r.I32("count", int32(len(a)))
	}

	if sum != 100*99/2 {
		t.Fatal("condition false")
	}

	r.I32("sum", int32(sum))
}

func runContainerTestEmptyArrayProperties(t *testing.T) {
	r := oracle.Open(t, "container/TestEmptyArrayProperties")

	var a []int

	a = Reserve(a, 0)
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

	if a != nil { //nolint:govet // reason: nilness preserves oracle/tests/upstream/test_container.c:292's deliberate destroyed-array nil check.
		t.Fatal("condition false")
	}

	r.I32("count", int32(len(a)))
	a = Push(a, 7)
	observeContainerArray(r, a)
	a, popped := Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)
}

func runContainerTestArrayReserveNoop(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayReserveNoop")

	var a []int

	a = Reserve(a, 16)
	if cap(a) < 16 {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)
	oldCapacity := cap(a)

	a = Reserve(a, 8)
	if cap(a) != oldCapacity {
		t.Fatal("condition false")
	}

	r.Bool("capacityUnchanged", cap(a) == oldCapacity)
	observeContainerArray(r, a)

	a = Reserve(a, oldCapacity)
	if cap(a) != oldCapacity {
		t.Fatal("condition false")
	}

	r.Bool("capacityUnchanged", cap(a) == oldCapacity)
	a = Push(a, 42)
	observeContainerArray(r, a)
}

func runContainerTestArrayResizeDown(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayResizeDown")

	var a []int
	for i := range 10 {
		a = Push(a, i*10)
	}

	if len(a) != 10 {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)

	a = Reserve(a, 5)

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
}

func runContainerTestArrayResizeUp(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayResizeUp")

	var a []int

	a = Reserve(a, 10)

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

	a = Reserve(a, 20)

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
}

func runContainerTestArrayPopOrder(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayPopOrder")

	var popped int

	var a []int

	a = Push(a, 10)
	a = Push(a, 20)
	a = Push(a, 30)

	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 30 {
		t.Fatal("condition false")
	}

	if len(a) != 2 {
		t.Fatal("condition false")
	}

	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 20 {
		t.Fatal("condition false")
	}

	if len(a) != 1 {
		t.Fatal("condition false")
	}

	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 10 {
		t.Fatal("condition false")
	}

	if len(a) != 0 {
		t.Fatal("condition false")
	}
}

func runContainerTestArrayRemoveSwapContents(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayRemoveSwapContents")

	var moved int

	var a []int

	a = Push(a, 100)
	a = Push(a, 200)
	a = Push(a, 300)
	a = Push(a, 400)

	a, moved = RemoveSwap(a, 1)
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

	a, moved = RemoveSwap(a, 2)
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

	a, moved = RemoveSwap(a, 0)
	r.I32("moved", int32(moved))
	observeContainerArray(r, a)

	if len(a) != 1 {
		t.Fatal("condition false")
	}

	if a[0] != 400 {
		t.Fatal("condition false")
	}

	a, moved = RemoveSwap(a, 0)
	r.I32("moved", int32(moved))
	observeContainerArray(r, a)

	if len(a) != 0 {
		t.Fatal("condition false")
	}
}

func runContainerTestArrayRemoveSwapAllFromEnd(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayRemoveSwapAllFromEnd")

	var moved int

	var a []int

	a = Push(a, 1)
	a = Push(a, 2)
	a = Push(a, 3)

	a, moved = RemoveSwap(a, len(a)-1)
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

	a, moved = RemoveSwap(a, len(a)-1)
	r.I32("moved", int32(moved))
	observeContainerArray(r, a)

	if len(a) != 1 {
		t.Fatal("condition false")
	}

	if a[0] != 1 {
		t.Fatal("condition false")
	}

	a, moved = RemoveSwap(a, len(a)-1)
	r.I32("moved", int32(moved))
	observeContainerArray(r, a)

	if len(a) != 0 {
		t.Fatal("condition false")
	}
}

func runContainerTestArrayGrowthIntegrity(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayGrowthIntegrity")

	var a []int

	for i := range 1000 {
		a = Push(a, i)
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
}

func runContainerTestArrayInterleavedPushPop(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayInterleavedPushPop")

	var popped int

	var a []int

	a = Push(a, 1)
	a = Push(a, 2)
	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 2 {
		t.Fatal("condition false")
	}

	if len(a) != 1 {
		t.Fatal("condition false")
	}

	a = Push(a, 3)

	a = Push(a, 4)
	if len(a) != 3 {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)
	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 4 {
		t.Fatal("condition false")
	}

	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 3 {
		t.Fatal("condition false")
	}

	a, popped = Pop(a)
	r.I32("popped", int32(popped))
	observeContainerArray(r, a)

	if popped != 1 {
		t.Fatal("condition false")
	}

	if len(a) != 0 {
		t.Fatal("condition false")
	}

	a = Push(a, 99)
	if len(a) != 1 {
		t.Fatal("condition false")
	}

	if a[0] != 99 {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)
}

func runContainerTestArrayEmplaceStruct(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayEmplaceStruct")

	var a []containerFoo

	for i := range 50 {
		var f *containerFoo

		a, f = Emplace(a)
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
}

func runContainerTestArrayPushAfterReserve(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayPushAfterReserve")

	var a []int

	a = Reserve(a, 50)
	if len(a) != 0 {
		t.Fatal("condition false")
	}

	if cap(a) < 50 {
		t.Fatal("condition false")
	}

	observeContainerArray(r, a)

	for i := range 50 {
		a = Push(a, i*3)
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
}

func runContainerTestArraySingleElement(t *testing.T) {
	r := oracle.Open(t, "container/TestArraySingleElement")

	var a []containerFoo

	var f *containerFoo

	a, f = Emplace(a)
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

	a, popped := Pop(a)
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
}

func runContainerTestArrayCreateN(t *testing.T) {
	r := oracle.Open(t, "container/TestArrayCreateN")

	var a []int

	a = Reserve(a, 16)
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
		a = Push(a, i*5)
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
}
