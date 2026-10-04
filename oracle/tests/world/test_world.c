// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT
#include "oracle.h"

#include "test_macros.h"
#include "benchmarks.h"
#include "dynamic_tree.h"
#include "joint.h"
#include "physics_world.h"
#include "solver_set.h"

#include "box2d/box2d.h"
#include "box2d/collision.h"
#include "box2d/constants.h"
#include "box2d/math_functions.h"

#include <stdio.h>

// This is a simple example of building and running a simulation
// using Box2D. Here we create a large ground box and a small dynamic
// box.
// There are no graphics for this example. Box2D is meant to be used
// with your rendering engine in your game engine.
int HelloWorld( void )
{
	// Construct a world object, which will hold and simulate the rigid bodies.
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = (b2Vec2){ 0.0f, -10.0f };

	b2WorldId worldId = b2CreateWorld( &worldDef );
	ENSURE( b2World_IsValid( worldId ) );
	ORACLE_BOOL( "L30", b2World_IsValid( worldId ) );

	// Define the ground body.
	b2BodyDef groundBodyDef = b2DefaultBodyDef();
	groundBodyDef.position = (b2Pos){ 0.0f, -10.0f };

	// Call the body factory which allocates memory for the ground body
	// from a pool and creates the ground box shape (also from a pool).
	// The body is also added to the world.
	b2BodyId groundId = b2CreateBody( worldId, &groundBodyDef );
	ENSURE( b2Body_IsValid( groundId ) );
	ORACLE_BOOL( "L40", b2Body_IsValid( groundId ) );

	// Define the ground box shape. The extents are the half-widths of the box.
	b2Polygon groundBox = b2MakeBox( 50.0f, 10.0f );

	// Add the box shape to the ground body.
	b2ShapeDef groundShapeDef = b2DefaultShapeDef();
	b2CreatePolygonShape( groundId, &groundShapeDef, &groundBox );

	// Define the dynamic body. We set its position and call the body factory.
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	bodyDef.position = (b2Pos){ 0.0f, 4.0f };

	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	// Define another box shape for our dynamic body.
	b2Polygon dynamicBox = b2MakeBox( 1.0f, 1.0f );

	// Define the dynamic body shape
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	// Set the box density to be non-zero, so it will be dynamic.
	shapeDef.density = 1.0f;

	// Override the default friction.
	shapeDef.material.friction = 0.3f;

	// Add the shape to the body.
	b2CreatePolygonShape( bodyId, &shapeDef, &dynamicBox );

	// Prepare for simulation. Typically we use a time step of 1/60 of a
	// second (60Hz) and 4 sub-steps. This provides a high quality simulation
	// in most game scenarios.
	float timeStep = 1.0f / 60.0f;
	int subStepCount = 4;

	b2Pos position = b2Body_GetPosition( bodyId );
	b2Rot rotation = b2Body_GetRotation( bodyId );
		ORACLE_POS( "hello.position", position );
		ORACLE_ROT( "hello.rotation", rotation );

	// This is our little game loop.
	for ( int i = 0; i < 90; ++i )
	{
		// Instruct the world to perform a single step of simulation.
		// It is generally best to keep the time step and iterations fixed.
		b2World_Step( worldId, timeStep, subStepCount );
		ORACLE_WORLD_STEP( worldId, "L85" );

		// Now print the position and angle of the body.
		position = b2Body_GetPosition( bodyId );
		rotation = b2Body_GetRotation( bodyId );
		ORACLE_POS( "hello.position", position );
		ORACLE_ROT( "hello.rotation", rotation );

		// printf("%4.2f %4.2f %4.2f\n", position.x, position.y, b2Rot_GetAngle(rotation));
	}

	// When the world destructor is called, all bodies and joints are freed. This can
	// create orphaned ids, so be careful about your world management.
	b2DestroyWorld( worldId );

	ENSURE( b2AbsFloat( position.x ) < 0.01f );
	ORACLE_BOOL( "L98", b2AbsFloat( position.x ) < 0.01f );
	ORACLE_F32( "L98:value", b2AbsFloat( position.x ) );
	ENSURE( b2AbsFloat( position.y - 1.00f ) < 0.01f );
	ORACLE_BOOL( "L99", b2AbsFloat( position.y - 1.00f ) < 0.01f );
	ORACLE_F32( "L99:value", b2AbsFloat( position.y - 1.00f ) );
	ENSURE( b2AbsFloat( b2Rot_GetAngle( rotation ) ) < 0.01f );
	ORACLE_BOOL( "L100", b2AbsFloat( b2Rot_GetAngle( rotation ) ) < 0.01f );
	ORACLE_F32( "L100:value", b2AbsFloat( b2Rot_GetAngle( rotation ) ) );

	return 0;
}

int EmptyWorld( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );
	ENSURE( b2World_IsValid( worldId ) == true );
	ORACLE_BOOL( "L109", b2World_IsValid( worldId ) == true );

	float timeStep = 1.0f / 60.0f;
	int32_t subStepCount = 1;

	for ( int32_t i = 0; i < 60; ++i )
	{
		b2World_Step( worldId, timeStep, subStepCount );
		ORACLE_WORLD_STEP( worldId, "L116" );
	}

	b2DestroyWorld( worldId );

	ENSURE( b2World_IsValid( worldId ) == false );
	ORACLE_BOOL( "L121", b2World_IsValid( worldId ) == false );

	return 0;
}

#define BODY_COUNT 10
int DestroyAllBodiesWorld( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );
	ENSURE( b2World_IsValid( worldId ) == true );
	ORACLE_BOOL( "L131", b2World_IsValid( worldId ) == true );

	int count = 0;
	bool creating = true;

	b2BodyId bodyIds[BODY_COUNT];
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	b2Polygon square = b2MakeSquare( 0.5f );

	for ( int32_t i = 0; i < 2 * BODY_COUNT + 10; ++i )
	{
		if ( creating )
		{
			if ( count < BODY_COUNT )
			{
				bodyIds[count] = b2CreateBody( worldId, &bodyDef );

				b2ShapeDef shapeDef = b2DefaultShapeDef();
				b2CreatePolygonShape( bodyIds[count], &shapeDef, &square );
				count += 1;
			}
			else
			{
				creating = false;
			}
		}
		else if ( count > 0 )
		{
			b2DestroyBody( bodyIds[count - 1] );
			bodyIds[count - 1] = b2_nullBodyId;
			count -= 1;
		}

		b2World_Step( worldId, 1.0f / 60.0f, 3 );
		ORACLE_WORLD_STEP( worldId, "L165" );
	}

	b2Counters counters = b2World_GetCounters( worldId );
	ENSURE( counters.bodyCount == 0 );
	ORACLE_BOOL( "L169", counters.bodyCount == 0 );
	ORACLE_I32( "L169:value", counters.bodyCount );

	b2DestroyWorld( worldId );

	ENSURE( b2World_IsValid( worldId ) == false );
	ORACLE_BOOL( "L173", b2World_IsValid( worldId ) == false );

	return 0;
}

static int TestIsValid( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );
	ENSURE( b2World_IsValid( worldId ) );
	ORACLE_BOOL( "L182", b2World_IsValid( worldId ) );

	b2BodyDef bodyDef = b2DefaultBodyDef();

	b2BodyId bodyId1 = b2CreateBody( worldId, &bodyDef );
	ENSURE( b2Body_IsValid( bodyId1 ) == true );
	ORACLE_BOOL( "L187", b2Body_IsValid( bodyId1 ) == true );

	b2BodyId bodyId2 = b2CreateBody( worldId, &bodyDef );
	ENSURE( b2Body_IsValid( bodyId2 ) == true );
	ORACLE_BOOL( "L190", b2Body_IsValid( bodyId2 ) == true );

	b2DestroyBody( bodyId1 );
	ENSURE( b2Body_IsValid( bodyId1 ) == false );
	ORACLE_BOOL( "L193", b2Body_IsValid( bodyId1 ) == false );

	b2DestroyBody( bodyId2 );
	ENSURE( b2Body_IsValid( bodyId2 ) == false );
	ORACLE_BOOL( "L196", b2Body_IsValid( bodyId2 ) == false );

	b2DestroyWorld( worldId );

	ENSURE( b2World_IsValid( worldId ) == false );
	ORACLE_BOOL( "L200", b2World_IsValid( worldId ) == false );
	ENSURE( b2Body_IsValid( bodyId2 ) == false );
	ORACLE_BOOL( "L201", b2Body_IsValid( bodyId2 ) == false );
	ENSURE( b2Body_IsValid( bodyId1 ) == false );
	ORACLE_BOOL( "L202", b2Body_IsValid( bodyId1 ) == false );

	return 0;
}

#define WORLD_COUNT ( B2_MAX_WORLDS / 2 )

int TestWorldRecycle( void )
{
	_Static_assert( WORLD_COUNT > 0, "world count" );

	int count = 100;

	b2WorldId worldIds[WORLD_COUNT];

	for ( int i = 0; i < count; ++i )
	{
		b2WorldDef worldDef = b2DefaultWorldDef();
		for ( int j = 0; j < WORLD_COUNT; ++j )
		{
			worldIds[j] = b2CreateWorld( &worldDef );
			ENSURE( b2World_IsValid( worldIds[j] ) == true );
			ORACLE_BOOL( "L223", b2World_IsValid( worldIds[j] ) == true );

			b2BodyDef bodyDef = b2DefaultBodyDef();
			b2CreateBody( worldIds[j], &bodyDef );
		}

		for ( int j = 0; j < WORLD_COUNT; ++j )
		{
			float timeStep = 1.0f / 60.0f;
			int subStepCount = 1;

			for ( int k = 0; k < 10; ++k )
			{
				b2World_Step( worldIds[j], timeStep, subStepCount );
				ORACLE_WORLD_STEP( worldIds[j], "L236" );
			}
		}

		for ( int j = WORLD_COUNT - 1; j >= 0; --j )
		{
			b2DestroyWorld( worldIds[j] );
			ENSURE( b2World_IsValid( worldIds[j] ) == false );
			ORACLE_BOOL( "L243", b2World_IsValid( worldIds[j] ) == false );
			worldIds[j] = b2_nullWorldId;
		}
	}

	return 0;
}

