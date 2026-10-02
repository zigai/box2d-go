// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// Result of computing the distance between two line segments
type SegmentDistanceResult struct {
	// The closest point on the first segment
	Closest1 Vec2
	// The closest point on the second segment
	Closest2 Vec2
	// The barycentric coordinate on the first segment
	Fraction1 float32
	// The barycentric coordinate on the second segment
	Fraction2 float32
	// The squared distance between the closest points
	DistanceSquared float32
}

// Used to warm start the GJK simplex. If you call this function multiple times with nearby
// transforms this might improve performance. Otherwise you can zero initialize this.
// The distance cache must be initialized to zero on the first call.
// Users should generally just zero initialize this structure for each call.
type SimplexCache struct {
	// The number of stored simplex points
	Count uint16
	// The cached simplex indices on shape A
	IndexA [3]uint8
	// The cached simplex indices on shape B
	IndexB [3]uint8
}

// Input for ShapeDistance
type DistanceInput struct {
	// The proxy for shape A
	ProxyA ShapeProxy
	// The proxy for shape B
	ProxyB ShapeProxy
	// Transform of shape B in shape A's frame, the relative pose B in A
	// (InvMulTransforms( worldA, worldB )). The query is origin independent and runs in frame A.
	Transform Transform
	// Should the proxy radius be considered?
	UseRadii bool
}

// Output for ShapeDistance
type DistanceOutput struct {
	// Closest point on shape A, in shape A's frame
	PointA Vec2
	// Closest point on shape B, in shape A's frame
	PointB Vec2
	// A to B normal in shape A's frame. Invalid if distance is zero.
	Normal Vec2
	// The final distance, zero if overlapped
	Distance float32
	// Number of GJK iterations used
	Iterations int
	// The number of simplexes stored in the simplex array
	SimplexCount int
}

// Simplex vertex for debugging the GJK algorithm
type SimplexVertex struct {
	// support point in ProxyA
	WA Vec2
	// support point in ProxyB
	WB Vec2
	// wB - wA
	W Vec2
	// barycentric coordinate for closest point
	A float32
	// wA index
	IndexA int
	// wB index
	IndexB int
}

// Simplex from the GJK algorithm
type Simplex struct {
	// vertices
	V1 SimplexVertex
	// vertices
	V2 SimplexVertex
	// vertices
	V3 SimplexVertex
	// number of valid vertices
	Count int
}

// Input parameters for ShapeCast
type ShapeCastPairInput struct {
	// The proxy for shape A
	ProxyA ShapeProxy
	// The proxy for shape B
	ProxyB ShapeProxy
	// Transform of shape B in shape A's frame, the relative pose B in A
	Transform Transform
	// The translation of shape B, in A's frame
	TranslationB Vec2
	// The fraction of the translation to consider, typically 1
	MaxFraction float32
	// Allows shapes with a radius to move slightly closer if already touching
	CanEncroach bool
}

// This describes the motion of a body/shape for TOI computation. Shapes are defined with respect to the body origin,
// which may not coincide with the center of mass. However, to support dynamics we must interpolate the center of mass
// position.
type Sweep struct {
	// Local center of mass position
	LocalCenter Vec2
	// Starting center of mass world position
	C1 Vec2
	// Ending center of mass world position
	C2 Vec2
	// Starting world rotation
	Q1 Rot
	// Ending world rotation
	Q2 Rot
}

// Time of impact input
type TOIInput struct {
	// The proxy for shape A
	ProxyA ShapeProxy
	// The proxy for shape B
	ProxyB ShapeProxy
	// The movement of shape A
	SweepA Sweep
	// The movement of shape B
	SweepB Sweep
	// Defines the sweep interval [0, maxFraction]
	MaxFraction float32
}

// Describes the TOI output
type TOIState int32

const (
	TOIStateUnknown TOIState = iota
	TOIStateFailed
	TOIStateOverlapped
	TOIStateHit
	TOIStateSeparated
)

// Time of impact output
type TOIOutput struct {
	// The type of result
	State TOIState
	// The hit point
	Point Vec2
	// The hit normal
	Normal Vec2
	// The sweep time of the collision
	Fraction float32
}

