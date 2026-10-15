package model

import (
	"time"
)

type AdminUser struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	Username         string     `gorm:"unique;not null;size:50" json:"username"`
	Password         string     `gorm:"not null;size:255" json:"password,omitempty"`
	Email            string     `gorm:"size:100" json:"email"`
	Role             string     `gorm:"not null;size:20;default:'admin'" json:"role"`
	Status           int        `gorm:"not null;default:1" json:"status"`
	FailedLoginCount int        `gorm:"not null;default:0" json:"-"`
	LockedUntil      *time.Time `json:"-"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}
