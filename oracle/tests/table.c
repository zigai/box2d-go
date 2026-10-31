// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"
#include <assert.h>

static unsigned char table_checks[100172];
static size_t table_check_count;

static void table_buffer( bool value )
{
 assert( table_check_count < sizeof( table_checks ) );
 table_checks[table_check_count++] = value ? 1 : 0;
}

static void table_flush( const char* key )
{
 ORACLE_BYTES( key, table_checks, table_check_count );
 table_check_count = 0;
}

#define ORACLE_TABLE_BUFFER( value ) table_buffer( value )
#define ORACLE_TABLE_FLUSH( key ) table_flush( key )

#include "upstream/test_table.c"

int main( void )
{
 return TableTest();
}
