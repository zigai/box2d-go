package box2d_test

import (
	"testing"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

// These fixtures come from the separately built pinned Debug scalar engine,
// not the normal RelWithDebInfo producer. Only reported live vertices are read.
func TestDistanceDebugPublic(t *testing.T) {
	t.Run("FullHistory", func(t *testing.T) {
		r := oracle.Open(t, "distance_debug_public/history")
		distancePublicCacheHistory(r)
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
