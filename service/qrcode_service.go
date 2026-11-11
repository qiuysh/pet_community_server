package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"pet-community/config"
)

// QRCodeService 二维码服务
type QRCodeService struct {
	secretKey string // 用于签名验证的密钥
}

// NewQRCodeService 从 config.yaml 的 qrcode.secret 读取密钥；
// 留空则启动时随机生成并告警（进程重启后旧码失效，生产务必配置）。
func NewQRCodeService() *QRCodeService {
	secret := ""
	if config.Cfg != nil {
		secret = strings.TrimSpace(config.Cfg.QRCode.Secret)
	}
	if secret == "" {
		secret = randomQRCodeSecret()
		log.Printf("WARN qrcode.secret is empty; using ephemeral key (restart invalidates active check-in codes). Set qrcode.secret in config.yaml")
	}
	return &QRCodeService{secretKey: secret}
}

// NewQRCodeServiceWithSecret 测试用：注入固定密钥
func NewQRCodeServiceWithSecret(secret string) *QRCodeService {
	return &QRCodeService{secretKey: secret}
}

func randomQRCodeSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		// 极端情况下退回时间戳派生，仍会打告警；正常路径不会走到这里
		return fmt.Sprintf("ephemeral-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

// GenerateQRCode 生成二维码内容
func (s *QRCodeService) GenerateQRCode(activityID uint) (string, string) {
	now := time.Now()
	expires := now.Add(30 * time.Minute) // 二维码有效期30分钟

	// 生成签名
	data := fmt.Sprintf("%d-%d", activityID, expires.Unix())
	signature := s.generateSignature(data)

	// 二维码内容格式：activityID-expires-signature
	qrCode := fmt.Sprintf("%d-%d-%s", activityID, expires.Unix(), signature)

	return qrCode, expires.Format(time.RFC3339)
}

// ValidateQRCode 验证二维码
func (s *QRCodeService) ValidateQRCode(qrCode string) (uint, error) {
	parts := splitQRCode(qrCode)
	if len(parts) != 3 {
		return 0, fmt.Errorf("二维码格式错误")
	}

	activityID, err := parseUint(parts[0])
	if err != nil {
		return 0, fmt.Errorf("活动ID无效")
	}

	expires, err := parseInt64(parts[1])
	if err != nil {
		return 0, fmt.Errorf("过期时间无效")
	}

	signature := parts[2]

	// 检查是否过期
	if time.Now().Unix() > expires {
		return 0, fmt.Errorf("二维码已过期")
	}

	// 验证签名（恒定时间比较，避免计时旁路）
	data := fmt.Sprintf("%d-%d", activityID, expires)
	expectedSignature := s.generateSignature(data)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return 0, fmt.Errorf("二维码签名无效")
	}

	return activityID, nil
}

// generateSignature 生成签名
func (s *QRCodeService) generateSignature(data string) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))[:16] // 使用前16位作为签名
}

// splitQRCode 分割二维码内容
func splitQRCode(qrCode string) []string {
	result := []string{}
	start := 0
	for i := 0; i < len(qrCode); i++ {
		if qrCode[i] == '-' {
			result = append(result, qrCode[start:i])
			start = i + 1
		}
	}
	if start < len(qrCode) {
		result = append(result, qrCode[start:])
	}
	return result
}

// parseUint 解析uint
func parseUint(s string) (uint, error) {
	var result uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + uint(c-'0')
		} else {
			return 0, fmt.Errorf("invalid character")
		}
	}
	return result, nil
}

// parseInt64 解析int64
func parseInt64(s string) (int64, error) {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		} else {
			return 0, fmt.Errorf("invalid character")
		}
	}
	return result, nil
}

// QRCodeData 二维码数据（用于前端展示）
type QRCodeData struct {
	Content string `json:"content"` // 二维码内容
	Expires string `json:"expires"` // 过期时间
}
