package controller

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postService    *service.PostService
	draftService   *service.DraftService
	topicService   *service.TopicService
	commentService *service.CommentService
	reportService  *service.ReportService
}

func NewPostController(
	postService *service.PostService,
	draftService *service.DraftService,
	topicService *service.TopicService,
	commentService *service.CommentService,
	reportService *service.ReportService,
) *PostController {
	return &PostController{
		postService:    postService,
		draftService:   draftService,
		topicService:   topicService,
		commentService: commentService,
		reportService:  reportService,
	}
}

func (c *PostController) CreatePost(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.PostCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	post, err := c.postService.CreatePost(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, post)
}

func (c *PostController) UpdatePost(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	var req types.PostUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	post, err := c.postService.UpdatePost(userID, uint(postID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, post)
}

func (c *PostController) CancelSchedule(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	draft, err := c.postService.CancelScheduleToDraft(userID, uint(postID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, draft)
}

func (c *PostController) DeletePost(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	if err := c.postService.DeletePost(userID, uint(postID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PostController) GetPost(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	post, err := c.postService.GetPostByID(userID, uint(postID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, post)
}

func (c *PostController) ListPosts(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	topicIDStr := ctx.Query("topic_id")
	petIDStr := ctx.Query("pet_id")
	postType := ctx.Query("type")
	keyword := ctx.Query("keyword")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	var topicID, petID uint
	if topicIDStr != "" {
		if id, err := strconv.ParseUint(topicIDStr, 10, 32); err == nil {
			topicID = uint(id)
		}
	}
	if petIDStr != "" {
		if id, err := strconv.ParseUint(petIDStr, 10, 32); err == nil {
			petID = uint(id)
		}
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.postService.ListPosts(userID, topicID, petID, postType, keyword, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) ToggleLike(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	result, err := c.postService.ToggleLike(userID, uint(postID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) ToggleFavorite(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	result, err := c.postService.ToggleFavorite(userID, uint(postID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) RecordShare(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	var req types.ShareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.postService.RecordShare(userID, uint(postID), req.Platform); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PostController) GetLikeList(ctx *gin.Context) {
	idStr := ctx.Param("id")
	postID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.postService.GetLikeList(uint(postID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) GetFavorites(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.postService.GetFavorites(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) CreateDraft(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.DraftCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	draft, err := c.draftService.CreateDraft(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, draft)
}

func (c *PostController) UpdateDraft(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	draftID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "草稿ID无效")
		return
	}

	var req types.DraftCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	draft, err := c.draftService.UpdateDraft(userID, uint(draftID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, draft)
}

func (c *PostController) GetDraft(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	draftID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "草稿ID无效")
		return
	}

	draft, err := c.draftService.GetDraft(userID, uint(draftID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, draft)
}

func (c *PostController) ListDrafts(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.draftService.ListDrafts(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) DeleteDraft(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	draftID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "草稿ID无效")
		return
	}

	if err := c.draftService.DeleteDraft(userID, uint(draftID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PostController) ListTopics(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.topicService.ListTopics(keyword, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) GetTopic(ctx *gin.Context) {
	idStr := ctx.Param("id")
	topicID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "话题ID无效")
		return
	}

	topic, err := c.topicService.GetTopic(uint(topicID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, topic)
}

func (c *PostController) SearchTopics(ctx *gin.Context) {
	keyword := ctx.Query("keyword")
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	result, err := c.topicService.SearchTopics(keyword, limit)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) CreateTopic(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	var req types.TopicCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	topic, err := c.topicService.CreateTopic(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, topic)
}

func (c *PostController) CreateComment(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.CommentCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	comment, err := c.commentService.CreateComment(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, comment)
}

func (c *PostController) UpdateComment(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	commentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "评论ID无效")
		return
	}

	var req types.CommentUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	comment, err := c.commentService.UpdateComment(userID, uint(commentID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, comment)
}

func (c *PostController) DeleteComment(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	commentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "评论ID无效")
		return
	}

	if err := c.commentService.DeleteComment(userID, uint(commentID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *PostController) ListComments(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	// 路由为 GET /posts/:id/comments，优先取路径参数；兼容旧版 ?post_id=
	postIDStr := ctx.Param("id")
	if postIDStr == "" {
		postIDStr = ctx.Query("post_id")
	}
	sort := ctx.DefaultQuery("sort", "latest")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.Query("page_size")
	if pageSizeStr == "" {
		pageSizeStr = ctx.DefaultQuery("pageSize", "20")
	}

	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "动态ID无效")
		return
	}

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.commentService.ListComments(userID, uint(postID), sort, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) ToggleCommentLike(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	commentID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "评论ID无效")
		return
	}

	result, err := c.commentService.ToggleLike(userID, uint(commentID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

func (c *PostController) CreateReport(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.ReportCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	report, err := c.reportService.CreateReport(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, report)
}

func (c *PostController) GetReport(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	reportID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "举报ID无效")
		return
	}

	report, err := c.reportService.GetReport(userID, uint(reportID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, report)
}

func (c *PostController) ListReports(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.reportService.ListReports(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}
