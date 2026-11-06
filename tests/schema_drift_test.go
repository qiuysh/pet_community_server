package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"pet-community/config"
	"pet-community/tests/testutil"

	"gorm.io/gorm"
)

// 建表语句块：CREATE TABLE IF NOT EXISTS xxx ( ... ) ENGINE=...
var createTableBlockRe = regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS (\w+) \((.*?)\n\) ENGINE`)

// 列定义行：可选反引号 + 列名 + 类型（索引/外键行会被显式排除）
var columnLineRe = regexp.MustCompile("^`?([a-zA-Z_][a-zA-Z0-9_]*)`?\\s+[a-zA-Z]")

var nonColumnKeywords = []string{
	"UNIQUE KEY", "INDEX ", "KEY ", "PRIMARY KEY", "FOREIGN KEY", "CONSTRAINT", "CHECK", "FULLTEXT",
}

// parseCreateSQLColumns 解析 init/create.sql，返回 表名 -> 列名集合
func parseCreateSQLColumns(t *testing.T) map[string]map[string]bool {
	t.Helper()

	path := filepath.Join("..", "sql", "init", "create.sql")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取 %s 失败: %v", path, err)
	}

	tables := make(map[string]map[string]bool)
	for _, block := range createTableBlockRe.FindAllStringSubmatch(string(raw), -1) {
		table, body := block[1], block[2]
		cols := make(map[string]bool)
		for _, line := range strings.Split(body, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}
			skipped := false
			for _, kw := range nonColumnKeywords {
				if strings.HasPrefix(strings.ToUpper(line), kw) {
					skipped = true
					break
				}
			}
			if skipped {
				continue
			}
			if m := columnLineRe.FindStringSubmatch(line); m != nil {
				cols[m[1]] = true
			}
		}
		tables[table] = cols
	}

	if len(tables) == 0 {
		t.Fatalf("从 %s 未解析到任何建表语句，解析规则可能失效", path)
	}
	return tables
}

// TestModelColumnsExistInCreateSQL 守住「模型字段必须出现在建表脚本里」。
//
// 背景：create.sql 曾长期落后于 model/*.go（users.token_version、topics.creator_id 等），
// 只跑 SQL 建库的环境（导库、QA、BI 只读库）会在登录/建话题时直接 1054 Unknown column。
// 通过 Go 启动的实例有 AutoMigrate 兜底，所以这类漂移在开发机上完全不报错，
// 只有换环境才炸 —— 因此需要一条不依赖数据库的静态校验。
func TestModelColumnsExistInCreateSQL(t *testing.T) {
	db := testutil.SetupTestDB(t)
	tables := parseCreateSQLColumns(t)

	for _, m := range config.Models() {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(m); err != nil {
			t.Fatalf("解析模型 %T 失败: %v", m, err)
		}
		table := stmt.Schema.Table
		cols, ok := tables[table]
		if !ok {
			t.Errorf("create.sql 缺少表 %q（模型 %T）", table, m)
			continue
		}
		for _, field := range stmt.Schema.Fields {
			if field.DBName == "" {
				continue // 关联字段，没有对应列
			}
			if !cols[field.DBName] {
				t.Errorf("%s.%s 在建表脚本中缺失（模型 %T 的字段 %s）——请同步 sql/init/create.sql 并补增量脚本",
					table, field.DBName, m, field.Name)
			}
		}
	}
}
