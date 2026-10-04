// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

const (
	// Maximum parallel workers. Used for some fixed size arrays.
	MaxWorkers = 32
	// Maximum number of tasks queued per world step. EnqueueTaskFunc will never be called
	// more than this per world step. This is related to MaxWorkers. With 32 workers,
	// the maximum observed task count is 130. This allows an external task system to use a fixed
	// size array for Box2D task, which may help with creating stable user task pointers.
	MaxTasks = 256
	// Maximum number of colors in the constraint graph. Constraints that cannot
	// find a color are added to the overflow set which are solved single-threaded.
	// The compound barrel benchmark has minor overflow with 24 colors
	GraphColorCount = 24
	// Maximum length of the body name. Can be 0 if you don't need names.
	NameLength = 10
	// The maximum rotation of a body per time step. This limit is very large and is used
	// to prevent numerical problems. You shouldn't need to adjust this.
	// @warning increasing this to 0.5f * Pi or greater will break continuous collision.
	MaxRotation float32 = 0.25 * Pi
	// The default contact recycling world angle threshold. 0.98 ~= 11.5 degrees
	ContactRecycleCosAngle float32 = 0.98
	// For small objects the margin is limited to this fraction times the maximum extent
	AABBMarginFraction float32 = 0.125
	// The time that a body must be still before it will go to sleep. In seconds.
	TimeToSleep float32 = 0.5
	// The maximum number of restitution iterations. Needed to avoid a solver overflow.
	MaxRestitutionIterations = 63
)

// Used to detect bad values. Positions greater than about 16km will have precision
// problems in single precision, so 100km as a limit should be fine in all cases.
// In large world mode the broad-phase starts to have excessive padding at 10,000km.
func Huge() float32 {
	return huge * lengthUnitsPerMeter
}

// A small length used as a collision and constraint tolerance. Usually it is
// chosen to be numerically significant, but visually insignificant. In meters.
// Normally this is 0.5cm.
// @warning modifying this can have a significant impact on stability
func LinearSlop() float32 {
	return 0.005 * lengthUnitsPerMeter
}

// Box2D uses limited speculative collision. This reduces jitter.
// Normally this is 2cm.
// @warning modifying this can have a significant impact on performance and stability
func SpeculativeDistance() float32 {
	return 4.0 * LinearSlop()
}

// The default contact recycling distance.
func ContactRecycleDistance() float32 {
	return 10.0 * LinearSlop()
}

// This is used to fatten AABBs in the dynamic tree. This allows proxies
// to move by a small amount without triggering a tree adjustment. This is in meters.
// Normally this is 5cm.
// This has been tuned on benchmarks and I found that keeping
// the margin small improves performance by reducing the number of contacts. I used
// to keep this larger in the past to avoid tree insert and remove churn. But the
// broad-phase no longer works that way and now uses a two-phase refit and rebuild
// setup similar to Jolt.
// @warning modifying this can have a significant impact on performance
func MaxAABBMargin() float32 {
	return 0.05 * lengthUnitsPerMeter
}
