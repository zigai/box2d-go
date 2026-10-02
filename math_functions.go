// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "math"

// @return the minimum of two integers
func MinInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

// @return the maximum of two integers
func MaxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// @return the absolute value of an integer
func AbsInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// @return an integer clamped between a lower and upper bound
func ClampInt(a int, lower int, upper int) int {
	if a < lower {
		return lower
	}
	if a > upper {
		return upper
	}
	return a
}

// https://en.wikipedia.org/wiki/Floor_and_ceiling_functions
func CeilingInt(numerator int, denominator int) int {
	validate(denominator > 0 && numerator >= 0, "denominator > 0 && numerator >= 0")
	return (numerator + denominator - 1) / denominator
}

// @return the minimum of two floats
func MinFloat(a float32, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

// @return the maximum of two floats
func MaxFloat(a float32, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// @return the absolute value of a float
func AbsFloat(a float32) float32 {
	if a < 0 {
		return -a
	}
	return a
}

// @return a float clamped between a lower and upper bound
func ClampFloat(a float32, lower float32, upper float32) float32 {
	if a < lower {
		return lower
	}
	if a > upper {
		return upper
	}
	return a
}

// Vector dot product
func Dot(a Vec2, b Vec2) float32 {
	return float32(a.X*b.X) + float32(a.Y*b.Y)
}

// Vector cross product. In 2D this yields a scalar.
func Cross(a Vec2, b Vec2) float32 {
	return float32(a.X*b.Y) - float32(a.Y*b.X)
}

// Perform the cross product on a vector and a scalar. In 2D this produces a vector.
func CrossVS(v Vec2, s float32) Vec2 {
	return Vec2{s * v.Y, -s * v.X}
}

// Perform the cross product on a scalar and a vector. In 2D this produces a vector.
func CrossSV(s float32, v Vec2) Vec2 {
	return Vec2{-s * v.Y, s * v.X}
}

// Get a left pointing perpendicular vector. Equivalent to CrossSV(1.0f, v)
func LeftPerp(v Vec2) Vec2 {
	return Vec2{-v.Y, v.X}
}

// Get a right pointing perpendicular vector. Equivalent to CrossVS(v, 1.0f)
func RightPerp(v Vec2) Vec2 {
	return Vec2{v.Y, -v.X}
}

// Vector addition
func Add(a Vec2, b Vec2) Vec2 {
	return Vec2{a.X + b.X, a.Y + b.Y}
}

// Vector subtraction
func Sub(a Vec2, b Vec2) Vec2 {
	return Vec2{a.X - b.X, a.Y - b.Y}
}

// Vector negation
func Neg(a Vec2) Vec2 {
	return Vec2{-a.X, -a.Y}
}

// Vector linear interpolation
// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
func Lerp(a Vec2, b Vec2, t float32) Vec2 {
	return Vec2{float32((1-t)*a.X) + float32(t*b.X), float32((1-t)*a.Y) + float32(t*b.Y)}
}

// Component-wise multiplication
func Mul(a Vec2, b Vec2) Vec2 {
	return Vec2{a.X * b.X, a.Y * b.Y}
}

// Multiply a scalar and vector
func MulSV(s float32, v Vec2) Vec2 {
	return Vec2{s * v.X, s * v.Y}
}

// a + s * b
func MulAdd(a Vec2, s float32, b Vec2) Vec2 {
	return Vec2{a.X + float32(s*b.X), a.Y + float32(s*b.Y)}
}

// a - s * b
func MulSub(a Vec2, s float32, b Vec2) Vec2 {
	return Vec2{a.X - float32(s*b.X), a.Y - float32(s*b.Y)}
}

// Component-wise absolute vector
func Abs(a Vec2) Vec2 {
	return Vec2{AbsFloat(a.X), AbsFloat(a.Y)}
}

// Component-wise minimum vector
func Min(a Vec2, b Vec2) Vec2 {
	return Vec2{MinFloat(a.X, b.X), MinFloat(a.Y, b.Y)}
}

// Component-wise maximum vector
func Max(a Vec2, b Vec2) Vec2 {
	return Vec2{MaxFloat(a.X, b.X), MaxFloat(a.Y, b.Y)}
}

// Component-wise clamp vector v into the range [a, b]
func Clamp(v Vec2, a Vec2, b Vec2) Vec2 {
	return Vec2{ClampFloat(v.X, a.X, b.X), ClampFloat(v.Y, a.Y, b.Y)}
}

// Get the length of this vector (the norm)
func Length(v Vec2) float32 {
	return float32(math.Sqrt(float64(float32(v.X*v.X) + float32(v.Y*v.Y))))
}

// Get the distance between two points
func Distance(a Vec2, b Vec2) float32 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return float32(math.Sqrt(float64(float32(dx*dx) + float32(dy*dy))))
}

// Convert a vector into a unit vector if possible, otherwise returns the zero vector.
// todo MSVC is not inlining this function in several places per warning 4710
func Normalize(a Vec2) Vec2 {
	lengthSquared := float32(a.X*a.X) + float32(a.Y*a.Y)
	if lengthSquared > float32(1000)*float32(0x1p-126) {
		s := 1 / float32(math.Sqrt(float64(lengthSquared)))
		return Vec2{s * a.X, s * a.Y}
	}
	return Vec2{}
}

// Determines if the provided vector is normalized (norm(a) == 1).
func IsNormalized(a Vec2) bool {
	aa := Dot(a, a)
	return AbsFloat(1-aa) < float32(100)*float32(0x1p-23)
}

// Convert a vector into a unit vector if possible, otherwise returns the zero vector. Also
// outputs the length.
func GetLengthAndNormalize(length *float32, a Vec2) Vec2 {
	lengthSquared := float32(a.X*a.X) + float32(a.Y*a.Y)
	if lengthSquared > float32(1000)*float32(0x1p-126) {
		*length = float32(math.Sqrt(float64(lengthSquared)))
		s := 1 / *length
		return Vec2{s * a.X, s * a.Y}
	}
	*length = 0
	return Vec2{}
}

// Normalize rotation
func NormalizeRot(q Rot) Rot {
	mag := float32(math.Sqrt(float64(float32(q.S*q.S) + float32(q.C*q.C))))
	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}
	return Rot{q.C * invMag, q.S * invMag}
}

