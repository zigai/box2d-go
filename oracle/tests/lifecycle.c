// SPDX-License-Identifier: MIT
#include "box2d/box2d.h"
#include "oracle.h"

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>

static uint64_t shape_key(b2ShapeId id) {
    return b2StoreShapeId(id) & UINT64_C(0xffffffff0000ffff);
}

static uint64_t body_key(b2BodyId id) {
    return b2StoreBodyId(id) & UINT64_C(0xffffffff0000ffff);
}

static void null_handles(void) {
    ORACLE_OPEN("lifecycle/NullHandles");
    ORACLE_BOOL("world.valid", b2World_IsValid(b2_nullWorldId));
    ORACLE_BOOL("body.valid", b2Body_IsValid(b2_nullBodyId));
    ORACLE_BOOL("shape.valid", b2Shape_IsValid(b2_nullShapeId));
    ORACLE_BOOL("chain.valid", b2Chain_IsValid(b2_nullChainId));
    ORACLE_BOOL("joint.valid", b2Joint_IsValid(b2_nullJointId));
    ORACLE_BOOL("contact.valid", b2Contact_IsValid(b2_nullContactId));
    ORACLE_CLOSE();
}

static b2ChainId make_chain(b2BodyId body) {
    b2Vec2            points[] = {{10, 0}, {14, 0}, {14, 4}, {10, 4}};
    b2SurfaceMaterial material = b2DefaultSurfaceMaterial();
    b2ChainDef        d        = b2DefaultChainDef();
    d.points                   = points;
    d.pointCount               = 4;
    d.materials                = &material;
    d.materialCount            = 1;
    d.isLoop                   = true;

    return b2CreateChain(body, &d);
}

static b2JointId make_joint(b2WorldId w, b2BodyId a, b2BodyId b) {
    b2RevoluteJointDef d    = b2DefaultRevoluteJointDef();
    d.base.bodyIdA          = a;
    d.base.bodyIdB          = b;
    d.base.localFrameA.p    = b2Body_GetLocalPoint(a, b2Body_GetPosition(b));
    d.base.collideConnected = true;

    return b2CreateRevoluteJoint(w, &d);
}

static void membership(
    b2WorldId w,
    b2BodyId  b,
    b2BodyId  peer,
    b2ShapeId shape,
    b2ChainId chain,
    b2JointId joint
) {
    ORACLE_BOOL("body.world", b2StoreWorldId(b2Body_GetWorld(b)) == b2StoreWorldId(w));
    ORACLE_BOOL("shape.world", b2StoreWorldId(b2Shape_GetWorld(shape)) == b2StoreWorldId(w));
    ORACLE_BOOL("shape.body", B2_ID_EQUALS(b2Shape_GetBody(shape), b));
    ORACLE_BOOL("chain.world", b2StoreWorldId(b2Chain_GetWorld(chain)) == b2StoreWorldId(w));
    ORACLE_BOOL("joint.world", b2StoreWorldId(b2Joint_GetWorld(joint)) == b2StoreWorldId(w));
    ORACLE_BOOL("joint.a", B2_ID_EQUALS(b2Joint_GetBodyA(joint), b));
    ORACLE_BOOL("joint.b", B2_ID_EQUALS(b2Joint_GetBodyB(joint), peer));

    b2ShapeId shapes[32];
    int       n = b2Body_GetShapes(b, shapes, 32);

    ORACLE_I32("shape.count", b2Body_GetShapeCount(b));
    ORACLE_I32("shape.enumerated", n);

    for (int i = 0; i < n; ++i) { ORACLE_U64("shape.member", shape_key(shapes[i])); }

    b2JointId joints[32];
    n = b2Body_GetJoints(b, joints, 32);

    ORACLE_I32("joint.count", b2Body_GetJointCount(b));
    ORACLE_I32("joint.enumerated", n);

    for (int i = 0; i < n; ++i) { ORACLE_BOOL("joint.member", B2_ID_EQUALS(joints[i], joint)); }

    n = b2Chain_GetSegments(chain, shapes, 32);

    ORACLE_I32("chain.count", b2Chain_GetSegmentCount(chain));
    ORACLE_I32("chain.enumerated", n);

    for (int i = 0; i < n; ++i) {
        ORACLE_U64("segment", shape_key(shapes[i]));
        ORACLE_BOOL("segment.parent", B2_ID_EQUALS(b2Shape_GetParentChain(shapes[i]), chain));
        ORACLE_BOOL("segment.body", B2_ID_EQUALS(b2Shape_GetBody(shapes[i]), b));
    }
}

