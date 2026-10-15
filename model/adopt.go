package model

import (
	"time"
)

// Adopt 领养信息模型
type Adopt struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	PetID           uint      `gorm:"default:0" json:"petId"`
	UserID          uint      `gorm:"not null;index" json:"userId"`
	PetName         string    `gorm:"size:50;not null" json:"petName"`
	PetType         string    `gorm:"size:20;not null;index" json:"petType"`
	PetAge          string    `gorm:"size:20" json:"petAge"`
	PetGender       int       `gorm:"default:0" json:"petGender"`
	HealthStatus    string    `gorm:"size:100" json:"healthStatus"`
	Photos          string    `gorm:"type:text" json:"photos"`
	Description     string    `gorm:"type:text" json:"description"`
	AdoptConditions string    `gorm:"type:text" json:"adoptConditions"`
	Location        string    `gorm:"size:100;index" json:"location"`
	ContactPhone    string    `gorm:"size:20" json:"contactPhone"`
	Status          int       `gorm:"default:0;index" json:"status"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Adopt) TableName() string {
	return "adoptions"
}

// AdoptionApplication 领养申请模型
type AdoptionApplication struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	AdoptionID      uint       `gorm:"not null;index" json:"adoptionId"`
	UserID          uint       `gorm:"not null;index" json:"userId"`
	ApplicationInfo string     `gorm:"type:text" json:"applicationInfo"`
	Status          int        `gorm:"default:0;index" json:"status"`
	ReviewedBy      uint       `gorm:"default:0" json:"reviewedBy"`
	ReviewedAt      *time.Time `json:"reviewedAt"`
	AgreementSigned int        `gorm:"default:0" json:"agreementSigned"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (AdoptionApplication) TableName() string {
	return "adoption_applications"
}

// AdoptionFollowUp 领养回访记录模型
type AdoptionFollowUp struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	AdoptionID    uint      `gorm:"not null;index" json:"adoptionId"`
	ApplicationID uint      `gorm:"not null;index" json:"applicationId"`
	UserID        uint      `gorm:"not null;index" json:"userId"`
	FollowUpType  string    `gorm:"size:20;not null" json:"followUpType"` // visit/photo/online
	Content       string    `gorm:"type:text" json:"content"`
	Photos        string    `gorm:"type:text" json:"photos"`
	Status        int       `gorm:"default:0" json:"status"` // 0:正常 1:异常
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (AdoptionFollowUp) TableName() string {
	return "adoption_follow_ups"
}
