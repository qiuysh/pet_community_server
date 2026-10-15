package types

type ChartDataPoint struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type TrendData struct {
	Title string           `json:"title"`
	XKey  string           `json:"xKey"`
	YKey  string           `json:"yKey"`
	Data  []ChartDataPoint `json:"data"`
}

type ComparisonData struct {
	Title      string           `json:"title"`
	Current    []ChartDataPoint `json:"current"`
	Previous   []ChartDataPoint `json:"previous"`
	Comparison string           `json:"comparison"`
}

type DistributionData struct {
	Title string           `json:"title"`
	Type  string           `json:"type"`
	Data  []ChartDataPoint `json:"data"`
}

type DashboardSummary struct {
	TotalUsers               int     `json:"totalUsers"`
	TodayNewUsers            int     `json:"todayNewUsers"`
	YesterdayNewUsers        int     `json:"yesterdayNewUsers"`
	TodayActiveUsers         int     `json:"todayActiveUsers"`
	YesterdayActiveUsers     int     `json:"yesterdayActiveUsers"`
	TotalPosts               int     `json:"totalPosts"`
	TodayPosts               int     `json:"todayPosts"`
	YesterdayPosts           int     `json:"yesterdayPosts"`
	TodayLikes               int     `json:"todayLikes"`
	YesterdayLikes           int     `json:"yesterdayLikes"`
	TodayComments            int     `json:"todayComments"`
	YesterdayComments        int     `json:"yesterdayComments"`
	TodayFavorites           int     `json:"todayFavorites"`
	YesterdayFavorites       int     `json:"yesterdayFavorites"`
	TodayShares              int     `json:"todayShares"`
	YesterdayShares          int     `json:"yesterdayShares"`
	TodayRegistrations       int     `json:"todayRegistrations"`
	YesterdayRegistrations   int     `json:"yesterdayRegistrations"`
	TodayVotes               int     `json:"todayVotes"`
	YesterdayVotes           int     `json:"yesterdayVotes"`
	PendingReviews           int     `json:"pendingReviews"`
	PendingRegistrations     int     `json:"pendingRegistrations"`
	PendingAdoptions         int     `json:"pendingAdoptions"`
	TotalPets                int     `json:"totalPets"`
	TotalLikes               int     `json:"totalLikes"`
	OngoingActivities        int     `json:"ongoingActivities"`
	OngoingLostPets          int     `json:"ongoingLostPets"`
	AdTotalViews             int     `json:"adTotalViews"`
	AdTotalClicks            int     `json:"adTotalClicks"`
	ReportHandleRate         float64 `json:"reportHandleRate"`
}

// DashboardOverview 驾驶舱概览响应（实时指标 + 近 7 日双趋势）
type DashboardOverview struct {
	Summary    DashboardSummary `json:"summary"`
	TodayTrend TrendData        `json:"todayTrend"`
	PostTrend  TrendData        `json:"postTrend"`
}

// TodoItem 待办聚合项
type TodoItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
	Link  string `json:"link"`
}

// TodoAggregate 待办聚合数据
type TodoAggregate struct {
	Items []TodoItem `json:"items"`
	Total int        `json:"total"`
}

// AdminResponse 是管理员用户的响应结构，不包含敏感信息
type AdminResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ReviewContent 待审核内容
type ReviewContent struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	Images    string `json:"images,omitempty"`
	Video     string `json:"video,omitempty"`
	PostID    uint   `json:"postId,omitempty"`
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// ReviewListResponse 待审核列表响应
type ReviewListResponse struct {
	List  []ReviewContent `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

// ReviewRequest 审核请求
type ReviewRequest struct {
	ID     uint   `json:"id" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

// BatchReviewRequest 批量审核请求
type BatchReviewRequest struct {
	IDs    []uint `json:"ids" binding:"required"`
	Status int    `json:"status" binding:"required,oneof=1 2"`
	Reason string `json:"reason,omitempty"`
}

// ReviewHistory 审核历史记录
type ReviewHistory struct {
	ID         uint   `json:"id"`
	TargetType string `json:"targetType"`
	TargetID   uint   `json:"targetId"`
	Action     string `json:"action"`
	Reason     string `json:"reason,omitempty"`
	AdminID    uint   `json:"adminId"`
	AdminName  string `json:"adminName"`
	CreatedAt  string `json:"createdAt"`
}

// ReviewHistoryResponse 审核历史响应
type ReviewHistoryResponse struct {
	List  []ReviewHistory `json:"list"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}

type AdPositionResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Size      string `json:"size"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type AdPositionRequest struct {
	Name     string `json:"name" binding:"required,max=50"`
	Location string `json:"location" binding:"required,oneof=home_banner splash detail_sidebar topic_banner"`
	Size     string `json:"size" binding:"required,max=20"`
	Status   int    `json:"status" binding:"omitempty,oneof=0 1"`
}

type AdvertisementResponse struct {
	ID         uint               `json:"id"`
	PositionID uint               `json:"positionId"`
	Position   AdPositionResponse `json:"position"`
	Title      string             `json:"title"`
	ImageURL   string             `json:"imageUrl"`
	LinkURL    string             `json:"linkUrl"`
	StartTime  string             `json:"startTime"`
	EndTime    string             `json:"endTime"`
	ViewCount  int                `json:"viewCount"`
	ClickCount int                `json:"clickCount"`
	Status     int                `json:"status"`
	CreatedBy  uint               `json:"createdBy"`
	CreatedAt  string             `json:"createdAt"`
	UpdatedAt  string             `json:"updatedAt"`
}

type AdvertisementRequest struct {
	PositionID uint   `json:"positionId" binding:"required"`
	Title      string `json:"title" binding:"required,max=50"`
	ImageURL   string `json:"imageUrl" binding:"required,max=255"`
	LinkURL    string `json:"linkUrl" binding:"omitempty,max=255"`
	StartTime  string `json:"startTime" binding:"required"`
	EndTime    string `json:"endTime" binding:"required"`
}

type AdvertisementListResponse struct {
	List  []AdvertisementResponse `json:"list"`
	Total int64                   `json:"total"`
	Page  int                     `json:"page"`
	Size  int                     `json:"size"`
}

type AdPositionListResponse struct {
	List  []AdPositionResponse `json:"list"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

type AdReviewRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int  `json:"status" binding:"required,oneof=0 1 2"`
}

type AdStatsResponse struct {
	TotalViews      int              `json:"totalViews"`
	TotalClicks     int              `json:"totalClicks"`
	CTR             float64          `json:"ctr"`
	StatsByDay      []ChartDataPoint `json:"statsByDay"`
	StatsByPosition []ChartDataPoint `json:"statsByPosition"`
}
