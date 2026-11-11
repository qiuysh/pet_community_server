package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPetRoutes(r *gin.RouterGroup, petController *controller.PetController) {
	pet := r.Group("/pets")
	{
		pet.GET("/breeds", petController.GetBreeds)

		// 公开读接口：可选登录，便于 R5 粉丝可见判定
		petPublic := pet.Group("")
		petPublic.Use(middleware.OptionalUserAuthMiddleware())
		{
			petPublic.GET("", petController.ListPets)
			petPublic.GET("/:id", petController.GetPet)
			petPublic.GET("/:id/album", petController.GetPetAlbum)
			petPublic.GET("/:id/growth-records", petController.GetGrowthRecords)
		}

		auth := pet.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.GET("/mine/list", petController.GetMyPets)
			auth.POST("/upload-avatar", petController.UploadAvatar)
			auth.POST("/upload-album-photo", petController.UploadAlbumPhoto)
			auth.POST("/upload-album-video", petController.UploadAlbumVideo)
			auth.POST("", petController.CreatePet)
			auth.PUT("/:id", petController.UpdatePet)
			auth.DELETE("/:id", petController.DeletePet)
			auth.POST("/:id/album", petController.AddAlbumPhoto)
			auth.PUT("/:id/album/sort", petController.ReorderAlbum)
			auth.DELETE("/:id/album/:photoId", petController.DeleteAlbumPhoto)
			auth.POST("/:id/growth-records", petController.CreateGrowthRecord)
			auth.PUT("/:id/growth-records/:recordId", petController.UpdateGrowthRecord)
			auth.DELETE("/:id/growth-records/:recordId", petController.DeleteGrowthRecord)
		}
	}

	lost := r.Group("/lost-pets")
	lost.Use(middleware.OptionalUserAuthMiddleware())
	{
		lost.GET("", petController.ListLostPets)
		lost.GET("/:id", petController.GetLostPet)
		// R2：浏览型分享，游客也可上报（打开面板不请求，选定平台才 POST）
		lost.POST("/:id/share", petController.RecordLostPetShare)

		auth := lost.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.POST("/upload-photo", petController.UploadLostPetPhoto)
			auth.POST("", petController.CreateLostPet)
			auth.PUT("/:id/status", petController.UpdateLostPetStatus)
		}
	}
}
