package types

// UploadMediaResponse 通用上传响应。
// Url 为对外可访问地址（本地驱动为 /api/v1/uploads/... 前缀，OSS 驱动为对象存储域名），
// Type 为后端按扩展名判定的媒体类型：image / video。
type UploadMediaResponse struct {
	Url  string `json:"url"`
	Type string `json:"type"`
}
