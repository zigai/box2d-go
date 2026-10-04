// SPDX-License-Identifier: MIT
// Compile and link against the separately configured pinned Debug scalar engine.
#include "oracle.h"

#include "box2d/collision.h"

#include <stdio.h>
#include <string.h>

#ifdef NDEBUG
#error "The distance debug oracle requires the actual pinned Debug engine profile"
#endif

static b2ShapeProxy debug_distance_box( float radius )
{
	const b2Vec2 points[] = { { -1, -1 }, { 1, -1 }, { 1, 1 }, { -1, 1 } };
	return b2MakeProxy( points, 4, radius );
}

static void debug_distance_output( const char* prefix, b2DistanceOutput o, b2SimplexCache cache, const b2Simplex* debug )
{
	char key[256];
	strcpy( key, prefix );
	ORACLE_VEC2( ORACLE_KEY( "%s/pointA", key ), o.pointA );
	ORACLE_VEC2( ORACLE_KEY( "%s/pointB", key ), o.pointB );
	ORACLE_F32( ORACLE_KEY( "%s/distance", key ), o.distance );
	if ( o.distance > 0 )
	{
		ORACLE_VEC2( ORACLE_KEY( "%s/normal", key ), o.normal );
	}
	ORACLE_I32( ORACLE_KEY( "%s/iterations", key ), o.iterations );
	ORACLE_I32( ORACLE_KEY( "%s/simplexCount", key ), o.simplexCount );
	ORACLE_U32( ORACLE_KEY( "%s/cache/count", key ), cache.count );
	for ( int i = 0; i < cache.count; ++i )
	{
		ORACLE_U32( ORACLE_KEY( "%s/cache/%d/indexA", key, i ), cache.indexA[i] );
		ORACLE_U32( ORACLE_KEY( "%s/cache/%d/indexB", key, i ), cache.indexB[i] );
	}
	for ( int i = 0; i < o.simplexCount; ++i )
	{
		const b2Simplex* s = debug + i;
		ORACLE_I32( ORACLE_KEY( "%s/simplex/%d/count", key, i ), s->count );
		const b2SimplexVertex* vertices[] = { &s->v1, &s->v2, &s->v3 };
		for ( int j = 0; j < s->count; ++j )
		{
			const b2SimplexVertex* v = vertices[j];
			char vk[256];
			snprintf( vk, sizeof( vk ), "%s/simplex/%d/vertex/%d", key, i, j );
			ORACLE_VEC2( ORACLE_KEY( "%s/wA", vk ), v->wA );
			ORACLE_VEC2( ORACLE_KEY( "%s/wB", vk ), v->wB );
			ORACLE_VEC2( ORACLE_KEY( "%s/w", vk ), v->w );
			ORACLE_F32( ORACLE_KEY( "%s/a", vk ), v->a );
			ORACLE_I32( ORACLE_KEY( "%s/indexA", vk ), v->indexA );
			ORACLE_I32( ORACLE_KEY( "%s/indexB", vk ), v->indexB );
		}
	}
}

int main( void )
{
	const b2Transform transforms[] = {
		{ { 4, 0.25f }, { 1, 0 } },		   { { 3.75f, 0.5f }, { 1, 0 } }, { { 3.5f, 1 }, { 0.6f, 0.8f } },
		{ { 2, 1 }, { 0.6f, 0.8f } },	   { { 0.5f, 0.25f }, { 1, 0 } }, { { -3, -1 }, { 0, -1 } },
		{ { -3.25f, -0.75f }, { 0, -1 } }, { { 4, 0.25f }, { 1, 0 } },
	};
	b2DistanceInput input = { .proxyA = debug_distance_box( 0.125f ), .proxyB = debug_distance_box( 0.25f ), .useRadii = true };
	b2SimplexCache cache = { 0 };
	ORACLE_OPEN( "distance_debug_public/history" );
	for ( int i = 0; i < 8; ++i )
	{
		input.transform = transforms[i];
		b2Simplex debug[24];
		b2DistanceOutput o = b2ShapeDistance( &input, &cache, debug, 24 );
		debug_distance_output( ORACLE_KEY( "%d", i ), o, cache, debug );
	}
	ORACLE_CLOSE();

	ORACLE_OPEN( "distance_debug_public/capacity-one" );
	const b2Vec2 segment[] = { { 0, -1 }, { 0, 1 } };
	input = (b2DistanceInput){
		.proxyA = debug_distance_box( 0 ), .proxyB = b2MakeProxy( segment, 2, 0 ), .transform = { { 3, 0 }, { 1, 0 } } };
	cache = (b2SimplexCache){ 0 };
	b2Simplex debug[1];
	b2DistanceOutput o = b2ShapeDistance( &input, &cache, debug, 1 );
	debug_distance_output( "output", o, cache, debug );
	ORACLE_CLOSE();
	return 0;
}
