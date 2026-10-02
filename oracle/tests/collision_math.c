// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"
#include <assert.h>
#include <string.h>

static uint8_t collision_data[1048576];
static size_t collision_bit;
static uint32_t collision_previous[1024];
static unsigned collision_leading[1024];
static unsigned collision_width[1024];

static void collision_bits( uint32_t value, unsigned count )
{
	assert( collision_bit + count <= 8 * sizeof( collision_data ) );
	for ( unsigned i = 0; i < count; ++i, ++collision_bit )
	{
		collision_data[collision_bit / 8] |= (uint8_t)( ( ( value >> i ) & 1 ) << ( collision_bit % 8 ) );
	}
}

static void collision_f32( unsigned site, float value )
{
	uint32_t bits;
	memcpy( &bits, &value, sizeof( bits ) );
	uint32_t delta = bits ^ collision_previous[site];
	collision_previous[site] = bits;
	if ( delta == 0 )
	{
		collision_bits( 0, 1 );
		return;
	}
	unsigned leading = 0;
	unsigned trailing = 0;
	for ( uint32_t mask = UINT32_C( 0x80000000 ); ( delta & mask ) == 0; mask >>= 1 )
	{
		++leading;
	}
	for ( uint32_t mask = 1; ( delta & mask ) == 0; mask <<= 1 )
	{
		++trailing;
	}
	unsigned width = 32 - leading - trailing;
	collision_bits( 1, 1 );
	unsigned previousWidth = collision_width[site];
	unsigned previousTrailing = 32 - collision_leading[site] - previousWidth;
	if ( previousWidth != 0 && leading >= collision_leading[site] && trailing >= previousTrailing && previousWidth <= width + 10 )
	{
		collision_bits( 0, 1 );
		collision_bits( delta >> previousTrailing, previousWidth );
		return;
	}
	collision_bits( 1, 1 );
	collision_leading[site] = leading;
	collision_width[site] = width;
	collision_bits( leading, 5 );
	collision_bits( width - 1, 5 );
	collision_bits( delta >> trailing, width );
}

#define ORACLE_COLLISION_F32( site, value ) collision_f32( site, value )
#define ORACLE_COLLISION_BOOL( site, value ) collision_bits( value, 1 )
#define ORACLE_COLLISION_CLOSE() do { ORACLE_U32( "bits", (uint32_t)collision_bit ); ORACLE_BYTES( "checks", collision_data, ( collision_bit + 7 ) / 8 ); ORACLE_CLOSE(); } while ( false )

#include "collision/test_math.c"

int main( void )
{
	return MathTest();
}
