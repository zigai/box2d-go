// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

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
	var shape Polygon = Polygon{}
	shape.Vertices[0] = p1
	shape.Vertices[1] = p2
	shape.Centroid = Lerp(p1, p2, 0.5)

	var d Vec2 = Sub(p2, p1)
	assert(LengthSquared(d) > float32(0x1p-23), "b2LengthSquared( d ) > FLT_EPSILON")
	var axis Vec2 = Normalize(d)
	var normal Vec2 = RightPerp(axis)

	shape.Normals[0] = normal
	shape.Normals[1] = Neg(normal)
	shape.Count = 2
	shape.Radius = radius

	return shape
}

// Compute the contact manifold between two circles
func CollideCircles(circleA, circleB *Circle, xf Transform) LocalManifold {
	var manifold LocalManifold = LocalManifold{}

	var pointA Vec2 = circleA.Center
	var pointB Vec2 = TransformPoint(xf, circleB.Center)

	var distance float32
	var normal Vec2 = GetLengthAndNormalize(&distance, Sub(pointB, pointA))

	var radiusA float32 = circleA.Radius
	var radiusB float32 = circleB.Radius

	var separation float32 = distance - radiusA - radiusB
	if separation > SpeculativeDistance() {
		return manifold
	}

	var cA Vec2 = MulAdd(pointA, radiusA, normal)
	var cB Vec2 = MulAdd(pointB, -radiusB, normal)

	manifold.Normal = normal
	var mp *LocalManifoldPoint = &manifold.Points[0]
	mp.Point = Lerp(cA, cB, 0.5)
	mp.Separation = separation
	mp.ID = 0
	manifold.PointCount = 1
	return manifold
}

// Compute the contact manifold between a capsule and circle
func CollideCapsuleAndCircle(capsuleA *Capsule, circleB *Circle, xf Transform) LocalManifold {
	var manifold LocalManifold = LocalManifold{}

	// Compute circle position in the frame of the capsule.
	var pB Vec2 = TransformPoint(xf, circleB.Center)

	// Compute closest point
	var p1 Vec2 = capsuleA.Center1
	var p2 Vec2 = capsuleA.Center2

	var e Vec2 = Sub(p2, p1)

	// dot(p - pA, e) = 0
	// dot(p - (p1 + s1 * e), e) = 0
	// s1 = dot(p - p1, e)
	var pA Vec2
	var s1 float32 = Dot(Sub(pB, p1), e)
	var s2 float32 = Dot(Sub(p2, pB), e)
	if s1 < 0.0 {
		// p1 region
		pA = p1
	} else if s2 < 0.0 {
		// p2 region
		pA = p2
	} else {
		// circle colliding with segment interior
		var s float32 = s1 / Dot(e, e)
		pA = MulAdd(p1, s, e)
	}

	var distance float32
	var normal Vec2 = GetLengthAndNormalize(&distance, Sub(pB, pA))

	var radiusA float32 = capsuleA.Radius
	var radiusB float32 = circleB.Radius
	var separation float32 = distance - radiusA - radiusB
	if separation > SpeculativeDistance() {
		return manifold
	}

	var cA Vec2 = MulAdd(pA, radiusA, normal)
	var cB Vec2 = MulAdd(pB, -radiusB, normal)

	manifold.Normal = normal
	var mp *LocalManifoldPoint = &manifold.Points[0]
	mp.Point = Lerp(cA, cB, 0.5)
	mp.Separation = separation
	mp.ID = 0
	manifold.PointCount = 1
	return manifold
}

// Compute the contact manifold between a polygon and a circle
func CollidePolygonAndCircle(polygonA *Polygon, circleB *Circle, xf Transform) LocalManifold {
	var manifold LocalManifold = LocalManifold{}
	var speculativeDistance float32 = SpeculativeDistance()

	// Compute circle position in the frame of the polygon.
	var center Vec2 = TransformPoint(xf, circleB.Center)
	var radiusA float32 = polygonA.Radius
	var radiusB float32 = circleB.Radius
	var radius float32 = radiusA + radiusB

	// Find the min separating edge.
	var normalIndex int = 0
	var separation float32 = -float32(0x1.fffffep127)
	var vertexCount int = polygonA.Count
	var vertices []Vec2 = polygonA.Vertices[:]
	var normals []Vec2 = polygonA.Normals[:]

	for i := 0; i < vertexCount; i++ {
		var s float32 = Dot(normals[i], Sub(center, vertices[i]))
		if s > separation {
			separation = s
			normalIndex = i
		}
	}

	if separation > radius+speculativeDistance {
		return manifold
	}

	// Vertices of the reference edge.
	var vertIndex1 int = normalIndex
	var vertIndex2 int
	if vertIndex1+1 < vertexCount {
		vertIndex2 = vertIndex1 + 1
	} else {
		vertIndex2 = 0
	}
	var v1 Vec2 = vertices[vertIndex1]
	var v2 Vec2 = vertices[vertIndex2]

	// Compute barycentric coordinates
	var u1 float32 = Dot(Sub(center, v1), Sub(v2, v1))
	var u2 float32 = Dot(Sub(center, v2), Sub(v1, v2))

	if u1 < 0.0 && separation > float32(0x1p-23) {
		// Circle center is closest to v1 and safely outside the polygon
		var normal Vec2 = Normalize(Sub(center, v1))
		separation = Dot(Sub(center, v1), normal)
		if separation > radius+speculativeDistance {
			return manifold
		}

		var cA Vec2 = MulAdd(v1, radiusA, normal)
		var cB Vec2 = MulSub(center, radiusB, normal)
		var contactPointA Vec2 = Lerp(cA, cB, 0.5)

		manifold.Normal = normal
		var mp *LocalManifoldPoint = &manifold.Points[0]
		mp.Point = contactPointA
		mp.Separation = Dot(Sub(cB, cA), normal)
		mp.ID = 0
		manifold.PointCount = 1
	} else if u2 < 0.0 && separation > float32(0x1p-23) {
		// Circle center is closest to v2 and safely outside the polygon
		var normal Vec2 = Normalize(Sub(center, v2))
		separation = Dot(Sub(center, v2), normal)
		if separation > radius+speculativeDistance {
			return manifold
		}

		var cA Vec2 = MulAdd(v2, radiusA, normal)
		var cB Vec2 = MulSub(center, radiusB, normal)
		var contactPointA Vec2 = Lerp(cA, cB, 0.5)

		manifold.Normal = normal
		var mp *LocalManifoldPoint = &manifold.Points[0]
		mp.Point = contactPointA
		mp.Separation = Dot(Sub(cB, cA), normal)
		mp.ID = 0
		manifold.PointCount = 1
	} else {
		// Circle center is between v1 and v2. Center may be inside polygon
		var normal Vec2 = normals[normalIndex]
		manifold.Normal = normal

		// cA is the projection of the circle center onto to the reference edge
		var cA Vec2 = MulAdd(center, radiusA-Dot(Sub(center, v1), normal), normal)

		// cB is the deepest point on the circle with respect to the reference edge
		var cB Vec2 = MulSub(center, radiusB, normal)

		// The contact point is the midpoint
		var mp *LocalManifoldPoint = &manifold.Points[0]
		mp.Point = Lerp(cA, cB, 0.5)
		mp.Separation = separation - radius
		mp.ID = 0
		manifold.PointCount = 1
	}

	return manifold
}

