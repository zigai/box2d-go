// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

func TestBoundedSceneTrig(t *testing.T) {
	r := oracle.Open(t, "scenes/Trig")

	dump := func(index int, angle float32) {
		t.Helper()

		sine, cosine := boundedSinCos(angle)
		r.F32(fmt.Sprintf("%d.angle", index), angle)
		r.F32(fmt.Sprintf("%d.sin", index), sine)
		r.F32(fmt.Sprintf("%d.cos", index), cosine)
		r.F32(fmt.Sprintf("%d.y", index), float32(25)*sine)
		r.F32(fmt.Sprintf("%d.x", index), float32(25)*cosine)
	}
	for step := range 300 {
		hash := hashSceneTrigRays(t, step, dump)

		r.U64(fmt.Sprintf("%d.hash", step), hash)
	}

	dump(1, box2d.Pi/float32(18))
	dump(2, float32(0.1)*(box2d.Pi/float32(18)))
	centers := []float32{
		0, math.Float32frombits(1), 0x1p-126, 0x1p-12, 0.75, box2d.Pi / 4, box2d.Pi / 2,
		float32(3) * box2d.Pi / 4, box2d.Pi, float32(5) * box2d.Pi / 4, float32(3) * box2d.Pi / 2, float32(7) * box2d.Pi / 4, float32(2) * box2d.Pi,
	}
	index := 3

	for _, center := range centers {
		for _, value := range []float32{math.Nextafter32(center, float32(math.Inf(-1))), center, math.Nextafter32(center, float32(math.Inf(1)))} {
			dump(index, value)
			dump(index+1, -value)
			index += 2
		}
	}

	if index != 81 {
		t.Fatal("condition false")
	}
}

func mixSceneTrigFloat(hash uint64, value float32) uint64 {
	bits := math.Float32bits(value)
	for b := range 4 {
		hash ^= uint64(byte(bits >> (8 * b)))
		hash *= 1099511628211
	}

	return hash
}

func hashSceneTrigRays(t *testing.T, step int, dump func(int, float32)) uint64 {
	t.Helper()

	state := uint32(1234) + uint32(7919)*uint32(step)
	hash := uint64(14695981039346656037)

	for ray := range tileQueryCount {
		tileRandom(&state, 0, 1)
		tileRandom(&state, -2, 18)
		angle := tileRandom(&state, 0, float32(2)*box2d.Pi)

		sine, cosine := boundedSinCos(angle)
		for _, value := range []float32{angle, sine, cosine, float32(25) * sine, float32(25) * cosine} {
			hash = mixSceneTrigFloat(hash, value)
		}

		if step == 0 && ray == 41 {
			if !(math.Float32bits(angle) == 0x40bd9c85) {
				t.Fatal("condition false")
			}

			if !(math.Float32bits(sine) == 0xbeb352db && math.Float32bits(float32(25)*sine) == 0xc10c18bb) {
				t.Fatal("condition false")
			}

			dump(0, angle)
		}
	}

	return hash
}
