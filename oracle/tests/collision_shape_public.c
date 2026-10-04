// SPDX-License-Identifier: MIT

#include "oracle.h"

#include "box2d/collision.h"
#include "box2d/math_functions.h"

#include <assert.h>
#include <string.h>

// Standalone additions: the four pinned ShapeTest cases remain in collision_shape.c.
static void shape_public_polygon( const char* key, b2Polygon p )
{
	char prefix[256];
	strcpy( prefix, key );
	key = prefix;
	ORACLE_I32( ORACLE_KEY( "%s/count", key ), p.count );
	ORACLE_VEC2( ORACLE_KEY( "%s/centroid", key ), p.centroid );
	ORACLE_F32( ORACLE_KEY( "%s/radius", key ), p.radius );
	for ( int i = 0; i < p.count; ++i )
	{
		ORACLE_VEC2( ORACLE_KEY( "%s/vertex/%d", key, i ), p.vertices[i] );
		ORACLE_VEC2( ORACLE_KEY( "%s/normal/%d", key, i ), p.normals[i] );
	}
}

static void shape_public_mass( const char* key, b2MassData m )
{
	ORACLE_F32( ORACLE_KEY( "%s/mass", key ), m.mass );
	ORACLE_VEC2( ORACLE_KEY( "%s/center", key ), m.center );
	ORACLE_F32( ORACLE_KEY( "%s/inertia", key ), m.rotationalInertia );
}

static void shape_public_aabb( const char* key, b2AABB a )
{
	ORACLE_VEC2( ORACLE_KEY( "%s/lower", key ), a.lowerBound );
	ORACLE_VEC2( ORACLE_KEY( "%s/upper", key ), a.upperBound );
}

static void shape_public_cast( const char* key, b2CastOutput c )
{
	// These C entry points initialize their entire output to zero, including misses.
	ORACLE_BOOL( ORACLE_KEY( "%s/hit", key ), c.hit );
	ORACLE_VEC2( ORACLE_KEY( "%s/point", key ), c.point );
	ORACLE_VEC2( ORACLE_KEY( "%s/normal", key ), c.normal );
	ORACLE_F32( ORACLE_KEY( "%s/fraction", key ), c.fraction );
	ORACLE_I32( ORACLE_KEY( "%s/iterations", key ), c.iterations );
}