static bool CustomFilter( b2ShapeId shapeIdA, b2ShapeId shapeIdB, void* context )
{
	(void)shapeIdA;
	(void)shapeIdB;
	ENSURE( context == NULL );
	ORACLE_BOOL( "L255", context == NULL );
	return true;
}

static void PreSolveStatic( b2ShapeId shapeIdA, b2ShapeId shapeIdB, b2Manifold* manifold, void* context )
{
	(void)shapeIdA;
	(void)shapeIdB;
	(void)manifold;
	(void)context;
}

// This test is here to ensure all API functions link correctly.
int TestWorldCoverage( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();

	b2WorldId worldId = b2CreateWorld( &worldDef );
	ENSURE( b2World_IsValid( worldId ) );
	ORACLE_BOOL( "L273", b2World_IsValid( worldId ) );

	b2World_EnableSleeping( worldId, true );
	b2World_EnableSleeping( worldId, false );
	bool flag = b2World_IsSleepingEnabled( worldId );
	ENSURE( flag == false );
	ORACLE_BOOL( "L278", flag == false );

	b2World_EnableContinuous( worldId, false );
	b2World_EnableContinuous( worldId, true );
	flag = b2World_IsContinuousEnabled( worldId );
	ENSURE( flag == true );
	ORACLE_BOOL( "L283", flag == true );

	b2World_SetRestitutionThreshold( worldId, 0.0f );
	b2World_SetRestitutionThreshold( worldId, 2.0f );
	float value = b2World_GetRestitutionThreshold( worldId );
	ENSURE( value == 2.0f );
	ORACLE_BOOL( "L288", value == 2.0f );
	ORACLE_F32( "L288:value", value );

	b2World_SetHitEventThreshold( worldId, 0.0f );
	b2World_SetHitEventThreshold( worldId, 100.0f );
	value = b2World_GetHitEventThreshold( worldId );
	ENSURE( value == 100.0f );
	ORACLE_BOOL( "L293", value == 100.0f );
	ORACLE_F32( "L293:value", value );

	b2World_SetCustomFilterCallback( worldId, CustomFilter, NULL );
	b2World_SetPreSolveCallback( worldId, PreSolveStatic, NULL, NULL );

	b2Vec2 g = { 1.0f, 2.0f };
	b2World_SetGravity( worldId, g );
	b2Vec2 v = b2World_GetGravity( worldId );
	ENSURE( v.x == g.x );
	ORACLE_BOOL( "L301", v.x == g.x );
	ORACLE_F32( "L301:value", v.x );
	ENSURE( v.y == g.y );
	ORACLE_BOOL( "L302", v.y == g.y );
	ORACLE_F32( "L302:value", v.y );

	b2ExplosionDef explosionDef = b2DefaultExplosionDef();
	b2World_Explode( worldId, &explosionDef );

	b2World_SetContactTuning( worldId, 10.0f, 2.0f, 4.0f );

	b2World_SetMaximumLinearSpeed( worldId, 10.0f );
	value = b2World_GetMaximumLinearSpeed( worldId );
	ENSURE( value == 10.0f );
	ORACLE_BOOL( "L311", value == 10.0f );
	ORACLE_F32( "L311:value", value );

	b2World_EnableWarmStarting( worldId, true );
	flag = b2World_IsWarmStartingEnabled( worldId );
	ENSURE( flag == true );
	ORACLE_BOOL( "L315", flag == true );

	int count = b2World_GetAwakeBodyCount( worldId );
	ENSURE( count == 0 );
	ORACLE_BOOL( "L318", count == 0 );
	ORACLE_I32( "L318:value", count );

	b2World_SetUserData( worldId, &value );
	void* userData = b2World_GetUserData( worldId );
	ENSURE( userData == &value );
	ORACLE_BOOL( "L322", userData == &value );

	b2World_Step( worldId, 1.0f, 1 );
	ORACLE_WORLD_STEP( worldId, "L324" );

	b2DestroyWorld( worldId );

	return 0;
}

static int TestSensor( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	// Wall from x = 1 to x = 2
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_staticBody;
	bodyDef.position.x = 1.5f;
	bodyDef.position.y = 11.0f;
	b2BodyId wallId = b2CreateBody( worldId, &bodyDef );
	b2Polygon box = b2MakeBox( 0.5f, 10.0f );
	b2ShapeDef shapeDef = b2DefaultShapeDef();
	shapeDef.enableSensorEvents = true;
	b2CreatePolygonShape( wallId, &shapeDef, &box );

	// Bullet fired towards the wall
	bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	bodyDef.isBullet = true;
	bodyDef.gravityScale = 0.0f;
	bodyDef.position = (b2Pos){ 7.39814f, 4.0f };
	bodyDef.linearVelocity = (b2Vec2){ -20.0f, 0.0f };
	b2BodyId bulletId = b2CreateBody( worldId, &bodyDef );
	shapeDef = b2DefaultShapeDef();
	shapeDef.isSensor = true;
	shapeDef.enableSensorEvents = true;
	b2Circle circle = { { 0.0f, 0.0f }, 0.1f };
	b2CreateCircleShape( bulletId, &shapeDef, &circle );

	int beginCount = 0;
	int endCount = 0;

	while ( true )
	{
		float timeStep = 1.0f / 60.0f;
		int subStepCount = 4;
		b2World_Step( worldId, timeStep, subStepCount );
		ORACLE_WORLD_STEP( worldId, "L368" );

		b2Pos bulletPos = b2Body_GetPosition( bulletId );
		// printf( "Bullet pos: %g %g\n", bulletPos.x, bulletPos.y );

		b2SensorEvents events = b2World_GetSensorEvents( worldId );
		ORACLE_WORLD_SENSOR( events );
		ORACLE_POS( "sensor.position", bulletPos );

		if ( events.beginCount > 0 )
		{
			beginCount += 1;
		}

		if ( events.endCount > 0 )
		{
			endCount += 1;
		}

		if ( bulletPos.x < -1.0f )
		{
			break;
		}
	}

	b2DestroyWorld( worldId );

	ENSURE( beginCount == 1 );
	ORACLE_BOOL( "L393", beginCount == 1 );
	ORACLE_I32( "L393:value", beginCount );
	ENSURE( endCount == 1 );
	ORACLE_BOOL( "L394", endCount == 1 );
	ORACLE_I32( "L394:value", endCount );

	return 0;
}

static int TestSetWorkerCount( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.workerCount = 1;
	b2WorldId worldId = b2CreateWorld( &worldDef );
	ENSURE( b2World_IsValid( worldId ) );
	ORACLE_BOOL( "L404", b2World_IsValid( worldId ) );
	ENSURE( b2World_GetWorkerCount( worldId ) == 1 );
	ORACLE_BOOL( "L405", b2World_GetWorkerCount( worldId ) == 1 );
	ORACLE_I32( "L405:value", b2World_GetWorkerCount( worldId ) );

	CreateJunkyard( worldId );
	StepJunkyard( worldId, 1 );
	ORACLE_WORLD_STEP( worldId, "target/L408" );

	b2World_SetWorkerCount( worldId, 4 );
	ENSURE( b2World_GetWorkerCount( worldId ) == 4 );
	ORACLE_BOOL( "L411", b2World_GetWorkerCount( worldId ) == 4 );
	ORACLE_I32( "L411:value", b2World_GetWorkerCount( worldId ) );

	StepJunkyard( worldId, 2 );
	ORACLE_WORLD_STEP( worldId, "target/L413" );

	b2World_SetWorkerCount( worldId, 4 );
	ENSURE( b2World_GetWorkerCount( worldId ) == 4 );
	ORACLE_BOOL( "L416", b2World_GetWorkerCount( worldId ) == 4 );
	ORACLE_I32( "L416:value", b2World_GetWorkerCount( worldId ) );

	StepJunkyard( worldId, 3 );
	ORACLE_WORLD_STEP( worldId, "target/L418" );

	b2World_SetWorkerCount( worldId, 0 );
	ENSURE( b2World_GetWorkerCount( worldId ) == 1 );
	ORACLE_BOOL( "L421", b2World_GetWorkerCount( worldId ) == 1 );
	ORACLE_I32( "L421:value", b2World_GetWorkerCount( worldId ) );

	StepJunkyard( worldId, 4 );
	ORACLE_WORLD_STEP( worldId, "target/L423" );

	b2World_SetWorkerCount( worldId, -5 );
	ENSURE( b2World_GetWorkerCount( worldId ) == 1 );
	ORACLE_BOOL( "L426", b2World_GetWorkerCount( worldId ) == 1 );
	ORACLE_I32( "L426:value", b2World_GetWorkerCount( worldId ) );

	StepJunkyard( worldId, 5 );
	ORACLE_WORLD_STEP( worldId, "target/L428" );

	b2World_SetWorkerCount( worldId, B2_MAX_WORKERS + 10 );
	ENSURE( b2World_GetWorkerCount( worldId ) == B2_MAX_WORKERS );
	ORACLE_BOOL( "L431", b2World_GetWorkerCount( worldId ) == B2_MAX_WORKERS );
	ORACLE_I32( "L431:value", b2World_GetWorkerCount( worldId ) );

	StepJunkyard( worldId, 2 );
	ORACLE_WORLD_STEP( worldId, "target/L433" );

	b2DestroyWorld( worldId );

	return 0;
}

