// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include "scenes/driver.h"
#include "scenes/human.c"
#include "scenes/utils.c"

int main( void )
{
	for ( int workers = 1; workers <= 4; workers += 3 )
	{
		g_randomSeed = 12345;
		char fixture[64];
		snprintf( fixture, sizeof( fixture ), "scenes/Human_workers%d", workers );
		ORACLE_OPEN( fixture );
		b2WorldDef def = b2DefaultWorldDef();
		def.workerCount = workers;
		b2WorldId world = b2CreateWorld( &def );
		Human human = { 0 };
		CreateHuman( &human, world, (b2Pos){ 0, 5 }, 1.0f, 0.05f, 5.0f, 0.5f, 1, NULL, true );
		scene_dump( world, -1 );
		for ( int step = 0; step < 300; ++step )
		{
			if ( step == 0 )
				Human_SetVelocity( &human, (b2Vec2){ 1, 0 } );
			if ( step == 20 )
				Human_ApplyRandomAngularImpulse( &human, 0.5f );
			if ( step == 40 )
				Human_SetJointFrictionTorque( &human, 0 );
			if ( step == 60 )
				Human_SetJointFrictionTorque( &human, 0.1f );
			if ( step == 80 )
				Human_SetJointSpringHertz( &human, 0 );
			if ( step == 100 )
				Human_SetJointSpringHertz( &human, 2.0f );
			if ( step == 120 )
				Human_SetJointDampingRatio( &human, 0.75f );
			if ( step == 140 )
				Human_EnableSensorEvents( &human, true );
			if ( step == 160 )
				Human_SetScale( &human, 1.5f );
			if ( step == 180 )
				Human_EnableSensorEvents( &human, false );
			if ( step == 140 || step == 180 )
			{
				b2ShapeId shape = b2_nullShapeId;
				int count = b2Body_GetShapes( human.bones[bone_torso].bodyId, &shape, 1 );
				ORACLE_I32( scene_key( step, "torsoShapes" ), count );
				bool enabled = b2Shape_AreSensorEventsEnabled( shape );
				ORACLE_BOOL( scene_key( step, "sensor" ), enabled );
				if ( count != 1 || enabled != ( step == 140 ) )
					abort();
			}
			if ( step == 0 || ( step >= 20 && step <= 180 && step % 20 == 0 ) )
				scene_state( world, 2000 + step );
			b2World_Step( world, 1.0f / 60.0f, 4 );
			scene_dump( world, step );
		}
		DestroyHuman( &human );
		scene_dump( world, 300 );
		b2DestroyWorld( world );
		scene_flush();
		ORACLE_CLOSE();
	}
	return 0;
}
