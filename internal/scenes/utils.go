// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"runtime"

	"github.com/zigai/box2d"
)

const (
	randLimit = 32767
	RandSeed  = 12345
)

var gRandomSeed uint32 = RandSeed

func SetRandomSeed(seed uint32) { gRandomSeed = seed }

func RandomPolygon(extent float32) box2d.Polygon {
	var points [box2d.MaxPolygonVertices]box2d.Vec2

	count := 3 + RandomInt()%6
	for i := range count {
		points[i] = RandomVec2(-extent, extent)
	}

	hull := box2d.ComputeHull(points[:count])
	if hull.Count > 0 {
		return box2d.MakePolygon(&hull, 0)
	}

	return box2d.MakeSquare(extent)
}

func GetNumberOfCores() int { return runtime.NumCPU() }

func RandomInt() int {
	x := gRandomSeed
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	gRandomSeed = x

	return int(x % (randLimit + 1))
}

func RandomIntRange(lo, hi int) int { return lo + RandomInt()%(hi-lo+1) }

func RandomFloat() float32 {
	r := float32(RandomInt() & randLimit)
	r /= randLimit
	r = float32(2*r) - 1

	return r
}

func RandomFloatRange(lo, hi float32) float32 {
	r := float32(RandomInt() & randLimit)
	r /= randLimit
	r = float32((hi-lo)*r) + lo

	return r
}

func RandomVec2(lo, hi float32) box2d.Vec2 {
	var v box2d.Vec2

	v.X = RandomFloatRange(lo, hi)
	v.Y = RandomFloatRange(lo, hi)

	return v
}

func RandomPos(lo, hi float32) box2d.Pos {
	var v box2d.Vec2

	v.X = RandomFloatRange(lo, hi)
	v.Y = RandomFloatRange(lo, hi)

	return v.Pos()
}

func RandomRot() box2d.Rot {
	angle := RandomFloatRange(-box2d.Pi, box2d.Pi)
	return box2d.MakeRot(angle)
}
