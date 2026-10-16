package controller

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ActivityController struct {
	activityService     *service.ActivityService
	registrationService *service.RegistrationService
	submissionService   *service.SubmissionService
	voteService         *service.VoteService
	signInService       *service.SignInService
	albumService        *service.AlbumService
	reviewService       *service.ActivityReviewService
}

func NewActivityController(
	activityService *service.ActivityService,
	registrationService *service.RegistrationService,
	submissionService *service.SubmissionService,
	voteService *service.VoteService,
	signInService *service.SignInService,
	albumService *service.AlbumService,
	reviewService *service.ActivityReviewService,
) *ActivityController {
	return &ActivityController{
		activityService:     activityService,
		registrationService: registrationService,
		submissionService:   submissionService,
		voteService:         voteService,
		signInService:       signInService,
		albumService:        albumService,
		reviewService:       reviewService,
	}
}

// CreateActivity 创建活动
func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	var req types.ActivityCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	activity, err := c.activityService.CreateActivity(&req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, activity)
}

// UpdateActivity 更新活动
func (c *ActivityController) UpdateActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	var req types.ActivityUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	activity, err := c.activityService.UpdateActivity(uint(activityID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, activity)
}

// DeleteActivity 删除活动
func (c *ActivityController) DeleteActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	if err := c.activityService.DeleteActivity(uint(activityID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// GetActivity 获取活动详情
func (c *ActivityController) GetActivity(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	activity, err := c.activityService.GetActivityByID(userID, uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, activity)
}

// ListActivities 活动列表（前台）
func (c *ActivityController) ListActivities(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	activityType := ctx.Query("type")
	status := ctx.Query("status")
	keyword := ctx.Query("keyword")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.activityService.ListActivities(userID, activityType, status, keyword, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ListAdminActivities 管理端活动列表（含草稿）
func (c *ActivityController) ListAdminActivities(ctx *gin.Context) {
	activityType := ctx.Query("type")
	status := ctx.Query("status")
	keyword := ctx.Query("keyword")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.activityService.ListActivitiesForAdmin(activityType, status, keyword, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// PublishActivity 发布活动
func (c *ActivityController) PublishActivity(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	if err := c.activityService.PublishActivity(uint(activityID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// Register 活动报名
func (c *ActivityController) Register(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.RegistrationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	registration, err := c.registrationService.Register(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, registration)
}

// CancelRegistration 取消报名
func (c *ActivityController) CancelRegistration(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	if err := c.registrationService.CancelRegistration(userID, uint(activityID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// ListRegistrations 报名列表
func (c *ActivityController) ListRegistrations(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	statusStr := ctx.Query("status")
	status := -1
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.registrationService.ListRegistrations(uint(activityID), status, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ApproveRegistration 审核报名
func (c *ActivityController) ApproveRegistration(ctx *gin.Context) {
	idStr := ctx.Param("id")
	registrationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "报名ID无效")
		return
	}

	approveStr := ctx.Query("approve")
	approve := approveStr == "true"

	if err := c.registrationService.ApproveRegistration(uint(registrationID), approve); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// CreateSubmission 创建作品
func (c *ActivityController) CreateSubmission(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.SubmissionCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	submission, err := c.submissionService.CreateSubmission(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, submission)
}

// UpdateSubmission 更新作品
func (c *ActivityController) UpdateSubmission(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "作品ID无效")
		return
	}

	var req types.SubmissionUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	submission, err := c.submissionService.UpdateSubmission(userID, uint(submissionID), &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, submission)
}

// DeleteSubmission 删除作品
func (c *ActivityController) DeleteSubmission(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "作品ID无效")
		return
	}

	if err := c.submissionService.DeleteSubmission(userID, uint(submissionID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// ListSubmissions 作品列表
func (c *ActivityController) ListSubmissions(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	sort := ctx.DefaultQuery("sort", "latest")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.submissionService.ListSubmissions(userID, uint(activityID), sort, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// GetSubmission 获取作品详情
func (c *ActivityController) GetSubmission(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "作品ID无效")
		return
	}

	submission, err := c.submissionService.GetSubmissionByID(userID, uint(submissionID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, submission)
}

// AwardSubmission 评选作品
func (c *ActivityController) AwardSubmission(ctx *gin.Context) {
	var req types.AwardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.submissionService.AwardSubmission(req.SubmissionID, req.AwardLevel); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// GetAwardPage 获取评选页面数据
func (c *ActivityController) GetAwardPage(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	result, err := c.submissionService.GetAwardPage(userID, uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// BatchAwardSubmission 批量设置获奖
func (c *ActivityController) BatchAwardSubmission(ctx *gin.Context) {
	var req types.BatchAwardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.submissionService.BatchAwardSubmission(req.Awards); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// PublishAwardResult 发布评选结果
func (c *ActivityController) PublishAwardResult(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	result, err := c.submissionService.PublishAwardResult(uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// Vote 为作品投票
func (c *ActivityController) Vote(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.VoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.voteService.Vote(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// CancelVote 取消投票
func (c *ActivityController) CancelVote(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	submissionID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "作品ID无效")
		return
	}

	if err := c.voteService.CancelVote(userID, uint(submissionID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// GetVoteHistory 获取投票历史
func (c *ActivityController) GetVoteHistory(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.voteService.GetVoteHistory(userID, uint(activityID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// GetUserVoteStats 获取用户投票统计
func (c *ActivityController) GetUserVoteStats(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	result, err := c.voteService.GetUserVoteStats(userID, uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ========== 签到相关接口 ==========

// GenerateQRCode 生成签到二维码
func (c *ActivityController) GenerateQRCode(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	result, err := c.signInService.GenerateQRCode(uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// SignIn 扫码签到
func (c *ActivityController) SignIn(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.SignInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.signInService.SignIn(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ListSignIns 签到列表
func (c *ActivityController) ListSignIns(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.signInService.ListSignIns(uint(activityID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// GetSignInStats 签到统计
func (c *ActivityController) GetSignInStats(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	signInCount, registrationCount, err := c.signInService.GetSignInStats(uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, map[string]int64{
		"signInCount":       signInCount,
		"registrationCount": registrationCount,
	})
}

// ========== 相册相关接口 ==========

// UploadPhoto 上传照片到相册
func (c *ActivityController) UploadPhoto(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.AlbumPhotoCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.albumService.UploadPhoto(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// DeletePhoto 删除照片
func (c *ActivityController) DeletePhoto(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	photoID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "照片ID无效")
		return
	}

	if err := c.albumService.DeletePhoto(userID, uint(photoID)); err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

// ListPhotos 照片列表
func (c *ActivityController) ListPhotos(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	statusStr := ctx.Query("status")
	status := -1
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, err := c.albumService.ListPhotos(uint(activityID), status, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// ========== 活动回顾相关接口 ==========

// PublishReview 发布活动回顾
func (c *ActivityController) PublishReview(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req types.ActivityReviewCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	result, err := c.reviewService.PublishReview(userID, &req)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}

// GetReview 获取活动回顾
func (c *ActivityController) GetReview(ctx *gin.Context) {
	idStr := ctx.Param("id")
	activityID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "活动ID无效")
		return
	}

	result, err := c.reviewService.GetReview(uint(activityID))
	if err != nil {
		utils.RespondError(ctx, types.CodeNotFound, err.Error())
		return
	}

	utils.RespondSuccess(ctx, result)
}
