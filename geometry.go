// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

// The maximum number of vertices on a convex polygon. Changing this affects performance even if you
// don't use more vertices.
const MaxPolygonVertices = 8

// Low level ray cast input data.
type RayCastInput struct {
	// Start point of the ray cast
	Origin Vec2
	// Translation of the ray cast
	Translation Vec2
	// The maximum fraction of the translation to consider, typically 1
	MaxFraction float32
}

// A distance proxy is used by the GJK algorithm. It encapsulates any shape.
// You can provide between 1 and MaxPolygonVertices and a radius.
type ShapeProxy struct {
	// The point cloud
	Points [MaxPolygonVertices]Vec2
	// The number of points. Must be greater than 0.
	Count int
	// The external radius of the point cloud. May be zero.
	Radius float32
}

// Low level shape cast input in generic form. This allows casting an arbitrary point
// cloud wrap with a radius. For example, a circle is a single point with a non-zero radius.
// A capsule is two points with a non-zero radius. A box is four points with a zero radius.
type ShapeCastInput struct {
	// A generic shape
	Proxy ShapeProxy
	// The translation of the shape cast
	Translation Vec2
	// The maximum fraction of the translation to consider, typically 1
	MaxFraction float32
	// Allow shape cast to encroach when initially touching. This only works if the radius is greater than zero.
	CanEncroach bool
}

// Low level ray cast or shape-cast output data. The hit point is in the local or relative frame
// of the input. Returns a zero fraction and normal in the case of initial overlap.
type CastOutput struct {
	// The surface normal at the hit point
	Normal Vec2
	// The surface hit point
	Point Vec2
	// The fraction of the input translation at collision
	Fraction float32
	// The number of iterations used
	Iterations int
	// Did the cast hit?
	Hit bool
}

// This holds the mass data computed for a shape.
type MassData struct {
	// The mass of the shape, usually in kilograms.
	Mass float32
	// The position of the shape's centroid relative to the shape's origin.
	Center Vec2
	// The rotational inertia of the shape about the shape center.
	RotationalInertia float32
}

// A solid circle.
type Circle struct {
	// The local center
	Center Vec2
	// The radius
	Radius float32
}

// A solid capsule can be viewed as two semicircles connected
// by a rectangle.
type Capsule struct {
	// Local center of the first semicircle
	Center1 Vec2
	// Local center of the second semicircle
	Center2 Vec2
	// The radius of the semicircles
	Radius float32
}

// A solid convex polygon. It is assumed that the interior of the polygon is to
// the left of each edge.
// Polygons have a maximum number of vertices equal to MaxPolygonVertices.
// In most cases you should not need many vertices for a convex polygon.
// @warning DO NOT fill this out manually, instead use a helper function like
// MakePolygon or MakeBox.
type Polygon struct {
	// The polygon vertices
	Vertices [MaxPolygonVertices]Vec2
	// The outward normal vectors of the polygon sides
	Normals [MaxPolygonVertices]Vec2
	// The centroid of the polygon
	Centroid Vec2
	// The external radius for rounded polygons
	Radius float32
	// The number of polygon vertices
	Count int
}

// A line segment with two-sided collision.
type Segment struct {
	// The first point
	Point1 Vec2
	// The second point
	Point2 Vec2
}

// A line segment with one-sided collision. Only collides on the right side.
// Several of these are generated for a chain shape.
// Ghost1 -> Segment.Point1 -> Segment.Point2 -> Ghost2.
type ChainSegment struct {
	// The tail ghost vertex
	Ghost1 Vec2
	// The line segment
	Segment Segment
	// The head ghost vertex
	Ghost2 Vec2
	// The owning chain shape index (internal usage only)
	ChainID int
}

// A convex hull. Used to create convex polygons.
// @warning Do not modify these values directly, instead use ComputeHull().
type Hull struct {
	// The final points of the hull
	Points [MaxPolygonVertices]Vec2
	// The number of points
	Count int
}

// Validate ray cast input data (NaN, etc).
func (input *RayCastInput) IsValid() bool {
	isValid := input.Origin.IsValid() && input.Translation.IsValid() &&
		IsValidFloat(input.MaxFraction) && 0.0 <= input.MaxFraction && input.MaxFraction < Huge()

	return isValid
}

// Make a convex polygon from a convex hull. This will assert if the hull is not valid.
// @warning Do not manually fill in the hull data, it must come directly from ComputeHull.
func MakePolygon(hull *Hull, radius float32) Polygon {
	assert(ValidateHull(hull), "b2ValidateHull( hull )")

	if hull.Count < 3 {
		// Handle a bad hull when assertions are disabled
		return MakeSquare(0.5)
	}

	var shape Polygon

	shape.Count = hull.Count
	shape.Radius = radius
	// Copy vertices
	for i := range shape.Count {
		shape.Vertices[i] = hull.Points[i]
	}

	// Compute normals. Ensure the edges have non-zero length.
	for i := range shape.Count {
		i1 := i

		i2 := i + 1
		if i2 == shape.Count {
			i2 = 0
		}

		edge := shape.Vertices[i2].Sub(shape.Vertices[i1])
		assert(edge.Dot(edge) > float32(0x1p-23)*float32(0x1p-23), "b2Dot( edge, edge ) > FLT_EPSILON * FLT_EPSILON")
		shape.Normals[i] = edge.CrossScalar(1.0).Normalize()
	}

	shape.Centroid = computePolygonCentroid(shape.Vertices[:], shape.Count)

	return shape
}

