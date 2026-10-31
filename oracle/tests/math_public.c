// SPDX-License-Identifier: MIT

#include "box2d/collision.h"
#include "box2d/constants.h"
#include "box2d/math_functions.h"
#include "oracle.h"

#include <limits.h>
#include <math.h>
#include <stdint.h>
#include <string.h>

static float math_public_float(uint32_t bits) {
    float value;

    memcpy(&value, &bits, sizeof(value));

    return value;
}

static void math_public_transform(const char* key, b2Transform value) {
    ORACLE_VEC2(ORACLE_KEY("%s/p", key), value.p);
    ORACLE_ROT(ORACLE_KEY("%s/q", key), value.q);
}

static void math_public_world_transform(const char* key, b2WorldTransform value) {
    ORACLE_POS(ORACLE_KEY("%s/p", key), value.p);
    ORACLE_ROT(ORACLE_KEY("%s/q", key), value.q);
}

static void math_public_scalars(void) {
    ORACLE_OPEN("math/PublicScalars");

    const int integers[] = {0, 1, -1, 17, -17, INT_MAX, INT_MIN + 1};
    const int count      = sizeof(integers) / sizeof(integers[0]);

    ORACLE_I32("integers/count", count);

    for (int i = 0; i < count; ++i) {
        int a = integers[i], b = integers[count - 1 - i];

        ORACLE_I32(ORACLE_KEY("integers/%d/min", i), b2MinInt(a, b));
        ORACLE_I32(ORACLE_KEY("integers/%d/max", i), b2MaxInt(a, b));
        ORACLE_I32(ORACLE_KEY("integers/%d/abs", i), b2AbsInt(a));
        ORACLE_I32(ORACLE_KEY("integers/%d/clamp", i), b2ClampInt(a, -1, 1));
    }

    const int divisions[][2] = {{0, 3}, {1, 3}, {3, 3}, {4, 3}, {17, 5}, {INT_MAX - 2, 2}};

    ORACLE_I32("ceiling/count", sizeof(divisions) / sizeof(divisions[0]));

    for (int i = 0; i < (int)(sizeof(divisions) / sizeof(divisions[0])); ++i) {
        ORACLE_I32(ORACLE_KEY("ceiling/%d", i), b2CeilingInt(divisions[i][0], divisions[i][1]));
    }

    const float floats[]   = {0.0f, -0.0f, -3.25f, 3.25f, -2.0f, 2.0f, FLT_MIN, 0x1p-149f, FLT_MAX};
    const int   floatCount = sizeof(floats) / sizeof(floats[0]);

    ORACLE_I32("floats/count", floatCount);

    for (int i = 0; i < floatCount; ++i) {
        float a = floats[i], b = floats[floatCount - 1 - i];

        ORACLE_F32(ORACLE_KEY("floats/%d/min", i), b2MinFloat(a, b));
        ORACLE_F32(ORACLE_KEY("floats/%d/max", i), b2MaxFloat(a, b));
        ORACLE_F32(ORACLE_KEY("floats/%d/abs", i), b2AbsFloat(a));
        ORACLE_F32(ORACLE_KEY("floats/%d/clamp", i), b2ClampFloat(a, -2.0f, 2.0f));
    }

    ORACLE_F32("zero/min-negative-first", b2MinFloat(-0.0f, 0.0f));
    ORACLE_F32("zero/min-positive-first", b2MinFloat(0.0f, -0.0f));
    ORACLE_F32("zero/max-negative-first", b2MaxFloat(-0.0f, 0.0f));
    ORACLE_F32("zero/max-positive-first", b2MaxFloat(0.0f, -0.0f));
    ORACLE_CLOSE();
}

