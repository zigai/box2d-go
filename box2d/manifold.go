// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"math"

	"github.com/zigai/box2d-go/internal/core"
)

const (
	normalSkip normalType = iota
	normalAdmit
	normalSnap
)

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

type chainSegmentParams struct {
	edge1   Vec2
	normal0 Vec2
	normal2 Vec2
	convex1 bool
	convex2 bool
}

// Compute the contact manifold between two circles.
func CollideCircles(circleA, circleB *Circle, xf Transform) LocalManifold {
	manifold := LocalManifold{}

	var (
		pointA = circleA.Center
		pointB = xf.Apply(circleB.Center)
	)

	distance, normal := pointB.Sub(pointA).LengthAndNormalize()

	var (
		radiusA = circleA.Radius
		radiusB = circleB.Radius
	)

	separation := distance - radiusA - radiusB
	if separation > SpeculativeDistance() {
		return manifold
	}

	var (
		cA = pointA.MulAdd(radiusA, normal)
		cB = pointB.MulAdd(-radiusB, normal)
	)

	manifold.Normal = normal

	mp := &manifold.Points[0]

	mp.Point = cA.Lerp(cB, 0.5)
	mp.Separation = separation
	mp.ID = 0
	manifold.PointCount = 1

	return manifold
}

// Compute the contact manifold between a capsule and circle.
func CollideCapsuleAndCircle(capsuleA *Capsule, circleB *Circle, xf Transform) LocalManifold {
	manifold := LocalManifold{}

	// Compute circle position in the frame of the capsule.
	pB := xf.Apply(circleB.Center)

	// Compute closest point
	var (
		p1 = capsuleA.Center1
		p2 = capsuleA.Center2
	)

	e := p2.Sub(p1)

	// dot(p - pA, e) = 0
	// dot(p - (p1 + s1 * e), e) = 0
	// s1 = dot(p - p1, e)
	var (
		pA Vec2
		s1 = pB.Sub(p1).Dot(e)
		s2 = p2.Sub(pB).Dot(e)
	)
	if s1 < 0.0 {
		// p1 region
		pA = p1
	} else if s2 < 0.0 {
		// p2 region
		pA = p2
	} else {
		// circle colliding with segment interior
		s := s1 / e.Dot(e)

		pA = p1.MulAdd(s, e)
	}

	distance, normal := pB.Sub(pA).LengthAndNormalize()

	var (
		radiusA    = capsuleA.Radius
		radiusB    = circleB.Radius
		separation = distance - radiusA - radiusB
	)
	if separation > SpeculativeDistance() {
		return manifold
	}

	var (
		cA = pA.MulAdd(radiusA, normal)
		cB = pB.MulAdd(-radiusB, normal)
	)

	manifold.Normal = normal

	mp := &manifold.Points[0]

	mp.Point = cA.Lerp(cB, 0.5)
	mp.Separation = separation
	mp.ID = 0
	manifold.PointCount = 1

	return manifold
}

// Compute the contact manifold between a polygon and a circle.
func CollidePolygonAndCircle(polygonA *Polygon, circleB *Circle, xf Transform) LocalManifold {
	var (
		manifold            = LocalManifold{}
		speculativeDistance = SpeculativeDistance()
	)

	// Compute circle position in the frame of the polygon.
	var (
		center  = xf.Apply(circleB.Center)
		radiusA = polygonA.Radius
		radiusB = circleB.Radius
		radius  = radiusA + radiusB
	)

	// Find the min separating edge.
	var (
		normalIndex = 0
		separation  = -float32(0x1.fffffep127)
		vertexCount = polygonA.Count
		vertices    = polygonA.Vertices[:]
		normals     = polygonA.Normals[:]
	)

	for i := range vertexCount {
		s := normals[i].Dot(center.Sub(vertices[i]))
		if s > separation {
			separation = s
			normalIndex = i
		}
	}

	if separation > radius+speculativeDistance {
		return manifold
	}

	// Vertices of the reference edge.
	var (
		vertIndex1 = normalIndex
		vertIndex2 int
	)
	if vertIndex1+1 < vertexCount {
		vertIndex2 = vertIndex1 + 1
	} else {
		vertIndex2 = 0
	}

	var (
		v1 = vertices[vertIndex1]
		v2 = vertices[vertIndex2]
	)

	// Compute barycentric coordinates
	var (
		u1 = center.Sub(v1).Dot(v2.Sub(v1))
		u2 = center.Sub(v2).Dot(v1.Sub(v2))
	)

	if u1 < 0.0 && separation > float32(0x1p-23) {
		// Circle center is closest to v1 and safely outside the polygon
		normal := center.Sub(v1).Normalize()

		separation = center.Sub(v1).Dot(normal)
		if separation > radius+speculativeDistance {
			return manifold
		}

		var (
			cA            = v1.MulAdd(radiusA, normal)
			cB            = center.MulSub(radiusB, normal)
			contactPointA = cA.Lerp(cB, 0.5)
		)

		manifold.Normal = normal

		mp := &manifold.Points[0]

		mp.Point = contactPointA
		mp.Separation = cB.Sub(cA).Dot(normal)
		mp.ID = 0
		manifold.PointCount = 1
	} else if u2 < 0.0 && separation > float32(0x1p-23) {
		// Circle center is closest to v2 and safely outside the polygon
		normal := center.Sub(v2).Normalize()

		separation = center.Sub(v2).Dot(normal)
		if separation > radius+speculativeDistance {
			return manifold
		}

		var (
			cA            = v2.MulAdd(radiusA, normal)
			cB            = center.MulSub(radiusB, normal)
			contactPointA = cA.Lerp(cB, 0.5)
		)

		manifold.Normal = normal

		mp := &manifold.Points[0]

		mp.Point = contactPointA
		mp.Separation = cB.Sub(cA).Dot(normal)
		mp.ID = 0
		manifold.PointCount = 1
	} else {
		// Circle center is between v1 and v2. Center may be inside polygon
		normal := normals[normalIndex]

		manifold.Normal = normal

		// cA is the projection of the circle center onto to the reference edge
		cA := center.MulAdd(radiusA-center.Sub(v1).Dot(normal), normal)

		// cB is the deepest point on the circle with respect to the reference edge
		cB := center.MulSub(radiusB, normal)

		// The contact point is the midpoint
		mp := &manifold.Points[0]

		mp.Point = cA.Lerp(cB, 0.5)
		mp.Separation = separation - radius
		mp.ID = 0
		manifold.PointCount = 1
	}

	return manifold
}

