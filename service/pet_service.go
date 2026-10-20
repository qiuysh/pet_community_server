package service

import (
	"errors"
	"io"
	"strings"

	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
	"pet-community/utils"
)

type PetService struct {
	petRepo *repository.PetRepository
}

func NewPetService(petRepo *repository.PetRepository) *PetService {
	return &PetService{petRepo: petRepo}
}

func (s *PetService) CreatePet(userID uint, req *types.PetCreateRequest) (*types.PetInfo, error) {
	pet := &model.Pet{
		UserID:       userID,
		Name:         FilterSensitiveText(req.Name),
		Avatar:       req.Avatar,
		PetType:      req.PetType,
		BreedID:      req.BreedID,
		BreedName:    req.BreedName,
		Gender:       req.Gender,
		Birthday:     req.Birthday,
		Weight:       req.Weight,
		Personality:  req.Personality,
		Bio:          req.Bio,
		IsSterilized: req.IsSterilized,
		IsVaccinated: req.IsVaccinated,
		Status:       1,
	}

	if req.BreedID > 0 && req.BreedName == "" {
		breed, err := s.petRepo.GetBreedByID(req.BreedID)
		if err == nil {
			pet.BreedName = breed.Name
		}
	}

	if err := s.petRepo.Create(pet); err != nil {
		return nil, err
	}

	return convertToPetInfo(pet), nil
}

func (s *PetService) UpdatePet(userID uint, petID uint, req *types.PetUpdateRequest) (*types.PetInfo, error) {
	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return nil, errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return nil, errors.New("无权限修改")
	}

	if req.Name != "" {
		pet.Name = FilterSensitiveText(req.Name)
	}
	if req.Avatar != "" {
		pet.Avatar = req.Avatar
	}
	if req.PetType > 0 {
		pet.PetType = req.PetType
	}
	if req.BreedID > 0 {
		pet.BreedID = req.BreedID
		if req.BreedName != "" {
			pet.BreedName = req.BreedName
		} else {
			breed, err := s.petRepo.GetBreedByID(req.BreedID)
			if err == nil {
				pet.BreedName = breed.Name
			}
		}
	}
	if req.BreedName != "" && req.BreedID == 0 {
		pet.BreedName = req.BreedName
	}
	if req.Gender > 0 {
		pet.Gender = req.Gender
	}
	if req.Birthday != nil {
		pet.Birthday = req.Birthday
	}
	if req.Weight > 0 {
		pet.Weight = req.Weight
	}
	if req.Personality != "" {
		pet.Personality = req.Personality
	}
	if req.Bio != "" {
		pet.Bio = req.Bio
	}
	if req.IsSterilized != nil {
		pet.IsSterilized = *req.IsSterilized
	}
	if req.IsVaccinated != nil {
		pet.IsVaccinated = *req.IsVaccinated
	}
	if req.Status >= 0 {
		pet.Status = req.Status
	}

	if err := s.petRepo.Update(pet); err != nil {
		return nil, err
	}

	return convertToPetInfo(pet), nil
}

func (s *PetService) GetPetByID(id uint) (*types.PetInfo, error) {
	pet, err := s.petRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("宠物不存在")
	}

	go s.petRepo.IncrementViewCount(id)

	return convertToPetInfo(pet), nil
}

func (s *PetService) DeletePet(userID uint, petID uint) error {
	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return errors.New("无权限删除")
	}

	pet.Status = 0
	return s.petRepo.Update(pet)
}

func (s *PetService) GetUserPets(userID uint) ([]types.PetInfo, error) {
	pets, err := s.petRepo.GetUserPets(userID)
	if err != nil {
		return nil, err
	}

	result := make([]types.PetInfo, len(pets))
	for i, p := range pets {
		result[i] = *convertToPetInfo(&p)
	}
	return result, nil
}

func (s *PetService) ListPets(req *types.PetListRequest) (*types.PetListResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	pets, total, err := s.petRepo.List(req.UserID, req.PetType, req.Keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.PetInfo, len(pets))
	for i, p := range pets {
		list[i] = *convertToPetInfo(&p)
	}

	return &types.PetListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PetService) GetBreeds(petType int) ([]types.PetBreedInfo, error) {
	breeds, err := s.petRepo.ListBreeds(petType)
	if err != nil {
		return nil, err
	}

	result := make([]types.PetBreedInfo, len(breeds))
	for i, b := range breeds {
		result[i] = types.PetBreedInfo{
			ID:          b.ID,
			PetType:     b.PetType,
			Name:        b.Name,
			EnglishName: b.EnglishName,
			Avatar:      b.Avatar,
			SortOrder:   b.SortOrder,
		}
	}
	return result, nil
}

