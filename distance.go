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
func (sweep *Sweep) Transform(time float32) Transform {
	// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
	var xf Transform
	xf.P = sweep.C1.Lerp(sweep.C2, time)
	q := Rot{
		float32((1.0-time)*sweep.Q1.C) + float32(time*sweep.Q2.C),
		float32((1.0-time)*sweep.Q1.S) + float32(time*sweep.Q2.S),
	}

	xf.Q = q.Normalize()
	// Shift to origin
	xf.P = xf.P.Sub(xf.Q.Rotate(sweep.LocalCenter))
	return xf
}

// Compute the distance between two line segments, clamping at the end points if needed.
func SegmentDistance(p1, q1, p2, q2 Vec2) SegmentDistanceResult {
	var result SegmentDistanceResult
	d1 := q1.Sub(p1)
	d2 := q2.Sub(p2)
	r := p1.Sub(p2)
	var dd1 float32 = d1.Dot(d1)
	var dd2 float32 = d2.Dot(d2)
	var rd1 float32 = r.Dot(d1)
	var rd2 float32 = r.Dot(d2)

	var epsSqr float32 = float32(0x1p-23) * float32(0x1p-23)

	if dd1 < epsSqr || dd2 < epsSqr {
		// Handle all degeneracies
		if dd1 >= epsSqr {
			// Segment 2 is degenerate
			result.Fraction1 = clampFloat(-rd1/dd1, 0.0, 1.0)
			result.Fraction2 = 0.0
		} else if dd2 >= epsSqr {
			// Segment 1 is degenerate
			result.Fraction1 = 0.0
			result.Fraction2 = clampFloat(rd2/dd2, 0.0, 1.0)
		} else {
			result.Fraction1 = 0.0
			result.Fraction2 = 0.0
		}
	} else {
		// Non-degenerate segments
		var d12 float32 = d1.Dot(d2)

		var denominator float32 = float32(dd1*dd2) - float32(d12*d12)

		// Fraction on segment 1
		var f1 float32 = 0.0
		if denominator != 0.0 {
			// not parallel
			f1 = clampFloat((float32(d12*rd2)-float32(rd1*dd2))/denominator, 0.0, 1.0)
		}

		// Compute point on segment 2 closest to p1 + f1 * d1
		var f2 float32 = (float32(d12*f1) + rd2) / dd2

		// Clamping of segment 2 requires a do over on segment 1
		if f2 < 0.0 {
			f2 = 0.0
			f1 = clampFloat(-rd1/dd1, 0.0, 1.0)
		} else if f2 > 1.0 {
			f2 = 1.0
			f1 = clampFloat((d12-rd1)/dd1, 0.0, 1.0)
		}

		result.Fraction1 = f1
		result.Fraction2 = f2
	}

	result.Closest1 = p1.MulAdd(result.Fraction1, d1)
	result.Closest2 = p2.MulAdd(result.Fraction2, d2)
	result.DistanceSquared = result.Closest1.DistanceSquared(result.Closest2)
	return result
}

// Make a proxy for use in overlap, shape cast, and related functions. This is a deep copy of the points.
func MakeProxy(points []Vec2, radius float32) ShapeProxy {
	count := min(len(points), MaxPolygonVertices)
	var proxy ShapeProxy
	for i := 0; i < count; i++ {
		proxy.Points[i] = points[i]
	}
	proxy.Count = count
	proxy.Radius = radius
	return proxy
}

// Make a proxy with a transform. This is a deep copy of the points.
func MakeOffsetProxy(points []Vec2, radius float32, position Vec2, rotation Rot) ShapeProxy {
	count := min(len(points), MaxPolygonVertices)
	transform := Transform{
		P: position,
		Q: rotation,
	}
	var proxy ShapeProxy
	for i := 0; i < count; i++ {
		proxy.Points[i] = transform.Apply(points[i])
	}
	proxy.Count = count
	proxy.Radius = radius
	return proxy
}

func weight2(a1 float32, w1 Vec2, a2 float32, w2 Vec2) Vec2 {
	return Vec2{float32(a1*w1.X) + float32(a2*w2.X), float32(a1*w1.Y) + float32(a2*w2.Y)}
}