static void
validity(b2WorldId w, b2BodyId b, b2ShapeId s, b2ChainId c, b2JointId j, b2ContactId contact) {
    ORACLE_BOOL("world.valid", b2World_IsValid(w));
    ORACLE_BOOL("body.valid", b2Body_IsValid(b));
    ORACLE_BOOL("shape.valid", b2Shape_IsValid(s));
    ORACLE_BOOL("chain.valid", b2Chain_IsValid(c));
    ORACLE_BOOL("joint.valid", b2Joint_IsValid(j));
    ORACLE_BOOL("contact.valid", b2Contact_IsValid(contact));
}

static void lifetimes(void) {
    ORACLE_OPEN("lifecycle/ReuseAndDestruction");

    b2WorldId   oldw       = b2_nullWorldId;
    b2BodyId    oldb       = b2_nullBodyId;
    b2ShapeId   olds       = b2_nullShapeId;
    b2ChainId   oldc       = b2_nullChainId;
    b2JointId   oldj       = b2_nullJointId;
    b2ContactId oldcontact = b2_nullContactId;

    for (int round = 0; round < 4; ++round) {
        b2WorldDef wd = b2DefaultWorldDef();
        wd.gravity    = (b2Vec2){0, 0};
        b2WorldId w   = b2CreateWorld(&wd);

        validity(oldw, oldb, olds, oldc, oldj, oldcontact);

        b2BodyDef  bd          = b2DefaultBodyDef();
        b2BodyId   ground      = b2CreateBody(w, &bd);
        b2ShapeDef sd          = b2DefaultShapeDef();
        sd.enableContactEvents = true;
        b2Circle circle        = {{0, 0}, 1};

        b2CreateCircleShape(ground, &sd, &circle);

        bd.type         = b2_dynamicBody;
        b2BodyId  b     = b2CreateBody(w, &bd);
        b2ShapeId s     = b2CreateCircleShape(b, &sd, &circle);
        circle.center   = (b2Vec2){4, 0};
        b2ShapeId extra = b2CreateCircleShape(b, &sd, &circle);
        circle.center   = (b2Vec2){0, 0};
        bd.position     = (b2Pos){0, 8};
        b2BodyId  peer  = b2CreateBody(w, &bd);
        b2ChainId c     = make_chain(b);
        b2JointId j     = make_joint(w, b, peer);

        membership(w, b, peer, s, c, j);
        validity(oldw, oldb, olds, oldc, oldj, oldcontact);
        b2World_Step(w, 1.0f / 60, 4);

        b2ContactData data[16];
        int           n = b2Shape_GetContactData(s, data, 16);

        assert(n == 1);
        ORACLE_I32("contact.count", n);

        b2ContactId contact = data[0].contactId;

        ORACLE_BOOL(
            "contact.member",
            B2_ID_EQUALS(data[0].shapeIdA, s) || B2_ID_EQUALS(data[0].shapeIdB, s)
        );
        ORACLE_I32("contact.points", data[0].manifold.pointCount);
        validity(w, b, s, c, j, contact);

        b2ContactData direct = b2Contact_GetData(contact);

        ORACLE_BOOL("contact.direct.id", B2_ID_EQUALS(direct.contactId, contact));
        ORACLE_U64("contact.direct.a", shape_key(direct.shapeIdA));
        ORACLE_U64("contact.direct.b", shape_key(direct.shapeIdB));
        ORACLE_VEC2("contact.normal", direct.manifold.normal);
        ORACLE_I32("contact.direct.points", direct.manifold.pointCount);

        b2ContactData bodyData[16];
        int           bodyCount = b2Body_GetContactData(b, bodyData, 16);

        ORACLE_I32("body.contact.count", bodyCount);

        int matches = 0;

        for (int i = 0; i < bodyCount; ++i) {
            if (B2_ID_EQUALS(bodyData[i].contactId, contact)) { ++matches; }
        }

        ORACLE_I32("body.contact.multiplicity", matches);
        assert(matches == 1);
        b2Body_SetTransform(b, (b2Pos){30, 0}, (b2Rot){1, 0});
        b2World_Step(w, 1.0f / 60, 4);
        ORACLE_BOOL("separated.contact.valid", b2Contact_IsValid(contact));
        b2Body_SetTransform(b, (b2Pos){0, 0}, (b2Rot){1, 0});
        b2World_Step(w, 1.0f / 60, 4);

        n = b2Shape_GetContactData(s, data, 16);

        ORACLE_I32("returned.contacts", n);
        assert(n == 1);
        ORACLE_BOOL("contact.reused", data[0].contactId.index1 == contact.index1);
        ORACLE_U32("contact.generation.delta", data[0].contactId.generation - contact.generation);
        ORACLE_BOOL("old.contact.valid", b2Contact_IsValid(contact));

        contact = data[0].contactId;

        b2DestroyShape(extra, true);

        b2ShapeId replacement = b2CreateCircleShape(b, &sd, &circle);

        ORACLE_BOOL("shape.old.valid", b2Shape_IsValid(extra));
        ORACLE_BOOL("shape.reused", extra.index1 == replacement.index1);
        ORACLE_U32("shape.generation.delta", (uint16_t)(replacement.generation - extra.generation));

        b2ShapeId segments[8];
        int       ns = b2Chain_GetSegments(c, segments, 8);

        b2DestroyChain(c);

        b2ChainId nc = make_chain(b);

        ORACLE_BOOL("chain.old.valid", b2Chain_IsValid(c));
        ORACLE_BOOL("chain.reused", c.index1 == nc.index1);
        ORACLE_U32("chain.generation.delta", (uint16_t)(nc.generation - c.generation));

        for (int i = 0; i < ns; ++i) {
            ORACLE_BOOL("segment.old.valid", b2Shape_IsValid(segments[i]));
        }

        c = nc;

        b2DestroyJoint(j);

        b2JointId nj = make_joint(w, b, peer);

        ORACLE_BOOL("joint.old.valid", b2Joint_IsValid(j));
        ORACLE_BOOL("joint.reused", j.index1 == nj.index1);
        ORACLE_U32("joint.generation.delta", (uint16_t)(nj.generation - j.generation));

        j = nj;

        membership(w, b, peer, s, c, j);

        ns = b2Chain_GetSegments(c, segments, 8);

        b2DestroyBody(b);
        validity(w, b, s, c, j, contact);
        ORACLE_BOOL("replacement.valid", b2Shape_IsValid(replacement));

        for (int i = 0; i < ns; ++i) {
            ORACLE_BOOL("dependent.segment.valid", b2Shape_IsValid(segments[i]));
        }

        ORACLE_I32("peer.joints", b2Body_GetJointCount(peer));

        bd.position = (b2Pos){20, 0};
        b2BodyId nb = b2CreateBody(w, &bd);

        ORACLE_BOOL("body.reused", b.index1 == nb.index1);
        ORACLE_U32("body.generation.delta", (uint16_t)(nb.generation - b.generation));
        ORACLE_BOOL("body.old.valid", b2Body_IsValid(b));
        b2World_Step(w, 1.0f / 60, 4);

        b2Counters counts = b2World_GetCounters(w);

        ORACLE_I32("bodies", counts.bodyCount);
        ORACLE_I32("shapes", counts.shapeCount);
        ORACLE_I32("joints", counts.jointCount);
        ORACLE_I32("contacts", counts.contactCount);
        ORACLE_POS("peer.position", b2Body_GetPosition(peer));
        b2Body_SetTransform(nb, (b2Pos){0, 0}, (b2Rot){1, 0});

        oldw = w;
        oldb = nb;
        olds = b2CreateCircleShape(nb, &sd, &circle);
        oldc = make_chain(nb);
        oldj = make_joint(w, nb, peer);

        b2World_Step(w, 1.0f / 60, 4);

        n = b2Shape_GetContactData(olds, data, 16);

        ORACLE_I32("world.destroy.contacts", n);
        assert(n == 1);

        oldcontact = data[0].contactId;

        validity(oldw, oldb, olds, oldc, oldj, oldcontact);
        b2DestroyWorld(w);
        validity(oldw, oldb, olds, oldc, oldj, oldcontact);
    }

    ORACLE_CLOSE();
}

