// SPDX-License-Identifier: MIT
#include "oracle.h"

#include "box2d/box2d.h"

#include <assert.h>
#include <float.h>
#include <math.h>
#include <stdint.h>
#include <stdlib.h>

// World slots are runner-owned. Preserve every object index/generation and explicitly
// assert its world membership rather than compare the runner's absolute slot number.
static b2WorldId evWorld;
static int evTags[128];
static b2ShapeId evShapes[128];
static int evShapeCount;
static b2ContactId evHistory[128];
static int evHistoryCount;
static void evShape( b2ShapeId s )
{
	ORACLE_I32( "shape.index", s.index1 );
	ORACLE_U32( "shape.generation", s.generation );
	ORACLE_BOOL( "shape.world", s.world0 == evWorld.index1 - 1 );
	ORACLE_BOOL( "shape.valid", b2Shape_IsValid( s ) );
	int tag = 0;
	for ( int i = 0; i < evShapeCount; ++i )
		if ( B2_ID_EQUALS( s, evShapes[i] ) )
			tag = evTags[i];
	ORACLE_I32( "shape.tag", tag );
}
static void evBodyId( b2BodyId b )
{
	ORACLE_I32( "body.index", b.index1 );
	ORACLE_U32( "body.generation", b.generation );
	ORACLE_BOOL( "body.world", b.world0 == evWorld.index1 - 1 );
	ORACLE_BOOL( "body.valid", b2Body_IsValid( b ) );
}
static void evContact( b2ContactId c )
{
	ORACLE_I32( "contact.index", c.index1 );
	ORACLE_U32( "contact.generation", c.generation );
	ORACLE_BOOL( "contact.world", c.world0 == evWorld.index1 - 1 );
	ORACLE_BOOL( "contact.valid", b2Contact_IsValid( c ) );
}
static int evPairCompare( b2ShapeId a, b2ShapeId b )
{
	if ( a.index1 != b.index1 )
		return a.index1 < b.index1 ? -1 : 1;
	return ( a.generation > b.generation ) - ( a.generation < b.generation );
}
#define EV_CMP( name, type, first, second )                                                                                      \
	static int name( const void* av, const void* bv )                                                                            \
	{                                                                                                                            \
		const type* a = av;                                                                                                      \
		const type* b = bv;                                                                                                      \
		int x = evPairCompare( a->first, b->first );                                                                             \
		return x ? x : evPairCompare( a->second, b->second );                                                                    \
	}
