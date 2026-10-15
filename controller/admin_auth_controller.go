package controller

import (
	"pet-community/model"
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminAuthController admin 登录/健康检查等无需 RBAC 路由的 handler。
type AdminAuthController struct {
	adminService *service.AdminAuthService
}

func NewAdminAuthController(adminService *service.AdminAuthService) *AdminAuthController {
	return &AdminAuthController{adminService: adminService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *AdminAuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}

	ip := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	token, admin, err := c.adminService.Login(req.Username, req.Password, ip, userAgent)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeUnauthorized, err.Error()))
		return
	}

	ctx.JSON(200, types.SuccessResponse(gin.H{
		"token": token,
		"admin": adminToResponse(admin),
	}))
}

func (c *AdminAuthController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, types.SuccessResponse(nil))
}

// ========= DTO mappers (从原 admin_controller.go 搬过来) =========

func adminToResponse(admin *model.AdminUser) types.AdminResponse {
	return types.AdminResponse{
		ID:        admin.ID,
		Username:  admin.Username,
		Email:     admin.Email,
		Role:      admin.Role,
		Status:    admin.Status,
		CreatedAt: admin.CreatedAt.Format(timestampFmt),
		UpdatedAt: admin.UpdatedAt.Format(timestampFmt),
	}
}

func adminListToResponse(admins []model.AdminUser) []types.AdminResponse {
	result := make([]types.AdminResponse, 0, len(admins))
	for _, admin := range admins {
		result = append(result, adminToResponse(&admin))
	}
	return result
}

// timestampFmt 统一时间格式化常量
const timestampFmt = "2006-01-02T15:04:05Z07:00"