func weight3(a1 float32, w1 Vec2, a2 float32, w2 Vec2, a3 float32, w3 Vec2) Vec2 {
	return Vec2{float32(a1*w1.X) + float32(a2*w2.X) + float32(a3*w3.X), float32(a1*w1.Y) + float32(a2*w2.Y) + float32(a3*w3.Y)}
}

func findSupport(proxy *ShapeProxy, direction Vec2) int {
	points := proxy.Points[:]
	count := proxy.Count

	bestIndex := 0
	var bestValue float32 = points[0].Dot(direction)
	for i := 1; i < count; i++ {
		var value float32 = points[i].Dot(direction)
		if value > bestValue {
			bestIndex = i
			bestValue = value
		}
	}

	return bestIndex
}

func makeSimplexFromCache(cache SimplexCache, proxyA, proxyB *ShapeProxy) Simplex {
	assert(cache.Count <= 3, "cache.count <= 3")
	var s Simplex

	// Copy data from cache.
	s.Count = int(cache.Count)
	vertices := [3]*SimplexVertex{&s.V1, &s.V2, &s.V3}
	for i := 0; i < s.Count; i++ {
		v := vertices[i]
		v.IndexA = int(cache.IndexA[i])
		v.IndexB = int(cache.IndexB[i])
		v.WA = proxyA.Points[v.IndexA]
		v.WB = proxyB.Points[v.IndexB]
		v.W = v.WA.Sub(v.WB)
		// invalid
		v.A = -1.0
	}

	// If the cache is empty or invalid ...
	if s.Count == 0 {
		v := vertices[0]
		v.IndexA = 0
		v.IndexB = 0
		v.WA = proxyA.Points[0]
		v.WB = proxyB.Points[0]
		v.W = v.WA.Sub(v.WB)
		v.A = 1.0
		s.Count = 1
	}

	return s
}

func makeSimplexCache(simplex *Simplex) SimplexCache {
	var cache SimplexCache
	cache.Count = uint16(simplex.Count)
	vertices := [3]*SimplexVertex{&simplex.V1, &simplex.V2, &simplex.V3}
	for i := 0; i < simplex.Count; i++ {
		cache.IndexA[i] = uint8(vertices[i].IndexA)
		cache.IndexB[i] = uint8(vertices[i].IndexB)
	}

	return cache
}

func computeWitnessPoints(s *Simplex, a, b *Vec2) {
	switch s.Count {
	case 1:
		*a = s.V1.WA
		*b = s.V1.WB

	case 2:
		*a = weight2(s.V1.A, s.V1.WA, s.V2.A, s.V2.WA)
		*b = weight2(s.V1.A, s.V1.WB, s.V2.A, s.V2.WB)

	case 3:
		*a = weight3(s.V1.A, s.V1.WA, s.V2.A, s.V2.WA, s.V3.A, s.V3.WA)
		// todo why are these not equal?
		//*b = weight3(s.V1.A, s.V1.WB, s.V2.A, s.V2.WB, s.V3.A, s.V3.WB)
		*b = *a

	default:
		*a = Vec2Zero
		*b = Vec2Zero
		assert(false, "false")

	}
}

func solveSimplex2(s *Simplex) Vec2 {
	w1 := s.V1.W
	w2 := s.V2.W
	e12 := w2.Sub(w1)

	// w1 region
	var d12_2 float32 = -w1.Dot(e12)
	if d12_2 <= 0.0 {
		// a2 <= 0, so we clamp it to 0
		s.V1.A = 1.0
		s.Count = 1
		return w1.Neg()
	}

	// w2 region
	var d12_1 float32 = w2.Dot(e12)
	if d12_1 <= 0.0 {
		// a1 <= 0, so we clamp it to 0
		s.V2.A = 1.0
		s.Count = 1
		s.V1 = s.V2
		return w2.Neg()
	}

	// Must be in e12 region.
	var inv_d12 float32 = 1.0 / (d12_1 + d12_2)
	s.V1.A = d12_1 * inv_d12
	s.V2.A = d12_2 * inv_d12
	s.Count = 2
	return e12.ScalarCross(w1.Add(w2).Cross(e12))
}