static void shape_public_hulls( void )
{
	const struct
	{
		const char* name;
		int count;
		b2Vec2 points[9];
	} cases[] = {
		{ "count0", 0, { { 0, 0 } } },
		{ "count1", 1, { { 1, 2 } } },
		{ "count2", 2, { { -1, 0 }, { 1, 0 } } },
		{ "count3", 3, { { 0, 2 }, { -2, -1 }, { 2, -1 } } },
		{ "count8", 8, { { 2, 1 }, { -1, -2 }, { -2, 1 }, { 1, 2 }, { 1, -2 }, { -2, -1 }, { -1, 2 }, { 2, -1 } } },
		{ "count9", 9, { { 2, 1 }, { -1, -2 }, { -2, 1 }, { 1, 2 }, { 1, -2 }, { -2, -1 }, { -1, 2 }, { 2, -1 }, { 0, 0 } } },
		{ "duplicates", 6, { { -1, -1 }, { 1, 1 }, { -1, 1 }, { 1, -1 }, { 1, 1 }, { -1, -1 } } },
		{ "near_coincident", 5, { { -1, -1 }, { 1, 1 }, { -1, 1 }, { 1, -1 }, { 1.001f, 1.001f } } },
		{ "all_coincident", 3, { { 1, 1 }, { 1.001f, 1 }, { 1, 1.001f } } },
		{ "collinear", 5, { { 2, 2 }, { -2, -2 }, { 0, 0 }, { 1, 1 }, { -1, -1 } } },
		{ "collinear_edge", 5, { { -1, -1 }, { 0, -1 }, { 1, -1 }, { 1, 1 }, { -1, 1 } } },
		{ "unordered_convex", 6, { { 3, 0 }, { -2, 1 }, { 1, -2 }, { 1, 2 }, { -2, -1 }, { 0, 0 } } },
	};
	const b2Rot q = { 0.6f, 0.8f };
	const b2Vec2 offset = { 2.5f, -3.0f };
	const b2Transform xf = { { -4.0f, 1.5f }, { 0.0f, 1.0f } };

	ORACLE_OPEN( "shape_public/HullAndConstructors" );
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		const char* name = cases[i].name;
		b2Hull h = b2ComputeHull( cases[i].points, cases[i].count );
		ORACLE_I32( ORACLE_KEY( "%s/count", name ), h.count );
		for ( int j = 0; j < h.count; ++j )
		{
			ORACLE_VEC2( ORACLE_KEY( "%s/point/%d", name, j ), h.points[j] );
		}
		bool valid = b2ValidateHull( &h );
		ORACLE_BOOL( ORACLE_KEY( "%s/valid", name ), valid );
		if ( valid )
		{
			b2Polygon plain = b2MakePolygon( &h, 0.0f );
			b2Polygon rounded = b2MakePolygon( &h, 0.25f );
			shape_public_polygon( ORACLE_KEY( "%s/plain", name ), plain );
			shape_public_polygon( ORACLE_KEY( "%s/rounded", name ), rounded );
			shape_public_polygon( ORACLE_KEY( "%s/offset", name ), b2MakeOffsetPolygon( &h, offset, q ) );
			shape_public_polygon( ORACLE_KEY( "%s/offset_rounded", name ), b2MakeOffsetRoundedPolygon( &h, offset, q, 0.25f ) );
			shape_public_polygon( ORACLE_KEY( "%s/transformed", name ), b2TransformPolygon( xf, &rounded ) );
		}
	}
	shape_public_polygon( "square", b2MakeSquare( 1.25f ) );
	shape_public_polygon( "box", b2MakeBox( 1.5f, 0.75f ) );
	b2Polygon roundedBox = b2MakeRoundedBox( 1.5f, 0.75f, 0.2f );
	shape_public_polygon( "rounded_box", roundedBox );
	shape_public_polygon( "offset_box", b2MakeOffsetBox( 1.5f, 0.75f, offset, q ) );
	shape_public_polygon( "offset_rounded_box", b2MakeOffsetRoundedBox( 1.5f, 0.75f, offset, q, 0.2f ) );
	shape_public_polygon( "transformed_box", b2TransformPolygon( xf, &roundedBox ) );
	ORACLE_CLOSE();
}

static void shape_public_masses( void )
{
	const b2Circle circles[] = { { { 0, 0 }, 1 }, { { 2, -3 }, 0.75f }, { { -4, 2 }, 0 } };
	const b2Capsule capsules[] = {
		{ { -1, 0 }, { 1, 0 }, 0.5f },
		{ { 2, -3 }, { 5, 1 }, 0.75f },
		{ { 1, 2 }, { 1, 2 }, 0.5f },
	};
	const b2Vec2 points[] = { { 0, 2 }, { -2, -1 }, { 2, -1 } };
	b2Hull h = b2ComputeHull( points, 3 );
	assert( b2ValidateHull( &h ) );
	const b2Polygon polygons[] = {
		b2MakeBox( 1.5f, 0.75f ),
		b2MakeRoundedBox( 1.5f, 0.75f, 0.25f ),
		b2MakeOffsetRoundedBox( 1.5f, 0.75f, (b2Vec2){ 2, -3 }, (b2Rot){ 0.6f, 0.8f }, 0.25f ),
		b2MakePolygon( &h, 0.125f ),
	};
	const float densities[] = { 0, 0.75f, 2.5f };
	ORACLE_OPEN( "shape_public/Mass" );
	for ( int d = 0; d < 3; ++d )
	{
		for ( int i = 0; i < 3; ++i )
		{
			shape_public_mass( ORACLE_KEY( "density/%d/circle/%d", d, i ), b2ComputeCircleMass( &circles[i], densities[d] ) );
			shape_public_mass( ORACLE_KEY( "density/%d/capsule/%d", d, i ), b2ComputeCapsuleMass( &capsules[i], densities[d] ) );
		}
		for ( int i = 0; i < 4; ++i )
		{
			shape_public_mass( ORACLE_KEY( "density/%d/polygon/%d", d, i ), b2ComputePolygonMass( &polygons[i], densities[d] ) );
		}
	}
	ORACLE_CLOSE();
}

