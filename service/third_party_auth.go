package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"pet-community/config"
	"pet-community/utils"
)

// 第三方平台官方接口地址（可被配置覆盖，便于自建代理与测试打桩）
// 声明为变量而非常量：测试可用 httptest 替换，配置项优先于这里的默认值。
var (
	defaultWeChatTokenURL    = "https://api.weixin.qq.com/sns/oauth2/access_token"
	defaultWeChatUserInfoURL = "https://api.weixin.qq.com/sns/userinfo"

	defaultWeiboTokenURL    = "https://api.weibo.com/oauth2/access_token"
	defaultWeiboUserInfoURL = "https://api.weibo.com/2/users/show.json"

	defaultQQTokenURL    = "https://graph.qq.com/oauth2.0/token"
	defaultQQMeURL       = "https://graph.qq.com/oauth2.0/me"
	defaultQQUserInfoURL = "https://graph.qq.com/user/get_user_info"
)

// thirdPartyHTTPTimeout 调用第三方 OAuth 接口的超时时间
const thirdPartyHTTPTimeout = 10 * time.Second

// thirdPartyMaxBodyBytes 第三方响应体最大读取量，防止异常响应撑爆内存
const thirdPartyMaxBodyBytes = 64 << 10

// thirdPartyHTTPClient 第三方 OAuth 专用 HTTP 客户端
var thirdPartyHTTPClient = &http.Client{Timeout: thirdPartyHTTPTimeout}

// thirdPartyIdentity 第三方平台返回的身份信息
type thirdPartyIdentity struct {
	// OpenID 平台侧用户唯一标识，是账号绑定的唯一依据
	OpenID string
	// Nickname 平台昵称，可能为空
	Nickname string
	// Avatar 平台头像，可能为空
	Avatar string
}

// platformDisplayName 平台中文名，用于给用户看的提示
func platformDisplayName(platform string) string {
	switch platform {
	case "wechat":
		return "微信"
	case "weibo":
		return "微博"
	case "qq":
		return "QQ"
	default:
		return platform
	}
}

// resolveThirdPartyIdentity 用授权码换取第三方身份。
//
// 安全约束：openID 必须由平台接口返回。禁止根据 code 在本地拼装 openID ——
// 那样任何客户端都能构造任意 openID，等同可伪造他人身份登录。
func resolveThirdPartyIdentity(platform, code string) (*thirdPartyIdentity, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("授权码无效")
	}

	tp := config.Cfg.ThirdParty
	if tp.DemoMode {
		// 演示模式只允许开发环境：config 加载时已对非开发环境强制关闭，
		// 这里再兜一层，避免热改配置绕过。
		if !config.IsDev() {
			utils.LogError(errors.New("demo mode enabled outside development"), "ThirdParty.DemoMode")
			return nil, errors.New("第三方登录暂未开放，请使用手机号登录")
		}
		return &thirdPartyIdentity{OpenID: platform + "_" + code}, nil
	}

	var pc config.ThirdPartyPlatformConfig
	switch platform {
	case "wechat":
		pc = tp.WeChat
	case "weibo":
		pc = tp.Weibo
	case "qq":
		pc = tp.QQ
	default:
		return nil, errors.New("不支持的平台")
	}

	if !pc.Configured() {
		return nil, fmt.Errorf("%s登录未配置，请使用手机号登录", platformDisplayName(platform))
	}

	switch platform {
	case "wechat":
		return exchangeWeChatIdentity(pc, code)
	case "weibo":
		return exchangeWeiboIdentity(pc, code)
	default:
		return exchangeQQIdentity(pc, code)
	}
}

// exchangeWeChatIdentity 微信网页授权：code → access_token/openid → 用户资料
func exchangeWeChatIdentity(pc config.ThirdPartyPlatformConfig, code string) (*thirdPartyIdentity, error) {
	tokenURL := firstNonEmpty(pc.TokenURL, defaultWeChatTokenURL)
	u, err := url.Parse(tokenURL)
	if err != nil {
		return nil, errors.New("微信登录配置有误")
	}
	q := u.Query()
	q.Set("appid", pc.AppID)
	q.Set("secret", pc.AppSecret)
	q.Set("code", code)
	q.Set("grant_type", "authorization_code")
	u.RawQuery = q.Encode()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		OpenID      string `json:"openid"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := httpGetJSON(u.String(), &tokenRes); err != nil {
		return nil, fmt.Errorf("微信授权失败：%w", err)
	}
	if tokenRes.ErrCode != 0 {
		return nil, fmt.Errorf("微信授权失败：%s", firstNonEmpty(tokenRes.ErrMsg, "code 无效"))
	}
	if tokenRes.OpenID == "" || tokenRes.AccessToken == "" {
		return nil, errors.New("微信授权失败：未返回 openid")
	}

	identity := &thirdPartyIdentity{OpenID: tokenRes.OpenID}
	// 昵称/头像属于锦上添花，失败不影响登录
	infoURL, err := url.Parse(firstNonEmpty(pc.UserInfoURL, defaultWeChatUserInfoURL))
	if err == nil {
		iq := infoURL.Query()
		iq.Set("access_token", tokenRes.AccessToken)
		iq.Set("openid", tokenRes.OpenID)
		iq.Set("lang", "zh_CN")
		infoURL.RawQuery = iq.Encode()
		var info struct {
			Nickname string `json:"nickname"`
			HeadImg  string `json:"headimgurl"`
		}
		if err := httpGetJSON(infoURL.String(), &info); err == nil {
			identity.Nickname = info.Nickname
			identity.Avatar = info.HeadImg
		}
	}
	return identity, nil
}

// exchangeWeiboIdentity 微博 OAuth：POST code 换 token/uid → 用户资料
func exchangeWeiboIdentity(pc config.ThirdPartyPlatformConfig, code string) (*thirdPartyIdentity, error) {
	form := url.Values{}
	form.Set("client_id", pc.AppID)
	form.Set("client_secret", pc.AppSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	if pc.RedirectURI != "" {
		form.Set("redirect_uri", pc.RedirectURI)
	}

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		UID         string `json:"uid"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := httpPostFormJSON(firstNonEmpty(pc.TokenURL, defaultWeiboTokenURL), form, &tokenRes); err != nil {
		return nil, fmt.Errorf("微博授权失败：%w", err)
	}
	if tokenRes.Error != "" {
		return nil, fmt.Errorf("微博授权失败：%s", firstNonEmpty(tokenRes.ErrorDesc, tokenRes.Error))
	}
	if tokenRes.UID == "" || tokenRes.AccessToken == "" {
		return nil, errors.New("微博授权失败：未返回 uid")
	}

	identity := &thirdPartyIdentity{OpenID: tokenRes.UID}
	infoURL, err := url.Parse(firstNonEmpty(pc.UserInfoURL, defaultWeiboUserInfoURL))
	if err == nil {
		iq := infoURL.Query()
		iq.Set("access_token", tokenRes.AccessToken)
		iq.Set("uid", tokenRes.UID)
		infoURL.RawQuery = iq.Encode()
		var info struct {
			ScreenName string `json:"screen_name"`
			Avatar     string `json:"profile_image_url"`
		}
		if err := httpGetJSON(infoURL.String(), &info); err == nil {
			identity.Nickname = info.ScreenName
			identity.Avatar = info.Avatar
		}
	}
	return identity, nil
}