static void math_public_vectors(void) {
    ORACLE_OPEN("math/PublicVectors");

    const struct {
        b2Vec2 a, b;
        float  s, t;
    } cases[] = {
        {{2.5f, -3.75f}, {-4.0f, 0.125f}, -1.25f, 0.3f},
        {{0.0f, -0.0f}, {-0.0f, 0.0f}, -0.0f, 0.0f},
        {{1.0f, 0.0f}, {0.0f, 1.0f}, 2.0f, 1.0f},
        {{-1.0f, -1.0f}, {-1.0f, -1.0f}, 0.0f, 0.5f},
        {{0x1p-60f, -0x1p-60f}, {0x1p-59f, 0x1p-60f}, 0.5f, -0.25f},
        {{1024.125f, -2048.5f}, {-16.75f, 33.25f}, 0.125f, 1.25f},
    };

    const int count = sizeof(cases) / sizeof(cases[0]);

    ORACLE_I32("count", count);

    for (int i = 0; i < count; ++i) {
        b2Vec2 a = cases[i].a, b = cases[i].b;
        float  s = cases[i].s, t = cases[i].t;

        ORACLE_F32(ORACLE_KEY("%d/dot", i), b2Dot(a, b));
        ORACLE_F32(ORACLE_KEY("%d/cross", i), b2Cross(a, b));
        ORACLE_VEC2(ORACLE_KEY("%d/cross-vs", i), b2CrossVS(a, s));
        ORACLE_VEC2(ORACLE_KEY("%d/cross-sv", i), b2CrossSV(s, a));
        ORACLE_VEC2(ORACLE_KEY("%d/left", i), b2LeftPerp(a));
        ORACLE_VEC2(ORACLE_KEY("%d/right", i), b2RightPerp(a));
        ORACLE_VEC2(ORACLE_KEY("%d/add", i), b2Add(a, b));
        ORACLE_VEC2(ORACLE_KEY("%d/sub", i), b2Sub(a, b));
        ORACLE_VEC2(ORACLE_KEY("%d/neg", i), b2Neg(a));
        ORACLE_VEC2(ORACLE_KEY("%d/lerp", i), b2Lerp(a, b, t));
        ORACLE_VEC2(ORACLE_KEY("%d/mul", i), b2Mul(a, b));
        ORACLE_VEC2(ORACLE_KEY("%d/mul-sv", i), b2MulSV(s, a));
        ORACLE_VEC2(ORACLE_KEY("%d/mul-add", i), b2MulAdd(a, s, b));
        ORACLE_VEC2(ORACLE_KEY("%d/mul-sub", i), b2MulSub(a, s, b));
        ORACLE_VEC2(ORACLE_KEY("%d/abs", i), b2Abs(a));
        ORACLE_VEC2(ORACLE_KEY("%d/min", i), b2Min(a, b));
        ORACLE_VEC2(ORACLE_KEY("%d/max", i), b2Max(a, b));
        ORACLE_VEC2(
            ORACLE_KEY("%d/clamp", i),
            b2Clamp(a, (b2Vec2){-1.0f, -2.0f}, (b2Vec2){1.0f, 2.0f})
        );
        ORACLE_F32(ORACLE_KEY("%d/length", i), b2Length(a));
        ORACLE_F32(ORACLE_KEY("%d/length-squared", i), b2LengthSquared(a));
        ORACLE_F32(ORACLE_KEY("%d/distance", i), b2Distance(a, b));
        ORACLE_F32(ORACLE_KEY("%d/distance-squared", i), b2DistanceSquared(a, b));
    }
    // Adjacent float inputs straddle sqrt(1000 * FLT_MIN). Expected outputs
    // come from the pinned routines, not a second normalization implementation.

    const b2Vec2 normals[] = {
        {0.0f, 0.0f},
        {-0.0f, -0.0f},
        {0x1p-149f, 0.0f},
        {FLT_MIN, -FLT_MIN},
        {math_public_float(0x227cfb71), 0.0f},
        {math_public_float(0x227cfb72), 0.0f},
        {math_public_float(0x227cfb73), 0.0f},
        {0.0f, -math_public_float(0x227cfb73)},
        {3.0f, -4.0f},
        {1.0f, 0.0f},
        {math_public_float(0x3f800031), 0.0f},
        {math_public_float(0x3f800032), 0.0f},
        {math_public_float(0x3f800033), 0.0f},
    };

    ORACLE_I32("normalize/count", sizeof(normals) / sizeof(normals[0]));

    for (int i = 0; i < (int)(sizeof(normals) / sizeof(normals[0])); ++i) {
        float length = -123.0f;

        ORACLE_VEC2(ORACLE_KEY("normalize/%d/value", i), b2Normalize(normals[i]));

        b2Vec2 value = b2GetLengthAndNormalize(&length, normals[i]);

        ORACLE_VEC2(ORACLE_KEY("normalize/%d/with-length", i), value);
        ORACLE_F32(ORACLE_KEY("normalize/%d/length", i), length);
        ORACLE_BOOL(ORACLE_KEY("normalize/%d/unit", i), b2IsNormalized(normals[i]));
    }

    ORACLE_CLOSE();
}

