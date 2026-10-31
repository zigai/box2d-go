// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"testing"
	"unsafe"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

func TestMath(t *testing.T) {
	checks := openMath(t)
	testMathAngles(t, checks)
	testMathAtanGrid(t, checks)
	testMathAtanAxes(t, checks)
	testMathVectors(t, checks)
	testMathTransforms(t, checks)
	testMathUnitRotations(t, checks)
	testMathRotationInterpolation(t, checks)
	testMathRelativeAngles(t, checks)
	testMathWorldTransforms(t, checks)
	testMathLargeWorldPositions(t, checks)
	checks.close()
}

func testMathAngles(t *testing.T, checks *mathChecks) {
	t.Helper()

	for angleT := float32(-10.0); angleT < 10.0; angleT += 0.01 {
		angle := float32(box2d.Pi * angleT)
		r := box2d.MakeRot(angle)
		c := checks.reference()
		s := checks.reference()

		// The cosine and sine approximations are accurate to about 0.1 degrees (0.002 radians)
		// printf( "%g %g\n", r.C - c, r.S - s )
		checks.f32(25, r.C-c)

		if v := r.C - c; v < -0.002 || 0.002 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.002)
		}

		checks.f32(26, r.S-s)

		if v := r.S - s; v < -0.002 || 0.002 < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.002)
		}

		checkMathUnwoundAngle(t, checks, angle, s, c)
	}
}

func checkMathUnwoundAngle(t *testing.T, checks *mathChecks, angle, s, c float32) {
	t.Helper()

	xn := box2d.UnwindAngle(angle)
	checks.f32(285, xn)
	checks.boolean(29, -box2d.Pi <= xn && xn <= box2d.Pi)

	if !(-box2d.Pi <= xn && xn <= box2d.Pi) {
		t.Fatal("condition false")
	}

	a := box2d.Atan2(s, c)
	checks.f32(288, a)
	checks.boolean(32, box2d.IsValidFloat(a))

	if !box2d.IsValidFloat(a) {
		t.Fatal("condition false")
	}

	diff := float32(math.Abs(float64(a - xn)))

	// The two results can be off by 360 degrees (-pi and pi)
	if diff > box2d.Pi {
		diff -= 2.0 * box2d.Pi
	}

	// The approximate atan2 is quite accurate
	checks.f32(43, diff)

	if v := diff; v < -float32(0.00004) || float32(0.00004) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
	}
}

func testMathAtanGrid(t *testing.T, checks *mathChecks) {
	t.Helper()

	for y := float32(-1.0); y <= 1.0; y += 0.01 {
		for x := float32(-1.0); x <= 1.0; x += 0.01 {
			a1 := box2d.Atan2(y, x)
			a2 := checks.reference()
			diff := float32(math.Abs(float64(a1 - a2)))
			checks.f32(309, a1)
			checks.boolean(53, box2d.IsValidFloat(a1))

			if !box2d.IsValidFloat(a1) {
				t.Fatal("condition false")
			}

			checks.f32(54, diff)

			if v := diff; v < -float32(0.00004) || float32(0.00004) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
			}
		}
	}
}

func testMathAtanAxes(t *testing.T, checks *mathChecks) {
	t.Helper()

	checkMathAtanAxis(t, checks, 1, 0, 318, 62, 63)
	checkMathAtanAxis(t, checks, -1, 0, 326, 70, 71)
	checkMathAtanAxis(t, checks, 0, 1, 334, 78, 79)
	checkMathAtanAxis(t, checks, 0, -1, 342, 86, 87)
	checkMathAtanAxis(t, checks, 0, 0, 350, 94, 95)
}

func checkMathAtanAxis(t *testing.T, checks *mathChecks, y, x float32, valueSite, validSite, diffSite int) {
	t.Helper()

	a1 := box2d.Atan2(y, x)
	a2 := checks.reference()
	diff := float32(math.Abs(float64(a1 - a2)))
	checks.f32(valueSite, a1)
	checks.boolean(validSite, box2d.IsValidFloat(a1))

	if !box2d.IsValidFloat(a1) {
		t.Fatal("condition false")
	}

	checks.f32(diffSite, diff)

	if v := diff; v < -float32(0.00004) || float32(0.00004) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
	}
}

