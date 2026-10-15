package model

import "time"

type UserBehaviorLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `json:"userId"`
	ActionType string    `gorm:"not null;size:50" json:"actionType"`
	TargetType string    `gorm:"size:50" json:"targetType"`
	TargetID   uint      `json:"targetId"`
	Detail     string    `gorm:"type:text" json:"detail"`
	IP         string    `gorm:"size:50" json:"ip"`
	Device     string    `gorm:"size:100" json:"device"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
}

func (UserBehaviorLog) TableName() string {
	return "user_behavior_logs"
}
