// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"log/slog"
	"math"
)

var defaultMaterial = SurfaceMaterial{Friction: 0.6}

// Use this to initialize your world definition
// @ingroup world.
func DefaultWorldDef() WorldDef {
	return WorldDef{
		Gravity:               Vec2{Y: -10},
		HitEventThreshold:     1 * lengthUnitsPerMeter,
		RestitutionThreshold:  1 * lengthUnitsPerMeter,
		RestitutionIterations: 2,
		ContactSpeed:          3 * lengthUnitsPerMeter,
		ContactHertz:          30,
		ContactDampingRatio:   10,
		MaximumLinearSpeed:    400 * lengthUnitsPerMeter,
		EnableSleep:           true,
		EnableContinuous:      true,
		Logger:                slog.Default(),
		internalValue:         secretCookie,
	}
}

// Use this to initialize your body definition
// @ingroup body.
func DefaultBodyDef() BodyDef {
	return BodyDef{
		Type:                   StaticBody,
		Rotation:               Rot{C: 1},
		SafetyFactor:           0.5,
		SleepThreshold:         0.05 * lengthUnitsPerMeter,
		GravityScale:           1,
		EnableSleep:            true,
		EnableContactRecycling: true,
		IsAwake:                true,
		IsEnabled:              true,
		internalValue:          secretCookie,
	}
}

// Use this to initialize your filter
// @ingroup shape.
func DefaultFilter() Filter {
	return Filter{CategoryBits: DefaultCategoryBits, MaskBits: DefaultMaskBits}
}

// Use this to initialize your query filter
// @ingroup shape.
func DefaultQueryFilter() QueryFilter {
	return QueryFilter{CategoryBits: DefaultCategoryBits, MaskBits: DefaultMaskBits}
}

// Use this to initialize your shape definition
// @ingroup shape.
func DefaultShapeDef() ShapeDef {
	return ShapeDef{
		Material:              SurfaceMaterial{Friction: 0.6},
		Density:               1,
		Filter:                DefaultFilter(),
		UpdateBodyMass:        true,
		InvokeContactCreation: true,
		internalValue:         secretCookie,
	}
}

// Use this to initialize your surface material
// @ingroup shape.
func DefaultSurfaceMaterial() SurfaceMaterial {
	return SurfaceMaterial{Friction: 0.6}
}

// Use this to initialize your chain definition
// @ingroup shape.
func DefaultChainDef() ChainDef {
	infinity := float32(math.Inf(1))

	return ChainDef{
		Materials:     []SurfaceMaterial{defaultMaterial},
		Ghost1:        Vec2{X: infinity, Y: infinity},
		Ghost2:        Vec2{X: infinity, Y: infinity},
		Filter:        DefaultFilter(),
		internalValue: secretCookie,
	}
}

// Use this to initialize your drawing interface. This allows you to implement a sub-set
// of the drawing functions.
func DefaultDebugDraw() DebugDraw {
	return DebugDraw{
		DrawingBounds: AABB{
			LowerBound: Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32},
			UpperBound: Vec2{X: math.MaxFloat32, Y: math.MaxFloat32},
		},
		ForceScale: 1,
		JointScale: 1,
		DrawShapes: true,
	}
}
