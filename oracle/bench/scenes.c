// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "benchmarks.h"
#include "determinism.h"
#include "human.h"
#include "utils.h"
#include "box2d/box2d.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char** argv) {
 int samples = argc > 1 ? atoi(argv[1]) : 6;
 int workers = argc > 2 ? atoi(argv[2]) : 1;
 int cpus = argc > 3 ? atoi(argv[3]) : 2;
 if (samples < 1 || workers < 1 || workers > B2_MAX_WORKERS || cpus < 1) return 1;
 const char* names[] = {"JointGrid","LargePyramid","ManyPyramids","Rain","Spinner","Smash","Tumbler","Washer","Junkyard","Sleep","Compounds","Queries","TreeCast","TileWorld","FallingHinges","Human"};
 for (int scene = 0; scene < 16; ++scene) {
  for (int sample = 0; sample < samples; ++sample) {
   g_randomSeed = 12345;
   b2WorldDef def = b2DefaultWorldDef();
   def.workerCount = workers;
   b2WorldId world = b2CreateWorld(&def);
   FallingHingeData hinges = {0};
   Human human = {0};
   switch (scene) {
case 0: CreateJointGrid(world); break;
case 1: CreateLargePyramid(world); break;
case 2: CreateManyPyramids(world); break;
case 3: CreateRain(world); break;
case 4: CreateSpinner(world); break;
case 5: CreateSmash(world); break;
case 6: CreateTumbler(world); break;
case 7: CreateWasher(world); break;
case 8: CreateJunkyard(world); break;
case 9: CreateSleep(world); break;
case 10: CreateCompounds(world); break;
case 11: CreateQueries(world); break;
case 12: CreateTreeCast(world); break;
case 13: CreateTileWorld(world); break;
case 14: hinges = CreateFallingHinges(world); break;
case 15: CreateHuman(&human, world, (b2Pos){0,5}, 1, 0.05f, 5, 0.5f, 1, NULL, true); break;
}
 uint64_t ticks = b2GetTicks();
 for (int step = 0; step < 300; ++step) {
 switch (scene) {
case 3: StepRain(world,step); break;
case 4: StepSpinner(world,step); break;
case 8: StepJunkyard(world,step); break;
case 9: StepSleep(world,step); break;
case 11: StepQueries(world,step); break;
case 12: StepTreeCast(world,step); break;
case 13: StepTileWorld(world,step); break;
case 15:
 switch (step) {
 case 0: Human_SetVelocity(&human, (b2Vec2){1,0}); break;
 case 20: Human_ApplyRandomAngularImpulse(&human,0.5f); break;
 case 40: Human_SetJointFrictionTorque(&human,0); break;
 case 60: Human_SetJointFrictionTorque(&human,0.1f); break;
 case 80: Human_SetJointSpringHertz(&human,0); break;
 case 100: Human_SetJointSpringHertz(&human,2); break;
 case 120: Human_SetJointDampingRatio(&human,0.75f); break;
 case 140: Human_EnableSensorEvents(&human,true); break;
 case 160: Human_SetScale(&human,1.5f); break;
 case 180: Human_EnableSensorEvents(&human,false); break;
 } break;
 }
 b2World_Step(world,1.0f/60.0f,4);
 if (scene == 14) UpdateFallingHinges(world,&hinges);
 }
 double nanoseconds = 1000000.0 * b2GetMilliseconds(ticks);
 if (scene == 14) DestroyFallingHinges(&hinges);
 if (scene == 15) DestroyHuman(&human);
 if (scene == 12) DestroyTreeCast();
 if (scene == 13) DestroyTileWorld();
 b2DestroyWorld(world);
 if (cpus == 1) {
  printf("BenchmarkScenes/%s/workers=%d 1 %.0f ns/op\n", names[scene], workers, nanoseconds);
 } else {
  printf("BenchmarkScenes/%s/workers=%d-%d 1 %.0f ns/op\n", names[scene], workers, cpus, nanoseconds);
 }
 }
 }
 return 0;
}
