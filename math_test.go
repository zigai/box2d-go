// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"

	"math"
	"os"
	"strconv"
	"testing"
	"unsafe"
)

func TestMath(t *testing.T) {

	checks := collisionOpenMath(t)
	for angleT := float32(-10.0); angleT < 10.0; angleT += 0.01 {
		angle := float32(box2d.Pi * angleT)
		r := box2d.MakeRot(angle)
		c := checks.reference(20)
		s := checks.reference(21)

		// The cosine and sine approximations are accurate to about 0.1 degrees (0.002 radians)
		// printf( "%g %g\n", r.C - c, r.S - s )
		checks.f32(25, r.C-c)
		if v := r.C - c; v < -(0.002) || (0.002) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.002)
		}
		checks.f32(26, r.S-s)
		if v := r.S - s; v < -(0.002) || (0.002) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 0.002)
		}

		xn := box2d.UnwindAngle(angle)
		checks.f32(285, xn)
		checks.boolean(29, -box2d.Pi <= xn && xn <= box2d.Pi)
		if !(-box2d.Pi <= xn && xn <= box2d.Pi) {
			t.Fatal("condition false")
		}

		a := box2d.Atan2(s, c)
		checks.f32(288, a)
		checks.boolean(32, box2d.IsValidFloat(a))
		if !(box2d.IsValidFloat(a)) {
			t.Fatal("condition false")
		}

		diff := box2d.AbsFloat(a - xn)

		// The two results can be off by 360 degrees (-pi and pi)
		if diff > box2d.Pi {
			diff -= 2.0 * box2d.Pi
		}

		// The approximate atan2 is quite accurate
		checks.f32(43, diff)
		if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
		}
	}

	for y := float32(-1.0); y <= 1.0; y += 0.01 {
		for x := float32(-1.0); x <= 1.0; x += 0.01 {
			a1 := box2d.Atan2(y, x)
			a2 := checks.reference(51)
			diff := box2d.AbsFloat(a1 - a2)
			checks.f32(309, a1)
			checks.boolean(53, box2d.IsValidFloat(a1))
			if !(box2d.IsValidFloat(a1)) {
				t.Fatal("condition false")
			}
			checks.f32(54, diff)
			if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
			}
		}
	}

	{
		a1 := box2d.Atan2(1.0, 0.0)
		a2 := checks.reference(60)
		diff := box2d.AbsFloat(a1 - a2)
		checks.f32(318, a1)
		checks.boolean(62, box2d.IsValidFloat(a1))
		if !(box2d.IsValidFloat(a1)) {
			t.Fatal("condition false")
		}
		checks.f32(63, diff)
		if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
		}
	}

	{
		a1 := box2d.Atan2(-1.0, 0.0)
		a2 := checks.reference(68)
		diff := box2d.AbsFloat(a1 - a2)
		checks.f32(326, a1)
		checks.boolean(70, box2d.IsValidFloat(a1))
		if !(box2d.IsValidFloat(a1)) {
			t.Fatal("condition false")
		}
		checks.f32(71, diff)
		if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
		}
	}

	{
		a1 := box2d.Atan2(0.0, 1.0)
		a2 := checks.reference(76)
		diff := box2d.AbsFloat(a1 - a2)
		checks.f32(334, a1)
		checks.boolean(78, box2d.IsValidFloat(a1))
		if !(box2d.IsValidFloat(a1)) {
			t.Fatal("condition false")
		}
		checks.f32(79, diff)
		if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
		}
	}

	{
		a1 := box2d.Atan2(0.0, -1.0)
		a2 := checks.reference(84)
		diff := box2d.AbsFloat(a1 - a2)
		checks.f32(342, a1)
		checks.boolean(86, box2d.IsValidFloat(a1))
		if !(box2d.IsValidFloat(a1)) {
			t.Fatal("condition false")
		}
		checks.f32(87, diff)
		if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
		}
	}

	{
		a1 := box2d.Atan2(0.0, 0.0)
		a2 := checks.reference(92)
		diff := box2d.AbsFloat(a1 - a2)
		checks.f32(350, a1)
		checks.boolean(94, box2d.IsValidFloat(a1))
		if !(box2d.IsValidFloat(a1)) {
			t.Fatal("condition false")
		}
		checks.f32(95, diff)
		if v := diff; v < -(float32(0.00004)) || (float32(0.00004)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, float32(0.00004))
		}
	}

	zero := box2d.Vec2Zero
	one := box2d.Vec2{1.0, 1.0}
	two := box2d.Vec2{2.0, 2.0}

	v := box2d.Add(one, two)
	checks.f32(359, v.X)
	checks.f32(615, v.Y)
	checks.boolean(103, v.X == 3.0 && v.Y == 3.0)
	if !(v.X == 3.0 && v.Y == 3.0) {
		t.Fatal("condition false")
	}

	v = box2d.Sub(zero, two)
	checks.f32(362, v.X)
	checks.f32(618, v.Y)
	checks.boolean(106, v.X == -2.0 && v.Y == -2.0)
	if !(v.X == -2.0 && v.Y == -2.0) {
		t.Fatal("condition false")
	}

	v = box2d.Add(two, two)
	checks.f32(365, v.X)
	checks.f32(621, v.Y)
	checks.boolean(109, v.X != 5.0 && v.Y != 5.0)
	if !(v.X != 5.0 && v.Y != 5.0) {
		t.Fatal("condition false")
	}

	transform1 := box2d.Transform{box2d.Vec2{-2.0, 3.0}, box2d.MakeRot(1.0)}
	transform2 := box2d.Transform{box2d.Vec2{1.0, 0.0}, box2d.MakeRot(-2.0)}

	transform := box2d.MulTransforms(transform2, transform1)

	v = box2d.TransformPoint(transform2, box2d.TransformPoint(transform1, two))
	u := box2d.TransformPoint(transform, two)

	checks.f32(120, u.X-v.X)
	if v := u.X - v.X; v < -(10.0*float32(0x1p-23)) || (10.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 10.0*float32(0x1p-23))
	}
	checks.f32(121, u.Y-v.Y)
	if v := u.Y - v.Y; v < -(10.0*float32(0x1p-23)) || (10.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 10.0*float32(0x1p-23))
	}

	v = box2d.TransformPoint(transform1, two)
	v = box2d.InvTransformPoint(transform1, v)
	checks.f32(126, v.X-two.X)
	if v := v.X - two.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}
	checks.f32(127, v.Y-two.Y)
	if v := v.Y - two.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
		t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
	}

	v = box2d.Normalize(box2d.Vec2{0.2, -0.5})
	for y := float32(-1.0); y <= 1.0; y += 0.01 {
		for x := float32(-1.0); x <= 1.0; x += 0.01 {
			if x == 0.0 && y == 0.0 {
				continue
			}

			u = box2d.Normalize(box2d.Vec2{x, y})
			r := box2d.ComputeRotationBetweenUnitVectors(v, u)

			w := box2d.RotateVector(r, v)
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

	// NLerp of Rot has an error of over 4 degrees.
	// 2D quaternions should have an error under 1 degree.
	q1 := box2d.RotIdentity
	q2 := box2d.MakeRot(0.5 * box2d.Pi)
	n := 100
	for i := 0; i <= n; i++ {
		alpha := float32(i) / float32(n)
		q := box2d.NLerp(q1, q2, alpha)
		angle := q.Angle()
		checks.f32(159, float32(float32(alpha*0.5)*box2d.Pi)-angle)
		if v := float32(float32(alpha*0.5)*box2d.Pi) - angle; v < -(5.0*box2d.Pi/180.0) || (5.0*box2d.Pi/180.0) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 5.0*box2d.Pi/180.0)
		}
		//printf("angle = [%g %g %g]\n", alpha, alpha * 0.5 * box2d.Pi, angle)
	}

	// Test relative angle
	baseAngle := 0.75 * box2d.Pi
	q1 = box2d.MakeRot(baseAngle)
	for angleT := float32(-10.0); angleT < 10.0; angleT += 0.01 {
		angle := float32(box2d.Pi * angleT)
		q2 = box2d.MakeRot(angle)
		relativeAngle := box2d.RelativeAngle(q1, q2)
		unwoundAngle := box2d.UnwindAngle(angle - baseAngle)
		tolerance := 0.1 * box2d.Pi / 180.0
		checks.f32(174, relativeAngle-unwoundAngle)
		if v := relativeAngle - unwoundAngle; v < -(tolerance) || (tolerance) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, tolerance)
		}
	}

	// World position boundary helpers. With large world mode off these collapse to the float
	// ops, so the round trips hold in both builds.
	{
		d := box2d.Vec2{0.25, -0.5}
		base := box2d.Pos{10.0, -20.0}
		p := box2d.OffsetPos(base, d)
		back := box2d.SubPos(p, base)
		checks.f32(184, back.X-d.X)
		if v := back.X - d.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}
		checks.f32(185, back.Y-d.Y)
		if v := back.Y - d.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}

		r := box2d.ToVec2(base)
		checks.f32(444, r.X)
		checks.f32(700, r.Y)
		checks.boolean(188, r.X == 10.0 && r.Y == -20.0)
		if !(r.X == 10.0 && r.Y == -20.0) {
			t.Fatal("condition false")
		}

		checks.boolean(190, box2d.IsValidPosition(p))
		if !(box2d.IsValidPosition(p)) {
			t.Fatal("condition false")
		}
		checks.boolean(191, box2d.IsValidPosition(box2d.PosZero))
		if !(box2d.IsValidPosition(box2d.PosZero)) {
			t.Fatal("condition false")
		}
		checks.boolean(192, box2d.IsValidWorldTransform(box2d.WorldTransformIdentity))
		if !(box2d.IsValidWorldTransform(box2d.WorldTransformIdentity)) {
			t.Fatal("condition false")
		}

		wt := box2d.WorldTransform{box2d.Pos{3.0, -4.0}, box2d.MakeRot(0.7)}
		checks.boolean(195, box2d.IsValidWorldTransform(wt))
		if !(box2d.IsValidWorldTransform(wt)) {
			t.Fatal("condition false")
		}

		// Local to world to local round trip
		local := box2d.Vec2{1.5, 2.5}
		world := box2d.TransformWorldPoint(wt, local)
		backLocal := box2d.InvTransformWorldPoint(wt, world)
		checks.f32(201, backLocal.X-local.X)
		if v := backLocal.X - local.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}
		checks.f32(202, backLocal.Y-local.Y)
		if v := backLocal.Y - local.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}

		// Relative transform of B in A matches a float reference at modest coordinates
		A := box2d.WorldTransform{box2d.Pos{-2.0, 3.0}, box2d.MakeRot(1.0)}
		B := box2d.WorldTransform{box2d.Pos{1.0, 0.0}, box2d.MakeRot(-2.0)}
		rel := box2d.InvMulWorldTransforms(A, B)
		refA := box2d.Transform{box2d.ToVec2(A.P), A.Q}
		refB := box2d.Transform{box2d.ToVec2(B.P), B.Q}
		ref := box2d.InvMulTransforms(refA, refB)
		checks.f32(211, rel.P.X-ref.P.X)
		if v := rel.P.X - ref.P.X; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}
		checks.f32(212, rel.P.Y-ref.P.Y)
		if v := rel.P.Y - ref.P.Y; v < -(8.0*float32(0x1p-23)) || (8.0*float32(0x1p-23)) < v {
			t.Fatalf("condition false: abs(%v) < %v", v, 8.0*float32(0x1p-23))
		}
	}

	if unsafe.Sizeof(box2d.Pos{}.X) == 8 {
		// Far from the origin a float vector cannot resolve sub meter motion, but a double world
		// position can. This is the whole point of large world mode.
		{
			d := box2d.Vec2{0.25, -0.5}
			base := box2d.Pos{1.0e7, 0.0}
			p := box2d.OffsetPos(base, d)
			back := box2d.SubPos(p, base)
			checks.f32(223, back.X-d.X)
			if v := back.X - d.X; v < -(1.0e-4) || (1.0e-4) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 1.0e-4)
			}
			checks.f32(224, back.Y-d.Y)
			if v := back.Y - d.Y; v < -(1.0e-4) || (1.0e-4) < v {
				t.Fatalf("condition false: abs(%v) < %v", v, 1.0e-4)
			}
		}
	}

	checks.close()

}

