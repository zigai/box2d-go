// SPDX-License-Identifier: MIT

#include "oracle.h"

#include <errno.h>
#include <inttypes.h>
#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>

#define ORACLE_KEY_BUFFERS 8
#define ORACLE_KEY_SIZE 256

static FILE* s_file;
static char s_name[256];
static char s_path[4096];
static uint64_t s_records;
static bool s_atexitRegistered;
static char s_keys[ORACLE_KEY_BUFFERS][ORACLE_KEY_SIZE];
static int s_keyIndex;

static void oracle_fail( const char* format, ... )
{
	va_list args;
	va_start( args, format );
	fprintf( stderr, "oracle: " );
	vfprintf( stderr, format, args );
	fprintf( stderr, "\n" );
	va_end( args );
	exit( 1 );
}

uint32_t oracle_bits_f32( float v )
{
	uint32_t bits;
	memcpy( &bits, &v, sizeof( bits ) );
	return bits;
}

uint64_t oracle_bits_f64( double v )
{
	uint64_t bits;
	memcpy( &bits, &v, sizeof( bits ) );
	return bits;
}

static void oracle_at_exit( void )
{
	if ( s_file != NULL )
	{
		fprintf( stderr, "oracle: fixture \"%s\" was opened but never closed\n", s_name );
		fclose( s_file );
		s_file = NULL;
		_Exit( 1 );
	}
}

static bool oracle_enabled( void )
{
	static int enabled = -1;
	if ( enabled < 0 )
	{
		const char* out = getenv( "ORACLE_OUT" );
		enabled = out != NULL && out[0] != '\0';
	}
	return enabled == 1;
}

static void oracle_make_dirs( char* path )
{
	for ( char* p = path + 1; *p != '\0'; ++p )
	{
		if ( *p == '/' )
		{
			*p = '\0';
			if ( mkdir( path, 0777 ) != 0 && errno != EEXIST )
			{
				oracle_fail( "cannot create directory %s: %s", path, strerror( errno ) );
			}
			*p = '/';
		}
	}
}

void oracle_open( const char* name )
{
	if ( oracle_enabled() == false )
	{
		return;
	}

	if ( s_file != NULL )
	{
		oracle_fail( "cannot open \"%s\": fixture \"%s\" is still open", name, s_name );
	}

	size_t length = strlen( name );
	if ( length == 0 || length >= sizeof( s_name ) || name[0] == '/' || name[length - 1] == '/' )
	{
		oracle_fail( "invalid fixture name \"%s\"", name );
	}

	for ( size_t i = 0; i < length; ++i )
	{
		char c = name[i];
		bool ok = ( c >= 'a' && c <= 'z' ) || ( c >= 'A' && c <= 'Z' ) || ( c >= '0' && c <= '9' ) || c == '_' || c == '-' ||
				  c == '.' || c == '/';
		if ( ok == false )
		{
			oracle_fail( "invalid character in fixture name \"%s\"", name );
		}
	}

	for ( const char* segment = name; *segment != '\0'; )
	{
		const char* end = strchr( segment, '/' );
		size_t segmentLength = end != NULL ? (size_t)( end - segment ) : strlen( segment );
		if ( segmentLength == 0 || segment[0] == '.' )
		{
			oracle_fail( "invalid fixture name \"%s\"", name );
		}
		segment += segmentLength;
		if ( *segment == '/' )
		{
			segment += 1;
		}
	}

	int written = snprintf( s_path, sizeof( s_path ), "%s/%s.jsonl", getenv( "ORACLE_OUT" ), name );
	if ( written < 0 || (size_t)written >= sizeof( s_path ) )
	{
		oracle_fail( "fixture path too long for \"%s\"", name );
	}

	oracle_make_dirs( s_path );

	s_file = fopen( s_path, "wbx" );
	if ( s_file == NULL )
	{
		oracle_fail( "cannot create %s: %s", s_path, strerror( errno ) );
	}

	memcpy( s_name, name, length + 1 );
	s_records = 0;

	if ( s_atexitRegistered == false )
	{
		atexit( oracle_at_exit );
		s_atexitRegistered = true;
	}
}