static void math_public_rotations(void) {
    ORACLE_OPEN("math/PublicRotations");

    const struct {
        float a, b, delta, t, invH;
    } cases[] = {
        {0.0f, 0.0f, 0.0f, 0.0f, 60.0f},
        {0.25f, -0.7f, 0.03f, 0.3f, 60.0f},
        {-2.5f, 2.75f, -0.125f, 1.0f, 120.0f},
        {0.0f, B2_PI, 0.5f, 0.5f, 0.0f},
        {-B2_PI, 0.5f * B2_PI, -0.5f, 0.0f, 1.0f},
    };

    const int count = sizeof(cases) / sizeof(cases[0]);

    ORACLE_I32("count", count);

    for (int i = 0; i < count; ++i) {
        b2Rot a = b2MakeRot(cases[i].a), b = b2MakeRot(cases[i].b);

        ORACLE_ROT(ORACLE_KEY("%d/a", i), a);
        ORACLE_ROT(ORACLE_KEY("%d/b", i), b);
        ORACLE_ROT(ORACLE_KEY("%d/integrate", i), b2IntegrateRotation(a, cases[i].delta));
        ORACLE_ROT(ORACLE_KEY("%d/from-vector", i), b2MakeRotFromUnitVector(b2Rot_GetXAxis(a)));
        ORACLE_ROT(
            ORACLE_KEY("%d/between-vectors", i),
            b2ComputeRotationBetweenUnitVectors(b2Rot_GetXAxis(a), b2Rot_GetXAxis(b))
        );
        ORACLE_ROT(ORACLE_KEY("%d/inverse", i), b2InvertRot(a));
        ORACLE_ROT(ORACLE_KEY("%d/nlerp", i), b2NLerp(a, b, cases[i].t));
        ORACLE_F32(
            ORACLE_KEY("%d/angular-velocity", i),
            b2ComputeAngularVelocity(a, b, cases[i].invH)
        );
        ORACLE_F32(ORACLE_KEY("%d/angle", i), b2Rot_GetAngle(a));
        ORACLE_VEC2(ORACLE_KEY("%d/x-axis", i), b2Rot_GetXAxis(a));
        ORACLE_VEC2(ORACLE_KEY("%d/y-axis", i), b2Rot_GetYAxis(a));
        ORACLE_ROT(ORACLE_KEY("%d/mul", i), b2MulRot(a, b));
        ORACLE_ROT(ORACLE_KEY("%d/inv-mul", i), b2InvMulRot(a, b));
        ORACLE_F32(ORACLE_KEY("%d/relative-angle", i), b2RelativeAngle(a, b));
        ORACLE_VEC2(ORACLE_KEY("%d/rotate", i), b2RotateVector(a, (b2Vec2){2.25f, -3.5f}));
        ORACLE_VEC2(ORACLE_KEY("%d/inv-rotate", i), b2InvRotateVector(a, (b2Vec2){2.25f, -3.5f}));
    }

    const b2Rot raw[] =
        {{0.0f, 0.0f}, {-0.0f, -0.0f}, {3.0f, -4.0f}, {1.0f, 0.0f}, {0x1p-60f, -0x1p-60f}};

    ORACLE_I32("normalize/count", sizeof(raw) / sizeof(raw[0]));

    for (int i = 0; i < (int)(sizeof(raw) / sizeof(raw[0])); ++i) {
        ORACLE_ROT(ORACLE_KEY("normalize/%d", i), b2NormalizeRot(raw[i]));
    }

    ORACLE_ROT("nlerp-opposite", b2NLerp((b2Rot){1.0f, 0.0f}, (b2Rot){-1.0f, 0.0f}, 0.5f));

    const b2Vec2 directions[] = {
        {0.0f, 0.0f},
        {-0.0f, -0.0f},
        {1.0f, 0.0f},
        {-1.0f, 0.0f},
        {0.0f, 1.0f},
        {0.0f, -1.0f},
        {1.0f, 1.0f},
        {-1.0f, 1.0f},
        {-1.0f, -1.0f},
        {1.0f, -1.0f},
        {0x1p-149f, 0x1p-148f}
    };

    ORACLE_I32("atan2/count", sizeof(directions) / sizeof(directions[0]));

    for (int i = 0; i < (int)(sizeof(directions) / sizeof(directions[0])); ++i) {
        ORACLE_F32(ORACLE_KEY("atan2/%d", i), b2Atan2(directions[i].y, directions[i].x));
    }

    ORACLE_CLOSE();
}