// Compute the contact manifold between a capsule and circle.
func CollideCapsules(capsuleA, capsuleB *Capsule, xf Transform) LocalManifold {
	origin := capsuleA.Center1

	// Shift to the origin in frame A for round-off, a pure translation in A's frame
	xfs := Transform{xf.P.Sub(origin), xf.Q}

	var (
		p1 = Vec2{}
		q1 = capsuleA.Center2.Sub(origin)
	)

	var (
		p2 = xfs.Apply(capsuleB.Center1)
		q2 = xfs.Apply(capsuleB.Center2)
	)

	var (
		d1 = q1.Sub(p1)
		d2 = q2.Sub(p2)
	)

	var (
		dd1 = d1.Dot(d1)
		dd2 = d2.Dot(d2)
	)

	epsSqr := float32(0x1p-23) * float32(0x1p-23)
	core.Assert(dd1 > epsSqr && dd2 > epsSqr, "dd1 > epsSqr && dd2 > epsSqr")

	var (
		r   = p1.Sub(p2)
		rd1 = r.Dot(d1)
		rd2 = r.Dot(d2)
	)

	d12 := d1.Dot(d2)

	denom := float32(dd1*dd2) - float32(d12*d12)

	// Fraction on segment 1
	var f1 float32 = 0.0
	if denom != 0.0 {
		// not parallel
		f1 = clampFloat((float32(d12*rd2)-float32(rd1*dd2))/denom, 0.0, 1.0)
	}

	// Compute point on segment 2 closest to p1 + f1 * d1
	f2 := (float32(d12*f1) + rd2) / dd2

	// Clamping of segment 2 requires a do over on segment 1
	if f2 < 0.0 {
		f2 = 0.0
		f1 = clampFloat(-rd1/dd1, 0.0, 1.0)
	} else if f2 > 1.0 {
		f2 = 1.0
		f1 = clampFloat((d12-rd1)/dd1, 0.0, 1.0)
	}

	var (
		closest1        = p1.MulAdd(f1, d1)
		closest2        = p2.MulAdd(f2, d2)
		distanceSquared = closest1.DistanceSquared(closest2)
	)

	var (
		manifold    = LocalManifold{}
		radiusA     = capsuleA.Radius
		radiusB     = capsuleB.Radius
		radius      = radiusA + radiusB
		maxDistance = radius + SpeculativeDistance()
	)

	if distanceSquared > maxDistance*maxDistance {
		return manifold
	}

	distance := float32(math.Sqrt(float64(distanceSquared)))

	length1, u1 := d1.LengthAndNormalize()
	length2, u2 := d2.LengthAndNormalize()

	// Does segment B project outside segment A?
	var (
		fp2      = p2.Sub(p1).Dot(u1)
		fq2      = q2.Sub(p1).Dot(u1)
		outsideA = (fp2 <= 0.0 && fq2 <= 0.0) || (fp2 >= length1 && fq2 >= length1)
	)

	// Does segment A project outside segment B?
	var (
		fp1      = p1.Sub(p2).Dot(u2)
		fq1      = q1.Sub(p2).Dot(u2)
		outsideB = (fp1 <= 0.0 && fq1 <= 0.0) || (fp1 >= length2 && fq1 >= length2)
	)

	if !outsideA && !outsideB {
		// attempt to clip
		// this may yield contact points with excessive separation
		// in that case the algorithm falls back to single point collision

		// find reference edge using SAT
		normalA, separationA := capsuleAxisSeparation(p1, u1, p2, q2)
		normalB, separationB := capsuleAxisSeparation(p2, u2, p1, q1)

		// biased to avoid feature flip-flop
		// todo more testing?
		if separationA+float32(0.1*LinearSlop()) >= separationB {
			manifold.Normal = normalA

			var (
				cp = p2
				cq = q2
			)

			// clip to p1
			if fp2 < 0.0 && fq2 > 0.0 {
				cp = p2.Lerp(q2, (0.0-fp2)/(fq2-fp2))
			} else if fq2 < 0.0 && fp2 > 0.0 {
				cq = q2.Lerp(p2, (0.0-fq2)/(fp2-fq2))
			}

			// clip to q1
			if fp2 > length1 && fq2 < length1 {
				cp = p2.Lerp(q2, (fp2-length1)/(fp2-fq2))
			} else if fq2 > length1 && fp2 < length1 {
				cq = q2.Lerp(p2, (fq2-length1)/(fq2-fp2))
			}

			var (
				sp = cp.Sub(p1).Dot(normalA)
				sq = cq.Sub(p1).Dot(normalA)
			)

			if sp <= distance+LinearSlop() || sq <= distance+LinearSlop() {
				var mp *LocalManifoldPoint

				mp = &manifold.Points[0]
				mp.Point = cp.MulAdd(0.5*(radiusA-radiusB-sp), normalA)
				mp.Separation = sp - radius
				mp.ID = (uint16(uint8(0))<<8 | uint16(uint8(0)))

				mp = &manifold.Points[1]
				mp.Point = cq.MulAdd(0.5*(radiusA-radiusB-sq), normalA)
				mp.Separation = sq - radius
				mp.ID = (uint16(uint8(0))<<8 | uint16(uint8(1)))
				manifold.PointCount = 2
			}
		} else {
			// normal always points from A to B
			manifold.Normal = normalB.Neg()

			var (
				cp = p1
				cq = q1
			)

			// clip to p2
			if fp1 < 0.0 && fq1 > 0.0 {
				cp = p1.Lerp(q1, (0.0-fp1)/(fq1-fp1))
			} else if fq1 < 0.0 && fp1 > 0.0 {
				cq = q1.Lerp(p1, (0.0-fq1)/(fp1-fq1))
			}

			// clip to q2
			if fp1 > length2 && fq1 < length2 {
				cp = p1.Lerp(q1, (fp1-length2)/(fp1-fq1))
			} else if fq1 > length2 && fp1 < length2 {
				cq = q1.Lerp(p1, (fq1-length2)/(fq1-fp1))
			}

			var (
				sp = cp.Sub(p2).Dot(normalB)
				sq = cq.Sub(p2).Dot(normalB)
			)

			if sp <= distance+LinearSlop() || sq <= distance+LinearSlop() {
				var mp *LocalManifoldPoint

				mp = &manifold.Points[0]
				mp.Point = cp.MulAdd(0.5*(radiusB-radiusA-sp), normalB)
				mp.Separation = sp - radius
				mp.ID = (uint16(uint8(0))<<8 | uint16(uint8(0)))
				mp = &manifold.Points[1]
				mp.Point = cq.MulAdd(0.5*(radiusB-radiusA-sq), normalB)
				mp.Separation = sq - radius
				mp.ID = (uint16(uint8(1))<<8 | uint16(uint8(0)))
				manifold.PointCount = 2
			}
		}
	}

	if manifold.PointCount == 0 {
		// single point collision
		normal := closest2.Sub(closest1)
		if normal.Dot(normal) > epsSqr {
			normal = normal.Normalize()
		} else {
			normal = u1.LeftPerp()
		}

		var (
			c1 = closest1.MulAdd(radiusA, normal)
			c2 = closest2.MulAdd(-radiusB, normal)
		)

		var i1 int
		if f1 == 0.0 {
			i1 = 0
		} else {
			i1 = 1
		}

		var i2 int
		if f2 == 0.0 {
			i2 = 0
		} else {
			i2 = 1
		}

		manifold.Normal = normal
		manifold.Points[0].Point = c1.Lerp(c2, 0.5)
		manifold.Points[0].Separation = float32(math.Sqrt(float64(distanceSquared))) - radius
		manifold.Points[0].ID = (uint16(uint8(i1))<<8 | uint16(uint8(i2)))
		manifold.PointCount = 1
	}

	// Undo the origin shift so points are in frame A
	for i := range manifold.PointCount {
		manifold.Points[i].Point = manifold.Points[i].Point.Add(origin)
	}

	return manifold
}

