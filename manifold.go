// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// A manifold point is a contact point belonging to a contact manifold.
// It holds details related to the geometry and dynamics of the contact points.
// Box2D uses speculative collision so some contact points may be separated.
// You may use the TotalNormalImpulse to determine if there was an interaction during
// the time step.
type ManifoldPoint struct {
	// Location of the contact point relative to body A's center of mass in world space.
	// This can be converted to a world point using:
	// Pos worldPointA = OffsetPos(Body.WorldCenter(myBodyA), AnchorA);
	// @note When used internally to the Box2D solver, this is relative to the body center of mass.
	AnchorA Vec2
	// Location of the contact point relative to body B's center of mass in world space.
	// This can be converted to a world point using:
	// Pos worldPointB = OffsetPos(Body.WorldCenter(myBodyB), AnchorB);
	// @note When used internally to the Box2D solver, this is relative to the body center of mass.
	AnchorB Vec2
	// The separation of the contact point, negative if penetrating
	Separation float32
	// The impulse along the manifold normal vector.
	NormalImpulse float32
	// The friction impulse
	TangentImpulse float32
	// The total normal impulse applied across sub-stepping and restitution. This is important
	// to identify speculative contact points that had an interaction in the time step.
	// This includes the warm starting impulse, the sub-step delta impulse, and the restitution
	// impulse.
	TotalNormalImpulse float32
	// Relative normal velocity pre-solve. Negative when approaching. This is only
	// computed if hit events are enabled. It may be computed otherwise as a side effect
	// of SIMD bundling, but that is not guaranteed.
	NormalVelocity float32
	// Cached separation used for contact recycling
	BaseSeparation float32
	// Uniquely identifies a contact point between two shapes
	ID uint16
	// Did this contact point exist the previous step?
	Persisted bool
}

// A contact manifold describes the contact points between colliding shapes.
// @note Box2D uses speculative collision so some contact points may be separated.
type Manifold struct {
	// The unit normal vector in world space, points from shape A to body B
	Normal Vec2
	// Angular impulse applied for rolling resistance. N * m * s = kg * m^2 / s
	RollingImpulse float32
	// The manifold points, up to two are possible in 2D
	Points [2]ManifoldPoint
	// The number of contacts points, will be 0, 1, or 2
	PointCount int
}

// Contact manifold point in local coordinates (frame A).
type LocalManifoldPoint struct {
	// Contact point in frame A
	Point Vec2
	// The separation of the contact point, negative if penetrating. May be positive or negative.
	Separation float32
	// Uniquely identifies a contact point between two shapes
	ID uint16
}

// Contact manifold in local coordinates (frame A).
type LocalManifold struct {
	// The unit normal vector in frame A, points from shape A to shape B
	Normal Vec2
	// The manifold points, up to two are possible in 2D
	Points [2]LocalManifoldPoint
	// The number of contacts points, will be 0, 1, or 2
	PointCount int
}

type normalType int32

const (
	normalSkip normalType = iota
	normalAdmit
	normalSnap
)

type chainSegmentParams struct {
	edge1   Vec2
	normal0 Vec2
	normal2 Vec2
	convex1 bool
	convex2 bool
}

func makeCapsule(p1, p2 Vec2, radius float32) Polygon {
	todo()
	return Polygon{}
}

// Compute the contact manifold between two circles
func CollideCircles(circleA, circleB *Circle, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between a capsule and circle
func CollideCapsuleAndCircle(capsuleA *Capsule, circleB *Circle, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between a polygon and a circle
func CollidePolygonAndCircle(polygonA *Polygon, circleB *Circle, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between a capsule and circle
func CollideCapsules(capsuleA, capsuleB *Capsule, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between an segment and a capsule
func CollideSegmentAndCapsule(segmentA *Segment, capsuleB *Capsule, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between a polygon and capsule
func CollidePolygonAndCapsule(polygonA *Polygon, capsuleB *Capsule, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

func clipPolygons(polyA, polyB *Polygon, edgeA, edgeB int, flip bool) LocalManifold {
	todo()
	return LocalManifold{}
}

func findMaxSeparation(edgeIndex *int, poly1, poly2 *Polygon) float32 {
	todo()
	return 0
}

// Compute the contact manifold between two polygons
func CollidePolygons(polygonA, polygonB *Polygon, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between an segment and a circle
func CollideSegmentAndCircle(segmentA *Segment, circleB *Circle, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between an segment and a polygon
func CollideSegmentAndPolygon(segmentA *Segment, polygonB *Polygon, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between a chain segment and a circle
func CollideChainSegmentAndCircle(segmentA *ChainSegment, circleB *Circle, xf Transform) LocalManifold {
	todo()
	return LocalManifold{}
}

// Compute the contact manifold between a chain segment and a capsule
func CollideChainSegmentAndCapsule(segmentA *ChainSegment, capsuleB *Capsule, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func clipSegments(a1, a2, b1, b2, normal Vec2, ra, rb float32, id1, id2 uint16) LocalManifold {
	todo()
	return LocalManifold{}
}

func classifyNormal(params chainSegmentParams, normal Vec2) normalType {
	todo()
	return normalType(0)
}

// Compute the contact manifold between a chain segment and a rounded polygon
func CollideChainSegmentAndPolygon(segmentA *ChainSegment, polygonB *Polygon, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}