static void math_public_matrix_spring(void) {
    ORACLE_OPEN("math/PublicMatrixSpring");

    const b2Mat22 matrices[] = {
        {{1.0f, 0.0f}, {0.0f, 1.0f}},
        {{3.0f, -2.0f}, {1.25f, 4.0f}},
        {{0.0f, 0.0f}, {0.0f, 0.0f}},
        {{1.0f, 2.0f}, {2.0f, 4.0f}},
        {{-2.0f, 0.0f}, {0.0f, 0.5f}},
        {{0x1p-10f, 1.0f}, {0.0f, 0x1p-10f}},
    };

    ORACLE_I32("matrix/count", sizeof(matrices) / sizeof(matrices[0]));

    for (int i = 0; i < (int)(sizeof(matrices) / sizeof(matrices[0])); ++i) {
        b2Vec2 rhs = {2.5f, -3.75f};

        ORACLE_VEC2(ORACLE_KEY("matrix/%d/mul", i), b2MulMV(matrices[i], rhs));

        b2Mat22 inverse = b2GetInverse22(matrices[i]);

        ORACLE_VEC2(ORACLE_KEY("matrix/%d/inverse-cx", i), inverse.cx);
        ORACLE_VEC2(ORACLE_KEY("matrix/%d/inverse-cy", i), inverse.cy);
        ORACLE_VEC2(ORACLE_KEY("matrix/%d/solve", i), b2Solve22(matrices[i], rhs));
    }

    const float springs[][5] = {
        {0.0f, 0.0f, 2.0f, -3.0f, 0.016f},
        {2.0f, 0.0f, 2.0f, -3.0f, 0.016f},
        {2.0f, 0.7f, -1.25f, 4.0f, 0.016f},
        {10.0f, 1.0f, 3.0f, 0.0f, 1.0f},
        {2.0f, 2.0f, 0.0f, -3.0f, 0.0f},
        {4.0f, 1.0f, -0.0f, 0.0f, 0.125f}
    };

    ORACLE_I32("spring/count", sizeof(springs) / sizeof(springs[0]));

    for (int i = 0; i < (int)(sizeof(springs) / sizeof(springs[0])); ++i) {
        ORACLE_F32(
            ORACLE_KEY("spring/%d", i),
            b2SpringDamper(
                springs[i][0],
                springs[i][1],
                springs[i][2],
                springs[i][3],
                springs[i][4]
            )
        );
    }

    ORACLE_CLOSE();
}

