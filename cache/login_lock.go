package cache

import (
	"context"
	"errors"
	"time"

	"pet-community/config"
)

const (
	LoginScopeUser  = "user"
	LoginScopeAdmin = "admin"
)

var ErrLoginLocked = errors.New("登录失败次数过多，请30分钟后再试")

// IsLoginLocked 是否处于登录锁定
func IsLoginLocked(scope string, id uint) (bool, error) {
	if !config.RedisEnabled() {
		return false, ErrUnavailable
	}
	n, err := config.RDB.Exists(context.Background(), keyLoginLock(scope, id)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// RecordLoginFailure 记录失败；达到阈值返回 ErrLoginLocked
func RecordLoginFailure(scope string, id uint) error {
	if !config.RedisEnabled() {
		return ErrUnavailable
	}
	ctx := context.Background()
	rdb := config.RDB
	failKey := keyLoginFail(scope, id)

	n, err := rdb.Incr(ctx, failKey).Result()
	if err != nil {
		return err
	}
	if n == 1 {
		_ = rdb.Expire(ctx, failKey, time.Duration(loginLockSeconds)*time.Second).Err()
	}
	if n >= int64(loginFailLimit) {
		_ = rdb.Set(ctx, keyLoginLock(scope, id), "1", time.Duration(loginLockSeconds)*time.Second).Err()
		_ = rdb.Del(ctx, failKey).Err()
		return ErrLoginLocked
	}
	return nil
}

// ClearLoginFailures 登录成功后清除失败计数与锁定
func ClearLoginFailures(scope string, id uint) {
	if !config.RedisEnabled() {
		return
	}
	ctx := context.Background()
	_ = config.RDB.Del(ctx, keyLoginFail(scope, id), keyLoginLock(scope, id)).Err()
}
