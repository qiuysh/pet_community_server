package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pet-community/config"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/utils"

	"gopkg.in/yaml.v3"
)

// 反馈回路：M4-04 要求语音/图片等上传走对象存储（OSS），且访问地址带可配置前缀。
// 下面三组断言分别覆盖：前缀可配、OSS 驱动真的把对象 PUT 到 S3 兼容端点、业务层不再写死本地盘。

func TestStorageLocalDriverUsesConfiguredPrefix(t *testing.T) {
	root := t.TempDir()
	st, err := utils.NewStorage(&config.StorageConfig{
		Driver: config.StorageDriverLocal,
		Local:  config.LocalStorageConfig{Root: root, BaseURL: "/api/v1/uploads"},
	})
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}

	got, err := st.Save(strings.NewReader("fakepng"), utils.SaveOptions{
		Category: "pet-album", Ext: "png", MaxSize: utils.MaxPostImageBytes, Kind: "图片",
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if !strings.HasPrefix(got, "/api/v1/uploads/pet-album/") {
		t.Fatalf("expected configured prefix, got %q", got)
	}
	rel := strings.TrimPrefix(got, "/api/v1/uploads/")
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatalf("object not persisted under root: %v", err)
	}
	if !strings.HasSuffix(got, ".png") {
		t.Errorf("unexpected url: %s", got)
	}
}

type capturedS3Request struct {
	method      string
	path        string
	auth        string
	contentType string
	body        string
}