// Compute the contact manifold between an segment and a capsule.
func CollideSegmentAndCapsule(segmentA *Segment, capsuleB *Capsule, xf Transform) LocalManifold {
	capsuleA := Capsule{segmentA.Point1, segmentA.Point2, 0.0}
	return CollideCapsules(&capsuleA, capsuleB, xf)
}

// Compute the contact manifold between a polygon and capsule.
func CollidePolygonAndCapsule(polygonA *Polygon, capsuleB *Capsule, xf Transform) LocalManifold {
	polyB := makeCapsule(capsuleB.Center1, capsuleB.Center2, capsuleB.Radius)
	return CollidePolygons(polygonA, &polyB, xf)
}

// Compute the contact manifold between two polygons.
func CollidePolygons(polygonA, polygonB *Polygon, xf Transform) LocalManifold {
	var (
		origin              = polygonA.Vertices[0]
		linearSlop          = LinearSlop()
		speculativeDistance = SpeculativeDistance()
	)

	// Shift to the origin in frame A for round-off, a pure translation in A's frame
	xfs := Transform{xf.P.Sub(origin), xf.Q}

	var localPolyA Polygon

	localPolyA.Count = polygonA.Count
	localPolyA.Radius = polygonA.Radius
	localPolyA.Vertices[0] = Vec2{}

	localPolyA.Normals[0] = polygonA.Normals[0]
	for i := 1; i < localPolyA.Count; i++ {
		localPolyA.Vertices[i] = polygonA.Vertices[i].Sub(origin)
		localPolyA.Normals[i] = polygonA.Normals[i]
	}

	// Put polyB in polyA's frame to reduce round-off error
	var localPolyB Polygon

	localPolyB.Count = polygonB.Count

	localPolyB.Radius = polygonB.Radius
	for i := range localPolyB.Count {
		localPolyB.Vertices[i] = xfs.Apply(polygonB.Vertices[i])
		localPolyB.Normals[i] = xfs.Q.Rotate(polygonB.Normals[i])
	}

	radius := localPolyA.Radius + localPolyB.Radius

	var (
		edgeA       = 0
		separationA = findMaxSeparation(&edgeA, &localPolyA, &localPolyB)
	)
	if separationA > speculativeDistance+radius {
		return LocalManifold{}
	}

	var (
		edgeB       = 0
		separationB = findMaxSeparation(&edgeB, &localPolyB, &localPolyA)
	)
	if separationB > speculativeDistance+radius {
		return LocalManifold{}
	}

	// Find incident edge
	var flip bool
	if separationA >= separationB {
		flip = false

		// Find the incident edge on polyB
		edgeB = findIncidentPolygonEdge(localPolyB.Normals[:localPolyB.Count], localPolyA.Normals[edgeA])
	} else {
		flip = true

		// Find the incident edge on polyA
		edgeA = findIncidentPolygonEdge(localPolyA.Normals[:localPolyA.Count], localPolyB.Normals[edgeB])
	}

	manifold := LocalManifold{}

	// Using slop here to ensure vertex-vertex normal vectors can be safely normalized
	// todo this means edge clipping needs to handle slightly non-overlapping edges.
	if separationA > float32(0.1*linearSlop) || separationB > float32(0.1*linearSlop) {
		// Edges are disjoint. Find closest points between reference edge and incident edge
		// Reference edge on polygon A
		var (
			i11 = edgeA
			i12 int
		)
		if edgeA+1 < localPolyA.Count {
			i12 = edgeA + 1
		} else {
			i12 = 0
		}

		var (
			i21 = edgeB
			i22 int
		)
		if edgeB+1 < localPolyB.Count {
			i22 = edgeB + 1
		} else {
			i22 = 0
		}

		var (
			v11 = localPolyA.Vertices[i11]
			v12 = localPolyA.Vertices[i12]
			v21 = localPolyB.Vertices[i21]
			v22 = localPolyB.Vertices[i22]
		)

		result := SegmentDistance(v11, v12, v21, v22)
		core.Assert(result.DistanceSquared > 0.0, "result.distanceSquared > 0.0f")

		var (
			distance   = float32(math.Sqrt(float64(result.DistanceSquared)))
			separation = distance - radius
		)

		if distance-radius > speculativeDistance {
			// This can happen in the vertex-vertex case
			return manifold
		}

		// Attempt to clip edges
		manifold = clipPolygons(&localPolyA, &localPolyB, edgeA, edgeB, flip)

		minSeparation := float32(0x1.fffffep127)
		for i := range manifold.PointCount {
			minSeparation = minFloat(minSeparation, manifold.Points[i].Separation)
		}

		// Does vertex-vertex have substantially larger separation?
		if separation+float32(0.1*linearSlop) < minSeparation {
			if result.Fraction1 == 0.0 && result.Fraction2 == 0.0 {
				// v11 - v21
				manifold.setVertexContact(v11, v21, distance, separation, localPolyA.Radius, localPolyB.Radius, uint16(uint8(i11))<<8|uint16(uint8(i21)))
			} else if result.Fraction1 == 0.0 && result.Fraction2 == 1.0 {
				// v11 - v22
				manifold.setVertexContact(v11, v22, distance, separation, localPolyA.Radius, localPolyB.Radius, uint16(uint8(i11))<<8|uint16(uint8(i22)))
			} else if result.Fraction1 == 1.0 && result.Fraction2 == 0.0 {
				// v12 - v21
				manifold.setVertexContact(v12, v21, distance, separation, localPolyA.Radius, localPolyB.Radius, uint16(uint8(i12))<<8|uint16(uint8(i21)))
			} else if result.Fraction1 == 1.0 && result.Fraction2 == 1.0 {
				// v12 - v22
				manifold.setVertexContact(v12, v22, distance, separation, localPolyA.Radius, localPolyB.Radius, uint16(uint8(i12))<<8|uint16(uint8(i22)))
			}
		}
	} else {
		// Polygons overlap
		manifold = clipPolygons(&localPolyA, &localPolyB, edgeA, edgeB, flip)
	}

	// Undo the origin shift so points are in frame A
	for i := range manifold.PointCount {
		manifold.Points[i].Point = manifold.Points[i].Point.Add(origin)
	}

	return manifold
}

