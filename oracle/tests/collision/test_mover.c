// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

#include "test_macros.h"

#include "box2d/collision.h"
#include "box2d/constants.h"
#include "box2d/math_functions.h"

#include <float.h>

static b2CollisionPlane MakePlane( b2Vec2 normal, float pushLimit )
{
	return ( b2CollisionPlane ){
		.plane = { normal, 0.0f },
		.pushLimit = pushLimit,
		.push = 0.0f,
		.clipVelocity = true,
	};
}

static int NoPlanesTest( void )
{
	ORACLE_OPEN( "mover/NoPlanesTest" );
	b2Vec2 targetDelta = { 0.5f, -0.25f };
	b2PlaneSolverResult result = b2SolvePlanes( targetDelta, NULL, 0 );

	ORACLE_F32( "mover:27/value0", result.delta.x );
	ORACLE_F32( "mover:27/value1", targetDelta.x );
	ORACLE_BOOL( "mover:27", result.delta.x == targetDelta.x );
	ENSURE( result.delta.x == targetDelta.x );
	ORACLE_F32( "mover:29/value0", result.delta.y );
	ORACLE_F32( "mover:29/value1", targetDelta.y );
	ORACLE_BOOL( "mover:29", result.delta.y == targetDelta.y );
	ENSURE( result.delta.y == targetDelta.y );
	ORACLE_I32( "mover:30/value0", result.iterationCount );
	ORACLE_BOOL( "mover:30", result.iterationCount == 0 );
	ENSURE( result.iterationCount == 0 );

	ORACLE_CLOSE();
	return 0;
}

static int SolvePlaneTest( void )
{
	ORACLE_OPEN( "mover/SolvePlaneTest" );
	b2CollisionPlane plane = MakePlane( ( b2Vec2 ){ 0.0f, 1.0f }, FLT_MAX );
	b2PlaneSolverResult result = b2SolvePlanes( ( b2Vec2 ){ 0.0f, -1.0f }, &plane, 1 );

	// An overshoot would eject the mover out the far side of the floor
	ORACLE_F32( "mover:43/value0", result.delta.y );
	ORACLE_BOOL( "mover:43", result.delta.y <= 0.0f );
	ENSURE( result.delta.y <= 0.0f );

	// The mover comes to rest a slop under the surface, not through it
	ORACLE_F32( "mover:46/value0", result.delta.y );
	ORACLE_BOOL( "mover:46", result.delta.y > -10.0f * B2_LINEAR_SLOP );
	ENSURE( result.delta.y > -10.0f * B2_LINEAR_SLOP );
	ORACLE_F32( "mover:47/value0", result.delta.x );
	ORACLE_BOOL( "mover:47", result.delta.x == 0.0f );
	ENSURE( result.delta.x == 0.0f );
	ORACLE_F32( "mover:48/value0", plane.push );
	ORACLE_BOOL( "mover:48", plane.push > 0.0f );
	ENSURE( plane.push > 0.0f );

	ORACLE_CLOSE();
	return 0;
}

static int PushLimitTest( void )
{
	ORACLE_OPEN( "mover/PushLimitTest" );
	b2CollisionPlane plane = MakePlane( ( b2Vec2 ){ 0.0f, 1.0f }, 0.1f );
	b2PlaneSolverResult result = b2SolvePlanes( ( b2Vec2 ){ 0.0f, -1.0f }, &plane, 1 );

	// Soft collision, so the mover sinks once the limit is spent
	ORACLE_F32( "mover:61/value0", plane.push );
	ORACLE_F32( "mover:61", plane.push - 0.1f );
	ENSURE_SMALL( plane.push - 0.1f, FLT_EPSILON );
	ORACLE_F32( "mover:62/value0", result.delta.y );
	ORACLE_F32( "mover:62", result.delta.y + 0.9f );
	ENSURE_SMALL( result.delta.y + 0.9f, FLT_EPSILON );

	ORACLE_CLOSE();
	return 0;
}

