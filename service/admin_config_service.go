// server/service/admin_config_service.go
package service

import (
	"errors"
	"strconv"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"

	"gorm.io/gorm"
)

// AdminConfigService 负责站点配置 + 审核配置 + 敏感词管理。
type AdminConfigService struct {
	adminRepo *repository.AdminRepository
}

// NewAdminConfigService 构造 AdminConfigService。
func NewAdminConfigService(adminRepo *repository.AdminRepository) *AdminConfigService {
	return &AdminConfigService{adminRepo: adminRepo}
}

func (s *AdminConfigService) GetSiteConfig() (map[string]string, error) {
	if cached, ok := cache.GetSiteConfig(); ok {
		return cached, nil
	}

	configs := make(map[string]string)
	keys := []string{"site_name", "site_logo", "site_description", "site_icp"}

	for _, key := range keys {
		config, err := s.adminRepo.GetSystemConfig(key)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				configs[key] = ""
				continue
			}
			return nil, err
		}
		configs[key] = config.ConfigValue
	}
	cache.SetSiteConfig(configs)
	return configs, nil
}

func (s *AdminConfigService) UpdateSiteConfig(configs map[string]string, adminID uint, ip string) error {
	if err := s.adminRepo.BatchUpdateSystemConfig(configs); err != nil {
		return err
	}
	cache.InvalidateSiteConfig()

	log := &model.AdminOperationLog{
		AdminID: adminID,
		Action:  "update_site_config",
		Detail:  "更新网站配置",
		IP:      ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminConfigService) GetAuditConfig() (map[string]bool, error) {
	configs := make(map[string]bool)

	if enabled, ok := cache.GetAuditFlag(cache.KeyAuditAuto()); ok {
		configs["auto_audit_enabled"] = enabled
	} else {
		autoConfig, err := s.adminRepo.GetSystemConfig("audit_auto_enabled")
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		enabled := autoConfig != nil && autoConfig.ConfigValue == "true"
		configs["auto_audit_enabled"] = enabled
		cache.SetAuditFlag(cache.KeyAuditAuto(), enabled)
	}

	if enabled, ok := cache.GetAuditFlag(cache.KeyAuditSensitive()); ok {
		configs["sensitive_filter_enabled"] = enabled
	} else {
		sensitiveConfig, err := s.adminRepo.GetSystemConfig("audit_sensitive_enabled")
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		enabled := sensitiveConfig != nil && sensitiveConfig.ConfigValue == "true"
		configs["sensitive_filter_enabled"] = enabled
		cache.SetAuditFlag(cache.KeyAuditSensitive(), enabled)
	}

	return configs, nil
}

func (s *AdminConfigService) UpdateAuditConfig(autoAudit, sensitiveFilter bool, adminID uint, ip string) error {
	configs := map[string]string{
		"audit_auto_enabled":      strconv.FormatBool(autoAudit),
		"audit_sensitive_enabled": strconv.FormatBool(sensitiveFilter),
	}

	if err := s.adminRepo.BatchUpdateSystemConfig(configs); err != nil {
		return err
	}
	cache.SetAuditFlag(cache.KeyAuditAuto(), autoAudit)
	cache.SetAuditFlag(cache.KeyAuditSensitive(), sensitiveFilter)

	log := &model.AdminOperationLog{
		AdminID: adminID,
		Action:  "update_audit_config",
		Detail:  "更新审核配置",
		IP:      ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminConfigService) GetSensitiveWords(keyword string, page, limit int) ([]model.SensitiveWord, int64, error) {
	return s.adminRepo.SearchSensitiveWords(keyword, page, limit)
}

func (s *AdminConfigService) AddSensitiveWord(word, replaceWith, category string, adminID uint, ip string) error {
	exists, err := s.adminRepo.CheckSensitiveWordExists(word)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("敏感词已存在")
	}

	newWord := &model.SensitiveWord{
		Word:        word,
		ReplaceWith: replaceWith,
		Category:    category,
	}

	if err := s.adminRepo.CreateSensitiveWord(newWord); err != nil {
		return err
	}
	cache.InvalidateSensitiveAndAudit()

	log := &model.AdminOperationLog{
		AdminID: adminID,
		Action:  "add_sensitive_word",
		Detail:  "添加敏感词: " + word,
		IP:      ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminConfigService) UpdateSensitiveWord(id uint, word, replaceWith, category string, adminID uint, ip string) error {
	existing, err := s.adminRepo.GetSensitiveWordByID(id)
	if err != nil {
		return err
	}

	existing.Word = word
	existing.ReplaceWith = replaceWith
	existing.Category = category

	if err := s.adminRepo.UpdateSensitiveWord(existing); err != nil {
		return err
	}
	cache.InvalidateSensitiveAndAudit()

	log := &model.AdminOperationLog{
		AdminID: adminID,
		Action:  "update_sensitive_word",
		Detail:  "更新敏感词: " + word,
		IP:      ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminConfigService) DeleteSensitiveWord(id uint, adminID uint, ip string) error {
	if err := s.adminRepo.DeleteSensitiveWord(id); err != nil {
		return err
	}
	cache.InvalidateSensitiveAndAudit()

	log := &model.AdminOperationLog{
		AdminID: adminID,
		Action:  "delete_sensitive_word",
		Detail:  "删除敏感词ID: " + strconv.FormatUint(uint64(id), 10),
		IP:      ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}
