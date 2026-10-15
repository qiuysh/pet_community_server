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

// Count 统计行为日志总数
func (r *BehaviorLogRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.UserBehaviorLog{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByDate 统计指定日期的行为日志数量
func (r *BehaviorLogRepository) CountByDate(date string) (int64, error) {
	var count int64
	if err := r.db.Model(&model.UserBehaviorLog{}).Where("DATE(created_at) = ?", date).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
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

func (r *BehaviorLogRepository) GetLogStats(userID uint, actionType string, startTime, endTime time.Time) (map[string]int64, error) {
	var results []struct {
		ActionType string `json:"actionType"`
		Count      int64  `json:"count"`
	}

	query := r.db.Model(&model.UserBehaviorLog{}).
		Select("action_type, COUNT(*) as count").
		Group("action_type")

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.ActionType] = result.Count
	}

	return stats, nil
}

func (r *BehaviorLogRepository) GetDailyLogStats(startTime, endTime time.Time) ([]struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}, error) {
	results := make([]struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}, 0)

	err := r.db.Model(&model.UserBehaviorLog{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}
