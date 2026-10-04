// SPDX-License-Identifier: MIT
// The input inventory is mirrored by internal/oracle/sequence.Corpus. These
// are inputs, not expected results: every observation is emitted by real C.
#define BOX2D_DIFFERENTIAL_NO_MAIN
#include "../differential.c"

static const uint8_t zero_partial[] = { 0 };
static const uint8_t lifecycle_reuse[] = { 0,  0, 0, 0, 2,	0, 0, 0, 0,	 1, 24, 0, 2,  1, 2, 0, 14, 0, 1,  3, 15, 0, 0, 0,
										   14, 0, 1, 3, 3,	0, 0, 0, 2,	 0, 3,	0, 1,  1, 0, 0, 0,	1, 24, 0, 2,  1, 2, 0,
										   14, 1, 0, 3, 12, 0, 0, 3, 13, 0, 0,	0, 20, 0, 0, 0, 0,	0, 0,  0, 2,  0, 0, 0 };
static const uint8_t contact_transitions[] = { 0,  0, 0,  0, 2,	 0, 4, 0, 9,  0, 0,	 0, 0, 1, 12, 0, 2,	 1,	 4,	 0, 12, 0,
											   0,  3, 12, 0, 0,	 3, 8, 1, 1,  0, 12, 0, 0, 3, 8,  1, 0,	 15, 12, 0, 0,	3,
											   10, 1, 0,  0, 12, 0, 0, 3, 11, 1, 0,	 0, 4, 1, 12, 0, 12, 0,	 0,	 3, 9,	1,
											   1,  0, 12, 0, 0,	 3, 9, 1, 2,  0, 12, 0, 0, 3, 1,  0, 0,	 0,	 12, 0, 0,	3 };
static const uint8_t snapshot_continuation[] = { 0, 0,	0,	0, 2, 0,  4,  0, 9, 0, 0, 0,  0,  1, 12, 0, 2, 1,  4,  0, 12, 0,  0,
												 3, 16, 0,	0, 0, 10, 1,  0, 0, 0, 2, 32, 32, 2, 2,	 2, 0, 14, 1,  2, 3,  12, 0,
												 0, 3,	17, 0, 0, 0,  12, 0, 0, 3, 4, 1,  48, 0, 13, 0, 0, 0,  12, 0, 0,  3 };
static const uint8_t replay_seek_restep[] = { 0, 0, 0, 0, 2,  0, 4, 0, 9,  0, 0, 0, 0,	1, 12, 0,	2,	1, 4, 0, 0,	 2, 24, 24,
											  2, 2, 2, 0, 14, 1, 2, 3, 12, 0, 0, 3, 19, 0, 16, 240, 13, 0, 0, 0, 12, 0, 0,	3 };
static const uint8_t force_velocity_zero_step[] = { 0,	 0,	  128, 127, 2, 0, 4,  0, 5,	  0,   255, 127, 6, 0, 128, 0, 7,	0,
													127, 128, 12,  0,	1, 3, 18, 0, 127, 128, 12,	0,	 0, 3, 13,	0, 128, 127 };
int main( void )
{
	uint8_t maximum[input_limit];
	for ( int i = 0; i < 64; ++i )
	{
		maximum[4 * i] = (uint8_t)( i % 21 );
		maximum[4 * i + 1] = maximum[4 * i + 2] = maximum[4 * i + 3] = 255;
	}
	for ( int i = 0; i < slot_count; ++i )
	{
		const uint8_t create[8] = { 0, (uint8_t)i, 255, 255, 2, (uint8_t)i, 255, 255 };
		memcpy( maximum + 8 * i, create, sizeof( create ) );
	}
	if ( run_sequence( "differential/corpus/empty", NULL, 0 ) != 0 )
		return 1;
	if ( run_sequence( "differential/corpus/zero_partial", zero_partial, sizeof( zero_partial ) ) != 0 )
		return 1;
	if ( run_sequence( "differential/corpus/maximum_operands", maximum, sizeof( maximum ) ) != 0 )
		return 1;
	if ( run_sequence( "differential/corpus/lifecycle_reuse", lifecycle_reuse, sizeof( lifecycle_reuse ) ) != 0 )
		return 1;
	if ( run_sequence( "differential/corpus/contact_transitions", contact_transitions, sizeof( contact_transitions ) ) != 0 )
		return 1;
	if ( run_sequence( "differential/corpus/snapshot_continuation", snapshot_continuation, sizeof( snapshot_continuation ) ) !=
		 0 )
		return 1;
	if ( run_sequence( "differential/corpus/replay_seek_restep", replay_seek_restep, sizeof( replay_seek_restep ) ) != 0 )
		return 1;
	if ( run_sequence( "differential/corpus/force_velocity_zero_step", force_velocity_zero_step,
					   sizeof( force_velocity_zero_step ) ) != 0 )
		return 1;
	return 0;
}
