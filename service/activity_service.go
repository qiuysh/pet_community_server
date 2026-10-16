package service

import (
	"errors"
	"strings"
	"time"

	"pet-community/cache"
	"pet-community/config"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"

	"gorm.io/gorm"
)

// ActivityService 活动服务
type ActivityService struct {
	activityRepo     *repository.ActivityRepository
	registrationRepo *repository.RegistrationRepository
	submissionRepo   *repository.SubmissionRepository
	voteRepo         *repository.VoteRepository
	userRepo         *repository.UserRepository
}

func NewActivityService(
	activityRepo *repository.ActivityRepository,
	registrationRepo *repository.RegistrationRepository,
	submissionRepo *repository.SubmissionRepository,
	voteRepo *repository.VoteRepository,
	userRepo *repository.UserRepository,
) *ActivityService {
	return &ActivityService{
		activityRepo:     activityRepo,
		registrationRepo: registrationRepo,
		submissionRepo:   submissionRepo,
		voteRepo:         voteRepo,
		userRepo:         userRepo,
	}
}

// CreateActivity 创建活动
func (s *ActivityService) CreateActivity(req *types.ActivityCreateRequest) (*types.ActivityInfo, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, errors.New("开始时间格式错误")
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, errors.New("结束时间格式错误")
	}

	if endTime.Before(startTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}

	if endTime.Sub(startTime) > 30*24*time.Hour {
		return nil, errors.New("活动时间不能超过30天")
	}

	maxSubmissions := req.MaxSubmissions
	if maxSubmissions == 0 {
		maxSubmissions = 1
	}

	votesPerDay := req.VotesPerDay
	if votesPerDay == 0 {
		votesPerDay = 10
	}

	activity := &model.Activity{
		Title:           req.Title,
		Description:     req.Description,
		Type:            req.Type,
		Cover:           req.Cover,
		Status:          0,
		StartTime:       startTime,
		EndTime:         endTime,
		Rules:           req.Rules,
		Prizes:          req.Prizes,
		MaxParticipants: req.MaxParticipants,
		MaxSubmissions:  maxSubmissions,
		VotesPerDay:     votesPerDay,
		RequireApproval: req.RequireApproval,
	}

	if err := s.activityRepo.Create(activity); err != nil {
		return nil, err
	}

	return s.convertToActivityInfo(activity, 0)
}

// UpdateActivity 更新活动
func (s *ActivityService) UpdateActivity(activityID uint, req *types.ActivityUpdateRequest) (*types.ActivityInfo, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 0 {
		return nil, errors.New("只能修改草稿状态的活动")
	}

	if req.Title != "" {
		activity.Title = req.Title
	}
	if req.Description != "" {
		activity.Description = req.Description
	}
	if req.Cover != "" {
		activity.Cover = req.Cover
	}
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			return nil, errors.New("开始时间格式错误")
		}
		activity.StartTime = startTime
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		activity.EndTime = endTime
	}
	if endTime := req.EndTime; endTime != "" {
		if req.StartTime != "" {
			startTime, _ := time.Parse(time.RFC3339, req.StartTime)
			end, _ := time.Parse(time.RFC3339, endTime)
			if end.Before(startTime) {
				return nil, errors.New("结束时间不能早于开始时间")
			}
		}
	}
	if req.Rules != "" {
		activity.Rules = req.Rules
	}
	if req.Prizes != "" {
		activity.Prizes = req.Prizes
	}
	if req.MaxParticipants > 0 {
		activity.MaxParticipants = req.MaxParticipants
	}
	if req.MaxSubmissions > 0 {
		activity.MaxSubmissions = req.MaxSubmissions
	}
	if req.VotesPerDay > 0 {
		activity.VotesPerDay = req.VotesPerDay
	}
	activity.RequireApproval = req.RequireApproval

	if err := s.activityRepo.Update(activity); err != nil {
		return nil, err
	}

	return s.convertToActivityInfo(activity, 0)
}

// DeleteActivity 删除活动
func (s *ActivityService) DeleteActivity(activityID uint) error {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return errors.New("活动不存在")
	}

	if activity.Status != 0 {
		return errors.New("只能删除草稿状态的活动")
	}

	return s.activityRepo.Delete(activityID)
}

