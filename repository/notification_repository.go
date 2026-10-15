package repository

import (
	"pet-community/cache"
	"pet-community/model"

	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *model.UserNotification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return err
	}
	cache.IncrNotifUnread(notification.UserID)
	return nil
}

func (r *NotificationRepository) GetByID(id uint) (*model.UserNotification, error) {
	var notification model.UserNotification
	if err := r.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) ListByUser(userID uint, page, pageSize int) ([]model.UserNotification, int64, error) {
	var notifications []model.UserNotification
	var total int64
	query := r.db.Model(&model.UserNotification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *NotificationRepository) MarkAsRead(userID, notificationID uint) error {
	if err := r.db.Model(&model.UserNotification{}).Where("id = ? AND user_id = ?", notificationID, userID).Update("is_read", 1).Error; err != nil {
		return err
	}
	cache.InvalidateNotifUnread(userID)
	return nil
}

func (r *NotificationRepository) MarkAllAsRead(userID uint) error {
	if err := r.db.Model(&model.UserNotification{}).Where("user_id = ? AND is_read = 0", userID).Update("is_read", 1).Error; err != nil {
		return err
	}
	cache.SetNotifUnread(userID, 0)
	return nil
}

func (r *NotificationRepository) GetUnreadCount(userID uint) (int64, error) {
	if n, ok := cache.GetNotifUnread(userID); ok {
		return n, nil
	}
	var count int64
	if err := r.db.Model(&model.UserNotification{}).Where("user_id = ? AND is_read = 0", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	cache.SetNotifUnread(userID, count)
	return count, nil
}

func (r *NotificationRepository) Delete(userID, notificationID uint) error {
	return r.db.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&model.UserNotification{}).Error
}

type FollowRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) *FollowRepository {
	return &FollowRepository{db: db}
}

func (r *FollowRepository) Create(follow *model.Follow) error {
	if err := r.db.Create(follow).Error; err != nil {
		return err
	}
	cache.SetFollowEdge(follow.FollowerID, follow.FollowingID, true)
	return nil
}

func (r *FollowRepository) Delete(followerID, followingID uint) error {
	if err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.Follow{}).Error; err != nil {
		return err
	}
	cache.SetFollowEdge(followerID, followingID, false)
	return nil
}

func (r *FollowRepository) Exists(followerID, followingID uint) (bool, error) {
	if exists, hit := cache.GetFollowEdge(followerID, followingID); hit {
		return exists, nil
	}
	var count int64
	if err := r.db.Model(&model.Follow{}).Where("follower_id = ? AND following_id = ? AND status = 1", followerID, followingID).Count(&count).Error; err != nil {
		return false, err
	}
	exists := count > 0
	cache.SetFollowEdge(followerID, followingID, exists)
	return exists, nil
}

// GetFollow 取关注关系记录（含 IsSecret）；不存在返回 gorm.ErrRecordNotFound。
// 用于个人主页关注按钮渲染「悄悄关注」状态。
func (r *FollowRepository) GetFollow(followerID, followingID uint) (*model.Follow, error) {
	var f model.Follow
	if err := r.db.Where("follower_id = ? AND following_id = ? AND status = 1", followerID, followingID).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// ExistsPublic 普通关注是否存在（排除悄悄关注，用于仅粉丝门禁）
func (r *FollowRepository) ExistsPublic(followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Follow{}).Where(
		"follower_id = ? AND following_id = ? AND status = 1 AND is_secret = 0",
		followerID, followingID,
	).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FollowRepository) GetFollowers(userID uint, page, pageSize int) ([]model.Follow, int64, error) {
	var follows []model.Follow
	var total int64
	query := r.db.Model(&model.Follow{}).Where("following_id = ? AND status = 1 AND is_secret = 0", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

func (r *FollowRepository) GetFollowing(userID uint, page, pageSize int) ([]model.Follow, int64, error) {
	var follows []model.Follow
	var total int64
	query := r.db.Model(&model.Follow{}).Where("follower_id = ? AND status = 1", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

func (r *FollowRepository) GetSecretFollowing(userID uint, page, pageSize int) ([]model.Follow, int64, error) {
	var follows []model.Follow
	var total int64
	query := r.db.Model(&model.Follow{}).Where("follower_id = ? AND status = 1 AND is_secret = 1", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	return follows, total, nil
}

func (r *FollowRepository) SetSecret(followerID, followingID uint, isSecret int) error {
	return r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND following_id = ? AND status = 1", followerID, followingID).
		Update("is_secret", isSecret).Error
}

func (r *FollowRepository) GetFollowersCount(userID uint) (int64, error) {
	var count int64
	// 悄悄关注不计入粉丝数（§4.7 / F6-01）
	if err := r.db.Model(&model.Follow{}).Where(
		"following_id = ? AND status = 1 AND is_secret = 0", userID,
	).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *FollowRepository) GetFollowingCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Follow{}).Where("follower_id = ? AND status = 1", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *FollowRepository) GetFollowingIDs(userID uint) ([]uint, error) {
	var follows []model.Follow
	if err := r.db.Where("follower_id = ? AND status = 1", userID).Find(&follows).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.FollowingID)
	}
	return ids, nil
}

// GetPublicFollowingIDs 普通关注对象 ID 列表（不含悄悄关注）
func (r *FollowRepository) GetPublicFollowingIDs(userID uint) ([]uint, error) {
	var follows []model.Follow
	if err := r.db.Where("follower_id = ? AND status = 1 AND is_secret = 0", userID).Find(&follows).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.FollowingID)
	}
	return ids, nil
}
