// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#pragma once

#include "oracle.h"
#include "box2d/box2d.h"
#include "body.h"
#include "physics_world.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct TraceSnapshot
{
	b2WorldTransform* transforms;
	int count;
} TraceSnapshot;

static TraceSnapshot* traceSnapshots;
static int traceSnapshotCount;

static void resetTrace( void )
{
	for ( int i = 0; i < traceSnapshotCount; ++i )
	{
		free( traceSnapshots[i].transforms );
	}
	free( traceSnapshots );
	traceSnapshots = NULL;
	traceSnapshotCount = 0;
}

static bool sameTransforms( const TraceSnapshot* a, const TraceSnapshot* b )
{
	if ( a->count != b->count )
	{
		return false;
	}
	for ( int i = 0; i < a->count; ++i )
	{
		const b2WorldTransform* x = a->transforms + i;
		const b2WorldTransform* y = b->transforms + i;
		if ( memcmp( &x->p.x, &y->p.x, sizeof( x->p.x ) ) != 0 ||
			 memcmp( &x->p.y, &y->p.y, sizeof( x->p.y ) ) != 0 ||
			 memcmp( &x->q.c, &y->q.c, sizeof( x->q.c ) ) != 0 ||
			 memcmp( &x->q.s, &y->q.s, sizeof( x->q.s ) ) != 0 )
		{
			return false;
		}
	}
	return true;
}

static void traceWorld( b2WorldId worldId, const char* key, bool fullState )
{
	// A trace emits more keys than ORACLE_KEY's scratch buffers can retain.
	char ownedKey[256];
	snprintf( ownedKey, sizeof( ownedKey ), "%s", key );
	key = ownedKey;
	b2World* world = b2GetWorldFromId( worldId );
	int capacity = world->bodies.count;
	TraceSnapshot snapshot = { calloc( capacity > 0 ? capacity : 1, sizeof( b2WorldTransform ) ), 0 };
	if ( snapshot.transforms == NULL )
	{
		abort();
	}
	uint64_t hash = ORACLE_HASH_INIT;
	for ( int i = 0; i < capacity; ++i )
	{
		b2Body* body = b2Array_Get( world->bodies, i );
		if ( body->id == B2_NULL_INDEX )
		{
			continue;
		}
		b2WorldTransform transform = b2Body_GetTransform( b2MakeBodyId( world, i ) );
		snapshot.transforms[snapshot.count++] = transform;
		hash = oracle_hash_transform( hash, transform.p, transform.q );
	}
	ORACLE_U64( key, hash );
	ORACLE_U64( ORACLE_KEY( "%s.canonical", key ), b2World_GetStateHash( worldId ) );
	b2Counters counters = b2World_GetCounters( worldId );
	ORACLE_I32( ORACLE_KEY( "%s.bodies", key ), counters.bodyCount );
	ORACLE_I32( ORACLE_KEY( "%s.shapes", key ), counters.shapeCount );
	ORACLE_I32( ORACLE_KEY( "%s.contacts", key ), counters.contactCount );
	ORACLE_I32( ORACLE_KEY( "%s.joints", key ), counters.jointCount );
	b2BodyEvents moves = b2World_GetBodyEvents( worldId );
	for ( int m = 0; fullState && m < moves.moveCount; ++m )
	{
		b2BodyMoveEvent event = moves.moveEvents[m];
		int ordinal = -1;
		int liveIndex = 0;
		for ( int i = 0; i < capacity; ++i )
		{
			b2Body* body = b2Array_Get( world->bodies, i );
			if ( body->id == B2_NULL_INDEX ) { continue; }
			if ( B2_ID_EQUALS( event.bodyId, b2MakeBodyId( world, i ) ) ) { ordinal = liveIndex; break; }
			liveIndex++;
		}
		ORACLE_I32( ORACLE_KEY( "%s.move.body", key ), ordinal );
		ORACLE_BOOL( ORACLE_KEY( "%s.move.sleep", key ), event.fellAsleep );
		ORACLE_POS( ORACLE_KEY( "%s.move.p", key ), event.transform.p );
		ORACLE_ROT( ORACLE_KEY( "%s.move.q", key ), event.transform.q );
	}
	ORACLE_I32( ORACLE_KEY( "%s.move.count", key ), moves.moveCount );
	if ( fullState )
	{
		for ( int i = 0; i < capacity; ++i )
		{
			b2Body* body = b2Array_Get( world->bodies, i );
			if ( body->id == B2_NULL_INDEX ) { continue; }
			b2BodyId id = b2MakeBodyId( world, i );
			ORACLE_VEC2( ORACLE_KEY( "%s.body.velocity", key ), b2Body_GetLinearVelocity( id ) );
			ORACLE_F32( ORACLE_KEY( "%s.body.angular", key ), b2Body_GetAngularVelocity( id ) );
			ORACLE_BOOL( ORACLE_KEY( "%s.body.awake", key ), b2Body_IsAwake( id ) );
			b2MassData mass = b2Body_GetMassData( id );
			ORACLE_F32( ORACLE_KEY( "%s.body.mass", key ), mass.mass );
			ORACLE_VEC2( ORACLE_KEY( "%s.body.center", key ), mass.center );
			ORACLE_F32( ORACLE_KEY( "%s.body.inertia", key ), mass.rotationalInertia );
		}
	}
	b2ContactEvents contacts = b2World_GetContactEvents( worldId );
	b2SensorEvents sensors = b2World_GetSensorEvents( worldId );
	b2JointEvents joints = b2World_GetJointEvents( worldId );
	ORACLE_I32( ORACLE_KEY( "%s.contact.begin", key ), contacts.beginCount );
	ORACLE_I32( ORACLE_KEY( "%s.contact.end", key ), contacts.endCount );
	ORACLE_I32( ORACLE_KEY( "%s.contact.hit", key ), contacts.hitCount );
	ORACLE_I32( ORACLE_KEY( "%s.sensor.begin", key ), sensors.beginCount );
	ORACLE_I32( ORACLE_KEY( "%s.sensor.end", key ), sensors.endCount );
	ORACLE_I32( ORACLE_KEY( "%s.joint.events", key ), joints.count );
	if ( fullState )
	{
		int index = 0;
		while ( index < traceSnapshotCount && !sameTransforms( &snapshot, traceSnapshots + index ) )
		{
			index += 1;
		}
		ORACLE_I32( "state", index );
		if ( index == traceSnapshotCount )
		{
			TraceSnapshot* resized = realloc( traceSnapshots, ( index + 1 ) * sizeof( TraceSnapshot ) );
			if ( resized == NULL )
			{
				abort();
			}
			traceSnapshots = resized;
			traceSnapshots[traceSnapshotCount++] = snapshot;
			ORACLE_I32( ORACLE_KEY( "state.%d.count", index ), snapshot.count );
			for ( int i = 0; i < snapshot.count; ++i )
			{
				ORACLE_POS( ORACLE_KEY( "state.%d.b%d.p", index, i ), snapshot.transforms[i].p );
				ORACLE_ROT( ORACLE_KEY( "state.%d.b%d.q", index, i ), snapshot.transforms[i].q );
			}
			return;
		}
	}
	free( snapshot.transforms );
}

#define ORACLE_TRACE_RESET() resetTrace()
#define ORACLE_TRACE_WORLD( worldId, key, fullState ) traceWorld( ( worldId ), ( key ), ( fullState ) )