static void math_public_transforms(void) {
    ORACLE_OPEN("math/PublicTransforms");

    const struct {
        b2Vec2 a, b, point;
        float  angleA, angleB;
        double worldX, worldY;
    } cases[] = {
        {{0.0f, 0.0f}, {0.0f, 0.0f}, {0.0f, -0.0f}, 0.0f, 0.0f, 0.0, -0.0},
        {{2.5f, -3.75f}, {-0.25f, 4.125f}, {1.25f, -2.0f}, 0.25f, -0.7f, 16.125, -32.25},
        {{-100.0f, 20.0f},
         {13.0f, -7.0f},
         {-3.5f, 8.25f},
         -2.5f,
         2.75f,
         10000000.125,
         -10000000.25},
        {{1024.125f, -2048.5f},
         {-1.25f, 3.5f},
         {0.125f, -0.25f},
         1.5f,
         -0.5f,
         1000000000.125,
         -1000000000.25},
    };

    const int count = sizeof(cases) / sizeof(cases[0]);

    ORACLE_I32("count", count);

    for (int i = 0; i < count; ++i) {
        b2Transform a = {cases[i].a, b2MakeRot(cases[i].angleA)};
        b2Transform b = {cases[i].b, b2MakeRot(cases[i].angleB)};
        b2Vec2      p = cases[i].point;

        ORACLE_VEC2(ORACLE_KEY("%d/point", i), b2TransformPoint(a, p));
        ORACLE_VEC2(ORACLE_KEY("%d/inv-point", i), b2InvTransformPoint(a, p));
        math_public_transform(ORACLE_KEY("%d/mul", i), b2MulTransforms(a, b));
        math_public_transform(ORACLE_KEY("%d/inv-mul", i), b2InvMulTransforms(a, b));

        b2Pos            world = oracle_pos_xy(cases[i].worldX, cases[i].worldY);
        b2WorldTransform wa    = {world, a.q};
        b2WorldTransform wb    = {b2OffsetPos(world, b.p), b.q};

        ORACLE_POS(ORACLE_KEY("%d/to-pos", i), b2ToPos(p));
        ORACLE_VEC2(ORACLE_KEY("%d/to-vec", i), b2ToVec2(world));
        ORACLE_POS(ORACLE_KEY("%d/offset", i), b2OffsetPos(world, p));
        ORACLE_VEC2(ORACLE_KEY("%d/sub-pos", i), b2SubPos(wb.p, wa.p));
        ORACLE_POS(ORACLE_KEY("%d/world-point", i), b2TransformWorldPoint(wa, p));
        ORACLE_VEC2(ORACLE_KEY("%d/inv-world-point", i), b2InvTransformWorldPoint(wa, wb.p));
        math_public_transform(ORACLE_KEY("%d/inv-world-mul", i), b2InvMulWorldTransforms(wa, wb));
        math_public_world_transform(ORACLE_KEY("%d/world-mul", i), b2MulWorldTransforms(wa, b));
        math_public_transform(ORACLE_KEY("%d/relative", i), b2ToRelativeTransform(wb, world));
        math_public_world_transform(ORACLE_KEY("%d/promote", i), b2MakeWorldTransform(a));

        const float samples[] = {0.0f, 0.3f, 0.5f, 1.0f};

        for (int j = 0; j < 4; ++j) {
            ORACLE_POS(
                ORACLE_KEY("%d/interpolate/%d", i, j),
                b2LerpPosition(wa.p, wb.p, samples[j])
            );
        }
    }

    ORACLE_CLOSE();
}

static void math_public_rounding(void) {
    ORACLE_OPEN("math/PublicRounding");

    const double values[] = {
        0.0,
        -0.0,
        0x1p-150,
        -0x1p-150,
        0x1p-149,
        -0x1p-149,
        1.0,
        1.0 + 0x1p-25,
        1.0 - 0x1p-25,
        -1.0,
        -1.0 + 0x1p-25,
        -1.0 - 0x1p-25,
        1.0 + 0x1p-24,
        -1.0 - 0x1p-24,
        10000000.125,
        -10000000.125,
        1000000000.25,
        -1000000000.25,
        (double)FLT_MAX,
        -(double)FLT_MAX,
        (double)math_public_float(0x7f7ffffe),
        -(double)math_public_float(0x7f7ffffe),
        nextafter((double)FLT_MAX, 0.0),
        nextafter(-(double)FLT_MAX, 0.0)
    };

    ORACLE_I32("count", sizeof(values) / sizeof(values[0]));

    for (int i = 0; i < (int)(sizeof(values) / sizeof(values[0])); ++i) {
        ORACLE_F32(ORACLE_KEY("%d/down", i), b2RoundDownFloat(values[i]));
        ORACLE_F32(ORACLE_KEY("%d/up", i), b2RoundUpFloat(values[i]));
    }

    ORACLE_CLOSE();
}

