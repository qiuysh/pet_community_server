package tests

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

func createActiveActivity(db *gorm.DB) *model.Activity {
	activity := &model.Activity{
		Title:           "Test Activity",
		Description:     "desc",
		Type:            "voting",
		Status:          1,
		StartTime:       time.Now().Add(-time.Hour),
		EndTime:         time.Now().Add(time.Hour),
		MaxParticipants: 10,
		MaxSubmissions:  2,
		VotesPerDay:     3,
	}
	db.Create(activity)
	return activity
}

func TestActivityServiceCreateActivity(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewActivityService(activityRepo, registrationRepo, submissionRepo, voteRepo, userRepo)

	now := time.Now()

	t.Run("success", func(t *testing.T) {
		_, err := srv.CreateActivity(&types.ActivityCreateRequest{
			Title:       "New Activity",
			Description: "desc",
			Type:        "voting",
			StartTime:   now.Add(time.Hour).Format(time.RFC3339),
			EndTime:     now.Add(2 * time.Hour).Format(time.RFC3339),
		})
		if err != nil {
			t.Fatalf("create activity failed: %v", err)
		}
	})

	t.Run("invalid time format", func(t *testing.T) {
		_, err := srv.CreateActivity(&types.ActivityCreateRequest{
			Title:       "Bad",
			Description: "desc",
			Type:        "voting",
			StartTime:   "not-a-time",
			EndTime:     now.Add(2 * time.Hour).Format(time.RFC3339),
		})
		if err == nil {
			t.Error("expected error for invalid start time")
		}
	})

	t.Run("end before start", func(t *testing.T) {
		_, err := srv.CreateActivity(&types.ActivityCreateRequest{
			Title:       "Bad",
			Description: "desc",
			Type:        "voting",
			StartTime:   now.Add(2 * time.Hour).Format(time.RFC3339),
			EndTime:     now.Add(time.Hour).Format(time.RFC3339),
		})
		if err == nil {
			t.Error("expected error for end before start")
		}
	})

	t.Run("too long duration", func(t *testing.T) {
		_, err := srv.CreateActivity(&types.ActivityCreateRequest{
			Title:       "Bad",
			Description: "desc",
			Type:        "voting",
			StartTime:   now.Format(time.RFC3339),
			EndTime:     now.AddDate(0, 0, 31).Format(time.RFC3339),
		})
		if err == nil {
			t.Error("expected error for duration over 30 days")
		}
	})
}