// Integrate rotation from angular velocity
// @param q1 initial rotation
// @param deltaAngle the angular displacement in radians
func IntegrateRotation(q1 Rot, deltaAngle float32) Rot {
	q2 := Rot{q1.C - float32(deltaAngle*q1.S), q1.S + float32(deltaAngle*q1.C)}
	mag := float32(math.Sqrt(float64(float32(q2.S*q2.S) + float32(q2.C*q2.C))))
	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}
	return Rot{q2.C * invMag, q2.S * invMag}
}

// Get the length squared of this vector
func LengthSquared(v Vec2) float32 {
	return float32(v.X*v.X) + float32(v.Y*v.Y)
}

// Get the distance squared between points
func DistanceSquared(a Vec2, b Vec2) float32 {
	c := Vec2{b.X - a.X, b.Y - a.Y}
	return float32(c.X*c.X) + float32(c.Y*c.Y)
}

// Make a rotation using an angle in radians
func MakeRot(radians float32) Rot {
	cs := ComputeCosSin(radians)
	return Rot{cs.Cosine, cs.Sine}
}

// Make a rotation using a unit vector
func MakeRotFromUnitVector(unitVector Vec2) Rot {
	assert(IsNormalized(unitVector), "IsNormalized(unitVector)")
	return Rot{unitVector.X, unitVector.Y}
}

// Is this rotation normalized?
func IsNormalizedRot(q Rot) bool {
	qq := float32(q.S*q.S) + float32(q.C*q.C)
	return float32(1)-float32(0.0006) < qq && qq < float32(1)+float32(0.0006)
}

// Get the inverse of a rotation
func InvertRot(a Rot) Rot {
	return Rot{a.C, -a.S}
}

// Normalized linear interpolation
// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
// https://web.archive.org/web/20170825184056/http://number-none.com/product/Understanding%20Slerp,%20Then%20Not%20Using%20It/
func NLerp(q1 Rot, q2 Rot, t float32) Rot {
	omt := 1 - t
	q := Rot{float32(omt*q1.C) + float32(t*q2.C), float32(omt*q1.S) + float32(t*q2.S)}
	mag := float32(math.Sqrt(float64(float32(q.S*q.S) + float32(q.C*q.C))))
	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}
	return Rot{q.C * invMag, q.S * invMag}
}

