cmake_minimum_required(VERSION 3.22)
project(box2d_oracle VERSION 0.0.0 LANGUAGES C)

set(BOX2D_UPSTREAM_DIR "" CACHE PATH "upstream Box2D checkout at the pin")
option(BOX2D_DISABLE_SIMD "Scalar math only, the reference configuration" ON)
option(BOX2D_DOUBLE_PRECISION "Large world (double precision position) configuration" OFF)
option(BOX2D_VALIDATE "Heavy validation. Inert in release builds, so off." OFF)

if(NOT EXISTS "${BOX2D_UPSTREAM_DIR}/src/CMakeLists.txt")
	message(FATAL_ERROR "BOX2D_UPSTREAM_DIR must name an upstream checkout, got '${BOX2D_UPSTREAM_DIR}'")
endif()

if(NOT CMAKE_BUILD_TYPE)
	set(CMAKE_BUILD_TYPE RelWithDebInfo)
endif()

add_compile_options(-ffp-contract=off)

set(ORACLE_SIMD_DEFINITIONS "")
if(BOX2D_DISABLE_SIMD)
	set(ORACLE_SIMD_DEFINITIONS BOX2D_DISABLE_SIMD)
endif()
set(BOX2D_SIMD_DEFINITIONS ${ORACLE_SIMD_DEFINITIONS})
set(BOX2D_SIMD_OPTIONS "")

add_subdirectory("${BOX2D_UPSTREAM_DIR}/src" box2d)
add_subdirectory("${BOX2D_UPSTREAM_DIR}/shared" shared)

# Build the unchanged upstream runner, including its own registered case list.
# Its original include path assumes it is the root project, so add the actual src.
add_subdirectory("${BOX2D_UPSTREAM_DIR}/test" upstream_tests)
target_include_directories(test PRIVATE "${BOX2D_UPSTREAM_DIR}/src")
set_target_properties(test PROPERTIES
	OUTPUT_NAME upstream_tests
	RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/upstream")

add_library(oracle STATIC oracle.c)
target_include_directories(oracle PUBLIC "${CMAKE_CURRENT_SOURCE_DIR}")
target_link_libraries(oracle PUBLIC box2d)
set_target_properties(oracle PROPERTIES C_STANDARD 17 C_STANDARD_REQUIRED YES)

file(GLOB ORACLE_SAMPLES CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/sample/*.c")
file(GLOB ORACLE_TESTS CONFIGURE_DEPENDS "${CMAKE_CURRENT_SOURCE_DIR}/tests/*.c")
list(SORT ORACLE_SAMPLES)
list(SORT ORACLE_TESTS)

function(oracle_program dir path)
	get_filename_component(name "${path}" NAME_WE)
	if(ARGC GREATER 2)
		set(target "${ARGV2}")
	else()
		set(target "${dir}_${name}")
	endif()
	add_executable(${target} "${path}")
	target_include_directories(${target} PRIVATE
		"${BOX2D_UPSTREAM_DIR}/include"
		"${BOX2D_UPSTREAM_DIR}/src"
		"${BOX2D_UPSTREAM_DIR}/shared"
		"${BOX2D_UPSTREAM_DIR}/test")
	target_compile_definitions(${target} PRIVATE ${ORACLE_SIMD_DEFINITIONS})
	target_link_libraries(${target} PRIVATE oracle shared box2d)
	set_target_properties(${target} PROPERTIES
		C_STANDARD 17
		C_STANDARD_REQUIRED YES
		C_EXTENSIONS NO
		RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bin")
endfunction()

foreach(path IN LISTS ORACLE_SAMPLES)
	oracle_program(sample "${path}")
endforeach()
foreach(path IN LISTS ORACLE_TESTS)
	oracle_program(tests "${path}")
	get_filename_component(program "${path}" NAME_WE)
	if(program STREQUAL "collision_distance_public")
		set_target_properties(tests_${program} PROPERTIES
			EXCLUDE_FROM_ALL TRUE
			RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/tools")
	endif()
endforeach()

# Input-driven comparisons are explicit integration runs, not fixture generators.
oracle_program(tools "${CMAKE_CURRENT_SOURCE_DIR}/differential.c")
set_target_properties(tools_differential PROPERTIES
	OUTPUT_NAME differential
	RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/tools")

# Requested simplex histories are an explicit Debug-engine reference profile.
oracle_program(tests "${CMAKE_CURRENT_SOURCE_DIR}/debug/distance.c" tests_collision_distance_debug_public)
set_target_properties(tests_collision_distance_debug_public PROPERTIES
	EXCLUDE_FROM_ALL TRUE
	RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/tools")

add_executable(bench_scenes bench/scenes.c)
target_include_directories(bench_scenes PRIVATE "${BOX2D_UPSTREAM_DIR}/include")
target_compile_definitions(bench_scenes PRIVATE ${ORACLE_SIMD_DEFINITIONS})
target_link_libraries(bench_scenes PRIVATE shared box2d)
set_target_properties(bench_scenes PROPERTIES
	C_STANDARD 17
	C_STANDARD_REQUIRED YES
	C_EXTENSIONS NO
	RUNTIME_OUTPUT_DIRECTORY "${CMAKE_BINARY_DIR}/bench")

include("${CMAKE_CURRENT_SOURCE_DIR}/reference.cmake")
