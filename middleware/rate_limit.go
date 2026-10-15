package middleware

import (
	"pet-community/cache"
	"pet-community/config"
	"pet-community/types"
	"pet-community/utils"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware IP 维度限流（全局 + 单路由）
func RateLimitMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if config.Cfg == nil || !config.Cfg.RateLimit.Enabled || !config.RedisEnabled() {
			ctx.Next()
			return
		}

		ip := ctx.ClientIP()
		route := ctx.FullPath()
		if route == "" {
			route = ctx.Request.URL.Path
		}

		if !cache.AllowIPGlobal(ip) || !cache.AllowIPRoute(ip, route) {
			utils.RespondError(ctx, types.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

// AllowUserRate 用户维度限流（在鉴权成功后调用）
func AllowUserRate(userID uint, route string) bool {
	if config.Cfg == nil || !config.Cfg.RateLimit.Enabled || !config.RedisEnabled() {
		return true
	}
	if route == "" {
		route = "/"
	}
	return cache.AllowUserRoute(userID, route)
}
