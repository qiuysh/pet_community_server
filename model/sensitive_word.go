package model

import "time"

type SensitiveWord struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Word        string    `gorm:"column:word;unique;not null;size:100" json:"word"`
	ReplaceWith string    `gorm:"column:replacement;size:100" json:"replaceWith"`
	Category    string    `gorm:"column:category;size:50" json:"category"`
	Status      uint8     `gorm:"column:status;default:1" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at;not null" json:"createdAt"`
}

func (SensitiveWord) TableName() string {
	return "sensitive_words"
}
