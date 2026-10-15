package types

// 活动信息结构
type ActivityInfo struct {
	ID                uint   `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Type              string `json:"type"`
	Cover             string `json:"cover"`
	Status            int    `json:"status"`
	StartTime         string `json:"startTime"`
	EndTime           string `json:"endTime"`
	Rules             string `json:"rules"`
	Prizes            string `json:"prizes"`
	MaxParticipants   int    `json:"maxParticipants"`
	MaxSubmissions    int    `json:"maxSubmissions"`
	VotesPerDay       int    `json:"votesPerDay"`
	RequireApproval   bool   `json:"requireApproval"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
	ParticipantsCount int    `json:"participantsCount"`
	SubmissionsCount  int    `json:"submissionsCount"`
	IsRegistered      bool   `json:"isRegistered"`
	HasSubmitted      bool   `json:"hasSubmitted"`
}

// 创建活动请求
type ActivityCreateRequest struct {
	Title           string `json:"title" binding:"required,max=100"`
	Description     string `json:"description" binding:"required,max=500"`
	Type            string `json:"type" binding:"required,oneof=voting contest collection"`
	Cover           string `json:"cover" binding:"omitempty,max=255"`
	StartTime       string `json:"startTime" binding:"required"`
	EndTime         string `json:"endTime" binding:"required"`
	Rules           string `json:"rules" binding:"omitempty"`
	Prizes          string `json:"prizes" binding:"omitempty"`
	MaxParticipants int    `json:"maxParticipants" binding:"omitempty,min=0"`
	MaxSubmissions  int    `json:"maxSubmissions" binding:"omitempty,min=1,max=10"`
	VotesPerDay     int    `json:"votesPerDay" binding:"omitempty,min=1,max=50"`
	RequireApproval bool   `json:"requireApproval"`
}

// 更新活动请求
type ActivityUpdateRequest struct {
	Title           string `json:"title" binding:"omitempty,max=100"`
	Description     string `json:"description" binding:"omitempty,max=500"`
	Cover           string `json:"cover" binding:"omitempty,max=255"`
	StartTime       string `json:"startTime" binding:"omitempty"`
	EndTime         string `json:"endTime" binding:"omitempty"`
	Rules           string `json:"rules" binding:"omitempty"`
	Prizes          string `json:"prizes" binding:"omitempty"`
	MaxParticipants int    `json:"maxParticipants" binding:"omitempty,min=0"`
	MaxSubmissions  int    `json:"maxSubmissions" binding:"omitempty,min=1,max=10"`
	VotesPerDay     int    `json:"votesPerDay" binding:"omitempty,min=1,max=50"`
	RequireApproval bool   `json:"requireApproval"`
}

// 活动列表请求
type ActivityListRequest struct {
	Type     string `form:"type"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// 活动列表响应
type ActivityListResponse struct {
	List     []ActivityInfo `json:"list"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"pageSize"`
}

// 报名信息结构
type RegistrationInfo struct {
	ID           uint   `json:"id"`
	ActivityID   uint   `json:"activityId"`
	UserID       uint   `json:"userId"`
	UserNickname string `json:"userNickname"`
	UserAvatar   string `json:"userAvatar"`
	Status       int    `json:"status"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Message      string `json:"message"`
	CreatedAt    string `json:"createdAt"`
}

// 报名请求
type RegistrationRequest struct {
	ActivityID uint   `json:"activityId" binding:"required"`
	Name       string `json:"name" binding:"required,max=50"`
	Phone      string `json:"phone" binding:"required,max=20"`
	Message    string `json:"message" binding:"omitempty,max=200"`
}

// 报名列表请求
type RegistrationListRequest struct {
	ActivityID uint `form:"activityId" binding:"required"`
	Status     int  `form:"status"`
	Page       int  `form:"page"`
	PageSize   int  `form:"pageSize"`
}

