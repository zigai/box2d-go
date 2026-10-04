// SPDX-License-Identifier: MIT
#include "oracle.h"

#include "box2d/box2d.h"

#include <assert.h>
#include <stdio.h>
static void collision_policy( b2JointType family );

// The pinned C dispatch rejects mover/pogo torque and separation getters.
// Compare callback-visible assertions as well as the handler-controlled return.
static int getter_assertions;
static int count_getter_assertion( const char* condition, const char* file, int line )
{
	(void)condition;
	(void)file;
	(void)line;
	++getter_assertions;
	return 0;
}
static int fatal_assertion( const char* condition, const char* file, int line )
{
	fprintf( stderr, "BOX2D ASSERTION: %s, %s:%d\n", condition, file, line );
	return 1;
}
static void scalar_getter( const char* phase, const char* field, b2JointId joint, float ( *get )( b2JointId ) )
{
	getter_assertions = 0;
	b2SetAssertFcn( count_getter_assertion );
	float value = get( joint );
	b2SetAssertFcn( fatal_assertion );
	ORACLE_I32( ORACLE_KEY( "%s/%sAssertions", phase, field ), getter_assertions );
	ORACLE_F32( ORACLE_KEY( "%s/%s", phase, field ), value );
}

// Direct API observations, not a transform digest: every field is labelled by action/step.
static void body_state( const char* phase, b2BodyId b )
{
	b2WorldTransform x = b2Body_GetTransform( b );
	ORACLE_POS( ORACLE_KEY( "%s/position", phase ), x.p );
	ORACLE_ROT( ORACLE_KEY( "%s/rotation", phase ), x.q );
	ORACLE_VEC2( ORACLE_KEY( "%s/velocity", phase ), b2Body_GetLinearVelocity( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/angularVelocity", phase ), b2Body_GetAngularVelocity( b ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/awake", phase ), b2Body_IsAwake( b ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/enabled", phase ), b2Body_IsEnabled( b ) );
	ORACLE_I32( ORACLE_KEY( "%s/type", phase ), b2Body_GetType( b ) );
}
static void observe( const char* phase, b2WorldId w, b2JointId j, b2BodyId a, b2BodyId b )
{
	body_state( ORACLE_KEY( "%s/bodyA", phase ), a );
	body_state( ORACLE_KEY( "%s/bodyB", phase ), b );
	ORACLE_I32( ORACLE_KEY( "%s/jointCount", phase ), b2World_GetCounters( w ).jointCount );
	ORACLE_I32( ORACLE_KEY( "%s/contactCount", phase ), b2World_GetCounters( w ).contactCount );
	ORACLE_I32( ORACLE_KEY( "%s/type", phase ), b2Joint_GetType( j ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/bodyA", phase ), B2_ID_EQUALS( b2Joint_GetBodyA( j ), a ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/bodyB", phase ), B2_ID_EQUALS( b2Joint_GetBodyB( j ), b ) );
	b2WorldId owner = b2Joint_GetWorld( j );
	ORACLE_BOOL( ORACLE_KEY( "%s/world", phase ), owner.index1 == w.index1 && owner.generation == w.generation );
	b2Transform fa = b2Joint_GetLocalFrameA( j ), fb = b2Joint_GetLocalFrameB( j );
	ORACLE_VEC2( ORACLE_KEY( "%s/frameA/position", phase ), fa.p );
	ORACLE_ROT( ORACLE_KEY( "%s/frameA/rotation", phase ), fa.q );
	ORACLE_VEC2( ORACLE_KEY( "%s/frameB/position", phase ), fb.p );
	ORACLE_ROT( ORACLE_KEY( "%s/frameB/rotation", phase ), fb.q );
	ORACLE_BOOL( ORACLE_KEY( "%s/collideConnected", phase ), b2Joint_GetCollideConnected( j ) );
	ORACLE_VEC2( ORACLE_KEY( "%s/constraintForce", phase ), b2Joint_GetConstraintForce( j ) );
	scalar_getter( phase, "constraintTorque", j, b2Joint_GetConstraintTorque );
	scalar_getter( phase, "linearSeparation", j, b2Joint_GetLinearSeparation );
	scalar_getter( phase, "angularSeparation", j, b2Joint_GetAngularSeparation );
	float hertz, damping;
	b2Joint_GetConstraintTuning( j, &hertz, &damping );
	ORACLE_F32( ORACLE_KEY( "%s/constraintHertz", phase ), hertz );
	ORACLE_F32( ORACLE_KEY( "%s/constraintDamping", phase ), damping );
	ORACLE_F32( ORACLE_KEY( "%s/forceThreshold", phase ), b2Joint_GetForceThreshold( j ) );
	ORACLE_F32( ORACLE_KEY( "%s/torqueThreshold", phase ), b2Joint_GetTorqueThreshold( j ) );
	switch ( b2Joint_GetType( j ) )
	{
		case b2_distanceJoint:
			ORACLE_F32( ORACLE_KEY( "%s/Length", phase ), b2DistanceJoint_GetLength( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsSpringEnabled", phase ), b2DistanceJoint_IsSpringEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringHertz", phase ), b2DistanceJoint_GetSpringHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringDampingRatio", phase ), b2DistanceJoint_GetSpringDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringForce", phase ), b2DistanceJoint_GetSpringForce( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsLimitEnabled", phase ), b2DistanceJoint_IsLimitEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MinLength", phase ), b2DistanceJoint_GetMinLength( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxLength", phase ), b2DistanceJoint_GetMaxLength( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/CurrentLength", phase ), b2DistanceJoint_GetCurrentLength( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsMotorEnabled", phase ), b2DistanceJoint_IsMotorEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorSpeed", phase ), b2DistanceJoint_GetMotorSpeed( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxMotorForce", phase ), b2DistanceJoint_GetMaxMotorForce( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorForce", phase ), b2DistanceJoint_GetMotorForce( j ) );
			{
				float lo, hi;
				b2DistanceJoint_GetSpringForceRange( j, &lo, &hi );
				ORACLE_F32( ORACLE_KEY( "%s/LowerSpringForce", phase ), lo );
				ORACLE_F32( ORACLE_KEY( "%s/UpperSpringForce", phase ), hi );
			}
			break;
		case b2_filterJoint:
			break;
		case b2_motorJoint:
			ORACLE_VEC2( ORACLE_KEY( "%s/LinearVelocity", phase ), b2MotorJoint_GetLinearVelocity( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/AngularVelocity", phase ), b2MotorJoint_GetAngularVelocity( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxVelocityForce", phase ), b2MotorJoint_GetMaxVelocityForce( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxVelocityTorque", phase ), b2MotorJoint_GetMaxVelocityTorque( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/LinearHertz", phase ), b2MotorJoint_GetLinearHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/LinearDampingRatio", phase ), b2MotorJoint_GetLinearDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxSpringForce", phase ), b2MotorJoint_GetMaxSpringForce( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/AngularHertz", phase ), b2MotorJoint_GetAngularHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/AngularDampingRatio", phase ), b2MotorJoint_GetAngularDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxSpringTorque", phase ), b2MotorJoint_GetMaxSpringTorque( j ) );
			break;
		case b2_moverJoint:
			ORACLE_VEC2( ORACLE_KEY( "%s/LinearVelocity", phase ), b2MoverJoint_GetLinearVelocity( j ) );
			ORACLE_VEC2( ORACLE_KEY( "%s/MaxVelocityForce", phase ), b2MoverJoint_GetMaxVelocityForce( j ) );
			break;
		case b2_pogoJoint:
			ORACLE_F32( ORACLE_KEY( "%s/RestLength", phase ), b2PogoJoint_GetRestLength( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringHertz", phase ), b2PogoJoint_GetSpringHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringDampingRatio", phase ), b2PogoJoint_GetSpringDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/Length", phase ), b2PogoJoint_GetLength( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/Impulse", phase ), b2PogoJoint_GetImpulse( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/Velocity", phase ), b2PogoJoint_GetVelocity( j ) );
			break;
		case b2_prismaticJoint:
			ORACLE_BOOL( ORACLE_KEY( "%s/IsSpringEnabled", phase ), b2PrismaticJoint_IsSpringEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringHertz", phase ), b2PrismaticJoint_GetSpringHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringDampingRatio", phase ), b2PrismaticJoint_GetSpringDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/TargetTranslation", phase ), b2PrismaticJoint_GetTargetTranslation( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsLimitEnabled", phase ), b2PrismaticJoint_IsLimitEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/LowerLimit", phase ), b2PrismaticJoint_GetLowerLimit( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/UpperLimit", phase ), b2PrismaticJoint_GetUpperLimit( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsMotorEnabled", phase ), b2PrismaticJoint_IsMotorEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorSpeed", phase ), b2PrismaticJoint_GetMotorSpeed( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxMotorForce", phase ), b2PrismaticJoint_GetMaxMotorForce( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorForce", phase ), b2PrismaticJoint_GetMotorForce( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/Translation", phase ), b2PrismaticJoint_GetTranslation( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/Speed", phase ), b2PrismaticJoint_GetSpeed( j ) );
			break;
		case b2_revoluteJoint:
			ORACLE_BOOL( ORACLE_KEY( "%s/IsSpringEnabled", phase ), b2RevoluteJoint_IsSpringEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringHertz", phase ), b2RevoluteJoint_GetSpringHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringDampingRatio", phase ), b2RevoluteJoint_GetSpringDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/TargetAngle", phase ), b2RevoluteJoint_GetTargetAngle( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/Angle", phase ), b2RevoluteJoint_GetAngle( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsLimitEnabled", phase ), b2RevoluteJoint_IsLimitEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/LowerLimit", phase ), b2RevoluteJoint_GetLowerLimit( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/UpperLimit", phase ), b2RevoluteJoint_GetUpperLimit( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsMotorEnabled", phase ), b2RevoluteJoint_IsMotorEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorSpeed", phase ), b2RevoluteJoint_GetMotorSpeed( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxMotorTorque", phase ), b2RevoluteJoint_GetMaxMotorTorque( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorTorque", phase ), b2RevoluteJoint_GetMotorTorque( j ) );
			break;
		case b2_weldJoint:
			ORACLE_F32( ORACLE_KEY( "%s/LinearHertz", phase ), b2WeldJoint_GetLinearHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/LinearDampingRatio", phase ), b2WeldJoint_GetLinearDampingRatio( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/AngularHertz", phase ), b2WeldJoint_GetAngularHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/AngularDampingRatio", phase ), b2WeldJoint_GetAngularDampingRatio( j ) );
			break;
		case b2_wheelJoint:
			ORACLE_BOOL( ORACLE_KEY( "%s/IsSpringEnabled", phase ), b2WheelJoint_IsSpringEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringHertz", phase ), b2WheelJoint_GetSpringHertz( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/SpringDampingRatio", phase ), b2WheelJoint_GetSpringDampingRatio( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsLimitEnabled", phase ), b2WheelJoint_IsLimitEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/LowerLimit", phase ), b2WheelJoint_GetLowerLimit( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/UpperLimit", phase ), b2WheelJoint_GetUpperLimit( j ) );
			ORACLE_BOOL( ORACLE_KEY( "%s/IsMotorEnabled", phase ), b2WheelJoint_IsMotorEnabled( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorSpeed", phase ), b2WheelJoint_GetMotorSpeed( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MaxMotorTorque", phase ), b2WheelJoint_GetMaxMotorTorque( j ) );
			ORACLE_F32( ORACLE_KEY( "%s/MotorTorque", phase ), b2WheelJoint_GetMotorTorque( j ) );
			break;
	}
}
static b2JointId create_joint( b2JointType family, b2WorldId w, b2BodyId a, b2BodyId b, b2Transform fa, b2Transform fb,
							   int partition, bool loaded )
{
	switch ( family )
	{
		case b2_distanceJoint:
		{
			b2DistanceJointDef d = b2DefaultDistanceJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			d.length = 1.0f;
			return b2CreateDistanceJoint( w, &d );
		}
		case b2_filterJoint:
		{
			b2FilterJointDef d = b2DefaultFilterJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreateFilterJoint( w, &d );
		}
		case b2_motorJoint:
		{
			b2MotorJointDef d = b2DefaultMotorJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreateMotorJoint( w, &d );
		}
		case b2_moverJoint:
		{
			b2MoverJointDef d = b2DefaultMoverJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreateMoverJoint( w, &d );
		}
		case b2_pogoJoint:
		{
			b2PogoJointDef d = b2DefaultPogoJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			d.normal = (b2Vec2){ 0.6f, 0.8f };
			d.restLength = 1.0f;
			d.hertz = 3.0f;
			d.dampingRatio = 0.2f;
			d.maxTensionForce = partition == 0 ? 0.0f : 2.0f;
			d.maxCompressionForce = partition == 2 ? 100.0f : 4.0f;
			d.impulse = loaded ? 0.2f : 0.0f;
			d.velocity = loaded ? -0.5f : 0.0f;
			return b2CreatePogoJoint( w, &d );
		}
		case b2_prismaticJoint:
		{
			b2PrismaticJointDef d = b2DefaultPrismaticJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreatePrismaticJoint( w, &d );
		}
		case b2_revoluteJoint:
		{
			b2RevoluteJointDef d = b2DefaultRevoluteJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreateRevoluteJoint( w, &d );
		}
		case b2_weldJoint:
		{
			b2WeldJointDef d = b2DefaultWeldJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreateWeldJoint( w, &d );
		}
		case b2_wheelJoint:
		{
			b2WheelJointDef d = b2DefaultWheelJointDef();
			d.base.bodyIdA = a;
			d.base.bodyIdB = b;
			d.base.localFrameA = fa;
			d.base.localFrameB = fb;
			return b2CreateWheelJoint( w, &d );
		}
	}
	assert( false );
	return b2_nullJointId;
}
static void mutate( b2JointId j, int stage )
{
	float cap = stage == 5 ? 0.0f : ( stage == 6 ? 100.0f : 3.0f );
	switch ( b2Joint_GetType( j ) )
	{
		case b2_distanceJoint:
			b2DistanceJoint_EnableSpring( j, stage != 0 );
			b2DistanceJoint_EnableLimit( j, stage >= 2 );
			b2DistanceJoint_EnableMotor( j, stage >= 3 );
			b2DistanceJoint_SetSpringHertz( j, stage == 1 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2DistanceJoint_SetSpringDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2DistanceJoint_SetMotorSpeed( j, stage == 4 ? -4.0f : 4.0f );
			b2DistanceJoint_SetMaxMotorForce( j, cap );
			b2DistanceJoint_SetLength( j, stage == 7 ? 0.7f : 1.3f );
			b2DistanceJoint_SetLengthRange( j, 0.7f, 1.3f );
			b2DistanceJoint_SetSpringForceRange( j, -cap, cap );
			break;
		case b2_filterJoint:
			break;
		case b2_motorJoint:
			b2MotorJoint_SetLinearHertz( j, stage == 0 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2MotorJoint_SetLinearDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2MotorJoint_SetAngularHertz( j, stage == 0 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2MotorJoint_SetAngularDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2MotorJoint_SetLinearVelocity( j, (b2Vec2){ stage == 4 ? -4.0f : 4.0f, -2.0f } );
			b2MotorJoint_SetMaxVelocityForce( j, cap );
			b2MotorJoint_SetMaxVelocityTorque( j, cap );
			b2MotorJoint_SetMaxSpringForce( j, cap );
			b2MotorJoint_SetMaxSpringTorque( j, cap );
			b2MotorJoint_SetAngularVelocity( j, stage == 4 ? -3.0f : 3.0f );
			break;
		case b2_moverJoint:
			b2MoverJoint_SetLinearVelocity( j, (b2Vec2){ stage == 4 ? -4.0f : 4.0f, -2.0f } );
			b2MoverJoint_SetMaxVelocityForce( j, (b2Vec2){ cap, 0.5f * cap } );
			break;
		case b2_pogoJoint:
			b2PogoJoint_SetRestLength( j, stage == 7 ? 0.3f : 1.3f );
			b2PogoJoint_SetSpringHertz( j, stage == 0 ? 0.0f : 5.0f );
			b2PogoJoint_SetSpringDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			break;
		case b2_prismaticJoint:
			b2PrismaticJoint_EnableSpring( j, stage != 0 );
			b2PrismaticJoint_EnableLimit( j, stage >= 2 );
			b2PrismaticJoint_EnableMotor( j, stage >= 3 );
			b2PrismaticJoint_SetSpringHertz( j, stage == 1 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2PrismaticJoint_SetSpringDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2PrismaticJoint_SetMotorSpeed( j, stage == 4 ? -4.0f : 4.0f );
			b2PrismaticJoint_SetMaxMotorForce( j, cap );
			b2PrismaticJoint_SetLimits( j, -0.5f, 0.5f );
			b2PrismaticJoint_SetTargetTranslation( j, stage == 7 ? -0.5f : 0.5f );
			break;
		case b2_revoluteJoint:
			b2RevoluteJoint_EnableSpring( j, stage != 0 );
			b2RevoluteJoint_EnableLimit( j, stage >= 2 );
			b2RevoluteJoint_EnableMotor( j, stage >= 3 );
			b2RevoluteJoint_SetSpringHertz( j, stage == 1 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2RevoluteJoint_SetSpringDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2RevoluteJoint_SetMotorSpeed( j, stage == 4 ? -4.0f : 4.0f );
			b2RevoluteJoint_SetMaxMotorTorque( j, cap );
			b2RevoluteJoint_SetLimits( j, -0.5f, 0.5f );
			b2RevoluteJoint_SetTargetAngle( j, stage == 7 ? -0.5f : 0.5f );
			break;
		case b2_weldJoint:
			b2WeldJoint_SetLinearHertz( j, stage == 0 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2WeldJoint_SetLinearDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2WeldJoint_SetAngularHertz( j, stage == 0 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2WeldJoint_SetAngularDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			break;
		case b2_wheelJoint:
			b2WheelJoint_EnableSpring( j, stage != 0 );
			b2WheelJoint_EnableLimit( j, stage >= 2 );
			b2WheelJoint_EnableMotor( j, stage >= 3 );
			b2WheelJoint_SetSpringHertz( j, stage == 1 ? 0.0f : ( stage == 7 ? 9.0f : 3.0f ) );
			b2WheelJoint_SetSpringDampingRatio( j, stage == 7 ? 1.0f : 0.2f );
			b2WheelJoint_SetMotorSpeed( j, stage == 4 ? -4.0f : 4.0f );
			b2WheelJoint_SetMaxMotorTorque( j, cap );
			b2WheelJoint_SetLimits( j, -0.5f, 0.5f );
			break;
	}
}

static void run_steps( const char* action, int count, bool loaded, b2WorldId w, b2JointId j, b2BodyId a, b2BodyId b )
{
	char label[128];
	for ( int i = 1; i <= count; ++i )
	{
		if ( loaded )
		{
			b2Body_ApplyForceToCenter( b, (b2Vec2){ 6.0f, -9.0f }, true );
			b2Body_ApplyTorque( b, 2.5f, true );
		}
		b2World_Step( w, 1.0f / 60.0f, 4 );
		if ( i == 1 || i == 4 || i == 12 || i == 60 )
		{
			snprintf( label, sizeof( label ), "%s/step%d", action, i );
			observe( label, w, j, a, b );
		}
	}
}
static void boundary( b2JointId j, b2BodyId a, b2BodyId b, float value )
{
	// Binary offset anchors and a quarter turn make exact translation boundaries
	// representable, while retaining an off-origin, rotated mechanism.
	b2Transform fa = { { 0.25f, -0.5f }, b2Rot_identity }, fb = { { -0.25f, 0.5f }, b2Rot_identity };
	b2Joint_SetLocalFrameA( j, fa );
	b2Joint_SetLocalFrameB( j, fb );
	b2Body_SetTransform( a, (b2Pos){ 7.0f, -3.0f }, (b2Rot){ 0.0f, 1.0f } );
	b2Rot axis = (b2Rot){ 0.0f, 1.0f };
	b2Rot q = axis;
	if ( b2Joint_GetType( j ) == b2_revoluteJoint )
		q = b2MulRot( b2MulRot( axis, b2MakeRot( value ) ), (b2Rot){ fb.q.c, -fb.q.s } );
	b2Vec2 delta = b2RotateVector( axis, (b2Vec2){ value, 0.0f } );
	if ( b2Joint_GetType( j ) == b2_revoluteJoint )
		delta = b2Vec2_zero;
	b2Vec2 offset = b2Sub( delta, b2RotateVector( q, fb.p ) );
	b2Body_SetTransform( b, b2OffsetPos( b2Body_GetWorldPoint( a, fa.p ), offset ), q );
	b2Body_SetLinearVelocity( b, b2Vec2_zero );
	b2Body_SetAngularVelocity( b, 0.0f );
	switch ( b2Joint_GetType( j ) )
	{
		case b2_distanceJoint:
			b2DistanceJoint_SetLengthRange( j, 0.5f, 1.5f );
			break;
		case b2_prismaticJoint:
			b2PrismaticJoint_SetLimits( j, -0.5f, 0.5f );
			break;
		case b2_wheelJoint:
			b2WheelJoint_SetLimits( j, -0.5f, 0.5f );
			break;
		case b2_revoluteJoint:
		{
			// An angle is reconstructed through atan2. Use its defined getter as the
			// input bound to test equality, rather than pretending trigonometry is exact.
			float angle = b2RevoluteJoint_GetAngle( j );
			b2RevoluteJoint_SetLimits( j, value == -0.5f ? angle : -0.5f, value == 0.5f ? angle : 0.5f );
			break;
		}
		default:
			break;
	}
}
static void scenario( b2JointType family, int partner, bool loaded, int partition )
{
	char name[128], action[128];
	snprintf( name, sizeof( name ), "joints/%d/partner%d/loaded%d/partition%d", family, partner, loaded, partition );
	ORACLE_OPEN( name );
	b2WorldDef wd = b2DefaultWorldDef();
	wd.gravity = b2Vec2_zero;
	b2WorldId w = b2CreateWorld( &wd );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = (b2BodyType)partner;
	bd.position = (b2Pos){ 7.0f, -3.0f };
	bd.rotation = b2MakeRot( 0.35f );
	bd.isAwake = loaded;
	b2BodyId a = b2CreateBody( w, &bd );
	bd.type = b2_dynamicBody;
	bd.position = (b2Pos){ 7.5f + 1.5f * partition, -2.0f };
	bd.rotation = b2MakeRot( -0.2f );
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.density = 2.0f;
	b2Circle circle = { { 0.1f, 0.2f }, 0.4f };
	b2CreateCircleShape( a, &sd, &circle );
	b2CreateCircleShape( b, &sd, &circle );
	if ( partition == 0 )
		b2Body_Disable( b );
	body_state( "beforeCreate/bodyA", a );
	body_state( "beforeCreate/bodyB", b );
	b2Transform fa = { { 0.2f, -0.3f }, b2MakeRot( 0.2f ) }, fb = { { -0.1f, 0.25f }, b2MakeRot( -0.1f ) };
	b2JointId j = create_joint( family, w, a, b, fa, fb, partition, loaded );
	assert( b2Joint_IsValid( j ) );
	observe( "created", w, j, a, b );
	if ( partition == 0 )
	{
		mutate( j, 3 );
		b2Joint_WakeBodies( j );
		observe( "createdDisabled/mutatedAndWoken", w, j, a, b );
		run_steps( "createdDisabled", 12, loaded, w, j, a, b );
		b2Body_Enable( b );
		observe( "createdDisabled/enabled", w, j, a, b );
	}
	run_steps( "initial", 60, loaded, w, j, a, b );
	const char* stages[] = { "featuresOff",	  "springOn", "limitsOn", "positiveMotor",
							 "negativeMotor", "zeroCaps", "highCaps", "retargetAndDamp" };
	for ( int stage = 0; stage < 8; ++stage )
	{
		mutate( j, stage );
		snprintf( action, sizeof( action ), "%s/afterSetters", stages[stage] );
		observe( action, w, j, a, b );
		b2Joint_WakeBodies( j );
		run_steps( stages[stage], 12, loaded, w, j, a, b );
	}
	if ( family == b2_distanceJoint || family == b2_prismaticJoint || family == b2_revoluteJoint || family == b2_wheelJoint )
	{
		const float values[] = { -0.8f, -0.5f, 0.0f, 0.5f, 0.8f };
		const float lengths[] = { 0.25f, 0.5f, 1.0f, 1.5f, 1.75f };
		const char* labels[] = { "belowLower", "atLower", "inside", "atUpper", "aboveUpper" };
		for ( int i = 0; i < 5; ++i )
		{
			float value = family == b2_distanceJoint ? lengths[i] : values[i];
			boundary( j, a, b, value );
			observe( labels[i], w, j, a, b );
			b2Joint_WakeBodies( j );
			run_steps( labels[i], 12, loaded, w, j, a, b );
		}
	}
	b2Body_SetAwake( b, false );
	body_state( "beforeSleepingMutation/bodyA", a );
	body_state( "beforeSleepingMutation/bodyB", b );
	mutate( j, 4 );
	observe( "sleepingFamilyMutation", w, j, a, b );
	b2Joint_SetConstraintTuning( j, 17.0f, 0.6f );
	b2Joint_SetForceThreshold( j, 4.0f );
	b2Joint_SetTorqueThreshold( j, 2.0f );
	fa.p = (b2Vec2){ -0.35f, 0.15f };
	fa.q = b2MakeRot( -0.4f );
	fb.p = (b2Vec2){ 0.3f, -0.2f };
	fb.q = b2MakeRot( 0.45f );
	b2Joint_SetLocalFrameA( j, fa );
	b2Joint_SetLocalFrameB( j, fb );
	b2Joint_SetUserData( j, &wd );
	ORACLE_BOOL( "userData", b2Joint_GetUserData( j ) == &wd );
	observe( "commonSetters", w, j, a, b );
	b2Joint_WakeBodies( j );
	observe( "explicitWake", w, j, a, b );
	run_steps( "changedFrames", 12, loaded, w, j, a, b );
	b2Joint_SetCollideConnected( j, true );
	observe( "collisionEnabled", w, j, a, b );
	run_steps( "collisionEnabled", 12, loaded, w, j, a, b );
	b2Joint_SetCollideConnected( j, false );
	observe( "collisionDisabled", w, j, a, b );
	run_steps( "collisionDisabled", 12, loaded, w, j, a, b );
	b2Body_Disable( b );
	observe( "bodyDisabled", w, j, a, b );
	run_steps( "bodyDisabled", 12, loaded, w, j, a, b );
	b2Body_Enable( b );
	observe( "bodyEnabled", w, j, a, b );
	run_steps( "bodyEnabled", 12, loaded, w, j, a, b );
	b2Body_SetType( a, b2_staticBody );
	b2Body_SetType( b, b2_kinematicBody );
	observe( "nonDynamicPair", w, j, a, b );
	run_steps( "nonDynamicPair", 12, false, w, j, a, b );
	b2Body_Disable( b );
	b2Body_Enable( b );
	run_steps( "nonDynamicReenabled", 12, false, w, j, a, b );
	b2Body_SetType( b, b2_dynamicBody );
	observe( "dynamicRestored", w, j, a, b );
	run_steps( "dynamicRestored", 12, loaded, w, j, a, b );
	b2DestroyJoint( j );
	ORACLE_I32( "jointDestroyed/count", b2World_GetCounters( w ).jointCount );
	b2Body_SetLinearVelocity( b, (b2Vec2){ 2.0f, -1.0f } );
	b2Body_SetAngularVelocity( b, 1.5f );
	for ( int i = 1; i <= 12; ++i )
	{
		b2World_Step( w, 1.0f / 60.0f, 4 );
		if ( i == 1 || i == 12 )
		{
			snprintf( action, sizeof( action ), "freeMotion/step%d", i );
			body_state( action, b );
		}
	}
	j = create_joint( family, w, a, b, fa, fb, partition, loaded );
	observe( "recreated", w, j, a, b );
	b2DestroyBody( b );
	ORACLE_I32( "bodyDestroyed/jointCount", b2World_GetCounters( w ).jointCount );
	b2World_Step( w, 1.0f / 60.0f, 4 );
	body_state( "survivor", a );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
int main( void )
{
	for ( int family = 0; family < 9; ++family )
		for ( int partner = 0; partner < 3; ++partner )
			for ( int loaded = 0; loaded < 2; ++loaded )
				for ( int partition = 0; partition < 3; ++partition )
					scenario( (b2JointType)family, partner, loaded != 0, partition );
	for ( int family = 0; family < 9; ++family )
		collision_policy( (b2JointType)family );
	return 0;
}

// Overlapping real shapes make collision policy observable independently of a
// joint's later trajectory. Full shape equality includes world and generation.
static void collision_observe( const char* phase, b2WorldId w, b2BodyId a, b2BodyId b, b2ShapeId sa, b2ShapeId sb )
{
	char key[160];
	ORACLE_I32( ORACLE_KEY( "%s/contactCount", phase ), b2World_GetCounters( w ).contactCount );
	ORACLE_I32( ORACLE_KEY( "%s/jointCount", phase ), b2World_GetCounters( w ).jointCount );
	body_state( ORACLE_KEY( "%s/bodyA", phase ), a );
	body_state( ORACLE_KEY( "%s/bodyB", phase ), b );
	b2ContactData contacts[2];
	int count = b2Body_GetContactData( a, contacts, 2 );
	ORACLE_I32( ORACLE_KEY( "%s/touchingCount", phase ), count );
	for ( int i = 0; i < count; ++i )
	{
		snprintf( key, sizeof( key ), "%s/contact%d", phase, i );
		ORACLE_BOOL( ORACLE_KEY( "%s/shapeA", key ), B2_ID_EQUALS( contacts[i].shapeIdA, sa ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/shapeB", key ), B2_ID_EQUALS( contacts[i].shapeIdB, sb ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/valid", key ), b2Contact_IsValid( contacts[i].contactId ) );
		b2Manifold m = contacts[i].manifold;
		ORACLE_VEC2( ORACLE_KEY( "%s/normal", key ), m.normal );
		ORACLE_I32( ORACLE_KEY( "%s/pointCount", key ), m.pointCount );
		for ( int p = 0; p < m.pointCount; ++p )
		{
			ORACLE_F32( ORACLE_KEY( "%s/point%d/separation", key, p ), m.points[p].separation );
			ORACLE_F32( ORACLE_KEY( "%s/point%d/normalImpulse", key, p ), m.points[p].normalImpulse );
		}
	}
}
static void collision_policy( b2JointType family )
{
	char name[64];
	snprintf( name, sizeof( name ), "joints/collision/%d", family );
	ORACLE_OPEN( name );
	b2WorldDef wd = b2DefaultWorldDef();
	wd.gravity = b2Vec2_zero;
	b2WorldId w = b2CreateWorld( &wd );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = (b2Pos){ 7.0f, -3.0f };
	b2BodyId a = b2CreateBody( w, &bd );
	bd.position = (b2Pos){ 7.5f, -3.0f };
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	b2Circle c = { { 0.0f, 0.0f }, 1.0f };
	b2ShapeId sa = b2CreateCircleShape( a, &sd, &c ), sb = b2CreateCircleShape( b, &sd, &c );
	b2World_Step( w, 1.0f / 60.0f, 4 );
	collision_observe( "beforeJoint", w, a, b, sa, sb );
	b2Body_SetAwake( a, false );
	b2Body_SetAwake( b, false );
	b2JointId j = create_joint( family, w, a, b, b2Transform_identity, b2Transform_identity, 1, false );
	collision_observe( "createdSleepingPair", w, a, b, sa, sb );
	b2World_Step( w, 1.0f / 60.0f, 4 );
	collision_observe( "suppressed", w, a, b, sa, sb );
	const char* actions[] = { "allow", "reject", "allowAgain" };
	for ( int i = 0; i < 3; ++i )
	{
		b2Body_SetTransform( a, (b2Pos){ 7.0f, -3.0f }, b2Rot_identity );
		b2Body_SetTransform( b, (b2Pos){ 7.5f, -3.0f }, b2Rot_identity );
		b2Body_SetLinearVelocity( a, b2Vec2_zero );
		b2Body_SetLinearVelocity( b, b2Vec2_zero );
		b2Joint_SetCollideConnected( j, i != 1 );
		b2Joint_WakeBodies( j );
		ORACLE_BOOL( ORACLE_KEY( "%s/collideConnected", actions[i] ), b2Joint_GetCollideConnected( j ) );
		b2World_Step( w, 1.0f / 60.0f, 4 );
		collision_observe( actions[i], w, a, b, sa, sb );
	}
	b2DestroyJoint( j );
	b2Body_SetTransform( a, (b2Pos){ 7.0f, -3.0f }, b2Rot_identity );
	b2Body_SetTransform( b, (b2Pos){ 7.5f, -3.0f }, b2Rot_identity );
	b2World_Step( w, 1.0f / 60.0f, 4 );
	collision_observe( "afterDestroy", w, a, b, sa, sb );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
