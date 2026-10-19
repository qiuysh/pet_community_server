package model

import (
	"time"
)

type Block struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BlockerID uint      `gorm:"index;not null" json:"blockerId"`
	BlockedID uint      `gorm:"index;not null" json:"blockedId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Block) TableName() string {
	return "blocks"
}

type Conversation struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	User1ID         uint       `gorm:"index;not null" json:"user1Id"`
	User2ID         uint       `gorm:"index;not null" json:"user2Id"`
	LastMessageID   uint       `json:"lastMessageId"`
	LastMessage     string     `gorm:"size:200" json:"lastMessage"`
	LastMessageTime *time.Time `json:"lastMessageTime"`
	UnreadCount     int        `gorm:"default:0" json:"unreadCount"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type Message struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ConversationID uint   `gorm:"index;not null" json:"conversationId"`
	FromID         uint   `gorm:"index;not null" json:"fromId"`
	ToID           uint   `gorm:"index;not null" json:"toId"`
	Content        string `gorm:"type:text" json:"content"`
	Type           string `gorm:"size:20;default:text" json:"type"`
	MediaURL       string `gorm:"size:255" json:"mediaUrl"`
	// Duration 语音时长（秒）；非语音消息为 0
	Duration   int       `gorm:"not null;default:0" json:"duration"`
	Status     int       `gorm:"default:0" json:"status"`
	IsRecalled int       `gorm:"default:0" json:"isRecalled"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Message) TableName() string {
	return "messages"
}

// 消息类型（与 messages.type 列一致）
const (
	MessageTypeText  = "text"
	MessageTypeImage = "image"
	MessageTypeVoice = "voice"
)

type UserLevel struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"uniqueIndex;not null" json:"userId"`
	Level       int       `gorm:"default:1" json:"level"`
	Points      int       `gorm:"default:0" json:"points"`
	TotalPoints int       `gorm:"default:0" json:"totalPoints"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserLevel) TableName() string {
	return "user_levels"
}

type PointsRecord struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"userId"`
	Points      int       `json:"points"`
	Type        string    `gorm:"size:20" json:"type"`
	TargetID    uint      `json:"targetId"`
	Description string    `gorm:"size:100" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (PointsRecord) TableName() string {
	return "points_records"
}

type PointsRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Type        string    `gorm:"size:20;unique" json:"type"`
	Points      int       `json:"points"`
	Description string    `gorm:"size:100" json:"description"`
	IsActive    int       `gorm:"default:1" json:"isActive"`
	DailyLimit  int       `gorm:"default:0" json:"dailyLimit"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PointsRule) TableName() string {
	return "points_rules"
}

type LevelBenefit struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Level        int       `gorm:"index" json:"level"`
	BenefitType  string    `gorm:"size:20" json:"benefitType"`
	BenefitValue string    `gorm:"size:100" json:"benefitValue"`
	Description  string    `gorm:"size:200" json:"description"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (LevelBenefit) TableName() string {
	return "level_benefits"
}

type NotificationSettings struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"uniqueIndex;not null" json:"userId"`
	LikeNotify    int       `gorm:"default:1" json:"likeNotify"`
	CommentNotify int       `gorm:"default:1" json:"commentNotify"`
	AtNotify      int       `gorm:"default:1" json:"atNotify"`
	FollowNotify  int       `gorm:"default:1" json:"followNotify"`
	SystemNotify  int       `gorm:"default:1" json:"systemNotify"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (NotificationSettings) TableName() string {
	return "notification_settings"
}
