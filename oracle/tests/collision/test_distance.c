// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "test_macros.h"

#include "box2d/collision.h"
#include "box2d/constants.h"
#include "box2d/math_functions.h"

#include <float.h>

static void collisionObserveDistance( b2DistanceOutput output, const b2SimplexCache* cache )
{
	ORACLE_F32( "full/distance", output.distance );
	ORACLE_VEC2( "full/point-a", output.pointA );
	ORACLE_VEC2( "full/point-b", output.pointB );
	if ( output.distance > 0 ) ORACLE_VEC2( "full/normal", output.normal );
	ORACLE_I32( "full/iterations", output.iterations );
	ORACLE_I32( "full/simplex-count", output.simplexCount );
	ORACLE_U32( "full/cache/count", cache->count );
	for ( int i = 0; i < cache->count; ++i )
	{
		ORACLE_U32( ORACLE_KEY( "full/cache/%d/a", i ), cache->indexA[i] );
		ORACLE_U32( ORACLE_KEY( "full/cache/%d/b", i ), cache->indexB[i] );
	}
}

static void collisionObserveCast( b2CastOutput output )
{
	ORACLE_BOOL( "full/cast/hit", output.hit );
	ORACLE_I32( "full/cast/iterations", output.iterations );
	if ( output.hit )
	{
		ORACLE_F32( "full/cast/fraction", output.fraction );
		ORACLE_VEC2( "full/cast/normal", output.normal );
		ORACLE_VEC2( "full/cast/point", output.point );
	}
}

static int SegmentDistanceTest( void )
{
	ORACLE_OPEN( "distance/SegmentDistanceTest" );
	b2Vec2 p1 = { -1.0f, -1.0f };
	b2Vec2 q1 = { -1.0f, 1.0f };
	b2Vec2 p2 = { 2.0f, 0.0f };
	b2Vec2 q2 = { 1.0f, 0.0f };

	b2SegmentDistanceResult result = b2SegmentDistance( p1, q1, p2, q2 );

	ORACLE_F32( "distance:21/value0", result.fraction1 );
	ORACLE_F32( "distance:21", result.fraction1 - 0.5f );
	ENSURE_SMALL( result.fraction1 - 0.5f, FLT_EPSILON );
	ORACLE_F32( "distance:23/value0", result.fraction2 );
	ORACLE_F32( "distance:23", result.fraction2 - 1.0f );
	ENSURE_SMALL( result.fraction2 - 1.0f, FLT_EPSILON );
	ORACLE_F32( "distance:24/value0", result.closest1.x );
	ORACLE_F32( "distance:24", result.closest1.x + 1.0f );
	ENSURE_SMALL( result.closest1.x + 1.0f, FLT_EPSILON );
	ORACLE_F32( "distance:25/value0", result.closest1.y );
	ORACLE_F32( "distance:25", result.closest1.y );
	ENSURE_SMALL( result.closest1.y, FLT_EPSILON );
	ORACLE_F32( "distance:26/value0", result.closest2.x );
	ORACLE_F32( "distance:26", result.closest2.x - 1.0f );
	ENSURE_SMALL( result.closest2.x - 1.0f, FLT_EPSILON );
	ORACLE_F32( "distance:27/value0", result.closest2.y );
	ORACLE_F32( "distance:27", result.closest2.y );
	ENSURE_SMALL( result.closest2.y, FLT_EPSILON );
	ORACLE_F32( "distance:28/value0", result.distanceSquared );
	ORACLE_F32( "distance:28", result.distanceSquared - 4.0f );
	ENSURE_SMALL( result.distanceSquared - 4.0f, FLT_EPSILON );

	ORACLE_CLOSE();
	return 0;
}

static int ShapeDistanceTest( void )
{
	ORACLE_OPEN( "distance/ShapeDistanceTest" );
	b2Vec2 vas[] = { ( b2Vec2 ){ -1.0f, -1.0f }, ( b2Vec2 ){ 1.0f, -1.0f }, ( b2Vec2 ){ 1.0f, 1.0f }, ( b2Vec2 ){ -1.0f, 1.0f } };

	b2Vec2 vbs[] = {
		( b2Vec2 ){ 2.0f, -1.0f },
		( b2Vec2 ){ 2.0f, 1.0f },
	};

	b2DistanceInput input;
	input.proxyA = b2MakeProxy( vas, ARRAY_COUNT( vas ), 0.0f );
	input.proxyB = b2MakeProxy( vbs, ARRAY_COUNT( vbs ), 0.0f );
	input.transform = b2Transform_identity;
	input.useRadii = false;

	b2SimplexCache cache = { 0 };
	b2DistanceOutput output = b2ShapeDistance(&input,  &cache, NULL, 0 );
	collisionObserveDistance( output, &cache );

	ORACLE_F32( "distance:52/value0", output.distance );
	ORACLE_F32( "distance:52", output.distance - 1.0f );
	ENSURE_SMALL( output.distance - 1.0f, FLT_EPSILON );

	ORACLE_CLOSE();
	return 0;
}

