package tests

import (
	"testing"
	"time"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/types"
)

func TestRedisRankingCache(t *testing.T) {
	setupMiniRedis(t)

	resp := &types.RankingResponse{
		Posts:  []types.RankingPost{{ID: 1, Content: "hi", HotScore: 10}},
		Users:  []types.RankingUser{{ID: 2, Nickname: "u"}},
		Topics: []types.RankingTopic{{ID: 3, Name: "t"}},
	}
	cache.SetRanking("day", resp)

	got, ok := cache.GetRanking("day")
	if !ok {
		t.Fatal("expected ranking cache hit")
	}
	if len(got.Posts) != 1 || got.Posts[0].ID != 1 {
		t.Fatalf("unexpected posts: %+v", got.Posts)
	}
}

func TestRedisHotKeywordZSet(t *testing.T) {
	setupMiniRedis(t)

	cache.IncrHotKeyword("猫粮")
	cache.IncrHotKeyword("猫粮")
	cache.IncrHotKeyword("狗狗")

	list := cache.TopHotKeywordsFromZSet(10)
	if len(list) < 2 {
		t.Fatalf("expected zset keywords, got %d", len(list))
	}
	if list[0].Keyword != "猫粮" || list[0].SearchCount != 2 {
		t.Fatalf("expected 猫粮 top, got %+v", list[0])
	}

	cache.SetHotKeywords(10, list)
	cached, ok := cache.GetHotKeywords(10)
	if !ok || len(cached) != len(list) {
		t.Fatalf("snapshot miss: ok=%v len=%d", ok, len(cached))
	}
}

func TestRedisActiveAdsCache(t *testing.T) {
	setupMiniRedis(t)

	ads := []model.Advertisement{{
		ID:        1,
		Title:     "banner",
		ImageURL:  "http://x/a.jpg",
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(time.Hour),
		Status:    1,
	}}
	cache.SetActiveAds("home_banner", ads)

	got, ok := cache.GetActiveAds("home_banner")
	if !ok || len(got) != 1 || got[0].Title != "banner" {
		t.Fatalf("ads cache miss: ok=%v got=%+v", ok, got)
	}

	cache.InvalidateActiveAds()
	if _, ok := cache.GetActiveAds("home_banner"); ok {
		t.Fatal("expected ads invalidated")
	}
}
