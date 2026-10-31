// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#define _POSIX_C_SOURCE 200809L
#include "body.h"
#include "oracle.h"
#include "physics_world.h"

#include "box2d/box2d.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "replay_events.h"
#include "replay_observe.h"
#include "replay_properties.h"

static const char* recording_diagPath = "diverge.b2rec";

static uint8_t* recording_states[1024];
static size_t recording_stateSizes[1024];
static int recording_stateCount;

// Named-field trace of all frames: exact canonical/transform hashes and complete
// event identity/payload digests.
static const char recording_trace_schema[] =
	"transform,canonicalState,moveCount,moveHash,sensorBeginCount,sensorEndCount,sensorHash,"
	"contactBeginCount,contactEndCount,hitCount,contactHash,jointCount,jointHash,frameIdentity";
static uint8_t* recording_history;
static size_t recording_historySize;
static size_t recording_historyCapacity;

static void recording_add_row( const uint64_t* row )
{
	if ( recording_historySize + 14 * 8 > recording_historyCapacity )
	{
		recording_historyCapacity = recording_historyCapacity ? 2 * recording_historyCapacity : 1024 * 1024;
		recording_history = realloc( recording_history, recording_historyCapacity );
		if ( recording_history == NULL )
			abort();
	}
	for ( int field = 0; field < 14; ++field )
		for ( int i = 0; i < 8; ++i )
			recording_history[recording_historySize++] = (uint8_t)( row[field] >> ( 8 * i ) );
}

static void recording_end_trace( void )
{
	ORACLE_BYTES( "trace/schema", recording_trace_schema, sizeof( recording_trace_schema ) - 1 );
	ORACLE_BYTES( "trace/rows-14u64-le", recording_history, recording_historySize );
}

static void recording_reset_trace( void )
{
	for ( int i = 0; i < recording_stateCount; ++i )
	{
		free( recording_states[i] );
	}
	recording_stateCount = 0;
	recording_historySize = 0;
}

static void recording_put( uint8_t* data, size_t* size, uint64_t value, int width )
{
	for ( int i = 0; i < width; ++i )
		data[( *size )++] = (uint8_t)( value >> ( 8 * i ) );
}

static void recording_trace( b2WorldId id, const char* prefix, int frame )
{
	// Nested observers reuse ORACLE_KEY's scratch buffers before this prefix's last use.
	char ownedPrefix[256];
	snprintf( ownedPrefix, sizeof( ownedPrefix ), "%s", prefix );
	prefix = ownedPrefix;
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
		uint8_t* raw = malloc( 5 + (size_t)count * 28 );
		if ( raw == NULL )
		{
			abort();
		}
		size_t size = 0;
#if defined( BOX2D_DOUBLE_PRECISION )
		const int width = 8;
#else
		const int width = 4;
#endif
		raw[size++] = (uint8_t)width;
		recording_put( raw, &size, (uint32_t)count, 4 );
		for ( int i = 0; i < world->bodies.count; ++i )
		{
			b2Body* body = world->bodies.data + i;
			if ( body->id != i )
			{
				continue;
			}
			b2WorldTransform xf = b2GetBodyTransformQuick( world, body );
			recording_put( raw, &size, (uint32_t)i, 4 );
#if defined( BOX2D_DOUBLE_PRECISION )
			uint64_t bits[4] = { oracle_bits_f64( xf.p.x ), oracle_bits_f64( xf.p.y ), oracle_bits_f32( xf.q.c ),
								 oracle_bits_f32( xf.q.s ) };
#else
			uint64_t bits[4] = { oracle_bits_f32( xf.p.x ), oracle_bits_f32( xf.p.y ), oracle_bits_f32( xf.q.c ),
								 oracle_bits_f32( xf.q.s ) };
#endif
			for ( int component = 0; component < 4; ++component )
			{
				recording_put( raw, &size, bits[component], component < 2 ? width : 4 );
			}
		}
		int reference = 0;
		while ( reference < recording_stateCount &&
				( recording_stateSizes[reference] != size || memcmp( recording_states[reference], raw, size ) != 0 ) )
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
			ORACLE_BYTES( ORACLE_KEY( "%s/%d/state", prefix, frame ), raw, size );
		}
		else
		{
			free( raw );
		}
	}
	if ( frame % 60 == 0 )
		ORACLE_U64( ORACLE_KEY( "%s/%d/xf", prefix, frame ), hash );
	uint64_t row[14] = { hash, b2World_GetStateHash( id ) };
	replay_event_row events = replay_event_observe( id );
	for ( int i = 0; i < 11; ++i )
		row[i + 2] = events.fields[i];
	char traceKey[320];
	snprintf( traceKey, sizeof( traceKey ), "%s/%d", prefix, frame );
	row[13] = oracle_hash_bytes( ORACLE_HASH_INIT, traceKey, strlen( traceKey ) );
	recording_add_row( row );
	if ( frame == 1 || frame == 31 || frame % 60 == 0 )
	{
		replay_properties( id, ORACLE_KEY( "%s/%d/properties", prefix, frame ) );
	}
}

