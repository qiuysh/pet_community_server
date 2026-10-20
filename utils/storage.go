package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"pet-community/config"
)

// 存储分类（object key 的目录部分），业务层统一引用，避免散落魔法字符串
const (
	CategoryUserAvatar = "avatar"
	CategoryPetAvatar  = "pet-avatar"
	CategoryPetAlbum   = "pet-album"
	CategoryLostPet    = "lost-pet"
	CategoryAdopt      = "adopt"
	CategoryMsgImage   = "messages/image"
	CategoryMsgVoice   = "messages/voice"
)

// SaveOptions 写入对象时的约束
type SaveOptions struct {
	// Category 对象分类（对应本地子目录 / OSS 对象键目录），见 Category* 常量
	Category string
	// Ext 扩展名（不含点，需由调用方按白名单校验）
	Ext string
	// MaxSize 体积上限（字节），超出时清理已写内容并返回 UploadSizeError
	MaxSize int64
	// Kind 超限文案中的资源名，如 图片 / 视频 / 语音 / 头像
	Kind string
}

// Storage 统一文件存储抽象：本地磁盘与对象存储（OSS / S3 兼容）共用同一套接口，
// 业务层只拿到对外 URL，不感知底层驱动。
type Storage interface {
	// Save 写入对象并返回对外可访问 URL；超限时返回 UploadSizeError
	Save(data io.Reader, opt SaveOptions) (string, error)
	// URL 由对象键（category/name）生成对外访问 URL
	URL(key string) string
	// Delete 删除对象，对象不存在时返回 nil
	Delete(key string) error
	// Driver 返回驱动名：local / oss / s3
	Driver() string
}

var (
	storageMu     sync.RWMutex
	storageInst   Storage
	fallbackOnce  sync.Once
	fallbackStore Storage
)

// NewStorage 按配置构建存储实例（cfg 为 nil 时使用默认本地配置）
func NewStorage(cfg *config.StorageConfig) (Storage, error) {
	if cfg == nil {
		def := config.DefaultStorageConfig()
		cfg = &def
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = config.StorageDriverLocal
	}
	switch driver {
	case config.StorageDriverLocal:
		return newLocalStorage(cfg.Local)
	case config.StorageDriverOSS, config.StorageDriverS3:
		return newOSSStorage(cfg.OSS, driver)
	default:
		return nil, fmt.Errorf("unknown storage.driver %q, want local/oss/s3", cfg.Driver)
	}
}

// InitStorage 构建并设置为全局存储实例
func InitStorage(cfg *config.StorageConfig) (Storage, error) {
	st, err := NewStorage(cfg)
	if err != nil {
		return nil, err
	}
	SetStorage(st)
	return st, nil
}

// SetStorage 设置全局存储实例（测试可替换为内存 / 桩实现）
func SetStorage(st Storage) {
	storageMu.Lock()
	storageInst = st
	storageMu.Unlock()
}

// CurrentStorage 返回当前存储实例；未初始化时回退为默认本地磁盘存储
func CurrentStorage() Storage {
	storageMu.RLock()
	st := storageInst
	storageMu.RUnlock()
	if st != nil {
		return st
	}
	fallbackOnce.Do(func() {
		def := config.DefaultStorageConfig()
		fallbackStore = &localStorage{root: def.Local.Root, baseURL: def.Local.BaseURL}
	})
	return fallbackStore
}

// NewObjectName 生成不重复的对象文件名：时间戳_随机串.扩展名
func NewObjectName(ext string) (string, error) {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", errors.New("generate random object name failed")
	}
	return fmt.Sprintf("%d_%s.%s", time.Now().UnixMilli(), hex.EncodeToString(randBytes), ext), nil
}

// NewObjectKey 生成对象键（category/name），category 与 ext 均做安全校验
func NewObjectKey(category, ext string) (string, error) {
	cat, err := normalizeCategory(category)
	if err != nil {
		return "", err
	}
	e, err := normalizeExt(ext)
	if err != nil {
		return "", err
	}
	name, err := NewObjectName(e)
	if err != nil {
		return "", err
	}
	return cat + "/" + name, nil
}

// normalizeCategory 分类只允许小写字母、数字、下划线、连字符与单级目录分隔符
func normalizeCategory(category string) (string, error) {
	c := strings.Trim(category, "/")
	if c == "" {
		return "", errors.New("storage category is required")
	}
	for _, part := range strings.Split(c, "/") {
		if part == "" {
			return "", fmt.Errorf("invalid storage category %q", category)
		}
		for _, r := range part {
			isLower := r >= 'a' && r <= 'z'
			isDigit := r >= '0' && r <= '9'
			if !isLower && !isDigit && r != '-' && r != '_' {
				return "", fmt.Errorf("invalid storage category %q", category)
			}
		}
	}
	return c, nil
}

// normalizeExt 扩展名只允许小写字母与数字
func normalizeExt(ext string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(ext, ".")))
	if e == "" {
		return "", errors.New("file extension is required")
	}
	for _, r := range e {
		isLower := r >= 'a' && r <= 'z'
		isDigit := r >= '0' && r <= '9'
		if !isLower && !isDigit {
			return "", fmt.Errorf("invalid file extension %q", ext)
		}
	}
	return e, nil
}

// checkSize 校验体积并生成统一超限错误
func checkSize(written, maxSize int64, kind string) error {
	if written <= maxSize {
		return nil
	}
	return &UploadSizeError{Kind: kind, LimitLabel: FormatBytes(maxSize)}
}

// FormatBytes 把字节数转成 MB 文案（本项目上限均为整 MB）
func FormatBytes(size int64) string {
	return fmt.Sprintf("%dMB", size/(1024*1024))
}
