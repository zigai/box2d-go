// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT
#include "contact.h"
#include "joint.h"
#include "shape.h"

static uint64_t scene_digest;

static void scene_hash_bytes(const uint8_t* p, size_t n) {
    for (size_t i = 0; i < n; ++i) {
        scene_digest ^= p[i];
        scene_digest *= UINT64_C(1099511628211);
    }
}

static void scene_hash_u32(uint32_t v) {
    uint8_t p[4];

    scene_u32(p, v);
    scene_hash_bytes(p, 4);
}

static void scene_hash_f32(float v) {
    uint32_t u;

    memcpy(&u, &v, 4);
    scene_hash_u32(u);
}

static void scene_hash_vec(b2Vec2 v) {
    scene_hash_f32(v.x);
    scene_hash_f32(v.y);
}

static void scene_hash_frame(b2Transform v) {
    scene_hash_vec(v.p);
    scene_hash_f32(v.q.c);
    scene_hash_f32(v.q.s);
}

static void scene_hash_id(int index1, uint32_t generation) {
    scene_hash_u32((uint32_t)index1);
    scene_hash_u32(generation);
}

static void scene_hash_manifold(b2Manifold m, bool hitEvents) {
    scene_hash_u32(hitEvents);
    scene_hash_vec(m.normal);
    scene_hash_f32(m.rollingImpulse);
    scene_hash_u32((uint32_t)m.pointCount);

    for (int i = 0; i < m.pointCount; ++i) {
        b2ManifoldPoint p = m.points[i];

        scene_hash_vec(p.anchorA);
        scene_hash_vec(p.anchorB);
        scene_hash_f32(p.separation);
        scene_hash_f32(p.normalImpulse);
        scene_hash_f32(p.tangentImpulse);
        scene_hash_f32(p.totalNormalImpulse);
        scene_hash_f32(p.baseSeparation);
        scene_hash_u32(p.id);
        scene_hash_u32(p.persisted);

        if (hitEvents) { scene_hash_f32(p.normalVelocity); }
    }
}

static void scene_observation(int step, const char* lane, int count) {
    char key[64];

    snprintf(key, sizeof(key), "%d.%s.count", step, lane);
    ORACLE_I32(key, count);
    snprintf(key, sizeof(key), "%d.%s.hash", step, lane);
    ORACLE_U64(key, scene_digest);
}

// Event records use explicit little-endian fields, sorted as a multiset. Sorting
// removes scheduling order only, preserving payloads, generations and duplicates.
static size_t scene_event_width;

static int scene_event_compare(const void* a, const void* b) {
    return memcmp(a, b, scene_event_width);
}

static void
scene_event_lane(int step, const char* lane, uint8_t* records, int count, size_t width) {
    scene_event_width = width;

    if (count > 1) { qsort(records, (size_t)count, width, scene_event_compare); }

    scene_digest = UINT64_C(14695981039346656037);

    scene_hash_bytes(records, (size_t)count * width);
    scene_observation(step, lane, count);
    free(records);
}

static void scene_event_id(uint8_t* p, int index1, uint32_t generation) {
    scene_u32(p, (uint32_t)index1);
    scene_u32(p + 4, generation);
}

static void scene_event_float(uint8_t* p, float v) {
    uint32_t u;

    memcpy(&u, &v, 4);
    scene_u32(p, u);
}

static void scene_event_pos(uint8_t* p, b2Pos v) {
    b2WorldTransform t = {.p = v, .q = {0, 0}};

    scene_transform(p, t);
}

static uint8_t* scene_event_allocate(int count, size_t width) {
    uint8_t* p = calloc(count ? (size_t)count : 1, width);
    if (!p) { abort(); }

    return p;
}

