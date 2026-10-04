// SPDX-License-Identifier: MIT
// Four-byte operations: opcode, slot, x, y. At most 64 operations; a final
// short operation is zero-padded. Coordinates/velocities are signed bytes / 16.
// The matching executor lives in internal/oracle/sequence.
#include "oracle.h"
#include "replay.h"

#include "box2d/box2d.h"
#include "box2d/math_functions.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

enum
{
	slot_count = 4,
	joint_count = 2,
	input_limit = 256
};
static b2WorldId world;
static b2BodyId bodies[slot_count], saved_bodies[slot_count];
static b2ShapeId shapes[slot_count], saved_shapes[slot_count];
static b2JointId joints[joint_count], saved_joints[joint_count];
static b2BodyId old_bodies[64];
static b2ShapeId old_shapes[64];
static b2JointId old_joints[64];
static int nb, ns, nj;
static uint8_t* image;
static int image_size;
static unsigned epoch;

// Only the process-local world slot is mapped. Allocation index and the full
// object generation are preserved, including event IDs for destroyed objects.
static uint64_t local_id( uint64_t id )
{
	return id & ~UINT64_C( 0xffff0000 );
}
static float scalar( uint8_t x )
{
	return (float)( x < 128 ? x : (int)x - 256 ) / 16.0f;
}
static b2Pos position( uint8_t x, uint8_t y )
{
	return (b2Pos){ scalar( x ), scalar( y ) };
}
static b2WorldId new_world( void )
{
	b2WorldDef d = b2DefaultWorldDef();
	d.gravity = (b2Vec2){ 0.0f, -10.0f };
	d.enableSleep = false;
	d.workerCount = 0;
	return b2CreateWorld( &d );
}
static void contact_id( b2ContactId id )
{
	ORACLE_U32( "contact.index", (uint32_t)id.index1 );
	ORACLE_U32( "contact.generation", id.generation );
	ORACLE_BOOL( "contact.valid", b2Contact_IsValid( id ) );
}
static void pair( b2ShapeId a, b2ShapeId b, b2ContactId c )
{
	ORACLE_U64( "pair.a", local_id( b2StoreShapeId( a ) ) );
	ORACLE_U64( "pair.b", local_id( b2StoreShapeId( b ) ) );
	ORACLE_BOOL( "pair.a.valid", b2Shape_IsValid( a ) );
	ORACLE_BOOL( "pair.b.valid", b2Shape_IsValid( b ) );
	contact_id( c );
}
static void events( b2WorldId w )
{
	b2ContactEvents e = b2World_GetContactEvents( w );
	ORACLE_I32( "begin.count", e.beginCount );
	for ( int i = 0; i < e.beginCount; ++i )
		pair( e.beginEvents[i].shapeIdA, e.beginEvents[i].shapeIdB, e.beginEvents[i].contactId );
	ORACLE_I32( "end.count", e.endCount );
	for ( int i = 0; i < e.endCount; ++i )
		pair( e.endEvents[i].shapeIdA, e.endEvents[i].shapeIdB, e.endEvents[i].contactId );
	ORACLE_I32( "hit.count", e.hitCount );
	for ( int i = 0; i < e.hitCount; ++i )
	{
		b2ContactHitEvent h = e.hitEvents[i];
		pair( h.shapeIdA, h.shapeIdB, h.contactId );
		ORACLE_POS( "hit.point", h.point );
		ORACLE_VEC2( "hit.normal", h.normal );
		ORACLE_F32( "hit.speed", h.approachSpeed );
	}
	b2BodyEvents be = b2World_GetBodyEvents( w );
	ORACLE_I32( "move.count", be.moveCount );
	for ( int i = 0; i < be.moveCount; ++i )
	{
		b2BodyMoveEvent b = be.moveEvents[i];
		ORACLE_U64( "move.body", local_id( b2StoreBodyId( b.bodyId ) ) );
		ORACLE_POS( "move.position", b.transform.p );
		ORACLE_ROT( "move.rotation", b.transform.q );
		ORACLE_BOOL( "move.sleep", b.fellAsleep );
	}
}
static void body_state( b2BodyId b )
{
	ORACLE_U64( "body.id", local_id( b2StoreBodyId( b ) ) );
	bool valid = b2Body_IsValid( b );
	ORACLE_BOOL( "body.valid", valid );
	if ( !valid )
		return;
	ORACLE_POS( "body.position", b2Body_GetPosition( b ) );
	ORACLE_ROT( "body.rotation", b2Body_GetRotation( b ) );
	ORACLE_VEC2( "body.velocity", b2Body_GetLinearVelocity( b ) );
	ORACLE_F32( "body.angular", b2Body_GetAngularVelocity( b ) );
	ORACLE_I32( "body.type", b2Body_GetType( b ) );
	ORACLE_BOOL( "body.enabled", b2Body_IsEnabled( b ) );
	ORACLE_BOOL( "body.awake", b2Body_IsAwake( b ) );
	ORACLE_F32( "body.mass", b2Body_GetMass( b ) );
	b2ContactData contacts[16];
	int n = b2Body_GetContactData( b, contacts, 16 );
	ORACLE_I32( "body.contacts", n );
	for ( int i = 0; i < n; ++i )
	{
		b2ContactData c = contacts[i];
		pair( c.shapeIdA, c.shapeIdB, c.contactId );
		ORACLE_VEC2( "manifold.normal", c.manifold.normal );
		ORACLE_F32( "manifold.rolling", c.manifold.rollingImpulse );
		ORACLE_I32( "manifold.count", c.manifold.pointCount );
		for ( int j = 0; j < c.manifold.pointCount; ++j )
		{
			b2ManifoldPoint p = c.manifold.points[j];
			ORACLE_VEC2( "point.a", p.anchorA );
			ORACLE_VEC2( "point.b", p.anchorB );
			ORACLE_F32( "point.separation", p.separation );
			ORACLE_F32( "point.normal", p.normalImpulse );
			ORACLE_F32( "point.tangent", p.tangentImpulse );
			ORACLE_F32( "point.total", p.totalNormalImpulse );
			ORACLE_F32( "point.velocity", p.normalVelocity );
			ORACLE_U32( "point.id", p.id );
			ORACLE_BOOL( "point.persisted", p.persisted );
		}
	}
}
static void world_state( b2WorldId w )
{
	b2Counters c = b2World_GetCounters( w );
	ORACLE_I32( "world.bodies", c.bodyCount );
	ORACLE_I32( "world.shapes", c.shapeCount );
	ORACLE_I32( "world.joints", c.jointCount );
	ORACLE_I32( "world.contacts", c.contactCount );
	ORACLE_I32( "world.islands", c.islandCount );
	events( w );
}
static void observe( void )
{
	ORACLE_U32( "world.epoch", epoch );
	world_state( world );
	for ( int i = 0; i < slot_count; ++i )
	{
		ORACLE_I32( "slot", i );
		body_state( bodies[i] );
		ORACLE_U64( "shape.id", local_id( b2StoreShapeId( shapes[i] ) ) );
		bool valid = b2Shape_IsValid( shapes[i] );
		ORACLE_BOOL( "shape.valid", valid );
		if ( valid )
		{
			b2Filter f = b2Shape_GetFilter( shapes[i] );
			ORACLE_U64( "shape.category", f.categoryBits );
			ORACLE_U64( "shape.mask", f.maskBits );
			ORACLE_I32( "shape.group", f.groupIndex );
		}
	}
	for ( int i = 0; i < joint_count; ++i )
	{
		ORACLE_U64( "joint.id", local_id( b2StoreJointId( joints[i] ) ) );
		bool valid = b2Joint_IsValid( joints[i] );
		ORACLE_BOOL( "joint.valid", valid );
		if ( valid )
		{
			ORACLE_U64( "joint.a", local_id( b2StoreBodyId( b2Joint_GetBodyA( joints[i] ) ) ) );
			ORACLE_U64( "joint.b", local_id( b2StoreBodyId( b2Joint_GetBodyB( joints[i] ) ) ) );
			ORACLE_F32( "joint.length", b2DistanceJoint_GetLength( joints[i] ) );
			ORACLE_VEC2( "joint.force", b2Joint_GetConstraintForce( joints[i] ) );
		}
	}
	ORACLE_I32( "retained.bodies", nb );
	for ( int i = 0; i < nb; ++i )
	{
		ORACLE_U64( "retained.body", local_id( b2StoreBodyId( old_bodies[i] ) ) );
		ORACLE_BOOL( "retained.body.valid", b2Body_IsValid( old_bodies[i] ) );
	}
	ORACLE_I32( "retained.shapes", ns );
	for ( int i = 0; i < ns; ++i )
	{
		ORACLE_U64( "retained.shape", local_id( b2StoreShapeId( old_shapes[i] ) ) );
		ORACLE_BOOL( "retained.shape.valid", b2Shape_IsValid( old_shapes[i] ) );
	}
	ORACLE_I32( "retained.joints", nj );
	for ( int i = 0; i < nj; ++i )
	{
		ORACLE_U64( "retained.joint", local_id( b2StoreJointId( old_joints[i] ) ) );
		ORACLE_BOOL( "retained.joint.valid", b2Joint_IsValid( old_joints[i] ) );
	}
	ORACLE_U64( "world.hash", b2World_GetStateHash( world ) );
}
struct hits
{
	uint64_t ids[slot_count + 1];
	int n;
	int limit;
};
static bool overlap( b2ShapeId id, void* context )
{
	struct hits* h = context;
	if ( h->n >= slot_count + 1 )
		abort();
	h->ids[h->n++] = local_id( b2StoreShapeId( id ) );
	return h->limit == 0 || h->n < h->limit;
}
static int cmp_id( const void* a, const void* b )
{
	uint64_t x = *(const uint64_t*)a, y = *(const uint64_t*)b;
	return ( x > y ) - ( x < y );
}
static void query( b2WorldId w, uint8_t x, uint8_t y )
{
	b2Pos p = position( x, y );
	b2AABB a = { { -2.0f, -2.0f }, { 2.0f, 2.0f } };
	b2QueryFilter f = b2DefaultQueryFilter();
	struct hits h = { 0 };
	b2World_OverlapAABB( w, p, a, f, overlap, &h );
	qsort( h.ids, (size_t)h.n, sizeof( h.ids[0] ), cmp_id );
	ORACLE_I32( "query.count", h.n );
	for ( int i = 0; i < h.n; ++i )
		ORACLE_U64( "query.shape", h.ids[i] );
	h.n = 0;
	h.limit = 1;
	b2World_OverlapAABB( w, p, a, f, overlap, &h );
	ORACLE_I32( "query.cancel.count", h.n );
	// The callback order is not promised. Compare the exhaustive multiset above
	// and cancellation multiplicity, not an arbitrary first hit identity.
	b2RayResult r = b2World_CastRayClosest( w, (b2Pos){ -10.0f, scalar( y ) }, (b2Vec2){ 20.0f, 0.0f }, f );
	ORACLE_BOOL( "ray.hit", r.hit );
	if ( r.hit )
	{
		ORACLE_U64( "ray.shape", local_id( b2StoreShapeId( r.shapeId ) ) );
		ORACLE_POS( "ray.point", r.point );
		ORACLE_VEC2( "ray.normal", r.normal );
		ORACLE_F32( "ray.fraction", r.fraction );
	}
}
static void replay_state( b2Replay* p )
{
	ORACLE_I32( "replay.frame", b2Replay_GetFrame( p ) );
	ORACLE_BOOL( "replay.end", b2Replay_IsAtEnd( p ) );
	ORACLE_BOOL( "replay.diverged", b2Replay_HasDiverged( p ) );
	ORACLE_I32( "replay.diverge.frame", b2Replay_GetDivergeFrame( p ) );
	world_state( b2Replay_GetWorldId( p ) );
	int n = b2Replay_GetBodyCount( p );
	ORACLE_I32( "replay.bodies", n );
	for ( int i = 0; i < n; ++i )
		body_state( b2Replay_GetBodyId( p, i ) );
	n = b2Replay_GetFrameQueryCount( p );
	ORACLE_I32( "replay.queries", n );
	for ( int i = 0; i < n; ++i )
	{
		b2ReplayQueryInfo q = b2Replay_GetFrameQuery( p, i );
		ORACLE_I32( "replay.query.type", q.type );
		ORACLE_I32( "replay.query.hits", q.hitCount );
		for ( int j = 0; j < q.hitCount; ++j )
		{
			if ( q.type == b2_replayQueryOverlapAABB || q.type == b2_replayQueryOverlapShape )
			{
				// Native overlap hits initialize only identity and callback control.
				// The generic getter also reads indeterminate cast-only payloads.
				b2ShapeId id = p->frameHits[p->frameQueries[i].hitStart + j].id;
				ORACLE_U64( "replay.query.shape", local_id( b2StoreShapeId( id ) ) );
			}
			else
			{
				b2ReplayQueryHit h = b2Replay_GetFrameQueryHit( p, i, j );
				ORACLE_U64( "replay.query.shape", local_id( b2StoreShapeId( h.shape ) ) );
				ORACLE_POS( "replay.query.point", h.point );
				ORACLE_VEC2( "replay.query.normal", h.normal );
				ORACLE_F32( "replay.query.fraction", h.fraction );
			}
		}
	}
	ORACLE_U64( "world.hash", b2World_GetStateHash( b2Replay_GetWorldId( p ) ) );
}
static void recording( uint8_t x, uint8_t y )
{
	b2Recording* r = b2CreateRecording( 0 );
	if ( !r )
		abort();
	b2World_StartRecording( world, r );
	b2BodyDef bd = b2DefaultBodyDef();
	bd.type = b2_dynamicBody;
	bd.position = position( x, y );
	b2BodyId transient = b2CreateBody( world, &bd );
	b2ShapeDef sd = b2DefaultShapeDef();
	sd.enableContactEvents = true;
	sd.enableHitEvents = true;
	b2Circle circle = { { 0.0f, 0.0f }, 0.375f };
	b2ShapeId first = b2CreateCircleShape( transient, &sd, &circle ), current = first;
	b2JointId transient_joint = b2_nullJointId;
	if ( b2Body_IsValid( bodies[0] ) )
	{
		b2DistanceJointDef jd = b2DefaultDistanceJointDef();
		jd.base.bodyIdA = bodies[0];
		jd.base.bodyIdB = transient;
		jd.base.collideConnected = true;
		transient_joint = b2CreateDistanceJoint( world, &jd );
	}
	for ( int i = 0; i < 3; ++i )
	{
		if ( i == 1 )
		{
			b2DestroyShape( current, true );
			current = b2CreateCircleShape( transient, &sd, &circle );
		}
		if ( i == 2 )
			b2DestroyBody( transient );
		if ( b2Body_IsValid( bodies[i] ) )
			b2Body_SetLinearVelocity( bodies[i], (b2Vec2){ scalar( x ), scalar( y ) } );
		query( world, x, y );
		b2World_Step( world, 1.0f / 60.0f, 4 );
		ORACLE_I32( "record.step", i );
		world_state( world );
		body_state( transient );
		ORACLE_U64( "record.first.shape", local_id( b2StoreShapeId( first ) ) );
		ORACLE_BOOL( "record.first.shape.valid", b2Shape_IsValid( first ) );
		ORACLE_U64( "record.current.shape", local_id( b2StoreShapeId( current ) ) );
		ORACLE_BOOL( "record.current.shape.valid", b2Shape_IsValid( current ) );
		ORACLE_U64( "record.joint", local_id( b2StoreJointId( transient_joint ) ) );
		ORACLE_BOOL( "record.joint.valid", b2Joint_IsValid( transient_joint ) );
		for ( int j = 0; j < slot_count; ++j )
			body_state( bodies[j] );
		ORACLE_U64( "world.hash", b2World_GetStateHash( world ) );
	}
	b2World_StopRecording( world );
	b2Replay* p = b2CreateReplay( b2Recording_GetData( r ), b2Recording_GetSize( r ), 0 );
	ORACLE_BOOL( "replay.created", p != NULL );
	if ( !p )
		abort();
	b2DestroyRecording( r );
	b2ReplayInfo info = b2Replay_GetInfo( p );
	ORACLE_I32( "replay.frames", info.frameCount );
	ORACLE_I32( "replay.workers", info.workerCount );
	ORACLE_F32( "replay.dt", info.timeStep );
	ORACLE_I32( "replay.substeps", info.subStepCount );
	replay_state( p );
	const int targets[] = { 3, 1, 2, 0, 3 };
	for ( int i = 0; i < 5; ++i )
	{
		b2Replay_SeekFrame( p, targets[i] );
		replay_state( p );
		ORACLE_BOOL( "replay.step", b2Replay_StepFrame( p ) );
		replay_state( p );
	}
	b2Replay_SeekFrame( p, 1 );
	b2WorldId rw = b2Replay_GetWorldId( p );
	int size = b2World_GetSnapshot( rw, NULL, 0 );
	uint8_t* snap = malloc( (size_t)size );
	if ( !snap || size <= 0 )
		abort();
	if ( b2World_GetSnapshot( rw, snap, size ) != size )
		abort();
	for ( int i = 0; i < 2; ++i )
	{
		ORACLE_BOOL( "replay.restore", b2World_Restore( rw, snap, size ) );
		query( rw, x, y );
		b2World_Step( rw, 1.0f / 60.0f, 4 );
		ORACLE_I32( "replay.restep", i );
		world_state( rw );
		for ( int j = 0; j < b2Replay_GetBodyCount( p ); ++j )
			body_state( b2Replay_GetBodyId( p, j ) );
		ORACLE_U64( "world.hash", b2World_GetStateHash( rw ) );
	}
	// Continue a replay-owned native world with a new mutation after re-stepping.
	for ( int i = 0; i < b2Replay_GetBodyCount( p ); ++i )
	{
		b2BodyId b = b2Replay_GetBodyId( p, i );
		if ( b2Body_IsValid( b ) )
			b2Body_ApplyForceToCenter( b, (b2Vec2){ scalar( y ), scalar( x ) }, true );
	}
	b2World_Step( rw, 1.0f / 60.0f, 4 );
	world_state( rw );
	for ( int j = 0; j < b2Replay_GetBodyCount( p ); ++j )
		body_state( b2Replay_GetBodyId( p, j ) );
	ORACLE_U64( "world.hash", b2World_GetStateHash( rw ) );
	free( snap );
	b2DestroyReplay( p );
}
static int run_sequence( const char* name, const uint8_t* input, size_t size )
{
	if ( size > input_limit )
		return 2;
	memset( bodies, 0, sizeof( bodies ) );
	memset( shapes, 0, sizeof( shapes ) );
	memset( joints, 0, sizeof( joints ) );
	memset( saved_bodies, 0, sizeof( saved_bodies ) );
	memset( saved_shapes, 0, sizeof( saved_shapes ) );
	memset( saved_joints, 0, sizeof( saved_joints ) );
	nb = ns = nj = 0;
	epoch = 0;
	image_size = 0;
	image = NULL;
	world = new_world();
	ORACLE_OPEN( name );
	ORACLE_BYTES( "input", input, size );
	observe();
	for ( size_t offset = 0; offset < size; offset += 4 )
	{
		uint8_t op[4] = { 0 };
		size_t n = size - offset < 4 ? size - offset : 4;
		memcpy( op, input + offset, n );
		unsigned code = op[0] % 21, s = op[1] % slot_count, j = op[1] % joint_count;
		b2BodyId b = bodies[s];
		b2ShapeId sh = shapes[s];
		bool bv = b2Body_IsValid( b ), sv = b2Shape_IsValid( sh ), acted = false;
		ORACLE_U32( "operation.index", (uint32_t)( offset / 4 ) );
		ORACLE_U32( "operation.code", code );
		switch ( code )
		{
			case 0:
				if ( !bv )
				{
					b2BodyDef d = b2DefaultBodyDef();
					d.type = b2_dynamicBody;
					d.position = position( op[2], op[3] );
					bodies[s] = b2CreateBody( world, &d );
					old_bodies[nb++] = bodies[s];
					acted = true;
				}
				break;
			case 1:
				if ( bv )
				{
					b2DestroyBody( b );
					acted = true;
				}
				break;
			case 2:
				if ( bv && !sv )
				{
					b2ShapeDef d = b2DefaultShapeDef();
					d.enableContactEvents = true;
					d.enableHitEvents = true;
					b2Circle c = { { 0.0f, 0.0f }, 0.25f + ( op[2] % 5 ) * 0.125f };
					shapes[s] = b2CreateCircleShape( b, &d, &c );
					old_shapes[ns++] = shapes[s];
					acted = true;
				}
				break;
			case 3:
				if ( sv )
				{
					b2DestroyShape( sh, true );
					acted = true;
				}
				break;
			case 4:
				if ( bv )
				{
					b2Body_SetTransform( b, position( op[2], op[3] ), b2MakeRot( ( op[2] % 9 ) * 0.125f ) );
					acted = true;
				}
				break;
			case 5:
				if ( bv )
				{
					b2Body_SetLinearVelocity( b, (b2Vec2){ scalar( op[2] ), scalar( op[3] ) } );
					acted = true;
				}
				break;
			case 6:
				if ( bv )
				{
					b2Body_SetAngularVelocity( b, scalar( op[2] ) );
					acted = true;
				}
				break;
			case 7:
				if ( bv )
				{
					b2Body_ApplyForceToCenter( b, (b2Vec2){ scalar( op[2] ), scalar( op[3] ) }, true );
					acted = true;
				}
				break;
			case 8:
				if ( sv )
				{
					b2Filter f = b2DefaultFilter();
					f.categoryBits = UINT64_C( 1 ) << ( op[2] % 4 );
					f.maskBits = op[3] & 15;
					b2Shape_SetFilter( sh, f );
					acted = true;
				}
				break;
			case 9:
				if ( bv )
				{
					b2Body_SetType( b, (b2BodyType)( op[2] % 3 ) );
					acted = true;
				}
				break;
			case 10:
				if ( bv )
				{
					b2Body_Disable( b );
					acted = true;
				}
				break;
			case 11:
				if ( bv )
				{
					b2Body_Enable( b );
					acted = true;
				}
				break;
			case 12:
				b2World_Step( world, ( op[2] & 1 ) ? 0.0f : 1.0f / 60.0f, 1 + op[3] % 4 );
				acted = true;
				break;
			case 13:
				query( world, op[2], op[3] );
				acted = true;
				break;
			case 14:
			{
				unsigned other = op[2] % slot_count;
				if ( !b2Joint_IsValid( joints[j] ) && bv && other != s && b2Body_IsValid( bodies[other] ) &&
					 ( b2Body_GetType( b ) == b2_dynamicBody || b2Body_GetType( bodies[other] ) == b2_dynamicBody ) )
				{
					b2DistanceJointDef d = b2DefaultDistanceJointDef();
					d.base.bodyIdA = b;
					d.base.bodyIdB = bodies[other];
					d.length = 0.5f + ( op[3] % 8 ) * 0.25f;
					d.base.collideConnected = true;
					joints[j] = b2CreateDistanceJoint( world, &d );
					old_joints[nj++] = joints[j];
					acted = true;
				}
				break;
			}
			case 15:
				if ( b2Joint_IsValid( joints[j] ) )
				{
					b2DestroyJoint( joints[j] );
					acted = true;
				}
				break;
			case 16:
				free( image );
				image_size = b2World_GetSnapshot( world, NULL, 0 );
				image = malloc( (size_t)image_size );
				if ( !image || image_size <= 0 || b2World_GetSnapshot( world, image, image_size ) != image_size )
					abort();
				memcpy( saved_bodies, bodies, sizeof( bodies ) );
				memcpy( saved_shapes, shapes, sizeof( shapes ) );
				memcpy( saved_joints, joints, sizeof( joints ) );
				acted = true;
				break;
			case 17:
				if ( image )
				{
					bool ok = b2World_Restore( world, image, image_size );
					ORACLE_BOOL( "snapshot.restore", ok );
					if ( !ok )
						abort();
					memcpy( bodies, saved_bodies, sizeof( bodies ) );
					memcpy( shapes, saved_shapes, sizeof( shapes ) );
					memcpy( joints, saved_joints, sizeof( joints ) );
					acted = true;
				}
				break;
			case 18:
				if ( bv )
				{
					b2Body_ApplyLinearImpulseToCenter( b, (b2Vec2){ scalar( op[2] ), scalar( op[3] ) }, true );
					acted = true;
				}
				break;
			case 19:
				recording( op[2], op[3] );
				acted = true;
				break;
			case 20:
			{
				b2WorldId old = world;
				b2DestroyWorld( world );
				ORACLE_BOOL( "world.destroyed", !b2World_IsValid( old ) );
				world = new_world();
				ORACLE_BOOL( "world.old.valid", b2World_IsValid( old ) );
				++epoch;
				free( image );
				image = NULL;
				image_size = 0;
				memset( bodies, 0, sizeof( bodies ) );
				memset( shapes, 0, sizeof( shapes ) );
				memset( joints, 0, sizeof( joints ) );
				nb = ns = nj = 0;
				acted = true;
				break;
			}
		}
		ORACLE_BOOL( "operation.acted", acted );
		observe();
	}
	ORACLE_U32( "operation.count", (uint32_t)( ( size + 3 ) / 4 ) );
	b2DestroyWorld( world );
	ORACLE_BOOL( "world.final.invalid", !b2World_IsValid( world ) );
	ORACLE_CLOSE();
	free( image );
	image = NULL;
	return 0;
}

#ifndef BOX2D_DIFFERENTIAL_NO_MAIN
int main( void )
{
	uint8_t input[input_limit];
	size_t size = fread( input, 1, sizeof( input ), stdin );
	if ( ferror( stdin ) || ( size == sizeof( input ) && fgetc( stdin ) != EOF ) )
		return 2;
	return run_sequence( "differential/sequence", input, size );
}
#endif
