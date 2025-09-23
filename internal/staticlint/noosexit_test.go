package staticlint

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestNoOsExitAnalyzer(t *testing.T) {
	tests := []struct {
		pkgPath   string
		wantError bool
	}{
		{"with_exit", true},
		{"without_exit", false},
		{"not_main", false},
	}

	testdata := analysistest.TestData()

	for _, tt := range tests {
		t.Run(tt.pkgPath, func(t *testing.T) {
			analysistest.Run(t, testdata, NoOsExitAnalyzer, tt.pkgPath)
		})
	}
}
