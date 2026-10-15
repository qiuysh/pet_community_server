package cache

import (
	"context"
	"time"

	"pet-community/config"
	"pet-community/types"
)

// GetRanking 读取排行榜缓存
func GetRanking(period string) (*types.RankingResponse, bool) {
	var resp types.RankingResponse
	if !GetJSON(keyRanking(period), &resp) {
		return nil, false
	}
	return &resp, true
}

// SetRanking 写入排行榜缓存
func SetRanking(period string, resp *types.RankingResponse) {
	if resp == nil {
		return
	}
	SetJSON(keyRanking(period), resp, time.Duration(rankingTTLSeconds)*time.Second)
}

// GetHotKeywords 读取热搜列表缓存
func GetHotKeywords(limit int) ([]types.HotKeywordInfo, bool) {
	var list []types.HotKeywordInfo
	if !GetJSON(keyHotKeywords(limit), &list) {
		return nil, false
	}
	return list, true
}

// SetHotKeywords 写入热搜列表缓存
func SetHotKeywords(limit int, list []types.HotKeywordInfo) {
	SetJSON(keyHotKeywords(limit), list, time.Duration(hotKeywordTTLSeconds)*time.Second)
}

// IncrHotKeyword 搜索热度 +1，并失效热搜快照
func IncrHotKeyword(keyword string) {
	if !config.RedisEnabled() || keyword == "" {
		return
	}
	ctx := context.Background()
	_ = config.RDB.ZIncrBy(ctx, keyHotKeywordZSet, 1, keyword).Err()
	DeleteByPattern("hot:keywords:top:*")
}

// TopHotKeywordsFromZSet 从 ZSET 取 TopN；不足或失败返回 nil
func TopHotKeywordsFromZSet(limit int) []types.HotKeywordInfo {
	if !config.RedisEnabled() || limit <= 0 {
		return nil
	}
	ctx := context.Background()
	vals, err := config.RDB.ZRevRangeWithScores(ctx, keyHotKeywordZSet, 0, int64(limit-1)).Result()
	if err != nil || len(vals) == 0 {
		return nil
	}
	out := make([]types.HotKeywordInfo, 0, len(vals))
	for i, z := range vals {
		out = append(out, types.HotKeywordInfo{
			Keyword:     z.Member.(string),
			SearchCount: int(z.Score),
			Rank:        i + 1,
		})
	}
	return out
}
