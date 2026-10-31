// SPDX-License-Identifier: MIT
#include "oracle.h"

#include "box2d/collision.h"
#include "box2d/math_functions.h"

#include <stdio.h>
#include <string.h>

static b2ShapeProxy distance_public_box( float radius )
{
	const b2Vec2 points[] = { { -1, -1 }, { 1, -1 }, { 1, 1 }, { -1, 1 } };
	return b2MakeProxy( points, 4, radius );
}

static void distance_public_proxy( const char* prefix, b2ShapeProxy p )
{
	char key[256];
	strcpy( key, prefix );
	ORACLE_I32( ORACLE_KEY( "%s/count", key ), p.count );
	ORACLE_F32( ORACLE_KEY( "%s/radius", key ), p.radius );
	for ( int i = 0; i < p.count; ++i )
	{
		ORACLE_VEC2( ORACLE_KEY( "%s/point/%d", key, i ), p.points[i] );
	}
}

static void distance_public_output( const char* prefix, b2DistanceOutput o, b2SimplexCache cache, const b2Simplex* debug )
{
	char key[256];
	strcpy( key, prefix );
	ORACLE_VEC2( ORACLE_KEY( "%s/pointA", key ), o.pointA );
	ORACLE_VEC2( ORACLE_KEY( "%s/pointB", key ), o.pointB );
	ORACLE_F32( ORACLE_KEY( "%s/distance", key ), o.distance );
	// The public normal is invalid at zero distance.
	if ( o.distance > 0 )
	{
		ORACLE_VEC2( ORACLE_KEY( "%s/normal", key ), o.normal );
	}
	ORACLE_I32( ORACLE_KEY( "%s/iterations", key ), o.iterations );
	ORACLE_I32( ORACLE_KEY( "%s/simplexCount", key ), o.simplexCount );
	ORACLE_U32( ORACLE_KEY( "%s/cache/count", key ), cache.count );
	for ( int i = 0; i < cache.count; ++i )
	{
		ORACLE_U32( ORACLE_KEY( "%s/cache/%d/indexA", key, i ), cache.indexA[i] );
		ORACLE_U32( ORACLE_KEY( "%s/cache/%d/indexB", key, i ), cache.indexB[i] );
	}
	for ( int i = 0; i < o.simplexCount; ++i )
	{
		const b2Simplex* s = debug + i;
		ORACLE_I32( ORACLE_KEY( "%s/simplex/%d/count", key, i ), s->count );
		const b2SimplexVertex* vertices[] = { &s->v1, &s->v2, &s->v3 };
		for ( int j = 0; j < s->count; ++j )
		{
			const b2SimplexVertex* v = vertices[j];
			char vk[256];
			snprintf( vk, sizeof( vk ), "%s/simplex/%d/vertex/%d", key, i, j );
			ORACLE_VEC2( ORACLE_KEY( "%s/wA", vk ), v->wA );
			ORACLE_VEC2( ORACLE_KEY( "%s/wB", vk ), v->wB );
			ORACLE_VEC2( ORACLE_KEY( "%s/w", vk ), v->w );
			ORACLE_F32( ORACLE_KEY( "%s/a", vk ), v->a );
			ORACLE_I32( ORACLE_KEY( "%s/indexA", vk ), v->indexA );
			ORACLE_I32( ORACLE_KEY( "%s/indexB", vk ), v->indexB );
		}
	}
}

static void distance_public_cast( b2CastOutput o )
{
	ORACLE_BOOL( "hit", o.hit );
	ORACLE_VEC2( "point", o.point );
	ORACLE_VEC2( "normal", o.normal );
	ORACLE_F32( "fraction", o.fraction );
	ORACLE_I32( "iterations", o.iterations );
}

static void distance_public_toi( const char* prefix, b2TOIOutput o )
{
	char key[256];
	strcpy( key, prefix );
	ORACLE_I32( ORACLE_KEY( "%s/state", key ), o.state );
	ORACLE_F32( ORACLE_KEY( "%s/fraction", key ), o.fraction );
	// Non-hit initialized fields are payload, not a promise of contact geometry.
	ORACLE_VEC2( ORACLE_KEY( "%s/point", key ), o.point );
	ORACLE_VEC2( ORACLE_KEY( "%s/normal", key ), o.normal );
}