static int CornerTest( void )
{
	ORACLE_OPEN( "mover/CornerTest" );
	b2CollisionPlane planes[2] = {
		MakePlane( ( b2Vec2 ){ 0.0f, 1.0f }, FLT_MAX ),
		MakePlane( ( b2Vec2 ){ 1.0f, 0.0f }, FLT_MAX ),
	};

	b2PlaneSolverResult result = b2SolvePlanes( ( b2Vec2 ){ -1.0f, -1.0f }, planes, ARRAY_COUNT( planes ) );

	// Both planes hold at once, rather than the second relaxing the first
	for ( int i = 0; i < ARRAY_COUNT( planes ); ++i )
	{
		float separation = b2PlaneSeparation( planes[i].plane, result.delta );
		ORACLE_F32( "mover:82", separation + B2_LINEAR_SLOP );
		ENSURE_SMALL( separation + B2_LINEAR_SLOP, B2_LINEAR_SLOP );
		ORACLE_BOOL( "mover:83", planes[i].push > 0.0f );
		ENSURE( planes[i].push > 0.0f );
	}

	ORACLE_CLOSE();
	return 0;
}

static int WedgeTest( void )
{
	ORACLE_OPEN( "mover/WedgeTest" );
	// An acute wedge takes several passes to settle, unlike a square corner
	b2CollisionPlane planes[2] = {
		MakePlane( ( b2Vec2 ){ 0.0f, 1.0f }, FLT_MAX ),
		MakePlane( ( b2Vec2 ){ 0.8f, 0.6f }, FLT_MAX ),
	};

	b2PlaneSolverResult result = b2SolvePlanes( ( b2Vec2 ){ -1.0f, -1.0f }, planes, ARRAY_COUNT( planes ) );

	// Settling means resting on both faces. Stopping early leaves the mover shoved
	// clear of one of them, which reads as a large positive separation
	for ( int i = 0; i < ARRAY_COUNT( planes ); ++i )
	{
		float separation = b2PlaneSeparation( planes[i].plane, result.delta );
		ORACLE_F32( "mover:106", separation + B2_LINEAR_SLOP );
		ENSURE_SMALL( separation + B2_LINEAR_SLOP, B2_LINEAR_SLOP );
	}

	ORACLE_CLOSE();
	return 0;
}

static int ClipVectorTest( void )
{
	ORACLE_OPEN( "mover/ClipVectorTest" );
	b2CollisionPlane plane = MakePlane( ( b2Vec2 ){ 0.0f, 1.0f }, FLT_MAX );
	plane.push = 0.5f;

	b2Vec2 v = b2ClipVector( ( b2Vec2 ){ 1.0f, -1.0f }, &plane, 1 );
	ORACLE_F32( "mover:120/value0", v.x );
	ORACLE_F32( "mover:120", v.x - 1.0f );
	ENSURE_SMALL( v.x - 1.0f, FLT_EPSILON );
	ORACLE_F32( "mover:121/value0", v.y );
	ORACLE_F32( "mover:121", v.y );
	ENSURE_SMALL( v.y, FLT_EPSILON );

	// Separating velocity passes through untouched
	v = b2ClipVector( ( b2Vec2 ){ 1.0f, 1.0f }, &plane, 1 );
	ORACLE_F32( "mover:125/value0", v.x );
	ORACLE_F32( "mover:125", v.x - 1.0f );
	ENSURE_SMALL( v.x - 1.0f, FLT_EPSILON );
	ORACLE_F32( "mover:126/value0", v.y );
	ORACLE_F32( "mover:126", v.y - 1.0f );
	ENSURE_SMALL( v.y - 1.0f, FLT_EPSILON );

	plane.clipVelocity = false;
	v = b2ClipVector( ( b2Vec2 ){ 1.0f, -1.0f }, &plane, 1 );
	ORACLE_F32( "mover:130/value0", v.y );
	ORACLE_F32( "mover:130", v.y + 1.0f );
	ENSURE_SMALL( v.y + 1.0f, FLT_EPSILON );

	// Zero push means the solver never touched this plane
	plane.clipVelocity = true;
	plane.push = 0.0f;
	v = b2ClipVector( ( b2Vec2 ){ 1.0f, -1.0f }, &plane, 1 );
	ORACLE_F32( "mover:136/value0", v.y );
	ORACLE_F32( "mover:136", v.y + 1.0f );
	ENSURE_SMALL( v.y + 1.0f, FLT_EPSILON );

	ORACLE_CLOSE();
	return 0;
}

int MoverTest( void )
{
	RUN_SUBTEST( NoPlanesTest );
	RUN_SUBTEST( SolvePlaneTest );
	RUN_SUBTEST( PushLimitTest );
	RUN_SUBTEST( CornerTest );
	RUN_SUBTEST( WedgeTest );
	RUN_SUBTEST( ClipVectorTest );

	return 0;
}
