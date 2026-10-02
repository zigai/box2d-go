// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"
#include "box2d/box2d.h"

#include "body.h"
#include "physics_world.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static uint8_t* recording_states[1024];
static size_t recording_stateSizes[1024];
static int recording_stateCount;

static void recording_reset_trace( void )
{
	for ( int i = 0; i < recording_stateCount; ++i )
	{
		free( recording_states[i] );
	}
	recording_stateCount = 0;
}

static void recording_varint( uint8_t* data, size_t* size, uint64_t value )
{
	while ( value >= 128 )
	{
		data[( *size )++] = (uint8_t)( value | 128 );
		value >>= 7;
	}
	data[( *size )++] = (uint8_t)value;
}

static uint32_t recording_lz_key( const uint8_t* data )
{
	return ( (uint32_t)data[0] * 251u + (uint32_t)data[1] * 31u + data[2] ) & 65535u;
}

static uint8_t* recording_compress( const uint8_t* raw, size_t size, size_t* packedSize )
{
	uint8_t* packed = malloc( size * 2 + 16 );
	int* recent = malloc( 65536 * sizeof( int ) );
	if ( packed == NULL || recent == NULL )
	{
		abort();
	}
	for ( int i = 0; i < 65536; ++i )
	{
		recent[i] = -1;
	}
	size_t out = 0;
	size_t pos = 0;
	size_t literal = 0;
	while ( pos < size )
	{
		size_t length = 0;
		size_t distance = 0;
		if ( pos + 2 < size )
		{
			int candidate = recent[recording_lz_key( raw + pos )];
			if ( candidate >= 0 && pos - (size_t)candidate <= 65535 )
			{
				distance = pos - (size_t)candidate;
				while ( length < 130 && pos + length < size && raw[candidate + length] == raw[pos + length] )
				{
					++length;
				}
			}
		}
		if ( length >= 4 )
		{
			if ( pos > literal )
			{
				packed[out++] = (uint8_t)( pos - literal - 1 );
				memcpy( packed + out, raw + literal, pos - literal );
				out += pos - literal;
			}
			packed[out++] = (uint8_t)( 128 + length - 3 );
			packed[out++] = (uint8_t)distance;
			packed[out++] = (uint8_t)( distance >> 8 );
			for ( size_t j = 0; j < length; ++j )
			{
				if ( pos + j + 2 < size )
				{
					recent[recording_lz_key( raw + pos + j )] = (int)( pos + j );
				}
			}
			pos += length;
			literal = pos;
		}
		else
		{
			if ( pos + 2 < size )
			{
				recent[recording_lz_key( raw + pos )] = (int)pos;
			}
			++pos;
			if ( pos - literal == 128 || pos == size )
			{
				packed[out++] = (uint8_t)( pos - literal - 1 );
				memcpy( packed + out, raw + literal, pos - literal );
				out += pos - literal;
				literal = pos;
			}
		}
	}
	free( recent );
	*packedSize = out;
	return packed;
}

static void recording_trace( b2WorldId id, const char* prefix, int frame )
{
	b2World* world = b2GetWorldFromId( id );
	uint64_t hash = ORACLE_HASH_INIT;
	int count = 0;
	for ( int i = 0; i < world->bodies.count; ++i )
	{
		b2Body* body = world->bodies.data + i;
		if ( body->id == i )
		{
			b2WorldTransform xf = b2GetBodyTransformQuick( world, body );
			hash = oracle_hash_transform( hash, xf.p, xf.q );
			++count;
		}
	}
	if ( frame % 60 == 0 )
	{
		uint8_t* raw = malloc( (size_t)count * 45 + 16 );
		if ( raw == NULL )
		{
			abort();
		}
		size_t size = 0;
#if defined( BOX2D_DOUBLE_PRECISION )
		raw[size++] = 8;
#else
		raw[size++] = 4;
#endif
		recording_varint( raw, &size, (uint64_t)count );
		int previousIndex = -1;
		uint64_t previous[4] = { 0, 0, 0, 0 };
		for ( int i = 0; i < world->bodies.count; ++i )
		{
			b2Body* body = world->bodies.data + i;
			if ( body->id != i )
			{
				continue;
			}
			b2WorldTransform xf = b2GetBodyTransformQuick( world, body );
			recording_varint( raw, &size, (uint64_t)( i - previousIndex ) );
			previousIndex = i;
#if defined( BOX2D_DOUBLE_PRECISION )
			uint64_t bits[4] = { oracle_bits_f64( xf.p.x ), oracle_bits_f64( xf.p.y ), oracle_bits_f32( xf.q.c ), oracle_bits_f32( xf.q.s ) };
#else
			uint64_t bits[4] = { oracle_bits_f32( xf.p.x ), oracle_bits_f32( xf.p.y ), oracle_bits_f32( xf.q.c ), oracle_bits_f32( xf.q.s ) };
#endif
			for ( int component = 0; component < 4; ++component )
			{
				recording_varint( raw, &size, bits[component] ^ previous[component] );
				previous[component] = bits[component];
			}
		}
		int reference = 0;
		while ( reference < recording_stateCount && ( recording_stateSizes[reference] != size || memcmp( recording_states[reference], raw, size ) != 0 ) )
		{
			++reference;
		}
		ORACLE_I32( ORACLE_KEY( "%s/%d/state/ref", prefix, frame ), reference );
		if ( reference == recording_stateCount )
		{
			if ( recording_stateCount == 1024 )
			{
				abort();
			}
			recording_states[recording_stateCount] = raw;
			recording_stateSizes[recording_stateCount++] = size;
			size_t packedSize;
			uint8_t* packed = recording_compress( raw, size, &packedSize );
			ORACLE_BYTES( ORACLE_KEY( "%s/%d/state/packed", prefix, frame ), packed, packedSize );
			free( packed );
		}
		else
		{
			free( raw );
		}
	}
	ORACLE_U64( ORACLE_KEY( "%s/%d/xf", prefix, frame ), hash );
}

#define ORACLE_RECORDING_TRACE( world, prefix, frame ) recording_trace( ( world ), ( prefix ), ( frame ) )

static uint64_t ReplayDeepHash( b2Replay* player );

#include "recording/test_recording.c"

int main( void )
{
	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingTest" );
	int RecordingTestResult = RecordingTest();
	ORACLE_CLOSE();
	if ( RecordingTestResult != 0 )
	{
		return RecordingTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingOutlinerTest" );
	int RecordingOutlinerTestResult = RecordingOutlinerTest();
	ORACLE_CLOSE();
	if ( RecordingOutlinerTestResult != 0 )
	{
		return RecordingOutlinerTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingKeyframeTest" );
	int RecordingKeyframeTestResult = RecordingKeyframeTest();
	ORACLE_CLOSE();
	if ( RecordingKeyframeTestResult != 0 )
	{
		return RecordingKeyframeTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingScrubTest" );
	int RecordingScrubTestResult = RecordingScrubTest();
	ORACLE_CLOSE();
	if ( RecordingScrubTestResult != 0 )
	{
		return RecordingScrubTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingQueryScrubTest" );
	int RecordingQueryScrubTestResult = RecordingQueryScrubTest();
	ORACLE_CLOSE();
	if ( RecordingQueryScrubTestResult != 0 )
	{
		return RecordingQueryScrubTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/ReStepRaceTest" );
	int ReStepRaceTestResult = ReStepRaceTest();
	ORACLE_CLOSE();
	if ( ReStepRaceTestResult != 0 )
	{
		return ReStepRaceTestResult;
	}
	recording_reset_trace();
	return 0;
}
