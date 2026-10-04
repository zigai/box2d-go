// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// 2D vector
// This can be used to represent a point or free vector.
type Vec2 struct {
	// coordinates
	X, Y float32
}

// Cosine and sine pair
// This uses a custom implementation designed for cross-platform determinism.
type CosSin struct {
	// cosine and sine
	Cosine float32
	Sine   float32
}

// 2D rotation
// This is similar to using a complex number for rotation.
type Rot struct {
	// cosine and sine
	C, S float32
}

// A 2D rigid transform.
type Transform struct {
	P Vec2
	Q Rot
}

// A 2-by-2 Matrix.
type Mat22 struct {
	// columns
	Cx, Cy Vec2
}

// Axis-aligned bounding box.
type AABB struct {
	LowerBound Vec2
	UpperBound Vec2
}

// separation = dot(normal, point) - offset.
type Plane struct {
	Normal Vec2
	Offset float32
}

// https://en.wikipedia.org/wiki/Pi
const Pi float32 = 3.14159265359

var (
	Vec2Zero               = Vec2{}
	RotIdentity            = Rot{C: 1}
	TransformIdentity      = Transform{Q: Rot{C: 1}}
	Mat22Zero              = Mat22{}
	PosZero                = Pos{}
	WorldTransformIdentity = WorldTransform{Q: Rot{C: 1}}
)
