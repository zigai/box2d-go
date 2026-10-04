// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func mathPublicKey(i int, field string) string {
	return fmt.Sprintf("%d/%s", i, field)
}

func mathPublicVec(r *oracle.Reader, key string, value box2d.Vec2) {
	r.Vec2(key, value.X, value.Y)
}

func mathPublicRot(r *oracle.Reader, key string, value box2d.Rot) {
	r.Rot(key, value.C, value.S)
}

func mathPublicPos(r *oracle.Reader, key string, value box2d.Pos) {
	r.Pos(key, value.X, value.Y)
}

func mathPublicTransform(r *oracle.Reader, key string, value box2d.Transform) {
	mathPublicVec(r, key+"/p", value.P)
	mathPublicRot(r, key+"/q", value.Q)
}

func mathPublicWorldTransform(r *oracle.Reader, key string, value box2d.WorldTransform) {
	mathPublicPos(r, key+"/p", value.P)
	mathPublicRot(r, key+"/q", value.Q)
}

// This is the complete 38-boundary/128-sweep input inventory from
// oracle/sample/make_rot.c. All outputs exercise the production API; no
// test-local copies of its trigonometry or angle reduction are involved.
func TestMathPublicSpecialAngles(t *testing.T) {
	r := oracle.Open(t, "make_rot")
	pi := box2d.Pi
	fb := math.Float32frombits
	negativeZero := fb(0x80000000)
	angles := []float32{
		0, negativeZero,
		pi, -pi,
		0.5 * pi, -0.5 * pi,
		0.25 * pi, -0.25 * pi,
		0.75 * pi, -0.75 * pi,
		fb(0x40490fdb), fb(0x40490fda), fb(0x40490fdc),
		2 * pi, -2 * pi, 3 * pi,
		1e-8, -1e-8, 1e-30,
		fb(0x00000001), fb(0x007fffff), fb(0x00800000),
		0.1, -0.1, 0.5, 1, -1, 2, 3, -3, 10, -10,
		100, -100, 1000, -1000, 9999.5, -9999.5,
	}
	for i := range 128 {
		angles = append(angles, float32(float32(i)*0.1)-6.4)
	}

	r.I32("count", int32(len(angles)))
	hash := oracle.HashInit
	for i, angle := range angles {
		cs := box2d.ComputeCosSin(angle)
		rotation := box2d.MakeRot(angle)
		r.F32(mathPublicKey(i, "angle"), angle)
		r.F32(mathPublicKey(i, "unwound"), box2d.UnwindAngle(angle))
		r.F32(mathPublicKey(i, "cosine"), cs.Cosine)
		r.F32(mathPublicKey(i, "sine"), cs.Sine)
		mathPublicRot(r, mathPublicKey(i, "rot"), rotation)
		hash = oracle.HashRot(hash, rotation.C, rotation.S)
	}
	r.U64("hash", hash)
}

func TestMathPublicScalars(t *testing.T) {
	r := oracle.Open(t, "math/PublicScalars")
	integers := []int{0, 1, -1, 17, -17, math.MaxInt32, math.MinInt32 + 1}
	r.I32("integers/count", int32(len(integers)))
	for i, a := range integers {
		b := integers[len(integers)-1-i]
		r.I32(fmt.Sprintf("integers/%d/min", i), int32(box2d.MinInt(a, b)))
		r.I32(fmt.Sprintf("integers/%d/max", i), int32(box2d.MaxInt(a, b)))
		r.I32(fmt.Sprintf("integers/%d/abs", i), int32(box2d.AbsInt(a)))
		r.I32(fmt.Sprintf("integers/%d/clamp", i), int32(box2d.ClampInt(a, -1, 1)))
	}
	divisions := [][2]int{{0, 3}, {1, 3}, {3, 3}, {4, 3}, {17, 5}, {math.MaxInt32 - 2, 2}}
	r.I32("ceiling/count", int32(len(divisions)))
	for i, pair := range divisions {
		r.I32(fmt.Sprintf("ceiling/%d", i), int32(box2d.CeilingInt(pair[0], pair[1])))
	}
	negativeZero := math.Float32frombits(0x80000000)
	floats := []float32{0, negativeZero, -3.25, 3.25, -2, 2, 0x1p-126, 0x1p-149, math.MaxFloat32}
	r.I32("floats/count", int32(len(floats)))
	for i, a := range floats {
		b := floats[len(floats)-1-i]
		r.F32(fmt.Sprintf("floats/%d/min", i), box2d.MinFloat(a, b))
		r.F32(fmt.Sprintf("floats/%d/max", i), box2d.MaxFloat(a, b))
		r.F32(fmt.Sprintf("floats/%d/abs", i), box2d.AbsFloat(a))
		r.F32(fmt.Sprintf("floats/%d/clamp", i), box2d.ClampFloat(a, -2, 2))
	}
	r.F32("zero/min-negative-first", box2d.MinFloat(negativeZero, 0))
	r.F32("zero/min-positive-first", box2d.MinFloat(0, negativeZero))
	r.F32("zero/max-negative-first", box2d.MaxFloat(negativeZero, 0))
	r.F32("zero/max-positive-first", box2d.MaxFloat(0, negativeZero))
}

