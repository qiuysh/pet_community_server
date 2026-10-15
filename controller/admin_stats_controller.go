package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminStatsController 统计与数据报表。
// RBAC: super_admin | data_admin。
type AdminStatsController struct {
	statsService *service.StatsService
}

func NewAdminStatsController(statsService *service.StatsService) *AdminStatsController {
	return &AdminStatsController{statsService: statsService}
}

type ExportStatsRequest struct {
	Type      string `json:"type" binding:"required,oneof=overview user_growth content interaction"`
	StartDate string `json:"startDate" binding:"required"`
	EndDate   string `json:"endDate" binding:"required"`
}

func (c *AdminStatsController) GetOverviewStats(ctx *gin.Context) {
	stats, err := c.statsService.GetOverviewStats()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(stats))
}

func (c *AdminStatsController) GetUserGrowthStats(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	if startDate == "" || endDate == "" {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "缺少日期参数"))
		return
	}
	stats, err := c.statsService.GetUserGrowthStats(startDate, endDate)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(stats))
}

func (c *AdminStatsController) GetContentStats(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	if startDate == "" || endDate == "" {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "缺少日期参数"))
		return
	}
	stats, err := c.statsService.GetContentStats(startDate, endDate)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(stats))
}

func (c *AdminStatsController) GetInteractionStats(ctx *gin.Context) {
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	if startDate == "" || endDate == "" {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "缺少日期参数"))
		return
	}
	stats, err := c.statsService.GetInteractionStats(startDate, endDate)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(stats))
}

func (c *AdminStatsController) GetTrendData(ctx *gin.Context) {
	trendType := ctx.DefaultQuery("type", "user_growth")
	days, _ := parseIntDefault(ctx.DefaultQuery("days", "7"), 7)
	data, err := c.statsService.GetTrendData(trendType, days)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(data))
}

func (c *AdminStatsController) GetComparisonStats(ctx *gin.Context) {
	trendType := ctx.DefaultQuery("type", "user_growth")
	compareType := ctx.DefaultQuery("compare", "week")
	data, err := c.statsService.GetComparisonData(trendType, compareType)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(data))
}

func (c *AdminStatsController) GetDistributionStats(ctx *gin.Context) {
	distType := ctx.DefaultQuery("type", "pet_type")
	data, err := c.statsService.GetDistributionData(distType)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(data))
}

func (c *AdminStatsController) ExportStats(ctx *gin.Context) {
	var req ExportStatsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	data, err := c.statsService.ExportStatsToExcel(req.Type, req.StartDate, req.EndDate)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	ctx.Header("Content-Disposition", "attachment; filename=stats_export.xlsx")
	ctx.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
