// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/benchmarks.c"
#include "scenes/human.c"
#include "scenes/utils.c"

int main(void)
{
 {
 ORACLE_OPEN("scenes/JointGrid");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateJointGrid(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/LargePyramid");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateLargePyramid(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/ManyPyramids");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateManyPyramids(world);
 scenes_dump(world, -1);
 b2Capacity capacity = GetManyPyramidsCapacity();
 ORACLE_I32("staticShapeCount", capacity.staticShapeCount);
 ORACLE_I32("staticBodyCount", capacity.staticBodyCount);
 ORACLE_I32("dynamicShapeCount", capacity.dynamicShapeCount);
 ORACLE_I32("dynamicBodyCount", capacity.dynamicBodyCount);
 ORACLE_I32("contactCount", capacity.contactCount);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Rain");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateRain(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepRain(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Spinner");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateSpinner(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepSpinner(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Smash");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateSmash(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Tumbler");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateTumbler(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Washer");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateWasher(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Junkyard");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateJunkyard(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepJunkyard(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Sleep");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateSleep(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepSleep(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Compounds");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateCompounds(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/Queries");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateQueries(world);
 scenes_dump(world, -1);
 ORACLE_I32("queryCount", GetQueryBenchmarkCount());
 ORACLE_F32("extent", GetQueryBenchmarkExtent());
 b2Pos origin0; b2Vec2 translation0;
 GetQueryBenchmarkRay(0, &origin0, &translation0);
 ORACLE_POS("origin0", origin0);
 ORACLE_VEC2("translation0", translation0);
 b2Pos origin999; b2Vec2 translation999;
 GetQueryBenchmarkRay(GetQueryBenchmarkCount()-1, &origin999, &translation999);
 ORACLE_POS("origin999", origin999);
 ORACLE_VEC2("translation999", translation999);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepQueries(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
  b2TreeStats stats = GetQueryBenchmarkStats();
  ORACLE_I32(scenes_key(step, "n"), stats.nodeVisits);
  ORACLE_I32(scenes_key(step, "l"), stats.leafVisits);
 }
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/TreeCast");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateTreeCast(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepTreeCast(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
  b2TreeStats stats = GetTreeCastBenchmarkStats();
  ORACLE_I32(scenes_key(step, "n"), stats.nodeVisits);
  ORACLE_I32(scenes_key(step, "l"), stats.leafVisits);
 }
 DestroyTreeCast();
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 {
 ORACLE_OPEN("scenes/TileWorld");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateTileWorld(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  float value = StepTileWorld(world, step);
  ORACLE_F32(scenes_key(step, "v"), value);
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
  b2TreeStats stats = GetTileWorldBenchmarkStats();
  ORACLE_I32(scenes_key(step, "n"), stats.nodeVisits);
  ORACLE_I32(scenes_key(step, "l"), stats.leafVisits);
 }
 DestroyTileWorld();
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 return 0;
}