// Make an offset convex polygon from a convex hull. This will assert if the hull is not valid.
// @warning Do not manually fill in the hull data, it must come directly from ComputeHull.
func MakeOffsetPolygon(hull *Hull, position Vec2, rotation Rot) Polygon {
	return MakeOffsetRoundedPolygon(hull, position, rotation, 0.0)
}

// Make an offset convex polygon from a convex hull. This will assert if the hull is not valid.
// @warning Do not manually fill in the hull data, it must come directly from ComputeHull.
func MakeOffsetRoundedPolygon(hull *Hull, position Vec2, rotation Rot, radius float32) Polygon {
	assert(ValidateHull(hull), "b2ValidateHull( hull )")

	if hull.Count < 3 {
		// Handle a bad hull when assertions are disabled
		return MakeSquare(0.5)
	}

	transform := Transform{position, rotation}

	var shape Polygon

	shape.Count = hull.Count
	shape.Radius = radius
	// Copy vertices
	for i := range shape.Count {
		shape.Vertices[i] = transform.Apply(hull.Points[i])
	}

	// Compute normals. Ensure the edges have non-zero length.
	for i := range shape.Count {
		i1 := i

		i2 := i + 1
		if i2 == shape.Count {
			i2 = 0
		}

		edge := shape.Vertices[i2].Sub(shape.Vertices[i1])
		assert(edge.Dot(edge) > float32(0x1p-23)*float32(0x1p-23), "b2Dot( edge, edge ) > FLT_EPSILON * FLT_EPSILON")
		shape.Normals[i] = edge.CrossScalar(1.0).Normalize()
	}

	shape.Centroid = computePolygonCentroid(shape.Vertices[:], shape.Count)

	return shape
}

// Make a square polygon, bypassing the need for a convex hull.
// @param halfWidth the half-width.
func MakeSquare(halfWidth float32) Polygon {
	return MakeBox(halfWidth, halfWidth)
}

// Make a box (rectangle) polygon, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis).
func MakeBox(halfWidth float32, halfHeight float32) Polygon {
	assert(IsValidFloat(halfWidth) && halfWidth > 0.0, "b2IsValidFloat( halfWidth ) && halfWidth > 0.0f")
	assert(IsValidFloat(halfHeight) && halfHeight > 0.0, "b2IsValidFloat( halfHeight ) && halfHeight > 0.0f")

	var shape Polygon

	shape.Count = 4
	shape.Vertices[0] = Vec2{-halfWidth, -halfHeight}
	shape.Vertices[1] = Vec2{halfWidth, -halfHeight}
	shape.Vertices[2] = Vec2{halfWidth, halfHeight}
	shape.Vertices[3] = Vec2{-halfWidth, halfHeight}
	shape.Normals[0] = Vec2{0.0, -1.0}
	shape.Normals[1] = Vec2{1.0, 0.0}
	shape.Normals[2] = Vec2{0.0, 1.0}
	shape.Normals[3] = Vec2{-1.0, 0.0}
	shape.Radius = 0.0
	shape.Centroid = Vec2Zero

	return shape
}

// Make a rounded box, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
// @param radius the radius of the rounded extension.
func MakeRoundedBox(halfWidth float32, halfHeight float32, radius float32) Polygon {
	assert(IsValidFloat(radius) && radius >= 0.0, "b2IsValidFloat( radius ) && radius >= 0.0f")

	shape := MakeBox(halfWidth, halfHeight)
	shape.Radius = radius

	return shape
}

// Make an offset box, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
// @param center the local center of the box
// @param rotation the local rotation of the box.
func MakeOffsetBox(halfWidth float32, halfHeight float32, center Vec2, rotation Rot) Polygon {
	xf := Transform{center, rotation}

	var shape Polygon

	shape.Count = 4
	shape.Vertices[0] = xf.Apply(Vec2{-halfWidth, -halfHeight})
	shape.Vertices[1] = xf.Apply(Vec2{halfWidth, -halfHeight})
	shape.Vertices[2] = xf.Apply(Vec2{halfWidth, halfHeight})
	shape.Vertices[3] = xf.Apply(Vec2{-halfWidth, halfHeight})
	shape.Normals[0] = xf.Q.Rotate(Vec2{0.0, -1.0})
	shape.Normals[1] = xf.Q.Rotate(Vec2{1.0, 0.0})
	shape.Normals[2] = xf.Q.Rotate(Vec2{0.0, 1.0})
	shape.Normals[3] = xf.Q.Rotate(Vec2{-1.0, 0.0})
	shape.Radius = 0.0
	shape.Centroid = xf.P

	return shape
}