EV_CMP( evCB, b2ContactBeginTouchEvent, shapeIdA, shapeIdB )
EV_CMP( evCE, b2ContactEndTouchEvent, shapeIdA, shapeIdB )
EV_CMP( evCH, b2ContactHitEvent, shapeIdA, shapeIdB )
EV_CMP( evSB, b2SensorBeginTouchEvent, sensorShapeId, visitorShapeId )
EV_CMP( evSE, b2SensorEndTouchEvent, sensorShapeId, visitorShapeId )
EV_CMP( evCD, b2ContactData, shapeIdA, shapeIdB )
static int evBM( const void* a, const void* b )
{
	return ( (const b2BodyMoveEvent*)a )->bodyId.index1 - ( (const b2BodyMoveEvent*)b )->bodyId.index1;
}
static int evJE( const void* a, const void* b )
{
	return ( (const b2JointEvent*)a )->jointId.index1 - ( (const b2JointEvent*)b )->jointId.index1;
}
static void evObserve( b2BodyId* bodies, int count )
{
	b2ContactEvents c = b2World_GetContactEvents( evWorld );
	b2SensorEvents s = b2World_GetSensorEvents( evWorld );
	b2BodyEvents m = b2World_GetBodyEvents( evWorld );
	b2JointEvents j = b2World_GetJointEvents( evWorld );
	// Sort copies: ordering is not promised, multiplicity remains visible.
	b2ContactBeginTouchEvent cb[32];
	b2ContactEndTouchEvent ce[32];
	b2ContactHitEvent ch[32];
	b2SensorBeginTouchEvent sb[32];
	b2SensorEndTouchEvent se[32];
	b2BodyMoveEvent bm[32];
	b2JointEvent je[32];
	assert( c.beginCount <= 32 && c.endCount <= 32 && c.hitCount <= 32 && s.beginCount <= 32 && s.endCount <= 32 &&
			m.moveCount <= 32 && j.count <= 32 );
	for ( int i = 0; i < c.beginCount; ++i )
		cb[i] = c.beginEvents[i];
	for ( int i = 0; i < c.endCount; ++i )
		ce[i] = c.endEvents[i];
	for ( int i = 0; i < c.hitCount; ++i )
		ch[i] = c.hitEvents[i];
	for ( int i = 0; i < s.beginCount; ++i )
		sb[i] = s.beginEvents[i];
	for ( int i = 0; i < s.endCount; ++i )
		se[i] = s.endEvents[i];
	for ( int i = 0; i < m.moveCount; ++i )
		bm[i] = m.moveEvents[i];
	for ( int i = 0; i < j.count; ++i )
		je[i] = j.jointEvents[i];
	qsort( cb, c.beginCount, sizeof( *cb ), evCB );
	qsort( ce, c.endCount, sizeof( *ce ), evCE );
	qsort( ch, c.hitCount, sizeof( *ch ), evCH );
	qsort( sb, s.beginCount, sizeof( *sb ), evSB );
	qsort( se, s.endCount, sizeof( *se ), evSE );
	qsort( bm, m.moveCount, sizeof( *bm ), evBM );
	qsort( je, j.count, sizeof( *je ), evJE );
	for ( int i = 0; i < c.beginCount; ++i )
	{
		b2ContactId id = cb[i].contactId;
		bool found = false;
		for ( int k = 0; k < evHistoryCount; ++k )
			found |= B2_ID_EQUALS( id, evHistory[k] );
		if ( !found )
		{
			assert( evHistoryCount < 128 );
			evHistory[evHistoryCount++] = id;
		}
	}
	ORACLE_I32( "history.count", evHistoryCount );
	for ( int i = 0; i < evHistoryCount; ++i )
		evContact( evHistory[i] );
	ORACLE_I32( "contact.begin", c.beginCount );
	for ( int i = 0; i < c.beginCount; ++i )
	{
		evShape( cb[i].shapeIdA );
		evShape( cb[i].shapeIdB );
		evContact( cb[i].contactId );
	}
	ORACLE_I32( "contact.end", c.endCount );
	for ( int i = 0; i < c.endCount; ++i )
	{
		evShape( ce[i].shapeIdA );
		evShape( ce[i].shapeIdB );
		evContact( ce[i].contactId );
	}
	ORACLE_I32( "contact.hit", c.hitCount );
	for ( int i = 0; i < c.hitCount; ++i )
	{
		evShape( ch[i].shapeIdA );
		evShape( ch[i].shapeIdB );
		evContact( ch[i].contactId );
		ORACLE_POS( "hit.point", ch[i].point );
		ORACLE_VEC2( "hit.normal", ch[i].normal );
		ORACLE_F32( "hit.speed", ch[i].approachSpeed );
	}
	ORACLE_I32( "sensor.begin", s.beginCount );
	for ( int i = 0; i < s.beginCount; ++i )
	{
		evShape( sb[i].sensorShapeId );
		evShape( sb[i].visitorShapeId );
	}
	ORACLE_I32( "sensor.end", s.endCount );
	for ( int i = 0; i < s.endCount; ++i )
	{
		evShape( se[i].sensorShapeId );
		evShape( se[i].visitorShapeId );
	}
	ORACLE_I32( "move.count", m.moveCount );
	for ( int i = 0; i < m.moveCount; ++i )
	{
		evBodyId( bm[i].bodyId );
		ORACLE_I32( "move.tag", *(int*)bm[i].userData );
		ORACLE_POS( "move.position", bm[i].transform.p );
		ORACLE_ROT( "move.rotation", bm[i].transform.q );
		ORACLE_BOOL( "move.sleep", bm[i].fellAsleep );
	}
	ORACLE_I32( "joint.count", j.count );
	for ( int i = 0; i < j.count; ++i )
	{
		ORACLE_I32( "joint.index", je[i].jointId.index1 );
		ORACLE_U32( "joint.generation", je[i].jointId.generation );
		ORACLE_BOOL( "joint.world", je[i].jointId.world0 == evWorld.index1 - 1 );
		ORACLE_BOOL( "joint.valid", b2Joint_IsValid( je[i].jointId ) );
		ORACLE_I32( "joint.tag", *(int*)je[i].userData );
	}
	ORACLE_I32( "state.count", count );
	for ( int i = 0; i < count; ++i )
	{
		b2BodyId b = bodies[i];
		evBodyId( b );
		if ( !b2Body_IsValid( b ) )
			continue;
		ORACLE_POS( "state.position", b2Body_GetPosition( b ) );
		ORACLE_ROT( "state.rotation", b2Body_GetRotation( b ) );
		ORACLE_VEC2( "state.velocity", b2Body_GetLinearVelocity( b ) );
		ORACLE_F32( "state.spin", b2Body_GetAngularVelocity( b ) );
		ORACLE_BOOL( "state.awake", b2Body_IsAwake( b ) );
		ORACLE_BOOL( "state.enabled", b2Body_IsEnabled( b ) );
		b2ContactData data[32];
		int n = b2Body_GetContactData( b, data, 32 );
		qsort( data, n, sizeof( *data ), evCD );
		ORACLE_I32( "state.contacts", n );
		for ( int k = 0; k < n; ++k )
		{
			evShape( data[k].shapeIdA );
			evShape( data[k].shapeIdB );
			evContact( data[k].contactId );
			ORACLE_VEC2( "manifold.normal", data[k].manifold.normal );
			ORACLE_F32( "manifold.rollingImpulse", data[k].manifold.rollingImpulse );
			ORACLE_I32( "manifold.count", data[k].manifold.pointCount );
			bool hit = b2Shape_AreHitEventsEnabled( data[k].shapeIdA ) || b2Shape_AreHitEventsEnabled( data[k].shapeIdB );
			ORACLE_BOOL( "manifold.hitEnabled", hit );
			for ( int p = 0; p < data[k].manifold.pointCount; ++p )
			{
				b2ManifoldPoint mp = data[k].manifold.points[p];
				ORACLE_VEC2( "manifold.anchorA", mp.anchorA );
				ORACLE_VEC2( "manifold.anchorB", mp.anchorB );
				ORACLE_F32( "manifold.separation", mp.separation );
				ORACLE_F32( "manifold.normalImpulse", mp.normalImpulse );
				ORACLE_F32( "manifold.tangentImpulse", mp.tangentImpulse );
				ORACLE_F32( "manifold.totalImpulse", mp.totalNormalImpulse );
				if ( hit )
					ORACLE_F32( "manifold.normalVelocity", mp.normalVelocity );
				ORACLE_U32( "manifold.id", mp.id );
				ORACLE_BOOL( "manifold.persisted", mp.persisted );
			}
		}
	}
}
static int evBodyTags[16];
static void evStart( const char* name, b2Vec2 gravity, bool sleep )
{
	ORACLE_OPEN( name );
	evShapeCount = 0;
	evHistoryCount = 0;
	b2WorldDef d = b2DefaultWorldDef();
	d.gravity = gravity;
	d.enableSleep = sleep;
	d.workerCount = 1;
	evWorld = b2CreateWorld( &d );
}
static b2BodyId evMakeBody( int tag, b2BodyType type, b2Pos pos )
{
	b2BodyDef d = b2DefaultBodyDef();
	d.type = type;
	d.position = pos;
	evBodyTags[tag] = tag;
	d.userData = &evBodyTags[tag];
	return b2CreateBody( evWorld, &d );
}
static b2ShapeId evCircle( b2BodyId b, int tag, float radius, b2ShapeDef* d )
{
	b2Circle c = { { 0, 0 }, radius };
	b2ShapeId s = b2CreateCircleShape( b, d, &c );
	evShapes[evShapeCount] = s;
	evTags[evShapeCount++] = tag;
	return s;
}
static void evStep( b2BodyId* b, int n )
{
	b2World_Step( evWorld, 1.0f / 60.0f, 4 );
	evObserve( b, n );
}
static void evFinish( void )
{
	b2DestroyWorld( evWorld );
	ORACLE_CLOSE();
}

