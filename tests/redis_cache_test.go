package tests

import (
	"testing"

	"pet-community/cache"
	"pet-community/config"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(func() {
		config.CloseRedis()
		mr.Close()
	})

	if config.Cfg == nil {
		config.Cfg = &config.Config{}
	}
	config.Cfg.RateLimit = config.RateLimitConfig{
		Enabled:         true,
		PerIPPerRoute:   100,
		PerIPGlobal:     500,
		PerUserPerRoute: 100,
		WindowSeconds:   60,
	}

	config.RDB = redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return mr
}

func TestRedisVerifyCode(t *testing.T) {
	setupMiniRedis(t)

	if err := cache.AllowSendCode("13800000000", "127.0.0.1"); err != nil {
		t.Fatalf("allow: %v", err)
	}
	if err := cache.StoreCode("13800000000", "login", "123456", "127.0.0.1"); err != nil {
		t.Fatalf("store: %v", err)
	}
	if err := cache.AllowSendCode("13800000000", "127.0.0.1"); err != cache.ErrSendTooFrequent {
		t.Fatalf("expected too frequent, got %v", err)
	}
	if err := cache.VerifyAndConsume("13800000000", "login", "123456"); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if err := cache.VerifyAndConsume("13800000000", "login", "123456"); err != cache.ErrCodeNotFound {
		t.Fatalf("expected consumed, got %v", err)
	}
}

func TestRedisLoginLock(t *testing.T) {
	setupMiniRedis(t)

	for i := 0; i < 4; i++ {
		if err := cache.RecordLoginFailure(cache.LoginScopeUser, 1); err != nil {
			t.Fatalf("fail %d: %v", i+1, err)
		}
	}
	if err := cache.RecordLoginFailure(cache.LoginScopeUser, 1); err != cache.ErrLoginLocked {
		t.Fatalf("expected lock, got %v", err)
	}
	locked, err := cache.IsLoginLocked(cache.LoginScopeUser, 1)
	if err != nil || !locked {
		t.Fatalf("expected locked, got locked=%v err=%v", locked, err)
	}
	cache.ClearLoginFailures(cache.LoginScopeUser, 1)
	locked, _ = cache.IsLoginLocked(cache.LoginScopeUser, 1)
	if locked {
		t.Fatal("expected unlocked after clear")
	}
}

func TestRedisSessionCache(t *testing.T) {
	setupMiniRedis(t)

	cache.SetSessionMeta(9, 1, 3)
	status, tv, ok := cache.GetSessionMeta(9)
	if !ok || status != 1 || tv != 3 {
		t.Fatalf("meta got status=%d tv=%d ok=%v", status, tv, ok)
	}
	cache.CacheDevice(9, "dev-a")
	if !cache.HasCachedDevice(9, "dev-a") {
		t.Fatal("expected device cached")
	}
	cache.InvalidateUserSession(9)
	_, _, ok = cache.GetSessionMeta(9)
	if ok {
		t.Fatal("expected session invalidated")
	}
	if cache.HasCachedDevice(9, "dev-a") {
		t.Fatal("expected device cleared")
	}
}

func TestRedisRateLimit(t *testing.T) {
	setupMiniRedis(t)
	config.Cfg.RateLimit.PerIPGlobal = 3

	ip := "10.0.0.1"
	if !cache.AllowIPGlobal(ip) || !cache.AllowIPGlobal(ip) || !cache.AllowIPGlobal(ip) {
		t.Fatal("first 3 should allow")
	}
	if cache.AllowIPGlobal(ip) {
		t.Fatal("4th should deny")
	}
}
