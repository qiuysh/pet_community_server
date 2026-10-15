// server/service/admin_log_service.go
package service

import (
	"time"

	"pet-community/model"
	"pet-community/repository"
)

// AdminLogService 负责各类日志查询(操作/活动/前台用户操作)。
type AdminLogService struct {
	adminRepo *repository.AdminRepository
}

// NewAdminLogService 构造 AdminLogService。
func NewAdminLogService(adminRepo *repository.AdminRepository) *AdminLogService {
	return &AdminLogService{adminRepo: adminRepo}
}

func (s *AdminLogService) GetOperationLogs(adminID uint, action string, startTime, endTime time.Time, page, limit int) ([]model.AdminOperationLog, int64, error) {
	return s.adminRepo.GetOperationLogs(adminID, action, startTime, endTime, page, limit)
}

func (s *AdminLogService) GetActivityLogs(userID uint, action string, startTime, endTime time.Time, page, limit int) ([]model.UserActivityLog, int64, error) {
	return s.adminRepo.GetActivityLogs(userID, action, startTime, endTime, page, limit)
}

func (s *AdminLogService) GetFrontUserOperationLogs(userID uint, action string, startTime, endTime time.Time, page, limit int) ([]model.AdminOperationLog, int64, error) {
	return s.adminRepo.GetFrontUserOperationLogs(userID, action, startTime, endTime, page, limit)
}