static int ShapeCastTest( void )
{
	ORACLE_OPEN( "distance/ShapeCastTest" );
	b2Vec2 vas[] = { ( b2Vec2 ){ -1.0f, -1.0f }, ( b2Vec2 ){ 1.0f, -1.0f }, ( b2Vec2 ){ 1.0f, 1.0f }, ( b2Vec2 ){ -1.0f, 1.0f } };

	b2Vec2 vbs[] = {
		( b2Vec2 ){ 2.0f, -1.0f },
		( b2Vec2 ){ 2.0f, 1.0f },
	};

	b2ShapeCastPairInput input = { 0 };
	input.proxyA = b2MakeProxy( vas, ARRAY_COUNT( vas ), 0.0f );
	input.proxyB = b2MakeProxy( vbs, ARRAY_COUNT( vbs ), 0.0f );
	input.transform = b2Transform_identity;
	input.translationB = ( b2Vec2 ){ -2.0f, 0.0f };
	input.maxFraction = 1.0f;
	input.canEncroach = false;

	b2CastOutput output = b2ShapeCast( &input );
	collisionObserveCast( output );

	ORACLE_BOOL( "distance:78/value0", output.hit );
	ORACLE_BOOL( "distance:78", output.hit );
	ENSURE( output.hit );
	ORACLE_F32( "distance:80/value0", output.fraction );
	ORACLE_F32( "distance:80", output.fraction - 0.5f );
	ENSURE_SMALL( output.fraction - 0.5f, 0.005f );

	ORACLE_CLOSE();
	return 0;
}

// A thin plank and a point a linear slop off the middle of its short end face. The closest feature
// is that end face, so GJK finishes on a two point simplex spanning it.
//
// b2SolveSimplex2 hands back a search direction of magnitude 2 * distance * edge^2, which the
// caller weighs against an absolute epsilon. That test is cubic in length, so it gives out on
// short edges: below an edge of about 3.5 mm at default length units it declares the shapes
// overlapped and returns a zero distance with no normal, though they are plainly apart.
//
// Measured cutoff matches 2 * distance * edge^2 < FLT_EPSILON to three digits, and it tracks the
// contact edge alone. Widening the plank from 1 m to 100 m changes nothing.
static int ShapeDistanceShortEdgeTest( void )
{
	ORACLE_OPEN( "distance/ShapeDistanceShortEdgeTest" );
	float halfThickness = 0.0016f;

	b2Vec2 corners[] = {
		( b2Vec2 ){ -0.5f, -halfThickness },
		( b2Vec2 ){ 0.5f, -halfThickness },
		( b2Vec2 ){ 0.5f, halfThickness },
		( b2Vec2 ){ -0.5f, halfThickness },
	};

	b2Vec2 point = { 0.5f + B2_LINEAR_SLOP, 0.0f };

	b2DistanceInput input = { 0 };
	input.proxyA = b2MakeProxy( corners, ARRAY_COUNT( corners ), 0.0f );
	input.proxyB = b2MakeProxy( &point, 1, 0.0f );
	input.transform = b2Transform_identity;
	input.useRadii = false;

	b2SimplexCache cache = { 0 };
	b2DistanceOutput output = b2ShapeDistance( &input, &cache, NULL, 0 );
	collisionObserveDistance( output, &cache );

	ORACLE_F32( "distance:118/value0", output.distance );
	ORACLE_BOOL( "distance:118", output.distance > 0.0f );
	ENSURE( output.distance > 0.0f );
	ORACLE_VEC2( "distance:120/value0", output.normal );
	ORACLE_BOOL( "distance:120", b2IsNormalized( output.normal ) );
	ENSURE( b2IsNormalized( output.normal ) );
	ORACLE_F32( "distance:121/value0", output.distance );
	ORACLE_F32( "distance:121", output.distance - B2_LINEAR_SLOP );
	ENSURE_SMALL( output.distance - B2_LINEAR_SLOP, 1e-6f );

	ORACLE_CLOSE();
	return 0;
}

