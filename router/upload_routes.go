package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUploadRoutes 注册通用上传路由（需登录）。
// POST /api/v1/upload —— multipart field=file，返回 {url, type}。
func SetupUploadRoutes(r *gin.RouterGroup, uploadController *controller.UploadController) {
	upload := r.Group("/upload")
	upload.Use(middleware.UserAuthMiddleware())
	{
		upload.POST("", uploadController.UploadMedia)
	}
}
