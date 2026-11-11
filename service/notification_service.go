package service

import (
	"errors"
	"fmt"
	"pet-community/model"
	"pet-community/repository"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NotificationService 通知服务
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	followRepo       *repository.FollowRepository
	userRepo         *repository.UserRepository
	settingsRepo     *repository.NotificationSettingsRepository
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	followRepo *repository.FollowRepository,
	userRepo *repository.UserRepository,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		followRepo:       followRepo,
		userRepo:         userRepo,
	}
}

// SetSettingsRepository 注入通知设置仓储，启用通知开关过滤（N8）。
func (s *NotificationService) SetSettingsRepository(settingsRepo *repository.NotificationSettingsRepository) {
	s.settingsRepo = settingsRepo
}

// NotifyAllowed 判断 userID 是否开启了 kind 对应的通知开关（N8-02~05, N8-07）。
//
// kind: like | comment | at | follow | system | message | security。
// repo 为 nil（未注入）、读取失败或用户未配置时视为开启——通知是纯副作用，
// 设置读取异常不应阻断主流程；用户主动关闭（值为 0）才拦截。
func NotifyAllowed(repo *repository.NotificationSettingsRepository, userID uint, kind string) bool {
	if repo == nil {
		return true
	}
	settings, err := repo.GetByUserID(userID)
	if err != nil || settings == nil {
		return true
	}
	var enabled int
	switch kind {
	case "like":
		enabled = settings.LikeNotify
	case "comment":
		enabled = settings.CommentNotify
	case "at":
		enabled = settings.AtNotify
	case "follow":
		enabled = settings.FollowNotify
	case "system":
		enabled = settings.SystemNotify
	case "message":
		enabled = settings.MessageNotify
	case "security":
		enabled = settings.SecurityNotify
	default:
		return true
	}
	return enabled == 1
}

// CreateNewDeviceLoginNotification R4：新设备登录站内通知（类型 security）。
// IP 归属地无本地库时降级为「归属地未知」；deviceName 空则「未知设备」。
func (s *NotificationService) CreateNewDeviceLoginNotification(userID uint, deviceName, ip string) error {
	if !NotifyAllowed(s.settingsRepo, userID, "security") {
		return nil
	}
	if strings.TrimSpace(deviceName) == "" {
		deviceName = "未知设备"
	}
	ipText := strings.TrimSpace(ip)
	if ipText == "" {
		ipText = "未知 IP"
	}
	content := fmt.Sprintf(
		"设备：%s；IP：%s（归属地未知）；时间：%s",
		deviceName,
		ipText,
		time.Now().Format("2006-01-02 15:04:05"),
	)
	notification := &model.UserNotification{
		UserID:      userID,
		Type:        "security",
		Title:       "新设备登录",
		Content:     content,
		TriggerType: "security",
	}
	return s.notificationRepo.Create(notification)
}

// CreateLikeNotification 创建点赞通知
func (s *NotificationService) CreateLikeNotification(triggerUserID, targetUserID uint, triggerType string, triggerID uint) error {
	if triggerUserID == targetUserID {
		return nil // 不通知自己
	}
	if !NotifyAllowed(s.settingsRepo, targetUserID, "like") {
		return nil // 用户已关闭点赞通知（N8-02）
	}

	triggerUser, _ := s.userRepo.GetByID(triggerUserID)

	var title, content string
	if triggerType == "post" {
		title = "点赞通知"
		content = fmt.Sprintf("%s 赞了你的动态", triggerUser.Nickname)
	} else {
		title = "点赞通知"
		content = fmt.Sprintf("%s 赞了你的评论", triggerUser.Nickname)
	}

	notification := &model.UserNotification{
		UserID:        targetUserID,
		Type:          "like",
		Title:         title,
		Content:       content,
		TriggerType:   triggerType,
		TriggerID:     triggerID,
		TriggerUserID: triggerUserID,
	}

	return s.notificationRepo.Create(notification)
}

// CreateCommentNotification 创建评论通知
func (s *NotificationService) CreateCommentNotification(triggerUserID, targetUserID uint, postID uint, commentContent string) error {
	if triggerUserID == targetUserID {
		return nil
	}
	if !NotifyAllowed(s.settingsRepo, targetUserID, "comment") {
		return nil // 用户已关闭评论通知（N8-03）
	}

	triggerUser, _ := s.userRepo.GetByID(triggerUserID)

	notification := &model.UserNotification{
		UserID:        targetUserID,
		Type:          "comment",
		Title:         "评论通知",
		Content:       fmt.Sprintf("%s 评论了你的动态：%s", triggerUser.Nickname, truncateContent(commentContent, 50)),
		TriggerType:   "post",
		TriggerID:     postID,
		TriggerUserID: triggerUserID,
	}

	return s.notificationRepo.Create(notification)
}