type collisionMathChecks struct {
	t        *testing.T
	data     []byte
	bit      uint32
	limit    uint32
	previous [1024]uint32
	leading  [1024]uint32
	width    [1024]uint32
	ordinal  int
}

func collisionOpenMath(t *testing.T) *collisionMathChecks {
	t.Helper()
	r := oracle.Open(t, "math/MathTest")
	config := "float"
	if unsafe.Sizeof(box2d.Pos{}.X) == 8 {
		config = "double"
	}
	data, err := os.ReadFile("testdata/oracle/" + config + "/math/MathTest.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	var records []struct {
		Key   string `json:"k"`
		Type  string `json:"t"`
		Value string `json:"v"`
	}
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		var rec struct {
			Key   string `json:"k"`
			Type  string `json:"t"`
			Value string `json:"v"`
		}
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatal(err)
		}
		records = append(records, rec)
	}
	if len(records) != 2 {
		t.Fatalf("math fixture: got %d records, want 2", len(records))
	}
	limit, err := strconv.ParseUint(records[0].Value, 16, 32)
	if err != nil {
		t.Fatal(err)
	}
	packed, err := hex.DecodeString(records[1].Value)
	if err != nil {
		t.Fatal(err)
	}
	r.U32("bits", uint32(limit))
	r.Bytes("checks", packed)
	if uint64(len(packed)) != (limit+7)/8 {
		t.Fatal("math fixture: invalid bit count")
	}
	return &collisionMathChecks{t: t, data: packed, limit: uint32(limit)}
}