// Compute the angular velocity necessary to rotate between two rotations over a give time
// @param q1 initial rotation
// @param q2 final rotation
// @param invH inverse time step
func ComputeAngularVelocity(q1 Rot, q2 Rot, invH float32) float32 {
	omega := invH * (float32(q2.S*q1.C) - float32(q2.C*q1.S))
	return omega
}

// Get the angle in radians in the range [-pi, pi]
func (q Rot) Angle() float32 {
	return Atan2(q.S, q.C)
}

// Get the x-axis
func (q Rot) XAxis() Vec2 {
	return Vec2{q.C, q.S}
}

// Get the y-axis
func (q Rot) YAxis() Vec2 {
	return Vec2{-q.S, q.C}
}

// Multiply two rotations: q * r
func MulRot(q Rot, r Rot) Rot {
	var qr Rot
	qr.S = float32(q.S*r.C) + float32(q.C*r.S)
	qr.C = float32(q.C*r.C) - float32(q.S*r.S)
	return qr
}

// Transpose multiply two rotations: inv(a) * b
// This rotates a vector local in frame b into a vector local in frame a
func InvMulRot(a Rot, b Rot) Rot {
	var r Rot
	r.S = float32(a.C*b.S) - float32(a.S*b.C)
	r.C = float32(a.C*b.C) + float32(a.S*b.S)
	return r
}

// Relative angle between a and b.
func RelativeAngle(a Rot, b Rot) float32 {
	s := float32(a.C*b.S) - float32(a.S*b.C)
	c := float32(a.C*b.C) + float32(a.S*b.S)
	return Atan2(s, c)
}

// Convert any angle into the range [-pi, pi].
func UnwindAngle(radians float32) float32 {
	validate(AbsFloat(radians) < 1e4, "AbsFloat(radians) < 1.0e4f")
	x := ClampFloat(radians, -1e6, 1e6)
	twoPi := float64(float32(2) * Pi)
	const roundToNearest float64 = 6755399441055744
	a := float64(x)
	k := (a/twoPi + roundToNearest) - roundToNearest
	return float32(a - float64(k*twoPi))
}

// Rotate a vector.
func RotateVector(q Rot, v Vec2) Vec2 {
	return Vec2{float32(q.C*v.X) - float32(q.S*v.Y), float32(q.S*v.X) + float32(q.C*v.Y)}
}

// Inverse rotate a vector.
func InvRotateVector(q Rot, v Vec2) Vec2 {
	return Vec2{float32(q.C*v.X) + float32(q.S*v.Y), float32(-q.S*v.X) + float32(q.C*v.Y)}
}

// Transform a point (e.g. local space to world space).
func TransformPoint(t Transform, p Vec2) Vec2 {
	x := (float32(t.Q.C*p.X) - float32(t.Q.S*p.Y)) + t.P.X
	y := (float32(t.Q.S*p.X) + float32(t.Q.C*p.Y)) + t.P.Y
	return Vec2{x, y}
}

// Inverse transform a point (e.g. world space to local space).
func InvTransformPoint(t Transform, p Vec2) Vec2 {
	vx := p.X - t.P.X
	vy := p.Y - t.P.Y
	return Vec2{float32(t.Q.C*vx) + float32(t.Q.S*vy), float32(-t.Q.S*vx) + float32(t.Q.C*vy)}
}

// Multiply two transforms. If the result is applied to a point p local to frame B,
// the transform would first convert p to a point local to frame A, then into a point
// in the world frame.
// v2 = a.Q.Rot(b.Q.Rot(v1) + b.P) + a.P
//
//	= (a.Q * b.Q).Rot(v1) + a.Q.Rot(b.P) + a.P
func MulTransforms(a Transform, b Transform) Transform {
	var c Transform
	c.Q = MulRot(a.Q, b.Q)
	c.P = Add(RotateVector(a.Q, b.P), a.P)
	return c
}

// Creates a transform that converts a local point in frame B to a local point in frame A.
// v2 = a.Q' * (b.Q * v1 + b.P - a.P)
//
//	= a.Q' * b.Q * v1 + a.Q' * (b.P - a.P)
func InvMulTransforms(a Transform, b Transform) Transform {
	var c Transform
	c.Q = InvMulRot(a.Q, b.Q)
	c.P = InvRotateVector(a.Q, Sub(b.P, a.P))
	return c
}

