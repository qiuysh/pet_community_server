package controller

import (
	"strconv"
	"time"

	"pet-community/middleware"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminActor 从 gin context 提取的 admin 操作主体信息
// 修复 B1:统一使用 auth.go 写入的 snake_case key "admin_id",
// 而不是之前 24 处复制粘贴的错误 camelCase key "adminID"。
type AdminActor struct {
	AdminID uint
	Role    string
	IP      string
}

// GetAdminActor 从 context 中统一提取 admin actor。
// 单点定义:auth.go ctx.Set("admin_id", claims.AdminID) 写入的是 snake_case。
// 之前 24 处 controller 手写 ctx.Get("adminID")(错误 key)导致:
//   - UpdateAdminStatus 端点 exists==false 永远报错
//   - 其余 23 处 adminIDVal 为 nil → adminID=0,审计日志全失源
//
// 现在通过此 helper 集中读取,任何 key 命名变更只改此处。
func GetAdminActor(ctx *gin.Context) (AdminActor, bool) {
	actor := AdminActor{IP: ctx.ClientIP()}

	idVal, exists := ctx.Get("admin_id")
	if !exists {
		return actor, false
	}
	switch v := idVal.(type) {
	case uint:
		actor.AdminID = v
	case uint64:
		actor.AdminID = uint(v)
	case float64:
		actor.AdminID = uint(v)
	default:
		return actor, false
	}

	roleVal, exists := ctx.Get("role")
	if exists {
		if role, ok := roleVal.(string); ok {
			actor.Role = role
		}
	}

	return actor, true
}

// MustGetAdminActor 取 actor 并在缺失时直接返回 401,handler 直接 return。
// 用于所有写操作,避免遗漏 exists 检查。
func MustGetAdminActor(ctx *gin.Context) (AdminActor, bool) {
	actor, ok := GetAdminActor(ctx)
	if !ok {
		ctx.JSON(200, types.ErrorResponse(types.CodeUnauthorized, "未授权"))
	}
	return actor, ok
}

// NormalizePage 统一分页守卫。消除 controller 层 10+ 处重复。
func NormalizePage(page, pageSize, maxSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if maxSize <= 0 {
		maxSize = 100
	}
	if pageSize > maxSize {
		pageSize = 20
	}
	return page, pageSize
}

// ParsePageQuery 从 query string 读取 page/page_size 并归一化。
func ParsePageQuery(ctx *gin.Context, maxSize int) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	return NormalizePage(page, pageSize, maxSize)
}

// ParseLimitQuery 别名用于 logs 接口 limit 命名。
func ParseLimitQuery(ctx *gin.Context, maxSize int) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	return NormalizePage(page, limit, maxSize)
}

// ParseTimePtr 解析 RFC3339 字符串指针为 *time.Time,失败则 nil。
// 消除 CreateNotification/CreateAnnouncement 中 6 处相同解析代码。
func ParseTimePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil
	}
	return &t
}

// ParseTimeRange 解析 "2006-01-02 15:04:05" 格式的起止时间。
// 消除 8 处日志接口中相同的 startTime/endTime 解析代码。
func ParseTimeRange(startStr, endStr string) (time.Time, time.Time) {
	var startTime, endTime time.Time
	if startStr != "" {
		startTime, _ = time.Parse("2006-01-02 15:04:05", startStr)
	}
	if endStr != "" {
		endTime, _ = time.Parse("2006-01-02 15:04:05", endStr)
	}
	return startTime, endTime
}

// ParseUintParam 解析 URL path 参数为 uint。
func ParseUintParam(ctx *gin.Context, key string) (uint, bool) {
	idStr := ctx.Param(key)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// MustParseUintParam 解析 path uint 参数,失败返回 400。
func MustParseUintParam(ctx *gin.Context, key, label string) (uint, bool) {
	id, ok := ParseUintParam(ctx, key)
	if !ok {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "无效的"+label))
	}
	return id, ok
}

// RequireSuperAdmin helper 用于 handler 内部的超管兜底检查(建议在路由层挂中间件)。
func (a AdminActor) IsSuperAdmin() bool {
	return a.Role == middleware.RoleSuperAdmin
}
