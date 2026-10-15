package cache

import "errors"

// ErrUnavailable Redis 未启用或不可用，调用方应降级到 MySQL/跳过
var ErrUnavailable = errors.New("redis unavailable")
