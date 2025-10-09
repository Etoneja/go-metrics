package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/etoneja/go-metrics/internal/common"
	"go.uber.org/zap"
)

type mockHandler struct {
	receivedBody string
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	m.receivedBody = string(body)
	w.WriteHeader(http.StatusOK)
}

func TestDecryptMiddleware_NoPrivateKey(t *testing.T) {
	mockHandler := &mockHandler{}
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.DecryptMiddleware(nil)
	handler := middleware(mockHandler)

	req := httptest.NewRequest("POST", "/", strings.NewReader("test data"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if mockHandler.receivedBody != "test data" {
		t.Error("Body should be passed through without decryption")
	}
}

func TestDecryptMiddleware_NoEncryptionHeader(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	mockHandler := &mockHandler{}
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.DecryptMiddleware(privKey)
	handler := middleware(mockHandler)

	req := httptest.NewRequest("POST", "/", strings.NewReader("test data"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if mockHandler.receivedBody != "test data" {
		t.Error("Body should be passed through without decryption")
	}
}

func TestDecryptMiddleware_EmptyBody(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	mockHandler := &mockHandler{}
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.DecryptMiddleware(privKey)
	handler := middleware(mockHandler)

	req := httptest.NewRequest("POST", "/", strings.NewReader(""))
	req.Header.Set("X-Encrypted", "true")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if mockHandler.receivedBody != "" {
		t.Error("Empty body should be passed through")
	}
}

func TestDecryptMiddleware_ShortData(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.DecryptMiddleware(privKey)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for short data")
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("short"))
	req.Header.Set("X-Encrypted", "true")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("Expected 400 for short encrypted data")
	}
}

func TestDecryptMiddleware_Success(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	originalData := []byte("test data")

	encryptedData, err := common.EncryptHybrid(&privKey.PublicKey, originalData)
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	mockHandler := &mockHandler{}
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.DecryptMiddleware(privKey)
	handler := middleware(mockHandler)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(encryptedData))
	req.Header.Set("X-Encrypted", "true")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if mockHandler.receivedBody != string(originalData) {
		t.Errorf("Decrypted data doesn't match. Expected '%s', got '%s'", originalData, mockHandler.receivedBody)
	}
}

func TestDecryptMiddleware_DecryptAESError(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	bmw := BaseMiddleware{logger: zap.NewNop()}
	middleware := bmw.DecryptMiddleware(privKey)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called on decrypt error")
	}))

	encryptedAESKey := make([]byte, 256)
	_, err := rand.Read(encryptedAESKey)
	if err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}
	invalidPayload := []byte("invalid encrypted data")
	encryptedData := append(encryptedAESKey, invalidPayload...)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(encryptedData))
	req.Header.Set("X-Encrypted", "true")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("Expected 400 for decrypt error")
	}
}
