package model

import (
	"time"
)

type SearchHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Keyword   string    `gorm:"size:100;not null" json:"keyword"`
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (SearchHistory) TableName() string {
	return "search_histories"
}

type UserInterest struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"userId"`
	Interest   string    `gorm:"size:50;not null" json:"interest"`
	Weight     float64   `gorm:"default:1" json:"weight"`
	LastActive time.Time `gorm:"autoCreateTime" json:"lastActive"`
}

func (UserInterest) TableName() string {
	return "user_interests"
}

type TopicCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;unique;not null" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	Icon        string    `gorm:"size:100" json:"icon"`
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
	Status      int       `gorm:"default:1" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (TopicCategory) TableName() string {
	return "topic_categories"
}

type TopicParticipant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TopicID   uint      `gorm:"index;not null" json:"topicId"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (TopicParticipant) TableName() string {
	return "topic_participants"
}

type HotKeyword struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Keyword     string    `gorm:"size:100;unique;not null" json:"keyword"`
	SearchCount int       `gorm:"default:0" json:"searchCount"`
	Rank        int       `gorm:"column:rank;default:0" json:"rank"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (HotKeyword) TableName() string {
	return "hot_keywords"
}
