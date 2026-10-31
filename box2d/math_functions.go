// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"math"

	"github.com/zigai/box2d-go/internal/core"
)

// Vector dot product.
func (a Vec2) Dot(b Vec2) float32 {
	return float32(a.X*b.X) + float32(a.Y*b.Y)
}

// Vector cross product. In 2D this yields a scalar.
func (a Vec2) Cross(b Vec2) float32 {
	return float32(a.X*b.Y) - float32(a.Y*b.X)
}

// Perform the cross product on a vector and a scalar. In 2D this produces a vector.
func (a Vec2) CrossScalar(s float32) Vec2 {
	return Vec2{s * a.Y, -s * a.X}
}

// Perform the cross product on a scalar and a vector. In 2D this produces a vector.
func (a Vec2) ScalarCross(s float32) Vec2 {
	return Vec2{float32(-s * a.Y), float32(s * a.X)}
}

// Get a left pointing perpendicular vector. Equivalent to CrossSV(1.0f, v).
func (a Vec2) LeftPerp() Vec2 {
	return Vec2{-a.Y, a.X}
}

// Get a right pointing perpendicular vector. Equivalent to CrossVS(v, 1.0f).
func (a Vec2) RightPerp() Vec2 {
	return Vec2{a.Y, -a.X}
}

// Vector addition.
func (a Vec2) Add(b Vec2) Vec2 {
	return Vec2{a.X + b.X, a.Y + b.Y}
}

// Vector subtraction.
func (a Vec2) Sub(b Vec2) Vec2 {
	return Vec2{a.X - b.X, a.Y - b.Y}
}

// Vector negation.
func (a Vec2) Neg() Vec2 {
	return Vec2{-a.X, -a.Y}
}

// Vector linear interpolation
// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
func (a Vec2) Lerp(b Vec2, t float32) Vec2 {
	return Vec2{float32((1-t)*a.X) + float32(t*b.X), float32((1-t)*a.Y) + float32(t*b.Y)}
}

// Component-wise multiplication.
func (a Vec2) Mul(b Vec2) Vec2 {
	return Vec2{a.X * b.X, a.Y * b.Y}
}

// Multiply a scalar and vector.
func (a Vec2) Scale(s float32) Vec2 {
	return Vec2{float32(s * a.X), float32(s * a.Y)}
}

// a + s * b.
func (a Vec2) MulAdd(s float32, b Vec2) Vec2 {
	return Vec2{a.X + float32(s*b.X), a.Y + float32(s*b.Y)}
}

// a - s * b.
func (a Vec2) MulSub(s float32, b Vec2) Vec2 {
	return Vec2{a.X - float32(s*b.X), a.Y - float32(s*b.Y)}
}

// Component-wise absolute vector.
func (a Vec2) Abs() Vec2 {
	return Vec2{absFloat(a.X), absFloat(a.Y)}
}

// Component-wise minimum vector.
func (a Vec2) Min(b Vec2) Vec2 {
	return Vec2{minFloat(a.X, b.X), minFloat(a.Y, b.Y)}
}

// Component-wise maximum vector.
func (a Vec2) Max(b Vec2) Vec2 {
	return Vec2{maxFloat(a.X, b.X), maxFloat(a.Y, b.Y)}
}

// Component-wise clamp vector v into the range [a, b].
func (a Vec2) Clamp(lower Vec2, b Vec2) Vec2 {
	return Vec2{clampFloat(a.X, lower.X, b.X), clampFloat(a.Y, lower.Y, b.Y)}
}

// Get the length of this vector (the norm).
func (a Vec2) Length() float32 {
	return float32(math.Sqrt(float64(float32(a.X*a.X) + float32(a.Y*a.Y))))
}

// Get the distance between two points.
func (a Vec2) Distance(b Vec2) float32 {
	dx := b.X - a.X
	dy := b.Y - a.Y

	return float32(math.Sqrt(float64(float32(dx*dx) + float32(dy*dy))))
}

// Convert a vector into a unit vector if possible, otherwise returns the zero vector.
func (a Vec2) Normalize() Vec2 {
	lengthSquared := float32(a.X*a.X) + float32(a.Y*a.Y)
	if math.IsInf(float64(lengthSquared), 1) {
		_, direction := a.LengthAndNormalize()
		return direction
	}

	if lengthSquared > float32(1000)*float32(0x1p-126) {
		s := 1 / float32(math.Sqrt(float64(lengthSquared)))
		return Vec2{s * a.X, s * a.Y}
	}

	return Vec2{}
}

