package repository

import (
	"pet-community/model"
	"time"

	"gorm.io/gorm"
)

type BehaviorLogRepository struct {
	db *gorm.DB
}

func NewBehaviorLogRepository(db *gorm.DB) *BehaviorLogRepository {
	return &BehaviorLogRepository{db: db}
}

func (r *BehaviorLogRepository) Create(log *model.UserBehaviorLog) error {
	return r.db.Create(log).Error
}

func (r *BehaviorLogRepository) GetBehaviorLogs(userID uint, actionType string, targetType string, startTime, endTime time.Time, page, limit int) ([]model.UserBehaviorLog, int64, error) {
	logs := make([]model.UserBehaviorLog, 0)
	var total int64
	query := r.db.Model(&model.UserBehaviorLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset((page - 1) * limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *BehaviorLogRepository) GetSecurityAuditLogs(userID uint, actionTypes []string, startTime, endTime time.Time, page, limit int) ([]model.UserBehaviorLog, int64, error) {
	logs := make([]model.UserBehaviorLog, 0)
	var total int64
	query := r.db.Model(&model.UserBehaviorLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if len(actionTypes) > 0 {
		query = query.Where("action_type IN (?)", actionTypes)
	}
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset((page - 1) * limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *BehaviorLogRepository) GetLogsForExport(userID uint, actionType string, targetType string, startTime, endTime time.Time) ([]model.UserBehaviorLog, error) {
	var logs []model.UserBehaviorLog
	query := r.db.Model(&model.UserBehaviorLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	return logs, query.Order("created_at DESC").Find(&logs).Error
}
