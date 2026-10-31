// SPDX-License-Identifier: MIT

#pragma once

#include "box2d/math_types.h"

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#ifdef __cplusplus
extern "C" {
#endif

void oracle_open(const char* name);

void oracle_close(void);

void oracle_f32(const char* key, float v);
void oracle_vec2(const char* key, b2Vec2 v);
void oracle_rot(const char* key, b2Rot q);
void oracle_pos(const char* key, b2Pos p);
void oracle_i32(const char* key, int32_t v);
void oracle_u32(const char* key, uint32_t v);
void oracle_u64(const char* key, uint64_t v);
void oracle_bool(const char* key, bool v);
void oracle_bytes(const char* key, const void* data, size_t size);

const char* oracle_key(const char* format, ...)
#if defined(__GNUC__) || defined(__clang__)
    __attribute__((format(printf, 1, 2)))
#endif
    ;

#define ORACLE_OPEN(name) oracle_open(name)
#define ORACLE_CLOSE() oracle_close()
#define ORACLE_F32(key, ...) oracle_f32((key), (__VA_ARGS__))
#define ORACLE_VEC2(key, ...) oracle_vec2((key), (__VA_ARGS__))
#define ORACLE_ROT(key, ...) oracle_rot((key), (__VA_ARGS__))
#define ORACLE_POS(key, ...) oracle_pos((key), (__VA_ARGS__))
#define ORACLE_I32(key, ...) oracle_i32((key), (__VA_ARGS__))
#define ORACLE_U32(key, ...) oracle_u32((key), (__VA_ARGS__))
#define ORACLE_U64(key, ...) oracle_u64((key), (__VA_ARGS__))
#define ORACLE_BOOL(key, ...) oracle_bool((key), (__VA_ARGS__))
#define ORACLE_BYTES(key, data, size) oracle_bytes((key), (data), (size))
#define ORACLE_KEY oracle_key

#if defined(BOX2D_DOUBLE_PRECISION)
typedef double oracle_pos_scalar;
#else
typedef float oracle_pos_scalar;
#endif

// Applies the conversion a b2Pos initializer would apply in the active precision.
static inline b2Pos oracle_pos_xy(double x, double y) {
    return (b2Pos){(oracle_pos_scalar)x, (oracle_pos_scalar)y};
}

// Copies a key out of the rotating ORACLE_KEY buffers before they are reused.
static inline void oracle_copy_key(char* dst, size_t size, const char* src) {
    size_t length = strlen(src);
    if (length >= size) { abort(); }

    memcpy(dst, src, length + 1);
}

#define ORACLE_HASH_INIT UINT64_C(14695981039346656037)

static inline uint64_t oracle_hash_bytes(uint64_t hash, const void* data, size_t size) {
    const uint8_t* bytes = (const uint8_t*)data;

    for (size_t i = 0; i < size; ++i) {
        hash ^= bytes[i];
        hash *= UINT64_C(1099511628211);
    }

    return hash;
}

static inline uint64_t oracle_hash_u32(uint64_t hash, uint32_t v) {
    const uint8_t bytes[4] =
        {(uint8_t)v, (uint8_t)(v >> 8), (uint8_t)(v >> 16), (uint8_t)(v >> 24)};
    return oracle_hash_bytes(hash, bytes, sizeof(bytes));
}

static inline uint64_t oracle_hash_u64(uint64_t hash, uint64_t v) {
    hash = oracle_hash_u32(hash, (uint32_t)v);
    return oracle_hash_u32(hash, (uint32_t)(v >> 32));
}

uint32_t oracle_bits_f32(float v);
uint64_t oracle_bits_f64(double v);

static inline uint64_t oracle_hash_f32(uint64_t hash, float v) {
    return oracle_hash_u32(hash, oracle_bits_f32(v));
}

static inline uint64_t oracle_hash_f64(uint64_t hash, double v) {
    return oracle_hash_u64(hash, oracle_bits_f64(v));
}

static inline uint64_t oracle_hash_pos(uint64_t hash, b2Pos p) {
#if defined(BOX2D_DOUBLE_PRECISION)
    hash = oracle_hash_f64(hash, p.x);
    return oracle_hash_f64(hash, p.y);
#else
    hash = oracle_hash_f32(hash, p.x);
    return oracle_hash_f32(hash, p.y);
#endif
}

static inline uint64_t oracle_hash_rot(uint64_t hash, b2Rot q) {
    hash = oracle_hash_f32(hash, q.c);
    return oracle_hash_f32(hash, q.s);
}

static inline uint64_t oracle_hash_transform(uint64_t hash, b2Pos p, b2Rot q) {
    hash = oracle_hash_pos(hash, p);
    return oracle_hash_rot(hash, q);
}

#ifdef __cplusplus
}
#endif
