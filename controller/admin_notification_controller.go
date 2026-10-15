package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminNotificationController 站内通知管理(对前台用户的系统通知)。
// RBAC: super_admin | service。
type AdminNotificationController struct {
	adminService *service.AdminNotificationService
}

func NewAdminNotificationController(adminService *service.AdminNotificationService) *AdminNotificationController {
	return &AdminNotificationController{adminService: adminService}
}

type CreateNotificationRequest struct {
	Title      string  `json:"title" binding:"required"`
	Content    string  `json:"content" binding:"required"`
	Type       int     `json:"type" binding:"required,oneof=1 2 3"`
	TargetType int     `json:"targetType" binding:"required,oneof=0 1 2"`
	TargetIDs  []uint  `json:"targetIds,omitempty"`
	PublishAt  *string `json:"publishAt,omitempty"`
}

type UpdateNotificationRequest struct {
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Type       int     `json:"type,omitempty" binding:"omitempty,oneof=1 2 3"`
	TargetType int     `json:"targetType,omitempty" binding:"omitempty,oneof=0 1 2"`
	TargetIDs  []uint  `json:"targetIds,omitempty"`
	PublishAt  *string `json:"publishAt,omitempty"`
}

func (c *AdminNotificationController) CreateNotification(ctx *gin.Context) {
	var req CreateNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	publishAt := ParseTimePtr(req.PublishAt)
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	notification, err := c.adminService.CreateNotification(req.Title, req.Content, req.Type, req.TargetType, req.TargetIDs, publishAt, actor.AdminID, actor.IP)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(notification))
}

func (c *AdminNotificationController) UpdateNotification(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "通知ID")
	if !ok {
		return
	}
	var req UpdateNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	publishAt := ParseTimePtr(req.PublishAt)
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	notification, err := c.adminService.UpdateNotification(id, req.Title, req.Content, req.Type, req.TargetType, req.TargetIDs, publishAt, actor.AdminID, actor.IP)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(notification))
}

func (c *AdminNotificationController) DeleteNotification(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "通知ID")
	if !ok {
		return
	}
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	if err := c.adminService.DeleteNotification(id, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminNotificationController) GetNotification(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "通知ID")
	if !ok {
		return
	}
	notification, err := c.adminService.GetNotificationByID(id)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeNotFound, "通知不存在"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(notification))
}

func (c *AdminNotificationController) ListNotifications(ctx *gin.Context) {
	notificationType, _ := parseIntDefault(ctx.DefaultQuery("type", "0"), 0)
	status, _ := parseIntDefault(ctx.DefaultQuery("status", "-1"), -1)
	page, limit := ParseLimitQuery(ctx, 100)
	notifications, total, err := c.adminService.ListNotifications(notificationType, status, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": notifications, "total": total}))
}

func (c *AdminNotificationController) SendNotification(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "通知ID")
	if !ok {
		return
	}
	actor, ok2 := MustGetAdminActor(ctx)
	if !ok2 {
		return
	}
	if err := c.adminService.SendNotification(id, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}
