// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"

#include <assert.h>
#include <stdlib.h>
#include <string.h>

static uint8_t* math_data;
static size_t   math_size;
static size_t   math_capacity;

static void math_put(uint32_t value, int width) {
    if (math_size + 4 > math_capacity) {
        math_capacity = math_capacity ? 2 * math_capacity : 1048576;
        void* grown   = realloc(math_data, math_capacity);
        if (grown == NULL) { abort(); }

        math_data = grown;
    }

    for (int i = 0; i < width; ++i) { math_data[math_size++] = (uint8_t)(value >> (8 * i)); }
}

static void math_f32(float value) {
    uint32_t bits;

    memcpy(&bits, &value, sizeof(bits));
    math_put(bits, 4);
}

#define ORACLE_MATH_F32(site, value) math_f32(value)
#define ORACLE_MATH_BOOL(site, value) math_put((value) ? 1 : 0, 1)
#define ORACLE_MATH_CLOSE()                                                                        \
    do {                                                                                           \
        ORACLE_BYTES("checks", math_data, math_size);                                              \
        ORACLE_CLOSE();                                                                            \
    } while (false)

#include "upstream/test_math.c"

int main(void) {
    return MathTest();
}
