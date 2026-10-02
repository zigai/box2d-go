// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "test_macros.h"

#include "box2d/math_functions.h"

#include <float.h>
#include <stdio.h>

// 0.0023 degrees
#define ATAN_TOL 0.00004f

int MathTest( void )
{
	ORACLE_OPEN( "math/MathTest" );
	for ( float t = -10.0f; t < 10.0f; t += 0.01f )
	{
		float angle = B2_PI * t;
		b2Rot r = b2MakeRot( angle );
		float c = cosf( angle );
		ORACLE_COLLISION_F32( 20, c );
		float s = sinf( angle );
		ORACLE_COLLISION_F32( 21, s );

		// The cosine and sine approximations are accurate to about 0.1 degrees (0.002 radians)
		// printf( "%g %g\n", r.c - c, r.s - s );
		ORACLE_COLLISION_F32( 25, r.c - c );
		ENSURE_SMALL( r.c - c, 0.002f );
		ORACLE_COLLISION_F32( 26, r.s - s );
		ENSURE_SMALL( r.s - s, 0.002f );

		float xn = b2UnwindAngle( angle );
		ORACLE_COLLISION_F32( 285, xn );
		ORACLE_COLLISION_BOOL( 29, -B2_PI <= xn && xn <= B2_PI );
		ENSURE( -B2_PI <= xn && xn <= B2_PI );

		float a = b2Atan2( s, c );
		ORACLE_COLLISION_F32( 288, a );
		ORACLE_COLLISION_BOOL( 32, b2IsValidFloat( a ) );
		ENSURE( b2IsValidFloat( a ) );

		float diff = b2AbsFloat( a - xn );

		// The two results can be off by 360 degrees (-pi and pi)
		if ( diff > B2_PI )
		{
			diff -= 2.0f * B2_PI;
		}

		// The approximate atan2 is quite accurate
		ORACLE_COLLISION_F32( 43, diff );
		ENSURE_SMALL( diff, ATAN_TOL );
	}

	for ( float y = -1.0f; y <= 1.0f; y += 0.01f )
	{
		for ( float x = -1.0f; x <= 1.0f; x += 0.01f )
		{
			float a1 = b2Atan2( y, x );
			float a2 = atan2f( y, x );
		ORACLE_COLLISION_F32( 51, a2 );
			float diff = b2AbsFloat( a1 - a2 );
			ORACLE_COLLISION_F32( 309, a1 );
			ORACLE_COLLISION_BOOL( 53, b2IsValidFloat( a1 ) );
			ENSURE( b2IsValidFloat( a1 ) );
			ORACLE_COLLISION_F32( 54, diff );
			ENSURE_SMALL( diff, ATAN_TOL );
		}
	}

	{
		float a1 = b2Atan2( 1.0f, 0.0f );
		float a2 = atan2f( 1.0f, 0.0f );
		ORACLE_COLLISION_F32( 60, a2 );
		float diff = b2AbsFloat( a1 - a2 );
		ORACLE_COLLISION_F32( 318, a1 );
		ORACLE_COLLISION_BOOL( 62, b2IsValidFloat( a1 ) );
		ENSURE( b2IsValidFloat( a1 ) );
		ORACLE_COLLISION_F32( 63, diff );
		ENSURE_SMALL( diff, ATAN_TOL );
	}

	{
		float a1 = b2Atan2( -1.0f, 0.0f );
		float a2 = atan2f( -1.0f, 0.0f );
		ORACLE_COLLISION_F32( 68, a2 );
		float diff = b2AbsFloat( a1 - a2 );
		ORACLE_COLLISION_F32( 326, a1 );
		ORACLE_COLLISION_BOOL( 70, b2IsValidFloat( a1 ) );
		ENSURE( b2IsValidFloat( a1 ) );
		ORACLE_COLLISION_F32( 71, diff );
		ENSURE_SMALL( diff, ATAN_TOL );
	}

	{
		float a1 = b2Atan2( 0.0f, 1.0f );
		float a2 = atan2f( 0.0f, 1.0f );
		ORACLE_COLLISION_F32( 76, a2 );
		float diff = b2AbsFloat( a1 - a2 );
		ORACLE_COLLISION_F32( 334, a1 );
		ORACLE_COLLISION_BOOL( 78, b2IsValidFloat( a1 ) );
		ENSURE( b2IsValidFloat( a1 ) );
		ORACLE_COLLISION_F32( 79, diff );
		ENSURE_SMALL( diff, ATAN_TOL );
	}

	{
		float a1 = b2Atan2( 0.0f, -1.0f );
		float a2 = atan2f( 0.0f, -1.0f );
		ORACLE_COLLISION_F32( 84, a2 );
		float diff = b2AbsFloat( a1 - a2 );
		ORACLE_COLLISION_F32( 342, a1 );
		ORACLE_COLLISION_BOOL( 86, b2IsValidFloat( a1 ) );
		ENSURE( b2IsValidFloat( a1 ) );
		ORACLE_COLLISION_F32( 87, diff );
		ENSURE_SMALL( diff, ATAN_TOL );
	}

	{
		float a1 = b2Atan2( 0.0f, 0.0f );
		float a2 = atan2f( 0.0f, 0.0f );
		ORACLE_COLLISION_F32( 92, a2 );
		float diff = b2AbsFloat( a1 - a2 );
		ORACLE_COLLISION_F32( 350, a1 );
		ORACLE_COLLISION_BOOL( 94, b2IsValidFloat( a1 ) );
		ENSURE( b2IsValidFloat( a1 ) );
		ORACLE_COLLISION_F32( 95, diff );
		ENSURE_SMALL( diff, ATAN_TOL );
	}

	b2Vec2 zero = b2Vec2_zero;
	b2Vec2 one = { 1.0f, 1.0f };
	b2Vec2 two = { 2.0f, 2.0f };

	b2Vec2 v = b2Add( one, two );
	ORACLE_COLLISION_F32( 359, v.x );
	ORACLE_COLLISION_F32( 615, v.y );
	ORACLE_COLLISION_BOOL( 103, v.x == 3.0f && v.y == 3.0f );
	ENSURE( v.x == 3.0f && v.y == 3.0f );

	v = b2Sub( zero, two );
	ORACLE_COLLISION_F32( 362, v.x );
	ORACLE_COLLISION_F32( 618, v.y );
	ORACLE_COLLISION_BOOL( 106, v.x == -2.0f && v.y == -2.0f );
	ENSURE( v.x == -2.0f && v.y == -2.0f );

	v = b2Add( two, two );
	ORACLE_COLLISION_F32( 365, v.x );
	ORACLE_COLLISION_F32( 621, v.y );
	ORACLE_COLLISION_BOOL( 109, v.x != 5.0f && v.y != 5.0f );
	ENSURE( v.x != 5.0f && v.y != 5.0f );

	b2Transform transform1 = { { -2.0f, 3.0f }, b2MakeRot( 1.0f ) };
	b2Transform transform2 = { { 1.0f, 0.0f }, b2MakeRot( -2.0f ) };

	b2Transform transform = b2MulTransforms( transform2, transform1 );

	v = b2TransformPoint( transform2, b2TransformPoint( transform1, two ) );

	b2Vec2 u = b2TransformPoint( transform, two );

	ORACLE_COLLISION_F32( 120, u.x - v.x );
	ENSURE_SMALL( u.x - v.x, 10.0f * FLT_EPSILON );
	ORACLE_COLLISION_F32( 121, u.y - v.y );
	ENSURE_SMALL( u.y - v.y, 10.0f * FLT_EPSILON );

	v = b2TransformPoint( transform1, two );
	v = b2InvTransformPoint( transform1, v );

	ORACLE_COLLISION_F32( 126, v.x - two.x );
	ENSURE_SMALL( v.x - two.x, 8.0f * FLT_EPSILON );
	ORACLE_COLLISION_F32( 127, v.y - two.y );
	ENSURE_SMALL( v.y - two.y, 8.0f * FLT_EPSILON );

	v = b2Normalize(( b2Vec2 ){ 0.2f, -0.5f });
	for ( float y = -1.0f; y <= 1.0f; y += 0.01f )
	{
		for ( float x = -1.0f; x <= 1.0f; x += 0.01f )
		{
			if (x == 0.0f && y == 0.0f)
			{
				continue;
			}

			u = b2Normalize( ( b2Vec2 ){ x, y } );

			b2Rot r = b2ComputeRotationBetweenUnitVectors( v, u );

			b2Vec2 w = b2RotateVector( r, v );
			ORACLE_COLLISION_F32( 144, w.x - u.x );
			ENSURE_SMALL( w.x - u.x, 4.0f * FLT_EPSILON );
			ORACLE_COLLISION_F32( 145, w.y - u.y );
			ENSURE_SMALL( w.y - u.y, 4.0f * FLT_EPSILON );
		}
	}

	// NLerp of b2Rot has an error of over 4 degrees.
	// 2D quaternions should have an error under 1 degree.
	b2Rot q1 = b2Rot_identity;
	b2Rot q2 = b2MakeRot(0.5f * B2_PI);
	int n = 100;
	for (int i = 0; i <= n; ++i)
	{
		float alpha = (float)i / (float)n;
		b2Rot q = b2NLerp(q1, q2, alpha);
		float angle = b2Rot_GetAngle(q);
		ORACLE_COLLISION_F32( 159, alpha * 0.5f * B2_PI - angle );
		ENSURE_SMALL( alpha * 0.5f * B2_PI - angle, 5.0f * B2_PI / 180.0f );
		//printf("angle = [%g %g %g]\n", alpha, alpha * 0.5f * B2_PI, angle);
	}
	
	// Test relative angle
	float baseAngle = 0.75f * B2_PI;
	q1 = b2MakeRot(baseAngle);
	for ( float t = -10.0f; t < 10.0f; t += 0.01f )
	{
		float angle = B2_PI * t;
		q2 = b2MakeRot(angle);

		float relativeAngle = b2RelativeAngle( q1, q2 );
		float unwoundAngle = b2UnwindAngle( angle - baseAngle );
		float tolerance = 0.1f * B2_PI / 180.0f;
		ORACLE_COLLISION_F32( 174, relativeAngle - unwoundAngle );
		ENSURE_SMALL( relativeAngle - unwoundAngle, tolerance );
	}

	// World position boundary helpers. With large world mode off these collapse to the float
	// ops, so the round trips hold in both builds.
	{
		b2Vec2 d = { 0.25f, -0.5f };
		b2Pos base = (b2Pos){ 10.0f, -20.0f };
		b2Pos p = b2OffsetPos( base, d );
		b2Vec2 back = b2SubPos( p, base );
		ORACLE_COLLISION_F32( 184, back.x - d.x );
		ENSURE_SMALL( back.x - d.x, 8.0f * FLT_EPSILON );
		ORACLE_COLLISION_F32( 185, back.y - d.y );
		ENSURE_SMALL( back.y - d.y, 8.0f * FLT_EPSILON );

		b2Vec2 r = b2ToVec2( base );
		ORACLE_COLLISION_F32( 444, r.x );
		ORACLE_COLLISION_F32( 700, r.y );
		ORACLE_COLLISION_BOOL( 188, r.x == 10.0f && r.y == -20.0f );
		ENSURE( r.x == 10.0f && r.y == -20.0f );

		ORACLE_COLLISION_BOOL( 190, b2IsValidPosition( p ) );
		ENSURE( b2IsValidPosition( p ) );
		ORACLE_COLLISION_BOOL( 191, b2IsValidPosition( b2Pos_zero ) );
		ENSURE( b2IsValidPosition( b2Pos_zero ) );
		ORACLE_COLLISION_BOOL( 192, b2IsValidWorldTransform( b2WorldTransform_identity ) );
		ENSURE( b2IsValidWorldTransform( b2WorldTransform_identity ) );

		b2WorldTransform wt = { (b2Pos){ 3.0f, -4.0f }, b2MakeRot( 0.7f ) };
		ORACLE_COLLISION_BOOL( 195, b2IsValidWorldTransform( wt ) );
		ENSURE( b2IsValidWorldTransform( wt ) );

		// Local to world to local round trip
		b2Vec2 local = { 1.5f, 2.5f };
		b2Pos world = b2TransformWorldPoint( wt, local );
		b2Vec2 backLocal = b2InvTransformWorldPoint( wt, world );
		ORACLE_COLLISION_F32( 201, backLocal.x - local.x );
		ENSURE_SMALL( backLocal.x - local.x, 8.0f * FLT_EPSILON );
		ORACLE_COLLISION_F32( 202, backLocal.y - local.y );
		ENSURE_SMALL( backLocal.y - local.y, 8.0f * FLT_EPSILON );

		// Relative transform of B in A matches a float reference at modest coordinates
		b2WorldTransform A = { (b2Pos){ -2.0f, 3.0f }, b2MakeRot( 1.0f ) };
		b2WorldTransform B = { (b2Pos){ 1.0f, 0.0f }, b2MakeRot( -2.0f ) };
		b2Transform rel = b2InvMulWorldTransforms( A, B );
		b2Transform refA = { b2ToVec2( A.p ), A.q };
		b2Transform refB = { b2ToVec2( B.p ), B.q };
		b2Transform ref = b2InvMulTransforms( refA, refB );
		ORACLE_COLLISION_F32( 211, rel.p.x - ref.p.x );
		ENSURE_SMALL( rel.p.x - ref.p.x, 8.0f * FLT_EPSILON );
		ORACLE_COLLISION_F32( 212, rel.p.y - ref.p.y );
		ENSURE_SMALL( rel.p.y - ref.p.y, 8.0f * FLT_EPSILON );
	}

#if defined( BOX2D_DOUBLE_PRECISION )
	// Far from the origin a float vector cannot resolve sub meter motion, but a double world
	// position can. This is the whole point of large world mode.
	{
		b2Vec2 d = { 0.25f, -0.5f };
		b2Pos base = (b2Pos){ 1.0e7f, 0.0f };
		b2Pos p = b2OffsetPos( base, d );
		b2Vec2 back = b2SubPos( p, base );
		ORACLE_COLLISION_F32( 223, back.x - d.x );
		ENSURE_SMALL( back.x - d.x, 1.0e-4f );
		ORACLE_COLLISION_F32( 224, back.y - d.y );
		ENSURE_SMALL( back.y - d.y, 1.0e-4f );
	}
#endif

	ORACLE_COLLISION_CLOSE();
	return 0;
}
