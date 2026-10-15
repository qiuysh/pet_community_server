package testutil

import (
	"testing"

	"pet-community/config"
	"pet-community/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an isolated in-memory sqlite database and migrates all models.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use an isolated in-memory database per test. With a single open connection,
	// GORM reuses the same connection for the duration of the test, keeping data
	// alive while preventing cross-test pollution.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Restrict to a single connection so the in-memory database stays alive
	// for the duration of the test.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get underlying sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&model.User{},
		&model.UserDevice{},
		&model.VerificationCode{},
		&model.UserThirdAccount{},
		&model.AdminUser{},
		&model.AdminLoginLog{},
		&model.AdminOperationLog{},
		&model.SystemConfig{},
		&model.SensitiveWord{},
		&model.SystemNotification{},
		&model.Announcement{},
		&model.UserNotification{},
		&model.Pet{},
		&model.PetBreed{},
		&model.PetAlbum{},
		&model.PetGrowthRecord{},
		&model.LostPet{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.Favorite{},
		&model.ShareRecord{},
		&model.Draft{},
		&model.Report{},
		&model.Topic{},
		&model.Follow{},
		&model.Block{},
		&model.Conversation{},
		&model.Message{},
		&model.UserLevel{},
		&model.PointsRecord{},
		&model.PointsRule{},
		&model.LevelBenefit{},
		&model.NotificationSettings{},
		&model.SearchHistory{},
		&model.UserInterest{},
		&model.TopicCategory{},
		&model.TopicParticipant{},
		&model.HotKeyword{},
		&model.Activity{},
		&model.Registration{},
		&model.Submission{},
		&model.Vote{},
		&model.Adopt{},
		&model.AdoptionApplication{},
		&model.AdPosition{},
		&model.Advertisement{},
		&model.Statistics{},
		&model.UserActivityLog{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	if config.Cfg == nil {
		config.Cfg = &config.Config{}
	}
	config.Cfg.JWT.Secret = "test_secret_key_for_testing"

	return db
}

// SetGlobalDB sets config.DB to the provided test database.
func SetGlobalDB(db *gorm.DB) {
	config.DB = db
}
