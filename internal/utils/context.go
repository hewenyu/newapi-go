package utils

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// 上下文键类型
type contextKey string

// 上下文键常量
const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	TraceIDKey   contextKey = "trace_id"
	TimeoutKey   contextKey = "timeout"
	RetryKey     contextKey = "retry_count"
	APIKeyKey    contextKey = "api_key"
	BaseURLKey   contextKey = "base_url"
	ModelKey     contextKey = "model"
)

// 默认超时时间
const (
	DefaultTimeout       = 30 * time.Second
	DefaultStreamTimeout = 60 * time.Second
	DefaultRetryCount    = 3
)

// WithRequestID 添加请求ID到上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithUserID 添加用户ID到上下文
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID 从上下文中获取用户ID
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithTraceID 添加跟踪ID到上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 从上下文中获取跟踪ID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithTimeout 添加超时时间到上下文
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return context.WithTimeout(ctx, timeout)
}

// WithRetryCount 添加重试次数到上下文
func WithRetryCount(ctx context.Context, retryCount int) context.Context {
	return context.WithValue(ctx, RetryKey, retryCount)
}

// GetRetryCount 从上下文中获取重试次数
func GetRetryCount(ctx context.Context) int {
	if count, ok := ctx.Value(RetryKey).(int); ok {
		return count
	}
	return 0
}

// WithAPIKey 添加API密钥到上下文
func WithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, APIKeyKey, apiKey)
}

// GetAPIKey 从上下文中获取API密钥
func GetAPIKey(ctx context.Context) string {
	if apiKey, ok := ctx.Value(APIKeyKey).(string); ok {
		return apiKey
	}
	return ""
}

// WithBaseURL 添加基础URL到上下文
func WithBaseURL(ctx context.Context, baseURL string) context.Context {
	return context.WithValue(ctx, BaseURLKey, baseURL)
}

// GetBaseURL 从上下文中获取基础URL
func GetBaseURL(ctx context.Context) string {
	if baseURL, ok := ctx.Value(BaseURLKey).(string); ok {
		return baseURL
	}
	return ""
}

// WithModel 添加模型名到上下文
func WithModel(ctx context.Context, model string) context.Context {
	return context.WithValue(ctx, ModelKey, model)
}

// GetModel 从上下文中获取模型名
func GetModel(ctx context.Context) string {
	if model, ok := ctx.Value(ModelKey).(string); ok {
		return model
	}
	return ""
}

// NewRequestContext 创建新的请求上下文
func NewRequestContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// 生成请求ID
	requestID := GenerateRequestID()

	// 添加请求ID到上下文
	ctx = WithRequestID(ctx, requestID)

	// 添加超时控制
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return context.WithTimeout(ctx, timeout)
}

// NewStreamContext 创建新的流式上下文
func NewStreamContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// 生成请求ID
	requestID := GenerateRequestID()

	// 添加请求ID到上下文
	ctx = WithRequestID(ctx, requestID)

	// 添加超时控制
	if timeout <= 0 {
		timeout = DefaultStreamTimeout
	}

	return context.WithTimeout(ctx, timeout)
}

// GenerateRequestID 生成唯一请求ID
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果生成失败，使用时间戳
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%x", bytes)
}

// GenerateTraceID 生成唯一跟踪ID
func GenerateTraceID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// 如果生成失败，使用时间戳
		return fmt.Sprintf("trace_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("trace_%x", bytes)
}

// IsTimeout 检查错误是否为超时错误
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	// 检查是否为上下文超时
	if err == context.DeadlineExceeded {
		return true
	}

	// 检查是否为上下文取消
	if err == context.Canceled {
		return true
	}

	return false
}

// IsCanceled 检查错误是否为取消错误
func IsCanceled(err error) bool {
	return err == context.Canceled
}

// GetEffectiveTimeout 获取有效的超时时间
func GetEffectiveTimeout(ctx context.Context, defaultTimeout time.Duration) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 {
			return remaining
		}
	}
	return defaultTimeout
}

// NewBackgroundContext 创建新的后台上下文
func NewBackgroundContext() context.Context {
	ctx := context.Background()
	ctx = WithRequestID(ctx, GenerateRequestID())
	ctx = WithTraceID(ctx, GenerateTraceID())
	return ctx
}

// CloneContext 克隆上下文，保留重要信息
func CloneContext(ctx context.Context) context.Context {
	newCtx := context.Background()

	// 复制请求ID
	if requestID := GetRequestID(ctx); requestID != "" {
		newCtx = WithRequestID(newCtx, requestID)
	}

	// 复制用户ID
	if userID := GetUserID(ctx); userID != "" {
		newCtx = WithUserID(newCtx, userID)
	}

	// 复制跟踪ID
	if traceID := GetTraceID(ctx); traceID != "" {
		newCtx = WithTraceID(newCtx, traceID)
	}

	// 复制API密钥
	if apiKey := GetAPIKey(ctx); apiKey != "" {
		newCtx = WithAPIKey(newCtx, apiKey)
	}

	// 复制基础URL
	if baseURL := GetBaseURL(ctx); baseURL != "" {
		newCtx = WithBaseURL(newCtx, baseURL)
	}

	// 复制模型名
	if model := GetModel(ctx); model != "" {
		newCtx = WithModel(newCtx, model)
	}

	return newCtx
}

// ContextWithValues 创建带有多个值的上下文
func ContextWithValues(ctx context.Context, values map[contextKey]interface{}) context.Context {
	for key, value := range values {
		ctx = context.WithValue(ctx, key, value)
	}
	return ctx
}

// GetContextInfo 获取上下文信息摘要
func GetContextInfo(ctx context.Context) map[string]interface{} {
	info := make(map[string]interface{})

	if requestID := GetRequestID(ctx); requestID != "" {
		info["request_id"] = requestID
	}

	if userID := GetUserID(ctx); userID != "" {
		info["user_id"] = userID
	}

	if traceID := GetTraceID(ctx); traceID != "" {
		info["trace_id"] = traceID
	}

	if model := GetModel(ctx); model != "" {
		info["model"] = model
	}

	if retryCount := GetRetryCount(ctx); retryCount > 0 {
		info["retry_count"] = retryCount
	}

	if deadline, ok := ctx.Deadline(); ok {
		info["deadline"] = deadline
		info["timeout"] = time.Until(deadline)
	}

	return info
}
