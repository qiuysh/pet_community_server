package service

import (
	"encoding/json"
	"strings"

	"pet-community/cache"
	"pet-community/config"
	"pet-community/model"
)

// 隐私枚举：与总览 §4.6 / §4.7 一致
// postVisibility: 0公开 1仅粉丝 2私密
// commentPermission / messagePermission: 0所有人 1仅粉丝 2禁止

const (
	VisibilityPublic    = 0
	VisibilityFollowers = 1
	VisibilityPrivate   = 2

	PermissionEveryone  = 0
	PermissionFollowers = 1
	PermissionDenied    = 2

	// 动态审核状态（与定时发布区分）
	PostStatusPending   = 0 // 待审核
	PostStatusPublished = 1 // 已发布
	PostStatusScheduled = 2 // 定时待发
	PostStatusRejected  = 3 // 已拒绝
)

// ParsePrivacySettings 解析用户隐私 JSON
func ParsePrivacySettings(raw string) (postVis, commentPerm, messagePerm int) {
	postVis, commentPerm, messagePerm = VisibilityPublic, PermissionEveryone, PermissionEveryone
	if raw == "" {
		return
	}
	var saved map[string]int
	if err := json.Unmarshal([]byte(raw), &saved); err != nil {
		return
	}
	if v, ok := saved["postVisibility"]; ok {
		postVis = v
	}
	if v, ok := saved["commentPermission"]; ok {
		commentPerm = v
	}
	if v, ok := saved["messagePermission"]; ok {
		messagePerm = v
	}
	return
}

// CanViewPost 是否可查看动态（悄悄关注不计入粉丝门禁）
func CanViewPost(viewerID, authorID uint, visibility int, isPublicFollower bool) bool {
	if viewerID == authorID {
		return true
	}
	switch visibility {
	case VisibilityPublic:
		return true
	case VisibilityFollowers:
		return viewerID > 0 && isPublicFollower
	case VisibilityPrivate:
		return false
	default:
		return false
	}
}

// MaskContact 联系方式脱敏（游客展示）
func MaskContact(contact string) string {
	if contact == "" {
		return ""
	}
	runes := []rune(contact)
	n := len(runes)
	if n <= 3 {
		return string(runes[0:1]) + "**"
	}
	if n <= 7 {
		return string(runes[0:2]) + "****" + string(runes[n-1:])
	}
	// 手机号等：保留前三后四
	if n >= 11 {
		return string(runes[0:3]) + "****" + string(runes[n-4:])
	}
	keepHead := 2
	keepTail := 2
	masked := make([]rune, 0, n)
	masked = append(masked, runes[0:keepHead]...)
	for i := 0; i < n-keepHead-keepTail; i++ {
		masked = append(masked, '*')
	}
	masked = append(masked, runes[n-keepTail:]...)
	return string(masked)
}

// ShouldRevealContact 登录用户可见完整联系方式；游客脱敏
func ShouldRevealContact(viewerID uint) bool {
	return viewerID > 0
}

// ApplyContactPrivacy 按查看者脱敏
func ApplyContactPrivacy(viewerID uint, contact string) string {
	if ShouldRevealContact(viewerID) {
		return contact
	}
	return MaskContact(contact)
}

// ResolvePostVisibility 发帖可见性：未显式传入时用账号默认（U8）
func ResolvePostVisibility(explicit *int, user *model.User) int {
	if explicit != nil {
		return *explicit
	}
	if user == nil {
		return VisibilityPublic
	}
	postVis, _, _ := ParsePrivacySettings(user.PrivacySettings)
	return postVis
}

// isAutoAuditEnabled 读取后台「自动审核」开关；未配置时默认自动通过
func isAutoAuditEnabled() bool {
	if enabled, ok := cache.GetAuditFlag(cache.KeyAuditAuto()); ok {
		return enabled
	}
	if config.DB == nil {
		return true
	}
	var cfg model.SystemConfig
	if err := config.DB.Where("config_key = ?", "audit_auto_enabled").First(&cfg).Error; err != nil {
		cache.SetAuditFlag(cache.KeyAuditAuto(), true)
		return true
	}
	enabled := cfg.ConfigValue == "true"
	cache.SetAuditFlag(cache.KeyAuditAuto(), enabled)
	return enabled
}

func isSensitiveFilterEnabled() bool {
	if enabled, ok := cache.GetAuditFlag(cache.KeyAuditSensitive()); ok {
		return enabled
	}
	if config.DB == nil {
		return false
	}
	var cfg model.SystemConfig
	if err := config.DB.Where("config_key = ?", "audit_sensitive_enabled").First(&cfg).Error; err != nil {
		cache.SetAuditFlag(cache.KeyAuditSensitive(), false)
		return false
	}
	enabled := cfg.ConfigValue == "true"
	cache.SetAuditFlag(cache.KeyAuditSensitive(), enabled)
	return enabled
}

func loadActiveSensitiveWords() []cache.SensitiveWordItem {
	if words, ok := cache.GetSensitiveWords(); ok {
		return words
	}
	if config.DB == nil {
		return nil
	}
	var rows []model.SensitiveWord
	if err := config.DB.Where("status = 1").Find(&rows).Error; err != nil || len(rows) == 0 {
		cache.SetSensitiveWords([]cache.SensitiveWordItem{})
		return nil
	}
	words := make([]cache.SensitiveWordItem, 0, len(rows))
	for _, w := range rows {
		words = append(words, cache.SensitiveWordItem{Word: w.Word, ReplaceWith: w.ReplaceWith})
	}
	cache.SetSensitiveWords(words)
	return words
}

// FilterSensitiveText 按敏感词库替换文本；未开启过滤时原样返回
func FilterSensitiveText(text string) string {
	if text == "" || !isSensitiveFilterEnabled() {
		return text
	}
	words := loadActiveSensitiveWords()
	if len(words) == 0 {
		return text
	}
	out := text
	for _, w := range words {
		if w.Word == "" || !strings.Contains(out, w.Word) {
			continue
		}
		repl := w.ReplaceWith
		if repl == "" {
			repl = "***"
		}
		out = strings.ReplaceAll(out, w.Word, repl)
	}
	return out
}
