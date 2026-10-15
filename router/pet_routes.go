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
		pet.GET("", petController.ListPets)

		auth := pet.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.GET("/mine/list", petController.GetMyPets)
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

		pet.GET("/:id", petController.GetPet)
		pet.GET("/:id/album", petController.GetPetAlbum)
		pet.GET("/:id/growth-records", petController.GetGrowthRecords)
	}

	lost := r.Group("/lost-pets")
	lost.Use(middleware.OptionalUserAuthMiddleware())
	{
		lost.GET("", petController.ListLostPets)
		lost.GET("/:id", petController.GetLostPet)

		auth := lost.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.POST("", petController.CreateLostPet)
			auth.PUT("/:id/status", petController.UpdateLostPetStatus)
		}
	}
}
