// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include <stdlib.h>

#include "input/trace.h"
#include "input/test_determinism.c"

int TestAssertFcn( const char* condition, const char* fileName, int lineNumber )
{
	fprintf( stderr, "%s:%d: %s\n", fileName, lineNumber, condition );
	abort();
}

void TestLogFcn( const char* message )
{
	fprintf( stderr, "%s\n", message );
}

int main( void )
{
	b2SetAssertFcn( TestAssertFcn );
	b2SetLogFcn( TestLogFcn );
	return DeterminismTest();
}
