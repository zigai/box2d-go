// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT
#include "scenes/driver.h"
#include "scenes/benchmarks.c"
#include "scenes/human.c"
#include "scenes/utils.c"

int main(void)
{
 // Correctness reaches every pinned CLI horizon AND the matched custom300 horizon.
 // Timing remains a separate custom300 workload, not the CLI's initial-step warm-up.
 const struct { const char* name; void (*create)(b2WorldId); float (*step)(b2WorldId,int); b2TreeStats (*stats)(void); void (*destroy)(void); int horizon; } cases[] = {
  {"JointGrid",CreateJointGrid,NULL,NULL,NULL,500},
  {"LargePyramid",CreateLargePyramid,NULL,NULL,NULL,500},
  {"ManyPyramids",CreateManyPyramids,NULL,NULL,NULL,300},
  {"Rain",CreateRain,StepRain,NULL,NULL,1000},
  {"Spinner",CreateSpinner,StepSpinner,NULL,NULL,500},
  {"Smash",CreateSmash,NULL,NULL,NULL,300},
  {"Tumbler",CreateTumbler,NULL,NULL,NULL,750},
  {"Washer",CreateWasher,NULL,NULL,NULL,500},
  {"Junkyard",CreateJunkyard,StepJunkyard,NULL,NULL,800},
  {"Sleep",CreateSleep,StepSleep,NULL,NULL,300},
  {"Compounds",CreateCompounds,NULL,NULL,NULL,500},
  {"Queries",CreateQueries,StepQueries,GetQueryBenchmarkStats,NULL,300},
  {"TreeCast",CreateTreeCast,StepTreeCast,GetTreeCastBenchmarkStats,DestroyTreeCast,300},
  {"TileWorld",CreateTileWorld,StepTileWorld,GetTileWorldBenchmarkStats,DestroyTileWorld,300},
 };
 for(int workers=1;workers<=4;workers+=3) {
  for(size_t i=0;i<sizeof(cases)/sizeof(cases[0]);++i) {
   g_randomSeed=12345;
   char name[64]; snprintf(name,sizeof(name),"scenes/%s_workers%d",cases[i].name,workers); ORACLE_OPEN(name);
   b2WorldDef def=b2DefaultWorldDef();def.workerCount=workers;b2WorldId world=b2CreateWorld(&def);
   cases[i].create(world);scene_dump(world,-1);
   if(strcmp(cases[i].name,"ManyPyramids")==0) {
    b2Capacity c=GetManyPyramidsCapacity();ORACLE_I32("staticShapeCount",c.staticShapeCount);ORACLE_I32("staticBodyCount",c.staticBodyCount);ORACLE_I32("dynamicShapeCount",c.dynamicShapeCount);ORACLE_I32("dynamicBodyCount",c.dynamicBodyCount);ORACLE_I32("contactCount",c.contactCount);
   }
   if(strcmp(cases[i].name,"Queries")==0) {
    ORACLE_I32("queryCount",GetQueryBenchmarkCount());ORACLE_F32("extent",GetQueryBenchmarkExtent());
    b2Pos origin;b2Vec2 translation;GetQueryBenchmarkRay(0,&origin,&translation);ORACLE_POS("origin0",origin);ORACLE_VEC2("translation0",translation);GetQueryBenchmarkRay(GetQueryBenchmarkCount()-1,&origin,&translation);ORACLE_POS("origin999",origin);ORACLE_VEC2("translation999",translation);
   }
   for(int step=0;step<cases[i].horizon;++step) {
    if(cases[i].step) ORACLE_F32(scene_key(step,"v"),cases[i].step(world,step));
    b2World_Step(world,1.0f/60.0f,4);scene_dump(world,step);
    if(cases[i].stats) {b2TreeStats s=cases[i].stats();ORACLE_I32(scene_key(step,"n"),s.nodeVisits);ORACLE_I32(scene_key(step,"l"),s.leafVisits);}
   }
   if(cases[i].destroy) cases[i].destroy();b2DestroyWorld(world);scene_flush();ORACLE_CLOSE();
  }
 }
 return 0;
}