func TestMathPublicVectors(t *testing.T) {
	r := oracle.Open(t, "math/PublicVectors")
	negativeZero := math.Float32frombits(0x80000000)
	cases := []struct {
		a, b box2d.Vec2
		s, t float32
	}{
		{box2d.Vec2{X: 2.5, Y: -3.75}, box2d.Vec2{X: -4, Y: 0.125}, -1.25, 0.3},
		{box2d.Vec2{X: 0, Y: negativeZero}, box2d.Vec2{X: negativeZero, Y: 0}, negativeZero, 0},
		{box2d.Vec2{X: 1, Y: 0}, box2d.Vec2{X: 0, Y: 1}, 2, 1},
		{box2d.Vec2{X: -1, Y: -1}, box2d.Vec2{X: -1, Y: -1}, 0, 0.5},
		{box2d.Vec2{X: 0x1p-60, Y: -0x1p-60}, box2d.Vec2{X: 0x1p-59, Y: 0x1p-60}, 0.5, -0.25},
		{box2d.Vec2{X: 1024.125, Y: -2048.5}, box2d.Vec2{X: -16.75, Y: 33.25}, 0.125, 1.25},
	}
	r.I32("count", int32(len(cases)))
	for i, c := range cases {
		r.F32(mathPublicKey(i, "dot"), box2d.Dot(c.a, c.b))
		r.F32(mathPublicKey(i, "cross"), box2d.Cross(c.a, c.b))
		mathPublicVec(r, mathPublicKey(i, "cross-vs"), box2d.CrossVS(c.a, c.s))
		mathPublicVec(r, mathPublicKey(i, "cross-sv"), box2d.CrossSV(c.s, c.a))
		mathPublicVec(r, mathPublicKey(i, "left"), box2d.LeftPerp(c.a))
		mathPublicVec(r, mathPublicKey(i, "right"), box2d.RightPerp(c.a))
		mathPublicVec(r, mathPublicKey(i, "add"), box2d.Add(c.a, c.b))
		mathPublicVec(r, mathPublicKey(i, "sub"), box2d.Sub(c.a, c.b))
		mathPublicVec(r, mathPublicKey(i, "neg"), box2d.Neg(c.a))
		mathPublicVec(r, mathPublicKey(i, "lerp"), box2d.Lerp(c.a, c.b, c.t))
		mathPublicVec(r, mathPublicKey(i, "mul"), box2d.Mul(c.a, c.b))
		mathPublicVec(r, mathPublicKey(i, "mul-sv"), box2d.MulSV(c.s, c.a))
		mathPublicVec(r, mathPublicKey(i, "mul-add"), box2d.MulAdd(c.a, c.s, c.b))
		mathPublicVec(r, mathPublicKey(i, "mul-sub"), box2d.MulSub(c.a, c.s, c.b))
		mathPublicVec(r, mathPublicKey(i, "abs"), box2d.Abs(c.a))
		mathPublicVec(r, mathPublicKey(i, "min"), box2d.Min(c.a, c.b))
		mathPublicVec(r, mathPublicKey(i, "max"), box2d.Max(c.a, c.b))
		mathPublicVec(r, mathPublicKey(i, "clamp"), box2d.Clamp(c.a, box2d.Vec2{X: -1, Y: -2}, box2d.Vec2{X: 1, Y: 2}))
		r.F32(mathPublicKey(i, "length"), box2d.Length(c.a))
		r.F32(mathPublicKey(i, "length-squared"), box2d.LengthSquared(c.a))
		r.F32(mathPublicKey(i, "distance"), box2d.Distance(c.a, c.b))
		r.F32(mathPublicKey(i, "distance-squared"), box2d.DistanceSquared(c.a, c.b))
	}

	// Adjacent float inputs straddle sqrt(1000 * FLT_MIN). Compare direct
	// pinned C outputs rather than reimplementing the branch in the test.
	fb := math.Float32frombits
	normals := []box2d.Vec2{
		{X: 0, Y: 0}, {X: negativeZero, Y: negativeZero},
		{X: 0x1p-149, Y: 0}, {X: 0x1p-126, Y: -0x1p-126},
		{X: fb(0x227cfb71), Y: 0}, {X: fb(0x227cfb72), Y: 0},
		{X: fb(0x227cfb73), Y: 0}, {X: 0, Y: -fb(0x227cfb73)},
		{X: 3, Y: -4}, {X: 1, Y: 0},
		{X: fb(0x3f800031), Y: 0}, {X: fb(0x3f800032), Y: 0}, {X: fb(0x3f800033), Y: 0},
	}
	r.I32("normalize/count", int32(len(normals)))
	for i, input := range normals {
		mathPublicVec(r, fmt.Sprintf("normalize/%d/value", i), box2d.Normalize(input))
		length := float32(-123)
		value := box2d.GetLengthAndNormalize(&length, input)
		mathPublicVec(r, fmt.Sprintf("normalize/%d/with-length", i), value)
		r.F32(fmt.Sprintf("normalize/%d/length", i), length)
		r.Bool(fmt.Sprintf("normalize/%d/unit", i), box2d.IsNormalized(input))
	}
}

