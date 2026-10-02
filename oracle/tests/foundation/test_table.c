// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"
#include "core.h"
#include "platform.h"
#include "table.h"
#include "test_macros.h"

#include "box2d/base.h"

#define SET_SPAN 317
#define ITEM_COUNT ( ( SET_SPAN * SET_SPAN - SET_SPAN ) / 2 )

static int BasicHashSetTest( void )
{
	ORACLE_OPEN( "table/BasicHashSetTest" );
	// Test basic creation and destruction
	b2HashSet set = b2CreateSet( 16 );
	ENSURE( b2GetSetCount( &set ) == 0 );
	ORACLE_I32( "line_18", b2GetSetCount( &set ) );
	ENSURE( b2GetSetCapacity( &set ) == 16 );
	ORACLE_I32( "line_19", b2GetSetCapacity( &set ) );

	b2DestroySet( &set );
	ENSURE( set.items == NULL );
	ORACLE_BOOL( "line_22", set.items == NULL );
	ENSURE( set.count == 0 );
	ORACLE_I32( "line_23", set.count );
	ENSURE( set.capacity == 0 );
	ORACLE_I32( "line_24", set.capacity );

	ORACLE_CLOSE();
	return 0;
}

static int HashSetCapacityTest( void )
{
	ORACLE_OPEN( "table/HashSetCapacityTest" );
	// Test capacity adjustments - capacity should be power of 2
	{
		b2HashSet set = b2CreateSet( 1 );
		ENSURE( b2GetSetCapacity( &set ) == 16 ); // Minimum capacity
		ORACLE_I32( "line_34", b2GetSetCapacity( &set ) );
		b2DestroySet( &set );
	}

	{
		b2HashSet set = b2CreateSet( 15 );
		ENSURE( b2GetSetCapacity( &set ) == 16 ); // Should round up to 16
		ORACLE_I32( "line_40", b2GetSetCapacity( &set ) );
		b2DestroySet( &set );
	}

	{
		b2HashSet set = b2CreateSet( 32 );
		ENSURE( b2GetSetCapacity( &set ) == 32 ); // Should stay at 32
		ORACLE_I32( "line_46", b2GetSetCapacity( &set ) );
		b2DestroySet( &set );
	}

	{
		b2HashSet set = b2CreateSet( 33 );
		ENSURE( b2GetSetCapacity( &set ) == 64 ); // Should round up to 64
		ORACLE_I32( "line_52", b2GetSetCapacity( &set ) );
		b2DestroySet( &set );
	}

	ORACLE_CLOSE();
	return 0;
}

