// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#pragma once

#include "oracle.h"
#include "box2d/box2d.h"
#include "body.h"
#include "physics_world.h"

#include <stdlib.h>
#include <string.h>

typedef struct InputSnapshot
{
	b2WorldTransform* transforms;
	int count;
} InputSnapshot;

static InputSnapshot* inputSnapshots;
static int inputSnapshotCount;

static void inputResetTrace( void )
{
	for ( int i = 0; i < inputSnapshotCount; ++i )
	{
		free( inputSnapshots[i].transforms );
	}
	free( inputSnapshots );
	inputSnapshots = NULL;
	inputSnapshotCount = 0;
}

static bool inputSameTransforms( const InputSnapshot* a, const InputSnapshot* b )
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

static void inputTraceWorld( b2WorldId worldId, const char* key, bool fullState )
{
	b2World* world = b2GetWorldFromId( worldId );
	int capacity = world->bodies.count;
	InputSnapshot snapshot = { calloc( capacity > 0 ? capacity : 1, sizeof( b2WorldTransform ) ), 0 };
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
	if ( fullState )
	{
		int index = 0;
		while ( index < inputSnapshotCount && !inputSameTransforms( &snapshot, inputSnapshots + index ) )
		{
			index += 1;
		}
		ORACLE_I32( "state", index );
		if ( index == inputSnapshotCount )
		{
			InputSnapshot* resized = realloc( inputSnapshots, ( index + 1 ) * sizeof( InputSnapshot ) );
			if ( resized == NULL )
			{
				abort();
			}
			inputSnapshots = resized;
			inputSnapshots[inputSnapshotCount++] = snapshot;
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

#define ORACLE_INPUT_RESET() inputResetTrace()
#define ORACLE_INPUT_TRACE( worldId, key, fullState ) inputTraceWorld( ( worldId ), ( key ), ( fullState ) )
