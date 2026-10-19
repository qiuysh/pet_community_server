package utils

import "errors"

// 上传大小上限（与 docs/specs 上传矩阵一致）
const (
	MaxAvatarBytes    int64 = 2 * 1024 * 1024 // 头像/海报/私信图片
	MaxPostImageBytes int64 = 5 * 1024 * 1024 // 动态配图/相册/领养图
	MaxVideoBytes     int64 = 50 * 1024 * 1024
	// MaxVoiceBytes 私信语音大小上限（60 秒录音远小于此值）
	MaxVoiceBytes int64 = 10 * 1024 * 1024
)

// VoiceDurationLimitSec 私信语音时长上限（秒）
const VoiceDurationLimitSec = 60

// ValidateImageSize 按场景校验图片大小。scene: avatar | post | adopt
func ValidateImageSize(size int64, scene string) error {
	limit := MaxPostImageBytes
	label := "5MB"
	switch scene {
	case "avatar", "poster", "message":
		limit = MaxAvatarBytes
		label = "2MB"
	case "post", "album", "adopt":
		limit = MaxPostImageBytes
		label = "5MB"
	}
	if size > limit {
		return &UploadSizeError{LimitLabel: label}
	}
	return nil
}

type UploadSizeError struct {
	Kind       string // 图片 | 视频
	LimitLabel string
}

func (e *UploadSizeError) Error() string {
	kind := e.Kind
	if kind == "" {
		kind = "图片"
	}
	return kind + "大小不得超过 " + e.LimitLabel
}

// ValidateVideoSize 校验视频大小（相册/动态统一 ≤50MB）
func ValidateVideoSize(size int64) error {
	if size > MaxVideoBytes {
		return &UploadSizeError{Kind: "视频", LimitLabel: "50MB"}
	}
	return nil
}

// ValidateVoiceSize 校验私信语音大小（≤10MB）
func ValidateVoiceSize(size int64) error {
	if size > MaxVoiceBytes {
		return &UploadSizeError{Kind: "语音", LimitLabel: "10MB"}
	}
	return nil
}

// ValidateVoiceDuration 校验语音时长（1-60 秒）
func ValidateVoiceDuration(seconds int) error {
	if seconds < 1 || seconds > VoiceDurationLimitSec {
		return errors.New("语音时长需在1-60秒之间")
	}
	return nil
}

// VideoDurationLimitSec 相册视频时长上限（秒）
const VideoDurationLimitSec = 60
