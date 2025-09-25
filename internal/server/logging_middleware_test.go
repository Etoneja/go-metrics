package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestLoggerMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := &BaseMiddleware{logger: logger}

	handler := middleware.LoggerMiddleware()
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

func TestLoggingResponseWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := &loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	// Test WriteHeader
	lw.WriteHeader(http.StatusNotFound)
	if responseData.status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, responseData.status)
	}

	// Test Write
	data := []byte("test data")
	n, err := lw.Write(data)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}
	if responseData.size != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), responseData.size)
	}
}

func TestLoggingResponseWriter_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := &BaseMiddleware{logger: logger}

	handler := middleware.LoggerMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("test response"))
		if err != nil {
			t.Logf("Write failed: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
	if recorder.Body.String() != "test response" {
		t.Errorf("Expected body 'test response', got '%s'", recorder.Body.String())
	}
}
