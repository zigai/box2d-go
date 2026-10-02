// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

//go:build box2d_double

package box2d

// A world position. Double precision in large world mode so coordinates stay accurate far
// from the origin.
type Pos struct {
	X, Y float64
}

// A world transform with double precision translation and float rotation. Rotation is frame
// local and never needs the extra range, the same split as Jolt's DMat44.
type WorldTransform struct {
	P Pos
	Q Rot
}

// World-space cast output. The hit point is a world position. The normal stays a float direction.
type WorldCastOutput struct {
	// The surface normal at the hit point
	Normal Vec2
	// The surface hit point in world space
	Point Pos
	// The fraction of the input translation at collision
	Fraction float32
	// The number of iterations used
	Iterations int
	// Did the cast hit?
	Hit bool
}

const huge float32 = 1.0e9
