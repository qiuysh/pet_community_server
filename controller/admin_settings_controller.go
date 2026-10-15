package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// AdminSettingsController 站点配置 + 审核配置 + 敏感词管理。
// RBAC: site config/audit config → super_admin only; sensitive words → super_admin | content_admin。
type AdminSettingsController struct {
	adminService *service.AdminConfigService
}

func NewAdminSettingsController(adminService *service.AdminConfigService) *AdminSettingsController {
	return &AdminSettingsController{adminService: adminService}
}

// ========= 站点配置 =========

type UpdateSiteConfigRequest struct {
	SiteName        string `json:"siteName"`
	SiteLogo        string `json:"siteLogo"`
	SiteDescription string `json:"siteDescription"`
	SiteICP         string `json:"siteICP"`
}

func (c *AdminSettingsController) GetSiteConfig(ctx *gin.Context) {
	config, err := c.adminService.GetSiteConfig()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(config))
}

func (c *AdminSettingsController) UpdateSiteConfig(ctx *gin.Context) {
	var req UpdateSiteConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	configs := map[string]string{
		"site_name":        req.SiteName,
		"site_logo":        req.SiteLogo,
		"site_description": req.SiteDescription,
		"site_icp":         req.SiteICP,
	}
	if err := c.adminService.UpdateSiteConfig(configs, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

// ========= 审核配置 =========

type UpdateAuditConfigRequest struct {
	AutoAuditEnabled       bool `json:"autoAuditEnabled"`
	SensitiveFilterEnabled bool `json:"sensitiveFilterEnabled"`
}

func (c *AdminSettingsController) GetAuditConfig(ctx *gin.Context) {
	config, err := c.adminService.GetAuditConfig()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(config))
}

func (c *AdminSettingsController) UpdateAuditConfig(ctx *gin.Context) {
	var req UpdateAuditConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.UpdateAuditConfig(req.AutoAuditEnabled, req.SensitiveFilterEnabled, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

// ========= 敏感词 =========

type AddSensitiveWordRequest struct {
	Word        string `json:"word"`
	ReplaceWith string `json:"replaceWith"`
	Category    string `json:"category"`
}

type UpdateSensitiveWordRequest struct {
	Word        string `json:"word"`
	ReplaceWith string `json:"replaceWith"`
	Category    string `json:"category"`
}

func (c *AdminSettingsController) GetSensitiveWords(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	page, limit := ParseLimitQuery(ctx, 100)
	words, total, err := c.adminService.GetSensitiveWords(keyword, page, limit)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(gin.H{"list": words, "total": total}))
}

func (c *AdminSettingsController) AddSensitiveWord(ctx *gin.Context) {
	var req AddSensitiveWordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	if req.ReplaceWith == "" {
		req.ReplaceWith = "***"
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.AddSensitiveWord(req.Word, req.ReplaceWith, req.Category, actor.AdminID, actor.IP); err != nil {
		if err.Error() == "敏感词已存在" {
			ctx.JSON(200, types.ErrorResponse(20001, err.Error()))
			return
		}
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminSettingsController) UpdateSensitiveWord(ctx *gin.Context) {
	id, err := parseInt(ctx.Param("id"))
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "无效的敏感词ID"))
		return
	}
	var req UpdateSensitiveWordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, err.Error()))
		return
	}
	if req.ReplaceWith == "" {
		req.ReplaceWith = "***"
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.UpdateSensitiveWord(uint(id), req.Word, req.ReplaceWith, req.Category, actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

func (c *AdminSettingsController) DeleteSensitiveWord(ctx *gin.Context) {
	id, err := parseInt(ctx.Param("id"))
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "无效的敏感词ID"))
		return
	}
	actor, ok := MustGetAdminActor(ctx)
	if !ok {
		return
	}
	if err := c.adminService.DeleteSensitiveWord(uint(id), actor.AdminID, actor.IP); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}
