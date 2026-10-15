// server/service/admin_notification_service.go
package service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"pet-community/model"
	"pet-community/repository"
	util "pet-community/utils"
)

// AdminNotificationService 负责系统通知管理:CRUD + 发送 + 定时处理。
type AdminNotificationService struct {
	adminRepo *repository.AdminRepository
}

// NewAdminNotificationService 构造 AdminNotificationService。
func NewAdminNotificationService(adminRepo *repository.AdminRepository) *AdminNotificationService {
	return &AdminNotificationService{adminRepo: adminRepo}
}

func (s *AdminNotificationService) CreateNotification(title, content string, notificationType, targetType int, targetIDs []uint, publishAt *time.Time, adminID uint, ip string) (*model.SystemNotification, error) {
	targetIDsStr := ""
	if len(targetIDs) > 0 {
		ids := make([]string, 0, len(targetIDs))
		for _, id := range targetIDs {
			ids = append(ids, strconv.FormatUint(uint64(id), 10))
		}
		targetIDsStr = strings.Join(ids, ",")
	}

	notification := &model.SystemNotification{
		Type:       notificationType,
		Title:      title,
		Content:    content,
		TargetType: targetType,
		TargetIDs:  targetIDsStr,
		Status:     0,
		PublishAt:  publishAt,
		AdminID:    adminID,
	}

	if err := s.adminRepo.CreateNotification(notification); err != nil {
		return nil, err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "create_notification",
		TargetType: "notification",
		TargetID:   notification.ID,
		Detail:     "创建通知: " + title,
		IP:         ip,
	}
	s.adminRepo.CreateOperationLog(log)

	return notification, nil
}

func (s *AdminNotificationService) UpdateNotification(id uint, title, content string, notificationType, targetType int, targetIDs []uint, publishAt *time.Time, adminID uint, ip string) (*model.SystemNotification, error) {
	notification, err := s.adminRepo.GetNotificationByID(id)
	if err != nil {
		return nil, errors.New("通知不存在")
	}

	targetIDsStr := ""
	if len(targetIDs) > 0 {
		ids := make([]string, 0, len(targetIDs))
		for _, tid := range targetIDs {
			ids = append(ids, strconv.FormatUint(uint64(tid), 10))
		}
		targetIDsStr = strings.Join(ids, ",")
	}

	notification.Title = title
	notification.Content = content
	notification.Type = notificationType
	notification.TargetType = targetType
	notification.TargetIDs = targetIDsStr
	notification.PublishAt = publishAt

	if err := s.adminRepo.UpdateNotification(notification); err != nil {
		return nil, err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "update_notification",
		TargetType: "notification",
		TargetID:   id,
		Detail:     "更新通知: " + title,
		IP:         ip,
	}
	s.adminRepo.CreateOperationLog(log)

	return notification, nil
}

func (s *AdminNotificationService) DeleteNotification(id uint, adminID uint, ip string) error {
	if err := s.adminRepo.DeleteNotification(id); err != nil {
		return err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "delete_notification",
		TargetType: "notification",
		TargetID:   id,
		Detail:     "删除通知ID: " + strconv.FormatUint(uint64(id), 10),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminNotificationService) GetNotificationByID(id uint) (*model.SystemNotification, error) {
	return s.adminRepo.GetNotificationByID(id)
}

func (s *AdminNotificationService) ListNotifications(notificationType, status int, page, limit int) ([]model.SystemNotification, int64, error) {
	return s.adminRepo.ListNotificationsByFilter(notificationType, status, page, limit)
}

func (s *AdminNotificationService) SendNotification(id uint, adminID uint, ip string) error {
	notification, err := s.adminRepo.GetNotificationByID(id)
	if err != nil {
		return errors.New("通知不存在")
	}

	if notification.Status == 1 {
		return errors.New("通知已发送")
	}

	var userIDs []uint
	switch notification.TargetType {
	case 0:
		users, _ := s.adminRepo.ListActiveUsers()
		for _, u := range users {
			userIDs = append(userIDs, u.ID)
		}
	case 1:
		if notification.TargetIDs != "" {
			userIDs = ParseAtUsersFromString(notification.TargetIDs)
		}
	case 2:
		if notification.TargetIDs != "" {
			userIDs = ParseAtUsersFromString(notification.TargetIDs)
		}
	}

	for _, uid := range userIDs {
		userNotification := &model.UserNotification{
			UserID:        uid,
			Type:          "system",
			Title:         notification.Title,
			Content:       notification.Content,
			TriggerType:   "system",
			TriggerID:     notification.ID,
			TriggerUserID: 0,
			IsRead:        0,
		}
		s.adminRepo.CreateUserNotification(userNotification)
	}

	notification.Status = 1
	if err := s.adminRepo.UpdateNotification(notification); err != nil {
		return err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "send_notification",
		TargetType: "notification",
		TargetID:   id,
		Detail:     "发送通知: " + notification.Title + ", 目标用户数: " + strconv.Itoa(len(userIDs)),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminNotificationService) ProcessScheduledNotifications() error {
	notifications, err := s.adminRepo.ListScheduledNotifications()
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := s.SendNotification(notification.ID, notification.AdminID, "system"); err != nil {
			util.LogError(err, "ProcessScheduledNotifications")
		}
	}

	return nil
}
