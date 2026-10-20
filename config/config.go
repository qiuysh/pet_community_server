package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置（统一从 config.yaml 加载）
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Database  DatabaseConfig  `yaml:"database"`
	JWT       JWTConfig       `yaml:"jwt"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Redis     RedisConfig     `yaml:"redis"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Storage   StorageConfig   `yaml:"storage"`
}

// 存储驱动：local 本地磁盘；oss / s3 为 S3 兼容对象存储（阿里云 OSS、腾讯云 COS、MinIO 等）
const (
	StorageDriverLocal = "local"
	StorageDriverOSS   = "oss"
	StorageDriverS3    = "s3"
)

// DefaultUploadBaseURL 本地驱动的默认对外访问前缀（需与静态路由挂载路径一致）
const DefaultUploadBaseURL = "/api/v1/uploads"

// DefaultUploadRoot 本地驱动的默认落盘根目录
const DefaultUploadRoot = "public/uploads"

// StorageConfig 文件存储配置。业务层统一通过 utils.Storage 读写，
// 切换本地磁盘 / 对象存储只改配置，不改业务代码。
type StorageConfig struct {
	// Driver 存储驱动：local | oss | s3
	Driver string             `yaml:"driver"`
	Local  LocalStorageConfig `yaml:"local"`
	OSS    OSSStorageConfig   `yaml:"oss"`
}

// LocalStorageConfig 本地磁盘驱动配置
type LocalStorageConfig struct {
	// Root 落盘根目录（相对进程工作目录，亦可填绝对路径）
	Root string `yaml:"root"`
	// BaseURL 对外访问前缀，必须与静态路由挂载路径一致
	BaseURL string `yaml:"base_url"`
}

// OSSStorageConfig S3 兼容对象存储配置
type OSSStorageConfig struct {
	// Endpoint 对象存储端点，形如 oss-cn-hangzhou.aliyuncs.com 或 127.0.0.1:9000
	Endpoint string `yaml:"endpoint"`
	// Region 签名用区域，MinIO 默认 us-east-1
	Region string `yaml:"region"`
	// Bucket 桶名
	Bucket string `yaml:"bucket"`
	// AccessKeyID / AccessKeySecret 访问密钥
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	// KeyPrefix object key 统一前缀（如 uploads），留空则直放桶根
	KeyPrefix string `yaml:"key_prefix"`
	// BaseURL 对外访问域名（CDN / 自定义域名），留空按 endpoint + bucket 自动拼接
	BaseURL string `yaml:"base_url"`
	// UseSSL 是否使用 https
	UseSSL bool `yaml:"use_ssl"`
	// PathStyle true=路径风格 https://endpoint/bucket/key（MinIO）；
	// false=虚拟主机风格 https://bucket.endpoint/key（阿里云 OSS / 腾讯云 COS / S3）
	PathStyle bool `yaml:"path_style"`
}

// DefaultStorageConfig 存储配置默认值（本地磁盘 + /api/v1/uploads 前缀）
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Driver: StorageDriverLocal,
		Local: LocalStorageConfig{
			Root:    DefaultUploadRoot,
			BaseURL: DefaultUploadBaseURL,
		},
		OSS: OSSStorageConfig{
			KeyPrefix: "uploads",
			Region:    "us-east-1",
			UseSSL:    true,
		},
	}
}

