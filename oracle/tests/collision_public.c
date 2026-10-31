// SPDX-License-Identifier: MIT
#include "box2d/collision.h"
#include "box2d/math_functions.h"
#include "oracle.h"

#include <string.h>

static void emit_manifold(const char* inputKey, b2LocalManifold m) {
    char key[256];

    oracle_copy_key(key, sizeof(key), inputKey);
    ORACLE_I32(ORACLE_KEY("%s/count", key), m.pointCount);
    // A normal and point slots are defined only for a nonempty manifold.

    if (m.pointCount > 0) { ORACLE_VEC2(ORACLE_KEY("%s/normal", key), m.normal); }
    for (int i = 0; i < m.pointCount; ++i) {
        ORACLE_VEC2(ORACLE_KEY("%s/%d/point", key, i), m.points[i].point);
        ORACLE_F32(ORACLE_KEY("%s/%d/separation", key, i), m.points[i].separation);
        ORACLE_U32(ORACLE_KEY("%s/%d/id", key, i), m.points[i].id);
    }
}

static void emit_cache(const char* inputKey, const b2SimplexCache* cache) {
    char key[256];

    oracle_copy_key(key, sizeof(key), inputKey);
    ORACLE_U32(ORACLE_KEY("%s/count", key), cache->count);

    for (int i = 0; i < cache->count; ++i) {
        ORACLE_U32(ORACLE_KEY("%s/%d/a", key, i), cache->indexA[i]);
        ORACLE_U32(ORACLE_KEY("%s/%d/b", key, i), cache->indexB[i]);
    }
}

int main(void) {
    ORACLE_OPEN("collision/PublicPairs");

    b2Circle  circle  = {{0.0f, 0.0f}, 0.5f};
    b2Capsule capsule = {{-0.75f, 0.0f}, {0.75f, 0.0f}, 0.25f};
    b2Segment segment = {{-1.5f, 0.0f}, {1.5f, 0.0f}};
    b2Polygon polygon = b2MakeRoundedBox(0.75f, 0.5f, 0.1f);
    // Straight, convex and concave ghost joins. The right side of the horizontal segment is below
    // it.
    b2ChainSegment chains[] = {
        {{-3.0f, 0.0f}, {{-1.5f, 0.0f}, {1.5f, 0.0f}}, {3.0f, 0.0f}, 0},
        {{-2.5f, -1.0f}, {{-1.5f, 0.0f}, {1.5f, 0.0f}}, {2.5f, -1.0f}, 0},
        {{-2.5f, 1.0f}, {{-1.5f, 0.0f}, {1.5f, 0.0f}}, {2.5f, 1.0f}, 0},
    };
    const b2Vec2 positions[] = {
        {0, 3},         {0, 1.21f},     {0, 1.2f},       {0, 1.01f},    {0, 1},
        {0, 0.86f},     {0, 0.85f},     {0, 0.76f},      {0, 0.75f},    {0, 0.61f},
        {0, 0.6f},      {0, 0.51f},     {0, 0.5f},       {0, 0.26f},    {0, 0.25f},
        {0, 0},         {0, -0.25f},    {0, -0.26f},     {0, -0.5f},    {0, -0.51f},
        {0, -0.6f},     {0, -0.61f},    {0, -1},         {0, -3},       {-1.6f, -0.2f},
        {-1.5f, -0.2f}, {-1.4f, -0.2f}, {1.4f, -0.2f},   {1.5f, -0.2f}, {1.6f, -0.2f},
        {-1.5f, 0.2f},  {1.5f, 0.2f},   {0.25f, -0.45f},
    };
    const float angles[] = {0, 0.35f, -0.65f};

    for (int a = 0; a < 3; ++a) {
        b2SimplexCache capsuleCache[3] = {0};
        b2SimplexCache polygonCache[3] = {0};

        for (int p = 0; p < (int)(sizeof(positions) / sizeof(positions[0])); ++p) {
            b2Transform xf = {positions[p], b2MakeRot(angles[a])};

            emit_manifold(
                ORACLE_KEY("%d/%d/circles", a, p),
                b2CollideCircles(&circle, &circle, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/capsule-circle", a, p),
                b2CollideCapsuleAndCircle(&capsule, &circle, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/segment-circle", a, p),
                b2CollideSegmentAndCircle(&segment, &circle, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/polygon-circle", a, p),
                b2CollidePolygonAndCircle(&polygon, &circle, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/capsules", a, p),
                b2CollideCapsules(&capsule, &capsule, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/segment-capsule", a, p),
                b2CollideSegmentAndCapsule(&segment, &capsule, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/polygon-capsule", a, p),
                b2CollidePolygonAndCapsule(&polygon, &capsule, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/polygons", a, p),
                b2CollidePolygons(&polygon, &polygon, xf)
            );
            emit_manifold(
                ORACLE_KEY("%d/%d/segment-polygon", a, p),
                b2CollideSegmentAndPolygon(&segment, &polygon, xf)
            );

            for (int c = 0; c < 3; ++c) {
                emit_manifold(
                    ORACLE_KEY("%d/%d/chain%d-circle", a, p, c),
                    b2CollideChainSegmentAndCircle(&chains[c], &circle, xf)
                );
                emit_manifold(
                    ORACLE_KEY("%d/%d/chain%d-capsule", a, p, c),
                    b2CollideChainSegmentAndCapsule(&chains[c], &capsule, xf, &capsuleCache[c])
                );
                emit_cache(ORACLE_KEY("%d/%d/chain%d-capsule-cache", a, p, c), &capsuleCache[c]);
                emit_manifold(
                    ORACLE_KEY("%d/%d/chain%d-polygon", a, p, c),
                    b2CollideChainSegmentAndPolygon(&chains[c], &polygon, xf, &polygonCache[c])
                );
                emit_cache(ORACLE_KEY("%d/%d/chain%d-polygon-cache", a, p, c), &polygonCache[c]);
            }
        }
    }

    ORACLE_CLOSE();
    return 0;
}
