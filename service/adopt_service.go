package service

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
	"pet-community/utils"

	"gorm.io/gorm"
)

// AdoptService 领养服务
type AdoptService struct {
	adoptRepo       *repository.AdoptRepository
	applicationRepo *repository.AdoptionApplicationRepository
	followUpRepo    *repository.AdoptionFollowUpRepository
	userRepo        *repository.UserRepository
}

func NewAdoptService(
	adoptRepo *repository.AdoptRepository,
	applicationRepo *repository.AdoptionApplicationRepository,
	followUpRepo *repository.AdoptionFollowUpRepository,
	userRepo *repository.UserRepository,
) *AdoptService {
	return &AdoptService{
		adoptRepo:       adoptRepo,
		applicationRepo: applicationRepo,
		followUpRepo:    followUpRepo,
		userRepo:        userRepo,
	}
}

// CreateAdopt 创建领养信息
func (s *AdoptService) CreateAdopt(userID uint, req *types.AdoptCreateRequest) (*types.AdoptInfo, error) {
	// 校验照片 URL，禁止占位符与未上传的本地预览 URL 入库
	if err := validatePhotos(req.Photos); err != nil {
		return nil, err
	}

	// 转换照片数组为JSON字符串
	photosJSON, err := json.Marshal(req.Photos)
	if err != nil {
		return nil, errors.New("照片数据格式错误")
	}

	adopt := &model.Adopt{
		PetID:           req.PetID,
		UserID:          userID,
		PetName:         req.PetName,
		PetType:         req.PetType,
		PetAge:          req.PetAge,
		PetGender:       req.PetGender,
		HealthStatus:    req.HealthStatus,
		Photos:          string(photosJSON),
		Description:     req.Description,
		AdoptConditions: req.AdoptConditions,
		Location:        req.Location,
		ContactPhone:    req.ContactPhone,
		Status:          0, // 待领养状态
	}

	if err := s.adoptRepo.Create(adopt); err != nil {
		return nil, err
	}

	return s.convertToAdoptInfo(adopt, userID)
}

// UpdateAdopt 更新领养信息
func (s *AdoptService) UpdateAdopt(userID, adoptID uint, req *types.AdoptUpdateRequest) (*types.AdoptInfo, error) {
	adopt, err := s.adoptRepo.GetByID(adoptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("领养信息不存在")
		}
		return nil, err
	}

	// 验证用户权限
	if adopt.UserID != userID {
		return nil, errors.New("无权修改此领养信息")
	}

	// 更新字段
	if req.PetName != "" {
		adopt.PetName = req.PetName
	}
	if req.PetType != "" {
		adopt.PetType = req.PetType
	}
	if req.PetAge != "" {
		adopt.PetAge = req.PetAge
	}
	if req.PetGender != 0 || adopt.PetGender != 0 {
		adopt.PetGender = req.PetGender
	}
	if req.HealthStatus != "" {
		adopt.HealthStatus = req.HealthStatus
	}
	if len(req.Photos) > 0 {
		if err := validatePhotos(req.Photos); err != nil {
			return nil, err
		}
		photosJSON, _ := json.Marshal(req.Photos)
		adopt.Photos = string(photosJSON)
	}
	if req.Description != "" {
		adopt.Description = req.Description
	}
	if req.AdoptConditions != "" {
		adopt.AdoptConditions = req.AdoptConditions
	}
	if req.Location != "" {
		adopt.Location = req.Location
	}
	if req.ContactPhone != "" {
		adopt.ContactPhone = req.ContactPhone
	}
	if req.Status != 0 || adopt.Status != 0 {
		adopt.Status = req.Status
	}

	if err := s.adoptRepo.Update(adopt); err != nil {
		return nil, err
	}

	return s.convertToAdoptInfo(adopt, userID)
}

