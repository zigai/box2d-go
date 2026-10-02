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
  ORACLE_I32(scenes_key(step, "int"), RandomInt());
  ORACLE_I32(scenes_key(step, "range"), RandomIntRange(-7, 13));
  ORACLE_F32(scenes_key(step, "float"), RandomFloat());
  ORACLE_F32(scenes_key(step, "floatRange"), RandomFloatRange(-0.75f, 2.5f));
  ORACLE_VEC2(scenes_key(step, "vec"), RandomVec2(-1.5f, 3.0f));
  ORACLE_POS(scenes_key(step, "pos"), RandomPos(-2.0f, 4.0f));
  ORACLE_ROT(scenes_key(step, "rot"), RandomRot());
  b2Polygon polygon = RandomPolygon(1.0f);
  ORACLE_I32(scenes_key(step, "count"), polygon.count);
  ORACLE_F32(scenes_key(step, "radius"), polygon.radius);
  ORACLE_VEC2(scenes_key(step, "centroid"), polygon.centroid);
  for (int i = 0; i < polygon.count; ++i)
  {
   ORACLE_VEC2(scenes_key(step, "vertex"), polygon.vertices[i]);
   ORACLE_VEC2(scenes_key(step, "normal"), polygon.normals[i]);
  }
 }
 ORACLE_CLOSE();
 return 0;
}
