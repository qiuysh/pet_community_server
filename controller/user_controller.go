package controller

import (
	"net/http"
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) Register(ctx *gin.Context) {
	var req types.UserRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	ip := getClientIP(ctx)

	user, err := c.userService.Register(&req, ip)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, user)
}

func (c *UserController) Login(ctx *gin.Context) {
	var req types.UserLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	ip := getClientIP(ctx)
	userAgent := ctx.GetHeader("User-Agent")

	resp, err := c.userService.Login(&req, ip, userAgent)
	if err != nil {
		utils.RespondError(ctx, types.CodeUnauthorized, err.Error())
		return
	}

	utils.RespondSuccess(ctx, resp)
}

func (c *UserController) SendCode(ctx *gin.Context) {
	var req types.SendCodeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	ip := getClientIP(ctx)

	expiresIn, debugCode, err := c.userService.SendCode(req.Account, req.Type, ip)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.SendCodeResponse{ExpiresIn: expiresIn, DebugCode: debugCode})
}

func (c *UserController) GetProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	profile, err := c.userService.GetProfile(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, profile)
}

// GetPublicProfile GET /users/:id/profile — 公开主页
func (c *UserController) GetPublicProfile(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.RespondError(ctx, types.CodeParamError, "无效的用户ID")
		return
	}

	viewerID := ctx.GetUint("user_id")
	profile, err := c.userService.GetPublicProfile(uint(id), viewerID)
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, profile)
}

func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.UserProfileUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	profile, err := c.userService.UpdateProfile(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, profile)
}

func (c *UserController) ChangePassword(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.PasswordChangeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.ChangePassword(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *UserController) ResetPassword(ctx *gin.Context) {
	var req types.PasswordResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.ResetPassword(&req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *UserController) Logout(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.Logout(userID, req.DeviceID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *UserController) GetDevices(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	devices, err := c.userService.GetDevices(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, devices)
}

func (c *UserController) RemoveDevice(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	deviceID := ctx.Param("deviceId")
	if deviceID == "" {
		utils.RespondError(ctx, types.CodeParamError, "设备ID无效")
		return
	}

	if err := c.userService.RemoveDeviceByID(userID, deviceID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *UserController) AdminListUsers(ctx *gin.Context) {
	var req types.UserListRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.userService.AdminListUsers(&req)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *UserController) AdminUpdateUserStatus(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	var body struct {
		Status int `json:"status" binding:"required,oneof=0 1 2"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.AdminUpdateUserStatus(uint(userID), body.Status); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, types.SuccessResponse(nil))
}

func (c *UserController) ThirdPartyLogin(ctx *gin.Context) {
	var req types.ThirdPartyLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	ip := getClientIP(ctx)
	userAgent := ctx.GetHeader("User-Agent")

	resp, err := c.userService.ThirdPartyLogin(&req, ip, userAgent)
	if err != nil {
		utils.RespondError(ctx, types.CodeUnauthorized, err.Error())
		return
	}

	utils.RespondSuccess(ctx, resp)
}

func (c *UserController) RefreshToken(ctx *gin.Context) {
	var req types.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	resp, err := c.userService.RefreshToken(&req)
	if err != nil {
		utils.RespondError(ctx, types.CodeUnauthorized, err.Error())
		return
	}

	utils.RespondSuccess(ctx, resp)
}

func (c *UserController) GetPrivacySettings(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	settings, err := c.userService.GetPrivacySettings(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, settings)
}

func (c *UserController) UpdatePrivacySettings(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.PrivacySettingsUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.UpdatePrivacySettings(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *UserController) UnbindPhone(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.UnbindPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.UnbindPhone(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *UserController) UnbindEmail(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.UnbindEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.UnbindEmail(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// BindPhone 绑定或更换手机号
func (c *UserController) BindPhone(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.BindPhoneRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.BindPhone(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	// 绑定成功后回写当前用户信息,前端可直接刷新本地 store
	user, _ := c.userService.GetProfile(userID)
	utils.RespondSuccess(ctx, user)
}

// BindEmail 绑定或更换邮箱
func (c *UserController) BindEmail(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.BindEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.userService.BindEmail(userID, &req); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	user, _ := c.userService.GetProfile(userID)
	utils.RespondSuccess(ctx, user)
}

func getClientIP(ctx *gin.Context) string {
	ip := ctx.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = ctx.ClientIP()
	}
	return ip
}
