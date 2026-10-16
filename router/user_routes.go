package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.RouterGroup, userController *controller.UserController) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", userController.Register)
		auth.POST("/login", userController.Login)
		auth.POST("/send-code", userController.SendCode)
		auth.POST("/reset-password", userController.ResetPassword)
		auth.POST("/logout", middleware.UserAuthMiddleware(), userController.Logout)
		auth.POST("/third-party-login", userController.ThirdPartyLogin)
		auth.POST("/refresh-token", userController.RefreshToken)
	}

	// 公开个人主页（游客可访问；登录时校验拉黑）
	r.GET("/users/:id/profile", middleware.OptionalUserAuthMiddleware(), userController.GetPublicProfile)

	user := r.Group("/user")
	user.Use(middleware.UserAuthMiddleware())
	{
		user.GET("/profile", userController.GetProfile)
		user.PUT("/profile", userController.UpdateProfile)
		user.POST("/avatar", userController.UploadAvatar)
		user.PUT("/password", userController.ChangePassword)
		user.GET("/devices", userController.GetDevices)
		user.DELETE("/devices/:deviceId", userController.RemoveDevice)
		user.GET("/privacy", userController.GetPrivacySettings)
		user.PUT("/privacy", userController.UpdatePrivacySettings)
		user.POST("/unbind-phone", userController.UnbindPhone)
		user.POST("/unbind-email", userController.UnbindEmail)
		user.POST("/bind-phone", userController.BindPhone)
		user.POST("/bind-email", userController.BindEmail)
		user.GET("/third-party", userController.ListThirdPartyAccounts)
		user.POST("/third-party/bind", userController.BindThirdPartyAccount)
		user.DELETE("/third-party/:platform", userController.UnbindThirdPartyAccount)
	}
}
