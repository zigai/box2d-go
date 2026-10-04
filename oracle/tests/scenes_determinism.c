// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/determinism.c"

int main(void)
{
 for (int workers = 1; workers <= 4; workers += 3) {
 char fixture[64]; snprintf(fixture, sizeof(fixture), "scenes/FallingHinges_workers%d", workers);
 ORACLE_OPEN(fixture);
 b2WorldDef def = b2DefaultWorldDef();
 def.workerCount = workers;
 b2WorldId world = b2CreateWorld(&def);
 FallingHingeData data = CreateFallingHinges(world);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  b2World_Step(world, 1.0f/60.0f, 4);
  bool done = UpdateFallingHinges(world, &data);
  ORACLE_BOOL(scenes_key(step, "done"), done);
  ORACLE_U32(scenes_key(step, "hash"), data.hash);
  ORACLE_I32(scenes_key(step, "sleep"), data.sleepStep);
  ORACLE_I32(scenes_key(step, "step"), data.stepCount);
  scenes_dump(world, step);
 }
 DestroyFallingHinges(&data);
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 }
 return 0;
}
