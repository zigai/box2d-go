// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"github.com/go-gl/glfw/v3.4/glfw"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const convexHullCount = box2d.MaxPolygonVertices

type ConvexHull struct {
	Sample

	points     [box2d.MaxPolygonVertices]box2d.Vec2 //exhaustruct:optional
	count      int
	generation int
	auto       bool
	bulk       bool
}

func CreateConvexHull(context *SampleContext) Sampler {
	const (
		cameraCenterX   = 0.5
		cameraZoom      = 25.0
		cameraZoomScale = 0.3
	)

	s := &ConvexHull{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		count:           0,
		generation:      0,
		auto:            false,
		bulk:            false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: cameraCenterX, Y: 0.0}
		s.context.camera.zoom = cameraZoom * cameraZoomScale
	}

	s.Generate()

	return s
}

func (s *ConvexHull) Generate() {
	const (
		clampExtent = 4.0
		pointSpread = 10.0
	)

	angle := box2d.Pi * scenes.RandomFloat()
	r := box2d.MakeRot(angle)

	lowerBound := box2d.Vec2{X: -clampExtent, Y: -clampExtent}
	upperBound := box2d.Vec2{X: clampExtent, Y: clampExtent}

	for i := range convexHullCount {
		x := pointSpread * scenes.RandomFloat()
		y := pointSpread * scenes.RandomFloat()

		v := box2d.Vec2{X: x, Y: y}.Clamp(lowerBound, upperBound)
		s.points[i] = r.Rotate(v)
	}

	s.count = convexHullCount

	s.generation += 1
}

func (s *ConvexHull) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyA {
		s.auto = !s.auto
	}

	if key == glfw.KeyB {
		s.bulk = !s.bulk
	}

	if key == glfw.KeyG {
		s.Generate()
	}
}

func (s *ConvexHull) Step() {
	const (
		inputPointSize = 5.0
		hullPointSize  = 6.0
		labelOffset    = 0.1
	)

	s.Sample.Step()

	s.DrawScreenTextLinef("Options: generate(g), auto(a), bulk(b)")

	var hull box2d.Hull

	valid := s.computeHull(&hull)

	var milliseconds float32

	if !valid {
		s.DrawScreenTextLinef("generation = %d, FAILED", s.generation)
	} else {
		s.DrawScreenTextLinef("generation = %d, count = %d", s.generation, hull.Count)
	}

	if milliseconds > 0.0 {
		s.DrawScreenTextLinef("milliseconds = %.6g", milliseconds)
	}

	if hull.Count > 0 {
		s.draw.DrawPolygon(box2d.WorldTransformIdentity, hull.Points[:hull.Count], box2d.ColorGray)
	}

	for i := range s.count {
		s.draw.DrawPoint(s.points[i].Pos(), inputPointSize, box2d.ColorBlue)
		s.draw.DrawStringf(
			s.camera,
			s.points[i].Add(box2d.Vec2{X: labelOffset, Y: labelOffset}).Pos(),
			box2d.ColorWhite,
			"%d",
			i,
		)
	}

	for i := range hull.Count {
		s.draw.DrawPoint(hull.Points[i].Pos(), hullPointSize, box2d.ColorGreen)
	}
}

func (s *ConvexHull) computeHull(hull *box2d.Hull) bool {
	const bulkGenerationCount = 10000

	valid := false

	if !s.bulk {
		if s.auto {
			s.Generate()
		}

		*hull = box2d.ComputeHull(s.points[:s.count])
		if hull.Count > 0 {
			valid = box2d.ValidateHull(hull)
			if !valid {
				s.auto = false
			}
		}

		return valid
	}

	for range bulkGenerationCount {
		s.Generate()

		*hull = box2d.ComputeHull(s.points[:s.count])
		if hull.Count == 0 {
			continue
		}

		valid = box2d.ValidateHull(hull)
		if !valid || !s.bulk {
			s.bulk = false
			break
		}
	}

	return valid
}

func registerGeometrySamples() {
	RegisterSample("Geometry", "Convex Hull", CreateConvexHull)
}