// Determines if the provided vector is normalized (norm(a) == 1).
func (a Vec2) IsNormalized() bool {
	aa := a.Dot(a)
	return absFloat(1-aa) < float32(100)*float32(0x1p-23)
}

// LengthAndNormalize returns the length of a and a unit vector in its direction, or zero and the
// zero vector if a is too short to normalize.
func (a Vec2) LengthAndNormalize() (float32, Vec2) {
	lengthSquared := float32(a.X*a.X) + float32(a.Y*a.Y)
	if math.IsInf(float64(lengthSquared), 1) {
		length := math.Hypot(float64(a.X), float64(a.Y))
		return float32(length), Vec2{float32(float64(a.X) / length), float32(float64(a.Y) / length)}
	}

	if lengthSquared > float32(1000)*float32(0x1p-126) {
		length := float32(math.Sqrt(float64(lengthSquared)))
		s := 1 / length

		return length, Vec2{s * a.X, s * a.Y}
	}

	return 0, Vec2{}
}

// Normalize rotation.
func (q Rot) Normalize() Rot {
	mag := float32(math.Sqrt(float64(float32(q.S*q.S) + float32(q.C*q.C))))

	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}

	return Rot{q.C * invMag, q.S * invMag}
}

// Integrate rotation from angular velocity
// @param q1 initial rotation
// @param deltaAngle the angular displacement in radians.
func (q Rot) Integrate(deltaAngle float32) Rot {
	q2 := Rot{q.C - float32(deltaAngle*q.S), q.S + float32(deltaAngle*q.C)}
	mag := float32(math.Sqrt(float64(float32(q2.S*q2.S) + float32(q2.C*q2.C))))

	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}

	return Rot{q2.C * invMag, q2.S * invMag}
}

// Get the length squared of this vector.
func (a Vec2) LengthSquared() float32 {
	return float32(a.X*a.X) + float32(a.Y*a.Y)
}

// Get the distance squared between points.
func (a Vec2) DistanceSquared(b Vec2) float32 {
	c := Vec2{b.X - a.X, b.Y - a.Y}
	return float32(c.X*c.X) + float32(c.Y*c.Y)
}

// Make a rotation using an angle in radians.
func MakeRot(radians float32) Rot {
	cs := ComputeCosSin(radians)
	return Rot{cs.Cosine, cs.Sine}
}

// Make a rotation using a unit vector.
func MakeRotFromUnitVector(unitVector Vec2) Rot {
	core.Assert(unitVector.IsNormalized(), "IsNormalized(unitVector)")
	return Rot{unitVector.X, unitVector.Y}
}

// Is this rotation normalized?
func (q Rot) IsNormalized() bool {
	qq := float32(q.S*q.S) + float32(q.C*q.C)
	return float32(1)-float32(0.0006) < qq && qq < float32(1)+float32(0.0006)
}

// Get the inverse of a rotation.
func (q Rot) Invert() Rot {
	return Rot{q.C, -q.S}
}

// Normalized linear interpolation
// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
// https://web.archive.org/web/20170825184056/http://number-none.com/product/Understanding%20Slerp,%20Then%20Not%20Using%20It/
func (q Rot) NLerp(q2 Rot, t float32) Rot {
	omt := 1 - t
	interpolated := Rot{float32(omt*q.C) + float32(t*q2.C), float32(omt*q.S) + float32(t*q2.S)}
	mag := float32(math.Sqrt(float64(float32(interpolated.S*interpolated.S) + float32(interpolated.C*interpolated.C))))

	var invMag float32
	if mag > 0 {
		invMag = 1 / mag
	}

	return Rot{interpolated.C * invMag, interpolated.S * invMag}
}

// Compute the angular velocity necessary to rotate between two rotations over a give time
// @param q1 initial rotation
// @param q2 final rotation
// @param invH inverse time step.
func (q Rot) AngularVelocity(q2 Rot, invH float32) float32 {
	omega := invH * (float32(q2.S*q.C) - float32(q2.C*q.S))
	return omega
}

// Get the angle in radians in the range [-pi, pi].
func (q Rot) Angle() float32 {
	return Atan2(q.S, q.C)
}

// Get the x-axis.
func (q Rot) XAxis() Vec2 {
	return Vec2{q.C, q.S}
}