static void segments( void )
{
	const struct
	{
		const char* name;
		b2Vec2 p1, q1, p2, q2;
	} cases[] = {
		{ "intersect", { -2, 0 }, { 2, 0 }, { 0, -3 }, { 0, 3 } },
		{ "parallel", { -2, 0 }, { 2, 0 }, { -1, 2 }, { 3, 2 } },
		{ "parallel-reversed", { -2, 0 }, { 2, 0 }, { 3, 2 }, { -1, 2 } },
		{ "clamp-start", { 0, 0 }, { 1, 0 }, { -2, 1 }, { -1, 2 } },
		{ "clamp-end", { 0, 0 }, { 1, 0 }, { 3, 2 }, { 2, 1 } },
		{ "first-point", { 1, 1 }, { 1, 1 }, { -2, 0 }, { 2, 0 } },
		{ "second-point", { -2, 0 }, { 2, 0 }, { 1, 1 }, { 1, 1 } },
		{ "both-points", { 1, 1 }, { 1, 1 }, { -2, 3 }, { -2, 3 } },
		{ "tiny-first", { 0, 0 }, { 0x1p-25f, 0 }, { -1, 2 }, { 1, 2 } },
	};
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		ORACLE_OPEN( ORACLE_KEY( "distance_public/segments/%s", cases[i].name ) );
		b2SegmentDistanceResult o = b2SegmentDistance( cases[i].p1, cases[i].q1, cases[i].p2, cases[i].q2 );
		ORACLE_VEC2( "closest1", o.closest1 );
		ORACLE_VEC2( "closest2", o.closest2 );
		ORACLE_F32( "fraction1", o.fraction1 );
		ORACLE_F32( "fraction2", o.fraction2 );
		ORACLE_F32( "distanceSquared", o.distanceSquared );
		ORACLE_CLOSE();
	}
}

static void proxies( void )
{
	const int counts[] = { 1, 8, 9 };
	for ( int j = 0; j < 3; ++j )
	{
		int count = counts[j];
		b2Vec2 points[] = { { -2, -1 }, { 0, -2 }, { 2, -1 }, { 3, 0 }, { 2, 1 }, { 0, 2 }, { -2, 1 }, { -3, 0 }, { 9, 9 } };
		ORACLE_OPEN( ORACLE_KEY( "distance_public/proxy/%d", count ) );
		b2ShapeProxy p = b2MakeProxy( points, count, 0.375f );
		b2ShapeProxy o = b2MakeOffsetProxy( points, count, 0.625f, (b2Vec2){ 3, -4 }, (b2Rot){ 0.6f, 0.8f } );
		distance_public_proxy( "plain", p );
		distance_public_proxy( "offset", o );
		for ( int i = 0; i < count; ++i )
		{
			points[i] = (b2Vec2){ 100 + (float)i, -100 - (float)i };
		}
		distance_public_proxy( "plain-after-source-mutation", p );
		distance_public_proxy( "offset-after-source-mutation", o );
		ORACLE_CLOSE();
	}
}

static void sweep_transforms( void )
{
	const b2Sweep sweeps[] = {
		{ { 2, -3 }, { -5, 4 }, { 7, -2 }, { 1, 0 }, { 0, 1 } },
		{ { -0.5f, 1.25f }, { 3, 2 }, { 3, 2 }, { 0.6f, -0.8f }, { 0.6f, 0.8f } },
	};
	const float times[] = { 0, 0.25f, 0.5f, 0.75f, 1 };
	ORACLE_OPEN( "distance_public/sweep-transforms" );
	for ( int i = 0; i < 2; ++i )
	{
		for ( int j = 0; j < 5; ++j )
		{
			b2Transform xf = b2GetSweepTransform( sweeps + i, times[j] );
			ORACLE_VEC2( ORACLE_KEY( "%d/%d/p", i, j ), xf.p );
			ORACLE_ROT( ORACLE_KEY( "%d/%d/q", i, j ), xf.q );
		}
	}
	ORACLE_CLOSE();
}

