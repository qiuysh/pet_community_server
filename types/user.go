package types

type UserRegisterRequest struct {
	Phone    string `json:"phone" binding:"omitempty,len=11"`
	Email    string `json:"email" binding:"omitempty,email,max=100"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Code     string `json:"code" binding:"required,len=6"`
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=20"`
}

// NotificationInfo 通知信息
type NotificationInfo struct {
	ID            uint   `json:"id"`
	Type          string `json:"type"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	TriggerType   string `json:"triggerType"`
	TriggerID     uint   `json:"triggerId"`
	TriggerUserID uint   `json:"triggerUserId"`
	IsRead        int    `json:"isRead"`
	CreatedAt     string `json:"createdAt"`
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	List     []NotificationInfo `json:"list"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

// UnreadCountResponse 未读消息数响应
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// FollowInfo 粉丝/关注信息
type FollowInfo struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"userId"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Signature   string `json:"signature"`
	IsFollowing bool   `json:"isFollowing"`
	// IsSecret 仅在查看自己关注列表时返回，用于渲染悄悄关注标识（F6-04）
	IsSecret bool `json:"isSecret"`
}

// FollowListResponse 粉丝/关注列表响应
type FollowListResponse struct {
	List     []FollowInfo `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

// FollowCountResponse 关注/粉丝数响应
type FollowCountResponse struct {
	FollowersCount int64 `json:"followersCount"`
	FollowingCount int64 `json:"followingCount"`
}

type UserLoginRequest struct {
	Account   string `json:"account" binding:"required"`
	Password  string `json:"password" binding:"omitempty,min=6"`
	Code      string `json:"code" binding:"omitempty,len=6"`
	LoginType string `json:"loginType" binding:"required,oneof=password code"`
	Remember  bool   `json:"remember"` // 记住登录：RefreshToken 7d；否则 ≤2h
}

type UserLoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	DeviceID     string   `json:"deviceId,omitempty"` // 当前会话设备，登出/踢设备用
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Gender    int    `json:"gender"`
	Signature string `json:"signature"`
	Location  string `json:"location"`
	PetYears  string `json:"petYears"`
	Status    int    `json:"status"`
	Level     int    `json:"level"`     // 用户等级
	IsVipName bool   `json:"isVipName"` // Lv10 专属昵称标识
}

// PublicUserProfile 公开个人主页字段（不含手机号/邮箱）
type PublicUserProfile struct {
	ID             uint   `json:"id"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Gender         int    `json:"gender"`
	Signature      string `json:"signature"`
	Location       string `json:"location"`
	PetYears       string `json:"petYears"`
	PostsCount     int64  `json:"postsCount"`
	FollowersCount int64  `json:"followersCount"`
	FollowingCount int64  `json:"followingCount"`
	PetsCount      int64  `json:"petsCount"`
	Level          int    `json:"level"`
	IsVipName      bool   `json:"isVipName"`
}

type SendCodeRequest struct {
	Account string `json:"account" binding:"required"`
	Type    string `json:"type" binding:"required,oneof=register login reset_password change_password bind_phone bind_email"`
}

type SendCodeResponse struct {
	ExpiresIn int    `json:"expiresIn"`
	DebugCode string `json:"debugCode,omitempty"` // 未配置短信/邮件时返回，便于联调；生产勿依赖
}

type UserProfileUpdateRequest struct {
	Nickname  string `json:"nickname" binding:"omitempty,min=2,max=20"`
	Avatar    string `json:"avatar" binding:"omitempty,max=255"`
	Signature string `json:"signature" binding:"omitempty,max=100"`
	Gender    *int   `json:"gender" binding:"omitempty,oneof=0 1 2"`
	Location  string `json:"location" binding:"omitempty,max=100"`
	PetYears  string `json:"petYears" binding:"omitempty,max=20"`
}

type PasswordChangeRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,min=6"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type PasswordResetRequest struct {
	Account     string `json:"account" binding:"required"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type PrivacySettingsUpdateRequest struct {
	PostVisibility    int `json:"postVisibility" binding:"omitempty,oneof=0 1 2"`
	CommentPermission int `json:"commentPermission" binding:"omitempty,oneof=0 1 2"`
	MessagePermission int `json:"messagePermission" binding:"omitempty,oneof=0 1 2"`
}

type UserListRequest struct {
	Keyword  string `form:"keyword"`
	Status   int    `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type UserListResponse struct {
	List     []UserInfo `json:"list"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

type LogoutRequest struct {
	DeviceID string `json:"deviceId"`
}

type DeviceInfo struct {
	ID           uint   `json:"id"`
	DeviceID     string `json:"deviceId"`
	DeviceName   string `json:"deviceName"`
	DeviceType   string `json:"deviceType"`
	IP           string `json:"ip"`
	LastActiveAt string `json:"lastActiveAt"`
}

type ThirdPartyLoginRequest struct {
	Platform string `json:"platform" binding:"required,oneof=wechat weibo qq"`
	Code     string `json:"code" binding:"required"`
}

type ThirdPartyLoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	DeviceID     string   `json:"deviceId,omitempty"`
	User         UserInfo `json:"user"`
	IsNewUser    bool     `json:"isNewUser"`
}

type ThirdPartyBindRequest struct {
	Platform string `json:"platform" binding:"required"`
	Phone    string `json:"phone" binding:"required,len=11"`
	Code     string `json:"code" binding:"required,len=6"`
}

// BindThirdPartyAccountRequest 已登录用户绑定第三方账号（演示模式用 code）
type BindThirdPartyAccountRequest struct {
	Platform string `json:"platform" binding:"required,oneof=wechat weibo qq"`
	Code     string `json:"code" binding:"required"`
}

type ThirdPartyAccountInfo struct {
	Platform  string `json:"platform"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"createdAt"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type PrivacySettings struct {
	PostVisibility    int `json:"postVisibility"`
	CommentPermission int `json:"commentPermission"`
	MessagePermission int `json:"messagePermission"`
}

type UnbindPhoneRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type UnbindEmailRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// BindPhoneRequest 绑定/更换手机号
// account 用作发送验证码的载体;code 由用户输入
type BindPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required,len=6"`
}

// BindEmailRequest 绑定/更换邮箱
type BindEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}
