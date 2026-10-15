package controller

import (
	"pet-community/repository"
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	notificationService *service.NotificationService
	notificationRepo    *repository.NotificationRepository
}

func NewNotificationController(
	notificationService *service.NotificationService,
	notificationRepo *repository.NotificationRepository,
) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
		notificationRepo:    notificationRepo,
	}
}

func (c *NotificationController) ListNotifications(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	notifications, total, err := c.notificationRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	list := make([]types.NotificationInfo, 0, len(notifications))
	for _, n := range notifications {
		list = append(list, types.NotificationInfo{
			ID:            n.ID,
			Type:          n.Type,
			Title:         n.Title,
			Content:       n.Content,
			TriggerType:   n.TriggerType,
			TriggerID:     n.TriggerID,
			TriggerUserID: n.TriggerUserID,
			IsRead:        n.IsRead,
			CreatedAt:     n.CreatedAt.Format(time.RFC3339),
		})
	}

	utils.RespondSuccess(ctx, types.NotificationListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (c *NotificationController) GetUnreadCount(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	count, err := c.notificationRepo.GetUnreadCount(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.UnreadCountResponse{Count: count})
}

func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	notificationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "通知ID无效")
		return
	}

	if err := c.notificationRepo.MarkAsRead(userID, uint(notificationID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	if err := c.notificationRepo.MarkAllAsRead(userID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *NotificationController) DeleteNotification(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	idStr := ctx.Param("id")
	notificationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "通知ID无效")
		return
	}

	if err := c.notificationRepo.Delete(userID, uint(notificationID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

type FollowController struct {
	followService *service.FollowService
}

func NewFollowController(followService *service.FollowService) *FollowController {
	return &FollowController{followService: followService}
}

func (c *FollowController) ToggleFollow(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	var req struct {
		IsSecret bool `json:"isSecret"`
	}
	ctx.ShouldBindJSON(&req)

	isFollowing, err := c.followService.ToggleFollow(userID, uint(targetID), req.IsSecret)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, gin.H{"isFollowing": isFollowing})
}

func (c *FollowController) Unfollow(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	if err := c.followService.Unfollow(userID, uint(targetID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, gin.H{"isFollowing": false})
}

func (c *FollowController) SetSecretFollow(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	var req struct {
		IsSecret bool `json:"isSecret" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	if err := c.followService.SetSecretFollow(userID, uint(targetID), req.IsSecret); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *FollowController) GetSecretFollowing(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	users, total, err := c.followService.GetSecretFollowing(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	list := make([]types.FollowInfo, 0, len(users))
	for _, u := range users {
		list = append(list, types.FollowInfo{
			ID:         u.ID,
			UserID:     u.ID,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Signature:  u.Signature,
			IsFollowing: true,
			IsSecret:   true, // 悄悄关注列表项均为悄悄关注（F6-02）
		})
	}

	utils.RespondSuccess(ctx, types.FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (c *FollowController) GetOnlineStatus(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	status, err := c.followService.GetOnlineStatus(uint(userID))
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, gin.H{"onlineStatus": status})
}

func (c *FollowController) GetFollowers(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	users, total, err := c.followService.GetFollowers(uint(userID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	list := make([]types.FollowInfo, 0, len(users))
	viewerID := ctx.GetUint("user_id")
	for _, u := range users {
		isFollowing := false
		if viewerID > 0 {
			isFollowing, _ = c.followService.IsFollowing(viewerID, u.ID)
		}
		list = append(list, types.FollowInfo{
			ID:          u.ID,
			UserID:      u.ID,
			Nickname:    u.Nickname,
			Avatar:      u.Avatar,
			Signature:   u.Signature,
			IsFollowing: isFollowing,
		})
	}

	utils.RespondSuccess(ctx, types.FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (c *FollowController) GetFollowing(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	users, total, err := c.followService.GetFollowing(uint(userID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	list := make([]types.FollowInfo, 0, len(users))
	viewerID := ctx.GetUint("user_id")
	// 仅查看自己关注列表时暴露悄悄关注标记（F6-04），他人视角不泄露
	revealSecret := viewerID > 0 && viewerID == uint(userID)
	for _, u := range users {
		isFollowing := true
		if viewerID > 0 && viewerID != uint(userID) {
			isFollowing, _ = c.followService.IsFollowing(viewerID, u.ID)
		}
		info := types.FollowInfo{
			ID:          u.ID,
			UserID:      u.ID,
			Nickname:    u.Nickname,
			Avatar:      u.Avatar,
			Signature:   u.Signature,
			IsFollowing: isFollowing,
		}
		if revealSecret {
			info.IsSecret = c.followService.IsSecretFollow(uint(userID), u.ID)
		}
		list = append(list, info)
	}

	utils.RespondSuccess(ctx, types.FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (c *FollowController) GetFollowStatus(ctx *gin.Context) {
	viewerID := ctx.GetUint("user_id")
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	isFollowing, isSecret, isMutual, err := c.followService.GetFollowState(viewerID, uint(targetID))
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	followersCount, _ := c.followService.GetFollowersCount(uint(targetID))
	followingCount, _ := c.followService.GetFollowingCount(uint(targetID))

	utils.RespondSuccess(ctx, gin.H{
		"isFollowing":    isFollowing,
		"isSecret":        isSecret,
		"isMutual":        isMutual,
		"followersCount":  followersCount,
		"followingCount":  followingCount,
	})
}
