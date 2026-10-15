package tests

import (
	"testing"

	"pet-community/cache"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/model"
	"time"
)

func TestRedisSensitiveWordsCache(t *testing.T) {
	setupMiniRedis(t)
	db := testutil.SetupTestDB(t)
	testutil.SetGlobalDB(db)
	t.Cleanup(func() { testutil.SetGlobalDB(nil) })

	_ = db.Create(&model.SystemConfig{ConfigKey: "audit_sensitive_enabled", ConfigValue: "true"}).Error
	_ = db.Create(&model.SensitiveWord{Word: "违禁", ReplaceWith: "**", Status: 1, CreatedAt: time.Now()}).Error

	got := service.FilterSensitiveText("这是违禁内容")
	if got != "这是**内容" {
		t.Fatalf("first filter got %q", got)
	}
	// 第二次应命中缓存（删库词条仍能过滤说明走了缓存）
	_ = db.Where("word = ?", "违禁").Delete(&model.SensitiveWord{}).Error
	got2 := service.FilterSensitiveText("这是违禁内容")
	if got2 != "这是**内容" {
		t.Fatalf("cached filter got %q", got2)
	}
	cache.InvalidateSensitiveAndAudit()
	got3 := service.FilterSensitiveText("这是违禁内容")
	if got3 != "这是违禁内容" {
		t.Fatalf("after invalidate expected original, got %q", got3)
	}
}

func TestRedisVoteReserve(t *testing.T) {
	setupMiniRedis(t)

	if err := cache.ReserveVote(1, 100, 2); err != nil {
		t.Fatalf("first vote: %v", err)
	}
	if err := cache.ReserveVote(1, 100, 2); err != cache.ErrVoteAlreadyToday {
		t.Fatalf("expected already today, got %v", err)
	}
	if err := cache.ReserveVote(1, 101, 2); err != nil {
		t.Fatalf("second submission: %v", err)
	}
	if err := cache.ReserveVote(1, 102, 2); err != cache.ErrVoteDayLimit {
		t.Fatalf("expected day limit, got %v", err)
	}
	cache.UndoVote(1, 101)
	if err := cache.ReserveVote(1, 102, 2); err != nil {
		t.Fatalf("after undo: %v", err)
	}
}

func TestRedisAdCounterBuffer(t *testing.T) {
	setupMiniRedis(t)

	if !cache.IncrAdView(7) || !cache.IncrAdView(7) || !cache.IncrAdClick(7) {
		t.Fatal("incr failed")
	}
	var gotViews, gotClicks int64
	cache.FlushAdCounters(func(id uint, views, clicks int64) error {
		if id != 7 {
			t.Fatalf("id=%d", id)
		}
		gotViews, gotClicks = views, clicks
		return nil
	})
	if gotViews != 2 || gotClicks != 1 {
		t.Fatalf("views=%d clicks=%d", gotViews, gotClicks)
	}
}
