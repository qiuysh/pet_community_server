package model

import (
	"time"
)

type Statistics struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	StatDate         time.Time `gorm:"uniqueIndex;type:date;not null" json:"statDate"`
	UserCount        int       `gorm:"default:0" json:"userCount"`
	NewUserCount     int       `gorm:"default:0" json:"newUserCount"`
	PostCount        int       `gorm:"default:0" json:"postCount"`
	NewPostCount     int       `gorm:"default:0" json:"newPostCount"`
	CommentCount     int       `gorm:"default:0" json:"commentCount"`
	NewCommentCount  int       `gorm:"default:0" json:"newCommentCount"`
	LikeCount        int       `gorm:"default:0" json:"likeCount"`
	NewLikeCount     int       `gorm:"default:0" json:"newLikeCount"`
	FavoriteCount    int       `gorm:"default:0" json:"favoriteCount"`
	NewFavoriteCount int       `gorm:"default:0" json:"newFavoriteCount"`
	ShareCount       int       `gorm:"default:0" json:"shareCount"`
	NewShareCount    int       `gorm:"default:0" json:"newShareCount"`
	TopicCount       int       `gorm:"default:0" json:"topicCount"`
	NewTopicCount    int       `gorm:"default:0" json:"newTopicCount"`
	ActiveUserCount  int       `gorm:"default:0" json:"activeUserCount"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Statistics) TableName() string {
	return "daily_stats"
}
