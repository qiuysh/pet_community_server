package model

import (
	"time"
)

type Pet struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index;not null" json:"userId"`
	Name         string     `gorm:"size:50;not null" json:"name"`
	Avatar       string     `gorm:"size:255" json:"avatar"`
	PetType      int        `gorm:"not null" json:"petType"`
	BreedID      uint       `json:"breedId"`
	BreedName    string     `gorm:"size:50" json:"breedName"`
	Gender       int        `gorm:"default:0" json:"gender"`
	Birthday     *time.Time `json:"birthday,omitempty"`
	Weight       float64    `gorm:"type:decimal(5,2)" json:"weight"`
	Personality  string     `gorm:"size:100" json:"personality"`
	Bio          string     `gorm:"size:500" json:"bio"`
	IsSterilized int        `gorm:"default:0" json:"isSterilized"`
	IsVaccinated int        `gorm:"default:0" json:"isVaccinated"`
	Status       int        `gorm:"default:1" json:"status"`
	ViewCount    int        `gorm:"default:0" json:"viewCount"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Pet) TableName() string {
	return "pets"
}

type PetBreed struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	PetType     int    `gorm:"not null;index" json:"petType"`
	Name        string `gorm:"size:50;not null" json:"name"`
	EnglishName string `gorm:"size:100" json:"englishName"`
	Avatar      string `gorm:"size:255" json:"avatar"`
	SortOrder   int    `gorm:"default:0" json:"sortOrder"`
	Status      int    `gorm:"default:1" json:"status"`
}

func (PetBreed) TableName() string {
	return "pet_breeds"
}

type PetAlbum struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PetID     uint      `gorm:"index;not null" json:"petId"`
	ImageURL  string    `gorm:"size:255;not null" json:"imageUrl"` // 图片或视频 URL
	ThumbURL  string    `gorm:"size:255" json:"thumbUrl"`
	Title     string    `gorm:"size:100" json:"title"`
	MediaType string    `gorm:"size:20;default:image" json:"mediaType"` // image | video
	SortOrder int       `gorm:"default:0" json:"sortOrder"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (PetAlbum) TableName() string {
	return "pet_albums"
}

type PetGrowthRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PetID      uint      `gorm:"index;not null" json:"petId"`
	RecordDate time.Time `json:"recordDate"`
	Weight     float64   `gorm:"type:decimal(5,2)" json:"weight"`
	Height     float64   `gorm:"type:decimal(5,2)" json:"height"`
	Title      string    `gorm:"size:100" json:"title"`
	Content    string    `gorm:"type:text" json:"content"`
	Images     string    `gorm:"type:text" json:"images,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PetGrowthRecord) TableName() string {
	return "pet_growth_records"
}

type LostPet struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	PetID        uint      `gorm:"index" json:"petId"`
	PetName      string    `gorm:"size:50;not null" json:"petName"`
	PetType      int       `gorm:"not null" json:"petType"`
	BreedID      uint      `json:"breedId"`
	BreedName    string    `gorm:"size:50" json:"breedName"`
	Gender       int       `json:"gender"`
	Age          string    `gorm:"size:20" json:"age"`
	Description  string    `gorm:"type:text" json:"description"`
	Images       string    `gorm:"type:text" json:"images,omitempty"`
	LostLocation string    `gorm:"size:200" json:"lostLocation"`
	LostTime     time.Time `json:"lostTime"`
	Contact      string    `gorm:"size:100;not null" json:"contact"`
	Status       int       `gorm:"default:1" json:"status"`
	ViewCount    int       `gorm:"default:0" json:"viewCount"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (LostPet) TableName() string {
	return "lost_pets"
}
