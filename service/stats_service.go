package service

import (
	"fmt"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/types"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type StatsService struct {
	statsRepo    *repository.StatsRepository
	userRepo     *repository.UserRepository
	postRepo     *repository.PostRepository
	commentRepo  *repository.CommentRepository
	likeRepo     *repository.LikeRepository
	favoriteRepo *repository.FavoriteRepository
	shareRepo    *repository.ShareRepository
	topicRepo    *repository.TopicRepository
}

func NewStatsService(
	statsRepo *repository.StatsRepository,
	userRepo *repository.UserRepository,
	postRepo *repository.PostRepository,
	commentRepo *repository.CommentRepository,
	likeRepo *repository.LikeRepository,
	favoriteRepo *repository.FavoriteRepository,
	shareRepo *repository.ShareRepository,
	topicRepo *repository.TopicRepository,
) *StatsService {
	return &StatsService{
		statsRepo:    statsRepo,
		userRepo:     userRepo,
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		likeRepo:     likeRepo,
		favoriteRepo: favoriteRepo,
		shareRepo:    shareRepo,
		topicRepo:    topicRepo,
	}
}

func (s *StatsService) GetOverviewStats() (*types.OverviewStats, error) {
	today := time.Now().Format("2006-01-02")
	stats, err := s.statsRepo.FindByDate(today)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err == gorm.ErrRecordNotFound || stats == nil {
		return s.calculateOverviewStats(today), nil
	}

	return &types.OverviewStats{
		UserCount:       stats.UserCount,
		NewUserCount:    stats.NewUserCount,
		PostCount:       stats.PostCount,
		NewPostCount:    stats.NewPostCount,
		CommentCount:    stats.CommentCount,
		NewCommentCount: stats.NewCommentCount,
		LikeCount:       stats.LikeCount,
		NewLikeCount:    stats.NewLikeCount,
	}, nil
}

func (s *StatsService) calculateOverviewStats(date string) *types.OverviewStats {
	userCount, _ := s.userRepo.Count()
	newUserCount, _ := s.userRepo.CountByDate(date)
	postCount, _ := s.postRepo.CountAll()
	newPostCount, _ := s.postRepo.CountByDate(date)
	commentCount, _ := s.commentRepo.CountAll()
	newCommentCount, _ := s.commentRepo.CountByDate(date)
	likeCount, _ := s.likeRepo.CountAll()
	newLikeCount, _ := s.likeRepo.CountByDate(date)

	return &types.OverviewStats{
		UserCount:       int(userCount),
		NewUserCount:    int(newUserCount),
		PostCount:       int(postCount),
		NewPostCount:    int(newPostCount),
		CommentCount:    int(commentCount),
		NewCommentCount: int(newCommentCount),
		LikeCount:       int(likeCount),
		NewLikeCount:    int(newLikeCount),
	}
}

