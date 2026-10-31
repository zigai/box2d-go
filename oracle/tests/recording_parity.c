// SPDX-License-Identifier: MIT
#define _POSIX_C_SOURCE 200809L
#include "box2d/box2d.h"
#include "oracle.h"
#include "recording.h"
#include "replay.h"
#include "replay_observe.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define REQUIRE(x)                                                                                 \
    do {                                                                                           \
        if (!(x)) {                                                                                \
            fprintf(stderr, "%s:%d: %s\n", __FILE__, __LINE__, #x);                                \
            abort();                                                                               \
        }                                                                                          \
    } while (0)

static bool all_overlap(b2ShapeId id, void* context) {
    (void)id;
    (void)context;

    return true;
}

static b2Recording* lifecycle_source(int steps) {
    b2WorldDef wd     = b2DefaultWorldDef();
    wd.gravity        = (b2Vec2){0};
    wd.workerCount    = 1;
    b2WorldId w       = b2CreateWorld(&wd);
    b2BodyDef bd      = b2DefaultBodyDef();
    bd.type           = b2_dynamicBody;
    bd.name           = "seed";
    bd.position       = (b2Pos){1, 2};
    b2BodyId   seed   = b2CreateBody(w, &bd);
    b2ShapeDef sd     = b2DefaultShapeDef();
    b2Circle   circle = {{0, 0}, 0.5f};

    b2CreateCircleShape(seed, &sd, &circle);

    b2Recording* rec = b2CreateRecording(0);

    b2World_StartRecording(w, rec);

    for (int frame = 0; frame < steps; ++frame) {
        if (frame == 1) {
            b2DestroyBody(seed);

            bd.name       = "reused";
            bd.position.x = 3;
            seed          = b2CreateBody(w, &bd);

            b2CreateCircleShape(seed, &sd, &circle);
        }

        if (frame == 2) {
            bd.name       = "future";
            bd.position.x = 5;
            b2BodyId body = b2CreateBody(w, &bd);

            b2CreateCircleShape(body, &sd, &circle);
        }

        b2World_Step(w, 1.0f / 30.0f, 2);

        b2QueryFilter filter = b2DefaultQueryFilter();
        filter.categoryBits  = 3;

        b2World_OverlapAABB(
            w,
            (b2Pos){2, -1},
            (b2AABB){{-3, -3}, {6, 5}},
            filter,
            all_overlap,
            NULL
        );

        filter.maskBits = 0;

        b2World_OverlapAABB(
            w,
            (b2Pos){2, -1},
            (b2AABB){{-3, -3}, {6, 5}},
            filter,
            all_overlap,
            NULL
        );
    }

    b2World_StopRecording(w);
    b2Body_SetName(seed, "not recorded");
    b2DestroyWorld(w);

    return rec;
}

static void lifecycle_parity(void) {
    ORACLE_OPEN("recording/LifecycleParity");

    float oldScale = b2GetLengthUnitsPerMeter();

    b2SetLengthUnitsPerMeter(2);

    int counts[] = {0, 1, 16}, workers[] = {0, 4};

    for (int n = 0; n < 3; ++n) {
        int          steps = counts[n];
        b2Recording* rec   = lifecycle_source(steps);

        for (int wc = 0; wc < 2; ++wc) {
            b2Replay* p =
                b2CreateReplay(b2Recording_GetData(rec), b2Recording_GetSize(rec), workers[wc]);

            REQUIRE(p);

            char prefix[80];

            snprintf(prefix, sizeof(prefix), "n%d/w%d", steps, workers[wc]);
            replay_observe(p, ORACLE_KEY("%s/open", prefix));

            for (int action = 0; action < steps + 3; ++action) {
                ORACLE_BOOL(ORACLE_KEY("%s/step%d/return", prefix, action), b2Replay_StepFrame(p));
                replay_observe(p, ORACLE_KEY("%s/step%d", prefix, action));
            }

            b2WorldId world = b2Replay_GetWorldId(p);

            b2Replay_Restart(p);
            ORACLE_BOOL(
                ORACLE_KEY("%s/restart/identity", prefix),
                b2StoreWorldId(world) == b2StoreWorldId(b2Replay_GetWorldId(p))
            );
            replay_observe(p, ORACLE_KEY("%s/restart", prefix));
            b2Replay_SetKeyframePolicy(p, 1u << 26, 2);
            ORACLE_U64(ORACLE_KEY("%s/policy/budget", prefix), b2Replay_GetKeyframeBudget(p));
            ORACLE_I32(ORACLE_KEY("%s/policy/min", prefix), b2Replay_GetKeyframeMinInterval(p));
            ORACLE_I32(ORACLE_KEY("%s/policy/interval", prefix), b2Replay_GetKeyframeInterval(p));
            ORACLE_BOOL(ORACLE_KEY("%s/policy/empty", prefix), b2Replay_GetKeyframeBytes(p) == 0);
            b2Replay_SetKeyframePolicy(p, 0, -1);
            ORACLE_U64(
                ORACLE_KEY("%s/policy/retainedBudget", prefix),
                b2Replay_GetKeyframeBudget(p)
            );
            ORACLE_I32(
                ORACLE_KEY("%s/policy/retainedMin", prefix),
                b2Replay_GetKeyframeMinInterval(p)
            );

            int targets[] = {steps + 8, -3, 2, 2, steps, 1, 0, steps + 1};

            for (int i = 0; i < 8; ++i) {
                b2Replay_SeekFrame(p, targets[i]);
                replay_observe(p, ORACLE_KEY("%s/seek%d", prefix, i));
            }

            ORACLE_BOOL(
                ORACLE_KEY("%s/policy/populated", prefix),
                b2Replay_GetKeyframeBytes(p) > 0
            );
            ORACLE_BOOL(
                ORACLE_KEY("%s/policy/bounded", prefix),
                b2Replay_GetKeyframeBytes(p) <= b2Replay_GetKeyframeBudget(p)
            );
            b2Replay_SetKeyframePolicy(p, 1, 1);
            b2Replay_Restart(p);

            while (b2Replay_StepFrame(p)) {}

            ORACLE_BOOL(
                ORACLE_KEY("%s/tight/bounded", prefix),
                b2Replay_GetKeyframeBytes(p) <= b2Replay_GetKeyframeBudget(p)
            );
            ORACLE_BOOL(
                ORACLE_KEY("%s/tight/softBudget", prefix),
                b2Replay_GetKeyframeBytes(p) <= b2Replay_GetKeyframeBudget(p) ||
                    p->keyframeCount == 1
            );
            ORACLE_BOOL(
                ORACLE_KEY("%s/tight/coarsened", prefix),
                b2Replay_GetKeyframeInterval(p) > b2Replay_GetKeyframeMinInterval(p)
            );
            b2Replay_SeekFrame(p, 3);
            replay_observe(p, ORACLE_KEY("%s/tight/seek", prefix));
            b2DestroyReplay(p);
            ORACLE_BOOL(ORACLE_KEY("%s/destroy/worldValid", prefix), b2World_IsValid(world));
        }

        b2DestroyRecording(rec);
    }

    b2SetLengthUnitsPerMeter(oldScale);
    ORACLE_CLOSE();
}

static void owned_source_parity(void) {
    ORACLE_OPEN("recording/OwnedSourceParity");

    b2Recording* rec    = lifecycle_source(16);
    int          size   = b2Recording_GetSize(rec);
    uint8_t*     source = malloc((size_t)size);

    REQUIRE(source);
    memcpy(source, b2Recording_GetData(rec), (size_t)size);

    b2Replay* p = b2CreateReplay(source, size, 0);

    REQUIRE(p);
    memset(source, 0xA5, (size_t)size);
    free(source);
    b2DestroyRecording(rec);
    replay_observe(p, "open");

    while (b2Replay_StepFrame(p)) { replay_observe(p, ORACLE_KEY("step%d", b2Replay_GetFrame(p))); }

    b2Replay_Restart(p);
    replay_observe(p, "restart");

    int targets[] = {16, 1, 0, 7, 99};

    for (int i = 0; i < 5; ++i) {
        b2Replay_SeekFrame(p, targets[i]);
        replay_observe(p, ORACLE_KEY("seek%d", i));
    }

    b2WorldId w = b2Replay_GetWorldId(p);

    b2DestroyReplay(p);
    ORACLE_BOOL("destroy/worldValid", b2World_IsValid(w));
    ORACLE_CLOSE();
}

static uint8_t* mismatch(const uint8_t* source, int size, bool query) {
    uint8_t* data = malloc((size_t)size);

    REQUIRE(data);
    memcpy(data, source, (size_t)size);

    b2RecHeader header;

    memcpy(&header, data, sizeof(header));

    int cursor = (int)sizeof(header) + (int)header.snapshotSize, frame = 0, changed = 0;
    while (cursor <= size - 4) {
        uint8_t op    = data[cursor];
        int     len   = data[cursor + 1] | data[cursor + 2] << 8 | data[cursor + 3] << 16;
        int     start = cursor + 4;

        REQUIRE(len <= size - start);

        if (op == 0x80) { ++frame; }
        if (frame == 7 || frame == 12) {
            b2RecReader rdr = {0};
            rdr.data        = data;
            rdr.size        = start + len;
            rdr.cursor      = start;
            rdr.ok          = true;

            if (!query && op == 0xf1) {
                (void)b2RecR_WORLDID(&rdr);

                int      at   = rdr.cursor;
                uint64_t hash = b2RecR_U64(&rdr);

                REQUIRE(rdr.ok);

                hash ^= 1;
                for (int i = 0; i < 8; ++i) { data[at + i] = (uint8_t)(hash >> (8 * i)); }

                ++changed;
            }

            if (query && op == 0xe0) {
                (void)b2RecR_WORLDID(&rdr);
                (void)b2RecR_POSITION(&rdr);
                (void)b2RecR_AABB(&rdr);
                (void)b2RecR_QUERYFILTER(&rdr);

                uint32_t count = b2RecR_U32(&rdr);
                if (count > 0) {
                    int       at = rdr.cursor;
                    b2ShapeId id = b2RecR_SHAPEID(&rdr);

                    REQUIRE(rdr.ok);

                    id.generation   ^= 1;
                    uint64_t packed  = b2StoreShapeId(id);
                    for (int i = 0; i < 8; ++i) { data[at + i] = (uint8_t)(packed >> (8 * i)); }

                    ++changed;
                }
            }
        }

        cursor = start + len;
    }

    REQUIRE(changed == 2);
    return data;
}

static void divergence_parity(void) {
    ORACLE_OPEN("recording/DivergenceParity");

    for (int kind = 0; kind < 2; ++kind) {
        const char*  name = kind == 0 ? "state" : "query";
        b2Recording* rec  = lifecycle_source(16);
        int          size = b2Recording_GetSize(rec);
        uint8_t*     data = mismatch(b2Recording_GetData(rec), size, kind == 1);

        b2DestroyRecording(rec);

        bool valid = b2ValidateReplay(data, size, 0);

        REQUIRE(!valid);
        ORACLE_BOOL(ORACLE_KEY("%s/validate", name), valid);

        b2Replay* p = b2CreateReplay(data, size, 0);

        REQUIRE(p);
        free(data);
        b2Replay_SetKeyframePolicy(p, 1u << 26, 2);
        replay_observe(p, ORACLE_KEY("%s/open", name));

        while (b2Replay_StepFrame(p)) {
            replay_observe(p, ORACLE_KEY("%s/step%d", name, b2Replay_GetFrame(p)));
        }

        REQUIRE(b2Replay_HasDiverged(p));
        REQUIRE(b2Replay_GetDivergeFrame(p) == 7);
        REQUIRE(b2Replay_GetFrame(p) == 16);
        b2Replay_Restart(p);
        REQUIRE(!b2Replay_HasDiverged(p));
        replay_observe(p, ORACLE_KEY("%s/restart", name));

        int targets[] = {6, 7, 16, 5, 12, 0, 16};

        for (int i = 0; i < 7; ++i) {
            b2Replay_SeekFrame(p, targets[i]);
            replay_observe(p, ORACLE_KEY("%s/seek%d", name, i));
        }

        b2DestroyReplay(p);
    }

    ORACLE_CLOSE();
}

static uint8_t* malformed(const uint8_t* source, int size, const char* name, int* outputSize) {
    uint8_t* data = malloc((size_t)size + 8);

    REQUIRE(data);
    memcpy(data, source, (size_t)size);

    *outputSize = size;
    b2RecHeader hdr;

    memcpy(&hdr, data, sizeof(hdr));

    int end = (int)sizeof(hdr) + (int)hdr.snapshotSize;

    if (strcmp(name, "empty") == 0) {
        *outputSize = 0;
    } else if (strcmp(name, "short") == 0) {
        *outputSize = (int)sizeof(hdr) - 1;
    } else if (strcmp(name, "magic") == 0) {
        hdr.magic ^= 1;
    } else if (strcmp(name, "major") == 0) {
        hdr.versionMajor ^= 1;
    } else if (strcmp(name, "minor") == 0) {
        hdr.versionMinor ^= 1;
    } else if (strcmp(name, "endian") == 0) {
        hdr.bigEndian = 1;
    } else if (strcmp(name, "snapshotZero") == 0) {
        hdr.snapshotSize = 0;
    } else if (strcmp(name, "snapshotOversize") == 0) {
        hdr.snapshotSize = (uint64_t)size;
    } else if (strcmp(name, "incompleteFrame") == 0) {
        data[end]     = 0xf1;
        data[end + 1] = 12;
        *outputSize   = end + 2;
    } else if (strcmp(name, "truncatedHash") == 0) {
        const uint8_t tail[] = {0xf1, 12, 0, 0, 0};

        memcpy(data + end, tail, sizeof(tail));

        *outputSize = end + (int)sizeof(tail);
    } else {
        const uint8_t tail[] = {0xff, 1, 0, 0, 0x55};

        memcpy(data + end, tail, sizeof(tail));

        if (strcmp(name, "unknownOverflow") == 0) { data[end + 1] = 2; }

        *outputSize = end + (int)sizeof(tail);
    }

    memcpy(data, &hdr, sizeof(hdr));
    return data;
}

static void file_and_framing_parity(void) {
    ORACLE_OPEN("recording/FileAndFramingParity");

    b2Recording* rec    = lifecycle_source(1);
    char         root[] = "/tmp/box2d-replay-XXXXXX";

    REQUIRE(mkdtemp(root));

    char path[256];

    snprintf(path, sizeof(path), "%s/absent/out.b2rec", root);
    ORACLE_BOOL("save/missingParent", !b2SaveRecordingToFile(rec, path));
    snprintf(path, sizeof(path), "%s/missing.b2rec", root);

    b2Recording* loaded = b2LoadRecordingFromFile(path);

    ORACLE_BOOL("load/missing", loaded == NULL);

    if (loaded) { b2DestroyRecording(loaded); }

    loaded = b2LoadRecordingFromFile(root);

    ORACLE_BOOL("load/directory", loaded == NULL);

    if (loaded) { b2DestroyRecording(loaded); }

    const char* names[] = {
        "empty",
        "short",
        "magic",
        "major",
        "minor",
        "endian",
        "snapshotZero",
        "snapshotOversize",
        "incompleteFrame",
        "truncatedHash",
        "unknown",
        "unknownOverflow"
    };

    for (int i = 0; i < 12; ++i) {
        int      size;
        uint8_t* data =
            malformed(b2Recording_GetData(rec), b2Recording_GetSize(rec), names[i], &size);
        b2Replay* p = b2CreateReplay(data, size, 0);

        ORACLE_BOOL(ORACLE_KEY("%s/open", names[i]), p != NULL);
        ORACLE_BOOL(ORACLE_KEY("%s/validate", names[i]), b2ValidateReplay(data, size, 0));

        if (p) {
            while (b2Replay_StepFrame(p)) {}

            ORACLE_I32(ORACLE_KEY("%s/frame", names[i]), b2Replay_GetFrame(p));
            ORACLE_BOOL(ORACLE_KEY("%s/diverged", names[i]), b2Replay_HasDiverged(p));
            b2DestroyReplay(p);
        }

        free(data);
    }
    // C's pointer-width field is an ABI rejection, not a common Go format requirement.

    int      size = b2Recording_GetSize(rec);
    uint8_t* data = malloc((size_t)size);

    REQUIRE(data);
    memcpy(data, b2Recording_GetData(rec), (size_t)size);

    b2RecHeader hdr;

    memcpy(&hdr, data, sizeof(hdr));

    hdr.pointerWidth ^= 0xff;

    memcpy(data, &hdr, sizeof(hdr));

    b2Replay* rejected = b2CreateReplay(data, size, 0);

    REQUIRE(rejected == NULL);
    free(data);
    b2DestroyRecording(rec);
    REQUIRE(rmdir(root) == 0);
    ORACLE_CLOSE();
}

int main(void) {
    lifecycle_parity();
    owned_source_parity();
    divergence_parity();
    file_and_framing_parity();

    return 0;
}
