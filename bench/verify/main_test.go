// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCustomMeasurementSteps(t *testing.T) {
	for _, precision := range []string{"float", "double"} {
		for _, goOutput := range []bool{false, true} {
			good := strings.Replace(measurementFixture(precision, 2, goOutput), "steps=300", "steps=500", 1)

			cases := []struct {
				name  string
				text  string
				steps int
				valid bool
			}{
				{name: "custom", text: good, steps: 500, valid: true},
				{name: "single", text: strings.Replace(good, "steps=500", "steps=1", 1), steps: 1, valid: true},
				{name: "wrong-expected-horizon", text: good, steps: 300},
				{name: "wrong-recorded-horizon", text: strings.Replace(good, "steps=500", "steps=300", 1), steps: 500},
				{name: "mixed-horizons", text: good + strings.SplitAfter(measurementFixture(precision, 2, goOutput), "\n")[0], steps: 500},
			}
			for _, testCase := range cases {
				t.Run(fmt.Sprintf("%s/go=%t/%s", precision, goOutput, testCase.name), func(test *testing.T) {
					path := filepath.Join(test.TempDir(), "measurements.txt")
					if err := os.WriteFile(path, []byte(testCase.text), 0o600); err != nil {
						test.Fatal(err)
					}

					err := validate(path, precision, 2, 2, testCase.steps, goOutput)
					if (err == nil) != testCase.valid {
						test.Fatalf("valid=%t, error=%v", testCase.valid, err)
					}
				})
			}
		}
	}
}

func TestMeasurementRowsAreCompleteMatchedAndFinite(t *testing.T) {
	for _, precision := range []string{"float", "double"} {
		for _, cpus := range []int{1, 4} {
			for _, goOutput := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/cpus=%d/go=%t", precision, cpus, goOutput), func(t *testing.T) {
					checkMeasurementVariants(t, precision, cpus, goOutput)
				})
			}
		}
	}
}

func measurementFixture(precision string, cpus int, goOutput bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "box2d-bench: precision=%s scene-config=release steps=300 timestep=1/60 substeps=4 seed=12345 continuous=true rebuild-static=false cpus=%d\n", precision, cpus)

	for _, name := range families {
		for _, workers := range []int{1, 4} {
			for range 2 {
				label := fmt.Sprintf("BenchmarkScenes/%s/workers=%d", name, workers)
				if cpus != 1 {
					label += fmt.Sprintf("-%d", cpus)
				}

				fmt.Fprintf(&b, "%s 1 12345 ns/op", label)

				if goOutput {
					b.WriteString(" 0 B/op 0 allocs/op")
				}

				b.WriteByte('\n')
			}
		}
	}

	return b.String()
}

func checkMeasurementVariants(t *testing.T, precision string, cpus int, goOutput bool) {
	t.Helper()

	good := measurementFixture(precision, cpus, goOutput)
	lines := strings.SplitAfter(good, "\n")

	cases := []struct {
		name, text string
		valid      bool
	}{
		{"debug-dimensions", strings.Replace(good, "scene-config=release", "scene-config=debug", 1), false},
		{"complete", good, true},
		{"missing-family", strings.Join(lines[:len(lines)-3], ""), false},
		{"extra-sample", good + lines[1], false},
		{"wrong-worker", strings.Replace(good, "workers=1", "workers=2", 1), false},
		{"wrong-precision", strings.Replace(good, "precision="+precision, "precision=other", 1), false},
		{"wrong-horizon", strings.Replace(good, "steps=300", "steps=299", 1), false},
		{"wrong-seed", strings.Replace(good, "seed=12345", "seed=42", 1), false},
		{"zero", strings.Replace(good, "12345 ns/op", "0 ns/op", 1), false},
		{"negative", strings.Replace(good, "12345 ns/op", "-1 ns/op", 1), false},
		{"nan", strings.Replace(good, "12345 ns/op", "NaN ns/op", 1), false},
		{"infinity", strings.Replace(good, "12345 ns/op", "+Inf ns/op", 1), false},
		{"partial-iterations", strings.Replace(good, " 1 12345", " 2 12345", 1), false},
		{"missing-metadata", strings.Join(lines[1:], ""), false},
		{"malformed", good + "BenchmarkScenes/Human/workers=1\n", false},
	}
	if goOutput {
		cases = append(cases,
			struct {
				name, text string
				valid      bool
			}{"negative-bytes", strings.Replace(good, "0 B/op", "-1 B/op", 1), false},
			struct {
				name, text string
				valid      bool
			}{"nan-allocations", strings.Replace(good, "0 allocs/op", "NaN allocs/op", 1), false},
			struct {
				name, text string
				valid      bool
			}{"missing-allocations", strings.Replace(good, " 0 B/op 0 allocs/op", "", 1), false})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "measurements.txt")
			if err := os.WriteFile(path, []byte(tc.text), 0o600); err != nil {
				t.Fatal(err)
			}

			err := validate(path, precision, 2, cpus, 300, goOutput)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t, error=%v", tc.valid, err)
			}
		})
	}
}
