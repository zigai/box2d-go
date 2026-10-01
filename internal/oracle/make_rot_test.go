// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package oracle

import (
	"fmt"
	"math"
	"testing"
)

const b2Pi float32 = 3.14159265359

func unwindAngle(radians float32) float32 {
	x := radians
	if x < -1.0e6 {
		x = -1.0e6
	} else if x > 1.0e6 {
		x = 1.0e6
	}

	// Work in doubles. Adding and removing this constant rounds to the nearest integer.
	twoPi := float64(2 * b2Pi)
	roundToNearest := 6755399441055744.0
	a := float64(x)

	k := (float64(a/twoPi) + roundToNearest) - roundToNearest
	return float32(a - float64(k*twoPi))
}

func computeCosSin(radians float32) (cosine, sine float32) {
	x := unwindAngle(radians)
	pi := b2Pi
	pi2 := float32(pi * pi)

	// cosine needs angle in [-pi/2, pi/2]
	var c float32
	switch {
	case x < float32(-0.5*pi):
		y := x + pi
		y2 := float32(y * y)
		c = -(pi2 - float32(4*y2)) / (pi2 + y2)
	case x > float32(0.5*pi):
		y := x - pi
		y2 := float32(y * y)
		c = -(pi2 - float32(4*y2)) / (pi2 + y2)
	default:
		y2 := float32(x * x)
		c = (pi2 - float32(4*y2)) / (pi2 + y2)
	}

	// sine needs angle in [0, pi]
	var s float32
	if x < 0 {
		y := x + pi
		num := float32(float32(-16*y) * (pi - y))
		den := float32(5*pi2) - float32(float32(4*y)*(pi-y))
		s = num / den
	} else {
		num := float32(float32(16*x) * (pi - x))
		den := float32(5*pi2) - float32(float32(4*x)*(pi-x))
		s = num / den
	}

	mag := float32(math.Sqrt(float64(float32(s*s) + float32(c*c))))
	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}

	return c * invMag, s * invMag
}

func makeRot(radians float32) (c, s float32) { return computeCosSin(radians) }

func makeRotAngles() []float32 {
	pi := b2Pi
	fb := math.Float32frombits
	negZero := float32(math.Copysign(0, -1))

	angles := []float32{
		0, negZero,
		pi, -pi,
		float32(0.5 * pi), float32(-0.5 * pi),
		float32(0.25 * pi), float32(-0.25 * pi),
		float32(0.75 * pi), float32(-0.75 * pi),
		fb(0x40490fdb), fb(0x40490fda), fb(0x40490fdc),
		float32(2 * pi), float32(-2 * pi),
		float32(3 * pi),
		1.0e-8, -1.0e-8, 1.0e-30,
		fb(0x00000001), fb(0x007fffff), fb(0x00800000),
		0.1, -0.1, 0.5, 1, -1, 2, 3, -3, 10, -10,
		100, -100, 1000, -1000, 9999.5, -9999.5,
	}

	for i := range 128 {
		angles = append(angles, float32(float32(i)*0.1)-6.4)
	}

	return angles
}

func key(i int, field string) string { return fmt.Sprintf("%d/%s", i, field) }

func TestMakeRotMatchesC(t *testing.T) {
	r := Open(t, "make_rot")
	angles := makeRotAngles()

	r.I32("count", int32(len(angles)))

	hash := HashInit
	for i, angle := range angles {
		cosine, sine := computeCosSin(angle)
		c, s := makeRot(angle)

		r.F32(key(i, "angle"), angle)
		r.F32(key(i, "unwound"), unwindAngle(angle))
		r.F32(key(i, "cosine"), cosine)
		r.F32(key(i, "sine"), sine)
		r.Rot(key(i, "rot"), c, s)

		hash = HashRot(hash, c, s)
	}

	r.U64("hash", hash)
}

func TestMakeRotWrongValueFails(t *testing.T) {
	angles := makeRotAngles()
	f := &fakeTB{}
	f.run(func() {
		r := Open(f, "make_rot")
		r.I32("count", int32(len(angles)))
		r.F32(key(0, "angle"), angles[0])
		r.F32(key(0, "unwound"), 0)
		r.F32(key(0, "cosine"), math.Float32frombits(0x3f800001))
	})

	mustContain(t, f.fatal,
		"first divergence in testdata/oracle/"+config+"/make_rot.jsonl, record 4",
		`key:      "0/cosine"`,
		"expected: 1 (0x3f800000)   (C oracle)",
		"got:      1.0000001 (0x3f800001)   (Go)",
		"ulp distance: 1",
	)
}