// Compute the contact manifold between a capsule and circle
func CollideCapsules(capsuleA, capsuleB *Capsule, xf Transform) LocalManifold {
	var origin Vec2 = capsuleA.Center1

	// Shift to the origin in frame A for round-off, a pure translation in A's frame
	var xfs Transform = Transform{Sub(xf.P, origin), xf.Q}

	var p1 Vec2 = Vec2{}
	var q1 Vec2 = Sub(capsuleA.Center2, origin)

	var p2 Vec2 = TransformPoint(xfs, capsuleB.Center1)
	var q2 Vec2 = TransformPoint(xfs, capsuleB.Center2)

	var d1 Vec2 = Sub(q1, p1)
	var d2 Vec2 = Sub(q2, p2)

	var dd1 float32 = Dot(d1, d1)
	var dd2 float32 = Dot(d2, d2)

	var epsSqr float32 = float32(0x1p-23) * float32(0x1p-23)
	assert(dd1 > epsSqr && dd2 > epsSqr, "dd1 > epsSqr && dd2 > epsSqr")

	var r Vec2 = Sub(p1, p2)
	var rd1 float32 = Dot(r, d1)
	var rd2 float32 = Dot(r, d2)

	var d12 float32 = Dot(d1, d2)

	var denom float32 = float32(dd1*dd2) - float32(d12*d12)

	// Fraction on segment 1
	var f1 float32 = 0.0
	if denom != 0.0 {
		// not parallel
		f1 = ClampFloat((float32(d12*rd2)-float32(rd1*dd2))/denom, 0.0, 1.0)
	}

	// Compute point on segment 2 closest to p1 + f1 * d1
	var f2 float32 = (float32(d12*f1) + rd2) / dd2

	// Clamping of segment 2 requires a do over on segment 1
	if f2 < 0.0 {
		f2 = 0.0
		f1 = ClampFloat(-rd1/dd1, 0.0, 1.0)
	} else if f2 > 1.0 {
		f2 = 1.0
		f1 = ClampFloat((d12-rd1)/dd1, 0.0, 1.0)
	}

	var closest1 Vec2 = MulAdd(p1, f1, d1)
	var closest2 Vec2 = MulAdd(p2, f2, d2)
	var distanceSquared float32 = DistanceSquared(closest1, closest2)

	var manifold LocalManifold = LocalManifold{}
	var radiusA float32 = capsuleA.Radius
	var radiusB float32 = capsuleB.Radius
	var radius float32 = radiusA + radiusB
	var maxDistance float32 = radius + SpeculativeDistance()

	if distanceSquared > maxDistance*maxDistance {
		return manifold
	}

	var distance float32 = float32(math.Sqrt(float64(distanceSquared)))

	var length1, length2 float32
	var u1 Vec2 = GetLengthAndNormalize(&length1, d1)
	var u2 Vec2 = GetLengthAndNormalize(&length2, d2)

	// Does segment B project outside segment A?
	var fp2 float32 = Dot(Sub(p2, p1), u1)
	var fq2 float32 = Dot(Sub(q2, p1), u1)
	var outsideA bool = (fp2 <= 0.0 && fq2 <= 0.0) || (fp2 >= length1 && fq2 >= length1)

	// Does segment A project outside segment B?
	var fp1 float32 = Dot(Sub(p1, p2), u2)
	var fq1 float32 = Dot(Sub(q1, p2), u2)
	var outsideB bool = (fp1 <= 0.0 && fq1 <= 0.0) || (fp1 >= length2 && fq1 >= length2)

	if outsideA == false && outsideB == false {
		// attempt to clip
		// this may yield contact points with excessive separation
		// in that case the algorithm falls back to single point collision

		// find reference edge using SAT
		var normalA Vec2
		var separationA float32

		{
			normalA = LeftPerp(u1)
			var ss1 float32 = Dot(Sub(p2, p1), normalA)
			var ss2 float32 = Dot(Sub(q2, p1), normalA)
			var s1p float32
			if ss1 < ss2 {
				s1p = ss1
			} else {
				s1p = ss2
			}
			var s1n float32
			if -ss1 < -ss2 {
				s1n = -ss1
			} else {
				s1n = -ss2
			}

			if s1p > s1n {
				separationA = s1p
			} else {
				separationA = s1n
				normalA = Neg(normalA)
			}
		}

		var normalB Vec2
		var separationB float32
		{
			normalB = LeftPerp(u2)
			var ss1 float32 = Dot(Sub(p1, p2), normalB)
			var ss2 float32 = Dot(Sub(q1, p2), normalB)
			var s1p float32
			if ss1 < ss2 {
				s1p = ss1
			} else {
				s1p = ss2
			}
			var s1n float32
			if -ss1 < -ss2 {
				s1n = -ss1
			} else {
				s1n = -ss2
			}

			if s1p > s1n {
				separationB = s1p
			} else {
				separationB = s1n
				normalB = Neg(normalB)
			}
		}

		// biased to avoid feature flip-flop
		// todo more testing?
		if separationA+float32(0.1*LinearSlop()) >= separationB {
			manifold.Normal = normalA

			var cp Vec2 = p2
			var cq Vec2 = q2

			// clip to p1
			if fp2 < 0.0 && fq2 > 0.0 {
				cp = Lerp(p2, q2, (0.0-fp2)/(fq2-fp2))
			} else if fq2 < 0.0 && fp2 > 0.0 {
				cq = Lerp(q2, p2, (0.0-fq2)/(fp2-fq2))
			}

			// clip to q1
			if fp2 > length1 && fq2 < length1 {
				cp = Lerp(p2, q2, (fp2-length1)/(fp2-fq2))
			} else if fq2 > length1 && fp2 < length1 {
				cq = Lerp(q2, p2, (fq2-length1)/(fq2-fp2))
			}

			var sp float32 = Dot(Sub(cp, p1), normalA)
			var sq float32 = Dot(Sub(cq, p1), normalA)

			if sp <= distance+LinearSlop() || sq <= distance+LinearSlop() {
				var mp *LocalManifoldPoint
				mp = &manifold.Points[0]
				mp.Point = MulAdd(cp, 0.5*(radiusA-radiusB-sp), normalA)
				mp.Separation = sp - radius
				mp.ID = (uint16(uint8(0))<<8 | uint16(uint8(0)))

				mp = &manifold.Points[1]
				mp.Point = MulAdd(cq, 0.5*(radiusA-radiusB-sq), normalA)
				mp.Separation = sq - radius
				mp.ID = (uint16(uint8(0))<<8 | uint16(uint8(1)))
				manifold.PointCount = 2
			}
		} else {
			// normal always points from A to B
			manifold.Normal = Neg(normalB)

			var cp Vec2 = p1
			var cq Vec2 = q1

			// clip to p2
			if fp1 < 0.0 && fq1 > 0.0 {
				cp = Lerp(p1, q1, (0.0-fp1)/(fq1-fp1))
			} else if fq1 < 0.0 && fp1 > 0.0 {
				cq = Lerp(q1, p1, (0.0-fq1)/(fp1-fq1))
			}

			// clip to q2
			if fp1 > length2 && fq1 < length2 {
				cp = Lerp(p1, q1, (fp1-length2)/(fp1-fq1))
			} else if fq1 > length2 && fp1 < length2 {
				cq = Lerp(q1, p1, (fq1-length2)/(fq1-fp1))
			}

			var sp float32 = Dot(Sub(cp, p2), normalB)
			var sq float32 = Dot(Sub(cq, p2), normalB)

			if sp <= distance+LinearSlop() || sq <= distance+LinearSlop() {
				var mp *LocalManifoldPoint
				mp = &manifold.Points[0]
				mp.Point = MulAdd(cp, 0.5*(radiusB-radiusA-sp), normalB)
				mp.Separation = sp - radius
				mp.ID = (uint16(uint8(0))<<8 | uint16(uint8(0)))
				mp = &manifold.Points[1]
				mp.Point = MulAdd(cq, 0.5*(radiusB-radiusA-sq), normalB)
				mp.Separation = sq - radius
				mp.ID = (uint16(uint8(1))<<8 | uint16(uint8(0)))
				manifold.PointCount = 2
			}
		}
	}

	if manifold.PointCount == 0 {
		// single point collision
		var normal Vec2 = Sub(closest2, closest1)
		if Dot(normal, normal) > epsSqr {
			normal = Normalize(normal)
		} else {
			normal = LeftPerp(u1)
		}

		var c1 Vec2 = MulAdd(closest1, radiusA, normal)
		var c2 Vec2 = MulAdd(closest2, -radiusB, normal)

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
		manifold.Points[0].Point = Lerp(c1, c2, 0.5)
		manifold.Points[0].Separation = float32(math.Sqrt(float64(distanceSquared))) - radius
		manifold.Points[0].ID = (uint16(uint8(i1))<<8 | uint16(uint8(i2)))
		manifold.PointCount = 1
	}

	// Undo the origin shift so points are in frame A
	for i := 0; i < manifold.PointCount; i++ {
		manifold.Points[i].Point = Add(manifold.Points[i].Point, origin)
	}

	return manifold
}

