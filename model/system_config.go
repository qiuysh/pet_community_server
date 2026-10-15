package model

import "time"

type SystemConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfigKey   string    `gorm:"unique;not null;size:100" json:"configKey"`
	ConfigValue string    `gorm:"type:text" json:"configValue"`
	Description string    `gorm:"size:200" json:"description"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
}

func (SystemConfig) TableName() string {
	return "system_config"
}
