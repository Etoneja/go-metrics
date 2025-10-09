package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

func GenerateECPrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func TestLoadPrivateKey_EmptyPath(t *testing.T) {
	key, err := LoadPrivateKey("")
	if err != nil {
		t.Errorf("Expected no error for empty path, got %v", err)
	}
	if key != nil {
		t.Error("Expected nil key for empty path")
	}
}

func TestLoadPrivateKey_FileNotExists(t *testing.T) {
	_, err := LoadPrivateKey("/nonexistent/private.key")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadPublicKey_EmptyPath(t *testing.T) {
	key, err := LoadPublicKey("")
	if err != nil {
		t.Errorf("Expected no error for empty path, got %v", err)
	}
	if key != nil {
		t.Error("Expected nil key for empty path")
	}
}

func TestLoadPublicKey_FileNotExists(t *testing.T) {
	_, err := LoadPublicKey("/nonexistent/public.key")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestEncryptDecryptAES(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	data := []byte("test data")

	encrypted, err := encryptAES(key, data)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptAES(key, encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Errorf("Decrypted data doesn't match original")
	}
}

func TestEncryptDecryptHybrid(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	data := []byte("test hybrid encryption data")

	encrypted, err := EncryptHybrid(&privateKey.PublicKey, data)
	if err != nil {
		t.Fatalf("Hybrid encryption failed: %v", err)
	}

	encryptedAESKey := encrypted[:256]
	encryptedData := encrypted[256:]

	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedAESKey, nil)
	if err != nil {
		t.Fatalf("RSA decryption failed: %v", err)
	}

	decrypted, err := DecryptAES(aesKey, encryptedData)
	if err != nil {
		t.Fatalf("AES decryption failed: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Errorf("Decrypted data doesn't match original")
	}
}

func TestDecryptAES_ShortData(t *testing.T) {
	key := make([]byte, 32)
	shortData := make([]byte, 10)

	_, err := DecryptAES(key, shortData)
	if err == nil {
		t.Error("Expected error for short ciphertext")
	}
}

func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "invalid_pem*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.WriteString("invalid pem data")
	tmpfile.Close()

	_, err = LoadPrivateKey(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}
}

func TestLoadPrivateKey_UnsupportedFormat(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "unsupported*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte("invalid key data"),
	}
	pemData := pem.EncodeToMemory(block)
	os.WriteFile(tmpfile.Name(), pemData, 0644)

	_, err = LoadPrivateKey(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestLoadPrivateKey_NotRSAKey(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "not_rsa*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	priv, err := GenerateECPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	privBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatal(err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	pemData := pem.EncodeToMemory(block)
	os.WriteFile(tmpfile.Name(), pemData, 0644)

	_, err = LoadPrivateKey(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for non-RSA key")
	}
}

func TestLoadPublicKey_InvalidPEM(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "invalid_pem_public*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.WriteString("invalid pem data")
	tmpfile.Close()

	_, err = LoadPublicKey(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid PEM")
	}
}

func TestLoadPublicKey_UnsupportedFormat(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "unsupported_public*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: []byte("invalid key data"),
	}
	pemData := pem.EncodeToMemory(block)
	os.WriteFile(tmpfile.Name(), pemData, 0644)

	_, err = LoadPublicKey(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for unsupported public key format")
	}
}

func TestLoadPublicKey_NotRSAKey(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "not_rsa_public*.key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}
	pemData := pem.EncodeToMemory(block)
	os.WriteFile(tmpfile.Name(), pemData, 0644)

	_, err = LoadPublicKey(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for non-RSA public key")
	}
}
