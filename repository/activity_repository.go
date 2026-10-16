package repository

import (
	"pet-community/model"
	"time"

	"gorm.io/gorm"
)

// ActivityRepository 活动仓储
type ActivityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

func (r *ActivityRepository) Create(activity *model.Activity) error {
	return r.db.Create(activity).Error
}

func (r *ActivityRepository) Update(activity *model.Activity) error {
	return r.db.Save(activity).Error
}

func (r *ActivityRepository) GetByID(id uint) (*model.Activity, error) {
	var activity model.Activity
	if err := r.db.First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) Delete(id uint) error {
	return r.db.Delete(&model.Activity{}, id).Error
}

func (r *ActivityRepository) List(activityType, status, keyword string, page, pageSize int, includeDraft bool) ([]model.Activity, int64, error) {
	var activities []model.Activity
	var total int64
	query := r.db.Model(&model.Activity{})

	if activityType != "" {
		query = query.Where("type = ?", activityType)
	}

	if status != "" {
		if status == "active" {
			query = query.Where("status = 1 AND start_time <= ? AND end_time >= ?", time.Now(), time.Now())
		} else if status == "upcoming" {
			query = query.Where("status = 1 AND start_time > ?", time.Now())
		} else if status == "ended" {
			query = query.Where("status = 1 AND end_time < ?", time.Now())
		} else if status == "draft" {
			query = query.Where("status = 0")
		} else if status == "published" {
			query = query.Where("status = 1")
		}
	} else if !includeDraft {
		// 前台默认不暴露草稿
		query = query.Where("status <> 0")
	}

	if keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

func (r *ActivityRepository) GetActiveActivities(limit int) ([]model.Activity, error) {
	var activities []model.Activity
	if err := r.db.Where("status = 1 AND start_time <= ? AND end_time >= ?", time.Now(), time.Now()).Limit(limit).Order("id DESC").Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}



// RegistrationRepository 报名仓储
type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Create(registration *model.Registration) error {
	return r.db.Create(registration).Error
}

func (r *RegistrationRepository) Update(registration *model.Registration) error {
	return r.db.Save(registration).Error
}

func (r *RegistrationRepository) GetByID(id uint) (*model.Registration, error) {
	var registration model.Registration
	if err := r.db.First(&registration, id).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *RegistrationRepository) GetByActivityAndUser(activityID, userID uint) (*model.Registration, error) {
	var registration model.Registration
	if err := r.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *RegistrationRepository) ListByActivity(activityID uint, status int, page, pageSize int) ([]model.Registration, int64, error) {
	var registrations []model.Registration
	var total int64
	query := r.db.Model(&model.Registration{}).Where("activity_id = ?", activityID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Lv8+ Priority 优先，同优先级按报名先后
	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).
		Order("priority DESC, id ASC").Find(&registrations).Error; err != nil {
		return nil, 0, err
	}

	return registrations, total, nil
}

func (r *RegistrationRepository) Exists(activityID, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Registration{}).Where("activity_id = ? AND user_id = ?", activityID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RegistrationRepository) CountByActivity(activityID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Registration{}).Where("activity_id = ?", activityID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Transaction 执行事务，由仓储封装内部 db 实例
func (r *RegistrationRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// SubmissionRepository 作品仓储
type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) Create(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

func (r *SubmissionRepository) Update(submission *model.Submission) error {
	return r.db.Save(submission).Error
}

func (r *SubmissionRepository) GetByID(id uint) (*model.Submission, error) {
	var submission model.Submission
	if err := r.db.First(&submission, id).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) Delete(id uint) error {
	return r.db.Delete(&model.Submission{}, id).Error
}

func (r *SubmissionRepository) ListByActivity(activityID uint, sort string, page, pageSize int) ([]model.Submission, int64, error) {
	var submissions []model.Submission
	var total int64
	query := r.db.Model(&model.Submission{}).Where("activity_id = ? AND status = 1", activityID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderStr := "id DESC"
	if sort == "votes" {
		orderStr = "votes_count DESC, id DESC"
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order(orderStr).Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, total, nil
}

func (r *SubmissionRepository) ListByUser(userID uint, page, pageSize int) ([]model.Submission, int64, error) {
	var submissions []model.Submission
	var total int64
	query := r.db.Model(&model.Submission{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&submissions).Error; err != nil {
		return nil, 0, err
	}

	return submissions, total, nil
}

func (r *SubmissionRepository) GetByActivityAndUser(activityID, userID uint) (*model.Submission, error) {
	var submission model.Submission
	if err := r.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&submission).Error; err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *SubmissionRepository) CountByActivity(activityID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Submission{}).Where("activity_id = ? AND status = 1", activityID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SubmissionRepository) CountByActivityAndUser(activityID, userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.Submission{}).Where("activity_id = ? AND user_id = ?", activityID, userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SubmissionRepository) IncrementVotes(id uint) error {
	return r.db.Model(&model.Submission{}).Where("id = ?", id).UpdateColumn("votes_count", gorm.Expr("votes_count + 1")).Error
}

func (r *SubmissionRepository) DecrementVotes(id uint) error {
	return r.db.Model(&model.Submission{}).Where("id = ?", id).UpdateColumn("votes_count", gorm.Expr("votes_count - 1")).Error
}

func (r *SubmissionRepository) SetAward(id uint, isAwarded bool, awardLevel string) error {
	return r.db.Model(&model.Submission{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_awarded":  isAwarded,
		"award_level": awardLevel,
	}).Error
}



// VoteRepository 投票仓储
type VoteRepository struct {
	db *gorm.DB
}

func NewVoteRepository(db *gorm.DB) *VoteRepository {
	return &VoteRepository{db: db}
}

func (r *VoteRepository) Create(vote *model.Vote) error {
	return r.db.Create(vote).Error
}

func (r *VoteRepository) Delete(id uint) error {
	return r.db.Delete(&model.Vote{}, id).Error
}

func (r *VoteRepository) GetByID(id uint) (*model.Vote, error) {
	var vote model.Vote
	if err := r.db.First(&vote, id).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *VoteRepository) GetByUserAndSubmission(userID, submissionID uint) (*model.Vote, error) {
	var vote model.Vote
	if err := r.db.Where("user_id = ? AND submission_id = ?", userID, submissionID).First(&vote).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *VoteRepository) Exists(userID, submissionID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Vote{}).Where("user_id = ? AND submission_id = ?", userID, submissionID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsToday 检查用户今天是否给某作品投过票
func (r *VoteRepository) ExistsToday(userID, submissionID uint) (bool, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	todayStart, _ := time.Parse("2006-01-02", today)
	todayEnd := todayStart.Add(24 * time.Hour)

	if err := r.db.Model(&model.Vote{}).Where("user_id = ? AND submission_id = ? AND created_at >= ? AND created_at < ?", userID, submissionID, todayStart, todayEnd).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *VoteRepository) CountByUserToday(userID uint) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	todayStart, _ := time.Parse("2006-01-02", today)
	todayEnd := todayStart.Add(24 * time.Hour)

	if err := r.db.Model(&model.Vote{}).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, todayStart, todayEnd).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountByUserInHour 统计用户最近1小时内投票数
func (r *VoteRepository) CountByUserInHour(userID uint) (int64, error) {
	var count int64
	hourAgo := time.Now().Add(-1 * time.Hour)

	if err := r.db.Model(&model.Vote{}).Where("user_id = ? AND created_at >= ?", userID, hourAgo).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *VoteRepository) ListByUser(userID uint, activityID uint, page, pageSize int) ([]model.Vote, int64, error) {
	var votes []model.Vote
	var total int64
	query := r.db.Model(&model.Vote{}).Where("user_id = ?", userID)

	if activityID > 0 {
		query = query.Where("activity_id = ?", activityID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&votes).Error; err != nil {
		return nil, 0, err
	}

	return votes, total, nil
}

func (r *VoteRepository) GetUserVoteMap(userID uint, submissionIDs []uint) (map[uint]bool, error) {
	var votes []model.Vote
	result := make(map[uint]bool)
	if len(submissionIDs) == 0 {
		return result, nil
	}

	if err := r.db.Where("user_id = ? AND submission_id IN ?", userID, submissionIDs).Find(&votes).Error; err != nil {
		return nil, err
	}

	for _, vote := range votes {
		result[vote.SubmissionID] = true
	}

	return result, nil
}

// Transaction 执行事务，由仓储封装内部 db 实例
func (r *VoteRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// SignInRepository 签到仓储
type SignInRepository struct {
	db *gorm.DB
}

func NewSignInRepository(db *gorm.DB) *SignInRepository {
	return &SignInRepository{db: db}
}

func (r *SignInRepository) Create(signIn *model.SignIn) error {
	return r.db.Create(signIn).Error
}

func (r *SignInRepository) GetByID(id uint) (*model.SignIn, error) {
	var signIn model.SignIn
	if err := r.db.First(&signIn, id).Error; err != nil {
		return nil, err
	}
	return &signIn, nil
}

func (r *SignInRepository) GetByActivityAndUser(activityID, userID uint) (*model.SignIn, error) {
	var signIn model.SignIn
	if err := r.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&signIn).Error; err != nil {
		return nil, err
	}
	return &signIn, nil
}

func (r *SignInRepository) Exists(activityID, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.SignIn{}).Where("activity_id = ? AND user_id = ?", activityID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *SignInRepository) ListByActivity(activityID uint, page, pageSize int) ([]model.SignIn, int64, error) {
	var signIns []model.SignIn
	var total int64
	query := r.db.Model(&model.SignIn{}).Where("activity_id = ?", activityID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&signIns).Error; err != nil {
		return nil, 0, err
	}

	return signIns, total, nil
}

func (r *SignInRepository) CountByActivity(activityID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.SignIn{}).Where("activity_id = ?", activityID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}



// ActivityAlbumRepository 活动相册仓储
type ActivityAlbumRepository struct {
	db *gorm.DB
}

func NewActivityAlbumRepository(db *gorm.DB) *ActivityAlbumRepository {
	return &ActivityAlbumRepository{db: db}
}

func (r *ActivityAlbumRepository) Create(album *model.ActivityAlbum) error {
	return r.db.Create(album).Error
}

func (r *ActivityAlbumRepository) Update(album *model.ActivityAlbum) error {
	return r.db.Save(album).Error
}

func (r *ActivityAlbumRepository) Delete(id uint) error {
	return r.db.Delete(&model.ActivityAlbum{}, id).Error
}

func (r *ActivityAlbumRepository) GetByID(id uint) (*model.ActivityAlbum, error) {
	var album model.ActivityAlbum
	if err := r.db.First(&album, id).Error; err != nil {
		return nil, err
	}
	return &album, nil
}

func (r *ActivityAlbumRepository) ListByActivity(activityID uint, status int, page, pageSize int) ([]model.ActivityAlbum, int64, error) {
	var albums []model.ActivityAlbum
	var total int64
	query := r.db.Model(&model.ActivityAlbum{}).Where("activity_id = ?", activityID)

	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&albums).Error; err != nil {
		return nil, 0, err
	}

	return albums, total, nil
}

func (r *ActivityAlbumRepository) ListByUser(userID uint, page, pageSize int) ([]model.ActivityAlbum, int64, error) {
	var albums []model.ActivityAlbum
	var total int64
	query := r.db.Model(&model.ActivityAlbum{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(pageSize).Offset((page - 1) * pageSize).Order("id DESC").Find(&albums).Error; err != nil {
		return nil, 0, err
	}

	return albums, total, nil
}

func (r *ActivityAlbumRepository) CountByActivity(activityID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&model.ActivityAlbum{}).Where("activity_id = ? AND status = 1", activityID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ActivityReviewRepository 活动回顾仓储
type ActivityReviewRepository struct {
	db *gorm.DB
}

func NewActivityReviewRepository(db *gorm.DB) *ActivityReviewRepository {
	return &ActivityReviewRepository{db: db}
}

func (r *ActivityReviewRepository) Create(review *model.ActivityReview) error {
	return r.db.Create(review).Error
}

func (r *ActivityReviewRepository) Update(review *model.ActivityReview) error {
	return r.db.Save(review).Error
}

func (r *ActivityReviewRepository) GetByID(id uint) (*model.ActivityReview, error) {
	var review model.ActivityReview
	if err := r.db.First(&review, id).Error; err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ActivityReviewRepository) GetByActivity(activityID uint) (*model.ActivityReview, error) {
	var review model.ActivityReview
	if err := r.db.Where("activity_id = ?", activityID).First(&review).Error; err != nil {
		return nil, err
	}
	return &review, nil
}


