package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"pet-community/controller"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
	"pet-community/types"
)

// setupFrontAdController 构建一个直接挂载 FrontAdController 的 gin 引擎,
// 便于在 HTTP 层面验证 location 校验、PV/Click 上报。
func setupFrontAdController(t *testing.T) (*gin.Engine, *service.AdService) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	adRepo := repository.NewAdRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	adService := service.NewAdService(adRepo, adminRepo)
	frontAdCtrl := controller.NewFrontAdController(adService)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/ads", frontAdCtrl.ListActiveAds)
	api.POST("/ads/:id/view", frontAdCtrl.IncrementAdView)
	api.POST("/ads/:id/click", frontAdCtrl.IncrementAdClick)
	return r, adService
}

// seedLiveAd 通过 service 层创建一条生效中的广告,返回其 ID。
// 包含 ReviewAd(status=1) 才会被 GetActiveAds 返回。
func seedLiveAd(t *testing.T, adService *service.AdService) uint {
	t.Helper()
	pos, err := adService.CreatePosition("Home Banner", "home_banner", "750x300", 1)
	if err != nil {
		t.Fatalf("create position failed: %v", err)
	}
	now := time.Now()
	start := now.Add(-time.Hour).Format("2006-01-02 15:04:05")
	end := now.Add(time.Hour).Format("2006-01-02 15:04:05")
	ad, err := adService.CreateAd(pos.ID, "live", "/img/live.jpg", "/live", start, end, 1)
	if err != nil {
		t.Fatalf("create ad failed: %v", err)
	}
	if err := adService.ReviewAd(ad.ID, 1, 1, "127.0.0.1"); err != nil {
		t.Fatalf("review ad failed: %v", err)
	}
	return ad.ID
}

// 接口层:location 缺失 / 非法 两个分支
func TestFrontAdController_ListActiveAds_Validation(t *testing.T) {
	r, _ := setupFrontAdController(t)

	t.Run("missing location returns param error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ads", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: %d", rec.Code)
		}
		var resp types.Response
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Code != types.CodeParamError {
			t.Errorf("expected CodeParamError, got %d", resp.Code)
		}
	})

	t.Run("invalid location returns param error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ads?location=evil", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		var resp types.Response
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Code != types.CodeParamError {
			t.Errorf("expected CodeParamError, got %d", resp.Code)
		}
	})
}

func TestFrontAdController_ListActiveAds_OK(t *testing.T) {
	r, adService := setupFrontAdController(t)
	adID := seedLiveAd(t, adService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ads?location=home_banner", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var resp struct {
		Code int                            `json:"code"`
		Data []types.AdvertisementResponse  `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Code != types.CodeSuccess {
		t.Fatalf("expected success, got code=%d msg=%s", resp.Code, rec.Body.String())
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 ad, got %d", len(resp.Data))
	}
	if resp.Data[0].Title != "live" {
		t.Errorf("expected title 'live', got %q", resp.Data[0].Title)
	}
	if resp.Data[0].ID != adID {
		t.Errorf("expected id %d, got %d", adID, resp.Data[0].ID)
	}
}

// PV/Click 上报:成功 + 无效 ID
func TestFrontAdController_ViewAndClick(t *testing.T) {
	r, adService := setupFrontAdController(t)
	adID := seedLiveAd(t, adService)

	t.Run("view increments view count", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/ads/%d/view", adID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: %d", rec.Code)
		}
		updated, _ := adService.GetAdByID(adID)
		if updated.ViewCount != 1 {
			t.Errorf("expected view count 1, got %d", updated.ViewCount)
		}
	})

	t.Run("click increments click count", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/ads/%d/click", adID), nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status: %d", rec.Code)
		}
		updated, _ := adService.GetAdByID(adID)
		if updated.ClickCount != 1 {
			t.Errorf("expected click count 1, got %d", updated.ClickCount)
		}
	})

	t.Run("invalid id returns param error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ads/abc/view", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		var resp types.Response
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp.Code != types.CodeParamError {
			t.Errorf("expected CodeParamError, got %d", resp.Code)
		}
	})
}