func solveSimplex3(s *Simplex) Vec2 {
	w1 := s.V1.W
	w2 := s.V2.W
	w3 := s.V3.W

	// Edge12
	// [1      1     ][a1] = [1]
	// [w1.e12 w2.e12][a2] = [0]
	// a3 = 0
	e12 := w2.Sub(w1)
	var w1e12 float32 = w1.Dot(e12)
	var w2e12 float32 = w2.Dot(e12)
	var d12_1 float32 = w2e12
	var d12_2 float32 = -w1e12

	// Edge13
	// [1      1     ][a1] = [1]
	// [w1.e13 w3.e13][a3] = [0]
	// a2 = 0
	e13 := w3.Sub(w1)
	var w1e13 float32 = w1.Dot(e13)
	var w3e13 float32 = w3.Dot(e13)
	var d13_1 float32 = w3e13
	var d13_2 float32 = -w1e13

	// Edge23
	// [1      1     ][a2] = [1]
	// [w2.e23 w3.e23][a3] = [0]
	// a1 = 0
	e23 := w3.Sub(w2)
	var w2e23 float32 = w2.Dot(e23)
	var w3e23 float32 = w3.Dot(e23)
	var d23_1 float32 = w3e23
	var d23_2 float32 = -w2e23

	// Triangle123
	var n123 float32 = e12.Cross(e13)

	var d123_1 float32 = float32(n123 * w2.Cross(w3))
	var d123_2 float32 = float32(n123 * w3.Cross(w1))
	var d123_3 float32 = float32(n123 * w1.Cross(w2))

	// w1 region
	if d12_2 <= 0.0 && d13_2 <= 0.0 {
		s.V1.A = 1.0
		s.Count = 1
		return w1.Neg()
	}

	// e12
	if d12_1 > 0.0 && d12_2 > 0.0 && d123_3 <= 0.0 {
		var inv_d12 float32 = 1.0 / (d12_1 + d12_2)
		s.V1.A = d12_1 * inv_d12
		s.V2.A = d12_2 * inv_d12
		s.Count = 2
		return e12.ScalarCross(w1.Add(w2).Cross(e12))
	}

	// e13
	if d13_1 > 0.0 && d13_2 > 0.0 && d123_2 <= 0.0 {
		var inv_d13 float32 = 1.0 / (d13_1 + d13_2)
		s.V1.A = d13_1 * inv_d13
		s.V3.A = d13_2 * inv_d13
		s.Count = 2
		s.V2 = s.V3
		return e13.ScalarCross(w1.Add(w3).Cross(e13))
	}

	// w2 region
	if d12_1 <= 0.0 && d23_2 <= 0.0 {
		s.V2.A = 1.0
		s.Count = 1
		s.V1 = s.V2
		return w2.Neg()
	}

	// w3 region
	if d13_1 <= 0.0 && d23_1 <= 0.0 {
		s.V3.A = 1.0
		s.Count = 1
		s.V1 = s.V3
		return w3.Neg()
	}

	// e23
	if d23_1 > 0.0 && d23_2 > 0.0 && d123_1 <= 0.0 {
		var inv_d23 float32 = 1.0 / (d23_1 + d23_2)
		s.V2.A = d23_1 * inv_d23
		s.V3.A = d23_2 * inv_d23
		s.Count = 2
		s.V1 = s.V3
		return e23.ScalarCross(w2.Add(w3).Cross(e23))
	}

	// Must be in triangle123
	var inv_d123 float32 = 1.0 / (d123_1 + d123_2 + d123_3)
	s.V1.A = d123_1 * inv_d123
	s.V2.A = d123_2 * inv_d123
	s.V3.A = d123_3 * inv_d123
	s.Count = 3
	// No search direction
	return Vec2Zero
}