// Make an offset rounded box, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
// @param center the local center of the box
// @param rotation the local rotation of the box
// @param radius the radius of the rounded extension.
func MakeOffsetRoundedBox(halfWidth float32, halfHeight float32, center Vec2, rotation Rot, radius float32) Polygon {
	assert(IsValidFloat(radius) && radius >= 0.0, "b2IsValidFloat( radius ) && radius >= 0.0f")

	xf := Transform{center, rotation}

	var shape Polygon

	shape.Count = 4
	shape.Vertices[0] = xf.Apply(Vec2{-halfWidth, -halfHeight})
	shape.Vertices[1] = xf.Apply(Vec2{halfWidth, -halfHeight})
	shape.Vertices[2] = xf.Apply(Vec2{halfWidth, halfHeight})
	shape.Vertices[3] = xf.Apply(Vec2{-halfWidth, halfHeight})
	shape.Normals[0] = xf.Q.Rotate(Vec2{0.0, -1.0})
	shape.Normals[1] = xf.Q.Rotate(Vec2{1.0, 0.0})
	shape.Normals[2] = xf.Q.Rotate(Vec2{0.0, 1.0})
	shape.Normals[3] = xf.Q.Rotate(Vec2{-1.0, 0.0})
	shape.Radius = radius
	shape.Centroid = xf.P

	return shape
}

// Transform a polygon. This is useful for transferring a shape from one body to another.
func TransformPolygon(transform Transform, polygon *Polygon) Polygon {
	p := *polygon

	for i := range p.Count {
		p.Vertices[i] = transform.Apply(p.Vertices[i])
		p.Normals[i] = transform.Q.Rotate(p.Normals[i])
	}

	p.Centroid = transform.Apply(p.Centroid)

	return p
}

// Compute mass properties of a circle.
func ComputeCircleMass(shape *Circle, density float32) MassData {
	rr := shape.Radius * shape.Radius

	var massData MassData

	massData.Mass = density * Pi * rr
	massData.Center = shape.Center
	// inertia about the center of mass
	massData.RotationalInertia = massData.Mass * 0.5 * rr

	return massData
}

// Compute mass properties of a capsule.
func ComputeCapsuleMass(shape *Capsule, density float32) MassData {
	var (
		radius = shape.Radius
		rr     = float32(radius * radius)
	)

	p1 := shape.Center1
	p2 := shape.Center2

	var (
		length = p2.Sub(p1).Length()
		ll     = float32(length * length)
	)

	var (
		circleMass = float32(density * (Pi * radius * radius))
		boxMass    = float32(density * (2.0 * radius * length))
	)

	var massData MassData

	massData.Mass = circleMass + boxMass
	massData.Center.X = 0.5 * (p1.X + p2.X)
	massData.Center.Y = 0.5 * (p1.Y + p2.Y)
	// two offset half circles, both halves add up to full circle and each half is offset by half length
	// semi-circle centroid = 4 r / 3 pi
	// Need to apply parallel-axis theorem twice:
	// 1. shift semi-circle centroid to origin
	// 2. shift semi-circle to box end
	// m * ((h + lc)^2 - lc^2) = m * (h^2 + 2 * h * lc)
	// See: https://en.wikipedia.org/wiki/Parallel_axis_theorem
	// I verified this formula by computing the convex hull of a 128 vertex capsule

	// half circle centroid
	lc := 4.0 * radius / (3.0 * Pi)

	// half length of rectangular portion of capsule
	h := 0.5 * length

	var (
		circleInertia = float32(circleMass * (float32(0.5*rr) + float32(h*h) + float32(2.0*h*lc)))
		boxInertia    = boxMass * (float32(4.0*rr) + ll) / 12.0
	)

	massData.RotationalInertia = circleInertia + boxInertia

	return massData
}

