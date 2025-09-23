package staticlint

import (
	"testing"
)

func TestGetAnalyzers(t *testing.T) {
	analyzers := GetAnalyzers()
	
	if len(analyzers) == 0 {
		t.Fatal("Expected non-empty analyzers slice")
	}
	
	t.Logf("Loaded %d analyzers", len(analyzers))
}

func TestGetStandardAnalyzers(t *testing.T) {
	analyzers := getStandardAnalyzers()

	if len(analyzers) == 0 {
		t.Error("Expected standard analyzers, got none")
	}

	found := false
	for _, a := range analyzers {
		if a.Name == "printf" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected printf analyzer not found")
	}
}

func TestGetCustomAnalyzers(t *testing.T) {
	analyzers := getCustomAnalyzers()

	if len(analyzers) != 1 {
		t.Errorf("Expected 1 custom analyzer, got %d", len(analyzers))
	}

	if analyzers[0].Name != "noosexit" {
		t.Errorf("Expected noosexit analyzer, got %s", analyzers[0].Name)
	}
}

func TestGetStaticCheckAnalyzers(t *testing.T) {
	analyzers := getStaticCheckAnalyzers()

	if len(analyzers) == 0 {
		t.Error("Expected staticcheck analyzers, got none")
	}

	for _, a := range analyzers {
		if a.Name == "ST1000" {
			t.Error("ST1000 should be excluded by blacklist")
		}
	}
}

func TestGetPublicAnalyzers(t *testing.T) {
	analyzers := getPublicAnalyzers()

	if len(analyzers) != 2 {
		t.Errorf("Expected 2 public analyzers, got %d", len(analyzers))
	}
}
