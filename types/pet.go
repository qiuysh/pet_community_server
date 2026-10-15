package types

import "time"

type PetInfo struct {
	ID           uint       `json:"id"`
	UserID       uint       `json:"userId"`
	Name         string     `json:"name"`
	Avatar       string     `json:"avatar"`
	PetType      int        `json:"petType"`
	BreedID      uint       `json:"breedId"`
	BreedName    string     `json:"breedName"`
	Gender       int        `json:"gender"`
	Birthday     *time.Time `json:"birthday,omitempty"`
	Weight       float64    `json:"weight"`
	Personality  string     `json:"personality"`
	Bio          string     `json:"bio"`
	IsSterilized int        `json:"isSterilized"`
	IsVaccinated int        `json:"isVaccinated"`
	Status       int        `json:"status"`
	ViewCount    int        `json:"viewCount"`
	CreatedAt    string     `json:"createdAt"`
}

type PetCreateRequest struct {
	Name         string     `json:"name" binding:"required,min=1,max=20"`
	Avatar       string     `json:"avatar" binding:"omitempty,max=255"`
	PetType      int        `json:"petType" binding:"required,oneof=1 2 3 4 5"`
	BreedID      uint       `json:"breedId" binding:"omitempty"`
	BreedName    string     `json:"breedName" binding:"omitempty,max=50"`
	Gender       int        `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Birthday     *time.Time `json:"birthday" binding:"omitempty"`
	Weight       float64    `json:"weight" binding:"omitempty"`
	Personality  string     `json:"personality" binding:"omitempty,max=100"`
	Bio          string     `json:"bio" binding:"omitempty,max=500"`
	IsSterilized int        `json:"isSterilized" binding:"omitempty,oneof=0 1"`
	IsVaccinated int        `json:"isVaccinated" binding:"omitempty,oneof=0 1"`
}

type PetUpdateRequest struct {
	Name         string     `json:"name" binding:"omitempty,min=1,max=20"`
	Avatar       string     `json:"avatar" binding:"omitempty,max=255"`
	PetType      int        `json:"petType" binding:"omitempty,oneof=1 2 3 4 5"`
	BreedID      uint       `json:"breedId" binding:"omitempty"`
	BreedName    string     `json:"breedName" binding:"omitempty,max=50"`
	Gender       int        `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Birthday     *time.Time `json:"birthday" binding:"omitempty"`
	Weight       float64    `json:"weight" binding:"omitempty"`
	Personality  string     `json:"personality" binding:"omitempty,max=100"`
	Bio          string     `json:"bio" binding:"omitempty,max=500"`
	IsSterilized *int       `json:"isSterilized" binding:"omitempty,oneof=0 1"`
	IsVaccinated *int       `json:"isVaccinated" binding:"omitempty,oneof=0 1"`
	Status       int        `json:"status" binding:"omitempty,oneof=0 1"`
}

type PetListRequest struct {
	UserID   uint   `form:"userId"`
	PetType  int    `form:"petType"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type PetListResponse struct {
	List     []PetInfo `json:"list"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
}

type PetBreedInfo struct {
	ID          uint   `json:"id"`
	PetType     int    `json:"petType"`
	Name        string `json:"name"`
	EnglishName string `json:"englishName"`
	Avatar      string `json:"avatar"`
	SortOrder   int    `json:"sortOrder"`
}

type PetAlbumInfo struct {
	ID        uint   `json:"id"`
	PetID     uint   `json:"petId"`
	ImageURL  string `json:"imageUrl"`
	ThumbURL  string `json:"thumbUrl"`
	Title     string `json:"title"`
	MediaType string `json:"mediaType"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
}

type PetAlbumCreateRequest struct {
	ImageURL  string `json:"imageUrl" binding:"required,max=255"`
	ThumbURL  string `json:"thumbUrl" binding:"omitempty,max=255"`
	Title     string `json:"title" binding:"omitempty,max=100"`
	MediaType string `json:"mediaType" binding:"omitempty,oneof=image video"`
	SortOrder int    `json:"sortOrder" binding:"omitempty"`
}

// PetAlbumSortRequest 相册排序（ids 按展示顺序，首位最先展示）
type PetAlbumSortRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1"`
}

type PetGrowthRecordInfo struct {
	ID         uint      `json:"id"`
	PetID      uint      `json:"petId"`
	RecordDate time.Time `json:"recordDate"`
	Weight     float64   `json:"weight"`
	Height     float64   `json:"height"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Images     []string  `json:"images"`
	CreatedAt  string    `json:"createdAt"`
}

type PetGrowthRecordCreateRequest struct {
	RecordDate time.Time `json:"recordDate" binding:"required"`
	Weight     float64   `json:"weight" binding:"omitempty"`
	Height     float64   `json:"height" binding:"omitempty"`
	Title      string    `json:"title" binding:"required,max=100"`
	Content    string    `json:"content" binding:"omitempty"`
	Images     []string  `json:"images" binding:"omitempty"`
}

type PetGrowthRecordUpdateRequest struct {
	RecordDate time.Time `json:"recordDate" binding:"omitempty"`
	Weight     float64   `json:"weight" binding:"omitempty"`
	Height     float64   `json:"height" binding:"omitempty"`
	Title      string    `json:"title" binding:"omitempty,max=100"`
	Content    string    `json:"content" binding:"omitempty"`
	Images     []string  `json:"images" binding:"omitempty"`
}

type LostPetInfo struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"userId"`
	PetID        uint      `json:"petId"`
	PetName      string    `json:"petName"`
	PetType      int       `json:"petType"`
	BreedID      uint      `json:"breedId"`
	BreedName    string    `json:"breedName"`
	Gender       int       `json:"gender"`
	Age          string    `json:"age"`
	Description  string    `json:"description"`
	Images       []string  `json:"images"`
	LostLocation string    `json:"lostLocation"`
	LostTime     time.Time `json:"lostTime"`
	Contact      string    `json:"contact"`
	Status       int       `json:"status"`
	ViewCount    int       `json:"viewCount"`
	CreatedAt    string    `json:"createdAt"`
}

type LostPetCreateRequest struct {
	PetID        uint      `json:"petId" binding:"omitempty"`
	PetName      string    `json:"petName" binding:"required,max=50"`
	PetType      int       `json:"petType" binding:"required"`
	BreedID      uint      `json:"breedId" binding:"omitempty"`
	BreedName    string    `json:"breedName" binding:"omitempty,max=50"`
	Gender       int       `json:"gender" binding:"omitempty"`
	Age          string    `json:"age" binding:"omitempty,max=20"`
	Description  string    `json:"description" binding:"required"`
	Images       []string  `json:"images" binding:"omitempty"`
	LostLocation string    `json:"lostLocation" binding:"required,max=200"`
	LostTime     time.Time `json:"lostTime" binding:"required"`
	Contact      string    `json:"contact" binding:"required,max=100"`
}

type LostPetListRequest struct {
	PetType      int    `form:"petType"`
	LostLocation string `form:"lostLocation"`
	Status       int    `form:"status"`
	Page         int    `form:"page"`
	PageSize     int    `form:"pageSize"`
}

type LostPetListResponse struct {
	List     []LostPetInfo `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}
