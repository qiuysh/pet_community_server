package controller

import (
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"pet-community/model"
	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"

	"github.com/gin-gonic/gin"
)

// 私信文本长度上限（字符数）
const maxMessageTextRunes = 500

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

	var req types.SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondError(ctx, types.CodeParamError, "参数无效")
		return
	}

	msgType := strings.ToLower(strings.TrimSpace(req.Type))
	if msgType == "" {
		msgType = model.MessageTypeText
	}
	content := strings.TrimSpace(req.Content)
	mediaURL := strings.TrimSpace(req.MediaURL)

	switch msgType {
	case model.MessageTypeText:
		if content == "" {
			utils.RespondError(ctx, types.CodeParamError, "消息内容不能为空")
			return
		}
		mediaURL = ""
		req.Duration = 0

	case model.MessageTypeImage:
		if err := service.ValidateMessageMediaURL(mediaURL, service.MessageImageCategory); err != nil {
			utils.RespondError(ctx, types.CodeParamError, "图片地址无效")
			return
		}
		req.Duration = 0
		if content == "" {
			content = "[图片]"
		}

	case model.MessageTypeVoice:
		if err := service.ValidateMessageMediaURL(mediaURL, service.MessageVoiceCategory); err != nil {
			utils.RespondError(ctx, types.CodeParamError, "语音地址无效")
			return
		}
		if err := utils.ValidateVoiceDuration(req.Duration); err != nil {
			utils.RespondError(ctx, types.CodeParamError, err.Error())
			return
		}
		if content == "" {
			content = "[语音]"
		}

	default:
		utils.RespondError(ctx, types.CodeParamError, "不支持的消息类型")
		return
	}

	if utf8.RuneCountInString(content) > maxMessageTextRunes {
		utils.RespondError(ctx, types.CodeParamError, "消息内容不能超过500字")
		return
	}

	message, err := c.messageService.SendMessageWithMedia(userID, req.ToID, content, msgType, mediaURL, req.Duration)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, message)
}

// UploadMessageImage 上传私信图片（multipart field=file，≤2MB）
func (c *MessageController) UploadMessageImage(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "未收到上传文件")
		return
	}
	if err := utils.ValidateImageSize(fileHeader.Size, "message"); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "无法读取上传文件")
		return
	}
	defer f.Close()

	url, err := c.messageService.UploadMessageImage(f, filepath.Ext(fileHeader.Filename))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, &types.UploadPhotoResponse{Url: url})
}

// UploadMessageVoice 上传私信语音（multipart field=file，≤10MB）
func (c *MessageController) UploadMessageVoice(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "未收到上传文件")
		return
	}
	if err := utils.ValidateVoiceSize(fileHeader.Size); err != nil {
		utils.RespondError(ctx, types.CodeParamError, err.Error())
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "无法读取上传文件")
		return
	}
	defer f.Close()

	url, err := c.messageService.UploadMessageVoice(f, filepath.Ext(fileHeader.Filename))
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}
	utils.RespondSuccess(ctx, &types.UploadPhotoResponse{Url: url})
}

// parsePageParams 解析分页参数：优先 pageSize，回退 page_size，范围 1-100
func parsePageParams(ctx *gin.Context) (int, int) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	sizeStr := ctx.Query("pageSize")
	if sizeStr == "" {
		sizeStr = ctx.DefaultQuery("page_size", "20")
	}
	pageSize, _ := strconv.Atoi(sizeStr)
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func (c *MessageController) GetConversations(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	page, pageSize := parsePageParams(ctx)

	conversations, total, err := c.messageService.GetConversations(userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeSystemError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.PageResponse(conversations, total, page, pageSize))
}

func (c *MessageController) GetMessages(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	conversationIDStr := ctx.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "对话ID无效")
		return
	}

	page, pageSize := parsePageParams(ctx)

	messages, total, err := c.messageService.GetMessages(uint(conversationID), userID, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
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
	userID := ctx.GetUint("user_id")
	conversationIDStr := ctx.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "对话ID无效")
		return
	}

	keyword := strings.TrimSpace(ctx.Query("keyword"))
	if keyword == "" {
		utils.RespondError(ctx, types.CodeParamError, "关键词不能为空")
		return
	}
	if utf8.RuneCountInString(keyword) > 50 {
		utils.RespondError(ctx, types.CodeParamError, "关键词不能超过50字")
		return
	}

	page, pageSize := parsePageParams(ctx)
	messages, total, err := c.messageService.SearchMessages(uint(conversationID), userID, keyword, page, pageSize)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, types.PageResponse(messages, total, page, pageSize))
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