// Lossy conversion of a world position to a float vector. no-op in single precision.
func ToVec2(p Pos) Vec2 {
	return Vec2{float32(p.X), float32(p.Y)}
}

// Narrow a world coordinate to float, rounding toward negative infinity. Use with
// RoundUpFloat to build a conservative float box that always contains double bounds,
// where plain rounding far from the origin could clip. nextafterf is an exact IEEE
// operation, so this is cross-platform deterministic. With large world mode off this is
// a plain conversion.
func RoundDownFloat(x float64) float32 {
	f := float32(x)
	if IsDoublePrecision() && float64(f) > x {
		return math.Nextafter32(f, -math.MaxFloat32)
	}
	return f
}

// Narrow a world coordinate to float, rounding toward positive infinity.
func RoundUpFloat(x float64) float32 {
	f := float32(x)
	if IsDoublePrecision() && float64(f) < x {
		return math.Nextafter32(f, math.MaxFloat32)
	}
	return f
}

// a - b, demoted to float. The primary precision boundary operation.
func SubPos(a Pos, b Pos) Vec2 {
	return Vec2{float32(a.X - b.X), float32(a.Y - b.Y)}
}

// Transform a local point to a world position. Rotation in float, translation in double.
func TransformWorldPoint(t WorldTransform, p Vec2) Pos {
	rx := float32(t.Q.C*p.X) - float32(t.Q.S*p.Y)
	ry := float32(t.Q.S*p.X) + float32(t.Q.C*p.Y)
	return OffsetPos(t.P, Vec2{rx, ry})
}

// Transform a world position to a local point. One double subtraction, then float.
func InvTransformWorldPoint(t WorldTransform, p Pos) Vec2 {
	vx := float32(p.X - t.P.X)
	vy := float32(p.Y - t.P.Y)
	return Vec2{float32(t.Q.C*vx) + float32(t.Q.S*vy), float32(-t.Q.S*vx) + float32(t.Q.C*vy)}
}

// Relative transform of frame B in frame A.
func InvMulWorldTransforms(a WorldTransform, b WorldTransform) Transform {
	var c Transform
	c.Q = InvMulRot(a.Q, b.Q)
	d := Vec2{float32(b.P.X - a.P.X), float32(b.P.Y - a.P.Y)}
	c.P = InvRotateVector(a.Q, d)
	return c
}

// Convert a local transform B into world space using world transform A.
func MulWorldTransforms(a WorldTransform, b Transform) WorldTransform {
	var c WorldTransform
	c.Q = MulRot(a.Q, b.Q)
	c.P = OffsetPos(a.P, RotateVector(a.Q, b.P))
	return c
}

// Shift a world transform into the frame of a base position.
func ToRelativeTransform(t WorldTransform, base Pos) Transform {
	var r Transform
	r.Q = t.Q
	r.P = Vec2{float32(t.P.X - base.X), float32(t.P.Y - base.Y)}
	return r
}

// Promote a float transform to a world transform. Lossless.
func MakeWorldTransform(t Transform) WorldTransform {
	var w WorldTransform
	w.P = ToPos(t.P)
	w.Q = t.Q
	return w
}

// Multiply a 2-by-2 matrix times a 2D vector
func MulMV(a Mat22, v Vec2) Vec2 {
	return Vec2{float32(a.Cx.X*v.X) + float32(a.Cy.X*v.Y), float32(a.Cx.Y*v.X) + float32(a.Cy.Y*v.Y)}
}

// Get the inverse of a 2-by-2 matrix
func GetInverse22(a Mat22) Mat22 {
	aa, b, c, d := a.Cx.X, a.Cy.X, a.Cx.Y, a.Cy.Y
	det := float32(aa*d) - float32(b*c)
	if det != 0 {
		det = 1 / det
	}
	return Mat22{Vec2{det * d, -det * c}, Vec2{-det * b, det * aa}}
}

// Solve A * x = b, where b is a column vector. This is more efficient
// than computing the inverse in one-shot cases.
func Solve22(a Mat22, b Vec2) Vec2 {
	a11, a12, a21, a22 := a.Cx.X, a.Cy.X, a.Cx.Y, a.Cy.Y
	det := float32(a11*a22) - float32(a12*a21)
	if det != 0 {
		det = 1 / det
	}
	return Vec2{det * (float32(a22*b.X) - float32(a12*b.Y)), det * (float32(a11*b.Y) - float32(a21*b.X))}
}