// Compute the closest points between two shapes represented as point clouds.
// SimplexCache cache is input/output. On the first call set SimplexCache.Count to zero.
// Passing a simplex buffer records the initial, reduced iteration, and final simplexes
// up to its length. Pass nil to skip the debug history.
func ShapeDistance(input *DistanceInput, cache *SimplexCache, simplexes []Simplex) DistanceOutput {

	assert(input.ProxyA.Count > 0 && input.ProxyB.Count > 0, "input->proxyA.count > 0 && input->proxyB.count > 0")
	assert(input.ProxyA.Radius >= 0.0, "input->proxyA.radius >= 0.0f")
	assert(input.ProxyB.Radius >= 0.0, "input->proxyB.radius >= 0.0f")
	var output DistanceOutput

	proxyA := &input.ProxyA

	// Get proxyB in frame A to avoid further transforms in the main loop.
	// This is still a performance gain at 8 points.
	var localProxyB ShapeProxy
	{
		localProxyB.Count = input.ProxyB.Count
		localProxyB.Radius = input.ProxyB.Radius
		for i := 0; i < localProxyB.Count; i++ {
			localProxyB.Points[i] = input.Transform.Apply(input.ProxyB.Points[i])
		}
	}

	// Initialize the simplex.
	simplex := makeSimplexFromCache(*cache, proxyA, &localProxyB)

	simplexIndex := 0
	if simplexes != nil && simplexIndex < len(simplexes) {
		simplexes[simplexIndex] = simplex
		simplexIndex += 1
	}

	// Get simplex vertices as an array.
	vertices := [3]*SimplexVertex{&simplex.V1, &simplex.V2, &simplex.V3}

	nonUnitNormal := Vec2Zero

	// These store the vertices of the last simplex so that we can check for duplicates and prevent cycling.
	var saveA, saveB [3]int
	// Main iteration loop. All computations are done in frame A.
	maxIterations := 20
	iteration := 0
	for iteration < maxIterations {
		// Copy simplex so we can identify duplicates.
		saveCount := simplex.Count
		for i := 0; i < saveCount; i++ {
			saveA[i] = vertices[i].IndexA
			saveB[i] = vertices[i].IndexB
		}

		var d Vec2
		switch simplex.Count {
		case 1:
			d = simplex.V1.W.Neg()

		case 2:
			d = solveSimplex2(&simplex)

		case 3:
			d = solveSimplex3(&simplex)

		default:
			assert(false, "false")
		}

		// If we have 3 points, then the origin is in the corresponding triangle.
		if simplex.Count == 3 {
			// Overlap
			*cache = makeSimplexCache(&simplex)
			var localPointA, localPointB Vec2
			computeWitnessPoints(&simplex, &localPointA, &localPointB)
			output.Distance = 0.0
			output.PointA = localPointA
			output.PointB = localPointB
			return output
		}

		if simplexIndex < len(simplexes) {
			simplexes[simplexIndex] = simplex
			simplexIndex++
		}

		// Ensure the search direction is numerically fit.
		if d.Dot(d) < 1000.0*float32(0x1p-126) {
			// This is unlikely but could lead to bad cycling.
			// The branch predictor seems to make this check have low cost.

			// The origin is probably contained by a line segment
			// or triangle. Thus the shapes are overlapped.

			*cache = makeSimplexCache(&simplex)
			// Must return overlap due to invalid normal.
			var localPointA, localPointB Vec2
			computeWitnessPoints(&simplex, &localPointA, &localPointB)
			output.Distance = 0.0
			output.PointA = localPointA
			output.PointB = localPointB
			return output
		}

		// Save the normal
		nonUnitNormal = d
		// Compute a tentative new simplex vertex using support points.
		// support = support(a, d) - support(b, -d)
		vertex := vertices[simplex.Count]
		vertex.IndexA = findSupport(proxyA, d)
		vertex.WA = proxyA.Points[vertex.IndexA]
		vertex.IndexB = findSupport(&localProxyB, d.Neg())
		vertex.WB = localProxyB.Points[vertex.IndexB]
		vertex.W = vertex.WA.Sub(vertex.WB)
		// Iteration count is equated to the number of support point calls.
		iteration++

		// Check for duplicate support points. This is the main termination criteria.
		duplicate := false
		for i := 0; i < saveCount; i++ {
			if vertex.IndexA == saveA[i] && vertex.IndexB == saveB[i] {
				duplicate = true
				break
			}
		}

		// If we found a duplicate support point we must exit to avoid cycling.
		if duplicate {
			break
		}

		// New vertex is valid and needed.
		simplex.Count += 1
	}

	if simplexIndex < len(simplexes) {
		simplexes[simplexIndex] = simplex
		simplexIndex++
	}

	// Prepare output in frame A
	normal := nonUnitNormal.Normalize()
	assert(normal.IsNormalized(), "b2IsNormalized( normal )")
	var localPointA, localPointB Vec2
	computeWitnessPoints(&simplex, &localPointA, &localPointB)
	output.Normal = normal
	output.Distance = localPointA.Distance(localPointB)
	output.PointA = localPointA
	output.PointB = localPointB
	output.Iterations = iteration
	output.SimplexCount = simplexIndex
	// Cache the simplex
	*cache = makeSimplexCache(&simplex)
	// Apply radii if requested
	if input.UseRadii {
		var radiusA float32 = input.ProxyA.Radius
		var radiusB float32 = input.ProxyB.Radius
		output.Distance = maxFloat(0.0, output.Distance-radiusA-radiusB)
		// Keep closest points on perimeter even if overlapped, this way the points move smoothly.
		output.PointA = output.PointA.MulAdd(radiusA, normal)
		output.PointB = output.PointB.MulSub(radiusB, normal)
	}

	return output
}