func (s *StatsService) CalculateStats() error {
	today := time.Now().Format("2006-01-02")
	todayDate, _ := time.Parse("2006-01-02", today)

	return s.statsRepo.Transaction(func(tx *gorm.DB) error {
		var userCount int64
		tx.Model(&model.User{}).Count(&userCount)

		var newUserCount int64
		tx.Model(&model.User{}).
			Where("DATE(created_at) = ?", today).Count(&newUserCount)

		var postCount int64
		tx.Model(&model.Post{}).Count(&postCount)

		var newPostCount int64
		tx.Model(&model.Post{}).
			Where("DATE(created_at) = ?", today).Count(&newPostCount)

		var commentCount int64
		tx.Model(&model.Comment{}).Count(&commentCount)

		var newCommentCount int64
		tx.Model(&model.Comment{}).
			Where("DATE(created_at) = ?", today).Count(&newCommentCount)

		var likeCount int64
		tx.Model(&model.Like{}).Count(&likeCount)

		var newLikeCount int64
		tx.Model(&model.Like{}).
			Where("DATE(created_at) = ?", today).Count(&newLikeCount)

		var favoriteCount int64
		tx.Model(&model.Favorite{}).Count(&favoriteCount)

		var newFavoriteCount int64
		tx.Model(&model.Favorite{}).
			Where("DATE(created_at) = ?", today).Count(&newFavoriteCount)

		var shareCount int64
		tx.Model(&model.ShareRecord{}).Count(&shareCount)

		var newShareCount int64
		tx.Model(&model.ShareRecord{}).
			Where("DATE(created_at) = ?", today).Count(&newShareCount)

		var topicCount int64
		tx.Model(&model.Topic{}).Count(&topicCount)

		var newTopicCount int64
		tx.Model(&model.Topic{}).
			Where("DATE(created_at) = ?", today).Count(&newTopicCount)

		var activeUserCount int64
		tx.Model(&model.UserDevice{}).
			Where("DATE(last_active_at) = ?", today).
			Distinct("user_id").Count(&activeUserCount)

		stats := &model.Statistics{
			StatDate:         todayDate,
			UserCount:        int(userCount),
			NewUserCount:     int(newUserCount),
			PostCount:        int(postCount),
			NewPostCount:     int(newPostCount),
			CommentCount:     int(commentCount),
			NewCommentCount:  int(newCommentCount),
			LikeCount:        int(likeCount),
			NewLikeCount:     int(newLikeCount),
			FavoriteCount:    int(favoriteCount),
			NewFavoriteCount: int(newFavoriteCount),
			ShareCount:       int(shareCount),
			NewShareCount:    int(newShareCount),
			TopicCount:       int(topicCount),
			NewTopicCount:    int(newTopicCount),
			ActiveUserCount:  int(activeUserCount),
			CreatedAt:        time.Now(),
		}

		return s.statsRepo.CreateOrUpdateWithTx(tx, stats)
	})
}

func (s *StatsService) GetUserGrowthStats(startDate, endDate string) (*types.UserGrowthStats, error) {
	return s.statsRepo.GetUserGrowthStats(startDate, endDate)
}

func (s *StatsService) GetContentStats(startDate, endDate string) (*types.ContentStats, error) {
	return s.statsRepo.GetContentStats(startDate, endDate)
}

func (s *StatsService) GetInteractionStats(startDate, endDate string) (*types.InteractionStats, error) {
	return s.statsRepo.GetInteractionStats(startDate, endDate)
}

// GetTrendData 获取趋势数据（分析报表）
func (s *StatsService) GetTrendData(trendType string, days int) (*types.TrendData, error) {
	data := make([]types.ChartDataPoint, 0)
	endTime := time.Now()

	switch trendType {
	case "user_growth":
		for i := days - 1; i >= 0; i-- {
			date := endTime.AddDate(0, 0, -i)
			dateStr := date.Format("2006-01-02")
			count, _ := s.userRepo.CountByDateAndStatus(dateStr, 1)
			data = append(data, types.ChartDataPoint{Label: dateStr, Value: int(count)})
		}
		return &types.TrendData{Title: "用户增长趋势", XKey: "日期", YKey: "新增用户", Data: data}, nil
	case "post_publish":
		for i := days - 1; i >= 0; i-- {
			date := endTime.AddDate(0, 0, -i)
			dateStr := date.Format("2006-01-02")
			count, _ := s.postRepo.CountByDate(dateStr)
			data = append(data, types.ChartDataPoint{Label: dateStr, Value: int(count)})
		}
		return &types.TrendData{Title: "内容发布趋势", XKey: "日期", YKey: "发布数量", Data: data}, nil
	case "activity_log":
		for i := days - 1; i >= 0; i-- {
			date := endTime.AddDate(0, 0, -i)
			dateStr := date.Format("2006-01-02")
			count, _ := s.statsRepo.CountActivityLogsByDate(dateStr)
			data = append(data, types.ChartDataPoint{Label: dateStr, Value: int(count)})
		}
		return &types.TrendData{Title: "用户活动趋势", XKey: "日期", YKey: "活动次数", Data: data}, nil
	default:
		return &types.TrendData{Title: trendType, XKey: "日期", YKey: "数值", Data: data}, nil
	}
}

