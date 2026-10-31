// SPDX-License-Identifier: MIT

package box2d_test

import (
	"testing"

	"github.com/zigai/box2d-go/internal/oracle"
	"github.com/zigai/box2d-go/internal/oracle/sequence"
)

// This offline comparison is collected by the ordinary root suite, including
// WASM. Native live exploration uses the identical executor and seed inputs.
func TestDifferentialCorpus(t *testing.T) {
	for _, seed := range sequence.Corpus() {
		t.Run(seed.Name, func(t *testing.T) {
			r := oracle.Open(t, "differential/corpus/"+seed.Name)
			sequence.Execute(t, r, seed.Input)
		})
	}
}