// exchangeQQIdentity QQ 互联：code → access_token → openid → 用户资料
func exchangeQQIdentity(pc config.ThirdPartyPlatformConfig, code string) (*thirdPartyIdentity, error) {
	tokenURL, err := url.Parse(firstNonEmpty(pc.TokenURL, defaultQQTokenURL))
	if err != nil {
		return nil, errors.New("QQ 登录配置有误")
	}
	q := tokenURL.Query()
	q.Set("grant_type", "authorization_code")
	q.Set("client_id", pc.AppID)
	q.Set("client_secret", pc.AppSecret)
	q.Set("code", code)
	if pc.RedirectURI != "" {
		q.Set("redirect_uri", pc.RedirectURI)
	}
	q.Set("fmt", "json")
	tokenURL.RawQuery = q.Encode()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		Error       int    `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := httpGetJSON(tokenURL.String(), &tokenRes); err != nil {
		return nil, fmt.Errorf("QQ 授权失败：%w", err)
	}
	if tokenRes.AccessToken == "" {
		return nil, fmt.Errorf("QQ 授权失败：%s", firstNonEmpty(tokenRes.ErrorDesc, "code 无效"))
	}

	meURL, err := url.Parse(defaultQQMeURL)
	if err != nil {
		return nil, errors.New("QQ 登录配置有误")
	}
	mq := meURL.Query()
	mq.Set("access_token", tokenRes.AccessToken)
	mq.Set("fmt", "json")
	meURL.RawQuery = mq.Encode()

	var meRes struct {
		OpenID string `json:"openid"`
		Error  int    `json:"error"`
	}
	if err := httpGetJSON(meURL.String(), &meRes); err != nil {
		return nil, fmt.Errorf("QQ 授权失败：%w", err)
	}
	if meRes.OpenID == "" {
		return nil, errors.New("QQ 授权失败：未返回 openid")
	}

	identity := &thirdPartyIdentity{OpenID: meRes.OpenID}
	infoURL, err := url.Parse(firstNonEmpty(pc.UserInfoURL, defaultQQUserInfoURL))
	if err == nil {
		iq := infoURL.Query()
		iq.Set("access_token", tokenRes.AccessToken)
		iq.Set("oauth_consumer_key", pc.AppID)
		iq.Set("openid", meRes.OpenID)
		infoURL.RawQuery = iq.Encode()
		var info struct {
			Ret      int    `json:"ret"`
			Nickname string `json:"nickname"`
			Avatar   string `json:"figureurl_qq_1"`
		}
		if err := httpGetJSON(infoURL.String(), &info); err == nil && info.Ret == 0 {
			identity.Nickname = info.Nickname
			identity.Avatar = info.Avatar
		}
	}
	return identity, nil
}

// httpGetJSON 发起 GET 并按 JSON 解析响应
func httpGetJSON(rawURL string, out interface{}) error {
	resp, err := thirdPartyHTTPClient.Get(rawURL)
	if err != nil {
		return errors.New("网络请求失败")
	}
	return decodeThirdPartyResponse(resp, out)
}

// httpPostFormJSON 以表单方式 POST 并按 JSON 解析响应
func httpPostFormJSON(rawURL string, form url.Values, out interface{}) error {
	resp, err := thirdPartyHTTPClient.PostForm(rawURL, form)
	if err != nil {
		return errors.New("网络请求失败")
	}
	return decodeThirdPartyResponse(resp, out)
}

// decodeThirdPartyResponse 校验状态码并解析响应体（调用方负责关闭 resp.Body 由本函数完成）
func decodeThirdPartyResponse(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, thirdPartyMaxBodyBytes))
	if err != nil {
		return errors.New("读取响应失败")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("平台返回状态码 %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return errors.New("响应格式异常")
	}
	return nil
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