// GetComparisonData 获取对比数据（分析报表）- 环比/同比
func (s *StatsService) GetComparisonData(trendType string, compareType string) (*types.ComparisonData, error) {
	currentData := make([]types.ChartDataPoint, 0)
	previousData := make([]types.ChartDataPoint, 0)

	days := 7
	if compareType != "week" {
		days = 30
	}
	endTime := time.Now()

	var previousEnd time.Time
	if compareType == "week" {
		previousEnd = endTime.AddDate(0, 0, -days)
	} else {
		previousEnd = endTime.AddDate(0, -1, 0)
	}

	switch trendType {
	case "user_growth":
		for i := days - 1; i >= 0; i-- {
			date := endTime.AddDate(0, 0, -i)
			dateStr := date.Format("2006-01-02")
			count, _ := s.userRepo.CountByDateAndStatus(dateStr, 1)
			currentData = append(currentData, types.ChartDataPoint{Label: dateStr, Value: int(count)})

			prevDate := previousEnd.AddDate(0, 0, -i)
			prevDateStr := prevDate.Format("2006-01-02")
			prevCount, _ := s.userRepo.CountByDateAndStatus(prevDateStr, 1)
			previousData = append(previousData, types.ChartDataPoint{Label: prevDateStr, Value: int(prevCount)})
		}
	case "post_publish":
		for i := days - 1; i >= 0; i-- {
			date := endTime.AddDate(0, 0, -i)
			dateStr := date.Format("2006-01-02")
			count, _ := s.postRepo.CountByDate(dateStr)
			currentData = append(currentData, types.ChartDataPoint{Label: dateStr, Value: int(count)})

			prevDate := previousEnd.AddDate(0, 0, -i)
			prevDateStr := prevDate.Format("2006-01-02")
			prevCount, _ := s.postRepo.CountByDate(prevDateStr)
			previousData = append(previousData, types.ChartDataPoint{Label: prevDateStr, Value: int(prevCount)})
		}
	}

	comparisonLabel := "环比"
	if compareType == "month" {
		comparisonLabel = "同比"
	}

	var title string
	switch trendType {
	case "user_growth":
		title = "用户增长" + comparisonLabel + "对比"
	case "post_publish":
		title = "内容发布" + comparisonLabel + "对比"
	default:
		title = trendType + comparisonLabel + "对比"
	}

	return &types.ComparisonData{Title: title, Current: currentData, Previous: previousData, Comparison: comparisonLabel}, nil
}

// GetDistributionData 获取分布数据（分析报表）
func (s *StatsService) GetDistributionData(distType string) (*types.DistributionData, error) {
	data := make([]types.ChartDataPoint, 0)

	switch distType {
	case "pet_type":
		pets, _ := s.statsRepo.GetPetTypeDistribution()
		petTypeLabels := map[int]string{1: "狗狗", 2: "猫咪", 3: "其他"}
		for _, p := range pets {
			label, ok := petTypeLabels[p.PetType]
			if !ok {
				label = "未知"
			}
			data = append(data, types.ChartDataPoint{Label: label, Value: p.Count})
		}
		return &types.DistributionData{Title: "宠物类型分布", Type: "pie", Data: data}, nil
	case "user_location":
		locations, _ := s.userRepo.GetLocationDistribution()
		for _, l := range locations {
			data = append(data, types.ChartDataPoint{Label: l.Location, Value: l.Count})
		}
		return &types.DistributionData{Title: "用户地域分布", Type: "bar", Data: data}, nil
	case "post_status":
		statuses, _ := s.postRepo.GetStatusDistribution()
		statusLabels := map[int]string{0: "草稿", 1: "待审核", 2: "已发布", 3: "已拒绝"}
		for _, st := range statuses {
			label, ok := statusLabels[st.Status]
			if !ok {
				label = "未知"
			}
			data = append(data, types.ChartDataPoint{Label: label, Value: st.Count})
		}
		return &types.DistributionData{Title: "内容状态分布", Type: "pie", Data: data}, nil
	case "user_gender":
		genders, _ := s.userRepo.GetGenderDistribution()
		genderLabels := map[int]string{0: "未知", 1: "男", 2: "女"}
		for _, g := range genders {
			label, ok := genderLabels[g.Gender]
			if !ok {
				label = "未知"
			}
			data = append(data, types.ChartDataPoint{Label: label, Value: g.Count})
		}
		return &types.DistributionData{Title: "用户性别分布", Type: "pie", Data: data}, nil
	default:
		return &types.DistributionData{Title: distType, Type: "bar", Data: data}, nil
	}
}

