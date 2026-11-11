package service

import (
	"testing"
	"time"
)

func TestGenerateAdminCaptchaAndVerifyOnce(t *testing.T) {
	id, img, err := GenerateAdminCaptcha()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if id == "" || img == "" {
		t.Fatal("expected non-empty id and image")
	}

	answer := adminCaptchaStore.Get(id, false)
	if answer == "" {
		t.Fatal("expected stored answer")
	}

	if err := VerifyAdminCaptcha(id, answer); err != nil {
		t.Fatalf("verify once: %v", err)
	}
	// 一次性：再次校验应失败
	if err := VerifyAdminCaptcha(id, answer); err == nil {
		t.Fatal("expected second verify to fail")
	}
}

func TestAdminCaptchaExpires(t *testing.T) {
	store := newTTLCaptchaStore()
	_ = store.Set("exp-id", "ABCD")
	store.mu.Lock()
	entry := store.data["exp-id"]
	entry.expiresAt = time.Now().Add(-time.Second)
	store.data["exp-id"] = entry
	store.mu.Unlock()

	if store.Verify("exp-id", "ABCD", true) {
		t.Fatal("expired captcha should not verify")
	}
}

func TestAdminCaptchaEmptyInvalid(t *testing.T) {
	if err := VerifyAdminCaptcha("", "ABCD"); err == nil {
		t.Fatal("empty id should fail")
	}
	if err := VerifyAdminCaptcha("x", ""); err == nil {
		t.Fatal("empty code should fail")
	}
}