// UploadAvatar 上传宠物头像（≤2MB），返回可访问 URL
// 落盘位置由存储驱动决定（本地磁盘 / 对象存储）
func (s *PetService) UploadAvatar(file io.Reader, originalExt string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	switch ext {
	case "jpg", "jpeg", "png", "webp":
	default:
		return "", errors.New("仅支持 jpg/jpeg/png/webp 格式的头像")
	}

	return utils.CurrentStorage().Save(file, utils.SaveOptions{
		Category: utils.CategoryPetAvatar,
		Ext:      ext,
		MaxSize:  utils.MaxAvatarBytes,
		Kind:     "头像",
	})
}

// UploadAlbumPhoto 上传宠物相册图片（≤5MB），返回可访问 URL
func (s *PetService) UploadAlbumPhoto(file io.Reader, originalExt string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp", "bmp":
	default:
		return "", errors.New("仅支持 jpg/jpeg/png/gif/webp/bmp 格式的图片")
	}

	return utils.CurrentStorage().Save(file, utils.SaveOptions{
		Category: utils.CategoryPetAlbum,
		Ext:      ext,
		MaxSize:  utils.MaxPostImageBytes,
		Kind:     "图片",
	})
}

// UploadAlbumVideo 上传宠物相册视频（MP4，≤50MB）
func (s *PetService) UploadAlbumVideo(file io.Reader, originalExt string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	if ext != "mp4" {
		return "", errors.New("相册视频仅支持 MP4 格式")
	}

	return utils.CurrentStorage().Save(file, utils.SaveOptions{
		Category: utils.CategoryPetAlbum,
		Ext:      ext,
		MaxSize:  utils.MaxVideoBytes,
		Kind:     "视频",
	})
}

func (s *PetService) AddAlbumPhoto(userID, petID uint, req *types.PetAlbumCreateRequest) (*types.PetAlbumInfo, error) {
	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return nil, errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return nil, errors.New("无权限操作")
	}

	mediaType := req.MediaType
	if mediaType == "" {
		mediaType = "image"
	}

	sortOrder := req.SortOrder
	if sortOrder == 0 {
		maxOrder, _ := s.petRepo.MaxAlbumSortOrder(petID)
		sortOrder = maxOrder + 1
	}

	photo := &model.PetAlbum{
		PetID:     petID,
		ImageURL:  req.ImageURL,
		ThumbURL:  req.ThumbURL,
		Title:     req.Title,
		MediaType: mediaType,
		SortOrder: sortOrder,
	}

	if err := s.petRepo.AddAlbumPhoto(photo); err != nil {
		return nil, err
	}

	return toAlbumInfo(photo), nil
}

func (s *PetService) GetPetAlbum(petID uint) ([]types.PetAlbumInfo, error) {
	photos, err := s.petRepo.GetPetAlbum(petID)
	if err != nil {
		return nil, err
	}

	result := make([]types.PetAlbumInfo, len(photos))
	for i, p := range photos {
		result[i] = *toAlbumInfo(&p)
	}
	return result, nil
}

func (s *PetService) ReorderAlbum(userID, petID uint, ids []uint) error {
	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return errors.New("无权限操作")
	}
	if len(ids) == 0 {
		return errors.New("排序列表不能为空")
	}

	existing, err := s.petRepo.GetPetAlbum(petID)
	if err != nil {
		return err
	}
	if len(ids) != len(existing) {
		return errors.New("排序列表与相册条目数量不一致")
	}
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			return errors.New("排序列表存在重复项")
		}
		seen[id] = struct{}{}
	}
	for _, p := range existing {
		if _, ok := seen[p.ID]; !ok {
			return errors.New("排序列表缺少相册条目")
		}
	}

	if err := s.petRepo.UpdateAlbumSortOrders(petID, ids); err != nil {
		return errors.New("更新排序失败")
	}
	return nil
}

