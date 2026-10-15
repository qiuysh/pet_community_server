package cache

import (
	"fmt"
	"strconv"
	"strings"

	"pet-community/config"
)

func keyAdView(id uint) string {
	return fmt.Sprintf("ad:view:%d", id)
}

func keyAdClick(id uint) string {
	return fmt.Sprintf("ad:click:%d", id)
}

// IncrAdView 缓冲广告曝光
func IncrAdView(id uint) bool {
	if !config.RedisEnabled() || id == 0 {
		return false
	}
	return config.RDB.Incr(ctxBG(), keyAdView(id)).Err() == nil
}

// IncrAdClick 缓冲广告点击
func IncrAdClick(id uint) bool {
	if !config.RedisEnabled() || id == 0 {
		return false
	}
	return config.RDB.Incr(ctxBG(), keyAdClick(id)).Err() == nil
}

// AdCounterFlushFunc 将缓冲计数刷入持久化存储
type AdCounterFlushFunc func(id uint, views, clicks int64) error

// FlushAdCounters 扫描并清空广告计数缓冲
func FlushAdCounters(flush AdCounterFlushFunc) {
	if !config.RedisEnabled() || flush == nil {
		return
	}
	ctx := ctxBG()
	rdb := config.RDB

	type pair struct{ views, clicks int64 }
	acc := map[uint]*pair{}

	collect := func(pattern, kind string) {
		iter := rdb.Scan(ctx, 0, pattern, 100).Iterator()
		for iter.Next(ctx) {
			key := iter.Val()
			parts := strings.Split(key, ":")
			if len(parts) < 3 {
				continue
			}
			id64, err := strconv.ParseUint(parts[2], 10, 64)
			if err != nil {
				continue
			}
			id := uint(id64)
			n, err := rdb.Get(ctx, key).Int64()
			if err != nil || n <= 0 {
				continue
			}
			_ = rdb.Del(ctx, key).Err()
			p := acc[id]
			if p == nil {
				p = &pair{}
				acc[id] = p
			}
			if kind == "view" {
				p.views += n
			} else {
				p.clicks += n
			}
		}
	}

	collect("ad:view:*", "view")
	collect("ad:click:*", "click")

	for id, p := range acc {
		_ = flush(id, p.views, p.clicks)
	}
}
