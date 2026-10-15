package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

func encryptWithPublicKey(plaintext string, publicKey *rsa.PublicKey) (string, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func TestInitRSA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		err := InitRSA(2048)
		if err != nil {
			t.Fatalf("init rsa failed: %v", err)
		}
		if rsaPrivateKey == nil {
			t.Error("expected private key to be initialized")
		}
		if rsaPublicKey == nil {
			t.Error("expected public key to be initialized")
		}
	})
}

func TestGetPublicKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := InitRSA(2048); err != nil {
			t.Fatalf("init rsa failed: %v", err)
		}

		key := GetPublicKey()
		if key == "" {
			t.Error("expected non-empty public key")
		}

		// Verify it's a valid public key.
		data, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			t.Fatalf("failed to decode public key: %v", err)
		}
		_, err = x509.ParsePKIXPublicKey(data)
		if err != nil {
			t.Errorf("parsed public key invalid: %v", err)
		}
	})
}

func TestDecryptRSA(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if err := InitRSA(2048); err != nil {
			t.Fatalf("init rsa failed: %v", err)
		}

		plaintext := "secret message"
		ciphertext, err := encryptWithPublicKey(plaintext, rsaPublicKey)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		decrypted, err := DecryptRSA(ciphertext)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}
		if decrypted != plaintext {
			t.Errorf("expected %q, got %q", plaintext, decrypted)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		// Temporarily clear keys.
		oldPrivateKey := rsaPrivateKey
		oldPublicKey := rsaPublicKey
		rsaPrivateKey = nil
		rsaPublicKey = nil
		defer func() {
			rsaPrivateKey = oldPrivateKey
			rsaPublicKey = oldPublicKey
		}()

		_, err := DecryptRSA("ciphertext")
		if err == nil {
			t.Error("expected error when RSA not initialized")
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		if err := InitRSA(2048); err != nil {
			t.Fatalf("init rsa failed: %v", err)
		}

		_, err := DecryptRSA("not-valid-base64!!!")
		if err == nil {
			t.Error("expected error for invalid base64")
		}
	})
}
