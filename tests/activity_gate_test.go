package tests

import (
	"strings"
	"testing"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"

	"gorm.io/gorm"
)

// createOfflineActivity 构造一个进行中的线下活动（签到/相册仅对此类型开放）
func createOfflineActivity(db *gorm.DB) *model.Activity {
	activity := &model.Activity{
		Title:           "线下见面会",
		Description:     "desc",
		Type:            "offline_meet",
		Status:          1,
		StartTime:       time.Now().Add(-time.Hour),
		EndTime:         time.Now().Add(time.Hour),
		MaxParticipants: 10,
	}
	db.Create(activity)
	return activity
}

func createGateTestUser(db *gorm.DB, phone, email, nickname string) *model.User {
	user := &model.User{Phone: phone, Email: email, Nickname: nickname, Status: 1}
	db.Create(user)
	return user
}

// TestSignInRequiresRegistration 未报名用户不能扫码签到，报名后才可签到。
func TestSignInRequiresRegistration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	srv := service.NewSignInService(
		repository.NewSignInRepository(db), activityRepo, registrationRepo, repository.NewUserRepository(db),
	)

	activity := createOfflineActivity(db)
	user := createGateTestUser(db, "13900139001", "signin@test.dev", "signer")

	qr, err := srv.GenerateQRCode(activity.ID)
	if err != nil {
		t.Fatalf("生成签到二维码失败: %v", err)
	}

	_, err = srv.SignIn(user.ID, &types.SignInRequest{ActivityID: activity.ID, QRCode: qr.QRCode})
	if err == nil {
		t.Fatal("未报名用户不应签到成功")
	}
	if !strings.Contains(err.Error(), "报名") {
		t.Fatalf("错误信息应提示先报名，实际: %v", err)
	}

	if err := registrationRepo.Create(&model.Registration{
		ActivityID: activity.ID, UserID: user.ID, Name: "报名者", Phone: "13900139001", Status: 1,
	}); err != nil {
		t.Fatalf("创建报名记录失败: %v", err)
	}

	if _, err := srv.SignIn(user.ID, &types.SignInRequest{ActivityID: activity.ID, QRCode: qr.QRCode}); err != nil {
		t.Fatalf("已报名用户应能签到，实际报错: %v", err)
	}
}

// TestUploadAlbumPhotoRequiresRegistration 未报名用户不能往活动相册上传照片。
func TestUploadAlbumPhotoRequiresRegistration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)
	srv := service.NewAlbumService(
		repository.NewActivityAlbumRepository(db), activityRepo, registrationRepo, repository.NewUserRepository(db),
	)

	activity := createOfflineActivity(db)
	user := createGateTestUser(db, "13900139002", "album@test.dev", "photographer")

	_, err := srv.UploadPhoto(user.ID, &types.AlbumPhotoCreateRequest{
		ActivityID: activity.ID, PhotoURL: "/api/v1/uploads/album/1.jpg", Description: "现场",
	})
	if err == nil {
		t.Fatal("未报名用户不应上传相册照片")
	}
	if !strings.Contains(err.Error(), "报名") {
		t.Fatalf("错误信息应提示先报名，实际: %v", err)
	}

	if err := registrationRepo.Create(&model.Registration{
		ActivityID: activity.ID, UserID: user.ID, Name: "报名者", Phone: "13900139002", Status: 1,
	}); err != nil {
		t.Fatalf("创建报名记录失败: %v", err)
	}

	if _, err := srv.UploadPhoto(user.ID, &types.AlbumPhotoCreateRequest{
		ActivityID: activity.ID, PhotoURL: "/api/v1/uploads/album/1.jpg", Description: "现场",
	}); err != nil {
		t.Fatalf("已报名用户应能上传相册照片，实际报错: %v", err)
	}
}