// Compute the contact manifold between an segment and a circle.
func CollideSegmentAndCircle(segmentA *Segment, circleB *Circle, xf Transform) LocalManifold {
	capsuleA := Capsule{segmentA.Point1, segmentA.Point2, 0.0}
	return CollideCapsuleAndCircle(&capsuleA, circleB, xf)
}

// Compute the contact manifold between an segment and a polygon.
func CollideSegmentAndPolygon(segmentA *Segment, polygonB *Polygon, xf Transform) LocalManifold {
	polygonA := makeCapsule(segmentA.Point1, segmentA.Point2, 0.0)
	return CollidePolygons(&polygonA, polygonB, xf)
}

// Compute the contact manifold between a chain segment and a circle.
func CollideChainSegmentAndCircle(segmentA *ChainSegment, circleB *Circle, xf Transform) LocalManifold {
	manifold := LocalManifold{}

	// Compute circle in frame of segment
	pB := xf.Apply(circleB.Center)

	var (
		p1 = segmentA.Segment.Point1
		p2 = segmentA.Segment.Point2
		e  = p2.Sub(p1)
	)

	// Normal points to the right
	offset := e.RightPerp().Dot(pB.Sub(p1))
	if offset < 0.0 {
		// collision is one-sided
		return manifold
	}

	pA, owned := chainCircleClosestPoint(segmentA, e, pB)
	if !owned {
		return manifold
	}

	distance, normal := pB.Sub(pA).LengthAndNormalize()

	var (
		radius     = circleB.Radius
		separation = distance - radius
	)
	if separation > SpeculativeDistance() {
		return manifold
	}

	var (
		cA = pA
		cB = pB.MulAdd(-radius, normal)
	)

	manifold.Normal = normal

	mp := &manifold.Points[0]

	mp.Point = cA.Lerp(cB, 0.5)
	mp.Separation = separation
	mp.ID = 0
	manifold.PointCount = 1

	return manifold
}

