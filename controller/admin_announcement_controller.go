package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminAnnouncementController 公告管理。
// RBAC: super_admin | content_admin (公告属于内容管理)。
type AdminAnnouncementController struct {
	adminService *service.AdminAnnouncementService
}

func NewAdminAnnouncementController(adminService *service.AdminAnnouncementService) *AdminAnnouncementController {
	return &AdminAnnouncementController{adminService: adminService}
}

type CreateAnnouncementRequest struct {
	Title     string  `json:"title" binding:"required"`
	Content   string  `json:"content" binding:"required"`
	Type      int     `json:"type,omitempty"`
	SortOrder int     `json:"sortOrder,omitempty"`
	StartAt   *string `json:"startAt,omitempty"`
	EndAt     *string `json:"endAt,omitempty"`
}

type UpdateAnnouncementRequest struct {
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Type      int     `json:"type,omitempty"`
	SortOrder int     `json:"sortOrder,omitempty"`
	StartAt   *string `json:"startAt,omitempty"`
	EndAt     *string `json:"endAt,omitempty"`
}

type UpdateAnnouncementStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=0 1"`
}

func (c *AdminAnnouncementController) CreateAnnouncement(ctx *gin.Context) {
	var req CreateAnnouncementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	startAt := ParseTimePtr(req.StartAt)
	endAt := ParseTimePtr(req.EndAt)
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	announcement, err := c.adminService.CreateAnnouncement(req.Title, req.Content, req.Type, req.SortOrder, startAt, endAt, actor.AdminID, actor.IP)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(announcement))
}

func (c *AdminAnnouncementController) UpdateAnnouncement(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "公告ID")
	if !ok {
		return
	}
	var req UpdateAnnouncementRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	startAt := ParseTimePtr(req.StartAt)
	endAt := ParseTimePtr(req.EndAt)
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	announcement, err := c.adminService.UpdateAnnouncement(id, req.Title, req.Content, req.Type, req.SortOrder, startAt, endAt, actor.AdminID, actor.IP)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(announcement))
}

func (c *AdminAnnouncementController) DeleteAnnouncement(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "公告ID")
	if !ok {
		return
	}
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	if err := c.adminService.DeleteAnnouncement(id, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminAnnouncementController) GetAnnouncement(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "公告ID")
	if !ok {
		return
	}
	announcement, err := c.adminService.GetAnnouncementByID(id)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeNotFound, "公告不存在"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(announcement))
}

func (c *AdminAnnouncementController) ListAnnouncements(ctx *gin.Context) {
	announcementType, _ := parseIntDefault(ctx.DefaultQuery("type", "0"), 0)
	status, _ := parseIntDefault(ctx.DefaultQuery("status", "-1"), -1)
	page, limit := ParseLimitQuery(ctx, 100)
	announcements, total, err := c.adminService.ListAnnouncements(announcementType, status, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": announcements, "total": total}))
}

func (c *AdminAnnouncementController) UpdateAnnouncementStatus(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "公告ID")
	if !ok {
		return
	}
	var req UpdateAnnouncementStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	if err := c.adminService.UpdateAnnouncementStatus(id, req.Status, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminAnnouncementController) PinAnnouncement(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "公告ID")
	if !ok {
		return
	}
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	if err := c.adminService.PinAnnouncement(id, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}
