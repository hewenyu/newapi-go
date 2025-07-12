package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hewenyu/newapi-go/internal/utils"
)

// HTTPTransport HTTP传输层接口
type HTTPTransport interface {
	// 基础HTTP请求
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
	DoJSON(ctx context.Context, req *http.Request, result interface{}) error
	DoStream(ctx context.Context, req *http.Request) (io.ReadCloser, error)

	// 便捷方法
	Get(ctx context.Context, path string, params url.Values) (*http.Response, error)
	Post(ctx context.Context, path string, body interface{}) (*http.Response, error)
	Put(ctx context.Context, path string, body interface{}) (*http.Response, error)
	Delete(ctx context.Context, path string) (*http.Response, error)

	// 流式方法
	PostStream(ctx context.Context, path string, body interface{}) (StreamReader, error)

	// 配置方法
	SetTimeout(timeout time.Duration)
	SetRetryPolicy(policy RetryPolicy)
	SetMiddleware(middleware ...Middleware)

	// 资源管理
	Close() error
}

// HTTPClient HTTP客户端实现
type HTTPClient struct {
	client          *http.Client
	requestBuilder  *RequestBuilder
	responseHandler *ResponseHandler
	retryPolicy     RetryPolicy
	middleware      []Middleware
	mu              sync.RWMutex
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(baseURL, apiKey string, options ...HTTPOption) *HTTPClient {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   10,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	httpClient := &HTTPClient{
		client:          client,
		requestBuilder:  NewRequestBuilder(baseURL, apiKey, 30*time.Second),
		responseHandler: NewResponseHandler(32 * 1024 * 1024), // 32MB
		retryPolicy:     NewDefaultRetryPolicy(),
		middleware:      make([]Middleware, 0),
	}

	// 应用选项
	for _, option := range options {
		option(httpClient)
	}

	return httpClient
}

// Do 执行HTTP请求
func (hc *HTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return hc.doWithRetry(ctx, req)
}

// DoJSON 执行HTTP请求并解析JSON响应
func (hc *HTTPClient) DoJSON(ctx context.Context, req *http.Request, result interface{}) error {
	resp, err := hc.Do(ctx, req)
	if err != nil {
		return err
	}

	startTime := time.Now()
	return hc.responseHandler.HandleJSONResponse(ctx, resp, result, startTime)
}

// DoStream 执行流式HTTP请求
func (hc *HTTPClient) DoStream(ctx context.Context, req *http.Request) (io.ReadCloser, error) {
	resp, err := hc.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	return hc.responseHandler.HandleStreamResponse(ctx, resp, startTime)
}

// Get 发送GET请求
func (hc *HTTPClient) Get(ctx context.Context, path string, params url.Values) (*http.Response, error) {
	fullPath := path
	if params != nil {
		fullPath = fmt.Sprintf("%s?%s", path, params.Encode())
	}

	req, err := hc.requestBuilder.BuildRequest(ctx, http.MethodGet, fullPath, nil)
	if err != nil {
		return nil, err
	}

	return hc.Do(ctx, req)
}

// Post 发送POST请求
func (hc *HTTPClient) Post(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	req, err := hc.requestBuilder.BuildRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	return hc.Do(ctx, req)
}

// Put 发送PUT请求
func (hc *HTTPClient) Put(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	req, err := hc.requestBuilder.BuildRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}

	return hc.Do(ctx, req)
}

// Delete 发送DELETE请求
func (hc *HTTPClient) Delete(ctx context.Context, path string) (*http.Response, error) {
	req, err := hc.requestBuilder.BuildRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	return hc.Do(ctx, req)
}

// PostStream 发送流式POST请求
func (hc *HTTPClient) PostStream(ctx context.Context, path string, body interface{}) (StreamReader, error) {
	req, err := hc.requestBuilder.BuildStreamRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	reader, err := hc.DoStream(ctx, req)
	if err != nil {
		return nil, err
	}

	return NewJSONStreamReader(ctx, reader), nil
}

// SetTimeout 设置超时时间
func (hc *HTTPClient) SetTimeout(timeout time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.client.Timeout = timeout
	hc.requestBuilder.SetTimeout(timeout)
}

// SetRetryPolicy 设置重试策略
func (hc *HTTPClient) SetRetryPolicy(policy RetryPolicy) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.retryPolicy = policy
}

// SetMiddleware 设置中间件
func (hc *HTTPClient) SetMiddleware(middleware ...Middleware) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.middleware = middleware
}

// Close 关闭客户端
func (hc *HTTPClient) Close() error {
	if transport, ok := hc.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// doWithRetry 执行带重试的请求
func (hc *HTTPClient) doWithRetry(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	retryCount := 0
	maxRetries := hc.retryPolicy.MaxRetries()

	for {
		// 应用中间件
		finalHandler := hc.executeRequest
		for i := len(hc.middleware) - 1; i >= 0; i-- {
			finalHandler = hc.middleware[i](finalHandler)
		}

		resp, err = finalHandler(ctx, req)

		// 成功或不可重试错误
		if err == nil || retryCount >= maxRetries {
			break
		}

		// 检查是否可重试
		if !hc.shouldRetry(ctx, req, resp, err, retryCount) {
			break
		}

		// 计算延迟时间
		delay := hc.retryPolicy.BackoffDelay(retryCount)
		utils.LogError(ctx, err, "Request failed, retrying")

		// 等待重试
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}

		retryCount++
	}

	return resp, err
}

