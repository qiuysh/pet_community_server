package service

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pet-community/config"
)

// setupThirdPartyConfig 注入测试用第三方配置，并在用例结束后还原
func setupThirdPartyConfig(t *testing.T, cfg config.ThirdPartyConfig, env string) {
	t.Helper()
	prev := config.Cfg
	config.Cfg = &config.Config{
		Server:     config.ServerConfig{Env: env},
		ThirdParty: cfg,
	}
	t.Cleanup(func() { config.Cfg = prev })
}

// withTempURL 临时替换平台默认地址，避免用例真的打到外网
func withTempURL(t *testing.T, target *string, value string) {
	t.Helper()
	prev := *target
	*target = value
	t.Cleanup(func() { *target = prev })
}

func TestResolveThirdPartyIdentityRejectsUnconfiguredPlatform(t *testing.T) {
	setupThirdPartyConfig(t, config.ThirdPartyConfig{DemoMode: false}, "production")

	identity, err := resolveThirdPartyIdentity("wechat", "any-code")
	if err == nil {
		t.Fatalf("未配置微信凭据时必须拒绝登录，实际返回 openID=%q", identity.OpenID)
	}
	if !strings.Contains(err.Error(), "未配置") {
		t.Fatalf("错误信息应说明未配置，实际: %v", err)
	}
}

func TestResolveThirdPartyIdentityDemoModeSkipsOAuthInDev(t *testing.T) {
	setupThirdPartyConfig(t, config.ThirdPartyConfig{DemoMode: true}, "development")

	identity, err := resolveThirdPartyIdentity("wechat", "demo-code")
	if err != nil {
		t.Fatalf("开发环境演示模式应可用，实际报错: %v", err)
	}
	if identity.OpenID != "wechat_demo-code" {
		t.Fatalf("演示模式 openID 派生不符合预期: %q", identity.OpenID)
	}
}

func TestResolveThirdPartyIdentityDemoModeRejectedInProduction(t *testing.T) {
	// 即使配置被改成 demo_mode=true，生产环境也不能用（否则可伪造任意身份）
	setupThirdPartyConfig(t, config.ThirdPartyConfig{DemoMode: true}, "production")

	if _, err := resolveThirdPartyIdentity("wechat", "any-code"); err == nil {
		t.Fatal("生产环境不允许演示模式第三方登录")
	}
}

func TestResolveThirdPartyIdentityWeChatUsesOpenIDFromPlatform(t *testing.T) {
	var tokenHit, infoHit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenHit = true
			if r.URL.Query().Get("code") != "real-code" {
				t.Errorf("换取 token 应带上 code，实际 %q", r.URL.Query().Get("code"))
			}
			if r.URL.Query().Get("secret") != "secret-1" {
				t.Errorf("换取 token 应带上 app secret")
			}
			_, _ = w.Write([]byte(`{"access_token":"at-1","openid":"platform-openid","errcode":0}`))
		case "/userinfo":
			infoHit = true
			_, _ = w.Write([]byte(`{"nickname":"平台昵称","headimgurl":"https://img.example/a.png"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	withTempURL(t, &defaultWeChatTokenURL, srv.URL+"/token")
	withTempURL(t, &defaultWeChatUserInfoURL, srv.URL+"/userinfo")

	setupThirdPartyConfig(t, config.ThirdPartyConfig{
		WeChat: config.ThirdPartyPlatformConfig{AppID: "app-1", AppSecret: "secret-1"},
	}, "production")

	identity, err := resolveThirdPartyIdentity("wechat", "real-code")
	if err != nil {
		t.Fatalf("换取身份失败: %v", err)
	}
	if !tokenHit || !infoHit {
		t.Fatalf("应调用平台接口，实际 token=%v userinfo=%v", tokenHit, infoHit)
	}
	// 关键断言：openID 来自平台，而不是 code 拼装
	if identity.OpenID != "platform-openid" {
		t.Fatalf("openID 必须取自平台响应，实际 %q", identity.OpenID)
	}
	if identity.Nickname != "平台昵称" || identity.Avatar != "https://img.example/a.png" {
		t.Fatalf("昵称/头像未回填: %+v", identity)
	}
}

func TestResolveThirdPartyIdentityWeChatErrorCodePropagates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40029,"errmsg":"invalid code"}`))
	}))
	defer srv.Close()
	withTempURL(t, &defaultWeChatTokenURL, srv.URL)

	setupThirdPartyConfig(t, config.ThirdPartyConfig{
		WeChat: config.ThirdPartyPlatformConfig{AppID: "app-1", AppSecret: "secret-1"},
	}, "production")

	if _, err := resolveThirdPartyIdentity("wechat", "bad-code"); err == nil {
		t.Fatal("平台返回 errcode 时必须报错，不能当成登录成功")
	}
}

