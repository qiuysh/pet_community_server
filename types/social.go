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