typedef struct scene {
    b2WorldId w;
    b2BodyId  bodies[12];
    b2ShapeId shapes[14];
    b2JointId joint;
} scene;

static scene make_scene(int variant, int workers) {
    scene      s           = {0};
    b2WorldDef wd          = b2DefaultWorldDef();
    wd.workerCount         = workers;
    wd.gravity             = (b2Vec2){variant ? 2 : 0, variant ? -6 : -10};
    wd.enableSleep         = false;
    s.w                    = b2CreateWorld(&wd);
    b2BodyDef bd           = b2DefaultBodyDef();
    bd.position            = oracle_pos_xy(0, -0.5);
    b2BodyId   floor       = b2CreateBody(s.w, &bd);
    b2Polygon  box         = b2MakeBox(20, 0.5f);
    b2ShapeDef sd          = b2DefaultShapeDef();
    sd.enableContactEvents = true;
    sd.enableSensorEvents  = true;
    s.shapes[0]            = b2CreatePolygonShape(floor, &sd, &box);
    bd.position            = (b2Pos){0, 2};
    b2BodyId sensor        = b2CreateBody(s.w, &bd);
    sd.isSensor            = true;
    box                    = b2MakeBox(8, 0.5f);
    s.shapes[1]            = b2CreatePolygonShape(sensor, &sd, &box);
    sd.isSensor            = false;
    box                    = b2MakeBox(0.4f, 0.4f);
    bd.type                = b2_dynamicBody;

    for (int i = 0; i < 12; ++i) {
        int row           = i / 4;
        bd.position       = (b2Pos){(float)(i % 4 - 2) * 1.1f, 0.5f + (float)row * 1.1f};
        bd.linearVelocity = (b2Vec2){variant ? 0.3f : -0.2f, 0};
        s.bodies[i]       = b2CreateBody(s.w, &bd);
        s.shapes[i + 2]   = b2CreatePolygonShape(s.bodies[i], &sd, &box);
    }

    s.joint = make_joint(s.w, s.bodies[8], s.bodies[9]);
    return s;
}

