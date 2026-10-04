package box2d_test

import (
	"strconv"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

// These fixtures come from the separately built pinned Debug scalar engine,
// not the normal RelWithDebInfo producer. Only reported live vertices are read.
func TestDistanceDebugPublic(t *testing.T) {
	t.Run("FullHistory", func(t *testing.T) {
		r := oracle.Open(t, "distance_debug_public/history")
		input := box2d.DistanceInput{ProxyA: distancePublicBox(0.125), ProxyB: distancePublicBox(0.25), UseRadii: true}
		transforms := []box2d.Transform{
			{P: box2d.Vec2{4, 0.25}, Q: box2d.RotIdentity},
			{P: box2d.Vec2{3.75, 0.5}, Q: box2d.RotIdentity},
			{P: box2d.Vec2{3.5, 1}, Q: box2d.Rot{0.6, 0.8}},
			{P: box2d.Vec2{2, 1}, Q: box2d.Rot{0.6, 0.8}},
			{P: box2d.Vec2{0.5, 0.25}, Q: box2d.RotIdentity},
			{P: box2d.Vec2{-3, -1}, Q: box2d.Rot{0, -1}},
			{P: box2d.Vec2{-3.25, -0.75}, Q: box2d.Rot{0, -1}},
			{P: box2d.Vec2{4, 0.25}, Q: box2d.RotIdentity},
		}
		cache := box2d.SimplexCache{}

		for i, xf := range transforms {
			input.Transform = xf
			debug := make([]box2d.Simplex, 24)
			o := box2d.ShapeDistance(&input, &cache, debug)
			distancePublicOutput(r, strconv.Itoa(i), o, cache, debug)
		}
	})

	t.Run("CapacityOne", func(t *testing.T) {
		r := oracle.Open(t, "distance_debug_public/capacity-one")
		input := box2d.DistanceInput{
			ProxyA:    distancePublicBox(0),
			ProxyB:    box2d.MakeProxy([]box2d.Vec2{{0, -1}, {0, 1}}, 0),
			Transform: box2d.Transform{P: box2d.Vec2{3, 0}, Q: box2d.RotIdentity},
		}
		cache := box2d.SimplexCache{}
		debug := make([]box2d.Simplex, 1)
		o := box2d.ShapeDistance(&input, &cache, debug)
		distancePublicOutput(r, "output", o, cache, debug)
	})
}
