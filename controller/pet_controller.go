package controller

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type PetController struct {
	petService *service.PetService
}

func NewPetController(petService *service.PetService) *PetController {
	return &PetController{petService: petService}
}

func (c *PetController) CreatePet(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.PetCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	pet, err := c.petService.CreatePet(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, pet)
}

func (c *PetController) UpdatePet(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	var req types.PetUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	pet, err := c.petService.UpdatePet(userID, uint(petID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, pet)
}

func (c *PetController) GetPet(ctx *gin.Context) {
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	pet, err := c.petService.GetPetByID(uint(petID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, pet)
}

func (c *PetController) DeletePet(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	if err := c.petService.DeletePet(userID, uint(petID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PetController) GetMyPets(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	pets, err := c.petService.GetUserPets(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, pets)
}

func (c *PetController) ListPets(ctx *gin.Context) {
	var req types.PetListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.petService.ListPets(&req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PetController) GetBreeds(ctx *gin.Context) {
	petTypeStr := ctx.Query("petType")
	petType := 0
	if petTypeStr != "" {
		if pt, err := strconv.Atoi(petTypeStr); err == nil {
			petType = pt
		}
	}

	breeds, err := c.petService.GetBreeds(petType)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, breeds)
}

// UploadAlbumPhoto 上传宠物相册图片（multipart field=file）
func (c *PetController) UploadAlbumPhoto(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "未收到上传文件："+err.Error())
		return
	}

	if err := utils.ValidateImageSize(fileHeader.Size, "album"); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "无法读取上传文件")
		return
	}
	defer f.Close()

	ext := ""
	if idx := strings.LastIndex(fileHeader.Filename, "."); idx >= 0 {
		ext = fileHeader.Filename[idx+1:]
	}
	if ext == "" {
		utils.RespondError(ctx, types.CodeParamError, "无法识别图片扩展名")
		return
	}

	url, err := c.petService.UploadAlbumPhoto(f, ext)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, &types.UploadPhotoResponse{Url: url})
}

// UploadAlbumVideo 上传宠物相册视频（multipart field=file，MP4 ≤50MB）
func (c *PetController) UploadAlbumVideo(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "未收到上传文件："+err.Error())
		return
	}

	if err := utils.ValidateVideoSize(fileHeader.Size); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "无法读取上传文件")
		return
	}
	defer f.Close()

	ext := ""
	if idx := strings.LastIndex(fileHeader.Filename, "."); idx >= 0 {
		ext = fileHeader.Filename[idx+1:]
	}
	if ext == "" {
		utils.RespondError(ctx, types.CodeParamError, "无法识别视频扩展名")
		return
	}

	url, err := c.petService.UploadAlbumVideo(f, ext)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, &types.UploadPhotoResponse{Url: url})
}

// ReorderAlbum 调整相册排序
func (c *PetController) ReorderAlbum(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	var req types.PetAlbumSortRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.petService.ReorderAlbum(userID, uint(petID), req.IDs); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, nil)
}

func (c *PetController) AddAlbumPhoto(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	var req types.PetAlbumCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	photo, err := c.petService.AddAlbumPhoto(userID, uint(petID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, photo)
}

func (c *PetController) GetPetAlbum(ctx *gin.Context) {
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	photos, err := c.petService.GetPetAlbum(uint(petID))
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, photos)
}

func (c *PetController) DeleteAlbumPhoto(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}
	photoIDStr := ctx.Param("photoId")
	photoID, err := strconv.ParseUint(photoIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "照片ID无效")
		return
	}

	if err := c.petService.DeleteAlbumPhoto(userID, uint(petID), uint(photoID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PetController) CreateGrowthRecord(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	var req types.PetGrowthRecordCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	record, err := c.petService.CreateGrowthRecord(userID, uint(petID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, record)
}

func (c *PetController) UpdateGrowthRecord(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}
	recordIDStr := ctx.Param("recordId")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "记录ID无效")
		return
	}

	var req types.PetGrowthRecordUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	record, err := c.petService.UpdateGrowthRecord(userID, uint(petID), uint(recordID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, record)
}

func (c *PetController) DeleteGrowthRecord(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}
	recordIDStr := ctx.Param("recordId")
	recordID, err := strconv.ParseUint(recordIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "记录ID无效")
		return
	}

	if err := c.petService.DeleteGrowthRecord(userID, uint(petID), uint(recordID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PetController) GetGrowthRecords(ctx *gin.Context) {
	petIDStr := ctx.Param("id")
	petID, err := strconv.ParseUint(petIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "宠物ID无效")
		return
	}

	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	page := 1
	pageSize := 20
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	records, total, err := c.petService.GetGrowthRecords(uint(petID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, gin.H{
		"list":     records,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (c *PetController) CreateLostPet(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.LostPetCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	lostPet, err := c.petService.CreateLostPet(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, lostPet)
}

func (c *PetController) GetLostPet(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "ID无效")
		return
	}

	viewerID := ctx.GetUint("user_id")
	lostPet, err := c.petService.GetLostPetByID(uint(id), viewerID)
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, lostPet)
}

func (c *PetController) ListLostPets(ctx *gin.Context) {
	var req types.LostPetListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	viewerID := ctx.GetUint("user_id")
	result, err := c.petService.ListLostPets(&req, viewerID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PetController) UpdateLostPetStatus(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "ID无效")
		return
	}

	var body struct {
		Status int `json:"status" binding:"required,oneof=1 2"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.petService.UpdateLostPetStatus(userID, uint(id), body.Status); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}
