# Deterministic reference identity; volatile logs remain in the build directory.
file(STRINGS "${CMAKE_CURRENT_SOURCE_DIR}/../UPSTREAM" ORACLE_PIN REGEX "^[0-9a-f]+$")
list(LENGTH ORACLE_PIN pin_count)
if(NOT pin_count EQUAL 1)
	message(FATAL_ERROR "UPSTREAM must contain exactly one commit hash")
endif()
string(LENGTH "${ORACLE_PIN}" pin_length)
if(NOT pin_length EQUAL 40)
	message(FATAL_ERROR "UPSTREAM must contain a 40-character commit hash")
endif()

file(GLOB_RECURSE adapter_sources CONFIGURE_DEPENDS
	"${CMAKE_CURRENT_SOURCE_DIR}/sample/*.c"
	"${CMAKE_CURRENT_SOURCE_DIR}/sample/*.h"
	"${CMAKE_CURRENT_SOURCE_DIR}/tests/*.c"
	"${CMAKE_CURRENT_SOURCE_DIR}/tests/*.h"
	"${CMAKE_CURRENT_SOURCE_DIR}/debug/*.c"
	"${CMAKE_CURRENT_SOURCE_DIR}/debug/*.h"
	"${CMAKE_CURRENT_SOURCE_DIR}/bench/*.c"
	"${CMAKE_CURRENT_SOURCE_DIR}/bench/*.h")
list(APPEND adapter_sources
	"${CMAKE_CURRENT_SOURCE_DIR}/oracle.c"
	"${CMAKE_CURRENT_SOURCE_DIR}/oracle.h"
	"${CMAKE_CURRENT_SOURCE_DIR}/differential.c"
	"${CMAKE_CURRENT_SOURCE_DIR}/CMakeLists.txt"
	"${CMAKE_CURRENT_SOURCE_DIR}/build.sh"
	"${CMAKE_CURRENT_SOURCE_DIR}/reference.cmake"
	"${CMAKE_CURRENT_SOURCE_DIR}/capture.cmake"
	"${CMAKE_CURRENT_SOURCE_DIR}/check-upstream.cmake"
	"${CMAKE_CURRENT_SOURCE_DIR}/cases.json")
list(SORT adapter_sources)
set(source_json "{}")
foreach(source IN LISTS adapter_sources)
	file(RELATIVE_PATH relative "${CMAKE_CURRENT_SOURCE_DIR}/.." "${source}")
	file(SHA256 "${source}" digest)
	string(JSON source_json SET "${source_json}" "${relative}" "\"${digest}\"")
endforeach()

execute_process(COMMAND "${CMAKE_C_COMPILER}" -dumpmachine
	RESULT_VARIABLE target_result OUTPUT_VARIABLE compiler_target
	OUTPUT_STRIP_TRAILING_WHITESPACE ERROR_VARIABLE target_error)
if(NOT target_result EQUAL 0)
	message(FATAL_ERROR "Cannot determine C compiler target: ${target_error}")
endif()
if(BOX2D_DOUBLE_PRECISION)
	set(precision double)
else()
	set(precision float)
endif()
if(BOX2D_DISABLE_SIMD)
	set(scalar true)
else()
	set(scalar false)
endif()
if(CMAKE_BUILD_TYPE STREQUAL "Debug" OR CMAKE_BUILD_TYPE STREQUAL "RelWithDebInfo")
	set(assertions true)
else()
	set(assertions false)
endif()
string(TOUPPER "${CMAKE_BUILD_TYPE}" build_type_upper)
set(reference "{}")
string(JSON reference SET "${reference}" schema 1)
string(JSON reference SET "${reference}" upstream "\"${ORACLE_PIN}\"")
string(JSON reference SET "${reference}" precision "\"${precision}\"")
string(JSON reference SET "${reference}" scalar "${scalar}")
string(JSON reference SET "${reference}" assertions "${assertions}")
string(JSON reference SET "${reference}" contraction "\"off\"")
string(JSON reference SET "${reference}" compiler "\"${CMAKE_C_COMPILER_ID}\"")
string(JSON reference SET "${reference}" compiler_version "\"${CMAKE_C_COMPILER_VERSION}\"")
string(JSON reference SET "${reference}" compiler_target "\"${compiler_target}\"")
string(JSON reference SET "${reference}" build_type "\"${CMAKE_BUILD_TYPE}\"")
string(JSON reference SET "${reference}" flags "\"${CMAKE_C_FLAGS} ${CMAKE_C_FLAGS_${build_type_upper}} -ffp-contract=off\"")
string(JSON reference SET "${reference}" sources "${source_json}")
file(WRITE "${CMAKE_BINARY_DIR}/reference.json" "${reference}\n")
