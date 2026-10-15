package model

import (
	"time"
)

// Activity 活动模型
type Activity struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:100;not null" json:"title"`
	Description     string    `gorm:"size:500;not null" json:"description"`
	Type            string    `gorm:"size:20;not null;index" json:"type"`
	Cover           string    `gorm:"size:255" json:"cover"`
	Status          int       `gorm:"default:0;index" json:"status"`
	StartTime       time.Time `gorm:"not null;index" json:"startTime"`
	EndTime         time.Time `gorm:"not null;index" json:"endTime"`
	Rules           string    `gorm:"type:text" json:"rules"`
	Prizes          string    `gorm:"type:text" json:"prizes"`
	MaxParticipants int       `gorm:"default:0" json:"maxParticipants"`
	MaxSubmissions  int       `gorm:"default:1" json:"maxSubmissions"`
	VotesPerDay     int       `gorm:"default:10" json:"votesPerDay"`
	RequireApproval bool      `gorm:"default:false" json:"requireApproval"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Activity) TableName() string {
	return "activities"
}

// Registration 活动报名模型
type Registration struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ActivityID uint      `gorm:"not null;index" json:"activityId"`
	UserID     uint      `gorm:"not null;index" json:"userId"`
	Status     int       `gorm:"default:0;index" json:"status"`
	// Priority Lv8+ 报名优先排队（1=优先，0=普通）
	Priority  int       `gorm:"default:0;index" json:"priority"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Phone     string    `gorm:"size:20;not null" json:"phone"`
	Message   string    `gorm:"size:200" json:"message"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Registration) TableName() string {
	return "activity_registrations"
}

// Submission 活动作品模型
type Submission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ActivityID  uint      `gorm:"not null;index" json:"activityId"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	Title       string    `gorm:"size:50;not null" json:"title"`
	Description string    `gorm:"size:500" json:"description"`
	Type        string    `gorm:"size:20;not null" json:"type"`
	Content     string    `gorm:"type:text" json:"content"`
	Images      string    `gorm:"type:text" json:"images"`
	Video       string    `gorm:"size:255" json:"video"`
	VotesCount  int       `gorm:"default:0" json:"votesCount"`
	Rank        int       `gorm:"default:0" json:"rank"`
	IsAwarded   bool      `gorm:"default:false" json:"isAwarded"`
	AwardLevel  string    `gorm:"size:20" json:"awardLevel"`
	Status      int       `gorm:"default:1;index" json:"status"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Submission) TableName() string {
	return "activity_submissions"
}

// Vote 活动投票模型
type Vote struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ActivityID   uint      `gorm:"not null;index" json:"activityId"`
	SubmissionID uint      `gorm:"not null;index" json:"submissionId"`
	UserID       uint      `gorm:"not null;index" json:"userId"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Vote) TableName() string {
	return "activity_votes"
}

// SignIn 线下活动签到模型
type SignIn struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ActivityID     uint      `gorm:"not null;index" json:"activityId"`
	UserID         uint      `gorm:"not null;index" json:"userId"`
	SignInTime     time.Time `gorm:"not null" json:"signInTime"`
	SignInLocation string    `gorm:"size:100" json:"signInLocation"`
	QRCode         string    `gorm:"size:100" json:"qrCode"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (SignIn) TableName() string {
	return "offline_activity_sign_ins"
}

// ActivityAlbum 活动相册模型
type ActivityAlbum struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ActivityID  uint      `gorm:"not null;index" json:"activityId"`
	UserID      uint      `gorm:"not null;index" json:"userId"`
	PhotoURL    string    `gorm:"size:255;not null" json:"photoUrl"`
	Description string    `gorm:"size:200" json:"description"`
	Status      int       `gorm:"default:1;index" json:"status"`
	ReviewedBy  uint      `json:"reviewedBy"`
	ReviewedAt  time.Time `json:"reviewedAt"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ActivityAlbum) TableName() string {
	return "activity_albums"
}

// ActivityReview 活动回顾模型
type ActivityReview struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ActivityID  uint      `gorm:"not null;unique;index" json:"activityId"`
	Title       string    `gorm:"size:100;not null" json:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Photos      string    `gorm:"type:text" json:"photos"`
	PublishedAt time.Time `json:"publishedAt"`
	PublishedBy uint      `gorm:"not null" json:"publishedBy"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ActivityReview) TableName() string {
	return "activity_reviews"
}
