package service

import (
	"errors"
	"time"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"
)

type BlockService struct {
	blockRepo  *repository.BlockRepository
	userRepo   *repository.UserRepository
	followRepo *repository.FollowRepository
}

func NewBlockService(blockRepo *repository.BlockRepository, userRepo *repository.UserRepository) *BlockService {
	return &BlockService{blockRepo: blockRepo, userRepo: userRepo}
}

func (s *BlockService) SetFollowRepository(followRepo *repository.FollowRepository) {
	s.followRepo = followRepo
}

func (s *BlockService) BlockUser(blockerID, blockedID uint) error {
	if blockerID == blockedID {
		return errors.New("不能拉黑自己")
	}

	exists, err := s.blockRepo.Exists(blockerID, blockedID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("已拉黑该用户")
	}

	block := &model.Block{
		BlockerID: blockerID,
		BlockedID: blockedID,
	}
	if err := s.blockRepo.Create(block); err != nil {
		return err
	}

	// F7：拉黑后自动双向取消关注
	if s.followRepo != nil {
		_ = s.followRepo.Delete(blockerID, blockedID)
		_ = s.followRepo.Delete(blockedID, blockerID)
	}

	return nil
}

func (s *BlockService) IsBlockedEither(userA, userB uint) (bool, error) {
	return s.blockRepo.ExistsEither(userA, userB)
}

func (s *BlockService) UnblockUser(blockerID, blockedID uint) error {
	return s.blockRepo.Delete(blockerID, blockedID)
}

