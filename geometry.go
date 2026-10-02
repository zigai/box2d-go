// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// The maximum number of vertices on a convex polygon. Changing this affects performance even if you
// don't use more vertices.
const MaxPolygonVertices = 8

// Low level ray cast input data
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

// A solid circle
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
// Ghost1 -> Segment.Point1 -> Segment.Point2 -> Ghost2
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
// @warning Do not modify these values directly, instead use ComputeHull()
type Hull struct {
	// The final points of the hull
	Points [MaxPolygonVertices]Vec2
	// The number of points
	Count int
}

// Validate ray cast input data (NaN, etc)
func IsValidRay(input *RayCastInput) bool {
	todo()
	return false
}

func computePolygonCentroid(vertices []Vec2, count int) Vec2 {
	todo()
	return Vec2{}
}

// Make a convex polygon from a convex hull. This will assert if the hull is not valid.
// @warning Do not manually fill in the hull data, it must come directly from ComputeHull
func MakePolygon(hull *Hull, radius float32) Polygon {
	todo()
	return Polygon{}
}

// Make an offset convex polygon from a convex hull. This will assert if the hull is not valid.
// @warning Do not manually fill in the hull data, it must come directly from ComputeHull
func MakeOffsetPolygon(hull *Hull, position Vec2, rotation Rot) Polygon {
	todo()
	return Polygon{}
}

// Make an offset convex polygon from a convex hull. This will assert if the hull is not valid.
// @warning Do not manually fill in the hull data, it must come directly from ComputeHull
func MakeOffsetRoundedPolygon(hull *Hull, position Vec2, rotation Rot, radius float32) Polygon {
	todo()
	return Polygon{}
}

// Make a square polygon, bypassing the need for a convex hull.
// @param halfWidth the half-width
func MakeSquare(halfWidth float32) Polygon {
	todo()
	return Polygon{}
}

// Make a box (rectangle) polygon, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
func MakeBox(halfWidth float32, halfHeight float32) Polygon {
	todo()
	return Polygon{}
}

// Make a rounded box, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
// @param radius the radius of the rounded extension
func MakeRoundedBox(halfWidth float32, halfHeight float32, radius float32) Polygon {
	todo()
	return Polygon{}
}

// Make an offset box, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
// @param center the local center of the box
// @param rotation the local rotation of the box
func MakeOffsetBox(halfWidth float32, halfHeight float32, center Vec2, rotation Rot) Polygon {
	todo()
	return Polygon{}
}

// Make an offset rounded box, bypassing the need for a convex hull.
// @param halfWidth the half-width (x-axis)
// @param halfHeight the half-height (y-axis)
// @param center the local center of the box
// @param rotation the local rotation of the box
// @param radius the radius of the rounded extension
func MakeOffsetRoundedBox(halfWidth float32, halfHeight float32, center Vec2, rotation Rot, radius float32) Polygon {
	todo()
	return Polygon{}
}

// Transform a polygon. This is useful for transferring a shape from one body to another.
func TransformPolygon(transform Transform, polygon *Polygon) Polygon {
	todo()
	return Polygon{}
}

// Compute mass properties of a circle
func ComputeCircleMass(shape *Circle, density float32) MassData {
	todo()
	return MassData{}
}

// Compute mass properties of a capsule
func ComputeCapsuleMass(shape *Capsule, density float32) MassData {
	todo()
	return MassData{}
}

// Compute mass properties of a polygon
func ComputePolygonMass(shape *Polygon, density float32) MassData {
	todo()
	return MassData{}
}

func computeCircleFatAABB(shape *Circle, xf WorldTransform, extra float32) AABB {
	todo()
	return AABB{}
}

func computeCapsuleFatAABB(shape *Capsule, xf WorldTransform, extra float32) AABB {
	todo()
	return AABB{}
}

func computePolygonFatAABB(shape *Polygon, xf WorldTransform, extra float32) AABB {
	todo()
	return AABB{}
}

func computeSegmentFatAABB(shape *Segment, xf WorldTransform, extra float32) AABB {
	todo()
	return AABB{}
}

// Compute the bounding box of a transformed circle
func ComputeCircleAABB(shape *Circle, xf WorldTransform) AABB {
	todo()
	return AABB{}
}

// Compute the bounding box of a transformed capsule
func ComputeCapsuleAABB(shape *Capsule, xf WorldTransform) AABB {
	todo()
	return AABB{}
}

// Compute the bounding box of a transformed polygon
func ComputePolygonAABB(shape *Polygon, xf WorldTransform) AABB {
	todo()
	return AABB{}
}

// Compute the bounding box of a transformed line segment
func ComputeSegmentAABB(shape *Segment, xf WorldTransform) AABB {
	todo()
	return AABB{}
}

// Conservative world AABB for a shape, inflated by extra margin. In large world mode this is
// computed in double and rounded outward so the inflation is not lost far from the origin.
func computeFatShapeAABB(shape *shape, xf WorldTransform, extra float32) AABB {
	todo()
	return AABB{}
}

// Test a point for overlap with a circle in local space
func PointInCircle(shape *Circle, point Vec2) bool {
	todo()
	return false
}

// Test a point for overlap with a capsule in local space
func PointInCapsule(shape *Capsule, point Vec2) bool {
	todo()
	return false
}

// Test a point for overlap with a convex polygon in local space
func PointInPolygon(shape *Polygon, point Vec2) bool {
	todo()
	return false
}

// Ray cast versus circle shape in local space.
func RayCastCircle(shape *Circle, input *RayCastInput) CastOutput {
	todo()
	return CastOutput{}
}

// Ray cast versus capsule shape in local space.
func RayCastCapsule(shape *Capsule, input *RayCastInput) CastOutput {
	todo()
	return CastOutput{}
}

// Ray cast versus segment shape in local space. Optionally treat the segment as one-sided with hits from
// the left side being treated as a miss.
func RayCastSegment(shape *Segment, input *RayCastInput, oneSided bool) CastOutput {
	todo()
	return CastOutput{}
}

// Ray cast versus polygon shape in local space.
func RayCastPolygon(shape *Polygon, input *RayCastInput) CastOutput {
	todo()
	return CastOutput{}
}

// Shape cast versus a circle.
func ShapeCastCircle(shape *Circle, input *ShapeCastInput) CastOutput {
	todo()
	return CastOutput{}
}

// Shape cast versus a capsule.
func ShapeCastCapsule(shape *Capsule, input *ShapeCastInput) CastOutput {
	todo()
	return CastOutput{}
}

// Shape cast versus a line segment.
func ShapeCastSegment(shape *Segment, input *ShapeCastInput) CastOutput {
	todo()
	return CastOutput{}
}

// Shape cast versus a convex polygon.
func ShapeCastPolygon(shape *Polygon, input *ShapeCastInput) CastOutput {
	todo()
	return CastOutput{}
}

func collideMoverAndCircle(mover *Capsule, shape *Circle) PlaneResult {
	todo()
	return PlaneResult{}
}

func collideMoverAndCapsule(mover *Capsule, shape *Capsule) PlaneResult {
	todo()
	return PlaneResult{}
}

func collideMoverAndPolygon(mover *Capsule, shape *Polygon) PlaneResult {
	todo()
	return PlaneResult{}
}

func collideMoverAndSegment(mover *Capsule, shape *Segment) PlaneResult {
	todo()
	return PlaneResult{}
}
