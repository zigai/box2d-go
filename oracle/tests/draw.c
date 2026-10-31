// SPDX-License-Identifier: MIT
#include "box2d/box2d.h"
#include "oracle.h"

#include <assert.h>
#include <string.h>

static int commandCount;

static void command(int type, b2HexColor color) {
    ++commandCount;

    ORACLE_I32("command", type);
    ORACLE_U32("color", color);
}

static void transform(b2WorldTransform t) {
    ORACLE_POS("position", t.p);
    ORACLE_ROT("rotation", t.q);
}

static void polygon(b2WorldTransform t, const b2Vec2* v, int n, b2HexColor c, void* ctx) {
    (void)ctx;
    command(1, c);
    transform(t);
    ORACLE_I32("vertices", n);

    for (int i = 0; i < n; ++i) { ORACLE_VEC2("vertex", v[i]); }
}

static void
solidPolygon(b2WorldTransform t, const b2Vec2* v, int n, float radius, b2HexColor c, void* ctx) {
    (void)ctx;
    command(2, c);
    transform(t);
    ORACLE_I32("vertices", n);

    for (int i = 0; i < n; ++i) { ORACLE_VEC2("vertex", v[i]); }

    ORACLE_F32("radius", radius);
}

static void circle(b2Pos p, float radius, b2HexColor c, void* ctx) {
    (void)ctx;
    command(3, c);
    ORACLE_POS("point", p);
    ORACLE_F32("radius", radius);
}

static void solidCircle(b2WorldTransform t, b2Vec2 center, float radius, b2HexColor c, void* ctx) {
    (void)ctx;
    command(4, c);
    transform(t);
    ORACLE_VEC2("center", center);
    ORACLE_F32("radius", radius);
}

static void capsule(b2Pos p, b2Pos q, float radius, b2HexColor c, void* ctx) {
    (void)ctx;
    command(5, c);
    ORACLE_POS("point", p);
    ORACLE_POS("point", q);
    ORACLE_F32("radius", radius);
}

static void line(b2Pos p, b2Pos q, b2HexColor c, void* ctx) {
    (void)ctx;
    command(6, c);
    ORACLE_POS("point", p);
    ORACLE_POS("point", q);
}

static void axes(b2WorldTransform t, void* ctx) {
    (void)ctx;
    command(7, 0);
    transform(t);
}

static void point(b2Pos p, float size, b2HexColor c, void* ctx) {
    (void)ctx;
    command(8, c);
    ORACLE_POS("point", p);
    ORACLE_F32("size", size);
}

static void string(b2Pos p, const char* text, b2HexColor c, void* ctx) {
    (void)ctx;
    command(9, c);
    ORACLE_POS("point", p);
    ORACLE_BYTES("text", text, strlen(text));
}

static void bounds(b2AABB a, b2HexColor c, void* ctx) {
    (void)ctx;
    command(10, c);
    ORACLE_VEC2("lower", a.lowerBound);
    ORACLE_VEC2("upper", a.upperBound);
}

static void inspect(b2WorldId w) {
    b2Counters c = b2World_GetCounters(w);

    ORACLE_I32("bodies", c.bodyCount);
    ORACLE_I32("shapes", c.shapeCount);
    ORACLE_I32("joints", c.jointCount);
    ORACLE_I32("contacts", c.contactCount);
    ORACLE_I32("islands", c.islandCount);
    ORACLE_I32("staticHeight", c.staticTreeHeight);
    ORACLE_I32("treeHeight", c.treeHeight);
    ORACLE_I32("awakeContacts", c.awakeContactCount);
    ORACLE_I32("recycled", c.recycledContactCount);

    for (int i = 0; i < 24; ++i) { ORACLE_I32("color", c.colorCounts[i]); }

    b2AABB a = b2World_GetBounds(w);

    ORACLE_VEC2("bounds/lower", a.lowerBound);
    ORACLE_VEC2("bounds/upper", a.upperBound);
    ORACLE_I32("awakeBodies", b2World_GetAwakeBodyCount(w));
}