func (s *BlockService) GetBlockedList(userID uint, page, pageSize int) ([]model.User, int64, error) {
	blocks, total, err := s.blockRepo.GetBlockedList(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	users := make([]model.User, 0, len(blocks))
	for _, block := range blocks {
		user, err := s.userRepo.GetByID(block.BlockedID)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, total, nil
}

func (s *BlockService) IsBlocked(blockerID, blockedID uint) (bool, error) {
	return s.blockRepo.Exists(blockerID, blockedID)
}

type MessageService struct {
	conversationRepo *repository.ConversationRepository
	messageRepo      *repository.MessageRepository
	userRepo         *repository.UserRepository
	blockRepo        *repository.BlockRepository
	followRepo       *repository.FollowRepository
	wsService        *WebSocketService
}

func NewMessageService(
	conversationRepo *repository.ConversationRepository,
	messageRepo *repository.MessageRepository,
	userRepo *repository.UserRepository,
	blockRepo *repository.BlockRepository,
) *MessageService {
	return &MessageService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		userRepo:         userRepo,
		blockRepo:        blockRepo,
	}
}

func (s *MessageService) SetFollowRepository(followRepo *repository.FollowRepository) {
	s.followRepo = followRepo
}

func (s *MessageService) SetWebSocketService(wsService *WebSocketService) {
	s.wsService = wsService
}

func (s *MessageService) SendMessage(fromID, toID uint, content string, msgType string) (*model.Message, error) {
	if fromID == toID {
		return nil, errors.New("不能给自己发消息")
	}

	blocked, err := s.blockRepo.Exists(toID, fromID)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, errors.New("对方已将你拉黑")
	}

	blocked2, err := s.blockRepo.Exists(fromID, toID)
	if err != nil {
		return nil, err
	}
	if blocked2 {
		return nil, errors.New("你已将对方拉黑")
	}

	// U8 私信权限：仅粉丝 = 发送方普通关注接收方（不含悄悄关注）
	if s.userRepo != nil {
		toUser, uErr := s.userRepo.GetByID(toID)
		if uErr == nil && toUser != nil {
			_, _, messagePerm := ParsePrivacySettings(toUser.PrivacySettings)
			switch messagePerm {
			case PermissionDenied:
				return nil, errors.New("对方已关闭私信")
			case PermissionFollowers:
				if s.followRepo == nil {
					return nil, errors.New("仅粉丝可私信")
				}
				ok, fErr := s.followRepo.ExistsPublic(fromID, toID)
				if fErr != nil {
					return nil, fErr
				}
				if !ok {
					return nil, errors.New("仅粉丝可私信")
				}
			}
		}
	}

	conversation, err := s.conversationRepo.GetByUsers(fromID, toID)
	if err != nil {
		return nil, err
	}

	if conversation == nil {
		conversation = &model.Conversation{
			User1ID: fromID,
			User2ID: toID,
		}
		if err := s.conversationRepo.Create(conversation); err != nil {
			return nil, err
		}
	}

	message := &model.Message{
		ConversationID: conversation.ID,
		FromID:         fromID,
		ToID:           toID,
		Content:        content,
		Type:           msgType,
		Status:         0,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	preview := content
	if len(content) > 20 {
		preview = content[:20] + "..."
	}

	now := time.Now()
	if err := s.conversationRepo.UpdateLastMessage(conversation.ID, message.ID, preview, &now); err != nil {
		return nil, err
	}

	if err := s.conversationRepo.IncrementUnread(conversation.ID); err != nil {
		return nil, err
	}
	cache.IncrMsgUnread(toID)

	if s.wsService != nil {
		s.wsService.SendMessage(toID, fromID, content, msgType)
	}

	return message, nil
}

func (s *MessageService) GetConversations(userID uint, page, pageSize int) ([]model.Conversation, int64, error) {
	return s.conversationRepo.ListByUser(userID, page, pageSize)
}

func (s *MessageService) GetMessages(conversationID uint, page, pageSize int) ([]model.Message, int64, error) {
	return s.messageRepo.ListByConversation(conversationID, page, pageSize)
}

func (s *MessageService) MarkAsRead(conversationID, userID uint) error {
	if err := s.messageRepo.MarkAsRead(conversationID, userID); err != nil {
		return err
	}
	if err := s.conversationRepo.ResetUnread(conversationID); err != nil {
		return err
	}
	cache.InvalidateMsgUnread(userID)
	return nil
}

func (s *MessageService) RecallMessage(messageID, userID uint) error {
	message, err := s.messageRepo.GetByID(messageID)
	if err != nil {
		return errors.New("消息不存在")
	}
	if message.FromID != userID {
		return errors.New("只能撤回自己发送的消息")
	}

	if time.Since(message.CreatedAt) > 2*time.Minute {
		return errors.New("消息已超过撤回时限")
	}

	return s.messageRepo.Recall(messageID)
}

func (s *MessageService) SearchMessages(conversationID uint, keyword string) ([]model.Message, error) {
	return s.messageRepo.Search(conversationID, keyword)
}

func (s *MessageService) DeleteConversation(conversationID, userID uint) error {
	return s.conversationRepo.Delete(conversationID, userID)
}

func (s *MessageService) GetUnreadCount(userID uint) (int64, error) {
	if n, ok := cache.GetMsgUnread(userID); ok {
		return n, nil
	}
	count, err := s.messageRepo.GetUnreadCount(userID)
	if err != nil {
		return 0, err
	}
	cache.SetMsgUnread(userID, count)
	return count, nil
}

type PointsService struct {
	pointsRepo       *repository.PointsRepository
	userLevelRepo    *repository.UserLevelRepository
	pointsRuleRepo   *repository.PointsRuleRepository
	levelBenefitRepo *repository.LevelBenefitRepository
}

func NewPointsService(
	pointsRepo *repository.PointsRepository,
	userLevelRepo *repository.UserLevelRepository,
	pointsRuleRepo *repository.PointsRuleRepository,
	levelBenefitRepo *repository.LevelBenefitRepository,
) *PointsService {
	return &PointsService{
		pointsRepo:       pointsRepo,
		userLevelRepo:    userLevelRepo,
		pointsRuleRepo:   pointsRuleRepo,
		levelBenefitRepo: levelBenefitRepo,
	}
}

const globalDailyPointsLimit = 100

func (s *PointsService) AddPoints(userID uint, pointType string, targetID uint, description string) error {
	rule, err := s.pointsRuleRepo.GetByType(pointType)
	if err != nil {
		return err
	}

	if rule.DailyLimit > 0 {
		todayPoints, err := s.pointsRepo.GetTodayPoints(userID, pointType)
		if err != nil {
			return err
		}
		if todayPoints >= rule.DailyLimit {
			return errors.New("今日积分已达上限")
		}
	}

	// 全局每日积分上限（规格 L2-06 默认 100）
	todayTotal, err := s.pointsRepo.GetTodayPointsTotal(userID)
	if err != nil {
		return err
	}
	if todayTotal >= globalDailyPointsLimit {
		return errors.New("今日积分已达上限")
	}
	award := rule.Points
	if todayTotal+award > globalDailyPointsLimit {
		award = globalDailyPointsLimit - todayTotal
	}
	if award <= 0 {
		return errors.New("今日积分已达上限")
	}

	level, err := s.userLevelRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	if level == nil {
		level = &model.UserLevel{
			UserID: userID,
			Level:  1,
			Points: 0,
		}
		if err := s.userLevelRepo.Create(level); err != nil {
			return err
		}
	}

	if err := s.userLevelRepo.AddPoints(userID, award); err != nil {
		return err
	}

	record := &model.PointsRecord{
		UserID:      userID,
		Points:      award,
		Type:        pointType,
		TargetID:    targetID,
		Description: description,
	}

	if err := s.pointsRepo.CreateRecord(record); err != nil {
		return err
	}

	newLevel := s.CalculateLevel(level.Points + award)
	if newLevel != level.Level {
		level.Level = newLevel
		level.Points = level.Points + award
		if err := s.userLevelRepo.Update(level); err != nil {
			return err
		}
	}

	return nil
}

// CalculateLevel 默认等级阈值：Lv1=0，每级 +200，至 Lv10=1800
func (s *PointsService) CalculateLevel(points int) int {
	level := points/200 + 1
	if level < 1 {
		return 1
	}
	if level > 10 {
		return 10
	}
	return level
}

func (s *PointsService) GetUserLevel(userID uint) (*model.UserLevel, error) {
	level, err := s.userLevelRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if level == nil {
		level = &model.UserLevel{
			UserID:      userID,
			Level:       1,
			Points:      0,
			TotalPoints: 0,
		}
		if err := s.userLevelRepo.Create(level); err != nil {
			return nil, err
		}
	}

	return level, nil
}

func (s *PointsService) GetPointsRecords(userID uint, page, pageSize int) ([]model.PointsRecord, int64, error) {
	return s.pointsRepo.ListByUser(userID, page, pageSize)
}

func (s *PointsService) GetPointsRules() ([]model.PointsRule, error) {
	return s.pointsRuleRepo.List()
}

func (s *PointsService) GetBenefits(level int) ([]model.LevelBenefit, error) {
	return s.levelBenefitRepo.ListByLevel(level)
}

func (s *PointsService) GetAllBenefits() ([]model.LevelBenefit, error) {
	return s.levelBenefitRepo.ListAll()
}

type NotificationSettingsService struct {
	settingsRepo *repository.NotificationSettingsRepository
}

func NewNotificationSettingsService(settingsRepo *repository.NotificationSettingsRepository) *NotificationSettingsService {
	return &NotificationSettingsService{settingsRepo: settingsRepo}
}

func (s *NotificationSettingsService) GetSettings(userID uint) (*model.NotificationSettings, error) {
	settings, err := s.settingsRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		settings = &model.NotificationSettings{
			UserID:        userID,
			LikeNotify:    1,
			CommentNotify: 1,
			AtNotify:      1,
			FollowNotify:  1,
			SystemNotify:  1,
		}
		if err := s.settingsRepo.Create(settings); err != nil {
			return nil, err
		}
	}

	return settings, nil
}

func (s *NotificationSettingsService) UpdateSettings(userID uint, likeNotify, commentNotify, atNotify, followNotify, systemNotify int) error {
	existing, err := s.settingsRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	if existing == nil {
		existing = &model.NotificationSettings{UserID: userID}
	}

	existing.LikeNotify = likeNotify
	existing.CommentNotify = commentNotify
	existing.AtNotify = atNotify
	existing.FollowNotify = followNotify
	existing.SystemNotify = systemNotify

	return s.settingsRepo.Update(existing)
}
