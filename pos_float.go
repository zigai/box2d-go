// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

//go:build !box2d_double

package box2d

const huge float32 = 1.0e5

const doublePrecision = false

type posScalar = float32
