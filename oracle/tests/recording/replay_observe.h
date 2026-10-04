// SPDX-License-Identifier: MIT
#pragma once
#include "replay.h"

static void replay_shape_id( const char* key, b2ShapeId id, b2WorldId world )
{
	char ownedKey[256];
	snprintf( ownedKey, sizeof( ownedKey ), "%s", key );
	key = ownedKey;
	ORACLE_I32( ORACLE_KEY( "%s/index", key ), id.index1 );
	ORACLE_U32( ORACLE_KEY( "%s/generation", key ), id.generation );
	ORACLE_BOOL( ORACLE_KEY( "%s/owner", key ), id.index1 == 0 || id.world0 == world.index1 - 1 );
}

static uint64_t replay_draw_hash;
static int replay_draw_count;
static void replay_draw_line( b2Pos a, b2Pos b, b2HexColor c, void* context )
{
	(void)context;
	++replay_draw_count;
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, 1 );
	replay_draw_hash = oracle_hash_pos( replay_draw_hash, a );
	replay_draw_hash = oracle_hash_pos( replay_draw_hash, b );
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, c );
}
static void replay_draw_point( b2Pos p, float size, b2HexColor c, void* context )
{
	(void)context;
	++replay_draw_count;
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, 2 );
	replay_draw_hash = oracle_hash_pos( replay_draw_hash, p );
	replay_draw_hash = oracle_hash_f32( replay_draw_hash, size );
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, c );
}
static void replay_draw_polygon( b2WorldTransform xf, const b2Vec2* vertices, int count, b2HexColor c, void* context )
{
	(void)context;
	++replay_draw_count;
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, 3 );
	replay_draw_hash = oracle_hash_transform( replay_draw_hash, xf.p, xf.q );
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, count );
	for ( int i = 0; i < count; ++i )
	{
		replay_draw_hash = oracle_hash_f32( replay_draw_hash, vertices[i].x );
		replay_draw_hash = oracle_hash_f32( replay_draw_hash, vertices[i].y );
	}
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, c );
}
static void replay_draw_capsule( b2Pos a, b2Pos b, float radius, b2HexColor c, void* context )
{
	(void)context;
	++replay_draw_count;
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, 4 );
	replay_draw_hash = oracle_hash_pos( replay_draw_hash, a );
	replay_draw_hash = oracle_hash_pos( replay_draw_hash, b );
	replay_draw_hash = oracle_hash_f32( replay_draw_hash, radius );
	replay_draw_hash = oracle_hash_u32( replay_draw_hash, c );
}
static void replay_observe_queries( b2Replay* player, const char* prefix )
{
	char ownedPrefix[256];
	snprintf( ownedPrefix, sizeof( ownedPrefix ), "%s", prefix );
	prefix = ownedPrefix;
	b2WorldId world = b2Replay_GetWorldId( player );
	int count = b2Replay_GetFrameQueryCount( player );
	ORACLE_I32( ORACLE_KEY( "%s/queries", prefix ), count );
	for ( int i = -1; i <= count; ++i )
	{
		b2ReplayQueryInfo q = b2Replay_GetFrameQuery( player, i );
		ORACLE_I32( ORACLE_KEY( "%s/q%d/type", prefix, i ), q.type );
		ORACLE_U64( ORACLE_KEY( "%s/q%d/category", prefix, i ), q.filter.categoryBits );
		ORACLE_U64( ORACLE_KEY( "%s/q%d/mask", prefix, i ), q.filter.maskBits );
		ORACLE_VEC2( ORACLE_KEY( "%s/q%d/lower", prefix, i ), q.aabb.lowerBound );
		ORACLE_VEC2( ORACLE_KEY( "%s/q%d/upper", prefix, i ), q.aabb.upperBound );
		ORACLE_POS( ORACLE_KEY( "%s/q%d/origin", prefix, i ), q.origin );
		ORACLE_VEC2( ORACLE_KEY( "%s/q%d/translation", prefix, i ), q.translation );
		replay_shape_id( ORACLE_KEY( "%s/q%d/shape", prefix, i ), q.shape, world );
		ORACLE_I32( ORACLE_KEY( "%s/q%d/hits", prefix, i ), q.hitCount );
		for ( int j = -1; j <= q.hitCount; ++j )
		{
			bool validHit = i >= 0 && i < count && j >= 0 && j < q.hitCount;
			bool identityOnly = validHit && ( q.type == b2_replayQueryOverlapAABB || q.type == b2_replayQueryOverlapShape ||
											  q.type == b2_replayQueryCollideMover );
			// Native scratch leaves the cast point/normal/fraction indeterminate for
			// overlaps and mover planes. Do not invoke the getter that reads those fields.
			b2ReplayQueryHit h = { 0 };
			if ( identityOnly )
			{
				h.shape = player->frameHits[player->frameQueries[i].hitStart + j].id;
			}
			else
			{
				h = b2Replay_GetFrameQueryHit( player, i, j );
			}
			replay_shape_id( ORACLE_KEY( "%s/q%d/h%d/shape", prefix, i, j ), h.shape, world );
			if ( !identityOnly )
			{
				ORACLE_POS( ORACLE_KEY( "%s/q%d/h%d/point", prefix, i, j ), h.point );
				ORACLE_VEC2( ORACLE_KEY( "%s/q%d/h%d/normal", prefix, i, j ), h.normal );
				ORACLE_F32( ORACLE_KEY( "%s/q%d/h%d/fraction", prefix, i, j ), h.fraction );
			}
			if ( validHit && q.type == b2_replayQueryCollideMover )
			{
				b2PlaneResult plane = player->frameHits[player->frameQueries[i].hitStart + j].plane;
				ORACLE_VEC2( ORACLE_KEY( "%s/q%d/h%d/planeNormal", prefix, i, j ), plane.plane.normal );
				ORACLE_F32( ORACLE_KEY( "%s/q%d/h%d/planeOffset", prefix, i, j ), plane.plane.offset );
				ORACLE_VEC2( ORACLE_KEY( "%s/q%d/h%d/planePoint", prefix, i, j ), plane.point );
				ORACLE_BOOL( ORACLE_KEY( "%s/q%d/h%d/planeHit", prefix, i, j ), plane.hit );
			}
		}
		if ( i >= 0 && i < count )
		{
			b2RecDrawQuery* internal = player->frameQueries + i;
			ORACLE_BOOL( ORACLE_KEY( "%s/q%d/bool", prefix, i ), internal->boolResult );
			ORACLE_F32( ORACLE_KEY( "%s/q%d/castFraction", prefix, i ), internal->castFraction );
			ORACLE_BOOL( ORACLE_KEY( "%s/q%d/castHit", prefix, i ), internal->castOut.hit );
			ORACLE_POS( ORACLE_KEY( "%s/q%d/castPoint", prefix, i ), internal->castOut.point );
			ORACLE_VEC2( ORACLE_KEY( "%s/q%d/castNormal", prefix, i ), internal->castOut.normal );
			ORACLE_F32( ORACLE_KEY( "%s/q%d/castOutputFraction", prefix, i ), internal->castOut.fraction );
			ORACLE_I32( ORACLE_KEY( "%s/q%d/castIterations", prefix, i ), internal->castOut.iterations );
		}
	}
	b2DebugDraw draw = b2DefaultDebugDraw();
	draw.DrawLineFcn = replay_draw_line;
	draw.DrawPointFcn = replay_draw_point;
	draw.DrawPolygonFcn = replay_draw_polygon;
	draw.DrawSolidCapsuleFcn = replay_draw_capsule;
	for ( int selected = -1; selected <= count; ++selected )
	{
		replay_draw_hash = ORACLE_HASH_INIT;
		replay_draw_count = 0;
		b2Replay_DrawFrameQueries( player, &draw, selected );
		ORACLE_I32( ORACLE_KEY( "%s/draw%d/count", prefix, selected ), replay_draw_count );
		ORACLE_U64( ORACLE_KEY( "%s/draw%d/hash", prefix, selected ), replay_draw_hash );
	}
}

