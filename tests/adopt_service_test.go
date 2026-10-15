package tests

import (
	"testing"

	"gorm.io/gorm"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

func setupAdoptService(db *gorm.DB) *service.AdoptService {
	adoptRepo := repository.NewAdoptRepository(db)
	applicationRepo := repository.NewAdoptionApplicationRepository(db)
	followUpRepo := repository.NewAdoptionFollowUpRepository(db)
	userRepo := repository.NewUserRepository(db)
	return service.NewAdoptService(adoptRepo, applicationRepo, followUpRepo, userRepo)
}

func createTestUser(db *gorm.DB, id uint, nickname string) *model.User {
	user := &model.User{
		ID:       id,
		Nickname: nickname,
		Phone:    nickname + "_phone",
		Email:    nickname + "@test.com",
		Status:   1,
	}
	db.Create(user)
	return user
}

func createTestAdopt(db *gorm.DB, userID uint) *model.Adopt {
	adopt := &model.Adopt{
		UserID:       userID,
		PetName:      "Kitty",
		PetType:      "cat",
		PetAge:       "1 year",
		PetGender:    1,
		HealthStatus: "healthy",
		Photos:       "[\"a.jpg\"]",
		Description:  "cute cat",
		Location:     "Beijing",
		ContactPhone: "13800138000",
		Status:       0,
	}
	db.Create(adopt)
	return adopt
}

func TestAdoptServiceCreateAdopt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")

	t.Run("success", func(t *testing.T) {
		info, err := srv.CreateAdopt(1, &types.AdoptCreateRequest{
			PetName:      "Kitty",
			PetType:      "cat",
			PetAge:       "1 year",
			PetGender:    1,
			HealthStatus: "healthy",
			Photos:       []string{"a.jpg"},
			Description:  "cute cat",
			Location:     "Beijing",
			ContactPhone: "13800138000",
		})
		if err != nil {
			t.Fatalf("create adopt failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("adopt ID should not be zero")
		}
		if info.Status != 0 {
			t.Errorf("expected status 0, got %d", info.Status)
		}
	})

	t.Run("reject placeholder URLs in photos", func(t *testing.T) {
		_, err := srv.CreateAdopt(1, &types.AdoptCreateRequest{
			PetName:      "Kitty",
			PetType:      "cat",
			PetAge:       "1 year",
			PetGender:    1,
			HealthStatus: "healthy",
			Photos:       []string{"placeholder-url-0", "placeholder-url-1"},
			Description:  "cute cat",
			Location:     "Beijing",
			ContactPhone: "13800138000",
		})
		if err == nil {
			t.Fatal("expected error for placeholder URLs but got nil")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("reject blob URLs in photos (unuploaded local preview)", func(t *testing.T) {
		_, err := srv.CreateAdopt(1, &types.AdoptCreateRequest{
			PetName:      "Kitty",
			PetType:      "cat",
			PetAge:       "1 year",
			PetGender:    1,
			HealthStatus: "healthy",
			Photos:       []string{"blob:http://localhost:5173/abcd-1234"},
			Description:  "cute cat",
			Location:     "Beijing",
			ContactPhone: "13800138000",
		})
		if err == nil {
			t.Fatal("expected error for blob URLs but got nil")
		}
	})

	t.Run("allow normal http(s) URLs in photos", func(t *testing.T) {
		info, err := srv.CreateAdopt(1, &types.AdoptCreateRequest{
			PetName:      "Kitty",
			PetType:      "cat",
			PetAge:       "1 year",
			PetGender:    1,
			HealthStatus: "healthy",
			Photos:       []string{"https://cdn.example.com/photo1.jpg", "/uploads/adopt/xyz.jpg"},
			Description:  "cute cat",
			Location:     "Beijing",
			ContactPhone: "13800138000",
		})
		if err != nil {
			t.Fatalf("expected success for real URLs, got: %v", err)
		}
		if info.ID == 0 {
			t.Error("adopt ID should not be zero")
		}
	})
}

func TestAdoptServiceGetAdoptByID(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	adopt := createTestAdopt(db, 1)

	t.Run("success", func(t *testing.T) {
		info, err := srv.GetAdoptByID(adopt.ID, 1)
		if err != nil {
			t.Fatalf("get adopt failed: %v", err)
		}
		if info.ID != adopt.ID {
			t.Errorf("expected ID %d, got %d", adopt.ID, info.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.GetAdoptByID(9999, 0)
		if err == nil {
			t.Error("expected error for non-existent adopt")
		}
	})
}

func TestAdoptServiceUpdateAdopt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	adopt := createTestAdopt(db, 1)

	t.Run("success", func(t *testing.T) {
		info, err := srv.UpdateAdopt(1, adopt.ID, &types.AdoptUpdateRequest{
			PetName: "New Kitty",
			Status:  1,
		})
		if err != nil {
			t.Fatalf("update adopt failed: %v", err)
		}
		if info.PetName != "New Kitty" {
			t.Errorf("expected name 'New Kitty', got %s", info.PetName)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		createTestUser(db, 2, "Bob")
		_, err := srv.UpdateAdopt(2, adopt.ID, &types.AdoptUpdateRequest{
			PetName: "Hacked",
		})
		if err == nil {
			t.Error("expected error for no permission")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := srv.UpdateAdopt(1, 9999, &types.AdoptUpdateRequest{
			PetName: "Ghost",
		})
		if err == nil {
			t.Error("expected error for non-existent adopt")
		}
	})
}

func TestAdoptServiceDeleteAdopt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	adopt := createTestAdopt(db, 1)

	t.Run("success", func(t *testing.T) {
		err := srv.DeleteAdopt(1, adopt.ID)
		if err != nil {
			t.Fatalf("delete adopt failed: %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		createTestUser(db, 2, "Bob")
		adopt2 := createTestAdopt(db, 1)
		err := srv.DeleteAdopt(2, adopt2.ID)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestAdoptServiceListAdopts(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	createTestAdopt(db, 1)
	createTestAdopt(db, 1)

	t.Run("list all", func(t *testing.T) {
		resp, err := srv.ListAdopts(&types.AdoptListRequest{Page: 1, PageSize: 10}, 0)
		if err != nil {
			t.Fatalf("list adopts failed: %v", err)
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		resp, err := srv.ListAdopts(&types.AdoptListRequest{PetType: "cat", Page: 1, PageSize: 10}, 1)
		if err != nil {
			t.Fatalf("list adopts failed: %v", err)
		}
		if resp.Total != 2 {
			t.Errorf("expected total 2, got %d", resp.Total)
		}
	})
}

func TestAdoptServiceCreateApplication(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	createTestUser(db, 2, "Bob")
	adopt := createTestAdopt(db, 1)

	t.Run("success", func(t *testing.T) {
		info, err := srv.CreateApplication(2, &types.AdoptionApplicationRequest{
			AdoptionID:      adopt.ID,
			ApplicationInfo: "I love cats",
		})
		if err != nil {
			t.Fatalf("create application failed: %v", err)
		}
		if info.ID == 0 {
			t.Error("application ID should not be zero")
		}
	})

	t.Run("apply own adopt", func(t *testing.T) {
		_, err := srv.CreateApplication(1, &types.AdoptionApplicationRequest{
			AdoptionID:      adopt.ID,
			ApplicationInfo: "I love cats",
		})
		if err == nil {
			t.Error("expected error for applying own adopt")
		}
	})

	t.Run("duplicate application", func(t *testing.T) {
		_, err := srv.CreateApplication(2, &types.AdoptionApplicationRequest{
			AdoptionID:      adopt.ID,
			ApplicationInfo: "Again",
		})
		if err == nil {
			t.Error("expected error for duplicate application")
		}
	})
}

func TestAdoptServiceReviewApplication(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	createTestUser(db, 2, "Bob")
	adopt := createTestAdopt(db, 1)
	application, _ := srv.CreateApplication(2, &types.AdoptionApplicationRequest{
		AdoptionID:      adopt.ID,
		ApplicationInfo: "I love cats",
	})

	t.Run("approve", func(t *testing.T) {
		err := srv.ReviewApplication(1, &types.AdoptionReviewRequest{
			ApplicationID: application.ID,
			Approve:       true,
		})
		if err != nil {
			t.Fatalf("review application failed: %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		err := srv.ReviewApplication(2, &types.AdoptionReviewRequest{
			ApplicationID: application.ID,
			Approve:       true,
		})
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestAdoptServiceUpdateAdoptStatus(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	adopt := createTestAdopt(db, 1)

	t.Run("success", func(t *testing.T) {
		err := srv.UpdateAdoptStatus(1, adopt.ID, 2)
		if err != nil {
			t.Fatalf("update status failed: %v", err)
		}
	})

	t.Run("no permission", func(t *testing.T) {
		createTestUser(db, 2, "Bob")
		err := srv.UpdateAdoptStatus(2, adopt.ID, 1)
		if err == nil {
			t.Error("expected error for no permission")
		}
	})
}

func TestAdoptServiceGetAdoptionStats(t *testing.T) {
	db := testutil.SetupTestDB(t)
	srv := setupAdoptService(db)

	createTestUser(db, 1, "Alice")
	createTestUser(db, 2, "Bob")
	adopt1 := createTestAdopt(db, 1)
	adopt2 := createTestAdopt(db, 1)
	adopt3 := createTestAdopt(db, 1)

	// Mark one as adopted and one as cancelled.
	srv.UpdateAdoptStatus(1, adopt1.ID, 1)
	srv.UpdateAdoptStatus(1, adopt2.ID, 2)

	// Bob applies for Alice's adoption (must apply to pending adoption with status=0)
	srv.CreateApplication(2, &types.AdoptionApplicationRequest{
		AdoptionID:      adopt3.ID,
		ApplicationInfo: "info",
	})

	t.Run("stats", func(t *testing.T) {
		stats, err := srv.GetAdoptionStats()
		if err != nil {
			t.Fatalf("get stats failed: %v", err)
		}
		if stats.TotalAdoptions != 3 {
			t.Errorf("expected total adoptions 3, got %d", stats.TotalAdoptions)
		}
		if stats.AdoptedAdoptions != 1 {
			t.Errorf("expected adopted 1, got %d", stats.AdoptedAdoptions)
		}
		if stats.CancelledAdoptions != 1 {
			t.Errorf("expected cancelled 1, got %d", stats.CancelledAdoptions)
		}
		if stats.TotalApplications != 1 {
			t.Errorf("expected total applications 1, got %d", stats.TotalApplications)
		}
	})
}
