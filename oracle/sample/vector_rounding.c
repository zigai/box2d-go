// SPDX-License-Identifier: MIT

#include "box2d/math_functions.h"
#include "oracle.h"

#include <stdint.h>
#include <string.h>

static float FloatFromBits(uint32_t bits) {
    float value;

    memcpy(&value, &bits, sizeof(value));

    return value;
}

int main(void) {
    const struct {
        const char* name;
        uint32_t    a[2];
        uint32_t    b[2];
        uint32_t    s;
    } cases[] = {
        {"MulSV", {0x3f6ceba4, 0x3f5e42b3}, {0xbf46a876, 0xbf2a855d}, 0xbf4fc718},
        {"CrossSV", {0xbf279f97, 0x3f47380a}, {0xbf5ac7e1, 0x3f27bfcc}, 0xbf4af8bb},
    };

    ORACLE_OPEN("vector_rounding");
    ORACLE_I32("count", 2);

    for (int i = 0; i < 2; ++i) {
        b2Vec2 a       = {FloatFromBits(cases[i].a[0]), FloatFromBits(cases[i].a[1])};
        b2Vec2 b       = {FloatFromBits(cases[i].b[0]), FloatFromBits(cases[i].b[1])};
        float  s       = FloatFromBits(cases[i].s);
        b2Vec2 product = i == 0 ? b2MulSV(s, b) : b2CrossSV(s, b);
        b2Vec2 result  = i == 0 ? b2Add(a, b2MulSV(s, b)) : b2Add(a, b2CrossSV(s, b));

        ORACLE_VEC2(ORACLE_KEY("%s/a", cases[i].name), a);
        ORACLE_VEC2(ORACLE_KEY("%s/b", cases[i].name), b);
        ORACLE_F32(ORACLE_KEY("%s/s", cases[i].name), s);
        ORACLE_VEC2(ORACLE_KEY("%s/product", cases[i].name), product);
        ORACLE_VEC2(ORACLE_KEY("%s/result", cases[i].name), result);
    }

    ORACLE_CLOSE();
    return 0;
}