func TestStorageOSSDriverUploadsToS3CompatibleEndpoint(t *testing.T) {
	var got capturedS3Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = capturedS3Request{
			method:      r.Method,
			path:        r.URL.Path,
			auth:        r.Header.Get("Authorization"),
			contentType: r.Header.Get("Content-Type"),
			body:        string(body),
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}

	st, err := utils.NewStorage(&config.StorageConfig{
		Driver: config.StorageDriverOSS,
		OSS: config.OSSStorageConfig{
			Endpoint:        u.Host,
			Region:          "cn-hangzhou",
			Bucket:          "pet-bucket",
			AccessKeyID:     "<REDACTED>",
			AccessKeySecret: "<REDACTED>",
			KeyPrefix:       "uploads",
			BaseURL:         "https://cdn.example.com",
			UseSSL:          false,
			PathStyle:       true,
		},
	})
	if err != nil {
		t.Fatalf("new oss storage: %v", err)
	}

	gotURL, err := st.Save(strings.NewReader("voicebytes"), utils.SaveOptions{
		Category: "messages/voice", Ext: "webm", MaxSize: utils.MaxVoiceBytes, Kind: "语音",
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if !strings.HasPrefix(gotURL, "https://cdn.example.com/uploads/messages/voice/") || !strings.HasSuffix(gotURL, ".webm") {
		t.Fatalf("expected oss public url, got %q", gotURL)
	}
	if got.method != http.MethodPut {
		t.Errorf("expected PUT to object storage, got %s", got.method)
	}
	if !strings.HasPrefix(got.path, "/pet-bucket/uploads/messages/voice/") {
		t.Errorf("unexpected object path: %s", got.path)
	}
	if !strings.HasPrefix(got.auth, "AWS4-HMAC-SHA256") {
		t.Errorf("missing SigV4 authorization header: %q", got.auth)
	}
	if got.body != "voicebytes" {
		t.Errorf("payload not uploaded: %q", got.body)
	}
	if got.contentType == "" {
		t.Errorf("content type should be set for browser playback")
	}

	// DELETE 也必须带 SigV4（空 body 哈希）
	if err := st.Delete("messages/voice/dummy.webm"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if got.method != http.MethodDelete {
		t.Errorf("expected DELETE after Delete(), got %s", got.method)
	}
	if !strings.HasPrefix(got.auth, "AWS4-HMAC-SHA256") {
		t.Errorf("DELETE missing SigV4 authorization: %q", got.auth)
	}
}

func TestStorageOSSDriverDefaultPublicURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)

	st, err := utils.NewStorage(&config.StorageConfig{
		Driver: config.StorageDriverOSS,
		OSS: config.OSSStorageConfig{
			Endpoint:        u.Host,
			Region:          "cn-hangzhou",
			Bucket:          "pet-bucket",
			AccessKeyID:     "<REDACTED>",
			AccessKeySecret: "<REDACTED>",
			KeyPrefix:       "uploads",
			UseSSL:          false,
		},
	})
	if err != nil {
		t.Fatalf("new oss storage: %v", err)
	}
	// 未配置 base_url 且 path_style=false 时按虚拟主机风格拼接（阿里云 OSS / COS / S3 默认）
	want := "http://pet-bucket." + u.Host + "/uploads/messages/image/a.jpg"
	if got := st.URL("messages/image/a.jpg"); got != want {
		t.Fatalf("unexpected default public url: got %s want %s", got, want)
	}
}

func TestStorageRejectsOversizeAndLeavesNoObject(t *testing.T) {
	root := t.TempDir()
	st, err := utils.NewStorage(&config.StorageConfig{
		Driver: config.StorageDriverLocal,
		Local:  config.LocalStorageConfig{Root: root, BaseURL: "/api/v1/uploads"},
	})
	if err != nil {
		t.Fatalf("new local storage: %v", err)
	}
	_, err = st.Save(strings.NewReader(strings.Repeat("x", 32)), utils.SaveOptions{
		Category: "avatar", Ext: "png", MaxSize: 8, Kind: "头像",
	})
	if err == nil {
		t.Fatal("expected oversize error")
	}
	entries, readErr := os.ReadDir(filepath.Join(root, "avatar"))
	if readErr != nil {
		t.Fatalf("read category dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Errorf("oversized object should be cleaned up, found %v", entries)
	}
}

func TestStorageConfigFromYAML(t *testing.T) {
	prev := config.Cfg
	t.Cleanup(func() { config.Cfg = prev })

	path := writeTempYAML(t, `
server:
  port: 8083
database:
  host: localhost
  port: 3306
  username: root
  password: secret
  database: pet_test
jwt:
  secret: test-secret-key-16
storage:
  driver: local
`)
	config.LoadFile(path)
	if config.Cfg.Storage.Driver != config.StorageDriverLocal {
		t.Fatalf("default driver: %s", config.Cfg.Storage.Driver)
	}
	if config.Cfg.Storage.Local.BaseURL != "/api/v1/uploads" {
		t.Errorf("default base_url: %s", config.Cfg.Storage.Local.BaseURL)
	}
	if config.Cfg.Storage.Local.Root != "public/uploads" {
		t.Errorf("default root: %s", config.Cfg.Storage.Local.Root)
	}
}

func TestStorageConfigValidationRules(t *testing.T) {
	// OSS 驱动缺少 bucket / 密钥时必须报错（启动时同样阻断）
	bad := config.StorageConfig{
		Driver: config.StorageDriverOSS,
		OSS:    config.OSSStorageConfig{Endpoint: "oss-cn-hangzhou.aliyuncs.com"},
	}
	if err := config.NormalizeStorageConfig(&bad); err == nil {
		t.Error("expected error for oss driver without bucket/credentials")
	}

	// 未知驱动必须报错
	unknown := config.StorageConfig{Driver: "ftp"}
	if err := config.NormalizeStorageConfig(&unknown); err == nil {
		t.Error("expected error for unknown driver")
	}

	// driver 缺省时回退本地磁盘，并补齐前缀
	def := config.StorageConfig{}
	if err := config.NormalizeStorageConfig(&def); err != nil {
		t.Fatalf("normalize default: %v", err)
	}
	if def.Driver != config.StorageDriverLocal || def.Local.BaseURL != "/api/v1/uploads" {
		t.Errorf("unexpected defaults: %+v", def)
	}

	// 前缀缺前导斜杠 / 多尾部斜杠都要规范化
	fixed := config.StorageConfig{Local: config.LocalStorageConfig{BaseURL: "uploads/"}}
	if err := config.NormalizeStorageConfig(&fixed); err != nil {
		t.Fatalf("normalize prefix: %v", err)
	}
	if fixed.Local.BaseURL != "/uploads" {
		t.Errorf("unexpected normalized prefix: %s", fixed.Local.BaseURL)
	}

	// 过宽前缀会让 CSRF 豁免覆盖写接口，必须拒绝（"/" 会回退默认合法前缀，不在此列）
	for _, broad := range []string{"/api", "/api/v1", "https://cdn.example.com/uploads"} {
		cfg := config.StorageConfig{Local: config.LocalStorageConfig{BaseURL: broad}}
		if err := config.NormalizeStorageConfig(&cfg); err == nil {
			t.Errorf("expected error for broad/invalid base_url %q", broad)
		}
	}
}

func TestRepoConfigYAMLDeclaresStorage(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "config.yaml"))
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}
	var cfg config.Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse config.yaml: %v", err)
	}
	if cfg.Storage.Driver != config.StorageDriverLocal {
		t.Fatalf("repo default driver should stay local, got %q", cfg.Storage.Driver)
	}
	if cfg.Storage.Local.BaseURL != "/api/v1/uploads" {
		t.Errorf("repo default upload prefix: %q", cfg.Storage.Local.BaseURL)
	}
}