void oracle_close( void )
{
	if ( oracle_enabled() == false )
	{
		return;
	}

	if ( s_file == NULL )
	{
		oracle_fail( "oracle_close without an open fixture" );
	}

	if ( fclose( s_file ) != 0 )
	{
		oracle_fail( "cannot write %s: %s", s_path, strerror( errno ) );
	}
	s_file = NULL;
}

static void oracle_begin( const char* key, const char* type )
{
	if ( s_file == NULL )
	{
		oracle_fail( "record \"%s\" written without an open fixture", key );
	}

	fputs( "{\"k\":\"", s_file );
	for ( const unsigned char* c = (const unsigned char*)key; *c != '\0'; ++c )
	{
		if ( *c == '"' || *c == '\\' )
		{
			fprintf( s_file, "\\%c", *c );
		}
		else if ( *c < 0x20 || *c == 0x7f )
		{
			fprintf( s_file, "\\u%04x", *c );
		}
		else
		{
			fputc( *c, s_file );
		}
	}
	fprintf( s_file, "\",\"t\":\"%s\",\"v\":\"", type );
	s_records += 1;
}

static void oracle_end( void )
{
	if ( fputs( "\"}\n", s_file ) < 0 )
	{
		oracle_fail( "cannot write %s: %s", s_path, strerror( errno ) );
	}
}

void oracle_f32( const char* key, float v )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "f32" );
	fprintf( s_file, "%08" PRIx32, oracle_bits_f32( v ) );
	oracle_end();
}

void oracle_vec2( const char* key, b2Vec2 v )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "vec2" );
	fprintf( s_file, "%08" PRIx32 ",%08" PRIx32, oracle_bits_f32( v.x ), oracle_bits_f32( v.y ) );
	oracle_end();
}

void oracle_rot( const char* key, b2Rot q )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "rot" );
	fprintf( s_file, "%08" PRIx32 ",%08" PRIx32, oracle_bits_f32( q.c ), oracle_bits_f32( q.s ) );
	oracle_end();
}

void oracle_pos( const char* key, b2Pos p )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "pos" );
#if defined( BOX2D_DOUBLE_PRECISION )
	fprintf( s_file, "%016" PRIx64 ",%016" PRIx64, oracle_bits_f64( p.x ), oracle_bits_f64( p.y ) );
#else
	fprintf( s_file, "%08" PRIx32 ",%08" PRIx32, oracle_bits_f32( p.x ), oracle_bits_f32( p.y ) );
#endif
	oracle_end();
}

void oracle_i32( const char* key, int32_t v )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "i32" );
	fprintf( s_file, "%08" PRIx32, (uint32_t)v );
	oracle_end();
}

void oracle_u32( const char* key, uint32_t v )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "u32" );
	fprintf( s_file, "%08" PRIx32, v );
	oracle_end();
}

void oracle_u64( const char* key, uint64_t v )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "u64" );
	fprintf( s_file, "%016" PRIx64, v );
	oracle_end();
}

void oracle_bool( const char* key, bool v )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "bool" );
	fputc( v ? '1' : '0', s_file );
	oracle_end();
}

void oracle_bytes( const char* key, const void* data, size_t size )
{
	if ( oracle_enabled() == false )
	{
		return;
	}
	oracle_begin( key, "bytes" );
	const uint8_t* bytes = (const uint8_t*)data;
	for ( size_t i = 0; i < size; ++i )
	{
		fprintf( s_file, "%02x", bytes[i] );
	}
	oracle_end();
}

const char* oracle_key( const char* format, ... )
{
	char* buffer = s_keys[s_keyIndex];
	s_keyIndex = ( s_keyIndex + 1 ) % ORACLE_KEY_BUFFERS;

	va_list args;
	va_start( args, format );
	int written = vsnprintf( buffer, ORACLE_KEY_SIZE, format, args );
	va_end( args );

	if ( written < 0 || written >= ORACLE_KEY_SIZE )
	{
		oracle_fail( "key too long: %s", format );
	}
	return buffer;
}
