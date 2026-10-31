// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

#include "body.h"
#include "box2d/box2d.h"
#include "physics_world.h"

#include <stdint.h>

#if defined(RESTITUTION_NO_ORACLE)
#define ORACLE_RESTITUTION_STEP(worldId) ((void)0)
#define ORACLE_RESTITUTION_END(worldId) ((void)0)
#define ORACLE_BOOL(...) ((void)0)
#define ORACLE_F32(...) ((void)0)
#define ORACLE_I32(...) ((void)0)
#define ORACLE_U64(...) ((void)0)
#else
#include "dynamics_observe.h"
#include "oracle.h"

static uint8_t restitution_steps[4000 * 8];
static uint8_t restitution_transform_steps[4000 * 8];
static size_t  restitution_step_count;
static uint8_t restitution_full_states[67 * (4 + 64 * 28)];
static size_t  restitution_full_size;

static void restitution_pack(uint8_t* data, size_t* size, uint64_t value, int width) {
    if (*size + (size_t)width > sizeof(restitution_full_states)) { abort(); }
    for (int i = 0; i < width; ++i) {
        data[*size]  = (uint8_t)(value >> (8 * i));
        *size       += 1;
    }
}

static void restitution_step(b2WorldId worldId) {
    char observation[128];

    snprintf(observation, sizeof(observation), "observation/%zu/", restitution_step_count + 1);
    dynamics_trace_add(dynamics_observe(
        worldId,
        observation,
        restitution_step_count == 0 || (restitution_step_count + 1) % 60 == 0
    ));

    b2World* world         = b2GetWorldFromId(worldId);
    uint64_t hash          = ORACLE_HASH_INIT;
    uint64_t transformHash = ORACLE_HASH_INIT;
    bool     sample        = (restitution_step_count + 1) % 60 == 0;
    size_t   countOffset   = restitution_full_size;

    if (sample) { restitution_full_size += 4; }

    int bodyCount = 0;

    for (int i = 0; i < world->bodies.count; ++i) {
        b2Body* body = b2Array_Get(world->bodies, i);
        if (body->id == B2_NULL_INDEX) { continue; }

        b2BodyId id       = b2MakeBodyId(world, i);
        b2Pos    position = b2Body_GetPosition(id);
        b2Rot    rotation = b2Body_GetRotation(id);
        hash              = oracle_hash_transform(hash, position, rotation);
        transformHash     = oracle_hash_transform(transformHash, position, rotation);

        if (sample) {
            if (bodyCount >= 64) { abort(); }
#if defined(BOX2D_DOUBLE_PRECISION)

            int      positionWidth = 8;
            uint64_t values[4]     = {
                oracle_bits_f64(position.x),
                oracle_bits_f64(position.y),
                oracle_bits_f32(rotation.c),
                oracle_bits_f32(rotation.s)
            };
#else
            int      positionWidth = 4;
            uint64_t values[4]     = {
                oracle_bits_f32(position.x),
                oracle_bits_f32(position.y),
                oracle_bits_f32(rotation.c),
                oracle_bits_f32(rotation.s)
            };
#endif

            restitution_pack(restitution_full_states, &restitution_full_size, (uint32_t)i, 4);

            for (int component = 0; component < 4; ++component) {
                restitution_pack(
                    restitution_full_states,
                    &restitution_full_size,
                    values[component],
                    component < 2 ? positionWidth : 4
                );
            }

            bodyCount += 1;
        }

        b2Vec2 v = b2Body_GetLinearVelocity(id);
        hash     = oracle_hash_f32(hash, v.x);
        hash     = oracle_hash_f32(hash, v.y);
        hash     = oracle_hash_f32(hash, b2Body_GetAngularVelocity(id));
        b2ContactData contacts[64];
        int           count = b2Body_GetContactData(id, contacts, 64);
        hash                = oracle_hash_u32(hash, (uint32_t)count);

        for (int c = 0; c < count; ++c) {
            const b2Manifold* m = &contacts[c].manifold;
            hash                = oracle_hash_u32(hash, (uint32_t)m->pointCount);

            for (int p = 0; p < m->pointCount; ++p) {
                hash = oracle_hash_f32(hash, m->points[p].separation);
                hash = oracle_hash_f32(hash, m->points[p].normalVelocity);
                hash = oracle_hash_f32(hash, m->points[p].totalNormalImpulse);
            }
        }
    }

    if (sample) {
        if (restitution_full_size > sizeof(restitution_full_states)) { abort(); }

        restitution_pack(restitution_full_states, &countOffset, (uint32_t)bodyCount, 4);
    }

    if (restitution_step_count >= 4000) { abort(); }
    for (int b = 0; b < 8; ++b) {
        restitution_steps[8 * restitution_step_count + (size_t)b] = (uint8_t)(hash >> (8 * b));
        restitution_transform_steps[8 * restitution_step_count + (size_t)b] =
            (uint8_t)(transformHash >> (8 * b));
    }

    restitution_step_count += 1;
}

static void restitution_end(b2WorldId worldId) {
    (void)worldId;
    dynamics_trace_end();
    ORACLE_BYTES("transforms/steps", restitution_transform_steps, 8 * restitution_step_count);
    ORACLE_BYTES("transforms/full", restitution_full_states, restitution_full_size);
    ORACLE_BYTES("steps", restitution_steps, 8 * restitution_step_count);

    restitution_step_count = 0;
    restitution_full_size  = 0;
}

#define ORACLE_RESTITUTION_STEP(worldId) restitution_step(worldId)
#define ORACLE_RESTITUTION_END(worldId) restitution_end(worldId)
#endif

#include "upstream/test_restitution.c"

int main(void) {
#if defined(RESTITUTION_NO_ORACLE)
    return RestitutionTest();
#else
    const char* names[] = {
        "HeadOnTest",
        "PhaseTest",
        "SupportedTest",
        "TwoPointTest",
        "SpinTest",
        "DropTest",
        "SingleBoxTest",
        "ThresholdTest",
        "RestingTest",
        "NormalVelocityTest",
        "OvershootTest",
        "EnergyTest",
        "ImpulseTest",
        "WorkerParityTest"
    };

    int (*tests[])(void) = {
        HeadOnTest,
        PhaseTest,
        SupportedTest,
        TwoPointTest,
        SpinTest,
        DropTest,
        SingleBoxTest,
        ThresholdTest,
        RestingTest,
        NormalVelocityTest,
        OvershootTest,
        EnergyTest,
        ImpulseTest,
        WorkerParityTest
    };

    int failed = 0;

    for (int i = 0; i < 14; ++i) {
        ORACLE_OPEN(ORACLE_KEY("restitution/%s", names[i]));

        failed |= tests[i]();

        ORACLE_CLOSE();
    }

    return failed;
#endif
}
