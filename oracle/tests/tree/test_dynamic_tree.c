// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

#include "test_macros.h"

#include "aabb.h"
#include "core.h"
#include "dynamic_tree.h"

#include "box2d/collision.h"
#include "box2d/math_functions.h"

#include <string.h>

static int TreeCreateDestroy( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeCreateDestroy" );
	ORACLE_TREE_BEGIN();
	b2AABB a = {
		.lowerBound = { -1.0f, -1.0f },
		.upperBound = { 2.0f, 2.0f },
	};

	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/22", &tree );
	b2CreateTreeProxyInternal( &tree, a, 1, 0, 0 );
	ORACLE_TREE_TREE( "step/23", &tree );

	ENSURE( tree.nodeEnd == 2 );
	ORACLE_TREE_CHECK( 25, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/25" );
	ORACLE_TREE_TREE( "check-state/25", &tree );
	ENSURE( tree.proxyCount == 1 );
	ORACLE_TREE_CHECK( 26, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/26" );
	ORACLE_TREE_TREE( "check-state/26", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/28", &tree );

	ENSURE( tree.nodeEnd == 0 );
	ORACLE_TREE_CHECK( 30, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/30" );
	ORACLE_TREE_TREE( "check-state/30", &tree );
	ENSURE( tree.proxyCount == 0 );
	ORACLE_TREE_CHECK( 31, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/31" );
	ORACLE_TREE_TREE( "check-state/31", &tree );

	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

float RayCastCallbackFcn( const b2RayCastInput* input, int proxyId, uint64_t userData, void* context )
{
	(void)input;
	(void)userData;

	int* proxyHit = context;
	*proxyHit = proxyId;
	return 0.0f;
}

static int TreeRayCastTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeRayCastTest" );
	ORACLE_TREE_BEGIN();
	// Test AABB centered at origin with bounds [-1, -1] to [1, 1]
	b2AABB a = { .lowerBound = { -1.0f, -1.0f }, .upperBound = { 1.0f, 1.0f }, };
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/50", &tree );
	int proxyId = b2CreateTreeProxyInternal( &tree, a, 1, 0, 0 );
	ORACLE_TREE_TREE( "step/51", &tree );

	b2RayCastInput input = {};
	input.maxFraction = 1.0f;

	// Test 1: Ray hits AABB from left side
	{
		b2Vec2 p1 = { -3.0f, 0.0f };
		b2Vec2 p2 = { 3.0f, 0.0f };

		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/65", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 67, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/67" );
		ORACLE_TREE_TREE( "check-state/67", &tree );
	}

	// Test 2: Ray hits AABB from right side
	{
		b2Vec2 p1 = { 3.0f, 0.0f };
		b2Vec2 p2 = { -3.0f, 0.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/78", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 80, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/80" );
		ORACLE_TREE_TREE( "check-state/80", &tree );
	}

	// Test 3: Ray hits AABB from bottom
	{
		b2Vec2 p1 = { 0.0f, -3.0f };
		b2Vec2 p2 = { 0.0f, 3.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/91", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 93, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/93" );
		ORACLE_TREE_TREE( "check-state/93", &tree );
	}

	// Test 4: Ray hits AABB from top
	{
		b2Vec2 p1 = { 0.0f, 3.0f };
		b2Vec2 p2 = { 0.0f, -3.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/104", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 106, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/106" );
		ORACLE_TREE_TREE( "check-state/106", &tree );
	}

	// Test 5: Ray misses AABB completely (parallel to x-axis)
	{
		b2Vec2 p1 = { -3.0f, 2.0f };
		b2Vec2 p2 = { 3.0f, 2.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/117", &tree );

		ENSURE( proxyHit == -1 );
		ORACLE_TREE_CHECK( 119, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/119" );
		ORACLE_TREE_TREE( "check-state/119", &tree );
	}

	// Test 6: Ray misses AABB completely (parallel to y-axis)
	{
		b2Vec2 p1 = { 2.0f, -3.0f };
		b2Vec2 p2 = { 2.0f, 3.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/130", &tree );

		ENSURE( proxyHit == -1 );
		ORACLE_TREE_CHECK( 132, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/132" );
		ORACLE_TREE_TREE( "check-state/132", &tree );
	}

	// Test 7: Ray starts inside AABB
	{
		b2Vec2 p1 = { 0.0f, 0.0f };
		b2Vec2 p2 = { 2.0f, 0.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/143", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 145, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/145" );
		ORACLE_TREE_TREE( "check-state/145", &tree );
	}

	// Test 8: Ray hits corner of AABB (diagonal ray)
	{
		b2Vec2 p1 = { -2.0f, -2.0f };
		b2Vec2 p2 = { 2.0f, 2.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/156", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 158, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/158" );
		ORACLE_TREE_TREE( "check-state/158", &tree );
	}

	// Test 9: Ray parallel to AABB edge but outside
	{
		b2Vec2 p1 = { -2.0f, 1.5f };
		b2Vec2 p2 = { 2.0f, 1.5f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/169", &tree );

		ENSURE( proxyHit == -1 );
		ORACLE_TREE_CHECK( 171, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/171" );
		ORACLE_TREE_TREE( "check-state/171", &tree );
	}

	// Test 10: Ray parallel to AABB edge and exactly on boundary
	{
		b2Vec2 p1 = { -2.0f, 1.0f };
		b2Vec2 p2 = { 2.0f, 1.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/182", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 184, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/184" );
		ORACLE_TREE_TREE( "check-state/184", &tree );
	}

	// Test 11: Very short ray that doesn't reach AABB
	{
		b2Vec2 p1 = { -3.0f, 0.0f };
		b2Vec2 p2 = { -2.5f, 0.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/195", &tree );

		ENSURE( proxyHit == -1 );
		ORACLE_TREE_CHECK( 197, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/197" );
		ORACLE_TREE_TREE( "check-state/197", &tree );
	}

	// Test 12: Zero-length ray (degenerate case)
	{
		b2Vec2 p1 = { 0.0f, 0.0f };
		b2Vec2 p2 = { 0.0f, 0.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/208", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 210, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/210" );
		ORACLE_TREE_TREE( "check-state/210", &tree );
	}

	// Test 13: Ray hits AABB at exact boundary condition (t = 1.0)
	{
		b2Vec2 p1 = { -2.0f, 0.0f };
		b2Vec2 p2 = { -1.0f, 0.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/221", &tree );

		ENSURE( proxyHit == proxyId );
		ORACLE_TREE_CHECK( 223, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/223" );
		ORACLE_TREE_TREE( "check-state/223", &tree );
	}

	// An off center box catches an axis aligned ray tested against the wrong axis, which the origin
	// centered box above cannot. Two leaves also give the root a child pair.
	b2AABB b = { .lowerBound = { 10.0f, 4.0f }, .upperBound = { 11.0f, 6.0f } };
	int proxyIdB = b2CreateTreeProxyInternal( &tree, b, 1, 1, 0 );
	ORACLE_TREE_TREE( "step/229", &tree );

	// Test 14: Horizontal ray through the off center box
	{
		b2Vec2 p1 = { 0.0f, 5.0f };
		b2Vec2 p2 = { 21.0f, 5.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/239", &tree );

		ENSURE( proxyHit == proxyIdB );
		ORACLE_TREE_CHECK( 241, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/241" );
		ORACLE_TREE_TREE( "check-state/241", &tree );
	}

	// Test 15: Vertical ray through the off center box
	{
		b2Vec2 p1 = { 10.5f, 0.0f };
		b2Vec2 p2 = { 10.5f, 21.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/252", &tree );

		ENSURE( proxyHit == proxyIdB );
		ORACLE_TREE_CHECK( 254, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/254" );
		ORACLE_TREE_TREE( "check-state/254", &tree );
	}

	// Test 16: Horizontal ray passing above the off center box
	{
		b2Vec2 p1 = { 0.0f, 7.0f };
		b2Vec2 p2 = { 21.0f, 7.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/265", &tree );

		ENSURE( proxyHit == -1 );
		ORACLE_TREE_CHECK( 267, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/267" );
		ORACLE_TREE_TREE( "check-state/267", &tree );
	}

	// Test 17: Vertical ray passing beside the off center box
	{
		b2Vec2 p1 = { 12.0f, 0.0f };
		b2Vec2 p2 = { 12.0f, 21.0f };
		input.origin = p1;
		input.translation = b2Sub( p2, p1 );

		int proxyHit = -1;
		b2DynamicTree_CastRay( &tree, &input, 1, RayCastCallbackFcn, &proxyHit );
		ORACLE_TREE_TREE( "step/278", &tree );

		ENSURE( proxyHit == -1 );
		ORACLE_TREE_CHECK( 280, (uint64_t)(proxyHit) );
		ORACLE_TREE_FLUSH( "check/280" );
		ORACLE_TREE_TREE( "check-state/280", &tree );
	}

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/283", &tree );

	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static bool QueryCollectCallback( int proxyId, uint64_t userData, void* context )
{
	(void)userData;
	int* out = context;
	out[proxyId] = 1;
	return true; // continue the query
}

static bool QueryCollectListCallback( int proxyId, uint64_t userData, void* context )
{
	(void)userData;
	int* list = context;
	int count = list[0];
	list[count + 1] = proxyId;
	list[0] = count + 1;
	return true;
}

static int TreeMultipleProxiesTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeMultipleProxiesTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/308", &tree );

	b2AABB a1 = { .lowerBound = { -5.0f, -1.0f }, .upperBound = { -3.0f, 1.0f } };
	b2AABB a2 = { .lowerBound = { -1.0f, -1.0f }, .upperBound = { 1.0f, 1.0f } };
	b2AABB a3 = { .lowerBound = { 3.0f, -1.0f }, .upperBound = { 5.0f, 1.0f } };

	int id1 = b2CreateTreeProxyInternal( &tree, a1, 0x1ull, 42, 0 );
	ORACLE_TREE_TREE( "step/314", &tree );
	int id2 = b2CreateTreeProxyInternal( &tree, a2, 0x2ull, 43, 0 );
	ORACLE_TREE_TREE( "step/315", &tree );
	int id3 = b2CreateTreeProxyInternal( &tree, a3, 0x4ull, 44, 0 );
	ORACLE_TREE_TREE( "step/316", &tree );

	ENSURE( b2DynamicTree_GetProxyCount( &tree ) == 3 );
	ORACLE_TREE_CHECK( 318, (uint64_t)(b2DynamicTree_GetProxyCount( &tree )) );
	ORACLE_TREE_FLUSH( "check/318" );
	ORACLE_TREE_TREE( "check-state/318", &tree );

	ENSURE( b2DynamicTree_GetUserData( &tree, id1 ) == 42 );
	ORACLE_TREE_CHECK( 320, (uint64_t)(b2DynamicTree_GetUserData( &tree, id1 )) );
	ORACLE_TREE_FLUSH( "check/320" );
	ORACLE_TREE_TREE( "check-state/320", &tree );
	ENSURE( b2DynamicTree_GetUserData( &tree, id2 ) == 43 );
	ORACLE_TREE_CHECK( 321, (uint64_t)(b2DynamicTree_GetUserData( &tree, id2 )) );
	ORACLE_TREE_FLUSH( "check/321" );
	ORACLE_TREE_TREE( "check-state/321", &tree );
	ENSURE( b2DynamicTree_GetUserData( &tree, id3 ) == 44 );
	ORACLE_TREE_CHECK( 322, (uint64_t)(b2DynamicTree_GetUserData( &tree, id3 )) );
	ORACLE_TREE_FLUSH( "check/322" );
	ORACLE_TREE_TREE( "check-state/322", &tree );

	ENSURE( b2DynamicTree_GetCategoryBits( &tree, id1 ) == 0x1ull );
	ORACLE_TREE_CHECK( 324, (uint64_t)(b2DynamicTree_GetCategoryBits( &tree, id1 )) );
	ORACLE_TREE_FLUSH( "check/324" );
	ORACLE_TREE_TREE( "check-state/324", &tree );
	ENSURE( b2DynamicTree_GetCategoryBits( &tree, id2 ) == 0x2ull );
	ORACLE_TREE_CHECK( 325, (uint64_t)(b2DynamicTree_GetCategoryBits( &tree, id2 )) );
	ORACLE_TREE_FLUSH( "check/325" );
	ORACLE_TREE_TREE( "check-state/325", &tree );
	ENSURE( b2DynamicTree_GetCategoryBits( &tree, id3 ) == 0x4ull );
	ORACLE_TREE_CHECK( 326, (uint64_t)(b2DynamicTree_GetCategoryBits( &tree, id3 )) );
	ORACLE_TREE_FLUSH( "check/326" );
	ORACLE_TREE_TREE( "check-state/326", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/328", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeQueryTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeQueryTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/334", &tree );

	b2AABB a1 = { .lowerBound = { -5.0f, -1.0f }, .upperBound = { -3.0f, 1.0f } };
	b2AABB a2 = { .lowerBound = { -1.0f, -1.0f }, .upperBound = { 1.0f, 1.0f } };
	b2AABB a3 = { .lowerBound = { 3.0f, -1.0f }, .upperBound = { 5.0f, 1.0f } };

	int id1 = b2CreateTreeProxyInternal( &tree, a1, 0xFFull, 0, 0 );
	ORACLE_TREE_TREE( "step/340", &tree );
	int id2 = b2CreateTreeProxyInternal( &tree, a2, 0xFFull, 0, 0 );
	ORACLE_TREE_TREE( "step/341", &tree );
	int id3 = b2CreateTreeProxyInternal( &tree, a3, 0xFFull, 0, 0 );
	ORACLE_TREE_TREE( "step/342", &tree );

	b2AABB queryA = { .lowerBound = { -2.0f, -2.0f }, .upperBound = { 2.0f, 2.0f } };

	int foundFlags[32] = { 0 };
	b2TreeStats stats = b2DynamicTree_Query( &tree, queryA, 0xFFFFFFFFull, QueryCollectCallback, foundFlags );
	ORACLE_TREE_TREE( "step/347", &tree );

	// We expect at least the middle proxy to be visited.
	ENSURE( foundFlags[id2] == 1 );
	ORACLE_TREE_CHECK( 350, (uint64_t)(foundFlags[id2]) );
	ORACLE_TREE_FLUSH( "check/350" );
	ORACLE_TREE_TREE( "check-state/350", &tree );
	ENSURE( stats.leafVisits >= 1 );
	ORACLE_TREE_CHECK( 351, (uint64_t)(stats.leafVisits) );
	ORACLE_TREE_FLUSH( "check/351" );
	ORACLE_TREE_TREE( "check-state/351", &tree );

	// Test QueryAll using list collector
	int list[16] = { 0 }; // list[0] holds count, following entries are ids
	b2TreeStats allStats = b2DynamicTree_QueryAll( &tree, queryA, QueryCollectListCallback, list );
	ORACLE_TREE_TREE( "step/355", &tree );
	ENSURE( list[0] >= 1 ); // at least one proxy should be collected
	ORACLE_TREE_CHECK( 356, (uint64_t)(list[0]) );
	ORACLE_TREE_FLUSH( "check/356" );
	ORACLE_TREE_TREE( "check-state/356", &tree );
	ENSURE( allStats.leafVisits >= 1 );
	ORACLE_TREE_CHECK( 357, (uint64_t)(allStats.leafVisits) );
	ORACLE_TREE_FLUSH( "check/357" );
	ORACLE_TREE_TREE( "check-state/357", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/359", &tree );
	(void)id1; (void)id3;
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeMoveAndEnlargeTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeMoveAndEnlargeTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/366", &tree );

	b2AABB a = { .lowerBound = { 0.0f, 0.0f }, .upperBound = { 1.0f, 1.0f } };
	int id = b2CreateTreeProxyInternal( &tree, a, 0x1ull, 100, 0 );
	ORACLE_TREE_TREE( "step/369", &tree );

	// Move proxy to a new place
	b2AABB moved = { .lowerBound = { 10.0f, 10.0f }, .upperBound = { 11.0f, 11.0f } };
	b2DynamicTree_MoveProxyInternal( &tree, id, moved, 0 );
	ORACLE_TREE_TREE( "step/373", &tree );

	b2AABB got = b2DynamicTree_GetAABB( &tree, id );
	ENSURE( got.lowerBound.x == moved.lowerBound.x );
	ORACLE_TREE_CHECK( 376, (uint64_t)oracle_bits_f32( got.lowerBound.x ) );
	ORACLE_TREE_FLUSH( "check/376" );
	ORACLE_TREE_TREE( "check-state/376", &tree );
	ENSURE( got.lowerBound.y == moved.lowerBound.y );
	ORACLE_TREE_CHECK( 377, (uint64_t)oracle_bits_f32( got.lowerBound.y ) );
	ORACLE_TREE_FLUSH( "check/377" );
	ORACLE_TREE_TREE( "check-state/377", &tree );
	ENSURE( got.upperBound.x == moved.upperBound.x );
	ORACLE_TREE_CHECK( 378, (uint64_t)oracle_bits_f32( got.upperBound.x ) );
	ORACLE_TREE_FLUSH( "check/378" );
	ORACLE_TREE_TREE( "check-state/378", &tree );
	ENSURE( got.upperBound.y == moved.upperBound.y );
	ORACLE_TREE_CHECK( 379, (uint64_t)oracle_bits_f32( got.upperBound.y ) );
	ORACLE_TREE_FLUSH( "check/379" );
	ORACLE_TREE_TREE( "check-state/379", &tree );

	// Now enlarge the proxy
	b2AABB enlarge = { .lowerBound = { 9.5f, 9.5f }, .upperBound = { 11.5f, 11.5f } };
	b2DynamicTree_EnlargeProxy( &tree, id, enlarge );
	ORACLE_TREE_TREE( "step/383", &tree );

	b2AABB got2 = b2DynamicTree_GetAABB( &tree, id );
	ENSURE( got2.lowerBound.x <= enlarge.lowerBound.x + 1e-6f );
	ORACLE_TREE_CHECK( 386, (uint64_t)oracle_bits_f32( got2.lowerBound.x ) );
	ORACLE_TREE_FLUSH( "check/386" );
	ORACLE_TREE_TREE( "check-state/386", &tree );
	ENSURE( got2.upperBound.x >= enlarge.upperBound.x - 1e-6f );
	ORACLE_TREE_CHECK( 387, (uint64_t)oracle_bits_f32( got2.upperBound.x ) );
	ORACLE_TREE_FLUSH( "check/387" );
	ORACLE_TREE_TREE( "check-state/387", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/389", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeRebuildAndValidateTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeRebuildAndValidateTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/395", &tree );

	// Create a number of proxies to make rebuild meaningful
	for ( int i = 0; i < 12; ++i )
	{
		float x = (float)i * 2.0f;
		b2AABB a = { .lowerBound = { x - 0.5f, -0.5f }, .upperBound = { x + 0.5f, 0.5f } };
		b2CreateTreeProxyInternal( &tree, a, 0xFFull, (uint64_t)i, 0 );
		ORACLE_TREE_TREE( "step/402", &tree );
	}

	int sorted = b2DynamicTree_Rebuild( &tree, true );
	ORACLE_TREE_TREE( "step/405", &tree );
	
	ENSURE( sorted >= 0 );
	ORACLE_TREE_CHECK( 407, (uint64_t)(sorted) );
	ORACLE_TREE_FLUSH( "check/407" );
	ORACLE_TREE_TREE( "check-state/407", &tree );
	ENSURE( b2DynamicTree_GetByteCount( &tree ) > 0 );
	ORACLE_TREE_CHECK( 408, (uint64_t)(b2DynamicTree_GetByteCount( &tree ) > 0) );
	ORACLE_TREE_FLUSH( "check/408" );
	ORACLE_TREE_TREE( "check-state/408", &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) > 0 );
	ORACLE_TREE_CHECK( 409, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/409" );
	ORACLE_TREE_TREE( "check-state/409", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/411", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeRowHeightTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeRowHeightTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/417", &tree );

	int columnCount = 200;
	for (int i = 0; i < columnCount; ++i)
	{
		float x = 1.0f * i;
		b2AABB a = { .lowerBound = { x, 0.0f }, .upperBound = { x + 1.0f, 1.0f } };
		b2CreateTreeProxyInternal( &tree, a, 1, (uint64_t)i, 0 );
		ORACLE_TREE_TREE( "step/424", &tree );
	}

	float minHeight = log2f((float)columnCount);
	ORACLE_F32( "min-height", minHeight );

	b2DynamicTree_Validate( &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) < 2.0f * minHeight );
	ORACLE_TREE_CHECK( 430, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/430" );
	ORACLE_TREE_TREE( "check-state/430", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/432", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeGridHeightTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeGridHeightTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/438", &tree );

	int columnCount = 20;
	int rowCount = 20;
	for (int i = 0; i < columnCount; ++i)
	{
		float x = 1.0f * i;
		for (int j = 0; j < rowCount; ++j)
		{
			float y = 1.0f * j;
			b2AABB a = { .lowerBound = { x, y }, .upperBound = { x + 1.0f, y + 1.0f } };
			b2CreateTreeProxyInternal( &tree, a, 1, (uint64_t)i, 0 );
			ORACLE_TREE_TREE( "step/449", &tree );
		}
	}

	float minHeight = log2f( (float)(rowCount * columnCount) );
	ORACLE_F32( "min-height", minHeight );

	b2DynamicTree_Validate( &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) < 2.0f * minHeight );
	ORACLE_TREE_CHECK( 456, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/456" );
	ORACLE_TREE_TREE( "check-state/456", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/458", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

#define GRID_COUNT 20

static int TreeGridMovementTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeGridMovementTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/466", &tree );

	int proxyIds[GRID_COUNT * GRID_COUNT];
	int index = 0;
	for (int i = 0; i < GRID_COUNT; ++i)
	{
		float x = 1.0f * i;
		for (int j = 0; j < GRID_COUNT; ++j)
		{
			float y = 1.0f * j;
			b2AABB a = { .lowerBound = { x, y }, .upperBound = { x + 1.0f, y + 1.0f } };
			proxyIds[index] = b2CreateTreeProxyInternal( &tree, a, 1, (uint64_t)i, 0 );
			ORACLE_TREE_TREE( "step/477", &tree );
			index += 1;
		}
	}

	ENSURE( index == GRID_COUNT * GRID_COUNT );
	ORACLE_TREE_CHECK( 482, (uint64_t)(index) );
	ORACLE_TREE_FLUSH( "check/482" );
	ORACLE_TREE_TREE( "check-state/482", &tree );

	float minHeight = log2f( (float)( GRID_COUNT * GRID_COUNT ) );
	ORACLE_F32( "min-height", minHeight );

	b2DynamicTree_Validate( &tree );
	int height1 = b2DynamicTree_GetHeight( &tree );
	ENSURE( height1 < 2.0f * minHeight );
	ORACLE_TREE_CHECK( 488, (uint64_t)(height1) );
	ORACLE_TREE_FLUSH( "check/488" );
	ORACLE_TREE_TREE( "check-state/488", &tree );

	b2Vec2 offset = {10.0f, 20.0f};
	index = 0;
	for (int i = 0; i < GRID_COUNT; ++i)
	{
		for (int j = 0; j < GRID_COUNT; ++j)
		{
			b2AABB a = b2DynamicTree_GetAABB( &tree, proxyIds[index] );
			a.lowerBound = b2Add( a.lowerBound, offset );
			a.upperBound = b2Add( a.upperBound, offset );
			b2DynamicTree_MoveProxyInternal( &tree, proxyIds[index], a, 0 );
			ORACLE_TREE_TREE( "step/499", &tree );
			index += 1;
		}
	}

	b2DynamicTree_Validate( &tree );
	int height2 = b2DynamicTree_GetHeight( &tree );
	ENSURE( height2 < 3.0f * minHeight );
	ORACLE_TREE_CHECK( 506, (uint64_t)(height2) );
	ORACLE_TREE_FLUSH( "check/506" );
	ORACLE_TREE_TREE( "check-state/506", &tree );

	b2DynamicTree_Rebuild( &tree, true );
	ORACLE_TREE_TREE( "step/508", &tree );

	b2DynamicTree_Validate( &tree );
	int height3 = b2DynamicTree_GetHeight( &tree );
	ENSURE( height3 < 2.0f * minHeight );
	ORACLE_TREE_CHECK( 512, (uint64_t)(height3) );
	ORACLE_TREE_FLUSH( "check/512" );
	ORACLE_TREE_TREE( "check-state/512", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/514", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

#define TREE_TEST_PROXY_COUNT 200

// Small generator so the random tests repeat run to run
static uint32_t treeRandomState = 12345u;

static float RandomTreeFloat( float lower, float upper )
{
	treeRandomState = 1664525u * treeRandomState + 1013904223u;
	float unit = (float)( treeRandomState >> 8 ) * ( 1.0f / 16777216.0f );
	return lower + ( upper - lower ) * unit;
}

static b2AABB RandomTreeBox( float maxHalfExtent )
{
	float x = RandomTreeFloat( -50.0f, 50.0f );
	float y = RandomTreeFloat( -50.0f, 50.0f );
	float hx = RandomTreeFloat( 0.25f, maxHalfExtent );
	float hy = RandomTreeFloat( 0.25f, maxHalfExtent );
	return (b2AABB){ { x - hx, y - hy }, { x + hx, y + hy } };
}

static bool BoxEquals( b2AABB a, b2AABB b )
{
	return a.lowerBound.x == b.lowerBound.x && a.lowerBound.y == b.lowerBound.y && a.upperBound.x == b.upperBound.x &&
		   a.upperBound.y == b.upperBound.y;
}

static b2AABB EnlargeTreeBox( b2AABB a, float margin )
{
	return (b2AABB){ { a.lowerBound.x - margin, a.lowerBound.y - margin }, { a.upperBound.x + margin, a.upperBound.y + margin } };
}

typedef struct TreeHitSet
{
	bool hit[TREE_TEST_PROXY_COUNT];
	int count;
	bool bad;
} TreeHitSet;

static bool CollectHitCallback( int proxyId, uint64_t userData, void* context )
{
	(void)userData;
	TreeHitSet* hits = context;
	if ( proxyId < 0 || TREE_TEST_PROXY_COUNT <= proxyId || hits->hit[proxyId] )
	{
		// Out of range or reported twice
		hits->bad = true;
		return true;
	}

	hits->hit[proxyId] = true;
	hits->count += 1;
	return true;
}

// Query the tree and compare against every live proxy box
static int CompareQuery( const b2DynamicTree* tree, const int* proxyIds, int proxyCount, b2AABB box )
{
	TreeHitSet hits = { 0 };
	b2DynamicTree_Query( tree, box, B2_DEFAULT_MASK_BITS, CollectHitCallback, &hits );
	ENSURE( hits.bad == false );
	ORACLE_TREE_CHECK( 578, (uint64_t)(hits.bad) );

	int expectedCount = 0;
	for ( int i = 0; i < proxyCount; ++i )
	{
		int proxyId = proxyIds[i];
		bool overlaps = b2AABB_Overlaps( b2DynamicTree_GetAABB( tree, proxyId ), box );
		ENSURE( hits.hit[proxyId] == overlaps );
		ORACLE_TREE_CHECK( 585, (uint64_t)(hits.hit[proxyId]) );
		expectedCount += overlaps ? 1 : 0;
	}

	ENSURE( hits.count == expectedCount );
	ORACLE_TREE_CHECK( 589, (uint64_t)(hits.count) );
	ORACLE_TREE_HITS( &hits );
	ORACLE_TREE_FLUSH( "return/590" );
	return 0;
}

typedef struct TreeCastHit
{
	const b2DynamicTree* tree;
	int proxyId;
	float fraction;
} TreeCastHit;

// Entry fraction along the full translation, negative for a miss
static float RayBoxFraction( b2AABB box, b2Vec2 origin, b2Vec2 translation )
{
	b2CastOutput output = b2AABB_RayCast( box, origin, b2Add( origin, translation ) );
	return output.hit ? output.fraction : -1.0f;
}

// A box sweep against a box is a ray against the box grown by the swept extents
static float BoxCastFraction( b2AABB box, const b2BoxCastInput* input )
{
	b2Vec2 extents = b2AABB_Extents( input->box );
	b2AABB grown = { b2Sub( box.lowerBound, extents ), b2Add( box.upperBound, extents ) };
	return RayBoxFraction( grown, b2AABB_Center( input->box ), input->translation );
}

// The fraction is computed against the full translation rather than the clipped input so it
// matches the brute force bit for bit
static float ClosestRayCallback( const b2RayCastInput* input, int proxyId, uint64_t userData, void* context )
{
	(void)userData;
	TreeCastHit* hit = context;
	float fraction = RayBoxFraction( b2DynamicTree_GetAABB( hit->tree, proxyId ), input->origin, input->translation );
	if ( fraction < 0.0f )
	{
		return -1.0f;
	}

	if ( fraction < hit->fraction )
	{
		hit->proxyId = proxyId;
		hit->fraction = fraction;
	}

	return fraction;
}

static float ClosestBoxCallback( const b2BoxCastInput* input, int proxyId, uint64_t userData, void* context )
{
	(void)userData;
	TreeCastHit* hit = context;
	float fraction = BoxCastFraction( b2DynamicTree_GetAABB( hit->tree, proxyId ), input );
	if ( fraction < 0.0f )
	{
		return -1.0f;
	}

	if ( fraction < hit->fraction )
	{
		hit->proxyId = proxyId;
		hit->fraction = fraction;
	}

	return fraction;
}

// Closest ray hit from the tree against the closest over every live proxy
static int CompareRayCast( const b2DynamicTree* tree, const int* proxyIds, int proxyCount, b2Vec2 origin, b2Vec2 translation )
{
	TreeCastHit hit = { tree, B2_NULL_INDEX, 1.0f };
	b2RayCastInput input = { .origin = origin, .translation = translation, .maxFraction = 1.0f };
	b2DynamicTree_CastRay( tree, &input, B2_DEFAULT_MASK_BITS, ClosestRayCallback, &hit );

	int bestId = B2_NULL_INDEX;
	float bestFraction = 1.0f;
	bool unique = true;
	for ( int i = 0; i < proxyCount; ++i )
	{
		int proxyId = proxyIds[i];
		float fraction = RayBoxFraction( b2DynamicTree_GetAABB( tree, proxyId ), origin, translation );
		if ( fraction < 0.0f )
		{
			continue;
		}

		if ( fraction < bestFraction )
		{
			bestId = proxyId;
			bestFraction = fraction;
			unique = true;
		}
		else if ( fraction == bestFraction )
		{
			unique = false;
		}
	}

	ENSURE( hit.fraction == bestFraction );
	ORACLE_TREE_CHECK( 686, (uint64_t)oracle_bits_f32( hit.fraction ) );
	ORACLE_F32( "fraction/686", hit.fraction );
	if ( unique )
	{
		ENSURE( hit.proxyId == bestId );
		ORACLE_TREE_CHECK( 689, (uint64_t)(hit.proxyId) );
	}

	ORACLE_TREE_FLUSH( "return/692" );
	return 0;
}

static int CompareBoxCast( const b2DynamicTree* tree, const int* proxyIds, int proxyCount, b2AABB box, b2Vec2 translation )
{
	TreeCastHit hit = { tree, B2_NULL_INDEX, 1.0f };
	b2BoxCastInput input = { .box = box, .translation = translation, .maxFraction = 1.0f };
	b2DynamicTree_CastBox( tree, &input, B2_DEFAULT_MASK_BITS, ClosestBoxCallback, &hit );

	int bestId = B2_NULL_INDEX;
	float bestFraction = 1.0f;
	bool unique = true;
	for ( int i = 0; i < proxyCount; ++i )
	{
		int proxyId = proxyIds[i];
		float fraction = BoxCastFraction( b2DynamicTree_GetAABB( tree, proxyId ), &input );
		if ( fraction < 0.0f )
		{
			continue;
		}

		if ( fraction < bestFraction )
		{
			bestId = proxyId;
			bestFraction = fraction;
			unique = true;
		}
		else if ( fraction == bestFraction )
		{
			unique = false;
		}
	}

	ENSURE( hit.fraction == bestFraction );
	ORACLE_TREE_CHECK( 725, (uint64_t)oracle_bits_f32( hit.fraction ) );
	ORACLE_F32( "fraction/725", hit.fraction );
	if ( unique )
	{
		ENSURE( hit.proxyId == bestId );
		ORACLE_TREE_CHECK( 728, (uint64_t)(hit.proxyId) );
	}

	ORACLE_TREE_FLUSH( "return/731" );
	return 0;
}

// A node is marked exactly when a leaf below it is marked. Leaves are checked against the
// set the test marked itself.
static int CheckMarks( const b2DynamicTree* tree, int nodeIndex, const bool* markedProxies, bool* markedBelow )
{
	const b2TreeNode* node = tree->nodes + nodeIndex;
	bool marked = b2IsNodeMoved( node );
	if ( b2IsLeaf( node ) )
	{
		ENSURE( marked == markedProxies[b2GetProxyId( node )] );
		ORACLE_TREE_CHECK( 742, (uint64_t)(marked) );
	}
	else
	{
		int pair = b2GetLeftChild( node );
		bool below1 = false;
		bool below2 = false;
		if ( CheckMarks( tree, pair, markedProxies, &below1 ) != 0 )
		{
			return 1;
		}

		if ( CheckMarks( tree, pair + 1, markedProxies, &below2 ) != 0 )
		{
			return 1;
		}

		ENSURE( marked == ( below1 || below2 ) );
		ORACLE_TREE_CHECK( 759, (uint64_t)(marked) );
	}

	*markedBelow = marked;
	ORACLE_TREE_FLUSH( "return/763" );
	return 0;
}

// Every internal node holds exactly the union of its children
static int CheckBounds( const b2DynamicTree* tree, int nodeIndex, b2AABB* bounds )
{
	const b2TreeNode* node = tree->nodes + nodeIndex;
	if ( b2IsLeaf( node ) == false )
	{
		int pair = b2GetLeftChild( node );
		b2AABB below1, below2;
		if ( CheckBounds( tree, pair, &below1 ) != 0 )
		{
			return 1;
		}

		if ( CheckBounds( tree, pair + 1, &below2 ) != 0 )
		{
			return 1;
		}

		ENSURE( BoxEquals( b2AABB_Union( below1, below2 ), node->aabb ) );
		ORACLE_TREE_CHECK( 784, (uint64_t)(BoxEquals( b2AABB_Union( below1, below2 ), node->aabb )) );
	}

	*bounds = node->aabb;
	ORACLE_TREE_BOX( "bounds", *bounds );
	ORACLE_TREE_FLUSH( "return/788" );
	return 0;
}

static bool RootMarked( const b2DynamicTree* tree )
{
	return b2HasTreeMoved( tree );
}

static int TreeOneProxyTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeOneProxyTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 1 );
	ORACLE_TREE_TREE( "step/798", &tree );
	b2AABB box = { { 0.0f, 0.0f }, { 1.0f, 1.0f } };
	int proxyId = b2CreateTreeProxyInternal( &tree, box, 1, 7, false );
	ORACLE_TREE_TREE( "step/800", &tree );
	int proxyIds[1] = { proxyId };
	b2DynamicTree_Validate( &tree );

	// One proxy is a leaf at the root beside the empty node
	ENSURE( tree.nodeEnd == 2 );
	ORACLE_TREE_CHECK( 805, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/805" );
	ORACLE_TREE_TREE( "check-state/805", &tree );
	ENSURE( tree.proxyCount == 1 );
	ORACLE_TREE_CHECK( 806, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/806" );
	ORACLE_TREE_TREE( "check-state/806", &tree );
	ENSURE( b2IsLeaf( tree.nodes + B2_ROOT_NODE ) );
	ORACLE_TREE_CHECK( 807, (uint64_t)(b2IsLeaf( tree.nodes + B2_ROOT_NODE )) );
	ORACLE_TREE_FLUSH( "check/807" );
	ORACLE_TREE_TREE( "check-state/807", &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) == 0 );
	ORACLE_TREE_CHECK( 808, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/808" );
	ORACLE_TREE_TREE( "check-state/808", &tree );
	ENSURE( b2DynamicTree_GetUserData( &tree, proxyId ) == 7 );
	ORACLE_TREE_CHECK( 809, (uint64_t)(b2DynamicTree_GetUserData( &tree, proxyId )) );
	ORACLE_TREE_FLUSH( "check/809" );
	ORACLE_TREE_TREE( "check-state/809", &tree );
	ENSURE( b2DynamicTree_GetCategoryBits( &tree, proxyId ) == 1 );
	ORACLE_TREE_CHECK( 810, (uint64_t)(b2DynamicTree_GetCategoryBits( &tree, proxyId )) );
	ORACLE_TREE_FLUSH( "check/810" );
	ORACLE_TREE_TREE( "check-state/810", &tree );
	ENSURE( BoxEquals( b2DynamicTree_GetAABB( &tree, proxyId ), box ) );
	ORACLE_TREE_CHECK( 811, (uint64_t)(BoxEquals( b2DynamicTree_GetAABB( &tree, proxyId ), box )) );
	ORACLE_TREE_FLUSH( "check/811" );
	ORACLE_TREE_TREE( "check-state/811", &tree );
	ENSURE( BoxEquals( b2DynamicTree_GetRootBounds( &tree ), box ) );
	ORACLE_TREE_CHECK( 812, (uint64_t)(BoxEquals( b2DynamicTree_GetRootBounds( &tree ), box )) );
	ORACLE_TREE_FLUSH( "check/812" );
	ORACLE_TREE_TREE( "check-state/812", &tree );

	b2AABB inside = { { 0.4f, 0.4f }, { 0.6f, 0.6f } };
	b2AABB outside = { { 5.0f, 5.0f }, { 6.0f, 6.0f } };
	ENSURE( CompareQuery( &tree, proxyIds, 1, inside ) == 0 );
	ORACLE_TREE_CHECK( 816, 0 );
	ORACLE_TREE_FLUSH( "check/816" );
	ORACLE_TREE_TREE( "check-state/816", &tree );
	ENSURE( CompareQuery( &tree, proxyIds, 1, outside ) == 0 );
	ORACLE_TREE_CHECK( 817, 0 );
	ORACLE_TREE_FLUSH( "check/817" );
	ORACLE_TREE_TREE( "check-state/817", &tree );

	b2Vec2 origin = { -1.0f, 0.5f };
	b2Vec2 translation = { 3.0f, 0.0f };
	TreeCastHit hit = { &tree, B2_NULL_INDEX, 1.0f };
	b2RayCastInput input = { .origin = origin, .translation = translation, .maxFraction = 1.0f };
	b2DynamicTree_CastRay( &tree, &input, B2_DEFAULT_MASK_BITS, ClosestRayCallback, &hit );
	ORACLE_TREE_TREE( "step/823", &tree );
	ENSURE( hit.proxyId == proxyId );
	ORACLE_TREE_CHECK( 824, (uint64_t)(hit.proxyId) );
	ORACLE_TREE_FLUSH( "check/824" );
	ORACLE_TREE_TREE( "check-state/824", &tree );
	ENSURE( CompareRayCast( &tree, proxyIds, 1, origin, translation ) == 0 );
	ORACLE_TREE_CHECK( 825, 0 );
	ORACLE_TREE_FLUSH( "check/825" );
	ORACLE_TREE_TREE( "check-state/825", &tree );

	b2Vec2 missOrigin = { -1.0f, 5.0f };
	ENSURE( CompareRayCast( &tree, proxyIds, 1, missOrigin, translation ) == 0 );
	ORACLE_TREE_CHECK( 828, 0 );
	ORACLE_TREE_FLUSH( "check/828" );
	ORACLE_TREE_TREE( "check-state/828", &tree );

	b2AABB castBox = { { -3.0f, 0.25f }, { -2.0f, 0.75f } };
	b2Vec2 castTranslation = { 10.0f, 0.0f };
	ENSURE( CompareBoxCast( &tree, proxyIds, 1, castBox, castTranslation ) == 0 );
	ORACLE_TREE_CHECK( 832, 0 );
	ORACLE_TREE_FLUSH( "check/832" );
	ORACLE_TREE_TREE( "check-state/832", &tree );

	// A full build keeps the leaf at the root
	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == 1 );
	ORACLE_TREE_CHECK( 835, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/835" );
	ORACLE_TREE_TREE( "check-state/835", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.nodeEnd == 2 );
	ORACLE_TREE_CHECK( 837, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/837" );
	ORACLE_TREE_TREE( "check-state/837", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 838, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/838" );
	ORACLE_TREE_TREE( "check-state/838", &tree );
	ENSURE( CompareQuery( &tree, proxyIds, 1, inside ) == 0 );
	ORACLE_TREE_CHECK( 839, 0 );
	ORACLE_TREE_FLUSH( "check/839" );
	ORACLE_TREE_TREE( "check-state/839", &tree );

	// The leaf is the root, so a mark and refit show up in the root bounds directly
	b2AABB enlarged = EnlargeTreeBox( box, 0.5f );
	b2DynamicTree_MarkProxyMoved( &tree, proxyId, enlarged );
	ORACLE_TREE_TREE( "step/843", &tree );
	ENSURE( RootMarked( &tree ) );
	ORACLE_TREE_CHECK( 844, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/844" );
	ORACLE_TREE_TREE( "check-state/844", &tree );
	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/845", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( BoxEquals( b2DynamicTree_GetAABB( &tree, proxyId ), enlarged ) );
	ORACLE_TREE_CHECK( 847, (uint64_t)(BoxEquals( b2DynamicTree_GetAABB( &tree, proxyId ), enlarged )) );
	ORACLE_TREE_FLUSH( "check/847" );
	ORACLE_TREE_TREE( "check-state/847", &tree );
	ENSURE( BoxEquals( b2DynamicTree_GetRootBounds( &tree ), enlarged ) );
	ORACLE_TREE_CHECK( 848, (uint64_t)(BoxEquals( b2DynamicTree_GetRootBounds( &tree ), enlarged )) );
	ORACLE_TREE_FLUSH( "check/848" );
	ORACLE_TREE_TREE( "check-state/848", &tree );
	ENSURE( b2DynamicTree_Rebuild( &tree, false ) == 1 );
	ORACLE_TREE_CHECK( 849, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/849" );
	ORACLE_TREE_TREE( "check-state/849", &tree );
	ENSURE( RootMarked( &tree ) == false );
	ORACLE_TREE_CHECK( 850, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/850" );
	ORACLE_TREE_TREE( "check-state/850", &tree );
	b2DynamicTree_ValidateNoMoved( &tree );

	b2DestroyTreeProxy( &tree, proxyId );
	ORACLE_TREE_TREE( "step/853", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.proxyCount == 0 );
	ORACLE_TREE_CHECK( 855, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/855" );
	ORACLE_TREE_TREE( "check-state/855", &tree );
	ENSURE( tree.nodeEnd == 2 );
	ORACLE_TREE_CHECK( 856, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/856" );
	ORACLE_TREE_TREE( "check-state/856", &tree );
	ENSURE( b2IsEmptyNode( tree.nodes + B2_ROOT_NODE ) );
	ORACLE_TREE_CHECK( 857, (uint64_t)(b2IsEmptyNode( tree.nodes + B2_ROOT_NODE )) );
	ORACLE_TREE_FLUSH( "check/857" );
	ORACLE_TREE_TREE( "check-state/857", &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) == 0 );
	ORACLE_TREE_CHECK( 858, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/858" );
	ORACLE_TREE_TREE( "check-state/858", &tree );
	ENSURE( CompareQuery( &tree, proxyIds, 0, inside ) == 0 );
	ORACLE_TREE_CHECK( 859, 0 );
	ORACLE_TREE_FLUSH( "check/859" );
	ORACLE_TREE_TREE( "check-state/859", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/861", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

// Removing down to one proxy puts the survivor at the root as a leaf, the next insert makes a
// pair for the two of them from the hole, and a full build compacts the holes.
static int TreeLeafRootTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeLeafRootTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/869", &tree );
	b2AABB boxA = { { 0.0f, 0.0f }, { 1.0f, 1.0f } };
	b2AABB boxB = { { 10.0f, 0.0f }, { 11.0f, 1.0f } };
	b2AABB boxC = { { 20.0f, 0.0f }, { 21.0f, 1.0f } };
	b2AABB boxD = { { 30.0f, 0.0f }, { 31.0f, 1.0f } };
	b2AABB all = { { -1.0f, -1.0f }, { 40.0f, 2.0f } };

	int idA = b2CreateTreeProxyInternal( &tree, boxA, 1, 0, false );
	ORACLE_TREE_TREE( "step/876", &tree );
	int idB = b2CreateTreeProxyInternal( &tree, boxB, 1, 1, false );
	ORACLE_TREE_TREE( "step/877", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.nodeEnd == 4 );
	ORACLE_TREE_CHECK( 879, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/879" );
	ORACLE_TREE_TREE( "check-state/879", &tree );
	ENSURE( tree.proxyCount == 2 );
	ORACLE_TREE_CHECK( 880, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/880" );
	ORACLE_TREE_TREE( "check-state/880", &tree );
	ENSURE( b2IsLeaf( tree.nodes + B2_ROOT_NODE ) == false );
	ORACLE_TREE_CHECK( 881, (uint64_t)(b2IsLeaf( tree.nodes + B2_ROOT_NODE )) );
	ORACLE_TREE_FLUSH( "check/881" );
	ORACLE_TREE_TREE( "check-state/881", &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) == 1 );
	ORACLE_TREE_CHECK( 882, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/882" );
	ORACLE_TREE_TREE( "check-state/882", &tree );

	b2DestroyTreeProxy( &tree, idA );
	ORACLE_TREE_TREE( "step/884", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( b2IsLeaf( tree.nodes + B2_ROOT_NODE ) );
	ORACLE_TREE_CHECK( 886, (uint64_t)(b2IsLeaf( tree.nodes + B2_ROOT_NODE )) );
	ORACLE_TREE_FLUSH( "check/886" );
	ORACLE_TREE_TREE( "check-state/886", &tree );
	ENSURE( tree.nodeEnd == 4 );
	ORACLE_TREE_CHECK( 887, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/887" );
	ORACLE_TREE_TREE( "check-state/887", &tree );
	ENSURE( tree.pairFreeList == 2 );
	ORACLE_TREE_CHECK( 888, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/888" );
	ORACLE_TREE_TREE( "check-state/888", &tree );
	ENSURE( tree.proxyCount == 1 );
	ORACLE_TREE_CHECK( 889, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/889" );
	ORACLE_TREE_TREE( "check-state/889", &tree );
	int idsB[1] = { idB };
	ENSURE( CompareQuery( &tree, idsB, 1, all ) == 0 );
	ORACLE_TREE_CHECK( 891, 0 );
	ORACLE_TREE_FLUSH( "check/891" );
	ORACLE_TREE_TREE( "check-state/891", &tree );
	ENSURE( CompareQuery( &tree, idsB, 1, boxA ) == 0 );
	ORACLE_TREE_CHECK( 892, 0 );
	ORACLE_TREE_FLUSH( "check/892" );
	ORACLE_TREE_TREE( "check-state/892", &tree );

	// The freed pair is reused
	int idC = b2CreateTreeProxyInternal( &tree, boxC, 1, 2, false );
	ORACLE_TREE_TREE( "step/895", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.nodeEnd == 4 );
	ORACLE_TREE_CHECK( 897, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/897" );
	ORACLE_TREE_TREE( "check-state/897", &tree );
	ENSURE( tree.pairFreeList == B2_NULL_INDEX );
	ORACLE_TREE_CHECK( 898, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/898" );
	ORACLE_TREE_TREE( "check-state/898", &tree );
	ENSURE( tree.proxyCount == 2 );
	ORACLE_TREE_CHECK( 899, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/899" );
	ORACLE_TREE_TREE( "check-state/899", &tree );
	int idsBC[2] = { idB, idC };
	ENSURE( CompareQuery( &tree, idsBC, 2, all ) == 0 );
	ORACLE_TREE_CHECK( 901, 0 );
	ORACLE_TREE_FLUSH( "check/901" );
	ORACLE_TREE_TREE( "check-state/901", &tree );

	// A third proxy extends the array
	int idD = b2CreateTreeProxyInternal( &tree, boxD, 1, 3, false );
	ORACLE_TREE_TREE( "step/904", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.nodeEnd == 6 );
	ORACLE_TREE_CHECK( 906, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/906" );
	ORACLE_TREE_TREE( "check-state/906", &tree );
	ENSURE( tree.proxyCount == 3 );
	ORACLE_TREE_CHECK( 907, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/907" );
	ORACLE_TREE_TREE( "check-state/907", &tree );
	int idsBCD[3] = { idB, idC, idD };
	ENSURE( CompareQuery( &tree, idsBCD, 3, all ) == 0 );
	ORACLE_TREE_CHECK( 909, 0 );
	ORACLE_TREE_FLUSH( "check/909" );
	ORACLE_TREE_TREE( "check-state/909", &tree );
	ENSURE( CompareQuery( &tree, idsBCD, 3, boxC ) == 0 );
	ORACLE_TREE_CHECK( 910, 0 );
	ORACLE_TREE_FLUSH( "check/910" );
	ORACLE_TREE_TREE( "check-state/910", &tree );

	// A remove leaves a hole that the full build compacts
	b2DestroyTreeProxy( &tree, idB );
	ORACLE_TREE_TREE( "step/913", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.nodeEnd == 6 );
	ORACLE_TREE_CHECK( 915, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/915" );
	ORACLE_TREE_TREE( "check-state/915", &tree );
	ENSURE( tree.pairFreeList != B2_NULL_INDEX );
	ORACLE_TREE_CHECK( 916, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/916" );
	ORACLE_TREE_TREE( "check-state/916", &tree );
	int idsCD[2] = { idC, idD };
	ENSURE( CompareQuery( &tree, idsCD, 2, all ) == 0 );
	ORACLE_TREE_CHECK( 918, 0 );
	ORACLE_TREE_FLUSH( "check/918" );
	ORACLE_TREE_TREE( "check-state/918", &tree );
	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == 2 );
	ORACLE_TREE_CHECK( 919, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/919" );
	ORACLE_TREE_TREE( "check-state/919", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.nodeEnd == 4 );
	ORACLE_TREE_CHECK( 921, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/921" );
	ORACLE_TREE_TREE( "check-state/921", &tree );
	ENSURE( tree.pairFreeList == B2_NULL_INDEX );
	ORACLE_TREE_CHECK( 922, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/922" );
	ORACLE_TREE_TREE( "check-state/922", &tree );
	ENSURE( CompareQuery( &tree, idsCD, 2, all ) == 0 );
	ORACLE_TREE_CHECK( 923, 0 );
	ORACLE_TREE_FLUSH( "check/923" );
	ORACLE_TREE_TREE( "check-state/923", &tree );

	b2DestroyTreeProxy( &tree, idC );
	ORACLE_TREE_TREE( "step/925", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( b2IsLeaf( tree.nodes + B2_ROOT_NODE ) );
	ORACLE_TREE_CHECK( 927, (uint64_t)(b2IsLeaf( tree.nodes + B2_ROOT_NODE )) );
	ORACLE_TREE_FLUSH( "check/927" );
	ORACLE_TREE_TREE( "check-state/927", &tree );

	b2DestroyTreeProxy( &tree, idD );
	ORACLE_TREE_TREE( "step/929", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( tree.proxyCount == 0 );
	ORACLE_TREE_CHECK( 931, (uint64_t)(tree.proxyCount) );
	ORACLE_TREE_FLUSH( "check/931" );
	ORACLE_TREE_TREE( "check-state/931", &tree );
	ENSURE( b2IsEmptyNode( tree.nodes + B2_ROOT_NODE ) );
	ORACLE_TREE_CHECK( 932, (uint64_t)(b2IsEmptyNode( tree.nodes + B2_ROOT_NODE )) );
	ORACLE_TREE_FLUSH( "check/932" );
	ORACLE_TREE_TREE( "check-state/932", &tree );
	ENSURE( b2DynamicTree_GetHeight( &tree ) == 0 );
	ORACLE_TREE_CHECK( 933, (uint64_t)(b2DynamicTree_GetHeight( &tree )) );
	ORACLE_TREE_FLUSH( "check/933" );
	ORACLE_TREE_TREE( "check-state/933", &tree );
	ENSURE( CompareQuery( &tree, idsBCD, 0, all ) == 0 );
	ORACLE_TREE_CHECK( 934, 0 );
	ORACLE_TREE_FLUSH( "check/934" );
	ORACLE_TREE_TREE( "check-state/934", &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/936", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeQueryBruteForceTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeQueryBruteForceTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_TREE( "step/942", &tree );
	int proxyIds[TREE_TEST_PROXY_COUNT];
	for ( int i = 0; i < TREE_TEST_PROXY_COUNT; ++i )
	{
		proxyIds[i] = b2CreateTreeProxyInternal( &tree, RandomTreeBox( 2.0f ), 1, (uint64_t)i, false );
		ORACLE_TREE_TREE( "step/946", &tree );
	}

	b2DynamicTree_Validate( &tree );
	for ( int i = 0; i < 32; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, TREE_TEST_PROXY_COUNT, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 952, 0 );
		ORACLE_TREE_FLUSH( "check/952" );
		ORACLE_TREE_TREE( "check-state/952", &tree );
	}

	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_CHECK( 955, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/955" );
	ORACLE_TREE_TREE( "check-state/955", &tree );
	b2DynamicTree_Validate( &tree );
	for ( int i = 0; i < 32; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, TREE_TEST_PROXY_COUNT, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 959, 0 );
		ORACLE_TREE_FLUSH( "check/959" );
		ORACLE_TREE_TREE( "check-state/959", &tree );
	}

	// Moves leave the built order behind
	for ( int i = 0; i < TREE_TEST_PROXY_COUNT; i += 2 )
	{
		b2DynamicTree_MoveProxyInternal( &tree, proxyIds[i], RandomTreeBox( 2.0f ), false );
		ORACLE_TREE_TREE( "step/965", &tree );
	}

	b2DynamicTree_Validate( &tree );
	for ( int i = 0; i < 32; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, TREE_TEST_PROXY_COUNT, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 971, 0 );
		ORACLE_TREE_FLUSH( "check/971" );
		ORACLE_TREE_TREE( "check-state/971", &tree );
	}

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/974", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

static int TreeCastBruteForceTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeCastBruteForceTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_TREE( "step/980", &tree );
	int proxyIds[TREE_TEST_PROXY_COUNT];
	for ( int i = 0; i < TREE_TEST_PROXY_COUNT; ++i )
	{
		proxyIds[i] = b2CreateTreeProxyInternal( &tree, RandomTreeBox( 2.0f ), 1, (uint64_t)i, false );
		ORACLE_TREE_TREE( "step/984", &tree );
	}

	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_CHECK( 987, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/987" );
	ORACLE_TREE_TREE( "check-state/987", &tree );
	b2DynamicTree_Validate( &tree );

	for ( int i = 0; i < 64; ++i )
	{
		b2Vec2 origin = { RandomTreeFloat( -60.0f, 60.0f ), RandomTreeFloat( -60.0f, 60.0f ) };
		float angle = RandomTreeFloat( -B2_PI, B2_PI );
		float length = RandomTreeFloat( 40.0f, 120.0f );
		ORACLE_F32( "angle", angle );
		ORACLE_F32( "length", length );
		b2Vec2 translation = { length * cosf( angle ), length * sinf( angle ) };
		ORACLE_VEC2( "translation", translation );
		ENSURE( CompareRayCast( &tree, proxyIds, TREE_TEST_PROXY_COUNT, origin, translation ) == 0 );
		ORACLE_TREE_CHECK( 996, 0 );
		ORACLE_TREE_FLUSH( "check/996" );
		ORACLE_TREE_TREE( "check-state/996", &tree );

		b2Vec2 halfExtent = { RandomTreeFloat( 0.25f, 2.0f ), RandomTreeFloat( 0.25f, 2.0f ) };
		b2AABB box = { b2Sub( origin, halfExtent ), b2Add( origin, halfExtent ) };
		ENSURE( CompareBoxCast( &tree, proxyIds, TREE_TEST_PROXY_COUNT, box, translation ) == 0 );
		ORACLE_TREE_CHECK( 1000, 0 );
		ORACLE_TREE_FLUSH( "check/1000" );
		ORACLE_TREE_TREE( "check-state/1000", &tree );
	}

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/1003", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

// Marks reach the root from any leaf, the sweep refit writes exact unions, and the rebuild clears
static int TreeMarkAndRefitTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeMarkAndRefitTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_TREE( "step/1010", &tree );
	int proxyIds[TREE_TEST_PROXY_COUNT];
	for ( int i = 0; i < TREE_TEST_PROXY_COUNT; ++i )
	{
		proxyIds[i] = b2CreateTreeProxyInternal( &tree, RandomTreeBox( 2.0f ), 1, (uint64_t)i, false );
		ORACLE_TREE_TREE( "step/1014", &tree );
	}

	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_CHECK( 1017, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1017" );
	ORACLE_TREE_TREE( "check-state/1017", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1018, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1018" );
	ORACLE_TREE_TREE( "check-state/1018", &tree );
	ENSURE( tree.nodeEnd == 2 * TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_CHECK( 1019, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/1019" );
	ORACLE_TREE_TREE( "check-state/1019", &tree );
	ENSURE( RootMarked( &tree ) == false );
	ORACLE_TREE_CHECK( 1020, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/1020" );
	ORACLE_TREE_TREE( "check-state/1020", &tree );

	bool marked[TREE_TEST_PROXY_COUNT] = { 0 };
	b2AABB enlarged[TREE_TEST_PROXY_COUNT] = { 0 };
	for ( int i = 0; i < TREE_TEST_PROXY_COUNT; i += 3 )
	{
		int proxyId = proxyIds[i];
		enlarged[proxyId] = EnlargeTreeBox( b2DynamicTree_GetAABB( &tree, proxyId ), 0.5f );
		b2DynamicTree_MarkProxyMoved( &tree, proxyId, enlarged[proxyId] );
		ORACLE_TREE_TREE( "step/1028", &tree );
		marked[proxyId] = true;
	}

	bool markedBelow = false;
	ENSURE( CheckMarks( &tree, B2_ROOT_NODE, marked, &markedBelow ) == 0 );
	ORACLE_TREE_CHECK( 1033, 0 );
	ORACLE_TREE_FLUSH( "check/1033" );
	ORACLE_TREE_TREE( "check-state/1033", &tree );
	ENSURE( markedBelow );
	ORACLE_TREE_CHECK( 1034, (uint64_t)(markedBelow) );
	ORACLE_TREE_FLUSH( "check/1034" );
	ORACLE_TREE_TREE( "check-state/1034", &tree );
	ENSURE( RootMarked( &tree ) );
	ORACLE_TREE_CHECK( 1035, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/1035" );
	ORACLE_TREE_TREE( "check-state/1035", &tree );

	// The leaf records took the fat boxes before the refit
	for ( int i = 0; i < TREE_TEST_PROXY_COUNT; ++i )
	{
		if ( marked[i] )
		{
			ENSURE( BoxEquals( b2DynamicTree_GetAABB( &tree, i ), enlarged[i] ) );
			ORACLE_TREE_CHECK( 1042, (uint64_t)(BoxEquals( b2DynamicTree_GetAABB( &tree, i ), enlarged[i] )) );
			ORACLE_TREE_FLUSH( "check/1042" );
			ORACLE_TREE_TREE( "check-state/1042", &tree );
		}
	}

	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/1046", &tree );
	b2DynamicTree_Validate( &tree );
	b2AABB bounds;
	ENSURE( CheckBounds( &tree, B2_ROOT_NODE, &bounds ) == 0 );
	ORACLE_TREE_CHECK( 1049, 0 );
	ORACLE_TREE_FLUSH( "check/1049" );
	ORACLE_TREE_TREE( "check-state/1049", &tree );
	ENSURE( BoxEquals( bounds, b2DynamicTree_GetRootBounds( &tree ) ) );
	ORACLE_TREE_CHECK( 1050, (uint64_t)(BoxEquals( bounds, b2DynamicTree_GetRootBounds( &tree ) )) );
	ORACLE_TREE_FLUSH( "check/1050" );
	ORACLE_TREE_TREE( "check-state/1050", &tree );

	// Marks survive the refit
	ENSURE( CheckMarks( &tree, B2_ROOT_NODE, marked, &markedBelow ) == 0 );
	ORACLE_TREE_CHECK( 1053, 0 );
	ORACLE_TREE_FLUSH( "check/1053" );
	ORACLE_TREE_TREE( "check-state/1053", &tree );
	ENSURE( markedBelow );
	ORACLE_TREE_CHECK( 1054, (uint64_t)(markedBelow) );
	ORACLE_TREE_FLUSH( "check/1054" );
	ORACLE_TREE_TREE( "check-state/1054", &tree );
	for ( int i = 0; i < 16; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, TREE_TEST_PROXY_COUNT, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 1057, 0 );
		ORACLE_TREE_FLUSH( "check/1057" );
		ORACLE_TREE_TREE( "check-state/1057", &tree );
	}

	// The rebuild clears every mark and restores the order
	ENSURE( b2DynamicTree_Rebuild( &tree, false ) > 0 );
	ORACLE_TREE_CHECK( 1061, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1061" );
	ORACLE_TREE_TREE( "check-state/1061", &tree );
	ENSURE( RootMarked( &tree ) == false );
	ORACLE_TREE_CHECK( 1062, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/1062" );
	ORACLE_TREE_TREE( "check-state/1062", &tree );
	memset( marked, 0, sizeof( marked ) );
	ENSURE( CheckMarks( &tree, B2_ROOT_NODE, marked, &markedBelow ) == 0 );
	ORACLE_TREE_CHECK( 1064, 0 );
	ORACLE_TREE_FLUSH( "check/1064" );
	ORACLE_TREE_TREE( "check-state/1064", &tree );
	ENSURE( markedBelow == false );
	ORACLE_TREE_CHECK( 1065, (uint64_t)(markedBelow) );
	ORACLE_TREE_FLUSH( "check/1065" );
	ORACLE_TREE_TREE( "check-state/1065", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1066, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1066" );
	ORACLE_TREE_TREE( "check-state/1066", &tree );
	ENSURE( CheckBounds( &tree, B2_ROOT_NODE, &bounds ) == 0 );
	ORACLE_TREE_CHECK( 1067, 0 );
	ORACLE_TREE_FLUSH( "check/1067" );
	ORACLE_TREE_TREE( "check-state/1067", &tree );
	for ( int i = 0; i < 16; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, TREE_TEST_PROXY_COUNT, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 1070, 0 );
		ORACLE_TREE_FLUSH( "check/1070" );
		ORACLE_TREE_TREE( "check-state/1070", &tree );
	}

	// A single deep leaf still marks the whole path to the root
	int loneId = proxyIds[TREE_TEST_PROXY_COUNT / 2];
	enlarged[loneId] = EnlargeTreeBox( b2DynamicTree_GetAABB( &tree, loneId ), 0.5f );
	b2DynamicTree_MarkProxyMoved( &tree, loneId, enlarged[loneId] );
	ORACLE_TREE_TREE( "step/1076", &tree );
	marked[loneId] = true;
	ENSURE( CheckMarks( &tree, B2_ROOT_NODE, marked, &markedBelow ) == 0 );
	ORACLE_TREE_CHECK( 1078, 0 );
	ORACLE_TREE_FLUSH( "check/1078" );
	ORACLE_TREE_TREE( "check-state/1078", &tree );
	ENSURE( markedBelow );
	ORACLE_TREE_CHECK( 1079, (uint64_t)(markedBelow) );
	ORACLE_TREE_FLUSH( "check/1079" );
	ORACLE_TREE_TREE( "check-state/1079", &tree );
	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/1080", &tree );
	ENSURE( CheckBounds( &tree, B2_ROOT_NODE, &bounds ) == 0 );
	ORACLE_TREE_CHECK( 1081, 0 );
	ORACLE_TREE_FLUSH( "check/1081" );
	ORACLE_TREE_TREE( "check-state/1081", &tree );
	ENSURE( b2DynamicTree_Rebuild( &tree, false ) > 0 );
	ORACLE_TREE_CHECK( 1082, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1082" );
	ORACLE_TREE_TREE( "check-state/1082", &tree );
	b2DynamicTree_ValidateNoMoved( &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/1085", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

// The rebuild orders the array, a remove keeps the order, an insert that moves an internal node
// below its children loses it, and the rebuild runs on an unordered array even when nothing moved
static int TreeDfsOrderTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeDfsOrderTest" );
	ORACLE_TREE_BEGIN();
	int proxyCount = 50;
	b2DynamicTree tree = b2CreateDynamicTree( TREE_TEST_PROXY_COUNT );
	ORACLE_TREE_TREE( "step/1094", &tree );
	int proxyIds[TREE_TEST_PROXY_COUNT];
	for ( int i = 0; i < proxyCount; ++i )
	{
		proxyIds[i] = b2CreateTreeProxyInternal( &tree, RandomTreeBox( 2.0f ), 1, (uint64_t)i, false );
		ORACLE_TREE_TREE( "step/1098", &tree );
	}

	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == proxyCount );
	ORACLE_TREE_CHECK( 1101, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1101" );
	ORACLE_TREE_TREE( "check-state/1101", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1102, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1102" );
	ORACLE_TREE_TREE( "check-state/1102", &tree );
	ENSURE( tree.nodeEnd == 2 * proxyCount );
	ORACLE_TREE_CHECK( 1103, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/1103" );
	ORACLE_TREE_TREE( "check-state/1103", &tree );
	ENSURE( tree.pairFreeList == B2_NULL_INDEX );
	ORACLE_TREE_CHECK( 1104, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/1104" );
	ORACLE_TREE_TREE( "check-state/1104", &tree );

	// A fresh pair lands above everything, so the order survives when the sibling is a leaf.
	// Validate checks the flag against the array either way.
	proxyIds[proxyCount] = b2CreateTreeProxyInternal( &tree, RandomTreeBox( 2.0f ), 1, (uint64_t)proxyCount, false );
	ORACLE_TREE_TREE( "step/1108", &tree );
	proxyCount += 1;
	b2DynamicTree_Validate( &tree );

	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == proxyCount );
	ORACLE_TREE_CHECK( 1112, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1112" );
	ORACLE_TREE_TREE( "check-state/1112", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1113, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1113" );
	ORACLE_TREE_TREE( "check-state/1113", &tree );

	b2DynamicTree_MoveProxyInternal( &tree, proxyIds[7], RandomTreeBox( 2.0f ), false );
	ORACLE_TREE_TREE( "step/1115", &tree );
	b2DynamicTree_Validate( &tree );

	ENSURE( b2DynamicTree_Rebuild( &tree, true ) == proxyCount );
	ORACLE_TREE_CHECK( 1118, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1118" );
	ORACLE_TREE_TREE( "check-state/1118", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1119, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1119" );
	ORACLE_TREE_TREE( "check-state/1119", &tree );
	int orderedEnd = tree.nodeEnd;

	// Removes leave holes but every survivor stays above its parent
	b2DestroyTreeProxy( &tree, proxyIds[3] );
	ORACLE_TREE_TREE( "step/1123", &tree );
	b2DestroyTreeProxy( &tree, proxyIds[20] );
	ORACLE_TREE_TREE( "step/1124", &tree );
	proxyIds[3] = proxyIds[proxyCount - 1];
	proxyIds[20] = proxyIds[proxyCount - 2];
	proxyCount -= 2;
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1128, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1128" );
	ORACLE_TREE_TREE( "check-state/1128", &tree );
	ENSURE( tree.nodeEnd == orderedEnd );
	ORACLE_TREE_CHECK( 1129, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/1129" );
	ORACLE_TREE_TREE( "check-state/1129", &tree );
	ENSURE( tree.pairFreeList != B2_NULL_INDEX );
	ORACLE_TREE_CHECK( 1130, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/1130" );
	ORACLE_TREE_TREE( "check-state/1130", &tree );
	b2DynamicTree_Validate( &tree );

	b2AABB enlarged = EnlargeTreeBox( b2DynamicTree_GetAABB( &tree, proxyIds[10] ), 0.5f );
	b2DynamicTree_MarkProxyMoved( &tree, proxyIds[10], enlarged );
	ORACLE_TREE_TREE( "step/1134", &tree );
	ENSURE( RootMarked( &tree ) );
	ORACLE_TREE_CHECK( 1135, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/1135" );
	ORACLE_TREE_TREE( "check-state/1135", &tree );
	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/1136", &tree );
	b2DynamicTree_Validate( &tree );
	b2AABB bounds;
	ENSURE( CheckBounds( &tree, B2_ROOT_NODE, &bounds ) == 0 );
	ORACLE_TREE_CHECK( 1139, 0 );
	ORACLE_TREE_FLUSH( "check/1139" );
	ORACLE_TREE_TREE( "check-state/1139", &tree );
	ENSURE( BoxEquals( bounds, b2DynamicTree_GetRootBounds( &tree ) ) );
	ORACLE_TREE_CHECK( 1140, (uint64_t)(BoxEquals( bounds, b2DynamicTree_GetRootBounds( &tree ) )) );
	ORACLE_TREE_FLUSH( "check/1140" );
	ORACLE_TREE_TREE( "check-state/1140", &tree );
	for ( int i = 0; i < 16; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, proxyCount, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 1143, 0 );
		ORACLE_TREE_FLUSH( "check/1143" );
		ORACLE_TREE_TREE( "check-state/1143", &tree );
	}

	ENSURE( b2DynamicTree_Rebuild( &tree, false ) > 0 );
	ORACLE_TREE_CHECK( 1146, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1146" );
	ORACLE_TREE_TREE( "check-state/1146", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1147, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1147" );
	ORACLE_TREE_TREE( "check-state/1147", &tree );
	ENSURE( tree.nodeEnd == 2 * proxyCount );
	ORACLE_TREE_CHECK( 1148, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/1148" );
	ORACLE_TREE_TREE( "check-state/1148", &tree );
	ENSURE( tree.pairFreeList == B2_NULL_INDEX );
	ORACLE_TREE_CHECK( 1149, (uint64_t)(tree.pairFreeList) );
	ORACLE_TREE_FLUSH( "check/1149" );
	ORACLE_TREE_TREE( "check-state/1149", &tree );
	b2DynamicTree_Validate( &tree );

	// A box around everything makes the root the best sibling, so the root's subtree moves into
	// the new pair below its own children and the order is lost with nothing marked
	b2AABB huge = { { -100.0f, -100.0f }, { 100.0f, 100.0f } };
	proxyIds[proxyCount] = b2CreateTreeProxyInternal( &tree, huge, 1, (uint64_t)proxyCount, false );
	ORACLE_TREE_TREE( "step/1155", &tree );
	proxyCount += 1;
	ENSURE( tree.dfsOrdered == false );
	ORACLE_TREE_CHECK( 1157, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1157" );
	ORACLE_TREE_TREE( "check-state/1157", &tree );
	ENSURE( RootMarked( &tree ) == false );
	ORACLE_TREE_CHECK( 1158, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/1158" );
	ORACLE_TREE_TREE( "check-state/1158", &tree );
	b2DynamicTree_Validate( &tree );

	// The refit falls back to a descent on an unordered array
	enlarged = EnlargeTreeBox( b2DynamicTree_GetAABB( &tree, proxyIds[10] ), 0.5f );
	b2DynamicTree_MarkProxyMoved( &tree, proxyIds[10], enlarged );
	ORACLE_TREE_TREE( "step/1163", &tree );
	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/1164", &tree );
	b2DynamicTree_Validate( &tree );
	ENSURE( CheckBounds( &tree, B2_ROOT_NODE, &bounds ) == 0 );
	ORACLE_TREE_CHECK( 1166, 0 );
	ORACLE_TREE_FLUSH( "check/1166" );
	ORACLE_TREE_TREE( "check-state/1166", &tree );
	b2DynamicTree_ClearMoved( &tree );
	ORACLE_TREE_TREE( "step/1167", &tree );
	ENSURE( RootMarked( &tree ) == false );
	ORACLE_TREE_CHECK( 1168, (uint64_t)(RootMarked( &tree )) );
	ORACLE_TREE_FLUSH( "check/1168" );
	ORACLE_TREE_TREE( "check-state/1168", &tree );

	// The rebuild runs for the order alone
	ENSURE( b2DynamicTree_Rebuild( &tree, false ) > 0 );
	ORACLE_TREE_CHECK( 1171, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1171" );
	ORACLE_TREE_TREE( "check-state/1171", &tree );
	ENSURE( tree.dfsOrdered );
	ORACLE_TREE_CHECK( 1172, (uint64_t)(tree.dfsOrdered) );
	ORACLE_TREE_FLUSH( "check/1172" );
	ORACLE_TREE_TREE( "check-state/1172", &tree );
	ENSURE( tree.nodeEnd == 2 * proxyCount );
	ORACLE_TREE_CHECK( 1173, (uint64_t)(tree.nodeEnd) );
	ORACLE_TREE_FLUSH( "check/1173" );
	ORACLE_TREE_TREE( "check-state/1173", &tree );
	b2DynamicTree_Validate( &tree );
	for ( int i = 0; i < 16; ++i )
	{
		ENSURE( CompareQuery( &tree, proxyIds, proxyCount, RandomTreeBox( 8.0f ) ) == 0 );
		ORACLE_TREE_CHECK( 1177, 0 );
		ORACLE_TREE_FLUSH( "check/1177" );
		ORACLE_TREE_TREE( "check-state/1177", &tree );
	}

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/1180", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

// A create and destroy between steps leaves no marks. The rebuild has to run anyway when the
// array lost its order, or the next enlarge and refit leave the ancestors stale.
static int TreeStaleAfterInsertRemoveTest( void )
{
	ORACLE_OPEN( "dynamic_tree/TreeStaleAfterInsertRemoveTest" );
	ORACLE_TREE_BEGIN();
	b2DynamicTree tree = b2CreateDynamicTree( 16 );
	ORACLE_TREE_TREE( "step/1188", &tree );
	int ids[16];
	for ( int i = 0; i < 16; ++i )
	{
		b2AABB box = { { 2.0f * i, 0.0f }, { 2.0f * i + 1.0f, 1.0f } };
		ids[i] = b2CreateTreeProxyInternal( &tree, box, 1, (uint64_t)i, true );
		ORACLE_TREE_TREE( "step/1193", &tree );
	}

	ENSURE( b2DynamicTree_Rebuild( &tree, false ) == 16 );
	ORACLE_TREE_CHECK( 1196, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1196" );
	ORACLE_TREE_TREE( "check-state/1196", &tree );

	// Control: enlarge and refit on the ordered tree
	b2AABB big = { { 0.0f, 0.0f }, { 40.0f, 40.0f } };
	b2DynamicTree_MarkProxyMoved( &tree, ids[0], big );
	ORACLE_TREE_TREE( "step/1200", &tree );
	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/1201", &tree );
	ENSURE( b2AABB_Contains( b2DynamicTree_GetRootBounds( &tree ), big ) );
	ORACLE_TREE_CHECK( 1202, (uint64_t)(b2AABB_Contains( b2DynamicTree_GetRootBounds( &tree ), big )) );
	ORACLE_TREE_FLUSH( "check/1202" );
	ORACLE_TREE_TREE( "check-state/1202", &tree );
	ENSURE( b2DynamicTree_Rebuild( &tree, false ) > 0 );
	ORACLE_TREE_CHECK( 1203, (uint64_t)treeLastRebuild );
	ORACLE_TREE_FLUSH( "check/1203" );
	ORACLE_TREE_TREE( "check-state/1203", &tree );

	// Insert and remove between steps leaves no marks
	b2AABB extra = { { 100.0f, 100.0f }, { 101.0f, 101.0f } };
	int extraId = b2CreateTreeProxyInternal( &tree, extra, 1, 99, true );
	ORACLE_TREE_TREE( "step/1207", &tree );
	b2DestroyTreeProxy( &tree, extraId );
	ORACLE_TREE_TREE( "step/1208", &tree );
	ENSURE( b2HasTreeMoved( &tree ) == false );
	ORACLE_TREE_CHECK( 1209, (uint64_t)(b2HasTreeMoved( &tree )) );
	ORACLE_TREE_FLUSH( "check/1209" );
	ORACLE_TREE_TREE( "check-state/1209", &tree );

	// The world's rebuild call, then a proxy enlarges during the step
	b2DynamicTree_Rebuild( &tree, false );
	ORACLE_TREE_TREE( "step/1212", &tree );
	b2AABB bigger = { { -10.0f, -10.0f }, { 60.0f, 60.0f } };
	b2DynamicTree_MarkProxyMoved( &tree, ids[0], bigger );
	ORACLE_TREE_TREE( "step/1214", &tree );
	b2DynamicTree_Refit( &tree );
	ORACLE_TREE_TREE( "step/1215", &tree );
	ENSURE( b2AABB_Contains( b2DynamicTree_GetRootBounds( &tree ), bigger ) );
	ORACLE_TREE_CHECK( 1216, (uint64_t)(b2AABB_Contains( b2DynamicTree_GetRootBounds( &tree ), bigger )) );
	ORACLE_TREE_FLUSH( "check/1216" );
	ORACLE_TREE_TREE( "check-state/1216", &tree );
	b2DynamicTree_Validate( &tree );

	b2DestroyDynamicTree( &tree );
	ORACLE_TREE_TREE( "step/1219", &tree );
	ORACLE_U32( "rng/end", treeRandomState );
	ORACLE_CLOSE();
	return 0;
}

int DynamicTreeTest( void )
{
	RUN_SUBTEST( TreeCreateDestroy );
	RUN_SUBTEST( TreeRayCastTest );
	RUN_SUBTEST( TreeMultipleProxiesTest );
	RUN_SUBTEST( TreeQueryTest );
	RUN_SUBTEST( TreeMoveAndEnlargeTest );
	RUN_SUBTEST( TreeRebuildAndValidateTest );
	RUN_SUBTEST( TreeRowHeightTest );
	RUN_SUBTEST( TreeGridHeightTest );
	RUN_SUBTEST( TreeGridMovementTest );
	RUN_SUBTEST( TreeOneProxyTest );
	RUN_SUBTEST( TreeLeafRootTest );
	RUN_SUBTEST( TreeQueryBruteForceTest );
	RUN_SUBTEST( TreeCastBruteForceTest );
	RUN_SUBTEST( TreeMarkAndRefitTest );
	RUN_SUBTEST( TreeDfsOrderTest );
	RUN_SUBTEST( TreeStaleAfterInsertRemoveTest );

	// todo test queries versus brute force

	return 0;
}
