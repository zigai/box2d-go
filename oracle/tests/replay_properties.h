// SPDX-License-Identifier: MIT
#pragma once
#include "body.h"
#include "joint.h"
#include "physics_world.h"
#include "shape.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
// Large scenes retain every typed observation, grouped by named field. Each
// digest includes the full relative key (hence object identity and geometry
// ordinal), type, and defined scalar bits. No native object memory is hashed.
typedef struct replay_property_field
{
	char name[128];
	uint64_t hash;
} replay_property_field;
typedef struct replay_property_fields
{
	const char* prefix;
	bool compact;
	int count;
	replay_property_field fields[256];
	unsigned short slots[512];
} replay_property_fields;

static uint64_t* replay_property_hash( replay_property_fields* fields, const char* key, const char* type )
{
	const char* relative = key + strlen( fields->prefix ) + 1;
	char family[128];
	const char* slash = strchr( relative, '/' );
	size_t stem = 0;
	if ( strncmp( relative, "body", 4 ) == 0 )
		stem = 4;
	else if ( strncmp( relative, "shape", 5 ) == 0 )
		stem = 5;
	else if ( strncmp( relative, "joint", 5 ) == 0 )
		stem = 5;
	if ( stem && slash )
	{
		snprintf( family, sizeof( family ), "%.*s%s", (int)stem, relative, slash );
	}
	else
	{
		snprintf( family, sizeof( family ), "%s", relative );
	}
	unsigned slot = (unsigned)( oracle_hash_bytes( ORACLE_HASH_INIT, family, strlen( family ) ) & 511 );
	while ( fields->slots[slot] && strcmp( fields->fields[fields->slots[slot] - 1].name, family ) != 0 )
	{
		slot = ( slot + 1 ) & 511;
	}
	int i = fields->slots[slot] ? fields->slots[slot] - 1 : fields->count;
	if ( i == fields->count )
	{
		if ( fields->count == 256 )
			abort();
		snprintf( fields->fields[i].name, sizeof( fields->fields[i].name ), "%s", family );
		fields->fields[i].hash = ORACLE_HASH_INIT;
		++fields->count;
		fields->slots[slot] = (unsigned short)( i + 1 );
	}
	uint64_t* hash = &fields->fields[i].hash;
	*hash = oracle_hash_u32( *hash, (uint32_t)strlen( relative ) );
	*hash = oracle_hash_bytes( *hash, relative, strlen( relative ) );
	*hash = oracle_hash_bytes( *hash, type, strlen( type ) + 1 );
	return hash;
}

