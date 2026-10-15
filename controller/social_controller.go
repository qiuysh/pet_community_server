package controller

import (
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BlockController struct {
	blockService *service.BlockService
}

func NewBlockController(blockService *service.BlockService) *BlockController {
	return &BlockController{blockService: blockService}
}

func (c *BlockController) BlockUser(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	if err := c.blockService.BlockUser(userID, uint(targetID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *BlockController) UnblockUser(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	targetIDStr := ctx.Param("id")
	targetID, err := strconv.ParseUint(targetIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "用户ID无效")
		return
	}

	if err := c.blockService.UnblockUser(userID, uint(targetID)); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *BlockController) GetBlockedList(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	users, total, err := c.blockService.GetBlockedList(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	list := make([]types.FollowInfo, 0, len(users))
	for _, u := range users {
		list = append(list, types.FollowInfo{
			ID:        u.ID,
			UserID:    u.ID,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
		})
	}

	utils.RespondSuccess(ctx, types.FollowListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

type MessageController struct {
	messageService *service.MessageService
	userService    *service.UserService
}

func NewMessageController(messageService *service.MessageService, userService *service.UserService) *MessageController {
	return &MessageController{messageService: messageService, userService: userService}
}

func (c *MessageController) SendMessage(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req struct {
		ToID    uint   `json:"toId" binding:"required"`
		Content string `json:"content" binding:"required"`
		Type    string `json:"type" binding:"omitempty"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, "参数无效")
		return
	}

	if req.Type == "" {
		req.Type = "text"
	}

	message, err := c.messageService.SendMessage(userID, req.ToID, req.Content, req.Type)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, message)
}

func (c *MessageController) GetConversations(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	conversations, total, err := c.messageService.GetConversations(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.PageResponse(conversations, total, page, pageSize))
}

func (c *MessageController) GetMessages(ctx *gin.Context) {
	conversationIDStr := ctx.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "对话ID无效")
		return
	}

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	messages, total, err := c.messageService.GetMessages(uint(conversationID), page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.PageResponse(messages, total, page, pageSize))
}

func (c *MessageController) MarkAsRead(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	conversationIDStr := ctx.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "对话ID无效")
		return
	}

	if err := c.messageService.MarkAsRead(uint(conversationID), userID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *MessageController) RecallMessage(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	messageIDStr := ctx.Param("msgId")
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "消息ID无效")
		return
	}

	if err := c.messageService.RecallMessage(uint(messageID), userID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *MessageController) SearchMessages(ctx *gin.Context) {
	conversationIDStr := ctx.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "对话ID无效")
		return
	}

	keyword := ctx.Query("keyword")
	if keyword == "" {
		utils.RespondError(ctx, types.CodeParamError, "关键词不能为空")
		return
	}

	messages, err := c.messageService.SearchMessages(uint(conversationID), keyword)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, messages)
}

func (c *MessageController) DeleteConversation(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	conversationIDStr := ctx.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "对话ID无效")
		return
	}

	if err := c.messageService.DeleteConversation(uint(conversationID), userID); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}

func (c *MessageController) GetUnreadCount(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	count, err := c.messageService.GetUnreadCount(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.UnreadCountResponse{Count: count})
}

type PointsController struct {
	pointsService *service.PointsService
}

func NewPointsController(pointsService *service.PointsService) *PointsController {
	return &PointsController{pointsService: pointsService}
}

func (c *PointsController) GetUserLevel(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	level, err := c.pointsService.GetUserLevel(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, level)
}

func (c *PointsController) GetPointsRecords(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	records, total, err := c.pointsService.GetPointsRecords(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.PageResponse(records, total, page, pageSize))
}

func (c *PointsController) GetPointsRules(ctx *gin.Context) {
	rules, err := c.pointsService.GetPointsRules()
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, rules)
}

func (c *PointsController) GetBenefits(ctx *gin.Context) {
	benefits, err := c.pointsService.GetAllBenefits()
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, benefits)
}

func (c *PointsController) GetBenefitsByLevel(ctx *gin.Context) {
	levelStr := ctx.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "等级无效")
		return
	}

	benefits, err := c.pointsService.GetBenefits(level)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, benefits)
}

type NotificationSettingsController struct {
	settingsService *service.NotificationSettingsService
}

func NewNotificationSettingsController(settingsService *service.NotificationSettingsService) *NotificationSettingsController {
	return &NotificationSettingsController{settingsService: settingsService}
}

func (c *NotificationSettingsController) GetSettings(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	settings, err := c.settingsService.GetSettings(userID)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, settings)
}

func (c *NotificationSettingsController) UpdateSettings(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req struct {
		LikeNotify    int `json:"likeNotify"`
		CommentNotify int `json:"commentNotify"`
		AtNotify      int `json:"atNotify"`
		FollowNotify  int `json:"followNotify"`
		SystemNotify  int `json:"systemNotify"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, "参数无效")
		return
	}

	if err := c.settingsService.UpdateSettings(userID, req.LikeNotify, req.CommentNotify, req.AtNotify, req.FollowNotify, req.SystemNotify); err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, nil)
}
