package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminDashboardController 运营驾驶舱 + 待办聚合。
// RBAC: 所有已登录 admin (概览无需区分)。
type AdminDashboardController struct {
	dashboardService *service.DashboardService
}

func NewAdminDashboardController(dashboardService *service.DashboardService) *AdminDashboardController {
	return &AdminDashboardController{dashboardService: dashboardService}
}

func (c *AdminDashboardController) GetDashboardOverview(ctx *gin.Context) {
	overview, err := c.dashboardService.GetDashboardOverview()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(overview))
}

func (c *AdminDashboardController) GetTodoAggregate(ctx *gin.Context) {
	aggregate, err := c.dashboardService.GetTodoAggregate()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(aggregate))
}
