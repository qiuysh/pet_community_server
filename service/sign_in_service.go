package service

import (
	"errors"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
	"time"
)

// SignInService 签到服务
type SignInService struct {
	signInRepo       *repository.SignInRepository
	activityRepo     *repository.ActivityRepository
	registrationRepo *repository.RegistrationRepository
	userRepo         *repository.UserRepository
	qrcodeService    *QRCodeService
}

func NewSignInService(
	signInRepo *repository.SignInRepository,
	activityRepo *repository.ActivityRepository,
	registrationRepo *repository.RegistrationRepository,
	userRepo *repository.UserRepository,
) *SignInService {
	return &SignInService{
		signInRepo:       signInRepo,
		activityRepo:     activityRepo,
		registrationRepo: registrationRepo,
		userRepo:         userRepo,
		qrcodeService:    NewQRCodeService(),
	}
}

// GenerateQRCode 生成签到二维码
func (s *SignInService) GenerateQRCode(activityID uint) (*types.QRCodeResponse, error) {
	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Type != "offline_meet" {
		return nil, errors.New("只有线下活动才能生成签到二维码")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	now := time.Now()
	if now.Before(activity.StartTime) {
		return nil, errors.New("活动尚未开始")
	}
	if now.After(activity.EndTime) {
		return nil, errors.New("活动已结束")
	}

	qrCode, expires := s.qrcodeService.GenerateQRCode(activityID)

	return &types.QRCodeResponse{
		QRCode:   qrCode,
		ImageURL: s.qrcodeService.GenerateQRCodeImageBase64(qrCode),
		Expires:  expires,
	}, nil
}

// SignIn 扫码签到
func (s *SignInService) SignIn(userID uint, req *types.SignInRequest) (*types.SignInInfo, error) {
	activityID, err := s.qrcodeService.ValidateQRCode(req.QRCode)
	if err != nil {
		return nil, err
	}

	if activityID != req.ActivityID {
		return nil, errors.New("二维码不属于该活动")
	}

	activity, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return nil, errors.New("活动不存在")
	}

	if activity.Status != 1 {
		return nil, errors.New("活动未发布")
	}

	now := time.Now()
	if now.Before(activity.StartTime) {
		return nil, errors.New("活动尚未开始")
	}
	if now.After(activity.EndTime) {
		return nil, errors.New("活动已结束")
	}

	// 检查是否已签到
	exists, _ := s.signInRepo.Exists(activityID, userID)
	if exists {
		return nil, errors.New("已经签到")
	}

	// 检查是否已报名（可选，根据活动要求）
	// isRegistered, _ := s.registrationRepo.Exists(activityID, userID)
	// if !isRegistered {
	//     return nil, errors.New("请先报名活动")
	// }

	signIn := &model.SignIn{
		ActivityID:     activityID,
		UserID:         userID,
		SignInTime:     now,
		SignInLocation: req.SignInLocation,
		QRCode:         req.QRCode,
	}

	if err := s.signInRepo.Create(signIn); err != nil {
		return nil, err
	}

	return s.convertToSignInInfo(signIn)
}

// ListSignIns 获取签到列表
func (s *SignInService) ListSignIns(activityID uint, page, pageSize int) (*types.SignInListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}

	signIns, total, err := s.signInRepo.ListByActivity(activityID, page, pageSize)
	if err != nil {
		return nil, err
	}

	signInInfos := make([]types.SignInInfo, 0, len(signIns))
	for _, signIn := range signIns {
		info, err := s.convertToSignInInfo(&signIn)
		if err == nil {
			signInInfos = append(signInInfos, *info)
		}
	}

	return &types.SignInListResponse{
		List:     signInInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetSignInStats 获取签到统计
func (s *SignInService) GetSignInStats(activityID uint) (int64, int64, error) {
	_, err := s.activityRepo.GetByID(activityID)
	if err != nil {
		return 0, 0, errors.New("活动不存在")
	}

	signInCount, _ := s.signInRepo.CountByActivity(activityID)
	registrationCount, _ := s.registrationRepo.CountByActivity(activityID)

	return signInCount, registrationCount, nil
}

// convertToSignInInfo 转换签到信息
func (s *SignInService) convertToSignInInfo(signIn *model.SignIn) (*types.SignInInfo, error) {
	user, err := s.userRepo.GetByID(signIn.UserID)
	userNickname := ""
	userAvatar := ""
	if err == nil {
		userNickname = user.Nickname
		userAvatar = user.Avatar
	}

	return &types.SignInInfo{
		ID:             signIn.ID,
		ActivityID:     signIn.ActivityID,
		UserID:         signIn.UserID,
		UserNickname:   userNickname,
		UserAvatar:     userAvatar,
		SignInTime:     signIn.SignInTime.Format(time.RFC3339),
		SignInLocation: signIn.SignInLocation,
		CreatedAt:      signIn.CreatedAt.Format(time.RFC3339),
	}, nil
}