func TestMathPublicRotations(t *testing.T) {
	r := oracle.Open(t, "math/PublicRotations")
	cases := []struct {
		a, b, delta, t, invH float32
	}{
		{0, 0, 0, 0, 60},
		{0.25, -0.7, 0.03, 0.3, 60},
		{-2.5, 2.75, -0.125, 1, 120},
		{0, box2d.Pi, 0.5, 0.5, 0},
		{-box2d.Pi, 0.5 * box2d.Pi, -0.5, 0, 1},
	}
	r.I32("count", int32(len(cases)))
	for i, c := range cases {
		a, b := box2d.MakeRot(c.a), box2d.MakeRot(c.b)
		mathPublicRot(r, mathPublicKey(i, "a"), a)
		mathPublicRot(r, mathPublicKey(i, "b"), b)
		mathPublicRot(r, mathPublicKey(i, "integrate"), box2d.IntegrateRotation(a, c.delta))
		mathPublicRot(r, mathPublicKey(i, "from-vector"), box2d.MakeRotFromUnitVector(a.XAxis()))
		mathPublicRot(r, mathPublicKey(i, "between-vectors"), box2d.ComputeRotationBetweenUnitVectors(a.XAxis(), b.XAxis()))
		mathPublicRot(r, mathPublicKey(i, "inverse"), box2d.InvertRot(a))
		mathPublicRot(r, mathPublicKey(i, "nlerp"), box2d.NLerp(a, b, c.t))
		r.F32(mathPublicKey(i, "angular-velocity"), box2d.ComputeAngularVelocity(a, b, c.invH))
		r.F32(mathPublicKey(i, "angle"), a.Angle())
		mathPublicVec(r, mathPublicKey(i, "x-axis"), a.XAxis())
		mathPublicVec(r, mathPublicKey(i, "y-axis"), a.YAxis())
		mathPublicRot(r, mathPublicKey(i, "mul"), box2d.MulRot(a, b))
		mathPublicRot(r, mathPublicKey(i, "inv-mul"), box2d.InvMulRot(a, b))
		r.F32(mathPublicKey(i, "relative-angle"), box2d.RelativeAngle(a, b))
		mathPublicVec(r, mathPublicKey(i, "rotate"), box2d.RotateVector(a, box2d.Vec2{X: 2.25, Y: -3.5}))
		mathPublicVec(r, mathPublicKey(i, "inv-rotate"), box2d.InvRotateVector(a, box2d.Vec2{X: 2.25, Y: -3.5}))
	}
	negativeZero := math.Float32frombits(0x80000000)
	raw := []box2d.Rot{
		{C: 0, S: 0}, {C: negativeZero, S: negativeZero},
		{C: 3, S: -4}, {C: 1, S: 0}, {C: 0x1p-60, S: -0x1p-60},
	}
	r.I32("normalize/count", int32(len(raw)))
	for i, input := range raw {
		mathPublicRot(r, fmt.Sprintf("normalize/%d", i), box2d.NormalizeRot(input))
	}
	mathPublicRot(r, "nlerp-opposite", box2d.NLerp(box2d.Rot{C: 1}, box2d.Rot{C: -1}, 0.5))
	directions := []box2d.Vec2{
		{X: 0, Y: 0}, {X: negativeZero, Y: negativeZero}, {X: 1, Y: 0}, {X: -1, Y: 0},
		{X: 0, Y: 1}, {X: 0, Y: -1}, {X: 1, Y: 1}, {X: -1, Y: 1},
		{X: -1, Y: -1}, {X: 1, Y: -1}, {X: 0x1p-149, Y: 0x1p-148},
	}
	r.I32("atan2/count", int32(len(directions)))
	for i, input := range directions {
		r.F32(fmt.Sprintf("atan2/%d", i), box2d.Atan2(input.Y, input.X))
	}
}

