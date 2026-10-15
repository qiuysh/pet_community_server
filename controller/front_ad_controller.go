package controller

import (
	"strings"

	"pet-community/service"
	"pet-community/types"

	"github.com/gin-gonic/gin"
)

// FrontAdController 前台广告展示控制器。
// 与 AdminAdController 不同:不挂 AuthMiddleware,任何访客均可查询启用中的广告,
// 并可上报 PV/Click 用于效果统计。
//
// 设计要点:
//   - location 必须命中枚举 home_banner/splash/detail_sidebar/topic_banner,
//     与 admin 控制器 CreateAdPositionRequest 的 binding 保持一致,避免脏值。
//   - 仅暴露查询与计数三类只读/幂等接口,CRUD 仍走 admin 域。
type FrontAdController struct {
	adService *service.AdService
}

func NewFrontAdController(adService *service.AdService) *FrontAdController {
	return &FrontAdController{adService: adService}
}

// allowedAdLocations 与前端 AdLocation 类型保持一致,避免拼写漂移。
var allowedAdLocations = map[string]bool{
	"home_banner":     true,
	"splash":          true,
	"detail_sidebar":  true,
	"topic_banner":    true,
}

// ListActiveAds GET /api/v1/ads?location=home_banner
// 返回当前生效中的广告列表(位置启用 + 时间窗 + status=1)。
func (c *FrontAdController) ListActiveAds(ctx *gin.Context) {
	location := strings.TrimSpace(ctx.Query("location"))
	if location == "" {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "location 参数必填"))
		return
	}
	if !allowedAdLocations[location] {
		ctx.JSON(200, types.ErrorResponse(types.CodeParamError, "无效的广告位置"))
		return
	}

	ads, err := c.adService.GetActiveAds(location)
	if err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}

	list := make([]types.AdvertisementResponse, len(ads))
	for i, a := range ads {
		list[i] = adToResponse(&a)
	}
	ctx.JSON(200, types.SuccessResponse(list))
}

// IncrementAdView POST /api/v1/ads/:id/view
// 前台广告曝光上报,失败不阻塞展示,仅返回错误码。
func (c *FrontAdController) IncrementAdView(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	if err := c.adService.IncrementViewCount(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}

// IncrementAdClick POST /api/v1/ads/:id/click
// 前台广告点击上报,用于统计 CTR。
func (c *FrontAdController) IncrementAdClick(ctx *gin.Context) {
	id, ok := MustParseUintParam(ctx, "id", "广告ID")
	if !ok {
		return
	}
	if err := c.adService.IncrementClickCount(id); err != nil {
		ctx.JSON(200, types.ErrorResponse(types.CodeSystemError, err.Error()))
		return
	}
	ctx.JSON(200, types.SuccessResponse(nil))
}