// Compute the contact manifold between a chain segment and a capsule.
func CollideChainSegmentAndCapsule(segmentA *ChainSegment, capsuleB *Capsule, xf Transform, cache *SimplexCache) LocalManifold {
	polyB := makeCapsule(capsuleB.Center1, capsuleB.Center2, capsuleB.Radius)
	return CollideChainSegmentAndPolygon(segmentA, &polyB, xf, cache)
}

// Compute the contact manifold between a chain segment and a rounded polygon.
func CollideChainSegmentAndPolygon(segmentA *ChainSegment, polygonB *Polygon, xf Transform, cache *SimplexCache) LocalManifold {
	manifold := LocalManifold{}

	var (
		centroidB = xf.Apply(polygonB.Centroid)
		radiusB   = polygonB.Radius
	)

	var (
		p1 = segmentA.Segment.Point1
		p2 = segmentA.Segment.Point2
	)

	edge1 := p2.Sub(p1).Normalize()

	smoothParams := chainSegmentParams{}

	smoothParams.edge1 = edge1

	var (
		convexTol float32 = 0.01
		edge0             = p1.Sub(segmentA.Ghost1).Normalize()
	)

	smoothParams.normal0 = edge0.RightPerp()
	smoothParams.convex1 = edge0.Cross(edge1) >= convexTol

	edge2 := segmentA.Ghost2.Sub(p2).Normalize()

	smoothParams.normal2 = edge2.RightPerp()
	smoothParams.convex2 = edge1.Cross(edge2) >= convexTol

	// Normal points to the right
	var (
		normal1 = edge1.RightPerp()
		behind1 = normal1.Dot(centroidB.Sub(p1)) < 0.0
		behind0 = true
		behind2 = true
	)

	if smoothParams.convex1 {
		behind0 = smoothParams.normal0.Dot(centroidB.Sub(p1)) < 0.0
	}

	if smoothParams.convex2 {
		behind2 = smoothParams.normal2.Dot(centroidB.Sub(p2)) < 0.0
	}

	if behind1 && behind0 && behind2 {
		// one-sided collision
		return manifold
	}

	// Get polygonB in frameA
	var (
		count    = polygonB.Count
		vertices [MaxPolygonVertices]Vec2
		normals  [MaxPolygonVertices]Vec2
	)

	for i := range count {
		vertices[i] = xf.Apply(polygonB.Vertices[i])
		normals[i] = xf.Q.Rotate(polygonB.Normals[i])
	}

	// Distance doesn't work correctly with partial polygons
	var input DistanceInput

	input.ProxyA = MakeProxy([]Vec2{segmentA.Segment.Point1, segmentA.Segment.Point2}, 0.0)
	input.ProxyB = MakeProxy(vertices[:count], 0.0)
	input.Transform = TransformIdentity
	input.UseRadii = false

	output := ShapeDistance(&input, cache, nil)

	if output.Distance > radiusB+SpeculativeDistance() {
		return manifold
	}

	// Snap concave normals for partial polygon
	var n0 Vec2
	if smoothParams.convex1 {
		n0 = smoothParams.normal0
	} else {
		n0 = normal1
	}

	var n2 Vec2
	if smoothParams.convex2 {
		n2 = smoothParams.normal2
	} else {
		n2 = normal1
	}

	// Index of incident vertex on polygon
	var (
		incidentIndex  = -1
		incidentNormal = -1
	)

	if !behind1 && output.Distance > float32(0.1*LinearSlop()) {
		// The closest features may be two vertices or an edge and a vertex even when there should
		// be face contact
		if cache.Count == 1 {
			// vertex-vertex collision
			var (
				pA = output.PointA
				pB = output.PointB
			)

			normal := pB.Sub(pA).Normalize()

			kind := classifyNormal(smoothParams, normal)
			if kind == normalSkip {
				return manifold
			}

			if kind == normalAdmit {
				manifold.Normal = normal

				cp := &manifold.Points[0]

				cp.Point = pA
				cp.Separation = output.Distance - radiusB
				cp.ID = (uint16(uint8(cache.IndexA[0]))<<8 | uint16(uint8(cache.IndexB[0])))
				manifold.PointCount = 1

				return manifold
			}

			// fall through b2_normalSnap
			incidentIndex = int(cache.IndexB[0])
		} else {
			// vertex-edge collision
			core.Assert(cache.Count == 2, "cache->count == 2")

			var (
				ia1 = int(cache.IndexA[0])
				ia2 = int(cache.IndexA[1])
				ib1 = int(cache.IndexB[0])
				ib2 = int(cache.IndexB[1])
			)

			if ia1 == ia2 {
				// 1 point on A, expect 2 points on B
				core.Assert(ib1 != ib2, "ib1 != ib2")

				// Find polygon normal most aligned with vector between closest points.
				// This effectively sorts ib1 and ib2
				var (
					normalB = output.PointA.Sub(output.PointB)
					dot1    = normalB.Dot(normals[ib1])
					dot2    = normalB.Dot(normals[ib2])
					ib      int
				)
				if dot1 > dot2 {
					ib = ib1
				} else {
					ib = ib2
				}

				// Use accurate normal
				normalB = normals[ib]

				kind := classifyNormal(smoothParams, normalB.Neg())
				if kind == normalSkip {
					return manifold
				}

				if kind == normalAdmit {
					// Get polygon edge associated with normal
					ib1 = ib
					if ib < count-1 {
						ib2 = ib + 1
					} else {
						ib2 = 0
					}

					var (
						b1 = vertices[ib1]
						b2 = vertices[ib2]
					)

					// Find incident segment vertex
					dot1 = normalB.Dot(p1.Sub(b1))
					dot2 = normalB.Dot(p2.Sub(b1))

					if dot1 < dot2 {
						if n0.Dot(normalB) < normal1.Dot(normalB) {
							// Neighbor is incident
							return manifold
						}
					} else {
						if n2.Dot(normalB) < normal1.Dot(normalB) {
							// Neighbor is incident
							return manifold
						}
					}

					manifold = clipSegments(b1, b2, p1, p2, normalB, radiusB, 0.0, (uint16(uint8(ib1))<<8 | uint16(uint8(1))), (uint16(uint8(ib2))<<8 | uint16(uint8(0))))

					core.Assert(manifold.PointCount == 0 || manifold.PointCount == 2, "manifold.pointCount == 0 || manifold.pointCount == 2")

					if manifold.PointCount == 2 {
						manifold.Normal = normalB.Neg()
					}

					return manifold
				}

				// fall through b2_normalSnap
				incidentNormal = ib
			} else {
				// Get index of incident polygonB vertex
				var (
					dot1 = normal1.Dot(vertices[ib1].Sub(p1))
					dot2 = normal1.Dot(vertices[ib2].Sub(p2))
				)
				if dot1 < dot2 {
					incidentIndex = ib1
				} else {
					incidentIndex = ib2
				}
			}
		}
	} else {
		// SAT edge normal
		edgeSeparation, incident := minimumNormalSeparation(vertices[:count], p1, normal1)
		incidentIndex = incident

		// Check convex neighbor for edge separation
		if smoothParams.convex1 {
			s0, _ := minimumNormalSeparation(vertices[:count], p1, smoothParams.normal0)

			if s0 > edgeSeparation {
				edgeSeparation = s0

				// Indicate neighbor owns edge separation
				incidentIndex = -1
			}
		}

		// Check convex neighbor for edge separation
		if smoothParams.convex2 {
			s2, _ := minimumNormalSeparation(vertices[:count], p2, smoothParams.normal2)

			if s2 > edgeSeparation {
				edgeSeparation = s2

				// Indicate neighbor owns edge separation
				incidentIndex = -1
			}
		}

		// SAT polygon normals
		var (
			polygonSeparation = -float32(0x1.fffffep127)
			referenceIndex    = -1
		)

		for i := range count {
			n := normals[i]

			kind := classifyNormal(smoothParams, n.Neg())
			if kind != normalAdmit {
				continue
			}

			// Check the infinite sides of the partial polygon
			// if ((smoothParams.convex1 && b2Cross(n0, n) > 0.0f) || (smoothParams.convex2 && b2Cross(n, n2) > 0.0f))
			//{
			//	continue;
			//}

			var (
				p = vertices[i]
				s = minFloat(n.Dot(p2.Sub(p)), n.Dot(p1.Sub(p)))
			)

			if s > polygonSeparation {
				polygonSeparation = s
				referenceIndex = i
			}
		}

		if polygonSeparation > edgeSeparation {
			var (
				ia1 = referenceIndex
				ia2 int
			)
			if ia1 < count-1 {
				ia2 = ia1 + 1
			} else {
				ia2 = 0
			}

			var (
				a1 = vertices[ia1]
				a2 = vertices[ia2]
			)

			n := normals[ia1]

			var (
				dot1 = n.Dot(p1.Sub(a1))
				dot2 = n.Dot(p2.Sub(a1))
			)

			if dot1 < dot2 {
				if n0.Dot(n) < normal1.Dot(n) {
					// Neighbor is incident
					return manifold
				}
			} else {
				if n2.Dot(n) < normal1.Dot(n) {
					// Neighbor is incident
					return manifold
				}
			}

			manifold = clipSegments(a1, a2, p1, p2, normals[ia1], radiusB, 0.0, (uint16(uint8(ia1))<<8 | uint16(uint8(1))), (uint16(uint8(ia2))<<8 | uint16(uint8(0))))

			core.Assert(manifold.PointCount == 0 || manifold.PointCount == 2, "manifold.pointCount == 0 || manifold.pointCount == 2")

			if manifold.PointCount == 2 {
				manifold.Normal = normals[ia1].Neg()
			}

			return manifold
		}

		if incidentIndex == -1 {
			// neighboring segment is the separating axis
			return manifold
		}

		// fall through segment normal axis
	}

	core.Assert(incidentNormal != -1 || incidentIndex != -1, "incidentNormal != -1 || incidentIndex != -1")

	// Segment normal

	// Find incident polygon normal: normal adjacent to deepest vertex that is most anti-parallel to segment normal
	var (
		b1, b2   Vec2
		ib1, ib2 int
	)

	if incidentNormal != -1 {
		ib1 = incidentNormal
		if ib1 < count-1 {
			ib2 = ib1 + 1
		} else {
			ib2 = 0
		}

		b1 = vertices[ib1]
		b2 = vertices[ib2]
	} else {
		var (
			i2 = incidentIndex
			i1 int
		)
		if i2 > 0 {
			i1 = i2 - 1
		} else {
			i1 = count - 1
		}

		var (
			d1 = normal1.Dot(normals[i1])
			d2 = normal1.Dot(normals[i2])
		)
		if d1 < d2 {
			ib1 = i1
			ib2 = i2
			b1 = vertices[ib1]
			b2 = vertices[ib2]
		} else {
			ib1 = i2
			if i2 < count-1 {
				ib2 = i2 + 1
			} else {
				ib2 = 0
			}

			b1 = vertices[ib1]
			b2 = vertices[ib2]
		}
	}

	manifold = clipSegments(p1, p2, b1, b2, normal1, 0.0, radiusB, (uint16(uint8(0))<<8 | uint16(uint8(ib2))), (uint16(uint8(1))<<8 | uint16(uint8(ib1))))

	core.Assert(manifold.PointCount == 0 || manifold.PointCount == 2, "manifold.pointCount == 0 || manifold.pointCount == 2")

	return manifold
}

