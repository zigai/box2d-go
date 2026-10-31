// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/benchmarks.c"
#include "scenes/human.c"
#include "scenes/utils.c"
#include <math.h>

static uint64_t trig_hash(uint64_t hash, float value)
{
 uint32_t bits;
 memcpy(&bits, &value, sizeof(bits));
 for (int b = 0; b < 4; ++b) { hash ^= (bits >> (8*b)) & 255u; hash *= UINT64_C(1099511628211); }
 return hash;
}

static void trig_dump(int index, float angle)
{
 volatile float argument = angle;
 float s = sinf(argument), c = cosf(argument);
 ORACLE_F32(scene_key(index, "angle"), angle);
 ORACLE_F32(scene_key(index, "sin"), s);
 ORACLE_F32(scene_key(index, "cos"), c);
 ORACLE_F32(scene_key(index, "y"), 25.0f*s);
 ORACLE_F32(scene_key(index, "x"), 25.0f*c);
}

int main(void)
{
 ORACLE_OPEN("scenes/Trig");
 for (int step = 0; step < 300; ++step)
 {
  uint32_t state = 1234u + 7919u*(uint32_t)step;
  uint64_t hash = UINT64_C(14695981039346656037);
  for (int ray = 0; ray < TILE_QUERY_COUNT; ++ray)
  {
   TileRandom(&state, 0, 1);
   TileRandom(&state, -2, 18);
   float angle = TileRandom(&state, 0, 2.0f*B2_PI);
   volatile float argument = angle;
   float s = sinf(argument), c = cosf(argument);
   hash = trig_hash(hash, angle);
   hash = trig_hash(hash, s);
   hash = trig_hash(hash, c);
   hash = trig_hash(hash, 25.0f*s);
   hash = trig_hash(hash, 25.0f*c);
   if (step == 0 && ray == 41) trig_dump(0, angle);
  }
  ORACLE_U64(scene_key(step, "hash"), hash);
 }
 trig_dump(1, B2_PI/18.0f);
 trig_dump(2, 0.1f*(B2_PI/18.0f));
 float centers[] = {0.0f, 0x1p-149f, 0x1p-126f, 0x1p-12f, 0.75f, B2_PI/4.0f, B2_PI/2.0f,
                    3.0f*B2_PI/4.0f, B2_PI, 5.0f*B2_PI/4.0f, 3.0f*B2_PI/2.0f, 7.0f*B2_PI/4.0f, 2.0f*B2_PI};
 int index = 3;
 for (unsigned i = 0; i < sizeof(centers)/sizeof(centers[0]); ++i)
 {
  float values[] = {nextafterf(centers[i], -INFINITY), centers[i], nextafterf(centers[i], INFINITY)};
  for (int j = 0; j < 3; ++j) { trig_dump(index++, values[j]); trig_dump(index++, -values[j]); }
 }
 ORACLE_CLOSE();
 return 0;
}
