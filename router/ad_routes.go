package router

import (
	"pet-community/controller"

	"github.com/gin-gonic/gin"
)

// SetupAdRoutes 注册前台广告展示路由。
// 与 admin 广告 CRUD/审核路由分离:此处全部为公开只读 + 计数上报,
// 不挂 AuthMiddleware,访客亦可调用。
//
// 路由:
//   GET    /api/v1/ads?location=home_banner  获取生效中的广告
//   POST   /api/v1/ads/:id/view              上报曝光 PV
//   POST   /api/v1/ads/:id/click             上报点击
func SetupAdRoutes(r *gin.RouterGroup, frontAdController *controller.FrontAdController) {
	ads := r.Group("/ads")
	{
		ads.GET("", frontAdController.ListActiveAds)
		ads.POST("/:id/view", frontAdController.IncrementAdView)
		ads.POST("/:id/click", frontAdController.IncrementAdClick)
	}
}
