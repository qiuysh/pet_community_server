package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"pet-community/config"
)

// ossStorage 对象存储驱动：以 AWS Signature V4 直连 S3 兼容端点
// （阿里云 OSS / 腾讯云 COS / MinIO / AWS S3 等），零第三方依赖。
type ossStorage struct {
	endpoint    string
	region      string
	bucket      string
	accessKey   string
	secretKey   string
	keyPrefix   string
	baseURL     string
	scheme      string
	pathStyle   bool
	client      *http.Client
	driverName  string
	host        string
	signingName string
}

const ossRequestTimeout = 60 * time.Second

func newOSSStorage(cfg config.OSSStorageConfig, driver string) (*ossStorage, error) {
	endpoint := strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://")
	if endpoint == "" {
		return nil, fmt.Errorf("oss endpoint is required")
	}
	if cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("oss bucket / access_key_id / access_key_secret are required")
	}
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	scheme := "https"
	if !cfg.UseSSL {
		scheme = "http"
	}

	s := &ossStorage{
		endpoint:    endpoint,
		region:      region,
		bucket:      cfg.Bucket,
		accessKey:   cfg.AccessKeyID,
		secretKey:   cfg.AccessKeySecret,
		keyPrefix:   strings.Trim(cfg.KeyPrefix, "/"),
		baseURL:     strings.TrimSuffix(cfg.BaseURL, "/"),
		scheme:      scheme,
		pathStyle:   cfg.PathStyle,
		driverName:  driver,
		signingName: "s3",
		client:      &http.Client{Timeout: ossRequestTimeout},
	}
	if s.pathStyle {
		s.host = endpoint
	} else {
		s.host = cfg.Bucket + "." + endpoint
	}
	return s, nil
}

// Save 上传到对象存储，返回对外可访问 URL
func (s *ossStorage) Save(data io.Reader, opt SaveOptions) (string, error) {
	key, err := NewObjectKey(opt.Category, opt.Ext)
	if err != nil {
		return "", err
	}

	// 体积上限内多读 1 字节用于判定超限，避免先落临时文件
	buf := new(bytes.Buffer)
	written, err := io.Copy(buf, io.LimitReader(data, opt.MaxSize+1))
	if err != nil {
		return "", fmt.Errorf("read upload data: %w", err)
	}
	if err := checkSize(written, opt.MaxSize, opt.Kind); err != nil {
		return "", err
	}

	if err := s.putObject(key, buf.Bytes(), contentTypeOf(opt.Ext)); err != nil {
		return "", err
	}
	return s.URL(key), nil
}

// URL 生成对外访问地址（优先使用配置的 CDN / 自定义域名）
func (s *ossStorage) URL(key string) string {
	objectKey := s.objectKey(key)
	if s.baseURL != "" {
		return s.baseURL + "/" + objectKey
	}
	if s.pathStyle {
		return s.scheme + "://" + s.endpoint + "/" + s.bucket + "/" + objectKey
	}
	return s.scheme + "://" + s.host + "/" + objectKey
}

// emptyPayloadHash SHA-256("")，DELETE 等无请求体操作签名用
const emptyPayloadHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// Delete 删除对象，不存在时视为成功（与 PUT 一样走 SigV4）
func (s *ossStorage) Delete(key string) error {
	objectKey := s.objectKey(key)
	req, err := s.newRequest(ossRequest{method: http.MethodDelete, objectKey: objectKey})
	if err != nil {
		return err
	}
	if err := s.sign(req, emptyPayloadHash); err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete object: %s", resp.Status)
	}
	return nil
}

// Driver 返回驱动名
func (s *ossStorage) Driver() string { return s.driverName }

// objectKey 给对象键拼上统一的 keyPrefix
func (s *ossStorage) objectKey(key string) string {
	key = strings.TrimPrefix(key, "/")
	if s.keyPrefix == "" {
		return key
	}
	return s.keyPrefix + "/" + key
}

