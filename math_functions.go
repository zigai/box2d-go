// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// @return the minimum of two integers
func MinInt(a int, b int) int {
	todo()
	return 0
}

// @return the maximum of two integers
func MaxInt(a int, b int) int {
	todo()
	return 0
}

// @return the absolute value of an integer
func AbsInt(a int) int {
	todo()
	return 0
}

// @return an integer clamped between a lower and upper bound
func ClampInt(a int, lower int, upper int) int {
	todo()
	return 0
}

// https://en.wikipedia.org/wiki/Floor_and_ceiling_functions
func CeilingInt(numerator int, denominator int) int {
	todo()
	return 0
}

// @return the minimum of two floats
func MinFloat(a float32, b float32) float32 {
	todo()
	return 0
}

// @return the maximum of two floats
func MaxFloat(a float32, b float32) float32 {
	todo()
	return 0
}

// @return the absolute value of a float
func AbsFloat(a float32) float32 {
	todo()
	return 0
}

// @return a float clamped between a lower and upper bound
func ClampFloat(a float32, lower float32, upper float32) float32 {
	todo()
	return 0
}

// Vector dot product
func Dot(a Vec2, b Vec2) float32 {
	todo()
	return 0
}

// Vector cross product. In 2D this yields a scalar.
func Cross(a Vec2, b Vec2) float32 {
	todo()
	return 0
}

// Perform the cross product on a vector and a scalar. In 2D this produces a vector.
func CrossVS(v Vec2, s float32) Vec2 {
	todo()
	return Vec2{}
}

