// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

//go:build !box2d_double

package box2d

// Alias in single precision.
type Pos = Vec2

// Alias in single precision.
type WorldTransform = Transform

// This is an alias in single precision.
type WorldCastOutput = CastOutput

const huge float32 = 1.0e5
