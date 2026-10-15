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

func extractActionType(path, method string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/users/login"):
		return "login"
	case strings.HasPrefix(path, "/api/v1/users/logout"):
		return "logout"
	case strings.HasPrefix(path, "/api/v1/users/register"):
		return "register"
	case strings.HasPrefix(path, "/api/v1/users/password"):
		if method == "PUT" || method == "POST" {
			return "password_change"
		}
	case strings.HasPrefix(path, "/api/v1/users/profile"):
		if method == "PUT" {
			return "profile_update"
		}
	case strings.HasPrefix(path, "/api/v1/pets"):
		if method == "POST" {
			return "pet_create"
		} else if method == "PUT" {
			return "pet_update"
		} else if method == "DELETE" {
			return "pet_delete"
		} else {
			return "pet_view"
		}
	case strings.HasPrefix(path, "/api/v1/posts"):
		if method == "POST" {
			return "post_create"
		} else if method == "PUT" {
			return "post_update"
		} else if method == "DELETE" {
			return "post_delete"
		} else {
			return "post_view"
		}
	case strings.HasPrefix(path, "/api/v1/comments"):
		if method == "POST" {
			return "comment_create"
		} else if method == "DELETE" {
			return "comment_delete"
		}
	case strings.HasPrefix(path, "/api/v1/likes"):
		return "like"
	case strings.HasPrefix(path, "/api/v1/favorites"):
		return "favorite"
	case strings.HasPrefix(path, "/api/v1/follows"):
		if method == "POST" {
			return "follow"
		} else if method == "DELETE" {
			return "unfollow"
		}
	case strings.HasPrefix(path, "/api/v1/activities"):
		if method == "POST" {
			return "activity_create"
		} else if method == "PUT" {
			return "activity_update"
		} else {
			return "activity_view"
		}
	case strings.HasPrefix(path, "/api/v1/adopt"):
		if method == "POST" {
			return "adopt_create"
		} else {
			return "adopt_view"
		}
	case strings.HasPrefix(path, "/api/v1/messages"):
		if method == "POST" {
			return "message_send"
		} else {
			return "message_view"
		}
	case strings.HasPrefix(path, "/api/v1/points"):
		return "points_query"
	default:
		return "api_request"
	}
	return "api_request"
}

func extractTargetType(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/pets"):
		return "pet"
	case strings.HasPrefix(path, "/api/v1/posts"):
		return "post"
	case strings.HasPrefix(path, "/api/v1/comments"):
		return "comment"
	case strings.HasPrefix(path, "/api/v1/likes"):
		return "like"
	case strings.HasPrefix(path, "/api/v1/favorites"):
		return "favorite"
	case strings.HasPrefix(path, "/api/v1/follows"):
		return "follow"
	case strings.HasPrefix(path, "/api/v1/activities"):
		return "activity"
	case strings.HasPrefix(path, "/api/v1/adopt"):
		return "adopt"
	case strings.HasPrefix(path, "/api/v1/messages"):
		return "message"
	case strings.HasPrefix(path, "/api/v1/users"):
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