func makeCapsule(p1, p2 Vec2, radius float32) Polygon {
	shape := Polygon{}

	shape.Vertices[0] = p1
	shape.Vertices[1] = p2
	shape.Centroid = p1.Lerp(p2, 0.5)

	d := p2.Sub(p1)
	core.Assert(d.LengthSquared() > float32(0x1p-23), "b2LengthSquared( d ) > FLT_EPSILON")

	var (
		axis   = d.Normalize()
		normal = axis.RightPerp()
	)

	shape.Normals[0] = normal
	shape.Normals[1] = normal.Neg()
	shape.Count = 2
	shape.Radius = radius

	return shape
}

func clipPolygons(polyA, polyB *Polygon, edgeA, edgeB int, flip bool) LocalManifold {
	manifold := LocalManifold{}

	// reference polygon
	var (
		poly1    *Polygon
		i11, i12 int
	)

	// incident polygon
	var (
		poly2    *Polygon
		i21, i22 int
	)

	if flip {
		poly1 = polyB
		poly2 = polyA

		i11 = edgeB
		i21 = edgeA
	} else {
		poly1 = polyA
		poly2 = polyB

		i11 = edgeA
		i21 = edgeB
	}

	i12 = i11 + 1
	if i12 >= poly1.Count {
		i12 = 0
	}

	i22 = i21 + 1
	if i22 >= poly2.Count {
		i22 = 0
	}

	normal := poly1.Normals[i11]

	// Reference edge vertices
	var (
		v11 = poly1.Vertices[i11]
		v12 = poly1.Vertices[i12]
	)

	// Incident edge vertices
	var (
		v21 = poly2.Vertices[i21]
		v22 = poly2.Vertices[i22]
	)

	tangent := normal.ScalarCross(1.0)

	var (
		lower1 float32 = 0.0
		upper1         = v12.Sub(v11).Dot(tangent)
	)

	// Incident edge points opposite of tangent due to CCW winding
	var (
		upper2 = v21.Sub(v11).Dot(tangent)
		lower2 = v22.Sub(v11).Dot(tangent)
	)

	// Are the segments disjoint?
	if upper2 < lower1 || upper1 < lower2 {
		return manifold
	}

	var vLower Vec2
	if lower2 < lower1 && upper2-lower2 > float32(0x1p-23) {
		vLower = v22.Lerp(v21, (lower1-lower2)/(upper2-lower2))
	} else {
		vLower = v22
	}

	var vUpper Vec2
	if upper2 > upper1 && upper2-lower2 > float32(0x1p-23) {
		vUpper = v22.Lerp(v21, (upper1-lower2)/(upper2-lower2))
	} else {
		vUpper = v21
	}

	// todo vLower can be very close to vUpper, reduce to one point?

	var (
		separationLower = vLower.Sub(v11).Dot(normal)
		separationUpper = vUpper.Sub(v11).Dot(normal)
	)

	var (
		r1 = poly1.Radius
		r2 = poly2.Radius
	)

	// Put contact points at midpoint, accounting for radii
	vLower = vLower.MulAdd(0.5*(r1-r2-separationLower), normal)
	vUpper = vUpper.MulAdd(0.5*(r1-r2-separationUpper), normal)

	radius := r1 + r2

	if !flip {
		manifold.Normal = normal

		cp := &manifold.Points[0]

		{
			cp.Point = vLower
			cp.Separation = separationLower - radius
			cp.ID = (uint16(uint8(i11))<<8 | uint16(uint8(i22)))
			manifold.PointCount += 1
			cp = &manifold.Points[1]
		}

		{
			cp.Point = vUpper
			cp.Separation = separationUpper - radius
			cp.ID = (uint16(uint8(i12))<<8 | uint16(uint8(i21)))
			manifold.PointCount += 1
		}
	} else {
		manifold.Normal = normal.Neg()

		cp := &manifold.Points[0]

		{
			cp.Point = vUpper
			cp.Separation = separationUpper - radius
			cp.ID = (uint16(uint8(i21))<<8 | uint16(uint8(i12)))
			manifold.PointCount += 1
			cp = &manifold.Points[1]
		}

		{
			cp.Point = vLower
			cp.Separation = separationLower - radius
			cp.ID = (uint16(uint8(i22))<<8 | uint16(uint8(i11)))
			manifold.PointCount += 1
		}
	}

	return manifold
}

