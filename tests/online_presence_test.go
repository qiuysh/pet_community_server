package tests

import (
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

type fakePresence map[uint]bool

func (f fakePresence) IsOnline(userID uint) bool { return f[userID] }

// TestGetOnlineStatusUsesPresence 在线判定必须看 WebSocket 连接态：
// 已退出但 5 分钟内登录过的用户不得显示为「在线」。
func TestGetOnlineStatusUsesPresence(t *testing.T) {
	db := testutil.SetupTestDB(t)

	loginAt := time.Now().Add(-30 * time.Minute)
	user := &model.User{
		Phone:       "presence_phone",
		Email:       "presence@test.com",
		Nickname:    "Presence",
		Status:      1,
		LastLoginAt: &loginAt,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	srv := service.NewFollowService(
		repository.NewFollowRepository(db),
		repository.NewUserRepository(db),
		repository.NewNotificationRepository(db),
	)

	// 有连接 → 在线
	srv.SetPresence(fakePresence{user.ID: true})
	if status, err := srv.GetOnlineStatus(user.ID); err != nil || status != 1 {
		t.Fatalf("with connection: status=%d err=%v, want 1", status, err)
	}

	// 无连接但 30 分钟内登录 → 最近活跃（旧实现会误判为在线）
	srv.SetPresence(fakePresence{})
	if status, err := srv.GetOnlineStatus(user.ID); err != nil || status != 2 {
		t.Fatalf("without connection: status=%d err=%v, want 2", status, err)
	}

	// 超过 24 小时未登录 → 离线
	stale := time.Now().Add(-25 * time.Hour)
	if err := db.Model(user).Update("last_login_at", stale).Error; err != nil {
		t.Fatalf("update last_login_at: %v", err)
	}
	if status, err := srv.GetOnlineStatus(user.ID); err != nil || status != 0 {
		t.Fatalf("stale login: status=%d err=%v, want 0", status, err)
	}

	// 未注入连接态时不 panic，按活跃时间降级
	srv.SetPresence(nil)
	if status, err := srv.GetOnlineStatus(user.ID); err != nil || status != 0 {
		t.Fatalf("nil presence: status=%d err=%v, want 0", status, err)
	}
}
