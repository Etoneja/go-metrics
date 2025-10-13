package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/etoneja/go-metrics/internal/common"
	"go.uber.org/zap"
)

func TestHashMiddleware_NoBody(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware("secret")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("Should pass through for no body")
	}
}

func TestHashMiddleware_NoHashKey(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware("")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("test"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("Should pass through for no hash key")
	}
}

func TestHashMiddleware_NoRequestHash(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware("secret")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("test"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("Should pass through for no request hash")
	}
}

func TestHashMiddleware_InvalidHash(t *testing.T) {
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware("secret")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid hash")
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("test"))
	req.Header.Set(common.HashHeaderKey, "invalid_hash")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("Expected 400 for invalid hash")
	}
}

func TestHashMiddleware_ValidHash(t *testing.T) {
	hashKey := "secret"
	data := []byte("test data")
	validHash := common.ComputeHash(hashKey, data)

	var handlerCalled bool
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware(hashKey)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if _, err := w.Write([]byte("response data")); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Set(common.HashHeaderKey, validHash)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("Handler should be called for valid hash")
	}
	if rr.Header().Get(common.HashHeaderKey) == "" {
		t.Error("Response should have hash header")
	}
}

func TestHashMiddleware_ResponseHash(t *testing.T) {
	hashKey := "secret"
	data := []byte("test data")
	validHash := common.ComputeHash(hashKey, data)
	responseData := []byte("response data")
	expectedResponseHash := common.ComputeHash(hashKey, responseData)

	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware(hashKey)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(responseData); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Set(common.HashHeaderKey, validHash)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	responseHash := rr.Header().Get(common.HashHeaderKey)
	if responseHash != expectedResponseHash {
		t.Errorf("Response hash mismatch. Expected %s, got %s", expectedResponseHash, responseHash)
	}
}

func TestHashMiddleware_EmptyResponse(t *testing.T) {
	hashKey := "secret"
	data := []byte("test data")
	validHash := common.ComputeHash(hashKey, data)

	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.HashMiddleware(hashKey)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest("POST", "/", bytes.NewReader(data))
	req.Header.Set(common.HashHeaderKey, validHash)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get(common.HashHeaderKey) != "" {
		t.Error("Should not set hash header for empty response")
	}
}
