package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminUsersController admin 账号管理:CRUD 后台管理员。
// RBAC: 全部 super_admin only。
type AdminUsersController struct {
	adminService *service.AdminAuthService
}

func NewAdminUsersController(adminService *service.AdminAuthService) *AdminUsersController {
	return &AdminUsersController{adminService: adminService}
}

type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"omitempty,email,max=100"`
	Role     string `json:"role" binding:"required,oneof=super_admin content_admin data_admin service admin"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type UpdateAdminRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email,max=100"`
	Role     string `json:"role" binding:"omitempty,oneof=super_admin content_admin data_admin service admin"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateStatusRequest struct {
	Status *int `json:"status" binding:"required,oneof=0 1"`
}

func (c *AdminUsersController) ListAdmins(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	page, pageSize := ParsePageQuery(ctx, 100)
	admins, total, err := c.adminService.ListAdmins(keyword, page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, "获取用户列表失败"))
		return
	}
	ctx.JSON(200, types.PageResponse(adminListToResponse(admins), total, page, pageSize))
}

func (c *AdminUsersController) GetAdmin(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	admin, err := c.adminService.GetAdminByID(id)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeNotFound, "用户不存在"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adminToResponse(admin)))
}

func (c *AdminUsersController) CreateAdmin(ctx *gin.Context) {
	var req CreateAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	admin, err := c.adminService.CreateAdmin(req.Username, req.Password, req.Email, req.Role, status)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adminToResponse(admin)))
}

func (c *AdminUsersController) UpdateAdmin(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	var req UpdateAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	admin, err := c.adminService.UpdateAdmin(id, req.Username, req.Email, req.Role, req.Status)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(adminToResponse(admin)))
}

func (c *AdminUsersController) UpdateAdminPassword(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	var req UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	if err := c.adminService.UpdateAdminPassword(id, req.Password); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminUsersController) DeleteAdmin(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	if err := c.adminService.DeleteAdmin(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, "删除用户失败"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

// UpdateAdminStatus 禁用/启用 admin 账号。修复 B1:使用 MustGetAdminActor 获取正确 adminID。
func (c *AdminUsersController) UpdateAdminStatus(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	var req UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	// super_admin_only 兜底:路由层已挂 RequireRole(super_admin),此处再兜一层
	if err := c.adminService.UpdateAdminStatus(id, actor.AdminID, *req.Status); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}
