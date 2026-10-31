// SPDX-License-Identifier: MIT
#pragma once
#include "box2d/box2d.h"
#include "oracle.h"

#include <stdint.h>
#include <stdio.h>

// Hash the defined identity fields, never padding or host world-slot numbers.
// Generation and world ownership remain observed, including destroyed handles.
static uint64_t
replay_event_id(uint64_t hash, int index, unsigned generation, unsigned world0, b2WorldId world) {
    hash = oracle_hash_u32(hash, (uint32_t)index);
    hash = oracle_hash_u32(hash, generation);

    return oracle_hash_u32(hash, world0 == (unsigned)(world.index1 - 1));
}

#define EVENT_ID(h, id, w) replay_event_id(h, (id).index1, (id).generation, (id).world0, w)

typedef struct replay_event_row {
    uint64_t fields[11];
} replay_event_row;

static replay_event_row replay_event_observe(b2WorldId world) {
    replay_event_row row;
    uint64_t         hash   = ORACLE_HASH_INIT;
    b2BodyEvents     bodies = b2World_GetBodyEvents(world);
    row.fields[0]           = (uint64_t)bodies.moveCount;

    for (int i = 0; i < bodies.moveCount; ++i) {
        b2BodyMoveEvent e = bodies.moveEvents[i];
        hash              = EVENT_ID(hash, e.bodyId, world);
        hash              = oracle_hash_transform(hash, e.transform.p, e.transform.q);
        hash              = oracle_hash_u32(hash, e.fellAsleep);
    }

    row.fields[1]          = hash;
    b2SensorEvents sensors = b2World_GetSensorEvents(world);
    hash                   = ORACLE_HASH_INIT;
    row.fields[2]          = (uint64_t)sensors.beginCount;
    row.fields[3]          = (uint64_t)sensors.endCount;

    for (int i = 0; i < sensors.beginCount; ++i) {
        b2SensorBeginTouchEvent e = sensors.beginEvents[i];
        hash                      = EVENT_ID(hash, e.sensorShapeId, world);
        hash                      = EVENT_ID(hash, e.visitorShapeId, world);
    }

    for (int i = 0; i < sensors.endCount; ++i) {
        b2SensorEndTouchEvent e = sensors.endEvents[i];
        hash                    = EVENT_ID(hash, e.sensorShapeId, world);
        hash                    = EVENT_ID(hash, e.visitorShapeId, world);
    }

    row.fields[4]            = hash;
    b2ContactEvents contacts = b2World_GetContactEvents(world);
    hash                     = ORACLE_HASH_INIT;
    row.fields[5]            = (uint64_t)contacts.beginCount;
    row.fields[6]            = (uint64_t)contacts.endCount;
    row.fields[7]            = (uint64_t)contacts.hitCount;

    for (int i = 0; i < contacts.beginCount; ++i) {
        b2ContactBeginTouchEvent e = contacts.beginEvents[i];
        hash                       = EVENT_ID(hash, e.shapeIdA, world);
        hash                       = EVENT_ID(hash, e.shapeIdB, world);
        hash                       = EVENT_ID(hash, e.contactId, world);
    }

    for (int i = 0; i < contacts.endCount; ++i) {
        b2ContactEndTouchEvent e = contacts.endEvents[i];
        hash                     = EVENT_ID(hash, e.shapeIdA, world);
        hash                     = EVENT_ID(hash, e.shapeIdB, world);
        hash                     = EVENT_ID(hash, e.contactId, world);
    }

    for (int i = 0; i < contacts.hitCount; ++i) {
        b2ContactHitEvent e = contacts.hitEvents[i];
        hash                = EVENT_ID(hash, e.shapeIdA, world);
        hash                = EVENT_ID(hash, e.shapeIdB, world);
        hash                = EVENT_ID(hash, e.contactId, world);
        hash                = oracle_hash_pos(hash, e.point);
        hash                = oracle_hash_f32(hash, e.normal.x);
        hash                = oracle_hash_f32(hash, e.normal.y);
        hash                = oracle_hash_f32(hash, e.approachSpeed);
    }

    row.fields[8]        = hash;
    b2JointEvents joints = b2World_GetJointEvents(world);
    hash                 = ORACLE_HASH_INIT;
    row.fields[9]        = (uint64_t)joints.count;

    for (int i = 0; i < joints.count; ++i) {
        hash = EVENT_ID(hash, joints.jointEvents[i].jointId, world);
    }

    row.fields[10] = hash;
    return row;
}

static inline void replay_events(b2WorldId world, const char* prefix) {
    char owned[256];

    snprintf(owned, sizeof(owned), "%s", prefix);

    prefix               = owned;
    replay_event_row row = replay_event_observe(world);

    ORACLE_I32(ORACLE_KEY("%s/events/moveCount", prefix), (int32_t)row.fields[0]);
    ORACLE_U64(ORACLE_KEY("%s/events/moveHash", prefix), row.fields[1]);
    ORACLE_I32(ORACLE_KEY("%s/events/sensorBeginCount", prefix), (int32_t)row.fields[2]);
    ORACLE_I32(ORACLE_KEY("%s/events/sensorEndCount", prefix), (int32_t)row.fields[3]);
    ORACLE_U64(ORACLE_KEY("%s/events/sensorHash", prefix), row.fields[4]);
    ORACLE_I32(ORACLE_KEY("%s/events/contactBeginCount", prefix), (int32_t)row.fields[5]);
    ORACLE_I32(ORACLE_KEY("%s/events/contactEndCount", prefix), (int32_t)row.fields[6]);
    ORACLE_I32(ORACLE_KEY("%s/events/hitCount", prefix), (int32_t)row.fields[7]);
    ORACLE_U64(ORACLE_KEY("%s/events/contactHash", prefix), row.fields[8]);
    ORACLE_I32(ORACLE_KEY("%s/events/jointCount", prefix), (int32_t)row.fields[9]);
    ORACLE_U64(ORACLE_KEY("%s/events/jointHash", prefix), row.fields[10]);
}

#undef EVENT_ID
