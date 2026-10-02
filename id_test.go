// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func TestId(t *testing.T) {

	r := oracle.Open(t, "id/IdTest")
	a := uint32(0x01234567)
	{
		id := box2d.LoadWorld(a)
		b := id.Store()
		if !(b == a) {
			t.Fatal("condition false")
		}
		r.U32("world", b)
	}
	x := uint64(0x0123456789ABCDEF)
	{
		id := box2d.LoadBody(x)
		y := id.Store()
		if !(x == y) {
			t.Fatal("condition false")
		}
		r.U64("body", y)
	}
	{
		id := box2d.LoadShape(x)
		y := id.Store()
		if !(x == y) {
			t.Fatal("condition false")
		}
		r.U64("shape", y)
	}
	{
		id := box2d.LoadChain(x)
		y := id.Store()
		if !(x == y) {
			t.Fatal("condition false")
		}
		r.U64("chain", y)
	}
	{
		id := box2d.LoadJoint(x)
		y := id.Store()
		if !(x == y) {
			t.Fatal("condition false")
		}
		r.U64("joint", y)
	}
}
