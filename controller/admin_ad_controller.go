package controller

import (
	"pet-community/model"
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminAdController 广告位与广告管理。
// RBAC: 广告位管理 → super_admin only; 广告 CRUD + 审核 → super_admin | content_admin。
type AdminAdController struct {
	adService *service.AdService
}

func NewAdminAdController(adService *service.AdService) *AdminAdController {
	return &AdminAdController{adService: adService}
}

func adPositionToResponse(position *model.AdPosition) types.AdPositionResponse {
	return types.AdPositionResponse{
		ID:        position.ID,
		Name:      position.Name,
		Location:  position.Location,
		Size:      position.Size,
		Status:    position.Status,
		CreatedAt: position.CreatedAt.Format(timestampFmt),
		UpdatedAt: position.UpdatedAt.Format(timestampFmt),
	}
}

func adToResponse(ad *model.Advertisement) types.AdvertisementResponse {
	return types.AdvertisementResponse{
		ID:         ad.ID,
		PositionID: ad.PositionID,
		Position:   adPositionToResponse(&ad.Position),
		Title:      ad.Title,
		ImageURL:   ad.ImageURL,
		LinkURL:    ad.LinkURL,
		StartTime:  ad.StartTime.Format(timestampFmt),
		EndTime:    ad.EndTime.Format(timestampFmt),
		ViewCount:  ad.ViewCount,
		ClickCount: ad.ClickCount,
		Status:     ad.Status,
		CreatedBy:  ad.CreatedBy,
		CreatedAt:  ad.CreatedAt.Format(timestampFmt),
		UpdatedAt:  ad.UpdatedAt.Format(timestampFmt),
	}
}

// ========= 广告位 =========

type CreateAdPositionRequest struct {
	Name     string `json:"name" binding:"required,max=50"`
	Location string `json:"location" binding:"required,oneof=home_banner splash detail_sidebar topic_banner"`
	Size     string `json:"size" binding:"required,max=20"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type UpdateAdPositionRequest struct {
	Name     string `json:"name" binding:"omitempty,max=50"`
	Location string `json:"location" binding:"omitempty,oneof=home_banner splash detail_sidebar topic_banner"`
	Size     string `json:"size" binding:"omitempty,max=20"`
	// 指针：未传时不覆盖，避免前端省略 status 时被写成 0（禁用）
	Status *int `json:"status" binding:"omitempty,oneof=0 1"`
}

func (c *AdminAdController) CreateAdPosition(ctx *gin.Context) {
	var req CreateAdPositionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	position, err := c.adService.CreatePosition(req.Name, req.Location, req.Size, req.Status)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adPositionToResponse(position)))
}

func (c *AdminAdController) GetAdPosition(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告位ID")
	if !ok {
		return
	}
	position, err := c.adService.GetPositionByID(id)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeNotFound, "广告位不存在"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adPositionToResponse(position)))
}

func (c *AdminAdController) ListAdPositions(ctx *gin.Context) {
	page, pageSize := ParsePageQuery(ctx, 100)
	positions, total, err := c.adService.ListPositions(page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	list := make([]types.AdPositionResponse, len(positions))
	for i, p := range positions {
		list[i] = adPositionToResponse(&p)
	}
	ctx.JSON(200, types.PageResponse(list, total, page, pageSize))
}

func (c *AdminAdController) UpdateAdPosition(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告位ID")
	if !ok {
		return
	}
	var req UpdateAdPositionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	position, err := c.adService.UpdatePosition(id, req.Name, req.Location, req.Size, req.Status)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adPositionToResponse(position)))
}

func (c *AdminAdController) DeleteAdPosition(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告位ID")
	if !ok {
		return
	}
	if err := c.adService.DeletePosition(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

// ========= 广告 =========

type CreateAdRequest struct {
	PositionID uint   `json:"positionId" binding:"required"`
	Title      string `json:"title" binding:"required,max=50"`
	ImageURL   string `json:"imageUrl" binding:"required,max=255"`
	LinkURL    string `json:"linkUrl" binding:"omitempty,max=255"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

type UpdateAdRequest struct {
	PositionID uint   `json:"positionId" binding:"omitempty"`
	Title      string `json:"title" binding:"omitempty,max=50"`
	ImageURL   string `json:"imageUrl" binding:"omitempty,max=255"`
	LinkURL    string `json:"linkUrl" binding:"omitempty,max=255"`
	StartTime  string `json:"startTime" binding:"omitempty"`
	EndTime    string `json:"endTime" binding:"omitempty"`
}

type ReviewAdRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int  `json:"status" binding:"required,oneof=0 1 2"`
}

func (c *AdminAdController) CreateAd(ctx *gin.Context) {
	var req CreateAdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	ad, err := c.adService.CreateAd(req.PositionID, req.Title, req.ImageURL, req.LinkURL, req.StartTime, req.EndTime, actor.AdminID)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adToResponse(ad)))
}

func (c *AdminAdController) GetAd(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	ad, err := c.adService.GetAdByID(id)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeNotFound, "广告不存在"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adToResponse(ad)))
}

func (c *AdminAdController) ListAds(ctx *gin.Context) {
	positionID, _ := parseIntDefault(ctx.Query("positionId"), 0)
	status, _ := parseIntDefault(ctx.DefaultQuery("status", "-1"), -1)
	page, pageSize := ParsePageQuery(ctx, 100)
	ads, total, err := c.adService.ListAds(uint(positionID), status, page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	list := make([]types.AdvertisementResponse, len(ads))
	for i, a := range ads {
		list[i] = adToResponse(&a)
	}
	ctx.JSON(200, types.PageResponse(list, total, page, pageSize))
}

func (c *AdminAdController) UpdateAd(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	var req UpdateAdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	ad, err := c.adService.UpdateAd(id, req.PositionID, req.Title, req.ImageURL, req.LinkURL, req.StartTime, req.EndTime)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adToResponse(ad)))
}

func (c *AdminAdController) DeleteAd(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	if err := c.adService.DeleteAd(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminAdController) ReviewAd(ctx *gin.Context) {
	var req ReviewAdRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adService.ReviewAd(req.ID, req.Status, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminAdController) IncrementAdView(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	if err := c.adService.IncrementViewCount(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminAdController) IncrementAdClick(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	if err := c.adService.IncrementClickCount(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminAdController) GetAdStats(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	stats, err := c.adService.GetAdStats(startDate, endDate)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(stats))
}
