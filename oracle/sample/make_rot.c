// SPDX-License-Identifier: MIT

#include "box2d/math_functions.h"
#include "oracle.h"

#include <stdint.h>
#include <string.h>

static float FloatFromBits(uint32_t bits) {
    float v;

    memcpy(&v, &bits, sizeof(v));

    return v;
}

int main(void) {
    const float table[] = {
        0.0f,
        -0.0f,
        B2_PI,
        -B2_PI,
        0.5f * B2_PI,
        -0.5f * B2_PI,
        0.25f * B2_PI,
        -0.25f * B2_PI,
        0.75f * B2_PI,
        -0.75f * B2_PI,
        FloatFromBits(0x40490fdb),
        FloatFromBits(0x40490fda),
        FloatFromBits(0x40490fdc),
        2.0f * B2_PI,
        -2.0f * B2_PI,
        3.0f * B2_PI,
        1.0e-8f,
        -1.0e-8f,
        1.0e-30f,
        FloatFromBits(0x00000001),
        FloatFromBits(0x007fffff),
        FloatFromBits(0x00800000),
        0.1f,
        -0.1f,
        0.5f,
        1.0f,
        -1.0f,
        2.0f,
        3.0f,
        -3.0f,
        10.0f,
        -10.0f,
        100.0f,
        -100.0f,
        1000.0f,
        -1000.0f,
        9999.5f,
        -9999.5f,
    };

    const int tableCount = (int)(sizeof(table) / sizeof(table[0]));

    const int sweepCount = 128;

    ORACLE_OPEN("make_rot");
    ORACLE_I32("count", tableCount + sweepCount);

    uint64_t hash = ORACLE_HASH_INIT;

    for (int i = 0; i < tableCount + sweepCount; ++i) {
        float angle = i < tableCount ? table[i] : (float)(i - tableCount) * 0.1f - 6.4f;

        float    unwound = b2UnwindAngle(angle);
        b2CosSin cs      = b2ComputeCosSin(angle);
        b2Rot    q       = b2MakeRot(angle);

        ORACLE_F32(ORACLE_KEY("%d/angle", i), angle);
        ORACLE_F32(ORACLE_KEY("%d/unwound", i), unwound);
        ORACLE_F32(ORACLE_KEY("%d/cosine", i), cs.cosine);
        ORACLE_F32(ORACLE_KEY("%d/sine", i), cs.sine);
        ORACLE_ROT(ORACLE_KEY("%d/rot", i), q);

        hash = oracle_hash_rot(hash, q);
    }

    ORACLE_U64("hash", hash);
    ORACLE_CLOSE();

    return 0;
}
