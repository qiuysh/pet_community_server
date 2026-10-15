package types

type OverviewStats struct {
	UserCount       int `json:"userCount"`
	NewUserCount    int `json:"newUserCount"`
	PostCount       int `json:"postCount"`
	NewPostCount    int `json:"newPostCount"`
	CommentCount    int `json:"commentCount"`
	NewCommentCount int `json:"newCommentCount"`
	LikeCount       int `json:"likeCount"`
	NewLikeCount    int `json:"newLikeCount"`
}

type UserGrowthDay struct {
	Date        string `json:"date"`
	NewUsers    int    `json:"newUsers"`
	TotalUsers  int    `json:"totalUsers"`
	ActiveUsers int    `json:"activeUsers"`
}

type UserGrowthStats struct {
	Data        []UserGrowthDay `json:"data"`
	TotalNew    int             `json:"totalNew"`
	TotalActive int             `json:"totalActive"`
}

type ContentStatsDay struct {
	Date     string `json:"date"`
	Posts    int    `json:"posts"`
	Comments int    `json:"comments"`
	Topics   int    `json:"topics"`
}

type ContentStats struct {
	Data          []ContentStatsDay `json:"data"`
	TotalPosts    int               `json:"totalPosts"`
	TotalComments int               `json:"totalComments"`
	TotalTopics   int               `json:"totalTopics"`
}

type InteractionStatsDay struct {
	Date      string `json:"date"`
	Likes     int    `json:"likes"`
	Favorites int    `json:"favorites"`
	Shares    int    `json:"shares"`
	Comments  int    `json:"comments"`
}

type InteractionStats struct {
	Data           []InteractionStatsDay `json:"data"`
	TotalLikes     int                   `json:"totalLikes"`
	TotalFavorites int                   `json:"totalFavorites"`
	TotalShares    int                   `json:"totalShares"`
	TotalComments  int                   `json:"totalComments"`
}

type ExportRequest struct {
	Type      string `json:"type" binding:"required,oneof=overview user_growth content interaction"`
	StartDate string `json:"startDate" binding:"required"`
	EndDate   string `json:"endDate" binding:"required"`
}
