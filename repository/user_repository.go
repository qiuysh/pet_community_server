package repository

import (
	"pet-community/model"
	"time"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByPhone(phone string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByNickname(nickname string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("nickname = ?", nickname).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByAccount(account string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("phone = ? OR email = ?", account, account).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(keyword string, status int, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	query := r.db.Model(&model.User{})

	if keyword != "" {
		query = query.Where("nickname LIKE ? OR phone LIKE ? OR email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *UserRepository) UpdateLoginInfo(id uint, ip string) error {
	now := time.Now()
	return r.db.Model(&model.User{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

func (r *UserRepository) UpdatePassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).
		Update("password", hashedPassword).Error
}

func (r *UserRepository) CreateVerificationCode(code *model.VerificationCode) error {
	return r.db.Create(code).Error
}

func (r *UserRepository) GetLatestVerificationCode(account, codeType string) (*model.VerificationCode, error) {
	var code model.VerificationCode
	if err := r.db.Where("account = ? AND type = ?", account, codeType).
		Order("created_at DESC").
		First(&code).Error; err != nil {
		return nil, err
	}
	return &code, nil
}

func (r *UserRepository) MarkCodeUsed(id uint) error {
	return r.db.Model(&model.VerificationCode{}).Where("id = ?", id).
		Update("used", true).Error
}

func (r *UserRepository) CountRecentCodes(account, ip string, since time.Time) int64 {
	var count int64
	r.db.Model(&model.VerificationCode{}).
		Where("(account = ? OR ip = ?) AND created_at >= ?", account, ip, since).
		Count(&count)
	return count
}

func (r *UserRepository) AddDevice(device *model.UserDevice) error {
	var existing model.UserDevice
	err := r.db.Where("user_id = ? AND device_id = ?", device.UserID, device.DeviceID).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Updates(map[string]interface{}{
			"last_active_at": time.Now(),
			"ip":             device.IP,
			"device_name":    device.DeviceName,
			"device_type":    device.DeviceType,
		}).Error
	}
	return r.db.Create(device).Error
}

func (r *UserRepository) GetUserDevices(userID uint) ([]model.UserDevice, error) {
	var devices []model.UserDevice
	if err := r.db.Where("user_id = ?", userID).Order("last_active_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

func (r *UserRepository) RemoveDevice(userID uint, deviceID string) error {
	return r.db.Where("user_id = ? AND device_id = ?", userID, deviceID).Delete(&model.UserDevice{}).Error
}

// DeviceExists 设备是否仍登记（用于 Token 设备级失效）
func (r *UserRepository) DeviceExists(userID uint, deviceID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.UserDevice{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Count(&count).Error
	return count > 0, err
}

// IncrementTokenVersion 使该用户全部 JWT 失效
func (r *UserRepository) IncrementTokenVersion(userID uint) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("token_version", gorm.Expr("token_version + 1")).Error
}

// RemoveAllDevices 清除用户全部登录设备
func (r *UserRepository) RemoveAllDevices(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.UserDevice{}).Error
}

func (r *UserRepository) GetThirdAccount(platform, openID string) (*model.UserThirdAccount, error) {
	var account model.UserThirdAccount
	if err := r.db.Where("platform = ? AND open_id = ?", platform, openID).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *UserRepository) CreateThirdAccount(account *model.UserThirdAccount) error {
	return r.db.Create(account).Error
}

func (r *UserRepository) GetUserThirdAccounts(userID uint) ([]model.UserThirdAccount, error) {
	var accounts []model.UserThirdAccount
	if err := r.db.Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *UserRepository) UnbindThirdAccount(userID uint, platform string) error {
	return r.db.Where("user_id = ? AND platform = ?", userID, platform).Delete(&model.UserThirdAccount{}).Error
}

func (r *UserRepository) SearchByNickname(keyword string, limit int) ([]model.User, error) {
	var users []model.User
	if err := r.db.Where("nickname LIKE ?", "%"+keyword+"%").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) CountByDate(date string) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("DATE(created_at) = ?", date).Count(&count).Error
	return count, err
}

func (r *UserRepository) CountByDateAndStatus(date string, status int) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("DATE(created_at) = ? AND status = ?", date, status).Count(&count).Error
	return count, err
}

type LocationCount struct {
	Location string
	Count    int64
}

func (r *UserRepository) GetLocationDistribution() ([]LocationCount, error) {
	var results []LocationCount
	err := r.db.Model(&model.User{}).Where("location != ''").
		Select("location, COUNT(*) as count").Group("location").
		Order("count DESC").Limit(10).Scan(&results).Error
	return results, err
}

type GenderCount struct {
	Gender int
	Count  int64
}

func (r *UserRepository) GetGenderDistribution() ([]GenderCount, error) {
	var results []GenderCount
	err := r.db.Model(&model.User{}).Where("status = 1").
		Select("gender, COUNT(*) as count").Group("gender").Scan(&results).Error
	return results, err
}

// FindActiveUsers 查询活跃用户（status = 1），按 limit 限制数量
func (r *UserRepository) FindActiveUsers(limit int) ([]model.User, error) {
	var users []model.User
	if err := r.db.Where("status = 1").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}