// Compute mass properties of a polygon.
func ComputePolygonMass(shape *Polygon, density float32) MassData {
	// Polygon mass, centroid, and inertia.
	// Let rho be the polygon density in mass per unit area.
	// Then:
	// mass = rho * int(dA)
	// centroid.X = (1/mass) * rho * int(x * dA)
	// centroid.Y = (1/mass) * rho * int(y * dA)
	// I = rho * int((x*x + y*y) * dA)
	//
	// We can compute these integrals by summing all the integrals
	// for each triangle of the polygon. To evaluate the integral
	// for a single triangle, we make a change of variables to
	// the (u,v) coordinates of the triangle:
	// x = x0 + e1x * u + e2x * v
	// y = y0 + e1y * u + e2y * v
	// where 0 <= u && 0 <= v && u + v <= 1.
	//
	// We integrate u from [0,1-v] and then v from [0,1].
	// We also need to use the Jacobian of the transformation:
	// D = cross(e1, e2)
	//
	// Simplification: triangle centroid = (1/3) * (p1 + p2 + p3)
	//
	// The rest of the derivation is handled by computer algebra.
	assert(shape.Count > 0, "shape->count > 0")

	if shape.Count == 1 {
		var circle Circle

		circle.Center = shape.Vertices[0]
		circle.Radius = shape.Radius

		return ComputeCircleMass(&circle, density)
	}

	if shape.Count == 2 {
		var capsule Capsule

		capsule.Center1 = shape.Vertices[0]
		capsule.Center2 = shape.Vertices[1]
		capsule.Radius = shape.Radius

		return ComputeCapsuleMass(&capsule, density)
	}

	vertices := [MaxPolygonVertices]Vec2{}

	count := shape.Count

	radius := shape.Radius

	if radius > 0.0 {
		// Approximate mass of rounded polygons by pushing out the vertices.
		var sqrt2 float32 = 1.412

		for i := range count {
			j := i - 1
			if i == 0 {
				j = count - 1
			}

			n1 := shape.Normals[j]
			n2 := shape.Normals[i]

			mid := n1.Add(n2).Normalize()
			vertices[i] = shape.Vertices[i].MulAdd(sqrt2*radius, mid)
		}
	} else {
		for i := range count {
			vertices[i] = shape.Vertices[i]
		}
	}

	var (
		center            Vec2
		area              float32 = 0.0
		rotationalInertia float32 = 0.0
	)

	// Get a reference point for forming triangles.
	// Use the first vertex to reduce round-off errors.
	r := vertices[0]

	var inv3 float32 = 1.0 / 3.0

	for i := 1; i < count-1; i++ {
		// Triangle edges
		e1 := vertices[i].Sub(r)
		e2 := vertices[i+1].Sub(r)

		d := e1.Cross(e2)

		triangleArea := float32(0.5 * d)

		area += triangleArea
		// Area weighted centroid, r at origin
		center = center.MulAdd(triangleArea*inv3, e1.Add(e2))
		ex1, ey1 := e1.X, e1.Y
		ex2, ey2 := e2.X, e2.Y

		var (
			intx2 = float32(ex1*ex1) + float32(ex2*ex1) + float32(ex2*ex2)
			inty2 = float32(ey1*ey1) + float32(ey2*ey1) + float32(ey2*ey2)
		)

		rotationalInertia += float32((0.25 * inv3 * d) * (intx2 + inty2))
	}

	var massData MassData

	// Total mass
	massData.Mass = density * area
	// Center of mass, shift back from origin at r
	assert(area > float32(0x1p-23), "area > FLT_EPSILON")

	invArea := 1.0 / area

	center.X = float32(center.X * invArea)
	center.Y = float32(center.Y * invArea)
	massData.Center = r.Add(center)
	// Inertia tensor relative to the local origin (point s).
	massData.RotationalInertia = float32(density * rotationalInertia)
	// Shift inertia to center of mass
	massData.RotationalInertia -= float32(massData.Mass * center.Dot(center))
	// If this goes negative we are hosed
	assert(massData.RotationalInertia >= 0.0, "massData.rotationalInertia >= 0.0f")

	return massData
}

// Compute the bounding box of a transformed circle.
func ComputeCircleAABB(shape *Circle, xf WorldTransform) AABB {
	return computeCircleFatAABB(shape, xf, 0.0)
}

// Compute the bounding box of a transformed capsule.
func ComputeCapsuleAABB(shape *Capsule, xf WorldTransform) AABB {
	return computeCapsuleFatAABB(shape, xf, 0.0)
}

// Compute the bounding box of a transformed polygon.
func ComputePolygonAABB(shape *Polygon, xf WorldTransform) AABB {
	return computePolygonFatAABB(shape, xf, 0.0)
}

// Compute the bounding box of a transformed line segment.
func ComputeSegmentAABB(shape *Segment, xf WorldTransform) AABB {
	return computeSegmentFatAABB(shape, xf, 0.0)
}

// Test a point for overlap with a circle in local space.
func PointInCircle(shape *Circle, point Vec2) bool {
	center := shape.Center
	return point.DistanceSquared(center) <= shape.Radius*shape.Radius
}

// Test a point for overlap with a capsule in local space.
func PointInCapsule(shape *Capsule, point Vec2) bool {
	rr := shape.Radius * shape.Radius

	p1 := shape.Center1
	p2 := shape.Center2

	d := p2.Sub(p1)

	dd := d.Dot(d)
	if dd == 0.0 {
		// Capsule is really a circle
		return point.DistanceSquared(p1) <= rr
	}

	// Get closest point on capsule segment
	// c = p1 + t * d
	// dot(point - c, d) = 0
	// dot(point - p1 - t * d, d) = 0
	// t = dot(point - p1, d) / dot(d, d)
	t := point.Sub(p1).Dot(d) / dd

	t = clampFloat(t, 0.0, 1.0)
	c := p1.MulAdd(t, d)

	// Is query point within radius around closest point?
	return point.DistanceSquared(c) <= rr
}

// Test a point for overlap with a convex polygon in local space.
func PointInPolygon(shape *Polygon, point Vec2) bool {
	var input DistanceInput

	input.ProxyA = MakeProxy(shape.Vertices[:shape.Count], 0.0)
	input.ProxyB = MakeProxy([]Vec2{point}, 0.0)
	input.Transform = TransformIdentity
	input.UseRadii = false

	var cache SimplexCache

	output := ShapeDistance(&input, &cache, nil)

	return output.Distance <= shape.Radius
}

