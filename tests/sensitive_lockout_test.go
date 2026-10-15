package tests

import (
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/utils"
)

func TestFilterSensitiveText(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	t.Cleanup(func() { testutil.SetGlobalDB(nil) })

	_ = db.Create(&model.SystemConfig{ConfigKey: "audit_sensitive_enabled", ConfigValue: "true"}).Error
	_ = db.Create(&model.SensitiveWord{Word: "违禁", ReplaceWith: "**", Status: 1, CreatedAt: time.Now()}).Error

	got := service.FilterSensitiveText("这是违禁内容")
	if got != "这是**内容" {
		t.Fatalf("expected replaced text, got %q", got)
	}
}

func TestAdminLoginLockout(t *testing.T) {
	db := testutil.SetupTestDB(t)
	adminRepo := repository.NewAdminRepository(db)
	srv := service.NewAdminAuthService(adminRepo)

	hash, err := utils.HashPassword("Pass1234")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	admin := &model.AdminUser{Username: "lockme", Password: hash, Role: "content_admin", Status: 1}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, _, err := srv.Login("lockme", "wrong", "127.0.0.1", "test")
		if err == nil {
			t.Fatalf("expected failure on attempt %d", i+1)
		}
	}
	_, _, err = srv.Login("lockme", "Pass1234", "127.0.0.1", "test")
	if err == nil || err.Error() != "登录失败次数过多，请30分钟后再试" {
		t.Fatalf("expected lockout, got %v", err)
	}
}