// CreateReplyNotification 创建回复通知
func (s *NotificationService) CreateReplyNotification(triggerUserID, targetUserID uint, postID uint, commentID uint, commentContent string) error {
	if triggerUserID == targetUserID {
		return nil
	}
	if !NotifyAllowed(s.settingsRepo, targetUserID, "comment") {
		return nil // 回复计入评论通知开关（N8-03）
	}

	triggerUser, _ := s.userRepo.GetByID(triggerUserID)

	notification := &model.UserNotification{
		UserID:        targetUserID,
		Type:          "comment",
		Title:         "回复通知",
		Content:       fmt.Sprintf("%s 回复了你的评论：%s", triggerUser.Nickname, truncateContent(commentContent, 50)),
		TriggerType:   "comment",
		TriggerID:     commentID,
		TriggerUserID: triggerUserID,
	}

	return s.notificationRepo.Create(notification)
}

// CreateAtNotification 创建@用户通知
func (s *NotificationService) CreateAtNotification(triggerUserID, targetUserID uint, triggerType string, triggerID uint, content string) error {
	if triggerUserID == targetUserID {
		return nil
	}
	if !NotifyAllowed(s.settingsRepo, targetUserID, "at") {
		return nil // 用户已关闭@通知（N8-04）
	}

	triggerUser, _ := s.userRepo.GetByID(triggerUserID)

	notification := &model.UserNotification{
		UserID:        targetUserID,
		Type:          "at",
		Title:         "@通知",
		Content:       fmt.Sprintf("%s 在动态中提到了你：%s", triggerUser.Nickname, truncateContent(content, 50)),
		TriggerType:   triggerType,
		TriggerID:     triggerID,
		TriggerUserID: triggerUserID,
	}

	return s.notificationRepo.Create(notification)
}