// Ray cast versus circle shape in local space.
func RayCastCircle(shape *Circle, input *RayCastInput) CastOutput {
	assert(input.IsValid(), "b2IsValidRay( input )")

	p := shape.Center

	var output CastOutput

	// Shift ray so circle center is the origin
	s := input.Origin.Sub(p)

	var (
		r  = shape.Radius
		rr = float32(r * r)
	)

	length, d := input.Translation.LengthAndNormalize()
	if length == 0.0 {
		// zero length ray
		if s.LengthSquared() < rr {
			// initial overlap
			output.Point = input.Origin
			output.Hit = true
		}

		return output
	}

	// Find closest point on ray to origin

	// solve: dot(s + t * d, d) = 0
	t := -s.Dot(d)

	// c is the closest point on the line to the origin
	c := s.MulAdd(t, d)

	cc := c.Dot(c)

	if cc > rr {
		// closest point is outside the circle
		return output
	}

	// Pythagoras
	h := float32(math.Sqrt(float64(rr - cc)))

	fraction := t - h

	if fraction < 0.0 || input.MaxFraction*length < fraction {
		// intersection is point outside the range of the ray segment
		if s.LengthSquared() < rr {
			// initial overlap
			output.Point = input.Origin
			output.Hit = true
		}

		return output
	}

	// hit point relative to center
	hitPoint := s.MulAdd(fraction, d)

	output.Fraction = fraction / length
	output.Normal = hitPoint.Normalize()
	output.Point = p.MulAdd(shape.Radius, output.Normal)
	output.Hit = true

	return output
}

// Ray cast versus capsule shape in local space.
func RayCastCapsule(shape *Capsule, input *RayCastInput) CastOutput {
	assert(input.IsValid(), "b2IsValidRay( input )")

	var output CastOutput

	v1 := shape.Center1
	v2 := shape.Center2

	e := v2.Sub(v1)

	capsuleLength, a := e.LengthAndNormalize()

	if capsuleLength < float32(0x1p-23) {
		// Capsule is really a circle
		circle := Circle{v1, shape.Radius}
		return RayCastCircle(&circle, input)
	}

	p1 := input.Origin
	d := input.Translation

	// Ray from capsule start to ray start
	q := p1.Sub(v1)

	qa := q.Dot(a)

	// Vector to ray start that is perpendicular to capsule axis
	qp := q.MulAdd(-qa, a)

	radius := shape.Radius

	// Does the ray start within the infinite length capsule?
	if qp.Dot(qp) < radius*radius {
		if qa < 0.0 {
			// start point behind capsule segment
			circle := Circle{v1, shape.Radius}
			return RayCastCircle(&circle, input)
		}

		if qa > capsuleLength {
			// start point ahead of capsule segment
			circle := Circle{v2, shape.Radius}
			return RayCastCircle(&circle, input)
		}

		// ray starts inside capsule -> no hit
		output.Point = input.Origin
		output.Hit = true

		return output
	}

	// Perpendicular to capsule axis, pointing right
	n := Vec2{a.Y, -a.X}

	rayLength, u := d.LengthAndNormalize()

	// Intersect ray with infinite length capsule
	// v1 + radius * n + s1 * a = p1 + s2 * u
	// v1 - radius * n + s1 * a = p1 + s2 * u

	// s1 * a - s2 * u = b
	// = q - radius * ap
	// or
	// b = q + radius * ap

	// Cramer's rule [a -u]
	den := float32(-a.X*u.Y) + float32(u.X*a.Y)
	if -float32(0x1p-23) < den && den < float32(0x1p-23) {
		// Ray is parallel to capsule and outside infinite length capsule
		return output
	}

	b1 := q.MulSub(radius, n)
	b2 := q.MulAdd(radius, n)

	invDen := 1.0 / den

	// Cramer's rule [a b1]
	s21 := (float32(a.X*b1.Y) - float32(b1.X*a.Y)) * invDen

	// Cramer's rule [a b2]
	s22 := (float32(a.X*b2.Y) - float32(b2.X*a.Y)) * invDen

	var (
		s2 float32
		b  Vec2
	)

	if s21 < s22 {
		s2 = s21
		b = b1
	} else {
		s2 = s22
		b = b2
		n = n.Neg()
	}

	if s2 < 0.0 || input.MaxFraction*rayLength < s2 {
		return output
	}

	// Cramer's rule [b -u]
	s1 := (float32(-b.X*u.Y) + float32(u.X*b.Y)) * invDen

	if s1 < 0.0 {
		// ray passes behind capsule segment
		circle := Circle{v1, shape.Radius}
		return RayCastCircle(&circle, input)
	}

	if capsuleLength < s1 {
		// ray passes ahead of capsule segment
		circle := Circle{v2, shape.Radius}
		return RayCastCircle(&circle, input)
	}
	// ray hits capsule side
	output.Fraction = s2 / rayLength
	output.Point = v1.Lerp(v2, s1/capsuleLength).Add(n.Scale(shape.Radius))
	output.Normal = n
	output.Hit = true

	return output
}

