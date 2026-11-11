package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"pet-community/controller"
	"pet-community/model"
	"pet-community/repository"
	"pet-community/service"
	"pet-community/tests/testutil"
)

// siteConfigPayload 与前端 SiteConfig 对齐的公开字段。
type siteConfigPayload struct {
	SiteName        string `json:"siteName"`
	SiteLogo        string `json:"siteLogo"`
	SiteDescription string `json:"siteDescription"`
	SiteICP         string `json:"siteICP"`
}

type siteConfigEnvelope struct {
	Code int               `json:"code"`
	Data siteConfigPayload `json:"data"`
}

// setupSiteController 挂载前台站点配置接口，返回引擎、DB 与配置服务。
func setupSiteController(t *testing.T) (*gin.Engine, *gorm.DB, *service.AdminConfigService) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	adminRepo := repository.NewAdminRepository(db)
	configSvc := service.NewAdminConfigService(adminRepo)
	siteCtrl := controller.NewSiteController(configSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/site/config", siteCtrl.GetSiteConfig)
	return r, db, configSvc
}

// requestSiteConfig 调用公开接口并解析响应。
func requestSiteConfig(t *testing.T, r *gin.Engine) siteConfigEnvelope {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/site/config", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var body siteConfigEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body failed: %v", err)
	}
	if body.Code != 0 {
		t.Fatalf("expected success code, got %d", body.Code)
	}
	return body
}

// 未配置任何 key 时接口应返回空串而不是系统错误（前台据此回退默认站点名）。
func TestSiteController_GetSiteConfig_Empty(t *testing.T) {
	r, _, _ := setupSiteController(t)

	body := requestSiteConfig(t, r)
	if body.Data.SiteName != "" || body.Data.SiteICP != "" {
		t.Fatalf("expected empty config, got %+v", body.Data)
	}
}

// 配置写入后公开接口应按 camelCase 回显。
func TestSiteController_GetSiteConfig_Configured(t *testing.T) {
	r, db, _ := setupSiteController(t)

	rows := []model.SystemConfig{
		{ConfigKey: "site_name", ConfigValue: "喵星社区"},
		{ConfigKey: "site_logo", ConfigValue: "https://cdn.example.com/logo.png"},
		{ConfigKey: "site_description", ConfigValue: "记录每只猫的日常"},
		{ConfigKey: "site_icp", ConfigValue: "京ICP备00000000号"},
	}
	for _, row := range rows {
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("seed config failed: %v", err)
		}
	}

	body := requestSiteConfig(t, r)
	if body.Data.SiteName != "喵星社区" {
		t.Errorf("siteName: %s", body.Data.SiteName)
	}
	if body.Data.SiteLogo != "https://cdn.example.com/logo.png" {
		t.Errorf("siteLogo: %s", body.Data.SiteLogo)
	}
	if body.Data.SiteDescription != "记录每只猫的日常" {
		t.Errorf("siteDescription: %s", body.Data.SiteDescription)
	}
	if body.Data.SiteICP != "京ICP备00000000号" {
		t.Errorf("siteICP: %s", body.Data.SiteICP)
	}
}

// seed 缺失时后台保存应插入而非静默丢失，公开接口随后可读到新值。
func TestSiteController_GetSiteConfig_AfterSaveWithoutSeed(t *testing.T) {
	r, _, configSvc := setupSiteController(t)

	configs := map[string]string{"site_name": "新站点名", "site_icp": "沪ICP备11111111号"}
	if err := configSvc.UpdateSiteConfig(configs, 1, "127.0.0.1"); err != nil {
		t.Fatalf("update site config failed: %v", err)
	}

	body := requestSiteConfig(t, r)
	if body.Data.SiteName != "新站点名" {
		t.Errorf("siteName: %s", body.Data.SiteName)
	}
	if body.Data.SiteICP != "沪ICP备11111111号" {
		t.Errorf("siteICP: %s", body.Data.SiteICP)
	}
}