func testMathVectors(t *testing.T, checks *mathChecks) {
	t.Helper()

	zero := box2d.Vec2Zero
	one := box2d.Vec2{1.0, 1.0}
	two := box2d.Vec2{2.0, 2.0}

	v := one.Add(two)
	checks.f32(359, v.X)
	checks.f32(615, v.Y)
	checks.boolean(103, v.X == 3.0 && v.Y == 3.0)

	if !(v.X == 3.0 && v.Y == 3.0) {
		t.Fatal("condition false")
	}

	v = zero.Sub(two)
	checks.f32(362, v.X)
	checks.f32(618, v.Y)
	checks.boolean(106, v.X == -2.0 && v.Y == -2.0)

	if !(v.X == -2.0 && v.Y == -2.0) {
		t.Fatal("condition false")
	}

	v = two.Add(two)
	checks.f32(365, v.X)
	checks.f32(621, v.Y)
	checks.boolean(109, v.X != 5.0 && v.Y != 5.0)

	if !(v.X != 5.0 && v.Y != 5.0) {
		t.Fatal("condition false")
	}
}

func testMathTransforms(t *testing.T, checks *mathChecks) {
	t.Helper()

	two := box2d.Vec2{2.0, 2.0}

	var v box2d.Vec2

	transform1 := box2d.Transform{box2d.Vec2{-2.0, 3.0}, box2d.MakeRot(1.0)}
	transform2 := box2d.Transform{box2d.Vec2{1.0, 0.0}, box2d.MakeRot(-2.0)}

	transform := transform2.Mul(transform1)

	v = transform2.Apply(transform1.Apply(two))
	u := transform.Apply(two)

	checks.f32(120, u.X-v.X)

	if v := u.X - v.X; v < -(10.0*float32(0x1p-23)) || (10.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 10.0*float32(0x1p-23))
	}

	checks.f32(121, u.Y-v.Y)

	if v := u.Y - v.Y; v < -(10.0*float32(0x1p-23)) || (10.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 10.0*float32(0x1p-23))
	}

	v = transform1.Apply(two)
	v = transform1.InvApply(v)
	checks.f32(126, v.X-two.X)

	if v := v.X - two.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}

	checks.f32(127, v.Y-two.Y)

	if v := v.Y - two.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}
}

func testMathUnitRotations(t *testing.T, checks *mathChecks) {
	t.Helper()

	var v, u box2d.Vec2

	v = box2d.Vec2{0.2, -0.5}.Normalize()

	for y := float32(-1.0); y <= 1.0; y += 0.01 {
		for x := float32(-1.0); x <= 1.0; x += 0.01 {
			if x == 0.0 && y == 0.0 {
				continue
			}

			u = box2d.Vec2{x, y}.Normalize()
			r := box2d.ComputeRotationBetweenUnitVectors(v, u)

			w := r.Rotate(v)
			checks.f32(144, w.X-u.X)

			if v := w.X - u.X; v < -(4.0*float32(0x1p-23)) || (4.0*float32(0x1p-23)) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 4.0*float32(0x1p-23))
			}

			checks.f32(145, w.Y-u.Y)

			if v := w.Y - u.Y; v < -(4.0*float32(0x1p-23)) || (4.0*float32(0x1p-23)) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 4.0*float32(0x1p-23))
			}
		}
	}
}