// Ray cast versus segment shape in local space. Optionally treat the segment as one-sided with hits from
// the left side being treated as a miss.
func RayCastSegment(shape *Segment, input *RayCastInput, oneSided bool) CastOutput {
	if oneSided {
		// Skip left-side collision
		offset := input.Origin.Sub(shape.Point1).Cross(shape.Point2.Sub(shape.Point1))
		if offset < 0.0 {
			var output CastOutput
			return output
		}
	}

	// Put the ray into the edge's frame of reference.
	p1 := input.Origin
	d := input.Translation

	v1 := shape.Point1
	v2 := shape.Point2
	e := v2.Sub(v1)

	var output CastOutput

	length, eUnit := e.LengthAndNormalize()
	if length == 0.0 {
		return output
	}

	// Normal points to the right, looking from v1 towards v2
	normal := eUnit.RightPerp()

	// Intersect ray with infinite segment using normal
	// Similar to intersecting a ray with an infinite plane
	// p = p1 + t * d
	// dot(normal, p - v1) = 0
	// dot(normal, p1 - v1) + t * dot(normal, d) = 0
	var (
		numerator   = normal.Dot(v1.Sub(p1))
		denominator = normal.Dot(d)
	)

	if denominator == 0.0 {
		// parallel
		return output
	}

	t := numerator / denominator
	if t < 0.0 || input.MaxFraction < t {
		// out of ray range
		return output
	}

	// Intersection point on infinite segment
	p := p1.MulAdd(t, d)

	// Compute position of p along segment
	// p = v1 + s * e
	// s = dot(p - v1, e) / dot(e, e)

	s := p.Sub(v1).Dot(eUnit)
	if s < 0.0 || length < s {
		// out of segment range
		return output
	}

	if numerator > 0.0 {
		normal = normal.Neg()
	}

	output.Fraction = t
	output.Point = p
	output.Normal = normal
	output.Hit = true

	return output
}

// Ray cast versus polygon shape in local space.
func RayCastPolygon(shape *Polygon, input *RayCastInput) CastOutput {
	assert(input.IsValid(), "b2IsValidRay( input )")

	if shape.Radius == 0.0 {
		return rayCastSharpPolygon(shape, input)
	}

	var castInput ShapeCastPairInput

	castInput.ProxyA = MakeProxy(shape.Vertices[:shape.Count], shape.Radius)
	castInput.ProxyB = MakeProxy([]Vec2{input.Origin}, 0.0)
	castInput.Transform = TransformIdentity
	castInput.TranslationB = input.Translation
	castInput.MaxFraction = input.MaxFraction
	castInput.CanEncroach = false

	return ShapeCast(&castInput)
}

// Shape cast versus a circle.
func ShapeCastCircle(shape *Circle, input *ShapeCastInput) CastOutput {
	var pairInput ShapeCastPairInput

	pairInput.ProxyA = MakeProxy([]Vec2{shape.Center}, shape.Radius)
	pairInput.ProxyB = input.Proxy
	pairInput.Transform = TransformIdentity
	pairInput.TranslationB = input.Translation
	pairInput.MaxFraction = input.MaxFraction
	pairInput.CanEncroach = input.CanEncroach

	return ShapeCast(&pairInput)
}

// Shape cast versus a capsule.
func ShapeCastCapsule(shape *Capsule, input *ShapeCastInput) CastOutput {
	var pairInput ShapeCastPairInput

	pairInput.ProxyA = MakeProxy([]Vec2{shape.Center1, shape.Center2}, shape.Radius)
	pairInput.ProxyB = input.Proxy
	pairInput.Transform = TransformIdentity
	pairInput.TranslationB = input.Translation
	pairInput.MaxFraction = input.MaxFraction
	pairInput.CanEncroach = input.CanEncroach

	return ShapeCast(&pairInput)
}

// Shape cast versus a line segment.
func ShapeCastSegment(shape *Segment, input *ShapeCastInput) CastOutput {
	var pairInput ShapeCastPairInput

	pairInput.ProxyA = MakeProxy([]Vec2{shape.Point1, shape.Point2}, 0.0)
	pairInput.ProxyB = input.Proxy
	pairInput.Transform = TransformIdentity
	pairInput.TranslationB = input.Translation
	pairInput.MaxFraction = input.MaxFraction
	pairInput.CanEncroach = input.CanEncroach

	return ShapeCast(&pairInput)
}

// Shape cast versus a convex polygon.
func ShapeCastPolygon(shape *Polygon, input *ShapeCastInput) CastOutput {
	var pairInput ShapeCastPairInput

	pairInput.ProxyA = MakeProxy(shape.Vertices[:shape.Count], shape.Radius)
	pairInput.ProxyB = input.Proxy
	pairInput.Transform = TransformIdentity
	pairInput.TranslationB = input.Translation
	pairInput.MaxFraction = input.MaxFraction
	pairInput.CanEncroach = input.CanEncroach

	return ShapeCast(&pairInput)
}