static void distances( void )
{
	const b2Vec2 sp[] = { { 0, -1 }, { 0, 1 } };
	const b2Vec2 pp[] = { { 0, 0 } };
	const b2Vec2 octagon[] = { { -2, -1 }, { 0, -2 }, { 2, -1 }, { 3, 0 }, { 2, 1 }, { 0, 2 }, { -2, 1 }, { -3, 0 } };
	b2ShapeProxy segment = b2MakeProxy( sp, 2, 0 );
	b2ShapeProxy point = b2MakeProxy( pp, 1, 0.5f );
	const struct
	{
		const char* name;
		b2ShapeProxy a, b;
		b2Transform xf;
		bool radii;
	} cases[] = {
		{ "separated", distance_public_box( 0 ), segment, { { 3, 0 }, { 1, 0 } }, false },
		{ "touch", distance_public_box( 0 ), segment, { { 1, 0 }, { 1, 0 } }, false },
		{ "overlap", distance_public_box( 0 ), distance_public_box( 0 ), { { 0.25f, 0.5f }, { 1, 0 } }, false },
		{ "radii-separated", point, point, { { 3, 0 }, { 1, 0 } }, true },
		{ "radii-touch", point, point, { { 1, 0 }, { 1, 0 } }, true },
		{ "radii-overlap", point, point, { { 0.75f, 0 }, { 1, 0 } }, true },
		{ "radii-ignored", point, point, { { 0.75f, 0 }, { 1, 0 } }, false },
		{ "transformed", distance_public_box( 0.125f ), segment, { { 3, -0.5f }, { 0.6f, 0.8f } }, true },
		{ "max-count", b2MakeProxy( octagon, 8, 0.25f ), point, { { 4, 2 }, { 1, 0 } }, true },
	};
	const int capacities[] = { 0, 1, 24 };
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		ORACLE_OPEN( ORACLE_KEY( "distance_public/distance/%s", cases[i].name ) );
		b2DistanceInput input = { cases[i].a, cases[i].b, cases[i].xf, cases[i].radii };
		for ( int j = 0; j < 3; ++j )
		{
			b2SimplexCache cache = { 0 };
			b2Simplex debug[24];
			int capacity = capacities[j];
			b2DistanceOutput o = b2ShapeDistance( &input, &cache, capacity == 0 ? NULL : debug, capacity );
			distance_public_output( ORACLE_KEY( "capacity/%d", capacity ), o, cache, debug );
		}
		ORACLE_CLOSE();
	}
}

static void cache_history( void )
{
	b2DistanceInput input = { .proxyA = distance_public_box( 0.125f ), .proxyB = distance_public_box( 0.25f ), .useRadii = true };
	const b2Transform transforms[] = {
		{ { 4, 0.25f }, { 1, 0 } },		   { { 3.75f, 0.5f }, { 1, 0 } }, { { 3.5f, 1 }, { 0.6f, 0.8f } },
		{ { 2, 1 }, { 0.6f, 0.8f } },	   { { 0.5f, 0.25f }, { 1, 0 } }, { { -3, -1 }, { 0, -1 } },
		{ { -3.25f, -0.75f }, { 0, -1 } }, { { 4, 0.25f }, { 1, 0 } },
	};
	b2SimplexCache cache = { 0 };
	ORACLE_OPEN( "distance_public/cache-history" );
	for ( int i = 0; i < 8; ++i )
	{
		input.transform = transforms[i];
		b2Simplex debug[24];
		b2DistanceOutput o = b2ShapeDistance( &input, &cache, debug, 24 );
		distance_public_output( ORACLE_KEY( "%d", i ), o, cache, debug );
	}
	ORACLE_CLOSE();
}

