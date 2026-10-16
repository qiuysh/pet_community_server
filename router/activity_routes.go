package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupActivityRoutes(r *gin.RouterGroup, activityController *controller.ActivityController) {
	// 活动管理路由（管理员：super_admin / content_admin）
	adminActivities := r.Group("/admin/activities")
	adminActivities.Use(
		middleware.AuthMiddleware(),
		middleware.RequireRole(middleware.RoleSuperAdmin, middleware.RoleContentAdmin),
	)
	{
		adminActivities.GET("", activityController.ListAdminActivities)
		adminActivities.POST("", activityController.CreateActivity)
		adminActivities.PUT("/:id", activityController.UpdateActivity)
		adminActivities.DELETE("/:id", activityController.DeleteActivity)
		adminActivities.POST("/:id/publish", activityController.PublishActivity)
		adminActivities.GET("/:id/registrations", activityController.ListRegistrations)
		adminActivities.POST("/registrations/:id/approve", activityController.ApproveRegistration)
		adminActivities.POST("/submissions/award", activityController.AwardSubmission)
		// 评选功能路由
		adminActivities.GET("/:id/award", activityController.GetAwardPage)
		adminActivities.POST("/submissions/batch-award", activityController.BatchAwardSubmission)
		adminActivities.POST("/:id/publish-award", activityController.PublishAwardResult)
		// 签到管理路由
		adminActivities.GET("/:id/sign-ins", activityController.ListSignIns)
		adminActivities.GET("/:id/sign-ins/stats", activityController.GetSignInStats)
		// 相册管理路由
		adminActivities.GET("/:id/album", activityController.ListPhotos)
		adminActivities.DELETE("/:id/album/photos/:photoId", activityController.DeletePhoto)
		// 活动回顾路由
		adminActivities.POST("/:id/review", activityController.PublishReview)
		adminActivities.GET("/:id/review", activityController.GetReview)
	}

	// 活动公开路由
	activities := r.Group("/activities")
	{
		activities.GET("", activityController.ListActivities)
		activities.GET("/:id", activityController.GetActivity)
		activities.GET("/:id/submissions", activityController.ListSubmissions)
		activities.GET("/submissions/:id", activityController.GetSubmission)
		// 签到公开路由
		activities.GET("/:id/sign-ins", activityController.ListSignIns)
		// 相册公开路由
		activities.GET("/:id/album", activityController.ListPhotos)
		// 活动回顾公开路由
		activities.GET("/:id/review", activityController.GetReview)

		// 需要用户认证的路由
		auth := activities.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.POST("/:id/register", activityController.Register)
			auth.POST("/:id/cancel", activityController.CancelRegistration)
			auth.POST("/submissions", activityController.CreateSubmission)
			auth.PUT("/submissions/:id", activityController.UpdateSubmission)
			auth.DELETE("/submissions/:id", activityController.DeleteSubmission)
			auth.POST("/vote", activityController.Vote)
			auth.POST("/submissions/:id/cancel-vote", activityController.CancelVote)
			auth.GET("/:id/votes/history", activityController.GetVoteHistory)
			auth.GET("/:id/votes/stats", activityController.GetUserVoteStats)
			// 签到认证路由
			auth.GET("/:id/qrcode", activityController.GenerateQRCode)
			auth.POST("/:id/sign-in", activityController.SignIn)
			// 相册认证路由
			auth.POST("/:id/album/photos", activityController.UploadPhoto)
			auth.DELETE("/:id/album/photos/:photoId", activityController.DeletePhoto)
		}
	}
}
