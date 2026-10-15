package tests

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

func setupDashboardService(db *gorm.DB) *service.DashboardService {
	adminRepo := repository.NewAdminRepository(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	petRepo := repository.NewPetRepository(db)
	statsRepo := repository.NewStatsRepository(db)
	reportRepo := repository.NewReportRepository(db)
	return service.NewDashboardService(adminRepo, userRepo, postRepo, petRepo, statsRepo, reportRepo)
}

func TestDashboardServiceGetDashboardOverview(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupDashboardService(db)

	now := time.Now()
	db.Create(&model.User{Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1, LastLoginAt: &now, CreatedAt: now})
	db.Create(&model.User{Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1, CreatedAt: now})
	db.Create(&model.User{Nickname: "Banned", Phone: "banned_phone", Email: "banned@test.com", Status: 0})

	db.Create(&model.Post{UserID: 1, Content: "post1", Status: 1, CreatedAt: now})
	db.Create(&model.Post{UserID: 1, Content: "post2", Status: 0, CreatedAt: now})
	db.Create(&model.Like{UserID: 1, TargetType: "post", TargetID: 1, CreatedAt: now})
	db.Create(&model.Comment{UserID: 1, PostID: 1, Content: "hi", CreatedAt: now})
	db.Create(&model.Favorite{UserID: 1, PostID: 1, CreatedAt: now})
	db.Create(&model.ShareRecord{UserID: 1, PostID: 1, Platform: "link", CreatedAt: now})
	db.Create(&model.Pet{UserID: 1, Name: "Doggo", PetType: 1, Status: 1})
	db.Create(&model.LostPet{UserID: 1, PetName: "Momo", PetType: 1, Contact: "13800000000", Status: 1, LostTime: now})
	db.Create(&model.Activity{Title: "Act", Description: "d", Type: "online", Status: 1, StartTime: now, EndTime: now.Add(24 * time.Hour)})
	db.Create(&model.Registration{ActivityID: 1, UserID: 1, Status: 0, Name: "A", Phone: "138", CreatedAt: now})
	db.Create(&model.Vote{ActivityID: 1, SubmissionID: 1, UserID: 1, CreatedAt: now})
	db.Create(&model.Advertisement{
		PositionID: 1, Title: "ad", ImageURL: "http://x.jpg",
		StartTime: now, EndTime: now.Add(time.Hour), Status: 1, CreatedBy: 1,
		ViewCount: 100, ClickCount: 10,
	})
	db.Create(&model.Report{ReporterID: 1, TargetType: "post", TargetID: 1, Reason: "spam", ReasonType: "spam", Status: 1})
	db.Create(&model.Report{ReporterID: 1, TargetType: "post", TargetID: 2, Reason: "spam", ReasonType: "spam", Status: 0})

	t.Run("overview", func(t *testing.T) {
		overview, err := srv.GetDashboardOverview()
		if err != nil {
			t.Fatalf("get overview failed: %v", err)
		}
		if overview.Summary.TotalUsers != 2 {
			t.Errorf("expected total users 2, got %d", overview.Summary.TotalUsers)
		}
		if overview.Summary.TodayFavorites != 1 {
			t.Errorf("expected today favorites 1, got %d", overview.Summary.TodayFavorites)
		}
		if overview.Summary.TodayShares != 1 {
			t.Errorf("expected today shares 1, got %d", overview.Summary.TodayShares)
		}
		if overview.Summary.TodayRegistrations != 1 {
			t.Errorf("expected today registrations 1, got %d", overview.Summary.TodayRegistrations)
		}
		if overview.Summary.TodayVotes != 1 {
			t.Errorf("expected today votes 1, got %d", overview.Summary.TodayVotes)
		}
		if overview.Summary.AdTotalViews != 100 {
			t.Errorf("expected ad views 100, got %d", overview.Summary.AdTotalViews)
		}
		if overview.Summary.AdTotalClicks != 10 {
			t.Errorf("expected ad clicks 10, got %d", overview.Summary.AdTotalClicks)
		}
		if overview.Summary.OngoingLostPets != 1 {
			t.Errorf("expected ongoing lost pets 1, got %d", overview.Summary.OngoingLostPets)
		}
		if overview.Summary.ReportHandleRate != 50 {
			t.Errorf("expected report handle rate 50, got %v", overview.Summary.ReportHandleRate)
		}
		if len(overview.TodayTrend.Data) != 7 {
			t.Errorf("expected 7 user trend points, got %d", len(overview.TodayTrend.Data))
		}
		if len(overview.PostTrend.Data) != 7 {
			t.Errorf("expected 7 post trend points, got %d", len(overview.PostTrend.Data))
		}
	})
}

func TestDashboardServiceGetTodoAggregate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupDashboardService(db)

	db.Create(&model.Post{UserID: 1, Content: "pending", Status: 0})
	db.Create(&model.Report{ReporterID: 1, TargetType: "post", TargetID: 1, Reason: "spam", ReasonType: "spam", Status: 0})
	db.Create(&model.Advertisement{PositionID: 1, Title: "ad", ImageURL: "http://x.jpg", StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), Status: 0, CreatedBy: 1})
	db.Create(&model.AdoptionApplication{AdoptionID: 1, UserID: 2, Status: 0})
	db.Create(&model.Registration{ActivityID: 1, UserID: 1, Status: 0, Name: "A", Phone: "138"})

	t.Run("aggregate", func(t *testing.T) {
		aggregate, err := srv.GetTodoAggregate()
		if err != nil {
			t.Fatalf("get todo aggregate failed: %v", err)
		}
		if aggregate.Total == 0 {
			t.Error("expected non-zero total todo count")
		}

		foundKeys := make(map[string]int)
		for _, item := range aggregate.Items {
			foundKeys[item.Key] = item.Count
		}
		if foundKeys["pending_posts"] != 1 {
			t.Errorf("expected pending_posts 1, got %d", foundKeys["pending_posts"])
		}
		if foundKeys["pending_reports"] != 1 {
			t.Errorf("expected pending_reports 1, got %d", foundKeys["pending_reports"])
		}
		if foundKeys["pending_ads"] != 1 {
			t.Errorf("expected pending_ads 1, got %d", foundKeys["pending_ads"])
		}
		// 领养申请一审由发布者处理，不进入运营驾驶舱待办
		if _, ok := foundKeys["pending_adoptions"]; ok {
			t.Error("pending_adoptions should not appear in admin todo aggregate")
		}
		if foundKeys["pending_registrations"] != 1 {
			t.Errorf("expected pending_registrations 1, got %d", foundKeys["pending_registrations"])
		}
	})
}
