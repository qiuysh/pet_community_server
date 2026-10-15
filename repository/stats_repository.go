package repository

import (
	"pet-community/model"
	"pet-community/types"

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
	err := r.db.Where("stat_date = ?", date).First(&stats).Error
	if err != nil {
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
	var existing model.Statistics
	err := tx.Where("stat_date = ?", stats.StatDate).First(&existing).Error
	if err == nil {
		stats.ID = existing.ID
		return tx.Save(stats).Error
	}
	if err == gorm.ErrRecordNotFound {
		return tx.Create(stats).Error
	}
	return err
}

func (r *StatsRepository) GetUserGrowthStats(startDate, endDate string) (*types.UserGrowthStats, error) {
	var results []model.Statistics
	err := r.db.Model(&model.Statistics{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Order("stat_date ASC").
		Find(&results).Error
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
	var results []model.Statistics
	err := r.db.Model(&model.Statistics{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Order("stat_date ASC").
		Find(&results).Error
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
	var results []model.Statistics
	err := r.db.Model(&model.Statistics{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Order("stat_date ASC").
		Find(&results).Error
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
	var results []model.Statistics
	err := r.db.Model(&model.Statistics{}).
		Where("stat_date >= ? AND stat_date <= ?", startDate, endDate).
		Order("stat_date ASC").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