var (
	toiTime                  float32
	toiMaxTime               float32
	toiCalls                 int
	toiDistanceIterations    int
	toiMaxDistanceIterations int
	toiRootIterations        int
	toiMaxRootIterations     int
	toiFailedCount           int
	toiOverlappedCount       int
	toiHitCount              int
	toiSeparatedCount        int
)

type separationType int32

const (
	pointsType separationType = iota
	faceAType
	faceBType
)

type separationFunction struct {
	proxyA     *ShapeProxy
	proxyB     *ShapeProxy
	sweepA     Sweep
	sweepB     Sweep
	localPoint Vec2
	axis       Vec2
	type_      separationType
}

// Evaluate the transform sweep at a specific time.
func GetSweepTransform(sweep *Sweep, time float32) Transform {
	todo()
	return Transform{}
}

// Compute the distance between two line segments, clamping at the end points if needed.
func SegmentDistance(p1, q1, p2, q2 Vec2) SegmentDistanceResult {
	todo()
	return SegmentDistanceResult{}
}

// Make a proxy for use in overlap, shape cast, and related functions. This is a deep copy of the points.
func MakeProxy(points []Vec2, radius float32) ShapeProxy {
	todo()
	return ShapeProxy{}
}

// Make a proxy with a transform. This is a deep copy of the points.
func MakeOffsetProxy(points []Vec2, radius float32, position Vec2, rotation Rot) ShapeProxy {
	todo()
	return ShapeProxy{}
}

func weight2(a1 float32, w1 Vec2, a2 float32, w2 Vec2) Vec2 {
	todo()
	return Vec2{}
}

func weight3(a1 float32, w1 Vec2, a2 float32, w2 Vec2, a3 float32, w3 Vec2) Vec2 {
	todo()
	return Vec2{}
}

func findSupport(proxy *ShapeProxy, direction Vec2) int {
	todo()
	return 0
}

func makeSimplexFromCache(cache SimplexCache, proxyA, proxyB *ShapeProxy) Simplex {
	todo()
	return Simplex{}
}

func makeSimplexCache(simplex *Simplex) SimplexCache {
	todo()
	return SimplexCache{}
}

func computeWitnessPoints(s *Simplex, a, b *Vec2) {
	todo()
}

func solveSimplex2(s *Simplex) Vec2 {
	todo()
	return Vec2{}
}

func solveSimplex3(s *Simplex) Vec2 {
	todo()
	return Vec2{}
}

// Compute the closest points between two shapes represented as point clouds.
// SimplexCache cache is input/output. On the first call set SimplexCache.Count to zero.
// The underlying GJK algorithm may be debugged by passing in debug simplexes. You may pass in nil for these.
func ShapeDistance(input *DistanceInput, cache *SimplexCache, simplexes []Simplex) DistanceOutput {
	todo()
	return DistanceOutput{}
}

// Perform a linear shape cast of shape B moving and shape A fixed. Determines the hit point, normal, and translation fraction.
// The query runs in frame A, so the hit point and normal are returned in frame A. Initially touching shapes are a miss.
func ShapeCast(input *ShapeCastPairInput) CastOutput {
	todo()
	return CastOutput{}
}

func makeSeparationFunction(cache SimplexCache, proxyA *ShapeProxy, sweepA *Sweep, proxyB *ShapeProxy, sweepB *Sweep, t1 float32) separationFunction {
	todo()
	return separationFunction{}
}

func findMinSeparation(f *separationFunction, indexA, indexB *int, t float32) float32 {
	todo()
	return 0
}

func evaluateSeparation(f *separationFunction, indexA, indexB int, t float32) float32 {
	todo()
	return 0
}

// Compute the upper bound on time before two shapes penetrate. Time is represented as
// a fraction between [0,tMax]. This uses a swept separating axis and may miss some intermediate,
// non-tunneling collisions. If you change the time interval, you should call this function
// again.
func TimeOfImpact(input *TOIInput) TOIOutput {
	todo()
	return TOIOutput{}
}