static int ChainSegmentShapeTest( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = (b2Vec2){ 0.0f, -10.0f };
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	b2BodyId groundId = b2CreateBody( worldId, &bodyDef );

	b2ChainSegment cs = { 0 };
	cs.ghost1 = (b2Vec2){ 2.0f, 0.0f };
	cs.segment.point1 = (b2Vec2){ 1.0f, 0.0f };
	cs.segment.point2 = (b2Vec2){ -1.0f, 0.0f };
	cs.ghost2 = (b2Vec2){ -2.0f, 0.0f };
	cs.chainId = 99;

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2ShapeId orphanShape = b2CreateChainSegmentShape( groundId, &shapeDef, &cs );
	ENSURE( B2_IS_NON_NULL( orphanShape ) );
	ORACLE_BOOL( "L458", B2_IS_NON_NULL( orphanShape ) );

	ENSURE( b2Shape_GetType( orphanShape ) == b2_chainSegmentShape );
	ORACLE_BOOL( "L460", b2Shape_GetType( orphanShape ) == b2_chainSegmentShape );
	ORACLE_I32( "L460:value", b2Shape_GetType( orphanShape ) );

	b2ChainId parentChain = b2Shape_GetParentChain( orphanShape );
	ENSURE( B2_IS_NULL( parentChain ) );
	ORACLE_BOOL( "L463", B2_IS_NULL( parentChain ) );

	b2ChainSegment got = b2Shape_GetChainSegment( orphanShape );
	ENSURE_SMALL( got.ghost1.x - cs.ghost1.x, 1e-5f );
	ORACLE_F32( "L466", got.ghost1.x - cs.ghost1.x );
	ENSURE_SMALL( got.ghost1.y - cs.ghost1.y, 1e-5f );
	ORACLE_F32( "L467", got.ghost1.y - cs.ghost1.y );
	ENSURE_SMALL( got.segment.point1.x - cs.segment.point1.x, 1e-5f );
	ORACLE_F32( "L468", got.segment.point1.x - cs.segment.point1.x );
	ENSURE_SMALL( got.segment.point1.y - cs.segment.point1.y, 1e-5f );
	ORACLE_F32( "L469", got.segment.point1.y - cs.segment.point1.y );
	ENSURE_SMALL( got.segment.point2.x - cs.segment.point2.x, 1e-5f );
	ORACLE_F32( "L470", got.segment.point2.x - cs.segment.point2.x );
	ENSURE_SMALL( got.segment.point2.y - cs.segment.point2.y, 1e-5f );
	ORACLE_F32( "L471", got.segment.point2.y - cs.segment.point2.y );
	ENSURE_SMALL( got.ghost2.x - cs.ghost2.x, 1e-5f );
	ORACLE_F32( "L472", got.ghost2.x - cs.ghost2.x );
	ENSURE_SMALL( got.ghost2.y - cs.ghost2.y, 1e-5f );
	ORACLE_F32( "L473", got.ghost2.y - cs.ghost2.y );
	ENSURE( got.chainId == B2_NULL_INDEX );
	ORACLE_BOOL( "L474", got.chainId == B2_NULL_INDEX );
	ORACLE_I32( "L474:value", got.chainId );

	b2BodyDef dynamicDef = b2DefaultBodyDef();
	dynamicDef.type = b2_dynamicBody;
	dynamicDef.position = (b2Pos){ 0.0f, 2.0f };
	b2BodyId circleBodyId = b2CreateBody( worldId, &dynamicDef );
	b2Circle circle = { { 0.0f, 0.0f }, 0.5f };
	b2ShapeDef circleShapeDef = b2DefaultShapeDef();
	b2CreateCircleShape( circleBodyId, &circleShapeDef, &circle );

	for ( int i = 0; i < 120; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_WORLD_STEP( worldId, "L486" );
	}

	b2Pos circlePos = b2Body_GetPosition( circleBodyId );
	ENSURE( circlePos.y > 0.0f );
	ORACLE_BOOL( "L490", circlePos.y > 0.0f );
	ORACLE_POS( "L490:value", circlePos );

	b2ChainSegment cs2 = { 0 };
	cs2.ghost1 = (b2Vec2){ 3.0f, 0.0f };
	cs2.segment.point1 = (b2Vec2){ 2.0f, 0.0f };
	cs2.segment.point2 = (b2Vec2){ -2.0f, 0.0f };
	cs2.ghost2 = (b2Vec2){ -3.0f, 0.0f };
	cs2.chainId = B2_NULL_INDEX;

	b2Shape_SetChainSegment( orphanShape, &cs2 );

	b2ChainSegment got2 = b2Shape_GetChainSegment( orphanShape );
	ENSURE_SMALL( got2.segment.point1.x - cs2.segment.point1.x, 1e-5f );
	ORACLE_F32( "L502", got2.segment.point1.x - cs2.segment.point1.x );
	ENSURE_SMALL( got2.segment.point1.y - cs2.segment.point1.y, 1e-5f );
	ORACLE_F32( "L503", got2.segment.point1.y - cs2.segment.point1.y );
	ENSURE_SMALL( got2.segment.point2.x - cs2.segment.point2.x, 1e-5f );
	ORACLE_F32( "L504", got2.segment.point2.x - cs2.segment.point2.x );
	ENSURE_SMALL( got2.segment.point2.y - cs2.segment.point2.y, 1e-5f );
	ORACLE_F32( "L505", got2.segment.point2.y - cs2.segment.point2.y );
	ENSURE_SMALL( got2.ghost1.x - cs2.ghost1.x, 1e-5f );
	ORACLE_F32( "L506", got2.ghost1.x - cs2.ghost1.x );
	ENSURE_SMALL( got2.ghost1.y - cs2.ghost1.y, 1e-5f );
	ORACLE_F32( "L507", got2.ghost1.y - cs2.ghost1.y );
	ENSURE_SMALL( got2.ghost2.x - cs2.ghost2.x, 1e-5f );
	ORACLE_F32( "L508", got2.ghost2.x - cs2.ghost2.x );
	ENSURE_SMALL( got2.ghost2.y - cs2.ghost2.y, 1e-5f );
	ORACLE_F32( "L509", got2.ghost2.y - cs2.ghost2.y );
	ENSURE( got2.chainId == B2_NULL_INDEX );
	ORACLE_BOOL( "L510", got2.chainId == B2_NULL_INDEX );
	ORACLE_I32( "L510:value", got2.chainId );

	b2ChainId parentChain2 = b2Shape_GetParentChain( orphanShape );
	ENSURE( B2_IS_NULL( parentChain2 ) );
	ORACLE_BOOL( "L513", B2_IS_NULL( parentChain2 ) );

	b2BodyId convBody = b2CreateBody( worldId, &bodyDef );
	b2Circle convCircle = { { 0.0f, 0.0f }, 0.25f };
	b2ShapeId convShape = b2CreateCircleShape( convBody, &shapeDef, &convCircle );
	ENSURE( b2Shape_GetType( convShape ) == b2_circleShape );
	ORACLE_BOOL( "L518", b2Shape_GetType( convShape ) == b2_circleShape );
	ORACLE_I32( "L518:value", b2Shape_GetType( convShape ) );

	b2Shape_SetChainSegment( convShape, &cs2 );

	ENSURE( b2Shape_GetType( convShape ) == b2_chainSegmentShape );
	ORACLE_BOOL( "L522", b2Shape_GetType( convShape ) == b2_chainSegmentShape );
	ORACLE_I32( "L522:value", b2Shape_GetType( convShape ) );
	b2ChainSegment got3 = b2Shape_GetChainSegment( convShape );
	ENSURE_SMALL( got3.ghost1.x - cs2.ghost1.x, 1e-5f );
	ORACLE_F32( "L524", got3.ghost1.x - cs2.ghost1.x );
	ENSURE_SMALL( got3.ghost1.y - cs2.ghost1.y, 1e-5f );
	ORACLE_F32( "L525", got3.ghost1.y - cs2.ghost1.y );
	ENSURE_SMALL( got3.ghost2.x - cs2.ghost2.x, 1e-5f );
	ORACLE_F32( "L526", got3.ghost2.x - cs2.ghost2.x );
	ENSURE_SMALL( got3.ghost2.y - cs2.ghost2.y, 1e-5f );
	ORACLE_F32( "L527", got3.ghost2.y - cs2.ghost2.y );
	ENSURE( got3.chainId == B2_NULL_INDEX );
	ORACLE_BOOL( "L528", got3.chainId == B2_NULL_INDEX );
	ORACLE_I32( "L528:value", got3.chainId );

	b2ChainId parentChain3 = b2Shape_GetParentChain( convShape );
	ENSURE( B2_IS_NULL( parentChain3 ) );
	ORACLE_BOOL( "L531", B2_IS_NULL( parentChain3 ) );

	b2DestroyShape( orphanShape, true );
	b2DestroyWorld( worldId );

	return 0;
}

static int DeferredMassFlagSyncTest( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	shapeDef.updateBodyMass = false;

	b2Circle circle = { { 0.0f, 0.0f }, 0.5f };
	b2CreateCircleShape( bodyId, &shapeDef, &circle );

	b2Body_UpdateMassFromShapes( bodyId );

	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L556" );

	b2DestroyWorld( worldId );
	return 0;
}

static int EnableSleepFlagSyncTest( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	bodyDef.enableSleep = false;
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	ENSURE( b2Body_IsSleepEnabled( bodyId ) == false );
	ORACLE_BOOL( "L572", b2Body_IsSleepEnabled( bodyId ) == false );

	b2Body_EnableSleep( bodyId, true );
	ENSURE( b2Body_IsSleepEnabled( bodyId ) == true );
	ORACLE_BOOL( "L575", b2Body_IsSleepEnabled( bodyId ) == true );

	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L577" );

	b2DestroyWorld( worldId );
	return 0;
}