// ServerConfig 服务监听相关配置
type ServerConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout int    `yaml:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"database"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret string `yaml:"secret"`
}

// SchedulerConfig 定制/定时任务 cron（含秒字段，配合 cron.WithSeconds）
type SchedulerConfig struct {
	StatsCron            string `yaml:"stats_cron"`
	NotificationCron     string `yaml:"notification_cron"`
	ScheduledPostCron    string `yaml:"scheduled_post_cron"`
	StatsStartupDelaySec int    `yaml:"stats_startup_delay_sec"`
}

// defaultConfig 代码内默认值（YAML 未写的字段保留默认）
func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8083,
			Timeout: 30,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "",
			DBName:   "pet_community",
		},
		JWT: JWTConfig{
			Secret: "pet-community-secret-key",
		},
		Scheduler: SchedulerConfig{
			StatsCron:            "0 0 2 * * *",
			NotificationCron:     "0 * * * * *",
			ScheduledPostCron:    "0 */5 * * * *",
			StatsStartupDelaySec: 5,
		},
		Redis: RedisConfig{
			Enabled:  false,
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		RateLimit: RateLimitConfig{
			Enabled:         true,
			PerIPPerRoute:   100,
			PerIPGlobal:     500,
			PerUserPerRoute: 100,
			WindowSeconds:   60,
		},
		Storage: DefaultStorageConfig(),
	}
}

// Load 从 config.yaml 加载配置（不分环境，不读取 .env）
func Load() {
	path, err := resolveConfigPath()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := loadFromFile(path); err != nil {
		log.Fatalf("config: %v", err)
	}
	log.Printf("config loaded from %s", path)
}

// LoadFile 从指定 YAML 路径加载，主要用于测试
func LoadFile(path string) {
	if err := loadFromFile(path); err != nil {
		log.Fatalf("config: %v", err)
	}
}

func loadFromFile(path string) error {
	cfg := defaultConfig()

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}
	if err := validateConfig(&cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	Cfg = &cfg
	return nil
}

// Addr 返回监听地址 host:port
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func resolveConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}

	// 与原 .env 同级：server/config.yaml（cwd 为 server 或仓库根均可）
	candidates := []string{
		filepath.Join(cwd, "config.yaml"),
		filepath.Join(cwd, "server", "config.yaml"),
		filepath.Join(cwd, "..", "config.yaml"),
		filepath.Join(cwd, "..", "server", "config.yaml"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("config.yaml not found, searched: %v", candidates)
}

func validateConfig(cfg *Config) error {
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port out of range: %d", cfg.Server.Port)
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return fmt.Errorf("database.port out of range: %d", cfg.Database.Port)
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database.username is required")
	}
	if cfg.Database.DBName == "" {
		return fmt.Errorf("database.database is required")
	}
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret is required")
	}
	if len(cfg.JWT.Secret) < 16 {
		return fmt.Errorf("jwt.secret length must be >= 16")
	}
	if cfg.Scheduler.StatsCron == "" || cfg.Scheduler.NotificationCron == "" || cfg.Scheduler.ScheduledPostCron == "" {
		return fmt.Errorf("scheduler cron expressions must not be empty")
	}
	if cfg.Redis.Enabled {
		if cfg.Redis.Host == "" {
			return fmt.Errorf("redis.host is required when redis.enabled")
		}
		if cfg.Redis.Port < 1 || cfg.Redis.Port > 65535 {
			return fmt.Errorf("redis.port out of range: %d", cfg.Redis.Port)
		}
	}
	if cfg.RateLimit.WindowSeconds <= 0 {
		cfg.RateLimit.WindowSeconds = 60
	}
	if cfg.RateLimit.PerIPPerRoute <= 0 {
		cfg.RateLimit.PerIPPerRoute = 100
	}
	if cfg.RateLimit.PerIPGlobal <= 0 {
		cfg.RateLimit.PerIPGlobal = 500
	}
	if cfg.RateLimit.PerUserPerRoute <= 0 {
		cfg.RateLimit.PerUserPerRoute = 100
	}
	if err := NormalizeStorageConfig(&cfg.Storage); err != nil {
		return err
	}
	return nil
}

// NormalizeStorageConfig 规范化并校验存储配置（导出便于测试与复用）
func NormalizeStorageConfig(cfg *StorageConfig) error {
	cfg.Driver = strings.ToLower(strings.TrimSpace(cfg.Driver))
	if cfg.Driver == "" {
		cfg.Driver = StorageDriverLocal
	}
	switch cfg.Driver {
	case StorageDriverLocal, StorageDriverOSS, StorageDriverS3:
	default:
		return fmt.Errorf("storage.driver must be one of local/oss/s3, got %q", cfg.Driver)
	}

	cfg.Local.Root = strings.TrimSpace(cfg.Local.Root)
	if cfg.Local.Root == "" {
		cfg.Local.Root = DefaultUploadRoot
	}
	cfg.Local.BaseURL = normalizeURLPrefix(cfg.Local.BaseURL, DefaultUploadBaseURL)
	if err := validateLocalBaseURL(cfg.Local.BaseURL); err != nil {
		return err
	}

	if cfg.Driver == StorageDriverLocal {
		return nil
	}

	oss := &cfg.OSS
	oss.Endpoint = strings.TrimSpace(oss.Endpoint)
	oss.Region = strings.TrimSpace(oss.Region)
	oss.Bucket = strings.TrimSpace(oss.Bucket)
	oss.AccessKeyID = strings.TrimSpace(oss.AccessKeyID)
	oss.AccessKeySecret = strings.TrimSpace(oss.AccessKeySecret)
	oss.KeyPrefix = strings.Trim(oss.KeyPrefix, "/")
	oss.BaseURL = strings.TrimSuffix(strings.TrimSpace(oss.BaseURL), "/")

	switch {
	case oss.Endpoint == "":
		return errors.New("storage.oss.endpoint is required when storage.driver is oss/s3")
	case oss.Bucket == "":
		return errors.New("storage.oss.bucket is required when storage.driver is oss/s3")
	case oss.AccessKeyID == "" || oss.AccessKeySecret == "":
		return errors.New("storage.oss.access_key_id / access_key_secret are required when storage.driver is oss/s3")
	}
	if oss.Region == "" {
		oss.Region = "us-east-1"
	}
	return nil
}

// normalizeURLPrefix 规范化 URL 前缀：去尾部斜杠，缺前导斜杠时补齐（绝对地址除外）
func normalizeURLPrefix(raw, fallback string) string {
	p := strings.TrimSpace(raw)
	if p == "" {
		return fallback
	}
	p = strings.TrimRight(p, "/")
	if p == "" {
		return fallback
	}
	if !strings.HasPrefix(p, "/") && !strings.Contains(p, "://") {
		p = "/" + p
	}
	return p
}

// validateLocalBaseURL 约束本地对外前缀，避免误配成 /api 导致 CSRF 写接口被整段豁免。
// 规则：相对路径；末段必须为 uploads（如 /api/v1/uploads、/uploads）。
func validateLocalBaseURL(baseURL string) error {
	if strings.Contains(baseURL, "://") {
		return errors.New("storage.local.base_url must be a relative path (e.g. /api/v1/uploads)")
	}
	if !strings.HasPrefix(baseURL, "/") {
		return fmt.Errorf("storage.local.base_url must start with /, got %q", baseURL)
	}
	if strings.Contains(baseURL, "..") {
		return errors.New("storage.local.base_url must not contain ..")
	}
	trimmed := strings.Trim(baseURL, "/")
	if trimmed == "" {
		return errors.New("storage.local.base_url is empty")
	}
	parts := strings.Split(trimmed, "/")
	if parts[len(parts)-1] != "uploads" {
		return fmt.Errorf("storage.local.base_url must end with /uploads (got %q); broader prefixes would bypass CSRF on API writes", baseURL)
	}
	return nil
}
