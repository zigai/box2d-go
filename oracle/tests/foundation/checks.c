// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

#include "oracle.h"

#undef ORACLE_OPEN
#undef ORACLE_CLOSE
#undef ORACLE_I32
#undef ORACLE_U32
#undef ORACLE_U64
#undef ORACLE_BOOL
#define ORACLE_OPEN( ... ) ( (void)0 )
#define ORACLE_CLOSE( ... ) ( (void)0 )
#define ORACLE_I32( ... ) ( (void)0 )
#define ORACLE_U32( ... ) ( (void)0 )
#define ORACLE_U64( ... ) ( (void)0 )
#define ORACLE_BOOL( ... ) ( (void)0 )
#define ORACLE_FOUNDATION_BUFFER( ... ) ( (void)0 )
#define ORACLE_FOUNDATION_FLUSH( ... ) ( (void)0 )

#include "test_container.c"
#include "test_table.c"
#include "test_thread.c"
#include "test_id.c"
#include "test_bitset.c"

int main( void )
{
 return ContainerTest() || TableTest() || ThreadTest() || IdTest() || BitSetTest();
}
