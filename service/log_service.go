package service

import (
	"encoding/csv"
	"pet-community/model"
	"pet-community/repository"
	"strconv"
	"strings"
	"time"
)

type LogService struct {
	behaviorLogRepo *repository.BehaviorLogRepository
}

func NewLogService(behaviorLogRepo *repository.BehaviorLogRepository) *LogService {
	return &LogService{behaviorLogRepo: behaviorLogRepo}
}

var SecurityActionTypes = []string{
	"login",
	"login_failed",
	"logout",
	"password_change",
	"password_reset",
	"email_change",
	"phone_change",
	"account_lock",
	"account_unlock",
	"token_refresh",
}

func (s *LogService) RecordBehaviorLog(userID uint, actionType, targetType string, targetID uint, detail, ip, device string) error {
	log := &model.UserBehaviorLog{
		UserID:     userID,
		ActionType: actionType,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     detail,
		IP:         ip,
		Device:     device,
	}
	return s.behaviorLogRepo.Create(log)
}

func (s *LogService) GetBehaviorLogs(userID uint, actionType, targetType string, startTime, endTime time.Time, page, limit int) ([]model.UserBehaviorLog, int64, error) {
	return s.behaviorLogRepo.GetBehaviorLogs(userID, actionType, targetType, startTime, endTime, page, limit)
}

func (s *LogService) GetSecurityAuditLogs(userID uint, startTime, endTime time.Time, page, limit int) ([]model.UserBehaviorLog, int64, error) {
	return s.behaviorLogRepo.GetSecurityAuditLogs(userID, SecurityActionTypes, startTime, endTime, page, limit)
}

func (s *LogService) GetOperationLogs(userID uint, actionType string, startTime, endTime time.Time, page, limit int) ([]model.UserBehaviorLog, int64, error) {
	excludedTypes := make(map[string]bool)
	for _, t := range SecurityActionTypes {
		excludedTypes[t] = true
	}

	logs, total, err := s.behaviorLogRepo.GetBehaviorLogs(userID, actionType, "", startTime, endTime, page, limit)
	if err != nil {
		return nil, 0, err
	}

	filtered := make([]model.UserBehaviorLog, 0)
	for _, log := range logs {
		if !excludedTypes[log.ActionType] {
			filtered = append(filtered, log)
		}
	}

	return filtered, total, nil
}

func (s *LogService) ExportLogs(userID uint, actionType, targetType string, startTime, endTime time.Time) ([]byte, error) {
	logs, err := s.behaviorLogRepo.GetLogsForExport(userID, actionType, targetType, startTime, endTime)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	header := []string{"ID", "用户ID", "操作类型", "目标类型", "目标ID", "详情", "IP地址", "设备", "创建时间"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, log := range logs {
		record := []string{
			strconv.FormatUint(uint64(log.ID), 10),
			strconv.FormatUint(uint64(log.UserID), 10),
			log.ActionType,
			log.TargetType,
			strconv.FormatUint(uint64(log.TargetID), 10),
			log.Detail,
			log.IP,
			log.Device,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(builder.String()), nil
}

