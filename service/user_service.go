package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"pet-community/cache"
	"pet-community/config"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo   *repository.UserRepository
	followRepo *repository.FollowRepository
	postRepo   *repository.PostRepository
	petRepo    *repository.PetRepository
	blockRepo  *repository.BlockRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// SetProfileRepos 注入主页统计所需仓库
func (s *UserService) SetProfileRepos(followRepo *repository.FollowRepository, postRepo *repository.PostRepository, petRepo *repository.PetRepository) {
	s.followRepo = followRepo
	s.postRepo = postRepo
	s.petRepo = petRepo
}

func (s *UserService) SetBlockRepository(blockRepo *repository.BlockRepository) {
	s.blockRepo = blockRepo
}

func (s *UserService) Register(req *types.UserRegisterRequest, ip string) (*types.UserInfo, error) {
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New("手机号或邮箱必填")
	}

	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	isValid, err := s.VerifyCode(req.Phone+req.Email, "register", req.Code)
	if err != nil || !isValid {
		return nil, errors.New("验证码错误或已过期")
	}

	if req.Phone != "" {
		existing, _ := s.userRepo.GetByPhone(req.Phone)
		if existing != nil {
			return nil, errors.New("该手机号已注册")
		}
	}
	if req.Email != "" {
		existing, _ := s.userRepo.GetByEmail(req.Email)
		if existing != nil {
			return nil, errors.New("该邮箱已注册")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = generateNickname()
	}
	nickname = FilterSensitiveText(nickname)

	user := &model.User{
		Phone:           req.Phone,
		Email:           req.Email,
		Password:        string(hashedPassword),
		Nickname:        nickname,
		Status:          1,
		PrivacySettings: "{}",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return convertToUserInfo(user), nil
}

func (s *UserService) Login(req *types.UserLoginRequest, ip, userAgent string) (*types.UserLoginResponse, error) {
	user, err := s.userRepo.GetByAccount(req.Account)
	if err != nil {
		return nil, errors.New("账号或密码错误")
	}

	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	if err := ensureNotLoginLocked(cache.LoginScopeUser, user.ID, user); err != nil {
		return nil, err
	}

	if req.LoginType == "password" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			if ferr := recordUserLoginFailure(s, user); errors.Is(ferr, cache.ErrLoginLocked) {
				return nil, cache.ErrLoginLocked
			}
			return nil, errors.New("账号或密码错误")
		}
	} else if req.LoginType == "code" {
		isValid, err := s.VerifyCode(req.Account, "login", req.Code)
		if err != nil || !isValid {
			if ferr := recordUserLoginFailure(s, user); errors.Is(ferr, cache.ErrLoginLocked) {
				return nil, cache.ErrLoginLocked
			}
			return nil, errors.New("验证码错误或已过期")
		}
	} else {
		return nil, errors.New("不支持的登录方式")
	}

	clearUserLoginFailure(s, user)

	s.userRepo.UpdateLoginInfo(user.ID, ip)

	deviceID := generateDeviceID(userAgent, ip)
	device := truncateString(userAgent, 100)
	s.userRepo.AddDevice(&model.UserDevice{
		UserID:       user.ID,
		DeviceID:     deviceID,
		DeviceName:   device,
		DeviceType:   "web",
		IP:           ip,
		LastActiveAt: time.Now(),
	})
	cache.SetSessionMeta(user.ID, user.Status, user.TokenVersion)
	cache.CacheDevice(user.ID, deviceID)

	token, refreshToken, err := GenerateUserToken(user.ID, user.TokenVersion, deviceID, req.Remember)
	if err != nil {
		return nil, err
	}

	return &types.UserLoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		User:         *convertToUserInfo(user),
	}, nil
}

func (s *UserService) GetProfile(userID uint) (*types.UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	return convertToUserInfo(user), nil
}

