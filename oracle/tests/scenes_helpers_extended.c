// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT
#include "scenes/driver.h"
#include "shared/benchmarks.c"
#include "shared/human.c"
#include "shared/utils.c"

static const char* helper_key(const char* family, int operation, int item, const char* field) {
    static char key[96];

    snprintf(key, sizeof(key), "%s.%d.%d.%s", family, operation, item, field);

    return key;
}

static void helper_counts(b2WorldId w, const char* family, int op) {
    b2Counters c = b2World_GetCounters(w);

    ORACLE_I32(helper_key(family, op, 0, "bodies"), c.bodyCount);
    ORACLE_I32(helper_key(family, op, 0, "shapes"), c.shapeCount);
    ORACLE_I32(helper_key(family, op, 0, "joints"), c.jointCount);
}

static void human_observe(Human* h, int op) {
    ORACLE_BOOL(helper_key("human", op, 0, "spawned"), h->isSpawned);
    ORACLE_F32(helper_key("human", op, 0, "scale"), h->scale);

    for (int i = 0; i < bone_count; ++i) {
        b2BodyId  b = h->bones[i].bodyId;
        b2JointId j = h->bones[i].jointId;

        ORACLE_BOOL(helper_key("human", op, i, "bodyValid"), b2Body_IsValid(b));
        ORACLE_BOOL(helper_key("human", op, i, "jointValid"), b2Joint_IsValid(j));

        if (B2_IS_NON_NULL(b)) {
            ORACLE_VEC2(helper_key("human", op, i, "velocity"), b2Body_GetLinearVelocity(b));
            ORACLE_F32(helper_key("human", op, i, "angular"), b2Body_GetAngularVelocity(b));

            b2MassData m = b2Body_GetMassData(b);

            ORACLE_F32(helper_key("human", op, i, "mass"), m.mass);
            ORACLE_VEC2(helper_key("human", op, i, "center"), m.center);
            ORACLE_F32(helper_key("human", op, i, "inertia"), m.rotationalInertia);

            b2ShapeId shapes[2];
            int       n = b2Body_GetShapes(b, shapes, 2);

            ORACLE_I32(helper_key("human", op, i, "shapes"), n);

            for (int s = 0; s < n; ++s) {
                b2ShapeType type = b2Shape_GetType(shapes[s]);

                ORACLE_I32(helper_key("human", op, i, "shapeType"), (int32_t)type);
                ORACLE_BOOL(
                    helper_key("human", op, i, "sensor"),
                    b2Shape_AreSensorEventsEnabled(shapes[s])
                );

                if (type == b2_capsuleShape) {
                    b2Capsule c = b2Shape_GetCapsule(shapes[s]);

                    ORACLE_VEC2(helper_key("human", op, i, "c1"), c.center1);
                    ORACLE_VEC2(helper_key("human", op, i, "c2"), c.center2);
                    ORACLE_F32(helper_key("human", op, i, "radius"), c.radius);
                } else if (type == b2_polygonShape) {
                    b2Polygon p = b2Shape_GetPolygon(shapes[s]);

                    ORACLE_I32(helper_key("human", op, i, "vertices"), p.count);
                    ORACLE_F32(helper_key("human", op, i, "radius"), p.radius);
                    ORACLE_VEC2(helper_key("human", op, i, "centroid"), p.centroid);

                    for (int v = 0; v < p.count; ++v) {
                        ORACLE_VEC2(helper_key("human", op, i, "vertex"), p.vertices[v]);
                        ORACLE_VEC2(helper_key("human", op, i, "normal"), p.normals[v]);
                    }
                }
            }
        }

        if (B2_IS_NON_NULL(j)) {
            ORACLE_BOOL(helper_key("human", op, i, "motor"), b2RevoluteJoint_IsMotorEnabled(j));
            ORACLE_F32(helper_key("human", op, i, "torque"), b2RevoluteJoint_GetMaxMotorTorque(j));
            ORACLE_BOOL(helper_key("human", op, i, "spring"), b2RevoluteJoint_IsSpringEnabled(j));
            ORACLE_F32(helper_key("human", op, i, "hertz"), b2RevoluteJoint_GetSpringHertz(j));
            ORACLE_F32(
                helper_key("human", op, i, "damping"),
                b2RevoluteJoint_GetSpringDampingRatio(j)
            );

            b2Transform a = b2Joint_GetLocalFrameA(j), bframe = b2Joint_GetLocalFrameB(j);

            ORACLE_VEC2(helper_key("human", op, i, "anchorA"), a.p);
            ORACLE_ROT(helper_key("human", op, i, "rotationA"), a.q);
            ORACLE_VEC2(helper_key("human", op, i, "anchorB"), bframe.p);
            ORACLE_ROT(helper_key("human", op, i, "rotationB"), bframe.q);
        }
    }
}