static void scene_state(b2WorldId worldId, int step) {
    b2World* world = b2GetWorldFromId(worldId);
    int      count = 0;
    scene_digest   = UINT64_C(14695981039346656037);

    for (int i = 0; i < world->bodies.count; ++i) {
        b2Body* b = world->bodies.data + i;
        if (b->setIndex == B2_NULL_INDEX) { continue; }

        b2BodyId id = b2MakeBodyId(world, i);

        scene_hash_id(id.index1, id.generation);
        scene_hash_u32(b2Body_GetType(id));
        scene_hash_vec(b2Body_GetLinearVelocity(id));
        scene_hash_f32(b2Body_GetAngularVelocity(id));
        scene_hash_u32(b2Body_IsAwake(id));
        scene_hash_u32(b2Body_IsEnabled(id));

        b2MassData m = b2Body_GetMassData(id);

        scene_hash_f32(m.mass);
        scene_hash_vec(m.center);
        scene_hash_f32(m.rotationalInertia);

        ++count;
    }

    scene_observation(step, "dynamic", count);

    count        = 0;
    scene_digest = UINT64_C(14695981039346656037);

    for (int i = 0; i < world->contacts.count; ++i) {
        b2Contact* c = world->contacts.data + i;
        if (c->contactId == B2_NULL_INDEX) { continue; }

        scene_hash_id(i + 1, c->generation);
        scene_hash_id(c->shapeIdA + 1, world->shapes.data[c->shapeIdA].generation);
        scene_hash_id(c->shapeIdB + 1, world->shapes.data[c->shapeIdB].generation);
        scene_hash_manifold(
            b2GetContactSim(world, c)->manifold,
            world->shapes.data[c->shapeIdA].enableHitEvents ||
                world->shapes.data[c->shapeIdB].enableHitEvents
        );

        ++count;
    }

    scene_observation(step, "contact", count);

    count        = 0;
    scene_digest = UINT64_C(14695981039346656037);

    for (int i = 0; i < world->joints.count; ++i) {
        b2Joint* j = world->joints.data + i;
        if (j->setIndex == B2_NULL_INDEX) { continue; }

        b2JointId id = {i + 1, world->worldId, j->generation};

        scene_hash_id(id.index1, id.generation);
        scene_hash_u32(b2Joint_GetType(id));

        b2BodyId a = b2Joint_GetBodyA(id), b = b2Joint_GetBodyB(id);

        scene_hash_id(a.index1, a.generation);
        scene_hash_id(b.index1, b.generation);
        scene_hash_frame(b2Joint_GetLocalFrameA(id));
        scene_hash_frame(b2Joint_GetLocalFrameB(id));
        scene_hash_vec(b2Joint_GetConstraintForce(id));
        scene_hash_f32(b2Joint_GetConstraintTorque(id));

        if (b2Joint_GetType(id) == b2_revoluteJoint) {
            scene_hash_u32(b2RevoluteJoint_IsMotorEnabled(id));
            scene_hash_f32(b2RevoluteJoint_GetMotorSpeed(id));
            scene_hash_f32(b2RevoluteJoint_GetMaxMotorTorque(id));
            scene_hash_f32(b2RevoluteJoint_GetMotorTorque(id));
            scene_hash_u32(b2RevoluteJoint_IsSpringEnabled(id));
            scene_hash_f32(b2RevoluteJoint_GetSpringHertz(id));
            scene_hash_f32(b2RevoluteJoint_GetSpringDampingRatio(id));
            scene_hash_u32(b2RevoluteJoint_IsLimitEnabled(id));
            scene_hash_f32(b2RevoluteJoint_GetLowerLimit(id));
            scene_hash_f32(b2RevoluteJoint_GetUpperLimit(id));
        }

        ++count;
    }

    scene_observation(step, "joint", count);

    b2BodyEvents moves   = b2World_GetBodyEvents(worldId);
    size_t       width   = 8 + sizeof(b2Pos) + 8 + 8;
    uint8_t*     records = scene_event_allocate(moves.moveCount, width);

    for (int i = 0; i < moves.moveCount; ++i) {
        b2BodyMoveEvent e = moves.moveEvents[i];
        uint8_t*        p = records + (size_t)i * width;

        scene_event_id(p, e.bodyId.index1, e.bodyId.generation);
        scene_transform(p + 8, e.transform);
        scene_u32(p + width - 8, e.fellAsleep);
        scene_u32(p + width - 4, e.userData == NULL);
    }

    scene_event_lane(step, "move", records, moves.moveCount, width);

    b2SensorEvents sensors = b2World_GetSensorEvents(worldId);
    width                  = 16;
    records                = scene_event_allocate(sensors.beginCount, width);

    for (int i = 0; i < sensors.beginCount; ++i) {
        b2SensorBeginTouchEvent e = sensors.beginEvents[i];
        uint8_t*                p = records + (size_t)i * width;

        scene_event_id(p, e.sensorShapeId.index1, e.sensorShapeId.generation);
        scene_event_id(p + 8, e.visitorShapeId.index1, e.visitorShapeId.generation);
    }

    scene_event_lane(step, "sensorBegin", records, sensors.beginCount, width);

    records = scene_event_allocate(sensors.endCount, width);

    for (int i = 0; i < sensors.endCount; ++i) {
        b2SensorEndTouchEvent e = sensors.endEvents[i];
        uint8_t*              p = records + (size_t)i * width;

        scene_event_id(p, e.sensorShapeId.index1, e.sensorShapeId.generation);
        scene_event_id(p + 8, e.visitorShapeId.index1, e.visitorShapeId.generation);
    }

    scene_event_lane(step, "sensorEnd", records, sensors.endCount, width);

    b2ContactEvents contacts = b2World_GetContactEvents(worldId);
    width                    = 24;
    records                  = scene_event_allocate(contacts.beginCount, width);

    for (int i = 0; i < contacts.beginCount; ++i) {
        b2ContactBeginTouchEvent e = contacts.beginEvents[i];
        uint8_t*                 p = records + (size_t)i * width;

        scene_event_id(p, e.shapeIdA.index1, e.shapeIdA.generation);
        scene_event_id(p + 8, e.shapeIdB.index1, e.shapeIdB.generation);
        scene_event_id(p + 16, e.contactId.index1, e.contactId.generation);
    }

    scene_event_lane(step, "contactBegin", records, contacts.beginCount, width);

    records = scene_event_allocate(contacts.endCount, width);

    for (int i = 0; i < contacts.endCount; ++i) {
        b2ContactEndTouchEvent e = contacts.endEvents[i];
        uint8_t*               p = records + (size_t)i * width;

        scene_event_id(p, e.shapeIdA.index1, e.shapeIdA.generation);
        scene_event_id(p + 8, e.shapeIdB.index1, e.shapeIdB.generation);
        scene_event_id(p + 16, e.contactId.index1, e.contactId.generation);
    }

    scene_event_lane(step, "contactEnd", records, contacts.endCount, width);

    width   = 24 + sizeof(b2Pos) + 12;
    records = scene_event_allocate(contacts.hitCount, width);

    for (int i = 0; i < contacts.hitCount; ++i) {
        b2ContactHitEvent e = contacts.hitEvents[i];
        uint8_t*          p = records + (size_t)i * width;

        scene_event_id(p, e.shapeIdA.index1, e.shapeIdA.generation);
        scene_event_id(p + 8, e.shapeIdB.index1, e.shapeIdB.generation);
        scene_event_id(p + 16, e.contactId.index1, e.contactId.generation);
        scene_event_pos(p + 24, e.point);
        scene_event_float(p + 24 + sizeof(b2Pos), e.normal.x);
        scene_event_float(p + 28 + sizeof(b2Pos), e.normal.y);
        scene_event_float(p + 32 + sizeof(b2Pos), e.approachSpeed);
    }

    scene_event_lane(step, "hit", records, contacts.hitCount, width);

    b2JointEvents joints = b2World_GetJointEvents(worldId);
    width                = 12;
    records              = scene_event_allocate(joints.count, width);

    for (int i = 0; i < joints.count; ++i) {
        b2JointEvent e = joints.jointEvents[i];

        scene_event_id(records + (size_t)i * width, e.jointId.index1, e.jointId.generation);
        scene_u32(records + (size_t)i * width + 8, e.userData == NULL);
    }

    scene_event_lane(step, "jointEvent", records, joints.count, width);
}