// GetPublicProfile 公开主页（U5）：不含手机号/邮箱；拉黑双向不可见（F7-03）
func (s *UserService) GetPublicProfile(userID uint, viewerID uint) (*types.PublicUserProfile, error) {
	if viewerID > 0 && viewerID != userID && s.blockRepo != nil {
		blocked, err := s.blockRepo.ExistsEither(viewerID, userID)
		if err == nil && blocked {
			return nil, errors.New("无法查看该用户主页")
		}
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if user.Status != 1 {
		return nil, errors.New("用户不存在")
	}

	var postsCount, followersCount, followingCount, petsCount int64
	if s.postRepo != nil {
		postsCount, _ = s.postRepo.CountByUser(userID)
	}
	if s.followRepo != nil {
		followersCount, _ = s.followRepo.GetFollowersCount(userID)
		followingCount, _ = s.followRepo.GetFollowingCount(userID)
	}
	if s.petRepo != nil {
		petsCount, _ = s.petRepo.CountByUser(userID)
	}

	lv := userLevelOf(user.ID)
	return &types.PublicUserProfile{
		ID:             user.ID,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Gender:         user.Gender,
		Signature:      user.Signature,
		Location:       user.Location,
		PetYears:       user.PetYears,
		PostsCount:     postsCount,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
		PetsCount:      petsCount,
		Level:          lv,
		IsVipName:      lv >= 10,
	}, nil
}

func (s *UserService) UpdateProfile(userID uint, req *types.UserProfileUpdateRequest) (*types.UserInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if req.Nickname != "" {
		user.Nickname = FilterSensitiveText(req.Nickname)
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Signature != "" {
		user.Signature = req.Signature
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Location != "" {
		user.Location = req.Location
	}
	if req.PetYears != "" {
		user.PetYears = req.PetYears
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return convertToUserInfo(user), nil
}

func (s *UserService) ChangePassword(userID uint, req *types.PasswordChangeRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, string(hashedPassword))
}

func (s *UserService) ResetPassword(req *types.PasswordResetRequest) error {
	user, err := s.userRepo.GetByAccount(req.Account)
	if err != nil {
		return errors.New("账号不存在")
	}

	isValid, err := s.VerifyCode(req.Account, "reset_password", req.Code)
	if err != nil || !isValid {
		return errors.New("验证码错误或已过期")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(user.ID, string(hashedPassword))
}

func (s *UserService) SendCode(account, codeType, ip string) (expiresIn int, debugCode string, err error) {
	if !isValidAccount(account) {
		return 0, "", errors.New("账号格式不正确")
	}

	code := generateCode()

	if config.RedisEnabled() {
		if err := cache.AllowSendCode(account, ip); err != nil {
			return 0, "", err
		}
		if err := cache.StoreCode(account, codeType, code, ip); err != nil {
			return 0, "", err
		}
		fmt.Printf("[SendCode] account=%s type=%s code=%s (redis)\n", account, codeType, code)
		return 300, code, nil
	}

	oneMinuteAgo := time.Now().Add(-1 * time.Minute)
	count := s.userRepo.CountRecentCodes(account, ip, oneMinuteAgo)
	if count >= 1 {
		return 0, "", errors.New("发送太频繁，请稍后再试")
	}

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	count = s.userRepo.CountRecentCodes(account, ip, oneHourAgo)
	if count >= 10 {
		return 0, "", errors.New("发送次数过多，请稍后再试")
	}

	verificationCode := &model.VerificationCode{
		Account:   account,
		Code:      code,
		Type:      codeType,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		IP:        ip,
	}
	if err := s.userRepo.CreateVerificationCode(verificationCode); err != nil {
		return 0, "", err
	}

	fmt.Printf("[SendCode] account=%s type=%s code=%s\n", account, codeType, code)
	return 300, code, nil
}

func (s *UserService) VerifyCode(account, codeType, code string) (bool, error) {
	if config.RedisEnabled() {
		if err := cache.VerifyAndConsume(account, codeType, code); err != nil {
			return false, err
		}
		return true, nil
	}

	latestCode, err := s.userRepo.GetLatestVerificationCode(account, codeType)
	if err != nil {
		return false, errors.New("请先获取验证码")
	}
	if latestCode.Used {
		return false, errors.New("验证码已使用")
	}
	if time.Now().After(latestCode.ExpiresAt) {
		return false, errors.New("验证码已过期")
	}
	if latestCode.Code != code {
		return false, errors.New("验证码错误")
	}
	s.userRepo.MarkCodeUsed(latestCode.ID)
	return true, nil
}

func (s *UserService) Logout(userID uint, deviceID string) error {
	if deviceID != "" {
		_ = s.userRepo.RemoveDevice(userID, deviceID)
		cache.RemoveCachedDevice(userID, deviceID)
		return nil
	}
	_ = s.userRepo.RemoveAllDevices(userID)
	err := s.userRepo.IncrementTokenVersion(userID)
	cache.InvalidateUserSession(userID)
	return err
}

func (s *UserService) RemoveDeviceByID(userID uint, deviceID string) error {
	if deviceID == "" {
		return errors.New("设备ID无效")
	}
	err := s.userRepo.RemoveDevice(userID, deviceID)
	cache.RemoveCachedDevice(userID, deviceID)
	return err
}

// InvalidateUserSessions 封禁/强制全端下线：递增 Token 版本并清设备
func (s *UserService) InvalidateUserSessions(userID uint) error {
	_ = s.userRepo.RemoveAllDevices(userID)
	err := s.userRepo.IncrementTokenVersion(userID)
	cache.InvalidateUserSession(userID)
	return err
}

func (s *UserService) GetDevices(userID uint) ([]types.DeviceInfo, error) {
	devices, err := s.userRepo.GetUserDevices(userID)
	if err != nil {
		return nil, err
	}

	result := make([]types.DeviceInfo, len(devices))
	for i, d := range devices {
		result[i] = types.DeviceInfo{
			ID:           d.ID,
			DeviceID:     d.DeviceID,
			DeviceName:   d.DeviceName,
			DeviceType:   d.DeviceType,
			IP:           d.IP,
			LastActiveAt: d.LastActiveAt.Format("2006-01-02 15:04:05"),
		}
	}
	return result, nil
}

func (s *UserService) AdminListUsers(req *types.UserListRequest) (*types.UserListResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	users, total, err := s.userRepo.List(req.Keyword, req.Status, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.UserInfo, len(users))
	for i, u := range users {
		list[i] = *convertToUserInfo(&u)
	}

	return &types.UserListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *UserService) AdminUpdateUserStatus(userID uint, status int) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	user.Status = status
	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	if status != 1 {
		_ = s.InvalidateUserSessions(userID)
	}
	return nil
}

// ThirdPartyLogin 第三方登录
func (s *UserService) ThirdPartyLogin(req *types.ThirdPartyLoginRequest, ip, userAgent string) (*types.ThirdPartyLoginResponse, error) {
	// 模拟通过code获取openID（实际应调用微信/微博/QQ接口）
	openID := fmt.Sprintf("%s_%s", req.Platform, req.Code)

	// 查找是否已有绑定记录
	thirdAccount, err := s.userRepo.GetThirdAccount(req.Platform, openID)
	if err == nil && thirdAccount != nil {
		// 已存在，直接登录
		user, err := s.userRepo.GetByID(thirdAccount.UserID)
		if err != nil {
			return nil, errors.New("用户不存在")
		}
		if user.Status != 1 {
			return nil, errors.New("账号已被禁用")
		}

		s.userRepo.UpdateLoginInfo(user.ID, ip)
		deviceID := generateDeviceID(userAgent, ip)
		s.userRepo.AddDevice(&model.UserDevice{
			UserID: user.ID, DeviceID: deviceID, DeviceName: truncateString(userAgent, 100),
			DeviceType: "web", IP: ip, LastActiveAt: time.Now(),
		})
		token, refreshToken, err := GenerateUserToken(user.ID, user.TokenVersion, deviceID, true)
		if err != nil {
			return nil, err
		}

		return &types.ThirdPartyLoginResponse{
			Token:        token,
			RefreshToken: refreshToken,
			DeviceID:     deviceID,
			User:         *convertToUserInfo(user),
			IsNewUser:    false,
		}, nil
	}

	// 新用户，创建账号
	nickname := generateNickname()
	user := &model.User{
		Nickname:        nickname,
		Status:          1,
		PrivacySettings: "{}",
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// 创建第三方账号绑定
	thirdAccount = &model.UserThirdAccount{
		UserID:   user.ID,
		Platform: req.Platform,
		OpenID:   openID,
		Nickname: nickname,
	}
	if err := s.userRepo.CreateThirdAccount(thirdAccount); err != nil {
		return nil, err
	}

	s.userRepo.UpdateLoginInfo(user.ID, ip)
	deviceID := generateDeviceID(userAgent, ip)
	s.userRepo.AddDevice(&model.UserDevice{
		UserID: user.ID, DeviceID: deviceID, DeviceName: truncateString(userAgent, 100),
		DeviceType: "web", IP: ip, LastActiveAt: time.Now(),
	})
	token, refreshToken, err := GenerateUserToken(user.ID, user.TokenVersion, deviceID, true)
	if err != nil {
		return nil, err
	}

	return &types.ThirdPartyLoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
		User:         *convertToUserInfo(user),
		IsNewUser:    true,
	}, nil
}

// RefreshToken 刷新Token
func (s *UserService) RefreshToken(req *types.RefreshTokenRequest) (*types.RefreshTokenResponse, error) {
	token, refreshToken, err := RefreshUserToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &types.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}

// GetPrivacySettings 获取隐私设置
func (s *UserService) GetPrivacySettings(userID uint) (*types.PrivacySettings, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	settings := &types.PrivacySettings{
		PostVisibility:    0,
		CommentPermission: 0,
		MessagePermission: 0,
	}

	if user.PrivacySettings != "" {
		var saved map[string]int
		if err := json.Unmarshal([]byte(user.PrivacySettings), &saved); err == nil {
			if v, ok := saved["postVisibility"]; ok {
				settings.PostVisibility = v
			}
			if v, ok := saved["commentPermission"]; ok {
				settings.CommentPermission = v
			}
			if v, ok := saved["messagePermission"]; ok {
				settings.MessagePermission = v
			}
		}
	}

	return settings, nil
}

// UpdatePrivacySettings 更新隐私设置
func (s *UserService) UpdatePrivacySettings(userID uint, req *types.PrivacySettingsUpdateRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	settings := map[string]int{
		"postVisibility":    req.PostVisibility,
		"commentPermission": req.CommentPermission,
		"messagePermission": req.MessagePermission,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	user.PrivacySettings = string(data)
	return s.userRepo.Update(user)
}

// UnbindPhone 解绑手机号
func (s *UserService) UnbindPhone(userID uint, req *types.UnbindPhoneRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.Phone == "" {
		return errors.New("未绑定手机号")
	}

	isValid, err := s.VerifyCode(user.Phone, "change_password", req.Code)
	if err != nil || !isValid {
		return errors.New("验证码错误或已过期")
	}

	user.Phone = ""
	return s.userRepo.Update(user)
}

// UnbindEmail 解绑邮箱
func (s *UserService) UnbindEmail(userID uint, req *types.UnbindEmailRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.Email == "" {
		return errors.New("未绑定邮箱")
	}

	isValid, err := s.VerifyCode(user.Email, "change_password", req.Code)
	if err != nil || !isValid {
		return errors.New("验证码错误或已过期")
	}

	user.Email = ""
	return s.userRepo.Update(user)
}

// BindPhone 绑定或更换手机号
// 流程:校验手机号格式 -> 校验验证码 -> 校验手机号是否已被他人占用 -> 更新
func (s *UserService) BindPhone(userID uint, req *types.BindPhoneRequest) error {
	if !isPhone(req.Phone) {
		return errors.New("手机号格式不正确")
	}

	// 校验验证码:bind_phone 类型,与 unbind 使用的 change_password 类型区分,
	// 避免绑/解绑流程误用同一验证码
	isValid, err := s.VerifyCode(req.Phone, "bind_phone", req.Code)
	if err != nil || !isValid {
		return errors.New("验证码错误或已过期")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 校验目标手机号未被他人占用
	if user.Phone != req.Phone {
		existing, err := s.userRepo.GetByPhone(req.Phone)
		if err == nil && existing != nil && existing.ID != userID {
			return errors.New("该手机号已被其他账号绑定")
		}
	}

	user.Phone = req.Phone
	return s.userRepo.Update(user)
}

// BindEmail 绑定或更换邮箱
func (s *UserService) BindEmail(userID uint, req *types.BindEmailRequest) error {
	if !isEmail(req.Email) {
		return errors.New("邮箱格式不正确")
	}

	isValid, err := s.VerifyCode(req.Email, "bind_email", req.Code)
	if err != nil || !isValid {
		return errors.New("验证码错误或已过期")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.Email != req.Email {
		existing, err := s.userRepo.GetByEmail(req.Email)
		if err == nil && existing != nil && existing.ID != userID {
			return errors.New("该邮箱已被其他账号绑定")
		}
	}

	user.Email = req.Email
	return s.userRepo.Update(user)
}

// GenerateUserToken 签发双 Token。remember=true → Refresh 7d；false → Refresh 2h。Access 恒 24h。
// tv=用户 TokenVersion；deviceID 用于设备级登出/强制下线。
func GenerateUserToken(userID uint, tokenVersion int, deviceID string, remember bool) (string, string, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"tv":      tokenVersion,
		"did":     deviceID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	refreshTTL := 2 * time.Hour
	if remember {
		refreshTTL = 7 * 24 * time.Hour
	}
	refreshClaims := jwt.MapClaims{
		"user_id":  userID,
		"type":     "refresh",
		"tv":       tokenVersion,
		"did":      deviceID,
		"remember": remember,
		"exp":      time.Now().Add(refreshTTL).Unix(),
		"iat":      time.Now().Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func claimInt(claims jwt.MapClaims, key string) int {
	switch v := claims[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// validateUserSession 校验用户状态、Token 版本、设备是否仍有效（优先 Redis，miss 回源 MySQL）
func validateUserSession(userID uint, tokenVersion int, deviceID string) error {
	if config.DB == nil {
		return nil
	}

	if status, tv, ok := cache.GetSessionMeta(userID); ok {
		if status != 1 {
			return errors.New("账号已被禁用")
		}
		if tv != tokenVersion {
			return errors.New("Token已失效")
		}
		if deviceID == "" {
			return nil
		}
		if cache.HasCachedDevice(userID, deviceID) {
			return nil
		}
		var count int64
		if err := config.DB.Model(&model.UserDevice{}).
			Where("user_id = ? AND device_id = ?", userID, deviceID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("设备已下线")
		}
		cache.CacheDevice(userID, deviceID)
		return nil
	}

	var user model.User
	if err := config.DB.Select("id", "status", "token_version").First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.Status != 1 {
		return errors.New("账号已被禁用")
	}
	if user.TokenVersion != tokenVersion {
		return errors.New("Token已失效")
	}
	cache.SetSessionMeta(user.ID, user.Status, user.TokenVersion)

	if deviceID != "" {
		var count int64
		if err := config.DB.Model(&model.UserDevice{}).
			Where("user_id = ? AND device_id = ?", userID, deviceID).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("设备已下线")
		}
		cache.CacheDevice(userID, deviceID)
	}
	return nil
}

func ensureNotLoginLocked(scope string, id uint, user *model.User) error {
	if locked, err := cache.IsLoginLocked(scope, id); err == nil && locked {
		return cache.ErrLoginLocked
	}
	if user != nil && user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return errors.New("登录失败次数过多，请30分钟后再试")
	}
	return nil
}

func recordUserLoginFailure(s *UserService, user *model.User) error {
	if err := cache.RecordLoginFailure(cache.LoginScopeUser, user.ID); err == nil || errors.Is(err, cache.ErrLoginLocked) {
		if errors.Is(err, cache.ErrLoginLocked) {
			until := time.Now().Add(30 * time.Minute)
			user.LockedUntil = &until
			user.FailedLoginCount = 0
			_ = s.userRepo.Update(user)
		}
		return err
	}
	user.FailedLoginCount++
	if user.FailedLoginCount >= 5 {
		until := time.Now().Add(30 * time.Minute)
		user.LockedUntil = &until
		user.FailedLoginCount = 0
	}
	return s.userRepo.Update(user)
}

func clearUserLoginFailure(s *UserService, user *model.User) {
	cache.ClearLoginFailures(cache.LoginScopeUser, user.ID)
	user.FailedLoginCount = 0
	user.LockedUntil = nil
	_ = s.userRepo.Update(user)
}

func RefreshUserToken(refreshTokenString string) (string, string, error) {
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Cfg.JWT.Secret), nil
	})
	if err != nil {
		return "", "", errors.New("refresh token无效或已过期")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid token")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return "", "", errors.New("token类型错误")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", "", errors.New("invalid token claims")
	}
	userID := uint(userIDFloat)
	tv := claimInt(claims, "tv")
	deviceID, _ := claims["did"].(string)
	if err := validateUserSession(userID, tv, deviceID); err != nil {
		return "", "", err
	}

	remember, _ := claims["remember"].(bool)
	return GenerateUserToken(userID, tv, deviceID, remember)
}

func ParseUserToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.Cfg.JWT.Secret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, _ := claims["type"].(string)
		if tokenType != "access" {
			return 0, errors.New("token类型错误")
		}
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("invalid token claims")
		}
		userID := uint(userIDFloat)
		tv := claimInt(claims, "tv")
		deviceID, _ := claims["did"].(string)
		if err := validateUserSession(userID, tv, deviceID); err != nil {
			return 0, err
		}
		return userID, nil
	}

	return 0, errors.New("invalid token")
}

func validatePassword(password string) error {
	if len(password) < 6 || len(password) > 50 {
		return errors.New("密码长度需在6-50位之间")
	}
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	if !hasLower || !hasUpper || !hasNumber {
		return errors.New("密码需包含大小写字母和数字")
	}
	return nil
}

func userLevelOf(userID uint) int {
	if config.DB == nil || userID == 0 {
		return 1
	}
	var level model.UserLevel
	if err := config.DB.Where("user_id = ?", userID).First(&level).Error; err != nil {
		return 1
	}
	if level.Level < 1 {
		return 1
	}
	return level.Level
}

func convertToUserInfo(user *model.User) *types.UserInfo {
	lv := userLevelOf(user.ID)
	return &types.UserInfo{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     maskPhone(user.Phone),
		Email:     maskEmail(user.Email),
		Gender:    user.Gender,
		Signature: user.Signature,
		Location:  user.Location,
		PetYears:  user.PetYears,
		Status:    user.Status,
		Level:     lv,
		IsVipName: lv >= 10,
	}
}

func generateCode() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < 6; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

func generateNickname() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("宠友%06d", rand.Intn(1000000))
}

func generateDeviceID(userAgent, ip string) string {
	return fmt.Sprintf("web_%s_%s", ip, userAgent[:min(len(userAgent), 20)])
}

func isValidAccount(account string) bool {
	if isPhone(account) {
		return true
	}
	if isEmail(account) {
		return true
	}
	return false
}

func isPhone(phone string) bool {
	match, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return match
}

func isEmail(email string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, email)
	return match
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	for i, c := range email {
		if c == '@' {
			if i <= 2 {
				return email[:1] + "***" + email[i:]
			}
			return email[:2] + "***" + email[i:]
		}
	}
	return email
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
