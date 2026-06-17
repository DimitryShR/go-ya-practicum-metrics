package osexitcheck_test

import (
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/cmd/staticlint/analyzers/osexitcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer tests the osexitcheck analyzer using testdata.
func TestAnalyzer(t *testing.T) {
	// Run the analyzer on the testdata directory
	analysistest.Run(t, analysistest.TestData(), osexitcheck.Analyzer, "./...")
}