// GetActivityByID 获取活动详情
func (s *ActivityService) GetActivityByID(userID, activityID uint) (*types.ActivityInfo, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	info, err := s.convertToActivityInfo(activity, userID)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// ListActivities 活动列表（前台：默认不含草稿）
func (s *ActivityService) ListActivities(userID uint, activityType, status, keyword string, page, pageSize int) (*types.ActivityListResponse, error) {
	return s.listActivities(userID, activityType, status, keyword, page, pageSize, false)
}

// ListActivitiesForAdmin 管理端活动列表（可含草稿）
func (s *ActivityService) ListActivitiesForAdmin(activityType, status, keyword string, page, pageSize int) (*types.ActivityListResponse, error) {
	return s.listActivities(0, activityType, status, keyword, page, pageSize, true)
}

func (s *ActivityService) listActivities(userID uint, activityType, status, keyword string, page, pageSize int, includeDraft bool) (*types.ActivityListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	activities, total, err := s.activityRepo.List(activityType, status, keyword, page, pageSize, includeDraft)
	if err != nil {
		return nil, err
	}

	activityInfos := make([]types.ActivityInfo, 0, len(activities))
	for _, activity := range activities {
		info, err := s.convertToActivityInfo(&activity, userID)
		if err == nil {
			activityInfos = append(activityInfos, *info)
		}
	}

	return &types.ActivityListResponse{
		List:     activityInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// PublishActivity 发布活动
func (s *ActivityService) PublishActivity(activityID uint) error {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return errors.New("活动不存在")
	}

	if activity.Status != 0 {
		return errors.New("只能发布草稿状态的活动")
	}

	activity.Status = 1
	return s.activityRepo.Update(activity)
}

// convertToActivityInfo 转换活动信息
func (s *ActivityService) convertToActivityInfo(activity *model.Activity, userID uint) (*types.ActivityInfo, error) {
	participantsCount, _ := s.registrationRepo.CountByActivity(activity.ID)
	submissionsCount, _ := s.submissionRepo.CountByActivity(activity.ID)

	isRegistered := false
	hasSubmitted := false
	if userID > 0 {
		isRegistered, _ = s.registrationRepo.Exists(activity.ID, userID)
		submissionCount, _ := s.submissionRepo.CountByActivityAndUser(activity.ID, userID)
		hasSubmitted = submissionCount > 0
	}

	return &types.ActivityInfo{
		ID:                activity.ID,
		Title:             activity.Title,
		Description:       activity.Description,
		Type:              activity.Type,
		Cover:             activity.Cover,
		Status:            activity.Status,
		StartTime:         activity.StartTime.Format(time.RFC3339),
		EndTime:           activity.EndTime.Format(time.RFC3339),
		Rules:             activity.Rules,
		Prizes:            activity.Prizes,
		MaxParticipants:   activity.MaxParticipants,
		MaxSubmissions:    activity.MaxSubmissions,
		VotesPerDay:       activity.VotesPerDay,
		RequireApproval:   activity.RequireApproval,
		CreatedAt:         activity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         activity.UpdatedAt.Format(time.RFC3339),
		ParticipantsCount: int(participantsCount),
		SubmissionsCount:  int(submissionsCount),
		IsRegistered:      isRegistered,
		HasSubmitted:      hasSubmitted,
	}, nil
}

// RegistrationService 报名服务
type RegistrationService struct {
	registrationRepo *repository.RegistrationRepository
	activityRepo     *repository.ActivityRepository
	userRepo         *repository.UserRepository
}

func NewRegistrationService(
	registrationRepo *repository.RegistrationRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
) *RegistrationService {
	return &RegistrationService{
		registrationRepo: registrationRepo,
		activityRepo:     activityRepo,
		userRepo:         userRepo,
	}
}

// Register 活动报名
func (s *RegistrationService) Register(userID uint, req *types.RegistrationRequest) (*types.RegistrationInfo, error) {
	activity, err := s.activityRepo.GetByID(req.ActivityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	if time.Now().Before(activity.StartTime) {
		return nil, errors.New("活动尚未开始")
	}

	if time.Now().After(activity.EndTime) {
		return nil, errors.New("活动已结束")
	}

	exists, _ := s.registrationRepo.Exists(req.ActivityID, userID)
	if exists {
		return nil, errors.New("已经报名")
	}

	if activity.MaxParticipants > 0 {
		count, _ := s.registrationRepo.CountByActivity(req.ActivityID)
		if count >= int64(activity.MaxParticipants) {
			return nil, errors.New("报名人数已达上限")
		}
	}

	priority := 0
	if config.DB != nil {
		var level model.UserLevel
		if err := config.DB.Where("user_id = ?", userID).First(&level).Error; err == nil && level.Level >= 8 {
			priority = 1
		}
	}

	registration := &model.Registration{
		ActivityID: req.ActivityID,
		UserID:     userID,
		Name:       req.Name,
		Phone:      req.Phone,
		Message:    req.Message,
		Status:     0,
		Priority:   priority,
	}

	if !activity.RequireApproval {
		registration.Status = 1
	}

	if err := s.registrationRepo.Create(registration); err != nil {
		return nil, err
	}

	return s.convertToRegistrationInfo(registration)
}

// CancelRegistration 取消报名
func (s *RegistrationService) CancelRegistration(userID, activityID uint) error {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return errors.New("活动不存在")
	}

	if time.Now().After(activity.StartTime) {
		return errors.New("活动已开始，无法取消报名")
	}

	registration, err := s.registrationRepo.GetByActivityAndUser(activityID, userID)
	if err != nil {
		return errors.New("未报名该活动")
	}

	return s.registrationRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(registration).Error; err != nil {
			return err
		}
		return nil
	})
}