// DeleteAdopt 删除领养信息
func (s *AdoptService) DeleteAdopt(userID, adoptID uint) error {
	adopt, err := s.adoptRepo.GetByID(adoptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("领养信息不存在")
		}
		return err
	}

	// 验证用户权限
	if adopt.UserID != userID {
		return errors.New("无权删除此领养信息")
	}

	return s.adoptRepo.Delete(adoptID)
}

// GetAdoptByID 获取领养信息详情
func (s *AdoptService) GetAdoptByID(adoptID, viewerID uint) (*types.AdoptInfo, error) {
	adopt, err := s.adoptRepo.GetByID(adoptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("领养信息不存在")
		}
		return nil, err
	}

	return s.convertToAdoptInfo(adopt, viewerID)
}

// ListAdopts 领养信息列表
func (s *AdoptService) ListAdopts(req *types.AdoptListRequest, viewerID uint) (*types.AdoptListResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	adopts, total, err := s.adoptRepo.List(req.PetType, req.Status, req.Location, req.Keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.AdoptInfo, 0, len(adopts))
	for _, adopt := range adopts {
		info, err := s.convertToAdoptInfo(&adopt, viewerID)
		if err != nil {
			continue
		}
		list = append(list, *info)
	}

	return &types.AdoptListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListMyAdopts 我的领养信息列表
func (s *AdoptService) ListMyAdopts(userID uint, page, pageSize int) (*types.AdoptListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	adopts, total, err := s.adoptRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.AdoptInfo, 0, len(adopts))
	for _, adopt := range adopts {
		info, err := s.convertToAdoptInfo(&adopt, userID)
		if err != nil {
			continue
		}
		list = append(list, *info)
	}

	return &types.AdoptListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CreateApplication 创建领养申请
func (s *AdoptService) CreateApplication(userID uint, req *types.AdoptionApplicationRequest) (*types.AdoptionApplicationInfo, error) {
	// 检查领养信息是否存在且状态正确
	adopt, err := s.adoptRepo.GetByID(req.AdoptionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("领养信息不存在")
		}
		return nil, err
	}

	if adopt.Status != 0 {
		return nil, errors.New("该宠物已不可领养")
	}

	// 不能申请自己发布的领养信息
	if adopt.UserID == userID {
		return nil, errors.New("不能申请自己发布的领养信息")
	}

	// 检查是否已经申请过
	exists, err := s.applicationRepo.Exists(req.AdoptionID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("已经申请过该领养信息")
	}

	application := &model.AdoptionApplication{
		AdoptionID:      req.AdoptionID,
		UserID:          userID,
		ApplicationInfo: req.ApplicationInfo,
		Status:          0, // 审核中
		AgreementSigned: 0, // 未签署
	}

	if err := s.applicationRepo.Create(application); err != nil {
		return nil, err
	}

	return s.convertToApplicationInfo(application)
}

// CancelApplication 取消领养申请
func (s *AdoptService) CancelApplication(userID, applicationID uint) error {
	application, err := s.applicationRepo.GetByID(applicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("申请不存在")
		}
		return err
	}

	// 验证用户权限
	if application.UserID != userID {
		return errors.New("无权取消此申请")
	}

	// 只能取消审核中的申请
	if application.Status != 0 {
		return errors.New("只能取消审核中的申请")
	}

	application.Status = 2 // 设置为已拒绝状态表示取消
	return s.applicationRepo.Update(application)
}

// ReviewApplication 审核领养申请
func (s *AdoptService) ReviewApplication(reviewedBy uint, req *types.AdoptionReviewRequest) error {
	application, err := s.applicationRepo.GetByID(req.ApplicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("申请不存在")
		}
		return err
	}

	// 检查领养信息所属用户是否为审核人
	adopt, err := s.adoptRepo.GetByID(application.AdoptionID)
	if err != nil {
		return errors.New("领养信息不存在")
	}

	if adopt.UserID != reviewedBy {
		return errors.New("无权审核此申请")
	}

	// 只能审核审核中的申请
	if application.Status != 0 {
		return errors.New("只能审核审核中的申请")
	}

	var status int
	if req.Approve {
		status = 1 // 已通过
	} else {
		status = 2 // 已拒绝
	}

	now := time.Now()
	application.Status = status
	application.ReviewedBy = reviewedBy
	application.ReviewedAt = &now

	if err := s.applicationRepo.Update(application); err != nil {
		return err
	}

	// 如果通过申请且签署了协议，更新领养信息状态
	if req.Approve && application.AgreementSigned == 1 {
		adopt.Status = 1 // 已领养
		if err := s.adoptRepo.Update(adopt); err != nil {
			return err
		}
	}

	return nil
}

