package types

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 错误码常量
const (
	// 客户端错误
	ErrCodeInvalidRequest    = "invalid_request"
	ErrCodeInvalidAPIKey     = "invalid_api_key"
	ErrCodeInvalidModel      = "invalid_model"
	ErrCodeInvalidParameter  = "invalid_parameter"
	ErrCodeMissingParameter  = "missing_parameter"
	ErrCodeRateLimitExceeded = "rate_limit_exceeded"
	ErrCodeQuotaExceeded     = "quota_exceeded"
	ErrCodeInsufficientQuota = "insufficient_quota"
	ErrCodeUnauthorized      = "unauthorized"
	ErrCodeForbidden         = "forbidden"
	ErrCodeNotFound          = "not_found"
	ErrCodeMethodNotAllowed  = "method_not_allowed"
	ErrCodeConflict          = "conflict"
	ErrCodeTooManyRequests   = "too_many_requests"
	ErrCodeRequestTimeout    = "request_timeout"
	ErrCodePayloadTooLarge   = "payload_too_large"
	ErrCodeUnsupportedMedia  = "unsupported_media_type"

	// 服务器错误
	ErrCodeInternalError      = "internal_error"
	ErrCodeServiceUnavailable = "service_unavailable"
	ErrCodeBadGateway         = "bad_gateway"
	ErrCodeGatewayTimeout     = "gateway_timeout"

	// 网络错误
	ErrCodeNetworkError      = "network_error"
	ErrCodeConnectionError   = "connection_error"
	ErrCodeConnectionTimeout = "connection_timeout"
	ErrCodeSSLError          = "ssl_error"

	// 数据处理错误
	ErrCodeParseError      = "parse_error"
	ErrCodeValidationError = "validation_error"
	ErrCodeEncodingError   = "encoding_error"
	ErrCodeDecodingError   = "decoding_error"

	// 流式处理错误
	ErrCodeStreamError   = "stream_error"
	ErrCodeStreamClosed  = "stream_closed"
	ErrCodeStreamTimeout = "stream_timeout"
)

// 错误类型常量
const (
	ErrTypeInvalidRequest = "invalid_request_error"
	ErrTypeAuthentication = "authentication_error"
	ErrTypePermission     = "permission_error"
	ErrTypeNotFound       = "not_found_error"
	ErrTypeRateLimit      = "rate_limit_error"
	ErrTypeAPIConnection  = "api_connection_error"
	ErrTypeAPIError       = "api_error"
	ErrTypeTimeout        = "timeout_error"
	ErrTypeValidation     = "validation_error"
	ErrTypeInternal       = "internal_error"
)

// APIError 自定义API错误类型
type APIError struct {
	Type           string      `json:"type"`
	Code           string      `json:"code"`
	Message        string      `json:"message"`
	Param          interface{} `json:"param,omitempty"`
	HTTPStatusCode int         `json:"-"`
	RequestID      string      `json:"request_id,omitempty"`
	Details        interface{} `json:"details,omitempty"`
	Cause          error       `json:"-"`
}

// ValidationError 验证错误类型
type ValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
}

// NetworkError 网络错误类型
type NetworkError struct {
	Type      string `json:"type"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	URL       string `json:"url,omitempty"`
	Method    string `json:"method,omitempty"`
	Cause     error  `json:"-"`
	Retryable bool   `json:"retryable"`
}

// StreamError 流式处理错误类型
type StreamError struct {
	Type      string `json:"type"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	StreamID  string `json:"stream_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
	Cause     error  `json:"-"`
}

// ErrorDetail 错误详情结构体
type ErrorDetail struct {
	Location string      `json:"location,omitempty"`
	Reason   string      `json:"reason,omitempty"`
	Domain   string      `json:"domain,omitempty"`
	Message  string      `json:"message,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// NewAPIError 创建新的API错误
func NewAPIError(errType, code, message string, httpStatusCode int) *APIError {
	return &APIError{
		Type:           errType,
		Code:           code,
		Message:        message,
		HTTPStatusCode: httpStatusCode,
	}
}

// NewValidationError 创建新的验证错误
func NewValidationError(field string, value interface{}, message, code string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    code,
	}
}

// NewNetworkError 创建新的网络错误
func NewNetworkError(errType, code, message string, retryable bool) *NetworkError {
	return &NetworkError{
		Type:      errType,
		Code:      code,
		Message:   message,
		Retryable: retryable,
	}
}

// NewStreamError 创建新的流式错误
func NewStreamError(errType, code, message string) *StreamError {
	return &StreamError{
		Type:    errType,
		Code:    code,
		Message: message,
	}
}

// Error 实现error接口
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 返回底层错误
func (e *APIError) Unwrap() error {
	return e.Cause
}

// Is 检查错误类型
func (e *APIError) Is(target error) bool {
	if t, ok := target.(*APIError); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

// WithCause 添加原因错误
func (e *APIError) WithCause(cause error) *APIError {
	e.Cause = cause
	return e
}

// WithRequestID 添加请求ID
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithDetails 添加错误详情
func (e *APIError) WithDetails(details interface{}) *APIError {
	e.Details = details
	return e
}

// WithParam 添加错误参数
func (e *APIError) WithParam(param interface{}) *APIError {
	e.Param = param
	return e
}

// IsRetryable 检查是否可重试
func (e *APIError) IsRetryable() bool {
	return e.HTTPStatusCode >= 500 || e.HTTPStatusCode == 429
}

// IsClientError 检查是否为客户端错误
func (e *APIError) IsClientError() bool {
	return e.HTTPStatusCode >= 400 && e.HTTPStatusCode < 500
}

// IsServerError 检查是否为服务器错误
func (e *APIError) IsServerError() bool {
	return e.HTTPStatusCode >= 500
}

// ToJSON 转换为JSON字符串
func (e *APIError) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Error 实现error接口
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// Error 实现error接口
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %s", e.Message)
}

