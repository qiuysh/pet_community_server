package model

import (
	"time"
)

type User struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Phone           string     `gorm:"uniqueIndex:idx_users_phone;size:20" json:"phone"`
	Email           string     `gorm:"uniqueIndex:idx_users_email;size:100" json:"email"`
	Password        string     `gorm:"size:255" json:"-"`
	Nickname        string     `gorm:"size:50;not null" json:"nickname"`
	Avatar          string     `gorm:"size:255" json:"avatar"`
	Signature       string     `gorm:"size:100" json:"signature"`
	Gender          int        `gorm:"default:0" json:"gender"`
	Location        string     `gorm:"size:100" json:"location"`
	PetYears        string     `gorm:"size:20" json:"petYears"`
	PrivacySettings string     `gorm:"type:json" json:"privacySettings,omitempty"`
	Status          int        `gorm:"default:0" json:"status"`
	// TokenVersion 递增后使既有 Access/Refresh 全部失效（登出全端、封禁）
	TokenVersion     int        `gorm:"default:0;not null" json:"-"`
	FailedLoginCount int        `gorm:"not null;default:0" json:"-"`
	LockedUntil      *time.Time `json:"-"`
	LastLoginAt      *time.Time `json:"lastLoginAt,omitempty"`
	LastLoginIP      string     `gorm:"size:50" json:"lastLoginIp,omitempty"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}

type UserDevice struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	DeviceID     string    `gorm:"size:100;not null" json:"deviceId"`
	DeviceName   string    `gorm:"size:100" json:"deviceName"`
	DeviceType   string    `gorm:"size:20" json:"deviceType"`
	IP           string    `gorm:"size:50" json:"ip"`
	LastActiveAt time.Time `json:"lastActiveAt"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserDevice) TableName() string {
	return "user_devices"
}

type VerificationCode struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Account   string    `gorm:"size:100;not null;index" json:"account"`
	Code      string    `gorm:"size:10;not null" json:"-"`
	Type      string    `gorm:"size:20;not null" json:"type"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `gorm:"default:false" json:"used"`
	IP        string    `gorm:"size:50" json:"ip"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

type UserThirdAccount struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Platform  string    `gorm:"size:20;not null" json:"platform"`
	OpenID    string    `gorm:"size:100;not null" json:"openId"`
	UnionID   string    `gorm:"size:100" json:"unionId,omitempty"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserThirdAccount) TableName() string {
	return "user_third_accounts"
}