// Perform a linear shape cast of shape B moving and shape A fixed. Determines the hit point, normal, and translation fraction.
// The query runs in frame A, so the hit point and normal are returned in frame A. Initially touching shapes are a miss.
func ShapeCast(input *ShapeCastPairInput) CastOutput {
	// Compute tolerance
	var linearSlop float32 = LinearSlop()
	var totalRadius float32 = input.ProxyA.Radius + input.ProxyB.Radius
	var target float32 = maxFloat(linearSlop, totalRadius-linearSlop)
	var tolerance float32 = float32(0.25 * linearSlop)

	assert(target > tolerance, "target > tolerance")
	// Prepare input for distance query
	var cache SimplexCache

	var fraction float32 = 0.0

	var distanceInput DistanceInput
	distanceInput.ProxyA = input.ProxyA
	distanceInput.ProxyB = input.ProxyB
	distanceInput.UseRadii = false
	// The whole cast runs in frame A. Advance the relative pose of B in float each iteration,
	// which keeps the math near the local origin and avoids re-relativizing world poses.
	distanceInput.Transform = input.Transform
	delta2 := input.TranslationB
	var output CastOutput

	iteration := 0
	maxIterations := 20

	for ; iteration < maxIterations; iteration++ {
		output.Iterations += 1
		distanceOutput := ShapeDistance(&distanceInput, &cache, nil)

		if distanceOutput.Distance < target+tolerance {
			if iteration == 0 {
				if input.CanEncroach && distanceOutput.Distance > 2.0*linearSlop {
					target = distanceOutput.Distance - linearSlop
				} else {
					// Initial overlap
					output.Hit = true
					// Compute a common point
					c1 := distanceOutput.PointA.MulAdd(input.ProxyA.Radius, distanceOutput.Normal)
					c2 := distanceOutput.PointB.MulAdd(-input.ProxyB.Radius, distanceOutput.Normal)
					output.Point = c1.Lerp(c2, 0.5)
					return output
				}
			} else {
				// Regular hit
				assert(distanceOutput.Distance > 0.0 && distanceOutput.Normal.IsNormalized(), "distanceOutput.distance > 0.0f && b2IsNormalized( distanceOutput.normal )")
				output.Fraction = fraction
				output.Point = distanceOutput.PointA.MulAdd(input.ProxyA.Radius, distanceOutput.Normal)
				output.Normal = distanceOutput.Normal
				output.Hit = true
				return output
			}
		}

		assert(distanceOutput.Distance > 0.0, "distanceOutput.distance > 0.0f")
		assert(distanceOutput.Normal.IsNormalized(), "b2IsNormalized( distanceOutput.normal )")
		// Check if shapes are approaching each other
		var denominator float32 = delta2.Dot(distanceOutput.Normal)
		if denominator >= 0.0 {
			// Miss
			return output
		}

		// Advance sweep
		fraction += (target - distanceOutput.Distance) / denominator
		if fraction >= input.MaxFraction {
			// Miss
			return output
		}

		distanceInput.Transform.P = input.Transform.P.MulAdd(fraction, delta2)
	}

	// Failure!
	return output
}

