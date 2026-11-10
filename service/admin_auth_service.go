// server/service/admin_auth_service.go
package service

import (
	"errors"
	"strings"
	"time"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"
	util "pet-community/utils"

	"gorm.io/gorm"
)

func normalizeAdminRole(role string) string {
	if role == "admin" {
		return "super_admin"
	}
	return role
}

// AdminAuthService 负责管理员认证 + 管理员账号管理。
type AdminAuthService struct {
	adminRepo *repository.AdminRepository
}

// NewAdminAuthService 构造 AdminAuthService。
func NewAdminAuthService(adminRepo *repository.AdminRepository) *AdminAuthService {
	return &AdminAuthService{adminRepo: adminRepo}
}

const maxLoginFailures = 5
const loginLockDuration = 30 * time.Minute

// adminFailCount 连续失败次数：优先 Redis，否则用库内 FailedLoginCount。
func (s *AdminAuthService) adminFailCount(admin *model.AdminUser) int {
	if n, err := cache.GetLoginFailureCount(cache.LoginScopeAdmin, admin.ID); err == nil {
		return n
	}
	return admin.FailedLoginCount
}

func (s *AdminAuthService) requireCaptcha(admin *model.AdminUser) bool {
	return s.adminFailCount(admin) >= adminCaptchaRequired
}

func (s *AdminAuthService) loginErr(msg string, requireCaptcha bool) error {
	return &AdminLoginError{Msg: msg, RequireCaptcha: requireCaptcha}
}

// recordPasswordFailure 记录一次失败并返回对外错误（含是否需要验证码）。
func (s *AdminAuthService) recordPasswordFailure(admin *model.AdminUser, username, ip, userAgent, reason string) error {
	if ferr := cache.RecordLoginFailure(cache.LoginScopeAdmin, admin.ID); errors.Is(ferr, cache.ErrLoginLocked) {
		until := time.Now().Add(loginLockDuration)
		admin.LockedUntil = &until
		admin.FailedLoginCount = 0
		if uerr := s.adminRepo.Update(admin); uerr != nil {
			util.LogError(uerr, "AdminAuthService.Login.persistLock")
		}
		s.recordLoginLog(admin.ID, username, ip, userAgent, 0, "账号已锁定")
		return s.loginErr(cache.ErrLoginLocked.Error(), true)
	} else if ferr == nil {
		// Redis 已计数
	} else {
		admin.FailedLoginCount++
		if admin.FailedLoginCount >= maxLoginFailures {
			until := time.Now().Add(loginLockDuration)
			admin.LockedUntil = &until
			admin.FailedLoginCount = 0
		}
		if uerr := s.adminRepo.Update(admin); uerr != nil {
			util.LogError(uerr, "AdminAuthService.Login.persistFailedLoginCount")
		}
	}
	s.recordLoginLog(admin.ID, username, ip, userAgent, 0, reason)
	return s.loginErr("用户名或密码错误", s.requireCaptcha(admin))
}

func (s *AdminAuthService) Login(username, password, ip, userAgent, captchaID, captchaCode string) (string, *model.AdminUser, error) {
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.recordLoginLog(0, username, ip, userAgent, 0, "用户不存在")
			return "", nil, s.loginErr("用户不存在", false)
		}
		util.LogError(err, "AdminService.Login")
		return "", nil, s.loginErr("系统错误", false)
	}

	if admin.Status != 1 {
		s.recordLoginLog(admin.ID, username, ip, userAgent, 0, "账号已被禁用")
		return "", nil, s.loginErr("该账号已被禁用", false)
	}

	if locked, err := cache.IsLoginLocked(cache.LoginScopeAdmin, admin.ID); err == nil && locked {
		s.recordLoginLog(admin.ID, username, ip, userAgent, 0, "账号已锁定")
		return "", nil, s.loginErr(cache.ErrLoginLocked.Error(), true)
	}
	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		s.recordLoginLog(admin.ID, username, ip, userAgent, 0, "账号已锁定")
		return "", nil, s.loginErr("登录失败次数过多，请30分钟后再试", true)
	}

	needCaptcha := s.requireCaptcha(admin)
	if needCaptcha {
		if strings.TrimSpace(captchaID) == "" || strings.TrimSpace(captchaCode) == "" {
			return "", nil, s.recordPasswordFailure(admin, username, ip, userAgent, "缺少验证码")
		}
		if err := VerifyAdminCaptcha(captchaID, captchaCode); err != nil {
			// 验证码失败计入失败次数（R9）
			return "", nil, s.recordPasswordFailure(admin, username, ip, userAgent, "验证码错误")
		}
	} else if strings.TrimSpace(captchaID) != "" || strings.TrimSpace(captchaCode) != "" {
		// 未达阈值时若主动带了验证码，错了同样计失败；对了则继续验密码
		if err := VerifyAdminCaptcha(captchaID, captchaCode); err != nil {
			return "", nil, s.recordPasswordFailure(admin, username, ip, userAgent, "验证码错误")
		}
	}

	decryptedPassword, err := util.DecryptRSA(password)
	if err != nil {
		decryptedPassword = password
	}

	if !util.CheckPasswordHash(decryptedPassword, admin.Password) {
		return "", nil, s.recordPasswordFailure(admin, username, ip, userAgent, "密码错误")
	}

	cache.ClearLoginFailures(cache.LoginScopeAdmin, admin.ID)
	admin.FailedLoginCount = 0
	admin.LockedUntil = nil
	// 登录成功清零失败计数；失败不阻断登录，原锁定状态仍会在下次登录时被校验
	if uerr := s.adminRepo.Update(admin); uerr != nil {
		util.LogError(uerr, "AdminAuthService.Login.resetLoginState")
	}

	token, err := util.GenerateToken(admin.ID, admin.Username, normalizeAdminRole(admin.Role))
	if err != nil {
		util.LogError(err, "AdminService.Login")
		s.recordLoginLog(admin.ID, username, ip, userAgent, 0, "生成token失败")
		return "", nil, s.loginErr("生成token失败", false)
	}

	s.recordLoginLog(admin.ID, username, ip, userAgent, 1, "")

	return token, admin, nil
}

