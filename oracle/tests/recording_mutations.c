// SPDX-License-Identifier: MIT
#include "box2d/box2d.h"
#include "oracle.h"
#include "replay_events.h"
#include "replay_observe.h"
#include "replay_properties.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#define REQUIRE(x)                                                                                 \
    do {                                                                                           \
        if (!(x)) abort();                                                                         \
    } while (0)

int main(void) {
    ORACLE_OPEN("recording/MutationParity");

    b2WorldDef wd        = b2DefaultWorldDef();
    wd.gravity           = (b2Vec2){0};
    b2WorldId    w       = b2CreateWorld(&wd);
    b2Recording* rec     = b2CreateRecording(0);
    b2Recording* ignored = b2CreateRecording(0);

    b2World_StartRecording(w, rec);
    b2World_StartRecording(w, ignored);

    b2BodyDef bd          = b2DefaultBodyDef();
    bd.name               = "ground";
    b2BodyId ground       = b2CreateBody(w, &bd);
    bd.type               = b2_dynamicBody;
    bd.position           = (b2Pos){2, 3};
    bd.name               = "moving";
    b2BodyId   body       = b2CreateBody(w, &bd);
    b2ShapeDef sd         = b2DefaultShapeDef();
    b2Circle   circle     = {{0, 0}, 0.5f};
    b2ShapeId  shape      = b2CreateCircleShape(body, &sd, &circle);
    b2ShapeDef sensorDef  = b2DefaultShapeDef();
    sensorDef.isSensor    = true;
    b2Circle sensorCircle = {{2, 3}, 2};

    b2CreateCircleShape(ground, &sensorDef, &sensorCircle);

    b2MoverJointDef md       = b2DefaultMoverJointDef();
    md.base.bodyIdA          = ground;
    md.base.bodyIdB          = body;
    b2JointId      mover     = b2CreateMoverJoint(w, &md);
    b2PogoJointDef pd        = b2DefaultPogoJointDef();
    pd.base.bodyIdA          = ground;
    pd.base.bodyIdB          = body;
    pd.normal                = (b2Vec2){0, 1};
    pd.restLength            = 3;
    pd.maxCompressionForce   = 15;
    pd.maxTensionForce       = 5;
    b2JointId       pogo     = b2CreatePogoJoint(w, &pd);
    b2WheelJointDef wheelDef = b2DefaultWheelJointDef();
    wheelDef.base.bodyIdA    = ground;
    wheelDef.base.bodyIdB    = body;
    b2JointId wheel          = b2CreateWheelJoint(w, &wheelDef);

    b2WheelJoint_EnableSpring(wheel, false);

    b2Vec2            points[] = {{-4, 0}, {-2, 0}, {0, 0}, {2, 0}, {4, 0}};
    b2SurfaceMaterial material = b2DefaultSurfaceMaterial();
    b2ChainDef        cd       = b2DefaultChainDef();
    cd.points                  = points;
    cd.pointCount              = 5;
    cd.ghost1                  = (b2Vec2){-6, 0};
    cd.ghost2                  = (b2Vec2){6, 0};
    cd.materials               = &material;
    cd.materialCount           = 1;
    b2ChainId chain            = b2CreateChain(ground, &cd);

    for (int phase = 0; phase < 18; ++phase) {
        switch (phase) {
            case 0: {
                b2Capsule cap = {{-0.4f, 0}, {0.6f, 0}, 0.3f};

                b2Shape_SetCapsule(shape, &cap);
                b2Body_SetMassData(body, (b2MassData){7, {0.2f, 0}, 2});
                b2Body_SetName(body, "mutated name");
                b2Body_EnableContactEvents(body, true);
                b2Body_EnableHitEvents(body, true);

                b2SurfaceMaterial surface = b2DefaultSurfaceMaterial();
                surface.friction          = 0.65f;
                surface.restitution       = 0.35f;
                surface.rollingResistance = 0.15f;
                surface.tangentSpeed      = 2;
                surface.userMaterialId    = UINT64_C(0x123456789abcdef0);
                surface.customColor       = 0xabcdef;

                b2Shape_SetSurfaceMaterial(shape, &surface);
                b2Shape_EnableSensorEvents(shape, true);
                b2Shape_EnablePreSolveEvents(shape, true);
                b2Chain_SetAllSurfaceMaterials(chain, &surface);

                surface.friction       = 0.85f;
                surface.userMaterialId = 19;

                b2Chain_SetSurfaceMaterial(chain, &surface, 2);
                b2MoverJoint_SetLinearVelocity(mover, (b2Vec2){0.7f, -0.1f});
                b2MoverJoint_SetMaxVelocityForce(mover, (b2Vec2){3, 4});
                b2PogoJoint_SetRestLength(pogo, 2.5f);
                b2PogoJoint_SetSpringHertz(pogo, 2);
                b2PogoJoint_SetSpringDampingRatio(pogo, 0.8f);
                b2Joint_SetLocalFrameA(mover, (b2Transform){{0.1f, 0.2f}, b2MakeRot(0.1f)});
                b2Joint_SetLocalFrameB(mover, (b2Transform){{-0.2f, 0}, b2MakeRot(-0.2f)});
                b2Joint_SetCollideConnected(mover, true);
                b2Joint_SetConstraintTuning(mover, 17, 0.3f);
                b2Joint_SetForceThreshold(mover, 0.01f);
                b2Joint_SetTorqueThreshold(mover, 0.02f);
                b2World_EnableSleeping(w, false);
                b2World_EnableContinuous(w, false);
                b2World_EnableWarmStarting(w, false);
                b2World_EnableRestitutionPropagation(w, false);

                break;
            }
            case 1: {
                b2Body_UpdateMassFromShapes(body);

                b2Polygon polygon = b2MakeBox(0.7f, 0.2f);

                b2Shape_SetPolygon(shape, &polygon);
                b2Shape_SetUserMaterial(shape, UINT64_C(0xfedcba9876543210));
                b2Body_EnableContactEvents(body, false);
                b2Body_EnableHitEvents(body, false);
                b2Shape_EnableSensorEvents(shape, false);
                b2Shape_EnablePreSolveEvents(shape, false);

                break;
            }
            case 2: {
                b2Body_SetType(body, b2_kinematicBody);
                b2Body_SetLinearVelocity(body, (b2Vec2){0.25f, 0});

                b2Segment segment = {{-0.6f, 0}, {0.8f, 0.3f}};

                b2Shape_SetSegment(shape, &segment);

                break;
            }
            case 3: {
                b2Body_SetType(body, b2_dynamicBody);

                b2ChainSegment segment = {{-2, 0}, {{-1, 0}, {1, 0}}, {2, 0}, -1};

                b2Shape_SetChainSegment(shape, &segment);

                break;
            }
            case 4:
                b2Shape_SetCircle(shape, &circle);
                b2Body_Disable(body);

                break;
            case 5:

                b2Body_Enable(body);
                b2Body_SetMotionLocks(body, (b2MotionLocks){true, false, true});
                b2Body_EnableSleep(body, true);
                b2World_EnableSleeping(w, true);

                break;
            case 6: b2Body_SetAwake(body, false); break;
            case 7:

                b2Joint_WakeBodies(mover);
                b2Body_SetMotionLocks(body, (b2MotionLocks){0});

                break;
            case 8:

                b2Body_EnableSleep(body, false);
                b2Body_ApplyForceToCenter(body, (b2Vec2){40, 20}, true);
                b2Body_ApplyTorque(body, 3, true);
                b2Body_ClearForces(body);

                break;
            case 9: {
                b2DestroyJoint(mover);
                b2DestroyJoint(pogo);
                b2DestroyChain(chain);
                b2DestroyShape(shape, true);

                b2ShapeId replacement = b2CreateCircleShape(body, &sd, &circle);

                b2Shape_SetDensity(replacement, 3, false);
                b2Body_UpdateMassFromShapes(body);

                break;
            }
            case 10:
                b2DestroyBody(body);

                bd.name     = "reused body";
                bd.position = (b2Pos){4, 6};
                body        = b2CreateBody(w, &bd);
                shape       = b2CreateCircleShape(body, &sd, &circle);
                break;
            case 11:

                b2Body_SetTransform(body, (b2Pos){5, 7}, b2MakeRot(0.3f));
                b2Body_SetLinearVelocity(body, (b2Vec2){0, -1});
                b2Body_SetAngularVelocity(body, 0.4f);

                break;
            case 12:

                b2Body_EnableSleep(body, true);
                b2Body_SetAwake(body, false);

                break;
            case 13: b2Body_SetAwake(body, true); break;
            case 14: {
                b2Circle support = {{5, 7}, 1};

                b2CreateCircleShape(ground, &sd, &support);
                b2Body_EnableContactEvents(body, true);
                b2Body_EnableHitEvents(body, true);
                b2World_SetHitEventThreshold(w, 0.01f);

                break;
            }
            case 15: b2Body_SetAwake(body, false); break;
            case 16: b2Body_WakeTouching(ground); break;
            case 17: {
                b2Pos point = b2OffsetPos(b2Body_GetPosition(body), (b2Vec2){0.3f, 0.2f});

                b2Body_ApplyForce(body, (b2Vec2){4, 6}, point, true);
                b2Body_ApplyForceToCenter(body, (b2Vec2){2, 3}, true);
                b2Body_ApplyTorque(body, 1.5f, true);
                b2Body_ApplyLinearImpulse(body, (b2Vec2){0.2f, 0.1f}, point, true);
                b2Body_ApplyLinearImpulseToCenter(body, (b2Vec2){0.3f, 0.2f}, true);
                b2Body_ApplyAngularImpulse(body, 0.4f, true);

                break;
            }
            default: break;
        }

        for (int sub = 0; sub < 2; ++sub) {
            b2World_Step(w, 1.0f / 60.0f, 4);

            char key[80];

            snprintf(key, sizeof(key), "live/%d", phase * 2 + sub + 1);
            ORACLE_U64(ORACLE_KEY("%s/state", key), b2World_GetStateHash(w));
            replay_properties(w, key);
            replay_events(w, key);
        }
    }

    b2World_StopRecording(w);
    b2World_StopRecording(w);
    ORACLE_BOOL("repeatedStart/ignoredEmpty", b2Recording_GetSize(ignored) == 0);
    b2DestroyRecording(ignored);
    b2Body_SetName(body, "after stop");
    b2DestroyWorld(w);

    b2Replay* p = b2CreateReplay(b2Recording_GetData(rec), b2Recording_GetSize(rec), 0);

    REQUIRE(p);
    b2Replay_SetKeyframePolicy(p, 1u << 26, 2);

    while (b2Replay_StepFrame(p)) {
        char key[80];

        snprintf(key, sizeof(key), "replay/%d", b2Replay_GetFrame(p));
        ORACLE_U64(ORACLE_KEY("%s/state", key), b2World_GetStateHash(b2Replay_GetWorldId(p)));
        replay_properties(b2Replay_GetWorldId(p), key);
        replay_events(b2Replay_GetWorldId(p), key);
    }

    ORACLE_BOOL("replay/diverged", b2Replay_HasDiverged(p));
    REQUIRE(!b2Replay_HasDiverged(p));

    int targets[] = {36, 1, 14, 10, 0, 18, 35, 99, -5};

    for (int i = 0; i < 9; ++i) {
        b2Replay_SeekFrame(p, targets[i]);

        char key[80];

        snprintf(key, sizeof(key), "seek/%d", i);
        replay_observe(p, key);
        replay_properties(b2Replay_GetWorldId(p), key);
        replay_events(b2Replay_GetWorldId(p), key);
    }

    b2DestroyReplay(p);

    wd.gravity = (b2Vec2){0, -2};
    w          = b2CreateWorld(&wd);

    b2World_StartRecording(w, rec);

    bd.name = "new session";
    body    = b2CreateBody(w, &bd);

    b2CreateCircleShape(body, &sd, &circle);
    b2World_Step(w, 1.0f / 20.0f, 3);
    b2World_StopRecording(w);
    b2DestroyWorld(w);

    p = b2CreateReplay(b2Recording_GetData(rec), b2Recording_GetSize(rec), 0);

    REQUIRE(p);
    replay_observe(p, "reuse/open");
    ORACLE_BOOL("reuse/step", b2Replay_StepFrame(p));
    replay_observe(p, "reuse/after");
    b2DestroyReplay(p);
    b2DestroyRecording(rec);
    ORACLE_CLOSE();

    return 0;
}
