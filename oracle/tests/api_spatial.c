// SPDX-License-Identifier: MIT
#include "oracle.h"

#include "box2d/box2d.h"

#include <float.h>
#include <stdlib.h>

static void box( const char* key, b2AABB a )
{
	ORACLE_VEC2( ORACLE_KEY( "%s/lower", key ), a.lowerBound );
	ORACLE_VEC2( ORACLE_KEY( "%s/upper", key ), a.upperBound );
}
static void stats( b2TreeStats s )
{
	ORACLE_I32( "nodes", s.nodeVisits );
	ORACLE_I32( "leaves", s.leafVisits );
}
typedef struct TreeHits
{
	int count;
	int ids[32];
	uint64_t data[32];
	bool stop;
	int policy;
} TreeHits;
static bool tq( int id, uint64_t data, void* ptr )
{
	TreeHits* h = ptr;
	h->ids[h->count] = id;
	h->data[h->count++] = data;
	return !h->stop;
}
static void treeHits( TreeHits* h )
{
	if ( !h->stop )
		for ( int i = 1; i < h->count; ++i )
			for ( int j = i; j > 0 && h->ids[j] < h->ids[j - 1]; --j )
			{
				int id = h->ids[j];
				h->ids[j] = h->ids[j - 1];
				h->ids[j - 1] = id;
				uint64_t data = h->data[j];
				h->data[j] = h->data[j - 1];
				h->data[j - 1] = data;
			}
	ORACLE_I32( "count", h->count );
	for ( int i = 0; i < h->count; ++i )
	{
		ORACLE_I32( "id", h->ids[i] );
		ORACLE_U64( "data", h->data[i] );
	}
}
static float tc( float fraction, int id, uint64_t data, TreeHits* h )
{
	float result = h->policy == 0 ? fraction : h->policy == 1 ? -1.0f : h->policy == 2 ? 0.0f : 0.5f;
	ORACLE_I32( "id", id );
	ORACLE_U64( "data", data );
	ORACLE_F32( "incoming", fraction );
	ORACLE_F32( "return", result );
	++h->count;
	return result;
}
static float tr( const b2RayCastInput* in, int id, uint64_t data, void* ptr )
{
	ORACLE_VEC2( "origin", in->origin );
	ORACLE_VEC2( "translation", in->translation );
	return tc( in->maxFraction, id, data, ptr );
}
static float tb( const b2BoxCastInput* in, int id, uint64_t data, void* ptr )
{
	box( "input", in->box );
	ORACLE_VEC2( "translation", in->translation );
	return tc( in->maxFraction, id, data, ptr );
}
static void treeObserve( b2DynamicTree* tree, int* ids, bool* live )
{
	ORACLE_I32( "proxies", b2DynamicTree_GetProxyCount( tree ) );
	ORACLE_I32( "height", b2DynamicTree_GetHeight( tree ) );
	ORACLE_F32( "ratio", b2DynamicTree_GetAreaRatio( tree ) );
	box( "root", b2DynamicTree_GetRootBounds( tree ) );
	for ( int i = 0; i < 4; ++i )
		if ( live[i] )
		{
			ORACLE_I32( "proxy", ids[i] );
			ORACLE_U64( "category", b2DynamicTree_GetCategoryBits( tree, ids[i] ) );
			ORACLE_U64( "userdata", b2DynamicTree_GetUserData( tree, ids[i] ) );
			box( "aabb", b2DynamicTree_GetAABB( tree, ids[i] ) );
		}
	b2AABB queries[3] = { { { -10, -10 }, { 10, 10 } }, { { -.25f, -.25f }, { .25f, .25f } }, { { 20, 20 }, { 21, 21 } } };
	uint64_t masks[4] = { UINT64_MAX, 1, 2, UINT64_C( 1 ) << 63 };
	for ( int q = 0; q < 3; ++q )
		for ( int m = 0; m < 4; ++m )
		{
			TreeHits h = { 0 };
			stats( b2DynamicTree_Query( tree, queries[q], masks[m], tq, &h ) );
			treeHits( &h );
		}
	for ( int stop = 0; stop < 2; ++stop )
	{
		TreeHits h = { .stop = stop };
		stats( b2DynamicTree_QueryAll( tree, queries[0], tq, &h ) );
		treeHits( &h );
	}
}
static void treeHistory( void )
{
	ORACLE_OPEN( "spatial/PublicTreeHistory" );
	b2DynamicTree tree = b2CreateDynamicTree( 4 );
	int ids[4] = { 0 };
	bool live[4] = { 0 };
	treeObserve( &tree, ids, live );
	for ( int i = 0; i < 4; ++i )
	{
		b2AABB a = { { (float)i - .5f, -.5f }, { (float)i + .5f, .5f } };
		ids[i] = b2CreateTreeProxy( &tree, a, i == 3 ? UINT64_C( 1 ) << 63 : UINT64_C( 1 ) << i, UINT64_C( 100 ) + i );
		live[i] = true;
		treeObserve( &tree, ids, live );
	}
	b2DynamicTree_MoveProxy( &tree, ids[0], (b2AABB){ { 4, -1 }, { 5, 1 } } );
	treeObserve( &tree, ids, live );
	b2DynamicTree_SetCategoryBits( &tree, ids[2], ( UINT64_C( 1 ) << 63 ) | 1 );
	treeObserve( &tree, ids, live );
	b2DestroyTreeProxy( &tree, ids[1] );
	live[1] = false;
	treeObserve( &tree, ids, live );
	ids[1] = b2CreateTreeProxy( &tree, (b2AABB){ { -.5f, -.5f }, { .5f, .5f } }, 2, 999 );
	live[1] = true;
	treeObserve( &tree, ids, live );
	for ( int i = 0; i < 4; ++i )
	{
		b2DestroyTreeProxy( &tree, ids[i] );
		live[i] = false;
		treeObserve( &tree, ids, live );
	}
	b2DestroyDynamicTree( &tree );
	ORACLE_CLOSE();
	ORACLE_OPEN( "spatial/TreeCallbackForwarding" );
	tree = b2CreateDynamicTree( 4 );
	for ( int i = 0; i < 3; ++i )
		b2CreateTreeProxy( &tree, (b2AABB){ { .5f + .25f * i, -1 }, { .75f + .25f * i, 1 } }, 1, 100 + i );
	b2CreateTreeProxy( &tree, (b2AABB){ { 3, -1 }, { 3.25f, 1 } }, 1, 200 );
	for ( int policy = 0; policy < 4; ++policy )
	{
		TreeHits h = { .policy = policy };
		b2RayCastInput ray = { { 0, 0 }, { 4, 0 }, 1 };
		stats( b2DynamicTree_CastRay( &tree, &ray, 1, tr, &h ) );
		ORACLE_I32( "count", h.count );
		h = (TreeHits){ .policy = policy };
		b2BoxCastInput cast = { { { -.1f, -.1f }, { .1f, .1f } }, { 4, 0 }, 1 };
		stats( b2DynamicTree_CastBox( &tree, &cast, 1, tb, &h ) );
		ORACLE_I32( "count", h.count );
	}
	b2DestroyDynamicTree( &tree );
	ORACLE_CLOSE();
}

