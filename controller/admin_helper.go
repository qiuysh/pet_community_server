package controller

import (
	"fmt"
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
//
// 兼容两种精度：后台日志页的日期筛选用 granularity='day'（值为 YYYY-MM-DD），
// 而报表/导出可能带时分秒。历史上这里用 `_ =` 吞掉解析错误，结果日期筛选
// 永远解析失败、静默退化成「不限时间」，用户看到的是全量数据。
// 因此解析失败必须返回错误由 controller 报 400；纯日期的结束时间补齐到
// 当天 23:59:59，否则 created_at <= endTime 会漏掉当天记录。
func ParseTimeRange(startStr, endStr string) (time.Time, time.Time, error) {
	var startTime, endTime time.Time

	if startStr != "" {
		parsed, _, err := parseTimeValue(startStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("开始时间格式无效: %s", startStr)
		}
		startTime = parsed // 起始时间取当天零点，语义正确
	}

	if endStr != "" {
		parsed, pureDate, err := parseTimeValue(endStr)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("结束时间格式无效: %s", endStr)
		}
		if pureDate {
			// 日期粒度代表「截止到这一天结束」
			parsed = parsed.Add(24*time.Hour - time.Second)
		}
		endTime = parsed
	}

	return startTime, endTime, nil
}

// timeLayouts 按「越精确越先匹配」排列，避免日期字符串被宽松格式误判。
var timeLayouts = []string{
	"2006-01-02 15:04:05",
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02",
}

// dateOnlyLayout 标记「只有日期」的布局，用于决定结束时间是否补齐到当天末尾。
const dateOnlyLayout = "2006-01-02"

// ParseIntQueryDefault 解析可选整数查询参数，非法值返回错误而不是静默回退默认值。
//
// 背景：`status, _ = strconv.Atoi(raw)` 这类写法会让 status=abc 变成 0，
// 而 0 往往是「待审核」这样的有效状态 —— 用户传错参数看到的是被悄悄改过的筛选结果，
// 比直接报错难排查得多。空值仍返回默认值（不筛选）。
func ParseIntQueryDefault(ctx *gin.Context, key string, def int) (int, error) {
	raw := ctx.Query(key)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def, fmt.Errorf("%s 参数无效: %s", key, raw)
	}
	return v, nil
}

// parseTimeValue 依次尝试已知布局，返回解析结果与是否为纯日期。
func parseTimeValue(raw string) (time.Time, bool, error) {
	for _, layout := range timeLayouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, layout == dateOnlyLayout, nil
		}
	}
	return time.Time{}, false, fmt.Errorf("unsupported time format: %s", raw)
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