func testMathRotationInterpolation(t *testing.T, checks *mathChecks) {
	t.Helper()

	// NLerp of Rot has an error of over 4 degrees.
	// 2D quaternions should have an error under 1 degree.
	q1 := box2d.RotIdentity
	q2 := box2d.MakeRot(0.5 * box2d.Pi)

	n := 100
	for i := 0; i <= n; i++ {
		alpha := float32(i) / float32(n)
		q := q1.NLerp(q2, alpha)
		angle := q.Angle()
		checks.f32(159, float32(float32(alpha*0.5)*box2d.Pi)-angle)

		if v := float32(float32(alpha*0.5)*box2d.Pi) - angle; v < -(5.0*box2d.Pi/180.0) || (5.0*box2d.Pi/180.0) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 5.0*box2d.Pi/180.0)
		}
		// printf("angle = [%g %g %g]\n", alpha, alpha * 0.5 * box2d.Pi, angle)
	}
}

func testMathRelativeAngles(t *testing.T, checks *mathChecks) {
	t.Helper()

	var q1, q2 box2d.Rot
	// Test relative angle
	baseAngle := 0.75 * box2d.Pi
	q1 = box2d.MakeRot(baseAngle)

	for angleT := float32(-10.0); angleT < 10.0; angleT += 0.01 {
		angle := float32(box2d.Pi * angleT)
		q2 = box2d.MakeRot(angle)
		relativeAngle := q1.RelativeAngle(q2)
		unwoundAngle := box2d.UnwindAngle(angle - baseAngle)
		tolerance := 0.1 * box2d.Pi / 180.0

		checks.f32(174, relativeAngle-unwoundAngle)

		if v := relativeAngle - unwoundAngle; v < -tolerance || tolerance < v {
			t.Fatalf("condition false: abs(%v) < %v", v, tolerance)
		}
	}
}

func testMathWorldTransforms(t *testing.T, checks *mathChecks) {
	t.Helper()

	// World position boundary helpers. With large world mode off these collapse to the float
	// ops, so the round trips hold in both builds.
	{
		d := box2d.Vec2{0.25, -0.5}
		base := box2d.Pos{10.0, -20.0}
		p := base.Add(d)
		back := p.Sub(base)
		checks.f32(184, back.X-d.X)

		if v := back.X - d.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}

		checks.f32(185, back.Y-d.Y)

		if v := back.Y - d.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}

		r := base.Vec2()
		checks.f32(444, r.X)
		checks.f32(700, r.Y)
		checks.boolean(188, r.X == 10.0 && r.Y == -20.0)

		if !(r.X == 10.0 && r.Y == -20.0) {
			t.Fatal("condition false")
		}

		checks.boolean(190, p.IsValid())

		if !p.IsValid() {
			t.Fatal("condition false")
		}

		checks.boolean(191, box2d.PosZero.IsValid())

		if !box2d.PosZero.IsValid() {
			t.Fatal("condition false")
		}

		checks.boolean(192, box2d.WorldTransformIdentity.IsValid())

		if !box2d.WorldTransformIdentity.IsValid() {
			t.Fatal("condition false")
		}
	}

	testMathWorldTransformRoundTrips(t, checks)
}

func testMathWorldTransformRoundTrips(t *testing.T, checks *mathChecks) {
	t.Helper()

	wt := box2d.WorldTransform{box2d.Pos{3.0, -4.0}, box2d.MakeRot(0.7)}
	checks.boolean(195, wt.IsValid())

	if !wt.IsValid() {
		t.Fatal("condition false")
	}

	// Local to world to local round trip
	local := box2d.Vec2{1.5, 2.5}
	world := wt.Apply(local)
	backLocal := wt.InvApply(world)
	checks.f32(201, backLocal.X-local.X)

	if v := backLocal.X - local.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}

	checks.f32(202, backLocal.Y-local.Y)

	if v := backLocal.Y - local.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}

	// Relative transform of B in a matches a float reference at modest coordinates
	a := box2d.WorldTransform{box2d.Pos{-2.0, 3.0}, box2d.MakeRot(1.0)}
	b := box2d.WorldTransform{box2d.Pos{1.0, 0.0}, box2d.MakeRot(-2.0)}
	rel := a.InvMul(b)
	refA := box2d.Transform{a.P.Vec2(), a.Q}
	refB := box2d.Transform{b.P.Vec2(), b.Q}
	ref := refA.InvMul(refB)
	checks.f32(211, rel.P.X-ref.P.X)

	if v := rel.P.X - ref.P.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}

	checks.f32(212, rel.P.Y-ref.P.Y)

	if v := rel.P.Y - ref.P.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}
}

