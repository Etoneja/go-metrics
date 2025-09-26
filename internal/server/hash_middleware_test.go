package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestHashMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := &BaseMiddleware{logger: logger}
	hashKey := "test-key"

	handler := middleware.HashMiddleware(hashKey)

	if handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("test response"))
		if err != nil {
			t.Logf("Write failed: %v", err)
		}
	})

	wrappedHandler := handler(nextHandler)
	if wrappedHandler == nil {
		t.Error("Expected wrapped handler, got nil")
	}
}

func TestResponseRecorder_Write(t *testing.T) {
	recorder := &responseRecorder{
		ResponseWriter: httptest.NewRecorder(),
		body:           &bytes.Buffer{},
	}

	data := []byte("test data")
	n, err := recorder.Write(data)

	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	if recorder.body.String() != "test data" {
		t.Errorf("Expected body 'test data', got '%s'", recorder.body.String())
	}
}
