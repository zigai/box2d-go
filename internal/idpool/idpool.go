// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package idpool

import (
	"slices"

	"github.com/zigai/box2d-go/internal/array"
	"github.com/zigai/box2d-go/internal/core"
)

type Pool struct {
	FreeArray []int
	NextIndex int
}

func New() Pool {
	return Pool{FreeArray: make([]int, 0, 32)}
}

func (p *Pool) Destroy() {
	*p = Pool{}
}

func (p *Pool) Alloc() int {
	if len(p.FreeArray) > 0 {
		var id int

		p.FreeArray, id = array.Pop(p.FreeArray)

		return id
	}

	id := p.NextIndex
	p.NextIndex++

	return id
}

func AllocWithGeneration[Generation ~uint16 | ~uint32](p *Pool, generation func(int) Generation) int {
	for freeIndex, id := range slices.Backward(p.FreeArray) {
		if generation(id) == ^Generation(0) {
			continue
		}

		p.FreeArray = slices.Delete(p.FreeArray, freeIndex, freeIndex+1)

		return id
	}

	id := p.NextIndex
	p.NextIndex++

	return id
}

func (p *Pool) Free(id int) {
	core.Assert(p.NextIndex > 0, "pool.nextIndex > 0")
	core.Assert(0 <= id && id < p.NextIndex, "0 <= id && id < pool.nextIndex")
	p.FreeArray = array.Push(p.FreeArray, id)
}

func (p *Pool) ValidateFree(id int) {
	if !core.ValidationEnabled {
		return
	}

	if slices.Contains(p.FreeArray, id) {
		return
	}

	core.Assert(false, "0")
}

func (p *Pool) ValidateUsed(id int) {
	if !core.ValidationEnabled {
		return
	}

	for _, freeID := range p.FreeArray {
		if freeID == id {
			core.Assert(false, "0")
		}
	}
}

func (p *Pool) Count() int {
	return p.NextIndex - len(p.FreeArray)
}

func (p *Pool) Capacity() int {
	return p.NextIndex
}