func findMaxSeparation(edgeIndex *int, poly1, poly2 *Polygon) float32 {
	var (
		count1 = poly1.Count
		count2 = poly2.Count
		n1s    = poly1.Normals[:]
		v1s    = poly1.Vertices[:]
		v2s    = poly2.Vertices[:]
	)

	var (
		bestIndex     = 0
		maxSeparation = -float32(0x1.fffffep127)
	)

	for i := range count1 {
		// Get poly1 normal in frame2.
		var (
			n  = n1s[i]
			v1 = v1s[i]
		)

		// Find the deepest point for normal i.
		si := float32(0x1.fffffep127)

		for j := range count2 {
			sij := n.Dot(v2s[j].Sub(v1))
			if sij < si {
				si = sij
			}
		}

		if si > maxSeparation {
			maxSeparation = si
			bestIndex = i
		}
	}

	*edgeIndex = bestIndex

	return maxSeparation
}

func clipSegments(a1, a2, b1, b2, normal Vec2, ra, rb float32, id1, id2 uint16) LocalManifold {
	manifold := LocalManifold{}

	tangent := normal.LeftPerp()

	// Barycentric coordinates of each point relative to a1 along tangent
	var (
		lower1 float32 = 0.0
		upper1         = a2.Sub(a1).Dot(tangent)
	)

	// Incident edge points opposite of tangent due to CCW winding
	var (
		upper2 = b1.Sub(a1).Dot(tangent)
		lower2 = b2.Sub(a1).Dot(tangent)
	)

	// Do segments overlap?
	if upper2 < lower1 || upper1 < lower2 {
		return manifold
	}

	var vLower Vec2
	if lower2 < lower1 && upper2-lower2 > float32(0x1p-23) {
		vLower = b2.Lerp(b1, (lower1-lower2)/(upper2-lower2))
	} else {
		vLower = b2
	}

	var vUpper Vec2
	if upper2 > upper1 && upper2-lower2 > float32(0x1p-23) {
		vUpper = b2.Lerp(b1, (upper1-lower2)/(upper2-lower2))
	} else {
		vUpper = b1
	}

	// todo vLower can be very close to vUpper, reduce to one point?

	var (
		separationLower = vLower.Sub(a1).Dot(normal)
		separationUpper = vUpper.Sub(a1).Dot(normal)
	)

	// Put contact points at midpoint, accounting for radii
	vLower = vLower.MulAdd(0.5*(ra-rb-separationLower), normal)
	vUpper = vUpper.MulAdd(0.5*(ra-rb-separationUpper), normal)

	radius := ra + rb

	manifold.Normal = normal
	{
		cp := &manifold.Points[0]

		cp.Point = vLower
		cp.Separation = separationLower - radius
		cp.ID = id1
	}

	{
		cp := &manifold.Points[1]

		cp.Point = vUpper
		cp.Separation = separationUpper - radius
		cp.ID = id2
	}

	manifold.PointCount = 2

	return manifold
}

