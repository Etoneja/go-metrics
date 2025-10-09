package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestIsCompressibleContentType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"text/html", true},
		{"application/xml", false},
		{"text/plain", false},
		{"invalid", false},
		{"", false},
	}

	for _, test := range tests {
		result := isCompressibleContentType(test.contentType)
		if result != test.expected {
			t.Errorf("isCompressibleContentType(%q) = %v, expected %v", test.contentType, result, test.expected)
		}
	}
}

func TestGzipMiddleware_DecompressRequest(t *testing.T) {
	originalData := "test data"
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte(originalData))
	gz.Close()

	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.GzipMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != originalData {
			t.Errorf("Decompressed data mismatch. Expected '%s', got '%s'", originalData, string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
}

func TestGzipMiddleware_InvalidGzip(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.GzipMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid gzip")
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("invalid gzip data"))
	req.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("Expected 400 for invalid gzip")
	}
}

func TestGzipMiddleware_NoAcceptEncoding(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.GzipMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Response should not be gzipped without Accept-Encoding")
	}
}

func TestGzipMiddleware_NonCompressibleContentType(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.GzipMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png") // не сжимаемый тип
		if _, err := w.Write([]byte("test")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Response should not be gzipped for non-compressible content type")
	}
}

func TestGzipResponseWriter_Close(t *testing.T) {
	rr := httptest.NewRecorder()
	gzw := &gzipResponseWriter{ResponseWriter: rr}

	err := gzw.Close()
	if err != nil {
		t.Errorf("Close should not fail when gz is nil: %v", err)
	}

	gzw.compress = true
	gzw.Write([]byte("test"))
	err = gzw.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}
