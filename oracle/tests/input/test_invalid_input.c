// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"
#include "test_macros.h"

#include "box2d/box2d.h"
#include "box2d/math_functions.h"

#include <math.h>
#include <stdint.h>

// The guard logs once per rejected call in every build config, while the assert only fires in
// debug. So the log count is the portable signal that bad input was turned away.
static int logCount;

static void CountingLogFcn( const char* message )
{
	MAYBE_UNUSED( message );
	logCount += 1;
}

static int SilentAssertFcn( const char* condition, const char* fileName, int lineNumber )
{
	MAYBE_UNUSED( condition );
	MAYBE_UNUSED( fileName );
	MAYBE_UNUSED( lineNumber );

	// Keep going so the subtest can report rather than break to the debugger
	return 0;
}

static void BeginRejectCount( void )
{
	logCount = 0;
	b2SetLogFcn( CountingLogFcn );
	b2SetAssertFcn( SilentAssertFcn );
}

static void EndRejectCount( void )
{
	b2SetAssertFcn( TestAssertFcn );
	b2SetLogFcn( TestLogFcn );
}

#define REJECTS( CALL )                                                                                                          \
	do                                                                                                                           \
	{                                                                                                                            \
		int countBefore = logCount; \
		uint64_t stateBefore = b2World_GetStateHash( worldId ); \
		ORACLE_U64( "reject.before", stateBefore ); \
		CALL;                                                                                                                    \
		ORACLE_I32( "reject", logCount - countBefore ); \
		uint64_t stateAfter = b2World_GetStateHash( worldId ); \
		ORACLE_U64( "reject.after", stateAfter ); \
		ENSURE( stateAfter == stateBefore ); \
		ObserveRejectedSettings( worldId ); \
		if ( logCount != countBefore + 1 )                                                                                       \
		{                                                                                                                        \
			printf( "not rejected: " #CALL "\n" );                                                                               \
			return 1;                                                                                                            \
		}                                                                                                                        \
	}                                                                                                                            \
	while ( false )

#define ACCEPTS( CALL )                                                                                                          \
	do                                                                                                                           \
	{                                                                                                                            \
		int countBefore = logCount;                                                                                              \
		CALL;                                                                                                                    \
		ORACLE_I32( "accept", logCount - countBefore ); \
		if ( logCount != countBefore )                                                                                           \
		{                                                                                                                        \
			printf( "wrongly rejected: " #CALL "\n" );                                                                           \
			return 1;                                                                                                            \
		}                                                                                                                        \
	}                                                                                                                            \
	while ( false )

static const float badValues[] = { NAN, INFINITY, -INFINITY };

static bool IsBodyFinite( b2BodyId bodyId )
{
	return b2IsValidPosition( b2Body_GetPosition( bodyId ) ) && b2IsValidRotation( b2Body_GetRotation( bodyId ) ) &&
		   b2IsValidVec2( b2Body_GetLinearVelocity( bodyId ) ) && b2IsValidFloat( b2Body_GetAngularVelocity( bodyId ) );
}

static void ObserveRejectedSettings( b2WorldId worldId )
{
	b2World* world = b2GetWorldFromId( worldId );
	ORACLE_F32( "settings.contact.hertz", world->contactHertz );
	ORACLE_F32( "settings.contact.damping", world->contactDampingRatio );
	ORACLE_F32( "settings.contact.push", world->contactSpeed );
	for ( int i = 0; i < world->bodies.count; ++i )
	{
		b2Body* body = b2Array_Get( world->bodies, i );
		if ( body->id == B2_NULL_INDEX ) { continue; }
		b2BodyId id = b2MakeBodyId( world, i );
		b2MassData mass = b2Body_GetMassData( id );
		ORACLE_F32( "settings.mass", mass.mass );
		ORACLE_VEC2( "settings.center", mass.center );
		ORACLE_F32( "settings.inertia", mass.rotationalInertia );
		ORACLE_F32( "settings.linearDamping", b2Body_GetLinearDamping( id ) );
		ORACLE_F32( "settings.angularDamping", b2Body_GetAngularDamping( id ) );
		ORACLE_F32( "settings.gravityScale", b2Body_GetGravityScale( id ) );
		ORACLE_F32( "settings.sleepThreshold", b2Body_GetSleepThreshold( id ) );
		ORACLE_F32( "settings.safetyFactor", b2Body_GetSafetyFactor( id ) );
		int count = b2Body_GetShapeCount( id );
		b2ShapeId* shapes = calloc( count > 0 ? count : 1, sizeof( b2ShapeId ) );
		if ( shapes == NULL ) { abort(); }
		int returned = b2Body_GetShapes( id, shapes, count );
		for ( int s = 0; s < returned; ++s )
		{
			b2ShapeId shape = shapes[s];
			ORACLE_BOOL( "settings.shape.valid", b2Shape_IsValid( shape ) );
			ORACLE_BOOL( "settings.shape.owner", B2_ID_EQUALS( b2Shape_GetBody( shape ), id ) );
			ORACLE_F32( "settings.shape.density", b2Shape_GetDensity( shape ) );
			b2SurfaceMaterial material = b2Shape_GetSurfaceMaterial( shape );
			ORACLE_F32( "settings.material.friction", material.friction );
			ORACLE_F32( "settings.material.restitution", material.restitution );
			ORACLE_F32( "settings.material.rolling", material.rollingResistance );
			ORACLE_F32( "settings.material.tangent", material.tangentSpeed );
			ORACLE_U64( "settings.material.tag", material.userMaterialId );
			ORACLE_U32( "settings.material.color", material.customColor );
			b2AABB bounds = b2Shape_GetAABB( shape );
			ORACLE_VEC2( "settings.aabb.lower", bounds.lowerBound );
			ORACLE_VEC2( "settings.aabb.upper", bounds.upperBound );
			b2ShapeType type = b2Shape_GetType( shape );
			ORACLE_I32( "settings.shape.type", type );
			if ( type == b2_polygonShape )
			{
				b2Polygon polygon = b2Shape_GetPolygon( shape );
				ORACLE_I32( "settings.polygon.count", polygon.count );
				ORACLE_F32( "settings.polygon.radius", polygon.radius );
				ORACLE_VEC2( "settings.polygon.centroid", polygon.centroid );
				for ( int p = 0; p < polygon.count; ++p )
				{
					ORACLE_VEC2( "settings.polygon.vertex", polygon.vertices[p] );
					ORACLE_VEC2( "settings.polygon.normal", polygon.normals[p] );
				}
			}
		}
		ORACLE_I32( "settings.body.shapes", returned );
		free( shapes );
	}
}

static void ObserveRejectedJoint( b2JointId joint )
{
	b2Transform a = b2Joint_GetLocalFrameA( joint ), b = b2Joint_GetLocalFrameB( joint );
	ORACLE_VEC2( "joint.frameA.p", a.p );
	ORACLE_ROT( "joint.frameA.q", a.q );
	ORACLE_VEC2( "joint.frameB.p", b.p );
	ORACLE_ROT( "joint.frameB.q", b.q );
	float hertz, damping;
	b2Joint_GetConstraintTuning( joint, &hertz, &damping );
	ORACLE_F32( "joint.hertz", hertz );
	ORACLE_F32( "joint.damping", damping );
	ORACLE_F32( "joint.forceThreshold", b2Joint_GetForceThreshold( joint ) );
	ORACLE_F32( "joint.torqueThreshold", b2Joint_GetTorqueThreshold( joint ) );
	ORACLE_F32( "joint.springHertz", b2RevoluteJoint_GetSpringHertz( joint ) );
	ORACLE_F32( "joint.springDamping", b2RevoluteJoint_GetSpringDampingRatio( joint ) );
	ORACLE_F32( "joint.target", b2RevoluteJoint_GetTargetAngle( joint ) );
	ORACLE_F32( "joint.lower", b2RevoluteJoint_GetLowerLimit( joint ) );
	ORACLE_F32( "joint.upper", b2RevoluteJoint_GetUpperLimit( joint ) );
	ORACLE_F32( "joint.motorSpeed", b2RevoluteJoint_GetMotorSpeed( joint ) );
	ORACLE_F32( "joint.maxTorque", b2RevoluteJoint_GetMaxMotorTorque( joint ) );
}

static bool ObserveConstructionCounts( b2WorldId worldId, int bodyCount )
{
	b2World* world = b2GetWorldFromId( worldId );
	b2Counters counts = b2World_GetCounters( worldId );
	int chains = b2GetIdCount( &world->chainIdPool );
	ORACLE_I32( "construction.bodies", counts.bodyCount );
	ORACLE_I32( "construction.shapes", counts.shapeCount );
	ORACLE_I32( "construction.joints", counts.jointCount );
	ORACLE_I32( "construction.contacts", counts.contactCount );
	ORACLE_I32( "construction.chains", chains );
	int shapes = 0, joints = 0;
	for ( int i = 0; i < world->bodies.count; ++i )
	{
		b2Body* body = b2Array_Get( world->bodies, i );
		if ( body->id == B2_NULL_INDEX ) { continue; }
		b2BodyId id = b2MakeBodyId( world, i );
		shapes += b2Body_GetShapeCount( id );
		joints += b2Body_GetJointCount( id );
	}
	ORACLE_I32( "construction.enumerated.shapes", shapes );
	ORACLE_I32( "construction.enumerated.joints", joints );
	return counts.bodyCount == bodyCount && counts.shapeCount == 1 && counts.jointCount == 0 &&
		counts.contactCount == 0 && chains == 0 && shapes == 1 && joints == 0;
}

static void IndependentWorld( void )
{
	b2WorldDef def = b2DefaultWorldDef();
	def.workerCount = 4;
	def.gravity = (b2Vec2){ 3.0f, -7.0f };
	b2WorldId world = b2CreateWorld( &def );
	ORACLE_BOOL( "independent.world.valid", b2World_IsValid( world ) );
	if ( !b2World_IsValid( world ) ) { abort(); }
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	bodyDef.linearVelocity = (b2Vec2){ 2.0f, 1.0f };
	b2BodyId body = b2CreateBody( world, &bodyDef );
	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2ShapeId shape = b2CreatePolygonShape( body, &shapeDef, &box );
	ORACLE_BOOL( "independent.body.valid", b2Body_IsValid( body ) );
	ORACLE_BOOL( "independent.shape.valid", b2Shape_IsValid( shape ) );
	if ( !b2Body_IsValid( body ) || !b2Shape_IsValid( shape ) ) { abort(); }
	for ( int i = 0; i < 2; ++i )
	{
		b2World_Step( world, 1.0f / 60.0f, 4 );
		ORACLE_INPUT_TRACE( world, ORACLE_KEY( "independent.%d", i ), false );
	}
	b2DestroyWorld( world );
	ORACLE_BOOL( "independent.world.destroyed", !b2World_IsValid( world ) );
	if ( b2World_IsValid( world ) ) { abort(); }
}

static int WorldInputTest( void )
{
	ORACLE_OPEN( "input/invalid_input/WorldInputTest" );
	ORACLE_INPUT_RESET();
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );
	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	bodyDef.linearVelocity = (b2Vec2){ 2.0f, 3.0f };
	b2BodyId body = b2CreateBody( worldId, &bodyDef );
	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( body, &shapeDef, &box );

	b2Vec2 gravity = b2World_GetGravity( worldId );
	float restitutionThreshold = b2World_GetRestitutionThreshold( worldId );
	float hitEventThreshold = b2World_GetHitEventThreshold( worldId );
	float recycleDistance = b2World_GetContactRecycleDistance( worldId );
	float maximumLinearSpeed = b2World_GetMaximumLinearSpeed( worldId );
	int restitutionIterations = b2World_GetRestitutionIterations( worldId );

	BeginRejectCount();

	for ( int i = 0; i < ARRAY_COUNT( badValues ); ++i )
	{
		float bad = badValues[i];

		REJECTS( b2World_SetGravity( worldId, (b2Vec2){ bad, 0.0f } ) );
		REJECTS( b2World_SetGravity( worldId, (b2Vec2){ 0.0f, bad } ) );
		REJECTS( b2World_SetRestitutionThreshold( worldId, bad ) );
		REJECTS( b2World_SetHitEventThreshold( worldId, bad ) );
		REJECTS( b2World_SetContactRecycleDistance( worldId, bad ) );
		REJECTS( b2World_SetMaximumLinearSpeed( worldId, bad ) );
		REJECTS( b2World_SetContactTuning( worldId, bad, 10.0f, 3.0f ) );
		REJECTS( b2World_SetContactTuning( worldId, 30.0f, bad, 3.0f ) );
		REJECTS( b2World_SetContactTuning( worldId, 30.0f, 10.0f, bad ) );
		REJECTS( b2World_Step( worldId, bad, 4 ) );
		ORACLE_INPUT_TRACE( worldId, ORACLE_KEY( "bad.%d.hash", i ), true );
	}

	REJECTS( b2World_SetRestitutionIterations( worldId, -1 ) );

	ACCEPTS( b2World_Step( worldId, 1.0f / 60.0f, 4 ) );
	ORACLE_INPUT_TRACE( worldId, "step.hash", true );
	ORACLE_VEC2( "step.gravity", b2World_GetGravity( worldId ) );

	EndRejectCount();

	ORACLE_BOOL( "check.115", b2World_GetGravity( worldId ).x == gravity.x && b2World_GetGravity( worldId ).y == gravity.y );
	ENSURE( b2World_GetGravity( worldId ).x == gravity.x && b2World_GetGravity( worldId ).y == gravity.y );
	ORACLE_BOOL( "check.116", b2World_GetRestitutionThreshold( worldId ) == restitutionThreshold );
	ENSURE( b2World_GetRestitutionThreshold( worldId ) == restitutionThreshold );
	ORACLE_BOOL( "check.117", b2World_GetHitEventThreshold( worldId ) == hitEventThreshold );
	ENSURE( b2World_GetHitEventThreshold( worldId ) == hitEventThreshold );
	ORACLE_BOOL( "check.118", b2World_GetContactRecycleDistance( worldId ) == recycleDistance );
	ENSURE( b2World_GetContactRecycleDistance( worldId ) == recycleDistance );
	ORACLE_BOOL( "check.119", b2World_GetMaximumLinearSpeed( worldId ) == maximumLinearSpeed );
	ENSURE( b2World_GetMaximumLinearSpeed( worldId ) == maximumLinearSpeed );
	ORACLE_BOOL( "check.120", b2World_GetRestitutionIterations( worldId ) == restitutionIterations );
	ENSURE( b2World_GetRestitutionIterations( worldId ) == restitutionIterations );

	b2DestroyWorld( worldId );
	ORACLE_INPUT_RESET();
	ORACLE_CLOSE();
	return 0;
}

static int BodyInputTest( void )
{
	ORACLE_OPEN( "input/invalid_input/BodyInputTest" );
	ORACLE_INPUT_RESET();
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( bodyId, &shapeDef, &box );

	BeginRejectCount();

	for ( int i = 0; i < ARRAY_COUNT( badValues ); ++i )
	{
		float bad = badValues[i];
		b2Vec2 badVec = { bad, 0.0f };
		b2Pos badPos = { bad, 0.0f };
		b2Rot badRot = { bad, 0.0f };
		b2WorldTransform badTarget = { badPos, b2Rot_identity };

		REJECTS( b2Body_SetLinearVelocity( bodyId, badVec ) );
		REJECTS( b2Body_SetAngularVelocity( bodyId, bad ) );
		REJECTS( b2Body_SetTransform( bodyId, badPos, b2Rot_identity ) );
		REJECTS( b2Body_SetTransform( bodyId, b2Pos_zero, badRot ) );
		REJECTS( b2Body_SetTargetTransform( bodyId, badTarget, 1.0f / 60.0f, true ) );
		REJECTS( b2Body_SetTargetTransform( bodyId, b2WorldTransform_identity, bad, true ) );
		REJECTS( b2Body_ApplyForce( bodyId, badVec, b2Pos_zero, true ) );
		REJECTS( b2Body_ApplyForce( bodyId, b2Vec2_zero, badPos, true ) );
		REJECTS( b2Body_ApplyForceToCenter( bodyId, badVec, true ) );
		REJECTS( b2Body_ApplyTorque( bodyId, bad, true ) );
		REJECTS( b2Body_ApplyLinearImpulse( bodyId, badVec, b2Pos_zero, true ) );
		REJECTS( b2Body_ApplyLinearImpulse( bodyId, b2Vec2_zero, badPos, true ) );
		REJECTS( b2Body_ApplyLinearImpulseToCenter( bodyId, badVec, true ) );
		REJECTS( b2Body_ApplyAngularImpulse( bodyId, bad, true ) );
		REJECTS( b2Body_SetMassData( bodyId, (b2MassData){ bad, b2Vec2_zero, 1.0f } ) );
		REJECTS( b2Body_SetMassData( bodyId, (b2MassData){ 1.0f, badVec, 1.0f } ) );
		REJECTS( b2Body_SetMassData( bodyId, (b2MassData){ 1.0f, b2Vec2_zero, bad } ) );
		REJECTS( b2Body_SetLinearDamping( bodyId, bad ) );
		REJECTS( b2Body_SetAngularDamping( bodyId, bad ) );
		REJECTS( b2Body_SetGravityScale( bodyId, bad ) );
		REJECTS( b2Body_SetSleepThreshold( bodyId, bad ) );
		REJECTS( b2Body_SetSafetyFactor( bodyId, bad ) );
	}

	// A negative value on a quantity the API documents as non-negative is turned away too,
	// since negative damping grows velocity without bound.
	REJECTS( b2Body_SetLinearDamping( bodyId, -1.0f ) );
	REJECTS( b2Body_SetAngularDamping( bodyId, -1.0f ) );

	ACCEPTS( b2Body_SetLinearVelocity( bodyId, (b2Vec2){ 1.0f, 2.0f } ) );
	ACCEPTS( b2Body_ApplyForceToCenter( bodyId, (b2Vec2){ 0.0f, 10.0f }, true ) );

	for ( int i = 0; i < 30; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_INPUT_TRACE( worldId, ORACLE_KEY( "step.%d.hash", i ), true );
		ORACLE_POS( ORACLE_KEY( "step.%d.position", i ), b2Body_GetPosition( bodyId ) );
		ORACLE_ROT( ORACLE_KEY( "step.%d.rotation", i ), b2Body_GetRotation( bodyId ) );
		ORACLE_VEC2( ORACLE_KEY( "step.%d.velocity", i ), b2Body_GetLinearVelocity( bodyId ) );
		ORACLE_F32( ORACLE_KEY( "step.%d.angular", i ), b2Body_GetAngularVelocity( bodyId ) );
	}

	EndRejectCount();

	ORACLE_BOOL( "check.188", IsBodyFinite( bodyId ) );
	ENSURE( IsBodyFinite( bodyId ) );

	b2DestroyWorld( worldId );
	ORACLE_INPUT_RESET();
	ORACLE_CLOSE();
	return 0;
}

static int ShapeInputTest( void )
{
	ORACLE_OPEN( "input/invalid_input/ShapeInputTest" );
	ORACLE_INPUT_RESET();
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2ShapeId shapeId = b2CreatePolygonShape( bodyId, &shapeDef, &box );

	float density = b2Shape_GetDensity( shapeId );
	float friction = b2Shape_GetFriction( shapeId );
	float restitution = b2Shape_GetRestitution( shapeId );

	BeginRejectCount();

	for ( int i = 0; i < ARRAY_COUNT( badValues ); ++i )
	{
		float bad = badValues[i];
		b2Vec2 badVec = { bad, 0.0f };

		b2Circle badCircle = { badVec, 1.0f };
		b2Capsule badCapsule = { { -0.5f, 0.0f }, { 0.5f, 0.0f }, bad };
		b2Segment badSegment = { badVec, { 1.0f, 0.0f } };
		b2ChainSegment badChainSegment = { badVec, { { 0.0f, 0.0f }, { 1.0f, 0.0f } }, { 2.0f, 0.0f }, -1 };

		b2Polygon badPolygon = box;
		badPolygon.radius = bad;

		b2Polygon badVertexPolygon = box;
		badVertexPolygon.vertices[0] = badVec;

		b2SurfaceMaterial badMaterial = b2DefaultSurfaceMaterial();
		badMaterial.friction = bad;

		REJECTS( b2Shape_SetDensity( shapeId, bad, true ) );
		REJECTS( b2Shape_SetFriction( shapeId, bad ) );
		REJECTS( b2Shape_SetRestitution( shapeId, bad ) );
		REJECTS( b2Shape_SetSurfaceMaterial( shapeId, &badMaterial ) );
		REJECTS( b2Shape_ApplyWind( shapeId, badVec, 1.0f, 1.0f, true ) );
		REJECTS( b2Shape_ApplyWind( shapeId, b2Vec2_zero, bad, 1.0f, true ) );
		REJECTS( b2Shape_ApplyWind( shapeId, b2Vec2_zero, 1.0f, bad, true ) );
		REJECTS( b2Shape_SetCircle( shapeId, &badCircle ) );
		REJECTS( b2Shape_SetCapsule( shapeId, &badCapsule ) );
		REJECTS( b2Shape_SetSegment( shapeId, &badSegment ) );
		REJECTS( b2Shape_SetChainSegment( shapeId, &badChainSegment ) );
		REJECTS( b2Shape_SetPolygon( shapeId, &badPolygon ) );
		REJECTS( b2Shape_SetPolygon( shapeId, &badVertexPolygon ) );
	}

	ACCEPTS( b2Shape_SetFriction( shapeId, 0.4f ) );

	for ( int i = 0; i < 10; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_INPUT_TRACE( worldId, ORACLE_KEY( "step.%d.hash", i ), true );
		ORACLE_POS( ORACLE_KEY( "step.%d.position", i ), b2Body_GetPosition( bodyId ) );
		ORACLE_ROT( ORACLE_KEY( "step.%d.rotation", i ), b2Body_GetRotation( bodyId ) );
		ORACLE_VEC2( ORACLE_KEY( "step.%d.velocity", i ), b2Body_GetLinearVelocity( bodyId ) );
		ORACLE_F32( ORACLE_KEY( "step.%d.angular", i ), b2Body_GetAngularVelocity( bodyId ) );
	}

	EndRejectCount();

	// The rejected geometry never reached the shape, so the proxy is still the original box
	ORACLE_BOOL( "check.257", b2Shape_GetType( shapeId ) == b2_polygonShape );
	ENSURE( b2Shape_GetType( shapeId ) == b2_polygonShape );
	ORACLE_BOOL( "check.258", b2IsValidAABB( b2Shape_GetAABB( shapeId ) ) );
	ENSURE( b2IsValidAABB( b2Shape_GetAABB( shapeId ) ) );
	ORACLE_BOOL( "check.259", b2IsValidAABB( b2World_GetBounds( worldId ) ) );
	ENSURE( b2IsValidAABB( b2World_GetBounds( worldId ) ) );
	ORACLE_BOOL( "check.260", b2Shape_GetDensity( shapeId ) == density );
	ENSURE( b2Shape_GetDensity( shapeId ) == density );
	ORACLE_BOOL( "check.261", b2Shape_GetFriction( shapeId ) == 0.4f && friction == friction );
	ENSURE( b2Shape_GetFriction( shapeId ) == 0.4f && friction == friction );
	ORACLE_BOOL( "check.262", b2Shape_GetRestitution( shapeId ) == restitution );
	ENSURE( b2Shape_GetRestitution( shapeId ) == restitution );
	ORACLE_BOOL( "check.263", IsBodyFinite( bodyId ) );
	ENSURE( IsBodyFinite( bodyId ) );

	b2DestroyWorld( worldId );
	ORACLE_INPUT_RESET();
	ORACLE_CLOSE();
	return 0;
}

static int JointInputTest( void )
{
	ORACLE_OPEN( "input/invalid_input/JointInputTest" );
	ORACLE_INPUT_RESET();
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	b2BodyId groundId = b2CreateBody( worldId, &bodyDef );

	bodyDef.type = b2_dynamicBody;
	bodyDef.position = (b2Pos){ 0.0f, -1.0f };
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( bodyId, &shapeDef, &box );

	b2RevoluteJointDef jointDef = b2DefaultRevoluteJointDef();
	jointDef.base.bodyIdA = groundId;
	jointDef.base.bodyIdB = bodyId;
	jointDef.base.localFrameB.p = (b2Vec2){ 0.0f, 1.0f };
	b2JointId jointId = b2CreateRevoluteJoint( worldId, &jointDef );

	BeginRejectCount();

	for ( int i = 0; i < ARRAY_COUNT( badValues ); ++i )
	{
		float bad = badValues[i];
		b2Transform badFrame = { { bad, 0.0f }, b2Rot_identity };

		REJECTS( b2Joint_SetLocalFrameA( jointId, badFrame ) );
		REJECTS( b2Joint_SetLocalFrameB( jointId, badFrame ) );
		REJECTS( b2Joint_SetConstraintTuning( jointId, bad, 1.0f ) );
		REJECTS( b2Joint_SetConstraintTuning( jointId, 60.0f, bad ) );
		REJECTS( b2Joint_SetForceThreshold( jointId, bad ) );
		REJECTS( b2Joint_SetTorqueThreshold( jointId, bad ) );
		REJECTS( b2RevoluteJoint_SetSpringHertz( jointId, bad ) );
		REJECTS( b2RevoluteJoint_SetSpringDampingRatio( jointId, bad ) );
		REJECTS( b2RevoluteJoint_SetTargetAngle( jointId, bad ) );
		REJECTS( b2RevoluteJoint_SetLimits( jointId, bad, 1.0f ) );
		REJECTS( b2RevoluteJoint_SetLimits( jointId, -1.0f, bad ) );
		REJECTS( b2RevoluteJoint_SetMotorSpeed( jointId, bad ) );
		REJECTS( b2RevoluteJoint_SetMaxMotorTorque( jointId, bad ) );
		ObserveRejectedJoint( jointId );
	}

	ACCEPTS( b2RevoluteJoint_SetMotorSpeed( jointId, 2.0f ) );
	ObserveRejectedJoint( jointId );

	for ( int i = 0; i < 30; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_INPUT_TRACE( worldId, ORACLE_KEY( "step.%d.hash", i ), true );
		ORACLE_POS( ORACLE_KEY( "step.%d.position", i ), b2Body_GetPosition( bodyId ) );
		ORACLE_ROT( ORACLE_KEY( "step.%d.rotation", i ), b2Body_GetRotation( bodyId ) );
		ORACLE_VEC2( ORACLE_KEY( "step.%d.velocity", i ), b2Body_GetLinearVelocity( bodyId ) );
		ORACLE_F32( ORACLE_KEY( "step.%d.angular", i ), b2Body_GetAngularVelocity( bodyId ) );
	}

	EndRejectCount();

	ORACLE_BOOL( "check.322", IsBodyFinite( bodyId ) );
	ENSURE( IsBodyFinite( bodyId ) );
	ORACLE_BOOL( "check.323", b2IsValidVec2( b2Joint_GetConstraintForce( jointId ) ) );
	ENSURE( b2IsValidVec2( b2Joint_GetConstraintForce( jointId ) ) );
	ORACLE_BOOL( "check.324", b2IsValidFloat( b2Joint_GetConstraintTorque( jointId ) ) );
	ENSURE( b2IsValidFloat( b2Joint_GetConstraintTorque( jointId ) ) );

	b2DestroyWorld( worldId );
	ORACLE_INPUT_RESET();
	ORACLE_CLOSE();
	return 0;
}

static int CreateInputTest( void )
{
	ORACLE_OPEN( "input/invalid_input/CreateInputTest" );
	ORACLE_INPUT_RESET();
	b2WorldDef worldDef = b2DefaultWorldDef();
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	bodyDef.type = b2_dynamicBody;
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( bodyId, &shapeDef, &box );

	BeginRejectCount();

	for ( int i = 0; i < ARRAY_COUNT( badValues ); ++i )
	{
		float bad = badValues[i];
		b2Vec2 badVec = { bad, 0.0f };
		uint64_t constructionHash = b2World_GetStateHash( worldId );

		b2WorldDef badWorldDef = b2DefaultWorldDef();
		badWorldDef.gravity = badVec;
		b2WorldId badWorldId = b2CreateWorld( &badWorldDef );
		ORACLE_BOOL( "check.353", b2World_IsValid( badWorldId ) == false );
		ENSURE( b2World_IsValid( badWorldId ) == false );

		b2BodyDef badBodyDef = b2DefaultBodyDef();
		badBodyDef.position = (b2Pos){ bad, 0.0f };
		bool bodyRejected = b2Body_IsValid( b2CreateBody( worldId, &badBodyDef ) ) == false;
		ENSURE( bodyRejected );
		ORACLE_I32( "check.357", logCount );
		ORACLE_BOOL( "check.357.result", bodyRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2ShapeDef badShapeDef = b2DefaultShapeDef();
		badShapeDef.density = bad;
		bool shapeRejected = b2Shape_IsValid( b2CreatePolygonShape( bodyId, &badShapeDef, &box ) ) == false;
		ENSURE( shapeRejected );
		ORACLE_I32( "check.361", logCount );
		ORACLE_BOOL( "check.361.result", shapeRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2Circle badCircle = { badVec, 1.0f };
		bool circleRejected = b2Shape_IsValid( b2CreateCircleShape( bodyId, &shapeDef, &badCircle ) ) == false;
		ENSURE( circleRejected );
		ORACLE_I32( "check.364", logCount );
		ORACLE_BOOL( "check.364.result", circleRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2Capsule badCapsule = { { -0.5f, 0.0f }, { 0.5f, 0.0f }, bad };
		bool capsuleRejected = b2Shape_IsValid( b2CreateCapsuleShape( bodyId, &shapeDef, &badCapsule ) ) == false;
		ENSURE( capsuleRejected );
		ORACLE_I32( "check.367", logCount );
		ORACLE_BOOL( "check.367.result", capsuleRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2Segment badSegment = { badVec, { 1.0f, 0.0f } };
		bool segmentRejected = b2Shape_IsValid( b2CreateSegmentShape( bodyId, &shapeDef, &badSegment ) ) == false;
		ENSURE( segmentRejected );
		ORACLE_I32( "check.370", logCount );
		ORACLE_BOOL( "check.370.result", segmentRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2Polygon badPolygon = box;
		badPolygon.radius = bad;
		bool polygonRejected = b2Shape_IsValid( b2CreatePolygonShape( bodyId, &shapeDef, &badPolygon ) ) == false;
		ENSURE( polygonRejected );
		ORACLE_I32( "check.374", logCount );
		ORACLE_BOOL( "check.374.result", polygonRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2Vec2 chainPoints[4] = { { -4.0f, 0.0f }, { 0.0f, 0.0f }, { 4.0f, 0.0f }, { 8.0f, 0.0f } };
		chainPoints[2] = badVec;
		b2SurfaceMaterial material = b2DefaultSurfaceMaterial();
		b2ChainDef badChainDef = b2DefaultChainDef();
		badChainDef.points = chainPoints;
		badChainDef.pointCount = 4;
		badChainDef.materials = &material;
		badChainDef.materialCount = 1;
		badChainDef.isLoop = true;
		bool chainRejected = b2Chain_IsValid( b2CreateChain( bodyId, &badChainDef ) ) == false;
		ENSURE( chainRejected );
		ORACLE_I32( "check.385", logCount );
		ORACLE_BOOL( "check.385.result", chainRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2RevoluteJointDef badJointDef = b2DefaultRevoluteJointDef();
		badJointDef.base.bodyIdA = bodyId;
		badJointDef.base.bodyIdB = b2CreateBody( worldId, &bodyDef );
		badJointDef.hertz = bad;
		constructionHash = b2World_GetStateHash( worldId );
		bool jointRejected = b2Joint_IsValid( b2CreateRevoluteJoint( worldId, &badJointDef ) ) == false;
		ENSURE( jointRejected );
		ORACLE_I32( "check.391", logCount );
		ORACLE_BOOL( "check.391.result", jointRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2ExplosionDef badExplosionDef = b2DefaultExplosionDef();
		badExplosionDef.radius = bad;
		REJECTS( b2World_Explode( worldId, &badExplosionDef ) );
		ENSURE( ObserveConstructionCounts( worldId, 2 + i ) );
		IndependentWorld();
	}

	for ( int i = 0; i < 10; ++i )
	{
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_INPUT_TRACE( worldId, ORACLE_KEY( "step.%d.hash", i ), true );
		ORACLE_POS( ORACLE_KEY( "step.%d.position", i ), b2Body_GetPosition( bodyId ) );
		ORACLE_ROT( ORACLE_KEY( "step.%d.rotation", i ), b2Body_GetRotation( bodyId ) );
		ORACLE_VEC2( ORACLE_KEY( "step.%d.velocity", i ), b2Body_GetLinearVelocity( bodyId ) );
		ORACLE_F32( ORACLE_KEY( "step.%d.angular", i ), b2Body_GetAngularVelocity( bodyId ) );
	}

	EndRejectCount();

	ORACLE_BOOL( "check.405", IsBodyFinite( bodyId ) );
	ENSURE( IsBodyFinite( bodyId ) );
	ORACLE_BOOL( "check.406", b2IsValidAABB( b2World_GetBounds( worldId ) ) );
	ENSURE( b2IsValidAABB( b2World_GetBounds( worldId ) ) );

	b2DestroyWorld( worldId );
	ORACLE_INPUT_RESET();
	ORACLE_CLOSE();
	return 0;
}

static int RecordReplayTest( void )
{
	ORACLE_OPEN( "input/invalid_input/RecordReplayTest" );
	ORACLE_INPUT_RESET();
	b2WorldDef worldDef = b2DefaultWorldDef();
	worldDef.gravity = (b2Vec2){ 0.0f, -10.0f };
	worldDef.workerCount = 1;
	b2WorldId worldId = b2CreateWorld( &worldDef );

	b2Recording* recording = b2CreateRecording( 0 );
	b2World_StartRecording( worldId, recording );

	b2BodyDef bodyDef = b2DefaultBodyDef();
	b2BodyId groundId = b2CreateBody( worldId, &bodyDef );

	b2ShapeDef shapeDef = b2DefaultShapeDef();
	b2Segment ground = { { -20.0f, 0.0f }, { 20.0f, 0.0f } };
	b2CreateSegmentShape( groundId, &shapeDef, &ground );

	bodyDef.type = b2_dynamicBody;
	bodyDef.position = (b2Pos){ 0.0f, 4.0f };
	b2BodyId bodyId = b2CreateBody( worldId, &bodyDef );

	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( bodyId, &shapeDef, &box );

	BeginRejectCount();

	// Interleave turned away calls with real ones so the op stream carries both. A mutator is
	// rejected before it is recorded, while a create is recorded with the null id it returned.
	for ( int i = 0; i < 30; ++i )
	{
		float bad = badValues[i % ARRAY_COUNT( badValues )];

		REJECTS( b2Body_ApplyForceToCenter( bodyId, (b2Vec2){ bad, 0.0f }, true ) );
		REJECTS( b2Body_SetAngularVelocity( bodyId, bad ) );

		b2ShapeDef badShapeDef = b2DefaultShapeDef();
		badShapeDef.density = bad;
		uint64_t constructionHash = b2World_GetStateHash( worldId );
		bool shapeRejected = b2Shape_IsValid( b2CreatePolygonShape( bodyId, &badShapeDef, &box ) ) == false;
		ENSURE( shapeRejected );
		ORACLE_I32( "check.449", logCount );
		ORACLE_BOOL( "check.449.result", shapeRejected );
		ORACLE_U64( "construction.state", b2World_GetStateHash( worldId ) );
		ENSURE( b2World_GetStateHash( worldId ) == constructionHash );

		b2Body_ApplyForceToCenter( bodyId, (b2Vec2){ 0.0f, 1.0f }, true );
		b2World_Step( worldId, 1.0f / 60.0f, 4 );
		ORACLE_INPUT_TRACE( worldId, ORACLE_KEY( "step.%d.hash", i ), true );
		ORACLE_POS( ORACLE_KEY( "step.%d.position", i ), b2Body_GetPosition( bodyId ) );
		ORACLE_ROT( ORACLE_KEY( "step.%d.rotation", i ), b2Body_GetRotation( bodyId ) );
		ORACLE_VEC2( ORACLE_KEY( "step.%d.velocity", i ), b2Body_GetLinearVelocity( bodyId ) );
		ORACLE_F32( ORACLE_KEY( "step.%d.angular", i ), b2Body_GetAngularVelocity( bodyId ) );
	}

	ORACLE_BOOL( "check.455", IsBodyFinite( bodyId ) );
	ENSURE( IsBodyFinite( bodyId ) );

	b2World_StopRecording( worldId );
	b2DestroyWorld( worldId );

	const uint8_t* data = b2Recording_GetData( recording );
	int size = b2Recording_GetSize( recording );
	ORACLE_BOOL( "check.462", size > 0 );
	ENSURE( size > 0 );
	bool replay0 = b2ValidateReplay( data, size, 0 );
	ENSURE( replay0 );
	ORACLE_I32( "check.463", logCount );
	ORACLE_BOOL( "check.463.result", replay0 );
	bool replay4 = b2ValidateReplay( data, size, 4 );
	ENSURE( replay4 );
	ORACLE_I32( "check.464", logCount );
	ORACLE_BOOL( "check.464.result", replay4 );

	EndRejectCount();

	b2DestroyRecording( recording );
	ORACLE_INPUT_RESET();
	ORACLE_CLOSE();
	return 0;
}

int InvalidInputTest( void )
{
	RUN_SUBTEST( WorldInputTest );
	RUN_SUBTEST( BodyInputTest );
	RUN_SUBTEST( ShapeInputTest );
	RUN_SUBTEST( JointInputTest );
	RUN_SUBTEST( CreateInputTest );
	RUN_SUBTEST( RecordReplayTest );

	return 0;
}
