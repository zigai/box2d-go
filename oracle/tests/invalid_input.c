// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

#include <stdlib.h>

#include "trace.h"
#include "upstream/test_invalid_input.c"

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
	return InvalidInputTest();
}
