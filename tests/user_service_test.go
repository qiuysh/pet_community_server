package tests

import (
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
	"pet-community/utils"
)

func TestUserServiceRegister(t *testing.T) {
	db := testutil.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	srv := service.NewUserService(userRepo)

	code := &model.VerificationCode{
		Account:   "13800138000",
		Code:      "123456",
		Type:      "register",
		ExpiresAt: time.Now().Add(5 * time.Minute),
		IP:        "127.0.0.1",
	}
	if err := db.Create(code).Error; err != nil {
		t.Fatalf("failed to create verification code: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := &types.UserRegisterRequest{
			Phone:    "13800138000",
			Password: "Pass1234",
			Code:     "123456",
		}
		info, err := srv.Register(req, "127.0.0.1")
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("user ID should not be zero")
		}
	})

	t.Run("phone already exists", func(t *testing.T) {
		newCode := &model.VerificationCode{
			Account:   "13800138000",
			Code:      "654321",
			Type:      "register",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			IP:        "127.0.0.1",
		}
		db.Create(newCode)
		req := &types.UserRegisterRequest{
			Phone:    "13800138000",
			Password: "Pass1234",
			Code:     "654321",
		}
		_, err := srv.Register(req, "127.0.0.1")
		if err == nil {
			t.Error("expected error for existing phone")
		}
	})

	t.Run("invalid code", func(t *testing.T) {
		req := &types.UserRegisterRequest{
			Phone:    "13800138001",
			Password: "Pass1234",
			Code:     "123456",
		}
		_, err := srv.Register(req, "127.0.0.1")
		if err == nil {
			t.Error("expected error for invalid code")
		}
	})
}

func TestUserServiceLogin(t *testing.T) {
	db := testutil.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	srv := service.NewUserService(userRepo)

	hashedPassword, _ := utils.HashPassword("Pass1234")
	user := &model.User{
		Phone:    "13800138000",
		Password: string(hashedPassword),
		Nickname: "user",
		Status:   1,
	}
	db.Create(user)

	t.Run("password correct", func(t *testing.T) {
		req := &types.UserLoginRequest{
			Account:   "13800138000",
			Password:  "Pass1234",
			LoginType: "password",
		}
		resp, err := srv.Login(req, "127.0.0.1", "test-agent")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}
		if resp.Token == "" {
			t.Error("token should not be empty")
		}
	})

	t.Run("password wrong", func(t *testing.T) {
		req := &types.UserLoginRequest{
			Account:   "13800138000",
			Password:  "wrongpassword",
			LoginType: "password",
		}
		_, err := srv.Login(req, "127.0.0.1", "test-agent")
		if err == nil {
			t.Error("expected error for wrong password")
		}
	})

	t.Run("disabled user", func(t *testing.T) {
		disabled := &model.User{
			Phone:    "13800138001",
			Email:    "disabled@test.com",
			Password: string(hashedPassword),
			Nickname: "disabled",
			Status:   0,
		}
		db.Create(disabled)
		req := &types.UserLoginRequest{
			Account:   "13800138001",
			Password:  "Pass1234",
			LoginType: "password",
		}
		_, err := srv.Login(req, "127.0.0.1", "test-agent")
		if err == nil {
			t.Error("expected error for disabled user")
		}
	})
}

func TestUserServiceChangePassword(t *testing.T) {
	db := testutil.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	srv := service.NewUserService(userRepo)

	hashedPassword, _ := utils.HashPassword("Oldpass123")
	user := &model.User{
		Phone:    "13800138000",
		Password: string(hashedPassword),
		Nickname: "user",
		Status:   1,
	}
	db.Create(user)

	t.Run("old password correct", func(t *testing.T) {
		err := srv.ChangePassword(user.ID, &types.PasswordChangeRequest{
			OldPassword: "Oldpass123",
			NewPassword: "Newpass123",
		})
		if err != nil {
			t.Fatalf("change password failed: %v", err)
		}
	})

	t.Run("old password wrong", func(t *testing.T) {
		err := srv.ChangePassword(user.ID, &types.PasswordChangeRequest{
			OldPassword: "wrongpass",
			NewPassword: "Newpass123",
		})
		if err == nil {
			t.Error("expected error for wrong old password")
		}
	})
}

func TestUserServiceUpdateProfile(t *testing.T) {
	db := testutil.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	srv := service.NewUserService(userRepo)

	user := &model.User{Phone: "13800138000", Nickname: "oldname", Status: 1}
	db.Create(user)

	gender := 1
	info, err := srv.UpdateProfile(user.ID, &types.UserProfileUpdateRequest{
		Nickname:  "newname",
		Signature: "hello",
		Gender:    &gender,
	})
	if err != nil {
		t.Fatalf("update profile failed: %v", err)
	}
	if info.Nickname != "newname" {
		t.Errorf("expected nickname newname, got %s", info.Nickname)
	}
	if info.Signature != "hello" {
		t.Errorf("expected signature hello, got %s", info.Signature)
	}
	if info.Gender != 1 {
		t.Errorf("expected gender 1, got %d", info.Gender)
	}
}

func TestUserServiceSendCode(t *testing.T) {
	db := testutil.SetupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	srv := service.NewUserService(userRepo)

	t.Run("success", func(t *testing.T) {
		expiresIn, debugCode, err := srv.SendCode("13800138000", "register", "127.0.0.1")
		if debugCode == "" {
			t.Error("expected debugCode in dev mode")
		}
		if err != nil {
			t.Fatalf("send code failed: %v", err)
		}
		if expiresIn != 300 {
			t.Errorf("expected expiresIn 300, got %d", expiresIn)
		}
	})

	t.Run("frequency limit", func(t *testing.T) {
		recentCode := &model.VerificationCode{
			Account:   "13800138001",
			Code:      "123456",
			Type:      "register",
			ExpiresAt: time.Now().Add(5 * time.Minute),
			IP:        "127.0.0.1",
		}
		db.Create(recentCode)

		_, _, err := srv.SendCode("13800138001", "register", "127.0.0.1")
		if err == nil {
			t.Error("expected error for frequency limit")
		}
	})
}