static int EnableContactRecyclingTest( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;

	// Default is enabled
	b2BodyId bodyA = b2CreateBody( worldId, &bodyDef );
	ENSURE( b2Body_IsContactRecyclingEnabled( bodyA ) == true );
	ORACLE_BOOL( "L593", b2Body_IsContactRecyclingEnabled( bodyA ) == true );

	b2Body_EnableContactRecycling( bodyA, false );
	ENSURE( b2Body_IsContactRecyclingEnabled( bodyA ) == false );
	ORACLE_BOOL( "L596", b2Body_IsContactRecyclingEnabled( bodyA ) == false );

	b2Body_EnableContactRecycling( bodyA, true );
	ENSURE( b2Body_IsContactRecyclingEnabled( bodyA ) == true );
	ORACLE_BOOL( "L599", b2Body_IsContactRecyclingEnabled( bodyA ) == true );

	// Per-def opt-out at creation
	bodyDef.enableContactRecycling = false;
	b2BodyId bodyB = b2CreateBody( worldId, &bodyDef );
	ENSURE( b2Body_IsContactRecyclingEnabled( bodyB ) == false );
	ORACLE_BOOL( "L604", b2Body_IsContactRecyclingEnabled( bodyB ) == false );

	// Stepping after toggling must not trip the flag-sync validator
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L607" );

	b2DestroyWorld( worldId );
	return 0;
}

static int SetBulletDriftTest( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	{
		b2BodyDef bodyDef = b2DefaultBodyDef();
		bodyDef.type = b2_dynamicBody;
		bodyDef.isBullet = false;
		b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

		ENSURE( b2Body_IsBullet( bodyId ) == false );
		ORACLE_BOOL( "L624", b2Body_IsBullet( bodyId ) == false );

		b2Body_SetBullet( bodyId, true );
		ENSURE( b2Body_IsBullet( bodyId ) == true );
		ORACLE_BOOL( "L627", b2Body_IsBullet( bodyId ) == true );

		b2MotionLocks locks = { 0 };
		locks.linearX = true;
		b2Body_SetMotionLocks( bodyId, locks );

		ENSURE( b2Body_IsBullet( bodyId ) == true );
		ORACLE_BOOL( "L633", b2Body_IsBullet( bodyId ) == true );
	}

	{
		b2BodyDef bodyDef = b2DefaultBodyDef();
		bodyDef.type = b2_dynamicBody;
		bodyDef.isBullet = true;
		b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

		ENSURE( b2Body_IsBullet( bodyId ) == true );
		ORACLE_BOOL( "L642", b2Body_IsBullet( bodyId ) == true );

		b2Body_SetBullet( bodyId, false );
		ENSURE( b2Body_IsBullet( bodyId ) == false );
		ORACLE_BOOL( "L645", b2Body_IsBullet( bodyId ) == false );

		b2MotionLocks locks = { 0 };
		locks.linearX = true;
		b2Body_SetMotionLocks( bodyId, locks );

		ENSURE( b2Body_IsBullet( bodyId ) == false );
		ORACLE_BOOL( "L651", b2Body_IsBullet( bodyId ) == false );
	}

	b2DestroyWorld( worldId );
	return 0;
}

// Counts marked nodes, the root included. Free nodes are never marked.
static int CountEnlargedNodes( const b2DynamicTree* tree )
{
	int count = 0;
	for ( int i = 0; i < tree->nodeEnd; ++i )
	{
		count += b2IsNodeMoved( tree->nodes + i ) ? 1 : 0;
	}

	return count;
}

// A proxy can be enlarged in one step and destroyed before the next. The remove recomputes the
// ancestor marks from the surviving children, so the mover's marks go with it and the next step
// starts clean.
static int EnlargedProxyDestroyed( int workerCount )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = b2Vec2_zero;
	worldDef.workerCount = workerCount;
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2World* world = b2GetWorldFromId( worldId );
	const b2DynamicTree* tree = world->broadPhase.trees + b2_dynamicBody;

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Circle circle = { b2Vec2_zero, 0.25f };

	// Resting bodies keep the tree deep enough that the mover has an ancestor that outlives it
	for ( int i = 0; i < 8; ++i )
	{
		bodyDef.position = (b2Pos){ 2.0f * i, 0.0f };
		b2BodyId restingId = b2CreateBody( worldId, &bodyDef );
		b2CreateCircleShape( restingId, &shapeDef, &circle );
	}

	// This body outruns its AABB margin every step
	bodyDef.position = (b2Pos){ 7.0f, 1.0f };
	bodyDef.linearVelocity = (b2Vec2){ 6.0f, 0.0f };
	b2BodyId moverId = b2CreateBody( worldId, &bodyDef );
	b2CreateCircleShape( moverId, &shapeDef, &circle );

	float timeStep = 1.0f / 60.0f;
	b2World_Step( worldId, timeStep, 4 );
	ORACLE_WORLD_STEP( worldId, "L704" );

	ENSURE( CountEnlargedNodes( tree ) > 0 );
	ORACLE_BOOL( "L706", CountEnlargedNodes( tree ) > 0 );
	ORACLE_I32( "L706:value", CountEnlargedNodes( tree ) );

	b2DestroyBody( moverId );

	ENSURE( CountEnlargedNodes( tree ) == 0 );
	ORACLE_BOOL( "L710", CountEnlargedNodes( tree ) == 0 );
	ORACLE_I32( "L710:value", CountEnlargedNodes( tree ) );

	b2World_Step( worldId, timeStep, 4 );
	ORACLE_WORLD_STEP( worldId, "L712" );

	ENSURE( CountEnlargedNodes( tree ) == 0 );
	ORACLE_BOOL( "L714", CountEnlargedNodes( tree ) == 0 );
	ORACLE_I32( "L714:value", CountEnlargedNodes( tree ) );

	b2DestroyWorld( worldId );
	return 0;
}

static int EnlargedProxyDestroyedTest( void )
{
	ENSURE( EnlargedProxyDestroyed( 1 ) == 0 );
	ORACLE_BOOL( "L722", true );
	ENSURE( EnlargedProxyDestroyed( 4 ) == 0 );
	ORACLE_BOOL( "L723", true );

	return 0;
}

// Broad phase pair creation through every route that marks a proxy: creation, teleport, a static
// shape that asks for contacts, refiltering, collide connected, enable, a type change, a static
// rebuild that must keep its pending marks, and the filters that reject a pair.
// Bodies never move on their own here, gravity is off and sleep is disabled, so a contact can only
// come from the route under test. The count is every live contact, touching or not, which is what
// the broad phase decides.

typedef struct PairBox
{
	b2BodyId bodyId;
	b2ShapeId shapeId;
} PairBox;

static PairBox CreatePairBox( b2WorldId worldId, b2BodyType type, float x, float y, const b2ShapeDef* shapeDef )
{
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = type;
	bodyDef.position = (b2Pos){ x, y };

	PairBox box;
	box.bodyId = b2CreateBody( worldId, &bodyDef );

	b2Polygon square = b2MakeSquare( 0.5f );
	box.shapeId = b2CreatePolygonShape( box.bodyId, shapeDef, &square );
	return box;
}

static b2WorldId CreatePairWorld( int workerCount )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = b2Vec2_zero;
	worldDef.enableSleep = false;
	worldDef.workerCount = workerCount;
	return b2CreateWorld( &worldDef );
}

static int PairContactCount( b2WorldId worldId )
{
	return b2World_GetCounters( worldId ).contactCount;
}

static void StepPairWorld( b2WorldId worldId )
{
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L771" );
}

// Every kind of proxy created over a resting body
static int PairCreation( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	// Dynamic created on static
	CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &shapeDef );
	CreatePairBox( worldId, b2_dynamicBody, 0.5f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L784", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L784:value", PairContactCount( worldId ) );

	// Dynamic created on a resting dynamic
	CreatePairBox( worldId, b2_dynamicBody, 10.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L789", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L789:value", PairContactCount( worldId ) );
	CreatePairBox( worldId, b2_dynamicBody, 10.5f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L792", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L792:value", PairContactCount( worldId ) );

	// Kinematic created on a resting dynamic, the mark is on the kinematic side
	CreatePairBox( worldId, b2_dynamicBody, 20.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	CreatePairBox( worldId, b2_kinematicBody, 20.5f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 3 );
	ORACLE_BOOL( "L799", PairContactCount( worldId ) == 3 );
	ORACLE_I32( "L799:value", PairContactCount( worldId ) );

	// A static shape created on a resting dynamic pairs by default and not when told to stay quiet
	b2ShapeDef quietDef = b2DefaultShapeDef();
	quietDef.invokeContactCreation = false;
	CreatePairBox( worldId, b2_dynamicBody, 30.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	CreatePairBox( worldId, b2_staticBody, 30.5f, 0.0f, &quietDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 3 );
	ORACLE_BOOL( "L808", PairContactCount( worldId ) == 3 );
	ORACLE_I32( "L808:value", PairContactCount( worldId ) );

	CreatePairBox( worldId, b2_staticBody, 29.5f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 4 );
	ORACLE_BOOL( "L812", PairContactCount( worldId ) == 4 );
	ORACLE_I32( "L812:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// A teleport marks the moved proxy, kinematic and static ones included
static int PairTeleport( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &shapeDef );
	PairBox mover = CreatePairBox( worldId, b2_dynamicBody, 10.0f, 0.0f, &shapeDef );
	CreatePairBox( worldId, b2_dynamicBody, 20.0f, 0.0f, &shapeDef );
	PairBox wall = CreatePairBox( worldId, b2_staticBody, 30.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L829", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L829:value", PairContactCount( worldId ) );

	b2Body_SetTransform( mover.bodyId, (b2Pos){ 0.5f, 0.0f }, b2Rot_identity );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L833", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L833:value", PairContactCount( worldId ) );

	b2Body_SetTransform( wall.bodyId, (b2Pos){ 20.5f, 0.0f }, b2Rot_identity );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L837", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L837:value", PairContactCount( worldId ) );

	// Kinematic onto a resting dynamic, the mark is on the kinematic side
	CreatePairBox( worldId, b2_dynamicBody, 40.0f, 0.0f, &shapeDef );
	PairBox pusher = CreatePairBox( worldId, b2_kinematicBody, 50.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L843", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L843:value", PairContactCount( worldId ) );

	b2Body_SetTransform( pusher.bodyId, (b2Pos){ 40.5f, 0.0f }, b2Rot_identity );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 3 );
	ORACLE_BOOL( "L847", PairContactCount( worldId ) == 3 );
	ORACLE_I32( "L847:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// A filter change must find the pairs the old filter rejected and drop the ones it now rejects
static int PairRefilter( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );

	b2ShapeDef pickyDef = b2DefaultShapeDef();
	pickyDef.filter.categoryBits = 2;
	pickyDef.filter.maskBits = 1;

	b2ShapeDef closedDef = b2DefaultShapeDef();
	closedDef.filter.categoryBits = 4;

	b2Filter openFilter = b2DefaultFilter();
	openFilter.categoryBits = 1;

	PairBox ground = CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &closedDef );
	CreatePairBox( worldId, b2_dynamicBody, 0.5f, 0.0f, &pickyDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L871", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L871:value", PairContactCount( worldId ) );

	b2Shape_SetFilter( ground.shapeId, openFilter );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L875", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L875:value", PairContactCount( worldId ) );

	b2Shape_SetFilter( ground.shapeId, closedDef.filter );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L879", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L879:value", PairContactCount( worldId ) );

	// Same on a dynamic shape
	PairBox other = CreatePairBox( worldId, b2_dynamicBody, 10.0f, 0.0f, &closedDef );
	CreatePairBox( worldId, b2_dynamicBody, 10.5f, 0.0f, &pickyDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L885", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L885:value", PairContactCount( worldId ) );

	b2Shape_SetFilter( other.shapeId, openFilter );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L889", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L889:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// Collide connected is off by default and turning it on must find the pair, static partner too.