func TestMathPublicMatrixSpring(t *testing.T) {
	r := oracle.Open(t, "math/PublicMatrixSpring")
	matrices := []box2d.Mat22{
		{Cx: box2d.Vec2{X: 1}, Cy: box2d.Vec2{Y: 1}},
		{Cx: box2d.Vec2{X: 3, Y: -2}, Cy: box2d.Vec2{X: 1.25, Y: 4}},
		{},
		{Cx: box2d.Vec2{X: 1, Y: 2}, Cy: box2d.Vec2{X: 2, Y: 4}},
		{Cx: box2d.Vec2{X: -2}, Cy: box2d.Vec2{Y: 0.5}},
		{Cx: box2d.Vec2{X: 0x1p-10, Y: 1}, Cy: box2d.Vec2{Y: 0x1p-10}},
	}
	r.I32("matrix/count", int32(len(matrices)))
	for i, matrix := range matrices {
		rhs := box2d.Vec2{X: 2.5, Y: -3.75}
		mathPublicVec(r, fmt.Sprintf("matrix/%d/mul", i), box2d.MulMV(matrix, rhs))
		inverse := box2d.GetInverse22(matrix)
		mathPublicVec(r, fmt.Sprintf("matrix/%d/inverse-cx", i), inverse.Cx)
		mathPublicVec(r, fmt.Sprintf("matrix/%d/inverse-cy", i), inverse.Cy)
		mathPublicVec(r, fmt.Sprintf("matrix/%d/solve", i), box2d.Solve22(matrix, rhs))
	}
	springs := [][5]float32{
		{0, 0, 2, -3, 0.016}, {2, 0, 2, -3, 0.016},
		{2, 0.7, -1.25, 4, 0.016}, {10, 1, 3, 0, 1},
		{2, 2, 0, -3, 0}, {4, 1, math.Float32frombits(0x80000000), 0, 0.125},
	}
	r.I32("spring/count", int32(len(springs)))
	for i, input := range springs {
		r.F32(fmt.Sprintf("spring/%d", i), box2d.SpringDamper(input[0], input[1], input[2], input[3], input[4]))
	}
}

