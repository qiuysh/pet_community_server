package middleware

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func UserAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondError(ctx, types.CodeUnauthorized, "未授权")
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.RespondError(ctx, types.CodeUnauthorized, "授权格式错误")
			ctx.Abort()
			return
		}

		userID, err := service.ParseUserToken(parts[1])
		if err != nil {
			utils.RespondError(ctx, types.CodeUnauthorized, "Token无效或已过期")
			ctx.Abort()
			return
		}

		route := ctx.FullPath()
		if route == "" {
			route = ctx.Request.URL.Path
		}
		if !AllowUserRate(userID, route) {
			utils.RespondError(ctx, types.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			ctx.Abort()
			return
		}

		ctx.Set("user_id", userID)
		ctx.Next()
	}
}

// OptionalUserAuthMiddleware 公开接口可选登录：有有效 Token 则注入 user_id，否则按游客继续
func OptionalUserAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			ctx.Next()
			return
		}
		userID, err := service.ParseUserToken(parts[1])
		if err == nil {
			ctx.Set("user_id", userID)
		}
		ctx.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondError(ctx, types.CodeUnauthorized, "未授权")
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.RespondError(ctx, types.CodeUnauthorized, "授权格式错误")
			ctx.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.RespondError(ctx, types.CodeUnauthorized, "Token无效或已过期")
			ctx.Abort()
			return
		}

		ctx.Set("admin_id", claims.AdminID)
		ctx.Set("username", claims.Username)
		ctx.Set("role", claims.Role)
		ctx.Next()
	}
}
