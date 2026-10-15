package controller

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DiscoveryController struct {
	discoveryService *service.DiscoveryService
}

func NewDiscoveryController(discoveryService *service.DiscoveryService) *DiscoveryController {
	return &DiscoveryController{discoveryService: discoveryService}
}

func (c *DiscoveryController) Search(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	keyword := ctx.Query("keyword")

	result, err := c.discoveryService.Search(keyword, userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetSearchHistory(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondSuccess(ctx, []interface{}{})
		return
	}

	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	result, err := c.discoveryService.GetSearchHistory(userID, limit)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) DeleteSearchHistory(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "ID无效")
		return
	}

	if err := c.discoveryService.DeleteSearchHistory(userID, uint(id)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *DiscoveryController) ClearSearchHistory(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if err := c.discoveryService.ClearSearchHistory(userID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *DiscoveryController) GetHotKeywords(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	result, err := c.discoveryService.GetHotKeywords(limit)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetRecommendPosts(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	postType := ctx.Query("type")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.discoveryService.GetRecommendPosts(userID, page, pageSize, postType)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetCategoryPosts(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	petType := ctx.Query("pet_type")
	postType := ctx.Query("post_type")
	sort := ctx.DefaultQuery("sort", "latest")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.discoveryService.GetCategoryPosts(userID, petType, postType, sort, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetTagCloud(ctx *gin.Context) {
	result, err := c.discoveryService.GetTagCloud()
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetHotRanking(ctx *gin.Context) {
	period := ctx.DefaultQuery("period", "day")

	result, err := c.discoveryService.GetHotRanking(period)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetTopicCategories(ctx *gin.Context) {
	result, err := c.discoveryService.GetTopicCategories()
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetTopicDetail(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	topicID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "话题ID无效")
		return
	}

	result, err := c.discoveryService.GetTopicDetail(uint(topicID), userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) GetTopicPosts(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	topicID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "话题ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.discoveryService.GetTopicPosts(uint(topicID), userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *DiscoveryController) ParticipateTopic(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, types.CodeUnauthorized, "请先登录")
		return
	}

	idStr := ctx.Param("id")
	topicID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "话题ID无效")
		return
	}

	if err := c.discoveryService.ParticipateTopic(uint(topicID), userID); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *DiscoveryController) CreateTopic(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		utils.RespondError(ctx, types.CodeUnauthorized, "请先登录")
		return
	}

	var req types.TopicCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.discoveryService.CreateTopic(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}
