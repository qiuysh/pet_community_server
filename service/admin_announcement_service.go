// server/service/admin_announcement_service.go
package service

import (
	"errors"
	"strconv"
	"time"

	"pet-community/model"
	"pet-community/repository"
)

// AdminAnnouncementService 负责公告管理:CRUD + 状态 + 置顶。
type AdminAnnouncementService struct {
	adminRepo *repository.AdminRepository
}

// NewAdminAnnouncementService 构造 AdminAnnouncementService。
func NewAdminAnnouncementService(adminRepo *repository.AdminRepository) *AdminAnnouncementService {
	return &AdminAnnouncementService{adminRepo: adminRepo}
}

func (s *AdminAnnouncementService) CreateAnnouncement(title, content string, announcementType int, sortOrder int, startAt, endAt *time.Time, adminID uint, ip string) (*model.Announcement, error) {
	announcement := &model.Announcement{
		Title:     title,
		Content:   content,
		Type:      announcementType,
		Status:    1,
		SortOrder: sortOrder,
		StartAt:   startAt,
		EndAt:     endAt,
		AdminID:   adminID,
	}

	if err := s.adminRepo.CreateAnnouncement(announcement); err != nil {
		return nil, err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "create_announcement",
		TargetType: "announcement",
		TargetID:   announcement.ID,
		Detail:     "创建公告: " + title,
		IP:         ip,
	}
	s.adminRepo.CreateOperationLog(log)

	return announcement, nil
}

func (s *AdminAnnouncementService) UpdateAnnouncement(id uint, title, content string, announcementType int, sortOrder int, startAt, endAt *time.Time, adminID uint, ip string) (*model.Announcement, error) {
	announcement, err := s.adminRepo.GetAnnouncementByID(id)
	if err != nil {
		return nil, errors.New("公告不存在")
	}

	announcement.Title = title
	announcement.Content = content
	announcement.Type = announcementType
	announcement.SortOrder = sortOrder
	announcement.StartAt = startAt
	announcement.EndAt = endAt

	if err := s.adminRepo.UpdateAnnouncement(announcement); err != nil {
		return nil, err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "update_announcement",
		TargetType: "announcement",
		TargetID:   id,
		Detail:     "更新公告: " + title,
		IP:         ip,
	}
	s.adminRepo.CreateOperationLog(log)

	return announcement, nil
}

func (s *AdminAnnouncementService) DeleteAnnouncement(id uint, adminID uint, ip string) error {
	if err := s.adminRepo.DeleteAnnouncement(id); err != nil {
		return err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "delete_announcement",
		TargetType: "announcement",
		TargetID:   id,
		Detail:     "删除公告ID: " + strconv.FormatUint(uint64(id), 10),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminAnnouncementService) GetAnnouncementByID(id uint) (*model.Announcement, error) {
	return s.adminRepo.GetAnnouncementByID(id)
}

func (s *AdminAnnouncementService) ListAnnouncements(announcementType, status int, page, limit int) ([]model.Announcement, int64, error) {
	return s.adminRepo.ListAnnouncements(status, page, limit)
}

func (s *AdminAnnouncementService) UpdateAnnouncementStatus(id uint, status int, adminID uint, ip string) error {
	announcement, err := s.adminRepo.GetAnnouncementByID(id)
	if err != nil {
		return errors.New("公告不存在")
	}

	announcement.Status = status
	if err := s.adminRepo.UpdateAnnouncement(announcement); err != nil {
		return err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "update_announcement_status",
		TargetType: "announcement",
		TargetID:   id,
		Detail:     "更新公告状态: ID=" + strconv.FormatUint(uint64(id), 10) + ", status=" + strconv.Itoa(status),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminAnnouncementService) PinAnnouncement(id uint, adminID uint, ip string) error {
	maxSortOrder, _ := s.adminRepo.GetMaxAnnouncementSortOrder()

	announcement, err := s.adminRepo.GetAnnouncementByID(id)
	if err != nil {
		return errors.New("公告不存在")
	}

	announcement.SortOrder = maxSortOrder + 1
	if announcement.Status == 0 {
		announcement.Status = 1
	}

	if err := s.adminRepo.UpdateAnnouncement(announcement); err != nil {
		return err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "pin_announcement",
		TargetType: "announcement",
		TargetID:   id,
		Detail:     "置顶公告: " + announcement.Title,
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}
