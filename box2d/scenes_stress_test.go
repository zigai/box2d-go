// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

//go:build box2d_stress

package box2d_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/zigai/box2d-go/internal/oracle"
)

func TestScenes(t *testing.T) { testScenes(t, 1) }

func TestParallelScenes(t *testing.T) { testScenes(t, 4) }

func TestSceneFixtureInventory(t *testing.T) {
	names := []string{"JointGrid", "LargePyramid", "ManyPyramids", "Rain", "Spinner", "Smash", "Tumbler", "Washer", "Junkyard", "Sleep", "Compounds", "Queries", "TreeCast", "TileWorld", "FallingHinges", "Human"}

	for _, config := range []string{"float", "double"} {
		t.Run(config, func(t *testing.T) {
			checkSceneFixtureInventory(t, config, names)
		})
	}
}

func checkSceneFixtureInventory(t *testing.T, config string, names []string) {
	t.Helper()

	width := 4
	if config == "double" {
		width = 8
	}

	benchmarks := int64(0)

	for _, workers := range []int{1, 4} {
		for index, name := range names {
			fixture := fmt.Sprintf("%s_workers%d", name, workers)
			t.Run(fixture, func(t *testing.T) {
				size := checkSceneWorkerFixture(t, config, fixture, name, workers, width)
				if index < 14 {
					benchmarks += size
				} else if size > 1048576 {
					t.Fatal("area exceeds default budget")
				}
			})
		}
	}

	if benchmarks > 67108864 {
		t.Fatalf("benchmark fixtures exceed named budget: %d", benchmarks)
	}

	checkSceneFixtureDirectory(t, config, names)
	checkSceneUtilityFixture(t, config, width)
}

func TestSceneHelperFixtureInventory(t *testing.T) {
	for _, config := range []string{"float", "double"} {
		t.Run(config, func(t *testing.T) {
			data, err := oracle.ReadFile(config, "scenes/Helpers")
			if err != nil {
				t.Fatal(err)
			}

			if len(data) > 1048576 {
				t.Fatal("helpers exceed default budget")
			}

			lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})

			width := 4
			if config == "double" {
				width = 8
			}

			c := sceneLineCursor{t: t, lines: lines, width: width}
			c.checkGroupLifecycle()
			c.checkPaddleBoundaries()

			if c.next != len(lines) {
				t.Fatal("extra helper records")
			}

			t.Logf("%d bytes; group destruction/recycling and four strict paddle boundaries", len(data))
		})
	}
}
