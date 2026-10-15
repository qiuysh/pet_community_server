package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// RDB Redis 客户端；未启用或初始化失败时为 nil（业务应降级）
var RDB *redis.Client

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	PerIPPerRoute     int  `yaml:"per_ip_per_route"`     // 单 IP 单路由，默认 100/分钟
	PerIPGlobal       int  `yaml:"per_ip_global"`        // 单 IP 全局，默认 500/分钟
	PerUserPerRoute   int  `yaml:"per_user_per_route"`   // 单用户单路由，默认 100/分钟
	WindowSeconds     int  `yaml:"window_seconds"`       // 窗口秒数，默认 60
}

// InitRedis 按配置初始化 Redis；enabled=false 时跳过
func InitRedis() {
	if Cfg == nil || !Cfg.Redis.Enabled {
		log.Println("redis: disabled, cache features will fall back to MySQL/no-op")
		RDB = nil
		return
	}

	addr := fmt.Sprintf("%s:%d", Cfg.Redis.Host, Cfg.Redis.Port)
	poolSize := Cfg.Redis.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     Cfg.Redis.Password,
		DB:           Cfg.Redis.DB,
		PoolSize:     poolSize,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		// 不阻断启动：缓存/限流等自动降级到 MySQL 或跳过
		log.Printf("redis: ping %s failed: %v; continue without redis (features degraded)", addr, err)
		RDB = nil
		return
	}

	RDB = client
	log.Printf("redis: connected to %s db=%d", addr, Cfg.Redis.DB)
}

// RedisEnabled 是否已接通 Redis
func RedisEnabled() bool {
	return RDB != nil
}

// CloseRedis 关闭连接（测试用）
func CloseRedis() {
	if RDB != nil {
		_ = RDB.Close()
		RDB = nil
	}
}