static void math_public_validity(void) {
    ORACLE_OPEN("math/PublicValidity");

    const float values[] = {
        0.0f,
        -0.0f,
        1.0f,
        -1.0f,
        FLT_MIN,
        FLT_MAX,
        math_public_float(0x7fc00001),
        INFINITY,
        -INFINITY
    };
    const int count = sizeof(values) / sizeof(values[0]);

    ORACLE_I32("numbers/count", count);

    for (int i = 0; i < count; ++i) {
        float x = values[i];

        ORACLE_BOOL(ORACLE_KEY("numbers/%d/float", i), b2IsValidFloat(x));
        ORACLE_BOOL(ORACLE_KEY("numbers/%d/vector-x", i), b2IsValidVec2((b2Vec2){x, 0.0f}));
        ORACLE_BOOL(ORACLE_KEY("numbers/%d/vector-y", i), b2IsValidVec2((b2Vec2){0.0f, x}));
        ORACLE_BOOL(ORACLE_KEY("numbers/%d/position-x", i), b2IsValidPosition((b2Pos){x, 0.0f}));
        ORACLE_BOOL(ORACLE_KEY("numbers/%d/position-y", i), b2IsValidPosition((b2Pos){0.0f, x}));
        ORACLE_BOOL(
            ORACLE_KEY("numbers/%d/transform", i),
            b2IsValidTransform((b2Transform){{x, 0.0f}, {1.0f, 0.0f}})
        );
        ORACLE_BOOL(
            ORACLE_KEY("numbers/%d/world-transform", i),
            b2IsValidWorldTransform((b2WorldTransform){{0.0f, x}, {1.0f, 0.0f}})
        );
        ORACLE_BOOL(
            ORACLE_KEY("numbers/%d/plane-offset", i),
            b2IsValidPlane((b2Plane){{1.0f, 0.0f}, x})
        );
    }

    const double worldValues[] = {
        0.0,
        -0.0,
        FLT_MAX,
        -FLT_MAX,
        NAN,
        INFINITY,
        -INFINITY,
#if defined(BOX2D_DOUBLE_PRECISION)
        1.0e100,
        -1.0e100,
        DBL_MAX,
        -DBL_MAX,
#endif
    };

    ORACLE_I32("world-numbers/count", sizeof(worldValues) / sizeof(worldValues[0]));

    for (int i = 0; i < (int)(sizeof(worldValues) / sizeof(worldValues[0])); ++i) {
        b2Pos px = oracle_pos_xy(worldValues[i], 0.0), py = oracle_pos_xy(0.0, worldValues[i]);

        ORACLE_BOOL(ORACLE_KEY("world-numbers/%d/x", i), b2IsValidPosition(px));
        ORACLE_BOOL(ORACLE_KEY("world-numbers/%d/y", i), b2IsValidPosition(py));
        ORACLE_BOOL(
            ORACLE_KEY("world-numbers/%d/transform", i),
            b2IsValidWorldTransform((b2WorldTransform){px, {1.0f, 0.0f}})
        );
    }

    const b2Rot rotations[] = {
        {1.0f, 0.0f},
        {0.0f, 1.0f},
        {-1.0f, -0.0f},
        {0.0f, 0.0f},
        {2.0f, 0.0f},
        {0.9997f, 0.0f},
        {1.0003f, 0.0f},
        {math_public_float(0x7fc00001), 0.0f},
        {1.0f, math_public_float(0x7fc00001)},
        {INFINITY, 0.0f},
        {1.0f, -INFINITY}
    };

    ORACLE_I32("rotation/count", sizeof(rotations) / sizeof(rotations[0]));

    for (int i = 0; i < (int)(sizeof(rotations) / sizeof(rotations[0])); ++i) {
        b2Rot q = rotations[i];

        ORACLE_BOOL(ORACLE_KEY("rotation/%d/normalized", i), b2IsNormalizedRot(q));
        ORACLE_BOOL(ORACLE_KEY("rotation/%d/valid", i), b2IsValidRotation(q));
        ORACLE_BOOL(
            ORACLE_KEY("rotation/%d/transform", i),
            b2IsValidTransform((b2Transform){{0.0f, 0.0f}, q})
        );
        ORACLE_BOOL(
            ORACLE_KEY("rotation/%d/world-transform", i),
            b2IsValidWorldTransform((b2WorldTransform){{0.0f, 0.0f}, q})
        );
        ORACLE_BOOL(
            ORACLE_KEY("rotation/%d/plane-normal", i),
            b2IsValidPlane((b2Plane){{q.c, q.s}, 0.0f})
        );
    }

    const b2AABB boxes[] = {
        {{0.0f, 0.0f}, {0.0f, 0.0f}},
        {{-0.0f, -0.0f}, {0.0f, 0.0f}},
        {{-1.0f, -2.0f}, {3.0f, 4.0f}},
        {{1.0f, 0.0f}, {0.0f, 1.0f}},
        {{0.0f, 2.0f}, {1.0f, 1.0f}},
        {{-INFINITY, 0.0f}, {1.0f, 1.0f}},
        {{0.0f, 0.0f}, {1.0f, INFINITY}},
        {{0.0f, 0.0f}, {math_public_float(0x7fc00001), 1.0f}}
    };

    ORACLE_I32("aabb/count", sizeof(boxes) / sizeof(boxes[0]));

    for (int i = 0; i < (int)(sizeof(boxes) / sizeof(boxes[0])); ++i) {
        ORACLE_BOOL(ORACLE_KEY("aabb/%d", i), b2IsValidAABB(boxes[i]));
    }
    // IsValidRay's maximum depends on precision and process-global length scale.
    // This executable owns its process; restore the old value before returning.

    float       oldScale = b2GetLengthUnitsPerMeter();
    const float scales[] = {1.0f, 32.0f};

    for (int scaleIndex = 0; scaleIndex < 2; ++scaleIndex) {
        b2SetLengthUnitsPerMeter(scales[scaleIndex]);
        ORACLE_F32(ORACLE_KEY("ray/%d/scale", scaleIndex), b2GetLengthUnitsPerMeter());
        ORACLE_F32(ORACLE_KEY("ray/%d/huge", scaleIndex), B2_HUGE);
        ORACLE_F32(ORACLE_KEY("ray/%d/linear-slop", scaleIndex), B2_LINEAR_SLOP);
        ORACLE_F32(ORACLE_KEY("ray/%d/speculative-distance", scaleIndex), B2_SPECULATIVE_DISTANCE);
        ORACLE_F32(
            ORACLE_KEY("ray/%d/contact-recycle-distance", scaleIndex),
            B2_CONTACT_RECYCLE_DISTANCE
        );
        ORACLE_F32(ORACLE_KEY("ray/%d/max-aabb-margin", scaleIndex), B2_MAX_AABB_MARGIN);

        const b2RayCastInput rays[] = {
            {{0.0f, 0.0f}, {0.0f, 0.0f}, 0.0f},
            {{-0.0f, 0.0f}, {1.0f, -0.0f}, -0.0f},
            {{1.0f, 2.0f}, {-3.0f, 4.0f}, 1.0f},
            {{0.0f, 0.0f}, {1.0f, 0.0f}, -0x1p-149f},
            {{0.0f, 0.0f}, {1.0f, 0.0f}, nextafterf(B2_HUGE, 0.0f)},
            {{0.0f, 0.0f}, {1.0f, 0.0f}, B2_HUGE},
            {{0.0f, 0.0f}, {1.0f, 0.0f}, nextafterf(B2_HUGE, INFINITY)},
            {{INFINITY, 0.0f}, {1.0f, 0.0f}, 1.0f},
            {{0.0f, 0.0f}, {0.0f, math_public_float(0x7fc00001)}, 1.0f},
            {{0.0f, 0.0f}, {1.0f, 0.0f}, INFINITY},
            {{0.0f, 0.0f}, {1.0f, 0.0f}, math_public_float(0x7fc00001)},
        };

        ORACLE_I32(ORACLE_KEY("ray/%d/count", scaleIndex), sizeof(rays) / sizeof(rays[0]));

        for (int i = 0; i < (int)(sizeof(rays) / sizeof(rays[0])); ++i) {
            ORACLE_BOOL(ORACLE_KEY("ray/%d/%d", scaleIndex, i), b2IsValidRay(&rays[i]));
        }
    }

    b2SetLengthUnitsPerMeter(oldScale);
    ORACLE_CLOSE();
}

