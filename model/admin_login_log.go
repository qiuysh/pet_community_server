package model

import (
	"time"
)

type AdminLoginLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AdminID    uint      `gorm:"index;not null" json:"adminId"`
	Username   string    `gorm:"size:50;not null" json:"username"`
	IP         string    `gorm:"size:50" json:"ip"`
	UserAgent  string    `gorm:"size:500" json:"userAgent"`
	Device     string    `gorm:"size:100" json:"device"`
	Status     int       `gorm:"not null;default:1" json:"status"`
	FailReason string    `gorm:"size:255" json:"failReason,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (AdminLoginLog) TableName() string {
	return "admin_login_logs"
}