static void replay_observe( b2Replay* player, const char* prefix )
{
	char ownedPrefix[256];
	snprintf( ownedPrefix, sizeof( ownedPrefix ), "%s", prefix );
	prefix = ownedPrefix;
	b2ReplayInfo info = b2Replay_GetInfo( player );
	b2WorldId world = b2Replay_GetWorldId( player );
	ORACLE_I32( ORACLE_KEY( "%s/frame", prefix ), b2Replay_GetFrame( player ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/end", prefix ), b2Replay_IsAtEnd( player ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/diverged", prefix ), b2Replay_HasDiverged( player ) );
	ORACLE_I32( ORACLE_KEY( "%s/divergeFrame", prefix ), b2Replay_GetDivergeFrame( player ) );
	ORACLE_U64( ORACLE_KEY( "%s/state", prefix ), b2World_GetStateHash( world ) );
	ORACLE_I32( ORACLE_KEY( "%s/info/frames", prefix ), info.frameCount );
	ORACLE_I32( ORACLE_KEY( "%s/info/workers", prefix ), info.workerCount );
	ORACLE_F32( ORACLE_KEY( "%s/info/dt", prefix ), info.timeStep );
	ORACLE_I32( ORACLE_KEY( "%s/info/substeps", prefix ), info.subStepCount );
	ORACLE_F32( ORACLE_KEY( "%s/info/scale", prefix ), info.lengthScale );
	ORACLE_VEC2( ORACLE_KEY( "%s/info/lower", prefix ), info.bounds.lowerBound );
	ORACLE_VEC2( ORACLE_KEY( "%s/info/upper", prefix ), info.bounds.upperBound );
	int count = b2Replay_GetBodyCount( player );
	ORACLE_I32( ORACLE_KEY( "%s/bodies", prefix ), count );
	for ( int i = -1; i <= count + 1; ++i )
	{
		b2BodyId id = b2Replay_GetBodyId( player, i );
		ORACLE_I32( ORACLE_KEY( "%s/body%d/index", prefix, i ), id.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/body%d/generation", prefix, i ), id.generation );
		ORACLE_BOOL( ORACLE_KEY( "%s/body%d/valid", prefix, i ), b2Body_IsValid( id ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/body%d/owner", prefix, i ),
					 id.index1 == 0 || b2StoreWorldId( b2Body_GetWorld( id ) ) == b2StoreWorldId( world ) );
		if ( b2Body_IsValid( id ) )
		{
			const char* name = b2Body_GetName( id );
			ORACLE_BYTES( ORACLE_KEY( "%s/body%d/name", prefix, i ), name, strlen( name ) );
		}
	}
	replay_observe_queries( player, prefix );
}
