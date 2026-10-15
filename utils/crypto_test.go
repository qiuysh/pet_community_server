package utils

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Run("hash and verify", func(t *testing.T) {
		hash, err := HashPassword("password123")
		if err != nil {
			t.Fatalf("hash password failed: %v", err)
		}
		if !CheckPasswordHash("password123", hash) {
			t.Error("expected password to match hash")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		hash, _ := HashPassword("password123")
		if CheckPasswordHash("wrongpassword", hash) {
			t.Error("expected wrong password to not match")
		}
	})
}

func TestSHA256(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		hash1 := SHA256("hello")
		hash2 := SHA256("hello")
		if hash1 != hash2 {
			t.Error("expected SHA256 to be deterministic")
		}
	})

	t.Run("different input", func(t *testing.T) {
		hash1 := SHA256("hello")
		hash2 := SHA256("world")
		if hash1 == hash2 {
			t.Error("expected different hashes for different inputs")
		}
	})
}

func TestGenerateRandomString(t *testing.T) {
	t.Run("length", func(t *testing.T) {
		s, err := GenerateRandomString(16)
		if err != nil {
			t.Fatalf("generate random string failed: %v", err)
		}
		if len(s) != 16 {
			t.Errorf("expected length 16, got %d", len(s))
		}
	})

	t.Run("unique", func(t *testing.T) {
		s1, _ := GenerateRandomString(16)
		s2, _ := GenerateRandomString(16)
		if s1 == s2 {
			t.Error("expected random strings to be different")
		}
	})
}

func TestEncryptDecryptAES(t *testing.T) {
	key := "0123456789abcdef" // 16 bytes for AES-128

	t.Run("success", func(t *testing.T) {
		plaintext := "hello world"
		ciphertext, err := EncryptAES(plaintext, key)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}
		decrypted, err := DecryptAES(ciphertext, key)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}
		if decrypted != plaintext {
			t.Errorf("expected %q, got %q", plaintext, decrypted)
		}
	})

	t.Run("wrong key length", func(t *testing.T) {
		_, err := EncryptAES("hello", "short")
		if err == nil {
			t.Error("expected error for invalid key length")
		}
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		_, err := DecryptAES("c2hvcnQ=", key)
		if err == nil {
			t.Error("expected error for short ciphertext")
		}
	})
}

func TestEncryptDecrypt(t *testing.T) {
	InitCrypto("pet-community-secret-key")

	t.Run("success", func(t *testing.T) {
		plaintext := "sensitive data"
		ciphertext, err := Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}
		decrypted, err := Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}
		if decrypted != plaintext {
			t.Errorf("expected %q, got %q", plaintext, decrypted)
		}
	})

	t.Run("not initialized", func(t *testing.T) {
		// Temporarily clear secret key.
		oldKey := secretKey
		secretKey = nil
		defer func() { secretKey = oldKey }()

		_, err := Encrypt("data")
		if err == nil {
			t.Error("expected error when secret key not initialized")
		}
	})
}
