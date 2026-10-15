package tests

import (
	"os"
	"path/filepath"
	"testing"

	"pet-community/config"
)

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return path
}

func TestLoadYAMLConfig(t *testing.T) {
	path := writeTempYAML(t, `
server:
  host: "127.0.0.1"
  port: 9090
  timeout: 15

database:
  host: db.local
  port: 3307
  username: app
  password: secret
  database: pet_test

jwt:
  secret: test-secret-key-16

scheduler:
  stats_cron: "0 30 3 * * *"
  notification_cron: "0 */2 * * * *"
  scheduled_post_cron: "0 */10 * * * *"
  stats_startup_delay_sec: 8
`)

	config.Cfg = nil
	config.LoadFile(path)

	if config.Cfg.Server.Host != "127.0.0.1" || config.Cfg.Server.Port != 9090 {
		t.Errorf("server: %+v", config.Cfg.Server)
	}
	if config.Cfg.Database.Host != "db.local" || config.Cfg.Database.Port != 3307 {
		t.Errorf("database: %+v", config.Cfg.Database)
	}
	if config.Cfg.Database.User != "app" || config.Cfg.Database.DBName != "pet_test" {
		t.Errorf("database user/name: %+v", config.Cfg.Database)
	}
	if config.Cfg.JWT.Secret != "test-secret-key-16" {
		t.Errorf("jwt: %s", config.Cfg.JWT.Secret)
	}
	s := config.Cfg.Scheduler
	if s.StatsCron != "0 30 3 * * *" || s.StatsStartupDelaySec != 8 {
		t.Errorf("scheduler: %+v", s)
	}
	if got := config.Cfg.Addr(); got != "127.0.0.1:9090" {
		t.Errorf("Addr: %s", got)
	}
}
