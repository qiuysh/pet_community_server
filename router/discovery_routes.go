package router

import (
	"pet-community/controller"
	"pet-community/middleware"

	"github.com/gin-gonic/gin"
)

func SetupDiscoveryRoutes(r *gin.RouterGroup, discoveryController *controller.DiscoveryController) {
	opt := middleware.OptionalUserAuthMiddleware()

	search := r.Group("/search")
	search.Use(opt)
	{
		search.GET("", discoveryController.Search)
		search.GET("/history", discoveryController.GetSearchHistory)
		search.DELETE("/history/:id", middleware.UserAuthMiddleware(), discoveryController.DeleteSearchHistory)
		search.DELETE("/history", middleware.UserAuthMiddleware(), discoveryController.ClearSearchHistory)
		search.GET("/hot-keywords", discoveryController.GetHotKeywords)
	}

	recommend := r.Group("/recommend")
	recommend.Use(opt)
	{
		recommend.GET("", discoveryController.GetRecommendPosts)
	}

	category := r.Group("/category")
	category.Use(opt)
	{
		category.GET("", discoveryController.GetCategoryPosts)
		category.GET("/tag-cloud", discoveryController.GetTagCloud)
	}

	ranking := r.Group("/ranking")
	ranking.Use(opt)
	{
		ranking.GET("", discoveryController.GetHotRanking)
	}

	topics := r.Group("/topics")
	topics.Use(opt)
	{
		topics.GET("/categories", discoveryController.GetTopicCategories)
		topics.GET("/:id/detail", discoveryController.GetTopicDetail)
		topics.GET("/:id/posts", discoveryController.GetTopicPosts)
		topics.GET("/:id/share", discoveryController.ShareTopic)
		topics.POST("/:id/participate", middleware.UserAuthMiddleware(), discoveryController.ParticipateTopic)
	}
}
