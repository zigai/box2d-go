// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/utils.c"
#include <assert.h>

int main(void)
{
 ORACLE_OPEN("scenes/Utils");
 assert(GetNumberOfCores() > 0);
 g_randomSeed = RAND_SEED;
 for (int step = 0; step < 300; ++step)
 {
  ORACLE_I32(scene_key(step, "int"), RandomInt());
  ORACLE_I32(scene_key(step, "range"), RandomIntRange(-7, 13));
  ORACLE_F32(scene_key(step, "float"), RandomFloat());
  ORACLE_F32(scene_key(step, "floatRange"), RandomFloatRange(-0.75f, 2.5f));
  ORACLE_VEC2(scene_key(step, "vec"), RandomVec2(-1.5f, 3.0f));
  ORACLE_POS(scene_key(step, "pos"), RandomPos(-2.0f, 4.0f));
  ORACLE_ROT(scene_key(step, "rot"), RandomRot());
  b2Polygon polygon = RandomPolygon(1.0f);
  ORACLE_I32(scene_key(step, "count"), polygon.count);
  ORACLE_F32(scene_key(step, "radius"), polygon.radius);
  ORACLE_VEC2(scene_key(step, "centroid"), polygon.centroid);
  for (int i = 0; i < polygon.count; ++i)
  {
   ORACLE_VEC2(scene_key(step, "vertex"), polygon.vertices[i]);
   ORACLE_VEC2(scene_key(step, "normal"), polygon.normals[i]);
  }
 }
 g_randomSeed = RAND_SEED;
 ORACLE_I32("equal.int", RandomIntRange(7, 7));
 ORACLE_F32("equal.float", RandomFloatRange(1.25f, 1.25f));
 ORACLE_VEC2("equal.vec", RandomVec2(-2.0f, -2.0f));
 ORACLE_POS("equal.pos", RandomPos(0.5f, 0.5f));
 ORACLE_U32("equal.seed", g_randomSeed);
 // XorShift's absorbing zero seed produces coincident points: exercise fallback.
 g_randomSeed = 0;
 b2Polygon fallback = RandomPolygon(1.0f);
 if (fallback.count != 4) abort();
 ORACLE_I32("fallback.count", fallback.count);
 ORACLE_F32("fallback.radius", fallback.radius);
 ORACLE_VEC2("fallback.centroid", fallback.centroid);
 for (int i = 0; i < fallback.count; ++i) {
  ORACLE_VEC2("fallback.vertex", fallback.vertices[i]);
  ORACLE_VEC2("fallback.normal", fallback.normals[i]);
 }
 ORACLE_CLOSE();
 return 0;
}