// ListRegistrations 报名列表
func (s *RegistrationService) ListRegistrations(activityID uint, status int, page, pageSize int) (*types.RegistrationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	registrations, total, err := s.registrationRepo.ListByActivity(activityID, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	registrationInfos := make([]types.RegistrationInfo, 0, len(registrations))
	for _, registration := range registrations {
		info, err := s.convertToRegistrationInfo(&registration)
		if err == nil {
			registrationInfos = append(registrationInfos, *info)
		}
	}

	return &types.RegistrationListResponse{
		List:     registrationInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ApproveRegistration 审核报名
func (s *RegistrationService) ApproveRegistration(registrationID uint, approve bool) error {
	registration, err := s.registrationRepo.GetByID(registrationID)
	if err != nil {
		return errors.New("报名记录不存在")
	}

	if approve {
		registration.Status = 1
	} else {
		registration.Status = 2
	}

	return s.registrationRepo.Update(registration)
}

// convertToRegistrationInfo 转换报名信息
func (s *RegistrationService) convertToRegistrationInfo(registration *model.Registration) (*types.RegistrationInfo, error) {
	user, err := s.userRepo.GetByID(registration.UserID)
	userNickname := ""
	userAvatar := ""
	if err == nil {
		userNickname = user.Nickname
		userAvatar = user.Avatar
	}

	return &types.RegistrationInfo{
		ID:           registration.ID,
		ActivityID:   registration.ActivityID,
		UserID:       registration.UserID,
		UserNickname: userNickname,
		UserAvatar:   userAvatar,
		Status:       registration.Status,
		Name:         registration.Name,
		Phone:        registration.Phone,
		Message:      registration.Message,
		CreatedAt:    registration.CreatedAt.Format(time.RFC3339),
	}, nil
}

// SubmissionService 作品服务
type SubmissionService struct {
	submissionRepo   *repository.SubmissionRepository
	activityRepo     *repository.ActivityRepository
	registrationRepo *repository.RegistrationRepository
	userRepo         *repository.UserRepository
	voteRepo         *repository.VoteRepository
}

func NewSubmissionService(
	submissionRepo *repository.SubmissionRepository,
	activityRepo *repository.ActivityRepository,
	registrationRepo *repository.RegistrationRepository,
	userRepo *repository.UserRepository,
	voteRepo *repository.VoteRepository,
) *SubmissionService {
	return &SubmissionService{
		submissionRepo:   submissionRepo,
		activityRepo:     activityRepo,
		registrationRepo: registrationRepo,
		userRepo:         userRepo,
		voteRepo:         voteRepo,
	}
}

// CreateSubmission 创建作品
func (s *SubmissionService) CreateSubmission(userID uint, req *types.SubmissionCreateRequest) (*types.SubmissionInfo, error) {
	activity, err := s.activityRepo.GetByID(req.ActivityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	if time.Now().Before(activity.StartTime) {
		return nil, errors.New("活动尚未开始")
	}

	if time.Now().After(activity.EndTime) {
		return nil, errors.New("活动已结束")
	}

	isRegistered, _ := s.registrationRepo.Exists(req.ActivityID, userID)
	if !isRegistered {
		return nil, errors.New("请先报名活动")
	}

	registration, _ := s.registrationRepo.GetByActivityAndUser(req.ActivityID, userID)
	if registration != nil && registration.Status != 1 {
		return nil, errors.New("报名未通过审核")
	}

	count, _ := s.submissionRepo.CountByActivityAndUser(req.ActivityID, userID)
	if count >= int64(activity.MaxSubmissions) {
		return nil, errors.New("已达到作品提交上限")
	}

	if req.Type == "image" && len(req.Images) == 0 {
		return nil, errors.New("图片作品必须上传图片")
	}
	if req.Type == "video" && req.Video == "" {
		return nil, errors.New("视频作品必须上传视频")
	}

	imagesStr := ""
	if len(req.Images) > 0 {
		imagesStr = strings.Join(req.Images, ",")
	}

	submission := &model.Submission{
		ActivityID:  req.ActivityID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Content:     req.Content,
		Images:      imagesStr,
		Video:       req.Video,
		Status:      1,
	}

	if err := s.submissionRepo.Create(submission); err != nil {
		return nil, err
	}

	return s.convertToSubmissionInfo(submission, userID)
}

// UpdateSubmission 更新作品
func (s *SubmissionService) UpdateSubmission(userID, submissionID uint, req *types.SubmissionUpdateRequest) (*types.SubmissionInfo, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, errors.New("作品不存在")
	}

	if submission.UserID != userID {
		return nil, errors.New("无权限修改")
	}

	activity, err := s.activityRepo.GetByID(submission.ActivityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 || time.Now().After(activity.EndTime) {
		return nil, errors.New("活动已结束，无法修改作品")
	}

	if req.Title != "" {
		submission.Title = req.Title
	}
	if req.Description != "" {
		submission.Description = req.Description
	}
	if req.Content != "" {
		submission.Content = req.Content
	}
	if req.Images != nil {
		imagesStr := ""
		if len(req.Images) > 0 {
			imagesStr = strings.Join(req.Images, ",")
		}
		submission.Images = imagesStr
	}
	if req.Video != "" {
		submission.Video = req.Video
	}

	if err := s.submissionRepo.Update(submission); err != nil {
		return nil, err
	}

	return s.convertToSubmissionInfo(submission, userID)
}

// DeleteSubmission 删除作品
func (s *SubmissionService) DeleteSubmission(userID, submissionID uint) error {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return errors.New("作品不存在")
	}

	if submission.UserID != userID {
		return errors.New("无权限删除")
	}

	activity, err := s.activityRepo.GetByID(submission.ActivityID)
	if err == nil && activity.Status == 1 && time.Now().Before(activity.EndTime) {
		return errors.New("活动进行中，无法删除作品")
	}

	return s.submissionRepo.Delete(submissionID)
}

// ListSubmissions 作品列表
func (s *SubmissionService) ListSubmissions(userID uint, activityID uint, sort string, page, pageSize int) (*types.SubmissionListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	submissions, total, err := s.submissionRepo.ListByActivity(activityID, sort, page, pageSize)
	if err != nil {
		return nil, err
	}

	submissionInfos := make([]types.SubmissionInfo, 0, len(submissions))
	var submissionIDs []uint
	for _, submission := range submissions {
		submissionIDs = append(submissionIDs, submission.ID)
	}

	voteMap := make(map[uint]bool)
	if userID > 0 && len(submissionIDs) > 0 {
		voteMap, _ = s.voteRepo.GetUserVoteMap(userID, submissionIDs)
	}

	for i, submission := range submissions {
		info, err := s.convertToSubmissionInfo(&submission, userID)
		if err == nil {
			info.IsVoted = voteMap[submission.ID]
			if sort == "votes" {
				info.Rank = i + 1
			}
			submissionInfos = append(submissionInfos, *info)
		}
	}

	return &types.SubmissionListResponse{
		List:     submissionInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetSubmissionByID 获取作品详情
func (s *SubmissionService) GetSubmissionByID(userID, submissionID uint) (*types.SubmissionInfo, error) {
	submission, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return nil, errors.New("作品不存在")
	}

	return s.convertToSubmissionInfo(submission, userID)
}

// AwardSubmission 评选作品
func (s *SubmissionService) AwardSubmission(submissionID uint, awardLevel string) error {
	_, err := s.submissionRepo.GetByID(submissionID)
	if err != nil {
		return errors.New("作品不存在")
	}

	return s.submissionRepo.SetAward(submissionID, true, awardLevel)
}

// GetAwardPage 获取评选页面数据
func (s *SubmissionService) GetAwardPage(userID uint, activityID uint) (*types.AwardPageResponse, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	// 检查活动状态（必须是已结束的活动才能评选）
	if activity.Status != 1 {
		return nil, errors.New("活动状态不允许评选")
	}

	// 获取所有作品（按票数排序）
	submissions, total, err := s.submissionRepo.ListByActivity(activityID, "votes", 1, 100)
	if err != nil {
		return nil, err
	}

	// 获取活动统计信息
	participantsCount, _ := s.registrationRepo.CountByActivity(activityID)
	submissionsCount, _ := s.submissionRepo.CountByActivity(activityID)

	activityInfo := &types.ActivityInfo{
		ID:                activity.ID,
		Title:             activity.Title,
		Description:       activity.Description,
		Type:              activity.Type,
		Cover:             activity.Cover,
		Status:            activity.Status,
		StartTime:         activity.StartTime.Format(time.RFC3339),
		EndTime:           activity.EndTime.Format(time.RFC3339),
		Rules:             activity.Rules,
		Prizes:            activity.Prizes,
		MaxParticipants:   activity.MaxParticipants,
		MaxSubmissions:    activity.MaxSubmissions,
		VotesPerDay:       activity.VotesPerDay,
		RequireApproval:   activity.RequireApproval,
		CreatedAt:         activity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         activity.UpdatedAt.Format(time.RFC3339),
		ParticipantsCount: int(participantsCount),
		SubmissionsCount:  int(submissionsCount),
	}

	submissionInfos := make([]types.SubmissionInfo, 0, len(submissions))
	var firstPrize *types.SubmissionInfo
	var secondPrizes []types.SubmissionInfo
	var thirdPrizes []types.SubmissionInfo
	var awardedCount int64

	for _, submission := range submissions {
		info, err := s.convertToSubmissionInfo(&submission, userID)
		if err == nil {
			submissionInfos = append(submissionInfos, *info)
			if submission.IsAwarded {
				awardedCount++
				switch submission.AwardLevel {
				case "first":
					firstPrize = info
				case "second":
					secondPrizes = append(secondPrizes, *info)
				case "third":
					thirdPrizes = append(thirdPrizes, *info)
				}
			}
		}
	}

	return &types.AwardPageResponse{
		Activity:     activityInfo,
		Submissions:  submissionInfos,
		Total:        total,
		AwardedCount: awardedCount,
		FirstPrize:   firstPrize,
		SecondPrizes: secondPrizes,
		ThirdPrizes:  thirdPrizes,
	}, nil
}

// BatchAwardSubmission 批量设置获奖
func (s *SubmissionService) BatchAwardSubmission(awards []types.AwardRequest) error {
	// 验证奖项设置是否符合规则（一等奖1个，二等奖2个，三等奖3个）
	firstCount := 0
	secondCount := 0
	thirdCount := 0

	for _, award := range awards {
		switch award.AwardLevel {
		case "first":
			firstCount++
		case "second":
			secondCount++
		case "third":
			thirdCount++
		}
	}

	if firstCount > 1 {
		return errors.New("一等奖只能设置1个")
	}
	if secondCount > 2 {
		return errors.New("二等奖最多设置2个")
	}
	if thirdCount > 3 {
		return errors.New("三等奖最多设置3个")
	}

	// 批量设置获奖
	for _, award := range awards {
		if err := s.AwardSubmission(award.SubmissionID, award.AwardLevel); err != nil {
			return err
		}
	}

	return nil
}

// PublishAwardResult 发布评选结果
func (s *SubmissionService) PublishAwardResult(activityID uint) (*types.PublishAwardResponse, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	// 检查活动状态（必须是已结束的活动）
	if activity.Status != 1 {
		return nil, errors.New("活动状态不允许发布结果")
	}

	// 获取获奖作品
	submissions, _, err := s.submissionRepo.ListByActivity(activityID, "votes", 1, 100)
	if err != nil {
		return nil, err
	}

	awardedSubmissions := make([]model.Submission, 0)
	notifiedUsers := make([]uint, 0)

	for _, submission := range submissions {
		if submission.IsAwarded {
			awardedSubmissions = append(awardedSubmissions, submission)
			notifiedUsers = append(notifiedUsers, submission.UserID)
		}
	}

	if len(awardedSubmissions) == 0 {
		return &types.PublishAwardResponse{
			Success: false,
			Message: "未设置获奖作品",
		}, nil
	}

	// 更新活动状态为已结束评选
	activity.Status = 3 // 已结束
	if err := s.activityRepo.Update(activity); err != nil {
		return nil, err
	}

	// 这里可以添加发送通知的逻辑（通知获奖用户）
	// 暂时返回成功响应

	return &types.PublishAwardResponse{
		Success:       true,
		Message:       "评选结果已发布",
		NotifiedUsers: notifiedUsers,
	}, nil
}

// convertToSubmissionInfo 转换作品信息
func (s *SubmissionService) convertToSubmissionInfo(submission *model.Submission, userID uint) (*types.SubmissionInfo, error) {
	images := []string{}
	if submission.Images != "" {
		images = strings.Split(submission.Images, ",")
	}

	user, err := s.userRepo.GetByID(submission.UserID)
	userNickname := ""
	userAvatar := ""
	if err == nil {
		userNickname = user.Nickname
		userAvatar = user.Avatar
	}

	isVoted := false
	if userID > 0 {
		isVoted, _ = s.voteRepo.Exists(userID, submission.ID)
	}

	return &types.SubmissionInfo{
		ID:           submission.ID,
		ActivityID:   submission.ActivityID,
		UserID:       submission.UserID,
		UserNickname: userNickname,
		UserAvatar:   userAvatar,
		Title:        submission.Title,
		Description:  submission.Description,
		Type:         submission.Type,
		Content:      submission.Content,
		Images:       images,
		Video:        submission.Video,
		VotesCount:   submission.VotesCount,
		Rank:         submission.Rank,
		IsAwarded:    submission.IsAwarded,
		AwardLevel:   submission.AwardLevel,
		Status:       submission.Status,
		CreatedAt:    submission.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    submission.UpdatedAt.Format(time.RFC3339),
		IsVoted:      isVoted,
	}, nil
}

// VoteService 投票服务
type VoteService struct {
	voteRepo       *repository.VoteRepository
	submissionRepo *repository.SubmissionRepository
	activityRepo   *repository.ActivityRepository
	userRepo       *repository.UserRepository
}

func NewVoteService(
	voteRepo *repository.VoteRepository,
	submissionRepo *repository.SubmissionRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
) *VoteService {
	return &VoteService{
		voteRepo:       voteRepo,
		submissionRepo: submissionRepo,
		activityRepo:   activityRepo,
		userRepo:       userRepo,
	}
}

// Vote 为作品投票
func (s *VoteService) Vote(userID uint, req *types.VoteRequest) (*types.VoteResult, error) {
	activity, err := s.activityRepo.GetByID(req.ActivityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	if time.Now().Before(activity.StartTime) || time.Now().After(activity.EndTime) {
		return nil, errors.New("活动不在进行时间内")
	}

	submission, err := s.submissionRepo.GetByID(req.SubmissionID)
	if err != nil {
		return nil, errors.New("作品不存在")
	}

	if submission.ActivityID != req.ActivityID {
		return nil, errors.New("作品不属于该活动")
	}

	reserved := false
	if err := cache.ReserveVote(userID, req.SubmissionID, activity.VotesPerDay); err == nil {
		reserved = true
	} else if !errors.Is(err, cache.ErrUnavailable) {
		return nil, err
	} else {
		// Redis 不可用：回退 MySQL 计数
		existsToday, _ := s.voteRepo.ExistsToday(userID, req.SubmissionID)
		if existsToday {
			return nil, errors.New("今天已经为该作品投票")
		}
		countToday, _ := s.voteRepo.CountByUserToday(userID)
		if countToday >= int64(activity.VotesPerDay) {
			return nil, errors.New("今日投票次数已达上限")
		}
		countHour, _ := s.voteRepo.CountByUserInHour(userID)
		if countHour >= 3 {
			return nil, errors.New("每小时投票次数已达上限")
		}
	}

	if err := s.voteRepo.Transaction(func(tx *gorm.DB) error {
		vote := &model.Vote{
			ActivityID:   req.ActivityID,
			SubmissionID: req.SubmissionID,
			UserID:       userID,
		}

		if err := tx.Create(vote).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Submission{}).Where("id = ?", req.SubmissionID).UpdateColumn("votes_count", gorm.Expr("votes_count + 1")).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		if reserved {
			cache.UndoVote(userID, req.SubmissionID)
		}
		return nil, err
	}

	countToday := cache.VoteDayCount(userID)
	if countToday < 0 {
		countToday, _ = s.voteRepo.CountByUserToday(userID)
	}
	remainingVotes := activity.VotesPerDay - int(countToday)
	if remainingVotes < 0 {
		remainingVotes = 0
	}
	return &types.VoteResult{
		Voted:          true,
		VotesCount:     submission.VotesCount + 1,
		RemainingVotes: remainingVotes,
	}, nil
}

// CancelVote 取消投票
func (s *VoteService) CancelVote(userID, submissionID uint) error {
	vote, err := s.voteRepo.GetByUserAndSubmission(userID, submissionID)
	if err != nil {
		return errors.New("投票不存在")
	}

	if time.Since(vote.CreatedAt) > 24*time.Hour {
		return errors.New("投票超过24小时，无法撤销")
	}

	err = s.voteRepo.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(vote).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Submission{}).Where("id = ?", submissionID).UpdateColumn("votes_count", gorm.Expr("votes_count - 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		cache.UndoVote(userID, submissionID)
	}
	return err
}

// GetVoteHistory 获取投票历史
func (s *VoteService) GetVoteHistory(userID uint, activityID uint, page, pageSize int) (*types.VoteHistoryResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 10
	}

	votes, total, err := s.voteRepo.ListByUser(userID, activityID, page, pageSize)
	if err != nil {
		return nil, err
	}

	voteInfos := make([]types.VoteInfo, 0, len(votes))
	for _, vote := range votes {
		voteInfos = append(voteInfos, types.VoteInfo{
			ID:           vote.ID,
			UserID:       vote.UserID,
			ActivityID:   vote.ActivityID,
			SubmissionID: vote.SubmissionID,
			CreatedAt:    vote.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.VoteHistoryResponse{
		List:     voteInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetUserVoteStats 获取用户投票统计
func (s *VoteService) GetUserVoteStats(userID, activityID uint) (*types.UserVoteStatsResponse, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	countToday, _ := s.voteRepo.CountByUserToday(userID)
	remainingVotes := activity.VotesPerDay - int(countToday)

	return &types.UserVoteStatsResponse{
		TotalVotesToday: int(countToday),
		RemainingVotes:  remainingVotes,
		VotesPerDay:     activity.VotesPerDay,
	}, nil
}