// Compute the contact manifold between an segment and a capsule
func CollideSegmentAndCapsule(segmentA *Segment, capsuleB *Capsule, xf Transform) LocalManifold {
	var capsuleA Capsule = Capsule{segmentA.Point1, segmentA.Point2, 0.0}
	return CollideCapsules(&capsuleA, capsuleB, xf)
}

// Compute the contact manifold between a polygon and capsule
func CollidePolygonAndCapsule(polygonA *Polygon, capsuleB *Capsule, xf Transform) LocalManifold {
	var polyB Polygon = makeCapsule(capsuleB.Center1, capsuleB.Center2, capsuleB.Radius)
	return CollidePolygons(polygonA, &polyB, xf)
}

func clipPolygons(polyA, polyB *Polygon, edgeA, edgeB int, flip bool) LocalManifold {
	var manifold LocalManifold = LocalManifold{}

	// reference polygon
	var poly1 *Polygon
	var i11, i12 int

	// incident polygon
	var poly2 *Polygon
	var i21, i22 int

	if flip {
		poly1 = polyB
		poly2 = polyA
		i11 = edgeB
		if edgeB+1 < polyB.Count {
			i12 = edgeB + 1
		} else {
			i12 = 0
		}
		i21 = edgeA
		if edgeA+1 < polyA.Count {
			i22 = edgeA + 1
		} else {
			i22 = 0
		}
	} else {
		poly1 = polyA
		poly2 = polyB
		i11 = edgeA
		if edgeA+1 < polyA.Count {
			i12 = edgeA + 1
		} else {
			i12 = 0
		}
		i21 = edgeB
		if edgeB+1 < polyB.Count {
			i22 = edgeB + 1
		} else {
			i22 = 0
		}
	}

	var normal Vec2 = poly1.Normals[i11]

	// Reference edge vertices
	var v11 Vec2 = poly1.Vertices[i11]
	var v12 Vec2 = poly1.Vertices[i12]

	// Incident edge vertices
	var v21 Vec2 = poly2.Vertices[i21]
	var v22 Vec2 = poly2.Vertices[i22]

	var tangent Vec2 = CrossSV(1.0, normal)

	var lower1 float32 = 0.0
	var upper1 float32 = Dot(Sub(v12, v11), tangent)

	// Incident edge points opposite of tangent due to CCW winding
	var upper2 float32 = Dot(Sub(v21, v11), tangent)
	var lower2 float32 = Dot(Sub(v22, v11), tangent)

	// Are the segments disjoint?
	if upper2 < lower1 || upper1 < lower2 {
		return manifold
	}

	var vLower Vec2
	if lower2 < lower1 && upper2-lower2 > float32(0x1p-23) {
		vLower = Lerp(v22, v21, (lower1-lower2)/(upper2-lower2))
	} else {
		vLower = v22
	}

	var vUpper Vec2
	if upper2 > upper1 && upper2-lower2 > float32(0x1p-23) {
		vUpper = Lerp(v22, v21, (upper1-lower2)/(upper2-lower2))
	} else {
		vUpper = v21
	}

	// todo vLower can be very close to vUpper, reduce to one point?

	var separationLower float32 = Dot(Sub(vLower, v11), normal)
	var separationUpper float32 = Dot(Sub(vUpper, v11), normal)

	var r1 float32 = poly1.Radius
	var r2 float32 = poly2.Radius

	// Put contact points at midpoint, accounting for radii
	vLower = MulAdd(vLower, 0.5*(r1-r2-separationLower), normal)
	vUpper = MulAdd(vUpper, 0.5*(r1-r2-separationUpper), normal)

	var radius float32 = r1 + r2

	if flip == false {
		manifold.Normal = normal
		var cp *LocalManifoldPoint = &manifold.Points[0]

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
		manifold.Normal = Neg(normal)
		var cp *LocalManifoldPoint = &manifold.Points[0]

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
	var count1 int = poly1.Count
	var count2 int = poly2.Count
	var n1s []Vec2 = poly1.Normals[:]
	var v1s []Vec2 = poly1.Vertices[:]
	var v2s []Vec2 = poly2.Vertices[:]

	var bestIndex int = 0
	var maxSeparation float32 = -float32(0x1.fffffep127)
	for i := 0; i < count1; i++ {
		// Get poly1 normal in frame2.
		var n Vec2 = n1s[i]
		var v1 Vec2 = v1s[i]

		// Find the deepest point for normal i.
		var si float32 = float32(0x1.fffffep127)
		for j := 0; j < count2; j++ {
			var sij float32 = Dot(n, Sub(v2s[j], v1))
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

// Compute the contact manifold between two polygons
func CollidePolygons(polygonA, polygonB *Polygon, xf Transform) LocalManifold {
	var origin Vec2 = polygonA.Vertices[0]
	var linearSlop float32 = LinearSlop()
	var speculativeDistance float32 = SpeculativeDistance()

	// Shift to the origin in frame A for round-off, a pure translation in A's frame
	var xfs Transform = Transform{Sub(xf.P, origin), xf.Q}

	var localPolyA Polygon
	localPolyA.Count = polygonA.Count
	localPolyA.Radius = polygonA.Radius
	localPolyA.Vertices[0] = Vec2{}
	localPolyA.Normals[0] = polygonA.Normals[0]
	for i := 1; i < localPolyA.Count; i++ {
		localPolyA.Vertices[i] = Sub(polygonA.Vertices[i], origin)
		localPolyA.Normals[i] = polygonA.Normals[i]
	}

	// Put polyB in polyA's frame to reduce round-off error
	var localPolyB Polygon
	localPolyB.Count = polygonB.Count
	localPolyB.Radius = polygonB.Radius
	for i := 0; i < localPolyB.Count; i++ {
		localPolyB.Vertices[i] = TransformPoint(xfs, polygonB.Vertices[i])
		localPolyB.Normals[i] = RotateVector(xfs.Q, polygonB.Normals[i])
	}

	var radius float32 = localPolyA.Radius + localPolyB.Radius

	var edgeA int = 0
	var separationA float32 = findMaxSeparation(&edgeA, &localPolyA, &localPolyB)
	if separationA > speculativeDistance+radius {
		return LocalManifold{}
	}

	var edgeB int = 0
	var separationB float32 = findMaxSeparation(&edgeB, &localPolyB, &localPolyA)
	if separationB > speculativeDistance+radius {
		return LocalManifold{}
	}

	// Find incident edge
	var flip bool
	if separationA >= separationB {
		flip = false

		var searchDirection Vec2 = localPolyA.Normals[edgeA]

		// Find the incident edge on polyB
		var count int = localPolyB.Count
		var normals []Vec2 = localPolyB.Normals[:]
		edgeB = 0
		var minDot float32 = float32(0x1.fffffep127)
		for i := 0; i < count; i++ {
			var dot float32 = Dot(searchDirection, normals[i])
			if dot < minDot {
				minDot = dot
				edgeB = i
			}
		}
	} else {
		flip = true

		var searchDirection Vec2 = localPolyB.Normals[edgeB]

		// Find the incident edge on polyA
		var count int = localPolyA.Count
		var normals []Vec2 = localPolyA.Normals[:]
		edgeA = 0
		var minDot float32 = float32(0x1.fffffep127)
		for i := 0; i < count; i++ {
			var dot float32 = Dot(searchDirection, normals[i])
			if dot < minDot {
				minDot = dot
				edgeA = i
			}
		}
	}

	var manifold LocalManifold = LocalManifold{}

	// Using slop here to ensure vertex-vertex normal vectors can be safely normalized
	// todo this means edge clipping needs to handle slightly non-overlapping edges.
	if separationA > float32(0.1*linearSlop) || separationB > float32(0.1*linearSlop) {
		// Edges are disjoint. Find closest points between reference edge and incident edge
		// Reference edge on polygon A
		var i11 int = edgeA
		var i12 int
		if edgeA+1 < localPolyA.Count {
			i12 = edgeA + 1
		} else {
			i12 = 0
		}
		var i21 int = edgeB
		var i22 int
		if edgeB+1 < localPolyB.Count {
			i22 = edgeB + 1
		} else {
			i22 = 0
		}

		var v11 Vec2 = localPolyA.Vertices[i11]
		var v12 Vec2 = localPolyA.Vertices[i12]
		var v21 Vec2 = localPolyB.Vertices[i21]
		var v22 Vec2 = localPolyB.Vertices[i22]

		var result SegmentDistanceResult = SegmentDistance(v11, v12, v21, v22)
		assert(result.DistanceSquared > 0.0, "result.distanceSquared > 0.0f")
		var distance float32 = float32(math.Sqrt(float64(result.DistanceSquared)))
		var separation float32 = distance - radius

		if distance-radius > speculativeDistance {
			// This can happen in the vertex-vertex case
			return manifold
		}

		// Attempt to clip edges
		manifold = clipPolygons(&localPolyA, &localPolyB, edgeA, edgeB, flip)

		var minSeparation float32 = float32(0x1.fffffep127)
		for i := 0; i < manifold.PointCount; i++ {
			minSeparation = MinFloat(minSeparation, manifold.Points[i].Separation)
		}

		// Does vertex-vertex have substantially larger separation?
		if separation+float32(0.1*linearSlop) < minSeparation {
			if result.Fraction1 == 0.0 && result.Fraction2 == 0.0 {
				// v11 - v21
				var normal Vec2 = Sub(v21, v11)
				var invDistance float32 = 1.0 / distance
				normal.X *= invDistance
				normal.Y *= invDistance

				var c1 Vec2 = MulAdd(v11, localPolyA.Radius, normal)
				var c2 Vec2 = MulAdd(v21, -localPolyB.Radius, normal)

				manifold.Normal = normal
				manifold.Points[0].Point = Lerp(c1, c2, 0.5)
				manifold.Points[0].Separation = distance - radius
				manifold.Points[0].ID = (uint16(uint8(i11))<<8 | uint16(uint8(i21)))
				manifold.PointCount = 1
			} else if result.Fraction1 == 0.0 && result.Fraction2 == 1.0 {
				// v11 - v22
				var normal Vec2 = Sub(v22, v11)
				var invDistance float32 = 1.0 / distance
				normal.X *= invDistance
				normal.Y *= invDistance

				var c1 Vec2 = MulAdd(v11, localPolyA.Radius, normal)
				var c2 Vec2 = MulAdd(v22, -localPolyB.Radius, normal)

				manifold.Normal = normal
				manifold.Points[0].Point = Lerp(c1, c2, 0.5)
				manifold.Points[0].Separation = distance - radius
				manifold.Points[0].ID = (uint16(uint8(i11))<<8 | uint16(uint8(i22)))
				manifold.PointCount = 1
			} else if result.Fraction1 == 1.0 && result.Fraction2 == 0.0 {
				// v12 - v21
				var normal Vec2 = Sub(v21, v12)
				var invDistance float32 = 1.0 / distance
				normal.X *= invDistance
				normal.Y *= invDistance

				var c1 Vec2 = MulAdd(v12, localPolyA.Radius, normal)
				var c2 Vec2 = MulAdd(v21, -localPolyB.Radius, normal)

				manifold.Normal = normal
				manifold.Points[0].Point = Lerp(c1, c2, 0.5)
				manifold.Points[0].Separation = distance - radius
				manifold.Points[0].ID = (uint16(uint8(i12))<<8 | uint16(uint8(i21)))
				manifold.PointCount = 1
			} else if result.Fraction1 == 1.0 && result.Fraction2 == 1.0 {
				// v12 - v22
				var normal Vec2 = Sub(v22, v12)
				var invDistance float32 = 1.0 / distance
				normal.X *= invDistance
				normal.Y *= invDistance

				var c1 Vec2 = MulAdd(v12, localPolyA.Radius, normal)
				var c2 Vec2 = MulAdd(v22, -localPolyB.Radius, normal)

				manifold.Normal = normal
				manifold.Points[0].Point = Lerp(c1, c2, 0.5)
				manifold.Points[0].Separation = distance - radius
				manifold.Points[0].ID = (uint16(uint8(i12))<<8 | uint16(uint8(i22)))
				manifold.PointCount = 1
			}
		}
	} else {
		// Polygons overlap
		manifold = clipPolygons(&localPolyA, &localPolyB, edgeA, edgeB, flip)
	}

	// Undo the origin shift so points are in frame A
	for i := 0; i < manifold.PointCount; i++ {
		manifold.Points[i].Point = Add(manifold.Points[i].Point, origin)
	}

	return manifold
}

// Compute the contact manifold between an segment and a circle
func CollideSegmentAndCircle(segmentA *Segment, circleB *Circle, xf Transform) LocalManifold {
	var capsuleA Capsule = Capsule{segmentA.Point1, segmentA.Point2, 0.0}
	return CollideCapsuleAndCircle(&capsuleA, circleB, xf)
}

// Compute the contact manifold between an segment and a polygon
func CollideSegmentAndPolygon(segmentA *Segment, polygonB *Polygon, xf Transform) LocalManifold {
	var polygonA Polygon = makeCapsule(segmentA.Point1, segmentA.Point2, 0.0)
	return CollidePolygons(&polygonA, polygonB, xf)
}

// Compute the contact manifold between a chain segment and a circle
func CollideChainSegmentAndCircle(segmentA *ChainSegment, circleB *Circle, xf Transform) LocalManifold {
	var manifold LocalManifold = LocalManifold{}

	// Compute circle in frame of segment
	var pB Vec2 = TransformPoint(xf, circleB.Center)

	var p1 Vec2 = segmentA.Segment.Point1
	var p2 Vec2 = segmentA.Segment.Point2
	var e Vec2 = Sub(p2, p1)

	// Normal points to the right
	var offset float32 = Dot(RightPerp(e), Sub(pB, p1))
	if offset < 0.0 {
		// collision is one-sided
		return manifold
	}

	// Barycentric coordinates
	var u float32 = Dot(e, Sub(p2, pB))
	var v float32 = Dot(e, Sub(pB, p1))

	var pA Vec2

	if v <= 0.0 {
		// Behind point1?
		// Is pB in the Voronoi region of the previous edge?
		var prevEdge Vec2 = Sub(p1, segmentA.Ghost1)
		var uPrev float32 = Dot(prevEdge, Sub(pB, p1))
		if uPrev <= 0.0 {
			return manifold
		}

		pA = p1
	} else if u <= 0.0 {
		// Ahead of point2?
		var nextEdge Vec2 = Sub(segmentA.Ghost2, p2)
		var vNext float32 = Dot(nextEdge, Sub(pB, p2))

		// Is pB in the Voronoi region of the next edge?
		if vNext > 0.0 {
			return manifold
		}

		pA = p2
	} else {
		var ee float32 = Dot(e, e)
		pA = Vec2{float32(u*p1.X) + float32(v*p2.X), float32(u*p1.Y) + float32(v*p2.Y)}
		if ee > 0.0 {
			pA = MulSV(1.0/ee, pA)
		} else {
			pA = p1
		}
	}

	var distance float32
	var normal Vec2 = GetLengthAndNormalize(&distance, Sub(pB, pA))

	var radius float32 = circleB.Radius
	var separation float32 = distance - radius
	if separation > SpeculativeDistance() {
		return manifold
	}

	var cA Vec2 = pA
	var cB Vec2 = MulAdd(pB, -radius, normal)

	manifold.Normal = normal

	var mp *LocalManifoldPoint = &manifold.Points[0]
	mp.Point = Lerp(cA, cB, 0.5)
	mp.Separation = separation
	mp.ID = 0
	manifold.PointCount = 1
	return manifold
}

// Compute the contact manifold between a chain segment and a capsule
func CollideChainSegmentAndCapsule(segmentA *ChainSegment, capsuleB *Capsule, xf Transform, cache *SimplexCache) LocalManifold {
	var polyB Polygon = makeCapsule(capsuleB.Center1, capsuleB.Center2, capsuleB.Radius)
	return CollideChainSegmentAndPolygon(segmentA, &polyB, xf, cache)
}

func clipSegments(a1, a2, b1, b2, normal Vec2, ra, rb float32, id1, id2 uint16) LocalManifold {
	var manifold LocalManifold = LocalManifold{}

	var tangent Vec2 = LeftPerp(normal)

	// Barycentric coordinates of each point relative to a1 along tangent
	var lower1 float32 = 0.0
	var upper1 float32 = Dot(Sub(a2, a1), tangent)

	// Incident edge points opposite of tangent due to CCW winding
	var upper2 float32 = Dot(Sub(b1, a1), tangent)
	var lower2 float32 = Dot(Sub(b2, a1), tangent)

	// Do segments overlap?
	if upper2 < lower1 || upper1 < lower2 {
		return manifold
	}

	var vLower Vec2
	if lower2 < lower1 && upper2-lower2 > float32(0x1p-23) {
		vLower = Lerp(b2, b1, (lower1-lower2)/(upper2-lower2))
	} else {
		vLower = b2
	}

	var vUpper Vec2
	if upper2 > upper1 && upper2-lower2 > float32(0x1p-23) {
		vUpper = Lerp(b2, b1, (upper1-lower2)/(upper2-lower2))
	} else {
		vUpper = b1
	}

	// todo vLower can be very close to vUpper, reduce to one point?

	var separationLower float32 = Dot(Sub(vLower, a1), normal)
	var separationUpper float32 = Dot(Sub(vUpper, a1), normal)

	// Put contact points at midpoint, accounting for radii
	vLower = MulAdd(vLower, 0.5*(ra-rb-separationLower), normal)
	vUpper = MulAdd(vUpper, 0.5*(ra-rb-separationUpper), normal)

	var radius float32 = ra + rb

	manifold.Normal = normal
	{
		var cp *LocalManifoldPoint = &manifold.Points[0]
		cp.Point = vLower
		cp.Separation = separationLower - radius
		cp.ID = id1
	}

	{
		var cp *LocalManifoldPoint = &manifold.Points[1]
		cp.Point = vUpper
		cp.Separation = separationUpper - radius
		cp.ID = id2
	}

	manifold.PointCount = 2

	return manifold
}

func classifyNormal(params chainSegmentParams, normal Vec2) normalType {
	var sinTol float32 = 0.01

	if Dot(normal, params.edge1) <= 0.0 {
		// Normal points towards the segment tail
		if params.convex1 {
			if Cross(normal, params.normal0) > sinTol {
				return normalSkip
			}

			return normalAdmit
		} else {
			return normalSnap
		}
	} else {
		// Normal points towards segment head
		if params.convex2 {
			if Cross(params.normal2, normal) > sinTol {
				return normalSkip
			}

			return normalAdmit
		} else {
			return normalSnap
		}
	}
}

// Compute the contact manifold between a chain segment and a rounded polygon
func CollideChainSegmentAndPolygon(segmentA *ChainSegment, polygonB *Polygon, xf Transform, cache *SimplexCache) LocalManifold {
	var manifold LocalManifold = LocalManifold{}

	var centroidB Vec2 = TransformPoint(xf, polygonB.Centroid)
	var radiusB float32 = polygonB.Radius

	var p1 Vec2 = segmentA.Segment.Point1
	var p2 Vec2 = segmentA.Segment.Point2

	var edge1 Vec2 = Normalize(Sub(p2, p1))

	var smoothParams chainSegmentParams = chainSegmentParams{}
	smoothParams.edge1 = edge1

	var convexTol float32 = 0.01
	var edge0 Vec2 = Normalize(Sub(p1, segmentA.Ghost1))
	smoothParams.normal0 = RightPerp(edge0)
	smoothParams.convex1 = Cross(edge0, edge1) >= convexTol

	var edge2 Vec2 = Normalize(Sub(segmentA.Ghost2, p2))
	smoothParams.normal2 = RightPerp(edge2)
	smoothParams.convex2 = Cross(edge1, edge2) >= convexTol

	// Normal points to the right
	var normal1 Vec2 = RightPerp(edge1)
	var behind1 bool = Dot(normal1, Sub(centroidB, p1)) < 0.0
	var behind0 bool = true
	var behind2 bool = true
	if smoothParams.convex1 {
		behind0 = Dot(smoothParams.normal0, Sub(centroidB, p1)) < 0.0
	}

	if smoothParams.convex2 {
		behind2 = Dot(smoothParams.normal2, Sub(centroidB, p2)) < 0.0
	}

	if behind1 && behind0 && behind2 {
		// one-sided collision
		return manifold
	}

	// Get polygonB in frameA
	var count int = polygonB.Count
	var vertices [MaxPolygonVertices]Vec2
	var normals [MaxPolygonVertices]Vec2
	for i := 0; i < count; i++ {
		vertices[i] = TransformPoint(xf, polygonB.Vertices[i])
		normals[i] = RotateVector(xf.Q, polygonB.Normals[i])
	}

	// Distance doesn't work correctly with partial polygons
	var input DistanceInput
	input.ProxyA = MakeProxy([]Vec2{segmentA.Segment.Point1, segmentA.Segment.Point2}, 0.0)
	input.ProxyB = MakeProxy(vertices[:count], 0.0)
	input.Transform = TransformIdentity
	input.UseRadii = false

	var output DistanceOutput = ShapeDistance(&input, cache, nil)

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
	var incidentIndex int = -1
	var incidentNormal int = -1

	if behind1 == false && output.Distance > float32(0.1*LinearSlop()) {
		// The closest features may be two vertices or an edge and a vertex even when there should
		// be face contact

		if cache.Count == 1 {
			// vertex-vertex collision
			var pA Vec2 = output.PointA
			var pB Vec2 = output.PointB

			var normal Vec2 = Normalize(Sub(pB, pA))

			var kind normalType = classifyNormal(smoothParams, normal)
			if kind == normalSkip {
				return manifold
			}

			if kind == normalAdmit {
				manifold.Normal = normal
				var cp *LocalManifoldPoint = &manifold.Points[0]
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
			assert(cache.Count == 2, "cache->count == 2")

			var ia1 int = int(cache.IndexA[0])
			var ia2 int = int(cache.IndexA[1])
			var ib1 int = int(cache.IndexB[0])
			var ib2 int = int(cache.IndexB[1])

			if ia1 == ia2 {
				// 1 point on A, expect 2 points on B
				assert(ib1 != ib2, "ib1 != ib2")

				// Find polygon normal most aligned with vector between closest points.
				// This effectively sorts ib1 and ib2
				var normalB Vec2 = Sub(output.PointA, output.PointB)
				var dot1 float32 = Dot(normalB, normals[ib1])
				var dot2 float32 = Dot(normalB, normals[ib2])
				var ib int
				if dot1 > dot2 {
					ib = ib1
				} else {
					ib = ib2
				}

				// Use accurate normal
				normalB = normals[ib]

				var kind normalType = classifyNormal(smoothParams, Neg(normalB))
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

					var b1 Vec2 = vertices[ib1]
					var b2 Vec2 = vertices[ib2]

					// Find incident segment vertex
					dot1 = Dot(normalB, Sub(p1, b1))
					dot2 = Dot(normalB, Sub(p2, b1))

					if dot1 < dot2 {
						if Dot(n0, normalB) < Dot(normal1, normalB) {
							// Neighbor is incident
							return manifold
						}
					} else {
						if Dot(n2, normalB) < Dot(normal1, normalB) {
							// Neighbor is incident
							return manifold
						}
					}

					manifold =
						clipSegments(b1, b2, p1, p2, normalB, radiusB, 0.0, (uint16(uint8(ib1))<<8 | uint16(uint8(1))), (uint16(uint8(ib2))<<8 | uint16(uint8(0))))

					assert(manifold.PointCount == 0 || manifold.PointCount == 2, "manifold.pointCount == 0 || manifold.pointCount == 2")
					if manifold.PointCount == 2 {
						manifold.Normal = Neg(normalB)
					}
					return manifold
				}

				// fall through b2_normalSnap
				incidentNormal = ib
			} else {
				// Get index of incident polygonB vertex
				var dot1 float32 = Dot(normal1, Sub(vertices[ib1], p1))
				var dot2 float32 = Dot(normal1, Sub(vertices[ib2], p2))
				if dot1 < dot2 {
					incidentIndex = ib1
				} else {
					incidentIndex = ib2
				}
			}
		}
	} else {
		// SAT edge normal
		var edgeSeparation float32 = float32(0x1.fffffep127)

		for i := 0; i < count; i++ {
			var s float32 = Dot(normal1, Sub(vertices[i], p1))
			if s < edgeSeparation {
				edgeSeparation = s
				incidentIndex = i
			}
		}

		// Check convex neighbor for edge separation
		if smoothParams.convex1 {
			var s0 float32 = float32(0x1.fffffep127)

			for i := 0; i < count; i++ {
				var s float32 = Dot(smoothParams.normal0, Sub(vertices[i], p1))
				if s < s0 {
					s0 = s
				}
			}

			if s0 > edgeSeparation {
				edgeSeparation = s0

				// Indicate neighbor owns edge separation
				incidentIndex = -1
			}
		}

		// Check convex neighbor for edge separation
		if smoothParams.convex2 {
			var s2 float32 = float32(0x1.fffffep127)

			for i := 0; i < count; i++ {
				var s float32 = Dot(smoothParams.normal2, Sub(vertices[i], p2))
				if s < s2 {
					s2 = s
				}
			}

			if s2 > edgeSeparation {
				edgeSeparation = s2

				// Indicate neighbor owns edge separation
				incidentIndex = -1
			}
		}

		// SAT polygon normals
		var polygonSeparation float32 = -float32(0x1.fffffep127)
		var referenceIndex int = -1

		for i := 0; i < count; i++ {
			var n Vec2 = normals[i]

			var kind normalType = classifyNormal(smoothParams, Neg(n))
			if kind != normalAdmit {
				continue
			}

			// Check the infinite sides of the partial polygon
			// if ((smoothParams.convex1 && b2Cross(n0, n) > 0.0f) || (smoothParams.convex2 && b2Cross(n, n2) > 0.0f))
			//{
			//	continue;
			//}

			var p Vec2 = vertices[i]
			var s float32 = MinFloat(Dot(n, Sub(p2, p)), Dot(n, Sub(p1, p)))

			if s > polygonSeparation {
				polygonSeparation = s
				referenceIndex = i
			}
		}

		if polygonSeparation > edgeSeparation {
			var ia1 int = referenceIndex
			var ia2 int
			if ia1 < count-1 {
				ia2 = ia1 + 1
			} else {
				ia2 = 0
			}
			var a1 Vec2 = vertices[ia1]
			var a2 Vec2 = vertices[ia2]

			var n Vec2 = normals[ia1]

			var dot1 float32 = Dot(n, Sub(p1, a1))
			var dot2 float32 = Dot(n, Sub(p2, a1))

			if dot1 < dot2 {
				if Dot(n0, n) < Dot(normal1, n) {
					// Neighbor is incident
					return manifold
				}
			} else {
				if Dot(n2, n) < Dot(normal1, n) {
					// Neighbor is incident
					return manifold
				}
			}

			manifold = clipSegments(a1, a2, p1, p2, normals[ia1], radiusB, 0.0, (uint16(uint8(ia1))<<8 | uint16(uint8(1))), (uint16(uint8(ia2))<<8 | uint16(uint8(0))))

			assert(manifold.PointCount == 0 || manifold.PointCount == 2, "manifold.pointCount == 0 || manifold.pointCount == 2")
			if manifold.PointCount == 2 {
				manifold.Normal = Neg(normals[ia1])
			}

			return manifold
		}

		if incidentIndex == -1 {
			// neighboring segment is the separating axis
			return manifold
		}

		// fall through segment normal axis
	}

	assert(incidentNormal != -1 || incidentIndex != -1, "incidentNormal != -1 || incidentIndex != -1")

	// Segment normal

	// Find incident polygon normal: normal adjacent to deepest vertex that is most anti-parallel to segment normal
	var b1, b2 Vec2
	var ib1, ib2 int

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
		var i2 int = incidentIndex
		var i1 int
		if i2 > 0 {
			i1 = i2 - 1
		} else {
			i1 = count - 1
		}
		var d1 float32 = Dot(normal1, normals[i1])
		var d2 float32 = Dot(normal1, normals[i2])
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

	assert(manifold.PointCount == 0 || manifold.PointCount == 2, "manifold.pointCount == 0 || manifold.pointCount == 2")

	return manifold
}
