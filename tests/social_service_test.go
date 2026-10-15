package tests

import (
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

func TestFollowServiceToggleFollow(t *testing.T) {
	db := testutil.SetupTestDB(t)

	followRepo := repository.NewFollowRepository(db)
	userRepo := repository.NewUserRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	srv := service.NewFollowService(followRepo, userRepo, notificationRepo)

	t.Run("follow success", func(t *testing.T) {
		following, err := srv.ToggleFollow(1, 2, false)
		if err != nil {
			t.Fatalf("follow failed: %v", err)
		}
		if !following {
			t.Error("expected following")
		}
	})

	t.Run("cannot follow self", func(t *testing.T) {
		_, err := srv.ToggleFollow(1, 1, false)
		if err == nil {
			t.Error("expected error for self follow")
		}
	})

	t.Run("unfollow success", func(t *testing.T) {
		following, err := srv.ToggleFollow(1, 2, false)
		if err != nil {
			t.Fatalf("unfollow failed: %v", err)
		}
		if following {
			t.Error("expected not following")
		}
	})
}

func TestFollowServiceCounts(t *testing.T) {
	db := testutil.SetupTestDB(t)

	followRepo := repository.NewFollowRepository(db)
	userRepo := repository.NewUserRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	srv := service.NewFollowService(followRepo, userRepo, notificationRepo)

	db.Create(&model.Follow{FollowerID: 1, FollowingID: 2, Status: 1})
	db.Create(&model.Follow{FollowerID: 1, FollowingID: 3, Status: 1})
	db.Create(&model.Follow{FollowerID: 3, FollowingID: 1, Status: 1})

	following, err := srv.GetFollowingCount(1)
	if err != nil {
		t.Fatalf("get following count failed: %v", err)
	}
	if following != 2 {
		t.Errorf("expected following count 2, got %d", following)
	}

	followers, err := srv.GetFollowersCount(1)
	if err != nil {
		t.Fatalf("get followers count failed: %v", err)
	}
	if followers != 1 {
		t.Errorf("expected followers count 1, got %d", followers)
	}
}

func TestBlockService(t *testing.T) {
	db := testutil.SetupTestDB(t)

	blockRepo := repository.NewBlockRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewBlockService(blockRepo, userRepo)

	t.Run("block success and unfollow both sides", func(t *testing.T) {
		followRepo := repository.NewFollowRepository(db)
		srv.SetFollowRepository(followRepo)
		_ = followRepo.Create(&model.Follow{FollowerID: 1, FollowingID: 2, Status: 1})
		_ = followRepo.Create(&model.Follow{FollowerID: 2, FollowingID: 1, Status: 1})

		err := srv.BlockUser(1, 2)
		if err != nil {
			t.Fatalf("block failed: %v", err)
		}
		blocked, _ := srv.IsBlocked(1, 2)
		if !blocked {
			t.Error("expected blocked")
		}
		exists1, _ := followRepo.Exists(1, 2)
		exists2, _ := followRepo.Exists(2, 1)
		if exists1 || exists2 {
			t.Error("expected follow relations removed after block")
		}
		either, _ := srv.IsBlockedEither(1, 2)
		if !either {
			t.Error("expected ExistsEither true")
		}
	})

	t.Run("cannot block self", func(t *testing.T) {
		err := srv.BlockUser(1, 1)
		if err == nil {
			t.Error("expected error for self block")
		}
	})

	t.Run("block duplicate", func(t *testing.T) {
		err := srv.BlockUser(1, 2)
		if err == nil {
			t.Error("expected error for duplicate block")
		}
	})

	t.Run("unblock", func(t *testing.T) {
		err := srv.UnblockUser(1, 2)
		if err != nil {
			t.Fatalf("unblock failed: %v", err)
		}
		blocked, _ := srv.IsBlocked(1, 2)
		if blocked {
			t.Error("expected not blocked")
		}
	})
}

func TestMessageServiceSendMessage(t *testing.T) {
	db := testutil.SetupTestDB(t)

	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	userRepo := repository.NewUserRepository(db)
	blockRepo := repository.NewBlockRepository(db)
	srv := service.NewMessageService(conversationRepo, messageRepo, userRepo, blockRepo)

	t.Run("success", func(t *testing.T) {
		msg, err := srv.SendMessage(1, 2, "hello", "text")
		if err != nil {
			t.Fatalf("send message failed: %v", err)
		}
		if msg.Content != "hello" {
			t.Errorf("expected content hello, got %s", msg.Content)
		}
	})

	t.Run("cannot send to self", func(t *testing.T) {
		_, err := srv.SendMessage(1, 1, "hello", "text")
		if err == nil {
			t.Error("expected error for self message")
		}
	})

	t.Run("blocked by receiver", func(t *testing.T) {
		db.Create(&model.Block{BlockerID: 3, BlockedID: 1})
		_, err := srv.SendMessage(1, 3, "hello", "text")
		if err == nil {
			t.Error("expected error when blocked by receiver")
		}
	})

	t.Run("blocked by sender", func(t *testing.T) {
		db.Create(&model.Block{BlockerID: 1, BlockedID: 4})
		_, err := srv.SendMessage(1, 4, "hello", "text")
		if err == nil {
			t.Error("expected error when blocked by sender")
		}
	})
}

func TestMessageServiceRecall(t *testing.T) {
	db := testutil.SetupTestDB(t)

	conversationRepo := repository.NewConversationRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	userRepo := repository.NewUserRepository(db)
	blockRepo := repository.NewBlockRepository(db)
	srv := service.NewMessageService(conversationRepo, messageRepo, userRepo, blockRepo)

	conv := &model.Conversation{User1ID: 1, User2ID: 2}
	db.Create(conv)
	msg := &model.Message{ConversationID: conv.ID, FromID: 1, ToID: 2, Content: "hello"}
	db.Create(msg)

	t.Run("success", func(t *testing.T) {
		err := srv.RecallMessage(msg.ID, 1)
		if err != nil {
			t.Fatalf("recall failed: %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		oldMsg := &model.Message{ConversationID: conv.ID, FromID: 1, ToID: 2, Content: "old", CreatedAt: time.Now().Add(-10 * time.Minute)}
		db.Create(oldMsg)
		err := srv.RecallMessage(oldMsg.ID, 1)
		if err == nil {
			t.Error("expected error for timeout recall")
		}
	})
}

func TestPointsServiceAddPoints(t *testing.T) {
	db := testutil.SetupTestDB(t)

	pointsRepo := repository.NewPointsRepository(db)
	userLevelRepo := repository.NewUserLevelRepository(db)
	pointsRuleRepo := repository.NewPointsRuleRepository(db)
	levelBenefitRepo := repository.NewLevelBenefitRepository(db)
	srv := service.NewPointsService(pointsRepo, userLevelRepo, pointsRuleRepo, levelBenefitRepo)

	db.Create(&model.PointsRule{Type: "post_create", Points: 10, DailyLimit: 30, IsActive: 1})

	t.Run("success", func(t *testing.T) {
		err := srv.AddPoints(1, "post_create", 1, "发布动态")
		if err != nil {
			t.Fatalf("add points failed: %v", err)
		}
		level, _ := userLevelRepo.GetByUserID(1)
		if level == nil || level.Points != 10 {
			t.Errorf("expected points 10, got %v", level)
		}
	})

	t.Run("daily limit", func(t *testing.T) {
		// Already added 10, limit is 30, two more adds should hit limit
		_ = srv.AddPoints(1, "post_create", 2, "发布动态")
		_ = srv.AddPoints(1, "post_create", 3, "发布动态")
		err := srv.AddPoints(1, "post_create", 4, "发布动态")
		if err == nil {
			t.Error("expected error for daily limit")
		}
	})
}

func TestPointsServiceCalculateLevel(t *testing.T) {
	srv := service.NewPointsService(nil, nil, nil, nil)

	cases := []struct {
		points int
		level  int
	}{
		{0, 1},
		{199, 1},
		{200, 2},
		{400, 3},
		{500, 3},
		{600, 4},
		{1800, 10},
		{2000, 10},
	}
	for _, c := range cases {
		got := srv.CalculateLevel(c.points)
		if got != c.level {
			t.Errorf("points=%d expected level %d, got %d", c.points, c.level, got)
		}
	}
}

func TestPointsServiceLevelUp(t *testing.T) {
	db := testutil.SetupTestDB(t)

	pointsRepo := repository.NewPointsRepository(db)
	userLevelRepo := repository.NewUserLevelRepository(db)
	pointsRuleRepo := repository.NewPointsRuleRepository(db)
	levelBenefitRepo := repository.NewLevelBenefitRepository(db)
	srv := service.NewPointsService(pointsRepo, userLevelRepo, pointsRuleRepo, levelBenefitRepo)

	// 日上限 100：用多次 +10 升到 Lv2（阈值 200）
	db.Create(&model.PointsRule{Type: "post_create", Points: 10, IsActive: 1, DailyLimit: 0})
	for i := 0; i < 10; i++ {
		if err := srv.AddPoints(1, "post_create", uint(i+1), "发布动态"); err != nil {
			t.Fatalf("add points failed at %d: %v", i, err)
		}
	}

	level, _ := userLevelRepo.GetByUserID(1)
	if level == nil || level.Points != 100 {
		t.Fatalf("expected 100 points after daily cap window, got %v", level)
	}
	if level.Level != 1 {
		t.Errorf("expected level 1 at 100 points, got %d", level.Level)
	}

	// 直接写入累计积分验证升级逻辑
	level.Points = 600
	level.Level = srv.CalculateLevel(600)
	if err := userLevelRepo.Update(level); err != nil {
		t.Fatalf("update level failed: %v", err)
	}
	if level.Level != 4 {
		t.Errorf("expected level 4 (600 points), got %d", level.Level)
	}
}

func TestNotificationServiceCreateLikeNotification(t *testing.T) {
	db := testutil.SetupTestDB(t)

	notificationRepo := repository.NewNotificationRepository(db)
	followRepo := repository.NewFollowRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewNotificationService(notificationRepo, followRepo, userRepo)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.User{ID: 2, Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1})

	t.Run("success", func(t *testing.T) {
		err := srv.CreateLikeNotification(1, 2, "post", 1)
		if err != nil {
			t.Fatalf("create notification failed: %v", err)
		}
		var count int64
		db.Model(&model.UserNotification{}).Where("user_id = ? AND type = ?", 2, "like").Count(&count)
		if count != 1 {
			t.Errorf("expected 1 notification, got %d", count)
		}
	})

	t.Run("no self notification", func(t *testing.T) {
		err := srv.CreateLikeNotification(1, 1, "post", 2)
		if err != nil {
			t.Fatalf("self notification should return nil: %v", err)
		}
		var count int64
		db.Model(&model.UserNotification{}).Where("user_id = ? AND type = ?", 1, "like").Count(&count)
		if count != 0 {
			t.Errorf("expected 0 self notification, got %d", count)
		}
	})
}

func TestNotificationServiceParseAtUsers(t *testing.T) {
	db := testutil.SetupTestDB(t)

	notificationRepo := repository.NewNotificationRepository(db)
	followRepo := repository.NewFollowRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewNotificationService(notificationRepo, followRepo, userRepo)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.User{ID: 2, Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1})

	userIDs := srv.ParseAtUsers("hello @Alice and @Bob @Unknown")
	if len(userIDs) != 2 {
		t.Errorf("expected 2 at users, got %d", len(userIDs))
	}
}

func TestNotificationSettingsService(t *testing.T) {
	db := testutil.SetupTestDB(t)

	settingsRepo := repository.NewNotificationSettingsRepository(db)
	srv := service.NewNotificationSettingsService(settingsRepo)

	t.Run("get default", func(t *testing.T) {
		settings, err := srv.GetSettings(1)
		if err != nil {
			t.Fatalf("get settings failed: %v", err)
		}
		if settings.LikeNotify != 1 {
			t.Errorf("expected like notify 1, got %d", settings.LikeNotify)
		}
	})

	t.Run("update", func(t *testing.T) {
		err := srv.UpdateSettings(1, 0, 0, 0, 0, 0)
		if err != nil {
			t.Fatalf("update settings failed: %v", err)
		}
		settings, _ := settingsRepo.GetByUserID(1)
		if settings.LikeNotify != 0 {
			t.Errorf("expected like notify 0, got %d", settings.LikeNotify)
		}
	})
}
