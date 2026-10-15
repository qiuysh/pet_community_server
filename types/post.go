package types

type PostInfo struct {
	ID             uint        `json:"id"`
	UserID         uint        `json:"userId"`
	UserNickname   string      `json:"userNickname"`
	UserAvatar     string      `json:"userAvatar"`
	Content        string      `json:"content"`
	Images         []string    `json:"images,omitempty"`
	Video          string      `json:"video,omitempty"`
	PetIDs         []uint      `json:"petIds,omitempty"`
	TopicIDs       []uint      `json:"topicIds,omitempty"`
	Topics         []TopicInfo `json:"topics,omitempty"`
	Pets           []PetInfo   `json:"pets,omitempty"`
	Location       string      `json:"location,omitempty"`
	Visibility     int         `json:"visibility"`
	LikesCount     int         `json:"likesCount"`
	CommentsCount  int         `json:"commentsCount"`
	SharesCount    int         `json:"sharesCount"`
	FavoritesCount int         `json:"favoritesCount"`
	IsPinned       int        `json:"isPinned"`
	IsFeatured     int        `json:"isFeatured"`
	IsLiked        bool        `json:"isLiked"`
	IsFavorited    bool        `json:"isFavorited"`
	Status         int        `json:"status"`
	PublishAt      string      `json:"publishAt,omitempty"`
	CreatedAt      string      `json:"createdAt"`
}

type PostCreateRequest struct {
	Content    string   `json:"content" binding:"omitempty,max=500"`
	Images     []string `json:"images" binding:"omitempty,max=9"`
	Video      string   `json:"video" binding:"omitempty,max=255"`
	PetIDs     []uint   `json:"petIds" binding:"omitempty"`
	TopicIDs   []uint   `json:"topicIds" binding:"omitempty,max=3"`
	Location   string   `json:"location" binding:"omitempty,max=100"`
	Visibility *int     `json:"visibility" binding:"omitempty,oneof=0 1 2"` // nil 时用 U8 账号默认
	AtUsers    []uint   `json:"atUsers" binding:"omitempty"`
	PublishAt  string   `json:"publishAt" binding:"omitempty"`
}

type PostUpdateRequest struct {
	Content    string   `json:"content" binding:"omitempty,max=500"`
	Images     []string `json:"images" binding:"omitempty,max=9"`
	PetIDs     []uint   `json:"petIds" binding:"omitempty"`
	TopicIDs   []uint   `json:"topicIds" binding:"omitempty,max=3"`
	Location   string   `json:"location" binding:"omitempty,max=100"`
	Visibility int      `json:"visibility" binding:"omitempty,oneof=0 1 2"`
	AtUsers    []uint   `json:"atUsers" binding:"omitempty"`
	PublishAt  string   `json:"publishAt" binding:"omitempty"`
}