// The weld frames coincide so the joint does not move anything.
static int PairCollideConnected( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	PairBox boxA = CreatePairBox( worldId, b2_dynamicBody, 0.0f, 0.0f, &shapeDef );
	PairBox boxB = CreatePairBox( worldId, b2_dynamicBody, 0.5f, 0.0f, &shapeDef );

	b2WeldJointDef jointDef = b2DefaultWeldJointDef();
	jointDef.base.bodyIdA = boxA.bodyId;
	jointDef.base.bodyIdB = boxB.bodyId;
	jointDef.base.localFrameB.p = (b2Vec2){ -0.5f, 0.0f };
	b2JointId jointId = b2CreateWeldJoint( worldId, &jointDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L911", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L911:value", PairContactCount( worldId ) );

	b2Joint_SetCollideConnected( jointId, true );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L915", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L915:value", PairContactCount( worldId ) );

	b2Joint_SetCollideConnected( jointId, false );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L919", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L919:value", PairContactCount( worldId ) );

	// Static partner, the mark goes on the dynamic body
	PairBox wall = CreatePairBox( worldId, b2_staticBody, 10.0f, 0.0f, &shapeDef );
	PairBox boxC = CreatePairBox( worldId, b2_dynamicBody, 10.5f, 0.0f, &shapeDef );
	jointDef.base.bodyIdA = wall.bodyId;
	jointDef.base.bodyIdB = boxC.bodyId;
	b2JointId wallJointId = b2CreateWeldJoint( worldId, &jointDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L928", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L928:value", PairContactCount( worldId ) );

	b2Joint_SetCollideConnected( wallJointId, true );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L932", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L932:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

static bool RejectTaggedShape( b2ShapeId shapeIdA, b2ShapeId shapeIdB, void* context )
{
	B2_UNUSED( context );
	return b2Shape_GetUserData( shapeIdA ) == NULL && b2Shape_GetUserData( shapeIdB ) == NULL;
}

// The filters that reject a pair. The benchmarks never exercise these, so this is the only guard.
static int PairFilters( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2World_SetCustomFilterCallback( worldId, RejectTaggedShape, NULL );

	// Category and mask must pass in both directions
	b2ShapeDef defA = b2DefaultShapeDef();
	defA.filter.categoryBits = 1;
	defA.filter.maskBits = 2;
	b2ShapeDef defB = b2DefaultShapeDef();
	defB.filter.categoryBits = 2;
	defB.filter.maskBits = 1;
	b2ShapeDef defC = b2DefaultShapeDef();
	defC.filter.categoryBits = 2;
	defC.filter.maskBits = 4;
	CreatePairBox( worldId, b2_dynamicBody, 0.0f, 0.0f, &defA );
	CreatePairBox( worldId, b2_dynamicBody, 0.5f, 0.0f, &defB );
	CreatePairBox( worldId, b2_dynamicBody, -0.5f, 0.0f, &defC );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L964", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L964:value", PairContactCount( worldId ) );

	// Group index overrides the bits, negative rejects and positive accepts
	b2ShapeDef negativeDef = b2DefaultShapeDef();
	negativeDef.filter.groupIndex = -1;
	CreatePairBox( worldId, b2_dynamicBody, 10.0f, 0.0f, &negativeDef );
	CreatePairBox( worldId, b2_dynamicBody, 10.5f, 0.0f, &negativeDef );
	b2ShapeDef positiveDef = defC;
	positiveDef.filter.groupIndex = 1;
	CreatePairBox( worldId, b2_dynamicBody, 20.0f, 0.0f, &positiveDef );
	CreatePairBox( worldId, b2_dynamicBody, 20.5f, 0.0f, &positiveDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L976", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L976:value", PairContactCount( worldId ) );

	// Custom filter rejects a tagged shape
	b2ShapeDef customDef = b2DefaultShapeDef();
	customDef.enableCustomFiltering = true;
	CreatePairBox( worldId, b2_dynamicBody, 30.0f, 0.0f, &customDef );
	CreatePairBox( worldId, b2_dynamicBody, 30.5f, 0.0f, &customDef );
	customDef.userData = (void*)1;
	CreatePairBox( worldId, b2_dynamicBody, 40.0f, 0.0f, &customDef );
	customDef.userData = NULL;
	CreatePairBox( worldId, b2_dynamicBody, 40.5f, 0.0f, &customDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 3 );
	ORACLE_BOOL( "L988", PairContactCount( worldId ) == 3 );
	ORACLE_I32( "L988:value", PairContactCount( worldId ) );

	// Sensors never make contacts
	b2ShapeDef sensorDef = b2DefaultShapeDef();
	sensorDef.isSensor = true;
	b2ShapeDef plainDef = b2DefaultShapeDef();
	CreatePairBox( worldId, b2_staticBody, 50.0f, 0.0f, &sensorDef );
	CreatePairBox( worldId, b2_dynamicBody, 50.5f, 0.0f, &plainDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 3 );
	ORACLE_BOOL( "L997", PairContactCount( worldId ) == 3 );
	ORACLE_I32( "L997:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// A body that outran its margin on the step it went to sleep still gets its contact. This is the
// hazard behind section G: the marks live in the tree and do not care which set the body is in.
static int PairSleepEscape( int workerCount, bool sleepBeforePairs )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = b2Vec2_zero;
	worldDef.workerCount = workerCount;
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &shapeDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	bodyDef.position = (b2Pos){ -2.0f, 0.0f };
	bodyDef.linearVelocity = (b2Vec2){ 58.0f, 0.0f };
	b2BodyId moverId = b2CreateBody( worldId, &bodyDef );
	b2Polygon square = b2MakeSquare( 0.5f );
	b2CreatePolygonShape( moverId, &shapeDef, &square );

	// One step carries the mover 0.97, far past its margin and into the fat box of the wall with a
	// few centimeters to spare, but not into the wall itself
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L1026", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L1026:value", PairContactCount( worldId ) );

	if ( sleepBeforePairs )
	{
		b2Body_SetAwake( moverId, false );
		ENSURE( b2Body_IsAwake( moverId ) == false );
		ORACLE_BOOL( "L1031", b2Body_IsAwake( moverId ) == false );
	}

	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1035", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1035:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// Enable recreates the proxies with marks, on either side of a resting pair. The fat boxes
// overlap across the gap while the shapes do not, so nothing pushes and nothing moves.
static int PairEnable( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	PairBox ground = CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &shapeDef );
	PairBox box = CreatePairBox( worldId, b2_dynamicBody, 1.1f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1051", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1051:value", PairContactCount( worldId ) );

	b2Body_Disable( box.bodyId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L1055", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L1055:value", PairContactCount( worldId ) );

	b2Body_Enable( box.bodyId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1059", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1059:value", PairContactCount( worldId ) );

	// The static side, the resting box has no mark of its own
	b2Body_Disable( ground.bodyId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L1064", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L1064:value", PairContactCount( worldId ) );

	b2Body_Enable( ground.bodyId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1068", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1068:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// A type change recreates the proxies with marks and must find the resting neighbors again
static int PairSetType( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &shapeDef );
	PairBox middle = CreatePairBox( worldId, b2_dynamicBody, 1.1f, 0.0f, &shapeDef );
	CreatePairBox( worldId, b2_dynamicBody, 2.2f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L1084", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L1084:value", PairContactCount( worldId ) );

	// Kinematic drops the static pair and keeps the dynamic one
	b2Body_SetType( middle.bodyId, b2_kinematicBody );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1089", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1089:value", PairContactCount( worldId ) );

	b2Body_SetType( middle.bodyId, b2_dynamicBody );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L1093", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L1093:value", PairContactCount( worldId ) );

	// Static drops the static pair too and the dynamic pair becomes static versus dynamic
	b2Body_SetType( middle.bodyId, b2_staticBody );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1098", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1098:value", PairContactCount( worldId ) );

	b2Body_SetType( middle.bodyId, b2_dynamicBody );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L1102", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L1102:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// Bullet is a flag only, the pair must be untouched