static void shape_public_bounds( void )
{
	const b2Circle circle = { { 1.25f, -2 }, 0.7f };
	const b2Capsule capsule = { { -1.25f, 0.5f }, { 2, -0.75f }, 0.3f };
	const b2Segment segment = { { -1.25f, 0.5f }, { 2, -0.75f } };
	const b2Polygon polygon = b2MakeOffsetRoundedBox( 1.5f, 0.75f, (b2Vec2){ 0.25f, -0.5f }, (b2Rot){ 0.6f, 0.8f }, 0.2f );
	const struct
	{
		const char* name;
		b2WorldTransform xf;
	} frames[] = {
		{ "identity", { { 0, 0 }, { 1, 0 } } },
		{ "translated", { { 7.25f, -4.5f }, { 1, 0 } } },
		{ "rotated", { { 7.25f, -4.5f }, { 0.6f, 0.8f } } },
		{ "distant", { { 10000000, -10000000 }, { 0.6f, 0.8f } } },
	};
	ORACLE_OPEN( "shape_public/AABB" );
	for ( int i = 0; i < 4; ++i )
	{
		shape_public_aabb( ORACLE_KEY( "%s/circle", frames[i].name ), b2ComputeCircleAABB( &circle, frames[i].xf ) );
		shape_public_aabb( ORACLE_KEY( "%s/capsule", frames[i].name ), b2ComputeCapsuleAABB( &capsule, frames[i].xf ) );
		shape_public_aabb( ORACLE_KEY( "%s/segment", frames[i].name ), b2ComputeSegmentAABB( &segment, frames[i].xf ) );
		shape_public_aabb( ORACLE_KEY( "%s/polygon", frames[i].name ), b2ComputePolygonAABB( &polygon, frames[i].xf ) );
	}
	ORACLE_CLOSE();
}

static void shape_public_contains( void )
{
	const b2Circle circle = { { 2, -1 }, 1 };
	const b2Capsule capsule = { { -1, 0 }, { 1, 0 }, 0.5f };
	const b2Capsule degenerate = { { 2, -1 }, { 2, -1 }, 1 };
	const b2Polygon box = b2MakeBox( 1, 1 );
	const b2Polygon rounded = b2MakeRoundedBox( 1, 1, 0.5f );
	const b2Polygon offset = b2MakeOffsetBox( 1, 1, (b2Vec2){ 2, -1 }, (b2Rot){ 0, 1 } );
	const b2Vec2 circlePoints[] = { { 2, -1 }, { 4, -1 }, { 3, -1 }, { 2, 0 }, { 2.999f, -1 }, { 3.001f, -1 } };
	const b2Vec2 capsulePoints[] = { { 0, 0 }, { 0, 2 }, { 0, 0.5f }, { -1.5f, 0 }, { 1.5f, 0 }, { 1.501f, 0 } };
	const b2Vec2 polygonPoints[] = { { 0, 0 }, { 3, 0 }, { 1, 0 }, { 1, 1 }, { 1.5f, 0 }, { 1.501f, 0 }, { 1.5f, 1.5f } };
	ORACLE_OPEN( "shape_public/Containment" );
	for ( int i = 0; i < 6; ++i )
	{
		ORACLE_BOOL( ORACLE_KEY( "circle/%d", i ), b2PointInCircle( &circle, circlePoints[i] ) );
		ORACLE_BOOL( ORACLE_KEY( "capsule/%d", i ), b2PointInCapsule( &capsule, capsulePoints[i] ) );
		ORACLE_BOOL( ORACLE_KEY( "degenerate_capsule/%d", i ), b2PointInCapsule( &degenerate, circlePoints[i] ) );
		ORACLE_BOOL( ORACLE_KEY( "offset_polygon/%d", i ), b2PointInPolygon( &offset, circlePoints[i] ) );
	}
	for ( int i = 0; i < 7; ++i )
	{
		ORACLE_BOOL( ORACLE_KEY( "polygon/%d", i ), b2PointInPolygon( &box, polygonPoints[i] ) );
		ORACLE_BOOL( ORACLE_KEY( "rounded_polygon/%d", i ), b2PointInPolygon( &rounded, polygonPoints[i] ) );
	}
	ORACLE_CLOSE();
}

