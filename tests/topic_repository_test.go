package tests

import (
	"testing"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/tests/testutil"
)

// TestTopicRepositoryCreateFillsCategoryID 回归「新建话题后读不出来」的线上故障。
//
// 故障成因：TopicRepository.Create 用显式 Select 写白名单，一旦漏掉 CategoryID，
// 该列就走 DB 默认值；而 create.sql 里 category_id 曾是可空、无默认、带外键，
// 于是落成 NULL，之后 SELECT * 把 NULL 扫进 model.Topic.CategoryID（uint）直接报
// converting NULL to uint is unsupported —— 表现为话题详情/后台待审列表 500。
//
// 这里刻意用「旧 DDL」（category_id 可空无默认）重建表，因为 AutoMigrate 建出的表带
// DEFAULT 0，测不出漏列问题。
func TestTopicRepositoryCreateFillsCategoryID(t *testing.T) {
	db := testutil.SetupTestDB(t)

	if err := db.Exec("DROP TABLE topics").Error; err != nil {
		t.Fatalf("drop topics: %v", err)
	}
	if err := db.Exec(`CREATE TABLE topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		category_id INTEGER,
		name varchar(50) NOT NULL,
		description varchar(200),
		cover varchar(255),
		creator_id INTEGER DEFAULT 0,
		posts_count INTEGER DEFAULT 0,
		shares_count INTEGER DEFAULT 0,
		status INTEGER DEFAULT 0,
		created_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create topics with legacy DDL: %v", err)
	}

	repo := repository.NewTopicRepository(db)

	// 新建话题的典型入参：未选分类（零值）、直接进入待审核（Status=0，也是零值）
	topic := &model.Topic{Name: "新话题", Status: 0}
	if err := repo.Create(topic); err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if topic.ID == 0 {
		t.Fatal("topic id not backfilled")
	}

	var categoryID *int64
	if err := db.Raw("SELECT category_id FROM topics WHERE id = ?", topic.ID).Scan(&categoryID).Error; err != nil {
		t.Fatalf("query category_id: %v", err)
	}
	if categoryID == nil {
		t.Fatal("category_id 落成了 NULL：Create 的 Select 白名单漏了 CategoryID，读取时会 converting NULL to uint 报错")
	}
	if *categoryID != 0 {
		t.Fatalf("category_id = %d, want 0 (未分类哨兵)", *categoryID)
	}

	// 读回验证：列值为 NULL 时这一步会直接报错
	got, err := repo.GetByID(topic.ID)
	if err != nil {
		t.Fatalf("GetByID after create: %v", err)
	}
	if got.CategoryID != 0 || got.Status != 0 {
		t.Fatalf("got categoryID=%d status=%d, want 0/0", got.CategoryID, got.Status)
	}
}