func computePolygonCentroid(vertices []Vec2, count int) Vec2 {
	var (
		center Vec2
		area   float32 = 0.0
	)

	// Get a reference point for forming triangles.
	// Use the first vertex to reduce round-off errors.
	origin := vertices[0]

	var inv3 float32 = 1.0 / 3.0

	for i := 1; i < count-1; i++ {
		// Triangle edges
		e1 := vertices[i].Sub(origin)
		e2 := vertices[i+1].Sub(origin)

		a := float32(0.5 * e1.Cross(e2))

		// Area weighted centroid
		center = center.MulAdd(a*inv3, e1.Add(e2))
		area += a
	}

	assert(area > float32(0x1p-23), "area > FLT_EPSILON")

	invArea := 1.0 / area

	center.X = float32(center.X * invArea)
	center.Y = float32(center.Y * invArea)
	// Restore offset
	center = origin.Add(center)

	return center
}

func computeCircleFatAABB(shape *Circle, xf WorldTransform, extra float32) AABB {
	c := xf.Apply(shape.Center)
	r := float64(shape.Radius) + float64(extra)

	return AABB{
		Vec2{RoundDownFloat(float64(c.X) - r), RoundDownFloat(float64(c.Y) - r)},
		Vec2{RoundUpFloat(float64(c.X) + r), RoundUpFloat(float64(c.Y) + r)},
	}
}

func computeCapsuleFatAABB(shape *Capsule, xf WorldTransform, extra float32) AABB {
	v1 := xf.Apply(shape.Center1)
	v2 := xf.Apply(shape.Center2)
	lx, ly := v2.X, v2.Y

	ux, uy := v2.X, v2.Y
	if v1.X < v2.X {
		lx = v1.X
	}

	if v1.Y < v2.Y {
		ly = v1.Y
	}

	if v1.X > v2.X {
		ux = v1.X
	}

	if v1.Y > v2.Y {
		uy = v1.Y
	}

	r := float64(shape.Radius) + float64(extra)

	return AABB{
		Vec2{RoundDownFloat(float64(lx) - r), RoundDownFloat(float64(ly) - r)},
		Vec2{RoundUpFloat(float64(ux) + r), RoundUpFloat(float64(uy) + r)},
	}
}

func computePolygonFatAABB(shape *Polygon, xf WorldTransform, extra float32) AABB {
	assert(shape.Count > 0, "shape->count > 0")
	v := xf.Apply(shape.Vertices[0])

	lx, ly, ux, uy := float64(v.X), float64(v.Y), float64(v.X), float64(v.Y)
	for i := 1; i < shape.Count; i++ {
		v = xf.Apply(shape.Vertices[i])
		if float64(v.X) < lx {
			lx = float64(v.X)
		}

		if float64(v.Y) < ly {
			ly = float64(v.Y)
		}

		if float64(v.X) > ux {
			ux = float64(v.X)
		}

		if float64(v.Y) > uy {
			uy = float64(v.Y)
		}
	}

	r := float64(shape.Radius) + float64(extra)

	return AABB{
		Vec2{RoundDownFloat(lx - r), RoundDownFloat(ly - r)},
		Vec2{RoundUpFloat(ux + r), RoundUpFloat(uy + r)},
	}
}

func computeSegmentFatAABB(shape *Segment, xf WorldTransform, extra float32) AABB {
	v1 := xf.Apply(shape.Point1)
	v2 := xf.Apply(shape.Point2)
	lx, ly := v2.X, v2.Y

	ux, uy := v2.X, v2.Y
	if v1.X < v2.X {
		lx = v1.X
	}

	if v1.Y < v2.Y {
		ly = v1.Y
	}

	if v1.X > v2.X {
		ux = v1.X
	}

	if v1.Y > v2.Y {
		uy = v1.Y
	}

	r := float64(extra)

	return AABB{
		Vec2{RoundDownFloat(float64(lx) - r), RoundDownFloat(float64(ly) - r)},
		Vec2{RoundUpFloat(float64(ux) + r), RoundUpFloat(float64(uy) + r)},
	}
}

// Conservative world AABB for a shape, inflated by extra margin. In large world mode this is
// computed in double and rounded outward so the inflation is not lost far from the origin.
func computeFatShapeAABB(shape *shape, xf WorldTransform, extra float32) AABB {
	switch shape.shapeType {
	case CapsuleShape:
		return computeCapsuleFatAABB(&shape.capsule, xf, extra)
	case CircleShape:
		return computeCircleFatAABB(&shape.circle, xf, extra)
	case PolygonShape:
		return computePolygonFatAABB(&shape.polygon, xf, extra)
	case SegmentShape:
		return computeSegmentFatAABB(&shape.segment, xf, extra)
	case ChainSegmentShape:
		return computeSegmentFatAABB(&shape.chainSegment.Segment, xf, extra)
	case ShapeTypeCount:
		assert(false, "false")
		return AABB{}
	default:
		{
			assert(false, "false")

			var empty AABB

			return empty
		}
	}
}

func collideMoverAndCircle(mover *Capsule, shape *Circle) PlaneResult {
	var distanceInput DistanceInput

	distanceInput.ProxyA = MakeProxy([]Vec2{shape.Center}, 0.0)
	distanceInput.ProxyB = MakeProxy([]Vec2{mover.Center1, mover.Center2}, mover.Radius)
	distanceInput.Transform = TransformIdentity
	distanceInput.UseRadii = false

	totalRadius := mover.Radius + shape.Radius

	var cache SimplexCache

	distanceOutput := ShapeDistance(&distanceInput, &cache, nil)

	if distanceOutput.Distance <= totalRadius {
		plane := Plane{distanceOutput.Normal, totalRadius - distanceOutput.Distance}

		return PlaneResult{
			Plane: plane,
			Point: distanceOutput.PointA,
			Hit:   true,
		}
	}

	return PlaneResult{}
}

