// SPDX-License-Identifier: MIT

#include "oracle.h"

#include <limits.h>
#include <stdint.h>
#include <string.h>

static float FloatFromBits(uint32_t bits) {
    float v;

    memcpy(&v, &bits, sizeof(v));

    return v;
}

int main(void) {
    ORACLE_OPEN("types");

    ORACLE_F32("f32/one", 1.0f);
    ORACLE_F32("f32/neg_zero", -0.0f);
    ORACLE_F32("f32/pi", B2_PI);
    ORACLE_F32("f32/denormal", FloatFromBits(0x00000001));
    ORACLE_F32("f32/inf", FloatFromBits(0x7f800000));
    ORACLE_F32("f32/neg_inf", FloatFromBits(0xff800000));
    ORACLE_F32("f32/nan", FloatFromBits(0x7fc00000));
    ORACLE_F32("f32/max", FloatFromBits(0x7f7fffff));

    ORACLE_VEC2("vec2/basic", (b2Vec2){1.5f, -2.25f});
    ORACLE_VEC2("vec2/zeros", (b2Vec2){0.0f, -0.0f});

    ORACLE_ROT("rot/identity", b2Rot_identity);
    ORACLE_ROT("rot/basic", (b2Rot){0.6f, -0.8f});

    ORACLE_POS("pos/basic", oracle_pos_xy(0.1, -1.0e-3));
    ORACLE_POS("pos/far", oracle_pos_xy(123456.789, -0.0));

    ORACLE_I32("i32/zero", 0);
    ORACLE_I32("i32/minus_one", -1);
    ORACLE_I32("i32/min", INT32_MIN);
    ORACLE_I32("i32/max", INT32_MAX);

    ORACLE_U32("u32/zero", 0);
    ORACLE_U32("u32/max", UINT32_MAX);

    ORACLE_U64("u64/zero", 0);
    ORACLE_U64("u64/high", UINT64_C(0x8000000000000001));
    ORACLE_U64("u64/max", UINT64_MAX);

    ORACLE_BOOL("bool/true", true);
    ORACLE_BOOL("bool/false", false);

    const uint8_t bytes[] = {0x00, 0x01, 0x7f, 0x80, 0xff};

    ORACLE_BYTES("bytes/five", bytes, sizeof(bytes));
    ORACLE_BYTES("bytes/empty", bytes, 0);

    ORACLE_BOOL("key \"quoted\" \\ back\tslash/\xc3\xa9", true);

    ORACLE_U64("hash/init", ORACLE_HASH_INIT);
    ORACLE_U64("hash/bytes", oracle_hash_bytes(ORACLE_HASH_INIT, "abc", 3));
    ORACLE_U64("hash/u32", oracle_hash_u32(ORACLE_HASH_INIT, 0x01020304u));
    ORACLE_U64("hash/u64", oracle_hash_u64(ORACLE_HASH_INIT, UINT64_C(0x0102030405060708)));
    ORACLE_U64("hash/f32", oracle_hash_f32(ORACLE_HASH_INIT, 1.5f));
    ORACLE_U64("hash/f64", oracle_hash_f64(ORACLE_HASH_INIT, -2.5));
    ORACLE_U64("hash/pos", oracle_hash_pos(ORACLE_HASH_INIT, (b2Pos){1.5, -2.25}));
    ORACLE_U64("hash/rot", oracle_hash_rot(ORACLE_HASH_INIT, (b2Rot){0.6f, -0.8f}));
    ORACLE_U64(
        "hash/transform",
        oracle_hash_transform(ORACLE_HASH_INIT, (b2Pos){1.5, -2.25}, (b2Rot){0.6f, -0.8f})
    );

    ORACLE_CLOSE();
    return 0;
}
