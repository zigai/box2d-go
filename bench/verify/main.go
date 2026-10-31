// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

// Command verify rejects incomplete or differently configured measurement files.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	defaultSamples = 6
	defaultSteps   = 300
	defaultCPUs    = 2
	usageExitCode  = 2
)

const (
	nameColumn       = 0
	iterationsColumn = 1
	timingColumn     = 2
	timingUnitColumn = 3
	bytesUnitColumn  = 5
	allocsUnitColumn = 7
	metricColumns    = 2
	cRowColumns      = 4
	goRowColumns     = 8
)

var families = []string{
	"JointGrid",
	"LargePyramid",
	"ManyPyramids",
	"Rain",
	"Spinner",
	"Smash",
	"Tumbler",
	"Washer",
	"Junkyard",
	"Sleep",
	"Compounds",
	"Queries",
	"TreeCast",
	"TileWorld",
	"FallingHinges",
	"Human",
}

func validate(
	path string,
	precision string,
	samples int,
	cpus int,
	steps int,
	goOutput bool,
) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open measurements: %w", err)
	}
	defer file.Close() //nolint:errcheck // Read-only file; closing cannot affect results.

	expected := expectedBenchmarks(samples, cpus)

	config := fmt.Sprintf(
		"box2d-bench: precision=%s scene-config=release steps=%d timestep=1/60 substeps=4 "+
			"seed=12345 continuous=true rebuild-static=false cpus=%d",
		precision,
		steps,
		cpus,
	)
	configurations, rows := 0, 0

	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		if strings.HasPrefix(text, "box2d-bench:") {
			if text != config {
				return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
					"%s:%d: unmatched workload configuration: %s",
					path,
					line,
					text,
				)
			}

			configurations++

			continue
		}

		if !strings.HasPrefix(text, "Benchmark") {
			continue
		}

		if err := validateBenchmarkRow(
			path,
			line,
			strings.Fields(text),
			goOutput,
			expected,
		); err != nil {
			return err
		}

		rows++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan measurements: %w", err)
	}

	if configurations == 0 {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s: missing workload configuration",
			path,
		)
	}

	for label, left := range expected {
		if left != 0 {
			return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
				"%s: missing %d samples for %s",
				path,
				left,
				label,
			)
		}
	}

	if rows != len(families)*2*samples {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s: wrong complete row count %d",
			path,
			rows,
		)
	}

	return nil
}

func expectedBenchmarks(samples int, cpus int) map[string]int {
	expected := make(map[string]int)

	for _, name := range families {
		for _, workers := range []int{1, 4} {
			label := fmt.Sprintf("BenchmarkScenes/%s/workers=%d", name, workers)
			if cpus != 1 {
				label += fmt.Sprintf("-%d", cpus)
			}

			expected[label] = samples
		}
	}

	return expected
}

func validateBenchmarkRow(
	path string,
	line int,
	f []string,
	goOutput bool,
	expected map[string]int,
) error {
	if len(f) < cRowColumns {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s:%d: incomplete benchmark row",
			path,
			line,
		)
	}

	left, ok := expected[f[nameColumn]]
	if !ok || left == 0 {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s:%d: unexpected or excess row %s",
			path,
			line,
			f[nameColumn],
		)
	}

	if f[iterationsColumn] != "1" {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s:%d: expected exactly one complete workload, got %s iterations",
			path,
			line,
			f[iterationsColumn],
		)
	}

	if err := validateMeasurements(path, line, f, goOutput); err != nil {
		return err
	}

	expected[f[nameColumn]] = left - 1

	return nil
}

func validateMeasurements(path string, line int, f []string, goOutput bool) error {
	want := cRowColumns
	if goOutput {
		want = goRowColumns
	}

	if len(f) != want {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s:%d: expected %d columns, got %d",
			path,
			line,
			want,
			len(f),
		)
	}

	if f[timingUnitColumn] != "ns/op" {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s:%d: missing timing metric",
			path,
			line,
		)
	}

	for i := timingColumn; i < len(f); i += metricColumns {
		if !validMeasurement(f[i], i == timingColumn) {
			return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
				"%s:%d: invalid measurement %q",
				path,
				line,
				f[i],
			)
		}
	}

	if goOutput && (f[bytesUnitColumn] != "B/op" || f[allocsUnitColumn] != "allocs/op") {
		return fmt.Errorf( //nolint:err113 // Terminal CLI error; callers never match it.
			"%s:%d: incomplete Go allocation metrics",
			path,
			line,
		)
	}

	return nil
}

func validMeasurement(text string, timing bool) bool {
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || timing && v == 0 {
		return false
	}

	return true
}

func main() {
	samples := flag.Int("samples", defaultSamples, "Samples per family and worker")
	steps := flag.Int("steps", defaultSteps, "Steps per complete workload")
	cpus := flag.Int("cpus", defaultCPUs, "Matched concurrency label")
	precision := flag.String("precision", "float", "Precision: float or double")

	flag.Parse()

	if *samples < 1 || *cpus < 1 || *steps < 1 || *steps > math.MaxInt32 ||
		(*precision != "float" && *precision != "double") ||
		flag.NArg() != 2 {
		fmt.Fprintln(
			os.Stderr,
			"usage: verify -samples <positive-count> -steps <1..2147483647> -cpus <positive-count> "+
				"-precision <float|double> <go-output> <c-output>",
		)
		os.Exit(usageExitCode)
	}

	for i, path := range flag.Args() {
		if err := validate(path, *precision, *samples, *cpus, *steps, i == 0); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Printf(
		"Verified %s: %d complete matched rows per process (16 families, workers 1/4, %d samples, %d steps).\n",
		*precision,
		len(families)*2**samples,
		*samples,
		*steps,
	)
}
