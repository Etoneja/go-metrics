package version

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	old := os.Stdout
	defer func() { os.Stdout = old }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe failed: %v", err)
	}
	os.Stdout = w

	Print()

	err = w.Close()
	if err != nil {
		t.Logf("Close writer failed: %v", err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Logf("Copy failed: %v", err)
	}
	output := buf.String()

	expected := []string{
		"Build version: " + BuildVersion,
		"Build date: " + BuildDate,
		"Build commit: " + BuildCommit,
	}

	for _, exp := range expected {
		if !bytes.Contains([]byte(output), []byte(exp)) {
			t.Errorf("Output should contain '%s'", exp)
		}
	}
}
