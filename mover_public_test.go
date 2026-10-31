// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func TestCastMoverSkipsSensors(t *testing.T) {
	worldDef := box2d.DefaultWorldDef()
	world := box2d.NewWorld(&worldDef)
	box2d.CleanupWorld(t, world)

	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Vec2{X: 2}.Pos()
	body := world.CreateBody(&bodyDef)
	box := box2d.MakeBox(0.5, 0.5)
	shapeDef := box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	body.CreatePolygonShape(&shapeDef, &box)

	mover := box2d.Capsule{Center1: box2d.Vec2{Y: -0.2}, Center2: box2d.Vec2{Y: 0.2}, Radius: 0.3}

	filter := box2d.DefaultQueryFilter()
	if got := world.CastMover(box2d.Pos{}, &mover, box2d.Vec2{X: 6}, filter); got != 1 {
		t.Fatalf("sensor-only cast fraction = %g, want 1", got)
	}

	bodyDef.Position = box2d.Vec2{X: 4}.Pos()
	body = world.CreateBody(&bodyDef)
	shapeDef.IsSensor = false
	body.CreatePolygonShape(&shapeDef, &box)

	if got := world.CastMover(box2d.Pos{}, &mover, box2d.Vec2{X: 6}, filter); math.Abs(float64(got-3.2/6)) > 0.01 {
		t.Fatalf("solid behind sensor cast fraction = %g, want near %g", got, 3.2/6)
	}
}

func TestCastMoverPreservesNearContactAndInitialOverlap(t *testing.T) {
	worldDef := box2d.DefaultWorldDef()
	world := box2d.NewWorld(&worldDef)
	box2d.CleanupWorld(t, world)

	bodyDef := box2d.DefaultBodyDef()
	body := world.CreateBody(&bodyDef)
	box := box2d.MakeBox(0.5, 0.5)
	shapeDef := box2d.DefaultShapeDef()
	body.CreatePolygonShape(&shapeDef, &box)

	near := box2d.Capsule{Center1: box2d.Vec2{X: -0.9955, Y: -0.2}, Center2: box2d.Vec2{X: -0.9955, Y: 0.2}, Radius: 0.5}

	filter := box2d.DefaultQueryFilter()
	if got := world.CastMover(box2d.Pos{}, &near, box2d.Vec2{X: 1}, filter); got <= 0 || got >= 0.02 {
		t.Fatalf("near-contact cast fraction = %g, want a small positive advance", got)
	}

	near.Center1.X, near.Center2.X = -0.9, -0.9
	if got := world.CastMover(box2d.Pos{}, &near, box2d.Vec2{X: -1}, filter); got != 1 {
		t.Fatalf("cast away from initial overlap = %g, want 1", got)
	}
}

func geometryMoverObserve(r *oracle.Reader, result box2d.PlaneSolverResult, planes []box2d.CollisionPlane) {
	r.Vec2("full/delta", result.Delta.X, result.Delta.Y)
	r.I32("full/iterations", int32(result.IterationCount))
	r.I32("full/plane-count", int32(len(planes)))

	for i := range planes {
		r.F32(fmt.Sprintf("full/plane%d/push", i), planes[i].Push)
	}
}

func geometryMoverClipObserve(r *oracle.Reader, v box2d.Vec2) {
	r.Vec2("full/clip", v.X, v.Y)
}

func TestMoverOffsetsLimitsAndClipPolicies(t *testing.T) {
	r := oracle.Open(t, "mover/OffsetsLimitsAndClipPolicies")

	cases := []struct {
		name   string
		target box2d.Vec2
		planes []box2d.CollisionPlane
	}{
		{"positive-offset-penetration", box2d.Vec2{0.75, -1}, []box2d.CollisionPlane{{Plane: box2d.Plane{Normal: box2d.Vec2{0, 1}, Offset: 0.25}, PushLimit: math.MaxFloat32, Push: 7, ClipVelocity: true}}},
		{"negative-offset-separated", box2d.Vec2{0.75, 0.25}, []box2d.CollisionPlane{{Plane: box2d.Plane{Normal: box2d.Vec2{0, 1}, Offset: -0.5}, PushLimit: math.MaxFloat32, Push: 7, ClipVelocity: true}}},
		{"zero-limit", box2d.Vec2{0.75, -1}, []box2d.CollisionPlane{{Plane: box2d.Plane{Normal: box2d.Vec2{0, 1}, Offset: 0.25}, PushLimit: 0, Push: 7, ClipVelocity: true}}},
		{"finite-slope", box2d.Vec2{-1, -1}, []box2d.CollisionPlane{{Plane: box2d.Plane{Normal: box2d.Vec2{0.8, 0.6}, Offset: 0.15}, PushLimit: 0.3, Push: 7, ClipVelocity: false}}},
		{"rigid-slope", box2d.Vec2{-1, -1}, []box2d.CollisionPlane{{Plane: box2d.Plane{Normal: box2d.Vec2{0.8, 0.6}, Offset: -0.15}, PushLimit: math.MaxFloat32, Push: 7, ClipVelocity: true}}},
		{"mixed-soft-rigid", box2d.Vec2{-1, -1}, []box2d.CollisionPlane{
			{Plane: box2d.Plane{Normal: box2d.Vec2{0, 1}, Offset: 0.1}, PushLimit: 0.2, Push: 7, ClipVelocity: false},
			{Plane: box2d.Plane{Normal: box2d.Vec2{0.8, 0.6}, Offset: -0.1}, PushLimit: math.MaxFloat32, Push: 7, ClipVelocity: true},
		}},
		{"three-level-corner", box2d.Vec2{-1, -1}, []box2d.CollisionPlane{
			{Plane: box2d.Plane{Normal: box2d.Vec2{0, 1}, Offset: 0.1}, PushLimit: math.MaxFloat32, Push: 7, ClipVelocity: true},
			{Plane: box2d.Plane{Normal: box2d.Vec2{1, 0}, Offset: -0.2}, PushLimit: math.MaxFloat32, Push: 7, ClipVelocity: true},
			{Plane: box2d.Plane{Normal: box2d.Vec2{0.8, 0.6}, Offset: 0.5}, PushLimit: 0.1, Push: 7, ClipVelocity: false},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := box2d.SolvePlanes(tc.target, tc.planes)
			geometryMoverObserve(r, result, tc.planes)

			for _, velocity := range []box2d.Vec2{{-2, -3}, {2, 3}, {1, -1}, {-1, 1}, {0, 0}} {
				geometryMoverClipObserve(r, box2d.ClipVector(velocity, tc.planes))
			}
		})
	}
	// Explicit independent clip policy partitions: active, zero push, disabled,
	// and a second active plane. A nonzero X component detects incomplete observers.
	planes := []box2d.CollisionPlane{
		{Plane: box2d.Plane{Normal: box2d.Vec2{0, 1}}, Push: 0.25, ClipVelocity: true},
		{Plane: box2d.Plane{Normal: box2d.Vec2{1, 0}}, Push: 0, ClipVelocity: true},
		{Plane: box2d.Plane{Normal: box2d.Vec2{0.8, 0.6}}, Push: 0.25, ClipVelocity: false},
	}
	for policy := range 4 {
		geometryMoverClipObserve(r, box2d.ClipVector(box2d.Vec2{-2, -3}, planes))

		switch policy {
		case 0:
			planes[1].Push = 0.25
		case 1:
			planes[2].ClipVelocity = true
		case 2:
			planes[0].Push = 0
		}
	}
}
