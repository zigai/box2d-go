#include "dynamics_observe.h"

#include <assert.h>
#include <math.h>

static void dynamics_open( const char* name )
{
	ORACLE_OPEN( name );
	ORACLE_BYTES( "dynamics/schema", dynamics_schema, sizeof( dynamics_schema ) - 1 );
}
static b2BodyId dynamics_box( b2WorldId w, b2BodyType type, float x, float y, float hx, float hy, b2SurfaceMaterial material,
							  b2ShapeId* shape )
{
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = type;
	bd.position = (b2Pos){ x, y };
	b2BodyId body = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.material = material;
	sd.enableContactEvents = true;
	sd.enableHitEvents = true;
	b2Polygon box = b2MakeBox( hx, hy );
	b2ShapeId s = b2CreatePolygonShape( body, &sd, &box );
	if ( shape )
		*shape = s;
	return body;
}
static void dynamics_steps( b2WorldId w, int n, const char* prefix )
{
	for ( int i = 0; i < n; ++i )
	{
		b2World_Step( w, 1.0f / 60.0f, 4 );
		char key[128];
		snprintf( key, sizeof( key ), "%sstep/%d/", prefix, i + 1 );
		dynamics_short( w, key, true );
	}
}
static float restitution_mix( float a, uint64_t ia, float b, uint64_t ib )
{
	(void)a;
	(void)b;
	return ia == 7 || ib == 7 ? .35f : .65f;
}
static float friction_mix( float a, uint64_t ia, float b, uint64_t ib )
{
	(void)a;
	(void)b;
	return ia == 7 || ib == 7 ? .1f : .8f;
}

