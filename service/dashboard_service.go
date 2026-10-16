package service

import (
	"pet-community/repository"
	"pet-community/types"
	"time"
)

// DashboardService 运营驾驶舱服务 - 实时概览与待办聚合
type DashboardService struct {
	adminRepo  *repository.AdminRepository
	userRepo   *repository.UserRepository
	postRepo   *repository.PostRepository
	petRepo    *repository.PetRepository
	statsRepo  *repository.StatsRepository
	reportRepo *repository.ReportRepository
}

func NewDashboardService(
	adminRepo *repository.AdminRepository,
	userRepo *repository.UserRepository,
	postRepo *repository.PostRepository,
	petRepo *repository.PetRepository,
	statsRepo *repository.StatsRepository,
	reportRepo *repository.ReportRepository,
) *DashboardService {
	return &DashboardService{
		adminRepo:  adminRepo,
		userRepo:   userRepo,
		postRepo:   postRepo,
		petRepo:    petRepo,
		statsRepo:  statsRepo,
		reportRepo: reportRepo,
	}
}

// GetDashboardOverview 获取驾驶舱概览数据（实时指标 + 近 7 日双趋势）
func (s *DashboardService) GetDashboardOverview() (*types.DashboardOverview, error) {
	summary, err := s.getDashboardSummary()
	if err != nil {
		return nil, err
	}

	todayTrend, err := s.getUserTrend()
	if err != nil {
		return nil, err
	}

	postTrend, err := s.getPostTrend()
	if err != nil {
		return nil, err
	}

	return &types.DashboardOverview{
		Summary:    *summary,
		TodayTrend: *todayTrend,
		PostTrend:  *postTrend,
	}, nil
}

func (s *DashboardService) getDashboardSummary() (*types.DashboardSummary, error) {
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	totalUsers, _ := s.adminRepo.CountUsersByStatus(1)
	todayNewUsers, _ := s.adminRepo.CountNewUsersByDate(today)
	yesterdayNewUsers, _ := s.adminRepo.CountNewUsersByDate(yesterday)

	todayActiveUsers, _ := s.adminRepo.CountActiveUsersByDate(today)
	yesterdayActiveUsers, _ := s.adminRepo.CountActiveUsersByDate(yesterday)

	totalPosts, _ := s.adminRepo.CountAllPosts()
	todayPosts, _ := s.adminRepo.CountPostsByCreatedDate(today)
	yesterdayPosts, _ := s.adminRepo.CountPostsByCreatedDate(yesterday)

	todayLikes, _ := s.adminRepo.CountLikesByDate(today)
	yesterdayLikes, _ := s.adminRepo.CountLikesByDate(yesterday)
	totalLikes, _ := s.adminRepo.CountAllLikes()

	todayComments, _ := s.adminRepo.CountCommentsByDate(today)
	yesterdayComments, _ := s.adminRepo.CountCommentsByDate(yesterday)

	todayFavorites, _ := s.adminRepo.CountFavoritesByDate(today)
	yesterdayFavorites, _ := s.adminRepo.CountFavoritesByDate(yesterday)
	todayShares, _ := s.adminRepo.CountSharesByDate(today)
	yesterdayShares, _ := s.adminRepo.CountSharesByDate(yesterday)

	todayRegistrations, _ := s.adminRepo.CountRegistrationsByDate(today)
	yesterdayRegistrations, _ := s.adminRepo.CountRegistrationsByDate(yesterday)
	todayVotes, _ := s.adminRepo.CountVotesByDate(today)
	yesterdayVotes, _ := s.adminRepo.CountVotesByDate(yesterday)

	pendingReviews, _ := s.adminRepo.CountPostsByStatus(0)
	pendingAdoptions, _ := s.adminRepo.CountAdoptionApplicationsByStatus(0)
	pendingRegistrations, _ := s.adminRepo.CountRegistrationsByStatus(0)

	totalPets, _ := s.adminRepo.CountAllPets()
	ongoingActivities, _ := s.adminRepo.CountOngoingActivities()
	ongoingLostPets, _ := s.adminRepo.CountLostPetsByStatus(1)

	adViews, _ := s.adminRepo.SumAdvertisementViews()
	adClicks, _ := s.adminRepo.SumAdvertisementClicks()

	var pendingReports, handledReports int64
	if s.reportRepo != nil {
		pendingReports, _ = s.reportRepo.CountByStatus(0)
		handledReports, _ = s.reportRepo.CountByStatus(1)
	} else {
		pendingReports, _ = s.adminRepo.CountReportsByStatus(0)
		handledReports, _ = s.adminRepo.CountReportsByStatus(1)
	}
	reportTotal := pendingReports + handledReports
	reportHandleRate := 0.0
	if reportTotal > 0 {
		reportHandleRate = float64(handledReports) / float64(reportTotal) * 100
		reportHandleRate = float64(int(reportHandleRate*10)) / 10
	}

	return &types.DashboardSummary{
		TotalUsers:             int(totalUsers),
		TodayNewUsers:          int(todayNewUsers),
		YesterdayNewUsers:      int(yesterdayNewUsers),
		TodayActiveUsers:       int(todayActiveUsers),
		YesterdayActiveUsers:   int(yesterdayActiveUsers),
		TotalPosts:             int(totalPosts),
		TodayPosts:             int(todayPosts),
		YesterdayPosts:         int(yesterdayPosts),
		TodayLikes:             int(todayLikes),
		YesterdayLikes:         int(yesterdayLikes),
		TodayComments:          int(todayComments),
		YesterdayComments:      int(yesterdayComments),
		TodayFavorites:         int(todayFavorites),
		YesterdayFavorites:     int(yesterdayFavorites),
		TodayShares:            int(todayShares),
		YesterdayShares:        int(yesterdayShares),
		TodayRegistrations:     int(todayRegistrations),
		YesterdayRegistrations: int(yesterdayRegistrations),
		TodayVotes:             int(todayVotes),
		YesterdayVotes:         int(yesterdayVotes),
		PendingReviews:         int(pendingReviews),
		PendingRegistrations:   int(pendingRegistrations),
		PendingAdoptions:       int(pendingAdoptions),
		TotalPets:              int(totalPets),
		TotalLikes:             int(totalLikes),
		OngoingActivities:      int(ongoingActivities),
		OngoingLostPets:        int(ongoingLostPets),
		AdTotalViews:           int(adViews),
		AdTotalClicks:          int(adClicks),
		ReportHandleRate:       reportHandleRate,
	}, nil
}