func TestStorageDriverConfigValidation(t *testing.T) {
	if _, err := utils.NewStorage(&config.StorageConfig{Driver: "ftp"}); err == nil {
		t.Error("expected error for unknown driver")
	}
	if _, err := utils.NewStorage(&config.StorageConfig{Driver: config.StorageDriverOSS}); err == nil {
		t.Error("expected error for oss driver without endpoint/bucket/credentials")
	}
}

func TestMessageVoiceUploadFollowsConfiguredStorage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)

	ossStorage, err := utils.NewStorage(&config.StorageConfig{
		Driver: config.StorageDriverOSS,
		OSS: config.OSSStorageConfig{
			Endpoint:        u.Host,
			Region:          "cn-hangzhou",
			Bucket:          "pet-bucket",
			AccessKeyID:     "<REDACTED>",
			AccessKeySecret: "<REDACTED>",
			KeyPrefix:       "uploads",
			BaseURL:         "https://cdn.example.com",
			UseSSL:          false,
			PathStyle:       true,
		},
	})
	if err != nil {
		t.Fatalf("new oss storage: %v", err)
	}

	prev := utils.CurrentStorage()
	utils.SetStorage(ossStorage)
	t.Cleanup(func() { utils.SetStorage(prev) })

	db := testutil.SetupTestDB(t)
	msgSrv := service.NewMessageService(
		repository.NewConversationRepository(db),
		repository.NewMessageRepository(db),
		repository.NewUserRepository(db),
		repository.NewBlockRepository(db),
	)

	voiceURL, err := msgSrv.UploadMessageVoice(strings.NewReader("voicebytes"), "webm")
	if err != nil {
		t.Fatalf("upload voice failed: %v", err)
	}
	if !strings.HasPrefix(voiceURL, "https://cdn.example.com/uploads/messages/voice/") {
		t.Fatalf("voice should be stored on OSS, got %q", voiceURL)
	}
	if !strings.HasPrefix(service.MessageVoiceURLPrefix(), "https://cdn.example.com/") {
		t.Fatalf("voice url prefix should follow storage config, got %q", service.MessageVoiceURLPrefix())
	}
	if err := service.ValidateMessageMediaURL(voiceURL, service.MessageVoiceCategory); err != nil {
		t.Fatalf("uploaded voice url rejected: %v", err)
	}
	if _, err := msgSrv.SendMessageWithMedia(1, 2, "", model.MessageTypeVoice, voiceURL, 5); err == nil {
		t.Log("send skipped: users absent in fixture")
	}
}
