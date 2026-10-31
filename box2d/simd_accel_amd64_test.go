//go:build goexperiment.simd && amd64

package box2d

import (
	"os"
	"os/exec"
	"strconv"
	"testing"
)

func TestSIMDEnvironment(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		value    string
		debug    string
		expected bool
	}{
		{"default", "", os.Getenv("GODEBUG"), simdSupported},
		{"enabled", "1", os.Getenv("GODEBUG"), simdSupported},
		{"disabled", "0", os.Getenv("GODEBUG"), false},
		{"unsupported_cpu", "1", os.Getenv("GODEBUG") + ",cpu.avx=off", false},
	}

	for _, scenario := range cases {
		t.Run(scenario.name, func(t *testing.T) {
			command := exec.CommandContext(
				t.Context(),
				binary,
				"-test.run=^TestSIMDEnvironmentChild$",
			)

			command.Env = append(
				os.Environ(),
				"BOX2D_SIMD="+scenario.value,
				"GODEBUG="+scenario.debug,
				"BOX2D_TEST_SIMD_ACTIVE="+strconv.FormatBool(scenario.expected),
			)

			output, runErr := command.CombinedOutput()
			if runErr != nil {
				t.Fatalf("SIMD subprocess failed: %v\n%s", runErr, output)
			}
		})
	}
}

func TestSIMDEnvironmentChild(t *testing.T) {
	state := os.Getenv("BOX2D_TEST_SIMD_ACTIVE")
	if state == "" {
		t.Skip("subprocess only")
	}

	expected, err := strconv.ParseBool(state)
	if err != nil {
		t.Fatal(err)
	}

	if simdEnabled != expected {
		t.Fatalf("SIMD active = %t, want %t", simdEnabled, expected)
	}

	checkAABBOverlap(
		t,
		[4]uint32{0, 0, 0x3f800000, 0x3f800000},
		[4]uint32{0x3f800000, 0, 0x40000000, 0x3f800000},
	)
	checkAABBOverlap(
		t,
		[4]uint32{0, 0, 0x3f800000, 0x3f800000},
		[4]uint32{0x40000000, 0, 0x40400000, 0x3f800000},
	)
}
