package cache

import (
	"context"
	"errors"
	"time"

	"pet-community/config"
)

var (
	ErrSendTooFrequent = errors.New("发送太频繁，请稍后再试")
	ErrSendTooMany     = errors.New("发送次数过多，请稍后再试")
	ErrCodeNotFound    = errors.New("请先获取验证码")
	ErrCodeMismatch    = errors.New("验证码错误")
)

// AllowSendCode 校验验证码发送频控（1 分钟 1 次、1 小时 10 次）
func AllowSendCode(account, ip string) error {
	if !config.RedisEnabled() {
		return ErrUnavailable
	}
	ctx := context.Background()
	rdb := config.RDB

	nMin, err := rdb.Get(ctx, keySendMinute(account, ip)).Int()
	if err == nil && nMin >= sendMinuteLimit {
		return ErrSendTooFrequent
	}
	nHour, err := rdb.Get(ctx, keySendHour(account, ip)).Int()
	if err == nil && nHour >= sendHourLimit {
		return ErrSendTooMany
	}
	return nil
}

// StoreCode 存储验证码并记录发送次数
func StoreCode(account, codeType, code, ip string) error {
	if !config.RedisEnabled() {
		return ErrUnavailable
	}
	ctx := context.Background()
	rdb := config.RDB

	if err := rdb.Set(ctx, keyVerifyCode(account, codeType), code, time.Duration(codeTTLSeconds)*time.Second).Err(); err != nil {
		return err
	}

	kMin := keySendMinute(account, ip)
	if n, err := rdb.Incr(ctx, kMin).Result(); err == nil && n == 1 {
		_ = rdb.Expire(ctx, kMin, time.Minute).Err()
	}
	kHour := keySendHour(account, ip)
	if n, err := rdb.Incr(ctx, kHour).Result(); err == nil && n == 1 {
		_ = rdb.Expire(ctx, kHour, time.Hour).Err()
	}
	return nil
}

// VerifyAndConsume 校验并消费验证码（一次性）
func VerifyAndConsume(account, codeType, code string) error {
	if !config.RedisEnabled() {
		return ErrUnavailable
	}
	ctx := context.Background()
	rdb := config.RDB
	key := keyVerifyCode(account, codeType)

	stored, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return ErrCodeNotFound
	}
	if stored != code {
		return ErrCodeMismatch
	}
	_ = rdb.Del(ctx, key).Err()
	return nil
}