static void observe_scene(scene* s, int step) {
    ORACLE_U64(ORACLE_KEY("step%d.state", step), b2World_GetStateHash(s->w));
    ORACLE_VEC2("gravity", b2World_GetGravity(s->w));
    ORACLE_I32("workers", b2World_GetWorkerCount(s->w));

    for (int i = 0; i < 12; ++i) {
        ORACLE_POS("position", b2Body_GetPosition(s->bodies[i]));
        ORACLE_ROT("rotation", b2Body_GetRotation(s->bodies[i]));
        ORACLE_VEC2("velocity", b2Body_GetLinearVelocity(s->bodies[i]));
        ORACLE_F32("angular", b2Body_GetAngularVelocity(s->bodies[i]));
        ORACLE_BOOL("awake", b2Body_IsAwake(s->bodies[i]));
    }

    b2ContactEvents ce = b2World_GetContactEvents(s->w);

    ORACLE_I32("begin.count", ce.beginCount);

    for (int i = 0; i < ce.beginCount; ++i) {
        b2ContactId id = ce.beginEvents[i].contactId;

        ORACLE_U64("begin.a", shape_key(ce.beginEvents[i].shapeIdA));
        ORACLE_U64("begin.b", shape_key(ce.beginEvents[i].shapeIdB));
        ORACLE_U32("begin.contact.index", (uint32_t)id.index1);
        ORACLE_U32("begin.contact.generation", id.generation);

        bool valid = b2Contact_IsValid(id);

        ORACLE_BOOL("begin.contact.valid", valid);

        if (valid) {
            b2ContactData data = b2Contact_GetData(id);

            ORACLE_VEC2("begin.normal", data.manifold.normal);
            ORACLE_I32("begin.points", data.manifold.pointCount);
        }
    }

    ORACLE_I32("end.count", ce.endCount);

    for (int i = 0; i < ce.endCount; ++i) {
        ORACLE_U64("end.a", shape_key(ce.endEvents[i].shapeIdA));
        ORACLE_U64("end.b", shape_key(ce.endEvents[i].shapeIdB));
        ORACLE_U32("end.contact.index", (uint32_t)ce.endEvents[i].contactId.index1);
        ORACLE_U32("end.contact.generation", ce.endEvents[i].contactId.generation);
        ORACLE_BOOL("end.contact.valid", b2Contact_IsValid(ce.endEvents[i].contactId));
    }

    ORACLE_I32("hit.count", ce.hitCount);

    for (int i = 0; i < ce.hitCount; ++i) {
        b2ContactHitEvent e = ce.hitEvents[i];

        ORACLE_U64("hit.a", shape_key(e.shapeIdA));
        ORACLE_U64("hit.b", shape_key(e.shapeIdB));
        ORACLE_U32("hit.contact.index", (uint32_t)e.contactId.index1);
        ORACLE_U32("hit.contact.generation", e.contactId.generation);
        ORACLE_POS("hit.point", e.point);
        ORACLE_VEC2("hit.normal", e.normal);
        ORACLE_F32("hit.speed", e.approachSpeed);
    }

    b2SensorEvents se = b2World_GetSensorEvents(s->w);

    ORACLE_I32("sensor.begin.count", se.beginCount);

    for (int i = 0; i < se.beginCount; ++i) {
        ORACLE_U64("sensor.begin.sensor", shape_key(se.beginEvents[i].sensorShapeId));
        ORACLE_U64("sensor.begin.visitor", shape_key(se.beginEvents[i].visitorShapeId));
    }

    ORACLE_I32("sensor.end.count", se.endCount);

    for (int i = 0; i < se.endCount; ++i) {
        ORACLE_U64("sensor.end.sensor", shape_key(se.endEvents[i].sensorShapeId));
        ORACLE_U64("sensor.end.visitor", shape_key(se.endEvents[i].visitorShapeId));
    }

    b2BodyEvents be = b2World_GetBodyEvents(s->w);

    ORACLE_I32("move.count", be.moveCount);

    for (int i = 0; i < be.moveCount; ++i) {
        ORACLE_U64("move.body", body_key(be.moveEvents[i].bodyId));
        ORACLE_POS("move.position", be.moveEvents[i].transform.p);
        ORACLE_ROT("move.rotation", be.moveEvents[i].transform.q);
        ORACLE_BOOL("move.sleep", be.moveEvents[i].fellAsleep);
    }
}