static void evLifecycle( bool sensor, int flags )
{
	evStart( ORACLE_KEY( "api/events/Lifecycle-%d-%d", sensor, flags ), (b2Vec2){ 0, 0 }, false );
	b2BodyId anchor = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.isSensor = sensor;
	d.enableSensorEvents = ( flags & 1 ) != 0;
	d.enableContactEvents = ( flags & 1 ) != 0;
	b2ShapeId root = evCircle( anchor, 10, 2, &d );
	d.isSensor = false;
	d.density = 1;
	d.enableSensorEvents = ( flags & 2 ) != 0;
	d.enableContactEvents = ( flags & 2 ) != 0;
	b2BodyId b[2] = { evMakeBody( 2, b2_dynamicBody, (b2Pos){ -1.9, 0 } ), evMakeBody( 3, b2_dynamicBody, (b2Pos){ 1.9, 0 } ) };
	b2ShapeId visitor[2] = { evCircle( b[0], 20, .25f, &d ), evCircle( b[1], 30, .25f, &d ) };
	evStep( b, 2 );
	evStep( b, 2 );
	b2Body_SetTransform( b[0], (b2Pos){ -5, 0 }, b2Rot_identity );
	evStep( b, 2 );
	b2Body_SetTransform( b[0], (b2Pos){ -1.9, 0 }, b2Rot_identity );
	evStep( b, 2 );
	b2Filter f = b2DefaultFilter();
	f.maskBits = 0;
	b2Shape_SetFilter( visitor[0], f );
	evStep( b, 2 );
	f = b2DefaultFilter();
	b2Shape_SetFilter( visitor[0], f );
	evStep( b, 2 );
	b2Body_SetType( b[0], b2_staticBody );
	evStep( b, 2 );
	b2Body_SetType( b[0], b2_dynamicBody );
	evStep( b, 2 );
	b2Body_Disable( b[1] );
	evStep( b, 2 );
	b2Body_Enable( b[1] );
	evStep( b, 2 );
	b2DestroyBody( b[0] );
	evStep( b, 2 );
	b[0] = evMakeBody( 4, b2_dynamicBody, (b2Pos){ -1.9, 0 } );
	visitor[0] = evCircle( b[0], 40, .25f, &d );
	evStep( b, 2 );
	b2DestroyShape( root, true );
	evStep( b, 2 );
	d.isSensor = sensor;
	d.enableSensorEvents = ( flags & 1 ) != 0;
	d.enableContactEvents = ( flags & 1 ) != 0;
	root = evCircle( anchor, 11, 2, &d );
	evStep( b, 2 );
	evStep( b, 2 );
	b2Shape_EnableSensorEvents( root, ( flags & 1 ) == 0 );
	b2Shape_EnableContactEvents( root, ( flags & 1 ) == 0 );
	b2Shape_EnableSensorEvents( visitor[0], ( flags & 2 ) == 0 );
	b2Shape_EnableContactEvents( visitor[0], ( flags & 2 ) == 0 );
	b2Body_SetTransform( b[0], (b2Pos){ -5, 0 }, b2Rot_identity );
	evStep( b, 2 );
	b2Body_SetTransform( b[0], (b2Pos){ -1.9, 0 }, b2Rot_identity );
	evStep( b, 2 );
	evFinish();
}