// The same false overlap seen through b2ShapeCast, which is how it was reported.
//
// Conservative advancement stops the cores a target apart and target is at least a linear slop, so
// the query can never legitimately answer overlap after the first iteration. Here iteration 1 does,
// and b2ShapeCast has no fallback: it trips
// B2_ASSERT( distanceOutput.distance > 0.0f && b2IsNormalized( distanceOutput.normal ) ).
//
// Enable once b2ShapeDistance stops reporting the false overlap. The assert aborts the run, so
// leaving it on would take the rest of the suite with it.
#define B2_SHORT_EDGE_CAST_REPRO 0
#if B2_SHORT_EDGE_CAST_REPRO
static int ShapeCastShortEdgeTest( void )
{
	ORACLE_OPEN( "distance/ShapeCastShortEdgeTest" );
	float halfThickness = 0.0002f;

	b2Vec2 corners[] = {
		( b2Vec2 ){ -0.5f, -halfThickness },
		( b2Vec2 ){ 0.5f, -halfThickness },
		( b2Vec2 ){ 0.5f, halfThickness },
		( b2Vec2 ){ -0.5f, halfThickness },
	};

	// falls past the end of the plank and drifts into it, so the second query lands on the end face
	b2Vec2 start = { 0.506f, 0.0035f };

	b2ShapeCastPairInput input = { 0 };
	input.proxyA = b2MakeProxy( corners, ARRAY_COUNT( corners ), 0.0f );
	input.proxyB = b2MakeProxy( &start, 1, 0.0f );
	input.transform = b2Transform_identity;
	input.translationB = ( b2Vec2 ){ -0.01f, -0.5f };
	input.maxFraction = 1.0f;
	input.canEncroach = false;

	b2CastOutput output = b2ShapeCast( &input );

	ORACLE_BOOL( "distance:162/value0", output.hit );
	ORACLE_VEC2( "distance:162/value1", output.normal );
	ORACLE_BOOL( "distance:162", output.hit == false || b2IsNormalized( output.normal ) );
	ENSURE( output.hit == false || b2IsNormalized( output.normal ) );

	ORACLE_CLOSE();
	return 0;
}
#endif

static int TimeOfImpactTest( void )
{
	ORACLE_OPEN( "distance/TimeOfImpactTest" );
	b2Vec2 vas[] = { { -1.0f, -1.0f }, { 1.0f, -1.0f }, { 1.0f, 1.0f }, { -1.0f, 1.0f } };

	b2Vec2 vbs[] = {
		{ 2.0f, -1.0f },
		{ 2.0f, 1.0f },
	};

	b2TOIInput input;
	input.proxyA = b2MakeProxy( vas, ARRAY_COUNT( vas ), 0.0f );
	input.proxyB = b2MakeProxy( vbs, ARRAY_COUNT( vbs ), 0.0f );
	input.sweepA = ( b2Sweep ){ b2Vec2_zero, b2Vec2_zero, b2Vec2_zero, b2Rot_identity, b2Rot_identity };
	input.sweepB = ( b2Sweep ){ b2Vec2_zero, b2Vec2_zero, ( b2Vec2 ){ -2.0f, 0.0f }, b2Rot_identity, b2Rot_identity };
	input.maxFraction = 1.0f;

	b2TOIOutput output = b2TimeOfImpact( &input );
	ORACLE_I32( "full/toi/state", output.state );
	ORACLE_F32( "full/toi/fraction", output.fraction );
	if ( output.state == b2_toiStateHit )
	{
		ORACLE_VEC2( "full/toi/point", output.point );
		ORACLE_VEC2( "full/toi/normal", output.normal );
	}

	ORACLE_I32( "distance:188/value0", output.state );
	ORACLE_BOOL( "distance:188", output.state == b2_toiStateHit );
	ENSURE( output.state == b2_toiStateHit );
	ORACLE_F32( "distance:190/value0", output.fraction );
	ORACLE_F32( "distance:190", output.fraction - 0.5f );
	ENSURE_SMALL( output.fraction - 0.5f, 0.005f );

	ORACLE_CLOSE();
	return 0;
}

int DistanceTest( void )
{
	RUN_SUBTEST( SegmentDistanceTest );
	RUN_SUBTEST( ShapeDistanceTest );
	RUN_SUBTEST( ShapeCastTest );
	RUN_SUBTEST( ShapeDistanceShortEdgeTest );
#if B2_SHORT_EDGE_CAST_REPRO
	RUN_SUBTEST( ShapeCastShortEdgeTest );
#endif
	RUN_SUBTEST( TimeOfImpactTest );

	return 0;
}
