package controller

import (
	"errors"

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
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId"`
	CaptchaCode string `json:"captchaCode"`
}

func (c *AdminAuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}

	ip := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")

	token, admin, err := c.adminService.Login(req.Username, req.Password, ip, userAgent, req.CaptchaID, req.CaptchaCode)
	if err != nil {
		requireCaptcha := false
		var loginErr *service.AdminLoginError
		if errors.As(err, &loginErr) {
			requireCaptcha = loginErr.RequireCaptcha
		}
		// 登录失败是业务错误，不能用 401：前端会把 401 当成会话过期并清掉已有登录态
		ctx.JSON(200, types.ErrorResponseWithData(types.CodeBusinessError, err.Error(), gin.H{
			"requireCaptcha": requireCaptcha,
		}))
		return
	}

	ctx.JSON(200, types.SuccessResponse(gin.H{
		"token": token,
		"admin": adminToResponse(admin),
	}))
}

// GetCaptcha R9：后台登录图形验证码（公开，60 秒有效，一次性）。
func (c *AdminAuthController) GetCaptcha(ctx *gin.Context) {
	id, b64s, err := service.GenerateAdminCaptcha()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, "生成验证码失败"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{
		"captchaId":    id,
		"captchaImage": b64s,
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