type PostListRequest struct {
	UserID   uint   `form:"userId"`
	PetID    uint   `form:"petId"`
	TopicID  uint   `form:"topicId"`
	Type     string `form:"type"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type PostListResponse struct {
	List     []PostInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

type CommentInfo struct {
	ID           uint          `json:"id"`
	PostID       uint          `json:"postId"`
	UserID       uint          `json:"userId"`
	UserNickname string        `json:"userNickname"`
	UserAvatar   string        `json:"userAvatar"`
	ParentID     uint          `json:"parentId"`
	ReplyToID    uint          `json:"replyToId,omitempty"`
	ReplyToName  string        `json:"replyToName,omitempty"`
	Content      string        `json:"content"`
	AtUsers      []uint        `json:"atUsers,omitempty"`
	LikesCount   int           `json:"likesCount"`
	IsLiked      bool          `json:"isLiked"`
	Children     []CommentInfo `json:"children,omitempty"`
	Status       int           `json:"status"`
	CreatedAt    string        `json:"createdAt"`
}

type CommentCreateRequest struct {
	PostID    uint   `json:"postId" binding:"required"`
	ParentID  uint   `json:"parentId" binding:"omitempty"`
	ReplyToID uint   `json:"replyToId" binding:"omitempty"`
	Content   string `json:"content" binding:"required,max=200"`
	AtUsers   []uint `json:"atUsers" binding:"omitempty"`
}

type CommentUpdateRequest struct {
	Content string `json:"content" binding:"required,max=200"`
}

type CommentListRequest struct {
	PostID   uint   `form:"postId" binding:"required"`
	Sort     string `form:"sort"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type CommentListResponse struct {
	List     []CommentInfo `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type TopicInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       string `json:"cover"`
	PostsCount  int    `json:"postsCount"`
	Status      int    `json:"status"`
}

type TopicCreateRequest struct {
	Name        string `json:"name" binding:"required,max=50"`
	Description string `json:"description" binding:"omitempty,max=200"`
	Cover       string `json:"cover" binding:"omitempty,max=255"`
}

type TopicListRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type TopicListResponse struct {
	List     []TopicInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

type DraftInfo struct {
	ID         uint     `json:"id"`
	UserID     uint     `json:"userId"`
	Content    string   `json:"content"`
	Images     []string `json:"images,omitempty"`
	Video      string   `json:"video,omitempty"`
	PetIDs     []uint   `json:"petIds,omitempty"`
	TopicIDs   []uint   `json:"topicIds,omitempty"`
	Location   string   `json:"location,omitempty"`
	Visibility int      `json:"visibility"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
}

type DraftCreateRequest struct {
	Content    string   `json:"content" binding:"omitempty,max=500"`
	Images     []string `json:"images" binding:"omitempty,max=9"`
	Video      string   `json:"video" binding:"omitempty,max=255"`
	PetIDs     []uint   `json:"petIds" binding:"omitempty"`
	TopicIDs   []uint   `json:"topicIds" binding:"omitempty,max=3"`
	Location   string   `json:"location" binding:"omitempty,max=100"`
	Visibility int      `json:"visibility" binding:"omitempty,oneof=0 1 2"`
}

type DraftListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type DraftListResponse struct {
	List     []DraftInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

type ReportInfo struct {
	ID           uint   `json:"id"`
	ReporterID   uint   `json:"reporterId"`
	ReporterName string `json:"reporterName,omitempty"`
	TargetType   string `json:"targetType"`
	TargetID     uint   `json:"targetId"`
	Reason       string `json:"reason"`
	ReasonType   string `json:"reasonType"`
	Status       int    `json:"status"`
	HandleResult string `json:"handleResult,omitempty"`
	HandledBy    uint   `json:"handledBy,omitempty"`
	HandledAt    string `json:"handledAt,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

// HandleReportRequest 处理举报请求（approve=通过/下架, reject=忽略, warn=警告）
type HandleReportRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject warn"`
	Result string `json:"result" binding:"max=200"`
}

// BatchHandleReportRequest 批量处理举报
type BatchHandleReportRequest struct {
	IDs    []uint `json:"ids" binding:"required,min=1"`
	Action string `json:"action" binding:"required,oneof=approve reject warn"`
	Result string `json:"result" binding:"max=200"`
}

// ReportTypeCount 举报类型统计项
type ReportTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

// ReportStatsResponse 举报统计
type ReportStatsResponse struct {
	Total        int64             `json:"total"`
	Pending      int64             `json:"pending"`
	Handled      int64             `json:"handled"`
	Ignored      int64             `json:"ignored"`
	HandleRate   float64           `json:"handleRate"`
	ByReasonType []ReportTypeCount `json:"byReasonType"`
	ByTargetType []ReportTypeCount `json:"byTargetType"`
}

type ReportCreateRequest struct {
	TargetType string `json:"targetType" binding:"required,oneof=post comment"`
	TargetID   uint   `json:"targetId" binding:"required"`
	Reason     string `json:"reason" binding:"required,max=500"`
	ReasonType string `json:"reasonType" binding:"required,oneof=spam adult attack other"`
}

type ReportListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type ReportListResponse struct {
	List     []ReportInfo `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

type LikeResult struct {
	Liked      bool `json:"liked"`
	LikesCount int  `json:"likesCount"`
}

type FavoriteResult struct {
	Favorited      bool `json:"favorited"`
	FavoritesCount int  `json:"favoritesCount"`
}

type ShareRequest struct {
	Platform string `json:"platform" binding:"omitempty,max=20"`
}

type LikeInfo struct {
	ID        uint   `json:"id"`
	UserID    uint   `json:"userId"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"createdAt"`
}

type LikeListResponse struct {
	List     []LikeInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

type UserSearchInfo struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