// putObject 以 SigV4 签名执行 PUT Object
func (s *ossStorage) putObject(key string, body []byte, contentType string) error {
	objectKey := s.objectKey(key)
	sum := sha256.Sum256(body)
	payloadHash := hex.EncodeToString(sum[:])

	// x-amz-content-sha256 由 sign 统一写入，此处只传业务头
	req, err := s.newRequest(ossRequest{
		method:      http.MethodPut,
		objectKey:   objectKey,
		body:        bytes.NewReader(body),
		contentType: contentType,
	})
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(body))

	if err := s.sign(req, payloadHash); err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("upload object: %s %s", resp.Status, strings.TrimSpace(string(respBody)))
	}
	return nil
}

// ossRequest 构造对象存储请求所需的参数（收拢字段，避免方法参数过多）
type ossRequest struct {
	// method HTTP 方法，目前为 PUT / DELETE
	method string
	// objectKey 已拼上 keyPrefix 的对象键
	objectKey string
	// body 请求体，DELETE 等无体请求留空
	body io.Reader
	// contentType 对象 Content-Type，仅 PUT 需要
	contentType string
}

// newRequest 构造对象存储请求（virtual-hosted 与 path-style 两种风格）
func (s *ossStorage) newRequest(spec ossRequest) (*http.Request, error) {
	var uri string
	if s.pathStyle {
		uri = "/" + s.bucket + "/" + spec.objectKey
	} else {
		uri = "/" + spec.objectKey
	}
	req, err := http.NewRequest(spec.method, s.scheme+"://"+s.host+uri, spec.body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if spec.contentType != "" {
		req.Header.Set("content-type", spec.contentType)
	}
	return req, nil
}

// sign 写入 AWS Signature V4 所需的 x-amz-date 与 Authorization 头
func (s *ossStorage) sign(req *http.Request, payloadHash string) error {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	req.Header.Set("x-amz-date", amzDate)
	if payloadHash != "" {
		req.Header.Set("x-amz-content-sha256", payloadHash)
	}

	canonicalURI := req.URL.EscapedPath()
	signedHeaders, canonicalHeaders := canonicalHeaderSet(req)

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	scope := strings.Join([]string{dateStamp, s.region, s.signingName, "aws4_request"}, "/")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		scope,
		hex.EncodeToString(sha256Sum([]byte(canonicalRequest))),
	}, "\n")

	signingKey := hmacSHA256([]byte("AWS4"+s.secretKey), dateStamp)
	signingKey = hmacSHA256(signingKey, s.region)
	signingKey = hmacSHA256(signingKey, s.signingName)
	signingKey = hmacSHA256(signingKey, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, scope, signedHeaders, signature,
	))
	return nil
}

// canonicalHeaderSet 收集需参与签名的头（host + x-amz-* + content-type）
func canonicalHeaderSet(req *http.Request) (signedHeaders, canonicalHeaders string) {
	host := req.URL.Host
	if host == "" {
		host = req.Host
	}
	pairs := [][2]string{{"host", strings.TrimSpace(host)}}
	for name, values := range req.Header {
		lower := strings.ToLower(name)
		if lower != "host" && !strings.HasPrefix(lower, "x-amz-") && lower != "content-type" {
			continue
		}
		if lower == "host" || len(values) == 0 {
			continue
		}
		pairs = append(pairs, [2]string{lower, strings.TrimSpace(values[0])})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i][0] < pairs[j][0] })

	names := make([]string, 0, len(pairs))
	var canonical strings.Builder
	for _, p := range pairs {
		names = append(names, p[0])
		canonical.WriteString(p[0])
		canonical.WriteString(":")
		canonical.WriteString(strings.Join(strings.Fields(p[1]), " "))
		canonical.WriteString("\n")
	}
	return strings.Join(names, ";"), canonical.String()
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// contentTypeOf 按扩展名给出 Content-Type，保证浏览器可直接渲染 / 播放
func contentTypeOf(ext string) string {
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "mp4":
		return "video/mp4"
	case "webm":
		return "audio/webm"
	case "ogg", "oga":
		return "audio/ogg"
	case "m4a":
		return "audio/mp4"
	case "aac":
		return "audio/aac"
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	default:
		return "application/octet-stream"
	}
}