static void REPLAY_PROPERTY_BOOL( replay_property_fields* fields, const char* key, bool value )
{
	if ( !fields->compact )
	{
		ORACLE_BOOL( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "bool" );
	*hash = oracle_hash_u32( *hash, value );
}
static void REPLAY_PROPERTY_I32( replay_property_fields* fields, const char* key, int32_t value )
{
	if ( !fields->compact )
	{
		ORACLE_I32( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "i32" );
	*hash = oracle_hash_u32( *hash, value );
}
static void REPLAY_PROPERTY_U32( replay_property_fields* fields, const char* key, uint32_t value )
{
	if ( !fields->compact )
	{
		ORACLE_U32( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "u32" );
	*hash = oracle_hash_u32( *hash, value );
}
static void REPLAY_PROPERTY_U64( replay_property_fields* fields, const char* key, uint64_t value )
{
	if ( !fields->compact )
	{
		ORACLE_U64( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "u64" );
	*hash = oracle_hash_u64( *hash, value );
}
static void REPLAY_PROPERTY_F32( replay_property_fields* fields, const char* key, float value )
{
	if ( !fields->compact )
	{
		ORACLE_F32( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "f32" );
	*hash = oracle_hash_f32( *hash, value );
}
static void REPLAY_PROPERTY_ROT( replay_property_fields* fields, const char* key, b2Rot value )
{
	if ( !fields->compact )
	{
		ORACLE_ROT( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "rot" );
	*hash = oracle_hash_rot( *hash, value );
}
static void REPLAY_PROPERTY_VEC2( replay_property_fields* fields, const char* key, b2Vec2 value )
{
	if ( !fields->compact )
	{
		ORACLE_VEC2( key, value );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "vec2" );
	*hash = oracle_hash_f32( oracle_hash_f32( *hash, value.x ), value.y );
}
static void REPLAY_PROPERTY_BYTES( replay_property_fields* fields, const char* key, const void* value, size_t size )
{
	if ( !fields->compact )
	{
		ORACLE_BYTES( key, value, size );
		return;
	}
	uint64_t* hash = replay_property_hash( fields, key, "bytes" );
	*hash = oracle_hash_bytes( oracle_hash_u64( *hash, size ), value, size );
}
static void replay_property_end( replay_property_fields* fields )
{
	if ( !fields->compact )
		return;
	ORACLE_I32( ORACLE_KEY( "%s/digest/fieldCount", fields->prefix ), fields->count );
	for ( int i = 0; i < fields->count; ++i )
	{
		ORACLE_U64( ORACLE_KEY( "%s/digest/%s", fields->prefix, fields->fields[i].name ), fields->fields[i].hash );
	}
}

static void replay_properties( b2WorldId world, const char* prefix )
{
	char owned[256];
	snprintf( owned, sizeof( owned ), "%s", prefix );
	prefix = owned;
	b2World* w = b2GetWorldFromId( world );
	replay_property_fields fields = { .prefix = prefix, .compact = b2World_GetCounters( world ).bodyCount > 32 };
	REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSleepingEnabled", prefix ), b2World_IsSleepingEnabled( world ) );
	REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsContinuousEnabled", prefix ), b2World_IsContinuousEnabled( world ) );
	REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsWarmStartingEnabled", prefix ), b2World_IsWarmStartingEnabled( world ) );
	REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsRestitutionPropagationEnabled", prefix ),
						  b2World_IsRestitutionPropagationEnabled( world ) );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/RestitutionThreshold", prefix ), b2World_GetRestitutionThreshold( world ) );
	REPLAY_PROPERTY_I32( &fields, ORACLE_KEY( "%s/RestitutionIterations", prefix ), b2World_GetRestitutionIterations( world ) );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/HitEventThreshold", prefix ), b2World_GetHitEventThreshold( world ) );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/ContactRecycleDistance", prefix ), b2World_GetContactRecycleDistance( world ) );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaximumLinearSpeed", prefix ), b2World_GetMaximumLinearSpeed( world ) );
	REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/Gravity", prefix ), b2World_GetGravity( world ) );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/contactHertz", prefix ), w->contactHertz );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/contactDamping", prefix ), w->contactDampingRatio );
	REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/contactSpeed", prefix ), w->contactSpeed );
	for ( int i = 0; i < w->bodies.count; ++i )
	{
		if ( w->bodies.data[i].id != i )
			continue;
		b2BodyId id = b2MakeBodyId( w, i );
		char key[320];
		snprintf( key, sizeof( key ), "%s/body%d", prefix, id.index1 );
		REPLAY_PROPERTY_U32( &fields, ORACLE_KEY( "%s/generation", key ), id.generation );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/owner", key ),
							  b2StoreWorldId( b2Body_GetWorld( id ) ) == b2StoreWorldId( world ) );
		REPLAY_PROPERTY_I32( &fields, ORACLE_KEY( "%s/Type", key ), b2Body_GetType( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/Mass", key ), b2Body_GetMass( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/RotationalInertia", key ), b2Body_GetRotationalInertia( id ) );
		REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/LocalCenter", key ), b2Body_GetLocalCenter( id ) );
		REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/LinearVelocity", key ), b2Body_GetLinearVelocity( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularVelocity", key ), b2Body_GetAngularVelocity( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LinearDamping", key ), b2Body_GetLinearDamping( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularDamping", key ), b2Body_GetAngularDamping( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/GravityScale", key ), b2Body_GetGravityScale( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SleepThreshold", key ), b2Body_GetSleepThreshold( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SafetyFactor", key ), b2Body_GetSafetyFactor( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsEnabled", key ), b2Body_IsEnabled( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsAwake", key ), b2Body_IsAwake( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSleepEnabled", key ), b2Body_IsSleepEnabled( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsBullet", key ), b2Body_IsBullet( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsContactRecyclingEnabled", key ),
							  b2Body_IsContactRecyclingEnabled( id ) );
		const char* name = b2Body_GetName( id );
		REPLAY_PROPERTY_BYTES( &fields, ORACLE_KEY( "%s/name", key ), name, strlen( name ) );
		b2MotionLocks locks = b2Body_GetMotionLocks( id );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/lockX", key ), locks.linearX );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/lockY", key ), locks.linearY );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/lockZ", key ), locks.angularZ );
	}
	for ( int i = 0; i < w->shapes.count; ++i )
	{
		if ( w->shapes.data[i].id != i )
			continue;
		b2ShapeId id = (b2ShapeId){ i + 1, (uint16_t)( world.index1 - 1 ), w->shapes.data[i].generation };
		char key[320];
		snprintf( key, sizeof( key ), "%s/shape%d", prefix, id.index1 );
		REPLAY_PROPERTY_U32( &fields, ORACLE_KEY( "%s/generation", key ), id.generation );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/owner", key ),
							  b2StoreWorldId( b2Shape_GetWorld( id ) ) == b2StoreWorldId( world ) );
		REPLAY_PROPERTY_I32( &fields, ORACLE_KEY( "%s/Type", key ), b2Shape_GetType( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/Density", key ), b2Shape_GetDensity( id ) );
		REPLAY_PROPERTY_U64( &fields, ORACLE_KEY( "%s/UserMaterial", key ), b2Shape_GetUserMaterial( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSensor", key ), b2Shape_IsSensor( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/AreSensorEventsEnabled", key ), b2Shape_AreSensorEventsEnabled( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/AreContactEventsEnabled", key ), b2Shape_AreContactEventsEnabled( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/AreHitEventsEnabled", key ), b2Shape_AreHitEventsEnabled( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/ArePreSolveEventsEnabled", key ), b2Shape_ArePreSolveEventsEnabled( id ) );
		b2SurfaceMaterial m = b2Shape_GetSurfaceMaterial( id );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/material/Friction", key ), m.friction );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/material/Restitution", key ), m.restitution );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/material/RollingResistance", key ), m.rollingResistance );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/material/TangentSpeed", key ), m.tangentSpeed );
		REPLAY_PROPERTY_U64( &fields, ORACLE_KEY( "%s/material/UserMaterialID", key ), m.userMaterialId );
		REPLAY_PROPERTY_U32( &fields, ORACLE_KEY( "%s/material/CustomColor", key ), m.customColor );
		b2Filter f = b2Shape_GetFilter( id );
		REPLAY_PROPERTY_U64( &fields, ORACLE_KEY( "%s/filter/category", key ), f.categoryBits );
		REPLAY_PROPERTY_U64( &fields, ORACLE_KEY( "%s/filter/mask", key ), f.maskBits );
		REPLAY_PROPERTY_I32( &fields, ORACLE_KEY( "%s/filter/group", key ), f.groupIndex );
		switch ( b2Shape_GetType( id ) )
		{
			case b2_circleShape:
			{
				b2Circle v = b2Shape_GetCircle( id );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Center", key ), v.center );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/geometry/Radius", key ), v.radius );
				break;
			}
			case b2_capsuleShape:
			{
				b2Capsule v = b2Shape_GetCapsule( id );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Center1", key ), v.center1 );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Center2", key ), v.center2 );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/geometry/Radius", key ), v.radius );
				break;
			}
			case b2_segmentShape:
			{
				b2Segment v = b2Shape_GetSegment( id );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Point1", key ), v.point1 );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Point2", key ), v.point2 );
				break;
			}
			case b2_chainSegmentShape:
			{
				b2ChainSegment v = b2Shape_GetChainSegment( id );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Ghost1", key ), v.ghost1 );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Segment.Point1", key ), v.segment.point1 );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Segment.Point2", key ), v.segment.point2 );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/Ghost2", key ), v.ghost2 );
				break;
			}
			case b2_polygonShape:
			{
				b2Polygon v = b2Shape_GetPolygon( id );
				REPLAY_PROPERTY_I32( &fields, ORACLE_KEY( "%s/geometry/count", key ), v.count );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/geometry/radius", key ), v.radius );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/centroid", key ), v.centroid );
				for ( int j = 0; j < v.count; ++j )
				{
					REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/v%d", key, j ), v.vertices[j] );
					REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/geometry/n%d", key, j ), v.normals[j] );
				}
				break;
			}
			default:
				abort();
		}
	}
	for ( int i = 0; i < w->joints.count; ++i )
	{
		if ( w->joints.data[i].jointId != i )
			continue;
		b2JointId id = (b2JointId){ i + 1, (uint16_t)( world.index1 - 1 ), w->joints.data[i].generation };
		char key[320];
		snprintf( key, sizeof( key ), "%s/joint%d", prefix, id.index1 );
		REPLAY_PROPERTY_U32( &fields, ORACLE_KEY( "%s/generation", key ), id.generation );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/owner", key ),
							  b2StoreWorldId( b2Joint_GetWorld( id ) ) == b2StoreWorldId( world ) );
		REPLAY_PROPERTY_I32( &fields, ORACLE_KEY( "%s/Type", key ), b2Joint_GetType( id ) );
		REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/CollideConnected", key ), b2Joint_GetCollideConnected( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/ForceThreshold", key ), b2Joint_GetForceThreshold( id ) );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/TorqueThreshold", key ), b2Joint_GetTorqueThreshold( id ) );
		b2Transform a = b2Joint_GetLocalFrameA( id ), b = b2Joint_GetLocalFrameB( id );
		REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/frameA/p", key ), a.p );
		REPLAY_PROPERTY_ROT( &fields, ORACLE_KEY( "%s/frameA/q", key ), a.q );
		REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/frameB/p", key ), b.p );
		REPLAY_PROPERTY_ROT( &fields, ORACLE_KEY( "%s/frameB/q", key ), b.q );
		float hertz, damping;
		b2Joint_GetConstraintTuning( id, &hertz, &damping );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/hertz", key ), hertz );
		REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/damping", key ), damping );
		switch ( b2Joint_GetType( id ) )
		{
			case b2_distanceJoint:
			{
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/Length", key ), b2DistanceJoint_GetLength( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MinLength", key ), b2DistanceJoint_GetMinLength( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxLength", key ), b2DistanceJoint_GetMaxLength( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringHertz", key ), b2DistanceJoint_GetSpringHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringDampingRatio", key ),
									 b2DistanceJoint_GetSpringDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MotorSpeed", key ), b2DistanceJoint_GetMotorSpeed( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxMotorForce", key ), b2DistanceJoint_GetMaxMotorForce( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSpringEnabled", key ), b2DistanceJoint_IsSpringEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsLimitEnabled", key ), b2DistanceJoint_IsLimitEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsMotorEnabled", key ), b2DistanceJoint_IsMotorEnabled( id ) );
				float low, high;
				b2DistanceJoint_GetSpringForceRange( id, &low, &high );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/forceLow", key ), low );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/forceHigh", key ), high );
				break;
			}
			case b2_motorJoint:
			{
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/LinearVelocity", key ), b2MotorJoint_GetLinearVelocity( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularVelocity", key ), b2MotorJoint_GetAngularVelocity( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxVelocityForce", key ), b2MotorJoint_GetMaxVelocityForce( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxVelocityTorque", key ),
									 b2MotorJoint_GetMaxVelocityTorque( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LinearHertz", key ), b2MotorJoint_GetLinearHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LinearDampingRatio", key ),
									 b2MotorJoint_GetLinearDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularHertz", key ), b2MotorJoint_GetAngularHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularDampingRatio", key ),
									 b2MotorJoint_GetAngularDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxSpringForce", key ), b2MotorJoint_GetMaxSpringForce( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxSpringTorque", key ), b2MotorJoint_GetMaxSpringTorque( id ) );
				break;
			}
			case b2_moverJoint:
			{
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/LinearVelocity", key ), b2MoverJoint_GetLinearVelocity( id ) );
				REPLAY_PROPERTY_VEC2( &fields, ORACLE_KEY( "%s/MaxVelocityForce", key ), b2MoverJoint_GetMaxVelocityForce( id ) );
				break;
			}
			case b2_pogoJoint:
			{
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/RestLength", key ), b2PogoJoint_GetRestLength( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringHertz", key ), b2PogoJoint_GetSpringHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringDampingRatio", key ),
									 b2PogoJoint_GetSpringDampingRatio( id ) );
				break;
			}
			case b2_prismaticJoint:
			{
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringHertz", key ), b2PrismaticJoint_GetSpringHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringDampingRatio", key ),
									 b2PrismaticJoint_GetSpringDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/TargetTranslation", key ),
									 b2PrismaticJoint_GetTargetTranslation( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LowerLimit", key ), b2PrismaticJoint_GetLowerLimit( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/UpperLimit", key ), b2PrismaticJoint_GetUpperLimit( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MotorSpeed", key ), b2PrismaticJoint_GetMotorSpeed( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxMotorForce", key ), b2PrismaticJoint_GetMaxMotorForce( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSpringEnabled", key ), b2PrismaticJoint_IsSpringEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsLimitEnabled", key ), b2PrismaticJoint_IsLimitEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsMotorEnabled", key ), b2PrismaticJoint_IsMotorEnabled( id ) );
				break;
			}
			case b2_revoluteJoint:
			{
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringHertz", key ), b2RevoluteJoint_GetSpringHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringDampingRatio", key ),
									 b2RevoluteJoint_GetSpringDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/TargetAngle", key ), b2RevoluteJoint_GetTargetAngle( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LowerLimit", key ), b2RevoluteJoint_GetLowerLimit( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/UpperLimit", key ), b2RevoluteJoint_GetUpperLimit( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MotorSpeed", key ), b2RevoluteJoint_GetMotorSpeed( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxMotorTorque", key ), b2RevoluteJoint_GetMaxMotorTorque( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSpringEnabled", key ), b2RevoluteJoint_IsSpringEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsLimitEnabled", key ), b2RevoluteJoint_IsLimitEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsMotorEnabled", key ), b2RevoluteJoint_IsMotorEnabled( id ) );
				break;
			}
			case b2_weldJoint:
			{
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LinearHertz", key ), b2WeldJoint_GetLinearHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LinearDampingRatio", key ),
									 b2WeldJoint_GetLinearDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularHertz", key ), b2WeldJoint_GetAngularHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/AngularDampingRatio", key ),
									 b2WeldJoint_GetAngularDampingRatio( id ) );
				break;
			}
			case b2_wheelJoint:
			{
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringHertz", key ), b2WheelJoint_GetSpringHertz( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/SpringDampingRatio", key ),
									 b2WheelJoint_GetSpringDampingRatio( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/LowerLimit", key ), b2WheelJoint_GetLowerLimit( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/UpperLimit", key ), b2WheelJoint_GetUpperLimit( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MotorSpeed", key ), b2WheelJoint_GetMotorSpeed( id ) );
				REPLAY_PROPERTY_F32( &fields, ORACLE_KEY( "%s/MaxMotorTorque", key ), b2WheelJoint_GetMaxMotorTorque( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsSpringEnabled", key ), b2WheelJoint_IsSpringEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsLimitEnabled", key ), b2WheelJoint_IsLimitEnabled( id ) );
				REPLAY_PROPERTY_BOOL( &fields, ORACLE_KEY( "%s/IsMotorEnabled", key ), b2WheelJoint_IsMotorEnabled( id ) );
				break;
			}
			case b2_filterJoint:
				break;
			default:
				abort();
		}
	}
	replay_property_end( &fields );
}
