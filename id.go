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
	todo()
	var result uint32
	return result
}

// Load a uint32 into a world id.
func LoadWorld(x uint32) World {
	todo()
	var result World
	return result
}

// Store a body id into a uint64.
func (r Body) Store() uint64 {
	todo()
	var result uint64
	return result
}

// Load a uint64 into a body id.
func LoadBody(x uint64) Body {
	todo()
	var result Body
	return result
}

// Store a shape id into a uint64.
func (r Shape) Store() uint64 {
	todo()
	var result uint64
	return result
}

// Load a uint64 into a shape id.
func LoadShape(x uint64) Shape {
	todo()
	var result Shape
	return result
}

// Store a chain id into a uint64.
func (r Chain) Store() uint64 {
	todo()
	var result uint64
	return result
}

// Load a uint64 into a chain id.
func LoadChain(x uint64) Chain {
	todo()
	var result Chain
	return result
}

// Store a joint id into a uint64.
func (r Joint) Store() uint64 {
	todo()
	var result uint64
	return result
}

// Load a uint64 into a joint id.
func LoadJoint(x uint64) Joint {
	todo()
	var result Joint
	return result
}

// Store a contact id into three uint32 values
func (c Contact) Store() [3]uint32 {
	todo()
	var result [3]uint32
	return result
}

// Load three uint32 values into a contact id.
func LoadContact(values [3]uint32) Contact {
	todo()
	var result Contact
	return result
}

// Determine if an id is null.
func (id World) IsNull() bool {
	todo()
	return false
}

// Determine if an id is non-null.
func (id World) IsNonNull() bool {
	todo()
	return false
}

// Determine if an id is null.
func (id Body) IsNull() bool {
	todo()
	return false
}

// Determine if an id is non-null.
func (id Body) IsNonNull() bool {
	todo()
	return false
}

// Determine if an id is null.
func (id Shape) IsNull() bool {
	todo()
	return false
}

// Determine if an id is non-null.
func (id Shape) IsNonNull() bool {
	todo()
	return false
}

// Determine if an id is null.
func (id Chain) IsNull() bool {
	todo()
	return false
}

// Determine if an id is non-null.
func (id Chain) IsNonNull() bool {
	todo()
	return false
}

// Determine if an id is null.
func (id Joint) IsNull() bool {
	todo()
	return false
}

// Determine if an id is non-null.
func (id Joint) IsNonNull() bool {
	todo()
	return false
}

// Determine if an id is null.
func (id Contact) IsNull() bool {
	todo()
	return false
}

// Determine if an id is non-null.
func (id Contact) IsNonNull() bool {
	todo()
	return false
}