func (s *PetService) DeleteAlbumPhoto(userID, petID, photoID uint) error {
	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return errors.New("无权限操作")
	}

	return s.petRepo.DeleteAlbumPhoto(photoID, petID)
}

func toAlbumInfo(photo *model.PetAlbum) *types.PetAlbumInfo {
	mediaType := photo.MediaType
	if mediaType == "" {
		mediaType = "image"
	}
	return &types.PetAlbumInfo{
		ID:        photo.ID,
		PetID:     photo.PetID,
		ImageURL:  photo.ImageURL,
		ThumbURL:  photo.ThumbURL,
		Title:     photo.Title,
		MediaType: mediaType,
		SortOrder: photo.SortOrder,
		CreatedAt: photo.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *PetService) CreateGrowthRecord(userID, petID uint, req *types.PetGrowthRecordCreateRequest) (*types.PetGrowthRecordInfo, error) {
	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return nil, errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return nil, errors.New("无权限操作")
	}

	imagesStr := ""
	if len(req.Images) > 0 {
		imagesStr = strings.Join(req.Images, ",")
	}

	record := &model.PetGrowthRecord{
		PetID:      petID,
		RecordDate: req.RecordDate,
		Weight:     req.Weight,
		Height:     req.Height,
		Title:      req.Title,
		Content:    req.Content,
		Images:     imagesStr,
	}

	if err := s.petRepo.CreateGrowthRecord(record); err != nil {
		return nil, err
	}

	return convertToGrowthRecordInfo(record), nil
}

func (s *PetService) UpdateGrowthRecord(userID, petID, recordID uint, req *types.PetGrowthRecordUpdateRequest) (*types.PetGrowthRecordInfo, error) {
	record, err := s.petRepo.GetGrowthRecordByID(recordID)
	if err != nil {
		return nil, errors.New("记录不存在")
	}
	if record.PetID != petID {
		return nil, errors.New("记录与宠物不匹配")
	}

	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return nil, errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return nil, errors.New("无权限操作")
	}

	if !req.RecordDate.IsZero() {
		record.RecordDate = req.RecordDate
	}
	if req.Weight > 0 {
		record.Weight = req.Weight
	}
	if req.Height > 0 {
		record.Height = req.Height
	}
	if req.Title != "" {
		record.Title = req.Title
	}
	if req.Content != "" {
		record.Content = req.Content
	}
	if req.Images != nil {
		record.Images = strings.Join(req.Images, ",")
	}

	if err := s.petRepo.UpdateGrowthRecord(record); err != nil {
		return nil, err
	}

	return convertToGrowthRecordInfo(record), nil
}

func (s *PetService) DeleteGrowthRecord(userID, petID, recordID uint) error {
	record, err := s.petRepo.GetGrowthRecordByID(recordID)
	if err != nil {
		return errors.New("记录不存在")
	}
	if record.PetID != petID {
		return errors.New("记录与宠物不匹配")
	}

	pet, err := s.petRepo.GetByID(petID)
	if err != nil {
		return errors.New("宠物不存在")
	}
	if pet.UserID != userID {
		return errors.New("无权限操作")
	}

	return s.petRepo.DeleteGrowthRecord(recordID, petID)
}

func (s *PetService) GetGrowthRecords(petID uint, page, pageSize int) ([]types.PetGrowthRecordInfo, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	records, total, err := s.petRepo.GetGrowthRecords(petID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	result := make([]types.PetGrowthRecordInfo, len(records))
	for i, r := range records {
		result[i] = *convertToGrowthRecordInfo(&r)
	}
	return result, total, nil
}

// UploadLostPetPhoto 上传寻宠启事图片（≤5MB）
func (s *PetService) UploadLostPetPhoto(file io.Reader, originalExt string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(originalExt, ".")))
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp", "bmp":
	default:
		return "", errors.New("仅支持 jpg/jpeg/png/gif/webp/bmp 格式的图片")
	}

	return utils.CurrentStorage().Save(file, utils.SaveOptions{
		Category: utils.CategoryLostPet,
		Ext:      ext,
		MaxSize:  utils.MaxPostImageBytes,
		Kind:     "图片",
	})
}

