// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

#if !defined( TREE_NO_ORACLE )
#include "oracle.h"
#include "dynamic_tree.h"

#include <assert.h>

static uint8_t treeChecks[4096];
static size_t treeCheckSize;
static int treeLastRebuild;

static void treeCheck( uint16_t line, uint64_t value )
{
	assert( treeCheckSize + 13 <= sizeof( treeChecks ) );
	treeChecks[treeCheckSize++] = (uint8_t)line;
	treeChecks[treeCheckSize++] = (uint8_t)( line >> 8 );
	do
	{
		uint8_t byte = (uint8_t)( value & 127 );
		value >>= 7;
		treeChecks[treeCheckSize++] = value ? (uint8_t)( byte | 128 ) : byte;
	}
	while ( value );
	treeChecks[treeCheckSize++] = 1;
}

static void treeFlush( const char* key )
{
	ORACLE_BYTES( key, treeChecks, treeCheckSize );
	treeCheckSize = 0;
}

static int treeRebuild( b2DynamicTree* tree, bool fullBuild )
{
	treeLastRebuild = b2DynamicTree_Rebuild( tree, fullBuild );
	return treeLastRebuild;
}

#define b2DynamicTree_Rebuild( tree, fullBuild ) treeRebuild( tree, fullBuild )
static uint32_t treeRandomState;

static uint64_t treeHashBox( uint64_t hash, b2AABB box )
{
	hash = oracle_hash_f32( hash, box.lowerBound.x );
	hash = oracle_hash_f32( hash, box.lowerBound.y );
	hash = oracle_hash_f32( hash, box.upperBound.x );
	return oracle_hash_f32( hash, box.upperBound.y );
}

static uint64_t treeHashNode( uint64_t hash, const b2DynamicTree* tree, int index )
{
	const b2TreeNode* node = tree->nodes + index;
	hash = oracle_hash_u32( hash, (uint32_t)index );
	hash = oracle_hash_u32( hash, (uint32_t)tree->parents[index] );
	hash = oracle_hash_u32( hash, node->flagIndex );
	hash = treeHashBox( hash, node->aabb );
	if ( node->flagIndex == B2_EMPTY_NODE )
	{
		return hash;
	}
	if ( ( node->flagIndex & B2_LEAF_NODE ) != 0 )
	{
		hash = oracle_hash_u32( hash, (uint32_t)node->shapeIndex );
		int proxyId = (int)( node->flagIndex & B2_NODE_INDEX_MASK );
		const b2TreeProxy* proxy = tree->proxies + proxyId;
		hash = oracle_hash_u64( hash, proxy->userData );
		hash = oracle_hash_u64( hash, proxy->categoryBits );
		return oracle_hash_u32( hash, (uint32_t)proxy->node );
	}
	hash = oracle_hash_u32( hash, (uint32_t)node->height );
	int pair = (int)( node->flagIndex & B2_NODE_INDEX_MASK );
	hash = treeHashNode( hash, tree, pair );
	return treeHashNode( hash, tree, pair + 1 );
}

static void treeTree( const char* key, const b2DynamicTree* tree )
{
	uint64_t hash = ORACLE_HASH_INIT;
	hash = oracle_hash_u32( hash, (uint32_t)tree->nodeEnd );
	hash = oracle_hash_u32( hash, (uint32_t)tree->nodeCapacity );
	hash = oracle_hash_u32( hash, (uint32_t)tree->pairFreeList );
	hash = oracle_hash_u32( hash, (uint32_t)tree->proxyCount );
	hash = oracle_hash_u32( hash, (uint32_t)tree->proxyCapacity );
	hash = oracle_hash_u32( hash, (uint32_t)tree->proxyFreeList );
	hash = oracle_hash_u32( hash, tree->dfsOrdered );
	if ( tree->nodeEnd > 0 )
	{
		hash = treeHashNode( hash, tree, B2_ROOT_NODE );
	}
	ORACLE_U64( key, hash );
}

#define ORACLE_TREE_BEGIN() ( treeCheckSize = 0, ORACLE_U32( "rng/start", treeRandomState ) )
#define ORACLE_TREE_CHECK( line, value ) treeCheck( ( line ), ( value ) )
#define ORACLE_TREE_FLUSH( key ) treeFlush( key )
#define ORACLE_TREE_TREE( key, tree ) treeTree( key, tree )
#define ORACLE_TREE_BOX( key, box ) treeBox( key, ( box ) )
#define ORACLE_TREE_HITS( hits ) treeHits( ( hits )->hit, ( hits )->count, ( hits )->bad )

static void treeHits( const bool* hit, int count, bool bad )
{
	uint8_t packed[30] = { 0 };
	for ( int i = 0; i < 200; ++i )
	{
		packed[i / 8] |= (uint8_t)( hit[i] << ( i % 8 ) );
	}
	for ( int i = 0; i < 4; ++i )
	{
		packed[25 + i] = (uint8_t)( (uint32_t)count >> ( 8 * i ) );
	}
	packed[29] = bad;
	ORACLE_BYTES( "query/hits", packed, sizeof( packed ) );
}

static void treeBox( const char* key, b2AABB box )
{
	uint32_t bits[4] = { oracle_bits_f32( box.lowerBound.x ), oracle_bits_f32( box.lowerBound.y ),
		oracle_bits_f32( box.upperBound.x ), oracle_bits_f32( box.upperBound.y ) };
	uint8_t packed[16];
	for ( int i = 0; i < 4; ++i )
	{
		for ( int j = 0; j < 4; ++j )
		{
			packed[4 * i + j] = (uint8_t)( bits[i] >> ( 8 * j ) );
		}
	}
	ORACLE_BYTES( key, packed, sizeof( packed ) );
}

#else
#define ORACLE_OPEN( ... ) ( (void)0 )
#define ORACLE_CLOSE( ... ) ( (void)0 )
#define ORACLE_U32( ... ) ( (void)0 )
#define ORACLE_F32( ... ) ( (void)0 )
#define ORACLE_VEC2( ... ) ( (void)0 )
#define ORACLE_TREE_BEGIN( ... ) ( (void)0 )
#define ORACLE_TREE_CHECK( ... ) ( (void)0 )
#define ORACLE_TREE_FLUSH( ... ) ( (void)0 )
#define ORACLE_TREE_TREE( ... ) ( (void)0 )
#define ORACLE_TREE_BOX( ... ) ( (void)0 )
#define ORACLE_TREE_HITS( ... ) ( (void)0 )
#endif

#include "upstream/test_dynamic_tree.c"

int main( void )
{
	return DynamicTreeTest();
}
