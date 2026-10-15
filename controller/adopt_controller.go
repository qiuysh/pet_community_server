package controller

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdoptController struct {
	adoptService *service.AdoptService
}

func NewAdoptController(adoptService *service.AdoptService) *AdoptController {
	return &AdoptController{
		adoptService: adoptService,
	}
}

// CreateAdopt 创建领养信息
func (c *AdoptController) CreateAdopt(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AdoptCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	adopt, err := c.adoptService.CreateAdopt(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, adopt)
}

// UpdateAdopt 更新领养信息
func (c *AdoptController) UpdateAdopt(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	adoptID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "领养ID无效")
		return
	}

	var req types.AdoptUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	adopt, err := c.adoptService.UpdateAdopt(userID, uint(adoptID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, adopt)
}

// DeleteAdopt 删除领养信息
func (c *AdoptController) DeleteAdopt(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	adoptID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "领养ID无效")
		return
	}

	if err := c.adoptService.DeleteAdopt(userID, uint(adoptID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// GetAdopt 获取领养信息详情
func (c *AdoptController) GetAdopt(ctx *gin.Context) {
	idStr := ctx.Param("id")
	adoptID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "领养ID无效")
		return
	}

	viewerID := ctx.GetUint("user_id")
	adopt, err := c.adoptService.GetAdoptByID(uint(adoptID), viewerID)
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, adopt)
}

// ListAdopts 领养信息列表
func (c *AdoptController) ListAdopts(ctx *gin.Context) {
	var req types.AdoptListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	viewerID := ctx.GetUint("user_id")
	result, err := c.adoptService.ListAdopts(&req, viewerID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ListMyAdopts 我的领养信息列表
func (c *AdoptController) ListMyAdopts(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.adoptService.ListMyAdopts(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// CreateApplication 创建领养申请
func (c *AdoptController) CreateApplication(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AdoptionApplicationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	application, err := c.adoptService.CreateApplication(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, application)
}

// CancelApplication 取消领养申请
func (c *AdoptController) CancelApplication(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	applicationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "申请ID无效")
		return
	}

	if err := c.adoptService.CancelApplication(userID, uint(applicationID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// ReviewApplication 审核领养申请
func (c *AdoptController) ReviewApplication(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AdoptionReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.adoptService.ReviewApplication(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// SignAgreement 签署领养协议
func (c *AdoptController) SignAgreement(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AdoptionAgreementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.adoptService.SignAgreement(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// ListApplicationsByAdoption 领养申请列表（按领养信息）
func (c *AdoptController) ListApplicationsByAdoption(ctx *gin.Context) {
	idStr := ctx.Param("id")
	adoptionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "领养ID无效")
		return
	}

	statusStr := ctx.Query("status")
	status := -1
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.adoptService.ListApplicationsByAdoption(uint(adoptionID), status, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ListMyApplications 我的领养申请列表
func (c *AdoptController) ListMyApplications(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	statusStr := ctx.Query("status")
	status := -1
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.adoptService.ListMyApplications(userID, status, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// UpdateAdoptStatus 更新领养状态
func (c *AdoptController) UpdateAdoptStatus(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AdoptionStatusUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.adoptService.UpdateAdoptStatus(userID, req.AdoptionID, req.Status); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// GetAdoptionStats 获取领养统计信息
func (c *AdoptController) GetAdoptionStats(ctx *gin.Context) {
	stats, err := c.adoptService.GetAdoptionStats()
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, stats)
}

// CreateFollowUp 创建领养回访记录
func (c *AdoptController) CreateFollowUp(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AdoptionFollowUpCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	followUp, err := c.adoptService.CreateFollowUp(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, followUp)
}

// ListFollowUpsByAdoption 获取领养回访记录列表
func (c *AdoptController) ListFollowUpsByAdoption(ctx *gin.Context) {
	idStr := ctx.Param("id")
	adoptionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "领养ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.adoptService.ListFollowUpsByAdoption(uint(adoptionID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// UploadAdoptPhoto 上传领养照片
func (c *AdoptController) UploadAdoptPhoto(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "未收到上传文件："+err.Error())
		return
	}

	// 动态/相册级上限 5MB（与上传矩阵一致）
	if err := utils.ValidateImageSize(fileHeader.Size, "adopt"); err != nil {
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
	if idx := lastDot(fileHeader.Filename); idx >= 0 {
		ext = fileHeader.Filename[idx+1:]
	}
	if ext == "" {
		utils.RespondError(ctx, types.CodeParamError, "无法识别图片扩展名")
		return
	}

	url, err := c.adoptService.UploadAdoptPhoto(f, ext)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, &types.UploadPhotoResponse{Url: url})
}

// lastDot 返回字符串最后一个 '.' 的索引（用于文件名扩展名分离）
func lastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}
