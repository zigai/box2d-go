// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// World references a world instance. This should be treated as an opaque handle.
type World struct {
	index1     uint16
	generation uint16
}

// Body references a body instance. This should be treated as an opaque handle.
type Body struct {
	index1     int32
	world0     uint16
	generation uint16
}

// Shape references a shape instance. This should be treated as an opaque handle.
type Shape struct {
	index1     int32
	world0     uint16
	generation uint16
}

// Chain references a chain instance. This should be treated as an opaque handle.
type Chain struct {
	index1     int32
	world0     uint16
	generation uint16
}

// Joint references a joint instance. This should be treated as an opaque handle.
type Joint struct {
	index1     int32
	world0     uint16
	generation uint16
}

// Contact references a contact instance. This should be treated as an opaque handle.
type Contact struct {
	index1     int32
	world0     uint16
	padding    int16
	generation uint32
}

// Store a world id into a uint32.
func (r World) Store() uint32 {
	return uint32(r.index1)<<16 | uint32(r.generation)
}

// Load a uint32 into a world id.
func LoadWorld(x uint32) World {
	return World{uint16(x >> 16), uint16(x)}
}

// Store a body id into a uint64.
func (r Body) Store() uint64 {
	return uint64(r.index1)<<32 | uint64(r.world0)<<16 | uint64(r.generation)
}

// Load a uint64 into a body id.
func LoadBody(x uint64) Body {
	return Body{int32(x >> 32), uint16(x >> 16), uint16(x)}
}

// Store a shape id into a uint64.
func (r Shape) Store() uint64 {
	return uint64(r.index1)<<32 | uint64(r.world0)<<16 | uint64(r.generation)
}

// Load a uint64 into a shape id.
func LoadShape(x uint64) Shape {
	return Shape{int32(x >> 32), uint16(x >> 16), uint16(x)}
}

// Store a chain id into a uint64.
func (r Chain) Store() uint64 {
	return uint64(r.index1)<<32 | uint64(r.world0)<<16 | uint64(r.generation)
}

// Load a uint64 into a chain id.
func LoadChain(x uint64) Chain {
	return Chain{int32(x >> 32), uint16(x >> 16), uint16(x)}
}

// Store a joint id into a uint64.
func (r Joint) Store() uint64 {
	return uint64(r.index1)<<32 | uint64(r.world0)<<16 | uint64(r.generation)
}

// Load a uint64 into a joint id.
func LoadJoint(x uint64) Joint {
	return Joint{int32(x >> 32), uint16(x >> 16), uint16(x)}
}

// Store a contact id into three uint32 values
func (c Contact) Store() [3]uint32 {
	return [3]uint32{uint32(c.index1), uint32(c.world0), c.generation}
}

// Load three uint32 values into a contact id.
func LoadContact(values [3]uint32) Contact {
	return Contact{index1: int32(values[0]), world0: uint16(values[1]), generation: values[2]}
}

// Determine if an id is null.
func (id World) IsNull() bool {
	return id.index1 == 0
}

// Determine if an id is non-null.
func (id World) IsNonNull() bool {
	return id.index1 != 0
}

// Determine if an id is null.
func (id Body) IsNull() bool {
	return id.index1 == 0
}

// Determine if an id is non-null.
func (id Body) IsNonNull() bool {
	return id.index1 != 0
}

// Determine if an id is null.
func (id Shape) IsNull() bool {
	return id.index1 == 0
}

// Determine if an id is non-null.
func (id Shape) IsNonNull() bool {
	return id.index1 != 0
}

// Determine if an id is null.
func (id Chain) IsNull() bool {
	return id.index1 == 0
}

// Determine if an id is non-null.
func (id Chain) IsNonNull() bool {
	return id.index1 != 0
}

// Determine if an id is null.
func (id Joint) IsNull() bool {
	return id.index1 == 0
}

// Determine if an id is non-null.
func (id Joint) IsNonNull() bool {
	return id.index1 != 0
}

// Determine if an id is null.
func (id Contact) IsNull() bool {
	return id.index1 == 0
}

// Determine if an id is non-null.
func (id Contact) IsNonNull() bool {
	return id.index1 != 0
}
