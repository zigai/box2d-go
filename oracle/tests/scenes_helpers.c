// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/benchmarks.c"
#include "scenes/human.c"
#include "scenes/utils.c"

static void counts(b2WorldId world, int step)
{
 b2Counters c = b2World_GetCounters(world);
 ORACLE_I32(scene_key(step, "b"), c.bodyCount);
 ORACLE_I32(scene_key(step, "s"), c.shapeCount);
 ORACLE_I32(scene_key(step, "j"), c.jointCount);
}

static void group_flags(int step, int column)
{
 bool spawned = true, cleared = true;
 for (int row = 0; row < RAIN_ROW_COUNT; ++row)
  for (int i = 0; i < RAIN_GROUP_SIZE; ++i)
  {
   Human* h = &g_rainData.groups[row * RAIN_COLUMN_COUNT + column].humans[i];
   spawned = spawned && h->isSpawned;
   for (int b = 0; b < bone_count; ++b)
    cleared = cleared && B2_IS_NULL(h->bones[b].bodyId) && B2_IS_NULL(h->bones[b].jointId);
  }
 ORACLE_BOOL(scene_key(step, "spawned"), spawned);
 ORACLE_BOOL(scene_key(step, "cleared"), cleared);
 ORACLE_I32(scene_key(step, "column"), g_rainData.columnIndex);
 if (spawned != (step != 1) || cleared != (step == 1)) abort();
}

int main(void)
{
 ORACLE_OPEN("scenes/Helpers");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 CreateRain(world);
 counts(world, -1);
 for (int row = 0; row < RAIN_ROW_COUNT; ++row) CreateGroup(world, row, RAIN_COLUMN_COUNT - 1);
 counts(world, 0);
 group_flags(0, RAIN_COLUMN_COUNT - 1);
 b2BodyId oldBody = g_rainData.groups[RAIN_COLUMN_COUNT - 1].humans[0].bones[bone_torso].bodyId;
 b2JointId oldJoint = g_rainData.groups[RAIN_COLUMN_COUNT - 1].humans[0].bones[bone_torso].jointId;
 for (int row = 0; row < RAIN_ROW_COUNT; ++row) DestroyGroup(row, RAIN_COLUMN_COUNT - 1);
 counts(world, 1);
 group_flags(1, RAIN_COLUMN_COUNT - 1);
 ORACLE_BOOL("1.oldBody", b2Body_IsValid(oldBody));
 ORACLE_BOOL("1.oldJoint", b2Joint_IsValid(oldJoint));
 for (int row = 0; row < RAIN_ROW_COUNT; ++row) CreateGroup(world, row, RAIN_COLUMN_COUNT - 1);
 g_rainData.columnCount = RAIN_COLUMN_COUNT;
 g_rainData.columnIndex = RAIN_COLUMN_COUNT - 1;
 oldBody = g_rainData.groups[RAIN_COLUMN_COUNT - 1].humans[0].bones[bone_torso].bodyId;
 oldJoint = g_rainData.groups[RAIN_COLUMN_COUNT - 1].humans[0].bones[bone_torso].jointId;
 ORACLE_F32("2.v", StepRain(world, 320));
 counts(world, 2);
 group_flags(2, RAIN_COLUMN_COUNT - 1);
 ORACLE_BOOL("2.oldBody", b2Body_IsValid(oldBody));
 ORACLE_BOOL("2.oldJoint", b2Joint_IsValid(oldJoint));
 if (g_rainData.columnIndex != 0 || g_rainData.columnCount != RAIN_COLUMN_COUNT) abort();
 b2DestroyWorld(world);
 world = b2CreateWorld(&def);
 CreateQueries(world);
 float positions[4] = {g_queryBenchmark.paddleMaxX + 1, g_queryBenchmark.paddleMaxX,
                       g_queryBenchmark.paddleMinX - 1, g_queryBenchmark.paddleMinX};
 for (int i = 0; i < 4; ++i)
 {
  b2Body_SetTransform(g_queryBenchmark.paddleId, (b2Pos){positions[i], 0}, b2Rot_identity);
  b2Body_SetLinearVelocity(g_queryBenchmark.paddleId, (b2Vec2){i < 2 ? 5 : -5, 0});
  ORACLE_F32(scene_key(i, "q"), StepQueries(world, i));
  ORACLE_POS(scene_key(i, "position"), b2Body_GetPosition(g_queryBenchmark.paddleId));
  ORACLE_VEC2(scene_key(i, "velocity"), b2Body_GetLinearVelocity(g_queryBenchmark.paddleId));
  float expected[4] = {-5, 5, 5, -5};
  b2Vec2 velocity = b2Body_GetLinearVelocity(g_queryBenchmark.paddleId);
  if (velocity.x != expected[i] || velocity.y != 0) abort();
  b2TreeStats s = GetQueryBenchmarkStats();
  ORACLE_I32(scene_key(i, "n"), s.nodeVisits);
  ORACLE_I32(scene_key(i, "l"), s.leafVisits);
 }
 b2DestroyWorld(world);
 ORACLE_CLOSE();
 return 0;
}
