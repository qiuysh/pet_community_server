package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupSocialRoutes(r *gin.RouterGroup, blockController *controller.BlockController, messageController *controller.MessageController, pointsController *controller.PointsController, notificationSettingsController *controller.NotificationSettingsController, followController *controller.FollowController) {
	social := r.Group("/social")
	social.Use(middleware.UserAuthMiddleware())

	block := social.Group("/block")
	{
		block.POST("/:id", blockController.BlockUser)
		block.DELETE("/:id", blockController.UnblockUser)
		block.GET("/list", blockController.GetBlockedList)
	}

	follow := social.Group("/follow")
	{
		follow.POST("/:id", followController.ToggleFollow)
		follow.DELETE("/:id", followController.Unfollow)
		follow.POST("/:id/secret", followController.SetSecretFollow)
		follow.GET("/secret/list", followController.GetSecretFollowing)
		follow.GET("/:id/followers", followController.GetFollowers)
		follow.GET("/:id/following", followController.GetFollowing)
		follow.GET("/:id/status", followController.GetFollowStatus)
		follow.GET("/status/:id", followController.GetFollowStatus)
		follow.GET("/:id/online", followController.GetOnlineStatus)
	}

	messages := social.Group("/messages")
	{
		messages.POST("/send", messageController.SendMessage)
		messages.POST("/upload/image", messageController.UploadMessageImage)
		messages.POST("/upload/voice", messageController.UploadMessageVoice)
		messages.GET("/conversations", messageController.GetConversations)
		messages.GET("/:id", messageController.GetMessages)
		messages.POST("/:id/read", messageController.MarkAsRead)
		messages.DELETE("/:id", messageController.DeleteConversation)
		messages.GET("/:id/search", messageController.SearchMessages)
		messages.POST("/recall/:msgId", messageController.RecallMessage)
		messages.GET("/unread/count", messageController.GetUnreadCount)
	}

	points := social.Group("/points")
	{
		points.GET("/level", pointsController.GetUserLevel)
		points.GET("/records", pointsController.GetPointsRecords)
		points.GET("/rules", pointsController.GetPointsRules)
		points.GET("/benefits", pointsController.GetBenefits)
		points.GET("/benefits/:level", pointsController.GetBenefitsByLevel)
	}

	notificationSettings := social.Group("/notification-settings")
	{
		notificationSettings.GET("", notificationSettingsController.GetSettings)
		notificationSettings.PUT("", notificationSettingsController.UpdateSettings)
	}
}