static void evMovement( void )
{
	evStart( "api/events/Movement", (b2Vec2){ 0, 0 }, true );
	b2BodyId b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ 3, 4 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.density = 1;
	evCircle( b, 20, .5f, &d );
	b2Body_SetLinearDamping( b, 3 );
	b2Body_SetLinearVelocity( b, (b2Vec2){ 2, 1 } );
	int step = 0;
	for ( ; step < 600 && b2Body_IsAwake( b ); ++step )
		evStep( &b, 1 );
	ORACLE_I32( "sleep.steps", step );
	assert( step < 600 );
	evStep( &b, 1 );
	b2Body_SetTransform( b, (b2Pos){ 8, 9 }, b2MakeRot( .3f ) );
	evObserve( &b, 1 );
	evStep( &b, 1 );
	b2Body_SetAwake( b, true );
	evStep( &b, 1 );
	b2Body_SetAwake( b, false );
	b2Body_ApplyForceToCenter( b, (b2Vec2){ 5, 2 }, false );
	evStep( &b, 1 );
	b2Body_ApplyForceToCenter( b, (b2Vec2){ 5, 2 }, true );
	evStep( &b, 1 );
	b2Body_SetAwake( b, false );
	b2Body_EnableSleep( b, false );
	evStep( &b, 1 );
	evFinish();
}

