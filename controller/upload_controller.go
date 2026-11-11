package controller

import (
	"strings"

	"pet-community/service"
	"pet-community/types"
	"pet-community/utils"

	"github.com/gin-gonic/gin"
)

// UploadController 通用上传接口控制器。
// 用于没有专属上传端点的场景（目前是活动作品提交的图片/视频）：
// 需要登录态，落盘统一走 service.UploadService → utils.Storage。
type UploadController struct {
	uploadService *service.UploadService
}

// NewUploadController 构造通用上传 controller
func NewUploadController(uploadService *service.UploadService) *UploadController {
	return &UploadController{uploadService: uploadService}
}

// UploadMedia 上传单个媒体文件（multipart field=file），返回 url 与 type(image|video)。
// 图片上限 5MB、视频上限 50MB，均由 Storage 层按扩展名判定后校验。
func (c *UploadController) UploadMedia(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "未收到上传文件："+err.Error())
		return
	}

	ext := ""
	if idx := strings.LastIndex(fileHeader.Filename, "."); idx >= 0 {
		ext = fileHeader.Filename[idx+1:]
	}
	if ext == "" {
		utils.RespondError(ctx, types.CodeParamError, "无法识别文件扩展名")
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		utils.RespondError(ctx, types.CodeParamError, "无法读取上传文件")
		return
	}
	defer f.Close()

	url, kind, err := c.uploadService.UploadMedia(f, ext)
	if err != nil {
		utils.RespondError(ctx, types.CodeBusinessError, err.Error())
		return
	}

	utils.RespondSuccess(ctx, &types.UploadMediaResponse{Url: url, Type: kind})
}
