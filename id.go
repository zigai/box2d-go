// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// Body references a body instance. This should be treated as an opaque handle. The zero value
// references no body.
type Body struct {
	index1     int32
	world      *World
	generation uint16
}

// Shape references a shape instance. This should be treated as an opaque handle. The zero value
// references no shape.
type Shape struct {
	index1     int32
	world      *World
	generation uint16
}

// Chain references a chain instance. This should be treated as an opaque handle. The zero value
// references no chain.
type Chain struct {
	index1     int32
	world      *World
	generation uint16
}

// Joint references a joint instance. This should be treated as an opaque handle. The zero value
// references no joint.
type Joint struct {
	index1     int32
	world      *World
	generation uint16
}

// Contact references a contact instance. This should be treated as an opaque handle. The zero
// value references no contact.
type Contact struct {
	index1     int32
	world      *World
	generation uint32
}

func packID(index1 int32, generation uint16) uint64 {
	return uint64(uint32(index1))<<32 | uint64(generation)
}

// ID returns a world-local identifier built from the body slot and generation. It is stable
// for the body's lifetime; a later body reusing the slot gets a different ID.
func (b Body) ID() uint64 { return packID(b.index1, b.generation) }

// ID returns a world-local identifier built from the shape slot and generation. It is stable
// for the shape's lifetime; a later shape reusing the slot gets a different ID.
func (s Shape) ID() uint64 { return packID(s.index1, s.generation) }

// ID returns a world-local identifier built from the chain slot and generation. It is stable
// for the chain's lifetime; a later chain reusing the slot gets a different ID.
func (c Chain) ID() uint64 { return packID(c.index1, c.generation) }

// ID returns a world-local identifier built from the joint slot and generation. It is stable
// for the joint's lifetime; a later joint reusing the slot gets a different ID.
func (j Joint) ID() uint64 { return packID(j.index1, j.generation) }

// ID returns a world-local identifier built from the contact slot and generation. It is stable
// for the contact's lifetime; a later contact reusing the slot gets a different ID.
func (c Contact) ID() uint64 { return uint64(uint32(c.index1))<<32 | uint64(c.generation) }

// Body resolves a world-local ID returned by Body.ID. Invalid slots, stale
// generations, nil worlds and closed worlds return the zero handle.
// IDs do not encode world ownership: use the world that produced the ID or
// restored its snapshot, not an unrelated world with coincident slots.
func (w *World) Body(id uint64) Body {
	if id&0xffff0000 != 0 {
		return Body{}
	}
	b := Body{index1: int32(id >> 32), generation: uint16(id), world: w}
	if !b.IsValid() {
		return Body{}
	}
	return b
}

// Shape resolves a world-local ID returned by Shape.ID, with the same
// validity and ownership contract as World.Body.
func (w *World) Shape(id uint64) Shape {
	if id&0xffff0000 != 0 {
		return Shape{}
	}
	s := Shape{index1: int32(id >> 32), generation: uint16(id), world: w}
	if !s.IsValid() {
		return Shape{}
	}
	return s
}