func TestResolveThirdPartyIdentityWeiboUsesUID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			if err := r.ParseForm(); err != nil {
				t.Errorf("解析表单失败: %v", err)
			}
			if r.Form.Get("code") != "wb-code" {
				t.Errorf("微博换取 token 应带上 code")
			}
			_, _ = w.Write([]byte(`{"access_token":"wb-at","uid":"1234567"}`))
		case "/userinfo":
			_, _ = w.Write([]byte(`{"screen_name":"微博昵称","profile_image_url":"https://img.example/w.png"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	withTempURL(t, &defaultWeiboTokenURL, srv.URL+"/token")
	withTempURL(t, &defaultWeiboUserInfoURL, srv.URL+"/userinfo")

	setupThirdPartyConfig(t, config.ThirdPartyConfig{
		Weibo: config.ThirdPartyPlatformConfig{AppID: "wb-app", AppSecret: "wb-secret"},
	}, "production")

	identity, err := resolveThirdPartyIdentity("weibo", "wb-code")
	if err != nil {
		t.Fatalf("微博换取身份失败: %v", err)
	}
	if identity.OpenID != "1234567" {
		t.Fatalf("openID 应取平台 uid，实际 %q", identity.OpenID)
	}
	if identity.Nickname != "微博昵称" {
		t.Fatalf("昵称未回填: %+v", identity)
	}
}

func TestResolveThirdPartyIdentityQQUsesMeEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"qq-at"}`))
		case "/me":
			_, _ = w.Write([]byte(`{"client_id":"qq-app","openid":"qq-openid"}`))
		case "/userinfo":
			_, _ = w.Write([]byte(`{"ret":0,"nickname":"QQ昵称","figureurl_qq_1":"https://img.example/q.png"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	withTempURL(t, &defaultQQTokenURL, srv.URL+"/token")
	withTempURL(t, &defaultQQMeURL, srv.URL+"/me")
	withTempURL(t, &defaultQQUserInfoURL, srv.URL+"/userinfo")

	setupThirdPartyConfig(t, config.ThirdPartyConfig{
		QQ: config.ThirdPartyPlatformConfig{AppID: "qq-app", AppSecret: "qq-secret"},
	}, "production")

	identity, err := resolveThirdPartyIdentity("qq", "qq-code")
	if err != nil {
		t.Fatalf("QQ 换取身份失败: %v", err)
	}
	if identity.OpenID != "qq-openid" {
		t.Fatalf("openID 应取 /me 返回，实际 %q", identity.OpenID)
	}
}

func TestResolveThirdPartyIdentityPlatformFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	withTempURL(t, &defaultWeChatTokenURL, srv.URL)

	setupThirdPartyConfig(t, config.ThirdPartyConfig{
		WeChat: config.ThirdPartyPlatformConfig{AppID: "app-1", AppSecret: "secret-1"},
	}, "production")

	if _, err := resolveThirdPartyIdentity("wechat", "code"); err == nil {
		t.Fatal("平台返回 5xx 时必须报错")
	}
}

func TestResolveThirdPartyIdentityRejectsEmptyCode(t *testing.T) {
	setupThirdPartyConfig(t, config.ThirdPartyConfig{DemoMode: true}, "development")
	if _, err := resolveThirdPartyIdentity("wechat", "   "); err == nil {
		t.Fatal("空 code 必须报错")
	}
}