static void ownership(void) {
    int counts[] = {1, 4, 8, 32, 0, -5, 4, 4};

    for (int variant = 0; variant < 2; ++variant) {
        for (int mode = 0; mode < 2; ++mode) {
            char name[80];

            snprintf(
                name,
                sizeof name,
                "ownership/Scene%d-%s",
                variant,
                mode ? "Reconfigure" : "Serial"
            );
            ORACLE_OPEN(name);

            scene s = make_scene(variant, 1);

            for (int step = 0; step < 48; ++step) {
                if (mode && step % 6 == 0) { b2World_SetWorkerCount(s.w, counts[step / 6]); }

                b2World_Step(s.w, 1.0f / 60, 4);
                observe_scene(&s, step);
            }

            b2DestroyWorld(s.w);
            ORACLE_CLOSE();
        }
    }
    // One world is destroyed/recreated while the other retains its own settings and progresses.

    ORACLE_OPEN("lifecycle/Multiworld");

    scene     a = make_scene(0, 1), b = make_scene(1, 4);
    b2WorldId old = a.w;

    for (int step = 0; step < 12; ++step) {
        b2World_Step(a.w, 1.0f / 60, 4);
        b2World_Step(b.w, 1.0f / 60, 4);
        observe_scene(&a, step);
        observe_scene(&b, step);
    }

    b2DestroyWorld(a.w);

    a = make_scene(0, 2);

    ORACLE_BOOL("old.world.valid", b2World_IsValid(old));
    ORACLE_BOOL("worlds.distinct", b2StoreWorldId(a.w) != b2StoreWorldId(b.w));

    for (int step = 12; step < 24; ++step) {
        b2World_Step(a.w, 1.0f / 60, 4);
        b2World_Step(b.w, 1.0f / 60, 4);
        observe_scene(&a, step);
        observe_scene(&b, step);
    }

    b2DestroyWorld(a.w);
    b2DestroyWorld(b.w);
    ORACLE_CLOSE();
}