// Does a fully contain b
func (a AABB) Contains(b AABB) bool {
	s := true
	s = s && a.LowerBound.X <= b.LowerBound.X
	s = s && a.LowerBound.Y <= b.LowerBound.Y
	s = s && b.UpperBound.X <= a.UpperBound.X
	s = s && b.UpperBound.Y <= a.UpperBound.Y
	return s
}

// Get the center of the AABB.
func (a AABB) Center() Vec2 {
	return Vec2{0.5 * (a.LowerBound.X + a.UpperBound.X), 0.5 * (a.LowerBound.Y + a.UpperBound.Y)}
}

// Get the extents of the AABB (half-widths).
func (a AABB) Extents() Vec2 {
	return Vec2{0.5 * (a.UpperBound.X - a.LowerBound.X), 0.5 * (a.UpperBound.Y - a.LowerBound.Y)}
}

// Union of two AABBs
func (a AABB) Union(b AABB) AABB {
	var c AABB
	c.LowerBound.X = MinFloat(a.LowerBound.X, b.LowerBound.X)
	c.LowerBound.Y = MinFloat(a.LowerBound.Y, b.LowerBound.Y)
	c.UpperBound.X = MaxFloat(a.UpperBound.X, b.UpperBound.X)
	c.UpperBound.Y = MaxFloat(a.UpperBound.Y, b.UpperBound.Y)
	return c
}

// Do a and b overlap
func (a AABB) Overlaps(b AABB) bool {
	return !(b.LowerBound.X > a.UpperBound.X || b.LowerBound.Y > a.UpperBound.Y || a.LowerBound.X > b.UpperBound.X || a.LowerBound.Y > b.UpperBound.Y)
}

// Compute the bounding box of an array of circles
func MakeAABB(points []Vec2, count int, radius float32) AABB {
	assert(count > 0, "count > 0")
	a := AABB{points[0], points[0]}
	for i := 1; i < count; i++ {
		a.LowerBound = Min(a.LowerBound, points[i])
		a.UpperBound = Max(a.UpperBound, points[i])
	}
	r := Vec2{radius, radius}
	a.LowerBound = Sub(a.LowerBound, r)
	a.UpperBound = Add(a.UpperBound, r)
	return a
}

// Signed separation of a point from a plane
func PlaneSeparation(plane Plane, point Vec2) float32 {
	return Dot(plane.Normal, point) - plane.Offset
}

// One-dimensional mass-spring-damper simulation. Returns the new velocity given the position and time step.
// You can then compute the new position using:
// position += timeStep * newVelocity
// This drives towards a zero position. By using implicit integration we get a stable solution
// that doesn't require transcendental functions.
func SpringDamper(hertz float32, dampingRatio float32, position float32, velocity float32, timeStep float32) float32 {
	omega := 2 * Pi * hertz
	omegaH := omega * timeStep
	return (velocity - float32(omega*omegaH*position)) / (1 + float32(2*dampingRatio*omegaH) + float32(omegaH*omegaH))
}

// Is this a valid number? Not NaN or infinity.
func IsValidFloat(a float32) bool {
	return !math.IsNaN(float64(a)) && !math.IsInf(float64(a), 0)
}

// Is this a valid vector? Not NaN or infinity.
func IsValidVec2(v Vec2) bool {
	if math.IsNaN(float64(v.X)) || math.IsNaN(float64(v.Y)) {
		return false
	}
	if math.IsInf(float64(v.X), 0) || math.IsInf(float64(v.Y), 0) {
		return false
	}
	return true
}

// Is this a valid rotation? Not NaN or infinity. Is normalized.
func IsValidRotation(q Rot) bool {
	if math.IsNaN(float64(q.S)) || math.IsNaN(float64(q.C)) {
		return false
	}
	if math.IsInf(float64(q.S), 0) || math.IsInf(float64(q.C), 0) {
		return false
	}
	return IsNormalizedRot(q)
}

// Is this a valid transform? Not NaN or infinity. Rotation is normalized.
func IsValidTransform(t Transform) bool {
	if !IsValidVec2(t.P) {
		return false
	}
	return IsValidRotation(t.Q)
}