// Get the y-axis.
func (q Rot) YAxis() Vec2 {
	return Vec2{-q.S, q.C}
}

// Multiply two rotations: q * r.
func (q Rot) Mul(r Rot) Rot {
	var qr Rot

	qr.S = float32(q.S*r.C) + float32(q.C*r.S)
	qr.C = float32(q.C*r.C) - float32(q.S*r.S)

	return qr
}

// Transpose multiply two rotations: inv(a) * b
// This rotates a vector local in frame b into a vector local in frame a.
func (q Rot) InvMul(b Rot) Rot {
	var r Rot

	r.S = float32(q.C*b.S) - float32(q.S*b.C)
	r.C = float32(q.C*b.C) + float32(q.S*b.S)

	return r
}

// Relative angle between a and b.
func (q Rot) RelativeAngle(b Rot) float32 {
	s := float32(q.C*b.S) - float32(q.S*b.C)
	c := float32(q.C*b.C) + float32(q.S*b.S)

	return Atan2(s, c)
}

// Convert any angle into the range [-pi, pi].
func UnwindAngle(radians float32) float32 {
	core.Validate(absFloat(radians) < 1e4, "AbsFloat(radians) < 1.0e4f")
	x := clampFloat(radians, -1e6, 1e6)
	twoPi := float64(float32(2) * Pi)

	const roundToNearest float64 = 6755399441055744

	a := float64(x)
	k := (a/twoPi + roundToNearest) - roundToNearest

	return float32(a - float64(k*twoPi))
}

// Rotate a vector.
func (q Rot) Rotate(v Vec2) Vec2 {
	return Vec2{float32(q.C*v.X) - float32(q.S*v.Y), float32(q.S*v.X) + float32(q.C*v.Y)}
}

// Inverse rotate a vector.
func (q Rot) InvRotate(v Vec2) Vec2 {
	return Vec2{float32(q.C*v.X) + float32(q.S*v.Y), float32(-q.S*v.X) + float32(q.C*v.Y)}
}

// Transform a point (e.g. local space to world space).
func (t Transform) Apply(p Vec2) Vec2 {
	x := (float32(t.Q.C*p.X) - float32(t.Q.S*p.Y)) + t.P.X
	y := (float32(t.Q.S*p.X) + float32(t.Q.C*p.Y)) + t.P.Y

	return Vec2{x, y}
}

