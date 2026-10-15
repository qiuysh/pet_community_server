// server/service/admin_front_user_service.go
package service

import (
	"errors"
	"strconv"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"
)

// AdminFrontUserService 负责前台用户管理:列表/详情/封禁/解禁/批量封禁。
type AdminFrontUserService struct {
	adminRepo *repository.AdminRepository
	userRepo  *repository.UserRepository
}

// NewAdminFrontUserService 构造 AdminFrontUserService。
func NewAdminFrontUserService(adminRepo *repository.AdminRepository, userRepo *repository.UserRepository) *AdminFrontUserService {
	return &AdminFrontUserService{adminRepo: adminRepo, userRepo: userRepo}
}

func (s *AdminFrontUserService) ListFrontUsers(keyword string, status int, page, pageSize int) ([]model.User, int64, error) {
	return s.userRepo.List(keyword, status, page, pageSize)
}

func (s *AdminFrontUserService) GetFrontUserDetail(userID uint) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *AdminFrontUserService) BanFrontUser(userID uint, reason string, adminID uint, ip string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.Status == 2 {
		return errors.New("用户已被封禁")
	}

	user.Status = 2
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	_ = s.userRepo.RemoveAllDevices(userID)
	_ = s.userRepo.IncrementTokenVersion(userID)
	cache.InvalidateUserSession(userID)

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "ban_front_user",
		TargetType: "user",
		TargetID:   userID,
		Detail:     "封禁用户ID: " + strconv.FormatUint(uint64(userID), 10) + ", 原因: " + reason,
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminFrontUserService) UnbanFrontUser(userID uint, adminID uint, ip string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.Status != 2 {
		return errors.New("用户未被封禁")
	}

	user.Status = 1
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "unban_front_user",
		TargetType: "user",
		TargetID:   userID,
		Detail:     "解封用户ID: " + strconv.FormatUint(uint64(userID), 10),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdminFrontUserService) BatchBanFrontUsers(userIDs []uint, reason string, adminID uint, ip string) error {
	if len(userIDs) == 0 {
		return errors.New("请选择要封禁的用户")
	}

	for _, userID := range userIDs {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			continue
		}
		if user.Status != 2 {
			user.Status = 2
			s.userRepo.Update(user)
			_ = s.userRepo.RemoveAllDevices(userID)
			_ = s.userRepo.IncrementTokenVersion(userID)
			cache.InvalidateUserSession(userID)
		}
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "batch_ban_front_users",
		TargetType: "user",
		TargetID:   userIDs[0],
		Detail:     "批量封禁用户, 数量: " + strconv.Itoa(len(userIDs)) + ", 原因: " + reason,
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}