static void casts( void )
{
	const b2Vec2 pp[] = { { 0, 0 } };
	const b2Vec2 sp[] = { { 0, -1 }, { 0, 1 } };
	b2ShapeProxy point = b2MakeProxy( pp, 1, 0.5f );
	b2ShapeProxy segment = b2MakeProxy( sp, 2, 0 );
	const struct
	{
		const char* name;
		b2ShapeProxy a, b;
		b2Vec2 p;
		b2Rot q;
		b2Vec2 delta;
		float max;
		bool encroach;
	} cases[] = {
		{ "hit", distance_public_box( 0 ), segment, { 3, 0 }, { 1, 0 }, { -4, 0 }, 1, false },
		{ "away", distance_public_box( 0 ), segment, { 3, 0 }, { 1, 0 }, { 4, 0 }, 1, false },
		{ "stationary", distance_public_box( 0 ), segment, { 3, 0 }, { 1, 0 }, { 0, 0 }, 1, false },
		{ "too-short", distance_public_box( 0 ), segment, { 3, 0 }, { 1, 0 }, { -4, 0 }, 0.25f, false },
		{ "fraction-boundary", distance_public_box( 0 ), segment, { 3, 0 }, { 1, 0 }, { -4, 0 }, 0.5f, false },
		{ "initial-touch", distance_public_box( 0 ), segment, { 1, 0 }, { 1, 0 }, { -1, 0 }, 1, false },
		{ "initial-overlap", distance_public_box( 0 ), distance_public_box( 0 ), { 0.25f, 0.5f }, { 1, 0 }, { -1, 0 }, 1, false },
		{ "graze", distance_public_box( 0 ), b2MakeProxy( pp, 1, 0 ), { 3, 1 }, { 1, 0 }, { -4, 0 }, 1, false },
		{ "graze-miss", distance_public_box( 0 ), b2MakeProxy( pp, 1, 0 ), { 3, 1.25f }, { 1, 0 }, { -4, 0 }, 1, false },
		{ "rounded-hit", point, point, { 3, 0.5f }, { 1, 0 }, { -4, 0 }, 1, false },
		{ "rounded-touch", point, point, { 1, 0 }, { 1, 0 }, { -1, 0 }, 1, false },
		{ "encroach-off", point, point, { 0.99f, 0 }, { 1, 0 }, { -1, 0 }, 1, false },
		{ "encroach-on", point, point, { 0.99f, 0 }, { 1, 0 }, { -1, 0 }, 1, true },
		{ "encroach-away", point, point, { 0.99f, 0 }, { 1, 0 }, { 1, 0 }, 1, true },
		{ "transformed", distance_public_box( 0.125f ), segment, { 4, 0.5f }, { 0.6f, 0.8f }, { -5, -0.5f }, 1, false },
	};
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		ORACLE_OPEN( ORACLE_KEY( "distance_public/cast/%s", cases[i].name ) );
		b2ShapeCastPairInput input = { cases[i].a,	   cases[i].b,	 { cases[i].p, cases[i].q },
									   cases[i].delta, cases[i].max, cases[i].encroach };
		distance_public_cast( b2ShapeCast( &input ) );
		ORACLE_CLOSE();
	}
}

static void toi_cases( void )
{
	const struct
	{
		const char* name;
		b2Vec2 start, end;
		b2Rot q1, q2;
		b2Vec2 localCenter;
		float max;
	} cases[] = {
		{ "translating-hit", { 4, 0 }, { 0, 0 }, { 1, 0 }, { 1, 0 }, { 0, 0 }, 1 },
		{ "translating-separated", { 4, 0 }, { 6, 0 }, { 1, 0 }, { 1, 0 }, { 0, 0 }, 1 },
		{ "initial-overlap", { 0.25f, 0.5f }, { 4, 0 }, { 1, 0 }, { 1, 0 }, { 0, 0 }, 1 },
		{ "initial-touch", { 2, 0 }, { 0, 0 }, { 1, 0 }, { 1, 0 }, { 0, 0 }, 1 },
		{ "radius-touch", { 2.25f, 0 }, { 0, 0 }, { 1, 0 }, { 1, 0 }, { 0, 0 }, 1 },
		{ "max-fraction", { 4, 0 }, { 0, 0 }, { 1, 0 }, { 1, 0 }, { 0, 0 }, 0.25f },
		{ "rotating-hit", { 4, 0.5f }, { 0, 0 }, { 1, 0 }, { 0, 1 }, { 0, 0 }, 1 },
		{ "local-com", { 5, 0.5f }, { 0, 0 }, { 0.6f, -0.8f }, { 0.6f, 0.8f }, { 0.75f, -0.25f }, 1 },
	};
	for ( int i = 0; i < 8; ++i )
	{
		ORACLE_OPEN( ORACLE_KEY( "distance_public/toi/%s", cases[i].name ) );
		b2TOIInput input = {
			.proxyA = distance_public_box( 0.125f ),
			.proxyB = distance_public_box( 0.125f ),
			.sweepA = { .c1 = { -1, 2 }, .c2 = { -1, 2 }, .q1 = { 1, 0 }, .q2 = { 1, 0 } },
			.sweepB = { .localCenter = cases[i].localCenter,
						.c1 = b2Add( cases[i].start, (b2Vec2){ -1, 2 } ),
						.c2 = b2Add( cases[i].end, (b2Vec2){ -1, 2 } ),
						.q1 = cases[i].q1,
						.q2 = cases[i].q2 },
			.maxFraction = cases[i].max,
		};
		distance_public_toi( "output", b2TimeOfImpact( &input ) );
		ORACLE_CLOSE();
	}
}

