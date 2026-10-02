// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

#include "aabb.h"
#include "shape.h"
#include "test_macros.h"

#include "box2d/collision.h"
#include "box2d/math_functions.h"

static int AABBTest( void )
{
	ORACLE_OPEN( "collision/AABBTest" );
	b2AABB a;
	a.lowerBound = (b2Vec2){ -1.0f, -1.0f };
	a.upperBound = (b2Vec2){ -2.0f, -2.0f };

	ORACLE_BOOL( "collision:17", b2IsValidAABB( a ) == false );
	ENSURE( b2IsValidAABB( a ) == false );

	a.upperBound = (b2Vec2){ 1.0f, 1.0f };
	ORACLE_BOOL( "collision:21", b2IsValidAABB( a ) == true );
	ENSURE( b2IsValidAABB( a ) == true );

	b2AABB b = { { 2.0f, 2.0f }, { 4.0f, 4.0f } };
	ORACLE_BOOL( "collision:24", b2AABB_Overlaps( a, b ) == false );
	ENSURE( b2AABB_Overlaps( a, b ) == false );
	ORACLE_BOOL( "collision:25", b2AABB_Contains( a, b ) == false );
	ENSURE( b2AABB_Contains( a, b ) == false );

	ORACLE_CLOSE();
	return 0;
}

