package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPostRoutes(r *gin.RouterGroup, postController *controller.PostController) {
	topics := r.Group("/topics")
	{
		topics.GET("", postController.ListTopics)
		topics.GET("/search", postController.SearchTopics)
		topics.GET("/:id", postController.GetTopic)
		topics.POST("", middleware.UserAuthMiddleware(), postController.CreateTopic)
	}

	posts := r.Group("/posts")
	posts.Use(middleware.OptionalUserAuthMiddleware())
	{
		posts.GET("", postController.ListPosts)
		posts.GET("/:id", postController.GetPost)
		posts.GET("/:id/comments", postController.ListComments)

		auth := posts.Group("")
		auth.Use(middleware.UserAuthMiddleware())
		{
			auth.POST("", postController.CreatePost)
			auth.PUT("/:id", postController.UpdatePost)
			auth.POST("/:id/cancel-schedule", postController.CancelSchedule)
			auth.DELETE("/:id", postController.DeletePost)
			auth.POST("/:id/like", postController.ToggleLike)
			auth.GET("/:id/likes", postController.GetLikeList)
			auth.POST("/:id/favorite", postController.ToggleFavorite)
			auth.POST("/:id/share", postController.RecordShare)

			auth.POST("/:id/comments", postController.CreateComment)
			auth.PUT("/comments/:id", postController.UpdateComment)
			auth.DELETE("/comments/:id", postController.DeleteComment)
			auth.POST("/comments/:id/like", postController.ToggleCommentLike)

			auth.POST("/reports", postController.CreateReport)
			auth.GET("/reports", postController.ListReports)
			auth.GET("/reports/:id", postController.GetReport)
		}
	}

	drafts := r.Group("/drafts")
	drafts.Use(middleware.UserAuthMiddleware())
	{
		drafts.POST("", postController.CreateDraft)
		drafts.PUT("/:id", postController.UpdateDraft)
		drafts.GET("", postController.ListDrafts)
		drafts.GET("/:id", postController.GetDraft)
		drafts.DELETE("/:id", postController.DeleteDraft)
	}

	favorites := r.Group("/favorites")
	favorites.Use(middleware.UserAuthMiddleware())
	{
		favorites.GET("", postController.GetFavorites)
	}

	likes := r.Group("/likes")
	likes.Use(middleware.UserAuthMiddleware())
	{
		likes.GET("", postController.GetLikedPosts)
	}
}
