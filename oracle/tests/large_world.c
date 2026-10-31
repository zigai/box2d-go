// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "body.h"
#include "physics_world.h"
#include "oracle.h"

#include <stdio.h>

static void transformTrace( b2WorldId worldId, const char* key, int step )
{
	char traceKey[256];
	snprintf( traceKey, sizeof( traceKey ), "%s", key );
	b2World* world = b2GetWorldFromId( worldId );
	bool full = step == 0 || ( step + 1 ) % 60 == 0;
	int count = 0;
	for ( int i = 0; i < world->bodies.count; ++i )
	{
		if ( world->bodies.data[i].id == i )
		{
			count += 1;
		}
	}
	if ( full )
	{
		ORACLE_I32( ORACLE_KEY( "%s/count", traceKey ), count );
	}
	uint64_t hash = ORACLE_HASH_INIT;
	for ( int i = 0; i < world->bodies.count; ++i )
	{
		b2Body* body = world->bodies.data + i;
		if ( body->id != i )
		{
			continue;
		}
		b2WorldTransform xf = b2GetBodySim( world, body )->transform;
		hash = oracle_hash_transform( hash, xf.p, xf.q );
		if ( full )
		{
			ORACLE_I32( ORACLE_KEY( "%s/body/%d/id", traceKey, i ), i );
			ORACLE_POS( ORACLE_KEY( "%s/body/%d/position", traceKey, i ), xf.p );
			ORACLE_ROT( ORACLE_KEY( "%s/body/%d/rotation", traceKey, i ), xf.q );
		}
	}
	ORACLE_U64( traceKey, hash );
}

#define ORACLE_TRANSFORM_TRACE( world, key, step ) transformTrace( ( world ), ( key ), ( step ) )

#include "upstream/test_large_world.c"

int main( void )
{
	return LargeWorldTest();
}