static void collection_observation(b2BodyId b) {
    b2ShapeId     shapes[16];
    b2JointId     joints[16];
    b2ContactData data[16];
    int           n = b2Body_GetShapes(b, shapes, 16);

    ORACLE_I32("shapes", n);

    for (int i = 0; i < n; ++i) { ORACLE_U64("shape", shape_key(shapes[i])); }

    n = b2Body_GetJoints(b, joints, 16);

    ORACLE_I32("joints", n);

    for (int i = 0; i < n; ++i) {
        ORACLE_U64("joint", b2StoreJointId(joints[i]) & UINT64_C(0xffffffff0000ffff));
    }

    ORACLE_I32("contact.capacity", b2Body_GetContactCapacity(b));

    n = b2Body_GetContactData(b, data, 16);

    ORACLE_I32("contacts", n);

    for (int i = 0; i < n; ++i) {
        ORACLE_U32("contact.index", (uint32_t)data[i].contactId.index1);
        ORACLE_U32("contact.generation", data[i].contactId.generation);
        ORACLE_U64("contact.a", shape_key(data[i].shapeIdA));
        ORACLE_U64("contact.b", shape_key(data[i].shapeIdB));
        ORACLE_VEC2("contact.normal", data[i].manifold.normal);
        ORACLE_I32("contact.points", data[i].manifold.pointCount);
    }
}

static void collections(void) {
    ORACLE_OPEN("lifecycle/CollectionTransitions");

    b2WorldDef wd    = b2DefaultWorldDef();
    wd.gravity       = (b2Vec2){0, 0};
    b2WorldId w      = b2CreateWorld(&wd);
    b2BodyDef bd     = b2DefaultBodyDef();
    b2BodyId  ground = b2CreateBody(w, &bd);
    bd.type          = b2_dynamicBody;
    b2BodyId b       = b2CreateBody(w, &bd);

    collection_observation(b);

    b2ShapeDef sd     = b2DefaultShapeDef();
    b2Circle   circle = {{0, 0}, 2};

    b2CreateCircleShape(ground, &sd, &circle);

    circle.radius = 0.5f;
    b2ShapeId a   = b2CreateCircleShape(b, &sd, &circle);

    collection_observation(b);

    circle.center.x  = 0.2f;
    b2ShapeId second = b2CreateCircleShape(b, &sd, &circle);

    collection_observation(b);
    b2World_Step(w, 1.0f / 60, 4);
    collection_observation(b);

    b2ContactData data[16];
    int           n = b2Body_GetContactData(b, data, 16);
    if (n != 2) { abort(); }

    b2ContactId held[2] = {data[0].contactId, data[1].contactId};

    b2DestroyShape(second, true);
    b2World_Step(w, 1.0f / 60, 4);
    collection_observation(b);

    for (int i = 0; i < 2; ++i) { ORACLE_BOOL("held.contact.valid", b2Contact_IsValid(held[i])); }

    bd.position    = (b2Pos){0, 5};
    b2BodyId  peer = b2CreateBody(w, &bd);
    b2JointId j1   = make_joint(w, b, peer);

    collection_observation(b);

    b2JointId j2 = make_joint(w, b, peer);

    collection_observation(b);
    b2DestroyJoint(j1);
    collection_observation(b);
    ORACLE_BOOL("removed.joint.valid", b2Joint_IsValid(j1));
    ORACLE_BOOL("survivor.joint.valid", b2Joint_IsValid(j2));
    b2DestroyBody(b);
    ORACLE_BOOL("dependent.shape.valid", b2Shape_IsValid(a));
    ORACLE_BOOL("dependent.joint.valid", b2Joint_IsValid(j2));
    b2World_Step(w, 1.0f / 60, 4);
    collection_observation(ground);
    collection_observation(peer);

    for (int i = 0; i < 2; ++i) { ORACLE_BOOL("held.contact.valid", b2Contact_IsValid(held[i])); }

    b2DestroyWorld(w);
    ORACLE_CLOSE();
}

int main(void) {
    null_handles();
    lifetimes();
    collections();
    ownership();

    return 0;
}