func (r *collisionMathChecks) bits(count uint32) uint32 {
	r.t.Helper()
	if r.bit+count > r.limit {
		r.t.Fatalf("math fixture exhausted at checkpoint %d, bit %d", r.ordinal, r.bit)
	}
	var value uint32
	for i := uint32(0); i < count; i++ {
		value |= uint32((r.data[r.bit/8]>>(r.bit%8))&1) << i
		r.bit++
	}
	return value
}

func (r *collisionMathChecks) reference(site int) float32 {
	r.t.Helper()
	r.ordinal++
	if r.bits(1) != 0 {
		if r.bits(1) != 0 {
			r.leading[site] = r.bits(5)
			r.width[site] = r.bits(5) + 1
		}
		leading, width := r.leading[site], r.width[site]
		if width == 0 || leading+width > 32 {
			r.t.Fatalf("math fixture: invalid delta at checkpoint %d", r.ordinal)
		}
		r.previous[site] ^= r.bits(width) << (32 - leading - width)
	}
	return math.Float32frombits(r.previous[site])
}

func (r *collisionMathChecks) f32(site int, value float32) {
	r.t.Helper()
	want := r.reference(site)
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

func (r *collisionMathChecks) boolean(site int, value bool) {
	r.t.Helper()
	r.ordinal++
	want := r.bits(1) != 0
	if want != value {
		r.t.Fatalf("math first divergence: upstream line %d, checkpoint %d: C %v, Go %v", site%256, r.ordinal, want, value)
	}
}

func (r *collisionMathChecks) close() {
	r.t.Helper()
	if r.bit != r.limit {
		r.t.Fatalf("math fixture: %d unread bits after checkpoint %d", r.limit-r.bit, r.ordinal)
	}
}
