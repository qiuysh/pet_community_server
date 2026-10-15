package repository

import (
	"pet-community/model"

	"gorm.io/gorm"
)

type SearchHistoryRepository struct {
	db *gorm.DB
}

func NewSearchHistoryRepository(db *gorm.DB) *SearchHistoryRepository {
	return &SearchHistoryRepository{db: db}
}

func (r *SearchHistoryRepository) Create(history *model.SearchHistory) error {
	return r.db.Create(history).Error
}

func (r *SearchHistoryRepository) ListByUser(userID uint, limit int) ([]model.SearchHistory, error) {
	var histories []model.SearchHistory
	if err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *SearchHistoryRepository) Delete(userID, id uint) error {
	return r.db.Where("user_id = ? AND id = ?", userID, id).Delete(&model.SearchHistory{}).Error
}

func (r *SearchHistoryRepository) Clear(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.SearchHistory{}).Error
}

func (r *SearchHistoryRepository) Exists(userID uint, keyword string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.SearchHistory{}).Where("user_id = ? AND keyword = ?", userID, keyword).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

type UserInterestRepository struct {
	db *gorm.DB
}

func NewUserInterestRepository(db *gorm.DB) *UserInterestRepository {
	return &UserInterestRepository{db: db}
}

func (r *UserInterestRepository) Create(interest *model.UserInterest) error {
	return r.db.Create(interest).Error
}

func (r *UserInterestRepository) Update(interest *model.UserInterest) error {
	return r.db.Save(interest).Error
}

func (r *UserInterestRepository) ListByUser(userID uint) ([]model.UserInterest, error) {
	var interests []model.UserInterest
	if err := r.db.Where("user_id = ?", userID).Order("weight DESC").Find(&interests).Error; err != nil {
		return nil, err
	}
	return interests, nil
}

func (r *UserInterestRepository) GetByUserAndInterest(userID uint, interest string) (*model.UserInterest, error) {
	var item model.UserInterest
	if err := r.db.Where("user_id = ? AND interest = ?", userID, interest).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *UserInterestRepository) UpdateWeight(userID uint, interest string, delta float64) error {
	return r.db.Model(&model.UserInterest{}).Where("user_id = ? AND interest = ?", userID, interest).Updates(map[string]interface{}{
		"weight":      gorm.Expr("weight + ?", delta),
		"last_active": gorm.Expr("NOW()"),
	}).Error
}

type TopicCategoryRepository struct {
	db *gorm.DB
}

func NewTopicCategoryRepository(db *gorm.DB) *TopicCategoryRepository {
	return &TopicCategoryRepository{db: db}
}

func (r *TopicCategoryRepository) Create(cat *model.TopicCategory) error {
	return r.db.Create(cat).Error
}

func (r *TopicCategoryRepository) List(status int) ([]model.TopicCategory, error) {
	var categories []model.TopicCategory
	query := r.db.Model(&model.TopicCategory{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	if err := query.Order("sort_order ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *TopicCategoryRepository) GetByID(id uint) (*model.TopicCategory, error) {
	var cat model.TopicCategory
	if err := r.db.First(&cat, id).Error; err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *TopicCategoryRepository) InitDefaultCategories() error {
	var count int64
	r.db.Model(&model.TopicCategory{}).Count(&count)
	if count > 0 {
		return nil
	}

	categories := []model.TopicCategory{
		{Name: "萌宠日常", Description: "分享宠物的日常生活点滴", Icon: "🐾", SortOrder: 1, Status: 1},
		{Name: "养宠知识", Description: "科学养宠、健康护理知识", Icon: "📚", SortOrder: 2, Status: 1},
		{Name: "宠物健康", Description: "宠物医疗、疾病预防", Icon: "❤️", SortOrder: 3, Status: 1},
		{Name: "搞笑萌宠", Description: "宠物搞笑瞬间", Icon: "😆", SortOrder: 4, Status: 1},
		{Name: "领养送养", Description: "宠物领养信息", Icon: "🏠", SortOrder: 5, Status: 1},
		{Name: "宠物活动", Description: "宠物相关活动聚会", Icon: "🎉", SortOrder: 6, Status: 1},
	}

	return r.db.Create(&categories).Error
}

type TopicParticipantRepository struct {
	db *gorm.DB
}

func NewTopicParticipantRepository(db *gorm.DB) *TopicParticipantRepository {
	return &TopicParticipantRepository{db: db}
}

func (r *TopicParticipantRepository) Create(participant *model.TopicParticipant) error {
	return r.db.Create(participant).Error
}

func (r *TopicParticipantRepository) Delete(topicID, userID uint) error {
	return r.db.Where("topic_id = ? AND user_id = ?", topicID, userID).Delete(&model.TopicParticipant{}).Error
}

func (r *TopicParticipantRepository) Exists(topicID, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.TopicParticipant{}).Where("topic_id = ? AND user_id = ?", topicID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *TopicParticipantRepository) CountByTopic(topicID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.TopicParticipant{}).Where("topic_id = ?", topicID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type HotKeywordRepository struct {
	db *gorm.DB
}

func NewHotKeywordRepository(db *gorm.DB) *HotKeywordRepository {
	return &HotKeywordRepository{db: db}
}

func (r *HotKeywordRepository) Create(keyword *model.HotKeyword) error {
	return r.db.Create(keyword).Error
}

func (r *HotKeywordRepository) Update(keyword *model.HotKeyword) error {
	return r.db.Save(keyword).Error
}

func (r *HotKeywordRepository) List(limit int) ([]model.HotKeyword, error) {
	var keywords []model.HotKeyword
	sql := "SELECT * FROM hot_keywords ORDER BY `rank` ASC, search_count DESC, id ASC LIMIT ?"
	if err := r.db.Raw(sql, limit).Scan(&keywords).Error; err != nil {
		return nil, err
	}
	return keywords, nil
}

func (r *HotKeywordRepository) GetByKeyword(keyword string) (*model.HotKeyword, error) {
	var item model.HotKeyword
	if err := r.db.Where("keyword = ?", keyword).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *HotKeywordRepository) IncrementSearchCount(keyword string) error {
	return r.db.Model(&model.HotKeyword{}).Where("keyword = ?", keyword).UpdateColumn("search_count", gorm.Expr("search_count + 1")).Error
}

func (r *HotKeywordRepository) UpdateRanks() error {
	var keywords []model.HotKeyword
	if err := r.db.Order("search_count DESC").Find(&keywords).Error; err != nil {
		return err
	}

	for i, kw := range keywords {
		kw.Rank = i + 1
		if err := r.db.Save(&kw).Error; err != nil {
			return err
		}
	}
	return nil
}
