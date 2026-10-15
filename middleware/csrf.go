package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"pet-community/types"
	"pet-community/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

// IssueCSRFToken 签发 CSRF Cookie 并返回 token（双重 Cookie：前端回传 Header）
func IssueCSRFToken(ctx *gin.Context) {
	token := ctx.GetHeader(csrfHeaderName)
	if cookie, err := ctx.Cookie(csrfCookieName); err == nil && cookie != "" {
		token = cookie
	}
	if token == "" {
		token = generateCSRFToken()
	}
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(csrfCookieName, token, 24*3600, "/", "", false, false)
	utils.RespondSuccess(ctx, gin.H{"csrfToken": token})
}

func generateCSRFToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return hex.EncodeToString([]byte("fallback-csrf-token-pet-community"))
	}
	return hex.EncodeToString(buf)
}

func csrfExemptPath(path string) bool {
	switch path {
	case "/api/v1/csrf-token",
		"/api/v1/public-key",
		"/health",
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/send-code",
		"/api/v1/auth/refresh-token",
		"/api/v1/auth/reset-password",
		"/api/v1/admin/login":
		return true
	default:
		return false
	}
}

// CSRFMiddleware 双重 Cookie 校验：写操作要求 Cookie 与 X-CSRF-Token 一致
func CSRFMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			ctx.Next()
			return
		}
		path := ctx.Request.URL.Path
		if csrfExemptPath(path) || strings.HasPrefix(path, "/uploads/") || path == "/ws" {
			ctx.Next()
			return
		}
		if !strings.HasPrefix(path, "/api/") {
			ctx.Next()
			return
		}

		cookie, err := ctx.Cookie(csrfCookieName)
		header := ctx.GetHeader(csrfHeaderName)
		if err != nil || cookie == "" || header == "" || cookie != header {
			utils.RespondError(ctx, types.CodeForbidden, "CSRF校验失败")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