func (s *StatsService) ExportStatsToExcel(exportType, startDate, endDate string) ([]byte, error) {
	f := excelize.NewFile()

	switch exportType {
	case "overview":
		s.exportOverviewToExcel(f)
	case "user_growth":
		stats, err := s.GetUserGrowthStats(startDate, endDate)
		if err != nil {
			return nil, err
		}
		s.exportUserGrowthToExcel(f, stats)
	case "content":
		stats, err := s.GetContentStats(startDate, endDate)
		if err != nil {
			return nil, err
		}
		s.exportContentToExcel(f, stats)
	case "interaction":
		stats, err := s.GetInteractionStats(startDate, endDate)
		if err != nil {
			return nil, err
		}
		s.exportInteractionToExcel(f, stats)
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *StatsService) exportOverviewToExcel(f *excelize.File) {
	sheet := "概览统计"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"指标", "数值"}
	for i, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), header)
	}

	stats, _ := s.GetOverviewStats()
	data := [][]interface{}{
		{"总用户数", stats.UserCount},
		{"新增用户数", stats.NewUserCount},
		{"总帖子数", stats.PostCount},
		{"新增帖子数", stats.NewPostCount},
		{"总评论数", stats.CommentCount},
		{"新增评论数", stats.NewCommentCount},
		{"总点赞数", stats.LikeCount},
		{"新增点赞数", stats.NewLikeCount},
	}

	for row, values := range data {
		for col, value := range values {
			f.SetCellValue(sheet, fmt.Sprintf("%c%d", 'A'+col, row+2), value)
		}
	}
}

func (s *StatsService) exportUserGrowthToExcel(f *excelize.File, stats *types.UserGrowthStats) {
	sheet := "用户增长"
	f.NewSheet(sheet)

	headers := []string{"日期", "新增用户", "总用户", "活跃用户"}
	for i, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), header)
	}

	for row, day := range stats.Data {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row+2), day.Date)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row+2), day.NewUsers)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row+2), day.TotalUsers)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row+2), day.ActiveUsers)
	}

	totalRow := len(stats.Data) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "合计")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", totalRow), stats.TotalNew)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), stats.TotalActive)
}

func (s *StatsService) exportContentToExcel(f *excelize.File, stats *types.ContentStats) {
	sheet := "内容统计"
	f.NewSheet(sheet)

	headers := []string{"日期", "新增帖子", "新增评论", "新增话题"}
	for i, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), header)
	}

	for row, day := range stats.Data {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row+2), day.Date)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row+2), day.Posts)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row+2), day.Comments)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row+2), day.Topics)
	}

	totalRow := len(stats.Data) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "合计")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", totalRow), stats.TotalPosts)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", totalRow), stats.TotalComments)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), stats.TotalTopics)
}

func (s *StatsService) exportInteractionToExcel(f *excelize.File, stats *types.InteractionStats) {
	sheet := "互动统计"
	f.NewSheet(sheet)

	headers := []string{"日期", "点赞数", "收藏数", "分享数", "评论数"}
	for i, header := range headers {
		f.SetCellValue(sheet, fmt.Sprintf("%c1", 'A'+i), header)
	}

	for row, day := range stats.Data {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row+2), day.Date)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row+2), day.Likes)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row+2), day.Favorites)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row+2), day.Shares)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row+2), day.Comments)
	}

	totalRow := len(stats.Data) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", totalRow), "合计")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", totalRow), stats.TotalLikes)
	f.SetCellValue(sheet, fmt.Sprintf("C%d", totalRow), stats.TotalFavorites)
	f.SetCellValue(sheet, fmt.Sprintf("D%d", totalRow), stats.TotalShares)
	f.SetCellValue(sheet, fmt.Sprintf("E%d", totalRow), stats.TotalComments)
}