func makeSeparationFunction(cache SimplexCache, proxyA *ShapeProxy, sweepA *Sweep, proxyB *ShapeProxy, sweepB *Sweep, t1 float32) separationFunction {
	var f separationFunction

	f.proxyA = proxyA
	f.proxyB = proxyB
	count := int(cache.Count)
	assert(0 < count && count < 3, "0 < count && count < 3")
	f.sweepA = *sweepA
	f.sweepB = *sweepB
	xfA := sweepA.Transform(t1)
	xfB := sweepB.Transform(t1)

	if count == 1 {
		f.type_ = pointsType
		localPointA := proxyA.Points[cache.IndexA[0]]
		localPointB := proxyB.Points[cache.IndexB[0]]
		pointA := xfA.Apply(localPointA)
		pointB := xfB.Apply(localPointB)
		f.axis = pointB.Sub(pointA).Normalize()
		f.localPoint = Vec2Zero
		return f
	}

	if cache.IndexA[0] == cache.IndexA[1] {
		// Two points on B and one on A.
		f.type_ = faceBType
		localPointB1 := proxyB.Points[cache.IndexB[0]]
		localPointB2 := proxyB.Points[cache.IndexB[1]]

		f.axis = localPointB2.Sub(localPointB1).CrossScalar(1.0)
		f.axis = f.axis.Normalize()
		normal := xfB.Q.Rotate(f.axis)

		f.localPoint = Vec2{0.5 * (localPointB1.X + localPointB2.X), 0.5 * (localPointB1.Y + localPointB2.Y)}
		pointB := xfB.Apply(f.localPoint)

		localPointA := proxyA.Points[cache.IndexA[0]]
		pointA := xfA.Apply(localPointA)

		var s float32 = pointA.Sub(pointB).Dot(normal)
		if s < 0.0 {
			f.axis = f.axis.Neg()
		}
		return f
	}

	// Two points on A and one or two points on B.
	f.type_ = faceAType
	localPointA1 := proxyA.Points[cache.IndexA[0]]
	localPointA2 := proxyA.Points[cache.IndexA[1]]

	f.axis = localPointA2.Sub(localPointA1).CrossScalar(1.0)
	f.axis = f.axis.Normalize()
	normal := xfA.Q.Rotate(f.axis)

	f.localPoint = Vec2{0.5 * (localPointA1.X + localPointA2.X), 0.5 * (localPointA1.Y + localPointA2.Y)}
	pointA := xfA.Apply(f.localPoint)

	localPointB := proxyB.Points[cache.IndexB[0]]
	pointB := xfB.Apply(localPointB)

	var s float32 = pointB.Sub(pointA).Dot(normal)
	if s < 0.0 {
		f.axis = f.axis.Neg()
	}
	return f
}

func findMinSeparation(f *separationFunction, indexA, indexB *int, t float32) float32 {
	xfA := f.sweepA.Transform(t)
	xfB := f.sweepB.Transform(t)

	switch f.type_ {
	case pointsType:
		{
			axisA := xfA.Q.InvRotate(f.axis)
			axisB := xfB.Q.InvRotate(f.axis.Neg())

			*indexA = findSupport(f.proxyA, axisA)
			*indexB = findSupport(f.proxyB, axisB)
			localPointA := f.proxyA.Points[*indexA]
			localPointB := f.proxyB.Points[*indexB]

			pointA := xfA.Apply(localPointA)
			pointB := xfB.Apply(localPointB)

			var separation float32 = pointB.Sub(pointA).Dot(f.axis)
			return separation
		}

	case faceAType:
		{
			normal := xfA.Q.Rotate(f.axis)
			pointA := xfA.Apply(f.localPoint)

			axisB := xfB.Q.InvRotate(normal.Neg())

			*indexA = -1
			*indexB = findSupport(f.proxyB, axisB)
			localPointB := f.proxyB.Points[*indexB]
			pointB := xfB.Apply(localPointB)

			var separation float32 = pointB.Sub(pointA).Dot(normal)
			return separation
		}

	case faceBType:
		{
			normal := xfB.Q.Rotate(f.axis)
			pointB := xfB.Apply(f.localPoint)

			axisA := xfA.Q.InvRotate(normal.Neg())

			*indexB = -1
			*indexA = findSupport(f.proxyA, axisA)
			localPointA := f.proxyA.Points[*indexA]
			pointA := xfA.Apply(localPointA)

			var separation float32 = pointA.Sub(pointB).Dot(normal)
			return separation
		}

	default:
		assert(false, "false")
		*indexA = -1
		*indexB = -1
		return 0.0
	}
}

