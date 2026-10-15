package model

import (
	"time"
)

type Post struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"index;not null" json:"userId"`
	Content        string     `gorm:"type:text" json:"content"`
	Images         string     `gorm:"type:text" json:"images,omitempty"`
	Video          string     `gorm:"size:255" json:"video,omitempty"`
	PetIDs         string     `gorm:"size:200" json:"petIds,omitempty"`
	TopicIDs       string     `gorm:"size:200" json:"topicIds,omitempty"`
	Location       string     `gorm:"size:100" json:"location,omitempty"`
	Visibility     int        `gorm:"default:0;index" json:"visibility"`
	LikesCount     int        `gorm:"default:0" json:"likesCount"`
	CommentsCount  int        `gorm:"default:0" json:"commentsCount"`
	SharesCount    int        `gorm:"default:0" json:"sharesCount"`
	FavoritesCount int        `gorm:"default:0" json:"favoritesCount"`
	IsPinned       int        `gorm:"default:0" json:"isPinned"`
	IsFeatured     int        `gorm:"default:0" json:"isFeatured"`
	PublishAt      *time.Time `json:"publishAt,omitempty"`
	Status         int        `gorm:"default:0;index" json:"status"`
	CreatedAt      time.Time  `gorm:"autoCreateTime;index" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Post) TableName() string {
	return "posts"
}

type Comment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PostID     uint      `gorm:"index;not null" json:"postId"`
	UserID     uint      `gorm:"index;not null" json:"userId"`
	ParentID   uint      `gorm:"index;default:0" json:"parentId"`
	ReplyToID  uint      `json:"replyToId,omitempty"`
	Content    string    `gorm:"size:500;not null" json:"content"`
	AtUsers    string    `gorm:"size:200" json:"atUsers,omitempty"`
	LikesCount int       `gorm:"default:0" json:"likesCount"`
	Status     int       `gorm:"default:0" json:"status"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Comment) TableName() string {
	return "comments"
}

type Like struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"userId"`
	TargetType string    `gorm:"size:20;not null" json:"targetType"`
	TargetID   uint      `gorm:"not null" json:"targetId"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Like) TableName() string {
	return "likes"
}

type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	PostID    uint      `gorm:"not null" json:"postId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Favorite) TableName() string {
	return "favorites"
}

type Draft struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"userId"`
	Content    string     `gorm:"type:text" json:"content"`
	Images     string     `gorm:"type:text" json:"images,omitempty"`
	Video      string     `gorm:"size:255" json:"video,omitempty"`
	PetIDs     string     `gorm:"size:200" json:"petIds,omitempty"`
	TopicIDs   string     `gorm:"size:200" json:"topicIds,omitempty"`
	Location   string     `gorm:"size:100" json:"location,omitempty"`
	Visibility int        `gorm:"default:0" json:"visibility"`
	PublishAt  *time.Time `json:"publishAt,omitempty"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Draft) TableName() string {
	return "drafts"
}

type Report struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	ReporterID   uint       `gorm:"index;not null" json:"reporterId"`
	TargetType   string     `gorm:"size:20;not null" json:"targetType"`
	TargetID     uint       `gorm:"not null" json:"targetId"`
	Reason       string     `gorm:"size:500" json:"reason"`
	ReasonType   string     `gorm:"size:50" json:"reasonType"`
	Status       int        `gorm:"default:0;index" json:"status"` // 0待处理 1已处理 2已忽略
	HandleResult string     `gorm:"size:200" json:"handleResult,omitempty"`
	HandledBy    uint       `gorm:"default:0" json:"handledBy,omitempty"`
	HandledAt    *time.Time `json:"handledAt,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
}

func (Report) TableName() string {
	return "reports"
}

type Topic struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;unique;not null" json:"name"`
	Description string    `gorm:"size:200" json:"description"`
	Cover       string    `gorm:"size:255" json:"cover"`
	CreatorID   uint      `gorm:"index;default:0" json:"creatorId"` // 创建者；0 表示系统/运营
	PostsCount  int       `gorm:"default:0" json:"postsCount"`
	Status      int       `gorm:"default:0;index" json:"status"` // 0待审核 1已通过（Create 须显式写 Status，避免零值被 default 覆盖）
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Topic) TableName() string {
	return "topics"
}

type ShareRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	PostID    uint      `gorm:"not null" json:"postId"`
	Platform  string    `gorm:"size:20" json:"platform"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (ShareRecord) TableName() string {
	return "share_records"
}