static void rain_helpers(b2WorldDef* def) {
    b2WorldId w = b2CreateWorld(def);

    CreateRain(w);

    b2BodyId  first      = b2_nullBodyId;
    b2JointId firstJoint = b2_nullJointId;

    for (int step = 0; step < 1000; ++step) {
        ORACLE_F32(helper_key("rain", step, 0, "result"), StepRain(w, step));

        if (step == 0) {
            first      = g_rainData.groups[0].humans[0].bones[bone_torso].bodyId;
            firstJoint = g_rainData.groups[0].humans[0].bones[bone_torso].jointId;
        }

        ORACLE_I32(helper_key("rain", step, 0, "columns"), g_rainData.columnCount);
        ORACLE_I32(helper_key("rain", step, 0, "head"), g_rainData.columnIndex);
        ORACLE_BOOL(helper_key("rain", step, 0, "firstBody"), b2Body_IsValid(first));
        ORACLE_BOOL(helper_key("rain", step, 0, "firstJoint"), b2Joint_IsValid(firstJoint));
        helper_counts(w, "rain", step);

        if (step == 320 && (b2Body_IsValid(first) || b2Joint_IsValid(firstJoint))) { abort(); }

        b2World_Step(w, 1.0f / 60, 4);
    }

    b2DestroyWorld(w);
}

static void sleep_helpers(b2WorldDef* def) {
    b2WorldId w = b2CreateWorld(def);

    CreateSleep(w);
    b2World_Step(w, 1.0f / 60, 4);

    for (int i = 0; i < SLEEP_PYRAMID_COUNT; ++i) {
        b2Body_SetAwake(g_sleepData.bodyIdA[i], false);
    }

    for (int op = 0; op < 3; ++op) {
        if (op > 0) { ORACLE_F32(helper_key("sleep", op, 0, "result"), StepSleep(w, op)); }
        for (int i = 0; i < SLEEP_PYRAMID_COUNT; ++i) {
            ORACLE_BOOL(helper_key("sleep", op, i, "a"), b2Body_IsAwake(g_sleepData.bodyIdA[i]));
            ORACLE_BOOL(helper_key("sleep", op, i, "b"), b2Body_IsAwake(g_sleepData.bodyIdB[i]));
        }

        helper_counts(w, "sleep", op);
    }

    b2DestroyWorld(w);
}

static void lifecycle_helpers(b2WorldDef* def) {
    for (int cycle = 0; cycle < 2; ++cycle) {
        b2WorldId w = b2CreateWorld(def);

        CreateTreeCast(w);
        ORACLE_BOOL(helper_key("tree", cycle, 0, "created"), g_treeCast.created);
        ORACLE_I32(helper_key("tree", cycle, 0, "proxies"), g_treeCast.proxyCount);
        ORACLE_F32(helper_key("tree", cycle, 0, "result"), StepTreeCast(w, 0));

        b2TreeStats s = GetTreeCastBenchmarkStats();

        ORACLE_I32(helper_key("tree", cycle, 0, "nodes"), s.nodeVisits);
        ORACLE_I32(helper_key("tree", cycle, 0, "leaves"), s.leafVisits);
        DestroyTreeCast();
        DestroyTreeCast();
        ORACLE_BOOL(helper_key("tree", cycle, 0, "destroyed"), !g_treeCast.created);
        ORACLE_I32(helper_key("tree", cycle, 0, "cleared"), g_treeCast.proxyCount);

        s = GetTreeCastBenchmarkStats();

        ORACLE_I32(helper_key("tree", cycle, 0, "retainedNodes"), s.nodeVisits);
        ORACLE_I32(helper_key("tree", cycle, 0, "retainedLeaves"), s.leafVisits);
        b2DestroyWorld(w);

        w = b2CreateWorld(def);

        CreateTileWorld(w);

        g_tileWorld.groundHead = g_tileWorld.groundCount - 1;
        b2BodyId old           = g_tileWorld.groundIds[g_tileWorld.groundHead];

        ORACLE_F32(helper_key("tile", cycle, 0, "result"), StepTileWorld(w, 0));
        ORACLE_I32(helper_key("tile", cycle, 0, "head"), g_tileWorld.groundHead);
        ORACLE_BOOL(helper_key("tile", cycle, 0, "old"), b2Body_IsValid(old));
        ORACLE_F32(helper_key("tile", cycle, 0, "left"), g_tileWorld.xLeft);
        ORACLE_F32(helper_key("tile", cycle, 0, "next"), g_tileWorld.xNext);
        helper_counts(w, "tile", cycle);

        s = GetTileWorldBenchmarkStats();

        ORACLE_I32(helper_key("tile", cycle, 0, "nodes"), s.nodeVisits);
        ORACLE_I32(helper_key("tile", cycle, 0, "leaves"), s.leafVisits);
        DestroyTileWorld();
        DestroyTileWorld();
        ORACLE_I32(helper_key("tile", cycle, 0, "cleared"), g_tileWorld.groundCount);

        s = GetTileWorldBenchmarkStats();

        ORACLE_I32(helper_key("tile", cycle, 0, "retainedNodes"), s.nodeVisits);
        ORACLE_I32(helper_key("tile", cycle, 0, "retainedLeaves"), s.leafVisits);
        b2DestroyWorld(w);
    }
}