func evaluateSeparation(f *separationFunction, indexA, indexB int, t float32) float32 {
	xfA := f.sweepA.Transform(t)
	xfB := f.sweepB.Transform(t)

	switch f.type_ {
	case pointsType:
		{
			localPointA := f.proxyA.Points[indexA]
			localPointB := f.proxyB.Points[indexB]

			pointA := xfA.Apply(localPointA)
			pointB := xfB.Apply(localPointB)

			var separation float32 = pointB.Sub(pointA).Dot(f.axis)
			return separation
		}

	case faceAType:
		{
			normal := xfA.Q.Rotate(f.axis)
			pointA := xfA.Apply(f.localPoint)

			localPointB := f.proxyB.Points[indexB]
			pointB := xfB.Apply(localPointB)

			var separation float32 = pointB.Sub(pointA).Dot(normal)
			return separation
		}

	case faceBType:
		{
			normal := xfB.Q.Rotate(f.axis)
			pointB := xfB.Apply(f.localPoint)

			localPointA := f.proxyA.Points[indexA]
			pointA := xfA.Apply(localPointA)

			var separation float32 = pointA.Sub(pointB).Dot(normal)
			return separation
		}

	default:
		assert(false, "false")
		return 0.0
	}
}

// Compute the upper bound on time before two shapes penetrate. Time is represented as
// a fraction between [0,tMax]. This uses a swept separating axis and may miss some intermediate,
// non-tunneling collisions. If you change the time interval, you should call this function
// again.
func TimeOfImpact(input *TOIInput) TOIOutput {
	var output TOIOutput
	output.State = TOIStateUnknown
	output.Fraction = input.MaxFraction
	sweepA := input.SweepA
	sweepB := input.SweepB
	assert(sweepA.Q1.IsNormalized() && sweepA.Q2.IsNormalized(), "b2IsNormalizedRot( sweepA.q1 ) && b2IsNormalizedRot( sweepA.q2 )")
	assert(sweepB.Q1.IsNormalized() && sweepB.Q2.IsNormalized(), "b2IsNormalizedRot( sweepB.q1 ) && b2IsNormalizedRot( sweepB.q2 )")
	// todo_erin
	// c1 can be at the origin yet the points are far away
	// Vec2 origin = Add(sweepA.C1, input.ProxyA.Points[0])
	proxyA := &input.ProxyA
	proxyB := &input.ProxyB

	var tMax float32 = input.MaxFraction

	// Setup target distance and tolerance
	var totalRadius float32 = proxyA.Radius + proxyB.Radius
	var target float32 = maxFloat(LinearSlop(), totalRadius-LinearSlop())
	var tolerance float32 = float32(0.25 * LinearSlop())
	assert(target > tolerance, "target > tolerance")
	var t1 float32 = 0.0
	k_maxIterations := 20
	distanceIterations := 0

	// Prepare input for distance query.
	var cache SimplexCache
	var distanceInput DistanceInput
	distanceInput.ProxyA = input.ProxyA
	distanceInput.ProxyB = input.ProxyB
	distanceInput.UseRadii = false
	// The outer loop progressively attempts to compute new separating axes.
	// This loop terminates when an axis is repeated (no progress is made).
	for {
		// Get the distance between shapes. We can also use the results to get a separating axis.
		xfA := sweepA.Transform(t1)
		xfB := sweepB.Transform(t1)
		distanceInput.Transform = xfA.InvMul(xfB)
		distanceOutput := ShapeDistance(&distanceInput, &cache, nil)

		// The distance query runs in frame A, project the witness data back to world
		worldNormal := xfA.Q.Rotate(distanceOutput.Normal)
		worldPointA := xfA.Apply(distanceOutput.PointA)
		worldPointB := xfA.Apply(distanceOutput.PointB)

		// Progressive time of impact. This handles slender geometry well but introduces
		// significant time loss.
		// if (distanceIterations == 0)
		//{
		//	if ( distanceOutput.Distance > totalRadius + B2_SPECULATIVE_DISTANCE )
		//	{
		//		target = totalRadius + B2_SPECULATIVE_DISTANCE - tolerance
		//	}
		//	else
		//	{
		//		target = distanceOutput.Distance - 1.5 * tolerance
		//		target = MaxFloat( target, 2.0 * tolerance )
		//	}
		//}

		distanceIterations += 1
		// If the shapes are overlapped, we give up on continuous collision.
		if distanceOutput.Distance <= 0.0 {
			// Failure!
			output.State = TOIStateOverlapped
			output.Fraction = 0.0
			break
		}

		if distanceOutput.Distance <= target+tolerance {
			// Success!
			output.State = TOIStateHit
			// Averaged hit point
			pA := worldPointA.MulAdd(proxyA.Radius, worldNormal)
			pB := worldPointB.MulAdd(-proxyB.Radius, worldNormal)
			output.Point = pA.Lerp(pB, 0.5)
			output.Normal = worldNormal
			output.Fraction = t1
			break
		}

		// Initialize the separating axis.
		fcn := makeSeparationFunction(cache, proxyA, &sweepA, proxyB, &sweepB, t1)

		// Compute the TOI on the separating axis. We do this by successively
		// resolving the deepest point. This loop is bounded by the number of vertices.
		done := false
		var t2 float32 = tMax
		pushBackIterations := 0
		for {
			// Find the deepest point at t2. Store the witness point indices.
			var indexA, indexB int
			var s2 float32 = findMinSeparation(&fcn, &indexA, &indexB, t2)

			// Is the final configuration separated?
			if s2 > target+tolerance {
				// Victory!
				output.State = TOIStateSeparated
				output.Fraction = tMax
				done = true
				break
			}

			// Has the separation reached tolerance?
			if s2 > target-tolerance {
				// Advance the sweeps
				t1 = t2
				break
			}

			// Compute the initial separation of the witness points.
			var s1 float32 = evaluateSeparation(&fcn, indexA, indexB, t1)

			// Check for initial overlap. This might happen if the root finder runs out of iterations.
			if s1 < target-tolerance {
				output.State = TOIStateFailed
				output.Fraction = t1
				done = true
				break
			}

			// Check for touching
			if s1 <= target+tolerance {
				// Success! t1 should hold the TOI (could be 0.0).
				output.State = TOIStateHit
				// Averaged hit point
				pA := worldPointA.MulAdd(proxyA.Radius, worldNormal)
				pB := worldPointB.MulAdd(-proxyB.Radius, worldNormal)
				output.Point = pA.Lerp(pB, 0.5)
				output.Normal = worldNormal
				output.Fraction = t1
				done = true
				break
			}

			// Compute 1D root of: f(x) - target = 0
			rootIterationCount := 0
			a1, a2 := t1, t2
			for {
				// Use a mix of false position and bisection.
				var t float32
				if rootIterationCount&1 != 0 {
					// False position to improve convergence.
					t = a1 + float32((target-s1)*(a2-a1))/(s2-s1)
				} else {
					// Bisection to guarantee progress.
					t = 0.5 * (a1 + a2)
				}

				rootIterationCount += 1
				var s float32 = evaluateSeparation(&fcn, indexA, indexB, t)

				// Has the separation reached tolerance?
				if absFloat(s-target) < tolerance {
					// t2 holds a tentative value for t1
					t2 = t
					break
				}

				// Ensure we continue to bracket the root.
				if s > target {
					a1 = t
					s1 = s
				} else {
					a2 = t
					s2 = s
				}

				if rootIterationCount == 50 {
					break
				}
			}

			pushBackIterations += 1
			if pushBackIterations == MaxPolygonVertices {
				break
			}
		}

		if done {
			break
		}

		if distanceIterations == k_maxIterations {
			// Root finder got stuck. Semi-victory.
			output.State = TOIStateFailed
			// Averaged hit point
			pA := worldPointA.MulAdd(proxyA.Radius, worldNormal)
			pB := worldPointB.MulAdd(-proxyB.Radius, worldNormal)
			output.Point = pA.Lerp(pB, 0.5)
			output.Normal = worldNormal
			output.Fraction = t1
			break
		}
	}

	return output
}