static int AABBRayCastTest( void )
{
	ORACLE_OPEN( "collision/AABBRayCastTest" );
	// Test AABB centered at origin with bounds [-1, -1] to [1, 1]
	b2AABB aabb = { { -1.0f, -1.0f }, { 1.0f, 1.0f } };

	// Test 1: Ray hits AABB from left side
	{
		b2Vec2 p1 = { -3.0f, 0.0f };
		b2Vec2 p2 = { 3.0f, 0.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:42/value0", output.hit );
		ORACLE_BOOL( "collision:42", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:44/value0", output.fraction );
		ORACLE_F32( "collision:44", output.fraction - 1.0f / 3.0f );
		ENSURE_SMALL( output.fraction - 1.0f / 3.0f, FLT_EPSILON );
		ORACLE_F32( "collision:45/value0", output.normal.x );
		ORACLE_F32( "collision:45", output.normal.x + 1.0f );
		ENSURE_SMALL( output.normal.x + 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:46/value0", output.normal.y );
		ORACLE_F32( "collision:46", output.normal.y );
		ENSURE_SMALL( output.normal.y, FLT_EPSILON );
		ORACLE_F32( "collision:47/value0", output.point.x );
		ORACLE_F32( "collision:47", output.point.x + 1.0f );
		ENSURE_SMALL( output.point.x + 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:48/value0", output.point.y );
		ORACLE_F32( "collision:48", output.point.y );
		ENSURE_SMALL( output.point.y, FLT_EPSILON );
	}

	// Test 2: Ray hits AABB from right side
	{
		b2Vec2 p1 = { 3.0f, 0.0f };
		b2Vec2 p2 = { -3.0f, 0.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:56/value0", output.hit );
		ORACLE_BOOL( "collision:56", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:58/value0", output.fraction );
		ORACLE_F32( "collision:58", output.fraction - 1.0f / 3.0f );
		ENSURE_SMALL( output.fraction - 1.0f / 3.0f, FLT_EPSILON );
		ORACLE_F32( "collision:59/value0", output.normal.x );
		ORACLE_F32( "collision:59", output.normal.x - 1.0f );
		ENSURE_SMALL( output.normal.x - 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:60/value0", output.normal.y );
		ORACLE_F32( "collision:60", output.normal.y );
		ENSURE_SMALL( output.normal.y, FLT_EPSILON );
		ORACLE_F32( "collision:61/value0", output.point.x );
		ORACLE_F32( "collision:61", output.point.x - 1.0f );
		ENSURE_SMALL( output.point.x - 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:62/value0", output.point.y );
		ORACLE_F32( "collision:62", output.point.y );
		ENSURE_SMALL( output.point.y, FLT_EPSILON );
	}

	// Test 3: Ray hits AABB from bottom
	{
		b2Vec2 p1 = { 0.0f, -3.0f };
		b2Vec2 p2 = { 0.0f, 3.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:70/value0", output.hit );
		ORACLE_BOOL( "collision:70", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:72/value0", output.fraction );
		ORACLE_F32( "collision:72", output.fraction - 1.0f / 3.0f );
		ENSURE_SMALL( output.fraction - 1.0f / 3.0f, FLT_EPSILON );
		ORACLE_F32( "collision:73/value0", output.normal.x );
		ORACLE_F32( "collision:73", output.normal.x );
		ENSURE_SMALL( output.normal.x, FLT_EPSILON );
		ORACLE_F32( "collision:74/value0", output.normal.y );
		ORACLE_F32( "collision:74", output.normal.y + 1.0f );
		ENSURE_SMALL( output.normal.y + 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:75/value0", output.point.x );
		ORACLE_F32( "collision:75", output.point.x );
		ENSURE_SMALL( output.point.x, FLT_EPSILON );
		ORACLE_F32( "collision:76/value0", output.point.y );
		ORACLE_F32( "collision:76", output.point.y + 1.0f );
		ENSURE_SMALL( output.point.y + 1.0f, FLT_EPSILON );
	}

	// Test 4: Ray hits AABB from top
	{
		b2Vec2 p1 = { 0.0f, 3.0f };
		b2Vec2 p2 = { 0.0f, -3.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:84/value0", output.hit );
		ORACLE_BOOL( "collision:84", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:86/value0", output.fraction );
		ORACLE_F32( "collision:86", output.fraction - 1.0f / 3.0f );
		ENSURE_SMALL( output.fraction - 1.0f / 3.0f, FLT_EPSILON );
		ORACLE_F32( "collision:87/value0", output.normal.x );
		ORACLE_F32( "collision:87", output.normal.x );
		ENSURE_SMALL( output.normal.x, FLT_EPSILON );
		ORACLE_F32( "collision:88/value0", output.normal.y );
		ORACLE_F32( "collision:88", output.normal.y - 1.0f );
		ENSURE_SMALL( output.normal.y - 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:89/value0", output.point.x );
		ORACLE_F32( "collision:89", output.point.x );
		ENSURE_SMALL( output.point.x, FLT_EPSILON );
		ORACLE_F32( "collision:90/value0", output.point.y );
		ORACLE_F32( "collision:90", output.point.y - 1.0f );
		ENSURE_SMALL( output.point.y - 1.0f, FLT_EPSILON );
	}

	// Test 5: Ray misses AABB completely (parallel to x-axis)
	{
		b2Vec2 p1 = { -3.0f, 2.0f };
		b2Vec2 p2 = { 3.0f, 2.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:98/value0", output.hit );
		ORACLE_BOOL( "collision:98", output.hit == false );
		ENSURE( output.hit == false );
	}

	// Test 6: Ray misses AABB completely (parallel to y-axis)
	{
		b2Vec2 p1 = { 2.0f, -3.0f };
		b2Vec2 p2 = { 2.0f, 3.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:107/value0", output.hit );
		ORACLE_BOOL( "collision:107", output.hit == false );
		ENSURE( output.hit == false );
	}

	// Test 7: Ray starts inside AABB
	{
		b2Vec2 p1 = { 0.0f, 0.0f };
		b2Vec2 p2 = { 2.0f, 0.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:116/value0", output.hit );
		ORACLE_BOOL( "collision:116", output.hit == false );
		ENSURE( output.hit == false );
	}

	// Test 8: Ray hits corner of AABB (diagonal ray)
	{
		b2Vec2 p1 = { -2.0f, -2.0f };
		b2Vec2 p2 = { 2.0f, 2.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:125/value0", output.hit );
		ORACLE_BOOL( "collision:125", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:127/value0", output.fraction );
		ORACLE_F32( "collision:127", output.fraction - 0.25f );
		ENSURE_SMALL( output.fraction - 0.25f, FLT_EPSILON );
		// Normal should be either (-1, 0) or (0, -1) depending on which edge is hit first
		ORACLE_F32( "collision:129/value0", output.normal.x );
		ORACLE_F32( "collision:129/value1", output.normal.y );
		ORACLE_BOOL( "collision:129", ( output.normal.x == -1.0f && output.normal.y == 0.0f ) || ( output.normal.x == 0.0f && output.normal.y == -1.0f ) );
		ENSURE( ( output.normal.x == -1.0f && output.normal.y == 0.0f ) ||
				( output.normal.x == 0.0f && output.normal.y == -1.0f ) );
	}

	// Test 9: Ray parallel to AABB edge but outside
	{
		b2Vec2 p1 = { -2.0f, 1.5f };
		b2Vec2 p2 = { 2.0f, 1.5f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:138/value0", output.hit );
		ORACLE_BOOL( "collision:138", output.hit == false );
		ENSURE( output.hit == false );
	}

	// Test 10: Ray parallel to AABB edge and exactly on boundary
	{
		b2Vec2 p1 = { -2.0f, 1.0f };
		b2Vec2 p2 = { 2.0f, 1.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:147/value0", output.hit );
		ORACLE_BOOL( "collision:147", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:149/value0", output.fraction );
		ORACLE_F32( "collision:149", output.fraction - 0.25f );
		ENSURE_SMALL( output.fraction - 0.25f, FLT_EPSILON );
		ORACLE_F32( "collision:150/value0", output.normal.x );
		ORACLE_F32( "collision:150", output.normal.x + 1.0f );
		ENSURE_SMALL( output.normal.x + 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:151/value0", output.normal.y );
		ORACLE_F32( "collision:151", output.normal.y );
		ENSURE_SMALL( output.normal.y, FLT_EPSILON );
	}

	// Test 11: Very short ray that doesn't reach AABB
	{
		b2Vec2 p1 = { -3.0f, 0.0f };
		b2Vec2 p2 = { -2.5f, 0.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:159/value0", output.hit );
		ORACLE_BOOL( "collision:159", output.hit == false );
		ENSURE( output.hit == false );
	}

	// Test 12: Zero-length ray (degenerate case)
	{
		b2Vec2 p1 = { 0.0f, 0.0f };
		b2Vec2 p2 = { 0.0f, 0.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:168/value0", output.hit );
		ORACLE_BOOL( "collision:168", output.hit == false );
		ENSURE( output.hit == false );
	}

	// Test 13: Ray hits AABB at exact boundary condition (t = 1.0)
	{
		b2Vec2 p1 = { -2.0f, 0.0f };
		b2Vec2 p2 = { -1.0f, 0.0f };
		b2CastOutput output = b2AABB_RayCast( aabb, p1, p2 );

		ORACLE_BOOL( "collision:177/value0", output.hit );
		ORACLE_BOOL( "collision:177", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:179/value0", output.fraction );
		ORACLE_F32( "collision:179", output.fraction - 1.0f );
		ENSURE_SMALL( output.fraction - 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:180/value0", output.normal.x );
		ORACLE_F32( "collision:180", output.normal.x + 1.0f );
		ENSURE_SMALL( output.normal.x + 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:181/value0", output.normal.y );
		ORACLE_F32( "collision:181", output.normal.y );
		ENSURE_SMALL( output.normal.y, FLT_EPSILON );
	}

	// Test 14: Different AABB position (not centered at origin)
	{
		b2AABB offsetAABB = { { 2.0f, 3.0f }, { 4.0f, 5.0f } };
		b2Vec2 p1 = { 0.0f, 4.0f };
		b2Vec2 p2 = { 6.0f, 4.0f };
		b2CastOutput output = b2AABB_RayCast( offsetAABB, p1, p2 );

		ORACLE_BOOL( "collision:190/value0", output.hit );
		ORACLE_BOOL( "collision:190", output.hit == true );
		ENSURE( output.hit == true );
		ORACLE_F32( "collision:192/value0", output.fraction );
		ORACLE_F32( "collision:192", output.fraction - 1.0f / 3.0f );
		ENSURE_SMALL( output.fraction - 1.0f / 3.0f, FLT_EPSILON );
		ORACLE_F32( "collision:193/value0", output.normal.x );
		ORACLE_F32( "collision:193", output.normal.x + 1.0f );
		ENSURE_SMALL( output.normal.x + 1.0f, FLT_EPSILON );
		ORACLE_F32( "collision:194/value0", output.normal.y );
		ORACLE_F32( "collision:194", output.normal.y );
		ENSURE_SMALL( output.normal.y, FLT_EPSILON );
		ORACLE_F32( "collision:195/value0", output.point.x );
		ORACLE_F32( "collision:195", output.point.x - 2.0f );
		ENSURE_SMALL( output.point.x - 2.0f, FLT_EPSILON );
		ORACLE_F32( "collision:196/value0", output.point.y );
		ORACLE_F32( "collision:196", output.point.y - 4.0f );
		ENSURE_SMALL( output.point.y - 4.0f, FLT_EPSILON );
	}

	ORACLE_CLOSE();
	return 0;
}

// The narrow phase differences the two world positions in double then works in frame A, so a
// manifold far from the origin must match the same manifold at the origin. Float loses this past
// ~1e7 m where the ULP grows larger than the overlap, which is the whole point of large world mode.
static int LargeWorldManifoldTest( void )
{
	ORACLE_OPEN( "collision/LargeWorldManifoldTest" );
	b2Polygon boxA = b2MakeBox( 0.5f, 0.5f );
	b2Polygon boxB = b2MakeBox( 0.5f, 0.5f );

	// Centers 0.9 apart so the boxes overlap by 0.1 along x
	b2Vec2 sep = { 0.9f, 0.0f };

	b2WorldTransform xfAo = b2WorldTransform_identity;
	b2WorldTransform xfBo = { b2OffsetPos( b2Pos_zero, sep ), b2Rot_identity };
	b2LocalManifold mOrigin = b2CollidePolygons( &boxA, &boxB, b2InvMulWorldTransforms( xfAo, xfBo ) );

	ORACLE_I32( "collision:218/value0", mOrigin.pointCount );
	ORACLE_BOOL( "collision:218", mOrigin.pointCount == 2 );
	ENSURE( mOrigin.pointCount == 2 );
	ORACLE_F32( "collision:220/value0", mOrigin.points[0].separation );
	ORACLE_F32( "collision:220", mOrigin.points[0].separation + 0.1f );
	ENSURE_SMALL( mOrigin.points[0].separation + 0.1f, 0.01f );
	ORACLE_F32( "collision:221/value0", mOrigin.points[1].separation );
	ORACLE_F32( "collision:221", mOrigin.points[1].separation + 0.1f );
	ENSURE_SMALL( mOrigin.points[1].separation + 0.1f, 0.01f );

#if defined( BOX2D_DOUBLE_PRECISION )
	// Same relative configuration shifted far from the origin. The relative pose differences the
	// world positions in double, so in double the frame A manifold is preserved to float precision.
	// In float it would collapse since the offset is below the ULP.
	b2Pos base = b2OffsetPos( b2Pos_zero, ( b2Vec2 ){ 1.0e7f, 1.0e7f } );
	b2WorldTransform xfAl = { base, b2Rot_identity };
	b2WorldTransform xfBl = { b2OffsetPos( base, sep ), b2Rot_identity };
	b2LocalManifold mLarge = b2CollidePolygons( &boxA, &boxB, b2InvMulWorldTransforms( xfAl, xfBl ) );

	ORACLE_I32( "collision:231/value0", mLarge.pointCount );
	ORACLE_I32( "collision:231/value1", mOrigin.pointCount );
	ORACLE_BOOL( "collision:231", mLarge.pointCount == mOrigin.pointCount );
	ENSURE( mLarge.pointCount == mOrigin.pointCount );
	ORACLE_F32( "collision:233/value0", mLarge.normal.x );
	ORACLE_F32( "collision:233/value1", mOrigin.normal.x );
	ORACLE_F32( "collision:233", mLarge.normal.x - mOrigin.normal.x );
	ENSURE_SMALL( mLarge.normal.x - mOrigin.normal.x, 1e-4f );
	ORACLE_F32( "collision:234/value0", mLarge.normal.y );
	ORACLE_F32( "collision:234/value1", mOrigin.normal.y );
	ORACLE_F32( "collision:234", mLarge.normal.y - mOrigin.normal.y );
	ENSURE_SMALL( mLarge.normal.y - mOrigin.normal.y, 1e-4f );
	for ( int i = 0; i < mLarge.pointCount; ++i )
	{
		ORACLE_F32( "collision:237/value0", mLarge.points[i].separation );
		ORACLE_F32( "collision:237/value1", mOrigin.points[i].separation );
		ORACLE_F32( "collision:237", mLarge.points[i].separation - mOrigin.points[i].separation );
		ENSURE_SMALL( mLarge.points[i].separation - mOrigin.points[i].separation, 1e-4f );
		ORACLE_F32( "collision:238/value0", mLarge.points[i].point.x );
		ORACLE_F32( "collision:238/value1", mOrigin.points[i].point.x );
		ORACLE_F32( "collision:238", mLarge.points[i].point.x - mOrigin.points[i].point.x );
		ENSURE_SMALL( mLarge.points[i].point.x - mOrigin.points[i].point.x, 1e-4f );
		ORACLE_F32( "collision:239/value0", mLarge.points[i].point.y );
		ORACLE_F32( "collision:239/value1", mOrigin.points[i].point.y );
		ORACLE_F32( "collision:239", mLarge.points[i].point.y - mOrigin.points[i].point.y );
		ENSURE_SMALL( mLarge.points[i].point.y - mOrigin.points[i].point.y, 1e-4f );
	}
#endif

	ORACLE_CLOSE();
	return 0;
}

// Broad-phase AABBs are built in double and narrowed to float with directed outward rounding, so a
// shape and its speculative margin stay inside their box far from the origin. A float build would
// round the extent away into the ULP (~1 m at 1e7) and clip the shape out of its own box.
static int LargeWorldAABBTest( void )
{
	ORACLE_OPEN( "collision/LargeWorldAABBTest" );
	// Rounded box: 0.5 half extents plus 0.1 radius, so the tight extent is 0.6 each way
	b2Polygon box = b2MakeRoundedBox( 0.5f, 0.5f, 0.1f );

	b2AABB aabbOrigin = b2ComputePolygonAABB( &box, b2WorldTransform_identity );
	ORACLE_F32( "collision:257/value0", aabbOrigin.lowerBound.x );
	ORACLE_F32( "collision:257", aabbOrigin.lowerBound.x + 0.6f );
	ENSURE_SMALL( aabbOrigin.lowerBound.x + 0.6f, FLT_EPSILON );
	ORACLE_F32( "collision:258/value0", aabbOrigin.lowerBound.y );
	ORACLE_F32( "collision:258", aabbOrigin.lowerBound.y + 0.6f );
	ENSURE_SMALL( aabbOrigin.lowerBound.y + 0.6f, FLT_EPSILON );
	ORACLE_F32( "collision:259/value0", aabbOrigin.upperBound.x );
	ORACLE_F32( "collision:259", aabbOrigin.upperBound.x - 0.6f );
	ENSURE_SMALL( aabbOrigin.upperBound.x - 0.6f, FLT_EPSILON );
	ORACLE_F32( "collision:260/value0", aabbOrigin.upperBound.y );
	ORACLE_F32( "collision:260", aabbOrigin.upperBound.y - 0.6f );
	ENSURE_SMALL( aabbOrigin.upperBound.y - 0.6f, FLT_EPSILON );

#if defined( BOX2D_DOUBLE_PRECISION )
	double d = 1.0e7;
	b2WorldTransform xfLarge = { { d, d }, b2Rot_identity };

	// Tight world AABB still contains the 0.6 m extent
	b2AABB tight = b2ComputePolygonAABB( &box, xfLarge );
	ORACLE_F32( "collision:268/value0", tight.lowerBound.x );
	ORACLE_BOOL( "collision:268", (double)tight.lowerBound.x <= d - 0.6 );
	ENSURE( (double)tight.lowerBound.x <= d - 0.6 );
	ORACLE_F32( "collision:269/value0", tight.lowerBound.y );
	ORACLE_BOOL( "collision:269", (double)tight.lowerBound.y <= d - 0.6 );
	ENSURE( (double)tight.lowerBound.y <= d - 0.6 );
	ORACLE_F32( "collision:270/value0", tight.upperBound.x );
	ORACLE_BOOL( "collision:270", (double)tight.upperBound.x >= d + 0.6 );
	ENSURE( (double)tight.upperBound.x >= d + 0.6 );
	ORACLE_F32( "collision:271/value0", tight.upperBound.y );
	ORACLE_BOOL( "collision:271", (double)tight.upperBound.y >= d + 0.6 );
	ENSURE( (double)tight.upperBound.y >= d + 0.6 );

	// The fat helper folds the extra into the double step before the single outward rounding, so a
	// margin smaller than a float ULP at this range survives instead of becoming a no-op subtract.
	float extra = 0.05f;
	b2Shape shape = { 0 };
	shape.type = b2_polygonShape;
	shape.polygon = box;
	b2AABB fat = b2ComputeFatShapeAABB( &shape, xfLarge, extra );
	ORACLE_F32( "collision:280/value0", fat.lowerBound.x );
	ORACLE_BOOL( "collision:280", (double)fat.lowerBound.x <= d - 0.6 - (double)extra );
	ENSURE( (double)fat.lowerBound.x <= d - 0.6 - (double)extra );
	ORACLE_F32( "collision:281/value0", fat.lowerBound.y );
	ORACLE_BOOL( "collision:281", (double)fat.lowerBound.y <= d - 0.6 - (double)extra );
	ENSURE( (double)fat.lowerBound.y <= d - 0.6 - (double)extra );
	ORACLE_F32( "collision:282/value0", fat.upperBound.x );
	ORACLE_BOOL( "collision:282", (double)fat.upperBound.x >= d + 0.6 + (double)extra );
	ENSURE( (double)fat.upperBound.x >= d + 0.6 + (double)extra );
	ORACLE_F32( "collision:283/value0", fat.upperBound.y );
	ORACLE_BOOL( "collision:283", (double)fat.upperBound.y >= d + 0.6 + (double)extra );
	ENSURE( (double)fat.upperBound.y >= d + 0.6 + (double)extra );
#endif

	ORACLE_CLOSE();
	return 0;
}

int CollisionTest( void )
{
	RUN_SUBTEST( AABBTest );
	RUN_SUBTEST( AABBRayCastTest );
	RUN_SUBTEST( LargeWorldManifoldTest );
	RUN_SUBTEST( LargeWorldAABBTest );

	return 0;
}
