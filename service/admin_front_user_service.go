// server/service/admin_front_user_service.go
package service

import (
	"errors"
	"fmt"
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
	if err := s.invalidateSessions(userID); err != nil {
		return fmt.Errorf("用户已封禁，但强制下线失败: %w", err)
	}
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

	var firstErr error
	banned := 0
	for _, userID := range userIDs {
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			continue
		}
		if user.Status == 2 {
			continue
		}
		user.Status = 2
		if err := s.userRepo.Update(user); err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("用户 %d 封禁失败: %w", userID, err)
			}
			continue
		}
		banned++
		if err := s.invalidateSessions(userID); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("用户 %d 已封禁但强制下线失败: %w", userID, err)
		}
		cache.InvalidateUserSession(userID)
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     "batch_ban_front_users",
		TargetType: "user",
		TargetID:   userIDs[0],
		Detail:     "批量封禁用户, 成功: " + strconv.Itoa(banned) + ", 请求: " + strconv.Itoa(len(userIDs)) + ", 原因: " + reason,
		IP:         ip,
	}
	if err := s.adminRepo.CreateOperationLog(log); err != nil {
		return err
	}
	return firstErr
}

// invalidateSessions 使目标用户所有登录态失效：清设备 + 递增 Token 版本。
// 两步都不能静默失败，否则旧 Token 在缓存/设备失效前仍可能通过校验。
func (s *AdminFrontUserService) invalidateSessions(userID uint) error {
	if err := s.userRepo.RemoveAllDevices(userID); err != nil {
		return err
	}
	return s.userRepo.IncrementTokenVersion(userID)
}
