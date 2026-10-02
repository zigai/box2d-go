// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "box2d/box2d.h"
#include "body.h"
#include "physics_world.h"
#include <assert.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>

#ifndef WORLD_NO_ORACLE
#include "oracle.h"
static uint8_t world_steps[1048576];
static size_t world_step_size;
static uint32_t world_step_count;
static uint32_t world_run_count;
static uint64_t world_run_hash;
static void world_flush( void )
{
    if ( world_run_count == 0 ) return;
    assert( world_step_size + 12 <= sizeof( world_steps ) );
    for ( int i = 0; i < 4; ++i ) world_steps[world_step_size++] = (uint8_t)(world_run_count >> (8 * i));
    for ( int i = 0; i < 8; ++i ) world_steps[world_step_size++] = (uint8_t)(world_run_hash >> (8 * i));
    world_run_count = 0;
}
static void world_step( b2WorldId worldId, const char* source )
{
    b2World* world = b2GetWorldFromId( worldId );
    uint64_t hash = ORACLE_HASH_INIT;
    for ( int i = 0; i < world->bodies.count; ++i )
    {
        b2Body* body = world->bodies.data + i;
        if ( body->id != i ) continue;
        b2BodySim* sim = b2GetBodySim( world, body );
        hash = oracle_hash_transform( hash, sim->transform.p, sim->transform.q );
    }
    if ( world_run_count != 0 && hash != world_run_hash ) world_flush();
    world_run_hash = hash;
    ++world_run_count;
    ++world_step_count;
    if ( world_step_count % 60 == 0 )
    {
        ORACLE_I32( ORACLE_KEY( "step/%u/%s/bodyCount", world_step_count, source ), b2World_GetCounters( worldId ).bodyCount );
        for ( int i = 0; i < world->bodies.count; ++i )
        {
            b2Body* body = world->bodies.data + i;
            if ( body->id != i ) continue;
            b2BodySim* sim = b2GetBodySim( world, body );
            ORACLE_I32( ORACLE_KEY( "step/%u/%s/body", world_step_count, source ), i );
            ORACLE_POS( ORACLE_KEY( "step/%u/%s/body/%d/position", world_step_count, source, i ), sim->transform.p );
            ORACLE_ROT( ORACLE_KEY( "step/%u/%s/body/%d/rotation", world_step_count, source, i ), sim->transform.q );
        }
    }
}
static void world_sensor( b2SensorEvents events )
{
    ORACLE_I32( "sensor.beginCount", events.beginCount );
    for ( int i = 0; i < events.beginCount; ++i )
    {
        ORACLE_I32( "sensor.begin.sensor", events.beginEvents[i].sensorShapeId.index1 );
        ORACLE_I32( "sensor.begin.visitor", events.beginEvents[i].visitorShapeId.index1 );
    }
    ORACLE_I32( "sensor.endCount", events.endCount );
    for ( int i = 0; i < events.endCount; ++i )
    {
        ORACLE_I32( "sensor.end.sensor", events.endEvents[i].sensorShapeId.index1 );
        ORACLE_I32( "sensor.end.visitor", events.endEvents[i].visitorShapeId.index1 );
    }
}
#define ORACLE_WORLD_SENSOR( events ) world_sensor( events )
#define ORACLE_WORLD_STEP( worldId, source ) world_step( worldId, source )
#else
#include "oracle.h"
#undef ORACLE_BOOL
#undef ORACLE_F32
#undef ORACLE_U64
#undef ORACLE_I32
#undef ORACLE_POS
#undef ORACLE_ROT
#define ORACLE_BOOL( ... ) ((void)0)
#define ORACLE_F32( ... ) ((void)0)
#define ORACLE_U64( ... ) ((void)0)
#define ORACLE_I32( ... ) ((void)0)
#define ORACLE_POS( ... ) ((void)0)
#define ORACLE_ROT( ... ) ((void)0)
#define ORACLE_WORLD_SENSOR( ... ) ((void)0)
#define ORACLE_WORLD_STEP( ... ) ((void)0)
#endif

#include "world/test_world.c"

int main( void )
{
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/HelloWorld" );
#endif
    if ( HelloWorld() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/EmptyWorld" );
#endif
    if ( EmptyWorld() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/DestroyAllBodiesWorld" );
#endif
    if ( DestroyAllBodiesWorld() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/TestIsValid" );
#endif
    if ( TestIsValid() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/TestWorldRecycle" );
#endif
    if ( TestWorldRecycle() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/TestWorldCoverage" );
#endif
    if ( TestWorldCoverage() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/TestSensor" );
#endif
    if ( TestSensor() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/TestSetWorkerCount" );
#endif
    if ( TestSetWorkerCount() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/ChainSegmentShapeTest" );
#endif
    if ( ChainSegmentShapeTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/SetBulletDriftTest" );
#endif
    if ( SetBulletDriftTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/DeferredMassFlagSyncTest" );
#endif
    if ( DeferredMassFlagSyncTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/EnableSleepFlagSyncTest" );
#endif
    if ( EnableSleepFlagSyncTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/EnableContactRecyclingTest" );
#endif
    if ( EnableContactRecyclingTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/EnlargedProxyDestroyedTest" );
#endif
    if ( EnlargedProxyDestroyedTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/BroadPhasePairsTest" );
#endif
    if ( BroadPhasePairsTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/BodySimLocatorTest" );
#endif
    if ( BodySimLocatorTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
#ifndef WORLD_NO_ORACLE
    world_step_size = 0;
    world_step_count = 0;
    world_run_count = 0;
    ORACLE_OPEN( "world/NonDynamicJointDisableTest" );
#endif
    if ( NonDynamicJointDisableTest() != 0 ) return 1;
#ifndef WORLD_NO_ORACLE
    world_flush();
    ORACLE_BYTES( "steps/rle-u32-u64-le", world_steps, world_step_size );
    ORACLE_CLOSE();
#endif
    return 0;
}
