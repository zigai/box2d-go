// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"math"
	"testing"
)

func TestContactSolverWideArithmetic(t *testing.T) {
	a := setW(1, -2, 3, -4)
	b := setW(2, 4, -6, -8)

	cases := []struct {
		name      string
		got, want floatW
	}{
		{"zero", zeroW(), floatW{}},
		{"splat", splatW(3), floatW{3, 3, 3, 3}},
		{"neg", negW(a), floatW{-1, 2, -3, 4}},
		{"add", addW(a, b), floatW{3, 2, -3, -12}},
		{"sub", subW(a, b), floatW{-1, -6, 9, 4}},
		{"mul", mulW(a, b), floatW{2, -8, -18, 32}},
		{"div", divW(a, b), floatW{0.5, -0.5, -0.5, 0.5}},
		{"min", minW(a, b), floatW{1, -2, -6, -8}},
		{"max", maxW(a, b), floatW{2, 4, 3, -4}},
		{"clamp", symClampW(a, splatW(2)), floatW{1, -2, 2, -2}},
		{"greater", greaterThanW(a, b), floatW{0, 0, 1, 1}},
		{"less", lessThanW(a, b), floatW{1, 1, 0, 0}},
		{"and", andW(a, setW(0, 1, -1, 0)), floatW{0, 1, 1, 0}},
		{"blend", blendW(a, b, setW(0, 1, -1, 0)), floatW{1, 4, -6, -4}},
		{"soft", softMaskW([]int{0, 1, 2, 0}, []int{1, 0, 3, 0}), floatW{1, 1, 0, 1}},
		{"lo", unpackLoW(a, b), floatW{1, 2, -2, 4}},
		{"hi", unpackHiW(a, b), floatW{3, -6, -4, -8}},
		{"mulAdd rounding", mulAddW(splatW(-1), splatW(math.Float32frombits(0x3f800001)), splatW(math.Float32frombits(0x3f7ffffe))), floatW{}},
		{"mulSub rounding", mulSubW(splatW(1), splatW(math.Float32frombits(0x3f800001)), splatW(math.Float32frombits(0x3f7ffffe))), floatW{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Fatalf("got %+v, want %+v", c.got, c.want)
			}
		})
	}
}

func TestContactSolverWideMaskReduction(t *testing.T) {
	a := setW(1, -2, 3, -4)

	if !allZeroW(zeroW()) || allZeroW(a) || anyTrueW(zeroW()) || !anyTrueW(setW(0, 0, 0, -1)) {
		t.Fatal("mask reduction")
	}
}

func TestContactSolverWideLoadStore(t *testing.T) {
	data := []float32{1, 2, 3, 4, 5}
	v := loadU(data)
	storeW(data, negW(v))

	if data[0] != -1 || data[1] != -2 || data[2] != -3 || data[3] != -4 || data[4] != 5 {
		t.Fatal(data)
	}
}

func TestContactSolverWideVectorArithmetic(t *testing.T) {
	v2 := vec2W{splatW(2), splatW(3)}

	w2 := vec2W{splatW(4), splatW(5)}
	if dotW(v2, w2) != splatW(23) || crossW(v2, w2) != splatW(-2) || subVW(v2, w2) != (vec2W{splatW(-2), splatW(-2)}) || rightPerpW(v2) != (vec2W{splatW(3), splatW(-2)}) || rotateVectorW(rotW{zeroW(), splatW(1)}, v2) != (vec2W{splatW(-3), splatW(2)}) {
		t.Fatal("wide vector arithmetic")
	}

	negativeZero := math.Float32frombits(0x80000000)
	upper := math.Float32frombits(0x3f800001)
	lower := math.Float32frombits(0x3f7ffffe)
	subnormal := math.Float32frombits(1)

	cases := []struct {
		name string
		got  floatW
		want [4]uint32
	}{
		{
			name: "dot rounding and signed zero",
			got: dotW(
				vec2W{setW(negativeZero, upper, subnormal, 2), setW(negativeZero, -1, 0, 3)},
				vec2W{setW(1, lower, 0.5, 4), setW(1, 1, 1, 5)},
			),
			want: [4]uint32{0x80000000, 0, 0, 0x41b80000},
		},
		{
			name: "cross rounding and signed zero",
			got: crossW(
				vec2W{setW(negativeZero, upper, subnormal, 2), setW(0, 1, 0, 3)},
				vec2W{setW(1, 1, 1, 4), setW(1, lower, 0.5, 5)},
			),
			want: [4]uint32{0x80000000, 0, 0, 0xc0000000},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := [4]uint32{math.Float32bits(test.got.x), math.Float32bits(test.got.y), math.Float32bits(test.got.z), math.Float32bits(test.got.w)}
			if got != test.want {
				t.Fatalf("got %08x, want %08x", got, test.want)
			}
		})
	}
}

