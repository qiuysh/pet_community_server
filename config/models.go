package config

import "pet-community/model"

// Models 返回由 ORM 管理的全部表模型，是「哪些表存在」的唯一代码来源。
//
// 新增模型时必须同时做两件事，缺一不可：
//  1. 加入本列表（AutoMigrate 才会建表/补列）；
//  2. 在 sql/init/create.sql 里同步建表与补列（只跑 SQL 的环境不做 AutoMigrate，
//     模型有列而建表脚本没有 → 运行时 1054 Unknown column）。
//
// tests/schema_drift_test.go 会拿本列表与 create.sql 逐列比对，漏了会直接测试失败。
func Models() []interface{} {
	return []interface{}{
		&model.AdminUser{},
		&model.AdminOperationLog{},
		&model.AdminLoginLog{},
		&model.SystemConfig{},
		&model.SensitiveWord{},
		&model.UserActivityLog{},
		&model.UserBehaviorLog{},
		&model.User{},
		&model.UserDevice{},
		&model.VerificationCode{},
		&model.UserThirdAccount{},
		&model.SystemNotification{},
		&model.Announcement{},
		&model.Pet{},
		&model.PetBreed{},
		&model.PetAlbum{},
		&model.PetGrowthRecord{},
		&model.LostPet{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.Favorite{},
		&model.Draft{},
		&model.Report{},
		&model.Topic{},
		&model.ShareRecord{},
		&model.UserNotification{},
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
		&model.SignIn{},
		&model.ActivityAlbum{},
		&model.ActivityReview{},
		&model.Adopt{},
		&model.AdoptionApplication{},
		&model.AdoptionFollowUp{},
		&model.Statistics{},
		&model.AdPosition{},
		&model.Advertisement{},
	}
}
