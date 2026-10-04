// SPDX-License-Identifier: MIT
#include "contact.h"
#include "oracle.h"
#include "physics_world.h"
#include "test_macros.h"

#include "box2d/box2d.h"

#include <stdlib.h>
#include <string.h>

static uint8_t* capture( b2WorldId w, int* size )
{
	*size = b2World_GetSnapshot( w, NULL, 0 );
	uint8_t* image = malloc( *size );
	b2World_GetSnapshot( w, image, *size );
	return image;
}

static void observe( b2WorldId w, b2Pos origin, const char* key )
{
	char owned[256];
	snprintf( owned, sizeof( owned ), "%s", key );
	key = owned;
	ORACLE_U64( ORACLE_KEY( "%s/state", key ), b2World_GetStateHash( w ) );
	b2SensorEvents se = b2World_GetSensorEvents( w );
	ORACLE_I32( ORACLE_KEY( "%s/sensorBegin", key ), se.beginCount );
	for ( int i = 0; i < se.beginCount; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/sensorBegin/%d/sensor", key, i ), se.beginEvents[i].sensorShapeId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/sensorBegin/%d/sensorGeneration", key, i ), se.beginEvents[i].sensorShapeId.generation );
		ORACLE_I32( ORACLE_KEY( "%s/sensorBegin/%d/visitor", key, i ), se.beginEvents[i].visitorShapeId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/sensorBegin/%d/visitorGeneration", key, i ), se.beginEvents[i].visitorShapeId.generation );
	}
	ORACLE_I32( ORACLE_KEY( "%s/sensorEnd", key ), se.endCount );
	for ( int i = 0; i < se.endCount; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/sensorEnd/%d/sensor", key, i ), se.endEvents[i].sensorShapeId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/sensorEnd/%d/sensorGeneration", key, i ), se.endEvents[i].sensorShapeId.generation );
		ORACLE_I32( ORACLE_KEY( "%s/sensorEnd/%d/visitor", key, i ), se.endEvents[i].visitorShapeId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/sensorEnd/%d/visitorGeneration", key, i ), se.endEvents[i].visitorShapeId.generation );
	}
	b2ContactEvents ce = b2World_GetContactEvents( w );
	ORACLE_I32( ORACLE_KEY( "%s/contactBegin", key ), ce.beginCount );
	for ( int i = 0; i < ce.beginCount; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/contactBegin/%d/a", key, i ), ce.beginEvents[i].shapeIdA.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactBegin/%d/aGeneration", key, i ), ce.beginEvents[i].shapeIdA.generation );
		ORACLE_I32( ORACLE_KEY( "%s/contactBegin/%d/b", key, i ), ce.beginEvents[i].shapeIdB.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactBegin/%d/bGeneration", key, i ), ce.beginEvents[i].shapeIdB.generation );
		ORACLE_I32( ORACLE_KEY( "%s/contactBegin/%d/contact", key, i ), ce.beginEvents[i].contactId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactBegin/%d/contactGeneration", key, i ), ce.beginEvents[i].contactId.generation );
	}
	ORACLE_I32( ORACLE_KEY( "%s/contactEnd", key ), ce.endCount );
	for ( int i = 0; i < ce.endCount; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/contactEnd/%d/a", key, i ), ce.endEvents[i].shapeIdA.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactEnd/%d/aGeneration", key, i ), ce.endEvents[i].shapeIdA.generation );
		ORACLE_I32( ORACLE_KEY( "%s/contactEnd/%d/b", key, i ), ce.endEvents[i].shapeIdB.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactEnd/%d/bGeneration", key, i ), ce.endEvents[i].shapeIdB.generation );
		ORACLE_I32( ORACLE_KEY( "%s/contactEnd/%d/contact", key, i ), ce.endEvents[i].contactId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactEnd/%d/contactGeneration", key, i ), ce.endEvents[i].contactId.generation );
	}
	ORACLE_I32( ORACLE_KEY( "%s/contactHit", key ), ce.hitCount );
	for ( int i = 0; i < ce.hitCount; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/contactHit/%d/a", key, i ), ce.hitEvents[i].shapeIdA.index1 );
		ORACLE_I32( ORACLE_KEY( "%s/contactHit/%d/b", key, i ), ce.hitEvents[i].shapeIdB.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactHit/%d/aGeneration", key, i ), ce.hitEvents[i].shapeIdA.generation );
		ORACLE_U32( ORACLE_KEY( "%s/contactHit/%d/bGeneration", key, i ), ce.hitEvents[i].shapeIdB.generation );
		ORACLE_I32( ORACLE_KEY( "%s/contactHit/%d/contact", key, i ), ce.hitEvents[i].contactId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/contactHit/%d/contactGeneration", key, i ), ce.hitEvents[i].contactId.generation );
		ORACLE_POS( ORACLE_KEY( "%s/contactHit/%d/point", key, i ), ce.hitEvents[i].point );
		ORACLE_VEC2( ORACLE_KEY( "%s/contactHit/%d/normal", key, i ), ce.hitEvents[i].normal );
		ORACLE_F32( ORACLE_KEY( "%s/contactHit/%d/speed", key, i ), ce.hitEvents[i].approachSpeed );
	}
	b2JointEvents je = b2World_GetJointEvents( w );
	ORACLE_I32( ORACLE_KEY( "%s/jointEvents", key ), je.count );
	for ( int i = 0; i < je.count; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "%s/jointEvent/%d/id", key, i ), je.jointEvents[i].jointId.index1 );
		ORACLE_U32( ORACLE_KEY( "%s/jointEvent/%d/generation", key, i ), je.jointEvents[i].jointId.generation );
		ORACLE_BOOL( ORACLE_KEY( "%s/jointEvent/%d/userDataNil", key, i ), je.jointEvents[i].userData == NULL );
	}
	(void)origin;
}

