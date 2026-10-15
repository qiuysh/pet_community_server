package model

import "time"

type AdminOperationLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID    uint      `gorm:"not null" json:"adminId"`
	Action     string    `gorm:"not null;size:100" json:"action"`
	TargetType string    `gorm:"size:50" json:"targetType"`
	TargetID   uint      `json:"targetId"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IP         string    `gorm:"size:50" json:"ip"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
}

func (AdminOperationLog) TableName() string {
	return "admin_operation_logs"
}
