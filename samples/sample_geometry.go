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
	points     [box2d.MaxPolygonVertices]box2d.Vec2
	count      int
	generation int
	auto       bool
	bulk       bool
}

func CreateConvexHull(context *SampleContext) Sampler {
	s := &ConvexHull{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.5, Y: 0.0}
		s.context.camera.zoom = 25.0 * 0.3
	}

	s.generation = 0
	s.auto = false
	s.bulk = false
	s.Generate()
	return s
}

func (s *ConvexHull) Generate() {
	angle := box2d.Pi * scenes.RandomFloat()
	r := box2d.MakeRot(angle)

	lowerBound := box2d.Vec2{X: -4.0, Y: -4.0}
	upperBound := box2d.Vec2{X: 4.0, Y: 4.0}

	for i := range convexHullCount {
		x := 10.0 * scenes.RandomFloat()
		y := 10.0 * scenes.RandomFloat()

		v := box2d.Vec2{X: x, Y: y}.Clamp(lowerBound, upperBound)
		s.points[i] = r.Rotate(v)
	}

	s.count = convexHullCount

	s.generation += 1
}

func (s *ConvexHull) Keyboard(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyA:
		s.auto = !s.auto

	case glfw.KeyB:
		s.bulk = !s.bulk

	case glfw.KeyG:
		s.Generate()
	}
}

func (s *ConvexHull) Step() {
	s.Sample.Step()

	s.DrawScreenTextLine("Options: generate(g), auto(a), bulk(b)")

	var hull box2d.Hull
	valid := false
	var milliseconds float32

	if s.bulk {
		for range 10000 {
			s.Generate()
			hull = box2d.ComputeHull(s.points[:s.count])
			if hull.Count == 0 {
				continue
			}

			valid = box2d.ValidateHull(&hull)
			if !valid || !s.bulk {
				s.bulk = false
				break
			}
		}
	} else {
		if s.auto {
			s.Generate()
		}

		hull = box2d.ComputeHull(s.points[:s.count])
		if hull.Count > 0 {
			valid = box2d.ValidateHull(&hull)
			if !valid {
				s.auto = false
			}
		}
	}

	if !valid {
		s.DrawScreenTextLine("generation = %d, FAILED", s.generation)
	} else {
		s.DrawScreenTextLine("generation = %d, count = %d", s.generation, hull.Count)
	}

	if milliseconds > 0.0 {
		s.DrawScreenTextLine("milliseconds = %.6g", milliseconds)
	}

	if hull.Count > 0 {
		s.draw.DrawPolygon(box2d.WorldTransformIdentity, hull.Points[:hull.Count], box2d.ColorGray)
	}

	for i := range s.count {
		s.draw.DrawPoint(s.points[i].Pos(), 5.0, box2d.ColorBlue)
		s.draw.DrawString(s.camera, s.points[i].Add(box2d.Vec2{X: 0.1, Y: 0.1}).Pos(), box2d.ColorWhite, "%d", i)
	}

	for i := range hull.Count {
		s.draw.DrawPoint(hull.Points[i].Pos(), 6.0, box2d.ColorGreen)
	}
}

func init() {
	RegisterSample("Geometry", "Convex Hull", CreateConvexHull)
}