func TestContactSolverWideSignedZeroAndNaN(t *testing.T) {
	negZero := math.Float32frombits(0x80000000)

	nan := math.Float32frombits(0x7fc12345)
	for _, f := range []func(floatW, floatW) floatW{minW, maxW} {
		got := f(setW(negZero, 0, nan, 1), setW(0, negZero, 2, nan))
		if math.Float32bits(got.x) != 0x80000000 || math.Float32bits(got.y) != 0 || got.z != 2 || math.Float32bits(got.w) != 0x7fc12345 {
			t.Fatalf("ties/NaN: %+v", got)
		}
	}
}

func TestContactSolverGatherScatter(t *testing.T) {
	states := []bodyState{
		{linearVelocity: Vec2{1, 2}, angularVelocity: 3, flags: dynamicFlag, deltaPosition: Vec2{4, 5}, deltaRotation: Rot{C: 6, S: 7}},
		{linearVelocity: Vec2{8, 9}, angularVelocity: 10, deltaPosition: Vec2{11, 12}, deltaRotation: Rot{C: 13, S: 14}},
		{linearVelocity: Vec2{15, 16}, angularVelocity: 17, flags: dynamicFlag},
	}
	before := append([]bodyState(nil), states...)
	indices := []int{1, 0, 2, 3}

	var wide bodyStateW
	gatherBodies(states, indices, &wide, true)

	if wide.v.x != (floatW{1, 0, 8, 15}) || wide.v.y != (floatW{2, 0, 9, 16}) || wide.w != (floatW{3, 0, 10, 17}) || wide.dp.x != (floatW{4, 0, 11, 0}) || wide.dp.y != (floatW{5, 0, 12, 0}) || wide.dq.c != (floatW{6, 1, 13, 0}) || wide.dq.s != (floatW{7, 0, 14, 0}) {
		t.Fatalf("gather: %+v", wide)
	}

	wide.v.x = floatW{21, 22, 23, 24}
	wide.v.y = floatW{31, 32, 33, 34}
	wide.w = floatW{41, 42, 43, 44}
	wide.dp = vec2W{}
	wide.dq = rotW{}
	scatterBodies(states, indices, &wide)

	before[0].linearVelocity = Vec2{21, 31}
	before[0].angularVelocity = 41
	before[2].linearVelocity = Vec2{24, 34}

	before[2].angularVelocity = 44
	for i := range states {
		if states[i] != before[i] {
			t.Fatalf("scatter[%d]: %+v want %+v", i, states[i], before[i])
		}
	}

	if getWideContactConstraintByteCount() <= 0 {
		t.Fatal("constraint byte count")
	}
}

func solverContext() (*stepContext, *contactConstraint, *contactConstraintWide) {
	w := &World{contactSpeed: 10, restitutionThreshold: 1, hitEventThreshold: 1, solverSets: make([]solverSet, 3), taskContexts: make([]taskContext, 2)}
	states := []bodyState{{linearVelocity: Vec2{2, -4}, angularVelocity: 2, flags: dynamicFlag, deltaRotation: Rot{C: 1}}}
	w.solverSets[awakeSet].bodyStates = states
	color := &w.constraintGraph.colors[overflowIndex]
	color.contactSims = make([]contactSim, 1)
	color.overflowConstraints = []contactConstraint{{indexB: 1, normal: Vec2{0, 1}, invMassB: 1, invIB: 1, rollingMass: 1, friction: 0.5, rollingResistance: 0.25, softness: softness{massScale: 1}, pointCount: 1, points: [2]contactConstraintPoint{{normalMass: 1, tangentMass: 1, relativeVelocity: -4}}}}
	wide := []contactConstraintWide{{indexB: [simdWidth]int{1}, normal: vec2W{y: floatW{x: 1}}, invMassB: floatW{x: 1}, invIB: floatW{x: 1}, normalMass1: floatW{x: 1}, tangentMass1: floatW{x: 1}, friction: floatW{x: 0.5}, rollingResistance: floatW{x: 0.25}, relativeVelocity1: floatW{x: -4}}}
	w.constraintGraph.colors[0].wideConstraints = wide

	return &stepContext{world: w, graph: &w.constraintGraph, states: states, invH: 1, contactSoftness: softness{massScale: 1}, staticSoftness: softness{massScale: 1}}, &color.overflowConstraints[0], &wide[0]
}

