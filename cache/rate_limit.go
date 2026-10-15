package cache

import (
	"context"
	"time"

	"pet-community/config"
)

// AllowRate 固定窗口计数限流；超限返回 false
func AllowRate(key string, limit int, window time.Duration) bool {
	if !config.RedisEnabled() || limit <= 0 {
		return true
	}
	ctx := context.Background()
	n, err := config.RDB.Incr(ctx, key).Result()
	if err != nil {
		// Redis 异常时放行，避免全站不可用；可由外围监控告警
		return true
	}
	if n == 1 {
		_ = config.RDB.Expire(ctx, key, window).Err()
	}
	return n <= int64(limit)
}

// AllowIPRoute 单 IP 单路由
func AllowIPRoute(ip, route string) bool {
	rl := config.Cfg.RateLimit
	window := time.Duration(rl.WindowSeconds) * time.Second
	return AllowRate(keyRateIPRoute(ip, route), rl.PerIPPerRoute, window)
}

// AllowIPGlobal 单 IP 全局
func AllowIPGlobal(ip string) bool {
	rl := config.Cfg.RateLimit
	window := time.Duration(rl.WindowSeconds) * time.Second
	return AllowRate(keyRateIPGlobal(ip), rl.PerIPGlobal, window)
}

// AllowUserRoute 单用户单路由
func AllowUserRoute(userID uint, route string) bool {
	rl := config.Cfg.RateLimit
	window := time.Duration(rl.WindowSeconds) * time.Second
	return AllowRate(keyRateUserRoute(userID, route), rl.PerUserPerRoute, window)
}
