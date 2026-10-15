package tests

import (
	"testing"

	"gorm.io/gorm"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

func setupAdminReportService(db *gorm.DB) *service.AdminReportService {
	return service.NewAdminReportService(
		repository.NewReportRepository(db),
		repository.NewAdminRepository(db),
		repository.NewUserRepository(db),
	)
}

func TestAdminReportService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdminReportService(db)

	db.Create(&model.User{Nickname: "Reporter", Phone: "r1", Email: "r1@test.com", Status: 1})
	db.Create(&model.Post{UserID: 1, Content: "bad post", Status: 1})
	db.Create(&model.Report{ReporterID: 1, TargetType: "post", TargetID: 1, Reason: "spam content", ReasonType: "spam", Status: 0})
	db.Create(&model.Report{ReporterID: 1, TargetType: "post", TargetID: 1, Reason: "attack", ReasonType: "attack", Status: 0})

	t.Run("list and filter", func(t *testing.T) {
		pending := 0
		res, err := srv.ListReports(&pending, "spam", "post", 1, 20)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}
		if res.Total != 1 {
			t.Fatalf("expected 1 spam pending, got %d", res.Total)
		}
		if res.List[0].ReporterName == "" {
			t.Error("expected reporter name")
		}
	})

	t.Run("handle approve", func(t *testing.T) {
		if err := srv.HandleReport(1, "approve", "下架违规内容", 99, "127.0.0.1"); err != nil {
			t.Fatalf("handle failed: %v", err)
		}
		var report model.Report
		db.First(&report, 1)
		if report.Status != 1 {
			t.Errorf("expected status 1, got %d", report.Status)
		}
		var post model.Post
		db.First(&post, 1)
		if post.Status != 2 {
			t.Errorf("expected post status 2 after approve, got %d", post.Status)
		}
	})

	t.Run("handle already processed", func(t *testing.T) {
		err := srv.HandleReport(1, "reject", "", 99, "127.0.0.1")
		if err == nil {
			t.Error("expected error for already handled report")
		}
	})

	t.Run("batch reject", func(t *testing.T) {
		if err := srv.BatchHandleReports([]uint{2}, "reject", "证据不足", 99, "127.0.0.1"); err != nil {
			t.Fatalf("batch failed: %v", err)
		}
		var report model.Report
		db.First(&report, 2)
		if report.Status != 2 {
			t.Errorf("expected status 2, got %d", report.Status)
		}
	})

	t.Run("stats", func(t *testing.T) {
		stats, err := srv.GetStats()
		if err != nil {
			t.Fatalf("stats failed: %v", err)
		}
		if stats.Total != 2 {
			t.Errorf("expected total 2, got %d", stats.Total)
		}
		if stats.Handled != 1 || stats.Ignored != 1 {
			t.Errorf("expected handled=1 ignored=1, got handled=%d ignored=%d", stats.Handled, stats.Ignored)
		}
		if stats.HandleRate < 99 {
			t.Errorf("expected handle rate ~100, got %f", stats.HandleRate)
		}
	})
}
