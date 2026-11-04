package controller

import (
	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// SiteController 前台站点配置只读接口。
// 与 admin 的 /admin/settings/site 分离：此处只暴露可公开的站点名称、Logo、描述与备案号，
// 不挂鉴权中间件，访客亦可调用（前台标题/Logo/页脚备案号渲染依赖它）。
type SiteController struct {
	configService *service.AdminConfigService
}

// NewSiteController 构造前台站点配置 controller。
func NewSiteController(configService *service.AdminConfigService) *SiteController {
	return &SiteController{configService: configService}
}

// GetSiteConfig 返回站点公开配置（camelCase，与后台接口保持同构）。
func (c *SiteController) GetSiteConfig(ctx *gin.Context) {
	config, err := c.configService.GetSiteConfig()
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(SiteConfigResponse{
		SiteName:        config["site_name"],
		SiteLogo:        config["site_logo"],
		SiteDescription: config["site_description"],
		SiteICP:         config["site_icp"],
	}))
}
