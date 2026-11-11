package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pet-community/config"
	"pet-community/service"
	"pet-community/utils"
)

// 通用上传（POST /api/v1/upload）的服务层契约：
// 按扩展名分流图片/视频、落到 activity-submission 分类、体积超限时不留残留文件。

func setupUploadTestStorage(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	st, err := utils.NewStorage(&config.StorageConfig{
		Driver: config.StorageDriverLocal,
		Local:  config.LocalStorageConfig{Root: root, BaseURL: "/api/v1/uploads"},
	})
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}
	prev := utils.CurrentStorage()
	utils.SetStorage(st)
	t.Cleanup(func() { utils.SetStorage(prev) })
	return root
}

func TestUploadServiceClassifiesImageAndVideo(t *testing.T) {
	setupUploadTestStorage(t)
	svc := service.NewUploadService()

	imageURL, imageKind, err := svc.UploadMedia(strings.NewReader("fakepng"), "JPG")
	if err != nil {
		t.Fatalf("upload image: %v", err)
	}
	if imageKind != service.UploadKindImage {
		t.Errorf("expected image kind, got %q", imageKind)
	}
	if !strings.HasPrefix(imageURL, "/api/v1/uploads/activity-submission/") {
		t.Errorf("image should land in activity-submission category, got %q", imageURL)
	}
	if !strings.HasSuffix(imageURL, ".jpg") {
		t.Errorf("extension should be normalized to lowercase, got %q", imageURL)
	}

	videoURL, videoKind, err := svc.UploadMedia(strings.NewReader("fakemp4"), "mp4")
	if err != nil {
		t.Fatalf("upload video: %v", err)
	}
	if videoKind != service.UploadKindVideo {
		t.Errorf("expected video kind, got %q", videoKind)
	}
	if !strings.HasPrefix(videoURL, "/api/v1/uploads/activity-submission/") {
		t.Errorf("video should land in activity-submission category, got %q", videoURL)
	}
}

func TestUploadServiceRejectsUnsupportedExtension(t *testing.T) {
	setupUploadTestStorage(t)
	svc := service.NewUploadService()

	for _, ext := range []string{"exe", "svg", "mov", "", "."} {
		if _, _, err := svc.UploadMedia(strings.NewReader("x"), ext); err == nil {
			t.Errorf("expected error for unsupported extension %q", ext)
		}
	}
}

func TestUploadServiceRejectsOversizeImage(t *testing.T) {
	root := setupUploadTestStorage(t)
	svc := service.NewUploadService()

	oversize := strings.Repeat("x", int(utils.MaxPostImageBytes)+1)
	if _, _, err := svc.UploadMedia(strings.NewReader(oversize), "png"); err == nil {
		t.Fatal("expected oversize error")
	}

	entries, err := os.ReadDir(filepath.Join(root, utils.CategorySubmission))
	if err != nil {
		t.Fatalf("read category dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("oversized object should be cleaned up, found %d entries", len(entries))
	}
}
