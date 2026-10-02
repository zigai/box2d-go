// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/human.c"
#include "scenes/utils.c"

int main(void)
{
 ORACLE_OPEN("scenes/Human");
 b2WorldDef def = b2DefaultWorldDef();
 b2WorldId world = b2CreateWorld(&def);
 Human human = {0};
 CreateHuman(&human, world, (b2Pos){0, 5}, 1.0f, 0.05f, 5.0f, 0.5f, 1, NULL, true);
 scenes_dump(world, -1);
 for (int step = 0; step < 300; ++step)
 {
  if (step == 0) Human_SetVelocity(&human, (b2Vec2){1, 0});
  if (step == 20) Human_ApplyRandomAngularImpulse(&human, 0.5f);
  if (step == 40) Human_SetJointFrictionTorque(&human, 0);
  if (step == 60) Human_SetJointFrictionTorque(&human, 0.1f);
  if (step == 80) Human_SetJointSpringHertz(&human, 0);
  if (step == 100) Human_SetJointSpringHertz(&human, 2.0f);
  if (step == 120) Human_SetJointDampingRatio(&human, 0.75f);
  if (step == 140) Human_EnableSensorEvents(&human, true);
  if (step == 160) Human_SetScale(&human, 1.5f);
  if (step == 180) Human_EnableSensorEvents(&human, false);
  if (step == 140 || step == 180)
  {
   b2ShapeId shape = b2_nullShapeId;
   int count = b2Body_GetShapes(human.bones[bone_torso].bodyId, &shape, 1);
   ORACLE_I32(scenes_key(step, "torsoShapes"), count);
   bool enabled = b2Shape_AreSensorEventsEnabled(shape);
   ORACLE_BOOL(scenes_key(step, "sensor"), enabled);
   if (count != 1 || enabled != (step == 140)) abort();
  }
  b2World_Step(world, 1.0f/60.0f, 4);
  scenes_dump(world, step);
 }
 DestroyHuman(&human);
 scenes_dump(world, 300);
 b2DestroyWorld(world);
 scenes_flush();
 ORACLE_CLOSE();
 return 0;
}
