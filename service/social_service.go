package service

import (
	"errors"
	"strings"
	"time"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/utils"
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

// SendMessage 发送纯文本消息（保留旧签名，兼容已有调用方与测试）
func (s *MessageService) SendMessage(fromID, toID uint, content string, msgType string) (*model.Message, error) {
	return s.SendMessageWithMedia(fromID, toID, content, msgType, "", 0)
}

// SendMessageWithMedia 发送消息，支持图片/语音等媒体消息。
//
// mediaURL 必须来自私信上传接口；duration 仅对语音消息生效（秒）。
func (s *MessageService) SendMessageWithMedia(fromID, toID uint, content string, msgType string, mediaURL string, duration int) (*model.Message, error) {
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

	// 归一化消息类型并按类型做参数校验（校验前置，避免非法请求先建会话）
	msgType = strings.ToLower(strings.TrimSpace(msgType))
	if msgType == "" {
		msgType = model.MessageTypeText
	}
	content = strings.TrimSpace(content)
	mediaURL = strings.TrimSpace(mediaURL)

	switch msgType {
	case model.MessageTypeText:
		if content == "" {
			return nil, errors.New("消息内容不能为空")
		}
		mediaURL = ""
		duration = 0
	case model.MessageTypeImage:
		if err := ValidateMessageMediaURL(mediaURL, MessageImageCategory); err != nil {
			return nil, errors.New("图片地址无效")
		}
		duration = 0
		if content == "" {
			content = "[图片]"
		}
	case model.MessageTypeVoice:
		if err := ValidateMessageMediaURL(mediaURL, MessageVoiceCategory); err != nil {
			return nil, errors.New("语音地址无效")
		}
		if err := utils.ValidateVoiceDuration(duration); err != nil {
			return nil, err
		}
		if content == "" {
			content = "[语音]"
		}
	default:
		return nil, errors.New("不支持的消息类型")
	}

	if err := ValidateMessageContentLength(content); err != nil {
		return nil, err
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
		MediaURL:       mediaURL,
		Duration:       duration,
		Status:         0,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, err
	}

	preview := buildMessagePreview(msgType, content)

	now := time.Now()
	if err := s.conversationRepo.UpdateLastMessage(conversation.ID, message.ID, preview, &now); err != nil {
		return nil, err
	}

	if err := s.conversationRepo.IncrementUnread(conversation.ID); err != nil {
		return nil, err
	}
	cache.IncrMsgUnread(toID)

	if s.wsService != nil {
		s.wsService.SendMessage(message)
	}

	return message, nil
}

func (s *MessageService) GetConversations(userID uint, page, pageSize int) ([]model.Conversation, int64, error) {
	return s.conversationRepo.ListByUser(userID, page, pageSize)
}

// GetMessages 拉取会话消息（仅会话成员可读）
func (s *MessageService) GetMessages(conversationID, userID uint, page, pageSize int) ([]model.Message, int64, error) {
	if err := s.EnsureConversationMember(conversationID, userID); err != nil {
		return nil, 0, err
	}
	return s.messageRepo.ListByConversation(conversationID, page, pageSize)
}

// EnsureConversationMember 校验用户是否属于该会话，防止越权读取/清空他人会话
func (s *MessageService) EnsureConversationMember(conversationID, userID uint) error {
	if conversationID == 0 || userID == 0 {
		return errors.New("无权访问该会话")
	}
	conv, err := s.conversationRepo.GetByID(conversationID)
	if err != nil || conv == nil {
		return errors.New("会话不存在")
	}
	if conv.User1ID != userID && conv.User2ID != userID {
		return errors.New("无权访问该会话")
	}
	return nil
}

func (s *MessageService) MarkAsRead(conversationID, userID uint) error {
	if err := s.EnsureConversationMember(conversationID, userID); err != nil {
		return err
	}
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

// SearchMessages 会话内搜索文本消息（仅会话成员可搜，支持分页）
func (s *MessageService) SearchMessages(conversationID, userID uint, keyword string, page, pageSize int) ([]model.Message, int64, error) {
	if err := s.EnsureConversationMember(conversationID, userID); err != nil {
		return nil, 0, err
	}
	return s.messageRepo.Search(conversationID, keyword, page, pageSize)
}

// buildMessagePreview 生成会话列表最后一条消息预览
func buildMessagePreview(msgType, content string) string {
	switch msgType {
	case model.MessageTypeImage:
		return "[图片]"
	case model.MessageTypeVoice:
		return "[语音]"
	}
	runes := []rune(content)
	if len(runes) > 20 {
		return string(runes[:20]) + "..."
	}
	return content
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