static void integration( void )
{
	for ( int partition = 0; partition < 12; ++partition )
	{
		dynamics_open( ORACLE_KEY( "dynamics/integration%d", partition ) );
		b2WorldDef wd = b2DefaultWorldDef();
		wd.gravity = (b2Vec2){ 2, -10 };
		wd.enableSleep = false;
		wd.maximumLinearSpeed = 10;
		b2WorldId w = b2CreateWorld( &wd );
		b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
		b2BodyId b = dynamics_box( w, b2_dynamicBody, 3, 4, .6f, .4f, m, NULL );
		b2Body_SetLinearVelocity( b, (b2Vec2){ 3, 2 } );
		b2Body_SetAngularVelocity( b, 2 );
		switch ( partition )
		{
			case 0:
				b2Body_SetGravityScale( b, 0 );
				break;
			case 1:
				b2Body_SetGravityScale( b, -1 );
				break;
			case 2:
				b2Body_SetLinearDamping( b, .7f );
				b2Body_SetAngularDamping( b, 1.3f );
				break;
			case 3:
				b2Body_ApplyForce( b, (b2Vec2){ 8, 11 }, (b2Pos){ 4, 5 }, true );
				break;
			case 4:
				b2Body_ApplyTorque( b, 9, true );
				break;
			case 5:
				b2Body_ApplyLinearImpulse( b, (b2Vec2){ 2, 3 }, (b2Pos){ 4, 5 }, true );
				break;
			case 6:
				b2Body_ApplyAngularImpulse( b, 3, true );
				break;
			case 7:
				b2Body_SetMotionLocks( b, (b2MotionLocks){ .linearX = true } );
				break;
			case 8:
				b2Body_SetMotionLocks( b, (b2MotionLocks){ .linearY = true, .angularZ = true } );
				break;
			default:
			{
				float speed = 10;
				if ( partition == 9 )
					speed = nextafterf( speed, -INFINITY );
				if ( partition == 11 )
					speed = nextafterf( speed, INFINITY );
				b2Body_SetGravityScale( b, 0 );
				b2Body_SetLinearVelocity( b, (b2Vec2){ speed, 0 } );
				break;
			}
		}
		dynamics_short( w, "initial/", true );
		dynamics_steps( w, 30, "" );
		b2World_SetGravity( w, (b2Vec2){ -3, 4 } );
		b2World_SetMaximumLinearSpeed( w, 5 );
		b2Body_SetLinearDamping( b, .2f );
		b2Body_SetAngularDamping( b, .4f );
		dynamics_steps( w, 12, "mutated/" );
		b2DestroyWorld( w );
		ORACLE_CLOSE();
	}
}
static void contact_tuning( void )
{
	for ( int partition = 0; partition < 12; ++partition )
	{
		dynamics_open( ORACLE_KEY( "dynamics/contact%d", partition ) );
		b2WorldDef wd = b2DefaultWorldDef();
		wd.enableSleep = false;
		b2WorldId w = b2CreateWorld( &wd );
		b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
		m.friction = .6f;
		b2ShapeId ground, shape;
		dynamics_box( w, b2_staticBody, 0, -1, 20, 1, m, &ground );
		b2BodyId b = dynamics_box( w, b2_dynamicBody, 0, .51f, .5f, .5f, m, &shape );
		dynamics_box( w, b2_dynamicBody, 0, 1.53f, .5f, .5f, m, NULL );
		b2Body_SetLinearVelocity( b, (b2Vec2){ 5, 0 } );
		b2Body_SetAngularVelocity( b, 1 );
		switch ( partition )
		{
			case 0:
				b2World_EnableWarmStarting( w, false );
				break;
			case 1:
				b2World_SetContactTuning( w, 5, .2f, .5f );
				break;
			case 2:
				b2World_SetContactTuning( w, 120, 3, 10 );
				break;
			case 3:
				b2Shape_SetFriction( shape, 0 );
				break;
			case 4:
				m.rollingResistance = .7f;
				b2Shape_SetSurfaceMaterial( shape, &m );
				break;
			case 5:
				m.tangentSpeed = 3;
				b2Shape_SetSurfaceMaterial( shape, &m );
				break;
			case 6:
				b2World_SetFrictionCallback( w, friction_mix );
				break;
			case 7:
				m.rollingResistance = .4f;
				b2Shape_SetSurfaceMaterial( shape, &m );
				break;
			case 8:
				b2Body_EnableContactRecycling( b, false );
				break;
			case 9:
				b2World_SetContactTuning( w, 30, 1, 0 );
				break;
			case 10:
				b2World_SetMaximumLinearSpeed( w, 2 );
				break;
			case 11:
				b2World_SetGravity( w, (b2Vec2){ 2, -5 } );
				break;
		}
		dynamics_steps( w, 30, "" );
		b2Shape_SetFriction( ground, .1f );
		m = b2Shape_GetSurfaceMaterial( shape );
		m.friction = .9f;
		m.rollingResistance = .2f;
		m.tangentSpeed = -2;
		b2Shape_SetSurfaceMaterial( shape, &m );
		b2World_EnableWarmStarting( w, true );
		b2World_SetContactTuning( w, 15, .5f, 2 );
		b2World_SetFrictionCallback( w, NULL );
		dynamics_steps( w, 30, "mutated/" );
		b2DestroyWorld( w );
		ORACLE_CLOSE();
	}
}
static void restitution_boundary( void )
{
	for ( int thresholdCase = 0; thresholdCase < 2; ++thresholdCase )
		for ( int neighbor = 0; neighbor < 3; ++neighbor )
		{
			dynamics_open( ORACLE_KEY( "dynamics/threshold%d_neighbor%d", thresholdCase, neighbor ) );
			b2WorldDef wd = b2DefaultWorldDef();
			wd.gravity = (b2Vec2){ 0, 0 };
			wd.enableSleep = false;
			wd.restitutionIterations = 4;
			if ( thresholdCase == 1 )
				wd.restitutionThreshold = 2.5f;
			float speed = wd.restitutionThreshold;
			if ( neighbor == 0 )
				speed = nextafterf( speed, -INFINITY );
			if ( neighbor == 2 )
				speed = nextafterf( speed, INFINITY );
			b2WorldId w = b2CreateWorld( &wd );
			b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
			m.friction = 0;
			m.restitution = .8f;
			dynamics_box( w, b2_staticBody, 0, -1, 20, 1, m, NULL );
			b2BodyDef bd = b2DefaultBodyDef();
			bd.type = b2_dynamicBody;
			bd.position = (b2Pos){ 0, .5f };
			bd.linearVelocity = (b2Vec2){ 0, -speed };
			b2BodyId b = b2CreateBody( w, &bd );
			b2ShapeDef sd = b2DefaultShapeDef();
			sd.material = m;
			sd.enableHitEvents = true;
			sd.enableContactEvents = true;
			b2Circle circle = { .radius = .5f };
			b2CreateCircleShape( b, &sd, &circle );
			ORACLE_F32( "approachSpeed", speed );
			dynamics_steps( w, 60, "" );
			b2DestroyWorld( w );
			ORACLE_CLOSE();
		}
	for ( int mode = 0; mode < 3; ++mode )
		for ( int swapped = 0; swapped < 2; ++swapped )
		{
			dynamics_open( ORACLE_KEY( "dynamics/mixing%d_swapped%d", mode, swapped ) );
			b2WorldDef wd = b2DefaultWorldDef();
			wd.gravity = (b2Vec2){ 0, 0 };
			wd.enableSleep = false;
			wd.restitutionIterations = 4;
			b2WorldId w = b2CreateWorld( &wd );
			if ( mode != 0 )
				b2World_SetRestitutionCallback( w, restitution_mix );
			if ( mode == 2 )
				b2World_SetRestitutionCallback( w, NULL );
			float a = .2f, b = .8f;
			if ( swapped == 1 )
			{
				a = .8f;
				b = .2f;
			}
			b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
			m.friction = 0;
			m.restitution = a;
			m.userMaterialId = 7;
			dynamics_box( w, b2_staticBody, 0, -1, 20, 1, m, NULL );
			m.restitution = b;
			m.userMaterialId = 11;
			b2BodyId ball = dynamics_box( w, b2_dynamicBody, 0, .6f, .5f, .5f, m, NULL );
			b2Body_SetLinearVelocity( ball, (b2Vec2){ 0, -5 } );
			dynamics_steps( w, 60, "" );
			b2DestroyWorld( w );
			ORACLE_CLOSE();
		}
	for ( int propagation = 0; propagation < 2; ++propagation )
		for ( int iterationCase = 0; iterationCase < 3; ++iterationCase )
			for ( int supportsCase = 0; supportsCase < 2; ++supportsCase )
			{
				dynamics_open(
					ORACLE_KEY( "dynamics/propagation%d_iterations%d_supports%d", propagation, iterationCase, supportsCase ) );
				b2WorldDef wd = b2DefaultWorldDef();
				wd.gravity = (b2Vec2){ 0, 0 };
				wd.enableSleep = false;
				wd.restitutionIterations = 1 << iterationCase;
				wd.enableRestitutionPropagation = propagation != 0;
				b2WorldId w = b2CreateWorld( &wd );
				b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
				m.friction = 0;
				m.restitution = .8f;
				dynamics_box( w, b2_staticBody, 0, -1, 20, 1, m, NULL );
				int supports = 1 + 2 * supportsCase;
				for ( int i = 0; i < supports; ++i )
					dynamics_box( w, b2_dynamicBody, 0, .5f + (float)i, .5f, .5f, m, NULL );
				b2BodyId ball = dynamics_box( w, b2_dynamicBody, 0, (float)supports + .6f, .5f, .5f, m, NULL );
				b2Body_SetLinearVelocity( ball, (b2Vec2){ 0, -5 } );
				dynamics_steps( w, 60, "" );
				b2DestroyWorld( w );
				ORACLE_CLOSE();
			}
}
static void continuous( void )
{
	for ( int target = 0; target < 3; ++target )
		for ( int bullet = 0; bullet < 2; ++bullet )
			for ( int enabled = 0; enabled < 2; ++enabled )
				for ( int geometry = 0; geometry < 4; ++geometry )
				{
					dynamics_open(
						ORACLE_KEY( "dynamics/target%d_bullet%d_continuous%d_geometry%d", target, bullet, enabled, geometry ) );
					b2WorldDef wd = b2DefaultWorldDef();
					wd.gravity = (b2Vec2){ 0, 0 };
					wd.enableSleep = false;
					wd.enableContinuous = enabled != 0;
					b2WorldId w = b2CreateWorld( &wd );
					b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
					m.friction = .2f;
					b2BodyId targetBody = dynamics_box( w, (b2BodyType)target, 0, 0, .05f, 3, m, NULL );
					if ( target != 0 )
						b2Body_SetLinearVelocity( targetBody, (b2Vec2){ -1, 0 } );
					float x = -2;
					if ( geometry == 1 )
						x = -.1f;
					b2BodyDef bd = b2DefaultBodyDef();
					bd.type = b2_dynamicBody;
					bd.position = (b2Pos){ x, 0 };
					bd.linearVelocity = (b2Vec2){ 80, 0 };
					bd.isBullet = bullet != 0;
					bd.safetyFactor = geometry == 0 ? .01f : .5f;
					if ( geometry == 2 )
						bd.angularVelocity = 10;
					b2BodyId b = b2CreateBody( w, &bd );
					b2ShapeDef sd = b2DefaultShapeDef();
					sd.material = m;
					sd.enableHitEvents = true;
					sd.enableContactEvents = true;
					b2Polygon box = b2MakeBox( .15f, .4f );
					b2CreatePolygonShape( b, &sd, &box );
					dynamics_steps( w, 30, "" );
					b2World_EnableContinuous( w, true );
					b2Body_SetBullet( b, true );
					b2Body_SetSafetyFactor( b, .1f );
					dynamics_steps( w, 12, "mutated/" );
					b2DestroyWorld( w );
					ORACLE_CLOSE();
				}
}
static void sleep_wake( void )
{
	for ( int operation = 0; operation < 12; ++operation )
		for ( int wake = 0; wake < 2; ++wake )
		{
			dynamics_open( ORACLE_KEY( "dynamics/sleep_operation%d_wake%d", operation, wake ) );
			b2WorldDef wd = b2DefaultWorldDef();
			b2WorldId w = b2CreateWorld( &wd );
			b2SurfaceMaterial m = b2DefaultSurfaceMaterial();
			m.friction = .6f;
			dynamics_box( w, b2_staticBody, 0, -1, 20, 1, m, NULL );
			b2BodyId lower = dynamics_box( w, b2_dynamicBody, 0, .51f, .5f, .5f, m, NULL );
			b2BodyId upper = dynamics_box( w, b2_dynamicBody, 0, 1.53f, .5f, .5f, m, NULL );
			bool asleep = false;
			for ( int i = 0; i < 900; ++i )
			{
				b2World_Step( w, 1.0f / 60.0f, 4 );
				char key[128];
				snprintf( key, sizeof( key ), "settle/%d/", i + 1 );
				dynamics_short( w, key, i == 0 || ( i + 1 ) % 60 == 0 );
				if ( !b2Body_IsAwake( lower ) && !b2Body_IsAwake( upper ) )
				{
					asleep = true;
					ORACLE_I32( "sleepStep", i + 1 );
					break;
				}
			}
			assert( asleep );
			switch ( operation )
			{
				case 0:
					b2Body_ApplyForceToCenter( lower, (b2Vec2){ 20, 5 }, wake != 0 );
					break;
				case 1:
					b2Body_ApplyForce( lower, (b2Vec2){ 20, 5 }, (b2Pos){ 1, 1 }, wake != 0 );
					break;
				case 2:
					b2Body_ApplyTorque( lower, 10, wake != 0 );
					break;
				case 3:
					b2Body_ApplyLinearImpulseToCenter( lower, (b2Vec2){ 2, 3 }, wake != 0 );
					break;
				case 4:
					b2Body_ApplyLinearImpulse( lower, (b2Vec2){ 2, 3 }, (b2Pos){ 1, 1 }, wake != 0 );
					break;
				case 5:
					b2Body_ApplyAngularImpulse( lower, 2, wake != 0 );
					break;
				case 6:
					b2Body_SetAwake( lower, true );
					break;
				case 7:
					b2Body_EnableSleep( lower, false );
					break;
				case 8:
					b2World_EnableSleeping( w, false );
					break;
				case 9:
					b2Body_Disable( lower );
					break;
				case 10:
					b2Body_SetType( lower, b2_kinematicBody );
					b2Body_SetLinearVelocity( lower, (b2Vec2){ 1, 0 } );
					break;
				case 11:
					b2Body_SetTransform( lower, (b2Pos){ 1, .5f }, (b2Rot){ 1, 0 } );
					break;
			}
			dynamics_short( w, "operation/", true );
			dynamics_steps( w, 12, "after/" );
			b2Body_Enable( lower );
			b2Body_SetType( lower, b2_dynamicBody );
			b2Body_SetAwake( lower, true );
			dynamics_steps( w, 12, "restored/" );
			b2DestroyWorld( w );
			ORACLE_CLOSE();
		}
}
int main( void )
{
	integration();
	contact_tuning();
	restitution_boundary();
	continuous();
	sleep_wake();
	return 0;
}
