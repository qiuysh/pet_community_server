package model

import "time"

type AdPosition struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;size:50" json:"name"`
	Location  string    `gorm:"not null;size:50" json:"location"`
	Size      string    `gorm:"not null;size:20" json:"size"`
	Status    int       `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

type Advertisement struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	PositionID uint       `gorm:"not null" json:"positionId"`
	Title      string     `gorm:"not null;size:50" json:"title"`
	ImageURL   string     `gorm:"not null;size:255" json:"imageUrl"`
	LinkURL    string     `gorm:"size:255" json:"linkUrl"`
	StartTime  time.Time  `gorm:"not null" json:"startTime"`
	EndTime    time.Time  `gorm:"not null" json:"endTime"`
	ViewCount  int        `gorm:"not null;default:0" json:"viewCount"`
	ClickCount int        `gorm:"not null;default:0" json:"clickCount"`
	Status     int        `gorm:"not null;default:0" json:"status"`
	CreatedBy  uint       `gorm:"not null" json:"createdBy"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	Position   AdPosition `gorm:"foreignKey:PositionID" json:"position"`
}
