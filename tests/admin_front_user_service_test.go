package tests

import (
	"strings"
	"testing"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

// TestBanFrontUserInvalidatesSessions 封禁必须递增 Token 版本并清空设备，
// 否则已登录用户拿旧 Token 仍能继续操作。
func TestBanFrontUserInvalidatesSessions(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewAdminFrontUserService(repository.NewAdminRepository(db), userRepo)

	user := &model.User{Phone: "13800138011", Email: "victim@test.dev", Nickname: "victim", Status: 1}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if err := db.Create(&model.UserDevice{UserID: user.ID, DeviceID: "dev-1", DeviceType: "web"}).Error; err != nil {
		t.Fatalf("创建设备失败: %v", err)
	}
	before, err := userRepo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("读取用户失败: %v", err)
	}

	if err := srv.BanFrontUser(user.ID, "违规", 1, "127.0.0.1"); err != nil {
		t.Fatalf("封禁失败: %v", err)
	}

	after, err := userRepo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("读取用户失败: %v", err)
	}
	if after.Status != 2 {
		t.Fatalf("封禁后状态应为 2，实际 %d", after.Status)
	}
	if after.TokenVersion <= before.TokenVersion {
		t.Fatalf("封禁必须递增 Token 版本（原 %d，现 %d），否则旧 Token 继续可用", before.TokenVersion, after.TokenVersion)
	}

	var devices int64
	db.Model(&model.UserDevice{}).Where("user_id = ?", user.ID).Count(&devices)
	if devices != 0 {
		t.Fatalf("封禁应清空全部登录设备，剩余 %d 条", devices)
	}
}

// TestBanFrontUserSurfacesInvalidationFailure 强制下线失败时不能静默返回成功，
// 否则管理员会以为用户已被踢下线。
func TestBanFrontUserSurfacesInvalidationFailure(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewAdminFrontUserService(repository.NewAdminRepository(db), userRepo)

	user := &model.User{Phone: "13800138012", Email: "victim2@test.dev", Nickname: "victim2", Status: 1}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	// 模拟设备表不可用：清设备会直接报错
	if err := db.Migrator().DropTable(&model.UserDevice{}); err != nil {
		t.Fatalf("删除设备表失败: %v", err)
	}

	err := srv.BanFrontUser(user.ID, "违规", 1, "127.0.0.1")
	if err == nil {
		t.Fatal("强制下线失败时必须返回错误，不能静默成功")
	}
	if !strings.Contains(err.Error(), "强制下线") {
		t.Fatalf("错误信息应说明强制下线失败，实际: %v", err)
	}
}

// TestBatchBanFrontUsersHandlesUpdateError 批量封禁中个别用户写库失败时，
// 其余用户仍应被处理，同时把失败透出给调用方。
func TestBatchBanFrontUsersHandlesUpdateError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewAdminFrontUserService(repository.NewAdminRepository(db), userRepo)

	u1 := &model.User{Phone: "13800138013", Email: "b1@test.dev", Nickname: "b1", Status: 1}
	u2 := &model.User{Phone: "13800138014", Email: "b2@test.dev", Nickname: "b2", Status: 1}
	if err := db.Create(u1).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if err := db.Create(u2).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 设备表不可用 -> 每个用户的强制下线都会失败，但状态仍应落库
	if err := db.Migrator().DropTable(&model.UserDevice{}); err != nil {
		t.Fatalf("删除设备表失败: %v", err)
	}

	err := srv.BatchBanFrontUsers([]uint{u1.ID, u2.ID}, "批量违规", 1, "127.0.0.1")
	if err == nil {
		t.Fatal("强制下线失败时必须返回错误")
	}

	got1, _ := userRepo.GetByID(u1.ID)
	got2, _ := userRepo.GetByID(u2.ID)
	if got1.Status != 2 || got2.Status != 2 {
		t.Fatalf("批量封禁应把状态落库，实际 %d / %d", got1.Status, got2.Status)
	}

	var logCount int64
	db.Model(&model.AdminOperationLog{}).Where("action = ?", "batch_ban_front_users").Count(&logCount)
	if logCount != 1 {
		t.Fatalf("批量封禁应写一条审计日志，实际 %d 条", logCount)
	}
}