func testMathLargeWorldPositions(t *testing.T, checks *mathChecks) {
	t.Helper()

	if unsafe.Sizeof(box2d.Pos{}.X) == 8 {
		// Far from the origin a float vector cannot resolve sub meter motion, but a double world
		// position can. This is the whole point of large world mode.
		{
			d := box2d.Vec2{0.25, -0.5}
			base := box2d.Pos{1.0e7, 0.0}
			p := base.Add(d)
			back := p.Sub(base)
			checks.f32(223, back.X-d.X)

			if v := back.X - d.X; v < -1.0e-4 || 1.0e-4 < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 1.0e-4)
			}

			checks.f32(224, back.Y-d.Y)

			if v := back.Y - d.Y; v < -1.0e-4 || 1.0e-4 < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 1.0e-4)
			}
		}
	}
}

type mathChecks struct {
	t       *testing.T
	data    []byte
	offset  int
	ordinal int
}

func openMath(t *testing.T) *mathChecks {
	t.Helper()
	r := oracle.Open(t, "math/MathTest")

	config := "float"
	if unsafe.Sizeof(box2d.Pos{}.X) == 8 {
		config = "double"
	}

	data, err := oracle.ReadFile(config, "math/MathTest")
	if err != nil {
		t.Fatal(err)
	}

	var rec struct {
		Key   string `json:"k"`
		Type  string `json:"t"`
		Value string `json:"v"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(data), &rec); err != nil || rec.Key != "checks" {
		t.Fatalf("math fixture: want one checks record: %v", err)
	}

	checks, err := hex.DecodeString(rec.Value)
	if err != nil {
		t.Fatal(err)
	}

	r.Bytes("checks", checks)

	return &mathChecks{t: t, data: checks}
}

func (r *mathChecks) take(width int) []byte {
	r.t.Helper()

	if r.offset+width > len(r.data) {
		r.t.Fatalf("math fixture exhausted at checkpoint %d", r.ordinal)
	}

	value := r.data[r.offset : r.offset+width]
	r.offset += width

	return value
}

func (r *mathChecks) reference() float32 {
	r.t.Helper()

	r.ordinal++

	return math.Float32frombits(binary.LittleEndian.Uint32(r.take(4)))
}

func (r *mathChecks) f32(site int, value float32) {
	r.t.Helper()

	want := r.reference()
	if math.Float32bits(want) != math.Float32bits(value) {
		ordered := func(v float32) int64 {
			bits := math.Float32bits(v)
			if bits>>31 != 0 {
				return -int64(bits&0x7fffffff) - 1
			}

			return int64(bits)
		}

		distance := ordered(want) - ordered(value)
		if distance < 0 {
			distance = -distance
		}

		r.t.Fatalf("math first divergence: upstream line %d, site %d, checkpoint %d: C 0x%08x (%g), Go 0x%08x (%g), ULP distance %d", site%256, site, r.ordinal, math.Float32bits(want), want, math.Float32bits(value), value, distance)
	}
}

func (r *mathChecks) boolean(site int, value bool) {
	r.t.Helper()
	r.ordinal++

	want := r.take(1)[0] != 0
	if want != value {
		r.t.Fatalf("math first divergence: upstream line %d, checkpoint %d: C %v, Go %v", site%256, r.ordinal, want, value)
	}
}

func (r *mathChecks) close() {
	r.t.Helper()

	if r.offset != len(r.data) {
		r.t.Fatalf("math fixture: %d unread bytes after checkpoint %d", len(r.data)-r.offset, r.ordinal)
	}
}
