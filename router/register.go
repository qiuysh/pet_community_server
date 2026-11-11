package router

import (
	"pet-community/controller"
	"pet-community/middleware"
	"pet-community/types"
	"pet-community/utils"

	"github.com/gin-gonic/gin"
)

// Handlers 聚合路由注册所需的全部 controller。
//
// 存在的意义是「路由表的唯一入口」：main 与契约测试（tests/api_contract_test.go）
// 都调用 RegisterRoutes，测试拿到的路由集合与线上完全一致。此前路由注册散落在
// cmd/main.go，任何静态校验都只能复制一份注册代码，复制品一旦落后就失去意义。
//
// 字段允许为 nil：gin 注册 handler 时只保存方法值，不会立即调用，
// 因此契约测试可以传零值 Handlers 拿到完整路由表而不需要数据库。
type Handlers struct {
	Admin                AdminControllers
	User                 *controller.UserController
	Pet                  *controller.PetController
	Post                 *controller.PostController
	Notification         *controller.NotificationController
	Follow               *controller.FollowController
	Block                *controller.BlockController
	Message              *controller.MessageController
	Points               *controller.PointsController
	NotificationSettings *controller.NotificationSettingsController
	Discovery            *controller.DiscoveryController
	Activity             *controller.ActivityController
	Adopt                *controller.AdoptController
	FrontAd              *controller.FrontAdController
	Site                 *controller.SiteController
	Upload               *controller.UploadController
}

// RegisterRoutes 注册全部 HTTP 路由（不含中间件装配与静态文件挂载）。
//
// WebSocket 入口 /ws 由调用方自行挂载：它需要把 WebSocketService 实例带进来，
// 而该实例属于运行时状态，不适合塞进 Handlers。
func RegisterRoutes(engine *gin.Engine, h Handlers) {
	engine.GET("/health", h.Admin.Auth.HealthCheck)
	engine.GET("/api/v1/csrf-token", middleware.IssueCSRFToken)
	engine.GET("/api/v1/public-key", func(c *gin.Context) {
		c.JSON(200, types.SuccessResponse(gin.H{
			"publicKey": utils.GetPublicKey(),
		}))
	})

	RegisterAdminRoutes(engine, h.Admin)

	apiV1 := engine.Group("/api/v1")
	SetupUserRoutes(apiV1, h.User)
	SetupPetRoutes(apiV1, h.Pet)
	SetupPostRoutes(apiV1, h.Post)
	SetupNotificationRoutes(apiV1, h.Notification, h.Follow)
	SetupSocialRoutes(apiV1, h.Block, h.Message, h.Points, h.NotificationSettings, h.Follow)
	SetupDiscoveryRoutes(apiV1, h.Discovery)
	SetupActivityRoutes(apiV1, h.Activity)
	SetupAdoptRoutes(apiV1, h.Adopt)
	SetupAdRoutes(apiV1, h.FrontAd)
	SetupSiteRoutes(apiV1, h.Site)
	SetupUploadRoutes(apiV1, h.Upload)
}