static void math_public_bounds_planes(void) {
    ORACLE_OPEN("math/PublicBoundsPlanes");

    const struct {
        b2AABB a, b;
    } cases[] = {
        {{{-2.0f, -3.0f}, {4.0f, 5.0f}}, {{-1.0f, -2.0f}, {3.0f, 4.0f}}},
        {{{0.0f, 0.0f}, {1.0f, 1.0f}}, {{1.0f, 1.0f}, {2.0f, 3.0f}}},
        {{{0.0f, 0.0f}, {1.0f, 1.0f}}, {{2.0f, -1.0f}, {3.0f, 2.0f}}},
        {{{0.0f, 0.0f}, {0.0f, 0.0f}}, {{-0.0f, -0.0f}, {0.0f, 0.0f}}},
        {{{1024.125f, -2048.5f}, {1025.25f, -2046.25f}},
         {{1023.5f, -2048.0f}, {1026.0f, -2047.0f}}},
    };

    ORACLE_I32("bounds/count", sizeof(cases) / sizeof(cases[0]));

    for (int i = 0; i < (int)(sizeof(cases) / sizeof(cases[0])); ++i) {
        b2AABB a = cases[i].a, b = cases[i].b;

        ORACLE_VEC2(ORACLE_KEY("bounds/%d/center", i), b2AABB_Center(a));
        ORACLE_VEC2(ORACLE_KEY("bounds/%d/extents", i), b2AABB_Extents(a));

        b2AABB combined = b2AABB_Union(a, b);

        ORACLE_VEC2(ORACLE_KEY("bounds/%d/union-lower", i), combined.lowerBound);
        ORACLE_VEC2(ORACLE_KEY("bounds/%d/union-upper", i), combined.upperBound);
        ORACLE_BOOL(ORACLE_KEY("bounds/%d/contains", i), b2AABB_Contains(a, b));
        ORACLE_BOOL(ORACLE_KEY("bounds/%d/contained", i), b2AABB_Contains(b, a));
        ORACLE_BOOL(ORACLE_KEY("bounds/%d/overlaps", i), b2AABB_Overlaps(a, b));
    }

    const b2Vec2 points[] = {{2.5f, -3.75f}, {-4.0f, 0.125f}, {1.0f, 6.25f}};

    const struct {
        int   count;
        float radius;
    } clouds[] = {{1, 0.0f}, {3, 0.0f}, {3, 0.125f}};

    ORACLE_I32("cloud/count", sizeof(clouds) / sizeof(clouds[0]));

    for (int i = 0; i < (int)(sizeof(clouds) / sizeof(clouds[0])); ++i) {
        b2AABB bounds = b2MakeAABB(points, clouds[i].count, clouds[i].radius);

        ORACLE_VEC2(ORACLE_KEY("cloud/%d/lower", i), bounds.lowerBound);
        ORACLE_VEC2(ORACLE_KEY("cloud/%d/upper", i), bounds.upperBound);
    }

    const struct {
        b2Plane plane;
        b2Vec2  point;
    } planes[] = {
        {{{1.0f, 0.0f}, 2.0f}, {2.0f, -3.0f}},
        {{{1.0f, 0.0f}, 2.0f}, {3.0f, -3.0f}},
        {{{0.0f, -1.0f}, -2.0f}, {3.0f, 4.0f}},
        {{{0.6f, 0.8f}, 0.25f}, {-1.25f, 2.5f}},
        {{{-1.0f, -0.0f}, -0.0f}, {-0.0f, 0.0f}},
    };

    ORACLE_I32("plane/count", sizeof(planes) / sizeof(planes[0]));

    for (int i = 0; i < (int)(sizeof(planes) / sizeof(planes[0])); ++i) {
        ORACLE_F32(
            ORACLE_KEY("plane/%d/separation", i),
            b2PlaneSeparation(planes[i].plane, planes[i].point)
        );
    }

    ORACLE_CLOSE();
}

int main(void) {
    math_public_scalars();
    math_public_vectors();
    math_public_rotations();
    math_public_matrix_spring();
    math_public_transforms();
    math_public_rounding();
    math_public_validity();
    math_public_bounds_planes();

    return 0;
}
