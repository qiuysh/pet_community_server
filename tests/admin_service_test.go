package tests

import (
	"testing"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/utils"
)

func TestAdminServiceLogin(t *testing.T) {
	db := testutil.SetupTestDB(t)

	adminRepo := repository.NewAdminRepository(db)
	srv := service.NewAdminAuthService(adminRepo)

	hashedPassword, err := utils.HashPassword("123456")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	admin := &model.AdminUser{
		Username: "admin",
		Password: hashedPassword,
		Email:    "admin@example.com",
		Role:     "admin",
		Status:   1,
	}
	if err := db.Create(admin).Error; err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		token, loggedIn, err := srv.Login("admin", "123456", "127.0.0.1", "")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}
		if token == "" {
			t.Error("token should not be empty")
		}
		if loggedIn.ID != admin.ID {
			t.Error("admin ID mismatch")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, _, err := srv.Login("admin", "wrongpassword", "127.0.0.1", "")
		if err == nil {
			t.Error("expected error for wrong password")
		}
	})

	t.Run("disabled admin", func(t *testing.T) {
		// GORM treats zero value status as the model default (1), so create as
		// active first and then explicitly update the status column to 0.
		disabled := &model.AdminUser{
			Username: "disabled",
			Password: hashedPassword,
			Email:    "disabled@example.com",
			Status:   1,
		}
		if err := db.Create(disabled).Error; err != nil {
			t.Fatalf("failed to create disabled admin: %v", err)
		}
		if err := db.Model(disabled).Update("status", 0).Error; err != nil {
			t.Fatalf("failed to disable admin: %v", err)
		}

		_, _, err := srv.Login("disabled", "123456", "127.0.0.1", "")
		if err == nil {
			t.Error("expected error for disabled admin")
		}
	})
}

func TestAdminServiceCreateAdmin(t *testing.T) {
	db := testutil.SetupTestDB(t)

	adminRepo := repository.NewAdminRepository(db)
	srv := service.NewAdminAuthService(adminRepo)

	_, err := srv.CreateAdmin("admin", "123456", "admin@example.com", "admin", 1)
	if err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	t.Run("duplicate username", func(t *testing.T) {
		_, err := srv.CreateAdmin("admin", "123456", "admin2@example.com", "admin", 1)
		if err == nil {
			t.Error("expected error for duplicate username")
		}
	})
}

func TestAdminServiceListAdminsKeyword(t *testing.T) {
	db := testutil.SetupTestDB(t)

	adminRepo := repository.NewAdminRepository(db)
	srv := service.NewAdminAuthService(adminRepo)

	if err := db.Create(&model.AdminUser{Username: "alice", Password: "hash", Email: "alice@example.com", Status: 1}).Error; err != nil {
		t.Fatalf("failed to create alice: %v", err)
	}
	if err := db.Create(&model.AdminUser{Username: "bob", Password: "hash", Email: "bob@pet.com", Status: 1}).Error; err != nil {
		t.Fatalf("failed to create bob: %v", err)
	}

	t.Run("search by username", func(t *testing.T) {
		list, total, err := srv.ListAdmins("ali", 1, 20)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if total != 1 || len(list) != 1 || list[0].Username != "alice" {
			t.Fatalf("expected alice only, got total=%d list=%v", total, list)
		}
	})

	t.Run("search by email", func(t *testing.T) {
		list, total, err := srv.ListAdmins("pet.com", 1, 20)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if total != 1 || len(list) != 1 || list[0].Username != "bob" {
			t.Fatalf("expected bob only, got total=%d list=%v", total, list)
		}
	})

	t.Run("empty keyword returns all", func(t *testing.T) {
		_, total, err := srv.ListAdmins("", 1, 20)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if total != 2 {
			t.Fatalf("expected 2 admins, got %d", total)
		}
	})
}

func TestAdminServiceUpdateAdminStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)

	adminRepo := repository.NewAdminRepository(db)
	srv := service.NewAdminAuthService(adminRepo)

	admin := &model.AdminUser{Username: "admin", Password: "hash", Status: 1}
	db.Create(admin)

	t.Run("cannot disable self", func(t *testing.T) {
		err := srv.UpdateAdminStatus(admin.ID, admin.ID, 0)
		if err == nil {
			t.Error("expected error when disabling self")
		}
	})

	t.Run("disable other", func(t *testing.T) {
		other := &model.AdminUser{Username: "other", Password: "hash", Status: 1}
		db.Create(other)
		err := srv.UpdateAdminStatus(other.ID, admin.ID, 0)
		if err != nil {
			t.Fatalf("failed to disable other admin: %v", err)
		}
	})
}

func TestAdminServiceBanFrontUser(t *testing.T) {
	db := testutil.SetupTestDB(t)

	adminRepo := repository.NewAdminRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewAdminFrontUserService(adminRepo, userRepo)

	user := &model.User{Phone: "13800138000", Nickname: "user", Status: 1}
	db.Create(user)

	t.Run("ban user", func(t *testing.T) {
		err := srv.BanFrontUser(user.ID, "test reason", 1, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed to ban user: %v", err)
		}

		banned, _ := userRepo.GetByID(user.ID)
		if banned.Status != 2 {
			t.Errorf("expected status 2, got %d", banned.Status)
		}
	})

	t.Run("ban already banned user", func(t *testing.T) {
		err := srv.BanFrontUser(user.ID, "test reason", 1, "127.0.0.1")
		if err == nil {
			t.Error("expected error for already banned user")
		}
	})

	t.Run("unban user", func(t *testing.T) {
		err := srv.UnbanFrontUser(user.ID, 1, "127.0.0.1")
		if err != nil {
			t.Fatalf("failed to unban user: %v", err)
		}

		unbanned, _ := userRepo.GetByID(user.ID)
		if unbanned.Status != 1 {
			t.Errorf("expected status 1, got %d", unbanned.Status)
		}
	})
}

func TestAdminServiceUpdateSiteConfig(t *testing.T) {
	db := testutil.SetupTestDB(t)

	adminRepo := repository.NewAdminRepository(db)
	srv := service.NewAdminConfigService(adminRepo)

	configs := map[string]string{
		"site_name": "Pet Community Test",
	}

	err := srv.UpdateSiteConfig(configs, 1, "127.0.0.1")
	if err != nil {
		t.Fatalf("failed to update site config: %v", err)
	}

	var logCount int64
	db.Model(&model.AdminOperationLog{}).Count(&logCount)
	if logCount != 1 {
		t.Errorf("expected 1 operation log, got %d", logCount)
	}
}
