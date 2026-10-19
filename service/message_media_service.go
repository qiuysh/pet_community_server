package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"pet-community/utils"
)

// MaxMessageContentRunes 私信正文上限（文本 / 图 / 语音占位文案共用）
const MaxMessageContentRunes = 500

// 私信媒体落盘目录与访问前缀
const (
	MessageImageDir       = "public/uploads/messages/image"
	MessageVoiceDir       = "public/uploads/messages/voice"
	MessageImageURLPrefix = "/uploads/messages/image/"
	MessageVoiceURLPrefix = "/uploads/messages/voice/"
)

var messageImageExts = map[string]struct{}{
	"jpg": {}, "jpeg": {}, "png": {}, "gif": {}, "webp": {},
}

// 语音扩展名需覆盖各浏览器 MediaRecorder 产出：
// Chrome/Edge=webm，Firefox=ogg，Safari=mp4(.m4a)
var messageVoiceExts = map[string]struct{}{
	"webm": {}, "ogg": {}, "mp4": {}, "m4a": {},
	"aac": {}, "mp3": {}, "wav": {},
}

// UploadMessageImage 保存私信图片（本地磁盘，≤2MB），返回可访问 URL
func (s *MessageService) UploadMessageImage(file io.Reader, originalExt string) (string, error) {
	return saveMessageMedia(file, originalExt, MessageImageDir, MessageImageURLPrefix, messageImageExts,
		utils.MaxAvatarBytes, "仅支持 jpg/jpeg/png/gif/webp 格式的图片", "图片")
}

// UploadMessageVoice 保存私信语音（本地磁盘，≤10MB），返回可访问 URL
func (s *MessageService) UploadMessageVoice(file io.Reader, originalExt string) (string, error) {
	return saveMessageMedia(file, originalExt, MessageVoiceDir, MessageVoiceURLPrefix, messageVoiceExts,
		utils.MaxVoiceBytes, "仅支持 webm/ogg/mp4/m4a/aac/mp3/wav 格式的语音", "语音")
}

// saveMessageMedia 校验扩展名后落盘，返回 URL
func saveMessageMedia(
	file io.Reader,
	originalExt, dir, urlPrefix string,
	allowed map[string]struct{},
	maxSize int64,
	invalidExtMsg, kind string,
) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	if _, ok := allowed[ext]; !ok {
		return "", errors.New(invalidExtMsg)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建上传目录失败: %w", err)
	}

	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("生成随机文件名失败: %w", err)
	}
	fileName := fmt.Sprintf("%d_%s.%s", time.Now().UnixMilli(), hex.EncodeToString(randBytes), ext)
	fullPath := filepath.Join(dir, fileName)

	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, io.LimitReader(file, maxSize+1))
	if err != nil {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("保存%s数据失败: %w", kind, err)
	}
	if written > maxSize {
		_ = os.Remove(fullPath)
		return "", fmt.Errorf("%s大小不得超过 %s", kind, formatBytes(maxSize))
	}

	return urlPrefix + fileName, nil
}

// formatBytes 把字节数转成 MB 文案（本项目上限均为整 MB）
func formatBytes(size int64) string {
	return fmt.Sprintf("%dMB", size/(1024*1024))
}

// ValidateMessageMediaURL 校验私信媒体 URL：必须落在指定前缀下，且仅为单段安全文件名。
// 拒绝 ..、反斜杠、query/fragment、多级路径，避免 Static 根下的路径穿越引用。
func ValidateMessageMediaURL(mediaURL, urlPrefix string) error {
	u := strings.TrimSpace(mediaURL)
	if u == "" || len(u) > 255 {
		return errors.New("媒体地址无效")
	}
	if strings.ContainsAny(u, "?#\\\x00") || strings.Contains(u, "//") {
		return errors.New("媒体地址无效")
	}
	if !strings.HasPrefix(u, urlPrefix) {
		return errors.New("媒体地址无效")
	}
	name := u[len(urlPrefix):]
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
