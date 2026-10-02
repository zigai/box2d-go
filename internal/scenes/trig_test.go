// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package scenes

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
)

func TestBoundedSceneTrig(t *testing.T) {
	defer porttest.Guard(t)
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
	mix := func(hash uint64, value float32) uint64 {
		bits := math.Float32bits(value)
		for b := 0; b < 4; b++ {
			hash ^= uint64(byte(bits >> (8 * b)))
			hash *= 1099511628211
		}
		return hash
	}
	for step := 0; step < 300; step++ {
		state := uint32(1234) + uint32(7919)*uint32(step)
		hash := uint64(14695981039346656037)
		for ray := 0; ray < tileQueryCount; ray++ {
			tileRandom(&state, 0, 1)
			tileRandom(&state, -2, 18)
			angle := tileRandom(&state, 0, float32(2)*box2d.Pi)
			sine, cosine := boundedSinCos(angle)
			for _, value := range []float32{angle, sine, cosine, float32(25) * sine, float32(25) * cosine} {
				hash = mix(hash, value)
			}
			if step == 0 && ray == 41 {
				porttest.Ensure(t, math.Float32bits(angle) == 0x40bd9c85)
				porttest.Ensure(t, math.Float32bits(sine) == 0xbeb352db && math.Float32bits(float32(25)*sine) == 0xc10c18bb)
				dump(0, angle)
			}
		}
		r.U64(fmt.Sprintf("%d.hash", step), hash)
	}
	dump(1, box2d.Pi/float32(18))
	dump(2, float32(0.1)*(box2d.Pi/float32(18)))
	centers := []float32{0, math.Float32frombits(1), 0x1p-126, 0x1p-12, 0.75, box2d.Pi / 4, box2d.Pi / 2,
		float32(3) * box2d.Pi / 4, box2d.Pi, float32(5) * box2d.Pi / 4, float32(3) * box2d.Pi / 2, float32(7) * box2d.Pi / 4, float32(2) * box2d.Pi}
	index := 3
	for _, center := range centers {
		for _, value := range []float32{math.Nextafter32(center, float32(math.Inf(-1))), center, math.Nextafter32(center, float32(math.Inf(1)))} {
			dump(index, value)
			dump(index+1, -value)
			index += 2
		}
	}
	porttest.Ensure(t, index == 81)
}
