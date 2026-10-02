// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

var defaultMaterial = SurfaceMaterial{Friction: 0.6}

func emptyDrawPolygon(transform WorldTransform, vertices []Vec2, color HexColor) {
}

func emptyDrawSolidPolygon(transform WorldTransform, vertices []Vec2, radius float32, color HexColor) {
}

func emptyDrawCircle(center Pos, radius float32, color HexColor) {
}

func emptyDrawSolidCircle(transform WorldTransform, center Vec2, radius float32, color HexColor) {
}

func emptyDrawSolidCapsule(p1 Pos, p2 Pos, radius float32, color HexColor) {
}

func emptyDrawSegment(p1 Pos, p2 Pos, color HexColor) {
}

func emptyDrawTransform(transform WorldTransform) {
}

func emptyDrawPoint(p Pos, size float32, color HexColor) {
}

func emptyDrawString(p Pos, s string, color HexColor) {
}

func emptyDrawBounds(aabb AABB, color HexColor) {
}

// Use this to initialize your world definition
// @ingroup world
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
		internalValue:         secretCookie,
	}
}

// Use this to initialize your body definition
// @ingroup body
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
// @ingroup shape
func DefaultFilter() Filter {
	return Filter{CategoryBits: DefaultCategoryBits, MaskBits: DefaultMaskBits}
}

// Use this to initialize your query filter
// @ingroup shape
func DefaultQueryFilter() QueryFilter {
	return QueryFilter{CategoryBits: DefaultCategoryBits, MaskBits: DefaultMaskBits}
}

// Use this to initialize your shape definition
// @ingroup shape
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
// @ingroup shape
func DefaultSurfaceMaterial() SurfaceMaterial {
	return SurfaceMaterial{Friction: 0.6}
}

// Use this to initialize your chain definition
// @ingroup shape
func DefaultChainDef() ChainDef {
	infinity := float32(math.Inf(1))
	return ChainDef{
		Materials:     []SurfaceMaterial{defaultMaterial},
		MaterialCount: 1,
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
		DrawPolygonFunc:      emptyDrawPolygon,
		DrawSolidPolygonFunc: emptyDrawSolidPolygon,
		DrawCircleFunc:       emptyDrawCircle,
		DrawSolidCircleFunc:  emptyDrawSolidCircle,
		DrawSolidCapsuleFunc: emptyDrawSolidCapsule,
		DrawLineFunc:         emptyDrawSegment,
		DrawTransformFunc:    emptyDrawTransform,
		DrawPointFunc:        emptyDrawPoint,
		DrawStringFunc:       emptyDrawString,
		DrawBoundsFunc:       emptyDrawBounds,
		DrawingBounds: AABB{
			LowerBound: Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32},
			UpperBound: Vec2{X: math.MaxFloat32, Y: math.MaxFloat32},
		},
		ForceScale: 1,
		JointScale: 1,
		DrawShapes: true,
	}
}