static void shape_public_rays( void )
{
	const b2Circle circle = { { 0, 0 }, 1 };
	const b2Capsule capsule = { { -1, 0 }, { 1, 0 }, 1 };
	const b2Capsule degenerate = { { 0, 0 }, { 0, 0 }, 1 };
	const b2Segment segment = { { 0, -1 }, { 0, 1 } };
	const b2Segment reversed = { { 0, 1 }, { 0, -1 } };
	const b2Polygon polygon = b2MakeBox( 1, 1 );
	const b2Polygon rounded = b2MakeRoundedBox( 1, 1, 0.25f );
	const struct
	{
		const char* name;
		b2RayCastInput input;
	} cases[] = {
		{ "hit", { { -4, 0 }, { 8, 0 }, 1 } },		   { "reverse", { { 4, 0 }, { -8, 0 }, 1 } },
		{ "miss", { { -4, 4 }, { 8, 0 }, 1 } },		   { "away", { { -4, 0 }, { -8, 0 }, 1 } },
		{ "overlap", { { 0, 0 }, { 8, 0 }, 1 } },	   { "boundary", { { -1, 0 }, { 8, 0 }, 1 } },
		{ "graze", { { -4, 1 }, { 8, 0 }, 1 } },	   { "max_fraction", { { -4, 0 }, { 8, 0 }, 0.2f } },
		{ "endpoint", { { -4, 0 }, { 8, 0 }, 0.5f } }, { "zero_translation", { { -4, 0 }, { 0, 0 }, 1 } },
		{ "capsule_cap", { { 1, 4 }, { 0, -8 }, 1 } },
	};
	ORACLE_OPEN( "shape_public/RayCasts" );
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		const char* n = cases[i].name;
		const b2RayCastInput* in = &cases[i].input;
		shape_public_cast( ORACLE_KEY( "%s/circle", n ), b2RayCastCircle( &circle, in ) );
		shape_public_cast( ORACLE_KEY( "%s/capsule", n ), b2RayCastCapsule( &capsule, in ) );
		shape_public_cast( ORACLE_KEY( "%s/degenerate_capsule", n ), b2RayCastCapsule( &degenerate, in ) );
		shape_public_cast( ORACLE_KEY( "%s/polygon", n ), b2RayCastPolygon( &polygon, in ) );
		shape_public_cast( ORACLE_KEY( "%s/rounded_polygon", n ), b2RayCastPolygon( &rounded, in ) );
		shape_public_cast( ORACLE_KEY( "%s/segment_two_sided", n ), b2RayCastSegment( &segment, in, false ) );
		shape_public_cast( ORACLE_KEY( "%s/segment_one_sided", n ), b2RayCastSegment( &segment, in, true ) );
		shape_public_cast( ORACLE_KEY( "%s/reversed_two_sided", n ), b2RayCastSegment( &reversed, in, false ) );
		shape_public_cast( ORACLE_KEY( "%s/reversed_one_sided", n ), b2RayCastSegment( &reversed, in, true ) );
	}
	ORACLE_CLOSE();
}

