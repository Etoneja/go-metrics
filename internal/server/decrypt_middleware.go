package server

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"net/http"

	"github.com/etoneja/go-metrics/internal/common"
	"go.uber.org/zap"
)

func (bmw *BaseMiddleware) DecryptMiddleware(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if privateKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("X-Encrypted") != "true" {
				next.ServeHTTP(w, r)
				return
			}

			encryptedData, err := io.ReadAll(r.Body)
			if err != nil {
				bmw.logger.Error("Failed to read encrypted body", zap.Error(err))
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer func() {
				if err = r.Body.Close(); err != nil {
					bmw.logger.Warn("Failed to close request body", zap.Error(err))
				}
			}()

			if len(encryptedData) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			keySize := 256
			if len(encryptedData) < keySize {
				bmw.logger.Error("Encrypted data too short")
				http.Error(w, "Invalid encrypted data", http.StatusBadRequest)
				return
			}

			encryptedAESKey := encryptedData[:keySize]
			encryptedPayload := encryptedData[keySize:]

			aesKey, err := rsa.DecryptOAEP(
				sha256.New(),
				rand.Reader,
				privateKey,
				encryptedAESKey,
				nil,
			)
			if err != nil {
				bmw.logger.Error("Failed to decrypt AES key", zap.Error(err))
				http.Error(w, "Failed to decrypt AES key", http.StatusBadRequest)
				return
			}

			decryptedData, err := common.DecryptAES(aesKey, encryptedPayload)
			if err != nil {
				bmw.logger.Error("Failed to decrypt data with AES", zap.Error(err))
				http.Error(w, "Failed to decrypt data", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			r.ContentLength = int64(len(decryptedData))

			bmw.logger.Debug("Request decrypted successfully",
				zap.Int("encrypted_size", len(encryptedData)),
				zap.Int("decrypted_size", len(decryptedData)),
			)

			next.ServeHTTP(w, r)
		})
	}
}
