// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "core.h"
#include "oracle.h"
#include "test_macros.h"

// b2Semaphore tests

static int SemaphoreCreateDestroyTest( void )
{
	ORACLE_OPEN( "thread/SemaphoreCreateDestroyTest" );
	b2Semaphore* s = b2CreateSemaphore( 0 );
	ENSURE( s != NULL );
	ORACLE_BOOL( "created", s != NULL );
	b2SignalSemaphore( s );
	b2WaitSemaphore( s );
	ORACLE_I32( "completedWaits", 1 );
	b2DestroySemaphore( s );
	s = b2CreateSemaphore( 1 );
	ENSURE( s != NULL );
	b2WaitSemaphore( s );
	b2DestroySemaphore( s );
	ORACLE_I32( "recreatedCompletedWaits", 1 );
	ORACLE_CLOSE();
	return 0;
}

typedef struct SemData
{
	b2Semaphore* sem;
	int value;
} SemData;

static void SemWorker( void* context )
{
	SemData* data = (SemData*)context;
	data->value = 99;
	b2SignalSemaphore( data->sem );
}

static int SemaphoreSignalWaitTest( void )
{
	ORACLE_OPEN( "thread/SemaphoreSignalWaitTest" );
	SemData data = { 0 };
	data.sem = b2CreateSemaphore( 0 );
	data.value = 0;

	b2Thread* thread = b2CreateThread( SemWorker, &data, "sem test" );
	b2WaitSemaphore( data.sem );

	int afterWait = data.value;

	b2JoinThread( thread );
	b2DestroySemaphore( data.sem );
	ENSURE( data.value == 99 );
	ENSURE( afterWait == 99 );
	ORACLE_I32( "afterWait", afterWait );
	ORACLE_I32( "afterJoin", data.value );
	ORACLE_CLOSE();
	return 0;
}

static int SemaphoreInitialCountTest( void )
{
	ORACLE_OPEN( "thread/SemaphoreInitialCountTest" );
	// Semaphore initialized to 3 should allow 3 waits without blocking
	b2Semaphore* s = b2CreateSemaphore( 3 );
	int completedWaits = 0;

	// These should not block since count starts at 3
	b2WaitSemaphore( s );
	ORACLE_I32( "completedWaits", ++completedWaits );
	b2WaitSemaphore( s );
	ORACLE_I32( "completedWaits", ++completedWaits );
	b2WaitSemaphore( s );
	ORACLE_I32( "completedWaits", ++completedWaits );

	// Signal once so the next wait doesn't block
	b2SignalSemaphore( s );
	b2WaitSemaphore( s );
	ORACLE_I32( "completedWaits", ++completedWaits );

	b2DestroySemaphore( s );
	ORACLE_CLOSE();
	return 0;
}

// b2Thread tests

static int ThreadCreateJoinTest( void )
{
	ORACLE_OPEN( "thread/ThreadCreateJoinTest" );
	SemData data = { 0 };
	data.sem = b2CreateSemaphore( 0 );
	data.value = 0;

	b2Thread* thread = b2CreateThread( SemWorker, &data, "join test" );
	b2JoinThread( thread );

	int afterJoin = data.value;
	b2WaitSemaphore( data.sem );

	b2DestroySemaphore( data.sem );
	ENSURE( data.value == 99 );
	ORACLE_I32( "afterJoin", afterJoin );
	ORACLE_I32( "completedWaits", 1 );
	ORACLE_CLOSE();
	return 0;
}

typedef struct SumData
{
	b2Mutex* mutex;
	int sum;
	int completed;
} SumData;

static void SumWorker( void* context )
{
	SumData* data = (SumData*)context;
	for ( int i = 0; i < 1000; ++i )
	{
		b2LockMutex( data->mutex );
		data->sum += 1;
		b2UnlockMutex( data->mutex );
	}
	b2LockMutex( data->mutex );
	data->completed += 1;
	b2UnlockMutex( data->mutex );
}

static int ThreadMultipleTest( void )
{
	ORACLE_OPEN( "thread/ThreadMultipleTest" );
	SumData data = { 0 };
	data.mutex = b2CreateMutex();
	data.sum = 0;

	enum
	{
		THREAD_COUNT = 4
	};
	b2Thread* threads[THREAD_COUNT];
	for ( int i = 0; i < THREAD_COUNT; ++i )
	{
		char name[16];
		snprintf( name, sizeof( name ), "sum test %d", i );
		threads[i] = b2CreateThread( SumWorker, &data, name );
	}
	int joined = 0;

	for ( int i = 0; i < THREAD_COUNT; ++i )
	{
		b2JoinThread( threads[i] );
		joined += 1;
	}

	int sum = data.sum;
	int completed = data.completed;

	b2DestroyMutex( data.mutex );
	ENSURE( data.sum == THREAD_COUNT * 1000 );
	ORACLE_I32( "sum", sum );
	ORACLE_I32( "completedWorkers", completed );
	ORACLE_I32( "joinedWorkers", joined );
	ORACLE_CLOSE();
	return 0;
}

int ThreadTest( void )
{
	RUN_SUBTEST( SemaphoreCreateDestroyTest );
	RUN_SUBTEST( SemaphoreSignalWaitTest );
	RUN_SUBTEST( SemaphoreInitialCountTest );
	RUN_SUBTEST( ThreadCreateJoinTest );
	RUN_SUBTEST( ThreadMultipleTest );
	return 0;
}
