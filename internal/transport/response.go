package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hewenyu/newapi-go/internal/utils"
	"github.com/hewenyu/newapi-go/types"
)

// ResponseHandler 响应处理器
type ResponseHandler struct {
	maxBodySize int64
}

// NewResponseHandler 创建新的响应处理器
func NewResponseHandler(maxBodySize int64) *ResponseHandler {
	if maxBodySize <= 0 {
		maxBodySize = 32 * 1024 * 1024 // 32MB default
	}
	return &ResponseHandler{
		maxBodySize: maxBodySize,
	}
}

// HandleResponse 处理HTTP响应
func (rh *ResponseHandler) HandleResponse(ctx context.Context, resp *http.Response, startTime time.Time) (*types.BaseResponse, error) {
	defer resp.Body.Close()

	// 计算响应时间
	duration := time.Since(startTime).Milliseconds()

	// 读取响应体
	body, err := rh.readBody(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 记录响应日志
	utils.LogAPIResponse(ctx, resp.StatusCode, rh.getHeaderMap(resp), string(body), duration)

	// 检查HTTP状态码
	if resp.StatusCode >= 400 {
		return nil, rh.handleErrorResponse(ctx, resp, body)
	}

	// 解析成功响应
	return rh.parseResponse(ctx, resp, body)
}

// HandleJSONResponse 处理JSON响应
func (rh *ResponseHandler) HandleJSONResponse(ctx context.Context, resp *http.Response, result interface{}, startTime time.Time) error {
	defer resp.Body.Close()

	// 计算响应时间
	duration := time.Since(startTime).Milliseconds()

	// 读取响应体
	body, err := rh.readBody(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// 记录响应日志
	utils.LogAPIResponse(ctx, resp.StatusCode, rh.getHeaderMap(resp), string(body), duration)

	// 检查HTTP状态码
	if resp.StatusCode >= 400 {
		return rh.handleErrorResponse(ctx, resp, body)
	}

	// 解析JSON响应
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeParseError,
				fmt.Sprintf("failed to parse JSON response: %v", err), resp.StatusCode)
		}
	}

	return nil
}

// HandleStreamResponse 处理流式响应
func (rh *ResponseHandler) HandleStreamResponse(ctx context.Context, resp *http.Response, startTime time.Time) (io.ReadCloser, error) {
	// 计算响应时间
	duration := time.Since(startTime).Milliseconds()

	// 记录响应日志
	utils.LogAPIResponse(ctx, resp.StatusCode, rh.getHeaderMap(resp), "[stream data]", duration)

	// 检查HTTP状态码
	if resp.StatusCode >= 400 {
		// 对于流式响应，需要读取错误信息
		body, err := rh.readBody(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response body: %w", err)
		}
		return nil, rh.handleErrorResponse(ctx, resp, body)
	}

	// 检查Content-Type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		return nil, types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeInvalidRequest,
			"invalid content type for stream response", resp.StatusCode)
	}

	// 返回响应体用于流式处理
	return resp.Body, nil
}

// readBody 读取响应体
func (rh *ResponseHandler) readBody(body io.Reader) ([]byte, error) {
	// 限制读取大小
	limitedReader := io.LimitReader(body, rh.maxBodySize)

	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	// 检查是否超过最大大小
	if int64(len(data)) >= rh.maxBodySize {
		return nil, types.NewAPIError(types.ErrTypeAPIError, types.ErrCodePayloadTooLarge,
			"response body too large", http.StatusRequestEntityTooLarge)
	}

	return data, nil
}

// handleErrorResponse 处理错误响应
func (rh *ResponseHandler) handleErrorResponse(ctx context.Context, resp *http.Response, body []byte) error {
	// 尝试解析标准错误响应
	var errorResp types.ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Message != "" {
		return types.NewAPIError(errorResp.Type, errorResp.Code, errorResp.Message, resp.StatusCode)
	}

	// 尝试解析OpenAI格式错误响应
	var openAIError struct {
		Error types.ErrorResponse `json:"error"`
	}
	if err := json.Unmarshal(body, &openAIError); err == nil && openAIError.Error.Message != "" {
		return types.NewAPIError(openAIError.Error.Type, openAIError.Error.Code,
			openAIError.Error.Message, resp.StatusCode)
	}

	// 如果无法解析，根据状态码生成错误
	return rh.createErrorFromStatusCode(resp.StatusCode, string(body))
}

