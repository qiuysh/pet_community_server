package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupNotificationRoutes(r *gin.RouterGroup, notificationController *controller.NotificationController, followController *controller.FollowController) {
	notifications := r.Group("/notifications")
	notifications.Use(middleware.UserAuthMiddleware())
	{
		notifications.GET("", notificationController.ListNotifications)
		notifications.GET("/unread", notificationController.GetUnreadCount)
		notifications.POST("/:id/read", notificationController.MarkAsRead)
		notifications.POST("/read-all", notificationController.MarkAllAsRead)
		notifications.DELETE("/:id", notificationController.DeleteNotification)
	}

	follows := r.Group("/users")
	{
		follows.POST("/:id/follow", middleware.UserAuthMiddleware(), followController.ToggleFollow)
		follows.DELETE("/:id/follow", middleware.UserAuthMiddleware(), followController.Unfollow)
		follows.GET("/:id/followers", followController.GetFollowers)
		follows.GET("/:id/following", followController.GetFollowing)
		follows.GET("/:id/follow-status", middleware.UserAuthMiddleware(), followController.GetFollowStatus)
	}
}