#define ORACLE_RECORDING_TRACE( world, prefix, frame ) recording_trace( ( world ), ( prefix ), ( frame ) )

static uint64_t ReplayDeepHash( b2Replay* player );

#include "upstream/test_recording.c"

int main( void )
{
	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingTest" );
	int RecordingTestResult = RecordingTest();
	recording_end_trace();
	ORACLE_CLOSE();
	if ( RecordingTestResult != 0 )
	{
		return RecordingTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingOutlinerTest" );
	int RecordingOutlinerTestResult = RecordingOutlinerTest();
	recording_end_trace();
	ORACLE_CLOSE();
	if ( RecordingOutlinerTestResult != 0 )
	{
		return RecordingOutlinerTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingKeyframeTest" );
	int RecordingKeyframeTestResult = RecordingKeyframeTest();
	recording_end_trace();
	ORACLE_CLOSE();
	if ( RecordingKeyframeTestResult != 0 )
	{
		return RecordingKeyframeTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingScrubTest" );
	int RecordingScrubTestResult = RecordingScrubTest();
	recording_end_trace();
	ORACLE_CLOSE();
	if ( RecordingScrubTestResult != 0 )
	{
		return RecordingScrubTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/RecordingQueryScrubTest" );
	int RecordingQueryScrubTestResult = RecordingQueryScrubTest();
	recording_end_trace();
	ORACLE_CLOSE();
	if ( RecordingQueryScrubTestResult != 0 )
	{
		return RecordingQueryScrubTestResult;
	}

	recording_reset_trace();
	ORACLE_OPEN( "recording/ReStepRaceTest" );
	int ReStepRaceTestResult = ReStepRaceTest();
	recording_end_trace();
	ORACLE_CLOSE();
	if ( ReStepRaceTestResult != 0 )
	{
		return ReStepRaceTestResult;
	}
	recording_reset_trace();
	// ReplayFileScrubDiag is a source diagnostic, NOT one of the six registered
	// upstream recording cases. Exercise it with isolated deterministic native data.
	ORACLE_OPEN( "recording/DiagnosticUnregistered" );
	char diagDir[] = "/tmp/box2d-diag-XXXXXX";
	if ( mkdtemp( diagDir ) == NULL )
	{
		return 1;
	}
	char diagPath[256];
	snprintf( diagPath, sizeof( diagPath ), "%s/diverge.b2rec", diagDir );
	recording_diagPath = diagPath;
	int missingResult = ReplayFileScrubDiag();
	ORACLE_I32( "missing/result", missingResult );
	if ( missingResult != 0 )
	{
		return missingResult;
	}
	b2Recording* diagRecording = RecordScene( BuildPyramidScene, 1, 12 );
	if ( b2SaveRecordingToFile( diagRecording, diagPath ) == false )
	{
		return 1;
	}
	b2DestroyRecording( diagRecording );
	int diagResult = ReplayFileScrubDiag();
	ORACLE_I32( "present/result", diagResult );
	if ( remove( diagPath ) != 0 || rmdir( diagDir ) != 0 )
	{
		return 1;
	}
	recording_end_trace();
	ORACLE_CLOSE();
	recording_reset_trace();
	if ( diagResult != 0 )
	{
		return diagResult;
	}
	return 0;
}
