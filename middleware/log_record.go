package middleware

import (
	"strings"
	"time"

	"pet-community/service"

	"github.com/gin-gonic/gin"
)

var logService *service.LogService

func SetLogService(ls *service.LogService) {
	logService = ls
}

func UserBehaviorLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if logService == nil {
			ctx.Next()
			return
		}

		startTime := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		ctx.Next()

		processTime := time.Since(startTime)

		if processTime > time.Second*5 {
			return
		}

		if shouldSkipLog(path) {
			return
		}

		userIDVal, exists := ctx.Get("user_id")
		if !exists {
			return
		}
		userID, ok := userIDVal.(uint)
		if !ok {
			return
		}

		ip := ctx.ClientIP()
		userAgent := ctx.GetHeader("User-Agent")
		device := parseDevice(userAgent)

		actionType := extractActionType(path, method)
		targetType := extractTargetType(path)
		targetID := extractTargetID(path)

		detail := method + " " + path + " | " + ctx.Request.UserAgent()

		go func() {
			logService.RecordBehaviorLog(userID, actionType, targetType, targetID, detail, ip, device)
		}()
	}
}

func shouldSkipLog(path string) bool {
	skipPaths := []string{
		"/api/v1/public-key",
		"/health",
		"/ws",
	}
	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

func parseDevice(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

// extractActionType 归档行为类型。
// 前缀必须与 server/router/*.go 的真实挂载路径一致：曾经写的是 /api/v1/users/login、
// /api/v1/likes 等旧路径，导致 login / like / favorite 长期被记为 api_request。
func extractActionType(path, method string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/auth/login"):
		return "login"
	case strings.HasPrefix(path, "/api/v1/auth/logout"):
		return "logout"
	case strings.HasPrefix(path, "/api/v1/auth/register"):
		return "register"
	case strings.HasPrefix(path, "/api/v1/auth/reset-password"),
		strings.HasPrefix(path, "/api/v1/user/password"):
		return "password_change"
	case strings.HasPrefix(path, "/api/v1/user/profile"):
		if method == "PUT" {
			return "profile_update"
		}
	case strings.HasPrefix(path, "/api/v1/posts") && strings.HasSuffix(path, "/like"):
		return "like"
	case strings.HasPrefix(path, "/api/v1/posts") && strings.HasSuffix(path, "/favorite"):
		return "favorite"
	case strings.HasPrefix(path, "/api/v1/posts") && strings.HasSuffix(path, "/share"):
		return "share"
	case strings.HasPrefix(path, "/api/v1/lost-pets") && strings.HasSuffix(path, "/share"):
		return "share"
	case strings.HasPrefix(path, "/api/v1/posts/comments"):
		switch method {
		case "POST":
			return "comment_create"
		case "PUT":
			return "comment_update"
		case "DELETE":
			return "comment_delete"
		}
	case strings.HasPrefix(path, "/api/v1/posts"):
		if strings.HasSuffix(path, "/comments") && method == "POST" {
			return "comment_create"
		}
		switch method {
		case "POST":
			return "post_create"
		case "PUT":
			return "post_update"
		case "DELETE":
			return "post_delete"
		default:
			return "post_view"
		}
	case strings.HasPrefix(path, "/api/v1/pets"):
		switch method {
		case "POST":
			return "pet_create"
		case "PUT":
			return "pet_update"
		case "DELETE":
			return "pet_delete"
		default:
			return "pet_view"
		}
	case strings.HasPrefix(path, "/api/v1/likes"):
		return "like"
	case strings.HasPrefix(path, "/api/v1/favorites"):
		return "favorite"
	case strings.HasPrefix(path, "/api/v1/social/follow"):
		if method == "DELETE" {
			return "unfollow"
		}
		return "follow"
	case strings.HasPrefix(path, "/api/v1/users/"):
		if strings.HasSuffix(path, "/follow") && (method == "POST" || method == "DELETE") {
			if method == "DELETE" {
				return "unfollow"
			}
			return "follow"
		}
	case strings.HasPrefix(path, "/api/v1/activities"), strings.HasPrefix(path, "/api/v1/admin/activities"):
		switch {
		case strings.HasSuffix(path, "/register"):
			return "activity_register"
		case strings.HasSuffix(path, "/cancel"):
			return "activity_cancel"
		case strings.HasSuffix(path, "/sign-in"):
			return "activity_sign_in"
		case strings.HasSuffix(path, "/vote") || strings.HasSuffix(path, "/cancel-vote"):
			return "activity_vote"
		case strings.HasSuffix(path, "/publish"), strings.HasSuffix(path, "/publish-award"):
			return "activity_publish"
		}
		switch method {
		case "POST":
			return "activity_create"
		case "PUT":
			return "activity_update"
		default:
			return "activity_view"
		}
	case strings.HasPrefix(path, "/api/v1/adopts"):
		if method == "POST" {
			return "adopt_create"
		}
		return "adopt_view"
	case strings.HasPrefix(path, "/api/v1/social/messages"):
		if method == "POST" {
			return "message_send"
		}
		return "message_view"
	case strings.HasPrefix(path, "/api/v1/social/points"):
		return "points_query"
	}
	return "api_request"
}

func extractTargetType(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/posts/comments"):
		return "comment"
	case strings.HasPrefix(path, "/api/v1/posts"):
		return "post"
	case strings.HasPrefix(path, "/api/v1/pets"):
		return "pet"
	case strings.HasPrefix(path, "/api/v1/users/"), strings.HasPrefix(path, "/api/v1/social/follow"):
		return "follow"
	case strings.HasPrefix(path, "/api/v1/favorites"):
		return "favorite"
	case strings.HasPrefix(path, "/api/v1/likes"):
		return "like"
	case strings.HasPrefix(path, "/api/v1/activities"), strings.HasPrefix(path, "/api/v1/admin/activities"):
		return "activity"
	case strings.HasPrefix(path, "/api/v1/adopts"):
		return "adopt"
	case strings.HasPrefix(path, "/api/v1/social/messages"):
		return "message"
	case strings.HasPrefix(path, "/api/v1/social/points"):
		return "points"
	case strings.HasPrefix(path, "/api/v1/auth"), strings.HasPrefix(path, "/api/v1/user"):
		return "user"
	default:
		return ""
	}
}

func extractTargetID(path string) uint {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" && parts[i] != "api" && parts[i] != "v1" {
				if id, err := parseUint(parts[i]); err == nil {
					return id
				}
			}
		}
	}
	return 0
}

func parseUint(s string) (uint, error) {
	var result uint
	var sign int
	var i int

	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}

	if i < len(s) && s[i] == '-' {
		sign = -1
		i++
	} else if i < len(s) && s[i] == '+' {
		i++
	}

	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		digit := uint(s[i] - '0')
		if result > (^uint(0)-digit)/10 {
			return 0, &parseError{}
		}
		result = result*10 + digit
		i++
	}

	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}

	if i != len(s) {
		return 0, &parseError{}
	}

	if sign == -1 {
		return 0, &parseError{}
	}

	return result, nil
}

type parseError struct{}

func (e *parseError) Error() string {
	return "invalid syntax"
}