func (s *AdminAuthService) recordLoginLog(adminID uint, username, ip, userAgent string, status int, failReason string) {
	device := parseDevice(userAgent)
	log := &model.AdminLoginLog{
		AdminID:    adminID,
		Username:   username,
		IP:         ip,
		UserAgent:  userAgent,
		Device:     device,
		Status:     status,
		FailReason: failReason,
	}
	s.adminRepo.CreateLoginLog(log)
}

func parseDevice(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	return "desktop"
}

func (s *AdminAuthService) GetLoginLogs(adminID uint, status int, startTime, endTime time.Time, page, limit int) ([]model.AdminLoginLog, int64, error) {
	return s.adminRepo.GetLoginLogs(adminID, status, startTime, endTime, page, limit)
}

func (s *AdminAuthService) HashPassword(password string) (string, error) {
	return util.HashPassword(password)
}

func (s *AdminAuthService) CreateAdmin(username, password, email, role string, status int) (*model.AdminUser, error) {
	// Check if username already exists
	existing, err := s.adminRepo.FindByUsername(username)
	if err == nil && existing != nil && existing.ID > 0 {
		return nil, errors.New("用户名已存在")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 只有非 record not found 的错误才记录日志
		util.LogError(err, "AdminService.CreateAdmin")
		return nil, errors.New("系统错误")
	}

	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	admin := &model.AdminUser{
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Role:     normalizeAdminRole(role),
		Status:   status,
	}

	err = s.adminRepo.Create(admin)
	return admin, err
}

func (s *AdminAuthService) GetAdminByID(id uint) (*model.AdminUser, error) {
	return s.adminRepo.FindByID(id)
}

func (s *AdminAuthService) ListAdmins(keyword string, page, pageSize int) ([]model.AdminUser, int64, error) {
	return s.adminRepo.List(keyword, page, pageSize)
}

func (s *AdminAuthService) UpdateAdmin(id uint, username, email, role string, status *int) (*model.AdminUser, error) {
	admin, err := s.adminRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if username != "" && username != admin.Username {
		existing, err := s.adminRepo.FindByUsername(username)
		if err == nil && existing != nil && existing.ID > 0 && existing.ID != id {
			return nil, errors.New("用户名已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			util.LogError(err, "AdminService.UpdateAdmin")
			return nil, errors.New("系统错误")
		}
		admin.Username = username
	}

	if email != "" {
		admin.Email = email
	}

	if role != "" {
		admin.Role = normalizeAdminRole(role)
	}

	if status != nil {
		admin.Status = *status
	}

	err = s.adminRepo.Update(admin)
	return admin, err
}

func (s *AdminAuthService) UpdateAdminPassword(id uint, password string) error {
	admin, err := s.adminRepo.FindByID(id)
	if err != nil {
		return errors.New("用户不存在")
	}

	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return err
	}

	admin.Password = hashedPassword
	return s.adminRepo.Update(admin)
}

func (s *AdminAuthService) DeleteAdmin(id uint) error {
	return s.adminRepo.Delete(id)
}

func (s *AdminAuthService) UpdateAdminStatus(id, currentAdminID uint, status int) error {
	// 检查是否是自己
	if id == currentAdminID {
		return errors.New("不能禁用自己的账号")
	}
	return s.adminRepo.UpdateStatus(id, status)
}
