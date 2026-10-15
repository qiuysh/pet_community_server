package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	return nil
}
