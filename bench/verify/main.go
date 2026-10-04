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

var families = []string{"JointGrid", "LargePyramid", "ManyPyramids", "Rain", "Spinner", "Smash", "Tumbler", "Washer", "Junkyard", "Sleep", "Compounds", "Queries", "TreeCast", "TileWorld", "FallingHinges", "Human"}

func validate(path, precision string, samples, cpus int, goOutput bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
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
	config := fmt.Sprintf("box2d-bench: precision=%s scene-config=release steps=300 timestep=1/60 substeps=4 seed=12345 continuous=true rebuild-static=false cpus=%d", precision, cpus)
	configurations, rows := 0, 0
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		text := scanner.Text()
		if strings.HasPrefix(text, "box2d-bench:") {
			if text != config {
				return fmt.Errorf("%s:%d: unmatched workload configuration: %s", path, line, text)
			}
			configurations++
			continue
		}
		if !strings.HasPrefix(text, "Benchmark") {
			continue
		}
		f := strings.Fields(text)
		if len(f) < 4 {
			return fmt.Errorf("%s:%d: incomplete benchmark row", path, line)
		}
		left, ok := expected[f[0]]
		if !ok || left == 0 {
			return fmt.Errorf("%s:%d: unexpected or excess row %s", path, line, f[0])
		}
		if f[1] != "1" {
			return fmt.Errorf("%s:%d: expected exactly one complete 300-step workload, got %s iterations", path, line, f[1])
		}
		want := 4
		if goOutput {
			want = 8
		}
		if len(f) != want {
			return fmt.Errorf("%s:%d: expected %d columns, got %d", path, line, want, len(f))
		}
		if f[3] != "ns/op" {
			return fmt.Errorf("%s:%d: missing timing metric", path, line)
		}
		for i := 2; i < len(f); i += 2 {
			v, e := strconv.ParseFloat(f[i], 64)
			if e != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || i == 2 && v == 0 {
				return fmt.Errorf("%s:%d: invalid measurement %q", path, line, f[i])
			}
		}
		if goOutput && (f[5] != "B/op" || f[7] != "allocs/op") {
			return fmt.Errorf("%s:%d: incomplete Go allocation metrics", path, line)
		}
		expected[f[0]] = left - 1
		rows++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if configurations == 0 {
		return fmt.Errorf("%s: missing workload configuration", path)
	}
	for label, left := range expected {
		if left != 0 {
			return fmt.Errorf("%s: missing %d samples for %s", path, left, label)
		}
	}
	if rows != len(families)*2*samples {
		return fmt.Errorf("%s: wrong complete row count %d", path, rows)
	}
	return nil
}

func main() {
	samples := flag.Int("samples", 6, "Samples per family and worker")
	cpus := flag.Int("cpus", 2, "Matched concurrency label")
	precision := flag.String("precision", "float", "Precision: float or double")
	flag.Parse()
	if *samples < 1 || *cpus < 1 || (*precision != "float" && *precision != "double") || flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: verify -samples <positive-count> -cpus <positive-count> -precision <float|double> <go-output> <c-output>")
		os.Exit(2)
	}
	for i, path := range flag.Args() {
		if err := validate(path, *precision, *samples, *cpus, i == 0); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	fmt.Printf("Verified %s: %d complete matched rows per process (16 families, workers 1/4, %d samples).\n", *precision, len(families)*2**samples, *samples)
}
