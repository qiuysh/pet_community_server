package types

// SendMessageRequest 发送私信请求体
//
// Content 对文本消息必填；图片/语音可为空，由服务端补 [图片]/[语音] 占位文案。
type SendMessageRequest struct {
	ToID uint `json:"toId" binding:"required"`
	// Content 消息文本（图片/语音可空）
	Content string `json:"content"`
	// Type 消息类型 text/image/voice，空值按 text 处理
	Type string `json:"type"`
	// MediaURL 图片/语音地址，必须来自本站上传接口
	MediaURL string `json:"mediaUrl"`
	// Duration 语音时长（秒），仅 voice 生效
	Duration int `json:"duration"`
}

// NotificationSettingsUpdateRequest 更新通知设置请求体（N8）。
//
// 六个开关与 model.NotificationSettings 一一对应，取值 0（关闭）/ 1（开启）；
// 客户端应回传完整开关集（PUT 全量语义），未传字段按 0 落库。
type NotificationSettingsUpdateRequest struct {
	LikeNotify    int `json:"likeNotify" binding:"gte=0,lte=1"`
	CommentNotify int `json:"commentNotify" binding:"gte=0,lte=1"`
	AtNotify      int `json:"atNotify" binding:"gte=0,lte=1"`
	FollowNotify  int `json:"followNotify" binding:"gte=0,lte=1"`
	SystemNotify  int `json:"systemNotify" binding:"gte=0,lte=1"`
	MessageNotify int `json:"messageNotify" binding:"gte=0,lte=1"`
}