func TestMathPublicTransforms(t *testing.T) {
	r := oracle.Open(t, "math/PublicTransforms")
	negativeZero := math.Float32frombits(0x80000000)
	cases := []struct {
		a, b, point    box2d.Vec2
		angleA, angleB float32
		worldX, worldY float64
	}{
		{box2d.Vec2{}, box2d.Vec2{}, box2d.Vec2{Y: negativeZero}, 0, 0, 0, math.Copysign(0, -1)},
		{box2d.Vec2{X: 2.5, Y: -3.75}, box2d.Vec2{X: -0.25, Y: 4.125}, box2d.Vec2{X: 1.25, Y: -2}, 0.25, -0.7, 16.125, -32.25},
		{box2d.Vec2{X: -100, Y: 20}, box2d.Vec2{X: 13, Y: -7}, box2d.Vec2{X: -3.5, Y: 8.25}, -2.5, 2.75, 10000000.125, -10000000.25},
		{box2d.Vec2{X: 1024.125, Y: -2048.5}, box2d.Vec2{X: -1.25, Y: 3.5}, box2d.Vec2{X: 0.125, Y: -0.25}, 1.5, -0.5, 1000000000.125, -1000000000.25},
	}
	r.I32("count", int32(len(cases)))
	for i, c := range cases {
		a := box2d.Transform{P: c.a, Q: box2d.MakeRot(c.angleA)}
		b := box2d.Transform{P: c.b, Q: box2d.MakeRot(c.angleB)}
		mathPublicVec(r, mathPublicKey(i, "point"), box2d.TransformPoint(a, c.point))
		mathPublicVec(r, mathPublicKey(i, "inv-point"), box2d.InvTransformPoint(a, c.point))
		mathPublicTransform(r, mathPublicKey(i, "mul"), box2d.MulTransforms(a, b))
		mathPublicTransform(r, mathPublicKey(i, "inv-mul"), box2d.InvMulTransforms(a, b))
		world := box2d.Pos{X: oracle.Real(c.worldX), Y: oracle.Real(c.worldY)}
		wa := box2d.WorldTransform{P: world, Q: a.Q}
		wb := box2d.WorldTransform{P: box2d.OffsetPos(world, b.P), Q: b.Q}
		mathPublicPos(r, mathPublicKey(i, "to-pos"), box2d.ToPos(c.point))
		mathPublicVec(r, mathPublicKey(i, "to-vec"), box2d.ToVec2(world))
		mathPublicPos(r, mathPublicKey(i, "offset"), box2d.OffsetPos(world, c.point))
		mathPublicVec(r, mathPublicKey(i, "sub-pos"), box2d.SubPos(wb.P, wa.P))
		mathPublicPos(r, mathPublicKey(i, "world-point"), box2d.TransformWorldPoint(wa, c.point))
		mathPublicVec(r, mathPublicKey(i, "inv-world-point"), box2d.InvTransformWorldPoint(wa, wb.P))
		mathPublicTransform(r, mathPublicKey(i, "inv-world-mul"), box2d.InvMulWorldTransforms(wa, wb))
		mathPublicWorldTransform(r, mathPublicKey(i, "world-mul"), box2d.MulWorldTransforms(wa, b))
		mathPublicTransform(r, mathPublicKey(i, "relative"), box2d.ToRelativeTransform(wb, world))
		mathPublicWorldTransform(r, mathPublicKey(i, "promote"), box2d.MakeWorldTransform(a))
		for j, sample := range []float32{0, 0.3, 0.5, 1} {
			mathPublicPos(r, fmt.Sprintf("%d/interpolate/%d", i, j), box2d.LerpPosition(wa.P, wb.P, sample))
		}
	}
}

func TestMathPublicRounding(t *testing.T) {
	r := oracle.Open(t, "math/PublicRounding")
	values := []float64{
		0, math.Copysign(0, -1), 0x1p-150, -0x1p-150, 0x1p-149, -0x1p-149,
		1, 1 + 0x1p-25, 1 - 0x1p-25, -1, -1 + 0x1p-25, -1 - 0x1p-25,
		1 + 0x1p-24, -1 - 0x1p-24,
		10000000.125, -10000000.125, 1000000000.25, -1000000000.25,
		math.MaxFloat32, -math.MaxFloat32,
		float64(math.Float32frombits(0x7f7ffffe)), -float64(math.Float32frombits(0x7f7ffffe)),
		math.Nextafter(math.MaxFloat32, 0), math.Nextafter(-math.MaxFloat32, 0),
	}
	r.I32("count", int32(len(values)))
	for i, input := range values {
		r.F32(mathPublicKey(i, "down"), box2d.RoundDownFloat(input))
		r.F32(mathPublicKey(i, "up"), box2d.RoundUpFloat(input))
	}
}