// executeRequest 执行请求
func (hc *HTTPClient) executeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return hc.client.Do(req)
}

// shouldRetry 判断是否应该重试
func (hc *HTTPClient) shouldRetry(ctx context.Context, req *http.Request, resp *http.Response, err error, retryCount int) bool {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return false
	}

	// 检查是否超过最大重试次数
	if retryCount >= hc.retryPolicy.MaxRetries() {
		return false
	}

	// 网络错误检查
	if err != nil {
		return hc.retryPolicy.ShouldRetry(ctx, req, resp, err, retryCount)
	}

	// HTTP状态码检查
	if resp != nil {
		return hc.responseHandler.ShouldRetry(resp) &&
			hc.retryPolicy.ShouldRetry(ctx, req, resp, err, retryCount)
	}

	return false
}

// HTTPOption HTTP客户端选项
type HTTPOption func(*HTTPClient)

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) HTTPOption {
	return func(hc *HTTPClient) {
		hc.SetTimeout(timeout)
	}
}

// WithRetryPolicy 设置重试策略
func WithRetryPolicy(policy RetryPolicy) HTTPOption {
	return func(hc *HTTPClient) {
		hc.SetRetryPolicy(policy)
	}
}

// WithMiddleware 添加中间件
func WithMiddleware(middleware ...Middleware) HTTPOption {
	return func(hc *HTTPClient) {
		hc.SetMiddleware(middleware...)
	}
}

// WithTLSConfig 设置TLS配置
func WithTLSConfig(config *tls.Config) HTTPOption {
	return func(hc *HTTPClient) {
		if transport, ok := hc.client.Transport.(*http.Transport); ok {
			transport.TLSClientConfig = config
		}
	}
}

// WithProxy 设置代理
func WithProxy(proxyURL string) HTTPOption {
	return func(hc *HTTPClient) {
		if proxyURL == "" {
			return
		}

		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return
		}

		if transport, ok := hc.client.Transport.(*http.Transport); ok {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}
}

// Middleware 中间件类型
type Middleware func(next HTTPHandler) HTTPHandler

// HTTPHandler HTTP处理器
type HTTPHandler func(ctx context.Context, req *http.Request) (*http.Response, error)

// RetryPolicy 重试策略接口
type RetryPolicy interface {
	MaxRetries() int
	BackoffDelay(retryCount int) time.Duration
	ShouldRetry(ctx context.Context, req *http.Request, resp *http.Response, err error, retryCount int) bool
}

// DefaultRetryPolicy 默认重试策略
type DefaultRetryPolicy struct {
	maxRetries int
	baseDelay  time.Duration
}

// NewDefaultRetryPolicy 创建默认重试策略
func NewDefaultRetryPolicy() *DefaultRetryPolicy {
	return &DefaultRetryPolicy{
		maxRetries: 3,
		baseDelay:  1 * time.Second,
	}
}

// MaxRetries 获取最大重试次数
func (p *DefaultRetryPolicy) MaxRetries() int {
	return p.maxRetries
}

// BackoffDelay 计算退避延迟
func (p *DefaultRetryPolicy) BackoffDelay(retryCount int) time.Duration {
	// 指数退避
	delay := p.baseDelay * time.Duration(math.Pow(2, float64(retryCount)))

	// 最大延迟30秒
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}

	return delay
}

// ShouldRetry 判断是否应该重试
func (p *DefaultRetryPolicy) ShouldRetry(ctx context.Context, req *http.Request, resp *http.Response, err error, retryCount int) bool {
	// 检查上下文
	if ctx.Err() != nil {
		return false
	}

	// 网络错误
	if err != nil {
		if netErr, ok := err.(net.Error); ok {
			return netErr.Temporary() || netErr.Timeout()
		}
		return true
	}

	// HTTP状态码
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout:      // 504
			return true
		}
	}

	return false
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(next HTTPHandler) HTTPHandler {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		resp, err := next(ctx, req)

		logger := utils.GetLogger().WithContext(ctx)
		if err != nil {
			utils.LogError(ctx, err, "HTTP request failed")
		} else {
			logger.Info("HTTP request completed")
		}

		return resp, err
	}
}

// UserAgentMiddleware 用户代理中间件
func UserAgentMiddleware(userAgent string) Middleware {
	return func(next HTTPHandler) HTTPHandler {
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			if userAgent != "" {
				req.Header.Set("User-Agent", userAgent)
			}
			return next(ctx, req)
		}
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(next HTTPHandler) HTTPHandler {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		// 这里可以实现速率限制逻辑
		// 暂时直接调用下一个处理器
		return next(ctx, req)
	}
}