func classifyNormal(params chainSegmentParams, normal Vec2) normalType {
	var sinTol float32 = 0.01

	if normal.Dot(params.edge1) <= 0.0 {
		// Normal points towards the segment tail
		if !params.convex1 {
			return normalSnap
		}

		if normal.Cross(params.normal0) > sinTol {
			return normalSkip
		}

		return normalAdmit
	}
	// Normal points towards segment head
	if !params.convex2 {
		return normalSnap
	}

	if params.normal2.Cross(normal) > sinTol {
		return normalSkip
	}

	return normalAdmit
}

func capsuleAxisSeparation(origin, edgeDirection, point1, point2 Vec2) (Vec2, float32) {
	normal := edgeDirection.LeftPerp()
	ss1 := point1.Sub(origin).Dot(normal)
	ss2 := point2.Sub(origin).Dot(normal)

	var positiveSeparation float32
	if ss1 < ss2 {
		positiveSeparation = ss1
	} else {
		positiveSeparation = ss2
	}

	var negativeSeparation float32
	if -ss1 < -ss2 {
		negativeSeparation = -ss1
	} else {
		negativeSeparation = -ss2
	}

	if positiveSeparation > negativeSeparation {
		return normal, positiveSeparation
	}

	return normal.Neg(), negativeSeparation
}

func (manifold *LocalManifold) setVertexContact(pointA, pointB Vec2, distance, separation, radiusA, radiusB float32, id uint16) {
	normal := pointB.Sub(pointA)
	invDistance := 1.0 / distance
	normal.X *= invDistance
	normal.Y *= invDistance
	c1 := pointA.MulAdd(radiusA, normal)
	c2 := pointB.MulAdd(-radiusB, normal)
	manifold.Normal = normal
	manifold.Points[0].Point = c1.Lerp(c2, 0.5)
	manifold.Points[0].Separation = separation
	manifold.Points[0].ID = id
	manifold.PointCount = 1
}

func findIncidentPolygonEdge(normals []Vec2, searchDirection Vec2) int {
	edge := 0
	minDot := float32(0x1.fffffep127)

	for i := range normals {
		dot := searchDirection.Dot(normals[i])
		if dot < minDot {
			minDot = dot
			edge = i
		}
	}

	return edge
}

func minimumNormalSeparation(vertices []Vec2, origin, normal Vec2) (float32, int) {
	separation := float32(0x1.fffffep127)
	index := -1

	for i := range vertices {
		s := normal.Dot(vertices[i].Sub(origin))
		if s < separation {
			separation = s
			index = i
		}
	}

	return separation, index
}

func chainCircleClosestPoint(segment *ChainSegment, edge, point Vec2) (Vec2, bool) {
	p1 := segment.Segment.Point1
	p2 := segment.Segment.Point2
	// Barycentric coordinates
	u := edge.Dot(p2.Sub(point))

	v := edge.Dot(point.Sub(p1))
	if v <= 0.0 {
		// Behind point1? Is the point in the previous edge's Voronoi region?
		prevEdge := p1.Sub(segment.Ghost1)

		uPrev := prevEdge.Dot(point.Sub(p1))
		if uPrev <= 0.0 {
			return Vec2{}, false
		}

		return p1, true
	}

	if u <= 0.0 {
		// Ahead of point2? Is the point in the next edge's Voronoi region?
		nextEdge := segment.Ghost2.Sub(p2)

		vNext := nextEdge.Dot(point.Sub(p2))
		if vNext > 0.0 {
			return Vec2{}, false
		}

		return p2, true
	}

	ee := edge.Dot(edge)

	pA := Vec2{float32(u*p1.X) + float32(v*p2.X), float32(u*p1.Y) + float32(v*p2.Y)}
	if ee > 0.0 {
		return pA.Scale(1.0 / ee), true
	}

	return p1, true
}