func (s *DashboardService) getUserTrend() (*types.TrendData, error) {
	data := make([]types.ChartDataPoint, 0, 7)
	endTime := time.Now()
	for i := 6; i >= 0; i-- {
		dateStr := endTime.AddDate(0, 0, -i).Format("2006-01-02")
		count, _ := s.adminRepo.CountNewUsersByDate(dateStr)
		data = append(data, types.ChartDataPoint{Label: dateStr, Value: int(count)})
	}
	return &types.TrendData{
		Title: "近 7 日用户增长",
		XKey:  "日期",
		YKey:  "新增用户",
		Data:  data,
	}, nil
}

func (s *DashboardService) getPostTrend() (*types.TrendData, error) {
	data := make([]types.ChartDataPoint, 0, 7)
	endTime := time.Now()
	for i := 6; i >= 0; i-- {
		dateStr := endTime.AddDate(0, 0, -i).Format("2006-01-02")
		count, _ := s.adminRepo.CountPostsByCreatedDate(dateStr)
		data = append(data, types.ChartDataPoint{Label: dateStr, Value: int(count)})
	}
	return &types.TrendData{
		Title: "近 7 日内容发布",
		XKey:  "日期",
		YKey:  "新增帖子",
		Data:  data,
	}, nil
}

// GetTodoAggregate 获取待办聚合数据（仅可处理队列）
func (s *DashboardService) GetTodoAggregate() (*types.TodoAggregate, error) {
	items := make([]types.TodoItem, 0)

	pendingPosts, _ := s.adminRepo.CountPostsByStatus(0)
	items = append(items, types.TodoItem{
		Key: "pending_posts", Label: "待审核帖子", Count: int(pendingPosts), Link: "/admin/reviews",
	})

	var pendingReports int64
	if s.reportRepo != nil {
		pendingReports, _ = s.reportRepo.CountByStatus(0)
	} else {
		pendingReports, _ = s.adminRepo.CountReportsByStatus(0)
	}
	items = append(items, types.TodoItem{
		Key: "pending_reports", Label: "待处理举报", Count: int(pendingReports), Link: "/admin/reports",
	})

	pendingAds, _ := s.adminRepo.CountAdvertisementsByStatus(0)
	items = append(items, types.TodoItem{
		Key: "pending_ads", Label: "待审核广告", Count: int(pendingAds), Link: "/admin/ads",
	})

	// 领养申请一审由发布者处理，不进入运营驾驶舱待办（总览 §4.10）

	pendingRegistrations, _ := s.adminRepo.CountRegistrationsByStatus(0)
	items = append(items, types.TodoItem{
		Key: "pending_registrations", Label: "待审活动报名", Count: int(pendingRegistrations), Link: "/admin/activities?focus=registrations",
	})

	total := 0
	for _, item := range items {
		total += item.Count
	}

	return &types.TodoAggregate{Items: items, Total: total}, nil
}
