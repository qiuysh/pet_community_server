package service

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
)

const (
	adminCaptchaTTL      = 60 * time.Second
	adminCaptchaLength   = 4
	adminCaptchaRequired = 3 // 连续失败达到该次数后强制验证码
)

var (
	ErrCaptchaRequired = errors.New("请填写图形验证码")
	ErrCaptchaInvalid  = errors.New("验证码错误或已过期")
)

// AdminLoginError 登录失败，并告知前端是否需要验证码。
type AdminLoginError struct {
	Msg            string
	RequireCaptcha bool
}

func (e *AdminLoginError) Error() string { return e.Msg }

func (e *AdminLoginError) Unwrap() error {
	if e.Msg == ErrCaptchaRequired.Error() {
		return ErrCaptchaRequired
	}
	if e.Msg == ErrCaptchaInvalid.Error() {
		return ErrCaptchaInvalid
	}
	return errors.New(e.Msg)
}

// ttlCaptchaStore 带 60 秒过期的内存验证码仓库（一次性消费）。
type ttlCaptchaStore struct {
	mu   sync.Mutex
	data map[string]captchaEntry
}

type captchaEntry struct {
	answer    string
	expiresAt time.Time
}

func newTTLCaptchaStore() *ttlCaptchaStore {
	return &ttlCaptchaStore{data: make(map[string]captchaEntry)}
}

func (s *ttlCaptchaStore) Set(id string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = captchaEntry{answer: value, expiresAt: time.Now().Add(adminCaptchaTTL)}
	return nil
}

func (s *ttlCaptchaStore) Get(id string, clear bool) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.data[id]
	if !ok {
		return ""
	}
	if time.Now().After(entry.expiresAt) {
		delete(s.data, id)
		return ""
	}
	if clear {
		delete(s.data, id)
	}
	return entry.answer
}

func (s *ttlCaptchaStore) Verify(id, answer string, clear bool) bool {
	got := s.Get(id, clear)
	if got == "" || answer == "" {
		return false
	}
	return strings.EqualFold(got, answer)
}

var adminCaptchaStore = newTTLCaptchaStore()

// GenerateAdminCaptcha 生成 4 位字符图形验证码（base64 图 + 一次性 id）。
func GenerateAdminCaptcha() (id string, b64s string, err error) {
	driver := base64Captcha.NewDriverString(
		80, 240, 0,
		base64Captcha.OptionShowHollowLine,
		adminCaptchaLength,
		"23456789ABCDEFGHJKLMNPQRSTUVWXYZ",
		nil, nil, nil,
	)
	c := base64Captcha.NewCaptcha(driver, adminCaptchaStore)
	id, b64s, _, err = c.Generate()
	return id, b64s, err
}

// VerifyAdminCaptcha 校验并消费验证码（一次性）；失败返回 ErrCaptchaInvalid。
func VerifyAdminCaptcha(id, code string) error {
	id = strings.TrimSpace(id)
	code = strings.TrimSpace(code)
	if id == "" || code == "" {
		return ErrCaptchaInvalid
	}
	if !adminCaptchaStore.Verify(id, code, true) {
		return ErrCaptchaInvalid
	}
	return nil
}

// SeedAdminCaptchaForTest 测试注入已知验证码（生产勿用）。
func SeedAdminCaptchaForTest(id, answer string) {
	_ = adminCaptchaStore.Set(id, answer)
}
