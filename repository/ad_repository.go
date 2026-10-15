package repository

import (
	"pet-community/model"
	"time"

	"gorm.io/gorm"
)

type AdRepository struct {
	db *gorm.DB
}

func NewAdRepository(db *gorm.DB) *AdRepository {
	return &AdRepository{db: db}
}

func (r *AdRepository) CreatePosition(position *model.AdPosition) error {
	return r.db.Create(position).Error
}

func (r *AdRepository) GetPositionByID(id uint) (*model.AdPosition, error) {
	var position model.AdPosition
	err := r.db.First(&position, id).Error
	return &position, err
}

func (r *AdRepository) ListPositions(page, pageSize int) ([]model.AdPosition, int64, error) {
	var positions []model.AdPosition
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.AdPosition{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&positions).Error
	return positions, total, err
}

func (r *AdRepository) UpdatePosition(position *model.AdPosition) error {
	return r.db.Save(position).Error
}

func (r *AdRepository) DeletePosition(id uint) error {
	return r.db.Delete(&model.AdPosition{}, id).Error
}

func (r *AdRepository) GetActivePositions(location string) ([]model.AdPosition, error) {
	var positions []model.AdPosition
	query := r.db.Where("status = 1")
	if location != "" {
		query = query.Where("location = ?", location)
	}
	err := query.Order("created_at DESC").Find(&positions).Error
	return positions, err
}

func (r *AdRepository) CreateAd(ad *model.Advertisement) error {
	return r.db.Create(ad).Error
}

func (r *AdRepository) GetAdByID(id uint) (*model.Advertisement, error) {
	var ad model.Advertisement
	err := r.db.Preload("Position").First(&ad, id).Error
	return &ad, err
}

func (r *AdRepository) ListAds(positionID uint, status int, page, pageSize int) ([]model.Advertisement, int64, error) {
	var ads []model.Advertisement
	var total int64

	offset := (page - 1) * pageSize
	query := r.db.Model(&model.Advertisement{}).Preload("Position")

	if positionID > 0 {
		query = query.Where("position_id = ?", positionID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&ads).Error
	return ads, total, err
}

func (r *AdRepository) UpdateAd(ad *model.Advertisement) error {
	return r.db.Save(ad).Error
}

func (r *AdRepository) DeleteAd(id uint) error {
	return r.db.Delete(&model.Advertisement{}, id).Error
}

func (r *AdRepository) UpdateAdStatus(id uint, status int) error {
	return r.db.Model(&model.Advertisement{}).Where("id = ?", id).Update("status", status).Error
}

func (r *AdRepository) IncrementViewCount(id uint) error {
	return r.AddViewCount(id, 1)
}

func (r *AdRepository) IncrementClickCount(id uint) error {
	return r.AddClickCount(id, 1)
}

func (r *AdRepository) AddViewCount(id uint, delta int64) error {
	if delta <= 0 {
		return nil
	}
	return r.db.Model(&model.Advertisement{}).Where("id = ?", id).Update("view_count", gorm.Expr("view_count + ?", delta)).Error
}

func (r *AdRepository) AddClickCount(id uint, delta int64) error {
	if delta <= 0 {
		return nil
	}
	return r.db.Model(&model.Advertisement{}).Where("id = ?", id).Update("click_count", gorm.Expr("click_count + ?", delta)).Error
}

func (r *AdRepository) GetActiveAds(location string) ([]model.Advertisement, error) {
	var ads []model.Advertisement
	now := time.Now()

	query := r.db.Preload("Position").
		Joins("JOIN ad_positions ON ad_positions.id = advertisements.position_id").
		Where("advertisements.status = 1").
		Where("ad_positions.status = 1").
		Where("advertisements.start_time <= ?", now).
		Where("advertisements.end_time >= ?", now)

	if location != "" {
		query = query.Where("ad_positions.location = ?", location)
	}

	err := query.Order("advertisements.created_at DESC").Find(&ads).Error
	return ads, err
}

func (r *AdRepository) GetAdStats(startTime, endTime time.Time) (int, int, error) {
	var result struct {
		TotalViews  int `gorm:"column:total_views"`
		TotalClicks int `gorm:"column:total_clicks"`
	}

	err := r.db.Model(&model.Advertisement{}).
		Where("created_at >= ?", startTime).
		Where("created_at <= ?", endTime).
		Select("SUM(view_count) as total_views, SUM(click_count) as total_clicks").
		Scan(&result).Error

	return result.TotalViews, result.TotalClicks, err
}

func (r *AdRepository) GetAdStatsByDay(startTime, endTime time.Time) ([]struct {
	Date       string
	ViewCount  int
	ClickCount int
}, error) {
	var stats []struct {
		Date       string
		ViewCount  int
		ClickCount int
	}

	err := r.db.Model(&model.Advertisement{}).
		Where("created_at >= ?", startTime).
		Where("created_at <= ?", endTime).
		Select("DATE(created_at) as date, SUM(view_count) as view_count, SUM(click_count) as click_count").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&stats).Error

	return stats, err
}

func (r *AdRepository) GetAdStatsByPosition() ([]struct {
	PositionName string
	ViewCount    int
	ClickCount   int
}, error) {
	var stats []struct {
		PositionName string
		ViewCount    int
		ClickCount   int
	}

	err := r.db.Model(&model.Advertisement{}).
		Joins("JOIN ad_positions ON ad_positions.id = advertisements.position_id").
		Select("ad_positions.name as position_name, SUM(advertisements.view_count) as view_count, SUM(advertisements.click_count) as click_count").
		Group("ad_positions.id, ad_positions.name").
		Order("view_count DESC").
		Scan(&stats).Error

	return stats, err
}