// Perform the cross product on a scalar and a vector. In 2D this produces a vector.
func CrossSV(s float32, v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Get a left pointing perpendicular vector. Equivalent to CrossSV(1.0f, v)
func LeftPerp(v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Get a right pointing perpendicular vector. Equivalent to CrossVS(v, 1.0f)
func RightPerp(v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Vector addition
func Add(a Vec2, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Vector subtraction
func Sub(a Vec2, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Vector negation
func Neg(a Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Vector linear interpolation
// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
func Lerp(a Vec2, b Vec2, t float32) Vec2 {
	todo()
	return Vec2{}
}

// Component-wise multiplication
func Mul(a Vec2, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Multiply a scalar and vector
func MulSV(s float32, v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// a + s * b
func MulAdd(a Vec2, s float32, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// a - s * b
func MulSub(a Vec2, s float32, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Component-wise absolute vector
func Abs(a Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Component-wise minimum vector
func Min(a Vec2, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Component-wise maximum vector
func Max(a Vec2, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Component-wise clamp vector v into the range [a, b]
func Clamp(v Vec2, a Vec2, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Get the length of this vector (the norm)
func Length(v Vec2) float32 {
	todo()
	return 0
}

// Get the distance between two points
func Distance(a Vec2, b Vec2) float32 {
	todo()
	return 0
}

// Convert a vector into a unit vector if possible, otherwise returns the zero vector.
// todo MSVC is not inlining this function in several places per warning 4710
func Normalize(a Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Determines if the provided vector is normalized (norm(a) == 1).
func IsNormalized(a Vec2) bool {
	todo()
	return false
}

// Convert a vector into a unit vector if possible, otherwise returns the zero vector. Also
// outputs the length.
func GetLengthAndNormalize(length *float32, a Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Normalize rotation
func NormalizeRot(q Rot) Rot {
	todo()
	return Rot{}
}

// Integrate rotation from angular velocity
// @param q1 initial rotation
// @param deltaAngle the angular displacement in radians
func IntegrateRotation(q1 Rot, deltaAngle float32) Rot {
	todo()
	return Rot{}
}

// Get the length squared of this vector
func LengthSquared(v Vec2) float32 {
	todo()
	return 0
}

// Get the distance squared between points
func DistanceSquared(a Vec2, b Vec2) float32 {
	todo()
	return 0
}

// Make a rotation using an angle in radians
func MakeRot(radians float32) Rot {
	todo()
	return Rot{}
}

// Make a rotation using a unit vector
func MakeRotFromUnitVector(unitVector Vec2) Rot {
	todo()
	return Rot{}
}

// Is this rotation normalized?
func IsNormalizedRot(q Rot) bool {
	todo()
	return false
}

// Get the inverse of a rotation
func InvertRot(a Rot) Rot {
	todo()
	return Rot{}
}

// Normalized linear interpolation
// https://fgiesen.wordpress.com/2012/08/15/linear-interpolation-past-present-and-future/
// https://web.archive.org/web/20170825184056/http://number-none.com/product/Understanding%20Slerp,%20Then%20Not%20Using%20It/
func NLerp(q1 Rot, q2 Rot, t float32) Rot {
	todo()
	return Rot{}
}

// Compute the angular velocity necessary to rotate between two rotations over a give time
// @param q1 initial rotation
// @param q2 final rotation
// @param invH inverse time step
func ComputeAngularVelocity(q1 Rot, q2 Rot, invH float32) float32 {
	todo()
	return 0
}

// Get the angle in radians in the range [-pi, pi]
func (q Rot) Angle() float32 {
	todo()
	return 0
}

// Get the x-axis
func (q Rot) XAxis() Vec2 {
	todo()
	return Vec2{}
}

// Get the y-axis
func (q Rot) YAxis() Vec2 {
	todo()
	return Vec2{}
}

// Multiply two rotations: q * r
func MulRot(q Rot, r Rot) Rot {
	todo()
	return Rot{}
}

// Transpose multiply two rotations: inv(a) * b
// This rotates a vector local in frame b into a vector local in frame a
func InvMulRot(a Rot, b Rot) Rot {
	todo()
	return Rot{}
}

// Relative angle between a and b.
func RelativeAngle(a Rot, b Rot) float32 {
	todo()
	return 0
}

// Convert any angle into the range [-pi, pi].
func UnwindAngle(radians float32) float32 {
	todo()
	return 0
}

// Rotate a vector.
func RotateVector(q Rot, v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Inverse rotate a vector.
func InvRotateVector(q Rot, v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Transform a point (e.g. local space to world space).
func TransformPoint(t Transform, p Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Inverse transform a point (e.g. world space to local space).
func InvTransformPoint(t Transform, p Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Multiply two transforms. If the result is applied to a point p local to frame B,
// the transform would first convert p to a point local to frame A, then into a point
// in the world frame.
// v2 = a.Q.Rot(b.Q.Rot(v1) + b.P) + a.P
//
//	= (a.Q * b.Q).Rot(v1) + a.Q.Rot(b.P) + a.P
func MulTransforms(a Transform, b Transform) Transform {
	todo()
	return Transform{}
}

// Creates a transform that converts a local point in frame B to a local point in frame A.
// v2 = a.Q' * (b.Q * v1 + b.P - a.P)
//
//	= a.Q' * b.Q * v1 + a.Q' * (b.P - a.P)
func InvMulTransforms(a Transform, b Transform) Transform {
	todo()
	return Transform{}
}

// Convert a vector to a world position. no-op in single precision.
func ToPos(v Vec2) Pos {
	todo()
	return Pos{}
}

// Lossy conversion of a world position to a float vector. no-op in single precision.
func ToVec2(p Pos) Vec2 {
	todo()
	return Vec2{}
}

// Narrow a world coordinate to float, rounding toward negative infinity. Use with
// RoundUpFloat to build a conservative float box that always contains double bounds,
// where plain rounding far from the origin could clip. nextafterf is an exact IEEE
// operation, so this is cross-platform deterministic. With large world mode off this is
// a plain conversion.
func RoundDownFloat(x float64) float32 {
	todo()
	return 0
}

// Narrow a world coordinate to float, rounding toward positive infinity.
func RoundUpFloat(x float64) float32 {
	todo()
	return 0
}

// a - b, demoted to float. The primary precision boundary operation.
func SubPos(a Pos, b Pos) Vec2 {
	todo()
	return Vec2{}
}

// p + d
func OffsetPos(p Pos, d Vec2) Pos {
	todo()
	return Pos{}
}

// World position interpolation for sweeps and sampling.
func LerpPosition(a Pos, b Pos, t float32) Pos {
	todo()
	return Pos{}
}

// Transform a local point to a world position. Rotation in float, translation in double.
func TransformWorldPoint(t WorldTransform, p Vec2) Pos {
	todo()
	return Pos{}
}

// Transform a world position to a local point. One double subtraction, then float.
func InvTransformWorldPoint(t WorldTransform, p Pos) Vec2 {
	todo()
	return Vec2{}
}

// Relative transform of frame B in frame A.
func InvMulWorldTransforms(a WorldTransform, b WorldTransform) Transform {
	todo()
	return Transform{}
}

// Convert a local transform B into world space using world transform A.
func MulWorldTransforms(a WorldTransform, b Transform) WorldTransform {
	todo()
	return WorldTransform{}
}

// Shift a world transform into the frame of a base position.
func ToRelativeTransform(t WorldTransform, base Pos) Transform {
	todo()
	return Transform{}
}

// Promote a float transform to a world transform. Lossless.
func MakeWorldTransform(t Transform) WorldTransform {
	todo()
	return WorldTransform{}
}

// Multiply a 2-by-2 matrix times a 2D vector
func MulMV(a Mat22, v Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Get the inverse of a 2-by-2 matrix
func GetInverse22(a Mat22) Mat22 {
	todo()
	return Mat22{}
}

// Solve A * x = b, where b is a column vector. This is more efficient
// than computing the inverse in one-shot cases.
func Solve22(a Mat22, b Vec2) Vec2 {
	todo()
	return Vec2{}
}

// Does a fully contain b
func (a AABB) Contains(b AABB) bool {
	todo()
	return false
}

// Get the center of the AABB.
func (a AABB) Center() Vec2 {
	todo()
	return Vec2{}
}

// Get the extents of the AABB (half-widths).
func (a AABB) Extents() Vec2 {
	todo()
	return Vec2{}
}

// Union of two AABBs
func (a AABB) Union(b AABB) AABB {
	todo()
	return AABB{}
}

// Do a and b overlap
func (a AABB) Overlaps(b AABB) bool {
	todo()
	return false
}

// Compute the bounding box of an array of circles
func MakeAABB(points []Vec2, count int, radius float32) AABB {
	todo()
	return AABB{}
}

// Signed separation of a point from a plane
func PlaneSeparation(plane Plane, point Vec2) float32 {
	todo()
	return 0
}

// One-dimensional mass-spring-damper simulation. Returns the new velocity given the position and time step.
// You can then compute the new position using:
// position += timeStep * newVelocity
// This drives towards a zero position. By using implicit integration we get a stable solution
// that doesn't require transcendental functions.
func SpringDamper(hertz float32, dampingRatio float32, position float32, velocity float32, timeStep float32) float32 {
	todo()
	return 0
}

// Is this a valid number? Not NaN or infinity.
func IsValidFloat(a float32) bool {
	todo()
	return false
}

// Is this a valid vector? Not NaN or infinity.
func IsValidVec2(v Vec2) bool {
	todo()
	return false
}

// Is this a valid rotation? Not NaN or infinity. Is normalized.
func IsValidRotation(q Rot) bool {
	todo()
	return false
}

// Is this a valid transform? Not NaN or infinity. Rotation is normalized.
func IsValidTransform(t Transform) bool {
	todo()
	return false
}

// Is this a valid plane? Normal is a unit vector. Not Nan or infinity.
func IsValidPlane(a Plane) bool {
	todo()
	return false
}

// Is this a valid world position? Not NaN or infinity.
func IsValidPosition(p Pos) bool {
	todo()
	return false
}

// Is this a valid world transform? Not NaN or infinity. Rotation is normalized.
func IsValidWorldTransform(t WorldTransform) bool {
	todo()
	return false
}

// Compute an approximate arctangent in the range [-pi, pi]
// This is hand coded for cross-platform determinism. The atan2f
// function in the standard library is not cross-platform deterministic.
// Accurate to around 0.0023 degrees
func Atan2(y float32, x float32) float32 {
	todo()
	return 0
}

// Compute the cosine and sine of an angle in radians. Implemented
// for cross-platform determinism.
func ComputeCosSin(radians float32) CosSin {
	todo()
	return CosSin{}
}

// Compute the rotation between two unit vectors
func ComputeRotationBetweenUnitVectors(v1 Vec2, v2 Vec2) Rot {
	todo()
	return Rot{}
}