static void evHits( int flags, int threshold, int side )
{
	evStart( ORACLE_KEY( "api/events/Hit-%d-%d-%d", flags, threshold, side ), (b2Vec2){ 0, 0 }, false );
	float t = threshold ? 5 : 2;
	b2World_SetHitEventThreshold( evWorld, t );
	b2World_EnableContinuous( evWorld, false );
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } ), b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ 0, 1.001 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.enableHitEvents = ( flags & 1 ) != 0;
	d.enableContactEvents = true;
	evCircle( a, 10, .5f, &d );
	d.density = 1;
	d.enableHitEvents = ( flags & 2 ) != 0;
	evCircle( b, 20, .5f, &d );
	float speed = side < 0 ? nextafterf( t, 0 ) : side > 0 ? nextafterf( t, INFINITY ) : t;
	ORACLE_F32( "input.speed", speed );
	b2Body_SetLinearVelocity( b, (b2Vec2){ 0, -speed } );
	for ( int i = 0; i < 10; ++i )
		evStep( &b, 1 );
	evFinish();
}

static void evJointThreshold( int axis, int side )
{
	evStart( ORACLE_KEY( "api/events/Joint-%d-%d", axis, side ), (b2Vec2){ 0, 0 }, false );
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } ), b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ 0, 0 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.density = 1;
	evCircle( b, 20, .5f, &d );
	b2WeldJointDef jd = b2DefaultWeldJointDef();
	jd.base.bodyIdA = a;
	jd.base.bodyIdB = b;
	int tag = 71;
	jd.base.userData = &tag;
	float t = side == -2 ? 0 : side == 2 ? FLT_MAX : side < 0 ? nextafterf( 10, 0 ) : side > 0 ? nextafterf( 10, INFINITY ) : 10;
	if ( axis == 0 )
		jd.base.forceThreshold = t;
	else
		jd.base.torqueThreshold = t;
	b2JointId j = b2CreateWeldJoint( evWorld, &jd );
	ORACLE_F32( "threshold", t );
	for ( int i = 0; i < 4; ++i )
	{
		b2Body_ApplyForceToCenter( b, (b2Vec2){ 10, 0 }, true );
		b2Body_ApplyTorque( b, 10, true );
		evStep( &b, 1 );
		ORACLE_VEC2( "joint.force", b2Joint_GetConstraintForce( j ) );
		ORACLE_F32( "joint.torque", b2Joint_GetConstraintTorque( j ) );
	}
	b2Body_SetAwake( b, false );
	evStep( &b, 1 );
	evFinish();
}