func TestContactSolverPhases(t *testing.T) {
	for _, wide := range []bool{false, true} {
		name := "overflow"
		if wide {
			name = "wide"
		}

		t.Run(name, func(t *testing.T) {
			ctx, c, cw := solverContext()

			block := solverBlock{count: 1}
			if wide {
				solveContactsWide(block, ctx)
			} else {
				solveContactsOverflow(ctx)
			}

			checkContactSolverSolveImpulses(t, ctx, c, cw, wide)

			c.restitution = 0.5
			cw.restitution.x = 0.5

			if wide {
				applyRestitutionWide(block, ctx)
			} else {
				applyRestitutionOverflow(ctx)
			}

			checkContactSolverBounceImpulses(t, ctx, c, cw, wide)

			if wide {
				applyRestitutionWide(block, ctx)
			} else {
				applyRestitutionOverflow(ctx)
			}

			if ctx.states[0].linearVelocity.Y != 2 {
				t.Fatal("repeated restitution gained energy")
			}
		})
	}
}

func TestContactSolverWarmStartAndPush(t *testing.T) {
	for _, wide := range []bool{false, true} {
		name := "overflow"
		if wide {
			name = "wide"
		}

		t.Run(name, func(t *testing.T) {
			testContactSolverWarmStartAndPush(t, wide)
		})
	}
}

func TestContactSolverStoreImpulses(t *testing.T) {
	ctx, c, cw := solverContext()
	for i := range ctx.world.taskContexts {
		ctx.world.taskContexts[i].hitEventBitSet = createBitSet(128)
		setBitCountAndClear(&ctx.world.taskContexts[i].hitEventBitSet, 128)
	}

	contacts := []contactSim{{contactID: 65, simFlags: simEnableHitEvent, manifold: Manifold{PointCount: 1, Points: [2]ManifoldPoint{{NormalVelocity: -2}}}}}
	ctx.graph.colors[overflowIndex].contactSims = contacts
	c.points[0].normalImpulse = 3
	c.points[0].tangentImpulse = -1
	c.points[0].totalNormalImpulse = 4
	c.rollingImpulse = 0.5

	storeImpulsesOverflow(ctx)

	storedOverflow := [4]float32{contacts[0].manifold.Points[0].NormalImpulse, contacts[0].manifold.Points[0].TangentImpulse, contacts[0].manifold.Points[0].TotalNormalImpulse, contacts[0].manifold.RollingImpulse}
	if storedOverflow != ([4]float32{3, -1, 4, 0.5}) || !ctx.world.taskContexts[0].hasHitEvents || !getBit(&ctx.world.taskContexts[0].hitEventBitSet, 65) {
		t.Fatal("overflow store/hit")
	}

	cw.normalImpulse1.x = 5
	cw.normalImpulse2.x = 6
	cw.tangentImpulse1.x = -2
	cw.tangentImpulse2.x = -3
	cw.totalNormalImpulse1.x = 7
	cw.totalNormalImpulse2.x = 8
	cw.rollingImpulse.x = 1
	ctx.wideContactConstraints = ctx.graph.colors[0].wideConstraints
	ctx.contactPrepareSpans = []contactPrepareSpan{{start: 0, count: 1, contacts: contacts}, {start: 1}}
	storeImpulsesWide(solverBlock{count: 1}, ctx, 1)

	m := contacts[0].manifold

	storedWide := [7]float32{m.Points[0].NormalImpulse, m.Points[1].NormalImpulse, m.Points[0].TangentImpulse, m.Points[1].TangentImpulse, m.Points[0].TotalNormalImpulse, m.Points[1].TotalNormalImpulse, m.RollingImpulse}
	if storedWide != ([7]float32{5, 6, -2, -3, 7, 8, 1}) || !ctx.world.taskContexts[1].hasHitEvents || !getBit(&ctx.world.taskContexts[1].hitEventBitSet, 65) {
		t.Fatal("wide store/hit")
	}
}

