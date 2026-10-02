// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"
#include <assert.h>

static unsigned char foundation_checks[100172];
static size_t foundation_check_count;

static void foundation_buffer( bool value )
{
 assert( foundation_check_count < sizeof( foundation_checks ) );
 foundation_checks[foundation_check_count++] = value ? 1 : 0;
}

static void foundation_flush( const char* key )
{
 ORACLE_BYTES( key, foundation_checks, foundation_check_count );
 foundation_check_count = 0;
}

#define ORACLE_FOUNDATION_BUFFER( value ) foundation_buffer( value )
#define ORACLE_FOUNDATION_FLUSH( key ) foundation_flush( key )

#include "foundation/test_table.c"

int main( void )
{
 return TableTest();
}