// Unwrap 返回底层错误
func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// WithCause 添加原因错误
func (e *NetworkError) WithCause(cause error) *NetworkError {
	e.Cause = cause
	return e
}

// WithURL 添加URL信息
func (e *NetworkError) WithURL(url string) *NetworkError {
	e.URL = url
	return e
}

// WithMethod 添加HTTP方法信息
func (e *NetworkError) WithMethod(method string) *NetworkError {
	e.Method = method
	return e
}

// Error 实现error接口
func (e *StreamError) Error() string {
	return fmt.Sprintf("stream error: %s", e.Message)
}

// Unwrap 返回底层错误
func (e *StreamError) Unwrap() error {
	return e.Cause
}

// WithCause 添加原因错误
func (e *StreamError) WithCause(cause error) *StreamError {
	e.Cause = cause
	return e
}

// WithStreamID 添加流ID
func (e *StreamError) WithStreamID(streamID string) *StreamError {
	e.StreamID = streamID
	return e
}

// WithEventType 添加事件类型
func (e *StreamError) WithEventType(eventType string) *StreamError {
	e.EventType = eventType
	return e
}

// WrapError 包装错误为APIError
func WrapError(err error, errType, code, message string, httpStatusCode int) *APIError {
	return &APIError{
		Type:           errType,
		Code:           code,
		Message:        message,
		HTTPStatusCode: httpStatusCode,
		Cause:          err,
	}
}

// WrapNetworkError 包装网络错误
func WrapNetworkError(err error, code, message string, retryable bool) *NetworkError {
	return &NetworkError{
		Type:      ErrTypeAPIConnection,
		Code:      code,
		Message:   message,
		Retryable: retryable,
		Cause:     err,
	}
}

// WrapStreamError 包装流式错误
func WrapStreamError(err error, code, message string) *StreamError {
	return &StreamError{
		Type:    ErrTypeInternal,
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// FromHTTPStatusCode 根据HTTP状态码创建错误
func FromHTTPStatusCode(statusCode int, message string) *APIError {
	switch statusCode {
	case http.StatusBadRequest:
		return NewAPIError(ErrTypeInvalidRequest, ErrCodeInvalidRequest, message, statusCode)
	case http.StatusUnauthorized:
		return NewAPIError(ErrTypeAuthentication, ErrCodeUnauthorized, message, statusCode)
	case http.StatusForbidden:
		return NewAPIError(ErrTypePermission, ErrCodeForbidden, message, statusCode)
	case http.StatusNotFound:
		return NewAPIError(ErrTypeNotFound, ErrCodeNotFound, message, statusCode)
	case http.StatusMethodNotAllowed:
		return NewAPIError(ErrTypeInvalidRequest, ErrCodeMethodNotAllowed, message, statusCode)
	case http.StatusConflict:
		return NewAPIError(ErrTypeInvalidRequest, ErrCodeConflict, message, statusCode)
	case http.StatusTooManyRequests:
		return NewAPIError(ErrTypeRateLimit, ErrCodeRateLimitExceeded, message, statusCode)
	case http.StatusRequestTimeout:
		return NewAPIError(ErrTypeTimeout, ErrCodeRequestTimeout, message, statusCode)
	case http.StatusRequestEntityTooLarge:
		return NewAPIError(ErrTypeInvalidRequest, ErrCodePayloadTooLarge, message, statusCode)
	case http.StatusUnsupportedMediaType:
		return NewAPIError(ErrTypeInvalidRequest, ErrCodeUnsupportedMedia, message, statusCode)
	case http.StatusInternalServerError:
		return NewAPIError(ErrTypeAPIError, ErrCodeInternalError, message, statusCode)
	case http.StatusBadGateway:
		return NewAPIError(ErrTypeAPIConnection, ErrCodeBadGateway, message, statusCode)
	case http.StatusServiceUnavailable:
		return NewAPIError(ErrTypeAPIError, ErrCodeServiceUnavailable, message, statusCode)
	case http.StatusGatewayTimeout:
		return NewAPIError(ErrTypeTimeout, ErrCodeGatewayTimeout, message, statusCode)
	default:
		return NewAPIError(ErrTypeAPIError, ErrCodeInternalError, message, statusCode)
	}
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsRetryable()
	}
	if netErr, ok := err.(*NetworkError); ok {
		return netErr.Retryable
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) string {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Code
	}
	if valErr, ok := err.(*ValidationError); ok {
		return valErr.Code
	}
	if netErr, ok := err.(*NetworkError); ok {
		return netErr.Code
	}
	if streamErr, ok := err.(*StreamError); ok {
		return streamErr.Code
	}
	return "unknown_error"
}

// GetErrorType 获取错误类型
func GetErrorType(err error) string {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.Type
	}
	if netErr, ok := err.(*NetworkError); ok {
		return netErr.Type
	}
	if streamErr, ok := err.(*StreamError); ok {
		return streamErr.Type
	}
	return "unknown_error"
}
