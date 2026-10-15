package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAdoptRoutes(r *gin.RouterGroup, adoptController *controller.AdoptController) {
	// 领养公开路由（可选登录，用于联系方式脱敏）
	adopts := r.Group("/adopts")
	adopts.Use(middleware.OptionalUserAuthMiddleware())
	{
		adopts.GET("", adoptController.ListAdopts)
		adopts.GET("/:id", adoptController.GetAdopt)
		adopts.GET("/:id/applications", adoptController.ListApplicationsByAdoption)
		adopts.GET("/stats", adoptController.GetAdoptionStats)

		// 需要用户认证的路由
		auth := adopts.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.POST("", adoptController.CreateAdopt)
			auth.POST("/upload-photo", adoptController.UploadAdoptPhoto)
			auth.PUT("/:id", adoptController.UpdateAdopt)
			auth.DELETE("/:id", adoptController.DeleteAdopt)
			auth.GET("/my", adoptController.ListMyAdopts)
			auth.POST("/status", adoptController.UpdateAdoptStatus)
			auth.POST("/applications", adoptController.CreateApplication)
			auth.GET("/applications/my", adoptController.ListMyApplications)
			auth.POST("/applications/:id/cancel", adoptController.CancelApplication)
			auth.POST("/applications/review", adoptController.ReviewApplication)
			auth.POST("/applications/sign", adoptController.SignAgreement)
			auth.POST("/follow-ups", adoptController.CreateFollowUp)
			auth.GET("/:id/follow-ups", adoptController.ListFollowUpsByAdoption)
		}
	}
}
