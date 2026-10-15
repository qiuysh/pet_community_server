package service

import (
	"errors"
	"strconv"
	"time"

	"pet-community/cache"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
)

type AdService struct {
	adRepo    *repository.AdRepository
	adminRepo *repository.AdminRepository
}

func NewAdService(adRepo *repository.AdRepository, adminRepo *repository.AdminRepository) *AdService {
	return &AdService{adRepo: adRepo, adminRepo: adminRepo}
}

func (s *AdService) CreatePosition(name, location, size string, status int) (*model.AdPosition, error) {
	position := &model.AdPosition{
		Name:     name,
		Location: location,
		Size:     size,
		Status:   status,
	}

	if position.Status == 0 {
		position.Status = 1
	}

	return position, s.adRepo.CreatePosition(position)
}

func (s *AdService) GetPositionByID(id uint) (*model.AdPosition, error) {
	return s.adRepo.GetPositionByID(id)
}

func (s *AdService) ListPositions(page, pageSize int) ([]model.AdPosition, int64, error) {
	return s.adRepo.ListPositions(page, pageSize)
}

func (s *AdService) UpdatePosition(id uint, name, location, size string, status int) (*model.AdPosition, error) {
	position, err := s.adRepo.GetPositionByID(id)
	if err != nil {
		return nil, errors.New("广告位不存在")
	}

	position.Name = name
	position.Location = location
	position.Size = size
	position.Status = status

	err = s.adRepo.UpdatePosition(position)
	if err == nil {
		cache.InvalidateActiveAds()
	}
	return position, err
}

func (s *AdService) DeletePosition(id uint) error {
	if _, err := s.adRepo.GetPositionByID(id); err != nil {
		return errors.New("广告位不存在")
	}

	err := s.adRepo.DeletePosition(id)
	if err == nil {
		cache.InvalidateActiveAds()
	}
	return err
}

func (s *AdService) GetActivePositions(location string) ([]model.AdPosition, error) {
	return s.adRepo.GetActivePositions(location)
}

func (s *AdService) CreateAd(positionID uint, title, imageURL, linkURL, startTime, endTime string, createdBy uint) (*model.Advertisement, error) {
	_, err := s.adRepo.GetPositionByID(positionID)
	if err != nil {
		return nil, errors.New("广告位不存在")
	}

	start, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		return nil, errors.New("开始时间格式错误")
	}

	end, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		return nil, errors.New("结束时间格式错误")
	}

	if end.Before(start) {
		return nil, errors.New("结束时间不能早于开始时间")
	}

	ad := &model.Advertisement{
		PositionID: positionID,
		Title:      title,
		ImageURL:   imageURL,
		LinkURL:    linkURL,
		StartTime:  start,
		EndTime:    end,
		Status:     0,
		CreatedBy:  createdBy,
	}

	err = s.adRepo.CreateAd(ad)
	if err == nil {
		cache.InvalidateActiveAds()
	}
	return ad, err
}

func (s *AdService) GetAdByID(id uint) (*model.Advertisement, error) {
	return s.adRepo.GetAdByID(id)
}

func (s *AdService) ListAds(positionID uint, status int, page, pageSize int) ([]model.Advertisement, int64, error) {
	return s.adRepo.ListAds(positionID, status, page, pageSize)
}

func (s *AdService) UpdateAd(id uint, positionID uint, title, imageURL, linkURL, startTime, endTime string) (*model.Advertisement, error) {
	ad, err := s.adRepo.GetAdByID(id)
	if err != nil {
		return nil, errors.New("广告不存在")
	}

	if positionID > 0 {
		_, err := s.adRepo.GetPositionByID(positionID)
		if err != nil {
			return nil, errors.New("广告位不存在")
		}
		ad.PositionID = positionID
	}

	if title != "" {
		ad.Title = title
	}
	if imageURL != "" {
		ad.ImageURL = imageURL
	}
	if linkURL != "" {
		ad.LinkURL = linkURL
	}

	if startTime != "" {
		start, err := time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			return nil, errors.New("开始时间格式错误")
		}
		ad.StartTime = start
	}

	if endTime != "" {
		end, err := time.Parse("2006-01-02 15:04:05", endTime)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		ad.EndTime = end
	}

	if ad.EndTime.Before(ad.StartTime) {
		return nil, errors.New("结束时间不能早于开始时间")
	}

	err = s.adRepo.UpdateAd(ad)
	if err == nil {
		cache.InvalidateActiveAds()
	}
	return ad, err
}

