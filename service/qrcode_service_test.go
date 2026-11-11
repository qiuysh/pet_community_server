package service

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestQRCodeServiceGenerateAndValidate(t *testing.T) {
	srv := NewQRCodeServiceWithSecret("unit-test-qrcode-secret")

	code, expires := srv.GenerateQRCode(42)
	if code == "" || expires == "" {
		t.Fatalf("expected non-empty qrcode and expires")
	}
	parts := strings.Split(code, "-")
	if len(parts) != 3 {
		t.Fatalf("expected 3 parts, got %d (%q)", len(parts), code)
	}

	activityID, err := srv.ValidateQRCode(code)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if activityID != 42 {
		t.Fatalf("expected activity 42, got %d", activityID)
	}
}

func TestQRCodeServiceRejectsForgedSignature(t *testing.T) {
	srv := NewQRCodeServiceWithSecret("unit-test-qrcode-secret")
	code, _ := srv.GenerateQRCode(7)
	parts := strings.Split(code, "-")
	parts[2] = "0000000000000000"
	forged := strings.Join(parts, "-")

	if _, err := srv.ValidateQRCode(forged); err == nil {
		t.Fatal("expected forged signature to fail")
	}
}

func TestQRCodeServiceRejectsExpired(t *testing.T) {
	srv := NewQRCodeServiceWithSecret("unit-test-qrcode-secret")
	expiredAt := time.Now().Add(-time.Minute).Unix()
	data := fmt.Sprintf("9-%d", expiredAt)
	sig := srv.generateSignature(data)
	code := fmt.Sprintf("9-%d-%s", expiredAt, sig)

	if _, err := srv.ValidateQRCode(code); err == nil {
		t.Fatal("expected expired code to fail")
	}
}

func TestQRCodeServiceRejectsCrossActivityReuse(t *testing.T) {
	srv := NewQRCodeServiceWithSecret("unit-test-qrcode-secret")
	code, _ := srv.GenerateQRCode(11)
	parts := strings.Split(code, "-")
	// 改活动 ID 但保留原签名 → 签名应不匹配
	parts[0] = "99"
	reused := strings.Join(parts, "-")

	if _, err := srv.ValidateQRCode(reused); err == nil {
		t.Fatal("expected cross-activity reuse to fail")
	}
}

func TestQRCodeServiceRejectsMalformed(t *testing.T) {
	srv := NewQRCodeServiceWithSecret("unit-test-qrcode-secret")
	if _, err := srv.ValidateQRCode("not-a-code"); err == nil {
		t.Fatal("expected malformed code to fail")
	}
}
