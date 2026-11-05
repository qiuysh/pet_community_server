package service

import (
	"errors"
	"io"
	"strings"

	"pet-community/utils"
)

// 通用上传的媒体类型（与 types.UploadMediaResponse.Type 对应）
const (
	UploadKindImage = "image"
	UploadKindVideo = "video"
)

// uploadImageExts 通用上传允许的图片扩展名
var uploadImageExts = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"gif":  true,
	"webp": true,
	"bmp":  true,
}

// uploadVideoExts 通用上传允许的视频扩展名（与宠物相册一致，仅 MP4）
var uploadVideoExts = map[string]bool{
	"mp4": true,
}

// UploadService 通用上传服务。
// 为没有专属上传端点的场景（目前是活动作品提交）提供落盘能力：
// 统一走 utils.Storage，业务层只拿到对外 URL，不感知本地磁盘还是 OSS。
type UploadService struct{}

// NewUploadService 构造通用上传服务
func NewUploadService() *UploadService {
	return &UploadService{}
}

// UploadMedia 按扩展名判定图片/视频并保存，返回对外 URL 与媒体类型。
// 体积上限由 SaveOptions.MaxSize 在 Storage 层校验（图片 5MB / 视频 50MB）。
func (s *UploadService) UploadMedia(file io.Reader, originalExt string) (string, string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))

	var opt utils.SaveOptions
	var kind string
	switch {
	case uploadImageExts[ext]:
		kind = UploadKindImage
		opt = utils.SaveOptions{
			Category: utils.CategorySubmission,
			Ext:      ext,
			MaxSize:  utils.MaxPostImageBytes,
			Kind:     "图片",
		}
	case uploadVideoExts[ext]:
		kind = UploadKindVideo
		opt = utils.SaveOptions{
			Category: utils.CategorySubmission,
			Ext:      ext,
			MaxSize:  utils.MaxVideoBytes,
			Kind:     "视频",
		}
	default:
		return "", "", errors.New("仅支持 jpg/jpeg/png/gif/webp/bmp 图片或 MP4 视频")
	}

	url, err := utils.CurrentStorage().Save(file, opt)
	if err != nil {
		return "", "", err
	}
	return url, kind, nil
}