static void rotating_toi_history( void )
{
	const b2Rot rotations[] = { { 1, 0 },  { 0.6f, 0.8f },	 { 0, 1 },	{ -0.6f, 0.8f },
								{ -1, 0 }, { -0.6f, -0.8f }, { 0, -1 }, { 0.6f, -0.8f } };
	const b2Vec2 ends[] = { { -3, -2 }, { 0, 0 }, { 3, 2 }, { -0.125f, 0.25f } };
	const b2Vec2 starts[] = { { 3, 0.125f }, { 0.125f, 3 }, { 2.25f, 2.25f }, { -3, -0.125f } };
	const b2Vec2 ap[] = { { -2, -0.125f }, { 2, -0.125f }, { 2, 0.125f }, { -2, 0.125f } };
	const b2Vec2 bp[] = { { -0.125f, -2 }, { 0.125f, -2 }, { 0.125f, 2 }, { -0.125f, 2 } };
	b2ShapeProxy a = b2MakeProxy( ap, 4, 0.025f );
	b2ShapeProxy b = b2MakeProxy( bp, 4, 0.025f );
	int32_t states[5] = { 0 };
	ORACLE_OPEN( "distance_public/toi-rotating-history" );
	for ( int i = 0; i < 8; ++i )
	{
		for ( int j = 0; j < 4; ++j )
		{
			for ( int k = 0; k < 4; ++k )
			{
				b2TOIInput input = {
					.proxyA = a,
					.proxyB = b,
					.sweepA = { .localCenter = { 0.25f, -0.125f },
								.c1 = { 0, 0 },
								.c2 = { 0.125f, -0.25f },
								.q1 = rotations[( i + 1 ) % 8],
								.q2 = rotations[( i + 2 ) % 8] },
					.sweepB = { .localCenter = { -0.125f, 0.25f },
								.c1 = starts[k],
								.c2 = ends[j],
								.q1 = rotations[( i + 1 ) % 8],
								.q2 = rotations[i] },
					.maxFraction = 1,
				};
				b2TOIOutput o = b2TimeOfImpact( &input );
				distance_public_toi( ORACLE_KEY( "%d/%d/%d", i, j, k ), o );
				states[o.state] += 1;
			}
		}
	}
	for ( int i = 0; i < 5; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "state-count/%d", i ), states[i] );
	}
	ORACLE_CLOSE();
}

static void pinned_rotating_toi_sample( void )
{
	const b2Vec2 ap[] = { { -16.25f, 44.75f }, { -15.75f, 44.75f }, { -15.75f, 45.25f }, { -16.25f, 45.25f } };
	const b2Vec2 bp[] = { { 0, -0.125f }, { 0, 0.125f } };
	b2TOIInput input = {
		.proxyA = b2MakeProxy( ap, 4, 0 ),
		.proxyB = b2MakeProxy( bp, 2, 0.0299999993f ),
		.sweepA = { .q1 = { 1, 0 }, .q2 = { 1, 0 } },
		.sweepB =
			{
				.c1 = { -15.8332710f, 45.3520279f },
				.c2 = { -15.8324337f, 45.3413048f },
				.q1 = { -0.540891349f, 0.841092527f },
				.q2 = { -0.457797021f, 0.889056742f },
			},
		.maxFraction = 1,
	};
	ORACLE_OPEN( "distance_public/toi-pinned-rotating-sample" );
	distance_public_toi( "output", b2TimeOfImpact( &input ) );
	ORACLE_CLOSE();
}

int main( void )
{
	segments();
	proxies();
	sweep_transforms();
	distances();
	cache_history();
	casts();
	toi_cases();
	rotating_toi_history();
	pinned_rotating_toi_sample();
	return 0;
}
