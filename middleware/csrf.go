package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"

	"pet-community/config"
	"pet-community/types"
	"pet-community/utils"

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

// 静态上传目录前缀（CSRF 豁免），来自 storage.local.base_url（默认 /api/v1/uploads）
var (
	uploadPrefixMu sync.RWMutex
	uploadPrefix   = config.DefaultUploadBaseURL + "/"
)

// SetUploadPrefix 设置上传文件的对外访问前缀（启动阶段由 storage.local.base_url 注入）
func SetUploadPrefix(prefix string) {
	p := strings.TrimSpace(prefix)
	if p == "" {
		return
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	uploadPrefixMu.Lock()
	uploadPrefix = p
	uploadPrefixMu.Unlock()
}

// isUploadPath 判断是否为静态上传资源访问（豁免 CSRF）
func isUploadPath(path string) bool {
	uploadPrefixMu.RLock()
	prefix := uploadPrefix
	uploadPrefixMu.RUnlock()
	return strings.HasPrefix(path, prefix)
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
		if csrfExemptPath(path) || isUploadPath(path) || path == "/ws" {
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
