package cache

import (
	"time"

	"pet-community/config"
)

const (
	keyAuditAuto      = "cfg:audit_auto_enabled"
	keyAuditSensitive = "cfg:audit_sensitive_enabled"
	keySensitiveWords = "cfg:sensitive_words"
	auditCfgTTL       = 30 * time.Minute
)

// SensitiveWordItem 过滤用敏感词项
type SensitiveWordItem struct {
	Word        string `json:"word"`
	ReplaceWith string `json:"replaceWith"`
}

// GetAuditFlag 读审核开关缓存；ok=false 表示未命中
func GetAuditFlag(key string) (enabled bool, ok bool) {
	if !config.RedisEnabled() {
		return false, false
	}
	val, err := config.RDB.Get(ctxBG(), key).Result()
	if err != nil {
		return false, false
	}
	return val == "true", true
}

// SetAuditFlag 写审核开关
func SetAuditFlag(key string, enabled bool) {
	if !config.RedisEnabled() {
		return
	}
	v := "false"
	if enabled {
		v = "true"
	}
	_ = config.RDB.Set(ctxBG(), key, v, auditCfgTTL).Err()
}

func KeyAuditAuto() string      { return keyAuditAuto }
func KeyAuditSensitive() string { return keyAuditSensitive }

// GetSensitiveWords 读敏感词列表缓存
func GetSensitiveWords() ([]SensitiveWordItem, bool) {
	var words []SensitiveWordItem
	if !GetJSON(keySensitiveWords, &words) {
		return nil, false
	}
	return words, true
}

// SetSensitiveWords 写敏感词列表
func SetSensitiveWords(words []SensitiveWordItem) {
	SetJSON(keySensitiveWords, words, auditCfgTTL)
}

// InvalidateSensitiveAndAudit 敏感词或审核配置变更后失效
func InvalidateSensitiveAndAudit() {
	DeleteKeys(keyAuditAuto, keyAuditSensitive, keySensitiveWords)
}

// InvalidateSiteConfig 站点配置变更
func InvalidateSiteConfig() {
	DeleteKeys("cfg:site")
}

// GetSiteConfig / SetSiteConfig
func GetSiteConfig() (map[string]string, bool) {
	var m map[string]string
	if !GetJSON("cfg:site", &m) {
		return nil, false
	}
	return m, true
}

func SetSiteConfig(m map[string]string) {
	SetJSON("cfg:site", m, auditCfgTTL)
}