static void run(b2Pos base, const char* name) {
    ORACLE_OPEN(name);

    b2WorldDef wd  = b2DefaultWorldDef();
    wd.workerCount = 1;
    b2WorldId w    = b2CreateWorld(&wd);

    inspect(w);

    b2BodyId bodies[6];

    for (int i = 0; i < 6; ++i) {
        b2BodyDef bd = b2DefaultBodyDef();
        bd.position  = b2OffsetPos(
            base,
            (b2Vec2){
                i == 0   ? 0
                : i == 1 ? 0
                         : 2.0f * (float)i,
                i == 0   ? -.5f
                : i == 1 ? .45f
                         : 2
            }
        );
        bd.type = i == 1 || i == 2 ? b2_dynamicBody : i == 3 ? b2_kinematicBody : b2_staticBody;
        bd.name = i == 1 ? "contact box" : i == 2 ? "capsule" : "fixture";

        if (i == 2) { bd.rotation = b2MakeRot(.25f); }

        bodies[i]     = b2CreateBody(w, &bd);
        b2ShapeDef sd = b2DefaultShapeDef();

        if (i == 0 || i == 1) {
            b2Polygon p = b2MakeBox(i == 0 ? 3 : .5f, .5f);

            b2CreatePolygonShape(bodies[i], &sd, &p);
        }

        if (i == 2) {
            b2CreateCapsuleShape(bodies[i], &sd, &(b2Capsule){{-.3f, 0}, {.7f, 0}, .3f});
        }

        if (i == 3) {
            sd.isSensor = true;

            b2CreateCircleShape(bodies[i], &sd, &(b2Circle){{.2f, .1f}, .6f});
        }

        if (i == 4) {
            sd.material.customColor = 0x123456;

            b2CreateSegmentShape(bodies[i], &sd, &(b2Segment){{-1, 0}, {1, .5f}});
        }

        if (i == 5) {
            b2ChainDef cd   = b2DefaultChainDef();
            b2Vec2     p[4] = {{-1, -1}, {1, -1}, {1, 1}, {-1, 1}};
            cd.points       = p;
            cd.pointCount   = 4;
            cd.isLoop       = true;

            b2CreateChain(bodies[i], &cd);
        }

        inspect(w);
    }

    b2DistanceJointDef jd = b2DefaultDistanceJointDef();
    jd.base.bodyIdA       = bodies[1];
    jd.base.bodyIdB       = bodies[2];
    jd.length             = 4;
    jd.enableSpring       = true;
    jd.hertz              = 2;
    jd.enableLimit        = true;
    jd.minLength          = 2;
    jd.maxLength          = 5;

    b2CreateDistanceJoint(w, &jd);
    inspect(w);

    for (int i = 0; i < 3; ++i) {
        b2World_Step(w, 1.0f / 60, 4);
        inspect(w);
    }

    for (int option = 0; option < 17; ++option) {
        b2DebugDraw d         = b2DefaultDebugDraw();
        d.DrawPolygonFcn      = polygon;
        d.DrawSolidPolygonFcn = solidPolygon;
        d.DrawCircleFcn       = circle;
        d.DrawSolidCircleFcn  = solidCircle;
        d.DrawSolidCapsuleFcn = capsule;
        d.DrawLineFcn         = line;
        d.DrawTransformFcn    = axes;
        d.DrawPointFcn        = point;
        d.DrawStringFcn       = string;
        d.DrawBoundsFcn       = bounds;
        d.drawShapes          = option == 0 || option == 1 || option == 15 || option == 16;
        d.forceScale          = .3f;
        d.jointScale          = .7f;
        d.drawChainNormals    = option == 1 || option == 15;
        d.drawJoints          = option == 2 || option == 3 || option == 15;
        d.drawJointExtras     = option == 3 || option == 15;
        d.drawBounds          = option == 4 || option == 15;
        d.drawMass            = option == 5 || option == 15;
        d.drawBodyNames       = option == 6 || option == 15;
        d.drawContacts        = (option >= 7 && option <= 13) || option == 15;
        d.drawAnchorA         = option == 13;
        d.drawGraphColors     = option == 8 || option == 15;
        d.drawContactFeatures = option == 9 || option == 15;
        d.drawContactNormals  = option == 10;
        d.drawContactForces   = option == 11 || option == 15;
        d.drawFrictionForces  = option == 12 || option == 15;
        d.drawIslands         = option == 14 || option == 15;

        if (option == 16) {
            b2Vec2 center   = b2ToVec2(base);
            d.drawingBounds = (b2AABB){{center.x - 1, center.y - 1}, {center.x + 1, center.y + 1}};
        }

        uint64_t before = b2World_GetStateHash(w);

        ORACLE_U64("before", before);

        commandCount = 0;

        b2World_Draw(w, &d);
        ORACLE_I32("commands", commandCount);

        uint64_t after = b2World_GetStateHash(w);

        ORACLE_U64("after", after);
        assert(before == after);
    }

    for (int i = 5; i >= 0; --i) {
        b2DestroyBody(bodies[i]);
        inspect(w);
    }

    b2DestroyWorld(w);
    ORACLE_CLOSE();
}

int main(void) {
    run(b2Pos_zero, "draw/Origin");
    run((b2Pos){32, -16}, "draw/Offset");
#if defined(BOX2D_DOUBLE_PRECISION)
    run((b2Pos){10000000, -10000000}, "draw/Far");
#endif
    return 0;
}