static int PairSetBullet( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	CreatePairBox( worldId, b2_staticBody, 0.0f, 0.0f, &shapeDef );
	PairBox box = CreatePairBox( worldId, b2_dynamicBody, 1.1f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1117", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1117:value", PairContactCount( worldId ) );

	b2Body_SetBullet( box.bodyId, true );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1121", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1121:value", PairContactCount( worldId ) );

	b2Body_SetBullet( box.bodyId, false );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1125", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1125:value", PairContactCount( worldId ) );

	b2DestroyWorld( worldId );
	return 0;
}

// A static rebuild must keep the pending marks. The resting box has none of its own, so a lost
// mark means the pair is never found.
static int PairRebuildStatic( int workerCount )
{
	b2WorldId worldId = CreatePairWorld( workerCount );
	b2ShapeDef shapeDef = b2DefaultShapeDef();

	PairBox box = CreatePairBox( worldId, b2_dynamicBody, 0.0f, 0.0f, &shapeDef );
	CreatePairBox( worldId, b2_staticBody, 10.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L1141", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L1141:value", PairContactCount( worldId ) );

	// No marks to keep
	b2World_RebuildStaticTree( worldId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 0 );
	ORACLE_BOOL( "L1146", PairContactCount( worldId ) == 0 );
	ORACLE_I32( "L1146:value", PairContactCount( worldId ) );

	CreatePairBox( worldId, b2_staticBody, 1.1f, 0.0f, &shapeDef );
	b2World_RebuildStaticTree( worldId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1151", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1151:value", PairContactCount( worldId ) );

	// A quiet static shape has no mark to keep
	b2ShapeDef quietDef = b2DefaultShapeDef();
	quietDef.invokeContactCreation = false;
	CreatePairBox( worldId, b2_staticBody, -1.1f, 0.0f, &quietDef );
	b2World_RebuildStaticTree( worldId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1159", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1159:value", PairContactCount( worldId ) );

	// Overlapping the resting box, the contact must touch on the first step after the rebuild
	CreatePairBox( worldId, b2_staticBody, 0.0f, 0.5f, &shapeDef );
	b2World_RebuildStaticTree( worldId );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 2 );
	ORACLE_BOOL( "L1165", PairContactCount( worldId ) == 2 );
	ORACLE_I32( "L1165:value", PairContactCount( worldId ) );

	b2ContactData contactData[4];
	int touchingCount = b2Body_GetContactData( box.bodyId, contactData, 4 );
	ENSURE( touchingCount == 1 );
	ORACLE_BOOL( "L1169", touchingCount == 1 );
	ORACLE_I32( "L1169:value", touchingCount );
	ENSURE( contactData[0].manifold.pointCount > 0 );
	ORACLE_BOOL( "L1170", contactData[0].manifold.pointCount > 0 );
	ORACLE_I32( "L1170:value", contactData[0].manifold.pointCount );

	b2DestroyWorld( worldId );
	return 0;
}

// A static shape created over a sleeping body still gets its contact, parked with the sleeper
// until it wakes. The world counter only sees awake contacts, so this counts on the body.
static int PairSleepingStatic( int workerCount )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = b2Vec2_zero;
	worldDef.workerCount = workerCount;
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	PairBox sleeper = CreatePairBox( worldId, b2_dynamicBody, 0.0f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	b2Body_SetAwake( sleeper.bodyId, false );
	ENSURE( b2Body_IsAwake( sleeper.bodyId ) == false );
	ORACLE_BOOL( "L1189", b2Body_IsAwake( sleeper.bodyId ) == false );

	CreatePairBox( worldId, b2_staticBody, 0.5f, 0.0f, &shapeDef );
	StepPairWorld( worldId );
	ENSURE( b2Body_GetContactCapacity( sleeper.bodyId ) == 1 );
	ORACLE_BOOL( "L1193", b2Body_GetContactCapacity( sleeper.bodyId ) == 1 );
	ORACLE_I32( "L1193:value", b2Body_GetContactCapacity( sleeper.bodyId ) );
	ENSURE( b2Body_IsAwake( sleeper.bodyId ) == false );
	ORACLE_BOOL( "L1194", b2Body_IsAwake( sleeper.bodyId ) == false );

	// The quiet kind never pairs with a sleeper
	b2ShapeDef quietDef = b2DefaultShapeDef();
	quietDef.invokeContactCreation = false;
	CreatePairBox( worldId, b2_staticBody, -0.5f, 0.0f, &quietDef );
	StepPairWorld( worldId );
	ENSURE( b2Body_GetContactCapacity( sleeper.bodyId ) == 1 );
	ORACLE_BOOL( "L1201", b2Body_GetContactCapacity( sleeper.bodyId ) == 1 );
	ORACLE_I32( "L1201:value", b2Body_GetContactCapacity( sleeper.bodyId ) );

	// Waking brings the parked contact into the step and it touches at once
	b2Body_SetAwake( sleeper.bodyId, true );
	StepPairWorld( worldId );
	ENSURE( PairContactCount( worldId ) == 1 );
	ORACLE_BOOL( "L1206", PairContactCount( worldId ) == 1 );
	ORACLE_I32( "L1206:value", PairContactCount( worldId ) );

	b2ContactData contactData[4];
	int touchingCount = b2Body_GetContactData( sleeper.bodyId, contactData, 4 );
	ENSURE( touchingCount == 1 );
	ORACLE_BOOL( "L1210", touchingCount == 1 );
	ORACLE_I32( "L1210:value", touchingCount );
	ENSURE( contactData[0].manifold.pointCount > 0 );
	ORACLE_BOOL( "L1211", contactData[0].manifold.pointCount > 0 );
	ORACLE_I32( "L1211:value", contactData[0].manifold.pointCount );

	b2DestroyWorld( worldId );
	return 0;
}

// The contact sim carries a persistent locator for each body sim instead of a copy written by the
// collide task every step. These scenarios drive every site that changes a body's set or local index
// while its contacts are live, then step under validation so b2ValidateSolverSets and the prepare
// asserts are the real check. Worker count parity on the state hash guards the multithreaded paths.

static void StepLocatorWorld( b2WorldId worldId, int stepCount )
{
	for ( int i = 0; i < stepCount; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_WORLD_STEP( worldId, "L1226" );
	}
}

static bool StepUntilAsleep( b2WorldId worldId, b2BodyId bodyId, int maxSteps )
{
	for ( int i = 0; i < maxSteps; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_WORLD_STEP( worldId, "L1234" );
		if ( b2Body_IsAwake( bodyId ) == false )
		{
			return true;
		}
	}

	return false;
}

static b2WorldId CreateLocatorWorld( int workerCount, bool enableSleep )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = (b2Vec2){ 0.0f, -10.0f };
	worldDef.enableSleep = enableSleep;
	worldDef.workerCount = workerCount;
	return b2CreateWorld( &worldDef );
}

static b2BodyId CreateLocatorGround( b2WorldId worldId, float x, float halfWidth )
{
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.position = (b2Pos){ x, -0.5f };
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( halfWidth, 0.5f );
	b2CreatePolygonShape( bodyId, &shapeDef, &box );
	return bodyId;
}

static b2BodyId CreateLocatorBox( b2WorldId worldId, b2BodyType type, float x, float y )
{
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = type;
	bodyDef.position = (b2Pos){ x, y };
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon square = b2MakeSquare( 0.5f );
	b2CreatePolygonShape( bodyId, &shapeDef, &square );
	return bodyId;
}

// A pile sleeps on static ground and a dropped body wakes it. Covers first touch, the graph add and
// the wake of a whole sleeping set.
static int LocatorSleepWake( int workerCount, uint64_t* hash )
{
	b2WorldId worldId = CreateLocatorWorld( workerCount, true );
	CreateLocatorGround( worldId, 0.0f, 10.0f );

	b2BodyId pile[3];
	for ( int i = 0; i < 3; ++i )
	{
		pile[i] = CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 0.5f + i );
	}

	ENSURE( StepUntilAsleep( worldId, pile[0], 600 ) );
	ORACLE_BOOL( "L1291", true );
	ENSURE( b2Body_IsAwake( pile[2] ) == false );
	ORACLE_BOOL( "L1292", b2Body_IsAwake( pile[2] ) == false );

	b2BodyId dropper = CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 6.0f );
	ENSURE( b2Body_IsAwake( pile[0] ) == false );
	ORACLE_BOOL( "L1295", b2Body_IsAwake( pile[0] ) == false );

	StepLocatorWorld( worldId, 60 );
	ENSURE( b2Body_IsAwake( pile[0] ) );
	ORACLE_BOOL( "L1298", b2Body_IsAwake( pile[0] ) );
	ENSURE( b2Body_IsAwake( dropper ) );
	ORACLE_BOOL( "L1299", b2Body_IsAwake( dropper ) );

	ENSURE( StepUntilAsleep( worldId, dropper, 900 ) );
	ORACLE_BOOL( "L1301", true );
	ENSURE( b2Body_IsAwake( pile[0] ) == false );
	ORACLE_BOOL( "L1302", b2Body_IsAwake( pile[0] ) == false );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorSleepWake:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// Destroying a body in the middle of the awake set swaps the last awake sim into the hole while that
// body's contacts are live.
static int LocatorDestroyAwakeBody( int workerCount, uint64_t* hash )
{
	b2WorldId worldId = CreateLocatorWorld( workerCount, false );
	CreateLocatorGround( worldId, 0.0f, 20.0f );

	b2BodyId stack[4];
	for ( int i = 0; i < 4; ++i )
	{
		stack[i] = CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 0.5f + i );
	}

	// A second stack far away owns the tail of the awake set
	b2BodyId neighbors[3];
	for ( int i = 0; i < 3; ++i )
	{
		neighbors[i] = CreateLocatorBox( worldId, b2_dynamicBody, 10.0f, 0.5f + i );
	}

	StepLocatorWorld( worldId, 120 );
	float restY = (float)b2Body_GetPosition( neighbors[2] ).y;

	b2DestroyBody( stack[0] );
	StepLocatorWorld( worldId, 30 );

	ENSURE_SMALL( (float)b2Body_GetPosition( neighbors[2] ).y - restY, 0.01f );
	ORACLE_F32( "L1335", (float)b2Body_GetPosition( neighbors[2] ).y - restY );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorDestroyAwakeBody:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// Destroying a static body swaps the last static sim into the hole. The moved body still has awake
