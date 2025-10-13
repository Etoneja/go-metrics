package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type EncryptionError struct {
	Operation string
	Cause     error
}

func (e *EncryptionError) Error() string {
	return fmt.Sprintf("encryption error in %s: %v", e.Operation, e.Cause)
}

func (e *EncryptionError) Unwrap() error {
	return e.Cause
}

func LoadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	if keyPath == "" {
		return nil, nil
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaPriv, ok := priv.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
		return rsaPriv, nil
	}

	rsaPriv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return rsaPriv, nil
	}

	return nil, fmt.Errorf("failed to parse private key: unsupported format")
}

func LoadPublicKey(keyPath string) (*rsa.PublicKey, error) {
	if keyPath == "" {
		return nil, nil
	}

	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return x509.ParsePKCS1PublicKey(block.Bytes)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

func DecryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &EncryptionError{Operation: "create cipher", Cause: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &EncryptionError{Operation: "create GCM mode", Cause: err}
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, &EncryptionError{Operation: "validate ciphertext", Cause: fmt.Errorf("ciphertext too short: got %d bytes, need at least %d", len(data), nonceSize)}
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, &EncryptionError{Operation: "decrypt data", Cause: err}
	}

	return plaintext, nil
}

func EncryptHybrid(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	aesKey := make([]byte, 32)
	if _, err := rand.Read(aesKey); err != nil {
		return nil, &EncryptionError{Operation: "AES key generation", Cause: err}
	}

	encryptedData, err := encryptAES(aesKey, data)
	if err != nil {
		return nil, &EncryptionError{Operation: "AES data encryption", Cause: err}
	}

	encryptedAESKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		aesKey,
		nil,
	)
	if err != nil {
		return nil, &EncryptionError{Operation: "RSA key encryption", Cause: err}
	}

	return append(encryptedAESKey, encryptedData...), nil
}

func encryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, &EncryptionError{Operation: "create cipher", Cause: err}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &EncryptionError{Operation: "create GCM mode", Cause: err}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, &EncryptionError{Operation: "generate nonce", Cause: err}
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}