typedef struct Hit
{
	b2ShapeId id;
	b2Pos point;
	b2Vec2 normal;
	float fraction;
	b2PlaneResult plane;
} Hit;
typedef struct Hits
{
	Hit hit[64];
	int count;
	int policy;
	b2WorldId world;
} Hits;
static int cmpHit( const void* a, const void* b )
{
	const Hit* x = a;
	const Hit* y = b;
	return ( x->id.index1 > y->id.index1 ) - ( x->id.index1 < y->id.index1 );
}
static void identity( b2ShapeId id, b2WorldId w )
{
	ORACLE_I32( "shape/index", id.index1 );
	ORACLE_U32( "shape/generation", id.generation );
	ORACLE_BOOL( "shape/owner", id.world0 == w.index1 - 1 );
}
static bool overlap( b2ShapeId id, void* ptr )
{
	Hits* h = ptr;
	h->hit[h->count++].id = id;
	return h->policy != 2;
}
static float cast( b2ShapeId id, b2Pos point, b2Vec2 normal, float fraction, void* ptr )
{
	Hits* h = ptr;
	h->hit[h->count++] = (Hit){ .id = id, .point = point, .normal = normal, .fraction = fraction };
	return h->policy == 0 ? 1 : h->policy == 1 ? -1 : h->policy == 2 ? 0 : fraction;
}
static bool plane( b2ShapeId id, const b2PlaneResult* p, void* ptr )
{
	Hits* h = ptr;
	h->hit[h->count++] = (Hit){ .id = id, .plane = *p };
	return h->policy != 2;
}
static void hits( Hits* h, int payload )
{
	if ( h->policy < 2 )
		qsort( h->hit, h->count, sizeof( Hit ), cmpHit );
	ORACLE_I32( "count", h->count );
	for ( int i = 0; i < h->count; ++i )
	{
		Hit* p = h->hit + i;
		identity( p->id, h->world );
		if ( payload == 1 )
		{
			ORACLE_POS( "point", p->point );
			ORACLE_VEC2( "normal", p->normal );
			ORACLE_F32( "fraction", p->fraction );
		}
		if ( payload == 2 )
		{
			ORACLE_BOOL( "hit", p->plane.hit );
			ORACLE_VEC2( "normal", p->plane.plane.normal );
			ORACLE_F32( "offset", p->plane.plane.offset );
			ORACLE_VEC2( "point", p->plane.point );
		}
	}
}
static void inspect( b2WorldId w )
{
	b2Counters c = b2World_GetCounters( w );
	ORACLE_I32( "bodies", c.bodyCount );
	ORACLE_I32( "shapes", c.shapeCount );
	ORACLE_I32( "joints", c.jointCount );
	ORACLE_I32( "contacts", c.contactCount );
	ORACLE_I32( "islands", c.islandCount );
	ORACLE_I32( "staticHeight", c.staticTreeHeight );
	ORACLE_I32( "treeHeight", c.treeHeight );
	ORACLE_I32( "awakeContacts", c.awakeContactCount );
	ORACLE_I32( "recycled", c.recycledContactCount );
	for ( int i = 0; i < 24; ++i )
		ORACLE_I32( "color", c.colorCounts[i] );
	box( "bounds", b2World_GetBounds( w ) );
}
static void worldQueries( b2Pos base, const char* name )
{
	ORACLE_OPEN( name );
	b2WorldDef wd = b2DefaultWorldDef();
	wd.gravity = b2Vec2_zero;
	wd.workerCount = 1;
	b2WorldId w = b2CreateWorld( &wd );
	inspect( w );
	b2QueryFilter all = { UINT64_MAX, UINT64_MAX };
	b2ShapeProxy proxy = b2MakeProxy( (b2Vec2[]){ { 0, 0 } }, 1, .3f );
	Hits h = { .world = w };
	stats( b2World_OverlapAABB( w, base, (b2AABB){ { -2, -2 }, { 16, 2 } }, all, overlap, &h ) );
	hits( &h, 0 );
	h = (Hits){ .world = w };
	stats( b2World_CastRay( w, base, (b2Vec2){ 16, 0 }, all, cast, &h ) );
	hits( &h, 1 );
	h = (Hits){ .world = w };
	stats( b2World_OverlapShape( w, base, &proxy, all, overlap, &h ) );
	hits( &h, 0 );
	h = (Hits){ .world = w };
	stats( b2World_CastShape( w, base, &proxy, (b2Vec2){ 16, 0 }, all, cast, &h ) );
	hits( &h, 1 );
	b2BodyId bodies[5];
	b2ShapeId shapes[8];
	for ( int i = 0; i < 5; ++i )
	{
		b2BodyDef bd = b2DefaultBodyDef();
		bd.position = b2OffsetPos( base, (b2Vec2){ 3.0f * i, 0 } );
		bd.type = i < 3 ? (b2BodyType)i : b2_staticBody;
		bodies[i] = b2CreateBody( w, &bd );
		b2ShapeDef sd = b2DefaultShapeDef();
		sd.filter.categoryBits = UINT64_C( 1 ) << i;
		sd.filter.maskBits = i % 2 ? 2 : 1;
		if ( i == 0 )
			shapes[i] = b2CreateCircleShape( bodies[i], &sd, &(b2Circle){ { 0, 0 }, .75f } );
		if ( i == 1 )
			shapes[i] = b2CreateCapsuleShape( bodies[i], &sd, &(b2Capsule){ { 0, -.5f }, { 0, .5f }, .5f } );
		if ( i == 2 )
		{
			b2Polygon p = b2MakeBox( .5f, .75f );
			shapes[i] = b2CreatePolygonShape( bodies[i], &sd, &p );
		}
		if ( i == 3 )
			shapes[i] = b2CreateSegmentShape( bodies[i], &sd, &(b2Segment){ { 0, 1 }, { 0, -1 } } );
		if ( i == 4 )
		{
			b2ChainDef cd = b2DefaultChainDef();
			b2Vec2 points[4] = { { -1, -1 }, { 1, -1 }, { 1, 1 }, { -1, 1 } };
			cd.points = points;
			cd.pointCount = 4;
			cd.isLoop = true;
			cd.filter = sd.filter;
			b2ChainId chain = b2CreateChain( bodies[i], &cd );
			ORACLE_I32( "segments", b2Chain_GetSegments( chain, shapes + 4, 4 ) );
		}
	}
	inspect( w );
	b2QueryFilter filters[4] = { all, { 1, 31 }, { 2, 31 }, { 1, 2 } };
	for ( int f = 0; f < 4; ++f )
	{
		for ( int stop = 0; stop < 2; ++stop )
		{
			h = (Hits){ .world = w, .policy = stop ? 2 : 0 };
			stats( b2World_OverlapAABB( w, base, (b2AABB){ { -2, -2 }, { 16, 2 } }, filters[f], overlap, &h ) );
			hits( &h, 0 );
		}
		for ( int i = 0; i < 5; ++i )
		{
			b2Pos origin = b2OffsetPos( base, (b2Vec2){ 3.0f * i, 0 } );
			h = (Hits){ .world = w };
			stats( b2World_OverlapShape( w, origin, &proxy, filters[f], overlap, &h ) );
			hits( &h, 0 );
		}
		for ( int start = 0; start < 3; ++start )
		{
			b2Pos origin = b2OffsetPos( base, (b2Vec2){ start == 0 ? -2 : 0, start == 2 ? 4 : 0 } );
			for ( int policy = 0; policy < 4; ++policy )
			{
				h = (Hits){ .world = w, .policy = policy };
				stats( b2World_CastRay( w, origin, (b2Vec2){ 18, 0 }, filters[f], cast, &h ) );
				hits( &h, 1 );
				h = (Hits){ .world = w, .policy = policy };
				stats( b2World_CastShape( w, origin, &proxy, (b2Vec2){ 18, 0 }, filters[f], cast, &h ) );
				hits( &h, 1 );
			}
			b2RayResult rr = b2World_CastRayClosest( w, origin, (b2Vec2){ 18, 0 }, filters[f] );
			ORACLE_BOOL( "closest/hit", rr.hit );
			stats( (b2TreeStats){ rr.nodeVisits, rr.leafVisits } );
			if ( rr.hit )
			{
				identity( rr.shapeId, w );
				ORACLE_POS( "point", rr.point );
				ORACLE_VEC2( "normal", rr.normal );
				ORACLE_F32( "fraction", rr.fraction );
			}
		}
	}
	for ( int i = 0; i < 8; ++i )
	{
		float x = 3.0f * ( i < 4 ? i : 4 );
		identity( shapes[i], w );
		ORACLE_BOOL( "inside", b2Shape_TestPoint( shapes[i], b2OffsetPos( base, (b2Vec2){ x, 0 } ) ) );
		ORACLE_BOOL( "outside", b2Shape_TestPoint( shapes[i], b2OffsetPos( base, (b2Vec2){ x, 4 } ) ) );
		b2Vec2 origins[4] = { { x - 2, 0 }, { x + 2, 0 }, { x, -2 }, { x, 2 } };
		b2Vec2 translations[4] = { { 4, 0 }, { -4, 0 }, { 0, 4 }, { 0, -4 } };
		for ( int direction = 0; direction < 4; ++direction )
		{
			b2WorldCastOutput out =
				b2Shape_RayCast( shapes[i], b2OffsetPos( base, origins[direction] ), translations[direction] );
			ORACLE_BOOL( "local/hit", out.hit );
			if ( out.hit )
			{
				ORACLE_POS( "point", out.point );
				ORACLE_VEC2( "normal", out.normal );
				ORACLE_F32( "fraction", out.fraction );
				ORACLE_I32( "iterations", out.iterations );
			}
		}
	}
	for ( int i = 0; i < 5; ++i )
	{
		b2Capsule mover = { { 0, -.25f }, { 0, .25f }, .4f };
		b2Pos origin = b2OffsetPos( base, (b2Vec2){ 3.0f * i + .7f, 0 } );
		for ( int stop = 0; stop < 2; ++stop )
		{
			h = (Hits){ .world = w, .policy = stop ? 2 : 0 };
			b2World_CollideMover( w, origin, &mover, all, plane, &h );
			hits( &h, 2 );
		}
		h = (Hits){ .world = w };
		b2World_CollideMover( w, origin, &mover, all, plane, &h );
		qsort( h.hit, h.count, sizeof( Hit ), cmpHit );
		for ( int limit = 0; limit < 3; ++limit )
		{
			b2CollisionPlane planes[64];
			int count = 0;
			for ( int j = 0; j < h.count; ++j )
				if ( h.hit[j].plane.hit )
					planes[count++] = (b2CollisionPlane){ .plane = h.hit[j].plane.plane,
														  .pushLimit = limit == 0	? 0
																	   : limit == 1 ? .1f
																					: FLT_MAX,
														  .clipVelocity = limit != 1 };
			b2PlaneSolverResult result = b2SolvePlanes( (b2Vec2){ -1, .2f }, planes, count );
			ORACLE_VEC2( "delta", result.delta );
			ORACLE_I32( "iterations", result.iterationCount );
			for ( int j = 0; j < count; ++j )
				ORACLE_F32( "push", planes[j].push );
			ORACLE_VEC2( "clip", b2ClipVector( (b2Vec2){ -2, .3f }, planes, count ) );
		}
		ORACLE_F32( "mover/fraction",
					b2World_CastMover( w, b2OffsetPos( base, (b2Vec2){ 3.0f * i - 2, 0 } ), &mover, (b2Vec2){ 4, 0 }, all ) );
	}
	b2Body_SetTransform( bodies[2], b2OffsetPos( base, (b2Vec2){ 6, 2 } ), b2MakeRot( .25f ) );
	inspect( w );
	b2World_Step( w, 1.0f / 60, 4 );
	inspect( w );
	for ( int i = 4; i >= 0; --i )
	{
		b2DestroyBody( bodies[i] );
		inspect( w );
	}
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void moverCases( b2Pos base, const char* name )
{
	ORACLE_OPEN( name );
	for ( int scenario = 0; scenario < 6; ++scenario )
	{
		b2WorldDef wd = b2DefaultWorldDef();
		wd.gravity = b2Vec2_zero;
		wd.workerCount = 1;
		b2WorldId w = b2CreateWorld( &wd );
		b2QueryFilter filter = { 2, UINT64_C( 1 ) << 63 };
		if ( scenario != 0 )
		{
			b2BodyDef bd = b2DefaultBodyDef();
			bd.position = base;
			b2BodyId body = b2CreateBody( w, &bd );
			b2ShapeDef sd = b2DefaultShapeDef();
			sd.filter.categoryBits = UINT64_C( 1 ) << 63;
			sd.filter.maskBits = 2;
			b2Polygon wall = b2MakeOffsetBox( .5f, 3, (b2Vec2){ 1, 0 }, b2Rot_identity );
			if ( scenario == 4 )
				wall = b2MakeOffsetBox( 3, .5f, (b2Vec2){ 0, -1 }, b2MakeRot( .25f ) );
			b2CreatePolygonShape( body, &sd, &wall );
			if ( scenario == 3 )
			{
				b2Polygon floor = b2MakeOffsetBox( 3, .5f, (b2Vec2){ 0, -1 }, b2Rot_identity );
				b2CreatePolygonShape( body, &sd, &floor );
			}
		}
		b2Capsule mover = { { 0, -.25f }, { 0, .25f }, .5f };
		b2Pos origin = b2OffsetPos( base, (b2Vec2){ scenario == 2 ? .3f : scenario == 5 ? -.5f : 0, scenario == 3 ? -.25f : 0 } );
		Hits h = { .world = w };
		b2World_CollideMover( w, origin, &mover, filter, plane, &h );
		hits( &h, 2 );
		for ( int limit = 0; limit < 3; ++limit )
		{
			b2CollisionPlane planes[64];
			int count = 0;
			for ( int i = 0; i < h.count; ++i )
				if ( h.hit[i].plane.hit )
					planes[count++] = (b2CollisionPlane){ .plane = h.hit[i].plane.plane,
														  .pushLimit = limit == 0	? 0
																	   : limit == 1 ? .1f
																					: FLT_MAX,
														  .clipVelocity = limit != 1 };
			b2PlaneSolverResult result = b2SolvePlanes( (b2Vec2){ 1, -1 }, planes, count );
			ORACLE_VEC2( "delta", result.delta );
			ORACLE_I32( "iterations", result.iterationCount );
			for ( int i = 0; i < count; ++i )
				ORACLE_F32( "push", planes[i].push );
			ORACLE_VEC2( "clip", b2ClipVector( (b2Vec2){ 2, -3 }, planes, count ) );
		}
		ORACLE_F32( "cast", b2World_CastMover( w, origin, &mover, (b2Vec2){ 2, -1 }, filter ) );
		ORACLE_F32( "filteredCast",
					b2World_CastMover( w, origin, &mover, (b2Vec2){ 2, -1 }, (b2QueryFilter){ 1, UINT64_C( 1 ) << 63 } ) );
		b2DestroyWorld( w );
	}
	ORACLE_CLOSE();
}
int main( void )
{
	treeHistory();
	worldQueries( b2Pos_zero, "spatial/WorldOrigin" );
	worldQueries( (b2Pos){ 32, -16 }, "spatial/WorldOffset" );
	moverCases( b2Pos_zero, "spatial/MoverOrigin" );
	moverCases( (b2Pos){ 32, -16 }, "spatial/MoverOffset" );
#if defined( BOX2D_DOUBLE_PRECISION )
	worldQueries( (b2Pos){ 10000000, -10000000 }, "spatial/WorldFar" );
	moverCases( (b2Pos){ 10000000, -10000000 }, "spatial/MoverFar" );
#endif
	return 0;
}
