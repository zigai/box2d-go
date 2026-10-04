// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT
#include "contact.h"
#include "joint.h"
#include "shape.h"

static uint64_t scenes_digest;
static void scenes_hash_bytes(const uint8_t* p, size_t n) { for(size_t i=0;i<n;++i) { scenes_digest ^= p[i]; scenes_digest *= UINT64_C(1099511628211); } }
static void scenes_hash_u32(uint32_t v) { uint8_t p[4]; scenes_u32(p,v); scenes_hash_bytes(p,4); }
static void scenes_hash_f32(float v) { uint32_t u; memcpy(&u,&v,4); scenes_hash_u32(u); }
static void scenes_hash_vec(b2Vec2 v) { scenes_hash_f32(v.x); scenes_hash_f32(v.y); }
static void scenes_hash_frame(b2Transform v) { scenes_hash_vec(v.p); scenes_hash_f32(v.q.c); scenes_hash_f32(v.q.s); }
static void scenes_hash_id(int index1,uint32_t generation) { scenes_hash_u32((uint32_t)index1); scenes_hash_u32(generation); }
static void scenes_hash_manifold(b2Manifold m,bool hitEvents) {
 scenes_hash_u32(hitEvents);scenes_hash_vec(m.normal); scenes_hash_f32(m.rollingImpulse); scenes_hash_u32((uint32_t)m.pointCount);
 for(int i=0;i<m.pointCount;++i) { b2ManifoldPoint p=m.points[i]; scenes_hash_vec(p.anchorA); scenes_hash_vec(p.anchorB); scenes_hash_f32(p.separation); scenes_hash_f32(p.normalImpulse); scenes_hash_f32(p.tangentImpulse); scenes_hash_f32(p.totalNormalImpulse); scenes_hash_f32(p.baseSeparation); scenes_hash_u32(p.id); scenes_hash_u32(p.persisted); if(hitEvents)scenes_hash_f32(p.normalVelocity); }
}
static void scenes_observation(int step,const char* lane,int count) {
 char key[64]; snprintf(key,sizeof(key),"%d.%s.count",step,lane); ORACLE_I32(key,count);
 snprintf(key,sizeof(key),"%d.%s.hash",step,lane); ORACLE_U64(key,scenes_digest);
}

