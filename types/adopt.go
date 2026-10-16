package types

// 领养信息结构
type AdoptInfo struct {
	ID               uint     `json:"id"`
	PetID            uint     `json:"petId"`
	UserID           uint     `json:"userId"`
	UserNickname     string   `json:"userNickname"`
	UserAvatar       string   `json:"userAvatar"`
	PetName          string   `json:"petName"`
	PetType          string   `json:"petType"`
	PetAge           string   `json:"petAge"`
	PetGender        int      `json:"petGender"`
	HealthStatus     string   `json:"healthStatus"`
	Photos           []string `json:"photos"`
	Description      string   `json:"description"`
	AdoptConditions  string   `json:"adoptConditions"`
	Location         string   `json:"location"`
	ContactPhone     string   `json:"contactPhone"`
	Status           int      `json:"status"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	ApplicationCount int      `json:"applicationCount"`
}

// 创建领养信息请求
type AdoptCreateRequest struct {
	PetID           uint     `json:"petId" binding:"omitempty"`
	PetName         string   `json:"petName" binding:"required,max=50"`
	PetType         string   `json:"petType" binding:"required,oneof=cat dog bird fish other"`
	PetAge          string   `json:"petAge" binding:"omitempty,max=20"`
	PetGender       int      `json:"petGender" binding:"omitempty,oneof=0 1 2"`
	HealthStatus    string   `json:"healthStatus" binding:"omitempty,max=100"`
	Photos          []string `json:"photos" binding:"required,max=9"`
	Description     string   `json:"description" binding:"required,max=500"`
	AdoptConditions string   `json:"adoptConditions" binding:"omitempty,max=500"`
	Location        string   `json:"location" binding:"required,max=100"`
	ContactPhone    string   `json:"contactPhone" binding:"required,max=20"`
}

// 更新领养信息请求
type AdoptUpdateRequest struct {
	PetName         string   `json:"petName" binding:"omitempty,max=50"`
	PetType         string   `json:"petType" binding:"omitempty,oneof=cat dog bird fish other"`
	PetAge          string   `json:"petAge" binding:"omitempty,max=20"`
	PetGender       int      `json:"petGender" binding:"omitempty,oneof=0 1 2"`
	HealthStatus    string   `json:"healthStatus" binding:"omitempty,max=100"`
	Photos          []string `json:"photos" binding:"omitempty,max=9"`
	Description     string   `json:"description" binding:"omitempty,max=500"`
	AdoptConditions string   `json:"adoptConditions" binding:"omitempty,max=500"`
	Location        string   `json:"location" binding:"omitempty,max=100"`
	ContactPhone    string   `json:"contactPhone" binding:"omitempty,max=20"`
	Status          int      `json:"status" binding:"omitempty,oneof=0 1 2"`
}

// 领养列表请求
type AdoptListRequest struct {
	PetType  string `form:"petType"`
	Status   int    `form:"status"`
	Location string `form:"location"`
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// 领养列表响应
type AdoptListResponse struct {
	List     []AdoptInfo `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// 领养申请信息结构
type AdoptionApplicationInfo struct {
	ID              uint   `json:"id"`
	AdoptionID      uint   `json:"adoptionId"`
	UserID          uint   `json:"userId"`
	UserNickname    string `json:"userNickname"`
	UserAvatar      string `json:"userAvatar"`
	ApplicationInfo string `json:"applicationInfo"`
	Status          int    `json:"status"`
	ReviewedBy      uint   `json:"reviewedBy"`
	ReviewedAt      string `json:"reviewedAt"`
	AgreementSigned bool   `json:"agreementSigned"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// 领养申请请求
type AdoptionApplicationRequest struct {
	AdoptionID      uint   `json:"adoptionId" binding:"required"`
	ApplicationInfo string `json:"applicationInfo" binding:"required,max=500"`
}

// 领养申请列表请求
type AdoptionApplicationListRequest struct {
	AdoptionID uint `form:"adoptionId" binding:"required"`
	Status     int  `form:"status"`
	Page       int  `form:"page"`
	PageSize   int  `form:"pageSize"`
}

// 领养申请列表响应
type AdoptionApplicationListResponse struct {
	List     []AdoptionApplicationInfo `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"pageSize"`
}

// 领养审核请求
type AdoptionReviewRequest struct {
	ApplicationID uint   `json:"applicationId" binding:"required"`
	Approve       bool   `json:"approve"` // 勿加 required：false 为零值会被 Gin 拒收
	Reason        string `json:"reason" binding:"omitempty,max=200"`
}

// 领养协议签署请求
type AdoptionAgreementRequest struct {
	ApplicationID uint   `json:"applicationId" binding:"required"`
	AgreementText string `json:"agreementText" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
}

// 领养状态更新请求
type AdoptionStatusUpdateRequest struct {
	AdoptionID uint `json:"adoptionId" binding:"required"`
	Status     int  `json:"status" binding:"required,oneof=0 1 2"`
}

// 领养统计信息
type AdoptionStats struct {
	TotalAdoptions     int64 `json:"totalAdoptions"`
	PendingAdoptions   int64 `json:"pendingAdoptions"`
	AdoptedAdoptions   int64 `json:"adoptedAdoptions"`
	CancelledAdoptions int64 `json:"cancelledAdoptions"`
	TotalApplications  int64 `json:"totalApplications"`
}

// 回访记录信息
type AdoptionFollowUpInfo struct {
	ID            uint     `json:"id"`
	AdoptionID    uint     `json:"adoptionId"`
	ApplicationID uint     `json:"applicationId"`
	UserID        uint     `json:"userId"`
	UserNickname  string   `json:"userNickname"`
	UserAvatar    string   `json:"userAvatar"`
	FollowUpType  string   `json:"followUpType"`
	Content       string   `json:"content"`
	Photos        []string `json:"photos,omitempty"`
	Status        int      `json:"status"`
	CreatedAt     string   `json:"createdAt"`
}

// 创建回访记录请求
type AdoptionFollowUpCreateRequest struct {
	AdoptionID    uint     `json:"adoptionId" binding:"required"`
	ApplicationID uint     `json:"applicationId" binding:"required"`
	FollowUpType  string   `json:"followUpType" binding:"required,oneof=visit photo online"`
	Content       string   `json:"content" binding:"required,max=500"`
	Photos        []string `json:"photos" binding:"omitempty,max=9"`
	Status        int      `json:"status" binding:"omitempty,oneof=0 1"`
}

// 回访记录列表响应
type AdoptionFollowUpListResponse struct {
	List     []AdoptionFollowUpInfo `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

// UploadPhotoResponse 领养照片上传响应
type UploadPhotoResponse struct {
	Url string `json:"url"`
}