// Inverse transform a point (e.g. world space to local space).
func (t Transform) InvApply(p Vec2) Vec2 {
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
func (t Transform) Mul(b Transform) Transform {
	var c Transform

	c.Q = t.Q.Mul(b.Q)
	c.P = t.Q.Rotate(b.P).Add(t.P)

	return c
}

// Creates a transform that converts a local point in frame B to a local point in frame A.
// v2 = a.Q' * (b.Q * v1 + b.P - a.P)
//
//	= a.Q' * b.Q * v1 + a.Q' * (b.P - a.P)
func (t Transform) InvMul(b Transform) Transform {
	var c Transform

	c.Q = t.Q.InvMul(b.Q)
	c.P = t.Q.InvRotate(b.P.Sub(t.P))

	return c
}

// Lossy conversion of a world position to a float vector. no-op in single precision.
func (p Pos) Vec2() Vec2 {
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
func (p Pos) Sub(b Pos) Vec2 {
	return Vec2{float32(p.X - b.X), float32(p.Y - b.Y)}
}

// Transform a local point to a world position. Rotation in float, translation in double.
func (t WorldTransform) Apply(p Vec2) Pos {
	rx := float32(t.Q.C*p.X) - float32(t.Q.S*p.Y)
	ry := float32(t.Q.S*p.X) + float32(t.Q.C*p.Y)

	return t.P.Add(Vec2{rx, ry})
}

// Transform a world position to a local point. One double subtraction, then float.
func (t WorldTransform) InvApply(p Pos) Vec2 {
	vx := float32(p.X - t.P.X)
	vy := float32(p.Y - t.P.Y)

	return Vec2{float32(t.Q.C*vx) + float32(t.Q.S*vy), float32(-t.Q.S*vx) + float32(t.Q.C*vy)}
}

// Relative transform of frame B in frame A.
func (t WorldTransform) InvMul(b WorldTransform) Transform {
	var c Transform

	c.Q = t.Q.InvMul(b.Q)
	d := Vec2{float32(b.P.X - t.P.X), float32(b.P.Y - t.P.Y)}
	c.P = t.Q.InvRotate(d)

	return c
}

// Convert a local transform B into world space using world transform A.
func (t WorldTransform) Mul(b Transform) WorldTransform {
	var c WorldTransform

	c.Q = t.Q.Mul(b.Q)
	c.P = t.P.Add(t.Q.Rotate(b.P))

	return c
}

// Shift a world transform into the frame of a base position.
func (t WorldTransform) Relative(base Pos) Transform {
	var r Transform

	r.Q = t.Q
	r.P = Vec2{float32(t.P.X - base.X), float32(t.P.Y - base.Y)}

	return r
}

// Promote a float transform to a world transform. Lossless.
func MakeWorldTransform(t Transform) WorldTransform {
	var w WorldTransform

	w.P = t.P.Pos()
	w.Q = t.Q

	return w
}

// Multiply a 2-by-2 matrix times a 2D vector.
func (a Mat22) MulVec(v Vec2) Vec2 {
	return Vec2{float32(a.Cx.X*v.X) + float32(a.Cy.X*v.Y), float32(a.Cx.Y*v.X) + float32(a.Cy.Y*v.Y)}
}

// Get the inverse of a 2-by-2 matrix.
func (a Mat22) Inverse() Mat22 {
	aa, b, c, d := a.Cx.X, a.Cy.X, a.Cx.Y, a.Cy.Y

	det := float32(aa*d) - float32(b*c)
	if det != 0 {
		det = 1 / det
	}

	return Mat22{Vec2{det * d, -det * c}, Vec2{-det * b, det * aa}}
}

// Solve A * x = b, where b is a column vector. This is more efficient
// than computing the inverse in one-shot cases.
func (a Mat22) Solve(b Vec2) Vec2 {
	a11, a12, a21, a22 := a.Cx.X, a.Cy.X, a.Cx.Y, a.Cy.Y

	det := float32(a11*a22) - float32(a12*a21)
	if det != 0 {
		det = 1 / det
	}

	return Vec2{det * (float32(a22*b.X) - float32(a12*b.Y)), det * (float32(a11*b.Y) - float32(a21*b.X))}
}

// Does a fully contain b.
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

// Union of two AABBs.
func (a AABB) Union(b AABB) AABB {
	var c AABB

	c.LowerBound.X = minFloat(a.LowerBound.X, b.LowerBound.X)
	c.LowerBound.Y = minFloat(a.LowerBound.Y, b.LowerBound.Y)
	c.UpperBound.X = maxFloat(a.UpperBound.X, b.UpperBound.X)
	c.UpperBound.Y = maxFloat(a.UpperBound.Y, b.UpperBound.Y)

	return c
}

// Do a and b overlap.
func (a AABB) Overlaps(b AABB) bool {
	return !(b.LowerBound.X > a.UpperBound.X || b.LowerBound.Y > a.UpperBound.Y || a.LowerBound.X > b.UpperBound.X || a.LowerBound.Y > b.UpperBound.Y)
}

// Compute the bounding box of an array of circles.
func MakeAABB(points []Vec2, radius float32) AABB {
	count := len(points)
	core.Assert(count > 0, "count > 0")

	a := AABB{points[0], points[0]}
	for i := 1; i < count; i++ {
		a.LowerBound = a.LowerBound.Min(points[i])
		a.UpperBound = a.UpperBound.Max(points[i])
	}

	r := Vec2{radius, radius}
	a.LowerBound = a.LowerBound.Sub(r)
	a.UpperBound = a.UpperBound.Add(r)

	return a
}

// Signed separation of a point from a plane.
func (plane Plane) Separation(point Vec2) float32 {
	return plane.Normal.Dot(point) - plane.Offset
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
func (a Vec2) IsValid() bool {
	if math.IsNaN(float64(a.X)) || math.IsNaN(float64(a.Y)) {
		return false
	}

	if math.IsInf(float64(a.X), 0) || math.IsInf(float64(a.Y), 0) {
		return false
	}

	return true
}

// Is this a valid rotation? Not NaN or infinity. Is normalized.
func (q Rot) IsValid() bool {
	if math.IsNaN(float64(q.S)) || math.IsNaN(float64(q.C)) {
		return false
	}

	if math.IsInf(float64(q.S), 0) || math.IsInf(float64(q.C), 0) {
		return false
	}

	return q.IsNormalized()
}

// Is this a valid transform? Not NaN or infinity. Rotation is normalized.
func (t Transform) IsValid() bool {
	if !t.P.IsValid() {
		return false
	}

	return t.Q.IsValid()
}

// Is this a valid plane? Normal is a unit vector. Not Nan or infinity.
func (plane Plane) IsValid() bool {
	return plane.Normal.IsValid() && plane.Normal.IsNormalized() && IsValidFloat(plane.Offset)
}

// Is this a valid world position? Not NaN or infinity.
func (p Pos) IsValid() bool {
	if math.IsNaN(float64(p.X)) || math.IsNaN(float64(p.Y)) {
		return false
	}

	if math.IsInf(float64(p.X), 0) || math.IsInf(float64(p.Y), 0) {
		return false
	}

	return true
}

// Is this a valid world transform? Not NaN or infinity. Rotation is normalized.
func (t WorldTransform) IsValid() bool {
	if !t.P.IsValid() {
		return false
	}

	return t.Q.IsValid()
}

// Compute an approximate arctangent in the range [-pi, pi]
// This is hand coded for cross-platform determinism. The atan2f
// function in the standard library is not cross-platform deterministic.
// Accurate to around 0.0023 degrees.
func Atan2(y float32, x float32) float32 {
	if x == 0 && y == 0 {
		return 0
	}

	ax := absFloat(x)
	ay := absFloat(y)
	mx := maxFloat(ay, ax)
	mn := minFloat(ay, ax)
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

	switch {
	case x < -0.5*Pi:
		y := x + Pi
		y2 := float32(y * y)
		c = -(pi2 - float32(4*y2)) / (pi2 + y2)
	case x > 0.5*Pi:
		y := x - Pi
		y2 := float32(y * y)
		c = -(pi2 - float32(4*y2)) / (pi2 + y2)
	default:
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

// Compute the rotation between two unit vectors.
func ComputeRotationBetweenUnitVectors(v1 Vec2, v2 Vec2) Rot {
	core.Assert(absFloat(1-v1.Length()) < float32(100)*float32(0x1p-23), "AbsFloat(1.0f - Length(v1)) < 100.0f * FLT_EPSILON")
	core.Assert(absFloat(1-v2.Length()) < float32(100)*float32(0x1p-23), "AbsFloat(1.0f - Length(v2)) < 100.0f * FLT_EPSILON")

	var rot Rot

	rot.C = v1.Dot(v2)
	rot.S = v1.Cross(v2)

	return rot.Normalize()
}

// Convert a vector to a world position. no-op in single precision.
func (a Vec2) Pos() Pos {
	return Pos{posScalar(a.X), posScalar(a.Y)}
}

// p + d.
func (p Pos) Add(d Vec2) Pos {
	return Pos{p.X + posScalar(d.X), p.Y + posScalar(d.Y)}
}

// World position interpolation for sweeps and sampling.
func (p Pos) Lerp(b Pos, t float32) Pos {
	return Pos{posScalar(posScalar(1-t)*p.X) + posScalar(posScalar(t)*b.X), posScalar(posScalar(1-t)*p.Y) + posScalar(posScalar(t)*b.Y)}
}

// @return the absolute value of an integer.
func absInt(a int) int {
	if a < 0 {
		return -a
	}

	return a
}

// https://en.wikipedia.org/wiki/Floor_and_ceiling_functions
func ceilingInt(numerator int, denominator int) int {
	core.Validate(denominator > 0 && numerator >= 0, "denominator > 0 && numerator >= 0")
	return (numerator + denominator - 1) / denominator
}

// @return the minimum of two floats.
func minFloat(a float32, b float32) float32 {
	if a < b {
		return a
	}

	return b
}

// @return the maximum of two floats.
func maxFloat(a float32, b float32) float32 {
	if a > b {
		return a
	}

	return b
}

// @return the absolute value of a float.
func absFloat(a float32) float32 {
	magnitude := math.Float32bits(a) & 0x7fffffff
	if magnitude == 0 || magnitude > 0x7f800000 {
		return a
	}

	return math.Float32frombits(magnitude)
}

// @return a float clamped between a lower and upper bound.
func clampFloat(a float32, lower float32, upper float32) float32 {
	if a < lower {
		return lower
	}

	if a > upper {
		return upper
	}

	return a
}