// SignAgreement 签署领养协议
func (s *AdoptService) SignAgreement(userID uint, req *types.AdoptionAgreementRequest) error {
	application, err := s.applicationRepo.GetByID(req.ApplicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("申请不存在")
		}
		return err
	}

	// 验证用户权限
	if application.UserID != userID {
		return errors.New("无权签署此协议")
	}

	// 只能签署已通过的申请
	if application.Status != 1 {
		return errors.New("只能签署已通过的申请")
	}

	// 检查是否已签署
	if application.AgreementSigned == 1 {
		return errors.New("已经签署过协议")
	}

	if err := s.applicationRepo.SignAgreement(req.ApplicationID); err != nil {
		return err
	}

	// 更新领养信息状态为已领养
	adopt, err := s.adoptRepo.GetByID(application.AdoptionID)
	if err != nil {
		return err
	}

	adopt.Status = 1 // 已领养
	return s.adoptRepo.Update(adopt)
}

// ListApplicationsByAdoption 领养申请列表（按领养信息）
func (s *AdoptService) ListApplicationsByAdoption(adoptionID uint, status int, page, pageSize int) (*types.AdoptionApplicationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	applications, total, err := s.applicationRepo.ListByAdoption(adoptionID, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.AdoptionApplicationInfo, 0, len(applications))
	for _, app := range applications {
		info, err := s.convertToApplicationInfo(&app)
		if err != nil {
			continue
		}
		list = append(list, *info)
	}

	return &types.AdoptionApplicationListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListMyApplications 我的领养申请列表
func (s *AdoptService) ListMyApplications(userID uint, status int, page, pageSize int) (*types.AdoptionApplicationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	applications, total, err := s.applicationRepo.ListByUser(userID, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.AdoptionApplicationInfo, 0, len(applications))
	for _, app := range applications {
		info, err := s.convertToApplicationInfo(&app)
		if err != nil {
			continue
		}
		list = append(list, *info)
	}

	return &types.AdoptionApplicationListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateAdoptStatus 更新领养状态（管理员或发布者）
func (s *AdoptService) UpdateAdoptStatus(userID, adoptID uint, status int) error {
	adopt, err := s.adoptRepo.GetByID(adoptID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("领养信息不存在")
		}
		return err
	}

	// 验证用户权限（只有发布者可以更新）
	if adopt.UserID != userID {
		return errors.New("无权更新此领养状态")
	}

	adopt.Status = status
	return s.adoptRepo.Update(adopt)
}

// GetAdoptionStats 获取领养统计信息
func (s *AdoptService) GetAdoptionStats() (*types.AdoptionStats, error) {
	var stats types.AdoptionStats

	// 统计总数
	totalAdoptions, err := s.adoptRepo.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalAdoptions = totalAdoptions

	// 统计待领养数量
	pendingAdoptions, err := s.adoptRepo.CountByStatus(0)
	if err != nil {
		return nil, err
	}
	stats.PendingAdoptions = pendingAdoptions

	// 统计已领养数量
	adoptedAdoptions, err := s.adoptRepo.CountByStatus(1)
	if err != nil {
		return nil, err
	}
	stats.AdoptedAdoptions = adoptedAdoptions

	// 统计已取消数量
	cancelledAdoptions, err := s.adoptRepo.CountByStatus(2)
	if err != nil {
		return nil, err
	}
	stats.CancelledAdoptions = cancelledAdoptions

	// 统计申请总数
	totalApplications, err := s.applicationRepo.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalApplications = totalApplications

	return &stats, nil
}

// validatePhotos 校验照片 URL 是否合法
// 禁止以下形式入库：
//   - "placeholder-url-N" 等前端占位符
//   - "blob:..." 未上传的本地预览 URL
//   - "file:..." / "data:..." 本地路径
// 允许：http://  https://  以及以 / 开头的相对路径（静态资源）
func validatePhotos(photos []string) error {
	for i, p := range photos {
		if p == "" {
			return errors.New("第 " + strconv.Itoa(i+1) + " 张照片地址为空")
		}
		if strings.HasPrefix(p, "placeholder-url") {
			return errors.New("第 " + strconv.Itoa(i+1) + " 张照片未完成上传，请等待上传结束后再提交")
		}
		if strings.HasPrefix(p, "blob:") {
			return errors.New("第 " + strconv.Itoa(i+1) + " 张照片为本地预览地址，请先上传至服务器")
		}
		if strings.HasPrefix(p, "file:") || strings.HasPrefix(p, "data:") {
			return errors.New("第 " + strconv.Itoa(i+1) + " 张照片格式不合法，请重新上传")
		}
		if !(strings.HasPrefix(p, "http://") ||
			strings.HasPrefix(p, "https://") ||
			strings.HasPrefix(p, "/") ||
			// 允许历史数据中的裸文件名（含扩展名即可通过，作为降级兼容）
			(strings.Contains(p, ".") && !strings.ContainsAny(p, " :\\\n\r\t<>&*?|"))) {
			return errors.New("第 " + strconv.Itoa(i+1) + " 张照片地址格式不正确")
		}
	}
	return nil
}

// UploadAdoptPhoto 上传领养照片（≤5MB），返回可被前端直接访问的 URL。
// 落盘位置由存储驱动决定（本地磁盘 / 对象存储）
func (s *AdoptService) UploadAdoptPhoto(file io.Reader, originalExt string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp", "bmp":
	default:
		return "", errors.New("仅支持 jpg/jpeg/png/gif/webp/bmp 格式的图片")
	}

	return utils.CurrentStorage().Save(file, utils.SaveOptions{
		Category: utils.CategoryAdopt,
		Ext:      ext,
		MaxSize:  utils.MaxPostImageBytes,
		Kind:     "图片",
	})
}

