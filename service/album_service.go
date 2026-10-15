package service

import (
	"errors"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
	"strings"
	"time"
)

// AlbumService 活动相册服务
type AlbumService struct {
	albumRepo        *repository.ActivityAlbumRepository
	activityRepo     *repository.ActivityRepository
	registrationRepo *repository.RegistrationRepository
	userRepo         *repository.UserRepository
}

func NewAlbumService(
	albumRepo *repository.ActivityAlbumRepository,
	activityRepo *repository.ActivityRepository,
	registrationRepo *repository.RegistrationRepository,
	userRepo *repository.UserRepository,
) *AlbumService {
	return &AlbumService{
		albumRepo:        albumRepo,
		activityRepo:     activityRepo,
		registrationRepo: registrationRepo,
		userRepo:         userRepo,
	}
}

// UploadPhoto 上传照片到活动相册
func (s *AlbumService) UploadPhoto(userID uint, req *types.AlbumPhotoCreateRequest) (*types.AlbumPhotoInfo, error) {
	activity, err := s.activityRepo.GetByID(req.ActivityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	// 只有线下活动才有相册
	if activity.Type != "offline_meet" {
		return nil, errors.New("只有线下活动才能上传相册照片")
	}

	// 检查是否已报名（可选）
	// isRegistered, _ := s.registrationRepo.Exists(req.ActivityID, userID)
	// if !isRegistered {
	//     return nil, errors.New("请先报名活动")
	// }

	album := &model.ActivityAlbum{
		ActivityID:  req.ActivityID,
		UserID:      userID,
		PhotoURL:    req.PhotoURL,
		Description: req.Description,
		Status:      1, // 默认已通过，如果需要审核可以设置为0
	}

	if err := s.albumRepo.Create(album); err != nil {
		return nil, err
	}

	return s.convertToAlbumPhotoInfo(album)
}

// DeletePhoto 删除照片
func (s *AlbumService) DeletePhoto(userID, photoID uint) error {
	album, err := s.albumRepo.GetByID(photoID)
	if err != nil {
		return errors.New("照片不存在")
	}

	if album.UserID != userID {
		return errors.New("无权限删除")
	}

	return s.albumRepo.Delete(photoID)
}

// ListPhotos 获取照片列表
func (s *AlbumService) ListPhotos(activityID uint, status int, page, pageSize int) (*types.AlbumPhotoListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	albums, total, err := s.albumRepo.ListByActivity(activityID, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	photoInfos := make([]types.AlbumPhotoInfo, 0, len(albums))
	for _, album := range albums {
		info, err := s.convertToAlbumPhotoInfo(&album)
		if err == nil {
			photoInfos = append(photoInfos, *info)
		}
	}

	return &types.AlbumPhotoListResponse{
		List:     photoInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ReviewPhoto 审核照片
func (s *AlbumService) ReviewPhoto(photoID uint, approve bool, reviewerID uint) error {
	album, err := s.albumRepo.GetByID(photoID)
	if err != nil {
		return errors.New("照片不存在")
	}

	if approve {
		album.Status = 1
	} else {
		album.Status = 2
	}
	album.ReviewedBy = reviewerID
	album.ReviewedAt = time.Now()

	return s.albumRepo.Update(album)
}

// convertToAlbumPhotoInfo 转换照片信息
func (s *AlbumService) convertToAlbumPhotoInfo(album *model.ActivityAlbum) (*types.AlbumPhotoInfo, error) {
	user, err := s.userRepo.GetByID(album.UserID)
	userNickname := ""
	userAvatar := ""
	if err == nil {
		userNickname = user.Nickname
		userAvatar = user.Avatar
	}

	return &types.AlbumPhotoInfo{
		ID:           album.ID,
		ActivityID:   album.ActivityID,
		UserID:       album.UserID,
		UserNickname: userNickname,
		UserAvatar:   userAvatar,
		PhotoURL:     album.PhotoURL,
		Description:  album.Description,
		Status:       album.Status,
		CreatedAt:    album.CreatedAt.Format(time.RFC3339),
	}, nil
}

// ActivityReviewService 活动回顾服务
type ActivityReviewService struct {
	reviewRepo   *repository.ActivityReviewRepository
	activityRepo *repository.ActivityRepository
	userRepo     *repository.UserRepository
}

func NewActivityReviewService(
	reviewRepo *repository.ActivityReviewRepository,
	activityRepo *repository.ActivityRepository,
	userRepo *repository.UserRepository,
) *ActivityReviewService {
	return &ActivityReviewService{
		reviewRepo:   reviewRepo,
		activityRepo: activityRepo,
		userRepo:     userRepo,
	}
}

// PublishReview 发布活动回顾
func (s *ActivityReviewService) PublishReview(userID uint, req *types.ActivityReviewCreateRequest) (*types.ActivityReviewInfo, error) {
	activity, err := s.activityRepo.GetByID(req.ActivityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	// 检查是否已结束
	if time.Now().Before(activity.EndTime) {
		return nil, errors.New("活动尚未结束，不能发布回顾")
	}

	// 检查是否已发布回顾
	existingReview, _ := s.reviewRepo.GetByActivity(req.ActivityID)
	if existingReview != nil {
		return nil, errors.New("该活动已发布回顾")
	}

	photosStr := ""
	if len(req.Photos) > 0 {
		photosStr = strings.Join(req.Photos, ",")
	}

	review := &model.ActivityReview{
		ActivityID:  req.ActivityID,
		Title:       req.Title,
		Content:     req.Content,
		Photos:      photosStr,
		PublishedAt: time.Now(),
		PublishedBy: userID,
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	return s.convertToReviewInfo(review)
}

// GetReview 获取活动回顾
func (s *ActivityReviewService) GetReview(activityID uint) (*types.ActivityReviewInfo, error) {
	review, err := s.reviewRepo.GetByActivity(activityID)
	if err != nil {
		return nil, errors.New("活动回顾不存在")
	}

	return s.convertToReviewInfo(review)
}

// UpdateReview 更新活动回顾
func (s *ActivityReviewService) UpdateReview(userID, activityID uint, req *types.ActivityReviewCreateRequest) (*types.ActivityReviewInfo, error) {
	review, err := s.reviewRepo.GetByActivity(activityID)
	if err != nil {
		return nil, errors.New("活动回顾不存在")
	}

	if review.PublishedBy != userID {
		return nil, errors.New("无权限修改")
	}

	review.Title = req.Title
	review.Content = req.Content
	photosStr := ""
	if len(req.Photos) > 0 {
		photosStr = strings.Join(req.Photos, ",")
	}
	review.Photos = photosStr

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, err
	}

	return s.convertToReviewInfo(review)
}

// convertToReviewInfo 转换回顾信息
func (s *ActivityReviewService) convertToReviewInfo(review *model.ActivityReview) (*types.ActivityReviewInfo, error) {
	photos := []string{}
	if review.Photos != "" {
		photos = strings.Split(review.Photos, ",")
	}

	return &types.ActivityReviewInfo{
		ID:          review.ID,
		ActivityID:  review.ActivityID,
		Title:       review.Title,
		Content:     review.Content,
		Photos:      photos,
		PublishedAt: review.PublishedAt.Format(time.RFC3339),
		PublishedBy: review.PublishedBy,
		CreatedAt:   review.CreatedAt.Format(time.RFC3339),
	}, nil
}
