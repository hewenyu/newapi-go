package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/types"
)

// RequestBuilder HTTP请求构建器
type RequestBuilder struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	headers map[string]string
}

// NewRequestBuilder 创建新的请求构建器
func NewRequestBuilder(baseURL, apiKey string, timeout time.Duration) *RequestBuilder {
	return &RequestBuilder{
		baseURL: baseURL,
		apiKey:  apiKey,
		timeout: timeout,
		headers: make(map[string]string),
	}
}

// WithHeader 添加头部
func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// WithHeaders 添加多个头部
func (rb *RequestBuilder) WithHeaders(headers map[string]string) *RequestBuilder {
	for key, value := range headers {
		rb.headers[key] = value
	}
	return rb
}

// BuildRequest 构建HTTP请求
func (rb *RequestBuilder) BuildRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// 构建完整URL
	fullURL, err := rb.buildURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	// 构建请求体
	var reader io.Reader
	var contentType string

	if body != nil {
		switch v := body.(type) {
		case string:
			reader = strings.NewReader(v)
			contentType = "text/plain"
		case []byte:
			reader = bytes.NewReader(v)
			contentType = "application/octet-stream"
		case io.Reader:
			reader = v
			contentType = "application/octet-stream"
		default:
			// JSON序列化
			data, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reader = bytes.NewReader(data)
			contentType = "application/json"
		}
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置通用头部
	rb.setCommonHeaders(req)

	// 设置Content-Type
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	// 设置自定义头部
	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	// 记录请求日志
	utils.LogAPIRequest(ctx, method, fullURL, rb.getHeaderMap(req), body)

	return req, nil
}

// BuildStreamRequest 构建流式请求
func (rb *RequestBuilder) BuildStreamRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	req, err := rb.BuildRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}

	// 设置流式头部
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req, nil
}

// BuildFormRequest 构建表单请求
func (rb *RequestBuilder) BuildFormRequest(ctx context.Context, method, path string, form url.Values) (*http.Request, error) {
	fullURL, err := rb.buildURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	var reader io.Reader
	if form != nil {
		reader = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置通用头部
	rb.setCommonHeaders(req)

	// 设置表单头部
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 设置自定义头部
	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	// 记录请求日志
	utils.LogAPIRequest(ctx, method, fullURL, rb.getHeaderMap(req), form)

	return req, nil
}

// BuildMultipartRequest 构建multipart请求
func (rb *RequestBuilder) BuildMultipartRequest(ctx context.Context, method, path, boundary string, body io.Reader) (*http.Request, error) {
	fullURL, err := rb.buildURL(path)
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置通用头部
	rb.setCommonHeaders(req)

	// 设置multipart头部
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

	// 设置自定义头部
	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	// 记录请求日志
	utils.LogAPIRequest(ctx, method, fullURL, rb.getHeaderMap(req), "[multipart data]")

	return req, nil
}

// buildURL 构建完整URL
func (rb *RequestBuilder) buildURL(path string) (string, error) {
	base, err := url.Parse(rb.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	rel, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	return base.ResolveReference(rel).String(), nil
}

// setCommonHeaders 设置通用头部
func (rb *RequestBuilder) setCommonHeaders(req *http.Request) {
	// 设置认证头部
	if rb.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rb.apiKey))
	}

	// 设置用户代理
	req.Header.Set("User-Agent", "newapi-go-sdk/1.0.0")

	// 设置接受类型
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}

	// 设置请求ID
	if requestID := utils.GetRequestID(req.Context()); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
}

// getHeaderMap 获取头部映射
func (rb *RequestBuilder) getHeaderMap(req *http.Request) map[string]string {
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// Clone 克隆请求构建器
func (rb *RequestBuilder) Clone() *RequestBuilder {
	headers := make(map[string]string)
	for key, value := range rb.headers {
		headers[key] = value
	}

	return &RequestBuilder{
		baseURL: rb.baseURL,
		apiKey:  rb.apiKey,
		timeout: rb.timeout,
		headers: headers,
	}
}

// SetTimeout 设置超时时间
func (rb *RequestBuilder) SetTimeout(timeout time.Duration) *RequestBuilder {
	rb.timeout = timeout
	return rb
}

// GetTimeout 获取超时时间
func (rb *RequestBuilder) GetTimeout() time.Duration {
	return rb.timeout
}

// SetBaseURL 设置基础URL
func (rb *RequestBuilder) SetBaseURL(baseURL string) *RequestBuilder {
	rb.baseURL = baseURL
	return rb
}

// GetBaseURL 获取基础URL
func (rb *RequestBuilder) GetBaseURL() string {
	return rb.baseURL
}

// SetAPIKey 设置API密钥
func (rb *RequestBuilder) SetAPIKey(apiKey string) *RequestBuilder {
	rb.apiKey = apiKey
	return rb
}

// ValidateRequest 验证请求
func (rb *RequestBuilder) ValidateRequest(method, path string, body interface{}) error {
	if rb.baseURL == "" {
		return types.NewAPIError(types.ErrTypeValidation, types.ErrCodeMissingParameter, "base URL is required", http.StatusBadRequest)
	}

	if rb.apiKey == "" {
		return types.NewAPIError(types.ErrTypeAuthentication, types.ErrCodeInvalidAPIKey, "API key is required", http.StatusUnauthorized)
	}

	if method == "" {
		return types.NewAPIError(types.ErrTypeValidation, types.ErrCodeMissingParameter, "HTTP method is required", http.StatusBadRequest)
	}

	if path == "" {
		return types.NewAPIError(types.ErrTypeValidation, types.ErrCodeMissingParameter, "path is required", http.StatusBadRequest)
	}

	// 验证URL格式
	_, err := url.Parse(rb.baseURL)
	if err != nil {
		return types.NewAPIError(types.ErrTypeValidation, types.ErrCodeInvalidParameter, "invalid base URL format", http.StatusBadRequest)
	}

	return nil
}