func TestContactSolverPrepare(t *testing.T) {
	for _, wide := range []bool{false, true} {
		name := "overflow"
		if wide {
			name = "wide"
		}

		t.Run(name, func(t *testing.T) {
			ctx, c, cw := solverContext()
			ctx.world.enableWarmStarting = true
			ctx.world.bodies = []body{{setIndex: staticSet, localIndex: 0}, {setIndex: awakeSet, localIndex: 0}}
			ctx.staticSoftness = softness{massScale: 0.5, biasRate: 3, impulseScale: 0.25}
			contacts := []contactSim{{bodyIDA: 0, bodyIDB: 1, encodedBodySimA: -2, encodedBodySimB: 0, invMassB: 1, invIB: 1, friction: 0.5, restitution: 0.5, rollingResistance: 0.25, tangentSpeed: 2, manifold: Manifold{Normal: Vec2{0, 1}, RollingImpulse: 0.5, PointCount: 1, Points: [2]ManifoldPoint{{AnchorB: Vec2{1, 0}, Separation: -0.25, NormalImpulse: 2, TangentImpulse: -1}}}}}
			ctx.graph.colors[overflowIndex].contactSims = contacts
			ctx.wideContactConstraints = ctx.graph.colors[0].wideConstraints

			ctx.contactPrepareSpans = []contactPrepareSpan{{count: 1, contacts: contacts}, {start: 1}}
			if wide {
				prepareContactsWide(solverBlock{count: 1}, ctx)
			} else {
				prepareContactsOverflow(ctx)
			}

			if ctx.anyRestitution.Load() != 1 || contacts[0].manifold.Points[0].NormalVelocity != -2 {
				t.Fatal("velocity sampling/restitution flag")
			}

			if wide {
				checkContactSolverWidePrepare(t, cw)
			} else {
				checkContactSolverOverflowPrepare(t, ctx, c)
			}
		})
	}
}

func TestContactSolverTwoPointWarmStart(t *testing.T) {
	ctx, c, _ := solverContext()
	ctx.states[0].linearVelocity = Vec2{}
	ctx.states[0].angularVelocity = 0
	c.pointCount = 2
	c.points[0].normalImpulse = 1
	c.points[1].normalImpulse = 2
	c.points[1].anchorB = Vec2{1, 0}
	c.rollingImpulse = -0.5

	warmStartContactsOverflow(ctx)

	if ctx.states[0].linearVelocity != (Vec2{0, 3}) || ctx.states[0].angularVelocity != 1.5 || c.points[0].totalNormalImpulse != 1 || c.points[1].totalNormalImpulse != 2 {
		t.Fatalf("overflow two points: %+v %+v", ctx.states[0], c)
	}

	ctx.states = []bodyState{{flags: dynamicFlag, deltaRotation: Rot{C: 1}}, {flags: dynamicFlag, deltaRotation: Rot{C: 1}}, {linearVelocity: Vec2{5, 6}, angularVelocity: 7, deltaRotation: Rot{C: 1}}}
	kin := ctx.states[2]
	cw := contactConstraintWide{indexB: [simdWidth]int{1, 2, 0, 3}, normal: vec2W{y: setW(1, 1, 0, 1)}, invMassB: setW(1, 1, 0, 0), invIB: setW(1, 1, 0, 0), anchorB2: vec2W{x: setW(1, 1, 0, 1)}, normalImpulse1: setW(1, 1, 0, 1), normalImpulse2: setW(2, 2, 0, 2), rollingImpulse: setW(-0.5, -0.5, 0, -0.5)}
	ctx.graph.colors[0].wideConstraints = []contactConstraintWide{cw}
	warmStartContactsWide(solverBlock{count: 1}, ctx)

	for i := range 2 {
		if ctx.states[i].linearVelocity != (Vec2{0, 3}) || ctx.states[i].angularVelocity != 1.5 {
			t.Fatalf("wide two points [%d]: %+v", i, ctx.states[i])
		}
	}

	if ctx.states[2] != kin {
		t.Fatal("kinematic state changed")
	}

	cw = ctx.graph.colors[0].wideConstraints[0]
	if cw.totalNormalImpulse1 != (floatW{1, 1, 0, 1}) || cw.totalNormalImpulse2 != (floatW{2, 2, 0, 2}) {
		t.Fatal("wide two point totals")
	}
}

