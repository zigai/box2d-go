#ifndef DYNAMICS_OBSERVE_H
#define DYNAMICS_OBSERVE_H
#include "box2d/box2d.h"
#include "body.h"
#include "physics_world.h"
#include "oracle.h"
#include <stdlib.h>
#include <stdio.h>

static inline uint64_t dynamics_hash_vec(uint64_t h,b2Vec2 v) { return oracle_hash_f32(oracle_hash_f32(h,v.x),v.y); }

static inline void dynamics_shape(const char* key,b2ShapeId s,uint16_t worldID,uint64_t* h,bool direct) {

 h[1]=oracle_hash_u32(h[1],(uint32_t)(s.index1));
 if(direct) ORACLE_I32(ORACLE_KEY("%sindex",key),s.index1);

 h[2]=oracle_hash_u32(h[2],(uint32_t)(s.generation));
 if(direct) ORACLE_I32(ORACLE_KEY("%sgeneration",key),s.generation);

 h[3]=oracle_hash_u32(h[3],(s.world0 == worldID) ? 1u : 0u);
 if(direct) ORACLE_BOOL(ORACLE_KEY("%sworld",key),s.world0 == worldID);

 h[34]=oracle_hash_u32(h[34],(b2Shape_IsValid(s)) ? 1u : 0u);
 if(direct) ORACLE_BOOL(ORACLE_KEY("%svalid",key),b2Shape_IsValid(s));
}
static inline void dynamics_contact(const char* key,b2ContactId c,b2ShapeId a,b2ShapeId b,uint16_t worldID,uint64_t* h,bool direct) {
 char child[256];

 h[1]=oracle_hash_u32(h[1],(uint32_t)(c.index1));
 if(direct) ORACLE_I32(ORACLE_KEY("%sindex",key),c.index1);

 h[2]=oracle_hash_u64(h[2],(c.generation));
 if(direct) ORACLE_U64(ORACLE_KEY("%sgeneration",key),c.generation);

 h[3]=oracle_hash_u32(h[3],(c.world0 == worldID) ? 1u : 0u);
 if(direct) ORACLE_BOOL(ORACLE_KEY("%sworld",key),c.world0 == worldID);

 h[34]=oracle_hash_u32(h[34],(b2Contact_IsValid(c)) ? 1u : 0u);
 if(direct) ORACLE_BOOL(ORACLE_KEY("%svalid",key),b2Contact_IsValid(c));
 if(direct) snprintf(child,sizeof(child),"%sshapeA/",key); else child[0]=0; dynamics_shape(child,a,worldID,h,direct);
 if(direct) snprintf(child,sizeof(child),"%sshapeB/",key); else child[0]=0; dynamics_shape(child,b,worldID,h,direct);
}
typedef struct dynamics_row { uint64_t fields[36]; } dynamics_row;
static inline dynamics_row dynamics_observe(b2WorldId id,const char* prefix,bool direct) {
 dynamics_row row; uint64_t* h=row.fields; for(int i=0;i<36;++i) h[i]=ORACLE_HASH_INIT;
 b2World* w = b2GetWorldFromId(id);

 h[0]=oracle_hash_u32(h[0],(uint32_t)(b2World_GetCounters(id).bodyCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%sbodyCount",prefix),b2World_GetCounters(id).bodyCount);
 for(int i=0;i<w->bodies.count;++i) {
  if(w->bodies.data[i].id != i) continue;
  b2BodyId b=b2MakeBodyId(w,i);
  char key[256]; if(direct) snprintf(key,sizeof(key),"%sbody/%d/",prefix,i); else key[0]=0;

  h[1]=oracle_hash_u32(h[1],(uint32_t)(b.index1));
  if(direct) ORACLE_I32(ORACLE_KEY("%sindex",key),b.index1);

  h[2]=oracle_hash_u32(h[2],(uint32_t)(b.generation));
  if(direct) ORACLE_I32(ORACLE_KEY("%sgeneration",key),b.generation);

  h[3]=oracle_hash_u32(h[3],(b.world0 == w->worldId) ? 1u : 0u);
  if(direct) ORACLE_BOOL(ORACLE_KEY("%sworld",key),b.world0 == w->worldId);

  h[4]=oracle_hash_pos(h[4],(b2Body_GetPosition(b)));
  if(direct) ORACLE_POS(ORACLE_KEY("%sposition",key),b2Body_GetPosition(b));

  h[5]=oracle_hash_rot(h[5],(b2Body_GetRotation(b)));
  if(direct) ORACLE_ROT(ORACLE_KEY("%srotation",key),b2Body_GetRotation(b));

  h[6]=dynamics_hash_vec(h[6],(b2Body_GetLinearVelocity(b)));
  if(direct) ORACLE_VEC2(ORACLE_KEY("%svelocity",key),b2Body_GetLinearVelocity(b));

  h[7]=oracle_hash_f32(h[7],(b2Body_GetAngularVelocity(b)));
  if(direct) ORACLE_F32(ORACLE_KEY("%sangularVelocity",key),b2Body_GetAngularVelocity(b));

  h[8]=oracle_hash_u32(h[8],(b2Body_IsAwake(b)) ? 1u : 0u);
  if(direct) ORACLE_BOOL(ORACLE_KEY("%sawake",key),b2Body_IsAwake(b));

  h[9]=oracle_hash_u32(h[9],(b2Body_IsEnabled(b)) ? 1u : 0u);
  if(direct) ORACLE_BOOL(ORACLE_KEY("%senabled",key),b2Body_IsEnabled(b));

  h[10]=oracle_hash_u32(h[10],(uint32_t)(b2Body_GetType(b)));
  if(direct) ORACLE_I32(ORACLE_KEY("%stype",key),b2Body_GetType(b));
  int capacity=b2Body_GetContactCapacity(b);
  b2ContactData contacts[capacity > 0 ? capacity : 1];
  int count=b2Body_GetContactData(b,contacts,capacity);
  for(int c=0;c<count;++c) {
   char ck[256]; if(direct) snprintf(ck,sizeof(ck),"%scontact/%d/",key,c); else ck[0]=0;
   dynamics_contact(ck,contacts[c].contactId,contacts[c].shapeIdA,contacts[c].shapeIdB,w->worldId,h,direct);
   b2Manifold* m=&contacts[c].manifold;

   h[11]=dynamics_hash_vec(h[11],(m->normal));
   if(direct) ORACLE_VEC2(ORACLE_KEY("%snormal",ck),m->normal);

   h[12]=oracle_hash_f32(h[12],(m->rollingImpulse));
   if(direct) ORACLE_F32(ORACLE_KEY("%srollingImpulse",ck),m->rollingImpulse);

   h[13]=oracle_hash_u32(h[13],(uint32_t)(m->pointCount));
   if(direct) ORACLE_I32(ORACLE_KEY("%spointCount",ck),m->pointCount);
   for(int j=0;j<m->pointCount;++j) {
    char pk[256]; if(direct) snprintf(pk,sizeof(pk),"%spoint/%d/",ck,j); else pk[0]=0;
    b2ManifoldPoint* p=&m->points[j];

    h[14]=dynamics_hash_vec(h[14],(p->anchorA));
    if(direct) ORACLE_VEC2(ORACLE_KEY("%sanchorA",pk),p->anchorA);

    h[15]=dynamics_hash_vec(h[15],(p->anchorB));
    if(direct) ORACLE_VEC2(ORACLE_KEY("%sanchorB",pk),p->anchorB);

    h[16]=oracle_hash_f32(h[16],(p->separation));
    if(direct) ORACLE_F32(ORACLE_KEY("%sseparation",pk),p->separation);

    h[17]=oracle_hash_f32(h[17],(p->normalImpulse));
    if(direct) ORACLE_F32(ORACLE_KEY("%snormalImpulse",pk),p->normalImpulse);

    h[18]=oracle_hash_f32(h[18],(p->tangentImpulse));
    if(direct) ORACLE_F32(ORACLE_KEY("%stangentImpulse",pk),p->tangentImpulse);

    h[19]=oracle_hash_f32(h[19],(p->totalNormalImpulse));
    if(direct) ORACLE_F32(ORACLE_KEY("%stotalNormalImpulse",pk),p->totalNormalImpulse);

    h[20]=oracle_hash_f32(h[20],(p->normalVelocity));
    if(direct) ORACLE_F32(ORACLE_KEY("%snormalVelocity",pk),p->normalVelocity);

    h[21]=oracle_hash_f32(h[21],(p->baseSeparation));
    if(direct) ORACLE_F32(ORACLE_KEY("%sbaseSeparation",pk),p->baseSeparation);

    h[22]=oracle_hash_u32(h[22],(uint32_t)(p->id));
    if(direct) ORACLE_I32(ORACLE_KEY("%sid",pk),p->id);

    h[23]=oracle_hash_u32(h[23],(p->persisted) ? 1u : 0u);
    if(direct) ORACLE_BOOL(ORACLE_KEY("%spersisted",pk),p->persisted);
   }
  }

  h[24]=oracle_hash_u32(h[24],(uint32_t)(count));
  if(direct) ORACLE_I32(ORACLE_KEY("%scontactCount",key),count);
 }
 b2BodyEvents moves=b2World_GetBodyEvents(id);
 for(int i=0;i<moves.moveCount;++i) {
  char key[256]; if(direct) snprintf(key,sizeof(key),"%smove/%d/",prefix,i); else key[0]=0;
  b2BodyMoveEvent* e=&moves.moveEvents[i];

  h[1]=oracle_hash_u32(h[1],(uint32_t)(e->bodyId.index1));
  if(direct) ORACLE_I32(ORACLE_KEY("%sindex",key),e->bodyId.index1);

  h[2]=oracle_hash_u32(h[2],(uint32_t)(e->bodyId.generation));
  if(direct) ORACLE_I32(ORACLE_KEY("%sgeneration",key),e->bodyId.generation);

  h[3]=oracle_hash_u32(h[3],(e->bodyId.world0 == w->worldId) ? 1u : 0u);
  if(direct) ORACLE_BOOL(ORACLE_KEY("%sworld",key),e->bodyId.world0 == w->worldId);

  h[4]=oracle_hash_pos(h[4],(e->transform.p));
  if(direct) ORACLE_POS(ORACLE_KEY("%sposition",key),e->transform.p);

  h[5]=oracle_hash_rot(h[5],(e->transform.q));
  if(direct) ORACLE_ROT(ORACLE_KEY("%srotation",key),e->transform.q);

  h[25]=oracle_hash_u32(h[25],(e->fellAsleep) ? 1u : 0u);
  if(direct) ORACLE_BOOL(ORACLE_KEY("%sfellAsleep",key),e->fellAsleep);
 }

 h[26]=oracle_hash_u32(h[26],(uint32_t)(moves.moveCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%smoveCount",prefix),moves.moveCount);
 b2ContactEvents events=b2World_GetContactEvents(id);
 for(int i=0;i<events.beginCount;++i) {
  char key[256]; if(direct) snprintf(key,sizeof(key),"%sbegin/%d/",prefix,i); else key[0]=0;
  b2ContactBeginTouchEvent* e=&events.beginEvents[i]; dynamics_contact(key,e->contactId,e->shapeIdA,e->shapeIdB,w->worldId,h,direct);
 }

 h[27]=oracle_hash_u32(h[27],(uint32_t)(events.beginCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%sbeginCount",prefix),events.beginCount);
 for(int i=0;i<events.endCount;++i) {
  char key[256]; if(direct) snprintf(key,sizeof(key),"%send/%d/",prefix,i); else key[0]=0;
  b2ContactEndTouchEvent* e=&events.endEvents[i]; dynamics_contact(key,e->contactId,e->shapeIdA,e->shapeIdB,w->worldId,h,direct);
 }

 h[28]=oracle_hash_u32(h[28],(uint32_t)(events.endCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%sendCount",prefix),events.endCount);
 for(int i=0;i<events.hitCount;++i) {
  char key[256]; if(direct) snprintf(key,sizeof(key),"%shit/%d/",prefix,i); else key[0]=0;
  b2ContactHitEvent* e=&events.hitEvents[i]; dynamics_contact(key,e->contactId,e->shapeIdA,e->shapeIdB,w->worldId,h,direct);

  h[29]=oracle_hash_pos(h[29],(e->point));
  if(direct) ORACLE_POS(ORACLE_KEY("%spoint",key),e->point);

  h[11]=dynamics_hash_vec(h[11],(e->normal));
  if(direct) ORACLE_VEC2(ORACLE_KEY("%snormal",key),e->normal);

  h[30]=oracle_hash_f32(h[30],(e->approachSpeed));
  if(direct) ORACLE_F32(ORACLE_KEY("%sapproachSpeed",key),e->approachSpeed);
 }

 h[31]=oracle_hash_u32(h[31],(uint32_t)(events.hitCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%shitCount",prefix),events.hitCount);
 b2SensorEvents sensors=b2World_GetSensorEvents(id);
 for(int i=0;i<sensors.beginCount;++i) {
  char key[256]; if(direct) snprintf(key,sizeof(key),"%ssensorBegin/%d/sensor/",prefix,i); else key[0]=0; dynamics_shape(key,sensors.beginEvents[i].sensorShapeId,w->worldId,h,direct);
  if(direct) snprintf(key,sizeof(key),"%ssensorBegin/%d/visitor/",prefix,i); dynamics_shape(key,sensors.beginEvents[i].visitorShapeId,w->worldId,h,direct);
 }

 h[32]=oracle_hash_u32(h[32],(uint32_t)(sensors.beginCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%ssensorBeginCount",prefix),sensors.beginCount);
 for(int i=0;i<sensors.endCount;++i) {
  char key[256]; if(direct) snprintf(key,sizeof(key),"%ssensorEnd/%d/sensor/",prefix,i); else key[0]=0; dynamics_shape(key,sensors.endEvents[i].sensorShapeId,w->worldId,h,direct);
  if(direct) snprintf(key,sizeof(key),"%ssensorEnd/%d/visitor/",prefix,i); dynamics_shape(key,sensors.endEvents[i].visitorShapeId,w->worldId,h,direct);
 }

 h[33]=oracle_hash_u32(h[33],(uint32_t)(sensors.endCount));
 if(direct) ORACLE_I32(ORACLE_KEY("%ssensorEndCount",prefix),sensors.endCount);
 h[35]=b2World_GetStateHash(id);
 return row;
}
static const char dynamics_schema[] = "bodyCount,index,generation,world,position,rotation,velocity,angularVelocity,awake,enabled,type,normal,rollingImpulse,pointCount,anchorA,anchorB,separation,normalImpulse,tangentImpulse,totalNormalImpulse,normalVelocity,baseSeparation,id,persisted,contactCount,fellAsleep,moveCount,beginCount,endCount,point,approachSpeed,hitCount,sensorBeginCount,sensorEndCount,valid,canonicalState";
static uint8_t dynamics_trace[32 * 1024 * 1024];
static size_t dynamics_trace_size;
static dynamics_row dynamics_previous;
static uint32_t dynamics_run;
static inline void dynamics_trace_flush(void) {
 if(!dynamics_run) return;
 B2_ASSERT(dynamics_trace_size+292<=sizeof(dynamics_trace));
 for(int i=0;i<4;++i) dynamics_trace[dynamics_trace_size++]=(uint8_t)(dynamics_run>>(8*i));
 for(int f=0;f<36;++f) for(int i=0;i<8;++i) dynamics_trace[dynamics_trace_size++]=(uint8_t)(dynamics_previous.fields[f]>>(8*i));
 dynamics_run=0;
}
static inline void dynamics_trace_add(dynamics_row row) {
 bool same=true;
 for(int i=0;i<36;++i) if(row.fields[i]!=dynamics_previous.fields[i]) same=false;
 if(dynamics_run && !same) dynamics_trace_flush();
 dynamics_previous=row; ++dynamics_run;
}
static inline void dynamics_trace_end(void) {
 dynamics_trace_flush();
 ORACLE_BYTES("dynamics/schema",dynamics_schema,sizeof(dynamics_schema)-1);
 ORACLE_BYTES("dynamics/rle-u32-36u64-le",dynamics_trace,dynamics_trace_size);
 dynamics_trace_size=0;
}
static inline void dynamics_short(b2WorldId w,const char* prefix,bool direct) {
 dynamics_row row=dynamics_observe(w,prefix,direct);
 uint8_t bytes[36*8];
 for(int f=0;f<36;++f) for(int i=0;i<8;++i) bytes[f*8+i]=(uint8_t)(row.fields[f]>>(8*i));
 ORACLE_U64(ORACLE_KEY("%scanonicalState",prefix),row.fields[35]);
 ORACLE_BYTES(ORACLE_KEY("%sdigests",prefix),bytes,sizeof(bytes));
}
#endif
