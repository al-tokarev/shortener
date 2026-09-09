package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestCheckExitInMain(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ErrExitAnalyzer, "./...")
}
