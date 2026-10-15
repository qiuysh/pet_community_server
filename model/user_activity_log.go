package model

import "time"

type UserActivityLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"not null" json:"userId"`
	Action     string    `gorm:"not null;size:100" json:"action"`
	TargetType string    `gorm:"size:50" json:"targetType"`
	TargetID   uint      `json:"targetId"`
	IP         string    `gorm:"size:50" json:"ip"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
}

func (UserActivityLog) TableName() string {
	return "user_activity_logs"
}