// ParseAtUsers 解析内容中的@用户，返回用户ID列表
func (s *NotificationService) ParseAtUsers(content string) []uint {
	re := regexp.MustCompile(`@(\S+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	userIDs := make([]uint, 0)
	for _, match := range matches {
		if len(match) > 1 {
			nickname := match[1]
			user, err := s.userRepo.GetByNickname(nickname)
			if err == nil {
				userIDs = append(userIDs, user.ID)
			}
		}
	}
	return userIDs
}

// SendAtNotifications 发送@用户通知给所有被@的用户
func (s *NotificationService) SendAtNotifications(triggerUserID uint, triggerType string, triggerID uint, content string, atUserIDs []uint) error {
	for _, userID := range atUserIDs {
		if err := s.CreateAtNotification(triggerUserID, userID, triggerType, triggerID, content); err != nil {
			return err
		}
	}
	return nil
}

// CheckVisibility 检查用户是否有权限查看动态（仅普通关注计入粉丝门禁）
func (s *NotificationService) CheckVisibility(viewerID, authorID uint, visibility int) bool {
	isPublicFollower := false
	if viewerID > 0 && viewerID != authorID && s.followRepo != nil {
		isPublicFollower, _ = s.followRepo.ExistsPublic(viewerID, authorID)
	}
	return CanViewPost(viewerID, authorID, visibility, isPublicFollower)
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

// FollowService 粉丝服务
type FollowService struct {
	followRepo       *repository.FollowRepository
	userRepo         *repository.UserRepository
	notificationRepo *repository.NotificationRepository
	settingsRepo     *repository.NotificationSettingsRepository
	blockRepo        *repository.BlockRepository
	pointsService    *PointsService
	presence         OnlinePresence
}

func NewFollowService(
	followRepo *repository.FollowRepository,
	userRepo *repository.UserRepository,
	notificationRepo *repository.NotificationRepository,
) *FollowService {
	return &FollowService{
		followRepo:       followRepo,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
	}
}

// SetNotificationSettingsRepository 注入通知设置仓储，启用关注通知开关过滤（N8-05）。
func (s *FollowService) SetNotificationSettingsRepository(settingsRepo *repository.NotificationSettingsRepository) {
	s.settingsRepo = settingsRepo
}

func (s *FollowService) SetBlockRepository(blockRepo *repository.BlockRepository) {
	s.blockRepo = blockRepo
}

func (s *FollowService) SetPointsService(ps *PointsService) {
	s.pointsService = ps
}

// SetPresence 注入在线连接态来源（WebSocket 连接表）；未注入时退化为仅按登录时间判断。
func (s *FollowService) SetPresence(p OnlinePresence) {
	s.presence = p
}

func (s *FollowService) Unfollow(followerID, followingID uint) error {
	exists, err := s.followRepo.Exists(followerID, followingID)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return s.followRepo.Delete(followerID, followingID)
}

func (s *FollowService) ToggleFollow(followerID, followingID uint, isSecret bool) (bool, error) {
	if followerID == followingID {
		return false, fmt.Errorf("不能关注自己")
	}

	if s.blockRepo != nil {
		blocked, err := s.blockRepo.ExistsEither(followerID, followingID)
		if err != nil {
			return false, err
		}
		if blocked {
			return false, fmt.Errorf("无法互动")
		}
	}

	exists, err := s.followRepo.Exists(followerID, followingID)
	if err != nil {
		return false, err
	}

	if exists {
		if err := s.followRepo.Delete(followerID, followingID); err != nil {
			return false, err
		}
		return false, nil
	}

	secret := 0
	if isSecret {
		secret = 1
	}

	follow := &model.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		Status:      1,
		IsSecret:    secret,
	}

	if err := s.followRepo.Create(follow); err != nil {
		return false, err
	}

	if s.pointsService != nil && !isSecret {
		go s.pointsService.AddPoints(followingID, "follow_received", followerID, "被用户关注")
	}

	// 发送关注通知
	if s.notificationRepo != nil && !isSecret && NotifyAllowed(s.settingsRepo, followingID, "follow") {
		triggerUser, _ := s.userRepo.GetByID(followerID)
		if triggerUser != nil {
			notification := &model.UserNotification{
				UserID:        followingID,
				Type:          "follow",
				Title:         "关注通知",
				Content:       fmt.Sprintf("%s 关注了你", triggerUser.Nickname),
				TriggerType:   "user",
				TriggerID:     followerID,
				TriggerUserID: followerID,
			}
			s.notificationRepo.Create(notification)
		}
	}

	return true, nil
}

// SetSecretFollow 设置悄悄关注
func (s *FollowService) SetSecretFollow(followerID, followingID uint, isSecret bool) error {
	exists, err := s.followRepo.Exists(followerID, followingID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("未关注该用户")
	}

	secret := 0
	if isSecret {
		secret = 1
	}
	return s.followRepo.SetSecret(followerID, followingID, secret)
}

// GetSecretFollowing 获取悄悄关注列表
func (s *FollowService) GetSecretFollowing(userID uint, page, pageSize int) ([]model.User, int64, error) {
	follows, total, err := s.followRepo.GetSecretFollowing(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	users := make([]model.User, 0, len(follows))
	for _, f := range follows {
		user, err := s.userRepo.GetByID(f.FollowingID)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, total, nil
}

// GetOnlineStatus 获取用户在线状态。
//
// 1=在线：当前持有 WebSocket 连接（连接态）；
// 2=最近活跃：24 小时内登录过，但当前无连接；
// 0=离线。
//
// 旧实现仅按 lastLoginAt 5 分钟内判定「在线」，与连接无关，会让已退出的用户长期显示在线。
func (s *FollowService) GetOnlineStatus(userID uint) (int, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return 0, err
	}

	if s.presence != nil && s.presence.IsOnline(userID) {
		return 1, nil
	}

	if user.LastLoginAt != nil && time.Since(*user.LastLoginAt) < 24*time.Hour {
		return 2, nil // 最近活跃
	}
	return 0, nil // 离线
}

func (s *FollowService) IsFollowing(followerID, followingID uint) (bool, error) {
	return s.followRepo.Exists(followerID, followingID)
}

// IsSecretFollow 判断 followerID 对 followingID 的关注是否为悄悄关注（F6）。
func (s *FollowService) IsSecretFollow(followerID, followingID uint) bool {
	f, err := s.followRepo.GetFollow(followerID, followingID)
	if err != nil || f == nil {
		return false
	}
	return f.IsSecret == 1
}

// GetFollowState 返回 viewerID 对 targetID 的关注状态（含悄悄关注、互关标识）。
// 用于个人主页关注按钮渲染：isFollowing / isSecret / isMutual（F1/F5/F6）。
func (s *FollowService) GetFollowState(viewerID, targetID uint) (isFollowing, isSecret, isMutual bool, err error) {
	isFollowing, err = s.followRepo.Exists(viewerID, targetID)
	if err != nil || !isFollowing {
		return
	}
	if f, e := s.followRepo.GetFollow(viewerID, targetID); e == nil && f.IsSecret == 1 {
		isSecret = true
	}
	if viewerID > 0 {
		isMutual, _ = s.followRepo.Exists(targetID, viewerID)
	}
	return
}

func (s *FollowService) GetFollowers(userID uint, page, pageSize int) ([]model.User, int64, error) {
	follows, total, err := s.followRepo.GetFollowers(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	users := make([]model.User, 0, len(follows))
	for _, f := range follows {
		user, err := s.userRepo.GetByID(f.FollowerID)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, total, nil
}

func (s *FollowService) GetFollowing(userID uint, page, pageSize int) ([]model.User, int64, error) {
	follows, total, err := s.followRepo.GetFollowing(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	users := make([]model.User, 0, len(follows))
	for _, f := range follows {
		user, err := s.userRepo.GetByID(f.FollowingID)
		if err == nil {
			users = append(users, *user)
		}
	}

	return users, total, nil
}

func (s *FollowService) GetFollowersCount(userID uint) (int64, error) {
	return s.followRepo.GetFollowersCount(userID)
}

func (s *FollowService) GetFollowingCount(userID uint) (int64, error) {
	return s.followRepo.GetFollowingCount(userID)
}

// ParseAtUsersFromString 从字符串解析@用户ID（如 "1,2,3"）
func ParseAtUsersFromString(atUsersStr string) []uint {
	if atUsersStr == "" {
		return []uint{}
	}

	userIDs := make([]uint, 0)
	for _, idStr := range strings.Split(atUsersStr, ",") {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			userIDs = append(userIDs, uint(id))
		}
	}
	return userIDs
}