// 报名列表响应
type RegistrationListResponse struct {
	List     []RegistrationInfo `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// 作品信息结构
type SubmissionInfo struct {
	ID           uint     `json:"id"`
	ActivityID   uint     `json:"activityId"`
	UserID       uint     `json:"userId"`
	UserNickname string   `json:"userNickname"`
	UserAvatar   string   `json:"userAvatar"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Content      string   `json:"content"`
	Images       []string `json:"images,omitempty"`
	Video        string   `json:"video,omitempty"`
	VotesCount   int      `json:"votesCount"`
	Rank         int      `json:"rank"`
	IsAwarded    bool     `json:"isAwarded"`
	AwardLevel   string   `json:"awardLevel"`
	Status       int      `json:"status"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	IsVoted      bool     `json:"isVoted"`
}

// 作品创建请求
type SubmissionCreateRequest struct {
	ActivityID  uint     `json:"activityId" binding:"required"`
	Title       string   `json:"title" binding:"required,max=50"`
	Description string   `json:"description" binding:"omitempty,max=500"`
	Type        string   `json:"type" binding:"required,oneof=image video text"`
	Content     string   `json:"content" binding:"omitempty"`
	Images      []string `json:"images" binding:"omitempty,max=9"`
	Video       string   `json:"video" binding:"omitempty,max=255"`
}

// 作品更新请求
type SubmissionUpdateRequest struct {
	Title       string   `json:"title" binding:"omitempty,max=50"`
	Description string   `json:"description" binding:"omitempty,max=500"`
	Content     string   `json:"content" binding:"omitempty"`
	Images      []string `json:"images" binding:"omitempty,max=9"`
	Video       string   `json:"video" binding:"omitempty,max=255"`
}

// 作品列表请求
type SubmissionListRequest struct {
	ActivityID uint   `form:"activityId" binding:"required"`
	Sort       string `form:"sort"`
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
}

// 作品列表响应
type SubmissionListResponse struct {
	List     []SubmissionInfo `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// 投票信息结构
type VoteInfo struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"userId"`
	ActivityID   uint   `json:"activityId"`
	SubmissionID uint   `json:"submissionId"`
	CreatedAt    string `json:"createdAt"`
}

// 投票请求
type VoteRequest struct {
	ActivityID   uint `json:"activityId" binding:"required"`
	SubmissionID uint `json:"submissionId" binding:"required"`
}

// 投票结果
type VoteResult struct {
	Voted          bool `json:"voted"`
	VotesCount     int  `json:"votesCount"`
	RemainingVotes int  `json:"remainingVotes"`
}

// 投票历史响应
type VoteHistoryResponse struct {
	List     []VoteInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

// 评选请求
type AwardRequest struct {
	SubmissionID uint   `json:"submissionId" binding:"required"`
	AwardLevel   string `json:"awardLevel" binding:"required,oneof=first second third"`
}

// 批量评选请求
type BatchAwardRequest struct {
	Awards []AwardRequest `json:"awards" binding:"required,min=1"`
}

// 评选页面查询响应
type AwardPageResponse struct {
	Activity     *ActivityInfo    `json:"activity"`
	Submissions  []SubmissionInfo `json:"submissions"`
	Total        int64            `json:"total"`
	AwardedCount int64            `json:"awardedCount"`
	FirstPrize   *SubmissionInfo  `json:"firstPrize,omitempty"`
	SecondPrizes []SubmissionInfo `json:"secondPrizes,omitempty"`
	ThirdPrizes  []SubmissionInfo `json:"thirdPrizes,omitempty"`
}

// 发布结果请求
type PublishAwardRequest struct {
	ActivityID uint `json:"activityId" binding:"required"`
}

// 发布结果响应
type PublishAwardResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	NotifiedUsers []uint `json:"notifiedUsers"`
}

// 用户投票统计响应
type UserVoteStatsResponse struct {
	TotalVotesToday int `json:"totalVotesToday"`
	RemainingVotes  int `json:"remainingVotes"`
	VotesPerDay     int `json:"votesPerDay"`
}

// ========== 签到相关类型 ==========

// SignInInfo 签到信息结构
type SignInInfo struct {
	ID             uint   `json:"id"`
	ActivityID     uint   `json:"activityId"`
	UserID         uint   `json:"userId"`
	UserNickname   string `json:"userNickname"`
	UserAvatar     string `json:"userAvatar"`
	SignInTime     string `json:"signInTime"`
	SignInLocation string `json:"signInLocation"`
	CreatedAt      string `json:"createdAt"`
}

// SignInRequest 签到请求
type SignInRequest struct {
	ActivityID     uint   `json:"activityId" binding:"required"`
	QRCode         string `json:"qrCode" binding:"required"`
	SignInLocation string `json:"signInLocation" binding:"omitempty,max=100"`
}

// QRCodeResponse 二维码响应
type QRCodeResponse struct {
	QRCode   string `json:"qrCode"`   // 二维码内容
	ImageURL string `json:"imageUrl"` // 二维码图片URL（base64格式）
	Expires  string `json:"expires"`  // 过期时间
}

// SignInListResponse 签到列表响应
type SignInListResponse struct {
	List     []SignInInfo `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// ========== 相册相关类型 ==========

// AlbumPhotoInfo 相册照片信息结构
type AlbumPhotoInfo struct {
	ID           uint   `json:"id"`
	ActivityID   uint   `json:"activityId"`
	UserID       uint   `json:"userId"`
	UserNickname string `json:"userNickname"`
	UserAvatar   string `json:"userAvatar"`
	PhotoURL     string `json:"photoUrl"`
	Description  string `json:"description"`
	Status       int    `json:"status"`
	CreatedAt    string `json:"createdAt"`
}

// AlbumPhotoCreateRequest 相册照片创建请求
type AlbumPhotoCreateRequest struct {
	ActivityID  uint   `json:"activityId" binding:"required"`
	PhotoURL    string `json:"photoUrl" binding:"required,max=255"`
	Description string `json:"description" binding:"omitempty,max=200"`
}

// AlbumPhotoListResponse 相册照片列表响应
type AlbumPhotoListResponse struct {
	List     []AlbumPhotoInfo `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

// ========== 活动回顾相关类型 ==========

// ActivityReviewInfo 活动回顾信息结构
type ActivityReviewInfo struct {
	ID          uint     `json:"id"`
	ActivityID  uint     `json:"activityId"`
	Title       string   `json:"title"`
	Content     string   `json:"content"`
	Photos      []string `json:"photos,omitempty"`
	PublishedAt string   `json:"publishedAt"`
	PublishedBy uint     `json:"publishedBy"`
	CreatedAt   string   `json:"createdAt"`
}

// ActivityReviewCreateRequest 活动回顾创建请求
type ActivityReviewCreateRequest struct {
	ActivityID uint     `json:"activityId" binding:"required"`
	Title      string   `json:"title" binding:"required,max=100"`
	Content    string   `json:"content" binding:"required"`
	Photos     []string `json:"photos" binding:"omitempty,max=20"`
}