// createErrorFromStatusCode 根据状态码创建错误
func (rh *ResponseHandler) createErrorFromStatusCode(statusCode int, body string) error {
	switch statusCode {
	case http.StatusBadRequest:
		return types.NewAPIError(types.ErrTypeInvalidRequest, types.ErrCodeInvalidRequest,
			"bad request", statusCode)
	case http.StatusUnauthorized:
		return types.NewAPIError(types.ErrTypeAuthentication, types.ErrCodeUnauthorized,
			"unauthorized", statusCode)
	case http.StatusForbidden:
		return types.NewAPIError(types.ErrTypePermission, types.ErrCodeForbidden,
			"forbidden", statusCode)
	case http.StatusNotFound:
		return types.NewAPIError(types.ErrTypeNotFound, types.ErrCodeNotFound,
			"not found", statusCode)
	case http.StatusTooManyRequests:
		return types.NewAPIError(types.ErrTypeRateLimit, types.ErrCodeTooManyRequests,
			"too many requests", statusCode)
	case http.StatusInternalServerError:
		return types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeInternalError,
			"internal server error", statusCode)
	case http.StatusBadGateway:
		return types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeBadGateway,
			"bad gateway", statusCode)
	case http.StatusServiceUnavailable:
		return types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeServiceUnavailable,
			"service unavailable", statusCode)
	case http.StatusGatewayTimeout:
		return types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeGatewayTimeout,
			"gateway timeout", statusCode)
	default:
		message := fmt.Sprintf("HTTP error %d", statusCode)
		if body != "" {
			message = fmt.Sprintf("%s: %s", message, body)
		}
		return types.NewAPIError(types.ErrTypeAPIError, types.ErrCodeInternalError,
			message, statusCode)
	}
}

// parseResponse 解析成功响应
func (rh *ResponseHandler) parseResponse(ctx context.Context, resp *http.Response, body []byte) (*types.BaseResponse, error) {
	var baseResp types.BaseResponse

	// 尝试解析JSON响应
	if err := json.Unmarshal(body, &baseResp); err != nil {
		// 如果解析失败，创建一个基本响应
		baseResp = types.BaseResponse{
			Object:  "response",
			Created: time.Now().Unix(),
			Data:    json.RawMessage(body),
		}
	}

	return &baseResp, nil
}

// getHeaderMap 获取头部映射
func (rh *ResponseHandler) getHeaderMap(resp *http.Response) map[string]string {
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// GetRetryAfter 获取重试延迟时间
func (rh *ResponseHandler) GetRetryAfter(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}

	// 尝试解析为秒数
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// 尝试解析为HTTP日期
	if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return time.Until(t)
	}

	return 0
}

// GetRateLimit 获取速率限制信息
func (rh *ResponseHandler) GetRateLimit(resp *http.Response) (remaining, limit, reset int64) {
	if remainingStr := resp.Header.Get("X-RateLimit-Remaining"); remainingStr != "" {
		remaining, _ = strconv.ParseInt(remainingStr, 10, 64)
	}

	if limitStr := resp.Header.Get("X-RateLimit-Limit"); limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}

	if resetStr := resp.Header.Get("X-RateLimit-Reset"); resetStr != "" {
		reset, _ = strconv.ParseInt(resetStr, 10, 64)
	}

	return
}

// IsRetryable 检查错误是否可重试
func (rh *ResponseHandler) IsRetryable(err error) bool {
	if apiErr, ok := err.(*types.APIError); ok {
		return apiErr.IsRetryable()
	}
	return false
}

// ShouldRetry 检查响应是否应该重试
func (rh *ResponseHandler) ShouldRetry(resp *http.Response) bool {
	switch resp.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}
