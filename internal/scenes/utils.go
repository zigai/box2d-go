// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"github.com/zigai/box2d"
	"runtime"
)

const (
	randLimit = 32767
	randSeed  = 12345
)

var gRandomSeed uint32 = randSeed

func RandomPolygon(extent float32) box2d.Polygon {
	todo()
	return box2d.Polygon{}
}

func GetNumberOfCores() int {
	todo()
	return 0
}

func RandomInt() int {
	todo()
	return 0
}

func RandomIntRange(lo, hi int) int {
	todo()
	return 0
}

func RandomFloat() float32 {
	todo()
	return 0
}

func RandomFloatRange(lo, hi float32) float32 {
	todo()
	return 0
}

func RandomVec2(lo, hi float32) box2d.Vec2 {
	todo()
	return box2d.Vec2{}
}

func RandomPos(lo, hi float32) box2d.Pos {
	todo()
	return box2d.Pos{}
}

func RandomRot() box2d.Rot {
	todo()
	return box2d.Rot{}
}

type unported string

func (u unported) Unported() string { return string(u) }

func todo() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic(unported("scenes"))
	}
	panic(unported(runtime.FuncForPC(pc).Name()))
}
