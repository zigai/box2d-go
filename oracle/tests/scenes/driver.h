// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "box2d/box2d.h"
#include "physics_world.h"
#include "body.h"
#include <stdlib.h>
#include <string.h>

#if defined(SCENE_NO_ORACLE)
#define ORACLE_OPEN(...) ((void)0)
#define ORACLE_CLOSE(...) ((void)0)
#define ORACLE_I32(key, value) ((void)(key), (void)(value))
#define ORACLE_U32(key, value) ((void)(key), (void)(value))
#define ORACLE_U64(key, value) ((void)(key), (void)(value))
#define ORACLE_F32(key, value) ((void)(key), (void)(value))
#define ORACLE_VEC2(key, value) ((void)(key), (void)(value))
#define ORACLE_POS(key, value) ((void)(key), (void)(value))
#define ORACLE_ROT(key, value) ((void)(key), (void)(value))
#define ORACLE_BOOL(key, value) ((void)(key), (void)(value))
#define ORACLE_BYTES(key, data, size) ((void)(key), (void)(data), (void)(size))
#else
#include "oracle.h"
#endif

#include <stdio.h>

static const char* scene_key(int step, const char* suffix)
{
 static char key[64];
 snprintf(key, sizeof(key), "%d.%s", step, suffix);
 return key;
}

static void scene_u32(uint8_t* bytes, uint32_t value)
{
 for (int i = 0; i < 4; ++i) bytes[i] = (uint8_t)(value >> (8*i));
}

static size_t scene_transform(uint8_t* bytes, b2WorldTransform transform)
{
 uint32_t word;
#if defined(BOX2D_DOUBLE_PRECISION)
 uint64_t position;
 memcpy(&position, &transform.p.x, 8);
 scene_u32(bytes, (uint32_t)position);
 scene_u32(bytes+4, (uint32_t)(position >> 32));
 memcpy(&position, &transform.p.y, 8);
 scene_u32(bytes+8, (uint32_t)position);
 scene_u32(bytes+12, (uint32_t)(position >> 32));
 const int offset = 16;
#else
 memcpy(&word, &transform.p.x, 4);
 scene_u32(bytes, word);
 memcpy(&word, &transform.p.y, 4);
 scene_u32(bytes+4, word);
 const int offset = 8;
#endif
 memcpy(&word, &transform.q.c, 4);
 scene_u32(bytes+offset, word);
 memcpy(&word, &transform.q.s, 4);
 scene_u32(bytes+offset+4, word);
 return offset+8;
}

#include "codec.h"
#include "state.h"

static void scene_dump(b2WorldId worldId, int step)
{
 b2Counters counters = b2World_GetCounters(worldId);
 ORACLE_I32(scene_key(step, "b"), counters.bodyCount);
 ORACLE_I32(scene_key(step, "s"), counters.shapeCount);
 b2World* world = b2GetWorldFromId(worldId);
 uint64_t hash = UINT64_C(14695981039346656037);
 size_t stride = sizeof(b2Pos)+12;
 uint8_t* full = malloc((size_t)counters.bodyCount*stride);
 size_t size = 0;
 for (int i = 0; i < world->bodies.count; ++i)
 {
  b2Body* body = world->bodies.data+i;
  if (body->setIndex == B2_NULL_INDEX) continue;
  b2WorldTransform transform = b2GetBodyTransformQuick(world, body);
  scene_u32(full+size, (uint32_t)i);
  size += 4;
  size_t count = scene_transform(full+size, transform);
  for (size_t j = 0; j < count; ++j)
  {
   hash ^= full[size+j];
   hash *= UINT64_C(1099511628211);
  }
  size += count;
 }
 ORACLE_U64(scene_key(step, "h"), hash);
 if (step == -1 || (step < 300 && step%60 == 59)) scene_full(step, full, size);
 free(full);
 scene_state(worldId, step);
}
