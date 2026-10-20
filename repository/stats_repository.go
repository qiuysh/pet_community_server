package repository

import (
	"errors"
	"pet-community/model"
	"pet-community/types"
	"time"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

type PetTypeCount struct {
	PetType int
	Count   int64
}

func (r *StatsRepository) GetPetTypeDistribution() ([]PetTypeCount, error) {
	var results []PetTypeCount
	err := r.db.Model(&model.Pet{}).
		Select("pet_type, COUNT(*) as count").Group("pet_type").Scan(&results).Error
	return results, err
}

func (r *StatsRepository) CountActivityLogsByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.UserActivityLog{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

func (r *StatsRepository) FindByDate(date string) (*model.Statistics, error) {
	var stats model.Statistics
	q, err := scopeStatDateOnDay(r.db, date)
	if err != nil {
		return nil, err
	}
	if err := q.First(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *StatsRepository) Create(stats *model.Statistics) error {
	return r.db.Create(stats).Error
}

func (r *StatsRepository) Update(stats *model.Statistics) error {
	return r.db.Save(stats).Error
}

func (r *StatsRepository) CreateOrUpdateWithTx(tx *gorm.DB, stats *model.Statistics) error {
	// 半开区间 [day, nextDay) + 日期字符串绑定，避免 time.Time 经驱动变成
	// 'YYYY-MM-DD 08:00:00' 导致 Find 失败、Create 撞唯一键 1062；且可走索引。
	dateStr := stats.StatDate.In(time.Local).Format("2006-01-02")
	dayStart, err := parseStatDay(dateStr)
	if err != nil {
		return err
	}
	stats.StatDate = dayStart

	q, err := scopeStatDateOnDay(tx, dateStr)
	if err != nil {
		return err
	}
	var existing model.Statistics
	err = q.First(&existing).Error
	if err == nil {
		stats.ID = existing.ID
		stats.StatDate = existing.StatDate
		return tx.Save(stats).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := tx.Create(stats).Error; err != nil {
		againQ, qErr := scopeStatDateOnDay(tx, dateStr)
		if qErr != nil {
			return err
		}
		var again model.Statistics
		if findErr := againQ.First(&again).Error; findErr != nil {
			return err
		}
		stats.ID = again.ID
		stats.StatDate = again.StatDate
		return tx.Save(stats).Error
	}
	return nil
}

func (r *StatsRepository) GetUserGrowthStats(startDate, endDate string) (*types.UserGrowthStats, error) {
	results, err := r.GetStatsByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	data := make([]types.UserGrowthDay, 0, len(results))
	totalNew := 0
	totalActive := 0

	for _, stat := range results {
		data = append(data, types.UserGrowthDay{
			Date:        stat.StatDate.Format("2006-01-02"),
			NewUsers:    stat.NewUserCount,
			TotalUsers:  stat.UserCount,
			ActiveUsers: stat.ActiveUserCount,
		})
		totalNew += stat.NewUserCount
		totalActive += stat.ActiveUserCount
	}

	return &types.UserGrowthStats{
		Data:        data,
		TotalNew:    totalNew,
		TotalActive: totalActive,
	}, nil
}

func (r *StatsRepository) GetContentStats(startDate, endDate string) (*types.ContentStats, error) {
	results, err := r.GetStatsByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	data := make([]types.ContentStatsDay, 0, len(results))
	totalPosts := 0
	totalComments := 0
	totalTopics := 0

	for _, stat := range results {
		data = append(data, types.ContentStatsDay{
			Date:     stat.StatDate.Format("2006-01-02"),
			Posts:    stat.NewPostCount,
			Comments: stat.NewCommentCount,
			Topics:   stat.NewTopicCount,
		})
		totalPosts += stat.NewPostCount
		totalComments += stat.NewCommentCount
		totalTopics += stat.NewTopicCount
	}

	return &types.ContentStats{
		Data:          data,
		TotalPosts:    totalPosts,
		TotalComments: totalComments,
		TotalTopics:   totalTopics,
	}, nil
}

func (r *StatsRepository) GetInteractionStats(startDate, endDate string) (*types.InteractionStats, error) {
	results, err := r.GetStatsByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	data := make([]types.InteractionStatsDay, 0, len(results))
	totalLikes := 0
	totalFavorites := 0
	totalShares := 0
	totalComments := 0

	for _, stat := range results {
		data = append(data, types.InteractionStatsDay{
			Date:      stat.StatDate.Format("2006-01-02"),
			Likes:     stat.NewLikeCount,
			Favorites: stat.NewFavoriteCount,
			Shares:    stat.NewShareCount,
			Comments:  stat.NewCommentCount,
		})
		totalLikes += stat.NewLikeCount
		totalFavorites += stat.NewFavoriteCount
		totalShares += stat.NewShareCount
		totalComments += stat.NewCommentCount
	}

	return &types.InteractionStats{
		Data:           data,
		TotalLikes:     totalLikes,
		TotalFavorites: totalFavorites,
		TotalShares:    totalShares,
		TotalComments:  totalComments,
	}, nil
}

func (r *StatsRepository) GetStatsByDateRange(startDate, endDate string) ([]model.Statistics, error) {
	q, err := scopeStatDateInclusiveRange(r.db.Model(&model.Statistics{}), startDate, endDate)
	if err != nil {
		return nil, err
	}
	var results []model.Statistics
	if err := q.Order("stat_date ASC").Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

// parseStatDay 解析 YYYY-MM-DD 为本地日 00:00
func parseStatDay(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, time.Local)
}

// scopeStatDateOnDay 单日半开区间 [day, nextDay)，参数用日期字符串以便走索引、规避时区偏移
func scopeStatDateOnDay(db *gorm.DB, dateStr string) (*gorm.DB, error) {
	start, err := parseStatDay(dateStr)
	if err != nil {
		return nil, err
	}
	next := start.AddDate(0, 0, 1).Format("2006-01-02")
	return db.Where("stat_date >= ? AND stat_date < ?", dateStr, next), nil
}

// scopeStatDateInclusiveRange 闭区间日历日 [startDate, endDate] → SQL 半开区间 [start, end+1day)
func scopeStatDateInclusiveRange(db *gorm.DB, startDate, endDate string) (*gorm.DB, error) {
	if _, err := parseStatDay(startDate); err != nil {
		return nil, err
	}
	end, err := parseStatDay(endDate)
	if err != nil {
		return nil, err
	}
	next := end.AddDate(0, 0, 1).Format("2006-01-02")
	return db.Where("stat_date >= ? AND stat_date < ?", startDate, next), nil
}
