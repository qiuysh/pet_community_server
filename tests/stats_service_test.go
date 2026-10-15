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

func setupStatsService(db *gorm.DB) *service.StatsService {
	statsRepo := repository.NewStatsRepository(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	shareRepo := repository.NewShareRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	return service.NewStatsService(statsRepo, userRepo, postRepo, commentRepo, likeRepo, favoriteRepo, shareRepo, topicRepo)
}

func TestStatsServiceGetOverviewStats(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupStatsService(db)

	today := time.Now().Format("2006-01-02")
	db.Create(&model.User{Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.User{Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1, CreatedAt: time.Now().AddDate(0, 0, -1)})
	db.Create(&model.Post{UserID: 1, Content: "post1", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.Comment{PostID: 1, UserID: 1, Content: "comment", CreatedAt: time.Now()})
	db.Create(&model.Like{TargetType: "post", TargetID: 1, UserID: 1, CreatedAt: time.Now()})

	t.Run("calculate from database", func(t *testing.T) {
		stats, err := srv.GetOverviewStats()
		if err != nil {
			t.Fatalf("get overview stats failed: %v", err)
		}
		if stats.UserCount != 2 {
			t.Errorf("expected user count 2, got %d", stats.UserCount)
		}
		if stats.NewUserCount != 1 {
			t.Errorf("expected new user count 1, got %d", stats.NewUserCount)
		}
		if stats.PostCount != 1 {
			t.Errorf("expected post count 1, got %d", stats.PostCount)
		}
		if stats.NewPostCount != 1 {
			t.Errorf("expected new post count 1, got %d", stats.NewPostCount)
		}
		if stats.CommentCount != 1 {
			t.Errorf("expected comment count 1, got %d", stats.CommentCount)
		}
		if stats.LikeCount != 1 {
			t.Errorf("expected like count 1, got %d", stats.LikeCount)
		}
	})

	t.Run("from existing record", func(t *testing.T) {
		// Insert a pre-existing daily_stats row directly so that the stored
		// stat_date is a plain date string and can be matched by FindByDate in
		// the sqlite in-memory test database.
		now := time.Now()
		if err := db.Exec(
			"INSERT INTO daily_stats (stat_date, user_count, created_at, updated_at) VALUES (?, ?, ?, ?)",
			today, 100, now, now,
		).Error; err != nil {
			t.Fatalf("insert stats record failed: %v", err)
		}

		stats, err := srv.GetOverviewStats()
		if err != nil {
			t.Fatalf("get overview stats failed: %v", err)
		}
		if stats.UserCount != 100 {
			t.Errorf("expected user count 100, got %d", stats.UserCount)
		}
	})
}

func TestStatsServiceCalculateStats(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupStatsService(db)

	db.Create(&model.User{Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.Post{UserID: 1, Content: "post1", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.Comment{PostID: 1, UserID: 1, Content: "comment", CreatedAt: time.Now()})
	db.Create(&model.Like{TargetType: "post", TargetID: 1, UserID: 1, CreatedAt: time.Now()})
	db.Create(&model.Favorite{PostID: 1, UserID: 1, CreatedAt: time.Now()})
	db.Create(&model.ShareRecord{PostID: 1, UserID: 1, CreatedAt: time.Now()})
	db.Create(&model.Topic{Name: "#Topic", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.UserDevice{UserID: 1, DeviceID: "device1", LastActiveAt: time.Now()})

	t.Run("success", func(t *testing.T) {
		err := srv.CalculateStats()
		if err != nil {
			t.Fatalf("calculate stats failed: %v", err)
		}

		stats, err := srv.GetOverviewStats()
		if err != nil {
			t.Fatalf("get overview stats failed: %v", err)
		}
		if stats.UserCount != 1 {
			t.Errorf("expected user count 1, got %d", stats.UserCount)
		}
		if stats.PostCount != 1 {
			t.Errorf("expected post count 1, got %d", stats.PostCount)
		}
	})
}

func TestStatsServiceGetTrendData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupStatsService(db)

	db.Create(&model.User{Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.Post{UserID: 1, Content: "post", Status: 1, CreatedAt: time.Now()})
	db.Create(&model.UserActivityLog{UserID: 1, Action: "login", CreatedAt: time.Now()})

	t.Run("user growth", func(t *testing.T) {
		data, err := srv.GetTrendData("user_growth", 7)
		if err != nil {
			t.Fatalf("get trend data failed: %v", err)
		}
		if len(data.Data) != 7 {
			t.Errorf("expected 7 data points, got %d", len(data.Data))
		}
	})

	t.Run("post publish", func(t *testing.T) {
		data, err := srv.GetTrendData("post_publish", 7)
		if err != nil {
			t.Fatalf("get trend data failed: %v", err)
		}
		if len(data.Data) != 7 {
			t.Errorf("expected 7 data points, got %d", len(data.Data))
		}
	})

	t.Run("unknown type", func(t *testing.T) {
		data, err := srv.GetTrendData("unknown", 7)
		if err != nil {
			t.Fatalf("get trend data failed: %v", err)
		}
		if len(data.Data) != 0 {
			t.Errorf("expected 0 data points, got %d", len(data.Data))
		}
	})
}

func TestStatsServiceGetDistributionData(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupStatsService(db)

	db.Create(&model.Pet{UserID: 1, Name: "Doggo", PetType: 1, Status: 1})
	db.Create(&model.Pet{UserID: 1, Name: "Kitty", PetType: 2, Status: 1})
	db.Create(&model.User{Nickname: "BeijingUser", Phone: "beijing_phone", Email: "beijing@test.com", Status: 1, Gender: 1, Location: "Beijing"})
	db.Create(&model.User{Nickname: "ShanghaiUser", Phone: "shanghai_phone", Email: "shanghai@test.com", Status: 1, Gender: 2, Location: "Shanghai"})
	db.Create(&model.Post{UserID: 1, Content: "draft", Status: 0})
	db.Create(&model.Post{UserID: 1, Content: "published", Status: 2})
	db.Create(&model.User{Nickname: "Male", Phone: "male_phone", Email: "male@test.com", Status: 1, Gender: 1})
	db.Create(&model.User{Nickname: "Female", Phone: "female_phone", Email: "female@test.com", Status: 1, Gender: 2})

	t.Run("pet type", func(t *testing.T) {
		dist, err := srv.GetDistributionData("pet_type")
		if err != nil {
			t.Fatalf("get distribution failed: %v", err)
		}
		if len(dist.Data) != 2 {
			t.Errorf("expected 2 pet types, got %d", len(dist.Data))
		}
	})

	t.Run("user location", func(t *testing.T) {
		dist, err := srv.GetDistributionData("user_location")
		if err != nil {
			t.Fatalf("get distribution failed: %v", err)
		}
		if len(dist.Data) != 2 {
			t.Errorf("expected 2 locations, got %d", len(dist.Data))
		}
	})

	t.Run("post status", func(t *testing.T) {
		dist, err := srv.GetDistributionData("post_status")
		if err != nil {
			t.Fatalf("get distribution failed: %v", err)
		}
		if len(dist.Data) != 2 {
			t.Errorf("expected 2 post statuses, got %d", len(dist.Data))
		}
	})

	t.Run("user gender", func(t *testing.T) {
		dist, err := srv.GetDistributionData("user_gender")
		if err != nil {
			t.Fatalf("get distribution failed: %v", err)
		}
		if len(dist.Data) != 2 {
			t.Errorf("expected 2 genders, got %d", len(dist.Data))
		}
	})
}

func TestStatsServiceExportStatsToExcel(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupStatsService(db)

	db.Create(&model.User{Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1, CreatedAt: time.Now()})

	t.Run("overview", func(t *testing.T) {
		data, err := srv.ExportStatsToExcel("overview", "", "")
		if err != nil {
			t.Fatalf("export overview failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty excel data")
		}
	})

	t.Run("user growth", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		todayDate, _ := time.Parse("2006-01-02", today)
		db.Create(&model.Statistics{
			StatDate:        todayDate,
			NewUserCount:    1,
			UserCount:       1,
			ActiveUserCount: 1,
			CreatedAt:       time.Now(),
		})
		data, err := srv.ExportStatsToExcel("user_growth", today, today)
		if err != nil {
			t.Fatalf("export user growth failed: %v", err)
		}
		if len(data) == 0 {
			t.Error("expected non-empty excel data")
		}
	})
}