// Event records use explicit little-endian fields, sorted as a multiset. Sorting
// removes scheduling order only, preserving payloads, generations and duplicates.
static size_t scenes_event_width;
static int scenes_event_compare(const void* a,const void* b) { return memcmp(a,b,scenes_event_width); }
static void scenes_event_lane(int step,const char* lane,uint8_t* records,int count,size_t width) {
 scenes_event_width=width; if(count>1) qsort(records,(size_t)count,width,scenes_event_compare);
 scenes_digest=UINT64_C(14695981039346656037); scenes_hash_bytes(records,(size_t)count*width); scenes_observation(step,lane,count); free(records);
}
static void scenes_event_id(uint8_t* p,int index1,uint32_t generation) { scenes_u32(p,(uint32_t)index1); scenes_u32(p+4,generation); }
static void scenes_event_float(uint8_t* p,float v) { uint32_t u; memcpy(&u,&v,4); scenes_u32(p,u); }
static void scenes_event_pos(uint8_t* p,b2Pos v) { b2WorldTransform t={.p=v,.q={0,0}}; scenes_transform(p,t); }
static uint8_t* scenes_event_allocate(int count,size_t width) { uint8_t* p=calloc(count ? (size_t)count : 1,width); if(!p) abort(); return p; }
static void scenes_state(b2WorldId worldId,int step) {
 b2World* world=b2GetWorldFromId(worldId); int count=0;
 scenes_digest=UINT64_C(14695981039346656037);
 for(int i=0;i<world->bodies.count;++i) { b2Body* b=world->bodies.data+i; if(b->setIndex==B2_NULL_INDEX) continue; b2BodyId id=b2MakeBodyId(world,i); scenes_hash_id(id.index1,id.generation); scenes_hash_u32(b2Body_GetType(id)); scenes_hash_vec(b2Body_GetLinearVelocity(id)); scenes_hash_f32(b2Body_GetAngularVelocity(id)); scenes_hash_u32(b2Body_IsAwake(id)); scenes_hash_u32(b2Body_IsEnabled(id)); b2MassData m=b2Body_GetMassData(id); scenes_hash_f32(m.mass); scenes_hash_vec(m.center); scenes_hash_f32(m.rotationalInertia); ++count; }
 scenes_observation(step,"dynamic",count); count=0; scenes_digest=UINT64_C(14695981039346656037);
 for(int i=0;i<world->contacts.count;++i) { b2Contact* c=world->contacts.data+i; if(c->contactId==B2_NULL_INDEX) continue; scenes_hash_id(i+1,c->generation); scenes_hash_id(c->shapeIdA+1,world->shapes.data[c->shapeIdA].generation); scenes_hash_id(c->shapeIdB+1,world->shapes.data[c->shapeIdB].generation); scenes_hash_manifold(b2GetContactSim(world,c)->manifold,world->shapes.data[c->shapeIdA].enableHitEvents || world->shapes.data[c->shapeIdB].enableHitEvents); ++count; }
 scenes_observation(step,"contact",count); count=0; scenes_digest=UINT64_C(14695981039346656037);
 for(int i=0;i<world->joints.count;++i) { b2Joint* j=world->joints.data+i; if(j->setIndex==B2_NULL_INDEX) continue; b2JointId id={i+1,world->worldId,j->generation}; scenes_hash_id(id.index1,id.generation); scenes_hash_u32(b2Joint_GetType(id)); b2BodyId a=b2Joint_GetBodyA(id),b=b2Joint_GetBodyB(id); scenes_hash_id(a.index1,a.generation); scenes_hash_id(b.index1,b.generation); scenes_hash_frame(b2Joint_GetLocalFrameA(id)); scenes_hash_frame(b2Joint_GetLocalFrameB(id)); scenes_hash_vec(b2Joint_GetConstraintForce(id)); scenes_hash_f32(b2Joint_GetConstraintTorque(id)); if(b2Joint_GetType(id)==b2_revoluteJoint) { scenes_hash_u32(b2RevoluteJoint_IsMotorEnabled(id)); scenes_hash_f32(b2RevoluteJoint_GetMotorSpeed(id)); scenes_hash_f32(b2RevoluteJoint_GetMaxMotorTorque(id)); scenes_hash_f32(b2RevoluteJoint_GetMotorTorque(id)); scenes_hash_u32(b2RevoluteJoint_IsSpringEnabled(id)); scenes_hash_f32(b2RevoluteJoint_GetSpringHertz(id)); scenes_hash_f32(b2RevoluteJoint_GetSpringDampingRatio(id)); scenes_hash_u32(b2RevoluteJoint_IsLimitEnabled(id)); scenes_hash_f32(b2RevoluteJoint_GetLowerLimit(id)); scenes_hash_f32(b2RevoluteJoint_GetUpperLimit(id)); } ++count; }
 scenes_observation(step,"joint",count);
 b2BodyEvents moves=b2World_GetBodyEvents(worldId); size_t width=8+sizeof(b2Pos)+8+8; uint8_t* records=scenes_event_allocate(moves.moveCount,width);
 for(int i=0;i<moves.moveCount;++i) { b2BodyMoveEvent e=moves.moveEvents[i]; uint8_t* p=records+(size_t)i*width; scenes_event_id(p,e.bodyId.index1,e.bodyId.generation); scenes_transform(p+8,e.transform); scenes_u32(p+width-8,e.fellAsleep); scenes_u32(p+width-4,e.userData==NULL); }
 scenes_event_lane(step,"move",records,moves.moveCount,width);
 b2SensorEvents sensors=b2World_GetSensorEvents(worldId); width=16;
 records=scenes_event_allocate(sensors.beginCount,width); for(int i=0;i<sensors.beginCount;++i) { b2SensorBeginTouchEvent e=sensors.beginEvents[i]; uint8_t* p=records+(size_t)i*width; scenes_event_id(p,e.sensorShapeId.index1,e.sensorShapeId.generation); scenes_event_id(p+8,e.visitorShapeId.index1,e.visitorShapeId.generation); } scenes_event_lane(step,"sensorBegin",records,sensors.beginCount,width);
 records=scenes_event_allocate(sensors.endCount,width); for(int i=0;i<sensors.endCount;++i) { b2SensorEndTouchEvent e=sensors.endEvents[i]; uint8_t* p=records+(size_t)i*width; scenes_event_id(p,e.sensorShapeId.index1,e.sensorShapeId.generation); scenes_event_id(p+8,e.visitorShapeId.index1,e.visitorShapeId.generation); } scenes_event_lane(step,"sensorEnd",records,sensors.endCount,width);
 b2ContactEvents contacts=b2World_GetContactEvents(worldId); width=24;
 records=scenes_event_allocate(contacts.beginCount,width); for(int i=0;i<contacts.beginCount;++i) { b2ContactBeginTouchEvent e=contacts.beginEvents[i]; uint8_t* p=records+(size_t)i*width; scenes_event_id(p,e.shapeIdA.index1,e.shapeIdA.generation); scenes_event_id(p+8,e.shapeIdB.index1,e.shapeIdB.generation); scenes_event_id(p+16,e.contactId.index1,e.contactId.generation); } scenes_event_lane(step,"contactBegin",records,contacts.beginCount,width);
 records=scenes_event_allocate(contacts.endCount,width); for(int i=0;i<contacts.endCount;++i) { b2ContactEndTouchEvent e=contacts.endEvents[i]; uint8_t* p=records+(size_t)i*width; scenes_event_id(p,e.shapeIdA.index1,e.shapeIdA.generation); scenes_event_id(p+8,e.shapeIdB.index1,e.shapeIdB.generation); scenes_event_id(p+16,e.contactId.index1,e.contactId.generation); } scenes_event_lane(step,"contactEnd",records,contacts.endCount,width);
 width=24+sizeof(b2Pos)+12; records=scenes_event_allocate(contacts.hitCount,width); for(int i=0;i<contacts.hitCount;++i) { b2ContactHitEvent e=contacts.hitEvents[i]; uint8_t* p=records+(size_t)i*width; scenes_event_id(p,e.shapeIdA.index1,e.shapeIdA.generation); scenes_event_id(p+8,e.shapeIdB.index1,e.shapeIdB.generation); scenes_event_id(p+16,e.contactId.index1,e.contactId.generation); scenes_event_pos(p+24,e.point); scenes_event_float(p+24+sizeof(b2Pos),e.normal.x); scenes_event_float(p+28+sizeof(b2Pos),e.normal.y); scenes_event_float(p+32+sizeof(b2Pos),e.approachSpeed); } scenes_event_lane(step,"hit",records,contacts.hitCount,width);
 b2JointEvents joints=b2World_GetJointEvents(worldId); width=12; records=scenes_event_allocate(joints.count,width); for(int i=0;i<joints.count;++i) { b2JointEvent e=joints.jointEvents[i]; scenes_event_id(records+(size_t)i*width,e.jointId.index1,e.jointId.generation); scenes_u32(records+(size_t)i*width+8,e.userData==NULL); } scenes_event_lane(step,"jointEvent",records,joints.count,width);
}
