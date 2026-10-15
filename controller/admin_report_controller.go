package controller

import (
	"strconv"

	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminReportController 举报管理：列表、处理、批量、统计。
// RBAC: super_admin | content_admin。
type AdminReportController struct {
	reportService *service.AdminReportService
}

func NewAdminReportController(reportService *service.AdminReportService) *AdminReportController {
	return &AdminReportController{reportService: reportService}
}

func (c *AdminReportController) ListReports(ctx *gin.Context) {
	page, pageSize := ParsePageQuery(ctx, 100)

	var statusPtr *int
	if statusStr := ctx.Query("status"); statusStr != "" {
		status, err := strconv.Atoi(statusStr)
		if err != nil || status < 0 || status > 2 {
			ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "无效的状态筛选"))
			return
		}
		statusPtr = &status
	}

	reasonType := ctx.Query("reasonType")
	if reasonType == "" {
		reasonType = ctx.Query("reason_type")
	}
	targetType := ctx.Query("targetType")
	if targetType == "" {
		targetType = ctx.Query("target_type")
	}

	result, err := c.reportService.ListReports(statusPtr, reasonType, targetType, page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(result))
}

func (c *AdminReportController) HandleReport(ctx *gin.Context) {
	reportID, ok := MustParseUintParam(ctx, "id", "举报ID")
	if !ok {
		return
	}
	var req types.HandleReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.reportService.HandleReport(reportID, req.Action, req.Result, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminReportController) BatchHandleReports(ctx *gin.Context) {
	var req types.BatchHandleReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.reportService.BatchHandleReports(req.IDs, req.Action, req.Result, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminReportController) GetReportStats(ctx *gin.Context) {
	stats, err := c.reportService.GetStats()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(stats))
}