static int HashSetAddRemoveTest( void )
{
	ORACLE_OPEN( "table/HashSetAddRemoveTest" );
	b2HashSet set = b2CreateSet( 16 );

	// Test adding new keys
	bool found = b2AddKey( &set, 42 );
	ENSURE( found == false ); // Should be new
	ORACLE_BOOL( "line_65", found );
	ENSURE( b2GetSetCount( &set ) == 1 );
	ORACLE_I32( "line_66", b2GetSetCount( &set ) );

	found = b2AddKey( &set, 123 );
	ENSURE( found == false ); // Should be new
	ORACLE_BOOL( "line_69", found );
	ENSURE( b2GetSetCount( &set ) == 2 );
	ORACLE_I32( "line_70", b2GetSetCount( &set ) );

	// Test adding duplicate key
	found = b2AddKey( &set, 42 );
	ENSURE( found == true );			  // Should already exist
	ORACLE_BOOL( "line_74", found );
	ENSURE( b2GetSetCount( &set ) == 2 ); // Count shouldn't change
	ORACLE_I32( "line_75", b2GetSetCount( &set ) );

	// Test contains
	ENSURE( b2ContainsKey( &set, 42 ) == true );
	ORACLE_BOOL( "line_78", b2ContainsKey( &set, 42 ) );
	ENSURE( b2ContainsKey( &set, 123 ) == true );
	ORACLE_BOOL( "line_79", b2ContainsKey( &set, 123 ) );
	ENSURE( b2ContainsKey( &set, 999 ) == false );
	ORACLE_BOOL( "line_80", b2ContainsKey( &set, 999 ) );

	// Test removal
	bool removed = b2RemoveKey( &set, 42 );
	ENSURE( removed == true );
	ORACLE_BOOL( "line_84", removed );
	ENSURE( b2GetSetCount( &set ) == 1 );
	ORACLE_I32( "line_85", b2GetSetCount( &set ) );
	ENSURE( b2ContainsKey( &set, 42 ) == false );
	ORACLE_BOOL( "line_86", b2ContainsKey( &set, 42 ) );
	ENSURE( b2ContainsKey( &set, 123 ) == true );
	ORACLE_BOOL( "line_87", b2ContainsKey( &set, 123 ) );

	// Test removing non-existent key
	removed = b2RemoveKey( &set, 999 );
	ENSURE( removed == false );
	ORACLE_BOOL( "line_91", removed );
	ENSURE( b2GetSetCount( &set ) == 1 );
	ORACLE_I32( "line_92", b2GetSetCount( &set ) );

	// Test removing same key twice
	removed = b2RemoveKey( &set, 42 );
	ENSURE( removed == false );
	ORACLE_BOOL( "line_96", removed );
	ENSURE( b2GetSetCount( &set ) == 1 );
	ORACLE_I32( "line_97", b2GetSetCount( &set ) );

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetClearTest( void )
{
	ORACLE_OPEN( "table/HashSetClearTest" );
	b2HashSet set = b2CreateSet( 16 );

	// Add some keys
	b2AddKey( &set, 10 );
	b2AddKey( &set, 20 );
	b2AddKey( &set, 30 );
	ENSURE( b2GetSetCount( &set ) == 3 );
	ORACLE_I32( "line_111", b2GetSetCount( &set ) );

	// Clear the set
	b2ClearSet( &set );
	ENSURE( b2GetSetCount( &set ) == 0 );
	ORACLE_I32( "line_115", b2GetSetCount( &set ) );
	ENSURE( b2ContainsKey( &set, 10 ) == false );
	ORACLE_BOOL( "line_116", b2ContainsKey( &set, 10 ) );
	ENSURE( b2ContainsKey( &set, 20 ) == false );
	ORACLE_BOOL( "line_117", b2ContainsKey( &set, 20 ) );
	ENSURE( b2ContainsKey( &set, 30 ) == false );
	ORACLE_BOOL( "line_118", b2ContainsKey( &set, 30 ) );

	// Test that we can add keys after clearing
	b2AddKey( &set, 40 );
	ENSURE( b2GetSetCount( &set ) == 1 );
	ORACLE_I32( "line_122", b2GetSetCount( &set ) );
	ENSURE( b2ContainsKey( &set, 40 ) == true );
	ORACLE_BOOL( "line_123", b2ContainsKey( &set, 40 ) );

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetGrowthTest( void )
{
	ORACLE_OPEN( "table/HashSetGrowthTest" );
	b2HashSet set = b2CreateSet( 16 );
	int initialCapacity = b2GetSetCapacity( &set );

	// Add enough keys to trigger growth (load factor is 0.5)
	// With capacity 16, growth should happen when count reaches 8
	for ( uint64_t i = 0; i < 8; ++i )
	{
		b2AddKey( &set, i + 1 );
	}

	// Should have grown
	int newCapacity = b2GetSetCapacity( &set );
	ENSURE( newCapacity >= initialCapacity );
	ORACLE_I32( "line_143", newCapacity );
	ORACLE_I32( "line_143_initial", initialCapacity );
	ENSURE( b2GetSetCount( &set ) == 8 );
	ORACLE_I32( "line_144", b2GetSetCount( &set ) );

	// Verify all keys are still present after growth
	for ( uint64_t i = 1; i <= 8; ++i )
	{
		ENSURE( b2ContainsKey( &set, i ) == true );
		ORACLE_BOOL( "line_149", b2ContainsKey( &set, i ) );
	}

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetEdgeCasesTest( void )
{
	ORACLE_OPEN( "table/HashSetEdgeCasesTest" );
	b2HashSet set = b2CreateSet( 16 );

	// Test large key values
	uint64_t largeKey = 0xFFFFFFFFFFFFFFFFULL - 1; // Max value minus 1 (since 0 is sentinel)
	b2AddKey( &set, largeKey );
	ENSURE( b2ContainsKey( &set, largeKey ) == true );
	ORACLE_BOOL( "line_163", b2ContainsKey( &set, largeKey ) );
	ENSURE( b2GetSetCount( &set ) == 1 );
	ORACLE_I32( "line_164", b2GetSetCount( &set ) );

	// Test keys that might cause hash collisions
	uint64_t key1 = 0x123456789ABCDEFULL;
	uint64_t key2 = 0x987654321FEDCBAULL;
	b2AddKey( &set, key1 );
	b2AddKey( &set, key2 );
	ENSURE( b2ContainsKey( &set, key1 ) == true );
	ORACLE_BOOL( "line_171", b2ContainsKey( &set, key1 ) );
	ENSURE( b2ContainsKey( &set, key2 ) == true );
	ORACLE_BOOL( "line_172", b2ContainsKey( &set, key2 ) );

	// Test pattern that could cause clustering
	for ( uint64_t i = 0x1000; i < 0x1010; ++i )
	{
		b2AddKey( &set, i );
	}

	for ( uint64_t i = 0x1000; i < 0x1010; ++i )
	{
		ENSURE( b2ContainsKey( &set, i ) == true );
		ORACLE_BOOL( "line_182", b2ContainsKey( &set, i ) );
	}

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetRemovalReorganizationTest( void )
{
	ORACLE_OPEN( "table/HashSetRemovalReorganizationTest" );
	b2HashSet set = b2CreateSet( 16 );

	// Add keys that might cluster together
	uint64_t keys[] = { 100, 116, 132, 148, 164 }; // These might hash to similar slots
	int keyCount = sizeof( keys ) / sizeof( keys[0] );

	for ( int i = 0; i < keyCount; ++i )
	{
		b2AddKey( &set, keys[i] );
	}

	// Verify all keys are present
	for ( int i = 0; i < keyCount; ++i )
	{
		ENSURE( b2ContainsKey( &set, keys[i] ) == true );
		ORACLE_BOOL( "line_205", b2ContainsKey( &set, keys[i] ) );
	}

	// Remove a key from the middle
	b2RemoveKey( &set, keys[2] );
	ENSURE( b2ContainsKey( &set, keys[2] ) == false );
	ORACLE_BOOL( "line_210", b2ContainsKey( &set, keys[2] ) );

	// Verify other keys are still present (tests reorganization)
	for ( int i = 0; i < keyCount; ++i )
	{
		if ( i != 2 )
		{
			ENSURE( b2ContainsKey( &set, keys[i] ) == true );
			ORACLE_BOOL( "line_217", b2ContainsKey( &set, keys[i] ) );
		}
	}

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

#define TEST_SIZE 1000

static int HashSetStressTest( void )
{
	ORACLE_OPEN( "table/HashSetStressTest" );
	b2HashSet set = b2CreateSet( 32 );

	const int testSize = TEST_SIZE;
	uint64_t keys[TEST_SIZE];

	// Generate test keys
	for ( int i = 0; i < testSize; ++i )
	{
		keys[i] = (uint64_t)( i * 7 + 13 ); // Some pattern to avoid zero
	}

	// Add all keys
	for ( int i = 0; i < testSize; ++i )
	{
		bool found = b2AddKey( &set, keys[i] );
		ENSURE( found == false );
		ORACLE_BOOL( "line_244", found );
	}

	ENSURE( b2GetSetCount( &set ) == testSize );
	ORACLE_I32( "line_247", b2GetSetCount( &set ) );

	// Verify all keys are present
	for ( int i = 0; i < testSize; ++i )
	{
		ENSURE( b2ContainsKey( &set, keys[i] ) == true );
		ORACLE_BOOL( "line_252", b2ContainsKey( &set, keys[i] ) );
	}

	// Remove every other key
	int removedCount = 0;
	for ( int i = 0; i < testSize; i += 2 )
	{
		bool removed = b2RemoveKey( &set, keys[i] );
		ENSURE( removed == true );
		ORACLE_BOOL( "line_260", removed );
		removedCount++;
	}

	ENSURE( b2GetSetCount( &set ) == testSize - removedCount );
	ORACLE_I32( "line_264", b2GetSetCount( &set ) );

	// Verify remaining keys are still present
	for ( int i = 0; i < testSize; ++i )
	{
		bool shouldBePresent = ( i % 2 == 1 );
		ENSURE( b2ContainsKey( &set, keys[i] ) == shouldBePresent );
		ORACLE_BOOL( "line_270", b2ContainsKey( &set, keys[i] ) );
	}

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetShapePairKeyTest( void )
{
	ORACLE_OPEN( "table/HashSetShapePairKeyTest" );
	b2HashSet set = b2CreateSet( 16 );

	// Test the B2_SHAPE_PAIR_KEY macro
	uint64_t key1 = B2_SHAPE_PAIR_KEY( 5, 10 );
	uint64_t key2 = B2_SHAPE_PAIR_KEY( 10, 5 ); // Should be same as key1
	ENSURE( key1 == key2 );
	ORACLE_U64( "line_284", key1 );
	ORACLE_U64( "line_284_other", key2 );

	b2AddKey( &set, key1 );
	ENSURE( b2ContainsKey( &set, key1 ) == true );
	ORACLE_BOOL( "line_287", b2ContainsKey( &set, key1 ) );
	ENSURE( b2ContainsKey( &set, key2 ) == true ); // Should find same key
	ORACLE_BOOL( "line_288", b2ContainsKey( &set, key2 ) );

	// Test different pairs
	uint64_t key3 = B2_SHAPE_PAIR_KEY( 1, 2 );
	uint64_t key4 = B2_SHAPE_PAIR_KEY( 2, 3 );
	ENSURE( key3 != key4 );
	ORACLE_U64( "line_293", key3 );
	ORACLE_U64( "line_293_other", key4 );

	b2AddKey( &set, key3 );
	b2AddKey( &set, key4 );
	ENSURE( b2GetSetCount( &set ) == 3 );
	ORACLE_I32( "line_297", b2GetSetCount( &set ) );

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetBytesTest( void )
{
	ORACLE_OPEN( "table/HashSetBytesTest" );
	b2HashSet set = b2CreateSet( 32 );

	int bytes = b2GetHashSetBytes( &set );
	int expectedBytes = 32 * (int)sizeof( b2SetItem );
	ENSURE( bytes == expectedBytes );
	ORACLE_I32( "line_309", bytes );

	// Add some items and verify bytes calculation doesn't change
	b2AddKey( &set, 100 );
	b2AddKey( &set, 200 );

	int bytesAfterAdd = b2GetHashSetBytes( &set );
	ENSURE( bytesAfterAdd == expectedBytes );
	ORACLE_I32( "line_316", bytesAfterAdd );

	b2DestroySet( &set );
	ORACLE_CLOSE();
	return 0;
}

static int HashSetTest( void )
{
	ORACLE_OPEN( "table/HashSetTest" );
	const int N = SET_SPAN;
	const uint32_t itemCount = ITEM_COUNT;
	bool removed[ITEM_COUNT] = { 0 };

	for ( int iter = 0; iter < 1; ++iter )
	{
		b2HashSet set = b2CreateSet( 16 );

		// Fill set
		for ( int i = 0; i < N; ++i )
		{
			for ( int j = i + 1; j < N; ++j )
			{
				uint64_t key = B2_SHAPE_PAIR_KEY( i, j );
				bool found = b2AddKey( &set, key );
				ENSURE( found == false );
				ORACLE_FOUNDATION_BUFFER( found );
			}
		}

		ORACLE_FOUNDATION_FLUSH( "fill" );
		ENSURE( b2GetSetCount( &set ) == itemCount );
		ORACLE_I32( "line_343", b2GetSetCount( &set ) );

		// Remove a portion of the set
		int k = 0;
		uint32_t removeCount = 0;
		for ( int i = 0; i < N; ++i )
		{
			for ( int j = i + 1; j < N; ++j )
			{
				if ( j == i + 1 )
				{
					uint64_t key = B2_SHAPE_PAIR_KEY( i, j );
					int size1 = b2GetSetCount( &set );
					bool found = b2RemoveKey( &set, key );
					ENSURE( found );
					ORACLE_BOOL( "line_357", found );
					int size2 = b2GetSetCount( &set );
					ENSURE( size2 == size1 - 1 );
					ORACLE_I32( "line_359", size2 );
					removed[k++] = true;
					removeCount += 1;
				}
				else
				{
					removed[k++] = false;
				}
			}
		}

		ENSURE( b2GetSetCount( &set ) == ( itemCount - removeCount ) );
		ORACLE_I32( "line_370", b2GetSetCount( &set ) );

#if B2_SNOOP_TABLE_COUNTERS
		extern b2AtomicInt b2_probeCount;
		b2AtomicStoreInt( &b2_probeCount, 0 );
#endif

		// Test key search
		// ~5ns per search on an AMD 7950x
		uint64_t ticks = b2GetTicks();

		k = 0;
		for ( int i = 0; i < N; ++i )
		{
			for ( int j = i + 1; j < N; ++j )
			{
				uint64_t key = B2_SHAPE_PAIR_KEY( j, i );
				bool found = b2ContainsKey( &set, key );
				ENSURE( found || removed[k] );
				ORACLE_FOUNDATION_BUFFER( found );
				ORACLE_FOUNDATION_BUFFER( removed[k] );
				k += 1;
			}
		}

		ORACLE_FOUNDATION_FLUSH( "search" );
		// uint64_t ticks = b2GetTicks(&timer);
		// printf("set ticks = %llu\n", ticks);

		float ms = b2GetMilliseconds( ticks );
		printf( "set: count = %d, b2ContainsKey = %.5f ms, ave = %.5f us\n", itemCount, ms, 1000.0f * ms / itemCount );

#if B2_SNOOP_TABLE_COUNTERS
		int probeCount = b2AtomicLoadInt( &b2_probeCount );
		float aveProbeCount = (float)probeCount / (float)itemCount;
		printf( "item count = %d, probe count = %d, ave probe count %.2f\n", itemCount, probeCount, aveProbeCount );
#endif

		// Remove all keys from set
		for ( int i = 0; i < N; ++i )
		{
			for ( int j = i + 1; j < N; ++j )
			{
				uint64_t key = B2_SHAPE_PAIR_KEY( i, j );
				b2RemoveKey( &set, key );
			}
		}

		ENSURE( b2GetSetCount( &set ) == 0 );
		ORACLE_I32( "line_415", b2GetSetCount( &set ) );

		b2DestroySet( &set );
	}

	ORACLE_CLOSE();
	return 0;
}

int TableTest( void )
{
	ORACLE_OPEN( "table/TableTest" );
	// Test helper functions first
	int power = b2BoundingPowerOf2( 3008 );
	ENSURE( power == 12 );
	ORACLE_I32( "line_427", power );

	int nextPowerOf2 = b2RoundUpPowerOf2( 3008 );
	ENSURE( nextPowerOf2 == ( 1 << power ) );
	ORACLE_I32( "line_430", nextPowerOf2 );

	ORACLE_CLOSE();
	// Run all hash set tests
	RUN_SUBTEST( BasicHashSetTest );
	RUN_SUBTEST( HashSetCapacityTest );
	RUN_SUBTEST( HashSetAddRemoveTest );
	RUN_SUBTEST( HashSetClearTest );
	RUN_SUBTEST( HashSetGrowthTest );
	RUN_SUBTEST( HashSetEdgeCasesTest );
	RUN_SUBTEST( HashSetRemovalReorganizationTest );
	RUN_SUBTEST( HashSetStressTest );
	RUN_SUBTEST( HashSetShapePairKeyTest );
	RUN_SUBTEST( HashSetBytesTest );
	RUN_SUBTEST( HashSetTest );

	return 0;
}
