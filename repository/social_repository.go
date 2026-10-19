package repository

import (
	"strings"
	"time"

	"pet-community/cache"
	"pet-community/model"

	"gorm.io/gorm"
)

type BlockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) *BlockRepository {
	return &BlockRepository{db: db}
}

func (r *BlockRepository) Create(block *model.Block) error {
	if err := r.db.Create(block).Error; err != nil {
		return err
	}
	cache.SetBlockEdge(block.BlockerID, block.BlockedID, true)
	cache.SetBlockEither(block.BlockerID, block.BlockedID, true)
	return nil
}

func (r *BlockRepository) Delete(blockerID, blockedID uint) error {
	if err := r.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&model.Block{}).Error; err != nil {
		return err
	}
	cache.InvalidateBlockRelation(blockerID, blockedID)
	return nil
}

func (r *BlockRepository) Exists(blockerID, blockedID uint) (bool, error) {
	if exists, hit := cache.GetBlockEdge(blockerID, blockedID); hit {
		return exists, nil
	}
	var count int64
	if err := r.db.Model(&model.Block{}).Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Count(&count).Error; err != nil {
		return false, err
	}
	exists := count > 0
	cache.SetBlockEdge(blockerID, blockedID, exists)
	return exists, nil
}

// ExistsEither 任一方拉黑另一方即返回 true（双向屏蔽语义）
func (r *BlockRepository) ExistsEither(userA, userB uint) (bool, error) {
	if exists, hit := cache.GetBlockEither(userA, userB); hit {
		return exists, nil
	}
	var count int64
	err := r.db.Model(&model.Block{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", userA, userB, userB, userA).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	exists := count > 0
	cache.SetBlockEither(userA, userB, exists)
	return exists, nil
}

// GetBlockedRelationIDs 返回与 userID 存在拉黑关系的全部对方用户 ID
func (r *BlockRepository) GetBlockedRelationIDs(userID uint) ([]uint, error) {
	if ids, ok := cache.GetBlockRelationIDs(userID); ok {
		return ids, nil
	}
	var blocks []model.Block
	if err := r.db.Where("blocker_id = ? OR blocked_id = ?", userID, userID).Find(&blocks).Error; err != nil {
		return nil, err
	}
	seen := make(map[uint]struct{}, len(blocks))
	ids := make([]uint, 0, len(blocks))
	for _, b := range blocks {
		other := b.BlockedID
		if b.BlockerID != userID {
			other = b.BlockerID
		}
		if _, ok := seen[other]; !ok {
			seen[other] = struct{}{}
			ids = append(ids, other)
		}
	}
	cache.SetBlockRelationIDs(userID, ids)
	return ids, nil
}

func (r *BlockRepository) GetBlockedList(userID uint, page, pageSize int) ([]model.Block, int64, error) {
	var blocks []model.Block
	var total int64
	query := r.db.Model(&model.Block{}).Where("blocker_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&blocks).Error; err != nil {
		return nil, 0, err
	}

	return blocks, total, nil
}

type ConversationRepository struct {
	db *gorm.DB
}

func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(conversation *model.Conversation) error {
	return r.db.Create(conversation).Error
}

func (r *ConversationRepository) GetByUsers(user1ID, user2ID uint) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.db.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)", user1ID, user2ID, user2ID, user1ID).First(&conversation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conversation, nil
}

func (r *ConversationRepository) GetByID(id uint) (*model.Conversation, error) {
	var conversation model.Conversation
	if err := r.db.First(&conversation, id).Error; err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *ConversationRepository) ListByUser(userID uint, page, pageSize int) ([]model.Conversation, int64, error) {
	var conversations []model.Conversation
	var total int64
	query := r.db.Model(&model.Conversation{}).Where("user1_id = ? OR user2_id = ?", userID, userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("updated_at DESC").Find(&conversations).Error; err != nil {
		return nil, 0, err
	}

	return conversations, total, nil
}

func (r *ConversationRepository) UpdateLastMessage(id uint, messageID uint, content string, timestamp *time.Time) error {
	return r.db.Model(&model.Conversation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_message_id":   messageID,
		"last_message":      content,
		"last_message_time": timestamp,
		"updated_at":        time.Now(),
	}).Error
}