func collideMoverAndCapsule(mover *Capsule, shape *Capsule) PlaneResult {
	var distanceInput DistanceInput

	distanceInput.ProxyA = MakeProxy([]Vec2{shape.Center1, shape.Center2}, 0.0)
	distanceInput.ProxyB = MakeProxy([]Vec2{mover.Center1, mover.Center2}, mover.Radius)
	distanceInput.Transform = TransformIdentity
	distanceInput.UseRadii = false

	totalRadius := mover.Radius + shape.Radius

	var cache SimplexCache

	distanceOutput := ShapeDistance(&distanceInput, &cache, nil)

	if distanceOutput.Distance <= totalRadius {
		plane := Plane{distanceOutput.Normal, totalRadius - distanceOutput.Distance}

		return PlaneResult{
			Plane: plane,
			Point: distanceOutput.PointA,
			Hit:   true,
		}
	}

	return PlaneResult{}
}

func collideMoverAndPolygon(mover *Capsule, shape *Polygon) PlaneResult {
	var distanceInput DistanceInput

	distanceInput.ProxyA = MakeProxy(shape.Vertices[:shape.Count], shape.Radius)
	distanceInput.ProxyB = MakeProxy([]Vec2{mover.Center1, mover.Center2}, mover.Radius)
	distanceInput.Transform = TransformIdentity
	distanceInput.UseRadii = false

	totalRadius := mover.Radius + shape.Radius

	var cache SimplexCache

	distanceOutput := ShapeDistance(&distanceInput, &cache, nil)

	if distanceOutput.Distance <= totalRadius {
		plane := Plane{distanceOutput.Normal, totalRadius - distanceOutput.Distance}

		return PlaneResult{
			Plane: plane,
			Point: distanceOutput.PointA,
			Hit:   true,
		}
	}

	return PlaneResult{}
}

func collideMoverAndSegment(mover *Capsule, shape *Segment) PlaneResult {
	var distanceInput DistanceInput

	distanceInput.ProxyA = MakeProxy([]Vec2{shape.Point1, shape.Point2}, 0.0)
	distanceInput.ProxyB = MakeProxy([]Vec2{mover.Center1, mover.Center2}, mover.Radius)
	distanceInput.Transform = TransformIdentity
	distanceInput.UseRadii = false

	totalRadius := mover.Radius

	var cache SimplexCache

	distanceOutput := ShapeDistance(&distanceInput, &cache, nil)

	if distanceOutput.Distance <= totalRadius {
		plane := Plane{distanceOutput.Normal, totalRadius - distanceOutput.Distance}

		return PlaneResult{
			Plane: plane,
			Point: distanceOutput.PointA,
			Hit:   true,
		}
	}

	return PlaneResult{}
}

func rayCastSharpPolygon(shape *Polygon, input *RayCastInput) CastOutput {
	// Shift all math to first vertex since the polygon may be far
	// from the origin.
	base := shape.Vertices[0]

	p1 := input.Origin.Sub(base)
	d := input.Translation

	var lower float32

	upper := input.MaxFraction

	index := -1

	var output CastOutput

	for edgeIndex := range shape.Count {
		// p = p1 + a * d
		// dot(normal, p - v) = 0
		// dot(normal, p1 - v) + a * dot(normal, d) = 0
		vertex := shape.Vertices[edgeIndex].Sub(base)

		var (
			numerator   = shape.Normals[edgeIndex].Dot(vertex.Sub(p1))
			denominator = shape.Normals[edgeIndex].Dot(d)
		)

		if denominator == 0.0 {
			// Parallel and runs outside edge
			if numerator < 0.0 {
				return output
			}

			if upper < lower {
				return output
			}

			continue
		}
		// Note: we want this predicate without division:
		// lower < numerator / denominator, where denominator < 0
		// Since denominator < 0, we have to flip the inequality:
		// lower < numerator / denominator <==> denominator * lower > numerator.
		if denominator < 0.0 && numerator < lower*denominator {
			// Increase lower.
			// The segment enters this half-space.
			lower = numerator / denominator
			index = edgeIndex
		} else if denominator > 0.0 && numerator < upper*denominator {
			// Decrease upper.
			// The segment exits this half-space.
			upper = numerator / denominator
		}

		if upper < lower {
			// Ray misses
			return output
		}
	}

	assert(0.0 <= lower && lower <= input.MaxFraction, "0.0f <= lower && lower <= input->maxFraction")

	if index >= 0 {
		output.Fraction = lower
		output.Normal = shape.Normals[index]
		output.Point = input.Origin.MulAdd(lower, d)
		output.Hit = true
	} else {
		// initial overlap
		output.Point = input.Origin
		output.Hit = true
	}

	return output
}
