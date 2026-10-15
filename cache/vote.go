package cache

import (
	"errors"
	"fmt"
	"time"

	"pet-community/config"
)

var (
	ErrVoteAlreadyToday = errors.New("今天已经为该作品投票")
	ErrVoteDayLimit     = errors.New("今日投票次数已达上限")
	ErrVoteHourLimit    = errors.New("每小时投票次数已达上限")
)

const voteHourLimit = 3

func keyVoteSub(userID, submissionID uint, day string) string {
	return fmt.Sprintf("vote:sub:%d:%d:%s", userID, submissionID, day)
}

func keyVoteDay(userID uint, day string) string {
	return fmt.Sprintf("vote:day:%d:%s", userID, day)
}

func keyVoteHour(userID uint) string {
	return fmt.Sprintf("vote:hour:%d", userID)
}

func todayStr() string {
	return time.Now().Format("2006-01-02")
}

func ttlUntilMidnight() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	d := next.Sub(now)
	if d < time.Minute {
		return time.Minute
	}
	return d
}

// ReserveVote 预占投票配额（Redis）；不可用返回 ErrUnavailable
func ReserveVote(userID, submissionID uint, votesPerDay int) error {
	if !config.RedisEnabled() {
		return ErrUnavailable
	}
	if votesPerDay <= 0 {
		votesPerDay = 10
	}
	ctx := ctxBG()
	rdb := config.RDB
	day := todayStr()
	subKey := keyVoteSub(userID, submissionID, day)

	ok, err := rdb.SetNX(ctx, subKey, "1", ttlUntilMidnight()).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrVoteAlreadyToday
	}

	dayKey := keyVoteDay(userID, day)
	nDay, err := rdb.Incr(ctx, dayKey).Result()
	if err != nil {
		_ = rdb.Del(ctx, subKey).Err()
		return err
	}
	if nDay == 1 {
		_ = rdb.Expire(ctx, dayKey, ttlUntilMidnight()).Err()
	}
	if nDay > int64(votesPerDay) {
		_ = rdb.Decr(ctx, dayKey).Err()
		_ = rdb.Del(ctx, subKey).Err()
		return ErrVoteDayLimit
	}

	hourKey := keyVoteHour(userID)
	nHour, err := rdb.Incr(ctx, hourKey).Result()
	if err != nil {
		_ = rdb.Decr(ctx, dayKey).Err()
		_ = rdb.Del(ctx, subKey).Err()
		return err
	}
	if nHour == 1 {
		_ = rdb.Expire(ctx, hourKey, time.Hour).Err()
	}
	if nHour > voteHourLimit {
		_ = rdb.Decr(ctx, hourKey).Err()
		_ = rdb.Decr(ctx, dayKey).Err()
		_ = rdb.Del(ctx, subKey).Err()
		return ErrVoteHourLimit
	}
	return nil
}

// UndoVote 回滚预占（DB 失败或取消投票）
func UndoVote(userID, submissionID uint) {
	if !config.RedisEnabled() {
		return
	}
	ctx := ctxBG()
	rdb := config.RDB
	day := todayStr()
	_ = rdb.Del(ctx, keyVoteSub(userID, submissionID, day)).Err()
	dayKey := keyVoteDay(userID, day)
	if n, err := rdb.Decr(ctx, dayKey).Result(); err == nil && n < 0 {
		_ = rdb.Set(ctx, dayKey, "0", ttlUntilMidnight()).Err()
	}
	hourKey := keyVoteHour(userID)
	if n, err := rdb.Decr(ctx, hourKey).Result(); err == nil && n < 0 {
		_ = rdb.Del(ctx, hourKey).Err()
	}
}

// VoteDayCount 今日已投票数（含已预占）
func VoteDayCount(userID uint) int64 {
	if !config.RedisEnabled() {
		return -1
	}
	n, err := config.RDB.Get(ctxBG(), keyVoteDay(userID, todayStr())).Int64()
	if err != nil {
		return 0
	}
	return n
}