// TestPublishAwardResultNotifiesWinners 发布评选结果要给获奖用户发站内通知。
func TestPublishAwardResultNotifiesWinners(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)
	userRepo := repository.NewUserRepository(db)

	srv := service.NewSubmissionService(service.SubmissionServiceDeps{
		SubmissionRepo:   submissionRepo,
		ActivityRepo:     activityRepo,
		RegistrationRepo: repository.NewRegistrationRepository(db),
		UserRepo:         userRepo,
		VoteRepo:         repository.NewVoteRepository(db),
		NotificationRepo: repository.NewNotificationRepository(db),
	})

	activity := createActiveActivity(db)
	winner := createGateTestUser(db, "13900139003", "winner@test.dev", "winner")

	submission := &model.Submission{
		ActivityID: activity.ID, UserID: winner.ID, Title: "获奖作品", Type: "image", Status: 1,
	}
	if err := db.Create(submission).Error; err != nil {
		t.Fatalf("创建作品失败: %v", err)
	}
	if err := submissionRepo.SetAward(submission.ID, true, "first"); err != nil {
		t.Fatalf("设置获奖失败: %v", err)
	}

	resp, err := srv.PublishAwardResult(activity.ID)
	if err != nil {
		t.Fatalf("发布评选结果失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("发布应成功，实际返回: %+v", resp)
	}
	if len(resp.NotifiedUsers) != 1 || resp.NotifiedUsers[0] != winner.ID {
		t.Fatalf("应通知 1 位获奖用户，实际 %v", resp.NotifiedUsers)
	}

	var notifications []model.UserNotification
	db.Where("user_id = ?", winner.ID).Find(&notifications)
	if len(notifications) != 1 {
		t.Fatalf("应为获奖者生成 1 条通知，实际 %d 条", len(notifications))
	}
	if !strings.Contains(notifications[0].Content, "一等奖") {
		t.Fatalf("通知内容应包含奖项等级，实际: %s", notifications[0].Content)
	}
	if !strings.Contains(notifications[0].Content, activity.Title) {
		t.Fatalf("通知内容应包含活动标题，实际: %s", notifications[0].Content)
	}
	if notifications[0].Type != "activity" {
		t.Fatalf("通知类型应为 activity，实际: %s", notifications[0].Type)
	}
}

// TestPublishAwardResultRejectsRepublish 重复发布应被拒绝（避免重复通知获奖者）。
func TestPublishAwardResultRejectsRepublish(t *testing.T) {
	db := testutil.SetupTestDB(t)
	activityRepo := repository.NewActivityRepository(db)
	submissionRepo := repository.NewSubmissionRepository(db)

	srv := service.NewSubmissionService(service.SubmissionServiceDeps{
		SubmissionRepo:   submissionRepo,
		ActivityRepo:     activityRepo,
		RegistrationRepo: repository.NewRegistrationRepository(db),
		UserRepo:         repository.NewUserRepository(db),
		VoteRepo:         repository.NewVoteRepository(db),
		NotificationRepo: repository.NewNotificationRepository(db),
	})

	activity := createActiveActivity(db)
	winner := createGateTestUser(db, "13900139004", "winner2@test.dev", "winner2")
	submission := &model.Submission{
		ActivityID: activity.ID, UserID: winner.ID, Title: "作品", Type: "image", Status: 1,
	}
	if err := db.Create(submission).Error; err != nil {
		t.Fatalf("创建作品失败: %v", err)
	}
	if err := submissionRepo.SetAward(submission.ID, true, "second"); err != nil {
		t.Fatalf("设置获奖失败: %v", err)
	}

	if _, err := srv.PublishAwardResult(activity.ID); err != nil {
		t.Fatalf("首次发布失败: %v", err)
	}
	if _, err := srv.PublishAwardResult(activity.ID); err == nil {
		t.Fatal("已发布评选结果的活动不应允许再次发布")
	}

	var count int64
	db.Model(&model.UserNotification{}).Where("user_id = ?", winner.ID).Count(&count)
	if count != 1 {
		t.Fatalf("重复发布不应重复通知，实际 %d 条", count)
	}
}