typedef struct Hits
{
	uint64_t mask;
	int count;
	uint32_t generations[64];
} Hits;
static bool hit( b2ShapeId s, void* context )
{
	Hits* h = context;
	if ( s.index1 <= 0 || s.index1 >= 64 )
		abort();
	h->mask |= UINT64_C( 1 ) << s.index1;
	h->count += 1;
	h->generations[s.index1] = s.generation;
	return true;
}
static void query( b2WorldId w, b2Pos origin, const char* key )
{
	char owned[256];
	snprintf( owned, sizeof( owned ), "%s", key );
	key = owned;
	Hits hits = { 0 };
	b2AABB aabb = { { -30, -5 }, { 30, 15 } };
	b2World_OverlapAABB( w, origin, aabb, b2DefaultQueryFilter(), hit, &hits );
	ORACLE_I32( ORACLE_KEY( "%s/queryCount", key ), hits.count );
	ORACLE_U64( ORACLE_KEY( "%s/querySet", key ), hits.mask );
	for ( int i = 1; i < 64; ++i )
		if ( hits.mask & ( UINT64_C( 1 ) << i ) )
			ORACLE_U32( ORACLE_KEY( "%s/query/%d/generation", key, i ), hits.generations[i] );
}

static int families( b2Pos origin, const char* name )
{
	ORACLE_OPEN( name );
	b2WorldDef wd = b2DefaultWorldDef();
	wd.gravity = (b2Vec2){ 0, -2 };
	b2WorldId w = b2CreateWorld( &wd );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.position = origin;
	b2BodyId ground = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.enableContactEvents = true;
	sd.enableHitEvents = true;
	sd.enableSensorEvents = true;
	b2Segment segment = { { -25, 0 }, { 25, 0 } };
	b2ShapeId edge = b2CreateSegmentShape( ground, &sd, &segment );
	b2BodyId bodies[5];
	b2ShapeId shapes[5];
	b2JointId joints[5];
	for ( int i = 0; i < 5; ++i )
	{
		bd = b2DefaultBodyDef();
		bd.type = b2_dynamicBody;
		bd.position = (b2Pos){ origin.x + 3 * i, origin.y + 3 };
		bodies[i] = b2CreateBody( w, &bd );
		if ( i % 2 == 0 )
		{
			b2Circle c = { { 0.125f, 0 }, 0.4f };
			shapes[i] = b2CreateCircleShape( bodies[i], &sd, &c );
		}
		else
		{
			b2Capsule c = { { -0.3f, 0 }, { 0.3f, 0 }, 0.25f };
			shapes[i] = b2CreateCapsuleShape( bodies[i], &sd, &c );
		}
		b2JointDef base = b2DefaultFilterJointDef().base;
		base.bodyIdA = ground;
		base.bodyIdB = bodies[i];
		base.localFrameA.p = (b2Vec2){ 3 * i, 3 };
		base.forceThreshold = 0.5f;
		base.torqueThreshold = 0.5f;
		if ( i == 0 )
		{
			b2WheelJointDef d = b2DefaultWheelJointDef();
			d.base = base;
			d.enableMotor = true;
			d.motorSpeed = 2;
			d.maxMotorTorque = 3;
			d.enableLimit = true;
			d.lowerTranslation = -1;
			d.upperTranslation = 1;
			joints[i] = b2CreateWheelJoint( w, &d );
		}
		if ( i == 1 )
		{
			b2MotorJointDef d = b2DefaultMotorJointDef();
			d.base = base;
			d.linearVelocity = (b2Vec2){ 0.5f, -0.25f };
			d.maxVelocityForce = 4;
			d.angularVelocity = 1;
			d.maxVelocityTorque = 2;
			d.linearHertz = 2;
			d.linearDampingRatio = 0.7f;
			d.maxSpringForce = 5;
			d.angularHertz = 1;
			d.angularDampingRatio = 0.5f;
			d.maxSpringTorque = 3;
			joints[i] = b2CreateMotorJoint( w, &d );
		}
		if ( i == 2 )
		{
			b2MoverJointDef d = b2DefaultMoverJointDef();
			d.base = base;
			d.linearVelocity = (b2Vec2){ 1, -0.5f };
			d.maxVelocityForce = (b2Vec2){ 2, 3 };
			joints[i] = b2CreateMoverJoint( w, &d );
		}
		if ( i == 3 )
		{
			b2PogoJointDef d = b2DefaultPogoJointDef();
			d.base = base;
			d.normal = (b2Vec2){ 0, 1 };
			d.hertz = 3;
			d.dampingRatio = 0.6f;
			d.restLength = 0.5f;
			d.maxTensionForce = 2;
			d.maxCompressionForce = 6;
			d.impulse = 0.125f;
			d.velocity = 0.25f;
			joints[i] = b2CreatePogoJoint( w, &d );
		}
		if ( i == 4 )
		{
			b2FilterJointDef d = b2DefaultFilterJointDef();
			d.base = base;
			joints[i] = b2CreateFilterJoint( w, &d );
		}
	}
	bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = (b2Pos){ origin.x - 10, origin.y + 5 };
	bd.isEnabled = false;
	b2BodyId disabled = b2CreateBody( w, &bd );
	b2Circle circle = { { 0, 0 }, 0.5f };
	b2CreateCircleShape( disabled, &sd, &circle );
	bd.position.x = origin.x - 15;
	bd.isEnabled = true;
	bd.isAwake = false;
	b2BodyId sleeping = b2CreateBody( w, &bd );
	b2CreateCircleShape( sleeping, &sd, &circle );
	bd = b2DefaultBodyDef();
	bd.position = (b2Pos){ origin.x - 5, origin.y + 4 };
	b2BodyId sensorBody = b2CreateBody( w, &bd );
	b2ShapeDef sensorDef = b2DefaultShapeDef();
	sensorDef.isSensor = true;
	sensorDef.enableSensorEvents = true;
	circle.radius = 1.25f;
	b2ShapeId sensor = b2CreateCircleShape( sensorBody, &sensorDef, &circle );
	bd.type = b2_dynamicBody;
	bd.position.x = origin.x - 8;
	bd.linearVelocity = (b2Vec2){ 4, 0 };
	bd.gravityScale = 0;
	b2BodyId visitor = b2CreateBody( w, &bd );
	circle.radius = 0.25f;
	b2CreateCircleShape( visitor, &sd, &circle );
	for ( int i = 0; i < 20; ++i )
		b2World_Step( w, 1.0f / 60, 4 );
	int size;
	uint8_t* image = capture( w, &size );
	uint64_t initial = b2World_GetStateHash( w ), tail[90];
	ORACLE_U64( "capture/state", initial );
	ORACLE_BOOL( "capture/disabled", b2Body_IsEnabled( disabled ) );
	ORACLE_BOOL( "capture/sleeping", b2Body_IsAwake( sleeping ) );
	for ( int i = 0; i < 90; ++i )
	{
		if ( i == 5 )
			b2Body_Enable( disabled );
		if ( i == 9 )
			b2Body_SetAwake( sleeping, true );
		b2World_Step( w, 1.0f / 60, 4 );
		tail[i] = b2World_GetStateHash( w );
		observe( w, origin, ORACLE_KEY( "reference/%d", i ) );
		query( w, origin, ORACLE_KEY( "reference/%d", i ) );
	}
	b2Body_Enable( disabled );
	b2Body_SetAwake( sleeping, true );
	b2Body_SetLinearVelocity( visitor, (b2Vec2){ -10, 2 } );
	b2WheelJoint_SetMotorSpeed( joints[0], -7 );
	b2MotorJoint_SetLinearVelocity( joints[1], (b2Vec2){ -4, 5 } );
	b2MoverJoint_SetLinearVelocity( joints[2], (b2Vec2){ -3, 2 } );
	b2PogoJoint_SetRestLength( joints[3], 2 );
	b2Shape_SetCircle( shapes[0], &circle );
	b2DestroyJoint( joints[4] );
	b2Capsule changedCapsule = { { -1, 0 }, { 1, 0 }, 0.75f };
	b2Shape_SetCapsule( shapes[1], &changedCapsule );
	b2Segment changedSegment = { { -10, -1 }, { 10, -1 } };
	b2Shape_SetSegment( edge, &changedSegment );
	b2Body_SetTransform( visitor, (b2Pos){ origin.x - 5, origin.y + 4 }, (b2Rot){ 1, 0 } );
	b2World_Step( w, 1.0f / 60, 4 );
	b2Body_Disable( visitor );
	b2World_Step( w, 1.0f / 60, 4 );
	b2World* internal = b2GetWorldFromId( w );
	int queuedSensorEnds = internal->sensorEndEvents[0].count + internal->sensorEndEvents[1].count;
	ENSURE( queuedSensorEnds >= 1 );
	ORACLE_I32( "mutation/queuedSensorEnds", queuedSensorEnds );
	ENSURE( b2World_Restore( w, image, size ) );
	ENSURE( b2World_GetStateHash( w ) == initial );
	ORACLE_U64( "restore/state", b2World_GetStateHash( w ) );
	ORACLE_I32( "restore/sensorEnd0", internal->sensorEndEvents[0].count );
	ORACLE_I32( "restore/sensorEnd1", internal->sensorEndEvents[1].count );
	ENSURE( internal->sensorEndEvents[0].count == 0 && internal->sensorEndEvents[1].count == 0 );
	ORACLE_BOOL( "restore/disabled", b2Body_IsEnabled( disabled ) );
	ORACLE_BOOL( "restore/sleeping", b2Body_IsAwake( sleeping ) );
	ORACLE_F32( "restore/wheelMotor", b2WheelJoint_GetMotorSpeed( joints[0] ) );
	ORACLE_VEC2( "restore/motorVelocity", b2MotorJoint_GetLinearVelocity( joints[1] ) );
	ORACLE_VEC2( "restore/moverVelocity", b2MoverJoint_GetLinearVelocity( joints[2] ) );
	ORACLE_F32( "restore/pogoLength", b2PogoJoint_GetRestLength( joints[3] ) );
	for ( int i = 0; i < 5; ++i )
	{
		ORACLE_I32( ORACLE_KEY( "restore/joint/%d/type", i ), b2Joint_GetType( joints[i] ) );
		ORACLE_I32( ORACLE_KEY( "restore/shape/%d/type", i ), b2Shape_GetType( shapes[i] ) );
	}
	b2Circle restoredCircle = b2Shape_GetCircle( shapes[0] );
	ORACLE_VEC2( "restore/circleCenter", restoredCircle.center );
	ORACLE_F32( "restore/circleRadius", restoredCircle.radius );
	b2Capsule cap = b2Shape_GetCapsule( shapes[1] );
	ORACLE_VEC2( "restore/capsule1", cap.center1 );
	ORACLE_VEC2( "restore/capsule2", cap.center2 );
	ORACLE_F32( "restore/capsuleRadius", cap.radius );
	segment = b2Shape_GetSegment( edge );
	ORACLE_VEC2( "restore/segment1", segment.point1 );
	ORACLE_VEC2( "restore/segment2", segment.point2 );
	ORACLE_I32( "restore/sensorOverlaps", b2Shape_GetSensorCapacity( sensor ) );
	for ( int i = 0; i < 90; ++i )
	{
		if ( i == 5 )
			b2Body_Enable( disabled );
		if ( i == 9 )
			b2Body_SetAwake( sleeping, true );
		b2World_Step( w, 1.0f / 60, 4 );
		ENSURE( b2World_GetStateHash( w ) == tail[i] );
		observe( w, origin, ORACLE_KEY( "rollback/%d", i ) );
		query( w, origin, ORACLE_KEY( "rollback/%d", i ) );
	}
	free( image );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
	return 0;
}

