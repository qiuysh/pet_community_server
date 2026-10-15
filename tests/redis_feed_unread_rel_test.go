package tests

import (
	"testing"

	"pet-community/cache"
	"pet-community/types"
)

func TestRedisGuestFeedCache(t *testing.T) {
	setupMiniRedis(t)

	resp := &types.RecommendResponse{
		List:     []types.RecommendPostInfo{{ID: 1, Content: "hello"}},
		Total:    1,
		Page:     1,
		PageSize: 20,
	}
	cache.SetGuestFeed(1, 20, "", resp)
	got, ok := cache.GetGuestFeed(1, 20, "")
	if !ok || len(got.List) != 1 || got.List[0].ID != 1 {
		t.Fatalf("feed miss: ok=%v got=%+v", ok, got)
	}
	cache.InvalidateGuestFeed()
	if _, ok := cache.GetGuestFeed(1, 20, ""); ok {
		t.Fatal("expected feed invalidated")
	}
}

func TestRedisUnreadCounters(t *testing.T) {
	setupMiniRedis(t)

	cache.IncrNotifUnread(3)
	cache.IncrNotifUnread(3)
	n, ok := cache.GetNotifUnread(3)
	if !ok || n != 2 {
		t.Fatalf("notif unread=%d ok=%v", n, ok)
	}
	cache.SetNotifUnread(3, 0)
	n, _ = cache.GetNotifUnread(3)
	if n != 0 {
		t.Fatalf("expected 0, got %d", n)
	}

	cache.IncrMsgUnread(4)
	m, ok := cache.GetMsgUnread(4)
	if !ok || m != 1 {
		t.Fatalf("msg unread=%d ok=%v", m, ok)
	}
	cache.InvalidateMsgUnread(4)
	if _, ok := cache.GetMsgUnread(4); ok {
		t.Fatal("expected msg unread invalidated")
	}
}

func TestRedisRelationEdges(t *testing.T) {
	setupMiniRedis(t)

	cache.SetFollowEdge(1, 2, true)
	exists, hit := cache.GetFollowEdge(1, 2)
	if !hit || !exists {
		t.Fatal("follow edge miss")
	}
	cache.SetFollowEdge(1, 2, false)
	exists, hit = cache.GetFollowEdge(1, 2)
	if !hit || exists {
		t.Fatal("expected follow false")
	}

	cache.SetBlockEither(5, 6, true)
	exists, hit = cache.GetBlockEither(6, 5)
	if !hit || !exists {
		t.Fatal("block either miss")
	}
	cache.SetBlockRelationIDs(5, []uint{6, 7})
	ids, ok := cache.GetBlockRelationIDs(5)
	if !ok || len(ids) != 2 {
		t.Fatalf("block ids=%v", ids)
	}
	cache.InvalidateBlockRelation(5, 6)
	if _, ok := cache.GetBlockRelationIDs(5); ok {
		t.Fatal("expected block ids cleared")
	}
}
