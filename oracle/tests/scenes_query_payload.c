// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT
#include "scenes/benchmarks.c"
#include "scenes/driver.h"
#include "scenes/human.c"
#include "scenes/utils.c"

static void query_pos( b2Pos p )
{
	uint8_t bytes[sizeof( b2Pos ) + 8];
	scene_event_pos( bytes, p );
	scene_hash_bytes( bytes, sizeof( b2Pos ) );
}
static void query_stats( b2TreeStats s )
{
	scene_hash_u32( s.nodeVisits );
	scene_hash_u32( s.leafVisits );
}
static void query_result( b2RayResult r )
{
	scene_hash_u32( r.hit );
	if ( r.hit )
	{
		scene_hash_id( r.shapeId.index1, r.shapeId.generation );
		query_pos( r.point );
		scene_hash_vec( r.normal );
		scene_hash_f32( r.fraction );
	}
	scene_hash_u32( r.nodeVisits );
	scene_hash_u32( r.leafVisits );
}
static void query_emit( int step, int partition, const char* kind, const char* field, uint64_t value )
{
	char key[80];
	snprintf( key, sizeof( key ), "%d.%d.%s.%s", step, partition, kind, field );
	ORACLE_U64( key, value );
}
typedef struct Closest
{
	b2ShapeId id;
	b2Pos point;
	b2Vec2 normal;
	float fraction;
	bool hit;
} Closest;
static float closest( b2ShapeId id, b2Pos point, b2Vec2 normal, float fraction, void* context )
{
	Closest* c = context;
	*c = (Closest){ id, point, normal, fraction, true };
	return fraction;
}
typedef struct Overlaps
{
	uint64_t* ids;
	int count, capacity;
} Overlaps;
static bool overlap( b2ShapeId id, void* context )
{
	Overlaps* o = context;
	if ( o->count == o->capacity )
	{
		o->capacity = o->capacity ? 2 * o->capacity : 64;
		o->ids = realloc( o->ids, (size_t)o->capacity * sizeof( uint64_t ) );
		if ( !o->ids )
			abort();
	}
	o->ids[o->count++] = ( (uint64_t)(uint32_t)id.index1 << 32 ) | id.generation;
	return true;
}
static int compare_ids( const void* a, const void* b )
{
	uint64_t x = *(const uint64_t*)a, y = *(const uint64_t*)b;
	return x < y ? -1 : x > y;
}
static void hash_overlaps( Overlaps* o )
{
	if ( o->count > 1 )
		qsort( o->ids, (size_t)o->count, sizeof( uint64_t ), compare_ids );
	scene_hash_u32( o->count );
	for ( int i = 0; i < o->count; ++i )
		scene_hash_id( (int)( o->ids[i] >> 32 ), (uint32_t)o->ids[i] );
	free( o->ids );
}
static void queries_payload( b2WorldId world, int step )
{
	b2QueryFilter filter = b2DefaultQueryFilter();
	b2Vec2 center = { 0, 0 };
	b2ShapeProxy circle = b2MakeProxy( &center, 1, .1f );
	for ( int f = 0; f < 2; ++f )
	{
		filter.maskBits = f ? 1 : B2_DEFAULT_MASK_BITS;
		for ( int kind = 0; kind < 3; ++kind )
		{
			const char* name = kind == 0 ? "ray" : kind == 1 ? "circle" : "overlap";
			scene_digest = UINT64_C( 14695981039346656037 );
			scene_hash_u32( (uint32_t)filter.maskBits );
			scene_hash_u32( (uint32_t)( filter.maskBits >> 32 ) );
			for ( int i = 0; i < QUERY_COUNT; ++i )
			{
				scene_hash_u32( i );
				query_pos( g_queryBenchmark.origins[i] );
				if ( kind != 2 )
					scene_hash_vec( g_queryBenchmark.translations[i] );
				else
				{
					scene_hash_vec( (b2Vec2){ -5, -5 } );
					scene_hash_vec( (b2Vec2){ 5, 5 } );
				}
				if ( kind == 1 )
					scene_hash_f32( .1f );
			}
			query_emit( step, f, name, "inputs", scene_digest );
			scene_digest = UINT64_C( 14695981039346656037 );
			for ( int i = 0; i < QUERY_COUNT; ++i )
			{
				scene_hash_u32( i );
				if ( kind == 0 )
					query_result(
						b2World_CastRayClosest( world, g_queryBenchmark.origins[i], g_queryBenchmark.translations[i], filter ) );
				else if ( kind == 1 )
				{
					Closest c = { .fraction = 1 };
					b2TreeStats s = b2World_CastShape( world, g_queryBenchmark.origins[i], &circle,
													   g_queryBenchmark.translations[i], filter, closest, &c );
					scene_hash_u32( c.hit );
					if ( c.hit )
					{
						scene_hash_id( c.id.index1, c.id.generation );
						query_pos( c.point );
						scene_hash_vec( c.normal );
						scene_hash_f32( c.fraction );
					}
					query_stats( s );
				}
				else
				{
					Overlaps o = { 0 };
					b2TreeStats s = b2World_OverlapAABB( world, g_queryBenchmark.origins[i], (b2AABB){ { -5, -5 }, { 5, 5 } },
														 filter, overlap, &o );
					hash_overlaps( &o );
					query_stats( s );
				}
			}
			query_emit( step, f, name, "payload", scene_digest );
		}
	}
}
typedef struct TreeClosest
{
	TreeCastContext cast;
	int proxy;
	uint64_t identity;
} TreeClosest;
static float tree_ray( const b2RayCastInput* in, int proxy, uint64_t id, void* context )
{
	TreeClosest* c = context;
	float f = TreeCastRayCallback( in, proxy, id, &c->cast );
	if ( f >= 0 )
	{
		c->proxy = proxy;
		c->identity = id;
	}
	return f;
}
static float tree_box( const b2BoxCastInput* in, int proxy, uint64_t id, void* context )
{
	TreeClosest* c = context;
	float f = TreeCastBoxCallback( in, proxy, id, &c->cast );
	if ( f >= 0 )
	{
		c->proxy = proxy;
		c->identity = id;
	}
	return f;
}
static void tree_payload( int step )
{
	for ( int k = 0; k < TREE_CAST_LENGTH_COUNT; ++k )
		for ( int kind = 0; kind < 3; ++kind )
		{
			uint64_t mask = kind == 1 ? 1 : B2_DEFAULT_MASK_BITS;
			const char* name = kind == 0 ? "rayAll" : kind == 1 ? "rayMask1" : "box";
			scene_digest = UINT64_C( 14695981039346656037 );
			scene_hash_u32( (uint32_t)mask );
			scene_hash_u32( (uint32_t)( mask >> 32 ) );
			for ( int i = 0; i < TREE_CAST_RAY_COUNT; ++i )
			{
				scene_hash_u32( i );
				scene_hash_vec( g_treeCast.origins[i] );
				scene_hash_vec( g_treeCast.translations[k][i] );
				scene_hash_f32( 1 );
				if ( kind == 2 )
					scene_hash_f32( TREE_CAST_BOX_EXTENT );
			}
			query_emit( step, k, name, "inputs", scene_digest );
			scene_digest = UINT64_C( 14695981039346656037 );
			for ( int i = 0; i < TREE_CAST_RAY_COUNT; ++i )
			{
				TreeClosest c = { { g_treeCast.tightBoxes, 1 }, -1, UINT64_MAX };
				b2TreeStats s;
				if ( kind < 2 )
				{
					b2RayCastInput in = { g_treeCast.origins[i], g_treeCast.translations[k][i], 1 };
					s = b2DynamicTree_CastRay( &g_treeCast.tree, &in, mask, tree_ray, &c );
				}
				else
				{
					b2Vec2 o = g_treeCast.origins[i];
					b2BoxCastInput in = { { { o.x - TREE_CAST_BOX_EXTENT, o.y - TREE_CAST_BOX_EXTENT },
											{ o.x + TREE_CAST_BOX_EXTENT, o.y + TREE_CAST_BOX_EXTENT } },
										  g_treeCast.translations[k][i],
										  1 };
					s = b2DynamicTree_CastBox( &g_treeCast.tree, &in, mask, tree_box, &c );
				}
				scene_hash_u32( i );
				scene_hash_u32( (uint32_t)c.proxy );
				scene_hash_u32( (uint32_t)c.identity );
				scene_hash_u32( (uint32_t)( c.identity >> 32 ) );
				scene_hash_f32( c.cast.fraction );
				query_stats( s );
			}
			query_emit( step, k, name, "payload", scene_digest );
		}
}
static void tile_payload( b2WorldId world, int step )
{
	b2QueryFilter filter = b2DefaultQueryFilter();
	for ( int kind = 0; kind < 2; ++kind )
	{
		const char* name = kind == 0 ? "ray" : "overlap";
		uint32_t state = 1234u + 7919u * (uint32_t)step;
		scene_digest = UINT64_C( 14695981039346656037 );
		for ( int i = 0; i < TILE_QUERY_COUNT; ++i )
		{
			b2Pos origin = { g_tileWorld.xLeft + TileRandom( &state, 0, g_tileWorld.groundCount * g_tileWorld.bodySpan ),
							 TileRandom( &state, -2, 18 ) };
			float a = TileRandom( &state, 0, 2 * B2_PI );
			b2Vec2 t = { TILE_RAY_LENGTH * cosf( a ), TILE_RAY_LENGTH * sinf( a ) };
			scene_hash_u32( i );
			query_pos( origin );
			if ( kind == 0 )
				scene_hash_vec( t );
			else
			{
				scene_hash_vec( (b2Vec2){ -5, -5 } );
				scene_hash_vec( (b2Vec2){ 5, 5 } );
			}
		}
		query_emit( step, 0, name, "inputs", scene_digest );
		state = 1234u + 7919u * (uint32_t)step;
		scene_digest = UINT64_C( 14695981039346656037 );
		for ( int i = 0; i < TILE_QUERY_COUNT; ++i )
		{
			b2Pos origin = { g_tileWorld.xLeft + TileRandom( &state, 0, g_tileWorld.groundCount * g_tileWorld.bodySpan ),
							 TileRandom( &state, -2, 18 ) };
			float a = TileRandom( &state, 0, 2 * B2_PI );
			b2Vec2 t = { TILE_RAY_LENGTH * cosf( a ), TILE_RAY_LENGTH * sinf( a ) };
			scene_hash_u32( i );
			if ( kind == 0 )
				query_result( b2World_CastRayClosest( world, origin, t, filter ) );
			else
			{
				Overlaps o = { 0 };
				b2TreeStats s = b2World_OverlapAABB( world, origin, (b2AABB){ { -5, -5 }, { 5, 5 } }, filter, overlap, &o );
				hash_overlaps( &o );
				query_stats( s );
			}
		}
		query_emit( step, 0, name, "payload", scene_digest );
	}
}
int main( void )
{
	for ( int workers = 1; workers <= 4; workers += 3 )
		for ( int kind = 0; kind < 3; ++kind )
		{
			char name[64];
			snprintf( name, sizeof( name ), "scenes/%sPayload_workers%d",
					  kind == 0	  ? "Queries"
					  : kind == 1 ? "TreeCast"
								  : "TileWorld",
					  workers );
			ORACLE_OPEN( name );
			b2WorldDef def = b2DefaultWorldDef();
			def.workerCount = workers;
			b2WorldId world = b2CreateWorld( &def );
			if ( kind == 0 )
				CreateQueries( world );
			else if ( kind == 1 )
				CreateTreeCast( world );
			else
				CreateTileWorld( world );
			for ( int step = 0; step < 300; ++step )
			{
				if ( kind == 0 )
				{
					queries_payload( world, step );
					StepQueries( world, step );
				}
				else if ( kind == 1 )
				{
					tree_payload( step );
					StepTreeCast( world, step );
				}
				else
				{
					StepTileWorld( world, step );
					tile_payload( world, step );
				}
				b2World_Step( world, 1.0f / 60.0f, 4 );
			}
			if ( kind == 1 )
				DestroyTreeCast();
			if ( kind == 2 )
				DestroyTileWorld();
			b2DestroyWorld( world );
			ORACLE_CLOSE();
		}
	return 0;
}
