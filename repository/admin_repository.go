package repository

import (
	"pet-community/model"
	"time"

	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (r *AdminRepository) Create(admin *model.AdminUser) error {
	return r.db.Create(admin).Error
}

func (r *AdminRepository) FindByID(id uint) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.First(&admin, id).Error
	return &admin, err
}

func (r *AdminRepository) FindByUsername(username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	result := r.db.Where("username = ?", username).First(&admin)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 用户不存在时直接返回错误，不包装
			return nil, result.Error
		}
		// 其他数据库错误
		return nil, result.Error
	}
	return &admin, nil
}

func (r *AdminRepository) List(keyword string, page, pageSize int) ([]model.AdminUser, int64, error) {
	var admins []model.AdminUser
	var total int64

	query := r.db.Model(&model.AdminUser{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&admins).Error
	return admins, total, err
}

func (r *AdminRepository) Update(admin *model.AdminUser) error {
	return r.db.Save(admin).Error
}

func (r *AdminRepository) Delete(id uint) error {
	return r.db.Delete(&model.AdminUser{}, id).Error
}

func (r *AdminRepository) UpdateStatus(id uint, status int) error {
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).Update("status", status).Error
}

func (r *AdminRepository) GetSystemConfig(key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	if err := r.db.Where("config_key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *AdminRepository) UpdateSystemConfig(key, value string) error {
	return r.db.Model(&model.SystemConfig{}).Where("config_key = ?", key).Update("config_value", value).Error
}

func (r *AdminRepository) BatchUpdateSystemConfig(configs map[string]string) error {
	tx := r.db.Begin()
	for key, value := range configs {
		if err := tx.Model(&model.SystemConfig{}).Where("config_key = ?", key).Update("config_value", value).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *AdminRepository) SearchSensitiveWords(keyword string, page, limit int) ([]model.SensitiveWord, int64, error) {
	var words []model.SensitiveWord
	var total int64
	query := r.db.Model(&model.SensitiveWord{})

	if keyword != "" {
		query = query.Where("word LIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset((page - 1) * limit).Order("created_at DESC").Find(&words).Error; err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

func (r *AdminRepository) CreateSensitiveWord(word *model.SensitiveWord) error {
	return r.db.Create(word).Error
}

func (r *AdminRepository) UpdateSensitiveWord(word *model.SensitiveWord) error {
	return r.db.Save(word).Error
}

func (r *AdminRepository) DeleteSensitiveWord(id uint) error {
	return r.db.Delete(&model.SensitiveWord{}, id).Error
}

func (r *AdminRepository) CheckSensitiveWordExists(word string) (bool, error) {
	var count int64
	if err := r.db.Model(&model.SensitiveWord{}).Where("word = ?", word).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AdminRepository) GetSensitiveWordByID(id uint) (*model.SensitiveWord, error) {
	var word model.SensitiveWord
	if err := r.db.Where("id = ?", id).First(&word).Error; err != nil {
		return nil, err
	}
	return &word, nil
}

// ListActiveSensitiveWords 启用中的敏感词（用于发布过滤）
func (r *AdminRepository) ListActiveSensitiveWords() ([]model.SensitiveWord, error) {
	var words []model.SensitiveWord
	err := r.db.Where("status = 1").Order("id ASC").Find(&words).Error
	return words, err
}

func (r *AdminRepository) CreateOperationLog(log *model.AdminOperationLog) error {
	return r.db.Create(log).Error
}

func (r *AdminRepository) GetOperationLogs(adminID uint, action string, startTime, endTime time.Time, page, limit int) ([]model.AdminOperationLog, int64, error) {
	var logs []model.AdminOperationLog
	var total int64
	query := r.db.Model(&model.AdminOperationLog{})

	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
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

func (r *AdminRepository) CreateActivityLog(log *model.UserActivityLog) error {
	return r.db.Create(log).Error
}

func (r *AdminRepository) GetActivityLogs(userID uint, action string, startTime, endTime time.Time, page, limit int) ([]model.UserActivityLog, int64, error) {
	var logs []model.UserActivityLog
	var total int64
	query := r.db.Model(&model.UserActivityLog{})

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
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

func (r *AdminRepository) CreateLoginLog(log *model.AdminLoginLog) error {
	return r.db.Create(log).Error
}

func (r *AdminRepository) GetLoginLogs(adminID uint, status int, startTime, endTime time.Time, page, limit int) ([]model.AdminLoginLog, int64, error) {
	var logs []model.AdminLoginLog
	var total int64
	query := r.db.Model(&model.AdminLoginLog{})

	if adminID > 0 {
		query = query.Where("admin_id = ?", adminID)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
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

func (r *AdminRepository) CreateNotification(notification *model.SystemNotification) error {
	return r.db.Create(notification).Error
}

func (r *AdminRepository) UpdateNotification(notification *model.SystemNotification) error {
	return r.db.Save(notification).Error
}

func (r *AdminRepository) DeleteNotification(id uint) error {
	return r.db.Delete(&model.SystemNotification{}, id).Error
}

func (r *AdminRepository) GetNotificationByID(id uint) (*model.SystemNotification, error) {
	var notification model.SystemNotification
	if err := r.db.First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *AdminRepository) ListNotifications(status int, page, limit int) ([]model.SystemNotification, int64, error) {
	var notifications []model.SystemNotification
	var total int64
	query := r.db.Model(&model.SystemNotification{})

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset((page - 1) * limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *AdminRepository) CreateAnnouncement(announcement *model.Announcement) error {
	return r.db.Create(announcement).Error
}

func (r *AdminRepository) UpdateAnnouncement(announcement *model.Announcement) error {
	return r.db.Save(announcement).Error
}

func (r *AdminRepository) DeleteAnnouncement(id uint) error {
	return r.db.Delete(&model.Announcement{}, id).Error
}

func (r *AdminRepository) GetAnnouncementByID(id uint) (*model.Announcement, error) {
	var announcement model.Announcement
	if err := r.db.First(&announcement, id).Error; err != nil {
		return nil, err
	}
	return &announcement, nil
}

func (r *AdminRepository) ListAnnouncements(status int, page, limit int) ([]model.Announcement, int64, error) {
	var announcements []model.Announcement
	var total int64
	query := r.db.Model(&model.Announcement{})

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset((page - 1) * limit).Order("sort_order DESC, created_at DESC").Find(&announcements).Error; err != nil {
		return nil, 0, err
	}

	return announcements, total, nil
}

func (r *AdminRepository) GetActiveAnnouncements() ([]model.Announcement, error) {
	var announcements []model.Announcement
	now := time.Now()
	err := r.db.Where("status = 1 AND (start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", now, now).
		Order("sort_order DESC, created_at DESC").
		Find(&announcements).Error
	return announcements, err
}

// GetFrontUserOperationLogs 查询针对前台用户的操作日志（target_type=user, target_id=userID）
func (r *AdminRepository) GetFrontUserOperationLogs(userID uint, action string, startTime, endTime time.Time, page, limit int) ([]model.AdminOperationLog, int64, error) {
	var logs []model.AdminOperationLog
	var total int64
	query := r.db.Model(&model.AdminOperationLog{})

	if userID > 0 {
		query = query.Where("target_type = ? AND target_id = ?", "user", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
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

// ListPendingPosts 返回待审核帖子（status=0）及其总数
func (r *AdminRepository) ListPendingPosts(page, limit int) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	query := r.db.Model(&model.Post{}).Where("status = 0")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// ListPendingComments 返回待审核评论（status=0）及其总数
func (r *AdminRepository) ListPendingComments(page, limit int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64
	query := r.db.Model(&model.Comment{}).Where("status = 0")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

// GetCommentByID 按 ID 查询评论
func (r *AdminRepository) GetCommentByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetPostByID 按 ID 查询帖子
func (r *AdminRepository) GetPostByID(id uint) (*model.Post, error) {
	var post model.Post
	if err := r.db.First(&post, id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

// UpdatePostStatus 更新帖子状态
func (r *AdminRepository) UpdatePostStatus(postID uint, status int) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Update("status", status).Error
}

// UpdateCommentStatus 更新评论状态
func (r *AdminRepository) UpdateCommentStatus(commentID uint, status int) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", commentID).Update("status", status).Error
}

// ListPendingTopics 返回待审核话题（status=0）
func (r *AdminRepository) ListPendingTopics(page, limit int) ([]model.Topic, int64, error) {
	var topics []model.Topic
	var total int64
	query := r.db.Model(&model.Topic{}).Where("status = 0")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Offset((page - 1) * limit).Limit(limit).Order("created_at DESC").Find(&topics).Error; err != nil {
		return nil, 0, err
	}
	return topics, total, nil
}

// UpdateTopicStatus 更新话题状态（0待审 1通过 3拒绝）
func (r *AdminRepository) UpdateTopicStatus(topicID uint, status int) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", topicID).Update("status", status).Error
}

// GetReviewHistory 查询审核历史（action 为 approve_review / reject_review）
func (r *AdminRepository) GetReviewHistory(targetType string, page, pageSize int) ([]model.AdminOperationLog, int64, error) {
	var logs []model.AdminOperationLog
	var total int64
	query := r.db.Model(&model.AdminOperationLog{}).
		Where("action IN ('approve_review', 'reject_review')")

	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListNotificationsByFilter 按类型/状态分页查询系统通知
func (r *AdminRepository) ListNotificationsByFilter(notificationType, status int, page, limit int) ([]model.SystemNotification, int64, error) {
	var notifications []model.SystemNotification
	var total int64
	query := r.db.Model(&model.SystemNotification{})

	if notificationType > 0 {
		query = query.Where("type = ?", notificationType)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset((page - 1) * limit).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// ListActiveUsers 返回所有状态正常（status=1）的用户
func (r *AdminRepository) ListActiveUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("status = 1").Find(&users).Error
	return users, err
}

// CreateUserNotification 创建用户通知
func (r *AdminRepository) CreateUserNotification(n *model.UserNotification) error {
	return r.db.Create(n).Error
}

// ListScheduledNotifications 返回已到发布时间但尚未发送的系统通知
func (r *AdminRepository) ListScheduledNotifications() ([]model.SystemNotification, error) {
	var notifications []model.SystemNotification
	now := time.Now()
	err := r.db.Where("status = 0 AND publish_at IS NOT NULL AND publish_at <= ?", now).
		Find(&notifications).Error
	return notifications, err
}

// GetMaxAnnouncementSortOrder 获取公告最大排序值
func (r *AdminRepository) GetMaxAnnouncementSortOrder() (int, error) {
	var maxSortOrder int
	err := r.db.Model(&model.Announcement{}).Select("MAX(sort_order)").Scan(&maxSortOrder).Error
	return maxSortOrder, err
}

// CountUsersByStatus 按状态统计用户数
func (r *AdminRepository) CountUsersByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountTodayActiveUsers 统计今日活跃用户（status=1 且今日登录过）
func (r *AdminRepository) CountTodayActiveUsers() (int64, error) {
	return r.CountActiveUsersByDate(time.Now().Format("2006-01-02"))
}

// CountActiveUsersByDate 统计指定日期活跃用户
func (r *AdminRepository) CountActiveUsersByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Where("status = 1 AND DATE(last_login_at) = ?", date).
		Count(&count).Error
	return count, err
}

// CountAllPosts 统计帖子总数
func (r *AdminRepository) CountAllPosts() (int64, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Count(&count).Error
	return count, err
}

// CountPostsByStatus 按状态统计帖子数
func (r *AdminRepository) CountPostsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountPostsByCreatedDate 统计指定日期创建的帖子数
func (r *AdminRepository) CountPostsByCreatedDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Post{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// CountAllPets 统计宠物总数
func (r *AdminRepository) CountAllPets() (int64, error) {
	var count int64
	err := r.db.Model(&model.Pet{}).Count(&count).Error
	return count, err
}

// CountNewUsersByDate 统计指定日期新增的活跃用户数（DATE(created_at)=date AND status=1）
func (r *AdminRepository) CountNewUsersByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Where("DATE(created_at) = ? AND status = 1", date).
		Count(&count).Error
	return count, err
}

// CountAdvertisementsByStatus 按状态统计广告数
func (r *AdminRepository) CountAdvertisementsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.Advertisement{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountSystemNotificationsByStatus 按状态统计系统通知数
func (r *AdminRepository) CountSystemNotificationsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.SystemNotification{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountLikesByDate 统计指定日期新增点赞数
func (r *AdminRepository) CountLikesByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// CountCommentsByDate 统计指定日期新增评论数
func (r *AdminRepository) CountCommentsByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// CountAllLikes 统计点赞总数
func (r *AdminRepository) CountAllLikes() (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).Count(&count).Error
	return count, err
}

// CountAdoptionApplicationsByStatus 按状态统计领养申请数
func (r *AdminRepository) CountAdoptionApplicationsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.AdoptionApplication{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountOngoingActivities 统计进行中活动数（status=1）
func (r *AdminRepository) CountOngoingActivities() (int64, error) {
	var count int64
	err := r.db.Model(&model.Activity{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

// CountFavoritesByDate 统计指定日期新增收藏数
func (r *AdminRepository) CountFavoritesByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Favorite{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// CountSharesByDate 统计指定日期分享数
func (r *AdminRepository) CountSharesByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.ShareRecord{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// CountRegistrationsByDate 统计指定日期活动报名数
func (r *AdminRepository) CountRegistrationsByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Registration{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// CountRegistrationsByStatus 按状态统计活动报名数
func (r *AdminRepository) CountRegistrationsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.Registration{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountVotesByDate 统计指定日期活动投票数
func (r *AdminRepository) CountVotesByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Vote{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

// SumAdvertisementViews 累计广告曝光
func (r *AdminRepository) SumAdvertisementViews() (int64, error) {
	var sum int64
	err := r.db.Model(&model.Advertisement{}).Select("COALESCE(SUM(view_count), 0)").Scan(&sum).Error
	return sum, err
}

// SumAdvertisementClicks 累计广告点击
func (r *AdminRepository) SumAdvertisementClicks() (int64, error) {
	var sum int64
	err := r.db.Model(&model.Advertisement{}).Select("COALESCE(SUM(click_count), 0)").Scan(&sum).Error
	return sum, err
}

// CountLostPetsByStatus 按状态统计寻宠启事（1=寻找中）
func (r *AdminRepository) CountLostPetsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.LostPet{}).Where("status = ?", status).Count(&count).Error
	return count, err
}

// CountReportsByStatus 按状态统计举报（reportRepo 不可用时的兜底）
func (r *AdminRepository) CountReportsByStatus(status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