// Is this a valid plane? Normal is a unit vector. Not Nan or infinity.
func IsValidPlane(a Plane) bool {
	return IsValidVec2(a.Normal) && IsNormalized(a.Normal) && IsValidFloat(a.Offset)
}

// Is this a valid world position? Not NaN or infinity.
func IsValidPosition(p Pos) bool {
	if math.IsNaN(float64(p.X)) || math.IsNaN(float64(p.Y)) {
		return false
	}
	if math.IsInf(float64(p.X), 0) || math.IsInf(float64(p.Y), 0) {
		return false
	}
	return true
}

// Is this a valid world transform? Not NaN or infinity. Rotation is normalized.
func IsValidWorldTransform(t WorldTransform) bool {
	if !IsValidPosition(t.P) {
		return false
	}
	return IsValidRotation(t.Q)
}

// Compute an approximate arctangent in the range [-pi, pi]
// This is hand coded for cross-platform determinism. The atan2f
// function in the standard library is not cross-platform deterministic.
// Accurate to around 0.0023 degrees
func Atan2(y float32, x float32) float32 {
	if x == 0 && y == 0 {
		return 0
	}
	ax := AbsFloat(x)
	ay := AbsFloat(y)
	mx := MaxFloat(ay, ax)
	mn := MinFloat(ay, ax)
	a := mn / mx
	s := a * a
	c := s * a
	q := s * s
	r := float32(float32(0.024840285)*q) + float32(0.18681418)
	t := float32(float32(-0.094097948)*q) - float32(0.33213072)
	r = float32(r*s) + t
	r = float32(r*c) + a
	if ay > ax {
		r = float32(1.57079637) - r
	}
	if x < 0 {
		r = float32(3.14159274) - r
	}
	if y < 0 {
		r = -r
	}
	return r
}

// Compute the cosine and sine of an angle in radians. Implemented
// for cross-platform determinism.
func ComputeCosSin(radians float32) CosSin {
	x := UnwindAngle(radians)
	pi2 := float32(Pi * Pi)
	var c float32
	if x < -0.5*Pi {
		y := x + Pi
		y2 := float32(y * y)
		c = -(pi2 - float32(4*y2)) / (pi2 + y2)
	} else if x > 0.5*Pi {
		y := x - Pi
		y2 := float32(y * y)
		c = -(pi2 - float32(4*y2)) / (pi2 + y2)
	} else {
		y2 := float32(x * x)
		c = (pi2 - float32(4*y2)) / (pi2 + y2)
	}
	var s float32
	if x < 0 {
		y := x + Pi
		s = -16 * y * (Pi - y) / (float32(5*pi2) - float32(4*y*(Pi-y)))
	} else {
		s = 16 * x * (Pi - x) / (float32(5*pi2) - float32(4*x*(Pi-x)))
	}
	mag := float32(math.Sqrt(float64(float32(s*s) + float32(c*c))))
	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}
	return CosSin{c * invMag, s * invMag}
}

// Compute the rotation between two unit vectors
func ComputeRotationBetweenUnitVectors(v1 Vec2, v2 Vec2) Rot {
	assert(AbsFloat(1-Length(v1)) < float32(100)*float32(0x1p-23), "AbsFloat(1.0f - Length(v1)) < 100.0f * FLT_EPSILON")
	assert(AbsFloat(1-Length(v2)) < float32(100)*float32(0x1p-23), "AbsFloat(1.0f - Length(v2)) < 100.0f * FLT_EPSILON")
	var rot Rot
	rot.C = Dot(v1, v2)
	rot.S = Cross(v1, v2)
	return NormalizeRot(rot)
}

// Convert a vector to a world position. no-op in single precision.
func ToPos(v Vec2) Pos {
	return Pos{posScalar(v.X), posScalar(v.Y)}
}

// p + d
func OffsetPos(p Pos, d Vec2) Pos {
	return Pos{p.X + posScalar(d.X), p.Y + posScalar(d.Y)}
}

// World position interpolation for sweeps and sampling.
func LerpPosition(a Pos, b Pos, t float32) Pos {
	return Pos{posScalar(posScalar(1-t)*a.X) + posScalar(posScalar(t)*b.X), posScalar(posScalar(1-t)*a.Y) + posScalar(posScalar(t)*b.Y)}
}
