package router

import (
	"pet-community/controller"

	"github.com/gin-gonic/gin"
)

// SetupSiteRoutes 注册前台站点配置只读路由。
// 全部公开，不挂 AuthMiddleware：前台首屏需要用站点名称/Logo/备案号渲染标题与页脚。
//
// 路由:
//
//	GET /api/v1/site/config  获取站点公开配置
func SetupSiteRoutes(r *gin.RouterGroup, siteController *controller.SiteController) {
	site := r.Group("/site")
	{
		site.GET("/config", siteController.GetSiteConfig)
	}
}