func TestMathPublicValidity(t *testing.T) {
	r := oracle.Open(t, "math/PublicValidity")
	fb := math.Float32frombits
	negativeZero := fb(0x80000000)
	nan, positiveInfinity, negativeInfinity := fb(0x7fc00001), float32(math.Inf(1)), float32(math.Inf(-1))
	values := []float32{0, negativeZero, 1, -1, 0x1p-126, math.MaxFloat32, nan, positiveInfinity, negativeInfinity}
	r.I32("numbers/count", int32(len(values)))
	for i, x := range values {
		r.Bool(fmt.Sprintf("numbers/%d/float", i), box2d.IsValidFloat(x))
		r.Bool(fmt.Sprintf("numbers/%d/vector-x", i), box2d.IsValidVec2(box2d.Vec2{X: x}))
		r.Bool(fmt.Sprintf("numbers/%d/vector-y", i), box2d.IsValidVec2(box2d.Vec2{Y: x}))
		r.Bool(fmt.Sprintf("numbers/%d/position-x", i), box2d.IsValidPosition(box2d.Pos{X: oracle.Real(x)}))
		r.Bool(fmt.Sprintf("numbers/%d/position-y", i), box2d.IsValidPosition(box2d.Pos{Y: oracle.Real(x)}))
		r.Bool(fmt.Sprintf("numbers/%d/transform", i), box2d.IsValidTransform(box2d.Transform{P: box2d.Vec2{X: x}, Q: box2d.Rot{C: 1}}))
		r.Bool(fmt.Sprintf("numbers/%d/world-transform", i), box2d.IsValidWorldTransform(box2d.WorldTransform{P: box2d.Pos{Y: oracle.Real(x)}, Q: box2d.Rot{C: 1}}))
		r.Bool(fmt.Sprintf("numbers/%d/plane-offset", i), box2d.IsValidPlane(box2d.Plane{Normal: box2d.Vec2{X: 1}, Offset: x}))
	}
	worldValues := []float64{0, math.Copysign(0, -1), math.MaxFloat32, -math.MaxFloat32, math.NaN(), math.Inf(1), math.Inf(-1)}
	if oracle.DoublePrecision {
		worldValues = append(worldValues, 1e100, -1e100, math.MaxFloat64, -math.MaxFloat64)
	}
	r.I32("world-numbers/count", int32(len(worldValues)))
	for i, value := range worldValues {
		px := box2d.Pos{X: oracle.Real(value)}
		py := box2d.Pos{Y: oracle.Real(value)}
		r.Bool(fmt.Sprintf("world-numbers/%d/x", i), box2d.IsValidPosition(px))
		r.Bool(fmt.Sprintf("world-numbers/%d/y", i), box2d.IsValidPosition(py))
		r.Bool(fmt.Sprintf("world-numbers/%d/transform", i), box2d.IsValidWorldTransform(box2d.WorldTransform{P: px, Q: box2d.Rot{C: 1}}))
	}
	rotations := []box2d.Rot{
		{C: 1}, {S: 1}, {C: -1, S: negativeZero}, {}, {C: 2}, {C: 0.9997}, {C: 1.0003},
		{C: nan}, {C: 1, S: nan}, {C: positiveInfinity}, {C: 1, S: negativeInfinity},
	}
	r.I32("rotation/count", int32(len(rotations)))
	for i, q := range rotations {
		r.Bool(fmt.Sprintf("rotation/%d/normalized", i), box2d.IsNormalizedRot(q))
		r.Bool(fmt.Sprintf("rotation/%d/valid", i), box2d.IsValidRotation(q))
		r.Bool(fmt.Sprintf("rotation/%d/transform", i), box2d.IsValidTransform(box2d.Transform{Q: q}))
		r.Bool(fmt.Sprintf("rotation/%d/world-transform", i), box2d.IsValidWorldTransform(box2d.WorldTransform{Q: q}))
		r.Bool(fmt.Sprintf("rotation/%d/plane-normal", i), box2d.IsValidPlane(box2d.Plane{Normal: box2d.Vec2{X: q.C, Y: q.S}}))
	}
	boxes := []box2d.AABB{
		{}, {LowerBound: box2d.Vec2{X: negativeZero, Y: negativeZero}},
		{LowerBound: box2d.Vec2{X: -1, Y: -2}, UpperBound: box2d.Vec2{X: 3, Y: 4}},
		{LowerBound: box2d.Vec2{X: 1}, UpperBound: box2d.Vec2{Y: 1}},
		{LowerBound: box2d.Vec2{Y: 2}, UpperBound: box2d.Vec2{X: 1, Y: 1}},
		{LowerBound: box2d.Vec2{X: negativeInfinity}, UpperBound: box2d.Vec2{X: 1, Y: 1}},
		{UpperBound: box2d.Vec2{X: 1, Y: positiveInfinity}},
		{UpperBound: box2d.Vec2{X: nan, Y: 1}},
	}
	r.I32("aabb/count", int32(len(boxes)))
	for i, input := range boxes {
		r.Bool(fmt.Sprintf("aabb/%d", i), box2d.IsValidAABB(input))
	}

	// This serial test temporarily owns process-global units. Restore even
	// when an oracle mismatch calls Fatal; never use t.Parallel here.
	oldScale := box2d.GetLengthUnitsPerMeter()
	t.Cleanup(func() { box2d.SetLengthUnitsPerMeter(oldScale) })
	for scaleIndex, scale := range []float32{1, 32} {
		box2d.SetLengthUnitsPerMeter(scale)
		r.F32(fmt.Sprintf("ray/%d/scale", scaleIndex), box2d.GetLengthUnitsPerMeter())
		huge := box2d.Huge()
		r.F32(fmt.Sprintf("ray/%d/huge", scaleIndex), huge)
		r.F32(fmt.Sprintf("ray/%d/linear-slop", scaleIndex), box2d.LinearSlop())
		r.F32(fmt.Sprintf("ray/%d/speculative-distance", scaleIndex), box2d.SpeculativeDistance())
		r.F32(fmt.Sprintf("ray/%d/contact-recycle-distance", scaleIndex), box2d.ContactRecycleDistance())
		r.F32(fmt.Sprintf("ray/%d/max-aabb-margin", scaleIndex), box2d.MaxAABBMargin())
		rays := []box2d.RayCastInput{
			{},
			{Origin: box2d.Vec2{X: negativeZero}, Translation: box2d.Vec2{X: 1, Y: negativeZero}, MaxFraction: negativeZero},
			{Origin: box2d.Vec2{X: 1, Y: 2}, Translation: box2d.Vec2{X: -3, Y: 4}, MaxFraction: 1},
			{Translation: box2d.Vec2{X: 1}, MaxFraction: -0x1p-149},
			{Translation: box2d.Vec2{X: 1}, MaxFraction: math.Nextafter32(huge, 0)},
			{Translation: box2d.Vec2{X: 1}, MaxFraction: huge},
			{Translation: box2d.Vec2{X: 1}, MaxFraction: math.Nextafter32(huge, positiveInfinity)},
			{Origin: box2d.Vec2{X: positiveInfinity}, Translation: box2d.Vec2{X: 1}, MaxFraction: 1},
			{Translation: box2d.Vec2{Y: nan}, MaxFraction: 1},
			{Translation: box2d.Vec2{X: 1}, MaxFraction: positiveInfinity},
			{Translation: box2d.Vec2{X: 1}, MaxFraction: nan},
		}
		r.I32(fmt.Sprintf("ray/%d/count", scaleIndex), int32(len(rays)))
		for i, input := range rays {
			r.Bool(fmt.Sprintf("ray/%d/%d", scaleIndex, i), box2d.IsValidRay(&input))
		}
	}
}

