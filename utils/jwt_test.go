package utils

import (
	"testing"

	"pet-community/config"
)

func initJWTConfig() {
	if config.Cfg == nil {
		config.Cfg = &config.Config{}
	}
	config.Cfg.JWT.Secret = "test-secret-key-for-jwt"
}

func TestGenerateToken(t *testing.T) {
	initJWTConfig()

	t.Run("success", func(t *testing.T) {
		token, err := GenerateToken(1, "admin", "admin")
		if err != nil {
			t.Fatalf("generate token failed: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}
	})
}

func TestParseToken(t *testing.T) {
	initJWTConfig()

	t.Run("valid token", func(t *testing.T) {
		token, err := GenerateToken(1, "admin", "admin")
		if err != nil {
			t.Fatalf("generate token failed: %v", err)
		}

		claims, err := ParseToken(token)
		if err != nil {
			t.Fatalf("parse token failed: %v", err)
		}
		if claims.AdminID != 1 {
			t.Errorf("expected admin ID 1, got %d", claims.AdminID)
		}
		if claims.Username != "admin" {
			t.Errorf("expected username 'admin', got %s", claims.Username)
		}
		if claims.Role != "admin" {
			t.Errorf("expected role 'admin', got %s", claims.Role)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := ParseToken("invalid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		token, _ := GenerateToken(1, "admin", "admin")

		// Temporarily change secret.
		oldSecret := config.Cfg.JWT.Secret
		config.Cfg.JWT.Secret = "wrong-secret"
		defer func() { config.Cfg.JWT.Secret = oldSecret }()

		_, err := ParseToken(token)
		if err == nil {
			t.Error("expected error for token signed with different secret")
		}
	})
}