// Callback records are an unordered multiset, not invocation-order expectations.
typedef struct evCall
{
	int kind;
	b2ShapeId a, b;
	b2Pos p;
	b2Vec2 n;
	float fa, fb;
	uint64_t ma, mb;
	int points;
} evCall;
static evCall evCalls[4096];
static int evCallCount;
static bool evReject;
static void evRecord( evCall c )
{
	assert( evCallCount < 4096 );
	evCalls[evCallCount++] = c;
}
static bool evFilter( b2ShapeId a, b2ShapeId b, void* ctx )
{
	(void)ctx;
	evRecord( (evCall){ .kind = 1, .a = a, .b = b } );
	return !evReject;
}
static void evPre( b2ShapeId a, b2ShapeId b, b2Manifold* m, void* ctx )
{
	(void)ctx;
	evRecord( (evCall){ .kind = 2, .a = a, .b = b, .n = m->normal, .points = m->pointCount } );
	if ( evReject )
		m->pointCount = 0;
}
static bool evContinuous( b2ShapeId a, b2ShapeId b, b2Pos p, b2Vec2 n, void* ctx )
{
	(void)ctx;
	evRecord( (evCall){ .kind = 3, .a = a, .b = b, .p = p, .n = n } );
	return !evReject;
}
static float evFriction( float a, uint64_t ma, float b, uint64_t mb )
{
	evRecord( (evCall){ .kind = 4, .fa = a, .fb = b, .ma = ma, .mb = mb } );
	return ma == 11 || mb == 11 ? 0 : 1;
}
static float evRestitution( float a, uint64_t ma, float b, uint64_t mb )
{
	evRecord( (evCall){ .kind = 5, .fa = a, .fb = b, .ma = ma, .mb = mb } );
	return ma == 11 || mb == 11 ? 1 : 0;
}
#define EV_ORDER( field )                                                                                                        \
	if ( a->field != b->field )                                                                                                  \
	return a->field < b->field ? -1 : 1
static int evCallCompare( const void* av, const void* bv )
{
	const evCall* a = av;
	const evCall* b = bv;
	EV_ORDER( kind );
	EV_ORDER( a.index1 );
	EV_ORDER( a.generation );
	EV_ORDER( b.index1 );
	EV_ORDER( b.generation );
	EV_ORDER( p.x );
	EV_ORDER( p.y );
	EV_ORDER( n.x );
	EV_ORDER( n.y );
	EV_ORDER( fa );
	EV_ORDER( fb );
	EV_ORDER( ma );
	EV_ORDER( mb );
	EV_ORDER( points );
	return 0;
}
static void evFlushCalls( void )
{
	qsort( evCalls, evCallCount, sizeof( *evCalls ), evCallCompare );
	ORACLE_I32( "callback.count", evCallCount );
	for ( int i = 0; i < evCallCount; ++i )
	{
		evCall c = evCalls[i];
		ORACLE_I32( "callback.kind", c.kind );
		if ( c.kind <= 3 )
		{
			evShape( c.a );
			evShape( c.b );
		}
		ORACLE_POS( "callback.point", c.p );
		ORACLE_VEC2( "callback.normal", c.n );
		ORACLE_F32( "callback.a", c.fa );
		ORACLE_F32( "callback.b", c.fb );
		ORACLE_U64( "callback.materialA", c.ma );
		ORACLE_U64( "callback.materialB", c.mb );
		ORACLE_I32( "callback.points", c.points );
	}
	evCallCount = 0;
}

static void evFilters( int row, int eligible, bool sensor, bool reject )
{
	evStart( ORACLE_KEY( "api/events/Filter-%d-%d-%d-%d", row, eligible, sensor, reject ), (b2Vec2){ 0, 0 }, false );
	evCallCount = 0;
	evReject = reject;
	b2World_SetCustomFilterCallback( evWorld, evFilter, NULL );
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } ), b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ 0, .8 } );
	b2ShapeDef da = b2DefaultShapeDef(), db = b2DefaultShapeDef();
	da.isSensor = sensor;
	da.enableSensorEvents = db.enableSensorEvents = true;
	da.enableContactEvents = db.enableContactEvents = true;
	da.enableCustomFiltering = ( eligible & 1 ) != 0;
	db.enableCustomFiltering = ( eligible & 2 ) != 0;
	da.filter.categoryBits = 1;
	db.filter.categoryBits = 2;
	da.filter.maskBits = 2;
	db.filter.maskBits = 1;
	if ( row == 1 )
		da.filter.maskBits = 0;
	if ( row == 2 )
		db.filter.maskBits = 0;
	if ( row == 3 )
	{
		da.filter.maskBits = db.filter.maskBits = 0;
		da.filter.groupIndex = db.filter.groupIndex = 7;
	}
	if ( row == 4 )
		da.filter.groupIndex = db.filter.groupIndex = -7;
	if ( row == 5 )
		da.filter.maskBits = db.filter.maskBits = 4;
	db.density = 1;
	evCircle( a, 10, .5f, &da );
	evCircle( b, 20, .5f, &db );
	for ( int i = 0; i < 3; ++i )
	{
		evStep( &b, 1 );
		evFlushCalls();
	}
	evFinish();
}