func checkContactSolverSolveImpulses(t *testing.T, ctx *stepContext, c *contactConstraint, cw *contactConstraintWide, wide bool) {
	t.Helper()

	if ctx.states[0].linearVelocity != (Vec2{0, 0}) || ctx.states[0].angularVelocity != 1 {
		t.Fatalf("solve: %+v", ctx.states[0])
	}

	if wide {
		if cw.normalImpulse1.x != 4 || cw.totalNormalImpulse1.x != 4 || cw.tangentImpulse1.x != -2 || cw.rollingImpulse.x != -1 {
			t.Fatalf("solve impulses: %+v", cw)
		}
	} else {
		if c.points[0].normalImpulse != 4 || c.points[0].totalNormalImpulse != 4 || c.points[0].tangentImpulse != -2 || c.rollingImpulse != -1 {
			t.Fatalf("solve impulses: %+v", c)
		}
	}
}

func checkContactSolverBounceImpulses(t *testing.T, ctx *stepContext, c *contactConstraint, cw *contactConstraintWide, wide bool) {
	t.Helper()

	if ctx.states[0].linearVelocity.Y != 2 {
		t.Fatalf("bounce: %+v", ctx.states[0])
	}

	if wide {
		if cw.normalImpulse1.x != 6 || cw.totalNormalImpulse1.x != 6 || cw.restitutionImpulse1.x != 2 {
			t.Fatal("bounce impulses")
		}
	} else {
		if c.points[0].normalImpulse != 6 || c.points[0].totalNormalImpulse != 6 || c.points[0].restitutionImpulse != 2 {
			t.Fatal("bounce impulses")
		}
	}
}

func testContactSolverWarmStartAndPush(t *testing.T, wide bool) {
	t.Helper()

	ctx, c, cw := solverContext()
	block := solverBlock{count: 1}
	c.points[0].normalImpulse = 2
	c.points[0].tangentImpulse = -1
	c.rollingImpulse = -0.5
	cw.normalImpulse1.x = 2
	cw.tangentImpulse1.x = -1
	cw.rollingImpulse.x = -0.5

	if wide {
		warmStartContactsWide(block, ctx)
	} else {
		warmStartContactsOverflow(ctx)
	}

	if ctx.states[0].linearVelocity != (Vec2{1, -2}) || ctx.states[0].angularVelocity != 1.5 {
		t.Fatalf("warm start: %+v", ctx.states[0])
	}

	c.points[0].baseSeparation = 1
	cw.baseSeparation1.x = 1

	if wide {
		pushContactsWide(block, ctx)
	} else {
		pushContactsOverflow(ctx)
	}

	if ctx.states[0].linearVelocity != (Vec2{1, -1}) || ctx.states[0].angularVelocity != 1.5 {
		t.Fatalf("speculative push: %+v", ctx.states[0])
	}

	c.points[0].baseSeparation = -2
	cw.baseSeparation1.x = -2
	c.softness = softness{massScale: 0.5, biasRate: 2, impulseScale: 0.25}

	ctx.staticSoftness = c.softness
	if wide {
		pushContactsWide(block, ctx)
	} else {
		pushContactsOverflow(ctx)
	}

	if ctx.states[0].linearVelocity != (Vec2{1, 0.75}) || ctx.states[0].angularVelocity != 1.5 {
		t.Fatalf("soft push: %+v", ctx.states[0])
	}
}

func checkContactSolverWidePrepare(t *testing.T, cw *contactConstraintWide) {
	t.Helper()

	if cw.indexA != ([simdWidth]int{}) || cw.indexB != ([simdWidth]int{1}) || cw.normalMass1.x != 0.5 || cw.tangentMass1.x != 1 || cw.normalMass2 != (floatW{}) || cw.tangentMass2 != (floatW{}) || cw.baseSeparation1.x != -0.25 || cw.normalImpulse1.x != 2 || cw.tangentImpulse1.x != -1 || cw.relativeVelocity1.x != -2 || cw.rollingImpulse.x != 0.5 {
		t.Fatalf("wide prepare: %+v", cw)
	}
}

func checkContactSolverOverflowPrepare(t *testing.T, ctx *stepContext, c *contactConstraint) {
	t.Helper()

	if c.indexA != 0 || c.indexB != 1 || c.points[0].normalMass != 0.5 || c.points[0].tangentMass != 1 || c.points[0].baseSeparation != -0.25 || c.points[0].normalImpulse != 2 || c.points[0].tangentImpulse != -1 || c.points[0].relativeVelocity != -2 || c.rollingMass != 1 || c.rollingImpulse != 0.5 || c.softness != ctx.staticSoftness {
		t.Fatalf("overflow prepare: %+v", c)
	}
}
