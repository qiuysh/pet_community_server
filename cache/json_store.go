package cache

import (
	"context"
	"encoding/json"
	"time"

	"pet-community/config"
)

// GetJSON 读取并反序列化；未命中或 Redis 不可用返回 false
func GetJSON(key string, dest interface{}) bool {
	if !config.RedisEnabled() {
		return false
	}
	val, err := config.RDB.Get(context.Background(), key).Bytes()
	if err != nil || len(val) == 0 {
		return false
	}
	return json.Unmarshal(val, dest) == nil
}

// SetJSON 序列化写入并设置 TTL
func SetJSON(key string, value interface{}, ttl time.Duration) {
	if !config.RedisEnabled() {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = config.RDB.Set(context.Background(), key, raw, ttl).Err()
}

// DeleteKeys 删除若干 key
func DeleteKeys(keys ...string) {
	if !config.RedisEnabled() || len(keys) == 0 {
		return
	}
	_ = config.RDB.Del(context.Background(), keys...).Err()
}

// DeleteByPattern SCAN 删除匹配 key
func DeleteByPattern(pattern string) {
	if !config.RedisEnabled() {
		return
	}
	ctx := context.Background()
	iter := config.RDB.Scan(ctx, 0, pattern, 50).Iterator()
	for iter.Next(ctx) {
		_ = config.RDB.Del(ctx, iter.Val()).Err()
	}
}