static int evQueryTags[32], evQueryCount;
static bool evQuery( b2ShapeId s, void* ctx )
{
	(void)ctx;
	evQueryTags[evQueryCount++] = s.index1;
	return true;
}
static int evIntCompare( const void* a, const void* b )
{
	return *(const int*)a - *(const int*)b;
}
static void evQueries( void )
{
	evStart( "api/events/QueryFilter", (b2Vec2){ 0, 0 }, false );
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } );
	b2ShapeDef d = b2DefaultShapeDef();
	for ( int i = 0; i < 3; ++i )
	{
		d.filter.categoryBits = 1u << i;
		d.filter.maskBits = i == 2 ? 0 : 3;
		d.filter.groupIndex = i == 1 ? 7 : -7;
		evCircle( a, 10 + i, .5f, &d );
	}
	for ( int i = 0; i < 4; ++i )
	{
		b2QueryFilter q = b2DefaultQueryFilter();
		q.categoryBits = i == 3 ? 4 : 1;
		q.maskBits = i == 0 ? 0 : i == 1 ? 1 : 7;
		evQueryCount = 0;
		b2World_OverlapAABB( evWorld, (b2Pos){ 0, 0 }, (b2AABB){ { -1, -1 }, { 1, 1 } }, q, evQuery, NULL );
		qsort( evQueryTags, evQueryCount, sizeof( int ), evIntCompare );
		ORACLE_I32( "query.count", evQueryCount );
		for ( int k = 0; k < evQueryCount; ++k )
			ORACLE_I32( "query.index", evQueryTags[k] );
	}
	evFinish();
}

static void evPolicies( int kind, int mode, int domain )
{
	evStart( ORACLE_KEY( "api/events/Policy-%d-%d-%d", kind, mode, domain ), (b2Vec2){ 0, -10 }, true );
	evCallCount = 0;
	evReject = mode == 1;
	if ( kind == 0 )
		b2World_SetPreSolveCallback( evWorld, evPre, NULL, NULL );
	if ( kind == 1 )
		b2World_SetPreSolveCallback( evWorld, NULL, evContinuous, NULL );
	if ( kind == 2 && mode != 0 )
		b2World_SetFrictionCallback( evWorld, evFriction );
	if ( kind == 3 && mode != 0 )
		b2World_SetRestitutionCallback( evWorld, evRestitution );
	if ( mode == 2 )
	{
		b2World_SetPreSolveCallback( evWorld, NULL, NULL, NULL );
		b2World_SetFrictionCallback( evWorld, NULL );
		b2World_SetRestitutionCallback( evWorld, NULL );
	}
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } );
	b2BodyId b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ kind == 1 ? -2 : 0, kind == 1 ? 0 : 1.2 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.material.friction = .8f;
	d.material.restitution = .2f;
	d.material.userMaterialId = 11;
	d.enablePreSolveEvents = domain != 3;
	d.enableContactEvents = true;
	d.enableSensorEvents = true;
	d.isSensor = domain == 1;
	if ( kind == 1 )
		evCircle( a, 10, .2f, &d );
	else
	{
		b2Polygon p = b2MakeBox( 10, .5f );
		b2ShapeId s = b2CreatePolygonShape( a, &d, &p );
		evShapes[evShapeCount] = s;
		evTags[evShapeCount++] = 10;
	}
	d.isSensor = false;
	d.density = 1;
	d.material.friction = .2f;
	d.material.restitution = .8f;
	d.material.userMaterialId = 22;
	evCircle( b, 20, .5f, &d );
	if ( kind == 1 )
	{
		b2Body_SetBullet( b, true );
		b2Body_SetGravityScale( b, 0 );
		b2Body_SetLinearVelocity( b, (b2Vec2){ 150, 0 } );
	}
	else
		b2Body_SetLinearVelocity( b, (b2Vec2){ 4, -4 } );
	if ( domain == 2 )
		b2Body_SetAwake( b, false );
	for ( int i = 0; i < ( kind == 1 ? 3 : 30 ); ++i )
	{
		evStep( &b, 1 );
		evFlushCalls();
	}
	evFinish();
}