func TestActivityServicePublishAndGet(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewActivityService(activityRepo, registrationRepo, submissionRepo, voteRepo, userRepo)

	info, err := srv.CreateActivity(&types.ActivityCreateRequest{
		Title:       "Draft",
		Description: "desc",
		Type:        "voting",
		StartTime:   time.Now().Format(time.RFC3339),
		EndTime:     time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("create activity failed: %v", err)
	}

	t.Run("publish", func(t *testing.T) {
		err := srv.PublishActivity(info.ID)
		if err != nil {
			t.Fatalf("publish activity failed: %v", err)
		}
	})

	t.Run("get by id", func(t *testing.T) {
		got, err := srv.GetActivityByID(0, info.ID)
		if err != nil {
			t.Fatalf("get activity failed: %v", err)
		}
		if got.Status != 1 {
			t.Errorf("expected status 1, got %d", got.Status)
		}
	})

	t.Run("list", func(t *testing.T) {
		resp, err := srv.ListActivities(0, "", "", "", 1, 10)
		if err != nil {
			t.Fatalf("list activities failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
	})
}

func TestRegistrationService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewRegistrationService(registrationRepo, activityRepo, userRepo)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.User{ID: 2, Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1})

	activity := createActiveActivity(db)

	t.Run("success", func(t *testing.T) {
		_, err := srv.Register(1, &types.RegistrationRequest{
			ActivityID: activity.ID,
			Name:       "Alice",
			Phone:      "13800138000",
		})
		if err != nil {
			t.Fatalf("register failed: %v", err)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		_, err := srv.Register(1, &types.RegistrationRequest{
			ActivityID: activity.ID,
			Name:       "Alice",
			Phone:      "13800138000",
		})
		if err == nil {
			t.Error("expected error for duplicate registration")
		}
	})

	t.Run("max participants", func(t *testing.T) {
		activity.MaxParticipants = 1
		activityRepo.Update(activity)
		_, err := srv.Register(2, &types.RegistrationRequest{
			ActivityID: activity.ID,
			Name:       "Bob",
			Phone:      "13800138001",
		})
		if err == nil {
			t.Error("expected error for max participants")
		}
	})
}

func TestSubmissionService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	userRepo := repository.NewUserRepository(db)
	voteRepo := repository.NewVoteRepository(db)
	srv := service.NewSubmissionService(submissionRepo, activityRepo, registrationRepo, userRepo, voteRepo)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	activity := createActiveActivity(db)
	registrationRepo.Create(&model.Registration{ActivityID: activity.ID, UserID: 1, Name: "Alice", Phone: "1", Status: 1})

	t.Run("success", func(t *testing.T) {
		_, err := srv.CreateSubmission(1, &types.SubmissionCreateRequest{
			ActivityID: activity.ID,
			Title:      "My Work",
			Type:       "image",
			Images:     []string{"a.jpg"},
		})
		if err != nil {
			t.Fatalf("create submission failed: %v", err)
		}
	})

	t.Run("not registered", func(t *testing.T) {
		_, err := srv.CreateSubmission(2, &types.SubmissionCreateRequest{
			ActivityID: activity.ID,
			Title:      "Hack",
			Type:       "text",
		})
		if err == nil {
			t.Error("expected error for not registered")
		}
	})

	t.Run("image required", func(t *testing.T) {
		_, err := srv.CreateSubmission(1, &types.SubmissionCreateRequest{
			ActivityID: activity.ID,
			Title:      "No Image",
			Type:       "image",
		})
		if err == nil {
			t.Error("expected error for missing image")
		}
	})

	t.Run("list", func(t *testing.T) {
		resp, err := srv.ListSubmissions(1, activity.ID, "votes", 1, 10)
		if err != nil {
			t.Fatalf("list submissions failed: %v", err)
		}
		if resp.Total != 1 {
			t.Errorf("expected total 1, got %d", resp.Total)
		}
	})
}

func TestVoteService(t *testing.T) {
	db := testutil.SetupTestDB(t)
	voteRepo := repository.NewVoteRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	userRepo := repository.NewUserRepository(db)
	srv := service.NewVoteService(voteRepo, submissionRepo, activityRepo, userRepo)

	db.Create(&model.User{ID: 1, Nickname: "Alice", Phone: "alice_phone", Email: "alice@test.com", Status: 1})
	db.Create(&model.User{ID: 2, Nickname: "Bob", Phone: "bob_phone", Email: "bob@test.com", Status: 1})
	activity := createActiveActivity(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	registrationRepo.Create(&model.Registration{ActivityID: activity.ID, UserID: 1, Name: "Alice", Phone: "1", Status: 1})
	registrationRepo.Create(&model.Registration{ActivityID: activity.ID, UserID: 2, Name: "Bob", Phone: "2", Status: 1})
	submission := &model.Submission{ActivityID: activity.ID, UserID: 1, Title: "Work", Type: "text", Status: 1}
	db.Create(submission)

	t.Run("success", func(t *testing.T) {
		result, err := srv.Vote(2, &types.VoteRequest{
			ActivityID:   activity.ID,
			SubmissionID: submission.ID,
		})
		if err != nil {
			t.Fatalf("vote failed: %v", err)
		}
		if !result.Voted {
			t.Error("expected voted")
		}
	})

	t.Run("duplicate vote", func(t *testing.T) {
		_, err := srv.Vote(2, &types.VoteRequest{
			ActivityID:   activity.ID,
			SubmissionID: submission.ID,
		})
		if err == nil {
			t.Error("expected error for duplicate vote")
		}
	})

	t.Run("daily limit", func(t *testing.T) {
		activity.VotesPerDay = 1
		activityRepo.Update(activity)
		submission2 := &model.Submission{ActivityID: activity.ID, UserID: 1, Title: "Work2", Type: "text", Status: 1}
		db.Create(submission2)
		// User 2 already voted once for submission1; voting again should exceed the daily limit.
		_, err := srv.Vote(2, &types.VoteRequest{
			ActivityID:   activity.ID,
			SubmissionID: submission2.ID,
		})
		if err == nil {
			t.Error("expected error for daily limit")
		}
	})
}
