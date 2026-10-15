package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminReviewController 内容审核:待审核列表、通过/拒绝、批量审核、历史记录。
// RBAC: super_admin | content_admin。
type AdminReviewController struct {
	adminService *service.AdminReviewService
}

func NewAdminReviewController(adminService *service.AdminReviewService) *AdminReviewController {
	return &AdminReviewController{adminService: adminService}
}

type ApproveReviewRequest struct {
	TargetType string `json:"targetType" binding:"required,oneof=post comment topic"`
	ID         uint   `json:"id" binding:"required"`
}

type RejectReviewRequest struct {
	TargetType string `json:"targetType" binding:"required,oneof=post comment topic"`
	ID         uint   `json:"id" binding:"required"`
	Reason     string `json:"reason"`
}

type BatchReviewRequest struct {
	TargetType string `json:"targetType" binding:"required,oneof=post comment topic"`
	IDs        []uint `json:"ids" binding:"required"`
	Status     int    `json:"status" binding:"required,oneof=1 2"`
	Reason     string `json:"reason,omitempty"`
}

func (c *AdminReviewController) GetPendingReviews(ctx *gin.Context) {
	reviewType := ctx.Query("type")
	page, pageSize := ParsePageQuery(ctx, 100)
	result, err := c.adminService.GetPendingReviews(reviewType, page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(result))
}

func (c *AdminReviewController) ApproveReview(ctx *gin.Context) {
	var req ApproveReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.ApproveReview(req.TargetType, req.ID, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminReviewController) RejectReview(ctx *gin.Context) {
	var req RejectReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.RejectReview(req.TargetType, req.ID, req.Reason, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminReviewController) BatchReview(ctx *gin.Context) {
	var req BatchReviewRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.BatchReview(req.TargetType, req.IDs, req.Status, req.Reason, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminReviewController) GetReviewHistory(ctx *gin.Context) {
	targetType := ctx.Query("type")
	page, pageSize := ParsePageQuery(ctx, 100)
	result, err := c.adminService.GetReviewHistory(targetType, page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(result))
}