static void evFastSensor( void )
{
	evStart( "api/events/FastSensor", (b2Vec2){ 0, 0 }, false );
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } ), b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ -3, 0 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.isSensor = true;
	d.enableSensorEvents = true;
	evCircle( a, 10, 1, &d );
	d.isSensor = false;
	d.density = 1;
	evCircle( b, 20, .1f, &d );
	b2Body_SetBullet( b, true );
	b2Body_SetLinearVelocity( b, (b2Vec2){ 20, 0 } );
	for ( int i = 0; i < 24; ++i )
		evStep( &b, 1 );
	evFinish();
}
static void evSleepingDestroy( void )
{
	evStart( "api/events/SleepingDestroy", (b2Vec2){ 0, 0 }, true );
	b2BodyId a = evMakeBody( 1, b2_staticBody, (b2Pos){ 0, 0 } );
	b2BodyId b = evMakeBody( 2, b2_dynamicBody, (b2Pos){ 0, .4 } );
	b2ShapeDef d = b2DefaultShapeDef();
	d.isSensor = true;
	d.enableSensorEvents = true;
	evCircle( a, 10, 1, &d );
	d.isSensor = false;
	d.density = 1;
	b2ShapeId visitor = evCircle( b, 20, .25f, &d );
	evStep( &b, 1 );
	b2Body_SetAwake( b, false );
	evStep( &b, 1 );
	b2DestroyShape( visitor, true );
	evStep( &b, 1 );
	visitor = evCircle( b, 21, .25f, &d );
	evStep( &b, 1 );
	b2Body_SetAwake( b, false );
	evStep( &b, 1 );
	b2DestroyBody( b );
	evStep( &b, 1 );
	b = evMakeBody( 3, b2_dynamicBody, (b2Pos){ 0, .4 } );
	evCircle( b, 30, .25f, &d );
	evStep( &b, 1 );
	b2Body_SetAwake( b, false );
	evStep( &b, 1 );
	b2DestroyBody( a );
	evStep( &b, 1 );
	evStep( &b, 1 );
	evFinish();
}
int main( void )
{
	for ( int s = 0; s < 2; ++s )
		for ( int f = 0; f < 4; ++f )
			evLifecycle( s, f );
	evMovement();
	evFastSensor();
	evSleepingDestroy();
	for ( int f = 0; f < 4; ++f )
		for ( int t = 0; t < 2; ++t )
			for ( int side = -1; side <= 1; ++side )
				evHits( f, t, side );
	for ( int a = 0; a < 2; ++a )
		for ( int side = -2; side <= 2; ++side )
			evJointThreshold( a, side );
	for ( int row = 0; row < 6; ++row )
		for ( int e = 0; e < 4; ++e )
			for ( int s = 0; s < 2; ++s )
				for ( int r = 0; r < 2; ++r )
					evFilters( row, e, s, r );
	evQueries();
	for ( int k = 0; k < 4; ++k )
		for ( int m = 0; m < 3; ++m )
			for ( int d = 0; d < 4; ++d )
				evPolicies( k, m, d );
	return 0;
}
