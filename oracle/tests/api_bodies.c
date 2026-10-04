// SPDX-License-Identifier: MIT

#include "oracle.h"

#include "box2d/box2d.h"

#include <assert.h>
#include <stdio.h>
#include <stdlib.h>

// Stable application tags are observations, not cross-process pointer values.
static int tags[256];
static void* Tag( int n )
{
	tags[n] = n;
	return &tags[n];
}
static int ShapeTag( b2ShapeId s )
{
	return b2Shape_IsValid( s ) ? *(int*)b2Shape_GetUserData( s ) : -1;
}
static bool SameBody( b2BodyId a, b2BodyId b )
{
	return b2StoreBodyId( a ) == b2StoreBodyId( b );
}
static bool SameShape( b2ShapeId a, b2ShapeId b )
{
	return b2StoreShapeId( a ) == b2StoreShapeId( b );
}
static bool SameChain( b2ChainId a, b2ChainId b )
{
	return b2StoreChainId( a ) == b2StoreChainId( b );
}
static bool SameWorld( b2WorldId a, b2WorldId b )
{
	return b2StoreWorldId( a ) == b2StoreWorldId( b );
}
static bool SameContact( b2ContactId a, b2ContactId b )
{
	return a.index1 == b.index1 && a.world0 == b.world0 && a.generation == b.generation;
}
static void Mass( const char* k, b2MassData m )
{
	ORACLE_F32( ORACLE_KEY( "%s/mass", k ), m.mass );
	ORACLE_VEC2( ORACLE_KEY( "%s/center", k ), m.center );
	ORACLE_F32( ORACLE_KEY( "%s/inertia", k ), m.rotationalInertia );
}
static void Bounds( const char* k, b2AABB a )
{
	ORACLE_VEC2( ORACLE_KEY( "%s/lower", k ), a.lowerBound );
	ORACLE_VEC2( ORACLE_KEY( "%s/upper", k ), a.upperBound );
}
static void Material( const char* k, b2SurfaceMaterial m )
{
	ORACLE_F32( ORACLE_KEY( "%s/friction", k ), m.friction );
	ORACLE_F32( ORACLE_KEY( "%s/restitution", k ), m.restitution );
	ORACLE_F32( ORACLE_KEY( "%s/rolling", k ), m.rollingResistance );
	ORACLE_F32( ORACLE_KEY( "%s/tangent", k ), m.tangentSpeed );
	ORACLE_U64( ORACLE_KEY( "%s/tag", k ), m.userMaterialId );
	ORACLE_U32( ORACLE_KEY( "%s/color", k ), m.customColor );
}
static void Body( const char* key, b2BodyId b )
{
	char k[256];
	ORACLE_POS( ORACLE_KEY( "%s/position", key ), b2Body_GetPosition( b ) );
	ORACLE_ROT( ORACLE_KEY( "%s/rotation", key ), b2Body_GetRotation( b ) );
	ORACLE_VEC2( ORACLE_KEY( "%s/velocity", key ), b2Body_GetLinearVelocity( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/omega", key ), b2Body_GetAngularVelocity( b ) );
	ORACLE_I32( ORACLE_KEY( "%s/type", key ), b2Body_GetType( b ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/awake", key ), b2Body_IsAwake( b ) );
	snprintf( k, sizeof( k ), "%s/massData", key );
	Mass( k, b2Body_GetMassData( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/mass", key ), b2Body_GetMass( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/inertia", key ), b2Body_GetRotationalInertia( b ) );
	ORACLE_VEC2( ORACLE_KEY( "%s/localCenter", key ), b2Body_GetLocalCenter( b ) );
	ORACLE_POS( ORACLE_KEY( "%s/worldCenter", key ), b2Body_GetWorldCenter( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/linearDamping", key ), b2Body_GetLinearDamping( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/angularDamping", key ), b2Body_GetAngularDamping( b ) );
	ORACLE_F32( ORACLE_KEY( "%s/gravity", key ), b2Body_GetGravityScale( b ) );
	b2MotionLocks locks = b2Body_GetMotionLocks( b );
	ORACLE_BOOL( ORACLE_KEY( "%s/lockX", key ), locks.linearX );
	ORACLE_BOOL( ORACLE_KEY( "%s/lockY", key ), locks.linearY );
	ORACLE_BOOL( ORACLE_KEY( "%s/lockZ", key ), locks.angularZ );
	int capacity = b2Body_GetShapeCount( b );
	ORACLE_I32( ORACLE_KEY( "%s/shapeCount", key ), capacity );
	b2ShapeId shapes[16];
	if ( capacity > 16 )
	{
		abort();
	}
	int count = b2Body_GetShapes( b, shapes, capacity );
	for ( int i = 0; i < count; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/shape/%d/tag", key, i ), ShapeTag( shapes[i] ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/shape/%d/owner", key, i ), SameBody( b2Shape_GetBody( shapes[i] ), b ) );
	}
	ORACLE_I32( ORACLE_KEY( "%s/enumerated", key ), count );
	snprintf( k, sizeof( k ), "%s/bounds", key );
	Bounds( k, b2Body_ComputeAABB( b ) );
}
static void Shape( const char* key, b2ShapeId s, b2BodyId b )
{
	char k[256];
	b2ShapeType type = b2Shape_GetType( s );
	ORACLE_I32( ORACLE_KEY( "%s/type", key ), type );
	ORACLE_I32( ORACLE_KEY( "%s/user", key ), ShapeTag( s ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/body", key ), SameBody( b2Shape_GetBody( s ), b ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/world", key ), SameWorld( b2Shape_GetWorld( s ), b2Body_GetWorld( b ) ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/orphan", key ), B2_IS_NULL( b2Shape_GetParentChain( s ) ) );
	switch ( type )
	{
		case b2_circleShape:
		{
			b2Circle c = b2Shape_GetCircle( s );
			ORACLE_VEC2( ORACLE_KEY( "%s/circle/center", key ), c.center );
			ORACLE_F32( ORACLE_KEY( "%s/circle/radius", key ), c.radius );
			break;
		}
		case b2_capsuleShape:
		{
			b2Capsule c = b2Shape_GetCapsule( s );
			ORACLE_VEC2( ORACLE_KEY( "%s/capsule/a", key ), c.center1 );
			ORACLE_VEC2( ORACLE_KEY( "%s/capsule/b", key ), c.center2 );
			ORACLE_F32( ORACLE_KEY( "%s/capsule/radius", key ), c.radius );
			break;
		}
		case b2_segmentShape:
		{
			b2Segment c = b2Shape_GetSegment( s );
			ORACLE_VEC2( ORACLE_KEY( "%s/segment/a", key ), c.point1 );
			ORACLE_VEC2( ORACLE_KEY( "%s/segment/b", key ), c.point2 );
			break;
		}
		case b2_polygonShape:
		{
			b2Polygon c = b2Shape_GetPolygon( s );
			ORACLE_I32( ORACLE_KEY( "%s/polygon/count", key ), c.count );
			ORACLE_F32( ORACLE_KEY( "%s/polygon/radius", key ), c.radius );
			ORACLE_VEC2( ORACLE_KEY( "%s/polygon/center", key ), c.centroid );
			for ( int i = 0; i < c.count; ++i )
			{
				ORACLE_VEC2( ORACLE_KEY( "%s/polygon/vertex/%d", key, i ), c.vertices[i] );
				ORACLE_VEC2( ORACLE_KEY( "%s/polygon/normal/%d", key, i ), c.normals[i] );
			}
			break;
		}
		case b2_chainSegmentShape:
		{
			b2ChainSegment c = b2Shape_GetChainSegment( s );
			ORACLE_VEC2( ORACLE_KEY( "%s/chain/ghost1", key ), c.ghost1 );
			ORACLE_VEC2( ORACLE_KEY( "%s/chain/a", key ), c.segment.point1 );
			ORACLE_VEC2( ORACLE_KEY( "%s/chain/b", key ), c.segment.point2 );
			ORACLE_VEC2( ORACLE_KEY( "%s/chain/ghost2", key ), c.ghost2 );
			break;
		}
		default:
			abort();
	}
	snprintf( k, sizeof( k ), "%s/material", key );
	Material( k, b2Shape_GetSurfaceMaterial( s ) );
	ORACLE_F32( ORACLE_KEY( "%s/density", key ), b2Shape_GetDensity( s ) );
	ORACLE_F32( ORACLE_KEY( "%s/friction", key ), b2Shape_GetFriction( s ) );
	ORACLE_F32( ORACLE_KEY( "%s/restitution", key ), b2Shape_GetRestitution( s ) );
	ORACLE_U64( ORACLE_KEY( "%s/materialTag", key ), b2Shape_GetUserMaterial( s ) );
	b2Filter f = b2Shape_GetFilter( s );
	ORACLE_U64( ORACLE_KEY( "%s/category", key ), f.categoryBits );
	ORACLE_U64( ORACLE_KEY( "%s/mask", key ), f.maskBits );
	ORACLE_I32( ORACLE_KEY( "%s/group", key ), f.groupIndex );
	ORACLE_BOOL( ORACLE_KEY( "%s/sensor", key ), b2Shape_IsSensor( s ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/sensorEvents", key ), b2Shape_AreSensorEventsEnabled( s ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/contactEvents", key ), b2Shape_AreContactEventsEnabled( s ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/hitEvents", key ), b2Shape_AreHitEventsEnabled( s ) );
	ORACLE_BOOL( ORACLE_KEY( "%s/preSolveEvents", key ), b2Shape_ArePreSolveEventsEnabled( s ) );
	snprintf( k, sizeof( k ), "%s/massData", key );
	Mass( k, b2Shape_ComputeMassData( s ) );
	snprintf( k, sizeof( k ), "%s/bounds", key );
	Bounds( k, b2Shape_GetAABB( s ) );
}
static void Cast( const char* key, b2WorldCastOutput out )
{
	ORACLE_BOOL( ORACLE_KEY( "%s/hit", key ), out.hit );
	if ( out.hit )
	{
		ORACLE_POS( ORACLE_KEY( "%s/point", key ), out.point );
		ORACLE_VEC2( ORACLE_KEY( "%s/normal", key ), out.normal );
		ORACLE_F32( ORACLE_KEY( "%s/fraction", key ), out.fraction );
		ORACLE_I32( ORACLE_KEY( "%s/iterations", key ), out.iterations );
	}
}
static void Contacts( const char* key, b2WorldId w, b2BodyId b )
{
	int capacity = b2Body_GetContactCapacity( b );
	b2ContactData contacts[16];
	if ( capacity > 16 )
	{
		abort();
	}
	int count = b2Body_GetContactData( b, contacts, capacity );
	for ( int n = 0; n < count; ++n )
	{
		b2ContactData c = contacts[n];
		char p[256];
		snprintf( p, sizeof( p ), "%s/contact/%d", key, n );
		ORACLE_I32( ORACLE_KEY( "%s/a", p ), ShapeTag( c.shapeIdA ) );
		ORACLE_I32( ORACLE_KEY( "%s/b", p ), ShapeTag( c.shapeIdB ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/valid", p ), b2Contact_IsValid( c.contactId ) );
		b2ContactData data = b2Contact_GetData( c.contactId );
		ORACLE_BOOL( ORACLE_KEY( "%s/identity", p ), SameContact( data.contactId, c.contactId ) &&
														 SameShape( data.shapeIdA, c.shapeIdA ) &&
														 SameShape( data.shapeIdB, c.shapeIdB ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/body", p ),
					 SameBody( b2Shape_GetBody( c.shapeIdA ), b ) || SameBody( b2Shape_GetBody( c.shapeIdB ), b ) );
		b2Manifold m = c.manifold;
		ORACLE_VEC2( ORACLE_KEY( "%s/normal", p ), m.normal );
		ORACLE_F32( ORACLE_KEY( "%s/rolling", p ), m.rollingImpulse );
		ORACLE_I32( ORACLE_KEY( "%s/points", p ), m.pointCount );
		for ( int i = 0; i < m.pointCount; ++i )
		{
			char q[300];
			snprintf( q, sizeof( q ), "%s/point/%d", p, i );
			b2ManifoldPoint v = m.points[i];
			ORACLE_VEC2( ORACLE_KEY( "%s/anchorA", q ), v.anchorA );
			ORACLE_VEC2( ORACLE_KEY( "%s/anchorB", q ), v.anchorB );
			ORACLE_F32( ORACLE_KEY( "%s/separation", q ), v.separation );
			ORACLE_F32( ORACLE_KEY( "%s/normalImpulse", q ), v.normalImpulse );
			ORACLE_F32( ORACLE_KEY( "%s/tangentImpulse", q ), v.tangentImpulse );
			ORACLE_F32( ORACLE_KEY( "%s/totalImpulse", q ), v.totalNormalImpulse );
			ORACLE_U32( ORACLE_KEY( "%s/id", q ), v.id );
			ORACLE_BOOL( ORACLE_KEY( "%s/persisted", q ), v.persisted );
		}
	}
	ORACLE_I32( ORACLE_KEY( "%s/contacts", key ), count );
	b2ContactEvents e = b2World_GetContactEvents( w );
	ORACLE_I32( ORACLE_KEY( "%s/begin", key ), e.beginCount );
	ORACLE_I32( ORACLE_KEY( "%s/end", key ), e.endCount );
	ORACLE_I32( ORACLE_KEY( "%s/hits", key ), e.hitCount );
	for ( int i = 0; i < e.beginCount; ++i )
	{
		b2ContactBeginTouchEvent c = e.beginEvents[i];
		ORACLE_I32( ORACLE_KEY( "%s/begin/%d/a", key, i ), ShapeTag( c.shapeIdA ) );
		ORACLE_I32( ORACLE_KEY( "%s/begin/%d/b", key, i ), ShapeTag( c.shapeIdB ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/begin/%d/valid", key, i ), b2Contact_IsValid( c.contactId ) );
	}
	for ( int i = 0; i < e.endCount; ++i )
	{
		b2ContactEndTouchEvent c = e.endEvents[i];
		ORACLE_I32( ORACLE_KEY( "%s/end/%d/a", key, i ), ShapeTag( c.shapeIdA ) );
		ORACLE_I32( ORACLE_KEY( "%s/end/%d/b", key, i ), ShapeTag( c.shapeIdB ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/end/%d/valid", key, i ), b2Contact_IsValid( c.contactId ) );
	}
	for ( int i = 0; i < e.hitCount; ++i )
	{
		b2ContactHitEvent c = e.hitEvents[i];
		ORACLE_I32( ORACLE_KEY( "%s/hit/%d/a", key, i ), ShapeTag( c.shapeIdA ) );
		ORACLE_I32( ORACLE_KEY( "%s/hit/%d/b", key, i ), ShapeTag( c.shapeIdB ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/hit/%d/valid", key, i ), b2Contact_IsValid( c.contactId ) );
		ORACLE_POS( ORACLE_KEY( "%s/hit/%d/point", key, i ), c.point );
		ORACLE_VEC2( ORACLE_KEY( "%s/hit/%d/normal", key, i ), c.normal );
		ORACLE_F32( ORACLE_KEY( "%s/hit/%d/speed", key, i ), c.approachSpeed );
	}
}
typedef struct QueryContext
{
	const char* key;
	int count;
} QueryContext;
static bool QueryCallback( b2ShapeId s, void* context )
{
	QueryContext* c = context;
	ORACLE_I32( ORACLE_KEY( "%s/tag/%d", c->key, c->count ), ShapeTag( s ) );
	c->count += 1;
	return true;
}
static void Query( const char* key, b2WorldId w, b2Pos base, float y, uint64_t mask )
{
	b2QueryFilter filter = { .categoryBits = 1, .maskBits = mask };
	QueryContext context = { key, 0 };
	b2World_OverlapAABB( w, base, (b2AABB){ { -.1f, y - .1f }, { .1f, y + .1f } }, filter, QueryCallback, &context );
	ORACLE_I32( ORACLE_KEY( "%s/count", key ), context.count );
}
static b2WorldId World( b2Vec2 gravity )
{
	b2WorldDef wd = b2DefaultWorldDef();
	wd.gravity = gravity;
	wd.enableContinuous = false;
	return b2CreateWorld( &wd );
}
static b2Pos Base( bool distant )
{
	if ( distant )
	{
		return (b2Pos){ 0x1p30, -0x1p29 };
	}
	return (b2Pos){ 7, -3 };
}
static b2ShapeId CreateGeometry( b2BodyId b, b2ShapeDef* sd, int family, float y )
{
	switch ( family )
	{
		case 0:
		{
			b2Circle g = { { 0, y }, 1 };
			return b2CreateCircleShape( b, sd, &g );
		}
		case 1:
		{
			b2Capsule g = { { -1, y }, { 1, y }, .5f };
			return b2CreateCapsuleShape( b, sd, &g );
		}
		case 2:
		{
			b2Segment g = { { -2, y }, { 2, y } };
			return b2CreateSegmentShape( b, sd, &g );
		}
		case 3:
		{
			b2Polygon g = b2MakeOffsetBox( 2, .5f, (b2Vec2){ 0, y }, (b2Rot){ 1, 0 } );
			return b2CreatePolygonShape( b, sd, &g );
		}
		default:
		{
			b2ChainSegment g = { .ghost1 = { -4, y }, .segment = { { -2, y }, { 2, y } }, .ghost2 = { 4, y } };
			return b2CreateChainSegmentShape( b, sd, &g );
		}
	}
}
static void SetGeometry( b2ShapeId s, int family, float y )
{
	switch ( family )
	{
		case 0:
		{
			b2Circle g = { { 0, y }, 1.25f };
			b2Shape_SetCircle( s, &g );
			break;
		}
		case 1:
		{
			b2Capsule g = { { -1.5f, y }, { 1.5f, y }, .75f };
			b2Shape_SetCapsule( s, &g );
			break;
		}
		case 2:
		{
			b2Segment g = { { -3, y }, { 3, y } };
			b2Shape_SetSegment( s, &g );
			break;
		}
		case 3:
		{
			b2Polygon g = b2MakeOffsetRoundedBox( 2.5f, .75f, (b2Vec2){ 0, y }, (b2Rot){ 1, 0 }, .1f );
			b2Shape_SetPolygon( s, &g );
			break;
		}
		case 4:
		{
			b2ChainSegment g = {
				.ghost1 = { -5, y - .5f }, .segment = { { -3, y }, { 3, y } }, .ghost2 = { 5, y + .5f }, .chainId = 42 };
			b2Shape_SetChainSegment( s, &g );
			break;
		}
	}
}
static void Construction( int type, int family )
{
	char name[128], k[128];
	snprintf( name, sizeof( name ), "api/bodies/construction/body%d_shape%d", type, family );
	ORACLE_OPEN( name );
	b2WorldId w = World( (b2Vec2){ 0, -2 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = (b2BodyType)type;
	bd.position = Base( false );
	bd.rotation = (b2Rot){ .6f, .8f };
	bd.linearVelocity = (b2Vec2){ 2, 1 };
	bd.angularVelocity = .5f;
	bd.linearDamping = .2f;
	bd.angularDamping = .3f;
	bd.gravityScale = .7f;
	bd.userData = Tag( 11 );
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	sd.density = 2;
	sd.material.friction = .4f;
	sd.material.userMaterialId = 21;
	b2Circle circle = { { 1, .5f }, .75f };
	b2Capsule capsule = { { -1, .5f }, { 1, .5f }, .4f };
	b2Segment segment = { { -2, 0 }, { 2, 0 } };
	b2Polygon polygon = b2MakeOffsetBox( 1, .5f, (b2Vec2){ .5f, .25f }, (b2Rot){ 1, 0 } );
	b2ChainSegment chain = { .ghost1 = { -4, 0 }, .segment = segment, .ghost2 = { 4, 0 } };
	b2ShapeId s;
	switch ( family )
	{
		case 0:
			s = b2CreateCircleShape( b, &sd, &circle );
			break;
		case 1:
			s = b2CreateCapsuleShape( b, &sd, &capsule );
			break;
		case 2:
			s = b2CreateSegmentShape( b, &sd, &segment );
			break;
		case 3:
			s = b2CreatePolygonShape( b, &sd, &polygon );
			break;
		default:
			s = b2CreateChainSegmentShape( b, &sd, &chain );
			break;
	}
	circle.center.x = 100;
	circle.radius = 8;
	capsule.center1.x = 100;
	capsule.center2.y = 100;
	capsule.radius = 8;
	segment.point1.x = 100;
	polygon.vertices[0].x = 100;
	polygon.normals[0].y = 100;
	polygon.centroid.x = 100;
	polygon.radius = 8;
	chain.ghost1.x = 100;
	chain.segment.point2.y = 100;
	sd.density = 9;
	sd.material.friction = .9f;
	sd.material.userMaterialId = 99;
	sd.filter.maskBits = 0;
	sd.userData = Tag( 99 );
	bd.position = (b2Pos){ 100, 100 };
	bd.type = b2_staticBody;
	bd.linearVelocity = (b2Vec2){ 0, 0 };
	bd.rotation = (b2Rot){ 1, 0 };
	bd.userData = Tag( 99 );
	b2BodyId other = b2CreateBody( w, &bd );
	ORACLE_I32( "bodyUser", *(int*)b2Body_GetUserData( b ) );
	ORACLE_I32( "reusedUser", *(int*)b2Body_GetUserData( other ) );
	Shape( "copied", s, b );
	Body( "constructed", b );
	b2ShapeId reused = CreateGeometry( other, &sd, family, 0 );
	Shape( "reusedShape", reused, other );
	b2ShapeDef ballastDef = b2DefaultShapeDef();
	ballastDef.userData = Tag( 8 );
	b2Circle ballast = { { -10, 0 }, .5f };
	b2CreateCircleShape( b, &ballastDef, &ballast );
	Cast( "ray", b2Shape_RayCast( s, (b2Pos){ 3, -3 }, (b2Vec2){ 8, 0 } ) );
	ORACLE_POS( "closest", b2Shape_GetClosestPoint( s, (b2Pos){ 10, 1 } ) );
	for ( int step = 0; step < 3; ++step )
	{
		b2World_Step( w, 1.0f / 60, 4 );
		snprintf( k, sizeof( k ), "step/%d", step );
		Body( k, b );
	}
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void Motion( int kind, int distant, int sleeping, int action, int wake )
{
	char name[128], k[128];
	snprintf( name, sizeof( name ), "api/bodies/motion/type%d_base%d_sleep%d_action%d_wake%d", kind, distant, sleeping, action,
			  wake );
	ORACLE_OPEN( name );
	b2Pos base = Base( distant == 1 );
	b2WorldId w = World( (b2Vec2){ 1, -4 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = (b2BodyType)kind;
	bd.position = base;
	bd.rotation = (b2Rot){ .6f, .8f };
	bd.userData = Tag( 11 );
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	sd.density = 2;
	b2Circle g = { { 1, .5f }, .75f };
	b2CreateCircleShape( b, &sd, &g );
	b2Body_SetLinearVelocity( b, (b2Vec2){ 2, -1 } );
	b2Body_SetAngularVelocity( b, .75f );
	ORACLE_VEC2( "localPoint", b2Body_GetLocalPoint( b, (b2Pos){ base.x + 3, base.y - 2 } ) );
	ORACLE_POS( "worldPoint", b2Body_GetWorldPoint( b, (b2Vec2){ 1.25f, -.75f } ) );
	ORACLE_VEC2( "localVector", b2Body_GetLocalVector( b, (b2Vec2){ 3, -2 } ) );
	ORACLE_VEC2( "worldVector", b2Body_GetWorldVector( b, (b2Vec2){ 1.25f, -.75f } ) );
	ORACLE_VEC2( "localVelocity", b2Body_GetLocalPointVelocity( b, (b2Vec2){ 1.25f, -.75f } ) );
	ORACLE_VEC2( "worldVelocity", b2Body_GetWorldPointVelocity( b, (b2Pos){ base.x + 3, base.y - 2 } ) );
	b2Body_SetLinearDamping( b, .4f );
	b2Body_SetAngularDamping( b, .7f );
	b2Body_SetGravityScale( b, .3f );
	if ( sleeping == 1 )
	{
		b2Body_SetAwake( b, false );
	}
	b2Pos point = { base.x + 2, base.y + 1 };
	bool flag = wake == 1;
	switch ( action )
	{
		case 0:
			b2Body_ApplyForce( b, (b2Vec2){ 5, 9 }, point, flag );
			break;
		case 1:
			b2Body_ApplyForceToCenter( b, (b2Vec2){ 5, 9 }, flag );
			break;
		case 2:
			b2Body_ApplyTorque( b, 7, flag );
			break;
		case 3:
			b2Body_ApplyLinearImpulse( b, (b2Vec2){ 3, 4 }, point, flag );
			break;
		case 4:
			b2Body_ApplyLinearImpulseToCenter( b, (b2Vec2){ 3, 4 }, flag );
			break;
		case 5:
			b2Body_ApplyAngularImpulse( b, 2, flag );
			break;
		case 6:
			b2Body_SetTargetTransform( b, (b2WorldTransform){ point, { 0, 1 } }, 1.0f / 60, flag );
			break;
		case 7:
			b2Body_SetTargetTransform( b, (b2WorldTransform){ { base.x + .0001f, base.y }, { .6f, .8f } }, 1.0f / 60, flag );
			break;
		case 8:
			b2Body_ApplyForce( b, (b2Vec2){ 5, 9 }, point, flag );
			b2Body_ApplyTorque( b, 7, flag );
			b2Body_ClearForces( b );
			break;
	}
	Body( "immediate", b );
	for ( int step = 0; step < 3; ++step )
	{
		b2World_Step( w, 1.0f / 60, 4 );
		snprintf( k, sizeof( k ), "step/%d", step );
		Body( k, b );
	}
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void Locks( int locks, int runtime )
{
	char name[128], k[128];
	snprintf( name, sizeof( name ), "api/bodies/motion/locks%d_runtime%d", locks, runtime );
	ORACLE_OPEN( name );
	b2WorldId w = World( (b2Vec2){ 1, -4 } );
	b2MotionLocks m = { .linearX = ( locks & 1 ) != 0, .linearY = ( locks & 2 ) != 0, .angularZ = ( locks & 4 ) != 0 };
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = Base( false );
	bd.linearVelocity = (b2Vec2){ 2, -1 };
	bd.angularVelocity = .75f;
	if ( runtime == 0 )
	{
		bd.motionLocks = m;
	}
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	CreateGeometry( b, &sd, 3, 0 );
	if ( runtime == 1 )
	{
		b2Body_SetMotionLocks( b, m );
	}
	b2Body_SetLinearDamping( b, .4f );
	b2Body_SetAngularDamping( b, .7f );
	b2Body_SetGravityScale( b, .3f );
	b2Body_ApplyForce( b, (b2Vec2){ 5, 9 }, (b2Pos){ 9, -2 }, true );
	b2Body_ApplyTorque( b, 7, true );
	b2Body_ApplyLinearImpulse( b, (b2Vec2){ 3, 4 }, (b2Pos){ 9, -2 }, true );
	b2Body_ApplyAngularImpulse( b, 2, true );
	Body( "loaded", b );
	for ( int step = 0; step < 3; ++step )
	{
		b2World_Step( w, 1.0f / 60, 4 );
		snprintf( k, sizeof( k ), "step/%d", step );
		Body( k, b );
	}
	b2Body_SetMotionLocks( b, (b2MotionLocks){ 0 } );
	b2Body_ApplyTorque( b, 7, true );
	b2Body_ApplyForceToCenter( b, (b2Vec2){ 5, 9 }, true );
	b2World_Step( w, 1.0f / 60, 4 );
	Body( "unlocked", b );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void Target( int kind )
{
	char name[128];
	snprintf( name, sizeof( name ), "api/bodies/motion/target_type%d", kind );
	ORACLE_OPEN( name );
	b2WorldId w = World( (b2Vec2){ 0, 0 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = (b2BodyType)kind;
	bd.position = Base( false );
	bd.rotation = (b2Rot){ 0, 1 };
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	CreateGeometry( b, &sd, 0, 0 );
	b2Body_SetTransform( b, (b2Pos){ 8, -2 }, (b2Rot){ .6f, .8f } );
	Body( "teleport", b );
	b2Body_SetTargetTransform( b, (b2WorldTransform){ { 9, -1 }, { 0, 1 } }, .25f, true );
	Body( "target", b );
	b2World_Step( w, .25f, 4 );
	Body( "arrived", b );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void Load( const char* key, b2WorldId w, b2BodyId b )
{
	char k[256];
	b2Body_SetLinearVelocity( b, (b2Vec2){ 2, -1 } );
	b2Body_SetAngularVelocity( b, .75f );
	b2Body_ApplyLinearImpulse( b, (b2Vec2){ 3, 4 }, (b2Pos){ 9, -2 }, true );
	b2Body_ApplyTorque( b, 7, true );
	b2Body_ApplyForceToCenter( b, (b2Vec2){ 5, 9 }, true );
	snprintf( k, sizeof( k ), "%s/loaded", key );
	Body( k, b );
	b2World_Step( w, 1.0f / 60, 4 );
	snprintf( k, sizeof( k ), "%s/step", key );
	Body( k, b );
}
static void MassMutation( void )
{
	ORACLE_OPEN( "api/bodies/mass" );
	b2WorldId w = World( (b2Vec2){ 0, -2 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = Base( false );
	bd.rotation = (b2Rot){ .6f, .8f };
	bd.angularVelocity = .75f;
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	sd.density = 2;
	b2Circle circle = { { 1, .5f }, .75f };
	b2ShapeId a = b2CreateCircleShape( b, &sd, &circle );
	Body( "circle", b );
	Load( "circle", w, b );
	sd.userData = Tag( 8 );
	sd.density = 0;
	b2Capsule capsule = { { -2, 0 }, { -1, 0 }, .4f };
	b2ShapeId z = b2CreateCapsuleShape( b, &sd, &capsule );
	Body( "zeroAdded", b );
	Load( "zeroAdded", w, b );
	sd.userData = Tag( 9 );
	sd.density = 3;
	sd.isSensor = true;
	sd.updateBodyMass = false;
	b2Polygon polygon = b2MakeOffsetBox( .5f, 1, (b2Vec2){ -1, 2 }, (b2Rot){ 1, 0 } );
	b2ShapeId p = b2CreatePolygonShape( b, &sd, &polygon );
	Body( "deferredAdded", b );
	b2Body_UpdateMassFromShapes( b );
	Body( "updatedAdded", b );
	Load( "updatedAdded", w, b );
	b2Shape_SetDensity( a, 4, false );
	Body( "densityDeferred", b );
	Load( "densityDeferred", w, b );
	b2Body_UpdateMassFromShapes( b );
	Body( "densityUpdated", b );
	Load( "densityUpdated", w, b );
	b2Shape_SetDensity( z, 2, true );
	Body( "densityAutomatic", b );
	Load( "densityAutomatic", w, b );
	b2Shape_SetDensity( z, 0, true );
	Body( "densityBackToZero", b );
	Load( "densityBackToZero", w, b );
	b2Body_SetMassData( b, (b2MassData){ 7, { -.25f, 1.5f }, 11 } );
	Body( "override", b );
	Load( "override", w, b );
	b2Body_UpdateMassFromShapes( b );
	Body( "restoreOverride", b );
	Load( "restoreOverride", w, b );
	b2Body_SetMassData( b, (b2MassData){ 7, { -.25f, 1.5f }, 11 } );
	Body( "overrideBeforeRemove", b );
	b2DestroyShape( z, false );
	ORACLE_BOOL( "removedZeroValid", b2Shape_IsValid( z ) );
	Body( "removeDeferred", b );
	Load( "removeDeferred", w, b );
	b2Body_UpdateMassFromShapes( b );
	Body( "removeUpdated", b );
	Load( "removeUpdated", w, b );
	b2Body_SetMassData( b, (b2MassData){ 9, { .5f, -.5f }, 13 } );
	sd.isSensor = false;
	sd.updateBodyMass = true;
	sd.userData = Tag( 10 );
	sd.density = 1;
	b2ShapeId extra = b2CreateCircleShape( b, &sd, &circle );
	Body( "addResetsOverride", b );
	Load( "addResetsOverride", w, b );
	b2DestroyShape( extra, true );
	Body( "removeAutomatic", b );
	Load( "removeAutomatic", w, b );
	for ( int kind = 0; kind < 3; ++kind )
	{
		char key[128];
		b2Body_SetMassData( b, (b2MassData){ 9, { .5f, -.5f }, 13 } );
		b2Body_SetType( b, (b2BodyType)kind );
		snprintf( key, sizeof( key ), "type/%d", kind );
		Body( key, b );
		Load( key, w, b );
	}
	b2Body_SetMotionLocks( b, (b2MotionLocks){ .angularZ = true } );
	b2Body_SetMassData( b, (b2MassData){ 7, { -.25f, 1.5f }, 11 } );
	Body( "lockedOverride", b );
	Load( "lockedOverride", w, b );
	b2Body_SetMotionLocks( b, (b2MotionLocks){ 0 } );
	Body( "unlockRecomputes", b );
	Load( "unlockRecomputes", w, b );
	b2DestroyShape( p, true );
	b2DestroyShape( a, true );
	Body( "empty", b );
	Load( "empty", w, b );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void MutationObservation( const char* key, b2WorldId w, b2Pos base, b2ShapeId s, b2BodyId ground, b2BodyId visitor )
{
	char k[256];
	snprintf( k, sizeof( k ), "%s/shape", key );
	Shape( k, s, ground );
	snprintf( k, sizeof( k ), "%s/oldQuery", key );
	Query( k, w, base, 0, 4 );
	snprintf( k, sizeof( k ), "%s/newQuery", key );
	Query( k, w, base, 3, 4 );
	snprintf( k, sizeof( k ), "%s/body", key );
	Body( k, visitor );
	snprintf( k, sizeof( k ), "%s/effects", key );
	Contacts( k, w, visitor );
}
static void Impact( const char* key, float y, b2WorldId w, b2Pos base, b2ShapeId s, b2BodyId ground, b2BodyId visitor )
{
	b2Body_SetTransform( visitor, (b2Pos){ base.x, base.y + y - .3f }, (b2Rot){ 1, 0 } );
	b2Body_SetLinearVelocity( visitor, (b2Vec2){ 2, 3 } );
	b2Body_SetAngularVelocity( visitor, .5f );
	for ( int step = 0; step < 4; ++step )
	{
		char k[128];
		b2World_Step( w, 1.0f / 60, 4 );
		snprintf( k, sizeof( k ), "%s/%d", key, step );
		MutationObservation( k, w, base, s, ground, visitor );
	}
}
static float TaggedFriction( float frictionA, uint64_t a, float frictionB, uint64_t b )
{
	(void)frictionA;
	(void)frictionB;
	return a == 91 || b == 91 ? .05f : .6f;
}
static void ShapeMutation( int family )
{
	char name[128];
	snprintf( name, sizeof( name ), "api/shapes/mutation/shape%d", family );
	ORACLE_OPEN( name );
	b2Pos base = Base( false );
	b2WorldId w = World( (b2Vec2){ 0, 0 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.position = base;
	b2BodyId ground = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	sd.filter.categoryBits = 4;
	sd.enableContactEvents = true;
	sd.enableHitEvents = true;
	sd.material.friction = .3f;
	b2ShapeId s = CreateGeometry( ground, &sd, family, 0 );
	bd.type = b2_dynamicBody;
	bd.position.y -= .3f;
	bd.enableSleep = false;
	b2BodyId visitor = b2CreateBody( w, &bd );
	sd.filter.categoryBits = 2;
	sd.userData = Tag( 9 );
	b2Circle g = { { 0, 0 }, .5f };
	b2CreateCircleShape( visitor, &sd, &g );
	Impact( "initial", 0, w, base, s, ground, visitor );
	SetGeometry( s, family, 3 );
	MutationObservation( "geometryImmediate", w, base, s, ground, visitor );
	b2World_Step( w, 1.0f / 60, 4 );
	MutationObservation( "oldLocationCleared", w, base, s, ground, visitor );
	Impact( "geometryImpact", 3, w, base, s, ground, visitor );
	b2Shape_SetDensity( s, 2, true );
	Impact( "density", 3, w, base, s, ground, visitor );
	b2Shape_SetFriction( s, .8f );
	Impact( "friction", 3, w, base, s, ground, visitor );
	b2Shape_SetRestitution( s, .7f );
	Impact( "restitution", 3, w, base, s, ground, visitor );
	b2World_SetFrictionCallback( w, TaggedFriction );
	b2Shape_SetUserMaterial( s, 91 );
	Impact( "tag", 3, w, base, s, ground, visitor );
	b2World_SetFrictionCallback( w, NULL );
	b2SurfaceMaterial material = { .friction = .6f,
								   .restitution = .2f,
								   .rollingResistance = .15f,
								   .tangentSpeed = 2,
								   .userMaterialId = 92,
								   .customColor = 0x123456 };
	b2Shape_SetSurfaceMaterial( s, &material );
	material.tangentSpeed = 99;
	Impact( "surface", 3, w, base, s, ground, visitor );
	b2Filter filter = b2Shape_GetFilter( s );
	filter.maskBits = 0;
	b2Shape_SetFilter( s, filter );
	MutationObservation( "filterImmediate", w, base, s, ground, visitor );
	Impact( "filtered", 3, w, base, s, ground, visitor );
	filter.maskBits = UINT64_MAX;
	filter.categoryBits = 8;
	b2Shape_SetFilter( s, filter );
	Impact( "categoryChanged", 3, w, base, s, ground, visitor );
	Query( "category8", w, base, 3, 8 );
	filter.categoryBits = 4;
	b2Shape_SetFilter( s, filter );
	b2Shape_EnableSensorEvents( s, true );
	b2Shape_EnableContactEvents( s, false );
	b2Shape_EnableHitEvents( s, false );
	b2Shape_EnablePreSolveEvents( s, true );
	Impact( "flagsOff", 3, w, base, s, ground, visitor );
	b2Shape_EnableSensorEvents( s, false );
	b2Shape_EnableContactEvents( s, true );
	b2Shape_EnableHitEvents( s, true );
	b2Shape_EnablePreSolveEvents( s, false );
	Impact( "flagsOn", 3, w, base, s, ground, visitor );
	for ( int next = 0; next < 5; ++next )
	{
		char key[128];
		SetGeometry( s, next, 3 );
		snprintf( key, sizeof( key ), "converted/%d", next );
		Impact( key, 3, w, base, s, ground, visitor );
	}
	b2DestroyShape( s, true );
	ORACLE_BOOL( "destroyed", b2Shape_IsValid( s ) );
	b2World_Step( w, 1.0f / 60, 4 );
	Query( "destroyedQuery", w, base, 3, 4 );
	Body( "destroyedBody", visitor );
	Contacts( "destroyedEffects", w, visitor );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void ShapeQueries( int distant, int family )
{
	char name[128], key[128], k[256];
	snprintf( name, sizeof( name ), "api/shapes/queries/base%d_shape%d", distant, family );
	ORACLE_OPEN( name );
	b2Pos base = Base( distant == 1 );
	if ( distant == 2 )
	{
		base = (b2Pos){ 0, 0 };
	}
	b2WorldId w = World( (b2Vec2){ 0, 0 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = base;
	bd.rotation = (b2Rot){ 0, 1 };
	b2BodyId b = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 7 );
	b2ShapeId s = CreateGeometry( b, &sd, family, 0 );
	b2ShapeDef ballastDef = b2DefaultShapeDef();
	ballastDef.userData = Tag( 8 );
	ballastDef.filter.categoryBits = 2;
	b2Circle ballast = { { -10, 0 }, .5f };
	b2CreateCircleShape( b, &ballastDef, &ballast );
	Shape( "shape", s, b );
	if ( family == 0 || family == 1 || family == 3 )
	{
		b2Pos boundary = { base.x + .5f, base.y };
		if ( family == 0 )
		{
			boundary.x = base.x + 1;
		}
		b2Pos points[4] = { base, { base.x + .5f, base.y + .25f }, { base.x + 4, base.y + 4 }, boundary };
		for ( int i = 0; i < 4; ++i )
		{
			ORACLE_BOOL( ORACLE_KEY( "point/%d", i ), b2Shape_TestPoint( s, points[i] ) );
		}
	}
	b2Pos origins[7] = { { base.x, base.y - 4 }, { base.x, base.y + 4 }, { base.x + 4, base.y - 4 },  base,
						 { base.x - 4, base.y }, { base.x, base.y - 4 }, { base.x + .5f, base.y - 4 } };
	b2Vec2 translations[7] = { { 0, 8 }, { 0, -8 }, { 0, 8 }, { 0, 8 }, { 8, 0 }, { 0, 1 }, { 0, 8 } };
	for ( int i = 0; i < 7; ++i )
	{
		snprintf( key, sizeof( key ), "ray/%d", i );
		Cast( key, b2Shape_RayCast( s, origins[i], translations[i] ) );
	}
	b2Pos closest[3] = { { base.x + 3, base.y + 5 }, base, { base.x - 4, base.y - 2 } };
	for ( int i = 0; i < 3; ++i )
	{
		ORACLE_POS( ORACLE_KEY( "closest/%d", i ), b2Shape_GetClosestPoint( s, closest[i] ) );
	}
	for ( int mode = 0; mode < 4; ++mode )
	{
		for ( int wake = 0; wake < 2; ++wake )
		{
			snprintf( key, sizeof( key ), "force/%d/wake/%d", mode, wake );
			b2Body_SetTransform( b, base, (b2Rot){ 0, 1 } );
			b2Body_SetLinearVelocity( b, (b2Vec2){ 1, -.5f } );
			b2Body_SetAngularVelocity( b, .3f );
			b2Body_SetAwake( b, false );
			if ( mode == 1 )
			{
				b2Body_SetAwake( b, true );
				b2Body_SetLinearVelocity( b, (b2Vec2){ 1, -.5f } );
				b2Body_SetAngularVelocity( b, .3f );
			}
			if ( mode < 2 )
			{
				b2Shape_ApplyWind( s, (b2Vec2){ 5, 2 }, .4f, .2f, wake == 1 );
			}
			else
			{
				if ( wake == 1 )
				{
					b2Body_SetAwake( b, true );
				}
				b2ExplosionDef def = b2DefaultExplosionDef();
				def.maskBits = 1;
				def.position = (b2Pos){ base.x - 2, base.y - 1 };
				def.radius = 1;
				def.falloff = 5;
				def.impulsePerLength = 3;
				if ( mode == 3 )
				{
					def.impulsePerLength = -3;
				}
				b2World_Explode( w, &def );
			}
			snprintf( k, sizeof( k ), "%s/immediate", key );
			Body( k, b );
			for ( int step = 0; step < 3; ++step )
			{
				b2World_Step( w, 1.0f / 60, 4 );
				snprintf( k, sizeof( k ), "%s/step/%d", key, step );
				Body( k, b );
			}
		}
	}
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static void ChainObservation( const char* key, b2ChainId chain, b2ShapeId* segments, int capacity, b2BodyId ground, b2WorldId w )
{
	char k[256], p[256];
	ORACLE_I32( ORACLE_KEY( "%s/count", key ), b2Chain_GetSegmentCount( chain ) );
	b2ShapeId current[4];
	if ( capacity > 4 )
	{
		abort();
	}
	int count = b2Chain_GetSegments( chain, current, capacity );
	for ( int n = 0; n < count; ++n )
	{
		snprintf( p, sizeof( p ), "%s/segment/%d", key, n );
		b2ShapeId s = current[n];
		ORACLE_BOOL( ORACLE_KEY( "%s/identity", p ), SameShape( s, segments[n] ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/parent", p ), SameChain( b2Shape_GetParentChain( s ), chain ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/valid", p ), b2Shape_IsValid( s ) );
		Shape( p, s, ground );
		snprintf( k, sizeof( k ), "%s/reported", p );
		Material( k, b2Chain_GetSurfaceMaterial( chain, n ) );
	}
	ORACLE_I32( ORACLE_KEY( "%s/enumerated", key ), count );
	ORACLE_BOOL( ORACLE_KEY( "%s/world", key ), SameWorld( b2Chain_GetWorld( chain ), w ) );
	snprintf( k, sizeof( k ), "%s/ground", key );
	Body( k, ground );
	snprintf( k, sizeof( k ), "%s/query", key );
	Query( k, w, b2Body_GetPosition( ground ), 0, 4 );
}
static void ChainMaterialImpact( const char* key, b2WorldId w, b2BodyId ground, b2BodyId visitor, b2ShapeId segment )
{
	b2Segment edge = b2Shape_GetChainSegment( segment ).segment;
	b2Vec2 tangent = b2Normalize( b2Sub( edge.point2, edge.point1 ) );
	b2Vec2 normal = b2RightPerp( tangent );
	b2Vec2 start = b2Add( b2Lerp( edge.point1, edge.point2, .5f ), b2MulSV( .3f, normal ) );
	b2Body_SetTransform( visitor, b2Body_GetWorldPoint( ground, start ), (b2Rot){ 1, 0 } );
	b2Body_SetLinearVelocity( visitor, b2Add( b2MulSV( -2, normal ), b2MulSV( .5f, tangent ) ) );
	b2Body_SetAngularVelocity( visitor, .5f );
	for ( int step = 0; step < 8; ++step )
	{
		char p[128], k[256];
		b2World_Step( w, 1.0f / 60, 4 );
		snprintf( p, sizeof( p ), "%s/%d", key, step );
		snprintf( k, sizeof( k ), "%s/body", p );
		Body( k, visitor );
		snprintf( k, sizeof( k ), "%s/effects", p );
		Contacts( k, w, visitor );
	}
}
static void Chains( int topology, int reverse )
{
	char name[128], key[128], k[256];
	snprintf( name, sizeof( name ), "api/chains/topology%d_reverse%d", topology, reverse );
	ORACLE_OPEN( name );
	b2Pos base = Base( false );
	b2WorldId w = World( (b2Vec2){ 0, 0 } );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.position = base;
	b2BodyId ground = b2CreateBody( w, &bd );
	b2Vec2 points[4] = { { -2, 0 }, { 2, 0 } };
	int pointCount = 2;
	bool loop = false;
	switch ( topology )
	{
		case 1:
			points[0] = (b2Vec2){ -2, 0 };
			points[1] = (b2Vec2){ 0, 0 };
			points[2] = (b2Vec2){ 2, 0 };
			points[3] = (b2Vec2){ 3, 1 };
			pointCount = 4;
			break;
		case 2:
			points[0] = (b2Vec2){ -2, 0 };
			points[1] = (b2Vec2){ 2, 0 };
			points[2] = (b2Vec2){ 0, 3 };
			pointCount = 3;
			loop = true;
			break;
		case 3:
			points[0] = (b2Vec2){ -2, 0 };
			points[1] = (b2Vec2){ 2, 0 };
			points[2] = (b2Vec2){ 2, 3 };
			points[3] = (b2Vec2){ -2, 3 };
			pointCount = 4;
			loop = true;
			break;
		case 4:
			points[0] = (b2Vec2){ -.02f, 0 };
			points[1] = (b2Vec2){ 0, 0 };
			points[2] = (b2Vec2){ .02f, 0 };
			pointCount = 3;
			break;
	}
	if ( reverse == 1 )
	{
		for ( int i = 0; i < pointCount / 2; ++i )
		{
			int j = pointCount - 1 - i;
			b2Vec2 temp = points[i];
			points[i] = points[j];
			points[j] = temp;
		}
	}
	b2SurfaceMaterial materials[4];
	for ( int i = 0; i < pointCount; ++i )
	{
		materials[i] = b2DefaultSurfaceMaterial();
		materials[i].friction = .2f;
		materials[i].userMaterialId = (uint64_t)( 20 + i );
	}
	b2ChainDef cd = b2DefaultChainDef();
	cd.points = points;
	cd.pointCount = pointCount;
	cd.materials = materials;
	cd.materialCount = loop ? pointCount : pointCount - 1;
	cd.isLoop = loop;
	cd.ghost1 = (b2Vec2){ -4, 0 };
	cd.ghost2 = (b2Vec2){ 4, 2 };
	if ( reverse == 1 )
	{
		b2Vec2 temp = cd.ghost1;
		cd.ghost1 = cd.ghost2;
		cd.ghost2 = temp;
	}
	cd.userData = Tag( 7 );
	cd.filter.categoryBits = 4;
	b2ChainId chain = b2CreateChain( ground, &cd );
	for ( int i = 0; i < pointCount; ++i )
	{
		points[i] = (b2Vec2){ 100, 100 };
		materials[i].friction = .9f;
		materials[i].userMaterialId = 99;
	}
	cd.ghost1.x = 100;
	cd.ghost2.y = 100;
	cd.filter.maskBits = 0;
	cd.userData = Tag( 99 );
	int capacity = b2Chain_GetSegmentCount( chain );
	b2ShapeId segments[4];
	if ( capacity > 4 )
	{
		abort();
	}
	int count = b2Chain_GetSegments( chain, segments, capacity );
	for ( int i = 0; i < count; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "copiedUser/%d", i ), ShapeTag( segments[i] ) );
		b2Shape_SetUserData( segments[i], Tag( 100 + i ) );
	}
	ChainObservation( "copied", chain, segments, capacity, ground, w );
	bd.type = b2_dynamicBody;
	bd.enableSleep = false;
	bd.position = (b2Pos){ base.x, base.y - 1 };
	b2BodyId visitor = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = Tag( 9 );
	sd.filter.categoryBits = 2;
	sd.enableContactEvents = true;
	sd.enableHitEvents = true;
	b2Circle g = { { 0, 0 }, .25f };
	b2CreateCircleShape( visitor, &sd, &g );
	ChainMaterialImpact( "originalMaterialImpact", w, ground, visitor, segments[0] );
	b2SurfaceMaterial m = { .friction = .7f,
							.restitution = .4f,
							.tangentSpeed = 1.5f,
							.rollingResistance = .1f,
							.userMaterialId = 91,
							.customColor = 0xabcdef };
	b2Chain_SetSurfaceMaterial( chain, &m, 0 );
	m.friction = .9f;
	ChainObservation( "oneMaterial", chain, segments, capacity, ground, w );
	ChainMaterialImpact( "oneMaterialImpact", w, ground, visitor, segments[0] );
	m.userMaterialId = 92;
	b2Chain_SetAllSurfaceMaterials( chain, &m );
	m.tangentSpeed = 99;
	ChainObservation( "allMaterials", chain, segments, capacity, ground, w );
	ChainMaterialImpact( "allMaterialImpact", w, ground, visitor, segments[0] );
	for ( int side = 0; side < 2; ++side )
	{
		float y = -1, speed = 2;
		if ( side == 1 )
		{
			y = 1;
			speed = -2;
		}
		b2Body_SetTransform( visitor, (b2Pos){ base.x, base.y + y }, (b2Rot){ 1, 0 } );
		b2Body_SetLinearVelocity( visitor, (b2Vec2){ 0, speed } );
		b2Body_SetAngularVelocity( visitor, 0 );
		for ( int step = 0; step < 60; ++step )
		{
			b2World_Step( w, 1.0f / 60, 4 );
			snprintf( key, sizeof( key ), "side/%d/step/%d", side, step );
			snprintf( k, sizeof( k ), "%s/body", key );
			Body( k, visitor );
			snprintf( k, sizeof( k ), "%s/effects", key );
			Contacts( k, w, visitor );
		}
	}
	Query( "beforeDestroy", w, base, 0, 4 );
	b2DestroyChain( chain );
	ORACLE_BOOL( "chainValid", b2Chain_IsValid( chain ) );
	for ( int i = 0; i < count; ++i )
	{
		ORACLE_BOOL( ORACLE_KEY( "segmentValid/%d", i ), b2Shape_IsValid( segments[i] ) );
	}
	Body( "afterDestroy", ground );
	b2World_Step( w, 1.0f / 60, 4 );
	Query( "afterDestroyQuery", w, base, 0, 4 );
	Contacts( "afterDestroyEffects", w, visitor );
	cd = b2DefaultChainDef();
	b2Vec2 replacementPoints[2] = { { -2, 0 }, { 2, 0 } };
	b2SurfaceMaterial replacementMaterial = b2DefaultSurfaceMaterial();
	cd.points = replacementPoints;
	cd.pointCount = 2;
	cd.ghost1 = (b2Vec2){ -4, 0 };
	cd.ghost2 = (b2Vec2){ 4, 0 };
	cd.materials = &replacementMaterial;
	cd.materialCount = 1;
	b2ChainId replacement = b2CreateChain( ground, &cd );
	ORACLE_BOOL( "replacementDifferent", !SameChain( replacement, chain ) );
	ORACLE_BOOL( "oldChainStillInvalid", b2Chain_IsValid( chain ) );
	for ( int i = 0; i < count; ++i )
	{
		ORACLE_BOOL( ORACLE_KEY( "oldSegmentStillInvalid/%d", i ), b2Shape_IsValid( segments[i] ) );
	}
	b2DestroyBody( ground );
	ORACLE_BOOL( "replacementDestroyedWithBody", b2Chain_IsValid( replacement ) );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
}
static int AssertFailure( const char* condition, const char* file, int line )
{
	fprintf( stderr, "%s:%d: %s\n", file, line, condition );
	abort();
}
int main( void )
{
	b2SetAssertFcn( AssertFailure );
	for ( int type = 0; type < 3; ++type )
	{
		for ( int family = 0; family < 5; ++family )
		{
			Construction( type, family );
		}
	}
	int baseCount = 1;
#if defined( BOX2D_DOUBLE_PRECISION )
	baseCount = 2;
#endif
	for ( int distant = 0; distant < baseCount; ++distant )
	{
		for ( int kind = 0; kind < 3; ++kind )
		{
			for ( int sleeping = 0; sleeping < 2; ++sleeping )
			{
				for ( int action = 0; action < 9; ++action )
				{
					for ( int wake = 0; wake < 2; ++wake )
					{
						Motion( kind, distant, sleeping, action, wake );
					}
				}
			}
		}
	}
	for ( int locks = 0; locks < 8; ++locks )
	{
		for ( int runtime = 0; runtime < 2; ++runtime )
		{
			Locks( locks, runtime );
		}
	}
	for ( int kind = 0; kind < 3; ++kind )
	{
		Target( kind );
	}
	MassMutation();
	for ( int family = 0; family < 5; ++family )
	{
		ShapeMutation( family );
	}
	for ( int distant = 0; distant < baseCount; ++distant )
	{
		for ( int family = 0; family < 5; ++family )
		{
			ShapeQueries( distant, family );
		}
	}
	for ( int family = 0; family < 5; ++family )
	{
		ShapeQueries( 2, family );
	}
	for ( int topology = 0; topology < 5; ++topology )
	{
		for ( int reverse = 0; reverse < 2; ++reverse )
		{
			Chains( topology, reverse );
		}
	}
	return 0;
}
