// SPDX-License-Identifier: MIT
#include "oracle.h"

#include "box2d/collision.h"

#include <float.h>

static void emit_solver( b2PlaneSolverResult result, const b2CollisionPlane* planes, int count )
{
	ORACLE_VEC2( "full/delta", result.delta );
	ORACLE_I32( "full/iterations", result.iterationCount );
	ORACLE_I32( "full/plane-count", count );
	for ( int i = 0; i < count; ++i )
		ORACLE_F32( ORACLE_KEY( "full/plane%d/push", i ), planes[i].push );
}

int main( void )
{
	ORACLE_OPEN( "mover/OffsetsLimitsAndClipPolicies" );
	struct
	{
		b2Vec2 target;
		int count;
		b2CollisionPlane planes[3];
	} cases[] = {
		{ { 0.75f, -1 }, 1, { { { { 0, 1 }, 0.25f }, FLT_MAX, 7, true } } },
		{ { 0.75f, 0.25f }, 1, { { { { 0, 1 }, -0.5f }, FLT_MAX, 7, true } } },
		{ { 0.75f, -1 }, 1, { { { { 0, 1 }, 0.25f }, 0, 7, true } } },
		{ { -1, -1 }, 1, { { { { 0.8f, 0.6f }, 0.15f }, 0.3f, 7, false } } },
		{ { -1, -1 }, 1, { { { { 0.8f, 0.6f }, -0.15f }, FLT_MAX, 7, true } } },
		{ { -1, -1 },
		  2,
		  {
			  { { { 0, 1 }, 0.1f }, 0.2f, 7, false },
			  { { { 0.8f, 0.6f }, -0.1f }, FLT_MAX, 7, true },
		  } },
		{ { -1, -1 },
		  3,
		  {
			  { { { 0, 1 }, 0.1f }, FLT_MAX, 7, true },
			  { { { 1, 0 }, -0.2f }, FLT_MAX, 7, true },
			  { { { 0.8f, 0.6f }, 0.5f }, 0.1f, 7, false },
		  } },
	};
	const b2Vec2 velocities[] = { { -2, -3 }, { 2, 3 }, { 1, -1 }, { -1, 1 }, { 0, 0 } };
	for ( int i = 0; i < (int)( sizeof( cases ) / sizeof( cases[0] ) ); ++i )
	{
		b2PlaneSolverResult result = b2SolvePlanes( cases[i].target, cases[i].planes, cases[i].count );
		emit_solver( result, cases[i].planes, cases[i].count );
		for ( int j = 0; j < 5; ++j )
			ORACLE_VEC2( "full/clip", b2ClipVector( velocities[j], cases[i].planes, cases[i].count ) );
	}
	b2CollisionPlane planes[] = {
		{ { { 0, 1 }, 0 }, 0, 0.25f, true },
		{ { { 1, 0 }, 0 }, 0, 0, true },
		{ { { 0.8f, 0.6f }, 0 }, 0, 0.25f, false },
	};
	for ( int policy = 0; policy < 4; ++policy )
	{
		ORACLE_VEC2( "full/clip", b2ClipVector( (b2Vec2){ -2, -3 }, planes, 3 ) );
		switch ( policy )
		{
			case 0:
				planes[1].push = 0.25f;
				break;
			case 1:
				planes[2].clipVelocity = true;
				break;
			case 2:
				planes[0].push = 0;
				break;
		}
	}
	ORACLE_CLOSE();
	return 0;
}