static int handles( void )
{
	ORACLE_OPEN( "snapshot/Handles" );
	b2WorldDef wd = b2DefaultWorldDef();
	b2WorldId w = b2CreateWorld( &wd );
	b2BodyDef bd = b2DefaultBodyDef();
	b2BodyId ground = b2CreateBody( w, &bd );
	bd.type = b2_dynamicBody;
	bd.position = (b2Pos){ 0, 4 };
	b2BodyId body = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	b2Circle c = { { 0, 0 }, 0.5f };
	b2ShapeId shape = b2CreateCircleShape( body, &sd, &c );
	b2FilterJointDef jd = b2DefaultFilterJointDef();
	jd.base.bodyIdA = ground;
	jd.base.bodyIdB = body;
	b2JointId joint = b2CreateFilterJoint( w, &jd );
	b2Vec2 points[] = { { -2, 0 }, { 0, 0 }, { 2, 0 } };
	b2SurfaceMaterial mat = b2DefaultSurfaceMaterial();
	b2ChainDef cd = b2DefaultChainDef();
	cd.points = points;
	cd.pointCount = 3;
	cd.ghost1 = (b2Vec2){ -4, 0 };
	cd.ghost2 = (b2Vec2){ 4, 0 };
	cd.materials = &mat;
	cd.materialCount = 1;
	b2ChainId chain = b2CreateChain( ground, &cd );
	int size;
	uint8_t* image = capture( w, &size );
	b2DestroyJoint( joint );
	b2DestroyChain( chain );
	b2DestroyShape( shape, true );
	b2DestroyBody( body );
	b2BodyId replacement = b2CreateBody( w, &bd );
	b2ShapeId replacementShape = b2CreateCircleShape( replacement, &sd, &c );
	jd.base.bodyIdB = replacement;
	b2JointId replacementJoint = b2CreateFilterJoint( w, &jd );
	b2ChainId replacementChain = b2CreateChain( ground, &cd );
	b2BodyId extra = b2CreateBody( w, &bd );
	b2ShapeId extraShape = b2CreateCircleShape( extra, &sd, &c );
	jd.base.bodyIdB = extra;
	b2JointId extraJoint = b2CreateFilterJoint( w, &jd );
	b2ChainId extraChain = b2CreateChain( ground, &cd );
	ORACLE_BOOL( "reuse/bodySlot", body.index1 == replacement.index1 );
	ORACLE_BOOL( "reuse/bodyGeneration", body.generation != replacement.generation );
	ORACLE_BOOL( "reuse/shapeSlot", shape.index1 == replacementShape.index1 );
	ORACLE_BOOL( "reuse/shapeGeneration", shape.generation != replacementShape.generation );
	ORACLE_BOOL( "reuse/jointSlot", joint.index1 == replacementJoint.index1 );
	ORACLE_BOOL( "reuse/jointGeneration", joint.generation != replacementJoint.generation );
	ORACLE_BOOL( "reuse/chainSlot", chain.index1 == replacementChain.index1 );
	ORACLE_BOOL( "reuse/chainGeneration", chain.generation != replacementChain.generation );
	ENSURE( b2World_Restore( w, image, size ) );
	ORACLE_BOOL( "held/body", b2Body_IsValid( body ) );
	ORACLE_BOOL( "held/shape", b2Shape_IsValid( shape ) );
	ORACLE_BOOL( "held/joint", b2Joint_IsValid( joint ) );
	ORACLE_BOOL( "held/chain", b2Chain_IsValid( chain ) );
	ORACLE_BOOL( "replacement/body", b2Body_IsValid( replacement ) );
	ORACLE_BOOL( "replacement/shape", b2Shape_IsValid( replacementShape ) );
	ORACLE_BOOL( "replacement/joint", b2Joint_IsValid( replacementJoint ) );
	ORACLE_BOOL( "replacement/chain", b2Chain_IsValid( replacementChain ) );
	ORACLE_BOOL( "extra/body", b2Body_IsValid( extra ) );
	ORACLE_BOOL( "extra/shape", b2Shape_IsValid( extraShape ) );
	ORACLE_BOOL( "extra/joint", b2Joint_IsValid( extraJoint ) );
	ORACLE_BOOL( "extra/chain", b2Chain_IsValid( extraChain ) );
	ENSURE( b2Body_IsValid( body ) && b2Shape_IsValid( shape ) && b2Joint_IsValid( joint ) && b2Chain_IsValid( chain ) );
	b2Body_SetLinearVelocity( body, (b2Vec2){ 2, 1 } );
	b2Shape_SetDensity( shape, 2, true );
	mat.friction = 0.75f;
	b2Chain_SetAllSurfaceMaterials( chain, &mat );
	b2Joint_SetCollideConnected( joint, true );
	b2World_Step( w, 1.0f / 60, 4 );
	ORACLE_POS( "held/position", b2Body_GetPosition( body ) );
	ORACLE_F32( "held/density", b2Shape_GetDensity( shape ) );
	ORACLE_F32( "held/chainFriction", b2Chain_GetSurfaceMaterial( chain, 0 ).friction );
	ORACLE_BOOL( "held/collide", b2Joint_GetCollideConnected( joint ) );
	ORACLE_U64( "held/state", b2World_GetStateHash( w ) );
	for ( int workers = 0; workers <= 4; workers += 4 )
	{
		b2WorldId clone = b2CreateWorldFromSnapshot( image, size, workers );
		ENSURE( b2World_IsValid( clone ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/identity", workers ), clone.index1 != w.index1 );
		ORACLE_I32( ORACLE_KEY( "new/%d/workers", workers ), b2GetWorldFromId( clone )->workerCount );
		b2BodyId mappedBody = body;
		mappedBody.world0 = clone.index1 - 1;
		b2ShapeId mappedShape = shape;
		mappedShape.world0 = clone.index1 - 1;
		b2JointId mappedJoint = joint;
		mappedJoint.world0 = clone.index1 - 1;
		b2ChainId mappedChain = chain;
		mappedChain.world0 = clone.index1 - 1;
		ORACLE_BOOL( ORACLE_KEY( "new/%d/bodyValid", workers ), b2Body_IsValid( mappedBody ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/shapeBody", workers ), B2_ID_EQUALS( b2Shape_GetBody( mappedShape ), mappedBody ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/jointBody", workers ), B2_ID_EQUALS( b2Joint_GetBodyB( mappedJoint ), mappedBody ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/bodyWorld", workers ),
					 b2StoreWorldId( b2Body_GetWorld( mappedBody ) ) == b2StoreWorldId( clone ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/shapeWorld", workers ),
					 b2StoreWorldId( b2Shape_GetWorld( mappedShape ) ) == b2StoreWorldId( clone ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/jointWorld", workers ),
					 b2StoreWorldId( b2Joint_GetWorld( mappedJoint ) ) == b2StoreWorldId( clone ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/chainWorld", workers ),
					 b2StoreWorldId( b2Chain_GetWorld( mappedChain ) ) == b2StoreWorldId( clone ) );
		ORACLE_BOOL( ORACLE_KEY( "new/%d/originOwner", workers ),
					 b2StoreWorldId( b2Body_GetWorld( body ) ) == b2StoreWorldId( w ) );
		b2World_Step( clone, 1.0f / 60, 4 );
		ORACLE_U64( ORACLE_KEY( "new/%d/state", workers ), b2World_GetStateHash( clone ) );
		b2DestroyWorld( clone );
	}
	free( image );
	b2DestroyWorld( w );
	ORACLE_CLOSE();
	return 0;
}

static int headers( void )
{
	ORACLE_OPEN( "snapshot/Headers" );
	b2WorldDef wd = b2DefaultWorldDef();
	b2WorldId w = b2CreateWorld( &wd );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = (b2Pos){ 2, 4 };
	bd.name = "header-held";
	b2BodyId body = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.material.friction = 0.625f;
	b2Circle c = { { 0, 0 }, 0.5f };
	b2ShapeId shape = b2CreateCircleShape( body, &sd, &c );
	b2BodyDef staticDef = b2DefaultBodyDef();
	b2BodyId anchor = b2CreateBody( w, &staticDef );
	b2FilterJointDef jd = b2DefaultFilterJointDef();
	jd.base.bodyIdA = anchor;
	jd.base.bodyIdB = body;
	b2JointId joint = b2CreateFilterJoint( w, &jd );
	b2Vec2 points[] = { { -2, 0 }, { 0, 0 }, { 2, 0 } };
	b2SurfaceMaterial material = b2DefaultSurfaceMaterial();
	material.friction = 0.375f;
	b2ChainDef cd = b2DefaultChainDef();
	cd.points = points;
	cd.pointCount = 3;
	cd.ghost1 = (b2Vec2){ -4, 0 };
	cd.ghost2 = (b2Vec2){ 4, 0 };
	cd.materials = &material;
	cd.materialCount = 1;
	b2ChainId chain = b2CreateChain( anchor, &cd );
	int size;
	uint8_t* image = capture( w, &size );
	uint8_t* bad = malloc( size );
	uint64_t hash = b2World_GetStateHash( w );
	const char* cases[] = { "nil", "short", "magic", "version", "layout", "precision" };
	for ( int i = 0; i < 6; ++i )
	{
		memcpy( bad, image, size );
		int n = size;
		const uint8_t* input = bad;
		if ( i == 0 )
		{
			input = NULL;
			n = 0;
		}
		if ( i == 1 )
			n = 15;
		// Native header consists of four uint32 fields; mutate each producer's own image only.
		if ( i >= 2 )
		{
			int offset = ( i - 2 ) * 4;
			uint32_t value;
			memcpy( &value, bad + offset, 4 );
			value ^= i == 5 ? 2u : 1u;
			memcpy( bad + offset, &value, 4 );
		}
		bool ok = b2World_Restore( w, input, n );
		ENSURE( !ok );
		ENSURE( b2World_GetStateHash( w ) == hash );
		ORACLE_BOOL( ORACLE_KEY( "%s/rejected", cases[i] ), !ok );
		ORACLE_U64( ORACLE_KEY( "%s/state", cases[i] ), b2World_GetStateHash( w ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/bodyValid", cases[i] ), b2Body_IsValid( body ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/shapeValid", cases[i] ), b2Shape_IsValid( shape ) );
		ORACLE_POS( ORACLE_KEY( "%s/position", cases[i] ), b2Body_GetPosition( body ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/name", cases[i] ), strcmp( b2Body_GetName( body ), "header-held" ) == 0 );
		ORACLE_F32( ORACLE_KEY( "%s/friction", cases[i] ), b2Shape_GetFriction( shape ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/jointValid", cases[i] ), b2Joint_IsValid( joint ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/chainValid", cases[i] ), b2Chain_IsValid( chain ) );
		ORACLE_BOOL( ORACLE_KEY( "%s/jointBody", cases[i] ), B2_ID_EQUALS( b2Joint_GetBodyB( joint ), body ) );
		ORACLE_F32( ORACLE_KEY( "%s/chainFriction", cases[i] ), b2Chain_GetSurfaceMaterial( chain, 0 ).friction );
		ORACLE_VEC2( ORACLE_KEY( "%s/gravity", cases[i] ), b2World_GetGravity( w ) );
		b2WorldId invalid = b2CreateWorldFromSnapshot( input, n, 0 );
		ORACLE_BOOL( ORACLE_KEY( "%s/newRejected", cases[i] ), !b2World_IsValid( invalid ) );
		ENSURE( !b2World_IsValid( invalid ) );
	}
	// Truncate each engine's own image at the final payload boundary.
	bool ok = b2World_Restore( w, image, size - 1 );
	ORACLE_BOOL( "late/rejected", !ok );
	ENSURE( !ok );
	b2DestroyWorld( w );
	b2WorldId invalid = b2CreateWorldFromSnapshot( image, size - 1, 0 );
	ORACLE_BOOL( "late/newRejected", !b2World_IsValid( invalid ) );
	ENSURE( !b2World_IsValid( invalid ) );
	wd = b2DefaultWorldDef();
	w = b2CreateWorld( &wd );
	body = b2CreateBody( w, &bd );
	ORACLE_BOOL( "cleanup/newBody", b2Body_IsValid( body ) );
	b2DestroyWorld( w );
	free( image );
	free( bad );
	ORACLE_CLOSE();
	return 0;
}

static int taskCount, finishCount, oldFilter, liveFilter, oldPre, livePre, oldFriction, liveFriction, oldRestitution,
	liveRestitution;
static int oldTag, liveTag;
static void* enqueue( b2TaskCallback* task, void* context, void* user )
{
	if ( user != &liveTag )
		abort();
	++taskCount;
	task( context );
	return &liveTag;
}
static void finish( void* task, void* user )
{
	if ( task != &liveTag || user != &liveTag )
		abort();
	++finishCount;
}
static bool filterOld( b2ShapeId a, b2ShapeId b, void* context )
{
	(void)a;
	(void)b;
	if ( context != &oldTag )
		abort();
	++oldFilter;
	return false;
}
static bool filterLive( b2ShapeId a, b2ShapeId b, void* context )
{
	(void)a;
	(void)b;
	if ( context != &liveTag )
		abort();
	++liveFilter;
	return true;
}
static void preOld( b2ShapeId a, b2ShapeId b, b2Manifold* m, void* context )
{
	(void)a;
	(void)b;
	(void)m;
	if ( context != &oldTag )
		abort();
	++oldPre;
}
static void preLive( b2ShapeId a, b2ShapeId b, b2Manifold* m, void* context )
{
	(void)a;
	(void)b;
	if ( context != &liveTag )
		abort();
	++livePre;
	m->pointCount = 0;
}
static float frictionOld( float a, uint64_t x, float b, uint64_t y )
{
	(void)a;
	(void)x;
	(void)b;
	(void)y;
	++oldFriction;
	return 0.9f;
}
static float frictionLive( float a, uint64_t x, float b, uint64_t y )
{
	(void)a;
	(void)x;
	(void)b;
	(void)y;
	++liveFriction;
	return 0.125f;
}
static float restitutionOld( float a, uint64_t x, float b, uint64_t y )
{
	(void)a;
	(void)x;
	(void)b;
	(void)y;
	++oldRestitution;
	return 0.9f;
}
static float restitutionLive( float a, uint64_t x, float b, uint64_t y )
{
	(void)a;
	(void)x;
	(void)b;
	(void)y;
	++liveRestitution;
	return 0.375f;
}
static void contactMaterials( b2WorldId id, const char* key )
{
	b2World* w = b2GetWorldFromId( id );
	int count = 0;
	for ( int i = 0; i < w->contacts.count; ++i )
		if ( w->contacts.data[i].contactId == i )
			++count;
	ORACLE_I32( ORACLE_KEY( "%s/contactCount", key ), count );
	for ( int i = 0; i < w->contacts.count; ++i )
	{
		b2Contact* c = w->contacts.data + i;
		if ( c->contactId != i )
			continue;
		b2ContactSim* sim = b2GetContactSim( w, c );
		ORACLE_I32( ORACLE_KEY( "%s/contact/%d/id", key, i ), i );
		ORACLE_F32( ORACLE_KEY( "%s/contact/%d/friction", key, i ), sim->friction );
		ORACLE_F32( ORACLE_KEY( "%s/contact/%d/restitution", key, i ), sim->restitution );
		ORACLE_I32( ORACLE_KEY( "%s/contact/%d/points", key, i ), sim->manifold.pointCount );
	}
}

static int wiring( void )
{
	ORACLE_OPEN( "snapshot/Wiring" );
	b2WorldDef wd = b2DefaultWorldDef();
	wd.workerCount = 2;
	wd.enqueueTask = enqueue;
	wd.finishTask = finish;
	wd.userTaskContext = &liveTag;
	wd.userData = &oldTag;
	wd.frictionCallback = frictionOld;
	wd.restitutionCallback = restitutionOld;
	b2WorldId w = b2CreateWorld( &wd );
	b2World_SetCustomFilterCallback( w, filterOld, &oldTag );
	b2World_SetPreSolveCallback( w, preOld, NULL, &oldTag );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.userData = &oldTag;
	b2BodyId ground = b2CreateBody( w, &bd );
	bd.type = b2_dynamicBody;
	bd.position = (b2Pos){ 0, 0.75f };
	b2BodyId body = b2CreateBody( w, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.userData = &oldTag;
	sd.enableCustomFiltering = true;
	sd.enablePreSolveEvents = true;
	sd.enableContactEvents = true;
	b2Polygon box = b2MakeBox( 0.5f, 0.5f );
	b2CreatePolygonShape( ground, &sd, &box );
	b2ShapeId shape = b2CreatePolygonShape( body, &sd, &box );
	b2FilterJointDef jd = b2DefaultFilterJointDef();
	jd.base.bodyIdA = ground;
	jd.base.bodyIdB = body;
	jd.base.collideConnected = true;
	jd.base.userData = &oldTag;
	b2JointId joint = b2CreateFilterJoint( w, &jd );
	int size;
	uint8_t* image = capture( w, &size );
	b2World_SetUserData( w, &liveTag );
	b2Body_SetUserData( body, &liveTag );
	b2Shape_SetUserData( shape, &liveTag );
	b2Joint_SetUserData( joint, &liveTag );
	b2World_SetCustomFilterCallback( w, filterLive, &liveTag );
	b2World_SetPreSolveCallback( w, preLive, NULL, &liveTag );
	b2World_SetFrictionCallback( w, frictionLive );
	b2World_SetRestitutionCallback( w, restitutionLive );
	ENSURE( b2World_Restore( w, image, size ) );
	ORACLE_BOOL( "restore/worldData", b2World_GetUserData( w ) == &liveTag );
	ORACLE_BOOL( "restore/bodyData", b2Body_GetUserData( body ) == NULL );
	ORACLE_BOOL( "restore/shapeData", b2Shape_GetUserData( shape ) == NULL );
	ORACLE_BOOL( "restore/jointData", b2Joint_GetUserData( joint ) == NULL );
	ORACLE_I32( "restore/workers", b2GetWorldFromId( w )->workerCount );
	b2World_Step( w, 1.0f / 60, 4 );
	ENSURE( taskCount > 0 && taskCount == finishCount );
	ENSURE( liveFilter > 0 && livePre > 0 && liveFriction > 0 && liveRestitution > 0 );
	ORACLE_BOOL( "live/tasks", taskCount > 0 );
	ORACLE_BOOL( "live/finished", taskCount == finishCount );
	ORACLE_I32( "live/filter", liveFilter );
	ORACLE_I32( "live/pre", livePre );
	ORACLE_I32( "live/friction", liveFriction );
	ORACLE_I32( "live/restitution", liveRestitution );
	ORACLE_I32( "old/filter", oldFilter );
	ORACLE_I32( "old/pre", oldPre );
	ORACLE_I32( "old/friction", oldFriction );
	ORACLE_I32( "old/restitution", oldRestitution );
	ORACLE_POS( "live/position", b2Body_GetPosition( body ) );
	ORACLE_U64( "live/state", b2World_GetStateHash( w ) );
	contactMaterials( w, "live" );
	int beforeFilter = liveFilter, beforePre = livePre, beforeFriction = liveFriction, beforeRestitution = liveRestitution,
		beforeTasks = taskCount;
	b2WorldId clone = b2CreateWorldFromSnapshot( image, size, 0 );
	ENSURE( b2World_IsValid( clone ) );
	b2BodyId cb = body;
	cb.world0 = clone.index1 - 1;
	b2ShapeId cs = shape;
	cs.world0 = clone.index1 - 1;
	b2JointId cj = joint;
	cj.world0 = clone.index1 - 1;
	ORACLE_BOOL( "new/worldData", b2World_GetUserData( clone ) == NULL );
	ORACLE_BOOL( "new/bodyData", b2Body_GetUserData( cb ) == NULL );
	ORACLE_BOOL( "new/shapeData", b2Shape_GetUserData( cs ) == NULL );
	ORACLE_BOOL( "new/jointData", b2Joint_GetUserData( cj ) == NULL );
	ORACLE_I32( "new/workers", b2GetWorldFromId( clone )->workerCount );
	b2World_Step( clone, 1.0f / 60, 4 );
	ORACLE_BOOL( "new/noLiveCalls", liveFilter == beforeFilter && livePre == beforePre && liveFriction == beforeFriction &&
										liveRestitution == beforeRestitution && taskCount == beforeTasks );
	ORACLE_POS( "new/position", b2Body_GetPosition( cb ) );
	ORACLE_U64( "new/state", b2World_GetStateHash( clone ) );
	contactMaterials( clone, "new" );
	b2DestroyWorld( clone );
	b2DestroyWorld( w );
	free( image );
	ORACLE_CLOSE();
	return 0;
}
int main( void )
{
	if ( families( (b2Pos){ 0, 0 }, "snapshot/Families" ) )
		return 1;
#if defined( BOX2D_DOUBLE_PRECISION )
	if ( families( (b2Pos){ 1e7, -1e7 }, "snapshot/DistantFamilies" ) )
		return 1;
#endif
	if ( handles() || headers() || wiring() )
		return 1;
	return 0;
}