func (s *PetService) CreateLostPet(userID uint, req *types.LostPetCreateRequest) (*types.LostPetInfo, error) {
	imagesStr := ""
	if len(req.Images) > 0 {
		imagesStr = strings.Join(req.Images, ",")
	}

	lostPet := &model.LostPet{
		UserID:       userID,
		PetID:        req.PetID,
		PetName:      req.PetName,
		PetType:      req.PetType,
		BreedID:      req.BreedID,
		BreedName:    req.BreedName,
		Gender:       req.Gender,
		Age:          req.Age,
		Description:  req.Description,
		Images:       imagesStr,
		LostLocation: req.LostLocation,
		LostTime:     req.LostTime,
		Contact:      req.Contact,
		Status:       1,
	}

	if err := s.petRepo.CreateLostPet(lostPet); err != nil {
		return nil, err
	}

	return convertToLostPetInfo(lostPet, userID), nil
}

func (s *PetService) GetLostPetByID(id, viewerID uint) (*types.LostPetInfo, error) {
	lostPet, err := s.petRepo.GetLostPetByID(id)
	if err != nil {
		return nil, errors.New("寻宠信息不存在")
	}

	go s.petRepo.IncrementLostPetViewCount(id)

	return convertToLostPetInfo(lostPet, viewerID), nil
}

func (s *PetService) ListLostPets(req *types.LostPetListRequest, viewerID uint) (*types.LostPetListResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	lostPets, total, err := s.petRepo.ListLostPets(req.PetType, req.LostLocation, req.Status, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]types.LostPetInfo, len(lostPets))
	for i, p := range lostPets {
		list[i] = *convertToLostPetInfo(&p, viewerID)
	}

	return &types.LostPetListResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PetService) UpdateLostPetStatus(userID, id uint, status int) error {
	lostPet, err := s.petRepo.GetLostPetByID(id)
	if err != nil {
		return errors.New("寻宠信息不存在")
	}
	if lostPet.UserID != userID {
		return errors.New("无权限操作")
	}

	lostPet.Status = status
	return s.petRepo.UpdateLostPet(lostPet)
}

func convertToPetInfo(pet *model.Pet) *types.PetInfo {
	return &types.PetInfo{
		ID:           pet.ID,
		UserID:       pet.UserID,
		Name:         pet.Name,
		Avatar:       pet.Avatar,
		PetType:      pet.PetType,
		BreedID:      pet.BreedID,
		BreedName:    pet.BreedName,
		Gender:       pet.Gender,
		Birthday:     pet.Birthday,
		Weight:       pet.Weight,
		Personality:  pet.Personality,
		Bio:          pet.Bio,
		IsSterilized: pet.IsSterilized,
		IsVaccinated: pet.IsVaccinated,
		Status:       pet.Status,
		ViewCount:    pet.ViewCount,
		CreatedAt:    pet.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func convertToGrowthRecordInfo(record *model.PetGrowthRecord) *types.PetGrowthRecordInfo {
	images := []string{}
	if record.Images != "" {
		images = strings.Split(record.Images, ",")
	}
	return &types.PetGrowthRecordInfo{
		ID:         record.ID,
		PetID:      record.PetID,
		RecordDate: record.RecordDate,
		Weight:     record.Weight,
		Height:     record.Height,
		Title:      record.Title,
		Content:    record.Content,
		Images:     images,
		CreatedAt:  record.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func convertToLostPetInfo(lostPet *model.LostPet, viewerID uint) *types.LostPetInfo {
	images := []string{}
	if lostPet.Images != "" {
		images = strings.Split(lostPet.Images, ",")
	}
	contact := lostPet.Contact
	// 发布者本人始终可见；登录用户可见；游客脱敏（总览 §4.11）
	if viewerID != lostPet.UserID {
		contact = ApplyContactPrivacy(viewerID, contact)
	}
	return &types.LostPetInfo{
		ID:           lostPet.ID,
		UserID:       lostPet.UserID,
		PetID:        lostPet.PetID,
		PetName:      lostPet.PetName,
		PetType:      lostPet.PetType,
		BreedID:      lostPet.BreedID,
		BreedName:    lostPet.BreedName,
		Gender:       lostPet.Gender,
		Age:          lostPet.Age,
		Description:  lostPet.Description,
		Images:       images,
		LostLocation: lostPet.LostLocation,
		LostTime:     lostPet.LostTime,
		Contact:      contact,
		Status:       lostPet.Status,
		ViewCount:    lostPet.ViewCount,
		CreatedAt:    lostPet.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