func (s *AdService) DeleteAd(id uint) error {
	if _, err := s.adRepo.GetAdByID(id); err != nil {
		return errors.New("广告不存在")
	}

	err := s.adRepo.DeleteAd(id)
	if err == nil {
		cache.InvalidateActiveAds()
	}
	return err
}

func (s *AdService) ReviewAd(id uint, status int, adminID uint, ip string) error {
	ad, err := s.adRepo.GetAdByID(id)
	if err != nil {
		return errors.New("广告不存在")
	}

	if ad.Status == status {
		return errors.New("广告状态未发生变化")
	}

	if err := s.adRepo.UpdateAdStatus(id, status); err != nil {
		return err
	}
	cache.InvalidateActiveAds()

	action := ""
	if status == 1 {
		action = "ad_approve"
	} else if status == 2 {
		action = "ad_reject"
	} else {
		action = "ad_pending"
	}

	log := &model.AdminOperationLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: "advertisement",
		TargetID:   id,
		Detail:     "广告审核: ID=" + strconv.FormatUint(uint64(id), 10) + ", 状态=" + strconv.Itoa(status),
		IP:         ip,
	}
	return s.adminRepo.CreateOperationLog(log)
}

func (s *AdService) IncrementViewCount(id uint) error {
	if cache.IncrAdView(id) {
		return nil
	}
	return s.adRepo.IncrementViewCount(id)
}

func (s *AdService) IncrementClickCount(id uint) error {
	if cache.IncrAdClick(id) {
		return nil
	}
	return s.adRepo.IncrementClickCount(id)
}

// FlushBufferedCounters 将 Redis 中的广告曝光/点击缓冲刷入 MySQL
func (s *AdService) FlushBufferedCounters() {
	cache.FlushAdCounters(func(id uint, views, clicks int64) error {
		if err := s.adRepo.AddViewCount(id, views); err != nil {
			return err
		}
		return s.adRepo.AddClickCount(id, clicks)
	})
}

func (s *AdService) GetActiveAds(location string) ([]model.Advertisement, error) {
	if ads, ok := cache.GetActiveAds(location); ok {
		return ads, nil
	}
	ads, err := s.adRepo.GetActiveAds(location)
	if err != nil {
		return nil, err
	}
	cache.SetActiveAds(location, ads)
	return ads, nil
}

func (s *AdService) GetAdStats(startDate, endDate string) (*types.AdStatsResponse, error) {
	var startTime, endTime time.Time
	var err error

	if startDate == "" {
		startTime = time.Now().AddDate(0, 0, -7)
	} else {
		startTime, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, errors.New("开始日期格式错误")
		}
	}

	if endDate == "" {
		endTime = time.Now()
	} else {
		endTime, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, errors.New("结束日期格式错误")
		}
	}

	totalViews, totalClicks, err := s.adRepo.GetAdStats(startTime, endTime)
	if err != nil {
		return nil, err
	}

	ctr := 0.0
	if totalViews > 0 {
		ctr = float64(totalClicks) / float64(totalViews) * 100
	}

	statsByDay, err := s.adRepo.GetAdStatsByDay(startTime, endTime)
	if err != nil {
		return nil, err
	}

	dayData := make([]types.ChartDataPoint, 0)
	for _, stat := range statsByDay {
		dayData = append(dayData, types.ChartDataPoint{
			Label: stat.Date,
			Value: stat.ViewCount,
		})
	}

	statsByPosition, err := s.adRepo.GetAdStatsByPosition()
	if err != nil {
		return nil, err
	}

	positionData := make([]types.ChartDataPoint, 0)
	for _, stat := range statsByPosition {
		positionData = append(positionData, types.ChartDataPoint{
			Label: stat.PositionName,
			Value: stat.ViewCount,
		})
	}

	return &types.AdStatsResponse{
		TotalViews:      totalViews,
		TotalClicks:     totalClicks,
		CTR:             ctr,
		StatsByDay:      dayData,
		StatsByPosition: positionData,
	}, nil
}
