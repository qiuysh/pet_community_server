package tests

import (
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

func setupAdService(db *gorm.DB) *service.AdService {
	adRepo := repository.NewAdRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	return service.NewAdService(adRepo, adminRepo)
}

func createTestAdPosition(db *gorm.DB) *model.AdPosition {
	position := &model.AdPosition{
		Name:     "Home Banner",
		Location: "home",
		Size:     "750x300",
		Status:   1,
	}
	db.Create(position)
	return position
}

func TestAdServiceCreatePosition(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	t.Run("success", func(t *testing.T) {
		position, err := srv.CreatePosition("Banner", "home", "750x300", 1)
		if err != nil {
			t.Fatalf("create position failed: %v", err)
		}
		if position.ID == 0 {
			t.Error("position ID should not be zero")
		}
	})

	t.Run("default status", func(t *testing.T) {
		position, err := srv.CreatePosition("Footer", "footer", "300x250", 0)
		if err != nil {
			t.Fatalf("create position failed: %v", err)
		}
		if position.Status != 1 {
			t.Errorf("expected status 1, got %d", position.Status)
		}
	})
}

func TestAdServiceUpdatePosition(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)

	t.Run("success", func(t *testing.T) {
		updated, err := srv.UpdatePosition(position.ID, "Updated", "home", "800x400", 1)
		if err != nil {
			t.Fatalf("update position failed: %v", err)
		}
		if updated.Name != "Updated" {
			t.Errorf("expected name 'Updated', got %s", updated.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.UpdatePosition(9999, "Ghost", "home", "800x400", 1)
		if err == nil {
			t.Error("expected error for non-existent position")
		}
	})
}

func TestAdServiceDeletePosition(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)

	t.Run("success", func(t *testing.T) {
		err := srv.DeletePosition(position.ID)
		if err != nil {
			t.Fatalf("delete position failed: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := srv.DeletePosition(9999)
		if err == nil {
			t.Error("expected error for non-existent position")
		}
	})
}

func TestAdServiceCreateAd(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(time.Hour).Format("2006-01-02 15:04:05")

	t.Run("success", func(t *testing.T) {
		ad, err := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", startTime, endTime, 1)
		if err != nil {
			t.Fatalf("create ad failed: %v", err)
		}
		if ad.ID == 0 {
			t.Error("ad ID should not be zero")
		}
	})

	t.Run("position not found", func(t *testing.T) {
		_, err := srv.CreateAd(9999, "Ad Title", "http://image.jpg", "http://link.com", startTime, endTime, 1)
		if err == nil {
			t.Error("expected error for non-existent position")
		}
	})

	t.Run("invalid time", func(t *testing.T) {
		_, err := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", "bad-time", endTime, 1)
		if err == nil {
			t.Error("expected error for invalid start time")
		}
	})

	t.Run("end before start", func(t *testing.T) {
		_, err := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", endTime, startTime, 1)
		if err == nil {
			t.Error("expected error for end before start")
		}
	})
}

func TestAdServiceUpdateAd(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(time.Hour).Format("2006-01-02 15:04:05")
	ad, _ := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", startTime, endTime, 1)

	t.Run("success", func(t *testing.T) {
		updated, err := srv.UpdateAd(ad.ID, position.ID, "Updated Title", "", "", "", "")
		if err != nil {
			t.Fatalf("update ad failed: %v", err)
		}
		if updated.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got %s", updated.Title)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.UpdateAd(9999, position.ID, "Ghost", "", "", "", "")
		if err == nil {
			t.Error("expected error for non-existent ad")
		}
	})
}

func TestAdServiceReviewAd(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(time.Hour).Format("2006-01-02 15:04:05")
	ad, _ := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", startTime, endTime, 1)

	t.Run("approve", func(t *testing.T) {
		err := srv.ReviewAd(ad.ID, 1, 1, "127.0.0.1")
		if err != nil {
			t.Fatalf("review ad failed: %v", err)
		}
	})

	t.Run("same status", func(t *testing.T) {
		err := srv.ReviewAd(ad.ID, 1, 1, "127.0.0.1")
		if err == nil {
			t.Error("expected error for same status")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := srv.ReviewAd(9999, 2, 1, "127.0.0.1")
		if err == nil {
			t.Error("expected error for non-existent ad")
		}
	})
}

func TestAdServiceIncrementCounts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(time.Hour).Format("2006-01-02 15:04:05")
	ad, _ := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", startTime, endTime, 1)

	t.Run("view", func(t *testing.T) {
		err := srv.IncrementViewCount(ad.ID)
		if err != nil {
			t.Fatalf("increment view failed: %v", err)
		}
	})

	t.Run("click", func(t *testing.T) {
		err := srv.IncrementClickCount(ad.ID)
		if err != nil {
			t.Fatalf("increment click failed: %v", err)
		}
	})
}

func TestAdServiceGetAdStats(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdService(db)

	position := createTestAdPosition(db)
	now := time.Now()
	startTime := now.Format("2006-01-02 15:04:05")
	endTime := now.Add(time.Hour).Format("2006-01-02 15:04:05")
	ad, _ := srv.CreateAd(position.ID, "Ad Title", "http://image.jpg", "http://link.com", startTime, endTime, 1)
	srv.ReviewAd(ad.ID, 1, 1, "127.0.0.1")
	srv.IncrementViewCount(ad.ID)
	srv.IncrementClickCount(ad.ID)

	t.Run("default range", func(t *testing.T) {
		stats, err := srv.GetAdStats("", "")
		if err != nil {
			t.Fatalf("get ad stats failed: %v", err)
		}
		if stats.TotalViews != 1 {
			t.Errorf("expected total views 1, got %d", stats.TotalViews)
		}
		if stats.TotalClicks != 1 {
			t.Errorf("expected total clicks 1, got %d", stats.TotalClicks)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		_, err := srv.GetAdStats("bad-date", "")
		if err == nil {
			t.Error("expected error for invalid date")
		}
	})
}

// TestAdServiceGetActiveAds 前台展示核心：按位置过滤 + 时间窗 + status=1 + 位置 status=1
func TestAdServiceGetActiveAds(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db = db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Info)})
	srv := setupAdService(db)

	// 准备：1 个启用位置 + 1 个禁用位置
	posOn := &model.AdPosition{Name: "Home Banner", Location: "home_banner", Size: "750x300", Status: 1}
	posOff := &model.AdPosition{Name: "Sidebar Off", Location: "detail_sidebar", Size: "300x250", Status: 1}
	db.Create(posOn)
	db.Create(posOff)
	// 模拟运营「先创建再禁用」：default:1 导致 Create 跳过 0 值，必须显式 Update 写 0
	if err := db.Model(&model.AdPosition{}).Where("id = ?", posOff.ID).Update("status", 0).Error; err != nil {
		t.Fatalf("failed to disable posOff: %v", err)
	}
	posOff.Status = 0

	// 启用位置下放 3 条广告：生效中 / 未到开始 / 已结束
	now := time.Now()
	db.Create(&model.Advertisement{
		PositionID: posOn.ID, Title: "live", ImageURL: "/img/live.jpg", LinkURL: "/live",
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), Status: 1,
	})
	db.Create(&model.Advertisement{
		PositionID: posOn.ID, Title: "future", ImageURL: "/img/future.jpg", LinkURL: "/future",
		StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour), Status: 1,
	})
	db.Create(&model.Advertisement{
		PositionID: posOn.ID, Title: "expired", ImageURL: "/img/expired.jpg", LinkURL: "/expired",
		StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-time.Hour), Status: 1,
	})
	// 启用位置下放 1 条审核未通过（status=0）
	db.Create(&model.Advertisement{
		PositionID: posOn.ID, Title: "pending", ImageURL: "/img/pending.jpg", LinkURL: "/pending",
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), Status: 0,
	})
	// 禁用位置下放 1 条「本应生效」
	db.Create(&model.Advertisement{
		PositionID: posOff.ID, Title: "pos-off", ImageURL: "/img/off.jpg", LinkURL: "/off",
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), Status: 1,
	})

	t.Run("home_banner returns only the live ad", func(t *testing.T) {
		ads, err := srv.GetActiveAds("home_banner")
		if err != nil {
			t.Fatalf("GetActiveAds failed: %v", err)
		}
		if len(ads) != 1 {
			t.Fatalf("expected 1 active ad, got %d: %+v", len(ads), ads)
		}
		if ads[0].Title != "live" {
			t.Errorf("expected 'live', got %q", ads[0].Title)
		}
	})

	t.Run("detail_sidebar (position disabled) returns nothing", func(t *testing.T) {
		ads, err := srv.GetActiveAds("detail_sidebar")
		if err != nil {
			t.Fatalf("GetActiveAds failed: %v", err)
		}
		if len(ads) != 0 {
			t.Errorf("expected 0 ads for disabled position, got %d", len(ads))
		}
	})

	t.Run("empty location returns all live ads across all enabled positions", func(t *testing.T) {
		ads, err := srv.GetActiveAds("")
		if err != nil {
			t.Fatalf("GetActiveAds failed: %v", err)
		}
		if len(ads) != 1 {
			t.Errorf("expected exactly 1 live ad overall, got %d", len(ads))
		}
	})

	t.Run("increment view and click counts", func(t *testing.T) {
		ads, _ := srv.GetActiveAds("home_banner")
		if len(ads) != 1 {
			t.Fatalf("setup failed: expected 1 ad")
		}
		before := ads[0].ViewCount
		if err := srv.IncrementViewCount(ads[0].ID); err != nil {
			t.Fatalf("IncrementViewCount failed: %v", err)
		}
		after, _ := srv.GetAdByID(ads[0].ID)
		if after.ViewCount != before+1 {
			t.Errorf("view count: expected %d, got %d", before+1, after.ViewCount)
		}
		clicksBefore := after.ClickCount
		if err := srv.IncrementClickCount(ads[0].ID); err != nil {
			t.Fatalf("IncrementClickCount failed: %v", err)
		}
		after2, _ := srv.GetAdByID(ads[0].ID)
		if after2.ClickCount != clicksBefore+1 {
			t.Errorf("click count: expected %d, got %d", clicksBefore+1, after2.ClickCount)
		}
	})
}