static void shape_public_sweeps( void )
{
	const b2Circle circle = { { 0, 0 }, 1 };
	const b2Capsule capsule = { { -1, 0 }, { 1, 0 }, 1 };
	const b2Segment segment = { { 0, -1 }, { 0, 1 } };
	const b2Polygon polygon = b2MakeBox( 1, 1 );
	const b2Polygon rounded = b2MakeRoundedBox( 1, 1, 0.25f );
	const struct
	{
		const char* name;
		b2Vec2 origin, translation;
		float radius, maxFraction;
		bool encroach;
	} cases[] = {
		{ "hit", { -4, 0 }, { 8, 0 }, 0.25f, 1, false },
		{ "reverse", { 4, 0 }, { -8, 0 }, 0.25f, 1, false },
		{ "miss", { -4, 4 }, { 8, 0 }, 0.25f, 1, false },
		{ "away", { -4, 0 }, { -8, 0 }, 0.25f, 1, false },
		{ "overlap", { 0, 0 }, { 8, 0 }, 0.25f, 1, false },
		{ "graze", { -4, 1.25f }, { 8, 0 }, 0.25f, 1, false },
		{ "max_fraction", { -4, 0 }, { 8, 0 }, 0.25f, 0.2f, false },
		{ "endpoint", { -4, 0 }, { 8, 0 }, 0.25f, 0.5f, false },
		{ "zero_translation", { -4, 0 }, { 0, 0 }, 0.25f, 1, false },
		{ "zero_radius", { -4, 0 }, { 8, 0 }, 0, 1, false },
	};
	ORACLE_OPEN( "shape_public/ShapeCasts" );
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		b2ShapeCastInput in = { 0 };
		in.proxy = b2MakeProxy( &cases[i].origin, 1, cases[i].radius );
		in.translation = cases[i].translation;
		in.maxFraction = cases[i].maxFraction;
		in.canEncroach = cases[i].encroach;
		const char* n = cases[i].name;
		shape_public_cast( ORACLE_KEY( "%s/circle", n ), b2ShapeCastCircle( &circle, &in ) );
		shape_public_cast( ORACLE_KEY( "%s/capsule", n ), b2ShapeCastCapsule( &capsule, &in ) );
		shape_public_cast( ORACLE_KEY( "%s/segment", n ), b2ShapeCastSegment( &segment, &in ) );
		shape_public_cast( ORACLE_KEY( "%s/polygon", n ), b2ShapeCastPolygon( &polygon, &in ) );
		shape_public_cast( ORACLE_KEY( "%s/rounded_polygon", n ), b2ShapeCastPolygon( &rounded, &in ) );
	}
	// Initially touching, moving inward: encroachment must be forwarded by every wrapper.
	const b2Vec2 touching[] = { { -1.25f, 0 }, { -2.25f, 0 }, { -0.25f, 0 }, { -1.25f, 0 }, { -1.5f, 0 } };
	for ( int e = 0; e < 2; ++e )
	{
		b2ShapeCastInput in = { 0 };
		in.translation = (b2Vec2){ 2, 0 };
		in.maxFraction = 1;
		in.canEncroach = e != 0;
		in.proxy = b2MakeProxy( &touching[0], 1, 0.25f );
		shape_public_cast( ORACLE_KEY( "touch/%d/circle", e ), b2ShapeCastCircle( &circle, &in ) );
		in.proxy = b2MakeProxy( &touching[1], 1, 0.25f );
		shape_public_cast( ORACLE_KEY( "touch/%d/capsule", e ), b2ShapeCastCapsule( &capsule, &in ) );
		in.proxy = b2MakeProxy( &touching[2], 1, 0.25f );
		shape_public_cast( ORACLE_KEY( "touch/%d/segment", e ), b2ShapeCastSegment( &segment, &in ) );
		in.proxy = b2MakeProxy( &touching[3], 1, 0.25f );
		shape_public_cast( ORACLE_KEY( "touch/%d/polygon", e ), b2ShapeCastPolygon( &polygon, &in ) );
		in.proxy = b2MakeProxy( &touching[4], 1, 0.25f );
		shape_public_cast( ORACLE_KEY( "touch/%d/rounded_polygon", e ), b2ShapeCastPolygon( &rounded, &in ) );
	}
	const b2Vec2 capsulePoints[] = { { -4, -0.5f }, { -4, 0.5f } };
	const b2Vec2 boxPoints[] = { { -4.5f, -0.5f }, { -3.5f, -0.5f }, { -3.5f, 0.5f }, { -4.5f, 0.5f } };
	const b2ShapeProxy proxies[] = { b2MakeProxy( capsulePoints, 2, 0.25f ), b2MakeProxy( boxPoints, 4, 0 ) };
	for ( int i = 0; i < 2; ++i )
	{
		b2ShapeCastInput in = { proxies[i], { 8, 0 }, 1, false };
		shape_public_cast( ORACLE_KEY( "proxy/%d/circle", i ), b2ShapeCastCircle( &circle, &in ) );
		shape_public_cast( ORACLE_KEY( "proxy/%d/capsule", i ), b2ShapeCastCapsule( &capsule, &in ) );
		shape_public_cast( ORACLE_KEY( "proxy/%d/segment", i ), b2ShapeCastSegment( &segment, &in ) );
		shape_public_cast( ORACLE_KEY( "proxy/%d/polygon", i ), b2ShapeCastPolygon( &polygon, &in ) );
		shape_public_cast( ORACLE_KEY( "proxy/%d/rounded_polygon", i ), b2ShapeCastPolygon( &rounded, &in ) );
	}
	ORACLE_CLOSE();
}

int main( void )
{
	shape_public_hulls();
	shape_public_masses();
	shape_public_bounds();
	shape_public_contains();
	shape_public_rays();
	shape_public_sweeps();
	return 0;
}