// contacts, so its side of them has to be re-encoded or the collide task reads past the static set.
static int LocatorDestroyStaticBody( int workerCount, uint64_t* hash )
{
	b2WorldId worldId = CreateLocatorWorld( workerCount, false );

	b2BodyId platforms[3];
	b2BodyId boxes[3];
	for ( int i = 0; i < 3; ++i )
	{
		float x = 10.0f * i;
		platforms[i] = CreateLocatorGround( worldId, x, 2.0f );
		boxes[i] = CreateLocatorBox( worldId, b2_dynamicBody, x, 0.5f );
	}

	StepLocatorWorld( worldId, 60 );
	float restY = (float)b2Body_GetPosition( boxes[2] ).y;

	b2DestroyBody( platforms[0] );
	StepLocatorWorld( worldId, 40 );

	ENSURE_SMALL( (float)b2Body_GetPosition( boxes[2] ).y - restY, 0.01f );
	ORACLE_F32( "L1363", (float)b2Body_GetPosition( boxes[2] ).y - restY );
	ENSURE( (float)b2Body_GetPosition( boxes[0] ).y < -1.0f );
	ORACLE_BOOL( "L1364", (float)b2Body_GetPosition( boxes[0] ).y < -1.0f );
	ORACLE_F32( "L1364:value", (float)b2Body_GetPosition( boxes[0] ).y );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorDestroyStaticBody:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// A kinematic that never sleeps keeps a non touching contact to a body that falls asleep. That
// contact stays in the awake set with its sleeping side on the cold path, then the kinematic closes
// the gap and wakes it.
static int LocatorSleepingNeighbor( int workerCount, uint64_t* hash )
{
	b2WorldId worldId = CreateLocatorWorld( workerCount, true );
	CreateLocatorGround( worldId, 0.0f, 10.0f );

	b2BodyId sleeper = CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 0.5f );

	// The fat boxes overlap across a tenth of a meter, well past the speculative distance
	b2BodyId mover = CreateLocatorBox( worldId, b2_kinematicBody, 1.1f, 0.5f );
	b2Body_EnableSleep( mover, false );

	ENSURE( StepUntilAsleep( worldId, sleeper, 600 ) );
	ORACLE_BOOL( "L1385", true );
	ENSURE( b2Body_IsAwake( mover ) );
	ORACLE_BOOL( "L1386", b2Body_IsAwake( mover ) );

	b2ContactData contactData[4];
	ENSURE( b2Body_GetContactCapacity( mover ) == 1 );
	ORACLE_BOOL( "L1389", b2Body_GetContactCapacity( mover ) == 1 );
	ORACLE_I32( "L1389:value", b2Body_GetContactCapacity( mover ) );
	ENSURE( b2Body_GetContactData( mover, contactData, 4 ) == 0 );
	ORACLE_BOOL( "L1390", b2Body_GetContactData( mover, contactData, 4 ) == 0 );

	StepLocatorWorld( worldId, 30 );
	ENSURE( b2Body_IsAwake( sleeper ) == false );
	ORACLE_BOOL( "L1393", b2Body_IsAwake( sleeper ) == false );

	b2Body_SetLinearVelocity( mover, (b2Vec2){ -2.0f, 0.0f } );
	StepLocatorWorld( worldId, 10 );
	ENSURE( b2Body_IsAwake( sleeper ) );
	ORACLE_BOOL( "L1397", b2Body_IsAwake( sleeper ) );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorSleepingNeighbor:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// Type change, disable and enable all route through the body transfer, which pulls a body sim out of
// the middle of a set while other bodies hold contacts.
static int LocatorTypeChange( int workerCount, uint64_t* hash )
{
	b2WorldId worldId = CreateLocatorWorld( workerCount, false );
	CreateLocatorGround( worldId, 0.0f, 20.0f );

	b2BodyId subject = CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 0.5f );
	CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 1.5f );

	b2BodyId neighbors[3];
	for ( int i = 0; i < 3; ++i )
	{
		neighbors[i] = CreateLocatorBox( worldId, b2_dynamicBody, 10.0f, 0.5f + i );
	}

	StepLocatorWorld( worldId, 120 );
	float restY = (float)b2Body_GetPosition( neighbors[2] ).y;

	b2Body_SetType( subject, b2_staticBody );
	StepLocatorWorld( worldId, 30 );
	ENSURE( b2Body_GetType( subject ) == b2_staticBody );
	ORACLE_BOOL( "L1425", b2Body_GetType( subject ) == b2_staticBody );
	ORACLE_I32( "L1425:value", b2Body_GetType( subject ) );

	b2Body_SetType( subject, b2_dynamicBody );
	StepLocatorWorld( worldId, 30 );
	ENSURE( b2Body_GetType( subject ) == b2_dynamicBody );
	ORACLE_BOOL( "L1429", b2Body_GetType( subject ) == b2_dynamicBody );
	ORACLE_I32( "L1429:value", b2Body_GetType( subject ) );

	b2Body_Disable( subject );
	StepLocatorWorld( worldId, 10 );
	ENSURE( b2Body_IsEnabled( subject ) == false );
	ORACLE_BOOL( "L1433", b2Body_IsEnabled( subject ) == false );

	b2Body_Enable( subject );
	StepLocatorWorld( worldId, 30 );
	ENSURE( b2Body_IsEnabled( subject ) );
	ORACLE_BOOL( "L1437", b2Body_IsEnabled( subject ) );

	ENSURE_SMALL( (float)b2Body_GetPosition( neighbors[2] ).y - restY, 0.01f );
	ORACLE_F32( "L1439", (float)b2Body_GetPosition( neighbors[2] ).y - restY );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorTypeChange:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// A joint between two sleeping bodies merges their sets instead of waking them, while an awake
// kinematic holds a non touching contact to one of them across the merge.
static int LocatorSetMerge( int workerCount, uint64_t* hash )
{
	b2WorldId worldId = CreateLocatorWorld( workerCount, true );
	CreateLocatorGround( worldId, 0.0f, 20.0f );

	b2BodyId left = CreateLocatorBox( worldId, b2_dynamicBody, -5.0f, 0.5f );
	b2BodyId right = CreateLocatorBox( worldId, b2_dynamicBody, 5.0f, 0.5f );

	b2BodyId watcher = CreateLocatorBox( worldId, b2_kinematicBody, -3.9f, 0.5f );
	b2Body_EnableSleep( watcher, false );

	ENSURE( StepUntilAsleep( worldId, left, 600 ) );
	ORACLE_BOOL( "L1459", true );
	ENSURE( b2Body_IsAwake( right ) == false );
	ORACLE_BOOL( "L1460", b2Body_IsAwake( right ) == false );
	ENSURE( b2Body_IsAwake( watcher ) );
	ORACLE_BOOL( "L1461", b2Body_IsAwake( watcher ) );
	ENSURE( b2Body_GetContactCapacity( watcher ) == 1 );
	ORACLE_BOOL( "L1462", b2Body_GetContactCapacity( watcher ) == 1 );
	ORACLE_I32( "L1462:value", b2Body_GetContactCapacity( watcher ) );

	b2DistanceJointDef jointDef = b2DefaultDistanceJointDef();
	jointDef.base.bodyIdA = left;
	jointDef.base.bodyIdB = right;
	jointDef.length = 10.0f;
	b2JointId jointId = b2CreateDistanceJoint( worldId, &jointDef );

	ENSURE( b2Body_IsAwake( left ) == false );
	ORACLE_BOOL( "L1470", b2Body_IsAwake( left ) == false );
	ENSURE( b2Body_IsAwake( right ) == false );
	ORACLE_BOOL( "L1471", b2Body_IsAwake( right ) == false );

	StepLocatorWorld( worldId, 30 );
	ENSURE( b2Body_IsAwake( left ) == false );
	ORACLE_BOOL( "L1474", b2Body_IsAwake( left ) == false );

	b2Body_SetAwake( left, true );
	StepLocatorWorld( worldId, 10 );
	ENSURE( b2Body_IsAwake( right ) );
	ORACLE_BOOL( "L1478", b2Body_IsAwake( right ) );

	b2DestroyJoint( jointId );
	StepLocatorWorld( worldId, 10 );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorSetMerge:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// A body moving far enough in one step carries the fast flag, which the collide task now reads off
// the sim. A public setter mid flight makes sure a flag sync does not drop it.
static int LocatorFastBody( int workerCount, uint64_t* hash )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = b2Vec2_zero;
	worldDef.workerCount = workerCount;
	b2WorldId worldId = b2CreateWorld( &worldDef );

	CreateLocatorBox( worldId, b2_staticBody, 10.0f, 0.0f );
	b2BodyId bullet = CreateLocatorBox( worldId, b2_dynamicBody, 0.0f, 0.0f );
	b2Body_SetLinearVelocity( bullet, (b2Vec2){ 40.0f, 0.0f } );

	// A step carries it two thirds of a meter, well past half its minimum extent
	StepLocatorWorld( worldId, 4 );
	ENSURE( (float)b2Body_GetPosition( bullet ).x > 2.0f );
	ORACLE_BOOL( "L1503", (float)b2Body_GetPosition( bullet ).x > 2.0f );
	ORACLE_F32( "L1503:value", (float)b2Body_GetPosition( bullet ).x );

	b2Body_EnableSleep( bullet, false );
	StepLocatorWorld( worldId, 4 );

	b2Body_EnableSleep( bullet, true );
	StepLocatorWorld( worldId, 30 );

	// It stopped at the wall instead of tunneling through
	ENSURE( (float)b2Body_GetPosition( bullet ).x < 9.1f );
	ORACLE_BOOL( "L1512", (float)b2Body_GetPosition( bullet ).x < 9.1f );
	ORACLE_F32( "L1512:value", (float)b2Body_GetPosition( bullet ).x );
	ENSURE( (float)b2Body_GetPosition( bullet ).x > 8.5f );
	ORACLE_BOOL( "L1513", (float)b2Body_GetPosition( bullet ).x > 8.5f );
	ORACLE_F32( "L1513:value", (float)b2Body_GetPosition( bullet ).x );

	*hash = b2World_GetStateHash( worldId );
	ORACLE_U64( "LocatorFastBody:hash", *hash );
	b2DestroyWorld( worldId );
	return 0;
}

