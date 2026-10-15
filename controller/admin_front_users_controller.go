package controller

import (
	"pet-community/model"
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminFrontUsersController 前台用户管理:列表/封禁/解禁/批量封禁。
// RBAC: super_admin | service。
type AdminFrontUsersController struct {
	adminService *service.AdminFrontUserService
	logService   *service.AdminLogService
}

func NewAdminFrontUsersController(adminService *service.AdminFrontUserService, logService *service.AdminLogService) *AdminFrontUsersController {
	return &AdminFrontUsersController{adminService: adminService, logService: logService}
}

func frontUserToResponse(user *model.User) types.UserInfo {
	return types.UserInfo{
		ID:        user.ID,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		Email:     user.Email,
		Gender:    user.Gender,
		Signature: user.Signature,
		Location:  user.Location,
		PetYears:  user.PetYears,
		Status:    user.Status,
	}
}

type BanFrontUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type BatchBanFrontUsersRequest struct {
	UserIDs []uint `json:"userIds" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
}

func (c *AdminFrontUsersController) ListFrontUsers(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	status, _ := parseIntDefault(ctx.Query("status"), -1)
	page, pageSize := ParsePageQuery(ctx, 100)
	users, total, err := c.adminService.ListFrontUsers(keyword, status, page, pageSize)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, "获取用户列表失败"))
		return
	}
	list := make([]types.UserInfo, len(users))
	for i, u := range users {
		list[i] = frontUserToResponse(&u)
	}
	ctx.JSON(200, types.PageResponse(list, total, page, pageSize))
}

func (c *AdminFrontUsersController) GetFrontUser(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	user, err := c.adminService.GetFrontUserDetail(id)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeNotFound, "用户不存在"))
		return
	}
	ctx.JSON(200, types.SuccessResponse(frontUserToResponse(user)))
}

func (c *AdminFrontUsersController) BanFrontUser(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	var req BanFrontUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	if err := c.adminService.BanFrontUser(id, req.Reason, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminFrontUsersController) UnbanFrontUser(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "用户ID")
	if !ok {
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.UnbanFrontUser(id, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminFrontUsersController) BatchBanFrontUsers(ctx *gin.Context) {
	var req BatchBanFrontUsersRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.BatchBanFrontUsers(req.UserIDs, req.Reason, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminFrontUsersController) GetFrontUserOperationLogs(ctx *gin.Context) {
	userID, _ := parseIntDefault(ctx.Query("userId"), 0)
	action := ctx.Query("action")
	startTime, endTime := ParseTimeRange(ctx.Query("startTime"), ctx.Query("endTime"))
	page, limit := ParseLimitQuery(ctx, 100)
	logs, total, err := c.logService.GetFrontUserOperationLogs(uint(userID), action, startTime, endTime, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": logs, "total": total}))
}

func parseIntDefault(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	n, err := parseInt(s)
	if err != nil {
		return def, err
	}
	return n, nil
}