func TestMathPublicBoundsPlanes(t *testing.T) {
	r := oracle.Open(t, "math/PublicBoundsPlanes")
	negativeZero := math.Float32frombits(0x80000000)
	cases := []struct {
		a, b box2d.AABB
	}{
		{
			box2d.AABB{LowerBound: box2d.Vec2{X: -2, Y: -3}, UpperBound: box2d.Vec2{X: 4, Y: 5}},
			box2d.AABB{LowerBound: box2d.Vec2{X: -1, Y: -2}, UpperBound: box2d.Vec2{X: 3, Y: 4}},
		},
		{
			box2d.AABB{UpperBound: box2d.Vec2{X: 1, Y: 1}},
			box2d.AABB{LowerBound: box2d.Vec2{X: 1, Y: 1}, UpperBound: box2d.Vec2{X: 2, Y: 3}},
		},
		{
			box2d.AABB{UpperBound: box2d.Vec2{X: 1, Y: 1}},
			box2d.AABB{LowerBound: box2d.Vec2{X: 2, Y: -1}, UpperBound: box2d.Vec2{X: 3, Y: 2}},
		},
		{
			box2d.AABB{},
			box2d.AABB{LowerBound: box2d.Vec2{X: negativeZero, Y: negativeZero}},
		},
		{
			box2d.AABB{LowerBound: box2d.Vec2{X: 1024.125, Y: -2048.5}, UpperBound: box2d.Vec2{X: 1025.25, Y: -2046.25}},
			box2d.AABB{LowerBound: box2d.Vec2{X: 1023.5, Y: -2048}, UpperBound: box2d.Vec2{X: 1026, Y: -2047}},
		},
	}
	r.I32("bounds/count", int32(len(cases)))
	for i, c := range cases {
		mathPublicVec(r, fmt.Sprintf("bounds/%d/center", i), c.a.Center())
		mathPublicVec(r, fmt.Sprintf("bounds/%d/extents", i), c.a.Extents())
		combined := c.a.Union(c.b)
		mathPublicVec(r, fmt.Sprintf("bounds/%d/union-lower", i), combined.LowerBound)
		mathPublicVec(r, fmt.Sprintf("bounds/%d/union-upper", i), combined.UpperBound)
		r.Bool(fmt.Sprintf("bounds/%d/contains", i), c.a.Contains(c.b))
		r.Bool(fmt.Sprintf("bounds/%d/contained", i), c.b.Contains(c.a))
		r.Bool(fmt.Sprintf("bounds/%d/overlaps", i), c.a.Overlaps(c.b))
	}
	points := []box2d.Vec2{{X: 2.5, Y: -3.75}, {X: -4, Y: 0.125}, {X: 1, Y: 6.25}}
	clouds := []struct {
		count  int
		radius float32
	}{{1, 0}, {3, 0}, {3, 0.125}}
	r.I32("cloud/count", int32(len(clouds)))
	for i, cloud := range clouds {
		bounds := box2d.MakeAABB(points, cloud.count, cloud.radius)
		mathPublicVec(r, fmt.Sprintf("cloud/%d/lower", i), bounds.LowerBound)
		mathPublicVec(r, fmt.Sprintf("cloud/%d/upper", i), bounds.UpperBound)
	}
	planes := []struct {
		plane box2d.Plane
		point box2d.Vec2
	}{
		{box2d.Plane{Normal: box2d.Vec2{X: 1}, Offset: 2}, box2d.Vec2{X: 2, Y: -3}},
		{box2d.Plane{Normal: box2d.Vec2{X: 1}, Offset: 2}, box2d.Vec2{X: 3, Y: -3}},
		{box2d.Plane{Normal: box2d.Vec2{Y: -1}, Offset: -2}, box2d.Vec2{X: 3, Y: 4}},
		{box2d.Plane{Normal: box2d.Vec2{X: 0.6, Y: 0.8}, Offset: 0.25}, box2d.Vec2{X: -1.25, Y: 2.5}},
		{box2d.Plane{Normal: box2d.Vec2{X: -1, Y: negativeZero}, Offset: negativeZero}, box2d.Vec2{X: negativeZero}},
	}
	r.I32("plane/count", int32(len(planes)))
	for i, input := range planes {
		r.F32(fmt.Sprintf("plane/%d/separation", i), box2d.PlaneSeparation(input.plane, input.point))
	}
}
