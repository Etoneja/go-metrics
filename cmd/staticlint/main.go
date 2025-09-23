package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/etoneja/go-metrics/internal/staticlint"
)

func main() {
	analyzers := staticlint.GetAnalyzers()
	multichecker.Main(analyzers...)
}
