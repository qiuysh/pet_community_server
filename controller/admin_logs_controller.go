package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminLogsController 各类日志的读操作:operation/activity/login/behavior/security/export。
// RBAC: 所有已登录 admin 可见(读接口无需区分 role,rolePermissions 全部 logs:view)。
type AdminLogsController struct {
	adminService *service.AdminLogService
	authService  *service.AdminAuthService
	logService   *service.LogService
}

func NewAdminLogsController(adminService *service.AdminLogService, authService *service.AdminAuthService, logService *service.LogService) *AdminLogsController {
	return &AdminLogsController{adminService: adminService, authService: authService, logService: logService}
}

func (c *AdminLogsController) GetOperationLogs(ctx *gin.Context) {
	adminID, _ := parseIntDefault(ctx.Query("adminId"), 0)
	action := ctx.Query("action")
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.adminService.GetOperationLogs(uint(adminID), action, startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func (c *AdminLogsController) GetActivityLogs(ctx *gin.Context) {
	userID, _ := parseIntDefault(ctx.Query("userId"), 0)
	action := ctx.Query("action")
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.adminService.GetActivityLogs(uint(userID), action, startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func (c *AdminLogsController) GetLoginLogs(ctx *gin.Context) {
	adminID, _ := parseIntDefault(ctx.Query("adminId"), 0)
	status, _ := parseIntDefault(ctx.Query("status"), -1)
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.authService.GetLoginLogs(uint(adminID), status, startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func (c *AdminLogsController) GetUserBehaviorLogs(ctx *gin.Context) {
	userID, _ := parseIntDefault(ctx.Query("userId"), 0)
	actionType := ctx.Query("actionType")
	targetType := ctx.Query("targetType")
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.logService.GetBehaviorLogs(uint(userID), actionType, targetType, startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func (c *AdminLogsController) GetOperationLogsV2(ctx *gin.Context) {
	userID, _ := parseIntDefault(ctx.Query("userId"), 0)
	actionType := ctx.Query("actionType")
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.logService.GetOperationLogs(uint(userID), actionType, startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func (c *AdminLogsController) GetSecurityAuditLogs(ctx *gin.Context) {
	userID, _ := parseIntDefault(ctx.Query("userId"), 0)
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.logService.GetSecurityAuditLogs(uint(userID), startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func (c *AdminLogsController) ExportLogs(ctx *gin.Context) {
	userID, _ := parseIntDefault(ctx.Query("userId"), 0)
	actionType := ctx.Query("actionType")
	targetType := ctx.Query("targetType")
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	data, err := c.logService.ExportLogs(uint(userID), actionType, targetType, startTime, endTime)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.Header("Content-Type", "text/csv; charset=utf-8")
	ctx.Header("Content-Disposition", "attachment; filename=behavior_logs.csv")
	ctx.Data(200, "text/csv; charset=utf-8", data)
}