// convertToAdoptInfo 转换领养信息模型为响应类型
func (s *AdoptService) convertToAdoptInfo(adopt *model.Adopt, viewerID uint) (*types.AdoptInfo, error) {
	// 解析照片JSON
	var photos []string
	if adopt.Photos != "" {
		if err := json.Unmarshal([]byte(adopt.Photos), &photos); err != nil {
			photos = []string{}
		}
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(adopt.UserID)
	if err != nil {
		return nil, err
	}

	// 统计申请数量
	applicationCount, _ := s.adoptRepo.CountApplications(adopt.ID)

	contact := adopt.ContactPhone
	if viewerID != adopt.UserID {
		contact = ApplyContactPrivacy(viewerID, contact)
	}

	return &types.AdoptInfo{
		ID:               adopt.ID,
		PetID:            adopt.PetID,
		UserID:           adopt.UserID,
		UserNickname:     user.Nickname,
		UserAvatar:       user.Avatar,
		PetName:          adopt.PetName,
		PetType:          adopt.PetType,
		PetAge:           adopt.PetAge,
		PetGender:        adopt.PetGender,
		HealthStatus:     adopt.HealthStatus,
		Photos:           photos,
		Description:      adopt.Description,
		AdoptConditions:  adopt.AdoptConditions,
		Location:         adopt.Location,
		ContactPhone:     contact,
		Status:           adopt.Status,
		CreatedAt:        adopt.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:        adopt.UpdatedAt.Format("2006-01-02 15:04:05"),
		ApplicationCount: int(applicationCount),
	}, nil
}

// convertToApplicationInfo 转换申请模型为响应类型
func (s *AdoptService) convertToApplicationInfo(application *model.AdoptionApplication) (*types.AdoptionApplicationInfo, error) {
	// 获取申请人信息
	user, err := s.userRepo.GetByID(application.UserID)
	if err != nil {
		return nil, err
	}

	var reviewedAt string
	if application.ReviewedAt != nil {
		reviewedAt = application.ReviewedAt.Format("2006-01-02 15:04:05")
	}

	return &types.AdoptionApplicationInfo{
		ID:              application.ID,
		AdoptionID:      application.AdoptionID,
		UserID:          application.UserID,
		UserNickname:    user.Nickname,
		UserAvatar:      user.Avatar,
		ApplicationInfo: application.ApplicationInfo,
		Status:          application.Status,
		ReviewedBy:      application.ReviewedBy,
		ReviewedAt:      reviewedAt,
		AgreementSigned: application.AgreementSigned == 1,
		CreatedAt:       application.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       application.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// CreateFollowUp 创建领养回访记录
func (s *AdoptService) CreateFollowUp(userID uint, req *types.AdoptionFollowUpCreateRequest) (*types.AdoptionFollowUpInfo, error) {
	// 验证领养信息是否存在
	adopt, err := s.adoptRepo.GetByID(req.AdoptionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("领养信息不存在")
		}
		return nil, err
	}

	// 验证申请是否存在
	application, err := s.applicationRepo.GetByID(req.ApplicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("领养申请不存在")
		}
		return nil, err
	}

	if application.AdoptionID != req.AdoptionID {
		return nil, errors.New("申请不属于该领养信息")
	}

	// 只有发布者可以创建回访记录
	if adopt.UserID != userID {
		return nil, errors.New("无权创建回访记录")
	}

	// 申请必须已通过且已签署协议
	if application.Status != 1 || application.AgreementSigned != 1 {
		return nil, errors.New("只能对已签署协议的通过申请创建回访记录")
	}

	photosJSON, _ := json.Marshal(req.Photos)
	followUp := &model.AdoptionFollowUp{
		AdoptionID:    req.AdoptionID,
		ApplicationID: req.ApplicationID,
		UserID:        userID,
		FollowUpType:  req.FollowUpType,
		Content:       req.Content,
		Photos:        string(photosJSON),
		Status:        req.Status,
	}

	if err := s.followUpRepo.Create(followUp); err != nil {
		return nil, err
	}

	return s.convertToFollowUpInfo(followUp)
}

// ListFollowUpsByAdoption 获取领养回访记录列表
func (s *AdoptService) ListFollowUpsByAdoption(adoptionID uint, page, pageSize int) (*types.AdoptionFollowUpListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	followUps, total, err := s.followUpRepo.ListByAdoption(adoptionID, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.AdoptionFollowUpInfo, 0, len(followUps))
	for _, fu := range followUps {
		info, err := s.convertToFollowUpInfo(&fu)
		if err != nil {
			continue
		}
		list = append(list, *info)
	}

	return &types.AdoptionFollowUpListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// convertToFollowUpInfo 转换回访记录模型为响应类型
func (s *AdoptService) convertToFollowUpInfo(followUp *model.AdoptionFollowUp) (*types.AdoptionFollowUpInfo, error) {
	user, err := s.userRepo.GetByID(followUp.UserID)
	if err != nil {
		return nil, err
	}

	var photos []string
	if followUp.Photos != "" {
		if err := json.Unmarshal([]byte(followUp.Photos), &photos); err != nil {
			photos = []string{}
		}
	}

	return &types.AdoptionFollowUpInfo{
		ID:            followUp.ID,
		AdoptionID:    followUp.AdoptionID,
		ApplicationID: followUp.ApplicationID,
		UserID:        followUp.UserID,
		UserNickname:  user.Nickname,
		UserAvatar:    user.Avatar,
		FollowUpType:  followUp.FollowUpType,
		Content:       followUp.Content,
		Photos:        photos,
		Status:        followUp.Status,
		CreatedAt:     followUp.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
