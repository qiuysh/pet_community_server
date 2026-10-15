package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"pet-community/config"
)

// GetSessionMeta 读取 status + token_version；ok=false 表示未命中
func GetSessionMeta(userID uint) (status, tv int, ok bool) {
	if !config.RedisEnabled() {
		return 0, 0, false
	}
	val, err := config.RDB.Get(context.Background(), keyTokenVersion(userID)).Result()
	if err != nil {
		return 0, 0, false
	}
	// 格式 "status:tv"，兼容旧纯数字 tv
	if parts := strings.SplitN(val, ":", 2); len(parts) == 2 {
		s, e1 := strconv.Atoi(parts[0])
		t, e2 := strconv.Atoi(parts[1])
		if e1 == nil && e2 == nil {
			return s, t, true
		}
		return 0, 0, false
	}
	t, err := strconv.Atoi(val)
	if err != nil {
		return 0, 0, false
	}
	return 1, t, true
}

// SetSessionMeta 缓存用户状态与 Token 版本
func SetSessionMeta(userID uint, status, tv int) {
	if !config.RedisEnabled() {
		return
	}
	val := fmt.Sprintf("%d:%d", status, tv)
	_ = config.RDB.Set(context.Background(), keyTokenVersion(userID), val, time.Duration(sessionTVTTL)*time.Second).Err()
}

// CacheDevice 标记设备在线
func CacheDevice(userID uint, deviceID string) {
	if !config.RedisEnabled() || deviceID == "" {
		return
	}
	_ = config.RDB.Set(context.Background(), keyDevice(userID, deviceID), "1", time.Duration(sessionDeviceTTL)*time.Second).Err()
}

// HasCachedDevice 设备缓存命中且存在
func HasCachedDevice(userID uint, deviceID string) bool {
	if !config.RedisEnabled() || deviceID == "" {
		return false
	}
	n, err := config.RDB.Exists(context.Background(), keyDevice(userID, deviceID)).Result()
	return err == nil && n > 0
}

// RemoveCachedDevice 删除单设备缓存
func RemoveCachedDevice(userID uint, deviceID string) {
	if !config.RedisEnabled() || deviceID == "" {
		return
	}
	_ = config.RDB.Del(context.Background(), keyDevice(userID, deviceID)).Err()
}

// InvalidateUserSession 全端失效会话缓存
func InvalidateUserSession(userID uint) {
	if !config.RedisEnabled() {
		return
	}
	ctx := context.Background()
	rdb := config.RDB
	_ = rdb.Del(ctx, keyTokenVersion(userID)).Err()

	iter := rdb.Scan(ctx, 0, keyDevice(userID, "*"), 50).Iterator()
	for iter.Next(ctx) {
		_ = rdb.Del(ctx, iter.Val()).Err()
	}
}
