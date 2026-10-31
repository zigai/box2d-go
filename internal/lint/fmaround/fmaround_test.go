package fmaround_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/zigai/box2d-go/internal/lint/fmaround"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), fmaround.Analyzer, "a", "clean", "partitions")
}

func TestSuggestedFixes(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), fmaround.Analyzer, "a")
}
