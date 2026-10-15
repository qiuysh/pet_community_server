package model

import (
	"time"
)

// UserNotification 用户通知模型
type UserNotification struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"userId"`   // 接收通知的用户
	Type          string    `gorm:"size:20;not null" json:"type"`   // like/comment/at/system
	Title         string    `gorm:"size:200" json:"title"`          // 通知标题
	Content       string    `gorm:"size:500" json:"content"`        // 通知内容
	TriggerType   string    `gorm:"size:20" json:"triggerType"`     // post/comment
	TriggerID     uint      `gorm:"default:0" json:"triggerId"`     // 触发对象ID
	TriggerUserID uint      `gorm:"default:0" json:"triggerUserId"` // 触发用户ID
	IsRead        int       `gorm:"default:0" json:"isRead"`        // 是否已读
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserNotification) TableName() string {
	return "user_notifications"
}

// Follow 粉丝关系模型
type Follow struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FollowerID  uint      `gorm:"index;not null" json:"followerId"`  // 粉丝ID（关注者）
	FollowingID uint      `gorm:"index;not null" json:"followingId"` // 被关注者ID
	Status      int       `gorm:"default:1" json:"status"`           // 1:正常 0:已取消
	IsSecret    int       `gorm:"default:0" json:"isSecret"`         // 0:公开关注 1:悄悄关注
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Follow) TableName() string {
	return "follows"
}

// SystemNotification 系统通知（管理员发布）
type SystemNotification struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Type       int        `gorm:"not null;default:1" json:"type"`
	Title      string     `gorm:"size:200;not null" json:"title"`
	Content    string     `gorm:"type:text" json:"content"`
	TargetType int        `gorm:"not null;default:0" json:"targetType"`
	TargetIDs  string     `gorm:"type:text" json:"targetIds,omitempty"`
	Status     int        `gorm:"not null;default:1" json:"status"`
	PublishAt  *time.Time `json:"publishAt,omitempty"`
	AdminID    uint       `gorm:"not null" json:"adminId"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (SystemNotification) TableName() string {
	return "system_notifications"
}

type Announcement struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Title     string     `gorm:"size:200;not null" json:"title"`
	Content   string     `gorm:"type:text" json:"content"`
	Type      int        `gorm:"not null;default:1" json:"type"`
	Status    int        `gorm:"not null;default:1" json:"status"`
	SortOrder int        `gorm:"not null;default:0" json:"sortOrder"`
	AdminID   uint       `gorm:"not null" json:"adminId"`
	StartAt   *time.Time `json:"startAt,omitempty"`
	EndAt     *time.Time `json:"endAt,omitempty"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Announcement) TableName() string {
	return "announcements"
}
