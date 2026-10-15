package cache

import (
	"fmt"
	"strconv"
	"time"

	"pet-community/config"
)

const unreadTTL = 10 * time.Minute

func keyNotifUnread(userID uint) string {
	return fmt.Sprintf("unread:notif:%d", userID)
}

func keyMsgUnread(userID uint) string {
	return fmt.Sprintf("unread:msg:%d", userID)
}

func getUnread(key string) (int64, bool) {
	if !config.RedisEnabled() {
		return 0, false
	}
	val, err := config.RDB.Get(ctxBG(), key).Result()
	if err != nil {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

func setUnread(key string, n int64) {
	if !config.RedisEnabled() {
		return
	}
	_ = config.RDB.Set(ctxBG(), key, strconv.FormatInt(n, 10), unreadTTL).Err()
}

// GetNotifUnread / SetNotifUnread / IncrNotifUnread / InvalidateNotifUnread
func GetNotifUnread(userID uint) (int64, bool) { return getUnread(keyNotifUnread(userID)) }
func SetNotifUnread(userID uint, n int64)      { setUnread(keyNotifUnread(userID), n) }
func InvalidateNotifUnread(userID uint)        { DeleteKeys(keyNotifUnread(userID)) }

func IncrNotifUnread(userID uint) {
	if !config.RedisEnabled() {
		return
	}
	key := keyNotifUnread(userID)
	n, err := config.RDB.Incr(ctxBG(), key).Result()
	if err != nil {
		return
	}
	if n == 1 {
		_ = config.RDB.Expire(ctxBG(), key, unreadTTL).Err()
	}
}

// GetMsgUnread / SetMsgUnread / IncrMsgUnread / InvalidateMsgUnread
func GetMsgUnread(userID uint) (int64, bool) { return getUnread(keyMsgUnread(userID)) }
func SetMsgUnread(userID uint, n int64)      { setUnread(keyMsgUnread(userID), n) }
func InvalidateMsgUnread(userID uint)        { DeleteKeys(keyMsgUnread(userID)) }

func IncrMsgUnread(userID uint) {
	if !config.RedisEnabled() {
		return
	}
	key := keyMsgUnread(userID)
	n, err := config.RDB.Incr(ctxBG(), key).Result()
	if err != nil {
		return
	}
	if n == 1 {
		_ = config.RDB.Expire(ctxBG(), key, unreadTTL).Err()
	}
}
