package service

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"pet-community/utils"
)

// MaxMessageContentRunes 私信正文上限（文本 / 图 / 语音占位文案共用）
const MaxMessageContentRunes = 500

// 私信媒体存储分类（对应存储层的 category，本地落盘目录与 OSS object key 共用）
const (
	// MessageImageCategory 私信图片分类
	MessageImageCategory = utils.CategoryMsgImage
	// MessageVoiceCategory 私信语音分类
	MessageVoiceCategory = utils.CategoryMsgVoice
)

// MessageImageURLPrefix 私信图片访问前缀，随存储配置变化：
// 本地驱动为配置的 URL 前缀（默认 /api/v1/uploads/messages/image/），OSS 驱动为对象存储域名。
func MessageImageURLPrefix() string {
	return utils.CurrentStorage().URL(MessageImageCategory) + "/"
}

// MessageVoiceURLPrefix 私信语音访问前缀，规则同 MessageImageURLPrefix
func MessageVoiceURLPrefix() string {
	return utils.CurrentStorage().URL(MessageVoiceCategory) + "/"
}

var messageImageExts = map[string]struct{}{
	"jpg": {}, "jpeg": {}, "png": {}, "gif": {}, "webp": {},
}

// 语音扩展名需覆盖各浏览器 MediaRecorder 产出：
// Chrome/Edge=webm，Firefox=ogg，Safari=mp4(.m4a)
var messageVoiceExts = map[string]struct{}{
	"webm": {}, "ogg": {}, "mp4": {}, "m4a": {},
	"aac": {}, "mp3": {}, "wav": {},
}

// UploadMessageImage 保存私信图片（≤2MB），返回可访问 URL。
// 实际落盘位置由存储驱动决定：本地磁盘或对象存储（OSS）。
func (s *MessageService) UploadMessageImage(file io.Reader, originalExt string) (string, error) {
	return saveMessageMedia(file, originalExt, MessageImageCategory, messageImageExts,
		utils.MaxAvatarBytes, "仅支持 jpg/jpeg/png/gif/webp 格式的图片", "图片")
}

// UploadMessageVoice 保存私信语音（≤10MB），返回可访问 URL。
// 语音与图片统一走存储抽象层，配置 storage.driver=oss 后即落到对象存储。
func (s *MessageService) UploadMessageVoice(file io.Reader, originalExt string) (string, error) {
	return saveMessageMedia(file, originalExt, MessageVoiceCategory, messageVoiceExts,
		utils.MaxVoiceBytes, "仅支持 webm/ogg/mp4/m4a/aac/mp3/wav 格式的语音", "语音")
}

// saveMessageMedia 校验扩展名后交给存储层保存，返回对外 URL
func saveMessageMedia(
	file io.Reader,
	originalExt, category string,
	allowed map[string]struct{},
	maxSize int64,
	invalidExtMsg, kind string,
) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	if _, ok := allowed[ext]; !ok {
		return "", errors.New(invalidExtMsg)
	}

	return utils.CurrentStorage().Save(file, utils.SaveOptions{
		Category: category,
		Ext:      ext,
		MaxSize:  maxSize,
		Kind:     kind,
	})
}

// ValidateMessageMediaURL 校验私信媒体 URL：必须落在当前存储前缀下，且仅为单段安全文件名。
// 拒绝 ..、反斜杠、query/fragment、多级路径，避免 Static 根下的路径穿越引用。
func ValidateMessageMediaURL(mediaURL, category string) error {
	u := strings.TrimSpace(mediaURL)
	if u == "" || len(u) > 255 {
		return errors.New("媒体地址无效")
	}
	if strings.ContainsAny(u, "?#\\\x00") {
		return errors.New("媒体地址无效")
	}
	// 对象存储为绝对地址（https://cdn/...），路径部分的 // 才代表异常
	pathProbe := u
	if i := strings.Index(pathProbe, "://"); i >= 0 {
		pathProbe = pathProbe[i+3:]
	}
	if strings.Contains(pathProbe, "//") {
		return errors.New("媒体地址无效")
	}

	// 前缀随存储配置变化：本地为配置的 URL 前缀，OSS 为对象存储域名
	prefix := utils.CurrentStorage().URL(category) + "/"
	if !strings.HasPrefix(u, prefix) {
		return errors.New("媒体地址无效")
	}
	name := u[len(prefix):]
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		return errors.New("媒体地址无效")
	}
	if filepath.Base(name) != name {
		return errors.New("媒体地址无效")
	}
	return nil
}

// ValidateMessageContentLength 校验私信正文长度（字符数）
func ValidateMessageContentLength(content string) error {
	if utf8.RuneCountInString(content) > MaxMessageContentRunes {
		return errors.New("消息内容不能超过500字")
	}
	return nil
}
