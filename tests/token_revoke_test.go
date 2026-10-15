package tests

import (
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

func TestTokenRevokeOnLogoutAndBan(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	t.Cleanup(func() { testutil.SetGlobalDB(nil) })

	userRepo := repository.NewUserRepository(db)
	srv := service.NewUserService(userRepo)

	user := &model.User{
		Nickname: "tok", Phone: "13900001111", Password: "x", Status: 1, PrivacySettings: "{}",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	deviceID := "dev-1"
	_ = userRepo.AddDevice(&model.UserDevice{
		UserID: user.ID, DeviceID: deviceID, DeviceName: "web", DeviceType: "web",
		IP: "127.0.0.1", LastActiveAt: time.Now(),
	})

	access, refresh, err := service.GenerateUserToken(user.ID, 0, deviceID, true)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := service.ParseUserToken(access); err != nil {
		t.Fatalf("parse before logout: %v", err)
	}

	if err := srv.Logout(user.ID, deviceID); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.ParseUserToken(access); err == nil {
		t.Fatal("expected access token invalid after device logout")
	}
	if _, _, err := service.RefreshUserToken(refresh); err == nil {
		t.Fatal("expected refresh token invalid after device logout")
	}

	// 重新登录设备后封禁应作废版本
	_ = userRepo.AddDevice(&model.UserDevice{
		UserID: user.ID, DeviceID: deviceID, DeviceName: "web", DeviceType: "web",
		IP: "127.0.0.1", LastActiveAt: time.Now(),
	})
	access2, _, err := service.GenerateUserToken(user.ID, 0, deviceID, true)
	if err != nil {
		t.Fatalf("generate2: %v", err)
	}
	adminRepo := repository.NewAdminRepository(db)
	frontSvc := service.NewAdminFrontUserService(adminRepo, userRepo)
	if err := frontSvc.BanFrontUser(user.ID, "spam", 1, "127.0.0.1"); err != nil {
		t.Fatalf("ban: %v", err)
	}
	if _, err := service.ParseUserToken(access2); err == nil {
		t.Fatal("expected token invalid after ban")
	}
}

func TestFollowersCountExcludesSecret(t *testing.T) {
	db := testutil.SetupTestDB(t)
	followRepo := repository.NewFollowRepository(db)

	_ = followRepo.Create(&model.Follow{FollowerID: 2, FollowingID: 1, Status: 1, IsSecret: 0})
	_ = followRepo.Create(&model.Follow{FollowerID: 3, FollowingID: 1, Status: 1, IsSecret: 1})

	count, err := followRepo.GetFollowersCount(1)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected public followers 1, got %d", count)
	}
}