static void human_helpers(b2WorldDef* def) {
    b2WorldId w = b2CreateWorld(def);
    Human     h = {0};

    CreateHuman(&h, w, (b2Pos){0, 5}, 1, .05f, 5, .5f, 1, NULL, true);
    human_observe(&h, 0);
    Human_SetVelocity(&h, (b2Vec2){1, -2});
    human_observe(&h, 1);
    Human_ApplyRandomAngularImpulse(&h, .5f);
    human_observe(&h, 2);
    Human_SetJointFrictionTorque(&h, 0);
    human_observe(&h, 3);
    Human_SetJointFrictionTorque(&h, .1f);
    human_observe(&h, 4);
    Human_SetJointSpringHertz(&h, 0);
    human_observe(&h, 5);
    Human_SetJointSpringHertz(&h, 2);
    human_observe(&h, 6);
    Human_SetJointDampingRatio(&h, .75f);
    human_observe(&h, 7);
    Human_EnableSensorEvents(&h, true);
    human_observe(&h, 8);
    Human_SetScale(&h, 1.5f);
    human_observe(&h, 9);
    Human_EnableSensorEvents(&h, false);
    human_observe(&h, 10);

    Human old = h;

    DestroyHuman(&h);
    human_observe(&h, 11);

    for (int i = 0; i < bone_count; ++i) {
        ORACLE_BOOL(helper_key("human", 11, i, "oldBody"), b2Body_IsValid(old.bones[i].bodyId));
        ORACLE_BOOL(helper_key("human", 11, i, "oldJoint"), b2Joint_IsValid(old.bones[i].jointId));
    }

    helper_counts(w, "human", 11);

    Human     sparse               = {0};
    b2BodyDef bd                   = b2DefaultBodyDef();
    bd.type                        = b2_dynamicBody;
    sparse.bones[bone_hip].bodyId  = b2CreateBody(w, &bd);
    sparse.bones[bone_head].bodyId = b2CreateBody(w, &bd);

    Human_SetVelocity(&sparse, (b2Vec2){2, -3});
    ORACLE_VEC2("human.sparse.hip", b2Body_GetLinearVelocity(sparse.bones[bone_hip].bodyId));
    ORACLE_VEC2("human.sparse.head", b2Body_GetLinearVelocity(sparse.bones[bone_head].bodyId));
    ORACLE_BOOL("human.sparse.torsoNull", B2_IS_NULL(sparse.bones[bone_torso].bodyId));
    b2DestroyWorld(w);
}

int main(void) {
    for (int workers = 1; workers <= 4; workers += 3) {
        g_randomSeed = 12345;
        char name[64];

        snprintf(name, sizeof(name), "scenes/ExtendedHelpers_workers%d", workers);
        ORACLE_OPEN(name);

        b2WorldDef def  = b2DefaultWorldDef();
        def.workerCount = workers;

        rain_helpers(&def);
        sleep_helpers(&def);
        lifecycle_helpers(&def);
        human_helpers(&def);
        ORACLE_CLOSE();
    }

    return 0;
}
