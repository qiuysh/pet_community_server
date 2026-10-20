package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pet-community/config"
)

// localStorage 本地磁盘驱动：落盘到 root 目录，通过 baseURL 前缀对外提供静态访问
type localStorage struct {
	root    string
	baseURL string
}

func newLocalStorage(cfg config.LocalStorageConfig) (*localStorage, error) {
	root := cfg.Root
	if root == "" {
		root = "public/uploads"
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "/api/v1/uploads"
	}
	return &localStorage{root: root, baseURL: baseURL}, nil
}

// Save 写入本地磁盘，返回对外访问 URL
func (s *localStorage) Save(data io.Reader, opt SaveOptions) (string, error) {
	key, err := NewObjectKey(opt.Category, opt.Ext)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(s.root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create upload dir: %w", err)
	}

	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	// 先关闭句柄再清理残留文件：Windows 下无法删除仍被打开的文件
	written, copyErr := io.Copy(f, io.LimitReader(data, opt.MaxSize+1))
	closeErr := f.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(fullPath)
		if copyErr != nil {
			return "", fmt.Errorf("save file: %w", copyErr)
		}
		return "", fmt.Errorf("save file: %w", closeErr)
	}
	if err := checkSize(written, opt.MaxSize, opt.Kind); err != nil {
		_ = os.Remove(fullPath)
		return "", err
	}

	return s.URL(key), nil
}

// URL 生成本地静态访问地址
func (s *localStorage) URL(key string) string {
	return s.baseURL + "/" + strings.TrimPrefix(key, "/")
}

// Delete 删除本地文件，不存在时返回 nil；拒绝越出 root 的 key
func (s *localStorage) Delete(key string) error {
	key = strings.TrimPrefix(filepath.ToSlash(key), "/")
	if key == "" || strings.Contains(key, "..") {
		return fmt.Errorf("invalid object key")
	}
	fullPath := filepath.Join(s.root, filepath.FromSlash(key))
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return fmt.Errorf("resolve storage root: %w", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fmt.Errorf("resolve object path: %w", err)
	}
	sep := string(os.PathSeparator)
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+sep) {
		return fmt.Errorf("invalid object key")
	}
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// Driver 返回驱动名
func (s *localStorage) Driver() string { return config.StorageDriverLocal }