func (r *ConversationRepository) IncrementUnread(id uint) error {
	return r.db.Model(&model.Conversation{}).Where("id = ?", id).UpdateColumn("unread_count", gorm.Expr("unread_count + 1")).Error
}

func (r *ConversationRepository) ResetUnread(id uint) error {
	return r.db.Model(&model.Conversation{}).Where("id = ?", id).UpdateColumn("unread_count", 0).Error
}

func (r *ConversationRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND (user1_id = ? OR user2_id = ?)", id, userID, userID).Delete(&model.Conversation{}).Error
}

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *MessageRepository) ListByConversation(conversationID uint, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64
	query := r.db.Model(&model.Message{}).Where("conversation_id = ?", conversationID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("created_at DESC, id DESC").Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

func (r *MessageRepository) GetByID(id uint) (*model.Message, error) {
	var message model.Message
	if err := r.db.First(&message, id).Error; err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *MessageRepository) MarkAsRead(conversationID, userID uint) error {
	return r.db.Model(&model.Message{}).Where("conversation_id = ? AND to_id = ? AND status = 0", conversationID, userID).Update("status", 1).Error
}

func (r *MessageRepository) Recall(id uint) error {
	return r.db.Model(&model.Message{}).Where("id = ?", id).Update("is_recalled", 1).Error
}

// Search 会话内搜索文本消息
// - 只命中未被撤回的文本消息，避免 [图片]/[语音] 占位文案污染结果
// - 关键词中的 % _ 会被转义，使用 ! 作为 ESCAPE 字符（MySQL/SQLite 一致）
func (r *MessageRepository) Search(conversationID uint, keyword string, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := r.db.Model(&model.Message{}).
		Where("conversation_id = ? AND is_recalled = 0 AND type = ? AND content LIKE ? ESCAPE '!'",
			conversationID, model.MessageTypeText, "%"+escapeLike(keyword)+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// escapeLike 转义 LIKE 通配符
func escapeLike(s string) string {
	return strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(s)
}

func (r *MessageRepository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Message{}).Where("to_id = ? AND status = 0", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type PointsRepository struct {
	db *gorm.DB
}

func NewPointsRepository(db *gorm.DB) *PointsRepository {
	return &PointsRepository{db: db}
}

func (r *PointsRepository) CreateRecord(record *model.PointsRecord) error {
	return r.db.Create(record).Error
}

func (r *PointsRepository) ListByUser(userID uint, page, pageSize int) ([]model.PointsRecord, int64, error) {
	var records []model.PointsRecord
	var total int64
	query := r.db.Model(&model.PointsRecord{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *PointsRepository) GetTodayPoints(userID uint, pointType string) (int, error) {
	var total int
	today := time.Now().Format("2006-01-02")
	if err := r.db.Model(&model.PointsRecord{}).
		Where("user_id = ? AND type = ? AND DATE(created_at) = ?", userID, pointType, today).
		Select("COALESCE(SUM(points), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PointsRepository) GetTodayPointsTotal(userID uint) (int, error) {
	var total int
	today := time.Now().Format("2006-01-02")
	if err := r.db.Model(&model.PointsRecord{}).
		Where("user_id = ? AND DATE(created_at) = ?", userID, today).
		Select("COALESCE(SUM(points), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PointsRuleRepository) InitDefaultRules() error {
	defaults := []model.PointsRule{
		{Type: "post_create", Points: 10, Description: "发布动态", IsActive: 1, DailyLimit: 0},
		{Type: "comment", Points: 2, Description: "发表评论", IsActive: 1, DailyLimit: 0},
		{Type: "post_liked", Points: 1, Description: "动态被点赞", IsActive: 1, DailyLimit: 0},
		{Type: "comment_received", Points: 2, Description: "动态被评论", IsActive: 1, DailyLimit: 0},
		{Type: "follow_received", Points: 5, Description: "被用户关注", IsActive: 1, DailyLimit: 0},
	}
	for _, rule := range defaults {
		var count int64
		if err := r.db.Model(&model.PointsRule{}).Where("type = ?", rule.Type).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := r.db.Create(&rule).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *LevelBenefitRepository) InitDefaultBenefits() error {
	defaults := []model.LevelBenefit{
		{Level: 1, BenefitType: "basic", BenefitValue: "1", Description: "基础发帖、评论、点赞"},
		{Level: 3, BenefitType: "badge", BenefitValue: "1", Description: "个人主页展示等级徽章"},
		{Level: 5, BenefitType: "topic_create", BenefitValue: "1", Description: "可申请创建话题"},
		{Level: 8, BenefitType: "activity_priority", BenefitValue: "1", Description: "活动报名排队优先"},
		{Level: 10, BenefitType: "nickname_style", BenefitValue: "1", Description: "专属昵称标识"},
	}
	for _, b := range defaults {
		var count int64
		if err := r.db.Model(&model.LevelBenefit{}).
			Where("level = ? AND benefit_type = ?", b.Level, b.BenefitType).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := r.db.Create(&b).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

type UserLevelRepository struct {
	db *gorm.DB
}

func NewUserLevelRepository(db *gorm.DB) *UserLevelRepository {
	return &UserLevelRepository{db: db}
}

func (r *UserLevelRepository) GetByUserID(userID uint) (*model.UserLevel, error) {
	var level model.UserLevel
	err := r.db.Where("user_id = ?", userID).First(&level).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &level, nil
}

func (r *UserLevelRepository) Create(level *model.UserLevel) error {
	return r.db.Create(level).Error
}

func (r *UserLevelRepository) Update(level *model.UserLevel) error {
	return r.db.Save(level).Error
}

func (r *UserLevelRepository) AddPoints(userID uint, points int) error {
	return r.db.Model(&model.UserLevel{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"points":       gorm.Expr("points + ?", points),
		"total_points": gorm.Expr("total_points + ?", points),
	}).Error
}

type PointsRuleRepository struct {
	db *gorm.DB
}

func NewPointsRuleRepository(db *gorm.DB) *PointsRuleRepository {
	return &PointsRuleRepository{db: db}
}

func (r *PointsRuleRepository) GetByType(pointType string) (*model.PointsRule, error) {
	var rule model.PointsRule
	if err := r.db.Where("type = ? AND is_active = 1", pointType).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *PointsRuleRepository) List() ([]model.PointsRule, error) {
	var rules []model.PointsRule
	if err := r.db.Where("is_active = 1").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *PointsRuleRepository) Create(rule *model.PointsRule) error {
	return r.db.Create(rule).Error
}

type LevelBenefitRepository struct {
	db *gorm.DB
}

func NewLevelBenefitRepository(db *gorm.DB) *LevelBenefitRepository {
	return &LevelBenefitRepository{db: db}
}

func (r *LevelBenefitRepository) ListByLevel(level int) ([]model.LevelBenefit, error) {
	var benefits []model.LevelBenefit
	if err := r.db.Where("level = ?", level).Find(&benefits).Error; err != nil {
		return nil, err
	}
	return benefits, nil
}

func (r *LevelBenefitRepository) ListAll() ([]model.LevelBenefit, error) {
	var benefits []model.LevelBenefit
	if err := r.db.Order("level ASC").Find(&benefits).Error; err != nil {
		return nil, err
	}
	return benefits, nil
}

type NotificationSettingsRepository struct {
	db *gorm.DB
}

func NewNotificationSettingsRepository(db *gorm.DB) *NotificationSettingsRepository {
	return &NotificationSettingsRepository{db: db}
}

func (r *NotificationSettingsRepository) GetByUserID(userID uint) (*model.NotificationSettings, error) {
	var settings model.NotificationSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *NotificationSettingsRepository) Create(settings *model.NotificationSettings) error {
	return r.db.Create(settings).Error
}

func (r *NotificationSettingsRepository) Update(settings *model.NotificationSettings) error {
	return r.db.Save(settings).Error
}