// A row of boxes at rest in zero gravity is unaffected by sleeping, so the run that sleeps and wakes
// on impact has to land where the run that never sleeps lands. Positions rather than the state hash
// because the hash folds in set and local index, which sleeping is entitled to change.
static int LocatorSleepParity( int workerCount )
{
	float finalX[2][3];

	for ( int pass = 0; pass < 2; ++pass )
	{
		b2WorldDef worldDef = b2DefaultWorldDef();
		worldDef.gravity = b2Vec2_zero;
		worldDef.enableSleep = pass == 0;
		worldDef.workerCount = workerCount;
		b2WorldId worldId = b2CreateWorld( &worldDef );

		b2BodyId boxes[3];
		for ( int i = 0; i < 3; ++i )
		{
			boxes[i] = CreateLocatorBox( worldId, b2_dynamicBody, (float)i, 0.0f );
		}

		StepLocatorWorld( worldId, 60 );
		ENSURE( b2Body_IsAwake( boxes[0] ) == ( pass == 1 ) );
		ORACLE_BOOL( "L1542", b2Body_IsAwake( boxes[0] ) == ( pass == 1 ) );

		b2BodyId impactor = CreateLocatorBox( worldId, b2_dynamicBody, -5.0f, 0.0f );
		b2Body_SetLinearVelocity( impactor, (b2Vec2){ 5.0f, 0.0f } );

		StepLocatorWorld( worldId, 120 );
		ENSURE( b2Body_IsAwake( boxes[2] ) );
		ORACLE_BOOL( "L1548", b2Body_IsAwake( boxes[2] ) );

		for ( int i = 0; i < 3; ++i )
		{
			finalX[pass][i] = (float)b2Body_GetPosition( boxes[i] ).x;
		}

		b2DestroyWorld( worldId );
	}

	for ( int i = 0; i < 3; ++i )
	{
		ENSURE_SMALL( finalX[0][i] - finalX[1][i], 0.01f );
		ORACLE_F32( "L1560", finalX[0][i] - finalX[1][i] );
	}

	return 0;
}

static int BodySimLocatorTest( void )
{
	int workerCounts[2] = { 1, 4 };
	uint64_t hashes[2];

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorSleepWake( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1573", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1575", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorDestroyAwakeBody( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1579", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1581", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorDestroyStaticBody( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1585", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1587", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorSleepingNeighbor( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1591", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1593", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorTypeChange( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1597", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1599", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorSetMerge( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1603", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1605", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorFastBody( workerCounts[i], hashes + i ) == 0 );
		ORACLE_BOOL( "L1609", true );
	}
	ENSURE( hashes[0] == hashes[1] );
	ORACLE_BOOL( "L1611", hashes[0] == hashes[1] );

	for ( int i = 0; i < 2; ++i )
	{
		ENSURE( LocatorSleepParity( workerCounts[i] ) == 0 );
		ORACLE_BOOL( "L1615", true );
	}

	return 0;
}

static int BroadPhasePairsTest( void )
{
	int workerCounts[2] = { 1, 4 };
	for ( int i = 0; i < 2; ++i )
	{
		int workerCount = workerCounts[i];
		ENSURE( PairCreation( workerCount ) == 0 );
		ORACLE_BOOL( "L1627", true );
		ENSURE( PairTeleport( workerCount ) == 0 );
		ORACLE_BOOL( "L1628", true );
		ENSURE( PairRefilter( workerCount ) == 0 );
		ORACLE_BOOL( "L1629", true );
		ENSURE( PairCollideConnected( workerCount ) == 0 );
		ORACLE_BOOL( "L1630", true );
		ENSURE( PairFilters( workerCount ) == 0 );
		ORACLE_BOOL( "L1631", true );
		ENSURE( PairEnable( workerCount ) == 0 );
		ORACLE_BOOL( "L1632", true );
		ENSURE( PairSetType( workerCount ) == 0 );
		ORACLE_BOOL( "L1633", true );
		ENSURE( PairSetBullet( workerCount ) == 0 );
		ORACLE_BOOL( "L1634", true );
		ENSURE( PairRebuildStatic( workerCount ) == 0 );
		ORACLE_BOOL( "L1635", true );
		ENSURE( PairSleepEscape( workerCount, false ) == 0 );
		ORACLE_BOOL( "L1636", true );
		ENSURE( PairSleepEscape( workerCount, true ) == 0 );
		ORACLE_BOOL( "L1637", true );
		ENSURE( PairSleepingStatic( workerCount ) == 0 );
		ORACLE_BOOL( "L1638", true );
	}

	return 0;
}

// A joint with no dynamic body lives in the static set even when one body is an awake kinematic.
// Disabling and enabling that body has to respect this, whether the joint was created that way or
// got there through body type changes. Box3D issue #164.
static int NonDynamicJointDisableTest( void )
{
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = b2Vec2_zero;
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	b2BodyId bodyIdA = b2CreateBody( worldId, &bodyDef );
	bodyDef.position = (b2Pos){ 2.0f, 0.0f };
	b2BodyId bodyIdB = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( bodyIdA, &shapeDef, &box );
	b2CreatePolygonShape( bodyIdB, &shapeDef, &box );

	b2WeldJointDef jointDef = b2DefaultWeldJointDef();
	jointDef.base.bodyIdA = bodyIdA;
	jointDef.base.bodyIdB = bodyIdB;
	b2JointId weldId = b2CreateWeldJoint( worldId, &jointDef );

	b2Body_SetType( bodyIdA, b2_staticBody );
	b2Body_SetType( bodyIdB, b2_kinematicBody );
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L1671" );

	b2Body_Disable( bodyIdB );
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L1674" );
	b2Body_Enable( bodyIdB );
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L1676" );
	ENSURE( b2Joint_IsValid( weldId ) );
	ORACLE_BOOL( "L1677", b2Joint_IsValid( weldId ) );

	b2World* world = b2GetWorldFromId( worldId );
	ENSURE( b2GetJointFullId( world, weldId )->setIndex == b2_staticSet );
	ORACLE_BOOL( "L1680", b2GetJointFullId( world, weldId )->setIndex == b2_staticSet );
	ORACLE_I32( "L1680:value", b2GetJointFullId( world, weldId )->setIndex );

	b2Body_SetType( bodyIdB, b2_dynamicBody );
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L1683" );
	ENSURE( b2GetJointFullId( world, weldId )->setIndex == b2_awakeSet );
	ORACLE_BOOL( "L1684", b2GetJointFullId( world, weldId )->setIndex == b2_awakeSet );
	ORACLE_I32( "L1684:value", b2GetJointFullId( world, weldId )->setIndex );

	// Created directly between a static and a kinematic body
	bodyDef.type = b2_kinematicBody;
	bodyDef.position = (b2Pos){ 4.0f, 0.0f };
	b2BodyId bodyIdC = b2CreateBody( worldId, &bodyDef );
	b2CreatePolygonShape( bodyIdC, &shapeDef, &box );

	b2RevoluteJointDef revoluteDef = b2DefaultRevoluteJointDef();
	revoluteDef.base.bodyIdA = bodyIdA;
	revoluteDef.base.bodyIdB = bodyIdC;
	b2JointId revoluteId = b2CreateRevoluteJoint( worldId, &revoluteDef );

	b2Body_Disable( bodyIdC );
	b2Body_Enable( bodyIdC );
	b2World_Step( worldId, 1.0f / 60.0f, 4 );
	ORACLE_WORLD_STEP( worldId, "L1699" );
	ENSURE( b2GetJointFullId( world, revoluteId )->setIndex == b2_staticSet );
	ORACLE_BOOL( "L1700", b2GetJointFullId( world, revoluteId )->setIndex == b2_staticSet );
	ORACLE_I32( "L1700:value", b2GetJointFullId( world, revoluteId )->setIndex );

	b2DestroyWorld( worldId );
	return 0;
}

int WorldTest( void )
{
	RUN_SUBTEST( HelloWorld );
	RUN_SUBTEST( EmptyWorld );
	RUN_SUBTEST( DestroyAllBodiesWorld );
	RUN_SUBTEST( TestIsValid );
	RUN_SUBTEST( TestWorldRecycle );
	RUN_SUBTEST( TestWorldCoverage );
	RUN_SUBTEST( TestSensor );
	RUN_SUBTEST( TestSetWorkerCount );
	RUN_SUBTEST( ChainSegmentShapeTest );
	RUN_SUBTEST( SetBulletDriftTest );
	RUN_SUBTEST( DeferredMassFlagSyncTest );
	RUN_SUBTEST( EnableSleepFlagSyncTest );
	RUN_SUBTEST( EnableContactRecyclingTest );
	RUN_SUBTEST( EnlargedProxyDestroyedTest );
	RUN_SUBTEST( BroadPhasePairsTest );
	RUN_SUBTEST( BodySimLocatorTest );
	RUN_SUBTEST( NonDynamicJointDisableTest );

	return 0;
}
